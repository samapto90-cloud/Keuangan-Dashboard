package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type GajiCategoryDef struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Group string `json:"group,omitempty"`
}

var gajiCategories = []GajiCategoryDef{
	{ID: "gaji_pns", Label: "Gaji PNS", Group: "gaji"},
	{ID: "gaji_pppk", Label: "Gaji PPPK", Group: "gaji"},
	{ID: "tpp_pns", Label: "TPP PNS", Group: "tpp"},
	{ID: "tpp_pppk", Label: "TPP PPPK", Group: "tpp"},
	{ID: "tpg_pns", Label: "TPG PNS", Group: "tpg"},
	{ID: "tpg_pppk", Label: "TPG PPPK", Group: "tpg"},
	{ID: "tamsil_pns", Label: "Tamsil PNS", Group: "tamsil"},
	{ID: "tamsil_pppk", Label: "Tamsil PPPK", Group: "tamsil"},
}

var gajiCategoryLabels = map[string]string{
	"gaji_pns":    "Gaji PNS",
	"gaji_pppk":   "Gaji PPPK",
	"tpp_pns":     "TPP PNS",
	"tpp_pppk":    "TPP PPPK",
	"tpg_pns":     "TPG PNS",
	"tpg_pppk":    "TPG PPPK",
	"tamsil_pns":  "Tamsil PNS",
	"tamsil_pppk": "Tamsil PPPK",
}

type GajiMonthCell struct {
	JumlahPegawai int     `json:"jumlah_pegawai"`
	Anggaran      float64 `json:"anggaran"`
	Realisasi     float64 `json:"realisasi"`
}

type GajiTunjanganState struct {
	Tahun           int                                 `json:"tahun"`
	Pagu            map[string]float64                  `json:"pagu,omitempty"`
	Pegawai         map[string]int                      `json:"pegawai,omitempty"`
	Rekening        []GajiRekeningDef                   `json:"rekening,omitempty"`
	RekeningCells   map[string]map[string]GajiMonthCell `json:"rekening_cells,omitempty"`
	KebutuhanManual map[string]map[string]float64       `json:"kebutuhan_manual,omitempty"`
	Cells           map[string]map[string]GajiMonthCell `json:"cells"`
	RealisasiLocked map[string]bool                     `json:"realisasi_locked"`
	ImportedAt      string                              `json:"imported_at"`
	Version         string                              `json:"version,omitempty"`
	VersionLabel    string                              `json:"version_label,omitempty"`
}

type GajiCategoryMonthRow struct {
	Bulan             string  `json:"bulan"`
	LabelBulan        string  `json:"label_bulan"`
	JumlahPegawai     int     `json:"jumlah_pegawai"`
	Anggaran          float64 `json:"anggaran"`
	Realisasi         float64 `json:"realisasi"`
	Sisa              float64 `json:"sisa"`
	Persen            float64 `json:"persen"`
	KekuranganSetahun float64 `json:"kekurangan_setahun"`
	RealisasiLocked   bool    `json:"realisasi_locked"`
	IsKebutuhanPeriod bool    `json:"is_kebutuhan_period"`
}

type GajiKebutuhanRow struct {
	CategoryID     string  `json:"category_id"`
	Label          string  `json:"label"`
	Pagu           float64 `json:"pagu"`
	RealisasiSD    float64 `json:"realisasi_sd"`
	KebutuhanSisa  float64 `json:"kebutuhan_sisa"`
	KebutuhanTahun float64 `json:"kebutuhan_tahun"`
	SisaAnggaran   float64 `json:"sisa_anggaran"`
	SelisihPagu    float64 `json:"selisih_pagu"`
	Accrual25      float64 `json:"accrual_25"`
	TotalDenganAcc float64 `json:"total_dengan_accrual"`
	Pegawai        int     `json:"pegawai"`
}

type GajiRekapBulanRow struct {
	Bulan          string                   `json:"bulan"`
	LabelBulan     string                   `json:"label_bulan"`
	Categories     map[string]GajiMonthCell `json:"categories"`
	TotalAnggaran  float64                  `json:"total_anggaran"`
	TotalRealisasi float64                  `json:"total_realisasi"`
	TotalSisa      float64                  `json:"total_sisa"`
	TotalPegawai   int                      `json:"total_pegawai"`
}

