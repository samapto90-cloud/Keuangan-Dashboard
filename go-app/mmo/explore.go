package mmo

import (
	"log"
	"math"
	"time"

	_ "embed"
)

//go:embed data/worlds.json
var worldsJSON []byte

//go:embed data/regions.json
var regionsJSON []byte

//go:embed data/landmarks.json
var landmarksJSON []byte

//go:embed data/lore.json
var loreJSON []byte

//go:embed data/mounts.json
var mountsJSON []byte

//go:embed data/worldEvents.json
var worldEventsJSON []byte

//go:embed data/secrets.json
var secretsJSON []byte

//go:embed data/guardians.json
var guardiansJSON []byte

type WorldCfg struct {
	ID                       string  `json:"id"`
	Name                     string  `json:"name"`
	InterestRadius           float64 `json:"interestRadius"`
	MaxEnemiesCell           int     `json:"maxEnemiesPerCell"`
	CellSize                 float64 `json:"cellSize"`
	DiscoveryExp             int     `json:"discoveryExp"`
	GameMinutesPerRealMinute float64 `json:"gameMinutesPerRealMinute"`
	StartClock               string  `json:"startClock"`
}

type RegionDef struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Title            string  `json:"title"`
	Index            int     `json:"index"`
	LevelMin         int     `json:"levelMin"`
	LevelMax         int     `json:"levelMax"`
	RecommendedLevel int     `json:"recommendedLevel"`
	DangerLevel      string  `json:"dangerLevel"`
	MinZ             float64 `json:"minZ"`
	MaxZ             float64 `json:"maxZ"`
	Environment      string  `json:"environment"`
	Weather          string  `json:"weather"`
	Music            string  `json:"music"`
	RequiredFlag     string  `json:"requiredFlag"`
	Landmark         string  `json:"landmark"`
	Atmosphere       string  `json:"atmosphere"`
	SafeZone         bool    `json:"safeZone"`
	CombatDisabled   bool    `json:"combatDisabled"`
	Siluman          string  `json:"siluman"`
	EnemyTier        string  `json:"enemyTier,omitempty"`
	ResourceTier     string  `json:"resourceTier,omitempty"`
	Gateway          string  `json:"gateway,omitempty"`
	Path             string  `json:"path,omitempty"`
	TeleportPoint    string  `json:"teleportPoint,omitempty"`
}

type LandmarkDef struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Region       string  `json:"region"`
	X            float64 `json:"x"`
	Z            float64 `json:"z"`
	Kind         string  `json:"kind"`
	Rest         bool    `json:"rest"`
	Lore         string  `json:"lore"`
	LoreJv       string  `json:"loreJv"`
	LoreID       string  `json:"loreId"`
	Hidden       bool    `json:"hidden"`
	RequiredFlag string  `json:"requiredFlag"`
}

type LoreDef struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Region      string  `json:"region"`
	Text        string  `json:"text"`
	X           float64 `json:"x"`
	Z           float64 `json:"z"`
	Kind        string  `json:"kind,omitempty"`
	Personality string  `json:"personality,omitempty"`
	Mechanic    string  `json:"mechanic,omitempty"`
	EnemyName   string  `json:"enemyName,omitempty"`
}

type MountDef struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Speed         float64  `json:"speed"`
	MaxSpeed      float64  `json:"maxSpeed"`
	Rarity        string   `json:"rarity"`
	RequiredLevel int      `json:"requiredLevel"`
	RequiredQuest string   `json:"requiredQuest"`
	UnlockFlag    string   `json:"unlockFlag"`
	Appearance    string   `json:"appearance"`
	Animation     []string `json:"animation"`
	Source        string   `json:"source"`
	Terrain       string   `json:"terrain"`
	CanFly        bool     `json:"canFly"`
	CanSwim       bool     `json:"canSwim"`
	CanGlide      bool     `json:"canGlide"`
}

type EventDef struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Region      string    `json:"region"`
	Announce    string    `json:"announce"`
	DurationSec float64   `json:"durationSec"`
	GateHP      int       `json:"gateHp"`
	Enemies     []string  `json:"enemies"`
	NightOnly      bool      `json:"nightOnly"`
	QuestionID     string    `json:"questionId"`
	Rewards        RewardDef `json:"rewards"`
	Objective      string    `json:"objective,omitempty"`
	ObjectiveNeed  int       `json:"objectiveNeed,omitempty"`
	AnnounceJV     string    `json:"announceJv,omitempty"`
	X              float64   `json:"x,omitempty"`
	Z              float64   `json:"z,omitempty"`
}

