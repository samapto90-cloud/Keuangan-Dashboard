package mmo

import (
	"encoding/json"
	"log"
	"math"
	"strings"
	"time"
)

func combatLog(format string, args ...any) {
	log.Printf("[COMBAT] "+format, args...)
}

func (p *Player) initCombat() {
	if p.MaxHP <= 0 {
		p.MaxHP = BaseMaxHP
	}
	dead := p.CombatState == "DEAD" || p.CombatState == "DOWNED" || p.CombatState == "RESPAWNING"
	if p.HP <= 0 && !dead {
		p.HP = p.MaxHP
	}
	if p.MaxEnergy <= 0 {
		p.MaxEnergy = EnergyMax
	}
	if p.Energy <= 0 {
		p.Energy = p.MaxEnergy
	}
	if p.MaxStamina <= 0 {
		p.MaxStamina = StaminaMax
	}
	if p.Stamina <= 0 {
		p.Stamina = p.MaxStamina
	}
	if p.BaseMaxHP <= 0 {
		p.BaseMaxHP = BaseMaxHP
	}
	if p.BaseMaxEnergy <= 0 {
		p.BaseMaxEnergy = EnergyMax
	}
	if p.BaseStrength <= 0 {
		p.BaseStrength = BaseStrength
	}
	if p.BaseDefense <= 0 {
		p.BaseDefense = BaseDefense
	}
	if p.Level <= 0 {
		p.Level = 1
	}
	if p.ExpToNext <= 0 {
		p.ExpToNext = expToNext(p.Level)
	}
	if p.SkillCD == nil {
		p.SkillCD = map[string]time.Time{}
	}
	if p.CombatState == "" {
		p.CombatState = "IDLE"
	}
	if p.FormID == "" {
		p.FormID = "normal"
	}
	if p.TransformState == "" {
		p.TransformState = "NORMAL"
	}
}

func (p *Player) alive() bool {
	return p.HP > 0 && p.CombatState != "DEAD" && p.CombatState != "RESPAWNING"
}

func (p *Player) invulnerable(now time.Time) bool {
	return now.Before(p.IFrameUntil) || now.Before(p.InvulnUntil)
}

func (p *Player) canAct(now time.Time) string {
	if !p.alive() {
		return "player mati"
	}
	if now.Before(p.PvpNoActUntil) {
		return "countdown"
	}
	if now.Before(p.StunUntil) {
		return "stun"
	}
	if now.Before(p.SilenceUntil) {
		return "silence"
	}
	if now.Before(p.DodgeUntil) || p.CombatState == "DODGING" {
		return "dodge"
	}
	if now.Before(p.HitUntil) {
		return "hit"
	}
	return ""
}

func validTimestamp(ts int64, now time.Time) bool {
	if ts <= 0 {
		return true
	}
	delta := now.UnixMilli() - ts
	if delta < 0 {
		delta = -delta
	}
	return time.Duration(delta)*time.Millisecond <= TimestampMaxAge
}

func (w *WorldState) ApplyCombat(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	p.LastInputAt = time.Now()
	if w.limited("combat:"+p.ID, CombatRatePerSec, time.Second) {
		return rejectFor(p.ID, env.Type, "rate")
	}
	var events [][]byte
	if p.Mounted {
		events = append(events, w.dismount(p, "combat")...)
	}
	switch env.Type {
	case TypePlayerAttack, TypePlayerCombo:
		var in AttackIn
		if unmarshal(env.Data, &in) != nil {
			events = rejectFor(p.ID, "PLAYER_ATTACK", "payload")
		} else {
			events = w.playerAttack(p, in)
		}
	case TypePlayerSkill:
		var in SkillIn
		if unmarshal(env.Data, &in) != nil {
			events = rejectFor(p.ID, "PLAYER_SKILL", "payload")
		} else {
			events = w.playerSkill(p, in)
		}
	case TypePlayerDodge:
		var in DodgeIn
		if unmarshal(env.Data, &in) != nil {
			events = rejectFor(p.ID, "PLAYER_DODGE", "payload")
		} else {
			events = w.playerDodge(p, in)
		}
	case TypePlayerRespawn:
		events = w.playerRespawn(p, false)
	case TypePlayerBlock:
		var in BlockIn
		_ = unmarshal(env.Data, &in)
		events = w.playerBlock(p, in.On)
	case TypePlayerCounter:
		events = w.playerCounter(p)
	case TypePlayerCharge, TypeSetEnergy, TypeSetCooldown, TypeUnlockTransform, TypeSetDamage:
		events = w.applyPhase25Combat(p, env)
	default:
		return nil
	}
	return events
}

