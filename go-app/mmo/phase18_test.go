package mmo

import (
	"encoding/json"
	"strings"
	"testing"
)

func payloadHas(evs [][]byte, typ, sub string) bool {
	for _, ev := range evs {
		var env Envelope
		if json.Unmarshal(ev, &env) != nil {
			continue
		}
		if typ != "" && env.Type != typ {
			continue
		}
		if sub == "" || strings.Contains(string(ev), sub) {
			return true
		}
	}
	return false
}

func TestPhase18StoryServiceAndChapters(t *testing.T) {
	if len(storyChapterCatalog) != 10 {
		t.Fatalf("chapters %d", len(storyChapterCatalog))
	}
	if storyChapterCatalog[0].Title != "Awal Perjalanan" || storyChapterCatalog[9].Title != "Perjalanan Menuju Masjid" {
		t.Fatal("chapter titles")
	}
	_, p := testVillagePlayer()
	log := p.ensureLog()
	if log.StoryState != StoryNotStarted || log.Language != "jv" {
		t.Fatalf("default state %s lang %s", log.StoryState, log.Language)
	}
	if storyChapterState(log, "st-ch02") != StoryLocked {
		t.Fatal("ch2 locked")
	}
}

func TestPhase18MbahJagatJawaAndLanguage(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 1.6, 4.2
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_jagat","kind":"talk"}`)})
	blob := ""
	for _, ev := range evs {
		blob += string(ev)
	}
	if !strings.Contains(blob, "dalanmu isih adoh") && !strings.Contains(blob, "wektu wis teka") {
		t.Fatalf("jawa: %s", blob)
	}
	if !strings.Contains(blob, "perjalananmu masih jauh") && !strings.Contains(blob, "waktunya sudah tiba") {
		t.Fatalf("subtitle id: %s", blob)
	}
	w.ApplyExplore(p.ID, Envelope{Type: TypeSetLanguage, Data: []byte(`{"language":"id"}`)})
	if p.ensureLog().Language != "id" {
		t.Fatal("lang id")
	}
	evs = w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_jagat","kind":"talk"}`)})
	raw := string(evs[0])
	if !strings.Contains(raw, "perjalananmu masih jauh") && !strings.Contains(raw, "waktunya sudah tiba") {
		t.Fatalf("indonesia: %s", raw)
	}
	if strings.Contains(raw, "dalanmu isih adoh") && strings.Contains(raw, `"text":"Le, dalanmu`) {
		t.Fatalf("npc should be indonesia: %s", raw)
	}
}

func TestPhase18StoryQuestsAndChapterUnlock(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 1.6, 4.2
	if p.quest("stq001").State != QuestAvailable {
		t.Fatal("stq001 available")
	}
	w.claimTalk(p, "stq001")
	if p.quest("stq001").State != QuestClaimed {
		t.Fatal("stq001 claimed")
	}
	p.X, p.Z = -1.6, 5.2
	w.acceptQuest(p, "stq002")
	p.credit("TALK", "raven", 1)
	p.credit("TALK", "lio", 1)
	w.claimQuest(p, "stq002")
	p.ensureLog().ForestUnlocked = true
	p.X, p.Z = 1.6, 4.2
	w.acceptQuest(p, "stq003")
	p.X, p.Z = 0, 36
	w.Simulate(0.05)
	p.credit("VISIT", "forest", 1)
	p.X, p.Z = 1.6, 4.2
	w.claimQuest(p, "stq003")
	if !p.ensureLog().Flags["story_chapter_1_complete"] {
		t.Fatal("ch1 complete")
	}
	if storyChapterState(p.ensureLog(), "st-ch02") != StoryActive && !p.ensureLog().Flags["story_chapter_2_available"] {
		t.Fatal("ch2 unlocked")
	}
}

func TestPhase18ChoicePersist(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 1.6, 4.2
	w.ApplyExplore(p.ID, Envelope{Type: TypeStoryChoice, Data: []byte(`{"choiceId":"mbah_jagat","option":"A"}`)})
	if p.ensureLog().StoryChoices["mbah_jagat"] != "A" {
		t.Fatal("choice saved")
	}
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeStoryChoice, Data: []byte(`{"choiceId":"mbah_jagat","option":"B"}`)})
	if !rejectAction(evs, TypeStoryChoice) {
		t.Fatal("historical choice locked")
	}
	w.persist(p)
	w2 := NewWorldState()
	w2.QuestRepo = w.QuestRepo
	p2 := addTestPlayer(w2, p.ID, "Raka")
	if p2.ensureLog().StoryChoices["mbah_jagat"] != "A" {
		t.Fatal("reconnect choice")
	}
}

