package main

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type SipkeuModule struct {
	ID              string
	BPKCode         string
	mu              sync.Mutex
	txs             []Transaction
	nextID          int
	settings        AppSettings
	defaultSettings AppSettings
}

var (
	sipkeuModules   map[string]*SipkeuModule
	sipkeuModulesMu sync.RWMutex
)

func initSipkeuModules() {
	sipkeuModules = map[string]*SipkeuModule{
		"sekretariat": newSipkeuModule("sekretariat", "B01", Pejabat{Nama: "HENDRI ARULAN, S.Pd", Nip: "NIP. 19670119 199103 1 009"}, Pejabat{Nama: "ELDINA SRIDHANTY, SE", Nip: "NIP. 19810610 201001 2 002"}),
		"paud":        newSipkeuModule("paud", "B02", Pejabat{Nama: "Sularno, S.Pd., M.Si", Nip: "NIP. 19690408 199303 1 010"}, Pejabat{Nama: "LIONI ASMIRELDA", Nip: "NIP. 19981102 202203 2 002"}),
		"sd":          newSipkeuModule("sd", "B03", Pejabat{Nama: "Yusal, S.Pd., M.M", Nip: "NIP. 19700809 199304 1 002"}, Pejabat{Nama: "JULIA ARSIA NAFHASYA, S.Pd", Nip: "NIP. 19920716 201903 2 001"}),
		"smp":         newSipkeuModule("smp", "B04", Pejabat{Nama: "Joni Satria Putra, S.Pd., M.Si", Nip: "NIP. 19730101 200502 1 006"}, Pejabat{Nama: "SHINTA WIDYASTATI", Nip: "NIP. 19790303 200502 2 005"}),
	}
}

func newSipkeuModule(id, bpk string, pa, bend Pejabat) *SipkeuModule {
	settings := AppSettings{
		PA:               pa,
		Bendahara:        bend,
		AnggaranKegiatan: map[string]float64{},
	}
	return &SipkeuModule{
		ID:              id,
		BPKCode:         bpk,
		nextID:          1,
		settings:        settings,
		defaultSettings: settings,
	}
}

func transactionBelongsToModule(mod *SipkeuModule, t Transaction) bool {
	if mod == nil {
		return false
	}
	no := strings.TrimSpace(t.NoBPK)
	if no == "" {
		return true
	}
	return strings.Contains(no, "/"+mod.BPKCode+"/")
}

func importTransactionAllowed(mod *SipkeuModule, t Transaction) bool {
	if mod == nil {
		return false
	}
	if !transactionBelongsToModule(mod, t) {
		return false
	}
	np2d := strings.TrimSpace(t.NoNP2D)
	if np2d != "" && !strings.Contains(np2d, "/"+mod.BPKCode+"/") {
		return false
	}
	return true
}

func moduleTransactionsCopy(mod *SipkeuModule) []Transaction {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	out := make([]Transaction, 0, len(mod.txs))
	for _, t := range mod.txs {
		if transactionBelongsToModule(mod, t) {
			out = append(out, t)
		}
	}
	sortTransactionsNewestFirst(out)
	return out
}

func sortTransactionsNewestFirst(txs []Transaction) {
	sort.Slice(txs, func(i, j int) bool {
		if txs[i].ID != txs[j].ID {
			return txs[i].ID > txs[j].ID
		}
		return txs[i].Tanggal > txs[j].Tanggal
	})
}

func repairModuleIsolation(mod *SipkeuModule) bool {
	if mod == nil {
		return false
	}

	mod.mu.Lock()
	defer mod.mu.Unlock()
	changed := false

	filtered := make([]Transaction, 0, len(mod.txs))
	for _, t := range mod.txs {
		if transactionBelongsToModule(mod, t) {
			filtered = append(filtered, t)
		} else {
			changed = true
		}
	}
	if len(filtered) != len(mod.txs) {
		mod.txs = filtered
		changed = true
	}

	if isPejabatFromOtherModule(mod.ID, mod.settings.PA.Nama, true) &&
		!isPejabatFromOtherModule(mod.ID, mod.defaultSettings.PA.Nama, true) {
		mod.settings.PA = mod.defaultSettings.PA
		changed = true
	}
	if isPejabatFromOtherModule(mod.ID, mod.settings.Bendahara.Nama, false) &&
		!isPejabatFromOtherModule(mod.ID, mod.defaultSettings.Bendahara.Nama, false) {
		mod.settings.Bendahara = mod.defaultSettings.Bendahara
		changed = true
	}
	return changed
}

func pejabatNameTokens(name string) []string {
	u := strings.ToUpper(strings.TrimSpace(name))
	if u == "" {
		return nil
	}
	raw := strings.FieldsFunc(u, func(r rune) bool {
		return r == ',' || r == ' ' || r == '.'
	})
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if len(p) >= 4 {
			out = append(out, p)
		}
	}
	return out
}

