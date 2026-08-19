package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRegionStreamingNearbyFar(t *testing.T) {
	forest := regionByID["forest"]
	valley := regionByID["valley"]
	canyon := regionByID["canyon"]
	if !regionStreamVisible(30, forest) {
		t.Fatal("Mistwood current harus load")
	}
	if !regionStreamVisible(30, valley) {
		t.Fatal("Stone Valley nearby dari Mistwood harus load")
	}
	if regionStreamVisible(30, canyon) {
		t.Fatal("Crimson Temple jauh dari Mistwood harus unload")
	}
	if !regionStreamVisible(61, valley) {
		t.Fatal("Stone Valley current")
	}
	if regionStreamVisible(104, forest) {
		t.Fatal("Mistwood harus unload dari Crimson Temple")
	}
}

func TestChannelSwitchHidesPlayers(t *testing.T) {
	w := NewWorldState()
	a := addTestPlayer(w, "p_a", "Raka")
	b := addTestPlayer(w, "p_b", "Lila")
	a.X, a.Z, b.X, b.Z = 0, 6, 1, 6
	if a.channelID() != b.channelID() {
		t.Fatalf("default channel %s %s", a.channelID(), b.channelID())
	}
	evs := w.ApplyExplore(a.ID, Envelope{Type: TypeSwitchChannel, Data: []byte(`{"channel":"2"}`)})
	if isReject(evs) {
		t.Fatal("switch channel 2")
	}
	snap := w.SnapshotFor(a.ID)
	for _, o := range snap.Players {
		if o.PlayerID == b.ID {
			t.Fatal("channel lain tidak boleh tampil")
		}
	}
}

func TestMonsterInvasionFlow(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	p.ensureLog().ForestUnlocked = true
	p.X, p.Z = 0, 30
	w.mu.Lock()
	w.startWorldEvent("monster-invasion")
	w.mu.Unlock()
	snap := w.SnapshotFor(p.ID)
	if snap.Event == nil || snap.Event.State != "ACTIVE" {
		t.Fatal("announcement/marker invasion")
	}
	w.ApplyExplore(p.ID, Envelope{Type: TypeJoinWorldEvent})
	w.mu.Lock()
	w.finishWorldEvent(w.Events.Active, true)
	w.mu.Unlock()
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeClaimEventReward, Data: []byte(`{"eventId":"monster-invasion"}`)})
	if isReject(evs) {
		t.Fatal("reward invasion")
	}
	again := w.ApplyExplore(p.ID, Envelope{Type: TypeClaimEventReward, Data: []byte(`{"eventId":"monster-invasion"}`)})
	if !isReject(again) {
		t.Fatal("event reward duplikat")
	}
}

func TestGuardianJournalOnce(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	w.mu.Lock()
	w.markGuardianDefeated(p, "ragha")
	w.mu.Unlock()
	j := p.worldJournal(w)
	if j.GuardiansDefeated != 1 || j.GuardiansTotal != 33 {
		t.Fatalf("journal %d/%d", j.GuardiansDefeated, j.GuardiansTotal)
	}
	if j.Tokens < 1 {
		t.Fatal("guardian token")
	}
	if j.Guardians[0].Name != "Kabutra" {
		t.Fatalf("nama original %s", j.Guardians[0].Name)
	}
	tok := p.ensureLog().GuardianTokens
	w.mu.Lock()
	w.markGuardianDefeated(p, "ragha")
	w.mu.Unlock()
	if p.ensureLog().GuardianTokens != tok {
		t.Fatal("reward guardian harus sekali")
	}
}

func TestMainQuestPersists(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	p.ensureLog().ForestUnlocked = true
	q := p.quest("mq010")
	if q == nil {
		t.Fatal("mq010")
	}
	q.State = QuestActive
	p.X, p.Z = 0, 30
	w.Simulate(0.05)
	if q.State != QuestCompleted && q.Progress[0] < 1 {
		t.Fatalf("visit mistwood progress %v state %s", q.Progress, q.State)
	}
	w.Remove(p)
	p2 := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 8)}
	p2.initCombat()
	w.Add(p2)
	q2 := p2.quest("mq010")
	if q2 == nil || (q2.State != QuestCompleted && q2.Progress[0] < 1) {
		t.Fatal("quest harus persist reconnect")
	}
}

func TestAuraITrialPersists(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	p.Level = 10
	p.Energy = p.MaxEnergy
	p.ensureLog().Quests["tq001"] = &QuestLog{ID: "tq001", State: QuestClaimed}
	raw, _ := json.Marshal(TransformIn{FormID: "aura-1"})
	evs := w.ApplyProgression(p.ID, Envelope{Type: TypeRequestTransformation, Data: raw})
	if hasType(evs, TypeTransformationRejected) {
		t.Fatal("Aura I trial")
	}
	w.Remove(p)
	p2 := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 8)}
	p2.initCombat()
	w.Add(p2)
	if !p2.hasForm("aura-1") && !p2.ensureLog().Flags["aura_1_unlocked"] {
		t.Fatal("Aura I persist reconnect")
	}
}

func TestMasjidCombatDisabled(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	p.ensureLog().Flags["celestial_gate_unlocked"] = true
	p.ensureLog().ForestUnlocked = true
	for _, r := range regionCatalog {
		p.ensureLog().Flags[r.RequiredFlag] = true
	}
	p.X, p.Z = 0, 166
	w.Simulate(0.05)
	if p.ZoneID != "masjid" {
		t.Fatalf("zone %s", p.ZoneID)
	}
	if !p.ensureLog().Flags["masjid_entered"] {
		t.Fatal("story masjid")
	}
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: []byte(`{"attackType":"light"}`)})
	if !isReject(evs) {
		t.Fatal("combat di Masjid Cahaya harus ditolak")
	}
}

