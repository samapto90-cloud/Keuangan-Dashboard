package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func pvpReadyPlayer(id, name string, level int) *Player {
	p := &Player{ID: id, Name: name, send: make(chan []byte, 64), Region: "ASIA"}
	p.initCombat()
	p.Level = level
	p.LastInputAt = time.Now()
	return p
}

func startActivePvp(t *testing.T, w *WorldState, a, b *Player, mode string) *PvpInstance {
	t.Helper()
	def, ok := pvpMode(mode)
	if !ok {
		t.Fatalf("mode %s", mode)
	}
	now := time.Now()
	inst := w.newPvpInstance([]string{a.ID, b.ID}, def, now)
	w.beginPvp(inst)
	inst.State = PvpActive
	inst.CountdownUntil = now.Add(-time.Second)
	inst.ProtectUntil = now.Add(-time.Second)
	a.PvpNoActUntil = time.Time{}
	b.PvpNoActUntil = time.Time{}
	a.InvulnUntil = time.Time{}
	b.InvulnUntil = time.Time{}
	a.AttackCDUntil = time.Time{}
	b.AttackCDUntil = time.Time{}
	a.X, a.Z = 0, 0
	b.X, b.Z = 1.2, 0
	return inst
}

func pvpPunch(w *WorldState, p *Player, target string) [][]byte {
	p.AttackCDUntil = time.Time{}
	p.PvpNoActUntil = time.Time{}
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: target, Timestamp: time.Now().UnixMilli()})
	return w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
}

func TestPvpADefeatsB(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pa", "Raka", 20)
	b := pvpReadyPlayer("pb", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "CASUAL_1V1")
	b.HP = 8
	var last [][]byte
	for i := 0; i < 20 && b.alive(); i++ {
		last = pvpPunch(w, a, b.ID)
	}
	if b.alive() {
		t.Fatalf("B should be defeated %s", string(last[0]))
	}
	w.mu.Lock()
	if inst.State != PvpCompleted && inst.State != PvpEnding {
		_ = w.tickPvpActive(inst, time.Now())
	}
	w.mu.Unlock()
	if inst.WinnerTeam != 1 {
		w.mu.Lock()
		_ = w.finishPvpOnTime(inst)
		w.mu.Unlock()
	}
	// wipe rule should already finish
	w.mu.Lock()
	if inst.State != PvpCompleted {
		_ = w.finishPvp(inst, 1, "defeat")
	}
	wonA := inst.WinnerTeam == 1
	w.mu.Unlock()
	if !wonA {
		t.Fatal("A should win")
	}
	if a.ensureLog().PvpWins < 1 {
		t.Fatal("A win stat")
	}
	if b.ensureLog().PvpLosses < 1 {
		t.Fatal("B loss stat")
	}
}

func TestPvpFakeDamageRejected(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pfd", "Raka", 20)
	w.Add(a)
	evs := w.ApplyCombat(a.ID, Envelope{Type: TypeSetDamage, Data: []byte(`{"damage":99999}`)})
	if !rejectAction(evs, TypeSetDamage) {
		t.Fatal("fake damage")
	}
}

