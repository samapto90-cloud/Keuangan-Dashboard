package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func endgamePlayer(t *testing.T) (*WorldState, *Player) {
	t.Helper()
	w := NewWorldState()
	p := addTestPlayer(w, "p_eg", "Raka")
	p.Level = 40
	p.ensureLog().Flags["storyCompleted"] = true
	p.ensureLog().DailyDay = utcDayKey()
	p.ensureLog().WeeklyWeek = raidWeekKey()
	return w, p
}

func TestDailyRewardOnce(t *testing.T) {
	w, p := endgamePlayer(t)
	log := p.ensureLog()
	log.DailyProgress["dq-dungeon"] = 1
	evs := w.ApplyEndgame(p.ID, Envelope{Type: TypeClaimDaily, Data: []byte(`{"id":"dq-dungeon","transactionId":"tx-daily-1"}`)})
	if rejectAction(evs, TypeClaimDaily) {
		t.Fatalf("claim daily: %s", dumpFirst(evs))
	}
	again := w.ApplyEndgame(p.ID, Envelope{Type: TypeClaimDaily, Data: []byte(`{"id":"dq-dungeon","transactionId":"tx-daily-1"}`)})
	if !hasMsgType(again, TypeEndgameState) && !rejectAction(again, TypeClaimDaily) {
		t.Fatal("duplikat tx harus replay atau ditolak")
	}
	dup := w.ApplyEndgame(p.ID, Envelope{Type: TypeClaimDaily, Data: []byte(`{"id":"dq-dungeon","transactionId":"tx-daily-2"}`)})
	if !rejectAction(dup, TypeClaimDaily) {
		t.Fatal("reward daily hanya sekali")
	}
}

func TestSeasonEndSavesHistory(t *testing.T) {
	w, p := endgamePlayer(t)
	log := p.ensureLog()
	log.SeasonLevel = 50
	log.SeasonXP = 5000
	w.mu.Lock()
	w.forceSeasonEnd(p, time.Now().UTC())
	w.mu.Unlock()
	if len(p.ensureLog().SeasonHistory) < 1 {
		t.Fatal("history season harus tersimpan")
	}
	if p.ensureLog().SeasonHistory[0].Level != 50 {
		t.Fatalf("final level %d", p.ensureLog().SeasonHistory[0].Level)
	}
	if p.ensureLog().SeasonLevel != 1 || p.ensureLog().SeasonXP != 0 {
		t.Fatalf("season baru level=%d xp=%d", p.ensureLog().SeasonLevel, p.ensureLog().SeasonXP)
	}
}

func TestAchievementProgressOnce(t *testing.T) {
	w, p := endgamePlayer(t)
	p.ensureLog().EduCorrect = 99
	p.ensureLog().EduAnswered = 99
	w.mu.Lock()
	first := w.recordEducation(p, "Matematika", true)
	second := w.recordEducation(p, "Matematika", true)
	w.mu.Unlock()
	if p.ensureLog().EduCorrect < 100 {
		t.Fatal("progress 100")
	}
	n := 0
	for _, a := range p.ensureLog().Achievements {
		if a == "little-scholar" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("achievement sekali, got %d first=%v second=%v", n, hasMsgType(first, TypeAchievementUnlocked), hasMsgType(second, TypeAchievementUnlocked))
	}
}

func TestCosmeticEquipRejected(t *testing.T) {
	w, p := endgamePlayer(t)
	evs := w.ApplyEconomy(p.ID, Envelope{Type: TypeSetCosmetic, Data: []byte(`{"id":"aura-festival-light"}`)})
	if !rejectAction(evs, TypeSetCosmetic) {
		t.Fatal("equip tanpa owned harus REJECT")
	}
}

