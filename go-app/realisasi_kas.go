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

// computeKasMonth menghitung laporan satu bulan memakai carry-in sisa bulan lalu
// (sisaPrev). Setelah selesai, sisaPrev diperbarui menjadi sisa s/d bulan ini agar
// bisa dipakai bulan berikutnya. Kunci "5." menyimpan sisa total.
func computeKasMonth(state KasBelanjaState, bulan string, sisaPrev map[string]float64) []KasReportRow {
	out := make([]KasReportRow, 0, len(angkasTemplate)+1)
	var belanjaDaerah KasReportRow
	hasBelanjaDaerah := false
	for _, tpl := range angkasTemplate {
		sisa := sisaPrev[tpl.Kode]
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
		row := KasReportRow{
			Kode:          tpl.Kode,
			Uraian:        tpl.Uraian,
			Level:         tpl.Level,
			SisaBulanLalu: sisa,
			AnggaranKas:   anggaranKas,
			Realisasi:     realisasi,
			SisaSD:        sisaSD,
			Persen:        persen,
			Editable:      true,
		}
		out = append(out, row)
		sisaPrev[tpl.Kode] = sisaSD
		if tpl.Kode == "5." {
			belanjaDaerah = row
			hasBelanjaDaerah = true
		}
	}

	// TOTAL BELANJA identik dengan baris induk BELANJA DAERAH (5.). Mengambil
	// langsung dari baris "5." menghindari tabrakan kunci carry sisa bulan lalu
	// (dulu kunci "5." dipakai bersama sehingga sisa bulan lalu/total salah).
	total := KasReportRow{Uraian: "TOTAL BELANJA", Editable: true}
	if hasBelanjaDaerah {
		total.SisaBulanLalu = belanjaDaerah.SisaBulanLalu
		total.AnggaranKas = belanjaDaerah.AnggaranKas
		total.Realisasi = belanjaDaerah.Realisasi
		total.SisaSD = belanjaDaerah.SisaSD
		total.Persen = belanjaDaerah.Persen
	}
	out = append(out, total)
	return out
}

// buildKasReport menghitung laporan kas s/d bulan target secara iteratif
// (dari Januari) sehingga linear — menghindari rekursi eksponensial sebelumnya.
func buildKasReport(state KasBelanjaState, bulan string) []KasReportRow {
	bulan = normalizeBulanKey(bulan)
	targetIdx := 0
	for i, b := range bulanKeys {
		if b == bulan {
			targetIdx = i
			break
		}
	}
	sisaPrev := map[string]float64{}
	var rows []KasReportRow
	for m := 0; m <= targetIdx; m++ {
		rows = computeKasMonth(state, bulanKeys[m], sisaPrev)
	}
	return rows
}

type kasPeriodRange struct {
	FromIdx int
	ToIdx   int
	Label   string
	Jenis   string
}

func kasPeriodRangeForKey(periode string) (kasPeriodRange, bool) {
	periode = strings.ToLower(strings.TrimSpace(periode))
	switch periode {
	case "tw1":
		return kasPeriodRange{0, 2, "TRIWULAN I (JANUARI – MARET)", "triwulan"}, true
	case "tw2":
		return kasPeriodRange{3, 5, "TRIWULAN II (APRIL – JUNI)", "triwulan"}, true
	case "tw3":
		return kasPeriodRange{6, 8, "TRIWULAN III (JULI – SEPTEMBER)", "triwulan"}, true
	case "tw4":
		return kasPeriodRange{9, 11, "TRIWULAN IV (OKTOBER – DESEMBER)", "triwulan"}, true
	case "sem1":
		return kasPeriodRange{0, 5, "SEMESTER I (JANUARI – JUNI)", "semester"}, true
	case "sem2":
		return kasPeriodRange{6, 11, "SEMESTER II (JULI – DESEMBER)", "semester"}, true
	case "tahun":
		return kasPeriodRange{0, 11, "TAHUNAN (JANUARI – DESEMBER)", "tahun"}, true
	default:
		return kasPeriodRange{}, false
	}
}

// buildKasReportRange mengagregasi laporan kas untuk rentang bulan [fromIdx..toIdx]
// dengan menjumlahkan nilai kolom 4 & 5 dari setiap bulan, kolom 3 dari bulan pertama,
// dan kolom 6 dari sisa akhir bulan terakhir periode.
func buildKasReportRange(state KasBelanjaState, fromIdx, toIdx int) []KasReportRow {
	return buildKasPeriodResult(state, fromIdx, toIdx).Report
}

type KasMonthlyReportSnap struct {
	Bulan  string         `json:"bulan"`
	Report []KasReportRow `json:"report"`
}

type kasPeriodResult struct {
	Report         []KasReportRow
	MonthlyReports []KasMonthlyReportSnap
}

