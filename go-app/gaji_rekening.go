package main

import (
	"sort"
	"strings"
)

type GajiRekeningDef struct {
	Kode        string  `json:"kode"`
	Nama        string  `json:"nama"`
	Kegiatan    string  `json:"kegiatan,omitempty"`
	SubKegiatan string  `json:"sub_kegiatan,omitempty"`
	Grup        string  `json:"grup"`
	Jenis       string  `json:"jenis"`
	Potongan    bool    `json:"potongan,omitempty"`
	Pagu        float64 `json:"pagu"`
	Urut        int     `json:"urut"`
}

type GajiRekeningRow struct {
	Kode          string  `json:"kode"`
	Nama          string  `json:"nama"`
	Grup          string  `json:"grup"`
	Jenis         string  `json:"jenis"`
	Potongan      bool    `json:"potongan,omitempty"`
	Attached      bool    `json:"attached,omitempty"`
	Pagu          float64 `json:"pagu"`
	JumlahPegawai int     `json:"jumlah_pegawai"`
	Anggaran      float64 `json:"anggaran"`
	Realisasi     float64 `json:"realisasi"`
	Sisa          float64 `json:"sisa"`
	Persen        float64 `json:"persen"`
	Locked        bool    `json:"locked"`
}

type GajiRekeningSummary struct {
	Grup           string  `json:"grup"`
	Label          string  `json:"label"`
	Bulan          string  `json:"bulan"`
	LabelBulan     string  `json:"label_bulan"`
	TotalPegawai   int     `json:"total_pegawai"`
	TotalAnggaran  float64 `json:"total_anggaran"`
	TotalRealisasi float64 `json:"total_realisasi"`
	TotalSisa      float64 `json:"total_sisa"`
	Locked         bool    `json:"locked"`
}

type GajiRekeningMatrixCell struct {
	Pegawai   int     `json:"pegawai"`
	Anggaran  float64 `json:"anggaran"`
	Realisasi float64 `json:"realisasi"`
	Sisa      float64 `json:"sisa"`
}

type GajiRekeningMatrixRow struct {
	Kode          string                            `json:"kode"`
	Nama          string                            `json:"nama"`
	Grup          string                            `json:"grup"`
	ViewGrup      string                            `json:"view_grup,omitempty"`
	Jenis         string                            `json:"jenis"`
	Potongan      bool                              `json:"potongan,omitempty"`
	Attached      bool                              `json:"attached,omitempty"`
	Pagu          float64                           `json:"pagu"`
	RealisasiSD   float64                           `json:"realisasi_sd"`
	KebutuhanSisa float64                           `json:"kebutuhan_sisa"`
	Bulan         map[string]GajiRekeningMatrixCell `json:"bulan"`
}

type GajiRekeningMatrixSummary struct {
	Grup    string                           `json:"grup"`
	Label   string                           `json:"label"`
	Bulan   map[string]GajiRekeningMatrixCell `json:"bulan"`
}

var gajiGrupLabels = map[string]string{
	"gaji":   "Realisasi Gaji",
	"tpp":    "Realisasi TPP",
	"tpg":    "TPG",
	"tamsil": "Tamsil",
}

var gajiGrupOrder = []string{"gaji", "tpp", "tpg", "tamsil"}

func isValidGajiGrup(grup string) bool {
	_, ok := gajiGrupLabels[grup]
	return ok
}

func parseGajiGrupList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	seen := map[string]bool{}
	var picked []string
	for _, part := range strings.Split(raw, ",") {
		g := strings.TrimSpace(part)
		if g == "" || !isValidGajiGrup(g) || seen[g] {
			continue
		}
		seen[g] = true
		picked = append(picked, g)
	}
	if len(picked) == 0 {
		return nil
	}
	out := make([]string, 0, len(picked))
	for _, g := range gajiGrupOrder {
		if seen[g] {
			out = append(out, g)
		}
	}
	return out
}

