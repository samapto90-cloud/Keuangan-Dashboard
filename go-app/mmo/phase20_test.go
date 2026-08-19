package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPhase20LevelCapAndTiers(t *testing.T) {
	if progressionCfg.MaxLevel < 60 {
		t.Fatalf("maxLevel %d want 60+", progressionCfg.MaxLevel)
	}
	w, p, _ := testWorldWithPlayer()
	if p.Level != 1 {
		t.Fatalf("start level %d", p.Level)
	}
	p.Level = 10
	if p.Level != 10 {
		t.Fatal("level 10")
	}
	p.Level = 30
	p.Level = 60
	w.grantExp(p, 99999)
	if p.Level > progressionCfg.MaxLevel {
		t.Fatalf("cap broken %d", p.Level)
	}
	p.Level = progressionCfg.MaxLevel
	w.grantExp(p, 500)
	if p.Level != 60 {
		t.Fatalf("level 60 cap %d", p.Level)
	}
}

func TestPhase20AttributesAliasAndSecurity(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	p.AttributePoints = 2
	w.ApplyProgression(p.ID, Envelope{Type: TypeAllocateAttribute, Data: []byte(`{"stat":"POWER"}`)})
	if p.SpentSTR != 1 {
		t.Fatal("POWER harus map ke STR")
	}
	w.ApplyProgression(p.ID, Envelope{Type: TypeAllocateAttribute, Data: []byte(`{"stat":"VITALITY"}`)})
	if p.SpentVIT != 1 {
		t.Fatal("VITALITY harus map ke VIT")
	}
	evs := w.ApplyProgression(p.ID, Envelope{Type: TypeAllocateAttribute, Data: []byte(`{"stat":"STR"}`)})
	if !rejectAction(evs, TypeAllocateAttribute) {
		t.Fatal("poin habis harus ditolak")
	}
	hp := p.MaxHP
	p.AttributePoints = 1
	w.ApplyProgression(p.ID, Envelope{Type: TypeAllocateAttribute, Data: []byte(`{"stat":"VIT"}`)})
	if p.MaxHP <= hp {
		t.Fatal("vitality harus menaikkan HP")
	}
}

func TestPhase20CombatStylesAndSwitch(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	evs := w.ApplyProgression(p.ID, Envelope{Type: TypeSetCombatStyle, Data: []byte(`{"styleId":"WIND_STEP"}`)})
	if rejectAction(evs, TypeSetCombatStyle) || p.CombatStyle != StyleWindStep {
		t.Fatal("switch style di luar combat")
	}
	p.InCombatUntil = time.Now().Add(time.Second)
	evs = w.ApplyProgression(p.ID, Envelope{Type: TypeSetCombatStyle, Data: []byte(`{"styleId":"IRON_GUARD"}`)})
	if !rejectAction(evs, TypeSetCombatStyle) {
		t.Fatal("switch style saat combat harus ditolak")
	}
}

func TestPhase20BuildsPersistAndCombatLock(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	p.AttributePoints = 3
	w.ApplyProgression(p.ID, Envelope{Type: TypeAllocateAttribute, Data: []byte(`{"stat":"STR"}`)})
	w.ApplyProgression(p.ID, Envelope{Type: TypeSetCombatStyle, Data: []byte(`{"styleId":"DAWN_FIST"}`)})
	w.ApplyProgression(p.ID, Envelope{Type: TypeSaveBuild, Data: []byte(`{"slot":0,"name":"Fist Master"}`)})
	w.ApplyProgression(p.ID, Envelope{Type: TypeSetCombatStyle, Data: []byte(`{"styleId":"WIND_STEP"}`)})
	w.ApplyProgression(p.ID, Envelope{Type: TypeSaveBuild, Data: []byte(`{"slot":1,"name":"Wind Runner"}`)})
	w.ApplyProgression(p.ID, Envelope{Type: TypeSetCombatStyle, Data: []byte(`{"styleId":"IRON_GUARD"}`)})
	w.ApplyProgression(p.ID, Envelope{Type: TypeSaveBuild, Data: []byte(`{"slot":2,"name":"Guardian"}`)})
	if p.ensureLog().Builds[0].Name != "Fist Master" || p.ensureLog().Builds[1].Name != "Wind Runner" || p.ensureLog().Builds[2].Name != "Guardian" {
		t.Fatalf("build names %+v", p.ensureLog().Builds)
	}
	p.InCombatUntil = time.Now().Add(2 * time.Second)
	if !rejectAction(w.ApplyProgression(p.ID, Envelope{Type: TypeLoadBuild, Data: []byte(`{"slot":0}`)}), TypeLoadBuild) {
		t.Fatal("load saat combat harus ditolak")
	}
	p.InCombatUntil = time.Time{}
	if rejectAction(w.ApplyProgression(p.ID, Envelope{Type: TypeLoadBuild, Data: []byte(`{"slot":0}`)}), TypeLoadBuild) {
		t.Fatal("load di luar combat")
	}
	if p.CombatStyle != StyleDawnFist {
		t.Fatalf("style %s", p.CombatStyle)
	}
	w.persist(p)
	w.persistGear(p)
	w.Remove(p)
	again := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 8)}
	again.initCombat()
	ok, _ := w.Add(again)
	if !ok {
		t.Fatal("reconnect")
	}
	if again.ensureLog().Builds[0].Name != "Fist Master" {
		t.Fatalf("build persist %v", again.ensureLog().Builds)
	}
}

