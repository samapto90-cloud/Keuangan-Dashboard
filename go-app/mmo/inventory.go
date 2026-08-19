package mmo

import (
	"math"
	"time"
)

type GearSave struct {
	Inv                                                                               *Inventory
	Bank                                                                              *Inventory
	Gear                                                                              EquipmentSet
	BaseMaxHP, BaseMaxEnergy, BaseStrength, BaseDefense, BaseAgility, BaseEnergyPower int
	Level, Exp, ExpToNext, HP, Energy                                                 int
	Stamina                                                                           float64
	AttributePoints, SkillPoints                                                      int
	SpentSTR, SpentDEF, SpentAGI, SpentENG, SpentVIT                                  int
	UnlockedSkills, UnlockedForms                                                     []string
	FormID                                                                            string
	TransEnergy, MaxTransEnergy                                                       int
	CombatStyle                                                                       string
	LoadoutSkills                                                                     []string
	LoadoutUlt                                                                        string
	AttrResetUsed, UltCharge                                                          int
	GearUp                                                                            map[string]int
	GearInst                                                                          map[string]string
	ShowCosmetic                                                                      bool
}

type InventoryRepository interface {
	Load(playerID string) *GearSave
	Save(playerID string, s *GearSave)
}

type MemoryInvRepo struct {
	data map[string]*GearSave
}

func NewMemoryInvRepo() *MemoryInvRepo {
	return &MemoryInvRepo{data: map[string]*GearSave{}}
}

func (r *MemoryInvRepo) Load(id string) *GearSave {
	s := r.data[id]
	if s == nil {
		return nil
	}
	cp := *s
	cp.Inv = cloneInv(s.Inv)
	cp.Bank = cloneInv(s.Bank)
	cp.UnlockedSkills = append([]string{}, s.UnlockedSkills...)
	cp.UnlockedForms = append([]string{}, s.UnlockedForms...)
	cp.LoadoutSkills = append([]string{}, s.LoadoutSkills...)
	cp.GearUp = copyIntMapVal(s.GearUp)
	cp.GearInst = copyStrMapVal(s.GearInst)
	return &cp
}

func (r *MemoryInvRepo) Save(id string, s *GearSave) {
	if s == nil {
		return
	}
	cp := *s
	cp.Inv = cloneInv(s.Inv)
	cp.Bank = cloneInv(s.Bank)
	cp.UnlockedSkills = append([]string{}, s.UnlockedSkills...)
	cp.UnlockedForms = append([]string{}, s.UnlockedForms...)
	cp.LoadoutSkills = append([]string{}, s.LoadoutSkills...)
	cp.GearUp = copyIntMapVal(s.GearUp)
	cp.GearInst = copyStrMapVal(s.GearInst)
	r.data[id] = &cp
}

func (p *Player) gearSave() *GearSave {
	p.ensureGear()
	return &GearSave{
		Inv: cloneInv(p.Bag), Bank: cloneInv(p.Bank), Gear: p.Gear,
		BaseMaxHP: p.BaseMaxHP, BaseMaxEnergy: p.BaseMaxEnergy, BaseStrength: p.BaseStrength,
		BaseDefense: p.BaseDefense, BaseAgility: p.BaseAgility, BaseEnergyPower: p.BaseEnergyPower,
		Level: p.Level, Exp: p.Exp, ExpToNext: p.ExpToNext, HP: p.HP, Energy: p.Energy, Stamina: p.Stamina,
		AttributePoints: p.AttributePoints, SkillPoints: p.SkillPoints,
		SpentSTR: p.SpentSTR, SpentDEF: p.SpentDEF, SpentAGI: p.SpentAGI, SpentENG: p.SpentENG, SpentVIT: p.SpentVIT,
		UnlockedSkills: append([]string{}, p.UnlockedSkills...), UnlockedForms: append([]string{}, p.UnlockedForms...),
		FormID: p.FormID, TransEnergy: p.TransEnergy, MaxTransEnergy: p.MaxTransEnergy,
		CombatStyle: p.CombatStyle, LoadoutSkills: append([]string{}, p.LoadoutSkills...), LoadoutUlt: p.LoadoutUlt,
		AttrResetUsed: p.AttrResetUsed, UltCharge: p.UltCharge,
		GearUp: copyIntMapVal(p.GearUp), GearInst: copyStrMapVal(p.GearInst), ShowCosmetic: p.ShowCosmetic,
	}
}

