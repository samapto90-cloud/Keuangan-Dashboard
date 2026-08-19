package mmo

import (
	"time"
)

type Party struct {
	ID, LeaderID string
	Members      []string
	CreatedAt    time.Time
	TargetID     string
	OfflineSince map[string]time.Time
	Activity     string
	MinLevel     int
	NeedRole     string
	Roles        map[string]string
	Ready        map[string]bool
	WaypointID   string
	WaypointX    float64
	WaypointZ    float64
}

type PartyInvite struct {
	PartyID, FromID, ToID string
	Until                 time.Time
}

type PartyHub struct {
	parties  map[string]*Party
	byPlayer map[string]string
	invites  map[string]*PartyInvite
}

type FinderHub struct {
	List []*FinderListing
}

func NewFinderHub() *FinderHub {
	return &FinderHub{}
}

func NewPartyHub() *PartyHub {
	return &PartyHub{parties: map[string]*Party{}, byPlayer: map[string]string{}, invites: map[string]*PartyInvite{}}
}

func (h *PartyHub) Get(id string) *Party {
	if h == nil {
		return nil
	}
	return h.parties[id]
}

func (h *PartyHub) Of(playerID string) *Party {
	if h == nil {
		return nil
	}
	return h.parties[h.byPlayer[playerID]]
}

func (w *WorldState) sameParty(a, b string) bool {
	if w.Parties == nil || a == "" || b == "" {
		return false
	}
	pa, pb := w.Parties.byPlayer[a], w.Parties.byPlayer[b]
	return pa != "" && pa == pb
}

