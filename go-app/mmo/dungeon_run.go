package mmo

import (
	"encoding/json"
	"math"
	"strings"
	"time"
)

func (w *WorldState) tickDungeons(dt float64, now time.Time) [][]byte {
	if w.Dungeons == nil {
		return nil
	}
	for _, ev := range w.tickQueue(now) {
		var body struct {
			ToID string `json:"toId"`
		}
		var env Envelope
		if json.Unmarshal(ev, &env) == nil {
			_ = json.Unmarshal(env.Data, &body)
		}
		if body.ToID != "" {
			w.SendToLocked(body.ToID, ev)
			continue
		}
		if env.Type == TypeDungeonReadyCheck {
			var ready DungeonReadyOut
			_ = json.Unmarshal(env.Data, &ready)
			for _, m := range ready.Members {
				w.SendToLocked(m.PlayerID, ev)
			}
			continue
		}
		for _, e := range w.Dungeons.Queue.Entries {
			w.SendToLocked(e.PlayerID, ev)
		}
	}
	for _, inst := range w.Dungeons.instances {
		events := w.tickDungeon(inst, dt, now)
		for _, ev := range events {
			for _, id := range inst.Players {
				w.SendToLocked(id, ev)
			}
		}
	}
	w.cleanupIdleInstances(now)
	return nil
}

func (w *WorldState) tickDungeon(inst *DungeonInstance, dt float64, now time.Time) [][]byte {
	if inst.State == DunClosing || inst.State == DunCompleted {
		return nil
	}
	if inst.State == DunFailed {
		return nil
	}
	if now.After(inst.EndsAt) {
		return w.failDungeon(inst, "Waktu habis.")
	}
	if w.dungeonWipe(inst) {
		return w.wipeToCheckpoint(inst)
	}
	w.expireOffline(inst, now)
	w.tickRevive(inst, dt)
	w.tickCrystals(inst)
	w.resetLeashedPacks(inst)
	var events [][]byte
	if inst.Telegraph != nil && !now.Before(inst.Telegraph.Until) {
		events = append(events, w.resolveTelegraph(inst)...)
	}
	allow := func(p *Player) bool {
		return p != nil && p.Connected && p.InstanceID == inst.ID
	}
	events = append(events, w.tickEnemyMap(inst.Enemies, allow, false, dt, now)...)
	if inst.State == DunBoss {
		events = append(events, w.tickBoss(inst, now)...)
		events = append(events, w.tickRaidMechanics(inst, now)...)
	}
	if inst.State == DunActive {
		events = append(events, w.checkDungeonProgress(inst)...)
	}
	return events
}

func (w *WorldState) dungeonWipe(inst *DungeonInstance) bool {
	alive := 0
	pending := 0
	for _, id := range inst.Players {
		p := w.players[id]
		if p != nil && p.Connected && p.alive() {
			alive++
			continue
		}
		if _, ok := inst.OfflineSince[id]; ok {
			pending++
			continue
		}
		if w.Dungeons.pending[id] != nil {
			pending++
		}
	}
	return alive == 0 && pending == 0 && len(inst.Players) > 0
}

func (w *WorldState) wipeToCheckpoint(inst *DungeonInstance) [][]byte {
	inst.WipeCount++
	inst.EndsAt = inst.EndsAt.Add(-DungeonWipePenalty)
	if time.Until(inst.EndsAt) < 30*time.Second {
		inst.EndsAt = time.Now().Add(30 * time.Second)
	}
	inst.Telegraph = nil
	inst.Votes = map[string]string{}
	if dungeonByID[inst.DefID].EducationBoss != "" && inst.Enemies[inst.BossID] != nil && inst.Enemies[inst.BossID].Alive {
		inst.EduShield = true
		inst.CrystalShield = true
	} else {
		inst.CrystalShield = false
	}
	for _, id := range inst.Players {
		p := w.players[id]
		if p == nil {
			continue
		}
		p.X, p.Y, p.Z = inst.CheckpointX, 0, inst.CheckpointZ
		p.VX, p.VZ = 0, 0
		p.HP = p.MaxHP
		p.CombatState = "IDLE"
		p.State = "IDLE"
		inst.ReviveToken[id] = 1
		delete(inst.DownedAt, id)
		delete(inst.ReviveProgress, id)
		delete(inst.Reviving, id)
	}
	w.resetBoss(inst)
	view := w.dungeonView(inst, firstPlayer(inst))
	view.Toast = "Party wipe. Kembali ke checkpoint."
	return [][]byte{marshal(TypeDungeonWipe, view), marshal(TypeDungeonState, view)}
}

