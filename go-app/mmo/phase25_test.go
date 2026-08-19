package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPhase25LandminesKept(t *testing.T) {
	if comboRecipe() != "LLLH" {
		t.Fatalf("combo %s", comboRecipe())
	}
	if transformByID["aura-1"].Name != "AURA ASCENSION I" || formDisplayName(transformByID["aura-1"]) != "AWAKENED FORM" {
		t.Fatal("keep AURA ASCENSION I / AWAKENED FORM")
	}
	if questionByID["q-add-4-3"].Correct != 1 || questionByID["q-apel-3-2"].Correct != 1 {
		t.Fatal("keep q-add-4-3 and q-apel-3-2")
	}
	if questionByID["q1"].Correct != 1 {
		t.Fatal("q1 2+3=5 index B")
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatal("keep Dawn City")
	}
	if mountByID["wind-runner"].Name != "Wind Runner" || mountByID["wind-runner"].Speed != 1.45 {
		t.Fatal("keep wind-runner")
	}
	if _, ok := skillCatalog["power_strike"]; !ok {
		t.Fatal("power_strike")
	}
	if _, ok := skillCatalog["celestial_impact"]; !ok {
		t.Fatal("celestial_impact")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
	if questByID["mq004"].Title != "Jalan Menuju Hutan" {
		t.Fatal("mq004")
	}
	if dialogueCatalog["pak_jaga"].Text != "Ati-ati, Le! Ana siluman teka saka alas!" {
		t.Fatal("pak jaga")
	}
	if len(storyChapterCatalog) != 10 {
		t.Fatalf("chapters %d", len(storyChapterCatalog))
	}
	if silumanByGuard["ragha"].Name != "Jaladara" {
		t.Fatal("jaladara")
	}
	if questByID["tq001"].Title == "Ngudi Kekuwatan." {
		t.Fatal("do not retitle tq001")
	}
	if comboCfg.MaxHits != 4 {
		t.Fatalf("LLLH maxHits %d", comboCfg.MaxHits)
	}
}

func TestPhase25OverlayContent(t *testing.T) {
	if questByID["nq-combat-1"].Title != "Latihan Dhasar." {
		t.Fatal("latihan dhasar")
	}
	if questByID["nq-combo-1"].Title != "Serangan Berantai." {
		t.Fatal("combo quest")
	}
	if questByID["nq-dodge-1"].Title != "Langkah Cepet." {
		t.Fatal("dodge quest")
	}
	if questByID["nq-ascend-1"].Title != "Ngudi Kekuwatan." {
		t.Fatal("ngudi")
	}
	if questByID["nq-ascend-2"].Title != "Awakmu Wis Siap?" {
		t.Fatal("awakmu")
	}
	if storyDialogue["p25.karya.punch"].JV != "Saiki coba mukul kayu iku." {
		t.Fatal("tutorial jawa")
	}
	if storyDialogue["p25.karya.ngudi"].JV != "Kowe kudu nglatih awakmu luwih sregep, Le." {
		t.Fatal("ngudi line")
	}
	if _, ok := skillCatalog["dawn_wave"]; !ok {
		t.Fatal("dawn_wave")
	}
	if _, ok := skillCatalog["light_burst"]; !ok {
		t.Fatal("light_burst")
	}
	if resolvePhase25Skill("light_strike") != "power_strike" || resolvePhase25Skill("final_light") != "celestial_impact" {
		t.Fatal("skill alias")
	}
	if AirComboMaxHits != 6 {
		t.Fatal("air combo 6")
	}
}

func TestPhase25PunchKickComboCharge(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	e.HP, e.MaxHP = 500, 500
	p.X, p.Z = e.X, e.Z
	punch, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	if rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: punch}), "PLAYER_ATTACK") {
		t.Fatal("punch")
	}
	p.AttackCDUntil = time.Time{}
	if rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: punch}), "PLAYER_ATTACK") {
		t.Fatal("punch 2")
	}
	p.AttackCDUntil = time.Time{}
	kick, _ := json.Marshal(AttackIn{AttackType: "kick", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	if rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: kick}), "PLAYER_ATTACK") {
		t.Fatal("kick")
	}
	if p.ensureLog().CombatHits < 1 {
		t.Fatal("attack credit")
	}
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerCharge, Data: []byte(`{"on":true}`)})
	if rejectAction(evs, TypePlayerCharge) || !p.Charging || p.CombatState != "CHARGE" {
		t.Fatal("charge")
	}
	w.mu.Lock()
	w.hitPlayer("src", "dummy", p, 20, 0, 0, "PHYSICAL")
	w.mu.Unlock()
	if p.Charging {
		t.Fatal("charge interrupt")
	}
}

