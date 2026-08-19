package mmo

import (
	"math"
	"sort"
	"strings"
	"time"

	_ "embed"
)

//go:embed data/factions.json
var factionsJSON []byte

const (
	RepNeutral  = "Neutral"
	RepFriendly = "Friendly"
	RepTrusted  = "Trusted"
	RepHonored  = "Honored"

	EncounterCooldownSec = 180
)

type FactionDef struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
	Title  string `json:"title"`
}

type npcVec struct {
	X, Z float64
}

type npcLive struct {
	X, Z, Yaw, Speed float64
	StuckT      float64
	LastX, LastZ float64
	WanderX, WanderZ float64
	WanderReady float64
	UsingWander bool
	MoveDelay   float64
}

var (
	factionCatalog []FactionDef
	factionByID    = map[string]FactionDef{}
)

var npcPlaces = map[string]npcVec{
	"square":    {0, 4},
	"market":    {9, 3},
	"school":    {-9, 2},
	"gate":      {0, 19.6},
	"clinic":    {7.2, 8.4},
	"home":      {-9, 8},
	"hall":      {-6.4, 1.2},
	"temple":    {0, 3.2},
	"masjid":    {0, 166},
	"forest":    {0, 36},
	"valley":    {0, 56},
	"horizon":   {0, 182},
	"village":   {0, 6},
	"workplace": {9, 3},
	"tavern":    {-6.4, 1.2},
	"house":     {-9, 8},
}

func init() {
	mustJSON("factions.json", factionsJSON, &factionCatalog)
	for _, f := range factionCatalog {
		factionByID[f.ID] = f
	}
}

func (t *WorldTimeSystem) ClockHour() int {
	if t.ClockMin < 0 {
		return 9
	}
	return (t.ClockMin / 60) % 24
}

func (t *WorldTimeSystem) ClockMinute() int {
	if t.ClockMin < 0 {
		return 0
	}
	return t.ClockMin % 60
}

func (t WorldTimeSystem) ClockLabel() string {
	h := t.ClockHour()
	switch {
	case h >= 6 && h < 12:
		return "Morning"
	case h >= 12 && h < 17:
		return "Afternoon"
	case h >= 17 && h < 21:
		return "Evening"
	default:
		return "Night"
	}
}

func (t WorldTimeSystem) ClockText() string {
	return pad2(t.ClockHour()) + ":" + pad2(t.ClockMinute())
}

func (t WorldTimeSystem) Slot() string {
	h := t.ClockHour()
	switch {
	case h >= 6 && h < 12:
		return "morning"
	case h >= 12 && h < 17:
		return "afternoon"
	case h >= 17 && h < 21:
		return "evening"
	default:
		return "night"
	}
}

func (t *WorldTimeSystem) AdvanceClock(gameMinutes int) {
	if gameMinutes < 0 {
		return
	}
	t.ClockMin = (t.ClockMin + gameMinutes) % 1440
	h := t.ClockHour()
	switch {
	case h >= 6 && h < 17:
		t.Phase = "DAY"
	case h >= 17 && h < 21:
		t.Phase = "EVENING"
	default:
		t.Phase = "NIGHT"
	}
}

func weatherForRegion(region, phase, global string) string {
	if global == "STORM" {
		return "STORM"
	}
	switch region {
	case "forest":
		if phase == "NIGHT" {
			return "FOG"
		}
		if global == "RAIN" {
			return "RAIN"
		}
		return "CLOUDY"
	case "plains":
		if global == "RAIN" {
			return "RAIN"
		}
		return "CLEAR"
	case "canyon":
		return "CLEAR"
	case "village", "masjid":
		if global == "" {
			return "CLEAR"
		}
		return global
	default:
		if global == "" {
			return weatherFor(phase)
		}
		return global
	}
}

func (w *WorldState) weatherForWatcher(p *Player) string {
	if p == nil {
		return w.Time.Weather
	}
	z := zoneAt(p.X, p.Z)
	if w.Time.RegionWeather != nil {
		if v := w.Time.RegionWeather[z.ID]; v != "" {
			return v
		}
	}
	return weatherForRegion(z.ID, w.Time.Phase, w.Time.Weather)
}