func pejabatNamesMatch(a, b string) bool {
	na := strings.ToUpper(strings.TrimSpace(a))
	nb := strings.ToUpper(strings.TrimSpace(b))
	if na == "" || nb == "" {
		return false
	}
	if na == nb {
		return true
	}
	ta := pejabatNameTokens(a)
	tb := pejabatNameTokens(b)
	for _, x := range ta {
		for _, y := range tb {
			if x == y {
				return true
			}
		}
	}
	return false
}

func isPejabatFromOtherModule(modID, nama string, isPA bool) bool {
	if strings.TrimSpace(nama) == "" {
		return false
	}
	sipkeuModulesMu.RLock()
	defer sipkeuModulesMu.RUnlock()
	for id, other := range sipkeuModules {
		if id == modID {
			continue
		}
		def := other.defaultSettings.PA
		if !isPA {
			def = other.defaultSettings.Bendahara
		}
		if pejabatNamesMatch(nama, def.Nama) {
			return true
		}
	}
	return false
}

func repairAllModulesIsolation() {
	sipkeuModulesMu.RLock()
	mods := make([]*SipkeuModule, 0, len(sipkeuModules))
	for _, m := range sipkeuModules {
		mods = append(mods, m)
	}
	sipkeuModulesMu.RUnlock()
	for _, mod := range mods {
		if repairModuleIsolation(mod) {
			persistModule(mod)
			log.Printf("Isolasi modul %s diperbaiki (transaksi/pejabat)", mod.ID)
		}
	}
}

func effectivePejabatValues(modID string, pa, bend, defPA, defBend Pejabat) (Pejabat, Pejabat) {
	if isPejabatFromOtherModule(modID, pa.Nama, true) {
		pa = defPA
	}
	if isPejabatFromOtherModule(modID, bend.Nama, false) {
		bend = defBend
	}
	return pa, bend
}

func moduleFromRequest(r *http.Request) *SipkeuModule {
	key := ""
	if sess := getSession(r); sess != nil && strings.TrimSpace(sess.AppModule) != "" {
		key = strings.TrimSpace(sess.AppModule)
	}
	if key == "" {
		key = strings.TrimSpace(r.Header.Get("X-SIPKEU-App"))
	}
	if key == "" {
		key = "sekretariat"
	}
	sipkeuModulesMu.RLock()
	mod, ok := sipkeuModules[key]
	fallback := sipkeuModules["sekretariat"]
	sipkeuModulesMu.RUnlock()
	if !ok {
		return fallback
	}
	return mod
}

func normalizeModuleIDs(mod *SipkeuModule) {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	if mod.nextID <= 0 {
		mod.nextID = 1
	}
	maxID := 0
	for i := range mod.txs {
		if mod.txs[i].ID <= 0 {
			mod.txs[i].ID = mod.nextID
			mod.nextID++
		}
		if mod.txs[i].ID > maxID {
			maxID = mod.txs[i].ID
		}
	}
	if maxID >= mod.nextID {
		mod.nextID = maxID + 1
	}
}

func cloneRakRows(rows []RakRow) []RakRow {
	out := make([]RakRow, len(rows))
	copy(out, rows)
	return out
}

func cloneAnggaranMap(src map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range src {
		out[k] = v
	}
	return out
}

func applyRakToModule(mod *SipkeuModule, rows []RakRow, meta *RakMeta) ImportAnggaranResult {
	anggaranMap := map[string]float64{}
	var totalAnggaran float64
	for i := range rows {
		rows[i].Kegiatan = strings.TrimSpace(rows[i].Kegiatan)
		rows[i].SubKegiatan = strings.TrimSpace(rows[i].SubKegiatan)
		rows[i].KodeRekening = strings.TrimSpace(rows[i].KodeRekening)
		rows[i].Pekerjaan = strings.TrimSpace(rows[i].Pekerjaan)
		if rows[i].PPTK == "" {
			rows[i].PPTK = pptkForRow(rows[i].KodeRekening, rows[i].Pekerjaan)
		}
		if rows[i].Kegiatan != "" {
			anggaranMap[rows[i].Kegiatan] += rows[i].Anggaran
		}
		totalAnggaran += rows[i].Anggaran
	}
	mod.mu.Lock()
	mod.settings.Rak = rows
	if mod.settings.AnggaranKegiatan == nil {
		mod.settings.AnggaranKegiatan = map[string]float64{}
	}
	for k, v := range anggaranMap {
		mod.settings.AnggaranKegiatan[k] = v
	}
	if meta != nil {
		mod.settings.RakMeta = *meta
	}
	outMeta := mod.settings.RakMeta
	mod.mu.Unlock()
	persistModule(mod)
	return ImportAnggaranResult{
		Rak:              rows,
		AnggaranKegiatan: anggaranMap,
		RakMeta:          outMeta,
		TotalBaris:       len(rows),
	}
}