func (p *Player) applyGearSave(s *GearSave) {
	if s == nil {
		p.ensureGear()
		p.applyDerived()
		return
	}
	p.Bag = cloneInv(s.Inv)
	p.Bank = cloneInv(s.Bank)
	p.Gear = s.Gear
	if s.BaseMaxHP > 0 {
		p.BaseMaxHP = s.BaseMaxHP
	}
	if s.BaseMaxEnergy > 0 {
		p.BaseMaxEnergy = s.BaseMaxEnergy
	}
	if s.BaseStrength > 0 {
		p.BaseStrength = s.BaseStrength
	}
	if s.BaseDefense > 0 {
		p.BaseDefense = s.BaseDefense
	}
	p.BaseAgility = s.BaseAgility
	p.BaseEnergyPower = s.BaseEnergyPower
	if s.Level > 0 {
		p.Level = s.Level
	}
	p.Exp, p.ExpToNext = s.Exp, s.ExpToNext
	p.AttributePoints, p.SkillPoints = s.AttributePoints, s.SkillPoints
	p.SpentSTR, p.SpentDEF, p.SpentAGI, p.SpentENG, p.SpentVIT = s.SpentSTR, s.SpentDEF, s.SpentAGI, s.SpentENG, s.SpentVIT
	p.UnlockedSkills = append([]string{}, s.UnlockedSkills...)
	p.UnlockedForms = append([]string{}, s.UnlockedForms...)
	if s.FormID != "" {
		p.FormID = s.FormID
		if p.FormID != "normal" {
			p.TransformState = "NORMAL"
			p.FormID = "normal"
		}
	}
	p.TransEnergy, p.MaxTransEnergy = s.TransEnergy, s.MaxTransEnergy
	if s.CombatStyle != "" {
		p.CombatStyle = s.CombatStyle
	}
	p.LoadoutSkills = append([]string{}, s.LoadoutSkills...)
	p.LoadoutUlt = s.LoadoutUlt
	p.AttrResetUsed = s.AttrResetUsed
	p.UltCharge = s.UltCharge
	p.GearUp = copyIntMapVal(s.GearUp)
	p.GearInst = copyStrMapVal(s.GearInst)
	p.ShowCosmetic = s.ShowCosmetic
	p.ensureGear()
	p.applyDerived()
	if s.HP > 0 {
		p.HP = s.HP
	}
	if s.Energy > 0 {
		p.Energy = s.Energy
	}
	if s.Stamina > 0 {
		p.Stamina = s.Stamina
	}
}

type WorldDrop struct {
	ID, ItemID, OwnerID, InstanceID string
	Qty                             int
	X, Z                            float64
	Until                           time.Time
}

func (w *WorldState) persistGear(p *Player) {
	if p == nil {
		return
	}
	if w.InvRepo != nil {
		w.InvRepo.Save(p.ID, p.gearSave())
	}
	w.persistJourney(p)
}

func (w *WorldState) ApplyInventory(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	p.ensureGear()
	switch env.Type {
	case TypeGiveItem, TypeGiveCurrency, TypeAddItem, TypeSetQuantity,
		TypeSetItemStats, TypeSetItemLevel, TypeDuplicateInstance, TypeSetInstanceID:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	case TypeGetInventory:
		return [][]byte{marshal(TypeInventoryUpdated, p.loadout("", nil))}
	case TypePickupItem:
		var in PickupIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypePickupItem, "payload")
		}
		return w.pickupDrop(p, in)
	case TypeUseItem:
		var in SlotIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypeUseItem, "payload")
		}
		return w.useItem(p, in.Slot)
	case TypeEquipItem:
		var in SlotIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypeEquipItem, "payload")
		}
		return w.equipItem(p, in.Slot)
	case TypeUnequipItem:
		var in SlotIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypeUnequipItem, "payload")
		}
		return w.unequipItem(p, in.EquipSlot)
	case TypeDiscardItem:
		var in SlotIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypeDiscardItem, "payload")
		}
		if in.Slot < 0 || in.Slot >= len(p.Bag.Slots) {
			return rejectFor(p.ID, TypeDiscardItem, "slot")
		}
		s := p.Bag.Slots[in.Slot]
		def := itemByID[s.ItemID]
		if s.ItemID == "" {
			return rejectFor(p.ID, TypeDiscardItem, "slot")
		}
		if s.Locked || s.Favorite {
			return rejectFor(p.ID, TypeDiscardItem, "protected")
		}
		if def.Type == "QUEST" || def.Type == "QUEST_ITEM" {
			return rejectFor(p.ID, TypeDiscardItem, "quest")
		}
		if !p.Bag.removeAt(in.Slot, 1) {
			return rejectFor(p.ID, TypeDiscardItem, "slot")
		}
		w.recordItemHist(p, s.InstanceID, s.ItemID, "discarded")
		w.persistGear(p)
		return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Dibuang.", []int{in.Slot}))}
	default:
		return w.ApplyPhase27(p, env)
	}
}

