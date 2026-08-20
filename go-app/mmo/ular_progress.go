package mmo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type progressBlob struct {
	Profiles     map[string]PlayerProfile `json:"profiles"`
	XP           []XPTransaction          `json:"xp"`
	Coins        []CoinTransaction        `json:"coins"`
	Unlocks      []UserAchievement        `json:"unlocks"`
	Daily        []DailyClaim             `json:"daily"`
	MatchResults []MatchPlayerResult      `json:"matchResults"`
	RankChanges  []RankChange             `json:"rankChanges"`
}

type ProgressStore struct {
	mu       sync.Mutex
	path     string
	failNext bool
	blob     progressBlob
}

func progressStorePath() string {
	if p := strings.TrimSpace(os.Getenv("ULAR_PROGRESS_STORE")); p != "" {
		return p
	}
	return filepath.Join("data", "ular-progress.json")
}

func OpenProgressStore(path string) *ProgressStore {
	s := &ProgressStore{path: path, blob: progressBlob{Profiles: map[string]PlayerProfile{}}}
	raw, err := os.ReadFile(path)
	if err == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &s.blob)
	}
	if s.blob.Profiles == nil {
		s.blob.Profiles = map[string]PlayerProfile{}
	}
	return s
}

func (s *ProgressStore) flushLocked() error {
	if s.failNext {
		s.failNext = false
		return os.ErrInvalid
	}
	raw, err := json.MarshalIndent(s.blob, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
	return os.WriteFile(s.path, raw, 0o644)
}

func (s *ProgressStore) SnapshotFailNext() {
	s.mu.Lock()
	s.failNext = true
	s.mu.Unlock()
}

func (s *ProgressStore) Ensure(userID, username string) PlayerProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, _ := s.ensureLocked(userID, username)
	_ = s.flushLocked()
	return p
}

func (s *ProgressStore) ensureLocked(userID, username string) (PlayerProfile, bool) {
	if p, ok := s.blob.Profiles[userID]; ok {
		changed := false
		if username != "" && p.Username != username {
			p.Username = username
			changed = true
		}
		if p.Trophies < p.Wins {
			p.Trophies = p.Wins
			changed = true
		}
		if changed {
			p.UpdatedAt = time.Now().UnixMilli()
			s.blob.Profiles[userID] = p
		}
		return p, false
	}
	now := time.Now().UnixMilli()
	p := PlayerProfile{
		ID: "pf-" + shortID(), UserID: userID, Username: username,
		Avatar: "bird", Title: "pemula", Level: 1, XP: 0, Coins: 0, Trophies: 0,
		RankTier: "BRONZE", RankDivision: "III", SeasonID: CurrentSeason.ID,
		AllowFriends: true, AllowInvites: true, ShowOnline: true,
		SubjectCorrect: map[string]int{}, SubjectTotal: map[string]int{},
		CreatedAt: now, UpdatedAt: now,
	}
	s.blob.Profiles[userID] = p
	return p, true
}

func (s *ProgressStore) Get(userID string) (PlayerProfile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.blob.Profiles[userID]
	return p, ok
}

func (s *ProgressStore) grantXPLocked(p *PlayerProfile, typ, ref string, amount int) XPTransaction {
	if amount < 0 {
		amount = 0
	}
	before := p.XP
	p.XP += amount
	after := p.XP
	p.Level = LevelFromXP(p.XP)
	if p.Level > MaxPlayerLevel {
		p.Level = MaxPlayerLevel
	}
	p.UpdatedAt = time.Now().UnixMilli()
	tx := XPTransaction{
		ID: "xp-" + shortID(), UserID: p.UserID, Type: typ, Amount: amount,
		BalanceBefore: before, BalanceAfter: after, ReferenceID: ref, CreatedAt: time.Now().UnixMilli(),
	}
	s.blob.XP = append(s.blob.XP, tx)
	return tx
}