func TestPhase18SilumanCodex(t *testing.T) {
	if len(silumanRoster) != 33 {
		t.Fatalf("siluman %d", len(silumanRoster))
	}
	if silumanByGuard["ragha"].Name != "Jaladara" {
		t.Fatalf("jaladara %s", silumanByGuard["ragha"].Name)
	}
	if silumanByGuard["avaron"].Name != "Adityara" {
		t.Fatalf("adityara %s", silumanByGuard["avaron"].Name)
	}
	if guardianByID["ragha"].Name != "Kabutra" || guardianByID["avaron"].Name != "Avaron" {
		t.Fatal("combat names must stay")
	}
	w, p := testVillagePlayer()
	w.markGuardianDefeated(p, "ragha")
	if p.ensureLog().SilumanProg["ragha"] != SilumanDefeated {
		t.Fatal("siluman prog")
	}
	j := p.worldJournal(w)
	if j.Guardians[0].StoryName != "Jaladara" {
		t.Fatalf("codex name %s", j.Guardians[0].StoryName)
	}
	w.ApplyExplore(p.ID, Envelope{Type: TypeStoryChoice, Data: []byte(`{"choiceId":"siluman:ragha","option":"A"}`)})
	if p.ensureLog().SilumanProg["ragha"] != SilumanAlly {
		t.Fatalf("ally %s", p.ensureLog().SilumanProg["ragha"])
	}
	w.persist(p)
	w2 := NewWorldState()
	w2.QuestRepo = w.QuestRepo
	p2 := addTestPlayer(w2, p.ID, "Raka")
	if p2.ensureLog().SilumanProg["ragha"] != SilumanAlly {
		t.Fatal("codex persist")
	}
}

func TestPhase18CinematicSkip(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 1.6, 4.2
	w.acceptQuest(p, "stq001")
	evs := w.claimQuest(p, "stq001")
	if !payloadHas(evs, TypeCinematicStart, "cin-opening") && !payloadHas(evs, TypeCinematicStart, "Wektu") {
		t.Fatalf("cinematic start %s", string(evs[0]))
	}
	skip := w.ApplyExplore(p.ID, Envelope{Type: TypeSkipCinematic, Data: []byte(`{}`)})
	if !payloadHas(skip, TypeCinematicSkipped, "") {
		t.Fatal("skip")
	}
	if !p.ensureLog().Flags["opening_cinematic_seen"] || !p.ensureLog().Flags["story_started"] {
		t.Fatal("flags after skip")
	}
	if p.ensureLog().PendingCinematic != "" {
		t.Fatal("pending cleared")
	}
}

func TestPhase18ChapterRewardSecurity(t *testing.T) {
	w, p := testVillagePlayer()
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeClaimStoryChapter, Data: []byte(`{"chapterId":"st-ch01"}`)})
	if !rejectAction(evs, TypeClaimStoryChapter) {
		t.Fatal("incomplete reject")
	}
	p.ensureLog().Flags["story_chapter_1_complete"] = true
	p.ensureLog().Flags["st-ch01_complete"] = true
	evs = w.ApplyExplore(p.ID, Envelope{Type: TypeClaimStoryChapter, Data: []byte(`{"chapterId":"st-ch01"}`)})
	if rejectAction(evs, TypeClaimStoryChapter) {
		t.Fatal("first claim")
	}
	evs = w.ApplyExplore(p.ID, Envelope{Type: TypeClaimStoryChapter, Data: []byte(`{"chapterId":"st-ch01"}`)})
	if !rejectAction(evs, TypeClaimStoryChapter) {
		t.Fatal("duplicate")
	}
	for _, cheat := range []string{TypeSetStoryFlag, TypeUnlockStoryChapter, TypeDefeatSiluman, TypeClaimStoryReward, TypeQuestComplete} {
		var evs [][]byte
		if cheat == TypeQuestComplete {
			evs = w.ApplyWorld(p.ID, Envelope{Type: cheat, Data: []byte(`{"questId":"stq001"}`)})
		} else {
			evs = w.ApplyExplore(p.ID, Envelope{Type: cheat, Data: []byte(`{}`)})
		}
		if !rejectAction(evs, cheat) {
			t.Fatalf("cheat %s", cheat)
		}
	}
}

