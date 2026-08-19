package mmo

import (
	"math"
	"time"
)

// Phase 19 overlay: TravelService + MountService on WorldState.
// Logical tables persist on PlayerLog (not a second store):
// mounts, player_mounts, mount_favorites, mount_cosmetics,
// waypoints, player_waypoints, travel_events, landmarks,
// player_landmarks, exploration_progress, mount_races, mount_race_results.
// Indexes: playerId, mountId, regionId, waypointId, landmarkId.

const (
	MaxMountMult     = 1.55
	MaxMoveSpeed     = 16.0
	SwimStaminaDrain = 2.0
	RaceMaxSec       = 180
	MountAnimIdle    = "IDLE"
	MountAnimSummon  = "SUMMONING"
	MountAnimMounted = "MOUNTED"
	MountAnimMoving  = "MOVING"
	MountAnimDismount = "DISMOUNTING"
)

type waterVolume struct {
	ID              string
	X, Z, R         float64
	ShoreX, ShoreZ  float64
}

type climbPoint struct {
	ID      string
	X, Z, H float64
}

type raceDef struct {
	ID          string
	Checkpoints []string
	Exp, Coin   int
	Cosmetic    string
}

type travelArch struct {
	CanFly   bool
	CanSwim  bool
	CanGlide bool
	CanClimb bool
}

var waterVolumes = []waterVolume{
	{ID: "mist-pool", X: -11, Z: 42, R: 5.5, ShoreX: -8, ShoreZ: 40},
	{ID: "river-light", X: 0, Z: 80, R: 7, ShoreX: 4, ShoreZ: 78},
}

var climbPoints = []climbPoint{
	{ID: "valley-cliff", X: 2, Z: 58, H: 4},
}

var dawnRace = raceDef{
	ID: "dawn-race", Checkpoints: []string{"race-cp-1", "race-cp-2", "race-cp-3"},
	Exp: 40, Coin: 12, Cosmetic: "mount-ornament-leaf",
}

func clampMountSpeed(v float64) float64 {
	if v < 1 {
		return 1
	}
	if v > MaxMountMult {
		return MaxMountMult
	}
	return v
}

func mountZoneAllowed(p *Player) bool {
	if p.InstanceID != "" {
		return false
	}
	z := zoneAt(p.X, p.Z)
	if z.ID == "masjid" || z.CombatDisabled {
		return false
	}
	return true
}

func (p *Player) ownsMount(id string) bool {
	for _, m := range p.ensureLog().Mounts {
		if m == id {
			return true
		}
	}
	return false
}

func (w *WorldState) grantMount(p *Player, mountID string) bool {
	if mountByID[mountID].ID == "" {
		return false
	}
	log := p.ensureLog()
	if p.ownsMount(mountID) {
		return false
	}
	log.Mounts = append(log.Mounts, mountID)
	if log.Flags == nil {
		log.Flags = map[string]bool{}
	}
	log.Flags["mount_"+mountID+"_owned"] = true
	if len(log.Mounts) == 1 {
		log.Flags["ach_first_mount"] = true
		p.grantTitle("rider")
	}
	if len(log.Mounts) >= 4 {
		log.Flags["ach_mount_collector"] = true
	}
	p.markDirty()
	w.persist(p)
	return true
}

func (w *WorldState) afterMountQuestClaim(p *Player, def QuestDef) [][]byte {
	var out [][]byte
	switch def.ID {
	case "mtq001":
		if w.grantMount(p, "dawn-wolf") {
			log := p.ensureLog()
			log.ActiveMount = "dawn-wolf"
			out = append(out, marshal(TypeMountUpdated, MountView{
				PlayerID: p.ID, MountID: "dawn-wolf", Mounted: false, Name: "Dawn Wolf", Reason: "unlock",
			}))
		}
		p.grantTitle("rider")
	case "mtq002":
		p.ensureLog().Flags["mount_training_done"] = true
		p.grantTitle("explorer")
	}
	out = append(out, w.refreshTravelAchievements(p)...)
	return out
}

