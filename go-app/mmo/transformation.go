package mmo

import (
	"math"
	"time"
)

func (w *WorldState) requestTransform(p *Player, formID string) [][]byte {
	if formID == "" || formID == "normal" {
		return w.endTransform(p, "cancel")
	}
	if w.limited("transform:"+p.ID, Phase25TransformRPS, time.Second) {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "rate", PlayerID: p.ID})}
	}
	def, ok := transformByID[formID]
	if !ok {
		return rejectFor(p.ID, TypeRequestTransformation, "form")
	}
	now := time.Now()
	if !p.alive() {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "dead", PlayerID: p.ID})}
	}
	if now.Before(p.StunUntil) {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "stunned", PlayerID: p.ID})}
	}
	if p.ensureLog().Quiz.Active {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "education", PlayerID: p.ID})}
	}
	if p.ensureLog().PendingCinematic != "" {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "cinematic", PlayerID: p.ID})}
	}
	if p.TransformState == "TRANSFORMING" {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "busy", PlayerID: p.ID})}
	}
	if p.InstanceID != "" && w.Dungeons != nil {
		if inst := w.dungeonOf(p.ID); inst != nil && (inst.State == DunStarting || inst.State == DunClosing) {
			return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "restricted", PlayerID: p.ID})}
		}
	}
	if !p.hasForm(formID) && !w.canAwaken(p, def) {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "locked", PlayerID: p.ID, FormID: formID})}
	}
	if p.Level < def.RequiredLevel {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "level", PlayerID: p.ID})}
	}
	if now.Before(p.TransformCDUntil) && p.FormID != formID {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "cooldown", PlayerID: p.ID})}
	}
	if p.Energy < def.EnergyCost {
		return [][]byte{marshal(TypeTransformationRejected, TransformReject{Reason: "energy", PlayerID: p.ID})}
	}
	was := p.hasForm(formID)
	p.unlockForm(formID)
	if !was {
		w.audit("transformationUnlocked", p.ID, formID)
	}
	p.Energy -= def.EnergyCost
	if p.Energy < 0 {
		p.Energy = 0
	}
	dur := def.Duration
	if p.hasSkill("aura_mastery") {
		dur *= 1.1
	}
	p.FormID = formID
	p.TransformState = "TRANSFORMING"
	p.TransformUntil = now.Add(time.Duration(dur * float64(time.Second)))
	p.MaxTransEnergy = int(dur)
	if p.MaxTransEnergy < 1 {
		p.MaxTransEnergy = 1
	}
	p.TransEnergy = p.MaxTransEnergy
	p.applyDerived()
	w.persistGear(p)
	w.persist(p)
	w.phase25OnTransform(p, formID)
	return [][]byte{marshal(TypeTransformationStarted, p.transformView())}
}

func (w *WorldState) canAwaken(p *Player, def TransformDef) bool {
	if p.hasForm(def.ID) {
		return true
	}
	if p.Level < def.RequiredLevel {
		return false
	}
	if def.RequiredQuest != "" {
		ql := p.quest(def.RequiredQuest)
		if ql != nil && (ql.State == QuestClaimed || ql.State == QuestCompleted) {
			return true
		}
	}
	if def.ID == "aura-1" && chapterTwoDone(p) {
		return true
	}
	for _, q := range awakenCatalog {
		if q.FormID != def.ID {
			continue
		}
		if p.Level < q.RequiredLevel {
			return false
		}
		log := p.ensureLog()
		if q.RequiredChapter != "" {
			st := chapterStatus(p, chapterByID[q.RequiredChapter])
			if st != "COMPLETED" && !log.Flags["chapter_"+q.RequiredChapter+"_complete"] {
				if !(q.RequiredChapter == "ch01" && log.Flags["chapter1_complete"]) {
					return false
				}
			}
		}
		for _, f := range q.RequiredFlags {
			if !log.Flags[f] {
				return false
			}
		}
		if q.RequiredKill != "" && !log.Flags["killed_"+q.RequiredKill] {
			return false
		}
		return true
	}
	return false
}

