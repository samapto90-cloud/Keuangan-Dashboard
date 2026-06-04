package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

var bulanKeys = []string{
	"januari", "februari", "maret", "april", "mei", "juni",
	"juli", "agustus", "september", "oktober", "november", "desember",
}

type RakBelanjaRow struct {
	KodeRekening    string             `json:"kode_rekening"`
	NamaRekening    string             `json:"nama_rekening"`
	NamaKegiatan    string             `json:"nama_kegiatan"`
	NamaSubKegiatan string             `json:"nama_sub_kegiatan"`
	Anggaran        float64            `json:"anggaran"`
	Bulan           map[string]float64 `json:"bulan"`
}

type KasReportRow struct {
	Kode           string  `json:"kode"`
	Uraian         string  `json:"uraian"`
	Level          int     `json:"level"`
	SisaBulanLalu  float64 `json:"sisa_bulan_lalu"`
	AnggaranKas    float64 `json:"anggaran_kas"`
	Realisasi      float64 `json:"realisasi"`
	SisaSD         float64 `json:"sisa_sd"`
	Persen         float64 `json:"persen"`
	Editable       bool    `json:"editable"`
}

type KasBelanjaState struct {
	Tahun           int                             `json:"tahun"`
	RakRows         []RakBelanjaRow                 `json:"rak_rows"`
	Realisasi       map[string]map[string]float64   `json:"realisasi"`
	SisaManual      map[string]map[string]float64   `json:"sisa_manual"`
	RealisasiLocked map[string]bool                 `json:"realisasi_locked"`
	ImportedAt      string                          `json:"imported_at"`
	Version         string                          `json:"version,omitempty"`
	VersionLabel    string                          `json:"version_label,omitempty"`
	Message         string                          `json:"message,omitempty"`
}

var (
	kasState = KasBelanjaState{
		Tahun:           2026,
		RakRows:         []RakBelanjaRow{},
		Realisasi:       map[string]map[string]float64{},
		SisaManual:      map[string]map[string]float64{},
		RealisasiLocked: map[string]bool{},
	}
	kasMu sync.RWMutex
)

var angkasTemplate = []struct {
	Kode   string
	Uraian string
	Level  int
}{
	{"5.", "BELANJA DAERAH", 0},
	{"5.1.", "BELANJA OPERASI", 1},
	{"5.1.01.", "Belanja Pegawai", 2},
	{"5.1.02.", "Belanja Barang dan Jasa", 2},
	{"5.1.05.", "Belanja Hibah", 2},
	{"5.1.06.", "Belanja Bantuan Sosial", 2},
	{"5.2.", "BELANJA MODAL", 1},
	{"5.2.02.", "Belanja Modal Peralatan dan Mesin", 2},
	{"5.2.03.", "Belanja Modal Gedung dan Bangunan", 2},
	{"5.2.04.", "Belanja Modal Jalan, Jaringan, dan Irigasi", 2},
	{"5.2.05.", "Belanja Modal Aset Tetap Lainnya", 2},
	{"5.3.", "BELANJA TIDAK TERDUGA", 1},
	{"5.3.01.", "Belanja Tidak Terduga", 2},
}

func currentBulanKey() string {
	now := time.Now()
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		now = now.In(loc)
	}
	idx := int(now.Month()) - 1
	if idx >= 0 && idx < len(bulanKeys) {
		return bulanKeys[idx]
	}
	return bulanKeys[0]
}

func normalizeBulanKey(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func matchesRekeningPrefix(kode, prefix string) bool {
	kode = strings.TrimSpace(kode)
	prefix = strings.TrimSpace(prefix)
	if kode == "" || prefix == "" {
		return false
	}
	if !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}
	base := strings.TrimSuffix(prefix, ".")
	if kode == base {
		return true
	}
	return strings.HasPrefix(kode, prefix)
}

func sumRakForPrefix(rows []RakBelanjaRow, prefix, field string) float64 {
	var total float64
	for _, r := range rows {
		if !matchesRekeningPrefix(r.KodeRekening, prefix) {
			continue
		}
		switch field {
		case "anggaran":
			total += r.Anggaran
		default:
			if r.Bulan != nil {
				total += r.Bulan[field]
			}
		}
	}
	return total
}