func (w *WorldState) pickupDrop(p *Player, in PickupIn) [][]byte {
	if !p.alive() {
		return rejectFor(p.ID, TypePickupItem, "dead")
	}
	if in.Quantity > 20 || in.ItemID != "" && in.DropID == "" {
		return rejectFor(p.ID, TypePickupItem, "quantity")
	}
	d := w.drops[in.DropID]
	if d == nil || time.Now().After(d.Until) {
		return rejectFor(p.ID, TypePickupItem, "drop")
	}
	if _, ok := itemByID[d.ItemID]; !ok {
		return rejectFor(p.ID, TypePickupItem, "item")
	}
	if math.Hypot(p.X-d.X, p.Z-d.Z) > 2.8 {
		return rejectFor(p.ID, TypePickupItem, "distance")
	}
	if d.InstanceID != p.InstanceID {
		return rejectFor(p.ID, TypePickupItem, "instance")
	}
	if d.OwnerID != "" && d.OwnerID != p.ID && !w.sameParty(d.OwnerID, p.ID) {
		return rejectFor(p.ID, TypePickupItem, "ownership")
	}
	before := p.Bag.count(d.ItemID)
	changed, ok := p.Bag.add(d.ItemID, d.Qty)
	def := itemByID[d.ItemID]
	toast := "Memperoleh " + def.Name + "."
	if !ok {
		left := d.Qty - (p.Bag.count(d.ItemID) - before)
		if left < 1 {
			left = 0
		}
		if left > 0 {
			w.stashTempLoot(p, d.ItemID, left)
		}
		delete(w.drops, d.ID)
		w.persistGear(p)
		return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Tas penuh. "+toast, changed))}
	}
	delete(w.drops, d.ID)
	w.recordItemHist(p, "", d.ItemID, "looted")
	w.persistGear(p)
	return [][]byte{
		marshal(TypeItemAdded, p.loadout(toast, changed)),
		marshal(TypeInventoryUpdated, p.loadout("", changed)),
	}
}

func (w *WorldState) useItem(p *Player, slot int) [][]byte {
	if slot < 0 || slot >= len(p.Bag.Slots) {
		return rejectFor(p.ID, TypeUseItem, "slot")
	}
	s := p.Bag.Slots[slot]
	def, ok := itemByID[s.ItemID]
	if !ok || s.Qty < 1 {
		return rejectFor(p.ID, TypeUseItem, "item")
	}
	if def.Type != "CONSUMABLE" {
		return rejectFor(p.ID, TypeUseItem, "type")
	}
	if !p.alive() {
		return rejectFor(p.ID, TypeUseItem, "dead")
	}
	if !p.Bag.removeAt(slot, 1) {
		return rejectFor(p.ID, TypeUseItem, "consume")
	}
	if def.Effects.HealPct > 0 {
		heal := int(float64(p.MaxHP)*def.Effects.HealPct + 0.5)
		p.HP += heal
		if p.HP > p.MaxHP {
			p.HP = p.MaxHP
		}
	}
	if def.Effects.EnergyPct > 0 {
		p.Energy += int(float64(p.MaxEnergy)*def.Effects.EnergyPct + 0.5)
		if p.Energy > p.MaxEnergy {
			p.Energy = p.MaxEnergy
		}
	}
	if def.Effects.StaminaPct > 0 {
		p.Stamina += p.MaxStamina * def.Effects.StaminaPct
		if p.Stamina > p.MaxStamina {
			p.Stamina = p.MaxStamina
		}
	}
	w.persistGear(p)
	return [][]byte{
		marshal(TypeItemConsumed, p.loadout(def.Name+" digunakan.", []int{slot})),
		marshal(TypeItemUsed, p.loadout("", []int{slot})),
		marshal(TypePlayerStatsUpdated, p.statsView()),
		marshal(TypeInventoryUpdated, p.loadout("", []int{slot})),
	}
}

