package mmo

import (
	"sort"
	"time"
)

const (
	Phase29WorldEventID = "bayangan-ing-alas"
	Phase29WorldBossID  = "raksa-cahaya-peteng"
	Phase29RaidID       = "raid-lampah-pungkasan"
)

type phase29RegionRow struct {
	ID, Theme, Purpose string
}

type phase29SilumanRow struct {
	ID, Name, Title, Region, Element, Weakness, Lore, BossType string
	Level                                                      int
	Skills                                                     []string
}

var (
	phase29Regions = []phase29RegionRow{
		{ID: "village", Theme: "Desa Awal", Purpose: "tutorial, NPC, quest, gathering, combat dasar"},
		{ID: "forest", Theme: "Hutan Kabut", Purpose: "siluman kecil, choice story, world event"},
		{ID: "valley", Theme: "Lembah Batu", Purpose: "guardian stone, mining, mini boss"},
		{ID: "temple", Theme: "Pegunungan Awan", Purpose: "flying enemies, elite route, chapter ascent"},
		{ID: "plains", Theme: "Danau Cahaya", Purpose: "fishing, water creatures, support quests"},
		{ID: "ruins", Theme: "Kota Tua", Purpose: "NPC hub, merchant, guild, endgame prep"},
		{ID: "canyon", Theme: "Hutan Gelap", Purpose: "elite siluman, live hunts, party content"},
		{ID: "masjid", Theme: "Gerbang Cahaya", Purpose: "endgame gate, raid, final journey"},
	}
	phase29Siluman = []phase29SilumanRow{
		{ID: "ragha", Name: "Klowo Alas", Title: "tutorial mini boss", Region: "Desa Awal", Level: 6, Element: "earth", Skills: []string{"basic attack"}, Weakness: "light", Lore: "Siluman cilik sing pisanan nyegat dalan mlebu alas.", BossType: "MINI_BOSS"},
		{ID: "gravon", Name: "Watu Ireng", Title: "tank", Region: "Lembah Batu", Level: 10, Element: "earth", Skills: []string{"stone armor"}, Weakness: "water", Lore: "Badan watu peteng iki nglindhungi jalur tambang lawas.", BossType: "REGIONAL"},
		{ID: "velra", Name: "Geni Lingsir", Title: "fire attacker", Region: "Hutan Kabut", Level: 12, Element: "fire", Skills: []string{"fire zone"}, Weakness: "water", Lore: "Geni abang petang iki kerep ninggal jejak kobaran ing alas.", BossType: "REGIONAL"},
		{ID: "kairoth", Name: "Kabut Sura", Title: "assassin", Region: "Hutan Kabut", Level: 13, Element: "dark", Skills: []string{"temporary invisibility"}, Weakness: "wind", Lore: "Sosok iki seneng nyerang saka pepeteng kabut.", BossType: "REGIONAL"},
		{ID: "nymra", Name: "Rembulan Peteng", Title: "caster", Region: "Hutan Kabut", Level: 14, Element: "dark", Skills: []string{"dark projectile"}, Weakness: "light", Lore: "Wulan peteng iki ngirim proyektil peteng saka pucuk wit.", BossType: "MINI_BOSS"},
		{ID: "vorak", Name: "Bayu Guntur", Title: "ranged", Region: "Pegunungan Awan", Level: 15, Element: "wind", Skills: []string{"wind attack"}, Weakness: "earth", Lore: "Panah angin lan gludhug dadi tandha rawuhe.", BossType: "REGIONAL"},
		{ID: "zeran", Name: "Tunggak Jagad", Title: "tank", Region: "Lembah Batu", Level: 16, Element: "earth", Skills: []string{"ground shockwave"}, Weakness: "water", Lore: "Akar raksasa iki nyebar kejut saka lemah.", BossType: "REGIONAL"},
		{ID: "malvek", Name: "Jaka Wulung", Title: "melee", Region: "Desa Awal", Level: 17, Element: "shadow", Skills: []string{"combo attack"}, Weakness: "light", Lore: "Prajurit peteng sing nyerang nganggo rangkaian tebasan cepet.", BossType: "REGIONAL"},
		{ID: "torga", Name: "Lawang Banyu", Title: "water guardian", Region: "Danau Cahaya", Level: 18, Element: "water", Skills: []string{"water shield"}, Weakness: "wind", Lore: "Penjaga gerbang banyu sing nganggo tameng aliran.", BossType: "REGIONAL"},
		{ID: "selka", Name: "Kembang Ireng", Title: "poison caster", Region: "Hutan Gelap", Level: 19, Element: "poison", Skills: []string{"poison field"}, Weakness: "fire", Lore: "Kembang peteng iki nyebar racun alon ing sela oyot.", BossType: "MINI_BOSS"},
		{ID: "dravon", Name: "Gagak Wengi", Title: "flying", Region: "Pegunungan Awan", Level: 20, Element: "wind", Skills: []string{"dive attack"}, Weakness: "earth", Lore: "Manuk wengi peteng sing nyelem saka awan.", BossType: "REGIONAL"},
		{ID: "morgha", Name: "Banyu Tiris", Title: "support", Region: "Danau Cahaya", Level: 21, Element: "water", Skills: []string{"heal"}, Weakness: "light", Lore: "Tetes banyu urip iki nambani sekutu siluman liyane.", BossType: "REGIONAL"},
		{ID: "kelran", Name: "Raja Suket", Title: "summoner", Region: "Hutan Kabut", Level: 22, Element: "earth", Skills: []string{"summon minions"}, Weakness: "fire", Lore: "Raja rumput liar sing nyeluk pengikut saka semak.", BossType: "REGIONAL"},
		{ID: "arvok", Name: "Guntur Watu", Title: "heavy", Region: "Lembah Batu", Level: 23, Element: "earth", Skills: []string{"ground smash"}, Weakness: "water", Lore: "Palu watu lan gludhug dadi siji ing pukulane.", BossType: "REGIONAL"},
		{ID: "nyxara", Name: "Jenggirat Angin", Title: "speed", Region: "Pegunungan Awan", Level: 24, Element: "wind", Skills: []string{"dash"}, Weakness: "earth", Lore: "Langkah cepete ninggal garis angin tajem.", BossType: "MINI_BOSS"},
		{ID: "torven", Name: "Mata Kabut", Title: "controller", Region: "Hutan Kabut", Level: 25, Element: "mist", Skills: []string{"blind zone"}, Weakness: "light", Lore: "Mripat kabut iki nutupi pandelengan wong sing sembrana.", BossType: "REGIONAL"},
		{ID: "veyra", Name: "Kethek Gunung", Title: "bruiser", Region: "Pegunungan Awan", Level: 26, Element: "earth", Skills: []string{"leap attack"}, Weakness: "wind", Lore: "Lompatan keras saka tebing dadi ciri khasé.", BossType: "REGIONAL"},
		{ID: "gharon", Name: "Naga Awan", Title: "flying boss", Region: "Pegunungan Awan", Level: 28, Element: "wind", Skills: []string{"air attack"}, Weakness: "earth", Lore: "Naga kabut iki muter ing pucuk awan lawas.", BossType: "REGIONAL_BOSS"},
		{ID: "lurak", Name: "Lembu Watu", Title: "tank", Region: "Lembah Batu", Level: 29, Element: "earth", Skills: []string{"charge"}, Weakness: "water", Lore: "Tanduk watu lan langkah aboté nabrak tanpa ampun.", BossType: "REGIONAL"},
		{ID: "sorya", Name: "Jalak Peteng", Title: "ranged", Region: "Hutan Gelap", Level: 30, Element: "dark", Skills: []string{"dark arrows"}, Weakness: "light", Lore: "Panah peteng mabur saka sela dahan peteng.", BossType: "REGIONAL"},
		{ID: "varok", Name: "Pangreksa Alas", Title: "forest guardian", Region: "Hutan Gelap", Level: 31, Element: "earth", Skills: []string{"root trap"}, Weakness: "fire", Lore: "Penjaga alas peteng iki ngiket mungsuh nganggo oyot.", BossType: "REGIONAL"},
		{ID: "zavra", Name: "Sura Geni", Title: "fire elite", Region: "Hutan Gelap", Level: 32, Element: "fire", Skills: []string{"flame dash"}, Weakness: "water", Lore: "Serudukan geni cepet saka lorong pepeteng.", BossType: "REGIONAL"},
		{ID: "orvak", Name: "Rembes Banyu", Title: "water caster", Region: "Danau Cahaya", Level: 33, Element: "water", Skills: []string{"water burst"}, Weakness: "wind", Lore: "Ledakan banyu cahya metu saka danau lan rerimbunan.", BossType: "REGIONAL"},
		{ID: "merra", Name: "Wulung Bayang", Title: "assassin", Region: "Kota Tua", Level: 34, Element: "shadow", Skills: []string{"shadow clone"}, Weakness: "light", Lore: "Bayang-bayang pecah dadi klon sakdurunge nyerang.", BossType: "REGIONAL"},
		{ID: "tharos", Name: "Buto Awan", Title: "giant", Region: "Pegunungan Awan", Level: 35, Element: "wind", Skills: []string{"area smash"}, Weakness: "earth", Lore: "Raksasa mendhung sing ngebanting lemah gunung.", BossType: "MINI_BOSS"},
		{ID: "ragna", Name: "Prajurit Watu", Title: "elite warrior", Region: "Lembah Batu", Level: 36, Element: "earth", Skills: []string{"shield stance"}, Weakness: "water", Lore: "Prajurit batu iki tahan serangan yen ora dipecah pertahanane.", BossType: "REGIONAL"},
		{ID: "velkor", Name: "Singa Kabut", Title: "beast", Region: "Hutan Kabut", Level: 37, Element: "mist", Skills: []string{"fear roar"}, Weakness: "light", Lore: "Aumané nggawe rombongan buyar yen ora kompak.", BossType: "REGIONAL"},
		{ID: "aranya", Name: "Garuda Peteng", Title: "flying elite", Region: "Pegunungan Awan", Level: 38, Element: "dark", Skills: []string{"air dive"}, Weakness: "earth", Lore: "Nyelem saka ndhuwur nganggo cakar peteng.", BossType: "REGIONAL"},
		{ID: "gorven", Name: "Ratu Akar", Title: "forest boss", Region: "Hutan Gelap", Level: 40, Element: "earth", Skills: []string{"summon roots"}, Weakness: "fire", Lore: "Ratu oyot sing marakaké alas peteng dadi urip.", BossType: "REGIONAL_BOSS"},
		{ID: "sylra", Name: "Raksasa Geni", Title: "fire boss", Region: "Kota Tua", Level: 42, Element: "fire", Skills: []string{"meteor zone"}, Weakness: "water", Lore: "Pecahan meteore murub ing reruntuhan kutha lawas.", BossType: "REGIONAL_BOSS"},
		{ID: "korvan", Name: "Panglima Bayangan", Title: "warrior boss", Region: "Kota Tua", Level: 45, Element: "shadow", Skills: []string{"multi combo"}, Weakness: "light", Lore: "Panglima peteng iki nggabungake teknik pedhang lan kabut.", BossType: "REGIONAL_BOSS"},
		{ID: "elyra", Name: "Penjaga Gerbang", Title: "guardian boss", Region: "Gerbang Cahaya", Level: 48, Element: "light", Skills: []string{"multi phase"}, Weakness: "earth", Lore: "Penjaga pungkasan sadurunge gerbang kebuka kanggo rombongan sing pantes.", BossType: "REGIONAL_BOSS"},
		{ID: "avaron", Name: "RAKSA CAHAYA PETENG", Title: "final regional boss", Region: "Gerbang Cahaya", Level: 50, Element: "dark", Skills: []string{"multiple phases"}, Weakness: "light", Lore: "Raksa ageng iki nyoba kekuwatan, ilmu, lan kabecikan bebarengan.", BossType: "FINAL_REGIONAL_BOSS"},
	}
)

