package mmo

import (
	"strings"
	"time"
)

// Phase 24 overlay: NPC intelligence + dialogue + story branching + lore + education.
// Reuses NPCService, DialogueService, StoryService, QuestService, WorldService,
// CharacterService, PartyService, GuildService, AchievementService,
// EducationService, LocalizationService, InventoryService.
// VoiceService is architecture-only (text prototype; voiceId reserved).
// Logical tables live on catalogs + PlayerLog:
// npcs, npc_schedules, npc_dialogues, dialogue_nodes, dialogue_choices,
// story_chapters, story_nodes, story_flags, player_story_flags,
// npc_relationships, npc_memory, lore_entries, player_lore, enemy_lore,
// cinematics, educational_questions, educational_attempts.
// Indexes: npcId, playerId, storyId, chapterId, dialogueId, relationshipId.

const (
	EduAskCooldownSec = 45
	RelStranger       = "STRANGER"
	RelAcquaintance   = "ACQUAINTANCE"
	RelFriend         = "FRIEND"
	RelTrusted        = "TRUSTED"
	MoodForestSafe    = "FOREST_SAFE"
	MoodForestDanger  = "FOREST_DANGER"
)

func init() {
	registerPhase24()
}

func registerPhase24() {
	registerPhase24Places()
	registerPhase24NPCs()
	registerPhase24Dialogue()
	registerPhase24Quests()
	registerPhase24Education()
	registerPhase24Lore()
	registerPhase24Cinematics()
	registerPhase24Meta()
}

func registerPhase24Places() {
	if _, ok := npcPlaces["tavern"]; !ok {
		npcPlaces["tavern"] = npcVec{-6.4, 1.2}
	}
	if _, ok := npcPlaces["house"]; !ok {
		npcPlaces["house"] = npcVec{-9, 8}
	}
}

func overlayNPC(id string, fn func(*NPCDef)) {
	n, ok := npcByID[id]
	if !ok {
		return
	}
	fn(&n)
	npcByID[id] = n
	for i := range npcCatalog {
		if npcCatalog[i].ID == id {
			npcCatalog[i] = n
			return
		}
	}
}

func overlayLoc(key, jv, idn string) {
	if _, ok := storyDialogue[key]; ok {
		return
	}
	storyDialogue[key] = LocLine{JV: jv, ID: idn}
}

func registerLore(l LoreDef) {
	if loreByID[l.ID].ID != "" {
		return
	}
	loreCatalog = append(loreCatalog, l)
	loreByID[l.ID] = l
}

func registerCinematic(c CinematicDef) {
	if cinematicByID[c.ID].ID != "" {
		return
	}
	if c.DurationSec < 5 {
		c.DurationSec = 5
	}
	if c.DurationSec > 30 {
		c.DurationSec = 30
	}
	cinematicCatalog = append(cinematicCatalog, c)
	cinematicByID[c.ID] = c
}

func phase24Sched(morning, afternoon, evening, night string) map[string]string {
	return map[string]string{
		"morning":   morning,
		"afternoon": afternoon,
		"evening":   evening,
		"night":     night,
		"day":       afternoon,
	}
}

func registerPhase24NPCs() {
	overlayNPC("mbah_karya", func(n *NPCDef) {
		n.Region, n.Trait = "village", "WISE"
		n.DialogueProfile, n.QuestProfile = "karya-chain", "nq-karya"
		n.VoiceLineID = "voice-mbah-karya"
	})
	overlayNPC("mbok_rasa", func(n *NPCDef) {
		n.Region, n.Trait = "village", "KIND"
		n.DialogueProfile = "rasa-cook"
		n.VoiceLineID = "voice-mbok-rasa"
	})
	overlayNPC("pak_jala", func(n *NPCDef) {
		n.Region, n.Trait = "village", "CHEERFUL"
		n.DialogueProfile = "jala-fish"
		n.VoiceLineID = "voice-pak-jala"
	})
	overlayNPC("mbah_batu", func(n *NPCDef) {
		n.Region, n.Trait = "village", "STRICT"
		n.DialogueProfile = "batu-mine"
		n.VoiceLineID = "voice-mbah-batu"
	})
	overlayNPC("pak_jaga", func(n *NPCDef) {
		n.Region, n.Trait = "forest", "BRAVE"
		n.DialogueProfile = "jaga-warn"
		n.VoiceLineID = "voice-pak-jaga"
	})

	registerNPC(NPCDef{
		ID: "mbah_guru", Name: "Mbah Guru", Role: "Guru Desa", Type: "MASTER",
		X: -8.6, Z: 2.4, Yaw: 1.4, DialogueID: "mbah_guru", InteractionRange: 2.6,
		Region: "village", Trait: "WISE", DialogueProfile: "edu-grade1", QuestProfile: "nq-guru",
		VoiceLineID: "voice-mbah-guru", Occupation: "SCHOLAR",
		Schedule: phase24Sched("school", "school", "home", "home"),
		QuestIDs: []string{"nq-guru-1", "nq-guru-2", "nq-guru-3"},
		Personality: "Sabar, jelas, lan seneng mulang.",
	})
	registerNPC(NPCDef{
		ID: "laras", Name: "Laras", Role: "Petani", Type: "FARMER",
		X: 2.2, Z: 6.2, Yaw: 3.1, DialogueID: "laras", InteractionRange: 2.6,
		Region: "village", Trait: "KIND", DialogueProfile: "farmer-day",
		VoiceLineID: "voice-laras", Occupation: "FARMER",
		Schedule: phase24Sched("market", "house", "tavern", "home"),
		Personality: "Ramah lan seneng nulung.",
	})
	registerNPC(NPCDef{
		ID: "jaka", Name: "Jaka", Role: "Anak Desa", Type: "TRAVELER",
		X: -2.4, Z: 5.6, Yaw: 0.4, DialogueID: "jaka", InteractionRange: 2.6,
		Region: "village", Trait: "FUNNY", DialogueProfile: "child-play",
		VoiceLineID: "voice-jaka", Occupation: "CHILD",
		Schedule: phase24Sched("square", "school", "square", "home"),
		Personality: "Ceria lan penasaran.",
	})
	registerNPC(NPCDef{
		ID: "bagas", Name: "Bagas", Role: "Pengawal Alun-alun", Type: "TRAVELER",
		X: 1.2, Z: 8.4, Yaw: 4.2, DialogueID: "bagas", InteractionRange: 2.6,
		Region: "village", Trait: "BRAVE", DialogueProfile: "guard-soft",
		VoiceLineID: "voice-bagas", Occupation: "GUARD",
		Schedule: phase24Sched("gate", "square", "tavern", "home"),
		Personality: "Wani nanging alus.",
	})
	registerNPC(NPCDef{
		ID: "sari", Name: "Sari", Role: "Penjaga Cerita", Type: "STORY",
		X: -3.6, Z: 3.8, Yaw: 2.1, DialogueID: "sari", InteractionRange: 2.6,
		Region: "village", Trait: "MYSTERIOUS", DialogueProfile: "story-clue",
		VoiceLineID: "voice-sari", Occupation: "STORY NPC",
		Schedule: phase24Sched("hall", "square", "tavern", "home"),
		Personality: "Nyimpen clue, ora mbukak kabeh.",
	})
	registerNPC(NPCDef{
		ID: "wira", Name: "Wira", Role: "Pengembara", Type: "TRAVELER",
		X: 4.4, Z: 7.2, Yaw: 5.0, DialogueID: "wira", InteractionRange: 2.6,
		Region: "village", Trait: "CAUTIOUS", DialogueProfile: "traveler-warn",
		VoiceLineID: "voice-wira", Occupation: "TRAVELER",
		Schedule: phase24Sched("market", "gate", "tavern", "home"),
		Personality: "Ati-ati lan ngelingake.",
	})
}

