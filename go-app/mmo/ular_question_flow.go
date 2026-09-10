package mmo

import (
	"log"
	"strings"
	"time"
)

const (
	TypeQuestionStart    = "question:start"
	TypeQuestionAnswer   = "question:answer"
	TypeQuestionResult   = "question:result"
	TypeQuestionTimeout  = "question:timeout"
	TypeQuestionPenalty  = "question:penalty"
	TypeQuestionComplete = "question:complete"
	TypeQuestionHistory  = "question:history"
)

func (m *UlarLiveMatch) ensurePool() {
	if m.UsedQuestionIDs == nil {
		m.UsedQuestionIDs = []string{}
	}
	if m.SubjectCursor < 0 {
		m.SubjectCursor = int(time.Now().UnixNano() % 4)
	}
}

func (m *UlarLiveMatch) nextSubject() string {
	m.ensurePool()
	s := eduSubjects[m.SubjectCursor%len(eduSubjects)]
	m.SubjectCursor = (m.SubjectCursor + 1) % len(eduSubjects)
	return s
}

func (h *Hub) startQuestionLocked(room *UlarRoom, userID string, final bool) {
	m := room.Match
	pl := h.Lobby.player(room, userID)
	if m == nil || pl == nil {
		return
	}
	m.ensurePool()
	bank := DefaultEduBank()
	diff := ""
	if final {
		diff = DiffHard
	}
	grade := normalizeGrade(m.Grade)
	if grade == "" {
		grade = normalizeGrade(room.Grade)
	}
	m.Grade = grade
	subject := ""
	if grade != GradeSD {
		subject = m.nextSubject()
	}
	// Coba filter ketat dulu, lalu longgarkan difficulty — jangan pernah ganti tingkat (SD/SMA).
	q, used := bank.reserve(m.UsedQuestionIDs, subject, diff, grade, true)
	if q.ID == "" && subject != "" {
		q, used = bank.reserve(m.UsedQuestionIDs, "", diff, grade, true)
	}
	if q.ID == "" && diff != "" {
		q, used = bank.reserve(m.UsedQuestionIDs, subject, "", grade, true)
	}
	if q.ID == "" && subject != "" {
		q, used = bank.reserve(m.UsedQuestionIDs, "", "", grade, true)
	}
	if q.ID == "" {
		q, used = bank.reserve(m.UsedQuestionIDs, "", "", grade, true)
	}
	if q.ID == "" {
		log.Printf("QUESTION_EMPTY room=%s grade=%s final=%v", room.RoomCode, grade, final)
		next := h.Lobby.nextAlive(room, userID)
		m.CurrentPlayerID = next
		m.TurnNumber++
		m.Phase = UlarPlayerTurn
		m.WaitingAnim = false
		m.LastAction = "TURN"
		snapState := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, ""))
		snapTurn := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, ""))
		h.emitLocked(room, TypeGameState, snapState)
		h.emitLocked(room, TypeGameTurn, snapTurn)
		return
	}
	bank.bumpUsage(q.ID)
	now := time.Now()
	m.UsedQuestionIDs = used
	m.CurrentQuestionID = q.ID
	m.QuestionPlayerID = userID
	m.QuestionStartedAt = now
	m.QuestionState = QStateActive
	m.AnswerSubmitted = false
	m.QuestionFinal = final
	m.QuestionNumber++
	m.Penalty = 0
	m.PositionBeforePenalty = pl.Position
	m.Phase = UlarOnQuestion
	m.WaitingAnim = false
	m.LastAction = "QUESTION"
	m.LastActionAt = now
	lim := questionTimeLimit
	if m.QuestionLimit > 0 {
		lim = m.QuestionLimit
	}
	ends := now.Add(lim).UnixMilli()
	pub := q.Public(userID, pl.Username, m.QuestionNumber, ends, final)
	if m.QuestionLimit > 0 {
		pub.TimeLimit = int(m.QuestionLimit.Seconds())
	}
	m.QuestionView = pub
	code := room.RoomCode
	matchID := m.ID
	qid := q.ID
	seqQ, evidQ, atQ := h.Lobby.nextSeq(room)
	snapState := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, ""))
	go func() {
		time.Sleep(lim)
		h.questionTimeout(code, matchID, qid)
	}()
	h.emitLocked(room, TypeQuestionStart, struct {
		UlarEnvelope
		Question QuestionPublic `json:"question"`
		State    string         `json:"questionState"`
	}{UlarEnvelope: UlarEnvelope{Seq: seqQ, EventID: evidQ, At: atQ}, Question: pub, State: QStateActive})
	h.emitLocked(room, TypeGameState, snapState)
}