func (w *WorldState) refreshTravelAchievements(p *Player) [][]byte {
	log := p.ensureLog()
	nZone := 0
	for _, ok := range log.DiscoveredZones {
		if ok {
			nZone++
		}
	}
	nMark := 0
	for _, ok := range log.Landmarks {
		if ok {
			nMark++
		}
	}
	if nMark >= 3 {
		log.Flags["ach_explorer"] = true
		p.grantTitle("explorer")
	}
	if nZone >= 4 {
		log.Flags["ach_world_traveler"] = true
		p.grantTitle("traveler")
	}
	if p.ownsMount("celestial-bird") {
		p.grantTitle("sky-seeker")
	}
	return w.refreshAchievements(p)
}

func (w *WorldState) ApplyTravel(id string, env Envelope) [][]byte {
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	switch env.Type {
	case TypeGetMounts:
		return [][]byte{marshal(TypeMountCollection, w.mountCollection(p))}
	case TypeFavoriteMount:
		var in MountIn
		_ = unmarshal(env.Data, &in)
		return w.favoriteMount(p, in.MountID)
	case TypeEquipMount:
		var in MountIn
		_ = unmarshal(env.Data, &in)
		return w.equipMount(p, in.MountID)
	case TypeSetMountCosmetic:
		var in struct {
			MountID   string `json:"mountId"`
			CosmeticID string `json:"cosmeticId"`
			Slot      string `json:"slot"`
		}
		_ = unmarshal(env.Data, &in)
		return w.setMountCosmetic(p, in.MountID, in.Slot, in.CosmeticID)
	case TypeMountEmote:
		var in struct{ Kind string `json:"kind"` }
		_ = unmarshal(env.Data, &in)
		return w.mountEmote(p, in.Kind)
	case TypeRaceStart:
		return w.raceStart(p)
	case TypeRaceCheckpoint:
		var in struct{ LandmarkID string `json:"landmarkId"` }
		_ = unmarshal(env.Data, &in)
		return w.raceCheckpoint(p, in.LandmarkID)
	case TypeRaceFinish:
		return w.raceFinish(p)
	case TypeTravelEvent:
		var in struct {
			Action string `json:"action"`
			Kind   string `json:"kind"`
		}
		_ = unmarshal(env.Data, &in)
		return w.resolveTravelEvent(p, in.Kind, in.Action)
	case TypeInspectLandmark:
		var in FastTravelIn
		_ = unmarshal(env.Data, &in)
		return w.inspectLandmark(p, in.LandmarkID)
	case TypeUnlockMount, TypeGrantMount, TypeClaimMount, TypeSetMountSpeed:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) mountCollection(p *Player) map[string]any {
	log := p.ensureLog()
	owned := map[string]bool{}
	for _, id := range log.Mounts {
		owned[id] = true
	}
	list := make([]map[string]any, 0, len(mountCatalog))
	for _, m := range mountCatalog {
		entry := map[string]any{
			"mountId": m.ID, "name": m.Name, "type": m.Type, "rarity": m.Rarity,
			"speed": clampMountSpeed(m.Speed), "source": m.Source, "terrain": m.Terrain,
			"appearance": m.Appearance, "animation": m.Animation,
			"canFly": m.CanFly, "canSwim": m.CanSwim, "canGlide": m.CanGlide,
			"owned": owned[m.ID], "favorite": log.MountFavorites[m.ID],
			"status": "LOCKED",
		}
		if owned[m.ID] {
			entry["status"] = "OWNED"
			entry["name"] = m.Name
		} else {
			entry["name"] = "???"
		}
		if log.ActiveMount == m.ID {
			entry["status"] = "ACTIVE"
			entry["name"] = m.Name
		}
		if log.MountCosmetics != nil {
			entry["cosmetic"] = log.MountCosmetics[m.ID]
		}
		list = append(list, entry)
	}
	return map[string]any{
		"toId": p.ID, "active": log.ActiveMount, "favorite": log.MountFavorite,
		"mounted": p.Mounted, "mountId": p.MountID, "state": p.MountState,
		"mounts": list, "canFlyArch": true, "canSwimArch": true, "canGlideArch": true, "canClimbArch": true,
	}
}

func (w *WorldState) favoriteMount(p *Player, id string) [][]byte {
	if !p.ownsMount(id) {
		return rejectFor(p.ID, TypeFavoriteMount, "owned")
	}
	log := p.ensureLog()
	if log.MountFavorites == nil {
		log.MountFavorites = map[string]bool{}
	}
	log.MountFavorites[id] = true
	log.MountFavorite = id
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeMountCollection, w.mountCollection(p))}
}

