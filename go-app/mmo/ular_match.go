package mmo

import (
	"fmt"
	"log"
	"time"
)

var ularCountdown = 3 * time.Second

func (l *UlarLobby) snapshotLocked(room *UlarRoom, you string) GameSnapshot {
	pls := make([]UlarPlayer, 0, len(room.Players))
	for _, p := range room.Players {
		if p == nil {
			continue
		}
		cp := *p
		pls = append(pls, cp)
	}
	snap := GameSnapshot{
		RoomCode:   room.RoomCode,
		HostID:     room.HostID,
		Status:     room.Status,
		Players:    pls,
		YouAre:     you,
		Chat:       append([]UlarChatLine{}, room.Chat...),
		MaxPlayers: room.MaxPlayers,
		Mode:       room.Mode,
		Visibility: room.Visibility,
		Grade:      room.Grade,
	}
	if room.Match != nil {
		snap.MatchID = room.Match.ID
		snap.Phase = room.Match.Phase
		snap.CurrentPlayerID = room.Match.CurrentPlayerID
		snap.TurnNumber = room.Match.TurnNumber
		snap.LastAction = room.Match.LastAction
		snap.LastDice = room.Match.LastDice
		snap.WinnerID = room.Match.WinnerID
		snap.CountdownEnd = room.Match.CountdownEnd
		snap.PowerCells = clonePowerCells(room.Match.PowerCells)
		if room.Match.Grade != "" {
			snap.Grade = room.Match.Grade
		}
		if len(room.Match.Attempts) > 0 {
			snap.Stats = summarizeAttempts(room.Match.Attempts)
			ps := map[string]any{}
			for _, p := range room.Players {
				if p == nil {
					continue
				}
				own := make([]QuestionAttempt, 0)
				for _, a := range room.Match.Attempts {
					if a.UserID == p.UserID {
						own = append(own, a)
					}
				}
				ps[p.UserID] = summarizeAttempts(own)
			}
			snap.PlayerStats = ps
		}
		if room.Match.CurrentQuestionID != "" {
			snap.CurrentQuestionID = room.Match.CurrentQuestionID
			snap.QuestionPlayerID = room.Match.QuestionPlayerID
			snap.QuestionState = room.Match.QuestionState
			snap.AnswerSubmitted = room.Match.AnswerSubmitted
			snap.Penalty = room.Match.Penalty
			if !room.Match.QuestionStartedAt.IsZero() {
				snap.QuestionStartedAt = room.Match.QuestionStartedAt.UnixMilli()
				lim := questionTimeLimit
				if room.Match.QuestionLimit > 0 {
					lim = room.Match.QuestionLimit
				}
				snap.QuestionEndsAt = room.Match.QuestionStartedAt.Add(lim).UnixMilli()
			}
			view := room.Match.QuestionView
			if room.Match.QuestionState == QStateActive || room.Match.QuestionState == QStatePending {
				snap.Question = &view
			}
		}
	}
	return snap
}

func (l *UlarLobby) nextSeq(room *UlarRoom) (int, string, int64) {
	room.Seq++
	if room.SeenEvent == nil {
		room.SeenEvent = map[string]bool{}
	}
	id := "ev-" + shortID()
	room.SeenEvent[id] = true
	return room.Seq, id, time.Now().UnixMilli()
}

func (l *UlarLobby) wrap(room *UlarRoom, extra GameSnapshot) GameSnapshot {
	seq, id, at := l.nextSeq(room)
	extra.Seq = seq
	extra.EventID = id
	extra.At = at
	return extra
}

func (h *Hub) emitLocked(room *UlarRoom, typ string, payload any) {
	if h == nil || h.Lobby == nil || room == nil {
		return
	}
	socks := make([]*Player, 0, len(room.Players))
	for _, pl := range room.Players {
		if pl == nil || !pl.IsConnected {
			continue
		}
		if s := h.Lobby.online[pl.UserID]; s != nil {
			socks = append(socks, s)
		}
	}
	raw := marshal(typ, payload)
	for _, s := range socks {
		select {
		case s.send <- raw:
		default:
		}
	}
}

