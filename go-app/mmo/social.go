package mmo

import (
	"math"
	"time"
)

type FriendRel struct {
	Since    time.Time
	LastSeen int64
}

type SocialHub struct {
	Friends  map[string]map[string]FriendRel
	Incoming map[string]map[string]bool
	Outgoing map[string]map[string]bool
	Blocked  map[string]map[string]bool
}

func NewSocialHub() *SocialHub {
	return &SocialHub{
		Friends:  map[string]map[string]FriendRel{},
		Incoming: map[string]map[string]bool{},
		Outgoing: map[string]map[string]bool{},
		Blocked:  map[string]map[string]bool{},
	}
}

func (s *SocialHub) blocked(a, b string) bool {
	if s == nil {
		return false
	}
	return s.Blocked[a][b] || s.Blocked[b][a]
}

func ensureSet(m map[string]map[string]bool, id string) map[string]bool {
	if m[id] == nil {
		m[id] = map[string]bool{}
	}
	return m[id]
}

func (w *WorldState) ApplySocial(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	if w.Social == nil {
		w.Social = NewSocialHub()
	}
	var in PartyActionIn
	_ = unmarshal(env.Data, &in)
	switch env.Type {
	case TypeChat:
		return w.ApplyChat(id, env)
	case TypeGetSocial, TypeNearbyPlayers:
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	case TypeInspectPlayer:
		return w.inspect(p, in.TargetID)
	case TypeFriendRequest:
		return w.friendRequest(p, in.TargetID)
	case TypeAcceptFriend:
		return w.acceptFriend(p, in.TargetID)
	case TypeDeclineFriend:
		delete(ensureSet(w.Social.Incoming, p.ID), in.TargetID)
		if w.Social.Outgoing[in.TargetID] != nil {
			delete(w.Social.Outgoing[in.TargetID], p.ID)
		}
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	case TypeRemoveFriend:
		delete(w.Social.Friends[p.ID], in.TargetID)
		delete(w.Social.Friends[in.TargetID], p.ID)
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	case TypeBlockPlayer:
		if in.TargetID == "" || in.TargetID == p.ID {
			return rejectFor(p.ID, TypeBlockPlayer, "target")
		}
		lim := socialCfg.BlockLimit
		if lim < 1 {
			lim = Phase26BlockLimit
		}
		if len(w.Social.Blocked[p.ID]) >= lim && !w.Social.Blocked[p.ID][in.TargetID] {
			return rejectFor(p.ID, TypeBlockPlayer, "limit")
		}
		ensureSet(w.Social.Blocked, p.ID)[in.TargetID] = true
		delete(w.Social.Friends[p.ID], in.TargetID)
		delete(w.Social.Friends[in.TargetID], p.ID)
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	case TypeUnblockPlayer:
		delete(ensureSet(w.Social.Blocked, p.ID), in.TargetID)
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	case TypeSetPrivacy, TypeGetPrivacy, TypeSetPresence, TypeLocalMute, TypeUnmuteLocal, TypeReportMessage,
		TypeSetFriendship, TypeSetPartyLeader, TypeSetTradeItem, TypeSetChatSender, TypeSetGuildMaster:
		return w.ApplyPhase26(p, env)
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) friendRequest(from *Player, targetID string) [][]byte {
	if w.limited("fr:"+from.ID, 8, time.Minute) {
		return rejectFor(from.ID, TypeFriendRequest, "rate")
	}
	to := w.players[targetID]
	if to == nil || !to.Connected {
		return rejectFor(from.ID, TypeFriendRequest, "target")
	}
	if from.ID == to.ID {
		return rejectFor(from.ID, TypeFriendRequest, "self")
	}
	if w.Social.blocked(from.ID, to.ID) {
		return rejectFor(from.ID, TypeFriendRequest, "blocked")
	}
	if !w.privacyAllows(to.ensureLog().privacyFriend(), from.ID, to.ID) {
		return rejectFor(from.ID, TypeFriendRequest, "privacy")
	}
	if friendCount(w.Social, from.ID) >= socialCfg.FriendLimit || friendCount(w.Social, to.ID) >= socialCfg.FriendLimit {
		return rejectFor(from.ID, TypeFriendRequest, "limit")
	}
	if w.Social.Friends[from.ID] != nil {
		if _, ok := w.Social.Friends[from.ID][to.ID]; ok {
			return rejectFor(from.ID, TypeFriendRequest, "already")
		}
	}
	ensureSet(w.Social.Outgoing, from.ID)[to.ID] = true
	ensureSet(w.Social.Incoming, to.ID)[from.ID] = true
	return [][]byte{
		marshal(TypeFriendRequest, map[string]any{"fromId": from.ID, "from": from.Name, "toId": to.ID}),
		marshal(TypeSocialNotification, SocialNote{Kind: "friend_request", Text: from.Name + " sent you a friend request.", FromID: from.ID, From: from.Name, ToID: to.ID, Priority: "IMPORTANT"}),
	}
}

func (w *WorldState) acceptFriend(p *Player, fromID string) [][]byte {
	if !w.Social.Incoming[p.ID][fromID] {
		return rejectFor(p.ID, TypeAcceptFriend, "request")
	}
	if friendCount(w.Social, p.ID) >= socialCfg.FriendLimit || friendCount(w.Social, fromID) >= socialCfg.FriendLimit {
		return rejectFor(p.ID, TypeAcceptFriend, "limit")
	}
	delete(w.Social.Incoming[p.ID], fromID)
	if w.Social.Outgoing[fromID] != nil {
		delete(w.Social.Outgoing[fromID], p.ID)
	}
	if w.Social.Friends[p.ID] == nil {
		w.Social.Friends[p.ID] = map[string]FriendRel{}
	}
	if w.Social.Friends[fromID] == nil {
		w.Social.Friends[fromID] = map[string]FriendRel{}
	}
	now := FriendRel{Since: time.Now(), LastSeen: time.Now().UnixMilli()}
	w.Social.Friends[p.ID][fromID] = now
	w.Social.Friends[fromID][p.ID] = now
	other := w.players[fromID]
	out := [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	out = append(out, w.grantSocialFlag(p, "ach_first_friend")...)
	if other != nil {
		st := w.socialState(other)
		st.ToID = fromID
		out = append(out, marshal(TypeFriendUpdated, st))
		out = append(out, marshal(TypeSocialNotification, SocialNote{Kind: "friend", Text: p.Name + " accepted your friend request.", From: p.Name, ToID: fromID, Priority: "NORMAL"}))
		out = append(out, w.grantSocialFlag(other, "ach_first_friend")...)
	}
	return out
}

func (w *WorldState) inspect(p *Player, targetID string) [][]byte {
	t := w.players[targetID]
	if t == nil {
		return rejectFor(p.ID, TypeInspectPlayer, "target")
	}
	t.ensureGear()
	w.refreshTitles(t)
	gname, gtag := "", t.GuildTag
	if g := w.guildOf(t.ID); g != nil {
		gname, gtag = g.Name, g.Tag
	}
	return [][]byte{marshal(TypeInspectResult, InspectOut{
		PlayerID: t.ID, Name: t.Name, Level: t.Level, Class: t.Class,
		Stats: t.statsView(), Equipment: t.Gear.view(), PowerRating: t.powerRating(),
		Guild: gname, GuildTag: gtag, Title: t.Title,
		Rank: t.ensureLog().PvpHighestRank, Season: seasonTrack.Name, SeasonLevel: t.ensureLog().SeasonLevel,
		Badge: t.ensureLog().ShowcaseBadge, Aura: t.ensureLog().ShowcaseAura, Mount: t.ensureLog().ShowcaseMount,
		Avatar: t.Class, Status: w.presenceOf(t), Region: w.regionLabel(t),
		AchievementScore: len(t.ensureLog().Achievements),
	})}
}

func (w *WorldState) socialState(p *Player) SocialState {
	st := SocialState{Friends: []FriendView{}, Pending: []FriendView{}, Outgoing: []FriendView{}, Blocked: []string{}, Nearby: []NearbyView{}}
	if pt := w.Parties.Of(p.ID); pt != nil {
		v := w.partyView(pt, p)
		st.Party = &v
	}
	for id := range w.Social.Friends[p.ID] {
		st.Friends = append(st.Friends, w.friendView(id))
	}
	for id := range w.Social.Incoming[p.ID] {
		st.Pending = append(st.Pending, w.friendView(id))
	}
	for id := range w.Social.Outgoing[p.ID] {
		st.Outgoing = append(st.Outgoing, w.friendView(id))
	}
	for id := range w.Social.Blocked[p.ID] {
		st.Blocked = append(st.Blocked, id)
	}
	for _, o := range w.players {
		if o.ID == p.ID || !o.Connected {
			continue
		}
		d := math.Hypot(p.X-o.X, p.Z-o.Z)
		if d <= NearbyRange {
			st.Nearby = append(st.Nearby, NearbyView{PlayerID: o.ID, Name: o.Name, Level: o.Level, Class: o.Class, Distance: d, Online: true})
		}
	}
	st.Notifies = p.ensureLog().Notices
	g := w.guildViewFor(p)
	if g.ID != "" {
		st.Guild = &g
	}
	st.Wallet = p.wallet()
	pv := w.privacyView(p)
	st.Privacy = &pv
	return st
}

func (w *WorldState) friendView(id string) FriendView {
	p := w.players[id]
	if p != nil {
		return FriendView{PlayerID: p.ID, Name: p.Name, Level: p.Level, Class: p.Class, Online: p.Connected, LastSeen: time.Now().UnixMilli(), Status: w.presenceOf(p), Guild: p.GuildTag, Title: p.Title, Avatar: p.Class, Region: w.regionLabel(p)}
	}
	return FriendView{PlayerID: id, Name: id, Online: false, Status: "OFFLINE"}
}

func (w *WorldState) LoadoutFor(id string) InventoryUpdated {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil {
		return InventoryUpdated{}
	}
	p.ensureGear()
	p.applyDerived()
	return p.loadout("", nil)
}

func (w *WorldState) SocialFor(id string) SocialState {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil {
		return SocialState{}
	}
	if w.Social == nil {
		w.Social = NewSocialHub()
	}
	if w.Parties == nil {
		w.Parties = NewPartyHub()
	}
	return w.socialState(p)
}
