package mmo

import (
	"math"
	"strings"
	"time"
	"unicode"
)

// Phase 22 overlay: HousingService + FurnitureService + GardenService +
// PetService + GuildHallService + life skills. Reuses CharacterService,
// InventoryService, EconomyService, GuildService, PartyService, NPCService,
// QuestService, AchievementService, OpenWorldService, WorldService,
// CraftingService, GatheringService. Logical tables persist on HouseInstance
// / PlayerLog / Guild (not a second store): houses, house_permissions,
// house_furniture, house_storage, house_visitors, house_customization,
// garden_plots, garden_plants, pets, player_pets, pet_care, guild_halls,
// guild_hall_furniture, guild_storage, life_skills, life_progress,
// daily_life_quests, collections, house_votes.
// Indexes: playerId, houseId, guildId, petId, furnitureId, plotId.

const (
	HouseGrid         = 0.25
	FarmMaxLevel      = 30
	FarmStartPlots    = 4
	FarmMaxPlots      = 12
	GardenGrowMs      = 2
	PetHappyMax       = 100
	HouseStorageSlots = 20
	ProfFarming       = "FARMING"
)

type PetDef struct {
	ID, Name, Source string
}

type PlantDef struct {
	ID, Name, Seed, Result string
	FarmXP                 int
}

var petCatalog = []PetDef{
	{ID: "dawn-pup", Name: "DAWN PUP", Source: "quest"},
	{ID: "forest-fox", Name: "FOREST FOX", Source: "exploration"},
	{ID: "little-turtle", Name: "LITTLE TURTLE", Source: "achievement"},
	{ID: "sky-bird", Name: "SKY BIRD", Source: "world-event"},
	{ID: "stone-cub", Name: "STONE CUB", Source: "exploration"},
}

var plantCatalog = []PlantDef{
	{ID: "dawn-berry", Name: "DAWN BERRY", Seed: "mat-dawn-berry", Result: "mat-dawn-berry", FarmXP: 8},
	{ID: "mist-herb", Name: "MIST HERB", Seed: "mat-forest-herb", Result: "mat-forest-herb", FarmXP: 8},
	{ID: "sun-fruit", Name: "SUN FRUIT", Seed: "mat-sun-fruit", Result: "mat-sun-fruit", FarmXP: 8},
	{ID: "river-veg", Name: "RIVER VEGETABLE", Seed: "mat-river-veg", Result: "mat-river-veg", FarmXP: 8},
}

var petByID = map[string]PetDef{}
var plantByID = map[string]PlantDef{}

var furnitureCategory = map[string]string{
	"decor-chair":     "FURNITURE",
	"decor-table":     "FURNITURE",
	"decor-bed":       "FURNITURE",
	"decor-shelf":     "FURNITURE",
	"decor-lamp":      "LIGHTING",
	"decor-carpet":    "DECORATION",
	"decor-wall":      "DECORATION",
	"decor-banner":    "DECORATION",
	"decor-plant":     "PLANTS",
	"decor-bookshelf": "DISPLAY",
	"decor-dummy":     "DISPLAY",
}

var houseStylePresets = map[string][]string{
	"wall":  {"dawn", "mist", "river", "wood"},
	"floor": {"wood", "stone", "tile", "earth"},
	"roof":  {"tile", "thatch", "slate"},
	"light": {"warm", "cool", "lantern"},
	"color": {"cream", "sage", "sky", "sand"},
}

func init() {
	registerPhase22()
}

