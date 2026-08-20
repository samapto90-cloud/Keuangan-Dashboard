package mmo

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type practiceSess struct {
	ID    string
	Q     EduQuestion
	Start time.Time
	Done  bool
	Final bool
}

var (
	practiceMu sync.Mutex
	practice   = map[string]*practiceSess{}
)

func adminAuthorized(r *http.Request) bool {
	tok := strings.TrimSpace(os.Getenv("ULAR_ADMIN_TOKEN"))
	if tok == "" {
		return false
	}
	got := strings.TrimSpace(r.Header.Get("X-Ular-Admin"))
	if got == "" {
		h := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(h), "bearer ") {
			got = strings.TrimSpace(h[7:])
		}
	}
	return got != "" && got == tok
}

func HandleQuestionValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	rep := DefaultEduBank().Validate()
	writeJSON(w, http.StatusOK, map[string]any{
		"pai": rep.PAI, "matematika": rep.Math, "bahasaInggris": rep.English, "bahasaJawa": rep.Jawa,
		"total": rep.Total, "invalid": rep.Invalid, "duplicate": rep.Duplicate,
		"easy": rep.Easy, "medium": rep.Medium, "hard": rep.Hard,
	})
}

func HandleQuestionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess := DefaultHub.Accounts.Lookup(bearerFrom(r))
	if sess == nil {
		writeAccountErr(w, http.StatusUnauthorized, "sesi tidak valid")
		return
	}
	if DefaultHub.Lobby == nil || DefaultHub.Lobby.Attempts == nil {
		writeJSON(w, http.StatusOK, summarizeAttempts(nil))
		return
	}
	writeJSON(w, http.StatusOK, summarizeAttempts(DefaultHub.Lobby.Attempts.ForUser(sess.PlayerID)))
}

func HandlePracticeQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in struct {
		Final bool     `json:"final"`
		Used  []string `json:"used"`
		Grade string   `json:"grade"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	diff := ""
	if in.Final {
		diff = DiffHard
	}
	grade := strings.ToUpper(strings.TrimSpace(in.Grade))
	if grade != GradeSD && grade != GradeSMA {
		grade = GradeSMA
	}
	q, used := DefaultEduBank().reserve(in.Used, "", diff, grade, true)
	if q.ID == "" {
		q, used = DefaultEduBank().reserve(in.Used, "", "", grade, true)
	}
	if q.ID == "" {
		writeAccountErr(w, http.StatusServiceUnavailable, "soal tidak tersedia")
		return
	}
	DefaultEduBank().bumpUsage(q.ID)
	id := "pq-" + shortID()
	practiceMu.Lock()
	practice[id] = &practiceSess{ID: id, Q: q, Start: time.Now(), Final: in.Final}
	practiceMu.Unlock()
	pub := q.Public("", "", 0, time.Now().Add(questionTimeLimit).UnixMilli(), in.Final)
	writeJSON(w, http.StatusOK, map[string]any{"practiceId": id, "question": pub, "used": used})
}

func HandlePracticeAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in struct {
		PracticeID string `json:"practiceId"`
		Answer     string `json:"answer"`
		Position   int    `json:"position"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	practiceMu.Lock()
	s := practice[in.PracticeID]
	if s == nil || s.Done {
		practiceMu.Unlock()
		writeAccountErr(w, http.StatusBadRequest, "sesi soal tidak valid")
		return
	}
	timeout := time.Since(s.Start) >= questionTimeLimit
	s.Done = true
	q := s.Q
	practiceMu.Unlock()
	ans := strings.ToUpper(strings.TrimSpace(in.Answer))
	correct := !timeout && ans == strings.ToUpper(q.CorrectAnswer)
	before := in.Position
	if before < OFFBOARD_START {
		before = OFFBOARD_START
	}
	after := before
	result := ResultCorrect
	if timeout {
		result = ResultTimeout
		after = PenaltyPosition(before)
	} else if !correct {
		result = ResultWrong
		after = PenaltyPosition(before)
	}
	fb := "Hebat! Kamu menjawab dengan tepat."
	if result != ResultCorrect {
		if timeout {
			fb = "Waktu habis. Yuk belajar lagi. Kesempatan berikutnya!"
		} else {
			fb = "Belum tepat. Yuk belajar lagi. Kesempatan berikutnya!"
		}
	}
	won := correct && s.Final
	writeJSON(w, http.StatusOK, map[string]any{
		"result": result, "correct": correct, "timeout": timeout,
		"correctAnswer": q.CorrectAnswer, "explanation": q.Explanation, "feedback": fb,
		"positionBeforePenalty": before, "penalty": before - after, "positionAfterPenalty": after,
		"path": PenaltyPath(before, after), "won": won, "final": s.Final,
	})
}

func HandleAdminQuestions(w http.ResponseWriter, r *http.Request) {
	perm := PermQuestionView
	switch r.Method {
	case http.MethodPost:
		perm = PermQuestionCreate
	case http.MethodPut, http.MethodPatch:
		perm = PermQuestionEdit
	case http.MethodDelete:
		perm = PermQuestionDelete
	}
	a, ok := requireAdmin(w, r, perm)
	if !ok {
		return
	}
	bank := DefaultEduBank()
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		writeJSON(w, http.StatusOK, bank.ListAdmin(q.Get("subject"), q.Get("difficulty"), q.Get("q")))
	case http.MethodPost:
		var item EduQuestion
		if json.NewDecoder(r.Body).Decode(&item) != nil {
			writeAccountErr(w, http.StatusBadRequest, "payload")
			return
		}
		if item.ID == "" {
			item.ID = "q-" + shortID()
		}
		item.Active = true
		if err := bank.Upsert(item); err != nil {
			writeAccountErr(w, http.StatusBadRequest, err.Error())
			return
		}
		a.audit("QUESTION_CREATED", "question", item.ID, "", clipJSON(item.ID), r.RemoteAddr)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": item.ID})
	case http.MethodPut:
		var item EduQuestion
		if json.NewDecoder(r.Body).Decode(&item) != nil || item.ID == "" {
			writeAccountErr(w, http.StatusBadRequest, "payload")
			return
		}
		old, _ := bank.Get(item.ID)
		if err := bank.Upsert(item); err != nil {
			writeAccountErr(w, http.StatusBadRequest, err.Error())
			return
		}
		a.audit("QUESTION_UPDATED", "question", item.ID, clipJSON(old.Question), clipJSON(item.Question), r.RemoteAddr)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" || !bank.Delete(id) {
			writeAccountErr(w, http.StatusNotFound, "not found")
			return
		}
		a.audit("QUESTION_DELETED", "question", id, "", "soft-delete", r.RemoteAddr)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	case http.MethodPatch:
		var in struct {
			ID     string `json:"id"`
			Active *bool  `json:"active"`
			Seed   bool   `json:"seed"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			writeAccountErr(w, http.StatusBadRequest, "payload")
			return
		}
		if in.Seed {
			if err := bank.SeedFromEmbedded(); err != nil {
				writeAccountErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, bank.Validate())
			return
		}
		if in.ID == "" || in.Active == nil {
			writeAccountErr(w, http.StatusBadRequest, "payload")
			return
		}
		if !bank.SetActive(in.ID, *in.Active) {
			writeAccountErr(w, http.StatusNotFound, "not found")
			return
		}
		a.audit("QUESTION_UPDATED", "question", in.ID, "", "active="+strconv.FormatBool(*in.Active), r.RemoteAddr)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
	}
}
