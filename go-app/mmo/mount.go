package mmo

import "time"

func (w *WorldState) requestMount(p *Player, mountID string) [][]byte {
	if mountID == "" {
		if p.ensureLog().ActiveMount != "" {
			mountID = p.ensureLog().ActiveMount
		} else {
			mountID = "wind-runner"
		}
	}
	def, ok := mountByID[mountID]
	if !ok {
		return rejectFor(p.ID, TypeRequestMount, "mount")
	}
	if !p.alive() || p.InstanceID != "" {
		return rejectFor(p.ID, TypeRequestMount, "restricted")
	}
	if p.inCombatDisabledZone() || !mountZoneAllowed(p) {
		return rejectFor(p.ID, TypeRequestMount, "sanctuary")
	}
	if p.ensureLog().PendingCinematic != "" {
		return rejectFor(p.ID, TypeRequestMount, "cinematic")
	}
	if w.Boss != nil && w.Boss.State == "ACTIVE" {
		if e := w.enemies[w.Boss.EnemyID]; e != nil && nearby(p.X, p.Z, e.X, e.Z, 18) {
			return rejectFor(p.ID, TypeRequestMount, "arena")
		}
	}
	if time.Now().Before(p.InCombatUntil) {
		return rejectFor(p.ID, TypeRequestMount, "combat")
	}
	if p.Level < def.RequiredLevel {
		return rejectFor(p.ID, TypeRequestMount, "level")
	}
	log := p.ensureLog()
	owned := p.ownsMount(def.ID)
	if !owned {
		questOK := false
		if def.RequiredQuest != "" {
			q := p.quest(def.RequiredQuest)
			if q != nil && (q.State == QuestClaimed || q.State == QuestCompleted) {
				questOK = true
			}
		}
		flagOK := def.UnlockFlag != "" && log.Flags[def.UnlockFlag]
		if !questOK && !flagOK {
			return rejectFor(p.ID, TypeRequestMount, "quest")
		}
		if !w.grantMount(p, def.ID) && !p.ownsMount(def.ID) {
			return rejectFor(p.ID, TypeRequestMount, "owned")
		}
	}
	log.ActiveMount = def.ID
	p.Mounted = true
	p.MountID = def.ID
	p.MountState = MountAnimMounted
	return [][]byte{marshal(TypeMountUpdated, MountView{
		PlayerID: p.ID, MountID: p.MountID, Mounted: true, Name: def.Name, State: p.MountState,
	})}
}

func (w *WorldState) dismount(p *Player, reason string) [][]byte {
	if !p.Mounted && p.MountID == "" {
		return nil
	}
	p.Mounted = false
	p.MountState = MountAnimIdle
	id := p.MountID
	if reason == "combat" {
		p.MountID = id
	}
	return [][]byte{marshal(TypeMountUpdated, MountView{
		PlayerID: p.ID, MountID: p.MountID, Mounted: false, Reason: reason, State: p.MountState,
	})}
}