func (h *Hub) tickUlarQuestions() {
	if h == nil || h.Lobby == nil {
		return
	}
	h.Lobby.mu.Lock()
	rooms := make([]*UlarRoom, 0, len(h.Lobby.rooms))
	for _, r := range h.Lobby.rooms {
		if r != nil && r.Match != nil && r.Status == UlarPlaying {
			rooms = append(rooms, r)
		}
	}
	for _, room := range rooms {
		h.overdueQuestionLocked(room)
	}
	h.Lobby.mu.Unlock()
}

func (h *Hub) questionTimeout(roomCode, matchID, questionID string) {
	l := h.Lobby
	l.mu.Lock()
	room := l.rooms[l.byCode[roomCode]]
	if room == nil || room.Match == nil || room.Match.ID != matchID {
		l.mu.Unlock()
		return
	}
	m := room.Match
	if m.CurrentQuestionID != questionID || m.AnswerSubmitted || (m.QuestionState != QStateActive && m.QuestionState != QStatePending) {
		l.mu.Unlock()
		return
	}
	h.settleQuestionLocked(room, "", true)
	l.mu.Unlock()
}

func (h *Hub) overdueQuestionLocked(room *UlarRoom) {
	m := room.Match
	if m == nil || m.CurrentQuestionID == "" {
		return
	}
	if m.QuestionState != QStateActive && m.QuestionState != QStatePending {
		return
	}
	if m.AnswerSubmitted {
		return
	}
	lim := questionTimeLimit
	if m.QuestionLimit > 0 {
		lim = m.QuestionLimit
	}
	if time.Since(m.QuestionStartedAt) < lim {
		return
	}
	h.settleQuestionLocked(room, "", true)
}

func (h *Hub) submitQuestion(p *Player, answer string) string {
	l := h.Lobby
	l.mu.Lock()
	defer l.mu.Unlock()
	room := l.rooms[l.inRoom[p.ID]]
	if room == nil || room.Match == nil {
		return ErrNotInRoom
	}
	m := room.Match
	if m.AnswerSubmitted {
		return ErrAlreadyAnswered
	}
	if m.CurrentQuestionID == "" || (m.QuestionState != QStateActive && m.QuestionState != QStatePending) {
		return ErrNoQuestion
	}
	if m.QuestionPlayerID != p.ID {
		return ErrNotQuestionPlayer
	}
	lim := questionTimeLimit
	if m.QuestionLimit > 0 {
		lim = m.QuestionLimit
	}
	if time.Since(m.QuestionStartedAt) >= lim {
		h.settleQuestionLocked(room, "", true)
		return ErrLateAnswer
	}
	ans := strings.ToUpper(strings.TrimSpace(answer))
	if ans != "A" && ans != "B" && ans != "C" && ans != "D" {
		return ErrInvalidRequest
	}
	h.settleQuestionLocked(room, ans, false)
	return ""
}

