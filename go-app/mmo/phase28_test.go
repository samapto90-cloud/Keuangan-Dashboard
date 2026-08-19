package mmo

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPhase28LandminesKept(t *testing.T) {
	if comboRecipe() != "LLLH" {
		t.Fatalf("combo %s", comboRecipe())
	}
	if transformByID["aura-1"].Name != "AURA ASCENSION I" || formDisplayName(transformByID["aura-1"]) != "AWAKENED FORM" {
		t.Fatal("keep AURA ASCENSION I / AWAKENED FORM")
	}
	if questionByID["q-add-4-3"].Correct != 1 || questionByID["q-apel-3-2"].Correct != 1 {
		t.Fatal("keep q-add-4-3 and q-apel-3-2")
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatal("keep Dawn City")
	}
	if mountByID["wind-runner"].Name != "Wind Runner" || mountByID["wind-runner"].Speed != 1.45 {
		t.Fatal("keep wind-runner")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
	if itemByID["dawn_blade"].Name != "DAWN BLADE" {
		t.Fatal("keep DAWN BLADE")
	}
	if itemByID["mat-iron-ore"].Name != "Iron Ore" {
		t.Fatal("keep Iron Ore")
	}
	if questByID["pq-mine-1"].Title != "Sinau Nambang." {
		t.Fatal("keep Sinau Nambang.")
	}
	if questByID["pq-edu-1"].Title != "Sinau Ngetung." {
		t.Fatal("keep Sinau Ngetung.")
	}
	if questionByID["q-iwak-2-3"].Correct != 1 {
		t.Fatal("keep fishing edu")
	}
	if GoldGenMult != 1.0 || ShopPriceMult != 1.0 || CraftFeeMult != 1.0 {
		t.Fatal("phase 21 price multipliers")
	}
}

func TestPhase28CatalogAndQuests(t *testing.T) {
	if resourceByID["sacred_wood"].Name != "Sacred Wood" || resourceByID["dawn_herb"].Name != "Dawn Herb" {
		t.Fatal("new resources")
	}
	if resourceByID["moon_fiber"].Name != "Moon Fiber" || resourceByID["spirit_stone"].Name != "Spirit Stone" {
		t.Fatal("fiber/stone")
	}
	if resourceByID["blue_river_fish"].Name != "Blue River Fish" {
		t.Fatal("blue river fish")
	}
	rec := recipeByID["rec-dawn-blade"]
	if rec.Result != "dawn_blade" || len(rec.Materials) != 3 {
		t.Fatal("dawn blade recipe")
	}
	if rec.Materials[0].ItemID != "mat-iron-ore" || rec.Materials[0].Qty != 10 {
		t.Fatal("iron ore x10")
	}
	if rec.Materials[1].ItemID != "sacred_wood" || rec.Materials[1].Qty != 5 {
		t.Fatal("sacred wood x5")
	}
	if rec.Materials[2].ItemID != "spirit_stone" || rec.Materials[2].Qty != 2 {
		t.Fatal("spirit stone x2")
	}
	if dialogueCatalog["pak_dagang"].Text != "monggo, Le. Pilihen barang sing kok butuhake." {
		t.Fatal("pak dagang dialogue")
	}
	if npcByID["pak_dagang"].Type != "PEDAGANG" {
		t.Fatal("pedagang type")
	}
	if questByID["pq-craft-1"].Title != "Sinau Nggawé." || questByID["pq-gather-1"].Title != "Ngumpulake Bahan." {
		t.Fatal("craft/gather quest")
	}
	if questByID["pq-fish-1"].Title != "Nelayan Cilik." {
		t.Fatal("fish quest")
	}
	if questionByID["q-kayu-3-2"].Correct != 1 || questionByID["q-emas-15-10"].Correct != 1 {
		t.Fatal("edu answers B")
	}
	if ProfessionMaxLevel != 50 || GoldMaxBalance != 10000000 {
		t.Fatal("caps")
	}
	it := npcShopByID["pedagang-shop"].Items[0]
	if it.SellPrice >= it.Price {
		t.Fatal("sell below buy")
	}
}

func TestPhase28GatherIronAndRareTool(t *testing.T) {
	w, p := testVillagePlayer()
	ore := nodeByID["node-ore-1"]
	p.X, p.Z = ore.X, ore.Z
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-ore-1","transactionId":"p28-ore"}`)}), TypeGather) {
		t.Fatal("gather iron ore")
	}
	if p.materialCount("mat-iron-ore") != 1 {
		t.Fatalf("iron %d", p.materialCount("mat-iron-ore"))
	}
	wood := nodeByID["node-sacred-1"]
	p.X, p.Z = wood.X, wood.Z
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-sacred-1","transactionId":"p28-wood0"}`)}), TypeGather) {
		t.Fatal("rare wood needs axe")
	}
	p.Bag.add("tool_axe", 1)
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-sacred-1","transactionId":"p28-wood1"}`)}), TypeGather) {
		t.Fatal("wood with axe")
	}
	if p.materialCount("sacred_wood") != 1 {
		t.Fatal("sacred wood")
	}
}

