package mmo

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func socialStore() *SocialStore {
	if DefaultHub != nil && DefaultHub.Social != nil {
		return DefaultHub.Social
	}
	return nil
}

func requireSocial(w http.ResponseWriter, r *http.Request) (*GameSession, *SocialStore, bool) {
	sess, _, ok := requireProgressSession(w, r)
	if !ok {
		return nil, nil, false
	}
	st := socialStore()
	if st == nil {
		writeAccountErr(w, http.StatusServiceUnavailable, "sosial tidak siap")
		return nil, nil, false
	}
	return sess, st, true
}

func HandleFriends(w http.ResponseWriter, r *http.Request) {
	sess, st, ok := requireSocial(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	ids := st.FriendIDs(sess.PlayerID)
	in, outg := st.PendingFor(sess.PlayerID)
	friends := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		friends = append(friends, friendCard(sess.PlayerID, id, st))
	}
	incoming := make([]map[string]any, 0, len(in))
	for _, r := range in {
		card := friendCard(sess.PlayerID, r.SenderID, st)
		incoming = append(incoming, map[string]any{
			"id": r.ID, "senderId": r.SenderID, "receiverId": r.ReceiverID, "status": r.Status,
			"username": card["username"], "avatar": card["avatar"],
		})
	}
	outgoing := make([]map[string]any, 0, len(outg))
	for _, r := range outg {
		card := friendCard(sess.PlayerID, r.ReceiverID, st)
		outgoing = append(outgoing, map[string]any{
			"id": r.ID, "senderId": r.SenderID, "receiverId": r.ReceiverID, "status": r.Status,
			"username": card["username"], "avatar": card["avatar"],
		})
	}
	sort.Slice(friends, func(i, j int) bool {
		oi, oj := statusOrder(friends[i]["status"]), statusOrder(friends[j]["status"])
		if oi != oj {
			return oi < oj
		}
		return strings.ToLower(str(friends[i]["username"])) < strings.ToLower(str(friends[j]["username"]))
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"friends": friends, "incoming": incoming, "outgoing": outgoing,
		"unread": st.UnreadCount(sess.PlayerID),
	})
}

