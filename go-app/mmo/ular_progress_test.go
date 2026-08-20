package mmo

import (
	"path/filepath"
	"testing"
	"time"
)

func testProgress(t *testing.T) *ProgressStore {
	t.Helper()
	return OpenProgressStore(filepath.Join(t.TempDir(), "progress.json"))
}

func TestLevelFormulaProgressive(t *testing.T) {
	if XPForLevel(1) != 0 || XPForLevel(2) != 100 || XPForLevel(3) != 250 || XPForLevel(4) != 450 {
		t.Fatalf("formula %d %d %d %d", XPForLevel(1), XPForLevel(2), XPForLevel(3), XPForLevel(4))
	}
	if LevelFromXP(0) != 1 || LevelFromXP(99) != 1 || LevelFromXP(100) != 2 || LevelFromXP(250) != 3 {
		t.Fatal("level from xp")
	}
	lvl, into, need, next := XPToNext(850)
	if lvl != 5 || into != 150 || need != 150 || next != 1000 {
		t.Fatalf("bar %d %d %d %d", lvl, into, need, next)
	}
}

func TestNewProfileDefaults(t *testing.T) {
	st := testProgress(t)
	p := st.Ensure("u1", "Andi")
	if p.Level != 1 || p.XP != 0 || p.Coins != 0 || p.TotalMatches != 0 || p.Wins != 0 {
		t.Fatalf("%+v", p)
	}
}

func TestRecordAnswerXPFromConfig(t *testing.T) {
	st := testProgress(t)
	st.Ensure("u1", "Andi")
	st.RecordAnswer("u1", "Andi", SubjectMath, "a1", true, false, XP_CORRECT_ANSWER, XP_WRONG_ANSWER, XP_TIMEOUT, nil)
	p, _ := st.Get("u1")
	if p.XP != XP_CORRECT_ANSWER || p.CorrectAnswers != 1 {
		t.Fatalf("correct %+v", p)
	}
	st.RecordAnswer("u1", "Andi", SubjectMath, "a2", false, false, XP_CORRECT_ANSWER, XP_WRONG_ANSWER, XP_TIMEOUT, nil)
	p, _ = st.Get("u1")
	if p.XP != XP_CORRECT_ANSWER+XP_WRONG_ANSWER || p.WrongAnswers != 1 {
		t.Fatalf("wrong %+v", p)
	}
	st.RecordAnswer("u1", "Andi", SubjectMath, "a3", false, true, XP_CORRECT_ANSWER, XP_WRONG_ANSWER, XP_TIMEOUT, nil)
	p, _ = st.Get("u1")
	if p.XP != XP_CORRECT_ANSWER+XP_WRONG_ANSWER+XP_TIMEOUT || p.TimeoutAnswers != 1 {
		t.Fatalf("timeout %+v", p)
	}
}

func TestLevelUpFromXP(t *testing.T) {
	st := testProgress(t)
	st.Ensure("u1", "Andi")
	for i := 0; i < 10; i++ {
		st.RecordAnswer("u1", "Andi", SubjectMath, "q"+string(rune('a'+i)), true, false, XP_CORRECT_ANSWER, XP_WRONG_ANSWER, XP_TIMEOUT, nil)
	}
	p, _ := st.Get("u1")
	if p.XP != 100 || p.Level != 2 {
		t.Fatalf("want L2 100xp got L%d %d", p.Level, p.XP)
	}
}

func TestSettleMatchCoinsAndDoubleReward(t *testing.T) {
	st := testProgress(t)
	in := MatchSettlement{MatchID: "m1", Players: []MatchSettlePlayer{
		{UserID: "w", Username: "Win", Rank: 1, Position: 100, Won: true, Reached100: true},
		{UserID: "l", Username: "Lose", Rank: 2, Position: 40, Won: false},
	}}
	in.Mode = "CASUAL"
	in.XPMatchComplete = XP_MATCH_COMPLETE
	in.CoinMatch = COIN_MATCH
	in.XPWin = XP_WIN
	in.CoinWin = COIN_WIN
	in.RankWinRR = RankWinRR
	in.RankLossRR = RankLossRR
	ev, msg := st.SettleMatch(in)
	if msg != "" || len(ev) != 2 {
		t.Fatalf("settle %s %+v", msg, ev)
	}
	w, _ := st.Get("w")
	if w.Coins != COIN_MATCH+COIN_WIN+20+40 { // FIRST_GAME 20 + WIN_FIRST 40 + CENTURY 40
		// wait CENTURY 40 + FIRST 20 + WIN_FIRST 40 = 100 extra coins
	}
	wantWinCoins := COIN_MATCH + COIN_WIN
	for _, a := range AchievementCatalog {
		if a.ID == "FIRST_GAME" || a.ID == "WIN_FIRST" || a.ID == "CENTURY" {
			wantWinCoins += a.RewardCoins
		}
	}
	if w.Coins != wantWinCoins || w.Wins != 1 || w.Trophies != 1 || w.CurrentWinStreak != 1 || w.Level < 1 {
		t.Fatalf("winner %+v wantCoins=%d", w, wantWinCoins)
	}
	if len(ev) < 1 || ev[0].Trophies != 1 || !ev[0].Won {
		t.Fatalf("trophy reward %+v", ev)
	}
	if w.XP < XP_MATCH_COMPLETE+XP_WIN {
		t.Fatalf("xp %d", w.XP)
	}
	lose, _ := st.Get("l")
	if lose.Coins != COIN_MATCH+20 || lose.CurrentWinStreak != 0 || lose.Losses != 1 {
		t.Fatalf("loser %+v", lose)
	}
	ev2, msg := st.SettleMatch(in)
	if msg != "" || len(ev2) != 0 {
		t.Fatalf("double %s %+v", msg, ev2)
	}
	w2, _ := st.Get("w")
	if w2.Coins != w.Coins || w2.XP != w.XP {
		t.Fatal("double reward applied")
	}
	n := 0
	for _, u := range st.Unlocks("w") {
		if u.AchievementID == "FIRST_GAME" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("duplicate first game %d", n)
	}
}

