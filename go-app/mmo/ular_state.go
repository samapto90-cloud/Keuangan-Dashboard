package mmo

type UlarMatchStatus string

const (
	UlarWaiting  UlarMatchStatus = "WAITING"
	UlarReady    UlarMatchStatus = "READY"
	UlarStarting UlarMatchStatus = "STARTING"
	UlarPlaying  UlarMatchStatus = "PLAYING"
	UlarClosed   UlarMatchStatus = "CLOSED"

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
	UlarTerminated UlarMatchStatus = "ADMIN_TERMINATED"
)

const (
	DisconnectGrace  = 60
	AnimCompleteWait = 12
	AFKWarnSeconds   = 45
	RoomCodeLen      = 6
	ChatMaxLen       = 200
)

const AdventureGameplayEnabled = false

const (
	ErrNotYourTurn       = "NOT_YOUR_TURN"
	ErrGameNotStarted    = "GAME_NOT_STARTED"
	ErrRoomFull          = "ROOM_FULL"
	ErrNotReady          = "NOT_READY"
	ErrAlreadyRolled     = "ALREADY_ROLLED"
	ErrNotInRoom         = "NOT_IN_ROOM"
	ErrRoomClosed        = "ROOM_CLOSED"
	ErrNeedPlayers       = "NEED_PLAYERS"
	ErrNotHost           = "NOT_HOST"
	ErrInvalidRequest    = "INVALID_REQUEST"
	ErrRateLimited       = "RATE_LIMITED"
	ErrLocked            = "GAME_LOCKED"
	ErrNoQuestion        = "NO_QUESTION"
	ErrNotQuestionPlayer = "NOT_QUESTION_PLAYER"
	ErrAlreadyAnswered   = "ALREADY_ANSWERED"
	ErrLateAnswer        = "LATE_ANSWER"
)
