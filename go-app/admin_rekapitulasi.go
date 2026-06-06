package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

var adminRekapSipkeuIDs = []string{"sekretariat", "paud", "sd", "smp"}

type adminRekapRow struct {
	PortalID     string  `json:"portal_id"`
	PortalLabel  string  `json:"portal_label"`
	Kegiatan     string  `json:"kegiatan,omitempty"`
	SubKegiatan  string  `json:"sub_kegiatan,omitempty"`
	Pekerjaan    string  `json:"pekerjaan,omitempty"`
	KodeRekening string  `json:"kode_rekening,omitempty"`
	PPTK         string  `json:"pptk,omitempty"`
	Anggaran     float64 `json:"anggaran"`
	Realisasi    float64 `json:"realisasi"`
	Sisa         float64 `json:"sisa"`
	Count        int     `json:"count"`
	Pajak        float64 `json:"pajak"`
	Pct          float64 `json:"pct"`
}

type rekapAgg struct {
	portalID, portalLabel          string
	kegiatan, sub, pekerjaan, kode string
	pptk                           string
	anggaran, realisasi, pajak     float64
	count                          int
}

type adminRekapPPTKStat struct {
	PortalID    string  `json:"portal_id"`
	PortalLabel string  `json:"portal_label"`
	PPTK        string  `json:"pptk"`
	Anggaran    float64 `json:"anggaran"`
	Realisasi   float64 `json:"realisasi"`
	Sisa        float64 `json:"sisa"`
	Count       int     `json:"count"`
	Pct         float64 `json:"pct"`
}

func normRekap(s string) string {
	return strings.TrimSpace(s)
}

func parseAdminRekapPortals(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "all" {
		return append([]string(nil), adminRekapSipkeuIDs...)
	}
	allowed := map[string]bool{}
	for _, id := range adminRekapSipkeuIDs {
		allowed[id] = true
	}
	out := []string{}
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if allowed[p] {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), adminRekapSipkeuIDs...)
	}
	return out
}

func transactionInDateRange(t Transaction, from, to string) bool {
	d := strings.TrimSpace(t.Tanggal)
	if d == "" {
		return false
	}
	if from != "" && d < from {
		return false
	}
	if to != "" && d > to {
		return false
	}
	return true
}

func rekapPct(realisasi, anggaran float64) float64 {
	if anggaran <= 0 {
		return 0
	}
	return (realisasi / anggaran) * 100
}

func moduleSettingsSnapshot(mod *SipkeuModule) (rak []RakRow, anggaranKeg map[string]float64) {
	if mod == nil {
		return nil, map[string]float64{}
	}
	mod.mu.Lock()
	defer mod.mu.Unlock()
	rak = append([]RakRow(nil), mod.settings.Rak...)
	anggaranKeg = map[string]float64{}
	for k, v := range mod.settings.AnggaranKegiatan {
		anggaranKeg[normRekap(k)] = v
	}
	return rak, anggaranKeg
}

func sumRakKegiatan(rak []RakRow, kegiatan string) float64 {
	kegiatan = normRekap(kegiatan)
	var sum float64
	for _, r := range rak {
		if normRekap(r.Kegiatan) == kegiatan {
			sum += r.Anggaran
		}
	}
	return sum
}

func sumRakSub(rak []RakRow, kegiatan, sub string) float64 {
	kegiatan = normRekap(kegiatan)
	sub = normRekap(sub)
	var sum float64
	for _, r := range rak {
		if normRekap(r.Kegiatan) == kegiatan && normRekap(r.SubKegiatan) == sub {
			sum += r.Anggaran
		}
	}
	return sum
}

