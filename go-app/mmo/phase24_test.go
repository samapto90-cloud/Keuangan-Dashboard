package mmo

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPhase24LandminesKept(t *testing.T) {
	if questionByID["q-add-4-3"].ID == "" || questionByID["q-add-4-3"].Correct != 1 {
		t.Fatal("keep q-add-4-3")
	}
	if questionByID["q-apel-3-2"].Correct != 1 {
		t.Fatal("keep q-apel-3-2")
	}
	if questionByID["q-add-5-3"].ID == "" || questionByID["q-add-5-3"].Correct != 1 {
		t.Fatal("keep q-add-5-3")
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatal("keep Dawn City")
	}
	if mountByID["wind-runner"].Name != "Wind Runner" || mountByID["wind-runner"].Speed != 1.45 {
		t.Fatal("keep wind-runner")
	}
	if shopCatalog.ID != "dawn-merchant" {
		t.Fatal("shops.json shape")
	}
	if len(storyChapterCatalog) != 10 {
		t.Fatalf("chapters %d", len(storyChapterCatalog))
	}
	if storyChapterByID["st-ch01"].Title != "Awal Perjalanan" {
		t.Fatal("st-ch01 title")
	}
	if silumanByGuard["ragha"].Name != "Jaladara" {
		t.Fatal("jaladara")
	}
	if npcByID["mira"].Name == "" || lineOr("mira_welcome", "") == "" {
		t.Fatal("mira kept")
	}
	if dialogueCatalog["pak_jaga"].Text != "Ati-ati, Le! Ana siluman teka saka alas!" {
		t.Fatal("pak jaga")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
	if questByID["mq004"].Title != "Jalan Menuju Hutan" {
		t.Fatal("mq004")
	}
	if eventByID["open-house"].Name != "OPEN HOUSE" || eventByID["mount-festival"].ID == "" {
		t.Fatal("events")
	}
}

func TestPhase24NPCService(t *testing.T) {
	if npcByID["mbah_guru"].Type != "MASTER" || npcByID["mbah_guru"].Trait != "WISE" {
		t.Fatal("mbah guru")
	}
	if npcByID["laras"].Name != "Laras" || npcByID["jaka"].Name != "Jaka" || npcByID["bagas"].Name != "Bagas" {
		t.Fatal("names")
	}
	if npcByID["sari"].Name != "Sari" || npcByID["wira"].Name != "Wira" {
		t.Fatal("sari wira")
	}
	if npcByID["mbah_karya"].Trait != "WISE" || npcByID["mbah_karya"].Name != "Mbah Karya" {
		t.Fatal("karya overlay")
	}
	if questByID["nq-karya-1"].Title != "Kayu kanggo Pandhe." || questByID["nq-karya-2"].Title != "Nggawe Gegaman." || questByID["nq-karya-3"].Title != "Njaga Desa." {
		t.Fatal("karya chain")
	}
	if questByID["nq-guru-1"].Title != "Sinau Ngitung." || questByID["nq-guru-2"].Title != "Sinau Maca." || questByID["nq-guru-3"].Title != "Sinau Basa Jawa." {
		t.Fatal("guru chain")
	}
	if loreByID["lore-siluman-01"].EnemyName != "Klawu" || loreByID["lore-siluman-02"].EnemyName != "Grembyang" {
		t.Fatal("klawu grembyang")
	}
	if loreByID["lore-siluman-03"].EnemyName != "Watu Geni" || loreByID["lore-siluman-05"].EnemyName != "Gandring" {
		t.Fatal("watu gandring")
	}
	n := 0
	for _, l := range loreCatalog {
		if strings.HasPrefix(l.ID, "lore-siluman-") {
			n++
		}
	}
	if n != 33 {
		t.Fatalf("siluman lore %d", n)
	}
}

func p24Near(p *Player, id string) {
	n := npcByID[id]
	x, z := npcDest(n, WorldTimeSystem{ClockMin: 9 * 60})
	p.X, p.Z = x, z
}

func TestPhase24DialogueChoiceAndLanguage(t *testing.T) {
	w, p := testVillagePlayer()
	p24Near(p, "mbah_karya")
	talk := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_karya","kind":"talk"}`)})
	if !payloadHas(talk, TypeInteractResult, "Kowe arep menyang ngendi") {
		t.Fatalf("where jv %s", talk)
	}
	a := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_karya","kind":"choice-where-a"}`)})
	if !payloadHas(a, TypeInteractResult, "Masjid") && !payloadHas(a, TypeInteractResult, "masjid") {
		t.Fatal("choice A")
	}
	b := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_karya","kind":"choice-where-b"}`)})
	if !payloadHas(b, TypeInteractResult, "kanca") && !payloadHas(b, TypeInteractResult, "kawan") {
		t.Fatal("choice B still allowed")
	}
	w.ApplyExplore(p.ID, Envelope{Type: TypeSetLanguage, Data: []byte(`{"language":"id"}`)})
	talkID := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_karya","kind":"talk"}`)})
	if !payloadHas(talkID, TypeInteractResult, "Kamu hendak ke mana") && !payloadHas(talkID, TypeInteractResult, "ke mana") {
		t.Fatalf("id lang %s", talkID)
	}
}

func TestPhase24MemoryPersist(t *testing.T) {
	w, p := testVillagePlayer()
	p24Near(p, "mbah_karya")
	w.acceptQuest(p, "nq-karya-1")
	help := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_karya","kind":"choice-help"}`)})
	if !payloadHas(help, TypeInteractResult, "mbiyen nulungi") && !payloadHas(help, TypeInteractResult, "Matur nuwun") {
		t.Fatalf("help %s", help)
	}
	if !p.ensureLog().NpcMemory["mbah_karya"] {
		t.Fatal("memory flag")
	}
	w.persist(p)
	w2 := NewWorldState()
	w2.QuestRepo = w.QuestRepo
	p2 := addTestPlayer(w2, p.ID, "Raka")
	p24Near(p2, "mbah_karya")
	if !p2.ensureLog().NpcMemory["mbah_karya"] {
		t.Fatal("reconnect memory")
	}
	again := w2.ApplyWorld(p2.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_karya","kind":"talk"}`)})
	if !payloadHas(again, TypeInteractResult, "mbiyen nulungi") {
		t.Fatalf("remember %s", again)
	}
}

func TestPhase24EducationRewardAndWrong(t *testing.T) {
	w, p := testVillagePlayer()
	p24Near(p, "mbah_guru")
	w.acceptQuest(p, "nq-guru-1")
	coin, exp := p.ensureLog().Coin, p.Exp
	start := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_guru","kind":"quiz-guru"}`)})
	if !payloadHas(start, TypeInteractResult, "q-apel-3-2") {
		t.Fatal("server question")
	}
	wrong := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: []byte(`{"questionId":"q-apel-3-2","choice":0}`)})
	if !payloadHas(wrong, TypeEducationFeedback, "Durung apa-apa") {
		t.Fatal("explain")
	}
	if p.ensureLog().Coin != coin || p.Exp != exp {
		t.Fatal("no penalty")
	}
	ok := w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: []byte(`{"questionId":"q-apel-3-2","choice":1}`)})
	if !payloadHas(ok, TypeEducationFeedback, `"correct":true`) && !payloadHas(ok, TypeEducationFeedback, `"correct": true`) {
		if !payloadHas(ok, TypeEducationFeedback, "Pinter") && !payloadHas(ok, TypeEducationFeedback, "3 ditambah 2") {
			t.Fatalf("correct %s", ok)
		}
	}
	if p.ensureLog().EduToken < 1 {
		t.Fatal("knowledge token")
	}
	if p.ensureLog().NpcRel["mbah_guru"] < 1 {
		t.Fatal("rel up")
	}
}

