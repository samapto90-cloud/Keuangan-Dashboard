package mmo

import (
	"math"
	"strings"
	"time"
)

func (w *WorldState) tickPvp(now time.Time) [][]byte {
	if w.PvP == nil {
		return nil
	}
	w.PvP.ensure()
	var events [][]byte
	events = append(events, w.tickPvpQueue(now)...)
	for id, inst := range w.PvP.instances {
		if inst.State == PvpCompleted {
			if !inst.EndedAt.IsZero() && now.Sub(inst.EndedAt) > 2*time.Second {
				inst.State = PvpCleanup
			}
			continue
		}
		if inst.State == PvpCleanup {
			if !inst.EndedAt.IsZero() && now.Sub(inst.EndedAt) > 8*time.Second {
				delete(w.PvP.instances, id)
			}
			continue
		}
		if inst.State == PvpCancelled {
			continue
		}
		if (inst.State == PvpMatched || inst.State == PvpReadySt) && now.After(inst.ReadyUntil) && !inst.ReadyUntil.IsZero() {
			w.cancelPvp(inst, "timeout")
			events = append(events, marshal(TypePvpState, w.pvpView(inst, "")))
			continue
		}
		if inst.State == PvpCountdown && !now.Before(inst.CountdownUntil) {
			inst.State = PvpActive
			for _, id := range inst.Players {
				if p := w.players[id]; p != nil {
					if now.After(inst.ProtectUntil) {
						p.PvpNoActUntil = time.Time{}
					}
				}
			}
			events = append(events, marshal(TypePvpState, w.pvpView(inst, "")))
			events = append(events, marshal(TypePvpCountdown, map[string]any{"matchId": inst.ID, "fight": true, "instanceId": inst.ID}))
		}
		if inst.State == PvpActive {
			events = append(events, w.tickPvpActive(inst, now)...)
		}
		events = append(events, w.tickPvpDisconnect(inst, now)...)
	}
	return events
}

func (w *WorldState) tickPvpActive(inst *PvpInstance, now time.Time) [][]byte {
	var events [][]byte
	events = append(events, w.tickPvpAfk(inst, now)...)
	events = append(events, w.tickPvpRespawn(inst, now)...)
	if pvpIsBattleground(inst) {
		events = append(events, w.tickPvpCapture(inst, now)...)
		if inst.ScoreA >= pvpMod.ScoreLimit {
			events = append(events, w.finishPvp(inst, 1, "score")...)
			return events
		}
		if inst.ScoreB >= pvpMod.ScoreLimit {
			events = append(events, w.finishPvp(inst, 2, "score")...)
			return events
		}
	} else if w.pvpTeamWiped(inst, 2) {
		events = append(events, w.finishPvp(inst, 1, "defeat")...)
		return events
	} else if w.pvpTeamWiped(inst, 1) {
		events = append(events, w.finishPvp(inst, 2, "defeat")...)
		return events
	}
	if now.After(inst.EndsAt) {
		events = append(events, w.finishPvpOnTime(inst)...)
	}
	return events
}

func (w *WorldState) pvpTeamWiped(inst *PvpInstance, team int) bool {
	ids := inst.TeamA
	if team == 2 {
		ids = inst.TeamB
	}
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		p := w.players[id]
		if p != nil && p.alive() && p.CombatState != "DEAD" && p.CombatState != "RESPAWNING" {
			return false
		}
		if p == nil {
			if f := inst.Fighters[id]; f != nil && f.Deaths == 0 {
				return false
			}
		}
	}
	return true
}

func (w *WorldState) tickPvpRespawn(inst *PvpInstance, now time.Time) [][]byte {
	def, _ := pvpMode(inst.Mode)
	if !def.Respawn {
		return nil
	}
	delay := pvpRespawnDelay(def)
	if delay <= 0 {
		delay = 8 * time.Second
	}
	var events [][]byte
	for _, id := range inst.Players {
		p := w.players[id]
		if p == nil {
			continue
		}
		if p.CombatState != "RESPAWNING" && p.CombatState != "DEAD" {
			continue
		}
		if now.Before(p.RespawnAt) {
			continue
		}
		p.HP = p.MaxHP
		p.Energy = p.MaxEnergy
		p.CombatState = "IDLE"
		p.State = "IDLE"
		f := inst.Fighters[id]
		if f != nil && f.Team == 1 {
			p.X, p.Z = -18, 0
		} else {
			p.X, p.Z = 18, 0
		}
		prot := time.Duration(pvpMod.SpawnProtect) * time.Second
		if prot <= 0 {
			prot = 3 * time.Second
		}
		p.InvulnUntil = now.Add(prot)
		p.PvpNoActUntil = now.Add(prot)
		events = append(events, marshal(TypePvpState, map[string]any{"playerId": id, "respawn": true, "instanceId": inst.ID}))
	}
	return events
}

