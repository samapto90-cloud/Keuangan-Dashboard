package mmo

import "math"

// Collision layers (logical — used for filtering).
const (
	LayerPlayer     = "PLAYER"
	LayerNPC        = "NPC"
	LayerEnemy      = "ENEMY"
	LayerBuilding   = "BUILDING"
	LayerWorld      = "WORLD"
	LayerObstacle   = "OBSTACLE"
	LayerProjectile = "PROJECTILE"
	LayerTrigger    = "TRIGGER"
)

const (
	PlayerRadius  = 0.38
	NPCRadius     = 0.32
	NPCRadiusChild = 0.26
)

type worldAABB struct {
	X, Z, W, D float64
	Layer        string
}

type worldCircle struct {
	X, Z, R  float64
	Layer    string
	Walkable bool // bridge segments
}

// villageObstacles mirrors Dawn Village props in cahaya-game/src/world/Environment.ts.
var villageObstacles = []worldAABB{
	{-11.5, 8.5, 3.4, 2.8, LayerBuilding},
	{-13.2, 5.2, 3.4, 2.8, LayerBuilding},
	{-10.4, 11.4, 3.4, 2.8, LayerBuilding},
	{-6.6, 0.6, 3.8, 3.0, LayerBuilding},
	{7.4, 8.2, 3.4, 2.8, LayerBuilding},
	{11.2, 4.8, 2.4, 1.6, LayerBuilding},
	{-9.2, 1.1, 1.8, 0.9, LayerObstacle},
	{-10.6, 2.6, 1.5, 0.4, LayerObstacle},
	{-8.1, 0.4, 1.9, 0.2, LayerObstacle},
	{-12.0, 14.5, 14.0, 3.4, LayerWorld}, // water — non-walkable
}

var villageCircles = []worldCircle{
	{-1.1, 4.6, 0.85, LayerObstacle, false},
	{-12.0, 14.5, 1.3, LayerWorld, true}, // bridge walkable carve-out
}

var villageTrees = func() [][3]float64 {
	out := make([][3]float64, 0, 32)
	for i := 0; i < 22; i++ {
		ang := float64(i) / 22 * 2 * math.Pi
		dist := 16 + float64(i%5)*2.2
		out = append(out, [3]float64{math.Cos(ang) * dist, math.Sin(ang)*dist*0.55 + 2, 0.9})
	}
	for i := 0; i < 10; i++ {
		out = append(out, [3]float64{(float64(i) - 4.5) * 2.1, 27 + float64(i%3)*1.8, 0.85})
	}
	return out
}()

var villageRocks = [][3]float64{
	{-7.0, 12.5, 0.38}, {-5.4, 13.5, 0.42}, {-3.8, 12.5, 0.35}, {-2.2, 13.5, 0.4},
	{-0.6, 12.5, 0.36}, {1.0, 13.5, 0.41}, {2.6, 12.5, 0.39}, {4.2, 13.5, 0.37},
	{5.8, 12.5, 0.4}, {7.4, 13.5, 0.38},
}

var villageFenceZ = 20.4

func insideAABB(x, z, pad float64, b worldAABB) bool {
	hw := b.W/2 + pad
	hd := b.D/2 + pad
	return math.Abs(x-b.X) < hw && math.Abs(z-b.Z) < hd
}

func pushOutAABB(x, z, pad float64, b worldAABB) (float64, float64) {
	hwx := b.W/2 + pad
	hdz := b.D/2 + pad
	dx := x - b.X
	dz := z - b.Z
	if math.Abs(dx) >= hwx || math.Abs(dz) >= hdz {
		return x, z
	}
	ox := hwx - math.Abs(dx)
	oz := hdz - math.Abs(dz)
	if ox < oz {
		if dx >= 0 {
			x = b.X + hwx
		} else {
			x = b.X - hwx
		}
	} else {
		if dz >= 0 {
			z = b.Z + hdz
		} else {
			z = b.Z - hdz
		}
	}
	return x, z
}

func onBridge(x, z float64) bool {
	return math.Abs(x+12) < 1.3 && math.Abs(z-14.5) < 2.2
}

func hitVillageFence(x, z, radius float64) bool {
	if math.Abs(z-villageFenceZ) > radius+0.12 {
		return false
	}
	if math.Abs(x) < 1.45 {
		return false
	}
	if math.Abs(x) > 6.2 {
		return false
	}
	return true
}