type SecretDef struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Region         string  `json:"region"`
	X              float64 `json:"x"`
	Z              float64 `json:"z"`
	EntranceRadius float64 `json:"entranceRadius"`
	Reward         struct {
		Crystal int    `json:"crystal"`
		Lore    string `json:"lore"`
	} `json:"reward"`
}

type GuardianDef struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Title          string   `json:"title"`
	ChapterID      string   `json:"chapterId"`
	Region         string   `json:"region"`
	Level          int      `json:"level"`
	Element        string   `json:"element"`
	Difficulty     string   `json:"difficulty"`
	Story          string   `json:"story"`
	BossArena      string   `json:"bossArena"`
	Skills         []string `json:"skills"`
	LootTable      string   `json:"lootTable"`
	UniqueItem     string   `json:"uniqueItem"`
	UniqueItemName string   `json:"uniqueItemName"`
	Phases         int      `json:"phases"`
	X              float64  `json:"x"`
	Z              float64  `json:"z"`
	Intro          string   `json:"intro"`
	Encounter      string   `json:"encounter"`
	Defeat         string   `json:"defeat"`
	Aftermath      string   `json:"aftermath"`
	Personality    string   `json:"personality"`
	Weakness       string   `json:"weakness"`
	EducationPool  []string `json:"educationQuestionPool"`
	StoryName      string   `json:"storyName,omitempty"`
}

var (
	worldCfg        WorldCfg
	regionCatalog   []RegionDef
	regionByID      = map[string]RegionDef{}
	landmarkCatalog []LandmarkDef
	landmarkByID    = map[string]LandmarkDef{}
	loreCatalog     []LoreDef
	loreByID        = map[string]LoreDef{}
	mountCatalog    []MountDef
	mountByID       = map[string]MountDef{}
	eventCatalog    []EventDef
	eventByID       = map[string]EventDef{}
	secretCatalog   []SecretDef
	guardianCatalog []GuardianDef
	guardianByID    = map[string]GuardianDef{}
	guardianByChap  = map[string]GuardianDef{}
)

func initExplore() {
	mustJSON("worlds.json", worldsJSON, &worldCfg)
	if worldCfg.InterestRadius < 20 {
		worldCfg.InterestRadius = 52
	}
	if worldCfg.MaxEnemiesCell < 1 {
		worldCfg.MaxEnemiesCell = 6
	}
	if worldCfg.DiscoveryExp < 1 {
		worldCfg.DiscoveryExp = 20
	}
	mustJSON("regions.json", regionsJSON, &regionCatalog)
	for _, r := range regionCatalog {
		regionByID[r.ID] = r
	}
	mustJSON("landmarks.json", landmarksJSON, &landmarkCatalog)
	for _, l := range landmarkCatalog {
		landmarkByID[l.ID] = l
	}
	mustJSON("lore.json", loreJSON, &loreCatalog)
	for _, l := range loreCatalog {
		loreByID[l.ID] = l
	}
	mustJSON("mounts.json", mountsJSON, &mountCatalog)
	for _, m := range mountCatalog {
		mountByID[m.ID] = m
	}
	mustJSON("worldEvents.json", worldEventsJSON, &eventCatalog)
	for _, e := range eventCatalog {
		eventByID[e.ID] = e
	}
	mustJSON("secrets.json", secretsJSON, &secretCatalog)
	mustJSON("guardians.json", guardiansJSON, &guardianCatalog)
	order := []string{"forest", "valley", "plains", "canyon", "temple", "ruins", "celestial"}
	for i := range guardianCatalog {
		g := &guardianCatalog[i]
		idx := i * len(order) / len(guardianCatalog)
		if idx >= len(order) {
			idx = len(order) - 1
		}
		g.Region = order[idx]
		if i == len(guardianCatalog)-1 {
			g.Region = "celestial"
			g.Difficulty = "CELESTIAL"
		}
	}
	applySilumanOverlay()
	for _, g := range guardianCatalog {
		guardianByID[g.ID] = g
		guardianByChap[g.ChapterID] = g
	}
	log.Printf("mmo explore regions=%d guardians=%d landmarks=%d events=%d", len(regionCatalog), len(guardianCatalog), len(landmarkCatalog), len(eventCatalog))
}

