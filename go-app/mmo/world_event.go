package mmo

import (
	"strconv"
	"strings"
	"time"
)

type WorldEvent struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type LiveEvent struct {
	Def       EventDef
	State     string
	Until     time.Time
	ActiveAt  time.Time
	GateHP    int
	MaxGateHP int
	Score     map[string]int
	Claimed   map[string]bool
	Region    string
	SpawnID   string
}

type WorldEventManager struct {
	Events []WorldEvent
	Active *LiveEvent
	Last   *LiveEvent
	nextAt time.Time
	rotate int
}

func NewWorldEventManager() WorldEventManager {
	return WorldEventManager{Events: defaultWorldEvents(), nextAt: time.Now().Add(12 * time.Minute)}
}

func defaultWorldEvents() []WorldEvent {
	out := make([]WorldEvent, 0, len(eventCatalog))
	for _, d := range eventCatalog {
		out = append(out, WorldEvent{ID: d.ID, Name: d.Name, State: "SCHEDULED"})
	}
	if len(out) == 0 {
		out = []WorldEvent{{ID: "village-defense", Name: "Village Under Attack", State: "SCHEDULED"}}
	}
	return out
}

func (w *WorldState) tickWorldEvents(now time.Time) [][]byte {
	ev := w.Events.Active
	if ev != nil {
		if ev.State == "WAITING" && !ev.ActiveAt.IsZero() && now.After(ev.ActiveAt) {
			ev.State = "ACTIVE"
			w.spawnEventMobs(ev)
			return [][]byte{marshal(TypeWorldEvent, ev.view())}
		}
		if ev.State == "ACTIVE" && ev.Until.Sub(now).Seconds() < ev.Def.DurationSec*0.2 {
			ev.State = "FINAL"
			return [][]byte{marshal(TypeWorldEvent, ev.view())}
		}
		if now.After(ev.Until) {
			success := eventSucceeded(ev)
			return w.finishWorldEvent(ev, success)
		}
		return nil
	}
	if now.After(w.Events.nextAt) {
		if !w.adventureReadyForWorldEvent() {
			w.Events.nextAt = now.Add(2 * time.Minute)
			return nil
		}
		id := "village-defense"
		if w.Events.rotate > 0 && len(eventCatalog) > 0 {
			id = eventCatalog[w.Events.rotate%len(eventCatalog)].ID
		}
		w.Events.rotate++
		return w.startWorldEvent(id)
	}
	return nil
}

func (w *WorldState) adventureReadyForWorldEvent() bool {
	for _, p := range w.players {
		if p == nil || !p.Connected {
			continue
		}
		log := p.ensureLog()
		if log.ForestUnlocked || log.Flags["forest_unlocked"] || log.Flags["forest_threat_known"] || log.Flags["training_complete"] {
			return true
		}
	}
	return false
}

func (w *WorldState) startWorldEvent(id string) [][]byte {
	def, ok := eventByID[id]
	if !ok {
		return nil
	}
	dur := def.DurationSec
	if dur < 20 {
		dur = 90
	}
	now := time.Now()
	live := &LiveEvent{
		Def: def, State: "ACTIVE", ActiveAt: now, Until: now.Add(time.Duration(dur * float64(time.Second))),
		GateHP: def.GateHP, MaxGateHP: def.GateHP, Score: map[string]int{}, Claimed: map[string]bool{}, Region: def.Region,
		SpawnID: strconv.FormatInt(now.UnixNano(), 10),
	}
	if live.GateHP <= 0 {
		live.GateHP, live.MaxGateHP = 200, 200
	}
	w.Events.Active = live
	for i := range w.Events.Events {
		if w.Events.Events[i].ID == def.ID {
			w.Events.Events[i].State = "ACTIVE"
		}
	}
	switch def.Kind {
	case "social", "education", "gather", "fog", "puzzle", "festival":
	default:
		w.spawnEventMobs(live)
	}
	view := live.view()
	return [][]byte{
		marshal(TypeWorldEvent, view),
		marshal(TypeChatMessage, ChatOut{Channel: "SYSTEM", From: "SYSTEM", Text: overlayEventAnnounce(def), System: true}),
	}
}

func (w *WorldState) spawnEventMobs(live *LiveEvent) {
	kinds := live.Def.Enemies
	if len(kinds) == 0 {
		kinds = []string{"forest_fang"}
	}
	byID := map[string]EnemyDef{}
	for _, d := range enemyCatalog {
		byID[d.ID] = d
	}
	n := live.scaledMobCount()
	spawned := 0
	max := worldCfg.MaxEnemiesCell
	if max < 1 {
		max = 6
	}
	for _, kind := range kinds {
		def, ok := byID[kind]
		if !ok {
			continue
		}
		for i := 0; i < n && spawned < max; i++ {
			e := spawnEnemy(def, float64((spawned-1)*4), 24.5)
			e.Elite = kind == "shadow_imp"
			w.enemies[e.ID] = e
			spawned++
		}
		if spawned >= max {
			break
		}
	}
}

func (w *WorldState) finishWorldEvent(live *LiveEvent, success bool) [][]byte {
	if success {
		live.State = "SUCCESS"
	} else {
		live.State = "EXPIRED"
	}
	for i := range w.Events.Events {
		if w.Events.Events[i].ID == live.Def.ID {
			w.Events.Events[i].State = live.State
		}
	}
	view := live.view()
	view.Success = success
	w.Events.Last = live
	w.Events.Active = nil
	w.Events.nextAt = time.Now().Add(3 * time.Minute)
	return [][]byte{marshal(TypeWorldEvent, view)}
}

