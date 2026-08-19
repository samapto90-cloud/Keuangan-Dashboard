package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPhase23LandminesKept(t *testing.T) {
	if questionByID["q-add-4-3"].ID == "" || questionByID["q-add-4-3"].Correct != 1 {
		t.Fatal("keep q-add-4-3")
	}
	if questionByID[WhispersEduID].ID == "" || questionByID[WhispersEduID].Correct != 1 {
		t.Fatal("5+3=8 index B")
	}
	if dungeonByID["dun-mistwood"].EducationBoss != "q-add-4-3" {
		t.Fatal("mistwood edu")
	}
	if dungeonByID["raid-celestial"].ID == "" {
		t.Fatal("keep raid-celestial")
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatal("keep Dawn City")
	}
	if mountByID["wind-runner"].Name != "Wind Runner" {
		t.Fatal("keep wind-runner")
	}
	if shopCatalog.ID != "dawn-merchant" {
		t.Fatal("shops.json shape")
	}
	if eventByID["mount-festival"].ID == "" || eventByID["festival-karya"].ID == "" || eventByID["open-house"].Name != "OPEN HOUSE" {
		t.Fatal("keep events")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
	if regionByID["ruins"].Title != "Ancient Ruins" {
		t.Fatal("ancient ruins overlay")
	}
	if dungeonByID[WhispersDungeonID].Name != "CAVE OF WHISPERS" {
		t.Fatal("cave of whispers")
	}
	if dungeonByID[GerbangRaidID].MaxPlayers != 8 || dungeonByID[GerbangRaidID].MinPlayers != 8 {
		t.Fatal("raid 8")
	}
	if bossByID[WhispersBossID].Name != "PENJAGA GUA" || bossByID[RajaSilumanID].Name != "RAJA SILUMAN" {
		t.Fatal("original bosses")
	}
	if eventByID[SeranganSilumanID].Name != "SERANGAN SILUMAN" {
		t.Fatal("serangan siluman")
	}
	if npcByID["pak_jaga"].ID == "" || dialogueCatalog["pak_jaga"].Text != "Ati-ati, Le! Ana siluman teka saka alas!" {
		t.Fatal("pak jaga jawa")
	}
}

func TestPhase23CheatsRejected(t *testing.T) {
	w, p := testVillagePlayer()
	p.Level = 12
	cheats := []string{TypeSetBossDead, TypeSpawnBoss, TypeCompleteDungeon, TypeSkipMechanic, TypeDamageBoss, TypeGiveLoot, TypeSetBossHP, TypeTeleport}
	for _, c := range cheats {
		evs := w.ApplyDungeon(p.ID, Envelope{Type: c, Data: []byte(`{"bossHp":0}`)})
		if c == TypeTeleport {
			evs = w.ApplyExplore(p.ID, Envelope{Type: c, Data: []byte(`{"landmarkId":"celestial-gate"}`)})
		}
		if !rejectAction(evs, c) {
			t.Fatalf("%s", c)
		}
	}
}

func TestPhase23WorldEventAnnounce(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().ForestUnlocked = true
	p.X, p.Z = 0, 36
	w.mu.Lock()
	evs := w.startWorldEvent(SeranganSilumanID)
	w.mu.Unlock()
	if len(evs) == 0 {
		t.Fatal("event")
	}
	live := w.Events.Active
	if live == nil || live.Def.Name != "SERANGAN SILUMAN" {
		t.Fatalf("name %+v", live)
	}
	v := live.view()
	if v.Announce == "" || v.Objective == "" || v.Kind != "hunt" {
		t.Fatalf("view %+v", v)
	}
	join := w.ApplyExplore(p.ID, Envelope{Type: TypeJoinWorldEvent})
	if rejectAction(join, TypeJoinWorldEvent) {
		t.Fatal("join tanpa party")
	}
	j := p.worldJournal(w)
	found := false
	for _, m := range j.Markers {
		if m.Kind == "World Event" && m.ID == SeranganSilumanID {
			found = true
		}
	}
	if !found {
		t.Fatal("event marker")
	}
}

func TestPhase23WorldBossRatu(t *testing.T) {
	w := NewWorldState()
	w.mu.Lock()
	evs := w.spawnNamedWorldBoss(RatuBayanganID)
	w.mu.Unlock()
	if len(evs) == 0 || w.Boss == nil || w.Boss.Def.Name != "RATU BAYANGAN" {
		t.Fatalf("boss %+v", w.Boss)
	}
	if w.Boss.Def.DurationSec != 1200 {
		t.Fatalf("timer %v", w.Boss.Def.DurationSec)
	}
	ow := w.openWorldView(addTestPlayer(w, "p23_wb", "Raka"))
	if ow.NextWorldBoss == nil || ow.WorldBoss == nil {
		t.Fatal("preview")
	}
}

func TestPhase23WhispersInstance(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p23_wh", "Raka", 8)
	evs := w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-whispers"}`)})
	if rejectAction(evs, TypeDungeonEnter) {
		t.Fatalf("enter %s", string(evs[0]))
	}
	inst := w.dungeonOf(p.ID)
	if inst == nil || inst.DefID != WhispersDungeonID {
		t.Fatal("instance")
	}
	view := w.dungeonView(inst, p.ID)
	if view.Room == "" || view.TimeLeft < 1700 {
		t.Fatalf("room/timer %+v", view)
	}
	if dungeonByID[inst.DefID].MaxPlayers != 5 || dungeonByID[inst.DefID].MinPlayers != 1 {
		t.Fatal("party 1-5")
	}
}

func TestPhase23WhispersEducation(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p23_edu", "Raka", 8)
	w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-whispers"}`)})
	inst := w.dungeonOf(p.ID)
	w.mu.Lock()
	w.beginBoss(inst)
	shield := inst.EduShield
	boss := inst.Enemies[inst.BossID]
	hp := 0
	if boss != nil {
		hp = boss.HP
	}
	w.mu.Unlock()
	if !shield || boss == nil {
		t.Fatal("edu shield")
	}
	p.X, p.Z = boss.X, boss.Z
	attackEnemy(w, p, boss.ID)
	if boss.HP < hp {
		t.Fatalf("shield block %d -> %d", hp, boss.HP)
	}
	coin := p.ensureLog().Coin
	wrong, _ := json.Marshal(EducationAnswerIn{QuestionID: WhispersEduID, Choice: 0})
	w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: wrong})
	if !inst.EduShield {
		t.Fatal("wrong keeps shield")
	}
	if p.ensureLog().Coin < coin {
		t.Fatal("no gold punishment")
	}
	ok, _ := json.Marshal(EducationAnswerIn{QuestionID: WhispersEduID, Choice: 1})
	w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: ok})
	if inst.EduShield {
		t.Fatal("correct reduces shield")
	}
}

