package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func addTestPlayer(w *WorldState, id, name string) *Player {
	p := &Player{ID: id, Name: name, Class: "WARRIOR", Level: 1, HP: 100, MaxHP: 100, send: make(chan []byte, 32)}
	p.initCombat()
	ok, _ := w.Add(p)
	if !ok {
		panic("add " + id)
	}
	return p
}

func TestWorldMultiplayerZones(t *testing.T) {
	w := NewWorldState()
	a := addTestPlayer(w, "p_a", "Raka")
	b := addTestPlayer(w, "p_b", "Lila")
	a.X, a.Z = 0, 6
	b.ensureLog().ForestUnlocked = true
	b.X, b.Z = 0, 36
	w.Simulate(0.05)
	if a.ZoneID == b.ZoneID {
		t.Fatalf("zona harus berbeda: %s %s", a.ZoneID, b.ZoneID)
	}
	sa := w.SnapshotFor(a.ID)
	sb := w.SnapshotFor(b.ID)
	if sa.Online < 2 || sb.Online < 2 {
		t.Fatal("keduanya harus tetap satu world multiplayer")
	}
}

func TestZoneDiscovered(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	p.ensureLog().ForestUnlocked = true
	p.X, p.Z = 0, 30
	evs := w.Simulate(0.05)
	if !p.ensureLog().DiscoveredZones["forest"] {
		t.Fatal("forest harus discovered")
	}
	found := false
	for _, raw := range evs {
		var env Envelope
		_ = json.Unmarshal(raw, &env)
		if env.Type == TypeZoneDiscovered {
			found = true
		}
	}
	if !found {
		t.Fatal("ZONE_DISCOVERED expected")
	}
}

func TestGuardianDefeatPersists(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	w.mu.Lock()
	evs := w.markGuardianDefeated(p, "ragha")
	w.mu.Unlock()
	if len(evs) == 0 {
		t.Fatal("guardian defeat event")
	}
	if !p.ensureLog().Guardians["ragha"] {
		t.Fatal("ragha harus defeated")
	}
	if !p.ensureLog().Flags["chapter_ch01_complete"] && !p.ensureLog().Flags["chapter1_complete"] {
		t.Fatal("chapter progress")
	}
	j := p.worldJournal(w)
	if j.Guardians[0].Status != "DEFEATED" {
		t.Fatalf("journal %s", j.Guardians[0].Status)
	}
	if j.Guardians[1].Status == "LOCKED" {
		t.Fatal("guardian 2 harus unlock")
	}
	w.Remove(p)
	p2 := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 8)}
	p2.initCombat()
	w.Add(p2)
	if !p2.ensureLog().Guardians["ragha"] {
		t.Fatal("reconnect harus tetap defeated")
	}
}

func TestGuardianCheatRejected(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeUnlockGuardian, Data: []byte(`{"id":"avaron"}`)})
	if !isReject(evs) {
		t.Fatal("unlockGuardian cheat harus REJECT")
	}
	evs = w.ApplyExplore(p.ID, Envelope{Type: TypeTeleport, Data: []byte(`{"landmarkId":"celestial-gate"}`)})
	if !isReject(evs) {
		t.Fatal("teleport cheat harus REJECT")
	}
	evs = w.ApplyExplore(p.ID, Envelope{Type: TypeFastTravel, Data: []byte(`{"landmarkId":"celestial-gate"}`)})
	if !isReject(evs) {
		t.Fatal("fast travel undiscovered harus REJECT")
	}
}

func TestFastTravelDiscovered(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	p.ensureLog().Landmarks["old-windmill"] = true
	p.ensureLog().FastTravel["old-windmill"] = true
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeFastTravel, Data: []byte(`{"landmarkId":"old-windmill"}`)})
	if isReject(evs) {
		t.Fatal("landmark discovered harus travel")
	}
	if p.Z < 7 || p.Z > 9 {
		t.Fatalf("posisi travel %v", p.Z)
	}
}

func TestMountLockAndUnlock(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	p.Level = 5
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeRequestMount, Data: []byte(`{"mountId":"wind-runner"}`)})
	if !isReject(evs) {
		t.Fatal("level 5 mount locked")
	}
	p.Level = 15
	q := p.quest("jq001")
	if q == nil {
		t.Fatal("jq001")
	}
	q.State = QuestClaimed
	evs = w.ApplyExplore(p.ID, Envelope{Type: TypeRequestMount, Data: []byte(`{"mountId":"wind-runner"}`)})
	if isReject(evs) || !p.Mounted || p.MountID != "wind-runner" {
		t.Fatalf("mount unlock gagal mounted=%v id=%s reject=%v", p.Mounted, p.MountID, isReject(evs))
	}
	b := addTestPlayer(w, "p_b", "Lila")
	b.X, b.Z = p.X, p.Z
	snap := w.SnapshotFor(b.ID)
	seen := false
	for _, o := range snap.Players {
		if o.PlayerID == p.ID && o.Mounted && o.MountID == "wind-runner" {
			seen = true
		}
	}
	if !seen {
		t.Fatal("player lain harus melihat mount")
	}
	w.Remove(p)
	p2 := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 8)}
	p2.initCombat()
	w.Add(p2)
	owned := false
	for _, id := range p2.ensureLog().Mounts {
		if id == "wind-runner" {
			owned = true
		}
	}
	if !owned {
		t.Fatal("mount persist reconnect")
	}
}

