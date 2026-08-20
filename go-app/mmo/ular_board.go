package mmo

const (
	BOARD_GRID      = 10
	BOARD_SIZE      = 100
	OFFBOARD_START  = 0
	MIN_POSITION    = 1
	MAX_POSITION    = 100
	MAX_PLAYERS     = 4
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

// Default snakes: head → tail.
var DefaultSnakes = map[int]int{
	97: 78,
	95: 75,
	92: 71,
	88: 48,
	87: 36,
	62: 18,
	54: 34,
	17: 7,
}

// Default ladders: bottom → top.
var DefaultLadders = map[int]int{
	4:  25,
	9:  31,
	20: 42,
	28: 55,
	40: 59,
	51: 73,
	63: 81,
	70: 93,
}

var PlayerTokenColors = [4]string{"#e23d3d", "#3d7dff", "#1f8a64", "#e6c84f"}