func (w *WorldState) tickRevive(inst *DungeonInstance, dt float64) {
	for targetID, reviverID := range inst.Reviving {
		t := w.players[targetID]
		r := w.players[reviverID]
		if t == nil || r == nil || t.CombatState != "DOWNED" || !r.alive() || dist2(r.X, r.Z, t.X, t.Z) > 3.0 {
			delete(inst.Reviving, targetID)
			delete(inst.ReviveProgress, targetID)
			continue
		}
		inst.ReviveProgress[targetID] += dt / ReviveChannelTime * 100
		if inst.ReviveProgress[targetID] >= 100 {
			for _, ev := range w.finishRevive(inst, t, r) {
				for _, id := range inst.Players {
					w.SendToLocked(id, ev)
				}
			}
		}
	}
}

func (w *WorldState) tickCrystals(inst *DungeonInstance) {
	if inst.EduShield {
		inst.CrystalShield = true
		return
	}
	need := 0
	held := 0
	for _, obj := range inst.Objects {
		if obj.Kind != "crystal" {
			continue
		}
		need++
		for _, id := range inst.Players {
			p := w.players[id]
			if p != nil && p.alive() && dist2(p.X, p.Z, obj.X, obj.Z) < 1.6 {
				held++
				break
			}
		}
	}
	inst.CrystalShield = need > 0 && held >= need
}

func (w *WorldState) resetLeashedPacks(inst *DungeonInstance) {
	for _, e := range inst.Enemies {
		if e == nil || !e.Alive || e.IsBoss {
			continue
		}
		leash := e.Def.Leash
		if leash <= 0 {
			leash = 14
		}
		if math.Hypot(e.X-e.SX, e.Z-e.SZ) > leash {
			e.X, e.Z = e.SX, e.SZ
			e.HP = e.MaxHP
			e.State = "IDLE"
			e.TargetID = ""
		}
	}
}

func (w *WorldState) expireOffline(inst *DungeonInstance, now time.Time) {
	for id, since := range inst.OfflineSince {
		if now.Sub(since) > DungeonRejoinTimeout {
			delete(inst.OfflineSince, id)
			inst.Players = removeID(inst.Players, id)
			delete(w.Dungeons.byPlayer, id)
			if p := w.players[id]; p != nil {
				p.InstanceID = ""
			}
		}
	}
}

func (w *WorldState) spawnDungeonWave(inst *DungeonInstance, waveIndex int) {
	def := dungeonByID[inst.DefID]
	if waveIndex < 1 || waveIndex > len(def.EnemyWaves) {
		return
	}
	waveID := def.EnemyWaves[waveIndex-1]
	wave := dungeonWaveByID[waveID]
	inst.WaveIndex = waveIndex
	for _, sp := range wave.Spawns {
		n := sp.Count
		if n < 1 {
			continue
		}
		ed, ok := enemyDef(sp.EnemyID)
		if !ok {
			continue
		}
		for i := 0; i < n; i++ {
			ang := float64(i) / float64(n) * 6.2832
			x := sp.X + math.Cos(ang)*1.6
			z := sp.Z + math.Sin(ang)*1.6
			e := spawnEnemy(ed, x, z)
			e.NoRespawn = true
			e.InstanceID = inst.ID
			if ed.Rank == "elite" {
				e.Elite = true
			}
			if ed.Rank == "boss" {
				e.IsBoss = true
				e.HP = ed.MaxHP
				e.MaxHP = ed.MaxHP
				inst.BossID = e.ID
			}
			inst.Enemies[e.ID] = e
			applyHorizonEnemy(inst, e)
			applyInstanceScale(inst, e)
		}
	}
}

func enemyDef(id string) (EnemyDef, bool) {
	for _, d := range enemyCatalog {
		if d.ID == id {
			return d, true
		}
	}
	if b, ok := bossByID[id]; ok {
		return EnemyDef{
			ID: b.ID, Name: b.Name, Level: b.Level, MaxHP: b.MaxHP, Attack: b.Attack, Defense: b.Defense,
			Speed: b.Speed, AttackRange: 2.2, AggroRange: 16, Leash: b.Leash, AttackCooldown: 2, ExpReward: 180,
			Behavior: "boss", LootTableID: b.LootTableID, Rank: "boss",
		}, true
	}
	return EnemyDef{}, false
}

func (w *WorldState) setObjective(inst *DungeonInstance, index int) {
	objs := dungeonObjByDun[inst.DefID]
	if index < 0 || index >= len(objs) {
		return
	}
	o := objs[index]
	inst.ObjIndex = index
	inst.ObjProgress = 0
	inst.ObjNeed = o.Count
	if inst.ObjNeed < 1 {
		inst.ObjNeed = 1
	}
	if o.Type == "REACH" {
		inst.Objects = append(inst.Objects, ObjectSnapshot{ID: o.Target, Kind: "gate", X: 0, Z: 16, Text: "Gerbang Kabut"})
	}
}

