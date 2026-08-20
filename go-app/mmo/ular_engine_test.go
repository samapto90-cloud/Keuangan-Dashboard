package mmo

import "testing"

func TestGetCellCoordinate(t *testing.T) {
	c, ok := GetCellCoordinate(1)
	if !ok || c.Row != 0 || c.Col != 0 {
		t.Fatalf("1 %+v", c)
	}
	c, ok = GetCellCoordinate(10)
	if !ok || c.Row != 0 || c.Col != 9 {
		t.Fatalf("10 %+v", c)
	}
	c, ok = GetCellCoordinate(11)
	if !ok || c.Row != 1 || c.Col != 9 {
		t.Fatalf("11 %+v", c)
	}
	c, ok = GetCellCoordinate(100)
	if !ok || c.Row != 9 || c.Col != 0 {
		t.Fatalf("100 %+v", c)
	}
	p, ok := GetPositionFromCoordinate(0, 0)
	if !ok || p != 1 {
		t.Fatal("coord 0,0")
	}
	p, ok = GetPositionFromCoordinate(9, 0)
	if !ok || p != 100 {
		t.Fatal("coord 9,0")
	}
	p, ok = GetPositionFromCoordinate(1, 9)
	if !ok || p != 11 {
		t.Fatal("coord 1,9")
	}
}

func TestOffboardEntry(t *testing.T) {
	cfg := DefaultBoardConfig()
	r := ResolveMove(cfg, OFFBOARD_START, 4)
	if r.WalkFinal != 4 || len(r.WalkPath) != 4 {
		t.Fatalf("0+4 walk=%d path=%v", r.WalkFinal, r.WalkPath)
	}
	if r.WalkPath[0] != 1 || r.WalkPath[3] != 4 {
		t.Fatalf("entry path %v", r.WalkPath)
	}
	r = ResolveMove(cfg, OFFBOARD_START, 1)
	if r.WalkFinal != 1 || r.Final != 1 {
		t.Fatalf("0+1 %d", r.Final)
	}
}

func TestMovementAndBounce(t *testing.T) {
	cfg := DefaultBoardConfig()
	if ResolveMove(cfg, 1, 1).WalkFinal != 2 {
		t.Fatal("1+1")
	}
	if ResolveMove(cfg, 10, 6).WalkFinal != 16 {
		t.Fatal("10+6")
	}
	if ResolveMove(cfg, 50, 6).WalkFinal != 56 {
		t.Fatal("50+6")
	}
	r := ResolveMove(cfg, 98, 5)
	if r.WalkFinal != 97 || len(r.WalkPath) != 5 {
		t.Fatalf("98+5 walk=%d path=%v", r.WalkFinal, r.WalkPath)
	}
	r = ResolveMove(cfg, 99, 2)
	if r.WalkFinal != 99 {
		t.Fatalf("99+2 %d %v", r.WalkFinal, r.WalkPath)
	}
	r = ResolveMove(cfg, 97, 3)
	if r.WalkFinal != 100 || !r.Reached100 {
		t.Fatalf("97+3 %d", r.WalkFinal)
	}
}

func TestSnakeLadder(t *testing.T) {
	cfg := DefaultBoardConfig()
	if err := ValidateBoardConfig(cfg); err != nil {
		t.Fatal(err)
	}
	r := ResolveMove(cfg, 87, 1)
	if r.WalkFinal != 88 || r.SnakeTo != 48 || r.Final != 48 {
		t.Fatalf("snake 88 %+v", r)
	}
	r = ResolveMove(cfg, 27, 1)
	if r.WalkFinal != 28 || r.LadderTo != 55 || r.Final != 55 {
		t.Fatalf("ladder 28 %+v", r)
	}
}

func TestTurns(t *testing.T) {
	if NextTurnIndex(0, 4) != 1 || NextTurnIndex(3, 4) != 0 {
		t.Fatal("4p")
	}
	if NextTurnIndex(0, 2) != 1 || NextTurnIndex(1, 2) != 0 {
		t.Fatal("2p")
	}
}

func TestTokenOffsetsNoOverlap(t *testing.T) {
	seen := map[[2]int]bool{}
	for i := 0; i < 4; i++ {
		dx, dy := TokenOffsets(4, i)
		key := [2]int{int(dx * 1000), int(dy * 1000)}
		if seen[key] {
			t.Fatal("overlap")
		}
		seen[key] = true
	}
}

func TestFinish(t *testing.T) {
	cfg := DefaultBoardConfig()
	r := ResolveMove(cfg, 99, 1)
	if !r.Reached100 || r.Final != 100 {
		t.Fatal("finish")
	}
}

func TestDiceRange(t *testing.T) {
	for i := 0; i < 40; i++ {
		n, err := RollDiceSecure()
		if err != nil || n < 1 || n > 6 {
			t.Fatal(n, err)
		}
	}
}