func unmarshal(raw []byte, dest any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dest)
}

func rejectFor(playerID, action, reason string) [][]byte {
	return [][]byte{marshal(TypeActionReject, RejectOut{Action: action, Reason: reason, PlayerID: playerID})}
}

func reject(action, reason string) [][]byte {
	return rejectFor("", action, reason)
}

func (w *WorldState) playerAttack(p *Player, in AttackIn) [][]byte {
	now := time.Now()
	if !validTimestamp(in.Timestamp, now) {
		return rejectFor(p.ID, "PLAYER_ATTACK", "timestamp")
	}
	if reason := p.canAct(now); reason != "" {
		return rejectFor(p.ID, "PLAYER_ATTACK", reason)
	}
	if now.Before(p.AttackCDUntil) {
		return rejectFor(p.ID, "PLAYER_ATTACK", "cooldown")
	}
	kind := normalizeAttack(in.AttackType)
	if kind == "" {
		return rejectFor(p.ID, "PLAYER_ATTACK", "attackType")
	}
	if p.InstanceID == "" && p.inCombatDisabledZone() {
		return rejectFor(p.ID, "PLAYER_ATTACK", "safe_zone")
	}
	if OpenWorldPvP == false && p.InstanceID == "" {
		if _, isPlayer := w.players[in.TargetID]; isPlayer {
			return rejectFor(p.ID, "PLAYER_ATTACK", "pvp_disabled")
		}
		if p.inOpenWorldSafeZone() {
			if _, isPlayer := w.players[in.TargetID]; isPlayer {
				return rejectFor(p.ID, "PLAYER_ATTACK", "safe_zone")
			}
		}
	}
	if in.Direction != 0 {
		p.Yaw = in.Direction
	}
	nowWin := comboWindow()
	letter := "L"
	if kind == "heavy" {
		letter = "H"
	}
	if now.After(p.ComboUntil) {
		p.ComboChain = ""
	}
	next := p.ComboChain + letter
	recipe := comboRecipe()
	finisher := next == recipe
	advanced := next == "LLH"
	if !finisher && !advanced && !comboPrefix(next, recipe) {
		next = letter
	}
	if len(next) > comboCfg.MaxHits {
		next = letter
	}
	p.ComboChain = next
	p.ComboUntil = now.Add(nowWin)
	hits := len(next)
	base, rng, kb, cd := comboNumbers(kind, hits, finisher)
	if p.hasSkill("enhanced_punch") && kind == "light" {
		base += 2
	}
	power := 1.0
	if p.hasSkill("combo_master") && hits >= 2 {
		power *= 1.05
	}
	if finisher && p.hasSkill("finisher_mastery") {
		power *= 1.15
	}
	rng += p.AttackRange
	cd = agilityCooldown(p, cd)
	if inst := w.pvpOf(p.ID); inst != nil {
		if inst.State != PvpActive || now.Before(inst.CountdownUntil) {
			return rejectFor(p.ID, "PLAYER_ATTACK", "countdown")
		}
		p.AttackCDUntil = now.Add(cd)
		p.CombatState = "ATTACKING"
		if hits > 1 {
			p.CombatState = "COMBO"
		}
		p.InCombatUntil = now.Add(CombatIdleAfter)
		display := kind
		if finisher {
			display = "finisher"
			p.ComboChain = ""
		}
		w.phase25OnAttack(p)
		return w.pvpAttack(p, inst, in, base, power, rng, kb, kind, hits, finisher, display)
	}
	target := w.resolveEnemyTarget(p, in.TargetID, rng+RangeSlack)
	if target == nil {
		return rejectFor(p.ID, "PLAYER_ATTACK", "target")
	}
	p.AttackCDUntil = now.Add(cd)
	p.CombatState = "ATTACKING"
	if hits > 1 {
		p.CombatState = "COMBO"
	}
	p.InCombatUntil = now.Add(CombatIdleAfter)
	display := kind
	if finisher {
		display = "finisher"
		p.ComboChain = ""
	}
	w.phase25OnAttack(p)
	events := [][]byte{marshal(TypeAttackResult, AttackResult{
		AttackerID: p.ID, AttackType: display, TargetIDs: []string{target.ID}, Timestamp: now.UnixMilli(),
		ComboHits: hits, Finisher: finisher,
	})}
	events = append(events, w.hitEnemy(p, target, float64(base), power, kb, kind)...)
	return events
}