func TestPvpReconnectWindow(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pr1", "Raka", 20)
	b := pvpReadyPlayer("pr2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "CASUAL_1V1")
	id := inst.ID
	w.Remove(a)
	c := pvpReadyPlayer("pr1", "Raka", 20)
	ok, _ := w.Add(c)
	if !ok {
		t.Fatal("rejoin add")
	}
	if c.InstanceID != id {
		t.Fatalf("expected return to match got %s", c.InstanceID)
	}
}

func TestPvpDisconnectLoss(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pd1", "Raka", 20)
	b := pvpReadyPlayer("pd2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "RANKED_1V1")
	w.mu.Lock()
	inst.Offline[a.ID] = time.Now().Add(-6 * time.Minute)
	delete(w.players, a.ID)
	evs := w.tickPvpDisconnect(inst, time.Now())
	w.mu.Unlock()
	if inst.WinnerTeam != 2 {
		t.Fatalf("ranked disconnect should loss team2 win state=%s winner=%d evs=%d", inst.State, inst.WinnerTeam, len(evs))
	}
}

func TestPvpAfkWarning(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pk1", "Raka", 20)
	b := pvpReadyPlayer("pk2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "CASUAL_1V1")
	a.LastInputAt = time.Now().Add(-25 * time.Second)
	w.mu.Lock()
	evs := w.tickPvpAfk(inst, time.Now())
	warned := inst.Fighters[a.ID].AfkWarned
	w.mu.Unlock()
	if !warned {
		t.Fatalf("expected afk warning %d", len(evs))
	}
	a.LastInputAt = time.Now().Add(-50 * time.Second)
	w.mu.Lock()
	_ = w.tickPvpAfk(inst, time.Now())
	flagged := inst.Fighters[a.ID].AfkFlagged
	w.mu.Unlock()
	if !flagged {
		t.Fatal("expected afk penalty")
	}
}

func TestPvpGoldPromotion(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pg1", "Raka", 20)
	b := pvpReadyPlayer("pg2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	a.ensureLog().PvpRating = 1490
	a.ensureLog().PvpMatches = 12
	a.ensureLog().PvpRankedMatches = 12
	a.ensureLog().PvpPlacementLeft = 0
	inst := startActivePvp(t, w, a, b, "RANKED_1V1")
	w.mu.Lock()
	_ = w.finishPvp(inst, 1, "defeat")
	w.mu.Unlock()
	if a.ensureLog().PvpRating < 1500 {
		t.Fatalf("rating %d", a.ensureLog().PvpRating)
	}
	if rankForRating(a.ensureLog().PvpRating).ID != "GOLD" {
		t.Fatalf("rank %s", rankForRating(a.ensureLog().PvpRating).ID)
	}
}

func TestPvpRatingLoss(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pl1", "Raka", 20)
	b := pvpReadyPlayer("pl2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	w.ensurePvpRating(a.ensureLog())
	before := a.ensureLog().PvpRating
	inst := startActivePvp(t, w, a, b, "RANKED_1V1")
	w.mu.Lock()
	_ = w.finishPvp(inst, 2, "defeat")
	w.mu.Unlock()
	if a.ensureLog().PvpRating >= before {
		t.Fatalf("rating should drop %d -> %d", before, a.ensureLog().PvpRating)
	}
}

func TestPvpDuplicateResult(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("px1", "Raka", 20)
	b := pvpReadyPlayer("px2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "RANKED_1V1")
	w.mu.Lock()
	_ = w.finishPvp(inst, 1, "defeat")
	wins := a.ensureLog().PvpWins
	rating := a.ensureLog().PvpRating
	again := w.finishPvp(inst, 1, "defeat")
	w.mu.Unlock()
	if a.ensureLog().PvpWins != wins || a.ensureLog().PvpRating != rating {
		t.Fatal("duplicate processed")
	}
	if again == nil {
		t.Fatal("duplicate should return cached tx")
	}
}

func TestBattlegroundCaptureScore(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pc1", "Raka", 16)
	b := pvpReadyPlayer("pc2", "Sinta", 16)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "BATTLEGROUND_5V5")
	a.X, a.Z = -12, 0
	b.X, b.Z = 18, 8
	w.mu.Lock()
	s := inst.Shrines[0]
	s.ProgressA = 100
	s.Owner = 1
	s.Contested = false
	s.LastTick = time.Time{}
	before := inst.ScoreA
	_ = w.tickPvpCapture(inst, time.Now())
	after := inst.ScoreA
	w.mu.Unlock()
	if after <= before {
		t.Fatal("captured shrine should tick score")
	}
}

func TestBattlegroundContested(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pt1", "Raka", 16)
	b := pvpReadyPlayer("pt2", "Sinta", 16)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "BATTLEGROUND_5V5")
	a.X, a.Z = -12, 0
	b.X, b.Z = -12, 0
	w.mu.Lock()
	s := inst.Shrines[0]
	s.Owner = 1
	s.ProgressA = 100
	s.LastTick = time.Time{}
	before := inst.ScoreA
	_ = w.tickPvpCapture(inst, time.Now())
	if !s.Contested {
		w.mu.Unlock()
		t.Fatal("expected contested")
	}
	if inst.ScoreA != before {
		w.mu.Unlock()
		t.Fatal("no score while contested")
	}
	w.mu.Unlock()
}

func TestPvpTransformModifier(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pf1", "Raka", 40)
	w.Add(a)
	a.FormID = "celestial-4"
	a.TransformState = "TRANSFORMED"
	f := pvpTransformFactor(a)
	if f >= 1 {
		t.Fatalf("pvp transform should reduce bonus %f", f)
	}
	if f > 0.9 {
		t.Fatalf("celestial-4 pvp factor too high %f", f)
	}
}

func TestPvpGoldReward(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("prw", "Raka", 20)
	w.Add(a)
	w.mu.Lock()
	w.grantPvpRankReward(a, "GOLD")
	w.mu.Unlock()
	found := false
	for _, c := range a.ensureLog().Cosmetics {
		if c == "badge-golden-arena" {
			found = true
		}
	}
	if !found {
		t.Fatal("gold badge")
	}
}

func TestPvpLeaderboardServer(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("plb", "Raka", 20)
	w.Add(a)
	a.ensureLog().PvpRating = 1800
	a.ensureLog().PvpWins = 4
	w.mu.Lock()
	w.rebuildLeaderboard()
	lb := w.pvpLeaderboard(a, "GLOBAL")
	w.mu.Unlock()
	entries, _ := lb["entries"].([]LBEntry)
	if len(entries) == 0 || entries[0].Rating != 1800 {
		t.Fatal("leaderboard rating")
	}
	evs := w.ApplyPvp(a.ID, Envelope{Type: TypeSetRating, Data: []byte(`{"rating":99999}`)})
	if !rejectAction(evs, TypeSetRating) {
		t.Fatal("client rating")
	}
}

func TestPvpSecurityRejects(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("ps1", "Raka", 20)
	w.Add(a)
	if !rejectAction(w.ApplyPvp(a.ID, Envelope{Type: TypeSetRank, Data: []byte(`{"rank":"CELESTIAL"}`)}), TypeSetRank) {
		t.Fatal("rank")
	}
	if !rejectAction(w.ApplyPvp(a.ID, Envelope{Type: TypePvpWin, Data: []byte(`{"win":true}`)}), TypePvpWin) {
		t.Fatal("win")
	}
}

func TestPvpMatchmakingReady(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("pm1", "Raka", 20)
	b := pvpReadyPlayer("pm2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	if rejectAction(w.ApplyPvp(a.ID, Envelope{Type: TypePvpQueueJoin, Data: []byte(`{"mode":"CASUAL_1V1"}`)}), TypePvpQueueJoin) {
		t.Fatal("queue a")
	}
	if rejectAction(w.ApplyPvp(b.ID, Envelope{Type: TypePvpQueueJoin, Data: []byte(`{"mode":"CASUAL_1V1"}`)}), TypePvpQueueJoin) {
		t.Fatal("queue b")
	}
	inst := w.pvpOf(a.ID)
	if inst == nil {
		t.Fatal("matched")
	}
	w.ApplyPvp(a.ID, Envelope{Type: TypePvpReady, Data: []byte(`{"ready":true}`)})
	w.ApplyPvp(b.ID, Envelope{Type: TypePvpReady, Data: []byte(`{"ready":true}`)})
	if inst.State != PvpCountdown && inst.State != PvpActive && inst.State != PvpLoading {
		t.Fatalf("state %s", inst.State)
	}
}

func TestPvpLobbyLevelGate(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("lv", "Raka", 5)
	w.Add(a)
	if !rejectAction(w.ApplyPvp(a.ID, Envelope{Type: TypePvpQueueJoin, Data: []byte(`{"mode":"CASUAL_1V1"}`)}), TypePvpQueueJoin) {
		t.Fatal("level 10 required")
	}
}
