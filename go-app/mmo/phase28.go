package mmo

import (
	"time"
)

// Phase 28 overlay: Economy + gathering + crafting + merchant.
// Reuses EconomyService, CurrencyService, CraftingService, GatheringService,
// ResourceService, MerchantService (NPC shops), InventoryService, ItemService,
// EquipmentService, QuestService, GuildService, WorldService, TradeService.
// Do not duplicate those services.
//
// Logical tables on PlayerLog / WorldState / catalogs:
// currencies, currency_transactions, resource_nodes, resource_states,
// resource_gather_logs, professions, profession_xp, tools, tool_instances,
// crafting_stations, crafting_recipes, recipe_materials, recipe_unlocks,
// crafting_history, merchants, merchant_items, merchant_prices, economy_logs,
// guild_contributions.
// Indexes: playerId, resourceId, nodeId, recipeId, merchantId, transactionId.

const (
	GoldMaxBalance              = 10000000
	DailyGoldLimit              = 0 // 0 = unlimited prototype
	GoldGenMult                 = 1.0
	ShopPriceMult               = 1.0
	CraftFeeMult                = 1.0
	Phase28RareRespawn          = 24
	Phase28GuildQuestResourceID = "gq-bahan-1"
)

func init() {
	registerPhase28()
}

func registerPhase28() {
	registerPhase28Resources()
	registerPhase28Items()
	registerPhase28Recipes()
	registerPhase28Merchant()
	registerPhase28Quests()
	registerPhase28Meta()
}

func registerPhase28Resources() {
	addRes := func(r ResourceDef) {
		if resourceByID[r.ID].ID != "" {
			return
		}
		resourceCatalog = append(resourceCatalog, r)
		resourceByID[r.ID] = r
		registerItem(ItemDef{
			ID: r.ID, Name: r.Name, Description: r.NameJV, Type: "MATERIAL",
			Rarity: r.Rarity, Stackable: true, MaxStack: MaterialStackMax, Icon: "ore", Value: 3,
		})
	}
	addRes(ResourceDef{ID: "sacred_wood", Name: "Sacred Wood", Type: "WOOD", Region: "forest", Rarity: "RARE", NameJV: "Kayu suci saka alas fajar.", RespawnSec: Phase28RareRespawn, GatherSec: 2})
	addRes(ResourceDef{ID: "dawn_herb", Name: "Dawn Herb", Type: "HERB", Region: "village", Rarity: "COMMON", NameJV: "Godhong esuk desa.", RespawnSec: 8, GatherSec: 1})
	addRes(ResourceDef{ID: "moon_fiber", Name: "Moon Fiber", Type: "FIBER", Region: "plains", Rarity: "UNCOMMON", NameJV: "Serat rembulan.", RespawnSec: 12, GatherSec: 1})
	addRes(ResourceDef{ID: "spirit_stone", Name: "Spirit Stone", Type: "STONE", Region: "ruins", Rarity: "RARE", NameJV: "Watu roh saka reruntuhan.", RespawnSec: Phase28RareRespawn, GatherSec: 2})
	addRes(ResourceDef{ID: "blue_river_fish", Name: "Blue River Fish", Type: "FISH", Region: "plains", Rarity: "UNCOMMON", NameJV: "Iwak biru saka kali cahaya.", RespawnSec: 10, GatherSec: 2})
	addRes(ResourceDef{ID: "fish_common", Name: "Common Fish", Type: "FISH", Region: "plains", Rarity: "COMMON", NameJV: "Iwak biasa.", RespawnSec: 8, GatherSec: 2})
	addRes(ResourceDef{ID: "fish_rare", Name: "Rare Fish", Type: "FISH", Region: "plains", Rarity: "RARE", NameJV: "Iwak langka.", RespawnSec: 14, GatherSec: 2})
	addRes(ResourceDef{ID: "fish_epic", Name: "Epic Fish", Type: "FISH", Region: "plains", Rarity: "EPIC", NameJV: "Iwak agung.", RespawnSec: 20, GatherSec: 2})
	addNode := func(n ResourceNodeDef) {
		if nodeByID[n.ID].ID != "" {
			return
		}
		nodeCatalog = append(nodeCatalog, n)
		nodeByID[n.ID] = n
		name := resourceByID[n.ResourceID].Name
		registerInteract(InteractDef{ID: n.ID, Kind: n.Kind, X: n.X, Z: n.Z, Text: name})
	}
	addNode(ResourceNodeDef{ID: "node-sacred-1", ResourceID: "sacred_wood", Kind: "gather-wood", X: -6, Z: 26})
	addNode(ResourceNodeDef{ID: "node-dawnherb-1", ResourceID: "dawn_herb", Kind: "gather-herb", X: 2.4, Z: 6.2})
	addNode(ResourceNodeDef{ID: "node-moonfiber-1", ResourceID: "moon_fiber", Kind: "gather-fiber", X: -5, Z: 82})
	addNode(ResourceNodeDef{ID: "node-spirit-1", ResourceID: "spirit_stone", Kind: "gather-stone", X: 2, Z: 132})
	registerInteract(InteractDef{ID: "station-alchemy", Kind: "alchemy", X: -5.4, Z: 6.2, Text: "ALCHEMY TABLE"})
}