func registerPhase24Dialogue() {
	overlayLoc("p24.karya.hello", "Le, nek arep ndandani gegamanmu, gawa mrene.", "Nak, kalau ingin memperbaiki perlengkapanmu, bawa kemari.")
	overlayLoc("p24.karya.where", "Kowe arep menyang ngendi?", "Kamu hendak ke mana?")
	overlayLoc("p24.karya.masjid", "Apik. Masjid kuwi tujuane lelakon. Ditindakake kanthi sopan, ya.", "Bagus. Masjid adalah tujuan perjalanan. Tempuhlah dengan hormat, ya.")
	overlayLoc("p24.karya.friend", "Golek kanca iku apik. Lelakon ora kanggo siji wong.", "Mencari kawan itu baik. Perjalanan bukan untuk seorang diri.")
	overlayLoc("p24.karya.unsure", "Durung ngerti kuwi biasa. Mlaku alon, takon wong desa.", "Belum mengerti itu biasa. Jalan pelan, tanya orang desa.")
	overlayLoc("p24.karya.memory", "Eh, kowe sing mbiyen nulungi aku, ta?", "Oh, kamu yang dulu menolongku, ya?")
	overlayLoc("p24.karya.help", "Matur nuwun, Le. Kayu iki kanggo pandhe.", "Terima kasih, Nak. Kayu ini untuk pandai besi.")
	overlayLoc("p24.karya.q1", "Le, gawa kayu kanggo pandhe, ya.", "Nak, bawakan kayu untuk pandai besi, ya.")
	overlayLoc("p24.karya.q2", "Saiki nggawe gegaman saka kayu kuwi.", "Sekarang buat perlengkapan dari kayu itu.")
	overlayLoc("p24.karya.q3", "Desa kudu dijaga. Aja kasar, nanging wani.", "Desa harus dijaga. Jangan kasar, tetapi berani.")
	overlayLoc("p24.karya.morning", "Esuk isih adhem. Pasar lagi mbukak.", "Pagi masih sejuk. Pasar baru buka.")
	overlayLoc("p24.karya.evening", "Sorene, wong desa kumpul ing alun-alun.", "Sorenya, orang desa berkumpul di alun-alun.")
	overlayLoc("p24.karya.rain", "Udan teka. Gawa klambi, aja kesusu.", "Hujan datang. Bawa pakaian, jangan terburu-buru.")
	overlayLoc("p24.karya.sun", "Srengenge padhang. Apik kanggo mlaku.", "Matahari cerah. Bagus untuk berjalan.")
	overlayLoc("p24.karya.storm", "Angin kenceng. Lungguh dhisik ing omah.", "Angin kencang. Duduk dulu di rumah.")
	overlayLoc("p24.karya.event", "Ati-ati! Ana makhluk saka alas!", "Hati-hati! Ada makhluk dari hutan!")
	overlayLoc("p24.karya.boss", "Ana bebaya gedhe teka. Jaga kanca, aja mungsuh dhewekan.", "Ada bahaya besar datang. Jaga kawan, jangan hadapi sendirian.")
	overlayLoc("p24.karya.safe", "Alas saiki luwih ayem. Tetep eling dalan.", "Hutan sekarang lebih tenang. Tetap ingat jalan.")
	overlayLoc("p24.karya.danger", "Alas isih mbebayani. Mlaku bareng kanca.", "Hutan masih berbahaya. Jalan bersama kawan.")
	overlayLoc("p24.guru.hello", "Le, ayo sinau bareng. Soal iki bagean critane.", "Nak, ayo belajar bersama. Soal ini bagian dari cerita.")
	overlayLoc("p24.guru.wrong", "Durung apa-apa, ayo diitung bareng.", "Belum apa-apa, ayo dihitung bersama.")
	overlayLoc("p24.guru.ok", "Pinter. Kawruhmu mundhak.", "Pintar. Pengetahuanmu bertambah.")
	overlayLoc("p24.guru.wait", "Istimewa. Ngaso dhisik, mengko sinau maneh.", "Bagus. Istirahat dulu, nanti belajar lagi.")
	overlayLoc("p24.laras.hello", "Le, kebon butuh tulung. Aja kesusu, nandur alon.", "Nak, kebun butuh bantuan. Jangan terburu, tanam pelan.")
	overlayLoc("p24.jaka.hello", "He, dolanan bareng, ya? Aku iso nyritakake dalan cilik.", "Hei, main bersama, ya? Aku bisa ceritakan jalan kecil.")
	overlayLoc("p24.bagas.hello", "Gerbang dijaga. Yen ana siluman, kandha aku.", "Gerbang dijaga. Kalau ada siluman, beri tahuku.")
	overlayLoc("p24.sari.hello", "Crita durung rampung. Rungokna, aja ngoyak kabeh rahasia.", "Cerita belum selesai. Dengarkan, jangan kejar semua rahasia.")
	overlayLoc("p24.wira.hello", "Dalan adoh. Ati-ati ing kabut.", "Jalan jauh. Hati-hati di kabut.")
	overlayLoc("p24.wira.dalan", "Le, dalan menyang alas isih adoh.", "Nak, jalan menuju hutan masih jauh.")
	overlayLoc("p24.choice.masjid", "Nyari masjid.", "Mencari masjid.")
	overlayLoc("p24.choice.friend", "Nyari kanca.", "Mencari kawan.")
	overlayLoc("p24.choice.unsure", "Aku durung ngerti.", "Aku belum mengerti.")
	overlayLoc("p24.choice.help", "Aku arep nulung.", "Aku ingin menolong.")
	overlayLoc("p24.cin.awal1", "Lelakon dimulai saka desa. Tujuane: dalan menyang masjid.", "Perjalanan dimulai dari desa. Tujuannya: jalan menuju masjid.")
	overlayLoc("p24.cin.awal2", "Nulung kanca, sinau, lan wani mlaku.", "Menolong kawan, belajar, dan berani melangkah.")
	overlayLoc("p24.cin.mist1", "Kabut Mistwood nyimpen jeneng, dudu musuh.", "Kabut Mistwood menyimpan nama, bukan musuh.")
	overlayLoc("p24.impact", "Pilihanmu memengaruhi hubungan dan cerita.", "Pilihanmu memengaruhi hubungan dan cerita.")
	dialogueCatalog["mbah_guru"] = DialogueLine{Speaker: "Mbah Guru", Text: "Le, ayo sinau bareng. Soal iki bagean critane.", VoiceLineID: "voice-mbah-guru"}
	dialogueCatalog["laras"] = DialogueLine{Speaker: "Laras", Text: "Le, kebon butuh tulung. Aja kesusu, nandur alon.", VoiceLineID: "voice-laras"}
	dialogueCatalog["jaka"] = DialogueLine{Speaker: "Jaka", Text: "He, dolanan bareng, ya?", VoiceLineID: "voice-jaka"}
	dialogueCatalog["bagas"] = DialogueLine{Speaker: "Bagas", Text: "Gerbang dijaga. Yen ana siluman, kandha aku.", VoiceLineID: "voice-bagas"}
	dialogueCatalog["sari"] = DialogueLine{Speaker: "Sari", Text: "Crita durung rampung. Rungokna, aja ngoyak kabeh rahasia.", VoiceLineID: "voice-sari"}
	dialogueCatalog["wira"] = DialogueLine{Speaker: "Wira", Text: "Le, dalan menyang alas isih adoh.", VoiceLineID: "voice-wira"}
}

