package main

import "testing"

func TestGajiSplitRealisasiKebutuhanGajiPNS(t *testing.T) {
	state := GajiTunjanganState{
		Pagu: map[string]float64{"gaji_pns": 100758404821},
		Cells: map[string]map[string]GajiMonthCell{
			"gaji_pns": {
				"januari": {Realisasi: 8002177160},
				"februari": {Realisasi: 7942207060},
				"maret": {Realisasi: 7944455460},
				"thr": {Realisasi: 7942207060},
				"april": {Realisasi: 7902663660},
				"gaji_13": {Realisasi: 7880759000},
				"juli": {Realisasi: 7880759000},
			},
		},
	}
	sd, sisa, tahun := gajiSplitRealisasiKebutuhan(state, "gaji_pns", "april")
	wantSD := 8002177160.0 + 7942207060 + 7944455460 + 7942207060 + 7902663660
	if sd != wantSD {
		t.Fatalf("realisasi sd = %v want %v", sd, wantSD)
	}
	if sisa <= 0 {
		t.Fatalf("kebutuhan sisa should be > 0, got %v", sisa)
	}
	if tahun != sd+sisa {
		t.Fatalf("kebutuhan tahun = %v want %v", tahun, sd+sisa)
	}
}

func TestGajiKebutuhanRowSelisih(t *testing.T) {
	state := GajiTunjanganState{
		Pagu: map[string]float64{"gaji_pns": 100000000},
		Cells: map[string]map[string]GajiMonthCell{
			"gaji_pns": {
				"januari": {Realisasi: 10000000},
				"februari": {Realisasi: 10000000},
				"gaji_13": {Realisasi: 10000000},
			},
		},
	}
	rows := buildGajiKebutuhan(state, "februari")
	if len(rows) == 0 {
		t.Fatal("expected rows")
	}
	var row *GajiKebutuhanRow
	for i := range rows {
		if rows[i].CategoryID == "gaji_pns" {
			row = &rows[i]
			break
		}
	}
	if row == nil {
		t.Fatal("gaji_pns missing")
	}
	if row.Pagu != 100000000 {
		t.Fatalf("pagu = %v", row.Pagu)
	}
	if row.RealisasiSD != 20000000 {
		t.Fatalf("realisasi sd = %v", row.RealisasiSD)
	}
	if row.SelisihPagu != row.Pagu-row.KebutuhanTahun {
		t.Fatalf("selisih mismatch")
	}
}

func TestNormalizeGajiPeriodThr(t *testing.T) {
	if got := normalizeGajiPeriod("gaji_pns", "THR"); got != "thr" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeGajiPeriod("tpp_pns", "TPP 13"); got != "tpp_13" {
		t.Fatalf("got %q", got)
	}
}
