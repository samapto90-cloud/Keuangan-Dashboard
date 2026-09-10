package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func defaultTahunAnggaran() int {
	return 2026
}

func getTahunAnggaran() int {
	sys := getSystemSettingsCopy()
	if sys.TahunAnggaran >= 2000 && sys.TahunAnggaran <= 2100 {
		return sys.TahunAnggaran
	}
	return defaultTahunAnggaran()
}

func setTahunAnggaran(tahun int, syncKasGaji bool) error {
	if tahun < 2000 || tahun > 2100 {
		return fmt.Errorf("tahun anggaran tidak valid")
	}
	systemSettingsMu.Lock()
	systemSettings.TahunAnggaran = tahun
	systemSettingsMu.Unlock()
	persistSystemSettings()
	invalidatePortalStatusCache()

	if syncKasGaji {
		kasMu.Lock()
		kasState.Tahun = tahun
		kasMu.Unlock()
		persistKasState()

		gajiMu.Lock()
		gajiState.Tahun = tahun
		gajiMu.Unlock()
		persistGajiState()
	}
	return nil
}

func sipkeuDataModuleIDs() []string {
	return []string{"sekretariat", "paud", "sd", "smp"}
}

func buildDataOpsSummary() map[string]interface{} {
	modules := []map[string]interface{}{}
	totalTx := 0
	totalRak := 0
	for _, id := range sipkeuDataModuleIDs() {
		mod := sipkeuModules[id]
		txCount, rakCount := 0, 0
		if mod != nil {
			mod.mu.Lock()
			txCount = len(mod.txs)
			rakCount = len(mod.settings.Rak)
			mod.mu.Unlock()
		}
		totalTx += txCount
		totalRak += rakCount
		modules = append(modules, map[string]interface{}{
			"id":        id,
			"label":     portalLabel(id),
			"tx_count":  txCount,
			"rak_count": rakCount,
		})
	}

	kasMu.RLock()
	kasRak := len(kasState.RakRows)
	kasTahun := kasState.Tahun
	kasMu.RUnlock()

	gajiMu.RLock()
	gajiTahun := gajiState.Tahun
	gajiRek := len(gajiState.Rekening)
	gajiMu.RUnlock()

	return map[string]interface{}{
		"tahun_anggaran": getTahunAnggaran(),
		"generated_at":   time.Now().Format(time.RFC3339),
		"modules":        modules,
		"totals": map[string]interface{}{
			"transactions": totalTx,
			"rak_rows":     totalRak,
		},
		"kas_belanja": map[string]interface{}{
			"tahun":    kasTahun,
			"rak_rows": kasRak,
		},
		"gaji_asn": map[string]interface{}{
			"tahun":    gajiTahun,
			"rekening": gajiRek,
		},
	}
}

func handleAdminDataOpsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	jsonResponse(w, http.StatusOK, buildDataOpsSummary())
}

func handleAdminDataExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	sess := getSession(r)
	tahun := getTahunAnggaran()
	stamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("SIPKEU_backup_TA%d_%s.zip", tahun, stamp)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	meta := map[string]interface{}{
		"exported_at":    time.Now().Format(time.RFC3339),
		"tahun_anggaran": tahun,
		"exported_by":    "",
		"app":            "SIPKEU Dinas Pendidikan Kota Batam",
		"note":           "Cadangan data inputan — transaksi, anggaran/RAK, kas belanja, gaji ASN. Tidak termasuk password sistem.",
	}
	if sess != nil {
		meta["exported_by"] = sess.Username
	}
	if err := writeZipJSON(zw, "manifest.json", meta); err != nil {
		_ = zw.Close()
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal menyusun cadangan"})
		return
	}

	for _, id := range sipkeuDataModuleIDs() {
		mod := sipkeuModules[id]
		if mod == nil {
			continue
		}
		mod.mu.Lock()
		snap := moduleSnapshot{
			NextID: mod.nextID,
			Txs:    append([]Transaction(nil), mod.txs...),
			Settings: AppSettings{
				PA:               mod.settings.PA,
				Bendahara:        mod.settings.Bendahara,
				AnggaranKegiatan: cloneAnggaranMap(mod.settings.AnggaranKegiatan),
				Rak:              cloneRakRows(mod.settings.Rak),
				RakMeta:          mod.settings.RakMeta,
			},
		}
		mod.mu.Unlock()
		if snap.Settings.AnggaranKegiatan == nil {
			snap.Settings.AnggaranKegiatan = map[string]float64{}
		}
		if err := writeZipJSON(zw, "modules/"+id+".json", snap); err != nil {
			_ = zw.Close()
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal menyusun cadangan modul"})
			return
		}
		if err := writeZipJSON(zw, "modules/"+id+"-pejabat.json", pejabatSnapshot{
			PA:        snap.Settings.PA,
			Bendahara: snap.Settings.Bendahara,
		}); err != nil {
			_ = zw.Close()
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal menyusun cadangan pejabat"})
			return
		}
	}

	kasMu.RLock()
	kasCopy := kasState
	kasMu.RUnlock()
	if err := writeZipJSON(zw, "kas-belanja.json", kasCopy); err != nil {
		_ = zw.Close()
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal menyusun cadangan kas"})
		return
	}

	gajiMu.RLock()
	gajiCopy := gajiState
	gajiMu.RUnlock()
	if err := writeZipJSON(zw, "gaji-tunjangan.json", gajiCopy); err != nil {
		_ = zw.Close()
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal menyusun cadangan gaji"})
		return
	}

	if err := zw.Close(); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal menutup arsip cadangan"})
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())

	if sess != nil {
		recordAudit(sess.Username, "export_all_data", "pengaturan",
			fmt.Sprintf("Unduh cadangan data TA %d", tahun), clientIP(r))
	}
}