func TestDailyClaimOnce(t *testing.T) {
	st := testProgress(t)
	st.Ensure("u1", "Andi")
	now := time.Date(2026, 8, 20, 10, 0, 0, 0, time.FixedZone("WIB", 7*3600))
	claim, ev, msg := st.ClaimDaily("u1", "Andi", now)
	if msg != "" || claim.Day != 1 || ev.Coins != DailyCoinTable[0] || DailyCoinTable[0] != COIN_DAILY {
		t.Fatalf("day1 %s %+v %+v", msg, claim, ev)
	}
	_, _, msg = st.ClaimDaily("u1", "Andi", now.Add(time.Hour))
	if msg == "" {
		t.Fatal("second claim")
	}
	st2, ev2, msg := st.ClaimDaily("u1", "Andi", now.Add(24*time.Hour))
	if msg != "" || st2.Day != 2 || ev2.Coins != 75 {
		t.Fatalf("day2 %s %+v", msg, st2)
	}
	_, _, msg = st.ClaimDaily("u1", "Andi", now.Add(72*time.Hour))
	if msg != "" {
		t.Fatal("gap should reset not reject")
	}
	p, _ := st.Get("u1")
	if p.DailyStreak != 1 {
		t.Fatalf("streak reset got %d", p.DailyStreak)
	}
}

func TestWinStreakResetOnLoss(t *testing.T) {
	st := testProgress(t)
	win := func(id string) {
		_, msg := st.SettleMatch(MatchSettlement{MatchID: id, Players: []MatchSettlePlayer{
			{UserID: "u", Username: "A", Rank: 1, Position: 100, Won: true},
		}})
		if msg != "" {
			t.Fatal(msg)
		}
	}
	win("m1")
	win("m2")
	p, _ := st.Get("u")
	if p.CurrentWinStreak != 2 || p.BestWinStreak != 2 {
		t.Fatalf("streak %+v", p)
	}
	_, msg := st.SettleMatch(MatchSettlement{MatchID: "m3", Players: []MatchSettlePlayer{
		{UserID: "u", Username: "A", Rank: 2, Position: 10, Won: false},
	}})
	if msg != "" {
		t.Fatal(msg)
	}
	p, _ = st.Get("u")
	if p.CurrentWinStreak != 0 || p.BestWinStreak != 2 {
		t.Fatalf("reset %+v", p)
	}
}

func TestAvatarAndTitleValidation(t *testing.T) {
	st := testProgress(t)
	st.Ensure("u1", "Andi")
	if _, msg := st.UpdateProfile("u1", "", "http://evil", ""); msg == "" {
		t.Fatal("url avatar")
	}
	if _, msg := st.UpdateProfile("u1", "", "bird", "master"); msg == "" {
		t.Fatal("locked title")
	}
	if _, msg := st.UpdateProfile("u1", "", "bird", "pemula"); msg != "" {
		t.Fatal(msg)
	}
}

func TestNegativeCoinsRejected(t *testing.T) {
	st := testProgress(t)
	st.Ensure("u1", "Andi")
	if _, msg := st.AdminAdjustCoins("u1", "admin", "test", -10); msg == "" {
		t.Fatal("negative")
	}
	if NewCurrencyService(st).Balance("u1") != 0 {
		t.Fatal("balance")
	}
}

func TestAtomicSettleRollback(t *testing.T) {
	st := testProgress(t)
	st.Ensure("u1", "Andi")
	st.SnapshotFailNext()
	_, msg := st.SettleMatch(MatchSettlement{MatchID: "m1", Players: []MatchSettlePlayer{
		{UserID: "u1", Username: "Andi", Rank: 1, Position: 100, Won: true},
	}})
	if msg == "" {
		t.Fatal("want fail")
	}
	p, _ := st.Get("u1")
	if p.Coins != 0 || p.XP != 0 || p.TotalMatches != 0 || len(st.Unlocks("u1")) != 0 || len(st.History("u1", 0)) != 0 {
		t.Fatalf("partial %+v", p)
	}
}

func TestReconnectSameProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.json")
	st := OpenProgressStore(path)
	st.Ensure("u1", "Andi")
	st.RecordAnswer("u1", "Andi", SubjectPAI, "a", true, false, XP_CORRECT_ANSWER, XP_WRONG_ANSWER, XP_TIMEOUT, nil)
	st2 := OpenProgressStore(path)
	p, ok := st2.Get("u1")
	if !ok || p.XP != XP_CORRECT_ANSWER || p.Username != "Andi" {
		t.Fatalf("reload %+v %v", p, ok)
	}
}

func TestWinRateAndAccuracyZero(t *testing.T) {
	st := testProgress(t)
	view := st.ViewFor("u1", "Baru")
	if view["winRate"] != 0.0 || view["accuracy"] != 0.0 || view["level"] != 1 {
		t.Fatalf("%+v", view)
	}
}

func TestSpendCoinsLedger(t *testing.T) {
	st := testProgress(t)
	st.Ensure("u1", "Andi")
	_, msg := st.AdminAdjustCoins("u1", "admin", "seed", 100)
	if msg != "" {
		t.Fatal(msg)
	}
	tx, msg := NewCurrencyService(st).Spend("u1", 40, "cosmetic")
	if msg != "" || tx.Amount != -40 || tx.BalanceAfter != 60 {
		t.Fatalf("%s %+v", msg, tx)
	}
}
