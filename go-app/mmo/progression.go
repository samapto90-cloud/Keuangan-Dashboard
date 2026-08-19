package mmo

import (
	"log"
	"time"

	_ "embed"
)

//go:embed data/progression.json
var progressionJSON []byte

//go:embed data/skillTree.json
var skillTreeJSON []byte

//go:embed data/transformations.json
var transformationsJSON []byte

//go:embed data/transformationQuests.json
var transformationQuestsJSON []byte

//go:embed data/combo.json
var comboJSON []byte

type ProgressionCfg struct {
	MaxLevel                int      `json:"maxLevel"`
	AttributePointsPerLevel int      `json:"attributePointsPerLevel"`
	SkillPointsPerLevel     int      `json:"skillPointsPerLevel"`
	BaseExp                 int      `json:"baseExp"`
	ExpPerLevel             int      `json:"expPerLevel"`
	HPPerLevel              int      `json:"hpPerLevel"`
	HPPerVitality           int      `json:"hpPerVitality"`
	AttackPerStrength       float64  `json:"attackPerStrength"`
	DefensePerDefense       float64  `json:"defensePerDefense"`
	MovePerAgility          float64  `json:"movePerAgility"`
	AttackSpeedPerAgility   float64  `json:"attackSpeedPerAgility"`
	EnergyPerEnergy         int      `json:"energyPerEnergy"`
	EnergyDamagePerEnergy   float64  `json:"energyDamagePerEnergy"`
	StarterSkills           []string `json:"starterSkills"`
	EducationExpBonusPct    float64  `json:"educationExpBonusPct"`
	EducationExpBonusSec    float64  `json:"educationExpBonusSec"`
}

type SkillTreeFile struct {
	Branches []SkillBranch `json:"branches"`
}

type SkillBranch struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Nodes []SkillTreeNode `json:"nodes"`
}

type SkillTreeNode struct {
	ID            string `json:"id"`
	SkillID       string `json:"skillId"`
	Cost          int    `json:"cost"`
	RequiredLevel int    `json:"requiredLevel"`
	Prerequisite  string `json:"prerequisite"`
	FormID        string `json:"formId"`
}

type TransformDef struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	ShortName       string  `json:"shortName"`
	Visual          string  `json:"visual"`
	Level           int     `json:"level"`
	EnergyCost      int     `json:"energyCost"`
	Duration        float64 `json:"duration"`
	DrainPerSec     float64 `json:"drainPerSec"`
	Cooldown        float64 `json:"cooldown"`
	AttackPct       float64 `json:"attackPct"`
	DefensePct      float64 `json:"defensePct"`
	EnergyPct       float64 `json:"energyPct"`
	RequiredLevel   int     `json:"requiredLevel"`
	RequiredQuest   string  `json:"requiredQuest"`
	RequiredChapter string  `json:"requiredChapter"`
	AuraColor       string  `json:"auraColor"`
	Particles       string  `json:"particles"`
	StoryName       string  `json:"storyName"`
	Passive         string  `json:"passive"`
	SpecialSkill    string  `json:"specialSkill"`
}

type TransformQuestDef struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	FormID          string   `json:"formId"`
	RequiredLevel   int      `json:"requiredLevel"`
	RequiredChapter string   `json:"requiredChapter"`
	RequiredFlags   []string `json:"requiredFlags"`
	RequiredKill    string   `json:"requiredKill"`
}

type ComboCfg struct {
	WindowMs          int      `json:"windowMs"`
	MaxHits           int      `json:"maxHits"`
	Finisher          []string `json:"finisher"`
	FinisherDamage    int      `json:"finisherDamage"`
	FinisherKnockback float64  `json:"finisherKnockback"`
	LightDamage       []int    `json:"lightDamage"`
	HeavyDamage       []int    `json:"heavyDamage"`
}

