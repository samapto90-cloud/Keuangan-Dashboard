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
			nextID:  1,
			settings: AppSettings{
				PA:               Pejabat{Nama: "HENDRI ARULAN, S.Pd", Nip: "NIP. 19670119 199103 1 009"},
				Bendahara:        Pejabat{Nama: "ELDINA SRIDHANTY, SE", Nip: "NIP. 19810610 201001 2 002"},
				AnggaranKegiatan: map[string]float64{},
			},
		},
		"paud": {
			ID:      "paud",
			BPKCode: "B02",
			nextID:  1,
			settings: AppSettings{
				PA:               Pejabat{Nama: "Sularno, S.Pd., M.Si", Nip: "NIP. 19690408 199303 1 010"},
				Bendahara:        Pejabat{Nama: "LIONI ASMIRELDA", Nip: "NIP. 19981102 202203 2 002"},
				AnggaranKegiatan: map[string]float64{},
			},
		},
		"sd": {
			ID:      "sd",
			BPKCode: "B03",
			nextID:  1,
			settings: AppSettings{
				PA:               Pejabat{Nama: "Yusal, S.Pd., M.M", Nip: "NIP. 19700809 199304 1 002"},
				Bendahara:        Pejabat{Nama: "JULIA ARSIA NAFHASYA, S.Pd", Nip: "NIP. 19920716 201903 2 001"},
				AnggaranKegiatan: map[string]float64{},
			},
		},
		"smp": {
			ID:      "smp",
			BPKCode: "B04",
			nextID:  1,
			settings: AppSettings{
				PA:               Pejabat{Nama: "Joni Satria Putra, S.Pd., M.Si", Nip: "NIP. 19730101 200502 1 006"},
				Bendahara:        Pejabat{Nama: "SHINTA WIDYASTATI", Nip: "NIP. 19790303 200502 2 005"},
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
