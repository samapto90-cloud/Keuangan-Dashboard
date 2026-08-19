package mmo

import (
	"math"
	"testing"
	"time"
)

func TestSimulateWalkStaysInWorld(t *testing.T) {
	p := &Player{Grounded: true, CamYaw: 0, AZ: 1, LastHeard: time.Now()}
	for i := 0; i < 40; i++ {
		p.Simulate(0.05)
	}
	if math.Abs(p.X) > WorldLimit || math.Abs(p.Z) > WorldLimit {
		t.Fatalf("keluar peta: %v %v", p.X, p.Z)
	}
	if p.State != "WALK" && p.State != "RUN" && p.State != "IDLE" {
		t.Fatalf("state aneh: %s", p.State)
	}
}

func TestInputRateLimit(t *testing.T) {
	p := &Player{}
	p.SetInput(MoveInput{AZ: 1, Seq: 1})
	p.SetInput(MoveInput{AZ: 0, Seq: 2})
	if p.AZ != 1 {
		t.Fatal("input kedua dalam interval harus di-drop")
	}
}

func TestStrafeRightMovesScreenRight(t *testing.T) {
	p := &Player{Grounded: true, CamYaw: 0, AX: 1, LastHeard: time.Now()}
	p.Simulate(0.05)
	if p.VX >= 0 {
		t.Fatalf("D/kanan harus ke kanan layar (-X), vx=%v", p.VX)
	}
	fwd := &Player{Grounded: true, CamYaw: 0, AZ: 1, LastHeard: time.Now()}
	fwd.Simulate(0.05)
	if fwd.VZ <= 0 {
		t.Fatalf("W harus maju ke +Z, vz=%v", fwd.VZ)
	}
}

func TestSanitizeName(t *testing.T) {
	if sanitizeName("  ") != "Raka" {
		t.Fatal("nama kosong harus Raka")
	}
}

func TestWorldJoinSnapshot(t *testing.T) {
	w := NewWorldState()
	a := &Player{ID: "p_a", Name: "Raka", Class: "WARRIOR", Level: 1, HP: 100, MaxHP: 100, send: make(chan []byte, 1)}
	ok, isNew := w.Add(a)
	if !ok || !isNew {
		t.Fatal("player pertama harus spawn baru")
	}
	snap := w.Snapshot()
	if snap.Online != 1 || snap.WorldID != WorldID || snap.Channel != ChannelID {
		t.Fatalf("snapshot salah: %+v", snap)
	}
	b := &Player{ID: "p_a", Name: "Raka", send: make(chan []byte, 1)}
	ok, isNew = w.Add(b)
	if !ok || isNew {
		t.Fatal("reconnect harus mengganti sesi, bukan spawn baru")
	}
	if w.Remove(a) != nil {
		t.Fatal("sesi lama tidak boleh menghapus sesi baru")
	}
	if w.Remove(b) == nil {
		t.Fatal("sesi aktif harus bisa dihapus")
	}
}