var (
	progressionCfg   ProgressionCfg
	skillTreeFile    SkillTreeFile
	skillNodeByID    = map[string]SkillTreeNode{}
	transformCatalog []TransformDef
	transformByID    = map[string]TransformDef{}
	awakenCatalog    []TransformQuestDef
	comboCfg         ComboCfg
)

func init() {
	mustJSON("progression.json", progressionJSON, &progressionCfg)
	if progressionCfg.MaxLevel < 1 {
		progressionCfg.MaxLevel = 60
	}
	if progressionCfg.AttributePointsPerLevel < 1 {
		progressionCfg.AttributePointsPerLevel = 3
	}
	if progressionCfg.SkillPointsPerLevel < 1 {
		progressionCfg.SkillPointsPerLevel = 1
	}
	mustJSON("skillTree.json", skillTreeJSON, &skillTreeFile)
	for _, b := range skillTreeFile.Branches {
		for _, n := range b.Nodes {
			skillNodeByID[n.ID] = n
		}
	}
	mustJSON("transformations.json", transformationsJSON, &transformCatalog)
	for _, f := range transformCatalog {
		transformByID[f.ID] = f
	}
	mustJSON("transformationQuests.json", transformationQuestsJSON, &awakenCatalog)
	mustJSON("combo.json", comboJSON, &comboCfg)
	if comboCfg.MaxHits < 1 {
		comboCfg.MaxHits = 4
	}
	if comboCfg.WindowMs < 1 {
		comboCfg.WindowMs = 800
	}
	log.Printf("mmo progression cap=%d forms=%d treeNodes=%d", progressionCfg.MaxLevel, len(transformCatalog), len(skillNodeByID))
	ensurePhase25Starters()
}

func expToNext(level int) int {
	if level < 1 {
		level = 1
	}
	base := progressionCfg.BaseExp
	if base <= 0 {
		base = 80
	}
	per := progressionCfg.ExpPerLevel
	if per <= 0 {
		per = 40
	}
	return base + level*per
}

func comboWindow() time.Duration {
	return time.Duration(comboCfg.WindowMs) * time.Millisecond
}

func (p *Player) hasSkill(id string) bool {
	for _, s := range p.UnlockedSkills {
		if s == id {
			return true
		}
	}
	for _, s := range progressionCfg.StarterSkills {
		if s == id {
			return true
		}
	}
	return false
}

func (p *Player) unlockSkill(id string) {
	if p.hasSkill(id) {
		return
	}
	p.UnlockedSkills = append(p.UnlockedSkills, id)
}

func (p *Player) hasForm(id string) bool {
	if id == "" || id == "normal" {
		return true
	}
	for _, f := range p.UnlockedForms {
		if f == id {
			return true
		}
	}
	log := p.ensureLog()
	if id == "aura-1" && (log.Flags["aura_1_unlocked"] || log.Flags["awakening_1"] || chapterTwoDone(p)) {
		return true
	}
	if id == "aura-2" && log.Flags["aura_2_unlocked"] {
		return true
	}
	if id == "aura-3" && log.Flags["aura_3_unlocked"] {
		return true
	}
	if id == "celestial-4" && log.Flags["celestial_4_unlocked"] {
		return true
	}
	return false
}

func (p *Player) unlockForm(id string) {
	if p.hasForm(id) && containsID(p.UnlockedForms, id) {
		return
	}
	if !containsID(p.UnlockedForms, id) {
		p.UnlockedForms = append(p.UnlockedForms, id)
	}
	log := p.ensureLog()
	switch id {
	case "aura-1":
		log.Flags["aura_1_unlocked"] = true
	case "aura-2":
		log.Flags["aura_2_unlocked"] = true
	case "aura-3":
		log.Flags["aura_3_unlocked"] = true
	case "celestial-4":
		log.Flags["celestial_4_unlocked"] = true
	}
}