func gajiJenisFromNama(nama string) string {
	n := strings.ToUpper(nama)
	if strings.Contains(n, "PPPK") {
		return "pppk"
	}
	if strings.Contains(n, "PNS") {
		return "pns"
	}
	return ""
}

// gajiJenisFromKodeJaminanKes — suffix RAK jaminan kesehatan (Templet Anggaran Gaji).
func gajiJenisFromKodeJaminanKes(kode string) string {
	k := strings.TrimSpace(kode)
	if strings.HasSuffix(k, ".00002") || strings.HasSuffix(k, "00002") {
		return "pppk"
	}
	if strings.HasSuffix(k, ".00001") || strings.HasSuffix(k, "00001") {
		return "pns"
	}
	return ""
}

func isGajiPotonganRekening(kode, nama string) bool {
	k := strings.TrimSpace(kode)
	n := strings.ToUpper(strings.TrimSpace(nama))
	return strings.Contains(k, ".007.") || strings.Contains(k, ".009.") ||
		strings.Contains(k, ".010.") || strings.Contains(k, ".011.") || strings.Contains(k, ".012.") ||
		strings.Contains(n, "PPh") || strings.Contains(n, "IURAN JAMINAN") ||
		strings.Contains(n, "IURAN SIMPANAN")
}

func isJaminanKesehatanPotongan(nama string) bool {
	n := strings.ToUpper(strings.TrimSpace(nama))
	return strings.Contains(n, "JAMINAN KESEHATAN") ||
		strings.Contains(n, "BPJS KESEHATAN") ||
		(strings.Contains(n, "IURAN JAMINAN") && strings.Contains(n, "KESEHATAN"))
}

func potonganTargetGrupFromNama(nama string) string {
	n := strings.ToUpper(strings.TrimSpace(nama))
	if strings.Contains(n, "TPP") || strings.Contains(n, "TUNJANGAN PENGHASILAN") {
		return "tpp"
	}
	if strings.Contains(n, "TAMSIL") {
		return "tamsil"
	}
	if strings.Contains(n, "TPG") || strings.Contains(n, "TUNJANGAN PROFESI") {
		return "tpg"
	}
	return "gaji"
}

func gajiRekeningAttachedJaminanKes(def GajiRekeningDef, grup string) bool {
	if def.Grup == grup || !def.Potongan || !isJaminanKesehatanPotongan(def.Nama) {
		return false
	}
	switch grup {
	case "tpg", "tamsil":
		jenis := def.Jenis
		if jenis == "" {
			jenis = gajiJenisFromKodeJaminanKes(def.Kode)
		}
		return jenis == "pns" || jenis == "pppk" || jenis == ""
	default:
		return false
	}
}

func gajiRekeningIncludedInGrup(def GajiRekeningDef, grup string) bool {
	if def.Grup == grup {
		return true
	}
	return gajiRekeningAttachedJaminanKes(def, grup)
}

func gajiSortRekeningReportRows(rows []GajiRekeningRow, grup string) {
	jenisOrder := func(j string) int {
		if j == "pppk" {
			return 1
		}
		return 0
	}
	sort.Slice(rows, func(i, j int) bool {
		if grup == "tpg" || grup == "tamsil" {
			ji, jj := jenisOrder(rows[i].Jenis), jenisOrder(rows[j].Jenis)
			if ji != jj {
				return ji < jj
			}
			if rows[i].Potongan != rows[j].Potongan {
				return !rows[i].Potongan
			}
		}
		return rows[i].Kode < rows[j].Kode
	})
}