func registerPhase28Items() {
	tool := func(id, name, desc string) {
		registerItem(ItemDef{ID: id, Name: name, Description: desc, Type: "UTILITY", Rarity: "COMMON", Icon: "gear", Value: 6, MaxStack: 1})
	}
	tool("tool_pickaxe", "Pickaxe", "Alat tambang. Architecture durability siap.")
	tool("tool_axe", "Axe", "Kapak tebang kayu.")
	tool("tool_rod", "Fishing Rod", "Pancing desa.")
	tool("tool_knife", "Gathering Knife", "Pisau herbalis.")
	registerItem(ItemDef{ID: "iron_ingot", Name: "Iron Ingot", Description: "Hasil olah bijih. Material processing.", Type: "MATERIAL", Rarity: "UNCOMMON", Stackable: true, MaxStack: 99, Icon: "ore", Value: 6})
	registerItem(ItemDef{ID: "potion_dawn", Name: "Dawn Potion", Description: "Ramuan alchemy desa.", Type: "CONSUMABLE", Rarity: "UNCOMMON", Stackable: true, MaxStack: 20, Icon: "potion", Value: 8, Effects: ItemEffects{HealPct: 0.12}})
}

func registerPhase28Recipes() {
	addRec := func(r RecipeDef) {
		if recipeByID[r.ID].ID != "" {
			return
		}
		recipeCatalog = append(recipeCatalog, r)
		recipeByID[r.ID] = r
	}
	addRec(RecipeDef{
		ID: "rec-dawn-blade", Name: "Dawn Blade", Profession: ProfBlacksmith, Category: "WEAPON", Result: "dawn_blade",
		RequiredLevel: 1, CraftSec: 0, Discover: "start",
		Materials: []RecipeMat{{ItemID: "mat-iron-ore", Qty: 10}, {ItemID: "sacred_wood", Qty: 5}, {ItemID: "spirit_stone", Qty: 2}},
	})
	addRec(RecipeDef{
		ID: "rec-dawn-potion", Name: "Dawn Potion", Profession: ProfCook, Category: "POTION", Result: "potion_dawn",
		RequiredLevel: 1, CraftSec: 0, Discover: "start",
		Materials: []RecipeMat{{ItemID: "dawn_herb", Qty: 2}},
	})
	addRec(RecipeDef{
		ID: "rec-iron-ingot", Name: "Iron Ingot", Profession: ProfBlacksmith, Category: "MATERIAL", Result: "iron_ingot",
		RequiredLevel: 1, CraftSec: 0, Discover: "start",
		Materials: []RecipeMat{{ItemID: "mat-iron-ore", Qty: 2}},
	})
}

func registerPhase28Merchant() {
	registerNPC(NPCDef{
		ID: "pak_dagang", Name: "Pak Dagang", Role: "Pedagang", Type: "PEDAGANG",
		X: 1.6, Z: 5.4, Yaw: 2.4, DialogueID: "pak_dagang", QuestIDs: []string{"pq-craft-1", "pq-gather-1"}, InteractionRange: 2.6,
	})
	dialogueCatalog["pak_dagang"] = DialogueLine{Speaker: "Pak Dagang", Text: "monggo, Le. Pilihen barang sing kok butuhake."}
	npcShopByID["pedagang-shop"] = NpcShopDef{
		ID: "pedagang-shop", Name: "Toko Pedagang",
		Items: []NpcShopItem{
			{ItemID: "mat-iron-ore", Price: 10, SellPrice: 4, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "sacred_wood", Price: 12, SellPrice: 5, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "dawn_herb", Price: 6, SellPrice: 2, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "potion_heal", Price: 8, SellPrice: 3, Currency: "gold", Buy: true, Sell: true},
			{ItemID: "knowledge_token", Price: 99, SellPrice: 1, Currency: "gold", Buy: false, Sell: true},
			{ItemID: "quest_dawn_note", Price: 99, SellPrice: 1, Currency: "gold", Buy: false, Sell: true},
		},
	}
}