func (w *WorldState) tickPvpAfk(inst *PvpInstance, now time.Time) [][]byte {
	warn := time.Duration(pvpMod.AfkWarn) * time.Second
	pen := time.Duration(pvpMod.AfkPenalty) * time.Second
	if warn <= 0 {
		warn = 20 * time.Second
	}
	if pen <= 0 {
		pen = 45 * time.Second
	}
	var events [][]byte
	for _, id := range inst.Players {
		p := w.players[id]
		f := inst.Fighters[id]
		if p == nil || f == nil || !p.Connected {
			continue
		}
		idle := now.Sub(p.LastInputAt)
		if idle < warn {
			f.AfkWarned = false
			continue
		}
		if !f.AfkWarned {
			f.AfkWarned = true
			events = append(events, marshal(TypePvpAfk, map[string]any{"playerId": id, "warning": true, "text": pvpAfkWarnText, "toId": id, "instanceId": inst.ID}))
			continue
		}
		if idle >= pen && !f.AfkFlagged {
			f.AfkFlagged = true
			ab := w.pvpAbuse(id)
			ab.Afk++
			if pvpIsBattleground(inst) && ab.Afk >= 2 {
				ab.QueueLockUntil = now.Add(time.Duration(pvpMod.LeaveLock) * time.Second)
			}
			if inst.Ranked && !pvpIsBattleground(inst) {
				events = append(events, w.finishPvp(inst, otherTeamOf(inst, id), "afk")...)
				return events
			}
			events = append(events, marshal(TypePvpAfk, map[string]any{"playerId": id, "penalty": true, "toId": id, "instanceId": inst.ID}))
		}
	}
	return events
}

func (w *WorldState) tickPvpDisconnect(inst *PvpInstance, now time.Time) [][]byte {
	var events [][]byte
	for _, id := range inst.Players {
		p := w.players[id]
		if p != nil && p.Connected {
			delete(inst.Offline, id)
			continue
		}
		since, ok := inst.Offline[id]
		if !ok {
			continue
		}
		if now.Sub(since) < pvpReconnectWindow() {
			continue
		}
		if inst.State == PvpActive || inst.State == PvpCountdown {
			if inst.Ranked {
				events = append(events, w.finishPvp(inst, otherTeamOf(inst, id), "disconnect")...)
				return events
			}
		}
	}
	return events
}

func (w *WorldState) tickPvpCapture(inst *PvpInstance, now time.Time) [][]byte {
	var events [][]byte
	rate := pvpMod.CaptureRate
	if rate <= 0 {
		rate = 28
	}
	dt := 0.05
	for _, s := range inst.Shrines {
		a, b := 0, 0
		for _, id := range inst.Players {
			p := w.players[id]
			f := inst.Fighters[id]
			if p == nil || f == nil || !p.alive() {
				continue
			}
			if math.Hypot(p.X-s.X, p.Z-s.Z) > s.Radius {
				continue
			}
			if f.Team == 1 {
				a++
			} else {
				b++
			}
		}
		prev := s.Owner
		s.Contested = a > 0 && b > 0
		if s.Contested {
			continue
		}
		if a > 0 {
			s.ProgressA = math.Min(100, s.ProgressA+rate*dt)
			s.ProgressB = math.Max(0, s.ProgressB-rate*dt*0.5)
			if s.ProgressA >= 100 {
				s.Owner = 1
				s.ProgressA = 100
			}
		} else if b > 0 {
			s.ProgressB = math.Min(100, s.ProgressB+rate*dt)
			s.ProgressA = math.Max(0, s.ProgressA-rate*dt*0.5)
			if s.ProgressB >= 100 {
				s.Owner = 2
				s.ProgressB = 100
			}
		}
		if s.Owner != prev && s.Owner != 0 {
			w.creditPvpCapture(inst, s, s.Owner)
			w.pvpReplay(inst, "objective", "", map[string]any{"shrine": s.ID, "owner": s.Owner})
		}
		if s.Owner != 0 && !s.Contested {
			if s.LastTick.IsZero() || now.Sub(s.LastTick) >= time.Second {
				s.LastTick = now
				add := pvpMod.ScoreTick
				if add <= 0 {
					add = 8
				}
				if s.Owner == 1 {
					inst.ScoreA += add
					if inst.ScoreA > pvpMod.ScoreLimit {
						inst.ScoreA = pvpMod.ScoreLimit
					}
				} else {
					inst.ScoreB += add
					if inst.ScoreB > pvpMod.ScoreLimit {
						inst.ScoreB = pvpMod.ScoreLimit
					}
				}
				events = append(events, marshal(TypePvpCapture, w.pvpCaptureView(inst)))
			}
		}
	}
	return events
}

