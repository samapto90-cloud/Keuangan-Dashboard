package mmo

import (
	"sync"
	"time"
)

type UlarQueueEntry struct {
	UserID        string
	Username      string
	Mode          string
	MMR           int
	Region        string
	Grade         string
	PreferredSize int
	EnqueuedAt    time.Time
}

type PendingMatch struct {
	ID       string
	Mode     string
	UserIDs  []string
	Ready    map[string]bool
	FoundAt  time.Time
	RoomCode string
	RoomID   string
}

type Matchmaker struct {
	mu      sync.Mutex
	casual  []UlarQueueEntry
	ranked  []UlarQueueEntry
	pending map[string]*PendingMatch
	inQueue map[string]string
}

func NewMatchmaker() *Matchmaker {
	return &Matchmaker{pending: map[string]*PendingMatch{}, inQueue: map[string]string{}}
}

func (m *Matchmaker) Enqueue(e UlarQueueEntry) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.inQueue[e.UserID] != "" {
		return ""
	}
	e.EnqueuedAt = time.Now()
	if e.Mode == "RANKED" {
		m.ranked = append(m.ranked, e)
	} else {
		e.Mode = "CASUAL"
		m.casual = append(m.casual, e)
	}
	m.inQueue[e.UserID] = e.Mode
	return ""
}

func (m *Matchmaker) Cancel(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeLocked(userID)
}

func (m *Matchmaker) Queued(userID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.inQueue[userID]
}

func (m *Matchmaker) PendingFor(userID string) *PendingMatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.pending {
		for _, id := range p.UserIDs {
			if id == userID {
				cp := *p
				return &cp
			}
		}
	}
	return nil
}

func (m *Matchmaker) removeLocked(userID string) {
	delete(m.inQueue, userID)
	m.casual = dropQueue(m.casual, userID)
	m.ranked = dropQueue(m.ranked, userID)
}

func dropQueue(in []UlarQueueEntry, userID string) []UlarQueueEntry {
	out := in[:0]
	for _, e := range in {
		if e.UserID != userID {
			out = append(out, e)
		}
	}
	return out
}

func (m *Matchmaker) MarkReady(userID string, ok bool) *PendingMatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.pending {
		for _, id := range p.UserIDs {
			if id != userID {
				continue
			}
			if p.Ready == nil {
				p.Ready = map[string]bool{}
			}
			p.Ready[userID] = ok
			cp := *p
			return &cp
		}
	}
	return nil
}

func (h *Hub) queueJoin(p *Player, mode, region, grade string, preferredSize int) string {
	if h.Matchmaker == nil || p == nil {
		return ErrInvalidRequest
	}
	if h.Ops != nil {
		if ban := h.Ops.ActiveBan(p.ID); ban != nil {
			return "akun diblokir"
		}
	}
	if mode == "RANKED" && !LiveFlags().EnableRanked {
		return "ranked nonaktif"
	}
	if h.Lobby != nil {
		h.Lobby.mu.Lock()
		in := h.Lobby.inRoom[p.ID]
		h.Lobby.mu.Unlock()
		if in != "" {
			return "sudah di room"
		}
	}
	if h.Social != nil {
		if until := h.restrictUntil(p.ID); until > time.Now().UnixMilli() {
			return "matchmaking dibatasi sementara"
		}
	}
	mmr := 0
	if mode == "RANKED" {
		if h.Progress == nil {
			return "progres tidak siap"
		}
		pr := h.Progress.Ensure(p.ID, p.Name)
		if pr.Level < RankedMinLevel {
			return "butuh level 5 untuk ranked"
		}
		st := pr.RankState()
		mmr = st.Index
	}
	if region == "" {
		region = "ID-JKT"
	}
	size := preferredSize
	if size < 2 {
		size = 2
	}
	if size > 4 {
		size = 4
	}
	if h.Social != nil {
		h.Social.Touch(p.ID, false)
	}
	h.Matchmaker.Enqueue(UlarQueueEntry{
		UserID: p.ID, Username: p.Name, Mode: mode, MMR: mmr, Region: region,
		Grade: normalizeGrade(grade), PreferredSize: size,
	})
	return ""
}

