package main

import "testing"

func TestGajiRekeningAttachedJaminanKesPPPK(t *testing.T) {
	def := GajiRekeningDef{
		Kode:     "5.1.01.01.009.00002",
		Nama:     "Belanja Iuran Jaminan Kesehatan PPPK",
		Grup:     "gaji",
		Jenis:    "pppk",
		Potongan: true,
	}
	for _, grup := range []string{"tpg", "tamsil"} {
		if !gajiRekeningAttachedJaminanKes(def, grup) {
			t.Fatalf("expected jaminan kesehatan PPPK attached to %s", grup)
		}
		if !gajiRekeningIncludedInGrup(def, grup) {
			t.Fatalf("expected jaminan kesehatan PPPK included in %s report", grup)
		}
	}
}

func TestGajiRekeningAttachedJaminanKesPNS(t *testing.T) {
	def := GajiRekeningDef{
		Kode:     "5.1.01.01.009.00001",
		Nama:     "Belanja Iuran Jaminan Kesehatan PNS",
		Grup:     "gaji",
		Jenis:    "pns",
		Potongan: true,
	}
	for _, grup := range []string{"tpg", "tamsil"} {
		if !gajiRekeningAttachedJaminanKes(def, grup) {
			t.Fatalf("expected jaminan kesehatan PNS attached to %s", grup)
		}
	}
}

func TestGajiJenisFromKodeJaminanKes(t *testing.T) {
	if gajiJenisFromKodeJaminanKes("5.1.01.01.009.00001") != "pns" {
		t.Fatal("expected pns from kode suffix 00001")
	}
	if gajiJenisFromKodeJaminanKes("5.1.01.01.009.00002") != "pppk" {
		t.Fatal("expected pppk from kode suffix 00002")
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
