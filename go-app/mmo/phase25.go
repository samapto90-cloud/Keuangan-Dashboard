package mmo

import (
	"strings"
	"time"
)

// Phase 25 overlay: advanced 3D combat + LIGHT ASCENSION HUD + combo streak +
// energy charge + original skills. Reuses CombatService, CharacterService,
// EnemyService, BossService, TransformationService, SkillService,
// AnimationService, VFXService, TargetService, EnergyService, EquipmentService,
// PartyService, QuestService, AchievementService, CameraService.
// Do not duplicate those services. Logical tables live on Player / PlayerLog:
// combat_profiles, combat_stats, skills, player_skills, skill_cooldowns,
// energy_states, combo_states, transformation_states, transformation_unlocks,
// combat_achievements, boss_combat_states.
// Indexes: playerId, skillId, transformationId, bossId.

const (
	CombatStreakMax     = 10
	AirComboMaxHits     = 6
	ChargeMoveFactor    = 0.28
	ChargeRegenBonus    = 18.0
	CounterBonusPower   = 1.18
	Phase25TransformRPS = 4
)

var phase25SkillAlias = map[string]string{
	"light_strike": "power_strike",
	"final_light":  "celestial_impact",
	"dawn_beam":    "focused_beam",
}

func init() {
	registerPhase25()
}

func registerPhase25() {
	registerPhase25Skills()
	registerPhase25Quests()
	registerPhase25Dialogue()
	registerPhase25Achievements()
	registerPhase25Cinematics()
	overlayQuestionPrompt("q1", "2 + 3 = ?")
	overlayNPC("mbah_karya", func(n *NPCDef) {
		n.QuestIDs = appendQuestIDs(n.QuestIDs, "nq-combat-1", "nq-combo-1", "nq-dodge-1", "nq-ascend-1", "nq-ascend-2", "nq-edu-combat")
	})
}

func registerPhase25Skills() {
	ensurePhase25Starters()
	if s, ok := skillCatalog["flash_step"]; ok && s.RequiredLevel > 1 {
		s.RequiredLevel = 1
		skillCatalog["flash_step"] = s
	}
	registerSkillDef(SkillDef{
		ID: "dawn_wave", Name: "Earth Break", Description: "Gelombang cahaya area original.",
		Kind: "ACTIVE", Type: "ENERGY", Target: "AREA", Damage: 20, EnergyCost: 28, Cooldown: 8,
		Range: 1.4, Radius: 2.8, RequiredLevel: 1, Animation: "KICK", VFX: "wave", Element: "LIGHT",
	})
	registerSkillDef(SkillDef{
		ID: "light_burst", Name: "Energy Burst", Description: "LIGHT BURST. Projectile energy original.",
		Kind: "ACTIVE", Type: "ENERGY", Target: "TARGET", Damage: 26, EnergyCost: 18, Cooldown: 5,
		Range: 16, RequiredLevel: 1, Animation: "ENERGY_ATTACK", VFX: "burst", Element: "LIGHT",
	})
	registerSkillDef(SkillDef{
		ID: "light_orb", Name: "Energy Burst", Description: "LIGHT ORB. Orb energy original.",
		Kind: "ACTIVE", Type: "ENERGY", Target: "TARGET", Damage: 22, EnergyCost: 16, Cooldown: 6,
		Range: 12, Radius: 1.6, RequiredLevel: 1, Animation: "ENERGY_ATTACK", VFX: "orb", Element: "LIGHT",
	})
}

func registerSkillDef(s SkillDef) {
	if _, ok := skillCatalog[s.ID]; ok {
		return
	}
	skillCatalog[s.ID] = s
}

func ensurePhase25Starters() {
	appendStarterSkill("flash_step")
	appendStarterSkill("dawn_wave")
	appendStarterSkill("light_burst")
}

func appendStarterSkill(id string) {
	for _, s := range progressionCfg.StarterSkills {
		if s == id {
			return
		}
	}
	progressionCfg.StarterSkills = append(progressionCfg.StarterSkills, id)
}

