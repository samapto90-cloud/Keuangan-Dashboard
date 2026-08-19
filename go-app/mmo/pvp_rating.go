package mmo

import (
	"sort"
	"strings"
	"time"
)

func (w *WorldState) ensurePvpRating(log *PlayerLog) {
	if log == nil {
		return
	}
	if log.PvpRating == 0 && log.PvpMatches == 0 && log.PvpRankedMatches == 0 {
		log.PvpRating = pvpMod.StartRating
		log.PvpPlacementLeft = pvpMod.Placement
	}
	if log.PvpHistory == nil {
		log.PvpHistory = []PvpHistoryRow{}
	}
	if log.PvpSeasonID == "" {
		log.PvpSeasonID = pvpSeasonDef.ID
	}
}

func (w *WorldState) finishPvp(inst *PvpInstance, winnerTeam int, reason string) [][]byte {
	if inst.Processed || inst.State == PvpCompleted {
		if prev, ok := w.txSeen("pvp:" + inst.ID); ok {
			return prev
		}
		return nil
	}
	inst.State = PvpEnding
	inst.WinnerTeam = winnerTeam
	inst.EndedAt = time.Now()
	inst.Result = reason
	tx := "pvp:" + inst.ID
	if prev, ok := w.txSeen(tx); ok {
		inst.Processed = true
		inst.State = PvpCompleted
		return prev
	}
	mvpID, _ := w.pvpPickMVP(inst)
	var events [][]byte
	for _, id := range inst.Players {
		p := w.players[id]
		if p == nil {
			continue
		}
		f := inst.Fighters[id]
		draw := winnerTeam == 0
		won := !draw && f != nil && f.Team == winnerTeam
		beforeRank := rankForRating(p.ensureLog().PvpRating).ID
		chg := w.applyPvpResult(p, inst, f, won, draw)
		if mvpID == id && !draw {
			xp := pvpMod.MvpSeasonXP
			if xp <= 0 {
				xp = 15
			}
			events = append(events, w.grantSeasonXP(p, xp, "pvp_mvp")...)
		}
		view := w.pvpResultView(inst, p, f, won, draw, chg, beforeRank, mvpID)
		events = append(events, marshal(TypePvpResult, view))
	}
	inst.Processed = true
	inst.State = PvpCompleted
	w.pvpReplay(inst, "end", "", map[string]any{"winner": winnerTeam, "reason": reason})
	w.PvP.replays[inst.ID] = append([]ReplayEvent{}, inst.Replay...)
	w.rebuildLeaderboard()
	for _, id := range inst.Players {
		if p := w.players[id]; p != nil {
			w.extractPvpPlayer(p, inst)
		} else {
			delete(w.PvP.byPlayer, id)
		}
	}
	return w.rememberTx(tx, events)
}

