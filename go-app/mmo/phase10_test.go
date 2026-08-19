package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestQueueMatchmaking(t *testing.T) {
	w := NewWorldState()
	a := dungeonReadyPlayer("p_q1", "Raka")
	b := dungeonReadyPlayer("p_q2", "Sinta")
	a.Level, b.Level = 12, 12
	w.Add(a)
	w.Add(b)
	join := []byte(`{"dungeonId":"dun-mistwood","role":"DPS"}`)
	if rejectAction(w.ApplyDungeon(a.ID, Envelope{Type: TypeQueueJoin, Data: join}), TypeQueueJoin) {
		t.Fatal("queue a")
	}
	if rejectAction(w.ApplyDungeon(b.ID, Envelope{Type: TypeQueueJoin, Data: join}), TypeQueueJoin) {
		t.Fatal("queue b")
	}
	w.mu.Lock()
	if len(w.Dungeons.Queue.Entries) != 2 {
		w.mu.Unlock()
		t.Fatalf("queued %d", len(w.Dungeons.Queue.Entries))
	}
	for i := range w.Dungeons.Queue.Entries {
		w.Dungeons.Queue.Entries[i].JoinedAt = time.Now().Add(-10 * time.Second)
	}
	evs := w.tickQueue(time.Now())
	w.mu.Unlock()
	found := false
	for _, ev := range evs {
		var env Envelope
		_ = json.Unmarshal(ev, &env)
		if env.Type == TypeDungeonReadyCheck {
			found = true
		}
	}
	if !found {
		t.Fatal("expected ready check after match")
	}
	ready, _ := json.Marshal(DungeonActionIn{Ready: true})
	w.ApplyDungeon(a.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	w.ApplyDungeon(b.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	if w.dungeonOf(a.ID) == nil || w.dungeonOf(b.ID) == nil {
		t.Fatal("matched instance missing")
	}
	if w.dungeonOf(a.ID) != w.dungeonOf(b.ID) {
		t.Fatal("queue should share instance")
	}
}

func TestDungeonRevive(t *testing.T) {
	w := NewWorldState()
	a := dungeonReadyPlayer("p_r1", "Raka")
	b := dungeonReadyPlayer("p_r2", "Sinta")
	w.Add(a)
	w.Add(b)
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_r2"}`)})
	w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	w.ApplyDungeon(a.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-ch01"}`)})
	ready, _ := json.Marshal(DungeonActionIn{Ready: true})
	w.ApplyDungeon(b.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	inst := w.dungeonOf(a.ID)
	if inst == nil {
		t.Fatal("instance")
	}
	w.mu.Lock()
	b.HP = 0
	b.CombatState = "DOWNED"
	b.X, b.Z = a.X, a.Z
	inst.DownedAt[b.ID] = time.Now()
	w.mu.Unlock()
	evs := w.ApplyDungeon(a.ID, Envelope{Type: TypeDungeonRevive, Data: []byte(`{"targetId":"p_r2"}`)})
	if rejectAction(evs, TypeDungeonRevive) {
		t.Fatalf("revive start %s", string(evs[0]))
	}
	w.mu.Lock()
	w.tickRevive(inst, ReviveChannelTime+0.1)
	hp := b.HP
	cs := b.CombatState
	w.mu.Unlock()
	if cs == "DOWNED" || hp <= 0 {
		t.Fatalf("expected revive hp=%d cs=%s", hp, cs)
	}
}

func TestWipeToCheckpoint(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_w1", "Raka")
	w.Add(p)
	inst := enterChapter1(t, w, p)
	w.mu.Lock()
	p.HP = 0
	p.CombatState = "DOWNED"
	evs := w.tickDungeon(inst, 0.016, time.Now())
	state := inst.State
	hp := p.HP
	w.mu.Unlock()
	if state == DunFailed {
		t.Fatalf("wipe should checkpoint, events=%d", len(evs))
	}
	if hp != p.MaxHP {
		t.Fatalf("checkpoint hp %d", hp)
	}
	if inst.WipeCount < 1 {
		t.Fatal("wipe count")
	}
}

func TestBossInterrupt(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_i1", "Raka")
	w.Add(p)
	inst := enterChapter1(t, w, p)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	inst.Telegraph = &BossTelegraph{SkillID: "charge_burst", X: boss.X, Z: boss.Z, Radius: 3, Until: time.Now().Add(2 * time.Second), Damage: 20, Interruptible: true}
	w.mu.Unlock()
	p.X, p.Z = boss.X, boss.Z
	boss.HP = boss.MaxHP
	attackEnemy(w, p, boss.ID)
	if inst.Telegraph != nil {
		t.Fatal("interruptible telegraph should cancel")
	}
}

func TestBossEnrage(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_e1", "Raka")
	w.Add(p)
	inst := enterChapter1(t, w, p)
	w.mu.Lock()
	w.beginBoss(inst)
	inst.EnrageAt = time.Now().Add(-time.Second)
	w.tickBoss(inst, time.Now())
	enraged := inst.Enraged
	w.mu.Unlock()
	if !enraged {
		t.Fatal("enrage")
	}
}

func TestDuplicateLootClaim(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_l1", "Raka")
	w.Add(p)
	inst := enterChapter1(t, w, p)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	w.mu.Unlock()
	p.X, p.Z = boss.X, boss.Z
	boss.HP = 1
	attackEnemy(w, p, boss.ID)
	raw, _ := json.Marshal(DungeonActionIn{ClaimID: inst.RewardClaimID})
	if rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeClaimLoot, Data: raw}), TypeClaimLoot) {
		t.Fatal("first claim")
	}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeClaimLoot, Data: raw}), TypeClaimLoot) {
		t.Fatal("duplicate claim must reject")
	}
}

