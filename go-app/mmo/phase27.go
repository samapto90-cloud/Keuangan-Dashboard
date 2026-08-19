package mmo

import (
	"strings"
	"time"
)

// Phase 27 overlay: inventory + equipment + item progression.
// Reuses InventoryService, EquipmentService, ItemService, LootService,
// CharacterService, CombatService, TransformationService, DungeonService,
// RaidService, QuestService, TradeService, CraftingService.
// Do not duplicate those services.
//
// Logical tables on Inventory / EquipmentSet / PlayerLog / WorldState:
// items, item_instances, item_stats, item_lore, inventories, inventory_slots,
// equipment, equipment_loadouts, item_sets, item_set_bonuses, item_upgrades,
// item_enchantments, loot_tables, loot_entries, temporary_loot, item_history,
// guild_storage, guild_storage_logs.
// Indexes: playerId, itemId, itemInstanceId, inventoryId, equipmentId, guildId.

const (
	Phase27BagMax      = 200
	Phase27BagStep     = 10
	Phase27UpgradeMax  = 10
	Phase27EnchantMax  = 3
	Phase27Loadouts    = 3
	Phase27TempLootSec = 180
	DawnSetID          = "dawn-set"
)

var dawnSetPieces = map[string]bool{
	"dawn_helm": true, "dawn_mail": true, "dawn_staff": true, "dawn_ring": true,
}

func init() {
	registerPhase27()
}

func registerPhase27() {
	registerPhase27Items()
	registerLore(LoreDef{
		ID: "lore-dawn-blade", Title: "Dawn Blade", Region: "village",
		Text: "Pedang fajar tempaan desa. Cahaya tipis di mata bilahnya mengingatkan Raka pada janji pagi.",
	})
	registerLore(LoreDef{
		ID: "lore-light-staff", Title: "Light Staff", Region: "village",
		Text: "Tongkat cahaya yang dipakai penuntun desa. Bukan senjata kerajaan, melainkan alat perjalanan.",
	})
}

func registerPhase27Items() {
	eq := func(id, name, slot, rarity, icon, desc string, atk, def, hp, en int, lv int) {
		registerItem(ItemDef{
			ID: id, Name: name, Description: desc, Type: "EQUIPMENT", Slot: slot, Rarity: rarity,
			Icon: icon, Value: 24, LevelRequirement: lv, MaxStack: 1,
			Effects: ItemEffects{Attack: atk, Defense: def, MaxHP: hp, MaxEnergy: en},
		})
	}
	eq("light_staff", "LIGHT STAFF", "WEAPON", "UNCOMMON", "staff", "Tongkat cahaya original desa. Ringan dan jernih.", 7, 0, 0, 4, 1)
	eq("dawn_blade", "DAWN BLADE", "WEAPON", "RARE", "sword", "Bilah fajar original. Terhubung dengan lore desa.", 8, 0, 0, 0, 1)
	eq("iron_gauntlet", "IRON GAUNTLET", "WEAPON", "UNCOMMON", "gloves", "Sarung tinju besi penempa desa.", 6, 1, 0, 0, 1)
	eq("wind_spear", "WIND SPEAR", "WEAPON", "RARE", "spear", "Tombak angin original. Jangkauan pendek-menengah.", 7, 0, 0, 0, 1)
	eq("dawn-set-bracelet", "Dawn Bracelet Set", "ACCESSORY_3", "UNCOMMON", "pendant", "Gelang set fajar. Aksesori ketiga.", 0, 1, 4, 0, 1)
	registerItem(ItemDef{
		ID: "enhancement_stone", Name: "Enhancement Stone", Description: "Batu penguat equipment. Bahan upgrade +1 sampai +10.",
		Type: "MATERIAL", Rarity: "UNCOMMON", Stackable: true, MaxStack: 99, Icon: "crystal", Value: 8,
	})
	registerItem(ItemDef{
		ID: "knowledge_token", Name: "Knowledge Token", Description: "Token pendidikan. Dipakai ekspansi tas dan progres soal.",
		Type: "MATERIAL", Rarity: "UNCOMMON", Stackable: true, MaxStack: 99, Icon: "token", Value: 0, Untradable: true,
	})
	registerItem(ItemDef{
		ID: "cloak-dawn-soft", Name: "Dawn Soft Cloak", Description: "Jubah kosmetik. Tidak menambah Combat Power.",
		Type: "COSMETIC", Slot: "BODY", Rarity: "COMMON", Icon: "cloak", Value: 12, MaxStack: 1,
	})
	registerItem(ItemDef{
		ID: "quest_dawn_note", Name: "Catatan Desa", Description: "Barang misi. Tidak dapat dibuang.",
		Type: "QUEST_ITEM", Rarity: "COMMON", Icon: "token", Value: 0, Untradable: true, MaxStack: 1,
	})
}

