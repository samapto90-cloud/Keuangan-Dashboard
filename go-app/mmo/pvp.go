package mmo

import (
	"strings"
	"time"
)

const (
	PvpWaiting    = "WAITING"
	PvpMatched    = "MATCHED"
	PvpReadySt    = "READY"
	PvpLoading    = "LOADING"
	PvpCountdown  = "COUNTDOWN"
	PvpActive     = "ACTIVE"
	PvpEnding     = "ENDING"
	PvpCompleted  = "COMPLETED"
	PvpCancelled  = "CANCELLED"
	PvpCleanup    = "CLEANUP"
	PvpRejoin     = 120 * time.Second
	PvpArenaLimit = 22.0
	PvpBgLimit    = 40.0
)

type PvpQueueEntry struct {
	PlayerID string
	Mode     string
	Rating   int
	Region   string
	Latency  int
	PartyID  string
	PartyN   int
	JoinedAt time.Time
}

type PvpShrine struct {
	ID           string
	X, Z, Radius float64
	Owner        int
	ProgressA    float64
	ProgressB    float64
	Contested    bool
	LastTick     time.Time
}

type ReplayEvent struct {
	T        int64          `json:"t"`
	Kind     string         `json:"kind"`
	PlayerID string         `json:"playerId"`
	Data     map[string]any `json:"data,omitempty"`
}

type PvpFighter struct {
	PlayerID      string
	Name          string
	Team          int
	Rating        int
	SnapLevel     int
	SnapAtk       int
	SnapDef       int
	SnapHP        int
	SnapMaxHP     int
	Skills        []string
	Forms         []string
	DmgDealt      int
	DmgTaken      int
	Healing       int
	Kills         int
	Assists       int
	Deaths        int
	Objective     int
	DamageFrom    map[string]int
	CCHits        map[string]int
	CCImmuneUntil map[string]time.Time
	LastInput     time.Time
	AfkWarned     bool
	AfkFlagged    bool
	DisconnectAt  time.Time
	SpectateID    string
	Return        [3]float64
	WarnedAt      time.Time
}

type PvpInstance struct {
	ID             string
	Mode           string
	Map            string
	State          string
	Region         string
	Players        []string
	TeamA, TeamB   []string
	Fighters       map[string]*PvpFighter
	CreatedAt      time.Time
	MatchedAt      time.Time
	StartedAt      time.Time
	EndedAt        time.Time
	EndsAt         time.Time
	CountdownUntil time.Time
	ProtectUntil   time.Time
	ReadyUntil     time.Time
	ScoreA, ScoreB int
	Shrines        []*PvpShrine
	KillFeed       []string
	Replay         []ReplayEvent
	Processed      bool
	RewardID       string
	Ready          map[string]bool
	WinnerTeam     int
	WinnerID       string
	Result         string
	Offline        map[string]time.Time
	Ranked         bool
}

type PvpHub struct {
	instances map[string]*PvpInstance
	byPlayer  map[string]string
	queue     []PvpQueueEntry
	pending   map[string]*rejoinSlot
	history   []PvpHistoryRow
	lb        LeaderboardCache
	abuse      map[string]*PvpAbuse
	reports    []PvpReport
	replays    map[string][]ReplayEvent
	duels      map[string]*pvpDuelInvite
	spectators map[string]string
}

type PvpAbuse struct {
	Leaves         int
	Afk            int
	Suspicious     float64
	QueueLockUntil time.Time
	PairWins       map[string]int
	LastOpp        string
	FeedFlags      int
}

type PvpReport struct {
	ID, MatchID, Reporter, Target, Category string
	At                                      time.Time
}

type PvpHistoryRow struct {
	MatchID      string `json:"matchId"`
	Mode         string `json:"mode"`
	Opponent     string `json:"opponent"`
	Result       string `json:"result"`
	RatingChange int    `json:"ratingChange"`
	Date         int64  `json:"date"`
	PlayerID     string `json:"playerId,omitempty"`
	Duration     int    `json:"duration,omitempty"`
	Kills        int    `json:"kills,omitempty"`
	Deaths       int    `json:"deaths,omitempty"`
	Assists      int    `json:"assists,omitempty"`
	Damage       int    `json:"damage,omitempty"`
	Objective    int    `json:"objective,omitempty"`
}

type LeaderboardCache struct {
	Global   []LBEntry            `json:"global"`
	Regional map[string][]LBEntry `json:"regional"`
	BuiltAt  time.Time            `json:"builtAt"`
}