func TestPhase28GatherCombatAndDistance(t *testing.T) {
	w, p := testVillagePlayer()
	node := nodeByID["node-dawnherb-1"]
	p.X, p.Z = 0, 0
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-dawnherb-1","transactionId":"p28-far"}`)}), TypeGather) {
		t.Fatal("distance")
	}
	p.X, p.Z = node.X, node.Z
	p.InCombatUntil = time.Now().Add(time.Minute)
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-dawnherb-1","transactionId":"p28-combat"}`)}), TypeGather) {
		t.Fatal("combat interrupt")
	}
	p.InCombatUntil = time.Time{}
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-dawnherb-1","transactionId":"p28-herb"}`)}), TypeGather) {
		t.Fatal("herb gather")
	}
	if p.materialCount("dawn_herb") != 1 {
		t.Fatal("dawn herb")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-dawnherb-1","transactionId":"p28-herb2"}`)}), TypeGather) {
		t.Fatal("respawn reject")
	}
	w.NodeCooldown[node.ID] = time.Now().Unix() - 1
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-dawnherb-1","transactionId":"p28-herb3"}`)}), TypeGather) {
		t.Fatal("respawn after timer")
	}
}

func TestPhase28CraftDawnBladeAndFullBag(t *testing.T) {
	w, p := testVillagePlayer()
	st := interactByID["station-forge"]
	p.X, p.Z = st.X, st.Z
	p.activateProfession(ProfBlacksmith)
	p.ensureLog().Coin = 20
	p.addMaterial("mat-iron-ore", 10)
	p.addMaterial("sacred_wood", 5)
	p.addMaterial("spirit_stone", 2)
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-dawn-blade","stationId":"station-forge","transactionId":"p28-c1"}`)}), TypeCraft) {
		t.Fatal("craft dawn blade")
	}
	if p.Bag.count("dawn_blade") != 1 {
		t.Fatal("blade created")
	}
	if p.materialCount("mat-iron-ore") != 0 || p.materialCount("sacred_wood") != 0 || p.materialCount("spirit_stone") != 0 {
		t.Fatal("materials consumed")
	}
	p.addMaterial("mat-iron-ore", 10)
	p.addMaterial("sacred_wood", 5)
	p.addMaterial("spirit_stone", 2)
	p.ensureGear()
	for i := 0; i < bagCapOf(p); i++ {
		if _, ok := p.Bag.add("dawn_blade", 1); !ok {
			break
		}
	}
	ore, wood, stone := p.materialCount("mat-iron-ore"), p.materialCount("sacred_wood"), p.materialCount("spirit_stone")
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-dawn-blade","stationId":"station-forge","transactionId":"p28-full"}`)}), TypeCraft) {
		t.Fatal("full bag craft")
	}
	if p.materialCount("mat-iron-ore") != ore || p.materialCount("sacred_wood") != wood || p.materialCount("spirit_stone") != stone {
		t.Fatal("full bag must not consume")
	}
}

