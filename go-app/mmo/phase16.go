package mmo

import (
	"math"
	"strings"
	"time"
)

const pvpAfkWarnText = "Anda tidak aktif. Bergerak atau melakukan aksi untuk tetap berada dalam pertandingan."

type pvpDuelInvite struct {
	From, To string
	Until    time.Time
}

type PvpSeasonHistoryRow struct {
	SeasonID    string `json:"seasonId"`
	Season      string `json:"season"`
	HighestRank string `json:"highestRank"`
	FinalRating int    `json:"finalRating"`
	Wins        int    `json:"wins"`
	Losses      int    `json:"losses"`
}

func pvpReconnectWindow() time.Duration {
	if pvpMod.Reconnect > 0 {
		return time.Duration(pvpMod.Reconnect) * time.Second
	}
	if PvpRejoin > 0 {
		return PvpRejoin
	}
	return 120 * time.Second
}

func pvpIsBattleground(inst *PvpInstance) bool {
	if inst == nil {
		return false
	}
	if inst.Mode == "BATTLEGROUND_5V5" {
		return true
	}
	return pvpIsBattlegroundMap(inst.Map)
}

func pvpIsBattlegroundMap(mapID string) bool {
	switch mapID {
	case "valley-of-dawn", "dawn-arena":
		return true
	default:
		return false
	}
}

func pvpFriendly(inst *PvpInstance) bool {
	if inst == nil {
		return false
	}
	def, ok := pvpMode(inst.Mode)
	return ok && (def.Kind == "DUEL" || inst.Mode == "DUEL_1V1")
}

func pvpAllowSpectate(inst *PvpInstance) bool {
	if inst == nil {
		return false
	}
	def, ok := pvpMode(inst.Mode)
	if !ok {
		return false
	}
	return def.Kind == "DUEL" || def.Kind == "CASUAL"
}

func clampPvpRating(v int) int {
	if v < 0 {
		return 0
	}
	cap := pvpMod.RatingCap
	if cap <= 0 {
		cap = 99999
	}
	if v > cap {
		return cap
	}
	return v
}

func rankDivision(rating int, rk PvpRankDef) string {
	switch rk.ID {
	case "MASTER", "CELESTIAL":
		return ""
	}
	span := rk.Max - rk.Min + 1
	if span < 3 {
		return "I"
	}
	third := span / 3
	off := rating - rk.Min
	if off < third {
		return "III"
	}
	if off < third*2 {
		return "II"
	}
	return "I"
}

func rankVisual(rk PvpRankDef) string {
	if rk.Visual != "" {
		return rk.Visual
	}
	switch rk.ID {
	case "BRONZE":
		return "bronze"
	case "SILVER":
		return "silver"
	case "GOLD":
		return "gold"
	case "PLATINUM", "DIAMOND":
		return "crystal"
	default:
		return "celestial"
	}
}

func rankDisplayName(rating int) string {
	rk := rankForRating(rating)
	div := rankDivision(rating, rk)
	if div == "" {
		return rk.Name
	}
	return rk.Name + " " + div
}

func pvpQueueEstimate(need, have int, waited time.Duration) (int64, string) {
	remain := need - have
	if remain < 1 {
		remain = 1
	}
	base := time.Duration(remain) * 15 * time.Second
	if waited > 45*time.Second {
		base = 20 * time.Second
	}
	return base.Milliseconds(), "perkiraan"
}

func (w *WorldState) pvpNearby(p *Player) []PvpNearbyView {
	out := []PvpNearbyView{}
	if p == nil {
		return out
	}
	for _, o := range w.players {
		if o == nil || o.ID == p.ID || !o.Connected {
			continue
		}
		if o.InstanceID != "" {
			continue
		}
		if hypot2(p.X, p.Z, o.X, o.Z) > 40*40 {
			continue
		}
		cosmetic := o.ensureLog().ActiveCosmetic
		out = append(out, PvpNearbyView{PlayerID: o.ID, Name: o.Name, Level: o.Level, Cosmetic: cosmetic})
		if len(out) >= 12 {
			break
		}
	}
	return out
}

