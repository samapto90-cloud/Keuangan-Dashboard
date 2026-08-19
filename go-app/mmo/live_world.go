package mmo

import (
	"time"
)

func (w *WorldState) ensureCommunity() *CommunityLive {
	cfg := liveWorldCfg.Community
	if w.Community == nil {
		w.Community = &CommunityLive{
			ID: cfg.ID, Name: cfg.Name, Target: cfg.Target, Cosmetic: cfg.Cosmetic,
			DailyCap: cfg.DailyCap, Granted: map[string]bool{}, Participants: map[string]int{},
		}
		if w.Community.DailyCap < 1 {
			w.Community.DailyCap = 500
		}
		if w.Community.Target < 1 {
			w.Community.Target = 1000000
		}
	}
	c := w.Community
	if c.Points >= c.Target {
		c.State = "COMPLETED"
	} else {
		c.State = communityState(time.Now().UTC())
	}
	return c
}

func communityState(now time.Time) string {
	cfg := liveWorldCfg.Community
	start, err1 := time.Parse(time.RFC3339, cfg.Start)
	end, err2 := time.Parse(time.RFC3339, cfg.End)
	if err1 != nil || err2 != nil {
		return "ACTIVE"
	}
	switch {
	case now.Before(start):
		return "SCHEDULED"
	case now.After(end):
		return "EXPIRED"
	default:
		return "ACTIVE"
	}
}

func (w *WorldState) addCommunityPoints(p *Player, amount int) [][]byte {
	if p == nil || amount <= 0 {
		return nil
	}
	c := w.ensureCommunity()
	if c.State != "ACTIVE" && c.State != "COMPLETED" {
		return nil
	}
	if c.State != "ACTIVE" {
		return w.maybeGrantFestival(p)
	}
	log := p.ensureLog()
	day := utcDayKey()
	if log.LiveDay != day {
		log.LiveDay = day
		log.LiveDayAmt = 0
	}
	cap := c.DailyCap
	remain := cap - log.LiveDayAmt
	if remain <= 0 {
		return nil
	}
	if amount > remain {
		amount = remain
	}
	log.LiveDayAmt += amount
	if log.LiveContrib == nil {
		log.LiveContrib = map[string]int{}
	}
	log.LiveContrib[c.ID] += amount
	c.Participants[p.ID] += amount
	c.Points += amount
	if c.Points >= c.Target {
		c.State = "COMPLETED"
		w.notify(p, "event", "Light Festival complete")
		return w.grantFestivalToAll()
	}
	return nil
}

func (w *WorldState) maybeGrantFestival(p *Player) [][]byte {
	c := w.ensureCommunity()
	if c.State != "COMPLETED" && c.Points < c.Target {
		return nil
	}
	if c.Granted[p.ID] {
		return nil
	}
	if c.Participants[p.ID] < 1 {
		return nil
	}
	c.Granted[p.ID] = true
	p.grantCosmetic(c.Cosmetic)
	w.refreshCollectionMilestones(p)
	w.audit("eventCompleted", p.ID, c.ID)
	return [][]byte{marshal(TypeEventReward, map[string]any{"eventId": c.ID, "cosmetic": c.Cosmetic, "toId": p.ID})}
}

func (w *WorldState) grantFestivalToAll() [][]byte {
	c := w.ensureCommunity()
	var out [][]byte
	for id := range c.Participants {
		p := w.players[id]
		if p == nil {
			continue
		}
		out = append(out, w.maybeGrantFestival(p)...)
	}
	return out
}

func (w *WorldState) communityView() map[string]any {
	c := w.ensureCommunity()
	return map[string]any{
		"id": c.ID, "name": c.Name, "points": c.Points, "target": c.Target,
		"state": c.State, "dailyCap": c.DailyCap,
	}
}