func TestSafeZoneOpenWorldPvpRejected(t *testing.T) {
	w := NewWorldState()
	a := addTestPlayer(w, "p_a", "Raka")
	b := addTestPlayer(w, "p_b", "Lila")
	a.X, a.Z, b.X, b.Z = 0, 6, 1, 6
	raw, _ := json.Marshal(AttackIn{AttackType: "light", TargetID: b.ID, Timestamp: time.Now().UnixMilli()})
	evs := w.ApplyCombat(a.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if !isReject(evs) {
		t.Fatal("PvP Dawn City harus ditolak")
	}
}

func TestWorldBossSharedHPAndContribution(t *testing.T) {
	w := NewWorldState()
	a := addTestPlayer(w, "p_a", "Raka")
	b := addTestPlayer(w, "p_b", "Lila")
	a.ensureLog().ForestUnlocked = true
	b.ensureLog().ForestUnlocked = true
	a.ensureLog().Flags["chapter_ch01_complete"] = true
	b.ensureLog().Flags["chapter_ch01_complete"] = true
	w.mu.Lock()
	w.spawnWorldBoss()
	w.mu.Unlock()
	if w.Boss == nil {
		t.Fatal("boss spawn")
	}
	e := w.enemies[w.Boss.EnemyID]
	a.X, a.Z = e.X, e.Z
	b.X, b.Z = e.X, e.Z
	hp := e.HP
	raw, _ := json.Marshal(AttackIn{AttackType: "light", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	w.ApplyCombat(a.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	raw, _ = json.Marshal(AttackIn{AttackType: "light", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	w.ApplyCombat(b.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if e.HP >= hp {
		t.Fatal("shared HP harus berkurang")
	}
	if w.Boss.Damage[a.ID] < 1 || w.Boss.Damage[b.ID] < 1 {
		t.Fatal("kontribusi server-side")
	}
	w.ApplyExplore(a.ID, Envelope{Type: TypeTriggerWorldBoss})
	if !isReject(w.ApplyExplore(a.ID, Envelope{Type: TypeSetWorldBossHP, Data: []byte(`{"hp":1}`)})) {
		t.Fatal("SET_WORLD_BOSS_HP")
	}
}

func TestWorldBossRewardNoDuplicate(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	w.mu.Lock()
	w.spawnWorldBoss()
	w.Boss.Damage[p.ID] = 40
	w.Boss.State = "SUCCESS"
	w.mu.Unlock()
	claim, _ := json.Marshal(map[string]any{"bossId": w.Boss.Def.ID, "transactionId": "tx-wb-1"})
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeClaimWorldBoss, Data: claim})
	if isReject(evs) {
		t.Fatalf("claim %s", string(evs[0]))
	}
	w.Remove(p)
	p2 := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 8)}
	p2.initCombat()
	w.Add(p2)
	bossClaim, _ := json.Marshal(map[string]any{"bossId": w.Boss.Def.ID, "transactionId": "tx-wb-1"})
	again := w.ApplyExplore(p2.ID, Envelope{Type: TypeClaimWorldBoss, Data: bossClaim})
	if !isReject(again) {
		dupRaw, _ := json.Marshal(map[string]any{"bossId": w.Boss.Def.ID, "transactionId": "tx-wb-2"})
		dup := w.ApplyExplore(p2.ID, Envelope{Type: TypeClaimWorldBoss, Data: dupRaw})
		if !isReject(dup) {
			t.Fatal("reward world boss tidak boleh duplikat setelah reconnect")
		}
	}
}

func TestOpenWorldQuestTravel(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	log := p.ensureLog()
	log.ForestUnlocked = true
	log.Flags["chapter_ch01_complete"] = true
	log.Flags["chapter_ch02_complete"] = true
	log.Flags["chapter_ch05_complete"] = true
	q := p.quest("mq013")
	if q == nil {
		t.Fatal("mq013")
	}
	q.State = QuestActive
	p.X, p.Z = 0, 30
	w.Simulate(0.05)
	p.X, p.Z = 0, 60
	w.Simulate(0.05)
	p.X, p.Z = 0, 100
	w.Simulate(0.05)
	if p.quest("mq013").State != QuestCompleted && p.quest("mq013").Progress[0] < 1 {
		t.Fatal("quest travel tidak boleh rusak")
	}
}

func TestMasjidIsSafeZone(t *testing.T) {
	r := regionByID["masjid"]
	if r.ID == "" || !r.SafeZone || !r.CombatDisabled {
		t.Fatal("Masjid Cahaya harus safe zone")
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatalf("dawn %s", regionByID["village"].Name)
	}
}

func TestSilumanRegistry33(t *testing.T) {
	if len(guardianCatalog) != 33 {
		t.Fatalf("%d", len(guardianCatalog))
	}
	last := guardianCatalog[32]
	if last.ID != "avaron" || last.Personality == "" {
		t.Fatal("Avaron personality")
	}
	if last.Defeat == "" || last.Name != "Avaron" {
		t.Fatal("avaron story")
	}
}

func TestTransformationNamesOriginal(t *testing.T) {
	if transformByID["aura-1"].Name != "AURA ASCENSION I" {
		t.Fatalf("%s", transformByID["aura-1"].Name)
	}
	if transformByID["celestial-4"].Name != "CELESTIAL AURA IV" {
		t.Fatalf("%s", transformByID["celestial-4"].Name)
	}
}

func TestWorldLimitAllowsMasjid(t *testing.T) {
	if WorldLimit < 175 {
		t.Fatal("world limit masjid")
	}
}
