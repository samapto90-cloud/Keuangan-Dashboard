package mmo

import (
	"math"
	"time"
)

func (w *WorldState) pvpQueueOf(playerID string) *PvpQueueEntry {
	if w.PvP == nil {
		return nil
	}
	for i := range w.PvP.queue {
		if w.PvP.queue[i].PlayerID == playerID {
			return &w.PvP.queue[i]
		}
	}
	return nil
}

func (w *WorldState) pvpQueueCount(mode string) int {
	n := 0
	for _, e := range w.PvP.queue {
		if e.Mode == mode {
			n++
		}
	}
	return n
}

func (w *WorldState) removePvpQueue(playerID string) {
	if w.PvP == nil {
		return
	}
	live := w.PvP.queue[:0]
	for _, e := range w.PvP.queue {
		if e.PlayerID != playerID {
			live = append(live, e)
		}
	}
	w.PvP.queue = live
}

func (w *WorldState) pvpQueueJoin(p *Player, mode string) [][]byte {
	def, ok := pvpMode(mode)
	if !ok || !def.Enabled || def.Kind == "DUEL" {
		return rejectFor(p.ID, TypePvpQueueJoin, "mode")
	}
	if p.Level < def.MinLevel {
		return rejectFor(p.ID, TypePvpQueueJoin, "level")
	}
	if p.InstanceID != "" {
		return rejectFor(p.ID, TypePvpQueueJoin, "already")
	}
	if w.pvpOf(p.ID) != nil {
		return rejectFor(p.ID, TypePvpQueueJoin, "match")
	}
	ab := w.pvpAbuse(p.ID)
	if time.Now().Before(ab.QueueLockUntil) {
		return rejectFor(p.ID, TypePvpQueueJoin, "lock")
	}
	partyN := 1
	partyID := p.PartyID
	if partyID != "" && w.Parties != nil {
		if pt := w.Parties.Of(p.ID); pt != nil {
			partyN = len(pt.Members)
		}
	}
	if def.TeamSize == 1 && partyN > 1 {
		return rejectFor(p.ID, TypePvpQueueJoin, "party")
	}
	if def.TeamSize > 1 && partyN != 1 && partyN != def.TeamSize {
		return rejectFor(p.ID, TypePvpQueueJoin, "size")
	}
	w.removePvpQueue(p.ID)
	log := p.ensureLog()
	w.ensurePvpRating(log)
	region := p.Region
	if region == "" {
		region = "ASIA"
	}
	w.PvP.queue = append(w.PvP.queue, PvpQueueEntry{
		PlayerID: p.ID, Mode: def.ID, Rating: log.PvpRating, Region: region,
		Latency: p.PingMs, PartyID: partyID, PartyN: partyN, JoinedAt: time.Now(),
	})
	evs := w.tickPvpQueue(time.Now())
	need := def.TeamSize * 2
	have := w.pvpQueueCount(def.ID)
	est, note := pvpQueueEstimate(need, have, 0)
	evs = append(evs, marshal(TypePvpQueueUpdate, PvpQueueView{
		State: "QUEUED", Mode: def.ID, Name: def.Name, Players: have,
		Need: need, WaitMs: 0, WaitEstMs: est, WaitNote: note, ToID: p.ID,
	}))
	return evs
}

func (w *WorldState) pvpQueueLeave(p *Player) [][]byte {
	w.removePvpQueue(p.ID)
	return [][]byte{marshal(TypePvpQueueUpdate, PvpQueueView{State: "IDLE", ToID: p.ID})}
}

func pvpRangeForWait(wait time.Duration) int {
	sec := wait.Seconds()
	if sec < 30 {
		return 100
	}
	if sec < 60 {
		return 200
	}
	return 300
}