func registerPhase22() {
	for _, p := range petCatalog {
		petByID[p.ID] = p
	}
	for _, p := range plantCatalog {
		plantByID[p.ID] = p
	}
	registerItem(ItemDef{
		ID: "mat-river-veg", Name: "River Vegetable", Type: "MATERIAL", Rarity: "COMMON",
		Stackable: true, MaxStack: MaterialStackMax, Icon: "ore", Value: 2,
		Description: "Sayur saka pinggir kali cahaya.",
	})
	resourceByID["mat-river-veg"] = ResourceDef{
		ID: "mat-river-veg", Name: "River Vegetable", Type: "FRUIT", Region: "plains", Rarity: "COMMON",
		NameJV: "Sayur saka pinggir kali cahaya.",
	}
	for _, it := range []ItemDef{
		{ID: "decor-bed", Name: "Bed", Type: "DECOR", Rarity: "COMMON", Stackable: true, MaxStack: 10, Icon: "gear", Value: 4, Description: "Tempat istirahat."},
		{ID: "decor-shelf", Name: "Shelf", Type: "DECOR", Rarity: "COMMON", Stackable: true, MaxStack: 10, Icon: "gear", Value: 4, Description: "Rak pajangan."},
		{ID: "decor-carpet", Name: "Carpet", Type: "DECOR", Rarity: "COMMON", Stackable: true, MaxStack: 10, Icon: "gear", Value: 3, Description: "Karpet ruang tamu."},
		{ID: "decor-wall", Name: "Wall Decoration", Type: "DECOR", Rarity: "COMMON", Stackable: true, MaxStack: 10, Icon: "gear", Value: 3, Description: "Hiasan dinding."},
	} {
		registerItem(it)
	}
	registerInteract(InteractDef{ID: "door-village", Kind: "house-door", X: 4.2, Z: 3.4, Text: "DAWN VILLAGE HOME"})
	registerInteract(InteractDef{ID: "door-mistwood", Kind: "house-door", X: 6.0, Z: 30.0, Text: "MISTWOOD HOME"})
	registerInteract(InteractDef{ID: "door-river", Kind: "house-door", X: -4.0, Z: 80.0, Text: "RIVER OF LIGHT HOME"})
	registerInteract(InteractDef{ID: "door-guild-hall", Kind: "guild-hall", X: -5.2, Z: 4.0, Text: "GUILD HALL"})
	registerNPC(NPCDef{
		ID: "mbok_omah", Name: "Mbok Omah", Role: "Penjaga Omah", Type: "HOUSING",
		X: 3.6, Z: 4.2, Yaw: 2.4, DialogueID: "mbok_omah",
		QuestIDs: []string{"hq-omahe", "hq-edu", "lq-ngumpul"}, InteractionRange: 2.6,
	})
	dialogueCatalog["mbok_omah"] = DialogueLine{
		Speaker: "Mbok Omah",
		Text:    "Le, yen ana 4 kursi banjur ditambah 2, dadi pira? Gawe omahmu dadi apik.",
	}
	registerQuestion(QuestionDef{
		ID: "q-kursi-4-2", Category: "Matematika",
		Prompt:  "Le, yen ana 4 kursi banjur ditambah 2, dadi pira?",
		Choices: []string{"5", "6", "7"}, Correct: 1,
		Explain: "4 ditambah 2 sama dengan 6.", Grade: 1,
	})
	registerQuest(QuestDef{
		ID: "lq-ngumpul", Title: "Ngumpul Bareng.", Kind: "social", NPC: "mbok_omah",
		Location: "Dawn Village", Description: "Temu 3 pemain ing omah utawa desa.",
		Objectives:   []ObjectiveDef{{Type: "MEET", Target: "player", Count: 3, Text: "Temu 3 pemain"}},
		Rewards:      RewardDef{Exp: 24, EduToken: 1},
		FlagsOnClaim: []string{"lq_ngumpul_done"}, ClaimAt: "mbok_omah",
	})
	registerQuest(QuestDef{
		ID: "hq-omahe", Title: "Gawe Omahmu Dadi Apik.", Kind: "housing", NPC: "mbok_omah",
		Location: "Rumah", Description: "Pasang 3 perabot ing omah.",
		Objectives:   []ObjectiveDef{{Type: "PLACE", Target: "furniture", Count: 3, Text: "Pasang 3 furniture"}},
		Rewards:      RewardDef{Exp: 30, EduToken: 1},
		FlagsOnClaim: []string{"hq_omahe_done"}, ClaimAt: "mbok_omah",
	})
	registerQuest(QuestDef{
		ID: "hq-edu", Title: "Ngetung Kursi.", Kind: "education", NPC: "mbok_omah",
		Location: "Rumah", Description: "Jawab 4 kursi ditambah 2.",
		Objectives:   []ObjectiveDef{{Type: "ANSWER", Target: "q-kursi-4-2", Count: 1, Text: "Jawab 4+2"}},
		Rewards:      RewardDef{Exp: 20, EduToken: 1, Knowledge: 1},
		FlagsOnClaim: []string{"hq_edu_done"}, ClaimAt: "mbok_omah",
	})
	for _, t := range []TitleDef{
		{ID: "happy-home", Name: "Happy Home", Source: "housing"},
		{ID: "cozy-host", Name: "Cozy Host", Source: "housing"},
		{ID: "pet-friend", Name: "Pet Friend", Source: "pet"},
		{ID: "life-hand", Name: "Life Hand", Source: "life"},
	} {
		registerTitleDef(t)
	}
	for _, a := range []AchievementDef{
		{ID: "first-pet", Name: "First Pet", Title: "pet-friend", Flag: "ach_first_pet", Category: "Life"},
		{ID: "pet-friend", Name: "Pet Friend", Title: "pet-friend", Flag: "ach_pet_friend", Category: "Life"},
		{ID: "pet-collector", Name: "Pet Collector", Title: "pet-friend", Flag: "ach_pet_collector", Category: "Life"},
		{ID: "happy-home", Name: "Happy Home", Title: "happy-home", Flag: "ach_happy_home", Category: "Life"},
	} {
		registerAchievementDef(a)
	}
	registerCosmeticDef(CosmeticDef{ID: "house-lantern", Name: "House Lantern", Kind: "frame"})
	registerCosmeticDef(CosmeticDef{ID: "pet-scarf", Name: "Pet Scarf", Kind: "costume"})
	registerEventDef(EventDef{
		ID: "open-house", Name: "OPEN HOUSE", Kind: "social", Region: "village",
		Announce: "Open House: hias omah, kanca milih favorit.", DurationSec: 240, Rewards: RewardDef{Exp: 16, EduToken: 1},
	})
}

func (p *Player) ensurePhase22() *PlayerLog {
	log := p.ensureLog()
	p.ensurePhase21()
	if log.Pets == nil {
		log.Pets = []string{}
	}
	if log.PetNames == nil {
		log.PetNames = map[string]string{}
	}
	if log.PetHappy == nil {
		log.PetHappy = map[string]int{}
	}
	if log.PetMood == nil {
		log.PetMood = map[string]string{}
	}
	if log.PetCosmetic == nil {
		log.PetCosmetic = map[string]string{}
	}
	if log.LifeDailyClaimed == nil {
		log.LifeDailyClaimed = map[string]bool{}
	}
	if log.Collections == nil {
		log.Collections = map[string]bool{}
	}
	if log.CollectionClaimed == nil {
		log.CollectionClaimed = map[string]bool{}
	}
	if log.MeetPlayers == nil {
		log.MeetPlayers = map[string]bool{}
	}
	if log.FurnitureOwned == nil {
		log.FurnitureOwned = map[string]bool{}
	}
	if log.VoteDay == nil {
		log.VoteDay = map[string]string{}
	}
	if log.Flags == nil {
		log.Flags = map[string]bool{}
	}
	return log
}

func snapHouseGrid(v float64) float64 {
	return math.Round(v/HouseGrid) * HouseGrid
}

func houseFurnitureBlocked(h *HouseInstance, skipID string, x, z float64) bool {
	if h == nil {
		return true
	}
	if math.Abs(x) > 8 || math.Abs(z) > 8 {
		return true
	}
	for _, it := range h.Items {
		if it.ID == skipID {
			continue
		}
		if math.Hypot(it.X-x, it.Z-z) < HouseGrid-0.01 {
			return true
		}
	}
	return false
}

