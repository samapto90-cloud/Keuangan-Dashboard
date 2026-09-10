package mmo

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

type UlarLobby struct {
	mu        sync.Mutex
	online    map[string]*Player
	rooms     map[string]*UlarRoom
	byCode    map[string]string
	inRoom    map[string]string
	matches   map[string]*UlarMatch
	joinAsks  map[string]joinAsk
	Store     *MatchStore
	Attempts  *AttemptStore
	limit     *ularLimiter
	OnAbandon func(userID, matchID string)
}

type joinAsk struct {
	ID        string
	FromID    string
	FromName  string
	HostID    string
	RoomCode  string
	ExpiresAt int64
}

func NewUlarLobby() *UlarLobby {
	return &UlarLobby{
		online:   map[string]*Player{},
		rooms:    map[string]*UlarRoom{},
		byCode:   map[string]string{},
		inRoom:   map[string]string{},
		matches:  map[string]*UlarMatch{},
		joinAsks: map[string]joinAsk{},
		Store:    OpenMatchStore(matchStorePath()),
		Attempts: OpenAttemptStore(attemptStorePath()),
		limit:    newUlarLimiter(),
	}
}

func shortID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func roomCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, RoomCodeLen)
	_, _ = rand.Read(b)
	out := make([]byte, RoomCodeLen)
	for i := 0; i < RoomCodeLen; i++ {
		out[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(out)
}

func (l *UlarLobby) Connect(p *Player) *UlarRoom {
	if p == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.online[p.ID] = p
	rid := l.inRoom[p.ID]
	room := l.rooms[rid]
	if room == nil {
		return nil
	}
	pl := l.player(room, p.ID)
	if pl != nil {
		pl.IsConnected = true
		pl.ConnState = "CONNECTED"
		pl.DisconnectedAt = time.Time{}
		pl.Username = p.Name
		log.Printf("PLAYER_RECONNECTED %s room=%s", p.ID, room.RoomCode)
	}
	return room
}

func (l *UlarLobby) Disconnect(p *Player) {
	if p == nil {
		return
	}
	l.mu.Lock()
	// Socket lama setelah refresh: jangan hapus sesi baru.
	if cur, ok := l.online[p.ID]; ok && cur != p {
		l.mu.Unlock()
		return
	}
	delete(l.online, p.ID)
	rid := l.inRoom[p.ID]
	room := l.rooms[rid]
	grace := DisconnectGrace
	if room != nil {
		pl := l.player(room, p.ID)
		if pl != nil {
			pl.IsConnected = false
			pl.ConnState = "DISCONNECTED"
			pl.DisconnectedAt = time.Now()
			log.Printf("PLAYER_DISCONNECTED %s room=%s", p.ID, room.RoomCode)
			// Jangan ganti host di sini — blip reconnect sering membuat MULAI gagal (NOT_HOST).
			waiting := room.Status == UlarWaiting || room.Status == UlarReady || room.Status == UlarFinished
			if waiting {
				grace = LobbyDisconnectGrace
			}
		}
	}
	l.mu.Unlock()
	go func(id string, waitSec int) {
		time.Sleep(time.Duration(waitSec) * time.Second)
		l.expireDisconnect(id)
	}(p.ID, grace)
}

func (l *UlarLobby) expireDisconnect(userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if cur := l.online[userID]; cur != nil {
		return
	}
	rid := l.inRoom[userID]
	room := l.rooms[rid]
	if room == nil {
		// Soft-leave mungkin sudah hapus inRoom tetapi kursi masih ada.
		for _, r := range l.rooms {
			if r == nil {
				continue
			}
			if pl := l.player(r, userID); pl != nil && !pl.IsConnected {
				l.removePlayerSeatLocked(r, userID)
				if l.connectedCount(r) == 0 {
					l.closeLocked(r)
				}
				return
			}
		}
		return
	}
	pl := l.player(room, userID)
	if pl == nil || pl.IsConnected {
		return
	}
	playing := room.Status == UlarPlaying || room.Status == UlarStarting
	if playing {
		if !pl.Abandoned && room.Mode == "RANKED" {
			pl.Abandoned = true
			mid := ""
			if room.Match != nil {
				mid = room.Match.ID
			}
			cb := l.OnAbandon
			go func(uid, matchID string, fn func(string, string)) {
				if fn != nil {
					fn(uid, matchID)
				}
			}(userID, mid, cb)
		}
		l.leaveLocked(userID, rid)
		if room.Status != UlarClosed && l.connectedCount(room) == 0 {
			l.closeLocked(room)
		}
		return
	}
	l.leaveLocked(userID, rid)
	if room.Status != UlarClosed && (l.connectedCount(room) == 0 || len(room.Players) == 0) {
		l.closeLocked(room)
	}
}

// freeGhostSeatsLocked membuang pemain offline dari kapasitas room (lobby / soft-leave).
func (l *UlarLobby) freeGhostSeatsLocked(room *UlarRoom) {
	if room == nil {
		return
	}
	waiting := room.Status == UlarWaiting || room.Status == UlarReady || room.Status == UlarFinished
	kept := room.Players[:0]
	for _, pl := range room.Players {
		if pl == nil {
			continue
		}
		if pl.IsConnected {
			kept = append(kept, pl)
			continue
		}
		if waiting {
			delete(l.inRoom, pl.UserID)
			continue
		}
		// Playing: buang yang sudah soft-leave (tidak ada inRoom) atau lewat grace.
		if _, mapped := l.inRoom[pl.UserID]; !mapped {
			continue
		}
		if !pl.DisconnectedAt.IsZero() && time.Since(pl.DisconnectedAt) >= time.Duration(DisconnectGrace)*time.Second {
			delete(l.inRoom, pl.UserID)
			continue
		}
		kept = append(kept, pl)
	}
	room.Players = kept
	if room.HostID != "" && l.player(room, room.HostID) == nil {
		room.HostID = l.pickHostLocked(room)
	}
}

func (l *UlarLobby) removePlayerSeatLocked(room *UlarRoom, playerID string) {
	if room == nil {
		return
	}
	kept := room.Players[:0]
	for _, pl := range room.Players {
		if pl != nil && pl.UserID != playerID {
			kept = append(kept, pl)
		}
	}
	room.Players = kept
	delete(l.inRoom, playerID)
	if room.HostID == playerID {
		room.HostID = l.pickHostLocked(room)
	}
}

func (l *UlarLobby) closeLocked(room *UlarRoom) {
	if room == nil {
		return
	}
	room.Status = UlarClosed
	delete(l.rooms, room.ID)
	delete(l.byCode, room.RoomCode)
	for _, p := range room.Players {
		if p != nil {
			delete(l.inRoom, p.UserID)
		}
	}
	log.Printf("ROOM_CLOSED %s", room.RoomCode)
}

func (l *UlarLobby) pickHostLocked(room *UlarRoom) string {
	var best *UlarPlayer
	for _, o := range room.Players {
		if o == nil || !o.IsConnected {
			continue
		}
		if best == nil || o.JoinedAt.Before(best.JoinedAt) {
			best = o
		}
	}
	if best != nil {
		return best.UserID
	}
	if len(room.Players) > 0 && room.Players[0] != nil {
		return room.Players[0].UserID
	}
	return room.HostID
}

func (l *UlarLobby) Create(host *Player) (*UlarRoom, string) {
	return l.CreateSized(host, MAX_PLAYERS, "")
}

func (l *UlarLobby) CreateSized(host *Player, maxPlayers int, grade string) (*UlarRoom, string) {
	if host == nil {
		return nil, ErrInvalidRequest
	}
	if DefaultHub != nil && DefaultHub.Ops != nil && DefaultHub.Ops.ActiveBan(host.ID) != nil {
		return nil, "akun diblokir"
	}
	if maxPlayers < 2 {
		maxPlayers = 2
	}
	if maxPlayers > MAX_PLAYERS {
		maxPlayers = MAX_PLAYERS
	}
	wantGrade := normalizeGrade(grade)
	l.mu.Lock()
	if rid, ok := l.inRoom[host.ID]; ok {
		if r := l.rooms[rid]; r != nil {
			// Room selesai / tertutup — lepaskan supaya bisa buat baru.
			if r.Status == UlarFinished || r.Status == UlarClosed {
				l.leaveLocked(host.ID, rid)
			} else if r.HostID == host.ID && r.Status == UlarWaiting && r.Visibility != "MATCHMADE" && len(r.Players) <= maxPlayers {
				r.MaxPlayers = maxPlayers
				if grade != "" || r.Grade == "" {
					r.Grade = wantGrade
				}
				l.mu.Unlock()
				return r, ""
			} else if r.Status == UlarWaiting || r.Status == UlarReady {
				if grade != "" && r.HostID == host.ID {
					r.Grade = wantGrade
				}
				l.mu.Unlock()
				return r, ""
			} else {
				// Playing/starting — jangan buat room baru; kembalikan yang aktif.
				l.mu.Unlock()
				return r, ""
			}
		} else {
			delete(l.inRoom, host.ID)
		}
	}
	l.mu.Unlock()
	if !l.limit.allow("create:"+host.ID, 8, 10*time.Minute) {
		return nil, ErrRateLimited
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if rid, ok := l.inRoom[host.ID]; ok {
		if r := l.rooms[rid]; r != nil {
			if grade != "" && r.HostID == host.ID && (r.Status == UlarWaiting || r.Status == UlarReady) {
				r.Grade = wantGrade
			}
			return r, ""
		}
		delete(l.inRoom, host.ID)
	}
	id := "rm-" + shortID()
	code := roomCode()
	for l.byCode[code] != "" {
		code = roomCode()
	}
	room := &UlarRoom{
		ID: id, RoomCode: code, HostID: host.ID, Status: UlarWaiting,
		MaxPlayers: maxPlayers, CreatedAt: time.Now().UTC(), Visibility: "PRIVATE",
		Grade:     wantGrade,
		Players:   []*UlarPlayer{NewUlarPlayer(host.ID, host.Name, 0)},
		SeenEvent: map[string]bool{},
	}
	l.rooms[id] = room
	l.byCode[code] = id
	l.inRoom[host.ID] = id
	log.Printf("ROOM_CREATED %s host=%s grade=%s", code, host.ID, wantGrade)
	return room, ""
}

func (l *UlarLobby) CreateFilled(players []*Player, mode string) (*UlarRoom, string) {
	if len(players) < 2 || players[0] == nil {
		return nil, ErrNeedPlayers
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, p := range players {
		if p == nil {
			continue
		}
		if rid := l.inRoom[p.ID]; rid != "" {
			l.leaveLocked(p.ID, rid)
		}
	}
	id := "rm-" + shortID()
	code := roomCode()
	for l.byCode[code] != "" {
		code = roomCode()
	}
	max := len(players)
	if max < 2 {
		max = 2
	}
	if max > MAX_PLAYERS {
		max = MAX_PLAYERS
	}
	host := players[0]
	room := &UlarRoom{
		ID: id, RoomCode: code, HostID: host.ID, Status: UlarWaiting,
		MaxPlayers: max, CreatedAt: time.Now().UTC(), Visibility: "MATCHMADE", Mode: mode,
		Players: []*UlarPlayer{}, SeenEvent: map[string]bool{},
	}
	for i, p := range players {
		if p == nil {
			continue
		}
		pl := NewUlarPlayer(p.ID, p.Name, i)
		room.Players = append(room.Players, pl)
		l.inRoom[p.ID] = id
	}
	l.rooms[id] = room
	l.byCode[code] = id
	return room, ""
}

func (l *UlarLobby) Kick(hostID, targetID string) (*UlarRoom, string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	room := l.rooms[l.inRoom[hostID]]
	if room == nil {
		return nil, ErrNotInRoom
	}
	if room.HostID != hostID {
		return nil, ErrNotHost
	}
	if targetID == hostID {
		return nil, ErrInvalidRequest
	}
	if room.Status == UlarPlaying || room.Status == UlarStarting {
		return nil, ErrLocked
	}
	if l.player(room, targetID) == nil {
		return nil, ErrNotInRoom
	}
	return l.leaveLocked(targetID, room.ID), ""
}

func (l *UlarLobby) Join(p *Player, code string) (*UlarRoom, string) {
	if p == nil {
		return nil, ErrInvalidRequest
	}
	if DefaultHub != nil && DefaultHub.Ops != nil && DefaultHub.Ops.ActiveBan(p.ID) != nil {
		return nil, "akun diblokir"
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if !l.limit.allow("join:"+p.ID, 20, 10*time.Minute) {
		return nil, ErrRateLimited
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	room := l.rooms[l.byCode[code]]
	if room == nil || room.Status == UlarClosed {
		return nil, ErrNotInRoom
	}
	l.freeGhostSeatsLocked(room)
	if room.Status == UlarClosed {
		return nil, ErrNotInRoom
	}
	if room.Status == UlarPlaying || room.Status == UlarStarting || room.Status == UlarFinished {
		if existing := l.player(room, p.ID); existing != nil {
			existing.IsConnected = true
			existing.ConnState = "CONNECTED"
			existing.DisconnectedAt = time.Time{}
			l.inRoom[p.ID] = room.ID
			return room, ""
		}
		// Izinkan bergabung terlambat jika masih ada slot (3rd/4th).
		if room.Status == UlarPlaying && len(room.Players) < room.MaxPlayers {
			slot := len(room.Players)
			np := NewUlarPlayer(p.ID, p.Name, slot)
			np.Position = OFFBOARD_START
			np.PlayState = "PLAYING"
			np.IsReady = true
			room.Players = append(room.Players, np)
			l.inRoom[p.ID] = room.ID
			log.Printf("PLAYER_JOINED_LATE %s room=%s n=%d", p.ID, room.RoomCode, len(room.Players))
			return room, ""
		}
		return nil, ErrGameNotStarted
	}
	if existing := l.inRoom[p.ID]; existing != "" && existing != room.ID {
		old := l.rooms[existing]
		if old != nil && (old.Status == UlarWaiting || old.Status == UlarReady || old.Status == UlarFinished) {
			l.leaveLocked(p.ID, existing)
		} else if old != nil {
			return nil, ErrInvalidRequest
		} else {
			delete(l.inRoom, p.ID)
		}
	}
	if pl := l.player(room, p.ID); pl != nil {
		pl.IsConnected = true
		pl.ConnState = "CONNECTED"
		pl.DisconnectedAt = time.Time{}
		pl.Username = p.Name
		l.inRoom[p.ID] = room.ID
		return room, ""
	}
	if len(room.Players) >= room.MaxPlayers {
		return nil, ErrRoomFull
	}
	slot := len(room.Players)
	room.Players = append(room.Players, NewUlarPlayer(p.ID, p.Name, slot))
	l.inRoom[p.ID] = room.ID
	log.Printf("PLAYER_JOINED %s room=%s n=%d", p.ID, room.RoomCode, len(room.Players))
	return room, ""
}

func (l *UlarLobby) Leave(playerID string) *UlarRoom {
	l.mu.Lock()
	defer l.mu.Unlock()
	rid := l.inRoom[playerID]
	room := l.leaveLocked(playerID, rid)
	if room != nil && room.Status != UlarClosed && l.connectedCount(room) == 0 {
		l.closeLocked(room)
		return nil
	}
	return room
}

func (l *UlarLobby) leaveLocked(playerID, rid string) *UlarRoom {
	room := l.rooms[rid]
	if room == nil {
		delete(l.inRoom, playerID)
		return nil
	}
	playing := room.Status == UlarPlaying || room.Status == UlarStarting
	if playing {
		if pl := l.player(room, playerID); pl != nil && !pl.Abandoned && room.Mode == "RANKED" {
			pl.Abandoned = true
			mid := ""
			if room.Match != nil {
				mid = room.Match.ID
			}
			cb := l.OnAbandon
			go func(uid, matchID string, fn func(string, string)) {
				if fn != nil {
					fn(uid, matchID)
				}
			}(playerID, mid, cb)
		}
	}
	kept := room.Players[:0]
	for _, pl := range room.Players {
		if pl != nil && pl.UserID != playerID {
			kept = append(kept, pl)
		}
	}
	room.Players = kept
	delete(l.inRoom, playerID)
	log.Printf("PLAYER_LEAVE %s room=%s", playerID, room.RoomCode)
	if len(room.Players) == 0 {
		l.closeLocked(room)
		return nil
	}
	if room.HostID == playerID {
		room.HostID = l.pickHostLocked(room)
	}
	return room
}

func (l *UlarLobby) SetReady(playerID string, ready bool) (*UlarRoom, string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	room := l.rooms[l.inRoom[playerID]]
	if room == nil {
		return nil, ErrNotInRoom
	}
	pl := l.player(room, playerID)
	if pl == nil {
		return nil, ErrNotInRoom
	}
	pl.IsReady = ready
	if ready {
		pl.PlayState = "READY"
	} else {
		pl.PlayState = "NOT_READY"
	}
	if l.allReady(room) {
		room.Status = UlarReady
	} else if room.Status != UlarStarting && room.Status != UlarPlaying {
		room.Status = UlarWaiting
	}
	log.Printf("PLAYER_READY %s ready=%v room=%s", playerID, ready, room.RoomCode)
	return room, ""
}

func (l *UlarLobby) PublicList() []RoomPublic {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]RoomPublic, 0, len(l.rooms))
	for _, r := range l.rooms {
		if r == nil || r.Status == UlarClosed || r.Status == UlarPlaying || r.Status == UlarFinished || r.Visibility == "PRIVATE" || r.Visibility == "MATCHMADE" {
			continue
		}
		out = append(out, RoomPublic{RoomCode: r.RoomCode, Players: len(r.Players), Max: r.MaxPlayers, Status: string(r.Status)})
	}
	return out
}

func (l *UlarLobby) List() []UlarRoom {
	pubs := l.PublicList()
	out := make([]UlarRoom, 0, len(pubs))
	for _, p := range pubs {
		out = append(out, UlarRoom{RoomCode: p.RoomCode, MaxPlayers: p.Max, Status: UlarMatchStatus(p.Status)})
	}
	return out
}

func (h *Hub) lobbyJoin(p *Player) {
	if h.Lobby == nil {
		h.Lobby = NewUlarLobby()
	}
	room := h.Lobby.Connect(p)
	if h.Social != nil {
		h.Social.Touch(p.ID, false)
	}
	hello := LobbyHelloOut{
		Title: GameTitle, Version: GameVersion, Phase: GamePhase,
		PlayerID: p.ID, Username: p.Name, BoardSize: BOARD_SIZE, MaxPlayers: MAX_PLAYERS,
	}
	select {
	case p.send <- marshal(TypeLobbyHello, hello):
	default:
	}
	if room != nil {
		h.Lobby.mu.Lock()
		h.overdueQuestionLocked(room)
		snap := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, p.ID))
		h.Lobby.mu.Unlock()
		h.pushRoom(room, TypePlayerRecon, snap)
		select {
		case p.send <- marshal(TypeGameState, snap):
		default:
		}
	}
	h.broadcastOnlineList()
	h.pushOnlineListTo(p)
}

func (h *Hub) lobbyLeave(p *Player) {
	if h.Lobby != nil {
		roomID := ""
		h.Lobby.mu.Lock()
		roomID = h.Lobby.inRoom[p.ID]
		room := h.Lobby.rooms[roomID]
		h.Lobby.mu.Unlock()
		h.Lobby.Disconnect(p)
		if room != nil {
			h.Lobby.mu.Lock()
			snap := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, ""))
			h.Lobby.mu.Unlock()
			h.pushRoom(room, TypePlayerDisc, snap)
		}
	}
	h.broadcastOnlineList()
	p.CloseSend()
}

