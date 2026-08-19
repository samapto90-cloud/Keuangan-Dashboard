package mmo

import (
	"strings"
	"time"
)

// Phase 20 overlay: CharacterService + CombatService + SkillService +
// TransformationService + InventoryService + EquipmentService + QuestService +
// StoryService + AchievementService + PartyService + GuildService.
// Logical tables persist on PlayerLog / GearSave (not a second store):
// character_attributes, character_styles, style_mastery, skill_nodes,
// player_skills, player_builds, build_skills, combat_loadouts,
// transformation_progress, transformation_mastery, combat_stats, training_records.
// Indexes: playerId, skillId, styleId, transformationId, buildId.

const (
	BuildSlots          = 3
	AttrResetFree       = 3
	AttrResetCooldown   = 30 * time.Minute
	SkillResetCooldown  = 10 * time.Minute
	MaxMoveBonus        = 0.18
	PerfectDodgeWindow  = 120 * time.Millisecond
	BlockDamageFactor   = 0.55
	GuardBreakStagger   = 40
	MaxUltCharge        = 100
	CombatRatePerSec    = 20
)

const (
	StyleDawnFist       = "DAWN_FIST"
	StyleWindStep       = "WIND_STEP"
	StyleIronGuard      = "IRON_GUARD"
	StyleCelestialFlow  = "CELESTIAL_FLOW"
)

var attrAlias = map[string]string{
	"STR": "STR", "POWER": "STR",
	"DEF": "DEF", "RESOLVE": "DEF",
	"AGI": "AGI", "AGILITY": "AGI",
	"ENG": "ENG", "FOCUS": "ENG",
	"VIT": "VIT", "VITALITY": "VIT",
}

var attrDisplay = map[string]string{
	"STR": "POWER", "DEF": "RESOLVE", "AGI": "AGILITY", "ENG": "FOCUS", "VIT": "VITALITY",
}

var styleCatalog = []CombatStyleDef{
	{ID: StyleDawnFist, Name: "DAWN FIST", Focus: "close-range attack", Ultimate: "celestial_impact"},
	{ID: StyleWindStep, Name: "WIND STEP", Focus: "mobility dodge rapid attack", Ultimate: "celestial_impact"},
	{ID: StyleIronGuard, Name: "IRON GUARD", Focus: "defense counter", Ultimate: "celestial_impact"},
	{ID: StyleCelestialFlow, Name: "CELESTIAL FLOW", Focus: "energy technique ranged attack", Ultimate: "celestial_impact"},
}

var formStoryName = map[string]string{
	"aura-1":      "AWAKENED FORM",
	"aura-2":      "ASCENDED FORM",
	"aura-3":      "RADIANT FORM",
	"celestial-4": "CELESTIAL FORM",
}

type CombatStyleDef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Focus    string `json:"focus"`
	Ultimate string `json:"ultimate"`
}

type CombatBuild struct {
	ID         string            `json:"id"`
	Slot       int               `json:"slot"`
	Name       string            `json:"name"`
	Style      string            `json:"style"`
	FormID     string            `json:"formId"`
	SpentSTR   int               `json:"spentStr"`
	SpentDEF   int               `json:"spentDef"`
	SpentAGI   int               `json:"spentAgi"`
	SpentENG   int               `json:"spentEng"`
	SpentVIT   int               `json:"spentVit"`
	Skills     []string          `json:"skills"`
	Loadout    []string          `json:"loadout"`
	Ultimate   string            `json:"ultimate"`
	Equipment  map[string]string `json:"equipment"`
	Score      int               `json:"score"`
	Active     bool              `json:"active"`
}

type TrainingRecord struct {
	DummyID string `json:"dummyId"`
	Damage  int    `json:"damage"`
	Hits    int    `json:"hits"`
	DPS     int    `json:"dps"`
	Combo   int    `json:"combo"`
	Score   int    `json:"score,omitempty"`
	Rank    string `json:"rank,omitempty"`
}

