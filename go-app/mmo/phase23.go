package mmo

import (
	"math"
	"strings"
	"time"
)

// Phase 23 overlay: OpenWorld + WorldEvent + WorldBoss + Dungeon + Raid.
// Reuses WorldService, OpenWorldService, CharacterService, CombatService,
// EnemyService, BossService, QuestService, PartyService, GuildService,
// InventoryService, EquipmentService, TransformationService,
// AchievementService, NPCService, StoryService, EconomyService,
// InstanceService, DungeonService, RaidService. Do not add duplicates.
// Logical tables persist on catalogs / DungeonInstance / PlayerLog / LiveEvent:
// regions, waypoints, world_events, world_event_participants, world_bosses,
// world_boss_instances, dungeons, dungeon_instances, dungeon_participants,
// dungeon_checkpoints, raids, raid_instances, raid_participants,
// boss_encounters, boss_phases, boss_mechanics, boss_threat, instance_rewards,
// raid_tokens, dungeon_completions, world_boss_schedule.
// Indexes: regionId, playerId, eventId, instanceId, bossId, partyId, guildId.

const (
	WhispersDungeonID = "dun-whispers"
	WhispersBossID    = "penjaga-gua"
	GerbangRaidID     = "raid-gerbang-33"
	RajaSilumanID     = "raja-siluman"
	RatuBayanganID    = "ratu-bayangan"
	SeranganSilumanID = "serangan-siluman"
	WhispersEduID     = "q-add-5-3"
	WhispersFastSec   = 1200
)

type MapMarkerView struct {
	ID     string  `json:"id"`
	Kind   string  `json:"kind"`
	Name   string  `json:"name"`
	Region string  `json:"region,omitempty"`
	X      float64 `json:"x"`
	Z      float64 `json:"z"`
}

type WorldBossPreview struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Announce string `json:"announce,omitempty"`
	When     string `json:"when,omitempty"`
}

type OpenWorldView struct {
	Regions       []RegionView       `json:"regions"`
	Markers       []MapMarkerView    `json:"markers"`
	NextWorldBoss *WorldBossPreview  `json:"nextWorldBoss,omitempty"`
	Event         *WorldEventView    `json:"event,omitempty"`
	WorldBoss     *WorldBossView     `json:"worldBoss,omitempty"`
}

var worldBossByID = map[string]WorldBossDef{}

func init() {
	registerPhase23()
}

func registerPhase23() {
	overlayPhase23Regions()
	registerWorldBossCatalog()
	registerPhase23Events()
	registerPhase23Education()
	registerPhase23Dungeon()
	registerPhase23Raid()
	registerPhase23NPC()
	registerPhase23Achievements()
}

func overlayPhase23Regions() {
	type overlay struct {
		Title, Enemy, Resource, Gateway, Path, Teleport string
		Rec, Min                                        int
	}
	rows := map[string]overlay{
		"village":   {Title: "Dawn Village", Enemy: "T1", Resource: "T1", Gateway: "old-windmill", Path: "village-road", Teleport: "old-windmill", Rec: 1, Min: 1},
		"forest":    {Title: "Mistwood Forest", Enemy: "T2", Resource: "T2", Gateway: "forgotten-bridge", Path: "mist-trail", Teleport: "forest-waypoint", Rec: 8, Min: 5},
		"valley":    {Title: "Stone Valley", Enemy: "T3", Resource: "T3", Gateway: "stone-gate", Path: "valley-pass", Teleport: "stone-gate", Rec: 12, Min: 10},
		"plains":    {Title: "River of Light", Enemy: "T4", Resource: "T3", Gateway: "moon-shrine", Path: "river-path", Teleport: "moon-shrine", Rec: 22, Min: 20},
		"canyon":    {Title: "Crimson Plains", Enemy: "T4", Resource: "T4", Gateway: "ancient-tree", Path: "crimson-road", Teleport: "ancient-tree", Rec: 18, Min: 15},
		"ruins":     {Title: "Ancient Ruins", Enemy: "T5", Resource: "T5", Gateway: "ancient-library", Path: "ruin-road", Teleport: "ancient-library", Rec: 35, Min: 35},
		"temple":    {Title: "Celestial Mountains", Enemy: "T5", Resource: "T5", Gateway: "sky-tower", Path: "celestial-ascent", Teleport: "sky-tower", Rec: 28, Min: 25},
		"celestial": {Title: "Holy Horizon", Enemy: "T6", Resource: "T6", Gateway: "celestial-gate", Path: "horizon-road", Teleport: "celestial-gate", Rec: 45, Min: 45},
	}
	for i := range regionCatalog {
		r := regionCatalog[i]
		ov, ok := rows[r.ID]
		if !ok {
			continue
		}
		if r.Title == "" {
			r.Title = ov.Title
		}
		if r.RecommendedLevel < 1 {
			r.RecommendedLevel = ov.Rec
		}
		if r.LevelMin < 1 {
			r.LevelMin = ov.Min
		}
		r.EnemyTier = ov.Enemy
		r.ResourceTier = ov.Resource
		r.Gateway = ov.Gateway
		r.Path = ov.Path
		r.TeleportPoint = ov.Teleport
		regionCatalog[i] = r
		regionByID[r.ID] = r
	}
}