func init() {
	registerPhase29()
}

func registerPhase29() {
	registerPhase29Education()
	registerPhase29Story()
	registerPhase29Bosses()
	registerPhase29Achievements()
}

func registerPhase29Education() {
	registerQuestion(QuestionDef{ID: "q-add-2-4", Category: "Matematika", Prompt: "2 + 4 = ?", Choices: []string{"5", "6", "7"}, Correct: 1, Explain: "2 ditambah 4 sama dengan 6.", Grade: 1})
	registerQuestion(QuestionDef{ID: "q-sub-7-3", Category: "Matematika", Prompt: "7 - 3 = ?", Choices: []string{"3", "4", "5"}, Correct: 1, Explain: "7 dikurangi 3 sama dengan 4.", Grade: 1})
	registerQuestion(QuestionDef{ID: "q-read-buku", Category: "Membaca", Prompt: "Manakah kata yang benar?", Choices: []string{"BUKU", "BUKUUU", "BUKK"}, Correct: 0, Explain: "Kata yang benar adalah BUKU.", Grade: 1})
	registerQuestion(QuestionDef{ID: "q-env-timur", Category: "Pengenalan lingkungan", Prompt: "Matahari terbit dari arah?", Choices: []string{"Barat", "Timur", "Selatan"}, Correct: 1, Explain: "Matahari terbit dari timur.", Grade: 1})
}

