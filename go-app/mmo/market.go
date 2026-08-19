package mmo

import (
	"sort"
	"strings"
	"time"
)

func (w *WorldState) ensureMarket() *MarketHub {
	if w.Market == nil {
		w.Market = &MarketHub{History: map[string][]MarketHistory{}}
	}
	if w.Market.History == nil {
		w.Market.History = map[string][]MarketHistory{}
	}
	return w.Market
}

func (w *WorldState) marketSearch(p *Player, category, rarity, sortBy string, minP, maxP, level, page int) [][]byte {
	hub := w.ensureMarket()
	pageSize := socialCfg.MarketPage
	if pageSize < 1 {
		pageSize = 20
	}
	if page < 1 {
		page = 1
	}
	minLim, maxLim := socialCfg.MarketMinPrice, socialCfg.MarketMaxPrice
	filtered := make([]MarketListing, 0, len(hub.Listings))
	for _, it := range hub.Listings {
		if category != "" && !strings.EqualFold(it.Category, category) {
			continue
		}
		if rarity != "" && !strings.EqualFold(it.Rarity, rarity) {
			continue
		}
		if minP > 0 && it.Price < minP {
			continue
		}
		if maxP > 0 && it.Price > maxP {
			continue
		}
		if level > 0 && it.Level > level {
			continue
		}
		if minLim > 0 && it.Price < minLim {
			continue
		}
		if maxLim > 0 && it.Price > maxLim {
			continue
		}
		filtered = append(filtered, it)
	}
	switch strings.ToLower(sortBy) {
	case "price_desc", "high":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Price > filtered[j].Price })
	case "newest":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Created > filtered[j].Created })
	default:
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Price < filtered[j].Price })
	}
	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	pageItems := filtered[start:end]
	return [][]byte{marshal(TypeMarketListings, map[string]any{
		"listings": pageItems, "page": page, "pageSize": pageSize, "total": total,
		"history": hub.History[p.ID], "wallet": p.wallet(), "toId": p.ID,
	})}
}

func (w *WorldState) marketList(p *Player, slot int, itemID string, qty, price int, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	p.ensureGear()
	if qty < 1 || price < 1 {
		return rejectFor(p.ID, TypeMarketList, "payload")
	}
	if socialCfg.MarketMinPrice > 0 && price < socialCfg.MarketMinPrice {
		return rejectFor(p.ID, TypeMarketList, "price")
	}
	if socialCfg.MarketMaxPrice > 0 && price > socialCfg.MarketMaxPrice {
		return rejectFor(p.ID, TypeMarketList, "price")
	}
	if slot < 0 || slot >= len(p.Bag.Slots) {
		return rejectFor(p.ID, TypeMarketList, "slot")
	}
	s := p.Bag.Slots[slot]
	if s.ItemID == "" || s.Qty < qty || (itemID != "" && s.ItemID != itemID) {
		return rejectFor(p.ID, TypeMarketList, "item")
	}
	if s.Locked || s.Favorite {
		return rejectFor(p.ID, TypeMarketList, "locked")
	}
	def, ok := itemByID[s.ItemID]
	if !ok || def.Untradable || def.Type == "QUEST" || def.Type == "KEY" {
		return rejectFor(p.ID, TypeMarketList, "tradable")
	}
	feePct := socialCfg.MarketFeePct
	if feePct < 0 {
		feePct = 5
	}
	fee := price * feePct / 100
	if fee < 1 && feePct > 0 {
		fee = 1
	}
	if !w.removeCurrency(p, "coin", fee, "MARKET") {
		return rejectFor(p.ID, TypeMarketList, "fee")
	}
	if !p.Bag.removeAt(slot, qty) {
		w.addCurrency(p, "coin", fee, "market_rollback")
		return rejectFor(p.ID, TypeMarketList, "qty")
	}
	hub := w.ensureMarket()
	listing := MarketListing{
		ID: randomID("mkt_"), SellerID: p.ID, Seller: p.Name, ItemID: s.ItemID,
		Category: def.Type, Rarity: def.Rarity, Qty: qty, Price: price, Level: def.LevelRequirement,
		Created: time.Now().UnixMilli(),
	}
	hub.Listings = append(hub.Listings, listing)
	w.appendMarketHist(p.ID, MarketHistory{ID: listing.ID, Kind: "LIST", ItemID: listing.ItemID, Qty: qty, Price: price, At: listing.Created})
	w.recordEconomy(p.ID, "coin", -fee, "MARKET", tx)
	w.persistGear(p)
	w.persist(p)
	w.saveRuntimeLocked()
	out := append([][]byte{marshal(TypeInventoryUpdated, p.loadout("Listed.", []int{slot}))}, w.marketSearch(p, "", "", "newest", 0, 0, 0, 1)...)
	return w.rememberTx(tx, out)
}

