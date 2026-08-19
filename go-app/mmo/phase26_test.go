package mmo

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPhase26LandminesKept(t *testing.T) {
	if comboRecipe() != "LLLH" {
		t.Fatalf("combo %s", comboRecipe())
	}
	if transformByID["aura-1"].Name != "AURA ASCENSION I" || formDisplayName(transformByID["aura-1"]) != "AWAKENED FORM" {
		t.Fatal("keep AURA ASCENSION I / AWAKENED FORM")
	}
	if questionByID["q-add-4-3"].Correct != 1 || questionByID["q-apel-3-2"].Correct != 1 {
		t.Fatal("keep q-add-4-3 and q-apel-3-2")
	}
	if questionByID["q1"].Correct != 1 {
		t.Fatal("q1 2+3=5 index B")
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
	if partyCap() != 5 {
		t.Fatalf("party cap %d", partyCap())
	}
	if socialCfg.FriendLimit != Phase26FriendLimit {
		t.Fatalf("friend limit %d", socialCfg.FriendLimit)
	}
	if chatRuneMax != Phase26ChatRunes {
		t.Fatalf("chat runes %d", chatRuneMax)
	}
	if guildQuestByID(Phase26GuildQuestID) == nil || guildQuestByID(Phase26GuildQuestID).Title != "Pertahanan Desa." {
		t.Fatal("guild quest Pertahanan Desa.")
	}
	if eventByID[CommunityEventID].Name != "Gotong Royong Desa." {
		t.Fatal("community event")
	}
	if itemByID["relic-soul-dawn"].Untradable != true {
		t.Fatal("soulbound")
	}
}

func TestPhase26FriendRequestNotifies(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	evs := w.ApplySocial(a.ID, Envelope{Type: TypeFriendRequest, Data: []byte(`{"targetId":"p_b"}`)})
	if rejectAction(evs, TypeFriendRequest) {
		t.Fatal("friend request")
	}
	if !hasNoteKind(evs, "friend_request") {
		t.Fatal("target notification")
	}
	if rejectAction(w.ApplySocial(b.ID, Envelope{Type: TypeAcceptFriend, Data: []byte(`{"targetId":"p_a"}`)}), TypeAcceptFriend) {
		t.Fatal("accept")
	}
	if !w.friendsOf(a.ID, b.ID) {
		t.Fatal("friendship")
	}
}

func TestPhase26PartyInviteCreates(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	if rejectAction(w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_b"}`)}), TypePartyInvite) {
		t.Fatal("invite")
	}
	if w.Parties.Of(a.ID) == nil {
		t.Fatal("party created")
	}
	if rejectAction(w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)}), TypePartyAccept) {
		t.Fatal("accept")
	}
	if !rejectAction(w.ApplyParty(b.ID, Envelope{Type: TypePartyKick, Data: []byte(`{"targetId":"p_a"}`)}), TypePartyKick) {
		t.Fatal("member kick must reject")
	}
}