func registerPhase29Story() {
	registerQuest(QuestDef{
		ID: "mq029", Title: "Lampah Pungkasan.", Kind: "main", NPC: "mbah_jagat",
		Location: "Gerbang Cahaya", Description: "Perjalananmu wis adoh banget, Le. Saiki wayahe mbuktekake kekuwatan lan kabecikanmu.",
		Objectives: []ObjectiveDef{
			{Type: "DUNGEON", Target: Phase29RaidID, Count: 1, Text: "Selesaikan Final Journey"},
			{Type: "ANSWER", Target: "q-read-buku", Count: 1, Text: "Jawab tantangan pendidikan"},
			{Type: "KILL", Target: "raksa_gate_alpha", Count: 1, Text: "Kalahkan elite penjaga"},
		},
		Rewards:      RewardDef{Exp: 600, Coin: 180, EduToken: 2, Knowledge: 2},
		FlagsOnClaim: []string{"phase29_final_journey_done"}, ClaimAt: "mbah_jagat",
	})
	dialogueCatalog["phase29-final"] = DialogueLine{
		Speaker: "Mbah Jagat",
		Text:    "Perjalananmu wis adoh banget, Le. Saiki wayahe mbuktekake kekuwatan lan kabecikanmu.",
	}
	addLore29 := func(id, title, region, text string) {
		if loreByID[id].ID != "" {
			return
		}
		ld := LoreDef{ID: id, Title: title, Region: region, Text: text}
		loreCatalog = append(loreCatalog, ld)
		loreByID[id] = ld
	}
	addLore29("lore-phase29-journey", "Petualangan Menuju Cahaya", "masjid", "Perjalanan iki ora mung babagan menang perang, nanging uga belajar, nulung warga, lan mbukak dalan bebarengan.")
	addLore29("lore-phase29-live", "Festival Cahaya", "ruins", "Festival Cahaya nggawa mini quest, kuis pendhidhikan, lan gotong royong antar petualang.")
}

