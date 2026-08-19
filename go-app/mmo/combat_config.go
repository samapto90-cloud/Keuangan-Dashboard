package mmo

import "time"

const (
	HpRegenPctPerSec   = 0.01
	EnergyRegenPerSec  = 8.0
	StaminaRegenPerSec = 15.0
	StaminaSprintDrain = 5.0
	StaminaMax         = 100.0
	EnergyMax          = 100
	BaseMaxHP          = 100
	BaseStrength       = 5
	BaseDefense        = 4
	EquipmentBonus     = 1.0
	CritChance         = 0.05
	CritMultiplier     = 1.5
	MinDamage          = 1
	MaxDamage          = 2500
	ComboWindow        = 800 * time.Millisecond
	ComboMaxStep       = 4
	PunchRange         = 1.55
	KickRange          = 1.8
	BasicAttackRange   = 1.6
	RangeSlack         = 0.55
	DodgeStamina       = 20.0
	DodgeDuration      = 400 * time.Millisecond
	DodgeIFrame        = 250 * time.Millisecond
	DodgeCooldown      = 550 * time.Millisecond
	DodgeSpeed         = 14.0
	AttackCooldown     = 180 * time.Millisecond
	KickCooldown       = 220 * time.Millisecond
	HitStun            = 180 * time.Millisecond
	CombatIdleAfter    = 4 * time.Second
	RespawnDelay       = 5 * time.Second
	EnemyRespawnDelay  = 12 * time.Second
	TimestampMaxAge    = 10 * time.Second
	KickKnockback      = 5.0
	PunchKnockback     = 2.4
)

var punchDamage = []int{10, 12, 14, 20}
var kickDamage = []int{15, 18, 22, 28}
