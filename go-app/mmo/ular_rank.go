package mmo

import "strings"

const (
	RankedMinLevel  = 5
	RankWinRR       = 25
	RankLossRR      = 20
	RRPerDivision   = 100
	MasterBand      = 200
	ChatBurstMax    = 5
	ChatBurstWin    = 10 // seconds
	InviteTTLSec    = 30
	QueueExpand1    = 60
	QueueExpand2    = 120
	MatchFoundSec   = 5
	ReadyCheckSec   = 15
	LeaderboardPage = 20
	LBCacheSec      = 15
	PresenceOnline  = 45
	PresenceAway    = 90
)

var AbandonRestrictMin = []int{5, 15, 30}

var RankTiers = []string{"BRONZE", "SILVER", "GOLD", "PLATINUM", "DIAMOND", "MASTER", "GRANDMASTER"}

var RankDivisions = []string{"III", "II", "I"}

type RankSeason struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	StartAt int64  `json:"startAt"`
	EndAt   int64  `json:"endAt"`
	Active  bool   `json:"active"`
}

var CurrentSeason = RankSeason{
	ID: "season-1", Name: "Nusantara Beginnings",
	StartAt: 1755648000000, EndAt: 1764547200000, Active: true,
}

type RankState struct {
	Tier         string `json:"tier"`
	Division     string `json:"division,omitempty"`
	RR           int    `json:"rr"`
	Index        int    `json:"index"`
	Protect      bool   `json:"protect,omitempty"`
	PeakTier     string `json:"peakTier,omitempty"`
	PeakDivision string `json:"peakDivision,omitempty"`
	PeakRR       int    `json:"peakRr,omitempty"`
	PeakIndex    int    `json:"peakIndex,omitempty"`
	SeasonID     string `json:"seasonId,omitempty"`
}

func defaultRank() RankState {
	return RankState{Tier: "BRONZE", Division: "III", RR: 0, Index: 0, SeasonID: CurrentSeason.ID, PeakTier: "BRONZE", PeakDivision: "III"}
}

func tierIndex(tier string) int {
	t := strings.ToUpper(strings.TrimSpace(tier))
	for i, n := range RankTiers {
		if n == t {
			return i
		}
	}
	return 0
}

func divIndex(div string) int {
	d := strings.ToUpper(strings.TrimSpace(div))
	for i, n := range RankDivisions {
		if n == d {
			return i
		}
	}
	return 0
}

// RankIndex maps a visible rank to a monotonic ladder index. Never negative.
func RankIndex(tier, division string, rr int) int {
	if rr < 0 {
		rr = 0
	}
	ti := tierIndex(tier)
	if ti >= 5 { // MASTER / GRANDMASTER
		base := 5 * 3 * RRPerDivision
		if ti == 5 {
			if rr > MasterBand-1 {
				rr = MasterBand - 1
			}
			return base + rr
		}
		return base + MasterBand + rr
	}
	if rr >= RRPerDivision {
		rr = RRPerDivision - 1
	}
	return ti*3*RRPerDivision + divIndex(division)*RRPerDivision + rr
}

func RankFromIndex(idx int) (tier, division string, rr int) {
	if idx < 0 {
		idx = 0
	}
	divSpan := 5 * 3 * RRPerDivision
	if idx < divSpan {
		ti := idx / (3 * RRPerDivision)
		rest := idx % (3 * RRPerDivision)
		di := rest / RRPerDivision
		rr = rest % RRPerDivision
		return RankTiers[ti], RankDivisions[di], rr
	}
	masterStart := divSpan
	if idx < masterStart+MasterBand {
		return "MASTER", "", idx - masterStart
	}
	return "GRANDMASTER", "", idx - masterStart - MasterBand
}

func RankLabel(tier, division string) string {
	tier = strings.ToUpper(tier)
	if division == "" {
		return tier
	}
	return tier + " " + strings.ToUpper(division)
}

func (r RankState) Label() string {
	return RankLabel(r.Tier, r.Division)
}

func (r RankState) Sync() RankState {
	r.Index = RankIndex(r.Tier, r.Division, r.RR)
	r.Tier, r.Division, r.RR = RankFromIndex(r.Index)
	if r.Index >= r.PeakIndex {
		r.PeakIndex = r.Index
		r.PeakTier, r.PeakDivision, r.PeakRR = r.Tier, r.Division, r.RR
	}
	if r.SeasonID == "" {
		r.SeasonID = CurrentSeason.ID
	}
	return r
}

