package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type moduleSnapshot struct {
	NextID   int           `json:"next_id"`
	Txs      []Transaction `json:"transactions"`
	Settings AppSettings   `json:"settings"`
}

var (
	dataDir   string
	storageMu sync.Mutex
)

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

func kasDataPath() string {
	return filepath.Join(dataDir, "kas-belanja.json")
}

func writeJSONAtomic(path string, v any) error {
	storageMu.Lock()
	defer storageMu.Unlock()

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
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

func loadModuleFromDisk(mod *SipkeuModule) bool {
	path := moduleDataPath(mod.ID)
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var snap moduleSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		log.Printf("Peringatan: file data %s rusak: %v", path, err)
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
	if snap.Settings.PA.Nama != "" {
		mod.settings.PA = snap.Settings.PA
	}
	if snap.Settings.Bendahara.Nama != "" {
		mod.settings.Bendahara = snap.Settings.Bendahara
	}
	mod.mu.Unlock()
	normalizeModuleIDs(mod)
	log.Printf("Data modul %s dimuat dari %s (%d transaksi)", mod.ID, path, len(snap.Txs))
	return true
}

func persistModule(mod *SipkeuModule) {
	mod.mu.Lock()
	snap := moduleSnapshot{
		NextID:   mod.nextID,
		Txs:      append([]Transaction(nil), mod.txs...),
		Settings: mod.settings,
	}
	mod.mu.Unlock()

	if snap.Settings.AnggaranKegiatan == nil {
		snap.Settings.AnggaranKegiatan = map[string]float64{}
	}
	if err := writeJSONAtomic(moduleDataPath(mod.ID), snap); err != nil {
		log.Printf("Gagal simpan modul %s: %v", mod.ID, err)
	}
}

func loadAllModulesFromDisk() {
	sipkeuModulesMu.RLock()
	defer sipkeuModulesMu.RUnlock()
	for _, mod := range sipkeuModules {
		loadModuleFromDisk(mod)
	}
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

func storageInfo() string {
	return fmt.Sprintf("DATA_DIR=%s", dataDir)
}
