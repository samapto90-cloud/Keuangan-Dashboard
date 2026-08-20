package mmo

import "time"

const (
	XP_CORRECT_ANSWER = 10
	XP_WRONG_ANSWER   = 2
	XP_TIMEOUT        = 1
	XP_MATCH_COMPLETE = 20
	XP_WIN            = 50
	XP_ACHIEVEMENT    = 25
	MaxPlayerLevel    = 100

	COIN_MATCH       = 30
	COIN_WIN         = 120
	COIN_DAILY       = 50
	COIN_ACHIEVEMENT = 20

	MaxMatchHistoryPage = 20
)

var DailyCoinTable = []int{50, 75, 100, 125, 150, 200, 500}

const (
	XPCorrect     = "CORRECT_ANSWER"
	XPWrong       = "WRONG_ANSWER"
	XPTimeout     = "TIMEOUT"
	XPMatchDone   = "MATCH_COMPLETE"
	XPWin         = "WIN"
	XPAchievement = "ACHIEVEMENT"
	XPAdmin       = "ADMIN_ADJUSTMENT"

	CoinMatch       = "MATCH_REWARD"
	CoinWin         = "WIN_REWARD"
	CoinDaily       = "DAILY_REWARD"
	CoinAchievement = "ACHIEVEMENT_REWARD"
	CoinAdmin       = "ADMIN_ADJUSTMENT"
)

type PlayerProfile struct {
	ID               string         `json:"id"`
	UserID           string         `json:"userId"`
	Username         string         `json:"username"`
	Avatar           string         `json:"avatar"`
	Title            string         `json:"title"`
	Level            int            `json:"level"`
	XP               int            `json:"xp"`
	Coins            int            `json:"coins"`
	Trophies         int            `json:"trophies"`
	TotalMatches     int            `json:"totalMatches"`
	Wins             int            `json:"wins"`
	Losses           int            `json:"losses"`
	Draws            int            `json:"draws"`
	TotalQuestions   int            `json:"totalQuestions"`
	CorrectAnswers   int            `json:"correctAnswers"`
	WrongAnswers     int            `json:"wrongAnswers"`
	TimeoutAnswers   int            `json:"timeoutAnswers"`
	CurrentWinStreak int            `json:"currentWinStreak"`
	BestWinStreak    int            `json:"bestWinStreak"`
	DailyStreak      int            `json:"dailyStreak"`
	LastDailyDate    string         `json:"lastDailyDate,omitempty"`
	SubjectCorrect   map[string]int `json:"subjectCorrect,omitempty"`
	SubjectTotal     map[string]int `json:"subjectTotal,omitempty"`
	LadderClimbs     int            `json:"ladderClimbs"`
	SnakeHits        int            `json:"snakeHits"`
	Reached100       int            `json:"reached100"`
	RankTier         string         `json:"rankTier,omitempty"`
	RankDivision     string         `json:"rankDivision,omitempty"`
	RankRR           int            `json:"rankRr"`
	RankProtect      bool           `json:"rankProtect,omitempty"`
	PeakTier         string         `json:"peakTier,omitempty"`
	PeakDivision     string         `json:"peakDivision,omitempty"`
	PeakRR           int            `json:"peakRr,omitempty"`
	PeakIndex        int            `json:"peakIndex,omitempty"`
	SeasonID         string         `json:"seasonId,omitempty"`
	RankedWins       int            `json:"rankedWins"`
	AbandonCount     int            `json:"abandonCount"`
	RestrictUntil    int64          `json:"restrictUntil,omitempty"`
	AllowFriends     bool           `json:"allowFriends"`
	AllowInvites     bool           `json:"allowInvites"`
	ShowOnline       bool           `json:"showOnline"`
	CreatedAt        int64          `json:"createdAt"`
	UpdatedAt        int64          `json:"updatedAt"`
}

type XPTransaction struct {
	ID            string `json:"id"`
	UserID        string `json:"userId"`
	Type          string `json:"type"`
	Amount        int    `json:"amount"`
	BalanceBefore int    `json:"balanceBefore"`
	BalanceAfter  int    `json:"balanceAfter"`
	ReferenceID   string `json:"referenceId,omitempty"`
	CreatedAt     int64  `json:"createdAt"`
}