func (w *WorldState) checkDungeonProgress(inst *DungeonInstance) [][]byte {
	objs := dungeonObjByDun[inst.DefID]
	if inst.ObjIndex >= len(objs) {
		return nil
	}
	o := objs[inst.ObjIndex]
	switch o.Type {
	case "KILL":
		def := dungeonByID[inst.DefID]
		if inst.ObjProgress >= inst.ObjNeed && inst.WaveIndex >= len(def.EnemyWaves)-1 && w.aliveCount(inst) == 0 {
			return w.advanceObjective(inst)
		}
		if w.aliveCount(inst) == 0 && inst.WaveIndex < len(def.EnemyWaves)-1 {
			w.spawnDungeonWave(inst, inst.WaveIndex+1)
			return [][]byte{marshal(TypeDungeonWave, w.dungeonView(inst, inst.Players[0]))}
		}
	case "REACH":
		for _, id := range inst.Players {
			p := w.players[id]
			if p != nil && p.alive() && dist2(p.X, p.Z, 0, 16) < 2.4 {
				return w.advanceObjective(inst)
			}
		}
	case "PUZZLE":
		if inst.PuzzleStep >= len(inst.PuzzleNeed) && len(inst.PuzzleNeed) > 0 {
			return w.advanceObjective(inst)
		}
	case "BOSS":
		if inst.State != DunBoss {
			return w.beginBoss(inst)
		}
	}
	return nil
}

func (w *WorldState) advanceObjective(inst *DungeonInstance) [][]byte {
	next := inst.ObjIndex + 1
	objs := dungeonObjByDun[inst.DefID]
	if next >= len(objs) {
		return nil
	}
	w.setObjective(inst, next)
	o := objs[next]
	if o.Type == "BOSS" {
		return w.beginBoss(inst)
	}
	if inst.WaveIndex < len(dungeonByID[inst.DefID].EnemyWaves)-1 && o.Type == "KILL" {
		w.spawnDungeonWave(inst, inst.WaveIndex+1)
	}
	return [][]byte{marshal(TypeDungeonObjective, w.dungeonView(inst, inst.Players[0]))}
}

func (w *WorldState) beginBoss(inst *DungeonInstance) [][]byte {
	inst.State = DunBoss
	inst.BossLocked = true
	def := dungeonByID[inst.DefID]
	wave := len(def.EnemyWaves)
	if dungeonKind(def) == "RAID" || len(def.Bosses) > 1 {
		wave = inst.EncounterIndex + 1
		if wave < 1 {
			wave = 1
		}
	}
	w.spawnDungeonWave(inst, wave)
	if b := inst.Enemies[inst.BossID]; b != nil {
		if bd, ok := bossByID[b.Def.ID]; ok && bd.EnrageTime > 0 {
			inst.EnrageAt = time.Now().Add(time.Duration(bd.EnrageTime) * time.Second)
		}
	}
	defEdu := dungeonByID[inst.DefID]
	if defEdu.EducationBoss != "" {
		inst.EduShield = true
		inst.EduQuestion = defEdu.EducationBoss
		inst.CrystalShield = true
		for _, id := range inst.Players {
			if m := w.players[id]; m != nil {
				events := w.offerDungeonQuiz(m, defEdu.EducationBoss)
				for _, ev := range events {
					w.SendToLocked(id, ev)
				}
			}
		}
	}
	view := w.dungeonView(inst, inst.Players[0])
	return [][]byte{marshal(TypeBossLock, map[string]any{"instanceId": inst.ID, "locked": true}), marshal(TypeBossSpawn, view), marshal(TypeDungeonState, view)}
}

func (w *WorldState) aliveCount(inst *DungeonInstance) int {
	n := 0
	for _, e := range inst.Enemies {
		if e.Alive && !e.IsBoss {
			n++
		}
	}
	return n
}

func (w *WorldState) onDungeonKill(p *Player, e *Enemy) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil || e == nil {
		return nil
	}
	objs := dungeonObjByDun[inst.DefID]
	if inst.ObjIndex < len(objs) {
		o := objs[inst.ObjIndex]
		if o.Type == "KILL" && (o.Target == e.Def.ID || o.Target == "") {
			inst.ObjProgress++
			if inst.ObjProgress > inst.ObjNeed {
				inst.ObjProgress = inst.ObjNeed
			}
		}
	}
	p.credit("KILL", e.Def.ID, 1)
	if e.IsBoss || e.Def.Rank == "boss" {
		def := dungeonByID[inst.DefID]
		if dungeonKind(def) == "RAID" && inst.EncounterIndex+1 < len(def.EnemyWaves) {
			inst.EncounterIndex++
			inst.BossLocked = false
			inst.BossID = ""
			inst.BossPhase = 1
			inst.Enraged = false
			inst.Telegraph = nil
			inst.State = DunActive
			return w.advanceObjective(inst)
		}
		return w.completeDungeon(inst, p)
	}
	if inst.State != DunActive {
		return [][]byte{marshal(TypeDungeonState, w.dungeonView(inst, p.ID))}
	}
	if w.aliveCount(inst) > 0 {
		return [][]byte{marshal(TypeDungeonState, w.dungeonView(inst, p.ID))}
	}
	def := dungeonByID[inst.DefID]
	if inst.WaveIndex < len(def.EnemyWaves)-1 {
		w.spawnDungeonWave(inst, inst.WaveIndex+1)
		return [][]byte{marshal(TypeDungeonWave, w.dungeonView(inst, p.ID))}
	}
	if inst.ObjIndex < len(objs) && objs[inst.ObjIndex].Type == "KILL" {
		return w.advanceObjective(inst)
	}
	return [][]byte{marshal(TypeDungeonState, w.dungeonView(inst, p.ID))}
}

