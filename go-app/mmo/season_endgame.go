package mmo

import (
	"strconv"
	"time"
)

const seasonDailyXPCap = 300

func (w *WorldState) grantSeasonXP(p *Player, amount int, reason string) [][]byte {
	if p == nil || amount <= 0 {
		return nil
	}
	log := p.ensureLog()
	day := utcDayKey()
	if log.SeasonXPDay != day {
		log.SeasonXPDay = day
		log.SeasonXPToday = 0
	}
	remain := seasonDailyXPCap - log.SeasonXPToday
	if remain <= 0 {
		return nil
	}
	if amount > remain {
		amount = remain
	}
	if log.SeasonLevel < 1 {
		log.SeasonLevel = 1
	}
	if log.SeasonTrackID == "" {
		log.SeasonTrackID = seasonTrack.ID
	}
	need := seasonTrack.XPPerLevel
	if need < 1 {
		need = 100
	}
	max := seasonTrack.MaxLevel
	if max < 1 {
		max = 50
	}
	log.SeasonXP += amount
	log.SeasonXPToday += amount
	for log.SeasonLevel < max {
		into := log.SeasonXP - (log.SeasonLevel-1)*need
		if into < need {
			break
		}
		log.SeasonLevel++
		w.notify(p, "season", "Season Level "+strconv.Itoa(log.SeasonLevel))
	}
	if log.SeasonLevel > max {
		log.SeasonLevel = max
	}
	p.markDirty()
	w.audit("seasonXP", p.ID, reason)
	return nil
}

func (w *WorldState) claimSeasonLevel(p *Player, level int, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	log := p.ensureLog()
	if level < 1 || level > log.SeasonLevel {
		return rejectFor(p.ID, TypeClaimSeason, "level")
	}
	key := strconv.Itoa(level)
	if log.SeasonClaimed[key] {
		return rejectFor(p.ID, TypeClaimSeason, "claimed")
	}
	var def SeasonRewardDef
	ok := false
	for _, r := range seasonTrack.Rewards {
		if r.Level == level {
			def, ok = r, true
			break
		}
	}
	if !ok {
		return rejectFor(p.ID, TypeClaimSeason, "reward")
	}
	if combatSeasonReward(def.Kind) {
		return rejectFor(p.ID, TypeClaimSeason, "pay_to_win")
	}
	log.SeasonClaimed[key] = true
	switch def.Kind {
	case "title":
		p.grantTitle(def.ID)
	case "coin":
		w.giveCurrency(p, 25, 0)
	case "material":
		log.HorizonFragments++
	default:
		p.grantCosmetic(def.ID)
	}
	w.refreshCollectionMilestones(p)
	p.markDirty()
	w.persist(p)
	w.audit("seasonCompleted", p.ID, key)
	return w.rememberTx(tx, [][]byte{marshal(TypeEndgameState, w.endgameView(p))})
}

func combatSeasonReward(kind string) bool {
	switch kind {
	case "damage", "hp", "multiplier", "attack", "power":
		return true
	}
	return false
}

func (w *WorldState) maybeRotateSeason(p *Player, now time.Time) {
	end, err := time.Parse(time.RFC3339, seasonTrack.End)
	if err != nil {
		return
	}
	log := p.ensureLog()
	flag := "season_" + seasonTrack.ID + "_archived"
	if !now.After(end) {
		return
	}
	if log.Flags[flag] {
		return
	}
	w.forceSeasonEnd(p, now)
}

func (w *WorldState) forceSeasonEnd(p *Player, at time.Time) {
	log := p.ensureLog()
	n := 0
	for _, ok := range log.Guardians {
		if ok {
			n++
		}
	}
	lvl := log.SeasonLevel
	if lvl < 1 {
		lvl = 1
	}
	log.SeasonHistory = append(log.SeasonHistory, SeasonHistoryRow{
		SeasonID:  seasonTrack.ID,
		Name:      seasonTrack.Name,
		Level:     lvl,
		XP:        log.SeasonXP,
		Guardians: n,
		Rank:      log.PvpHighestRank,
		Ended:     at.UnixMilli(),
	})
	log.Flags["season_"+seasonTrack.ID+"_archived"] = true
	log.SeasonXP = 0
	log.SeasonLevel = 1
	log.SeasonClaimed = map[string]bool{}
	log.SeasonXPToday = 0
	log.SeasonTrackID = seasonTrack.ID
	p.markDirty()
	w.audit("seasonCompleted", p.ID, seasonTrack.ID)
	w.notify(p, "season", "Season ended. History saved.")
}
