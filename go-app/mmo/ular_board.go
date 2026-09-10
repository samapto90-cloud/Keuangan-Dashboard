package mmo

const (
	BOARD_GRID      = 10
	BOARD_SIZE      = 100
	OFFBOARD_START  = 0
	MIN_POSITION    = 1
	MAX_POSITION    = 100
	MAX_PLAYERS     = 8
	MOVE_DURATION   = 200
	SNAKE_DURATION  = 1000
	LADDER_DURATION = 1000
	DICE_DURATION   = 1000
)

type BoardConfig struct {
	Size       int         `json:"size"`
	TotalCells int         `json:"totalCells"`
	Snakes     map[int]int `json:"snakes"`
	Ladders    map[int]int `json:"ladders"`
}

func DefaultBoardConfig() BoardConfig {
	return BoardConfig{
		Size:       BOARD_GRID,
		TotalCells: BOARD_SIZE,
		Snakes:     cloneIntMap(DefaultSnakes),
		Ladders:    cloneIntMap(DefaultLadders),
	}
}

func cloneIntMap(src map[int]int) map[int]int {
	out := make(map[int]int, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// Default snakes: head → tail (visual snakes off; gameplay only).
var DefaultSnakes = map[int]int{
	97: 75,
	93: 55,
	87: 52,
	66: 34,
	57: 24,
	43: 18,
	35: 12,
}

// Default ladders: dimatikan (kosong).
var DefaultLadders = map[int]int{}

var PlayerTokenColors = [4]string{"#e23d3d", "#3d7dff", "#1f8a64", "#e6c84f"}