func (s *ProgressStore) grantCoinsLocked(p *PlayerProfile, typ, ref, admin, reason string, amount int) (CoinTransaction, bool) {
	before := p.Coins
	after := before + amount
	if after < 0 {
		return CoinTransaction{}, false
	}
	p.Coins = after
	p.UpdatedAt = time.Now().UnixMilli()
	tx := CoinTransaction{
		ID: "cn-" + shortID(), UserID: p.UserID, Type: typ, Amount: amount,
		BalanceBefore: before, BalanceAfter: after, ReferenceID: ref,
		AdminID: admin, Reason: reason, CreatedAt: time.Now().UnixMilli(),
	}
	s.blob.Coins = append(s.blob.Coins, tx)
	return tx, true
}

func (s *ProgressStore) SpendCoins(userID string, amount int, ref string) (CoinTransaction, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	backup := cloneBlob(s.blob)
	p, ok := s.blob.Profiles[userID]
	if !ok {
		return CoinTransaction{}, "profil tidak ada"
	}
	if amount <= 0 {
		return CoinTransaction{}, "jumlah tidak valid"
	}
	tx, ok := s.grantCoinsLocked(&p, "SPEND", ref, "", "", -amount)
	if !ok {
		return CoinTransaction{}, "koin tidak cukup"
	}
	s.blob.Profiles[userID] = p
	if err := s.flushLocked(); err != nil {
		s.blob = backup
		return CoinTransaction{}, "gagal menyimpan"
	}
	return tx, ""
}

func (s *ProgressStore) hasUnlockLocked(userID, achID string) bool {
	for _, u := range s.blob.Unlocks {
		if u.UserID == userID && u.AchievementID == achID {
			return true
		}
	}
	return false
}

func (s *ProgressStore) hasMatchLocked(matchID, userID string) bool {
	for _, m := range s.blob.MatchResults {
		if m.MatchID == matchID && m.UserID == userID {
			return true
		}
	}
	return false
}

func (s *ProgressStore) hasDailyLocked(userID, date string) bool {
	for _, d := range s.blob.Daily {
		if d.UserID == userID && d.RewardDate == date {
			return true
		}
	}
	return false
}

func (s *ProgressStore) unlockLocked(p *PlayerProfile, ach UlarAchievement, ref string) RewardEvent {
	ev := RewardEvent{UserID: p.UserID, LevelBefore: p.Level}
	if !ach.Active || s.hasUnlockLocked(p.UserID, ach.ID) {
		ev.LevelAfter = p.Level
		return ev
	}
	s.blob.Unlocks = append(s.blob.Unlocks, UserAchievement{UserID: p.UserID, AchievementID: ach.ID, UnlockedAt: time.Now().UnixMilli()})
	if ach.RewardXP > 0 {
		s.grantXPLocked(p, XPAchievement, ref+":"+ach.ID, ach.RewardXP)
		ev.XP += ach.RewardXP
	}
	if ach.RewardCoins > 0 {
		_, _ = s.grantCoinsLocked(p, CoinAchievement, ref+":"+ach.ID, "", "", ach.RewardCoins)
		ev.Coins += ach.RewardCoins
	}
	ev.LevelAfter = p.Level
	ev.LevelUp = ev.LevelAfter > ev.LevelBefore
	ev.Achievements = []string{ach.ID}
	return ev
}

func (s *ProgressStore) RecordAnswer(
	userID, username, subject, attemptID string,
	correct, timeout bool,
	xpCorrect, xpWrong, xpTimeout int,
	achByID map[string]UlarAchievement,
) RewardEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	backup := cloneBlob(s.blob)
	p, _ := s.ensureLocked(userID, username)
	before := p.Level
	p.TotalQuestions++
	if p.SubjectTotal == nil {
		p.SubjectTotal = map[string]int{}
		p.SubjectCorrect = map[string]int{}
	}
	if subject != "" {
		p.SubjectTotal[subject]++
	}
	xpType := XPWrong
	amt := xpWrong
	if timeout {
		p.TimeoutAnswers++
		xpType = XPTimeout
		amt = xpTimeout
	} else if correct {
		p.CorrectAnswers++
		if subject != "" {
			p.SubjectCorrect[subject]++
		}
		xpType = XPCorrect
		amt = xpCorrect
	} else {
		p.WrongAnswers++
		amt = xpWrong
	}
	s.grantXPLocked(&p, xpType, attemptID, amt)
	ev := RewardEvent{UserID: userID, XP: amt, LevelBefore: before, LevelAfter: p.Level, LevelUp: p.Level > before}
	ev.Achievements = s.checkQuizLocked(&p, attemptID, achByID)
	s.blob.Profiles[userID] = p
	if err := s.flushLocked(); err != nil {
		s.blob = backup
		return RewardEvent{}
	}
	return ev
}