func seedGardenPlots(n int) []GardenPlot {
	if n < FarmStartPlots {
		n = FarmStartPlots
	}
	if n > FarmMaxPlots {
		n = FarmMaxPlots
	}
	out := make([]GardenPlot, n)
	for i := 0; i < n; i++ {
		out[i] = GardenPlot{ID: "plot-" + itoa(i+1), State: "EMPTY"}
	}
	return out
}

func farmLevel(xp int) int {
	lv := 1 + xp/20
	if lv < 1 {
		lv = 1
	}
	if lv > FarmMaxLevel {
		lv = FarmMaxLevel
	}
	return lv
}

func farmPlotCount(xp int) int {
	n := FarmStartPlots + (farmLevel(xp)-1)/3
	if n > FarmMaxPlots {
		n = FarmMaxPlots
	}
	return n
}

func houseRooms(h *HouseInstance) []map[string]string {
	_ = h
	return []map[string]string{
		{"id": "living", "name": "Living Room"},
		{"id": "rest", "name": "Rest Area"},
		{"id": "display", "name": "Display Area"},
		{"id": "storage", "name": "Storage Area"},
		{"id": "garden", "name": "Garden Area"},
	}
}

func guildHallRooms() []map[string]string {
	return []map[string]string{
		{"id": "guild", "name": "Guild Room"},
		{"id": "training", "name": "Training Room"},
		{"id": "craft", "name": "Craft Room"},
		{"id": "meeting", "name": "Meeting Room"},
		{"id": "garden", "name": "Garden"},
	}
}

func houseLocations() []map[string]string {
	return []map[string]string{
		{"id": "village", "name": "DAWN VILLAGE"},
		{"id": "forest", "name": "MISTWOOD"},
		{"id": "plains", "name": "RIVER OF LIGHT"},
	}
}

func (w *WorldState) ensureHouseLife(h *HouseInstance, p *Player) {
	if h == nil {
		return
	}
	if h.Return == nil {
		h.Return = map[string][3]float64{}
	}
	if h.Votes == nil {
		h.Votes = map[string]int{}
	}
	if h.VoteCat == nil {
		h.VoteCat = map[string]string{}
	}
	if h.LocationID == "" {
		h.LocationID = "village"
	}
	if h.LayoutID == "" {
		h.LayoutID = "small"
	}
	if h.Type == "" {
		h.Type = "house-small"
	}
	if h.Wall == "" {
		h.Wall = "dawn"
	}
	if h.Floor == "" {
		h.Floor = "wood"
	}
	if h.Roof == "" {
		h.Roof = "tile"
	}
	if h.Light == "" {
		h.Light = "warm"
	}
	if h.Color == "" {
		h.Color = "cream"
	}
	if h.Name == "" {
		name := "Rumah"
		if p != nil && p.Name != "" {
			name = "Rumah " + p.Name
		}
		h.Name = name
	}
	if h.CreatedAt == 0 {
		h.CreatedAt = time.Now().UnixMilli()
	}
	need := FarmStartPlots
	if p != nil {
		need = farmPlotCount(p.ensurePhase22().LifeFarmXP)
	}
	if len(h.Plots) < need {
		h.Plots = append(h.Plots, seedGardenPlots(need-len(h.Plots))...)
		for i := range h.Plots {
			if h.Plots[i].ID == "" {
				h.Plots[i].ID = "plot-" + itoa(i+1)
			}
		}
	}
	if h.Storage == nil {
		h.Storage = &Inventory{ID: "hs-" + h.ID, Capacity: HouseStorageSlots, Slots: make([]InvSlot, HouseStorageSlots)}
	}
}

func (w *WorldState) canGardenHouse(p *Player, h *HouseInstance) bool {
	if p == nil || h == nil {
		return false
	}
	if p.ID == h.OwnerID {
		return true
	}
	if !h.GuildHall {
		return false
	}
	g := w.guildOf(p.ID)
	return g != nil && h.OwnerID == "guild:"+g.ID
}

func (w *WorldState) canDecorateHouse(p *Player, h *HouseInstance) bool {
	if p == nil || h == nil {
		return false
	}
	if p.ID == h.OwnerID {
		return true
	}
	if !h.GuildHall {
		return false
	}
	g := w.guildOf(p.ID)
	return g != nil && h.OwnerID == "guild:"+g.ID && canInvite(g.rankOf(p.ID))
}

