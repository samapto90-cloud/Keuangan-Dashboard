package main

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	bpkUnitCode = "1.01.0.00.0.00.01.0000"
	np2dDocType = "NP2D-PPTK"
)

var np2dSeqPrefixRe = regexp.MustCompile(`^(\d+)/`)

func parseNp2dSeqValue(noNp2d, bpkCode string) int {
	noNp2d = strings.TrimSpace(noNp2d)
	if noNp2d == "" || !strings.Contains(noNp2d, "/"+bpkCode+"/") {
		return 0
	}
	m := np2dSeqPrefixRe.FindStringSubmatch(noNp2d)
	if len(m) < 2 {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func isNewNp2dFormat(noNp2d, bpkCode string) bool {
	noNp2d = strings.TrimSpace(noNp2d)
	if noNp2d == "" {
		return false
	}
	want := fmt.Sprintf("/%s/%s/%s/", np2dDocType, bpkUnitCode, bpkCode)
	return strings.Contains(noNp2d, want)
}

func formatNp2dNumber(seq int, bpkCode, month, year string) string {
	return fmt.Sprintf("%04d/%s/%s/%s/%s/%s", seq, np2dDocType, bpkUnitCode, bpkCode, month, year)
}

func tanggalParts(tanggal string) (year, month string, ok bool) {
	parts := strings.Split(strings.TrimSpace(tanggal), "-")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func backfillModuleNP2D(mod *SipkeuModule) int {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	bpkCode := mod.BPKCode
	maxSeq := 0
	for _, t := range mod.txs {
		maxSeq = maxInt(maxSeq, parseNp2dSeqValue(t.NoNP2D, bpkCode))
	}

	updated := 0

	// Konversi format lama ke format baru (pertahankan nomor urut).
	for i := range mod.txs {
		t := &mod.txs[i]
		existing := strings.TrimSpace(t.NoNP2D)
		if existing == "" || isNewNp2dFormat(existing, bpkCode) {
			continue
		}
		year, month, ok := tanggalParts(t.Tanggal)
		if !ok {
			continue
		}
		seq := parseNp2dSeqValue(existing, bpkCode)
		if seq <= 0 {
			continue
		}
		newNo := formatNp2dNumber(seq, bpkCode, month, year)
		if newNo != existing {
			t.NoNP2D = newNo
			updated++
		}
		maxSeq = maxInt(maxSeq, seq)
	}

	// Isi NP2D kosong — urut berdasarkan tanggal lalu ID.
	type txRef struct {
		idx int
	}
	var empty []txRef
	for i := range mod.txs {
		if strings.TrimSpace(mod.txs[i].NoNP2D) != "" {
			continue
		}
		if _, _, ok := tanggalParts(mod.txs[i].Tanggal); !ok {
			continue
		}
		empty = append(empty, txRef{idx: i})
	}
	sort.Slice(empty, func(a, b int) bool {
		ta, tb := mod.txs[empty[a].idx], mod.txs[empty[b].idx]
		if ta.Tanggal != tb.Tanggal {
			return ta.Tanggal < tb.Tanggal
		}
		return ta.ID < tb.ID
	})

	for _, ref := range empty {
		t := &mod.txs[ref.idx]
		year, month, ok := tanggalParts(t.Tanggal)
		if !ok {
			continue
		}
		maxSeq++
		t.NoNP2D = formatNp2dNumber(maxSeq, bpkCode, month, year)
		updated++
	}

	return updated
}

func backfillAllModulesNP2D() map[string]int {
	sipkeuModulesMu.RLock()
	modules := make([]*SipkeuModule, 0, len(sipkeuModules))
	for _, mod := range sipkeuModules {
		modules = append(modules, mod)
	}
	sipkeuModulesMu.RUnlock()

	out := map[string]int{}
	for _, mod := range modules {
		n := backfillModuleNP2D(mod)
		out[mod.ID] = n
		if n > 0 {
			persistModule(mod)
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func handleBackfillNP2D(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	results := backfillAllModulesNP2D()
	total := 0
	for _, n := range results {
		total += n
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("%d transaksi diperbarui nomor NP2D-nya", total),
		"updated": total,
		"modules": results,
	})
}