func (w *WorldState) joinWorldEvent(p *Player) [][]byte {
	live := w.Events.Active
	if live == nil || (live.State != "ACTIVE" && live.State != "WAITING" && live.State != "FINAL") {
		return rejectFor(p.ID, TypeJoinWorldEvent, "inactive")
	}
	z := zoneAt(p.X, p.Z)
	if z.ID != live.Region && live.Region != "" {
		return rejectFor(p.ID, TypeJoinWorldEvent, "region")
	}
	if live.Score[p.ID] == 0 {
		live.Score[p.ID] = 1
	}
	w.noteActivity(p, "EVENT", live.Def.ID, 1)
	out := [][]byte{marshal(TypeWorldEvent, live.view())}
	if live.Def.ID == "darkness-rising" {
		p.credit("EVENT", "darkness-rising", 1)
		if strings.Contains(live.eventPhase(time.Now()), "PHASE 4") || live.State == "FINAL" {
			out = append(out, w.startStoryQuiz(p, "darkness-rising", "q-add-5-4")...)
		}
	}
	if live.Def.Kind == "education" && live.Def.QuestionID != "" {
		for i, q := range questionCatalog {
			if q.ID == live.Def.QuestionID {
				out = append(out, marshal(TypeEducationQuestion, questionOut(i)))
				break
			}
		}
	}
	if live.Def.Kind == "social" {
		p.credit("VISIT", "lost-child", 1)
	}
	return out
}

func (w *WorldState) addEventScore(p *Player, amount int) {
	if w.Events.Active == nil || amount <= 0 {
		return
	}
	w.Events.Active.Score[p.ID] += amount
}

func (w *WorldState) claimEventReward(p *Player, eventID string, clientScore int) [][]byte {
	_ = clientScore
	live := w.Events.Active
	if live == nil || live.Def.ID != eventID {
		if w.Events.Last != nil && (eventID == "" || w.Events.Last.Def.ID == eventID) {
			live = w.Events.Last
		}
	}
	if live == nil || (eventID != "" && live.Def.ID != eventID) {
		return rejectFor(p.ID, TypeClaimEventReward, "event")
	}
	if live.State != "COMPLETED" && live.State != "SUCCESS" && live.State != "ACTIVE" {
		return rejectFor(p.ID, TypeClaimEventReward, "state")
	}
	pid := live.Def.ID + ":" + live.SpawnID + ":" + p.ID
	log := p.ensureLog()
	if log.EventClaims == nil {
		log.EventClaims = map[string]bool{}
	}
	if live.Claimed[p.ID] || log.EventClaims[pid] {
		return rejectFor(p.ID, TypeClaimEventReward, "claimed")
	}
	score := live.Score[p.ID]
	if score < 1 {
		return rejectFor(p.ID, TypeClaimEventReward, "participation")
	}
	live.Claimed[p.ID] = true
	log.EventClaims[pid] = true
	r := live.Def.Rewards
	exp, coin, crystal := r.Exp, r.Coin, r.Crystal
	if score < 2 {
		exp /= 2
		coin /= 2
		if crystal > 0 {
			crystal = 0
		}
	}
	events := w.giveExp(p, exp)
	w.giveCurrency(p, coin, crystal)
	if r.EduToken > 0 && score >= 2 {
		log.EduToken += r.EduToken
	}
	p.credit("EVENT", live.Def.ID, 1)
	p.bumpRegionRep(live.Region, 4)
	w.guildContribute(p, 20, 12)
	log.Flags["event_"+live.Def.ID+"_done"] = true
	p.markDirty()
	w.persist(p)
	w.audit("questCompleted", p.ID, "event:"+pid)
	events = append(events, marshal(TypeEventReward, map[string]any{"eventId": live.Def.ID, "participationId": pid, "score": score, "exp": exp, "coin": coin, "afk": score < 2, "toId": p.ID}))
	events = append(events, w.afterTravelEventClaim(p, live.Def.ID)...)
	return events
}

func (e *LiveEvent) view() WorldEventView {
	now := time.Now()
	phase := e.eventPhase(now)
	startsIn, endsIn := 0, 0
	if !e.ActiveAt.IsZero() && now.Before(e.ActiveAt) {
		startsIn = int(e.ActiveAt.Sub(now).Seconds())
	}
	if !e.Until.IsZero() && now.Before(e.Until) {
		endsIn = int(e.Until.Sub(now).Seconds())
	}
	prog, need := eventProgress(e)
	return WorldEventView{
		ID: e.Def.ID, Name: e.Def.Name, Kind: e.Def.Kind, State: e.State, Phase: phase, Region: e.Region,
		Announce: overlayEventAnnounce(e.Def), AnnounceJV: e.Def.AnnounceJV, Objective: e.Def.Objective,
		Progress: prog, Need: need,
		GateHP: e.GateHP, MaxGateHP: e.MaxGateHP, Until: e.Until.UnixMilli(), StartsAt: e.ActiveAt.UnixMilli(),
		StartsIn: startsIn, EndsIn: endsIn, Participants: len(e.Score), X: e.Def.X, Z: e.Def.Z,
	}
}

func (w *WorldState) eventViewFor(p *Player) *WorldEventView {
	live := w.Events.Active
	if live == nil {
		return nil
	}
	if p != nil && zoneAt(p.X, p.Z).ID != live.Region && live.Region != "" {
		return nil
	}
	v := live.view()
	return &v
}