func statusOrder(v any) int {
	switch str(v) {
	case PresOnline:
		return 0
	case PresInGame:
		return 1
	case PresAway:
		return 2
	default:
		return 3
	}
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func friendCard(viewer, id string, st *SocialStore) map[string]any {
	card := map[string]any{"userId": id, "status": st.Status(id, viewer)}
	if DefaultHub != nil && DefaultHub.Progress != nil {
		if p, ok := DefaultHub.Progress.Get(id); ok {
			rs := p.RankState()
			card["username"] = p.Username
			card["avatar"] = p.Avatar
			card["level"] = p.Level
			card["rankLabel"] = rs.Label()
			card["rankRr"] = rs.RR
		}
	}
	if card["username"] == nil && DefaultHub != nil && DefaultHub.Accounts != nil {
		if acc := DefaultHub.Accounts.AccountByID(id); acc != nil {
			card["username"] = acc.Username
		}
	}
	return card
}

func HandleFriendSearch(w http.ResponseWriter, r *http.Request) {
	sess, st, ok := requireSocial(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	found := DefaultHub.Accounts.SearchUsers(q, 20)
	out := make([]map[string]any, 0, len(found))
	for _, acc := range found {
		if acc.PlayerID == sess.PlayerID {
			continue
		}
		if st.Blocked(sess.PlayerID, acc.PlayerID) {
			continue
		}
		card := friendCard(sess.PlayerID, acc.PlayerID, st)
		card["username"] = acc.Username
		card["friends"] = st.AreFriends(sess.PlayerID, acc.PlayerID)
		out = append(out, card)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func HandleFriendAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, st, ok := requireSocial(w, r)
	if !ok {
		return
	}
	raw, _ := io.ReadAll(r.Body)
	if rejectClientEconomy(w, raw) {
		return
	}
	var in struct {
		Username  string `json:"username"`
		UserID    string `json:"userId"`
		RequestID string `json:"requestId"`
		Action    string `json:"action"`
	}
	if json.Unmarshal(raw, &in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/request"):
		uid := in.UserID
		if uid == "" && in.Username != "" {
			if acc := DefaultHub.Accounts.AccountByUsername(in.Username); acc != nil {
				uid = acc.PlayerID
			}
		}
		req, msg := st.RequestFriend(sess.PlayerID, uid)
		if msg != "" {
			writeAccountErr(w, http.StatusBadRequest, msg)
			return
		}
		if pl := DefaultHub.Lobby.online[uid]; pl != nil {
			DefaultHub.pushUser(uid, TypeSocialPush, map[string]any{"type": NoteFriendReq, "from": sess.Username})
		}
		writeJSON(w, http.StatusOK, req)
	case strings.HasSuffix(path, "/respond"):
		req, msg := st.RespondFriend(sess.PlayerID, in.RequestID, in.Action)
		if msg != "" {
			writeAccountErr(w, http.StatusBadRequest, msg)
			return
		}
		writeJSON(w, http.StatusOK, req)
	case strings.HasSuffix(path, "/remove"):
		if msg := st.RemoveFriend(sess.PlayerID, in.UserID); msg != "" {
			writeAccountErr(w, http.StatusBadRequest, msg)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	case strings.HasSuffix(path, "/block"):
		if msg := st.Block(sess.PlayerID, in.UserID); msg != "" {
			writeAccountErr(w, http.StatusBadRequest, msg)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	case strings.HasSuffix(path, "/unblock"):
		st.Unblock(sess.PlayerID, in.UserID)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		writeAccountErr(w, http.StatusNotFound, "not found")
	}
}

func HandlePublicPlayer(w http.ResponseWriter, r *http.Request) {
	sess, st, ok := requireSocial(w, r)
	if !ok {
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("userId"))
	if id == "" {
		writeAccountErr(w, http.StatusBadRequest, "userId")
		return
	}
	view := DefaultHub.Progress.PublicSafe(id)
	if view == nil {
		writeAccountErr(w, http.StatusNotFound, "tidak ditemukan")
		return
	}
	view["status"] = st.Status(id, sess.PlayerID)
	writeJSON(w, http.StatusOK, view)
}

func HandleNotifications(w http.ResponseWriter, r *http.Request) {
	sess, st, ok := requireSocial(w, r)
	if !ok {
		return
	}
	if r.Method == http.MethodPost {
		st.MarkRead(sess.PlayerID)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": st.Notes(sess.PlayerID, false), "unread": st.UnreadCount(sess.PlayerID)})
}

func HandlePrivacy(w http.ResponseWriter, r *http.Request) {
	sess, st, ok := requireSocial(w, r)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, st.Privacy(sess.PlayerID))
		return
	}
	var in SocialPrivacy
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	st.SetPrivacy(sess.PlayerID, in)
	writeJSON(w, http.StatusOK, in)
}

func HandleReportPlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	sess, st, ok := requireSocial(w, r)
	if !ok {
		return
	}
	var in struct {
		UserID, Reason, Description, MatchID string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	rep, msg := st.Report(sess.PlayerID, in.UserID, in.Reason, in.Description, in.MatchID)
	if msg != "" {
		writeAccountErr(w, http.StatusBadRequest, msg)
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

func HandleSeason(w http.ResponseWriter, r *http.Request) {
	if o := liveOps(); o != nil {
		writeJSON(w, http.StatusOK, o.ActiveSeason())
		return
	}
	writeJSON(w, http.StatusOK, CurrentSeason)
}

func HandleRankMe(w http.ResponseWriter, r *http.Request) {
	sess, st, ok := requireProgressSession(w, r)
	if !ok {
		return
	}
	view := st.ViewFor(sess.PlayerID, sess.Username)
	rs := st.Ensure(sess.PlayerID, sess.Username).RankState()
	view["myRank"] = st.rankBy(sess.PlayerID, func(o PlayerProfile) int { return o.RankState().Index })
	view["rankLabel"] = rs.Label()
	writeJSON(w, http.StatusOK, view)
}

func HandleRankHistoryHTTP(w http.ResponseWriter, r *http.Request) {
	sess, st, ok := requireProgressSession(w, r)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	writeJSON(w, http.StatusOK, map[string]any{"items": st.RankHistory(sess.PlayerID, page), "page": page, "pageSize": MaxMatchHistoryPage})
}

type lbRow struct {
	UserID    string  `json:"userId"`
	Username  string  `json:"username"`
	Avatar    string  `json:"avatar"`
	Level     int     `json:"level"`
	Tier      string  `json:"tier"`
	Division  string  `json:"division,omitempty"`
	RR        int     `json:"rr"`
	Wins      int     `json:"wins"`
	XP        int     `json:"xp"`
	Accuracy  float64 `json:"accuracy"`
	Coins     int     `json:"coins"`
	RankLabel string  `json:"rankLabel"`
	Score     int     `json:"score"`
}

var (
	lbMu    sync.Mutex
	lbAt    time.Time
	lbCache map[string][]lbRow
)

func InvalidateLeaderboard() {
	lbMu.Lock()
	lbAt = time.Time{}
	lbMu.Unlock()
}

func leaderboardRows(kind string) []lbRow {
	lbMu.Lock()
	defer lbMu.Unlock()
	if time.Since(lbAt) < time.Duration(LBCacheSec)*time.Second && lbCache != nil {
		return lbCache[kind]
	}
	lbCache = map[string][]lbRow{}
	if DefaultHub == nil || DefaultHub.Progress == nil {
		return nil
	}
	all := DefaultHub.Progress.AllProfiles()
	build := func(score func(PlayerProfile) int) []lbRow {
		rows := make([]lbRow, 0, len(all))
		for _, p := range all {
			acc := 0.0
			if p.TotalQuestions > 0 {
				acc = float64(p.CorrectAnswers) / float64(p.TotalQuestions) * 100
			}
			rs := p.RankState()
			rows = append(rows, lbRow{
				UserID: p.UserID, Username: p.Username, Avatar: p.Avatar, Level: p.Level,
				Tier: rs.Tier, Division: rs.Division, RR: rs.RR, Wins: p.Wins, XP: p.XP,
				Accuracy: acc, Coins: p.Coins, RankLabel: rs.Label(), Score: score(p),
			})
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Score != rows[j].Score {
				return rows[i].Score > rows[j].Score
			}
			return rows[i].Username < rows[j].Username
		})
		return rows
	}
	lbCache["rank"] = build(func(p PlayerProfile) int { return p.RankState().Index })
	lbCache["wins"] = build(func(p PlayerProfile) int { return p.Wins })
	lbCache["xp"] = build(func(p PlayerProfile) int { return p.XP })
	lbCache["coins"] = build(func(p PlayerProfile) int { return p.Coins })
	lbCache["accuracy"] = build(func(p PlayerProfile) int {
		if p.TotalQuestions == 0 {
			return 0
		}
		return (p.CorrectAnswers * 10000) / p.TotalQuestions
	})
	lbAt = time.Now()
	return lbCache[kind]
}

func HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	sess, st, ok := requireSocial(w, r)
	if !ok {
		return
	}
	kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
	if kind == "" {
		kind = "rank"
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 0 {
		page = 0
	}
	scope := strings.ToLower(r.URL.Query().Get("scope"))
	rows := leaderboardRows(kind)
	if scope == "friends" {
		allow := map[string]bool{sess.PlayerID: true}
		for _, id := range st.FriendIDs(sess.PlayerID) {
			allow[id] = true
		}
		filt := rows[:0]
		for _, row := range rows {
			if allow[row.UserID] {
				filt = append(filt, row)
			}
		}
		rows = filt
	}
	my := 0
	for i, row := range rows {
		if row.UserID == sess.PlayerID {
			my = i + 1
			break
		}
	}
	start := page * LeaderboardPage
	if start > len(rows) {
		start = len(rows)
	}
	end := start + LeaderboardPage
	if end > len(rows) {
		end = len(rows)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": rows[start:end], "page": page, "pageSize": LeaderboardPage,
		"total": len(rows), "myRank": my, "kind": kind, "scope": scope,
	})
}
