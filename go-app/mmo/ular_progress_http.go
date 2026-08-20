package mmo

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func progressStore() *ProgressStore {
	if DefaultHub != nil && DefaultHub.Progress != nil {
		return DefaultHub.Progress
	}
	return nil
}

func requireProgressSession(w http.ResponseWriter, r *http.Request) (*GameSession, *ProgressStore, bool) {
	if DefaultHub == nil || DefaultHub.Accounts == nil {
		writeAccountErr(w, http.StatusServiceUnavailable, "akun tidak siap")
		return nil, nil, false
	}
	sess := DefaultHub.Accounts.Lookup(bearerFrom(r))
	if sess == nil {
		writeAccountErr(w, http.StatusUnauthorized, "sesi tidak valid")
		return nil, nil, false
	}
	st := progressStore()
	if st == nil {
		writeAccountErr(w, http.StatusServiceUnavailable, "progres tidak siap")
		return nil, nil, false
	}
	return sess, st, true
}

func rejectClientEconomy(w http.ResponseWriter, raw []byte) bool {
	var probe map[string]json.RawMessage
	if json.Unmarshal(raw, &probe) != nil {
		return false
	}
	for _, k := range []string{"xp", "coins", "level", "xpAmount", "coinAmount", "achievementUnlocked"} {
		if _, ok := probe[k]; ok {
			writeAccountErr(w, http.StatusForbidden, "server menentukan XP, koin, level, dan achievement")
			return true
		}
	}
	return false
}

func HandleProgressStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, st, ok := requireProgressSession(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, st.ViewFor(sess.PlayerID, sess.Username))
}

func HandleProgressAchievements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, st, ok := requireProgressSession(w, r)
	if !ok {
		return
	}
	view := st.ViewFor(sess.PlayerID, sess.Username)
	writeJSON(w, http.StatusOK, map[string]any{"achievements": view["achievements"], "titles": view["titles"]})
}

func HandleProgressHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, st, ok := requireProgressSession(w, r)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	writeJSON(w, http.StatusOK, map[string]any{"items": st.History(sess.PlayerID, page), "page": page, "pageSize": MaxMatchHistoryPage})
}

func HandleProgressUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, st, ok := requireProgressSession(w, r)
	if !ok {
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	if rejectClientEconomy(w, raw) {
		return
	}
	var in struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
		Title    string `json:"title"`
	}
	if json.Unmarshal(raw, &in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	name := strings.TrimSpace(in.Username)
	if name != "" {
		if v := validateUsername(name); v == "" || blockedName(name) {
			writeAccountErr(w, http.StatusBadRequest, "username tidak valid")
			return
		}
		if msg := DefaultHub.Accounts.Rename(sess.PlayerID, name); msg != "" {
			writeAccountErr(w, http.StatusBadRequest, msg)
			return
		}
		sess.Username = name
	}
	st.Ensure(sess.PlayerID, sess.Username)
	p, msg := st.UpdateProfile(sess.PlayerID, name, strings.TrimSpace(in.Avatar), strings.TrimSpace(in.Title))
	if msg != "" {
		writeAccountErr(w, http.StatusBadRequest, msg)
		return
	}
	writeJSON(w, http.StatusOK, st.ViewFor(p.UserID, p.Username))
}

func HandleDailyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, st, ok := requireProgressSession(w, r)
	if !ok {
		return
	}
	st.Ensure(sess.PlayerID, sess.Username)
	writeJSON(w, http.StatusOK, st.DailyStatus(sess.PlayerID, time.Now()))
}

func HandleDailyClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, st, ok := requireProgressSession(w, r)
	if !ok {
		return
	}
	if !LiveFlags().EnableDailyReward {
		writeAccountErr(w, http.StatusForbidden, "daily reward nonaktif")
		return
	}
	raw, _ := io.ReadAll(r.Body)
	if rejectClientEconomy(w, raw) {
		return
	}
	claim, ev, msg := st.ClaimDaily(sess.PlayerID, sess.Username, time.Now())
	if msg != "" {
		writeAccountErr(w, http.StatusConflict, msg)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"claim": claim, "reward": ev, "profile": st.ViewFor(sess.PlayerID, sess.Username)})
}

func HandleAdminAdjustCoins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	a, ok := requireAdmin(w, r, PermPlayerEdit)
	if !ok {
		return
	}
	var in struct {
		UserID string `json:"userId"`
		Amount int    `json:"amount"`
		Reason string `json:"reason"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.UserID == "" || strings.TrimSpace(in.Reason) == "" {
		writeAccountErr(w, http.StatusBadRequest, "userId dan reason wajib")
		return
	}
	st := progressStore()
	if st == nil {
		writeAccountErr(w, http.StatusServiceUnavailable, "progres tidak siap")
		return
	}
	tx, msg := st.AdminAdjustCoins(in.UserID, a.ID, strings.TrimSpace(in.Reason), in.Amount)
	if msg != "" {
		writeAccountErr(w, http.StatusBadRequest, msg)
		return
	}
	a.audit("PLAYER_EDIT", "player", in.UserID, "", clipJSON(tx), r.RemoteAddr)
	writeJSON(w, http.StatusOK, tx)
}

func blockedName(name string) bool {
	n := strings.ToLower(name)
	for _, b := range []string{"admin", "anjing", "kontol", "babi"} {
		if n == b || strings.Contains(n, b) {
			return true
		}
	}
	return false
}
