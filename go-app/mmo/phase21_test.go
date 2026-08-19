package mmo

import (
	"encoding/json"
	"testing"
)

func TestPhase21LandminesKept(t *testing.T) {
	if questionByID["q-add-4-3"].ID == "" || questionByID["q-add-4-3"].Correct != 1 {
		t.Fatal("keep q-add-4-3")
	}
	if questionByID["q-iwak-2-3"].Correct != 1 {
		t.Fatal("fishing question B = 5")
	}
	if questionByID["q-add-2-2"].Correct != 1 {
		t.Fatal("2+2 correct index")
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatal("keep Dawn City")
	}
	if mountByID["wind-runner"].Name != "Wind Runner" {
		t.Fatal("keep wind-runner")
	}
	if shopCatalog.ID != "dawn-merchant" {
		t.Fatal("shops.json shape")
	}
	if eventByID["mount-festival"].ID == "" {
		t.Fatal("keep mount-festival")
	}
	if eventByID["festival-karya"].Name != "FESTIVAL KARYA" {
		t.Fatal("festival karya")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
}

func TestPhase21GatherAndNodeCooldown(t *testing.T) {
	w, p := testVillagePlayer()
	node := nodeByID["node-stone-1"]
	p.X, p.Z = node.X, node.Z
	raw, _ := json.Marshal(map[string]any{"nodeId": node.ID, "transactionId": "g1"})
	evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: raw})
	if rejectAction(evs, TypeGather) {
		t.Fatal("gather first")
	}
	if p.materialCount("mat-valley-stone") != 1 {
		t.Fatalf("bag %d", p.materialCount("mat-valley-stone"))
	}
	evs = w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: raw})
	if rejectAction(evs, TypeGather) {
		t.Fatal("same tx replay")
	}
	if p.materialCount("mat-valley-stone") != 1 {
		t.Fatal("dup gather")
	}
	evs = w.ApplyEconomy(p.ID, Envelope{Type: TypeGather, Data: []byte(`{"nodeId":"node-stone-1","transactionId":"g2"}`)})
	if !rejectAction(evs, TypeGather) {
		t.Fatal("respawn reject")
	}
	if p.ensureLog().ProfessionXP[ProfMiner] <= 0 {
		t.Fatal("profession xp")
	}
}

func TestPhase21CraftAtomicAndLocked(t *testing.T) {
	w, p := testVillagePlayer()
	st := interactByID["station-forge"]
	p.X, p.Z = st.X, st.Z
	p.activateProfession(ProfBlacksmith)
	p.ensurePhase21().ProfessionXP[ProfBlacksmith] = 80
	evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-dawn-staff","stationId":"station-forge","transactionId":"c0"}`)})
	if !rejectAction(evs, TypeCraft) {
		t.Fatal("craft without material")
	}
	p.addMaterial("mat-mistwood", 2)
	p.addMaterial("mat-valley-stone", 1)
	p.ensureLog().Coin = 20
	evs = w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-dawn-staff","stationId":"station-forge","transactionId":"c1"}`)})
	if rejectAction(evs, TypeCraft) {
		t.Fatal("craft")
	}
	if p.materialCount("mat-mistwood") != 0 || p.Bag.count("dawn_staff") != 1 {
		t.Fatalf("craft result wood=%d staff=%d", p.materialCount("mat-mistwood"), p.Bag.count("dawn_staff"))
	}
	p.ensurePhase21().Recipes["rec-dawn-ring"] = RecipeLocked
	p.addMaterial("mat-iron-ore", 1)
	p.addMaterial("mat-relic-fragment", 1)
	p.X, p.Z = interactByID["station-bench"].X, interactByID["station-bench"].Z
	evs = w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-dawn-ring","stationId":"station-bench","transactionId":"c2"}`)})
	if !rejectAction(evs, TypeCraft) {
		t.Fatal("locked recipe")
	}
	w.Remove(p)
	again := testOnline("p_a", "Raka")
	w.Add(again)
	if again.Bag.count("dawn_staff") != 1 {
		t.Fatalf("reconnect staff %d", again.Bag.count("dawn_staff"))
	}
}

func TestPhase21FishingAndEduToken(t *testing.T) {
	w, p := testVillagePlayer()
	spot := spotByID["spot-village"]
	p.X, p.Z = spot.X, spot.Z
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeFishStart, Data: []byte(`{"spotId":"spot-village"}`)}), TypeFishStart) {
		t.Fatal("fish start")
	}
	evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeFishCatch, Data: []byte(`{"spotId":"spot-village","progress":50}`)})
	if rejectAction(evs, TypeFishCatch) {
		t.Fatal("fish catch")
	}
	if p.materialCount("mat-river-fish") != 1 {
		t.Fatal("fish reward")
	}
	tok := p.ensureLog().EduToken
	raw, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-iwak-2-3", Choice: 1})
	evs = w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: raw})
	if rejectAction(evs, TypeEducationAnswer) {
		t.Fatal("edu fish")
	}
	if p.ensureLog().EduToken <= tok {
		t.Fatal("knowledge token")
	}
}