func (s *ProgressStore) checkQuizLocked(p *PlayerProfile, ref string, achByID map[string]UlarAchievement) []string {
	var out []string
	try := func(id string, ok bool) {
		if !ok || s.hasUnlockLocked(p.UserID, id) {
			return
		}
		var def UlarAchievement
		var found bool
		if achByID != nil {
			def, found = achByID[id]
		} else {
			def, found = achievementByID(id)
		}
		if !found {
			return
		}
		ev := s.unlockLocked(p, def, ref)
		out = append(out, ev.Achievements...)
	}
	try("QUIZ_MASTER", p.CorrectAnswers >= 50)
	try("MATH_MASTER", p.SubjectCorrect[SubjectMath] >= 25)
	try("PAI_MASTER", p.SubjectCorrect[SubjectPAI] >= 25)
	try("ENGLISH_MASTER", p.SubjectCorrect[SubjectEnglish] >= 25)
	try("JAVA_MASTER", p.SubjectCorrect[SubjectJawa] >= 25)
	return out
}

func (s *ProgressStore) SettleMatch(in MatchSettlement) ([]RewardEvent, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	backup := cloneBlob(s.blob)
	events := make([]RewardEvent, 0, len(in.Players))
	now := time.Now().UnixMilli()
	for _, pl := range in.Players {
		if s.hasMatchLocked(in.MatchID, pl.UserID) {
			continue
		}
		p, _ := s.ensureLocked(pl.UserID, pl.Username)
		beforeLvl := p.Level
		xp := 0
		coins := 0
		s.grantXPLocked(&p, XPMatchDone, in.MatchID, in.XPMatchComplete)
		xp += in.XPMatchComplete
		_, _ = s.grantCoinsLocked(&p, CoinMatch, in.MatchID, "", "", in.CoinMatch)
		coins += in.CoinMatch
		p.TotalMatches++
		p.LadderClimbs += pl.Ladders
		p.SnakeHits += pl.Snakes
		if pl.Reached100 {
			p.Reached100++
		}
		if pl.Won {
			p.Wins++
			p.Trophies++
			p.CurrentWinStreak++
			if p.CurrentWinStreak > p.BestWinStreak {
				p.BestWinStreak = p.CurrentWinStreak
			}
			s.grantXPLocked(&p, XPWin, in.MatchID, in.XPWin)
			xp += in.XPWin
			_, _ = s.grantCoinsLocked(&p, CoinWin, in.MatchID, "", "", in.CoinWin)
			coins += in.CoinWin
		} else {
			p.Losses++
			p.CurrentWinStreak = 0
		}
		achs := []string{}
		unlock := func(id string, cond bool) {
			if !cond {
				return
			}
			var def UlarAchievement
			var ok bool
			if in.AchievementsByID != nil {
				def, ok = in.AchievementsByID[id]
			} else {
				def, ok = achievementByID(id)
			}
			if !ok {
				return
			}
			ev := s.unlockLocked(&p, def, in.MatchID)
			achs = append(achs, ev.Achievements...)
			xp += ev.XP
			coins += ev.Coins
		}
		unlock("FIRST_GAME", p.TotalMatches >= 1)
		unlock("WIN_FIRST", p.Wins >= 1)
		unlock("CLIMBER", p.LadderClimbs >= 10)
		unlock("SNAKE_SURVIVOR", p.SnakeHits >= 10)
		unlock("WIN_STREAK_3", p.CurrentWinStreak >= 3)
		unlock("WIN_STREAK_5", p.CurrentWinStreak >= 5)
		unlock("CENTURY", p.Reached100 >= 1)
		achs = append(achs, s.checkQuizLocked(&p, in.MatchID, in.AchievementsByID)...)
		res := MatchPlayerResult{
			MatchID: in.MatchID, UserID: pl.UserID, Rank: pl.Rank, Position: pl.Position,
			Correct: pl.Correct, Wrong: pl.Wrong, Timeout: pl.Timeout,
			XPEarned: xp, CoinsEarned: coins, Won: pl.Won, Mode: in.Mode, CreatedAt: now,
		}
		if in.Mode == "RANKED" {
			before := p.RankState()
			after := ApplyRankRR(before, pl.Won, in.RankWinRR, in.RankLossRR)
			if pl.Won {
				p.RankedWins++
			}
			res.RRBefore = before.Index
			res.RRChange = after.Index - before.Index
			res.RRAfter = after.Index
			applyRankToProfile(&p, after)
			s.blob.RankChanges = append(s.blob.RankChanges, RankChange{
				ID: "rk-" + shortID(), UserID: p.UserID, MatchID: in.MatchID, SeasonID: after.SeasonID,
				Result:     map[bool]string{true: "WIN", false: "LOSS"}[pl.Won],
				TierBefore: before.Tier, DivBefore: before.Division, RRBefore: before.RR,
				RRChange: res.RRChange, RRAfter: after.RR, TierAfter: after.Tier, DivAfter: after.Division,
				CreatedAt: now,
			})
			if after.Label() != before.Label() {
				kind := NoteRankUp
				if after.Index < before.Index {
					kind = NoteRankDown
				}
				_ = kind
			}
		}
		s.blob.MatchResults = append(s.blob.MatchResults, res)
		p.UpdatedAt = now
		s.blob.Profiles[pl.UserID] = p
		trophyGain := 0
		if pl.Won {
			trophyGain = 1
		}
		ev := RewardEvent{
			UserID: pl.UserID, XP: xp, Coins: coins,
			Trophies: trophyGain, TrophyTotal: p.Trophies, Won: pl.Won,
			LevelBefore: beforeLvl, LevelAfter: p.Level, LevelUp: p.Level > beforeLvl,
			Achievements: achs,
		}
		if in.Mode == "RANKED" {
			ev.RankLabel = p.RankState().Label()
			ev.RR = p.RankRR
			ev.RRDelta = res.RRChange
			ev.RankUp = res.RRChange > 0 && p.RankState().Label() != ""
			ev.RankDown = res.RRChange < 0
		}
		events = append(events, ev)
	}
	if err := s.flushLocked(); err != nil {
		s.blob = backup
		return nil, "gagal menyimpan reward"
	}
	InvalidateLeaderboard()
	return events, ""
}