func (w *WorldState) equipMount(p *Player, id string) [][]byte {
	if !p.ownsMount(id) {
		return rejectFor(p.ID, TypeEquipMount, "owned")
	}
	if p.Mounted {
		w.dismount(p, "swap")
	}
	p.ensureLog().ActiveMount = id
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeMountCollection, w.mountCollection(p))}
}

func (w *WorldState) setMountCosmetic(p *Player, mountID, slot, cosmeticID string) [][]byte {
	if !p.ownsMount(mountID) {
		return rejectFor(p.ID, TypeSetMountCosmetic, "owned")
	}
	log := p.ensureLog()
	ownedCos := false
	for _, c := range log.Cosmetics {
		if c == cosmeticID {
			ownedCos = true
		}
	}
	if cosmeticID != "" && !ownedCos {
		return rejectFor(p.ID, TypeSetMountCosmetic, "cosmetic")
	}
	if log.MountCosmetics == nil {
		log.MountCosmetics = map[string]string{}
	}
	if slot == "" {
		slot = "saddle"
	}
	log.MountCosmetics[mountID] = slot + ":" + cosmeticID
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeMountCollection, w.mountCollection(p))}
}

func (w *WorldState) mountEmote(p *Player, kind string) [][]byte {
	if !p.Mounted {
		return rejectFor(p.ID, TypeMountEmote, "mounted")
	}
	if kind == "" {
		kind = "pose"
	}
	return [][]byte{marshal(TypeMountEmote, map[string]any{
		"playerId": p.ID, "mountId": p.MountID, "kind": kind, "toId": p.ID,
	})}
}

func (w *WorldState) raceStart(p *Player) [][]byte {
	if p.InstanceID != "" || time.Now().Before(p.InCombatUntil) {
		return rejectFor(p.ID, TypeRaceStart, "restricted")
	}
	lm := landmarkByID[dawnRace.Checkpoints[0]]
	if lm.ID == "" || !nearby(p.X, p.Z, lm.X, lm.Z, 4) {
		return rejectFor(p.ID, TypeRaceStart, "start")
	}
	p.RaceID = dawnRace.ID
	p.RaceStep = 0
	p.RaceStartedAt = time.Now().Unix()
	p.credit("DISCOVER", dawnRace.Checkpoints[0], 1)
	return [][]byte{marshal(TypeRaceUpdated, map[string]any{
		"raceId": dawnRace.ID, "step": 0, "toId": p.ID, "state": "RUNNING",
	})}
}

