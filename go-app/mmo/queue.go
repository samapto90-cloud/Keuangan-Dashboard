package mmo

import (
	"strings"
	"time"
)

func dungeonKind(def DungeonDef) string {
	if def.Kind != "" {
		return strings.ToUpper(def.Kind)
	}
	if def.ChapterID != "" {
		return "CHAPTER"
	}
	return "DUNGEON"
}

func dungeonDifficulty(def DungeonDef) string {
	if def.Difficulty != "" {
		return strings.ToUpper(def.Difficulty)
	}
	return "NORMAL"
}

func dungeonMinPlayers(def DungeonDef) int {
	if def.MinPlayers > 0 {
		return def.MinPlayers
	}
	return 1
}

func raidWeekKey() string {
	y, w := time.Now().UTC().ISOWeek()
	return itoa(y) + "-W" + pad2(w)
}

func utcDayKey() string {
	return time.Now().UTC().Format("2006-01-02")
}

func (w *WorldState) raidLocked(p *Player, def DungeonDef) bool {
	if p == nil || !def.WeeklyLockout {
		return false
	}
	log := p.ensureLog()
	if log.RaidLockout == nil {
		return false
	}
	return log.RaidLockout[def.ID] == raidWeekKey()
}

func (w *WorldState) dungeonList(p *Player) DungeonListOut {
	out := DungeonListOut{}
	for _, def := range dungeonCatalog {
		kind := dungeonKind(def)
		if kind != "DUNGEON" && kind != "RAID" && def.ID != "dun-ch01" {
			continue
		}
		st := "AVAILABLE"
		if def.ChapterID != "" {
			st = chapterStatus(p, chapterByID[def.ChapterID])
		}
		if w.raidLocked(p, def) {
			st = "LOCKOUT"
		}
		if p.Level < def.MinimumLevel {
			st = "LEVEL"
		}
		resetAt := int64(0)
		resetLabel := ""
		if def.WeeklyLockout {
			resetAt = raidResetAt(time.Now()).UnixMilli()
			resetLabel = raidLockoutLabel(time.Now())
		}
		diffs := []string{"NORMAL", "HARD"}
		if kind == "RAID" {
			diffs = []string{"NORMAL"}
		}
		out.Dungeons = append(out.Dungeons, DungeonOffer{
			DungeonID: def.ID, Name: def.Name, ChapterID: def.ChapterID, Kind: kind, Description: def.Description,
			Difficulty: dungeonDifficulty(def), RecommendedLevel: def.RecommendedLevel,
			MinPlayers: dungeonMinPlayers(def), MaxPlayers: def.MaxPlayers,
			TimeLimit: def.TimeLimit, Rewards: rewardView(def.Rewards, false), Status: st,
			Region: def.Region, Difficulties: diffs, LockoutResetAt: resetAt, LockoutLabel: resetLabel,
		})
	}
	out.Dungeons = append([]DungeonOffer{{
		DungeonID: "RANDOM", Name: "Random Dungeon", Kind: "DUNGEON", Description: "Server memilih dungeon sesuai level dan unlock.",
		Difficulty: "NORMAL", RecommendedLevel: p.Level, MinPlayers: 1, MaxPlayers: 5, TimeLimit: 1800,
		Status: "AVAILABLE", Difficulties: []string{"NORMAL", "HARD"},
	}}, out.Dungeons...)
	out.History = w.dungeonHistoryFor(p)
	out.Board = w.dungeonBoardFor("")
	out.RaidShop = raidShopCatalog
	out.RaidTokens = p.ensureLog().RaidTokens
	out.LockoutResetAt = raidResetAt(time.Now()).UnixMilli()
	out.LockoutLabel = raidLockoutLabel(time.Now())
	if q := w.queueOf(p.ID); q != nil {
		def := dungeonByID[q.DungeonID]
		out.Queue = &QueueView{
			State: q.State, DungeonID: q.DungeonID, Name: def.Name, Role: q.Role,
			Players: w.queueCount(q.DungeonID, q.Difficulty), MinPlayers: dungeonMinPlayers(def),
			MaxPlayers: def.MaxPlayers, WaitMs: time.Since(q.JoinedAt).Milliseconds(), ToID: p.ID,
		}
	}
	return out
}

func (w *WorldState) queueOf(playerID string) *QueueEntry {
	if w.Dungeons == nil || w.Dungeons.Queue == nil {
		return nil
	}
	for i := range w.Dungeons.Queue.Entries {
		if w.Dungeons.Queue.Entries[i].PlayerID == playerID {
			return &w.Dungeons.Queue.Entries[i]
		}
	}
	return nil
}

