package mmo

import (
	"math"
	"time"
)

// Phase 21 overlay: EconomyService + GatheringService + FishingService +
// ProfessionService + CraftingService + TradeService + stall architecture.
// Reuses InventoryService, EquipmentService, CharacterService, QuestService,
// NPCService, WorldService, OpenWorldService, AchievementService, GuildService,
// PartyService, and existing EconomyService (Coin + EduToken).
// Logical tables persist on PlayerLog / WorldState (not a second store):
// resources, resource_nodes, professions, player_professions, recipes,
// player_recipes, crafting_jobs, crafting_stations, player_materials,
// fishing_spots, fish, npc_shops, npc_shop_items, trade_sessions,
// trade_items, trade_logs, player_stalls, economy_transactions.
// Indexes: playerId, resourceId, professionId, recipeId, stationId, tradeId.

const (
	MaterialStackMax     = 99
	MaterialStackCap     = 999
	ProfessionMaxLevel   = 50
	GatherProfessionCap  = 3
	CraftProfessionCap   = 2
	ProfessionResetLimit = 2
	CraftQueueMax        = 3
	CraftFeeGold         = 1
	RepairFeeGold        = 5
	GatherRange          = 3.2
	RecipeLocked         = "LOCKED"
	RecipeDiscovered     = "DISCOVERED"
	RecipeMastered       = "MASTERED"
	QualityNormal        = "NORMAL"
	QualityFine          = "FINE"
	QualityMasterwork    = "MASTERWORK"
)

const (
	ProfMiner      = "MINER"
	ProfWoodcutter = "WOODCUTTER"
	ProfHerbalist  = "HERBALIST"
	ProfFisher     = "FISHER"
	ProfCook       = "COOK"
	ProfBlacksmith = "BLACKSMITH"
	ProfArtisan    = "ARTISAN"
)

type ResourceDef struct {
	ID, Name, Type, Region, Rarity, NameJV string
	RespawnSec, GatherSec                  int
}

type ResourceNodeDef struct {
	ID, ResourceID, Kind string
	X, Z                 float64
}

type ProfessionDef struct {
	ID, Name, Kind, Title string
}

type RecipeMat struct {
	ItemID string `json:"itemId"`
	Qty    int    `json:"qty"`
}

type RecipeDef struct {
	ID, Name, Profession, Category, Result string
	RequiredLevel, CraftSec                int
	Materials                              []RecipeMat
	Discover                               string
}

type FishSpotDef struct {
	ID, Name, Place, Reward string
	X, Z                    float64
}

type NpcShopItem struct {
	ItemID    string
	Price     int
	SellPrice int
	Currency  string
	Buy       bool
	Sell      bool
}

type NpcShopDef struct {
	ID    string
	Name  string
	Items []NpcShopItem
}

type CraftJob struct {
	ID, RecipeID, Result, Quality string
	ReadyAt                       int64
}

type StallListing struct {
	ItemID string
	Qty    int
	Price  int
}

type PlayerStall struct {
	PlayerID string
	Open     bool
	Items    []StallListing
}

type CraftOrder struct {
	ID, Requester, Crafter, RecipeID, Status string
	Reward                                   int
}

type TradeLogRow struct {
	ID, AID, BID string
	At           int64
}

var (
	resourceByID = map[string]ResourceDef{}
	nodeByID     = map[string]ResourceNodeDef{}
	recipeByID   = map[string]RecipeDef{}
	spotByID     = map[string]FishSpotDef{}
	npcShopByID  = map[string]NpcShopDef{}
	professionBy = map[string]ProfessionDef{}
)

var resourceCatalog = []ResourceDef{
	{ID: "mat-mistwood", Name: "Mistwood", Type: "WOOD", Region: "forest", Rarity: "COMMON", NameJV: "Kayu saka alas Mistwood.", RespawnSec: 8, GatherSec: 1},
	{ID: "mat-forest-herb", Name: "Forest Herb", Type: "HERB", Region: "forest", Rarity: "COMMON", NameJV: "Godhong obat saka alas.", RespawnSec: 8, GatherSec: 1},
	{ID: "mat-dawn-berry", Name: "Dawn Berry", Type: "FRUIT", Region: "forest", Rarity: "UNCOMMON", NameJV: "Woh abang saka esuk desa.", RespawnSec: 10, GatherSec: 1},
	{ID: "mat-valley-stone", Name: "Valley Stone", Type: "STONE", Region: "valley", Rarity: "COMMON", NameJV: "Watu saka Stone Valley.", RespawnSec: 8, GatherSec: 1},
	{ID: "mat-iron-ore", Name: "Iron Ore", Type: "ORE", Region: "valley", Rarity: "UNCOMMON", NameJV: "Bijih wesi saka lembah.", RespawnSec: 12, GatherSec: 1},
	{ID: "mat-river-fish", Name: "River Fish", Type: "FISH", Region: "plains", Rarity: "COMMON", NameJV: "Iwak saka kali cahaya.", RespawnSec: 10, GatherSec: 2},
	{ID: "mat-glow-herb", Name: "Glow Herb", Type: "HERB", Region: "plains", Rarity: "RARE", NameJV: "Godhong sumunar ing pinggir kali.", RespawnSec: 14, GatherSec: 1},
	{ID: "mat-red-fiber", Name: "Red Fiber", Type: "FIBER", Region: "canyon", Rarity: "COMMON", NameJV: "Serat abang saka Crimson Plains.", RespawnSec: 10, GatherSec: 1},
	{ID: "mat-sun-fruit", Name: "Sun Fruit", Type: "FRUIT", Region: "canyon", Rarity: "UNCOMMON", NameJV: "Woh srengenge saka dataran abang.", RespawnSec: 12, GatherSec: 1},
	{ID: "mat-celestial-ore", Name: "Celestial Ore", Type: "ORE", Region: "temple", Rarity: "EPIC", NameJV: "Bijih langit saka gunung suci.", RespawnSec: 18, GatherSec: 2},
	{ID: "mat-mountain-herb", Name: "Mountain Herb", Type: "HERB", Region: "temple", Rarity: "RARE", NameJV: "Godhong gunung salju.", RespawnSec: 14, GatherSec: 1},
	{ID: "mat-ancient-stone", Name: "Ancient Stone", Type: "STONE", Region: "ruins", Rarity: "RARE", NameJV: "Watu kuna saka reruntuhan.", RespawnSec: 16, GatherSec: 1},
	{ID: "mat-relic-fragment", Name: "Relic Fragment", Type: "STONE", Region: "ruins", Rarity: "EPIC", NameJV: "Pecahan peninggalan kuna.", RespawnSec: 20, GatherSec: 2},
	{ID: "mat-mushroom", Name: "Forest Mushroom", Type: "FRUIT", Region: "forest", Rarity: "COMMON", NameJV: "Jamur alas sing aman dipangan.", RespawnSec: 8, GatherSec: 1},
}

var nodeCatalog = []ResourceNodeDef{
	{ID: "node-wood-1", ResourceID: "mat-mistwood", Kind: "gather-wood", X: -4, Z: 28},
	{ID: "node-herb-1", ResourceID: "mat-forest-herb", Kind: "gather-herb", X: 5, Z: 30},
	{ID: "node-berry-1", ResourceID: "mat-dawn-berry", Kind: "gather-fruit", X: 7, Z: 26},
	{ID: "node-shroom-1", ResourceID: "mat-mushroom", Kind: "gather-fruit", X: -7, Z: 32},
	{ID: "node-stone-1", ResourceID: "mat-valley-stone", Kind: "gather-stone", X: -3, Z: 58},
	{ID: "node-ore-1", ResourceID: "mat-iron-ore", Kind: "gather-ore", X: 4, Z: 60},
	{ID: "node-glow-1", ResourceID: "mat-glow-herb", Kind: "gather-herb", X: -6, Z: 80},
	{ID: "node-fiber-1", ResourceID: "mat-red-fiber", Kind: "gather-fiber", X: 3, Z: 100},
	{ID: "node-sun-1", ResourceID: "mat-sun-fruit", Kind: "gather-fruit", X: -4, Z: 102},
	{ID: "node-celest-1", ResourceID: "mat-celestial-ore", Kind: "gather-ore", X: 2, Z: 120},
	{ID: "node-mherb-1", ResourceID: "mat-mountain-herb", Kind: "gather-herb", X: -5, Z: 118},
	{ID: "node-ancient-1", ResourceID: "mat-ancient-stone", Kind: "gather-stone", X: 1, Z: 134},
	{ID: "node-relic-1", ResourceID: "mat-relic-fragment", Kind: "gather-stone", X: -3, Z: 136},
}