func kasRowByKode(rows []KasReportRow, kode string) (KasReportRow, bool) {
	for _, r := range rows {
		if r.Kode == kode {
			return r, true
		}
	}
	return KasReportRow{}, false
}

func buildKasPeriodResult(state KasBelanjaState, fromIdx, toIdx int) kasPeriodResult {
	if fromIdx < 0 {
		fromIdx = 0
	}
	if toIdx >= len(bulanKeys) {
		toIdx = len(bulanKeys) - 1
	}
	if fromIdx > toIdx {
		fromIdx, toIdx = toIdx, fromIdx
	}

	sisaPrev := map[string]float64{}
	for m := 0; m < fromIdx; m++ {
		computeKasMonth(state, bulanKeys[m], sisaPrev)
	}

	tempSisa := make(map[string]float64, len(sisaPrev))
	for k, v := range sisaPrev {
		tempSisa[k] = v
	}

	angSum := map[string]float64{}
	realSum := map[string]float64{}
	monthly := make([]KasMonthlyReportSnap, 0, toIdx-fromIdx+1)
	var firstRows, lastRows []KasReportRow

	for m := fromIdx; m <= toIdx; m++ {
		rows := computeKasMonth(state, bulanKeys[m], tempSisa)
		monthly = append(monthly, KasMonthlyReportSnap{
			Bulan:  bulanKeys[m],
			Report: append([]KasReportRow(nil), rows...),
		})
		if m == fromIdx {
			firstRows = rows
		}
		lastRows = rows
		for _, row := range rows {
			if row.Kode == "" {
				continue
			}
			angSum[row.Kode] += row.AnggaranKas
			realSum[row.Kode] += row.Realisasi
		}
	}

	out := make([]KasReportRow, 0, len(angkasTemplate)+1)
	var belanjaDaerah KasReportRow
	hasBelanjaDaerah := false
	for _, tpl := range angkasTemplate {
		kode := tpl.Kode
		ang := angSum[kode]
		real := realSum[kode]
		sisaLalu := 0.0
		if r, ok := kasRowByKode(firstRows, kode); ok {
			sisaLalu = r.SisaBulanLalu
		}
		sisaSD := sisaLalu + ang - real
		if r, ok := kasRowByKode(lastRows, kode); ok {
			sisaSD = r.SisaSD
		}
		persen := 0.0
		if ang > 0 {
			persen = (real / ang) * 100
		}
		row := KasReportRow{
			Kode:          kode,
			Uraian:        tpl.Uraian,
			Level:         tpl.Level,
			SisaBulanLalu: sisaLalu,
			AnggaranKas:   ang,
			Realisasi:     real,
			SisaSD:        sisaSD,
			Persen:        persen,
			Editable:      true,
		}
		out = append(out, row)
		if kode == "5." {
			belanjaDaerah = row
			hasBelanjaDaerah = true
		}
	}

	total := KasReportRow{Uraian: "TOTAL BELANJA", Editable: true}
	if hasBelanjaDaerah {
		total.SisaBulanLalu = belanjaDaerah.SisaBulanLalu
		total.AnggaranKas = belanjaDaerah.AnggaranKas
		total.Realisasi = belanjaDaerah.Realisasi
		total.SisaSD = belanjaDaerah.SisaSD
		total.Persen = belanjaDaerah.Persen
	}
	out = append(out, total)
	return kasPeriodResult{Report: out, MonthlyReports: monthly}
}

func verifyKasPeriodMonthlySum(monthly []KasMonthlyReportSnap, report []KasReportRow) bool {
	if len(monthly) == 0 {
		return true
	}
	angSum := map[string]float64{}
	realSum := map[string]float64{}
	for _, snap := range monthly {
		for _, row := range snap.Report {
			if row.Kode == "" {
				continue
			}
			angSum[row.Kode] += row.AnggaranKas
			realSum[row.Kode] += row.Realisasi
		}
	}
	const eps = 0.01
	for _, row := range report {
		if row.Kode == "" {
			continue
		}
		if absFloat(row.AnggaranKas-angSum[row.Kode]) > eps || absFloat(row.Realisasi-realSum[row.Kode]) > eps {
			return false
		}
	}
	return true
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func handleKasLaporanTahunan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	periode := normalizeBulanKey(r.URL.Query().Get("periode"))
	if periode == "" {
		periode = "tw1"
	}
	pr, ok := kasPeriodRangeForKey(periode)
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Periode tidak valid"})
		return
	}
	kasMu.RLock()
	state := kasState
	bundle := buildKasPeriodResult(state, pr.FromIdx, pr.ToIdx)
	kasMu.RUnlock()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"tahun":            state.Tahun,
		"periode":          periode,
		"periode_label":    pr.Label,
		"periode_jenis":    pr.Jenis,
		"from_bulan":       bulanKeys[pr.FromIdx],
		"to_bulan":         bulanKeys[pr.ToIdx],
		"report":           bundle.Report,
		"monthly_reports":  bundle.MonthlyReports,
		"monthly_sum_ok":   verifyKasPeriodMonthlySum(bundle.MonthlyReports, bundle.Report),
		"rak_rows":         state.RakRows,
		"realisasi":        state.Realisasi,
		"total_pagu":       totalPaguFromRak(state.RakRows),
		"version":          state.Version,
		"version_label":    state.VersionLabel,
		"bulan_list":       bulanKeys,
	})
}