func registerPhase28Quests() {
	registerQuestion(QuestionDef{
		ID: "q-kayu-3-2", Category: "Matematika",
		Prompt: "Yen duwe 3 kayu lan entuk 2 kayu maneh, dadi pira?",
		Choices: []string{"4", "5", "6"}, Correct: 1,
		Explain: "3 ditambah 2 sama dengan 5.", Grade: 1,
	})
	registerQuestion(QuestionDef{
		ID: "q-emas-15-10", Category: "Matematika",
		Prompt: "Yen regane barang 10 emas lan kowe duwe 15 emas, sisane pira?",
		Choices: []string{"3", "5", "10"}, Correct: 1,
		Explain: "15 dikurangi 10 sama dengan 5.", Grade: 1,
	})
	registerQuest(QuestDef{
		ID: "pq-craft-1", Title: "Sinau Nggawé.", Kind: "profession", NPC: "pak_dagang",
		Location: "Dawn Village", Description: "Yen kepengin maju, kowe kudu sinau nggawe barang dhewe.",
		Objectives: []ObjectiveDef{{Type: "CRAFT", Target: "dawn_blade", Count: 1, Text: "Craft Dawn Blade"}},
		Rewards:    RewardDef{Exp: 50, Coin: 12, EduToken: 1, Knowledge: 1},
		FlagsOnClaim: []string{"pq_craft_done"}, ClaimAt: "pak_dagang",
	})
	registerQuest(QuestDef{
		ID: "pq-gather-1", Title: "Ngumpulake Bahan.", Kind: "gather", NPC: "pak_dagang",
		Location: "Alas lan Lembah", Description: "Kumpulkan Sacred Wood, Iron Ore, Dawn Herb.",
		Objectives: []ObjectiveDef{
			{Type: "GATHER", Target: "sacred_wood", Count: 5, Text: "5 Sacred Wood"},
			{Type: "GATHER", Target: "mat-iron-ore", Count: 5, Text: "5 Iron Ore"},
			{Type: "GATHER", Target: "dawn_herb", Count: 3, Text: "3 Dawn Herb"},
		},
		Rewards:      RewardDef{Exp: 40, Coin: 10},
		FlagsOnClaim: []string{"pq_gather_done"}, ClaimAt: "pak_dagang",
	})
	registerQuest(QuestDef{
		ID: "pq-fish-1", Title: "Nelayan Cilik.", Kind: "gather", NPC: "pak_jala",
		Location: "Kolam Desa", Description: "Tangkap 3 iwak.",
		Objectives: []ObjectiveDef{{Type: "GATHER", Target: "mat-river-fish", Count: 3, Text: "Catch 3 fish"}},
		Rewards:      RewardDef{Exp: 30, Coin: 8},
		FlagsOnClaim: []string{"pq_fish_done"}, ClaimAt: "pak_jala",
	})
	registerQuest(QuestDef{
		ID: Phase28GuildQuestResourceID, Title: "Bahan Desa.", Kind: "guild", NPC: "pak_dagang",
		Location: "Guild", Description: "Sumbang bahan untuk gudang guild.",
		Objectives: []ObjectiveDef{{Type: "GATHER", Target: "mat-iron-ore", Count: 10, Text: "Sumbang Iron Ore"}},
		Rewards:      RewardDef{Exp: 20, Coin: 6},
		FlagsOnClaim: []string{"gq_bahan_done"},
	})
}

func registerPhase28Meta() {
	registerAchievementDef(AchievementDef{ID: "first-fish", Name: "First Fish", Title: "fisher", Flag: "ach_first_fish", Category: "Profession"})
	registerAchievementDef(AchievementDef{ID: "master-miner", Name: "Master Miner", Title: "gatherer", Flag: "ach_master_miner", Category: "Profession"})
	registerAchievementDef(AchievementDef{ID: "master-fisher", Name: "Master Fisher", Title: "fisher", Flag: "ach_master_fisher", Category: "Profession"})
	registerAchievementDef(AchievementDef{ID: "master-crafter", Name: "Master Crafter", Title: "craftsman", Flag: "ach_master_crafter", Category: "Profession"})
	registerTitleDef(TitleDef{ID: "master-crafter", Name: "Master Crafter", Source: "profession", Rarity: "RARE"})
	registerCosmeticDef(CosmeticDef{ID: "badge-crafter", Name: "Crafter Badge", Kind: "badge"})
}

