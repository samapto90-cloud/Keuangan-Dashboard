package mmo

import (
	"strings"
	"sync"
	"time"
)

type ularLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func newUlarLimiter() *ularLimiter {
	return &ularLimiter{hits: map[string][]time.Time{}}
}

func (l *ularLimiter) allow(key string, max int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cut := now.Add(-window)
	rows := l.hits[key]
	kept := rows[:0]
	for _, t := range rows {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= max {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}

func clampChat(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > ChatMaxLen {
		s = s[:ChatMaxLen]
	}
	return s
}

func allowedEmote(s string) string {
	switch s {
	case "clap", "lol", "fire", "wow", "gg", "👏", "😂", "🔥", "😮", "GG":
		return s
	default:
		return ""
	}
}
