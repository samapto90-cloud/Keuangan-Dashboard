package mmo

import (
	"math"
	"strconv"

	_ "embed"
)

//go:embed data/siluman.json
var silumanJSON []byte

//go:embed data/achievements.json
var achievementsJSON []byte

type silumanOverlay struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Title                 string   `json:"title"`
	Personality           string   `json:"personality"`
	Weakness              string   `json:"weakness"`
	EducationQuestionPool []string `json:"educationQuestionPool"`
	Story                 string   `json:"story"`
	Intro                 string   `json:"intro"`
	Encounter             string   `json:"encounter"`
	Defeat                string   `json:"defeat"`
	Aftermath             string   `json:"aftermath"`
}

type AchievementDef struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	Flag         string `json:"flag"`
	MinGuardians int    `json:"minGuardians"`
	MinEdu       int    `json:"minEdu"`
	Category     string `json:"category"`
}

var achievementCatalog []AchievementDef

func applySilumanOverlay() {
	var overlay []silumanOverlay
	mustJSON("siluman.json", silumanJSON, &overlay)
	byID := map[string]silumanOverlay{}
	for _, o := range overlay {
		byID[o.ID] = o
	}
	for i := range guardianCatalog {
		g := &guardianCatalog[i]
		o, ok := byID[g.ID]
		if !ok {
			continue
		}
		if o.Name != "" {
			g.Name = o.Name
		}
		if o.Title != "" {
			g.Title = o.Title
		}
		g.Personality = o.Personality
		g.Weakness = o.Weakness
		g.EducationPool = o.EducationQuestionPool
		if o.Story != "" {
			g.Story = o.Story
		}
		if o.Intro != "" {
			g.Intro = o.Intro
		}
		if o.Encounter != "" {
			g.Encounter = o.Encounter
		}
		if o.Defeat != "" {
			g.Defeat = o.Defeat
		}
		if o.Aftermath != "" {
			g.Aftermath = o.Aftermath
		}
	}
	if len(achievementsJSON) > 0 {
		mustJSON("achievements.json", achievementsJSON, &achievementCatalog)
	}
}

func regionStreamVisible(playerZ float64, r RegionDef) bool {
	if playerZ >= r.MinZ && playerZ < r.MaxZ {
		return true
	}
	dist := math.Min(math.Abs(playerZ-r.MinZ), math.Abs(playerZ-r.MaxZ))
	return dist < 36
}

func (p *Player) channelID() string {
	if p.Channel == "" {
		return "1"
	}
	return p.Channel
}

func (w *WorldState) channelCount(ch string) int {
	n := 0
	for _, p := range w.players {
		if p.Connected && p.channelID() == ch {
			n++
		}
	}
	return n
}

func (w *WorldState) leastChannel() string {
	best, bestN := "1", ChannelMaxPop+1
	for i := 1; i <= ChannelCount; i++ {
		id := strconv.Itoa(i)
		n := w.channelCount(id)
		if n < bestN {
			best, bestN = id, n
		}
	}
	return best
}

func (w *WorldState) switchChannel(p *Player, ch string) [][]byte {
	if ch == "" {
		ch = "1"
	}
	ok := false
	for i := 1; i <= ChannelCount; i++ {
		if ch == strconv.Itoa(i) {
			ok = true
			break
		}
	}
	if !ok {
		return rejectFor(p.ID, TypeSwitchChannel, "channel")
	}
	if p.InstanceID != "" {
		return rejectFor(p.ID, TypeSwitchChannel, "instance")
	}
	if p.channelID() == ch {
		return [][]byte{marshal(TypeChannelList, w.channelList(p))}
	}
	n := w.channelCount(ch)
	if n >= ChannelMaxPop {
		return rejectFor(p.ID, TypeSwitchChannel, "full")
	}
	p.Channel = ch
	return [][]byte{marshal(TypeChannelList, w.channelList(p))}
}

