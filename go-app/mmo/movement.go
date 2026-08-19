package mmo

import (
	"math"
	"time"
)

func clamp01(v float64) float64 {
	if v < -1 {
		return -1
	}
	if v > 1 {
		return 1
	}
	return v
}

func (p *Player) SetInput(in MoveInput) {
	now := time.Now()
	if now.Sub(p.LastInputAt) < InputMinInterval {
		p.DroppedInputs++
		if in.Jump {
			p.JumpQueued = true
		}
		return
	}
	p.LastInputAt = now
	p.LastHeard = now
	p.AX = clamp01(in.AX)
	p.AZ = clamp01(in.AZ)
	p.CamYaw = in.Yaw
	p.Sprint = in.Sprint
	if in.Jump {
		p.JumpQueued = true
	}
	p.Seq = in.Seq
}

func (p *Player) Simulate(dt float64) {
	if dt <= 0 || dt > 0.08 {
		dt = 1.0 / float64(ServerTickRate)
	}
	if p.CombatState == "DEAD" || p.CombatState == "RESPAWNING" || p.CombatState == "DOWNED" {
		p.AX, p.AZ = 0, 0
		p.VX, p.VZ = 0, 0
		p.State = "DEAD"
		p.LastUpdate = time.Now()
		return
	}
	now := time.Now()
	if now.Before(p.DodgeUntil) {
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.Z += p.VZ * dt
		p.VY -= Gravity * dt
		if p.Y <= 0 || math.IsNaN(p.Y) || math.IsInf(p.Y, 0) {
			p.Y = 0
			p.VY = 0
			p.Grounded = true
		}
		p.X = clamp(p.X, -WorldLimit, WorldLimit)
		p.Z = clamp(p.Z, -WorldLimit, WorldLimit)
		p.State = "DODGE"
		p.LastUpdate = now
		return
	}
	if now.Before(p.StunUntil) || now.Before(p.HitUntil) {
		p.AX, p.AZ = 0, 0
	}
	mag := math.Hypot(p.AX, p.AZ)
	moving := mag > 0.08
	speed := WalkSpeed * (1 + p.MoveSpeedBonus)
	if p.Sprint && moving && p.Grounded && p.Stamina > 4 && !p.Mounted {
		speed = RunSpeed * (1 + p.MoveSpeedBonus)
		p.Stamina = math.Max(0, p.Stamina-StaminaSprintDrain*dt)
	}
	if p.Mounted {
		mult := 1.45
		if m, ok := mountByID[p.MountID]; ok && m.Speed > 1 {
			mult = clampMountSpeed(m.Speed)
		} else {
			mult = clampMountSpeed(mult)
		}
		speed = RunSpeed * (1 + p.MoveSpeedBonus) * mult
	}
	if p.Charging {
		speed *= ChargeMoveFactor
		p.Sprint = false
	}
	if p.Swimming {
		speed = WalkSpeed * (1 + p.MoveSpeedBonus)
	}
	cap := RunSpeed * MaxMountMult
	if speed > cap {
		speed = cap
	}
	if moving {
		yaw := headingFromCamera(p.CamYaw, p.AX, p.AZ)
		tx := math.Sin(yaw) * speed
		tz := math.Cos(yaw) * speed
		a := 1 - math.Exp(-Acceleration*dt)
		p.VX += (tx - p.VX) * a
		p.VZ += (tz - p.VZ) * a
		p.Yaw = yaw
	} else {
		d := math.Exp(-Deceleration * dt)
		p.VX *= d
		p.VZ *= d
		if math.Hypot(p.VX, p.VZ) < 0.04 {
			p.VX, p.VZ = 0, 0
		}
	}
	if p.JumpQueued && p.Grounded && p.CombatState != "ATTACKING" && p.CombatState != "COMBO" && p.CombatState != "CASTING" {
		p.VY = JumpForce
		p.Grounded = false
	}
	p.JumpQueued = false
	p.VY -= Gravity * dt
	p.X += p.VX * dt
	p.Y += p.VY * dt
	p.Z += p.VZ * dt
	if p.Y <= 0 || math.IsNaN(p.Y) || math.IsInf(p.Y, 0) {
		p.Y = 0
		if p.VY < 0 || math.IsNaN(p.VY) || math.IsInf(p.VY, 0) {
			p.VY = 0
		}
		p.Grounded = true
	} else {
		p.Grounded = false
	}
	p.X = clamp(p.X, -WorldLimit, WorldLimit)
	p.Z = clamp(p.Z, -WorldLimit, WorldLimit)
	if spd := math.Hypot(p.VX, p.VZ); spd > RunSpeed*MaxMountMult {
		s := (RunSpeed * MaxMountMult) / spd
		p.VX *= s
		p.VZ *= s
		p.Suspicious++
	}
	p.tickSwim(dt)
	spd := math.Hypot(p.VX, p.VZ)
	switch {
	case !p.Grounded && p.VY > 0.6:
		p.State = "JUMP"
	case !p.Grounded:
		p.State = "FALL"
	case spd > 4.5:
		p.State = "RUN"
	case spd > 0.35:
		p.State = "WALK"
	default:
		p.State = "IDLE"
	}
	p.LastUpdate = time.Now()
}

// headingFromCamera maps WASD/stick (ax = strafe, az = forward) onto camera yaw.
// Strafe is negated so D/right matches Three.js camera local +X (screen right).
func headingFromCamera(camYaw, ax, az float64) float64 {
	return camYaw + math.Atan2(-ax, az)
}