func registerWorldBossCatalog() {
	if worldBossDef.ID != "" {
		worldBossByID[worldBossDef.ID] = worldBossDef
	}
	registerWorldBossDef(WorldBossDef{
		ID: RatuBayanganID, Name: "RATU BAYANGAN", Title: "World Boss", Region: "forest",
		X: 2, Z: 36, Level: 20, MaxHP: 12000, MaxPlayers: 40,
		Announce: "RATU BAYANGAN muncul di Mistwood Forest!", DurationSec: 1200,
		Rewards: RewardDef{Exp: 160, Coin: 60, Crystal: 3},
	})
}

func registerPhase23Events() {
	registerEventDef(EventDef{
		ID: SeranganSilumanID, Name: "SERANGAN SILUMAN", Kind: "hunt", Region: "forest",
		Announce: "⚠ Serangan makhluk liar muncul di Mistwood!", AnnounceJV: "Ati-ati, Le! Ana siluman teka saka alas!",
		DurationSec: 180, GateHP: 200, Enemies: []string{"forest_fang", "shadow_imp"},
		Objective: "Kalahkan 20 musuh.", ObjectiveNeed: 20, X: 0, Z: 36,
		Rewards: RewardDef{Exp: 40, Coin: 16, EduToken: 1},
	})
	registerEventDef(EventDef{
		ID: "defend-pak-jaga", Name: "DEFEND PAK JAGA", Kind: "defend", Region: "forest",
		Announce: "Lindungi Pak Jaga di tepi alas!", AnnounceJV: "Jaga Pak Jaga, Le!",
		DurationSec: 150, GateHP: 220, Enemies: []string{"forest_fang"},
		Objective: "Lindungi NPC.", ObjectiveNeed: 1, X: 4, Z: 30,
		Rewards: RewardDef{Exp: 28, Coin: 12},
	})
	registerEventDef(EventDef{
		ID: "escort-cahaya", Name: "ESCORT CAHAYA", Kind: "escort", Region: "valley",
		Announce: "Kawal peziarah menuju Stone Valley.", AnnounceJV: "Kawal wong tuwa menyang lembah.",
		DurationSec: 140, Enemies: []string{"forest_fang"},
		Objective: "Kawal NPC menuju lokasi.", ObjectiveNeed: 1, X: 0, Z: 58,
		Rewards: RewardDef{Exp: 30, Coin: 14},
	})
	registerEventDef(EventDef{
		ID: "hunt-siluman-20", Name: "HUNT SILUMAN", Kind: "hunt", Region: "forest",
		Announce: "Buruan siluman di Mistwood.", AnnounceJV: "Burunen siluman ing alas.",
		DurationSec: 160, Enemies: []string{"shadow_imp"},
		Objective: "Kalahkan enemy tertentu.", ObjectiveNeed: 20, X: -2, Z: 34,
		Rewards: RewardDef{Exp: 32, Coin: 14},
	})
	registerEventDef(EventDef{
		ID: "collect-cahaya", Name: "COLLECT CAHAYA", Kind: "collect", Region: "plains",
		Announce: "Kumpulkan pecahan cahaya di River of Light.", AnnounceJV: "Kumpulna pecahan cahaya.",
		DurationSec: 120, Objective: "Kumpulkan resource/event item.", ObjectiveNeed: 8, X: 0, Z: 80,
		Rewards: RewardDef{Exp: 22, Coin: 10, EduToken: 1},
	})
	registerEventDef(EventDef{
		ID: "rescue-mbah", Name: "RESCUE MBAH", Kind: "rescue", Region: "forest",
		Announce: "Selamatkan Mbah dari musuh di hutan.", AnnounceJV: "Slametna Mbah saka musuh.",
		DurationSec: 150, Enemies: []string{"forest_fang"},
		Objective: "Selamatkan NPC dari musuh.", ObjectiveNeed: 1, X: 3, Z: 32,
		Rewards: RewardDef{Exp: 26, Coin: 12},
	})
	registerEventDef(EventDef{
		ID: "bayangan-attack", Name: "BOSS ATTACK", Kind: "boss-attack", Region: "forest",
		Announce: "Bayangan besar menyerang Mistwood.", AnnounceJV: "Bayangan gedhe nyerang alas.",
		DurationSec: 200, Enemies: []string{"shadow_imp"},
		Objective: "Hadapi serangan boss.", ObjectiveNeed: 1, X: 2, Z: 36,
		Rewards: RewardDef{Exp: 50, Coin: 20, Crystal: 1},
	})
}

