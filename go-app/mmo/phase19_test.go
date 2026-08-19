package mmo

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestPhase19MountCatalogOriginal(t *testing.T) {
	for _, id := range []string{"dawn-wolf", "sky-fox", "forest-deer", "stone-beast", "celestial-bird", "wind-runner"} {
		if mountByID[id].ID == "" {
			t.Fatalf("mount %s", id)
		}
		if mountByID[id].Speed > MaxMountMult {
			t.Fatalf("p2w speed %s", id)
		}
	}
	if mountByID["wind-runner"].Name != "Wind Runner" || mountByID["wind-runner"].Speed != 1.45 {
		t.Fatal("wind-runner harus tetap")
	}
	if mountByID["dawn-wolf"].Name != "Dawn Wolf" {
		t.Fatal("dawn wolf")
	}
	if eventByID["mount-festival"].DurationSec != 600 {
		t.Fatalf("festival %v", eventByID["mount-festival"].DurationSec)
	}
	if questionByID["q-add-4-3"].ID == "" || questionByID["q-add-1-4"].Correct != 1 {
		t.Fatal("questions")
	}
	if questByID["mtq001"].Title != "Teman Perjalanan" || questByID["mtq002"].Title != "Latihan Menunggang" {
		t.Fatal("mount quest")
	}
	if landmarkByID["hidden-trail"].Hidden != true {
		t.Fatal("hidden path")
	}
}

func TestPhase19WalkSprintMount(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19a", "Raka")
	p.AX, p.AZ = 0, 1
	p.Simulate(0.05)
	walk := math.Hypot(p.VX, p.VZ)
	p.Sprint = true
	p.Stamina = 100
	p.Simulate(0.05)
	run := math.Hypot(p.VX, p.VZ)
	if run <= walk {
		t.Fatalf("sprint %v walk %v", run, walk)
	}
	p.Level = 15
	p.quest("jq001").State = QuestClaimed
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeRequestMount, Data: []byte(`{"mountId":"wind-runner"}`)})
	if isReject(evs) || !p.Mounted {
		t.Fatal("summon")
	}
	p.VX, p.VZ = 0, 0
	p.AX, p.AZ = 0, 1
	p.Simulate(0.05)
	mt := math.Hypot(p.VX, p.VZ)
	if mt <= walk {
		t.Fatalf("mount %v walk %v", mt, walk)
	}
	w.ApplyExplore(p.ID, Envelope{Type: TypeDismount, Data: []byte(`{}`)})
	if p.Mounted {
		t.Fatal("dismiss")
	}
}

func TestPhase19DawnWolfQuestAndDuplicate(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19b", "Raka")
	p.quest("mq001").State = QuestClaimed
	w.mu.Lock()
	w.refreshAvailability(p)
	w.mu.Unlock()
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeQuestAccept, Data: []byte(`{"questId":"mtq001"}`)})
	if isReject(evs) {
		t.Fatal("accept mtq001")
	}
	nara := npcByID["nara"]
	p.X, p.Z = nara.X, nara.Z
	evs = w.ApplyWorld(p.ID, Envelope{Type: TypeQuestClaim, Data: []byte(`{"questId":"mtq001"}`)})
	if isReject(evs) {
		t.Fatal("claim mtq001")
	}
	if !p.ownsMount("dawn-wolf") {
		t.Fatal("dawn wolf reward")
	}
	w.mu.Lock()
	if w.grantMount(p, "dawn-wolf") {
		t.Fatal("duplicate mount")
	}
	w.mu.Unlock()
	evs = w.ApplyExplore(p.ID, Envelope{Type: TypeGrantMount, Data: []byte(`{"mountId":"sky-fox"}`)})
	if !isReject(evs) {
		t.Fatal("client grant ditolak")
	}
	evs = w.ApplyExplore(p.ID, Envelope{Type: TypeSetMountSpeed, Data: []byte(`{"speed":9}`)})
	if !isReject(evs) {
		t.Fatal("client speed ditolak")
	}
	evs = w.ApplyExplore(p.ID, Envelope{Type: TypeRequestMount, Data: []byte(`{"mountId":"dawn-wolf"}`)})
	if isReject(evs) || !p.Mounted || p.MountID != "dawn-wolf" {
		t.Fatal("summon dawn wolf")
	}
}