func registerPhase24Quests() {
	registerQuest(QuestDef{
		ID: "nq-karya-1", Title: "Kayu kanggo Pandhe.", Kind: "side", NPC: "mbah_karya",
		Location: "Dawn Village", Description: "Tulung Mbah Karya nggawa kayu kanggo pandhe.",
		Objectives: []ObjectiveDef{{Type: "HELP", Target: "mbah_karya", Count: 1, Text: "Tulung Mbah Karya"}},
		Rewards:    RewardDef{Exp: 18, Coin: 4, EduToken: 0, Knowledge: 1},
		FlagsOnClaim: []string{"helped_villagers", "met_mbah_karya"}, ClaimAt: "mbah_karya",
	})
	registerQuest(QuestDef{
		ID: "nq-karya-2", Title: "Nggawe Gegaman.", Kind: "side", NPC: "mbah_karya",
		Location: "Dawn Village", Description: "Nggawe gegaman bareng Mbah Karya. Ora kanggo perang kasar.",
		Prereq:     []string{"nq-karya-1"},
		Objectives: []ObjectiveDef{{Type: "TALK", Target: "mbah_karya", Count: 1, Text: "Gunem karo Mbah Karya"}},
		Rewards:    RewardDef{Exp: 22, Coin: 6, Knowledge: 1},
		FlagsOnClaim: []string{"karya_gear_done"}, ClaimAt: "mbah_karya",
	})
	registerQuest(QuestDef{
		ID: "nq-karya-3", Title: "Njaga Desa.", Kind: "side", NPC: "mbah_karya",
		Location: "Dawn Village", Description: "Njaga desa kanthi wani lan alus.",
		Prereq:     []string{"nq-karya-2"},
		Objectives: []ObjectiveDef{{Type: "TALK", Target: "mbah_karya", Count: 1, Text: "Njaga desa bareng"}},
		Rewards:    RewardDef{Exp: 28, Coin: 8, Knowledge: 1},
		FlagsOnClaim: []string{"guarded_village", "forest_safe_story"}, ClaimAt: "mbah_karya", CinematicID: "cin-ch-mistwood",
	})
	registerQuest(QuestDef{
		ID: "nq-guru-1", Title: "Sinau Ngitung.", Kind: "education", NPC: "mbah_guru",
		Location: "Sekolah Desa", Description: "Jawab soal apel bareng Mbah Guru. Kelas 1.",
		Objectives: []ObjectiveDef{{Type: "ANSWER", Target: "q-apel-3-2", Count: 1, Text: "Jawab 3 apel + 2"}},
		Rewards:    RewardDef{Exp: 16, EduToken: 1, Knowledge: 1},
		FlagsOnClaim: []string{"answered_math_question", "edu_ngitung"}, ClaimAt: "mbah_guru",
	})
	registerQuest(QuestDef{
		ID: "nq-guru-2", Title: "Sinau Maca.", Kind: "education", NPC: "mbah_guru",
		Location: "Sekolah Desa", Description: "Sinau huruf A. Kelas 1.",
		Prereq:     []string{"nq-guru-1"},
		Objectives: []ObjectiveDef{{Type: "ANSWER", Target: "q-huruf-a", Count: 1, Text: "Kenali huruf A"}},
		Rewards:    RewardDef{Exp: 16, EduToken: 1, Knowledge: 1},
		FlagsOnClaim: []string{"edu_maca"}, ClaimAt: "mbah_guru",
	})
	registerQuest(QuestDef{
		ID: "nq-guru-3", Title: "Sinau Basa Jawa.", Kind: "education", NPC: "mbah_guru",
		Location: "Sekolah Desa", Description: "Sinau werna ijo. Kelas 1.",
		Prereq:     []string{"nq-guru-2"},
		Objectives: []ObjectiveDef{{Type: "ANSWER", Target: "q-werna-ijo", Count: 1, Text: "Ijo iku werna apa"}},
		Rewards:    RewardDef{Exp: 18, EduToken: 1, Knowledge: 1},
		FlagsOnClaim: []string{"edu_jawa", "edu_story_done"}, ClaimAt: "mbah_guru",
	})
}