func TestPhase18DarknessRisingAndEndings(t *testing.T) {
	if eventByID["darkness-rising"].DurationSec != 1200 {
		t.Fatalf("duration %v", eventByID["darkness-rising"].DurationSec)
	}
	w, p := testVillagePlayer()
	p.X, p.Z = 0, 6
	w.startWorldEvent("darkness-rising")
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeJoinWorldEvent, Data: []byte(`{}`)})
	if !payloadHas(evs, TypeWorldEvent, "DARKNESS") && !payloadHas(evs, TypeWorldEvent, "darkness-rising") {
		t.Fatal("join darkness")
	}
	help := pickStoryEnding(&PlayerLog{HelpScore: 9, PowerScore: 1, RedeemScore: 1})
	if help.ID != "ending-a" {
		t.Fatalf("ending a %s", help.ID)
	}
	pow := pickStoryEnding(&PlayerLog{HelpScore: 1, PowerScore: 9, RedeemScore: 1})
	if pow.ID != "ending-b" {
		t.Fatalf("ending b %s", pow.ID)
	}
	red := pickStoryEnding(&PlayerLog{HelpScore: 1, PowerScore: 1, RedeemScore: 9})
	if red.ID != "ending-c" {
		t.Fatalf("ending c %s", red.ID)
	}
}

func TestPhase18FinalBattlePostGame(t *testing.T) {
	w, p := testVillagePlayer()
	p.X, p.Z = 0, 6
	w.markGuardianDefeated(p, "avaron")
	evs := w.ApplyExplore(p.ID, Envelope{Type: TypeStoryChoice, Data: []byte(`{"choiceId":"adityara","option":"B"}`)})
	if !payloadHas(evs, TypeCinematicStart, "ending") && !p.ensureLog().Flags["storyCompleted"] {
		t.Fatalf("ending cinematic %v flags %v", payloadHas(evs, TypeCinematicStart, ""), p.ensureLog().Flags["storyCompleted"])
	}
	if p.ZoneID != "village" && p.Z != 6 {
		t.Fatalf("postgame z=%v zone=%s", p.Z, p.ZoneID)
	}
	if p.ensureLog().StoryState != StoryCompleted {
		t.Fatal("story completed")
	}
	ng := w.ApplyExplore(p.ID, Envelope{Type: TypeStartNGPlus, Data: []byte(`{}`)})
	if rejectAction(ng, TypeStartNGPlus) {
		t.Fatal("ng+")
	}
	if p.ensureLog().NGPlus < 1 || p.Level < 1 {
		t.Fatal("character kept")
	}
}

func TestPhase18EducationalAndTransformAlias(t *testing.T) {
	if questionByID["q-apel-3-2"].Correct != 1 || questionByID["q-add-5-4"].Correct != 1 || questionByID["q-add-6-2"].Correct != 1 {
		t.Fatal("edu questions")
	}
	if questionByID["q-add-4-3"].ID == "" {
		t.Fatal("keep dungeon q-add-4-3")
	}
	w, p := testVillagePlayer()
	p.X, p.Z = 1.6, 4.2
	w.claimTalk(p, "stq001")
	p.X, p.Z = 0, 4
	w.acceptQuest(p, "stq-edu-apel")
	evs := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"petani","kind":"quiz-apel"}`)})
	if !payloadHas(evs, TypeInteractResult, "q-apel-3-2") && !payloadHas(evs, TypeInteractResult, "apel") {
		t.Fatalf("apel quiz %s", string(evs[0]))
	}
	w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: []byte(`{"questionId":"q-apel-3-2","choice":1}`)})
	if p.ensureLog().KnowledgePoints < 1 {
		t.Fatal("knowledge")
	}
	view := w.storyStateView(p)
	if view.Forms[0]["story"] != "Awakened" || transformByID["aura-1"].Name != "AURA ASCENSION I" {
		t.Fatal("transform alias")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
	if npcByID["mira"].Type != "TEACHER" {
		t.Fatal("mira kept")
	}
}
