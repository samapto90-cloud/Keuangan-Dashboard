package mmo

import "time"

type TradeOffer struct {
	Slots   []TradeSlotView
	Coin    int
	Ready   bool
	Confirm bool
}

type TradeSession struct {
	ID, TxID, AID, BID string
	A, B               TradeOffer
	Done               bool
}

type TradeHub struct {
	ByPlayer map[string]*TradeSession
	Pending  map[string]string
}

func NewTradeHub() *TradeHub {
	return &TradeHub{ByPlayer: map[string]*TradeSession{}, Pending: map[string]string{}}
}

func (w *WorldState) ApplyTrade(id string, env Envelope) [][]byte {
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	if w.Trades == nil {
		w.Trades = NewTradeHub()
	}
	var in struct {
		TargetID      string          `json:"targetId"`
		Slots         []TradeSlotView `json:"slots"`
		Coin          int             `json:"coin"`
		TransactionID string          `json:"transactionId"`
	}
	_ = unmarshal(env.Data, &in)
	switch env.Type {
	case TypeTradeRequest:
		return w.tradeRequest(p, in.TargetID)
	case TypeTradeAccept:
		return w.tradeAccept(p)
	case TypeTradeDecline, TypeTradeCancel:
		return w.tradeCancel(p, "cancel")
	case TypeTradeOffer:
		return w.tradeOffer(p, in.Slots, in.Coin)
	case TypeTradeReady:
		return w.tradeReady(p)
	case TypeTradeConfirm:
		return w.tradeConfirm(p, in.TransactionID)
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) tradeRequest(from *Player, targetID string) [][]byte {
	to := w.players[targetID]
	if to == nil || !to.Connected {
		return rejectFor(from.ID, TypeTradeRequest, "target")
	}
	if from.ID == to.ID {
		return rejectFor(from.ID, TypeTradeRequest, "self")
	}
	if w.Social != nil && w.Social.blocked(from.ID, to.ID) {
		return rejectFor(from.ID, TypeTradeRequest, "blocked")
	}
	if !w.privacyAllows(to.ensureLog().privacyTrade(), from.ID, to.ID) {
		return rejectFor(from.ID, TypeTradeRequest, "privacy")
	}
	if w.limited("trade:"+from.ID, 6, time.Minute) {
		return rejectFor(from.ID, TypeTradeRequest, "rate")
	}
	if w.Trades.ByPlayer[from.ID] != nil || w.Trades.ByPlayer[to.ID] != nil {
		return rejectFor(from.ID, TypeTradeRequest, "busy")
	}
	w.Trades.Pending[to.ID] = from.ID
	ev := w.notify(to, "trade_request", from.Name+" wants to trade.")
	return [][]byte{marshal(TypeTradeRequest, map[string]any{"fromId": from.ID, "from": from.Name, "toId": to.ID}), ev}
}

func (w *WorldState) tradeAccept(p *Player) [][]byte {
	fromID := w.Trades.Pending[p.ID]
	from := w.players[fromID]
	if from == nil {
		return rejectFor(p.ID, TypeTradeAccept, "request")
	}
	delete(w.Trades.Pending, p.ID)
	s := &TradeSession{ID: randomID("trd_"), TxID: randomID("tx_"), AID: from.ID, BID: p.ID}
	w.Trades.ByPlayer[from.ID] = s
	w.Trades.ByPlayer[p.ID] = s
	view := w.tradeView(s)
	return [][]byte{marshal(TypeTradeUpdated, view)}
}

func (w *WorldState) tradeCancel(p *Player, reason string) [][]byte {
	delete(w.Trades.Pending, p.ID)
	s := w.Trades.ByPlayer[p.ID]
	if s == nil {
		return [][]byte{marshal(TypeTradeUpdated, TradeView{State: "CANCEL", Result: reason})}
	}
	delete(w.Trades.ByPlayer, s.AID)
	delete(w.Trades.ByPlayer, s.BID)
	view := w.tradeView(s)
	view.State, view.Result = "CANCEL", reason
	return [][]byte{marshal(TypeTradeUpdated, view)}
}

func (w *WorldState) cancelTradeOf(id string) {
	if w.Trades == nil {
		return
	}
	if s := w.Trades.ByPlayer[id]; s != nil {
		delete(w.Trades.ByPlayer, s.AID)
		delete(w.Trades.ByPlayer, s.BID)
	}
	delete(w.Trades.Pending, id)
}

func (w *WorldState) tradeOffer(p *Player, slots []TradeSlotView, coin int) [][]byte {
	s := w.Trades.ByPlayer[p.ID]
	if s == nil {
		return rejectFor(p.ID, TypeTradeOffer, "session")
	}
	if s.A.Confirm || s.B.Confirm {
		return rejectFor(p.ID, TypeTradeOffer, "locked")
	}
	if coin < 0 || len(slots) > 5 {
		return rejectFor(p.ID, TypeTradeOffer, "offer")
	}
	off := TradeOffer{Slots: slots, Coin: coin}
	if p.ID == s.AID {
		s.A = off
	} else {
		s.B = off
	}
	s.A.Ready, s.B.Ready, s.A.Confirm, s.B.Confirm = false, false, false, false
	return [][]byte{marshal(TypeTradeUpdated, w.tradeView(s))}
}

func (w *WorldState) tradeReady(p *Player) [][]byte {
	s := w.Trades.ByPlayer[p.ID]
	if s == nil {
		return rejectFor(p.ID, TypeTradeReady, "session")
	}
	if p.ID == s.AID {
		s.A.Ready = true
	} else {
		s.B.Ready = true
	}
	return [][]byte{marshal(TypeTradeUpdated, w.tradeView(s))}
}

func (w *WorldState) tradeConfirm(p *Player, txID string) [][]byte {
	if prev, ok := w.txSeen(txID); ok {
		return prev
	}
	s := w.Trades.ByPlayer[p.ID]
	if s == nil {
		return rejectFor(p.ID, TypeTradeConfirm, "session")
	}
	if txID == "" {
		txID = s.TxID
	}
	if prev, ok := w.txSeen(txID); ok {
		return prev
	}
	if p.ID == s.AID {
		s.A.Confirm = true
	} else {
		s.B.Confirm = true
	}
	if !s.A.Ready || !s.B.Ready {
		return rejectFor(p.ID, TypeTradeConfirm, "ready")
	}
	if !s.A.Confirm || !s.B.Confirm {
		return [][]byte{marshal(TypeTradeUpdated, w.tradeView(s))}
	}
	a, b := w.players[s.AID], w.players[s.BID]
	if a == nil || b == nil || !a.Connected || !b.Connected {
		return w.tradeCancel(p, "offline")
	}
	if err := w.applyTradeAtomic(s, a, b); err != "" {
		return rejectFor(p.ID, TypeTradeConfirm, err)
	}
	s.Done = true
	view := w.tradeView(s)
	view.State, view.Result = "DONE", "ok"
	out := [][]byte{marshal(TypeTradeUpdated, view), marshal(TypeInventoryUpdated, a.loadout("", nil)), marshal(TypeInventoryUpdated, b.loadout("", nil))}
	delete(w.Trades.ByPlayer, s.AID)
	delete(w.Trades.ByPlayer, s.BID)
	w.audit("TRADE", a.ID, s.ID+":"+b.ID)
	w.TradeVolume++
	w.TradeLogs = append(w.TradeLogs, TradeLogRow{
		ID: s.ID, AID: a.ID, BID: b.ID, At: time.Now().UnixMilli(),
	})
	if len(w.TradeLogs) > 200 {
		w.TradeLogs = w.TradeLogs[len(w.TradeLogs)-120:]
	}
	return w.rememberTx(txID, out)
}

func (w *WorldState) applyTradeAtomic(s *TradeSession, a, b *Player) string {
	if !w.canGiveOffer(a, s.A) || !w.canGiveOffer(b, s.B) {
		return "offer"
	}
	if !w.canReceiveOffer(a, s.B) || !w.canReceiveOffer(b, s.A) {
		return "space"
	}
	if !w.takeOffer(a, s.A) {
		return "take"
	}
	if !w.takeOffer(b, s.B) {
		w.giveOffer(a, s.A)
		return "take"
	}
	w.giveOffer(a, s.B)
	w.giveOffer(b, s.A)
	w.persist(a)
	w.persist(b)
	w.persistGear(a)
	w.persistGear(b)
	return ""
}

func (w *WorldState) canGiveOffer(p *Player, o TradeOffer) bool {
	if o.Coin > 0 && currencyOf(p.ensureLog(), "coin") < o.Coin {
		return false
	}
	p.ensureGear()
	need := map[string]int{}
	for _, s := range o.Slots {
		if s.Qty < 1 || s.ItemID == "" {
			continue
		}
		if itemByID[s.ItemID].Untradable || itemByID[s.ItemID].ID == "" {
			return false
		}
		if s.Slot >= 0 && s.Slot < len(p.Bag.Slots) {
			if p.Bag.Slots[s.Slot].Locked || p.Bag.Slots[s.Slot].Favorite {
				return false
			}
		}
		need[s.ItemID] += s.Qty
	}
	for id, n := range need {
		if p.Bag.count(id) < n {
			return false
		}
	}
	return true
}

func (w *WorldState) canReceiveOffer(p *Player, o TradeOffer) bool {
	p.ensureGear()
	for _, s := range o.Slots {
		if s.Qty < 1 {
			continue
		}
		if !p.Bag.canFit(s.ItemID, s.Qty) {
			return false
		}
	}
	return true
}

func (w *WorldState) takeOffer(p *Player, o TradeOffer) bool {
	if o.Coin > 0 && !w.removeCurrency(p, "coin", o.Coin, "trade") {
		return false
	}
	for _, s := range o.Slots {
		if s.Qty < 1 {
			continue
		}
		if !p.Bag.takeItem(s.ItemID, s.Qty) {
			return false
		}
	}
	return true
}

func (w *WorldState) giveOffer(p *Player, o TradeOffer) {
	if o.Coin > 0 {
		w.addCurrency(p, "coin", o.Coin, "trade")
	}
	for _, s := range o.Slots {
		if s.Qty > 0 {
			w.giveItem(p, s.ItemID, s.Qty)
		}
	}
}

func (w *WorldState) tradeView(s *TradeSession) TradeView {
	if s == nil {
		return TradeView{}
	}
	return TradeView{
		ID: s.ID, TxID: s.TxID, State: "OPEN", NotifyIDs: []string{s.AID, s.BID},
		A: TradeOfferView{PlayerID: s.AID, Slots: s.A.Slots, Coin: s.A.Coin, Ready: s.A.Ready, Confirm: s.A.Confirm},
		B: TradeOfferView{PlayerID: s.BID, Slots: s.B.Slots, Coin: s.B.Coin, Ready: s.B.Ready, Confirm: s.B.Confirm},
	}
}
