package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

var adminRekapSipkeuIDs = []string{"sekretariat", "paud", "sd", "smp"}

type adminRekapRow struct {
	PortalID     string  `json:"portal_id"`
	PortalLabel  string  `json:"portal_label"`
	Kegiatan     string  `json:"kegiatan,omitempty"`
	SubKegiatan  string  `json:"sub_kegiatan,omitempty"`
	Pekerjaan    string  `json:"pekerjaan,omitempty"`
	KodeRekening string  `json:"kode_rekening,omitempty"`
	Anggaran     float64 `json:"anggaran"`
	Realisasi    float64 `json:"realisasi"`
	Sisa         float64 `json:"sisa"`
	Count        int     `json:"count"`
	Pajak        float64 `json:"pajak"`
	Pct          float64 `json:"pct"`
}

type rekapAgg struct {
	portalID, portalLabel          string
	kegiatan, sub, pekerjaan, kode string
	anggaran, realisasi, pajak     float64
	count                          int
}

func parseAdminRekapPortals(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "all" {
		return append([]string(nil), adminRekapSipkeuIDs...)
	}
	allowed := map[string]bool{}
	for _, id := range adminRekapSipkeuIDs {
		allowed[id] = true
	}
	out := []string{}
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if allowed[p] {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), adminRekapSipkeuIDs...)
	}
	return out
}

func transactionInDateRange(t Transaction, from, to string) bool {
	d := strings.TrimSpace(t.Tanggal)
	if d == "" {
		return false
	}
	if from != "" && d < from {
		return false
	}
	if to != "" && d > to {
		return false
	}
	return true
}

func rekapPct(realisasi, anggaran float64) float64 {
	if anggaran <= 0 {
		return 0
	}
	return (realisasi / anggaran) * 100
}

func anggaranKegiatanFor(mod *SipkeuModule, kegiatan string, realisasi float64) float64 {
	if mod == nil || kegiatan == "" {
		return 0
	}
	mod.mu.Lock()
	defer mod.mu.Unlock()
	if mod.settings.AnggaranKegiatan != nil {
		if v := mod.settings.AnggaranKegiatan[kegiatan]; v > 0 {
			return v
		}
	}
	var sum float64
	for _, r := range mod.settings.Rak {
		if r.Kegiatan == kegiatan {
			sum += r.Anggaran
		}
	}
	if sum > 0 {
		return sum
	}
	if realisasi > 0 {
		padded := int(realisasi*1.25/1000000) + 1
		if padded < 100 {
			padded = 100
		}
		return float64(padded) * 1000000
	}
	return 0
}

func anggaranSubFor(mod *SipkeuModule, kegiatan, sub string) float64 {
	if mod == nil || kegiatan == "" || sub == "" {
		return 0
	}
	mod.mu.Lock()
	defer mod.mu.Unlock()
	var sum float64
	for _, r := range mod.settings.Rak {
		if r.Kegiatan == kegiatan && r.SubKegiatan == sub {
			sum += r.Anggaran
		}
	}
	return sum
}

func anggaranPekerjaanFor(mod *SipkeuModule, kegiatan, sub, kode, pekerjaan string) float64 {
	if mod == nil {
		return 0
	}
	mod.mu.Lock()
	defer mod.mu.Unlock()
	for _, r := range mod.settings.Rak {
		if r.Kegiatan == kegiatan && r.SubKegiatan == sub &&
			r.KodeRekening == kode && r.Pekerjaan == pekerjaan {
			return r.Anggaran
		}
	}
	return 0
}

func portalTotalAnggaran(mod *SipkeuModule) float64 {
	if mod == nil {
		return 0
	}
	mod.mu.Lock()
	defer mod.mu.Unlock()
	var total float64
	for _, r := range mod.settings.Rak {
		total += r.Anggaran
	}
	if total > 0 {
		return total
	}
	for _, v := range mod.settings.AnggaranKegiatan {
		if v > 0 {
			total += v
		}
	}
	return total
}

func rekapAggKey(mode string, a *rekapAgg) string {
	switch mode {
	case "portal":
		return a.portalID
	case "sub":
		return a.portalID + "\x00" + a.kegiatan + "\x00" + a.sub
	case "pekerjaan":
		return a.portalID + "\x00" + a.kegiatan + "\x00" + a.sub + "\x00" + a.kode + "\x00" + a.pekerjaan
	default:
		return a.portalID + "\x00" + a.kegiatan
	}
}

