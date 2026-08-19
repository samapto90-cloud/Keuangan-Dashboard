package mmo

import "math"

// World +Z is north, +X is east. All navigation UI must use this mapping.
func CardinalFromDelta(dx, dz float64) string {
	if dx == 0 && dz == 0 {
		return "N"
	}
	deg := math.Atan2(dx, dz) * 180 / math.Pi
	if deg < 0 {
		deg += 360
	}
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	i := int(math.Round(deg/45.0)) % 8
	if i < 0 {
		i += 8
	}
	return dirs[i]
}

func CardinalID(cardinal string) string {
	switch cardinal {
	case "N":
		return "utara"
	case "NE":
		return "timur laut"
	case "E":
		return "timur"
	case "SE":
		return "tenggara"
	case "S":
		return "selatan"
	case "SW":
		return "barat daya"
	case "W":
		return "barat"
	case "NW":
		return "barat laut"
	default:
		return "utara"
	}
}

func CardinalJV(cardinal string) string {
	switch cardinal {
	case "N":
		return "ngalor"
	case "NE":
		return "ngetan-lor"
	case "E":
		return "ngetan"
	case "SE":
		return "ngetan-kidul"
	case "S":
		return "ngidul"
	case "SW":
		return "ngulon-kidul"
	case "W":
		return "ngulon"
	case "NW":
		return "ngulon-lor"
	default:
		return "ngalor"
	}
}

func HeightMark(dy float64) string {
	if dy > 1.4 {
		return "up"
	}
	if dy < -1.4 {
		return "down"
	}
	return "level"
}

// WrongWayReady is true when the player has stayed 100m+ past their closest
// approach for at least 15 seconds. No teleport, no forced movement.
func WrongWayReady(awayMeters, awaySeconds float64) bool {
	return awayMeters > 100 && awaySeconds >= 15
}

func (p *Player) journeyHint() (jv, id, landmark, cardinal string) {
	j := p.journeyView()
	return j.HintJv, j.HintId, j.Landmark, j.Cardinal
}

func applyJourneyGuidance(p *Player, j *JourneyView) {
	dx := j.NavX - p.X
	dz := j.NavZ - p.Z
	j.Cardinal = CardinalFromDelta(dx, dz)
	if j.Landmark == "" {
		j.Landmark = landmarkNameForNav(j.NavX, j.NavZ, j.NextRegion)
	}
	j.HintJv = "Le, yen arep menyang " + j.Landmark + ", mlakua " + CardinalJV(j.Cardinal) + "."
	j.HintId = "Kalau ingin menuju " + j.Landmark + ", berjalan ke " + CardinalID(j.Cardinal) + "."
}