func registerPhase23Education() {
	registerQuestion(QuestionDef{
		ID: WhispersEduID, Category: "Matematika",
		Prompt:  "5 + 3 = ?",
		Choices: []string{"7", "8", "9"}, Correct: 1,
		Explain: "5 ditambah 3 sama dengan 8. Perisai boss berkurang.", Grade: 1,
	})
	registerQuestion(QuestionDef{
		ID: "q-huruf-a", Category: "Bahasa Indonesia",
		Prompt:  "Huruf apakah ini: A ?",
		Choices: []string{"B", "A", "C"}, Correct: 1,
		Explain: "Itu huruf A.", Grade: 1,
	})
	registerQuestion(QuestionDef{
		ID: "q-werna-ijo", Category: "Bahasa Jawa",
		Prompt:  "Ijo iku werna apa?",
		Choices: []string{"Merah", "Hijau", "Biru"}, Correct: 1,
		Explain: "Ijo artinya hijau.", Grade: 1,
	})
	registerQuestion(QuestionDef{
		ID: "q-matahari", Category: "Pengetahuan umum",
		Prompt:  "Matahari terbit di?",
		Choices: []string{"Barat", "Timur", "Utara"}, Correct: 1,
		Explain: "Matahari terbit di timur.", Grade: 1,
	})
}

