package mmo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	ReqPending   = "PENDING"
	ReqAccepted  = "ACCEPTED"
	ReqRejected  = "REJECTED"
	ReqCancelled = "CANCELLED"
	InvExpired   = "EXPIRED"

	NoteFriendReq  = "FRIEND_REQUEST"
	NoteFriendAcc  = "FRIEND_ACCEPTED"
	NoteGameInvite = "GAME_INVITE"
	NoteMatchFound = "MATCH_FOUND"
	NoteAchieve    = "ACHIEVEMENT"
	NoteRankUp     = "RANK_UP"
	NoteRankDown   = "RANK_DOWN"

	PresOnline  = "ONLINE"
	PresInGame  = "IN_GAME"
	PresAway    = "AWAY"
	PresOffline = "OFFLINE"
)

type FriendRequest struct {
	ID         string `json:"id"`
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	Status     string `json:"status"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

type Friendship struct {
	ID        string `json:"id"`
	UserA     string `json:"userA"`
	UserB     string `json:"userB"`
	CreatedAt int64  `json:"createdAt"`
}

type UserBlock struct {
	BlockerID string `json:"blockerId"`
	BlockedID string `json:"blockedId"`
	CreatedAt int64  `json:"createdAt"`
}

type GameInvite struct {
	ID         string `json:"id"`
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	RoomID     string `json:"roomId"`
	RoomCode   string `json:"roomCode"`
	Status     string `json:"status"`
	ExpiresAt  int64  `json:"expiresAt"`
	CreatedAt  int64  `json:"createdAt"`
}

type UlarNote struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	RefID     string `json:"refId,omitempty"`
	Read      bool   `json:"read"`
	CreatedAt int64  `json:"createdAt"`
}

type PlayerReport struct {
	ID             string `json:"id"`
	ReporterID     string `json:"reporterId"`
	ReportedID     string `json:"reportedId"`
	Reason         string `json:"reason"`
	Description    string `json:"description,omitempty"`
	MatchID        string `json:"matchId,omitempty"`
	Status         string `json:"status"`
	ResolutionNote string `json:"resolutionNote,omitempty"`
	ResolvedBy     string `json:"resolvedBy,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
}

type SocialPrivacy struct {
	AllowFriendRequests bool `json:"allowFriendRequests"`
	AllowGameInvites    bool `json:"allowGameInvites"`
	ShowOnlineStatus    bool `json:"showOnlineStatus"`
}

type Presence struct {
	UserID   string `json:"userId"`
	LastSeen int64  `json:"lastSeen"`
	InGame   bool   `json:"inGame"`
}

type socialBlob struct {
	Requests []FriendRequest          `json:"requests"`
	Friends  []Friendship             `json:"friends"`
	Blocks   []UserBlock              `json:"blocks"`
	Invites  []GameInvite             `json:"invites"`
	Notes    []UlarNote               `json:"notes"`
	Reports  []PlayerReport           `json:"reports"`
	Privacy  map[string]SocialPrivacy `json:"privacy"`
}

type SocialStore struct {
	mu       sync.Mutex
	path     string
	blob     socialBlob
	presence map[string]Presence
}

func socialStorePath() string {
	if p := strings.TrimSpace(os.Getenv("ULAR_SOCIAL_STORE")); p != "" {
		return p
	}
	return filepath.Join("data", "ular-social.json")
}

func OpenSocialStore(path string) *SocialStore {
	s := &SocialStore{path: path, blob: socialBlob{Privacy: map[string]SocialPrivacy{}}, presence: map[string]Presence{}}
	raw, err := os.ReadFile(path)
	if err == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &s.blob)
	}
	if s.blob.Privacy == nil {
		s.blob.Privacy = map[string]SocialPrivacy{}
	}
	return s
}

