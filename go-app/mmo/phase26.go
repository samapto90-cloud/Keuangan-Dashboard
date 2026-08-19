package mmo

import (
	"strings"
	"time"
	"unicode"
)

// Phase 26 overlay: social MMORPG (friend, party, guild, chat, trade, presence,
// moderation, privacy). Reuses PartyService, GuildService, ChatService,
// FriendService, TradeService, PlayerService, PresenceService,
// NotificationService, QuestService, DungeonService, RaidService, WorldService,
// InventoryService, EquipmentService, AchievementService, ModerationService.
// Do not duplicate those services.
//
// Logical tables on SocialHub / GuildHub / PartyHub / TradeHub / PlayerLog:
// friends, friend_requests, blocks, parties, party_members, guilds,
// guild_members, guild_ranks, guild_permissions, guild_quests, guild_progress,
// guild_announcements, guild_tokens, chat_messages, chat_channels, chat_mutes,
// chat_reports, trade_sessions, trade_items, trade_confirmations,
// player_presence, social_notifications.
// Indexes: playerId, friendId, guildId, partyId, channelId, tradeId.

const (
	PrivacyEveryone     = "EVERYONE"
	PrivacyFriends      = "FRIENDS"
	PrivacyNone         = "NONE"
	Phase26FriendLimit  = 200
	Phase26BlockLimit   = 100
	Phase26GuildMembers = 50
	Phase26ChatRunes    = 200
	Phase26DescMax      = 200
	Phase26GuildQuestID = "gq-desa-1"
	CommunityEventID    = "gotong-royong-desa"
)

func init() {
	registerPhase26()
}

func registerPhase26() {
	socialCfg.FriendLimit = Phase26FriendLimit
	socialCfg.BlockLimit = Phase26BlockLimit
	socialCfg.GuildMemberBase = Phase26GuildMembers
	socialCfg.GuildDescMax = Phase26DescMax
	if guildCfg.MemberBase < Phase26GuildMembers {
		guildCfg.MemberBase = Phase26GuildMembers
	}
	guildCfg.DescMax = Phase26DescMax
	chatRuneMax = Phase26ChatRunes
	guildAnnounceMax = Phase26DescMax
	registerPhase26Content()
}