func TestPhase19MountCombatRestricted(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19c", "Raka")
	p.Level = 15
	p.quest("jq001").State = QuestClaimed
	p.InCombatUntil = time.Now().Add(time.Second)
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeRequestMount, Data: []byte(`{"mountId":"wind-runner"}`)})) {
		t.Fatal("combat mount")
	}
	p.InCombatUntil = time.Time{}
	p.InstanceID = "dun-1"
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeRequestMount, Data: []byte(`{"mountId":"wind-runner"}`)})) {
		t.Fatal("dungeon mount")
	}
	p.InstanceID = ""
	p.ensureLog().PendingCinematic = "cin-1"
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeRequestMount, Data: []byte(`{"mountId":"wind-runner"}`)})) {
		t.Fatal("cinematic mount")
	}
}

func TestPhase19FastTravelAndStoryLock(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19d", "Raka")
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeFastTravel, Data: []byte(`{"landmarkId":"forest-waypoint"}`)})) {
		t.Fatal("locked waypoint")
	}
	p.ensureLog().Landmarks["old-windmill"] = true
	p.ensureLog().FastTravel["old-windmill"] = true
	p.InCombatUntil = time.Now().Add(time.Second)
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeFastTravel, Data: []byte(`{"landmarkId":"old-windmill"}`)})) {
		t.Fatal("combat travel")
	}
	p.InCombatUntil = time.Time{}
	p.ensureLog().Flags["fast_travel_locked"] = true
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeFastTravel, Data: []byte(`{"landmarkId":"old-windmill"}`)})) {
		t.Fatal("story lock")
	}
	p.ensureLog().Flags["fast_travel_locked"] = false
	p.Mounted = true
	p.MountID = "wind-runner"
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeFastTravel, Data: []byte(`{"landmarkId":"old-windmill"}`)})
	if isReject(evs) || p.Mounted {
		t.Fatal("travel dismount")
	}
}

func TestPhase19AntiCheat(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19e", "Raka")
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeTeleport, Data: []byte(`{"landmarkId":"celestial-gate"}`)})) {
		t.Fatal("teleport")
	}
	p.MoveSpeedBonus = 40
	p.AX, p.AZ = 0, 1
	p.Sprint = true
	p.Simulate(0.05)
	if math.Hypot(p.VX, p.VZ) > RunSpeed*MaxMountMult+0.05 {
		t.Fatalf("speed hack %v", math.Hypot(p.VX, p.VZ))
	}
}

func TestPhase19SwimShore(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19f", "Raka")
	p.X, p.Z = -11, 42
	p.Stamina = 0.01
	p.HP = 80
	p.Simulate(0.05)
	if p.HP != 80 {
		t.Fatal("swim death")
	}
	if !nearby(p.X, p.Z, -8, 40, 1.5) {
		t.Fatalf("shore %v %v", p.X, p.Z)
	}
}

func TestPhase19RaceOrder(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19g", "Raka")
	p.X, p.Z = 4, 8
	if isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeRaceStart, Data: []byte(`{}`)})) {
		t.Fatal("race start")
	}
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeRaceCheckpoint, Data: []byte(`{"landmarkId":"race-cp-3"}`)})) {
		t.Fatal("out of order")
	}
	p.X, p.Z = 0, 18
	if isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeRaceCheckpoint, Data: []byte(`{"landmarkId":"race-cp-2"}`)})) {
		t.Fatal("cp2")
	}
	p.X, p.Z = 0, 32
	if isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeRaceCheckpoint, Data: []byte(`{"landmarkId":"race-cp-3"}`)})) {
		t.Fatal("cp3")
	}
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeRaceFinish, Data: []byte(`{}`)})
	if isReject(evs) {
		t.Fatal("finish")
	}
	if !isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeRaceFinish, Data: []byte(`{}`)})) {
		t.Fatal("second finish")
	}
}

