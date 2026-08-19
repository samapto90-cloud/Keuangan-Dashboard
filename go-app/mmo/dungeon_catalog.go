package mmo

import (
	"log"

	_ "embed"
)

//go:embed data/chapters.json
var chaptersJSON []byte

//go:embed data/dungeons.json
var dungeonsJSON []byte

//go:embed data/bosses.json
var bossesJSON []byte

//go:embed data/bossSkills.json
var bossSkillsJSON []byte

//go:embed data/lootTables.json
var lootTablesJSON []byte

//go:embed data/dungeonWaves.json
var dungeonWavesJSON []byte

//go:embed data/dungeonObjectives.json
var dungeonObjectivesJSON []byte

type ChapterDef struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Region          string    `json:"region"`
	Story           string    `json:"story"`
	BossID          string    `json:"bossId"`
	BossName        string    `json:"bossName"`
	RequiredLevel   int       `json:"requiredLevel"`
	RequiredChapter string    `json:"requiredChapter"`
	Reward          RewardDef `json:"reward"`
	DungeonID       string    `json:"dungeonId"`
}

type DungeonDef struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	ChapterID         string    `json:"chapterId"`
	Kind              string    `json:"kind"`
	Description       string    `json:"description"`
	Difficulty        string    `json:"difficulty"`
	MinimumLevel      int       `json:"minimumLevel"`
	RecommendedLevel  int       `json:"recommendedLevel"`
	MinPlayers        int       `json:"minPlayers"`
	MaxPlayers        int       `json:"maxPlayers"`
	TimeLimit         int       `json:"timeLimit"`
	Environment       string    `json:"environment"`
	EnemyWaves        []string  `json:"enemyWaves"`
	BossID            string    `json:"bossId"`
	Bosses            []string  `json:"bosses"`
	LootTableID       string    `json:"lootTableId"`
	WeeklyLockout     bool      `json:"weeklyLockout"`
	StoryChapter      string    `json:"storyChapter"`
	Rewards           RewardDef `json:"rewards"`
	EducationQuestion string    `json:"educationQuestion"`
	EducationBoss     string    `json:"educationBoss"`
	Region            string    `json:"region"`
}

type BossPhaseDef struct {
	ID    int     `json:"id"`
	HPPct float64 `json:"hpPct"`
	Label string  `json:"label"`
}

type BossDef struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Title       string         `json:"title"`
	Type        string         `json:"type"`
	Level       int            `json:"level"`
	MaxHP       int            `json:"maxHp"`
	Attack      int            `json:"attack"`
	Defense     int            `json:"defense"`
	Speed       float64        `json:"speed"`
	Leash       float64        `json:"leash"`
	EnrageTime  float64        `json:"enrageTime"`
	LootTableID string         `json:"lootTableId"`
	Skills      []string       `json:"skills"`
	Phases      []BossPhaseDef `json:"phases"`
}

type BossSkillDef struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Damage        int     `json:"damage"`
	Radius        float64 `json:"radius"`
	Range         float64 `json:"range"`
	Telegraph     float64 `json:"telegraph"`
	Cooldown      float64 `json:"cooldown"`
	Phase         int     `json:"phase"`
	VFX           string  `json:"vfx"`
	SummonID      string  `json:"summonId"`
	SummonCount   int     `json:"summonCount"`
	Shape         string  `json:"shape"`
	Interruptible bool    `json:"interruptible"`
}

type LootEntryDef struct {
	ItemID      string  `json:"itemId"`
	MinQuantity int     `json:"minQuantity"`
	MaxQuantity int     `json:"maxQuantity"`
	Chance      float64 `json:"chance"`
}

type LootTableDef struct {
	ID       string         `json:"id"`
	RollType string         `json:"rollType"`
	Entries  []LootEntryDef `json:"entries"`
}

type WaveSpawnDef struct {
	EnemyID string  `json:"enemyId"`
	Count   int     `json:"count"`
	X       float64 `json:"x"`
	Z       float64 `json:"z"`
}

type DungeonWaveDef struct {
	ID        string         `json:"id"`
	DungeonID string         `json:"dungeonId"`
	Index     int            `json:"index"`
	Name      string         `json:"name"`
	Spawns    []WaveSpawnDef `json:"spawns"`
}

type DungeonObjectiveDef struct {
	ID        string `json:"id"`
	DungeonID string `json:"dungeonId"`
	Index     int    `json:"index"`
	Type      string `json:"type"`
	Target    string `json:"target"`
	Count     int    `json:"count"`
	Text      string `json:"text"`
}

var (
	chapterCatalog  []ChapterDef
	chapterByID     = map[string]ChapterDef{}
	dungeonCatalog  []DungeonDef
	dungeonByID     = map[string]DungeonDef{}
	bossCatalog     []BossDef
	bossByID        = map[string]BossDef{}
	bossSkillByID   = map[string]BossSkillDef{}
	lootTableByID   = map[string]LootTableDef{}
	dungeonWaveByID = map[string]DungeonWaveDef{}
	dungeonObjByDun = map[string][]DungeonObjectiveDef{}
)

func init() {
	mustJSON("chapters.json", chaptersJSON, &chapterCatalog)
	for _, c := range chapterCatalog {
		chapterByID[c.ID] = c
	}
	mustJSON("dungeons.json", dungeonsJSON, &dungeonCatalog)
	for _, d := range dungeonCatalog {
		dungeonByID[d.ID] = d
	}
	mustJSON("bosses.json", bossesJSON, &bossCatalog)
	for _, b := range bossCatalog {
		bossByID[b.ID] = b
	}
	var skills []BossSkillDef
	mustJSON("bossSkills.json", bossSkillsJSON, &skills)
	for _, s := range skills {
		bossSkillByID[s.ID] = s
	}
	var tables []LootTableDef
	mustJSON("lootTables.json", lootTablesJSON, &tables)
	for _, t := range tables {
		lootTableByID[t.ID] = t
	}
	var waves []DungeonWaveDef
	mustJSON("dungeonWaves.json", dungeonWavesJSON, &waves)
	for _, w := range waves {
		dungeonWaveByID[w.ID] = w
	}
	var objs []DungeonObjectiveDef
	mustJSON("dungeonObjectives.json", dungeonObjectivesJSON, &objs)
	for _, o := range objs {
		dungeonObjByDun[o.DungeonID] = append(dungeonObjByDun[o.DungeonID], o)
	}
	log.Printf("mmo chapters=%d dungeons=%d bosses=%d", len(chapterCatalog), len(dungeonCatalog), len(bossCatalog))
}
