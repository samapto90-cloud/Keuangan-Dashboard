package mmo

import (
	"math"
	"testing"
)

func TestWorldColliderResolve(t *testing.T) {
	x, z := resolveWorldXZ(-9, 8, NPCRadius)
	if insideAABB(x, z, NPCRadius, villageObstacles[0]) {
		t.Fatalf("npc still inside house at %.2f, %.2f", x, z)
	}
	if !isWalkableXZ(0, 5, NPCRadius) {
		t.Fatal("plaza should be walkable")
	}
	if isWalkableXZ(-12, 14.5, NPCRadius) && !onBridge(-12, 14.5) {
		t.Fatal("water should block unless on bridge")
	}
	if !onBridge(-12, 14.5) {
		t.Fatal("bridge center")
	}
}

func TestNPCDestinationSlots(t *testing.T) {
	a := npcByID["lio"]
	b := npcByID["dawn_merchant"]
	if a.ID == "" || b.ID == "" {
		t.Fatal("need merchants")
	}
	ax, az := npcDestWithSlot(a, "market", WorldTimeSystem{ClockMin: 10 * 60})
	bx, bz := npcDestWithSlot(b, "market", WorldTimeSystem{ClockMin: 10 * 60})
	if math.Hypot(ax-bx, az-bz) < 0.4 {
		t.Fatalf("merchants should not share same slot %.2f,%.2f vs %.2f,%.2f", ax, az, bx, bz)
	}
}

func TestNPCSeparationNoOverlap(t *testing.T) {
	w := NewWorldState()
	w.Time.ClockMin = 10 * 60
	for i := 0; i < 240; i++ {
		w.tickNPCs(0.1)
	}
	minD := 999.0
	ids := make([]string, 0, len(w.npcPos))
	for id := range w.npcPos {
		ids = append(ids, id)
	}
	for i := 0; i < len(ids); i++ {
		a := w.npcPos[ids[i]]
		for j := i + 1; j < len(ids); j++ {
			b := w.npcPos[ids[j]]
			d := math.Hypot(a.X-b.X, a.Z-b.Z)
			if d < minD {
				minD = d
			}
			if d < 0.42 {
				t.Fatalf("npc overlap %s vs %s dist %.2f", ids[i], ids[j], d)
			}
		}
	}
	if minD < 0.42 {
		t.Fatalf("min dist %.2f", minD)
	}
}

func TestNPCStuckRecovery(t *testing.T) {
	w := NewWorldState()
	n := npcByID["child_lina"]
	if n.ID == "" {
		t.Fatal("child_lina missing")
	}
	cur := npcLive{X: -9.05, Z: 8.05, Yaw: 0, LastX: -9.05, LastZ: 8.05}
	w.npcPos = map[string]npcLive{n.ID: cur}
	for i := 0; i < 30; i++ {
		w.tickNPCs(0.1)
	}
	live := w.npcPos[n.ID]
	if live.StuckT > 0 && !live.UsingWander {
		t.Fatalf("expected wander or unstuck, stuck %.2f", live.StuckT)
	}
}

func TestPlayerBuildingCollision(t *testing.T) {
	x, z := resolveWorldXZ(-9, 8, PlayerRadius)
	if insideAABB(x, z, PlayerRadius, villageObstacles[0]) {
		t.Fatalf("player inside house %.2f, %.2f", x, z)
	}
}

func TestPickWanderWalkable(t *testing.T) {
	x, z := pickWanderPoint(0, 6, 12, "test_wander")
	if !isWalkableXZ(x, z, NPCRadius) {
		t.Fatalf("wander point not walkable %.2f, %.2f", x, z)
	}
}
