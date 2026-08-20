package mmo

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

var feedbackLimit = newUlarLimiter()

type PlayerFeedback struct {
	ID          string `json:"id"`
	UserID      string `json:"userId,omitempty"`
	Username    string `json:"username,omitempty"`
	Category    string `json:"category"`
	QuestionSub string `json:"questionSub,omitempty"`
	Message     string `json:"message"`
	Page        string `json:"page,omitempty"`
	MatchID     string `json:"matchId,omitempty"`
	QuestionID  string `json:"questionId,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
}

func (s *OpsStore) AddFeedback(f PlayerFeedback) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if f.ID == "" {
		f.ID = "fb-" + shortID()
	}
	if f.CreatedAt == 0 {
		f.CreatedAt = time.Now().UnixMilli()
	}
	// Store as audit entry for admin visibility without new blob migration.
	s.blob.Audit = append(s.blob.Audit, AuditLog{
		ID:         f.ID,
		AdminID:    f.UserID,
		AdminName:  f.Username,
		Action:     "PLAYER_FEEDBACK:" + f.Category,
		TargetType: "feedback",
		TargetID:   f.ID,
		AfterData:  clipJSON(f),
		CreatedAt:  f.CreatedAt,
	})
	if len(s.blob.Audit) > 5000 {
		s.blob.Audit = s.blob.Audit[len(s.blob.Audit)-4000:]
	}
	_ = s.flushLocked()
}

func HandleFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	ip := r.RemoteAddr
	if !feedbackLimit.allow("fb:"+ip, 5, 10*time.Minute) {
		writeAccountErr(w, http.StatusTooManyRequests, "rate limit")
		return
	}
	var in struct {
		Category    string `json:"category"`
		QuestionSub string `json:"questionSub"`
		Message     string `json:"message"`
		Page        string `json:"page"`
		MatchID     string `json:"matchId"`
		QuestionID  string `json:"questionId"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	msg := strings.TrimSpace(in.Message)
	if len(msg) < 5 || len(msg) > 1000 {
		writeAccountErr(w, http.StatusBadRequest, "pesan tidak valid")
		return
	}
	cat := strings.ToUpper(strings.TrimSpace(in.Category))
	switch cat {
	case "BUG", "SUGGESTION", "QUESTION_ISSUE", "OTHER":
	default:
		cat = "OTHER"
	}
	fb := PlayerFeedback{
		ID:          "fb-" + shortID(),
		Category:    cat,
		QuestionSub: strings.TrimSpace(in.QuestionSub),
		Message:     msg,
		Page:        clipStr(in.Page, 120),
		MatchID:     clipStr(in.MatchID, 64),
		QuestionID:  clipStr(in.QuestionID, 64),
		CreatedAt:   time.Now().UnixMilli(),
	}
	if DefaultHub != nil && DefaultHub.Accounts != nil {
		if sess := DefaultHub.Accounts.Lookup(bearerFrom(r)); sess != nil {
			fb.UserID = sess.PlayerID
			fb.Username = sess.Username
		}
	}
	if DefaultHub != nil && DefaultHub.Ops != nil {
		DefaultHub.Ops.AddFeedback(fb)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": fb.ID})
}

func clipStr(v string, n int) string {
	v = strings.TrimSpace(v)
	if len(v) <= n {
		return v
	}
	return v[:n]
}