func (w *WorldState) playerSkill(p *Player, in SkillIn) [][]byte {
	now := time.Now()
	if !validTimestamp(in.Timestamp, now) {
		return rejectFor(p.ID, "PLAYER_SKILL", "timestamp")
	}
	if reason := p.canAct(now); reason != "" {
		return rejectFor(p.ID, "PLAYER_SKILL", reason)
	}
	if in.SkillID = resolvePhase25Skill(in.SkillID); in.SkillID == "" {
		return rejectFor(p.ID, "PLAYER_SKILL", "skill")
	}
	def, ok := skillCatalog[in.SkillID]
	if !ok {
		return rejectFor(p.ID, "PLAYER_SKILL", "skill")
	}
	if def.Kind == "PASSIVE" || def.Kind == "TRANSFORMATION" {
		return rejectFor(p.ID, "PLAYER_SKILL", "passive")
	}
	if !p.hasSkill(def.ID) {
		return rejectFor(p.ID, "PLAYER_SKILL", "locked")
	}
	if p.Level < def.RequiredLevel {
		return rejectFor(p.ID, "PLAYER_SKILL", "level")
	}
	if def.RequiredSkill != "" && !p.hasSkill(def.RequiredSkill) {
		return rejectFor(p.ID, "PLAYER_SKILL", "prerequisite")
	}
	if until, exists := p.SkillCD[def.ID]; exists && now.Before(until) {
		return rejectFor(p.ID, "PLAYER_SKILL", "cooldown")
	}
	if def.Kind == "ULTIMATE" && p.Energy < p.MaxEnergy {
		return rejectFor(p.ID, "PLAYER_SKILL", "energy")
	}
	if p.Energy < def.EnergyCost {
		return rejectFor(p.ID, "PLAYER_SKILL", "energy")
	}
	if inst := w.pvpOf(p.ID); inst != nil && pvpSkillLocked(def.ID) {
		return rejectFor(p.ID, "PLAYER_SKILL", "disabled")
	}
	if p.Stamina < def.StaminaCost {
		return rejectFor(p.ID, "PLAYER_SKILL", "stamina")
	}
	if in.Direction != 0 {
		p.Yaw = in.Direction
	}
	if inst := w.pvpOf(p.ID); inst != nil {
		if inst.State != PvpActive || now.Before(inst.CountdownUntil) {
			return rejectFor(p.ID, "PLAYER_SKILL", "countdown")
		}
		p.Energy -= def.EnergyCost
		p.Stamina -= def.StaminaCost
		p.SkillCD[def.ID] = now.Add(agilityCooldown(p, time.Duration(def.Cooldown*float64(time.Second))))
		p.CombatState = "CASTING"
		if def.Type == "ENERGY" {
			p.CombatState = "ENERGY_ATTACK"
		}
		p.InCombatUntil = now.Add(CombatIdleAfter)
		w.applySkillEffect(p, def)
		w.phase25OnSkill(p, def.ID)
		power := 1.0
		if def.Type == "ENERGY" {
			power += float64(p.EnergyPower) * progressionCfg.EnergyDamagePerEnergy
		}
		if def.Kind == "ULTIMATE" {
			p.Energy = 0
		}
		return w.pvpSkill(p, inst, def, in, power)
	}
	self := def.Target == "SELF" || def.Effect == "dash" || def.Effect == "guard" || def.Damage <= 0
	var targets []*Enemy
	if !self {
		if def.Radius > 0 || def.Target == "AREA" || def.Target == "DIRECTION" {
			r := def.Radius
			if r < 1 {
				r = def.Range
			}
			targets = w.enemiesInRadiusFor(p, r+RangeSlack)
		} else {
			maxR := def.Range + RangeSlack
			if def.Range > 8 {
				maxR = def.Range
			}
			t := w.resolveEnemyTarget(p, in.TargetID, maxR)
			if t != nil {
				targets = []*Enemy{t}
			}
		}
		if len(targets) == 0 {
			return rejectFor(p.ID, "PLAYER_SKILL", "target")
		}
	}
	p.Energy -= def.EnergyCost
	p.Stamina -= def.StaminaCost
	if def.Kind == "ULTIMATE" {
		p.UltCharge = 0
	}
	if p.ComboMasteryReady {
		p.ensureLog().Flags["advanced_combo"] = true
		p.ensureLog().StyleMastery[p.ensureStyle()] += 8
		p.ComboMasteryReady = false
	}
	p.SkillCD[def.ID] = now.Add(agilityCooldown(p, time.Duration(def.Cooldown*float64(time.Second))))
	p.CombatState = "CASTING"
	if def.Type == "ENERGY" {
		p.CombatState = "ENERGY_ATTACK"
	}
	if def.Kind == "ULTIMATE" {
		p.CombatState = "ULTIMATE"
	}
	p.InCombatUntil = now.Add(CombatIdleAfter)
	w.applySkillEffect(p, def)
	w.phase25OnSkill(p, def.ID)
	ids := make([]string, 0, len(targets))
	for _, t := range targets {
		ids = append(ids, t.ID)
	}
	events := [][]byte{
		marshal(TypeAttackResult, AttackResult{AttackerID: p.ID, AttackType: "skill", SkillID: def.ID, TargetIDs: ids, Timestamp: now.UnixMilli()}),
		marshal(TypeSkillUsed, map[string]any{"playerId": p.ID, "skillId": def.ID}),
	}
	power := 1.0
	if def.Type == "ENERGY" {
		power += float64(p.EnergyPower) * progressionCfg.EnergyDamagePerEnergy
	}
	for _, t := range targets {
		atk := def.Type
		if def.Element != "" {
			atk = def.Element
		}
		events = append(events, w.hitEnemy(p, t, float64(def.Damage), power, 3.2, atk)...)
	}
	return events
}

