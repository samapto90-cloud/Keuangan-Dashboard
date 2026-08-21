package mmo

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTP realtime bridge: fallback saat WebSocket tidak tersedia di edge (Hostinger :443).

type bridgeStore struct {
	mu      sync.Mutex
	players map[string]*Player
}

var httpBridge = &bridgeStore{players: map[string]*Player{}}

func (b *bridgeStore) ensure(sess *GameSession) *Player {
	b.mu.Lock()
	defer b.mu.Unlock()
	if p := b.players[sess.PlayerID]; p != nil {
		return p
	}
	p := &Player{
		ID:        sess.PlayerID,
		SessionID: sess.Token,
		Name:      sess.Username,
		Level:     1,
		State:     "IDLE",
		send:      make(chan []byte, 256),
	}
	b.players[sess.PlayerID] = p
	if DefaultHub != nil {
		DefaultHub.lobbyJoin(p)
	}
	select {
	case p.send <- marshal(TypeAuthOk, AuthOkOut{Token: sess.Token, PlayerID: sess.PlayerID, SessionID: sess.Token}):
	default:
	}
	return p
}

func requireBridgeSession(w http.ResponseWriter, r *http.Request) (*GameSession, bool) {
	tok := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(tok), "bearer ") {
		tok = strings.TrimSpace(tok[7:])
	}
	if tok == "" {
		tok = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if DefaultHub == nil || DefaultHub.Accounts == nil {
		writeAccountErr(w, http.StatusServiceUnavailable, "hub tidak siap")
		return nil, false
	}
	sess := DefaultHub.Accounts.Lookup(tok)
	if sess == nil {
		writeAccountErr(w, http.StatusUnauthorized, "sesi tidak valid")
		return nil, false
	}
	return sess, true
}

func HandleRealtimePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, ok := requireBridgeSession(w, r)
	if !ok {
		return
	}
	p := httpBridge.ensure(sess)
	if DefaultHub != nil && DefaultHub.Social != nil {
		DefaultHub.Social.Touch(p.ID, false)
	}
	timer := time.NewTimer(20 * time.Second)
	defer timer.Stop()
	select {
	case msg := <-p.send:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(msg)
	case <-timer.C:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(`{"type":"PONG","data":{"t":0}}`))
	case <-r.Context().Done():
		return
	}
}

func HandleRealtimeSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, ok := requireBridgeSession(w, r)
	if !ok {
		return
	}
	p := httpBridge.ensure(sess)
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeAccountErr(w, http.StatusBadRequest, "body")
		return
	}
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil || env.Type == "" {
		writeAccountErr(w, http.StatusBadRequest, "json")
		return
	}
	if env.Type == TypeAuth || env.Type == TypeJoinLobby {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	if DefaultHub != nil {
		DefaultHub.Enqueue(p, env)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
