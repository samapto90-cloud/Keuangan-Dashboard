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