func registerPhase24Education() {
	registerQuestion(QuestionDef{
		ID: "q-langit-biru", Category: "Pengetahuan umum", Grade: 1,
		Prompt:  "Langit awan cerah warnane?",
		Choices: []string{"Ijo", "Biru", "Abang"}, Correct: 1,
		Explain: "Langit cerah berwarna biru.",
	})
}

func registerPhase24Lore() {
	registerLore(LoreDef{ID: "lore-world-asal", Title: "Asal Donya", Region: "village", Kind: "history",
		X: 210, Z: 410, Text: "Donya iki diwiwiti saka cahya cilik ing desa. Wong mlaku, sinau, lan nulung, supaya dalan menyang masjid ketemu."})
	registerLore(LoreDef{ID: "lore-masjid-tujuan", Title: "Dalan Menyang Masjid", Region: "horizon", Kind: "culture",
		X: 214, Z: 414, Text: "Masjid Cahaya kuwi tujuane lelakon. Dituju kanthi sopan, dudu papan perang, dudu lelucon."})
	registerLore(LoreDef{ID: "lore-npc-karya", Title: "Mbah Karya", Region: "village", Kind: "npc",
		X: 216, Z: 416, Text: "Mbah Karya njaga pandhe desa. Gegaman kanggo njaga, dudu kanggo nyakiti."})
	registerLore(LoreDef{ID: "lore-33-clue", Title: "Clue 33 Siluman", Region: "forest", Kind: "history",
		X: 218, Z: 418, Text: "Telung puluh telu jeneng isih kasimpen. Saben chapter mung mbukak sethithik."})
	for _, e := range phase24Siluman() {
		registerLore(LoreDef{
			ID: e.id, Title: e.name, Region: e.region, Kind: "siluman", EnemyName: e.name,
			Personality: e.personality, Mechanic: e.mechanic, X: 220 + float64(e.idx), Z: 420,
			Text: e.lore,
		})
	}
}

type p24Siluman struct {
	idx                  int
	id, name, region     string
	personality, mechanic, lore string
}

func phase24Siluman() []p24Siluman {
	return []p24Siluman{
		{1, "lore-siluman-01", "Klawu", "forest", "hati-hati", "kabut pelan", "Klawu nyimpen dalan ing kabut. Ora medeni, mung ngelingake alon-alon."},
		{2, "lore-siluman-02", "Grembyang", "plains", "ceria", "tiupan angin", "Grembyang nyurung angin alus. Isa mundur yen kalah wicara."},
		{3, "lore-siluman-03", "Watu Geni", "valley", "teguh", "zona hangat", "Watu Geni anget, dudu geni perang. Ngajarake aja kesusu."},
		{4, "lore-siluman-04", "Rembulan", "forest", "penjaga malam", "cahaya malam", "Rembulan madhangi dalan wengi supaya bocah ora kesasar."},
		{5, "lore-siluman-05", "Gandring", "forest", "penjaga groves", "jerat daun", "Gandring njaga wit. Yen dihormati, dalan mbukak."},
		{6, "lore-siluman-06", "Silem", "plains", "tenang", "percikan air", "Silem nyimpen kali. Nyilem, banjur metu adoh."},
		{7, "lore-siluman-07", "Kutilang", "plains", "lincah", "lompat cepat", "Kutilang mabur cendhak. Nggodha, banjur mlayu."},
		{8, "lore-siluman-08", "Lumut", "forest", "sabar", "langkah licin", "Lumut ngelingake mlaku alon ing watu teles."},
		{9, "lore-siluman-09", "Alun", "plains", "lembut", "ombak kecil", "Alun nyurung sethithik, kaya ombak kali."},
		{10, "lore-siluman-10", "Jenar", "canyon", "waspada", "debu kilau", "Jenar nyebar pasir cemlorot, dudu tatu."},
		{11, "lore-siluman-11", "Riwayat", "ruins", "pendendam lembut", "gaung panggilan", "Riwayat njaluk kanca kanthi swara lawas."},
		{12, "lore-siluman-12", "Kebul", "forest", "pemalu", "asap sembunyi", "Kebul ndhelik yen wedi. Aja ngoyak kasar."},
		{13, "lore-siluman-13", "Tirta", "plains", "penolong", "panggil kanca", "Tirta njaluk kanca saka sumber banyu."},
		{14, "lore-siluman-14", "Gendhis", "village", "ramah", "pikat manis", "Gendhis nggodha karo rasa manis, banjur ngajar aja rakus."},
		{15, "lore-siluman-15", "Liring", "ruins", "waspada", "mundur cermin", "Liring mundur yen ketok. Ngajarake aja kesusu nyerang."},
		{16, "lore-siluman-16", "Pepet", "canyon", "teguh", "jalur sempit", "Pepet njaga celah. Liwati siji-siji."},
		{17, "lore-siluman-17", "Kidung", "village", "ceria", "nyanyian godaan", "Kidung nembang. Iku godaan, dudu umpatan."},
		{18, "lore-siluman-18", "Wana", "forest", "penjaga", "akar pelan", "Wana nahan langkah supaya wong mikir."},
		{19, "lore-siluman-19", "Suryalemah", "canyon", "teguh", "kilat panas", "Suryalemah anget ing lemah. Nggoleki teduh."},
		{20, "lore-siluman-20", "Embun", "forest", "lembut", "embun pelan", "Embun nglambatake. Istirahat, banjur mlaku."},
		{21, "lore-siluman-21", "Bayulembut", "plains", "pemalu", "mundur angin", "Bayulembut mlayu yen kalah ramah."},
		{22, "lore-siluman-22", "Candrahalim", "forest", "penjaga malam", "kabut bulan", "Candrahalim nyampur cahya wengi karo kabut."},
		{23, "lore-siluman-23", "Lintangcilik", "horizon", "ceria", "percik bintang", "Lintangcilik mung kilatan cilik, kanggo tenger."},
		{24, "lore-siluman-24", "Pasirhalus", "canyon", "waspada", "debu halus", "Pasirhalus nutupi jejak. Goleki tenger watu."},
		{25, "lore-siluman-25", "Daunturu", "forest", "tenang", "daun tidur", "Daunturu ngajak ngaso. Dudu nggawe wedi."},
		{26, "lore-siluman-26", "Kilatcilik", "plains", "lincah", "kilat kecil", "Kilatcilik nyentuh banjur lunga. Telegraph cendhak."},
		{27, "lore-siluman-27", "Tlagasuram", "forest", "penjaga", "tarikan danau", "Tlagasuram nahan ing pinggir tlaga. Aja nyilem dhewe."},
		{28, "lore-siluman-28", "Gunungsepi", "valley", "teguh", "hentakan lembut", "Gunungsepi ngelingake langkah abot. Telegraph dawa."},
		{29, "lore-siluman-29", "Jalansesat", "ruins", "penyesat lembut", "salah arah", "Jalansesat muter dalan. Clue ana ing papan."},
		{30, "lore-siluman-30", "Bendungan", "plains", "penjaga", "perisai blok", "Bendungan nahan, dudu nyerang. Tunggu bolongane."},
		{31, "lore-siluman-31", "Kalaangin", "plains", "pemanggil", "panggil kanca", "Kalaangin njaluk kanca angin. Aja nguber kabeh."},
		{32, "lore-siluman-32", "Banyumurni", "horizon", "penolong", "bersih pelan", "Banyumurni ngresiki kabut sethithik."},
		{33, "lore-siluman-33", "Penjagagerbang", "ruins", "penjaga gerbang", "panggil kanca", "Penjagagerbang njaga lawang 33. Jenenge dudu musuh utama."},
	}
}