func (h *Hub) onlineLobbyCards(viewerID string) []map[string]any {
	if h.Lobby == nil {
		return nil
	}
	h.Lobby.mu.Lock()
	defer h.Lobby.mu.Unlock()
	out := make([]map[string]any, 0, len(h.Lobby.online))
	for id, pl := range h.Lobby.online {
		if pl == nil || id == viewerID {
			continue
		}
		if h.Social != nil && h.Social.Blocked(viewerID, id) {
			continue
		}
		status := PresOnline
		inGame := false
		seats, maxP := 0, 0
		roomCode := ""
		canAsk := false
		if rid := h.Lobby.inRoom[id]; rid != "" {
			if rm := h.Lobby.rooms[rid]; rm != nil {
				seats = len(rm.Players)
				maxP = rm.MaxPlayers
				roomCode = rm.RoomCode
				if rm.Status == UlarPlaying || rm.Status == UlarStarting {
					status = PresInGame
					inGame = true
					canAsk = seats < maxP
				} else if rm.Status == UlarWaiting || rm.Status == UlarReady {
					canAsk = seats < maxP
				}
			}
		} else {
			canAsk = true // online di lobby — bisa diundang / diajak
		}
		friends := false
		if h.Social != nil {
			friends = h.Social.AreFriends(viewerID, id)
		}
		out = append(out, map[string]any{
			"userId": id, "username": pl.Name, "status": status, "inGame": inGame,
			"friends": friends, "seats": seats, "maxPlayers": maxP, "roomCode": roomCode, "canAsk": canAsk,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(str(out[i]["username"])) < strings.ToLower(str(out[j]["username"]))
	})
	return out
}

func (h *Hub) pushOnlineListTo(p *Player) {
	if p == nil {
		return
	}
	select {
	case p.send <- marshal(TypeOnlineList, map[string]any{"players": h.onlineLobbyCards(p.ID)}):
	default:
	}
}

func (h *Hub) broadcastOnlineList() {
	if h.Lobby == nil {
		return
	}
	h.Lobby.mu.Lock()
	online := make([]*Player, 0, len(h.Lobby.online))
	for _, pl := range h.Lobby.online {
		if pl != nil {
			online = append(online, pl)
		}
	}
	h.Lobby.mu.Unlock()
	for _, pl := range online {
		h.pushOnlineListTo(pl)
	}
}

func (h *Hub) emitRoom(p *Player, room *UlarRoom) {
	if room == nil || p == nil {
		return
	}
	h.Lobby.mu.Lock()
	snap := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, p.ID))
	h.Lobby.mu.Unlock()
	h.pushRoom(room, TypeRoomUpdated, snap)
}

