package mmo

import "time"

type UlarPlayer struct {
	ID             string         `json:"id"`
	UserID         string         `json:"userId"`
	Username       string         `json:"username"`
	Avatar         string         `json:"avatar"`
	Color          string         `json:"color"`
	Position       int            `json:"position"`
	Items          map[string]int `json:"items,omitempty"`
	IsReady        bool           `json:"isReady"`
	IsConnected    bool           `json:"isConnected"`
	ConnState      string         `json:"connState"`
	PlayState      string         `json:"playState"`
	DisconnectedAt time.Time      `json:"-"`
	JoinedAt       time.Time      `json:"-"`
	Abandoned      bool           `json:"-"`
}

func NewUlarPlayer(userID, username string, slot int) *UlarPlayer {
	return &UlarPlayer{
		ID:          userID,
		UserID:      userID,
		Username:    username,
		Avatar:      "default",
		Color:       PlayerTokenColors[slot%len(PlayerTokenColors)],
		Position:    OFFBOARD_START,
		Items:       map[string]int{},
		IsReady:     false,
		IsConnected: true,
		ConnState:   "CONNECTED",
		PlayState:   "NOT_READY",
		JoinedAt:    time.Now().UTC(),
	}
}

type UlarChatLine struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Text     string `json:"text"`
	Emote    string `json:"emote,omitempty"`
	At       int64  `json:"at"`
}

type UlarRoom struct {
	ID         string          `json:"id"`
	RoomCode   string          `json:"roomCode"`
	HostID     string          `json:"hostId"`
	Status     UlarMatchStatus `json:"status"`
	MaxPlayers int             `json:"maxPlayers"`
	CreatedAt  time.Time       `json:"createdAt"`
	Visibility string          `json:"visibility"`
	Mode       string          `json:"mode,omitempty"`
	Grade      string          `json:"grade,omitempty"`
	Players    []*UlarPlayer   `json:"players"`
	Seq        int             `json:"seq"`
	Match      *UlarLiveMatch  `json:"match,omitempty"`
	Chat       []UlarChatLine  `json:"chat,omitempty"`
	SeenEvent  map[string]bool `json:"-"`
}

type UlarLiveMatch struct {
	ID                    string            `json:"id"`
	RoomID                string            `json:"roomId"`
	Status                UlarMatchStatus   `json:"status"`
	Phase                 UlarMatchStatus   `json:"phase"`
	CurrentPlayerID       string            `json:"currentPlayerId"`
	TurnNumber            int               `json:"turnNumber"`
	WinnerID              string            `json:"winnerId,omitempty"`
	LastDice              int               `json:"lastDice"`
	LastAction            string            `json:"lastAction"`
	CountdownEnd          int64             `json:"countdownEnd,omitempty"`
	CreatedAt             time.Time         `json:"createdAt"`
	StartedAt             time.Time         `json:"startedAt,omitempty"`
	FinishedAt            *time.Time        `json:"finishedAt,omitempty"`
	LastActionAt          time.Time         `json:"-"`
	WaitingAnim           bool              `json:"-"`
	PendingMove           MoveResult        `json:"-"`
	CurrentQuestionID     string            `json:"currentQuestionId,omitempty"`
	QuestionPlayerID      string            `json:"questionPlayerId,omitempty"`
	QuestionStartedAt     time.Time         `json:"questionStartedAt,omitempty"`
	QuestionState         string            `json:"questionState,omitempty"`
	AnswerSubmitted       bool              `json:"answerSubmitted,omitempty"`
	Penalty               int               `json:"penalty,omitempty"`
	PositionBeforePenalty int               `json:"positionBeforePenalty,omitempty"`
	QuestionFinal         bool              `json:"questionFinal,omitempty"`
	QuestionNumber        int               `json:"questionNumber,omitempty"`
	UsedQuestionIDs       []string          `json:"-"`
	SubjectCursor         int               `json:"-"`
	QuestionView          QuestionPublic    `json:"-"`
	Attempts              []QuestionAttempt `json:"-"`
	SnakeHits             map[string]int    `json:"-"`
	LadderHits            map[string]int    `json:"-"`
	QuestionLimit         time.Duration     `json:"-"`
	PenaltyN              int               `json:"-"`
	// Frozen reward config for match integrity. Must not depend on LiveConfig()
	// after the match has started (admin config changes should affect only future matches).
	XPCorrectAmt          int               `json:"-"`
	XPWrongAmt            int               `json:"-"`
	XPTimeoutAmt          int              `json:"-"`
	XPMatchCompleteAmt   int              `json:"-"`
	XPWinAmt              int              `json:"-"`
	CoinMatchAmt          int              `json:"-"`
	CoinWinAmt            int              `json:"-"`
	RankWinRrAmt          int              `json:"-"`
	RankLossRrAmt         int              `json:"-"`
	AchievementsByID      map[string]UlarAchievement `json:"-"`
	Terminated            bool              `json:"-"`
	TerminateReason       string            `json:"-"`
	PowerCells            map[int]string    `json:"powerCells,omitempty"`
	Grade                 string            `json:"grade,omitempty"`
}

type UlarMatch struct {
	ID              string          `json:"id"`
	RoomID          string          `json:"roomId"`
	Status          UlarMatchStatus `json:"status"`
	CurrentPlayerID string          `json:"currentPlayerId"`
	WinnerID        string          `json:"winnerId,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	FinishedAt      *time.Time      `json:"finishedAt,omitempty"`
}

type UlarQuestion struct {
	ID            string `json:"id"`
	Category      string `json:"category"`
	Subject       string `json:"subject"`
	Grade         string `json:"grade"`
	Difficulty    string `json:"difficulty"`
	Question      string `json:"question"`
	OptionA       string `json:"optionA"`
	OptionB       string `json:"optionB"`
	OptionC       string `json:"optionC"`
	OptionD       string `json:"optionD"`
	CorrectAnswer string `json:"correctAnswer"`
	Explanation   string `json:"explanation"`
	Active        bool   `json:"active"`
}

type StoredMatchPlayer struct {
	UserID        string `json:"userId"`
	Username      string `json:"username"`
	Color         string `json:"playerColor"`
	FinalPosition int    `json:"finalPosition"`
	FinishOrder   int    `json:"finishOrder"`
}

type StoredMatch struct {
	ID         string              `json:"id"`
	RoomID     string              `json:"roomId"`
	RoomCode   string              `json:"roomCode"`
	Status     string              `json:"status"`
	Mode       string              `json:"mode,omitempty"`
	WinnerID   string              `json:"winnerId,omitempty"`
	StartedAt  int64               `json:"startedAt"`
	FinishedAt int64               `json:"finishedAt"`
	Players    []StoredMatchPlayer `json:"players"`
}