func init() {
	initExplore()
}

func registerGuardianItems() {
	for _, g := range guardianCatalog {
		if g.UniqueItem == "" || itemByID[g.UniqueItem].ID != "" {
			continue
		}
		def := ItemDef{
			ID: g.UniqueItem, Name: g.UniqueItemName, Type: "EQUIPMENT", Slot: "ACCESSORY_1",
			Rarity: "RARE", Icon: "sigil", Value: 40, LevelRequirement: g.Level,
			Description: "Signature relic of " + g.Name + ".", Effects: ItemEffects{Defense: 1, MaxHP: 8}, Untradable: true,
		}
		itemCatalog = append(itemCatalog, def)
		itemByID[def.ID] = def
	}
}

func zoneAt(x, z float64) RegionDef {
	for i := len(regionCatalog) - 1; i >= 0; i-- {
		r := regionCatalog[i]
		if z >= r.MinZ && z < r.MaxZ && math.Abs(x) <= WorldLimit {
			return r
		}
	}
	if len(regionCatalog) > 0 {
		return regionCatalog[0]
	}
	return RegionDef{ID: "village", Name: "Village of Dawn"}
}

func (p *Player) regionUnlocked(r RegionDef) bool {
	if r.RequiredFlag == "" {
		return true
	}
	log := p.ensureLog()
	if log.Flags[r.RequiredFlag] {
		return true
	}
	if r.RequiredFlag == "forest_unlocked" {
		return log.ForestUnlocked
	}
	if r.RequiredFlag == "chapter_ch01_complete" && (log.Flags["chapter1_complete"] || log.Flags["chapter_ch01_complete"]) {
		return true
	}
	if (r.ID == "celestial" || r.ID == "masjid") && log.Flags["celestial_gate_unlocked"] {
		return true
	}
	if r.ID == "horizon" {
		return p.endgameUnlocked()
	}
	return false
}

func (p *Player) maxExploreZ() float64 {
	maxZ := GateMaxZ
	for _, r := range regionCatalog {
		if p.regionUnlocked(r) && r.MaxZ > maxZ {
			maxZ = r.MaxZ
		}
	}
	if !p.ensureLog().ForestUnlocked {
		return GateMaxZ
	}
	return maxZ
}

func (w *WorldState) tickExplore(p *Player) [][]byte {
	if p.InstanceID != "" {
		return nil
	}
	z := zoneAt(p.X, p.Z)
	p.ZoneID = z.ID
	log := p.ensureLog()
	if log.DiscoveredZones == nil {
		log.DiscoveredZones = map[string]bool{}
	}
	p.credit("VISIT", z.ID, 1)
	p.credit("DISCOVER", z.ID, 1)
	w.offerDynamicQuests(p)
	var events [][]byte
	events = append(events, w.maybeEncounter(p)...)
	if z.ID == "masjid" {
		first := !log.Flags["masjid_entered"]
		events = append(events, w.enterMasjid(p, first)...)
		if p.Mounted {
			events = append(events, w.dismount(p, "sanctuary")...)
		}
	}
	if log.DiscoveredZones[z.ID] {
		return events
	}
	log.DiscoveredZones[z.ID] = true
	log.Flags["zone_"+z.ID+"_discovered"] = true
	w.noteActivity(p, "VISIT", z.ID, 1)
	exp := worldCfg.DiscoveryExp
	p.grantKnowledge(1)
	log.ExplorationXP += exp
	nZone := 0
	for _, ok := range log.DiscoveredZones {
		if ok {
			nZone++
		}
	}
	if nZone >= 4 {
		log.Flags["ach_world_traveler"] = true
	}
	p.markDirty()
	events = append(events, w.grantExp(p, exp)...)
	toastName := z.Title
	if toastName == "" {
		toastName = z.Name
	}
	events = append([][]byte{marshal(TypeZoneDiscovered, ZoneDiscovered{
		PlayerID: p.ID, ZoneID: z.ID, Name: toastName, Exp: exp, Toast: toastName + " DISCOVERED",
	})}, events...)
	if cin := regionIntroCinematic(z.ID); cin != "" && log.PendingCinematic == "" {
		events = append(events, w.startCinematic(p, cin, false)...)
	}
	events = append(events, w.refreshAchievements(p)...)
	w.persist(p)
	return events
}