func TestJoinWithoutMembershipRejected(t *testing.T) {
	w := NewWorldState()
	a := dungeonReadyPlayer("p_j1", "Raka")
	b := dungeonReadyPlayer("p_j2", "Sinta")
	c := dungeonReadyPlayer("p_j3", "Budi")
	w.Add(a)
	w.Add(b)
	w.Add(c)
	inst := enterChapter1(t, w, a)
	raw, _ := json.Marshal(DungeonActionIn{InstanceID: inst.ID})
	if !rejectAction(w.ApplyDungeon(b.ID, Envelope{Type: TypeDungeonJoin, Data: raw}), TypeDungeonJoin) {
		t.Fatal("join without membership")
	}
	w.mu.Lock()
	w.beginBoss(inst)
	w.mu.Unlock()
	if !rejectAction(w.ApplyDungeon(c.ID, Envelope{Type: TypeDungeonFill, Data: raw}), TypeDungeonFill) {
		t.Fatal("fill after boss lock")
	}
}

func TestRaidWeeklyLockout(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_lock", "Raka")
	p.Level = 40
	w.Add(p)
	p.ensureLog().RaidLockout = map[string]string{"raid-celestial": raidWeekKey()}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeQueueJoin, Data: []byte(`{"dungeonId":"raid-celestial"}`)}), TypeQueueJoin) {
		t.Fatal("raid lockout queue")
	}
	a := dungeonReadyPlayer("p_ra", "A")
	b := dungeonReadyPlayer("p_rb", "B")
	c := dungeonReadyPlayer("p_rc", "C")
	d := dungeonReadyPlayer("p_rd", "D")
	for _, x := range []*Player{a, b, c, d} {
		x.Level = 40
		w.Add(x)
	}
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_rb"}`)})
	w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_rc"}`)})
	w.ApplyParty(c.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_rd"}`)})
	w.ApplyParty(d.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	evs := w.ApplyDungeon(a.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"raid-celestial"}`)})
	if rejectAction(evs, TypeDungeonEnter) {
		t.Fatalf("raid enter %s", string(evs[0]))
	}
	ready, _ := json.Marshal(DungeonActionIn{Ready: true})
	w.ApplyDungeon(b.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	w.ApplyDungeon(c.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	w.ApplyDungeon(d.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	inst := w.dungeonOf(a.ID)
	if inst == nil {
		t.Fatal("raid instance")
	}
	for i := 0; i < 3; i++ {
		w.mu.Lock()
		if inst.State != DunBoss {
			w.beginBoss(inst)
		}
		boss := inst.Enemies[inst.BossID]
		w.mu.Unlock()
		if boss == nil {
			t.Fatalf("encounter %d boss missing", i)
		}
		a.X, a.Z = boss.X, boss.Z
		boss.HP = 1
		attackEnemy(w, a, boss.ID)
	}
	if inst.State != DunCompleted {
		t.Fatalf("raid state %s", inst.State)
	}
	if a.ensureLog().RaidLockout["raid-celestial"] != raidWeekKey() {
		t.Fatal("lockout not set")
	}
	owned := false
	for _, title := range a.ensureLog().Titles {
		if title == "celestial-explorer" {
			owned = true
		}
	}
	if !owned {
		t.Fatal("raid title")
	}
}

func TestLootStealRejected(t *testing.T) {
	w := NewWorldState()
	a := dungeonReadyPlayer("p_s1", "Raka")
	b := dungeonReadyPlayer("p_s2", "Sinta")
	w.Add(a)
	w.Add(b)
	inst := enterChapter1(t, w, a)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	w.mu.Unlock()
	a.X, a.Z = boss.X, boss.Z
	boss.HP = 1
	attackEnemy(w, a, boss.ID)
	raw, _ := json.Marshal(DungeonActionIn{ClaimID: inst.RewardClaimID})
	if !rejectAction(w.ApplyDungeon(b.ID, Envelope{Type: TypeClaimLoot, Data: raw}), TypeClaimLoot) {
		t.Fatal("outsider loot steal")
	}
}

func TestTeleportInDungeonRejected(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_tp", "Raka")
	w.Add(p)
	enterChapter1(t, w, p)
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeTeleport, Data: []byte(`{"landmarkId":"celestial-gate"}`)}), TypeTeleport) {
		t.Fatal("teleport")
	}
}

func TestRaidChatRequiresInstance(t *testing.T) {
	w := NewWorldState()
	p := dungeonReadyPlayer("p_chat", "Raka")
	w.Add(p)
	if !rejectAction(w.ApplyChat(p.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"RAID","text":"go"}`)}), TypeChat) {
		t.Fatal("raid chat outside")
	}
}