func (h *Hub) restrictUntil(userID string) int64 {
	if h.Progress == nil {
		return 0
	}
	p, ok := h.Progress.Get(userID)
	if !ok {
		return 0
	}
	return p.RestrictUntil
}

func (h *Hub) tickMatchmaking() {
	if h.Matchmaker == nil || h.Lobby == nil {
		return
	}
	now := time.Now()
	h.Matchmaker.ExpireReady(h, now)
	h.Matchmaker.form(h, "CASUAL", now)
	h.Matchmaker.form(h, "RANKED", now)
	if h.Social != nil {
		h.Social.ExpireInvites(now.UnixMilli())
	}
}

func (m *Matchmaker) form(h *Hub, mode string, now time.Time) {
	m.mu.Lock()
	list := m.casual
	if mode == "RANKED" {
		list = m.ranked
	}
	if len(list) < 2 {
		m.mu.Unlock()
		return
	}
	used := map[string]bool{}
	var groups [][]UlarQueueEntry
	for i, seed := range list {
		if used[seed.UserID] {
			continue
		}
		wait := int(now.Sub(seed.EnqueuedAt).Seconds())
		need := seed.PreferredSize
		if need < 2 {
			need = 2
		}
		if need > 4 {
			need = 4
		}
		// Setelah 15 detik tunggu, longgarkan ke 2 pemain agar online tidak macet.
		if wait >= 15 && need > 2 {
			need = 2
		}
		g := []UlarQueueEntry{seed}
		for j := i + 1; j < len(list) && len(g) < need; j++ {
			o := list[j]
			if used[o.UserID] {
				continue
			}
			if mode == "RANKED" {
				wa := int(now.Sub(o.EnqueuedAt).Seconds())
				if wa > wait {
					wait = wa
				}
				sa := RankFromIndexSafe(seed.MMR)
				sb := RankFromIndexSafe(o.MMR)
				if !RanksCompatible(sa, sb, wait) {
					continue
				}
			}
			if normalizeGrade(o.Grade) != normalizeGrade(seed.Grade) {
				continue
			}
			// Hormati preferred size lawan: jangan masukkan ke grup lebih besar dari yang mereka mau,
			// kecuali mereka juga sudah menunggu lama.
			oNeed := o.PreferredSize
			if oNeed < 2 {
				oNeed = 2
			}
			oWait := int(now.Sub(o.EnqueuedAt).Seconds())
			if oWait >= 15 && oNeed > 2 {
				oNeed = 2
			}
			if len(g)+1 > oNeed && oNeed < need {
				continue
			}
			g = append(g, o)
		}
		if len(g) >= need {
			groups = append(groups, g)
			for _, e := range g {
				used[e.UserID] = true
			}
		}
	}
	m.mu.Unlock()
	for _, g := range groups {
		h.openMatchmade(g, mode)
	}
}

func RankFromIndexSafe(idx int) RankState {
	t, d, rr := RankFromIndex(idx)
	return RankState{Tier: t, Division: d, RR: rr, Index: idx}.Sync()
}