func (w *WorldState) tickBoss(inst *DungeonInstance, now time.Time) [][]byte {
	boss := inst.Enemies[inst.BossID]
	if boss == nil || !boss.Alive {
		return nil
	}
	if !inst.Enraged && now.After(inst.EnrageAt) {
		inst.Enraged = true
		boss.Def.Attack = int(float64(boss.Def.Attack) * 1.25)
		boss.Def.AttackCooldown *= 0.85
		return [][]byte{marshal(TypeBossEnrage, w.dungeonView(inst, firstPlayer(inst)))}
	}
	pct := float64(boss.HP) / float64(boss.MaxHP)
	phase := bossPhaseAt(boss.Def.ID, pct)
	var events [][]byte
	if phase != inst.BossPhase {
		inst.BossPhase = phase
		inst.PhaseLockUntil = now.Add(PhaseLockDuration)
		events = append(events, marshal(TypeBossPhase, BossPhaseOut{InstanceID: inst.ID, Phase: phase, Label: phaseLabelFor(boss.Def.ID, phase)}))
	}
	if inst.Telegraph != nil {
		return events
	}
	skill := w.pickBossSkill(inst, phase, now)
	if skill.ID == "" {
		return events
	}
	target := w.closestInstancePlayer(inst, boss)
	if target == nil {
		home := math.Hypot(boss.X-boss.SX, boss.Z-boss.SZ)
		if home > boss.Def.Leash {
			w.resetBoss(inst)
			events = append(events, marshal(TypeBossReset, w.dungeonView(inst, firstPlayer(inst))))
		}
		return events
	}
	if math.Hypot(boss.X-boss.SX, boss.Z-boss.SZ) > boss.Def.Leash {
		w.resetBoss(inst)
		return append(events, marshal(TypeBossReset, w.dungeonView(inst, firstPlayer(inst))))
	}
	inst.SkillCD[skill.ID] = now.Add(time.Duration(skill.Cooldown * float64(time.Second)))
	tx, tz := target.X, target.Z
	if skill.Type == "MELEE" {
		tx, tz = boss.X, boss.Z
	}
	_, dmgM, telM := dungeonScale(inst)
	wait := skill.Telegraph * telM
	if wait < 0.35 {
		wait = 0.35
	}
	dmg := int(float64(skill.Damage) * dmgM)
	inst.Telegraph = &BossTelegraph{
		SkillID: skill.ID, X: tx, Z: tz, Radius: skill.Radius,
		Until: now.Add(time.Duration(wait * float64(time.Second))), Damage: dmg,
		Shape: skill.Shape, Yaw: math.Atan2(tx-boss.X, tz-boss.Z), Interruptible: skill.Interruptible,
	}
	if inst.Telegraph.Shape == "" {
		inst.Telegraph.Shape = "circle"
	}
	if inst.Enraged {
		inst.Telegraph.Damage = int(float64(inst.Telegraph.Damage) * 1.2)
		if inst.Telegraph.Radius > 0 {
			inst.Telegraph.Radius += 0.6
		}
	}
	if inst.CrystalShield {
		inst.Telegraph.Damage = int(float64(inst.Telegraph.Damage) * 0.7)
	}
	return append(events, marshal(TypeBossTelegraph, BossTelegraphOut{
		InstanceID: inst.ID, Skill: skill.Name, X: inst.Telegraph.X, Z: inst.Telegraph.Z,
		Radius: inst.Telegraph.Radius, Until: inst.Telegraph.Until.UnixMilli(), VFX: skill.VFX,
		Shape: inst.Telegraph.Shape, Interruptible: inst.Telegraph.Interruptible, Pulse: true,
	}))
}