func (h *Hub) pushRoom(room *UlarRoom, typ string, payload any) {
	if h == nil || h.Lobby == nil || room == nil {
		return
	}
	h.Lobby.mu.Lock()
	socks := make([]*Player, 0, len(room.Players))
	for _, pl := range room.Players {
		if pl == nil || !pl.IsConnected {
			continue
		}
		if s := h.Lobby.online[pl.UserID]; s != nil {
			socks = append(socks, s)
		}
	}
	h.Lobby.mu.Unlock()
	raw := marshal(typ, payload)
	for _, s := range socks {
		select {
		case s.send <- raw:
		default:
		}
	}
}

func (h *Hub) pushUser(userID, typ string, payload any) {
	if h == nil || h.Lobby == nil || userID == "" {
		return
	}
	h.Lobby.mu.Lock()
	s := h.Lobby.online[userID]
	h.Lobby.mu.Unlock()
	if s == nil {
		return
	}
	select {
	case s.send <- marshal(typ, payload):
	default:
	}
}

func (h *Hub) rejectCode(p *Player, action, code, msg string) {
	if p == nil {
		return
	}
	select {
	case p.send <- marshal(TypeUlarError, UlarEnvelope{Code: code, Message: msg, At: time.Now().UnixMilli()}):
	default:
	}
	log.Printf("ular reject %s action=%s code=%s", p.ID, action, code)
}

func (l *UlarLobby) player(room *UlarRoom, userID string) *UlarPlayer {
	for _, p := range room.Players {
		if p != nil && p.UserID == userID {
			return p
		}
	}
	return nil
}

func (l *UlarLobby) connectedCount(room *UlarRoom) int {
	n := 0
	for _, p := range room.Players {
		if p != nil && p.IsConnected {
			n++
		}
	}
	return n
}

func (l *UlarLobby) allReady(room *UlarRoom) bool {
	if len(room.Players) < 2 {
		return false
	}
	for _, p := range room.Players {
		if p == nil || !p.IsReady {
			return false
		}
	}
	return true
}

func (l *UlarLobby) nextAlive(room *UlarRoom, currentID string) string {
	if len(room.Players) == 0 {
		return ""
	}
	idx := 0
	for i, p := range room.Players {
		if p != nil && p.UserID == currentID {
			idx = i
			break
		}
	}
	for n := 1; n <= len(room.Players); n++ {
		p := room.Players[(idx+n)%len(room.Players)]
		if p != nil && p.IsConnected && p.PlayState != "FINISHED" {
			return p.UserID
		}
	}
	return ""
}