func (h *Hub) openMatchmade(group []UlarQueueEntry, mode string) {
	players := make([]*Player, 0, len(group))
	h.Matchmaker.mu.Lock()
	for _, e := range group {
		h.Matchmaker.removeLocked(e.UserID)
		if h.Lobby != nil {
			if p := h.Lobby.online[e.UserID]; p != nil {
				players = append(players, p)
			}
		}
	}
	h.Matchmaker.mu.Unlock()
	if len(players) < 2 {
		for _, e := range group {
			h.Matchmaker.Enqueue(e)
		}
		return
	}
	room, errc := h.Lobby.CreateFilled(players, mode)
	if errc != "" || room == nil {
		for _, e := range group {
			h.Matchmaker.Enqueue(e)
		}
		return
	}
	if len(group) > 0 {
		h.Lobby.mu.Lock()
		room.Grade = normalizeGrade(group[0].Grade)
		h.Lobby.mu.Unlock()
	}
	ids := make([]string, 0, len(players))
	for _, p := range players {
		ids = append(ids, p.ID)
		if h.Social != nil {
			h.Social.Notify(p.ID, NoteMatchFound, "MATCH FOUND!", "Siap-siap bertanding.", room.ID)
			h.Social.Touch(p.ID, false)
		}
	}
	pm := &PendingMatch{
		ID: "pm-" + shortID(), Mode: mode, UserIDs: ids, Ready: map[string]bool{},
		FoundAt: time.Now(), RoomCode: room.RoomCode, RoomID: room.ID,
	}
	h.Matchmaker.mu.Lock()
	h.Matchmaker.pending[pm.ID] = pm
	h.Matchmaker.mu.Unlock()
	h.Lobby.mu.Lock()
	snap := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, ""))
	h.Lobby.mu.Unlock()
	h.pushRoom(room, TypeRoomUpdated, snap)
	h.pushRoom(room, TypeMatchFound, map[string]any{
		"matchId": pm.ID, "roomCode": room.RoomCode, "mode": mode,
		"countdown": MatchFoundSec, "playerIds": ids,
	})
}

func (m *Matchmaker) ExpireReady(h *Hub, now time.Time) {
	m.mu.Lock()
	var due []*PendingMatch
	for _, p := range m.pending {
		if now.Sub(p.FoundAt) >= time.Duration(MatchFoundSec+ReadyCheckSec)*time.Second {
			due = append(due, p)
		}
	}
	m.mu.Unlock()
	for _, p := range due {
		h.resolvePending(p)
	}
}

func (h *Hub) resolvePending(p *PendingMatch) {
	h.Matchmaker.mu.Lock()
	cur := h.Matchmaker.pending[p.ID]
	if cur == nil {
		h.Matchmaker.mu.Unlock()
		return
	}
	readyIDs := make([]string, 0)
	for _, id := range cur.UserIDs {
		if cur.Ready[id] {
			readyIDs = append(readyIDs, id)
		}
	}
	delete(h.Matchmaker.pending, p.ID)
	h.Matchmaker.mu.Unlock()
	if len(readyIDs) >= 2 {
		h.Lobby.mu.Lock()
		room := h.Lobby.rooms[p.RoomID]
		if room != nil {
			for _, id := range cur.UserIDs {
				if !cur.Ready[id] {
					h.Lobby.leaveLocked(id, room.ID)
				} else if pl := h.Lobby.player(room, id); pl != nil {
					pl.IsReady = true
					pl.PlayState = "READY"
				}
			}
		}
		h.Lobby.mu.Unlock()
		if room != nil && room.HostID != "" {
			_ = h.startMatch(room.HostID)
		}
		return
	}
	if h.Lobby != nil {
		h.Lobby.mu.Lock()
		grade := ""
		if room := h.Lobby.rooms[p.RoomID]; room != nil {
			grade = room.Grade
			if room.Match == nil {
				h.Lobby.closeLocked(room)
			}
		}
		h.Lobby.mu.Unlock()
		for _, id := range readyIDs {
			if pl := h.Lobby.online[id]; pl != nil {
				h.Matchmaker.Enqueue(UlarQueueEntry{
					UserID: id, Username: pl.Name, Mode: p.Mode,
					Grade: normalizeGrade(grade), PreferredSize: 2,
				})
			}
		}
		return
	}
	for _, id := range readyIDs {
		h.Matchmaker.Enqueue(UlarQueueEntry{UserID: id, Mode: p.Mode, PreferredSize: 2})
	}
}

func (h *Hub) matchReady(p *Player, ok bool) {
	pm := h.Matchmaker.MarkReady(p.ID, ok)
	if pm == nil {
		return
	}
	all := true
	for _, id := range pm.UserIDs {
		if !pm.Ready[id] {
			all = false
			break
		}
	}
	if all && len(pm.UserIDs) >= 2 {
		h.resolvePending(pm)
	}
}
