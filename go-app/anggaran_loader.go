package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type RakRow struct {
	Kegiatan     string  `json:"kegiatan"`
	SubKegiatan  string  `json:"sub_kegiatan"`
	KodeRekening string  `json:"kode_rekening"`
	Pekerjaan    string  `json:"pekerjaan"`
	PPTK         string  `json:"pptk"`
	PPTKNip      string  `json:"pptk_nip,omitempty"`
	Anggaran     float64 `json:"anggaran"`
}

type ImportAnggaranResult struct {
	Rak              []RakRow           `json:"rak"`
	AnggaranKegiatan map[string]float64 `json:"anggaran_kegiatan"`
	RakMeta          RakMeta            `json:"rak_meta"`
	TotalBaris       int                `json:"total_baris"`
	Message          string             `json:"message"`
}

type RakMeta struct {
	Version       string  `json:"version"`
	Label         string  `json:"label"`
	ImportedAt    string  `json:"imported_at"`
	RowCount      int     `json:"row_count"`
	TotalAnggaran float64 `json:"total_anggaran"`
}

var rakVersionLabels = map[string]string{
	"apbd":              "APBD (Murni)",
	"pergeseran-1":      "Pergeseran APBD I",
	"pergeseran-2":      "Pergeseran APBD II",
	"pergeseran-3":      "Pergeseran APBD III",
	"pergeseran-4":      "Pergeseran APBD IV",
	"apbdp":             "APBDP (Perubahan)",
	"apbdp-pergeseran-1": "APBDP Pergeseran I",
	"apbdp-pergeseran-2": "APBDP Pergeseran II",
	"apbdp-pergeseran-3": "APBDP Pergeseran III",
}

func rakVersionLabel(key string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	if l, ok := rakVersionLabels[k]; ok {
		return l
	}
	if k != "" {
		return strings.ToUpper(k)
	}
	return "APBD (Murni)"
}

var legacyPPTK = map[string]string{
	"5.1.02.02.001.00080|Belanja Honorarium Penanggungjawaban Pengelola Keuangan": "RAMA WARNI, MM",
	"5.1.02.02.001.00067|Belanja Pembayaran Pajak, Bea, dan Perizinan":             "RAMA WARNI, MM",
	"5.1.02.01.001.00024|Belanja Alat/Bahan untuk Kegiatan Kantor-Alat Tulis Kantor": "ARIOS ZEUS SANDRY, S.KOM",
	"5.1.02.03.002.00405|Belanja Pemeliharaan Komputer-Komputer Unit-Personal Computer": "ARIOS ZEUS SANDRY, S.KOM",
	"5.2.02.05.003.00001|Belanja Modal Meja Kerja Pejabat":                          "RAMA WARNI, MM",
}

func pptkForRow(kode, pekerjaan string) string {
	key := strings.TrimSpace(kode) + "|" + strings.TrimSpace(pekerjaan)
	if p, ok := legacyPPTK[key]; ok {
		return p
	}
	return ""
}

func normalizeHeaderCell(h string) string {
	return strings.Join(strings.Fields(strings.ToUpper(strings.TrimSpace(h))), " ")
}

func findAnggaranHeaderRow(rows [][]string) int {
	limit := 20
	if len(rows) < limit {
		limit = len(rows)
	}
	for r := 0; r < limit; r++ {
		for _, h := range rows[r] {
			u := normalizeHeaderCell(h)
			if strings.Contains(u, "NAMA KEGIATAN") || u == "KEGIATAN" || strings.Contains(u, "URAIAN KEGIATAN") {
				return r
			}
		}
	}
	return 0
}

func headerIndexExact(header map[string]int, names ...string) int {
	for _, n := range names {
		if i, ok := header[normalizeHeaderCell(n)]; ok {
			return i
		}
	}
	return -1
}

func headerIndexContains(row []string, skip func(string) bool, needles ...string) int {
	for i, h := range row {
		u := normalizeHeaderCell(h)
		if u == "" || (skip != nil && skip(u)) {
			continue
		}
		for _, n := range needles {
			if strings.Contains(u, normalizeHeaderCell(n)) {
				return i
			}
		}
	}
	return -1
}

func skipNonAnggaranHeader(u string) bool {
	if strings.Contains(u, "SUB KEG") || strings.Contains(u, "KODE REK") || strings.Contains(u, "NIP") {
		return !strings.Contains(u, "ANGGARAN") && !strings.Contains(u, "PAGU") && !strings.Contains(u, "JUMLAH")
	}
	return strings.Contains(u, "NO.") || strings.Contains(u, "NOMOR") || strings.Contains(u, "TAHUN") ||
		strings.Contains(u, "BULAN") || strings.Contains(u, "SIFAT") || strings.Contains(u, "SKPD") || strings.Contains(u, "OPD")
}