func (w *WorldState) channelList(p *Player) map[string]any {
	chs := make([]map[string]any, 0, ChannelCount)
	for i := 1; i <= ChannelCount; i++ {
		id := strconv.Itoa(i)
		chs = append(chs, map[string]any{"id": id, "name": "Channel " + id, "online": w.channelCount(id), "max": ChannelMaxPop})
	}
	cur := "1"
	if p != nil {
		cur = p.channelID()
	}
	return map[string]any{"channel": cur, "channels": chs, "toId": p.ID}
}

func (w *WorldState) enterMasjid(p *Player, first bool) [][]byte {
	log := p.ensureLog()
	if p.Mounted {
		w.dismount(p, "sanctuary")
	}
	p.credit("VISIT", "masjid", 1)
	if !first && log.Flags["masjid_entered"] {
		return nil
	}
	log.Flags["masjid_entered"] = true
	log.Flags["masjid_path_open"] = true
	p.grantTitle("light-seeker")
	w.refreshAchievements(p)
	p.markDirty()
	w.persist(p)
	w.audit("questCompleted", p.ID, "enter_masjid")
	return [][]byte{marshal(TypeEnterMasjid, map[string]any{
		"playerId": p.ID, "combat": false, "music": "calm",
		"text": "Tujuan perjalanan bukan hanya kekuatan. Ilmu, kesabaran, keberanian, dan kebaikan.",
		"skip": true, "duration": 20,
	})}
}

func (w *WorldState) completeStoryChapter(p *Player) [][]byte {
	log := p.ensureLog()
	if log.Flags["storyCompleted"] && log.Flags["chapter_complete_shown"] {
		return nil
	}
	log.Flags["storyCompleted"] = true
	log.Flags["explorer_mode"] = true
	log.Flags["chapter_complete_shown"] = true
	p.X, p.Y, p.Z = 0, 0, 6
	p.VX, p.VZ = 0, 0
	p.ZoneID = "village"
	p.grantTitle("light-seeker")
	w.refreshAchievements(p)
	p.markDirty()
	w.persist(p)
	n := 0
	for _, ok := range log.Guardians {
		if ok {
			n++
		}
	}
	return [][]byte{marshal(TypeChapterComplete, map[string]any{
		"title": "CHAPTER COMPLETE", "guardians": n, "guardiansTotal": len(guardianCatalog),
		"masjid": "DISCOVERED", "toId": p.ID, "explorerMode": true,
	})}
}

func (w *WorldState) collectionBook(p *Player) CollectionBook {
	log := p.ensureLog()
	g := 0
	for _, ok := range log.Guardians {
		if ok {
			g++
		}
	}
	loc := 0
	for _, ok := range log.DiscoveredZones {
		if ok {
			loc++
		}
	}
	auraN, mountN := 0, 0
	for _, c := range log.Cosmetics {
		if cosmeticByID[c].Kind == "aura" {
			auraN++
		}
		if cosmeticByID[c].Kind == "mount" {
			mountN++
		}
	}
	return CollectionBook{
		PlayerID: p.ID, Guardians: g, GuardiansTotal: len(guardianCatalog), Tokens: log.GuardianTokens,
		Locations: loc, Titles: append([]string{}, log.Titles...), Cosmetics: append([]string{}, log.Cosmetics...),
		Achievements:   append([]string{}, log.Achievements...),
		StoryCompleted: log.Flags["storyCompleted"], ExplorerMode: log.Flags["explorer_mode"] || log.Flags["storyCompleted"],
		Aura: auraN, AuraTotal: countCosmeticKind("aura"), Mount: mountN, MountTotal: countCosmeticKind("mount"),
	}
}

