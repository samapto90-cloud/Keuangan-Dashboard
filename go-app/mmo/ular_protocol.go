package mmo

const (
	TypeJoinLobby     = "JOIN_LOBBY"
	TypeLobbyHello    = "LOBBY_HELLO"
	TypeRoomCreate    = "ROOM_CREATE"
	TypeRoomJoin      = "ROOM_JOIN"
	TypeRoomLeave     = "ROOM_LEAVE"
	TypeRoomList      = "ROOM_LIST"
	TypeRoomUpdated   = "ROOM_UPDATED"
	TypeRoomStart     = "ROOM_START"
	TypeRoomClose     = "ROOM_CLOSE"
	TypePlayerReady   = "PLAYER_READY"
	TypeUlarGameInfo  = "ULAR_GAME_INFO"
	TypeGameState     = "GAME_STATE"
	TypeGameTurn      = "TURN_CHANGED"
	TypeGameRoll      = "GAME_ROLL_REQUEST"
	TypeDiceResult    = "DICE_RESULT"
	TypePlayerMoving  = "PLAYER_MOVING"
	TypeSnakeTrigger  = "SNAKE_TRIGGERED"
	TypeLadderTrigger = "LADDER_TRIGGERED"
	TypeGameFinish    = "GAME_FINISHED"
	TypeAnimDone      = "MOVEMENT_ANIMATION_COMPLETE"
	TypePlayerJoinEv  = "PLAYER_JOIN"
	TypePlayerLeaveEv = "PLAYER_LEAVE"
	TypePlayerDisc    = "PLAYER_DISCONNECT"
	TypePlayerRecon   = "PLAYER_RECONNECT"
	TypeRoomChat      = "ROOM_CHAT"
	TypeRoomEmote     = "ROOM_EMOTE"
	TypeUlarError     = "ULAR_ERROR"
	TypeMatchFound    = "MATCH_FOUND"
	TypeMatchReady    = "MATCH_READY"
	TypeRoomKick      = "ROOM_KICK"
	TypeGameInviteEv  = "GAME_INVITE"
	TypeInviteRespond = "INVITE_RESPOND"
	TypeSocialPush    = "SOCIAL_NOTIFY"
	TypeOnlineList    = "ONLINE_LIST"
	TypeOnlineListReq = "ONLINE_LIST_REQ"
	TypeJoinAsk       = "JOIN_ASK"
	TypeJoinAskEv     = "JOIN_ASK_EV"
	TypeJoinAskRespond = "JOIN_ASK_RESPOND"
	TypeInviteResult  = "INVITE_RESULT"
)

type LobbyHelloOut struct {
	Title      string `json:"title"`
	Version    string `json:"version"`
	Phase      string `json:"phase"`
	PlayerID   string `json:"playerId"`
	Username   string `json:"username"`
	BoardSize  int    `json:"boardSize"`
	MaxPlayers int    `json:"maxPlayers"`
}

type RoomCodeIn struct {
	RoomCode string `json:"roomCode"`
	UserID   string `json:"userId"`
}

type PlayerReadyIn struct {
	Ready  bool   `json:"ready"`
	UserID string `json:"userId"`
}

type RoomListOut struct {
	Rooms []RoomPublic `json:"rooms"`
}

type RoomPublic struct {
	RoomCode string `json:"roomCode"`
	Players  int    `json:"players"`
	Max      int    `json:"maxPlayers"`
	Status   string `json:"status"`
}

type UlarEnvelope struct {
	Seq     int    `json:"seq"`
	EventID string `json:"eventId"`
	At      int64  `json:"at"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type GameSnapshot struct {
	UlarEnvelope
	RoomCode          string          `json:"roomCode"`
	RoomIDHidden      bool            `json:"-"`
	HostID            string          `json:"hostId"`
	Status            UlarMatchStatus `json:"status"`
	MatchID           string          `json:"matchId,omitempty"`
	Phase             UlarMatchStatus `json:"phase,omitempty"`
	CurrentPlayerID   string          `json:"currentPlayerId,omitempty"`
	TurnNumber        int             `json:"turnNumber,omitempty"`
	Players           []UlarPlayer    `json:"players"`
	LastAction        string          `json:"lastAction,omitempty"`
	LastDice          int             `json:"lastDice,omitempty"`
	WinnerID          string          `json:"winnerId,omitempty"`
	CountdownEnd      int64           `json:"countdownEnd,omitempty"`
	Chat              []UlarChatLine  `json:"chat,omitempty"`
	YouAre            string          `json:"youAre,omitempty"`
	CurrentQuestionID string          `json:"currentQuestionId,omitempty"`
	QuestionPlayerID  string          `json:"questionPlayerId,omitempty"`
	QuestionState     string          `json:"questionState,omitempty"`
	QuestionStartedAt int64           `json:"questionStartedAt,omitempty"`
	QuestionEndsAt    int64           `json:"questionEndsAt,omitempty"`
	AnswerSubmitted   bool            `json:"answerSubmitted,omitempty"`
	Penalty           int             `json:"penalty,omitempty"`
	Question          *QuestionPublic `json:"question,omitempty"`
	Stats             map[string]any  `json:"stats,omitempty"`
	PlayerStats       map[string]any  `json:"playerStats,omitempty"`
	Rewards           []RewardEvent   `json:"rewards,omitempty"`
	MaxPlayers        int             `json:"maxPlayers,omitempty"`
	Mode              string          `json:"mode,omitempty"`
	Visibility        string          `json:"visibility,omitempty"`
	PowerCells        map[int]string  `json:"powerCells,omitempty"`
	Grade             string          `json:"grade,omitempty"`
}

type MoveBroadcast struct {
	UlarEnvelope
	UserID   string     `json:"userId"`
	Username string     `json:"username"`
	Move     MoveResult `json:"move"`
}

type ChatIn struct {
	Text  string `json:"text"`
	Emote string `json:"emote"`
}