func anggaranKegiatanForSnapshot(rak []RakRow, anggaranKeg map[string]float64, kegiatan string, realisasi float64) float64 {
	kegiatan = normRekap(kegiatan)
	if kegiatan == "" {
		return 0
	}
	// Selaras dengan getAnggaranForKegiatan + pagu dari RAK jika belum di-set manual
	if v := anggaranKeg[kegiatan]; v > 0 {
		return v
	}
	if rakSum := sumRakKegiatan(rak, kegiatan); rakSum > 0 {
		return rakSum
	}
	if realisasi > 0 {
		padded := int(realisasi*1.25/1000000) + 1
		if padded < 100 {
			padded = 100
		}
		return float64(padded) * 1000000
	}
	return 0
}

func anggaranSubForSnapshot(rak []RakRow, kegiatan, sub string) float64 {
	return sumRakSub(rak, kegiatan, sub)
}

func anggaranPekerjaanForSnapshot(rak []RakRow, kegiatan, sub, kode, pekerjaan string) float64 {
	kegiatan = normRekap(kegiatan)
	sub = normRekap(sub)
	kode = normRekap(kode)
	pekerjaan = normRekap(pekerjaan)
	for _, r := range rak {
		if normRekap(r.Kegiatan) == kegiatan &&
			normRekap(r.SubKegiatan) == sub &&
			normRekap(r.KodeRekening) == kode &&
			normRekap(r.Pekerjaan) == pekerjaan {
			return r.Anggaran
		}
	}
	return 0
}

func portalTotalAnggaranSnapshot(rak []RakRow, anggaranKeg map[string]float64) float64 {
	var total float64
	for _, r := range rak {
		total += r.Anggaran
	}
	if total > 0 {
		return total
	}
	for _, v := range anggaranKeg {
		if v > 0 {
			total += v
		}
	}
	return total
}

func rekapAggKey(mode string, a *rekapAgg) string {
	switch mode {
	case "portal":
		return a.portalID
	case "sub":
		return a.portalID + "\x00" + normRekap(a.kegiatan) + "\x00" + normRekap(a.sub)
	case "pekerjaan":
		return a.portalID + "\x00" + normRekap(a.kegiatan) + "\x00" + normRekap(a.sub) + "\x00" + normRekap(a.kode) + "\x00" + normRekap(a.pekerjaan)
	case "pptk":
		return a.portalID + "\x00" + normRekap(a.pptk)
	default:
		return a.portalID + "\x00" + normRekap(a.kegiatan)
	}
}

func rekapPortalMeta(sourceModID, kegiatan string) (portalID, label string) {
	portalID = rekapPortalForKegiatan(sourceModID, kegiatan)
	return portalID, portalLabel(portalID)
}

