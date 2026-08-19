package mmo

import (
	"encoding/json"
	"testing"
)

func TestJourneyStartsInVillage(t *testing.T) {
	w, p := testVillagePlayer()
	j := p.journeyView()
	if j.Objective == "" {
		t.Fatal("journey objective wajib ada")
	}
	if j.NextRegion != "forest" {
		t.Fatalf("tujuan awal harus hutan, got %s", j.NextRegion)
	}
	out := p.progressOut(w.Time.Phase)
	if out.Journey == nil || out.Journey.Objective == "" {
		t.Fatal("progress harus membawa journey")
	}
}

func TestReachVillageGateCompletesMq003(t *testing.T) {
	w, p := testVillagePlayer()
	w.acceptQuest(p, "mq001")
	w.claimTalk(p, "mq001")
	w.claimQuest(p, "mq001")
	w.acceptQuest(p, "mq002")
	p.credit("KILL", "training_dummy", 3)
	w.claimQuest(p, "mq002")
	w.acceptQuest(p, "mq003")
	if p.quest("mq003").State != QuestActive {
		t.Fatalf("mq003 %s", p.quest("mq003").State)
	}
	p.X, p.Z = 0, 19.2
	w.Simulate(0.1)
	if p.quest("mq003").State != QuestCompleted {
		t.Fatalf("REACH gerbang harus menyelesaikan mq003, got %s", p.quest("mq003").State)
	}
}

func TestVillageNpcDensity(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 0, 6
	snap := w.SnapshotFor(p.ID)
	n := 0
	boards := 0
	for _, npc := range snap.NPCs {
		n++
		if npc.Type == "QUEST_BOARD" {
			boards++
		}
	}
	if n > 16 {
		t.Fatalf("NPC desa terlalu ramai: %d", n)
	}
	if boards > 1 {
		t.Fatalf("papan misi tidak boleh memenuhi alun-alun: %d", boards)
	}
}

func TestRegionIntroCinematicOnForest(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().ForestUnlocked = true
	p.X, p.Z = 0, 30
	evs := w.Simulate(0.05)
	found := false
	for _, raw := range evs {
		var env Envelope
		_ = json.Unmarshal(raw, &env)
		if env.Type == TypeCinematicStart {
			found = true
		}
	}
	if !found {
		t.Fatal("masuk hutan pertama kali harus cinematic pendek")
	}
}