func (w *WorldState) refreshAchievements(p *Player) [][]byte {
	log := p.ensureLog()
	if log.Achievements == nil {
		log.Achievements = []string{}
	}
	n := 0
	for _, ok := range log.Guardians {
		if ok {
			n++
		}
	}
	var out [][]byte
	grant := func(id, title string) {
		for _, a := range log.Achievements {
			if a == id {
				return
			}
		}
		log.Achievements = append(log.Achievements, id)
		if title != "" {
			p.grantTitle(title)
		}
		w.audit("achievementUnlocked", p.ID, id)
		out = append(out, marshal(TypeAchievementUnlocked, map[string]any{"id": id, "toId": p.ID}))
		out = append(out, marshal(TypeEventReward, map[string]any{"achievement": id, "toId": p.ID}))
		w.notify(p, "achievement", "Achievement unlocked")
	}
	if log.DiscoveredZones["village"] || log.Flags["zone_village_discovered"] {
		grant("first-step", "first-step")
	}
	if log.DiscoveredZones["forest"] {
		grant("forest-explorer", "forest-explorer")
	}
	if n >= 1 {
		grant("guardian-hunter", "guardian-hunter")
	}
	if n >= 33 {
		grant("33-guardians", "33-guardians")
	}
	if log.Flags["masjid_entered"] || log.Flags["storyCompleted"] {
		grant("light-seeker", "light-seeker")
	}
	if log.Flags["ach_first_journey"] || log.Flags["story_started"] || log.Flags["opening_cinematic_seen"] {
		grant("first-journey", "first-journey")
	}
	if log.Flags["ach_siluman_scholar"] {
		grant("siluman-scholar", "siluman-scholar")
	}
	if log.Flags["ach_peace_bringer"] {
		grant("peace-bringer", "peace-bringer")
	}
	if log.Flags["ach_guardian_of_dawn"] || log.Flags["storyCompleted"] {
		grant("guardian-of-dawn", "guardian-of-dawn")
	}
	w.refreshProfessionAchievements(p, grant)
	for _, def := range achievementCatalog {
		if def.Flag != "" && log.Flags[def.Flag] {
			grant(def.ID, def.Title)
		}
		if def.MinGuardians > 0 && n >= def.MinGuardians {
			grant(def.ID, def.Title)
		}
		if def.MinEdu > 0 && log.EduCorrect >= def.MinEdu {
			grant(def.ID, def.Title)
		}
	}
	return out
}

func (w *WorldState) grantGuardianCollection(p *Player, n int) {
	switch n {
	case 5:
		p.grantTitle("mist-walker")
	case 10:
		p.grantCosmetic("cloak-mistwood")
		p.grantCosmetic("emote-bow")
	case 20:
		p.grantCosmetic("aura-guardian-20")
	case 33:
		p.grantTitle("33-guardians")
		p.grantCosmetic("aura-celestial")
	}
}

func (p *Player) inCombatDisabledZone() bool {
	if houseID(p.InstanceID) {
		return true
	}
	r := zoneAt(p.X, p.Z)
	return r.CombatDisabled || r.ID == "masjid"
}

func (p *Player) inOpenWorldSafeZone() bool {
	r := zoneAt(p.X, p.Z)
	return r.SafeZone || r.ID == "village" || r.ID == "masjid"
}

func fastTravelCost(level int) int {
	if level >= 25 {
		return 5
	}
	return 0
}

func (w *WorldState) breakEducationShield(p *Player) {
	log := p.ensureLog()
	for _, e := range w.enemies {
		if e == nil || !e.Alive {
			continue
		}
		g, ok := guardianByID[e.Def.ID]
		if !ok || len(g.EducationPool) == 0 {
			continue
		}
		if nearby(p.X, p.Z, e.X, e.Z, 16) {
			log.Flags["edu_boss_"+g.ID] = true
		}
	}
	if w.Boss != nil && w.Boss.State == "ACTIVE" {
		if e := w.enemies[w.Boss.EnemyID]; e != nil && nearby(p.X, p.Z, e.X, e.Z, 18) {
			log.Flags["wb_shield_down"] = true
		}
	}
}

func auraUnlockByGuardians(n int) string {
	switch {
	case n >= 30:
		return "celestial-4"
	case n >= 20:
		return "aura-3"
	case n >= 10:
		return "aura-2"
	case n >= 3:
		return "aura-1"
	default:
		return ""
	}
}
