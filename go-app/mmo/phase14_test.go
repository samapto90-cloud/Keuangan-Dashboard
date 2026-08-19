package mmo

import (
	"encoding/json"
	"testing"
)

func addWorldPlayer(w *WorldState, id, name string) *Player {
	p := &Player{ID: id, Name: name, send: make(chan []byte, 32)}
	p.initCombat()
	w.Add(p)
	return p
}

func TestPhase14MarketPurchase(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_buy", "Bima")
	a.ensureLog().Coin = 500
	b.ensureLog().Coin = 200
	if !w.giveItem(a, "iron_ore", 10) {
		t.Fatal("give ore")
	}
	slot := a.Bag.firstIndex("iron_ore")
	raw, _ := json.Marshal(map[string]any{"slot": slot, "itemId": "iron_ore", "qty": 10, "price": 40, "transactionId": "mkt-list-1"})
	if rejectAction(w.ApplyPhase14(a.ID, Envelope{Type: TypeMarketList, Data: raw}), TypeMarketList) {
		t.Fatal("list rejected")
	}
	if len(w.Market.Listings) != 1 {
		t.Fatal("listing missing")
	}
	lid := w.Market.Listings[0].ID
	buy, _ := json.Marshal(map[string]any{"listingId": lid, "transactionId": "mkt-buy-1"})
	if rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeMarketBuy, Data: buy}), TypeMarketBuy) {
		t.Fatal("buy rejected")
	}
	if b.Bag.count("iron_ore") != 10 {
		t.Fatalf("buyer ore %d", b.Bag.count("iron_ore"))
	}
	if a.ensureLog().Coin != 538 {
		t.Fatalf("seller coin %d", a.ensureLog().Coin)
	}
	if b.ensureLog().Coin != 160 {
		t.Fatalf("buyer coin %d", b.ensureLog().Coin)
	}
	if len(w.Market.Listings) != 0 {
		t.Fatal("listing should be gone")
	}
	if rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeMarketBuy, Data: buy}), TypeMarketBuy) {
		t.Fatal("duplicate buy should replay not reject")
	}
	if b.ensureLog().Coin != 160 {
		t.Fatalf("dup buy coin %d", b.ensureLog().Coin)
	}
}

func TestPhase14GuildStorageLog(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_off", "Sam")
	a.Level = 10
	a.ensureLog().Coin = 1000
	if rejectAction(w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Iron Hand","tag":"IRON"}`)}), TypeGuildCreate) {
		t.Fatal("create")
	}
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildInvite, Data: []byte(`{"targetId":"p_off"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeGuildAccept, Data: []byte(`{}`)})
	if !w.giveItem(b, "iron_ore", 10) {
		t.Fatal("ore")
	}
	slot := b.Bag.firstIndex("iron_ore")
	dep, _ := json.Marshal(map[string]any{"slot": slot, "qty": 10})
	if rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeGuildDeposit, Data: dep}), TypeGuildDeposit) {
		t.Fatal("deposit")
	}
	g := w.guildOf(a.ID)
	if g == nil || g.Storage == nil || g.Storage.count("iron_ore") != 10 {
		t.Fatal("storage empty")
	}
	stSlot := g.Storage.firstIndex("iron_ore")
	wd, _ := json.Marshal(map[string]any{"slot": stSlot, "qty": 10})
	if rejectAction(w.ApplyPhase14(a.ID, Envelope{Type: TypeGuildWithdraw, Data: wd}), TypeGuildWithdraw) {
		t.Fatal("withdraw")
	}
	if len(g.Logs) < 2 {
		t.Fatalf("logs %d", len(g.Logs))
	}
}

func TestPhase14PartyInviteCreatesParty(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_pt", "Nara")
	if rejectAction(w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_pt"}`)}), TypePartyInvite) {
		t.Fatal("invite")
	}
	if rejectAction(w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)}), TypePartyAccept) {
		t.Fatal("accept")
	}
	pt := w.Parties.Of(a.ID)
	if pt == nil || len(pt.Members) != 2 {
		t.Fatal("party not created")
	}
}

