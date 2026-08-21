package mmo

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRankPromotionGoldIII(t *testing.T) {
	st := RankState{Tier: "GOLD", Division: "III", RR: 99}
	st = ApplyRank(st, true)
	if st.Tier != "GOLD" || st.Division != "II" {
		t.Fatalf("want GOLD II got %s %s %d", st.Tier, st.Division, st.RR)
	}
	if st.RR < 0 {
		t.Fatal("negative")
	}
}

func TestRankDemotionProtectionAndFloor(t *testing.T) {
	st := defaultRank()
	st = ApplyRank(st, false)
	if st.Index != 0 || st.RR != 0 || st.Tier != "BRONZE" {
		t.Fatalf("floor %+v", st)
	}
	st = RankState{Tier: "GOLD", Division: "II", RR: 0, Protect: true}
	st = ApplyRank(st, false)
	if st.Tier != "GOLD" || st.Division != "II" || st.RR != 0 {
		t.Fatalf("protect %+v", st)
	}
	st = ApplyRank(st, false)
	if st.Index < 0 {
		t.Fatal("negative index")
	}
	if st.RR < 0 {
		t.Fatal("negative rr")
	}
}

func TestRankNeverBronzeVsMaster(t *testing.T) {
	a := RankState{Tier: "BRONZE", Division: "III", RR: 10}.Sync()
	b := RankState{Tier: "MASTER", RR: 10}.Sync()
	if RanksCompatible(a, b, 200) {
		t.Fatal("bronze vs master")
	}
}

func TestFriendAcceptUnique(t *testing.T) {
	st := OpenSocialStore(filepath.Join(t.TempDir(), "s.json"))
	if _, msg := st.RequestFriend("a", "b"); msg != "" {
		t.Fatal(msg)
	}
	in, _ := st.PendingFor("b")
	if len(in) != 1 {
		t.Fatalf("incoming %d", len(in))
	}
	if _, msg := st.RespondFriend("b", in[0].ID, "accept"); msg != "" {
		t.Fatal(msg)
	}
	if !st.AreFriends("a", "b") || !st.AreFriends("b", "a") {
		t.Fatal("friendship")
	}
	n := 0
	for _, f := range st.blob.Friends {
		if (f.UserA == "a" && f.UserB == "b") || (f.UserA == "b" && f.UserB == "a") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("duplicate friendship %d", n)
	}
}

func TestBlockRejectsInvite(t *testing.T) {
	st := OpenSocialStore(filepath.Join(t.TempDir(), "s.json"))
	if msg := st.Block("a", "b"); msg != "" {
		t.Fatal(msg)
	}
	if _, msg := st.CreateInvite("b", "a", "rm", "ABC123"); msg == "" {
		t.Fatal("blocked invite")
	}
	if _, msg := st.RequestFriend("b", "a"); msg == "" {
		t.Fatal("blocked friend")
	}
}

func TestCasualMatchmakingFour(t *testing.T) {
	h := &Hub{Lobby: NewUlarLobby(), Matchmaker: NewMatchmaker()}
	ps := []*Player{testPlayer("u1", "A"), testPlayer("u2", "B"), testPlayer("u3", "C"), testPlayer("u4", "D")}
	for _, p := range ps {
		h.Lobby.Connect(p)
		if msg := h.queueJoin(p, "CASUAL", "ID-JKT", "", 2); msg != "" {
			t.Fatal(msg)
		}
	}
	h.Matchmaker.mu.Lock()
	for i := range h.Matchmaker.casual {
		h.Matchmaker.casual[i].EnqueuedAt = time.Now().Add(-25 * time.Second)
	}
	h.Matchmaker.mu.Unlock()
	h.tickMatchmaking()
	if h.Lobby.inRoom["u1"] == "" {
		t.Fatal("no room")
	}
	h.Matchmaker.mu.Lock()
	npend := len(h.Matchmaker.pending)
	h.Matchmaker.mu.Unlock()
	if npend != 1 {
		t.Fatalf("pending %d", npend)
	}
	for _, p := range ps {
		h.matchReady(p, true)
	}
}