func (w *WorldState) applyPvpResult(p *Player, inst *PvpInstance, f *PvpFighter, won, draw bool) int {
	log := p.ensureLog()
	w.ensurePvpRating(log)
	def, _ := pvpMode(inst.Mode)
	chg := 0
	if def.Kind == "RANKED" && !draw {
		if won {
			chg = pvpMod.WinRating
			log.PvpRating = clampPvpRating(log.PvpRating + chg)
			log.PvpWins++
			log.PvpWinStreak++
			log.PvpLossStreak = 0
			if log.PvpWinStreak >= 3 {
				chg++
				log.PvpRating = clampPvpRating(log.PvpRating + 1)
			}
		} else {
			chg = -pvpMod.LossRating
			log.PvpRating = clampPvpRating(log.PvpRating + chg)
			log.PvpLosses++
			log.PvpLossStreak++
			log.PvpWinStreak = 0
		}
		log.PvpRankedMatches++
		if log.PvpPlacementLeft > 0 {
			log.PvpPlacementLeft--
		}
		rk := rankForRating(log.PvpRating)
		prev := log.PvpHighestRank
		if rankIndex(rk.ID) > rankIndex(prev) {
			log.PvpHighestRank = rk.ID
			w.grantPvpRankReward(p, rk.ID)
		} else if log.PvpHighestRank == "" && log.PvpPlacementLeft == 0 {
			log.PvpHighestRank = rk.ID
			w.grantPvpRankReward(p, rk.ID)
		}
		w.detectAbuse(p.ID, inst, f, won)
	} else if def.Kind == "BATTLEGROUND" {
		w.addCurrency(p, "battle", 8, "bg_"+inst.ID)
		if won && !draw {
			w.addCurrency(p, "battle", 6, "bg_win_"+inst.ID)
			log.PvpWins++
		} else if !draw {
			log.PvpLosses++
		}
		w.addCurrency(p, "coin", 12, "bg_coin_"+inst.ID)
		p.Exp += 8
		w.grantSeasonXP(p, 20, "pvp_bg")
	} else if def.Kind == "DUEL" {
		if won && !draw {
			log.PvpWins++
		} else if !draw {
			log.PvpLosses++
		}
	} else {
		w.addCurrency(p, "coin", 8, "casual_"+inst.ID)
		p.Exp += 6
		if won && !draw {
			log.PvpWins++
		} else if !draw {
			log.PvpLosses++
		}
	}
	log.PvpMatches++
	if f != nil {
		log.PvpDamage += f.DmgDealt
		log.PvpKills += f.Kills
		log.PvpDeaths += f.Deaths
	}
	opp := ""
	if f != nil {
		ids := inst.TeamB
		if f.Team == 2 {
			ids = inst.TeamA
		}
		if len(ids) > 0 {
			if o := w.players[ids[0]]; o != nil {
				opp = o.Name
			} else {
				opp = ids[0]
			}
		}
	}
	res := "DEFEAT"
	if draw {
		res = "DRAW"
	} else if won {
		res = "VICTORY"
	}
	dur := 0
	if !inst.StartedAt.IsZero() {
		end := inst.EndedAt
		if end.IsZero() {
			end = time.Now()
		}
		dur = int(end.Sub(inst.StartedAt).Seconds())
	}
	kills, deaths, assists, dmg, obj := 0, 0, 0, 0, 0
	if f != nil {
		kills, deaths, assists, dmg, obj = f.Kills, f.Deaths, f.Assists, f.DmgDealt, f.Objective
	}
	row := PvpHistoryRow{
		MatchID: inst.ID, Mode: inst.Mode, Opponent: opp, Result: res, RatingChange: chg,
		Date: time.Now().UnixMilli(), PlayerID: p.ID, Duration: dur,
		Kills: kills, Deaths: deaths, Assists: assists, Damage: dmg, Objective: obj,
	}
	log.PvpHistory = append(log.PvpHistory, row)
	if len(log.PvpHistory) > 20 {
		log.PvpHistory = log.PvpHistory[len(log.PvpHistory)-20:]
	}
	w.PvP.history = append(w.PvP.history, row)
	p.markDirty()
	w.noteActivity(p, "PVP_MATCH", inst.Mode, 1)
	if won && !draw {
		w.noteActivity(p, "PVP_WIN", inst.Mode, 1)
		if def.Kind == "RANKED" {
			w.noteActivity(p, "RANKED_WIN", inst.Mode, 1)
		}
	}
	scoreN := obj*20 + dmg/10
	if pvpIsBattleground(inst) {
		if f != nil && f.Team == 1 {
			scoreN += inst.ScoreA
		} else if f != nil {
			scoreN += inst.ScoreB
		}
	}
	if scoreN > 0 {
		w.noteActivity(p, "PVP_SCORE", inst.Mode, scoreN)
	}
	w.grantPvpAchievements(p, inst, f, won, draw)
	w.persist(p)
	return chg
}

func rankIndex(id string) int {
	order := []string{"BRONZE", "SILVER", "GOLD", "PLATINUM", "DIAMOND", "MASTER", "CELESTIAL"}
	for i, v := range order {
		if v == id {
			return i
		}
	}
	return -1
}