func npcDest(n NPCDef, t WorldTimeSystem) (float64, float64) {
	place := ""
	if n.Schedule != nil {
		place = n.Schedule[t.Slot()]
		if place == "" {
			switch t.Phase {
			case "NIGHT":
				place = n.Schedule["night"]
			case "EVENING":
				place = n.Schedule["evening"]
			default:
				place = n.Schedule["day"]
			}
		}
	}
	if n.Type == "MERCHANT" && (t.ClockHour() < 8 || t.ClockHour() >= 20) {
		if home, ok := n.Schedule["night"]; ok && home != "" {
			place = home
		} else {
			place = "home"
		}
	}
	if place == "" {
		return n.X, n.Z
	}
	if v, ok := npcPlaces[place]; ok {
		if place == "village" || place == "workplace" {
			return n.X, n.Z
		}
		return v.X, v.Z
	}
	return n.X, n.Z
}

func (w *WorldState) npcLive(n NPCDef) (float64, float64) {
	if w.npcPos != nil {
		if v, ok := w.npcPos[n.ID]; ok {
			return v.X, v.Z
		}
	}
	return npcDest(n, w.Time)
}

func (w *WorldState) npcLiveYaw(n NPCDef) float64 {
	if w.npcPos != nil {
		if v, ok := w.npcPos[n.ID]; ok {
			return v.Yaw
		}
	}
	return n.Yaw
}

func (w *WorldState) tickNPCs(dt float64) {
	if dt <= 0 {
		return
	}
	if dt > 0.25 {
		steps := int(dt/0.1 + 0.5)
		if steps < 1 {
			steps = 1
		}
		if steps > 400 {
			steps = 400
		}
		sub := dt / float64(steps)
		for i := 0; i < steps; i++ {
			w.stepNPCs(sub)
		}
		return
	}
	w.stepNPCs(dt)
}

func (w *WorldState) stepNPCs(dt float64) {
	if w.npcPos == nil {
		w.npcPos = map[string]npcLive{}
	}
	eventFlee := w.Events.Active != nil && (w.Events.Active.State == "ACTIVE" || w.Events.Active.State == "FINAL")
	players := w.villagePlayers()
	npcs := activeNPCList()
	// Build proximity lists once per tick.
	type npcRef struct {
		id string
		n  NPCDef
	}
	order := make([]npcRef, 0, len(npcs))
	for _, n := range npcs {
		order = append(order, npcRef{id: n.ID, n: n})
	}
	sort.Slice(order, func(i, j int) bool {
		return npcPriority(order[i].n.Type) > npcPriority(order[j].n.Type)
	})
	for _, ref := range order {
		n := ref.n
		cur, ok := w.npcPos[n.ID]
		if !ok {
			cur = npcLive{X: n.X, Z: n.Z, Yaw: n.Yaw, LastX: n.X, LastZ: n.Z, MoveDelay: npcStartDelay(n.ID)}
		}
		if cur.MoveDelay > 0 {
			cur.MoveDelay -= dt
			w.npcPos[n.ID] = cur
			continue
		}
		if n.Type == "QUEST_BOARD" {
			cur.X, cur.Z = n.X, n.Z
			cur.Speed = 0
			w.npcPos[n.ID] = cur
			continue
		}
		dx, dz := w.npcDestination(n)
		rad := npcBodyRadius(n)
		walk := npcWalkSpeed(n, eventFlee)
		hx, hz := dx-cur.X, dz-cur.Z
		dist := math.Hypot(hx, hz)
		dirX, dirZ := 0.0, 0.0
		if dist > 0.08 {
			dirX, dirZ = hx/dist, hz/dist
		}
		// Local avoidance: separation from NPCs + soft player steer.
		others := make([][3]float64, 0, len(w.npcPos))
		for oid, ol := range w.npcPos {
			if oid == n.ID {
				continue
			}
			orad := NPCRadius
			for _, cn := range npcs {
				if cn.ID == oid {
					orad = npcBodyRadius(cn)
					break
				}
			}
			others = append(others, [3]float64{ol.X, ol.Z, orad})
		}
		sepX, sepZ := softSeparation(cur.X, cur.Z, rad, others)
		for _, pl := range players {
			pdx, pdz := cur.X-pl[0], cur.Z-pl[1]
			pd := math.Hypot(pdx, pdz)
			if pd < npcAvoidanceRadius && pd > 0.001 {
				push := (npcAvoidanceRadius - pd) / pd * 0.45
				sepX += pdx * push
				sepZ += pdz * push
			}
		}
		if dirX != 0 || dirZ != 0 {
			dirX, dirZ = steerClearOfObstacles(cur.X, cur.Z, dirX, dirZ, rad)
			dirX += sepX * 1.6
			dirZ += sepZ * 1.6
			l := math.Hypot(dirX, dirZ)
			if l > 0.001 {
				dirX /= l
				dirZ /= l
			}
		} else if sepX != 0 || sepZ != 0 {
			dirX, dirZ = sepX, sepZ
			l := math.Hypot(dirX, dirZ)
			if l > 0.001 {
				dirX /= l
				dirZ /= l
			}
		}
		want := 0.0
		if dist > 0.22 && (dirX != 0 || dirZ != 0) {
			want = walk
			if dist < 1.6 {
				want = walk * math.Max(0.12, dist/1.6)
			}
			if npcStationary(n) && dist < 0.55 {
				want = 0
			}
			face := math.Atan2(dirX, dirZ)
			cur.Yaw = npcTurnToward(cur.Yaw, face, 8*dt)
			aligned := math.Abs(npcAngleDiff(cur.Yaw, face)) < 0.55
			if !aligned {
				want *= 0.25
			}
		}
		accel := 3.2
		if want < cur.Speed {
			accel = 5.5
		}
		if cur.Speed < want {
			cur.Speed += accel * dt
			if cur.Speed > want {
				cur.Speed = want
			}
		} else {
			cur.Speed -= accel * dt
			if cur.Speed < want {
				cur.Speed = want
			}
		}
		if cur.Speed < 0.04 {
			cur.Speed = 0
		}
		prevX, prevZ := cur.X, cur.Z
		if cur.Speed > 0 {
			cur.X += dirX * cur.Speed * dt
			cur.Z += dirZ * cur.Speed * dt
			cur.X, cur.Z = resolveWorldXZ(cur.X, cur.Z, rad)
			sx, sz := softSeparation(cur.X, cur.Z, rad, others)
			cur.X += sx * dt * 2.2
			cur.Z += sz * dt * 2.2
			cur.X, cur.Z = resolveWorldXZ(cur.X, cur.Z, rad)
		}
		moved := math.Hypot(cur.X-prevX, cur.Z-prevZ)
		if moved < 0.015 && dist > 0.6 {
			cur.StuckT += dt
		} else {
			cur.StuckT = 0
			cur.LastX, cur.LastZ = cur.X, cur.Z
		}
		if cur.StuckT > 2.5 {
			cur.WanderX, cur.WanderZ = pickAlternateDest(n, cur, w.Time.Seconds(), w.npcPos)
			cur.UsingWander = true
			cur.StuckT = 0
		}
		w.npcPos[n.ID] = cur
	}
}