func (w *WorldState) playerDodge(p *Player, in DodgeIn) [][]byte {
	now := time.Now()
	if !validTimestamp(in.Timestamp, now) {
		return rejectFor(p.ID, "PLAYER_DODGE", "timestamp")
	}
	if !p.alive() {
		return rejectFor(p.ID, "PLAYER_DODGE", "dead")
	}
	if now.Before(p.DodgeCDUntil) {
		return rejectFor(p.ID, "PLAYER_DODGE", "cooldown")
	}
	if now.Before(p.StunUntil) {
		return rejectFor(p.ID, "PLAYER_DODGE", "stun")
	}
	if p.Stamina < DodgeStamina {
		return rejectFor(p.ID, "PLAYER_DODGE", "stamina")
	}
	p.Stamina -= DodgeStamina
	yaw := in.Yaw
	if math.Hypot(in.AX, in.AZ) > 0.12 {
		yaw = headingFromCamera(in.Yaw, in.AX, in.AZ)
	} else if yaw == 0 {
		yaw = p.Yaw
	}
	p.Yaw = yaw
	p.VX = math.Sin(yaw) * DodgeSpeed
	p.VZ = math.Cos(yaw) * DodgeSpeed
	p.CombatState = "DODGING"
	p.DodgeUntil = now.Add(DodgeDuration)
	p.IFrameUntil = now.Add(DodgeIFrame)
	p.PerfectDodgeUntil = now.Add(PerfectDodgeWindow)
	p.DodgeCDUntil = now.Add(DodgeCooldown)
	w.phase25OnDodge(p)
	return [][]byte{marshal(TypeAttackResult, AttackResult{
		AttackerID: p.ID,
		AttackType: "dodge",
		TargetIDs:  []string{},
		Timestamp:  now.UnixMilli(),
	})}
}

func (w *WorldState) resolveEnemyTarget(p *Player, targetID string, maxDist float64) *Enemy {
	if targetID != "" {
		if _, isPlayer := w.players[targetID]; isPlayer {
			return nil
		}
		e := w.enemyByID(targetID)
		if e == nil || !e.Alive || e.HP <= 0 {
			return nil
		}
		if p.InstanceID != e.InstanceID {
			return nil
		}
		if e.Dist(p.X, p.Z) <= maxDist {
			return e
		}
		return nil
	}
	return w.nearestEnemyFor(p, maxDist)
}

