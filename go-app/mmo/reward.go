package mmo

func (inv *Inventory) canFit(itemID string, qty int) bool {
	def, ok := itemByID[itemID]
	if !ok || qty <= 0 {
		return false
	}
	max := def.MaxStack
	if max < 1 {
		max = 1
	}
	left := qty
	if def.Stackable {
		for _, s := range inv.Slots {
			if s.ItemID == itemID && s.Qty < max {
				space := max - s.Qty
				if space > left {
					space = left
				}
				left -= space
			}
		}
	}
	for _, s := range inv.Slots {
		if left <= 0 {
			return true
		}
		if s.ItemID != "" {
			continue
		}
		if def.Stackable {
			take := left
			if take > max {
				take = max
			}
			left -= take
		} else {
			left--
		}
	}
	return left <= 0
}

func (w *WorldState) giveExp(p *Player, amount int) [][]byte {
	return w.grantExp(p, amount)
}

func (w *WorldState) giveItem(p *Player, itemID string, qty int) bool {
	p.ensureGear()
	if !p.Bag.canFit(itemID, qty) {
		return false
	}
	_, ok := p.Bag.add(itemID, qty)
	if ok {
		if itemID == "potion_heal" {
			p.ensureLog().Potion += qty
		}
		if itemID == "potion_energy" {
			p.ensureLog().EnergyPotion += qty
		}
		if itemID == "crystal_shard" {
			p.ensureLog().Crystal += qty
		}
		if itemID == "edu_token" {
			p.ensureLog().EduToken += qty
			if p.Bag.canFit("knowledge_token", qty) {
				p.Bag.add("knowledge_token", qty)
			}
		}
		w.recordItemHist(p, "", itemID, "created")
		p.markDirty()
	}
	return ok
}

func (w *WorldState) giveCurrency(p *Player, coin, crystal int) {
	if coin > 0 {
		n := int(float64(coin)*GoldGenMult + 0.5)
		if n < 1 {
			n = 1
		}
		w.addCurrency(p, "gold", n, "reward")
		w.GoldCreated += n
		w.recordEconomy(p.ID, "gold", n, "reward", "")
	}
	if crystal > 0 {
		w.addCurrency(p, "crystal", crystal, "reward")
	}
}

func (w *WorldState) giveEducationToken(p *Player, n int) {
	if n > 0 {
		w.giveItem(p, "edu_token", n)
	}
}

func (w *WorldState) giveRewardBundle(p *Player, r RewardDef) [][]byte {
	p.ensureGear()
	w.giveCurrency(p, r.Coin, r.Crystal)
	if r.Potion > 0 {
		if !w.giveItem(p, "potion_heal", r.Potion) {
			w.stashTempLoot(p, "potion_heal", r.Potion)
		}
	}
	if r.EduToken > 0 {
		if p.Bag.canFit("edu_token", r.EduToken) {
			w.giveEducationToken(p, r.EduToken)
		} else {
			w.stashTempLoot(p, "edu_token", r.EduToken)
		}
	}
	if r.Crystal > 0 {
		if !w.giveItem(p, "crystal_shard", r.Crystal) {
			w.stashTempLoot(p, "crystal_shard", r.Crystal)
		}
	}
	if r.Knowledge > 0 {
		p.grantKnowledge(r.Knowledge)
	}
	events := w.giveExp(p, r.Exp)
	w.persistGear(p)
	w.persist(p)
	return events
}

func (w *WorldState) partyShareKill(killer *Player, kind string) {
	if w.Parties == nil || killer.PartyID == "" {
		w.creditKill(killer, kind)
		return
	}
	pt := w.Parties.Get(killer.PartyID)
	if pt == nil {
		w.creditKill(killer, kind)
		return
	}
	w.creditKill(killer, kind)
	for _, id := range pt.Members {
		if id == killer.ID {
			continue
		}
		m := w.players[id]
		if m == nil || !m.Connected || !m.alive() {
			continue
		}
		if hypot2(killer.X, killer.Z, m.X, m.Z) > PartyShareRange*PartyShareRange {
			continue
		}
		w.creditKill(m, kind)
		w.giveExp(m, 8)
	}
}

func (w *WorldState) partyShareExp(killer *Player, exp int) [][]byte {
	events := w.giveExp(killer, exp)
	if w.Parties == nil || killer.PartyID == "" || exp <= 0 {
		return events
	}
	pt := w.Parties.Get(killer.PartyID)
	if pt == nil {
		return events
	}
	share := exp / 4
	if share < 1 {
		share = 1
	}
	for _, id := range pt.Members {
		if id == killer.ID {
			continue
		}
		m := w.players[id]
		if m == nil || !m.Connected || !m.alive() {
			continue
		}
		if hypot2(killer.X, killer.Z, m.X, m.Z) > PartyShareRange*PartyShareRange {
			continue
		}
		events = append(events, w.giveExp(m, share)...)
	}
	return events
}