func (h *Hub) handlePhase1(msg inbound) {
	p := msg.player
	if p == nil || h.Lobby == nil {
		return
	}
	switch msg.env.Type {
	case TypePing:
		var ping PingIn
		_ = json.Unmarshal(msg.env.Data, &ping)
		if h.Social != nil {
			inGame := false
			h.Lobby.mu.Lock()
			if rid := h.Lobby.inRoom[p.ID]; rid != "" {
				if rm := h.Lobby.rooms[rid]; rm != nil && (rm.Status == UlarPlaying || rm.Status == UlarStarting) {
					inGame = true
				}
			}
			h.Lobby.mu.Unlock()
			h.Social.Touch(p.ID, inGame)
		}
		select {
		case p.send <- marshal(TypePong, PongOut{T: ping.T, St: time.Now().UnixMilli()}):
		default:
		}
	case TypeRoomCreate:
		if h.Ops != nil {
			if ban := h.Ops.ActiveBan(p.ID); ban != nil {
				h.rejectCode(p, msg.env.Type, "banned", "Akun diblokir.")
				return
			}
		}
		var cin struct {
			MaxPlayers int    `json:"maxPlayers"`
			Grade      string `json:"grade"`
		}
		_ = json.Unmarshal(msg.env.Data, &cin)
		max := cin.MaxPlayers
		if max == 0 {
			max = MAX_PLAYERS
		}
		room, errc := h.Lobby.CreateSized(p, max, cin.Grade)
		if errc != "" {
			h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
			return
		}
		h.emitRoom(p, room)
	case TypeRoomJoin:
		var in RoomCodeIn
		_ = json.Unmarshal(msg.env.Data, &in)
		if h.Social != nil {
			h.Lobby.mu.Lock()
			rm := h.Lobby.rooms[h.Lobby.byCode[strings.ToUpper(strings.TrimSpace(in.RoomCode))]]
			host := ""
			if rm != nil {
				host = rm.HostID
			}
			h.Lobby.mu.Unlock()
			if host != "" && h.Social.Blocked(p.ID, host) {
				h.rejectCode(p, msg.env.Type, "blocked", "Tidak dapat bergabung.")
				return
			}
		}
		if h.Ops != nil {
			if ban := h.Ops.ActiveBan(p.ID); ban != nil {
				h.rejectCode(p, msg.env.Type, "banned", "Akun diblokir.")
				return
			}
		}
		room, errc := h.Lobby.Join(p, in.RoomCode)
		if errc != "" {
			h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
			return
		}
		h.emitRoom(p, room)
	case TypeRoomKick:
		var in struct {
			UserID string `json:"userId"`
		}
		_ = json.Unmarshal(msg.env.Data, &in)
		room, errc := h.Lobby.Kick(p.ID, in.UserID)
		if errc != "" {
			h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
			return
		}
		if tgt := h.Lobby.online[in.UserID]; tgt != nil {
			select {
			case tgt.send <- marshal(TypeUlarError, map[string]string{"code": "KICKED", "message": "You were removed from room."}):
			default:
			}
		}
		h.emitRoom(p, room)
	case TypeQueueJoin:
		var in struct {
			Mode          string `json:"mode"`
			Region        string `json:"region"`
			Grade         string `json:"grade"`
			Rank          string `json:"rank"`
			RR            int    `json:"rr"`
			MMR           int    `json:"mmr"`
			UserID        string `json:"userId"`
			PreferredSize int    `json:"preferredSize"`
			MaxPlayers    int    `json:"maxPlayers"`
		}
		_ = json.Unmarshal(msg.env.Data, &in)
		mode := strings.ToUpper(strings.TrimSpace(in.Mode))
		if mode != "RANKED" {
			mode = "CASUAL"
		}
		size := in.PreferredSize
		if size == 0 {
			size = in.MaxPlayers
		}
		if qerr := h.queueJoin(p, mode, in.Region, in.Grade, size); qerr != "" {
			h.rejectCode(p, msg.env.Type, qerr, qerr)
			return
		}
		select {
		case p.send <- marshal(TypeQueueUpdate, map[string]any{"mode": mode, "searching": true, "preferredSize": size}):
		default:
		}
	case TypeQueueLeave:
		if h.Matchmaker != nil {
			h.Matchmaker.Cancel(p.ID)
		}
		select {
		case p.send <- marshal(TypeQueueUpdate, map[string]any{"searching": false}):
		default:
		}
	case TypeMatchReady:
		var in struct {
			Ready bool `json:"ready"`
		}
		_ = json.Unmarshal(msg.env.Data, &in)
		h.matchReady(p, in.Ready)
	case TypeGameInviteEv:
		var in struct {
			UserID     string `json:"userId"`
			MaxPlayers int    `json:"maxPlayers"`
			Grade      string `json:"grade"`
		}
		_ = json.Unmarshal(msg.env.Data, &in)
		h.Lobby.mu.Lock()
		room := h.Lobby.rooms[h.Lobby.inRoom[p.ID]]
		code, rid := "", ""
		playingInvitee := false
		if room != nil {
			if pl := h.Lobby.player(room, in.UserID); pl != nil && (room.Status == UlarPlaying || room.Status == UlarStarting) {
				playingInvitee = true
			}
			code, rid = room.RoomCode, room.ID
		}
		tgtRoom := h.Lobby.rooms[h.Lobby.inRoom[in.UserID]]
		if tgtRoom != nil && (tgtRoom.Status == UlarPlaying || tgtRoom.Status == UlarStarting) {
			playingInvitee = true
		}
		h.Lobby.mu.Unlock()
		if playingInvitee {
			h.rejectCode(p, msg.env.Type, "in_game", "Pemain sedang bermain. Minta bergabung atau buat permainan sendiri.")
			return
		}
		if room == nil {
			max := in.MaxPlayers
			if max < 2 {
				max = 2
			}
			if max > MAX_PLAYERS {
				max = MAX_PLAYERS
			}
			created, errc := h.Lobby.CreateSized(p, max, in.Grade)
			if errc != "" {
				h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
				return
			}
			room = created
			code, rid = room.RoomCode, room.ID
			h.emitRoom(p, room)
		} else if in.Grade != "" {
			h.Lobby.mu.Lock()
			if room.HostID == p.ID && (room.Status == UlarWaiting || room.Status == UlarReady) {
				room.Grade = normalizeGrade(in.Grade)
			}
			h.Lobby.mu.Unlock()
		}
		if h.Social == nil {
			h.rejectCode(p, msg.env.Type, "social", "Sistem undangan tidak siap.")
			return
		}
		inv, ierr := h.Social.CreateInvite(p.ID, in.UserID, rid, code)
		if ierr != "" {
			h.rejectCode(p, msg.env.Type, ierr, ierr)
			return
		}
		h.pushUser(in.UserID, TypeGameInviteEv, map[string]any{
			"inviteId": inv.ID, "senderId": p.ID, "username": p.Name, "roomCode": code, "expiresAt": inv.ExpiresAt,
			"maxPlayers": room.MaxPlayers,
		})
		h.emitRoom(p, room)
	case TypeOnlineListReq:
		h.pushOnlineListTo(p)
	case TypeJoinAsk:
		var in struct {
			UserID string `json:"userId"`
			Grade  string `json:"grade"`
		}
		_ = json.Unmarshal(msg.env.Data, &in)
		h.handleJoinAsk(p, in.UserID, in.Grade)
	case TypeJoinAskRespond:
		var in struct {
			AskID  string `json:"askId"`
			Accept bool   `json:"accept"`
		}
		_ = json.Unmarshal(msg.env.Data, &in)
		h.handleJoinAskRespond(p, in.AskID, in.Accept)
	case TypeInviteRespond:
		var in struct {
			InviteID string `json:"inviteId"`
			Accept   bool   `json:"accept"`
		}
		_ = json.Unmarshal(msg.env.Data, &in)
		if h.Social == nil {
			return
		}
		act := "reject"
		if in.Accept {
			act = "accept"
		}
		inv, ierr := h.Social.RespondInvite(p.ID, in.InviteID, act)
		if ierr != "" {
			h.rejectCode(p, msg.env.Type, ierr, ierr)
			return
		}
		if in.Accept && inv.RoomCode != "" {
			room, errc := h.Lobby.Join(p, inv.RoomCode)
			if errc != "" {
				h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
				return
			}
			h.emitRoom(p, room)
			h.pushUser(inv.SenderID, TypeInviteResult, map[string]any{
				"inviteId": inv.ID, "accepted": true, "username": p.Name,
			})
		} else {
			h.pushUser(inv.SenderID, TypeInviteResult, map[string]any{
				"inviteId": inv.ID, "accepted": false, "username": p.Name,
			})
			h.pushUser(inv.SenderID, TypeSocialPush, map[string]any{
				"type": "INVITE_DECLINED", "title": "Undangan ditolak",
				"message": p.Name + " menolak undangan bermain.",
			})
		}
	case TypeRoomLeave:
		room := h.Lobby.Leave(p.ID)
		if room != nil {
			h.emitRoom(p, room)
		}
	case TypeRoomList:
		select {
		case p.send <- marshal(TypeRoomList, RoomListOut{Rooms: h.Lobby.PublicList()}):
		default:
		}
	case TypePlayerReady:
		var in PlayerReadyIn
		_ = json.Unmarshal(msg.env.Data, &in)
		room, errc := h.Lobby.SetReady(p.ID, in.Ready)
		if errc != "" {
			h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
			return
		}
		h.emitRoom(p, room)
		// Semua ready → langsung mulai (tanpa menunggu tombol MULAI yang sering gagal).
		if room != nil && room.Status == UlarReady {
			host := room.HostID
			if host == "" {
				host = p.ID
			}
			if errStart := h.startMatch(host); errStart != "" {
				log.Printf("AUTO_START_FAIL room=%s err=%s", room.RoomCode, errStart)
			}
		}
	case TypeRoomStart:
		if errc := h.startMatch(p.ID); errc != "" {
			h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
		}
	case TypeRoomClose:
		h.Lobby.mu.Lock()
		room := h.Lobby.rooms[h.Lobby.inRoom[p.ID]]
		if room == nil || room.HostID != p.ID {
			h.Lobby.mu.Unlock()
			h.rejectCode(p, msg.env.Type, ErrNotHost, friendlyUlar(ErrNotHost))
			return
		}
		h.Lobby.closeLocked(room)
		h.Lobby.mu.Unlock()
	case TypeGameRoll:
		if !h.Lobby.limit.allow("roll:"+p.ID, 8, 2*time.Second) {
			h.rejectCode(p, msg.env.Type, ErrRateLimited, friendlyUlar(ErrRateLimited))
			return
		}
		if errc := h.rollMatch(p); errc != "" {
			h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
		}
	case TypeAnimDone:
		h.animDone(p)
	case TypeQuestionAnswer:
		var in QuestionAnswerIn
		_ = json.Unmarshal(msg.env.Data, &in)
		if errc := h.submitQuestion(p, in.Answer); errc != "" {
			h.rejectCode(p, msg.env.Type, errc, friendlyUlar(errc))
		}
	case TypeUsePower:
		var in UsePowerIn
		_ = json.Unmarshal(msg.env.Data, &in)
		if !h.Lobby.limit.allow("power:"+p.ID, 6, 2*time.Second) {
			h.rejectCode(p, msg.env.Type, ErrRateLimited, friendlyUlar(ErrRateLimited))
			return
		}
		if errc := h.usePower(p, in.Item, in.TargetID); errc != "" {
			msgText := friendlyUlar(errc)
			if errc == "no_item" {
				msgText = "Item tidak tersedia."
			}
			h.rejectCode(p, msg.env.Type, errc, msgText)
		}
	case TypeQuestionHistory:
		h.sendQuestionHistory(p)
	case TypeRoomChat, TypeRoomEmote:
		h.handleChat(p, msg.env)
	case TypeUlarGameInfo:
		select {
		case p.send <- marshal(TypeUlarGameInfo, map[string]any{
			"boardSize": BOARD_SIZE, "min": MIN_POSITION, "max": MAX_POSITION,
			"snakes": DefaultSnakes, "ladders": DefaultLadders, "gameplay": "multiplayer",
		}):
		default:
		}
	default:
		h.rejectCode(p, msg.env.Type, "legacy_gameplay_retired", "Perintah tidak tersedia.")
	}
}

