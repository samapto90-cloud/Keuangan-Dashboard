package mmo

import "time"

type WorldTimeSystem struct {
	Phase          string
	Weather        string
	cycleAt        time.Time
	ClockMin       int
	clockAcc       float64
	lastTick       time.Time
	GameMinPerReal float64
	RegionWeather  map[string]string
}

func NewWorldTimeSystem() WorldTimeSystem {
	rate := worldCfg.GameMinutesPerRealMinute
	if rate <= 0 {
		rate = 5
	}
	return WorldTimeSystem{
		Phase: "DAY", Weather: "CLEAR", cycleAt: time.Now(), lastTick: time.Now(),
		ClockMin: 9 * 60, GameMinPerReal: rate, RegionWeather: map[string]string{},
	}
}

func (t *WorldTimeSystem) Tick(now time.Time) bool {
	if t.lastTick.IsZero() {
		t.lastTick = now
	}
	if t.ClockMin <= 0 {
		t.ClockMin = 9 * 60
	}
	rate := t.GameMinPerReal
	if rate <= 0 {
		rate = 5
	}
	dt := now.Sub(t.lastTick).Seconds()
	if dt < 0 {
		dt = 0
	}
	t.lastTick = now
	t.clockAcc += dt * (rate / 60.0)
	add := int(t.clockAcc)
	if add > 0 {
		t.clockAcc -= float64(add)
		t.ClockMin = (t.ClockMin + add) % 1440
	}
	if t.RegionWeather == nil {
		t.RegionWeather = map[string]string{}
	}
	if t.cycleAt.IsZero() {
		t.cycleAt = now
		if t.Phase == "" {
			t.Phase = "DAY"
		}
		if t.Weather == "" {
			t.Weather = "CLEAR"
		}
		return false
	}
	if now.Sub(t.cycleAt).Seconds() <= DayCycleSec {
		return false
	}
	switch t.Phase {
	case "DAY":
		t.Phase = "EVENING"
		t.ClockMin = 18 * 60
	case "EVENING":
		t.Phase = "NIGHT"
		t.ClockMin = 21 * 60
	default:
		t.Phase = "DAY"
		t.ClockMin = 9 * 60
	}
	t.cycleAt = now
	t.Weather = weatherFor(t.Phase)
	t.RegionWeather["village"] = t.Weather
	t.RegionWeather["forest"] = weatherForRegion("forest", t.Phase, t.Weather)
	t.RegionWeather["plains"] = weatherForRegion("plains", t.Phase, t.Weather)
	t.RegionWeather["canyon"] = weatherForRegion("canyon", t.Phase, t.Weather)
	return true
}

func weatherFor(phase string) string {
	switch phase {
	case "EVENING":
		return "FOG"
	case "NIGHT":
		return "RAIN"
	default:
		return "CLEAR"
	}
}

func (t WorldTimeSystem) IsNight() bool {
	return t.Phase == "NIGHT"
}