func (w *WorldState) refreshAwakening(p *Player) [][]byte {
	var out [][]byte
	for _, q := range awakenCatalog {
		if p.hasForm(q.FormID) {
			continue
		}
		def := transformByID[q.FormID]
		if def.ID == "" {
			continue
		}
		if !w.canAwaken(p, def) {
			continue
		}
		p.unlockForm(q.FormID)
		out = append(out, marshal(TypeSkillUnlocked, map[string]any{"skillId": q.FormID, "formId": q.FormID, "toId": p.ID}))
	}
	if len(out) > 0 {
		w.persist(p)
		w.persistGear(p)
		out = append(out, marshal(TypeProgressionState, p.progressionView()))
	}
	return out
}

func (w *WorldState) tickTransform(p *Player, dt float64, now time.Time) [][]byte {
	if p.TransformState == "COOLDOWN" && now.After(p.TransformCDUntil) {
		p.TransformState = "NORMAL"
		return [][]byte{marshal(TypeTransformationUpdated, p.transformView())}
	}
	if p.FormID == "" || p.FormID == "normal" {
		if p.TransformState == "TRANSFORMING" || p.TransformState == "TRANSFORMED" || p.TransformState == "ENDING" {
			return w.endTransform(p, "idle")
		}
		return nil
	}
	def := transformByID[p.FormID]
	if p.TransformState == "TRANSFORMING" && now.After(p.TransformUntil.Add(-time.Duration(def.Duration*float64(time.Second))+800*time.Millisecond)) {
		p.TransformState = "TRANSFORMED"
		p.applyDerived()
		return [][]byte{marshal(TypeTransformationUpdated, p.transformView())}
	}
	if p.TransformState == "TRANSFORMING" || p.TransformState == "TRANSFORMED" {
		drain := def.DrainPerSec
		if drain <= 0 {
			drain = 1
		}
		p.TransEnergyAcc += drain * dt
		if p.TransEnergyAcc >= 1 {
			n := int(p.TransEnergyAcc)
			p.TransEnergyAcc -= float64(n)
			p.TransEnergy -= n
		}
		if p.TransEnergy < 0 {
			p.TransEnergy = 0
		}
		if p.TransEnergy <= 0 || now.After(p.TransformUntil) || !p.alive() {
			return w.endTransform(p, "energy")
		}
	}
	return nil
}

func (w *WorldState) endTransform(p *Player, reason string) [][]byte {
	def := transformByID[p.FormID]
	cd := 8.0
	if def.Cooldown > 0 {
		cd = def.Cooldown
	}
	p.FormID = "normal"
	p.TransformState = "ENDING"
	p.TransEnergy = 0
	p.TransformCDUntil = time.Now().Add(time.Duration(cd * float64(time.Second)))
	p.applyDerived()
	view := p.transformView()
	view.Reason = reason
	p.TransformState = "COOLDOWN"
	w.persistGear(p)
	return [][]byte{marshal(TypeTransformationEnded, view)}
}

func (p *Player) transformView() TransformView {
	def := transformByID[p.FormID]
	state := p.TransformState
	if state == "" {
		state = "NORMAL"
	}
	return TransformView{
		PlayerID: p.ID, FormID: p.FormID, Name: def.Name, Visual: def.Visual, State: state,
		Energy: p.TransEnergy, MaxEnergy: p.MaxTransEnergy, Until: p.TransformUntil.UnixMilli(),
		AuraColor: def.AuraColor, Particles: def.Particles,
	}
}

func (p *Player) transformBuff() StatBlock {
	if p.FormID == "" || p.FormID == "normal" {
		return StatBlock{}
	}
	if p.TransformState != "TRANSFORMING" && p.TransformState != "TRANSFORMED" {
		return StatBlock{}
	}
	def := transformByID[p.FormID]
	atk := int(float64(p.BaseStrength)*def.AttackPct + 0.5)
	defv := int(float64(p.BaseDefense)*def.DefensePct + 0.5)
	eng := int(float64(p.BaseMaxEnergy)*def.EnergyPct + 0.5)
	return StatBlock{Attack: atk, Strength: atk, Defense: defv, MaxEnergy: eng, EnergyPower: eng / 4}
}

func (w *WorldState) applySkillEffect(p *Player, def SkillDef) {
	switch def.Effect {
	case "dash":
		yaw := p.Yaw
		p.VX = math.Sin(yaw) * DodgeSpeed * 0.85
		p.VZ = math.Cos(yaw) * DodgeSpeed * 0.85
		p.CombatState = "DODGING"
		p.DodgeUntil = time.Now().Add(220 * time.Millisecond)
	case "guard":
		p.InvulnUntil = time.Now().Add(2 * time.Second)
	}
}