var stationCatalog = []InteractDef{
	{ID: "station-forge", Kind: "forge", X: 6.8, Z: 5.2, Text: "FORGE"},
	{ID: "station-bench", Kind: "workbench", X: 7.6, Z: 6.4, Text: "WORKBENCH"},
	{ID: "station-cook", Kind: "cooking-fire", X: -6.6, Z: 5.4, Text: "COOKING FIRE"},
	{ID: "spot-village", Kind: "fishing-spot", X: -8.4, Z: 8.2, Text: "Kolam desa"},
	{ID: "spot-river", Kind: "fishing-spot", X: -8, Z: 80, Text: "River of Light"},
	{ID: "spot-lake", Kind: "fishing-spot", X: 8, Z: 36, Text: "Danau Mistwood"},
	{ID: "spot-coast", Kind: "fishing-spot", X: 8, Z: 104, Text: "Pantai Crimson"},
}

var professionCatalog = []ProfessionDef{
	{ID: ProfMiner, Name: "Miner", Kind: "gather", Title: "gatherer"},
	{ID: ProfWoodcutter, Name: "Woodcutter", Kind: "gather", Title: "gatherer"},
	{ID: ProfHerbalist, Name: "Herbalist", Kind: "gather", Title: "gatherer"},
	{ID: ProfFisher, Name: "Fisher", Kind: "gather", Title: "fisher"},
	{ID: ProfCook, Name: "Cook", Kind: "craft", Title: "cook"},
	{ID: ProfBlacksmith, Name: "Blacksmith", Kind: "craft", Title: "craftsman"},
	{ID: ProfArtisan, Name: "Artisan", Kind: "craft", Title: "artisan"},
}

var recipeCatalog = []RecipeDef{
	{ID: "rec-dawn-staff", Name: "Dawn Staff", Profession: ProfBlacksmith, Category: "WEAPON", Result: "dawn_staff", RequiredLevel: 1, CraftSec: 0, Materials: []RecipeMat{{"mat-mistwood", 2}, {"mat-valley-stone", 1}}, Discover: "start"},
	{ID: "rec-wind-blade", Name: "Wind Blade", Profession: ProfBlacksmith, Category: "WEAPON", Result: "wind_blade", RequiredLevel: 3, CraftSec: 5, Materials: []RecipeMat{{"mat-iron-ore", 2}, {"mat-mistwood", 1}}, Discover: "start"},
	{ID: "rec-stone-hammer", Name: "Stone Hammer", Profession: ProfBlacksmith, Category: "WEAPON", Result: "stone_hammer", RequiredLevel: 2, CraftSec: 5, Materials: []RecipeMat{{"mat-valley-stone", 3}}, Discover: "start"},
	{ID: "rec-forest-bow", Name: "Forest Bow", Profession: ProfArtisan, Category: "WEAPON", Result: "forest_bow", RequiredLevel: 2, CraftSec: 5, Materials: []RecipeMat{{"mat-mistwood", 3}, {"mat-red-fiber", 1}}, Discover: "start"},
	{ID: "rec-dawn-helm", Name: "Dawn Helm", Profession: ProfBlacksmith, Category: "ARMOR", Result: "dawn_helm", RequiredLevel: 2, CraftSec: 0, Materials: []RecipeMat{{"mat-iron-ore", 1}, {"mat-valley-stone", 1}}, Discover: "start"},
	{ID: "rec-dawn-mail", Name: "Dawn Mail", Profession: ProfBlacksmith, Category: "ARMOR", Result: "dawn_mail", RequiredLevel: 3, CraftSec: 8, Materials: []RecipeMat{{"mat-iron-ore", 2}, {"mat-red-fiber", 1}}, Discover: "start"},
	{ID: "rec-dawn-gloves", Name: "Dawn Gloves", Profession: ProfArtisan, Category: "ARMOR", Result: "dawn_gloves", RequiredLevel: 1, CraftSec: 0, Materials: []RecipeMat{{"mat-red-fiber", 2}}, Discover: "start"},
	{ID: "rec-dawn-boots", Name: "Dawn Boots", Profession: ProfArtisan, Category: "ARMOR", Result: "dawn_boots", RequiredLevel: 1, CraftSec: 0, Materials: []RecipeMat{{"mat-red-fiber", 1}, {"mat-mistwood", 1}}, Discover: "start"},
	{ID: "rec-dawn-ring", Name: "Dawn Ring", Profession: ProfArtisan, Category: "ACCESSORY", Result: "dawn_ring", RequiredLevel: 4, CraftSec: 8, Materials: []RecipeMat{{"mat-iron-ore", 1}, {"mat-relic-fragment", 1}}, Discover: "explore"},
	{ID: "rec-dawn-amulet", Name: "Dawn Amulet", Profession: ProfArtisan, Category: "ACCESSORY", Result: "dawn_amulet", RequiredLevel: 5, CraftSec: 8, Materials: []RecipeMat{{"mat-glow-herb", 1}, {"mat-celestial-ore", 1}}, Discover: "explore"},
	{ID: "rec-dawn-bracelet", Name: "Dawn Bracelet", Profession: ProfArtisan, Category: "ACCESSORY", Result: "dawn_bracelet", RequiredLevel: 4, CraftSec: 8, Materials: []RecipeMat{{"mat-red-fiber", 1}, {"mat-ancient-stone", 1}}, Discover: "explore"},
	{ID: "rec-madu-hutan", Name: "Madu Hutan", Profession: ProfCook, Category: "FOOD", Result: "food_madu_hutan", RequiredLevel: 1, CraftSec: 0, Materials: []RecipeMat{{"mat-dawn-berry", 1}, {"mat-forest-herb", 1}}, Discover: "start"},
	{ID: "rec-sup-sayur", Name: "Sup Sayur", Profession: ProfCook, Category: "FOOD", Result: "food_sup_sayur", RequiredLevel: 1, CraftSec: 0, Materials: []RecipeMat{{"mat-forest-herb", 2}, {"mat-mushroom", 1}}, Discover: "start"},
	{ID: "rec-ikan-bakar", Name: "Ikan Bakar", Profession: ProfCook, Category: "FOOD", Result: "food_ikan_bakar", RequiredLevel: 1, CraftSec: 0, Materials: []RecipeMat{{"mat-river-fish", 1}}, Discover: "start"},
	{ID: "rec-torch", Name: "Dawn Torch", Profession: ProfArtisan, Category: "UTILITY", Result: "dawn_torch", RequiredLevel: 1, CraftSec: 0, Materials: []RecipeMat{{"mat-mistwood", 1}}, Discover: "start"},
	{ID: "rec-apron", Name: "Cook Apron", Profession: ProfCook, Category: "COSMETIC", Result: "cook_apron", RequiredLevel: 8, CraftSec: 5, Materials: []RecipeMat{{"mat-red-fiber", 2}}, Discover: "master"},
}

var fishSpotCatalog = []FishSpotDef{
	{ID: "spot-village", Name: "Kolam Desa", Place: "lake", Reward: "mat-river-fish", X: -8.4, Z: 8.2},
	{ID: "spot-river", Name: "River of Light", Place: "river", Reward: "mat-river-fish", X: -8, Z: 80},
	{ID: "spot-lake", Name: "Danau Mistwood", Place: "lake", Reward: "mat-river-fish", X: 8, Z: 36},
	{ID: "spot-coast", Name: "Pantai Crimson", Place: "coast", Reward: "mat-river-fish", X: 8, Z: 104},
}

func init() {
	registerPhase21()
}

