package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	TotalBaris       int                `json:"total_baris"`
	Message          string             `json:"message"`
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

func parseAnggaranValue(raw string) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
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
	for _, mod := range mods {
		applyRakToModule(mod, cloneRakRows(rows))
	}
	return ImportAnggaranResult{
		Rak:              rows,
		AnggaranKegiatan: anggaranMap,
		TotalBaris:       len(rows),
		Message:          fmt.Sprintf("%d baris kegiatan dan pagu anggaran berhasil dimuat", len(rows)),
	}
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

	header := map[string]int{}
	for i, h := range rows[0] {
		header[strings.ToUpper(strings.TrimSpace(h))] = i
	}

	idx := func(names ...string) int {
		for _, n := range names {
			if i, ok := header[strings.ToUpper(n)]; ok {
				return i
			}
		}
		return -1
	}

	iKeg := idx("NAMA KEGIATAN")
	iSubKode := idx("KODE SUB KEGIATAN")
	iSubNama := idx("NAMA SUB KEGIATAN")
	iKodeRek := idx("KODE REKENING")
	iNamaRek := idx("NAMA REKENING", "PEKERJAAN")
	iPPTK := idx("PPTK", "NAMA PPTK")
	iPPTKNip := idx("NIP PPTK")
	iAnggaran := idx("ANGGARAN", "PAGU")

	if iKeg < 0 || iAnggaran < 0 {
		return fmt.Errorf("kolom wajib tidak ditemukan (NAMA KEGIATAN, ANGGARAN)")
	}

	var rak []RakRow
	for _, row := range rows[1:] {
		get := func(i int) string {
			if i < 0 || i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}
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
		Rak []RakRow `json:"rak"`
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
	mod := moduleFromRequest(r)
	rows := cloneRakRows(payload.Rak)
	result := applyRakToModule(mod, rows)
	result.Message = fmt.Sprintf("%d baris kegiatan dan pagu anggaran berhasil dimuat", result.TotalBaris)
	jsonResponse(w, http.StatusOK, result)
}
