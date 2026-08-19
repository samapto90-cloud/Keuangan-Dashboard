package mmo

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

type UlarLobby struct {
	mu      sync.Mutex
	online  map[string]*Player
	rooms   map[string]*UlarRoom
	byCode  map[string]string
	inRoom  map[string]string
	matches map[string]*UlarMatch
}

func NewUlarLobby() *UlarLobby {
	return &UlarLobby{
		online:  map[string]*Player{},
		rooms:   map[string]*UlarRoom{},
		byCode:  map[string]string{},
		inRoom:  map[string]string{},
		matches: map[string]*UlarMatch{},
	}
}

func (l *UlarLobby) Connect(p *Player) {
	if p == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.online[p.ID] = p
}

func (l *UlarLobby) Disconnect(p *Player) {
	if p == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.online, p.ID)
	if rid, ok := l.inRoom[p.ID]; ok {
		l.leaveLocked(p.ID, rid)
	}
}

func shortID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func roomCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	out := make([]byte, 4)
	for i := 0; i < 4; i++ {
		out[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(out)
}

func (l *UlarLobby) Create(host *Player) *UlarRoom {
	l.mu.Lock()
	defer l.mu.Unlock()
	if host == nil {
		return nil
	}
	if rid, ok := l.inRoom[host.ID]; ok {
		return l.rooms[rid]
	}
	id := "rm-" + shortID()
	code := roomCode()
	for l.byCode[code] != "" {
		code = roomCode()
	}
	room := &UlarRoom{
		ID:         id,
		RoomCode:   code,
		HostID:     host.ID,
		Status:     UlarWaiting,
		MaxPlayers: MAX_PLAYERS,
		CreatedAt:  time.Now().UTC(),
		Players:    []*UlarPlayer{NewUlarPlayer(host.ID, host.Name)},
	}
	l.rooms[id] = room
	l.byCode[code] = id
	l.inRoom[host.ID] = id
	return room
}

func (l *UlarLobby) Join(p *Player, code string) (*UlarRoom, string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if p == nil {
		return nil, "pemain tidak valid"
	}
	id := l.byCode[code]
	room := l.rooms[id]
	if room == nil {
		return nil, "ruangan tidak ditemukan"
	}
	if existing := l.inRoom[p.ID]; existing != "" && existing != room.ID {
		return nil, "sudah di ruangan lain"
	}
	for _, pl := range room.Players {
		if pl.UserID == p.ID {
			pl.IsConnected = true
			pl.Username = p.Name
			return room, ""
		}
	}
	if len(room.Players) >= room.MaxPlayers {
		return nil, "ruangan penuh"
	}
	room.Players = append(room.Players, NewUlarPlayer(p.ID, p.Name))
	l.inRoom[p.ID] = room.ID
	return room, ""
}

func (l *UlarLobby) Leave(playerID string) *UlarRoom {
	l.mu.Lock()
	defer l.mu.Unlock()
	rid := l.inRoom[playerID]
	return l.leaveLocked(playerID, rid)
}

func (l *UlarLobby) leaveLocked(playerID, rid string) *UlarRoom {
	room := l.rooms[rid]
	if room == nil {
		delete(l.inRoom, playerID)
		return nil
	}
	kept := room.Players[:0]
	for _, pl := range room.Players {
		if pl.UserID != playerID {
			kept = append(kept, pl)
		}
	}
	room.Players = kept
	delete(l.inRoom, playerID)
	if len(room.Players) == 0 {
		delete(l.rooms, room.ID)
		delete(l.byCode, room.RoomCode)
		delete(l.matches, room.ID)
		return nil
	}
	if room.HostID == playerID {
		room.HostID = room.Players[0].UserID
	}
	return room
}

func (l *UlarLobby) SetReady(playerID string, ready bool) *UlarRoom {
	l.mu.Lock()
	defer l.mu.Unlock()
	rid := l.inRoom[playerID]
	room := l.rooms[rid]
	if room == nil {
		return nil
	}
	allReady := true
	for _, pl := range room.Players {
		if pl.UserID == playerID {
			pl.IsReady = ready
		}
		if !pl.IsReady {
			allReady = false
		}
	}
	if allReady && len(room.Players) >= 2 {
		room.Status = UlarReady
	} else {
		room.Status = UlarWaiting
	}
	return room
}

func (l *UlarLobby) List() []UlarRoom {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]UlarRoom, 0, len(l.rooms))
	for _, r := range l.rooms {
		if r != nil {
			cp := *r
			cp.Players = append([]*UlarPlayer{}, r.Players...)
			out = append(out, cp)
		}
	}
	return out
}