type MatchSettlement struct {
	MatchID string
	Mode    string
	Players []MatchSettlePlayer
	// Frozen reward config for match integrity.
	XPMatchComplete int
	CoinMatch       int
	XPWin           int
	CoinWin         int
	// Frozen RR values for ranked history integrity.
	RankWinRR  int
	RankLossRR int
	// Frozen achievement overlay for match integrity.
	AchievementsByID map[string]UlarAchievement
}

type MatchSettlePlayer struct {
	UserID     string
	Username   string
	Rank       int
	Position   int
	Correct    int
	Wrong      int
	Timeout    int
	Won        bool
	Ladders    int
	Snakes     int
	Reached100 bool
}

func cloneBlob(b progressBlob) progressBlob {
	raw, _ := json.Marshal(b)
	var out progressBlob
	_ = json.Unmarshal(raw, &out)
	if out.Profiles == nil {
		out.Profiles = map[string]PlayerProfile{}
	}
	return out
}

func (s *ProgressStore) ClaimDaily(userID, username string, now time.Time) (DailyClaim, RewardEvent, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	backup := cloneBlob(s.blob)
	p, _ := s.ensureLocked(userID, username)
	date := JakartaDate(now)
	if s.hasDailyLocked(userID, date) {
		return DailyClaim{}, RewardEvent{}, "sudah diklaim hari ini"
	}
	streak := 1
	if p.LastDailyDate == YesterdayJakarta(now) {
		streak = p.DailyStreak + 1
	}
	if streak > 7 {
		streak = ((streak - 1) % 7) + 1
	}
	if streak < 1 {
		streak = 1
	}
	coins := LiveDailyCoins()[streak-1]
	p.DailyStreak = streak
	p.LastDailyDate = date
	before := p.Level
	_, ok := s.grantCoinsLocked(&p, CoinDaily, date, "", "", coins)
	if !ok {
		s.blob = backup
		return DailyClaim{}, RewardEvent{}, "gagal memberikan koin"
	}
	xpAmt := 0
	if xs := LiveConfig().DailyXP; LiveConfig().Version > 0 && len(xs) == 7 {
		xpAmt = xs[streak-1]
		if xpAmt < 0 {
			xpAmt = 0
		}
		if xpAmt > 0 {
			s.grantXPLocked(&p, CoinDaily, date, xpAmt)
		}
	}
	claim := DailyClaim{UserID: userID, RewardDate: date, Day: streak, Coins: coins, CreatedAt: time.Now().UnixMilli()}
	s.blob.Daily = append(s.blob.Daily, claim)
	s.blob.Profiles[userID] = p
	if err := s.flushLocked(); err != nil {
		s.blob = backup
		return DailyClaim{}, RewardEvent{}, "gagal menyimpan"
	}
	return claim, RewardEvent{UserID: userID, Coins: coins, XP: xpAmt, LevelBefore: before, LevelAfter: p.Level}, ""
}

