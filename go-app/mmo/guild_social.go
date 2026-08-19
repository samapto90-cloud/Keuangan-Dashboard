package mmo

import (
	"strings"
	"time"
)

func (w *WorldState) guildSetRank(p *Player, targetID, rank string) [][]byte {
	rank = strings.ToUpper(strings.TrimSpace(rank))
	if targetID == "" || rank == "LEADER" || rank == "GUILD_MASTER" || rank == "MASTER" {
		return rejectFor(p.ID, TypeSetGuildRank, "server_authoritative")
	}
	g := w.guildOf(p.ID)
	if g == nil || g.LeaderID != p.ID {
		return rejectFor(p.ID, TypeSetGuildRank, "leader")
	}
	if g.Members[targetID] == nil || targetID == g.LeaderID {
		return rejectFor(p.ID, TypeSetGuildRank, "member")
	}
	if rank != "OFFICER" && rank != "MEMBER" && rank != "RECRUIT" {
		return rejectFor(p.ID, TypeSetGuildRank, "rank")
	}
	g.Members[targetID].Rank = rank
	w.audit("GUILD_CHANGE", p.ID, "rank:"+targetID+":"+rank)
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
}

func (w *WorldState) guildApply(p *Player, guildID string) [][]byte {
	if w.Guilds.ByPlayer[p.ID] != "" {
		return rejectFor(p.ID, TypeGuildApply, "already")
	}
	g := w.Guilds.ByID[guildID]
	if g == nil {
		return rejectFor(p.ID, TypeGuildApply, "guild")
	}
	if len(g.Members) >= guildMemberCap(g.Level) {
		return rejectFor(p.ID, TypeGuildApply, "full")
	}
	if g.Apps == nil {
		g.Apps = map[string]bool{}
	}
	g.Apps[p.ID] = true
	for _, id := range g.memberIDs() {
		if m := g.Members[id]; m != nil && canInvite(m.Rank) {
			if off := w.players[id]; off != nil {
				w.notify(off, "guild_apply", p.Name+" applied to "+g.Name+".")
			}
		}
	}
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g)), w.notify(p, "guild_apply", "Application sent.")}
}

func (w *WorldState) guildReview(p *Player, targetID string, accept bool) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil || !canInvite(g.rankOf(p.ID)) {
		return rejectFor(p.ID, TypeGuildReview, "perm")
	}
	if g.Apps == nil || !g.Apps[targetID] {
		return rejectFor(p.ID, TypeGuildReview, "app")
	}
	delete(g.Apps, targetID)
	if !accept {
		return [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
	}
	if w.Guilds.ByPlayer[targetID] != "" {
		return rejectFor(p.ID, TypeGuildReview, "already")
	}
	if len(g.Members) >= guildMemberCap(g.Level) {
		return rejectFor(p.ID, TypeGuildReview, "full")
	}
	g.Members[targetID] = &GuildMember{PlayerID: targetID, Rank: "MEMBER", JoinedAt: time.Now()}
	w.Guilds.ByPlayer[targetID] = g.ID
	if t := w.players[targetID]; t != nil {
		t.GuildID, t.GuildTag = g.ID, g.Tag
		t.ensureLog().GuildID = g.ID
		w.persist(t)
		w.notify(t, "guild_join", "Welcome to ["+g.Tag+"] "+g.Name+".")
	}
	w.audit("GUILD_CHANGE", p.ID, "accept:"+targetID)
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
}

func (w *WorldState) guildStorageMove(p *Player, deposit bool, slot, qty int) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil {
		return rejectFor(p.ID, TypeGuildDeposit, "guild")
	}
	if g.Storage == nil {
		g.Storage = &Inventory{ID: "gst-" + g.ID, Capacity: 40, Slots: make([]InvSlot, 40)}
	}
	if qty < 1 {
		qty = 1
	}
	p.ensureGear()
	src, dst := p.Bag, g.Storage
	act := "deposit"
	if !deposit {
		if !canInvite(g.rankOf(p.ID)) {
			return rejectFor(p.ID, TypeGuildWithdraw, "perm")
		}
		src, dst = g.Storage, p.Bag
		act = "withdraw"
	}
	if slot < 0 || slot >= len(src.Slots) {
		return rejectFor(p.ID, TypeGuildDeposit, "slot")
	}
	s := src.Slots[slot]
	if s.ItemID == "" || s.Qty < qty {
		return rejectFor(p.ID, TypeGuildDeposit, "item")
	}
	if deposit && (s.Locked || s.Favorite) {
		return rejectFor(p.ID, TypeGuildDeposit, "locked")
	}
	if !src.removeAt(slot, qty) {
		return rejectFor(p.ID, TypeGuildDeposit, "qty")
	}
	if _, ok := dst.add(s.ItemID, qty); !ok {
		src.add(s.ItemID, qty)
		return rejectFor(p.ID, TypeGuildDeposit, "space")
	}
	name := p.Name
	g.Logs = append(g.Logs, GuildLog{
		ID: randomID("gl_"), Player: name, ItemID: s.ItemID, Qty: qty, Action: act, At: time.Now().UnixMilli(),
	})
	if len(g.Logs) > 80 {
		g.Logs = g.Logs[len(g.Logs)-60:]
	}
	w.audit("GUILD_STORAGE", p.ID, act+":"+s.ItemID+":"+itoa(qty))
	w.persistGear(p)
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Guild storage.", nil)), marshal(TypeGuildUpdated, w.guildView(g))}
}

