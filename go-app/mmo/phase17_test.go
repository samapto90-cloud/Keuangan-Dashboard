package mmo

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

func TestPhase17OpenWorldRegions(t *testing.T) {
	if worldCfg.Name != "World of Dawn" {
		t.Fatalf("world name %s", worldCfg.Name)
	}
	need := []string{"village", "forest", "valley", "plains", "canyon", "temple", "masjid"}
	for _, id := range need {
		if regionByID[id].ID == "" {
			t.Fatalf("missing region %s", id)
		}
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatalf("keep Dawn City name, got %s", regionByID["village"].Name)
	}
	if regionByID["village"].Title != "Dawn Village" {
		t.Fatalf("title %s", regionByID["village"].Title)
	}
	w := NewWorldState()
	p := addTestPlayer(w, "p17_a", "Raka")
	p.ensureLog().ForestUnlocked = true
	p.X, p.Z = 0, 6
	w.Simulate(0.05)
	if p.ZoneID != "village" {
		t.Fatalf("village zone %s", p.ZoneID)
	}
	p.X, p.Z = 0, 36
	w.Simulate(0.05)
	if p.ZoneID != "forest" {
		t.Fatalf("forest zone %s", p.ZoneID)
	}
	p.X, p.Z = 0, 58
	w.Simulate(0.05)
	if p.ZoneID != "valley" {
		t.Fatalf("valley zone %s", p.ZoneID)
	}
}

func TestPhase17MiraWelcomeAndQuest(t *testing.T) {
	w, p := testVillagePlayer()
	w.claimTalk(p, "mq001")
	p.X, p.Z = -9.2, 2.2
	w.offerDynamicQuests(p)
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mira","kind":"talk"}`)})
	text := string(evs[0])
	if !containsStr(text, "Selamat datang") && !containsStr(text, "Dawn Village") {
		t.Fatalf("mira welcome: %s", text)
	}
	w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mira","kind":"choice-a"}`)})
	if p.ensureLog().RegionRep["village"] < 1 {
		t.Fatal("choice A harus menambah reputasi desa")
	}
	if p.quest("oq001").State == QuestLocked {
		t.Fatal("oq001 harus terbuka setelah mq001")
	}
	w.acceptQuest(p, "oq001")
	p.ensureLog().ForestUnlocked = true
	p.X, p.Z = 0, 34
	w.Simulate(0.05)
	lm := landmarkByID["edu-shrine"]
	p.X, p.Z = lm.X, lm.Z
	w.autoDiscoverNearby(p)
	if p.quest("oq001").Progress[0] < 1 {
		t.Fatal("talk mira")
	}
	if p.quest("oq001").Progress[1] < 1 {
		t.Fatal("visit forest")
	}
	if p.quest("oq001").Progress[2] < 1 {
		t.Fatal("discover shrine")
	}
}

func TestPhase17ShadowAttackEvent(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 0, 6
	w.mu.Lock()
	evs := w.startWorldEvent("shadow-attack")
	w.mu.Unlock()
	if len(evs) == 0 {
		t.Fatal("event start")
	}
	live := w.Events.Active
	if live == nil || live.Def.Name != "Makhluk Bayangan Menyerang" {
		t.Fatalf("name %+v", live)
	}
	if live.Def.DurationSec != 600 {
		t.Fatalf("duration %v", live.Def.DurationSec)
	}
	join := w.ApplyExplore(p.ID, Envelope{Type: TypeJoinWorldEvent})
	if rejectAction(join, TypeJoinWorldEvent) {
		t.Fatal("join tanpa party harus boleh")
	}
	if live.Score[p.ID] < 1 {
		t.Fatal("contribution")
	}
	v := live.view()
	if v.Phase == "" || v.EndsIn <= 0 {
		t.Fatalf("timer phase=%s endsIn=%d", v.Phase, v.EndsIn)
	}
}

func TestPhase17DayNightSchedule(t *testing.T) {
	w := NewWorldState()
	dayX, dayZ := npcDest(npcByID["child_lina"], w.Time)
	w.Time.AdvanceClock(12 * 60)
	if w.Time.Phase != "NIGHT" {
		t.Fatalf("phase %s clock %s", w.Time.Phase, w.Time.ClockText())
	}
	w.tickNPCs(30)
	nx, nz := w.npcLive(npcByID["child_lina"])
	home := npcPlaces["home"]
	if math.Hypot(nx-home.X, nz-home.Z) > 2.5 {
		t.Fatalf("lina harus pulang malam: %v %v (siang %v %v)", nx, nz, dayX, dayZ)
	}
	kx, kz := npcDest(npcByID["kian"], w.Time)
	if kz < 50 {
		t.Fatalf("kian malam ke valley, z=%v x=%v", kz, kx)
	}
}

func TestPhase17WeatherAndClock(t *testing.T) {
	w, p := testVillagePlayer()
	w.Time.cycleAt = time.Now().Add(-time.Duration(DayCycleSec+1) * time.Second)
	evs := w.Simulate(0.05)
	if w.Time.Phase == "DAY" && w.Time.Weather == "CLEAR" {
		t.Fatal("cycle harus mengubah phase/weather")
	}
	found := false
	for _, raw := range evs {
		var env Envelope
		_ = json.Unmarshal(raw, &env)
		if env.Type == TypeWeatherUpdated {
			found = true
		}
	}
	if !found {
		t.Fatal("WEATHER_UPDATED")
	}
	snap := w.SnapshotFor(p.ID)
	if snap.Clock == "" || snap.ClockLabel == "" {
		t.Fatalf("clock %s %s", snap.Clock, snap.ClockLabel)
	}
	if w.Time.RegionWeather["forest"] == "" && snap.Weather == "" {
		t.Fatal("weather snapshot")
	}
}