func (h *Hub) handleChat(p *Player, env Envelope) {
	if !LiveFlags().EnableChat {
		h.rejectCode(p, env.Type, ErrLocked, "Chat nonaktif.")
		return
	}
	if h.Ops != nil && h.Ops.ChatMuted(p.ID) {
		h.rejectCode(p, env.Type, "muted", "Chat dibatasi.")
		return
	}
	var in ChatIn
	_ = json.Unmarshal(env.Data, &in)
	text := clampChat(in.Text)
	emote := allowedEmote(in.Emote)
	if text == "" && emote == "" {
		return
	}
	if !h.Lobby.limit.allow("chat:"+p.ID, ChatBurstMax, ChatBurstWin*time.Second) {
		h.rejectCode(p, env.Type, ErrRateLimited, friendlyUlar(ErrRateLimited))
		return
	}
	h.Lobby.mu.Lock()
	room := h.Lobby.rooms[h.Lobby.inRoom[p.ID]]
	if room == nil {
		h.Lobby.mu.Unlock()
		h.rejectCode(p, env.Type, ErrNotInRoom, friendlyUlar(ErrNotInRoom))
		return
	}
	line := UlarChatLine{UserID: p.ID, Username: p.Name, Text: text, Emote: emote, At: time.Now().UnixMilli()}
	room.Chat = append(room.Chat, line)
	if len(room.Chat) > 40 {
		room.Chat = room.Chat[len(room.Chat)-40:]
	}
	seq, id, at := h.Lobby.nextSeq(room)
	payload := struct {
		UlarEnvelope
		UlarChatLine
	}{UlarEnvelope: UlarEnvelope{Seq: seq, EventID: id, At: at}, UlarChatLine: line}
	h.Lobby.mu.Unlock()
	h.pushRoom(room, TypeRoomChat, payload)
}