func TestPhase21NpcShopBuySell(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().Coin = 40
	buy := w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopBuy, Data: []byte(`{"shopId":"mbah-karya-shop","itemId":"mat-valley-stone","transactionId":"nb1"}`)})
	if rejectAction(buy, TypeNpcShopBuy) {
		t.Fatal("buy")
	}
	if p.ensureLog().Coin != 36 || p.materialCount("mat-valley-stone") != 1 {
		t.Fatalf("buy gold=%d mat=%d", p.ensureLog().Coin, p.materialCount("mat-valley-stone"))
	}
	w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopBuy, Data: []byte(`{"shopId":"mbah-karya-shop","itemId":"mat-valley-stone","transactionId":"nb1"}`)})
	if p.ensureLog().Coin != 36 {
		t.Fatal("buy spam")
	}
	sell := w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopSell, Data: []byte(`{"shopId":"mbah-karya-shop","itemId":"mat-valley-stone","qty":1,"transactionId":"ns1"}`)})
	if rejectAction(sell, TypeNpcShopSell) {
		t.Fatal("sell")
	}
	if p.materialCount("mat-valley-stone") != 0 || p.ensureLog().Coin != 40 {
		t.Fatalf("sell gold=%d mat=%d", p.ensureLog().Coin, p.materialCount("mat-valley-stone"))
	}
	w.ApplyEconomy(p.ID, Envelope{Type: TypeNpcShopSell, Data: []byte(`{"shopId":"mbah-karya-shop","itemId":"mat-valley-stone","qty":1,"transactionId":"ns1"}`)})
	if p.ensureLog().Coin != 40 {
		t.Fatal("sell spam")
	}
}

func TestPhase21TradeLockAfterConfirm(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_b", "Sinta")
	a.X, a.Z = 0, 4
	b.X, b.Z = 0.4, 4
	if !w.giveItem(a, "potion_heal", 1) {
		t.Fatal("potion")
	}
	slot := a.Bag.firstIndex("potion_heal")
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeAccept, Data: []byte(`{}`)})
	off, _ := json.Marshal(map[string]any{"slots": []map[string]any{{"slot": slot, "itemId": "potion_heal", "qty": 1}}, "coin": 0})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeOffer, Data: off})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeReady, Data: []byte(`{}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeReady, Data: []byte(`{}`)})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tr-lock"}`)})
	evs := w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeOffer, Data: off})
	if !rejectAction(evs, TypeTradeOffer) {
		t.Fatal("offer after confirm")
	}
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tr-lock"}`)})
	if b.Bag.count("potion_heal") < 1 {
		t.Fatal("trade deliver")
	}
}

func TestPhase21SecurityRejects(t *testing.T) {
	w, p := testVillagePlayer()
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeAddGold, Data: []byte(`{"amount":99}`)}), TypeAddGold) {
		t.Fatal("addGold")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeAddMaterial, Data: []byte(`{"itemId":"mat-mistwood"}`)}), TypeAddMaterial) {
		t.Fatal("addMaterial")
	}
	if !rejectAction(w.ApplyShop(p.ID, Envelope{Type: TypeSetCoin, Data: []byte(`{"coin":999}`)}), TypeSetCoin) {
		t.Fatal("setCoin")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGetEconomy, Data: []byte(`{}`)}), TypeGetEconomy) {
		t.Fatal("getEconomy admin")
	}
}

func TestPhase21CookingAndProfessionLimit(t *testing.T) {
	w, p := testVillagePlayer()
	st := interactByID["station-cook"]
	p.X, p.Z = st.X, st.Z
	p.addMaterial("mat-dawn-berry", 1)
	p.addMaterial("mat-forest-herb", 1)
	p.ensureLog().Coin = 10
	evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeCraft, Data: []byte(`{"recipeId":"rec-madu-hutan","stationId":"station-cook","transactionId":"food1"}`)})
	if rejectAction(evs, TypeCraft) {
		t.Fatal("cook")
	}
	if p.Bag.count("food_madu_hutan") != 1 {
		t.Fatal("food")
	}
	for _, id := range []string{ProfMiner, ProfWoodcutter, ProfHerbalist, ProfFisher} {
		w.ApplyEconomy(p.ID, Envelope{Type: TypeSetProfession, Data: []byte(`{"professionId":"` + id + `"}`)})
	}
	if len(p.ensurePhase21().ActiveGather) != 3 {
		t.Fatalf("gather cap %d", len(p.ensurePhase21().ActiveGather))
	}
}

func TestPhase21StallAndWorkshop(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_stall", "Bima")
	a.Level = 10
	a.ensureLog().Coin = 1000
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Karya Guild","tag":"KARYA"}`)})
	if rejectAction(w.ApplyEconomy(a.ID, Envelope{Type: TypeGetWorkshop, Data: []byte(`{}`)}), TypeGetWorkshop) {
		t.Fatal("workshop")
	}
	if g := w.guildOf(a.ID); g == nil || !g.Workshop {
		t.Fatal("guild workshop")
	}
	a.addMaterial("mat-mistwood", 1)
	b.ensureLog().Coin = 20
	if rejectAction(w.ApplyEconomy(a.ID, Envelope{Type: TypeStallList, Data: []byte(`{"itemId":"mat-mistwood","qty":1,"price":5}`)}), TypeStallList) {
		t.Fatal("stall list")
	}
	evs := w.ApplyEconomy(b.ID, Envelope{Type: TypeStallBuy, Data: []byte(`{"sellerId":"p_a","itemId":"mat-mistwood","transactionId":"st1"}`)})
	if rejectAction(evs, TypeStallBuy) {
		t.Fatal("stall buy")
	}
	if b.materialCount("mat-mistwood") != 1 {
		t.Fatal("stall deliver")
	}
}