func classifyGajiRekening(kode, nama string) (grup, jenis string, potongan bool) {
	k := strings.TrimSpace(kode)
	n := strings.ToUpper(strings.TrimSpace(nama))
	jenis = gajiJenisFromNama(nama)
	if jenis == "" {
		jenis = gajiJenisFromKodeJaminanKes(k)
	}

	if isGajiPotonganRekening(k, n) {
		potongan = true
		target := potonganTargetGrupFromNama(n)
		if target == "tpg" || target == "tamsil" {
			if !isJaminanKesehatanPotongan(n) {
				if strings.Contains(n, "TPP") {
					target = "tpp"
				} else {
					target = "gaji"
				}
			}
		}
		return target, jenis, true
	}

	if strings.Contains(k, "5.1.01.02.006") || strings.Contains(n, "TPG") || strings.Contains(n, "TAMSIL") {
		if strings.Contains(n, "TAMSIL") {
			return "tamsil", jenis, false
		}
		return "tpg", jenis, false
	}
	if strings.HasPrefix(k, "5.1.01.02.") {
		return "tpp", jenis, false
	}
	return "gaji", jenis, false
}

func gajiRekeningLockKey(grup, bulan string) string {
	return "rekening:" + grup + ":" + normalizeBulanKey(bulan)
}

func gajiRekeningRowLockKey(grup, kode, bulan string) string {
	return "rekening:" + grup + ":" + strings.TrimSpace(kode) + ":" + normalizeBulanKey(bulan)
}

func isGajiRekeningRowLocked(state GajiTunjanganState, grup, kode, bulan string) bool {
	if state.RealisasiLocked == nil {
		return false
	}
	bulan = normalizeBulanKey(bulan)
	if state.RealisasiLocked[gajiRekeningRowLockKey(grup, kode, bulan)] {
		return true
	}
	return state.RealisasiLocked[gajiRekeningLockKey(grup, bulan)]
}

func unlockGajiRekeningMonth(state *GajiTunjanganState, grup, bulan string) {
	if state.RealisasiLocked == nil {
		return
	}
	bulan = normalizeBulanKey(bulan)
	delete(state.RealisasiLocked, gajiRekeningLockKey(grup, bulan))
	prefix := "rekening:" + grup + ":"
	suffix := ":" + bulan
	for k := range state.RealisasiLocked {
		if strings.HasPrefix(k, prefix) && strings.HasSuffix(k, suffix) && k != gajiRekeningLockKey(grup, bulan) {
			delete(state.RealisasiLocked, k)
		}
	}
}

func ensureGajiRekening(state *GajiTunjanganState) {
	if state.RekeningCells == nil {
		state.RekeningCells = map[string]map[string]GajiMonthCell{}
	}
}

func gajiFindRekeningDef(state GajiTunjanganState, kode string) *GajiRekeningDef {
	for i := range state.Rekening {
		if state.Rekening[i].Kode == kode {
			return &state.Rekening[i]
		}
	}
	return nil
}

func gajiGetRekeningCell(state GajiTunjanganState, kode, bulan string) GajiMonthCell {
	bulan = normalizeBulanKey(bulan)
	if state.RekeningCells == nil || state.RekeningCells[kode] == nil {
		return GajiMonthCell{}
	}
	return state.RekeningCells[kode][bulan]
}

// gajiRekeningCellStoreKey — kode rekening yang ditampilkan lintas menu (mis. jaminan kesehatan di TPG/Tamsil)
// disimpan terpisah per menu agar realisasi tidak saling timpa.
func gajiRekeningCellStoreKey(viewGrup string, def GajiRekeningDef) string {
	if def.Grup != viewGrup {
		return viewGrup + "|" + def.Kode
	}
	return def.Kode
}

