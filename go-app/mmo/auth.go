package mmo

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

type Session struct {
	Token    string
	PlayerID string
	Name     string
	Expires  time.Time
}

type AuthStore struct {
	mu    sync.Mutex
	byTok map[string]*Session
}

func NewAuthStore() *AuthStore {
	return &AuthStore{byTok: map[string]*Session{}}
}

func (a *AuthStore) Issue(name string) *Session {
	name = sanitizeName(name)
	s := &Session{
		Token:    randomID("tok_"),
		PlayerID: randomID("p_"),
		Name:     name,
		Expires:  time.Now().Add(12 * time.Hour),
	}
	a.mu.Lock()
	a.byTok[s.Token] = s
	a.mu.Unlock()
	return s
}

func (a *AuthStore) Lookup(token string) *Session {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.byTok[token]
	if s == nil || time.Now().After(s.Expires) {
		return nil
	}
	return s
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Raka"
	}
	runes := []rune(name)
	if len(runes) > 18 {
		runes = runes[:18]
	}
	return string(runes)
}

func randomID(prefix string) string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return prefix + hex.EncodeToString(b[:])
}
