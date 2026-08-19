package mmo

import (
	"strings"
	"time"
	"unicode"
)

var guildAnnounceMax = 160

type GuildMember struct {
	PlayerID     string
	Rank         string
	JoinedAt     time.Time
	Contribution int
}

type Guild struct {
	ID, Name, Tag, LeaderID, Announcement, EmblemID, Desc string
	Level, Exp                                           int
	Members                                              map[string]*GuildMember
	CreatedAt                                            time.Time
	QuestID                                              string
	QuestProgress                                        int
	QuestDay                                             string
	WeeklyWeek                                           string
	WeeklyDungeon                                        int
	WeeklyRewarded                                       bool
	Apps                                                 map[string]bool
	Storage                                              *Inventory
	Logs                                                 []GuildLog
	PendingLeader                                        string
	PendingUntil                                         time.Time
	DailyRewarded                                        bool
	WeeklyQuestWeek                                      string
	WeeklyQuestProg                                      int
	WeeklyQuestDone                                      bool
	Workshop                                             bool
	HallEvent                                            string
}

func (g *Guild) memberIDs() []string {
	if g == nil {
		return nil
	}
	ids := make([]string, 0, len(g.Members))
	for id := range g.Members {
		ids = append(ids, id)
	}
	return ids
}

type GuildHub struct {
	ByID     map[string]*Guild
	ByName   map[string]string
	ByTag    map[string]string
	ByPlayer map[string]string
	Invites  map[string]string
}

func NewGuildHub() *GuildHub {
	return &GuildHub{
		ByID: map[string]*Guild{}, ByName: map[string]string{}, ByTag: map[string]string{},
		ByPlayer: map[string]string{}, Invites: map[string]string{},
	}
}