func registerPhase24Cinematics() {
	registerCinematic(CinematicDef{
		ID: "cin-ch-awal", Title: "Awal Perjalanan", DurationSec: 8, Skippable: true,
		Camera: "follow", Music: "story-dawn",
		Lines: []string{"p24.cin.awal1", "p24.cin.awal2"},
	})
	registerCinematic(CinematicDef{
		ID: "cin-ch-mistwood", Title: "Kabut Mistwood", DurationSec: 8, Skippable: true,
		Camera: "pan", Music: "story-mist",
		Lines: []string{"p24.cin.mist1"},
	})
	registerCinematic(CinematicDef{
		ID: "cin-ch-outro", Title: "Apa yang sudah terjadi", DurationSec: 6, Skippable: true,
		Camera: "focus", Music: "story-dawn",
		Lines: []string{"p24.cin.awal2"},
	})
}

func registerPhase24Meta() {
	registerTitleDef(TitleDef{ID: "kanca-desa", Name: "Kanca Desa", Source: "npc"})
	registerTitleDef(TitleDef{ID: "peneliti-siluman", Name: "Peneliti Siluman", Source: "lore"})
	registerTitleDef(TitleDef{ID: "murid-cahaya", Name: "Murid Cahaya", Source: "education"})
	registerCosmeticDef(CosmeticDef{ID: "cloak-karya", Name: "Karya Cloak", Kind: "cloak"})
	registerAchievementDef(AchievementDef{ID: "helped-karya", Name: "Tulung Pandhe", Title: "kanca-desa", Flag: "helped_villagers", Category: "Story"})
}

func phase24NPC(id string) bool {
	switch id {
	case "mbah_karya", "mbok_rasa", "pak_jala", "mbah_batu", "mbah_guru", "laras", "jaka", "bagas", "sari", "wira":
		return true
	default:
		return false
	}
}

func phase24SkipCore(id string) bool {
	switch id {
	case "mira", "elder_ardan", "lio", "nara", "raven", "mbah_jagat", "ibu_desa", "petani", "child_lina", "dawn_merchant":
		return true
	default:
		return false
	}
}

func relTier(xp int) string {
	switch {
	case xp >= 60:
		return RelTrusted
	case xp >= 30:
		return RelFriend
	case xp >= 10:
		return RelAcquaintance
	default:
		return RelStranger
	}
}

func relNext(xp int) string {
	switch {
	case xp < 10:
		return "Dialog anyar"
	case xp < 30:
		return "Gelar Kanca Desa"
	case xp < 60:
		return "Cloak Karya"
	default:
		return "Wis dipercaya"
	}
}

func (p *Player) bumpNpcRel(id string, n int) {
	if n <= 0 || id == "" {
		return
	}
	log := p.ensureLog()
	if log.NpcRel == nil {
		log.NpcRel = map[string]int{}
	}
	log.NpcRel[id] += n
	if log.NpcRel[id] > 100 {
		log.NpcRel[id] = 100
	}
	xp := log.NpcRel[id]
	if xp >= 30 {
		p.grantTitle("kanca-desa")
	}
	if xp >= 60 {
		p.grantCosmetic("cloak-karya")
	}
}

func (p *Player) rememberNPC(id string) {
	log := p.ensureLog()
	if log.NpcMemory == nil {
		log.NpcMemory = map[string]bool{}
	}
	log.NpcMemory[id] = true
	if log.Flags == nil {
		log.Flags = map[string]bool{}
	}
	log.Flags["helped_villagers"] = true
}

func (p *Player) appendDialogueHistory(line string) {
	if line == "" {
		return
	}
	log := p.ensureLog()
	log.DialogueHistory = append(log.DialogueHistory, line)
	if len(log.DialogueHistory) > 24 {
		log.DialogueHistory = log.DialogueHistory[len(log.DialogueHistory)-24:]
	}
}

func p24Lang(p *Player) string {
	return playerLang(p)
}

func p24Pair(key string, p *Player) (text, sub string) {
	lang := p24Lang(p)
	jv := locText(key, "jv")
	idn := locText(key, "id")
	if lang == "id" {
		return idn, jv
	}
	return jv, idn
}

func forestMood(p *Player) string {
	log := p.ensureLog()
	if log.Flags["forest_safe_story"] || log.Flags["chapter1_complete"] || log.Flags["story_chapter_1_complete"] || log.Flags["st-ch01_complete"] {
		return MoodForestSafe
	}
	return MoodForestDanger
}