func buildKasReport(state KasBelanjaState, bulan string) []KasReportRow {
	bulan = normalizeBulanKey(bulan)
	prev := previousBulan(bulan)
	out := make([]KasReportRow, 0, len(angkasTemplate)+1)

	for _, tpl := range angkasTemplate {
		sisa := sumPrevSisa(state, tpl.Kode, prev)
		if manual, ok := state.SisaManual[bulan][tpl.Kode]; ok {
			sisa = manual
		}
		anggaranKas := sumRakForPrefix(state.RakRows, tpl.Kode, bulan)
		realisasi := 0.0
		if state.Realisasi[bulan] != nil {
			realisasi = state.Realisasi[bulan][tpl.Kode]
		}
		sisaSD := sisa + anggaranKas - realisasi
		persen := 0.0
		if anggaranKas > 0 {
			persen = (realisasi / anggaranKas) * 100
		}
		out = append(out, KasReportRow{
			Kode:          tpl.Kode,
			Uraian:        tpl.Uraian,
			Level:         tpl.Level,
			SisaBulanLalu: sisa,
			AnggaranKas:   anggaranKas,
			Realisasi:     realisasi,
			SisaSD:        sisaSD,
			Persen:        persen,
			Editable:      true,
		})
	}

	totalAnggaranKas := sumRakForPrefix(state.RakRows, "5.", bulan)
	totalRealisasi := 0.0
	if state.Realisasi[bulan] != nil {
		for _, v := range state.Realisasi[bulan] {
			totalRealisasi += v
		}
		// use top-level only to avoid double count
		totalRealisasi = state.Realisasi[bulan]["5."]
	}
	totalSisaLalu := sumPrevSisa(state, "5.", prev)
	if manual, ok := state.SisaManual[bulan]["5."]; ok {
		totalSisaLalu = manual
	}
	totalSisaSD := totalSisaLalu + totalAnggaranKas - totalRealisasi
	totalPersen := 0.0
	if totalAnggaranKas > 0 {
		totalPersen = (totalRealisasi / totalAnggaranKas) * 100
	}
	out = append(out, KasReportRow{
		Kode:          "",
		Uraian:        "TOTAL BELANJA",
		Level:         0,
		SisaBulanLalu: totalSisaLalu,
		AnggaranKas:   totalAnggaranKas,
		Realisasi:     totalRealisasi,
		SisaSD:        totalSisaSD,
		Persen:        totalPersen,
		Editable:      true,
	})
	return out
}

func previousBulan(bulan string) string {
	for i, b := range bulanKeys {
		if b == bulan && i > 0 {
			return bulanKeys[i-1]
		}
	}
	return ""
}

func sumPrevSisa(state KasBelanjaState, kode, prevBulan string) float64 {
	if prevBulan == "" {
		return 0
	}
	rows := buildKasReport(state, prevBulan)
	for _, r := range rows {
		if r.Kode == kode || (kode == "5." && r.Uraian == "TOTAL BELANJA") {
			return r.SisaSD
		}
	}
	return 0
}

func totalPaguFromRak(rows []RakBelanjaRow) float64 {
	var total float64
	for _, r := range rows {
		total += r.Anggaran
	}
	return total
}

