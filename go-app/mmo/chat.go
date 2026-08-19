package mmo

import (
	"strings"
	"time"
)

var chatRuneMax = 180

func (w *WorldState) ApplyChat(id string, env Envelope) [][]byte {
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	var in struct {
		Channel, Text, Target string
	}
	_ = unmarshal(env.Data, &in)
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return rejectFor(p.ID, TypeChat, "empty")
	}
	if w.muted(p.ID) {
		return rejectFor(p.ID, TypeChat, "mute")
	}
	chIn := strings.ToUpper(in.Channel)
	if chIn == "SYSTEM" {
		return rejectFor(p.ID, TypeChat, "system")
	}
	burst, win := 5, 5*time.Second
	if chIn == "WORLD" || chIn == "GLOBAL" {
		n, sec := socialCfg.ChatBurst, socialCfg.ChatWindowSec
		if n < 1 {
			n = 1
		}
		if sec < 1 {
			sec = 2
		}
		burst, win = n, time.Duration(sec)*time.Second
	}
	if w.limited("chat:"+p.ID, burst, win) {
		if w.Mutes == nil {
			w.Mutes = map[string]time.Time{}
		}
		w.Mutes[p.ID] = time.Now().Add(15 * time.Second)
		return rejectFor(p.ID, TypeChat, "spam")
	}
	if strings.HasPrefix(text, "/w ") || strings.HasPrefix(text, "/whisper ") {
		in.Channel = "WHISPER"
		parts := strings.SplitN(text, " ", 3)
		if len(parts) >= 3 {
			in.Target, text = parts[1], parts[2]
		}
	}
	if strings.HasPrefix(text, "/guild ") {
		in.Channel = "GUILD"
		text = strings.TrimPrefix(text, "/guild ")
	}
	if strings.HasPrefix(text, "/party ") {
		in.Channel = "PARTY"
		text = strings.TrimPrefix(text, "/party ")
	}
	ch := strings.ToUpper(in.Channel)
	if ch == "GLOBAL" {
		ch = "WORLD"
	}
	if ch == "AREA" {
		ch = "LOCAL"
	}
	if ch == "PRIVATE" {
		ch = "WHISPER"
	}
	if inst := w.pvpOf(p.ID); inst != nil {
		if ch == "WORLD" || ch == "" || ch == "LOCAL" {
			ch = "MATCH"
		}
	}
	if ch == "" {
		ch = "LOCAL"
	}
	text = filterChat(text)
	maxRunes := chatRuneMax
	if maxRunes < 1 {
		maxRunes = 180
	}
	if len([]rune(text)) > maxRunes {
		text = string([]rune(text)[:maxRunes])
	}
	if w.lastChat == nil {
		w.lastChat = map[string]string{}
	}
	if w.lastChat[p.ID] == text && text != "" {
		return rejectFor(p.ID, TypeChat, "repeat")
	}
	w.lastChat[p.ID] = text
	msg := ChatOut{Channel: ch, FromID: p.ID, From: p.Name, Text: text}
	switch ch {
	case "SYSTEM":
		return rejectFor(p.ID, TypeChat, "system")
	case "WHISPER":
		to := w.findByNameOrID(in.Target)
		if to == nil {
			return rejectFor(p.ID, TypeChat, "target")
		}
		if w.Social != nil && w.Social.blocked(p.ID, to.ID) {
			return rejectFor(p.ID, TypeChat, "blocked")
		}
		if !w.privacyAllows(to.ensureLog().privacyPM(), p.ID, to.ID) {
			return rejectFor(p.ID, TypeChat, "privacy")
		}
		msg.ToID = to.ID
		msg.NotifyIDs = w.filterMutedNotify(p.ID, []string{p.ID, to.ID})
		return [][]byte{marshal(TypeChatMessage, msg)}
	case "RAID":
		inst := w.dungeonOf(p.ID)
		if inst == nil || dungeonKind(dungeonByID[inst.DefID]) != "RAID" {
			return rejectFor(p.ID, TypeChat, "raid")
		}
		msg.NotifyIDs = append([]string{}, inst.Players...)
		msg.NotifyIDs = w.filterMutedNotify(p.ID, msg.NotifyIDs)
		return [][]byte{marshal(TypeChatMessage, msg)}
	case "PARTY":
		pt := w.Parties.Of(p.ID)
		if pt == nil {
			return rejectFor(p.ID, TypeChat, "party")
		}
		msg.NotifyIDs = w.filterMutedNotify(p.ID, append([]string{}, pt.Members...))
		return [][]byte{marshal(TypeChatMessage, msg)}
	case "TEAM":
		inst := w.pvpOf(p.ID)
		if inst == nil {
			return rejectFor(p.ID, TypeChat, "match")
		}
		f := inst.Fighters[p.ID]
		ids := []string{}
		if f != nil {
			side := inst.TeamA
			if f.Team == 2 {
				side = inst.TeamB
			}
			ids = append([]string{}, side...)
		}
		msg.NotifyIDs = ids
		return [][]byte{marshal(TypeChatMessage, msg)}
	case "MATCH":
		inst := w.pvpOf(p.ID)
		if inst == nil {
			return rejectFor(p.ID, TypeChat, "match")
		}
		msg.NotifyIDs = append([]string{}, inst.Players...)
		return [][]byte{marshal(TypeChatMessage, msg)}
	case "GUILD":
		g := w.guildOf(p.ID)
		if g == nil {
			return rejectFor(p.ID, TypeChat, "guild")
		}
		msg.NotifyIDs = g.memberIDs()
		msg.NotifyIDs = w.filterMutedNotify(p.ID, msg.NotifyIDs)
		return [][]byte{marshal(TypeChatMessage, msg)}
	case "WORLD":
		ids := make([]string, 0, len(w.players))
		for _, o := range w.players {
			if o.Connected {
				ids = append(ids, o.ID)
			}
		}
		msg.NotifyIDs = w.filterMutedNotify(p.ID, ids)
		w.appendChatLog(msg)
		return [][]byte{marshal(TypeChatMessage, msg)}
	default:
		ids := []string{p.ID}
		for _, o := range w.players {
			if o.ID == p.ID || !o.Connected {
				continue
			}
			if hypot2(p.X, p.Z, o.X, o.Z) > 30*30 {
				continue
			}
			ids = append(ids, o.ID)
		}
		msg.NotifyIDs = w.filterMutedNotify(p.ID, ids)
		return [][]byte{marshal(TypeChatMessage, msg)}
	}
}

func (w *WorldState) findByNameOrID(q string) *Player {
	if q == "" {
		return nil
	}
	if p := w.players[q]; p != nil {
		return p
	}
	low := strings.ToLower(q)
	for _, p := range w.players {
		if strings.ToLower(p.Name) == low {
			return p
		}
	}
	return nil
}

func (w *WorldState) systemChat(text string) [][]byte {
	ids := make([]string, 0, len(w.players))
	for _, o := range w.players {
		if o.Connected {
			ids = append(ids, o.ID)
		}
	}
	msg := ChatOut{Channel: "SYSTEM", From: "SERVER", Text: filterChat(text), System: true, NotifyIDs: ids}
	w.appendChatLog(msg)
	return [][]byte{marshal(TypeChatMessage, msg)}
}

func (w *WorldState) appendChatLog(msg ChatOut) {
	w.ChatLog = append(w.ChatLog, msg)
	if len(w.ChatLog) > 40 {
		w.ChatLog = w.ChatLog[len(w.ChatLog)-30:]
	}
}