func (w *WorldState) hitEnemy(p *Player, e *Enemy, base, skillPower, knockback float64, attackType string) [][]byte {
	now := time.Now()
	equip := EquipmentBonus + float64(p.EquipAttack)*0.04
	chance := CritChance
	if p.CritChance > 0 {
		chance = p.CritChance
	}
	dealt, crit := calcDamageChance(base, float64(p.Strength), skillPower, equip, float64(e.Def.Defense), chance)
	el := skillElement(attackType, p)
	dealt = int(float64(dealt)*elementBonus(el) + 0.5)
	if e.statusActive("WEAKEN", now) {
		dealt = int(float64(dealt)*1.08 + 0.5)
	}
	if dealt > MaxDamage {
		dealt = MaxDamage
	}
	if e.HP <= 0 {
		return nil
	}
	if g, ok := guardianByID[e.Def.ID]; ok && len(g.EducationPool) > 0 {
		if !p.ensureLog().Flags["edu_boss_"+g.ID] {
			dealt = dealt / 4
			if dealt < 1 {
				dealt = 1
			}
		}
	}
	if w.Boss != nil && w.Boss.State == "ACTIVE" && e.ID == w.Boss.EnemyID {
		_, name := w.Boss.phase()
		if name == "SHIELD" && !p.ensureLog().Flags["wb_shield_down"] {
			dealt = dealt / 3
			if dealt < 1 {
				dealt = 1
			}
		}
	}
	var extra [][]byte
	if inst := w.dungeonOf(p.ID); inst != nil {
		if inst.Threat == nil {
			inst.Threat = map[string]float64{}
		}
		if inst.DamageDealt == nil {
			inst.DamageDealt = map[string]int{}
		}
		if inst.Suspicious == nil {
			inst.Suspicious = map[string]int{}
		}
		if partySynergy(inst) {
			dealt = int(float64(dealt) * 1.05)
		}
		if inst.LastHitID != "" && inst.LastHitID != p.ID && time.Since(inst.LastHitAt) < 1500*time.Millisecond {
			dealt = int(float64(dealt) * 1.05)
		}
		if inst.EduShield && e.IsBoss {
			return extra
		}
		if !inst.PhaseLockUntil.IsZero() && now.Before(inst.PhaseLockUntil) && e.IsBoss {
			return extra
		}
		if dealt < 1 {
			dealt = 1
		}
		add := float64(dealt)
		if inst.Roles[p.ID] == "TANK" {
			add *= 2
		}
		inst.Threat[p.ID] += add
		inst.DamageDealt[p.ID] += dealt
		if dealt > p.Strength*12+80 {
			inst.Suspicious[p.ID]++
		}
		inst.LastHitID = p.ID
		inst.LastHitAt = now
		if inst.Telegraph != nil && inst.Telegraph.Interruptible {
			inst.Telegraph = nil
			extra = append(extra, marshal(TypeBossInterrupt, map[string]any{"instanceId": inst.ID, "playerId": p.ID}))
		}
	}
	e.HP -= dealt
	if e.HP < 0 {
		e.HP = 0
	}
	w.noteWorldBossHit(p, e, dealt)
	e.LastHitBy = p.ID
	e.State = "HIT"
	if knockback > 0 {
		dx := e.X - p.X
		dz := e.Z - p.Z
		if math.Hypot(dx, dz) < 0.01 {
			dx = math.Sin(p.Yaw)
			dz = math.Cos(p.Yaw)
		} else {
			n := math.Hypot(dx, dz)
			dx /= n
			dz /= n
		}
		e.VX += dx * knockback
		e.VZ += dz * knockback
	}
	killed := e.HP <= 0
	events := extra
	events = append(events,
		marshal(TypeDamageResult, DamageResult{
			AttackerID:  p.ID,
			TargetID:    e.ID,
			Damage:      dealt,
			IsCritical:  crit,
			HitX:        e.X,
			HitY:        e.Y + 1.2,
			HitZ:        e.Z,
			AttackType:  attackType,
			Timestamp:   now.UnixMilli(),
			TargetHP:    e.HP,
			TargetMaxHP: e.MaxHP,
			Killed:      killed,
			Kind:        "enemy",
		}),
		marshal(TypeEnemyHit, DamageResult{
			AttackerID:  p.ID,
			TargetID:    e.ID,
			Damage:      dealt,
			IsCritical:  crit,
			HitX:        e.X,
			HitY:        e.Y + 1.2,
			HitZ:        e.Z,
			AttackType:  attackType,
			Timestamp:   now.UnixMilli(),
			TargetHP:    e.HP,
			TargetMaxHP: e.MaxHP,
			Killed:      killed,
			Kind:        "enemy",
		}),
	)
	combatLog("%s attacked %s Damage: %d Critical: %v Target HP: %d/%d", p.Name, e.Def.Name, dealt, crit, e.HP, e.MaxHP)
	w.afterCombatHit(p, e, dealt, crit, attackType)
	if e.Def.ID == "training_dummy" && p.ensureLog().Training.Hits > 0 {
		tr := p.ensureLog().Training
		events = append(events, marshal(TypeTrainingMeter, map[string]any{
			"playerId": p.ID, "dps": tr.DPS, "hits": tr.Hits, "damage": tr.Damage, "combo": len(p.ComboChain), "toId": p.ID,
		}))
	}
	if killed {
		events = append(events, w.killEnemy(p, e)...)
	}
	return events
}