type LBEntry struct {
	Place    int     `json:"rank"`
	PlayerID string  `json:"playerId"`
	Name     string  `json:"player"`
	RankName string  `json:"rankName"`
	Rating   int     `json:"rating"`
	Wins     int     `json:"wins"`
	Losses   int     `json:"losses"`
	WinRate  float64 `json:"winRate"`
	Region   string  `json:"region,omitempty"`
	Guild    string  `json:"guild,omitempty"`
}

func NewPvpHub() *PvpHub {
	return &PvpHub{
		instances: map[string]*PvpInstance{},
		byPlayer:  map[string]string{},
		pending:   map[string]*rejoinSlot{},
		abuse:      map[string]*PvpAbuse{},
		replays:    map[string][]ReplayEvent{},
		duels:      map[string]*pvpDuelInvite{},
		spectators: map[string]string{},
		lb:         LeaderboardCache{Regional: map[string][]LBEntry{}},
	}
}

func (w *WorldState) ApplyPvp(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	p.LastInputAt = time.Now()
	if w.PvP == nil {
		w.PvP = NewPvpHub()
	}
	var in PvpActionIn
	_ = unmarshal(env.Data, &in)
	switch env.Type {
	case TypeSetRating, TypeSetRank, TypePvpWin, TypeSetDamage:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	case TypeGetPvp:
		return [][]byte{marshal(TypePvpLobby, w.pvpLobby(p))}
	case TypePvpQueueJoin:
		return w.pvpQueueJoin(p, in.Mode)
	case TypePvpQueueLeave:
		return w.pvpQueueLeave(p)
	case TypePvpReady:
		return w.pvpReady(p, true)
	case TypePvpDecline:
		return w.pvpReady(p, false)
	case TypePvpLeave:
		return w.pvpLeave(p, true)
	case TypePvpEmote:
		return w.pvpEmote(p, in.Emote)
	case TypePvpSpectate:
		return w.pvpSpectate(p, in.TargetID)
	case TypePvpReport:
		return w.pvpReport(p, in)
	case TypePvpLeaderboard:
		return [][]byte{marshal(TypePvpLeaderboard, w.pvpLeaderboard(p, in.Board))}
	case TypePvpHistory:
		return [][]byte{marshal(TypePvpHistory, map[string]any{"history": w.pvpHistory(p), "toId": p.ID})}
	case TypePvpShopBuy:
		return w.pvpShopBuy(p, in.ShopItemID, in.TransactionID)
	case TypePvpTraining:
		return w.pvpEnterTraining(p)
	case TypePvpDuel:
		return w.pvpDuelChallenge(p, in.TargetID)
	case TypePvpDuelAccept:
		return w.pvpDuelRespond(p, true)
	case TypePvpDuelDecline:
		return w.pvpDuelRespond(p, false)
	case TypeGetReplay:
		return w.pvpGetReplay(p, in.MatchID)
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) pvpOf(playerID string) *PvpInstance {
	if w.PvP == nil {
		return nil
	}
	id := w.PvP.byPlayer[playerID]
	if id == "" {
		return nil
	}
	return w.PvP.instances[id]
}

func (w *WorldState) pvpLimit(p *Player) float64 {
	inst := w.pvpOf(p.ID)
	if inst == nil {
		return DungeonLimit
	}
	if pvpIsBattleground(inst) {
		return PvpBgLimit
	}
	return PvpArenaLimit
}

func (w *WorldState) pvpLobby(p *Player) PvpLobbyOut {
	log := p.ensureLog()
	w.ensurePvpRating(log)
	rk := rankForRating(log.PvpRating)
	rankName := rankDisplayName(log.PvpRating)
	if log.PvpPlacementLeft > 0 && log.PvpRankedMatches < pvpMod.Placement {
		rankName = "UNRANKED"
	}
	modes := make([]PvpModeView, 0, len(pvpModeCatalog))
	for _, m := range pvpModeCatalog {
		st := "LOCKED"
		if p.Level >= m.MinLevel {
			st = "AVAILABLE"
		}
		if !m.Enabled {
			st = "SOON"
		}
		modes = append(modes, PvpModeView{
			ID: m.ID, Name: m.Name, Kind: m.Kind, TeamSize: m.TeamSize, MinLevel: m.MinLevel,
			Duration: m.Duration, Map: m.Map, Enabled: m.Enabled, Status: st,
		})
	}
	shop := make([]PvpShopView, 0, len(pvpShop.Items))
	for _, it := range pvpShop.Items {
		shop = append(shop, PvpShopView{ShopItemID: it.ShopItemID, ItemID: it.ItemID, Name: it.Name, Kind: it.Kind, Price: it.Price, Currency: it.Currency})
	}
	out := PvpLobbyOut{
		Modes: modes,
		Profile: PvpProfileView{
			Rating: log.PvpRating, Rank: rk.ID, RankName: rankName, Division: rankDivision(log.PvpRating, rk),
			RankVisual: rankVisual(rk), Wins: log.PvpWins, Losses: log.PvpLosses,
			WinRate: winRate(log.PvpWins, log.PvpLosses), PlacementLeft: log.PvpPlacementLeft,
			WinStreak: log.PvpWinStreak, LossStreak: log.PvpLossStreak, HighestRank: log.PvpHighestRank,
			BattleToken: log.BattleToken, Season: pvpSeasonDef.Name, SeasonID: pvpSeasonDef.ID,
		},
		Season:        PvpSeasonView{ID: pvpSeasonDef.ID, Name: pvpSeasonDef.Name, Number: pvpSeasonDef.Number, Start: pvpSeasonDef.Start, End: pvpSeasonDef.End, Weeks: pvpSeasonDef.Weeks},
		Shop:          shop,
		Rewards:       w.pvpRewardViews(log),
		History:       w.pvpHistory(p),
		SeasonHistory: append([]PvpSeasonHistoryRow{}, log.PvpSeasonHistory...),
		Nearby:        w.pvpNearby(p),
		ToID:          p.ID,
	}
	if q := w.pvpQueueOf(p.ID); q != nil {
		def, _ := pvpMode(q.Mode)
		need := def.TeamSize * 2
		have := w.pvpQueueCount(q.Mode)
		est, note := pvpQueueEstimate(need, have, time.Since(q.JoinedAt))
		out.Queue = &PvpQueueView{
			State: "QUEUED", Mode: q.Mode, Name: def.Name, Players: have,
			Need: need, WaitMs: time.Since(q.JoinedAt).Milliseconds(), WaitEstMs: est, WaitNote: note, ToID: p.ID,
		}
	}
	if inst := w.pvpOf(p.ID); inst != nil {
		out.Match = w.pvpView(inst, p.ID)
	}
	return out
}

func winRate(w, l int) float64 {
	t := w + l
	if t <= 0 {
		return 0
	}
	return float64(w) / float64(t)
}

func (w *WorldState) pvpEmote(p *Player, emote string) [][]byte {
	emote = strings.ToLower(strings.TrimSpace(emote))
	switch emote {
	case "wave", "cheer", "bow", "laugh":
	default:
		return rejectFor(p.ID, TypePvpEmote, "emote")
	}
	if w.limited("emote:"+p.ID, 3, time.Duration(pvpMod.EmoteCooldown)*time.Second) {
		return rejectFor(p.ID, TypePvpEmote, "rate")
	}
	inst := w.pvpOf(p.ID)
	id := ""
	if inst != nil {
		id = inst.ID
		w.pvpReplay(inst, "emote", p.ID, map[string]any{"emote": emote})
	}
	return [][]byte{marshal(TypePvpEmote, map[string]any{"playerId": p.ID, "emote": emote, "instanceId": id})}
}

func (w *WorldState) pvpSpectate(p *Player, targetID string) [][]byte {
	inst := w.pvpOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypePvpSpectate, "match")
	}
	if !pvpAllowSpectate(inst) {
		return rejectFor(p.ID, TypePvpSpectate, "friendly")
	}
	f := inst.Fighters[p.ID]
	if f == nil {
		return rejectFor(p.ID, TypePvpSpectate, "fighter")
	}
	pl := w.players[p.ID]
	if pl != nil && pl.alive() && pl.CombatState != "DEAD" && pl.CombatState != "RESPAWNING" {
		return rejectFor(p.ID, TypePvpSpectate, "alive")
	}
	t := inst.Fighters[targetID]
	if t == nil {
		return rejectFor(p.ID, TypePvpSpectate, "target")
	}
	sameTeam := t.Team == f.Team
	if !sameTeam && len(inst.TeamA) == 1 && len(inst.TeamB) == 1 {
		sameTeam = true
	}
	if !sameTeam {
		return rejectFor(p.ID, TypePvpSpectate, "team")
	}
	f.SpectateID = targetID
	return [][]byte{marshal(TypePvpSpectate, map[string]any{
		"playerId": p.ID, "targetId": targetID, "mode": "PLAYER", "instanceId": inst.ID, "toId": p.ID,
	})}
}