func (w *WorldState) resolveTelegraph(inst *DungeonInstance) [][]byte {
	tg := inst.Telegraph
	inst.Telegraph = nil
	skill := bossSkillByID[tg.SkillID]
	var events [][]byte
	if skill.Type == "SUMMON" {
		ed, ok := enemyDef(skill.SummonID)
		if ok {
			n := skill.SummonCount
			if n < 1 {
				n = 2
			}
			boss := inst.Enemies[inst.BossID]
			bx, bz := 0.0, 20.0
			if boss != nil {
				bx, bz = boss.X, boss.Z
			}
			for i := 0; i < n; i++ {
				e := spawnEnemy(ed, bx+float64(i)*1.4-0.7, bz+1.2)
				e.NoRespawn = true
				e.InstanceID = inst.ID
				inst.Enemies[e.ID] = e
				events = append(events, marshal(TypeEnemySpawn, e.Snap()))
			}
		}
		return events
	}
	events = append(events, marshal(TypeBossAOE, BossAOEOut{InstanceID: inst.ID, X: tg.X, Z: tg.Z, Radius: math.Max(tg.Radius, 1.6), Damage: tg.Damage}))
	boss := inst.Enemies[inst.BossID]
	bx, bz, yaw := tg.X, tg.Z, tg.Yaw
	if boss != nil {
		bx, bz = boss.X, boss.Z
	}
	srcName := "Boss"
	if boss != nil {
		srcName = boss.Def.Name
	}
	for _, id := range inst.Players {
		p := w.players[id]
		if p == nil || !p.alive() {
			continue
		}
		if !inTelegraphHit(tg, p.X, p.Z, bx, bz, yaw) {
			continue
		}
		events = append(events, w.hitPlayer(inst.BossID, srcName, p, float64(tg.Damage), 0, 2.4, "boss")...)
	}
	return events
}

func inTelegraphHit(tg *BossTelegraph, px, pz, bx, bz, yaw float64) bool {
	r := tg.Radius
	if r < 1.4 {
		r = 1.8
	}
	shape := strings.ToLower(tg.Shape)
	dx, dz := px-tg.X, pz-tg.Z
	dist := math.Hypot(dx, dz)
	switch shape {
	case "cone":
		if dist > r {
			return false
		}
		ang := math.Atan2(px-bx, pz-bz) - yaw
		for ang > math.Pi {
			ang -= 2 * math.Pi
		}
		for ang < -math.Pi {
			ang += 2 * math.Pi
		}
		return math.Abs(ang) <= math.Pi/5
	case "line":
		return distToYawLine(px, pz, bx, bz, yaw, r+8) < math.Max(0.9, r)
	case "ring":
		inner := math.Max(0.6, r-1.4)
		return dist >= inner && dist <= r+0.5
	default:
		return dist <= r
	}
}

func distToYawLine(px, pz, ox, oz, yaw, length float64) float64 {
	lx := math.Sin(yaw) * length
	lz := math.Cos(yaw) * length
	dx, dz := px-ox, pz-oz
	seg := lx*lx + lz*lz
	if seg < 0.001 {
		return math.Hypot(dx, dz)
	}
	t := (dx*lx + dz*lz) / seg
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return math.Hypot(dx-lx*t, dz-lz*t)
}

func (w *WorldState) pickBossSkill(inst *DungeonInstance, phase int, now time.Time) BossSkillDef {
	def := dungeonByID[inst.DefID]
	bossID := def.BossID
	if live := inst.Enemies[inst.BossID]; live != nil {
		bossID = live.Def.ID
	}
	b := bossByID[bossID]
	var pick BossSkillDef
	for _, id := range b.Skills {
		s := bossSkillByID[id]
		if s.ID == "" || s.Phase > phase {
			continue
		}
		if until, ok := inst.SkillCD[s.ID]; ok && now.Before(until) {
			continue
		}
		pick = s
		if s.Phase == phase {
			return s
		}
	}
	return pick
}

func (w *WorldState) resetBoss(inst *DungeonInstance) {
	boss := inst.Enemies[inst.BossID]
	if boss == nil {
		return
	}
	boss.X, boss.Z = boss.SX, boss.SZ
	boss.HP = boss.MaxHP
	boss.State = "IDLE"
	inst.BossPhase = 1
	inst.Enraged = false
	inst.Telegraph = nil
	for id, e := range inst.Enemies {
		if !e.IsBoss {
			delete(inst.Enemies, id)
		}
	}
}

func (w *WorldState) closestInstancePlayer(inst *DungeonInstance, e *Enemy) *Player {
	now := time.Now()
	var taunt *Player
	for id, until := range inst.TauntUntil {
		if now.After(until) {
			continue
		}
		p := w.players[id]
		if p != nil && p.Connected && p.alive() && p.InstanceID == inst.ID {
			taunt = p
			break
		}
	}
	if taunt != nil {
		return taunt
	}
	var best *Player
	bestScore := -1.0
	bestD := e.Def.AggroRange
	if bestD <= 0 {
		bestD = 16
	}
	for _, id := range inst.Players {
		p := w.players[id]
		if p == nil || !p.Connected || !p.alive() {
			continue
		}
		d := e.Dist(p.X, p.Z)
		if d > bestD {
			continue
		}
		score := inst.Threat[id] - d
		if best == nil || score > bestScore {
			bestScore = score
			best = p
		}
	}
	return best
}

