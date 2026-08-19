package mmo

type StatBlock struct {
	Attack, Defense, MaxHP, MaxEnergy, Strength, Agility, EnergyPower int
	CritChance, MoveSpeed, Range                                      float64
}

func (a StatBlock) add(b StatBlock) StatBlock {
	a.Attack += b.Attack
	a.Defense += b.Defense
	a.MaxHP += b.MaxHP
	a.MaxEnergy += b.MaxEnergy
	a.Strength += b.Strength
	a.Agility += b.Agility
	a.EnergyPower += b.EnergyPower
	a.CritChance += b.CritChance
	a.MoveSpeed += b.MoveSpeed
	a.Range += b.Range
	return a
}

func effectsToStats(e ItemEffects) StatBlock {
	return StatBlock{
		Attack: e.Attack, Defense: e.Defense, MaxHP: e.MaxHP, MaxEnergy: e.MaxEnergy,
		Strength: e.Strength, Agility: e.Agility, EnergyPower: e.EnergyPower,
		CritChance: e.CriticalChance, MoveSpeed: e.MovementSpeed, Range: e.Range,
	}
}

func (p *Player) ensureGear() {
	if p.Bag == nil {
		p.Bag = newInventory(p.ID)
	}
	if p.GearUp == nil {
		p.GearUp = map[string]int{}
	}
	if p.GearInst == nil {
		p.GearInst = map[string]string{}
	}
	cap := bagCapOf(p)
	if len(p.Bag.Slots) < cap {
		slots := make([]InvSlot, cap)
		copy(slots, p.Bag.Slots)
		p.Bag.Slots = slots
	}
	if p.Bag.Capacity < cap {
		p.Bag.Capacity = cap
	}
	bankCap := socialCfg.BankSlots
	if bankCap < 1 {
		bankCap = 100
	}
	if p.Bank == nil {
		p.Bank = &Inventory{ID: "bank-" + p.ID, Capacity: bankCap, Slots: make([]InvSlot, bankCap)}
	}
	if len(p.Bank.Slots) != bankCap {
		slots := make([]InvSlot, bankCap)
		copy(slots, p.Bank.Slots)
		p.Bank.Slots = slots
		p.Bank.Capacity = bankCap
	}
	if p.BaseMaxHP <= 0 {
		p.BaseMaxHP = BaseMaxHP
	}
	if p.BaseMaxEnergy <= 0 {
		p.BaseMaxEnergy = EnergyMax
	}
	if p.BaseStrength <= 0 {
		p.BaseStrength = BaseStrength
	}
	if p.BaseDefense <= 0 {
		p.BaseDefense = BaseDefense
	}
}

func (p *Player) baseStats() StatBlock {
	p.ensureGear()
	return StatBlock{
		Attack: p.BaseStrength, Defense: p.BaseDefense, MaxHP: p.BaseMaxHP, MaxEnergy: p.BaseMaxEnergy,
		Strength: p.BaseStrength, Agility: p.BaseAgility, EnergyPower: p.BaseEnergyPower,
		CritChance: CritChance, MoveSpeed: 0, Range: 0,
	}
}

func (p *Player) equipmentStats() StatBlock {
	p.ensureGear()
	var s StatBlock
	for _, slot := range []string{"HEAD", "BODY", "LEGS", "WEAPON", "ACCESSORY_1", "ACCESSORY_2", "ACCESSORY_3"} {
		id := p.Gear.get(slot)
		if id == "" {
			continue
		}
		def, ok := itemByID[id]
		if !ok || def.Type == "COSMETIC" {
			continue
		}
		s = s.add(effectsToStats(def.Effects))
		up := p.GearUp[slot]
		if up > Phase27UpgradeMax {
			up = Phase27UpgradeMax
		}
		if slot == "WEAPON" {
			s.Attack += up
		} else {
			s.Defense += up
		}
	}
	s = s.add(p.setBonusStats())
	return capItemStats(s)
}

func (p *Player) buffStats() StatBlock {
	s := p.transformBuff()
	s.Strength += p.SpentSTR
	s.Defense += p.SpentDEF
	s.Agility += p.SpentAGI
	s.EnergyPower += p.SpentENG
	s.MaxHP += p.SpentVIT * progressionCfg.HPPerVitality
	s.MaxEnergy += p.SpentENG * progressionCfg.EnergyPerEnergy
	s.Attack += int(float64(p.SpentSTR)*progressionCfg.AttackPerStrength + 0.5)
	s.Defense += int(float64(p.SpentDEF)*progressionCfg.DefensePerDefense + 0.5)
	s.MoveSpeed += float64(p.SpentAGI) * progressionCfg.MovePerAgility
	s = s.add(p.styleBuff())
	s = s.add(p.formPassive())
	if p.hasSkill("iron_guard") {
		s.Defense += 2
	}
	if p.hasSkill("energy_control") {
		s.MaxEnergy += 5
	}
	if p.hasSkill("quick_step") {
		s.MoveSpeed += 0.04
	}
	return s
}

