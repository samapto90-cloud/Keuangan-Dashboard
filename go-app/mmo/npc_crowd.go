package mmo

import (
	"math"
	"os"
)

const (
	npcPersonalSpace   = 0.78
	npcAvoidanceRadius = 1.15
	npcMinSpawnDist    = 2.0
)

var placeSlots = map[string][]npcVec{
	"market": {
		{8.2, 2.2}, {10.2, 2.8}, {9.8, 4.6}, {11.4, 3.8}, {7.6, 3.4}, {10.8, 1.6}, {8.8, 4.2}, {12.0, 2.4},
	},
	"square": {
		{-2.2, 3.8}, {2.0, 4.2}, {-1.0, 6.2}, {1.4, 5.6}, {-3.0, 5.0}, {3.2, 3.4},
		{0.8, 7.0}, {2.8, 6.8}, {-2.8, 4.8}, {1.0, 3.2}, {3.6, 5.2}, {-1.6, 7.2},
	},
	"school": {
		{-8.0, 4.4}, {-7.2, 3.6}, {-8.6, 5.0}, {-6.8, 4.2}, {-9.0, 4.8}, {-7.6, 5.4},
	},
	"clinic": {
		{5.4, 6.2}, {8.8, 6.0}, {6.2, 6.6}, {9.2, 6.4},
	},
	"home": {
		{-8.6, 6.6}, {-9.2, 7.0}, {-8.2, 7.4}, {-9.8, 6.4},
	},
	"hall": {
		{-4.4, 2.8}, {-4.0, 3.4}, {-8.8, 3.6}, {-4.8, 4.0},
	},
	"gate": {
		{-1.4, 19.0}, {1.4, 19.0}, {0.0, 18.2}, {-2.6, 18.8}, {2.6, 18.8},
	},
	"temple": {
		{-1.2, 3.0}, {1.2, 3.0}, {0.0, 2.2}, {-2.0, 3.6}, {2.0, 3.6},
	},
}

var placeCapacity = map[string]int{
	"market": 8, "square": 12, "school": 6, "clinic": 4, "gate": 5, "home": 4, "hall": 4, "temple": 5,
}

var guardPatrol = []npcVec{
	{0, 19.6}, {-3.0, 18.4}, {3.0, 18.4}, {0, 17.6},
}

var crowdNPCExtra []NPCDef

func init() {
	if os.Getenv("CAHAYA_CROWD_TEST") == "1" {
		crowdNPCExtra = buildCrowdTestNPCs(27)
	}
}

func buildCrowdTestNPCs(n int) []NPCDef {
	out := make([]NPCDef, 0, n)
	zones := []struct {
		cx, cz, r float64
	}{
		{0, 5, 6}, {9, 3, 4}, {-8, 3, 4}, {0, 10, 5}, {-6, 8, 4},
	}
	for i := 0; i < n; i++ {
		z := zones[i%len(zones)]
		ang := float64(i)*2.399 + float64(z.cx)
		dist := 1.8 + float64(i%5)*1.4
		x := z.cx + math.Cos(ang)*dist
		cz := z.cz + math.Sin(ang)*dist*0.7
		if !isWalkableXZ(x, cz, NPCRadius) {
			x, cz = z.cx, z.cz
		}
		out = append(out, NPCDef{
			ID:   "crowd_" + pad2(i+1),
			Name: "Villager",
			Role: "Villager",
			Type: "VILLAGER",
			X:    x, Y: 0, Z: cz,
			Yaw: float64(i) * 0.7,
			Schedule: map[string]string{
				"day": "village", "morning": "village", "afternoon": "village",
				"evening": "village", "night": "home",
			},
			InteractionRange: 2.2,
		})
	}
	return out
}

func activeNPCList() []NPCDef {
	if len(crowdNPCExtra) == 0 {
		return npcCatalog
	}
	out := make([]NPCDef, len(npcCatalog)+len(crowdNPCExtra))
	copy(out, npcCatalog)
	copy(out[len(npcCatalog):], crowdNPCExtra)
	return out
}

func npcPriority(t string) int {
	switch t {
	case "QUEST_BOARD", "ELDER", "TEACHER", "SCHOLAR", "HEALER", "MERCHANT", "GUILD":
		return 3
	case "GUARD":
		return 2
	default:
		return 1
	}
}