func (w *WorldState) ApplyGuild(id string, env Envelope) [][]byte {
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	if w.Guilds == nil {
		w.Guilds = NewGuildHub()
	}
	switch env.Type {
	case TypeGetGuild:
		return [][]byte{marshal(TypeGuildUpdated, w.guildViewFor(p))}
	case TypeGuildCreate:
		var in struct {
			Name, Tag string
		}
		_ = unmarshal(env.Data, &in)
		return w.guildCreate(p, in.Name, in.Tag)
	case TypeGuildInvite:
		var in PartyActionIn
		_ = unmarshal(env.Data, &in)
		return w.guildInvite(p, in.TargetID)
	case TypeGuildAccept:
		return w.guildAccept(p)
	case TypeGuildDecline:
		delete(w.Guilds.Invites, p.ID)
		return [][]byte{marshal(TypeGuildUpdated, w.guildViewFor(p))}
	case TypeGuildLeave:
		return w.guildLeave(p)
	case TypeGuildKick:
		var in PartyActionIn
		_ = unmarshal(env.Data, &in)
		return w.guildKick(p, in.TargetID)
	case TypeGuildDisband:
		return w.guildDisband(p)
	case TypeGuildTransfer:
		var in PartyActionIn
		_ = unmarshal(env.Data, &in)
		return w.guildTransfer(p, in.TargetID)
	case TypeGuildAnnounce:
		var in struct{ Text string }
		_ = unmarshal(env.Data, &in)
		return w.guildAnnounce(p, in.Text)
	case TypeSetGuildRank:
		var in struct {
			TargetID, Rank string
		}
		_ = unmarshal(env.Data, &in)
		return w.guildSetRank(p, in.TargetID, in.Rank)
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) guildCreate(p *Player, name, tag string) [][]byte {
	name = strings.TrimSpace(name)
	tag = strings.ToUpper(strings.TrimSpace(tag))
	if p.Level < guildCfg.CreateLevel {
		return rejectFor(p.ID, TypeGuildCreate, "level")
	}
	if currencyOf(p.ensureLog(), "coin") < guildCfg.CreateCost {
		return rejectFor(p.ID, TypeGuildCreate, "coin")
	}
	if w.Guilds.ByPlayer[p.ID] != "" {
		return rejectFor(p.ID, TypeGuildCreate, "already")
	}
	if !validGuildName(name) || !validGuildTag(tag) || !validGuildText(name) {
		return rejectFor(p.ID, TypeGuildCreate, "name")
	}
	if w.Guilds.ByName[strings.ToLower(name)] != "" || w.Guilds.ByTag[tag] != "" {
		return rejectFor(p.ID, TypeGuildCreate, "duplicate")
	}
	if !w.removeCurrency(p, "coin", guildCfg.CreateCost, "guild_create") {
		return rejectFor(p.ID, TypeGuildCreate, "coin")
	}
	g := &Guild{
		ID: randomID("gld_"), Name: name, Tag: tag, LeaderID: p.ID, Level: 1, EmblemID: "emblem-dawn",
		Members:   map[string]*GuildMember{p.ID: {PlayerID: p.ID, Rank: "LEADER", JoinedAt: time.Now()}},
		CreatedAt: time.Now(), QuestID: "gq-shadow", QuestDay: utcDay(),
	}
	w.Guilds.ByID[g.ID] = g
	w.Guilds.ByName[strings.ToLower(name)] = g.ID
	w.Guilds.ByTag[tag] = g.ID
	w.Guilds.ByPlayer[p.ID] = g.ID
	p.GuildID, p.GuildTag = g.ID, g.Tag
	p.ensureLog().GuildID = g.ID
	p.grantTitle("dawn-defender")
	w.audit("GUILD_CHANGE", p.ID, "create:"+g.ID)
	w.persist(p)
	out := [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
	out = append(out, w.grantSocialFlag(p, "ach_guild_founder")...)
	out = append(out, w.grantSocialFlag(p, "ach_guild_member")...)
	return out
}

func validGuildName(name string) bool {
	n := len([]rune(name))
	if n < guildCfg.NameMin || n > guildCfg.NameMax {
		return false
	}
	for _, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-') {
			return false
		}
	}
	low := strings.ToLower(name)
	for _, w := range guildCfg.Reserved {
		if low == w {
			return false
		}
	}
	return true
}