func (s *ProgressStore) DailyStatus(userID string, now time.Time) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.blob.Profiles[userID]
	date := JakartaDate(now)
	claimed := s.hasDailyLocked(userID, date)
	next := 1
	if p.LastDailyDate == YesterdayJakarta(now) {
		next = p.DailyStreak + 1
		if next > 7 {
			next = 1
		}
	} else if claimed {
		next = p.DailyStreak
	}
	if next < 1 {
		next = 1
	}
	return map[string]any{
		"claimed": claimed, "day": next, "coins": LiveDailyCoins()[next-1],
		"streak": p.DailyStreak, "date": date,
	}
}

func (s *ProgressStore) History(userID string, page int) []MatchPlayerResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	if page < 0 {
		page = 0
	}
	own := make([]MatchPlayerResult, 0)
	for i := len(s.blob.MatchResults) - 1; i >= 0; i-- {
		if s.blob.MatchResults[i].UserID == userID {
			own = append(own, s.blob.MatchResults[i])
		}
	}
	start := page * MaxMatchHistoryPage
	if start >= len(own) {
		return []MatchPlayerResult{}
	}
	end := start + MaxMatchHistoryPage
	if end > len(own) {
		end = len(own)
	}
	return own[start:end]
}

func (s *ProgressStore) Unlocks(userID string) []UserAchievement {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]UserAchievement, 0)
	for _, u := range s.blob.Unlocks {
		if u.UserID == userID {
			out = append(out, u)
		}
	}
	return out
}

func (s *ProgressStore) UpdateProfile(userID, username, avatar, title string) (PlayerProfile, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	backup := cloneBlob(s.blob)
	p, ok := s.blob.Profiles[userID]
	if !ok {
		return PlayerProfile{}, "profil tidak ada"
	}
	if avatar != "" {
		if !validAvatar(avatar) {
			return PlayerProfile{}, "avatar tidak valid"
		}
		p.Avatar = avatar
	}
	if title != "" {
		if !validTitle(title) {
			return PlayerProfile{}, "gelar tidak valid"
		}
		req := ""
		for _, t := range TitleCatalog {
			if t.ID == title {
				req = t.Requires
			}
		}
		if req != "" && !s.hasUnlockLocked(userID, req) {
			return PlayerProfile{}, "gelar belum terbuka"
		}
		p.Title = title
	}
	if username != "" {
		p.Username = username
	}
	p.UpdatedAt = time.Now().UnixMilli()
	s.blob.Profiles[userID] = p
	if err := s.flushLocked(); err != nil {
		s.blob = backup
		return PlayerProfile{}, "gagal menyimpan"
	}
	return p, ""
}