func (w *WorldState) pvpCaptureView(inst *PvpInstance) map[string]any {
	pts := make([]map[string]any, 0, len(inst.Shrines))
	for _, s := range inst.Shrines {
		pts = append(pts, map[string]any{
			"id": s.ID, "owner": s.Owner, "contested": s.Contested,
			"progressA": int(s.ProgressA + 0.5), "progressB": int(s.ProgressB + 0.5),
			"x": s.X, "z": s.Z,
		})
	}
	return map[string]any{"scoreA": inst.ScoreA, "scoreB": inst.ScoreB, "points": pts, "instanceId": inst.ID}
}

func (w *WorldState) resolvePvpTarget(p *Player, inst *PvpInstance, targetID string, maxDist float64) *Player {
	if inst == nil || inst.State != PvpActive {
		return nil
	}
	now := time.Now()
	if now.Before(inst.CountdownUntil) {
		return nil
	}
	self := inst.Fighters[p.ID]
	if self == nil {
		return nil
	}
	pick := func(t *Player) *Player {
		if t == nil || t.ID == p.ID || !t.alive() {
			return nil
		}
		if t.InstanceID != inst.ID {
			return nil
		}
		f := inst.Fighters[t.ID]
		if f == nil || f.Team == self.Team {
			return nil
		}
		if math.Hypot(t.X-p.X, t.Z-p.Z) > maxDist {
			w.flagSuspicious(p.ID, 4, "range")
			return nil
		}
		return t
	}
	if targetID != "" {
		return pick(w.players[targetID])
	}
	var best *Player
	bestD := maxDist
	for _, id := range inst.Players {
		t := w.players[id]
		if pick(t) == nil {
			continue
		}
		d := math.Hypot(t.X-p.X, t.Z-p.Z)
		if d < bestD {
			best, bestD = t, d
		}
	}
	return best
}

func (w *WorldState) pvpAttack(p *Player, inst *PvpInstance, in AttackIn, base int, power, rng, kb float64, kind string, hits int, finisher bool, display string) [][]byte {
	target := w.resolvePvpTarget(p, inst, in.TargetID, rng+RangeSlack)
	if target == nil {
		return rejectFor(p.ID, "PLAYER_ATTACK", "target")
	}
	now := time.Now()
	p.AttackCDUntil = now.Add(agilityCooldown(p, 280*time.Millisecond))
	if kind == "heavy" {
		p.AttackCDUntil = now.Add(agilityCooldown(p, 420*time.Millisecond))
	}
	p.CombatState = "ATTACKING"
	if hits > 1 {
		p.CombatState = "COMBO"
	}
	p.InCombatUntil = now.Add(CombatIdleAfter)
	events := [][]byte{marshal(TypeAttackResult, AttackResult{
		AttackerID: p.ID, AttackType: display, TargetIDs: []string{target.ID}, Timestamp: now.UnixMilli(),
		ComboHits: hits, Finisher: finisher,
	})}
	cc := ""
	if finisher || kind == "heavy" {
		cc = "stun"
	}
	events = append(events, w.hitPvp(p, target, inst, float64(base), power, kb, kind, cc)...)
	return events
}