func (w *WorldState) applyPhase24Talk(p *Player, npc NPCDef, res *InteractResult) {
	if res == nil || phase24SkipCore(npc.ID) {
		return
	}
	log := p.ensureLog()
	if log.Flags == nil {
		log.Flags = map[string]bool{}
	}
	lang := p24Lang(p)
	res.VoiceID = npc.VoiceLineID
	if res.VoiceID == "" {
		res.VoiceID = "voice-" + npc.ID
	}
	if npc.ID == "mbah_karya" && log.Flags["met_mbah_karya"] == false {
		log.Flags["met_mbah_karya"] = true
	}
	if npc.ID == "wira" && p.ZoneID == "forest" {
		log.Flags["found_mistwood"] = true
	}

	key := phase24LineKey(w, p, npc)
	text, sub := p24Pair(key, p)
	if text != "" && (phase24NPC(npc.ID) || npc.ID == "pak_jaga" && res.Text == "") {
		if npc.ID != "pak_jaga" {
			res.Text, res.Subtitle = text, sub
		}
	}
	if npc.ID == "pak_jaga" && res.Text == "" {
		res.Text = lineOr("pak_jaga", "Ati-ati, Le! Ana siluman teka saka alas!")
	}

	res.Emotion, res.Gesture = phase24Emotion(w, p, npc)
	if log.NpcMemory[npc.ID] && npc.ID == "mbah_karya" {
		res.Emotion, res.Gesture = "happy", "wave"
	}

	opts := res.Options
	if phase24NPC(npc.ID) {
		opts = phase24Options(p, npc, lang, opts)
	}
	res.Options = opts
	res.History = append([]string{}, log.DialogueHistory...)
	p.appendDialogueHistory(res.Text)
	if npc.ID == "mbah_karya" || npc.ID == "sari" {
		w.maybeDiscoverLore(p, "lore-npc-karya")
		w.maybeDiscoverLore(p, "lore-world-asal")
		w.maybeDiscoverLore(p, "lore-masjid-tujuan")
	}
	if npc.ID == "wira" {
		w.maybeDiscoverLore(p, "lore-siluman-01")
		w.maybeDiscoverLore(p, "lore-33-clue")
	}
	p.markDirty()
	w.persist(p)
}

func phase24LineKey(w *WorldState, p *Player, npc NPCDef) string {
	log := p.ensureLog()
	if npc.ID == "mbah_karya" && log.NpcMemory["mbah_karya"] {
		return "p24.karya.memory"
	}
	if w.Events.Active != nil && (w.Events.Active.State == "ACTIVE" || w.Events.Active.State == "FINAL") {
		return "p24.karya.event"
	}
	if w.Boss != nil && (w.Boss.State == "ACTIVE" || w.Boss.State == "ANNOUNCED") {
		return "p24.karya.boss"
	}
	weather := strings.ToUpper(w.Time.Weather)
	if npc.ID != "mbah_karya" {
		switch weather {
		case "RAIN":
			return "p24.karya.rain"
		case "STORM":
			return "p24.karya.storm"
		}
		switch w.Time.Slot() {
		case "morning":
			if npc.ID == "laras" {
				return "p24.karya.morning"
			}
		case "evening", "night":
			if npc.ID == "laras" {
				return "p24.karya.evening"
			}
		}
		if forestMood(p) == MoodForestSafe && (npc.ID == "wira" || npc.ID == "bagas") {
			return "p24.karya.safe"
		}
		if forestMood(p) == MoodForestDanger && (npc.ID == "wira" || npc.ID == "bagas") {
			return "p24.karya.danger"
		}
	} else if weather == "RAIN" {
		return "p24.karya.rain"
	} else if weather == "STORM" {
		return "p24.karya.storm"
	}
	switch npc.ID {
	case "mbah_karya":
		return "p24.karya.where"
	case "mbah_guru":
		if log.EduAskAt > 0 && time.Now().Unix()-log.EduAskAt < EduAskCooldownSec {
			return "p24.guru.wait"
		}
		return "p24.guru.hello"
	case "laras":
		return "p24.laras.hello"
	case "jaka":
		return "p24.jaka.hello"
	case "bagas":
		return "p24.bagas.hello"
	case "sari":
		return "p24.sari.hello"
	case "wira":
		return "p24.wira.dalan"
	default:
		return ""
	}
}

func phase24Emotion(w *WorldState, p *Player, npc NPCDef) (emotion, gesture string) {
	if w.Events.Active != nil && (w.Events.Active.State == "ACTIVE" || w.Events.Active.State == "FINAL") {
		return "worried", "point"
	}
	if forestMood(p) == MoodForestDanger && (npc.ID == "wira" || npc.ID == "bagas") {
		return "worried", "point"
	}
	switch npc.Trait {
	case "FUNNY":
		return "happy", "laugh"
	case "WISE":
		return "happy", "think"
	case "KIND":
		return "happy", "wave"
	case "STRICT":
		return "surprised", "sit"
	default:
		return "happy", "wave"
	}
}

func phase24Options(p *Player, npc NPCDef, lang string, base []DialogOption) []DialogOption {
	opts := make([]DialogOption, 0, 8)
	if npc.ID == "mbah_karya" {
		opts = append(opts,
			DialogOption{ID: "choice-where-a", Label: locText("p24.choice.masjid", lang)},
			DialogOption{ID: "choice-where-b", Label: locText("p24.choice.friend", lang)},
			DialogOption{ID: "choice-where-c", Label: locText("p24.choice.unsure", lang)},
			DialogOption{ID: "choice-help", Label: locText("p24.choice.help", lang)},
		)
		opts = append(opts, questOption(p, "nq-karya-1", "Kayu kanggo Pandhe.")...)
		opts = append(opts, questOption(p, "nq-karya-2", "Nggawe Gegaman.")...)
		opts = append(opts, questOption(p, "nq-karya-3", "Njaga Desa.")...)
		opts = append(opts, phase25KaryaOptions(p)...)
		opts = append(opts, DialogOption{ID: "cin:cin-ch-awal", Label: "Crita"})
	}
	if npc.ID == "mbah_guru" {
		opts = append(opts, DialogOption{ID: "quiz-guru", Label: "Jawab soal"})
		opts = append(opts, questOption(p, "nq-guru-1", "Sinau Ngitung.")...)
		opts = append(opts, questOption(p, "nq-guru-2", "Sinau Maca.")...)
		opts = append(opts, questOption(p, "nq-guru-3", "Sinau Basa Jawa.")...)
	}
	seen := map[string]bool{}
	out := make([]DialogOption, 0, len(opts)+len(base)+1)
	for _, o := range append(opts, base...) {
		if o.ID == "" || seen[o.ID] {
			continue
		}
		seen[o.ID] = true
		out = append(out, o)
	}
	if !seen["close"] {
		out = append(out, DialogOption{ID: "close", Label: "Tutup"})
	}
	return out
}