func (w *WorldState) pvpReport(p *Player, in PvpActionIn) [][]byte {
	if in.TargetID == "" || in.Category == "" {
		return rejectFor(p.ID, TypePvpReport, "payload")
	}
	if w.limited("pvprep:"+p.ID, 8, time.Hour) {
		return rejectFor(p.ID, TypePvpReport, "rate")
	}
	matchID := in.MatchID
	if inst := w.pvpOf(p.ID); inst != nil {
		matchID = inst.ID
	}
	rep := PvpReport{ID: randomID("prep_"), MatchID: matchID, Reporter: p.ID, Target: in.TargetID, Category: in.Category, At: time.Now()}
	w.PvP.reports = append(w.PvP.reports, rep)
	if w.Reports == nil {
		w.Reports = []Report{}
	}
	w.Reports = append(w.Reports, Report{
		ID: rep.ID, Reporter: p.ID, Target: in.TargetID, Category: in.Category,
		Evidence: "pvp:" + matchID, State: "OPEN", At: time.Now(), MatchID: matchID,
	})
	return [][]byte{w.notify(p, "system", "Laporan PvP diterima.")}
}

func (w *WorldState) pvpLeave(p *Player, penalize bool) [][]byte {
	inst := w.pvpOf(p.ID)
	if inst == nil {
		return w.pvpQueueLeave(p)
	}
	if inst.State == PvpActive || inst.State == PvpCountdown || inst.State == PvpLoading {
		if inst.Ranked && penalize {
			w.flagLeaver(p.ID)
			w.finishPvp(inst, otherTeamOf(inst, p.ID), "leave")
		} else if penalize && inst.State == PvpActive {
			w.flagLeaver(p.ID)
			w.finishPvp(inst, otherTeamOf(inst, p.ID), "leave")
		}
	}
	w.extractPvpPlayer(p, inst)
	return [][]byte{marshal(TypePvpLeft, map[string]any{"toId": p.ID})}
}