func (p *Player) powerRating() int {
	n := p.Level*10 + p.Strength*2 + p.Defense*2 + p.Agility + p.EnergyPower + p.SpentVIT*2 + len(p.UnlockedSkills)*5 + p.EquipAttack
	n += len(p.UnlockedForms) * 20
	if p.FormID != "" && p.FormID != "normal" {
		n += 15
	}
	if n < 1 {
		n = 1
	}
	return n
}

func (w *WorldState) scaledExp(p *Player, amount int) int {
	if amount <= 0 || p == nil {
		return 0
	}
	if p.ExpBoostUntil.After(time.Now()) && p.ExpBoostPct > 0 {
		amount = int(float64(amount)*(1+p.ExpBoostPct) + 0.5)
	}
	return amount
}

func (w *WorldState) grantExp(p *Player, amount int) [][]byte {
	if amount <= 0 || !p.alive() {
		return nil
	}
	amount = w.scaledExp(p, amount)
	if p.Level >= progressionCfg.MaxLevel {
		return nil
	}
	p.Exp += amount
	var events [][]byte
	for p.Exp >= p.ExpToNext && p.Level < progressionCfg.MaxLevel {
		p.Exp -= p.ExpToNext
		from := p.Level
		p.Level++
		p.ExpToNext = expToNext(p.Level)
		p.AttributePoints += progressionCfg.AttributePointsPerLevel
		p.SkillPoints += progressionCfg.SkillPointsPerLevel
		if progressionCfg.HPPerLevel > 0 {
			p.BaseMaxHP += progressionCfg.HPPerLevel
		}
		p.applyDerived()
		p.HP = p.MaxHP
		p.Energy = p.MaxEnergy
		p.Stamina = p.MaxStamina
		events = append(events, marshal(TypePlayerLevelUp, LevelUpOut{
			PlayerID: p.ID, FromLevel: from, NewLevel: p.Level, MaxHP: p.MaxHP,
			AttributePoints: p.AttributePoints, SkillPoints: p.SkillPoints,
			Reward: "Attribute + Skill Point",
		}))
		events = append(events, marshal(TypePlayerStatsUpdated, p.statsView()))
		events = append(events, marshal(TypePowerRatingUpdated, map[string]any{"playerId": p.ID, "powerRating": p.powerRating()}))
	}
	w.persistGear(p)
	return events
}

func (w *WorldState) ApplyProgression(id string, env Envelope) [][]byte {
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
	case TypeAllocateAttribute:
		var in AttributeIn
		_ = unmarshal(env.Data, &in)
		return w.allocateAttribute(p, in.Stat)
	case TypeResetAttributes:
		return w.resetAttributes(p)
	case TypeUnlockSkill:
		var in SkillUnlockIn
		_ = unmarshal(env.Data, &in)
		return w.unlockSkillNode(p, in.NodeID, in.SkillID)
	case TypeRequestTransformation:
		var in TransformIn
		_ = unmarshal(env.Data, &in)
		return w.requestTransform(p, in.FormID)
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
	case TypeSetTransformation, TypeSetLevel, TypeSetSkillPoints, TypeSetEnergy, TypeSetCooldown, TypeUnlockTransform:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) allocateAttribute(p *Player, stat string) [][]byte {
	if p.AttributePoints < 1 {
		return rejectFor(p.ID, TypeAllocateAttribute, "points")
	}
	stat = normalizeAttr(stat)
	switch stat {
	case "STR":
		p.SpentSTR++
	case "DEF":
		p.SpentDEF++
	case "AGI":
		p.SpentAGI++
	case "ENG":
		p.SpentENG++
	case "VIT":
		p.SpentVIT++
	default:
		return rejectFor(p.ID, TypeAllocateAttribute, "stat")
	}
	p.AttributePoints--
	p.applyDerived()
	w.persistGear(p)
	return [][]byte{
		marshal(TypePlayerStatsUpdated, p.statsView()),
		marshal(TypeProgressionState, p.progressionView()),
		marshal(TypePowerRatingUpdated, map[string]any{"playerId": p.ID, "powerRating": p.powerRating()}),
	}
}

