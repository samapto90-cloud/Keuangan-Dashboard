package mmo

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

//go:embed data/sma-edu-questions.json
var smaEduQuestionsJSON []byte

const (
	SubjectPAI     = "PAI"
	SubjectMath    = "MATEMATIKA"
	SubjectEnglish = "BAHASA_INGGRIS"
	SubjectJawa    = "BAHASA_JAWA"

	GradeSD  = "SD"
	GradeSMP = "SMP"
	GradeSMA = "SMA"

	DiffEasy   = "EASY"
	DiffMedium = "MEDIUM"
	DiffHard   = "HARD"

	QStatePending   = "QUESTION_PENDING"
	QStateActive    = "QUESTION_ACTIVE"
	QStateSubmitted = "ANSWER_SUBMITTED"
	QStateResult    = "QUESTION_RESULT"
	QStatePenalty   = "PENALTY"
	QStateComplete  = "QUESTION_COMPLETE"

	ResultCorrect = "CORRECT"
	ResultWrong   = "WRONG"
	ResultTimeout = "TIMEOUT"

	QUESTION_TIME_LIMIT_SEC = 15
	QUESTION_PENALTY        = 10
)

var questionTimeLimit = QUESTION_TIME_LIMIT_SEC * time.Second

var eduSubjects = []string{SubjectPAI, SubjectMath, SubjectEnglish, SubjectJawa, "IPA", "IPS", "UMUM"}

