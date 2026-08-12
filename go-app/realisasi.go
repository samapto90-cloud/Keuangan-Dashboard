package main

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type realisasiFilters struct {
	Tahun        string
	Kegiatan     string
	SubKegiatan  string
	KodeRekening string
	PPTK         string
	Bulan        string
	Search       string
	Page         int
	PageSize     int
	Sort         string
	SortDir      string
}

type RealisasiRow struct {
	RowKey         string  `json:"row_key"`
	Kegiatan       string  `json:"kegiatan"`
	SubKegiatan    string  `json:"sub_kegiatan"`
	KodeRekening   string  `json:"kode_rekening"`
	UraianRekening string  `json:"uraian_rekening"`
	PPTK           string  `json:"pptk"`
	Anggaran       float64 `json:"anggaran"`
	Realisasi      float64 `json:"realisasi"`
	Sisa           float64 `json:"sisa"`
	Pct            float64 `json:"pct"`
	Count          int     `json:"count"`
	Status         string  `json:"status"`
	StatusLabel    string  `json:"status_label"`
	OverBudget     bool    `json:"over_budget"`
}

type RealisasiSummary struct {
	TotalAnggaran   float64 `json:"total_anggaran"`
	TotalTransaksi  int     `json:"total_transaksi"`
	TotalRealisasi  float64 `json:"total_realisasi"`
	SisaAnggaran    float64 `json:"sisa_anggaran"`
	PersenRealisasi float64 `json:"persen_realisasi"`
	JumlahPPTK      int     `json:"jumlah_pptk"`
	OverBudgetCount int     `json:"over_budget_count"`
}

type RealisasiChartSub struct {
	SubKegiatan string  `json:"sub_kegiatan"`
	Anggaran    float64 `json:"anggaran"`
	Realisasi   float64 `json:"realisasi"`
}

type RealisasiChartMonth struct {
	Bulan     string  `json:"bulan"`
	BulanLabel string `json:"bulan_label"`
	Realisasi float64 `json:"realisasi"`
}

type RealisasiChartRekening struct {
	KodeRekening string  `json:"kode_rekening"`
	Uraian       string  `json:"uraian"`
	Realisasi    float64 `json:"realisasi"`
}

type RealisasiStatusBucket struct {
	Status string `json:"status"`
	Label  string `json:"label"`
	Count  int    `json:"count"`
}

type RealisasiPPTKRow struct {
	PPTK        string  `json:"pptk"`
	SubKegiatan string  `json:"sub_kegiatan"`
	Anggaran    float64 `json:"anggaran"`
	Realisasi   float64 `json:"realisasi"`
	Sisa        float64 `json:"sisa"`
	Pct         float64 `json:"pct"`
	Count       int     `json:"count"`
	Status      string  `json:"status"`
	StatusLabel string  `json:"status_label"`
}

type RealisasiFilterOptions struct {
	Tahun         []string            `json:"tahun"`
	Kegiatan      []string            `json:"kegiatan"`
	SubKegiatan   []string            `json:"sub_kegiatan"`
	KodeRekening  []filterKodeOption  `json:"kode_rekening"`
	PPTK          []string            `json:"pptk"`
	Bulan         []filterBulanOption `json:"bulan"`
}

type filterKodeOption struct {
	Kode  string `json:"kode"`
	Uraian string `json:"uraian"`
	Sub   string `json:"sub_kegiatan"`
}

type filterBulanOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type RealisasiReport struct {
	Summary       RealisasiSummary        `json:"summary"`
	Rows          []RealisasiRow          `json:"rows"`
	TotalRows     int                     `json:"total_rows"`
	Page          int                     `json:"page"`
	PageSize      int                     `json:"page_size"`
	ChartSub      []RealisasiChartSub     `json:"chart_sub"`
	ChartMonth    []RealisasiChartMonth   `json:"chart_month"`
	ChartRekening []RealisasiChartRekening `json:"chart_rekening"`
	StatusDistrib []RealisasiStatusBucket `json:"status_distrib"`
	PPTKRows      []RealisasiPPTKRow      `json:"pptk_rows"`
	OverBudget    []RealisasiRow          `json:"over_budget_rows"`
	FilterOptions RealisasiFilterOptions  `json:"filter_options"`
	SourceCount   int                     `json:"source_transaction_count"`
	GeneratedAt   string                  `json:"generated_at"`
	PortalLabel   string                  `json:"portal_label"`
	TahunLabel    string                  `json:"tahun_label"`
}

