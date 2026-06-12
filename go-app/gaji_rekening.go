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
	Kode         string                              `json:"kode"`
	Nama         string                              `json:"nama"`
	Grup         string                              `json:"grup"`
	Jenis        string                              `json:"jenis"`
	Potongan     bool                                `json:"potongan,omitempty"`
	Attached     bool                                `json:"attached,omitempty"`
	Pagu         float64                             `json:"pagu"`
	RealisasiSD  float64                             `json:"realisasi_sd"`
	KebutuhanSisa float64                            `json:"kebutuhan_sisa"`
	Bulan        map[string]GajiRekeningMatrixCell   `json:"bulan"`
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
	case "tpg":
		return def.Jenis == "pns" || def.Jenis == ""
	case "tamsil":
		return def.Jenis == "pns" || def.Jenis == "pppk" || def.Jenis == ""
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
		Locked:     state.RealisasiLocked != nil && state.RealisasiLocked[gajiRekeningLockKey(grup, bulan)],
	}
	var rows []GajiRekeningRow
	maxPNS, maxPPPK := state.Pegawai["pns"], state.Pegawai["pppk"]
	for _, def := range state.Rekening {
		if !gajiRekeningIncludedInGrup(def, grup) {
			continue
		}
		cell := gajiGetRekeningCell(state, def.Kode, bulan)
		pegawai := gajiRekeningPegawaiFallback(state, def, cell)
		if def.Jenis == "pppk" && pegawai > maxPPPK {
			maxPPPK = pegawai
		} else if def.Jenis != "pppk" && pegawai > maxPNS {
			maxPNS = pegawai
		}
		sisa := cell.Anggaran - cell.Realisasi
		persen := 0.0
		if cell.Anggaran > 0 {
			persen = (cell.Realisasi / cell.Anggaran) * 100
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
			JumlahPegawai: pegawai,
			Anggaran:      cell.Anggaran,
			Realisasi:     cell.Realisasi,
			Sisa:          sisa,
			Persen:        persen,
			Locked:        summary.Locked,
		})
		summary.TotalAnggaran += cell.Anggaran
		summary.TotalRealisasi += cell.Realisasi
		summary.TotalSisa += sisa
	}
	summary.TotalPegawai = maxPNS + maxPPPK
	gajiSortRekeningReportRows(rows, grup)
	return rows, summary
}

func gajiRekeningPegawaiFallback(state GajiTunjanganState, def GajiRekeningDef, cell GajiMonthCell) int {
	pegawai := cell.JumlahPegawai
	if pegawai == 0 {
		if def.Jenis == "pppk" {
			pegawai = state.Pegawai["pppk"]
		} else {
			pegawai = state.Pegawai["pns"]
		}
	}
	return pegawai
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
			cell := gajiGetRekeningCell(state, def.Kode, b)
			pegawai := gajiRekeningPegawaiFallback(state, def, cell)
			sisa := cell.Anggaran - cell.Realisasi
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
		row.KebutuhanSisa = row.Pagu - row.RealisasiSD
		if row.KebutuhanSisa < 0 {
			row.KebutuhanSisa = 0
		}
		rows = append(rows, row)
	}
	gajiSortRekeningMatrixRows(rows, grup)
	return rows, summary
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

func gajiSyncCategoryFromRekening(state *GajiTunjanganState) {
	ensureGajiCells(state)
	ensureGajiRekening(state)

	catPagu := map[string]float64{}
	catAng := map[string]map[string]float64{}
	catReal := map[string]map[string]float64{}
	catPeg := map[string]map[string]int{}

	for _, def := range state.Rekening {
		if def.Potongan {
			continue
		}
		catID := gajiCategoryFromRekening(def)
		if catID == "" {
			continue
		}
		catPagu[catID] += def.Pagu
		if catAng[catID] == nil {
			catAng[catID] = map[string]float64{}
			catReal[catID] = map[string]float64{}
			catPeg[catID] = map[string]int{}
		}
		for b, cell := range state.RekeningCells[def.Kode] {
			catAng[catID][b] += cell.Anggaran
			catReal[catID][b] += cell.Realisasi
			if cell.JumlahPegawai > 0 {
				catPeg[catID][b] += cell.JumlahPegawai
			}
		}
	}

	for catID, pagu := range catPagu {
		state.Pagu[catID] = pagu
		for b, ang := range catAng[catID] {
			cell := state.Cells[catID][b]
			cell.Anggaran = ang
			cell.Realisasi = catReal[catID][b]
			if p := catPeg[catID][b]; p > 0 {
				cell.JumlahPegawai = p
			}
			state.Cells[catID][b] = cell
		}
	}
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
		return "tpg"
	case "tamsil":
		return "tamsil"
	default:
		return ""
	}
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
	case "tpg":
		return "tpg"
	case "tamsil":
		return "tamsil"
	default:
		return category
	}
}