func (p *Player) inCombatNow() bool {
	return time.Now().Before(p.InCombatUntil)
}

func (p *Player) ensureStyle() string {
	if p.CombatStyle == "" {
		p.CombatStyle = StyleDawnFist
	}
	return p.CombatStyle
}

func (p *Player) ensurePhase20() {
	log := p.ensureLog()
	if log.StyleMastery == nil {
		log.StyleMastery = map[string]int{}
	}
	if log.FormMastery == nil {
		log.FormMastery = map[string]int{}
	}
	if log.StatusResist == nil {
		log.StatusResist = map[string]float64{}
	}
	if log.Builds == nil {
		log.Builds = make([]CombatBuild, BuildSlots)
		for i := range log.Builds {
			log.Builds[i].Slot = i
			log.Builds[i].ID = "build-" + itoa(i)
		}
	}
	if len(log.Builds) < BuildSlots {
		next := make([]CombatBuild, BuildSlots)
		copy(next, log.Builds)
		for i := range next {
			if next[i].ID == "" {
				next[i].Slot = i
				next[i].ID = "build-" + itoa(i)
			}
		}
		log.Builds = next
	}
	if p.CombatStyle == "" {
		if log.CombatStyle != "" {
			p.CombatStyle = log.CombatStyle
		} else {
			p.CombatStyle = StyleDawnFist
		}
	}
	if p.StaggerMax < 1 {
		p.StaggerMax = 80 + p.SpentDEF*4
	}
	if p.StatusUntil == nil {
		p.StatusUntil = map[string]time.Time{}
	}
}

func normalizeAttr(stat string) string {
	return attrAlias[strings.ToUpper(strings.TrimSpace(stat))]
}

func validStyle(id string) bool {
	switch strings.ToUpper(id) {
	case StyleDawnFist, StyleWindStep, StyleIronGuard, StyleCelestialFlow:
		return true
	}
	return false
}

