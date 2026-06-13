package main

import (
	"strings"
)

var bulanLabels = map[string]string{
	"januari": "Januari", "februari": "Februari", "maret": "Maret", "april": "April",
	"mei": "Mei", "juni": "Juni", "juli": "Juli", "agustus": "Agustus",
	"september": "September", "oktober": "Oktober", "november": "November", "desember": "Desember",
}

type gajiPeriodDef struct {
	Key   string
	Label string
}

var gajiRealisasiPeriods = []gajiPeriodDef{
	{Key: "januari", Label: "Jan"},
	{Key: "februari", Label: "Feb"},
	{Key: "maret", Label: "Mar"},
	{Key: "thr", Label: "THR"},
	{Key: "april", Label: "April"},
	{Key: "mei", Label: "Mei"},
	{Key: "juni", Label: "Juni"},
	{Key: "gaji_13", Label: "Gaji 13"},
	{Key: "juli", Label: "Juli"},
	{Key: "agustus", Label: "Agustus"},
	{Key: "september", Label: "September"},
	{Key: "oktober", Label: "Oktober"},
	{Key: "november", Label: "November"},
	{Key: "desember", Label: "Desember"},
}

var tppRealisasiPeriods = []gajiPeriodDef{
	{Key: "januari", Label: "Jan"},
	{Key: "februari", Label: "Feb"},
	{Key: "thr", Label: "THR"},
	{Key: "maret", Label: "Mar"},
	{Key: "april", Label: "Apr"},
	{Key: "mei", Label: "Mei"},
	{Key: "juni", Label: "Juni"},
	{Key: "tpp_13", Label: "TPP 13"},
	{Key: "juli", Label: "Juli"},
	{Key: "agustus", Label: "Agustus"},
	{Key: "september", Label: "Sept"},
	{Key: "oktober", Label: "Okt"},
	{Key: "november", Label: "Nov"},
	{Key: "desember", Label: "Des"},
}

var tpgRealisasiPeriods = []gajiPeriodDef{
	{Key: "tw1", Label: "TW I"},
	{Key: "tw2", Label: "TW II"},
	{Key: "tw3", Label: "TW III"},
	{Key: "tw4", Label: "TW IV"},
}

func gajiPeriodsForCategory(category string) []gajiPeriodDef {
	switch category {
	case "gaji_pns", "gaji_pppk":
		return gajiRealisasiPeriods
	case "tpp_pns", "tpp_pppk":
		return tppRealisasiPeriods
	case "tpg_pns", "tpg_pppk", "tamsil_pns", "tamsil_pppk":
		return tpgRealisasiPeriods
	default:
		return nil
	}
}

// gajiRekapLookupPeriod — baris rekap memakai periode gaji bulanan; TPG/Tamsil disimpan per TW.
// gajiCategoryPeriodKeyForCalendarMonth — bulan kalender (grid realisasi) → kunci periode kategori.
func gajiCategoryPeriodKeyForCalendarMonth(categoryID, calendarMonth string) string {
	calendarMonth = normalizeBulanKey(calendarMonth)
	if calendarMonth == "" {
		return ""
	}
	switch categoryID {
	case "tpg_pns", "tpg_pppk", "tamsil_pns", "tamsil_pppk":
		switch calendarMonth {
		case "januari", "februari", "maret":
			return "tw1"
		case "april", "mei", "juni":
			return "tw2"
		case "juli", "agustus", "september":
			return "tw3"
		case "oktober", "november", "desember":
			return "tw4"
		}
		return ""
	default:
		if gajiPeriodIndex(categoryID, calendarMonth) >= 0 {
			return calendarMonth
		}
		return ""
	}
}

func gajiRekapLookupPeriod(categoryID, rowPeriodKey string) string {
	switch categoryID {
	case "tpg_pns", "tpg_pppk", "tamsil_pns", "tamsil_pppk":
		switch rowPeriodKey {
		case "januari":
			return "tw1"
		case "april":
			return "tw2"
		case "juli":
			return "tw3"
		case "oktober":
			return "tw4"
		default:
			return ""
		}
	default:
		return rowPeriodKey
	}
}

func gajiPeriodLabel(category, key string) string {
	for _, p := range gajiPeriodsForCategory(category) {
		if p.Key == key {
			return p.Label
		}
	}
	if l, ok := bulanLabels[key]; ok {
		return l
	}
	return key
}

func gajiPeriodIndex(category, key string) int {
	key = normalizeBulanKey(key)
	for i, p := range gajiPeriodsForCategory(category) {
		if p.Key == key {
			return i
		}
	}
	return -1
}

func normalizeGajiPeriod(category, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	low := strings.ToLower(raw)
	for _, p := range gajiPeriodsForCategory(category) {
		if strings.ToLower(p.Key) == low || strings.ToLower(p.Label) == low {
			return p.Key
		}
	}
	aliases := map[string]string{
		"jan": "januari", "feb": "februari", "mar": "maret", "apr": "april", "april": "april",
		"may": "mei", "jun": "juni", "jul": "juli", "aug": "agustus", "sep": "september",
		"oct": "oktober", "okt": "oktober", "nov": "november", "dec": "desember", "des": "desember",
		"gaji13": "gaji_13", "gaji 13": "gaji_13", "tpp13": "tpp_13", "tpp 13": "tpp_13",
		"tw i": "tw1", "tw 1": "tw1", "tw1": "tw1", "tw ii": "tw2", "tw2": "tw2",
		"tw iii": "tw3", "tw3": "tw3", "tw iv": "tw4", "tw4": "tw4",
	}
	if v, ok := aliases[low]; ok && gajiPeriodIndex(category, v) >= 0 {
		return v
	}
	if gajiPeriodIndex(category, low) >= 0 {
		return low
	}
	return normalizeBulanKey(raw)
}