func (w *WorldState) queueCount(dungeonID, diff string) int {
	n := 0
	for _, e := range w.Dungeons.Queue.Entries {
		if e.DungeonID == dungeonID && (diff == "" || e.Difficulty == diff) {
			n++
		}
	}
	return n
}

func (w *WorldState) removeQueue(playerID string) {
	if w.Dungeons == nil || w.Dungeons.Queue == nil {
		return
	}
	live := w.Dungeons.Queue.Entries[:0]
	for _, e := range w.Dungeons.Queue.Entries {
		if e.PlayerID != playerID {
			live = append(live, e)
		}
	}
	w.Dungeons.Queue.Entries = live
}

func normalizeRole(role string) string {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case "TANK":
		return "TANK"
	case "SUPPORT", "HEALER":
		return "SUPPORT"
	case "DPS":
		return "DPS"
	default:
		return "FLEX"
	}
}

func (w *WorldState) queueJoin(p *Player, in DungeonActionIn) [][]byte {
	if p.InstanceID != "" {
		return rejectFor(p.ID, TypeQueueJoin, "already")
	}
	if w.queuePenaltyActive(p.ID) {
		return rejectFor(p.ID, TypeQueueJoin, "penalty")
	}
	diff := normalizeDungeonDiff(in.Difficulty)
	if in.Difficulty != "" && !prototypeDiffOK(diff) {
		return rejectFor(p.ID, TypeQueueJoin, "difficulty")
	}
	if diff == "" {
		diff = "NORMAL"
	}
	dungeonID := in.DungeonID
	if strings.EqualFold(dungeonID, "RANDOM") {
		picked, ok := w.pickRandomDungeon(p, diff)
		if !ok {
			return rejectFor(p.ID, TypeQueueJoin, "dungeon")
		}
		dungeonID = picked.ID
	}
	def, ok := dungeonByID[dungeonID]
	if !ok {
		return rejectFor(p.ID, TypeQueueJoin, "dungeon")
	}
	kind := dungeonKind(def)
	if kind != "DUNGEON" && kind != "RAID" {
		return rejectFor(p.ID, TypeQueueJoin, "queue")
	}
	if p.Level < def.MinimumLevel {
		return rejectFor(p.ID, TypeQueueJoin, "level")
	}
	if w.raidLocked(p, def) {
		return rejectFor(p.ID, TypeQueueJoin, "lockout")
	}
	if diff == "" {
		diff = dungeonDifficulty(def)
	}
	role := normalizeRole(in.Role)
	region := in.Region
	if region == "" {
		region = "dawn"
	}
	w.removeQueue(p.ID)
	ids := []string{p.ID}
	partyID := p.PartyID
	if pt := w.Parties.Of(p.ID); pt != nil {
		if pt.LeaderID != p.ID {
			return rejectFor(p.ID, TypeQueueJoin, "not_leader")
		}
		ids = append([]string{}, pt.Members...)
		partyID = pt.ID
		if len(ids) > def.MaxPlayers {
			return rejectFor(p.ID, TypeQueueJoin, "full")
		}
	}
	now := time.Now()
	for _, id := range ids {
		m := w.players[id]
		if m == nil || !m.Connected {
			return rejectFor(p.ID, TypeQueueJoin, "offline")
		}
		if m.InstanceID != "" {
			return rejectFor(p.ID, TypeQueueJoin, "busy")
		}
		if m.Level < def.MinimumLevel {
			return rejectFor(p.ID, TypeQueueJoin, "level")
		}
		if w.raidLocked(m, def) {
			return rejectFor(p.ID, TypeQueueJoin, "lockout")
		}
		w.removeQueue(id)
		r := role
		if pt := w.Parties.Of(id); pt != nil && pt.Roles[id] != "" {
			r = normalizeRole(pt.Roles[id])
		}
		w.Dungeons.Queue.Entries = append(w.Dungeons.Queue.Entries, QueueEntry{
			PlayerID: id, Role: r, DungeonID: def.ID, Difficulty: diff, Region: region,
			PartyID: partyID, JoinedAt: now, State: QueueQueued,
		})
		w.noteQueueJoin(id)
	}
	out := w.tickQueue(now)
	for _, id := range ids {
		out = append(out, marshal(TypeQueueUpdate, w.queueViewFor(id)))
	}
	return out
}