func gajiGetRekeningCellForGrup(state GajiTunjanganState, viewGrup string, def GajiRekeningDef, bulan string) GajiMonthCell {
	bulan = normalizeBulanKey(bulan)
	native := gajiGetRekeningCell(state, def.Kode, bulan)
	cell := GajiMonthCell{
		Anggaran:      native.Anggaran,
		JumlahPegawai: native.JumlahPegawai,
	}
	storeKey := gajiRekeningCellStoreKey(viewGrup, def)
	if storeKey != def.Kode {
		if state.RekeningCells != nil && state.RekeningCells[storeKey] != nil {
			cell.Realisasi = state.RekeningCells[storeKey][bulan].Realisasi
		}
		return cell
	}
	cell.Realisasi = native.Realisasi
	if cell.Realisasi == 0 && storeKey == def.Kode {
		scopedKey := viewGrup + "|" + def.Kode
		if state.RekeningCells != nil && state.RekeningCells[scopedKey] != nil {
			cell.Realisasi = state.RekeningCells[scopedKey][bulan].Realisasi
		}
	}
	return cell
}

func gajiRekeningIsSharedAcrossMenus(def GajiRekeningDef) bool {
	for _, grup := range gajiGrupOrder {
		if def.Grup != grup && gajiRekeningAttachedJaminanKes(def, grup) {
			return true
		}
	}
	return false
}

func gajiGrupsIncludingRekening(def GajiRekeningDef) []string {
	var grups []string
	for _, g := range gajiGrupOrder {
		if gajiRekeningIncludedInGrup(def, g) {
			grups = append(grups, g)
		}
	}
	return grups
}

// gajiSumRekeningRealisasiAllGrups — jumlah realisasi semua menu untuk satu kode rekening (anggaran shared).
func gajiSumRekeningRealisasiAllGrups(state GajiTunjanganState, def GajiRekeningDef, bulan string) float64 {
	bulan = normalizeBulanKey(bulan)
	if !gajiRekeningIsSharedAcrossMenus(def) {
		return gajiGetRekeningCell(state, def.Kode, bulan).Realisasi
	}
	var total float64
	hasScoped := false
	for _, grup := range gajiGrupsIncludingRekening(def) {
		if def.Grup == grup {
			continue
		}
		storeKey := gajiRekeningCellStoreKey(grup, def)
		if state.RekeningCells == nil || state.RekeningCells[storeKey] == nil {
			continue
		}
		r := state.RekeningCells[storeKey][bulan].Realisasi
		if r != 0 {
			hasScoped = true
		}
		total += r
	}
	if !hasScoped {
		return gajiGetRekeningCell(state, def.Kode, bulan).Realisasi
	}
	return total
}

func gajiSetRekeningCellForGrup(state *GajiTunjanganState, viewGrup string, def *GajiRekeningDef, kode, bulan string, cell GajiMonthCell) {
	ensureGajiRekening(state)
	bulan = normalizeBulanKey(bulan)
	storeKey := kode
	if def != nil {
		storeKey = gajiRekeningCellStoreKey(viewGrup, *def)
	}
	if state.RekeningCells[storeKey] == nil {
		state.RekeningCells[storeKey] = map[string]GajiMonthCell{}
	}
	if def != nil && storeKey != def.Kode {
		prev := state.RekeningCells[storeKey][bulan]
		prev.Realisasi = cell.Realisasi
		if cell.JumlahPegawai > 0 {
			prev.JumlahPegawai = cell.JumlahPegawai
		}
		state.RekeningCells[storeKey][bulan] = prev
		return
	}
	state.RekeningCells[storeKey][bulan] = cell
}

func migrateGajiAttachedRekeningCells(state *GajiTunjanganState) {
	// Tidak menduplikasi realisasi ke tiap menu — data legacy tetap di kode asli
	// sampai pengguna menyimpan per menu; akumulasi sisa membaca semua menu.
	_ = state
}

func gajiUpsertRekeningDef(state *GajiTunjanganState, kode, nama string, pagu float64, urut int) *GajiRekeningDef {
	grup, jenis, potongan := classifyGajiRekening(kode, nama)
	for i := range state.Rekening {
		if state.Rekening[i].Kode == kode {
			if nama != "" {
				state.Rekening[i].Nama = nama
			}
			if pagu > 0 {
				state.Rekening[i].Pagu = pagu
			}
			state.Rekening[i].Grup = grup
			state.Rekening[i].Potongan = potongan
			if jenis != "" {
				state.Rekening[i].Jenis = jenis
			}
			return &state.Rekening[i]
		}
	}
	def := GajiRekeningDef{Kode: kode, Nama: nama, Grup: grup, Jenis: jenis, Potongan: potongan, Pagu: pagu, Urut: urut}
	if urut == 0 {
		def.Urut = len(state.Rekening) + 1
	}
	state.Rekening = append(state.Rekening, def)
	return &state.Rekening[len(state.Rekening)-1]
}

