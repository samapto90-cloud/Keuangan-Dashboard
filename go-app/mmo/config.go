package mmo

import "time"

const (
	ServerTickRate   = 20
	WorldID          = "world-01"
	ChannelID        = "channel-01"
	MaxPlayers       = 100
	WorldLimit       = 200.0
	GateMaxZ         = 20.8
	ChannelCount     = 3
	ChannelMaxPop    = 40
	WorldBossMaxPop  = 20
	DayCycleSec      = 400.0
	OpenWorldPvP     = false
	WalkSpeed        = 3.0
	RunSpeed         = 6.0
	Acceleration     = 28.0
	Deceleration     = 22.0
	Gravity          = 22.0
	JumpForce        = 7.2
	MaxInputRate     = 20
	HeartbeatTimeout = 15 * time.Second
	InputMinInterval = time.Second / 24
)

var spawnPoints = [][3]float64{
	{0, 0, 6},
	{2.4, 0, 7.2},
	{-2.4, 0, 7.2},
}