type GajiCategorySummary struct {
	CategoryID     string  `json:"category_id"`
	Label          string  `json:"label"`
	RealisasiSD    float64 `json:"realisasi_sd"`
	KebutuhanSisa  float64 `json:"kebutuhan_sisa"`
	KebutuhanTahun float64 `json:"kebutuhan_tahun"`
	Pagu           float64 `json:"pagu"`
	SelisihPagu    float64 `json:"selisih_pagu"`
}

var (
	gajiState = GajiTunjanganState{
		Tahun:           2026,
		Pagu:            map[string]float64{},
		Pegawai:         map[string]int{},
		Rekening:        []GajiRekeningDef{},
		RekeningCells:   map[string]map[string]GajiMonthCell{},
		Cells:           map[string]map[string]GajiMonthCell{},
		RealisasiLocked: map[string]bool{},
	}
	gajiMu sync.RWMutex
)

func isValidGajiCategory(id string) bool {
	_, ok := gajiCategoryLabels[id]
	return ok
}

func gajiLockKey(category, period string) string {
	return category + ":" + period
}

func ensureGajiCells(state *GajiTunjanganState) {
	if state.Cells == nil {
		state.Cells = map[string]map[string]GajiMonthCell{}
	}
	if state.Pagu == nil {
		state.Pagu = map[string]float64{}
	}
	if state.Pegawai == nil {
		state.Pegawai = map[string]int{}
	}
	for _, cat := range gajiCategories {
		if state.Cells[cat.ID] == nil {
			state.Cells[cat.ID] = map[string]GajiMonthCell{}
		}
	}
}

func gajiGetCell(state GajiTunjanganState, category, period string) GajiMonthCell {
	if state.Cells == nil || state.Cells[category] == nil {
		return GajiMonthCell{}
	}
	return state.Cells[category][period]
}

func buildGajiKebutuhan(state GajiTunjanganState, reportingMonth string) []GajiKebutuhanRow {
	out := make([]GajiKebutuhanRow, 0, len(gajiCategories))
	for _, cat := range gajiCategories {
		pagu := gajiGetPagu(state, cat.ID)
		realisasiSD, kebutuhanSisa, kebutuhanTahun := gajiSplitRealisasiKebutuhan(state, cat.ID, reportingMonth)
		sisaAnggaran := pagu - realisasiSD
		selisihPagu := pagu - kebutuhanTahun
		accrual := kebutuhanSisa * 0.025
		out = append(out, GajiKebutuhanRow{
			CategoryID:     cat.ID,
			Label:          cat.Label,
			Pagu:           pagu,
			RealisasiSD:    realisasiSD,
			KebutuhanSisa:  kebutuhanSisa,
			KebutuhanTahun: kebutuhanTahun,
			SisaAnggaran:   sisaAnggaran,
			SelisihPagu:    selisihPagu,
			Accrual25:      accrual,
			TotalDenganAcc: kebutuhanTahun + accrual,
			Pegawai:        gajiPegawaiForCategory(state, cat.ID),
		})
	}
	return out
}

func buildGajiCategoryReport(state GajiTunjanganState, category, reportingMonth string) []GajiCategoryMonthRow {
	pagu := gajiGetPagu(state, category)
	periods := gajiPeriodsForCategory(category)
	endSD := gajiRealisasiSDEndIndex(category, reportingMonth)
	out := make([]GajiCategoryMonthRow, 0, len(periods))
	var totalRealisasi float64
	for i, p := range periods {
		cell := gajiGetCell(state, category, p.Key)
		totalRealisasi += cell.Realisasi
		isKebutuhan := i > endSD
		sisa := cell.Anggaran - cell.Realisasi
		persen := 0.0
		if cell.Anggaran > 0 {
			persen = (cell.Realisasi / cell.Anggaran) * 100
		}
		var kekurangan float64
		if pagu > 0 {
			kekurangan = pagu - totalRealisasi
		}
		locked := state.RealisasiLocked != nil && state.RealisasiLocked[gajiLockKey(category, p.Key)]
		pegawai := cell.JumlahPegawai
		if pegawai == 0 {
			pegawai = gajiPegawaiForCategory(state, category)
		}
		out = append(out, GajiCategoryMonthRow{
			Bulan:             p.Key,
			LabelBulan:        p.Label,
			JumlahPegawai:     pegawai,
			Anggaran:          cell.Anggaran,
			Realisasi:         cell.Realisasi,
			Sisa:              sisa,
			Persen:            persen,
			KekuranganSetahun: kekurangan,
			RealisasiLocked:   locked,
			IsKebutuhanPeriod: isKebutuhan,
		})
	}
	return out
}