func (s *ProgressStore) AdminAdjustCoins(userID, adminID, reason string, amount int) (CoinTransaction, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	backup := cloneBlob(s.blob)
	p, ok := s.blob.Profiles[userID]
	if !ok {
		return CoinTransaction{}, "profil tidak ada"
	}
	tx, ok := s.grantCoinsLocked(&p, CoinAdmin, "admin", adminID, reason, amount)
	if !ok {
		return CoinTransaction{}, "saldo tidak boleh negatif"
	}
	s.blob.Profiles[userID] = p
	if err := s.flushLocked(); err != nil {
		s.blob = backup
		return CoinTransaction{}, "gagal menyimpan"
	}
	return tx, ""
}

func (s *ProgressStore) PublicView(p PlayerProfile, unlocks []UserAchievement) map[string]any {
	lvl, into, need, nextAt := XPToNext(p.XP)
	winRate := 0.0
	if p.TotalMatches > 0 {
		winRate = float64(p.Wins) / float64(p.TotalMatches) * 100
	}
	acc := 0.0
	if p.TotalQuestions > 0 {
		acc = float64(p.CorrectAnswers) / float64(p.TotalQuestions) * 100
	}
	sub := map[string]any{}
	for _, name := range eduSubjects {
		t := p.SubjectTotal[name]
		c := p.SubjectCorrect[name]
		pct := 0.0
		if t > 0 {
			pct = float64(c) / float64(t) * 100
		}
		sub[name] = map[string]any{"correct": c, "total": t, "accuracy": pct}
	}
	got := map[string]bool{}
	for _, u := range unlocks {
		got[u.AchievementID] = true
	}
	achs := make([]map[string]any, 0, len(AchievementCatalog))
	for _, a := range AchievementCatalog {
		achs = append(achs, map[string]any{
			"id": a.ID, "name": a.Name, "description": a.Description, "icon": a.Icon,
			"rewardXp": a.RewardXP, "rewardCoins": a.RewardCoins, "unlocked": got[a.ID],
		})
	}
	titles := make([]map[string]any, 0, len(TitleCatalog))
	for _, t := range TitleCatalog {
		open := t.Requires == "" || got[t.Requires]
		titles = append(titles, map[string]any{"id": t.ID, "name": t.Name, "unlocked": open})
	}
	rs := p.RankState()
	return map[string]any{
		"id": p.ID, "userId": p.UserID, "username": p.Username, "avatar": p.Avatar, "title": p.Title,
		"level": lvl, "xp": p.XP, "xpIntoLevel": into, "xpToNext": need, "xpNextAt": nextAt,
		"coins": p.Coins, "trophies": p.Trophies, "totalMatches": p.TotalMatches, "wins": p.Wins, "losses": p.Losses, "draws": p.Draws,
		"winRate": winRate, "totalQuestions": p.TotalQuestions, "correctAnswers": p.CorrectAnswers,
		"wrongAnswers": p.WrongAnswers, "timeoutAnswers": p.TimeoutAnswers, "accuracy": acc,
		"currentWinStreak": p.CurrentWinStreak, "bestWinStreak": p.BestWinStreak,
		"dailyStreak": p.DailyStreak, "subjectAccuracy": sub, "achievements": achs,
		"titles": titles, "avatars": ValidAvatars, "createdAt": p.CreatedAt, "updatedAt": p.UpdatedAt,
		"rankXp": 0, "rankWins": 0, "rankAccuracy": 0,
		"rankTier": rs.Tier, "rankDivision": rs.Division, "rankRr": rs.RR, "rankLabel": rs.Label(),
		"rankIndex": rs.Index, "peakTier": rs.PeakTier, "peakDivision": rs.PeakDivision, "peakRr": rs.PeakRR,
		"rankedWins": p.RankedWins, "seasonId": rs.SeasonID, "seasonName": CurrentSeason.Name,
		"abandonCount": p.AbandonCount,
	}
}

func (s *ProgressStore) PublicSafe(userID string) map[string]any {
	p, ok := s.Get(userID)
	if !ok {
		return nil
	}
	view := s.PublicView(p, s.Unlocks(userID))
	delete(view, "coins")
	delete(view, "email")
	delete(view, "avatars")
	delete(view, "titles")
	return view
}