func registerPhase21() {
	for _, r := range resourceCatalog {
		resourceByID[r.ID] = r
		registerItem(ItemDef{
			ID: r.ID, Name: r.Name, Description: r.NameJV, Type: "MATERIAL",
			Rarity: r.Rarity, Stackable: true, MaxStack: MaterialStackMax, Icon: "ore", Value: 2,
		})
	}
	for _, n := range nodeCatalog {
		nodeByID[n.ID] = n
		registerInteract(InteractDef{ID: n.ID, Kind: n.Kind, X: n.X, Z: n.Z, Text: resourceByID[n.ResourceID].Name})
	}
	for _, s := range stationCatalog {
		registerInteract(s)
	}
	for _, p := range professionCatalog {
		professionBy[p.ID] = p
	}
	for _, r := range recipeCatalog {
		recipeByID[r.ID] = r
	}
	for _, s := range fishSpotCatalog {
		spotByID[s.ID] = s
	}
	registerCraftedGear()
	registerPhase21NPCs()
	registerPhase21Quests()
	registerPhase21Meta()
	npcShopByID["mbah-karya-shop"] = NpcShopDef{
		ID: "mbah-karya-shop", Name: "Toko Mbah Karya",
		Items: []NpcShopItem{
			{ItemID: "mat-valley-stone", Price: 4, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "mat-mistwood", Price: 4, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "mat-iron-ore", Price: 8, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "mat-forest-herb", Price: 3, Currency: "gold", Buy: true, Sell: true},
		},
	}
	npcShopByID["mbok-rasa-shop"] = NpcShopDef{
		ID: "mbok-rasa-shop", Name: "Dapur Mbok Rasa",
		Items: []NpcShopItem{
			{ItemID: "mat-dawn-berry", Price: 5, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "mat-mushroom", Price: 3, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "mat-river-fish", Price: 6, Currency: "gold", Buy: true, Sell: true},
		},
	}
}

func registerItem(def ItemDef) {
	if itemByID[def.ID].ID != "" {
		return
	}
	if def.MaxStack < 1 {
		def.MaxStack = 1
	}
	itemCatalog = append(itemCatalog, def)
	itemByID[def.ID] = def
}

func registerInteract(o InteractDef) {
	if interactByID[o.ID].ID != "" {
		return
	}
	interactCatalog = append(interactCatalog, o)
	interactByID[o.ID] = o
}

func registerNPC(n NPCDef) {
	if npcByID[n.ID].ID != "" {
		return
	}
	npcCatalog = append(npcCatalog, n)
	npcByID[n.ID] = n
}

func registerQuest(q QuestDef) {
	if questByID[q.ID].ID != "" {
		return
	}
	questCatalog = append(questCatalog, q)
	questByID[q.ID] = q
}

func registerQuestion(q QuestionDef) {
	if questionByID[q.ID].ID != "" {
		return
	}
	questionCatalog = append(questionCatalog, q)
	questionByID[q.ID] = q
}

func registerTitleDef(t TitleDef) {
	if titleByID[t.ID].ID != "" {
		return
	}
	titleCatalog = append(titleCatalog, t)
	titleByID[t.ID] = t
}

func registerAchievementDef(a AchievementDef) {
	for _, e := range achievementCatalog {
		if e.ID == a.ID {
			return
		}
	}
	achievementCatalog = append(achievementCatalog, a)
}

func registerEventDef(e EventDef) {
	if eventByID[e.ID].ID != "" {
		return
	}
	eventCatalog = append(eventCatalog, e)
	eventByID[e.ID] = e
}

func registerCosmeticDef(c CosmeticDef) {
	if cosmeticByID[c.ID].ID != "" {
		return
	}
	cosmeticCat = append(cosmeticCat, c)
	cosmeticByID[c.ID] = c
}

func registerCraftedGear() {
	eq := func(id, name, slot string, atk, def, hp int) {
		registerItem(ItemDef{
			ID: id, Name: name, Type: "EQUIPMENT", Slot: slot, Rarity: "UNCOMMON",
			Icon: "gear", Value: 18, LevelRequirement: 1,
			Description: "Karya tangan Dawn Village.", Effects: ItemEffects{Attack: atk, Defense: def, MaxHP: hp},
		})
	}
	eq("dawn_staff", "Dawn Staff", "WEAPON", 3, 0, 0)
	eq("wind_blade", "Wind Blade", "WEAPON", 4, 0, 0)
	eq("stone_hammer", "Stone Hammer", "WEAPON", 3, 1, 0)
	eq("forest_bow", "Forest Bow", "WEAPON", 3, 0, 0)
	eq("dawn_helm", "Dawn Helm", "HEAD", 0, 2, 6)
	eq("dawn_mail", "Dawn Mail", "BODY", 0, 3, 10)
	eq("dawn_gloves", "Dawn Gloves", "ACCESSORY_1", 0, 1, 4)
	eq("dawn_boots", "Dawn Boots", "LEGS", 0, 1, 4)
	eq("dawn_ring", "Dawn Ring", "ACCESSORY_1", 1, 0, 4)
	eq("dawn_amulet", "Dawn Amulet", "ACCESSORY_2", 0, 1, 8)
	eq("dawn_bracelet", "Dawn Bracelet", "ACCESSORY_2", 1, 1, 4)
	registerItem(ItemDef{ID: "food_madu_hutan", Name: "Madu Hutan", Type: "CONSUMABLE", Rarity: "COMMON", Stackable: true, MaxStack: 20, Icon: "potion", Value: 6, Description: "Memulihkan HP kecil.", Effects: ItemEffects{HealPct: 0.08}})
	registerItem(ItemDef{ID: "food_sup_sayur", Name: "Sup Sayur", Type: "CONSUMABLE", Rarity: "COMMON", Stackable: true, MaxStack: 20, Icon: "potion", Value: 6, Description: "Memulihkan energi kecil.", Effects: ItemEffects{EnergyPct: 0.10}})
	registerItem(ItemDef{ID: "food_ikan_bakar", Name: "Ikan Bakar", Type: "CONSUMABLE", Rarity: "COMMON", Stackable: true, MaxStack: 20, Icon: "potion", Value: 6, Description: "Bonus stamina kecil.", Effects: ItemEffects{StaminaPct: 0.12}})
	registerItem(ItemDef{ID: "dawn_torch", Name: "Dawn Torch", Type: "UTILITY", Rarity: "COMMON", Stackable: true, MaxStack: 10, Icon: "gear", Value: 4, Description: "Obor desa. Bukan senjata."})
	registerItem(ItemDef{ID: "cook_apron", Name: "Cook Apron", Type: "COSMETIC", Rarity: "UNCOMMON", Icon: "cloak", Value: 12, Description: "Celemek dapur Mbok Rasa."})
}

func registerPhase21NPCs() {
	registerNPC(NPCDef{ID: "mbah_karya", Name: "Mbah Karya", Role: "Pandai Besi", Type: "BLACKSMITH", X: 6.2, Z: 4.6, Yaw: 3.6, DialogueID: "mbah_karya", QuestIDs: []string{"pq-mine-1"}, InteractionRange: 2.6})
	registerNPC(NPCDef{ID: "mbok_rasa", Name: "Mbok Rasa", Role: "Juru Masak", Type: "COOK", X: -6.4, Z: 4.8, Yaw: 1.2, DialogueID: "mbok_rasa", QuestIDs: []string{"pq-edu-1"}, InteractionRange: 2.6})
	registerNPC(NPCDef{ID: "pak_jala", Name: "Pak Jala", Role: "Nelayan", Type: "FISHER", X: -8.0, Z: 7.4, Yaw: 2.2, DialogueID: "pak_jala", InteractionRange: 2.6})
	registerNPC(NPCDef{ID: "mbah_batu", Name: "Mbah Batu", Role: "Penambang", Type: "MINER", X: 8.0, Z: 8.2, Yaw: 4.0, DialogueID: "mbah_batu", QuestIDs: []string{"pq-mine-1"}, InteractionRange: 2.6})
	dialogueCatalog["mbah_karya"] = DialogueLine{Speaker: "Mbah Karya", Text: "Le, nek arep ndandani gegamanmu, gawa mrene."}
	dialogueCatalog["mbok_rasa"] = DialogueLine{Speaker: "Mbok Rasa", Text: "Le, yen ngelih, masak ing kene. Aja mangan barang mbebayani."}
	dialogueCatalog["pak_jala"] = DialogueLine{Speaker: "Pak Jala", Text: "Le, iwak ana ing kali lan tlaga. Tunggu wektune, banjur tarik."}
	dialogueCatalog["mbah_batu"] = DialogueLine{Speaker: "Mbah Batu", Text: "Le, bahan iki isa digawe dadi gegaman. Coba kok gawa menyang pandhe."}
}

