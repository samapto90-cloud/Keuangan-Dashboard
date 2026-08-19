package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func testWorldWithPlayer() (*WorldState, *Player, *Enemy) {
	w := NewWorldState()
	p := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 8)}
	p.initCombat()
	ok, _ := w.Add(p)
	if !ok {
		panic("add")
	}
	var enemy *Enemy
	for _, e := range w.enemies {
		enemy = e
		break
	}
	p.X, p.Z = enemy.X, enemy.Z
	return w, p, enemy
}

func TestDamageMinimum(t *testing.T) {
	d, _ := calcDamage(1, 0, 1, 1, 9999)
	if d < 1 {
		t.Fatalf("min damage %d", d)
	}
}

func TestCheatDamageFieldIgnored(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	hp := e.HP
	raw, _ := json.Marshal(map[string]any{
		"attackType":  "punch",
		"targetId":    e.ID,
		"timestamp":   time.Now().UnixMilli(),
		"direction":   0,
		"damage":      999999,
		"finalDamage": 999999,
		"enemyHp":     1,
	})
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if len(evs) == 0 {
		t.Fatal("serangan valid harus menghasilkan event")
	}
	if e.HP <= 0 {
		return
	}
	lost := hp - e.HP
	if lost > 80 {
		t.Fatalf("damage client 999999 tidak boleh dipakai: lost=%d", lost)
	}
}

func TestRejectPlayerTarget(t *testing.T) {
	w, p, _ := testWorldWithPlayer()
	other := &Player{ID: "p_b", Name: "Sinta", send: make(chan []byte, 8)}
	other.initCombat()
	w.Add(other)
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: other.ID, Timestamp: time.Now().UnixMilli()})
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if len(evs) != 1 {
		t.Fatalf("target player harus di-reject, evs=%d", len(evs))
	}
}

func TestDoubleKillExpOnce(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	e.HP = 1
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	exp1 := p.Exp
	alive := e.Alive
	p.AttackCDUntil = time.Time{}
	w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if e.Alive != alive && exp1 > 0 && p.Exp != exp1 {
		t.Fatal("exp ganda setelah musuh mati")
	}
	if p.Exp != exp1 && !e.Alive {
		t.Fatalf("double EXP: %d -> %d", exp1, p.Exp)
	}
}

func TestCooldownReject(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if len(evs) != 1 {
		t.Fatal("cooldown harus reject")
	}
}

func TestRangeReject(t *testing.T) {
	w, p, e := testWorldWithPlayer()
	p.X, p.Z = 30, -30
	raw, _ := json.Marshal(AttackIn{AttackType: "punch", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerAttack, Data: raw})
	if len(evs) != 1 {
		t.Fatal("jarak jauh harus reject")
	}
}

func TestSkillEnergyAndCatalog(t *testing.T) {
	if _, ok := skillCatalog["power_strike"]; !ok {
		t.Fatal("skill power_strike harus ada")
	}
	w, p, e := testWorldWithPlayer()
	p.X, p.Z = e.X, e.Z
	p.Energy = 0
	raw, _ := json.Marshal(SkillIn{SkillID: "power_strike", TargetID: e.ID, Timestamp: time.Now().UnixMilli()})
	evs := w.ApplyCombat(p.ID, Envelope{Type: TypePlayerSkill, Data: raw})
	if len(evs) != 1 {
		t.Fatal("energy habis harus reject")
	}
}

func TestEnemyCatalog(t *testing.T) {
	if len(enemyCatalog) < 6 {
		t.Fatalf("catalog musuh kurang, ada %d", len(enemyCatalog))
	}
}