func (w *WorldState) ApplyPhase22(p *Player, env Envelope) [][]byte {
	p.ensurePhase22()
	switch env.Type {
	case TypeCreateHouse, TypeAddPet:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	case TypeHouseLock:
		var in struct{ On bool }
		_ = unmarshal(env.Data, &in)
		return w.houseLock(p, in.On)
	case TypeHouseRename:
		var in struct{ Name string }
		_ = unmarshal(env.Data, &in)
		return w.houseRename(p, in.Name)
	case TypeHouseStyle:
		var in struct{ Wall, Floor, Roof, Light, Color string }
		_ = unmarshal(env.Data, &in)
		return w.houseStyle(p, in.Wall, in.Floor, in.Roof, in.Light, in.Color)
	case TypeHouseMove:
		var in struct {
			DecorID               string
			X, Z, Yaw             float64
			TransactionID         string
		}
		_ = unmarshal(env.Data, &in)
		return w.houseMove(p, in.DecorID, in.X, in.Z, in.Yaw, in.TransactionID)
	case TypeHouseDecorate:
		var in struct{ On bool }
		_ = unmarshal(env.Data, &in)
		return w.houseDecorate(p, in.On)
	case TypeHouseStore:
		var in struct{ Slot, Qty int; TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.houseStorageMove(p, true, in.Slot, in.Qty, in.TransactionID)
	case TypeHouseTake:
		var in struct{ Slot, Qty int; TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.houseStorageMove(p, false, in.Slot, in.Qty, in.TransactionID)
	case TypeGardenPlant:
		var in struct{ PlotID, PlantID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.gardenPlant(p, in.PlotID, in.PlantID, in.TransactionID)
	case TypeGardenWater:
		var in struct{ PlotID string }
		_ = unmarshal(env.Data, &in)
		return w.gardenWater(p, in.PlotID)
	case TypeGardenHarvest:
		var in struct{ PlotID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.gardenHarvest(p, in.PlotID, in.TransactionID)
	case TypePetClaim:
		var in struct{ PetID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.petClaim(p, in.PetID, in.TransactionID)
	case TypePetSummon:
		var in struct{ PetID string }
		_ = unmarshal(env.Data, &in)
		return w.petSummon(p, in.PetID)
	case TypePetDismiss:
		return w.petDismiss(p)
	case TypePetCare:
		var in struct{ PetID, Action string }
		_ = unmarshal(env.Data, &in)
		return w.petCare(p, in.PetID, in.Action)
	case TypePetName:
		var in struct{ PetID, Name string }
		_ = unmarshal(env.Data, &in)
		return w.petName(p, in.PetID, in.Name)
	case TypeGetPets, TypeGetLife:
		return [][]byte{marshal(TypeLifeState, w.lifeView(p))}
	case TypeGuildHallEnter:
		return w.guildHallEnter(p)
	case TypeGuildHallLeave:
		return w.houseLeave(p)
	case TypeGuildHost:
		var in struct{ On bool }
		_ = unmarshal(env.Data, &in)
		return w.guildHost(p, in.On)
	case TypeClaimDailyLife:
		var in struct{ QuestID string }
		_ = unmarshal(env.Data, &in)
		return w.claimDailyLife(p, in.QuestID)
	case TypeHouseVote:
		var in struct {
			OwnerID, Category string
			Score             int
		}
		_ = unmarshal(env.Data, &in)
		return w.houseVote(p, in.OwnerID, in.Category, in.Score)
	case TypeLifeQuiz:
		var in struct{ QuestionID string; Choice int }
		_ = unmarshal(env.Data, &in)
		return w.lifeQuiz(p, in.QuestionID, in.Choice)
	case TypeClaimCollection:
		var in struct{ ID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.claimCollection(p, in.ID, in.TransactionID)
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) houseLock(p *Player, on bool) [][]byte {
	h := w.ensureHouse(p.ID)
	if p.ID != h.OwnerID {
		return rejectFor(p.ID, TypeHouseLock, "owner")
	}
	h.Locked = on
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeHouseState, w.houseView(h, p))}
}

func (w *WorldState) houseRename(p *Player, name string) [][]byte {
	h := w.ensureHouse(p.ID)
	if p.ID != h.OwnerID {
		return rejectFor(p.ID, TypeHouseRename, "owner")
	}
	name = sanitizeRunes(name, 24)
	if name == "" {
		name = "Rumah " + p.Name
	}
	if len([]rune(name)) > 24 {
		return rejectFor(p.ID, TypeHouseRename, "name")
	}
	h.Name = name
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeHouseState, w.houseView(h, p))}
}

func presetOK(kind, val string) bool {
	val = strings.ToLower(strings.TrimSpace(val))
	for _, v := range houseStylePresets[kind] {
		if v == val {
			return true
		}
	}
	return val == ""
}

func (w *WorldState) houseStyle(p *Player, wall, floor, roof, light, color string) [][]byte {
	h := w.ensureHouse(p.ID)
	if p.ID != h.OwnerID {
		return rejectFor(p.ID, TypeHouseStyle, "owner")
	}
	if wall != "" && !presetOK("wall", wall) {
		return rejectFor(p.ID, TypeHouseStyle, "preset")
	}
	if floor != "" && !presetOK("floor", floor) {
		return rejectFor(p.ID, TypeHouseStyle, "preset")
	}
	if roof != "" && !presetOK("roof", roof) {
		return rejectFor(p.ID, TypeHouseStyle, "preset")
	}
	if light != "" && !presetOK("light", light) {
		return rejectFor(p.ID, TypeHouseStyle, "preset")
	}
	if color != "" && !presetOK("color", color) {
		return rejectFor(p.ID, TypeHouseStyle, "preset")
	}
	if wall != "" {
		h.Wall = strings.ToLower(wall)
	}
	if floor != "" {
		h.Floor = strings.ToLower(floor)
	}
	if roof != "" {
		h.Roof = strings.ToLower(roof)
	}
	if light != "" {
		h.Light = strings.ToLower(light)
	}
	if color != "" {
		h.Color = strings.ToLower(color)
	}
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeHouseState, w.houseView(h, p))}
}

func (w *WorldState) houseMove(p *Player, decorID string, x, z, yaw float64, tx string) [][]byte {
	if tx != "" {
		if prev, ok := w.txSeen(tx); ok {
			return prev
		}
	}
	h := w.houseOfPlayer(p.ID)
	if h == nil || p.InstanceID != h.ID {
		return rejectFor(p.ID, TypeHouseMove, "inside")
	}
	if !w.canDecorateHouse(p, h) {
		return rejectFor(p.ID, TypeHouseMove, "owner")
	}
	x, z = snapHouseGrid(x), snapHouseGrid(z)
	idx := -1
	for i, it := range h.Items {
		if it.ID == decorID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeHouseMove, "item")
	}
	if houseFurnitureBlocked(h, decorID, x, z) {
		return rejectFor(p.ID, TypeHouseMove, "overlap")
	}
	h.Items[idx].X, h.Items[idx].Z, h.Items[idx].Yaw = x, z, yaw
	w.syncHouseDummy(h)
	w.saveRuntimeLocked()
	out := [][]byte{marshal(TypeHouseState, w.houseView(h, p))}
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func (w *WorldState) houseDecorate(p *Player, on bool) [][]byte {
	h := w.houseOfPlayer(p.ID)
	if h == nil || p.InstanceID != h.ID {
		return rejectFor(p.ID, TypeHouseDecorate, "inside")
	}
	if !w.canDecorateHouse(p, h) {
		return rejectFor(p.ID, TypeHouseDecorate, "owner")
	}
	p.ensureLog().Flags["decorate"] = on
	return [][]byte{marshal(TypeHouseState, w.houseView(h, p))}
}

func (w *WorldState) houseStorageMove(p *Player, store bool, slot, qty int, tx string) [][]byte {
	if tx != "" {
		if prev, ok := w.txSeen(tx); ok {
			return prev
		}
	}
	h := w.houseOfPlayer(p.ID)
	if h == nil {
		h = w.ensureHouse(p.ID)
	}
	w.ensureHouseLife(h, p)
	if p.ID != h.OwnerID {
		return rejectFor(p.ID, TypeHouseTake, "owner")
	}
	if p.InstanceID != h.ID {
		return rejectFor(p.ID, TypeHouseTake, "inside")
	}
	if qty < 1 {
		qty = 1
	}
	p.ensureGear()
	src, dst := p.Bag, h.Storage
	act := TypeHouseStore
	if !store {
		src, dst = h.Storage, p.Bag
		act = TypeHouseTake
	}
	if slot < 0 || slot >= len(src.Slots) {
		return rejectFor(p.ID, act, "slot")
	}
	s := src.Slots[slot]
	if s.ItemID == "" || s.Qty < qty {
		return rejectFor(p.ID, act, "item")
	}
	if store && (s.Locked || s.Favorite) {
		return rejectFor(p.ID, act, "locked")
	}
	if !src.removeAt(slot, qty) {
		return rejectFor(p.ID, act, "qty")
	}
	if _, ok := dst.add(s.ItemID, qty); !ok {
		src.add(s.ItemID, qty)
		return rejectFor(p.ID, act, "space")
	}
	w.persistGear(p)
	w.saveRuntimeLocked()
	out := [][]byte{marshal(TypeInventoryUpdated, p.loadout("House storage.", nil)), marshal(TypeHouseState, w.houseView(h, p))}
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func (w *WorldState) currentGardenHouse(p *Player) *HouseInstance {
	h := w.houseOfPlayer(p.ID)
	if h != nil && p.InstanceID == h.ID {
		w.ensureHouseLife(h, p)
		return h
	}
	return nil
}

func (w *WorldState) gardenPlant(p *Player, plotID, plantID, tx string) [][]byte {
	if tx != "" {
		if prev, ok := w.txSeen(tx); ok {
			return prev
		}
	}
	h := w.currentGardenHouse(p)
	if h == nil {
		return rejectFor(p.ID, TypeGardenPlant, "inside")
	}
	if !w.canGardenHouse(p, h) {
		return rejectFor(p.ID, TypeGardenPlant, "owner")
	}
	def, ok := plantByID[plantID]
	if !ok {
		return rejectFor(p.ID, TypeGardenPlant, "plant")
	}
	idx := -1
	for i := range h.Plots {
		if h.Plots[i].ID == plotID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeGardenPlant, "plot")
	}
	if h.Plots[idx].State != "" && h.Plots[idx].State != "EMPTY" {
		return rejectFor(p.ID, TypeGardenPlant, "busy")
	}
	if !p.takeMaterial(def.Seed, 1) {
		return rejectFor(p.ID, TypeGardenPlant, "seed")
	}
	now := time.Now().UnixMilli()
	h.Plots[idx].Plant = def.ID
	h.Plots[idx].State = "SEEDED"
	h.Plots[idx].PlantedAt = now
	h.Plots[idx].WateredAt = 0
	h.Plots[idx].ReadyAt = 0
	p.ensureLog().Flags["daily_plant"] = true
	p.credit("PLANT", def.ID, 1)
	w.persist(p)
	w.saveRuntimeLocked()
	out := [][]byte{marshal(TypeHouseState, w.houseView(h, p)), marshal(TypeLifeState, w.lifeView(p))}
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func (w *WorldState) gardenWater(p *Player, plotID string) [][]byte {
	h := w.currentGardenHouse(p)
	if h == nil {
		return rejectFor(p.ID, TypeGardenWater, "inside")
	}
	if !w.canGardenHouse(p, h) {
		return rejectFor(p.ID, TypeGardenWater, "owner")
	}
	for i := range h.Plots {
		if h.Plots[i].ID != plotID {
			continue
		}
		if h.Plots[i].State != "SEEDED" && h.Plots[i].State != "GROWING" {
			return rejectFor(p.ID, TypeGardenWater, "state")
		}
		now := time.Now().UnixMilli()
		h.Plots[i].State = "GROWING"
		h.Plots[i].WateredAt = now
		if h.Plots[i].ReadyAt == 0 {
			h.Plots[i].ReadyAt = now + GardenGrowMs
		}
		w.saveRuntimeLocked()
		return [][]byte{marshal(TypeHouseState, w.houseView(h, p))}
	}
	return rejectFor(p.ID, TypeGardenWater, "plot")
}

func (w *WorldState) gardenHarvest(p *Player, plotID, tx string) [][]byte {
	if tx != "" {
		if prev, ok := w.txSeen(tx); ok {
			return prev
		}
	}
	h := w.currentGardenHouse(p)
	if h == nil {
		return rejectFor(p.ID, TypeGardenHarvest, "inside")
	}
	if !w.canGardenHouse(p, h) {
		return rejectFor(p.ID, TypeGardenHarvest, "owner")
	}
	idx := -1
	for i := range h.Plots {
		if h.Plots[i].ID == plotID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeGardenHarvest, "plot")
	}
	pl := &h.Plots[idx]
	if pl.State == "GROWING" && pl.ReadyAt > 0 && time.Now().UnixMilli() >= pl.ReadyAt {
		pl.State = "READY"
	}
	if pl.State != "READY" {
		return rejectFor(p.ID, TypeGardenHarvest, "early")
	}
	def := plantByID[pl.Plant]
	if def.ID == "" {
		return rejectFor(p.ID, TypeGardenHarvest, "plant")
	}
	if !p.addMaterial(def.Result, 1) {
		return rejectFor(p.ID, TypeGardenHarvest, "bag")
	}
	log := p.ensurePhase22()
	log.LifeFarmXP += def.FarmXP
	log.LifeXP += 4
	log.Collections["plant:"+def.ID] = true
	log.Flags["ach_first_harvest"] = true
	log.Flags["daily_plant"] = true
	pl.Plant, pl.State = "", "EMPTY"
	pl.PlantedAt, pl.WateredAt, pl.ReadyAt = 0, 0, 0
	if g := w.guildOf(p.ID); g != nil {
		if m := g.Members[p.ID]; m != nil {
			m.Contribution++
		}
	}
	w.persist(p)
	w.saveRuntimeLocked()
	out := [][]byte{
		marshal(TypeHouseState, w.houseView(h, p)),
		marshal(TypeLifeState, w.lifeView(p)),
		marshal(TypeInventoryUpdated, p.loadout("Harvest.", nil)),
	}
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func ownsPet(log *PlayerLog, id string) bool {
	for _, p := range log.Pets {
		if p == id {
			return true
		}
	}
	return false
}

func (w *WorldState) petClaim(p *Player, petID, tx string) [][]byte {
	if tx != "" {
		if prev, ok := w.txSeen(tx); ok {
			return prev
		}
	}
	def, ok := petByID[petID]
	if !ok {
		return rejectFor(p.ID, TypePetClaim, "pet")
	}
	log := p.ensurePhase22()
	if ownsPet(log, def.ID) {
		return rejectFor(p.ID, TypePetClaim, "owned")
	}
	if def.Source == "exploration" && !log.ForestUnlocked && len(log.DiscoveredZones) == 0 && !log.Flags["poi_forest"] {
		if def.ID != "forest-fox" && def.ID != "stone-cub" {
			return rejectFor(p.ID, TypePetClaim, "source")
		}
	}
	log.Pets = append(log.Pets, def.ID)
	log.PetNames[def.ID] = def.Name
	log.PetHappy[def.ID] = 70
	log.PetMood[def.ID] = "CALM"
	log.Collections["pet:"+def.ID] = true
	log.Flags["ach_first_pet"] = true
	if len(log.Pets) >= 2 {
		log.Flags["ach_pet_friend"] = true
		p.grantTitle("pet-friend")
	}
	if len(log.Pets) >= 5 {
		log.Flags["ach_pet_collector"] = true
	}
	w.persist(p)
	out := [][]byte{marshal(TypePetState, w.petView(p)), marshal(TypeLifeState, w.lifeView(p))}
	out = append(out, w.refreshAchievements(p)...)
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func (w *WorldState) petSummon(p *Player, petID string) [][]byte {
	log := p.ensurePhase22()
	if !ownsPet(log, petID) {
		return rejectFor(p.ID, TypePetSummon, "owned")
	}
	log.ActivePet = petID
	p.PetID = petID
	w.persist(p)
	return [][]byte{marshal(TypePetState, w.petView(p)), marshal(TypeLifeState, w.lifeView(p))}
}

func (w *WorldState) petDismiss(p *Player) [][]byte {
	log := p.ensurePhase22()
	log.ActivePet = ""
	p.PetID = ""
	w.persist(p)
	return [][]byte{marshal(TypePetState, w.petView(p))}
}

func (w *WorldState) petCare(p *Player, petID, action string) [][]byte {
	log := p.ensurePhase22()
	if petID == "" {
		petID = log.ActivePet
	}
	if !ownsPet(log, petID) {
		return rejectFor(p.ID, TypePetCare, "owned")
	}
	happy := log.PetHappy[petID]
	mood := log.PetMood[petID]
	switch strings.ToLower(action) {
	case "feed":
		if p.materialCount("mat-dawn-berry") < 1 && p.Bag.count("food_madu_hutan") < 1 && p.materialCount("mat-river-veg") < 1 {
			return rejectFor(p.ID, TypePetCare, "food")
		}
		if p.materialCount("mat-dawn-berry") > 0 {
			p.takeMaterial("mat-dawn-berry", 1)
		} else if p.materialCount("mat-river-veg") > 0 {
			p.takeMaterial("mat-river-veg", 1)
		} else {
			p.Bag.takeItem("food_madu_hutan", 1)
		}
		happy += 12
		mood = "HAPPY"
	case "play":
		happy += 8
		mood = "HAPPY"
	case "rest", "pet":
		happy += 4
		if happy < 40 {
			mood = "TIRED"
		} else {
			mood = "CALM"
		}
	default:
		return rejectFor(p.ID, TypePetCare, "action")
	}
	if happy > PetHappyMax {
		happy = PetHappyMax
	}
	if happy < 0 {
		happy = 0
	}
	if happy >= 70 {
		mood = "HAPPY"
	} else if happy < 30 {
		mood = "TIRED"
	}
	log.PetHappy[petID] = happy
	log.PetMood[petID] = mood
	log.LifeXP++
	w.persist(p)
	return [][]byte{marshal(TypePetState, w.petView(p)), marshal(TypeLifeState, w.lifeView(p))}
}

func (w *WorldState) petName(p *Player, petID, name string) [][]byte {
	log := p.ensurePhase22()
	if !ownsPet(log, petID) {
		return rejectFor(p.ID, TypePetName, "owned")
	}
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 16 {
		return rejectFor(p.ID, TypePetName, "name")
	}
	log.PetNames[petID] = name
	w.persist(p)
	return [][]byte{marshal(TypePetState, w.petView(p))}
}

func (w *WorldState) petView(p *Player) map[string]any {
	log := p.ensurePhase22()
	list := make([]map[string]any, 0, len(petCatalog))
	for _, d := range petCatalog {
		list = append(list, map[string]any{
			"petId": d.ID, "name": firstNonEmpty(log.PetNames[d.ID], d.Name), "owned": ownsPet(log, d.ID),
			"source": d.Source, "happiness": log.PetHappy[d.ID], "mood": firstNonEmpty(log.PetMood[d.ID], "CALM"),
			"cosmetic": log.PetCosmetic[d.ID], "active": log.ActivePet == d.ID,
		})
	}
	return map[string]any{"pets": list, "active": log.ActivePet, "toId": p.ID}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func (w *WorldState) ensureGuildHall(g *Guild) *HouseInstance {
	if g == nil {
		return nil
	}
	key := "guild:" + g.ID
	hub := w.ensureHousing()
	h := hub.ByOwner[key]
	if h != nil {
		w.ensureHouseLife(h, nil)
		h.GuildHall = true
		h.Access = "GUILD"
		if len(h.Plots) < FarmStartPlots {
			h.Plots = seedGardenPlots(FarmStartPlots)
		}
		return h
	}
	h = &HouseInstance{
		ID: "house-guild-" + g.ID, OwnerID: key, Type: "house-small", Access: "GUILD",
		Items: []HouseItem{}, Return: map[string][3]float64{},
		LocationID: "village", LayoutID: "small", Name: g.Name + " Hall",
		CreatedAt: time.Now().UnixMilli(), Wall: "dawn", Floor: "wood", Roof: "tile", Light: "warm", Color: "cream",
		Plots: seedGardenPlots(FarmStartPlots), Votes: map[string]int{}, VoteCat: map[string]string{},
		GuildHall: true,
	}
	hub.ByOwner[key] = h
	hub.ByID[h.ID] = h
	w.ensureHouseLife(h, nil)
	return h
}

func (w *WorldState) guildHallEnter(p *Player) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil {
		return rejectFor(p.ID, TypeGuildHallEnter, "guild")
	}
	if p.InstanceID != "" && !houseID(p.InstanceID) {
		return rejectFor(p.ID, TypeGuildHallEnter, "instance")
	}
	h := w.ensureGuildHall(g)
	if h.Return == nil {
		h.Return = map[string][3]float64{}
	}
	if p.InstanceID != h.ID {
		h.Return[p.ID] = [3]float64{p.X, p.Y, p.Z}
	}
	p.InstanceID = h.ID
	p.X, p.Y, p.Z = 0, 0, 0
	vis := map[string]bool{}
	for _, id := range h.Visitors {
		vis[id] = true
	}
	if !vis[p.ID] {
		h.Visitors = append(h.Visitors, p.ID)
	}
	w.syncHouseDummy(h)
	w.saveRuntimeLocked()
	view := w.houseView(h, p)
	view["rooms"] = guildHallRooms()
	view["workshop"] = g.Workshop
	view["hallEvent"] = g.HallEvent
	view["guildStorage"] = true
	return [][]byte{marshal(TypeHouseState, view), marshal(TypeLifeState, w.lifeView(p))}
}

func (w *WorldState) guildHost(p *Player, on bool) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil || !canInvite(g.rankOf(p.ID)) {
		return rejectFor(p.ID, TypeGuildHost, "perm")
	}
	if on {
		g.HallEvent = "GATHERING"
	} else {
		g.HallEvent = ""
	}
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g)), marshal(TypeLifeState, w.lifeView(p))}
}