func (w *WorldState) pvpSkill(p *Player, inst *PvpInstance, def SkillDef, in SkillIn, power float64) [][]byte {
	now := time.Now()
	self := def.Target == "SELF" || def.Effect == "dash" || def.Effect == "guard" || def.Damage <= 0
	var targets []*Player
	if !self {
		maxR := def.Range + RangeSlack
		if def.Radius > 0 || def.Target == "AREA" || def.Target == "DIRECTION" {
			r := def.Radius
			if r < 1 {
				r = def.Range
			}
			for _, id := range inst.Players {
				t := w.resolvePvpTarget(p, inst, id, r+RangeSlack)
				if t != nil {
					targets = append(targets, t)
				}
			}
		} else {
			t := w.resolvePvpTarget(p, inst, in.TargetID, maxR)
			if t != nil {
				targets = []*Player{t}
			}
		}
		if len(targets) == 0 {
			return rejectFor(p.ID, "PLAYER_SKILL", "target")
		}
	}
	ids := make([]string, 0, len(targets))
	for _, t := range targets {
		ids = append(ids, t.ID)
	}
	events := [][]byte{
		marshal(TypeAttackResult, AttackResult{AttackerID: p.ID, AttackType: "skill", SkillID: def.ID, TargetIDs: ids, Timestamp: now.UnixMilli()}),
		marshal(TypeSkillUsed, map[string]any{"playerId": p.ID, "skillId": def.ID}),
	}
	w.pvpReplay(inst, "skill", p.ID, map[string]any{"skillId": def.ID, "targets": ids})
	cc := ""
	if def.Kind == "ULTIMATE" {
		cc = "stun"
	} else if def.Type == "ENERGY" {
		cc = "slow"
	}
	for _, t := range targets {
		events = append(events, w.hitPvp(p, t, inst, float64(def.Damage), power, 3.2, def.Kind, cc)...)
	}
	return events
}

func (w *WorldState) hitPvp(atk, tgt *Player, inst *PvpInstance, base, skillPower, knockback float64, attackType, cc string) [][]byte {
	now := time.Now()
	if !tgt.alive() || tgt.invulnerable(now) {
		return nil
	}
	if now.Before(inst.ProtectUntil) {
		return nil
	}
	equip := EquipmentBonus + float64(atk.EquipAttack)*0.04
	chance := CritChance
	if atk.CritChance > 0 {
		chance = atk.CritChance
	}
	dealt, crit := calcDamageChance(base, float64(atk.Strength), skillPower, equip, float64(tgt.Defense), chance)
	dealt = w.applyPvpModifiers(atk, tgt, inst, dealt, attackType)
	if dealt < 1 {
		dealt = 1
	}
	if dealt > atk.MaxHP*4 {
		w.flagSuspicious(atk.ID, 25, "damage")
		return rejectFor(atk.ID, "PLAYER_ATTACK", "damage")
	}
	tgt.HP -= dealt
	if tgt.HP < 0 {
		tgt.HP = 0
	}
	tgt.CombatState = "HIT"
	tgt.HitUntil = now.Add(HitStun)
	tgt.InCombatUntil = now.Add(CombatIdleAfter)
	if knockback > 0 {
		dur := w.pvpCCDuration(inst, tgt.ID, "knockback", 220*time.Millisecond)
		if dur > 0 {
			tgt.VX += math.Sin(atk.Yaw) * knockback * 0.18
			tgt.VZ += math.Cos(atk.Yaw) * knockback * 0.18
		}
	}
	af := inst.Fighters[atk.ID]
	tf := inst.Fighters[tgt.ID]
	if af != nil {
		af.DmgDealt += dealt
		af.LastInput = now
	}
	if tf != nil {
		tf.DmgTaken += dealt
		if tf.DamageFrom == nil {
			tf.DamageFrom = map[string]int{}
		}
		tf.DamageFrom[atk.ID] += dealt
	}
	if cc != "" {
		w.applyPvpCC(inst, atk, tgt, cc)
	}
	killed := tgt.HP <= 0
	events := [][]byte{
		marshal(TypeDamageResult, DamageResult{
			AttackerID: atk.ID, TargetID: tgt.ID, Damage: dealt, IsCritical: crit,
			HitX: tgt.X, HitY: tgt.Y + 1.4, HitZ: tgt.Z, AttackType: attackType,
			Timestamp: now.UnixMilli(), TargetHP: tgt.HP, TargetMaxHP: tgt.MaxHP, Killed: killed, Kind: "player",
		}),
		marshal(TypePlayerHit, DamageResult{
			AttackerID: atk.ID, TargetID: tgt.ID, Damage: dealt, IsCritical: crit,
			HitX: tgt.X, HitY: tgt.Y + 1.4, HitZ: tgt.Z, AttackType: attackType,
			Timestamp: now.UnixMilli(), TargetHP: tgt.HP, TargetMaxHP: tgt.MaxHP, Killed: killed, Kind: "player",
		}),
	}
	if killed {
		events = append(events, w.killPvpPlayer(atk, tgt, inst)...)
	}
	return events
}

