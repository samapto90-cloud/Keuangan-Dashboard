package mmo

import (
	"math"
	"testing"
	"time"
)

func TestPickupPartialStackDoesNotDuplicate(t *testing.T) {
	w, p := testVillagePlayer()
	for i := range p.Bag.Slots {
		if p.Bag.Slots[i].ItemID == "" {
			p.Bag.Slots[i] = InvSlot{ItemID: "potion_heal", Qty: 20}
		} else if p.Bag.Slots[i].ItemID == "potion_heal" {
			p.Bag.Slots[i].Qty = 20
		}
	}
	last := -1
	for i := range p.Bag.Slots {
		if p.Bag.Slots[i].ItemID == "potion_heal" {
			last = i
			break
		}
	}
	if last < 0 {
		t.Fatal("need potion slot")
	}
	p.Bag.Slots[last].Qty = 19
	before := p.Bag.count("potion_heal")
	w.drops["drop-partial"] = &WorldDrop{
		ID: "drop-partial", ItemID: "potion_heal", Qty: 2, OwnerID: p.ID, InstanceID: p.InstanceID,
		X: p.X, Z: p.Z, Until: time.Now().Add(time.Minute),
	}
	evs := w.ApplyInventory(p.ID, Envelope{Type: TypePickupItem, Data: []byte(`{"dropId":"drop-partial"}`)})
	if rejectAction(evs, TypePickupItem) {
		t.Fatal("partial pickup should stash remainder")
	}
	got := p.Bag.count("potion_heal")
	stash := 0
	for _, row := range p.ensureLog().TempLoot {
		if row.ItemID == "potion_heal" {
			stash += row.Qty
		}
	}
	if got != before+1 {
		t.Fatalf("bag %d want %d", got, before+1)
	}
	if stash != 1 {
		t.Fatalf("stash remainder want 1 got %d (total would dup as %d)", stash, got+stash-before)
	}
}

func TestDeadReconnectDoesNotRefillHP(t *testing.T) {
	p := &Player{CombatState: "DEAD", HP: 0, MaxHP: 80}
	p.initCombat()
	if p.HP != 0 {
		t.Fatalf("dead reconnect filled HP to %d", p.HP)
	}
}

func TestHitStunExpiresToIdle(t *testing.T) {
	w, p := testVillagePlayer()
	p.CombatState = "HIT"
	p.HitUntil = time.Now().Add(-time.Millisecond)
	w.tickCombat(0.05)
	if p.CombatState != "IDLE" {
		t.Fatalf("hit stun should clear, got %s", p.CombatState)
	}
}

func TestGhostHousesRemovedMiraReachable(t *testing.T) {
	if !isWalkableXZ(-9.2, 2.2, PlayerRadius) {
		t.Fatal("Mira harus bisa didekati")
	}
	if !isWalkableXZ(9.4, 3.4, PlayerRadius) {
		t.Fatal("Lio harus bisa didekati")
	}
	x, z := resolveWorldXZ(-9.2, 2.2, PlayerRadius)
	if math.Hypot(x+9.2, z-2.2) > 0.6 {
		t.Fatalf("player terdorong jauh dari Mira: %.2f %.2f", x, z)
	}
}

func TestNPCSlotsWalkable(t *testing.T) {
	mira := npcByID["mira"]
	if mira.ID == "" {
		t.Fatal("mira")
	}
	x, z := npcDestWithSlot(mira, "school", WorldTimeSystem{ClockMin: 10 * 60})
	if !isWalkableXZ(x, z, NPCRadius) {
		t.Fatalf("school slot di dalam bangunan: %.2f %.2f", x, z)
	}
}

func TestThirtyThreeGuardiansSeeded(t *testing.T) {
	w := NewWorldState()
	n := 0
	seen := map[string]bool{}
	for _, e := range w.enemies {
		if _, ok := guardianByID[e.Def.ID]; ok && !seen[e.Def.ID] {
			seen[e.Def.ID] = true
			n++
		}
	}
	if n != 33 {
		t.Fatalf("guardian seeded %d want 33", n)
	}
}

func TestFinishCinematicDoesNotMarkSkip(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().PendingCinematic = "cin-opening"
	p.ensureLog().CinematicsSeen = map[string]bool{}
	w.finishCinematic(p)
	if p.ensureLog().Flags["cinematic_skipped"] {
		t.Fatal("selesai natural tidak boleh cinematic_skipped")
	}
	if !p.ensureLog().CinematicsSeen["cin-opening"] {
		t.Fatal("cinematic harus marked seen")
	}
}

func TestNaNHeightRecovers(t *testing.T) {
	_, p := testVillagePlayer()
	p.Y = math.NaN()
	p.VY = math.NaN()
	p.Simulate(0.05)
	if math.IsNaN(p.Y) || p.Y != 0 {
		t.Fatalf("Y %v", p.Y)
	}
}