func (h *Hub) startMatch(hostID string) string {
	l := h.Lobby
	l.mu.Lock()
	rid := l.inRoom[hostID]
	room := l.rooms[rid]
	if room == nil {
		l.mu.Unlock()
		return ErrNotInRoom
	}
	if room.HostID != hostID {
		l.mu.Unlock()
		return ErrNotHost
	}
	if len(room.Players) < 2 {
		l.mu.Unlock()
		return ErrNeedPlayers
	}
	if !l.allReady(room) {
		l.mu.Unlock()
		return ErrNotReady
	}
	if room.Status != UlarWaiting && room.Status != UlarReady {
		l.mu.Unlock()
		return ErrInvalidRequest
	}
	now := time.Now()
	room.Status = UlarStarting
	for _, p := range room.Players {
		p.Position = OFFBOARD_START
		p.Items = map[string]int{}
		p.PlayState = "PLAYING"
		p.IsReady = true
	}
	cfg := LiveConfig()
	achByID := map[string]UlarAchievement{}
	for _, a := range liveAchievements() {
		if a.Active {
			achByID[a.ID] = a
		}
	}
	m := &UlarLiveMatch{
		ID:            "mt-" + shortID(),
		RoomID:        room.ID,
		Status:        UlarStarting,
		Phase:         UlarStarting,
		TurnNumber:    0,
		CreatedAt:     now,
		CountdownEnd:  now.Add(ularCountdown).UnixMilli(),
		LastAction:    "GAME_STARTING",
		LastActionAt:  now,
		SubjectCursor: int(now.UnixNano() % 4),
		SnakeHits:     map[string]int{},
		LadderHits:    map[string]int{},
		QuestionLimit: LiveQuestionTime(),
		PenaltyN:      LivePenaltyN(),
		XPCorrectAmt:        liveReward(cfg.XPCorrect, XP_CORRECT_ANSWER),
		XPWrongAmt:          liveReward(cfg.XPWrong, XP_WRONG_ANSWER),
		XPTimeoutAmt:        liveReward(cfg.XPTimeout, XP_TIMEOUT),
		XPMatchCompleteAmt: liveReward(cfg.XPMatchComplete, XP_MATCH_COMPLETE),
		XPWinAmt:            liveReward(cfg.XPWin, XP_WIN),
		CoinMatchAmt:        liveReward(cfg.CoinMatch, COIN_MATCH),
		CoinWinAmt:          liveReward(cfg.CoinWin, COIN_WIN),
		RankWinRrAmt:       LiveRankWin(),
		RankLossRrAmt:      LiveRankLoss(),
		AchievementsByID:   achByID,
		PowerCells:         SpawnPowerCells(DefaultSnakes, DefaultLadders),
		Grade:              normalizeGrade(room.Grade),
	}
	room.Match = m
	l.matches[m.ID] = &UlarMatch{ID: m.ID, RoomID: room.ID, Status: UlarStarting, CreatedAt: now}
	log.Printf("GAME_STARTED room=%s match=%s", room.RoomCode, m.ID)
	snap := l.wrap(room, l.snapshotLocked(room, ""))
	code := room.RoomCode
	matchID := m.ID
	l.mu.Unlock()
	h.pushRoom(room, TypeRoomStart, snap)
	go func() {
		time.Sleep(ularCountdown)
		l.mu.Lock()
		rm := l.rooms[l.byCode[code]]
		if rm == nil || rm.Match == nil || rm.Match.ID != matchID {
			l.mu.Unlock()
			return
		}
		rm.Status = UlarPlaying
		rm.Match.Status = UlarPlaying
		rm.Match.Phase = UlarPlayerTurn
		rm.Match.StartedAt = time.Now()
		rm.Match.TurnNumber = 1
		rm.Match.CurrentPlayerID = rm.Players[0].UserID
		rm.Match.LastAction = "TURN"
		snap := l.wrap(rm, l.snapshotLocked(rm, ""))
		cur := rm.Match.CurrentPlayerID
		l.mu.Unlock()
		h.pushRoom(rm, TypeGameState, snap)
		h.pushRoom(rm, TypeGameTurn, snap)
		log.Printf("TURN_CHANGED room=%s player=%s", code, cur)
	}()
	return ""
}