func registerPhase23Dungeon() {
	registerBossSkillDef(BossSkillDef{ID: "gua-melee", Name: "Cakar Gua", Type: "MELEE", Damage: 16, Range: 2.4, Telegraph: 0.5, Cooldown: 2.2, Phase: 1, VFX: "slash", Shape: "circle"})
	registerBossSkillDef(BossSkillDef{ID: "shadow-zone", Name: "Shadow Zone", Type: "AREA", Damage: 20, Radius: 3.6, Range: 9, Telegraph: 1.1, Cooldown: 7, Phase: 1, VFX: "shadow", Shape: "circle"})
	registerBossSkillDef(BossSkillDef{ID: "whisper-summon", Name: "Summoned Creatures", Type: "SUMMON", Radius: 0, Range: 6, Telegraph: 0.8, Cooldown: 14, Phase: 2, VFX: "summon", SummonID: "shadow_imp", SummonCount: 2, Shape: "circle"})
	registerBossSkillDef(BossSkillDef{ID: "gua-charge", Name: "Charge Attack", Type: "CHARGE", Damage: 26, Radius: 3.4, Range: 8, Telegraph: 1.4, Cooldown: 12, Phase: 2, VFX: "charge", Shape: "circle", Interruptible: true})
	registerBossSkillDef(BossSkillDef{ID: "gua-projectile", Name: "Whisper Bolt", Type: "PROJECTILE", Damage: 18, Radius: 1.2, Range: 10, Telegraph: 0.8, Cooldown: 6, Phase: 2, VFX: "bolt", Shape: "line"})
	registerBossSkillDef(BossSkillDef{ID: "gua-ultimate", Name: "Echo Burst", Type: "ULTIMATE", Damage: 28, Radius: 5.0, Range: 12, Telegraph: 1.5, Cooldown: 16, Phase: 3, VFX: "burst", Shape: "ring"})
	registerBossDef(BossDef{
		ID: WhispersBossID, Name: "PENJAGA GUA", Title: "Cave Warden", Type: "Original Echo",
		Level: 10, MaxHP: 900, Attack: 14, Defense: 8, Speed: 1.8, Leash: 18, EnrageTime: 240,
		LootTableID: "lt-whispers",
		Skills:      []string{"gua-melee", "shadow-zone", "whisper-summon", "gua-charge", "gua-projectile", "gua-ultimate"},
		Phases: []BossPhaseDef{
			{ID: 1, HPPct: 1.0, Label: "PHASE 1"},
			{ID: 2, HPPct: 0.6, Label: "PHASE 2"},
			{ID: 3, HPPct: 0.25, Label: "PHASE 3"},
		},
	})
	registerLootTableDef(LootTableDef{
		ID: "lt-whispers", RollType: "PERSONAL",
		Entries: []LootEntryDef{
			{ItemID: "potion_heal", MinQuantity: 1, MaxQuantity: 2, Chance: 1},
			{ItemID: "crystal_shard", MinQuantity: 1, MaxQuantity: 2, Chance: 0.7},
			{ItemID: "dawn_cap", MinQuantity: 1, MaxQuantity: 1, Chance: 0.2},
		},
	})
	registerWaveDef(DungeonWaveDef{ID: "wh-pack", DungeonID: WhispersDungeonID, Index: 1, Name: "ENTRY", Spawns: []WaveSpawnDef{{EnemyID: "forest_fang", Count: 4, X: -3, Z: 8}}})
	registerWaveDef(DungeonWaveDef{ID: "wh-bridge", DungeonID: WhispersDungeonID, Index: 2, Name: "BRIDGE", Spawns: []WaveSpawnDef{{EnemyID: "shadow_imp", Count: 3, X: 0, Z: 10}}})
	registerWaveDef(DungeonWaveDef{ID: "wh-elite", DungeonID: WhispersDungeonID, Index: 3, Name: "CHAMBER", Spawns: []WaveSpawnDef{{EnemyID: "elite_shadow_beast", Count: 1, X: 0, Z: 12}}})
	registerWaveDef(DungeonWaveDef{ID: "wh-boss", DungeonID: WhispersDungeonID, Index: 4, Name: "BOSS ROOM", Spawns: []WaveSpawnDef{{EnemyID: WhispersBossID, Count: 1, X: 0, Z: 20}}})
	registerObjectiveDef(DungeonObjectiveDef{ID: "wh-kill", DungeonID: WhispersDungeonID, Index: 0, Type: "KILL", Target: "", Count: 7, Text: "Kalahkan enemy waves."})
	registerObjectiveDef(DungeonObjectiveDef{ID: "wh-boss", DungeonID: WhispersDungeonID, Index: 1, Type: "BOSS", Target: WhispersBossID, Count: 1, Text: "Hadapi PENJAGA GUA."})
	registerDungeonDef(DungeonDef{
		ID: WhispersDungeonID, Name: "CAVE OF WHISPERS", Kind: "DUNGEON", Description: "Gua bisikan. Penjaga Gua menguji keberanian dengan zona bayangan.",
		Difficulty: "NORMAL", MinimumLevel: 5, RecommendedLevel: 8, MinPlayers: 1, MaxPlayers: 5, TimeLimit: 1800,
		Environment: "whispers", Region: "forest", EnemyWaves: []string{"wh-pack", "wh-bridge", "wh-elite", "wh-boss"},
		BossID: WhispersBossID, Bosses: []string{WhispersBossID}, LootTableID: "lt-whispers",
		Rewards: RewardDef{Exp: 420, Coin: 100, Crystal: 2, Knowledge: 1}, EducationBoss: WhispersEduID,
	})
}

