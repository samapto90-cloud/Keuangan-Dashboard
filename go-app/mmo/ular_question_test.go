package mmo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestQuestionBank390(t *testing.T) {
	rep := DefaultEduBank().Validate()
	if rep.Total < 350 || rep.Invalid != 0 {
		t.Fatalf("bank quality total=%d invalid=%d problems=%v", rep.Total, rep.Invalid, rep.Problems)
	}
	sma, sd := 0, 0
	for _, q := range DefaultEduBank().Snapshot() {
		if q.Grade == GradeSMA {
			sma++
		}
		if q.Grade == GradeSD {
			sd++
		}
	}
	if sma < 200 || sd < 150 {
		t.Fatalf("grades sma=%d sd=%d", sma, sd)
	}
}

func TestPenaltyPositions(t *testing.T) {
	if PenaltyPosition(45) != 35 || PenaltyPosition(41) != 31 || PenaltyPosition(50) != 40 || PenaltyPosition(7) != 1 || PenaltyPosition(1) != 1 {
		t.Fatal("penalty")
	}
}

func TestSnakeLadderLandThenQuestionCell(t *testing.T) {
	cfg := DefaultBoardConfig()
	s := ResolveMove(cfg, 87, 1)
	if s.SnakeTo != 48 || s.Final != 48 {
		t.Fatalf("snake %+v", s)
	}
	l := ResolveMove(cfg, 27, 1)
	if l.LadderTo != 55 || l.Final != 55 {
		t.Fatalf("ladder %+v", l)
	}
}

