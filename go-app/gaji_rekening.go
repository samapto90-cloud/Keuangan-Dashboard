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
	Pagu        float64 `json:"pagu"`
	Urut        int     `json:"urut"`
}

type GajiRekeningRow struct {
	Kode          string  `json:"kode"`
	Nama          string  `json:"nama"`
	Grup          string  `json:"grup"`
	Jenis         string  `json:"jenis"`
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
	"gaji":     "Realisasi Gaji",
	"tpp":      "Realisasi TPP",
	"tpg":      "TPG & Tamsil",
	"potongan": "Potongan",
}

var gajiGrupOrder = []string{"gaji", "tpp", "tpg", "potongan"}

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

func classifyGajiRekening(kode, nama string) (grup, jenis string) {
	k := strings.TrimSpace(kode)
	n := strings.ToUpper(strings.TrimSpace(nama))
	jenis = gajiJenisFromNama(nama)

	if strings.Contains(k, "5.1.01.02.006") {
		if strings.Contains(n, "TAMSIL") {
			return "tpg", jenis
		}
		return "tpg", jenis
	}
	if strings.Contains(k, ".007.") || strings.Contains(k, ".009.") ||
		strings.Contains(k, ".010.") || strings.Contains(k, ".011.") || strings.Contains(k, ".012.") ||
		strings.Contains(n, "PPh") || strings.Contains(n, "IURAN JAMINAN") ||
		strings.Contains(n, "IURAN SIMPANAN") {
		return "potongan", jenis
	}
	if strings.HasPrefix(k, "5.1.01.02.") {
		return "tpp", jenis
	}
	return "gaji", jenis
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
	grup, jenis := classifyGajiRekening(kode, nama)
	for i := range state.Rekening {
		if state.Rekening[i].Kode == kode {
			if nama != "" {
				state.Rekening[i].Nama = nama
			}
			if pagu > 0 {
				state.Rekening[i].Pagu = pagu
			}
			state.Rekening[i].Grup = grup
			if jenis != "" {
				state.Rekening[i].Jenis = jenis
			}
			return &state.Rekening[i]
		}
	}
	def := GajiRekeningDef{Kode: kode, Nama: nama, Grup: grup, Jenis: jenis, Pagu: pagu, Urut: urut}
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
		return state.Rekening[i].Urut < state.Rekening[j].Urut
	})
	gajiSyncCategoryFromRekening(state)
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
		if def.Grup != grup {
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
		rows = append(rows, GajiRekeningRow{
			Kode:          def.Kode,
			Nama:          def.Nama,
			Grup:          def.Grup,
			Jenis:         def.Jenis,
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
	sort.Slice(rows, func(i, j int) bool { return rows[i].Kode < rows[j].Kode })
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
		if def.Grup != grup {
			continue
		}
		row := GajiRekeningMatrixRow{
			Kode:  def.Kode,
			Nama:  def.Nama,
			Grup:  def.Grup,
			Jenis: def.Jenis,
			Pagu:  def.Pagu,
			Bulan: map[string]GajiRekeningMatrixCell{},
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
	sort.Slice(rows, func(i, j int) bool { return rows[i].Kode < rows[j].Kode })
	return rows, summary
}

func gajiSyncCategoryFromRekening(state *GajiTunjanganState) {
	ensureGajiCells(state)
	ensureGajiRekening(state)

	catPagu := map[string]float64{}
	catAng := map[string]map[string]float64{}
	catReal := map[string]map[string]float64{}
	catPeg := map[string]map[string]int{}
	potPagu := 0.0
	potReal := map[string]float64{}
	potAng := map[string]float64{}

	for _, def := range state.Rekening {
		if def.Grup == "potongan" {
			potPagu += def.Pagu
			for b, cell := range state.RekeningCells[def.Kode] {
				potReal[b] += cell.Realisasi
				potAng[b] += cell.Anggaran
			}
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
	if potPagu > 0 {
		state.Pagu["potongan"] = potPagu
	}
	if state.Cells["potongan"] == nil {
		state.Cells["potongan"] = map[string]GajiMonthCell{}
	}
	for b, real := range potReal {
		cell := state.Cells["potongan"][b]
		cell.Realisasi = real
		state.Cells["potongan"][b] = cell
	}
	for b, ang := range potAng {
		cell := state.Cells["potongan"][b]
		cell.Anggaran = ang
		state.Cells["potongan"][b] = cell
	}
	if potPagu > 0 {
		state.Pagu["potongan"] = potPagu
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
		if strings.Contains(strings.ToUpper(def.Nama), "TAMSIL") {
			return "tamsil"
		}
		if def.Jenis == "pppk" {
			return "tpg"
		}
		return "tpg"
	case "potongan":
		return "potongan"
	default:
		return ""
	}
}

func buildGajiPotonganDashboard(state GajiTunjanganState, bulan string) map[string]interface{} {
	bulan = normalizeBulanKey(bulan)
	endIdx := gajiBulanIndex(bulan)
	var pagu, realisasiSD float64
	for _, def := range state.Rekening {
		if def.Grup != "potongan" {
			continue
		}
		pagu += def.Pagu
		for b, cell := range state.RekeningCells[def.Kode] {
			if gajiBulanIndex(b) <= endIdx {
				realisasiSD += cell.Realisasi
			}
		}
	}
	return map[string]interface{}{
		"pagu":         pagu,
		"realisasi_sd": realisasiSD,
		"sisa":         pagu - realisasiSD,
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
	case "tpg", "tamsil":
		return "tpg"
	case "potongan":
		return "potongan"
	default:
		return category
	}
}