func (w *WorldState) killEnemy(p *Player, e *Enemy) [][]byte {
	if !e.Alive {
		return nil
	}
	e.Alive = false
	e.HP = 0
	e.State = "DEAD"
	e.TargetID = ""
	e.DeadUntil = time.Now().Add(EnemyRespawnDelay)
	exp := e.Def.ExpReward
	if e.Def.ID == "training_dummy" {
		exp = 0
	}
	events := [][]byte{marshal(TypeEnemyDeath, EnemyDeathOut{EnemyID: e.ID, KillerID: p.ID, Exp: exp})}
	combatLog("%s killed %s EXP: %d", p.Name, e.Def.Name, exp)
	events = append(events, w.partyShareExp(p, exp)...)
	w.partyShareKill(p, e.Def.ID)
	w.guildQuestTick(p, "KILL", e.Def.ID, 1)
	w.spawnDrop(p, e)
	log := p.ensureLog()
	log.Flags["killed_"+e.Def.ID] = true
	if e.Elite || e.Def.Rank == "ELITE" || e.Def.ID == "elite_shadow_beast" {
		log.Flags["killed_elite_shadow_beast"] = true
	}
	p.markDirty()
	events = append(events, w.refreshAwakening(p)...)
	w.addEventScore(p, 8)
	if g, ok := guardianByID[e.Def.ID]; ok {
		events = append(events, w.markGuardianDefeated(p, g.ID)...)
	}
	if p.InstanceID != "" || e.InstanceID != "" {
		events = append(events, w.onDungeonKill(p, e)...)
	}
	return events
}

func (w *WorldState) hitPlayer(srcID, srcName string, p *Player, base, defense, knockback float64, attackType string) [][]byte {
	now := time.Now()
	if !p.alive() || p.invulnerable(now) {
		if now.Before(p.PerfectDodgeUntil) {
			p.Energy = min(p.MaxEnergy, p.Energy+4)
			p.UltCharge += 6
			if p.UltCharge > MaxUltCharge {
				p.UltCharge = MaxUltCharge
			}
			w.phase25OnPerfectDodge(p)
			return [][]byte{marshal(TypeAttackResult, AttackResult{AttackerID: p.ID, AttackType: "perfect_dodge", Timestamp: now.UnixMilli()})}
		}
		return nil
	}
	resist := float64(p.Defense) + defense
	if attackType == "ENERGY" || attackType == "energy" {
		resist += float64(p.EnergyPower) * progressionCfg.DefensePerDefense
	}
	dealt, crit := calcDamage(base, 1, 1, EquipmentBonus, resist)
	blocked := false
	if p.Blocking || now.Before(p.GuardUntil) {
		dealt = int(float64(dealt)*BlockDamageFactor + 0.5)
		blocked = true
		if p.ensureStyle() == StyleIronGuard {
			dealt = int(float64(dealt)*0.85 + 0.5)
		}
	}
	if dealt > MaxDamage {
		dealt = MaxDamage
	}
	if dealt < MinDamage {
		dealt = MinDamage
	}
	p.HP -= dealt
	if p.HP < 0 {
		p.HP = 0
	}
	p.stopCharge()
	p.Stagger += dealt / 4
	if p.StaggerMax < 1 {
		p.StaggerMax = 80 + p.SpentDEF*4
	}
	if p.Stagger >= p.StaggerMax {
		p.Stagger = 0
		p.StunUntil = now.Add(400 * time.Millisecond)
		p.CombatState = "STUNNED"
	} else {
		p.CombatState = "HIT"
	}
	p.HitUntil = now.Add(HitStun)
	p.InCombatUntil = now.Add(CombatIdleAfter)
	if knockback > 0 {
		p.VX += math.Sin(p.Yaw+math.Pi) * knockback * 0.15
		p.VZ += math.Cos(p.Yaw+math.Pi) * knockback * 0.15
	}
	killed := p.HP <= 0
	events := [][]byte{
		marshal(TypeDamageResult, DamageResult{
			AttackerID:  srcID,
			TargetID:    p.ID,
			Damage:      dealt,
			IsCritical:  crit,
			Blocked:     blocked,
			HitX:        p.X,
			HitY:        p.Y + 1.4,
			HitZ:        p.Z,
			AttackType:  attackType,
			Timestamp:   now.UnixMilli(),
			TargetHP:    p.HP,
			TargetMaxHP: p.MaxHP,
			Killed:      killed,
			Kind:        "player",
		}),
		marshal(TypePlayerHit, DamageResult{
			AttackerID:  srcID,
			TargetID:    p.ID,
			Damage:      dealt,
			IsCritical:  crit,
			Blocked:     blocked,
			HitX:        p.X,
			HitY:        p.Y + 1.4,
			HitZ:        p.Z,
			AttackType:  attackType,
			Timestamp:   now.UnixMilli(),
			TargetHP:    p.HP,
			TargetMaxHP: p.MaxHP,
			Killed:      killed,
			Kind:        "player",
		}),
	}
	combatLog("%s hit %s Damage: %d Critical: %v Target HP: %d/%d", srcName, p.Name, dealt, crit, p.HP, p.MaxHP)
	if killed {
		events = append(events, w.killPlayer(p)...)
	}
	return events
}