func TestPhase20SkillCooldownEnergyAndDamage(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	p.X, p.Z = e.X, e.Z
	e.HP, e.MaxHP = 999, 999
	raw, _ := json.Marshal(map[string]any{"skillId": "power_strike", "targetId": e.ID, "timestamp": time.Now().UnixMilli(), "damage": 999999})
	hp := e.HP
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: raw})
	if rejectAction(evs, "PLAYER_SKILL") {
		t.Fatalf("skill 1 %s", string(evs[0]))
	}
	lost := hp - e.HP
	if lost < 1 || lost > 80 {
		t.Fatalf("server damage %d", lost)
	}
	evs = w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: raw})
	if !rejectAction(evs, "PLAYER_SKILL") {
		t.Fatal("cooldown harus menolak spam")
	}
	p.SkillCD = map[string]time.Time{}
	p.Energy = 0
	raw, _ = json.Marshal(map[string]any{"skillId": "energy_bolt", "targetId": e.ID, "timestamp": time.Now().UnixMilli()})
	evs = w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: raw})
	if !rejectAction(evs, "PLAYER_SKILL") {
		t.Fatal("energy habis harus menolak")
	}
}

func TestPhase20TransformationOverlay(t *testing.T) {
	if transformByID["aura-1"].Name != "AURA ASCENSION I" || formDisplayName(transformByID["aura-1"]) != "AWAKENED FORM" {
		t.Fatal("keep AURA name, overlay AWAKENED FORM")
	}
	if transformByID["celestial-4"].Name != "CELESTIAL AURA IV" {
		t.Fatal("celestial name")
	}
	w, p, _ := testWorldWithPlayer()
	p.Level = 10
	p.Energy = p.MaxEnergy
	p.ensureLog().Quests["tq001"] = &QuestLog{ID: "tq001", State: QuestClaimed}
	evs := w.ApplyProgression(p.ID, Envelope{Type: TypeRequestTransformation, Data: []byte(`{"formId":"aura-1"}`)})
	if hasType(evs, TypeTransformationRejected) || p.TransformState != "TRANSFORMING" {
		t.Fatal("awakened start")
	}
	p.unlockForm("aura-2")
	p.FormID = "aura-2"
	p.TransformState = "TRANSFORMED"
	p.TransEnergy = 0
	p.TransformUntil = time.Now().Add(time.Minute)
	evs = w.tickTransform(p, 0.2, time.Now())
	if !hasType(evs, TypeTransformationEnded) {
		t.Fatal("energy 0 auto end")
	}
	p.ensureLog().PendingCinematic = "cin-1"
	p.FormID = "normal"
	p.TransformState = "NORMAL"
	p.unlockForm("aura-1")
	p.Energy = p.MaxEnergy
	evs = w.ApplyProgression(p.ID, Envelope{Type: TypeRequestTransformation, Data: []byte(`{"formId":"aura-1"}`)})
	if !hasType(evs, TypeTransformationRejected) {
		t.Fatal("cinematic harus menolak transform")
	}
	p.ensureLog().PendingCinematic = ""
	p.FormID = "normal"
	p.TransformState = "NORMAL"
	p.Energy = p.MaxEnergy
	evs = w.ApplyProgression(p.ID, Envelope{Type: TypeRequestTransformation, Data: []byte(`{"formId":"celestial-4"}`)})
	if !hasType(evs, TypeTransformationRejected) {
		t.Fatal("celestial tanpa syarat")
	}
}

func TestPhase20EducationAndLandmines(t *testing.T) {
	if questionByID["q-add-4-3"].ID == "" || questionByID["q-add-4-3"].Correct != 1 {
		t.Fatal("keep q-add-4-3")
	}
	if questionByID["q-sub-7-3"].ID == "" || questionByID["q-sub-7-3"].Correct != 1 {
		t.Fatal("7-3 = 4 index 1")
	}
	if questByID["ctq001"].Title != "Control Your Power" {
		t.Fatal("training quest")
	}
	if comboRecipe() != "LLLH" {
		t.Fatalf("combo %s", comboRecipe())
	}
	w, p, _ := testWorldWithPlayer()
	p.X, p.Z = 12.4, 7.2
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"training-quiz","kind":"training-quiz"}`)})
	if !payloadHas(evs, TypeInteractResult, "q-sub-7-3") {
		t.Fatalf("training quiz %s", string(evs[0]))
	}
	w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: []byte(`{"questionId":"q-sub-7-3","choice":1}`)})
	if !p.ensureLog().Flags["training_edu_bonus"] {
		t.Fatal("educational bonus")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
}

func TestPhase20RateLimitAndBlock(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	p.X, p.Z = e.X, e.Z
	rejected := false
	for i := 0; i < 28; i++ {
		raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
		p.AttackCDUntil = time.Time{}
		evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
		if rejectAction(evs, "PLAYER_ATTACK") || rejectAction(evs, TypePlayerAttack) {
			rejected = true
			break
		}
	}
	if !rejected {
		t.Fatal("combat rate limit")
	}
	p2 := &Player{ID: "p_block", Name: "Raka", send: make(chan []byte, 8)}
	p2.initCombat()
	w.Add(p2)
	w.ApplyCombat(p2.ID, Envelope{Type: TypePlayerBlock, Data: []byte(`{"on":true}`)})
	if !p2.Blocking {
		t.Fatal("block")
	}
}
