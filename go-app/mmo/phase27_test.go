package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPhase27LandminesKept(t *testing.T) {
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
	if _, ok := skillCatalog["power_strike"]; !ok {
		t.Fatal("power_strike")
	}
	if _, ok := skillCatalog["celestial_impact"]; !ok {
		t.Fatal("celestial_impact")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
	if questByID["mq004"].Title != "Jalan Menuju Hutan" {
		t.Fatal("mq004")
	}
	if dialogueCatalog["pak_jaga"].Text != "Ati-ati, Le! Ana siluman teka saka alas!" {
		t.Fatal("pak_jaga")
	}
	if len(storyChapterCatalog) != 10 {
		t.Fatalf("chapters %d", len(storyChapterCatalog))
	}
	if silumanByGuard["ragha"].Name != "Jaladara" {
		t.Fatal("Jaladara")
	}
	if itemByID["wind_blade"].Name != "Wind Blade" || itemByID["stone_hammer"].Name != "Stone Hammer" {
		t.Fatal("keep crafted weapon names")
	}
	if InvCapacity != 50 {
		t.Fatalf("default bag %d", InvCapacity)
	}
	if itemByID["light_staff"].Name != "LIGHT STAFF" || itemByID["dawn_blade"].Name != "DAWN BLADE" {
		t.Fatal("original weapons")
	}
	if itemByID["edu_token"].Name != "Education Token" {
		t.Fatal("keep Education Token name")
	}
	if itemByID["knowledge_token"].ID == "" {
		t.Fatal("knowledge token")
	}
}

func TestPhase27ItemSecurityRejected(t *testing.T) {
	w, p := testVillagePlayer()
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeGiveItem, Data: []byte(`{"itemId":"dawn_blade","quantity":1}`)}), TypeGiveItem) {
		t.Fatal("GIVE_ITEM")
	}
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeSetItemStats, Data: []byte(`{"attack":999}`)}), TypeSetItemStats) {
		t.Fatal("SET_ITEM_STATS")
	}
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeDuplicateInstance, Data: []byte(`{}`)}), TypeDuplicateInstance) {
		t.Fatal("DUPLICATE_ITEM")
	}
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeSetInstanceID, Data: []byte(`{}`)}), TypeSetInstanceID) {
		t.Fatal("SET_INSTANCE")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeSetCoin, Data: []byte(`{"coin":9}`)}), TypeSetCoin) {
		t.Fatal("SET_COIN")
	}
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeEquipItem, Data: []byte(`{"slot":99}`)}), TypeEquipItem) {
		t.Fatal("equip missing")
	}
}

func TestPhase27InstanceSplitLockDiscard(t *testing.T) {
	w, p := testVillagePlayer()
	idx := p.Bag.firstIndex("potion_heal")
	if p.Bag.Slots[idx].InstanceID == "" {
		t.Fatal("instance id")
	}
	raw, _ := json.Marshal(map[string]int{"slot": idx, "qty": 1})
	if rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeSplitStack, Data: raw}), TypeSplitStack) {
		t.Fatal("split")
	}
	if p.Bag.count("potion_heal") != 3 || p.Bag.emptySlots() != InvCapacity-4 {
		t.Fatalf("split stacks count=%d empty=%d", p.Bag.count("potion_heal"), p.Bag.emptySlots())
	}
	lock, _ := json.Marshal(map[string]any{"slot": idx, "on": true})
	w.ApplyPhase14(p.ID, Envelope{Type: TypeLockItem, Data: lock})
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeDiscardItem, Data: []byte(`{"slot":` + itoa(idx) + `}`)}), TypeDiscardItem) {
		t.Fatal("locked discard")
	}
	w.giveItem(p, "quest_dawn_note", 1)
	q := p.Bag.firstIndex("quest_dawn_note")
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeDiscardItem, Data: []byte(`{"slot":` + itoa(q) + `}`)}), TypeDiscardItem) {
		t.Fatal("quest discard")
	}
}

func TestPhase27UpgradeAndCap(t *testing.T) {
	w, p := testVillagePlayer()
	w.giveItem(p, "light_staff", 1)
	slot := p.Bag.firstIndex("light_staff")
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeUpgradeItem, Data: []byte(`{"slot":` + itoa(slot) + `}`)}), TypeUpgradeItem) {
		t.Fatal("upgrade without material")
	}
	w.giveItem(p, "enhancement_stone", 20)
	w.giveCurrency(p, 400, 0)
	if rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeUpgradeItem, Data: []byte(`{"slot":` + itoa(slot) + `}`)}), TypeUpgradeItem) {
		t.Fatal("upgrade +1")
	}
	if p.Bag.Slots[slot].Upgrade != 1 {
		t.Fatalf("upgrade %d", p.Bag.Slots[slot].Upgrade)
	}
	p.ensureLog().Flags["upgrade_force_fail"] = true
	before := p.Bag.Slots[slot].Upgrade
	stones := p.Bag.count("enhancement_stone")
	w.ApplyInventory(p.ID, Envelope{Type: TypeUpgradeItem, Data: []byte(`{"slot":` + itoa(slot) + `}`)})
	if p.Bag.Slots[slot].Upgrade != before {
		t.Fatal("fail must not destroy or raise")
	}
	if p.Bag.count("enhancement_stone") >= stones {
		t.Fatal("material consumed")
	}
	p.Bag.Slots[slot].Upgrade = 10
	p.ensureLog().Flags["upgrade_force_fail"] = false
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeUpgradeItem, Data: []byte(`{"slot":` + itoa(slot) + `}`)}), TypeUpgradeItem) {
		t.Fatal("cap +10")
	}
}