func registerPhase29Bosses() {
	registerEventDef(EventDef{
		ID: Phase29WorldEventID, Name: "Bayangan Ing Alas", Kind: "hunt", Region: "forest",
		Announce: "Bayangan Ing Alas dimulai! Kalahkan invasi siluman di Mistwood.",
		AnnounceJV: "Bayangan Ing Alas diwiwiti! Tumpes serangan siluman ing alas.",
		DurationSec: 180, GateHP: 260, Enemies: []string{"forest_fang", "shadow_imp"},
		Objective: "Kalahkan invasi siluman", ObjectiveNeed: 20, Rewards: RewardDef{Exp: 60, Coin: 24, EduToken: 1}, X: 0, Z: 34,
	})
	registerWorldBossDef(WorldBossDef{
		ID: Phase29WorldBossID, Name: "RAKSA CAHAYA PETENG", Title: "World Boss", Region: "masjid",
		X: 0, Z: 154, Level: 50, MaxHP: 24000, MaxPlayers: 100,
		Announce: "RAKSA CAHAYA PETENG muncul di Gerbang Cahaya!", DurationSec: 900,
		Rewards: RewardDef{Exp: 260, Coin: 120, Crystal: 4, Knowledge: 1},
	})
	registerBossSkillDef(BossSkillDef{ID: "raid-positional-wave", Name: "Positional Wave", Type: "AREA", Damage: 26, Radius: 4.0, Range: 10, Telegraph: 1.0, Cooldown: 8, Phase: 1, VFX: "wave", Shape: "line"})
	registerBossSkillDef(BossSkillDef{ID: "raid-add-call", Name: "Add Call", Type: "SUMMON", Range: 8, Telegraph: 1.0, Cooldown: 12, Phase: 2, VFX: "summon", SummonID: "shadow_imp", SummonCount: 4, Shape: "circle"})
	registerBossSkillDef(BossSkillDef{ID: "raid-gate-break", Name: "Gate Break", Type: "ULTIMATE", Damage: 34, Radius: 5.4, Range: 12, Telegraph: 1.6, Cooldown: 16, Phase: 3, VFX: "burst", Shape: "ring"})
	for _, b := range []BossDef{
		{
			ID: "raksa_gate_alpha", Name: "KLOWO ALAS AGUNG", Title: "Raid Boss 1", Type: "Original Raid",
			Level: 46, MaxHP: 5000, Attack: 26, Defense: 18, Speed: 1.6, Leash: 20, EnrageTime: 360, LootTableID: "lt-phase29-raid",
			Skills: []string{"raid-positional-wave"}, Phases: []BossPhaseDef{{ID: 1, HPPct: 1.0, Label: "PHASE 1"}, {ID: 2, HPPct: 0.45, Label: "PHASE 2"}},
		},
		{
			ID: "raksa_gate_beta", Name: "PANGREKSA ALAS AGUNG", Title: "Raid Boss 2", Type: "Original Raid",
			Level: 48, MaxHP: 6200, Attack: 28, Defense: 20, Speed: 1.55, Leash: 20, EnrageTime: 360, LootTableID: "lt-phase29-raid",
			Skills: []string{"raid-add-call"}, Phases: []BossPhaseDef{{ID: 1, HPPct: 1.0, Label: "PHASE 1"}, {ID: 2, HPPct: 0.4, Label: "PHASE 2"}},
		},
		{
			ID: "raksa_gate_omega", Name: "RAKSA CAHAYA PETENG", Title: "Raid Boss 3", Type: "Original Raid",
			Level: 50, MaxHP: 9000, Attack: 30, Defense: 22, Speed: 1.5, Leash: 22, EnrageTime: 420, LootTableID: "lt-phase29-raid",
			Skills: []string{"raid-positional-wave", "raid-add-call", "raid-gate-break"},
			Phases: []BossPhaseDef{{ID: 1, HPPct: 1.0, Label: "PHASE 1"}, {ID: 2, HPPct: 0.66, Label: "PHASE 2"}, {ID: 3, HPPct: 0.33, Label: "PHASE 3"}},
		},
	} {
		registerBossDef(b)
	}
	registerLootTableDef(LootTableDef{
		ID: "lt-phase29-raid", RollType: "PERSONAL",
		Entries: []LootEntryDef{
			{ItemID: "crystal_shard", MinQuantity: 2, MaxQuantity: 4, Chance: 1},
			{ItemID: "knowledge_token", MinQuantity: 1, MaxQuantity: 1, Chance: 0.7},
			{ItemID: "banner-gerbang", MinQuantity: 1, MaxQuantity: 1, Chance: 0.1},
		},
	})
	registerWaveDef(DungeonWaveDef{ID: "p29-raid-1", DungeonID: Phase29RaidID, Index: 1, Name: "POSITIONING", Spawns: []WaveSpawnDef{{EnemyID: "raksa_gate_alpha", Count: 1, X: -2, Z: 20}}})
	registerWaveDef(DungeonWaveDef{ID: "p29-raid-2", DungeonID: Phase29RaidID, Index: 2, Name: "ADDS", Spawns: []WaveSpawnDef{{EnemyID: "raksa_gate_beta", Count: 1, X: 0, Z: 26}}})
	registerWaveDef(DungeonWaveDef{ID: "p29-raid-3", DungeonID: Phase29RaidID, Index: 3, Name: "FINAL GATE", Spawns: []WaveSpawnDef{{EnemyID: "raksa_gate_omega", Count: 1, X: 0, Z: 32}}})
	registerObjectiveDef(DungeonObjectiveDef{ID: "p29-r1", DungeonID: Phase29RaidID, Index: 0, Type: "BOSS", Target: "raksa_gate_alpha", Count: 1, Text: "Atur posisi dan kalahkan boss pertama."})
	registerObjectiveDef(DungeonObjectiveDef{ID: "p29-r2", DungeonID: Phase29RaidID, Index: 1, Type: "BOSS", Target: "raksa_gate_beta", Count: 1, Text: "Tahan adds dan kalahkan boss kedua."})
	registerObjectiveDef(DungeonObjectiveDef{ID: "p29-r3", DungeonID: Phase29RaidID, Index: 2, Type: "BOSS", Target: "raksa_gate_omega", Count: 1, Text: "Buka final gate dan kalahkan boss terakhir."})
	registerDungeonDef(DungeonDef{
		ID: Phase29RaidID, Name: "LAMPAH PUNGKASAN", Kind: "RAID", Description: "Raid final 10-20 pemain dengan 3 boss, checkpoint, lan final gate.",
		Difficulty: "HARD", MinimumLevel: 50, RecommendedLevel: 50, MinPlayers: 10, MaxPlayers: 20, TimeLimit: 2700,
		Environment: "final-gate", Region: "masjid", EnemyWaves: []string{"p29-raid-1", "p29-raid-2", "p29-raid-3"},
		BossID: "raksa_gate_omega", Bosses: []string{"raksa_gate_alpha", "raksa_gate_beta", "raksa_gate_omega"},
		LootTableID: "lt-phase29-raid", WeeklyLockout: true, EducationBoss: "q-sub-7-3",
		Rewards: RewardDef{Exp: 1200, Coin: 320, Crystal: 8, Knowledge: 2},
	})
}