func seedRekapFromRak(mod *SipkeuModule, mode string, aggs map[string]*rekapAgg, rak []RakRow, anggaranKeg map[string]float64) {
	if mod == nil {
		return
	}
	label := portalLabel(mod.ID)

	switch mode {
	case "portal":
		key := mod.ID
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{
				portalID: mod.ID, portalLabel: label,
				anggaran: portalTotalAnggaranSnapshot(rak, anggaranKeg),
			}
		}
	case "kegiatan":
		seen := map[string]bool{}
		for k, v := range anggaranKeg {
			if k == "" || !moduleOwnsKegiatan(mod.ID, k) {
				continue
			}
			seen[k] = true
			pid, plabel := rekapPortalMeta(mod.ID, k)
			a := &rekapAgg{portalID: pid, portalLabel: plabel, kegiatan: k, anggaran: v}
			aggs[rekapAggKey(mode, a)] = a
		}
		for _, r := range rak {
			k := normRekap(r.Kegiatan)
			if k == "" || seen[k] || !moduleOwnsKegiatan(mod.ID, k) {
				continue
			}
			seen[k] = true
			pid, plabel := rekapPortalMeta(mod.ID, k)
			a := &rekapAgg{portalID: pid, portalLabel: plabel, kegiatan: k, anggaran: sumRakKegiatan(rak, k)}
			aggs[rekapAggKey(mode, a)] = a
		}
	case "sub":
		for _, r := range rak {
			keg := normRekap(r.Kegiatan)
			sub := normRekap(r.SubKegiatan)
			if keg == "" || sub == "" || !moduleOwnsKegiatan(mod.ID, keg) {
				continue
			}
			pid, plabel := rekapPortalMeta(mod.ID, keg)
			key := rekapAggKey(mode, &rekapAgg{portalID: pid, kegiatan: keg, sub: sub})
			if aggs[key] == nil {
				aggs[key] = &rekapAgg{
					portalID: pid, portalLabel: plabel,
					kegiatan: keg, sub: sub,
					anggaran: sumRakSub(rak, keg, sub),
				}
			}
		}
	case "pekerjaan":
		for _, r := range rak {
			keg := normRekap(r.Kegiatan)
			sub := normRekap(r.SubKegiatan)
			pk := normRekap(r.Pekerjaan)
			if keg == "" || sub == "" || pk == "" || !moduleOwnsKegiatan(mod.ID, keg) {
				continue
			}
			pid, plabel := rekapPortalMeta(mod.ID, keg)
			key := rekapAggKey(mode, &rekapAgg{
				portalID: pid, kegiatan: keg, sub: sub,
				kode: normRekap(r.KodeRekening), pekerjaan: pk,
			})
			if aggs[key] == nil {
				aggs[key] = &rekapAgg{
					portalID: pid, portalLabel: plabel,
					kegiatan: keg, sub: sub,
					kode: normRekap(r.KodeRekening), pekerjaan: pk,
					anggaran: r.Anggaran,
				}
			}
		}
	case "pptk":
		for _, r := range rak {
			keg := normRekap(r.Kegiatan)
			if keg != "" && !moduleOwnsKegiatan(mod.ID, keg) {
				continue
			}
			pptk := normRekap(r.PPTK)
			if pptk == "" {
				continue
			}
			key := rekapAggKey(mode, &rekapAgg{portalID: mod.ID, pptk: pptk})
			if aggs[key] == nil {
				aggs[key] = &rekapAgg{portalID: mod.ID, portalLabel: label, pptk: pptk}
			}
			aggs[key].anggaran += r.Anggaran
		}
	}
}

func sumRakPPTK(rak []RakRow, pptk string) float64 {
	pptk = normRekap(pptk)
	var sum float64
	for _, r := range rak {
		if normRekap(r.PPTK) == pptk {
			sum += r.Anggaran
		}
	}
	return sum
}

func pptkForTransaction(rak []RakRow, t Transaction) string {
	if p := normRekap(t.PPTK); p != "" {
		return p
	}
	keg := normRekap(t.Kegiatan)
	sub := normRekap(t.SubKegiatan)
	kode := normRekap(t.KodeRekening)
	pk := normRekap(t.Pekerjaan)
	for _, r := range rak {
		if normRekap(r.Kegiatan) == keg &&
			normRekap(r.SubKegiatan) == sub &&
			normRekap(r.KodeRekening) == kode &&
			normRekap(r.Pekerjaan) == pk {
			return normRekap(r.PPTK)
		}
	}
	return ""
}

func rekapRowVisible(mode string, a *rekapAgg) bool {
	switch mode {
	case "portal":
		return a.portalID != ""
	case "kegiatan":
		return a.kegiatan != ""
	case "sub":
		return a.kegiatan != "" && a.sub != ""
	case "pekerjaan":
		return a.kegiatan != "" && a.sub != "" && a.pekerjaan != ""
	case "pptk":
		return a.pptk != ""
	default:
		return false
	}
}

