package mmo

import (
	"strings"
	"time"
)

type Report struct {
	ID, Reporter, Target, Category, Evidence, State, MatchID string
	At                                                       time.Time
}

func (w *WorldState) ApplyModeration(id string, env Envelope) [][]byte {
	p := w.players[id]
	if p == nil {
		return nil
	}
	switch env.Type {
	case TypeReportPlayer:
		var in struct{ TargetID, Category, Evidence string }
		_ = unmarshal(env.Data, &in)
		if in.TargetID == "" || in.Category == "" {
			return rejectFor(p.ID, TypeReportPlayer, "payload")
		}
		if !reportCategoryOK(in.Category) {
			return rejectFor(p.ID, TypeReportPlayer, "category")
		}
		if w.limited("rep:"+p.ID, 6, time.Hour) {
			return rejectFor(p.ID, TypeReportPlayer, "rate")
		}
		if w.Reports == nil {
			w.Reports = []Report{}
		}
		w.Reports = append(w.Reports, Report{
			ID: randomID("rep_"), Reporter: p.ID, Target: in.TargetID, Category: in.Category,
			Evidence: in.Evidence, State: "OPEN", At: time.Now(),
		})
		return [][]byte{w.notify(p, "system", "Laporan diterima.")}
	case TypeSearchPlayer:
		var in struct{ Name string }
		_ = unmarshal(env.Data, &in)
		if w.limited("search:"+p.ID, 8, time.Minute) {
			return rejectFor(p.ID, TypeSearchPlayer, "rate")
		}
		hit := w.findByNameOrID(in.Name)
		if hit == nil || hit.ID == p.ID {
			return [][]byte{marshal(TypeSearchResult, map[string]any{"results": []SearchHit{}})}
		}
		return [][]byte{marshal(TypeSearchResult, map[string]any{"results": []SearchHit{{
			PlayerID: hit.ID, Name: hit.Name, Level: hit.Level, Guild: hit.GuildTag, Status: w.presenceOf(hit),
		}}})}
	case TypeGetNotifies:
		return [][]byte{marshal(TypeNotifyList, map[string]any{"notifies": p.ensureLog().Notices})}
	case TypeReportMessage:
		return w.reportMessage(p, env.Data)
	case TypeMutePlayer, TypeModKick, TypeModBan:
		return rejectFor(p.ID, env.Type, "moderator")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) presenceOf(p *Player) string {
	if p == nil || !p.Connected {
		return "OFFLINE"
	}
	if mode := strings.ToUpper(strings.TrimSpace(p.ensureLog().PresenceMode)); mode == "AWAY" || mode == "BUSY" || mode == "ONLINE" {
		if mode != "ONLINE" {
			return mode
		}
	}
	if w.pvpOf(p.ID) != nil {
		return "IN PVP"
	}
	if p.InstanceID != "" && !houseID(p.InstanceID) {
		return "IN DUNGEON"
	}
	if time.Now().Before(p.InCombatUntil) {
		return "BUSY"
	}
	if time.Since(p.LastHeard) > 2*time.Minute {
		return "AWAY"
	}
	return "ONLINE"
}
