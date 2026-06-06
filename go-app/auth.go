package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
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
	Username  string
	Role      string
	Name      string
	AppModule string
	CreatedAt time.Time
	LastSeen  time.Time
	ClientIP  string
	Expires   time.Time
}

var (
	defaultUsers map[string]UserAccount
	sessions     = map[string]Session{}
	sessionsMu   sync.RWMutex
)

func initAuth() {
	adminUser := envOr("SIPKEU_ADMIN_USER", "admin")
	adminPass := envOr("SIPKEU_ADMIN_PASSWORD", "admin2026")
	opUser := envOr("SIPKEU_OPERATOR_USER", "operator")
	opPass := envOr("SIPKEU_OPERATOR_PASSWORD", "operator2026")

	defaultUsers = map[string]UserAccount{
		strings.ToLower(adminUser): {Password: adminPass, Role: "admin", Name: "Administrator SIPKEU"},
		strings.ToLower(opUser):    {Password: opPass, Role: "operator", Name: "Operator SIPKEU"},
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func isValidAppModule(id string) bool {
	switch id {
	case "sekretariat", "paud", "sd", "smp", "kas-belanja", "pengaturan":
		return true
	default:
		return false
	}
}

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
	now := time.Now()

	sessionsMu.RLock()
	sess, ok := sessions[token]
	sessionsMu.RUnlock()
	if !ok {
		return nil
	}
	if now.After(sess.Expires) {
		sessionsMu.Lock()
		delete(sessions, token)
		sessionsMu.Unlock()
		return nil
	}

	if now.Sub(sess.LastSeen) > 2*time.Minute {
		sessionsMu.Lock()
		if s, ok2 := sessions[token]; ok2 && now.Before(s.Expires) {
			half := sessionLifetime() / 2
			if now.After(s.Expires.Add(-half)) {
				s.Expires = now.Add(sessionLifetime())
			}
			s.LastSeen = now
			sessions[token] = s
			sess = s
		}
		sessionsMu.Unlock()
	}

	out := sess
	return &out
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
	password := body.Password
	if username == "" || password == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Username dan password wajib diisi"})
		return
	}
	if len(username) > 64 || len(password) > 128 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Username atau password tidak valid"})
		return
	}
	ip := clientIP(r)
	if loginLimiter.check(ip) {
		jsonResponse(w, http.StatusTooManyRequests, map[string]string{
			"error": "Terlalu banyak percobaan login. Coba lagi dalam 15 menit.",
		})
		return
	}
	appModule := strings.TrimSpace(r.Header.Get("X-SIPKEU-App"))
	if appModule == "" {
		appModule = "sekretariat"
	}
	if !isValidAppModule(appModule) {
		appModule = "sekretariat"
	}
	if appModule != "pengaturan" && !isPortalEnabled(appModule) {
		jsonResponse(w, http.StatusForbidden, map[string]string{
			"error": "Portal sedang dinonaktifkan oleh Super Admin. Hubungi administrator sistem.",
		})
		return
	}
	user, ok := authenticatePortalUser(username, password, appModule)
	if !ok {
		loginLimiter.recordFail(ip)
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Username atau password salah"})
		return
	}
	loginLimiter.recordSuccess(ip)
	token, err := newSessionToken()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal membuat sesi"})
		return
	}
	now := time.Now()
	sess := Session{
		Username:  username,
		Role:      user.Role,
		Name:      user.Name,
		AppModule: appModule,
		CreatedAt: now,
		LastSeen:  now,
		ClientIP:  ip,
		Expires:   now.Add(sessionLifetime()),
	}
	sessionsMu.Lock()
	sessions[token] = sess
	sessionsMu.Unlock()
	recordAudit(username, "login", appModule, "Login berhasil", ip)
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"token":       token,
		"username":    username,
		"role":        user.Role,
		"name":        user.Name,
		"app_module":  appModule,
		"permissions": permissionsForSession(&sess),
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
		"username":    sess.Username,
		"role":        sess.Role,
		"name":        sess.Name,
		"app_module":  sess.AppModule,
		"permissions": permissionsForSession(sess),
	})
}