var realisasiStatusLabels = map[string]string{
	"belum":            "Belum Realisasi",
	"rendah":           "Rendah",
	"sedang":           "Sedang",
	"tinggi":           "Tinggi",
	"perlu_perhatian":  "Perlu Perhatian",
	"selesai":          "Selesai",
	"over_budget":      "Over Budget",
}

var realisasiMonthNames = []string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

func parseRealisasiFilters(r *http.Request) realisasiFilters {
	q := r.URL.Query()
	page, _ := strconv.Atoi(strings.TrimSpace(q.Get("page")))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(strings.TrimSpace(q.Get("page_size")))
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 200 {
		pageSize = 200
	}
	sortDir := strings.ToLower(strings.TrimSpace(q.Get("sort_dir")))
	if sortDir != "asc" {
		sortDir = "desc"
	}
	return realisasiFilters{
		Tahun:        strings.TrimSpace(q.Get("tahun")),
		Kegiatan:     strings.TrimSpace(q.Get("kegiatan")),
		SubKegiatan:  strings.TrimSpace(q.Get("sub_kegiatan")),
		KodeRekening: strings.TrimSpace(q.Get("kode_rekening")),
		PPTK:         strings.TrimSpace(q.Get("pptk")),
		Bulan:        strings.TrimSpace(q.Get("bulan")),
		Search:       strings.TrimSpace(q.Get("search")),
		Page:         page,
		PageSize:     pageSize,
		Sort:         strings.TrimSpace(q.Get("sort")),
		SortDir:      sortDir,
	}
}

func realisasiRowKey(kegiatan, sub, kode, pekerjaan, pptk string) string {
	return strings.Join([]string{
		normRekap(kegiatan), normRekap(sub), normRekap(kode),
		normRekap(pekerjaan), normRekap(pptk),
	}, "\x1e")
}

func realisasiRowKeyFromTrx(rak []RakRow, t Transaction) string {
	pptk := pptkForTransaction(rak, t)
	return realisasiRowKey(t.Kegiatan, t.SubKegiatan, t.KodeRekening, t.Pekerjaan, pptk)
}

func transactionMatchesRealisasiFilters(t Transaction, rak []RakRow, f realisasiFilters) bool {
	if !trxIsApproved(t) {
		return false
	}
	if f.Tahun != "" {
		if len(t.Tanggal) < 4 || t.Tanggal[:4] != f.Tahun {
			return false
		}
	}
	if f.Bulan != "" {
		if len(t.Tanggal) < 7 || t.Tanggal[5:7] != f.Bulan {
			return false
		}
	}
	if f.Kegiatan != "" && normRekap(t.Kegiatan) != normRekap(f.Kegiatan) {
		return false
	}
	if f.SubKegiatan != "" && normRekap(t.SubKegiatan) != normRekap(f.SubKegiatan) {
		return false
	}
	if f.KodeRekening != "" && normRekap(t.KodeRekening) != normRekap(f.KodeRekening) {
		return false
	}
	if f.PPTK != "" {
		p := pptkForTransaction(rak, t)
		if normRekap(p) != normRekap(f.PPTK) {
			return false
		}
	}
	return true
}

func rakRowMatchesRealisasiFilters(r RakRow, f realisasiFilters) bool {
	if f.Kegiatan != "" && normRekap(r.Kegiatan) != normRekap(f.Kegiatan) {
		return false
	}
	if f.SubKegiatan != "" && normRekap(r.SubKegiatan) != normRekap(f.SubKegiatan) {
		return false
	}
	if f.KodeRekening != "" && normRekap(r.KodeRekening) != normRekap(f.KodeRekening) {
		return false
	}
	if f.PPTK != "" && normRekap(r.PPTK) != normRekap(f.PPTK) {
		return false
	}
	return true
}