func (w *WorldState) raceCheckpoint(p *Player, landmarkID string) [][]byte {
	if p.RaceID != dawnRace.ID {
		return rejectFor(p.ID, TypeRaceCheckpoint, "race")
	}
	next := p.RaceStep + 1
	if next >= len(dawnRace.Checkpoints) {
		return rejectFor(p.ID, TypeRaceCheckpoint, "done")
	}
	want := dawnRace.Checkpoints[next]
	if landmarkID != want {
		return rejectFor(p.ID, TypeRaceCheckpoint, "order")
	}
	lm := landmarkByID[want]
	if lm.ID == "" || !nearby(p.X, p.Z, lm.X, lm.Z, 4.5) {
		return rejectFor(p.ID, TypeRaceCheckpoint, "distance")
	}
	if time.Now().Unix()-p.RaceStartedAt > RaceMaxSec {
		p.RaceID = ""
		return rejectFor(p.ID, TypeRaceCheckpoint, "time")
	}
	p.RaceStep = next
	p.credit("DISCOVER", want, 1)
	w.discoverLandmark(p, lm)
	return [][]byte{marshal(TypeRaceUpdated, map[string]any{
		"raceId": p.RaceID, "step": p.RaceStep, "checkpoint": want, "toId": p.ID,
	})}
}

func (w *WorldState) raceFinish(p *Player) [][]byte {
	if p.RaceID != dawnRace.ID || p.RaceStep != len(dawnRace.Checkpoints)-1 {
		return rejectFor(p.ID, TypeRaceFinish, "order")
	}
	elapsed := time.Now().Unix() - p.RaceStartedAt
	if elapsed > RaceMaxSec {
		p.RaceID = ""
		return rejectFor(p.ID, TypeRaceFinish, "time")
	}
	if elapsed < 0 {
		elapsed = 0
	}
	log := p.ensureLog()
	if log.RaceClaimed == nil {
		log.RaceClaimed = map[string]bool{}
	}
	if log.RaceBest == nil {
		log.RaceBest = map[string]int64{}
	}
	dup := log.RaceClaimed[dawnRace.ID]
	events := [][]byte{}
	if !dup {
		log.RaceClaimed[dawnRace.ID] = true
		events = append(events, w.giveExp(p, dawnRace.Exp)...)
		w.giveCurrency(p, dawnRace.Coin, 0)
		p.grantCosmetic(dawnRace.Cosmetic)
		log.Flags["ach_race_champion"] = true
		p.grantTitle("rider")
	}
	if log.RaceBest[dawnRace.ID] == 0 || elapsed < log.RaceBest[dawnRace.ID] {
		log.RaceBest[dawnRace.ID] = elapsed
	}
	p.RaceID = ""
	p.markDirty()
	w.persist(p)
	events = append(events, marshal(TypeRaceUpdated, map[string]any{
		"raceId": dawnRace.ID, "state": "FINISH", "elapsed": elapsed, "duplicate": dup, "toId": p.ID,
	}))
	events = append(events, w.refreshTravelAchievements(p)...)
	return events
}

func (w *WorldState) resolveTravelEvent(p *Player, kind, action string) [][]byte {
	log := p.ensureLog()
	if log.TravelEvent == "" {
		return rejectFor(p.ID, TypeTravelEvent, "none")
	}
	if kind == "" {
		kind = log.TravelEvent
	}
	log.TravelEvent = ""
	if action == "help" {
		p.bumpRegionRep("village", 3)
		log.HelpScore++
		p.markDirty()
		w.persist(p)
		return [][]byte{marshal(TypeTravelEvent, map[string]any{
			"kind": kind, "action": "help", "toast": "Reputation +3", "toId": p.ID,
		})}
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeTravelEvent, map[string]any{
		"kind": kind, "action": "ignore", "toast": "Kamu melewatinya.", "toId": p.ID,
	})}
}

func (w *WorldState) inspectLandmark(p *Player, id string) [][]byte {
	lm, ok := landmarkByID[id]
	if !ok {
		return rejectFor(p.ID, TypeInspectLandmark, "landmark")
	}
	if !nearby(p.X, p.Z, lm.X, lm.Z, 3.6) {
		return rejectFor(p.ID, TypeInspectLandmark, "distance")
	}
	events := w.discoverLandmark(p, lm)
	text := lm.LoreJv
	sub := lm.LoreID
	if text == "" {
		text = lm.Lore
	}
	if sub == "" {
		sub = lm.Lore
	}
	events = append(events, marshal(TypeInteractResult, InteractResult{
		Kind: "landmark", TargetID: lm.ID, Title: lm.Name, Speaker: lm.Name,
		Text: text, Subtitle: sub, Options: []DialogOption{{ID: "close", Label: "Tutup"}},
	}))
	return events
}