func applyRekapTransaction(mod *SipkeuModule, mode, from, to string, t Transaction, aggs map[string]*rekapAgg, rak []RakRow, anggaranKeg map[string]float64) {
	if mod == nil || !trxIsApproved(t) || !transactionBelongsToModule(mod, t) {
		return
	}
	if !transactionInDateRange(t, from, to) {
		return
	}
	label := portalLabel(mod.ID)
	keg := normRekap(t.Kegiatan)
	sub := normRekap(t.SubKegiatan)
	pk := normRekap(t.Pekerjaan)
	kode := normRekap(t.KodeRekening)

	var a *rekapAgg
	switch mode {
	case "portal":
		key := mod.ID
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{
				portalID: mod.ID, portalLabel: label,
				anggaran: portalTotalAnggaranSnapshot(rak, anggaranKeg),
			}
		}
		a = aggs[key]
	case "sub":
		if keg == "" || sub == "" {
			return
		}
		pid, plabel := rekapPortalMeta(mod.ID, keg)
		key := rekapAggKey(mode, &rekapAgg{portalID: pid, kegiatan: keg, sub: sub})
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{
				portalID: pid, portalLabel: plabel,
				kegiatan: keg, sub: sub,
				anggaran: sumRakSub(rak, keg, sub),
			}
		}
		a = aggs[key]
	case "pekerjaan":
		if keg == "" || sub == "" || pk == "" {
			return
		}
		pid, plabel := rekapPortalMeta(mod.ID, keg)
		key := rekapAggKey(mode, &rekapAgg{portalID: pid, kegiatan: keg, sub: sub, kode: kode, pekerjaan: pk})
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{
				portalID: pid, portalLabel: plabel,
				kegiatan: keg, sub: sub, kode: kode, pekerjaan: pk,
				anggaran: anggaranPekerjaanForSnapshot(rak, keg, sub, kode, pk),
			}
		}
		a = aggs[key]
	case "pptk":
		pptk := pptkForTransaction(rak, t)
		if pptk == "" {
			return
		}
		pid, plabel := mod.ID, label
		if keg != "" {
			pid = rekapPortalForKegiatan(mod.ID, keg)
			plabel = portalLabel(pid)
		}
		key := rekapAggKey(mode, &rekapAgg{portalID: pid, pptk: pptk})
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{portalID: pid, portalLabel: plabel, pptk: pptk}
		}
		a = aggs[key]
	case "kegiatan":
		if keg == "" {
			return
		}
		pid, plabel := rekapPortalMeta(mod.ID, keg)
		key := rekapAggKey(mode, &rekapAgg{portalID: pid, kegiatan: keg})
		if aggs[key] == nil {
			aggs[key] = &rekapAgg{
				portalID: pid, portalLabel: plabel, kegiatan: keg,
				anggaran: anggaranKegiatanForSnapshot(rak, anggaranKeg, keg, 0),
			}
		}
		a = aggs[key]
	default:
		return
	}
	a.realisasi += t.Nilai
	a.pajak += t.Pajak
	a.count++
}

func finalizeRekapAnggaran(mod *SipkeuModule, mode string, aggs map[string]*rekapAgg, rak []RakRow, anggaranKeg map[string]float64) {
	for _, a := range aggs {
		if a.portalID != mod.ID {
			continue
		}
		switch mode {
		case "kegiatan":
			a.anggaran = anggaranKegiatanForSnapshot(rak, anggaranKeg, a.kegiatan, a.realisasi)
		case "sub":
			a.anggaran = anggaranSubForSnapshot(rak, a.kegiatan, a.sub)
		case "pekerjaan":
			a.anggaran = anggaranPekerjaanForSnapshot(rak, a.kegiatan, a.sub, a.kode, a.pekerjaan)
		case "portal":
			a.anggaran = portalTotalAnggaranSnapshot(rak, anggaranKeg)
		case "pptk":
			a.anggaran = sumRakPPTK(rak, a.pptk)
		}
	}
}