func (h *Hub) handleJoinAsk(p *Player, targetID, grade string) {
	if p == nil || targetID == "" || targetID == p.ID {
		h.rejectCode(p, TypeJoinAsk, "invalid", "Pemain tidak valid.")
		return
	}
	h.Lobby.mu.Lock()
	tgt := h.Lobby.online[targetID]
	if tgt == nil {
		h.Lobby.mu.Unlock()
		h.rejectCode(p, TypeJoinAsk, "offline", "Pemain tidak online.")
		return
	}
	rid := h.Lobby.inRoom[targetID]
	room := h.Lobby.rooms[rid]
	if room == nil {
		h.Lobby.mu.Unlock()
		// Target di lobby — undang ke room baru milik peminta.
		created, errc := h.Lobby.CreateSized(p, 4, grade)
		if errc != "" {
			h.rejectCode(p, TypeJoinAsk, errc, friendlyUlar(errc))
			return
		}
		h.emitRoom(p, created)
		if h.Social == nil {
			return
		}
		inv, ierr := h.Social.CreateInvite(p.ID, targetID, created.ID, created.RoomCode)
		if ierr != "" {
			h.rejectCode(p, TypeJoinAsk, ierr, ierr)
			return
		}
		h.pushUser(targetID, TypeGameInviteEv, map[string]any{
			"inviteId": inv.ID, "senderId": p.ID, "username": p.Name, "roomCode": created.RoomCode,
			"expiresAt": inv.ExpiresAt, "maxPlayers": created.MaxPlayers,
		})
		return
	}
	if len(room.Players) >= room.MaxPlayers {
		h.Lobby.freeGhostSeatsLocked(room)
	}
	if len(room.Players) >= room.MaxPlayers {
		h.Lobby.mu.Unlock()
		h.rejectCode(p, TypeJoinAsk, "full", "Partai penuh. Buat permainan sendiri.")
		return
	}
	if h.Lobby.player(room, p.ID) != nil {
		h.Lobby.mu.Unlock()
		h.rejectCode(p, TypeJoinAsk, "already", "Kamu sudah di partai ini.")
		return
	}
	askID := "ja-" + shortID()
	ask := joinAsk{
		ID: askID, FromID: p.ID, FromName: p.Name, HostID: room.HostID,
		RoomCode: room.RoomCode, ExpiresAt: time.Now().UnixMilli() + 45_000,
	}
	h.Lobby.joinAsks[askID] = ask
	hostID := room.HostID
	code := room.RoomCode
	seats, maxP := len(room.Players), room.MaxPlayers
	h.Lobby.mu.Unlock()
	h.pushUser(hostID, TypeJoinAskEv, map[string]any{
		"askId": askID, "userId": p.ID, "username": p.Name,
		"roomCode": code, "seats": seats, "maxPlayers": maxP,
		"expiresAt": ask.ExpiresAt,
	})
	select {
	case p.send <- marshal(TypeSocialPush, map[string]any{
		"type": "JOIN_ASK_SENT", "title": "Permintaan terkirim",
		"message": "Menunggu persetujuan host partai.",
	}):
	default:
	}
}