func registerPhase23Raid() {
	registerBossSkillDef(BossSkillDef{ID: "siluman-melee", Name: "Ground Combat", Type: "MELEE", Damage: 22, Range: 2.8, Telegraph: 0.6, Cooldown: 2.4, Phase: 1, VFX: "slash", Shape: "circle"})
	registerBossSkillDef(BossSkillDef{ID: "siluman-summon", Name: "Summoning", Type: "SUMMON", Range: 7, Telegraph: 0.9, Cooldown: 14, Phase: 2, VFX: "summon", SummonID: "shadow_imp", SummonCount: 3, Shape: "circle"})
	registerBossSkillDef(BossSkillDef{ID: "arena-collapse", Name: "Arena Collapse", Type: "AREA", Damage: 28, Radius: 5.2, Range: 12, Telegraph: 1.4, Cooldown: 10, Phase: 3, VFX: "collapse", Shape: "ring"})
	registerBossSkillDef(BossSkillDef{ID: "final-attack", Name: "Final Attack", Type: "ULTIMATE", Damage: 34, Radius: 6.0, Range: 14, Telegraph: 1.6, Cooldown: 16, Phase: 4, VFX: "ultimate", Shape: "circle"})
	registerBossDef(BossDef{
		ID: RajaSilumanID, Name: "RAJA SILUMAN", Title: "Gerbang 33", Type: "Original Raid",
		Level: 28, MaxHP: 4200, Attack: 24, Defense: 16, Speed: 1.7, Leash: 24, EnrageTime: 420,
		LootTableID: "lt-gerbang-33",
		Skills:      []string{"siluman-melee", "siluman-summon", "arena-collapse", "final-attack", "charge_burst"},
		Phases: []BossPhaseDef{
			{ID: 1, HPPct: 1.0, Label: "PHASE 1"},
			{ID: 2, HPPct: 0.7, Label: "PHASE 2"},
			{ID: 3, HPPct: 0.4, Label: "PHASE 3"},
			{ID: 4, HPPct: 0.15, Label: "PHASE 4"},
		},
	})
	registerLootTableDef(LootTableDef{
		ID: "lt-gerbang-33", RollType: "PERSONAL",
		Entries: []LootEntryDef{
			{ItemID: "potion_heal", MinQuantity: 1, MaxQuantity: 2, Chance: 1},
			{ItemID: "crystal_shard", MinQuantity: 1, MaxQuantity: 3, Chance: 0.8},
			{ItemID: "dawn_cap", MinQuantity: 1, MaxQuantity: 1, Chance: 0.18},
		},
	})
	registerWaveDef(DungeonWaveDef{ID: "g33-ground", DungeonID: GerbangRaidID, Index: 1, Name: "Ground Combat", Spawns: []WaveSpawnDef{{EnemyID: RajaSilumanID, Count: 1, X: 0, Z: 20}}})
	registerWaveDef(DungeonWaveDef{ID: "g33-summon", DungeonID: GerbangRaidID, Index: 2, Name: "Summoning", Spawns: []WaveSpawnDef{{EnemyID: RajaSilumanID, Count: 1, X: 0, Z: 20}}})
	registerWaveDef(DungeonWaveDef{ID: "g33-collapse", DungeonID: GerbangRaidID, Index: 3, Name: "Arena Collapse", Spawns: []WaveSpawnDef{{EnemyID: RajaSilumanID, Count: 1, X: 0, Z: 20}}})
	registerWaveDef(DungeonWaveDef{ID: "g33-final", DungeonID: GerbangRaidID, Index: 4, Name: "Final Attack", Spawns: []WaveSpawnDef{{EnemyID: RajaSilumanID, Count: 1, X: 0, Z: 20}}})
	registerObjectiveDef(DungeonObjectiveDef{ID: "g33-e1", DungeonID: GerbangRaidID, Index: 0, Type: "BOSS", Target: RajaSilumanID, Count: 1, Text: "Melewati area dan kalahkan elite."})
	registerObjectiveDef(DungeonObjectiveDef{ID: "g33-e2", DungeonID: GerbangRaidID, Index: 1, Type: "BOSS", Target: RajaSilumanID, Count: 1, Text: "Aktifkan seal."})
	registerObjectiveDef(DungeonObjectiveDef{ID: "g33-e3", DungeonID: GerbangRaidID, Index: 2, Type: "BOSS", Target: RajaSilumanID, Count: 1, Text: "Lindungi NPC, dodge serangan."})
	registerObjectiveDef(DungeonObjectiveDef{ID: "g33-e4", DungeonID: GerbangRaidID, Index: 3, Type: "BOSS", Target: RajaSilumanID, Count: 1, Text: "Hadapi RAJA SILUMAN."})
	registerDungeonDef(DungeonDef{
		ID: GerbangRaidID, Name: "GERBANG 33 SILUMAN", Kind: "RAID", Description: "Raid 8 pemain. Gerbang 33 siluman. Desain original.",
		Difficulty: "NORMAL", MinimumLevel: 8, RecommendedLevel: 20, MinPlayers: 8, MaxPlayers: 8, TimeLimit: 2400,
		Environment: "siluman-gate", Region: "forest", EnemyWaves: []string{"g33-ground", "g33-summon", "g33-collapse", "g33-final"},
		BossID: RajaSilumanID, Bosses: []string{RajaSilumanID, RajaSilumanID, RajaSilumanID, RajaSilumanID},
		LootTableID: "lt-gerbang-33", WeeklyLockout: false,
		Rewards: RewardDef{Exp: 900, Coin: 280, Crystal: 6},
	})
}

