package mmo

import (
	"encoding/json"
	"testing"
)

func testOnline(id, name string) *Player {
	p := &Player{ID: id, Name: name, send: make(chan []byte, 32)}
	p.initCombat()
	return p
}

func TestSetCoinRejected(t *testing.T) {
	w, p := testVillagePlayer()
	evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeSetCoin, Data: []byte(`{"coin":999999999}`)})
	if !rejectAction(evs, TypeSetCoin) {
		t.Fatal("SET_COIN harus ditolak")
	}
	if p.ensureLog().Coin == 999999999 {
		t.Fatal("wallet client tidak boleh menang")
	}
}

func TestSetGuildRankRejected(t *testing.T) {
	w, p := testVillagePlayer()
	evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeSetGuildRank, Data: []byte(`{"rank":"LEADER"}`)})
	if !rejectAction(evs, TypeSetGuildRank) {
		t.Fatal("SET_GUILD_RANK harus ditolak")
	}
}

func TestGuildCreateLevelAndCoin(t *testing.T) {
	w, p := testVillagePlayer()
	p.Level = 5
	p.ensureLog().Coin = 1000
	evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Dawn Guard","tag":"DAWN"}`)})
	if !rejectAction(evs, TypeGuildCreate) {
		t.Fatal("level 5 harus ditolak")
	}
	p.Level = 10
	evs = w.ApplyEconomy(p.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Dawn Guard","tag":"DAWN"}`)})
	if rejectAction(evs, TypeGuildCreate) {
		t.Fatal("level 10 + coin harus sukses")
	}
	if p.ensureLog().Coin != 0 {
		t.Fatalf("coin tersisa %d", p.ensureLog().Coin)
	}
	if w.guildOf(p.ID) == nil || w.guildOf(p.ID).Tag != "DAWN" {
		t.Fatal("guild tidak terbentuk")
	}
}

func TestTradeConfirmRequiresReady(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	b.X, b.Z = 1, 3
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeAccept, Data: []byte(`{}`)})
	evs := w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{}`)})
	if !rejectAction(evs, TypeTradeConfirm) {
		t.Fatal("confirm tanpa READY harus ditolak")
	}
}

func TestTradeRejectsOtherPlayerItem(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	a.ensureGear()
	b.ensureGear()
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeAccept, Data: []byte(`{}`)})
	raw, _ := json.Marshal(map[string]any{"slots": []TradeSlotView{{Slot: 0, ItemID: "potion_heal", Qty: 1}}, "coin": 0})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeOffer, Data: raw})
	steal, _ := json.Marshal(map[string]any{"slots": []TradeSlotView{{Slot: 0, ItemID: "potion_heal", Qty: 99}}, "coin": 0})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeOffer, Data: steal})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeReady, Data: []byte(`{}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeReady, Data: []byte(`{}`)})
	evs := w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tx-steal"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tx-steal"}`)})
	if a.Bag.count("potion_heal") < 3 {
		t.Fatal("item A tidak boleh hilang pada trade invalid")
	}
	_ = evs
}

func TestTradeIdempotent(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	a.ensureLog().Coin = 40
	b.ensureLog().Coin = 0
	a.ensureGear()
	b.ensureGear()
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeAccept, Data: []byte(`{}`)})
	off, _ := json.Marshal(map[string]any{"slots": []TradeSlotView{{ItemID: "potion_heal", Qty: 1}}, "coin": 10})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeOffer, Data: off})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeOffer, Data: []byte(`{"slots":[],"coin":0}`)})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeReady, Data: []byte(`{}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeReady, Data: []byte(`{}`)})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tx-dup"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tx-dup"}`)})
	first := a.Bag.count("potion_heal")
	coinB := b.ensureLog().Coin
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tx-dup"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tx-dup"}`)})
	if a.Bag.count("potion_heal") != first {
		t.Fatal("trade duplikat menggandakan item")
	}
	if b.ensureLog().Coin != coinB {
		t.Fatal("trade duplikat menggandakan coin")
	}
}

func TestShopBuyIdempotent(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().Coin = 100
	p.X, p.Z = 11.2, 4.8
	buy := []byte(`{"shopItemId":"shop-potion","itemId":"potion_heal","transactionId":"buy-1"}`)
	before := p.Bag.count("potion_heal")
	w.ApplyEconomy(p.ID, Envelope{Type: TypeShopBuy, Data: buy})
	mid := p.Bag.count("potion_heal")
	w.ApplyEconomy(p.ID, Envelope{Type: TypeShopBuy, Data: buy})
	if p.Bag.count("potion_heal") != mid || mid != before+1 {
		t.Fatalf("shop duplikat potion %d -> %d -> %d", before, mid, p.Bag.count("potion_heal"))
	}
}