func (w *WorldState) marketBuy(p *Player, listingID, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	if listingID == "" {
		return rejectFor(p.ID, TypeMarketBuy, "listing")
	}
	hub := w.ensureMarket()
	idx := -1
	var it MarketListing
	for i, l := range hub.Listings {
		if l.ID == listingID {
			idx, it = i, l
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeMarketBuy, "gone")
	}
	if it.SellerID == p.ID {
		hub.Flags = append(hub.Flags, "self-trade:"+p.ID)
		return rejectFor(p.ID, TypeMarketBuy, "self")
	}
	if it.Qty < 1 || it.Price < 1 {
		return rejectFor(p.ID, TypeMarketBuy, "payload")
	}
	seller := w.players[it.SellerID]
	if !w.removeCurrency(p, "coin", it.Price, "MARKET") {
		return rejectFor(p.ID, TypeMarketBuy, "coin")
	}
	if !w.giveItem(p, it.ItemID, it.Qty) {
		w.addCurrency(p, "coin", it.Price, "market_rollback")
		return rejectFor(p.ID, TypeMarketBuy, "space")
	}
	if seller != nil {
		w.addCurrency(seller, "coin", it.Price, "MARKET")
		w.persist(seller)
		w.persistGear(seller)
		w.recordEconomy(seller.ID, "coin", it.Price, "MARKET", tx+"-seller")
	} else {
		w.ensureMarket()
		w.appendMarketHist(it.SellerID, MarketHistory{ID: randomID("mh_"), Kind: "SOLD", ItemID: it.ItemID, Other: p.Name, Qty: it.Qty, Price: it.Price, At: time.Now().UnixMilli()})
	}
	hub.Listings = append(hub.Listings[:idx], hub.Listings[idx+1:]...)
	now := time.Now().UnixMilli()
	w.appendMarketHist(p.ID, MarketHistory{ID: randomID("mh_"), Kind: "BOUGHT", ItemID: it.ItemID, Other: it.Seller, Qty: it.Qty, Price: it.Price, At: now})
	w.appendMarketHist(it.SellerID, MarketHistory{ID: randomID("mh_"), Kind: "SOLD", ItemID: it.ItemID, Other: p.Name, Qty: it.Qty, Price: it.Price, At: now})
	w.recordEconomy(p.ID, "coin", -it.Price, "MARKET", tx)
	w.persist(p)
	w.persistGear(p)
	w.saveRuntimeLocked()
	out := append([][]byte{marshal(TypeInventoryUpdated, p.loadout("Purchased.", nil))}, w.marketSearch(p, "", "", "newest", 0, 0, 0, 1)...)
	return w.rememberTx(tx, out)
}