func (s *ProgressStore) ViewFor(userID, username string) map[string]any {
	p := s.Ensure(userID, username)
	view := s.PublicView(p, s.Unlocks(userID))
	view["rankXp"] = s.rankBy(userID, func(o PlayerProfile) int { return o.XP })
	view["rankWins"] = s.rankBy(userID, func(o PlayerProfile) int { return o.Wins })
	view["myRank"] = s.rankBy(userID, func(o PlayerProfile) int { return o.RankState().Index })
	view["rankAccuracy"] = s.rankBy(userID, func(o PlayerProfile) int {
		if o.TotalQuestions <= 0 {
			return 0
		}
		return (o.CorrectAnswers * 10000) / o.TotalQuestions
	})
	return view
}

func (s *ProgressStore) rankBy(userID string, score func(PlayerProfile) int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	self, ok := s.blob.Profiles[userID]
	if !ok {
		return 0
	}
	mine := score(self)
	rank := 1
	for _, o := range s.blob.Profiles {
		if score(o) > mine {
			rank++
		}
	}
	return rank
}

func (s *ProgressStore) RecordAbandon(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.blob.Profiles[userID]
	if !ok {
		return
	}
	p.AbandonCount++
	idx := p.AbandonCount - 1
	if idx >= len(AbandonRestrictMin) {
		idx = len(AbandonRestrictMin) - 1
	}
	p.RestrictUntil = time.Now().Add(time.Duration(AbandonRestrictMin[idx]) * time.Minute).UnixMilli()
	p.UpdatedAt = time.Now().UnixMilli()
	s.blob.Profiles[userID] = p
	_ = s.flushLocked()
}

func (s *ProgressStore) RankHistory(userID string, page int) []RankChange {
	s.mu.Lock()
	defer s.mu.Unlock()
	if page < 0 {
		page = 0
	}
	own := make([]RankChange, 0)
	for i := len(s.blob.RankChanges) - 1; i >= 0; i-- {
		if s.blob.RankChanges[i].UserID == userID {
			own = append(own, s.blob.RankChanges[i])
		}
	}
	start := page * MaxMatchHistoryPage
	if start >= len(own) {
		return []RankChange{}
	}
	end := start + MaxMatchHistoryPage
	if end > len(own) {
		end = len(own)
	}
	return own[start:end]
}

func (s *ProgressStore) TxCoins(userID string, page int) []CoinTransaction {
	s.mu.Lock()
	defer s.mu.Unlock()
	own := make([]CoinTransaction, 0)
	for i := len(s.blob.Coins) - 1; i >= 0; i-- {
		if s.blob.Coins[i].UserID == userID {
			own = append(own, s.blob.Coins[i])
		}
	}
	return pageSlice(own, page, 20)
}

func (s *ProgressStore) TxXP(userID string, page int) []XPTransaction {
	s.mu.Lock()
	defer s.mu.Unlock()
	own := make([]XPTransaction, 0)
	for i := len(s.blob.XP) - 1; i >= 0; i-- {
		if s.blob.XP[i].UserID == userID {
			own = append(own, s.blob.XP[i])
		}
	}
	return pageSlice(own, page, 20)
}

func (s *ProgressStore) SoftResetAllRanks() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	backup := cloneBlob(s.blob)
	n := 0
	season := CurrentSeason.ID
	if o := liveOps(); o != nil {
		season = o.ActiveSeason().ID
	}
	for id, p := range s.blob.Profiles {
		rs := SoftResetRank(p.RankState())
		rs.SeasonID = season
		applyRankToProfile(&p, rs)
		s.blob.Profiles[id] = p
		n++
	}
	if err := s.flushLocked(); err != nil {
		s.blob = backup
		return 0
	}
	return n
}

func (s *ProgressStore) AllProfiles() []PlayerProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]PlayerProfile, 0, len(s.blob.Profiles))
	for _, p := range s.blob.Profiles {
		out = append(out, p)
	}
	return out
}

func (s *ProgressStore) SetRankForTest(userID string, st RankState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.blob.Profiles[userID]
	applyRankToProfile(&p, st)
	s.blob.Profiles[userID] = p
}