func npcParkOffset(id string) (float64, float64) {
	h := 0
	for i := 0; i < len(id); i++ {
		h = h*31 + int(id[i])
	}
	a := float64(h%628) / 100.0
	return math.Cos(a) * 0.5, math.Sin(a) * 0.5
}

func npcAngleDiff(from, to float64) float64 {
	d := to - from
	for d > math.Pi {
		d -= 2 * math.Pi
	}
	for d < -math.Pi {
		d += 2 * math.Pi
	}
	return d
}

func npcTurnToward(from, to, maxStep float64) float64 {
	d := npcAngleDiff(from, to)
	if d > maxStep {
		d = maxStep
	}
	if d < -maxStep {
		d = -maxStep
	}
	return from + d
}

func npcActivity(n NPCDef, t WorldTimeSystem, eventActive bool) string {
	if eventActive && n.Type != "GUARD" {
		if n.Type == "CHILD" || n.Type == "TEACHER" {
			return "SHELTER"
		}
		return "GATHER"
	}
	if t.Slot() == "night" && n.Schedule["night"] == "home" {
		return "SLEEP"
	}
	if n.Type == "MERCHANT" && (t.ClockHour() < 8 || t.ClockHour() >= 20) {
		return "CLOSED"
	}
	return strings.ToUpper(t.Slot())
}

func repTier(v int) string {
	switch {
	case v >= 75:
		return RepHonored
	case v >= 50:
		return RepTrusted
	case v >= 25:
		return RepFriendly
	default:
		return RepNeutral
	}
}

func (p *Player) bumpRegionRep(region string, amount int) {
	log := p.ensureLog()
	if log.RegionRep == nil {
		log.RegionRep = map[string]int{}
	}
	log.RegionRep[region] += amount
	if log.RegionRep[region] < 0 {
		log.RegionRep[region] = 0
	}
	if log.RegionRep[region] > 100 {
		log.RegionRep[region] = 100
	}
	tier := repTier(log.RegionRep[region])
	switch tier {
	case RepHonored:
		p.grantTitle("dawn-honored")
		p.grantCosmetic("dawn-sash")
	case RepTrusted:
		p.grantTitle("dawn-trusted")
	case RepFriendly:
		p.grantTitle("dawn-friend")
	}
	p.markDirty()
}

