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

func TestGajiRekeningCellIndependentPerGrup(t *testing.T) {
	def := GajiRekeningDef{
		Kode:     "5.1.01.01.009.00001",
		Nama:     "Belanja Iuran Jaminan Kesehatan PNS",
		Grup:     "gaji",
		Jenis:    "pns",
		Potongan: true,
	}
	state := GajiTunjanganState{
		Rekening:      []GajiRekeningDef{def},
		RekeningCells: map[string]map[string]GajiMonthCell{},
	}
	ensureGajiRekening(&state)
	bulan := "januari"

	tpgCell := GajiMonthCell{Anggaran: 1_000_000, Realisasi: 400_000}
	gajiSetRekeningCellForGrup(&state, "tpg", &def, def.Kode, bulan, tpgCell)

	tamsilCell := GajiMonthCell{Anggaran: 1_000_000, Realisasi: 250_000}
	gajiSetRekeningCellForGrup(&state, "tamsil", &def, def.Kode, bulan, tamsilCell)

	gotTPG := gajiGetRekeningCellForGrup(state, "tpg", def, bulan)
	gotTamsil := gajiGetRekeningCellForGrup(state, "tamsil", def, bulan)
	if gotTPG.Realisasi != 400_000 {
		t.Fatalf("expected tpg realisasi 400000, got %v", gotTPG.Realisasi)
	}
	if gotTamsil.Realisasi != 250_000 {
		t.Fatalf("expected tamsil realisasi 250000, got %v", gotTamsil.Realisasi)
	}
	if sisa := gotTPG.Anggaran - gotTPG.Realisasi; sisa != 600_000 {
		t.Fatalf("expected tpg sisa 600000, got %v", sisa)
	}
}