func (h *Hub) settleQuestionLocked(room *UlarRoom, answer string, timeout bool) {
	m := room.Match
	if m == nil || m.CurrentQuestionID == "" {
		return
	}
	if m.AnswerSubmitted && m.QuestionState != QStateActive && m.QuestionState != QStatePending {
		return
	}
	pl := h.Lobby.player(room, m.QuestionPlayerID)
	q, ok := DefaultEduBank().Get(m.CurrentQuestionID)
	if pl == nil || !ok {
		log.Printf("QUESTION_SETTLE_ABORT room=%s q=%s plNil=%v ok=%v", room.RoomCode, m.CurrentQuestionID, pl == nil, ok)
		m.QuestionState = QStateComplete
		m.CurrentQuestionID = ""
		m.QuestionView = QuestionPublic{}
		m.AnswerSubmitted = false
		m.WaitingAnim = false
		next := h.Lobby.nextAlive(room, m.QuestionPlayerID)
		if next == "" && pl != nil {
			next = pl.UserID
		}
		m.CurrentPlayerID = next
		m.TurnNumber++
		m.Phase = UlarPlayerTurn
		m.QuestionPlayerID = ""
		m.LastAction = "TURN"
		snapState := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, ""))
		snapTurn := h.Lobby.wrap(room, h.Lobby.snapshotLocked(room, ""))
		h.emitLocked(room, TypeGameState, snapState)
		h.emitLocked(room, TypeGameTurn, snapTurn)
		return
	}
	m.AnswerSubmitted = true
	m.QuestionState = QStateSubmitted
	elapsed := time.Since(m.QuestionStartedAt).Milliseconds()
	correct := !timeout && strings.EqualFold(answer, q.CorrectAnswer)
	before := pl.Position
	after := before
	result := ResultCorrect
	if timeout {
		result = ResultTimeout
	} else if !correct {
		result = ResultWrong
	}
	feedback := "Hebat! Kamu menjawab dengan tepat."
	if result != ResultCorrect {
		after = PenaltyPositionN(before, m.PenaltyN)
		m.Penalty = before - after
		m.PositionBeforePenalty = before
		m.QuestionState = QStatePenalty
		m.Phase = UlarPenalty
		pl.Position = after
		if timeout {
			feedback = "Waktu habis. Yuk belajar lagi. Kesempatan berikutnya!"
		} else {
			feedback = "Belum tepat. Yuk belajar lagi. Kesempatan berikutnya!"
		}
	} else {
		m.Penalty = 0
		m.QuestionState = QStateResult
		m.Phase = UlarOnQuestion
	}
	att := QuestionAttempt{
		ID: "qa-" + shortID(), MatchID: m.ID, UserID: pl.UserID, QuestionID: q.ID, Subject: q.Subject,
		Answer: answer, Correct: correct, Timeout: timeout, PositionBefore: before, PositionAfter: after,
		TimeTaken: elapsed, CreatedAt: time.Now().UnixMilli(),
	}
	m.Attempts = append(m.Attempts, att)
	if h.Lobby.Attempts != nil {
		h.Lobby.Attempts.Append(att)
	}
	var reward RewardEvent
	if h.Progress != nil {
		reward = h.Progress.RecordAnswer(
			pl.UserID,
			pl.Username,
			q.Subject,
			att.ID,
			correct,
			timeout,
			m.XPCorrectAmt,
			m.XPWrongAmt,
			m.XPTimeoutAmt,
			m.AchievementsByID,
		)
	}
	seq, evid, at := h.Lobby.nextSeq(room)
	env := UlarEnvelope{Seq: seq, EventID: evid, At: at}
	payload := QuestionResultOut{
		UlarEnvelope:          env,
		UserID:                pl.UserID,
		Username:              pl.Username,
		Result:                result,
		Correct:               correct,
		Timeout:               timeout,
		CorrectAnswer:         q.CorrectAnswer,
		Explanation:           q.Explanation,
		Feedback:              feedback,
		PositionBeforePenalty: before,
		Penalty:               m.Penalty,
		PositionAfterPenalty:  after,
		Path:                  PenaltyPath(before, after),
		Final:                 m.QuestionFinal,
		Won:                   correct && m.QuestionFinal,
		QuestionID:            q.ID,
		Subject:               q.Subject,
		Reward:                reward,
		Items:                 cloneItemMap(pl.Items),
	}
	typ := TypeQuestionResult
	if timeout {
		typ = TypeQuestionTimeout
	}
	h.emitLocked(room, typ, payload)
	if result != ResultCorrect {
		seq2, evid2, at2 := h.Lobby.nextSeq(room)
		pen := payload
		pen.UlarEnvelope = UlarEnvelope{Seq: seq2, EventID: evid2, At: at2}
		h.emitLocked(room, TypeQuestionPenalty, pen)
	}
	code := room.RoomCode
	matchID := m.ID
	won := payload.Won
	wait := 900 * time.Millisecond
	if result != ResultCorrect {
		wait = time.Duration(len(payload.Path))*time.Millisecond*MOVE_DURATION + 400*time.Millisecond
	}
	go func() {
		time.Sleep(wait)
		h.completeQuestion(code, matchID, won, pl.UserID)
	}()
}