func gajiMergeRekeningImport(state *GajiTunjanganState, lines []GajiRekeningDef, cells map[string]map[string]GajiMonthCell) {
	ensureGajiRekening(state)
	for _, line := range lines {
		def := gajiUpsertRekeningDef(state, line.Kode, line.Nama, line.Pagu, line.Urut)
		if def.Urut == 0 && line.Urut > 0 {
			def.Urut = line.Urut
		}
		if v := strings.TrimSpace(line.Kegiatan); v != "" {
			def.Kegiatan = v
		}
		if v := strings.TrimSpace(line.SubKegiatan); v != "" {
			def.SubKegiatan = v
		}
	}
	for kode, months := range cells {
		if state.RekeningCells[kode] == nil {
			state.RekeningCells[kode] = map[string]GajiMonthCell{}
		}
		for bulan, cell := range months {
			b := normalizeBulanKey(bulan)
			if b == "" {
				continue
			}
			cur := state.RekeningCells[kode][b]
			if cell.Anggaran > 0 {
				cur.Anggaran = cell.Anggaran
			}
			if cell.Realisasi > 0 {
				cur.Realisasi = cell.Realisasi
			}
			if cell.JumlahPegawai > 0 {
				cur.JumlahPegawai = cell.JumlahPegawai
			}
			state.RekeningCells[kode][b] = cur
		}
	}
	sort.Slice(state.Rekening, func(i, j int) bool {
		if state.Rekening[i].Grup != state.Rekening[j].Grup {
			return gajiGrupIndex(state.Rekening[i].Grup) < gajiGrupIndex(state.Rekening[j].Grup)
		}
		if state.Rekening[i].Potongan != state.Rekening[j].Potongan {
			return !state.Rekening[i].Potongan
		}
		return state.Rekening[i].Urut < state.Rekening[j].Urut
	})
	gajiSyncCategoryFromRekening(state)
}

func normalizeGajiRekeningGrups(state *GajiTunjanganState) {
	for i := range state.Rekening {
		grup, jenis, potongan := classifyGajiRekening(state.Rekening[i].Kode, state.Rekening[i].Nama)
		state.Rekening[i].Grup = grup
		state.Rekening[i].Potongan = potongan
		if jenis != "" {
			state.Rekening[i].Jenis = jenis
		}
	}
	delete(state.Pagu, "potongan")
	delete(state.Cells, "potongan")
	if state.RealisasiLocked != nil {
		for k := range state.RealisasiLocked {
			if strings.HasPrefix(k, "rekening:potongan:") {
				delete(state.RealisasiLocked, k)
			}
		}
	}
	sort.Slice(state.Rekening, func(i, j int) bool {
		if state.Rekening[i].Grup != state.Rekening[j].Grup {
			return gajiGrupIndex(state.Rekening[i].Grup) < gajiGrupIndex(state.Rekening[j].Grup)
		}
		if state.Rekening[i].Potongan != state.Rekening[j].Potongan {
			return !state.Rekening[i].Potongan
		}
		return state.Rekening[i].Urut < state.Rekening[j].Urut
	})
}

func gajiGrupIndex(grup string) int {
	for i, g := range gajiGrupOrder {
		if g == grup {
			return i
		}
	}
	return 99
}