func TestPhase17WaypointFastTravel(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().ForestUnlocked = true
	lm := landmarkByID["forest-waypoint"]
	p.X, p.Z = lm.X, lm.Z
	w.mu.Lock()
	w.discoverLandmark(p, lm)
	w.mu.Unlock()
	p.X, p.Z = 0, 6
	p.InCombatUntil = time.Time{}
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeFastTravel, Data: []byte(`{"landmarkId":"forest-waypoint"}`)})
	if rejectAction(evs, TypeFastTravel) {
		t.Fatalf("fast travel %s", evs)
	}
	if math.Hypot(p.X-lm.X, p.Z-lm.Z) > 0.5 {
		t.Fatalf("pos %v %v", p.X, p.Z)
	}
}

func TestPhase17EducationalShrine(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().ForestUnlocked = true
	obj := interactByID["edu-shrine"]
	p.X, p.Z = obj.X, obj.Z
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"edu-shrine"}`)})
	raw := string(evs[0])
	if containsStr(raw, `"correct"`) {
		t.Fatal("client tidak boleh menerima correctAnswer")
	}
	if !containsStr(raw, "3 + 2") {
		t.Fatalf("soal %s", raw)
	}
	beforeKP := p.ensureLog().KnowledgePoints
	p.Energy = 10
	ans, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-add-3-2", Choice: 1})
	w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: ans})
	if p.ensureLog().KnowledgePoints <= beforeKP {
		t.Fatal("knowledge point")
	}
	if p.Energy < 20 {
		t.Fatal("energy reward")
	}
}

func TestPhase17DiscoveryTreasureReconnect(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().ForestUnlocked = true
	fall := landmarkByID["mist-waterfall"]
	p.X, p.Z = fall.X, fall.Z
	w.mu.Lock()
	w.discoverLandmark(p, fall)
	w.mu.Unlock()
	if !p.ensureLog().POI["mist-waterfall"] && !p.ensureLog().Landmarks["mist-waterfall"] {
		t.Fatal("waterfall poi")
	}
	chest := interactByID["chest-1"]
	p.X, p.Z = chest.X, chest.Z
	w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"chest-1"}`)})
	if !p.ensureLog().Claimed["chest-1"] {
		t.Fatal("chest claim")
	}
	again := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"chest-1"}`)})
	if containsStr(string(again[0]), "persediaan") && p.ensureLog().Claimed["chest-1"] {
		// already open text is ok
	}
	w.Remove(p)
	p2 := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 16)}
	p2.initCombat()
	w.Add(p2)
	if !p2.ensureLog().Landmarks["mist-waterfall"] {
		t.Fatal("discovery harus persist setelah reconnect")
	}
}

func TestPhase17Security(t *testing.T) {
	w, p := testVillagePlayer()
	if !rejectAction(w.ApplyWorld(p.ID, Envelope{Type: TypeQuestComplete, Data: []byte(`{"questId":"mq001"}`)}), TypeQuestComplete) {
		t.Fatal("COMPLETE_QUEST")
	}
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeSpawnTreasure}), TypeSpawnTreasure) {
		t.Fatal("SPAWN_TREASURE")
	}
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeSetWorldTime, Data: []byte(`{"clock":999}`)}), TypeSetWorldTime) {
		t.Fatal("SET_WORLD_TIME")
	}
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeStartWorldEvent}), TypeStartWorldEvent) {
		t.Fatal("START_WORLD_EVENT")
	}
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeSetWeather}), TypeSetWeather) {
		t.Fatal("SET_WEATHER")
	}
}

func TestPhase17InterestAndPopulation(t *testing.T) {
	w := NewWorldState()
	for i := 0; i < 100; i++ {
		p := &Player{ID: "c" + itoa(i), Name: "P", send: make(chan []byte, 4)}
		p.initCombat()
		ok, _ := w.Add(p)
		if !ok {
			t.Fatalf("add %d", i)
		}
		p.X, p.Z = float64((i%8)*3), float64(i)*1.4
		if i > 40 {
			p.Z = 90 + float64(i)
		}
	}
	if len(w.players) != 100 {
		t.Fatalf("players %d", len(w.players))
	}
	snap := w.SnapshotFor("c0")
	if snap.Online != 100 {
		t.Fatalf("online %d", snap.Online)
	}
	if len(snap.Players) > 80 {
		t.Fatalf("interest harus memangkas: %d", len(snap.Players))
	}
}

func TestPhase17SubzonesAndFactions(t *testing.T) {
	if areaAt(0, 40) == nil || areaAt(0, 40).Name != "North Forest" {
		t.Fatalf("north forest %+v", areaAt(0, 40))
	}
	if len(factionCatalog) < 4 {
		t.Fatal("factions")
	}
	if eventByID["shadow-attack"].ID == "" || eventByID["dark-fog"].ID == "" {
		t.Fatal("world events catalog")
	}
	if questionByID["q-add-3-2"].Correct != 1 {
		t.Fatal("shrine question")
	}
}

func containsStr(s, sub string) bool {
	return strings.Contains(s, sub)
}