func registerPhase29Achievements() {
	for _, t := range []TitleDef{
		{ID: "pemburu-siluman", Name: "Pemburu Siluman", Source: "phase29", Rarity: "UNCOMMON"},
		{ID: "penjaga-desa", Name: "Penjaga Desa", Source: "phase29", Rarity: "UNCOMMON"},
		{ID: "sahabat-cahaya", Name: "Sahabat Cahaya", Source: "phase29", Rarity: "RARE"},
		{ID: "penjelajah-dunia", Name: "Penjelajah Dunia", Source: "phase29", Rarity: "RARE"},
		{ID: "pahlawan-pendidikan", Name: "Pahlawan Pendidikan", Source: "phase29", Rarity: "RARE"},
		{ID: "guild-hero", Name: "Guild Hero", Source: "phase29", Rarity: "RARE"},
	} {
		registerTitleDef(t)
	}
	for _, c := range []CosmeticDef{
		{ID: "badge-phase29-raid", Name: "Raid of Light Badge", Kind: "badge"},
		{ID: "cloak-phase29-final", Name: "Final Journey Cloak", Kind: "cloak"},
	} {
		registerCosmeticDef(c)
	}
	for _, a := range []AchievementDef{
		{ID: "defeat-1-siluman", Name: "Defeat 1 Siluman", Title: "pemburu-siluman", Flag: "ach_defeat_1_siluman", Category: "Phase29"},
		{ID: "defeat-10-siluman", Name: "Defeat 10 Siluman", Title: "pemburu-siluman", Flag: "ach_defeat_10_siluman", Category: "Phase29"},
		{ID: "defeat-33-siluman", Name: "Defeat 33 Siluman", Title: "sahabat-cahaya", Flag: "ach_defeat_33_siluman", Category: "Phase29"},
		{ID: "complete-chapter-phase29", Name: "Complete Chapter", Title: "penjelajah-dunia", Flag: "ach_complete_chapter_phase29", Category: "Phase29"},
		{ID: "complete-raid-phase29", Name: "Complete Raid", Title: "penjaga-desa", Flag: "ach_complete_raid_phase29", Category: "Phase29"},
		{ID: "defeat-world-boss-phase29", Name: "Defeat World Boss", Title: "penjaga-desa", Flag: "ach_defeat_world_boss_phase29", Category: "Phase29"},
		{ID: "educational-hero-phase29", Name: "Educational Hero", Title: "pahlawan-pendidikan", Flag: "ach_educational_hero_phase29", Category: "Phase29"},
		{ID: "guild-hero-phase29", Name: "Guild Hero", Title: "guild-hero", Flag: "ach_guild_hero_phase29", Category: "Phase29"},
	} {
		registerAchievementDef(a)
	}
}

