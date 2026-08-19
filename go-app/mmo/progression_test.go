package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLevelUpGrantsPoints(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	from := p.Level
	need := p.ExpToNext
	w.grantExp(p, need)
	if p.Level != from+1 {
		t.Fatalf("level %d want %d", p.Level, from+1)
	}
	if p.AttributePoints < progressionCfg.AttributePointsPerLevel {
		t.Fatalf("attribute points %d", p.AttributePoints)
	}
	if p.SkillPoints < progressionCfg.SkillPointsPerLevel {
		t.Fatalf("skill points %d", p.SkillPoints)
	}
}

func TestUnlockSkillRejectThenOk(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	raw, _ := json.Marshal(SkillUnlockIn{NodeID: "combo_master"})
	evs := w.ApplyProgression(p.ID, Envelope{Type: TypeUnlockSkill, Data: raw})
	if !rejectAction(evs, TypeUnlockSkill) {
		t.Fatal("skill terkunci harus ditolak")
	}
	p.Level = 1
	p.SkillPoints = 1
	raw, _ = json.Marshal(SkillUnlockIn{NodeID: "enhanced_punch"})
	evs = w.ApplyProgression(p.ID, Envelope{Type: TypeUnlockSkill, Data: raw})
	if rejectAction(evs, TypeUnlockSkill) || !p.hasSkill("enhanced_punch") {
		t.Fatal("enhanced_punch harus terbuka")
	}
}

func TestComboFinisherRecipe(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	e.HP, e.MaxHP = 999, 999
	p.X, p.Z = e.X, e.Z
	var last AttackResult
	for _, kind := range []string{"punch", "punch", "punch", "kick"} {
		p.AttackCDUntil = time.Time{}
		raw, _ := json.Marshal(AttackIn{AttackType: kind, TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
		evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
		for _, ev := range evs {
			var env Envelope
			_ = json.Unmarshal(ev, &env)
			if env.Type == TypeAttackResult {
				_ = json.Unmarshal(env.Data, &last)
			}
		}
	}
	if !last.Finisher || last.ComboHits != 4 {
		t.Fatalf("finisher want 4 hits, got hits=%d finisher=%v type=%s", last.ComboHits, last.Finisher, last.AttackType)
	}
}

func TestTransformRejectLowLevel(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	p.Level = 5
	p.Energy = 100
	raw, _ := json.Marshal(TransformIn{FormID: "aura-1"})
	evs := w.ApplyProgression(p.ID, Envelope{Type: TypeRequestTransformation, Data: raw})
	if !hasType(evs, TypeTransformationRejected) {
		t.Fatal("level 5 Aura I harus ditolak")
	}
}

func TestTransformUnlockAndStart(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	p.Level = 10
	p.Energy = p.MaxEnergy
	p.ensureLog().Quests["tq001"] = &QuestLog{ID: "tq001", State: QuestClaimed}
	raw, _ := json.Marshal(TransformIn{FormID: "aura-1"})
	evs := w.ApplyProgression(p.ID, Envelope{Type: TypeRequestTransformation, Data: raw})
	if hasType(evs, TypeTransformationRejected) {
		t.Fatalf("Aura I harus mulai: %s", string(evs[0]))
	}
	if p.TransformState != "TRANSFORMING" || p.FormID != "aura-1" {
		t.Fatalf("state=%s form=%s", p.TransformState, p.FormID)
	}
	w.tickTransform(p, 0.05, time.Now().Add(900*time.Millisecond))
	if p.TransformState != "TRANSFORMED" {
		t.Fatalf("setelah 800ms harus TRANSFORMED, got %s", p.TransformState)
	}
}

func TestTransformEnergyEnds(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	p.Level = 10
	p.unlockForm("aura-1")
	p.FormID = "aura-1"
	p.TransformState = "TRANSFORMED"
	p.TransEnergy = 0
	p.TransformUntil = time.Now().Add(time.Minute)
	evs := w.tickTransform(p, 0.2, time.Now())
	if !hasType(evs, TypeTransformationEnded) {
		t.Fatal("energy habis harus mengakhiri transformasi")
	}
	if p.FormID != "normal" {
		t.Fatalf("form %s", p.FormID)
	}
}

func TestTransformMultiplayerSnapshot(t *testing.T) {
	w, a, _ := testWorldWithPlayer()
	b := &Player{ID: "p_b", Name: "Sinta", send: make(chan []byte, 8)}
	b.initCombat()
	w.Add(b)
	a.Level = 10
	a.unlockForm("aura-1")
	a.Energy = a.MaxEnergy
	raw, _ := json.Marshal(TransformIn{FormID: "aura-1"})
	w.ApplyProgression(a.ID, Envelope{Type: TypeRequestTransformation, Data: raw})
	if a.Snap().FormID != "aura-1" {
		t.Fatal("snapshot A harus memuat formId")
	}
	cheat, _ := json.Marshal(TransformIn{FormID: "celestial-4"})
	evs := w.ApplyProgression(b.ID, Envelope{Type: TypeSetTransformation, Data: cheat})
	if !rejectAction(evs, TypeSetTransformation) {
		t.Fatal("B tidak boleh SET_TRANSFORMATION")
	}
	if a.FormID != "aura-1" {
		t.Fatal("B tidak boleh mengubah state A")
	}
}

func TestProgressionCheatsRejected(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	if !rejectAction(w.ApplyProgression(p.ID, Envelope{Type: TypeSetLevel, Data: []byte(`{"level":99}`)}), TypeSetLevel) {
		t.Fatal("SET_LEVEL harus ditolak")
	}
	if !rejectAction(w.ApplyProgression(p.ID, Envelope{Type: TypeSetSkillPoints, Data: []byte(`{"skillPoints":99}`)}), TypeSetSkillPoints) {
		t.Fatal("SET_SKILL_POINTS harus ditolak")
	}
	p.X, p.Z = e.X, e.Z
	e.HP, e.MaxHP = 999, 999
	raw, _ := json.Marshal(map[string]any{"skillId": "power_strike", "targetId": e.ID, "timestamp": time.Now().UnixMilli(), "damage": 999999, "energyCost": 0})
	hp := e.HP
	w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: raw})
	if hp-e.HP > 80 {
		t.Fatalf("damage client tidak boleh dipakai, lost=%d", hp-e.HP)
	}
	raw, _ = json.Marshal(TransformIn{FormID: "celestial-4"})
	evs := w.ApplyProgression(p.ID, Envelope{Type: TypeRequestTransformation, Data: raw})
	if !hasType(evs, TypeTransformationRejected) {
		t.Fatal("celestial-4 tanpa syarat harus ditolak")
	}
}