func (w *WorldState) queueLeave(p *Player) [][]byte {
	if w.queueOf(p.ID) == nil {
		return rejectFor(p.ID, TypeQueueLeave, "queue")
	}
	w.noteQueueLeave(p.ID)
	w.removeQueue(p.ID)
	return [][]byte{marshal(TypeQueueUpdate, QueueView{State: QueueCancel, ToID: p.ID})}
}

func (w *WorldState) queueViewFor(id string) QueueView {
	q := w.queueOf(id)
	if q == nil {
		return QueueView{State: QueueIdle, ToID: id}
	}
	def := dungeonByID[q.DungeonID]
	return QueueView{
		State: q.State, DungeonID: q.DungeonID, Name: def.Name, Role: q.Role,
		Players: w.queueCount(q.DungeonID, q.Difficulty), MinPlayers: dungeonMinPlayers(def),
		MaxPlayers: def.MaxPlayers, WaitMs: time.Since(q.JoinedAt).Milliseconds(), ToID: id,
	}
}

func (w *WorldState) tickQueue(now time.Time) [][]byte {
	if w.Dungeons == nil || w.Dungeons.Queue == nil {
		return nil
	}
	groups := map[string][]QueueEntry{}
	for _, e := range w.Dungeons.Queue.Entries {
		if e.State != QueueQueued {
			continue
		}
		key := e.DungeonID + "|" + e.Difficulty + "|" + e.Region
		groups[key] = append(groups[key], e)
	}
	var out [][]byte
	for _, list := range groups {
		if len(list) == 0 {
			continue
		}
		def := dungeonByID[list[0].DungeonID]
		minN := dungeonMinPlayers(def)
		maxN := def.MaxPlayers
		if maxN < 1 {
			maxN = 4
		}
		picked := pickQueueMatch(list, minN, maxN)
		if len(picked) < minN {
			continue
		}
		oldest := picked[0].JoinedAt
		for _, e := range picked {
			if e.JoinedAt.Before(oldest) {
				oldest = e.JoinedAt
			}
		}
		waited := !now.Before(oldest.Add(QueueFillWait))
		if len(picked) < maxN && !waited {
			continue
		}
		members := make([]string, 0, len(picked))
		ready := &ReadyCheck{
			DungeonID: def.ID, LeaderID: picked[0].PlayerID, Members: map[string]bool{},
			Until: now.Add(DungeonReadyTimeout), FromQueue: true, PartyID: picked[0].PartyID,
			Difficulty: picked[0].Difficulty,
		}
		for _, e := range picked {
			members = append(members, e.PlayerID)
			ready.Members[e.PlayerID] = false
			w.removeQueue(e.PlayerID)
		}
		ready.Members[ready.LeaderID] = false
		w.Dungeons.ready[ready.LeaderID] = ready
		msg := marshal(TypeDungeonReadyCheck, DungeonReadyOut{
			DungeonID: def.ID, LeaderID: ready.LeaderID, Until: ready.Until.UnixMilli(),
			Members: readyMembers(members, ready.Members), FromQueue: true,
		})
		out = append(out, msg)
		for _, id := range members {
			out = append(out, marshal(TypeQueueUpdate, QueueView{State: QueueReady, DungeonID: def.ID, Name: def.Name, ToID: id}))
		}
	}
	return out
}

func pickQueueMatch(list []QueueEntry, minN, maxN int) []QueueEntry {
	seen := map[string]bool{}
	var picked []QueueEntry
	for _, e := range list {
		if seen[e.PlayerID] {
			continue
		}
		unit := []QueueEntry{e}
		if e.PartyID != "" {
			unit = nil
			for _, o := range list {
				if o.PartyID == e.PartyID && !seen[o.PlayerID] {
					unit = append(unit, o)
				}
			}
		}
		if len(picked)+len(unit) > maxN {
			continue
		}
		for _, u := range unit {
			seen[u.PlayerID] = true
			picked = append(picked, u)
		}
		if len(picked) >= maxN {
			break
		}
	}
	if len(picked) < minN {
		return nil
	}
	return picked
}

func readyMemberIDs(check *ReadyCheck) []string {
	ids := make([]string, 0, len(check.Members))
	if check.LeaderID != "" {
		ids = append(ids, check.LeaderID)
	}
	for id := range check.Members {
		if id != check.LeaderID {
			ids = append(ids, id)
		}
	}
	return ids
}