func (w *WorldState) ApplyParty(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	if w.Parties == nil {
		w.Parties = NewPartyHub()
	}
	var in PartyActionIn
	_ = unmarshal(env.Data, &in)
	switch env.Type {
	case TypePartyCreate:
		return w.partyCreate(p)
	case TypePartyInvite:
		return w.partyInvite(p, in.TargetID)
	case TypePartyAccept:
		return w.partyAccept(p)
	case TypePartyDecline:
		delete(w.Parties.invites, p.ID)
		return [][]byte{marshal(TypeSocialNotification, SocialNote{Kind: "party", Text: "Undangan ditolak."})}
	case TypePartyLeave:
		return w.partyLeave(p, false)
	case TypePartyKick:
		return w.partyKick(p, in.TargetID)
	case TypePartyDisband:
		return w.partyDisband(p)
	case TypePartySetTarget:
		return w.partySetTarget(p, in.TargetID)
	case TypePartyTransfer:
		return w.partyTransfer(p, in.TargetID)
	case TypePartySetRole:
		return w.partySetRole(p, in.Role)
	case TypePartyReady:
		return w.partyReady(p)
	case TypePartyFinderCreate:
		return w.finderCreate(p, in.Activity, in.Role, in.MinLevel)
	case TypePartyFinderList:
		return w.finderList(p)
	case TypePartyFinderJoin:
		return w.finderJoin(p, in.ListingID)
	case TypePartySetWaypoint:
		return w.partySetWaypoint(p, in.LandmarkID)
	case TypeFollowParty:
		return w.followParty(p)
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) partyInvite(from *Player, targetID string) [][]byte {
	to := w.players[targetID]
	if to == nil || !to.Connected {
		return rejectFor(from.ID, TypePartyInvite, "target")
	}
	if from.ID == to.ID {
		return rejectFor(from.ID, TypePartyInvite, "self")
	}
	if w.Social != nil && w.Social.blocked(from.ID, to.ID) {
		return rejectFor(from.ID, TypePartyInvite, "blocked")
	}
	if !w.privacyAllows(to.ensureLog().privacyParty(), from.ID, to.ID) {
		return rejectFor(from.ID, TypePartyInvite, "privacy")
	}
	if w.Parties.byPlayer[to.ID] != "" {
		return rejectFor(from.ID, TypePartyInvite, "already_in_party")
	}
	pt := w.Parties.Of(from.ID)
	if pt == nil {
		pt = &Party{ID: randomID("pty_"), LeaderID: from.ID, Members: []string{from.ID}, CreatedAt: time.Now(), OfflineSince: map[string]time.Time{}}
		w.Parties.parties[pt.ID] = pt
		w.Parties.byPlayer[from.ID] = pt.ID
		from.PartyID = pt.ID
	}
	if pt.LeaderID != from.ID {
		return rejectFor(from.ID, TypePartyInvite, "not_leader")
	}
	if len(pt.Members) >= partyCap() {
		return rejectFor(from.ID, TypePartyInvite, "full")
	}
	sec := socialCfg.PartyInviteSec
	if sec < 1 {
		sec = 30
	}
	w.Parties.invites[to.ID] = &PartyInvite{PartyID: pt.ID, FromID: from.ID, ToID: to.ID, Until: time.Now().Add(time.Duration(sec) * time.Second)}
	return [][]byte{
		marshal(TypePartyInvite, map[string]any{"partyId": pt.ID, "fromId": from.ID, "from": from.Name, "toId": to.ID}),
		marshal(TypeSocialNotification, SocialNote{Kind: "party_invite", Text: from.Name + " invited you to a party.", FromID: from.ID, From: from.Name, ToID: to.ID, Priority: "IMPORTANT"}),
	}
}

func (w *WorldState) partyAccept(p *Player) [][]byte {
	inv := w.Parties.invites[p.ID]
	if inv == nil || time.Now().After(inv.Until) {
		return rejectFor(p.ID, TypePartyAccept, "invite")
	}
	pt := w.Parties.Get(inv.PartyID)
	if pt == nil {
		return rejectFor(p.ID, TypePartyAccept, "party")
	}
	if w.Parties.byPlayer[p.ID] != "" {
		return rejectFor(p.ID, TypePartyAccept, "already_in_party")
	}
	if len(pt.Members) >= partyCap() {
		return rejectFor(p.ID, TypePartyAccept, "full")
	}
	delete(w.Parties.invites, p.ID)
	pt.Members = append(pt.Members, p.ID)
	w.Parties.byPlayer[p.ID] = pt.ID
	p.PartyID = pt.ID
	view := w.partyView(pt, p)
	out := [][]byte{
		marshal(TypePartyMemberJoined, view),
		marshal(TypePartyUpdated, view),
		marshal(TypeSocialNotification, SocialNote{Kind: "party_join", Text: p.Name + " joined your party.", From: p.Name, Priority: "NORMAL"}),
	}
	out = append(out, w.grantSocialFlag(p, "ach_party_hero")...)
	return out
}

func (w *WorldState) partyLeave(p *Player, kicked bool) [][]byte {
	pt := w.Parties.Of(p.ID)
	if pt == nil {
		return rejectFor(p.ID, TypePartyLeave, "not_in_party")
	}
	pt.Members = removeID(pt.Members, p.ID)
	delete(w.Parties.byPlayer, p.ID)
	p.PartyID = ""
	if len(pt.Members) == 0 {
		delete(w.Parties.parties, pt.ID)
		return [][]byte{marshal(TypePartyUpdated, PartyView{NotifyIDs: []string{p.ID}})}
	}
	if pt.LeaderID == p.ID {
		pt.LeaderID = pt.Members[0]
	}
	view := w.partyView(pt, p)
	kind := TypePartyMemberLeft
	text := p.Name + " left the party."
	if kicked {
		text = p.Name + " was removed from the party."
	}
	return [][]byte{marshal(kind, view), marshal(TypePartyUpdated, view), marshal(TypeSocialNotification, SocialNote{Kind: "party_leave", Text: text})}
}

func (w *WorldState) partyKick(leader *Player, targetID string) [][]byte {
	pt := w.Parties.Of(leader.ID)
	if pt == nil || pt.LeaderID != leader.ID {
		return rejectFor(leader.ID, TypePartyKick, "not_leader")
	}
	if targetID == leader.ID {
		return rejectFor(leader.ID, TypePartyKick, "self")
	}
	t := w.players[targetID]
	if t == nil {
		pt.Members = removeID(pt.Members, targetID)
		delete(w.Parties.byPlayer, targetID)
		return [][]byte{marshal(TypePartyUpdated, w.partyView(pt, leader))}
	}
	return w.partyLeave(t, true)
}

func (w *WorldState) partyDisband(p *Player) [][]byte {
	pt := w.Parties.Of(p.ID)
	if pt == nil || pt.LeaderID != p.ID {
		return rejectFor(p.ID, TypePartyDisband, "not_leader")
	}
	ids := append([]string{}, pt.Members...)
	for _, id := range ids {
		delete(w.Parties.byPlayer, id)
		if m := w.players[id]; m != nil {
			m.PartyID = ""
		}
	}
	delete(w.Parties.parties, pt.ID)
	return [][]byte{marshal(TypePartyUpdated, PartyView{NotifyIDs: ids})}
}

func (w *WorldState) partySetTarget(p *Player, targetID string) [][]byte {
	pt := w.Parties.Of(p.ID)
	if pt == nil || pt.LeaderID != p.ID {
		return rejectFor(p.ID, TypePartySetTarget, "not_leader")
	}
	pt.TargetID = targetID
	return [][]byte{marshal(TypePartyUpdated, w.partyView(pt, p))}
}

func (w *WorldState) partyView(pt *Party, watcher *Player) PartyView {
	if pt == nil {
		return PartyView{}
	}
	members := make([]PartyMemberView, 0, len(pt.Members))
	for _, id := range pt.Members {
		m := w.players[id]
		online := m != nil && m.Connected
		name, class := id, "WARRIOR"
		lv, hp, max, en, maxE := 1, 0, 100, 0, 100
		dist := 0.0
		if m != nil {
			name, class, lv = m.Name, m.Class, m.Level
			hp, max, en, maxE = m.HP, m.MaxHP, m.Energy, m.MaxEnergy
			if watcher != nil {
				dist = mathHypot(watcher.X-m.X, watcher.Z-m.Z)
			}
		}
		role := "FLEX"
		if pt.Roles != nil && pt.Roles[id] != "" {
			role = pt.Roles[id]
		}
		members = append(members, PartyMemberView{
			PlayerID: id, Name: name, Level: lv, Class: class, HP: hp, MaxHP: max,
			Energy: en, MaxEnergy: maxE, Distance: dist, Online: online, Role: role, Leader: id == pt.LeaderID,
			Ready: pt.Ready[id], Status: w.presenceOf(m),
		})
	}
	state := "WAITING"
	if len(pt.Members) == 0 {
		state = "DISBANDED"
	} else if watcher != nil && watcher.InstanceID != "" && !houseID(watcher.InstanceID) {
		state = "ACTIVE"
	}
	view := PartyView{PartyID: pt.ID, LeaderID: pt.LeaderID, Members: members, Activity: pt.Activity, MinLevel: pt.MinLevel, NeedRole: pt.NeedRole, State: state, Ready: pt.Ready, NotifyIDs: append([]string{}, pt.Members...), WaypointID: pt.WaypointID, WaypointX: pt.WaypointX, WaypointZ: pt.WaypointZ}
	if e := w.enemies[pt.TargetID]; e != nil && e.Alive {
		view.TargetID, view.TargetName, view.TargetHP, view.TargetMax, view.TargetLv = e.ID, e.Def.Name, e.HP, e.MaxHP, e.Def.Level
	}
	return view
}

func (w *WorldState) fanoutParty(pt *Party, events [][]byte) {
	if pt == nil {
		return
	}
	for _, id := range pt.Members {
		for _, ev := range events {
			w.SendTo(id, ev)
		}
	}
}

func removeID(ids []string, drop string) []string {
	out := ids[:0]
	for _, id := range ids {
		if id != drop {
			out = append(out, id)
		}
	}
	return append([]string{}, out...)
}

func mathHypot(a, b float64) float64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	if a > b {
		return a + b*0.4
	}
	return b + a*0.4
}