func validGuildTag(tag string) bool {
	if len(tag) < guildCfg.TagMin || len(tag) > guildCfg.TagMax {
		return false
	}
	for _, r := range tag {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

func (w *WorldState) guildOf(id string) *Guild {
	if w.Guilds == nil {
		return nil
	}
	return w.Guilds.ByID[w.Guilds.ByPlayer[id]]
}

func (g *Guild) rankOf(id string) string {
	if g == nil || g.Members[id] == nil {
		return ""
	}
	return g.Members[id].Rank
}

func canInvite(rank string) bool {
	return rank == "LEADER" || rank == "MASTER" || rank == "OFFICER"
}

func (w *WorldState) guildInvite(from *Player, targetID string) [][]byte {
	g := w.guildOf(from.ID)
	if g == nil || !canInvite(g.rankOf(from.ID)) {
		return rejectFor(from.ID, TypeGuildInvite, "perm")
	}
	if w.limited("ginv:"+from.ID, 8, time.Minute) {
		return rejectFor(from.ID, TypeGuildInvite, "rate")
	}
	to := w.players[targetID]
	if to == nil || !to.Connected {
		return rejectFor(from.ID, TypeGuildInvite, "target")
	}
	if w.Social != nil && w.Social.blocked(from.ID, to.ID) {
		return rejectFor(from.ID, TypeGuildInvite, "blocked")
	}
	if w.Guilds.ByPlayer[to.ID] != "" {
		return rejectFor(from.ID, TypeGuildInvite, "already")
	}
	if len(g.Members) >= guildMemberCap(g.Level) {
		return rejectFor(from.ID, TypeGuildInvite, "full")
	}
	w.Guilds.Invites[to.ID] = g.ID
	ev := w.notify(to, "guild_invite", from.Name+" invited you to ["+g.Tag+"] "+g.Name+".")
	return [][]byte{marshal(TypeGuildInvite, map[string]any{"guildId": g.ID, "from": from.Name, "fromId": from.ID, "toId": to.ID}), ev}
}

func (w *WorldState) guildAccept(p *Player) [][]byte {
	gid := w.Guilds.Invites[p.ID]
	g := w.Guilds.ByID[gid]
	if g == nil {
		return rejectFor(p.ID, TypeGuildAccept, "invite")
	}
	if w.Guilds.ByPlayer[p.ID] != "" {
		return rejectFor(p.ID, TypeGuildAccept, "already")
	}
	if len(g.Members) >= guildMemberCap(g.Level) {
		return rejectFor(p.ID, TypeGuildAccept, "full")
	}
	delete(w.Guilds.Invites, p.ID)
	g.Members[p.ID] = &GuildMember{PlayerID: p.ID, Rank: "MEMBER", JoinedAt: time.Now()}
	w.Guilds.ByPlayer[p.ID] = g.ID
	p.GuildID, p.GuildTag = g.ID, g.Tag
	p.ensureLog().GuildID = g.ID
	w.persist(p)
	out := [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
	out = append(out, w.grantSocialFlag(p, "ach_guild_member")...)
	return out
}

func (w *WorldState) guildLeave(p *Player) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil {
		return rejectFor(p.ID, TypeGuildLeave, "none")
	}
	if g.LeaderID == p.ID && len(g.Members) > 1 {
		return rejectFor(p.ID, TypeGuildLeave, "leader")
	}
	if g.LeaderID == p.ID {
		return w.guildDisband(p)
	}
	delete(g.Members, p.ID)
	delete(w.Guilds.ByPlayer, p.ID)
	p.GuildID, p.GuildTag = "", ""
	p.ensureLog().GuildID = ""
	w.persist(p)
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g)), marshal(TypeGuildUpdated, GuildView{})}
}

func (w *WorldState) guildKick(leader *Player, targetID string) [][]byte {
	g := w.guildOf(leader.ID)
	if g == nil || !canInvite(g.rankOf(leader.ID)) {
		return rejectFor(leader.ID, TypeGuildKick, "perm")
	}
	if targetID == g.LeaderID || targetID == leader.ID {
		return rejectFor(leader.ID, TypeGuildKick, "target")
	}
	if g.rankOf(leader.ID) != "LEADER" && g.rankOf(targetID) == "OFFICER" {
		return rejectFor(leader.ID, TypeGuildKick, "rank")
	}
	delete(g.Members, targetID)
	delete(w.Guilds.ByPlayer, targetID)
	w.audit("GUILD_KICK", leader.ID, targetID)
	if t := w.players[targetID]; t != nil {
		t.GuildID, t.GuildTag = "", ""
		t.ensureLog().GuildID = ""
		w.persist(t)
	}
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
}

func (w *WorldState) guildDisband(p *Player) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil || g.LeaderID != p.ID {
		return rejectFor(p.ID, TypeGuildDisband, "leader")
	}
	for id := range g.Members {
		delete(w.Guilds.ByPlayer, id)
		if m := w.players[id]; m != nil {
			m.GuildID, m.GuildTag = "", ""
			m.ensureLog().GuildID = ""
			w.persist(m)
		}
	}
	delete(w.Guilds.ByID, g.ID)
	delete(w.Guilds.ByName, strings.ToLower(g.Name))
	delete(w.Guilds.ByTag, g.Tag)
	w.audit("GUILD_CHANGE", p.ID, "disband:"+g.ID)
	return [][]byte{marshal(TypeGuildUpdated, GuildView{})}
}