func npcBodyRadius(n NPCDef) float64 {
	if n.Type == "CHILD" {
		return NPCRadiusChild
	}
	return NPCRadius
}

func npcWalkSpeed(n NPCDef, flee bool) float64 {
	if flee {
		return 3.0
	}
	switch n.Type {
	case "CHILD":
		return 1.15
	case "ELDER":
		return 0.95
	case "GUARD":
		return 1.1
	case "MERCHANT", "TEACHER", "HEALER", "SCHOLAR":
		return 0.85
	case "VILLAGER":
		return 1.25
	default:
		return 1.35
	}
}

func npcWanderRadius(n NPCDef) float64 {
	switch n.Type {
	case "CHILD":
		return 6
	case "MERCHANT", "TEACHER", "HEALER":
		return 2.5
	case "GUARD":
		return 0
	case "VILLAGER", "GUIDE", "TRAVELER":
		return 14
	default:
		return 8
	}
}

func npcStationary(n NPCDef) bool {
	switch n.Type {
	case "QUEST_BOARD", "ELDER", "HEALER", "SCHOLAR", "GUILD", "ARENA":
		return true
	case "MERCHANT":
		return true
	case "TEACHER":
		return true
	case "MENTOR", "GUIDE":
		return true
	case "CHILD":
		return true
	default:
		return false
	}
}

func npcHash(id string) int {
	h := 0
	for i := 0; i < len(id); i++ {
		h = h*31 + int(id[i])
	}
	if h < 0 {
		h = -h
	}
	return h
}

func npcStartDelay(id string) float64 {
	return float64(npcHash(id)%25) + 3
}

func npcSlotIndex(place, id string, cap int) int {
	if cap < 1 {
		return 0
	}
	return npcHash(id+"@"+place) % cap
}

func npcDestWithSlot(n NPCDef, place string, t WorldTimeSystem) (float64, float64) {
	if slots, ok := placeSlots[place]; ok && len(slots) > 0 {
		cap := len(slots)
		if c, ok := placeCapacity[place]; ok && c < cap {
			cap = c
		}
		idx := npcSlotIndex(place, n.ID, cap)
		s := slots[idx]
		if isWalkableXZ(s.X, s.Z, NPCRadius) {
			return s.X, s.Z
		}
		return pickWanderPoint(0, 5.2, 4.5, n.ID+place)
	}
	if v, ok := npcPlaces[place]; ok {
		if place == "village" || place == "workplace" {
			return n.X, n.Z
		}
		return v.X, v.Z
	}
	return n.X, n.Z
}

func pickWanderPoint(homeX, homeZ, radius float64, id string) (float64, float64) {
	h := npcHash(id + "w")
	for i := 0; i < 12; i++ {
		a := float64(h%628)/100.0 + float64(i)*0.52
		d := radius * (0.35 + float64((h+i*17)%70)/100.0)
		x := homeX + math.Cos(a)*d
		z := homeZ + math.Sin(a)*d*0.85
		if isWalkableXZ(x, z, NPCRadius) {
			return x, z
		}
	}
	if isWalkableXZ(homeX, homeZ, NPCRadius) {
		return homeX, homeZ
	}
	return 0, 5.2
}

func guardDest(n NPCDef, t float64) (float64, float64) {
	if len(guardPatrol) == 0 {
		return n.X, n.Z
	}
	step := 8.0
	idx := int(t/step) % len(guardPatrol)
	idx = (idx + npcHash(n.ID)%len(guardPatrol)) % len(guardPatrol)
	p := guardPatrol[idx]
	return p.X, p.Z
}

func crowdCountNear(x, z, radius float64, positions map[string]npcLive) int {
	n := 0
	for id, v := range positions {
		if id == "" {
			continue
		}
		if math.Hypot(v.X-x, v.Z-z) <= radius {
			n++
		}
	}
	return n
}

func pickAlternateDest(n NPCDef, cur npcLive, t float64, positions map[string]npcLive) (float64, float64) {
	r := npcWanderRadius(n)
	if r < 1 {
		r = 6
	}
	x, z := pickWanderPoint(n.X, n.Z, r, n.ID+"alt"+pad2(int(t)))
	if crowdCountNear(x, z, 2.5, positions) > 3 {
		x, z = pickWanderPoint(n.X, n.Z, r*0.6, n.ID+"alt2")
	}
	return x, z
}

