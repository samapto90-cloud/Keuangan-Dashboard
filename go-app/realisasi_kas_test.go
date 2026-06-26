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

func TestMergeKasRakNoLockedUsesNew(t *testing.T) {
	old := []RakBelanjaRow{{KodeRekening: "5.1.01.", Bulan: map[string]float64{"januari": 1}}}
	newRows := []RakBelanjaRow{{KodeRekening: "5.1.01.", Bulan: map[string]float64{"januari": 9}}}
	got := mergeKasRakPreservingLockedMonths(old, newRows, nil)
	if got[0].Bulan["januari"] != 9 {
		t.Fatalf("tanpa kunci harus pakai RAK baru, got %v", got[0].Bulan["januari"])
	}
}