func (w *WorldState) ApplyExplore(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	switch env.Type {
	case TypeGetWorldJournal:
		return [][]byte{marshal(TypeWorldJournal, p.worldJournal(w))}
	case TypeGetOpenWorld, TypeSetBossDead, TypeSpawnBoss, TypeCompleteDungeon, TypeSkipMechanic, TypeDamageBoss:
		return w.ApplyPhase23(p, env)
	case TypeGetAdventure, TypeSetRelationship, TypeSetNpcMemory, TypeUnlockLore,
		TypeCompleteQuest, TypeSetWorldState, TypeClaimEducationReward, TypeAddRelationship:
		return w.ApplyPhase24(p, env)
	case TypeFastTravel:
		var in FastTravelIn
		_ = unmarshal(env.Data, &in)
		return w.fastTravel(p, in.LandmarkID)
	case TypeRequestMount:
		var in MountIn
		_ = unmarshal(env.Data, &in)
		return w.requestMount(p, in.MountID)
	case TypeDismount:
		return w.dismount(p, "manual")
	case TypeSkipCinematic:
		return w.skipCinematic(p)
	case TypeSetLanguage, TypeGetStory, TypeStoryChoice, TypeClaimStoryChapter,
		TypeReplayCinematic, TypeReplayChapter, TypeStartNGPlus, TypeCinematicDone,
		TypeSetStoryFlag, TypeUnlockStoryChapter, TypeDefeatSiluman, TypeClaimStoryReward:
		return w.ApplyStory(p.ID, env)
	case TypeJoinWorldEvent:
		return w.joinWorldEvent(p)
	case TypeClaimEventReward:
		var in EventClaimIn
		_ = unmarshal(env.Data, &in)
		return w.claimEventReward(p, in.EventID, in.Score)
	case TypeSwitchChannel:
		var in ChannelIn
		_ = unmarshal(env.Data, &in)
		return w.switchChannel(p, in.Channel)
	case TypeChannelList:
		return [][]byte{marshal(TypeChannelList, w.channelList(p))}
	case TypeClaimWorldBoss:
		var in WorldBossClaimIn
		_ = unmarshal(env.Data, &in)
		return w.claimWorldBoss(p, in.BossID, in.Contribution, in.TransactionID)
	case TypeGetCollections:
		return [][]byte{marshal(TypeCollectionBook, w.collectionBook(p))}
	case TypeManualSave:
		return w.manualSave(p)
	case TypeGetMounts, TypeFavoriteMount, TypeEquipMount, TypeSetMountCosmetic, TypeMountEmote,
		TypeUnlockMount, TypeGrantMount, TypeClaimMount, TypeSetMountSpeed, TypeRaceStart,
		TypeRaceCheckpoint, TypeRaceFinish, TypeTravelEvent, TypeInspectLandmark:
		return w.ApplyTravel(p.ID, env)
	case TypeUnlockGuardian, TypeUnlockRegion, TypeSetWeather, TypeTeleport, TypeTriggerWorldBoss, TypeSetWorldBossHP, TypeSetBossHP, TypeSetWorldTime, TypeSpawnTreasure, TypeStartWorldEvent:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) fastTravel(p *Player, landmarkID string) [][]byte {
	lm, ok := landmarkByID[landmarkID]
	if !ok {
		return rejectFor(p.ID, TypeFastTravel, "landmark")
	}
	if p.InstanceID != "" {
		return rejectFor(p.ID, TypeFastTravel, "dungeon")
	}
	if !p.alive() {
		return rejectFor(p.ID, TypeFastTravel, "dead")
	}
	if time.Now().Before(p.InCombatUntil) {
		return rejectFor(p.ID, TypeFastTravel, "combat")
	}
	if w.bossEncounterLocked(p) {
		return rejectFor(p.ID, TypeFastTravel, "boss")
	}
	log := p.ensureLog()
	if log.Flags["fast_travel_locked"] {
		return rejectFor(p.ID, TypeFastTravel, "story")
	}
	if !log.FastTravel[lm.ID] && !log.Landmarks[lm.ID] {
		return rejectFor(p.ID, TypeFastTravel, "undiscovered")
	}
	if lm.RequiredFlag != "" && !log.Flags[lm.RequiredFlag] {
		return rejectFor(p.ID, TypeFastTravel, "locked")
	}
	reg := regionByID[lm.Region]
	if reg.ID != "" && !p.regionUnlocked(reg) {
		return rejectFor(p.ID, TypeFastTravel, "region")
	}
	cost := fastTravelCost(p.Level)
	if cost > 0 {
		if log.Coin < cost {
			return rejectFor(p.ID, TypeFastTravel, "coin")
		}
		log.Coin -= cost
		p.markDirty()
		w.persist(p)
	}
	w.dismount(p, "travel")
	p.TravelOK = true
	p.X, p.Y, p.Z = lm.X, 0, lm.Z
	p.VX, p.VZ = 0, 0
	p.ZoneID = lm.Region
	return [][]byte{marshal(TypeFastTravelOk, map[string]any{"landmarkId": lm.ID, "x": lm.X, "z": lm.Z, "toId": p.ID, "cost": cost})}
}

func (w *WorldState) discoverLandmark(p *Player, lm LandmarkDef) [][]byte {
	log := p.ensureLog()
	if log.Landmarks == nil {
		log.Landmarks = map[string]bool{}
	}
	if log.FastTravel == nil {
		log.FastTravel = map[string]bool{}
	}
	if log.Landmarks[lm.ID] {
		return nil
	}
	log.Landmarks[lm.ID] = true
	log.FastTravel[lm.ID] = true
	p.credit("DISCOVER", lm.ID, 1)
	if lm.Kind == "poi" || lm.Kind == "shrine" || lm.Kind == "waypoint" {
		p.grantKnowledge(1)
		if log.POI == nil {
			log.POI = map[string]bool{}
		}
		log.POI[lm.ID] = true
	}
	p.markDirty()
	w.persist(p)
	if lm.Rest {
		p.HP, p.Energy, p.Stamina = p.MaxHP, p.MaxEnergy, p.MaxStamina
	}
	return [][]byte{marshal(TypeLandmarkDiscovered, map[string]any{"id": lm.ID, "name": lm.Name, "toId": p.ID})}
}

func (w *WorldState) discoverLore(p *Player, def LoreDef) [][]byte {
	log := p.ensureLog()
	if log.Lore == nil {
		log.Lore = map[string]bool{}
	}
	if log.Lore[def.ID] {
		return nil
	}
	log.Lore[def.ID] = true
	complete := true
	for _, l := range loreCatalog {
		if !log.Lore[l.ID] {
			complete = false
			break
		}
	}
	if complete {
		p.grantTitle("lore-keeper")
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeLoreDiscovered, LoreView{ID: def.ID, Title: def.Title, Text: def.Text})}
}