func TestPhase24CheatsRejected(t *testing.T) {
	w, p := testVillagePlayer()
	cheats := []string{TypeSetStoryFlag, TypeQuestComplete, TypeCompleteQuest, TypeSetRelationship, TypeAddRelationship, TypeSetNpcMemory, TypeUnlockLore, TypeSetWorldState, TypeClaimEducationReward}
	for _, c := range cheats {
		raw := []byte(`{"npcId":"mbah_karya","xp":1000,"storyFlag":true,"questId":"nq-karya-1"}`)
		var evs [][]byte
		if c == TypeQuestComplete {
			evs = w.ApplyWorld(p.ID, Envelope{Type: c, Data: raw})
		} else if c == TypeSetStoryFlag {
			evs = w.ApplyExplore(p.ID, Envelope{Type: c, Data: raw})
		} else {
			evs = w.ApplyExplore(p.ID, Envelope{Type: c, Data: raw})
		}
		if !rejectAction(evs, c) {
			t.Fatalf("%s", c)
		}
	}
	if p.ensureLog().NpcRel["mbah_karya"] != 0 {
		t.Fatal("rel cheat")
	}
}

func TestPhase24ScheduleAndWorldState(t *testing.T) {
	n := npcByID["laras"]
	morn := WorldTimeSystem{ClockMin: 8 * 60}
	eve := WorldTimeSystem{ClockMin: 19 * 60}
	mx, mz := npcDest(n, morn)
	ex, ez := npcDest(n, eve)
	if mx == ex && mz == ez {
		t.Fatal("schedule move")
	}
	w, p := testVillagePlayer()
	p24Near(p, "wira")
	danger := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"wira","kind":"talk"}`)})
	if !payloadHas(danger, TypeInteractResult, "mbebayani") && !payloadHas(danger, TypeInteractResult, "berbahaya") && !payloadHas(danger, TypeInteractResult, "adoh") {
		t.Fatalf("danger %s", danger)
	}
	p.ensureLog().Flags["chapter1_complete"] = true
	safe := w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"wira","kind":"talk"}`)})
	if !payloadHas(safe, TypeInteractResult, "ayem") && !payloadHas(safe, TypeInteractResult, "tenang") {
		t.Fatalf("safe %s", safe)
	}
}