func (p *Player) bumpFactionRep(faction string, amount int) {
	if faction == "" {
		return
	}
	log := p.ensureLog()
	if log.FactionRep == nil {
		log.FactionRep = map[string]int{}
	}
	log.FactionRep[faction] += amount
	if log.FactionRep[faction] < 0 {
		log.FactionRep[faction] = 0
	}
	if log.FactionRep[faction] > 100 {
		log.FactionRep[faction] = 100
	}
	if def, ok := factionByID[faction]; ok && repTier(log.FactionRep[faction]) == RepHonored && def.Title != "" {
		p.grantTitle(def.Title)
	}
	p.markDirty()
}

func (p *Player) grantKnowledge(n int) {
	if n <= 0 {
		return
	}
	log := p.ensureLog()
	log.KnowledgePoints += n
	p.markDirty()
}

func (w *WorldState) npcChoice(p *Player, npc NPCDef, choice string) [][]byte {
	log := p.ensureLog()
	if log.DialogueChoices == nil {
		log.DialogueChoices = map[string]string{}
	}
	log.DialogueChoices[npc.ID] = choice
	text := lineOr("mira_welcome", "Selamat datang di Dawn Village.")
	switch choice {
	case "choice-a", "A", "a":
		text = lineOr("mira_choice_a", "Baik. Pelan-pelan saja. Desa akan mengingat kebaikanmu.")
		p.bumpRegionRep("village", 8)
		p.bumpFactionRep("dawn-keepers", 6)
	case "choice-b", "B", "b":
		text = lineOr("mira_choice_b", "Di hutan ada kuil kecil. Jawab soal sederhana, lalu cahaya membantumu.")
		p.grantKnowledge(1)
	case "choice-c", "C", "c":
		text = lineOr("mira_choice_c", "Kalau terburu-buru, tetap ikuti jalan utama. Cerita utamamu tidak akan hilang.")
	}
	p.credit("TALK", npc.ID, 1)
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "npc", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
		Text: text, Marker: npcMarker(p, npc),
		Options: append(miraOptions(p, npc), DialogOption{ID: "close", Label: "Tutup"}),
	})}
}

func (w *WorldState) startShrineQuiz(p *Player, obj InteractDef) [][]byte {
	qid := obj.QuestionID
	if qid == "" {
		qid = "q-add-3-2"
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
	log := p.ensureLog()
	log.Quiz = QuizSession{QuestID: "edu-shrine", Index: idx, Active: true}
	p.credit("DISCOVER", "edu-shrine", 1)
	p.markDirty()
	w.persist(p)
	q := questionCatalog[idx]
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "edu-shrine", TargetID: obj.ID, Title: "Educational Shrine", Speaker: "Shrine of Dawn",
		Text: "Jawab soal kecil. Kelas 1.",
		Question: &QuestionOut{
			ID: q.ID, Index: 1, Total: 1, Category: q.Category, Prompt: q.Prompt, Choices: q.Choices, ToID: p.ID,
		},
	})}
}

func (w *WorldState) answerShrine(p *Player, in EducationAnswerIn) [][]byte {
	log := p.ensureLog()
	if !log.Quiz.Active || log.Quiz.QuestID != "edu-shrine" {
		return rejectFor(p.ID, TypeEducationAnswer, "no_session")
	}
	if log.Quiz.Index < 0 || log.Quiz.Index >= len(questionCatalog) {
		return rejectFor(p.ID, TypeEducationAnswer, "question")
	}
	def := questionCatalog[log.Quiz.Index]
	if def.ID != in.QuestionID {
		return rejectFor(p.ID, TypeEducationAnswer, "order")
	}
	if in.Choice != def.Correct {
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: false, Explain: def.Explain, Retry: true, Toast: "Coba lagi.",
			Question: &QuestionOut{ID: def.ID, Index: 1, Total: 1, Category: def.Category, Prompt: def.Prompt, Choices: def.Choices},
		})}
	}
	log.Quiz.Active = false
	p.Energy = min(p.MaxEnergy, p.Energy+12)
	p.grantKnowledge(2)
	p.credit("ANSWER", "edu-shrine", 1)
	p.credit("DISCOVER", "edu-shrine", 1)
	p.bumpFactionRep("celestial-scholars", 4)
	events := w.grantExp(p, 12)
	events = append(events, w.recordEducation(p, def.Category, true)...)
	p.markDirty()
	w.persist(p)
	events = append(events, marshal(TypeEducationFeedback, EducationFeedback{
		Correct: true, Explain: def.Explain, Toast: "+Energy · +EXP · +Knowledge Point",
	}))
	return events
}