func bagCapOf(p *Player) int {
	cap := InvCapacity
	if socialCfg.InventorySlots > cap {
		cap = socialCfg.InventorySlots
	}
	if p != nil && p.Bag != nil && p.Bag.Capacity > cap {
		cap = p.Bag.Capacity
	}
	if p != nil {
		if n := p.ensureLog().BagBonus; n > 0 && InvCapacity+n > cap {
			cap = InvCapacity + n
		}
	}
	if cap > Phase27BagMax {
		cap = Phase27BagMax
	}
	if cap < InvCapacity {
		cap = InvCapacity
	}
	return cap
}

func itemBind(def ItemDef) string {
	if def.Type == "QUEST" || def.Type == "QUEST_ITEM" {
		return "QUESTBOUND"
	}
	if def.Untradable {
		return "SOULBOUND"
	}
	return "TRADABLE"
}

func itemEquippable(def ItemDef) bool {
	switch def.Type {
	case "EQUIPMENT", "WEAPON", "ARMOR", "HELM", "ACCESSORY", "COSMETIC":
		return def.Slot != ""
	default:
		return false
	}
}

func rarityRank(r string) int {
	switch strings.ToUpper(r) {
	case "MYTHIC":
		return 6
	case "LEGENDARY":
		return 5
	case "EPIC":
		return 4
	case "RARE":
		return 3
	case "UNCOMMON":
		return 2
	default:
		return 1
	}
}

func capItemStats(s StatBlock) StatBlock {
	if s.Attack > 420 {
		s.Attack = 420
	}
	if s.Defense > 420 {
		s.Defense = 420
	}
	if s.MaxHP > 6000 {
		s.MaxHP = 6000
	}
	if s.MaxEnergy > 800 {
		s.MaxEnergy = 800
	}
	if s.CritChance > 0.45 {
		s.CritChance = 0.45
	}
	if s.MoveSpeed > 0.55 {
		s.MoveSpeed = 0.55
	}
	return s
}

func (p *Player) setBonusStats() StatBlock {
	n := 0
	for _, id := range p.Gear.ids() {
		if dawnSetPieces[id] {
			n++
		}
	}
	var s StatBlock
	if n >= 2 {
		s.MaxHP += 8
	}
	if n >= 3 {
		s.Attack += 3
	}
	if n >= 4 {
		s.MaxEnergy += 4
	}
	return s
}

func (p *Player) setCount() int {
	n := 0
	for _, id := range p.Gear.ids() {
		if dawnSetPieces[id] {
			n++
		}
	}
	return n
}

func (p *Player) inSafeZone() bool {
	if houseID(p.InstanceID) {
		return true
	}
	return p.inOpenWorldSafeZone()
}

func (w *WorldState) recordItemHist(p *Player, instanceID, itemID, action string) {
	row := ItemHistRow{InstanceID: instanceID, PlayerID: p.ID, ItemID: itemID, Action: action, At: time.Now().UnixMilli()}
	w.ItemHist = append(w.ItemHist, row)
	if len(w.ItemHist) > 200 {
		w.ItemHist = w.ItemHist[len(w.ItemHist)-120:]
	}
	log := p.ensureLog()
	log.ItemHist = append(log.ItemHist, row)
	if len(log.ItemHist) > 40 {
		log.ItemHist = log.ItemHist[len(log.ItemHist)-30:]
	}
}