type CoinTransaction struct {
	ID            string `json:"id"`
	UserID        string `json:"userId"`
	Type          string `json:"type"`
	Amount        int    `json:"amount"`
	BalanceBefore int    `json:"balanceBefore"`
	BalanceAfter  int    `json:"balanceAfter"`
	ReferenceID   string `json:"referenceId,omitempty"`
	AdminID       string `json:"adminId,omitempty"`
	Reason        string `json:"reason,omitempty"`
	CreatedAt     int64  `json:"createdAt"`
}

type UlarAchievement struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Requirement string `json:"requirement,omitempty"`
	RewardXP    int    `json:"rewardXp"`
	RewardCoins int    `json:"rewardCoins"`
	Active      bool   `json:"active"`
}

type UserAchievement struct {
	UserID        string `json:"userId"`
	AchievementID string `json:"achievementId"`
	UnlockedAt    int64  `json:"unlockedAt"`
}

type DailyClaim struct {
	UserID     string `json:"userId"`
	RewardDate string `json:"rewardDate"`
	Day        int    `json:"day"`
	Coins      int    `json:"coins"`
	CreatedAt  int64  `json:"createdAt"`
}

type MatchPlayerResult struct {
	MatchID     string `json:"matchId"`
	UserID      string `json:"userId"`
	Rank        int    `json:"rank"`
	Position    int    `json:"position"`
	Correct     int    `json:"correct"`
	Wrong       int    `json:"wrong"`
	Timeout     int    `json:"timeout"`
	XPEarned    int    `json:"xpEarned"`
	CoinsEarned int    `json:"coinsEarned"`
	Won         bool   `json:"won"`
	Mode        string `json:"mode,omitempty"`
	RRBefore    int    `json:"rrBefore,omitempty"`
	RRChange    int    `json:"rrChange,omitempty"`
	RRAfter     int    `json:"rrAfter,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
}

type RewardEvent struct {
	UserID       string   `json:"userId"`
	XP           int      `json:"xp"`
	Coins        int      `json:"coins"`
	Trophies     int      `json:"trophies,omitempty"`
	TrophyTotal  int      `json:"trophyTotal,omitempty"`
	Won          bool     `json:"won,omitempty"`
	LevelBefore  int      `json:"levelBefore"`
	LevelAfter   int      `json:"levelAfter"`
	LevelUp      bool     `json:"levelUp"`
	Achievements []string `json:"achievements,omitempty"`
	RankLabel    string   `json:"rankLabel,omitempty"`
	RR           int      `json:"rr,omitempty"`
	RRDelta      int      `json:"rrDelta,omitempty"`
	RankUp       bool     `json:"rankUp,omitempty"`
	RankDown     bool     `json:"rankDown,omitempty"`
}

type UlarTitle struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Requires string `json:"requires,omitempty"`
}

// LevelService: progressive XP thresholds. XPForLevel(n) = 25*(n-1)*(n+2) → 0, 100, 250, 450, …
func XPForLevel(level int) int {
	if level <= 1 {
		return 0
	}
	if level > MaxPlayerLevel {
		level = MaxPlayerLevel
	}
	return 25 * (level - 1) * (level + 2)
}

func LevelFromXP(xp int) int {
	if xp < 0 {
		xp = 0
	}
	lvl := 1
	for n := 2; n <= MaxPlayerLevel; n++ {
		if xp >= XPForLevel(n) {
			lvl = n
			continue
		}
		break
	}
	return lvl
}

func XPToNext(xp int) (level, into, need, nextAt int) {
	level = LevelFromXP(xp)
	cur := XPForLevel(level)
	if level >= MaxPlayerLevel {
		return level, xp - cur, 0, cur
	}
	nextAt = XPForLevel(level + 1)
	return level, xp - cur, nextAt - xp, nextAt
}

func JakartaDate(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	return t.In(loc).Format("2006-01-02")
}

func YesterdayJakarta(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	return t.In(loc).AddDate(0, 0, -1).Format("2006-01-02")
}

var AchievementCatalog = []UlarAchievement{
	{ID: "FIRST_GAME", Code: "FIRST_GAME", Name: "Pertandingan Pertama", Description: "Selesaikan pertandingan pertamamu.", Icon: "play", RewardXP: 20, RewardCoins: 20, Active: true},
	{ID: "WIN_FIRST", Code: "WIN_FIRST", Name: "Kemenangan Pertama", Description: "Menangkan satu pertandingan.", Icon: "trophy", RewardXP: 40, RewardCoins: 40, Active: true},
	{ID: "QUIZ_MASTER", Code: "QUIZ_MASTER", Name: "Quiz Master", Description: "Jawab 50 soal dengan benar.", Icon: "book", RewardXP: 80, RewardCoins: 50, Active: true},
	{ID: "MATH_MASTER", Code: "MATH_MASTER", Name: "Math Master", Description: "Jawab 25 soal matematika benar.", Icon: "hash", RewardXP: 60, RewardCoins: 40, Active: true},
	{ID: "PAI_MASTER", Code: "PAI_MASTER", Name: "PAI Master", Description: "Jawab 25 soal PAI benar.", Icon: "star", RewardXP: 60, RewardCoins: 40, Active: true},
	{ID: "ENGLISH_MASTER", Code: "ENGLISH_MASTER", Name: "English Master", Description: "Jawab 25 soal Bahasa Inggris benar.", Icon: "globe", RewardXP: 60, RewardCoins: 40, Active: true},
	{ID: "JAVA_MASTER", Code: "JAVA_MASTER", Name: "Jawa Master", Description: "Jawab 25 soal Bahasa Jawa benar.", Icon: "leaf", RewardXP: 60, RewardCoins: 40, Active: true},
	{ID: "CLIMBER", Code: "CLIMBER", Name: "Pendaki", Description: "Naik tangga 10 kali.", Icon: "ladder", RewardXP: 40, RewardCoins: 30, Active: true},
	{ID: "SNAKE_SURVIVOR", Code: "SNAKE_SURVIVOR", Name: "Penyintas Ular", Description: "Terkena ular 10 kali.", Icon: "snake", RewardXP: 40, RewardCoins: 30, Active: true},
	{ID: "WIN_STREAK_3", Code: "WIN_STREAK_3", Name: "Hat-trick", Description: "Menang 3 kali berturut-turut.", Icon: "fire", RewardXP: 70, RewardCoins: 50, Active: true},
	{ID: "WIN_STREAK_5", Code: "WIN_STREAK_5", Name: "Tak Terkalahkan", Description: "Menang 5 kali berturut-turut.", Icon: "crown", RewardXP: 120, RewardCoins: 80, Active: true},
	{ID: "CENTURY", Code: "CENTURY", Name: "Seratus", Description: "Mencapai kotak 100.", Icon: "flag", RewardXP: 50, RewardCoins: 40, Active: true},
}

var TitleCatalog = []UlarTitle{
	{ID: "pemula", Name: "Pemula"},
	{ID: "pembelajar", Name: "Pembelajar", Requires: "FIRST_GAME"},
	{ID: "penjelajah", Name: "Penjelajah", Requires: "CENTURY"},
	{ID: "ahli", Name: "Ahli", Requires: "QUIZ_MASTER"},
	{ID: "master", Name: "Master", Requires: "WIN_STREAK_5"},
}

var ValidAvatars = []string{"bird", "batik", "gunung", "candi", "wayang", "kapal"}

func achievementByID(id string) (UlarAchievement, bool) {
	for _, a := range liveAchievements() {
		if a.ID == id {
			if !a.Active {
				return UlarAchievement{}, false
			}
			return a, true
		}
	}
	return UlarAchievement{}, false
}

func validAvatar(id string) bool {
	for _, a := range ValidAvatars {
		if a == id {
			return true
		}
	}
	return false
}

func validTitle(id string) bool {
	for _, t := range TitleCatalog {
		if t.ID == id {
			return true
		}
	}
	return false
}
