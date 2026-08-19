package mmo

import "math/rand"

func calcDamage(base, strength, skillPower, equipment, defense float64) (dealt int, crit bool) {
	return calcDamageChance(base, strength, skillPower, equipment, defense, CritChance)
}

func calcDamageChance(base, strength, skillPower, equipment, defense, critChance float64) (dealt int, crit bool) {
	if skillPower <= 0 {
		skillPower = 1
	}
	if equipment <= 0 {
		equipment = 1
	}
	if critChance <= 0 {
		critChance = CritChance
	}
	raw := base * (1 + strength/10) * skillPower * equipment
	mit := 1 - defense/(defense+50)
	if mit < 0.2 {
		mit = 0.2
	}
	raw *= mit
	crit = rand.Float64() < critChance
	if crit {
		raw *= CritMultiplier
	}
	dealt = int(raw + 0.5)
	if dealt < MinDamage {
		dealt = MinDamage
	}
	if dealt > MaxDamage {
		dealt = MaxDamage
	}
	return dealt, crit
}
