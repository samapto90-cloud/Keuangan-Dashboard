package mmo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type AttemptStore struct {
	mu   sync.Mutex
	path string
	all  []QuestionAttempt
}

func attemptStorePath() string {
	if p := os.Getenv("ULAR_ATTEMPT_STORE"); p != "" {
		return p
	}
	return filepath.Join("data", "ular-attempts.json")
}

func OpenAttemptStore(path string) *AttemptStore {
	s := &AttemptStore{path: path}
	raw, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(raw, &s.all)
	}
	return s
}

func (s *AttemptStore) Append(a QuestionAttempt) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.all = append(s.all, a)
	raw, _ := json.MarshalIndent(s.all, "", "  ")
	_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
	_ = os.WriteFile(s.path, raw, 0o644)
}

func (s *AttemptStore) All() []QuestionAttempt {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]QuestionAttempt, len(s.all))
	copy(out, s.all)
	return out
}

func (s *AttemptStore) ForUser(userID string) []QuestionAttempt {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]QuestionAttempt, 0)
	for _, a := range s.all {
		if a.UserID == userID {
			out = append(out, a)
		}
	}
	return out
}

func summarizeAttempts(atts []QuestionAttempt) map[string]any {
	total := len(atts)
	correct, wrong, timeout := 0, 0, 0
	type bucket struct{ t, c int }
	sub := map[string]*bucket{}
	for _, a := range atts {
		if sub[a.Subject] == nil {
			sub[a.Subject] = &bucket{}
		}
		sub[a.Subject].t++
		if a.Timeout {
			timeout++
		} else if a.Correct {
			correct++
			sub[a.Subject].c++
		} else {
			wrong++
		}
	}
	acc := 0
	if total > 0 {
		acc = int(float64(correct) / float64(total) * 100)
	}
	subject := map[string]any{}
	for _, name := range eduSubjects {
		b := sub[name]
		if b == nil {
			subject[name] = map[string]any{"correct": 0, "total": 0, "accuracy": 0}
			continue
		}
		pct := 0
		if b.t > 0 {
			pct = int(float64(b.c) / float64(b.t) * 100)
		}
		subject[name] = map[string]any{"correct": b.c, "total": b.t, "accuracy": pct}
	}
	return map[string]any{
		"totalQuestions":  total,
		"correctAnswers":  correct,
		"wrongAnswers":    wrong,
		"timeoutAnswers":  timeout,
		"accuracy":        acc,
		"subjectAccuracy": subject,
	}
}