func handleKasBelanja(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if sess == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		bulan := normalizeBulanKey(r.URL.Query().Get("bulan"))
		if bulan == "" {
			bulan = currentBulanKey()
		}
		kasMu.RLock()
		state := kasState
		report := buildKasReport(state, bulan)
		kasMu.RUnlock()
		locked := state.RealisasiLocked != nil && state.RealisasiLocked[bulan]
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"tahun":            state.Tahun,
			"rak_rows":         state.RakRows,
			"realisasi":        state.Realisasi,
			"sisa_manual":      state.SisaManual,
			"realisasi_locked": locked,
			"imported_at":      state.ImportedAt,
			"total_pagu":       totalPaguFromRak(state.RakRows),
			"bulan":            bulan,
			"report":           report,
			"bulan_list":       bulanKeys,
			"version":          state.Version,
			"version_label":    state.VersionLabel,
		})
	default:
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func handleKasImportRAK(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	var payload struct {
		Tahun        int             `json:"tahun"`
		RakRows      []RakBelanjaRow `json:"rak_rows"`
		Version      string          `json:"version"`
		VersionLabel string          `json:"version_label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	if len(payload.RakRows) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Data RAK kosong"})
		return
	}
	version := strings.ToLower(strings.TrimSpace(payload.Version))
	if version == "" {
		version = "apbd"
	}
	versionLabel := strings.TrimSpace(payload.VersionLabel)
	if versionLabel == "" {
		versionLabel = rakVersionLabel(version)
	}
	kasMu.Lock()
	if payload.Tahun > 0 {
		kasState.Tahun = payload.Tahun
	}
	kasState.RakRows = payload.RakRows
	kasState.Version = version
	kasState.VersionLabel = versionLabel
	kasState.ImportedAt = time.Now().Format("2006-01-02 15:04:05")
	kasMu.Unlock()
	persistKasState()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":       "Data RAK " + versionLabel + " berhasil diimpor",
		"total":         len(payload.RakRows),
		"rak_rows":      payload.RakRows,
		"version":       version,
		"version_label": versionLabel,
		"imported_at":   kasState.ImportedAt,
	})
}

var kasMainLeafKodes = []string{
	"5.1.01.", "5.1.02.", "5.1.05.", "5.1.06.",
	"5.2.02.", "5.2.03.", "5.2.04.", "5.2.05.",
	"5.3.01.",
}

var kasPenDetailLeafKodes = []string{
	"5.1.02.03.002.00035",
	"5.1.02.03.002.00038",
	"5.1.02.02.001.00059",
	"5.1.02.02.001.00060",
	"5.1.02.02.001.00061",
	"5.1.02.02.001.00063",
	"5.1.02.04.001.00001",
	"5.1.02.04.01.00002",
	"5.1.02.04.001.00003",
	"5.1.02.04.001.00004",
	"5.1.02.02.005.00043",
}

func sumRealisasiKeys(m map[string]float64, keys ...string) float64 {
	var total float64
	for _, k := range keys {
		total += m[k]
	}
	return total
}

// rollupRealisasi menerapkan rumus sheet BELANJA: baris induk = jumlah anak (kolom realisasi).
func rollupRealisasi(raw map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(raw)+16)
	for _, k := range kasMainLeafKodes {
		out[k] = raw[k]
	}
	for _, k := range kasPenDetailLeafKodes {
		out[k] = raw[k]
	}
	out["5.1."] = sumRealisasiKeys(out, "5.1.01.", "5.1.02.", "5.1.05.", "5.1.06.")
	out["5.2."] = sumRealisasiKeys(out, "5.2.02.", "5.2.03.", "5.2.04.", "5.2.05.")
	out["5.3.01."] = raw["5.3.01."]
	out["5.3."] = out["5.3.01."]
	out["5."] = sumRealisasiKeys(out, "5.1.", "5.2.", "5.3.")
	return out
}

func handleKasSaveRealisasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	var payload struct {
		Bulan      string             `json:"bulan"`
		Realisasi  map[string]float64 `json:"realisasi"`
		SisaManual map[string]float64 `json:"sisa_manual"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	bulan := normalizeBulanKey(payload.Bulan)
	if bulan == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Bulan wajib diisi"})
		return
	}
	kasMu.Lock()
	if kasState.RealisasiLocked != nil && kasState.RealisasiLocked[bulan] {
		kasMu.Unlock()
		jsonResponse(w, http.StatusForbidden, map[string]string{
			"error": "Realisasi bulan ini terkunci. Klik Perbaiki terlebih dahulu.",
		})
		return
	}
	if kasState.Realisasi == nil {
		kasState.Realisasi = map[string]map[string]float64{}
	}
	if kasState.SisaManual == nil {
		kasState.SisaManual = map[string]map[string]float64{}
	}
	if kasState.RealisasiLocked == nil {
		kasState.RealisasiLocked = map[string]bool{}
	}
	if payload.Realisasi != nil {
		kasState.Realisasi[bulan] = rollupRealisasi(payload.Realisasi)
	}
	if payload.SisaManual != nil {
		kasState.SisaManual[bulan] = payload.SisaManual
	}
	kasState.RealisasiLocked[bulan] = true
	report := buildKasReport(kasState, bulan)
	kasMu.Unlock()
	persistKasState()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":          "Realisasi bulan " + bulan + " disimpan",
		"report":           report,
		"realisasi_locked": true,
	})
}

func handleKasUnlockRealisasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	var payload struct {
		Bulan string `json:"bulan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	bulan := normalizeBulanKey(payload.Bulan)
	if bulan == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Bulan wajib diisi"})
		return
	}
	kasMu.Lock()
	if kasState.RealisasiLocked == nil {
		kasState.RealisasiLocked = map[string]bool{}
	}
	delete(kasState.RealisasiLocked, bulan)
	kasMu.Unlock()
	persistKasState()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":          "Mode perbaikan aktif — data realisasi dapat diubah",
		"realisasi_locked": false,
	})
}