func phase29ChapterRows(log *PlayerLog) []OverlayChapterView {
	titles := []string{
		"Desa Awal", "Hutan Kabut", "Lembah Batu", "Pegunungan Awan",
		"Danau Cahaya", "Kota Tua", "Hutan Gelap", "Gerbang Cahaya",
	}
	out := make([]OverlayChapterView, 0, len(titles))
	for i, title := range titles {
		state := StoryLocked
		if i == 0 {
			state = StoryActive
		}
		if log.Flags["story_chapter_"+itoa(i+1)+"_complete"] || log.Flags["st-ch0"+itoa(i+1)+"_complete"] {
			state = StoryCompleted
		} else if i > 0 && (log.Flags["story_chapter_"+itoa(i)+"_complete"] || log.Flags["st-ch0"+itoa(i)+"_complete"]) {
			state = StoryActive
		}
		out = append(out, OverlayChapterView{Index: i + 1, Title: title, State: state, Locked: state == StoryLocked})
	}
	return out
}

func phase29EnemyLore(log *PlayerLog) []EnemyLoreView {
	out := make([]EnemyLoreView, 0, len(phase29Siluman))
	for _, row := range phase29Siluman {
		state := log.SilumanProg[row.ID]
		defeated := state == SilumanDefeated || log.Guardians[row.ID] || log.Flags["guardian_"+row.ID+"_defeated"]
		met := defeated || state == SilumanUnderstood || state == SilumanAlly || log.Lore["phase29:"+row.ID]
		out = append(out, EnemyLoreView{
			ID: row.ID, Name: row.Name, Region: row.Region, Personality: row.Element, Mechanic: firstOr(row.Skills, "basic attack"),
			Encountered: met, Defeated: defeated, Discovered: met, Lore: row.Lore,
		})
	}
	return out
}

func phase29WorldMood(w *WorldState) string {
	if w != nil && w.Boss != nil && (w.Boss.State == "ACTIVE" || w.Boss.State == "ANNOUNCED") {
		return "Gerbang Cahaya bergetar. Aura boss agung nyebar ing sakupenge."
	}
	if w != nil && w.Events.Active != nil {
		return "Bayangan lan gotong royong dadi warna utama petualangan dina iki."
	}
	return "Perjalanan tumuju cahaya terus mbukak wilayah anyar, siluman anyar, lan piwulang anyar."
}

