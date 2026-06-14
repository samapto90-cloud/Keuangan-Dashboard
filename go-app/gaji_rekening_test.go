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

func TestGajiSyncCategoryTPGRekapRealisasi(t *testing.T) {
	def := GajiRekeningDef{
		Kode: "5.1.01.02.006.00071", Nama: "TPG PNS", Grup: "tpg", Jenis: "pns", Pagu: 50_000_000,
	}
	state := GajiTunjanganState{
		Rekening:      []GajiRekeningDef{def},
		RekeningCells: map[string]map[string]GajiMonthCell{},
	}
	ensureGajiRekening(&state)
	gajiSetRekeningCellForGrup(&state, "tpg", &def, def.Kode, "januari", GajiMonthCell{Realisasi: 100_000})
	gajiSetRekeningCellForGrup(&state, "tpg", &def, def.Kode, "februari", GajiMonthCell{Realisasi: 200_000})
	gajiSyncCategoryFromRekening(&state)
	cell := gajiGetCell(state, "tpg_pns", "tw1")
	if cell.Realisasi != 300_000 {
		t.Fatalf("expected tw1 realisasi 300000, got %v", cell.Realisasi)
	}
	rekap := buildGajiRekap(state)
	var janRow *GajiRekapBulanRow
	for i := range rekap {
		if rekap[i].Bulan == "januari" {
			janRow = &rekap[i]
			break
		}
	}
	if janRow == nil {
		t.Fatal("expected januari rekap row")
	}
	if janRow.Categories["tpg_pns"].Realisasi != 300_000 {
		t.Fatalf("expected rekap tpg_pns januari realisasi 300000, got %v", janRow.Categories["tpg_pns"].Realisasi)
	}
	dash := buildGajiDashboard(state, "juni")
	var tpgSum float64
	for _, c := range dash["category_summaries"].([]GajiCategorySummary) {
		if c.CategoryID == "tpg_pns" {
			tpgSum = c.RealisasiSD
		}
	}
	if tpgSum != 300_000 {
		t.Fatalf("expected dashboard tpg_pns sd 300000, got %v", tpgSum)
	}
}

func TestGajiRekeningMatrixMultiGrup(t *testing.T) {
	defGaji := GajiRekeningDef{Kode: "5.1.01.02.001.00001", Nama: "Gaji PNS", Grup: "gaji", Jenis: "pns", Pagu: 10_000_000}
	defTpp := GajiRekeningDef{Kode: "5.1.01.02.002.00001", Nama: "TPP PNS", Grup: "tpp", Jenis: "pns", Pagu: 5_000_000}
	state := GajiTunjanganState{
		Rekening:      []GajiRekeningDef{defGaji, defTpp},
		RekeningCells: map[string]map[string]GajiMonthCell{},
	}
	ensureGajiRekening(&state)
	gajiSetRekeningCellForGrup(&state, "gaji", &defGaji, defGaji.Kode, "januari", GajiMonthCell{Realisasi: 100_000})
	gajiSetRekeningCellForGrup(&state, "tpp", &defTpp, defTpp.Kode, "januari", GajiMonthCell{Realisasi: 50_000})
	rows, summary := buildGajiRekeningMatrixMulti(state, []string{"gaji", "tpp"}, "januari")
	if len(rows) != 2 {
		t.Fatalf("expected 2 matrix rows, got %d", len(rows))
	}
	if summary.Bulan["januari"].Realisasi != 150_000 {
		t.Fatalf("expected summary realisasi 150000, got %v", summary.Bulan["januari"].Realisasi)
	}
}

func TestGajiRekeningMonthlyTotals(t *testing.T) {
	defGaji := GajiRekeningDef{Kode: "5.1.01.02.001.00001", Nama: "Gaji PNS", Grup: "gaji", Jenis: "pns", Pagu: 100_000_000}
	defTpp := GajiRekeningDef{Kode: "5.1.01.02.002.00001", Nama: "TPP PNS", Grup: "tpp", Jenis: "pns", Pagu: 50_000_000}
	state := GajiTunjanganState{
		Rekening: []GajiRekeningDef{defGaji, defTpp},
		RekeningCells: map[string]map[string]GajiMonthCell{
			defGaji.Kode: {"januari": {Anggaran: 10_000_000, Realisasi: 3_000_000}, "februari": {Anggaran: 10_000_000}},
			defTpp.Kode:  {"januari": {Anggaran: 5_000_000, Realisasi: 1_500_000}},
		},
	}
	totals := buildGajiRekeningMonthlyTotals(state)
	if len(totals) != 12 {
		t.Fatalf("expected 12 months, got %d", len(totals))
	}
	if totals[0].Anggaran != 15_000_000 {
		t.Fatalf("jan anggaran = %v want 15000000", totals[0].Anggaran)
	}
	if totals[0].Realisasi != 4_500_000 {
		t.Fatalf("jan realisasi = %v want 4500000", totals[0].Realisasi)
	}
	if totals[1].Anggaran != 10_000_000 {
		t.Fatalf("feb anggaran = %v want 10000000", totals[1].Anggaran)
	}
}

func TestGajiRekeningPekerjaanSeries(t *testing.T) {
	state := GajiTunjanganState{
		Rekening: []GajiRekeningDef{{
			Kode: "5.1.01.02.001.00001", Nama: "Belanja Gaji Pokok PNS",
			Grup: "gaji", Jenis: "pns", Pagu: 100_000_000,
			Kegiatan: "Penyediaan Gaji dan Tunjangan ASN",
			SubKegiatan: "Penyediaan Gaji dan Tunjangan ASN Pemerintah",
		}},
		RekeningCells: map[string]map[string]GajiMonthCell{
			"5.1.01.02.001.00001": {"januari": {Anggaran: 10_000_000, Realisasi: 2_000_000}},
		},
	}
	series := buildGajiRekeningPekerjaanSeries(state)
	if len(series) != 1 {
		t.Fatalf("expected 1 pekerjaan row, got %d", len(series))
	}
	if series[0].Pekerjaan != "Belanja Gaji Pokok PNS" {
		t.Fatalf("unexpected pekerjaan label %q", series[0].Pekerjaan)
	}
	meta := gajiProgramMetaFromRekening(state)
	if meta.Kegiatan != "Penyediaan Gaji dan Tunjangan ASN" {
		t.Fatalf("unexpected kegiatan %q", meta.Kegiatan)
	}
}
