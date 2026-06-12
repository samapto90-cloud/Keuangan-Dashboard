package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

type GajiKebutuhanRekeningRow struct {
	Kode          string  `json:"kode"`
	Nama          string  `json:"nama"`
	Kegiatan      string  `json:"kegiatan"`
	SubKegiatan   string  `json:"sub_kegiatan"`
	CategoryID    string  `json:"category_id"`
	CategoryLabel string  `json:"category_label"`
	FilterKey     string  `json:"filter_key"`
	Pagu          float64 `json:"pagu"`
	Anggaran      float64 `json:"anggaran"`
	RealisasiSD   float64 `json:"realisasi_sd"`
	Kebutuhan     float64 `json:"kebutuhan"`
	Surplus       float64 `json:"surplus"`
}

var gajiKebutuhanFilterLabels = map[string]string{
	"gaji_pns":  "Gaji PNS",
	"gaji_pppk": "Gaji PPPK",
	"tpp":       "TPP",
	"tpg":       "TPG",
	"tamsil":    "Tamsil",
}

func gajiKebutuhanFilterKeys() []string {
	return []string{"gaji_pns", "gaji_pppk", "tpp", "tpg", "tamsil"}
}

func gajiRekeningFilterKey(catID string) string {
	switch catID {
	case "tpp_pns", "tpp_pppk":
		return "tpp"
	case "tpg_pns", "tpg_pppk":
		return "tpg"
	case "tamsil_pns", "tamsil_pppk":
		return "tamsil"
	case "tpg", "tamsil":
		return catID
	default:
		return catID
	}
}

func parseGajiKebutuhanFilters(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return gajiKebutuhanFilterKeys()
	}
	allowed := map[string]bool{
		"gaji_pns": true, "gaji_pppk": true, "tpp": true, "tpg": true, "tamsil": true,
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		k := strings.TrimSpace(strings.ToLower(p))
		if allowed[k] {
			out = append(out, k)
		}
	}
	if len(out) == 0 {
		return gajiKebutuhanFilterKeys()
	}
	return out
}

func gajiRekeningMatchesKebutuhanFilter(catID string, filters []string) bool {
	if catID == "" {
		return false
	}
	fk := gajiRekeningFilterKey(catID)
	for _, f := range filters {
		if f == fk {
			return true
		}
	}
	return false
}

func gajiGetKebutuhanManual(state GajiTunjanganState, kode, bulan string) float64 {
	bulan = normalizeBulanKey(bulan)
	if state.KebutuhanManual == nil {
		return 0
	}
	if byBulan, ok := state.KebutuhanManual[kode]; ok {
		return byBulan[bulan]
	}
	return 0
}

func gajiSumRekeningRealisasiSD(state GajiTunjanganState, kode, sdBulan string) float64 {
	endIdx := gajiBulanIndex(normalizeBulanKey(sdBulan))
	if endIdx < 0 {
		return 0
	}
	var total float64
	for i, b := range bulanKeys {
		if i > endIdx {
			break
		}
		cell := gajiGetRekeningCell(state, kode, b)
		total += cell.Realisasi
	}
	return total
}

func parseGajiKebutuhanBulanList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{currentBulanKey()}
	}
	seen := map[string]bool{}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		b := normalizeBulanKey(strings.TrimSpace(p))
		if b == "" || seen[b] {
			continue
		}
		seen[b] = true
		out = append(out, b)
	}
	if len(out) == 0 {
		return []string{currentBulanKey()}
	}
	sort.Slice(out, func(i, j int) bool {
		return gajiBulanIndex(out[i]) < gajiBulanIndex(out[j])
	})
	return out
}

func gajiSumRekeningRealisasiMonths(state GajiTunjanganState, kode string, bulanList []string) float64 {
	var total float64
	for _, b := range bulanList {
		total += gajiGetRekeningCell(state, kode, b).Realisasi
	}
	return total
}

func buildGajiKebutuhanRekening(state GajiTunjanganState, bulan string, filters []string) []GajiKebutuhanRekeningRow {
	return buildGajiKebutuhanRekeningMulti(state, parseGajiKebutuhanBulanList(bulan), filters)
}

