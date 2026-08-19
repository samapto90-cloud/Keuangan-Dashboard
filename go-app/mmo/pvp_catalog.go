package mmo

import (
	"log"
	"strings"

	_ "embed"
)

//go:embed data/pvpModes.json
var pvpModesJSON []byte

//go:embed data/pvpRanks.json
var pvpRanksJSON []byte

//go:embed data/pvpSeason.json
var pvpSeasonJSON []byte

//go:embed data/pvpModifiers.json
var pvpModifiersJSON []byte

//go:embed data/pvpShop.json
var pvpShopJSON []byte

//go:embed data/pvpRewards.json
var pvpRewardsJSON []byte

type PvpModeDef struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	TeamSize   int    `json:"teamSize"`
	MinLevel   int    `json:"minLevel"`
	Duration   int    `json:"duration"`
	Map        string `json:"map"`
	Respawn    bool   `json:"respawn"`
	RespawnSec int    `json:"respawnSec"`
	ScoreLimit int    `json:"scoreLimit"`
	Enabled    bool   `json:"enabled"`
}

type PvpRankDef struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Min    int    `json:"min"`
	Max    int    `json:"max"`
	Reward string `json:"reward"`
	Visual string `json:"visual"`
}

type PvpSeasonDef struct {
	ID        string  `json:"id"`
	Number    int     `json:"number"`
	Name      string  `json:"name"`
	Start     string  `json:"start"`
	End       string  `json:"end"`
	Weeks     int     `json:"weeks"`
	SoftReset float64 `json:"softReset"`
	Floor     int     `json:"floor"`
}

type PvpModifiers struct {
	DamageDealt    float64            `json:"damageDealt"`
	DamageTaken    float64            `json:"damageTaken"`
	SkillDamage    float64            `json:"skillDamage"`
	Ultimate       float64            `json:"ultimate"`
	LevelScale     float64            `json:"levelScale"`
	SpawnProtect   float64            `json:"spawnProtect"`
	Countdown      float64            `json:"countdown"`
	BgRespawn      float64            `json:"bgRespawn"`
	Reconnect      float64            `json:"reconnect"`
	AfkWarn        float64            `json:"afkWarn"`
	AfkPenalty     float64            `json:"afkPenalty"`
	WinRating      int                `json:"winRating"`
	LossRating     int                `json:"lossRating"`
	Placement      int                `json:"placement"`
	StartRating    int                `json:"startRating"`
	LeaveLock      float64            `json:"leaveLock"`
	ReadyTimeout   float64            `json:"readyTimeout"`
	EmoteCooldown  float64            `json:"emoteCooldown"`
	ScoreTick      int                `json:"scoreTick"`
	CaptureRate    float64            `json:"captureRate"`
	ScoreLimit     int                `json:"scoreLimit"`
	MvpSeasonXP    int                `json:"mvpSeasonXP"`
	RatingCap      int                `json:"ratingCap"`
	DisabledSkills []string           `json:"disabledSkills"`
	TransformPct   map[string]float64 `json:"transformPct"`
}

type PvpShopItem struct {
	ShopItemID string `json:"shopItemId"`
	ItemID     string `json:"itemId"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Price      int    `json:"price"`
	Currency   string `json:"currency"`
}

type PvpShopDef struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	NPC   string        `json:"npc"`
	Items []PvpShopItem `json:"items"`
}

type PvpRewardDef struct {
	ID     string `json:"id"`
	Rank   string `json:"rank"`
	Kind   string `json:"kind"`
	ItemID string `json:"itemId"`
	Name   string `json:"name"`
}

var (
	pvpModeCatalog []PvpModeDef
	pvpModeByID    = map[string]PvpModeDef{}
	pvpRankCatalog []PvpRankDef
	pvpSeasonDef   PvpSeasonDef
	pvpMod         PvpModifiers
	pvpShop        PvpShopDef
	pvpRewardList  []PvpRewardDef
)

func init() {
	mustJSON("pvpModes.json", pvpModesJSON, &pvpModeCatalog)
	for _, m := range pvpModeCatalog {
		pvpModeByID[m.ID] = m
	}
	mustJSON("pvpRanks.json", pvpRanksJSON, &pvpRankCatalog)
	mustJSON("pvpSeason.json", pvpSeasonJSON, &pvpSeasonDef)
	mustJSON("pvpModifiers.json", pvpModifiersJSON, &pvpMod)
	mustJSON("pvpShop.json", pvpShopJSON, &pvpShop)
	mustJSON("pvpRewards.json", pvpRewardsJSON, &pvpRewardList)
	if pvpMod.StartRating <= 0 {
		pvpMod.StartRating = 1000
	}
	if pvpMod.WinRating <= 0 {
		pvpMod.WinRating = 20
	}
	if pvpMod.LossRating <= 0 {
		pvpMod.LossRating = 15
	}
	if pvpMod.Placement <= 0 {
		pvpMod.Placement = 5
	}
	if pvpMod.SkillDamage <= 0 {
		pvpMod.SkillDamage = 0.85
	}
	if pvpMod.DamageDealt <= 0 {
		pvpMod.DamageDealt = 0.85
	}
	if pvpMod.RatingCap <= 0 {
		pvpMod.RatingCap = 99999
	}
	if pvpMod.ScoreLimit <= 0 {
		pvpMod.ScoreLimit = 1000
	}
	if len(pvpModeCatalog) == 0 {
		log.Print("pvp modes empty")
	}
}

func pvpMode(id string) (PvpModeDef, bool) {
	def, ok := pvpModeByID[strings.ToUpper(strings.TrimSpace(id))]
	if ok {
		return def, true
	}
	def, ok = pvpModeByID[id]
	return def, ok
}

func rankForRating(rating int) PvpRankDef {
	for _, r := range pvpRankCatalog {
		if rating >= r.Min && rating <= r.Max {
			return r
		}
	}
	if rating >= 3500 {
		return PvpRankDef{ID: "CELESTIAL", Name: "Celestial", Min: 3500, Max: 99999}
	}
	return PvpRankDef{ID: "BRONZE", Name: "Bronze", Min: 0, Max: 999}
}
