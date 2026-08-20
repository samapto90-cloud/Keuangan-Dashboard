package mmo

import (
	"crypto/rand"
	"fmt"
)

type CellCoord struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type MoveResult struct {
	Dice        int   `json:"dice"`
	WalkPath    []int `json:"walkPath"`
	WalkFinal   int   `json:"walkFinal"`
	SnakeFrom   int   `json:"snakeFrom,omitempty"`
	SnakeTo     int   `json:"snakeTo,omitempty"`
	LadderFrom  int   `json:"ladderFrom,omitempty"`
	LadderTo    int   `json:"ladderTo,omitempty"`
	Final       int   `json:"final"`
	Reached100  bool  `json:"reached100"`
	Bounced     bool  `json:"bounced"`
	Log         string `json:"log"`
}

func GetCellCoordinate(position int) (CellCoord, bool) {
	if position < MIN_POSITION || position > MAX_POSITION {
		return CellCoord{}, false
	}
	idx := position - 1
	row := idx / BOARD_GRID
	offset := idx % BOARD_GRID
	col := offset
	if row%2 == 1 {
		col = BOARD_GRID - 1 - offset
	}
	return CellCoord{Row: row, Col: col}, true
}

func GetPositionFromCoordinate(row, col int) (int, bool) {
	if row < 0 || row >= BOARD_GRID || col < 0 || col >= BOARD_GRID {
		return 0, false
	}
	offset := col
	if row%2 == 1 {
		offset = BOARD_GRID - 1 - col
	}
	return row*BOARD_GRID + offset + 1, true
}

func ValidateBoardConfig(cfg BoardConfig) error {
	if cfg.Size != BOARD_GRID || cfg.TotalCells != BOARD_SIZE {
		return fmt.Errorf("board size")
	}
	starts := map[int]string{}
	dests := map[int]string{}
	for from, to := range cfg.Snakes {
		if from <= to {
			return fmt.Errorf("snake %d -> %d", from, to)
		}
		if err := checkRange(from, to); err != nil {
			return err
		}
		if _, ok := starts[from]; ok {
			return fmt.Errorf("duplicate start %d", from)
		}
		if _, ok := dests[to]; ok {
			return fmt.Errorf("duplicate dest %d", to)
		}
		starts[from] = "snake"
		dests[to] = "snake"
	}
	for from, to := range cfg.Ladders {
		if to <= from {
			return fmt.Errorf("ladder %d -> %d", from, to)
		}
		if err := checkRange(from, to); err != nil {
			return err
		}
		if _, ok := starts[from]; ok {
			return fmt.Errorf("duplicate start %d", from)
		}
		if _, ok := dests[to]; ok {
			return fmt.Errorf("duplicate dest %d", to)
		}
		starts[from] = "ladder"
		dests[to] = "ladder"
	}
	for from, to := range cfg.Snakes {
		if starts[to] == "snake" {
			return fmt.Errorf("snake to snake %d -> %d", from, to)
		}
	}
	for from, to := range cfg.Ladders {
		if starts[to] == "ladder" {
			return fmt.Errorf("ladder to ladder %d -> %d", from, to)
		}
	}
	return nil
}

func checkRange(a, b int) error {
	if a < MIN_POSITION || a > MAX_POSITION || b < MIN_POSITION || b > MAX_POSITION {
		return fmt.Errorf("out of range")
	}
	return nil
}

func WalkPath(from, dice int) ([]int, bool) {
	if dice < 1 || dice > 6 || from < OFFBOARD_START || from > MAX_POSITION {
		return nil, false
	}
	path := make([]int, 0, dice+4)
	pos := from
	over := from + dice - MAX_POSITION
	if over > 0 {
		up := MAX_POSITION - from
		for i := 1; i <= up; i++ {
			pos++
			path = append(path, pos)
		}
		for i := 0; i < over; i++ {
			pos--
			path = append(path, pos)
		}
		return path, true
	}
	for i := 0; i < dice; i++ {
		pos++
		path = append(path, pos)
	}
	return path, false
}

func ResolveMove(cfg BoardConfig, from, dice int) MoveResult {
	path, bounced := WalkPath(from, dice)
	walkFinal := from
	if len(path) > 0 {
		walkFinal = path[len(path)-1]
	}
	final := walkFinal
	out := MoveResult{
		Dice: dice, WalkPath: path, WalkFinal: walkFinal, Final: final, Bounced: bounced,
	}
	if to, ok := cfg.Snakes[walkFinal]; ok {
		out.SnakeFrom = walkFinal
		out.SnakeTo = to
		final = to
	} else if to, ok := cfg.Ladders[walkFinal]; ok {
		out.LadderFrom = walkFinal
		out.LadderTo = to
		final = to
	}
	out.Final = final
	out.Reached100 = final == MAX_POSITION
	out.Log = fmt.Sprintf("dice=%d before=%d move=%d", dice, from, walkFinal)
	if out.SnakeTo != 0 {
		out.Log += fmt.Sprintf(" snake=%d final=%d", out.SnakeTo, final)
	} else if out.LadderTo != 0 {
		out.Log += fmt.Sprintf(" ladder=%d final=%d", out.LadderTo, final)
	} else {
		out.Log += fmt.Sprintf(" final=%d", final)
	}
	return out
}

func RollDiceSecure() (int, error) {
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return int(b[0]%6) + 1, nil
}

func NextTurnIndex(current, playerCount int) int {
	if playerCount <= 0 {
		return 0
	}
	return (current + 1) % playerCount
}

func TokenOffsets(count, index int) (dx, dy float64) {
	if count <= 1 {
		return 0, 0
	}
	switch count {
	case 2:
		if index == 0 {
			return -0.18, 0
		}
		return 0.18, 0
	case 3:
		pts := [][2]float64{{0, -0.16}, {-0.16, 0.12}, {0.16, 0.12}}
		p := pts[index%3]
		return p[0], p[1]
	default:
		pts := [][2]float64{{-0.18, -0.18}, {0.18, -0.18}, {-0.18, 0.18}, {0.18, 0.18}}
		p := pts[index%4]
		return p[0], p[1]
	}
}