// ApplyRank applies a ranked win/loss. RR never goes negative. Bronze III is a floor.
// One loss at a new-division floor is absorbed by demotion protection.
func ApplyRank(st RankState, won bool) RankState {
	st = st.Sync()
	before := st.Index
	beforeLabel := st.Label()
	if won {
		st.Index += LiveRankWin()
		st.Protect = false
	} else {
		if st.Index == 0 {
			st.RR = 0
			return st.Sync()
		}
		atFloor := st.RR == 0 && st.Index > 0
		if atFloor && st.Protect {
			st.Protect = false
			return st.Sync()
		}
		st.Index -= LiveRankLoss()
		if st.Index < 0 {
			st.Index = 0
		}
		st.Protect = false
	}
	st.Tier, st.Division, st.RR = RankFromIndex(st.Index)
	if won && st.Label() != beforeLabel && st.Index > before {
		st.Protect = true
	}
	return st.Sync()
}

// ApplyRankRR applies a ranked win/loss using frozen RR values.
// It is used to keep active match history stable even if admin changes RR config mid-match.
func ApplyRankRR(st RankState, won bool, rrWin, rrLoss int) RankState {
	st = st.Sync()
	before := st.Index
	beforeLabel := st.Label()

	if won {
		if rrWin > 0 {
			st.Index += rrWin
		}
		st.Protect = false
	} else {
		if st.Index == 0 {
			st.RR = 0
			return st.Sync()
		}
		atFloor := st.RR == 0 && st.Index > 0
		if atFloor && st.Protect {
			st.Protect = false
			return st.Sync()
		}
		if rrLoss > 0 {
			st.Index -= rrLoss
		}
		if st.Index < 0 {
			st.Index = 0
		}
		st.Protect = false
	}

	st.Tier, st.Division, st.RR = RankFromIndex(st.Index)
	if won && st.Label() != beforeLabel && st.Index > before {
		st.Protect = true
	}
	return st.Sync()
}

func SoftResetRank(st RankState) RankState {
	st = st.Sync()
	st.Index = st.Index * 70 / 100
	st.Protect = true
	st.Tier, st.Division, st.RR = RankFromIndex(st.Index)
	st.SeasonID = CurrentSeason.ID
	if o := liveOps(); o != nil {
		st.SeasonID = o.ActiveSeason().ID
	}
	return st.Sync()
}

func SeasonReward(peakIndex int) (title, badge string) {
	_, _, _ = RankFromIndex(peakIndex)
	ti := 0
	idx := peakIndex
	if idx >= 5*3*RRPerDivision+MasterBand {
		return "master", "gm-crest"
	}
	if idx >= 5*3*RRPerDivision {
		return "master", "master-crest"
	}
	ti = idx / (3 * RRPerDivision)
	switch ti {
	case 4:
		return "", "diamond-frame"
	case 3, 2:
		return "", "gold-frame"
	default:
		return "", ""
	}
}

func RankRangeWidth(waitSec int) int {
	// ±1 tier (300 RR) then ±2 tiers after 60s. Never bronze↔master.
	w := 3 * RRPerDivision
	if waitSec >= QueueExpand1 {
		w = 6 * RRPerDivision
	}
	if waitSec >= QueueExpand2 {
		w = 9 * RRPerDivision
	}
	return w
}

func (p PlayerProfile) RankState() RankState {
	if p.RankTier == "" {
		return defaultRank()
	}
	st := RankState{
		Tier: p.RankTier, Division: p.RankDivision, RR: p.RankRR, Protect: p.RankProtect,
		PeakTier: p.PeakTier, PeakDivision: p.PeakDivision, PeakRR: p.PeakRR, PeakIndex: p.PeakIndex,
		SeasonID: p.SeasonID,
	}
	return st.Sync()
}

func applyRankToProfile(p *PlayerProfile, st RankState) {
	st = st.Sync()
	p.RankTier, p.RankDivision, p.RankRR = st.Tier, st.Division, st.RR
	p.RankProtect = st.Protect
	p.PeakTier, p.PeakDivision, p.PeakRR, p.PeakIndex = st.PeakTier, st.PeakDivision, st.PeakRR, st.PeakIndex
	p.SeasonID = st.SeasonID
}

type RankChange struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	MatchID    string `json:"matchId"`
	SeasonID   string `json:"seasonId"`
	Result     string `json:"result"`
	TierBefore string `json:"tierBefore"`
	DivBefore  string `json:"divBefore,omitempty"`
	RRBefore   int    `json:"rrBefore"`
	RRChange   int    `json:"rrChange"`
	RRAfter    int    `json:"rrAfter"`
	TierAfter  string `json:"tierAfter"`
	DivAfter   string `json:"divAfter,omitempty"`
	CreatedAt  int64  `json:"createdAt"`
}

func RanksCompatible(a, b RankState, waitSec int) bool {
	a, b = a.Sync(), b.Sync()
	w := RankRangeWidth(waitSec)
	d := a.Index - b.Index
	if d < 0 {
		d = -d
	}
	if d > w {
		return false
	}
	// Hard cap: bronze (0-2) never matches master (5+)
	ta, tb := tierIndex(a.Tier), tierIndex(b.Tier)
	if (ta <= 0 && tb >= 5) || (tb <= 0 && ta >= 5) {
		return false
	}
	return true
}
