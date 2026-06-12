package main

import "testing"

func TestGajiRekeningAttachedJaminanKesTamsil(t *testing.T) {
	def := GajiRekeningDef{
		Kode:     "5.1.01.01.009.00002",
		Nama:     "Belanja Iuran Jaminan Kesehatan PPPK",
		Grup:     "gaji",
		Jenis:    "pppk",
		Potongan: true,
	}
	if !gajiRekeningAttachedJaminanKes(def, "tamsil") {
		t.Fatal("expected jaminan kesehatan PPPK attached to tamsil")
	}
	if gajiRekeningAttachedJaminanKes(def, "tpg") {
		t.Fatal("PPPK jaminan should not attach to tpg")
	}
}

func TestGajiRekeningIncludedInGrupTamsilBenefit(t *testing.T) {
	def := GajiRekeningDef{
		Kode:  "5.1.01.02.006.00072",
		Nama:  "Belanja Tambahan Penghasilan (Tamsil) Guru PPPK",
		Grup:  "tamsil",
		Jenis: "pppk",
	}
	if !gajiRekeningIncludedInGrup(def, "tamsil") {
		t.Fatal("expected tamsil benefit in tamsil grup")
	}
}