func appendQuestIDs(cur []string, ids ...string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(cur)+len(ids))
	for _, id := range cur {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func overlayQuestionPrompt(id, prompt string) {
	q, ok := questionByID[id]
	if !ok || q.ID == "" {
		return
	}
	q.Prompt = prompt
	questionByID[id] = q
	for i := range questionCatalog {
		if questionCatalog[i].ID == id {
			questionCatalog[i].Prompt = prompt
			return
		}
	}
}

func registerPhase25Quests() {
	registerQuest(QuestDef{
		ID: "nq-combat-1", Title: "Latihan Dhasar.", Kind: "combat", NPC: "mbah_karya",
		Location: "Dawn Village", Description: "10 serangan, 3 dodge, 1 block, 1 skill.",
		Objectives: []ObjectiveDef{
			{Type: "ATTACK", Target: "any", Count: 10, Text: "10 attacks"},
			{Type: "DODGE", Target: "any", Count: 3, Text: "3 dodges"},
			{Type: "BLOCK", Target: "any", Count: 1, Text: "1 block"},
			{Type: "SKILL", Target: "any", Count: 1, Text: "1 skill"},
		},
		Rewards:      RewardDef{Exp: 24, Coin: 6, Knowledge: 1},
		FlagsOnClaim: []string{"combat_tutorial_done"}, ClaimAt: "mbah_karya",
	})
	registerQuest(QuestDef{
		ID: "nq-combo-1", Title: "Serangan Berantai.", Kind: "combat", NPC: "mbah_karya",
		Location: "Dawn Village", Description: "Capai combo x5.",
		Objectives:   []ObjectiveDef{{Type: "COMBO", Target: "x5", Count: 1, Text: "combo x5"}},
		Rewards:      RewardDef{Exp: 22, Coin: 5, Knowledge: 1},
		FlagsOnClaim: []string{"combo_x5_done"}, ClaimAt: "mbah_karya",
	})
	registerQuest(QuestDef{
		ID: "nq-dodge-1", Title: "Langkah Cepet.", Kind: "combat", NPC: "mbah_karya",
		Location: "Dawn Village", Description: "3 perfect dodge.",
		Objectives:   []ObjectiveDef{{Type: "PERFECT_DODGE", Target: "any", Count: 3, Text: "3 perfect dodge"}},
		Rewards:      RewardDef{Exp: 22, Coin: 5, Knowledge: 1},
		FlagsOnClaim: []string{"perfect_dodge_done"}, ClaimAt: "mbah_karya",
	})
	registerQuest(QuestDef{
		ID: "nq-ascend-1", Title: "Ngudi Kekuwatan.", Kind: "combat", NPC: "mbah_karya",
		Location: "Guild Training Room", Description: "Latih awakmu. LIGHT ASCENSION original.",
		Objectives:   []ObjectiveDef{{Type: "KILL", Target: "training_dummy", Count: 1, Text: "Kalahake training dummy"}},
		Rewards:      RewardDef{Exp: 28, Coin: 8, Knowledge: 1},
		FlagsOnClaim: []string{"ngudi_kekuwatan"}, ClaimAt: "mbah_karya",
	})
	registerQuest(QuestDef{
		ID: "nq-ascend-2", Title: "Awakmu Wis Siap?", Kind: "combat", NPC: "mbah_karya",
		Location: "Guild Training Room", Description: "Rampungake latihan transformasi.",
		Prereq:       []string{"nq-ascend-1"},
		Objectives:   []ObjectiveDef{{Type: "TALK", Target: "mbah_karya", Count: 1, Text: "Rampungake training"}},
		Rewards:      RewardDef{Exp: 30, Coin: 8, Knowledge: 1},
		FlagsOnClaim: []string{"awakmu_wis_siap", "training_complete"}, ClaimAt: "mbah_karya",
	})
	registerQuest(QuestDef{
		ID: "nq-edu-combat", Title: "Pitakonan Latihan.", Kind: "education", NPC: "mbah_karya",
		Location: "Dawn Village", Description: "2 + 3 = ?",
		Objectives:   []ObjectiveDef{{Type: "ANSWER", Target: "q1", Count: 1, Text: "Jawab 2 + 3"}},
		Rewards:      RewardDef{Exp: 12, EduToken: 1, Knowledge: 1},
		FlagsOnClaim: []string{"combat_edu_done"}, ClaimAt: "mbah_karya",
	})
}

func registerPhase25Dialogue() {
	overlayLoc("p25.karya.punch", "Saiki coba mukul kayu iku.", "Sekarang coba pukul kayu itu.")
	overlayLoc("p25.karya.ngudi", "Kowe kudu nglatih awakmu luwih sregep, Le.", "Kamu harus melatih dirimu lebih giat, Nak.")
	overlayLoc("p25.karya.siap", "Awakmu wis siap, Le?", "Kamu sudah siap, Nak?")
}

func registerPhase25Achievements() {
	registerAchievementDef(AchievementDef{ID: "first-strike", Name: "First Strike", Title: "first-strike", Category: "combat"})
	registerAchievementDef(AchievementDef{ID: "combo-beginner", Name: "Combo Beginner", Title: "combo-beginner", Category: "combat"})
	registerAchievementDef(AchievementDef{ID: "combo-master", Name: "Combo Master", Title: "combo-master", Category: "combat"})
	registerAchievementDef(AchievementDef{ID: "perfect-dodge", Name: "Perfect Dodge", Title: "perfect-dodge", Category: "combat"})
	registerAchievementDef(AchievementDef{ID: "energy-master", Name: "Energy Master", Title: "energy-master", Category: "combat"})
	registerAchievementDef(AchievementDef{ID: "first-ascension", Name: "First Ascension", Title: "first-ascension", Category: "combat"})
	registerAchievementDef(AchievementDef{ID: "ascension-iv", Name: "Ascension IV", Title: "ascension-iv", Category: "combat"})
	registerAchievementDef(AchievementDef{ID: "boss-breaker", Name: "Boss Breaker", Title: "boss-breaker", Category: "combat"})
	registerTitleDef(TitleDef{ID: "first-strike", Name: "First Strike", Source: "combat", Rarity: "COMMON"})
	registerTitleDef(TitleDef{ID: "combo-beginner", Name: "Combo Beginner", Source: "combat", Rarity: "COMMON"})
	registerTitleDef(TitleDef{ID: "combo-master", Name: "Combo Master", Source: "combat", Rarity: "RARE"})
	registerTitleDef(TitleDef{ID: "perfect-dodge", Name: "Perfect Dodge", Source: "combat", Rarity: "RARE"})
	registerTitleDef(TitleDef{ID: "energy-master", Name: "Energy Master", Source: "combat", Rarity: "RARE"})
	registerTitleDef(TitleDef{ID: "first-ascension", Name: "First Ascension", Source: "combat", Rarity: "RARE"})
	registerTitleDef(TitleDef{ID: "ascension-iv", Name: "Ascension IV", Source: "combat", Rarity: "EPIC"})
	registerTitleDef(TitleDef{ID: "boss-breaker", Name: "Boss Breaker", Source: "combat", Rarity: "EPIC"})
	registerTitleDef(TitleDef{ID: "rank-ss", Name: "Combat SS", Source: "training", Rarity: "EPIC"})
	registerTitleDef(TitleDef{ID: "rank-s", Name: "Combat S", Source: "training", Rarity: "RARE"})
	registerCosmeticDef(CosmeticDef{ID: "aura-dawn-spark", Name: "Dawn Spark Aura", Kind: "aura"})
	registerCosmeticDef(CosmeticDef{ID: "cloak-combo", Name: "Combo Cloak", Kind: "cloak"})
}

func registerPhase25Cinematics() {
	registerCinematic(CinematicDef{
		ID: "cin-light-ascension", Title: "LIGHT ASCENSION", DurationSec: 8, Skippable: true,
		Camera: "zoom", Music: "combat-dawn", VFX: "aura",
		Lines: []string{"Cahaya munggah.", "Aura original.", "Ascension."},
	})
	registerCinematic(CinematicDef{
		ID: "cin-boss-break", Title: "BOSS BREAK", DurationSec: 6, Skippable: true,
		Camera: "focus", Music: "combat-boss",
		Lines: []string{"Boss vulnerable.", "Break window."},
	})
}

func resolvePhase25Skill(id string) string {
	if alias, ok := phase25SkillAlias[id]; ok {
		return alias
	}
	return id
}

func (w *WorldState) applyPhase25Combat(p *Player, env Envelope) [][]byte {
	switch env.Type {
	case TypePlayerCharge:
		var in BlockIn
		_ = unmarshal(env.Data, &in)
		return w.playerCharge(p, in.On)
	case TypeSetEnergy, TypeSetCooldown, TypeUnlockTransform, TypeSetDamage:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return nil
	}
}

func (w *WorldState) playerCharge(p *Player, on bool) [][]byte {
	now := time.Now()
	if !p.alive() {
		return rejectFor(p.ID, TypePlayerCharge, "dead")
	}
	if now.Before(p.StunUntil) || p.CombatState == "DOWNED" || p.CombatState == "DEAD" {
		return rejectFor(p.ID, TypePlayerCharge, "busy")
	}
	p.Charging = on
	if on {
		p.CombatState = "CHARGE"
		p.ChargeUntil = now.Add(12 * time.Second)
		p.Sprint = false
	} else if p.CombatState == "CHARGE" {
		p.CombatState = "IDLE"
	}
	return [][]byte{marshal(TypeAttackResult, AttackResult{AttackerID: p.ID, AttackType: "charge", Timestamp: now.UnixMilli()})}
}

func (p *Player) stopCharge() {
	if !p.Charging && p.CombatState != "CHARGE" {
		return
	}
	p.Charging = false
	if p.CombatState == "CHARGE" {
		p.CombatState = "IDLE"
	}
}

func (w *WorldState) phase25SafeReconnect(p *Player) {
	if p == nil {
		return
	}
	switch p.CombatState {
	case "ATTACKING", "COMBO", "CASTING", "DODGING", "CHARGE", "COUNTER", "ENERGY_ATTACK", "ULTIMATE", "GUARD":
		p.CombatState = "IDLE"
	}
	p.ComboChain = ""
	p.CombatStreak = 0
	p.Charging = false
	p.Blocking = false
	p.AttackCDUntil = time.Time{}
}

func (w *WorldState) phase25OnAttack(p *Player) {
	p.stopCharge()
	p.credit("ATTACK", "any", 1)
	log := p.ensureLog()
	log.CombatHits++
	w.phase25Grant(p, "first-strike", "first-strike")
}

func (w *WorldState) phase25OnDodge(p *Player) {
	p.stopCharge()
	p.credit("DODGE", "any", 1)
}

func (w *WorldState) phase25OnBlock(p *Player) {
	p.credit("BLOCK", "any", 1)
}

func (w *WorldState) phase25OnSkill(p *Player, skillID string) {
	p.stopCharge()
	p.credit("SKILL", "any", 1)
	if skillID == "celestial_impact" || skillID == "final_light" {
		w.phase25Grant(p, "energy-master", "energy-master")
	}
}

func (w *WorldState) phase25OnHit(p *Player, e *Enemy, attackType string) {
	now := time.Now()
	if now.After(p.ComboUntil) {
		p.CombatStreak = 0
	}
	p.CombatStreak++
	if p.CombatStreak > CombatStreakMax {
		p.CombatStreak = CombatStreakMax
	}
	if p.CombatStreak >= 3 {
		w.phase25Grant(p, "combo-beginner", "combo-beginner")
	}
	if p.CombatStreak >= 8 {
		w.phase25Grant(p, "combo-master", "combo-master")
	}
	if p.CombatStreak >= 5 {
		p.credit("COMBO", "x5", 1)
	}
	if e != nil && (e.IsBoss || strings.Contains(strings.ToLower(e.Def.Rank), "boss")) && e.Stagger >= e.StaggerMax && e.StaggerMax > 0 {
		w.phase25Grant(p, "boss-breaker", "boss-breaker")
	}
	if attackType == "heavy" || attackType == "kick" {
		if e != nil && e.Y < 1.4 {
			e.Y = 1.2
		}
	}
	w.phase25TrainingRank(p)
}

func (w *WorldState) phase25OnPerfectDodge(p *Player) {
	p.credit("PERFECT_DODGE", "any", 1)
	p.ensureLog().PerfectDodges++
	if p.ensureLog().PerfectDodges >= 3 {
		w.phase25Grant(p, "perfect-dodge", "perfect-dodge")
	}
}

func (w *WorldState) phase25OnTransform(p *Player, formID string) {
	if formID == "aura-1" {
		w.phase25Grant(p, "first-ascension", "first-ascension")
	}
	if formID == "celestial-4" {
		w.phase25Grant(p, "ascension-iv", "ascension-iv")
	}
}

func (w *WorldState) phase25Grant(p *Player, id, title string) {
	if p == nil {
		return
	}
	log := p.ensureLog()
	for _, a := range log.Achievements {
		if a == id {
			return
		}
	}
	log.Achievements = append(log.Achievements, id)
	p.grantTitle(title)
	if id == "first-ascension" || id == "combo-master" {
		p.grantCosmetic("aura-dawn-spark")
	}
	if id == "combo-beginner" {
		p.grantCosmetic("cloak-combo")
	}
	w.audit("achievementUnlocked", p.ID, id)
}

func (w *WorldState) phase25TrainingRank(p *Player) {
	tr := p.ensureLog().Training
	if tr.Hits < 1 {
		return
	}
	score := tr.Damage + tr.Hits*4 + tr.Combo*12 + p.ensureLog().PerfectDodges*20
	rank := "C"
	switch {
	case score >= 420:
		rank = "SS"
		p.grantTitle("rank-ss")
	case score >= 280:
		rank = "S"
		p.grantTitle("rank-s")
	case score >= 180:
		rank = "A"
	case score >= 90:
		rank = "B"
	}
	tr.Score = score
	tr.Rank = rank
	p.ensureLog().Training = tr
}

func (w *WorldState) tickPhase25(p *Player, dt float64, now time.Time) {
	if now.After(p.ComboUntil) {
		p.CombatStreak = 0
	}
	if p.Charging {
		if now.After(p.ChargeUntil) || !p.alive() {
			p.stopCharge()
			return
		}
		if p.Energy < p.MaxEnergy {
			p.energyAcc += (EnergyRegenPerSec + ChargeRegenBonus + p.energyRegenBonus()) * dt
			if p.energyAcc >= 1 {
				add := int(p.energyAcc)
				p.energyAcc -= float64(add)
				p.Energy = min(p.MaxEnergy, p.Energy+add)
			}
		}
	}
}

func phase25KaryaOptions(p *Player) []DialogOption {
	opts := []DialogOption{
		{ID: "choice-train", Label: "Latihan"},
		{ID: "quiz-combat", Label: "2 + 3 = ?"},
	}
	opts = append(opts, questOption(p, "nq-combat-1", "Latihan Dhasar.")...)
	opts = append(opts, questOption(p, "nq-combo-1", "Serangan Berantai.")...)
	opts = append(opts, questOption(p, "nq-dodge-1", "Langkah Cepet.")...)
	opts = append(opts, questOption(p, "nq-ascend-1", "Ngudi Kekuwatan.")...)
	opts = append(opts, questOption(p, "nq-ascend-2", "Awakmu Wis Siap?")...)
	opts = append(opts, questOption(p, "nq-edu-combat", "Pitakonan Latihan.")...)
	return opts
}

func (w *WorldState) startPhase25Quiz(p *Player, npc NPCDef) [][]byte {
	idx := -1
	for i, q := range questionCatalog {
		if q.ID == "q1" {
			idx = i
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeInteract, "question")
	}
	qlog := p.quest("nq-edu-combat")
	if qlog != nil && qlog.State == QuestAvailable {
		qlog.State = QuestActive
	}
	p.ensureLog().Quiz = QuizSession{QuestID: "nq-edu-combat", Index: idx, Active: true}
	p.markDirty()
	w.persist(p)
	q := questionCatalog[idx]
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "edu-story", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
		Text: q.Prompt, VoiceID: npc.VoiceLineID,
		Question: &QuestionOut{ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices, ToID: p.ID},
	})}
}