func (w *WorldState) pvpEnterTraining(p *Player) [][]byte {
	if p.InstanceID != "" || w.pvpOf(p.ID) != nil {
		return rejectFor(p.ID, TypePvpTraining, "busy")
	}
	p.X, p.Y, p.Z = 0, 0, 9.6
	p.AX, p.AZ, p.VX, p.VZ = 0, 0, 0, 0
	p.markDirty()
	lobby := w.pvpLobby(p)
	lobby.Training = true
	return [][]byte{
		marshal(TypePvpTraining, map[string]any{"ok": true, "x": p.X, "z": p.Z, "toId": p.ID}),
		marshal(TypePvpLobby, lobby),
		w.notify(p, "system", "Training Arena: uji combo, skill, dodge, transformasi, dan damage. Tanpa ranking."),
	}
}

func (w *WorldState) pvpDuelChallenge(p *Player, targetID string) [][]byte {
	if targetID == "" || targetID == p.ID {
		return rejectFor(p.ID, TypePvpDuel, "target")
	}
	if p.InstanceID != "" || w.pvpOf(p.ID) != nil {
		return rejectFor(p.ID, TypePvpDuel, "busy")
	}
	tgt := w.players[targetID]
	if tgt == nil || !tgt.Connected {
		return rejectFor(p.ID, TypePvpDuel, "target")
	}
	if tgt.InstanceID != "" || w.pvpOf(tgt.ID) != nil {
		return rejectFor(p.ID, TypePvpDuel, "busy")
	}
	if hypot2(p.X, p.Z, tgt.X, tgt.Z) > 48*48 {
		return rejectFor(p.ID, TypePvpDuel, "range")
	}
	w.PvP.ensure()
	if w.PvP.duels == nil {
		w.PvP.duels = map[string]*pvpDuelInvite{}
	}
	w.PvP.duels[tgt.ID] = &pvpDuelInvite{From: p.ID, To: tgt.ID, Until: time.Now().Add(30 * time.Second)}
	return [][]byte{
		marshal(TypePvpDuelRequest, map[string]any{
			"fromId": p.ID, "from": p.Name, "level": p.Level, "toId": tgt.ID,
		}),
		w.notify(p, "system", "Undangan duel dikirim ke "+tgt.Name+"."),
	}
}

func (w *WorldState) pvpDuelRespond(p *Player, accept bool) [][]byte {
	w.PvP.ensure()
	if w.PvP.duels == nil {
		w.PvP.duels = map[string]*pvpDuelInvite{}
	}
	inv := w.PvP.duels[p.ID]
	if inv == nil || time.Now().After(inv.Until) {
		delete(w.PvP.duels, p.ID)
		return rejectFor(p.ID, TypePvpDuelDecline, "expired")
	}
	from := w.players[inv.From]
	delete(w.PvP.duels, p.ID)
	if !accept {
		evs := [][]byte{w.notify(p, "system", "Duel ditolak.")}
		if from != nil {
			evs = append(evs, w.notify(from, "system", p.Name+" menolak duel."))
		}
		return evs
	}
	if from == nil || !from.Connected || from.InstanceID != "" || p.InstanceID != "" {
		return rejectFor(p.ID, TypePvpDuelAccept, "busy")
	}
	def, ok := pvpMode("DUEL_1V1")
	if !ok {
		return rejectFor(p.ID, TypePvpDuelAccept, "mode")
	}
	inst := w.newPvpInstance([]string{from.ID, p.ID}, def, time.Now())
	return w.beginPvp(inst)
}

func (w *WorldState) pvpGetReplay(p *Player, matchID string) [][]byte {
	if matchID == "" {
		if inst := w.pvpOf(p.ID); inst != nil {
			matchID = inst.ID
		}
	}
	if matchID == "" || w.PvP == nil {
		return rejectFor(p.ID, TypeGetReplay, "match")
	}
	events := w.PvP.replays[matchID]
	if events == nil {
		if inst := w.PvP.instances[matchID]; inst != nil {
			events = inst.Replay
		}
	}
	clean := make([]ReplayEvent, 0, len(events))
	for _, ev := range events {
		if ev.Kind == "chat" {
			continue
		}
		clean = append(clean, ev)
	}
	return [][]byte{marshal(TypePvpReplay, map[string]any{"matchId": matchID, "events": clean, "toId": p.ID})}
}