func TestPhase28CraftWithoutMatsAndSpam(t *testing.T) {
	w, p := testVillagePlayer()
	st := interactByID["station-forge"]
	p.X, p.Z = st.X, st.Z
	p.activateProfession(ProfBlacksmith)
	p.ensureLog().Coin = 20
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-dawn-blade","stationId":"station-forge","transactionId":"p28-nom"}`)}), TypeCraft) {
		t.Fatal("no mats")
	}
	limited := false
	for i := 0; i < 8; i++ {
		raw, _ := json.Marshal(map[string]any{"recipeId": "rec-dawn-blade", "stationId": "station-forge", "transactionId": "spam-" + itoa(i)})
		evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: raw})
		if rejectAction(evs, TypeCraft) {
			var env Envelope
			json.Unmarshal(evs[0], &env)
			var r RejectOut
			json.Unmarshal(env.Data, &r)
			if r.Reason == "cooldown" {
				limited = true
			}
		}
	}
	if !limited {
		t.Fatal("craft rate limit")
	}
}

func TestPhase28PedagangBuySellAndSoulbound(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().Coin = 20
	buy := w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopBuy, Data: []byte(`{"shopId":"pedagang-shop","itemId":"mat-iron-ore","transactionId":"p28-b1"}`)})
	if rejectAction(buy, TypeNpcShopBuy) {
		t.Fatal("buy")
	}
	if p.ensureLog().Coin != 10 || p.materialCount("mat-iron-ore") != 1 {
		t.Fatalf("buy gold=%d ore=%d", p.ensureLog().Coin, p.materialCount("mat-iron-ore"))
	}
	w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopBuy, Data: []byte(`{"shopId":"pedagang-shop","itemId":"mat-iron-ore","transactionId":"p28-b1"}`)})
	if p.ensureLog().Coin != 10 {
		t.Fatal("buy tx spam")
	}
	sell := w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopSell, Data: []byte(`{"shopId":"pedagang-shop","itemId":"mat-iron-ore","qty":1,"transactionId":"p28-s1"}`)})
	if rejectAction(sell, TypeNpcShopSell) {
		t.Fatal("sell")
	}
	if p.materialCount("mat-iron-ore") != 0 || p.ensureLog().Coin != 14 {
		t.Fatalf("sell gold=%d ore=%d", p.ensureLog().Coin, p.materialCount("mat-iron-ore"))
	}
	p.Bag.add("knowledge_token", 1)
	p.addMaterial("knowledge_token", 1)
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopSell, Data: []byte(`{"shopId":"pedagang-shop","itemId":"knowledge_token","qty":1,"transactionId":"p28-sb"}`)}), TypeNpcShopSell) {
		t.Fatal("soulbound")
	}
	p.addMaterial("quest_dawn_note", 1)
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopSell, Data: []byte(`{"shopId":"pedagang-shop","itemId":"quest_dawn_note","qty":1,"transactionId":"p28-q"}`)}), TypeNpcShopSell) {
		t.Fatal("quest item")
	}
}

func TestPhase28SecurityRejects(t *testing.T) {
	w, p := testVillagePlayer()
	before := p.ensureLog().Coin
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGiveGold, Data: []byte(`{"amount":99}`)}), TypeGiveGold) {
		t.Fatal("GIVE_GOLD")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeSetGold, Data: []byte(`{"gold":999}`)}), TypeSetGold) {
		t.Fatal("SET_GOLD")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeSetPrice, Data: []byte(`{"price":1}`)}), TypeSetPrice) {
		t.Fatal("SET_PRICE")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeCreateRecipe, Data: []byte(`{}`)}), TypeCreateRecipe) {
		t.Fatal("CREATE_RECIPE")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeAddGold, Data: []byte(`{"amount":9}`)}), TypeAddGold) {
		t.Fatal("ADD_GOLD")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeAddMaterial, Data: []byte(`{"itemId":"sacred_wood"}`)}), TypeAddMaterial) {
		t.Fatal("ADD_MATERIAL")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeSetCoin, Data: []byte(`{"coin":9}`)}), TypeSetCoin) {
		t.Fatal("SET_COIN")
	}
	if p.ensureLog().Coin != before {
		t.Fatal("gold unchanged")
	}
	if p.materialCount("sacred_wood") != 0 {
		t.Fatal("no free mats")
	}
}