func computeRealisasiStatus(pct, realisasi, anggaran float64) string {
	if anggaran > 0 && realisasi > anggaran {
		return "over_budget"
	}
	if realisasi <= 0 || pct <= 0 {
		return "belum"
	}
	if pct >= 100 {
		return "selesai"
	}
	if pct >= 95 {
		return "perlu_perhatian"
	}
	if pct > 80 {
		return "tinggi"
	}
	if pct > 50 {
		return "sedang"
	}
	return "rendah"
}

func realisasiStatusLabel(status string) string {
	if l, ok := realisasiStatusLabels[status]; ok {
		return l
	}
	return status
}

func buildRealisasiFilterOptions(mod *SipkeuModule, f realisasiFilters) RealisasiFilterOptions {
	rak, _ := moduleSettingsSnapshot(mod)
	txs := moduleTransactionsCopy(mod)

	tahunSet := map[string]bool{}
	kegiatanSet := map[string]bool{}
	subSet := map[string]bool{}
	pptkSet := map[string]bool{}
	kodeMap := map[string]filterKodeOption{}

	for _, t := range txs {
		if !transactionBelongsToModule(mod, t) || !trxIsApproved(t) {
			continue
		}
		if len(t.Tanggal) >= 4 {
			tahunSet[t.Tanggal[:4]] = true
		}
		if k := normRekap(t.Kegiatan); k != "" {
			kegiatanSet[k] = true
		}
		if s := normRekap(t.SubKegiatan); s != "" {
			subSet[s] = true
		}
		if p := pptkForTransaction(rak, t); p != "" {
			pptkSet[p] = true
		}
	}
	for _, r := range rak {
		if !moduleOwnsKegiatan(mod.ID, r.Kegiatan) {
			continue
		}
		if k := normRekap(r.Kegiatan); k != "" {
			kegiatanSet[k] = true
		}
		if s := normRekap(r.SubKegiatan); s != "" {
			subSet[s] = true
		}
		if p := normRekap(r.PPTK); p != "" {
			pptkSet[p] = true
		}
		k := normRekap(r.KodeRekening)
		if k != "" {
			kodeMap[k] = filterKodeOption{
				Kode:  k,
				Uraian: normRekap(r.Pekerjaan),
				Sub:   normRekap(r.SubKegiatan),
			}
		}
	}

	opts := RealisasiFilterOptions{
		Tahun:    sortedKeys(tahunSet),
		Kegiatan: sortedKeys(kegiatanSet),
		SubKegiatan: sortedKeys(subSet),
		PPTK:     sortedKeys(pptkSet),
	}
	if len(opts.Tahun) == 0 {
		opts.Tahun = []string{"2026"}
	}
	for i, name := range realisasiMonthNames {
		opts.Bulan = append(opts.Bulan, filterBulanOption{
			Value: fmtMonth(i + 1),
			Label: name,
		})
	}
	subFilter := f.SubKegiatan
	for _, ko := range kodeMap {
		if subFilter != "" && normRekap(ko.Sub) != normRekap(subFilter) {
			continue
		}
		opts.KodeRekening = append(opts.KodeRekening, ko)
	}
	sort.Slice(opts.KodeRekening, func(i, j int) bool {
		return opts.KodeRekening[i].Kode < opts.KodeRekening[j].Kode
	})
	return opts
}