// gajiRealisasiSDEndIndex: indeks periode terakhir yang masuk kolom "Total Realisasi" template Excel.
func gajiRealisasiSDEndIndex(category, reportingMonth string) int {
	reportingMonth = normalizeBulanKey(reportingMonth)
	if reportingMonth == "" {
		reportingMonth = currentBulanKey()
	}
	gajiMap := map[string]int{
		"januari": 0, "februari": 1, "maret": 2, "april": 4, "mei": 5, "juni": 6,
		"juli": 8, "agustus": 9, "september": 10, "oktober": 11, "november": 12, "desember": 13,
	}
	tppMap := map[string]int{
		"januari": 0, "februari": 1, "maret": 3, "april": 4, "mei": 5, "juni": 6,
		"juli": 8, "agustus": 9, "september": 10, "oktober": 11, "november": 12, "desember": 13,
	}
	switch category {
	case "gaji_pns", "gaji_pppk":
		if idx, ok := gajiMap[reportingMonth]; ok {
			return idx
		}
		return 6
	case "tpp_pns", "tpp_pppk":
		if idx, ok := tppMap[reportingMonth]; ok {
			return idx
		}
		return 6
	case "tpg_pns", "tpg_pppk", "tamsil_pns", "tamsil_pppk":
		q := map[string]int{
			"januari": 0, "februari": 0, "maret": 0,
			"april": 1, "mei": 1, "juni": 1,
			"juli": 2, "agustus": 2, "september": 2,
			"oktober": 3, "november": 3, "desember": 3,
		}
		if idx, ok := q[reportingMonth]; ok {
			return idx
		}
		return 1
	default:
		periods := gajiPeriodsForCategory(category)
		if len(periods) == 0 {
			return -1
		}
		return len(periods) - 1
	}
}

func gajiSumRealisasiPeriods(state GajiTunjanganState, category string, from, to int) float64 {
	periods := gajiPeriodsForCategory(category)
	if len(periods) == 0 {
		return 0
	}
	if from < 0 {
		from = 0
	}
	if to >= len(periods) {
		to = len(periods) - 1
	}
	var sum float64
	for i := from; i <= to; i++ {
		sum += gajiGetCell(state, category, periods[i].Key).Realisasi
	}
	return sum
}

func gajiSumAllRealisasi(state GajiTunjanganState, category string) float64 {
	periods := gajiPeriodsForCategory(category)
	if len(periods) == 0 {
		return 0
	}
	return gajiSumRealisasiPeriods(state, category, 0, len(periods)-1)
}

func gajiSplitRealisasiKebutuhan(state GajiTunjanganState, category, reportingMonth string) (realisasiSD, kebutuhanSisa, kebutuhanTahun float64) {
	periods := gajiPeriodsForCategory(category)
	if len(periods) == 0 {
		return 0, 0, 0
	}
	endIdx := gajiRealisasiSDEndIndex(category, reportingMonth)
	realisasiSD = gajiSumRealisasiPeriods(state, category, 0, endIdx)
	kebutuhanTahun = gajiSumAllRealisasi(state, category)
	if endIdx+1 < len(periods) {
		kebutuhanSisa = gajiSumRealisasiPeriods(state, category, endIdx+1, len(periods)-1)
	}
	if kebutuhanTahun == 0 {
		kebutuhanTahun = realisasiSD + kebutuhanSisa
	} else if kebutuhanSisa == 0 && kebutuhanTahun > realisasiSD {
		kebutuhanSisa = kebutuhanTahun - realisasiSD
	}
	return realisasiSD, kebutuhanSisa, kebutuhanTahun
}

func gajiGetPagu(state GajiTunjanganState, category string) float64 {
	if state.Pagu != nil {
		if v, ok := state.Pagu[category]; ok && v > 0 {
			return v
		}
	}
	var total float64
	for _, p := range gajiPeriodsForCategory(category) {
		total += gajiGetCell(state, category, p.Key).Anggaran
	}
	return total
}

func gajiPegawaiForCategory(state GajiTunjanganState, category string) int {
	if state.Pegawai != nil {
		switch category {
		case "gaji_pns", "tpp_pns", "tpg_pns", "tamsil_pns":
			if v := state.Pegawai["pns"]; v > 0 {
				return v
			}
		case "gaji_pppk", "tpp_pppk", "tpg_pppk", "tamsil_pppk":
			if v := state.Pegawai["pppk"]; v > 0 {
				return v
			}
		}
	}
	for _, p := range gajiPeriodsForCategory(category) {
		if v := gajiGetCell(state, category, p.Key).JumlahPegawai; v > 0 {
			return v
		}
	}
	return 0
}