func (w *WorldState) phase24Choice(p *Player, npc NPCDef, choice string) [][]byte {
	log := p.ensureLog()
	if log.StoryChoices == nil {
		log.StoryChoices = map[string]string{}
	}
	if log.DialogueChoices == nil {
		log.DialogueChoices = map[string]string{}
	}
	option := strings.ToLower(choice)
	key := "p24.karya.where"
	switch option {
	case "choice-where-a", "choice-a":
		if log.StoryChoices["dlg-where"] == "" {
			log.StoryChoices["dlg-where"] = "A"
		}
		log.DialogueChoices["dlg-where"] = "A"
		log.Flags["seek_masjid"] = true
		key = "p24.karya.masjid"
		p.bumpNpcRel(npc.ID, 4)
		p.grantKnowledge(1)
		w.maybeDiscoverLore(p, "lore-masjid-tujuan")
	case "choice-where-b", "choice-b":
		if log.StoryChoices["dlg-where"] == "" {
			log.StoryChoices["dlg-where"] = "B"
		}
		log.DialogueChoices["dlg-where"] = "B"
		key = "p24.karya.friend"
		p.bumpNpcRel(npc.ID, 6)
	case "choice-where-c", "choice-c":
		if log.StoryChoices["dlg-where"] == "" {
			log.StoryChoices["dlg-where"] = "C"
		}
		log.DialogueChoices["dlg-where"] = "C"
		key = "p24.karya.unsure"
		p.bumpNpcRel(npc.ID, 2)
	case "choice-help":
		log.DialogueChoices[npc.ID+"-help"] = "help"
		p.rememberNPC(npc.ID)
		p.bumpNpcRel(npc.ID, 8)
		p.credit("HELP", npc.ID, 1)
		key = "p24.karya.help"
		w.maybeDiscoverLore(p, "lore-npc-karya")
	case "choice-train":
		log.DialogueChoices[npc.ID+"-train"] = "train"
		p.credit("TALK", "mbah_karya", 1)
		key = "p25.karya.punch"
		if q := p.quest("nq-ascend-1"); q != nil && q.State == QuestActive {
			key = "p25.karya.ngudi"
		}
		p.bumpNpcRel(npc.ID, 4)
	default:
		key = "p24.karya.where"
	}
	text, sub := p24Pair(key, p)
	p.appendDialogueHistory(text)
	p.markDirty()
	w.persist(p)
	res := InteractResult{
		Kind: "npc", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
		Text: text, Subtitle: sub, VoiceID: npc.VoiceLineID, Emotion: "happy", Gesture: "wave",
		Options: phase24Options(p, npc, p24Lang(p), nil), History: append([]string{}, log.DialogueHistory...),
	}
	return [][]byte{marshal(TypeInteractResult, res)}
}

func (w *WorldState) startPhase24Quiz(p *Player, npc NPCDef) [][]byte {
	log := p.ensureLog()
	if log.EduAskAt > 0 && time.Now().Unix()-log.EduAskAt < EduAskCooldownSec && log.Flags["answered_math_question"] {
		text, sub := p24Pair("p24.guru.wait", p)
		return [][]byte{marshal(TypeInteractResult, InteractResult{
			Kind: "npc", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
			Text: text, Subtitle: sub, Options: phase24Options(p, npc, p24Lang(p), nil),
		})}
	}
	qid := "q-apel-3-2"
	questID := "nq-guru-1"
	if q := p.quest("nq-guru-2"); q != nil && (q.State == QuestActive || q.State == QuestAvailable) && p.quest("nq-guru-1") != nil && p.quest("nq-guru-1").State == QuestClaimed {
		qid, questID = "q-huruf-a", "nq-guru-2"
	}
	if q := p.quest("nq-guru-3"); q != nil && (q.State == QuestActive || q.State == QuestAvailable) && p.quest("nq-guru-2") != nil && p.quest("nq-guru-2").State == QuestClaimed {
		qid, questID = "q-werna-ijo", "nq-guru-3"
	}
	idx := -1
	for i, q := range questionCatalog {
		if q.ID == qid {
			idx = i
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeInteract, "question")
	}
	qlog := p.quest(questID)
	if qlog != nil && qlog.State == QuestAvailable {
		qlog.State = QuestActive
	}
	log.Quiz = QuizSession{QuestID: questID, Index: idx, Active: true}
	p.markDirty()
	w.persist(p)
	q := questionCatalog[idx]
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "edu-story", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
		Text: q.Prompt, VoiceID: npc.VoiceLineID,
		Question: &QuestionOut{ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices, ToID: p.ID},
	})}
}

