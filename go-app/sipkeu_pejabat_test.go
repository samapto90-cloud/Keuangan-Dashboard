package main

import "testing"

func TestPejabatNamesMatchExactOnly(t *testing.T) {
	if !pejabatNamesMatch("HENDRI ARULAN, S.Pd", "hendri arulan, s.pd") {
		t.Fatal("exact match (case-insensitive) should succeed")
	}
	if pejabatNamesMatch("HENDRI ARULAN, S.Pd", "HENDRI ARULAN") {
		t.Fatal("partial/fuzzy match must not succeed")
	}
	if pejabatNamesMatch("BUDI SANTOSO", "BUDI") {
		t.Fatal("token overlap must not count as match")
	}
}

func TestEffectivePejabatKeepsAdminRename(t *testing.T) {
	defPA := Pejabat{Nama: "DEFAULT PA", Nip: "NIP. 1"}
	defBend := Pejabat{Nama: "DEFAULT BEND", Nip: "NIP. 2"}
	savedPA := Pejabat{Nama: "NAMA BARU PA", Nip: "NIP. 999"}
	savedBend := Pejabat{Nama: "NAMA BARU BEND", Nip: "NIP. 888"}

	pa, bend := effectivePejabatValues("sekretariat", savedPA, savedBend, defPA, defBend)
	if pa.Nama != savedPA.Nama || pa.Nip != savedPA.Nip {
		t.Fatalf("PA should keep admin values, got %+v", pa)
	}
	if bend.Nama != savedBend.Nama || bend.Nip != savedBend.Nip {
		t.Fatalf("Bendahara should keep admin values, got %+v", bend)
	}

	emptyPA, emptyBend := effectivePejabatValues("sekretariat", Pejabat{}, Pejabat{}, defPA, defBend)
	if emptyPA.Nama != defPA.Nama || emptyBend.Nama != defBend.Nama {
		t.Fatalf("empty pejabat should fall back to defaults")
	}
}

func TestPejabatSettingsPersistRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dataDir = dir
	initSipkeuModules()
	mod := sipkeuModules["sekretariat"]

	mod.mu.Lock()
	mod.settings.PA = Pejabat{Nama: "NAMA BARU PA", Nip: "NIP. 999"}
	mod.settings.Bendahara = Pejabat{Nama: "NAMA BARU BEND", Nip: "NIP. 888"}
	pa, bend := mod.settings.PA, mod.settings.Bendahara
	mod.mu.Unlock()
	if err := persistPejabat(mod, pa, bend); err != nil {
		t.Fatalf("persistPejabat: %v", err)
	}
	if err := persistModule(mod); err != nil {
		t.Fatalf("persistModule: %v", err)
	}

	reloaded := newSipkeuModule("sekretariat", "B01", mod.defaultSettings.PA, mod.defaultSettings.Bendahara)
	if !loadModuleFromDisk(reloaded) {
		t.Fatal("expected module data file to load")
	}
	reloaded.mu.Lock()
	gotPA := reloaded.settings.PA
	gotBend := reloaded.settings.Bendahara
	reloaded.mu.Unlock()
	if gotPA.Nama != "NAMA BARU PA" || gotPA.Nip != "NIP. 999" {
		t.Fatalf("PA not persisted, got %+v", gotPA)
	}
	if gotBend.Nama != "NAMA BARU BEND" || gotBend.Nip != "NIP. 888" {
		t.Fatalf("Bendahara not persisted, got %+v", gotBend)
	}
}

func TestPejabatFileWinsOverStaleModuleJSON(t *testing.T) {
	dir := t.TempDir()
	dataDir = dir
	initSipkeuModules()
	mod := sipkeuModules["sekretariat"]

	// Simulasikan race: module.json punya pejabat lama, file pejabat punya yang baru.
	mod.mu.Lock()
	mod.settings.PA = Pejabat{Nama: "LAMA PA", Nip: "1"}
	mod.settings.Bendahara = Pejabat{Nama: "LAMA BEND", Nip: "2"}
	mod.mu.Unlock()
	if err := persistModule(mod); err != nil {
		t.Fatal(err)
	}
	if err := persistPejabat(mod, Pejabat{Nama: "BARU PA", Nip: "9"}, Pejabat{Nama: "BARU BEND", Nip: "8"}); err != nil {
		t.Fatal(err)
	}

	reloaded := newSipkeuModule("sekretariat", "B01", mod.defaultSettings.PA, mod.defaultSettings.Bendahara)
	if !loadModuleFromDisk(reloaded) {
		t.Fatal("load failed")
	}
	reloaded.mu.Lock()
	got := reloaded.settings.PA.Nama
	reloaded.mu.Unlock()
	if got != "BARU PA" {
		t.Fatalf("pejabat file should win, got %q", got)
	}
}

func TestStampCurrentPejabatOnEditUsesSettings(t *testing.T) {
	initSipkeuModules()
	mod := sipkeuModules["sekretariat"]
	mod.mu.Lock()
	mod.settings.PA = Pejabat{Nama: "PA BARU", Nip: "NIP. 111"}
	mod.settings.Bendahara = Pejabat{Nama: "BEND BARU", Nip: "NIP. 222"}
	mod.mu.Unlock()

	trx := Transaction{
		PenggunaAnggaran:    "PA LAMA",
		PenggunaAnggaranNip: "NIP. OLD",
		Bendahara:           "BEND LAMA",
		BendaharaNip:        "NIP. OLD2",
	}
	stampCurrentPejabatOnEdit(mod, &trx)
	if trx.PenggunaAnggaran != "PA BARU" || trx.PenggunaAnggaranNip != "NIP. 111" {
		t.Fatalf("PA should follow current settings, got %q / %q", trx.PenggunaAnggaran, trx.PenggunaAnggaranNip)
	}
	if trx.Bendahara != "BEND BARU" || trx.BendaharaNip != "NIP. 222" {
		t.Fatalf("Bendahara should follow current settings, got %q / %q", trx.Bendahara, trx.BendaharaNip)
	}
}

func TestRepairModuleIsolationKeepsCustomPejabat(t *testing.T) {
	initSipkeuModules()
	mod := sipkeuModules["sekretariat"]
	otherDefault := sipkeuModules["paud"].defaultSettings.PA

	mod.mu.Lock()
	mod.settings.PA = otherDefault
	mod.mu.Unlock()

	if repairModuleIsolation(mod) {
		t.Fatal("repair should not reset admin/custom pejabat")
	}
	mod.mu.Lock()
	got := mod.settings.PA
	mod.mu.Unlock()
	if got.Nama != otherDefault.Nama {
		t.Fatalf("pejabat changed after repair, got %+v", got)
	}
}