func otherTeamOf(inst *PvpInstance, playerID string) int {
	f := inst.Fighters[playerID]
	if f == nil {
		return 1
	}
	if f.Team == 1 {
		return 2
	}
	return 1
}

func (w *WorldState) extractPvpPlayer(p *Player, inst *PvpInstance) {
	if inst == nil || p == nil {
		return
	}
	if f := inst.Fighters[p.ID]; f != nil {
		p.X, p.Y, p.Z = f.Return[0], f.Return[1], f.Return[2]
	}
	p.InstanceID = ""
	delete(w.PvP.byPlayer, p.ID)
	if p.CombatState == "DEAD" || p.CombatState == "RESPAWNING" || p.CombatState == "DOWNED" {
		p.CombatState = "IDLE"
		p.State = "IDLE"
		if p.HP <= 0 {
			p.HP = p.MaxHP
		}
	}
}

func (w *WorldState) flagLeaver(id string) {
	ab := w.pvpAbuse(id)
	ab.Leaves++
	lock := time.Duration(pvpMod.LeaveLock) * time.Second
	if ab.Leaves >= 3 {
		lock *= time.Duration(ab.Leaves)
	}
	ab.QueueLockUntil = time.Now().Add(lock)
}

func (w *WorldState) pvpAbuse(id string) *PvpAbuse {
	if w.PvP.abuse[id] == nil {
		w.PvP.abuse[id] = &PvpAbuse{PairWins: map[string]int{}}
	}
	return w.PvP.abuse[id]
}

func (w *WorldState) pvpReplay(inst *PvpInstance, kind, playerID string, data map[string]any) {
	if inst == nil {
		return
	}
	inst.Replay = append(inst.Replay, ReplayEvent{T: time.Now().UnixMilli(), Kind: kind, PlayerID: playerID, Data: data})
	if len(inst.Replay) > 500 {
		inst.Replay = inst.Replay[len(inst.Replay)-400:]
	}
}

func (h *PvpHub) ensure() {
	if h.instances == nil {
		h.instances = map[string]*PvpInstance{}
	}
	if h.byPlayer == nil {
		h.byPlayer = map[string]string{}
	}
	if h.pending == nil {
		h.pending = map[string]*rejoinSlot{}
	}
	if h.abuse == nil {
		h.abuse = map[string]*PvpAbuse{}
	}
	if h.replays == nil {
		h.replays = map[string][]ReplayEvent{}
	}
	if h.duels == nil {
		h.duels = map[string]*pvpDuelInvite{}
	}
	if h.spectators == nil {
		h.spectators = map[string]string{}
	}
	if h.lb.Regional == nil {
		h.lb.Regional = map[string][]LBEntry{}
	}
}