func (p *Player) finalStats() StatBlock {
	return p.baseStats().add(p.equipmentStats()).add(p.buffStats())
}

func (p *Player) applyDerived() {
	f := capItemStats(p.finalStats())
	oldMax := p.MaxHP
	p.Strength = f.Strength
	p.Defense = f.Defense
	p.Agility = f.Agility
	p.EnergyPower = f.EnergyPower
	p.CritChance = f.CritChance
	p.MoveSpeedBonus = capMoveBonus(f.MoveSpeed)
	p.AttackRange = f.Range
	p.EquipAttack = f.Attack
	p.MaxHP = f.MaxHP
	p.MaxEnergy = f.MaxEnergy
	if p.MaxHP < 1 {
		p.MaxHP = 1
	}
	if oldMax > 0 && p.HP > 0 {
		p.HP = p.HP * p.MaxHP / oldMax
	}
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	if p.Energy > p.MaxEnergy {
		p.Energy = p.MaxEnergy
	}
}

func (p *Player) statsView() StatsView {
	f := p.finalStats()
	return StatsView{
		Level: p.Level, Class: p.Class, HP: p.HP, MaxHP: p.MaxHP, Energy: p.Energy, MaxEnergy: p.MaxEnergy,
		Stamina: int(p.Stamina + 0.5), Attack: f.Attack, Defense: f.Defense, Strength: f.Strength,
		Agility: f.Agility, EnergyPower: f.EnergyPower, Vitality: p.SpentVIT, CriticalChance: f.CritChance, MoveSpeed: f.MoveSpeed,
		AttributePoints: p.AttributePoints, SkillPoints: p.SkillPoints, PowerRating: p.powerRating(),
		FormID: p.FormID, TransformState: p.TransformState,
		Dodge: p.Agility / 40, EnergyRegen: p.energyRegenBonus(), ShowCosmetic: p.ShowCosmetic,
	}
}

func (p *Player) slotViews(indexes []int) []InvSlotView {
	p.ensureGear()
	if len(indexes) == 0 {
		indexes = make([]int, len(p.Bag.Slots))
		for i := range indexes {
			indexes[i] = i
		}
	}
	out := make([]InvSlotView, 0, len(indexes))
	for _, i := range indexes {
		if i < 0 || i >= len(p.Bag.Slots) {
			continue
		}
		s := p.Bag.Slots[i]
		v := InvSlotView{Index: i, Qty: s.Qty, Locked: s.Locked, Favorite: s.Favorite, InstanceID: s.InstanceID, Upgrade: s.Upgrade, ItemLevel: s.ItemLevel}
		if s.ItemID != "" {
			if def, ok := itemByID[s.ItemID]; ok {
				d := def.view()
				if s.ItemLevel > 0 {
					d.ItemLevel = s.ItemLevel
				}
				v.Item = &d
			}
		}
		out = append(out, v)
	}
	return out
}

func (p *Player) loadout(toast string, changed []int) InventoryUpdated {
	p.ensureGear()
	log := p.ensureLog()
	full := len(changed) == 0
	upd := InventoryUpdated{
		PlayerID: p.ID, Version: p.Bag.Version, ChangedSlots: p.slotViews(changed),
		Equipment: p.Gear.view(), Stats: p.statsView(),
		Coin: log.Coin, Crystal: log.Crystal, EduToken: log.EduToken,
		BattleToken: log.BattleToken, GuardianToken: log.GuardianTokens, RaidToken: log.RaidTokens, Toast: toast,
		TempLoot: append([]TempLootRow{}, log.TempLoot...), SetPieces: p.setCount(), BagCapacity: p.Bag.Capacity,
		ShowCosmetic: p.ShowCosmetic, ItemHist: append([]ItemHistRow{}, log.ItemHist...),
	}
	if full {
		upd.Slots = p.slotViews(nil)
		upd.ChangedSlots = upd.Slots
		if p.Bank != nil {
			upd.Bank = p.Bank.views()
		}
	}
	return upd
}