func registerPhase26Content() {
	desa := GuildQuestDef{ID: Phase26GuildQuestID, Title: "Pertahanan Desa.", Kind: "KILL", Count: 100, Reset: "weekly"}
	desa.Rewards.GuildExp = 80
	desa.Rewards.Contribution = 40
	desa.Rewards.Coin = 20
	registerGuildQuest(desa)
	registerEventDef(EventDef{
		ID: CommunityEventID, Name: "Gotong Royong Desa.", Kind: "COMMUNITY", Region: "village",
		Announce: "Gotong Royong Desa. Kumpulkan bahan bersama.", DurationSec: 1800,
		Objective: "collect materials", ObjectiveNeed: 40, X: 0, Z: 4,
	})
	registerAchievementDef(AchievementDef{ID: "first-friend", Name: "First Friend", Title: "first-friend", Flag: "ach_first_friend", Category: "SOCIAL"})
	registerAchievementDef(AchievementDef{ID: "party-hero", Name: "Party Hero", Title: "party-hero", Flag: "ach_party_hero", Category: "SOCIAL"})
	registerAchievementDef(AchievementDef{ID: "guild-member", Name: "Guild Member", Title: "guild-member", Flag: "ach_guild_member", Category: "SOCIAL"})
	registerAchievementDef(AchievementDef{ID: "guild-founder", Name: "Guild Founder", Title: "guild-founder", Flag: "ach_guild_founder", Category: "SOCIAL"})
	registerAchievementDef(AchievementDef{ID: "helpful-player", Name: "Helpful Player", Title: "helpful-player", Flag: "ach_helpful_player", Category: "SOCIAL"})
	registerAchievementDef(AchievementDef{ID: "community-hero", Name: "Community Hero", Title: "community-hero", Flag: "ach_community_hero", Category: "SOCIAL"})
	registerTitleDef(TitleDef{ID: "first-friend", Name: "Sahabat Pertama", Source: "social", Rarity: "COMMON"})
	registerTitleDef(TitleDef{ID: "party-hero", Name: "Pahlawan Party", Source: "social", Rarity: "UNCOMMON"})
	registerTitleDef(TitleDef{ID: "guild-member", Name: "Anggota Guild", Source: "social", Rarity: "COMMON"})
	registerTitleDef(TitleDef{ID: "guild-founder", Name: "Pendiri Guild", Source: "social", Rarity: "RARE"})
	registerTitleDef(TitleDef{ID: "helpful-player", Name: "Pemain Penolong", Source: "social", Rarity: "UNCOMMON"})
	registerTitleDef(TitleDef{ID: "community-hero", Name: "Pahlawan Komunitas", Source: "social", Rarity: "RARE"})
	registerCosmeticDef(CosmeticDef{ID: "badge-friend", Name: "Lencana Sahabat", Kind: "badge"})
	registerCosmeticDef(CosmeticDef{ID: "badge-guild", Name: "Lencana Guild", Kind: "badge"})
	registerCosmeticDef(CosmeticDef{ID: "banner-desa", Name: "Banner Desa", Kind: "banner"})
	registerItem(ItemDef{
		ID: "relic-soul-dawn", Name: "Relik Jiwa Fajar", Type: "COSMETIC", Rarity: "RARE",
		Icon: "relic", Value: 0, Untradable: true, Description: "Soulbound. Tidak dapat ditrade.",
	})
	registerInteract(InteractDef{ID: "guild-board", Kind: "guild-board", X: -5.0, Z: 4.4, Text: "GUILD BOARD"})
	registerNPC(NPCDef{
		ID: "mbah_balai", Name: "Mbah Balai", Role: "Penjaga Balai", Type: "GUILD",
		X: -5.4, Z: 4.2, Yaw: 1.4, DialogueID: "mbah_balai", InteractionRange: 2.6,
	})
	if _, ok := dialogueCatalog["mbah_balai"]; !ok {
		dialogueCatalog["mbah_balai"] = DialogueLine{Speaker: "Mbah Balai", Text: "Le, iki balai guild. Rapat, latihan, lan papan kabar ana kene."}
	}
}

func registerGuildQuest(q GuildQuestDef) {
	for _, e := range guildQuestList {
		if e.ID == q.ID {
			return
		}
	}
	guildQuestList = append(guildQuestList, q)
}

func guildQuestByID(id string) *GuildQuestDef {
	for i := range guildQuestList {
		if guildQuestList[i].ID == id {
			return &guildQuestList[i]
		}
	}
	return nil
}

func normalizePrivacy(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case PrivacyFriends:
		return PrivacyFriends
	case PrivacyNone:
		return PrivacyNone
	default:
		return PrivacyEveryone
	}
}

func (w *WorldState) friendsOf(a, b string) bool {
	if w.Social == nil || a == "" || b == "" {
		return false
	}
	if w.Social.Friends[a] == nil {
		return false
	}
	_, ok := w.Social.Friends[a][b]
	return ok
}

func (w *WorldState) privacyAllows(setting, fromID, toID string) bool {
	switch normalizePrivacy(setting) {
	case PrivacyNone:
		return false
	case PrivacyFriends:
		return w.friendsOf(toID, fromID)
	default:
		return true
	}
}

func (log *PlayerLog) privacyFriend() string { return normalizePrivacy(log.PrivacyFriend) }
func (log *PlayerLog) privacyParty() string  { return normalizePrivacy(log.PrivacyParty) }
func (log *PlayerLog) privacyTrade() string  { return normalizePrivacy(log.PrivacyTrade) }
func (log *PlayerLog) privacyPM() string     { return normalizePrivacy(log.PrivacyPM) }

func guildDisplayRank(rank string) string {
	if strings.EqualFold(rank, "LEADER") {
		return "MASTER"
	}
	return rank
}