func (w *WorldState) grantPvpRankReward(p *Player, rankID string) {
	log := p.ensureLog()
	if log.PvpClaimed == nil {
		log.PvpClaimed = map[string]bool{}
	}
	key := pvpSeasonDef.ID + ":" + rankID
	if log.PvpClaimed[key] {
		return
	}
	for _, rw := range pvpRewardList {
		if rw.Rank != rankID {
			continue
		}
		if rw.Kind == "title" {
			p.grantTitle(rw.ItemID)
		} else {
			p.grantCosmetic(rw.ItemID)
		}
	}
	log.PvpClaimed[key] = true
}

func (w *WorldState) detectAbuse(id string, inst *PvpInstance, f *PvpFighter, won bool) {
	ab := w.pvpAbuse(id)
	opp := ""
	if f != nil {
		ids := inst.TeamB
		if f.Team == 2 {
			ids = inst.TeamA
		}
		if len(ids) == 1 {
			opp = ids[0]
		}
	}
	if opp != "" && won {
		if ab.LastOpp == opp {
			ab.PairWins[opp]++
		} else {
			ab.PairWins[opp] = 1
		}
		ab.LastOpp = opp
		if ab.PairWins[opp] >= 5 {
			ab.Suspicious += 20
		}
	}
	if f != nil && f.Deaths >= 3 && f.DmgDealt < 10 {
		ab.FeedFlags++
		ab.Suspicious += 8
	}
}

func (w *WorldState) pvpResultView(inst *PvpInstance, p *Player, f *PvpFighter, won, draw bool, chg int, beforeRank, mvpID string) PvpResultOut {
	log := p.ensureLog()
	rk := rankForRating(log.PvpRating)
	out := PvpResultOut{
		MatchID: inst.ID, Mode: inst.Mode, Victory: won && !draw, Draw: draw, RatingChange: chg,
		Rating: log.PvpRating, Rank: rk.ID, RankName: rankDisplayName(log.PvpRating),
		Promoted:    rankIndex(rk.ID) > rankIndex(beforeRank),
		Demoted:     rankIndex(rk.ID) < rankIndex(beforeRank) && rankIndex(rk.ID) >= 0,
		BattleToken: log.BattleToken, InstanceID: inst.ID, ToID: p.ID, Mvp: mvpID == p.ID,
	}
	if draw {
		out.Title = "DRAW"
	} else if won {
		out.Title = "VICTORY"
	} else {
		out.Title = "DEFEAT"
	}
	if f != nil {
		out.Kills, out.Deaths, out.Assists, out.Damage, out.Objective = f.Kills, f.Deaths, f.Assists, f.DmgDealt, f.Objective
	}
	if !inst.StartedAt.IsZero() && !inst.EndedAt.IsZero() {
		out.Duration = int(inst.EndedAt.Sub(inst.StartedAt).Seconds())
	}
	if mvpID != "" {
		if m := w.players[mvpID]; m != nil {
			out.MvpName = m.Name
		} else if mf := inst.Fighters[mvpID]; mf != nil {
			out.MvpName = mf.Name
		}
	}
	if inst.Ranked && log.PvpPlacementLeft == 0 {
		out.RankName = rankDisplayName(log.PvpRating)
	}
	return out
}