func TestPhase25DodgeBlockCounter(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	p.X, p.Z = e.X, e.Z
	raw, _ := json.Marshal(DodgeIn{Timestamp: time.Now().UnixMilli(), Yaw: 0, AX: 1, AZ: 0})
	if rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypePlayerDodge, Data: raw}), "PLAYER_DODGE") {
		t.Fatal("dodge")
	}
	if p.CombatState != "DODGING" {
		t.Fatal("dodge state")
	}
	p.DodgeUntil = time.Time{}
	p.DodgeCDUntil = time.Time{}
	p.CombatState = "IDLE"
	w.ApplyCombat(p.ID, Envelope{Type: TypePlayerBlock, Data: []byte(`{"on":true}`)})
	if !p.Blocking {
		t.Fatal("block")
	}
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerCounter, Data: []byte(`{}`)})
	if rejectAction(evs, TypePlayerCounter) {
		t.Fatal("counter window")
	}
	p.PerfectDodgeUntil = time.Now().Add(time.Second)
	p.IFrameUntil = time.Now().Add(200 * time.Millisecond)
	w.mu.Lock()
	out := w.hitPlayer("e1", "enemy", p, 30, 0, 0, "PHYSICAL")
	w.mu.Unlock()
	if len(out) == 0 {
		t.Fatal("perfect dodge event")
	}
	if p.ensureLog().PerfectDodges < 1 {
		t.Fatal("perfect dodge credit")
	}
}

func TestPhase25SkillsAndUltimate(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	e.HP, e.MaxHP = 800, 800
	p.X, p.Z = e.X, e.Z
	p.Energy = p.MaxEnergy
	raw, _ := json.Marshal(SkillIn{SkillID: "power_strike", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	if rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: raw}), "PLAYER_SKILL") {
		t.Fatal("light strike")
	}
	p.Energy = p.MaxEnergy
	alias, _ := json.Marshal(SkillIn{SkillID: "light_strike", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	p.SkillCD["power_strike"] = time.Time{}
	if rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: alias}), "PLAYER_SKILL") {
		t.Fatal("alias light_strike")
	}
	p.Energy = p.MaxEnergy
	wave, _ := json.Marshal(SkillIn{SkillID: "dawn_wave", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: wave})
	if rejectAction(evs, "PLAYER_SKILL") {
		t.Fatalf("dawn wave %s", string(evs[0]))
	}
	p.Energy = p.MaxEnergy
	p.Level = 30
	p.unlockSkill("celestial_impact")
	ult, _ := json.Marshal(SkillIn{SkillID: "celestial_impact", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	if rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: ult}), "PLAYER_SKILL") {
		t.Fatal("ultimate")
	}
}

func TestPhase25TransformationLocked(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	p.Level = 60
	p.Energy = 999
	evs := w.requestTransform(p, "celestial-4")
	ok := false
	for _, ev := range evs {
		var env Envelope
		if json.Unmarshal(ev, &env) != nil {
			continue
		}
		if env.Type == TypeTransformationRejected {
			ok = true
		}
	}
	if !ok {
		t.Fatal("transform IV without unlock must reject")
	}
}

