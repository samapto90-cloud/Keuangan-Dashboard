package mmo

// Air combat architecture for Phase 7 — jump / air attack / air combo / air dash.
// Free flight is intentionally not implemented.

func (p *Player) airborne() bool {
	return p.Y > 0.35 && !p.Grounded
}

func (p *Player) canAirAttack() bool {
	return p.alive() && p.airborne()
}

func (p *Player) canAirDash() bool {
	return p.canAirAttack() && (p.hasSkill("air_dash") || p.hasSkill("flash_movement"))
}

func (p *Player) airComboCap() int {
	if p.hasSkill("flash_movement") {
		return AirComboMaxHits
	}
	if AirComboMaxHits > 2 {
		return AirComboMaxHits
	}
	return 2
}