func (w *WorldState) requeueMembers(check *ReadyCheck) [][]byte {
	now := time.Now()
	var out [][]byte
	for id := range check.Members {
		m := w.players[id]
		if m == nil || !m.Connected || m.InstanceID != "" {
			continue
		}
		w.Dungeons.Queue.Entries = append(w.Dungeons.Queue.Entries, QueueEntry{
			PlayerID: id, Role: "FLEX", DungeonID: check.DungeonID, Difficulty: "NORMAL",
			Region: "dawn", JoinedAt: now, State: QueueQueued,
		})
		out = append(out, marshal(TypeQueueUpdate, w.queueViewFor(id)))
	}
	return out
}

func (w *WorldState) dungeonJoin(p *Player, instanceID string) [][]byte {
	if instanceID == "" {
		return rejectFor(p.ID, TypeDungeonJoin, "instance")
	}
	inst := w.Dungeons.instances[instanceID]
	if inst == nil {
		return rejectFor(p.ID, TypeDungeonJoin, "instance")
	}
	if !containsID(inst.Players, p.ID) {
		return rejectFor(p.ID, TypeDungeonJoin, "membership")
	}
	return rejectFor(p.ID, TypeDungeonJoin, "already")
}

func (w *WorldState) dungeonFill(p *Player, instanceID string) [][]byte {
	inst := w.Dungeons.instances[instanceID]
	if inst == nil {
		return rejectFor(p.ID, TypeDungeonFill, "instance")
	}
	if inst.BossLocked || inst.State == DunBoss {
		return rejectFor(p.ID, TypeDungeonFill, "boss_lock")
	}
	if containsID(inst.Players, p.ID) || p.InstanceID != "" {
		return rejectFor(p.ID, TypeDungeonFill, "already")
	}
	def := dungeonByID[inst.DefID]
	if len(inst.Players) >= def.MaxPlayers {
		return rejectFor(p.ID, TypeDungeonFill, "full")
	}
	if p.Level < def.MinimumLevel {
		return rejectFor(p.ID, TypeDungeonFill, "level")
	}
	if w.raidLocked(p, def) {
		return rejectFor(p.ID, TypeDungeonFill, "lockout")
	}
	inst.Players = append(inst.Players, p.ID)
	inst.Return[p.ID] = [3]float64{p.X, p.Y, p.Z}
	inst.ReviveToken[p.ID] = 1
	p.InstanceID = inst.ID
	w.Dungeons.byPlayer[p.ID] = inst.ID
	p.X, p.Y, p.Z = inst.CheckpointX, 0, inst.CheckpointZ
	view := w.dungeonView(inst, p.ID)
	return [][]byte{marshal(TypeDungeonLoading, DungeonLoading{InstanceID: inst.ID, DungeonID: def.ID, Name: def.Name, ToID: p.ID}), marshal(TypeDungeonStarted, view), marshal(TypeDungeonState, view)}
}

func (w *WorldState) dungeonRevive(p *Player, targetID string) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypeDungeonRevive, "instance")
	}
	if !p.alive() {
		return rejectFor(p.ID, TypeDungeonRevive, "downed")
	}
	if targetID == "" || targetID == p.ID {
		return rejectFor(p.ID, TypeDungeonRevive, "target")
	}
	t := w.players[targetID]
	if t == nil || t.InstanceID != inst.ID {
		return rejectFor(p.ID, TypeDungeonRevive, "target")
	}
	if t.CombatState != "DOWNED" {
		return rejectFor(p.ID, TypeDungeonRevive, "not_downed")
	}
	if dist2(p.X, p.Z, t.X, t.Z) > 2.8 {
		return rejectFor(p.ID, TypeDungeonRevive, "distance")
	}
	if inst.ReviveToken[targetID] < 1 && inst.ReviveToken[p.ID] < 1 {
		return rejectFor(p.ID, TypeDungeonRevive, "token")
	}
	inst.Reviving[targetID] = p.ID
	if inst.ReviveProgress == nil {
		inst.ReviveProgress = map[string]float64{}
	}
	return [][]byte{marshal(TypeDungeonState, w.dungeonView(inst, p.ID))}
}