func TestWorldEventNearbyAndScore(t *testing.T) {
	w := NewWorldState()
	near := addTestPlayer(w, "p_a", "Raka")
	far := addTestPlayer(w, "p_b", "Lila")
	near.X, near.Z = 0, 6
	far.ensureLog().ForestUnlocked = true
	far.X, far.Z = 0, 40
	w.mu.Lock()
	w.startWorldEvent("village-defense")
	w.mu.Unlock()
	nearSnap := w.SnapshotFor(near.ID)
	farSnap := w.SnapshotFor(far.ID)
	if nearSnap.Event == nil {
		t.Fatal("player dekat harus menerima event")
	}
	if farSnap.Event != nil {
		t.Fatal("player jauh tidak menerima full event")
	}
	w.ApplyExplore(near.ID, Envelope{Type: TypeJoinWorldEvent})
	evs := w.ApplyExplore(far.ID, Envelope{Type: TypeClaimEventReward, Data: []byte(`{"eventId":"village-defense","score":999999}`)})
	if !isReject(evs) {
		t.Fatal("claim tanpa partisipasi harus REJECT")
	}
}

func TestWeatherServerState(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	w.mu.Lock()
	w.Time.cycleAt = time.Now().Add(-time.Duration(DayCycleSec+1) * time.Second)
	w.mu.Unlock()
	w.Simulate(0.05)
	if w.Time.Phase == "DAY" && w.Time.Weather == "CLEAR" {
		t.Fatal("waktu/weather harus berubah")
	}
	snap := w.SnapshotFor(p.ID)
	if snap.Weather != w.Time.Weather {
		t.Fatalf("weather snapshot %s vs %s", snap.Weather, w.Time.Weather)
	}
}

func TestGuardianCatalog(t *testing.T) {
	if len(guardianCatalog) != 33 {
		t.Fatalf("33 guardians, dapat %d", len(guardianCatalog))
	}
	last := guardianCatalog[32]
	if last.ID != "avaron" || last.Difficulty != "CELESTIAL" {
		t.Fatalf("guardian 33 %s %s", last.ID, last.Difficulty)
	}
}

func TestCelestialGateAfterAvaron(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p_a", "Raka")
	w.mu.Lock()
	w.markGuardianDefeated(p, "avaron")
	w.mu.Unlock()
	if !p.ensureLog().Flags["celestial_gate_unlocked"] {
		t.Fatal("celestial gate harus unlock setelah Avaron")
	}
}

func TestLoreDiscoverOnce(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = -4.2, 4.4
	first := w.Simulate(0.05)
	n := 0
	for _, raw := range first {
		var env Envelope
		_ = json.Unmarshal(raw, &env)
		if env.Type == TypeLoreDiscovered {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("lore pertama harus 1 event, got %d", n)
	}
	again := w.Simulate(0.05)
	for _, raw := range again {
		var env Envelope
		_ = json.Unmarshal(raw, &env)
		if env.Type == TypeLoreDiscovered {
			t.Fatal("lore yang sudah ditemukan tidak boleh spam popup")
		}
	}
}

func TestVillageSafeFromHostileAggro(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 0, 8
	p.HP, p.MaxHP = 100, 100
	w.Simulate(0.2)
	for _, e := range w.enemies {
		if e.Def.Attack <= 0 {
			continue
		}
		if e.Z < 16 {
			t.Fatalf("musuh hostile tidak boleh di alun-alun desa z=%.1f id=%s", e.Z, e.ID)
		}
		if e.TargetID == p.ID {
			t.Fatalf("player di desa tidak boleh di-aggro oleh %s", e.ID)
		}
	}
}

func TestWorldEventWaitsForAdventure(t *testing.T) {
	w, _ := testVillagePlayer()
	w.Events.nextAt = time.Now().Add(-time.Second)
	evs := w.tickWorldEvents(time.Now())
	if w.Events.Active != nil {
		t.Fatal("world event tidak boleh mulai sebelum latihan/hutan terbuka")
	}
	if len(evs) != 0 {
		t.Fatal("tidak boleh ada event combat di awal petualangan")
	}
}

func isReject(evs [][]byte) bool {
	for _, raw := range evs {
		var env Envelope
		_ = json.Unmarshal(raw, &env)
		if env.Type == TypeActionReject {
			return true
		}
	}
	return len(evs) == 0
}