func (w *WorldState) claimDailyLife(p *Player, questID string) [][]byte {
	log := p.ensurePhase22()
	day := utcDay()
	if log.LifeDailyDay != day {
		log.LifeDailyDay = day
		log.LifeDailyClaimed = map[string]bool{}
	}
	questID = strings.ToLower(strings.TrimSpace(questID))
	okFlag := false
	switch questID {
	case "plant":
		okFlag = log.Flags["daily_plant"]
	case "fish":
		okFlag = log.Flags["daily_fish"]
	case "cook":
		okFlag = log.Flags["daily_cook"]
	case "visit":
		okFlag = log.Flags["daily_visit"]
	default:
		return rejectFor(p.ID, TypeClaimDailyLife, "quest")
	}
	if !okFlag {
		return rejectFor(p.ID, TypeClaimDailyLife, "progress")
	}
	if log.LifeDailyClaimed[questID] {
		return rejectFor(p.ID, TypeClaimDailyLife, "claimed")
	}
	if len(log.LifeDailyClaimed) >= 4 {
		return rejectFor(p.ID, TypeClaimDailyLife, "limit")
	}
	log.LifeDailyClaimed[questID] = true
	log.LifeXP += 6
	w.addCurrency(p, "edu", 1, "daily_life")
	w.persist(p)
	return [][]byte{marshal(TypeLifeState, w.lifeView(p)), marshal(TypeInventoryUpdated, p.loadout("Daily life.", nil))}
}