func buildGajiRekeningReport(state GajiTunjanganState, grup, bulan string) ([]GajiRekeningRow, GajiRekeningSummary) {
	bulan = normalizeBulanKey(bulan)
	summary := GajiRekeningSummary{
		Grup:       grup,
		Label:      gajiGrupLabels[grup],
		Bulan:      bulan,
		LabelBulan: bulanLabels[bulan],
		Locked:     false,
	}
	var rows []GajiRekeningRow
	anyLocked := false
	for _, def := range state.Rekening {
		if !gajiRekeningIncludedInGrup(def, grup) {
			continue
		}
		cell := gajiGetRekeningCellForGrup(state, grup, def, bulan)
		realisasiMenu := cell.Realisasi
		realisasiTotal := gajiSumRekeningRealisasiAllGrups(state, def, bulan)
		sisa := cell.Anggaran - realisasiTotal
		persen := 0.0
		if cell.Anggaran > 0 {
			persen = (realisasiTotal / cell.Anggaran) * 100
		}
		rowLocked := isGajiRekeningRowLocked(state, grup, def.Kode, bulan)
		if rowLocked {
			anyLocked = true
		}
		attached := def.Grup != grup && gajiRekeningAttachedJaminanKes(def, grup)
		rows = append(rows, GajiRekeningRow{
			Kode:          def.Kode,
			Nama:          def.Nama,
			Grup:          def.Grup,
			Jenis:         def.Jenis,
			Potongan:      def.Potongan,
			Attached:      attached,
			Pagu:          def.Pagu,
			Anggaran:      cell.Anggaran,
			Realisasi:     realisasiMenu,
			Sisa:          sisa,
			Persen:        persen,
			Locked:        rowLocked,
		})
		summary.TotalAnggaran += cell.Anggaran
		summary.TotalRealisasi += realisasiMenu
		summary.TotalSisa += sisa
	}
	summary.Locked = anyLocked
	summary.TotalPegawai = state.Pegawai["pns"] + state.Pegawai["pppk"]
	gajiSortRekeningReportRows(rows, grup)
	return rows, summary
}

func gajiRekeningPegawaiFallback(state GajiTunjanganState, def GajiRekeningDef, cell GajiMonthCell) int {
	if def.Jenis == "pppk" {
		return state.Pegawai["pppk"]
	}
	return state.Pegawai["pns"]
}

func buildGajiRekeningMatrix(state GajiTunjanganState, grup, sdBulan string) ([]GajiRekeningMatrixRow, GajiRekeningMatrixSummary) {
	sdBulan = normalizeBulanKey(sdBulan)
	endIdx := gajiBulanIndex(sdBulan)
	summary := GajiRekeningMatrixSummary{
		Grup:  grup,
		Label: gajiGrupLabels[grup],
		Bulan: map[string]GajiRekeningMatrixCell{},
	}
	for _, b := range bulanKeys {
		summary.Bulan[b] = GajiRekeningMatrixCell{}
	}

	var rows []GajiRekeningMatrixRow
	for _, def := range state.Rekening {
		if !gajiRekeningIncludedInGrup(def, grup) {
			continue
		}
		attached := def.Grup != grup && gajiRekeningAttachedJaminanKes(def, grup)
		row := GajiRekeningMatrixRow{
			Kode:     def.Kode,
			Nama:     def.Nama,
			Grup:     def.Grup,
			Jenis:    def.Jenis,
			Potongan: def.Potongan,
			Attached: attached,
			Pagu:     def.Pagu,
			Bulan:    map[string]GajiRekeningMatrixCell{},
		}
		for i, b := range bulanKeys {
			cell := gajiGetRekeningCellForGrup(state, grup, def, b)
			realisasiTotal := gajiSumRekeningRealisasiAllGrups(state, def, b)
			pegawai := gajiRekeningPegawaiFallback(state, def, cell)
			sisa := cell.Anggaran - realisasiTotal
			mc := GajiRekeningMatrixCell{
				Pegawai:   pegawai,
				Anggaran:  cell.Anggaran,
				Realisasi: cell.Realisasi,
				Sisa:      sisa,
			}
			row.Bulan[b] = mc
			sum := summary.Bulan[b]
			sum.Pegawai += pegawai
			sum.Anggaran += cell.Anggaran
			sum.Realisasi += cell.Realisasi
			sum.Sisa += sisa
			summary.Bulan[b] = sum
			if endIdx >= 0 && i <= endIdx {
				row.RealisasiSD += cell.Realisasi
			}
		}
		row.ViewGrup = grup
		row.KebutuhanSisa = row.Pagu - row.RealisasiSD
		if row.KebutuhanSisa < 0 {
			row.KebutuhanSisa = 0
		}
		rows = append(rows, row)
	}
	gajiSortRekeningMatrixRows(rows, grup)
	return rows, summary
}