func (w *WorldState) handleTravelPuzzle(p *Player, obj InteractDef) [][]byte {
	qid := obj.QuestionID
	if qid == "" {
		qid = "q-add-1-4"
	}
	idx := -1
	for i, q := range questionCatalog {
		if q.ID == qid {
			idx = i
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeInteract, "question")
	}
	log := p.ensureLog()
	log.Quiz = QuizSession{QuestID: obj.ID, Index: idx, Active: true}
	p.markDirty()
	q := questionCatalog[idx]
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "travel-puzzle", TargetID: obj.ID, Title: "Educational Path", Speaker: "Hidden Path",
		Text: "1 + 4 = ?", Subtitle: "Soal kelas 1.",
		Question: &QuestionOut{ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices, ToID: p.ID},
	})}
}

func (w *WorldState) answerTravelPuzzle(p *Player, in EducationAnswerIn) [][]byte {
	log := p.ensureLog()
	if !log.Quiz.Active || log.Quiz.QuestID != "travel-puzzle" {
		return rejectFor(p.ID, TypeEducationAnswer, "no_session")
	}
	if log.Quiz.Index < 0 || log.Quiz.Index >= len(questionCatalog) {
		return rejectFor(p.ID, TypeEducationAnswer, "question")
	}
	def := questionCatalog[log.Quiz.Index]
	if def.ID != in.QuestionID {
		return rejectFor(p.ID, TypeEducationAnswer, "order")
	}
	log.Quiz.Active = false
	if in.Choice != def.Correct {
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: false, Explain: def.Explain, Retry: true, Toast: "Coba lagi.",
		})}
	}
	events := w.giveExp(p, 12)
	p.grantKnowledge(1)
	log.Flags["hidden_path_open"] = true
	log.Flags["mount_sky_fox"] = true
	w.grantMount(p, "sky-fox")
	w.discoverLandmark(p, landmarkByID["hidden-trail"])
	p.markDirty()
	w.persist(p)
	events = append(events, marshal(TypeEducationFeedback, EducationFeedback{
		Correct: true, Explain: def.Explain, Toast: "+EXP · +Knowledge Point",
	}))
	return events
}

func (w *WorldState) openTravelPath(p *Player, obj InteractDef) [][]byte {
	log := p.ensureLog()
	log.Flags["path_"+obj.ID] = true
	if obj.Kind == "lever" || obj.Kind == "pressure-plate" {
		log.Flags["hidden_path_open"] = true
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: obj.Kind, TargetID: obj.ID, Title: "Hidden Path", Speaker: "Trail",
		Text: obj.Text, Subtitle: "Jalan tersembunyi terbuka.",
		Options: []DialogOption{{ID: "close", Label: "Tutup"}},
	})}
}

func (w *WorldState) partySetWaypoint(p *Player, landmarkID string) [][]byte {
	pt := w.Parties.Of(p.ID)
	if pt == nil || pt.LeaderID != p.ID {
		return rejectFor(p.ID, TypePartySetWaypoint, "not_leader")
	}
	lm, ok := landmarkByID[landmarkID]
	if !ok {
		return rejectFor(p.ID, TypePartySetWaypoint, "landmark")
	}
	pt.WaypointID = lm.ID
	pt.WaypointX, pt.WaypointZ = lm.X, lm.Z
	return [][]byte{marshal(TypePartyUpdated, w.partyView(pt, p))}
}

