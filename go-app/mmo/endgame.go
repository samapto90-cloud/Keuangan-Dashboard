package mmo

import (
	"strings"
	"time"

	_ "embed"
)

//go:embed data/endgame.json
var endgameJSON []byte

//go:embed data/dailyQuests.json
var dailyQuestsJSON []byte

//go:embed data/weeklyQuests.json
var weeklyQuestsJSON []byte

//go:embed data/challenges.json
var challengesJSON []byte

//go:embed data/seasonTrack.json
var seasonTrackJSON []byte

//go:embed data/horizon.json
var horizonJSON []byte

//go:embed data/liveEvents.json
var liveEventsJSON []byte

type EndgameCfg struct {
	UnlockLevel int    `json:"unlockLevel"`
	HubRegion   string `json:"hubRegion"`
	Timezone    string `json:"timezone"`
}

type EndgameQuestDef struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Kind          string    `json:"kind"`
	Target        string    `json:"target"`
	Count         int       `json:"count"`
	Rewards       RewardDef `json:"rewards"`
	SeasonXP      int       `json:"seasonXP"`
	BattleToken   int       `json:"battleToken"`
	GuardianToken int       `json:"guardianToken"`
	Fragment      int       `json:"fragment"`
}

type ChallengeDef struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	Tier           string `json:"tier"`
	Title          string `json:"title"`
	Metric         string `json:"metric"`
	Target         string `json:"target"`
	Count          int    `json:"count"`
	SeasonXP       int    `json:"seasonXP"`
	RewardTitle    string `json:"rewardTitle"`
	RewardCosmetic string `json:"rewardCosmetic"`
}

type SeasonTrackDef struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Theme      string            `json:"theme"`
	Start      string            `json:"start"`
	End        string            `json:"end"`
	Weeks      int               `json:"weeks"`
	MaxLevel   int               `json:"maxLevel"`
	XPPerLevel int               `json:"xpPerLevel"`
	FreeTrack  bool              `json:"freeTrack"`
	Activities []SeasonActivity  `json:"activities"`
	Rewards    []SeasonRewardDef `json:"rewards"`
}

type SeasonActivity struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	XP    int    `json:"xp"`
}

type SeasonRewardDef struct {
	Level int    `json:"level"`
	Kind  string `json:"kind"`
	ID    string `json:"id"`
}

type SeasonHistoryRow struct {
	SeasonID   string `json:"seasonId"`
	Name       string `json:"name"`
	Level      int    `json:"level"`
	XP         int    `json:"xp"`
	Guardians  int    `json:"guardians"`
	Rank       string `json:"rank"`
	Ended      int64  `json:"ended"`
}

type HorizonCfg struct {
	DungeonID       string          `json:"dungeonId"`
	MaxLevel        int             `json:"maxLevel"`
	MaxModifiers    int             `json:"maxModifiers"`
	ScoreTimeWeight int             `json:"scoreTimeWeight"`
	ScoreDeathPen   int             `json:"scoreDeathPenalty"`
	Modifiers       []HorizonModDef `json:"modifiers"`
	InvalidPairs    [][]string      `json:"invalidPairs"`
}

type HorizonModDef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type HorizonScore struct {
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
	Level    int    `json:"level"`
	Week     string `json:"week"`
}

type LiveWorldCfg struct {
	Timezone          string              `json:"timezone"`
	WorldBossRotation []BossRotationDef   `json:"worldBossRotation"`
	EventRotation     []string            `json:"eventRotation"`
	Community         CommunityEventDef   `json:"community"`
}

