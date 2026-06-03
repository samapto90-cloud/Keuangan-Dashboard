package main

import (
	"net/http"
	"strings"
	"sync"
)

type SipkeuModule struct {
	ID      string
	BPKCode string
	mu      sync.Mutex
	txs     []Transaction
	nextID  int
	settings AppSettings
}

var (
	sipkeuModules   map[string]*SipkeuModule
	sipkeuModulesMu sync.RWMutex
)

func initSipkeuModules() {
	sipkeuModules = map[string]*SipkeuModule{
		"sekretariat": {
			ID:      "sekretariat",
			BPKCode: "B01",
			settings: AppSettings{
				PA:               Pejabat{Nama: "HENDRI ARULAN, S.Pd", Nip: "NIP. 19670119 199103 1 009"},
				Bendahara:        Pejabat{Nama: "ELDINA SRIDHANTY, SE", Nip: "NIP. 19810610 201001 2 002"},
				AnggaranKegiatan: map[string]float64{},
			},
		},
		"paud": {
			ID:      "paud",
			BPKCode: "B02",
			settings: AppSettings{
				PA:               Pejabat{Nama: "Sularno, S.Pd., M.Si", Nip: "NIP. 19690408 199303 1 010"},
				Bendahara:        Pejabat{Nama: "LIONI ASMIRELDA", Nip: "NIP. 19981102 202203 2 002"},
				AnggaranKegiatan: map[string]float64{},
			},
		},
	}
}

func moduleFromRequest(r *http.Request) *SipkeuModule {
	key := strings.TrimSpace(r.Header.Get("X-SIPKEU-App"))
	if key == "" {
		key = "sekretariat"
	}
	sipkeuModulesMu.RLock()
	mod, ok := sipkeuModules[key]
	sipkeuModulesMu.RUnlock()
	if !ok {
		return sipkeuModules["sekretariat"]
	}
	return mod
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

func applyRakToModule(mod *SipkeuModule, rows []RakRow) ImportAnggaranResult {
	anggaranMap := map[string]float64{}
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
	}
	mod.mu.Lock()
	mod.settings.Rak = rows
	if mod.settings.AnggaranKegiatan == nil {
		mod.settings.AnggaranKegiatan = map[string]float64{}
	}
	for k, v := range anggaranMap {
		mod.settings.AnggaranKegiatan[k] = v
	}
	mod.mu.Unlock()
	return ImportAnggaranResult{
		Rak:              rows,
		AnggaranKegiatan: anggaranMap,
		TotalBaris:       len(rows),
	}
}