func (w *WorldState) eventCalendar() map[string]any {
	now := time.Now().UTC()
	wd := int(now.Weekday())
	todayBoss := ""
	todayName := ""
	week := make([]map[string]any, 0)
	for _, r := range liveWorldCfg.WorldBossRotation {
		row := map[string]any{"weekday": r.Weekday, "id": r.ID, "name": r.Name, "region": r.Region, "kind": "WORLD_BOSS"}
		week = append(week, row)
		if r.Weekday == wd {
			todayBoss, todayName = r.ID, r.Name
		}
	}
	evIdx := (int(now.YearDay()) / 1) % len(liveWorldCfg.EventRotation)
	if evIdx < 0 {
		evIdx = 0
	}
	todayEvent := ""
	if len(liveWorldCfg.EventRotation) > 0 {
		todayEvent = liveWorldCfg.EventRotation[evIdx]
	}
	upcoming := make([]map[string]any, 0)
	for i := 1; i <= 3; i++ {
		d := now.AddDate(0, 0, i)
		day := int(d.Weekday())
		for _, r := range liveWorldCfg.WorldBossRotation {
			if r.Weekday == day {
				upcoming = append(upcoming, map[string]any{"when": d.Format("2006-01-02"), "id": r.ID, "name": r.Name, "kind": "WORLD_BOSS"})
			}
		}
	}
	return map[string]any{
		"timezone": liveWorldCfg.Timezone, "serverTime": now.Format(time.RFC3339),
		"today": map[string]any{"worldBoss": todayBoss, "worldBossName": todayName, "event": todayEvent},
		"week": week, "upcoming": upcoming,
	}
}

func (w *WorldState) rotatedWorldBoss(base WorldBossDef) WorldBossDef {
	wd := int(time.Now().UTC().Weekday())
	for _, r := range liveWorldCfg.WorldBossRotation {
		if r.Weekday != wd {
			continue
		}
		base.ID = r.ID
		base.Name = r.Name
		base.Announce = r.Announce
		if r.Region != "" {
			base.Region = r.Region
		}
		if r.ID == "celestial-avaron-echo" {
			base.Name = "Celestial Avaron Echo"
			base.Title = "Echo of the Last Guardian"
			base.Z = 150
			base.Region = "celestial"
		}
		return base
	}
	return base
}

func worldBossChest(score int) string {
	switch {
	case score >= 800:
		return "LEGENDARY"
	case score >= 400:
		return "EPIC"
	case score >= 120:
		return "RARE"
	default:
		return "COMMON"
	}
}

func (w *WorldState) refreshCollectionMilestones(p *Player) {
	if p == nil {
		return
	}
	log := p.ensureLog()
	n := len(log.Cosmetics)
	if n >= 10 {
		p.grantTitle("world-adventurer")
	}
	if n >= 25 {
		p.grantCosmetic("badge-dawn-33")
	}
	if n >= 50 {
		p.grantCosmetic("aura-horizon-free")
	}
}

func (w *WorldState) noteGuildDungeon(inst *DungeonInstance) {
	if inst == nil || w.Guilds == nil {
		return
	}
	var gid string
	for _, id := range inst.Players {
		g := w.guildOf(id)
		if g == nil {
			return
		}
		if gid == "" {
			gid = g.ID
		} else if g.ID != gid {
			return
		}
	}
	g := w.Guilds.ByID[gid]
	if g == nil {
		return
	}
	week := raidWeekKey()
	if g.WeeklyWeek != week {
		g.WeeklyWeek = week
		g.WeeklyDungeon = 0
		g.WeeklyRewarded = false
	}
	g.WeeklyDungeon++
	for _, id := range inst.Players {
		if p := w.players[id]; p != nil {
			w.noteActivity(p, "GUILD_DUNGEON", inst.DefID, 1)
		}
	}
	if g.WeeklyDungeon >= 20 && !g.WeeklyRewarded {
		g.WeeklyRewarded = true
		g.Exp += 80
		for _, id := range g.memberIDs() {
			if p := w.players[id]; p != nil {
				p.grantTitle("guild-banner")
			}
		}
		w.audit("guildChallenge", gid, week)
	}
}

func (w *WorldState) recordEducation(p *Player, category string, correct bool) [][]byte {
	log := p.ensureLog()
	log.EduAnswered++
	day := utcDayKey()
	if correct {
		log.EduCorrect++
		if log.EduCats == nil {
			log.EduCats = map[string]int{}
		}
		if category == "" {
			category = "Umum"
		}
		log.EduCats[category]++
		if log.EduLastDay == day {
			// same day, streak unchanged
		} else if log.EduLastDay == "" {
			log.EduStreak = 1
		} else {
			prev, err := time.Parse("2006-01-02", log.EduLastDay)
			if err == nil && utcDayKeyAt(prev.Add(24*time.Hour)) == day {
				log.EduStreak++
			} else {
				log.EduStreak = 1
			}
		}
		log.EduLastDay = day
		w.noteActivity(p, "ANSWER", category, 1)
	}
	return w.refreshAchievements(p)
}

func utcDayKeyAt(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}