func (w *WorldState) answerPhase25Edu(p *Player, in EducationAnswerIn) [][]byte {
	log := p.ensureLog()
	if !log.Quiz.Active {
		return rejectFor(p.ID, TypeEducationAnswer, "no_session")
	}
	idx := log.Quiz.Index
	if idx < 0 || idx >= len(questionCatalog) {
		return rejectFor(p.ID, TypeEducationAnswer, "question")
	}
	def := questionCatalog[idx]
	if in.QuestionID != "" && in.QuestionID != def.ID {
		return rejectFor(p.ID, TypeEducationAnswer, "order")
	}
	if in.Choice != def.Correct {
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: false, Explain: def.Explain, Retry: true, Question: questionOut(idx),
		})}
	}
	p.credit("ANSWER", def.ID, 1)
	p.grantKnowledge(1)
	w.grantEducationBonus(p)
	p.bumpNpcRel("mbah_karya", 6)
	log.Flags["combat_edu_done"] = true
	log.Quiz.Active = false
	if q := p.quest("nq-edu-combat"); q != nil && q.State == QuestActive {
		if d, ok := questByID[q.ID]; ok && objectivesDone(q, d) {
			q.State = QuestCompleted
		}
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
		Correct: true, Explain: def.Explain, Toast: "Knowledge Token", Question: questionOut(idx),
	})}
}