func registerPhase23NPC() {
	registerNPC(NPCDef{
		ID: "pak_jaga", Name: "Pak Jaga", Role: "Penjaga Alas", Type: "WORLD",
		X: 4.2, Z: 28.5, Yaw: 2.8, DialogueID: "pak_jaga", InteractionRange: 2.6,
	})
	dialogueCatalog["pak_jaga"] = DialogueLine{
		Speaker: "Pak Jaga",
		Text:    "Ati-ati, Le! Ana siluman teka saka alas!",
	}
	registerInteract(InteractDef{ID: "cave-whispers-door", Kind: "dungeon", X: 5.2, Z: 31.0, Text: "CAVE OF WHISPERS"})
	registerInteract(InteractDef{ID: "gerbang-33-door", Kind: "dungeon", X: 6.4, Z: 33.2, Text: "GERBANG 33 SILUMAN"})
}

func registerPhase23Achievements() {
	for _, t := range []TitleDef{
		{ID: "whisper-explorer", Name: "Whisper Explorer", Source: "dungeon"},
		{ID: "dungeon-master", Name: "Dungeon Master", Source: "dungeon"},
		{ID: "learning-hero", Name: "Learning Hero", Source: "education"},
		{ID: "gerbang-raider", Name: "Gerbang Raider", Source: "raid"},
	} {
		registerTitleDef(t)
	}
	for _, a := range []AchievementDef{
		{ID: "first-dungeon", Name: "First Dungeon", Title: "whisper-explorer", Flag: "ach_first_dungeon", Category: "Dungeon"},
		{ID: "dungeon-master", Name: "Dungeon Master", Title: "dungeon-master", Flag: "ach_dungeon_master", Category: "Dungeon"},
		{ID: "fast-clear", Name: "Fast Clear", Title: "speed-runner", Flag: "ach_dungeon_fast", Category: "Dungeon"},
		{ID: "perfect-clear", Name: "Perfect Clear", Title: "whisper-explorer", Flag: "ach_dungeon_no_death", Category: "Dungeon"},
		{ID: "learning-hero", Name: "Learning Hero", Title: "learning-hero", Flag: "ach_learning_hero", Category: "Education"},
	} {
		registerAchievementDef(a)
	}
	registerCosmeticDef(CosmeticDef{ID: "cloak-whispers", Name: "Whisper Cloak", Kind: "cloak"})
	registerCosmeticDef(CosmeticDef{ID: "banner-gerbang", Name: "Gerbang Banner", Kind: "banner"})
}

func registerDungeonDef(d DungeonDef) {
	if dungeonByID[d.ID].ID != "" {
		return
	}
	dungeonCatalog = append(dungeonCatalog, d)
	dungeonByID[d.ID] = d
}

func registerBossDef(b BossDef) {
	if bossByID[b.ID].ID != "" {
		return
	}
	bossCatalog = append(bossCatalog, b)
	bossByID[b.ID] = b
}

func registerBossSkillDef(s BossSkillDef) {
	if bossSkillByID[s.ID].ID != "" {
		return
	}
	bossSkillByID[s.ID] = s
}

func registerWaveDef(w DungeonWaveDef) {
	if dungeonWaveByID[w.ID].ID != "" {
		return
	}
	dungeonWaveByID[w.ID] = w
}

func registerObjectiveDef(o DungeonObjectiveDef) {
	for _, cur := range dungeonObjByDun[o.DungeonID] {
		if cur.ID == o.ID {
			return
		}
	}
	dungeonObjByDun[o.DungeonID] = append(dungeonObjByDun[o.DungeonID], o)
}

func registerLootTableDef(t LootTableDef) {
	if lootTableByID[t.ID].ID != "" {
		return
	}
	lootTableByID[t.ID] = t
}

func registerWorldBossDef(d WorldBossDef) {
	if worldBossByID[d.ID].ID != "" {
		return
	}
	worldBossByID[d.ID] = d
}

func phase23Cheat(t string) bool {
	switch t {
	case TypeSetBossDead, TypeSpawnBoss, TypeCompleteDungeon, TypeSkipMechanic, TypeDamageBoss:
		return true
	default:
		return false
	}
}