func (w *WorldState) answerPhase24Edu(p *Player, in EducationAnswerIn) [][]byte {
	log := p.ensureLog()
	if !log.Quiz.Active {
		return rejectFor(p.ID, TypeEducationAnswer, "no_session")
	}
	idx := log.Quiz.Index
	if idx < 0 || idx >= len(questionCatalog) {
		return rejectFor(p.ID, TypeEducationAnswer, "question")
	}
	def := questionCatalog[idx]
	if in.QuestionID != "" && in.QuestionID != def.ID {
		return rejectFor(p.ID, TypeEducationAnswer, "order")
	}
	if in.Choice != def.Correct {
		explain := def.Explain
		if def.ID == "q-apel-3-2" {
			explain = locText("p24.guru.wrong", p24Lang(p)) + " " + def.Explain
		}
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: false, Explain: explain, Retry: true, Toast: locText("p24.guru.wrong", p24Lang(p)),
			Question: questionOut(idx),
		})}
	}
	coin, exp := log.Coin, p.Exp
	p.credit("ANSWER", def.ID, 1)
	p.grantKnowledge(1)
	w.grantEducationBonus(p)
	p.bumpNpcRel("mbah_guru", 10)
	log.Flags["answered_math_question"] = true
	log.EduAskAt = time.Now().Unix()
	log.Quiz.Active = false
	if q := p.quest(log.Quiz.QuestID); q != nil && q.State == QuestActive {
		if d, ok := questByID[q.ID]; ok && objectivesDone(q, d) {
			q.State = QuestCompleted
		}
	}
	if log.Flags["edu_jawa"] || (p.quest("nq-guru-3") != nil && p.quest("nq-guru-3").State == QuestCompleted) {
		p.grantTitle("murid-cahaya")
	}
	p.markDirty()
	w.persist(p)
	if log.Coin < coin {
		log.Coin = coin
	}
	if p.Exp < exp {
		p.Exp = exp
	}
	return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
		Correct: true, Explain: def.Explain, Toast: locText("p24.guru.ok", p24Lang(p)),
	}), marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase))}
}

func (w *WorldState) maybeDiscoverLore(p *Player, id string) {
	def, ok := loreByID[id]
	if !ok {
		return
	}
	log := p.ensureLog()
	if log.Lore[id] {
		return
	}
	w.discoverLore(p, def)
	n := 0
	for _, e := range phase24Siluman() {
		if log.Lore[e.id] {
			n++
		}
	}
	if n >= 5 {
		p.grantTitle("peneliti-siluman")
		if log.EduToken < 200 {
			log.EduToken++
		}
	}
}

func (w *WorldState) ApplyPhase24(p *Player, env Envelope) [][]byte {
	if p == nil {
		return nil
	}
	switch env.Type {
	case TypeGetAdventure:
		return [][]byte{marshal(TypeWorldJournal, p.worldJournal(w))}
	case TypeSetRelationship, TypeAddRelationship, TypeSetNpcMemory, TypeUnlockLore,
		TypeCompleteQuest, TypeSetWorldState, TypeClaimEducationReward:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func attachPhase24Journal(p *Player, w *WorldState, j *WorldJournal) {
	if j == nil {
		return
	}
	log := p.ensureLog()
	j.WorldMood = forestMood(p)
	j.NpcBook = phase24NpcBook(p)
	j.ChoiceHistory = phase24ChoiceHistory(log)
	j.OverlayChapters = phase24OverlayChapters(log)
	j.EnemyLore = phase24EnemyLore(log)
	cards := make([]LoreView, 0, len(loreCatalog))
	for _, l := range loreCatalog {
		if !strings.HasPrefix(l.ID, "lore-siluman") && !strings.HasPrefix(l.ID, "lore-world") && !strings.HasPrefix(l.ID, "lore-masjid") && !strings.HasPrefix(l.ID, "lore-npc") && !strings.HasPrefix(l.ID, "lore-33") {
			continue
		}
		cards = append(cards, LoreView{
			ID: l.ID, Title: l.Title, Text: l.Text, Region: l.Region, Kind: l.Kind,
			Discovered: log.Lore[l.ID], Personality: l.Personality, Mechanic: l.Mechanic,
		})
	}
	j.LoreCards = cards
}

func phase24NpcBook(p *Player) []NpcRelView {
	log := p.ensureLog()
	ids := []string{"mbah_karya", "mbok_rasa", "pak_jala", "mbah_batu", "mbah_guru", "laras", "jaka", "bagas", "sari", "wira"}
	out := make([]NpcRelView, 0, len(ids))
	for _, id := range ids {
		n := npcByID[id]
		if n.ID == "" {
			continue
		}
		xp := 0
		if log.NpcRel != nil {
			xp = log.NpcRel[id]
		}
		out = append(out, NpcRelView{
			ID: id, Name: n.Name, Role: n.Role, Trait: n.Trait, XP: xp,
			Relationship: relTier(xp), NextReward: relNext(xp), Memory: log.NpcMemory[id],
		})
	}
	return out
}

func phase24ChoiceHistory(log *PlayerLog) []ChoiceHistView {
	out := make([]ChoiceHistView, 0)
	if v := log.StoryChoices["dlg-where"]; v != "" {
		out = append(out, ChoiceHistView{ID: "dlg-where", Choice: v, Impact: locText("p24.impact", playerLangFromLog(log))})
	}
	for k, v := range log.StoryChoices {
		if k == "dlg-where" {
			continue
		}
		out = append(out, ChoiceHistView{ID: k, Choice: v, Impact: locText("p24.impact", playerLangFromLog(log))})
	}
	return out
}

func playerLangFromLog(log *PlayerLog) string {
	if log != nil && log.Language != "" {
		return log.Language
	}
	return "jv"
}

func phase24OverlayChapters(log *PlayerLog) []OverlayChapterView {
	titles := []string{
		"Awal Perjalanan", "Kabut Mistwood", "Lembah Batu", "Sungai Cahaya",
		"Dataran Merah", "Reruntuhan Kuno", "Gerbang 33 Siluman", "Jalan Menuju Cahaya",
	}
	out := make([]OverlayChapterView, 0, 8)
	for i, title := range titles {
		st := StoryLocked
		idx := i + 1
		if idx == 8 {
			out = append(out, OverlayChapterView{Index: idx, Title: title, State: StoryLocked, Locked: true})
			continue
		}
		sid := "st-ch0" + itoa(idx)
		st = storyChapterState(log, sid)
		out = append(out, OverlayChapterView{Index: idx, Title: title, State: st, Locked: st == StoryLocked})
	}
	return out
}

func phase24EnemyLore(log *PlayerLog) []EnemyLoreView {
	out := make([]EnemyLoreView, 0, 33)
	for _, e := range phase24Siluman() {
		enc := log.Lore[e.id] || log.Flags["met_"+e.id]
		def := log.Flags["defeated_"+e.id]
		out = append(out, EnemyLoreView{
			ID: e.id, Name: e.name, Region: e.region, Personality: e.personality, Mechanic: e.mechanic,
			Encountered: enc, Defeated: def, Discovered: log.Lore[e.id],
			Lore: func() string {
				if log.Lore[e.id] {
					return e.lore
				}
				return ""
			}(),
		})
	}
	return out
}
