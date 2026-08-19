package mmo

import (
	"strconv"
	"strings"
	"time"

	_ "embed"
)

//go:embed data/storyChapters.json
var storyChaptersJSON []byte

//go:embed data/storyDialogues.json
var storyDialoguesJSON []byte

//go:embed data/storyCinematics.json
var storyCinematicsJSON []byte

//go:embed data/storyEndings.json
var storyEndingsJSON []byte

const (
	StoryNotStarted = "NOT_STARTED"
	StoryActive     = "ACTIVE"
	StoryCompleted  = "COMPLETED"
	StoryFailed     = "FAILED"
	StoryLocked     = "LOCKED"

	SilumanDefeated   = "DEFEATED"
	SilumanUnderstood = "UNDERSTOOD"
	SilumanAlly       = "ALLY"
)

type StoryChapterDef struct {
	ID          string `json:"id"`
	Index       int    `json:"index"`
	Title       string `json:"title"`
	Music       string `json:"music"`
	Region      string `json:"region"`
	CinematicID string `json:"cinematicId"`
	RewardTitle string `json:"rewardTitle"`
	RewardExp   int    `json:"rewardExp"`
}

type LocLine struct {
	JV string `json:"jv"`
	ID string `json:"id"`
	EN string `json:"en"`
}

type CinematicDef struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	DurationSec     int      `json:"durationSec"`
	Skippable       bool     `json:"skippable"`
	Camera          string   `json:"camera"`
	Music           string   `json:"music"`
	VFX             string   `json:"vfx"`
	Lines           []string `json:"lines"`
	FlagsOnComplete []string `json:"flagsOnComplete"`
}

type StoryEndingDef struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	CinematicID string `json:"cinematicId"`
	TitleReward string `json:"titleReward"`
	Cosmetic    string `json:"cosmetic"`
	Need        string `json:"need"`
}

type SilumanStory struct {
	Index       int
	GuardianID  string
	Name        string
	Region      string
	Level       int
	Personality string
	Weakness    string
	Story       string
	MiniBoss    bool
	Final       bool
	AllyOK      bool
}