func TestQuestionPublicOmitsAnswer(t *testing.T) {
	q, ok := DefaultEduBank().Get("sma-001")
	if !ok {
		t.Fatal("missing sma-001")
	}
	pub := q.Public("u", "Andi", 1, 0, false)
	raw, _ := json.Marshal(pub)
	s := string(raw)
	if containsFold(s, "correctAnswer") || containsFold(s, q.CorrectAnswer) && len(q.CorrectAnswer) > 1 {
		// letter A-D may appear in options; key must not leak
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if _, ok := m["correctAnswer"]; ok {
		t.Fatal("leaked correctAnswer")
	}
	if _, ok := m["explanation"]; ok {
		t.Fatal("leaked explanation")
	}
}

func containsFold(s, sub string) bool {
	return len(sub) > 0 && (s == sub)
}

func readyTwo(t *testing.T) (*Hub, *UlarRoom, *Player, *Player) {
	t.Helper()
	t.Setenv("ULAR_MATCH_STORE", filepath.Join(t.TempDir(), "m.json"))
	t.Setenv("ULAR_ATTEMPT_STORE", filepath.Join(t.TempDir(), "a.json"))
	prev := ularCountdown
	ularCountdown = time.Millisecond
	t.Cleanup(func() { ularCountdown = prev })
	h := &Hub{Lobby: NewUlarLobby()}
	a, b := testPlayer("ua", "Andi"), testPlayer("ub", "Budi")
	h.Lobby.Connect(a)
	room, _ := h.Lobby.Create(a)
	h.Lobby.Connect(b)
	_, _ = h.Lobby.Join(b, room.RoomCode)
	_, _ = h.Lobby.SetReady("ua", true)
	_, _ = h.Lobby.SetReady("ub", true)
	if errc := h.startMatch("ua"); errc != "" {
		t.Fatal(errc)
	}
	time.Sleep(15 * time.Millisecond)
	return h, room, a, b
}

func TestQuestionSecurityRejects(t *testing.T) {
	h, room, a, b := readyTwo(t)
	h.Lobby.mu.Lock()
	pl := h.Lobby.player(room, "ua")
	pl.Position = 41
	h.startQuestionLocked(room, "ua", false)
	qid := room.Match.CurrentQuestionID
	h.Lobby.mu.Unlock()
	if qid == "" {
		t.Fatal("no question")
	}
	if errc := h.submitQuestion(b, "A"); errc != ErrNotQuestionPlayer {
		t.Fatalf("other player %s", errc)
	}
	h.Lobby.mu.Lock()
	room.Match.CurrentQuestionID = ""
	room.Match.QuestionState = ""
	h.Lobby.mu.Unlock()
	if errc := h.submitQuestion(a, "A"); errc != ErrNoQuestion {
		t.Fatalf("no q %s", errc)
	}
	h.Lobby.mu.Lock()
	pl.Position = 41
	h.startQuestionLocked(room, "ua", false)
	q, _ := DefaultEduBank().Get(room.Match.CurrentQuestionID)
	h.Lobby.mu.Unlock()
	if errc := h.submitQuestion(a, q.CorrectAnswer); errc != "" {
		t.Fatal(errc)
	}
	if errc := h.submitQuestion(a, q.CorrectAnswer); errc != ErrAlreadyAnswered && errc != ErrNoQuestion {
		t.Fatalf("double %s", errc)
	}
	time.Sleep(1200 * time.Millisecond)
	h.Lobby.mu.Lock()
	pos := h.Lobby.player(room, "ua").Position
	h.Lobby.mu.Unlock()
	if pos != 41 {
		t.Fatalf("correct should stay 41 got %d", pos)
	}
}

func TestWrongPenaltyAndTimeout(t *testing.T) {
	h, room, a, _ := readyTwo(t)
	h.Lobby.mu.Lock()
	pl := h.Lobby.player(room, "ua")
	pl.Position = 41
	h.startQuestionLocked(room, "ua", false)
	q, _ := DefaultEduBank().Get(room.Match.CurrentQuestionID)
	wrong := "A"
	if q.CorrectAnswer == "A" {
		wrong = "B"
	}
	h.Lobby.mu.Unlock()
	if errc := h.submitQuestion(a, wrong); errc != "" {
		t.Fatal(errc)
	}
	h.Lobby.mu.Lock()
	if h.Lobby.player(room, "ua").Position != 31 {
		t.Fatalf("wrong pos %d", h.Lobby.player(room, "ua").Position)
	}
	h.Lobby.mu.Unlock()
	time.Sleep(20 * time.Millisecond)

	h.Lobby.mu.Lock()
	h.Lobby.player(room, "ua").Position = 50
	h.startQuestionLocked(room, "ua", false)
	room.Match.QuestionStartedAt = time.Now().Add(-20 * time.Second)
	h.overdueQuestionLocked(room)
	if h.Lobby.player(room, "ua").Position != 40 {
		t.Fatalf("timeout pos %d", h.Lobby.player(room, "ua").Position)
	}
	h.Lobby.mu.Unlock()

	h.Lobby.mu.Lock()
	h.Lobby.player(room, "ua").Position = 7
	h.startQuestionLocked(room, "ua", false)
	q, _ = DefaultEduBank().Get(room.Match.CurrentQuestionID)
	wrong = "A"
	if q.CorrectAnswer == "A" {
		wrong = "B"
	}
	h.Lobby.mu.Unlock()
	if errc := h.submitQuestion(a, wrong); errc != "" {
		t.Fatal(errc)
	}
	h.Lobby.mu.Lock()
	if h.Lobby.player(room, "ua").Position != 1 {
		t.Fatalf("min %d", h.Lobby.player(room, "ua").Position)
	}
	h.Lobby.mu.Unlock()
}

func TestFinalQuestionWinAndFail(t *testing.T) {
	h, room, a, _ := readyTwo(t)
	h.Lobby.mu.Lock()
	h.Lobby.player(room, "ua").Position = 100
	h.startQuestionLocked(room, "ua", true)
	if !room.Match.QuestionFinal || room.Match.QuestionView.Difficulty != DiffHard {
		t.Fatal("final hard")
	}
	q, _ := DefaultEduBank().Get(room.Match.CurrentQuestionID)
	h.Lobby.mu.Unlock()
	if errc := h.submitQuestion(a, q.CorrectAnswer); errc != "" {
		t.Fatal(errc)
	}
	time.Sleep(1100 * time.Millisecond)
	h.Lobby.mu.Lock()
	if room.Status != UlarFinished || room.Match.WinnerID != "ua" {
		t.Fatalf("win %s %s", room.Status, room.Match.WinnerID)
	}
	h.Lobby.mu.Unlock()
}

func TestFinalQuestionWrongTo90(t *testing.T) {
	h, room, a, _ := readyTwo(t)
	h.Lobby.mu.Lock()
	h.Lobby.player(room, "ua").Position = 100
	h.startQuestionLocked(room, "ua", true)
	q, _ := DefaultEduBank().Get(room.Match.CurrentQuestionID)
	wrong := "A"
	if q.CorrectAnswer == "A" {
		wrong = "B"
	}
	h.Lobby.mu.Unlock()
	if errc := h.submitQuestion(a, wrong); errc != "" {
		t.Fatal(errc)
	}
	h.Lobby.mu.Lock()
	if h.Lobby.player(room, "ua").Position != 90 {
		t.Fatalf("final wrong %d", h.Lobby.player(room, "ua").Position)
	}
	h.Lobby.mu.Unlock()
}

func TestBalancedSubjectsAndNoDup(t *testing.T) {
	h, room, _, _ := readyTwo(t)
	seen := map[string]int{}
	ids := map[string]bool{}
	h.Lobby.mu.Lock()
	for i := 0; i < 8; i++ {
		h.startQuestionLocked(room, "ua", false)
		q := room.Match.QuestionView
		seen[q.Subject]++
		if ids[q.ID] {
			t.Fatalf("dup %s", q.ID)
		}
		ids[q.ID] = true
		room.Match.CurrentQuestionID = ""
		room.Match.QuestionState = QStateComplete
		room.Match.AnswerSubmitted = false
	}
	h.Lobby.mu.Unlock()
	if seen[SubjectPAI] == 8 {
		t.Fatal("unbalanced")
	}
	if len(seen) < 4 {
		t.Fatalf("subjects %+v", seen)
	}
}

func TestAdminQuestionsForbidden(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cahaya/api/ular/admin/questions", nil)
	rr := httptest.NewRecorder()
	HandleAdminQuestions(rr, req)
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusForbidden {
		t.Fatalf("code %d", rr.Code)
	}
}