type BossRotationDef struct {
	Weekday  int    `json:"weekday"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Announce string `json:"announce"`
}

type CommunityEventDef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Target   int    `json:"target"`
	Cosmetic string `json:"cosmetic"`
	DailyCap int    `json:"dailyCap"`
	Start    string `json:"start"`
	End      string `json:"end"`
}

type CommunityLive struct {
	ID           string
	Name         string
	Points       int
	Target       int
	State        string
	Cosmetic     string
	DailyCap     int
	Granted      map[string]bool
	Participants map[string]int
}

type CraftingService struct {
	Ready bool
}

var (
	endgameCfg     EndgameCfg
	dailyCatalog   []EndgameQuestDef
	weeklyCatalog  []EndgameQuestDef
	challengeCat   []ChallengeDef
	seasonTrack    SeasonTrackDef
	horizonCfg     HorizonCfg
	liveWorldCfg   LiveWorldCfg
	craftingHook   = CraftingService{Ready: true}
)

func init() {
	mustJSON("endgame.json", endgameJSON, &endgameCfg)
	if endgameCfg.UnlockLevel < 1 {
		endgameCfg.UnlockLevel = 40
	}
	mustJSON("dailyQuests.json", dailyQuestsJSON, &dailyCatalog)
	mustJSON("weeklyQuests.json", weeklyQuestsJSON, &weeklyCatalog)
	mustJSON("challenges.json", challengesJSON, &challengeCat)
	mustJSON("seasonTrack.json", seasonTrackJSON, &seasonTrack)
	mustJSON("horizon.json", horizonJSON, &horizonCfg)
	mustJSON("liveEvents.json", liveEventsJSON, &liveWorldCfg)
	if seasonTrack.XPPerLevel < 1 {
		seasonTrack.XPPerLevel = 100
	}
	if seasonTrack.MaxLevel < 1 {
		seasonTrack.MaxLevel = 50
	}
}

func copyIntMap(src, dst map[string]int) {
	for k, v := range src {
		dst[k] = v
	}
}

func copyBoolMap(src, dst map[string]bool) {
	for k, v := range src {
		dst[k] = v
	}
}

func (p *Player) endgameUnlocked() bool {
	if p == nil {
		return false
	}
	log := p.ensureLog()
	if log.Flags["storyCompleted"] || log.Flags["endgame_unlocked"] {
		return true
	}
	return p.Level >= endgameCfg.UnlockLevel
}

func (w *WorldState) ensureEndgameClock(p *Player) {
	log := p.ensureLog()
	day := utcDayKey()
	week := raidWeekKey()
	if log.DailyDay != day {
		log.DailyDay = day
		log.DailyProgress = map[string]int{}
		log.DailyClaimed = map[string]bool{}
		log.SeasonXPToday = 0
		log.SeasonXPDay = day
		log.LiveDay = day
		log.LiveDayAmt = 0
		w.notify(p, "daily", "Daily Reset")
	}
	if log.WeeklyWeek != week {
		log.WeeklyWeek = week
		log.WeeklyProgress = map[string]int{}
		log.WeeklyClaimed = map[string]bool{}
		log.WeeklyBonusClaimed = false
		if log.HorizonWeek != week {
			log.HorizonWeek = week
			log.HorizonBest = 0
			log.HorizonClaimWeek = ""
		}
		w.notify(p, "weekly", "Weekly Reset")
	}
	if log.SeasonTrackID == "" {
		log.SeasonTrackID = seasonTrack.ID
	}
	if log.SeasonLevel < 1 {
		log.SeasonLevel = 1
	}
	if p.endgameUnlocked() {
		log.Flags["endgame_unlocked"] = true
	}
	w.maybeRotateSeason(p, time.Now().UTC())
}

func (w *WorldState) ApplyEndgame(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	w.ensureEndgameClock(p)
	switch env.Type {
	case TypeGetEndgame, TypeGetSeason, TypeGetHorizon, TypeGetCalendar, TypeGetLiveEvent, TypeGetLearning, TypeGetLoreBook, TypeGetLeaderboards:
		return [][]byte{marshal(TypeEndgameState, w.endgameView(p))}
	case TypeClaimDaily:
		var in struct{ ID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.claimEndgameQuest(p, in.ID, in.TransactionID, true)
	case TypeClaimWeekly:
		var in struct{ ID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.claimEndgameQuest(p, in.ID, in.TransactionID, false)
	case TypeClaimChallenge:
		var in struct{ ID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.claimChallenge(p, in.ID, in.TransactionID)
	case TypeClaimSeason:
		var in struct{ Level int; TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.claimSeasonLevel(p, in.Level, in.TransactionID)
	case TypeSetShowcase:
		var in struct{ Title, Badge, Aura, Mount string }
		_ = unmarshal(env.Data, &in)
		return w.setShowcase(p, in.Title, in.Badge, in.Aura, in.Mount)
	case TypeGetPublicProfile:
		var in struct{ TargetID string }
		_ = unmarshal(env.Data, &in)
		return w.publicProfile(p, in.TargetID)
	case TypeSetSeasonXP, TypeUnlockAchievement, TypeUnlockCosmetic, TypeSetAchievement, TypeContributeEvent:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) claimEndgameQuest(p *Player, id, tx string, daily bool) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	if !p.endgameUnlocked() {
		return rejectFor(p.ID, TypeClaimDaily, "locked")
	}
	log := p.ensureLog()
	var def EndgameQuestDef
	claimed := log.DailyClaimed
	progress := log.DailyProgress
	action := TypeClaimDaily
	cat := dailyCatalog
	if !daily {
		claimed, progress, action, cat = log.WeeklyClaimed, log.WeeklyProgress, TypeClaimWeekly, weeklyCatalog
	}
	ok := false
	for _, d := range cat {
		if d.ID == id {
			def, ok = d, true
			break
		}
	}
	if !ok {
		return rejectFor(p.ID, action, "quest")
	}
	need := def.Count
	if need < 1 {
		need = 1
	}
	if progress[id] < need {
		return rejectFor(p.ID, action, "progress")
	}
	if claimed[id] {
		return rejectFor(p.ID, action, "claimed")
	}
	claimed[id] = true
	events := w.giveExp(p, def.Rewards.Exp)
	w.giveCurrency(p, def.Rewards.Coin, def.Rewards.Crystal)
	if def.Rewards.EduToken > 0 {
		log.EduToken += def.Rewards.EduToken
	}
	if def.Rewards.BattleToken > 0 {
		log.BattleToken += def.Rewards.BattleToken
	}
	if def.BattleToken > 0 {
		log.BattleToken += def.BattleToken
	}
	if def.GuardianToken > 0 {
		log.GuardianTokens += def.GuardianToken
	}
	if def.Fragment > 0 {
		log.HorizonFragments += def.Fragment
	}
	events = append(events, w.grantSeasonXP(p, def.SeasonXP, "quest:"+id)...)
	if !daily {
		n := 0
		for _, d := range weeklyCatalog {
			if log.WeeklyClaimed[d.ID] {
				n++
			}
		}
		if n >= 3 && !log.WeeklyBonusClaimed {
			log.WeeklyBonusClaimed = true
			log.HorizonFragments++
			events = append(events, w.grantSeasonXP(p, 40, "weekly-milestone")...)
		}
	}
	p.markDirty()
	w.persist(p)
	w.audit("rewardClaimed", p.ID, id)
	events = append(events, marshal(TypeEndgameState, w.endgameView(p)))
	return w.rememberTx(tx, events)
}

func (w *WorldState) claimChallenge(p *Player, id, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	log := p.ensureLog()
	var def ChallengeDef
	ok := false
	for _, d := range challengeCat {
		if d.ID == id {
			def, ok = d, true
			break
		}
	}
	if !ok {
		return rejectFor(p.ID, TypeClaimChallenge, "challenge")
	}
	if log.ChallengeClaimed[id] {
		return rejectFor(p.ID, TypeClaimChallenge, "claimed")
	}
	need := def.Count
	if need < 1 {
		need = 1
	}
	if log.ChallengeProgress[id] < need {
		return rejectFor(p.ID, TypeClaimChallenge, "progress")
	}
	log.ChallengeClaimed[id] = true
	if def.RewardTitle != "" {
		p.grantTitle(def.RewardTitle)
	}
	if def.RewardCosmetic != "" {
		p.grantCosmetic(def.RewardCosmetic)
	}
	events := w.grantSeasonXP(p, def.SeasonXP, "challenge:"+id)
	p.markDirty()
	w.persist(p)
	w.audit("rewardClaimed", p.ID, "ch:"+id)
	events = append(events, marshal(TypeEndgameState, w.endgameView(p)))
	return w.rememberTx(tx, events)
}

func (w *WorldState) noteActivity(p *Player, kind, target string, n int) [][]byte {
	if p == nil || n <= 0 {
		return nil
	}
	w.ensureEndgameClock(p)
	log := p.ensureLog()
	bump := func(m map[string]int, id string, add, cap int) {
		if m == nil {
			return
		}
		m[id] += add
		if cap > 0 && m[id] > cap {
			m[id] = cap
		}
	}
	for _, d := range dailyCatalog {
		if d.Kind == kind && (d.Target == "" || d.Target == target) {
			bump(log.DailyProgress, d.ID, n, d.Count)
		}
	}
	for _, d := range weeklyCatalog {
		if d.Kind == kind && (d.Target == "" || d.Target == target) {
			bump(log.WeeklyProgress, d.ID, n, d.Count)
		}
	}
	for _, d := range challengeCat {
		match := false
		switch d.Metric {
		case kind, strings.ToUpper(kind):
			match = d.Target == "" || d.Target == target
		case "KILL_NODEATH":
			match = kind == "KILL"
			if match {
				log.DeathlessKills += n
				log.ChallengeProgress[d.ID] = log.DeathlessKills
				continue
			}
		case "ZONE":
			match = kind == "VISIT" && target == d.Target
		case "ANSWER":
			match = kind == "ANSWER"
		case "DUNGEON_TIME":
			if kind == "DUNGEON_TIME" && target == d.Target && n > 0 && n <= d.Count {
				log.ChallengeProgress[d.ID] = d.Count
			}
			continue
		case "RAID_NOWIPE":
			match = kind == "RAID_NOWIPE"
		case "RANKED_WIN":
			match = kind == "RANKED_WIN"
		case "COSMETIC_KIND":
			match = kind == "COSMETIC"
		case "GUILD_DUNGEON":
			match = kind == "GUILD_DUNGEON"
		}
		if match {
			bump(log.ChallengeProgress, d.ID, n, d.Count)
		}
	}
	xp := 0
	switch kind {
	case "DUNGEON":
		xp = 100
	case "RAID":
		xp = 200
	case "EVENT":
		xp = 40
	case "VISIT":
		xp = 20
	case "PVP_MATCH", "PVP_WIN":
		xp = 60
	case "PVP_SCORE", "PVP_CAPTURE":
		xp = 8
	case "ANSWER":
		xp = 10
	}
	var out [][]byte
	if xp > 0 {
		out = append(out, w.grantSeasonXP(p, xp, kind)...)
	}
	out = append(out, w.addCommunityPoints(p, n*5)...)
	return out
}

func (w *WorldState) setShowcase(p *Player, title, badge, aura, mount string) [][]byte {
	log := p.ensureLog()
	if title != "" {
		owned := false
		for _, t := range log.Titles {
			if t == title {
				owned = true
			}
		}
		if !owned {
			return rejectFor(p.ID, TypeSetShowcase, "title")
		}
		log.ActiveTitle = title
		p.Title = titleByID[title].Name
	}
	if badge != "" || aura != "" {
		c := badge
		if aura != "" {
			c = aura
		}
		owned := false
		for _, x := range log.Cosmetics {
			if x == c {
				owned = true
			}
		}
		if !owned {
			return rejectFor(p.ID, TypeSetShowcase, "cosmetic")
		}
		if badge != "" {
			log.ShowcaseBadge = badge
		}
		if aura != "" {
			log.ShowcaseAura = aura
			log.ActiveCosmetic = aura
		}
	}
	if mount != "" {
		ok := false
		for _, m := range log.Mounts {
			if m == mount {
				ok = true
			}
		}
		if !ok {
			return rejectFor(p.ID, TypeSetShowcase, "mount")
		}
		log.ShowcaseMount = mount
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeEndgameState, w.endgameView(p))}
}

func (w *WorldState) publicProfile(viewer *Player, targetID string) [][]byte {
	if targetID == "" {
		targetID = viewer.ID
	}
	t := w.players[targetID]
	if t == nil {
		return rejectFor(viewer.ID, TypeGetPublicProfile, "target")
	}
	log := t.ensureLog()
	w.refreshTitles(t)
	gname, gtag := "", t.GuildTag
	if g := w.guildOf(t.ID); g != nil {
		gname, gtag = g.Name, g.Tag
	}
	return [][]byte{marshal(TypeInspectResult, InspectOut{
		PlayerID: t.ID, Name: t.Name, Level: t.Level, Class: t.Class,
		Stats: StatsView{HP: t.HP, MaxHP: t.MaxHP}, PowerRating: t.powerRating(),
		Guild: gname, GuildTag: gtag, Title: t.Title,
		Rank: log.PvpHighestRank, Season: seasonTrack.Name, SeasonLevel: log.SeasonLevel,
		Badge: log.ShowcaseBadge, Aura: log.ShowcaseAura, Mount: log.ShowcaseMount,
	})}
}

func (w *WorldState) endgameView(p *Player) map[string]any {
	log := p.ensureLog()
	type qv struct {
		ID, Title, Kind string
		Progress, Count int
		Claimed         bool
		Tier            string
	}
	dailies := make([]qv, 0, len(dailyCatalog))
	for _, d := range dailyCatalog {
		need := d.Count
		if need < 1 {
			need = 1
		}
		dailies = append(dailies, qv{ID: d.ID, Title: d.Title, Kind: d.Kind, Progress: log.DailyProgress[d.ID], Count: need, Claimed: log.DailyClaimed[d.ID]})
	}
	weeklies := make([]qv, 0, len(weeklyCatalog))
	for _, d := range weeklyCatalog {
		need := d.Count
		if need < 1 {
			need = 1
		}
		weeklies = append(weeklies, qv{ID: d.ID, Title: d.Title, Kind: d.Kind, Progress: log.WeeklyProgress[d.ID], Count: need, Claimed: log.WeeklyClaimed[d.ID]})
	}
	chs := make([]qv, 0, len(challengeCat))
	for _, d := range challengeCat {
		need := d.Count
		if need < 1 {
			need = 1
		}
		chs = append(chs, qv{ID: d.ID, Title: d.Title, Kind: d.Kind, Progress: log.ChallengeProgress[d.ID], Count: need, Claimed: log.ChallengeClaimed[d.ID], Tier: d.Tier})
	}
	need := seasonTrack.XPPerLevel
	if need < 1 {
		need = 100
	}
	into := log.SeasonXP % need
	next := nilSeasonReward(log.SeasonLevel + 1)
	cur := nilSeasonReward(log.SeasonLevel)
	acc := 0
	cats := map[string]int{}
	if log.EduAnswered > 0 {
		acc = 100 * log.EduCorrect / log.EduAnswered
	}
	for k, v := range log.EduCats {
		cats[k] = v
	}
	auraN, mountN, titleN := 0, 0, len(log.Titles)
	for _, c := range log.Cosmetics {
		if cosmeticByID[c].Kind == "aura" {
			auraN++
		}
		if cosmeticByID[c].Kind == "mount" {
			mountN++
		}
	}
	view := map[string]any{
		"unlocked": p.endgameUnlocked(), "hub": "Hall of Horizon", "level": p.Level,
		"seasonName": seasonTrack.Name, "seasonId": seasonTrack.ID, "seasonLevel": log.SeasonLevel,
		"seasonXP": into, "seasonXPNeed": need, "seasonEnd": seasonTrack.End,
		"currentReward": cur, "nextReward": next, "history": log.SeasonHistory,
		"daily": dailies, "weekly": weeklies, "challenges": chs,
		"achievements": log.Achievements, "titles": log.Titles, "cosmetics": log.Cosmetics,
		"collection": map[string]any{"aura": auraN, "auraTotal": countCosmeticKind("aura"), "mount": mountN, "mountTotal": countCosmeticKind("mount"), "titles": titleN, "titleTotal": len(titleByID)},
		"learning": map[string]any{"answered": log.EduAnswered, "correct": log.EduCorrect, "accuracy": acc, "streak": log.EduStreak, "categories": cats},
		"lore": loreBook(p), "horizon": w.horizonView(p), "calendar": w.eventCalendar(),
		"community": w.communityView(), "leaderboards": w.endgameBoards(),
		"fragments": log.HorizonFragments, "showcase": map[string]string{"title": log.ActiveTitle, "badge": log.ShowcaseBadge, "aura": log.ShowcaseAura, "mount": log.ShowcaseMount},
		"serverTime": time.Now().UTC().Format(time.RFC3339), "toId": p.ID,
	}
	phase29EnrichEndgameView(w, p, view)
	phase30EnrichEndgameView(view)
	return view
}

func countCosmeticKind(kind string) int {
	n := 0
	for _, c := range cosmeticByID {
		if c.Kind == kind {
			n++
		}
	}
	return n
}

func nilSeasonReward(level int) map[string]any {
	for _, r := range seasonTrack.Rewards {
		if r.Level == level {
			return map[string]any{"level": r.Level, "kind": r.Kind, "id": r.ID}
		}
	}
	return map[string]any{"level": level}
}

func loreBook(p *Player) []LoreView {
	log := p.ensureLog()
	out := make([]LoreView, 0)
	for _, l := range loreCatalog {
		if log.Lore[l.ID] {
			out = append(out, LoreView{ID: l.ID, Title: l.Title, Text: l.Text})
		}
	}
	return out
}

func (w *WorldState) endgameBoards() map[string]any {
	type row struct {
		Rank int
		Name string
		Score, Level int
		ID string
	}
	horizon := make([]row, 0)
	season := make([]row, 0)
	guilds := make([]row, 0)
	i := 1
	for _, p := range w.players {
		if p == nil {
			continue
		}
		log := p.ensureLog()
		if log.HorizonBest > 0 {
			horizon = append(horizon, row{Rank: i, Name: p.Name, Score: log.HorizonBest, Level: p.Level, ID: p.ID})
		}
		if log.SeasonXP > 0 || log.SeasonLevel > 0 {
			season = append(season, row{Rank: i, Name: p.Name, Score: log.SeasonLevel, Level: log.SeasonLevel, ID: p.ID})
		}
		i++
	}
	if w.Guilds != nil {
		r := 1
		for _, g := range w.Guilds.ByID {
			if g == nil {
				continue
			}
			guilds = append(guilds, row{Rank: r, Name: g.Name, Score: g.Exp, Level: g.Level, ID: g.ID})
			r++
		}
	}
	return map[string]any{"horizon": horizon, "season": season, "guilds": guilds, "week": raidWeekKey()}
}