func fmtMonth(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func buildRealisasiForModule(mod *SipkeuModule, f realisasiFilters) RealisasiReport {
	rak, _ := moduleSettingsSnapshot(mod)
	txs := moduleTransactionsCopy(mod)

	rowMap := map[string]*RealisasiRow{}
	for _, r := range rak {
		if !moduleOwnsKegiatan(mod.ID, r.Kegiatan) {
			continue
		}
		if !rakRowMatchesRealisasiFilters(r, f) {
			continue
		}
		key := realisasiRowKey(r.Kegiatan, r.SubKegiatan, r.KodeRekening, r.Pekerjaan, r.PPTK)
		if rowMap[key] == nil {
			rowMap[key] = &RealisasiRow{
				RowKey:         key,
				Kegiatan:       normRekap(r.Kegiatan),
				SubKegiatan:    normRekap(r.SubKegiatan),
				KodeRekening:   normRekap(r.KodeRekening),
				UraianRekening: normRekap(r.Pekerjaan),
				PPTK:           normRekap(r.PPTK),
				Anggaran:       r.Anggaran,
			}
		}
	}

	var filteredTxs []Transaction
	for _, t := range txs {
		if !transactionBelongsToModule(mod, t) {
			continue
		}
		if transactionMatchesRealisasiFilters(t, rak, f) {
			filteredTxs = append(filteredTxs, t)
		}
	}

	for _, t := range filteredTxs {
		key := realisasiRowKeyFromTrx(rak, t)
		row := rowMap[key]
		if row == nil {
			pptk := pptkForTransaction(rak, t)
			row = &RealisasiRow{
				RowKey:         key,
				Kegiatan:       normRekap(t.Kegiatan),
				SubKegiatan:    normRekap(t.SubKegiatan),
				KodeRekening:   normRekap(t.KodeRekening),
				UraianRekening: normRekap(t.Pekerjaan),
				PPTK:           pptk,
				Anggaran:       anggaranPekerjaanForSnapshot(rak, t.Kegiatan, t.SubKegiatan, t.KodeRekening, t.Pekerjaan),
			}
			rowMap[key] = row
		}
		row.Realisasi += t.Nilai
		row.Count++
	}

	allRows := make([]RealisasiRow, 0, len(rowMap))
	search := strings.ToLower(f.Search)
	for _, row := range rowMap {
		if f.Search != "" {
			hay := strings.ToLower(strings.Join([]string{
				row.SubKegiatan, row.KodeRekening, row.UraianRekening, row.PPTK,
			}, " "))
			if !strings.Contains(hay, search) {
				continue
			}
		}
		if row.Realisasi <= 0 && row.Anggaran <= 0 {
			continue
		}
		row.Sisa = row.Anggaran - row.Realisasi
		row.Pct = rekapPct(row.Realisasi, row.Anggaran)
		row.Status = computeRealisasiStatus(row.Pct, row.Realisasi, row.Anggaran)
		row.StatusLabel = realisasiStatusLabel(row.Status)
		row.OverBudget = row.Anggaran > 0 && row.Realisasi > row.Anggaran
		allRows = append(allRows, *row)
	}

	sortRealisasiRows(allRows, f.Sort, f.SortDir)

	report := RealisasiReport{
		TotalRows:     len(allRows),
		Page:          f.Page,
		PageSize:      f.PageSize,
		SourceCount:   len(filteredTxs),
		GeneratedAt:   time.Now().Format(time.RFC3339),
		PortalLabel:   portalLabel(mod.ID),
		FilterOptions: buildRealisasiFilterOptions(mod, f),
	}
	if f.Tahun != "" {
		report.TahunLabel = f.Tahun
	} else if len(report.FilterOptions.Tahun) > 0 {
		report.TahunLabel = report.FilterOptions.Tahun[len(report.FilterOptions.Tahun)-1]
	} else {
		report.TahunLabel = "2026"
	}

	start := (f.Page - 1) * f.PageSize
	if start > len(allRows) {
		start = len(allRows)
	}
	end := start + f.PageSize
	if end > len(allRows) {
		end = len(allRows)
	}
	report.Rows = allRows[start:end]

	report.Summary = summarizeRealisasiRows(allRows, filteredTxs)
	report.ChartSub = buildRealisasiChartSub(allRows)
	report.ChartMonth = buildRealisasiChartMonth(filteredTxs, f.Tahun)
	report.ChartRekening = buildRealisasiChartRekening(allRows)
	report.StatusDistrib = buildRealisasiStatusDistrib(allRows)
	report.PPTKRows = buildRealisasiPPTKRows(allRows)
	report.OverBudget = filterOverBudgetRows(allRows)

	return report
}

func sortRealisasiRows(rows []RealisasiRow, sortField, sortDir string) {
	less := func(i, j int) bool {
		var vi, vj float64
		var si, sj string
		switch sortField {
		case "anggaran":
			vi, vj = rows[i].Anggaran, rows[j].Anggaran
		case "realisasi":
			vi, vj = rows[i].Realisasi, rows[j].Realisasi
		case "sisa":
			vi, vj = rows[i].Sisa, rows[j].Sisa
		case "pct":
			vi, vj = rows[i].Pct, rows[j].Pct
		case "sub_kegiatan":
			si, sj = rows[i].SubKegiatan, rows[j].SubKegiatan
			return si < sj
		case "kode_rekening":
			si, sj = rows[i].KodeRekening, rows[j].KodeRekening
			return si < sj
		default:
			vi, vj = rows[i].Realisasi, rows[j].Realisasi
		}
		if sortDir == "asc" {
			return vi < vj
		}
		return vi > vj
	}
	sort.Slice(rows, less)
}

func summarizeRealisasiRows(rows []RealisasiRow, txs []Transaction) RealisasiSummary {
	var sum RealisasiSummary
	pptkActive := map[string]bool{}
	for _, row := range rows {
		sum.TotalAnggaran += row.Anggaran
		sum.TotalRealisasi += row.Realisasi
		if row.OverBudget {
			sum.OverBudgetCount++
		}
		if row.PPTK != "" && row.Count > 0 {
			pptkActive[row.PPTK] = true
		}
	}
	sum.TotalTransaksi = len(txs)
	sum.SisaAnggaran = sum.TotalAnggaran - sum.TotalRealisasi
	sum.PersenRealisasi = rekapPct(sum.TotalRealisasi, sum.TotalAnggaran)
	sum.JumlahPPTK = len(pptkActive)
	return sum
}

func buildRealisasiChartSub(rows []RealisasiRow) []RealisasiChartSub {
	m := map[string]*RealisasiChartSub{}
	for _, row := range rows {
		sub := row.SubKegiatan
		if sub == "" {
			sub = "—"
		}
		if m[sub] == nil {
			m[sub] = &RealisasiChartSub{SubKegiatan: sub}
		}
		m[sub].Anggaran += row.Anggaran
		m[sub].Realisasi += row.Realisasi
	}
	out := make([]RealisasiChartSub, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Realisasi > out[j].Realisasi })
	if len(out) > 12 {
		out = out[:12]
	}
	return out
}

