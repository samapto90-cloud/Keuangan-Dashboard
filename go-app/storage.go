package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type moduleSnapshot struct {
	NextID   int           `json:"next_id"`
	Txs      []Transaction `json:"transactions"`
	Settings AppSettings   `json:"settings"`
}

var dataDir string

func initStorage() {
	dataDir = os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Printf("Peringatan: tidak bisa buat folder data (%s): %v", dataDir, err)
	}
}

func moduleDataPath(id string) string {
	return filepath.Join(dataDir, id+".json")
}

func pejabatDataPath(id string) string {
	return filepath.Join(dataDir, id+"-pejabat.json")
}

type pejabatSnapshot struct {
	PA        Pejabat `json:"pa"`
	Bendahara Pejabat `json:"bendahara"`
}

func kasDataPath() string {
	return filepath.Join(dataDir, "kas-belanja.json")
}

func gajiDataPath() string {
	return filepath.Join(dataDir, "gaji-tunjangan.json")
}

func writeJSONAtomic(path string, v any) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	if !jsonCompactEnabled() {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func jsonCompactEnabled() bool {
	return strings.TrimSpace(os.Getenv("SIPKEU_COMPACT_JSON")) == "1"
}

func loadModuleFromDisk(mod *SipkeuModule) bool {
	path := moduleDataPath(mod.ID)
	raw, err := os.ReadFile(path)
	if err != nil {
		// Tetap coba pejabat file jika modul belum punya json utama.
		loadPejabatFromDisk(mod)
		return false
	}
	var snap moduleSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		log.Printf("Peringatan: file data %s rusak: %v", path, err)
		loadPejabatFromDisk(mod)
		return false
	}
	mod.mu.Lock()
	mod.nextID = snap.NextID
	if mod.nextID <= 0 {
		mod.nextID = 1
	}
	mod.txs = snap.Txs
	if mod.txs == nil {
		mod.txs = []Transaction{}
	}
	if snap.Settings.AnggaranKegiatan != nil {
		mod.settings.AnggaranKegiatan = snap.Settings.AnggaranKegiatan
	}
	if len(snap.Settings.Rak) > 0 {
		mod.settings.Rak = snap.Settings.Rak
	}
	mod.settings.RakMeta = snap.Settings.RakMeta
	if snap.Settings.PA.Nama != "" {
		mod.settings.PA = snap.Settings.PA
	}
	if snap.Settings.Bendahara.Nama != "" {
		mod.settings.Bendahara = snap.Settings.Bendahara
	}
	mod.mu.Unlock()
	normalizeModuleIDs(mod)
	// File pejabat khusus menang — tidak boleh tertimpa race persist transaksi.
	loadPejabatFromDisk(mod)
	log.Printf("Data modul %s dimuat dari %s (%d transaksi)", mod.ID, path, len(snap.Txs))
	return true
}

func loadPejabatFromDisk(mod *SipkeuModule) bool {
	if mod == nil {
		return false
	}
	raw, err := os.ReadFile(pejabatDataPath(mod.ID))
	if err != nil {
		return false
	}
	var snap pejabatSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		log.Printf("Peringatan: file pejabat %s rusak: %v", pejabatDataPath(mod.ID), err)
		return false
	}
	mod.mu.Lock()
	if strings.TrimSpace(snap.PA.Nama) != "" {
		mod.settings.PA = snap.PA
	}
	if strings.TrimSpace(snap.Bendahara.Nama) != "" {
		mod.settings.Bendahara = snap.Bendahara
	}
	paName := mod.settings.PA.Nama
	bendName := mod.settings.Bendahara.Nama
	mod.mu.Unlock()
	log.Printf("Pejabat modul %s dimuat: PA=%s Bendahara=%s", mod.ID, paName, bendName)
	return true
}

