package mmo

func (w *WorldState) ApplyShop(id string, env Envelope) [][]byte {
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	switch env.Type {
	case TypeShopOpen:
		return w.shopCatalogFor(p)
	case TypeShopBuy:
		var in struct {
			ShopItemID, ItemID, TransactionID string
		}
		_ = unmarshal(env.Data, &in)
		return w.shopBuy(p, in.ShopItemID, in.ItemID, in.TransactionID)
	case TypeShopSell:
		var in struct {
			Slot, Qty int
			ItemID    string
		}
		_ = unmarshal(env.Data, &in)
		return w.shopSell(p, in.Slot, in.ItemID, in.Qty)
	case TypeSetCoin:
		return rejectFor(p.ID, TypeSetCoin, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) shopCatalogFor(p *Player) [][]byte {
	log := p.ensureLog()
	w.resetBuys(log)
	items := make([]ShopItemView, 0, len(shopCatalog.Items))
	for _, it := range shopCatalog.Items {
		def := itemByID[it.ItemID]
		bought := log.BuyCounts[it.ShopItemID]
		items = append(items, ShopItemView{
			ShopItemID: it.ShopItemID, ItemID: it.ItemID, Name: def.Name, Price: it.Price,
			Currency: it.Currency, Stock: it.Stock, PurchaseLimit: it.PurchaseLimit, Bought: bought,
		})
	}
	return [][]byte{marshal(TypeShopCatalog, map[string]any{"id": shopCatalog.ID, "name": shopCatalog.Name, "items": items, "wallet": p.wallet()})}
}

func (w *WorldState) shopBuy(p *Player, shopItemID, itemID, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	var def ShopItemDef
	found := false
	for _, it := range shopCatalog.Items {
		if it.ShopItemID == shopItemID || it.ItemID == itemID && shopItemID == "" {
			def, found = it, true
			break
		}
	}
	if !found {
		return rejectFor(p.ID, TypeShopBuy, "item")
	}
	log := p.ensureLog()
	w.resetBuys(log)
	if def.PurchaseLimit > 0 && log.BuyCounts[def.ShopItemID] >= def.PurchaseLimit {
		return rejectFor(p.ID, TypeShopBuy, "limit")
	}
	if !w.removeCurrency(p, def.Currency, def.Price, "shop_buy") {
		return rejectFor(p.ID, TypeShopBuy, "currency")
	}
	if !w.giveItem(p, def.ItemID, 1) {
		w.addCurrency(p, def.Currency, def.Price, "shop_rollback")
		return rejectFor(p.ID, TypeShopBuy, "space")
	}
	if log.BuyCounts == nil {
		log.BuyCounts = map[string]int{}
	}
	log.BuyCounts[def.ShopItemID]++
	p.markDirty()
	w.persist(p)
	w.persistGear(p)
	w.audit("SHOP_PURCHASE", p.ID, def.ItemID)
	cat := w.shopCatalogFor(p)
	out := append([][]byte{marshal(TypeInventoryUpdated, p.loadout("Dibeli.", nil))}, cat...)
	return w.rememberTx(tx, out)
}

func (w *WorldState) shopSell(p *Player, slot int, itemID string, qty int) [][]byte {
	p.ensureGear()
	if qty < 1 {
		qty = 1
	}
	if slot < 0 || slot >= len(p.Bag.Slots) {
		return rejectFor(p.ID, TypeShopSell, "slot")
	}
	s := p.Bag.Slots[slot]
	if s.ItemID == "" || (itemID != "" && s.ItemID != itemID) {
		return rejectFor(p.ID, TypeShopSell, "item")
	}
	def := itemByID[s.ItemID]
	if def.Untradable || def.Value <= 0 || s.Locked || s.Favorite {
		return rejectFor(p.ID, TypeShopSell, "tradable")
	}
	if !p.Bag.removeAt(slot, qty) {
		return rejectFor(p.ID, TypeShopSell, "qty")
	}
	w.addCurrency(p, "coin", def.Value*qty, "shop_sell")
	w.persist(p)
	w.persistGear(p)
	w.audit("ITEM_CHANGE", p.ID, "sell:"+s.ItemID)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Terjual.", []int{slot}))}
}

func (w *WorldState) resetBuys(log *PlayerLog) {
	day := utcDay()
	if log.BuyDay != day {
		log.BuyDay = day
		log.BuyCounts = map[string]int{}
	}
	if log.BuyCounts == nil {
		log.BuyCounts = map[string]int{}
	}
}