func buildGajiRekap(state GajiTunjanganState) []GajiRekapBulanRow {
	periodOrder := gajiRealisasiPeriods
	out := make([]GajiRekapBulanRow, 0, len(periodOrder))
	for _, p := range periodOrder {
		row := GajiRekapBulanRow{
			Bulan:      p.Key,
			LabelBulan: p.Label,
			Categories: map[string]GajiMonthCell{},
		}
		for _, cat := range gajiCategories {
			lookup := gajiRekapLookupPeriod(cat.ID, p.Key)
			cell := GajiMonthCell{}
			if lookup != "" {
				cell = gajiGetCell(state, cat.ID, lookup)
			}
			row.Categories[cat.ID] = cell
			row.TotalAnggaran += cell.Anggaran
			row.TotalRealisasi += cell.Realisasi
			if cell.JumlahPegawai > 0 {
				row.TotalPegawai += cell.JumlahPegawai
			}
		}
		if row.TotalPegawai == 0 {
			if pns := state.Pegawai["pns"]; pns > 0 {
				row.TotalPegawai += pns
			}
			if pppk := state.Pegawai["pppk"]; pppk > 0 {
				row.TotalPegawai += pppk
			}
		}
		row.TotalSisa = row.TotalAnggaran - row.TotalRealisasi
		out = append(out, row)
	}
	return out
}

func buildGajiDashboard(state GajiTunjanganState, reportingMonth string) map[string]interface{} {
	reportingMonth = normalizeBulanKey(reportingMonth)
	if reportingMonth == "" {
		reportingMonth = currentBulanKey()
	}
	var totalPagu, totalRealisasiSD, totalKebutuhanTahun, totalSelisih float64
	pegawaiPNS := state.Pegawai["pns"]
	pegawaiPPPK := state.Pegawai["pppk"]
	summaries := make([]GajiCategorySummary, 0, len(gajiCategories))
	for _, cat := range gajiCategories {
		pagu := gajiGetPagu(state, cat.ID)
		realisasiSD, kebutuhanSisa, kebutuhanTahun := gajiSplitRealisasiKebutuhan(state, cat.ID, reportingMonth)
		totalPagu += pagu
		totalRealisasiSD += realisasiSD
		totalKebutuhanTahun += kebutuhanTahun
		totalSelisih += pagu - kebutuhanTahun
		summaries = append(summaries, GajiCategorySummary{
			CategoryID:     cat.ID,
			Label:          cat.Label,
			RealisasiSD:    realisasiSD,
			KebutuhanSisa:  kebutuhanSisa,
			KebutuhanTahun: kebutuhanTahun,
			Pagu:           pagu,
			SelisihPagu:    pagu - kebutuhanTahun,
		})
	}
	return map[string]interface{}{
		"bulan":                   reportingMonth,
		"total_pagu":              totalPagu,
		"total_realisasi_sd":      totalRealisasiSD,
		"total_kebutuhan_tahun":   totalKebutuhanTahun,
		"total_sisa_anggaran":     totalPagu - totalRealisasiSD,
		"total_selisih_pagu":      totalSelisih,
		"pegawai_pns":             pegawaiPNS,
		"pegawai_pppk":            pegawaiPPPK,
		"pegawai_total":           pegawaiPNS + pegawaiPPPK,
		"category_summaries":      summaries,
	}
}