func TestPhase28GuildContributeAndEdu(t *testing.T) {
	w, p := testVillagePlayer()
	p.Level = 10
	p.ensureLog().Coin = 1000
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Bahan Desa","tag":"BAHAN"}`)}), TypeGuildCreate) {
		t.Fatal("guild")
	}
	p.addMaterial("mat-iron-ore", 2)
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGuildContribute, Data: []byte(`{"itemId":"mat-iron-ore","qty":1}`)}), TypeGuildContribute) {
		t.Fatal("contribute")
	}
	if p.materialCount("mat-iron-ore") != 1 || len(w.GuildContrib) < 1 {
		t.Fatal("contrib log")
	}
	if w.GuildContrib[0].PlayerID != p.ID || w.GuildContrib[0].Qty != 1 {
		t.Fatal("contrib row")
	}
	evs := w.startEconQuiz(p)
	if len(evs) < 1 {
		t.Fatal("econ quiz")
	}
	raw, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-emas-15-10", Choice: 1})
	ans := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: raw})
	if rejectAction(ans, TypeEducationAnswer) {
		t.Fatal("econ answer")
	}
	if p.Bag.count("knowledge_token") < 1 && p.ensureLog().EduToken < 1 {
		t.Fatal("knowledge token")
	}
	w.startCraftQuiz(p)
	raw, _ = json.Marshal(EducationAnswerIn{QuestionID: "q-kayu-3-2", Choice: 1})
	if rejectAction(w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: raw}), TypeEducationAnswer) {
		t.Fatal("craft edu")
	}
}

func TestPhase28FishingStillRiverFish(t *testing.T) {
	w, p := testVillagePlayer()
	spot := spotByID["spot-village"]
	p.X, p.Z = spot.X, spot.Z
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeFishStart, Data: []byte(`{"spotId":"spot-village"}`)}), TypeFishStart) {
		t.Fatal("fish start")
	}
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeFishCatch, Data: []byte(`{"spotId":"spot-village","progress":50}`)}), TypeFishCatch) {
		t.Fatal("fish catch")
	}
	if p.materialCount("mat-river-fish") != 1 {
		t.Fatal("river fish")
	}
	if p.materialCount("fish_common")+p.materialCount("fish_rare")+p.materialCount("fish_epic") < 1 {
		t.Fatal("rarity fish overlay")
	}
}

func TestPhase28MerchantTalkAndPotionCraft(t *testing.T) {
	w, p := testVillagePlayer()
	npc := npcByID["pak_dagang"]
	p.X, p.Z = npc.X, npc.Z
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"pak_dagang"}`)})
	found := false
	for _, ev := range evs {
		var env Envelope
		json.Unmarshal(ev, &env)
		if env.Type != TypeInteractResult {
			continue
		}
		var res InteractResult
		json.Unmarshal(env.Data, &res)
		if strings.Contains(res.Text, "monggo, Le") && strings.Contains(res.Subtitle, "Silakan, Nak") {
			found = true
		}
	}
	if !found {
		t.Fatal("merchant talk")
	}
	al := interactByID["station-alchemy"]
	p.X, p.Z = al.X, al.Z
	p.activateProfession(ProfCook)
	p.ensureLog().Coin = 5
	p.addMaterial("dawn_herb", 2)
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-dawn-potion","stationId":"station-alchemy","transactionId":"p28-pot"}`)}), TypeCraft) {
		t.Fatal("potion craft")
	}
	if p.Bag.count("potion_dawn") != 1 && p.materialCount("potion_dawn") != 1 {
		t.Fatal("potion result")
	}
}

func TestPhase28IngotAndGoldCap(t *testing.T) {
	w, p := testVillagePlayer()
	st := interactByID["station-forge"]
	p.X, p.Z = st.X, st.Z
	p.activateProfession(ProfBlacksmith)
	p.ensureLog().Coin = 5
	p.addMaterial("mat-iron-ore", 2)
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-iron-ingot","stationId":"station-forge","transactionId":"p28-ing"}`)}), TypeCraft) {
		t.Fatal("ingot")
	}
	if p.materialCount("iron_ingot") != 1 && p.Bag.count("iron_ingot") != 1 {
		t.Fatal("ingot result")
	}
	p.ensureLog().Coin = GoldMaxBalance + 50
	capGold(p.ensureLog())
	if p.ensureLog().Coin != GoldMaxBalance {
		t.Fatal("gold cap")
	}
}