func (w *WorldState) pvpView(inst *PvpInstance, watcher string) PvpView {
	if inst == nil {
		return PvpView{}
	}
	left := 0
	if !inst.EndsAt.IsZero() && time.Now().Before(inst.EndsAt) {
		left = int(time.Until(inst.EndsAt).Seconds())
	}
	members := make([]PvpMemberView, 0, len(inst.Players))
	for _, id := range inst.Players {
		pl := w.players[id]
		f := inst.Fighters[id]
		mv := PvpMemberView{PlayerID: id}
		if f != nil {
			mv.Name, mv.Team = f.Name, f.Team
			mv.Kills, mv.Deaths, mv.Assists, mv.Damage = f.Kills, f.Deaths, f.Assists, f.DmgDealt
			mv.SpectateID = f.SpectateID
		}
		if pl != nil {
			mv.Name = pl.Name
			mv.HP, mv.MaxHP = pl.HP, pl.MaxHP
			mv.Alive = pl.alive()
			mv.PingMs = pl.PingMs
		}
		members = append(members, mv)
	}
	pts := make([]PvpPointView, 0, len(inst.Shrines))
	for _, s := range inst.Shrines {
		pts = append(pts, PvpPointView{ID: s.ID, Owner: s.Owner, Contested: s.Contested, ProgressA: int(s.ProgressA + 0.5), ProgressB: int(s.ProgressB + 0.5), X: s.X, Z: s.Z})
	}
	self := inst.Fighters[watcher]
	team := 0
	if self != nil {
		team = self.Team
	}
	return PvpView{
		MatchID: inst.ID, Mode: inst.Mode, Map: inst.Map, State: pvpViewState(inst), TimeLeft: left,
		ScoreA: inst.ScoreA, ScoreB: inst.ScoreB, Members: members, Points: pts, KillFeed: append([]string{}, inst.KillFeed...),
		Team: team, InstanceID: inst.ID, Countdown: inst.CountdownUntil.UnixMilli(),
	}
}

func (w *WorldState) pvpHistory(p *Player) []PvpHistoryRow {
	log := p.ensureLog()
	out := append([]PvpHistoryRow{}, log.PvpHistory...)
	for i := range out {
		out[i].PlayerID = ""
	}
	return out
}

func (w *WorldState) pvpRewardViews(log *PlayerLog) []PvpRewardView {
	out := make([]PvpRewardView, 0, len(pvpRewardList))
	for _, rw := range pvpRewardList {
		unlocked := log.PvpClaimed[pvpSeasonDef.ID+":"+rw.Rank] || rankIndex(log.PvpHighestRank) >= rankIndex(rw.Rank)
		out = append(out, PvpRewardView{ID: rw.ID, Rank: rw.Rank, Kind: rw.Kind, Name: rw.Name, Unlocked: unlocked})
	}
	return out
}

func (w *WorldState) pvpShopBuy(p *Player, shopItemID, tx string) [][]byte {
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	var def PvpShopItem
	found := false
	for _, it := range pvpShop.Items {
		if it.ShopItemID == shopItemID {
			def, found = it, true
			break
		}
	}
	if !found {
		return rejectFor(p.ID, TypePvpShopBuy, "item")
	}
	kind := strings.ToLower(def.Kind)
	if kind == "weapon" || kind == "damage" || kind == "hp" || kind == "power" || kind == "boost" {
		return rejectFor(p.ID, TypePvpShopBuy, "pay_to_win")
	}
	if def.Currency != "battle" {
		return rejectFor(p.ID, TypePvpShopBuy, "currency")
	}
	if !w.removeCurrency(p, "battle", def.Price, "pvp_shop") {
		return rejectFor(p.ID, TypePvpShopBuy, "currency")
	}
	switch kind {
	case "title":
		p.grantTitle(def.ItemID)
	case "cosmetic", "emote":
		p.grantCosmetic(def.ItemID)
	default:
		w.addCurrency(p, "battle", def.Price, "pvp_shop_rollback")
		return rejectFor(p.ID, TypePvpShopBuy, "kind")
	}
	p.markDirty()
	w.persist(p)
	out := [][]byte{marshal(TypePvpLobby, w.pvpLobby(p)), w.notify(p, "system", "Membeli "+def.Name)}
	return w.rememberTx(tx, out)
}

func (w *WorldState) pvpLeaderboard(p *Player, board string) map[string]any {
	w.rebuildLeaderboardIfStale()
	board = strings.ToUpper(board)
	list := w.PvP.lb.Global
	switch board {
	case "REGIONAL", "REGION":
		region := p.Region
		if region == "" {
			region = "ASIA"
		}
		list = w.PvP.lb.Regional[region]
	case "FRIENDS":
		list = w.filterFriendsLB(p)
	case "GUILD":
		list = w.filterGuildLB(p)
	}
	if list == nil {
		list = []LBEntry{}
	}
	return map[string]any{"board": board, "entries": list, "toId": p.ID, "cachedAt": w.PvP.lb.BuiltAt.UnixMilli()}
}