func (w *WorldState) equipItem(p *Player, slot int) [][]byte {
	if slot < 0 || slot >= len(p.Bag.Slots) {
		return rejectFor(p.ID, TypeEquipItem, "slot")
	}
	s := p.Bag.Slots[slot]
	def, ok := itemByID[s.ItemID]
	if !ok || !itemEquippable(def) || def.Slot == "" || def.Type == "COSMETIC" {
		return rejectFor(p.ID, TypeEquipItem, "type")
	}
	if p.Level < def.LevelRequirement {
		return rejectFor(p.ID, TypeEquipItem, "level")
	}
	if s.Qty < 1 {
		return rejectFor(p.ID, TypeEquipItem, "ownership")
	}
	old := p.Gear.get(def.Slot)
	up, inst := s.Upgrade, s.InstanceID
	ench := append([]string{}, s.Enchants...)
	if !p.Bag.removeAt(slot, 1) {
		return rejectFor(p.ID, TypeEquipItem, "item")
	}
	if old != "" {
		changed, ok := p.Bag.add(old, 1)
		if !ok {
			p.Bag.add(s.ItemID, 1)
			return rejectFor(p.ID, TypeEquipItem, "capacity")
		}
		if len(changed) > 0 {
			back := &p.Bag.Slots[changed[len(changed)-1]]
			back.Upgrade = p.GearUp[def.Slot]
			back.InstanceID = p.GearInst[def.Slot]
		}
	}
	p.Gear.set(def.Slot, def.ID)
	if p.GearUp == nil {
		p.GearUp = map[string]int{}
	}
	if p.GearInst == nil {
		p.GearInst = map[string]string{}
	}
	p.GearUp[def.Slot] = up
	p.GearInst[def.Slot] = inst
	_ = ench
	p.applyDerived()
	w.recordItemHist(p, inst, def.ID, "equipped")
	w.persistGear(p)
	return [][]byte{
		marshal(TypeEquipmentUpdated, p.loadout("Dipasang: "+def.Name, nil)),
		marshal(TypePlayerStatsUpdated, p.statsView()),
		marshal(TypeInventoryUpdated, p.loadout("", nil)),
	}
}

func (w *WorldState) unequipItem(p *Player, slotName string) [][]byte {
	id := p.Gear.get(slotName)
	if id == "" {
		return rejectFor(p.ID, TypeUnequipItem, "empty")
	}
	changed, ok := p.Bag.add(id, 1)
	if !ok {
		return rejectFor(p.ID, TypeUnequipItem, "capacity")
	}
	if len(changed) > 0 {
		back := &p.Bag.Slots[changed[len(changed)-1]]
		back.Upgrade = p.GearUp[slotName]
		back.InstanceID = p.GearInst[slotName]
	}
	p.Gear.set(slotName, "")
	delete(p.GearUp, slotName)
	delete(p.GearInst, slotName)
	p.applyDerived()
	w.persistGear(p)
	def := itemByID[id]
	return [][]byte{
		marshal(TypeEquipmentUpdated, p.loadout("Dilepas: "+def.Name, nil)),
		marshal(TypePlayerStatsUpdated, p.statsView()),
		marshal(TypeInventoryUpdated, p.loadout("", nil)),
	}
}

func (w *WorldState) spawnDrop(owner *Player, e *Enemy) {
	itemID, qty := dropFor(e.Def.ID)
	if itemID == "" || qty < 1 {
		return
	}
	id := "drop-" + itoa(int(time.Now().UnixNano()%100000000))
	if w.drops == nil {
		w.drops = map[string]*WorldDrop{}
	}
	w.drops[id] = &WorldDrop{
		ID: id, ItemID: itemID, Qty: qty, OwnerID: owner.ID, InstanceID: owner.InstanceID,
		X: e.X, Z: e.Z, Until: time.Now().Add(45 * time.Second),
	}
}

func dropFor(kind string) (string, int) {
	switch kind {
	case "training_dummy":
		return "potion_heal", 1
	case "forest_fang":
		if time.Now().UnixNano()%10 < 5 {
			return "potion_heal", 1
		}
		return "crystal_shard", 1
	case "shadow_imp":
		return "potion_energy", 1
	case "stone_beast":
		return "crystal_shard", 2
	case "elite_shadow_beast":
		return "potion_heal", 1
	}
	return "", 0
}

func (w *WorldState) dropSnaps() []DropSnapshot {
	return w.dropSnapsFor("")
}

func (w *WorldState) dropSnapsFor(instanceID string) []DropSnapshot {
	now := time.Now()
	out := make([]DropSnapshot, 0, len(w.drops))
	for id, d := range w.drops {
		if d == nil || now.After(d.Until) {
			delete(w.drops, id)
			continue
		}
		if d.InstanceID != instanceID {
			continue
		}
		name, rarity := d.ItemID, "COMMON"
		if def, ok := itemByID[d.ItemID]; ok {
			name, rarity = def.Name, def.Rarity
		}
		out = append(out, DropSnapshot{ID: d.ID, ItemID: d.ItemID, Name: name, Qty: d.Qty, X: d.X, Z: d.Z, Rarity: rarity})
	}
	return out
}