func (w *WorldState) resetAttributes(p *Player) [][]byte {
	now := time.Now()
	log := p.ensureLog()
	used := p.AttrResetUsed
	if used == 0 {
		used = log.AttrResetUsed
	}
	if used >= AttrResetFree && !log.Flags["reset_token"] && log.AttrResetUntil > 0 && now.UnixMilli() < log.AttrResetUntil {
		return rejectFor(p.ID, TypeResetAttributes, "cooldown")
	}
	spent := p.SpentSTR + p.SpentDEF + p.SpentAGI + p.SpentENG + p.SpentVIT
	p.AttributePoints += spent
	p.SpentSTR, p.SpentDEF, p.SpentAGI, p.SpentENG, p.SpentVIT = 0, 0, 0, 0, 0
	p.AttrResetUsed = used + 1
	log.AttrResetUsed = p.AttrResetUsed
	if p.AttrResetUsed >= AttrResetFree && !log.Flags["reset_token"] {
		log.AttrResetUntil = now.Add(AttrResetCooldown).UnixMilli()
	}
	if log.Flags["reset_token"] {
		log.Flags["reset_token"] = false
	}
	p.applyDerived()
	w.persistGear(p)
	w.persist(p)
	return [][]byte{marshal(TypePlayerStatsUpdated, p.statsView()), marshal(TypeProgressionState, p.progressionView())}
}

func (w *WorldState) unlockSkillNode(p *Player, nodeID, skillID string) [][]byte {
	node, ok := skillNodeByID[nodeID]
	if !ok {
		for _, n := range skillNodeByID {
			if n.SkillID == skillID {
				node, ok = n, true
				break
			}
		}
	}
	if !ok {
		return rejectFor(p.ID, TypeUnlockSkill, "node")
	}
	if p.Level < node.RequiredLevel {
		return rejectFor(p.ID, TypeUnlockSkill, "level")
	}
	cost := node.Cost
	if cost < 1 {
		cost = 1
	}
	if p.SkillPoints < cost {
		return rejectFor(p.ID, TypeUnlockSkill, "points")
	}
	if node.Prerequisite != "" {
		pre := skillNodeByID[node.Prerequisite]
		if pre.SkillID != "" && !p.hasSkill(pre.SkillID) {
			return rejectFor(p.ID, TypeUnlockSkill, "prerequisite")
		}
	}
	if p.hasSkill(node.SkillID) {
		return rejectFor(p.ID, TypeUnlockSkill, "owned")
	}
	p.SkillPoints -= cost
	p.unlockSkill(node.SkillID)
	if node.FormID != "" {
		p.unlockForm(node.FormID)
	}
	p.applyDerived()
	w.persistGear(p)
	w.persist(p)
	return [][]byte{
		marshal(TypeSkillUnlocked, map[string]any{"skillId": node.SkillID, "nodeId": node.ID, "toId": p.ID}),
		marshal(TypeProgressionState, p.progressionView()),
		marshal(TypePlayerStatsUpdated, p.statsView()),
	}
}