func buildRealisasiChartMonth(txs []Transaction, tahun string) []RealisasiChartMonth {
	m := map[string]float64{}
	for _, t := range txs {
		if len(t.Tanggal) < 7 {
			continue
		}
		bulan := t.Tanggal[:7]
		if tahun != "" && !strings.HasPrefix(bulan, tahun) {
			continue
		}
		m[bulan] += t.Nilai
	}
	out := make([]RealisasiChartMonth, 0, 12)
	prefix := tahun
	if prefix == "" {
		prefix = time.Now().Format("2006")
	}
	for i := 1; i <= 12; i++ {
		key := prefix + "-" + fmtMonth(i)
		label := realisasiMonthNames[i-1]
		if tahun != "" {
			label = realisasiMonthNames[i-1] + " " + tahun
		}
		out = append(out, RealisasiChartMonth{
			Bulan:      key,
			BulanLabel: label,
			Realisasi:  m[key],
		})
	}
	return out
}

func buildRealisasiChartRekening(rows []RealisasiRow) []RealisasiChartRekening {
	sorted := append([]RealisasiRow(nil), rows...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Realisasi > sorted[j].Realisasi })
	limit := 10
	if len(sorted) < limit {
		limit = len(sorted)
	}
	out := make([]RealisasiChartRekening, 0, limit)
	for i := 0; i < limit; i++ {
		r := sorted[i]
		if r.Realisasi <= 0 {
			break
		}
		out = append(out, RealisasiChartRekening{
			KodeRekening: r.KodeRekening,
			Uraian:       r.UraianRekening,
			Realisasi:    r.Realisasi,
		})
	}
	return out
}