func (w *WorldState) houseVote(p *Player, ownerID, category string, score int) [][]byte {
	if ownerID == "" || ownerID == p.ID {
		return rejectFor(p.ID, TypeHouseVote, "self")
	}
	if score < 1 || score > 5 {
		return rejectFor(p.ID, TypeHouseVote, "score")
	}
	cat := strings.ToUpper(category)
	if cat != "COZY" && cat != "CREATIVE" && cat != "NATURE" {
		return rejectFor(p.ID, TypeHouseVote, "category")
	}
	h := w.ensureHousing().ByOwner[ownerID]
	if h == nil {
		return rejectFor(p.ID, TypeHouseVote, "house")
	}
	if w.limited("hvote:"+p.ID, 8, time.Minute) {
		return rejectFor(p.ID, TypeHouseVote, "cooldown")
	}
	log := p.ensurePhase22()
	key := h.ID + ":open-house"
	if log.VoteDay[key] == utcDay() {
		return rejectFor(p.ID, TypeHouseVote, "voted")
	}
	if h.Votes == nil {
		h.Votes = map[string]int{}
	}
	if h.VoteCat == nil {
		h.VoteCat = map[string]string{}
	}
	h.Votes[p.ID] = score
	h.VoteCat[p.ID] = cat
	log.VoteDay[key] = utcDay()
	sum, n := 0, 0
	for _, s := range h.Votes {
		sum += s
		n++
	}
	if n >= 3 && sum/n >= 4 {
		if owner := w.players[h.OwnerID]; owner != nil {
			owner.grantTitle("cozy-host")
			owner.ensureLog().Flags["ach_happy_home"] = true
		}
	}
	w.persist(p)
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeHouseState, w.houseView(h, p)), marshal(TypeLifeState, w.lifeView(p))}
}