func buildAdminRekap(portals []string, mode, from, to string) []adminRekapRow {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = "kegiatan"
	}
	aggs := map[string]*rekapAgg{}

	for _, portalID := range portals {
		mod := sipkeuModules[portalID]
		if mod == nil {
			continue
		}
		rak, anggaranKeg := moduleSettingsSnapshot(mod)
		seedRekapFromRak(mod, mode, aggs, rak, anggaranKeg)
		txs := moduleTransactionsCopy(mod)
		for i := range txs {
			applyRekapTransaction(mod, mode, from, to, txs[i], aggs, rak, anggaranKeg)
		}
		finalizeRekapAnggaran(mod, mode, aggs, rak, anggaranKeg)
	}

	rows := make([]adminRekapRow, 0, len(aggs))
	for _, a := range aggs {
		if !rekapRowVisible(mode, a) {
			continue
		}
		sisa := a.anggaran - a.realisasi
		rows = append(rows, adminRekapRow{
			PortalID:     a.portalID,
			PortalLabel:  a.portalLabel,
			Kegiatan:     a.kegiatan,
			SubKegiatan:  a.sub,
			Pekerjaan:    a.pekerjaan,
			KodeRekening: a.kode,
			PPTK:         a.pptk,
			Anggaran:     a.anggaran,
			Realisasi:    a.realisasi,
			Sisa:         sisa,
			Count:        a.count,
			Pajak:        a.pajak,
			Pct:          rekapPct(a.realisasi, a.anggaran),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].PortalID != rows[j].PortalID {
			return rows[i].PortalID < rows[j].PortalID
		}
		if mode == "pptk" {
			return rows[i].PPTK < rows[j].PPTK
		}
		if rows[i].Anggaran != rows[j].Anggaran {
			return rows[i].Anggaran > rows[j].Anggaran
		}
		if rows[i].Realisasi != rows[j].Realisasi {
			return rows[i].Realisasi > rows[j].Realisasi
		}
		return rows[i].Kegiatan < rows[j].Kegiatan
	})
	return rows
}

func buildAdminRekapPPTKStats(portals []string, from, to string) []adminRekapPPTKStat {
	rows := cachedAdminRekapRows(portals, "pptk", from, to, func() []adminRekapRow {
		return buildAdminRekap(portals, "pptk", from, to)
	})
	return adminRekapRowsToPPTKStats(rows)
}

func adminRekapSummary(rows []adminRekapRow) map[string]interface{} {
	var totAng, totReal, totPajak float64
	var totTrx int
	for _, row := range rows {
		totAng += row.Anggaran
		totReal += row.Realisasi
		totPajak += row.Pajak
		totTrx += row.Count
	}
	return map[string]interface{}{
		"anggaran":  totAng,
		"realisasi": totReal,
		"sisa":      totAng - totReal,
		"pajak":     totPajak,
		"count":     totTrx,
		"pct":       rekapPct(totReal, totAng),
	}
}

func handleAdminRekapitulasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	mode := strings.TrimSpace(r.URL.Query().Get("mode"))
	if mode == "" {
		mode = "kegiatan"
	}
	validModes := map[string]bool{"kegiatan": true, "sub": true, "pekerjaan": true, "portal": true, "pptk": true}
	if !validModes[mode] {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Mode rekap tidak valid"})
		return
	}
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	to := strings.TrimSpace(r.URL.Query().Get("to"))
	portals := parseAdminRekapPortals(r.URL.Query().Get("portals"))
	rows := cachedAdminRekapRows(portals, mode, from, to, func() []adminRekapRow {
		return buildAdminRekap(portals, mode, from, to)
	})
	var pptkStats []adminRekapPPTKStat
	if mode == "pptk" {
		pptkStats = adminRekapRowsToPPTKStats(rows)
	} else {
		pptkStats = buildAdminRekapPPTKStats(portals, from, to)
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"mode":         mode,
		"from":         from,
		"to":           to,
		"portals":      portals,
		"rows":         rows,
		"pptk_stats":   pptkStats,
		"generated_at": time.Now().Format(time.RFC3339),
		"summary":      adminRekapSummary(rows),
		"labels": map[string]string{
			"mode": fmt.Sprintf("Rekapitulasi %s", map[string]string{
				"kegiatan": "Per Kegiatan", "sub": "Per Sub Kegiatan",
				"pekerjaan": "Per Pekerjaan", "portal": "Per Portal", "pptk": "Per PPTK",
			}[mode]),
		},
	})
}