func registerPhase21Quests() {
	registerQuestion(QuestionDef{
		ID: "q-iwak-2-3", Category: "Matematika",
		Prompt: "Yen ana 2 iwak banjur oleh 3 maneh, dadi pira?",
		Choices: []string{"4", "5", "6"}, Correct: 1,
		Explain: "2 ditambah 3 sama dengan 5.", Grade: 1,
	})
	registerQuestion(QuestionDef{
		ID: "q-add-2-2", Category: "Matematika", Prompt: "2 + 2 = ?",
		Choices: []string{"3", "4", "5"}, Correct: 1,
		Explain: "2 ditambah 2 sama dengan 4.", Grade: 1,
	})
	registerQuest(QuestDef{
		ID: "pq-mine-1", Title: "Sinau Nambang.", Kind: "profession", NPC: "mbah_batu",
		Location: "Stone Valley", Description: "Ambil 3 Valley Stone bersama Mbah Batu.",
		Objectives: []ObjectiveDef{{Type: "GATHER", Target: "mat-valley-stone", Count: 3, Text: "Ambil 3 Stone"}},
		Rewards:    RewardDef{Exp: 40, Coin: 8},
		FlagsOnClaim: []string{"pq_mine_done"}, ClaimAt: "mbah_batu",
	})
	registerQuest(QuestDef{
		ID: "pq-edu-1", Title: "Sinau Ngetung.", Kind: "education", NPC: "mbok_rasa",
		Location: "Dapur Desa", Description: "Jawab 2 + 2 bersama Mbok Rasa.",
		Objectives: []ObjectiveDef{{Type: "ANSWER", Target: "q-add-2-2", Count: 1, Text: "Jawab 2 + 2"}},
		Rewards:    RewardDef{Exp: 20, EduToken: 1, Knowledge: 1},
		FlagsOnClaim: []string{"pq_edu_done"}, ClaimAt: "mbok_rasa",
	})
}

func registerPhase21Meta() {
	for _, t := range []TitleDef{
		{ID: "gatherer", Name: "Gatherer", Source: "profession"},
		{ID: "craftsman", Name: "Craftsman", Source: "profession"},
		{ID: "cook", Name: "Cook", Source: "profession"},
		{ID: "fisher", Name: "Fisher", Source: "profession"},
		{ID: "artisan", Name: "Artisan", Source: "profession"},
	} {
		registerTitleDef(t)
	}
	for _, a := range []AchievementDef{
		{ID: "first-craft", Name: "First Craft", Title: "craftsman", Flag: "ach_first_craft", Category: "Profession"},
		{ID: "first-catch", Name: "First Catch", Title: "fisher", Flag: "ach_first_catch", Category: "Profession"},
		{ID: "first-mine", Name: "First Mine", Title: "gatherer", Flag: "ach_first_mine", Category: "Profession"},
		{ID: "first-harvest", Name: "First Harvest", Title: "gatherer", Flag: "ach_first_harvest", Category: "Profession"},
		{ID: "master-cook", Name: "Master Cook", Title: "cook", Flag: "ach_master_cook", Category: "Profession"},
		{ID: "master-artisan", Name: "Master Artisan", Title: "artisan", Flag: "ach_master_artisan", Category: "Profession"},
		{ID: "world-gatherer", Name: "World Gatherer", Title: "gatherer", Flag: "ach_world_gatherer", Category: "Profession"},
	} {
		registerAchievementDef(a)
	}
	registerCosmeticDef(CosmeticDef{ID: "apron-karya", Name: "Festival Karya Apron", Kind: "costume"})
	registerCosmeticDef(CosmeticDef{ID: "title-karya", Name: "Festival Karya Frame", Kind: "frame"})
	registerEventDef(EventDef{
		ID: "festival-karya", Name: "FESTIVAL KARYA", Kind: "social", Region: "village",
		Announce: "Festival Karya: craft, masak, mancing, lan pamer karya.", DurationSec: 180, Rewards: RewardDef{Exp: 20, EduToken: 2},
	})
}

func (p *Player) ensurePhase21() *PlayerLog {
	log := p.ensureLog()
	if log.Materials == nil {
		log.Materials = map[string]int{}
	}
	if log.MaterialCodex == nil {
		log.MaterialCodex = map[string]bool{}
	}
	if log.ProfessionXP == nil {
		log.ProfessionXP = map[string]int{}
	}
	if log.Recipes == nil {
		log.Recipes = map[string]string{}
	}
	if log.RecipeCrafts == nil {
		log.RecipeCrafts = map[string]int{}
	}
	if log.CraftQueue == nil {
		log.CraftQueue = []CraftJob{}
	}
	if log.StallListings == nil {
		log.StallListings = []StallListing{}
	}
	if log.Quests == nil {
		log.Quests = map[string]*QuestLog{}
	}
	for _, def := range questCatalog {
		if log.Quests[def.ID] != nil {
			continue
		}
		ql := &QuestLog{ID: def.ID, State: QuestLocked, Progress: make([]int, len(def.Objectives)), Perfect: true}
		if len(def.Prereq) == 0 {
			ql.State = QuestAvailable
		}
		log.Quests[def.ID] = ql
	}
	for _, rec := range recipeCatalog {
		if log.Recipes[rec.ID] != "" {
			continue
		}
		if rec.Discover == "start" {
			log.Recipes[rec.ID] = RecipeDiscovered
		} else {
			log.Recipes[rec.ID] = RecipeLocked
		}
	}
	if log.GearDurability <= 0 {
		log.GearDurability = 100
	}
	return log
}

