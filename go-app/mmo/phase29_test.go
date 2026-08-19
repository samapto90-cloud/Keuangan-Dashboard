package mmo

import (
	"encoding/json"
	"testing"
)

func TestPhase29LandminesKept(t *testing.T) {
	if len(guardianCatalog) != 33 {
		t.Fatalf("guardians %d", len(guardianCatalog))
	}
	if silumanByGuard["ragha"].Name != "Jaladara" || silumanByGuard["avaron"].Name != "Adityara" {
		t.Fatal("keep Jaladara / Adityara story names")
	}
	if len(storyChapterCatalog) != 10 {
		t.Fatalf("story chapters %d", len(storyChapterCatalog))
	}
	if questionByID["q-add-4-3"].Correct != 1 || questionByID["q-apel-3-2"].Correct != 1 {
		t.Fatal("keep old education questions")
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatal("keep Dawn City")
	}
	if mountByID["wind-runner"].Speed != 1.45 {
		t.Fatal("keep wind runner speed")
	}
}

func TestPhase29RosterAndContentRegistered(t *testing.T) {
	if len(phase29Regions) != 8 {
		t.Fatalf("regions %d", len(phase29Regions))
	}
	if len(phase29Siluman) != 33 {
		t.Fatalf("siluman %d", len(phase29Siluman))
	}
	if phase29Siluman[0].Name != "Klowo Alas" || phase29Siluman[32].Name != "RAKSA CAHAYA PETENG" {
		t.Fatal("phase29 names")
	}
	if eventByID[Phase29WorldEventID].Name != "Bayangan Ing Alas" {
		t.Fatal("world event")
	}
	if worldBossByID[Phase29WorldBossID].Name != "RAKSA CAHAYA PETENG" {
		t.Fatal("world boss")
	}
	raid := dungeonByID[Phase29RaidID]
	if raid.ID == "" || raid.MinPlayers != 10 || raid.MaxPlayers != 20 || len(raid.Bosses) != 3 {
		t.Fatal("final raid")
	}
	if questByID["mq029"].Title != "Lampah Pungkasan." {
		t.Fatal("final quest")
	}
	if questionByID["q-add-2-4"].Correct != 1 || questionByID["q-sub-7-3"].Correct != 1 || questionByID["q-read-buku"].Correct != 0 {
		t.Fatal("phase29 education")
	}
}

func TestPhase29JournalAndStoryOverlay(t *testing.T) {
	w, p := testVillagePlayer()
	j := p.worldJournal(w)
	if len(j.OverlayChapters) != 8 {
		t.Fatalf("overlay chapters %d", len(j.OverlayChapters))
	}
	if j.OverlayChapters[0].Title == "" || j.OverlayChapters[7].Title == "" {
		t.Fatal("chapter roadmap")
	}
	if len(j.EnemyLore) != 33 {
		t.Fatalf("enemy lore %d", len(j.EnemyLore))
	}
	if j.WorldMood == "" {
		t.Fatal("journal phase29 overlay")
	}
	s := w.storyStateView(p)
	if s.FinalBoss["name"] != "RAKSA CAHAYA PETENG" {
		t.Fatal("final boss story overlay")
	}
	if s.Objective == "" {
		t.Fatal("story objective")
	}
}

func TestPhase29EndgameViewAndLeaderboards(t *testing.T) {
	w, a := testVillagePlayer()
	a.Level = 55
	a.ensurePhase21()
	log := a.ensureLog()
	log.SeasonXP = 120
	log.RecipeCrafts["rec-dawn-blade"] = 2
	log.ProfessionXP[ProfMiner] = 90
	log.Claimed["mq001"] = true
	log.Flags["guardian_ragha_defeated"] = true
	b := addWorldPlayer(w, "p_b", "Sinta")
	b.Level = 40
	view := w.endgameView(a)
	p29, ok := view["phase29"].(map[string]any)
	if !ok || p29["mainStory"] == nil || p29["raid"] == nil {
		t.Fatal("phase29 endgame block")
	}
	sil, ok := p29["siluman33"].([]phase29SilumanRow)
	if !ok || len(sil) != 33 || sil[0].Name != "Klowo Alas" {
		t.Fatal("phase29 roster")
	}
	boards, ok := view["leaderboards"].(map[string]any)
	if !ok || boards["level"] == nil || boards["combat"] == nil || boards["crafting"] == nil || boards["gathering"] == nil {
		t.Fatal("expanded leaderboards")
	}
	_ = b
}

func TestPhase29EventAndBossSecurity(t *testing.T) {
	w, p := testVillagePlayer()
	w.startWorldEvent(Phase29WorldEventID)
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeClaimEventReward, Data: []byte(`{"eventId":"bayangan-ing-alas","score":999}`)}), TypeClaimEventReward) {
		t.Fatal("event reward without participation")
	}
	w.spawnNamedWorldBoss(Phase29WorldBossID)
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeClaimWorldBoss, Data: []byte(`{"bossId":"raksa-cahaya-peteng","transactionId":"wb29","contribution":999}`)}), TypeClaimWorldBoss) {
		t.Fatal("world boss reward without participation")
	}
}

func TestPhase29RaidEntryRequirement(t *testing.T) {
	w, p := testVillagePlayer()
	p.Level = 20
	evs := w.ApplyDungeon(p.ID, Envelope{Type: TypeDungeonEnter, Data: []byte(`{"dungeonId":"raid-lampah-pungkasan","difficulty":"HARD"}`)})
	if !rejectAction(evs, TypeDungeonEnter) {
		t.Fatal("raid requirement reject")
	}
}

func TestPhase29EducationAnswerRetryAndReward(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureLog().Quiz = QuizSession{QuestID: "craft-edu", Active: true}
	wrong, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-kayu-3-2", Choice: 0})
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: wrong})
	if !payloadHas(evs, TypeEducationFeedback, "Coba") {
		t.Fatal("retry feedback")
	}
	p.ensureLog().Quiz = QuizSession{QuestID: "econ-edu", Active: true}
	before := p.ensureLog().EduToken
	ok, _ := json.Marshal(EducationAnswerIn{QuestionID: "q-emas-15-10", Choice: 1})
	evs = w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: ok})
	if rejectAction(evs, TypeEducationAnswer) {
		t.Fatal("correct answer")
	}
	if p.ensureLog().EduToken <= before {
		t.Fatal("education reward")
	}
}