func TestPhase26GuildSecurity(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_mem", "Sinta")
	c := testOnline("p_c", "Budi")
	w.Add(b)
	w.Add(c)
	a.Level = 10
	a.ensureLog().Coin = 1000
	if rejectAction(w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Desa Jaga","tag":"DESA"}`)}), TypeGuildCreate) {
		t.Fatal("create guild")
	}
	if w.guildOf(a.ID) == nil {
		t.Fatal("guild missing")
	}
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildInvite, Data: []byte(`{"targetId":"p_mem"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeGuildAccept, Data: []byte(`{}`)})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildInvite, Data: []byte(`{"targetId":"p_c"}`)})
	w.ApplyEconomy(c.ID, Envelope{Type: TypeGuildAccept, Data: []byte(`{}`)})
	if !rejectAction(w.ApplyEconomy(b.ID, Envelope{Type: TypeGuildKick, Data: []byte(`{"targetId":"p_c"}`)}), TypeGuildKick) {
		t.Fatal("member kick")
	}
	if !rejectAction(w.ApplyEconomy(b.ID, Envelope{Type: TypeGuildDisband, Data: []byte(`{}`)}), TypeGuildDisband) {
		t.Fatal("member disband")
	}
	if !rejectAction(w.ApplyEconomy(b.ID, Envelope{Type: TypeSetGuildRank, Data: []byte(`{"targetId":"p_c","rank":"MASTER"}`)}), TypeSetGuildRank) {
		t.Fatal("fake master")
	}
}

func TestPhase26WorldChatAndFilter(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	evs := w.ApplySocial(a.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"WORLD","text":"halo dunia"}`)})
	msg := chatMsg(evs)
	if msg.Channel != "WORLD" || !containsID(msg.NotifyIDs, b.ID) {
		t.Fatalf("world chat %+v", msg)
	}
	if msg.FromID != a.ID {
		t.Fatal("sender")
	}
	evs = w.ApplySocial(a.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"AREA","text":"area hi"}`)})
	if chatMsg(evs).Channel != "LOCAL" {
		t.Fatal("AREA alias")
	}
	evs = w.ApplySocial(a.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"LOCAL","text":"bodoh sekali"}`)})
	if got := chatMsg(evs).Text; strings.Contains(strings.ToLower(got), "bodoh") {
		t.Fatalf("filter %q", got)
	}
	fake := w.ApplySocial(a.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"LOCAL","text":"aku budi","fromId":"p_b","from":"Sinta"}`)})
	if chatMsg(fake).FromID != a.ID {
		t.Fatal("fake sender")
	}
	link := w.ApplySocial(b.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"WORLD","text":"klik https://evil.test"}`)})
	if got := chatMsg(link).Text; strings.Contains(got, "https://") {
		t.Fatalf("link %q", got)
	}
}

func TestPhase26TradeAtomicDisconnectSoulbound(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	a.ensureGear()
	b.ensureGear()
	before := a.Bag.count("potion_heal")
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeAccept, Data: []byte(`{}`)})
	off, _ := json.Marshal(map[string]any{"slots": []TradeSlotView{{ItemID: "potion_heal", Qty: 1, Slot: 0}}, "coin": 0})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeOffer, Data: off})
	w.Remove(a)
	if w.Trades.ByPlayer[a.ID] != nil || w.Trades.ByPlayer[b.ID] != nil {
		t.Fatal("trade must cancel on disconnect")
	}
	if a.Bag.count("potion_heal") != before {
		t.Fatal("rollback items")
	}
	a.Connected = true
	w.players[a.ID] = a
	a.Bag.add("relic-soul-dawn", 1)
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeAccept, Data: []byte(`{}`)})
	soul, _ := json.Marshal(map[string]any{"slots": []TradeSlotView{{ItemID: "relic-soul-dawn", Qty: 1}}, "coin": 0})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeOffer, Data: soul})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeReady, Data: []byte(`{}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeReady, Data: []byte(`{}`)})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tx-soul"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeConfirm, Data: []byte(`{"transactionId":"tx-soul"}`)})
	if b.Bag.count("relic-soul-dawn") > 0 {
		t.Fatal("soulbound transferred")
	}
}

func TestPhase26BlockAndPrivacy(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	w.ApplySocial(a.ID, Envelope{Type: TypeBlockPlayer, Data: []byte(`{"targetId":"p_b"}`)})
	if !rejectAction(w.ApplySocial(b.ID, Envelope{Type: TypeFriendRequest, Data: []byte(`{"targetId":"p_a"}`)}), TypeFriendRequest) {
		t.Fatal("block friend")
	}
	if !rejectAction(w.ApplySocial(b.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"WHISPER","target":"Raka","text":"hi"}`)}), TypeChat) {
		t.Fatal("block whisper")
	}
	if !rejectAction(w.ApplyParty(b.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_a"}`)}), TypePartyInvite) {
		t.Fatal("block party")
	}
	if !rejectAction(w.ApplyEconomy(b.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_a"}`)}), TypeTradeRequest) {
		t.Fatal("block trade")
	}
	w.ApplySocial(a.ID, Envelope{Type: TypeUnblockPlayer, Data: []byte(`{"targetId":"p_b"}`)})
	c := testOnline("p_c", "Mira")
	w.Add(c)
	w.ApplySocial(c.ID, Envelope{Type: TypeSetPrivacy, Data: []byte(`{"friend":"NONE","party":"NONE","trade":"NONE","pm":"NONE"}`)})
	if !rejectAction(w.ApplySocial(a.ID, Envelope{Type: TypeFriendRequest, Data: []byte(`{"targetId":"p_c"}`)}), TypeFriendRequest) {
		t.Fatal("privacy friend")
	}
	if !rejectAction(w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_c"}`)}), TypePartyInvite) {
		t.Fatal("privacy party")
	}
	if !rejectAction(w.ApplyEconomy(a.ID, Envelope{Type: TypeTradeRequest, Data: []byte(`{"targetId":"p_c"}`)}), TypeTradeRequest) {
		t.Fatal("privacy trade")
	}
	if !rejectAction(w.ApplySocial(a.ID, Envelope{Type: TypeChat, Data: []byte(`{"channel":"PRIVATE","target":"Mira","text":"halo"}`)}), TypeChat) {
		t.Fatal("privacy pm")
	}
}

func TestPhase26AntiCheatAndModeration(t *testing.T) {
	w, p := testVillagePlayer()
	if !rejectAction(w.ApplySocial(p.ID, Envelope{Type: TypeSetFriendship, Data: []byte(`{}`)}), TypeSetFriendship) {
		t.Fatal("SET_FRIENDSHIP")
	}
	if !rejectAction(w.ApplySocial(p.ID, Envelope{Type: TypeSetPartyLeader, Data: []byte(`{}`)}), TypeSetPartyLeader) {
		t.Fatal("SET_PARTY_LEADER")
	}
	if !rejectAction(w.ApplySocial(p.ID, Envelope{Type: TypeSetTradeItem, Data: []byte(`{}`)}), TypeSetTradeItem) {
		t.Fatal("SET_TRADE_ITEM")
	}
	if !rejectAction(w.ApplySocial(p.ID, Envelope{Type: TypeSetChatSender, Data: []byte(`{}`)}), TypeSetChatSender) {
		t.Fatal("SET_CHAT_SENDER")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeSetCoin, Data: []byte(`{"coin":9}`)}), TypeSetCoin) {
		t.Fatal("SET_COIN")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeMutePlayer, Data: []byte(`{"targetId":"x"}`)}), TypeMutePlayer) {
		t.Fatal("MUTE_PLAYER")
	}
	if !reportCategoryOK("SCAM") || !reportCategoryOK("INAPPROPRIATE") || !reportCategoryOK("OTHER") {
		t.Fatal("report reasons")
	}
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeReportMessage, Data: []byte(`{"targetId":"x","category":"Spam","evidence":"chat"}`)}), TypeReportMessage) {
		t.Fatal("report message")
	}
}