func (w *WorldState) killPlayer(p *Player) [][]byte {
	now := time.Now()
	p.HP = 0
	log := p.ensureLog()
	log.DeathlessKills = 0
	for _, d := range challengeCat {
		if d.Metric == "KILL_NODEATH" && !log.ChallengeClaimed[d.ID] {
			log.ChallengeProgress[d.ID] = 0
		}
	}
	p.AX, p.AZ = 0, 0
	p.VX, p.VZ = 0, 0
	if inst := w.pvpOf(p.ID); inst != nil {
		def, _ := pvpMode(inst.Mode)
		if def.Respawn {
			p.CombatState = "RESPAWNING"
			p.State = "RESPAWNING"
			d := time.Duration(pvpMod.BgRespawn) * time.Second
			if d <= 0 {
				d = 5 * time.Second
			}
			p.RespawnAt = now.Add(d)
		} else {
			p.CombatState = "DEAD"
			p.State = "DEAD"
			p.RespawnAt = now.Add(24 * time.Hour)
		}
		events := [][]byte{marshal(TypePlayerDeath, DeathOut{PlayerID: p.ID, RespawnAt: p.RespawnAt.UnixMilli()})}
		if p.FormID != "" && p.FormID != "normal" {
			events = append(events, w.endTransform(p, "death")...)
		}
		return events
	}
	if p.InstanceID != "" {
		p.CombatState = "DOWNED"
		p.State = "DOWNED"
		p.RespawnAt = now.Add(DungeonDownedWindow)
		if inst := w.dungeonOf(p.ID); inst != nil {
			inst.Deaths[p.ID]++
			if inst.DownedAt == nil {
				inst.DownedAt = map[string]time.Time{}
			}
			inst.DownedAt[p.ID] = now
		}
	} else {
		p.CombatState = "DEAD"
		p.State = "DEAD"
		p.RespawnAt = now.Add(RespawnDelay)
	}
	if p.Energy < 0 {
		p.Energy = 0
	}
	var events [][]byte
	if p.FormID != "" && p.FormID != "normal" {
		events = append(events, w.endTransform(p, "death")...)
	}
	if p.CombatState == "DOWNED" {
		events = append(events, marshal(TypePlayerDowned, map[string]any{"playerId": p.ID, "until": p.RespawnAt.UnixMilli()}))
	}
	combatLog("%s died respawn in %.0fs", p.Name, time.Until(p.RespawnAt).Seconds())
	events = append(events, marshal(TypePlayerDeath, DeathOut{PlayerID: p.ID, RespawnAt: p.RespawnAt.UnixMilli()}))
	return events
}

func (w *WorldState) playerRespawn(p *Player, auto bool) [][]byte {
	now := time.Now()
	if p.CombatState != "DEAD" && p.CombatState != "RESPAWNING" {
		if !auto {
			return rejectFor(p.ID, "PLAYER_RESPAWN", "not-dead")
		}
		return nil
	}
	if p.InstanceID != "" {
		if w.pvpOf(p.ID) != nil {
			return rejectFor(p.ID, "PLAYER_RESPAWN", "pvp")
		}
		if !auto {
			return rejectFor(p.ID, "PLAYER_RESPAWN", "downed")
		}
		return nil
	}
	if now.Before(p.RespawnAt) {
		if !auto {
			return rejectFor(p.ID, "PLAYER_RESPAWN", "timer")
		}
		return nil
	}
	pt := spawnPoints[w.spawnIdx%len(spawnPoints)]
	w.spawnIdx++
	p.X, p.Y, p.Z = pt[0], pt[1], pt[2]
	if p.InstanceID != "" {
		if inst := w.dungeonOf(p.ID); inst != nil {
			p.X, p.Y, p.Z = inst.CheckpointX, 0, inst.CheckpointZ
		}
	}
	p.VX, p.VY, p.VZ = 0, 0, 0
	p.HP = p.MaxHP
	p.Energy = p.MaxEnergy
	p.Stamina = p.MaxStamina
	p.CombatState = "IDLE"
	p.State = "IDLE"
	p.Grounded = true
	combatLog("%s respawned", p.Name)
	return [][]byte{marshal(TypePlayerRespawn, RespawnOut{PlayerID: p.ID, X: p.X, Y: p.Y, Z: p.Z, HP: p.HP})}
}

