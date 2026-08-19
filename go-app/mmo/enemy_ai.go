package mmo

import (
	"math"
	"time"
)

func (w *WorldState) tickEnemies(dt float64, now time.Time) [][]byte {
	return w.tickEnemyMap(w.enemies, func(p *Player) bool { return p != nil && p.InstanceID == "" }, true, dt, now)
}

func (w *WorldState) tickEnemyMap(enemies map[string]*Enemy, allow func(*Player) bool, respawn bool, dt float64, now time.Time) [][]byte {
	var events [][]byte
	for _, e := range enemies {
		if e.IsBoss {
			continue
		}
		if !e.Alive {
			if respawn && !e.NoRespawn && now.After(e.DeadUntil) && w.canRespawnInCell(e.SX, e.SZ) {
				e.X, e.Y, e.Z = e.SX, e.SY, e.SZ
				e.VX, e.VZ = 0, 0
				e.HP = e.MaxHP
				e.Alive = true
				e.State = "IDLE"
				e.TargetID = ""
				e.LastHitBy = ""
				events = append(events, marshal(TypeEnemySpawn, e.Snap()))
			}
			continue
		}
		e.VX *= math.Exp(-4 * dt)
		e.VZ *= math.Exp(-4 * dt)
		if e.Def.Behavior == "dummy" || e.Def.Attack <= 0 || e.Def.AggroRange <= 0 {
			e.State = "IDLE"
			e.TargetID = ""
			e.VX, e.VZ = 0, 0
			continue
		}
		home := math.Hypot(e.X-e.SX, e.Z-e.SZ)
		target := w.players[e.TargetID]
		if target == nil || !target.Connected || !target.alive() || !allow(target) || target.inOpenWorldSafeZone() {
			e.TargetID = ""
			target = nil
		}
		if target == nil {
			e.TargetID = w.closestPlayerFiltered(e, allow)
			target = w.players[e.TargetID]
		}
		if home > e.Def.Leash {
			e.TargetID = ""
			target = nil
			e.State = "RETURN"
		}
		switch {
		case e.State == "RETURN" || (target == nil && home > 0.4):
			e.State = "RETURN"
			w.seek(e, e.SX, e.SZ, dt)
			if math.Hypot(e.X-e.SX, e.Z-e.SZ) < 0.35 {
				e.State = "IDLE"
				e.VX, e.VZ = 0, 0
			}
		case target != nil && e.Dist(target.X, target.Z) <= e.Def.AttackRange:
			e.State = "ATTACK"
			e.Yaw = math.Atan2(target.X-e.X, target.Z-e.Z)
			if !now.Before(e.NextAttack) {
				e.NextAttack = now.Add(time.Duration(e.Def.AttackCooldown * float64(time.Second)))
				events = append(events, w.hitPlayer(e.ID, e.Def.Name, target, float64(e.Def.Attack), 0, 2.2, "enemy")...)
				e.State = "COOLDOWN"
			}
		case target != nil:
			e.State = "CHASE"
			w.seek(e, target.X, target.Z, dt)
		default:
			e.State = "IDLE"
		}
		e.X += e.VX * dt
		e.Z += e.VZ * dt
		limit := WorldLimit
		if e.InstanceID != "" {
			limit = DungeonLimit
		}
		e.X = clamp(e.X, -limit, limit)
		e.Z = clamp(e.Z, -limit, limit)
	}
	return events
}

func (w *WorldState) closestPlayer(e *Enemy) string {
	return w.closestPlayerFiltered(e, func(p *Player) bool { return p != nil && p.InstanceID == e.InstanceID })
}

func (w *WorldState) closestPlayerFiltered(e *Enemy, allow func(*Player) bool) string {
	bestID := ""
	best := e.Def.AggroRange
	for _, p := range w.players {
		if !p.Connected || !p.alive() || !allow(p) {
			continue
		}
		if p.inOpenWorldSafeZone() {
			continue
		}
		d := e.Dist(p.X, p.Z)
		if d <= best {
			best = d
			bestID = p.ID
		}
	}
	return bestID
}

func (w *WorldState) seek(e *Enemy, x, z, dt float64) {
	dx := x - e.X
	dz := z - e.Z
	n := math.Hypot(dx, dz)
	if n < 0.05 {
		return
	}
	e.Yaw = math.Atan2(dx, dz)
	e.VX = dx / n * e.Def.Speed
	e.VZ = dz / n * e.Def.Speed
	_ = dt
}