func phase29Objective(log *PlayerLog) string {
	if log == nil {
		return "Mulai perjalanan menuju cahaya."
	}
	defeated := 0
	for _, row := range phase29Siluman {
		if log.Guardians[row.ID] || log.Flags["guardian_"+row.ID+"_defeated"] || log.SilumanProg[row.ID] == SilumanDefeated {
			defeated++
		}
	}
	switch {
	case log.Flags["phase29_final_journey_done"]:
		return "Final Journey selesai. World Events, raid, guild expedition, lan seasonal content tetep mbukak."
	case defeated >= 33:
		return "33 siluman wis rampung. Kumpulake party lan bukak Lampah Pungkasan."
	case defeated >= 20:
		return "Terusake perjalanan menyang Hutan Gelap lan Gerbang Cahaya."
	case defeated >= 10:
		return "Siluman saya kuwat. Siapna party, guild, lan latihan pendidikan."
	default:
		return "Lelakonmu diwiwiti saka desa, banjur mlebu kabut kanggo ngadhepi siluman siji-siji."
	}
}

func phase29EnrichJournal(p *Player, w *WorldState, j *WorldJournal) {
	if p == nil || j == nil {
		return
	}
	log := p.ensureLog()
	if len(j.OverlayChapters) == 0 {
		j.OverlayChapters = phase29ChapterRows(log)
	}
	if len(j.EnemyLore) == 0 {
		j.EnemyLore = phase29EnemyLore(log)
	}
	j.WorldMood = phase29WorldMood(w)
	if obj := phase29Objective(log); obj != "" {
		j.Objective = obj
	}
}

func phase29EnrichStoryState(p *Player, out *StoryStateView) {
	if p == nil || out == nil {
		return
	}
	log := p.ensureLog()
	out.Objective = phase29Objective(log)
	out.FinalBoss = map[string]any{
		"id": "raksa_gate_omega", "name": "RAKSA CAHAYA PETENG",
		"phases": []string{"Phase 1", "Phase 2", "Phase 3"},
		"questionId": "q-read-buku",
		"raidId": Phase29RaidID,
	}
}

func phase29BoardRows(w *WorldState) map[string][]map[string]any {
	type scoreRow struct {
		ID    string
		Name  string
		Score int
		Level int
	}
	build := func(scoreFn func(*Player) int) []map[string]any {
		rows := []scoreRow{}
		for _, p := range w.players {
			if p == nil {
				continue
			}
			rows = append(rows, scoreRow{ID: p.ID, Name: p.Name, Score: scoreFn(p), Level: p.Level})
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Score == rows[j].Score {
				return rows[i].Name < rows[j].Name
			}
			return rows[i].Score > rows[j].Score
		})
		if len(rows) > 10 {
			rows = rows[:10]
		}
		out := make([]map[string]any, 0, len(rows))
		for i, r := range rows {
			out = append(out, map[string]any{"rank": i + 1, "id": r.ID, "name": r.Name, "score": r.Score, "level": r.Level})
		}
		return out
	}
	sumMap := func(m map[string]int) int {
		n := 0
		for _, v := range m {
			n += v
		}
		return n
	}
	return map[string][]map[string]any{
		"level": build(func(p *Player) int { return p.Level }),
		"combat": build(func(p *Player) int { return p.powerRating() }),
		"bossDefeated": build(func(p *Player) int { return len(p.ensureLog().Guardians) + len(p.ensureLog().WorldBossClaims) }),
		"quest": build(func(p *Player) int { return len(p.ensureLog().Claimed) }),
		"crafting": build(func(p *Player) int { return sumMap(p.ensureLog().RecipeCrafts) }),
		"gathering": build(func(p *Player) int { return p.ensureLog().ProfessionXP[ProfMiner] + p.ensureLog().ProfessionXP[ProfWoodcutter] + p.ensureLog().ProfessionXP[ProfHerbalist] + p.ensureLog().ProfessionXP[ProfFisher] }),
	}
}

func firstOr(xs []string, fallback string) string {
	if len(xs) > 0 && xs[0] != "" {
		return xs[0]
	}
	return fallback
}