func (w *WorldState) ApplyBuild(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	p.ensurePhase20()
	switch env.Type {
	case TypeGetProgression:
		return [][]byte{marshal(TypeProgressionState, p.progressionView())}
	case TypeSetCombatStyle:
		var in StyleIn
		_ = unmarshal(env.Data, &in)
		return w.setCombatStyle(p, in.StyleID)
	case TypeSaveBuild:
		var in BuildIn
		_ = unmarshal(env.Data, &in)
		return w.saveBuild(p, in)
	case TypeLoadBuild, TypeSwitchBuild:
		var in BuildIn
		_ = unmarshal(env.Data, &in)
		return w.loadBuild(p, in.Slot)
	case TypeSetLoadout:
		var in LoadoutIn
		_ = unmarshal(env.Data, &in)
		return w.setLoadout(p, in)
	case TypeResetSkills:
		return w.resetSkills(p)
	case TypeGetBuilds:
		return [][]byte{marshal(TypeBuildList, p.buildListView())}
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) setCombatStyle(p *Player, style string) [][]byte {
	style = strings.ToUpper(strings.TrimSpace(style))
	if !validStyle(style) {
		return rejectFor(p.ID, TypeSetCombatStyle, "style")
	}
	if p.inCombatNow() {
		return rejectFor(p.ID, TypeSetCombatStyle, "combat")
	}
	p.CombatStyle = style
	p.ensureLog().CombatStyle = style
	w.persist(p)
	w.persistGear(p)
	return [][]byte{marshal(TypeProgressionState, p.progressionView())}
}

func (w *WorldState) saveBuild(p *Player, in BuildIn) [][]byte {
	p.ensurePhase20()
	slot := in.Slot
	if slot < 0 || slot >= BuildSlots {
		return rejectFor(p.ID, TypeSaveBuild, "slot")
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = "Build " + itoa(slot+1)
	}
	if len(name) > 24 {
		name = name[:24]
	}
	style := p.ensureStyle()
	if in.Style != "" && validStyle(strings.ToUpper(in.Style)) {
		style = strings.ToUpper(in.Style)
	}
	skills := append([]string{}, p.UnlockedSkills...)
	loadout := append([]string{}, p.LoadoutSkills...)
	ult := p.LoadoutUlt
	if ult == "" {
		ult = "celestial_impact"
	}
	eq := map[string]string{
		"HEAD": p.Gear.HEAD, "BODY": p.Gear.BODY, "LEGS": p.Gear.LEGS,
		"WEAPON": p.Gear.WEAPON, "ACCESSORY_1": p.Gear.ACCESSORY_1, "ACCESSORY_2": p.Gear.ACCESSORY_2,
	}
	b := CombatBuild{
		ID: "build-" + itoa(slot), Slot: slot, Name: name, Style: style, FormID: p.FormID,
		SpentSTR: p.SpentSTR, SpentDEF: p.SpentDEF, SpentAGI: p.SpentAGI, SpentENG: p.SpentENG, SpentVIT: p.SpentVIT,
		Skills: skills, Loadout: loadout, Ultimate: ult, Equipment: eq, Score: p.buildScore(),
	}
	log := p.ensureLog()
	for i := range log.Builds {
		log.Builds[i].Active = i == slot
	}
	b.Active = true
	log.Builds[slot] = b
	log.ActiveBuild = slot
	w.persist(p)
	w.persistGear(p)
	return [][]byte{marshal(TypeBuildSaved, p.buildListView()), marshal(TypeProgressionState, p.progressionView())}
}

func (w *WorldState) loadBuild(p *Player, slot int) [][]byte {
	p.ensurePhase20()
	if p.inCombatNow() {
		return rejectFor(p.ID, TypeLoadBuild, "combat")
	}
	if slot < 0 || slot >= BuildSlots {
		return rejectFor(p.ID, TypeLoadBuild, "slot")
	}
	log := p.ensureLog()
	b := log.Builds[slot]
	if b.Name == "" && b.Style == "" && b.SpentSTR+b.SpentDEF+b.SpentAGI+b.SpentENG+b.SpentVIT == 0 {
		return rejectFor(p.ID, TypeLoadBuild, "empty")
	}
	if reason := w.validateBuild(p, b); reason != "" {
		return rejectFor(p.ID, TypeLoadBuild, reason)
	}
	spent := p.SpentSTR + p.SpentDEF + p.SpentAGI + p.SpentENG + p.SpentVIT
	p.AttributePoints += spent
	need := b.SpentSTR + b.SpentDEF + b.SpentAGI + b.SpentENG + b.SpentVIT
	if p.AttributePoints < need {
		return rejectFor(p.ID, TypeLoadBuild, "points")
	}
	p.SpentSTR, p.SpentDEF, p.SpentAGI, p.SpentENG, p.SpentVIT = b.SpentSTR, b.SpentDEF, b.SpentAGI, b.SpentENG, b.SpentVIT
	p.AttributePoints -= need
	if validStyle(b.Style) {
		p.CombatStyle = b.Style
		log.CombatStyle = b.Style
	}
	if len(b.Loadout) > 0 {
		p.LoadoutSkills = append([]string{}, b.Loadout...)
	}
	if b.Ultimate != "" {
		p.LoadoutUlt = b.Ultimate
	}
	pref := b.FormID
	if pref != "" && pref != "normal" && pref != p.FormID && p.hasForm(pref) && p.ensureLog().PendingCinematic == "" && p.alive() {
		_ = w.requestTransform(p, pref)
	}
	for i := range log.Builds {
		log.Builds[i].Active = i == slot
	}
	log.ActiveBuild = slot
	p.applyDerived()
	w.persist(p)
	w.persistGear(p)
	return [][]byte{
		marshal(TypeBuildLoaded, p.buildListView()),
		marshal(TypeProgressionState, p.progressionView()),
		marshal(TypePlayerStatsUpdated, p.statsView()),
	}
}

func (w *WorldState) validateBuild(p *Player, b CombatBuild) string {
	need := b.SpentSTR + b.SpentDEF + b.SpentAGI + b.SpentENG + b.SpentVIT
	pool := p.AttributePoints + p.SpentSTR + p.SpentDEF + p.SpentAGI + p.SpentENG + p.SpentVIT
	if need > pool {
		return "points"
	}
	for _, id := range b.Skills {
		if id == "" {
			continue
		}
		if !p.hasSkill(id) {
			return "skill"
		}
	}
	for _, id := range b.Loadout {
		if id == "" {
			continue
		}
		if !p.hasSkill(id) {
			return "skill"
		}
		if def, ok := skillCatalog[id]; ok && p.Level < def.RequiredLevel {
			return "level"
		}
	}
	if b.Ultimate != "" && !p.hasSkill(b.Ultimate) && b.Ultimate != "celestial_impact" {
		return "skill"
	}
	if b.FormID != "" && b.FormID != "normal" && !p.hasForm(b.FormID) {
		return "form"
	}
	for _, itemID := range b.Equipment {
		if itemID == "" {
			continue
		}
		owned := p.Gear.hasItem(itemID) || (p.Bag != nil && p.Bag.count(itemID) > 0)
		if !owned {
			return "equipment"
		}
	}
	if b.Style != "" && !validStyle(b.Style) {
		return "style"
	}
	return ""
}

func (w *WorldState) setLoadout(p *Player, in LoadoutIn) [][]byte {
	if p.inCombatNow() {
		return rejectFor(p.ID, TypeSetLoadout, "combat")
	}
	skills := in.Skills
	if len(skills) > 4 {
		skills = skills[:4]
	}
	for _, id := range skills {
		if id == "" {
			continue
		}
		if !p.hasSkill(id) {
			return rejectFor(p.ID, TypeSetLoadout, "skill")
		}
	}
	ult := in.Ultimate
	if ult == "" {
		ult = "celestial_impact"
	}
	if ult != "celestial_impact" && !p.hasSkill(ult) {
		return rejectFor(p.ID, TypeSetLoadout, "ultimate")
	}
	p.LoadoutSkills = append([]string{}, skills...)
	p.LoadoutUlt = ult
	w.persistGear(p)
	return [][]byte{marshal(TypeProgressionState, p.progressionView())}
}

func (w *WorldState) resetSkills(p *Player) [][]byte {
	now := time.Now()
	if now.Before(p.SkillResetUntil) {
		return rejectFor(p.ID, TypeResetSkills, "cooldown")
	}
	refund := 0
	keep := map[string]bool{}
	for _, id := range progressionCfg.StarterSkills {
		keep[id] = true
	}
	var next []string
	for _, id := range p.UnlockedSkills {
		if keep[id] {
			next = append(next, id)
			continue
		}
		cost := 1
		for _, n := range skillNodeByID {
			if n.SkillID == id && n.Cost > 0 {
				cost = n.Cost
				break
			}
		}
		refund += cost
	}
	p.UnlockedSkills = next
	p.SkillPoints += refund
	p.SkillResetUntil = now.Add(SkillResetCooldown)
	p.ensureLog().SkillResetAt = p.SkillResetUntil.UnixMilli()
	p.applyDerived()
	w.persist(p)
	w.persistGear(p)
	return [][]byte{marshal(TypeProgressionState, p.progressionView())}
}

func (p *Player) buildScore() int {
	n := p.powerRating()
	n += styleMasteryLevel(p.ensureLog().StyleMastery[p.ensureStyle()]) * 4
	return n
}

func (p *Player) combatRating() int {
	return p.powerRating()
}

func styleMasteryLevel(xp int) int {
	lv := 1 + xp/80
	if lv > 20 {
		lv = 20
	}
	return lv
}

func (p *Player) buildListView() BuildListView {
	p.ensurePhase20()
	log := p.ensureLog()
	out := make([]CombatBuild, len(log.Builds))
	copy(out, log.Builds)
	return BuildListView{PlayerID: p.ID, Slots: out, Active: log.ActiveBuild, ToID: p.ID}
}

func (w *WorldState) playerBlock(p *Player, on bool) [][]byte {
	now := time.Now()
	if !p.alive() {
		return rejectFor(p.ID, TypePlayerBlock, "dead")
	}
	if now.Before(p.StunUntil) {
		return rejectFor(p.ID, TypePlayerBlock, "stun")
	}
	p.Blocking = on
	if on {
		p.GuardUntil = now.Add(2 * time.Second)
		p.CombatState = "GUARD"
		w.phase25OnBlock(p)
	} else if p.CombatState == "GUARD" {
		p.CombatState = "IDLE"
	}
	return [][]byte{marshal(TypeAttackResult, AttackResult{AttackerID: p.ID, AttackType: "block", Timestamp: now.UnixMilli()})}
}

func (w *WorldState) playerCounter(p *Player) [][]byte {
	now := time.Now()
	window := p.Blocking || now.Before(p.GuardUntil) || now.Before(p.PerfectDodgeUntil)
	if p.ensureStyle() != StyleIronGuard && !now.Before(p.PerfectDodgeUntil) && !p.Blocking && now.After(p.GuardUntil) {
		return rejectFor(p.ID, TypePlayerCounter, "style")
	}
	if !p.alive() || now.Before(p.StunUntil) {
		return rejectFor(p.ID, TypePlayerCounter, "busy")
	}
	if !window {
		return rejectFor(p.ID, TypePlayerCounter, "window")
	}
	p.CombatState = "COUNTER"
	p.IFrameUntil = now.Add(90 * time.Millisecond)
	p.InCombatUntil = now.Add(CombatIdleAfter)
	e := w.nearestEnemyFor(p, KickRange+0.8)
	power := 1.05
	if now.Before(p.PerfectDodgeUntil) {
		power = CounterBonusPower
	}
	var events [][]byte
	if e != nil {
		events = append(events, w.hitEnemy(p, e, 12, power, 1.6, "PHYSICAL")...)
		e.Stagger += 18
	}
	events = append(events, marshal(TypeAttackResult, AttackResult{AttackerID: p.ID, AttackType: "counter", Timestamp: now.UnixMilli()}))
	return events
}

func (p *Player) styleBuff() StatBlock {
	s := StatBlock{}
	switch p.ensureStyle() {
	case StyleDawnFist:
		s.Attack += 2
	case StyleWindStep:
		s.Agility += 2
		s.MoveSpeed += 0.03
	case StyleIronGuard:
		s.Defense += 3
	case StyleCelestialFlow:
		s.EnergyPower += 2
		s.MaxEnergy += 6
	}
	return s
}

func (p *Player) formPassive() StatBlock {
	s := StatBlock{}
	if p.FormID == "" || p.FormID == "normal" {
		return s
	}
	if p.TransformState != "TRANSFORMING" && p.TransformState != "TRANSFORMED" {
		return s
	}
	switch p.FormID {
	case "aura-1":
		s.Attack += 1
	case "aura-2":
		s.Agility += 1
	case "aura-3":
		s.MaxEnergy += 4
	case "celestial-4":
		s.Attack += 2
		s.Defense += 1
	}
	return s
}

func capMoveBonus(v float64) float64 {
	if v > MaxMoveBonus {
		return MaxMoveBonus
	}
	if v < 0 {
		return 0
	}
	return v
}

func (p *Player) energyRegenBonus() float64 {
	bonus := float64(p.SpentDEF) * 0.12
	if p.FormID == "aura-3" && (p.TransformState == "TRANSFORMED" || p.TransformState == "TRANSFORMING") {
		bonus += 1.2
	}
	if now, ok := p.StatusUntil["FOCUS"]; ok && time.Now().Before(now) {
		bonus += 1.5
	}
	if p.setCount() >= 4 {
		bonus += 0.6
	}
	return bonus
}

func (w *WorldState) afterCombatHit(p *Player, e *Enemy, dealt int, crit bool, attackType string) {
	p.ensurePhase20()
	log := p.ensureLog()
	style := p.ensureStyle()
	log.StyleMastery[style] += 1
	if p.FormID != "" && p.FormID != "normal" {
		log.FormMastery[p.FormID] += 1
	}
	gain := 6
	if crit {
		gain += 4
	}
	if attackType == "ENERGY" || attackType == "energy" {
		gain += 2
	}
	p.UltCharge += gain
	if p.UltCharge > MaxUltCharge {
		p.UltCharge = MaxUltCharge
	}
	if e != nil && e.Def.ID == "training_dummy" {
		w.recordTrainingHit(p, e, dealt)
	}
	if e != nil {
		w.applyCombatStatus(p, e, attackType)
		w.applyStagger(p, e, dealt, attackType)
	}
	if p.ComboChain == "LLH" || strings.HasSuffix(p.ComboChain, "LLH") {
		p.ComboMasteryReady = true
		log.StyleMastery[style] += 4
	}
	w.grantMasteryRewards(p, style)
	w.phase25OnHit(p, e, attackType)
}

func (w *WorldState) recordTrainingHit(p *Player, e *Enemy, dealt int) {
	now := time.Now()
	if p.TrainingStart.IsZero() || now.Sub(p.TrainingStart) > 8*time.Second {
		p.TrainingStart = now
		p.TrainingDmg = 0
		p.TrainingHits = 0
	}
	p.TrainingDmg += dealt
	p.TrainingHits++
	sec := now.Sub(p.TrainingStart).Seconds()
	if sec < 0.25 {
		sec = 0.25
	}
	dps := int(float64(p.TrainingDmg)/sec + 0.5)
	p.ensureLog().Training = TrainingRecord{DummyID: e.ID, Damage: p.TrainingDmg, Hits: p.TrainingHits, DPS: dps, Combo: len(p.ComboChain)}
}

func (w *WorldState) applyStagger(p *Player, e *Enemy, dealt int, attackType string) {
	if e.StaggerMax < 1 {
		e.StaggerMax = 60 + e.Def.Level*4
	}
	add := dealt / 3
	if attackType == "heavy" || attackType == "kick" {
		add += GuardBreakStagger / 4
		if e.Guarding {
			e.Guarding = false
			add += GuardBreakStagger
		}
	}
	e.Stagger += add
	if e.Stagger >= e.StaggerMax {
		e.Stagger = 0
		e.State = "STUNNED"
		e.NextAttack = time.Now().Add(900 * time.Millisecond)
	}
}

func (w *WorldState) applyCombatStatus(p *Player, e *Enemy, attackType string) {
	if e.StatusUntil == nil {
		e.StatusUntil = map[string]time.Time{}
	}
	now := time.Now()
	el := skillElement(attackType, p)
	switch el {
	case "WIND":
		e.StatusUntil["SLOW"] = now.Add(1800 * time.Millisecond)
	case "SOLAR":
		e.StatusUntil["BURN"] = now.Add(2 * time.Second)
	case "EARTH":
		e.StatusUntil["WEAKEN"] = now.Add(2 * time.Second)
	case "AETHER":
		p.StatusUntil["FOCUS"] = now.Add(2500 * time.Millisecond)
	}
}

func skillElement(attackType string, p *Player) string {
	switch strings.ToUpper(attackType) {
	case "ENERGY", "AETHER":
		return "AETHER"
	case "WIND", "SOLAR", "EARTH":
		return strings.ToUpper(attackType)
	}
	switch p.ensureStyle() {
	case StyleWindStep:
		return "WIND"
	case StyleDawnFist:
		return "SOLAR"
	case StyleIronGuard:
		return "EARTH"
	case StyleCelestialFlow:
		return "AETHER"
	}
	return "SOLAR"
}

func elementBonus(el string) float64 {
	switch el {
	case "WIND":
		return 1.04
	case "SOLAR":
		return 1.03
	case "EARTH":
		return 1.03
	case "AETHER":
		return 1.04
	}
	return 1
}

func (w *WorldState) grantMasteryRewards(p *Player, style string) {
	lv := styleMasteryLevel(p.ensureLog().StyleMastery[style])
	log := p.ensureLog()
	if lv >= 3 {
		log.Flags["style_passive_"+style] = true
	}
	if lv >= 5 && !log.Flags["mastery_sp_"+style] {
		log.Flags["mastery_sp_"+style] = true
		p.SkillPoints++
	}
	if lv >= 5 {
		log.Flags["style_aura_"+style] = true
		if !containsID(log.Cosmetics, "aura-"+strings.ToLower(style)) {
			log.Cosmetics = append(log.Cosmetics, "aura-"+strings.ToLower(style))
		}
	}
	if lv >= 8 && p.ComboMasteryReady {
		log.Flags["combo_variation"] = true
	}
}

func (w *WorldState) startTrainingQuiz(p *Player, obj InteractDef) [][]byte {
	qid := obj.QuestionID
	if qid == "" {
		qid = "q-sub-7-3"
	}
	idx := -1
	for i, q := range questionCatalog {
		if q.ID == qid {
			idx = i
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeInteract, "question")
	}
	log := p.ensureLog()
	log.Quiz = QuizSession{QuestID: "ctq001", Index: idx, Active: true}
	p.markDirty()
	w.persist(p)
	q := questionCatalog[idx]
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "training-quiz", TargetID: obj.ID, Title: "Control Your Power", Speaker: "Arena Master Kael",
		Text: "7 - 3 = ?",
		Question: &QuestionOut{
			ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices, ToID: p.ID,
		},
	})}
}