func (w *WorldState) tickPvpQueue(now time.Time) [][]byte {
	if w.PvP == nil {
		return nil
	}
	var events [][]byte
	used := map[string]bool{}
	for i := range w.PvP.queue {
		a := w.PvP.queue[i]
		if used[a.PlayerID] {
			continue
		}
		def, ok := pvpMode(a.Mode)
		if !ok {
			continue
		}
		need := def.TeamSize * 2
		cands := []PvpQueueEntry{a}
		for j := i + 1; j < len(w.PvP.queue); j++ {
			b := w.PvP.queue[j]
			if used[b.PlayerID] || b.Mode != a.Mode || b.Region != a.Region {
				continue
			}
			wait := now.Sub(a.JoinedAt)
			if now.Sub(b.JoinedAt) > wait {
				wait = now.Sub(b.JoinedAt)
			}
			rng := pvpRangeForWait(wait)
			if def.Kind == "RANKED" {
				pa := w.players[a.PlayerID]
				if pa != nil && pa.ensureLog().PvpRankedMatches < pvpMod.Placement {
					rng += 100
				}
			}
			if absInt(a.Rating-b.Rating) > rng {
				continue
			}
			if a.Latency > 0 && b.Latency > 0 && absInt(a.Latency-b.Latency) > 140 && wait < 45*time.Second {
				continue
			}
			cands = append(cands, b)
			if len(cands) >= need {
				break
			}
		}
		if len(cands) < need {
			continue
		}
		ids := make([]string, 0, need)
		for _, c := range cands[:need] {
			ids = append(ids, c.PlayerID)
			used[c.PlayerID] = true
		}
		events = append(events, w.openPvpReady(ids, def, now)...)
	}
	if len(used) > 0 {
		live := w.PvP.queue[:0]
		for _, e := range w.PvP.queue {
			if !used[e.PlayerID] {
				live = append(live, e)
			}
		}
		w.PvP.queue = live
	}
	return events
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (w *WorldState) openPvpReady(ids []string, def PvpModeDef, now time.Time) [][]byte {
	inst := w.newPvpInstance(ids, def, now)
	inst.State = PvpMatched
	inst.ReadyUntil = now.Add(time.Duration(pvpMod.ReadyTimeout) * time.Second)
	if inst.ReadyUntil.IsZero() || pvpMod.ReadyTimeout <= 0 {
		inst.ReadyUntil = now.Add(15 * time.Second)
	}
	members := make([]PvpMemberView, 0, len(ids))
	for _, id := range ids {
		pl := w.players[id]
		name := id
		if pl != nil {
			name = pl.Name
		}
		f := inst.Fighters[id]
		team := 0
		if f != nil {
			team = f.Team
		}
		members = append(members, PvpMemberView{PlayerID: id, Name: name, Team: team, Ready: false})
	}
	return [][]byte{marshal(TypePvpReadyCheck, PvpReadyOut{
		MatchID: inst.ID, Mode: def.ID, Name: def.Name, Until: inst.ReadyUntil.UnixMilli(),
		Members: members, InstanceID: inst.ID,
	})}
}

func (w *WorldState) pvpReady(p *Player, accept bool) [][]byte {
	inst := w.pvpOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypePvpReady, "match")
	}
	if inst.State != PvpMatched && inst.State != PvpReadySt {
		return rejectFor(p.ID, TypePvpReady, "state")
	}
	if !accept {
		w.cancelPvp(inst, "decline")
		return [][]byte{marshal(TypePvpState, w.pvpView(inst, p.ID))}
	}
	if inst.Ready == nil {
		inst.Ready = map[string]bool{}
	}
	inst.Ready[p.ID] = true
	all := true
	for _, id := range inst.Players {
		if !inst.Ready[id] {
			all = false
			break
		}
	}
	if all {
		return w.beginPvp(inst)
	}
	inst.State = PvpReadySt
	return [][]byte{marshal(TypePvpReadyCheck, w.pvpReadyOut(inst))}
}

func (w *WorldState) pvpReadyOut(inst *PvpInstance) PvpReadyOut {
	members := make([]PvpMemberView, 0, len(inst.Players))
	for _, id := range inst.Players {
		name := id
		if pl := w.players[id]; pl != nil {
			name = pl.Name
		}
		team := 0
		if f := inst.Fighters[id]; f != nil {
			team = f.Team
		}
		members = append(members, PvpMemberView{PlayerID: id, Name: name, Team: team, Ready: inst.Ready[id]})
	}
	return PvpReadyOut{MatchID: inst.ID, Mode: inst.Mode, Until: inst.ReadyUntil.UnixMilli(), Members: members, InstanceID: inst.ID}
}