func (p *Player) progressionView() ProgressionView {
	p.ensurePhase20()
	log := p.ensureLog()
	forms := make([]FormView, 0, len(transformCatalog)+1)
	forms = append(forms, FormView{ID: "normal", Name: "NORMAL FORM", ShortName: "NORMAL", Unlocked: true, Active: p.FormID == "" || p.FormID == "normal"})
	for _, f := range transformCatalog {
		forms = append(forms, FormView{
			ID: f.ID, Name: f.Name, ShortName: f.ShortName, Visual: f.Visual,
			Unlocked: p.hasForm(f.ID), Active: p.FormID == f.ID, EnergyCost: f.EnergyCost, Level: f.RequiredLevel,
			StoryName: formDisplayName(f), Passive: f.Passive, Mastery: styleMasteryLevel(log.FormMastery[f.ID]),
		})
	}
	nodes := make([]SkillNodeView, 0, len(skillNodeByID))
	owned := map[string]bool{}
	for _, id := range p.UnlockedSkills {
		owned[id] = true
	}
	for _, id := range progressionCfg.StarterSkills {
		owned[id] = true
	}
	for _, b := range skillTreeFile.Branches {
		branchName := b.Name
		if b.ID == "COMBAT" {
			branchName = "OFFENSE"
		}
		for _, n := range b.Nodes {
			unlocked := p.hasSkill(n.SkillID)
			available := !unlocked && p.Level >= n.RequiredLevel && (n.Prerequisite == "" || owned[skillNodeByID[n.Prerequisite].SkillID] || p.hasSkill(skillNodeByID[n.Prerequisite].SkillID))
			sv := SkillNodeView{
				ID: n.ID, SkillID: n.SkillID, Branch: b.ID, BranchName: branchName, Cost: n.Cost, RequiredLevel: n.RequiredLevel,
				Prerequisite: n.Prerequisite, Unlocked: unlocked, Available: available,
			}
			if def, ok := skillCatalog[n.SkillID]; ok {
				sv.Name, sv.Description, sv.EnergyCost, sv.Cooldown, sv.Effect, sv.Damage, sv.Range = def.Name, def.Description, def.EnergyCost, def.Cooldown, def.Effect, def.Damage, def.Range
			} else {
				sv.Name = n.SkillID
			}
			nodes = append(nodes, sv)
		}
	}
	state := p.TransformState
	if state == "" {
		state = "NORMAL"
	}
	idle := state == "NORMAL" || state == "COOLDOWN"
	left := AttrResetFree - p.AttrResetUsed
	if left < 0 {
		left = 0
	}
	var train *TrainingRecord
	if log.Training.Hits > 0 {
		cp := log.Training
		train = &cp
	}
	return ProgressionView{
		PlayerID: p.ID, Level: p.Level, Exp: p.Exp, ExpToNext: p.ExpToNext,
		AttributePoints: p.AttributePoints, SkillPoints: p.SkillPoints,
		SpentSTR: p.SpentSTR, SpentDEF: p.SpentDEF, SpentAGI: p.SpentAGI, SpentENG: p.SpentENG, SpentVIT: p.SpentVIT,
		UnlockedSkills: append([]string{}, p.UnlockedSkills...),
		UnlockedForms:  append([]string{}, p.UnlockedForms...),
		FormID:         p.FormID, TransformState: state, TransEnergy: p.TransEnergy, MaxTransEnergy: p.MaxTransEnergy,
		PowerRating: p.powerRating(), Forms: forms, Nodes: nodes,
		TransformReady: idle && time.Now().After(p.TransformCDUntil),
		CombatStyle: p.ensureStyle(), StyleMastery: log.StyleMastery, FormMastery: log.FormMastery,
		UltCharge: p.UltCharge, MaxUltCharge: MaxUltCharge, Loadout: p.loadoutView(),
		Builds: append([]CombatBuild{}, log.Builds...), ActiveBuild: log.ActiveBuild,
		CombatRating: p.combatRating(), AttrResetLeft: left, Training: train, Styles: styleCatalog,
	}
}

func (w *WorldState) ProgressionFor(id string) ProgressionView {
	w.mu.Lock()
	defer w.mu.Unlock()
	if p := w.players[id]; p != nil {
		return p.progressionView()
	}
	return ProgressionView{}
}

func (w *WorldState) grantEducationBonus(p *Player) {
	sec := progressionCfg.EducationExpBonusSec
	if sec <= 0 {
		sec = 600
	}
	p.ExpBoostPct = progressionCfg.EducationExpBonusPct
	if p.ExpBoostPct <= 0 {
		p.ExpBoostPct = 0.05
	}
	p.ExpBoostUntil = time.Now().Add(time.Duration(sec) * time.Second)
	p.ensureLog().Flags["education_challenge_1"] = true
	p.ensureLog().EduToken++
	p.markDirty()
}
