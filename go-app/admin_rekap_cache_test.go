package main

import "testing"

func TestAdminRekapCache(t *testing.T) {
	invalidateAdminRekapCache()
	portals := []string{"sekretariat", "paud"}
	calls := 0
	build := func() []adminRekapRow {
		calls++
		return []adminRekapRow{{PortalID: "sekretariat", Kegiatan: "Test", Anggaran: 100}}
	}
	_ = cachedAdminRekapRows(portals, "kegiatan", "", "", build)
	_ = cachedAdminRekapRows(portals, "kegiatan", "", "", build)
	if calls != 1 {
		t.Fatalf("expected 1 build, got %d", calls)
	}
	invalidateAdminRekapCache()
	_ = cachedAdminRekapRows(portals, "kegiatan", "", "", build)
	if calls != 2 {
		t.Fatalf("expected 2 builds after invalidate, got %d", calls)
	}
}

func TestSessionPublicID(t *testing.T) {
	id1 := sessionPublicID("abc123")
	id2 := sessionPublicID("abc123")
	id3 := sessionPublicID("other")
	if id1 != id2 || id1 == id3 || len(id1) != 24 {
		t.Fatalf("unexpected session ids: %q %q %q", id1, id2, id3)
	}
}

func TestAdminRekapRowsToPPTKStatsSort(t *testing.T) {
	stats := adminRekapRowsToPPTKStats([]adminRekapRow{
		{PPTK: "A", Realisasi: 10},
		{PPTK: "B", Realisasi: 50},
	})
	if len(stats) != 2 || stats[0].PPTK != "B" {
		t.Fatalf("expected B first, got %+v", stats)
	}
}