func TestPhase23BossPhaseAndEnrage(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p23_ph", "Raka", 8)
	w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-whispers"}`)})
	inst := w.dungeonOf(p.ID)
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	if boss == nil {
		w.mu.Unlock()
		t.Fatal("boss")
	}
	boss.HP = boss.MaxHP / 2
	w.tickBoss(inst, time.Now())
	if inst.BossPhase != 2 {
		w.mu.Unlock()
		t.Fatalf("phase %d", inst.BossPhase)
	}
	inst.EnrageAt = time.Now().Add(-time.Second)
	atk := boss.Def.Attack
	w.tickBoss(inst, time.Now())
	w.mu.Unlock()
	if !inst.Enraged {
		t.Fatal("enrage")
	}
	if boss.Def.Attack <= atk {
		t.Fatal("enrage more aggressive")
	}
}

func TestPhase23RaidQueueEight(t *testing.T) {
	w := NewWorldState()
	var members []*Player
	for i := 0; i < 8; i++ {
		p := phase15Player(w, "p23r"+itoa(i), "R"+itoa(i), 12)
		members = append(members, p)
		if rejectAction(w.ApplyDungeon(p.ID, Envelope{Type: TypeQueueJoin, Data: []byte(`{"dungeonId":"raid-gerbang-33","role":"FLEX"}`)}), TypeQueueJoin) {
			t.Fatalf("queue %d", i)
		}
	}
	ready, _ := json.Marshal(DungeonActionIn{Ready: true})
	for _, m := range members {
		w.ApplyDungeon(m.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	}
	w.mu.Lock()
	if w.Dungeons != nil && w.Dungeons.Queue != nil && len(w.Dungeons.Queue.Entries) > 0 {
		for i := range w.Dungeons.Queue.Entries {
			w.Dungeons.Queue.Entries[i].JoinedAt = time.Now().Add(-10 * time.Second)
		}
		_ = w.tickQueue(time.Now())
	}
	w.mu.Unlock()
	for _, m := range members {
		w.ApplyDungeon(m.ID, Envelope{Type: TypeDungeonReady, Data: ready})
	}
	inst := w.dungeonOf(members[0].ID)
	if inst == nil || inst.Kind != "RAID" || inst.DefID != GerbangRaidID {
		t.Fatalf("raid inst %+v", inst)
	}
	if len(inst.Players) != 8 {
		t.Fatalf("players %d", len(inst.Players))
	}
}

func TestPhase23InstanceSecurityAndDupLoot(t *testing.T) {
	w := NewWorldState()
	a := phase15Player(w, "p23_a", "Raka", 8)
	b := phase15Player(w, "p23_b", "Lila", 8)
	w.ApplyDungeon(a.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-whispers"}`)})
	inst := w.dungeonOf(a.ID)
	raw, _ := json.Marshal(DungeonActionIn{InstanceID: inst.ID})
	if !rejectAction(w.ApplyDungeon(b.ID, Envelope{Type: TypeDungeonJoin, Data: raw}), TypeDungeonJoin) {
		t.Fatal("random join")
	}
	w.mu.Lock()
	w.beginBoss(inst)
	boss := inst.Enemies[inst.BossID]
	w.mu.Unlock()
	a.X, a.Z = boss.X, boss.Z
	w.ApplyWorld(a.ID, Envelope{Type: TypeEducationAnswer, Data: []byte(`{"questionId":"q-add-5-3","choice":1}`)})
	boss.HP = 1
	attackEnemy(w, a, boss.ID)
	if inst.State != DunCompleted {
		t.Fatalf("state %s", inst.State)
	}
	claim, _ := json.Marshal(DungeonActionIn{ClaimID: inst.RewardClaimID})
	if rejectAction(w.ApplyDungeon(a.ID, Envelope{Type: TypeClaimLoot, Data: claim}), TypeClaimLoot) {
		t.Fatal("first claim")
	}
	if !rejectAction(w.ApplyDungeon(a.ID, Envelope{Type: TypeClaimLoot, Data: claim}), TypeClaimLoot) {
		t.Fatal("dup claim")
	}
	if !rejectAction(w.ApplyDungeon(b.ID, Envelope{Type: TypeClaimLoot, Data: claim}), TypeClaimLoot) {
		t.Fatal("loot steal")
	}
}

func TestPhase23OpenWorldAndFastTravelBoss(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().Landmarks["old-windmill"] = true
	p.ensureLog().FastTravel["old-windmill"] = true
	ow := w.ApplyExplore(p.ID, Envelope{Type: TypeGetOpenWorld})
	if len(ow) == 0 {
		t.Fatal("open world")
	}
	w.mu.Lock()
	w.spawnNamedWorldBoss(RatuBayanganID)
	p.X, p.Z = 2, 36
	w.mu.Unlock()
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeFastTravel, Data: []byte(`{"landmarkId":"old-windmill"}`)}), TypeFastTravel) {
		t.Fatal("fast travel during boss")
	}
}

func TestPhase23ReconnectAndCleanup(t *testing.T) {
	w := NewWorldState()
	p := phase15Player(w, "p23_rc", "Raka", 8)
	w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-whispers"}`)})
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		t.Fatal("inst")
	}
	id := inst.ID
	w.mu.Lock()
	inst.OfflineSince[p.ID] = time.Now()
	w.Dungeons.pending[p.ID] = &rejoinSlot{InstanceID: id, X: p.X, Y: p.Y, Z: p.Z, Until: time.Now().Add(DungeonRejoinTimeout)}
	w.mu.Unlock()
	if w.Dungeons.pending[p.ID] == nil || time.Until(w.Dungeons.pending[p.ID].Until) < 4*time.Minute {
		t.Fatal("grace 5m")
	}
	leave := w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonLeave, Data: []byte(`{}`)})
	if rejectAction(leave, TypeDungeonLeave) && p.InstanceID != "" {
		t.Fatalf("leave %s", string(leave[0]))
	}
}