func findAnggaranColumnIndex(headerRow []string, dataRows [][]string) int {
	header := make([]string, len(headerRow))
	headerMap := map[string]int{}
	for i, h := range headerRow {
		header[i] = normalizeHeaderCell(h)
		if header[i] != "" {
			headerMap[header[i]] = i
		}
	}

	candidates := map[int]bool{}
	for _, name := range []string{"ANGGARAN", "PAGU ANGGARAN", "JUMLAH ANGGARAN", "NILAI ANGGARAN", "TOTAL ANGGARAN", "JUMLAH BIAYA", "PAGU"} {
		if i := headerIndexExact(headerMap, name); i >= 0 {
			candidates[i] = true
		}
	}
	for i, h := range header {
		if h == "" {
			continue
		}
		if strings.Contains(h, "ANGGARAN") || strings.Contains(h, "PAGU") || strings.Contains(h, "JUMLAH") ||
			strings.Contains(h, "NILAI") || strings.Contains(h, "TOTAL") || strings.Contains(h, "BIAYA") ||
			strings.Contains(h, "VOLUME") || strings.Contains(h, "RPA") {
			if strings.Contains(h, "SUB KEG") || strings.Contains(h, "KODE REK") || strings.Contains(h, "NIP PPTK") ||
				strings.Contains(h, "RENCANA") || strings.Contains(h, "INDIKATOR") {
				continue
			}
			candidates[i] = true
		}
	}
	if len(candidates) == 0 {
		return -1
	}

	sample := dataRows
	if len(sample) > 40 {
		sample = sample[:40]
	}
	bestCol := -1
	bestScore := -1.0
	for col := range candidates {
		hits := 0
		total := 0.0
		for _, row := range sample {
			val := parseAnggaranValue(cellAt(row, col))
			if val > 0 {
				hits++
				total += val
			}
		}
		score := float64(hits)*10000 + total/1e6
		if strings.Contains(header[col], "ANGGARAN") || strings.Contains(header[col], "PAGU") ||
			strings.Contains(header[col], "JUMLAH") || strings.Contains(header[col], "TOTAL") {
			score += 100
		}
		if score > bestScore {
			bestScore = score
			bestCol = col
		}
	}
	return bestCol
}

func cellAt(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseAnggaranValue(raw string) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	raw = strings.TrimPrefix(strings.TrimPrefix(raw, "Rp."), "Rp")
	raw = strings.TrimSpace(raw)
	if strings.Count(raw, ".") > 1 {
		raw = strings.ReplaceAll(raw, ".", "")
		raw = strings.ReplaceAll(raw, ",", ".")
	} else if strings.Contains(raw, ",") {
		raw = strings.ReplaceAll(raw, ".", "")
		raw = strings.ReplaceAll(raw, ",", ".")
	}
	var ang float64
	fmt.Sscanf(raw, "%f", &ang)
	return ang
}

func normalizeAnggaranUnit(rows []RakRow) {
	var max float64
	for _, r := range rows {
		if r.Anggaran > max {
			max = r.Anggaran
		}
	}
	// File Anggaran.xlsx menyimpan pagu dalam juta rupiah (nilai tipikal < 100.000)
	if max > 0 && max < 100_000 {
		for i := range rows {
			rows[i].Anggaran *= 1_000_000
		}
	}
}

func applyRakImport(rows []RakRow) ImportAnggaranResult {
	return applyRakToAllModules(rows)
}

func applyRakToAllModules(rows []RakRow) ImportAnggaranResult {
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
	sipkeuModulesMu.RLock()
	mods := make([]*SipkeuModule, 0, len(sipkeuModules))
	for _, m := range sipkeuModules {
		mods = append(mods, m)
	}
	sipkeuModulesMu.RUnlock()
	var totalAnggaran float64
	for _, r := range rows {
		totalAnggaran += r.Anggaran
	}
	meta := RakMeta{
		Version:       "apbd",
		Label:         rakVersionLabel("apbd"),
		ImportedAt:    time.Now().Format(time.RFC3339),
		RowCount:      len(rows),
		TotalAnggaran: totalAnggaran,
	}
	var last ImportAnggaranResult
	for _, mod := range mods {
		last = applyRakToModule(mod, cloneRakRows(rows), &meta)
	}
	last.Message = fmt.Sprintf("%d baris kegiatan dan pagu anggaran berhasil dimuat", len(rows))
	return last
}