func TestCommunityEventRewardOnce(t *testing.T) {
	w, p := endgamePlayer(t)
	w.mu.Lock()
	c := w.ensureCommunity()
	c.Target = 15
	c.DailyCap = 100
	c.State = "ACTIVE"
	c.Points = 0
	evs := w.addCommunityPoints(p, 20)
	_ = evs
	if c.Points < c.Target {
		t.Fatalf("points %d target %d state %s cap %d amt %d", c.Points, c.Target, c.State, c.DailyCap, p.ensureLog().LiveDayAmt)
	}
	w.maybeGrantFestival(p)
	owned := 0
	for _, id := range p.ensureLog().Cosmetics {
		if id == "aura-festival-light" {
			owned++
		}
	}
	w.maybeGrantFestival(p)
	w.mu.Unlock()
	again := 0
	for _, id := range p.ensureLog().Cosmetics {
		if id == "aura-festival-light" {
			again++
		}
	}
	if owned != 1 || again != 1 {
		t.Fatalf("festival cosmetic sekali owned=%d again=%d", owned, again)
	}
}

func TestHorizonScoreLeaderboard(t *testing.T) {
	w, p := endgamePlayer(t)
	evs := w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"dun-horizon"}`)})
	if rejectAction(evs, TypeDungeonEnter) {
		t.Fatalf("horizon enter: %s", dumpFirst(evs))
	}
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		t.Fatal("horizon instance")
	}
	if len(inst.Modifiers) < 1 {
		t.Fatal("server harus memilih modifier")
	}
	if !validHorizonCombo(inst.Modifiers) {
		t.Fatalf("kombinasi tidak valid %v", inst.Modifiers)
	}
	w.mu.Lock()
	w.completeDungeon(inst, p)
	w.mu.Unlock()
	if p.ensureLog().HorizonBest < 1 {
		t.Fatal("score harus tercatat")
	}
	if len(w.HorizonLB) < 1 {
		t.Fatal("leaderboard weekly")
	}
}

func TestDailyWeeklyResetServerClock(t *testing.T) {
	w, p := endgamePlayer(t)
	log := p.ensureLog()
	log.DailyProgress["dq-dungeon"] = 1
	log.DailyClaimed["dq-dungeon"] = true
	log.DailyDay = "1999-01-01"
	log.WeeklyProgress["wq-raid"] = 1
	log.WeeklyClaimed["wq-raid"] = true
	log.WeeklyWeek = "1999-W01"
	w.mu.Lock()
	w.ensureEndgameClock(p)
	w.mu.Unlock()
	if len(p.ensureLog().DailyProgress) != 0 || p.ensureLog().DailyClaimed["dq-dungeon"] {
		t.Fatal("daily harus refresh")
	}
	if len(p.ensureLog().WeeklyProgress) != 0 || p.ensureLog().WeeklyClaimed["wq-raid"] {
		t.Fatal("weekly harus refresh")
	}
}

func TestGuildChallengeOnce(t *testing.T) {
	w := NewWorldState()
	a := addTestPlayer(w, "p_ga", "Raka")
	b := addTestPlayer(w, "p_gb", "Lila")
	a.Level, b.Level = 20, 20
	a.ensureLog().Coin = 5000
	evs := w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Horizon Hands","tag":"HORZ"}`)})
	if rejectAction(evs, TypeGuildCreate) {
		t.Fatalf("guild: %s", dumpFirst(evs))
	}
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildInvite, Data: []byte(`{"targetId":"p_gb"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeGuildAccept, Data: []byte(`{}`)})
	g := w.guildOf(a.ID)
	if g == nil {
		t.Fatal("guild missing")
	}
	exp := g.Exp
	inst := &DungeonInstance{DefID: "dun-ch01", Players: []string{a.ID, b.ID}}
	w.mu.Lock()
	for i := 0; i < 20; i++ {
		w.noteGuildDungeon(inst)
	}
	w.noteGuildDungeon(inst)
	w.mu.Unlock()
	g = w.guildOf(a.ID)
	if g.WeeklyDungeon < 20 || !g.WeeklyRewarded {
		t.Fatalf("weekly dungeon %d rewarded %v", g.WeeklyDungeon, g.WeeklyRewarded)
	}
	if g.Exp <= exp {
		t.Fatal("guild XP sekali saat syarat terpenuhi")
	}
}

func TestEducationAdd52(t *testing.T) {
	w, p := testVillagePlayer()
	q := p.quest("eq001")
	if q == nil {
		t.Fatal("eq001")
	}
	q.State = QuestActive
	idx := -1
	for i, def := range questionCatalog {
		if def.ID == "q-add52" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("q-add52")
	}
	log := p.ensureLog()
	log.Quiz.Active = true
	log.Quiz.QuestID = "eq001"
	log.Quiz.Index = idx
	before := log.EduCorrect
	raw, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-add52", Choice: 1})
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: raw})
	if rejectAction(evs, TypeEducationAnswer) {
		t.Fatalf("answer: %s", dumpFirst(evs))
	}
	if p.ensureLog().EduCorrect != before+1 {
		t.Fatalf("progress %d", p.ensureLog().EduCorrect)
	}
}

func TestEducationNoSessionRejected(t *testing.T) {
	w, p := testVillagePlayer()
	raw, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-add52", Choice: 1})
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: raw})
	if !rejectAction(evs, TypeEducationAnswer) {
		t.Fatal("tanpa session harus REJECT")
	}
}

func TestSeasonXPCheatRejected(t *testing.T) {
	w, p := endgamePlayer(t)
	before := p.ensureLog().SeasonXP
	evs := w.ApplyEndgame(p.ID, Envelope{Type: TypeSetSeasonXP, Data: []byte(`{"seasonXP":999999}`)})
	if !rejectAction(evs, TypeSetSeasonXP) {
		t.Fatal("SET_SEASON_XP harus REJECT")
	}
	if p.ensureLog().SeasonXP != before {
		t.Fatal("season XP client tidak boleh masuk")
	}
}

func TestAchievementCheatRejected(t *testing.T) {
	w, p := endgamePlayer(t)
	if !rejectAction(w.ApplyEndgame(p.ID, Envelope{Type: TypeUnlockAchievement, Data: []byte(`{"id":"little-scholar"}`)}), TypeUnlockAchievement) {
		t.Fatal("UNLOCK_ACHIEVEMENT")
	}
	if !rejectAction(w.ApplyEndgame(p.ID, Envelope{Type: TypeSetAchievement, Data: []byte(`{"achievementCompleted":true}`)}), TypeSetAchievement) {
		t.Fatal("SET_ACHIEVEMENT")
	}
}

func TestCosmeticUnlockCheatRejected(t *testing.T) {
	w, p := endgamePlayer(t)
	evs := w.ApplyEndgame(p.ID, Envelope{Type: TypeUnlockCosmetic, Data: []byte(`{"id":"aura-horizon-free"}`)})
	if !rejectAction(evs, TypeUnlockCosmetic) {
		t.Fatal("UNLOCK_COSMETIC harus REJECT")
	}
}

func TestEndgameTickStable(t *testing.T) {
	w := NewWorldState()
	for i := 0; i < 40; i++ {
		p := addTestPlayer(w, "p_load_"+string(rune('A'+i%26))+string(rune('0'+i/26)), "P")
		p.X, p.Z = float64(i%8), float64(6+i%4)
	}
	start := time.Now()
	for i := 0; i < 8; i++ {
		w.Simulate(0.05)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("tick tidak stabil")
	}
}

func hasMsgType(evs [][]byte, kind string) bool {
	for _, ev := range evs {
		var env Envelope
		if json.Unmarshal(ev, &env) == nil && env.Type == kind {
			return true
		}
	}
	return false
}

func dumpFirst(evs [][]byte) string {
	if len(evs) == 0 {
		return "empty"
	}
	return string(evs[0])
}