func pvpMVPScore(f *PvpFighter, inst *PvpInstance, p *Player) int {
	if f == nil {
		return 0
	}
	score := f.Objective*25 + f.DmgDealt/40 + f.Assists*12 + f.Kills*4
	if f.Deaths == 0 {
		score += 40
	}
	if p != nil && p.MaxHP > 0 {
		score += int(float64(p.HP) / float64(p.MaxHP) * 20)
	}
	if pvpIsBattleground(inst) {
		score += f.Objective * 10
	}
	return score
}

func (w *WorldState) pvpPickMVP(inst *PvpInstance) (string, int) {
	bestID, best := "", -1
	for _, id := range inst.Players {
		f := inst.Fighters[id]
		pl := w.players[id]
		s := pvpMVPScore(f, inst, pl)
		if s > best {
			best, bestID = s, id
		}
	}
	return bestID, best
}

func (w *WorldState) grantPvpAchievements(p *Player, inst *PvpInstance, f *PvpFighter, won, draw bool) {
	log := p.ensureLog()
	if log.Flags == nil {
		log.Flags = map[string]bool{}
	}
	if won && !draw {
		log.Flags["ach_pvp_first_win"] = true
		p.grantTitle("arena-challenger")
	}
	if log.PvpWins >= 10 {
		log.Flags["ach_pvp_arena_champ"] = true
		p.grantTitle("arena-champion")
	}
	if log.PvpCaptures >= 3 {
		log.Flags["ach_pvp_shrine"] = true
		p.grantTitle("shrine-guardian")
	}
	if pvpFriendly(inst) && won && f != nil && f.Deaths == 0 {
		log.Flags["ach_pvp_perfect_duel"] = true
	}
	if log.PvpRankedMatches >= 20 || (log.PvpPlacementLeft == 0 && log.PvpRankedMatches >= 5) {
		log.Flags["ach_pvp_ranked_vet"] = true
		p.grantTitle("celestial-competitor")
	}
	w.refreshAchievements(p)
}

func (w *WorldState) creditPvpCapture(inst *PvpInstance, s *PvpShrine, team int) {
	for _, id := range inst.Players {
		p := w.players[id]
		f := inst.Fighters[id]
		if p == nil || f == nil || f.Team != team || !p.alive() {
			continue
		}
		if math.Hypot(p.X-s.X, p.Z-s.Z) > s.Radius {
			continue
		}
		f.Objective++
		p.ensureLog().PvpCaptures++
		w.noteActivity(p, "PVP_CAPTURE", s.ID, 1)
	}
}

func (w *WorldState) appendPvpSeasonHistory(log *PlayerLog) {
	if log == nil {
		return
	}
	row := PvpSeasonHistoryRow{
		SeasonID: log.PvpSeasonID, Season: pvpSeasonDef.Name, HighestRank: log.PvpHighestRank,
		FinalRating: log.PvpRating, Wins: log.PvpWins, Losses: log.PvpLosses,
	}
	if row.SeasonID == "" {
		row.SeasonID = pvpSeasonDef.ID
	}
	log.PvpSeasonHistory = append(log.PvpSeasonHistory, row)
	if len(log.PvpSeasonHistory) > 12 {
		log.PvpSeasonHistory = log.PvpSeasonHistory[len(log.PvpSeasonHistory)-12:]
	}
}

func pvpSkillLocked(skillID string) bool {
	id := strings.ToLower(strings.TrimSpace(skillID))
	for _, s := range pvpMod.DisabledSkills {
		if strings.ToLower(s) == id {
			return true
		}
	}
	return false
}

func pvpRespawnDelay(def PvpModeDef) time.Duration {
	if def.RespawnSec > 0 {
		return time.Duration(def.RespawnSec) * time.Second
	}
	if pvpMod.BgRespawn > 0 {
		return time.Duration(pvpMod.BgRespawn) * time.Second
	}
	return 8 * time.Second
}

func pvpViewState(inst *PvpInstance) string {
	if inst == nil {
		return ""
	}
	switch inst.State {
	case PvpMatched, PvpReadySt:
		return "MATCH_FOUND"
	case PvpCompleted, PvpEnding:
		return "FINISHED"
	case PvpCleanup:
		return "CLEANUP"
	case PvpWaiting:
		return "QUEUED"
	default:
		return inst.State
	}
}