func (h *Hub) handleJoinAskRespond(host *Player, askID string, accept bool) {
	if host == nil || askID == "" {
		return
	}
	h.Lobby.mu.Lock()
	ask, ok := h.Lobby.joinAsks[askID]
	if !ok {
		h.Lobby.mu.Unlock()
		h.rejectCode(host, TypeJoinAskRespond, "missing", "Permintaan tidak ditemukan.")
		return
	}
	delete(h.Lobby.joinAsks, askID)
	if ask.HostID != host.ID {
		h.Lobby.mu.Unlock()
		h.rejectCode(host, TypeJoinAskRespond, "host", "Hanya host yang dapat menyetujui.")
		return
	}
	if time.Now().UnixMilli() > ask.ExpiresAt {
		h.Lobby.mu.Unlock()
		h.pushUser(ask.FromID, TypeSocialPush, map[string]any{
			"type": "JOIN_ASK_EXPIRED", "title": "Kedaluwarsa", "message": "Permintaan bergabung kedaluwarsa.",
		})
		return
	}
	code := ask.RoomCode
	fromID := ask.FromID
	fromName := ask.FromName
	h.Lobby.mu.Unlock()

	if !accept {
		h.pushUser(fromID, TypeInviteResult, map[string]any{"askId": askID, "accepted": false, "username": host.Name})
		h.pushUser(fromID, TypeSocialPush, map[string]any{
			"type": "JOIN_ASK_DECLINED", "title": "Ditolak",
			"message": host.Name + " menolak permintaan bergabung. Buat permainan sendiri.",
		})
		return
	}
	asker := (*Player)(nil)
	h.Lobby.mu.Lock()
	asker = h.Lobby.online[fromID]
	h.Lobby.mu.Unlock()
	if asker == nil {
		h.rejectCode(host, TypeJoinAskRespond, "offline", "Pemain sudah offline.")
		return
	}
	room, errc := h.Lobby.Join(asker, code)
	if errc != "" {
		h.pushUser(fromID, TypeSocialPush, map[string]any{
			"type": "JOIN_ASK_FAIL", "title": "Gagal bergabung", "message": friendlyUlar(errc),
		})
		h.rejectCode(host, TypeJoinAskRespond, errc, friendlyUlar(errc))
		return
	}
	h.emitRoom(host, room)
	h.emitRoom(asker, room)
	h.pushUser(fromID, TypeInviteResult, map[string]any{"askId": askID, "accepted": true, "username": host.Name})
	h.broadcastOnlineList()
	_ = fromName
}