func pushVillageFence(x, z, radius float64) (float64, float64) {
	if !hitVillageFence(x, z, radius) {
		return x, z
	}
	if z >= villageFenceZ {
		z = villageFenceZ + radius + 0.12
	} else {
		z = villageFenceZ - radius - 0.12
	}
	return x, z
}

func isWalkableXZ(x, z, radius float64) bool {
	if onBridge(x, z) {
		return true
	}
	for _, b := range villageObstacles {
		if b.Layer == LayerWorld && insideAABB(x, z, radius, b) {
			return false
		}
		if b.Layer != LayerWorld && insideAABB(x, z, radius, b) {
			return false
		}
	}
	for _, c := range villageCircles {
		if c.Walkable {
			continue
		}
		if math.Hypot(x-c.X, z-c.Z) < c.R+radius {
			return false
		}
	}
	for _, t := range villageTrees {
		if math.Hypot(x-t[0], z-t[1]) < t[2]+radius {
			return false
		}
	}
	for _, r := range villageRocks {
		if math.Hypot(x-r[0], z-r[1]) < r[2]+radius {
			return false
		}
	}
	if hitVillageFence(x, z, radius) {
		return false
	}
	return true
}

func resolveWorldXZ(x, z, radius float64) (float64, float64) {
	for pass := 0; pass < 4; pass++ {
		for _, b := range villageObstacles {
			if b.Layer == LayerWorld {
				if onBridge(x, z) {
					continue
				}
			}
			x, z = pushOutAABB(x, z, radius, b)
		}
		for _, c := range villageCircles {
			if c.Walkable {
				continue
			}
			d := math.Hypot(x-c.X, z-c.Z)
			minD := c.R + radius
			if d > 0.001 && d < minD {
				push := (minD - d) / d
				x += (x - c.X) * push
				z += (z - c.Z) * push
			}
		}
		for _, t := range villageTrees {
			d := math.Hypot(x-t[0], z-t[1])
			minD := t[2] + radius
			if d > 0.001 && d < minD {
				push := (minD - d) / d
				x += (x - t[0]) * push
				z += (z - t[1]) * push
			}
		}
		for _, r := range villageRocks {
			d := math.Hypot(x-r[0], z-r[1])
			minD := r[2] + radius
			if d > 0.001 && d < minD {
				push := (minD - d) / d
				x += (x - r[0]) * push
				z += (z - r[1]) * push
			}
		}
		x, z = pushVillageFence(x, z, radius)
	}
	return x, z
}

func steerClearOfObstacles(x, z, dirX, dirZ, radius float64) (float64, float64) {
	probe := 1.2
	nx, nz := x+dirX*probe, z+dirZ*probe
	if isWalkableXZ(nx, nz, radius) {
		return dirX, dirZ
	}
	bestX, bestZ := dirX, dirZ
	bestScore := -1.0
	for _, ang := range []float64{-0.85, -0.55, 0.55, 0.85, -1.2, 1.2, -1.57, 1.57} {
		c, s := math.Cos(ang), math.Sin(ang)
		tx := dirX*c - dirZ*s
		tz := dirX*s + dirZ*c
		l := math.Hypot(tx, tz)
		if l < 0.001 {
			continue
		}
		tx /= l
		tz /= l
		px, pz := x+tx*probe, z+tz*probe
		if !isWalkableXZ(px, pz, radius) {
			continue
		}
		score := tx*dirX + tz*dirZ
		if score > bestScore {
			bestScore = score
			bestX, bestZ = tx, tz
		}
	}
	return bestX, bestZ
}

func softSeparation(x, z, radius float64, others [][3]float64) (float64, float64) {
	sx, sz := 0.0, 0.0
	for _, o := range others {
		dx, dz := x-o[0], z-o[1]
		d := math.Hypot(dx, dz)
		minD := radius + o[2]
		if d < 0.001 || d >= minD {
			continue
		}
		push := (minD - d) / d
		sx += dx * push
		sz += dz * push
	}
	return sx, sz
}

// avoidBuildings kept for compatibility — delegates to resolveWorldXZ.
func avoidBuildings(x, z float64) (float64, float64) {
	return resolveWorldXZ(x, z, NPCRadius)
}