func totalPaguFromRak(rows []RakBelanjaRow) float64 {
	var total float64
	for _, r := range rows {
		total += r.Anggaran
	}
	return total
}

func kasRakMergeKey(r RakBelanjaRow) string {
	return strings.TrimSpace(r.KodeRekening) + "\x00" +
		strings.TrimSpace(r.NamaKegiatan) + "\x00" +
		strings.TrimSpace(r.NamaSubKegiatan)
}

func lockedKasMonths(locked map[string]bool) []string {
	if len(locked) == 0 {
		return nil
	}
	out := make([]string, 0, len(bulanKeys))
	for _, b := range bulanKeys {
		if locked[b] {
			out = append(out, b)
		}
	}
	return out
}

func cloneRakBelanjaRows(rows []RakBelanjaRow) []RakBelanjaRow {
	out := make([]RakBelanjaRow, len(rows))
	for i, r := range rows {
		out[i] = r
		if r.Bulan != nil {
			out[i].Bulan = make(map[string]float64, len(r.Bulan))
			for k, v := range r.Bulan {
				out[i].Bulan[k] = v
			}
		}
	}
	return out
}

// mergeKasRakPreservingLockedMonths menggabungkan RAK baru dengan RAK lama:
// bulan yang realisasinya sudah terkunci mempertahankan rencana kas (kolom bulan) lama.
func mergeKasRakPreservingLockedMonths(oldRows, newRows []RakBelanjaRow, locked map[string]bool) []RakBelanjaRow {
	lockedMonths := lockedKasMonths(locked)
	if len(lockedMonths) == 0 || len(oldRows) == 0 {
		return cloneRakBelanjaRows(newRows)
	}
	oldMap := make(map[string]RakBelanjaRow, len(oldRows))
	for _, r := range oldRows {
		oldMap[kasRakMergeKey(r)] = r
	}
	out := cloneRakBelanjaRows(newRows)
	for i := range out {
		old, ok := oldMap[kasRakMergeKey(out[i])]
		if !ok || old.Bulan == nil {
			continue
		}
		if out[i].Bulan == nil {
			out[i].Bulan = map[string]float64{}
		}
		for _, bulan := range lockedMonths {
			out[i].Bulan[bulan] = old.Bulan[bulan]
		}
	}
	return out
}

func formatKasLockedMonthList(months []string) string {
	if len(months) == 0 {
		return ""
	}
	labels := make([]string, len(months))
	for i, m := range months {
		if m == "" {
			continue
		}
		labels[i] = strings.ToUpper(m[:1]) + m[1:]
	}
	return strings.Join(labels, ", ")
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
			"tahun":                   state.Tahun,
			"rak_rows":                state.RakRows,
			"realisasi":               state.Realisasi,
			"sisa_manual":             state.SisaManual,
			"realisasi_locked":        locked,
			"realisasi_locked_months": lockedKasMonths(state.RealisasiLocked),
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
	oldRows := cloneRakBelanjaRows(kasState.RakRows)
	locked := kasState.RealisasiLocked
	if payload.Tahun > 0 {
		kasState.Tahun = payload.Tahun
	}
	merged := mergeKasRakPreservingLockedMonths(oldRows, payload.RakRows, locked)
	kasState.RakRows = merged
	kasState.Version = version
	kasState.VersionLabel = versionLabel
	kasState.ImportedAt = time.Now().Format("2006-01-02 15:04:05")
	preserved := lockedKasMonths(locked)
	importedAt := kasState.ImportedAt
	kasMu.Unlock()
	persistKasState()
	msg := "Data RAK " + versionLabel + " berhasil diimpor"
	if len(preserved) > 0 {
		msg += ". Rencana kas & realisasi bulan terkunci (" + formatKasLockedMonthList(preserved) + ") tetap dipertahankan."
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":                 msg,
		"total":                   len(merged),
		"rak_rows":                merged,
		"version":                 version,
		"version_label":           versionLabel,
		"imported_at":             importedAt,
		"preserved_locked_months": preserved,
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