func (w *WorldState) ApplyPhase23(p *Player, env Envelope) [][]byte {
	if phase23Cheat(env.Type) {
		return rejectFor(p.ID, env.Type, "server_authoritative")
	}
	switch env.Type {
	case TypeGetOpenWorld:
		return [][]byte{marshal(TypeOpenWorld, w.openWorldView(p))}
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) openWorldView(p *Player) OpenWorldView {
	j := p.worldJournal(w)
	out := OpenWorldView{Regions: j.Regions, Markers: j.Markers, NextWorldBoss: j.NextWorldBoss}
	if ev := w.eventViewFor(p); ev != nil {
		cp := *ev
		out.Event = &cp
	}
	if w.Boss != nil && (w.Boss.State == "ACTIVE" || w.Boss.State == "ANNOUNCED") {
		v := w.Boss.view()
		out.WorldBoss = &v
	}
	return out
}

func attachPhase23Journal(p *Player, w *WorldState, j *WorldJournal) {
	if j == nil {
		return
	}
	for i := range j.Regions {
		if r, ok := regionByID[j.Regions[i].ID]; ok {
			j.Regions[i].Title = r.Title
			j.Regions[i].RecommendedLevel = r.RecommendedLevel
			j.Regions[i].MinimumLevel = r.LevelMin
			j.Regions[i].EnemyTier = r.EnemyTier
			j.Regions[i].ResourceTier = r.ResourceTier
		}
	}
	j.Markers = phase23Markers(p, w)
	j.NextWorldBoss = nextWorldBossPreview()
}

func phase23Markers(p *Player, w *WorldState) []MapMarkerView {
	log := p.ensureLog()
	out := []MapMarkerView{
		{ID: "town-dawn", Kind: "Town", Name: "Dawn Village", Region: "village", X: 0, Z: 6},
		{ID: "guild-hall", Kind: "Guild Hall", Name: "Guild Hall", Region: "village", X: -5.2, Z: 4},
		{ID: "res-mist", Kind: "Resource", Name: "Mistwood", Region: "forest", X: -4, Z: 28},
		{ID: WhispersDungeonID, Kind: "Dungeon", Name: "CAVE OF WHISPERS", Region: "forest", X: 5.2, Z: 31},
		{ID: GerbangRaidID, Kind: "Raid", Name: "GERBANG 33 SILUMAN", Region: "forest", X: 6.4, Z: 33.2},
		{ID: "npc-pak-jaga", Kind: "NPC", Name: "Pak Jaga", Region: "forest", X: 4.2, Z: 28.5},
		{ID: "quest-whispers", Kind: "Quest", Name: "Jalan Menuju Hutan", Region: "forest", X: 1, Z: 24},
	}
	if live := w.Events.Active; live != nil {
		x, z := live.Def.X, live.Def.Z
		if x == 0 && z == 0 {
			z = 36
		}
		out = append(out, MapMarkerView{ID: live.Def.ID, Kind: "World Event", Name: live.Def.Name, Region: live.Region, X: x, Z: z})
	}
	if w.Boss != nil && (w.Boss.State == "ACTIVE" || w.Boss.State == "ANNOUNCED") {
		out = append(out, MapMarkerView{ID: w.Boss.Def.ID, Kind: "Boss", Name: w.Boss.Def.Name, Region: w.Boss.Def.Region, X: w.Boss.Def.X, Z: w.Boss.Def.Z})
	}
	_ = log
	return out
}

func nextWorldBossPreview() *WorldBossPreview {
	def, ok := worldBossByID[RatuBayanganID]
	if !ok {
		def = worldBossDef
	}
	return &WorldBossPreview{ID: def.ID, Name: def.Name, Region: def.Region, Announce: def.Announce, When: "scheduled"}
}

func (w *WorldState) spawnNamedWorldBoss(id string) [][]byte {
	def, ok := worldBossByID[id]
	if !ok {
		return nil
	}
	if w.Boss != nil && (w.Boss.State == "ACTIVE" || w.Boss.State == "ANNOUNCED") {
		return nil
	}
	hp := def.MaxHP
	if hp < 500 {
		hp = 8000
	}
	eDef := EnemyDef{
		ID: def.ID, Name: def.Name, Level: def.Level, MaxHP: hp, Attack: 16, Defense: 10, Speed: 1.5,
		AttackRange: 2.6, AggroRange: 16, Leash: 22, AttackCooldown: 1.8, ExpReward: 90, Behavior: "boss", Rank: "boss",
	}
	e := spawnEnemy(eDef, def.X, def.Z)
	e.NoRespawn = true
	w.enemies[e.ID] = e
	dur := def.DurationSec
	if dur < 30 {
		dur = 1200
	}
	w.Boss = &WorldBossLive{
		Def: def, State: "ACTIVE", SpawnID: randomID("wb_"),
		EnemyID: e.ID, HP: e.HP, MaxHP: e.MaxHP, Until: time.Now().Add(time.Duration(dur * float64(time.Second))),
		Damage: map[string]int{}, Heal: map[string]int{}, Support: map[string]int{}, Revive: map[string]int{}, Claimed: map[string]bool{},
	}
	v := w.Boss.view()
	return [][]byte{
		marshal(TypeWorldBossAnnounce, v),
		marshal(TypeWorldBossState, v),
		marshal(TypeChatMessage, ChatOut{Channel: "SYSTEM", From: "SYSTEM", Text: def.Announce, System: true}),
	}
}

func (w *WorldState) applyPhase23Rewards(p *Player, inst *DungeonInstance, elapsed float64, deaths int) {
	if p == nil || inst == nil {
		return
	}
	def := dungeonByID[inst.DefID]
	log := p.ensureLog()
	log.Flags["ach_first_dungeon"] = true
	if def.ID == WhispersDungeonID || def.ID == GerbangRaidID {
		log.Flags["ach_dungeon_master"] = true
	}
	if inst.WipeCount == 0 && deaths == 0 {
		log.Flags["ach_dungeon_no_death"] = true
	}
	if elapsed > 0 && elapsed <= WhispersFastSec {
		log.Flags["ach_dungeon_fast"] = true
	}
	if def.ID == WhispersDungeonID {
		p.grantTitle("whisper-explorer")
		p.grantCosmetic("cloak-whispers")
	}
	if def.ID == GerbangRaidID {
		w.addCurrency(p, "raid", 15, "raid_gerbang")
		p.grantTitle("gerbang-raider")
		p.grantCosmetic("banner-gerbang")
	}
}

func dungeonRoomLabel(inst *DungeonInstance) string {
	if inst == nil {
		return ""
	}
	if inst.DefID != WhispersDungeonID && inst.DefID != GerbangRaidID {
		return ""
	}
	if inst.State == DunBoss || inst.BossLocked {
		return "BOSS ROOM"
	}
	if inst.DefID == GerbangRaidID {
		switch inst.EncounterIndex {
		case 0:
			return "GROUND COMBAT"
		case 1:
			return "SUMMONING"
		case 2:
			return "ARENA COLLAPSE"
		default:
			return "FINAL ATTACK"
		}
	}
	switch inst.WaveIndex {
	case 0, 1:
		return "ENTRY"
	case 2:
		return "FOREST"
	case 3:
		return "BRIDGE"
	case 4:
		return "CHAMBER"
	default:
		return "BOSS ROOM"
	}
}

func eventProgress(live *LiveEvent) (int, int) {
	if live == nil {
		return 0, 0
	}
	need := live.Def.ObjectiveNeed
	if need < 1 {
		need = live.MaxGateHP
	}
	n := 0
	for _, s := range live.Score {
		n += s
	}
	if live.Def.Kind == "defend" || live.Def.Kind == "defense" {
		return live.GateHP, live.MaxGateHP
	}
	return n, need
}

func eventSucceeded(live *LiveEvent) bool {
	if live == nil {
		return false
	}
	if live.Def.ObjectiveNeed > 0 && (live.Def.Kind == "hunt" || live.Def.Kind == "collect" || live.Def.Kind == "rescue") {
		n, need := eventProgress(live)
		return n >= need && need > 0
	}
	return live.GateHP > 0
}

func (w *WorldState) bossEncounterLocked(p *Player) bool {
	if p == nil {
		return false
	}
	if inst := w.dungeonOf(p.ID); inst != nil && (inst.BossLocked || inst.State == DunBoss) {
		return true
	}
	if w.Boss != nil && (w.Boss.State == "ACTIVE" || w.Boss.State == "ANNOUNCED") {
		if math.Hypot(p.X-w.Boss.Def.X, p.Z-w.Boss.Def.Z) < 18 {
			return true
		}
		z := zoneAt(p.X, p.Z)
		if z.ID != "" && z.ID == w.Boss.Def.Region {
			return true
		}
	}
	return false
}

func grantWhisperEducation(p *Player, questionID string) {
	if p == nil || questionID != WhispersEduID {
		return
	}
	log := p.ensureLog()
	log.Flags["ach_learning_hero"] = true
	if log.EduToken < 99 {
		log.EduToken++
	}
	p.grantKnowledge(1)
	p.grantTitle("learning-hero")
}

func overlayEventAnnounce(def EventDef) string {
	if strings.TrimSpace(def.Announce) != "" {
		return def.Announce
	}
	return def.Name
}
