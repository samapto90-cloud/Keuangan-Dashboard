package mmo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	SanctionWarning  = "WARNING"
	SanctionChatMute = "CHAT_MUTE"
	SanctionTempBan  = "TEMP_BAN"
	SanctionPermBan  = "PERMANENT_BAN"

	ReportOpen     = "OPEN"
	ReportReview   = "UNDER_REVIEW"
	ReportResolved = "RESOLVED"
	ReportDismiss  = "DISMISSED"
)

type UserSanction struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	IssuedBy  string `json:"issuedBy"`
	StartAt   int64  `json:"startAt"`
	EndAt     int64  `json:"endAt,omitempty"`
	Active    bool   `json:"active"`
	CreatedAt int64  `json:"createdAt"`
}

type AuditLog struct {
	ID         string `json:"id"`
	AdminID    string `json:"adminId"`
	AdminName  string `json:"adminName,omitempty"`
	Action     string `json:"action"`
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	BeforeData string `json:"beforeData,omitempty"`
	AfterData  string `json:"afterData,omitempty"`
	IPHash     string `json:"ipHash,omitempty"`
	CreatedAt  int64  `json:"createdAt"`
}

type GameConfig struct {
	Version            int     `json:"version"`
	QuestionTimeLimit  int     `json:"questionTimeLimit"`
	WrongAnswerPenalty int     `json:"wrongAnswerPenalty"`
	MaxPlayers         int     `json:"maxPlayers"`
	MinPlayers         int     `json:"minPlayers"`
	ReconnectGrace     int     `json:"reconnectGrace"`
	MatchmakingTimeout int     `json:"matchmakingTimeout"`
	XPCorrect          int     `json:"xpCorrect"`
	XPWrong            int     `json:"xpWrong"`
	XPTimeout          int     `json:"xpTimeout"`
	XPMatchComplete    int     `json:"xpMatchComplete"`
	XPWin              int     `json:"xpWin"`
	CoinMatch          int     `json:"coinMatch"`
	CoinWin            int     `json:"coinWin"`
	CoinAchievement    int     `json:"coinAchievement"`
	RankWinRR          int     `json:"rankWinRr"`
	RankLossRR         int     `json:"rankLossRr"`
	ReviewAccuracy     float64 `json:"reviewAccuracy"`
	DailyCoins         []int   `json:"dailyCoins,omitempty"`
	DailyXP            []int   `json:"dailyXp,omitempty"`
	CreatedAt          int64   `json:"createdAt"`
	CreatedBy          string  `json:"createdBy,omitempty"`
}

type FeatureFlags struct {
	EnableRanked      bool `json:"enableRanked"`
	EnableDailyReward bool `json:"enableDailyReward"`
	EnableChat        bool `json:"enableChat"`
	EnableNewBoard    bool `json:"enableNewBoard"`
}

type OpsSeason struct {
	RankSeason
	EndedBy string `json:"endedBy,omitempty"`
}

type OpsError struct {
	Time    int64  `json:"time"`
	Service string `json:"service"`
	Type    string `json:"type"`
	Count   int    `json:"count"`
}

type opsBlob struct {
	Sanctions    []UserSanction    `json:"sanctions"`
	Audit        []AuditLog        `json:"audit"`
	Configs      []GameConfig      `json:"configs"`
	ActiveVer    int               `json:"activeVersion"`
	Flags        FeatureFlags      `json:"flags"`
	FlagsSet     bool              `json:"flagsSet,omitempty"`
	Seasons      []OpsSeason       `json:"seasons"`
	Achievements []UlarAchievement `json:"achievements,omitempty"`
	Errors       []OpsError        `json:"errors,omitempty"`
}

type OpsStore struct {
	mu   sync.Mutex
	path string
	blob opsBlob
}

func opsStorePath() string {
	if p := strings.TrimSpace(os.Getenv("ULAR_OPS_STORE")); p != "" {
		return p
	}
	return filepath.Join("data", "ular-ops.json")
}

func DefaultGameConfig() GameConfig {
	return GameConfig{
		Version: 0, QuestionTimeLimit: QUESTION_TIME_LIMIT_SEC, WrongAnswerPenalty: QUESTION_PENALTY,
		MaxPlayers: MAX_PLAYERS, MinPlayers: 2, ReconnectGrace: DisconnectGrace, MatchmakingTimeout: QueueExpand2,
		XPCorrect: XP_CORRECT_ANSWER, XPWrong: XP_WRONG_ANSWER, XPTimeout: XP_TIMEOUT,
		XPMatchComplete: XP_MATCH_COMPLETE, XPWin: XP_WIN,
		CoinMatch: COIN_MATCH, CoinWin: COIN_WIN, CoinAchievement: COIN_ACHIEVEMENT,
		RankWinRR: RankWinRR, RankLossRR: RankLossRR, ReviewAccuracy: 25,
		DailyCoins: append([]int{}, DailyCoinTable...),
	}
}

