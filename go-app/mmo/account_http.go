package mmo

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type accountReq struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type sessionOut struct {
	Token     string `json:"token"`
	PlayerID  string `json:"playerId"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expiresAt"`
}

type ProfileOut struct {
	PlayerID         string `json:"playerId"`
	Username         string `json:"username"`
	Email            string `json:"email,omitempty"`
	Level            int    `json:"level"`
	Chapter          string `json:"chapter"`
	ChapterTitle     string `json:"chapterTitle"`
	ChapterIndex     int    `json:"chapterIndex"`
	Checkpoint       string `json:"checkpoint"`
	CheckpointName   string `json:"checkpointName"`
	Region           string `json:"region"`
	NewJourney       bool   `json:"newJourney"`
	HasPasswordPlain bool   `json:"-"`
}

var (
	loginHits   = map[string][]time.Time{}
	loginHitsMu sync.Mutex
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeAccountErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func readAccountReq(r *http.Request) (accountReq, error) {
	var in accountReq
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&in)
	return in, err
}

func loginLimited(r *http.Request) bool {
	ip := r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i > 0 {
		ip = ip[:i]
	}
	now := time.Now()
	loginHitsMu.Lock()
	defer loginHitsMu.Unlock()
	cut := now.Add(-10 * time.Minute)
	rows := loginHits[ip]
	kept := rows[:0]
	for _, t := range rows {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= 12 {
		loginHits[ip] = kept
		return true
	}
	loginHits[ip] = append(kept, now)
	return false
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if loginLimited(r) {
		writeAccountErr(w, http.StatusTooManyRequests, "terlalu banyak percobaan")
		return
	}
	in, err := readAccountReq(r)
	if err != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	sess, msg := DefaultHub.Accounts.Register(in.Username, in.Email, in.Password, in.ConfirmPassword)
	if msg != "" {
		writeAccountErr(w, http.StatusBadRequest, msg)
		return
	}
	writeJSON(w, http.StatusOK, sessionOut{Token: sess.Token, PlayerID: sess.PlayerID, Username: sess.Username, ExpiresAt: sess.Expires.UnixMilli()})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if loginLimited(r) {
		writeAccountErr(w, http.StatusTooManyRequests, "terlalu banyak percobaan")
		return
	}
	in, err := readAccountReq(r)
	if err != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	user := in.Username
	if user == "" {
		user = in.Email
	}
	sess, msg := DefaultHub.Accounts.Login(user, in.Password)
	if msg != "" {
		writeAccountErr(w, http.StatusUnauthorized, msg)
		return
	}
	writeJSON(w, http.StatusOK, sessionOut{Token: sess.Token, PlayerID: sess.PlayerID, Username: sess.Username, ExpiresAt: sess.Expires.UnixMilli()})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	token := bearerFrom(r)
	DefaultHub.Accounts.Logout(token)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if loginLimited(r) {
		writeAccountErr(w, http.StatusTooManyRequests, "terlalu banyak percobaan")
		return
	}
	in, err := readAccountReq(r)
	if err != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	// Mode lupa password: cukup username → reset ke default batam2026.
	if strings.TrimSpace(in.Password) == "" && strings.TrimSpace(in.ConfirmPassword) == "" {
		user := strings.TrimSpace(in.Username)
		if user == "" {
			user = strings.TrimSpace(in.Email) // form bisa kirim field user saja
		}
		if msg := DefaultHub.Accounts.ResetToDefaultPassword(user); msg != "" {
			writeAccountErr(w, http.StatusBadRequest, msg)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "defaultPassword": DefaultResetPassword})
		return
	}
	if msg := DefaultHub.Accounts.ResetPassword(in.Username, in.Email, in.Password, in.ConfirmPassword); msg != "" {
		writeAccountErr(w, http.StatusBadRequest, msg)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess := DefaultHub.Accounts.Lookup(bearerFrom(r))
	if sess == nil {
		writeAccountErr(w, http.StatusUnauthorized, "sesi tidak valid")
		return
	}
	acc := DefaultHub.Accounts.AccountByID(sess.PlayerID)
	if acc == nil {
		writeAccountErr(w, http.StatusUnauthorized, "akun tidak ditemukan")
		return
	}
	out := ProfileOut{
		PlayerID: acc.PlayerID, Username: acc.Username, Email: acc.Email,
		Level: 1, NewJourney: true,
	}
	if st := progressStore(); st != nil {
		writeJSON(w, http.StatusOK, mergeProfile(out, st.ViewFor(acc.PlayerID, acc.Username)))
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func mergeProfile(base ProfileOut, extra map[string]any) map[string]any {
	out := map[string]any{
		"playerId": base.PlayerID, "username": base.Username, "email": base.Email,
		"newJourney": base.NewJourney,
	}
	for k, v := range extra {
		out[k] = v
	}
	out["playerId"] = base.PlayerID
	out["username"] = extra["username"]
	if out["username"] == nil || out["username"] == "" {
		out["username"] = base.Username
	}
	return out
}

func bearerFrom(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return strings.TrimSpace(r.Header.Get("X-Session-Token"))
}