func (w *WorldState) completeDungeon(inst *DungeonInstance, killer *Player) [][]byte {
	if inst.State == DunCompleted {
		return nil
	}
	inst.State = DunCompleted
	elapsed := time.Since(inst.StartedAt).Seconds()
	deaths := 0
	for _, n := range inst.Deaths {
		deaths += n
	}
	inst.Rating = dungeonRating(elapsed, deaths)
	inst.RewardClaimID = inst.ID + ":complete"
	inst.ChestReady = true
	inst.Objects = append(inst.Objects, ObjectSnapshot{ID: "dungeon-chest", Kind: "chest", X: 0, Z: 20, Text: "Dungeon Reward Chest"})
	def := dungeonByID[inst.DefID]
	var events [][]byte
	events = append(events, marshal(TypeBossDefeated, map[string]any{"instanceId": inst.ID, "bossId": lootBossID(def)}))
	for _, id := range inst.Players {
		p := w.players[id]
		if p == nil {
			continue
		}
		p.credit("BOSS_DEFEATED", dungeonByID[inst.DefID].BossID, 1)
		if inst.ChapterID != "" {
			w.unlockChapter(p, inst.ChapterID)
		}
		events = append(events, w.markGuardianDefeated(p, dungeonByID[inst.DefID].BossID)...)
		p.ensureLog().Flags["dungeon_boss_clear"] = true
		w.applyDungeonRewards(p, inst)
		p.markDirty()
		events = append(events, w.refreshAwakening(p)...)
		loot := rollPersonalLoot(lootBossID(def), id)
		if inst.DamageDealt[id] == 0 {
			loot = filterAFKLoot(loot)
		}
		inst.Loot[id] = loot
		if log := p.ensureLog(); log != nil {
			if log.PendingLoot == nil {
				log.PendingLoot = map[string][]LootItemView{}
			}
			log.PendingLoot[inst.RewardClaimID] = loot
		}
		bonus := def.Rewards
		if w.dailyDungeonBonus(p, def.ID) {
			bonus.Exp = int(float64(bonus.Exp) * 1.2)
			bonus.Coin = int(float64(bonus.Coin) * 1.1)
		}
		events = append(events, w.giveExp(p, bonus.Exp)...)
		w.giveCurrency(p, bonus.Coin, bonus.Crystal)
		w.guildContribute(p, 10, 8)
		w.guildQuestTick(p, "DUNGEON", def.ID, 1)
		kind := "DUNGEON"
		if dungeonKind(def) == "RAID" {
			kind = "RAID"
		}
		events = append(events, w.noteActivity(p, kind, def.ID, 1)...)
		if def.ID == "dun-mistwood" && elapsed <= 900 {
			w.noteActivity(p, "DUNGEON_TIME", def.ID, int(elapsed))
		}
		if dungeonKind(def) == "RAID" && inst.WipeCount == 0 {
			w.noteActivity(p, "RAID_NOWIPE", def.ID, 1)
		}
		events = append(events, w.finishHorizon(inst, p)...)
		w.applyPhase15Rewards(p, inst, elapsed, deaths)
		w.persist(p)
		w.persistGear(p)
		events = append(events, marshal(TypeLootResult, LootResult{
			PlayerID: id, ClaimID: inst.RewardClaimID, Items: loot, Exp: bonus.Exp, Coin: bonus.Coin, Crystal: bonus.Crystal, ToID: id,
		}))
		events = append(events, marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)))
	}
	w.noteGuildDungeon(inst)
	view := w.dungeonView(inst, killer.ID)
	events = append(events, marshal(TypeDungeonComplete, view))
	return events
}

func (w *WorldState) failDungeon(inst *DungeonInstance, reason string) [][]byte {
	inst.State = DunFailed
	view := w.dungeonView(inst, firstPlayer(inst))
	view.Toast = reason
	return [][]byte{marshal(TypeDungeonFailed, view)}
}

