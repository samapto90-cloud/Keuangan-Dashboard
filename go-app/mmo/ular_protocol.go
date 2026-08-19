package mmo

const (
	TypeJoinLobby    = "JOIN_LOBBY"
	TypeLobbyHello   = "LOBBY_HELLO"
	TypeRoomCreate   = "ROOM_CREATE"
	TypeRoomJoin     = "ROOM_JOIN"
	TypeRoomLeave    = "ROOM_LEAVE"
	TypeRoomList     = "ROOM_LIST"
	TypeRoomUpdated  = "ROOM_UPDATED"
	TypePlayerReady  = "PLAYER_READY"
	TypeUlarGameInfo = "ULAR_GAME_INFO"
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
}

type PlayerReadyIn struct {
	Ready bool `json:"ready"`
}

type RoomListOut struct {
	Rooms []UlarRoom `json:"rooms"`
}