func TestPhase27DawnSetBonus(t *testing.T) {
	w, p := testVillagePlayer()
	w.equipItem(p, p.Bag.firstIndex("training_gloves"))
	w.unequipItem(p, "ACCESSORY_1")
	w.giveItem(p, "dawn_helm", 1)
	w.giveItem(p, "dawn_mail", 1)
	w.giveItem(p, "dawn_staff", 1)
	w.giveItem(p, "dawn_ring", 1)
	baseHP := p.finalStats().MaxHP
	w.equipItem(p, p.Bag.firstIndex("dawn_helm"))
	w.equipItem(p, p.Bag.firstIndex("dawn_mail"))
	if p.setCount() != 2 || p.finalStats().MaxHP < baseHP+8 {
		t.Fatalf("2pc hp %d set %d", p.finalStats().MaxHP, p.setCount())
	}
	atk := p.finalStats().Attack
	w.equipItem(p, p.Bag.firstIndex("dawn_staff"))
	if p.setCount() != 3 || p.finalStats().Attack < atk+3 {
		t.Fatalf("3pc atk %d set %d", p.finalStats().Attack, p.setCount())
	}
	w.equipItem(p, p.Bag.firstIndex("dawn_ring"))
	if p.setCount() != 4 {
		t.Fatalf("4pc %d", p.setCount())
	}
	if p.energyRegenBonus() < 0.6 {
		t.Fatalf("4pc regen %f", p.energyRegenBonus())
	}
}

func TestPhase27TempLootAndBagExpand(t *testing.T) {
	w, p := testVillagePlayer()
	for p.Bag.emptySlots() > 0 {
		p.Bag.add("potion_heal", 1)
	}
	w.drops["drop-full"] = &WorldDrop{
		ID: "drop-full", ItemID: "dawn_blade", Qty: 1, OwnerID: p.ID, InstanceID: p.InstanceID,
		X: p.X, Z: p.Z, Until: time.Now().Add(time.Minute),
	}
	evs := w.ApplyInventory(p.ID, Envelope{Type: TypePickupItem, Data: []byte(`{"dropId":"drop-full"}`)})
	if rejectAction(evs, TypePickupItem) {
		t.Fatal("full bag should stash")
	}
	if len(p.ensureLog().TempLoot) < 1 || p.Bag.count("dawn_blade") != 0 {
		t.Fatal("temp loot")
	}
}

func TestPhase27BagExpandAndLoadout(t *testing.T) {
	w, p := testVillagePlayer()
	w.giveItem(p, "knowledge_token", 5)
	before := bagCapOf(p)
	if rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeExpandBag, Data: []byte(`{}`)}), TypeExpandBag) {
		t.Fatal("expand")
	}
	if bagCapOf(p) < before+Phase27BagStep || bagCapOf(p) > Phase27BagMax {
		t.Fatalf("cap %d", bagCapOf(p))
	}
	p.ensureLog().BagBonus = Phase27BagMax
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeExpandBag, Data: []byte(`{}`)}), TypeExpandBag) {
		t.Fatal("max 200")
	}
	w.equipItem(p, p.Bag.firstIndex("training_gloves"))
	if rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeSaveGearLoadout, Data: []byte(`{"slot":0}`)}), TypeSaveGearLoadout) {
		t.Fatal("save loadout")
	}
	p.InCombatUntil = time.Now().Add(time.Minute)
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeLoadGearLoadout, Data: []byte(`{"slot":0}`)}), TypeLoadGearLoadout) {
		t.Fatal("combat loadout")
	}
	p.InCombatUntil = time.Time{}
	if rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeLoadGearLoadout, Data: []byte(`{"slot":0}`)}), TypeLoadGearLoadout) {
		t.Fatal("safe loadout")
	}
}

func TestPhase27CosmeticNoPower(t *testing.T) {
	w, p := testVillagePlayer()
	w.giveItem(p, "cloak-dawn-soft", 1)
	before := p.powerRating()
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeEquipItem, Data: []byte(`{"slot":` + itoa(p.Bag.firstIndex("cloak-dawn-soft")) + `}`)}), TypeEquipItem) {
		t.Fatal("cosmetic combat equip")
	}
	if p.powerRating() != before {
		t.Fatal("cosmetic power")
	}
}

func TestPhase27TraceAdminOnly(t *testing.T) {
	w, p := testVillagePlayer()
	if !rejectAction(w.ApplyInventory(p.ID, Envelope{Type: TypeTraceItem, Data: []byte(`{}`)}), TypeTraceItem) {
		t.Fatal("trace")
	}
}