func (h *Hub) rollMatch(p *Player) string {
	l := h.Lobby
	l.mu.Lock()
	rid := l.inRoom[p.ID]
	room := l.rooms[rid]
	if room == nil || room.Match == nil {
		l.mu.Unlock()
		return ErrNotInRoom
	}
	if room.Status != UlarPlaying {
		l.mu.Unlock()
		return ErrGameNotStarted
	}
	m := room.Match
	if m.Phase != UlarPlayerTurn {
		l.mu.Unlock()
		if m.Phase == UlarRolling || m.Phase == UlarMoving || m.Phase == UlarSnake || m.Phase == UlarLadder {
			return ErrAlreadyRolled
		}
		return ErrLocked
	}
	if m.CurrentPlayerID != p.ID {
		l.mu.Unlock()
		return ErrNotYourTurn
	}
	pl := l.player(room, p.ID)
	if pl == nil || !pl.IsConnected {
		l.mu.Unlock()
		return ErrNotInRoom
	}
	dice, err := RollDiceSecure()
	if err != nil {
		l.mu.Unlock()
		return ErrInvalidRequest
	}
	cfg := DefaultBoardConfig()
	move := ResolveMove(cfg, pl.Position, dice)
	if m.SnakeHits == nil {
		m.SnakeHits = map[string]int{}
	}
	if m.LadderHits == nil {
		m.LadderHits = map[string]int{}
	}
	if move.SnakeTo != 0 {
		m.SnakeHits[p.ID]++
	}
	if move.LadderTo != 0 {
		m.LadderHits[p.ID]++
	}
	m.Phase = UlarRolling
	m.LastDice = dice
	m.LastAction = "ROLL"
	m.WaitingAnim = true
	m.PendingMove = move
	m.LastActionAt = time.Now()
	seq, evid, at := l.nextSeq(room)
	env := UlarEnvelope{Seq: seq, EventID: evid, At: at}
	diceMsg := struct {
		UlarEnvelope
		UserID   string `json:"userId"`
		Username string `json:"username"`
		Dice     int    `json:"dice"`
	}{UlarEnvelope: env, UserID: p.ID, Username: p.Name, Dice: dice}
	moveMsg := MoveBroadcast{UlarEnvelope: env, UserID: p.ID, Username: p.Name, Move: move}
	code := room.RoomCode
	matchID := m.ID
	l.mu.Unlock()
	log.Printf("ROLL_REQUEST %s dice=%d before=%d", p.ID, dice, pl.Position)
	h.pushRoom(room, TypeDiceResult, diceMsg)
	h.pushRoom(room, TypePlayerMoving, moveMsg)
	if move.SnakeTo != 0 {
		h.pushRoom(room, TypeSnakeTrigger, moveMsg)
		log.Printf("SNAKE_TRIGGERED %s %d->%d", p.ID, move.SnakeFrom, move.SnakeTo)
	} else if move.LadderTo != 0 {
		h.pushRoom(room, TypeLadderTrigger, moveMsg)
		log.Printf("LADDER_TRIGGERED %s %d->%d", p.ID, move.LadderFrom, move.LadderTo)
	}
	wait := time.Duration(len(move.WalkPath))*time.Millisecond*MOVE_DURATION + 400*time.Millisecond
	if move.SnakeTo != 0 || move.LadderTo != 0 {
		wait += time.Duration(SNAKE_DURATION) * time.Millisecond
	}
	go func() {
		time.Sleep(wait)
		h.finishRoll(code, matchID, p.ID, move)
	}()
	return ""
}

func (h *Hub) finishRoll(roomCode, matchID, userID string, move MoveResult) {
	l := h.Lobby
	l.mu.Lock()
	room := l.rooms[l.byCode[roomCode]]
	if room == nil || room.Match == nil || room.Match.ID != matchID {
		l.mu.Unlock()
		return
	}
	if !room.Match.WaitingAnim {
		l.mu.Unlock()
		return
	}
	pl := l.player(room, userID)
	if pl != nil {
		pl.Position = move.Final
		if kind, ok := h.tryPickupPowerLocked(room, pl); ok {
			seq, evid, at := l.nextSeq(room)
			h.emitLocked(room, TypePowerPickup, PowerPickupOut{
				UlarEnvelope: UlarEnvelope{Seq: seq, EventID: evid, At: at},
				UserID:       pl.UserID,
				Username:     pl.Username,
				Item:         kind,
				Position:     pl.Position,
				Items:        cloneItemMap(pl.Items),
				PowerCells:   clonePowerCells(room.Match.PowerCells),
				Message:      fmt.Sprintf("%s mengambil item di kotak %d!", pl.Username, pl.Position),
			})
		}
	}
	room.Match.WaitingAnim = false
	h.startQuestionLocked(room, userID, move.Reached100)
	l.mu.Unlock()
}

