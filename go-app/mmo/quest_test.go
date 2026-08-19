package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRejectQuestCompleteCheat(t *testing.T) {
	w, p := testVillagePlayer()
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeQuestComplete, Data: []byte(`{"questId":"mq001"}`)})
	if !rejectAction(evs, TypeQuestComplete) {
		t.Fatal("QUEST_COMPLETE harus ditolak server")
	}
	if p.quest("mq001").State != QuestAvailable {
		t.Fatal("quest tidak boleh selesai dari cheat")
	}
}

func TestRejectCollectItemQuantity(t *testing.T) {
	w, p := testVillagePlayer()
	raw, _ := json.Marshal(CollectItemIn{ItemID: "crystal", Quantity: 100})
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeCollectItem, Data: raw})
	if !rejectAction(evs, TypeCollectItem) {
		t.Fatal("COLLECT_ITEM x100 harus ditolak")
	}
}

func TestRejectForestUnlock(t *testing.T) {
	w, p := testVillagePlayer()
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeForestUnlock, Data: []byte(`{}`)})
	if !rejectAction(evs, TypeForestUnlock) {
		t.Fatal("FOREST_UNLOCK harus ditolak")
	}
	if p.ensureLog().ForestUnlocked {
		t.Fatal("hutan tidak boleh terbuka dari cheat")
	}
}

func TestRejectEducationCorrectWithoutAnswer(t *testing.T) {
	w, p := testVillagePlayer()
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationCorrect, Data: []byte(`{}`)})
	if !rejectAction(evs, TypeEducationCorrect) {
		t.Fatal("EDUCATION_CORRECT tanpa soal harus ditolak")
	}
}

func TestQuestProgressPerPlayer(t *testing.T) {
	w, a := testVillagePlayer()
	b := &Player{ID: "p_b", Name: "Sinta", send: make(chan []byte, 8)}
	b.initCombat()
	w.Add(b)
	w.acceptQuest(a, "mq001")
	if a.quest("mq001").State == QuestAvailable {
		t.Fatal("A harus menerima quest")
	}
	if b.quest("mq001").State != QuestAvailable {
		t.Fatalf("B tidak boleh otomatis menerima quest: %s", b.quest("mq001").State)
	}
}

func TestTalkQuestDoesNotShare(t *testing.T) {
	w, a := testVillagePlayer()
	b := &Player{ID: "p_b", Name: "Sinta", send: make(chan []byte, 8)}
	b.initCombat()
	w.Add(b)
	w.claimTalk(a, "mq001")
	if a.quest("mq001").State != QuestClaimed {
		t.Fatalf("A claim %s", a.quest("mq001").State)
	}
	if b.quest("mq001").State != QuestAvailable {
		t.Fatalf("B harus tetap available, got %s", b.quest("mq001").State)
	}
}

func TestKillCreditAttackerOnly(t *testing.T) {
	w, a := testVillagePlayer()
	b := &Player{ID: "p_b", Name: "Sinta", send: make(chan []byte, 8)}
	b.initCombat()
	w.Add(b)
	w.acceptQuest(a, "mq001")
	w.claimTalk(a, "mq001")
	w.acceptQuest(a, "mq002")
	w.acceptQuest(b, "mq001")
	dummy := firstKind(w, "training_dummy")
	if dummy == nil {
		t.Fatal("dummy")
	}
	a.X, a.Z = dummy.X, dummy.Z
	dummy.HP = 1
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: dummy.ID, Timestamp: time.Now().UnixMilli()})
	w.ApplyCombat(a.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if a.quest("mq002").Progress[0] != 1 {
		t.Fatalf("A kill credit %v", a.quest("mq002").Progress)
	}
	if b.quest("mq002").State == QuestActive && b.quest("mq002").Progress[0] != 0 {
		t.Fatal("B tidak boleh mendapat kill credit")
	}
}

func TestGateLockedUntilQuest(t *testing.T) {
	w, p := testVillagePlayer()
	p.Z = 30
	w.Simulate(0.05)
	if p.Z > GateMaxZ+0.05 {
		t.Fatalf("gerbang terkunci harus menahan player: z=%v", p.Z)
	}
	p.ensureLog().ForestUnlocked = true
	p.Z = 30
	w.Simulate(0.05)
	if p.Z < 29 {
		t.Fatal("setelah unlock player boleh masuk hutan")
	}
}

func TestEducationAnswerRequiresSession(t *testing.T) {
	w, p := testVillagePlayer()
	raw, _ := json.Marshal(EducationAnswerIn{QuestionID: "q1", Choice: 1})
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: raw})
	if !rejectAction(evs, TypeEducationAnswer) {
		t.Fatal("jawaban tanpa sesi harus ditolak")
	}
}

func testVillagePlayer() (*WorldState, *Player) {
	w := NewWorldState()
	p := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 16)}
	p.initCombat()
	w.Add(p)
	p.X, p.Z = 0, 3.2
	return w, p
}

func rejectAction(evs [][]byte, action string) bool {
	for _, ev := range evs {
		var env Envelope
		if json.Unmarshal(ev, &env) != nil || env.Type != TypeActionReject {
			continue
		}
		var r RejectOut
		_ = json.Unmarshal(env.Data, &r)
		if r.Action == action {
			return true
		}
	}
	return false
}

func firstKind(w *WorldState, kind string) *Enemy {
	for _, e := range w.enemies {
		if e.Def.ID == kind && e.Alive {
			return e
		}
	}
	return nil
}

func (w *WorldState) claimTalk(p *Player, id string) {
	w.acceptQuest(p, id)
	w.claimQuest(p, id)
}