func handleGajiTunjangan(w http.ResponseWriter, r *http.Request) {
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	reportingMonth := normalizeBulanKey(r.URL.Query().Get("bulan"))
	if reportingMonth == "" {
		reportingMonth = currentBulanKey()
	}
	kebFilters := parseGajiKebutuhanFilters(r.URL.Query().Get("keb_filter"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	grup := strings.TrimSpace(r.URL.Query().Get("grup"))
	if grup == "" && category != "" {
		grup = gajiGrupFromCategory(category)
	}

	gajiMu.RLock()
	state := gajiState
	gajiMu.RUnlock()

	periodDefs := map[string][]gajiPeriodDef{}
	for _, cat := range gajiCategories {
		periodDefs[cat.ID] = gajiPeriodsForCategory(cat.ID)
	}

	resp := map[string]interface{}{
		"tahun":         state.Tahun,
		"imported_at":   state.ImportedAt,
		"version":       state.Version,
		"version_label": state.VersionLabel,
		"bulan":         reportingMonth,
		"bulan_list":    bulanKeys,
		"categories":    gajiCategories,
		"grup_list":     gajiGrupOrder,
		"grup_labels":   gajiGrupLabels,
		"period_defs":   periodDefs,
		"pagu":          state.Pagu,
		"pegawai":       state.Pegawai,
		"rekening":      state.Rekening,
		"dashboard":     buildGajiDashboard(state, reportingMonth),
		"kebutuhan":           buildGajiKebutuhan(state, reportingMonth),
		"kebutuhan_rekening":  buildGajiKebutuhanRekening(state, reportingMonth, kebFilters),
		"keb_filter_labels":   gajiKebutuhanFilterLabels,
		"keb_filter_keys":     gajiKebutuhanFilterKeys(),
		"rekap":               buildGajiRekap(state),
	}
	if grup != "" && isValidGajiGrup(grup) {
		rows, summary := buildGajiRekeningReport(state, grup, reportingMonth)
		matrixRows, matrixSummary := buildGajiRekeningMatrix(state, grup, reportingMonth)
		resp["grup"] = grup
		resp["rekening_report"] = rows
		resp["rekening_summary"] = summary
		resp["rekening_matrix"] = matrixRows
		resp["rekening_matrix_summary"] = matrixSummary
	}
	if category != "" && isValidGajiCategory(category) {
		resp["category"] = category
		resp["category_report"] = buildGajiCategoryReport(state, category, reportingMonth)
	}
	jsonResponse(w, http.StatusOK, resp)
}

func handleGajiImportAnggaran(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	var payload struct {
		Tahun        int    `json:"tahun"`
		Version      string `json:"version"`
		VersionLabel string `json:"version_label"`
		Pagu         map[string]float64                  `json:"pagu"`
		Pegawai       map[string]int                      `json:"pegawai"`
		Rekening      []GajiRekeningDef                   `json:"rekening"`
		RekeningCells map[string]map[string]GajiMonthCell `json:"rekening_cells"`
		Cells         map[string]map[string]GajiMonthCell `json:"cells"`
		Rows         []struct {
			Category      string  `json:"category"`
			Bulan         string  `json:"bulan"`
			JumlahPegawai int     `json:"jumlah_pegawai"`
			Anggaran      float64 `json:"anggaran"`
			Realisasi     float64 `json:"realisasi"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	if len(payload.Rows) == 0 && len(payload.Cells) == 0 && len(payload.Pagu) == 0 && len(payload.Rekening) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Data anggaran kosong"})
		return
	}

	gajiMu.Lock()
	if payload.Tahun > 0 {
		gajiState.Tahun = payload.Tahun
	}
	ensureGajiCells(&gajiState)
	if payload.Pagu != nil {
		for k, v := range payload.Pagu {
			gajiState.Pagu[k] = v
		}
	}
	if payload.Pegawai != nil {
		for k, v := range payload.Pegawai {
			gajiState.Pegawai[k] = v
		}
	}
	if len(payload.Rekening) > 0 || len(payload.RekeningCells) > 0 {
		gajiMergeRekeningImport(&gajiState, payload.Rekening, payload.RekeningCells)
	}
	if payload.Cells != nil {
		for cat, periods := range payload.Cells {
			if !isValidGajiCategory(cat) {
				continue
			}
			for period, cell := range periods {
				p := normalizeGajiPeriod(cat, period)
				if p == "" {
					continue
				}
				cur := gajiState.Cells[cat][p]
				if cell.JumlahPegawai > 0 {
					cur.JumlahPegawai = cell.JumlahPegawai
				}
				if cell.Anggaran > 0 {
					cur.Anggaran = cell.Anggaran
				}
				if cell.Realisasi > 0 || payload.Version == "template-perhitungan" {
					cur.Realisasi = cell.Realisasi
				}
				gajiState.Cells[cat][p] = cur
			}
		}
	}
	for _, row := range payload.Rows {
		cat := strings.TrimSpace(row.Category)
		period := normalizeGajiPeriod(cat, row.Bulan)
		if !isValidGajiCategory(cat) || period == "" {
			continue
		}
		cell := gajiState.Cells[cat][period]
		if row.JumlahPegawai > 0 {
			cell.JumlahPegawai = row.JumlahPegawai
		}
		if row.Anggaran > 0 {
			cell.Anggaran = row.Anggaran
		}
		if row.Realisasi > 0 {
			cell.Realisasi = row.Realisasi
		}
		gajiState.Cells[cat][period] = cell
	}
	if v := strings.TrimSpace(payload.Version); v != "" {
		gajiState.Version = v
	}
	if vl := strings.TrimSpace(payload.VersionLabel); vl != "" {
		gajiState.VersionLabel = vl
	}
	gajiState.ImportedAt = time.Now().Format("2006-01-02 15:04:05")
	gajiSyncCategoryFromRekening(&gajiState)
	gajiMu.Unlock()
	persistGajiState()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "Anggaran gaji berhasil diimpor",
		"total":       len(payload.Rows),
		"imported_at": gajiState.ImportedAt,
	})
}

func handleGajiSaveRealisasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	var payload struct {
		Grup     string `json:"grup"`
		Category string `json:"category"`
		Bulan    string `json:"bulan"`
		Rows     []struct {
			Kode          string  `json:"kode"`
			Bulan         string  `json:"bulan"`
			JumlahPegawai int     `json:"jumlah_pegawai"`
			Realisasi     float64 `json:"realisasi"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	grup := strings.TrimSpace(payload.Grup)
	if grup == "" {
		grup = gajiGrupFromCategory(strings.TrimSpace(payload.Category))
	}
	bulan := normalizeBulanKey(payload.Bulan)
	if bulan == "" {
		bulan = currentBulanKey()
	}
	if !isValidGajiGrup(grup) {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Grup realisasi tidak valid"})
		return
	}
	if len(payload.Rows) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Data realisasi kosong"})
		return
	}

	gajiMu.Lock()
	ensureGajiRekening(&gajiState)
	lockKey := gajiRekeningLockKey(grup, bulan)
	if gajiState.RealisasiLocked != nil && gajiState.RealisasiLocked[lockKey] {
		gajiMu.Unlock()
		jsonResponse(w, http.StatusForbidden, map[string]string{
			"error": "Realisasi bulan ini terkunci. Klik Perbaiki terlebih dahulu.",
		})
		return
	}
	for _, row := range payload.Rows {
		kode := strings.TrimSpace(row.Kode)
		if kode == "" {
			continue
		}
		if gajiState.RekeningCells[kode] == nil {
			gajiState.RekeningCells[kode] = map[string]GajiMonthCell{}
		}
		cell := gajiState.RekeningCells[kode][bulan]
		if row.JumlahPegawai > 0 {
			cell.JumlahPegawai = row.JumlahPegawai
		}
		cell.Realisasi = row.Realisasi
		gajiState.RekeningCells[kode][bulan] = cell
	}
	if gajiState.RealisasiLocked == nil {
		gajiState.RealisasiLocked = map[string]bool{}
	}
	gajiState.RealisasiLocked[lockKey] = true
	gajiSyncCategoryFromRekening(&gajiState)
	rows, summary := buildGajiRekeningReport(gajiState, grup, bulan)
	gajiMu.Unlock()
	persistGajiState()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":          "Realisasi per rekening disimpan",
		"rekening_report":  rows,
		"rekening_summary": summary,
	})
}

func handleGajiUnlockRealisasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	var payload struct {
		Grup     string `json:"grup"`
		Category string `json:"category"`
		Bulan    string `json:"bulan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	grup := strings.TrimSpace(payload.Grup)
	if grup == "" {
		grup = gajiGrupFromCategory(strings.TrimSpace(payload.Category))
	}
	bulan := normalizeBulanKey(payload.Bulan)
	if !isValidGajiGrup(grup) || bulan == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Grup dan bulan wajib diisi"})
		return
	}
	gajiMu.Lock()
	if gajiState.RealisasiLocked == nil {
		gajiState.RealisasiLocked = map[string]bool{}
	}
	delete(gajiState.RealisasiLocked, gajiRekeningLockKey(grup, bulan))
	gajiMu.Unlock()
	persistGajiState()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Mode perbaiki aktif — data realisasi dapat diubah",
	})
}

func handleGajiSavePegawai(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	var payload struct {
		PNS  int `json:"pns"`
		PPPK int `json:"pppk"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	if payload.PNS <= 0 && payload.PPPK <= 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Jumlah pegawai PNS atau PPPK wajib diisi"})
		return
	}
	gajiMu.Lock()
	if gajiState.Pegawai == nil {
		gajiState.Pegawai = map[string]int{}
	}
	if payload.PNS > 0 {
		gajiState.Pegawai["pns"] = payload.PNS
	}
	if payload.PPPK > 0 {
		gajiState.Pegawai["pppk"] = payload.PPPK
	}
	gajiMu.Unlock()
	persistGajiState()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Data pegawai berhasil disimpan",
		"pegawai": gajiState.Pegawai,
	})
}