func (w *WorldState) claimDungeonLoot(p *Player, claimID string) [][]byte {
	log := p.ensureLog()
	key := claimID + ":" + p.ID
	if claimID == "" {
		return rejectFor(p.ID, TypeClaimLoot, "claim")
	}
	if log.Claimed[key] {
		return rejectFor(p.ID, TypeClaimLoot, "duplicate")
	}
	inst := w.dungeonOf(p.ID)
	var items []LootItemView
	if inst != nil {
		if claimID != inst.RewardClaimID {
			return rejectFor(p.ID, TypeClaimLoot, "claim")
		}
		if inst.Claims[key] {
			return rejectFor(p.ID, TypeClaimLoot, "duplicate")
		}
		inst.Claims[key] = true
		items = inst.Loot[p.ID]
	} else if log.PendingLoot != nil {
		items = log.PendingLoot[claimID]
		if len(items) == 0 {
			return rejectFor(p.ID, TypeClaimLoot, "instance")
		}
	} else {
		return rejectFor(p.ID, TypeClaimLoot, "instance")
	}
	log.Claimed[key] = true
	delete(log.PendingLoot, claimID)
	for _, it := range items {
		if !w.giveItem(p, it.ItemID, it.Qty) {
			w.stashTempLoot(p, it.ItemID, it.Qty)
		}
	}
	w.persistGear(p)
	w.persist(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Hadiah dungeon diterima.", nil))}
}

func rollPersonalLoot(bossID, playerID string) []LootItemView {
	tableID := "lt-placeholder"
	if b, ok := bossByID[bossID]; ok && b.LootTableID != "" {
		tableID = b.LootTableID
	}
	table, ok := lootTableByID[tableID]
	if !ok {
		return nil
	}
	seed := time.Now().UnixNano() + int64(len(playerID))*13
	var out []LootItemView
	for i, e := range table.Entries {
		roll := float64((seed+int64(i)*97)%1000) / 1000
		if roll > e.Chance {
			continue
		}
		qty := e.MinQuantity
		if e.MaxQuantity > e.MinQuantity {
			qty += int((seed + int64(i)) % int64(e.MaxQuantity-e.MinQuantity+1))
		}
		if qty < 1 {
			continue
		}
		name, rarity := e.ItemID, "COMMON"
		if def, ok := itemByID[e.ItemID]; ok {
			name, rarity = def.Name, def.Rarity
		}
		out = append(out, LootItemView{ItemID: e.ItemID, Name: name, Qty: qty, Rarity: rarity})
	}
	return out
}

func dungeonRating(seconds float64, deaths int) string {
	if seconds <= 360 && deaths == 0 {
		return "S+"
	}
	if seconds <= 480 && deaths == 0 {
		return "S"
	}
	if seconds <= 720 && deaths <= 1 {
		return "A"
	}
	if seconds <= 1080 && deaths <= 3 {
		return "B"
	}
	return "C"
}

func lootBossID(def DungeonDef) string {
	if len(def.Bosses) > 0 {
		return def.Bosses[len(def.Bosses)-1]
	}
	return def.BossID
}

func (w *WorldState) applyDungeonRewards(p *Player, inst *DungeonInstance) {
	def := dungeonByID[inst.DefID]
	log := p.ensureLog()
	if dungeonKind(def) == "RAID" {
		if log.RaidLockout == nil {
			log.RaidLockout = map[string]string{}
		}
		log.RaidLockout[def.ID] = raidWeekKey()
		p.grantTitle("celestial-explorer")
		p.grantCosmetic("aura-celestial")
		w.giveCurrency(p, 0, 2)
		log.Flags["raid_celestial_clear"] = true
	}
	if log.WeeklyChallengeWeek != raidWeekKey() {
		log.WeeklyChallengeWeek = raidWeekKey()
		log.WeeklyChallengeID = "weekly-dungeon-clear"
		log.WeeklyChallengeProg = 0
	}
	if log.WeeklyChallengeID == "weekly-dungeon-clear" {
		log.WeeklyChallengeProg++
	}
}

func (w *WorldState) dailyDungeonBonus(p *Player, dungeonID string) bool {
	log := p.ensureLog()
	day := utcDayKey()
	if log.DailyDungeonDay == day {
		return false
	}
	log.DailyDungeonDay = day
	log.DailyDungeonID = dungeonID
	log.DailyDungeonCount++
	return true
}

func bossPhaseAt(bossID string, pct float64) int {
	b := bossByID[bossID]
	if len(b.Phases) == 0 {
		if pct <= 0.3 {
			return 3
		}
		if pct <= 0.6 {
			return 2
		}
		return 1
	}
	phase := b.Phases[0].ID
	for _, ph := range b.Phases {
		if pct <= ph.HPPct+0.0001 {
			phase = ph.ID
		}
	}
	return phase
}

func phaseLabelFor(bossID string, phase int) string {
	b := bossByID[bossID]
	for _, ph := range b.Phases {
		if ph.ID == phase && ph.Label != "" {
			return ph.Label
		}
	}
	return phaseLabel(phase)
}

func phaseLabel(p int) string {
	if p >= 3 {
		return "ENRAGED"
	}
	if p == 2 {
		return "PHASE 2"
	}
	return "PHASE 1"
}

func firstPlayer(inst *DungeonInstance) string {
	if len(inst.Players) == 0 {
		return ""
	}
	return inst.Players[0]
}

func (w *WorldState) dungeonView(inst *DungeonInstance, watcher string) DungeonView {
	def := dungeonByID[inst.DefID]
	ch := chapterByID[inst.ChapterID]
	alive := w.aliveCount(inst)
	objText, objType := "", ""
	objs := dungeonObjByDun[inst.DefID]
	if inst.ObjIndex < len(objs) {
		objText, objType = objs[inst.ObjIndex].Text, objs[inst.ObjIndex].Type
	}
	left := int(time.Until(inst.EndsAt).Seconds())
	if left < 0 {
		left = 0
	}
	view := DungeonView{
		InstanceID: inst.ID, DungeonID: inst.DefID, ChapterID: inst.ChapterID, Kind: dungeonKind(def), Name: def.Name,
		Title: titleOf(ch, def), State: inst.State, Wave: inst.WaveIndex, WaveTotal: len(def.EnemyWaves),
		Encounter: inst.EncounterIndex, Enemies: alive, Objective: objText, ObjectiveType: objType,
		Progress: inst.ObjProgress, Count: inst.ObjNeed,
		TimeLeft: left, Rating: inst.Rating, ClaimID: inst.RewardClaimID, Chest: inst.ChestReady,
		Elapsed: int(time.Since(inst.StartedAt).Seconds()), BossLocked: inst.BossLocked,
		CrystalShield: inst.CrystalShield, PuzzleStep: inst.PuzzleStep, WipeCount: inst.WipeCount,
		Synergy: partySynergy(inst), Votes: inst.Votes, Difficulty: inst.Difficulty,
		Mechanic: inst.Mechanic, GuideHP: inst.GuideHP, EduShield: inst.EduShield,
	}
	if inst.State == DunStarting {
		view.State = DunLoading
	}
	if def.WeeklyLockout {
		view.LockoutLabel = raidLockoutLabel(time.Now())
	}
	if boss := inst.Enemies[inst.BossID]; boss != nil {
		bdef := bossByID[boss.Def.ID]
		if bdef.ID == "" {
			bdef = bossByID[def.BossID]
		}
		view.Boss = &BossView{
			ID: boss.Def.ID, Name: bdef.Name, Title: bdef.Title, Level: bdef.Level,
			HP: boss.HP, MaxHP: boss.MaxHP, Phase: inst.BossPhase, Enraged: inst.Enraged, Alive: boss.Alive,
		}
	}
	watcherP := w.players[watcher]
	for _, id := range inst.Players {
		p := w.players[id]
		if p == nil {
			continue
		}
		d := 0.0
		if watcherP != nil {
			d = dist2(p.X, p.Z, watcherP.X, watcherP.Z)
		}
		view.Members = append(view.Members, DungeonMemberView{
			PlayerID: p.ID, Name: p.Name, Level: p.Level, HP: p.HP, MaxHP: p.MaxHP,
			Dead: !p.alive(), Downed: p.CombatState == "DOWNED", Online: p.Connected, Distance: d,
			Role: inst.Roles[id], ReviveProgress: int(inst.ReviveProgress[id]), ReviveToken: inst.ReviveToken[id],
			Energy: p.Energy, MaxEnergy: p.MaxEnergy,
		})
	}
	view.Loot = inst.Loot[watcher]
	view.Room = dungeonRoomLabel(inst)
	if view.Room != "" {
		view.Checkpoint = "CHECKPOINT"
	}
	return view
}

func titleOf(ch ChapterDef, def DungeonDef) string {
	if ch.Title != "" {
		return ch.Title
	}
	return def.Name
}

func partySynergy(inst *DungeonInstance) bool {
	has := map[string]bool{}
	for _, r := range inst.Roles {
		has[normalizeRole(r)] = true
	}
	return has["TANK"] && has["DPS"] && has["SUPPORT"]
}

func chapterStatus(p *Player, ch ChapterDef) string {
	if ch.ID == "" {
		return "AVAILABLE"
	}
	log := p.ensureLog()
	if log.Flags["chapter_"+ch.ID+"_complete"] || (ch.ID == "ch01" && log.Flags["chapter1_complete"]) {
		return "COMPLETED"
	}
	if ch.RequiredChapter == "" {
		return "AVAILABLE"
	}
	if log.Flags["chapter_"+ch.RequiredChapter+"_complete"] || (ch.RequiredChapter == "ch01" && log.Flags["chapter1_complete"]) {
		return "AVAILABLE"
	}
	return "LOCKED"
}

func (w *WorldState) unlockChapter(p *Player, completedID string) {
	if completedID == "" {
		return
	}
	log := p.ensureLog()
	log.Flags["chapter_"+completedID+"_complete"] = true
	if completedID == "ch01" {
		log.Flags["chapter1_complete"] = true
		log.Flags["chapter2_unlocked"] = true
	}
	ch := chapterByID[completedID]
	for _, next := range chapterCatalog {
		if next.RequiredChapter == completedID {
			log.Flags["chapter_"+next.ID+"_available"] = true
		}
	}
	if completedID == "ch01" || completedID == "ch03" || completedID == "ch05" || completedID == "ch08" {
		log.Flags["education_"+completedID] = true
	}
	_ = ch
	w.persist(p)
}

func (w *WorldState) chapterList(p *Player) ChapterListOut {
	list := make([]ChapterView, 0, len(chapterCatalog))
	for _, c := range chapterCatalog {
		st := chapterStatus(p, c)
		list = append(list, ChapterView{
			ID: c.ID, Title: c.Title, Region: c.Region, Story: c.Story, BossID: c.BossID, BossName: c.BossName,
			RequiredLevel: c.RequiredLevel, Reward: rewardView(c.Reward, false), Status: st, DungeonID: c.DungeonID,
		})
	}
	return ChapterListOut{Chapters: list}
}

func playerChapterViews(p *Player) []ChapterView {
	return (&WorldState{}).chapterList(p).Chapters
}
