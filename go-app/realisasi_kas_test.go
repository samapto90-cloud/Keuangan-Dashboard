package main

import "testing"

func TestMergeKasRakPreservingLockedMonths(t *testing.T) {
	old := []RakBelanjaRow{
		{
			KodeRekening: "5.1.01.",
			Anggaran:     100,
			Bulan: map[string]float64{
				"januari": 10, "februari": 10, "maret": 20,
			},
		},
	}
	newRows := []RakBelanjaRow{
		{
			KodeRekening: "5.1.01.",
			Anggaran:     200,
			Bulan: map[string]float64{
				"januari": 99, "februari": 99, "maret": 50,
			},
		},
	}
	locked := map[string]bool{"januari": true, "februari": true}
	got := mergeKasRakPreservingLockedMonths(old, newRows, locked)
	if len(got) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got))
	}
	if got[0].Anggaran != 200 {
		t.Fatalf("pagu harus dari RAK baru, got %v", got[0].Anggaran)
	}
	if got[0].Bulan["januari"] != 10 {
		t.Fatalf("januari terkunci harus 10, got %v", got[0].Bulan["januari"])
	}
	if got[0].Bulan["februari"] != 10 {
		t.Fatalf("februari terkunci harus 10, got %v", got[0].Bulan["februari"])
	}
	if got[0].Bulan["maret"] != 50 {
		t.Fatalf("maret belum terkunci harus 50, got %v", got[0].Bulan["maret"])
	}
}

func TestBuildKasReportRangeTW1(t *testing.T) {
	state := KasBelanjaState{
		Tahun: 2026,
		RakRows: []RakBelanjaRow{{
			KodeRekening: "5.1.01.",
			Bulan: map[string]float64{
				"januari": 100, "februari": 100, "maret": 100,
			},
		}},
		Realisasi: map[string]map[string]float64{
			"januari":  {"5.1.01.": 40},
			"februari": {"5.1.01.": 30},
			"maret":    {"5.1.01.": 20},
		},
	}
	report := buildKasReportRange(state, 0, 2)
	row := findKasReportRow(report, "5.1.01.")
	if row == nil {
		t.Fatal("missing 5.1.01.")
	}
	if row.AnggaranKas != 300 {
		t.Fatalf("anggaran TW1 want 300, got %v", row.AnggaranKas)
	}
	if row.Realisasi != 90 {
		t.Fatalf("realisasi TW1 want 90, got %v", row.Realisasi)
	}
	if row.SisaSD != 210 {
		t.Fatalf("sisa TW1 want 210, got %v", row.SisaSD)
	}
}

func TestBuildKasReportRangeSem2Carry(t *testing.T) {
	state := KasBelanjaState{
		RakRows: []RakBelanjaRow{{
			KodeRekening: "5.1.01.",
			Bulan: map[string]float64{
				"juni": 50, "juli": 50,
			},
		}},
		Realisasi: map[string]map[string]float64{
			"juni": {"5.1.01.": 50},
			"juli": {"5.1.01.": 10},
		},
	}
	report := buildKasReportRange(state, 6, 7)
	row := findKasReportRow(report, "5.1.01.")
	if row == nil {
		t.Fatal("missing 5.1.01.")
	}
	if row.SisaBulanLalu != 0 {
		t.Fatalf("sisa awal jul want 0 (sisa akhir jun), got %v", row.SisaBulanLalu)
	}
	if row.AnggaranKas != 50 {
		t.Fatalf("anggaran jul want 50, got %v", row.AnggaranKas)
	}
	if row.Realisasi != 10 {
		t.Fatalf("realisasi jul want 10, got %v", row.Realisasi)
	}
	if row.SisaSD != 40 {
		t.Fatalf("sisa akhir jul want 40, got %v", row.SisaSD)
	}
}

func findKasReportRow(rows []KasReportRow, kode string) *KasReportRow {
	for i := range rows {
		if rows[i].Kode == kode {
			return &rows[i]
		}
	}
	return nil
}

func TestBuildKasPeriodMatchesMonthlySum(t *testing.T) {
	state := KasBelanjaState{
		RakRows: []RakBelanjaRow{{
			KodeRekening: "5.1.01.",
			Bulan: map[string]float64{
				"januari": 100, "februari": 80, "maret": 120,
			},
		}},
		Realisasi: map[string]map[string]float64{
			"januari":  {"5.1.01.": 40},
			"februari": {"5.1.01.": 30},
			"maret":    {"5.1.01.": 50},
		},
	}
	bundle := buildKasPeriodResult(state, 0, 2)
	if !verifyKasPeriodMonthlySum(bundle.MonthlyReports, bundle.Report) {
		t.Fatal("period report must equal sum of monthly rows")
	}
	if len(bundle.MonthlyReports) != 3 {
		t.Fatalf("expected 3 monthly snaps, got %d", len(bundle.MonthlyReports))
	}
	row := findKasReportRow(bundle.Report, "5.1.01.")
	if row == nil || row.AnggaranKas != 300 || row.Realisasi != 120 {
		t.Fatalf("aggregated row mismatch: %+v", row)
	}
}

func TestMergeKasRakNoLockedUsesNew(t *testing.T) {
	old := []RakBelanjaRow{{KodeRekening: "5.1.01.", Bulan: map[string]float64{"januari": 1}}}
	newRows := []RakBelanjaRow{{KodeRekening: "5.1.01.", Bulan: map[string]float64{"januari": 9}}}
	got := mergeKasRakPreservingLockedMonths(old, newRows, nil)
	if got[0].Bulan["januari"] != 9 {
		t.Fatalf("tanpa kunci harus pakai RAK baru, got %v", got[0].Bulan["januari"])
	}
}
