package mmo

// UlarMatchStatus is the phase-1 match/room state machine.
// Gameplay (dice, movement, questions) is not implemented yet.
type UlarMatchStatus string

const (
	UlarWaiting    UlarMatchStatus = "WAITING"
	UlarReady      UlarMatchStatus = "READY"
	UlarStarting   UlarMatchStatus = "STARTING"
	UlarPlayerTurn UlarMatchStatus = "PLAYER_TURN"
	UlarRolling    UlarMatchStatus = "ROLLING"
	UlarMoving     UlarMatchStatus = "MOVING"
	UlarSnake      UlarMatchStatus = "SNAKE"
	UlarLadder     UlarMatchStatus = "LADDER"
	UlarOnQuestion UlarMatchStatus = "QUESTION"
	UlarAnswering  UlarMatchStatus = "ANSWERING"
	UlarPenalty    UlarMatchStatus = "PENALTY"
	UlarNextTurn   UlarMatchStatus = "NEXT_TURN"
	UlarFinished   UlarMatchStatus = "FINISHED"
)

// AdventureGameplayEnabled keeps the old MMORPG simulation off.
const AdventureGameplayEnabled = false
