package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPhase16DuelAccept(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("d1", "Raka", 12)
	b := pvpReadyPlayer("d2", "Sinta", 12)
	a.X, a.Z = 0, 0
	b.X, b.Z = 2, 0
	w.Add(a)
	w.Add(b)
	evs := w.ApplyPvp(a.ID, Envelope{Type: TypePvpDuel, Data: []byte(`{"targetId":"d2"}`)})
	if rejectAction(evs, TypePvpDuel) {
		t.Fatalf("challenge rejected %s", string(evs[0]))
	}
	evs = w.ApplyPvp(b.ID, Envelope{Type: TypePvpDuelAccept, Data: []byte(`{}`)})
	inst := w.pvpOf(a.ID)
	if inst == nil {
		t.Fatalf("expected duel instance %d", len(evs))
	}
	if inst.Mode != "DUEL_1V1" {
		t.Fatalf("mode %s", inst.Mode)
	}
	if inst.Ranked {
		t.Fatal("duel must not be ranked")
	}
}

func TestPhase16DuelDecline(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("dd1", "Raka", 12)
	b := pvpReadyPlayer("dd2", "Sinta", 12)
	w.Add(a)
	w.Add(b)
	w.ApplyPvp(a.ID, Envelope{Type: TypePvpDuel, Data: []byte(`{"targetId":"dd2"}`)})
	w.ApplyPvp(b.ID, Envelope{Type: TypePvpDuelDecline, Data: []byte(`{}`)})
	if w.pvpOf(a.ID) != nil {
		t.Fatal("declined duel should not start")
	}
}