func (w *WorldState) newPvpInstance(ids []string, def PvpModeDef, now time.Time) *PvpInstance {
	id := randomID("pvp_")
	inst := &PvpInstance{
		ID: id, Mode: def.ID, Map: def.Map, State: PvpWaiting, Region: "ASIA",
		Players: append([]string{}, ids...), Fighters: map[string]*PvpFighter{},
		CreatedAt: now, MatchedAt: now, Ready: map[string]bool{}, Offline: map[string]time.Time{},
		Ranked: def.Kind == "RANKED", RewardID: randomID("prw_"),
	}
	size := def.TeamSize
	if size < 1 {
		size = 1
	}
	half := size
	if len(ids) < size*2 {
		half = len(ids) / 2
		if half < 1 {
			half = 1
		}
	}
	for i, pid := range ids {
		team := 1
		if i >= half {
			team = 2
		}
		if team == 1 {
			inst.TeamA = append(inst.TeamA, pid)
		} else {
			inst.TeamB = append(inst.TeamB, pid)
		}
		pl := w.players[pid]
		f := &PvpFighter{PlayerID: pid, Team: team, DamageFrom: map[string]int{}, CCHits: map[string]int{}, CCImmuneUntil: map[string]time.Time{}, LastInput: now}
		if pl != nil {
			f.Name = pl.Name
			f.SnapLevel = pl.Level
			f.SnapAtk = pl.Strength
			f.SnapDef = pl.Defense
			f.SnapHP = pl.MaxHP
			f.SnapMaxHP = pl.MaxHP
			f.Skills = append([]string{}, pl.UnlockedSkills...)
			f.Forms = append([]string{}, pl.UnlockedForms...)
			f.Return = [3]float64{pl.X, pl.Y, pl.Z}
			f.Rating = pl.ensureLog().PvpRating
			w.ensurePvpRating(pl.ensureLog())
			f.Rating = pl.ensureLog().PvpRating
			pl.InstanceID = id
		}
		inst.Fighters[pid] = f
		w.PvP.byPlayer[pid] = id
	}
	if def.Map == "valley-of-dawn" || def.Map == "dawn-arena" || def.Kind == "BATTLEGROUND" {
		inst.Shrines = []*PvpShrine{
			{ID: "A", X: -12, Z: 0, Radius: 4.2},
			{ID: "B", X: 0, Z: 0, Radius: 4.2},
			{ID: "C", X: 12, Z: 0, Radius: 4.2},
		}
	}
	w.PvP.instances[id] = inst
	return inst
}

func (w *WorldState) beginPvp(inst *PvpInstance) [][]byte {
	now := time.Now()
	def, _ := pvpMode(inst.Mode)
	inst.State = PvpLoading
	dur := time.Duration(def.Duration) * time.Second
	if dur <= 0 {
		dur = 5 * time.Minute
	}
	cd := time.Duration(pvpMod.Countdown) * time.Second
	if cd <= 0 {
		cd = 3 * time.Second
	}
	prot := time.Duration(pvpMod.SpawnProtect) * time.Second
	if prot <= 0 {
		prot = 3 * time.Second
	}
	w.placePvpSpawns(inst)
	inst.State = PvpCountdown
	inst.StartedAt = now
	inst.CountdownUntil = now.Add(cd)
	inst.ProtectUntil = inst.CountdownUntil.Add(prot)
	inst.EndsAt = inst.CountdownUntil.Add(dur)
	var events [][]byte
	events = append(events, marshal(TypePvpLoading, map[string]any{"matchId": inst.ID, "mode": inst.Mode, "map": inst.Map, "name": def.Name, "instanceId": inst.ID}))
	for _, id := range inst.Players {
		if p := w.players[id]; p != nil {
			p.HP = p.MaxHP
			p.Energy = p.MaxEnergy
			p.CombatState = "IDLE"
			p.PvpNoActUntil = inst.ProtectUntil
			p.InvulnUntil = inst.ProtectUntil
			p.AX, p.AZ, p.VX, p.VZ = 0, 0, 0, 0
		}
	}
	events = append(events, marshal(TypePvpCountdown, map[string]any{"matchId": inst.ID, "until": inst.CountdownUntil.UnixMilli(), "instanceId": inst.ID}))
	events = append(events, marshal(TypePvpState, w.pvpView(inst, "")))
	w.pvpReplay(inst, "start", "", map[string]any{"mode": inst.Mode})
	return events
}

func (w *WorldState) placePvpSpawns(inst *PvpInstance) {
	bg := pvpIsBattleground(inst)
	for i, id := range inst.TeamA {
		p := w.players[id]
		if p == nil {
			continue
		}
		if bg {
			p.X, p.Y, p.Z = -18, 0, -6+float64(i)*3
		} else {
			p.X, p.Y, p.Z = -8, 0, float64(i)*2.2
		}
		p.Yaw = 1.57
	}
	for i, id := range inst.TeamB {
		p := w.players[id]
		if p == nil {
			continue
		}
		if bg {
			p.X, p.Y, p.Z = 18, 0, -6+float64(i)*3
		} else {
			p.X, p.Y, p.Z = 8, 0, float64(i)*2.2
		}
		p.Yaw = -1.57
	}
	_ = math.Pi
}

func (w *WorldState) cancelPvp(inst *PvpInstance, reason string) {
	if inst == nil || inst.State == PvpCompleted || inst.State == PvpCancelled {
		return
	}
	inst.State = PvpCancelled
	inst.EndedAt = time.Now()
	inst.Result = reason
	for _, id := range inst.Players {
		if p := w.players[id]; p != nil {
			w.extractPvpPlayer(p, inst)
		} else {
			delete(w.PvP.byPlayer, id)
		}
	}
}
