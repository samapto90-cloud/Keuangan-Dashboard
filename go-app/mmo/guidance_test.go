package mmo

import "testing"

func TestCardinalNorthIsPlusZ(t *testing.T) {
	if got := CardinalFromDelta(0, 10); got != "N" {
		t.Fatalf("utara harus +Z, got %s", got)
	}
	if got := CardinalFromDelta(10, 0); got != "E" {
		t.Fatalf("timur harus +X, got %s", got)
	}
	if got := CardinalFromDelta(0, -10); got != "S" {
		t.Fatalf("selatan harus -Z, got %s", got)
	}
	if got := CardinalFromDelta(-10, 0); got != "W" {
		t.Fatalf("barat harus -X, got %s", got)
	}
}

func TestJourneyNavAndHintSameCardinal(t *testing.T) {
	_, p := testVillagePlayer()
	p.X, p.Z = 0, 0
	j := p.journeyView()
	want := CardinalFromDelta(j.NavX-p.X, j.NavZ-p.Z)
	if j.Cardinal != want {
		t.Fatalf("journey cardinal %s != nav %s", j.Cardinal, want)
	}
	jv, id, _, card := p.journeyHint()
	if card != want {
		t.Fatalf("NPC hint cardinal %s != nav %s", card, want)
	}
	if jv == "" || id == "" {
		t.Fatal("hint Jawa dan Indonesia wajib ada")
	}
}

func TestWrongWayThreshold(t *testing.T) {
	if WrongWayReady(99, 20) {
		t.Fatal("belum 100m tidak boleh guidance")
	}
	if WrongWayReady(120, 10) {
		t.Fatal("belum 15 detik tidak boleh guidance")
	}
	if !WrongWayReady(120, 15) {
		t.Fatal("100m + 15s harus guidance lembut")
	}
}

func TestGuideNpcUsesJourneyCardinal(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 0, 6
	j := p.journeyView()
	want := CardinalFromDelta(j.NavX-p.X, j.NavZ-p.Z)
	npc := NPCDef{ID: "guide_faris", Name: "Guide Faris", Type: "GUIDE", X: 4.2, Z: 8.5}
	evs := w.talkNPC(p, npc)
	if len(evs) == 0 {
		t.Fatal("talk GUIDE harus ada hasil")
	}
	_, id, _, card := p.journeyHint()
	if card != want {
		t.Fatalf("GUIDE %s vs journey %s", card, want)
	}
	if CardinalID(card) == "" || id == "" {
		t.Fatal("petunjuk harus punya terjemahan")
	}
}

func TestNoAutoTeleportOnWrongWay(t *testing.T) {
	_, p := testVillagePlayer()
	p.X, p.Z = 8, -4
	x, z := p.X, p.Z
	_ = p.journeyView()
	if WrongWayReady(120, 16) && (p.X != x || p.Z != z) {
		t.Fatal("guidance tidak boleh memindahkan pemain")
	}
	if p.X != x || p.Z != z {
		t.Fatal("posisi tidak boleh berubah hanya karena journeyView")
	}
}