func seedRekapFromRak(mod *SipkeuModule, mode string, aggs map[string]*rekapAgg) {
	if mod == nil {
		return
	}
	label := portalLabel(mod.ID)
	mod.mu.Lock()
	rak := append([]RakRow(nil), mod.settings.Rak...)
	anggaranKeg := map[string]float64{}
	for k, v := range mod.settings.AnggaranKegiatan {
		anggaranKeg[k] = v
	}
	mod.mu.Unlock()

	switch mode {
	case "portal":
		key := mod.ID
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{portalID: mod.ID, portalLabel: label, anggaran: portalTotalAnggaran(mod)}
		}
	case "kegiatan":
		seen := map[string]bool{}
		for k := range anggaranKeg {
			seen[k] = true
			a := &rekapAgg{portalID: mod.ID, portalLabel: label, kegiatan: k, anggaran: anggaranKeg[k]}
			aggs[rekapAggKey(mode, a)] = a
		}
		for _, r := range rak {
			if r.Kegiatan == "" {
				continue
			}
			key := rekapAggKey(mode, &rekapAgg{portalID: mod.ID, kegiatan: r.Kegiatan})
			if aggs[key] != nil {
				continue
			}
			if !seen[r.Kegiatan] {
				seen[r.Kegiatan] = true
				aggs[key] = &rekapAgg{portalID: mod.ID, portalLabel: label, kegiatan: r.Kegiatan}
			}
		}
	case "sub":
		for _, r := range rak {
			if r.Kegiatan == "" || r.SubKegiatan == "" {
				continue
			}
			key := rekapAggKey(mode, &rekapAgg{portalID: mod.ID, kegiatan: r.Kegiatan, sub: r.SubKegiatan})
			if aggs[key] == nil {
				aggs[key] = &rekapAgg{
					portalID: mod.ID, portalLabel: label,
					kegiatan: r.Kegiatan, sub: r.SubKegiatan,
				}
			}
		}
	case "pekerjaan":
		for _, r := range rak {
			if r.Kegiatan == "" || r.SubKegiatan == "" || r.Pekerjaan == "" {
				continue
			}
			key := rekapAggKey(mode, &rekapAgg{
				portalID: mod.ID, kegiatan: r.Kegiatan, sub: r.SubKegiatan,
				kode: r.KodeRekening, pekerjaan: r.Pekerjaan,
			})
			if aggs[key] == nil {
				aggs[key] = &rekapAgg{
					portalID: mod.ID, portalLabel: label,
					kegiatan: r.Kegiatan, sub: r.SubKegiatan,
					kode: r.KodeRekening, pekerjaan: r.Pekerjaan,
					anggaran: r.Anggaran,
				}
			}
		}
	}
}

func applyRekapTransaction(mod *SipkeuModule, mode, from, to string, t Transaction, aggs map[string]*rekapAgg) {
	if mod == nil || !trxIsApproved(t) || !transactionBelongsToModule(mod, t) {
		return
	}
	if !transactionInDateRange(t, from, to) {
		return
	}
	label := portalLabel(mod.ID)
	var a *rekapAgg
	switch mode {
	case "portal":
		key := mod.ID
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{portalID: mod.ID, portalLabel: label, anggaran: portalTotalAnggaran(mod)}
		}
		a = aggs[key]
	case "sub":
		if t.Kegiatan == "" || t.SubKegiatan == "" {
			return
		}
		key := rekapAggKey(mode, &rekapAgg{portalID: mod.ID, kegiatan: t.Kegiatan, sub: t.SubKegiatan})
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{
				portalID: mod.ID, portalLabel: label,
				kegiatan: t.Kegiatan, sub: t.SubKegiatan,
			}
		}
		a = aggs[key]
	case "pekerjaan":
		if t.Kegiatan == "" || t.SubKegiatan == "" || t.Pekerjaan == "" {
			return
		}
		key := rekapAggKey(mode, &rekapAgg{
			portalID: mod.ID, kegiatan: t.Kegiatan, sub: t.SubKegiatan,
			kode: t.KodeRekening, pekerjaan: t.Pekerjaan,
		})
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{
				portalID: mod.ID, portalLabel: label,
				kegiatan: t.Kegiatan, sub: t.SubKegiatan,
				kode: t.KodeRekening, pekerjaan: t.Pekerjaan,
			}
		}
		a = aggs[key]
	default:
		if t.Kegiatan == "" {
			return
		}
		key := rekapAggKey(mode, &rekapAgg{portalID: mod.ID, kegiatan: t.Kegiatan})
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{portalID: mod.ID, portalLabel: label, kegiatan: t.Kegiatan}
		}
		a = aggs[key]
	}
	a.realisasi += t.Nilai
	a.pajak += t.Pajak
	a.count++
}