func (h *Hub) sendQuestionHistory(p *Player) {
	if p == nil || h.Lobby == nil || h.Lobby.Attempts == nil {
		return
	}
	items := h.Lobby.Attempts.ForUser(p.ID)
	select {
	case p.send <- marshal(TypeQuestionHistory, map[string]any{"attempts": items, "stats": summarizeAttempts(items)}):
	default:
	}
}

func (h *Hub) sendRoom(p *Player, room *UlarRoom) {
	h.emitRoom(p, room)
}

func (h *Hub) rejectUlar(p *Player, action, reason string) {
	h.rejectCode(p, action, reason, friendlyUlar(reason))
}

func friendlyUlar(code string) string {
	switch code {
	case ErrNotYourTurn:
		return "Belum giliranmu."
	case ErrGameNotStarted:
		return "Permainan belum dimulai."
	case ErrRoomFull:
		return "Ruangan penuh."
	case ErrNotReady:
		return "Semua pemain online harus READY."
	case ErrAlreadyRolled:
		return "Dadunya sedang berjalan."
	case ErrNotInRoom:
		return "Kamu belum di ruangan."
	case ErrNeedPlayers:
		return "Minimal 2 pemain online."
	case ErrNotHost:
		return "Hanya host yang dapat memulai. Jika host putus, tunggu reconnect atau buat room baru."
	case ErrRateLimited:
		return "Terlalu banyak percobaan. Tunggu sebentar."
	case ErrLocked:
		return "Tunggu animasi selesai."
	case ErrNoQuestion:
		return "Tidak ada soal aktif."
	case ErrNotQuestionPlayer:
		return "Bukan giliranmu menjawab."
	case ErrAlreadyAnswered:
		return "Jawaban sudah terkirim."
	case ErrLateAnswer:
		return "Waktu sudah habis."
	default:
		return "Terjadi kesalahan. Silakan coba lagi."
	}
}