func (w *WorldState) followParty(p *Player) [][]byte {
	pt := w.Parties.Of(p.ID)
	if pt == nil || pt.WaypointID == "" {
		return rejectFor(p.ID, TypeFollowParty, "waypoint")
	}
	p.FollowParty = true
	return [][]byte{marshal(TypeTravelSuggestion, map[string]any{
		"toId": p.ID, "landmarkId": pt.WaypointID, "x": pt.WaypointX, "z": pt.WaypointZ,
		"follow": true, "autoTeleport": false,
		"text": "FOLLOW PARTY — pilih perjalanan sendiri.",
	})}
}

func (w *WorldState) afterTravelEventClaim(p *Player, eventID string) [][]byte {
	if eventID != "mount-festival" {
		return nil
	}
	log := p.ensureLog()
	log.FestivalTokens++
	p.grantCosmetic("festival-token")
	p.grantCosmetic("mount-festival-saddle")
	p.grantCosmetic("emote-mount-pose")
	if w.grantMount(p, "celestial-bird") {
		log.Flags["mount_celestial_bird"] = true
		p.grantTitle("sky-seeker")
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeEventReward, map[string]any{
		"eventId": eventID, "cosmetic": true, "token": 1, "toId": p.ID,
	})}
}

func waterAt(x, z float64) *waterVolume {
	for i := range waterVolumes {
		v := &waterVolumes[i]
		if math.Hypot(x-v.X, z-v.Z) <= v.R {
			return v
		}
	}
	return nil
}

func climbAt(x, z float64) *climbPoint {
	for i := range climbPoints {
		c := &climbPoints[i]
		if math.Hypot(x-c.X, z-c.Z) <= 2.4 {
			return c
		}
	}
	return nil
}

func (p *Player) tickSwim(dt float64) {
	v := waterAt(p.X, p.Z)
	if v == nil {
		p.Swimming = false
		return
	}
	p.Swimming = true
	if p.Mounted {
		p.Mounted = false
		p.MountState = MountAnimIdle
	}
	p.Stamina = math.Max(0, p.Stamina-SwimStaminaDrain*dt)
	if p.Stamina <= 0.05 {
		p.X, p.Y, p.Z = v.ShoreX, 0, v.ShoreZ
		p.VX, p.VZ = 0, 0
		p.Stamina = 12
		p.Swimming = false
	}
}

func (w *WorldState) tickTravel(p *Player) [][]byte {
	if p.Mounted {
		if math.Hypot(p.VX, p.VZ) > 0.35 {
			p.MountState = MountAnimMoving
		} else if p.MountState != MountAnimSummon {
			p.MountState = MountAnimMounted
		}
	}
	z := zoneAt(p.X, p.Z)
	if p.Mounted && !mountZoneAllowed(p) {
		return w.dismount(p, "zone")
	}
	log := p.ensureLog()
	if z.ID == "forest" && log.Landmarks["great-oak"] {
		log.Flags["mount_forest_deer"] = true
	}
	if z.ID == "canyon" || z.ID == "valley" {
		if log.Landmarks["stone-gate"] || log.Landmarks["stone-bridge"] {
			log.Flags["mount_stone_beast"] = true
		}
	}
	_ = climbAt(p.X, p.Z)
	return nil
}

func (w *WorldState) maybeAmbush(p *Player) [][]byte {
	if w.Time.Phase != "NIGHT" {
		return nil
	}
	z := zoneAt(p.X, p.Z)
	if z.SafeZone || z.ID == "village" || z.ID == "masjid" {
		return nil
	}
	var def EnemyDef
	for _, d := range enemyCatalog {
		if d.ID == "forest_fang" {
			def = d
			break
		}
	}
	if def.ID == "" {
		return nil
	}
	e := spawnEnemy(def, p.X+3, p.Z+2)
	w.enemies[e.ID] = e
	return [][]byte{marshal(TypeEnemySpawn, e.Snap())}
}