type EduQuestion struct {
	ID            string `json:"id"`
	Subject       string `json:"subject"`
	Category      string `json:"category"`
	Grade         string `json:"grade"`
	Difficulty    string `json:"difficulty"`
	Question      string `json:"question"`
	OptionA       string `json:"optionA"`
	OptionB       string `json:"optionB"`
	OptionC       string `json:"optionC"`
	OptionD       string `json:"optionD"`
	CorrectAnswer string `json:"correctAnswer"`
	Explanation   string `json:"explanation"`
	Active        bool   `json:"active"`
	Deleted       bool   `json:"deleted,omitempty"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
	Usage         int    `json:"usage,omitempty"`
}

type QuestionPublic struct {
	ID         string `json:"id"`
	Subject    string `json:"subject"`
	Category   string `json:"category"`
	Grade      string `json:"grade"`
	Difficulty string `json:"difficulty"`
	Question   string `json:"question"`
	OptionA    string `json:"optionA"`
	OptionB    string `json:"optionB"`
	OptionC    string `json:"optionC"`
	OptionD    string `json:"optionD"`
	Number     int    `json:"number,omitempty"`
	TimeLimit  int    `json:"timeLimit"`
	EndsAt     int64  `json:"endsAt,omitempty"`
	PlayerID   string `json:"playerId,omitempty"`
	Username   string `json:"username,omitempty"`
	Final      bool   `json:"final,omitempty"`
	QuestionNo int    `json:"questionNo,omitempty"`
}

type QuestionAttempt struct {
	ID             string `json:"id"`
	MatchID        string `json:"matchId"`
	UserID         string `json:"userId"`
	QuestionID     string `json:"questionId"`
	Subject        string `json:"subject"`
	Answer         string `json:"answer,omitempty"`
	Correct        bool   `json:"correct"`
	Timeout        bool   `json:"timeout"`
	PositionBefore int    `json:"positionBefore"`
	PositionAfter  int    `json:"positionAfter"`
	TimeTaken      int64  `json:"timeTaken"`
	CreatedAt      int64  `json:"createdAt"`
}

type BankReport struct {
	PAI       int      `json:"pai"`
	Math      int      `json:"matematika"`
	English   int      `json:"bahasaInggris"`
	Jawa      int      `json:"bahasaJawa"`
	Total     int      `json:"total"`
	Invalid   int      `json:"invalid"`
	Duplicate int      `json:"duplicate"`
	Easy      int      `json:"easy"`
	Medium    int      `json:"medium"`
	Hard      int      `json:"hard"`
	Problems  []string `json:"problems,omitempty"`
}

type EduBank struct {
	mu    sync.RWMutex
	items []EduQuestion
	byID  map[string]int
}

var defaultEduBank *EduBank

func init() {
	b, err := LoadEduBankBytes(smaEduQuestionsJSON)
	if err != nil {
		panic("sma question bank: " + err.Error())
	}
	defaultEduBank = b
}

func DefaultEduBank() *EduBank {
	return defaultEduBank
}

func LoadEduBankBytes(raw []byte) (*EduBank, error) {
	var items []EduQuestion
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	b := &EduBank{items: items, byID: map[string]int{}}
	for i := range b.items {
		b.byID[b.items[i].ID] = i
	}
	rep := b.Validate()
	if rep.Total < 350 || rep.Invalid != 0 {
		return nil, fmt.Errorf("bank invalid: %+v %v", rep, rep.Problems)
	}
	return b, nil
}

func (b *EduBank) Snapshot() []EduQuestion {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]EduQuestion, len(b.items))
	copy(out, b.items)
	return out
}

func (b *EduBank) Validate() BankReport {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return validateItems(b.items)
}

func validateItems(items []EduQuestion) BankReport {
	rep := BankReport{}
	seenID := map[string]bool{}
	seenQ := map[string]bool{}
	for i, q := range items {
		if q.Deleted {
			continue
		}
		bad := false
		if strings.TrimSpace(q.ID) == "" || strings.TrimSpace(q.Question) == "" || strings.TrimSpace(q.Explanation) == "" {
			bad = true
		}
		if strings.TrimSpace(q.OptionA) == "" || strings.TrimSpace(q.OptionB) == "" || strings.TrimSpace(q.OptionC) == "" || strings.TrimSpace(q.OptionD) == "" {
			bad = true
		}
		ans := strings.ToUpper(strings.TrimSpace(q.CorrectAnswer))
		if ans != "A" && ans != "B" && ans != "C" && ans != "D" {
			bad = true
		}
		switch q.Subject {
		case SubjectPAI:
			rep.PAI++
		case SubjectMath:
			rep.Math++
		case SubjectEnglish:
			rep.English++
		case SubjectJawa:
			rep.Jawa++
		case "IPA", "IPS", "UMUM":
			// ok — bank dari PDF
		default:
			bad = true
		}
		switch q.Difficulty {
		case DiffEasy:
			rep.Easy++
		case DiffMedium:
			rep.Medium++
		case DiffHard:
			rep.Hard++
		default:
			bad = true
		}
		if q.Grade != GradeSD && q.Grade != GradeSMP && q.Grade != GradeSMA {
			bad = true
		}
		if seenID[q.ID] {
			rep.Duplicate++
			bad = true
		}
		seenID[q.ID] = true
		key := strings.ToLower(strings.TrimSpace(q.Question))
		if seenQ[key] {
			rep.Duplicate++
			bad = true
		}
		seenQ[key] = true
		rep.Total++
		if bad {
			rep.Invalid++
			if len(rep.Problems) < 12 {
				rep.Problems = append(rep.Problems, fmt.Sprintf("idx=%d id=%s", i, q.ID))
			}
		}
	}
	return rep
}

func (q EduQuestion) Public(playerID, username string, number int, endsAt int64, final bool) QuestionPublic {
	return QuestionPublic{
		ID: q.ID, Subject: q.Subject, Category: q.Category, Grade: q.Grade, Difficulty: q.Difficulty,
		Question: q.Question, OptionA: q.OptionA, OptionB: q.OptionB, OptionC: q.OptionC, OptionD: q.OptionD,
		Number: number, QuestionNo: number, TimeLimit: QUESTION_TIME_LIMIT_SEC, EndsAt: endsAt,
		PlayerID: playerID, Username: username, Final: final,
	}
}

func PenaltyPosition(from int) int {
	return PenaltyPositionN(from, LivePenaltyN())
}

func PenaltyPositionN(from, n int) int {
	if n < 1 {
		n = QUESTION_PENALTY
	}
	x := from - n
	if x < MIN_POSITION {
		return MIN_POSITION
	}
	return x
}

func PenaltyPath(from, to int) []int {
	if to > from {
		return []int{from, to}
	}
	out := make([]int, 0, from-to+1)
	for p := from; p >= to; p-- {
		out = append(out, p)
	}
	return out
}

func (b *EduBank) Get(id string) (EduQuestion, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	i, ok := b.byID[id]
	if !ok {
		return EduQuestion{}, false
	}
	return b.items[i], true
}

func (b *EduBank) bumpUsage(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if i, ok := b.byID[id]; ok {
		b.items[i].Usage++
	}
}

func (b *EduBank) List(subject, difficulty, q string, includeInactive bool) []EduQuestion {
	b.mu.RLock()
	defer b.mu.RUnlock()
	q = strings.ToLower(strings.TrimSpace(q))
	out := make([]EduQuestion, 0)
	for _, it := range b.items {
		if it.Deleted {
			continue
		}
		if !includeInactive && !it.Active {
			continue
		}
		if subject != "" && it.Subject != subject {
			continue
		}
		if difficulty != "" && it.Difficulty != difficulty {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(it.Question), q) && !strings.Contains(strings.ToLower(it.ID), q) {
			continue
		}
		cp := it
		cp.CorrectAnswer = ""
		cp.Explanation = ""
		out = append(out, cp)
	}
	return out
}

func (b *EduBank) ListAdmin(subject, difficulty, q string) []EduQuestion {
	b.mu.RLock()
	defer b.mu.RUnlock()
	q = strings.ToLower(strings.TrimSpace(q))
	out := make([]EduQuestion, 0)
	for _, it := range b.items {
		if subject != "" && it.Subject != subject {
			continue
		}
		if difficulty != "" && it.Difficulty != difficulty {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(it.Question), q) && !strings.Contains(strings.ToLower(it.ID), q) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func (b *EduBank) Upsert(q EduQuestion) error {
	q.CorrectAnswer = strings.ToUpper(strings.TrimSpace(q.CorrectAnswer))
	q.UpdatedAt = time.Now().UnixMilli()
	if q.CreatedAt == 0 {
		q.CreatedAt = q.UpdatedAt
	}
	tmp := append([]EduQuestion{}, b.items...)
	if i, ok := b.byID[q.ID]; ok {
		tmp[i] = q
	} else {
		tmp = append(tmp, q)
	}
	rep := validateItems(tmp)
	if rep.Invalid > 0 {
		return fmt.Errorf("invalid question")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = tmp
	b.byID = map[string]int{}
	for i := range b.items {
		b.byID[b.items[i].ID] = i
	}
	return nil
}

func (b *EduBank) Delete(id string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	i, ok := b.byID[id]
	if !ok {
		return false
	}
	b.items[i].Deleted = true
	b.items[i].Active = false
	b.items[i].UpdatedAt = time.Now().UnixMilli()
	return true
}

func (b *EduBank) SetActive(id string, active bool) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	i, ok := b.byID[id]
	if !ok {
		return false
	}
	b.items[i].Active = active
	b.items[i].UpdatedAt = time.Now().UnixMilli()
	return true
}

func (b *EduBank) SeedFromEmbedded() error {
	nb, err := LoadEduBankBytes(smaEduQuestionsJSON)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = nb.items
	b.byID = nb.byID
	return nil
}

func (b *EduBank) reserve(used []string, subject, difficulty, grade string, preferUnused bool) (EduQuestion, []string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	usedSet := map[string]bool{}
	for _, id := range used {
		usedSet[id] = true
	}
	grade = strings.ToUpper(strings.TrimSpace(grade))
	collect := func(strictUsed bool) []int {
		out := make([]int, 0, 32)
		for i, q := range b.items {
			if !q.Active || q.Deleted {
				continue
			}
			if grade != "" && q.Grade != grade {
				continue
			}
			if subject != "" && q.Subject != subject {
				continue
			}
			if difficulty != "" && q.Difficulty != difficulty {
				continue
			}
			if strictUsed && usedSet[q.ID] {
				continue
			}
			out = append(out, i)
		}
		return out
	}
	pick := func(strictUsed bool) (EduQuestion, bool) {
		cands := collect(strictUsed)
		if len(cands) == 0 {
			return EduQuestion{}, false
		}
		i := cands[rand.Intn(len(cands))]
		return b.items[i], true
	}
	if q, ok := pick(preferUnused); ok {
		u := append([]string{}, used...)
		if !usedSet[q.ID] {
			u = append(u, q.ID)
		} else if len(usedSet) >= len(b.items) {
			u = []string{q.ID}
		}
		return q, u
	}
	if preferUnused {
		if q, ok := pick(false); ok {
			return q, []string{q.ID}
		}
	}
	return EduQuestion{}, used
}

func questionStorePath() string {
	if p := os.Getenv("ULAR_QUESTION_STORE"); p != "" {
		return p
	}
	return ""
}

func normalizeGrade(g string) string {
	g = strings.ToUpper(strings.TrimSpace(g))
	if g == GradeSD {
		return GradeSD
	}
	return GradeSMA
}