func DefaultFlags() FeatureFlags {
	return FeatureFlags{EnableRanked: true, EnableDailyReward: true, EnableChat: true, EnableNewBoard: false}
}

func OpenOpsStore(path string) *OpsStore {
	s := &OpsStore{path: path, blob: opsBlob{Flags: DefaultFlags(), Seasons: []OpsSeason{{RankSeason: CurrentSeason}}}}
	raw, err := os.ReadFile(path)
	if err == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &s.blob)
	}
	if !s.blob.FlagsSet {
		s.blob.Flags = DefaultFlags()
	}
	if len(s.blob.Seasons) == 0 {
		s.blob.Seasons = []OpsSeason{{RankSeason: CurrentSeason}}
	}
	return s
}

func (s *OpsStore) flushLocked() error {
	raw, err := json.MarshalIndent(s.blob, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
	return os.WriteFile(s.path, raw, 0o644)
}

func (s *OpsStore) ActiveConfig() GameConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeLocked()
}

func (s *OpsStore) activeLocked() GameConfig {
	if s.blob.ActiveVer > 0 {
		for i := len(s.blob.Configs) - 1; i >= 0; i-- {
			if s.blob.Configs[i].Version == s.blob.ActiveVer {
				return s.blob.Configs[i]
			}
		}
	}
	return DefaultGameConfig()
}

func (s *OpsStore) Flags() FeatureFlags {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.blob.FlagsSet {
		return DefaultFlags()
	}
	return s.blob.Flags
}

func (s *OpsStore) SetFlags(f FeatureFlags) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blob.Flags = f
	s.blob.FlagsSet = true
	_ = s.flushLocked()
}

func (s *OpsStore) PutConfig(c GameConfig, by string) GameConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := 1
	for _, x := range s.blob.Configs {
		if x.Version >= next {
			next = x.Version + 1
		}
	}
	c.Version = next
	c.CreatedAt = time.Now().UnixMilli()
	c.CreatedBy = by
	s.blob.Configs = append(s.blob.Configs, c)
	s.blob.ActiveVer = next
	_ = s.flushLocked()
	return c
}

func (s *OpsStore) Rollback(to int) (GameConfig, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.blob.Configs {
		if c.Version == to {
			s.blob.ActiveVer = to
			_ = s.flushLocked()
			return c, ""
		}
	}
	return GameConfig{}, "versi tidak ada"
}

func (s *OpsStore) Configs() []GameConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]GameConfig, len(s.blob.Configs))
	copy(out, s.blob.Configs)
	return out
}

func (s *OpsStore) AddSanction(sn UserSanction) UserSanction {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	sn.ID = "sn-" + shortID()
	sn.CreatedAt = now
	if sn.StartAt == 0 {
		sn.StartAt = now
	}
	sn.Active = true
	s.blob.Sanctions = append(s.blob.Sanctions, sn)
	_ = s.flushLocked()
	return sn
}

func (s *OpsStore) DeactivateSanctions(userID, typ string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sn := range s.blob.Sanctions {
		if sn.UserID == userID && sn.Active && (typ == "" || sn.Type == typ) {
			s.blob.Sanctions[i].Active = false
		}
	}
	_ = s.flushLocked()
}

func (s *OpsStore) ActiveBan(userID string) *UserSanction {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	for i, sn := range s.blob.Sanctions {
		if sn.UserID != userID || !sn.Active {
			continue
		}
		if sn.Type != SanctionTempBan && sn.Type != SanctionPermBan {
			continue
		}
		if sn.Type == SanctionTempBan && sn.EndAt > 0 && now > sn.EndAt {
			s.blob.Sanctions[i].Active = false
			_ = s.flushLocked()
			continue
		}
		cp := s.blob.Sanctions[i]
		return &cp
	}
	return nil
}

func (s *OpsStore) ChatMuted(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	for i, sn := range s.blob.Sanctions {
		if sn.UserID != userID || !sn.Active || sn.Type != SanctionChatMute {
			continue
		}
		if sn.EndAt > 0 && now > sn.EndAt {
			s.blob.Sanctions[i].Active = false
			continue
		}
		return true
	}
	return false
}

func (s *OpsStore) SanctionsFor(userID string) []UserSanction {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]UserSanction, 0)
	for _, sn := range s.blob.Sanctions {
		if sn.UserID == userID {
			out = append(out, sn)
		}
	}
	return out
}

func (s *OpsStore) Audit(adminID, adminName, action, targetType, targetID, before, after, ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blob.Audit = append(s.blob.Audit, AuditLog{
		ID: "au-" + shortID(), AdminID: adminID, AdminName: adminName, Action: action,
		TargetType: targetType, TargetID: targetID, BeforeData: before, AfterData: after,
		IPHash: hashIP(ip), CreatedAt: time.Now().UnixMilli(),
	})
	if len(s.blob.Audit) > 5000 {
		s.blob.Audit = s.blob.Audit[len(s.blob.Audit)-4000:]
	}
	_ = s.flushLocked()
}