func (p *Player) worldJournal(w *WorldState) WorldJournal {
	log := p.ensureLog()
	guardians := make([]GuardianView, 0, len(guardianCatalog))
	defeated := 0
	for i, g := range guardianCatalog {
		st := "LOCKED"
		if log.Guardians[g.ID] || log.Flags["guardian_"+g.ID+"_defeated"] {
			st = "DEFEATED"
			defeated++
		} else if i == 0 || log.Flags["chapter_"+g.ChapterID+"_available"] || log.Flags["chapter_"+prevChapter(g.ChapterID)+"_complete"] || (g.ChapterID == "ch01") {
			st = "AVAILABLE"
		}
		if g.ChapterID == "ch01" && !log.Flags["guardian_"+g.ID+"_defeated"] {
			st = "AVAILABLE"
		}
		guardians = append(guardians, GuardianView{
			ID: g.ID, Name: g.Name, Title: g.Title, Status: st, ChapterID: g.ChapterID, Region: g.Region, Index: i + 1,
			Personality: g.Personality, Weakness: g.Weakness, Story: g.Story,
		})
		silumanViewExtra(log, g, &guardians[len(guardians)-1])
	}
	regions := make([]RegionView, 0, len(regionCatalog))
	found := 0
	for _, r := range regionCatalog {
		disc := log.DiscoveredZones[r.ID]
		if disc {
			found++
		}
		pct := 0
		if disc {
			pct = 20
		}
		if log.Landmarks[r.Landmark] {
			pct += 20
		}
		for _, lm := range landmarkCatalog {
			if lm.Region == r.ID && log.Landmarks[lm.ID] {
				pct += 5
				if pct > 80 {
					pct = 80
				}
			}
		}
		if log.Flags["secret_"+r.ID+"_found"] {
			pct += 10
		}
		if log.Flags["region_"+r.ID+"_complete"] {
			pct = 100
		}
		regions = append(regions, RegionView{ID: r.ID, Name: r.Name, Discovered: disc, Completion: pct, Unlocked: p.regionUnlocked(r)})
	}
	lores := make([]LoreView, 0)
	for _, l := range loreCatalog {
		if log.Lore[l.ID] {
			lores = append(lores, LoreView{ID: l.ID, Title: l.Title, Text: l.Text})
		}
	}
	lms := make([]LandmarkView, 0)
	for _, lm := range landmarkCatalog {
		found := log.Landmarks[lm.ID] || log.FastTravel[lm.ID]
		if lm.Hidden && !found {
			continue
		}
		lms = append(lms, LandmarkView{ID: lm.ID, Name: lm.Name, Region: lm.Region, Discovered: found, X: lm.X, Z: lm.Z})
	}
	gate := log.Flags["celestial_gate_unlocked"]
	obj := "Temukan jalan menuju Masjid Cahaya."
	if log.Flags["storyCompleted"] {
		obj = "Explorer Mode. Dunia tetap terbuka."
	} else if log.Flags["masjid_entered"] {
		obj = "Selesaikan cerita di Masjid Cahaya."
	} else if gate {
		obj = "Masuk Masjid Cahaya."
	}
	tokens := log.GuardianTokens
	achs := append([]string{}, log.Achievements...)
	ch := p.channelID()
	j := WorldJournal{
		PlayerID: p.ID, Guardians: guardians, Regions: regions, Lore: lores, Landmarks: lms,
		GuardiansDefeated: defeated, GuardiansTotal: len(guardianCatalog),
		RegionsDiscovered: found, RegionsTotal: len(regionCatalog),
		Mounts: log.Mounts, CelestialGate: gate, Objective: obj,
		Tokens: tokens, StoryCompleted: log.Flags["storyCompleted"], ExplorerMode: log.Flags["explorer_mode"] || log.Flags["storyCompleted"],
		Achievements: achs, Channel: ch,
	}
	storyJournalFields(p, &j)
	attachPhase23Journal(p, w, &j)
	attachPhase24Journal(p, w, &j)
	phase29EnrichJournal(p, w, &j)
	return j
}

