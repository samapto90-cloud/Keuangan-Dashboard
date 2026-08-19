package mmo

const (
	BOARD_SIZE   = 100
	MIN_POSITION = 1
	MAX_POSITION = 100
	MAX_PLAYERS  = 4
)

// Default snakes: head → tail. Animation is phase 2.
var DefaultSnakes = map[int]int{
	97: 78,
	92: 71,
	64: 36,
}

// Default ladders: bottom → top. Animation is phase 2.
var DefaultLadders = map[int]int{
	4:  25,
	14: 45,
	28: 56,
}
