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

func TestGajiRekeningLockIndependentPerGrup(t *testing.T) {
	state := GajiTunjanganState{
		Rekening: []GajiRekeningDef{{
			Kode: "5.1.01.01.009.00001", Nama: "Belanja Iuran Jaminan Kesehatan PNS",
			Grup: "gaji", Jenis: "pns", Potongan: true,
		}},
		RealisasiLocked: map[string]bool{
			gajiRekeningRowLockKey("tpg", "5.1.01.01.009.00001", "januari"): true,
		},
	}
	if !isGajiRekeningRowLocked(state, "tpg", "5.1.01.01.009.00001", "januari") {
		t.Fatal("expected tpg row locked")
	}
	if isGajiRekeningRowLocked(state, "tamsil", "5.1.01.01.009.00001", "januari") {
		t.Fatal("expected tamsil row not locked when only tpg saved")
	}
	unlockGajiRekeningMonth(&state, "tpg", "januari")
	if isGajiRekeningRowLocked(state, "tpg", "5.1.01.01.009.00001", "januari") {
		t.Fatal("expected tpg row unlocked after perbaiki bulan")
	}
}

func TestGajiRekeningSharedAnggaranAccumulatedSisa(t *testing.T) {
	def := GajiRekeningDef{
		Kode:     "5.1.01.01.009.00001",
		Nama:     "Belanja Iuran Jaminan Kesehatan PNS",
		Grup:     "gaji",
		Jenis:    "pns",
		Potongan: true,
		Pagu:     10_000_000,
	}
	state := GajiTunjanganState{
		Rekening: []GajiRekeningDef{def},
		RekeningCells: map[string]map[string]GajiMonthCell{
			def.Kode: {
				"januari": {Anggaran: 1_000_000},
			},
		},
	}
	ensureGajiRekening(&state)
	bulan := "januari"

	gajiSetRekeningCellForGrup(&state, "tpg", &def, def.Kode, bulan, GajiMonthCell{Realisasi: 400_000})
	gajiSetRekeningCellForGrup(&state, "tamsil", &def, def.Kode, bulan, GajiMonthCell{Realisasi: 250_000})

	tpgCell := gajiGetRekeningCellForGrup(state, "tpg", def, bulan)
	if tpgCell.Realisasi != 400_000 {
		t.Fatalf("expected tpg menu realisasi 400000, got %v", tpgCell.Realisasi)
	}
	if tpgCell.Anggaran != 1_000_000 {
		t.Fatalf("expected shared anggaran 1000000, got %v", tpgCell.Anggaran)
	}
	total := gajiSumRekeningRealisasiAllGrups(state, def, bulan)
	if total != 650_000 {
		t.Fatalf("expected accumulated realisasi 650000, got %v", total)
	}
	sisa := tpgCell.Anggaran - total
	if sisa != 350_000 {
		t.Fatalf("expected accumulated sisa 350000, got %v", sisa)
	}

	rows, _ := buildGajiRekeningReport(state, "tpg", bulan)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row in tpg report, got %d", len(rows))
	}
	if rows[0].Realisasi != 400_000 {
		t.Fatalf("expected row realisasi menu 400000, got %v", rows[0].Realisasi)
	}
	if rows[0].Sisa != 350_000 {
		t.Fatalf("expected row sisa 350000, got %v", rows[0].Sisa)
	}
}

func TestGajiSyncCategoryDashboardRealisasi(t *testing.T) {
	def := GajiRekeningDef{
		Kode: "5.1.01.02.001.00001", Nama: "Gaji PNS", Grup: "gaji", Jenis: "pns", Pagu: 100_000_000,
	}
	state := GajiTunjanganState{
		Rekening:      []GajiRekeningDef{def},
		RekeningCells: map[string]map[string]GajiMonthCell{},
	}
	ensureGajiRekening(&state)
	gajiSetRekeningCellForGrup(&state, "gaji", &def, def.Kode, "januari", GajiMonthCell{Anggaran: 10_000_000, Realisasi: 3_500_000})
	gajiSyncCategoryFromRekening(&state)
	dash := buildGajiDashboard(state, "januari")
	if dash["total_realisasi_sd"].(float64) != 3_500_000 {
		t.Fatalf("expected dashboard realisasi 3500000, got %v", dash["total_realisasi_sd"])
	}
}