func persistPejabat(mod *SipkeuModule, pa, bend Pejabat) error {
	if mod == nil {
		return fmt.Errorf("modul kosong")
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	snap := pejabatSnapshot{PA: pa, Bendahara: bend}
	if err := writeJSONAtomic(pejabatDataPath(mod.ID), snap); err != nil {
		return err
	}
	invalidateSettingsCache(mod.ID)
	return nil
}

func persistModule(mod *SipkeuModule) error {
	if mod == nil {
		return fmt.Errorf("modul kosong")
	}
	// Serialisasi seluruh tulis modul agar snapshot pejabat tidak tertimpa race.
	mod.persistMu.Lock()
	defer mod.persistMu.Unlock()

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
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Printf("Gagal simpan modul %s: %v", mod.ID, err)
		return err
	}
	if err := writeJSONAtomic(moduleDataPath(mod.ID), snap); err != nil {
		log.Printf("Gagal simpan modul %s: %v", mod.ID, err)
		return err
	}
	invalidateDashboardCache(mod.ID)
	invalidateTransactionsCache(mod.ID)
	invalidateSettingsCache(mod.ID)
	invalidateAdminRekapCache()
	invalidateRealisasiCache(mod.ID)
	return nil
}

func loadAllModulesFromDisk() {
	sipkeuModulesMu.RLock()
	defer sipkeuModulesMu.RUnlock()
	for _, mod := range sipkeuModules {
		loadModuleFromDisk(mod)
		ensurePejabatFile(mod)
	}
}

// ensurePejabatFile: migrasi pejabat dari module.json ke file khusus jika belum ada.
func ensurePejabatFile(mod *SipkeuModule) {
	if mod == nil {
		return
	}
	if _, err := os.Stat(pejabatDataPath(mod.ID)); err == nil {
		return
	}
	mod.mu.Lock()
	pa, bend := mod.settings.PA, mod.settings.Bendahara
	mod.mu.Unlock()
	if strings.TrimSpace(pa.Nama) == "" && strings.TrimSpace(bend.Nama) == "" {
		return
	}
	if err := persistPejabat(mod, pa, bend); err != nil {
		log.Printf("Peringatan: gagal migrasi pejabat modul %s: %v", mod.ID, err)
		return
	}
	log.Printf("Pejabat modul %s dimigrasi ke file khusus", mod.ID)
}

func moduleHasData(mod *SipkeuModule) bool {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	return len(mod.txs) > 0 || len(mod.settings.Rak) > 0
}

func loadKasFromDisk() bool {
	raw, err := os.ReadFile(kasDataPath())
	if err != nil {
		return false
	}
	var state KasBelanjaState
	if err := json.Unmarshal(raw, &state); err != nil {
		log.Printf("Peringatan: file kas rusak: %v", err)
		return false
	}
	kasMu.Lock()
	kasState = state
	if kasState.Realisasi == nil {
		kasState.Realisasi = map[string]map[string]float64{}
	}
	if kasState.SisaManual == nil {
		kasState.SisaManual = map[string]map[string]float64{}
	}
	if kasState.RealisasiLocked == nil {
		kasState.RealisasiLocked = map[string]bool{}
	}
	kasMu.Unlock()
	log.Printf("Data kas belanja dimuat dari %s", kasDataPath())
	return true
}

func persistKasState() {
	kasMu.RLock()
	state := kasState
	kasMu.RUnlock()
	if err := writeJSONAtomic(kasDataPath(), state); err != nil {
		log.Printf("Gagal simpan kas belanja: %v", err)
	}
}

func loadGajiFromDisk() bool {
	raw, err := os.ReadFile(gajiDataPath())
	if err != nil {
		return false
	}
	var state GajiTunjanganState
	if err := json.Unmarshal(raw, &state); err != nil {
		log.Printf("Peringatan: file gaji tunjangan rusak: %v", err)
		return false
	}
	gajiMu.Lock()
	gajiState = state
	ensureGajiCells(&gajiState)
	if gajiState.RealisasiLocked == nil {
		gajiState.RealisasiLocked = map[string]bool{}
	}
	if gajiState.Pagu == nil {
		gajiState.Pagu = map[string]float64{}
	}
	if gajiState.Pegawai == nil {
		gajiState.Pegawai = map[string]int{}
	}
	normalizeGajiRekeningGrups(&gajiState)
	migrateGajiAttachedRekeningCells(&gajiState)
	gajiSyncCategoryFromRekening(&gajiState)
	gajiMu.Unlock()
	log.Printf("Data gaji tunjangan dimuat dari %s", gajiDataPath())
	return true
}

func persistGajiState() {
	gajiMu.RLock()
	state := gajiState
	gajiMu.RUnlock()
	if err := writeJSONAtomic(gajiDataPath(), state); err != nil {
		log.Printf("Gagal simpan gaji tunjangan: %v", err)
	}
}

func storageInfo() string {
	return fmt.Sprintf("DATA_DIR=%s", dataDir)
}
