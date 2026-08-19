package mmo

import (
	"encoding/json"
	"log"

	_ "embed"
)

//go:embed data/npcs.json
var npcsJSON []byte

//go:embed data/quests.json
var questsJSON []byte

//go:embed data/questions.json
var questionsJSON []byte

//go:embed data/interactions.json
var interactionsJSON []byte

//go:embed data/dialogues.json
var dialoguesJSON []byte

//go:embed data/storyFlags.json
var storyFlagsJSON []byte

//go:embed data/areas.json
var areasJSON []byte

type NPCDef struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Role             string            `json:"role"`
	Type             string            `json:"type"`
	X                float64           `json:"x"`
	Y                float64           `json:"y"`
	Z                float64           `json:"z"`
	Yaw              float64           `json:"yaw"`
	DialogueID       string            `json:"dialogueId"`
	QuestIDs         []string          `json:"questIds"`
	ShopID           string            `json:"shopId"`
	InteractionRange float64           `json:"interactionRange"`
	Schedule         map[string]string `json:"schedule"`
	Occupation       string            `json:"occupation"`
	Personality      string            `json:"personality"`
	Description      string            `json:"description"`
	VoiceLineID      string            `json:"voiceLineId"`
	Faction          string            `json:"faction"`
	MountID          string            `json:"mountId"`
	Region           string            `json:"region"`
	Trait            string            `json:"trait"`
	DialogueProfile  string            `json:"dialogueProfile"`
	QuestProfile     string            `json:"questProfile"`
}

type ObjectiveDef struct {
	Type   string `json:"type"`
	Target string `json:"target"`
	Count  int    `json:"count"`
	Text   string `json:"text"`
}

type RewardDef struct {
	Exp         int `json:"exp"`
	Coin        int `json:"coin"`
	Crystal     int `json:"crystal"`
	Potion      int `json:"potion"`
	EduToken    int `json:"eduToken"`
	BattleToken int `json:"battleToken"`
	Knowledge   int `json:"knowledge"`
}

type QuestDef struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Kind         string         `json:"kind"`
	NPC          string         `json:"npc"`
	Location     string         `json:"location"`
	Description  string         `json:"description"`
	Prereq       []string       `json:"prereq"`
	Objectives   []ObjectiveDef `json:"objectives"`
	Rewards      RewardDef      `json:"rewards"`
	FlagsOnClaim []string       `json:"flagsOnClaim"`
	UnlockForest bool           `json:"unlockForest"`
	ClaimAt      string         `json:"claimAt"`
	PerfectBonus *RewardDef     `json:"perfectBonus"`
	CinematicID  string         `json:"cinematicId"`
	StoryChapter string         `json:"storyChapter"`
}

type QuestionDef struct {
	ID       string   `json:"id"`
	Category string   `json:"category"`
	Prompt   string   `json:"prompt"`
	Choices  []string `json:"choices"`
	Correct  int      `json:"correct"`
	Explain  string   `json:"explain"`
	Grade    int      `json:"grade"`
}

type InteractDef struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	X          float64        `json:"x"`
	Z          float64        `json:"z"`
	Text       string         `json:"text"`
	Loot       map[string]int `json:"loot"`
	QuestionID string         `json:"questionId"`
}

type AreaDef struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Region   string  `json:"region"`
	X        float64 `json:"x"`
	Z        float64 `json:"z"`
	Radius   float64 `json:"radius"`
	Locked   bool    `json:"locked"`
	Kind     string  `json:"kind"`
	LevelMin int     `json:"levelMin"`
	LevelMax int     `json:"levelMax"`
	Hidden   bool    `json:"hidden"`
}

type DialogueLine struct {
	Speaker     string           `json:"speaker"`
	Text        string           `json:"text"`
	VoiceLineID string           `json:"voiceLineId,omitempty"`
	Choices     []DialogueChoice `json:"choices,omitempty"`
}

type DialogueChoice struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

var (
	npcCatalog      []NPCDef
	npcByID         = map[string]NPCDef{}
	questCatalog    []QuestDef
	questByID       = map[string]QuestDef{}
	questionCatalog []QuestionDef
	questionByID    = map[string]QuestionDef{}
	interactCatalog []InteractDef
	interactByID    = map[string]InteractDef{}
	dialogueCatalog = map[string]DialogueLine{}
	defaultFlags    = map[string]bool{}
	areaCatalog     []AreaDef
)

func init() {
	mustJSON("npcs.json", npcsJSON, &npcCatalog)
	for _, n := range npcCatalog {
		npcByID[n.ID] = n
	}
	mustJSON("quests.json", questsJSON, &questCatalog)
	for _, q := range questCatalog {
		questByID[q.ID] = q
	}
	mustJSON("questions.json", questionsJSON, &questionCatalog)
	for _, q := range questionCatalog {
		questionByID[q.ID] = q
	}
	mustJSON("interactions.json", interactionsJSON, &interactCatalog)
	for _, o := range interactCatalog {
		interactByID[o.ID] = o
	}
	if err := json.Unmarshal(dialoguesJSON, &dialogueCatalog); err != nil {
		log.Printf("mmo dialogues.json: %v", err)
	}
	if err := json.Unmarshal(storyFlagsJSON, &defaultFlags); err != nil {
		log.Printf("mmo storyFlags.json: %v", err)
	}
	mustJSON("areas.json", areasJSON, &areaCatalog)
}

func mustJSON(name string, raw []byte, dest any) {
	if err := json.Unmarshal(raw, dest); err != nil {
		log.Printf("mmo %s: %v", name, err)
	}
}

func npcSnap(n NPCDef) NPCSnapshot {
	return NPCSnapshot{ID: n.ID, Name: n.Name, Role: n.Role, Type: n.Type, X: n.X, Y: n.Y, Z: n.Z, Yaw: n.Yaw, VoiceLineID: n.VoiceLineID}
}

func objectSnap(o InteractDef) ObjectSnapshot {
	return ObjectSnapshot{ID: o.ID, Kind: o.Kind, X: o.X, Z: o.Z, Text: o.Text}
}