func (w *WorldState) applyPvpModifiers(atk, tgt *Player, inst *PvpInstance, dealt int, attackType string) int {
	v := float64(dealt) * pvpMod.DamageDealt * pvpMod.DamageTaken
	v *= pvpTransformFactor(atk)
	kind := strings.ToLower(attackType)
	if kind != "punch" && kind != "light" && kind != "heavy" && kind != "finisher" && pvpMod.SkillDamage > 0 {
		v *= pvpMod.SkillDamage
	}
	if attackType == "ULTIMATE" || attackType == "ultimate" {
		ult := pvpMod.Ultimate
		if ult <= 0 {
			ult = 0.75
		}
		v *= ult
	}
	if inst.Ranked && pvpMod.LevelScale > 0 {
		diff := atk.Level - tgt.Level
		if diff > 0 {
			scale := pvpMod.LevelScale * math.Min(0.5, float64(diff)/40.0)
			v *= 1 - scale
		}
	}
	if v < 1 {
		return 1
	}
	return int(v + 0.5)
}

func pvpTransformFactor(p *Player) float64 {
	if p.FormID == "" || p.FormID == "normal" {
		return 1
	}
	def := transformByID[p.FormID]
	if def.AttackPct <= 0 {
		return 1
	}
	pvpPct := pvpMod.TransformPct[p.FormID]
	if pvpPct <= 0 {
		pvpPct = def.AttackPct * 0.58
	}
	return (1 + pvpPct) / (1 + def.AttackPct)
}

func (w *WorldState) applyPvpCC(inst *PvpInstance, atk, tgt *Player, kind string) {
	dur := 900 * time.Millisecond
	if kind == "slow" {
		dur = 1200 * time.Millisecond
	}
	dur = w.pvpCCDuration(inst, tgt.ID, kind, dur)
	if dur <= 0 {
		return
	}
	now := time.Now()
	switch kind {
	case "stun":
		tgt.StunUntil = now.Add(dur)
		tgt.CombatState = "STUNNED"
	case "silence":
		tgt.SilenceUntil = now.Add(dur)
	case "slow":
		tgt.SlowUntil = now.Add(dur)
		tgt.SlowFactor = 0.55
	}
	_ = atk
}

func (w *WorldState) pvpCCDuration(inst *PvpInstance, targetID, kind string, base time.Duration) time.Duration {
	f := inst.Fighters[targetID]
	if f == nil {
		return base
	}
	now := time.Now()
	if f.CCImmuneUntil == nil {
		f.CCImmuneUntil = map[string]time.Time{}
	}
	if now.Before(f.CCImmuneUntil[kind]) {
		return 0
	}
	n := f.CCHits[kind]
	f.CCHits[kind] = n + 1
	mult := 1.0
	switch n {
	case 0:
		mult = 1
	case 1:
		mult = 0.5
	case 2:
		mult = 0.25
	default:
		f.CCImmuneUntil[kind] = now.Add(4 * time.Second)
		return 0
	}
	return time.Duration(float64(base) * mult)
}

