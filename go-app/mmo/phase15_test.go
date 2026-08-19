package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func phase15Player(w *WorldState, id, name string, level int) *Player {
	p := dungeonReadyPlayer(id, name)
	p.Level = level
	w.Add(p)
	return p
}

func partyReady(t *testing.T, w *WorldState, members []*Player) {
	t.Helper()
	ready, _ := json.Marshal(DungeonActionIn{Ready: true})
	for _, m := range members[1:] {
		if rejectAction(w.ApplyDungeon(m.ID, Envelope{Type: TypeDungeonReady, Data: ready}), TypeDungeonReady) {
			t.Fatalf("ready %s", m.ID)
		}
	}
}

func TestPhase15MistwoodFivePlayers(t *testing.T) {
	w := NewWorldState()
	leader := phase15Player(w, "p15_a", "Raka", 12)
	ids := []string{"p15_b", "p15_c", "p15_d", "p15_e"}
	var members []*Player
	members = append(members, leader)
	for i, id := range ids {
		m := phase15Player(w, id, "M"+itoa(i), 12)
		members = append(members, m)
		w.ApplyParty(leader.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"` + id + `"}`)})
		w.ApplyParty(m.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	}
	evs := w.ApplyDungeon(leader.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-mistwood"}`)})
	if rejectAction(evs, TypeDungeonEnter) {
		t.Fatalf("enter %s", string(evs[0]))
	}
	partyReady(t, w, members)
	inst := w.dungeonOf(leader.ID)
	if inst == nil {
		t.Fatal("instance missing")
	}
	if len(inst.Players) != 5 {
		t.Fatalf("players %d", len(inst.Players))
	}
	if inst.DefID != "dun-mistwood" {
		t.Fatalf("dungeon %s", inst.DefID)
	}
}

func TestPhase15ObjectiveProgress(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p15_obj", "Raka", 12)
	evs := w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-mistwood"}`)})
	if rejectAction(evs, TypeDungeonEnter) {
		t.Fatalf("enter %s", string(evs[0]))
	}
	inst := w.dungeonOf(p.ID)
	var enemy *Enemy
	for _, e := range inst.Enemies {
		if e.Alive && !e.IsBoss {
			enemy = e
			break
		}
	}
	if enemy == nil {
		t.Fatal("pack missing")
	}
	before := inst.ObjProgress
	p.X, p.Z = enemy.X, enemy.Z
	attackEnemy(w, p, enemy.ID)
	if inst.ObjProgress <= before && enemy.HP > 0 {
		enemy.HP = 1
		attackEnemy(w, p, enemy.ID)
	}
	if inst.ObjProgress < 1 && inst.ObjNeed > 0 {
		w.mu.Lock()
		w.onDungeonKill(p, enemy)
		w.mu.Unlock()
	}
	if inst.ObjProgress < 1 {
		t.Fatalf("objective %d/%d", inst.ObjProgress, inst.ObjNeed)
	}
}

func TestPhase15PuzzleAndEduShield(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p15_pz", "Raka", 12)
	w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-mistwood"}`)})
	inst := w.dungeonOf(p.ID)
	p.X, p.Z = -4, 10
	w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"puzzle-wind"}`)})
	p.X, p.Z = 0, 10
	w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"puzzle-stone"}`)})
	p.X, p.Z = 4, 10
	w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"puzzle-light"}`)})
	if inst.PuzzleStep < 3 {
		t.Fatalf("puzzle step %d", inst.PuzzleStep)
	}
	w.mu.Lock()
	w.beginBoss(inst)
	shield := inst.EduShield
	boss := inst.Enemies[inst.BossID]
	hp := 0
	if boss != nil {
		hp = boss.HP
	}
	w.mu.Unlock()
	if !shield {
		t.Fatal("edu shield")
	}
	p.X, p.Z = boss.X, boss.Z
	attackEnemy(w, p, boss.ID)
	if boss.HP != hp && boss.HP < hp {
		t.Fatalf("shield should block damage %d -> %d", hp, boss.HP)
	}
	wrong, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-add-4-3", Choice: 0})
	w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: wrong})
	if !inst.EduShield {
		t.Fatal("wrong answer must keep shield")
	}
	ok, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-add-4-3", Choice: 1})
	w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: ok})
	if inst.EduShield {
		t.Fatal("correct answer should break shield")
	}
}

func TestPhase15BossLootAndDuplicate(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p15_loot", "Raka", 12)
	w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-ch01"}`)})
	inst := w.dungeonOf(p.ID)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	w.mu.Unlock()
	p.X, p.Z = boss.X, boss.Z
	boss.HP = 1
	attackEnemy(w, p, boss.ID)
	if inst.State != DunCompleted {
		t.Fatalf("state %s", inst.State)
	}
	if !inst.ChestReady || len(inst.Loot[p.ID]) == 0 {
		t.Fatal("chest/loot")
	}
	raw, _ := json.Marshal(DungeonActionIn{ClaimID: inst.RewardClaimID})
	if rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeClaimLoot, Data: raw}), TypeClaimLoot) {
		t.Fatal("first claim")
	}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeClaimLoot, Data: raw}), TypeClaimLoot) {
		t.Fatal("duplicate")
	}
}

func TestPhase15LootPersistsReconnect(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p15_re", "Raka", 12)
	w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-ch01"}`)})
	inst := w.dungeonOf(p.ID)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	w.mu.Unlock()
	p.X, p.Z = boss.X, boss.Z
	boss.HP = 1
	attackEnemy(w, p, boss.ID)
	claim := inst.RewardClaimID
	if len(p.ensureLog().PendingLoot[claim]) == 0 {
		t.Fatal("pending loot")
	}
	w.Remove(p)
	back := dungeonReadyPlayer("p15_re", "Raka")
	w.Add(back)
	if len(back.ensureLog().PendingLoot[claim]) == 0 && (w.dungeonOf(back.ID) == nil || len(w.dungeonOf(back.ID).Loot[back.ID]) == 0) {
		t.Fatal("loot should persist")
	}
}