func (w *WorldState) guildTransfer(p *Player, targetID string) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil || g.LeaderID != p.ID {
		return rejectFor(p.ID, TypeGuildTransfer, "leader")
	}
	m := g.Members[targetID]
	if m == nil {
		return rejectFor(p.ID, TypeGuildTransfer, "member")
	}
	now := time.Now()
	if g.PendingLeader != targetID || now.After(g.PendingUntil) {
		g.PendingLeader = targetID
		g.PendingUntil = now.Add(30 * time.Second)
		return [][]byte{marshal(TypeGuildUpdated, w.guildView(g)), w.notify(p, "guild_transfer", "Confirm transfer to complete leadership change.")}
	}
	g.Members[p.ID].Rank = "OFFICER"
	m.Rank = "LEADER"
	g.LeaderID = targetID
	g.PendingLeader = ""
	w.audit("GUILD_CHANGE", p.ID, "transfer:"+targetID)
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
}

func (w *WorldState) guildAnnounce(p *Player, text string) [][]byte {
	g := w.guildOf(p.ID)
	if g == nil || !canInvite(g.rankOf(p.ID)) {
		return rejectFor(p.ID, TypeGuildAnnounce, "perm")
	}
	if w.limited("gann:"+p.ID, 3, time.Minute) {
		return rejectFor(p.ID, TypeGuildAnnounce, "rate")
	}
	maxAnn := guildAnnounceMax
	if maxAnn < 1 {
		maxAnn = 160
	}
	if len([]rune(text)) > maxAnn {
		text = string([]rune(text)[:maxAnn])
	}
	g.Announcement = filterChat(text)
	return [][]byte{marshal(TypeGuildUpdated, w.guildView(g))}
}

func (w *WorldState) guildContribute(p *Player, contrib, gexp int) {
	g := w.guildOf(p.ID)
	if g == nil || contrib <= 0 {
		return
	}
	if g.Members[p.ID] != nil {
		g.Members[p.ID].Contribution += contrib
	}
	g.Exp += gexp
	for i := len(guildCfg.Levels) - 1; i >= 0; i-- {
		if g.Exp >= guildCfg.Levels[i] {
			g.Level = i + 1
			break
		}
	}
	if g.Level >= 2 {
		p.ensureLog().Flags["guild_emblem"] = true
	}
	if g.Level >= 5 {
		p.grantTitle("dawn-defender")
	}
}

func (w *WorldState) guildView(g *Guild) GuildView {
	if g == nil {
		return GuildView{}
	}
	if g.QuestDay != utcDay() {
		g.QuestProgress = 0
		g.QuestDay = utcDay()
	}
	members := make([]GuildMemberView, 0, len(g.Members))
	for id, m := range g.Members {
		name := id
		if p := w.players[id]; p != nil {
			name = p.Name
		}
		members = append(members, GuildMemberView{PlayerID: id, Name: name, Rank: guildDisplayRank(m.Rank), Contribution: m.Contribution, JoinedAt: m.JoinedAt.UnixMilli()})
	}
	apps := make([]string, 0, len(g.Apps))
	for id := range g.Apps {
		apps = append(apps, id)
	}
	var storage []InvSlotView
	if g.Storage != nil {
		storage = g.Storage.views()
	}
	return GuildView{
		ID: g.ID, Name: g.Name, Tag: g.Tag, LeaderID: g.LeaderID, Level: g.Level, Exp: g.Exp, Members: members,
		Announcement: g.Announcement, EmblemID: g.EmblemID, Quest: g.QuestID, Description: g.Desc,
		Capacity: guildMemberCap(g.Level), Storage: storage, Logs: g.Logs, Apps: apps,
		PendingLead: g.PendingLeader, QuestProg: g.QuestProgress, NotifyIDs: g.memberIDs(),
	}
}

func (w *WorldState) guildViewFor(p *Player) GuildView {
	return w.guildView(w.guildOf(p.ID))
}

func utcDay() string {
	return time.Now().UTC().Format("2006-01-02")
}