func (w *WorldState) filterFriendsLB(p *Player) []LBEntry {
	out := []LBEntry{}
	if w.Social == nil {
		return out
	}
	friends := w.Social.Friends[p.ID]
	for _, e := range w.PvP.lb.Global {
		if e.PlayerID == p.ID {
			out = append(out, e)
			continue
		}
		if friends != nil {
			if _, ok := friends[e.PlayerID]; ok {
				out = append(out, e)
			}
		}
	}
	return out
}

func (w *WorldState) filterGuildLB(p *Player) []LBEntry {
	out := []LBEntry{}
	gid := p.GuildID
	if gid == "" {
		return out
	}
	for _, e := range w.PvP.lb.Global {
		pl := w.players[e.PlayerID]
		if pl != nil && pl.GuildID == gid {
			out = append(out, e)
		}
	}
	return out
}

func (w *WorldState) rebuildLeaderboardIfStale() {
	if time.Since(w.PvP.lb.BuiltAt) < 12*time.Second && len(w.PvP.lb.Global) > 0 {
		return
	}
	w.rebuildLeaderboard()
}

func (w *WorldState) rebuildLeaderboard() {
	type row struct {
		id, name, region, guild string
		rating, wins, losses    int
	}
	seen := map[string]*row{}
	for _, p := range w.players {
		log := p.ensureLog()
		w.ensurePvpRating(log)
		region := p.Region
		if region == "" {
			region = "ASIA"
		}
		seen[p.ID] = &row{id: p.ID, name: p.Name, region: region, guild: p.GuildTag, rating: log.PvpRating, wins: log.PvpWins, losses: log.PvpLosses}
	}
	if repo, ok := w.QuestRepo.(*MemoryQuestRepo); ok {
		repo.mu.Lock()
		for id, log := range repo.logs {
			if seen[id] != nil || log == nil {
				continue
			}
			seen[id] = &row{id: id, name: id, region: "ASIA", rating: log.PvpRating, wins: log.PvpWins, losses: log.PvpLosses}
		}
		repo.mu.Unlock()
	}
	rows := make([]row, 0, len(seen))
	for _, r := range seen {
		if r.rating == 0 && r.wins == 0 && r.losses == 0 {
			continue
		}
		rows = append(rows, *r)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].rating == rows[j].rating {
			return rows[i].wins > rows[j].wins
		}
		return rows[i].rating > rows[j].rating
	})
	global := make([]LBEntry, 0, len(rows))
	regional := map[string][]LBEntry{}
	for i, r := range rows {
		rk := rankForRating(r.rating)
		e := LBEntry{Place: i + 1, PlayerID: r.id, Name: r.name, RankName: rk.Name, Rating: r.rating, Wins: r.wins, Losses: r.losses, WinRate: winRate(r.wins, r.losses), Region: r.region, Guild: r.guild}
		global = append(global, e)
		regional[r.region] = append(regional[r.region], e)
	}
	for region, list := range regional {
		for i := range list {
			list[i].Place = i + 1
		}
		regional[region] = list
	}
	w.PvP.lb = LeaderboardCache{Global: global, Regional: regional, BuiltAt: time.Now()}
}

func (w *WorldState) seasonSoftReset() {
	if repo, ok := w.QuestRepo.(*MemoryQuestRepo); ok {
		repo.mu.Lock()
		for _, log := range repo.logs {
			if log == nil {
				continue
			}
			base := pvpMod.StartRating
			if pvpSeasonDef.Floor > 0 {
				base = pvpSeasonDef.Floor
			}
			factor := pvpSeasonDef.SoftReset
			if factor <= 0 {
				factor = 0.5
			}
			w.appendPvpSeasonHistory(log)
			log.PvpRating = clampPvpRating(base + int(float64(log.PvpRating-base)*factor))
			log.PvpSeasonID = pvpSeasonDef.ID
			log.PvpPlacementLeft = pvpMod.Placement
		}
		repo.mu.Unlock()
	}
}