func (w *WorldState) stashTempLoot(p *Player, itemID string, qty int) {
	if p == nil || itemID == "" || qty < 1 {
		return
	}
	log := p.ensureLog()
	name := itemID
	if def, ok := itemByID[itemID]; ok {
		name = def.Name
	}
	log.TempLoot = append(log.TempLoot, TempLootRow{
		ItemID: itemID, Name: name, Qty: qty, Until: time.Now().Add(Phase27TempLootSec * time.Second).UnixMilli(),
	})
	p.markDirty()
}

func (w *WorldState) expireTempLoot(p *Player) {
	log := p.ensureLog()
	if len(log.TempLoot) == 0 {
		return
	}
	now := time.Now().UnixMilli()
	keep := log.TempLoot[:0]
	for _, row := range log.TempLoot {
		if row.Until > now {
			keep = append(keep, row)
		}
	}
	log.TempLoot = keep
}

func (w *WorldState) ApplyPhase27(p *Player, env Envelope) [][]byte {
	if p == nil {
		return nil
	}
	p.ensureGear()
	switch env.Type {
	case TypeSplitStack:
		return w.splitStack(p, env.Data)
	case TypeUpgradeItem:
		return w.upgradeItem(p, env.Data)
	case TypeEnchantItem:
		return w.enchantItem(p, env.Data)
	case TypeExpandBag:
		return w.expandBag(p)
	case TypeSaveGearLoadout:
		return w.saveGearLoadout(p, env.Data)
	case TypeLoadGearLoadout:
		return w.loadGearLoadout(p, env.Data)
	case TypeClaimTempLoot:
		return w.claimTempLoot(p)
	case TypeSalvageItems:
		return w.salvageItems(p, env.Data)
	case TypeToggleCosmetic:
		p.ShowCosmetic = !p.ShowCosmetic
		w.persistGear(p)
		return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Cosmetic toggle.", nil))}
	case TypeTraceItem:
		return rejectFor(p.ID, TypeTraceItem, "admin_only")
	case TypeSetItemStats, TypeSetItemLevel, TypeDuplicateInstance, TypeSetInstanceID:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) splitStack(p *Player, data []byte) [][]byte {
	var in struct{ Slot, Qty int }
	_ = unmarshal(data, &in)
	if in.Slot < 0 || in.Slot >= len(p.Bag.Slots) || in.Qty < 1 {
		return rejectFor(p.ID, TypeSplitStack, "payload")
	}
	s := &p.Bag.Slots[in.Slot]
	def := itemByID[s.ItemID]
	if !def.Stackable || s.Qty <= in.Qty {
		return rejectFor(p.ID, TypeSplitStack, "stack")
	}
	if p.Bag.emptySlots() < 1 {
		return rejectFor(p.ID, TypeSplitStack, "capacity")
	}
	dest := -1
	for i := range p.Bag.Slots {
		if p.Bag.Slots[i].ItemID == "" {
			dest = i
			break
		}
	}
	if dest < 0 {
		return rejectFor(p.ID, TypeSplitStack, "capacity")
	}
	s.Qty -= in.Qty
	p.Bag.Slots[dest] = InvSlot{ItemID: s.ItemID, Qty: in.Qty, InstanceID: randomID("itm_"), ItemLevel: s.ItemLevel}
	p.Bag.bump()
	w.persistGear(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Stack dipisah.", []int{in.Slot, dest}))}
}

func (w *WorldState) upgradeItem(p *Player, data []byte) [][]byte {
	var in struct{ Slot int }
	_ = unmarshal(data, &in)
	if in.Slot < 0 || in.Slot >= len(p.Bag.Slots) {
		return rejectFor(p.ID, TypeUpgradeItem, "slot")
	}
	s := &p.Bag.Slots[in.Slot]
	def := itemByID[s.ItemID]
	if s.Qty < 1 || !itemEquippable(def) || def.Type == "COSMETIC" {
		return rejectFor(p.ID, TypeUpgradeItem, "item")
	}
	if s.Upgrade >= Phase27UpgradeMax {
		return rejectFor(p.ID, TypeUpgradeItem, "cap")
	}
	need := s.Upgrade + 1
	if p.Bag.count("enhancement_stone") < need {
		return rejectFor(p.ID, TypeUpgradeItem, "material")
	}
	cost := 20 * need
	if currencyOf(p.ensureLog(), "coin") < cost {
		return rejectFor(p.ID, TypeUpgradeItem, "cost")
	}
	if !p.Bag.takeItem("enhancement_stone", need) {
		return rejectFor(p.ID, TypeUpgradeItem, "material")
	}
	if !w.removeCurrency(p, "coin", cost, "upgrade") {
		p.Bag.add("enhancement_stone", need)
		return rejectFor(p.ID, TypeUpgradeItem, "cost")
	}
	ok := true
	if p.ensureLog().Flags["upgrade_force_fail"] {
		ok = false
	} else if need > 1 {
		ok = time.Now().UnixNano()%10 < 8
	}
	toast := "Upgrade gagal. Material terpakai."
	if ok {
		s.Upgrade++
		toast = "Upgrade +" + itoa(s.Upgrade) + " berhasil."
		w.recordItemHist(p, s.InstanceID, s.ItemID, "upgraded")
	}
	w.persistGear(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout(toast, []int{in.Slot})), marshal(TypePlayerStatsUpdated, p.statsView())}
}

func (w *WorldState) enchantItem(p *Player, data []byte) [][]byte {
	var in struct {
		Slot int
		Kind string
	}
	_ = unmarshal(data, &in)
	if in.Slot < 0 || in.Slot >= len(p.Bag.Slots) {
		return rejectFor(p.ID, TypeEnchantItem, "item")
	}
	s := &p.Bag.Slots[in.Slot]
	if s.Qty < 1 || !itemEquippable(itemByID[s.ItemID]) {
		return rejectFor(p.ID, TypeEnchantItem, "item")
	}
	if len(s.Enchants) >= Phase27EnchantMax {
		return rejectFor(p.ID, TypeEnchantItem, "limit")
	}
	kind := strings.ToUpper(in.Kind)
	switch kind {
	case "HP", "ATTACK", "ENERGY":
	default:
		kind = "HP"
	}
	if p.Bag.count("crystal_shard") < 1 {
		return rejectFor(p.ID, TypeEnchantItem, "material")
	}
	p.Bag.takeItem("crystal_shard", 1)
	s.Enchants = append(s.Enchants, kind)
	w.recordItemHist(p, s.InstanceID, s.ItemID, "enchanted")
	w.persistGear(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Enchant "+kind+".", []int{in.Slot}))}
}

func (w *WorldState) expandBag(p *Player) [][]byte {
	cap := bagCapOf(p)
	if cap >= Phase27BagMax {
		return rejectFor(p.ID, TypeExpandBag, "cap")
	}
	if p.Bag.count("knowledge_token") < 5 {
		return rejectFor(p.ID, TypeExpandBag, "material")
	}
	p.Bag.takeItem("knowledge_token", 5)
	p.ensureLog().BagBonus += Phase27BagStep
	p.Bag.Capacity = bagCapOf(p)
	p.ensureGear()
	w.persistGear(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Tas diperluas.", nil))}
}

func (w *WorldState) saveGearLoadout(p *Player, data []byte) [][]byte {
	var in struct{ Slot int }
	_ = unmarshal(data, &in)
	if in.Slot < 0 || in.Slot >= Phase27Loadouts {
		return rejectFor(p.ID, TypeSaveGearLoadout, "slot")
	}
	log := p.ensureLog()
	if len(log.GearLoadouts) < Phase27Loadouts {
		next := make([]GearLoadout, Phase27Loadouts)
		copy(next, log.GearLoadouts)
		log.GearLoadouts = next
	}
	up := map[string]int{}
	for k, v := range p.GearUp {
		up[k] = v
	}
	log.GearLoadouts[in.Slot] = GearLoadout{Slot: in.Slot, Gear: p.Gear, Up: up}
	p.markDirty()
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Loadout "+itoa(in.Slot+1)+" disimpan.", nil))}
}

func (w *WorldState) loadGearLoadout(p *Player, data []byte) [][]byte {
	if p.inCombatNow() {
		return rejectFor(p.ID, TypeLoadGearLoadout, "combat")
	}
	if !p.inSafeZone() {
		return rejectFor(p.ID, TypeLoadGearLoadout, "safe")
	}
	var in struct{ Slot int }
	_ = unmarshal(data, &in)
	log := p.ensureLog()
	if in.Slot < 0 || in.Slot >= len(log.GearLoadouts) {
		return rejectFor(p.ID, TypeLoadGearLoadout, "slot")
	}
	lo := log.GearLoadouts[in.Slot]
	empty := true
	for _, id := range lo.Gear.ids() {
		if id != "" {
			empty = false
			break
		}
	}
	if empty {
		return rejectFor(p.ID, TypeLoadGearLoadout, "empty")
	}
	p.Gear = lo.Gear
	p.GearUp = map[string]int{}
	for k, v := range lo.Up {
		p.GearUp[k] = v
	}
	p.applyDerived()
	w.persistGear(p)
	return [][]byte{marshal(TypeEquipmentUpdated, p.loadout("Loadout dipasang.", nil)), marshal(TypePlayerStatsUpdated, p.statsView())}
}

func (w *WorldState) claimTempLoot(p *Player) [][]byte {
	w.expireTempLoot(p)
	log := p.ensureLog()
	left := []TempLootRow{}
	got := 0
	for _, row := range log.TempLoot {
		if p.Bag.canFit(row.ItemID, row.Qty) && w.giveItem(p, row.ItemID, row.Qty) {
			got++
			continue
		}
		left = append(left, row)
	}
	log.TempLoot = left
	if got < 1 {
		return rejectFor(p.ID, TypeClaimTempLoot, "space")
	}
	w.persistGear(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Loot sementara diambil.", nil))}
}

func (w *WorldState) salvageItems(p *Player, data []byte) [][]byte {
	var in struct{ Slots []int }
	_ = unmarshal(data, &in)
	stones := 0
	for _, slot := range in.Slots {
		if slot < 0 || slot >= len(p.Bag.Slots) {
			continue
		}
		s := p.Bag.Slots[slot]
		def := itemByID[s.ItemID]
		if s.ItemID == "" || s.Locked || s.Favorite || rarityRank(def.Rarity) >= 3 || def.Type == "QUEST_ITEM" {
			continue
		}
		if p.Bag.removeAt(slot, s.Qty) {
			stones++
			w.recordItemHist(p, s.InstanceID, s.ItemID, "salvaged")
		}
	}
	if stones < 1 {
		return rejectFor(p.ID, TypeSalvageItems, "none")
	}
	p.Bag.add("enhancement_stone", stones)
	w.persistGear(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Salvage +"+itoa(stones)+" stone.", nil))}
}

type TempLootRow struct {
	ItemID string `json:"itemId"`
	Name   string `json:"name"`
	Qty    int    `json:"qty"`
	Until  int64  `json:"until"`
}

type ItemHistRow struct {
	InstanceID string `json:"itemInstanceId"`
	PlayerID   string `json:"playerId"`
	ItemID     string `json:"itemId"`
	Action     string `json:"action"`
	At         int64  `json:"at"`
}

type GearLoadout struct {
	Slot int            `json:"slot"`
	Gear EquipmentSet   `json:"gear"`
	Up   map[string]int `json:"up,omitempty"`
}

func copyIntMapVal(src map[string]int) map[string]int {
	if src == nil {
		return nil
	}
	out := make(map[string]int, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func copyStrMapVal(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