func (w *WorldState) finishRevive(inst *DungeonInstance, target, reviver *Player) [][]byte {
	if inst.ReviveToken[target.ID] > 0 {
		inst.ReviveToken[target.ID]--
	} else if inst.ReviveToken[reviver.ID] > 0 {
		inst.ReviveToken[reviver.ID]--
	}
	delete(inst.DownedAt, target.ID)
	delete(inst.ReviveProgress, target.ID)
	delete(inst.Reviving, target.ID)
	target.HP = int(float64(target.MaxHP) * 0.4)
	if target.HP < 1 {
		target.HP = 1
	}
	target.CombatState = "IDLE"
	target.State = "IDLE"
	target.IFrameUntil = time.Now().Add(1500 * time.Millisecond)
	return [][]byte{marshal(TypePlayerRevived, map[string]any{
		"playerId": target.ID, "reviverId": reviver.ID, "hp": target.HP, "instanceId": inst.ID,
	}), marshal(TypeDungeonState, w.dungeonView(inst, target.ID))}
}

func (w *WorldState) dungeonVote(p *Player, vote string) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypeDungeonVote, "instance")
	}
	vote = strings.ToLower(vote)
	if vote != "leave" && vote != "restart" {
		return rejectFor(p.ID, TypeDungeonVote, "vote")
	}
	if inst.Votes == nil {
		inst.Votes = map[string]string{}
	}
	if inst.Votes[p.ID] != "" {
		return rejectFor(p.ID, TypeDungeonVote, "already")
	}
	inst.Votes[p.ID] = vote
	need := (len(inst.Players)/2 + 1)
	count := 0
	for _, id := range inst.Players {
		if inst.Votes[id] == vote {
			count++
		}
	}
	out := [][]byte{marshal(TypeDungeonVoteUpdate, map[string]any{
		"instanceId": inst.ID, "vote": vote, "count": count, "need": need, "votes": inst.Votes,
	})}
	if count < need {
		return out
	}
	if vote == "leave" {
		return append(out, w.closeDungeon(inst, DunClosing, "Party meninggalkan dungeon.")...)
	}
	def := dungeonByID[inst.DefID]
	members := append([]string{}, inst.Players...)
	partyID := inst.PartyID
	w.closeDungeon(inst, DunClosing, "")
	return append(out, w.startDungeon(def, members, partyID, inst.Difficulty)...)
}

func (w *WorldState) dungeonTaunt(p *Player) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypeDungeonTaunt, "instance")
	}
	if inst.Roles[p.ID] != "TANK" && inst.Roles[p.ID] != "FLEX" {
		return rejectFor(p.ID, TypeDungeonTaunt, "role")
	}
	if inst.TauntUntil == nil {
		inst.TauntUntil = map[string]time.Time{}
	}
	inst.TauntUntil[p.ID] = time.Now().Add(8 * time.Second)
	if inst.Threat == nil {
		inst.Threat = map[string]float64{}
	}
	inst.Threat[p.ID] += 80
	return [][]byte{marshal(TypeDungeonState, w.dungeonView(inst, p.ID))}
}

func (w *WorldState) skipDungeonIntro(p *Player) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypeSkipDungeonIntro, "instance")
	}
	if inst.SkipIntro == nil {
		inst.SkipIntro = map[string]bool{}
	}
	inst.SkipIntro[p.ID] = true
	return [][]byte{marshal(TypeCinematicSkipped, map[string]any{"playerId": p.ID, "toId": p.ID})}
}

func (w *WorldState) seedDungeonMechanics(inst *DungeonInstance) {
	def := dungeonByID[inst.DefID]
	if def.Environment == "cavern" || def.Environment == "temple" || def.Environment == "depths" {
		inst.Objects = append(inst.Objects,
			ObjectSnapshot{ID: "puzzle-wind", Kind: "puzzle", X: -4, Z: 10, Text: "Wind"},
			ObjectSnapshot{ID: "puzzle-stone", Kind: "puzzle", X: 0, Z: 10, Text: "Stone"},
			ObjectSnapshot{ID: "puzzle-light", Kind: "puzzle", X: 4, Z: 10, Text: "Light"},
		)
	}
	if dungeonKind(def) == "RAID" {
		inst.Objects = append(inst.Objects,
			ObjectSnapshot{ID: "crystal-1", Kind: "crystal", X: -4, Z: 16, Text: "Raid Crystal"},
			ObjectSnapshot{ID: "crystal-2", Kind: "crystal", X: 0, Z: 16, Text: "Raid Crystal"},
			ObjectSnapshot{ID: "crystal-3", Kind: "crystal", X: 4, Z: 16, Text: "Raid Crystal"},
		)
	}
}

func (p *Player) grantCosmetic(id string) {
	if id == "" {
		return
	}
	log := p.ensureLog()
	for _, c := range log.Cosmetics {
		if c == id {
			return
		}
	}
	log.Cosmetics = append(log.Cosmetics, id)
	p.markDirty()
}