func (s *SocialStore) flushLocked() error {
	raw, err := json.MarshalIndent(s.blob, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
	return os.WriteFile(s.path, raw, 0o644)
}

func pairKey(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

func defaultPrivacy() SocialPrivacy {
	return SocialPrivacy{AllowFriendRequests: true, AllowGameInvites: true, ShowOnlineStatus: true}
}

func (s *SocialStore) privacyLocked(userID string) SocialPrivacy {
	p, ok := s.blob.Privacy[userID]
	if !ok {
		return defaultPrivacy()
	}
	// Entri JSON lama sering tersimpan semua-false (zero value) — anggap default.
	if !p.AllowFriendRequests && !p.AllowGameInvites && !p.ShowOnlineStatus {
		return defaultPrivacy()
	}
	return p
}

func (s *SocialStore) SetPrivacy(userID string, p SocialPrivacy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blob.Privacy[userID] = p
	_ = s.flushLocked()
}

func (s *SocialStore) Privacy(userID string) SocialPrivacy {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.privacyLocked(userID)
}

func (s *SocialStore) Touch(userID string, inGame bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.presence[userID] = Presence{UserID: userID, LastSeen: time.Now().UnixMilli(), InGame: inGame}
}

func (s *SocialStore) SetInGame(userID string, inGame bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.presence[userID]
	p.UserID = userID
	p.InGame = inGame
	if p.LastSeen == 0 {
		p.LastSeen = time.Now().UnixMilli()
	}
	s.presence[userID] = p
}

func (s *SocialStore) Status(userID string, viewer string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	priv := s.privacyLocked(userID)
	if !priv.ShowOnlineStatus && viewer != userID {
		return PresOffline
	}
	p := s.presence[userID]
	if p.LastSeen == 0 {
		return PresOffline
	}
	age := time.Now().UnixMilli() - p.LastSeen
	if p.InGame && age < int64(PresenceAway)*1000 {
		return PresInGame
	}
	if age <= int64(PresenceOnline)*1000 {
		return PresOnline
	}
	if age <= int64(PresenceAway)*1000 {
		return PresAway
	}
	return PresOffline
}

func (s *SocialStore) Blocked(a, b string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.blockedLocked(a, b)
}

func (s *SocialStore) blockedLocked(a, b string) bool {
	for _, x := range s.blob.Blocks {
		if (x.BlockerID == a && x.BlockedID == b) || (x.BlockerID == b && x.BlockedID == a) {
			return true
		}
	}
	return false
}

func (s *SocialStore) AreFriends(a, b string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.friendsLocked(a, b)
}

func (s *SocialStore) friendsLocked(a, b string) bool {
	aa, bb := pairKey(a, b)
	for _, f := range s.blob.Friends {
		if f.UserA == aa && f.UserB == bb {
			return true
		}
	}
	return false
}

func (s *SocialStore) FriendIDs(userID string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0)
	for _, f := range s.blob.Friends {
		if f.UserA == userID {
			out = append(out, f.UserB)
		} else if f.UserB == userID {
			out = append(out, f.UserA)
		}
	}
	return out
}

func (s *SocialStore) noteLocked(userID, typ, title, body, ref string) {
	s.blob.Notes = append(s.blob.Notes, UlarNote{
		ID: "nt-" + shortID(), UserID: userID, Type: typ, Title: title, Body: body, RefID: ref, CreatedAt: time.Now().UnixMilli(),
	})
}

func (s *SocialStore) RequestFriend(from, to string) (FriendRequest, string) {
	if from == "" || to == "" || from == to {
		return FriendRequest{}, "tidak valid"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blockedLocked(from, to) {
		return FriendRequest{}, "diblokir"
	}
	priv := s.privacyLocked(to)
	if !priv.AllowFriendRequests {
		return FriendRequest{}, "tidak menerima permintaan"
	}
	if s.friendsLocked(from, to) {
		return FriendRequest{}, "sudah berteman"
	}
	now := time.Now().UnixMilli()
	for i, r := range s.blob.Requests {
		if r.Status != ReqPending {
			continue
		}
		if r.SenderID == from && r.ReceiverID == to {
			return r, ""
		}
		if r.SenderID == to && r.ReceiverID == from {
			// auto-accept reverse pending
			s.blob.Requests[i].Status = ReqAccepted
			s.blob.Requests[i].UpdatedAt = now
			s.addFriendLocked(from, to, now)
			s.noteLocked(from, NoteFriendAcc, "Teman baru", "Permintaan diterima.", to)
			_ = s.flushLocked()
			return s.blob.Requests[i], ""
		}
	}
	req := FriendRequest{ID: "fr-" + shortID(), SenderID: from, ReceiverID: to, Status: ReqPending, CreatedAt: now, UpdatedAt: now}
	s.blob.Requests = append(s.blob.Requests, req)
	s.noteLocked(to, NoteFriendReq, "Permintaan teman", "Seseorang ingin berteman.", req.ID)
	_ = s.flushLocked()
	return req, ""
}

func (s *SocialStore) addFriendLocked(a, b string, now int64) {
	aa, bb := pairKey(a, b)
	if s.friendsLocked(a, b) {
		return
	}
	s.blob.Friends = append(s.blob.Friends, Friendship{ID: "fs-" + shortID(), UserA: aa, UserB: bb, CreatedAt: now})
}

func (s *SocialStore) RespondFriend(actor, requestID, action string) (FriendRequest, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	for i, r := range s.blob.Requests {
		if r.ID != requestID {
			continue
		}
		if r.Status != ReqPending {
			return r, "sudah diproses"
		}
		switch action {
		case "accept":
			if r.ReceiverID != actor {
				return FriendRequest{}, "bukan penerima"
			}
			s.blob.Requests[i].Status = ReqAccepted
			s.blob.Requests[i].UpdatedAt = now
			s.addFriendLocked(r.SenderID, r.ReceiverID, now)
			s.noteLocked(r.SenderID, NoteFriendAcc, "Teman baru", "Permintaan diterima.", actor)
		case "reject":
			if r.ReceiverID != actor {
				return FriendRequest{}, "bukan penerima"
			}
			s.blob.Requests[i].Status = ReqRejected
			s.blob.Requests[i].UpdatedAt = now
		case "cancel":
			if r.SenderID != actor {
				return FriendRequest{}, "bukan pengirim"
			}
			s.blob.Requests[i].Status = ReqCancelled
			s.blob.Requests[i].UpdatedAt = now
		default:
			return FriendRequest{}, "aksi tidak valid"
		}
		_ = s.flushLocked()
		return s.blob.Requests[i], ""
	}
	return FriendRequest{}, "tidak ditemukan"
}

func (s *SocialStore) RemoveFriend(actor, other string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	aa, bb := pairKey(actor, other)
	kept := s.blob.Friends[:0]
	found := false
	for _, f := range s.blob.Friends {
		if f.UserA == aa && f.UserB == bb {
			found = true
			continue
		}
		kept = append(kept, f)
	}
	if !found {
		return "bukan teman"
	}
	s.blob.Friends = kept
	_ = s.flushLocked()
	return ""
}

func (s *SocialStore) Block(blocker, blocked string) string {
	if blocker == blocked || blocked == "" {
		return "tidak valid"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blockedLocked(blocker, blocked) {
		return ""
	}
	s.blob.Blocks = append(s.blob.Blocks, UserBlock{BlockerID: blocker, BlockedID: blocked, CreatedAt: time.Now().UnixMilli()})
	aa, bb := pairKey(blocker, blocked)
	kept := s.blob.Friends[:0]
	for _, f := range s.blob.Friends {
		if f.UserA == aa && f.UserB == bb {
			continue
		}
		kept = append(kept, f)
	}
	s.blob.Friends = kept
	_ = s.flushLocked()
	return ""
}

func (s *SocialStore) Unblock(blocker, blocked string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.blob.Blocks[:0]
	for _, b := range s.blob.Blocks {
		if b.BlockerID == blocker && b.BlockedID == blocked {
			continue
		}
		kept = append(kept, b)
	}
	s.blob.Blocks = kept
	_ = s.flushLocked()
}

func (s *SocialStore) CreateInvite(from, to, roomID, code string) (GameInvite, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blockedLocked(from, to) {
		return GameInvite{}, "diblokir"
	}
	priv := s.privacyLocked(to)
	if !priv.AllowGameInvites {
		// Tetap izinkan undangan bermain (fitur inti); privacy hanya untuk teman.
		_ = priv
	}
	now := time.Now().UnixMilli()
	inv := GameInvite{
		ID: "gi-" + shortID(), SenderID: from, ReceiverID: to, RoomID: roomID, RoomCode: code,
		Status: ReqPending, ExpiresAt: now + int64(InviteTTLSec)*1000, CreatedAt: now,
	}
	s.blob.Invites = append(s.blob.Invites, inv)
	s.noteLocked(to, NoteGameInvite, "Undangan bermain", "Kamu diundang ke room "+code+".", inv.ID)
	_ = s.flushLocked()
	return inv, ""
}

func (s *SocialStore) RespondInvite(actor, inviteID, action string) (GameInvite, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	for i, inv := range s.blob.Invites {
		if inv.ID != inviteID {
			continue
		}
		if inv.Status != ReqPending {
			return inv, "sudah diproses"
		}
		if now > inv.ExpiresAt {
			s.blob.Invites[i].Status = InvExpired
			_ = s.flushLocked()
			return s.blob.Invites[i], "kedaluwarsa"
		}
		if inv.ReceiverID != actor {
			return GameInvite{}, "bukan penerima"
		}
		if action == "accept" {
			s.blob.Invites[i].Status = ReqAccepted
		} else {
			s.blob.Invites[i].Status = ReqRejected
		}
		_ = s.flushLocked()
		return s.blob.Invites[i], ""
	}
	return GameInvite{}, "tidak ditemukan"
}

func (s *SocialStore) ExpireInvites(now int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := false
	for i, inv := range s.blob.Invites {
		if inv.Status == ReqPending && now > inv.ExpiresAt {
			s.blob.Invites[i].Status = InvExpired
			changed = true
		}
	}
	if changed {
		_ = s.flushLocked()
	}
}

func (s *SocialStore) Notes(userID string, unreadOnly bool) []UlarNote {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]UlarNote, 0)
	for i := len(s.blob.Notes) - 1; i >= 0; i-- {
		n := s.blob.Notes[i]
		if n.UserID != userID {
			continue
		}
		if unreadOnly && n.Read {
			continue
		}
		out = append(out, n)
		if len(out) >= 50 {
			break
		}
	}
	return out
}

func (s *SocialStore) UnreadCount(userID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, x := range s.blob.Notes {
		if x.UserID == userID && !x.Read {
			n++
		}
	}
	return n
}

func (s *SocialStore) MarkRead(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, n := range s.blob.Notes {
		if n.UserID == userID {
			s.blob.Notes[i].Read = true
		}
	}
	_ = s.flushLocked()
}

func (s *SocialStore) Notify(userID, typ, title, body, ref string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.noteLocked(userID, typ, title, body, ref)
	_ = s.flushLocked()
}

func (s *SocialStore) PendingFor(userID string) (incoming, outgoing []FriendRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.blob.Requests {
		if r.Status != ReqPending {
			continue
		}
		if r.ReceiverID == userID {
			incoming = append(incoming, r)
		} else if r.SenderID == userID {
			outgoing = append(outgoing, r)
		}
	}
	return
}

func (s *SocialStore) Report(from, to, reason, desc, matchID string) (PlayerReport, string) {
	if from == to || to == "" {
		return PlayerReport{}, "tidak valid"
	}
	switch reason {
	case "Spam", "Harassment", "Inappropriate Name", "Cheating", "Other":
	default:
		return PlayerReport{}, "alasan tidak valid"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cut := time.Now().UnixMilli() - 3600*1000
	n := 0
	for _, r := range s.blob.Reports {
		if r.ReporterID == from && r.CreatedAt >= cut {
			n++
		}
	}
	if n >= 5 {
		return PlayerReport{}, "terlalu banyak laporan"
	}
	rep := PlayerReport{
		ID: "rp-" + shortID(), ReporterID: from, ReportedID: to, Reason: reason,
		Description: strings.TrimSpace(desc), MatchID: matchID, Status: "OPEN", CreatedAt: time.Now().UnixMilli(),
	}
	s.blob.Reports = append(s.blob.Reports, rep)
	_ = s.flushLocked()
	return rep, ""
}

func (s *SocialStore) AllReports() []PlayerReport {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]PlayerReport, len(s.blob.Reports))
	copy(out, s.blob.Reports)
	return out
}

func (s *SocialStore) ResolveReport(id, status, note, by string) (PlayerReport, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.blob.Reports {
		if r.ID != id {
			continue
		}
		switch status {
		case ReportResolved, ReportDismiss, ReportReview:
		default:
			return PlayerReport{}, "status tidak valid"
		}
		s.blob.Reports[i].Status = status
		s.blob.Reports[i].ResolutionNote = strings.TrimSpace(note)
		s.blob.Reports[i].ResolvedBy = strings.TrimSpace(by)
		_ = s.flushLocked()
		return s.blob.Reports[i], ""
	}
	return PlayerReport{}, "tidak ditemukan"
}

func allowedReportReason(r string) bool {
	switch r {
	case "Spam", "Harassment", "Inappropriate Name", "Cheating", "Other":
		return true
	}
	return false
}