func phase29EnrichEndgameView(w *WorldState, p *Player, view map[string]any) {
	if w == nil || p == nil || view == nil {
		return
	}
	now := time.Now().UTC()
	boards := phase29BoardRows(w)
	view["leaderboards"] = map[string]any{
		"horizon": view["leaderboards"].(map[string]any)["horizon"],
		"season": view["leaderboards"].(map[string]any)["season"],
		"guilds": view["leaderboards"].(map[string]any)["guilds"],
		"level": boards["level"],
		"combat": boards["combat"],
		"bossDefeated": boards["bossDefeated"],
		"quest": boards["quest"],
		"crafting": boards["crafting"],
		"gathering": boards["gathering"],
	}
	view["phase29"] = map[string]any{
		"mainStory": map[string]any{
			"title": "Petualangan Menuju Cahaya",
			"objective": phase29Objective(p.ensureLog()),
			"chapters": phase29ChapterRows(p.ensureLog()),
			"regions": phase29Regions,
		},
		"siluman33": phase29Siluman,
		"worldEvent": map[string]any{
			"id": Phase29WorldEventID, "name": "Bayangan Ing Alas",
			"active": w.Events.Active != nil && w.Events.Active.Def.ID == Phase29WorldEventID,
			"globalProgress": func() int {
				if w.Events.Active != nil && w.Events.Active.Def.ID == Phase29WorldEventID {
					pg, need := eventProgress(w.Events.Active)
					if need > 0 {
						return 100 * pg / need
					}
				}
				return 0
			}(),
			"timer": func() int {
				if w.Events.Active != nil && w.Events.Active.Def.ID == Phase29WorldEventID {
					return int(w.Events.Active.Until.Sub(now).Seconds())
				}
				return 0
			}(),
		},
		"worldBoss": map[string]any{
			"id": Phase29WorldBossID, "name": "RAKSA CAHAYA PETENG",
			"state": func() string {
				if w.Boss != nil {
					return w.Boss.State
				}
				return "SCHEDULED"
			}(),
			"dailyLimit": "server_authoritative",
		},
		"raid": map[string]any{
			"id": Phase29RaidID, "name": "Lampah Pungkasan", "minPlayers": 10, "maxPlayers": 20,
			"difficulties": []string{"HARD", "NIGHTMARE"}, "checkpoint": true, "bosses": []string{"raksa_gate_alpha", "raksa_gate_beta", "raksa_gate_omega"},
		},
		"seasonal": map[string]any{
			"theme": seasonTrack.Theme, "seasonName": seasonTrack.Name,
			"festival": "Festival Cahaya", "upcoming": w.eventCalendar(),
		},
		"guildEvent": map[string]any{
			"name": "Guild Expedition", "weekly": true, "hallEvent": func() string {
				if g := w.guildOf(p.ID); g != nil {
					return g.HallEvent
				}
				return ""
			}(),
			"contributions": len(w.GuildContrib),
		},
		"finalJourney": map[string]any{
			"questId": "mq029", "questName": "Lampah Pungkasan.", "raidId": Phase29RaidID,
			"gateLocked": !p.ensureLog().Flags["phase29_final_journey_done"],
			"partyRequired": 10, "cinematic": "cin-final-war",
		},
	}
	phase29RefreshFlags(w, p)
}

func phase29RefreshFlags(w *WorldState, p *Player) {
	if w == nil || p == nil {
		return
	}
	log := p.ensureLog()
	defeated := 0
	for _, row := range phase29Siluman {
		if log.Guardians[row.ID] || log.Flags["guardian_"+row.ID+"_defeated"] || log.SilumanProg[row.ID] == SilumanDefeated {
			defeated++
		}
	}
	if defeated >= 1 {
		log.Flags["ach_defeat_1_siluman"] = true
		p.grantTitle("pemburu-siluman")
	}
	if defeated >= 10 {
		log.Flags["ach_defeat_10_siluman"] = true
	}
	if defeated >= 33 {
		log.Flags["ach_defeat_33_siluman"] = true
		p.grantTitle("sahabat-cahaya")
	}
	if len(log.WorldBossClaims) > 0 {
		log.Flags["ach_defeat_world_boss_phase29"] = true
		p.grantTitle("penjaga-desa")
	}
	if log.EduCorrect >= 3 {
		log.Flags["ach_educational_hero_phase29"] = true
		p.grantTitle("pahlawan-pendidikan")
	}
	if g := w.guildOf(p.ID); g != nil && g.Exp >= 100 {
		log.Flags["ach_guild_hero_phase29"] = true
		p.grantTitle("guild-hero")
	}
	if log.Flags["phase29_final_journey_done"] {
		log.Flags["ach_complete_raid_phase29"] = true
		log.Flags["ach_complete_chapter_phase29"] = true
		p.grantCosmetic("cloak-phase29-final")
	}
}