func (w *WorldState) killPvpPlayer(atk, tgt *Player, inst *PvpInstance) [][]byte {
	now := time.Now()
	tgt.HP = 0
	tgt.AX, tgt.AZ, tgt.VX, tgt.VZ = 0, 0, 0, 0
	def, _ := pvpMode(inst.Mode)
	if def.Respawn {
		tgt.CombatState = "RESPAWNING"
		tgt.State = "RESPAWNING"
		d := pvpRespawnDelay(def)
		if d <= 0 {
			d = 8 * time.Second
		}
		tgt.RespawnAt = now.Add(d)
	} else {
		tgt.CombatState = "DEAD"
		tgt.State = "DEAD"
		tgt.RespawnAt = now.Add(24 * time.Hour)
	}
	tf := inst.Fighters[tgt.ID]
	af := inst.Fighters[atk.ID]
	if tf != nil {
		tf.Deaths++
		for id, dmg := range tf.DamageFrom {
			if id == atk.ID || dmg < 1 {
				continue
			}
			if as := inst.Fighters[id]; as != nil {
				as.Assists++
			}
		}
	}
	if af != nil {
		af.Kills++
	}
	line := "[" + atk.Name + "] defeated [" + tgt.Name + "]"
	inst.KillFeed = append(inst.KillFeed, line)
	if len(inst.KillFeed) > 12 {
		inst.KillFeed = inst.KillFeed[len(inst.KillFeed)-12:]
	}
	w.pvpReplay(inst, "kill", atk.ID, map[string]any{"target": tgt.ID})
	spec := ""
	if pvpAllowSpectate(inst) {
		if def.TeamSize == 1 {
			spec = atk.ID
		} else if tf != nil {
			for _, id := range inst.Players {
				f := inst.Fighters[id]
				pl := w.players[id]
				if f != nil && pl != nil && f.Team == tf.Team && pl.alive() {
					spec = id
					break
				}
			}
		}
	}
	if tf != nil {
		tf.SpectateID = spec
	}
	var events [][]byte
	if tgt.FormID != "" && tgt.FormID != "normal" {
		events = append(events, w.endTransform(tgt, "death")...)
	}
	events = append(events,
		marshal(TypePlayerDeath, DeathOut{PlayerID: tgt.ID, RespawnAt: tgt.RespawnAt.UnixMilli()}),
		marshal(TypePvpKillFeed, map[string]any{"text": line, "killerId": atk.ID, "targetId": tgt.ID, "instanceId": inst.ID}),
	)
	if spec != "" {
		events = append(events, marshal(TypePvpSpectate, map[string]any{"playerId": tgt.ID, "targetId": spec, "mode": "PLAYER", "toId": tgt.ID, "instanceId": inst.ID}))
	}
	return events
}

func (w *WorldState) finishPvpOnTime(inst *PvpInstance) [][]byte {
	if pvpIsBattleground(inst) {
		if inst.ScoreA > inst.ScoreB {
			return w.finishPvp(inst, 1, "timer")
		}
		if inst.ScoreB > inst.ScoreA {
			return w.finishPvp(inst, 2, "timer")
		}
		return w.finishPvp(inst, 0, "draw")
	}
	pct := func(ids []string) float64 {
		sum, n := 0.0, 0.0
		for _, id := range ids {
			p := w.players[id]
			if p == nil || p.MaxHP <= 0 {
				continue
			}
			sum += float64(p.HP) / float64(p.MaxHP)
			n++
		}
		if n == 0 {
			return 0
		}
		return sum / n
	}
	a, b := pct(inst.TeamA), pct(inst.TeamB)
	if a > b {
		return w.finishPvp(inst, 1, "hp")
	}
	if b > a {
		return w.finishPvp(inst, 2, "hp")
	}
	return w.finishPvp(inst, 0, "draw")
}

func (w *WorldState) suddenTeam(inst *PvpInstance) int {
	da, db := 0, 0
	for _, id := range inst.TeamA {
		if f := inst.Fighters[id]; f != nil {
			da += f.DmgDealt
		}
	}
	for _, id := range inst.TeamB {
		if f := inst.Fighters[id]; f != nil {
			db += f.DmgDealt
		}
	}
	if da > db {
		return 1
	}
	if db > da {
		return 2
	}
	return 1
}

func (w *WorldState) flagSuspicious(id string, score float64, reason string) {
	if w.PvP == nil {
		return
	}
	ab := w.pvpAbuse(id)
	ab.Suspicious += score
	if p := w.players[id]; p != nil {
		p.Suspicious += score
	}
	_ = reason
}

func (w *WorldState) validatePvpMove(p *Player, dt float64) {
	if dt <= 0 {
		return
	}
	speed := math.Hypot(p.VX, p.VZ)
	max := 16.0
	if p.Sprint {
		max = 22
	}
	if now := time.Now(); now.Before(p.DodgeUntil) {
		max = 38
	}
	if speed > max {
		w.flagSuspicious(p.ID, 2, "speed")
		scale := max / speed
		p.VX *= scale
		p.VZ *= scale
	}
}