func TestPhase16DrawNoRating(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("dr1", "Raka", 20)
	b := pvpReadyPlayer("dr2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	w.ensurePvpRating(a.ensureLog())
	before := a.ensureLog().PvpRating
	inst := startActivePvp(t, w, a, b, "RANKED_1V1")
	w.mu.Lock()
	_ = w.finishPvp(inst, 0, "draw")
	w.mu.Unlock()
	if a.ensureLog().PvpRating != before {
		t.Fatalf("draw changed rating %d -> %d", before, a.ensureLog().PvpRating)
	}
	if len(a.ensureLog().PvpHistory) == 0 || a.ensureLog().PvpHistory[len(a.ensureLog().PvpHistory)-1].Result != "DRAW" {
		t.Fatal("expected DRAW history")
	}
}

func TestPhase16PlacementFive(t *testing.T) {
	if pvpMod.Placement != 5 {
		t.Fatalf("placement %d", pvpMod.Placement)
	}
	w := NewWorldState()
	a := pvpReadyPlayer("pl5", "Raka", 20)
	w.Add(a)
	w.ensurePvpRating(a.ensureLog())
	if a.ensureLog().PvpPlacementLeft != 5 {
		t.Fatalf("left %d", a.ensureLog().PvpPlacementLeft)
	}
}

func TestPhase16Arena3v3Enabled(t *testing.T) {
	c, ok := pvpMode("CASUAL_3V3")
	if !ok || !c.Enabled || c.TeamSize != 3 {
		t.Fatalf("casual 3v3 %+v", c)
	}
	r, ok := pvpMode("RANKED_3V3")
	if !ok || !r.Enabled {
		t.Fatal("ranked 3v3 disabled")
	}
	if c.Map != "mistwood-battlefield" {
		t.Fatalf("map %s", c.Map)
	}
}

func TestPhase16TrainingArena(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("tr1", "Raka", 10)
	a.X, a.Z = 40, 40
	w.Add(a)
	evs := w.ApplyPvp(a.ID, Envelope{Type: TypePvpTraining, Data: []byte(`{}`)})
	if rejectAction(evs, TypePvpTraining) {
		t.Fatal("training rejected")
	}
	if a.Z < 8 {
		t.Fatalf("should move to training dummy pad z=%v", a.Z)
	}
}

func TestPhase16TimerDraw(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("tm1", "Raka", 20)
	b := pvpReadyPlayer("tm2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "CASUAL_1V1")
	a.HP, b.HP = a.MaxHP, b.MaxHP
	w.mu.Lock()
	_ = w.finishPvpOnTime(inst)
	w.mu.Unlock()
	if inst.WinnerTeam != 0 {
		t.Fatalf("equal hp should draw winner=%d", inst.WinnerTeam)
	}
}

func TestPhase16BattlegroundMaps(t *testing.T) {
	def, ok := pvpMode("BATTLEGROUND_5V5")
	if !ok || def.Map != "dawn-arena" {
		t.Fatalf("bg map %s", def.Map)
	}
	w := NewWorldState()
	a := pvpReadyPlayer("bm1", "Raka", 16)
	b := pvpReadyPlayer("bm2", "Sinta", 16)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "BATTLEGROUND_5V5")
	if len(inst.Shrines) != 3 {
		t.Fatalf("shrines %d", len(inst.Shrines))
	}
	if !pvpIsBattleground(inst) {
		t.Fatal("expected battleground")
	}
}

func TestPhase16RankDivision(t *testing.T) {
	rk := rankForRating(1500)
	if rk.ID != "GOLD" {
		t.Fatalf("gold id %s", rk.ID)
	}
	if rankDivision(1500, rk) != "III" {
		t.Fatalf("div %s", rankDivision(1500, rk))
	}
	gm := rankForRating(3600)
	if gm.Name != "Grandmaster" {
		t.Fatalf("grandmaster name %s", gm.Name)
	}
}

func TestPhase16FakeVictoryRejected(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("fv1", "Raka", 20)
	w.Add(a)
	if !rejectAction(w.ApplyPvp(a.ID, Envelope{Type: TypePvpWin, Data: []byte(`{"win":true}`)}), TypePvpWin) {
		t.Fatal("fake win")
	}
	if !rejectAction(w.ApplyPvp(a.ID, Envelope{Type: TypeSetRating, Data: []byte(`{"rating":9000}`)}), TypeSetRating) {
		t.Fatal("fake rating")
	}
}

func TestPhase16SpectateRankedRejected(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("sp1", "Raka", 20)
	b := pvpReadyPlayer("sp2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "RANKED_1V1")
	a.CombatState = "DEAD"
	a.HP = 0
	_ = inst
	if !rejectAction(w.ApplyPvp(a.ID, Envelope{Type: TypePvpSpectate, Data: []byte(`{"targetId":"sp2"}`)}), TypePvpSpectate) {
		t.Fatal("ranked spectate should reject")
	}
}

func TestPhase16ReconnectWindow(t *testing.T) {
	if pvpReconnectWindow() != 120*time.Second {
		t.Fatalf("window %s", pvpReconnectWindow())
	}
}

func TestPhase16ConcurrentMatches(t *testing.T) {
	w := NewWorldState()
	for i := 0; i < 10; i++ {
		a := pvpReadyPlayer("c"+string(rune('a'+i)), "A", 20)
		b := pvpReadyPlayer("c"+string(rune('k'+i)), "B", 20)
		w.Add(a)
		w.Add(b)
		inst := startActivePvp(t, w, a, b, "CASUAL_1V1")
		if inst == nil {
			t.Fatal("match")
		}
	}
	if len(w.PvP.instances) < 10 {
		t.Fatalf("instances %d", len(w.PvP.instances))
	}
}

func TestPhase16PartySize3v3(t *testing.T) {
	w := NewWorldState()
	a := pvpReadyPlayer("ps1", "Raka", 20)
	b := pvpReadyPlayer("ps2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	if w.Parties == nil {
		w.Parties = NewPartyHub()
	}
	pt := &Party{ID: "pty", LeaderID: a.ID, Members: []string{a.ID, b.ID}}
	w.Parties.parties[pt.ID] = pt
	w.Parties.byPlayer[a.ID] = pt.ID
	w.Parties.byPlayer[b.ID] = pt.ID
	a.PartyID, b.PartyID = pt.ID, pt.ID
	if !rejectAction(w.ApplyPvp(a.ID, Envelope{Type: TypePvpQueueJoin, Data: []byte(`{"mode":"CASUAL_3V3"}`)}), TypePvpQueueJoin) {
		t.Fatal("party of 2 cannot queue 3v3")
	}
}

func TestPhase16AfkWarningText(t *testing.T) {
	if pvpAfkWarnText == "" {
		t.Fatal("missing afk copy")
	}
	w := NewWorldState()
	a := pvpReadyPlayer("af1", "Raka", 20)
	b := pvpReadyPlayer("af2", "Sinta", 20)
	w.Add(a)
	w.Add(b)
	inst := startActivePvp(t, w, a, b, "CASUAL_1V1")
	a.LastInputAt = time.Now().Add(-25 * time.Second)
	w.mu.Lock()
	evs := w.tickPvpAfk(inst, time.Now())
	w.mu.Unlock()
	if !inst.Fighters[a.ID].AfkWarned {
		t.Fatalf("expected warning %d", len(evs))
	}
	found := false
	for _, raw := range evs {
		var env Envelope
		_ = json.Unmarshal(raw, &env)
		if env.Type == TypePvpAfk {
			found = true
		}
	}
	if !found {
		t.Fatal("expected PVP_AFK event")
	}
}