func (w *WorldState) ApplyPhase28(p *Player, env Envelope) [][]byte {
	if p == nil {
		return nil
	}
	switch env.Type {
	case TypeSetPrice, TypeSetGold, TypeGiveGold, TypeCreateRecipe:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	case TypeGuildContribute:
		return w.guildContributeMaterial(p, env.Data)
	case TypeGetGoldLog:
		return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
	case TypeAddGold, TypeAddMaterial:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) phase28GatherGate(p *Player, node ResourceNodeDef) [][]byte {
	if p.inCombatNow() {
		return rejectFor(p.ID, TypeGather, "combat")
	}
	if w.limited("gather:"+p.ID, 8, 2*time.Second) {
		return rejectFor(p.ID, TypeGather, "cooldown")
	}
	res := resourceByID[node.ResourceID]
	if res.Rarity == "RARE" || res.Rarity == "EPIC" {
		need := toolFor(res.Type)
		if need != "" && p.Bag.count(need) < 1 && p.materialCount(need) < 1 {
			return rejectFor(p.ID, TypeGather, "tool")
		}
	}
	return nil
}

func (w *WorldState) phase28AfterGather(p *Player, node ResourceNodeDef, res ResourceDef) {
	log := p.ensureLog()
	if log.NodeFound == nil {
		log.NodeFound = map[string]bool{}
	}
	log.NodeFound[node.ID] = true
	if professionLevel(log.ProfessionXP[ProfMiner]) >= 20 {
		log.Flags["ach_master_miner"] = true
	}
	_ = res
}

func (w *WorldState) phase28AfterFish(p *Player, progress int) {
	log := p.ensurePhase21()
	log.Flags["ach_first_fish"] = true
	if professionLevel(log.ProfessionXP[ProfFisher]) >= 20 {
		log.Flags["ach_master_fisher"] = true
	}
	mid := (log.FishTargetA + log.FishTargetB) / 2
	id := "fish_common"
	if absInt(progress-mid) <= 4 {
		id = "fish_epic"
	} else if absInt(progress-mid) <= 10 {
		id = "fish_rare"
	}
	p.addMaterial(id, 1)
}

func (w *WorldState) phase28AfterCraft(p *Player, recipeID string) {
	log := p.ensureLog()
	if professionLevel(log.ProfessionXP[ProfBlacksmith]) >= 20 || professionLevel(log.ProfessionXP[ProfArtisan]) >= 20 {
		log.Flags["ach_master_crafter"] = true
		p.grantTitle("master-crafter")
		p.grantCosmetic("badge-crafter")
	}
	if recipeID == "rec-dawn-blade" && !log.Flags["craft_edu_done"] {
		log.Flags["craft_edu_done"] = true
		log.Quiz = QuizSession{QuestID: "craft-edu", Active: true}
	}
	p.credit("CRAFT", recipeID, 1)
}

func phase28SellBlocked(itemID string) string {
	def := itemByID[itemID]
	if def.Untradable {
		return "soulbound"
	}
	if def.Type == "QUEST" || def.Type == "QUEST_ITEM" {
		return "quest"
	}
	return ""
}

func shopSellGain(it NpcShopItem) int {
	if it.SellPrice > 0 && it.SellPrice < it.Price {
		return int(float64(it.SellPrice)*ShopPriceMult + 0.5)
	}
	if it.SellPrice > 0 {
		return it.SellPrice
	}
	return it.Price
}

func shopBuyCost(it NpcShopItem) int {
	n := int(float64(it.Price)*ShopPriceMult + 0.5)
	if n < 1 {
		n = 1
	}
	return n
}

func craftFeeNow() int {
	mult := CraftFeeMult
	n := int(float64(CraftFeeGold)*mult + 0.5)
	if n < 0 {
		n = 0
	}
	return n
}

func toolFor(resType string) string {
	switch resType {
	case "ORE", "STONE":
		return "tool_pickaxe"
	case "WOOD":
		return "tool_axe"
	case "FISH":
		return "tool_rod"
	case "HERB", "FIBER", "FRUIT":
		return "tool_knife"
	}
	return ""
}

func capGold(log *PlayerLog) {
	if log == nil {
		return
	}
	if log.Coin > GoldMaxBalance {
		log.Coin = GoldMaxBalance
	}
	if DailyGoldLimit > 0 && log.GoldToday > DailyGoldLimit {
		overflow := log.GoldToday - DailyGoldLimit
		log.Coin -= overflow
		if log.Coin < 0 {
			log.Coin = 0
		}
		log.GoldToday = DailyGoldLimit
	}
}

func (p *Player) recordGoldHist(src string, amount int, dest string) {
	log := p.ensureLog()
	log.GoldHist = append(log.GoldHist, GoldHistRow{Source: src, Amount: amount, Dest: dest, At: time.Now().UnixMilli()})
	if len(log.GoldHist) > 40 {
		log.GoldHist = log.GoldHist[len(log.GoldHist)-30:]
	}
}

func (w *WorldState) guildContributeMaterial(p *Player, data []byte) [][]byte {
	var in struct {
		ItemID string
		Qty    int
	}
	_ = unmarshal(data, &in)
	if in.Qty < 1 {
		in.Qty = 1
	}
	g := w.guildOf(p.ID)
	if g == nil {
		return rejectFor(p.ID, TypeGuildContribute, "guild")
	}
	if p.materialCount(in.ItemID) < in.Qty && p.Bag.count(in.ItemID) < in.Qty {
		return rejectFor(p.ID, TypeGuildContribute, "owned")
	}
	if p.materialCount(in.ItemID) >= in.Qty {
		p.takeMaterial(in.ItemID, in.Qty)
	} else if !p.Bag.takeItem(in.ItemID, in.Qty) {
		return rejectFor(p.ID, TypeGuildContribute, "owned")
	}
	w.GuildContrib = append(w.GuildContrib, GuildContribRow{
		PlayerID: p.ID, ItemID: in.ItemID, Qty: in.Qty, At: time.Now().UnixMilli(),
	})
	if len(w.GuildContrib) > 80 {
		w.GuildContrib = w.GuildContrib[len(w.GuildContrib)-50:]
	}
	w.guildContribute(p, in.Qty, 2)
	p.credit("GATHER", in.ItemID, in.Qty)
	w.persist(p)
	return [][]byte{marshal(TypeCraftingState, w.craftingView(p))}
}

func professionLabel(id, fallback string) string {
	switch id {
	case ProfMiner:
		return "Mining"
	case ProfWoodcutter:
		return "Woodcutting"
	case ProfHerbalist:
		return "Herbalism"
	case ProfFisher:
		return "Fishing"
	default:
		return fallback
	}
}

func nextProfessionXP(xp int) int {
	lv, need, left := 1, 40, xp
	for lv < ProfessionMaxLevel && left >= need {
		left -= need
		lv++
		need += 12
	}
	return need - left
}

type GoldHistRow struct {
	Source string `json:"source"`
	Amount int    `json:"amount"`
	Dest   string `json:"destination"`
	At     int64  `json:"timestamp"`
}

type GuildContribRow struct {
	PlayerID string `json:"playerId"`
	ItemID   string `json:"itemId"`
	Qty      int    `json:"qty"`
	At       int64  `json:"at"`
}

func merchantInteract(p *Player) [][]byte {
	opts := []DialogOption{
		{ID: "npc-shop:pedagang-shop", Label: "Buka toko"},
		{ID: "quiz-econ", Label: "Tanya harga"},
	}
	opts = append(opts, questOption(p, "pq-craft-1", "Sinau Nggawé.")...)
	opts = append(opts, questOption(p, "pq-gather-1", "Ngumpulake Bahan.")...)
	opts = append(opts, DialogOption{ID: "close", Label: "Tutup"})
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "npc", TargetID: "pak_dagang", Title: "Pak Dagang", Speaker: "Pak Dagang",
		Text: "monggo, Le. Pilihen barang sing kok butuhake.",
		Subtitle: "Silakan, Nak. Pilih barang yang kamu butuhkan.",
		Options: opts,
	})}
}

func (w *WorldState) startEconQuiz(p *Player) [][]byte {
	q := questionByID["q-emas-15-10"]
	p.ensureLog().Quiz = QuizSession{QuestID: "econ-edu", Active: true}
	return [][]byte{marshal(TypeEducationQuestion, QuestionOut{
		ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices,
	})}
}

func (w *WorldState) startCraftQuiz(p *Player) [][]byte {
	q := questionByID["q-kayu-3-2"]
	p.ensureLog().Quiz = QuizSession{QuestID: "craft-edu", Active: true}
	return [][]byte{marshal(TypeEducationQuestion, QuestionOut{
		ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices,
	})}
}