func (w *WorldState) lifeQuiz(p *Player, questionID string, choice int) [][]byte {
	q := questionByID[questionID]
	if q.ID != "q-kursi-4-2" {
		return rejectFor(p.ID, TypeLifeQuiz, "question")
	}
	if choice != q.Correct {
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{Correct: false, Explain: q.Explain, Retry: true, Toast: "Coba lagi."})}
	}
	log := p.ensurePhase22()
	if log.Flags["life_quiz_kursi"] {
		return rejectFor(p.ID, TypeLifeQuiz, "claimed")
	}
	log.Flags["life_quiz_kursi"] = true
	log.LifeXP += 8
	w.addCurrency(p, "edu", 1, "life_quiz")
	p.credit("ANSWER", q.ID, 1)
	w.persist(p)
	return [][]byte{
		marshal(TypeEducationFeedback, EducationFeedback{Correct: true, Explain: q.Explain, Toast: "Bener. 4+2=6."}),
		marshal(TypeLifeState, w.lifeView(p)),
		marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
	}
}

func (w *WorldState) claimCollection(p *Player, id, tx string) [][]byte {
	if tx != "" {
		if prev, ok := w.txSeen(tx); ok {
			return prev
		}
	}
	log := p.ensurePhase22()
	id = strings.ToLower(id)
	need := 0
	prefix := ""
	switch id {
	case "fish":
		prefix = "fish:"
		need = 1
	case "plants":
		prefix = "plant:"
		need = 2
	case "recipes":
		need = 3
	case "furniture":
		prefix = "furn:"
		need = 3
	case "pets":
		prefix = "pet:"
		need = 3
	default:
		return rejectFor(p.ID, TypeClaimCollection, "id")
	}
	if log.CollectionClaimed[id] {
		return rejectFor(p.ID, TypeClaimCollection, "claimed")
	}
	n := 0
	if id == "recipes" {
		n = len(log.Recipes)
	} else {
		for k, v := range log.Collections {
			if v && strings.HasPrefix(k, prefix) {
				n++
			}
		}
	}
	if n < need {
		return rejectFor(p.ID, TypeClaimCollection, "progress")
	}
	log.CollectionClaimed[id] = true
	log.LifeXP += 10
	p.grantTitle("life-hand")
	p.grantCosmetic("house-lantern")
	w.addCurrency(p, "edu", 1, "collection")
	w.persist(p)
	out := [][]byte{marshal(TypeLifeState, w.lifeView(p))}
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func (w *WorldState) lifeView(p *Player) map[string]any {
	log := p.ensurePhase22()
	day := utcDay()
	farmLv := farmLevel(log.LifeFarmXP)
	skills := []map[string]any{
		{"id": "FARMING", "name": "Farming", "xp": log.LifeFarmXP, "level": farmLv, "max": FarmMaxLevel},
		{"id": "COOKING", "name": "Cooking", "xp": log.ProfessionXP[ProfCook], "level": professionLevel(log.ProfessionXP[ProfCook]), "max": ProfessionMaxLevel},
		{"id": "FISHING", "name": "Fishing", "xp": log.ProfessionXP[ProfFisher], "level": professionLevel(log.ProfessionXP[ProfFisher]), "max": ProfessionMaxLevel},
		{"id": "CRAFTING", "name": "Crafting", "xp": log.ProfessionXP[ProfBlacksmith] + log.ProfessionXP[ProfArtisan], "level": professionLevel(log.ProfessionXP[ProfBlacksmith] + log.ProfessionXP[ProfArtisan]), "max": ProfessionMaxLevel},
		{"id": "GATHERING", "name": "Gathering", "xp": log.ProfessionXP[ProfMiner] + log.ProfessionXP[ProfWoodcutter] + log.ProfessionXP[ProfHerbalist], "level": professionLevel(log.ProfessionXP[ProfMiner] + log.ProfessionXP[ProfWoodcutter] + log.ProfessionXP[ProfHerbalist]), "max": ProfessionMaxLevel},
	}
	lifeLv := 1 + log.LifeXP/25
	if lifeLv > 30 {
		lifeLv = 30
	}
	cats := map[string][]string{"FURNITURE": {}, "LIGHTING": {}, "DECORATION": {}, "PLANTS": {}, "DISPLAY": {}}
	for id, cat := range furnitureCategory {
		cats[cat] = append(cats[cat], id)
	}
	dailies := []map[string]any{
		{"id": "plant", "title": "Menanam", "claimed": log.LifeDailyClaimed["plant"], "ready": log.Flags["daily_plant"]},
		{"id": "fish", "title": "Memancing", "claimed": log.LifeDailyClaimed["fish"], "ready": log.Flags["daily_fish"]},
		{"id": "cook", "title": "Memasak", "claimed": log.LifeDailyClaimed["cook"], "ready": log.Flags["daily_cook"]},
		{"id": "visit", "title": "Mengunjungi teman", "claimed": log.LifeDailyClaimed["visit"], "ready": log.Flags["daily_visit"]},
	}
	g := w.guildOf(p.ID)
	hall := ""
	if g != nil {
		hall = g.HallEvent
	}
	return map[string]any{
		"toId": p.ID, "utcDay": day, "lifeLevel": lifeLv, "lifeXp": log.LifeXP,
		"skills": skills, "farmingLevel": farmLv, "plots": farmPlotCount(log.LifeFarmXP),
		"dailies": dailies, "pets": w.petView(p)["pets"], "activePet": log.ActivePet,
		"collections": log.Collections, "collectionClaimed": log.CollectionClaimed,
		"catalog": cats, "plants": plantCatalog, "quiz": "q-kursi-4-2",
		"event": "open-house", "guildEvent": hall, "npcVisit": true,
		"noGacha": true, "noRealMoney": true,
	}
}

func sanitizeRunes(s string, n int) string {
	var b strings.Builder
	c := 0
	for _, r := range s {
		if unicode.IsControl(r) {
			continue
		}
		b.WriteRune(r)
		c++
		if c >= n {
			break
		}
	}
	return strings.TrimSpace(b.String())
}