func TestQueueCancel(t *testing.T) {
	h := &Hub{Lobby: NewUlarLobby(), Matchmaker: NewMatchmaker()}
	p := testPlayer("u1", "A")
	h.Lobby.Connect(p)
	_ = h.queueJoin(p, "CASUAL", "ID-JKT", "", 2)
	h.Matchmaker.Cancel("u1")
	if h.Matchmaker.Queued("u1") != "" {
		t.Fatal("still queued")
	}
}

func TestRankedSettlementAndLeaderboard(t *testing.T) {
	dir := t.TempDir()
	st := OpenProgressStore(filepath.Join(dir, "p.json"))
	st.Ensure("w", "Win")
	st.Ensure("l", "Lose")
	st.SetRankForTest("w", RankState{Tier: "GOLD", Division: "III", RR: 99})
	st.SetRankForTest("l", RankState{Tier: "GOLD", Division: "III", RR: 40})
	_, msg := st.SettleMatch(MatchSettlement{MatchID: "rm1", Mode: "RANKED", Players: []MatchSettlePlayer{
		{UserID: "w", Username: "Win", Rank: 1, Position: 100, Won: true, Reached100: true},
		{UserID: "l", Username: "Lose", Rank: 2, Position: 20, Won: false},
	}, RankWinRR: RankWinRR, RankLossRR: RankLossRR})
	if msg != "" {
		t.Fatal(msg)
	}
	w, _ := st.Get("w")
	if w.RankState().Division != "II" {
		t.Fatalf("promo %+v", w.RankState())
	}
	l, _ := st.Get("l")
	if l.RankRR < 0 || l.RankState().Index < 0 {
		t.Fatal("negative")
	}
	hist := st.RankHistory("w", 0)
	if len(hist) != 1 || hist[0].RRBefore < 0 {
		t.Fatalf("history %+v", hist)
	}
	_, msg = st.SettleMatch(MatchSettlement{MatchID: "rm1", Mode: "RANKED", Players: []MatchSettlePlayer{
		{UserID: "w", Username: "Win", Rank: 1, Position: 100, Won: true},
	}, RankWinRR: RankWinRR, RankLossRR: RankLossRR})
	if msg != "" {
		t.Fatal(msg)
	}
	if w2, _ := st.Get("w"); w2.RankedWins != 1 {
		t.Fatal("double ranked")
	}
}

func TestKickAndRoomFull(t *testing.T) {
	h := &Hub{Lobby: NewUlarLobby()}
	a, b, c, d, e := testPlayer("ua", "A"), testPlayer("ub", "B"), testPlayer("uc", "C"), testPlayer("ud", "D"), testPlayer("ue", "E")
	h.Lobby.Connect(a)
	room, _ := h.Lobby.CreateSized(a, 4)
	for _, p := range []*Player{b, c, d} {
		h.Lobby.Connect(p)
		if _, errc := h.Lobby.Join(p, room.RoomCode); errc != "" {
			t.Fatal(errc)
		}
	}
	h.Lobby.Connect(e)
	if _, errc := h.Lobby.Join(e, room.RoomCode); errc != ErrRoomFull {
		t.Fatalf("full %s", errc)
	}
	if _, errc := h.Lobby.Kick("ua", "ub"); errc != "" {
		t.Fatal(errc)
	}
	if h.Lobby.inRoom["ub"] != "" {
		t.Fatal("kicked still in")
	}
}

func TestAbandonRestriction(t *testing.T) {
	st := OpenProgressStore(filepath.Join(t.TempDir(), "p.json"))
	st.Ensure("u", "A")
	st.RecordAbandon("u")
	p, _ := st.Get("u")
	if p.AbandonCount != 1 || p.RestrictUntil < time.Now().UnixMilli() {
		t.Fatalf("%+v", p)
	}
}

func TestSeasonSoftResetKeepsXP(t *testing.T) {
	st := OpenProgressStore(filepath.Join(t.TempDir(), "p.json"))
	st.Ensure("u", "A")
	st.SetRankForTest("u", RankState{Tier: "GOLD", Division: "I", RR: 80})
	p, _ := st.Get("u")
	xp, coins := p.XP, p.Coins
	rs := SoftResetRank(p.RankState())
	st.SetRankForTest("u", rs)
	p2, _ := st.Get("u")
	if p2.XP != xp || p2.Coins != coins {
		t.Fatal("reset xp/coins")
	}
	if p2.RankState().Index >= p.RankState().Index && p.RankState().Index > 0 {
		t.Fatal("not soft")
	}
}