func (h *Hub) completeQuestion(roomCode, matchID string, won bool, userID string) {
	l := h.Lobby
	l.mu.Lock()
	room := l.rooms[l.byCode[roomCode]]
	if room == nil || room.Match == nil || room.Match.ID != matchID {
		l.mu.Unlock()
		return
	}
	m := room.Match
	m.QuestionState = QStateComplete
	m.CurrentQuestionID = ""
	m.QuestionView = QuestionPublic{}
	if won {
		now := time.Now()
		pl := l.player(room, userID)
		if pl != nil {
			pl.PlayState = "FINISHED"
			pl.Position = MAX_POSITION
		}
		room.Status = UlarFinished
		m.Status = UlarFinished
		m.Phase = UlarFinished
		m.WinnerID = userID
		m.FinishedAt = &now
		h.persistMatchLocked(room)
		events := h.settleMatchRewardsLocked(room)
		snapDone := l.wrap(room, l.snapshotLocked(room, ""))
		snapDone.Rewards = events
		snapFinish := l.wrap(room, l.snapshotLocked(room, ""))
		snapFinish.Rewards = events
		l.mu.Unlock()
		h.pushRoom(room, TypeQuestionComplete, snapDone)
		h.pushRoom(room, TypeGameFinish, snapFinish)
		return
	}
	next := l.nextAlive(room, userID)
	m.CurrentPlayerID = next
	m.TurnNumber++
	m.Phase = UlarPlayerTurn
	m.QuestionPlayerID = ""
	m.AnswerSubmitted = false
	m.QuestionFinal = false
	m.WaitingAnim = false
	m.LastAction = "TURN"
	snapComplete := l.wrap(room, l.snapshotLocked(room, ""))
	snapState := l.wrap(room, l.snapshotLocked(room, ""))
	snapTurn := l.wrap(room, l.snapshotLocked(room, ""))
	l.mu.Unlock()
	h.pushRoom(room, TypeQuestionComplete, snapComplete)
	h.pushRoom(room, TypeGameState, snapState)
	h.pushRoom(room, TypeGameTurn, snapTurn)
}

type QuestionResultOut struct {
	UlarEnvelope
	UserID                string      `json:"userId"`
	Username              string      `json:"username"`
	Result                string      `json:"result"`
	Correct               bool        `json:"correct"`
	Timeout               bool        `json:"timeout"`
	CorrectAnswer         string      `json:"correctAnswer"`
	Explanation           string      `json:"explanation"`
	Feedback              string      `json:"feedback"`
	PositionBeforePenalty int         `json:"positionBeforePenalty"`
	Penalty               int         `json:"penalty"`
	PositionAfterPenalty  int         `json:"positionAfterPenalty"`
	Path                  []int       `json:"path"`
	Final                 bool        `json:"final"`
	Won                   bool        `json:"won"`
	QuestionID            string      `json:"questionId"`
	Subject               string      `json:"subject"`
	Reward                RewardEvent `json:"reward,omitempty"`
	PowerGrant            string      `json:"powerGrant,omitempty"`
	Items                 map[string]int `json:"items,omitempty"`
}

type QuestionAnswerIn struct {
	Answer        string `json:"answer"`
	CorrectAnswer string `json:"correctAnswer"`
	QuestionID    string `json:"questionId"`
	Position      int    `json:"position"`
	Elapsed       int    `json:"elapsed"`
}