func TestPhase24CinematicSkipAndJournal(t *testing.T) {
	w, p := testVillagePlayer()
	if cinematicByID["cin-ch-awal"].ID == "" || cinematicByID["cin-opening"].ID == "" {
		t.Fatal("cinematics")
	}
	evs := w.startCinematic(p, "cin-ch-awal", false)
	if !payloadHas(evs, TypeCinematicStart, "cin-ch-awal") {
		t.Fatal("start")
	}
	skip := w.skipCinematic(p)
	if !payloadHas(skip, TypeCinematicSkipped, "ok") {
		t.Fatal("skip")
	}
	j := p.worldJournal(w)
	if j.WorldMood == "" || len(j.OverlayChapters) != 8 || len(j.EnemyLore) != 33 {
		t.Fatalf("journal mood=%s ch=%d lore=%d", j.WorldMood, len(j.OverlayChapters), len(j.EnemyLore))
	}
	if j.OverlayChapters[7].Locked == false || j.OverlayChapters[7].Title != "Jalan Menuju Cahaya" {
		t.Fatal("final locked")
	}
	adv := w.ApplyExplore(p.ID, Envelope{Type: TypeGetAdventure, Data: []byte(`{}`)})
	if !payloadHas(adv, TypeWorldJournal, "npcBook") && !payloadHas(adv, TypeWorldJournal, "overlayChapters") {
		t.Fatal("adventure")
	}
}

func TestPhase24ChapterUnlock(t *testing.T) {
	w, p := testVillagePlayer()
	log := p.ensureLog()
	log.Flags["story_chapter_1_complete"] = true
	log.StoryState = StoryActive
	j := p.worldJournal(w)
	if j.OverlayChapters[0].State != StoryCompleted && j.OverlayChapters[1].State == StoryLocked {
		t.Fatal("next chapter")
	}
	if storyChapterState(log, "st-ch02") == StoryLocked {
		t.Fatal("st-ch02 unlock")
	}
}

func TestPhase24ChoicePersist(t *testing.T) {
	w, p := testVillagePlayer()
	p24Near(p, "mbah_karya")
	w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_karya","kind":"choice-where-a"}`)})
	if p.ensureLog().StoryChoices["dlg-where"] != "A" {
		t.Fatal("choice saved")
	}
	w.persist(p)
	w2 := NewWorldState()
	w2.QuestRepo = w.QuestRepo
	p2 := addTestPlayer(w2, p.ID, "Raka")
	if p2.ensureLog().StoryChoices["dlg-where"] != "A" {
		t.Fatal("choice reconnect")
	}
}

func TestPhase24CorrectAnswerNotClient(t *testing.T) {
	w, p := testVillagePlayer()
	p24Near(p, "mbah_guru")
	w.ApplyWorld(p.ID, Envelope{Type: TypeInteract, Data: []byte(`{"targetId":"mbah_guru","kind":"quiz-guru"}`)})
	raw, _ := json.Marshal(map[string]any{"questionId": "q-apel-3-2", "choice": 1, "correct": true})
	_ = w.ApplyWorld(p.ID, Envelope{Type: TypeEducationAnswer, Data: raw})
	if questionByID["q-apel-3-2"].Correct != 1 {
		t.Fatal("server answer")
	}
}