func phase26StripLinks(s string) string {
	low := strings.ToLower(s)
	if !strings.Contains(low, "http://") && !strings.Contains(low, "https://") && !strings.Contains(low, "www.") {
		return s
	}
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); {
		rest := strings.ToLower(string(runes[i:]))
		if strings.HasPrefix(rest, "https://") || strings.HasPrefix(rest, "http://") || strings.HasPrefix(rest, "www.") {
			b.WriteString("***")
			for i < len(runes) && !unicode.IsSpace(runes[i]) {
				i++
			}
			continue
		}
		b.WriteRune(runes[i])
		i++
	}
	return b.String()
}

func notifyPriority(kind string) string {
	switch kind {
	case "friend_request", "party_invite", "guild_invite", "trade_request":
		return "IMPORTANT"
	case "friend", "party", "party_join", "achievement", "guild":
		return "NORMAL"
	default:
		return "LOW"
	}
}

func (w *WorldState) phase26RestoreSocial(p *Player) {
	if p == nil {
		return
	}
	log := p.ensureLog()
	log.PresenceMode = "ONLINE"
	if w.Parties != nil {
		if id := w.Parties.byPlayer[p.ID]; id != "" {
			p.PartyID = id
			if pt := w.Parties.Get(id); pt != nil && pt.OfflineSince != nil {
				delete(pt.OfflineSince, p.ID)
			}
		}
	}
	if w.Guilds != nil {
		if gid := w.Guilds.ByPlayer[p.ID]; gid != "" {
			p.GuildID = gid
			if g := w.Guilds.ByID[gid]; g != nil {
				p.GuildTag = g.Tag
			}
		}
	}
}

func (w *WorldState) phase26MarkPartyOffline(id string) {
	if w.Parties == nil {
		return
	}
	pt := w.Parties.Of(id)
	if pt == nil {
		return
	}
	if pt.OfflineSince == nil {
		pt.OfflineSince = map[string]time.Time{}
	}
	pt.OfflineSince[id] = time.Now()
}

func (w *WorldState) phase26GuildQuestComplete(p *Player, q GuildQuestDef) {
	if p == nil {
		return
	}
	if q.ID == Phase26GuildQuestID {
		p.ensureLog().GuildTokens += 8
		p.markDirty()
	}
}

func (w *WorldState) grantSocialFlag(p *Player, flag string) [][]byte {
	if p == nil {
		return nil
	}
	log := p.ensureLog()
	if log.Flags == nil {
		log.Flags = map[string]bool{}
	}
	log.Flags[flag] = true
	if flag == "ach_helpful_player" {
		log.HelpScore++
	}
	p.markDirty()
	return w.refreshAchievements(p)
}

func (w *WorldState) listenerMutedSpeaker(listener, speaker string) bool {
	p := w.players[listener]
	if p == nil {
		return false
	}
	log := p.ensureLog()
	return log.LocalMutes != nil && log.LocalMutes[speaker]
}

func (w *WorldState) filterMutedNotify(fromID string, ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == fromID || !w.listenerMutedSpeaker(id, fromID) {
			out = append(out, id)
		}
	}
	return out
}

func (w *WorldState) regionLabel(p *Player) string {
	if p == nil {
		return ""
	}
	z := zoneAt(p.X, p.Z)
	if z.Title != "" {
		return z.Title
	}
	if z.Name != "" {
		return z.Name
	}
	return p.Region
}

func (w *WorldState) privacyView(p *Player) PrivacyView {
	log := p.ensureLog()
	return PrivacyView{
		Friend: log.privacyFriend(),
		Party:  log.privacyParty(),
		Trade:  log.privacyTrade(),
		PM:     log.privacyPM(),
	}
}