func (w *WorldState) maybeEncounter(p *Player) [][]byte {
	z := zoneAt(p.X, p.Z)
	if z.SafeZone || z.ID == "village" || z.ID == "masjid" {
		return nil
	}
	log := p.ensureLog()
	now := time.Now()
	if log.EncounterAt > 0 && now.Unix()-log.EncounterAt < EncounterCooldownSec {
		return nil
	}
	if now.Unix()%17 != 0 {
		return nil
	}
	log.EncounterAt = now.Unix()
	kind := []string{"merchant", "lost-child", "broken-cart", "wild-beast", "mysterious-traveler", "lost-traveler"}[int(now.Unix()/17)%6]
	text := "Seseorang menyeberang jalan."
	switch kind {
	case "merchant":
		text = "Pedagang keliling menawarkan ramuan sederhana."
	case "lost-child":
		text = "Seorang anak mencari jalan pulang ke Dawn Village."
		p.bumpRegionRep("village", 2)
	case "broken-cart":
		text = "Gerobak rusak di tepi jalan. Bantu dorong sebentar."
	case "wild-beast":
		text = "Jejak beast liar. Tetap di jalan utama."
		if w.Time.Phase == "NIGHT" {
			events := w.maybeAmbush(p)
			if len(events) > 0 {
				return append([][]byte{marshal(TypeRandomEncounter, map[string]any{
					"kind": kind, "text": text, "toId": p.ID,
				})}, events...)
			}
		}
	case "mysterious-traveler":
		text = "Pengembara misterius menunjuk ke Sanctum of Light."
	case "lost-traveler":
		text = "Le, aku kesasar. Iso tulung aku?"
		log.TravelEvent = "lost-traveler"
	}
	p.markDirty()
	return [][]byte{marshal(TypeRandomEncounter, map[string]any{
		"kind": kind, "text": text, "toId": p.ID,
	})}
}

func (w *WorldState) offerDynamicQuests(p *Player) {
	q := p.quest("oq001")
	if q != nil && q.State == QuestLocked && prereqMet(p, questByID["oq001"]) {
		q.State = QuestAvailable
		p.markDirty()
	}
	wq := p.quest("wq-save-dawn")
	if wq != nil && wq.State == QuestLocked && prereqMet(p, questByID["wq-save-dawn"]) {
		if w.Events.Active != nil && w.Events.Active.Def.ID == "shadow-attack" {
			wq.State = QuestAvailable
			p.markDirty()
		}
	}
}

func (e *LiveEvent) eventPhase(now time.Time) string {
	if e == nil {
		return ""
	}
	if p := e.darknessPhase(now); p != "" {
		return p
	}
	switch e.State {
	case "WAITING", "ACTIVE", "FINAL", "COMPLETE", "FAILED", "SUCCESS", "EXPIRED":
		if e.State == "ACTIVE" && !e.Until.IsZero() && e.Until.Sub(now).Seconds() < e.Def.DurationSec*0.2 {
			return "FINAL"
		}
		if e.State == "SUCCESS" {
			return "COMPLETE"
		}
		if e.State == "EXPIRED" {
			return "FAILED"
		}
		return e.State
	default:
		return e.State
	}
}

func (e *LiveEvent) scaledMobCount() int {
	n := 2 + len(e.Score)
	if n < 2 {
		n = 2
	}
	max := worldCfg.MaxEnemiesCell
	if max < 2 {
		max = 6
	}
	if n > max {
		n = max
	}
	return n
}

func areaAt(x, z float64) *AreaDef {
	var best *AreaDef
	bestR := 1e9
	for i := range areaCatalog {
		a := &areaCatalog[i]
		d := math.Hypot(x-a.X, z-a.Z)
		if d <= a.Radius && d < bestR {
			best = a
			bestR = d
		}
	}
	return best
}

func (w *WorldState) discoverPOI(p *Player, id, name string) [][]byte {
	log := p.ensureLog()
	if log.POI == nil {
		log.POI = map[string]bool{}
	}
	if log.POI[id] {
		return nil
	}
	log.POI[id] = true
	p.grantKnowledge(1)
	p.bumpRegionRep(zoneAt(p.X, p.Z).ID, 3)
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeLandmarkDiscovered, map[string]any{
		"id": id, "name": name, "poi": true, "toId": p.ID, "toast": "POINT OF INTEREST",
	})}
}