func finalizeRekapAnggaran(mod *SipkeuModule, mode string, aggs map[string]*rekapAgg) {
	for _, a := range aggs {
		if a.portalID != mod.ID {
			continue
		}
		switch mode {
		case "kegiatan":
			if a.anggaran <= 0 {
				a.anggaran = anggaranKegiatanFor(mod, a.kegiatan, a.realisasi)
			}
		case "sub":
			if a.anggaran <= 0 {
				a.anggaran = anggaranSubFor(mod, a.kegiatan, a.sub)
			}
		case "pekerjaan":
			if a.anggaran <= 0 {
				a.anggaran = anggaranPekerjaanFor(mod, a.kegiatan, a.sub, a.kode, a.pekerjaan)
			}
		case "portal":
			if a.anggaran <= 0 {
				a.anggaran = portalTotalAnggaran(mod)
			}
		}
	}
}

func buildAdminRekap(portals []string, mode, from, to string) []adminRekapRow {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = "kegiatan"
	}
	aggs := map[string]*rekapAgg{}

	for _, portalID := range portals {
		mod := sipkeuModules[portalID]
		if mod == nil {
			continue
		}
		seedRekapFromRak(mod, mode, aggs)
		txs := moduleTransactionsCopy(mod)
		for i := range txs {
			applyRekapTransaction(mod, mode, from, to, txs[i], aggs)
		}
		finalizeRekapAnggaran(mod, mode, aggs)
	}

	rows := make([]adminRekapRow, 0, len(aggs))
	for _, a := range aggs {
		if a.count == 0 && a.anggaran <= 0 && mode != "portal" {
			continue
		}
		sisa := a.anggaran - a.realisasi
		rows = append(rows, adminRekapRow{
			PortalID:     a.portalID,
			PortalLabel:  a.portalLabel,
			Kegiatan:     a.kegiatan,
			SubKegiatan:  a.sub,
			Pekerjaan:    a.pekerjaan,
			KodeRekening: a.kode,
			Anggaran:     a.anggaran,
			Realisasi:    a.realisasi,
			Sisa:         sisa,
			Count:        a.count,
			Pajak:        a.pajak,
			Pct:          rekapPct(a.realisasi, a.anggaran),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].PortalID != rows[j].PortalID {
			return rows[i].PortalID < rows[j].PortalID
		}
		if rows[i].Anggaran != rows[j].Anggaran {
			return rows[i].Anggaran > rows[j].Anggaran
		}
		return rows[i].Realisasi > rows[j].Realisasi
	})
	return rows
}

func handleAdminRekapitulasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	mode := strings.TrimSpace(r.URL.Query().Get("mode"))
	if mode == "" {
		mode = "kegiatan"
	}
	validModes := map[string]bool{"kegiatan": true, "sub": true, "pekerjaan": true, "portal": true}
	if !validModes[mode] {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Mode rekap tidak valid"})
		return
	}
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	to := strings.TrimSpace(r.URL.Query().Get("to"))
	portals := parseAdminRekapPortals(r.URL.Query().Get("portals"))
	rows := buildAdminRekap(portals, mode, from, to)

	var totAng, totReal, totPajak float64
	var totTrx int
	for _, row := range rows {
		totAng += row.Anggaran
		totReal += row.Realisasi
		totPajak += row.Pajak
		totTrx += row.Count
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"mode":         mode,
		"from":         from,
		"to":           to,
		"portals":      portals,
		"rows":         rows,
		"generated_at": time.Now().Format(time.RFC3339),
		"summary": map[string]interface{}{
			"anggaran":  totAng,
			"realisasi": totReal,
			"sisa":      totAng - totReal,
			"pajak":     totPajak,
			"count":     totTrx,
			"pct":       rekapPct(totReal, totAng),
		},
		"labels": map[string]string{
			"mode": fmt.Sprintf("Rekapitulasi %s", map[string]string{
				"kegiatan": "Per Kegiatan", "sub": "Per Sub Kegiatan",
				"pekerjaan": "Per Pekerjaan", "portal": "Per Portal",
			}[mode]),
		},
	})
}