func (w *WorldState) npcDestination(n NPCDef) (float64, float64) {
	cur, has := w.npcPos[n.ID]
	eventFlee := w.Events.Active != nil && (w.Events.Active.State == "ACTIVE" || w.Events.Active.State == "FINAL")
	if eventFlee && n.Type != "GUARD" && zoneAt(n.X, n.Z).ID == "village" {
		if n.Type == "CHILD" || n.Type == "TEACHER" || n.Type == "MERCHANT" {
			h := npcPlaces["home"]
			return h.X, h.Z
		}
		s := npcPlaces["square"]
		return s.X, s.Z
	}
	if has && cur.UsingWander {
		if math.Hypot(cur.WanderX-cur.X, cur.WanderZ-cur.Z) < 0.35 {
			cur.UsingWander = false
			w.npcPos[n.ID] = cur
		} else {
			return cur.WanderX, cur.WanderZ
		}
	}
	if n.Type == "GUARD" {
		return guardDest(n, w.Time.Seconds())
	}
	place := ""
	if n.Schedule != nil {
		place = n.Schedule[w.Time.Slot()]
		if place == "" {
			switch w.Time.Phase {
			case "NIGHT":
				place = n.Schedule["night"]
			case "EVENING":
				place = n.Schedule["evening"]
			default:
				place = n.Schedule["day"]
			}
		}
	}
	if n.Type == "MERCHANT" && (w.Time.ClockHour() < 8 || w.Time.ClockHour() >= 20) {
		if home, ok := n.Schedule["night"]; ok && home != "" {
			place = home
		} else {
			place = "home"
		}
	}
	if place == "" {
		return n.X, n.Z
	}
	dx, dz := npcDestWithSlot(n, place, w.Time)
	if npcStationary(n) {
		ox, oz := npcParkOffset(n.ID)
		return dx + ox*0.35, dz + oz*0.35
	}
	// Wandering types near home when at destination
	if has {
		distHome := math.Hypot(cur.X-n.X, cur.Z-n.Z)
		atDest := math.Hypot(dx-cur.X, dz-cur.Z) < 0.8
		wanderR := npcWanderRadius(n)
		if wanderR > 1 && atDest && w.Time.Seconds() > cur.WanderReady && distHome < wanderR+2 {
			wx, wz := pickWanderPoint(n.X, n.Z, wanderR, n.ID+pad2(int(w.Time.Seconds()/30)))
			cur.WanderX, cur.WanderZ = wx, wz
			cur.UsingWander = true
			cur.WanderReady = w.Time.Seconds() + 12 + float64(npcHash(n.ID)%28)
			w.npcPos[n.ID] = cur
			return wx, wz
		}
	}
	ox, oz := npcParkOffset(n.ID)
	return dx + ox, dz + oz
}

func (w *WorldState) villagePlayers() [][2]float64 {
	out := make([][2]float64, 0, len(w.players))
	for _, p := range w.players {
		if !p.Connected || p.InstanceID != "" {
			continue
		}
		if zoneAt(p.X, p.Z).ID != "village" {
			continue
		}
		out = append(out, [2]float64{p.X, p.Z})
	}
	return out
}

func resolvePlayerNPCPush(px, pz, pRad float64, npcPos map[string]npcLive, catalog []NPCDef) (float64, float64) {
	byID := map[string]NPCDef{}
	for _, n := range catalog {
		byID[n.ID] = n
	}
	for id, live := range npcPos {
		n, ok := byID[id]
		if !ok {
			continue
		}
		nRad := npcBodyRadius(n)
		dx, dz := px-live.X, pz-live.Z
		d := math.Hypot(dx, dz)
		minD := pRad + nRad + 0.05
		if d < 0.001 || d >= minD {
			continue
		}
		push := (minD - d) / d * 0.55
		px += dx * push
		pz += dz * push
	}
	return px, pz
}

// WorldTimeSystem.Seconds returns elapsed in-game seconds for wander timers.
func (t WorldTimeSystem) Seconds() float64 {
	if t.ClockMin < 0 {
		return 0
	}
	return float64(t.ClockMin) * 60
}