func prevChapter(id string) string {
	if len(id) < 4 {
		return ""
	}
	n := 0
	for _, c := range id[2:] {
		n = n*10 + int(c-'0')
	}
	n--
	if n < 1 {
		return ""
	}
	return "ch" + pad2(n)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func (w *WorldState) skipCinematic(p *Player) [][]byte {
	log := p.ensureLog()
	log.Flags["cinematic_skipped"] = true
	cid := log.PendingCinematic
	if cid != "" {
		w.completeCinematic(p, cid, true)
	} else {
		p.markDirty()
		w.persist(p)
	}
	return [][]byte{marshal(TypeCinematicSkipped, map[string]any{"ok": true, "cinematicId": cid, "toId": p.ID})}
}

func (w *WorldState) finishCinematic(p *Player) [][]byte {
	log := p.ensureLog()
	cid := log.PendingCinematic
	if cid != "" {
		w.completeCinematic(p, cid, false)
	} else {
		p.markDirty()
		w.persist(p)
	}
	return [][]byte{marshal(TypeCinematicSkipped, map[string]any{"ok": true, "cinematicId": cid, "done": true, "toId": p.ID})}
}

func nearby(ax, az, bx, bz, r float64) bool {
	return math.Hypot(ax-bx, az-bz) <= r
}

func npcLivePos(n NPCDef, phase string) (float64, float64) {
	t := WorldTimeSystem{Phase: phase, ClockMin: 9 * 60}
	if phase == "NIGHT" {
		t.ClockMin = 21 * 60
	} else if phase == "EVENING" {
		t.ClockMin = 18 * 60
	}
	return npcDest(n, t)
}

func exploreObjects(wx, wz, radius float64, filter bool) []ObjectSnapshot {
	out := make([]ObjectSnapshot, 0, len(landmarkCatalog)+len(loreCatalog)+len(secretCatalog))
	add := func(id, kind string, x, z float64) {
		if filter && !nearby(wx, wz, x, z, radius) {
			return
		}
		out = append(out, ObjectSnapshot{ID: id, Kind: kind, X: x, Z: z, Text: id})
	}
	for _, lm := range landmarkCatalog {
		kind := "landmark"
		if lm.Rest {
			kind = "checkpoint"
		}
		if lm.Kind == "gate" {
			kind = "gate"
		}
		add(lm.ID, kind, lm.X, lm.Z)
	}
	for _, l := range loreCatalog {
		add(l.ID, "lore", l.X, l.Z)
	}
	for _, s := range secretCatalog {
		add(s.ID, "secret", s.X, s.Z)
	}
	add("chest-common", "chest", -10, 10)
	add("chest-rare", "chest", 8, 38)
	return out
}

func (w *WorldState) autoDiscoverNearby(p *Player) [][]byte {
	var events [][]byte
	for _, lm := range landmarkCatalog {
		if nearby(p.X, p.Z, lm.X, lm.Z, 3.4) {
			events = append(events, w.discoverLandmark(p, lm)...)
		}
	}
	for _, l := range loreCatalog {
		if nearby(p.X, p.Z, l.X, l.Z, 2.4) {
			events = append(events, w.discoverLore(p, l)...)
		}
	}
	events = append(events, w.checkRegionComplete(p)...)
	return events
}

func (w *WorldState) trySecret(p *Player, targetID string) [][]byte {
	for _, s := range secretCatalog {
		if s.ID != targetID {
			continue
		}
		r := s.EntranceRadius
		if r < 1 {
			r = 2.2
		}
		if math.Hypot(p.X-s.X, p.Z-s.Z) > r+0.6 {
			return rejectFor(p.ID, TypeInteract, "distance")
		}
		log := p.ensureLog()
		if log.Flags["secret_"+s.ID] {
			return [][]byte{marshal(TypeInteractResult, InteractResult{
				Kind: "secret", TargetID: s.ID, Title: s.Name, Speaker: s.Name,
				Text: "Gua ini sudah kau jelajahi.", Options: []DialogOption{{ID: "close", Label: "Tutup"}},
			})}
		}
		log.Flags["secret_"+s.ID] = true
		log.Flags["secret_"+s.Region+"_found"] = true
		p.grantTitle("secret-walker")
		p.grantCosmetic("cloak-mistwood")
		if s.Reward.Crystal > 0 {
			w.giveCurrency(p, 0, s.Reward.Crystal)
		}
		if s.Reward.Lore != "" {
			if lore, ok := loreByID[s.Reward.Lore]; ok {
				w.discoverLore(p, lore)
			}
		}
		p.markDirty()
		w.persist(p)
		return [][]byte{marshal(TypeInteractResult, InteractResult{
			Kind: "secret", TargetID: s.ID, Title: "SECRET AREA", Speaker: s.Name,
			Text: "Hidden Cave. Peti dan kristal menanti di dalam.", Options: []DialogOption{{ID: "close", Label: "Tutup"}},
			Toast: "SECRET AREA",
		})}
	}
	return nil
}

func (w *WorldState) checkRegionComplete(p *Player) [][]byte {
	log := p.ensureLog()
	var events [][]byte
	for _, r := range regionCatalog {
		if !log.DiscoveredZones[r.ID] || log.Flags["region_"+r.ID+"_complete"] {
			continue
		}
		need := 0
		have := 0
		for _, lm := range landmarkCatalog {
			if lm.Region != r.ID {
				continue
			}
			need++
			if log.Landmarks[lm.ID] || log.FastTravel[lm.ID] {
				have++
			}
		}
		if need > 0 && have >= need {
			log.Flags["region_"+r.ID+"_complete"] = true
			p.markDirty()
			events = append(events, w.grantExp(p, 40)...)
			events = append(events, marshal(TypeZoneDiscovered, ZoneDiscovered{
				PlayerID: p.ID, ZoneID: r.ID, Name: r.Name, Exp: 40, Toast: "REGION COMPLETE",
			}))
		}
	}
	if len(events) > 0 {
		w.persist(p)
	}
	return events
}

func cellKey(x, z float64) string {
	s := worldCfg.CellSize
	if s < 8 {
		s = 16
	}
	return itoa(int(math.Floor(x/s))) + ":" + itoa(int(math.Floor(z/s)))
}