func (w *WorldState) partyTransfer(p *Player, targetID string) [][]byte {
	pt := w.Parties.Of(p.ID)
	if pt == nil || pt.LeaderID != p.ID {
		return rejectFor(p.ID, TypePartyTransfer, "not_leader")
	}
	ok := false
	for _, id := range pt.Members {
		if id == targetID {
			ok = true
		}
	}
	if !ok {
		return rejectFor(p.ID, TypePartyTransfer, "member")
	}
	pt.LeaderID = targetID
	return [][]byte{marshal(TypePartyUpdated, w.partyView(pt, p))}
}

func (w *WorldState) partySetRole(p *Player, role string) [][]byte {
	pt := w.Parties.Of(p.ID)
	if pt == nil {
		return rejectFor(p.ID, TypePartySetRole, "party")
	}
	switch role {
	case "DPS", "TANK", "SUPPORT", "FLEX":
	default:
		role = "FLEX"
	}
	if pt.Roles == nil {
		pt.Roles = map[string]string{}
	}
	pt.Roles[p.ID] = role
	return [][]byte{marshal(TypePartyUpdated, w.partyView(pt, p))}
}

func (w *WorldState) partyReady(p *Player) [][]byte {
	pt := w.Parties.Of(p.ID)
	if pt == nil {
		return rejectFor(p.ID, TypePartyReady, "party")
	}
	if pt.Ready == nil {
		pt.Ready = map[string]bool{}
	}
	pt.Ready[p.ID] = true
	return [][]byte{marshal(TypePartyUpdated, w.partyView(pt, p))}
}

