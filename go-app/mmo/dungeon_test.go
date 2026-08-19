package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func dungeonReadyPlayer(id, name string) *Player {
	p := &Player{ID: id, Name: name, send: make(chan []byte, 32)}
	p.initCombat()
	p.Level = 5
	return p
}

func enterChapter1(t *testing.T, w *WorldState, p *Player) *DungeonInstance {
	t.Helper()
	evs := w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-ch01"}`)})
	if rejectAction(evs, TypeDungeonEnter) {
		t.Fatalf("enter rejected: %s", string(evs[0]))
	}
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		t.Fatal("instance missing")
	}
	return inst
}

func attackEnemy(w *WorldState, p *Player, enemyID string) [][]byte {
	p.AttackCDUntil = time.Time{}
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: enemyID, Timestamp: time.Now().UnixMilli()})
	return w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
}

func TestDungeonInstancesIsolated(t *testing.T) {
	w := NewWorldState()
	a := dungeonReadyPlayer("p_a", "Raka")
	b := dungeonReadyPlayer("p_b", "Sinta")
	w.Add(a)
	w.Add(b)
	ia := enterChapter1(t, w, a)
	ib := enterChapter1(t, w, b)
	if ia.ID == ib.ID {
		t.Fatal("party A and B must not share instance")
	}
	snapA := w.SnapshotFor(a.ID)
	snapB := w.SnapshotFor(b.ID)
	if snapA.InstanceID != ia.ID || snapB.InstanceID != ib.ID {
		t.Fatal("snapshot instance mismatch")
	}
	for _, e := range snapA.Enemies {
		if e.ID == "" {
			continue
		}
		for _, other := range snapB.Enemies {
			if e.ID == other.ID {
				t.Fatal("enemy leaked across instances")
			}
		}
	}
	var ea *Enemy
	for _, e := range ia.Enemies {
		if e.Alive {
			ea = e
			break
		}
	}
	if ea == nil {
		t.Fatal("no dungeon enemy")
	}
	hpB := 0
	for _, e := range ib.Enemies {
		if e.Alive {
			hpB = e.HP
			break
		}
	}
	a.X, a.Z = ea.X, ea.Z
	attackEnemy(w, a, ea.ID)
	if ea.HP >= ea.MaxHP {
		t.Fatal("A should damage own instance")
	}
	for _, e := range ib.Enemies {
		if e.Alive && e.HP != hpB && hpB > 0 {
			t.Fatal("B enemy HP changed by A")
		}
	}
}

func TestDungeonWaveAdvances(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_a", "Raka")
	w.Add(p)
	inst := enterChapter1(t, w, p)
	if inst.WaveIndex != 1 {
		t.Fatalf("wave %d", inst.WaveIndex)
	}
	ids := make([]string, 0, len(inst.Enemies))
	for id := range inst.Enemies {
		ids = append(ids, id)
	}
	for _, id := range ids {
		e := inst.Enemies[id]
		if e == nil || !e.Alive {
			continue
		}
		p.X, p.Z = e.X, e.Z
		e.HP = 1
		attackEnemy(w, p, e.ID)
	}
	if inst.WaveIndex != 2 {
		t.Fatalf("expected wave 2, got %d", inst.WaveIndex)
	}
}

func TestBossPhasesAtThresholds(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_a", "Raka")
	w.Add(p)
	inst := enterChapter1(t, w, p)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	if boss == nil {
		w.mu.Unlock()
		t.Fatal("boss missing")
	}
	boss.HP = int(float64(boss.MaxHP) * 0.55)
	evs := w.tickBoss(inst, time.Now())
	if inst.BossPhase != 2 {
		w.mu.Unlock()
		t.Fatalf("phase 2 want, got %d events=%d", inst.BossPhase, len(evs))
	}
	boss.HP = int(float64(boss.MaxHP) * 0.2)
	w.tickBoss(inst, time.Now())
	phase := inst.BossPhase
	w.mu.Unlock()
	if phase != 3 {
		t.Fatalf("phase 3 want, got %d", phase)
	}
}

func TestBossDeathCompletesChapter(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_a", "Raka")
	w.Add(p)
	q := p.quest("mq005")
	q.State = QuestActive
	inst := enterChapter1(t, w, p)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	w.mu.Unlock()
	if boss == nil {
		t.Fatal("boss missing")
	}
	p.X, p.Z = boss.X, boss.Z
	boss.HP = 1
	attackEnemy(w, p, boss.ID)
	if inst.State != DunCompleted {
		t.Fatalf("state %s", inst.State)
	}
	if !p.ensureLog().Flags["chapter1_complete"] || !p.ensureLog().Flags["chapter2_unlocked"] {
		t.Fatal("chapter flags missing")
	}
	if q.State != QuestCompleted && q.Progress[0] < 1 {
		t.Fatalf("mq005 progress %+v %s", q.Progress, q.State)
	}
	if len(inst.Loot[p.ID]) == 0 {
		t.Fatal("personal loot missing")
	}
}

func TestDungeonCheatsRejected(t *testing.T) {
	w, p := testVillagePlayer()
	p.Level = 5
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeUnlockChapter, Data: []byte(`{"dungeonId":"ch02"}`)}), TypeUnlockChapter) {
		t.Fatal("UNLOCK_CHAPTER")
	}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeSetBossHP, Data: []byte(`{"bossHp":1}`)}), TypeSetBossHP) {
		t.Fatal("SET_BOSS_HP")
	}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeSetChapter, Data: []byte(`{}`)}), TypeSetChapter) {
		t.Fatal("SET_CHAPTER")
	}
	inst := enterChapter1(t, w, p)
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeClaimLoot, Data: []byte(`{"claimId":"fake"}`)}), TypeClaimLoot) {
		t.Fatal("fake loot")
	}
	_ = inst
}

func TestDungeonReconnectWithinTimeout(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_a", "Raka")
	w.Add(p)
	inst := enterChapter1(t, w, p)
	id := inst.ID
	w.Remove(p)
	if inst.State == DunFailed {
		t.Fatal("offline should not wipe immediately")
	}
	back := dungeonReadyPlayer("p_a", "Raka")
	w.Add(back)
	if back.InstanceID != id {
		t.Fatalf("rejoin instance %q want %s", back.InstanceID, id)
	}
}

func TestNonLeaderAbandonRejected(t *testing.T) {
	w := NewWorldState()
	a := dungeonReadyPlayer("p_a", "Raka")
	b := dungeonReadyPlayer("p_b", "Sinta")
	w.Add(a)
	w.Add(b)
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	evs := w.ApplyDungeon(a.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-ch01"}`)})
	if rejectAction(evs, TypeDungeonEnter) {
		t.Fatal("leader enter")
	}
	ready, _ := json.Marshal(DungeonActionIn{Ready: true})
	w.ApplyDungeon(b.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	if w.dungeonOf(a.ID) == nil {
		t.Fatal("party dungeon missing")
	}
	if !rejectAction(w.ApplyDungeon(b.ID, Envelope{Type: TypeDungeonAbandon, Data: []byte(`{}`)}), TypeDungeonAbandon) {
		t.Fatal("member abandon must reject")
	}
}

func TestMemberCannotEnterForParty(t *testing.T) {
	w := NewWorldState()
	a := dungeonReadyPlayer("p_a", "Raka")
	b := dungeonReadyPlayer("p_b", "Sinta")
	w.Add(a)
	w.Add(b)
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	if !rejectAction(w.ApplyDungeon(b.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-ch01"}`)}), TypeDungeonEnter) {
		t.Fatal("member enter must reject")
	}
}