func (w *WorldState) ApplyPhase21(p *Player, env Envelope) [][]byte {
	p.ensurePhase21()
	switch env.Type {
	case TypeGather:
		var in struct{ NodeID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.gatherNode(p, in.NodeID, in.TransactionID)
	case TypeCraft:
		var in struct{ RecipeID, StationID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.craftRecipe(p, in.RecipeID, in.StationID, in.TransactionID)
	case TypeGetCrafting:
		return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
	case TypeSetProfession:
		var in struct{ ProfessionID string }
		_ = unmarshal(env.Data, &in)
		return w.setProfession(p, in.ProfessionID)
	case TypeResetProfession:
		return w.resetProfessions(p)
	case TypeFishStart:
		var in struct{ SpotID string }
		_ = unmarshal(env.Data, &in)
		return w.fishStart(p, in.SpotID)
	case TypeFishCatch:
		var in struct{ SpotID string; Progress int }
		_ = unmarshal(env.Data, &in)
		return w.fishCatch(p, in.SpotID, in.Progress)
	case TypeNpcShopOpen:
		var in struct{ ShopID string }
		_ = unmarshal(env.Data, &in)
		return w.npcShopOpen(p, in.ShopID)
	case TypeNpcShopBuy:
		var in struct{ ShopID, ItemID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.npcShopBuy(p, in.ShopID, in.ItemID, in.TransactionID)
	case TypeNpcShopSell:
		var in struct{ ShopID, ItemID, TransactionID string; Qty int }
		_ = unmarshal(env.Data, &in)
		return w.npcShopSell(p, in.ShopID, in.ItemID, in.Qty, in.TransactionID)
	case TypeNpcRepair:
		var in struct{ TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.npcRepair(p, in.TransactionID)
	case TypeStallOpen:
		return w.stallSet(p, true)
	case TypeStallClose:
		return w.stallSet(p, false)
	case TypeStallList:
		var in struct{ ItemID string; Qty, Price int }
		_ = unmarshal(env.Data, &in)
		return w.stallList(p, in.ItemID, in.Qty, in.Price)
	case TypeStallBuy:
		var in struct{ SellerID, ItemID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.stallBuy(p, in.SellerID, in.ItemID, in.TransactionID)
	case TypeCraftOrder:
		var in struct{ RecipeID string; Reward int }
		_ = unmarshal(env.Data, &in)
		return w.craftOrderCreate(p, in.RecipeID, in.Reward)
	case TypeCraftOrderAccept:
		var in struct{ OrderID string }
		_ = unmarshal(env.Data, &in)
		return w.craftOrderAccept(p, in.OrderID)
	case TypeGetWorkshop:
		return w.guildWorkshop(p)
	case TypeAddGold, TypeAddMaterial:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return w.ApplyPhase28(p, env)
	}
}

func professionLevel(xp int) int {
	lv, need, left := 1, 40, xp
	for lv < ProfessionMaxLevel && left >= need {
		left -= need
		lv++
		need += 12
	}
	return lv
}

func gatherProfessionFor(resType string) string {
	switch resType {
	case "STONE", "ORE":
		return ProfMiner
	case "WOOD":
		return ProfWoodcutter
	case "HERB", "FRUIT", "FIBER":
		return ProfHerbalist
	case "FISH":
		return ProfFisher
	}
	return ""
}

func (p *Player) professionActive(id string) bool {
	log := p.ensurePhase21()
	for _, a := range log.ActiveGather {
		if a == id {
			return true
		}
	}
	for _, a := range log.ActiveCraft {
		if a == id {
			return true
		}
	}
	return false
}

func (p *Player) activateProfession(id string) bool {
	def, ok := professionBy[id]
	if !ok {
		return false
	}
	log := p.ensurePhase21()
	if p.professionActive(id) {
		return true
	}
	if def.Kind == "gather" {
		if len(log.ActiveGather) >= GatherProfessionCap {
			return false
		}
		log.ActiveGather = append(log.ActiveGather, id)
		return true
	}
	if len(log.ActiveCraft) >= CraftProfessionCap {
		return false
	}
	log.ActiveCraft = append(log.ActiveCraft, id)
	return true
}

func (p *Player) addProfessionXP(id string, amount int) {
	if amount <= 0 || professionBy[id].ID == "" {
		return
	}
	log := p.ensurePhase21()
	log.ProfessionXP[id] += amount
	if professionLevel(log.ProfessionXP[id]) >= 20 {
		if id == ProfCook {
			log.Flags["ach_master_cook"] = true
		}
		if id == ProfArtisan {
			log.Flags["ach_master_artisan"] = true
		}
	}
	if professionLevel(log.ProfessionXP[id]) >= 8 {
		for _, rec := range recipeCatalog {
			if rec.Discover == "master" && rec.Profession == id && log.Recipes[rec.ID] == RecipeLocked {
				log.Recipes[rec.ID] = RecipeDiscovered
			}
		}
	}
}

func (p *Player) materialCount(id string) int {
	return p.ensurePhase21().Materials[id]
}

func (p *Player) addMaterial(id string, qty int) bool {
	if qty <= 0 || resourceByID[id].ID == "" && itemByID[id].Type != "MATERIAL" && itemByID[id].ID == "" {
		if itemByID[id].ID == "" && resourceByID[id].ID == "" {
			return false
		}
	}
	log := p.ensurePhase21()
	have := log.Materials[id]
	if have+qty > MaterialStackMax {
		return false
	}
	log.Materials[id] = have + qty
	log.MaterialCodex[id] = true
	if len(log.MaterialCodex) >= 5 {
		log.Flags["ach_world_gatherer"] = true
	}
	p.markDirty()
	return true
}

func (p *Player) takeMaterial(id string, qty int) bool {
	log := p.ensurePhase21()
	if log.Materials[id] < qty {
		return false
	}
	log.Materials[id] -= qty
	if log.Materials[id] <= 0 {
		delete(log.Materials, id)
	}
	p.markDirty()
	return true
}

func (w *WorldState) gatherNode(p *Player, nodeID, tx string) [][]byte {
	if tx != "" {
		if prev, ok := w.txSeen(tx); ok {
			return prev
		}
	}
	node, ok := nodeByID[nodeID]
	if !ok {
		obj, found := interactByID[nodeID]
		if !found {
			return rejectFor(p.ID, TypeGather, "node")
		}
		node = ResourceNodeDef{ID: obj.ID, Kind: obj.Kind, X: obj.X, Z: obj.Z}
		for _, n := range nodeCatalog {
			if n.ID == obj.ID {
				node = n
				break
			}
		}
		if node.ResourceID == "" {
			return rejectFor(p.ID, TypeGather, "node")
		}
	}
	if math.Hypot(p.X-node.X, p.Z-node.Z) > GatherRange {
		return rejectFor(p.ID, TypeGather, "distance")
	}
	if w.NodeCooldown == nil {
		w.NodeCooldown = map[string]int64{}
	}
	if time.Now().Unix() < w.NodeCooldown[node.ID] {
		return rejectFor(p.ID, TypeGather, "respawn")
	}
	res := resourceByID[node.ResourceID]
	if res.ID == "" {
		return rejectFor(p.ID, TypeGather, "resource")
	}
	prof := gatherProfessionFor(res.Type)
	if prof != "" && !p.professionActive(prof) && !p.activateProfession(prof) {
		return rejectFor(p.ID, TypeGather, "profession")
	}
	if gate := w.phase28GatherGate(p, node); gate != nil {
		return gate
	}
	if !p.addMaterial(res.ID, 1) {
		return rejectFor(p.ID, TypeGather, "bag")
	}
	sec := res.RespawnSec
	if sec < 4 {
		sec = 8
	}
	w.NodeCooldown[node.ID] = time.Now().Unix() + int64(sec)
	p.credit("GATHER", res.ID, 1)
	p.addProfessionXP(prof, 8)
	w.giveExp(p, 4)
	switch res.Type {
	case "STONE", "ORE":
		p.ensureLog().Flags["ach_first_mine"] = true
	case "HERB", "FRUIT", "FIBER", "WOOD":
		p.ensureLog().Flags["ach_first_harvest"] = true
	}
	p.grantTitle(professionBy[prof].Title)
	w.phase28AfterGather(p, node, res)
	w.persist(p)
	w.audit("GATHER", p.ID, res.ID)
	out := [][]byte{
		marshal(TypeGatherResult, map[string]any{"nodeId": node.ID, "resourceId": res.ID, "anim": gatherAnim(res.Type), "toId": p.ID}),
		marshal(TypeCraftingState, w.craftingView(p)),
		marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
	}
	out = append(out, w.refreshAchievements(p)...)
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func gatherAnim(resType string) string {
	switch resType {
	case "STONE", "ORE":
		return "mining"
	case "WOOD":
		return "chop"
	case "FISH":
		return "cast"
	default:
		return "collect"
	}
}

func (w *WorldState) setProfession(p *Player, id string) [][]byte {
	if professionBy[id].ID == "" {
		return rejectFor(p.ID, TypeSetProfession, "profession")
	}
	if !p.activateProfession(id) {
		return rejectFor(p.ID, TypeSetProfession, "limit")
	}
	w.persist(p)
	return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
}

func (w *WorldState) resetProfessions(p *Player) [][]byte {
	log := p.ensurePhase21()
	if log.ProfessionResets >= ProfessionResetLimit {
		return rejectFor(p.ID, TypeResetProfession, "limit")
	}
	log.ProfessionResets++
	log.ActiveGather, log.ActiveCraft = nil, nil
	w.persist(p)
	return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
}

func stationKindFor(cat string) string {
	switch cat {
	case "WEAPON", "ARMOR":
		return "forge"
	case "FOOD":
		return "cooking-fire"
	case "POTION":
		return "alchemy"
	case "MATERIAL":
		return "forge"
	default:
		return "workbench"
	}
}

func craftQuality(level int) string {
	if level >= 35 {
		return QualityMasterwork
	}
	if level >= 15 {
		return QualityFine
	}
	return QualityNormal
}

func (w *WorldState) craftRecipe(p *Player, recipeID, stationID, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	if w.limited("craft:"+p.ID, 6, 2*time.Second) {
		return rejectFor(p.ID, TypeCraft, "cooldown")
	}
	rec, ok := recipeByID[recipeID]
	if !ok {
		return rejectFor(p.ID, TypeCraft, "recipe")
	}
	log := p.ensurePhase21()
	if log.Recipes[recipeID] == RecipeLocked || log.Recipes[recipeID] == "" {
		return rejectFor(p.ID, TypeCraft, "locked")
	}
	if !p.professionActive(rec.Profession) && !p.activateProfession(rec.Profession) {
		return rejectFor(p.ID, TypeCraft, "profession")
	}
	if professionLevel(log.ProfessionXP[rec.Profession]) < rec.RequiredLevel {
		return rejectFor(p.ID, TypeCraft, "level")
	}
	st := interactByID[stationID]
	if st.ID == "" || st.Kind != stationKindFor(rec.Category) {
		return rejectFor(p.ID, TypeCraft, "station")
	}
	if math.Hypot(p.X-st.X, p.Z-st.Z) > GatherRange {
		return rejectFor(p.ID, TypeCraft, "distance")
	}
	w.flushCraftQueue(p)
	if len(log.CraftQueue) >= CraftQueueMax {
		return rejectFor(p.ID, TypeCraft, "queue")
	}
	for _, m := range rec.Materials {
		if p.materialCount(m.ItemID) < m.Qty {
			return rejectFor(p.ID, TypeCraft, "material")
		}
	}
	fee := craftFeeNow()
	if g := w.guildOf(p.ID); g != nil && g.Workshop {
		fee = 0
	}
	if log.Recipes[recipeID] == RecipeMastered {
		fee = 0
	}
	if fee > 0 && !w.removeCurrency(p, "gold", fee, "craft_fee") {
		return rejectFor(p.ID, TypeCraft, "gold")
	}
	if fee > 0 {
		w.GoldRemoved += fee
		w.recordEconomy(p.ID, "gold", -fee, "craft_fee", tx)
	}
	taken := []RecipeMat{}
	for _, m := range rec.Materials {
		if !p.takeMaterial(m.ItemID, m.Qty) {
			for _, t := range taken {
				p.addMaterial(t.ItemID, t.Qty)
			}
			if fee > 0 {
				w.addCurrency(p, "gold", fee, "craft_rollback")
			}
			return rejectFor(p.ID, TypeCraft, "material")
		}
		taken = append(taken, m)
	}
	q := craftQuality(professionLevel(log.ProfessionXP[rec.Profession]))
	sec := rec.CraftSec
	if g := w.guildOf(p.ID); g != nil && g.Workshop && sec > 0 {
		sec = int(float64(sec) * 0.9)
	}
	if sec <= 0 {
		if !w.giveItem(p, rec.Result, 1) {
			for _, t := range taken {
				p.addMaterial(t.ItemID, t.Qty)
			}
			if fee > 0 {
				w.addCurrency(p, "gold", fee, "craft_rollback")
			}
			return rejectFor(p.ID, TypeCraft, "space")
		}
	} else {
		log.CraftQueue = append(log.CraftQueue, CraftJob{
			ID: "job-" + recipeID + "-" + itoa(int(time.Now().UnixNano()%100000)),
			RecipeID: recipeID, Result: rec.Result, Quality: q,
			ReadyAt: time.Now().UnixMilli() + int64(sec)*1000,
		})
	}
	log.RecipeCrafts[recipeID]++
	if log.RecipeCrafts[recipeID] >= 8 {
		log.Recipes[recipeID] = RecipeMastered
	}
	log.Flags["ach_first_craft"] = true
	w.phase28AfterCraft(p, recipeID)
	if rec.Category == "FOOD" {
		log.Flags["daily_cook"] = true
	}
	p.addProfessionXP(rec.Profession, 10)
	p.grantTitle(professionBy[rec.Profession].Title)
	w.CraftVolume++
	w.persist(p)
	w.persistGear(p)
	out := [][]byte{
		marshal(TypeCraftResult, map[string]any{"recipeId": recipeID, "result": rec.Result, "quality": q, "toId": p.ID}),
		marshal(TypeCraftingState, w.craftingView(p)),
		marshal(TypeInventoryUpdated, p.loadout("", nil)),
	}
	out = append(out, w.refreshAchievements(p)...)
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func (w *WorldState) flushCraftQueue(p *Player) {
	log := p.ensurePhase21()
	now := time.Now().UnixMilli()
	keep := log.CraftQueue[:0]
	for _, job := range log.CraftQueue {
		if job.ReadyAt > now {
			keep = append(keep, job)
			continue
		}
		if !w.giveItem(p, job.Result, 1) {
			keep = append(keep, job)
		}
	}
	log.CraftQueue = keep
}

func (w *WorldState) stationInteract(p *Player, obj InteractDef) [][]byte {
	p.ensurePhase21()
	w.flushCraftQueue(p)
	return [][]byte{
		marshal(TypeInteractResult, InteractResult{
			Kind: obj.Kind, TargetID: obj.ID, Title: obj.Text, Speaker: obj.Text,
			Text: "Le, bahan iki isa digawe dadi gegaman. Coba kok gawa menyang pandhe.",
			Subtitle: "Nak, bahan ini bisa dibuat menjadi senjata. Coba bawa ke pandai besi.",
			Options: []DialogOption{{ID: "craft-ui", Label: "Buka crafting"}, {ID: "close", Label: "Tutup"}},
		}),
		marshal(TypeCraftingState, w.craftingView(p)),
	}
}

func (w *WorldState) fishStart(p *Player, spotID string) [][]byte {
	spot, ok := spotByID[spotID]
	if !ok {
		obj := interactByID[spotID]
		if obj.Kind != "fishing-spot" {
			return rejectFor(p.ID, TypeFishStart, "spot")
		}
		spot = FishSpotDef{ID: obj.ID, X: obj.X, Z: obj.Z, Reward: "mat-river-fish", Place: "lake"}
	}
	if math.Hypot(p.X-spot.X, p.Z-spot.Z) > GatherRange {
		return rejectFor(p.ID, TypeFishStart, "distance")
	}
	if !p.professionActive(ProfFisher) && !p.activateProfession(ProfFisher) {
		return rejectFor(p.ID, TypeFishStart, "profession")
	}
	if w.limited("fish:"+p.ID, 8, 2*time.Second) {
		return rejectFor(p.ID, TypeFishStart, "cooldown")
	}
	log := p.ensurePhase21()
	log.FishSpot = spot.ID
	log.FishTargetA, log.FishTargetB = 42, 68
	w.persist(p)
	return [][]byte{marshal(TypeFishState, map[string]any{
		"spotId": spot.ID, "place": spot.Place, "targetA": log.FishTargetA, "targetB": log.FishTargetB,
		"prompt": "Tap / SPACE saat di zona.", "toId": p.ID, "anim": "cast",
	})}
}

func (w *WorldState) fishCatch(p *Player, spotID string, progress int) [][]byte {
	log := p.ensurePhase21()
	if log.FishSpot == "" || (spotID != "" && log.FishSpot != spotID) {
		return rejectFor(p.ID, TypeFishCatch, "session")
	}
	spot := spotByID[log.FishSpot]
	if progress < log.FishTargetA || progress > log.FishTargetB {
		log.FishSpot = ""
		w.persist(p)
		return rejectFor(p.ID, TypeFishCatch, "timing")
	}
	reward := spot.Reward
	if reward == "" {
		reward = "mat-river-fish"
	}
	if !p.addMaterial(reward, 1) {
		return rejectFor(p.ID, TypeFishCatch, "bag")
	}
	w.phase28AfterFish(p, progress)
	log.FishSpot = ""
	log.Flags["ach_first_catch"] = true
	log.Flags["daily_fish"] = true
	p.ensurePhase22()
	log.Collections["fish:"+reward] = true
	p.addProfessionXP(ProfFisher, 10)
	p.grantTitle("fisher")
	w.giveExp(p, 6)
	p.credit("GATHER", reward, 1)
	events := [][]byte{
		marshal(TypeFishState, map[string]any{"spotId": spot.ID, "caught": true, "reward": reward, "anim": "reel", "toId": p.ID}),
		marshal(TypeCraftingState, w.craftingView(p)),
	}
	events = append(events, w.refreshAchievements(p)...)
	if !log.Flags["fish_edu_done"] {
		log.Flags["fish_edu_done"] = true
		log.Quiz = QuizSession{QuestID: "fish-edu", Active: true}
		q := questionByID["q-iwak-2-3"]
		events = append(events, marshal(TypeEducationQuestion, QuestionOut{
			ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices,
		}))
	}
	w.persist(p)
	return events
}

func (w *WorldState) startProfessionQuiz(p *Player) [][]byte {
	p.ensurePhase21()
	q := questionByID["q-add-2-2"]
	p.ensureLog().Quiz = QuizSession{QuestID: "pq-edu-1", Active: true}
	quest := p.quest("pq-edu-1")
	if quest != nil && quest.State == QuestAvailable {
		quest.State = QuestActive
	}
	return [][]byte{marshal(TypeEducationQuestion, QuestionOut{
		ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices,
	})}
}

func (w *WorldState) answerProfessionQuiz(p *Player, in EducationAnswerIn) [][]byte {
	log := p.ensurePhase21()
	q := questionByID[in.QuestionID]
	if q.ID == "" {
		return rejectFor(p.ID, TypeEducationAnswer, "question")
	}
	want := "q-add-2-2"
	switch log.Quiz.QuestID {
	case "fish-edu":
		want = "q-iwak-2-3"
	case "craft-edu":
		want = "q-kayu-3-2"
	case "econ-edu":
		want = "q-emas-15-10"
	}
	if in.QuestionID != want {
		return rejectFor(p.ID, TypeEducationAnswer, "order")
	}
	if in.Choice != q.Correct {
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{Correct: false, Explain: q.Explain, Retry: true, Toast: "Coba lagi."})}
	}
	log.Quiz.Active = false
	w.addCurrency(p, "edu", 1, "profession_quiz")
	log.KnowledgePoints++
	if p.Bag.canFit("knowledge_token", 1) {
		p.Bag.add("knowledge_token", 1)
	}
	p.credit("ANSWER", in.QuestionID, 1)
	w.recordEconomy(p.ID, "edu", 1, "profession_quiz", "")
	events := w.grantExp(p, 6)
	events = append(events, marshal(TypeEducationFeedback, EducationFeedback{Correct: true, Explain: q.Explain, Toast: "Knowledge Token +1"}))
	events = append(events, marshal(TypeCraftingState, w.craftingView(p)))
	events = append(events, marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)))
	w.persist(p)
	return events
}

func professionNpcOptions(p *Player, npc NPCDef, questID, acceptLabel string) []DialogOption {
	opts := []DialogOption{}
	if npc.Type == "BLACKSMITH" || npc.Type == "MINER" {
		opts = append(opts, DialogOption{ID: "npc-shop:mbah-karya-shop", Label: "Toko bahan"})
		opts = append(opts, DialogOption{ID: "repair", Label: "Ndandani gegaman"})
	}
	if npc.Type == "COOK" {
		opts = append(opts, DialogOption{ID: "npc-shop:mbok-rasa-shop", Label: "Toko dapur"})
		opts = append(opts, DialogOption{ID: "quiz-prof", Label: "2 + 2 = ?"})
	}
	if npc.Type == "FISHER" {
		opts = append(opts, DialogOption{ID: "fish-ui", Label: "Sinau mancing"})
		opts = append(opts, questOption(p, "pq-fish-1", "Nelayan Cilik.")...)
	}
	if questID != "" {
		opts = append(opts, questOption(p, questID, acceptLabel)...)
	}
	opts = append(opts, DialogOption{ID: "close", Label: "Tutup"})
	return opts
}

func (w *WorldState) npcShopOpen(p *Player, shopID string) [][]byte {
	shop, ok := npcShopByID[shopID]
	if !ok {
		return rejectFor(p.ID, TypeNpcShopOpen, "shop")
	}
	items := make([]map[string]any, 0, len(shop.Items))
	for _, it := range shop.Items {
		def := itemByID[it.ItemID]
		items = append(items, map[string]any{
			"itemId": it.ItemID, "name": def.Name, "price": shopBuyCost(it), "buyPrice": shopBuyCost(it),
			"sellPrice": shopSellGain(it), "currency": "gold", "rarity": def.Rarity,
			"buy": it.Buy, "sell": it.Sell, "owned": p.materialCount(it.ItemID),
		})
	}
	return [][]byte{marshal(TypeShopCatalog, map[string]any{
		"id": shop.ID, "name": shop.Name, "items": items, "wallet": p.wallet(), "mode": "material",
	})}
}

func (w *WorldState) npcShopBuy(p *Player, shopID, itemID, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	if w.limited("buy:"+p.ID, 8, 2*time.Second) {
		return rejectFor(p.ID, TypeNpcShopBuy, "cooldown")
	}
	shop, ok := npcShopByID[shopID]
	if !ok {
		return rejectFor(p.ID, TypeNpcShopBuy, "shop")
	}
	var it NpcShopItem
	found := false
	for _, row := range shop.Items {
		if row.ItemID == itemID && row.Buy {
			it, found = row, true
			break
		}
	}
	if !found {
		return rejectFor(p.ID, TypeNpcShopBuy, "item")
	}
	if !w.removeCurrency(p, "gold", shopBuyCost(it), "npc_buy") {
		return rejectFor(p.ID, TypeNpcShopBuy, "gold")
	}
	if !p.addMaterial(it.ItemID, 1) {
		w.addCurrency(p, "gold", shopBuyCost(it), "npc_buy_rollback")
		return rejectFor(p.ID, TypeNpcShopBuy, "bag")
	}
	w.GoldRemoved += shopBuyCost(it)
	w.recordEconomy(p.ID, "gold", -shopBuyCost(it), "npc_buy", tx)
	w.persist(p)
	out := [][]byte{marshal(TypeCraftingState, w.craftingView(p)), marshal(TypeInventoryUpdated, p.loadout("", nil))}
	out = append(out, w.npcShopOpen(p, shopID)...)
	return w.rememberTx(tx, out)
}

func (w *WorldState) npcShopSell(p *Player, shopID, itemID string, qty int, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	if w.limited("sell:"+p.ID, 8, 2*time.Second) {
		return rejectFor(p.ID, TypeNpcShopSell, "cooldown")
	}
	if qty != 1 {
		qty = 1
	}
	shop, ok := npcShopByID[shopID]
	if !ok {
		return rejectFor(p.ID, TypeNpcShopSell, "shop")
	}
	var it NpcShopItem
	found := false
	for _, row := range shop.Items {
		if row.ItemID == itemID && row.Sell {
			it, found = row, true
			break
		}
	}
	if !found {
		return rejectFor(p.ID, TypeNpcShopSell, "item")
	}
	if why := phase28SellBlocked(itemID); why != "" {
		return rejectFor(p.ID, TypeNpcShopSell, why)
	}
	if !p.takeMaterial(it.ItemID, qty) {
		return rejectFor(p.ID, TypeNpcShopSell, "owned")
	}
	gain := shopSellGain(it)
	if !w.addCurrency(p, "gold", gain, "npc_sell") {
		p.addMaterial(it.ItemID, qty)
		return rejectFor(p.ID, TypeNpcShopSell, "gold")
	}
	w.GoldCreated += gain
	w.recordEconomy(p.ID, "gold", gain, "npc_sell", tx)
	w.persist(p)
	out := [][]byte{marshal(TypeCraftingState, w.craftingView(p)), marshal(TypeInventoryUpdated, p.loadout("", nil))}
	out = append(out, w.npcShopOpen(p, shopID)...)
	return w.rememberTx(tx, out)
}

func (w *WorldState) npcRepair(p *Player, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	log := p.ensurePhase21()
	if !w.removeCurrency(p, "gold", RepairFeeGold, "repair") {
		return rejectFor(p.ID, TypeNpcRepair, "gold")
	}
	log.GearDurability = 100
	w.GoldRemoved += RepairFeeGold
	w.recordEconomy(p.ID, "gold", -RepairFeeGold, "repair", tx)
	w.persist(p)
	out := [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func (w *WorldState) stallSet(p *Player, open bool) [][]byte {
	log := p.ensurePhase21()
	log.StallOpen = open
	if w.Stalls == nil {
		w.Stalls = map[string]*PlayerStall{}
	}
	w.Stalls[p.ID] = &PlayerStall{PlayerID: p.ID, Open: open, Items: append([]StallListing{}, log.StallListings...)}
	w.persist(p)
	return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
}

func (w *WorldState) stallList(p *Player, itemID string, qty, price int) [][]byte {
	if qty != 1 {
		qty = 1
	}
	if price < 1 || price > 100000 {
		return rejectFor(p.ID, TypeStallList, "price")
	}
	def := itemByID[itemID]
	if def.Untradable {
		return rejectFor(p.ID, TypeStallList, "bound")
	}
	if !p.takeMaterial(itemID, qty) {
		return rejectFor(p.ID, TypeStallList, "owned")
	}
	log := p.ensurePhase21()
	log.StallOpen = true
	log.StallListings = append(log.StallListings, StallListing{ItemID: itemID, Qty: qty, Price: price})
	if w.Stalls == nil {
		w.Stalls = map[string]*PlayerStall{}
	}
	w.Stalls[p.ID] = &PlayerStall{PlayerID: p.ID, Open: true, Items: append([]StallListing{}, log.StallListings...)}
	w.persist(p)
	return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
}

func (w *WorldState) stallBuy(p *Player, sellerID, itemID, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	if sellerID == p.ID {
		return rejectFor(p.ID, TypeStallBuy, "self")
	}
	seller := w.players[sellerID]
	if seller == nil {
		return rejectFor(p.ID, TypeStallBuy, "seller")
	}
	slog := seller.ensurePhase21()
	idx := -1
	var row StallListing
	for i, it := range slog.StallListings {
		if it.ItemID == itemID {
			idx, row = i, it
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeStallBuy, "item")
	}
	if !w.removeCurrency(p, "gold", row.Price, "stall_buy") {
		return rejectFor(p.ID, TypeStallBuy, "gold")
	}
	if !p.addMaterial(row.ItemID, row.Qty) {
		w.addCurrency(p, "gold", row.Price, "stall_rollback")
		return rejectFor(p.ID, TypeStallBuy, "bag")
	}
	w.addCurrency(seller, "gold", row.Price, "stall_sale")
	slog.StallListings = append(slog.StallListings[:idx], slog.StallListings[idx+1:]...)
	w.GoldCreated += 0
	w.recordEconomy(p.ID, "gold", -row.Price, "stall_buy", tx)
	w.recordEconomy(seller.ID, "gold", row.Price, "stall_sale", tx+"-s")
	w.persist(p)
	w.persist(seller)
	out := [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
	return w.rememberTx(tx, out)
}

func (w *WorldState) craftOrderCreate(p *Player, recipeID string, reward int) [][]byte {
	if recipeByID[recipeID].ID == "" {
		return rejectFor(p.ID, TypeCraftOrder, "recipe")
	}
	if reward < 1 {
		reward = 10
	}
	if !w.removeCurrency(p, "gold", reward, "craft_order_escrow") {
		return rejectFor(p.ID, TypeCraftOrder, "gold")
	}
	if w.CraftOrders == nil {
		w.CraftOrders = map[string]*CraftOrder{}
	}
	id := "ord-" + p.ID + "-" + itoa(int(time.Now().UnixNano()%100000))
	w.CraftOrders[id] = &CraftOrder{ID: id, Requester: p.ID, RecipeID: recipeID, Reward: reward, Status: "OPEN"}
	w.persist(p)
	return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
}

func (w *WorldState) craftOrderAccept(p *Player, orderID string) [][]byte {
	ord := w.CraftOrders[orderID]
	if ord == nil || ord.Status != "OPEN" {
		return rejectFor(p.ID, TypeCraftOrderAccept, "order")
	}
	rec := recipeByID[ord.RecipeID]
	for _, m := range rec.Materials {
		if p.materialCount(m.ItemID) < m.Qty {
			return rejectFor(p.ID, TypeCraftOrderAccept, "material")
		}
	}
	for _, m := range rec.Materials {
		p.takeMaterial(m.ItemID, m.Qty)
	}
	req := w.players[ord.Requester]
	if req == nil || !w.giveItem(req, rec.Result, 1) {
		for _, m := range rec.Materials {
			p.addMaterial(m.ItemID, m.Qty)
		}
		return rejectFor(p.ID, TypeCraftOrderAccept, "deliver")
	}
	w.addCurrency(p, "gold", ord.Reward, "craft_order")
	ord.Status = "DONE"
	ord.Crafter = p.ID
	w.persist(p)
	if req != nil {
		w.persist(req)
		w.persistGear(req)
	}
	return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
}

func (w *WorldState) guildWorkshop(p *Player) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil {
		return rejectFor(p.ID, TypeGetWorkshop, "guild")
	}
	g.Workshop = true
	return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
}

func (w *WorldState) refreshProfessionAchievements(p *Player, grant func(id, title string)) {
	log := p.ensurePhase21()
	if professionLevel(log.ProfessionXP[ProfCook]) >= 20 {
		log.Flags["ach_master_cook"] = true
	}
	if professionLevel(log.ProfessionXP[ProfArtisan]) >= 20 {
		log.Flags["ach_master_artisan"] = true
	}
	if len(log.MaterialCodex) >= 5 {
		log.Flags["ach_world_gatherer"] = true
	}
	_ = grant
}

func (w *WorldState) craftingView(p *Player) map[string]any {
	log := p.ensurePhase21()
	w.flushCraftQueue(p)
	profs := make([]map[string]any, 0, len(professionCatalog))
	for _, d := range professionCatalog {
		profs = append(profs, map[string]any{
			"id": d.ID, "name": professionLabel(d.ID, d.Name), "kind": d.Kind, "title": d.Title,
			"xp": log.ProfessionXP[d.ID], "level": professionLevel(log.ProfessionXP[d.ID]),
			"next": nextProfessionXP(log.ProfessionXP[d.ID]),
			"active": p.professionActive(d.ID),
		})
	}
	recipes := make([]map[string]any, 0, len(recipeCatalog))
	for _, r := range recipeCatalog {
		owned := map[string]int{}
		for _, m := range r.Materials {
			owned[m.ItemID] = p.materialCount(m.ItemID)
		}
		recipes = append(recipes, map[string]any{
			"id": r.ID, "name": r.Name, "profession": r.Profession, "category": r.Category,
			"requiredLevel": r.RequiredLevel, "craftTime": r.CraftSec, "result": r.Result,
			"materials": r.Materials, "owned": owned, "status": log.Recipes[r.ID],
			"station": stationKindFor(r.Category),
		})
	}
	mats := make([]map[string]any, 0, len(log.Materials))
	for id, n := range log.Materials {
		def := resourceByID[id]
		name := def.Name
		if name == "" {
			name = itemByID[id].Name
		}
		mats = append(mats, map[string]any{
			"id": id, "name": name, "qty": n, "desc": def.NameJV, "region": def.Region, "type": def.Type, "rarity": def.Rarity,
			"codex": log.MaterialCodex[id],
		})
	}
	codex := make([]map[string]any, 0, len(resourceCatalog))
	for _, r := range resourceCatalog {
		if !log.MaterialCodex[r.ID] {
			continue
		}
		codex = append(codex, map[string]any{"id": r.ID, "name": r.Name, "desc": r.NameJV, "region": r.Region, "rarity": r.Rarity})
	}
	festival := false
	if w.Events.Active != nil && w.Events.Active.Def.ID == "festival-karya" {
		festival = true
	}
	workshop := false
	if g := w.guildOf(p.ID); g != nil {
		workshop = g.Workshop
	}
	stalls := []map[string]any{}
	for id, s := range w.Stalls {
		if s == nil || !s.Open {
			continue
		}
		stalls = append(stalls, map[string]any{"playerId": id, "items": s.Items})
	}
	orders := []map[string]any{}
	for _, o := range w.CraftOrders {
		if o == nil {
			continue
		}
		orders = append(orders, map[string]any{"id": o.ID, "requester": o.Requester, "crafter": o.Crafter, "recipe": o.RecipeID, "reward": o.Reward, "status": o.Status})
	}
	merchants := []map[string]any{}
	for _, sid := range []string{"mbah-karya-shop", "mbok-rasa-shop", "pedagang-shop"} {
		shop, ok := npcShopByID[sid]
		if !ok {
			continue
		}
		its := make([]map[string]any, 0, len(shop.Items))
		for _, it := range shop.Items {
			def := itemByID[it.ItemID]
			its = append(its, map[string]any{
				"itemId": it.ItemID, "name": def.Name, "price": shopBuyCost(it), "buyPrice": shopBuyCost(it),
				"sellPrice": shopSellGain(it), "rarity": def.Rarity, "buy": it.Buy, "sell": it.Sell,
			})
		}
		merchants = append(merchants, map[string]any{"id": shop.ID, "name": shop.Name, "items": its})
	}
	return map[string]any{
		"gold": log.Coin, "goldBalance": log.Coin, "knowledgeToken": log.EduToken, "knowledgePoints": log.KnowledgePoints,
		"goldHistory": log.GoldHist, "guildContrib": w.GuildContrib, "merchants": merchants,
		"professions": profs, "recipes": recipes, "materials": mats, "codex": codex,
		"queue": log.CraftQueue, "activeGather": log.ActiveGather, "activeCraft": log.ActiveCraft,
		"resetsLeft": ProfessionResetLimit - log.ProfessionResets, "durability": log.GearDurability,
		"stallOpen": log.StallOpen, "stall": log.StallListings, "stalls": stalls, "orders": orders,
		"workshop": workshop, "festival": festival, "festivalName": "FESTIVAL KARYA",
		"economy": map[string]any{
			"goldCreated": w.GoldCreated, "goldSpent": w.GoldRemoved, "goldRemoved": w.GoldRemoved,
			"itemsGenerated": w.CraftVolume, "itemsDestroyed": w.GoldRemoved, "tradeVolume": w.TradeVolume, "craftVolume": w.CraftVolume,
		},
		"stackMax": MaterialStackMax, "stackCap": MaterialStackCap, "toId": p.ID,
		"language": log.Language, "achievements": log.Achievements, "titles": log.Titles,
	}
}