func (w *WorldState) guildLogView(p *Player) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil {
		return rejectFor(p.ID, TypeGetGuildLog, "guild")
	}
	return [][]byte{marshal(TypeGuildLog, map[string]any{"logs": g.Logs, "storage": g.Storage.views(), "toId": p.ID})}
}

func (w *WorldState) guildSetEmblem(p *Player, emblemID string) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil || g.LeaderID != p.ID {
		return rejectFor(p.ID, TypeGuildSetEmblem, "leader")
	}
	if _, ok := emblemByID[emblemID]; !ok {
		return rejectFor(p.ID, TypeGuildSetEmblem, "emblem")
	}
	g.EmblemID = emblemID
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
}

func (w *WorldState) guildSetDesc(p *Player, text string) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil || g.LeaderID != p.ID {
		return rejectFor(p.ID, TypeGuildSetDesc, "leader")
	}
	max := guildCfg.DescMax
	if max < 1 {
		max = socialCfg.GuildDescMax
	}
	if max < 1 {
		max = 300
	}
	text = strings.TrimSpace(filterChat(text))
	if len([]rune(text)) > max {
		text = string([]rune(text)[:max])
	}
	g.Desc = text
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
}

func (w *WorldState) guildQuestTick(p *Player, kind, target string, n int) {
	g := w.guildOf(p.ID)
	if g == nil || n < 1 {
		return
	}
	day := utcDay()
	week := raidWeekKey()
	if g.QuestDay != day {
		g.QuestProgress = 0
		g.QuestDay = day
		g.DailyRewarded = false
	}
	for _, q := range guildQuestList {
		if !strings.EqualFold(q.Kind, kind) {
			continue
		}
		if q.Target != "" && q.Target != target && !(kind == "KILL" && q.Target == "") {
			if q.Target != "" && q.Target != target {
				continue
			}
		}
		if q.Reset == "weekly" {
			if g.WeeklyQuestWeek != week {
				g.WeeklyQuestProg = 0
				g.WeeklyQuestWeek = week
				g.WeeklyQuestDone = false
			}
			if g.WeeklyQuestDone {
				continue
			}
			g.WeeklyQuestProg += n
			if g.WeeklyQuestProg >= q.Count {
				g.WeeklyQuestDone = true
				w.guildContribute(p, q.Rewards.Contribution, q.Rewards.GuildExp)
				if q.Rewards.Coin > 0 {
					w.addCurrency(p, "coin", q.Rewards.Coin, "GUILD")
				}
				w.phase26GuildQuestComplete(p, q)
			}
			continue
		}
		if g.DailyRewarded {
			continue
		}
		if g.QuestID == "" {
			g.QuestID = q.ID
		}
		if g.QuestID != q.ID && q.ID != g.QuestID {
			continue
		}
		g.QuestProgress += n
		if g.QuestProgress >= q.Count && !g.DailyRewarded {
			g.DailyRewarded = true
			w.guildContribute(p, q.Rewards.Contribution, q.Rewards.GuildExp)
			if q.Rewards.Coin > 0 {
				w.addCurrency(p, "coin", q.Rewards.Coin, "GUILD")
			}
			w.phase26GuildQuestComplete(p, q)
		}
	}
}

func validGuildText(name string) bool {
	low := strings.ToLower(name)
	for _, w := range badWords {
		if w != "" && strings.Contains(low, strings.ToLower(w)) {
			return false
		}
	}
	return true
}