func TestPhase25CheatsRejected(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	p.X, p.Z = e.X, e.Z
	if !rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypeSetDamage, Data: []byte(`{"damage":999999}`)}), TypeSetDamage) {
		t.Fatal("SET_DAMAGE")
	}
	if !rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypeSetEnergy, Data: []byte(`{"energy":999}`)}), TypeSetEnergy) {
		t.Fatal("SET_ENERGY")
	}
	if !rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypeSetCooldown, Data: []byte(`{"cooldown":0}`)}), TypeSetCooldown) {
		t.Fatal("SET_COOLDOWN")
	}
	if !rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypeUnlockTransform, Data: []byte(`{"formId":"celestial-4"}`)}), TypeUnlockTransform) {
		t.Fatal("UNLOCK_TRANSFORM")
	}
	p.X, p.Z = e.X+40, e.Z+40
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	if !rejectAction(w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw}), "PLAYER_ATTACK") {
		t.Fatal("teleport attack")
	}
}

func TestPhase25ComboQuestAndReconnect(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	e.HP, e.MaxHP = 800, 800
	p.X, p.Z = e.X, e.Z
	w.acceptQuest(p, "nq-combo-1")
	w.acceptQuest(p, "nq-combat-1")
	for i := 0; i < 5; i++ {
		p.AttackCDUntil = time.Time{}
		raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
		w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	}
	if p.CombatStreak < 5 {
		t.Fatalf("streak %d", p.CombatStreak)
	}
	if q := p.quest("nq-combo-1"); q == nil || q.Progress[0] < 1 {
		t.Fatal("combo x5 quest")
	}
	p.CombatState = "ATTACKING"
	p.ComboChain = "LLL"
	p.Charging = true
	again := &Player{ID: p.ID, Name: p.Name, send: make(chan []byte, 8)}
	again.initCombat()
	ok, isNew := w.Add(again)
	if !ok || isNew {
		t.Fatal("reconnect")
	}
	if again.CombatState != "IDLE" || again.Charging || again.ComboChain != "" {
		t.Fatalf("safe reconnect state=%s charge=%v chain=%s", again.CombatState, again.Charging, again.ComboChain)
	}
}

func TestPhase25TrainingNoInfiniteExp(t *testing.T) {
	w, p := testVillagePlayer()
	var dummy *Enemy
	for _, e := range w.enemies {
		if e.Def.ID == "training_dummy" {
			dummy = e
			break
		}
	}
	if dummy == nil {
		t.Fatal("dummy")
	}
	p.X, p.Z = dummy.X, dummy.Z
	before := p.Exp
	dummy.HP = 1
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: dummy.ID, Timestamp: time.Now().UnixMilli()})
	w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if p.Exp != before {
		t.Fatalf("training dummy EXP %d -> %d", before, p.Exp)
	}
}

func TestPhase25EducationQ1(t *testing.T) {
	w, p := testVillagePlayer()
	n := npcByID["mbah_karya"]
	w.mu.Lock()
	nx, nz := w.npcLive(n)
	w.mu.Unlock()
	p.X, p.Z = nx, nz
	start := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_karya","kind":"quiz-combat"}`)})
	if !payloadHas(start, TypeInteractResult, "q1") && !payloadHas(start, TypeInteractResult, "2 + 3") {
		t.Fatalf("quiz %s", string(start[0]))
	}
	token := p.ensureLog().EduToken
	ok := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: []byte(`{"questionId":"q1","choice":1}`)})
	if rejectAction(ok, TypeEducationAnswer) {
		t.Fatal("correct B")
	}
	if p.ensureLog().EduToken <= token {
		t.Fatal("knowledge token")
	}
}

func TestPhase25BossPhaseStillWorks(t *testing.T) {
	if bossByID[WhispersBossID].Name != "PENJAGA GUA" {
		t.Fatal("boss catalog")
	}
	w := NewWorldState()
	p := dungeonReadyPlayer("p_a", "Raka")
	w.Add(p)
	inst := enterChapter1(t, w, p)
	if inst.BossPhase < 1 {
		t.Fatal("boss phase")
	}
}
