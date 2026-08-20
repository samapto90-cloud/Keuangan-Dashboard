package mmo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type MatchStore struct {
	mu   sync.Mutex
	path string
	all  []StoredMatch
}

func matchStorePath() string {
	if p := os.Getenv("ULAR_MATCH_STORE"); p != "" {
		return p
	}
	return filepath.Join("data", "ular-matches.json")
}

func OpenMatchStore(path string) *MatchStore {
	s := &MatchStore{path: path}
	raw, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(raw, &s.all)
	}
	return s
}

func (s *MatchStore) Append(m StoredMatch) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.all = append(s.all, m)
	raw, _ := json.MarshalIndent(s.all, "", "  ")
	_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
	_ = os.WriteFile(s.path, raw, 0o644)
}

func (s *MatchStore) All() []StoredMatch {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]StoredMatch, len(s.all))
	copy(out, s.all)
	return out
}
