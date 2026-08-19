package mmo

// SpawnServer is a logical slice of WorldState (monolithic deployment).
func (w *WorldState) canRespawnInCell(x, z float64) bool {
	max := worldCfg.MaxEnemiesCell
	if max < 1 {
		max = 6
	}
	key := cellKey(x, z)
	n := 0
	for _, e := range w.enemies {
		if !e.Alive {
			continue
		}
		if cellKey(e.X, e.Z) == key {
			n++
			if n >= max {
				return false
			}
		}
	}
	return true
}