func TestPhase14FriendAccept(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_fr", "Kira")
	w.ApplySocial(a.ID, Envelope{Type: TypeFriendRequest, Data: []byte(`{"targetId":"p_fr"}`)})
	if rejectAction(w.ApplySocial(b.ID, Envelope{Type: TypeAcceptFriend, Data: []byte(`{"targetId":"p_a"}`)}), TypeAcceptFriend) {
		t.Fatal("accept friend")
	}
	if w.Social.Friends[a.ID]["p_fr"].Since.IsZero() {
		t.Fatal("friend list not updated")
	}
}

func TestPhase14HousingPersist(t *testing.T) {
	w, p := testVillagePlayer()
	if !w.giveItem(p, "decor-chair", 1) {
		t.Fatal("chair")
	}
	if rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeHouseEnter, Data: []byte(`{}`)}), TypeHouseEnter) {
		t.Fatal("enter")
	}
	slot := p.Bag.firstIndex("decor-chair")
	place, _ := json.Marshal(map[string]any{"slot": slot, "itemId": "decor-chair", "x": 1.0, "z": 2.0, "yaw": 0.5})
	if rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeHousePlace, Data: place}), TypeHousePlace) {
		t.Fatal("place")
	}
	blob := w.exportRuntime()
	w2 := NewWorldState()
	w2.importRuntime(blob)
	h := w2.Housing.ByOwner[p.ID]
	if h == nil || len(h.Items) != 1 || h.Items[0].ItemID != "decor-chair" {
		t.Fatal("chair missing after reconnect")
	}
}

func TestPhase14EconomyIdempotentSpend(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().Coin = 1000
	w.mu.Lock()
	if !w.economyDebit(p, "coin", 100, "SHOP", "spend-100") {
		w.mu.Unlock()
		t.Fatal("debit")
	}
	if p.ensureLog().Coin != 900 {
		w.mu.Unlock()
		t.Fatalf("coin %d", p.ensureLog().Coin)
	}
	if !w.economyDebit(p, "coin", 100, "SHOP", "spend-100") {
		w.mu.Unlock()
		t.Fatal("dup debit")
	}
	w.mu.Unlock()
	if p.ensureLog().Coin != 900 {
		t.Fatalf("dup coin %d", p.ensureLog().Coin)
	}
}

func TestPhase14SecurityRejects(t *testing.T) {
	w, p := testVillagePlayer()
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeSetCoin, Data: []byte(`{"coin":999999999}`)}), TypeSetCoin) {
		t.Fatal("coin")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeSetGuildRank, Data: []byte(`{"rank":"GUILD_MASTER"}`)}), TypeSetGuildRank) {
		t.Fatal("role")
	}
	if !rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeSetOwned, Data: []byte(`{"itemOwned":true}`)}), TypeSetOwned) {
		t.Fatal("owned")
	}
	if !rejectAction(w.ApplySocial(p.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"SYSTEM","text":"hack"}`)}), TypeChat) {
		t.Fatal("system chat")
	}
}

func TestPhase14BlockWhisper(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_blk", "Vex")
	w.ApplySocial(a.ID, Envelope{Type: TypeBlockPlayer, Data: []byte(`{"targetId":"p_blk"}`)})
	if !rejectAction(w.ApplySocial(b.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"WHISPER","target":"Raka","text":"hi"}`)}), TypeChat) {
		t.Fatal("whisper should reject")
	}
}

func TestPhase14BankAndLock(t *testing.T) {
	w, p := testVillagePlayer()
	if !w.giveItem(p, "iron_ore", 5) {
		t.Fatal("ore")
	}
	slot := p.Bag.firstIndex("iron_ore")
	lock, _ := json.Marshal(map[string]any{"slot": slot, "on": true})
	w.ApplyPhase14(p.ID, Envelope{Type: TypeLockItem, Data: lock})
	dep, _ := json.Marshal(map[string]any{"slot": slot, "qty": 1})
	if !rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeBankDeposit, Data: dep}), TypeBankDeposit) {
		t.Fatal("locked deposit")
	}
	unlock, _ := json.Marshal(map[string]any{"slot": slot, "on": false})
	w.ApplyPhase14(p.ID, Envelope{Type: TypeLockItem, Data: unlock})
	if rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeBankDeposit, Data: dep}), TypeBankDeposit) {
		t.Fatal("bank deposit")
	}
	if p.Bank.count("iron_ore") != 1 {
		t.Fatalf("bank %d", p.Bank.count("iron_ore"))
	}
}