func TestProgressionPersistence(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	p.Level = 20
	p.Exp = 40
	p.AttributePoints = 6
	p.SkillPoints = 2
	p.unlockForm("aura-2")
	p.unlockSkill("energy_control")
	w.persistGear(p)
	w.persist(p)
	w.Remove(p)
	again := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 8)}
	again.initCombat()
	ok, _ := w.Add(again)
	if !ok {
		t.Fatal("reconnect add")
	}
	if again.Level != 20 || !again.hasForm("aura-2") || !again.hasSkill("energy_control") {
		t.Fatalf("persist gagal level=%d forms=%v skills=%v", again.Level, again.UnlockedForms, again.UnlockedSkills)
	}
}

func TestAllocateAndResetAttributes(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	p.AttributePoints = 2
	w.ApplyProgression(p.ID, Envelope{Type: TypeAllocateAttribute, Data: []byte(`{"stat":"STR"}`)})
	w.ApplyProgression(p.ID, Envelope{Type: TypeAllocateAttribute, Data: []byte(`{"stat":"VIT"}`)})
	if p.SpentSTR != 1 || p.SpentVIT != 1 || p.AttributePoints != 0 {
		t.Fatal("alokasi atribut gagal")
	}
	w.ApplyProgression(p.ID, Envelope{Type: TypeResetAttributes, Data: []byte(`{}`)})
	if p.SpentSTR != 0 || p.AttributePoints != 2 {
		t.Fatal("reset atribut harus mengembalikan poin")
	}
}

func hasType(evs [][]byte, typ string) bool {
	for _, ev := range evs {
		var env Envelope
		if json.Unmarshal(ev, &env) == nil && env.Type == typ {
			return true
		}
	}
	return false
}