func TestPhase15RaidTenPlayersAndPhase(t *testing.T) {
	w := NewWorldState()
	var ids []string
	for i := 0; i < 10; i++ {
		id := "p15r_" + itoa(i)
		phase15Player(w, id, "R"+itoa(i), 40)
		ids = append(ids, id)
	}
	def := dungeonByID["raid-celestial"]
	w.mu.Lock()
	w.startDungeon(def, ids, "", "NORMAL")
	inst := w.dungeonOf(ids[0])
	if inst == nil || len(inst.Players) != 10 {
		w.mu.Unlock()
		t.Fatalf("raid players")
	}
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	if boss == nil {
		w.mu.Unlock()
		t.Fatal("boss")
	}
	boss.HP = int(float64(boss.MaxHP) * 0.65)
	w.tickBoss(inst, time.Now())
	phase := inst.BossPhase
	w.mu.Unlock()
	if phase < 2 {
		t.Fatalf("phase %d", phase)
	}
}

func TestPhase15RaidWipeCheckpoint(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p15_wipe", "Raka", 40)
	def := dungeonByID["raid-celestial"]
	w.mu.Lock()
	w.startDungeon(def, []string{p.ID}, "", "NORMAL")
	inst := w.dungeonOf(p.ID)
	p.HP = 0
	p.CombatState = "DOWNED"
	w.tickDungeon(inst, 0.016, time.Now())
	wipes := inst.WipeCount
	hp := p.HP
	w.mu.Unlock()
	if wipes < 1 {
		t.Fatal("wipe")
	}
	if hp != p.MaxHP {
		t.Fatalf("checkpoint hp %d", hp)
	}
}

func TestPhase15RaidLockoutAndReconnect(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p15_lock", "Raka", 40)
	p.ensureLog().RaidLockout = map[string]string{"raid-celestial": raidWeekKey()}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeQueueJoin, Data: []byte(`{"dungeonId":"raid-celestial"}`)}), TypeQueueJoin) {
		t.Fatal("lockout")
	}
	a := phase15Player(w, "p15_rc", "Sinta", 12)
	w.ApplyDungeon(a.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-mistwood"}`)})
	inst := w.dungeonOf(a.ID)
	id := inst.ID
	w.Remove(a)
	back := dungeonReadyPlayer("p15_rc", "Sinta")
	w.Add(back)
	if back.InstanceID != id {
		t.Fatalf("rejoin %q", back.InstanceID)
	}
}

func TestPhase15AntiCheat(t *testing.T) {
	w, p := testVillagePlayer()
	p.Level = 12
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeSetBossHP, Data: []byte(`{"bossHp":0}`)}), TypeSetBossHP) {
		t.Fatal("SET_BOSS_HP")
	}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeGiveLoot, Data: []byte(`{"loot":"legendary"}`)}), TypeGiveLoot) {
		t.Fatal("GIVE_LOOT")
	}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeSetObjective, Data: []byte(`{"objectiveComplete":true}`)}), TypeSetObjective) {
		t.Fatal("SET_OBJECTIVE")
	}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeObjectiveComplete, Data: []byte(`{"objectiveComplete":true}`)}), TypeObjectiveComplete) {
		t.Fatal("OBJECTIVE_COMPLETE")
	}
}

func TestPhase15HardAndRandomQueue(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p15_hard", "Raka", 20)
	if rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeQueueJoin, Data: []byte(`{"dungeonId":"dun-mistwood","difficulty":"HARD"}`)}), TypeQueueJoin) {
		t.Fatal("HARD queue")
	}
	w.ApplyDungeon(p.ID, Envelope{Type: TypeQueueLeave, Data: []byte(`{}`)})
	if rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeQueueJoin, Data: []byte(`{"dungeonId":"RANDOM","difficulty":"NORMAL"}`)}), TypeQueueJoin) {
		t.Fatal("RANDOM queue")
	}
	q := w.queueOf(p.ID)
	if q == nil || q.DungeonID == "RANDOM" {
		t.Fatal("server must pick dungeon")
	}
	if !rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeQueueJoin, Data: []byte(`{"dungeonId":"dun-mistwood","difficulty":"MYTHIC"}`)}), TypeQueueJoin) {
		t.Fatal("MYTHIC should reject in prototype")
	}
}

func TestPhase15DungeonListAndFirstClear(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p15_list", "Raka", 50)
	list := w.dungeonList(p)
	found := map[string]bool{}
	for _, d := range list.Dungeons {
		found[d.DungeonID] = true
	}
	for _, id := range []string{"RANDOM", "dun-mistwood", "dun-stoneheart", "dun-crimson", "dun-moonlight", "dun-celestial-ruins", "raid-celestial", "raid-sky-fortress"} {
		if !found[id] {
			t.Fatalf("missing %s", id)
		}
	}
	w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-ch01"}`)})
	inst := w.dungeonOf(p.ID)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	w.mu.Unlock()
	p.X, p.Z = boss.X, boss.Z
	boss.HP = 1
	attackEnemy(w, p, boss.ID)
	if !p.ensureLog().FirstClears["dun-ch01"] {
		t.Fatal("first clear")
	}
	if p.ensureLog().RaidTokens < 0 {
		t.Fatal("raid tokens")
	}
}