func (w *WorldState) finderCreate(p *Player, activity, role string, minLv int) [][]byte {
	if w.Finder == nil {
		w.Finder = NewFinderHub()
	}
	pt := w.Parties.Of(p.ID)
	if pt == nil {
		pt = &Party{ID: randomID("pty_"), LeaderID: p.ID, Members: []string{p.ID}, CreatedAt: time.Now(), OfflineSince: map[string]time.Time{}, Roles: map[string]string{}, Ready: map[string]bool{}}
		w.Parties.parties[pt.ID] = pt
		w.Parties.byPlayer[p.ID] = pt.ID
		p.PartyID = pt.ID
	}
	if pt.LeaderID != p.ID {
		return rejectFor(p.ID, TypePartyFinderCreate, "not_leader")
	}
	pt.Activity, pt.NeedRole, pt.MinLevel = activity, role, minLv
	w.Finder.List = append(w.Finder.List, &FinderListing{ID: pt.ID, Leader: p.Name, Activity: activity, Level: p.Level, Role: role, Players: len(pt.Members), Cap: PartyMaxSize})
	return w.finderList(p)
}

func (w *WorldState) finderList(p *Player) [][]byte {
	if w.Finder == nil {
		w.Finder = NewFinderHub()
	}
	live := []*FinderListing{}
	for _, pt := range w.Parties.parties {
		if pt.Activity == "" {
			continue
		}
		name := pt.LeaderID
		if l := w.players[pt.LeaderID]; l != nil {
			name = l.Name
		}
		live = append(live, &FinderListing{ID: pt.ID, Leader: name, Activity: pt.Activity, Level: pt.MinLevel, Role: pt.NeedRole, Players: len(pt.Members), Cap: PartyMaxSize})
	}
	w.Finder.List = live
	return [][]byte{marshal(TypePartyFinderList, map[string]any{"listings": live})}
}

func (w *WorldState) finderJoin(p *Player, listingID string) [][]byte {
	pt := w.Parties.Get(listingID)
	if pt == nil || pt.Activity == "" {
		return rejectFor(p.ID, TypePartyFinderJoin, "listing")
	}
	if p.Level < pt.MinLevel {
		return rejectFor(p.ID, TypePartyFinderJoin, "level")
	}
	if w.Social != nil && w.Social.blocked(p.ID, pt.LeaderID) {
		return rejectFor(p.ID, TypePartyFinderJoin, "blocked")
	}
	if w.Parties.byPlayer[p.ID] != "" {
		return rejectFor(p.ID, TypePartyFinderJoin, "already_in_party")
	}
	if len(pt.Members) >= PartyMaxSize {
		return rejectFor(p.ID, TypePartyFinderJoin, "full")
	}
	pt.Members = append(pt.Members, p.ID)
	w.Parties.byPlayer[p.ID] = pt.ID
	p.PartyID = pt.ID
	return [][]byte{marshal(TypePartyMemberJoined, w.partyView(pt, p)), marshal(TypePartyUpdated, w.partyView(pt, p))}
}