func (w *WorldState) marketCancel(p *Player, listingID string) [][]byte {
	hub := w.ensureMarket()
	idx := -1
	var it MarketListing
	for i, l := range hub.Listings {
		if l.ID == listingID {
			idx, it = i, l
			break
		}
	}
	if idx < 0 || it.SellerID != p.ID {
		return rejectFor(p.ID, TypeMarketCancel, "listing")
	}
	if !w.giveItem(p, it.ItemID, it.Qty) {
		return rejectFor(p.ID, TypeMarketCancel, "space")
	}
	hub.Listings = append(hub.Listings[:idx], hub.Listings[idx+1:]...)
	w.appendMarketHist(p.ID, MarketHistory{ID: listingID, Kind: "CANCELLED", ItemID: it.ItemID, Qty: it.Qty, Price: it.Price, At: time.Now().UnixMilli()})
	w.persistGear(p)
	w.saveRuntimeLocked()
	out := append([][]byte{marshal(TypeInventoryUpdated, p.loadout("Listing cancelled.", nil))}, w.marketSearch(p, "", "", "newest", 0, 0, 0, 1)...)
	return out
}

func (w *WorldState) appendMarketHist(playerID string, row MarketHistory) {
	hub := w.ensureMarket()
	hub.History[playerID] = append(hub.History[playerID], row)
	if len(hub.History[playerID]) > 40 {
		hub.History[playerID] = hub.History[playerID][len(hub.History[playerID])-30:]
	}
}

func (w *WorldState) bankMove(p *Player, deposit bool, slot, qty int) [][]byte {
	p.ensureGear()
	if qty < 1 {
		qty = 1
	}
	src, dst := p.Bag, p.Bank
	if !deposit {
		src, dst = p.Bank, p.Bag
	}
	if src == nil || dst == nil || slot < 0 || slot >= len(src.Slots) {
		return rejectFor(p.ID, TypeBankDeposit, "slot")
	}
	s := src.Slots[slot]
	if s.ItemID == "" || s.Qty < qty {
		return rejectFor(p.ID, TypeBankDeposit, "item")
	}
	if deposit && (s.Locked || s.Favorite) {
		return rejectFor(p.ID, TypeBankDeposit, "locked")
	}
	if !src.removeAt(slot, qty) {
		return rejectFor(p.ID, TypeBankDeposit, "qty")
	}
	if _, ok := dst.add(s.ItemID, qty); !ok {
		src.add(s.ItemID, qty)
		return rejectFor(p.ID, TypeBankDeposit, "space")
	}
	w.persistGear(p)
	kind := TypeBankUpdated
	if deposit {
		kind = TypeBankUpdated
	}
	_ = kind
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Bank.", nil)), marshal(TypeBankUpdated, w.bankView(p))}
}

func (w *WorldState) bankView(p *Player) map[string]any {
	p.ensureGear()
	return map[string]any{"slots": p.Bank.views(), "capacity": p.Bank.Capacity, "wallet": p.wallet(), "toId": p.ID}
}

func (inv *Inventory) views() []InvSlotView {
	if inv == nil {
		return nil
	}
	out := make([]InvSlotView, 0, len(inv.Slots))
	for i, s := range inv.Slots {
		v := InvSlotView{Index: i, Qty: s.Qty, Locked: s.Locked, Favorite: s.Favorite, InstanceID: s.InstanceID, Upgrade: s.Upgrade, ItemLevel: s.ItemLevel}
		if s.ItemID != "" {
			if def, ok := itemByID[s.ItemID]; ok {
				d := def.view()
				v.Item = &d
			}
		}
		out = append(out, v)
	}
	return out
}

func (w *WorldState) toggleItemFlag(p *Player, slot int, lock bool, on bool) [][]byte {
	p.ensureGear()
	if slot < 0 || slot >= len(p.Bag.Slots) || p.Bag.Slots[slot].ItemID == "" {
		return rejectFor(p.ID, TypeLockItem, "slot")
	}
	if lock {
		p.Bag.Slots[slot].Locked = on
	} else {
		p.Bag.Slots[slot].Favorite = on
	}
	w.persistGear(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Item flag.", []int{slot}))}
}