func TestLocalAndPartyChat(t *testing.T) {
	w, a := testVillagePlayer()
	near := testOnline("p_n", "Near")
	far := testOnline("p_f", "Far")
	w.Add(near)
	w.Add(far)
	a.X, a.Z = 0, 0
	near.X, near.Z = 2, 2
	far.X, far.Z = 80, 80
	evs := w.ApplySocial(a.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"LOCAL","text":"halo"}`)})
	var msg ChatOut
	for _, ev := range evs {
		var env Envelope
		_ = json.Unmarshal(ev, &env)
		if env.Type == TypeChatMessage {
			_ = json.Unmarshal(env.Data, &msg)
		}
	}
	if !containsID(msg.NotifyIDs, near.ID) || containsID(msg.NotifyIDs, far.ID) {
		t.Fatalf("local radius salah: %v", msg.NotifyIDs)
	}
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_f"}`)})
	w.ApplyParty(far.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	evs = w.ApplySocial(a.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"PARTY","text":"party hi"}`)})
	msg = ChatOut{}
	for _, ev := range evs {
		var env Envelope
		_ = json.Unmarshal(ev, &env)
		if env.Type == TypeChatMessage {
			_ = json.Unmarshal(env.Data, &msg)
		}
	}
	if !containsID(msg.NotifyIDs, far.ID) {
		t.Fatal("party chat harus sampai anggota jauh")
	}
}

func TestBlockRejectsSocial(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	w.ApplySocial(a.ID, Envelope{Type: TypeBlockPlayer, Data: []byte(`{"targetId":"p_b"}`)})
	if !rejectAction(w.ApplySocial(b.ID, Envelope{Type: TypeFriendRequest, Data: []byte(`{"targetId":"p_a"}`)}), TypeFriendRequest) {
		t.Fatal("friend request blocked harus ditolak")
	}
	if !rejectAction(w.ApplySocial(b.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"WHISPER","target":"Raka","text":"hi"}`)}), TypeChat) {
		t.Fatal("whisper blocked harus ditolak")
	}
	if !rejectAction(w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_a"}`)}), TypeTradeRequest) {
		t.Fatal("trade blocked harus ditolak")
	}
	if !rejectAction(w.ApplyParty(b.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_a"}`)}), TypePartyInvite) {
		t.Fatal("party invite blocked harus ditolak")
	}
}

func TestReconnectKeepsSocialWallet(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	a.Level = 10
	a.ensureLog().Coin = 1500
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Star Guard","tag":"STAR"}`)})
	a.grantTitle("forest-explorer")
	w.ApplySocial(a.ID, Envelope{Type: TypeFriendRequest, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplySocial(b.ID, Envelope{Type: TypeAcceptFriend, Data: []byte(`{"targetId":"p_a"}`)})
	gid := a.GuildID
	coin := a.ensureLog().Coin
	w.Remove(a)
	again := testOnline("p_a", "Raka")
	w.Add(again)
	if again.GuildID != gid {
		t.Fatalf("guild reconnect %s", again.GuildID)
	}
	if again.ensureLog().Coin != coin {
		t.Fatalf("wallet reconnect %d", again.ensureLog().Coin)
	}
	if w.Social.Friends[again.ID]["p_b"].Since.IsZero() {
		t.Fatal("friends reconnect hilang")
	}
	owned := false
	for _, tlt := range again.ensureLog().Titles {
		if tlt == "forest-explorer" || tlt == "dawn-defender" {
			owned = true
		}
	}
	if !owned {
		t.Fatal("title reconnect hilang")
	}
}

func TestRuntimeRestartPersistence(t *testing.T) {
	w, a := testVillagePlayer()
	a.Level = 10
	a.ensureLog().Coin = 1000
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Raven Guard","tag":"RAVEN"}`)})
	blob := w.exportRuntime()
	repo, inv := w.QuestRepo, w.InvRepo
	w2 := NewWorldState()
	w2.QuestRepo, w2.InvRepo = repo, inv
	w2.importRuntime(blob)
	p := testOnline("p_a", "Raka")
	w2.Add(p)
	if w2.guildOf(p.ID) == nil || w2.guildOf(p.ID).Tag != "RAVEN" {
		t.Fatal("guild harus tetap setelah restart")
	}
	if p.ensureLog().Coin != 0 {
		t.Fatalf("wallet restart %d", p.ensureLog().Coin)
	}
}

func TestGuildChatReachesMembers(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	a.Level = 10
	a.ensureLog().Coin = 1000
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Guard Order","tag":"GUARD"}`)})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildInvite, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeGuildAccept, Data: []byte(`{}`)})
	evs := w.ApplySocial(a.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"GUILD","text":"guild hi"}`)})
	var msg ChatOut
	for _, ev := range evs {
		var env Envelope
		_ = json.Unmarshal(ev, &env)
		if env.Type == TypeChatMessage {
			_ = json.Unmarshal(env.Data, &msg)
		}
	}
	if !containsID(msg.NotifyIDs, b.ID) {
		t.Fatal("guild chat harus sampai anggota")
	}
}