func (h *Hub) lobbyJoin(p *Player) {
	if h.Lobby == nil {
		h.Lobby = NewUlarLobby()
	}
	h.Lobby.Connect(p)
	hello := LobbyHelloOut{
		Title:      GameTitle,
		Version:    GameVersion,
		Phase:      GamePhase,
		PlayerID:   p.ID,
		Username:   p.Name,
		BoardSize:  BOARD_SIZE,
		MaxPlayers: MAX_PLAYERS,
	}
	select {
	case p.send <- marshal(TypeLobbyHello, hello):
	default:
	}
}

func (h *Hub) lobbyLeave(p *Player) {
	if h.Lobby != nil {
		h.Lobby.Disconnect(p)
	}
	p.CloseSend()
}

func (h *Hub) handlePhase1(msg inbound) {
	p := msg.player
	if p == nil {
		return
	}
	switch msg.env.Type {
	case TypePing:
		var ping PingIn
		_ = json.Unmarshal(msg.env.Data, &ping)
		select {
		case p.send <- marshal(TypePong, PongOut{T: ping.T, St: time.Now().UnixMilli()}):
		default:
		}
	case TypeRoomCreate:
		room := h.Lobby.Create(p)
		if room == nil {
			h.rejectUlar(p, msg.env.Type, "gagal membuat ruangan")
			return
		}
		h.sendRoom(p, room)
	case TypeRoomJoin:
		var in RoomCodeIn
		_ = json.Unmarshal(msg.env.Data, &in)
		room, errMsg := h.Lobby.Join(p, in.RoomCode)
		if errMsg != "" {
			h.rejectUlar(p, msg.env.Type, errMsg)
			return
		}
		h.sendRoom(p, room)
	case TypeRoomLeave:
		h.Lobby.Leave(p.ID)
		select {
		case p.send <- marshal(TypeRoomUpdated, UlarRoom{Status: UlarWaiting}):
		default:
		}
	case TypeRoomList:
		select {
		case p.send <- marshal(TypeRoomList, RoomListOut{Rooms: h.Lobby.List()}):
		default:
		}
	case TypePlayerReady:
		var in PlayerReadyIn
		_ = json.Unmarshal(msg.env.Data, &in)
		room := h.Lobby.SetReady(p.ID, in.Ready)
		if room == nil {
			h.rejectUlar(p, msg.env.Type, "belum di ruangan")
			return
		}
		h.sendRoom(p, room)
	case TypeUlarGameInfo:
		select {
		case p.send <- marshal(TypeUlarGameInfo, map[string]any{
			"boardSize": BOARD_SIZE,
			"min":       MIN_POSITION,
			"max":       MAX_POSITION,
			"snakes":    DefaultSnakes,
			"ladders":   DefaultLadders,
			"gameplay":  "not_implemented",
		}):
		default:
		}
	default:
		h.rejectUlar(p, msg.env.Type, "legacy_gameplay_retired")
	}
}

func (h *Hub) sendRoom(p *Player, room *UlarRoom) {
	if p == nil || room == nil {
		return
	}
	select {
	case p.send <- marshal(TypeRoomUpdated, room):
	default:
	}
}

func (h *Hub) rejectUlar(p *Player, action, reason string) {
	select {
	case p.send <- marshal(TypeActionReject, RejectOut{Action: action, Reason: reason, PlayerID: p.ID}):
	default:
	}
}