func loadAnggaranFromExcel(path string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return err
	}
	if len(rows) < 2 {
		return fmt.Errorf("file kosong")
	}

	headerRowIdx := findAnggaranHeaderRow(rows)
	headerRow := rows[headerRowIdx]
	dataRows := rows[headerRowIdx+1:]

	header := map[string]int{}
	for i, h := range headerRow {
		header[normalizeHeaderCell(h)] = i
	}

	idx := func(names ...string) int { return headerIndexExact(header, names...) }
	idxContains := func(names ...string) int {
		return headerIndexContains(headerRow, skipNonAnggaranHeader, names...)
	}

	iKeg := idx("NAMA KEGIATAN", "KEGIATAN", "URAIAN KEGIATAN")
	if iKeg < 0 {
		iKeg = idxContains("NAMA KEGIATAN", "KEGIATAN")
	}
	iSubKode := idx("KODE SUB KEGIATAN")
	iSubNama := idx("NAMA SUB KEGIATAN", "SUB KEGIATAN", "URAIAN SUB KEGIATAN")
	if iSubNama < 0 {
		iSubNama = idxContains("SUB KEGIATAN")
	}
	iKodeRek := idx("KODE REKENING", "KODE REK")
	if iKodeRek < 0 {
		iKodeRek = idxContains("KODE REKENING", "KODE REK")
	}
	iNamaRek := idx("NAMA REKENING", "PEKERJAAN", "URAIAN REKENING")
	if iNamaRek < 0 {
		iNamaRek = idxContains("NAMA REKENING", "PEKERJAAN", "URAIAN REKENING")
	}
	iPPTK := idx("PPTK", "NAMA PPTK")
	if iPPTK < 0 {
		iPPTK = idxContains("PPTK")
	}
	iPPTKNip := idx("NIP PPTK", "NIP")
	iAnggaran := findAnggaranColumnIndex(headerRow, dataRows)

	if iKeg < 0 {
		return fmt.Errorf("kolom wajib tidak ditemukan (NAMA KEGIATAN)")
	}
	if iAnggaran < 0 {
		return fmt.Errorf("kolom anggaran/pagu tidak ditemukan")
	}

	var rak []RakRow
	for _, row := range dataRows {
		get := func(i int) string { return cellAt(row, i) }
		keg := get(iKeg)
		if keg == "" {
			continue
		}
		sub := get(iSubNama)
		if sub == "" {
			sub = get(iSubKode)
		}
		rak = append(rak, RakRow{
			Kegiatan:     keg,
			SubKegiatan:  sub,
			KodeRekening: get(iKodeRek),
			Pekerjaan:    get(iNamaRek),
			PPTK:         get(iPPTK),
			PPTKNip:      get(iPPTKNip),
			Anggaran:     parseAnggaranValue(get(iAnggaran)),
		})
	}

	normalizeAnggaranUnit(rak)
	applyRakImport(rak)
	return nil
}

func tryLoadDefaultAnggaran() {
	sipkeuModulesMu.RLock()
	sek := sipkeuModules["sekretariat"]
	sipkeuModulesMu.RUnlock()
	if sek != nil {
		sek.mu.Lock()
		hasRak := len(sek.settings.Rak) > 0
		sek.mu.Unlock()
		if hasRak {
			return
		}
	}
	candidates := []string{
		"Anggaran.xlsx",
		filepath.Join("go-app", "Anggaran.xlsx"),
		`d:\Anggaran.xlsx`,
	}
	if env := os.Getenv("ANGGARAN_FILE"); env != "" {
		candidates = append([]string{env}, candidates...)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := loadAnggaranFromExcel(p); err == nil {
				sipkeuModulesMu.RLock()
				n := len(sipkeuModules["sekretariat"].settings.Rak)
				sipkeuModulesMu.RUnlock()
				fmt.Printf("Pagu anggaran dimuat dari %s (%d baris RAK)\n", p, n)
				return
			}
		}
	}
}

func handleImportAnggaran(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	var payload struct {
		Rak          []RakRow `json:"rak"`
		Version      string   `json:"version"`
		VersionLabel string   `json:"version_label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	if len(payload.Rak) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Data RAK kosong"})
		return
	}
	normalizeAnggaranUnit(payload.Rak)
	var totalAnggaran float64
	for _, r := range payload.Rak {
		totalAnggaran += r.Anggaran
	}
	if totalAnggaran <= 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Kolom pagu/anggaran tidak terbaca — semua nilai anggaran 0. Periksa nama kolom pagu di file Excel.",
		})
		return
	}
	version := strings.ToLower(strings.TrimSpace(payload.Version))
	if version == "" {
		version = "apbd"
	}
	label := strings.TrimSpace(payload.VersionLabel)
	if label == "" {
		label = rakVersionLabel(version)
	}
	meta := RakMeta{
		Version:       version,
		Label:         label,
		ImportedAt:    time.Now().Format(time.RFC3339),
		RowCount:      len(payload.Rak),
		TotalAnggaran: totalAnggaran,
	}
	mod := moduleFromRequest(r)
	rows := cloneRakRows(payload.Rak)
	result := applyRakToModule(mod, rows, &meta)
	result.Message = fmt.Sprintf("Anggaran %s — %d baris kegiatan berhasil dimuat (mengganti anggaran aktif sebelumnya)", label, result.TotalBaris)
	jsonResponse(w, http.StatusOK, result)
}