type StoryChapterView struct {
	ID     string `json:"id"`
	Index  int    `json:"index"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Music  string `json:"music,omitempty"`
	Reward bool   `json:"rewardClaimed,omitempty"`
}

type CinematicView struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	DurationSec int      `json:"durationSec"`
	Skippable   bool     `json:"skippable"`
	Camera      string   `json:"camera"`
	Music       string   `json:"music"`
	VFX         string   `json:"vfx,omitempty"`
	Lines       []string `json:"lines"`
	Subtitle    []string `json:"subtitle,omitempty"`
	Replay      bool     `json:"replay,omitempty"`
	ToID        string   `json:"toId,omitempty"`
}

type StoryStateView struct {
	PlayerID        string              `json:"playerId"`
	Language        string              `json:"language"`
	Chapter         string              `json:"chapter"`
	State           string              `json:"state"`
	Objective       string              `json:"objective"`
	Chapters        []StoryChapterView  `json:"chapters"`
	Choices         map[string]string   `json:"choices"`
	DefeatedSiluman []string            `json:"defeatedSiluman"`
	Siluman         []GuardianView      `json:"siluman"`
	Allies          []string            `json:"allies"`
	Archive         []CinematicView     `json:"archive"`
	EndingID        string              `json:"endingId,omitempty"`
	NGPlus          int                 `json:"ngPlus"`
	StoryCompleted  bool                `json:"storyCompleted"`
	ExplorerMode    bool                `json:"explorerMode"`
	Forms           []map[string]string `json:"forms"`
	FinalBoss       map[string]any      `json:"finalBoss,omitempty"`
	ToID            string              `json:"toId,omitempty"`
}

var (
	storyChapterCatalog []StoryChapterDef
	storyChapterByID    = map[string]StoryChapterDef{}
	storyDialogue       = map[string]LocLine{}
	cinematicCatalog    []CinematicDef
	cinematicByID       = map[string]CinematicDef{}
	storyEndingCatalog  []StoryEndingDef
	silumanRoster       []SilumanStory
	silumanByGuard      = map[string]SilumanStory{}
)

var silumanNames = []string{
	"Jaladara", "Wulansa", "Gandaru", "Kertala", "Bramasta",
	"Ranjala", "Sudara", "Larasati", "Tirtaka", "Kalasura",
	"Wiraga", "Nagara", "Jayanta", "Raksana", "Madrasa",
	"Sangkala", "Wiratma", "Arunika", "Kamala", "Darmaga",
	"Ranggala", "Baskara", "Candraka", "Surajaya", "Tunggara",
	"Mandala", "Kartika", "Jayengra", "Rakata", "Wiratara",
	"Kaladipa", "Mahesara", "Adityara",
}

var silumanPersonalities = []string{
	"sombong", "sedih", "pelindung", "pemarah", "bijaksana",
	"licik", "penasaran", "takut", "adil", "penipu",
	"pendiam", "keras", "dingin", "berubah-ubah", "angkuh",
	"curiga", "teliti", "lembut", "ceria", "ringan",
	"kaku", "penghitung", "impulsif", "waspada", "lapar cahaya",
	"tajam", "liar", "sunyi", "menggema", "gelisah",
	"penipu lembut", "berwibawa", "tenang penguji",
}

func init() {
	mustJSON("storyChapters.json", storyChaptersJSON, &storyChapterCatalog)
	for _, c := range storyChapterCatalog {
		storyChapterByID[c.ID] = c
	}
	if err := unmarshal(storyDialoguesJSON, &storyDialogue); err != nil {
		mustJSON("storyDialogues.json", storyDialoguesJSON, &storyDialogue)
	}
	mustJSON("storyCinematics.json", storyCinematicsJSON, &cinematicCatalog)
	for i := range cinematicCatalog {
		c := cinematicCatalog[i]
		if c.DurationSec < 5 {
			c.DurationSec = 5
		}
		if c.DurationSec > 30 {
			c.DurationSec = 30
		}
		cinematicCatalog[i] = c
		cinematicByID[c.ID] = c
	}
	mustJSON("storyEndings.json", storyEndingsJSON, &storyEndingCatalog)
	buildSilumanRoster()
	seedStorySideQuests()
}

func buildSilumanRoster() {
	silumanRoster = silumanRoster[:0]
	for i, g := range guardianCatalog {
		name := silumanNames[i%len(silumanNames)]
		if i < len(silumanNames) {
			name = silumanNames[i]
		}
		region, level := silumanBand(i)
		st := SilumanStory{
			Index: i + 1, GuardianID: g.ID, Name: name, Region: region, Level: level,
			Personality: silumanPersonalities[i%len(silumanPersonalities)],
			Weakness:    g.Weakness,
			Story:       name + " njaga " + region + ". Ora kabeh jahat; sawetara mung njaga dalan.",
			MiniBoss:    i == 4 || i == 9 || i == 14 || i == 19 || i == 24 || i == 29,
			Final:       i == 32,
			AllyOK:      i == 0 || i == 4 || i == 7 || i == 17 || i == 26 || i == 31 || i == 32,
		}
		if st.Weakness == "" {
			st.Weakness = "kebaikan lan jawaban jujur"
		}
		guardianCatalog[i].StoryName = name
		silumanRoster = append(silumanRoster, st)
		silumanByGuard[g.ID] = st
		guardianByID[g.ID] = guardianCatalog[i]
	}
}

func silumanBand(i int) (string, int) {
	switch {
	case i < 10:
		if i < 5 {
			return "Mistwood", 10 + i*2
		}
		return "Stone Valley", 18 + (i - 5)
	case i < 20:
		if i < 15 {
			return "River of Light", 25 + (i-10)*2
		}
		return "Crimson Plains", 33 + (i - 15)
	case i < 30:
		if i < 25 {
			return "Celestial Mountains", 40 + (i-20)*2
		}
		return "Ancient Ruins", 48 + (i - 25)
	default:
		if i == 32 {
			return "Final Sanctum", 60
		}
		return "Dark Gate", 55 + (i-30)*3
	}
}

func seedStorySideQuests() {
	regions := []struct{ id, loc, npc string }{
		{"village", "Dawn Village", "petani"},
		{"forest", "Mistwood Forest", "mbah_jagat"},
		{"valley", "Stone Valley", "mbah_jagat"},
		{"plains", "River of Light", "mbah_jagat"},
		{"canyon", "Crimson Plains", "mbah_jagat"},
		{"temple", "Ancient Ruins", "mbah_jagat"},
		{"masjid", "Gerbang Cahaya", "mbah_jagat"},
	}
	titles := []string{"Membantu Petani", "Mencari Anak Hilang", "Memperbaiki Jembatan", "Membantu Pedagang", "Menjaga Jalan"}
	for _, r := range regions {
		for n := 1; n <= 5; n++ {
			id := "ssq-" + r.id + "-" + strconv.Itoa(n)
			if questByID[id].ID != "" {
				continue
			}
			npc := r.npc
			if npcByID[npc].ID == "" {
				npc = "mbah_jagat"
			}
			q := QuestDef{
				ID: id, Title: titles[(n-1)%len(titles)] + " · " + r.loc, Kind: "side",
				NPC: npc, Location: r.loc, Description: "Quest sampingan. Menolong, bukan menyerang.",
				Prereq:     []string{"stq001"},
				Objectives: []ObjectiveDef{{Type: "TALK", Target: npc, Count: 1, Text: "Bicara dan bantu"}},
				Rewards:    RewardDef{Exp: 20, Coin: 8, Knowledge: 1},
				ClaimAt:    npc,
			}
			questCatalog = append(questCatalog, q)
			questByID[id] = q
		}
	}
}

func playerLang(p *Player) string {
	lang := strings.ToLower(strings.TrimSpace(p.ensureLog().Language))
	if lang == "id" || lang == "indonesia" {
		return "id"
	}
	if lang == "en" {
		return "en"
	}
	return "jv"
}

func locText(key, lang string) string {
	d, ok := storyDialogue[key]
	if !ok {
		return key
	}
	switch lang {
	case "id":
		if d.ID != "" {
			return d.ID
		}
	case "en":
		if d.EN != "" {
			return d.EN
		}
	}
	if d.JV != "" {
		return d.JV
	}
	if d.ID != "" {
		return d.ID
	}
	return key
}

func locPair(key, lang string) (text, sub string) {
	text = locText(key, lang)
	if lang == "jv" {
		sub = locText(key, "id")
		if sub == text {
			sub = ""
		}
		return text, sub
	}
	return text, ""
}

func (w *WorldState) ApplyStory(id string, env Envelope) [][]byte {
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	switch env.Type {
	case TypeSetLanguage:
		var in struct {
			Language string `json:"language"`
		}
		_ = unmarshal(env.Data, &in)
		return w.setStoryLanguage(p, in.Language)
	case TypeGetStory:
		return w.getStory(p, false)
	case TypeStoryChoice:
		var in struct {
			ChoiceID string `json:"choiceId"`
			Option   string `json:"option"`
			TargetID string `json:"targetId"`
		}
		_ = unmarshal(env.Data, &in)
		return w.submitStoryChoice(p, in.ChoiceID, in.Option, in.TargetID)
	case TypeClaimStoryChapter:
		var in struct {
			ChapterID string `json:"chapterId"`
		}
		_ = unmarshal(env.Data, &in)
		return w.claimStoryChapter(p, in.ChapterID)
	case TypeReplayCinematic:
		var in struct {
			CinematicID string `json:"cinematicId"`
		}
		_ = unmarshal(env.Data, &in)
		return w.replayCinematic(p, in.CinematicID)
	case TypeReplayChapter:
		var in struct {
			ChapterID string `json:"chapterId"`
		}
		_ = unmarshal(env.Data, &in)
		return w.replayChapter(p, in.ChapterID)
	case TypeStartNGPlus:
		return w.startNewGamePlus(p)
	case TypeCinematicDone:
		return w.finishCinematic(p)
	case TypeSetStoryFlag, TypeUnlockStoryChapter, TypeDefeatSiluman, TypeClaimStoryReward:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) setStoryLanguage(p *Player, lang string) [][]byte {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch lang {
	case "id", "indonesia", "indonesian":
		lang = "id"
	case "en", "english":
		lang = "en"
	default:
		lang = "jv"
	}
	log := p.ensureLog()
	log.Language = lang
	p.markDirty()
	w.persist(p)
	return w.getStory(p, true)
}

func (w *WorldState) getStory(p *Player, replay bool) [][]byte {
	log := p.ensureLog()
	if log.StoryState == "" {
		log.StoryState = StoryNotStarted
	}
	if log.StoryChapter == "" {
		log.StoryChapter = "st-ch01"
	}
	out := [][]byte{marshal(TypeStoryState, w.storyStateView(p))}
	if !replay && !log.CinematicsSeen["cin-opening"] && (log.StoryState == StoryNotStarted || log.StoryState == "") {
		out = append(out, w.startCinematic(p, "cin-opening", false)...)
	}
	return out
}

func (w *WorldState) storyStateView(p *Player) StoryStateView {
	log := p.ensureLog()
	lang := playerLang(p)
	chs := make([]StoryChapterView, 0, len(storyChapterCatalog))
	for _, c := range storyChapterCatalog {
		chs = append(chs, StoryChapterView{
			ID: c.ID, Index: c.Index, Title: c.Title, State: storyChapterState(log, c.ID), Music: c.Music,
			Reward: log.ChapterRewards[c.ID],
		})
	}
	sil := make([]GuardianView, 0, len(silumanRoster))
	defeated := []string{}
	for i, g := range guardianCatalog {
		st := silumanByGuard[g.ID]
		status := ""
		if log.SilumanProg != nil {
			status = log.SilumanProg[g.ID]
		}
		if status == "" && (log.Guardians[g.ID] || log.Flags["guardian_"+g.ID+"_defeated"]) {
			status = SilumanDefeated
		}
		if status == SilumanDefeated || status == SilumanUnderstood || status == SilumanAlly {
			defeated = append(defeated, g.ID)
		}
		sil = append(sil, GuardianView{
			ID: g.ID, Name: g.Name, StoryName: st.Name, Title: g.Title, Status: status, ChapterID: g.ChapterID,
			Region: st.Region, Index: i + 1, Personality: st.Personality, Weakness: st.Weakness, Story: st.Story,
			CodexStatus: status, Ally: status == SilumanAlly, MiniBoss: st.MiniBoss,
		})
	}
	archive := make([]CinematicView, 0, len(cinematicCatalog))
	for _, c := range cinematicCatalog {
		if log.CinematicsSeen != nil && log.CinematicsSeen[c.ID] {
			archive = append(archive, w.cinematicView(p, c, true))
		}
	}
	choices := map[string]string{}
	for k, v := range log.StoryChoices {
		choices[k] = v
	}
	out := StoryStateView{
		PlayerID: p.ID, Language: lang, Chapter: log.StoryChapter, State: log.StoryState,
		Objective: storyObjective(log, lang), Chapters: chs, Choices: choices, DefeatedSiluman: defeated,
		Siluman: sil, Allies: append([]string{}, log.StoryAllies...), Archive: archive, EndingID: log.EndingID,
		NGPlus: log.NGPlus, StoryCompleted: log.Flags["storyCompleted"], ExplorerMode: log.Flags["explorer_mode"] || log.Flags["storyCompleted"],
		Forms: []map[string]string{
			{"id": "aura-1", "name": "AURA ASCENSION I", "story": "Awakened"},
			{"id": "aura-2", "name": "AURA ASCENSION II", "story": "Ascended"},
			{"id": "aura-3", "name": "AURA ASCENSION III", "story": "Radiant"},
			{"id": "celestial-4", "name": "CELESTIAL AURA IV", "story": "Celestial"},
		},
		FinalBoss: map[string]any{
			"id": "raksa-kala", "name": "Raksa Kala",
			"phases":     []string{"Shadow Form", "Chaos Form", "Darkness Core"},
			"questionId": "q-add-5-4",
		},
		ToID: p.ID,
	}
	phase29EnrichStoryState(p, &out)
	return out
}

func storyChapterState(log *PlayerLog, id string) string {
	def, ok := storyChapterByID[id]
	if !ok {
		return StoryLocked
	}
	if log.Flags["story_chapter_"+strconv.Itoa(def.Index)+"_complete"] || log.Flags[id+"_complete"] || log.ChapterRewards[id] {
		return StoryCompleted
	}
	if def.Index == 1 {
		if log.StoryState == StoryActive || log.Flags["story_started"] || log.Flags["opening_cinematic_seen"] {
			return StoryActive
		}
		return StoryNotStarted
	}
	if log.Flags["story_chapter_"+strconv.Itoa(def.Index)+"_available"] || log.Flags[id+"_available"] ||
		log.Flags["story_chapter_"+strconv.Itoa(def.Index-1)+"_complete"] {
		return StoryActive
	}
	return StoryLocked
}

func storyObjective(log *PlayerLog, lang string) string {
	if log.Flags["storyCompleted"] {
		return "Explorer Mode. Dunia tetap terbuka."
	}
	if log.Flags["story_chapter_1_complete"] {
		return locText("story.objective.ch2", lang)
	}
	return locText("story.objective.ch1", lang)
}

func (w *WorldState) cinematicView(p *Player, c CinematicDef, replay bool) CinematicView {
	lang := playerLang(p)
	lines := make([]string, 0, len(c.Lines))
	subs := make([]string, 0, len(c.Lines))
	for _, key := range c.Lines {
		t, s := locPair(key, lang)
		lines = append(lines, t)
		if s != "" {
			subs = append(subs, s)
		}
	}
	dur := c.DurationSec
	if dur < 5 {
		dur = 5
	}
	if dur > 30 {
		dur = 30
	}
	return CinematicView{
		ID: c.ID, Title: c.Title, DurationSec: dur, Skippable: true, Camera: c.Camera, Music: c.Music, VFX: c.VFX,
		Lines: lines, Subtitle: subs, Replay: replay, ToID: p.ID,
	}
}

func (w *WorldState) startCinematic(p *Player, id string, replay bool) [][]byte {
	c, ok := cinematicByID[id]
	if !ok {
		return nil
	}
	log := p.ensureLog()
	if !replay {
		log.PendingCinematic = id
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeCinematicStart, w.cinematicView(p, c, replay))}
}

func (w *WorldState) completeCinematic(p *Player, id string, skipped bool) {
	if id == "" {
		return
	}
	c, ok := cinematicByID[id]
	if !ok {
		return
	}
	log := p.ensureLog()
	if log.CinematicsSeen == nil {
		log.CinematicsSeen = map[string]bool{}
	}
	log.CinematicsSeen[id] = true
	if skipped {
		log.Flags["cinematic_skipped"] = true
	}
	for _, f := range c.FlagsOnComplete {
		log.Flags[f] = true
	}
	if log.StoryState == "" || log.StoryState == StoryNotStarted {
		log.StoryState = StoryActive
		log.StoryChapter = "st-ch01"
	}
	if id == "cin-opening" {
		p.grantTitle("traveler")
		p.grantTitle("first-journey")
		log.Flags["ach_first_journey"] = true
	}
	if strings.HasPrefix(id, "cin-ending") {
		w.applyEndingFlags(p, id)
	}
	log.PendingCinematic = ""
	p.markDirty()
	w.persist(p)
	w.refreshAchievements(p)
}

func (w *WorldState) applyEndingFlags(p *Player, cinematicID string) {
	log := p.ensureLog()
	for _, e := range storyEndingCatalog {
		if e.CinematicID == cinematicID {
			log.EndingID = e.ID
			p.grantTitle(e.TitleReward)
			if e.Cosmetic != "" {
				p.grantCosmetic(e.Cosmetic)
			}
		}
	}
	log.StoryState = StoryCompleted
	log.Flags["ach_guardian_of_dawn"] = true
	p.grantTitle("guardian-of-dawn")
}

func (w *WorldState) replayCinematic(p *Player, id string) [][]byte {
	log := p.ensureLog()
	if id == "" || log.CinematicsSeen == nil || !log.CinematicsSeen[id] {
		return rejectFor(p.ID, TypeReplayCinematic, "locked")
	}
	before, ch := log.StoryState, log.StoryChapter
	out := w.startCinematic(p, id, true)
	log.StoryState = before
	log.StoryChapter = ch
	log.PendingCinematic = ""
	p.markDirty()
	w.persist(p)
	return out
}

func (w *WorldState) replayChapter(p *Player, chapterID string) [][]byte {
	def, ok := storyChapterByID[chapterID]
	if !ok {
		return rejectFor(p.ID, TypeReplayChapter, "chapter")
	}
	st, ch := p.ensureLog().StoryState, p.ensureLog().StoryChapter
	out := w.replayCinematic(p, def.CinematicID)
	p.ensureLog().StoryState = st
	p.ensureLog().StoryChapter = ch
	w.persist(p)
	return out
}

func (w *WorldState) submitStoryChoice(p *Player, choiceID, option, target string) [][]byte {
	if choiceID == "" {
		choiceID = target
	}
	if choiceID == "" {
		return rejectFor(p.ID, TypeStoryChoice, "payload")
	}
	option = normalizeStoryOption(option)
	log := p.ensureLog()
	if log.StoryChoices == nil {
		log.StoryChoices = map[string]string{}
	}
	if prev := log.StoryChoices[choiceID]; prev != "" {
		return rejectFor(p.ID, TypeStoryChoice, "historical")
	}
	log.StoryChoices[choiceID] = option
	if log.DialogueChoices == nil {
		log.DialogueChoices = map[string]string{}
	}
	log.DialogueChoices[choiceID] = option
	w.applyStoryChoiceEffect(p, choiceID, option)
	p.markDirty()
	w.persist(p)
	return w.afterStoryChoice(p, choiceID, option, target)
}

func normalizeStoryOption(opt string) string {
	opt = strings.ToLower(strings.TrimSpace(opt))
	switch opt {
	case "a", "choice-a", "ready", "spare", "ngapura", "fight", "ngadepi":
		return "A"
	case "b", "choice-b", "afraid", "leave", "talk", "nyoba":
		return "B"
	case "c", "choice-c", "learn", "ask", "friends", "critakna":
		return "C"
	}
	if opt == "" {
		return "A"
	}
	return strings.ToUpper(opt[:1])
}

func (w *WorldState) applyStoryChoiceEffect(p *Player, choiceID, option string) {
	log := p.ensureLog()
	gid := choiceID
	gid = strings.TrimPrefix(gid, "siluman:")
	gid = strings.TrimPrefix(gid, "guardian:")
	if _, ok := silumanByGuard[gid]; ok {
		w.applySilumanMoral(p, gid, option)
		return
	}
	if choiceID == "adityara" {
		log.Flags["adityara_choice_"+option] = true
		if option == "A" {
			log.PowerScore += 3
		}
		if option == "B" {
			log.HelpScore += 2
		}
		if option == "C" {
			log.RedeemScore += 3
		}
		return
	}
	switch option {
	case "A":
		log.HelpScore++
		p.bumpRegionRep("village", 4)
	case "B":
		log.PowerScore++
	case "C":
		p.grantKnowledge(1)
		log.HelpScore++
	}
}

func (w *WorldState) applySilumanMoral(p *Player, gid, option string) {
	log := p.ensureLog()
	if log.SilumanProg == nil {
		log.SilumanProg = map[string]string{}
	}
	st := silumanByGuard[gid]
	status := SilumanDefeated
	switch option {
	case "A":
		log.RedeemScore += 2
		if st.AllyOK {
			status = SilumanAlly
			w.addStoryAlly(p, gid)
			log.PendingCinematic = "cin-redemption"
		} else {
			status = SilumanUnderstood
		}
	case "B":
		log.PowerScore += 2
		status = SilumanDefeated
	case "C":
		log.HelpScore += 2
		p.grantKnowledge(1)
		status = SilumanUnderstood
	}
	log.SilumanProg[gid] = status
	if st.Final {
		log.Flags["adityara_choice_"+option] = true
	}
	p.markDirty()
}

func (w *WorldState) addStoryAlly(p *Player, id string) {
	log := p.ensureLog()
	for _, a := range log.StoryAllies {
		if a == id {
			return
		}
	}
	log.StoryAllies = append(log.StoryAllies, id)
	log.Flags["ally_"+id] = true
}

func (w *WorldState) afterStoryChoice(p *Player, choiceID, option, target string) [][]byte {
	log := p.ensureLog()
	lang := playerLang(p)
	text, sub := locPair("story.chapter1.mbahjagat.ready", lang)
	switch option {
	case "B":
		text, sub = locPair("story.chapter1.mbahjagat.afraid", lang)
	case "C":
		text, sub = locPair("story.chapter1.mbahjagat.learn", lang)
	}
	gid := strings.TrimPrefix(strings.TrimPrefix(choiceID, "siluman:"), "guardian:")
	if _, ok := silumanByGuard[gid]; ok {
		text, sub = locPair("story.siluman.jaladara.truth", lang)
		if option == "A" && log.SilumanProg[gid] == SilumanAlly {
			text, sub = locPair("story.siluman.redeem", lang)
		}
	}
	out := [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "story-choice", TargetID: target, Title: "PILIHAN", Speaker: "Mbah Jagat", Role: "Mentor",
		Text: text, Subtitle: sub, Options: []DialogOption{{ID: "close", Label: "Tutup"}},
	})}
	if log.PendingCinematic == "cin-redemption" {
		out = append(out, w.startCinematic(p, "cin-redemption", false)...)
	}
	if choiceID == "adityara" || gid == "avaron" {
		out = append(out, w.finishMainStory(p)...)
	}
	return out
}

func (w *WorldState) talkMbahJagat(p *Player, npc NPCDef) [][]byte {
	p.credit("TALK", npc.ID, 1)
	log := p.ensureLog()
	if log.StoryState == StoryNotStarted || log.StoryState == "" {
		log.StoryState = StoryActive
		log.StoryChapter = "st-ch01"
		log.Flags["story_started"] = true
	}
	lang := playerLang(p)
	text, sub := locPair("story.chapter1.mbahjagat.intro", lang)
	opts := []DialogOption{
		{ID: "choice-a", Label: locText("story.choice.ready", lang)},
		{ID: "choice-b", Label: locText("story.choice.afraid", lang)},
		{ID: "choice-c", Label: locText("story.choice.learn", lang)},
	}
	opts = append(opts, questOption(p, "stq001", locText("story.choice.ready", lang))...)
	opts = append(opts, questOption(p, "stq003", "Menyang wetan")...)
	opts = append(opts, questOption(p, "stq-allies", "Ngumpulake kanca")...)
	opts = append(opts, DialogOption{ID: "close", Label: "Tutup"})
	out := [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "npc", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
		Text: text, Subtitle: sub, Marker: npcMarker(p, npc), Options: opts, CinematicID: "cin-opening",
	})}
	if log.CinematicsSeen == nil || !log.CinematicsSeen["cin-opening"] {
		out = append(out, w.startCinematic(p, "cin-opening", false)...)
	}
	p.markDirty()
	w.persist(p)
	return out
}

func (w *WorldState) talkStoryVillager(p *Player, npc NPCDef) [][]byte {
	p.credit("TALK", npc.ID, 1)
	lang := playerLang(p)
	text, sub := locPair("story.ibu.farewell", lang)
	opts := questOption(p, "stq002", "Pamit")
	if npc.ID == "petani" {
		text, sub = locPair("story.petani.apel", lang)
		opts = append(questOption(p, "stq-edu-apel", "Aku nyoba"), DialogOption{ID: "quiz-apel", Label: "Jawab"})
	}
	opts = append(opts, DialogOption{ID: "close", Label: "Tutup"})
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "npc", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
		Text: text, Subtitle: sub, Marker: npcMarker(p, npc), Options: opts,
	})}
}

func (w *WorldState) storyNpcChoice(p *Player, npc NPCDef, choice string) [][]byte {
	choiceID := npc.ID
	option := normalizeStoryOption(choice)
	log := p.ensureLog()
	if log.StoryChoices == nil {
		log.StoryChoices = map[string]string{}
	}
	if log.StoryChoices[choiceID] == "" {
		log.StoryChoices[choiceID] = option
		w.applyStoryChoiceEffect(p, choiceID, option)
	}
	if log.DialogueChoices == nil {
		log.DialogueChoices = map[string]string{}
	}
	if log.DialogueChoices[npc.ID] == "" {
		log.DialogueChoices[npc.ID] = choice
	}
	p.credit("TALK", npc.ID, 1)
	lang := playerLang(p)
	key := "story.chapter1.mbahjagat.ready"
	switch option {
	case "B":
		key = "story.chapter1.mbahjagat.afraid"
	case "C":
		key = "story.chapter1.mbahjagat.learn"
	}
	text, sub := locPair(key, lang)
	p.markDirty()
	w.persist(p)
	opts := []DialogOption{
		{ID: "choice-a", Label: locText("story.choice.ready", lang)},
		{ID: "choice-b", Label: locText("story.choice.afraid", lang)},
		{ID: "choice-c", Label: locText("story.choice.learn", lang)},
		{ID: "close", Label: "Tutup"},
	}
	opts = append(questOption(p, "stq001", locText("story.choice.ready", lang)), opts...)
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "npc", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
		Text: text, Subtitle: sub, Marker: npcMarker(p, npc), Options: opts,
	})}
}

func (w *WorldState) startStoryQuiz(p *Player, questID, questionID string) [][]byte {
	idx := -1
	for i, q := range questionCatalog {
		if q.ID == questionID {
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
	p.ensureLog().Quiz = QuizSession{QuestID: questID, Index: idx, Active: true}
	p.markDirty()
	w.persist(p)
	q := questionCatalog[idx]
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "edu-story", TargetID: questID, Title: "Pitakonan", Speaker: "Mbah Jagat",
		Text:     q.Prompt,
		Question: &QuestionOut{ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices, ToID: p.ID},
	})}
}

func (w *WorldState) answerStoryQuiz(p *Player, in EducationAnswerIn) [][]byte {
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
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: false, Explain: def.Explain, Retry: true, Toast: "Coba lagi.",
			Question: questionOut(idx),
		})}
	}
	q := p.quest(log.Quiz.QuestID)
	p.credit("ANSWER", def.ID, 1)
	p.grantKnowledge(1)
	w.grantEducationBonus(p)
	log.Quiz.Active = false
	if q != nil && q.State == QuestActive {
		if def, ok := questByID[q.ID]; ok && objectivesDone(q, def) {
			q.State = QuestCompleted
		}
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
		Correct: true, Explain: def.Explain, Toast: "Knowledge Point +1",
	}), marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase))}
}

func isStoryQuiz(questID string) bool {
	return questID == "stq-edu-apel" || questID == "stq-latihan" || questID == "darkness-rising" || strings.HasPrefix(questID, "stq-")
}

func (w *WorldState) afterStoryQuestClaim(p *Player, def QuestDef) [][]byte {
	log := p.ensureLog()
	var out [][]byte
	switch def.ID {
	case "stq001":
		log.StoryState = StoryActive
		log.Flags["story_started"] = true
		log.Flags["ach_first_journey"] = true
		p.grantTitle("first-journey")
		p.grantTitle("traveler")
		out = append(out, w.startCinematic(p, "cin-opening", false)...)
	case "stq003":
		log.Flags["story_chapter_1_complete"] = true
		log.Flags["st-ch01_complete"] = true
		log.Flags["story_chapter_2_available"] = true
		log.Flags["st-ch02_available"] = true
		log.StoryChapter = "st-ch02"
		p.grantTitle("forest-friend")
		out = append(out, w.startCinematic(p, "cin-chapter-2", false)...)
	case "stq-allies":
		log.Flags["story_chapter_7_complete"] = true
		log.Flags["st-ch07_complete"] = true
		log.StoryChapter = "st-ch08"
		out = append(out, w.startCinematic(p, "cin-final-war", false)...)
		if w.Events.Active == nil || w.Events.Active.Def.ID != "darkness-rising" {
			out = append(out, w.startWorldEvent("darkness-rising")...)
		}
	case "stq-latihan", "tq001":
		out = append(out, w.startCinematic(p, "cin-awakened", false)...)
	}
	if def.CinematicID != "" && def.ID != "stq001" && def.ID != "stq003" {
		out = append(out, w.startCinematic(p, def.CinematicID, false)...)
	}
	p.markDirty()
	w.persist(p)
	w.refreshAchievements(p)
	return out
}

func (w *WorldState) afterGuardianStory(p *Player, g GuardianDef) [][]byte {
	log := p.ensureLog()
	if log.SilumanProg == nil {
		log.SilumanProg = map[string]string{}
	}
	if log.SilumanProg[g.ID] == "" {
		log.SilumanProg[g.ID] = SilumanDefeated
	}
	n := 0
	for _, st := range log.SilumanProg {
		if st != "" {
			n++
		}
	}
	if n >= 5 {
		log.Flags["ach_siluman_scholar"] = true
		p.grantTitle("siluman-scholar")
	}
	if n >= 33 {
		log.Flags["ach_33_siluman"] = true
	}
	lang := playerLang(p)
	st := silumanByGuard[g.ID]
	text, sub := locPair("story.siluman.jaladara.intro", lang)
	if st.Final {
		text, sub = locPair("story.siluman.adityara.intro", lang)
	}
	opts := []DialogOption{
		{ID: "choice-a", Label: locText("story.choice.spare", lang)},
		{ID: "choice-b", Label: locText("story.choice.leave", lang)},
		{ID: "choice-c", Label: locText("story.choice.ask", lang)},
	}
	if st.Final {
		opts = []DialogOption{
			{ID: "choice-a", Label: locText("story.choice.fight", lang)},
			{ID: "choice-b", Label: locText("story.choice.talk", lang)},
			{ID: "choice-c", Label: locText("story.choice.friends", lang)},
		}
	}
	out := [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "siluman", TargetID: g.ID, Title: st.Name, Speaker: st.Name, Role: "Siluman",
		Text: text, Subtitle: sub, Options: opts,
	})}
	cin := "cin-jaladara"
	if st.Final {
		cin = "cin-adityara"
	} else if st.MiniBoss {
		cin = "cin-chapter-5"
	}
	out = append(out, w.startCinematic(p, cin, false)...)
	if n >= 1 && log.Flags["story_chapter_1_complete"] && !log.Flags["story_chapter_2_complete"] {
		log.Flags["story_chapter_2_complete"] = true
		log.Flags["st-ch02_complete"] = true
		log.StoryChapter = "st-ch03"
	}
	p.markDirty()
	w.persist(p)
	return out
}

func (w *WorldState) claimStoryChapter(p *Player, chapterID string) [][]byte {
	def, ok := storyChapterByID[chapterID]
	if !ok {
		return rejectFor(p.ID, TypeClaimStoryChapter, "chapter")
	}
	log := p.ensureLog()
	if log.ChapterRewards == nil {
		log.ChapterRewards = map[string]bool{}
	}
	if log.ChapterRewards[chapterID] {
		return rejectFor(p.ID, TypeClaimStoryChapter, "claimed")
	}
	complete := storyChapterState(log, chapterID) == StoryCompleted ||
		log.Flags["story_chapter_"+strconv.Itoa(def.Index)+"_complete"] ||
		log.Flags[chapterID+"_complete"]
	if !complete {
		return rejectFor(p.ID, TypeClaimStoryChapter, "incomplete")
	}
	log.ChapterRewards[chapterID] = true
	if def.RewardTitle != "" {
		p.grantTitle(def.RewardTitle)
	}
	events := w.giveExp(p, def.RewardExp)
	p.markDirty()
	w.persist(p)
	out := [][]byte{marshal(TypeEventReward, map[string]any{
		"chapterId": chapterID, "exp": def.RewardExp, "title": def.RewardTitle, "toId": p.ID,
	})}
	out = append(out, events...)
	return out
}

func (w *WorldState) finishMainStory(p *Player) [][]byte {
	log := p.ensureLog()
	end := pickStoryEnding(log)
	log.EndingID = end.ID
	log.Flags["story_chapter_8_complete"] = true
	log.Flags["st-ch08_complete"] = true
	log.Flags["story_chapter_9_complete"] = true
	log.Flags["st-ch09_complete"] = true
	log.StoryChapter = "st-ch09"
	log.StoryState = StoryCompleted
	p.grantTitle(end.TitleReward)
	if end.Cosmetic != "" {
		p.grantCosmetic(end.Cosmetic)
	}
	if log.RedeemScore >= log.HelpScore && log.RedeemScore >= log.PowerScore {
		p.grantTitle("peace-bringer")
		log.Flags["ach_peace_bringer"] = true
	}
	log.Flags["ach_guardian_of_dawn"] = true
	p.grantTitle("guardian-of-dawn")
	out := w.startCinematic(p, end.CinematicID, false)
	out = append(out, w.completeStoryChapter(p)...)
	p.markDirty()
	w.persist(p)
	return out
}

func pickStoryEnding(log *PlayerLog) StoryEndingDef {
	if len(storyEndingCatalog) == 0 {
		return StoryEndingDef{ID: "ending-a", Title: "CAHAYA UNTUK KABEH", CinematicID: "cin-ending-a", TitleReward: "guardian-of-dawn"}
	}
	help, power, redeem := log.HelpScore, log.PowerScore, log.RedeemScore
	redeem += len(log.StoryAllies) * 2
	id := "ending-a"
	if power > help && power > redeem {
		id = "ending-b"
	} else if redeem > help && redeem >= power {
		id = "ending-c"
	}
	for _, e := range storyEndingCatalog {
		if e.ID == id {
			return e
		}
	}
	return storyEndingCatalog[0]
}

func (w *WorldState) startNewGamePlus(p *Player) [][]byte {
	log := p.ensureLog()
	if !log.Flags["storyCompleted"] {
		return rejectFor(p.ID, TypeStartNGPlus, "locked")
	}
	log.NGPlus++
	log.StoryChapter = "st-ch01"
	log.StoryState = StoryNotStarted
	log.PendingCinematic = ""
	log.EndingID = ""
	log.Flags["opening_cinematic_seen"] = false
	log.Flags["story_started"] = false
	p.markDirty()
	w.persist(p)
	return w.getStory(p, true)
}

func (e *LiveEvent) darknessPhase(now time.Time) string {
	if e == nil || e.Def.ID != "darkness-rising" {
		return ""
	}
	left := e.Until.Sub(now).Seconds()
	dur := e.Def.DurationSec
	if dur < 1 {
		dur = 1200
	}
	elapsed := dur - left
	pct := elapsed / dur
	switch {
	case pct < 0.25:
		return "PHASE 1 · Protect Village"
	case pct < 0.5:
		return "PHASE 2 · Defeat Shadow Army"
	case pct < 0.75:
		return "PHASE 3 · Open Gate"
	default:
		return "PHASE 4 · Final Boss"
	}
}

func storyJournalFields(p *Player, j *WorldJournal) {
	log := p.ensureLog()
	j.Language = playerLang(p)
	j.StoryChapter = log.StoryChapter
	j.StoryState = log.StoryState
	j.Allies = append([]string{}, log.StoryAllies...)
	j.EndingID = log.EndingID
	j.NGPlus = log.NGPlus
	if obj := storyObjective(log, j.Language); obj != "" && !log.Flags["storyCompleted"] {
		j.Objective = obj
	}
	chs := make([]StoryChapterView, 0, len(storyChapterCatalog))
	for _, c := range storyChapterCatalog {
		chs = append(chs, StoryChapterView{ID: c.ID, Index: c.Index, Title: c.Title, State: storyChapterState(log, c.ID), Music: c.Music})
	}
	j.StoryChapters = chs
}

func silumanViewExtra(log *PlayerLog, g GuardianDef, view *GuardianView) {
	st := silumanByGuard[g.ID]
	view.StoryName = st.Name
	status := ""
	if log.SilumanProg != nil {
		status = log.SilumanProg[g.ID]
	}
	if status == "" && (log.Guardians[g.ID] || log.Flags["guardian_"+g.ID+"_defeated"]) {
		status = SilumanDefeated
	}
	view.CodexStatus = status
	view.Ally = status == SilumanAlly
	view.MiniBoss = st.MiniBoss
	if status == SilumanAlly {
		view.Status = SilumanAlly
	} else if status == SilumanUnderstood {
		view.Status = SilumanUnderstood
	}
}