func TestPhase19PartyWaypointNoTeleport(t *testing.T) {
	w := NewWorldState()
	a := addTestPlayer(w, "p19h", "Raka")
	b := addTestPlayer(w, "p19i", "Lila")
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p19i"}`)})
	w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	bx, bz := b.X, b.Z
	evs := w.ApplyParty(a.ID, Envelope{Type: TypePartySetWaypoint, Data: []byte(`{"landmarkId":"old-windmill"}`)})
	if isReject(evs) {
		t.Fatal("waypoint")
	}
	pt := w.Parties.Of(a.ID)
	if pt == nil || pt.WaypointID != "old-windmill" {
		t.Fatal("marker")
	}
	view := w.partyView(pt, b)
	if view.WaypointID != "old-windmill" {
		t.Fatal("member marker")
	}
	w.ApplyParty(b.ID, Envelope{Type: TypeFollowParty, Data: []byte(`{}`)})
	if b.X != bx || b.Z != bz {
		t.Fatal("auto teleport")
	}
	if !b.FollowParty {
		t.Fatal("follow")
	}
}

func TestPhase19ExplorationPuzzleAndPersist(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19j", "Raka")
	p.ensureLog().ForestUnlocked = true
	p.X, p.Z = 0, 36
	w.Simulate(0.05)
	if !p.ensureLog().DiscoveredZones["forest"] {
		t.Fatal("region")
	}
	p.X, p.Z = -11, 42
	w.Simulate(0.05)
	if !p.ensureLog().Landmarks["mist-waterfall"] {
		t.Fatal("landmark")
	}
	p.X, p.Z = 10, 39
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"travel-puzzle"}`)})
	if isReject(evs) {
		t.Fatal("puzzle open")
	}
	evs = w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: []byte(`{"questionId":"q-add-1-4","choice":1}`)})
	if isReject(evs) {
		t.Fatal("puzzle answer")
	}
	if !p.ensureLog().Flags["hidden_path_open"] {
		t.Fatal("hidden")
	}
	col := w.mountCollection(p)
	raw, _ := json.Marshal(col)
	if !jsonContains(raw, "dawn-wolf") && !jsonContains(raw, "???") {
		t.Fatal("collection")
	}
	w.Remove(p)
	p2 := &Player{ID: "p19j", Name: "Raka", send: make(chan []byte, 8)}
	p2.initCombat()
	w.Add(p2)
	if !p2.ensureLog().Landmarks["mist-waterfall"] {
		t.Fatal("persist landmark")
	}
}

func TestPhase19TravelEventHelp(t *testing.T) {
	w := NewWorldState()
	p := addTestPlayer(w, "p19k", "Raka")
	p.ensureLog().TravelEvent = "lost-traveler"
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeTravelEvent, Data: []byte(`{"kind":"lost-traveler","action":"help"}`)})
	if isReject(evs) {
		t.Fatal("help")
	}
	p.ensureLog().TravelEvent = "lost-traveler"
	if isReject(w.ApplyExplore(p.ID, Envelope{Type: TypeTravelEvent, Data: []byte(`{"action":"ignore"}`)})) {
		t.Fatal("ignore")
	}
}

func jsonContains(raw []byte, s string) bool {
	return len(raw) > 0 && (string(raw) == s || len(s) > 0 && (func() bool {
		for i := 0; i+len(s) <= len(raw); i++ {
			if string(raw[i:i+len(s)]) == s {
				return true
			}
		}
		return false
	})())
}