func (w *WorldState) answerTrainingQuiz(p *Player, in EducationAnswerIn) [][]byte {
	log := p.ensureLog()
	if !log.Quiz.Active || log.Quiz.QuestID != "ctq001" {
		return nil
	}
	q := questionCatalog[log.Quiz.Index]
	log.Quiz.Active = false
	if in.Choice == q.Correct {
		w.grantEducationBonus(p)
		p.credit("ANSWER", q.ID, 1)
		log.Flags["training_edu_bonus"] = true
		log.KnowledgePoints++
		w.persist(p)
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: true, Explain: q.Explain, ToID: p.ID,
		}), marshal(TypeProgressionState, p.progressionView())}
	}
	w.persist(p)
	return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
		Correct: false, Explain: q.Explain, ToID: p.ID,
	})}
}

func formDisplayName(def TransformDef) string {
	if def.StoryName != "" {
		return def.StoryName
	}
	if n := formStoryName[def.ID]; n != "" {
		return n
	}
	return def.Name
}

func chapterTwoDone(p *Player) bool {
	log := p.ensureLog()
	return log.Flags["chapter_ch02_complete"] || log.Flags["ch02_complete"] || log.Flags["story_chapter_2_complete"] || log.Flags["chapter2_complete"]
}

func (e *Enemy) statusActive(kind string, now time.Time) bool {
	if e.StatusUntil == nil {
		return false
	}
	t, ok := e.StatusUntil[kind]
	return ok && now.Before(t)
}

func (p *Player) resistStatus(kind string) float64 {
	p.ensurePhase20()
	v := p.ensureLog().StatusResist[kind]
	v += float64(p.SpentDEF) * 0.01
	if v > 0.45 {
		v = 0.45
	}
	return v
}

func (p *Player) loadoutView() []string {
	out := append([]string{}, p.LoadoutSkills...)
	for len(out) < 4 {
		out = append(out, "")
	}
	if len(out) > 4 {
		out = out[:4]
	}
	ult := p.LoadoutUlt
	if ult == "" {
		ult = "celestial_impact"
	}
	return append(out, ult)
}

func (g EquipmentSet) hasItem(id string) bool {
	return g.HEAD == id || g.BODY == id || g.LEGS == id || g.WEAPON == id || g.ACCESSORY_1 == id || g.ACCESSORY_2 == id
}