func writeZipJSON(zw *zip.Writer, name string, v any) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func handleAdminDataClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	var body struct {
		Confirm           string   `json:"confirm"`
		Modules           []string `json:"modules"`
		ClearTransactions bool     `json:"clear_transactions"`
		ClearAnggaran     bool     `json:"clear_anggaran"`
		ClearKas          bool     `json:"clear_kas"`
		ClearGaji         bool     `json:"clear_gaji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "JSON tidak valid"})
		return
	}
	if strings.ToUpper(strings.TrimSpace(body.Confirm)) != "HAPUS" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Ketik HAPUS untuk konfirmasi"})
		return
	}
	if !body.ClearTransactions && !body.ClearAnggaran && !body.ClearKas && !body.ClearGaji {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Pilih minimal satu jenis data yang dibersihkan"})
		return
	}

	allowedMods := map[string]bool{}
	for _, id := range sipkeuDataModuleIDs() {
		allowedMods[id] = true
	}
	wanted := map[string]bool{}
	if len(body.Modules) == 0 {
		for id := range allowedMods {
			wanted[id] = true
		}
	} else {
		for _, id := range body.Modules {
			id = strings.TrimSpace(strings.ToLower(id))
			if id == "" {
				continue
			}
			if id == "all" {
				for mid := range allowedMods {
					wanted[mid] = true
				}
				continue
			}
			if !allowedMods[id] {
				jsonResponse(w, http.StatusBadRequest, map[string]string{
					"error": "Portal tidak dikenal: " + id,
				})
				return
			}
			wanted[id] = true
		}
		if (body.ClearTransactions || body.ClearAnggaran) && len(wanted) == 0 {
			jsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "Pilih minimal satu portal SIPKEU",
			})
			return
		}
	}

	cleared := map[string]interface{}{}
	txDeleted, rakCleared := 0, 0

	for _, id := range sipkeuDataModuleIDs() {
		if !wanted[id] {
			continue
		}
		mod := sipkeuModules[id]
		if mod == nil {
			continue
		}
		mod.mu.Lock()
		info := map[string]int{}
		if body.ClearTransactions {
			info["transactions"] = len(mod.txs)
			txDeleted += len(mod.txs)
			mod.txs = []Transaction{}
			mod.nextID = 1
		}
		if body.ClearAnggaran {
			info["rak_rows"] = len(mod.settings.Rak)
			rakCleared += len(mod.settings.Rak)
			mod.settings.Rak = []RakRow{}
			mod.settings.AnggaranKegiatan = map[string]float64{}
			mod.settings.RakMeta = RakMeta{}
		}
		mod.mu.Unlock()
		if len(info) > 0 {
			persistModule(mod)
			cleared[id] = info
		}
	}

	if body.ClearKas {
		kasMu.Lock()
		info := map[string]int{"rak_rows": len(kasState.RakRows)}
		tahun := kasState.Tahun
		if tahun <= 0 {
			tahun = getTahunAnggaran()
		}
		kasState = KasBelanjaState{
			Tahun:           tahun,
			RakRows:         []RakBelanjaRow{},
			Realisasi:       map[string]map[string]float64{},
			SisaManual:      map[string]map[string]float64{},
			RealisasiLocked: map[string]bool{},
		}
		kasMu.Unlock()
		persistKasState()
		cleared["kas-belanja"] = info
	}

	if body.ClearGaji {
		gajiMu.Lock()
		info := map[string]int{"rekening": len(gajiState.Rekening)}
		tahun := gajiState.Tahun
		if tahun <= 0 {
			tahun = getTahunAnggaran()
		}
		gajiState = GajiTunjanganState{
			Tahun:           tahun,
			Pagu:            map[string]float64{},
			Pegawai:         map[string]int{},
			Rekening:        nil,
			RekeningCells:   map[string]map[string]GajiMonthCell{},
			KebutuhanManual: map[string]map[string]float64{},
			Cells:           map[string]map[string]GajiMonthCell{},
			RealisasiLocked: map[string]bool{},
		}
		ensureGajiCells(&gajiState)
		gajiMu.Unlock()
		persistGajiState()
		cleared["gaji-asn"] = info
	}

	sess := getSession(r)
	detail := fmt.Sprintf("tx=%d rak=%d kas=%v gaji=%v", txDeleted, rakCleared, body.ClearKas, body.ClearGaji)
	if sess != nil {
		recordAudit(sess.Username, "clear_data", "pengaturan", detail, clientIP(r))
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Data berhasil dibersihkan",
		"cleared": cleared,
		"summary": buildDataOpsSummary(),
	})
}

func handleAdminTahun(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		kasMu.RLock()
		kasTahun := kasState.Tahun
		kasMu.RUnlock()
		gajiMu.RLock()
		gajiTahun := gajiState.Tahun
		gajiMu.RUnlock()
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"tahun_anggaran": getTahunAnggaran(),
			"kas_tahun":      kasTahun,
			"gaji_tahun":     gajiTahun,
		})
	case http.MethodPut, http.MethodPost:
		var body struct {
			Tahun         int   `json:"tahun"`
			TahunAnggaran int   `json:"tahun_anggaran"`
			SyncKasGaji   *bool `json:"sync_kas_gaji"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "JSON tidak valid"})
			return
		}
		tahun := body.Tahun
		if tahun == 0 {
			tahun = body.TahunAnggaran
		}
		if tahun < 2000 || tahun > 2100 {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Tahun harus antara 2000–2100"})
			return
		}
		sync := true
		if body.SyncKasGaji != nil {
			sync = *body.SyncKasGaji
		}
		if err := setTahunAnggaran(tahun, sync); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		sess := getSession(r)
		if sess != nil {
			recordAudit(sess.Username, "change_tahun", "pengaturan",
				"Tahun anggaran diganti menjadi "+strconv.Itoa(tahun), clientIP(r))
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"ok":             true,
			"tahun_anggaran": tahun,
			"message":        fmt.Sprintf("Tahun anggaran SIPKEU diganti ke %d", tahun),
			"summary":        buildDataOpsSummary(),
		})
	default:
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func handlePublicTahun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"tahun_anggaran": getTahunAnggaran(),
	})
}