func TestPhase26PresenceInspectPartyStay(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	w.ApplyParty(a.ID, Envelope{Type: TypePartyCreate, Data: []byte(`{}`)})
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	pty := w.Parties.Of(a.ID)
	if pty == nil {
		t.Fatal("party")
	}
	w.ApplySocial(a.ID, Envelope{Type: TypeSetPresence, Data: []byte(`{"status":"AWAY"}`)})
	if w.presenceOf(a) != "AWAY" {
		t.Fatalf("presence %s", w.presenceOf(a))
	}
	evs := w.ApplySocial(a.ID, Envelope{Type: TypeInspectPlayer, Data: []byte(`{"targetId":"p_b"}`)})
	var env Envelope
	for _, ev := range evs {
		_ = json.Unmarshal(ev, &env)
		if env.Type == TypeInspectResult {
			var out InspectOut
			_ = json.Unmarshal(env.Data, &out)
			if out.Region == "" || out.PlayerID != b.ID {
				t.Fatalf("inspect %+v", out)
			}
		}
	}
	w.Remove(b)
	if w.Parties.Get(pty.ID) == nil {
		t.Fatal("party dropped on disconnect")
	}
	if w.guildOf(a.ID) != nil {
		t.Fatal("no guild expected")
	}
}

func hasNoteKind(evs [][]byte, kind string) bool {
	for _, ev := range evs {
		var env Envelope
		if json.Unmarshal(ev, &env) != nil || env.Type != TypeSocialNotification {
			continue
		}
		var n SocialNote
		_ = json.Unmarshal(env.Data, &n)
		if n.Kind == kind {
			return true
		}
	}
	return false
}

func chatMsg(evs [][]byte) ChatOut {
	var msg ChatOut
	for _, ev := range evs {
		var env Envelope
		_ = json.Unmarshal(ev, &env)
		if env.Type == TypeChatMessage {
			_ = json.Unmarshal(env.Data, &msg)
		}
	}
	return msg
}