func buildRealisasiStatusDistrib(rows []RealisasiRow) []RealisasiStatusBucket {
	counts := map[string]int{}
	order := []string{"belum", "rendah", "sedang", "tinggi", "perlu_perhatian", "selesai", "over_budget"}
	for _, row := range rows {
		counts[row.Status]++
	}
	out := make([]RealisasiStatusBucket, 0, len(order))
	for _, st := range order {
		if counts[st] == 0 {
			continue
		}
		out = append(out, RealisasiStatusBucket{
			Status: st,
			Label:  realisasiStatusLabel(st),
			Count:  counts[st],
		})
	}
	return out
}

func buildRealisasiPPTKRows(rows []RealisasiRow) []RealisasiPPTKRow {
	type key struct{ pptk, sub string }
	m := map[key]*RealisasiPPTKRow{}
	for _, row := range rows {
		k := key{pptk: row.PPTK, sub: row.SubKegiatan}
		if m[k] == nil {
			m[k] = &RealisasiPPTKRow{PPTK: row.PPTK, SubKegiatan: row.SubKegiatan}
		}
		m[k].Anggaran += row.Anggaran
		m[k].Realisasi += row.Realisasi
		m[k].Count += row.Count
	}
	out := make([]RealisasiPPTKRow, 0, len(m))
	for _, v := range m {
		if v.PPTK == "" && v.Realisasi <= 0 {
			continue
		}
		v.Sisa = v.Anggaran - v.Realisasi
		v.Pct = rekapPct(v.Realisasi, v.Anggaran)
		v.Status = computeRealisasiStatus(v.Pct, v.Realisasi, v.Anggaran)
		v.StatusLabel = realisasiStatusLabel(v.Status)
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Realisasi > out[j].Realisasi })
	return out
}

func filterOverBudgetRows(rows []RealisasiRow) []RealisasiRow {
	out := []RealisasiRow{}
	for _, row := range rows {
		if row.OverBudget {
			out = append(out, row)
		}
	}
	return out
}

func buildRealisasiTransactions(mod *SipkeuModule, f realisasiFilters, rowKey string) []Transaction {
	rak, _ := moduleSettingsSnapshot(mod)
	txs := moduleTransactionsCopy(mod)
	out := []Transaction{}
	for _, t := range txs {
		if !transactionBelongsToModule(mod, t) {
			continue
		}
		if !transactionMatchesRealisasiFilters(t, rak, f) {
			continue
		}
		if rowKey != "" && realisasiRowKeyFromTrx(rak, t) != rowKey {
			continue
		}
		if f.PPTK != "" && rowKey == "" {
			p := pptkForTransaction(rak, t)
			if normRekap(p) != normRekap(f.PPTK) {
				continue
			}
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Tanggal != out[j].Tanggal {
			return out[i].Tanggal > out[j].Tanggal
		}
		return out[i].ID > out[j].ID
	})
	return out
}

func handleRealisasi(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if sess == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
		return
	}
	if !sessionHasPermission(sess, "view_rekap") {
		jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses ditolak — hak operator tidak mencukupi"})
		return
	}
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	mod := moduleFromRequest(r)
	f := parseRealisasiFilters(r)
	report := cachedRealisasiReport(mod.ID, f, func() RealisasiReport {
		return buildRealisasiForModule(mod, f)
	})
	jsonResponse(w, http.StatusOK, report)
}

func handleRealisasiTransactions(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if sess == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
		return
	}
	if !sessionHasPermission(sess, "view_rekap") {
		jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses ditolak — hak operator tidak mencukupi"})
		return
	}
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	mod := moduleFromRequest(r)
	f := parseRealisasiFilters(r)
	rowKey := strings.TrimSpace(r.URL.Query().Get("row_key"))
	pptk := strings.TrimSpace(r.URL.Query().Get("pptk"))
	if pptk != "" {
		f.PPTK = pptk
	}
	txs := buildRealisasiTransactions(mod, f, rowKey)
	var total float64
	for _, t := range txs {
		total += t.Nilai
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"transactions": txs,
		"count":        len(txs),
		"total":        total,
		"generated_at": time.Now().Format(time.RFC3339),
	})
}