func buildGajiRekeningMatrixMulti(state GajiTunjanganState, grups []string, sdBulan string) ([]GajiRekeningMatrixRow, GajiRekeningMatrixSummary) {
	summary := GajiRekeningMatrixSummary{
		Grup:  "multi",
		Label: "Gabungan",
		Bulan: map[string]GajiRekeningMatrixCell{},
	}
	for _, b := range bulanKeys {
		summary.Bulan[b] = GajiRekeningMatrixCell{}
	}
	var allRows []GajiRekeningMatrixRow
	labels := make([]string, 0, len(grups))
	for _, grup := range grups {
		rows, part := buildGajiRekeningMatrix(state, grup, sdBulan)
		allRows = append(allRows, rows...)
		if lbl, ok := gajiGrupLabels[grup]; ok {
			labels = append(labels, lbl)
		}
		for b, cell := range part.Bulan {
			sum := summary.Bulan[b]
			sum.Pegawai += cell.Pegawai
			sum.Anggaran += cell.Anggaran
			sum.Realisasi += cell.Realisasi
			sum.Sisa += cell.Sisa
			summary.Bulan[b] = sum
		}
	}
	summary.Label = strings.Join(labels, " + ")
	return allRows, summary
}

func gajiSortRekeningMatrixRows(rows []GajiRekeningMatrixRow, grup string) {
	jenisOrder := func(j string) int {
		if j == "pppk" {
			return 1
		}
		return 0
	}
	sort.Slice(rows, func(i, j int) bool {
		if grup == "tpg" || grup == "tamsil" {
			ji, jj := jenisOrder(rows[i].Jenis), jenisOrder(rows[j].Jenis)
			if ji != jj {
				return ji < jj
			}
			if rows[i].Potongan != rows[j].Potongan {
				return !rows[i].Potongan
			}
		}
		return rows[i].Kode < rows[j].Kode
	})
}

func gajiCategoryFromRekening(def GajiRekeningDef) string {
	switch def.Grup {
	case "gaji":
		if def.Jenis == "pppk" {
			return "gaji_pppk"
		}
		return "gaji_pns"
	case "tpp":
		if def.Jenis == "pppk" {
			return "tpp_pppk"
		}
		return "tpp_pns"
	case "tpg":
		if def.Jenis == "pppk" {
			return "tpg_pppk"
		}
		return "tpg_pns"
	case "tamsil":
		if def.Jenis == "pppk" {
			return "tamsil_pppk"
		}
		return "tamsil_pns"
	default:
		return ""
	}
}

func gajiCategoryForGrup(def GajiRekeningDef, viewGrup string) string {
	d := def
	d.Grup = viewGrup
	return gajiCategoryFromRekening(d)
}