func (w *WorldState) ApplyPhase26(p *Player, env Envelope) [][]byte {
	if p == nil {
		return nil
	}
	switch env.Type {
	case TypeSetPrivacy:
		return w.setPrivacy(p, env.Data)
	case TypeGetPrivacy:
		return [][]byte{marshal(TypePrivacyUpdated, w.privacyView(p))}
	case TypeSetPresence:
		var in struct{ Status string }
		_ = unmarshal(env.Data, &in)
		st := strings.ToUpper(strings.TrimSpace(in.Status))
		switch st {
		case "ONLINE", "AWAY", "BUSY":
			p.ensureLog().PresenceMode = st
			p.LastHeard = time.Now()
			p.markDirty()
			return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
		default:
			return rejectFor(p.ID, TypeSetPresence, "status")
		}
	case TypeLocalMute:
		var in struct{ TargetID string }
		_ = unmarshal(env.Data, &in)
		if in.TargetID == "" || in.TargetID == p.ID {
			return rejectFor(p.ID, TypeLocalMute, "target")
		}
		log := p.ensureLog()
		if log.LocalMutes == nil {
			log.LocalMutes = map[string]bool{}
		}
		log.LocalMutes[in.TargetID] = true
		p.markDirty()
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	case TypeUnmuteLocal:
		var in struct{ TargetID string }
		_ = unmarshal(env.Data, &in)
		log := p.ensureLog()
		if log.LocalMutes != nil {
			delete(log.LocalMutes, in.TargetID)
			p.markDirty()
		}
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	case TypeReportMessage:
		return w.reportMessage(p, env.Data)
	case TypePartyCreate:
		return w.partyCreate(p)
	case TypeSetFriendship, TypeSetPartyLeader, TypeSetTradeItem, TypeSetChatSender, TypeSetGuildMaster:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) setPrivacy(p *Player, data []byte) [][]byte {
	var in struct {
		Friend, Party, Trade, PM string
	}
	_ = unmarshal(data, &in)
	log := p.ensureLog()
	if in.Friend != "" {
		log.PrivacyFriend = normalizePrivacy(in.Friend)
	}
	if in.Party != "" {
		log.PrivacyParty = normalizePrivacy(in.Party)
	}
	if in.Trade != "" {
		log.PrivacyTrade = normalizePrivacy(in.Trade)
	}
	if in.PM != "" {
		log.PrivacyPM = normalizePrivacy(in.PM)
	}
	p.markDirty()
	return [][]byte{marshal(TypePrivacyUpdated, w.privacyView(p))}
}

func (w *WorldState) reportMessage(p *Player, data []byte) [][]byte {
	var in struct {
		TargetID, Category, Evidence, Channel string
	}
	_ = unmarshal(data, &in)
	if in.Category == "" {
		in.Category = "SPAM"
	}
	if !reportCategoryOK(in.Category) {
		return rejectFor(p.ID, TypeReportMessage, "category")
	}
	if w.limited("repmsg:"+p.ID, 8, time.Hour) {
		return rejectFor(p.ID, TypeReportMessage, "rate")
	}
	if w.Reports == nil {
		w.Reports = []Report{}
	}
	w.Reports = append(w.Reports, Report{
		ID: randomID("rep_"), Reporter: p.ID, Target: in.TargetID, Category: strings.ToUpper(in.Category),
		Evidence: in.Evidence, State: "OPEN", MatchID: in.Channel, At: time.Now(),
	})
	return [][]byte{w.notify(p, "system", "Laporan pesan diterima.")}
}

func (w *WorldState) partyCreate(p *Player) [][]byte {
	if w.Parties == nil {
		w.Parties = NewPartyHub()
	}
	if w.Parties.Of(p.ID) != nil {
		return rejectFor(p.ID, TypePartyCreate, "already_in_party")
	}
	pt := &Party{ID: randomID("pty_"), LeaderID: p.ID, Members: []string{p.ID}, CreatedAt: time.Now(), OfflineSince: map[string]time.Time{}, Ready: map[string]bool{}}
	w.Parties.parties[pt.ID] = pt
	w.Parties.byPlayer[p.ID] = pt.ID
	p.PartyID = pt.ID
	out := [][]byte{marshal(TypePartyUpdated, w.partyView(pt, p))}
	out = append(out, w.grantSocialFlag(p, "ach_party_hero")...)
	return out
}

func (w *WorldState) displayNameTaken(name, exceptID string) bool {
	low := strings.ToLower(strings.TrimSpace(name))
	if low == "" {
		return false
	}
	for _, o := range w.players {
		if o.ID == exceptID {
			continue
		}
		if strings.ToLower(o.Name) == low {
			return true
		}
	}
	return false
}
