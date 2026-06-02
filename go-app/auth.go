package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type UserAccount struct {
	Password string
	Role     string
	Name     string
}

type Session struct {
	Username string
	Role     string
	Name     string
	Expires  time.Time
}

var (
	defaultUsers = map[string]UserAccount{
		"admin":    {Password: "admin2026", Role: "admin", Name: "Administrator SIPKEU"},
		"operator": {Password: "operator2026", Role: "operator", Name: "Operator SIPKEU"},
	}
	sessions   = map[string]Session{}
	sessionsMu sync.RWMutex
)

func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func bearerToken(r *http.Request) string {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return strings.TrimSpace(r.Header.Get("X-Session-Token"))
}

func getSession(r *http.Request) *Session {
	token := bearerToken(r)
	if token == "" {
		return nil
	}
	sessionsMu.RLock()
	sess, ok := sessions[token]
	sessionsMu.RUnlock()
	if !ok || time.Now().After(sess.Expires) {
		if ok {
			sessionsMu.Lock()
			delete(sessions, token)
			sessionsMu.Unlock()
		}
		return nil
	}
	return &sess
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if getSession(r) == nil {
			jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
			return
		}
		next(w, r)
	}
}

func requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := getSession(r)
		if sess == nil {
			jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
			return
		}
		if sess.Role != "admin" {
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses hanya untuk Admin"})
			return
		}
		next(w, r)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	username := strings.TrimSpace(strings.ToLower(body.Username))
	user, ok := defaultUsers[username]
	if !ok || user.Password != body.Password {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Username atau password salah"})
		return
	}
	token, err := newSessionToken()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal membuat sesi"})
		return
	}
	sess := Session{
		Username: username,
		Role:     user.Role,
		Name:     user.Name,
		Expires:  time.Now().Add(8 * time.Hour),
	}
	sessionsMu.Lock()
	sessions[token] = sess
	sessionsMu.Unlock()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"token":    token,
		"username": username,
		"role":     user.Role,
		"name":     user.Name,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	token := bearerToken(r)
	if token != "" {
		sessionsMu.Lock()
		delete(sessions, token)
		sessionsMu.Unlock()
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Logout berhasil"})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	sess := getSession(r)
	if sess == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid"})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"username": sess.Username,
		"role":     sess.Role,
		"name":     sess.Name,
	})
}