func (s *OpsStore) AuditPage(page, size int, action string) ([]AuditLog, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := make([]AuditLog, 0, len(s.blob.Audit))
	for i := len(s.blob.Audit) - 1; i >= 0; i-- {
		a := s.blob.Audit[i]
		if action != "" && a.Action != action {
			continue
		}
		all = append(all, a)
	}
	return pageSlice(all, page, size), len(all)
}

func (s *OpsStore) Seasons() []OpsSeason {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]OpsSeason, len(s.blob.Seasons))
	copy(out, s.blob.Seasons)
	return out
}

func (s *OpsStore) ActiveSeason() RankSeason {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, se := range s.blob.Seasons {
		if se.Active {
			return se.RankSeason
		}
	}
	return CurrentSeason
}

func (s *OpsStore) UpsertSeason(se OpsSeason) (OpsSeason, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if se.ID == "" {
		se.ID = "season-" + shortID()
	}
	found := false
	for i, x := range s.blob.Seasons {
		if x.ID == se.ID {
			s.blob.Seasons[i] = se
			found = true
			break
		}
	}
	if !found {
		s.blob.Seasons = append(s.blob.Seasons, se)
	}
	if se.Active {
		for i := range s.blob.Seasons {
			if s.blob.Seasons[i].ID != se.ID {
				s.blob.Seasons[i].Active = false
			}
		}
		CurrentSeason = se.RankSeason
	}
	n := 0
	for _, x := range s.blob.Seasons {
		if x.Active {
			n++
		}
	}
	if n > 1 {
		return se, "hanya satu season aktif"
	}
	_ = s.flushLocked()
	return se, ""
}

func (s *OpsStore) RecordError(service, typ string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	for i, e := range s.blob.Errors {
		if e.Service == service && e.Type == typ && now-e.Time < 3600*1000 {
			s.blob.Errors[i].Count++
			s.blob.Errors[i].Time = now
			_ = s.flushLocked()
			return
		}
	}
	s.blob.Errors = append(s.blob.Errors, OpsError{Time: now, Service: service, Type: typ, Count: 1})
	if len(s.blob.Errors) > 200 {
		s.blob.Errors = s.blob.Errors[len(s.blob.Errors)-150:]
	}
	_ = s.flushLocked()
}

func (s *OpsStore) Errors() []OpsError {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]OpsError, len(s.blob.Errors))
	copy(out, s.blob.Errors)
	return out
}

func (s *OpsStore) OverlayAchievements(items []UlarAchievement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blob.Achievements = items
	_ = s.flushLocked()
}

func (s *OpsStore) AchievementOverlay() []UlarAchievement {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]UlarAchievement, len(s.blob.Achievements))
	copy(out, s.blob.Achievements)
	return out
}

func hashIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:8])
}

func pageSlice[T any](all []T, page, size int) []T {
	if size <= 0 {
		size = 20
	}
	if size > 50 {
		size = 50
	}
	if page < 0 {
		page = 0
	}
	start := page * size
	if start >= len(all) {
		return []T{}
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}
	return all[start:end]
}

func liveOps() *OpsStore {
	if DefaultHub != nil {
		return DefaultHub.Ops
	}
	return nil
}

func LiveConfig() GameConfig {
	if o := liveOps(); o != nil {
		return o.ActiveConfig()
	}
	return DefaultGameConfig()
}

func LiveFlags() FeatureFlags {
	if o := liveOps(); o != nil {
		return o.Flags()
	}
	return DefaultFlags()
}

func LiveQuestionTime() time.Duration {
	n := LiveConfig().QuestionTimeLimit
	if n < 5 {
		n = QUESTION_TIME_LIMIT_SEC
	}
	return time.Duration(n) * time.Second
}

func LivePenaltyN() int {
	n := LiveConfig().WrongAnswerPenalty
	if n < 1 {
		return QUESTION_PENALTY
	}
	return n
}

func LiveRankWin() int {
	n := LiveConfig().RankWinRR
	if n <= 0 {
		return RankWinRR
	}
	return n
}

func LiveRankLoss() int {
	n := LiveConfig().RankLossRR
	if n <= 0 {
		return RankLossRR
	}
	return n
}

func LiveDailyCoins() []int {
	c := LiveConfig().DailyCoins
	if len(c) == 7 {
		for _, n := range c {
			if n < 0 {
				return DailyCoinTable
			}
		}
		return c
	}
	return DailyCoinTable
}

func liveOr(n, fallback int) int {
	if n > 0 {
		return n
	}
	return fallback
}

func liveReward(n, fallback int) int {
	if LiveConfig().Version > 0 {
		if n < 0 {
			return 0
		}
		return n
	}
	return fallback
}

func clipJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	s := string(raw)
	if len(s) > 4000 {
		return s[:4000]
	}
	return s
}