func (h *Hub) persistMatchLocked(room *UlarRoom) {
	if room.Match == nil || h.Lobby.Store == nil {
		return
	}
	players := make([]StoredMatchPlayer, 0, len(room.Players))
	order := 1
	if room.Match.WinnerID != "" {
		for _, p := range room.Players {
			if p != nil && p.UserID == room.Match.WinnerID {
				players = append(players, StoredMatchPlayer{UserID: p.UserID, Username: p.Username, Color: p.Color, FinalPosition: p.Position, FinishOrder: 1})
			}
		}
		order = 2
	}
	rest := append([]*UlarPlayer{}, room.Players...)
	for i := 0; i < len(rest); i++ {
		for j := i + 1; j < len(rest); j++ {
			if rest[j].Position > rest[i].Position {
				rest[i], rest[j] = rest[j], rest[i]
			}
		}
	}
	for _, p := range rest {
		if p == nil || p.UserID == room.Match.WinnerID {
			continue
		}
		players = append(players, StoredMatchPlayer{UserID: p.UserID, Username: p.Username, Color: p.Color, FinalPosition: p.Position, FinishOrder: order})
		order++
	}
	fin := int64(0)
	if room.Match.FinishedAt != nil {
		fin = room.Match.FinishedAt.UnixMilli()
	}
	h.Lobby.Store.Append(StoredMatch{
		ID: room.Match.ID, RoomID: room.ID, RoomCode: room.RoomCode, Status: string(room.Match.Status),
		Mode: room.Mode, WinnerID: room.Match.WinnerID, StartedAt: room.Match.StartedAt.UnixMilli(), FinishedAt: fin, Players: players,
	})
}

func (h *Hub) settleMatchRewardsLocked(room *UlarRoom) []RewardEvent {
	if h.Progress == nil || room == nil || room.Match == nil {
		return nil
	}
	m := room.Match
	type tally struct{ c, w, t int }
	stats := map[string]tally{}
	for _, a := range m.Attempts {
		s := stats[a.UserID]
		if a.Timeout {
			s.t++
		} else if a.Correct {
			s.c++
		} else {
			s.w++
		}
		stats[a.UserID] = s
	}
	ranked := append([]*UlarPlayer{}, room.Players...)
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			ai, aj := ranked[i], ranked[j]
			if ai == nil || aj == nil {
				continue
			}
			aWin := ai.UserID == m.WinnerID
			bWin := aj.UserID == m.WinnerID
			if bWin && !aWin {
				ranked[i], ranked[j] = ranked[j], ranked[i]
				continue
			}
			if aWin == bWin && aj.Position > ai.Position {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}
	in := MatchSettlement{
		MatchID:        m.ID,
		Mode:           room.Mode,
		XPMatchComplete: m.XPMatchCompleteAmt,
		CoinMatch:       m.CoinMatchAmt,
		XPWin:           m.XPWinAmt,
		CoinWin:         m.CoinWinAmt,
		RankWinRR:       m.RankWinRrAmt,
		RankLossRR:      m.RankLossRrAmt,
		AchievementsByID: m.AchievementsByID,
	}
	rank := 1
	for _, p := range ranked {
		if p == nil {
			continue
		}
		st := stats[p.UserID]
		snakes, ladders := 0, 0
		if m.SnakeHits != nil {
			snakes = m.SnakeHits[p.UserID]
		}
		if m.LadderHits != nil {
			ladders = m.LadderHits[p.UserID]
		}
		in.Players = append(in.Players, MatchSettlePlayer{
			UserID: p.UserID, Username: p.Username, Rank: rank, Position: p.Position,
			Correct: st.c, Wrong: st.w, Timeout: st.t, Won: p.UserID == m.WinnerID,
			Ladders: ladders, Snakes: snakes, Reached100: p.Position >= MAX_POSITION,
		})
		rank++
	}
	events, _ := h.Progress.SettleMatch(in)
	return events
}

func (h *Hub) animDone(p *Player) {
	l := h.Lobby
	l.mu.Lock()
	room := l.rooms[l.inRoom[p.ID]]
	if room == nil || room.Match == nil || !room.Match.WaitingAnim || room.Match.CurrentPlayerID != p.ID {
		l.mu.Unlock()
		return
	}
	move := room.Match.PendingMove
	code := room.RoomCode
	matchID := room.Match.ID
	l.mu.Unlock()
	h.finishRoll(code, matchID, p.ID, move)
}