func (w *WorldState) tickCombat(dt float64) [][]byte {
	now := time.Now()
	var events [][]byte
	for _, p := range w.players {
		if !p.Connected {
			continue
		}
		if (p.CombatState == "DEAD" || p.CombatState == "RESPAWNING") && p.InstanceID == "" && !now.Before(p.RespawnAt) {
			events = append(events, w.playerRespawn(p, true)...)
		}
		if !p.alive() {
			continue
		}
		if now.After(p.DodgeUntil) && p.CombatState == "DODGING" {
			p.CombatState = "IDLE"
			p.VX *= 0.35
			p.VZ *= 0.35
		}
		if now.After(p.HitUntil) && (p.CombatState == "HIT" || p.CombatState == "STUNNED") {
			p.CombatState = "IDLE"
		}
		if now.After(p.InCombatUntil) && p.CombatState != "DODGING" && p.CombatState != "ATTACKING" && p.CombatState != "COMBO" && p.CombatState != "CASTING" && p.CombatState != "ENERGY_ATTACK" {
			p.hpRegenAcc += float64(p.MaxHP) * HpRegenPctPerSec * dt
			if p.hpRegenAcc >= 1 {
				add := int(p.hpRegenAcc)
				p.hpRegenAcc -= float64(add)
				p.HP = min(p.MaxHP, p.HP+add)
			}
			if p.Energy < p.MaxEnergy {
				regen := EnergyRegenPerSec + p.energyRegenBonus()
				if p.hasSkill("energy_control") {
					regen *= 1.05
				}
				p.energyAcc += regen * dt
				if p.energyAcc >= 1 {
					add := int(p.energyAcc)
					p.energyAcc -= float64(add)
					p.Energy = min(p.MaxEnergy, p.Energy+add)
				}
			}
		}
		events = append(events, w.tickTransform(p, dt, now)...)
		w.tickPhase25(p, dt, now)
		if now.After(p.DodgeUntil) && p.Stamina < p.MaxStamina {
			p.Stamina = math.Min(p.MaxStamina, p.Stamina+StaminaRegenPerSec*dt)
		}
		if p.CombatState == "ATTACKING" || p.CombatState == "COMBO" || p.CombatState == "CASTING" || p.CombatState == "ENERGY_ATTACK" {
			if now.After(p.AttackCDUntil) && now.After(p.ComboUntil) {
				p.CombatState = "IDLE"
			}
		}
	}
	events = append(events, w.tickEnemies(dt, now)...)
	return events
}

func normalizeAttack(t string) string {
	switch t {
	case "punch", "basic", "ATTACK", "PUNCH", "Power Punch", "light", "LIGHT":
		return "light"
	case "kick", "KICK", "Heavy Kick", "heavy", "HEAVY":
		return "heavy"
	default:
		return ""
	}
}

func comboRecipe() string {
	var b strings.Builder
	for _, step := range comboCfg.Finisher {
		if step == "heavy" {
			b.WriteByte('H')
		} else {
			b.WriteByte('L')
		}
	}
	if b.Len() == 0 {
		return "LLLH"
	}
	return b.String()
}

func comboPrefix(next, recipe string) bool {
	return strings.HasPrefix(recipe, next)
}

func comboNumbers(kind string, hits int, finisher bool) (base int, rng, kb float64, cd time.Duration) {
	rng, kb, cd = PunchRange, PunchKnockback, AttackCooldown
	dmg := comboCfg.LightDamage
	if kind == "heavy" {
		rng, kb, cd = KickRange, KickKnockback, KickCooldown
		dmg = comboCfg.HeavyDamage
	}
	if len(dmg) == 0 {
		if kind == "heavy" {
			base = 15
		} else {
			base = 10
		}
	} else {
		idx := hits - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(dmg) {
			idx = len(dmg) - 1
		}
		base = dmg[idx]
	}
	if finisher {
		if comboCfg.FinisherDamage > 0 {
			base = comboCfg.FinisherDamage
		}
		if comboCfg.FinisherKnockback > 0 {
			kb = comboCfg.FinisherKnockback
		}
	}
	return
}

func agilityCooldown(p *Player, cd time.Duration) time.Duration {
	scale := 1 - float64(p.Agility)*progressionCfg.AttackSpeedPerAgility
	if scale < 0.65 {
		scale = 0.65
	}
	if scale > 1 {
		scale = 1
	}
	return time.Duration(float64(cd) * scale)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