func gajiSyncCategoryFromRekening(state *GajiTunjanganState) {
	ensureGajiCells(state)
	ensureGajiRekening(state)

	catPagu := map[string]float64{}
	catAng := map[string]map[string]float64{}
	catReal := map[string]map[string]float64{}
	catPeg := map[string]map[string]int{}

	ensureCatMaps := func(catID string) {
		if catAng[catID] == nil {
			catAng[catID] = map[string]float64{}
			catReal[catID] = map[string]float64{}
			catPeg[catID] = map[string]int{}
		}
	}

	// Rekening induk (non-potongan): anggaran + realisasi menu asal
	for _, def := range state.Rekening {
		if def.Potongan {
			continue
		}
		catID := gajiCategoryFromRekening(def)
		if catID == "" {
			continue
		}
		catPagu[catID] += def.Pagu
		ensureCatMaps(catID)
		viewGrup := def.Grup
		for _, b := range bulanKeys {
			periodKey := gajiCategoryPeriodKeyForCalendarMonth(catID, b)
			if periodKey == "" {
				continue
			}
			native := gajiGetRekeningCell(*state, def.Kode, b)
			menuCell := gajiGetRekeningCellForGrup(*state, viewGrup, def, b)
			if native.Anggaran > 0 {
				catAng[catID][periodKey] += native.Anggaran
			}
			if menuCell.Realisasi > 0 {
				catReal[catID][periodKey] += menuCell.Realisasi
			} else if native.Realisasi > 0 {
				catReal[catID][periodKey] += native.Realisasi
			}
			peg := native.JumlahPegawai
			if peg == 0 {
				peg = menuCell.JumlahPegawai
			}
			if peg > 0 {
				catPeg[catID][periodKey] += peg
			}
		}
	}

	// Potongan terlampir (mis. jaminan kesehatan di TPG/Tamsil): realisasi per menu → kategori menu
	for _, def := range state.Rekening {
		if !def.Potongan {
			continue
		}
		for _, grup := range gajiGrupOrder {
			if !gajiRekeningAttachedJaminanKes(def, grup) {
				continue
			}
			catID := gajiCategoryForGrup(def, grup)
			if catID == "" {
				continue
			}
			ensureCatMaps(catID)
			for _, b := range bulanKeys {
				periodKey := gajiCategoryPeriodKeyForCalendarMonth(catID, b)
				if periodKey == "" {
					continue
				}
				cell := gajiGetRekeningCellForGrup(*state, grup, def, b)
				if cell.Realisasi > 0 {
					catReal[catID][periodKey] += cell.Realisasi
				}
			}
		}
	}

	for catID, pagu := range catPagu {
		state.Pagu[catID] = pagu
	}
	catIDs := map[string]bool{}
	for id := range catAng {
		catIDs[id] = true
	}
	for id := range catReal {
		catIDs[id] = true
	}
	for catID := range catIDs {
		periods := gajiPeriodsForCategory(catID)
		if len(periods) == 0 {
			continue
		}
		for _, p := range periods {
			pk := p.Key
			ang := catAng[catID][pk]
			rea := catReal[catID][pk]
			if ang == 0 && rea == 0 && catPeg[catID][pk] == 0 {
				continue
			}
			cell := state.Cells[catID][pk]
			cell.Anggaran = ang
			cell.Realisasi = rea
			if peg := catPeg[catID][pk]; peg > 0 {
				cell.JumlahPegawai = peg
			}
			state.Cells[catID][pk] = cell
		}
	}
	delete(state.Pagu, "tpg")
	delete(state.Pagu, "tamsil")
	delete(state.Cells, "tpg")
	delete(state.Cells, "tamsil")
}

func gajiBulanIndex(bulan string) int {
	bulan = normalizeBulanKey(bulan)
	for i, b := range bulanKeys {
		if b == bulan {
			return i
		}
	}
	return -1
}

func gajiGrupFromCategory(category string) string {
	switch category {
	case "gaji_pns", "gaji_pppk":
		return "gaji"
	case "tpp_pns", "tpp_pppk":
		return "tpp"
	case "tpg_pns", "tpg_pppk":
		return "tpg"
	case "tamsil_pns", "tamsil_pppk":
		return "tamsil"
	case "tpg", "tamsil":
		return category
	default:
		return category
	}
}