func buildGajiKebutuhanRekeningMulti(state GajiTunjanganState, bulanList []string, filters []string) []GajiKebutuhanRekeningRow {
	if len(bulanList) == 0 {
		bulanList = []string{currentBulanKey()}
	}
	if len(filters) == 0 {
		filters = gajiKebutuhanFilterKeys()
	}
	multi := len(bulanList) > 1
	var rows []GajiKebutuhanRekeningRow
	for _, def := range state.Rekening {
		if def.Potongan {
			continue
		}
		catID := gajiCategoryFromRekening(def)
		if !gajiRekeningMatchesKebutuhanFilter(catID, filters) {
			continue
		}
		var anggaran, kebutuhan, realisasi float64
		for _, b := range bulanList {
			cell := gajiGetRekeningCell(state, def.Kode, b)
			anggaran += cell.Anggaran
			kebutuhan += gajiGetKebutuhanManual(state, def.Kode, b)
		}
		if multi {
			realisasi = gajiSumRekeningRealisasiMonths(state, def.Kode, bulanList)
		} else {
			realisasi = gajiSumRekeningRealisasiSD(state, def.Kode, bulanList[0])
		}
		surplus := anggaran - kebutuhan
		kegiatan := strings.TrimSpace(def.Kegiatan)
		subKeg := strings.TrimSpace(def.SubKegiatan)
		if kegiatan == "" {
			kegiatan = "Penyediaan Gaji dan Tunjangan ASN"
		}
		if subKeg == "" {
			subKeg = "Penyediaan Gaji dan Tunjangan ASN Pemerintah"
		}
		rows = append(rows, GajiKebutuhanRekeningRow{
			Kode:          def.Kode,
			Nama:          def.Nama,
			Kegiatan:      kegiatan,
			SubKegiatan:   subKeg,
			CategoryID:    catID,
			CategoryLabel: gajiCategoryLabels[catID],
			FilterKey:     gajiRekeningFilterKey(catID),
			Pagu:          def.Pagu,
			Anggaran:      anggaran,
			RealisasiSD:   realisasi,
			Kebutuhan:     kebutuhan,
			Surplus:       surplus,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Kode < rows[j].Kode })
	return rows
}

func handleGajiSaveKebutuhan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	if getSession(r) == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	var payload struct {
		Tahun     int      `json:"tahun"`
		Bulan     string   `json:"bulan"`
		BulanList []string `json:"bulan_list"`
		Rows      []struct {
			Kode      string  `json:"kode"`
			Kebutuhan float64 `json:"kebutuhan"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	bulanList := payload.BulanList
	if len(bulanList) == 0 {
		bulanList = parseGajiKebutuhanBulanList(payload.Bulan)
	}
	if len(payload.Rows) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Data kebutuhan kosong"})
		return
	}

	gajiMu.Lock()
	if payload.Tahun > 0 {
		gajiState.Tahun = payload.Tahun
	}
	if gajiState.KebutuhanManual == nil {
		gajiState.KebutuhanManual = map[string]map[string]float64{}
	}
	for _, row := range payload.Rows {
		kode := strings.TrimSpace(row.Kode)
		if kode == "" {
			continue
		}
		if gajiState.KebutuhanManual[kode] == nil {
			gajiState.KebutuhanManual[kode] = map[string]float64{}
		}
		var totalAng float64
		angByBulan := map[string]float64{}
		for _, b := range bulanList {
			ang := gajiGetRekeningCell(gajiState, kode, b).Anggaran
			angByBulan[b] = ang
			totalAng += ang
		}
		for _, b := range bulanList {
			var share float64
			if totalAng > 0 {
				share = row.Kebutuhan * (angByBulan[b] / totalAng)
			} else if len(bulanList) > 0 {
				share = row.Kebutuhan / float64(len(bulanList))
			}
			gajiState.KebutuhanManual[kode][b] = share
		}
	}
	rows := buildGajiKebutuhanRekeningMulti(gajiState, bulanList, gajiKebutuhanFilterKeys())
	gajiMu.Unlock()
	persistGajiState()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":           "Kebutuhan per rekening disimpan",
		"kebutuhan_rekening": rows,
	})
}
