package mmo

import (
	"math/rand"
	"time"
)

func dungeonHasMod(inst *DungeonInstance, id string) bool {
	if inst == nil {
		return false
	}
	for _, m := range inst.Modifiers {
		if m == id {
			return true
		}
	}
	return false
}

func (w *WorldState) applyHorizonStart(inst *DungeonInstance, def DungeonDef) {
	if inst == nil || def.ID != horizonCfg.DungeonID {
		return
	}
	inst.HorizonLevel = 1
	inst.Modifiers = pickHorizonModifiers(inst.ID)
	if dungeonHasMod(inst, "LIMITED_REVIVE") {
		for id := range inst.ReviveToken {
			inst.ReviveToken[id] = 0
		}
	}
}

func pickHorizonModifiers(seed string) []string {
	ids := make([]string, 0, len(horizonCfg.Modifiers))
	for _, m := range horizonCfg.Modifiers {
		ids = append(ids, m.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	max := horizonCfg.MaxModifiers
	if max < 1 {
		max = 2
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(len(seed))))
	for attempt := 0; attempt < 12; attempt++ {
		n := 1 + r.Intn(max)
		if n > len(ids) {
			n = len(ids)
		}
		perm := r.Perm(len(ids))
		out := make([]string, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, ids[perm[i]])
		}
		if validHorizonCombo(out) {
			return out
		}
	}
	return []string{"FAST_ENEMIES"}
}

func validHorizonCombo(ids []string) bool {
	set := map[string]bool{}
	for _, id := range ids {
		set[id] = true
	}
	for _, pair := range horizonCfg.InvalidPairs {
		if len(pair) < 2 {
			continue
		}
		if set[pair[0]] && set[pair[1]] {
			return false
		}
	}
	return true
}

func applyHorizonEnemy(inst *DungeonInstance, e *Enemy) {
	if inst == nil || e == nil {
		return
	}
	if dungeonHasMod(inst, "FAST_ENEMIES") {
		e.Def.Speed *= 1.35
		if e.Def.AttackCooldown > 0.4 {
			e.Def.AttackCooldown *= 0.85
		}
	}
	if e.IsBoss && dungeonHasMod(inst, "STRONG_BOSS") {
		e.Def.Attack = int(float64(e.Def.Attack) * 1.15)
		if e.Def.AttackCooldown > 0.35 {
			e.Def.AttackCooldown *= 0.7
		}
	}
}

func horizonHealAmount(inst *DungeonInstance, maxHP int) int {
	if dungeonHasMod(inst, "LOW_HEAL") {
		return maxHP * 4 / 10
	}
	return maxHP
}

func (w *WorldState) finishHorizon(inst *DungeonInstance, p *Player) [][]byte {
	if inst == nil || p == nil || inst.DefID != horizonCfg.DungeonID {
		return nil
	}
	elapsed := time.Since(inst.StartedAt).Seconds()
	deaths := 0
	for _, n := range inst.Deaths {
		deaths += n
	}
	timeScore := horizonCfg.ScoreTimeWeight * (1200 - int(elapsed))
	if timeScore < 0 {
		timeScore = 0
	}
	modBonus := 80 * len(inst.Modifiers)
	obj := 200
	if inst.ObjNeed > 0 && inst.ObjProgress >= inst.ObjNeed {
		obj = 320
	}
	score := timeScore + modBonus + obj - deaths*horizonCfg.ScoreDeathPen
	if score < 1 {
		score = 1
	}
	inst.HorizonScore = score
	log := p.ensureLog()
	week := raidWeekKey()
	if log.HorizonWeek != week {
		log.HorizonWeek = week
		log.HorizonBest = 0
	}
	if score > log.HorizonBest {
		log.HorizonBest = score
	}
	w.upsertHorizonBoard(p, score, log.HorizonBest)
	if log.HorizonClaimWeek != week {
		log.HorizonClaimWeek = week
		log.HorizonFragments++
	}
	p.markDirty()
	w.audit("horizonScore", p.ID, inst.ID)
	return nil
}

func (w *WorldState) upsertHorizonBoard(p *Player, score, best int) {
	week := raidWeekKey()
	if w.HorizonLBWeek != week {
		if len(w.HorizonLB) > 0 {
			w.HorizonHistory = append(w.HorizonHistory, w.HorizonLB...)
			if len(w.HorizonHistory) > 80 {
				w.HorizonHistory = w.HorizonHistory[len(w.HorizonHistory)-80:]
			}
		}
		w.HorizonLB = nil
		w.HorizonLBWeek = week
	}
	found := false
	for i := range w.HorizonLB {
		if w.HorizonLB[i].PlayerID == p.ID {
			if score > w.HorizonLB[i].Score {
				w.HorizonLB[i].Score = score
				w.HorizonLB[i].Level = p.Level
				w.HorizonLB[i].Name = p.Name
			}
			found = true
			break
		}
	}
	if !found {
		w.HorizonLB = append(w.HorizonLB, HorizonScore{
			PlayerID: p.ID, Name: p.Name, Score: best, Level: p.Level, Week: week,
		})
	}
}

func (w *WorldState) horizonView(p *Player) map[string]any {
	log := p.ensureLog()
	rows := make([]HorizonScore, 0, len(w.HorizonLB))
	for _, r := range w.HorizonLB {
		rows = append(rows, r)
	}
	return map[string]any{
		"dungeonId": horizonCfg.DungeonID, "maxLevel": horizonCfg.MaxLevel,
		"best": log.HorizonBest, "week": raidWeekKey(), "board": rows,
		"history": w.HorizonHistory, "modifiers": horizonCfg.Modifiers,
	}
}
