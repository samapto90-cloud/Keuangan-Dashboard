package mmo

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var adminLimit = newUlarLimiter()

type adminCtx struct {
	Sess *GameSession
	Role string
	ID   string
	Name string
}

func adminRel(path string) string {
	path = strings.TrimSuffix(path, "/")
	for _, p := range []string{"/cahaya/api/admin", "/admin/api"} {
		if path == p {
			return ""
		}
		if strings.HasPrefix(path, p+"/") {
			return strings.TrimPrefix(path, p+"/")
		}
	}
	return strings.TrimPrefix(path, "/")
}

func requireAdmin(w http.ResponseWriter, r *http.Request, perm string) (*adminCtx, bool) {
	tok := strings.TrimSpace(os.Getenv("ULAR_ADMIN_TOKEN"))
	got := strings.TrimSpace(r.Header.Get("X-Ular-Admin"))
	if got == "" {
		h := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(h), "bearer ") {
			got = strings.TrimSpace(h[7:])
		}
	}
	if tok != "" && got == tok {
		if !HasPerm(RoleSuperAdmin, perm) {
			writeAccountErr(w, http.StatusForbidden, "forbidden")
			return nil, false
		}
		return &adminCtx{Role: RoleSuperAdmin, ID: "env-token", Name: "env-admin"}, true
	}
	if DefaultHub == nil || DefaultHub.Accounts == nil {
		writeAccountErr(w, http.StatusServiceUnavailable, "tidak siap")
		return nil, false
	}
	sess := DefaultHub.Accounts.Lookup(bearerFrom(r))
	if sess == nil {
		writeAccountErr(w, http.StatusUnauthorized, "sesi tidak valid")
		return nil, false
	}
	role := DefaultHub.Accounts.RoleOf(sess.PlayerID)
	if role == RolePlayer || !HasPerm(role, perm) {
		writeAccountErr(w, http.StatusForbidden, "forbidden")
		return nil, false
	}
	if !adminLimit.allow("adm:"+sess.PlayerID+":"+r.Method, 60, 10*time.Second) {
		writeAccountErr(w, http.StatusTooManyRequests, "rate limit")
		return nil, false
	}
	return &adminCtx{Sess: sess, Role: role, ID: sess.PlayerID, Name: sess.Username}, true
}

func (a *adminCtx) audit(action, typ, id, before, after, ip string) {
	if DefaultHub != nil && DefaultHub.Ops != nil {
		DefaultHub.Ops.Audit(a.ID, a.Name, action, typ, id, before, after, ip)
	}
}

func HandleAdminAPI(w http.ResponseWriter, r *http.Request) {
	rel := adminRel(r.URL.Path)
	parts := strings.Split(rel, "/")
	if rel == "" {
		parts = []string{"me"}
	}
	switch {
	case rel == "me" && r.Method == http.MethodGet:
		a, ok := requireAdmin(w, r, PermPlayerView)
		if !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"role": a.Role, "perms": RolePerms(a.Role), "adminId": a.ID, "username": a.Name})
	case rel == "dashboard" && r.Method == http.MethodGet:
		a, ok := requireAdmin(w, r, PermPlayerView)
		if !ok {
			return
		}
		_ = a
		writeJSON(w, http.StatusOK, adminDashboard(r))
	case rel == "status" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermPlayerView); !ok {
			return
		}
		writeJSON(w, http.StatusOK, adminStatus())
	case rel == "errors" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermAuditView); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": DefaultHub.Ops.Errors()})
	case rel == "backup/restore-test" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermConfigEdit)
		if !ok {
			return
		}
		adminUlarRestoreTest(w, r, a)
	case rel == "players" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermPlayerView); !ok {
			return
		}
		adminListPlayers(w, r)
	case len(parts) == 2 && parts[0] == "players" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermPlayerView); !ok {
			return
		}
		adminPlayerDetail(w, parts[1])
	case len(parts) == 3 && parts[0] == "players" && parts[2] == "sanction" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermPlayerBan)
		if !ok {
			return
		}
		adminSanction(w, r, a, parts[1])
	case len(parts) == 3 && parts[0] == "players" && parts[2] == "unban" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermPlayerBan)
		if !ok {
			return
		}
		DefaultHub.Ops.DeactivateSanctions(parts[1], "")
		a.audit("PLAYER_UNBANNED", "player", parts[1], "", "", r.RemoteAddr)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	case len(parts) == 3 && parts[0] == "players" && parts[2] == "role" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermAdminManage)
		if !ok {
			return
		}
		var in struct {
			Role string `json:"role"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			writeAccountErr(w, http.StatusBadRequest, "payload")
			return
		}
		if msg := DefaultHub.Accounts.SetRole(parts[1], in.Role); msg != "" {
			writeAccountErr(w, http.StatusBadRequest, msg)
			return
		}
		a.audit("ROLE_CHANGED", "player", parts[1], "", in.Role, r.RemoteAddr)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	case rel == "questions" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermQuestionView); !ok {
			return
		}
		adminListQuestions(w, r)
	case rel == "questions" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermQuestionCreate)
		if !ok {
			return
		}
		adminUpsertQuestion(w, r, a, true)
	case rel == "questions/export" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermQuestionView); !ok {
			return
		}
		adminExportQuestions(w)
	case rel == "questions/import" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermQuestionCreate)
		if !ok {
			return
		}
		adminImportQuestions(w, r, a)
	case len(parts) == 2 && parts[0] == "questions" && r.Method == http.MethodPut:
		a, ok := requireAdmin(w, r, PermQuestionEdit)
		if !ok {
			return
		}
		adminUpsertQuestionID(w, r, a, parts[1])
	case len(parts) == 2 && parts[0] == "questions" && r.Method == http.MethodDelete:
		a, ok := requireAdmin(w, r, PermQuestionDelete)
		if !ok {
			return
		}
		if !DefaultEduBank().Delete(parts[1]) {
			writeAccountErr(w, http.StatusNotFound, "not found")
			return
		}
		a.audit("QUESTION_DELETED", "question", parts[1], "", "soft-delete", r.RemoteAddr)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	case rel == "matches" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermMatchView); !ok {
			return
		}
		adminListMatches(w, r)
	case len(parts) == 2 && parts[0] == "matches" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermMatchView); !ok {
			return
		}
		adminMatchDetail(w, parts[1])
	case len(parts) == 3 && parts[0] == "matches" && parts[2] == "terminate" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermMatchTerminate)
		if !ok {
			return
		}
		adminTerminate(w, r, a, parts[1])
	case rel == "reports" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermReportView); !ok {
			return
		}
		adminListReports(w, r)
	case len(parts) == 3 && parts[0] == "reports" && parts[2] == "resolve" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermReportResolve)
		if !ok {
			return
		}
		adminResolveReport(w, r, a, parts[1])
	case rel == "achievements" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermConfigView); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": liveAchievements()})
	case rel == "achievements" && (r.Method == http.MethodPut || r.Method == http.MethodPost):
		a, ok := requireAdmin(w, r, PermAchievementEdit)
		if !ok {
			return
		}
		adminSaveAchievements(w, r, a)
	case rel == "rewards" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermConfigView); !ok {
			return
		}
		c := LiveConfig()
		writeJSON(w, http.StatusOK, map[string]any{"dailyCoins": LiveDailyCoins(), "dailyXp": c.DailyXP, "xpCorrect": c.XPCorrect, "xpWrong": c.XPWrong, "xpTimeout": c.XPTimeout, "xpMatch": c.XPMatchComplete, "xpWin": c.XPWin, "coinMatch": c.CoinMatch, "coinWin": c.CoinWin, "coinAchievement": c.CoinAchievement})
	case rel == "rewards" && r.Method == http.MethodPut:
		a, ok := requireAdmin(w, r, PermRewardEdit)
		if !ok {
			return
		}
		adminPutRewards(w, r, a)
	case rel == "ranks" && r.Method == http.MethodPut:
		a, ok := requireAdmin(w, r, PermRankEdit)
		if !ok {
			return
		}
		adminPutRanks(w, r, a)
	case rel == "config" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermConfigView); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"active": LiveConfig(), "versions": DefaultHub.Ops.Configs(), "activeVersion": DefaultHub.Ops.ActiveConfig().Version})
	case rel == "config" && r.Method == http.MethodPut:
		a, ok := requireAdmin(w, r, PermConfigEdit)
		if !ok {
			return
		}
		adminPutConfig(w, r, a)
	case rel == "config/rollback" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermConfigEdit)
		if !ok {
			return
		}
		adminRollback(w, r, a)
	case rel == "flags" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermConfigView); !ok {
			return
		}
		writeJSON(w, http.StatusOK, LiveFlags())
	case rel == "flags" && r.Method == http.MethodPut:
		a, ok := requireAdmin(w, r, PermConfigEdit)
		if !ok {
			return
		}
		var f FeatureFlags
		if json.NewDecoder(r.Body).Decode(&f) != nil {
			writeAccountErr(w, http.StatusBadRequest, "payload")
			return
		}
		before := clipJSON(LiveFlags())
		DefaultHub.Ops.SetFlags(f)
		a.audit("CONFIG_CHANGED", "flags", "feature", before, clipJSON(f), r.RemoteAddr)
		writeJSON(w, http.StatusOK, f)
	case rel == "ranks" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermConfigView); !ok {
			return
		}
		c := LiveConfig()
		writeJSON(w, http.StatusOK, map[string]any{"tiers": RankTiers, "divisions": RankDivisions, "winRr": c.RankWinRR, "lossRr": c.RankLossRR, "rrPerDivision": RRPerDivision})
	case rel == "seasons" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermConfigView); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": DefaultHub.Ops.Seasons(), "active": DefaultHub.Ops.ActiveSeason()})
	case rel == "seasons" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermSeasonManage)
		if !ok {
			return
		}
		adminUpsertSeason(w, r, a)
	case len(parts) == 3 && parts[0] == "seasons" && parts[2] == "preview" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermSeasonManage); !ok {
			return
		}
		adminSeasonPreview(w)
	case len(parts) == 3 && parts[0] == "seasons" && parts[2] == "end" && r.Method == http.MethodPost:
		a, ok := requireAdmin(w, r, PermSeasonManage)
		if !ok {
			return
		}
		adminEndSeason(w, r, a, parts[1])
	case rel == "audit-logs" && r.Method == http.MethodGet:
		if _, ok := requireAdmin(w, r, PermAuditView); !ok {
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		items, total := DefaultHub.Ops.AuditPage(page, 20, r.URL.Query().Get("action"))
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page})
	default:
		writeAccountErr(w, http.StatusNotFound, "not found")
	}
}

func adminDashboard(r *http.Request) map[string]any {
	from, to := adminRange(r)
	accs := DefaultHub.Accounts.AllAccounts()
	online := 0
	if DefaultHub.Lobby != nil {
		DefaultHub.Lobby.mu.Lock()
		online = len(DefaultHub.Lobby.online)
		DefaultHub.Lobby.mu.Unlock()
	}
	active := 0
	matchesToday := 0
	newToday := 0
	today := JakartaDate(time.Now())
	var matches []StoredMatch
	if DefaultHub.Lobby != nil && DefaultHub.Lobby.Store != nil {
		matches = DefaultHub.Lobby.Store.All()
	}
	for _, m := range matches {
		if m.StartedAt >= from && (to == 0 || m.StartedAt <= to) {
			matchesToday++
		}
	}
	if DefaultHub.Lobby != nil {
		DefaultHub.Lobby.mu.Lock()
		for _, rm := range DefaultHub.Lobby.rooms {
			if rm != nil && (rm.Status == UlarPlaying || rm.Status == UlarStarting) {
				active++
			}
		}
		DefaultHub.Lobby.mu.Unlock()
	}
	for _, a := range accs {
		if JakartaDate(time.UnixMilli(a.CreatedAt)) == today {
			newToday++
		}
	}
	pending := 0
	if DefaultHub.Social != nil {
		for _, rp := range DefaultHub.Social.AllReports() {
			if rp.Status == ReportOpen || rp.Status == "" {
				pending++
			}
		}
	}
	bank := DefaultEduBank().Validate()
	charts := adminCharts(from, to, accs, matches)
	return map[string]any{
		"totalPlayers": len(accs), "onlinePlayers": online, "activeMatches": active,
		"matchesToday": matchesToday, "questions": bank.Total, "reportsPending": pending,
		"newUsersToday": newToday, "charts": charts, "bank": bank,
	}
}

func adminRange(r *http.Request) (from, to int64) {
	q := r.URL.Query().Get("range")
	now := time.Now()
	to = now.UnixMilli()
	switch q {
	case "30":
		from = now.AddDate(0, 0, -30).UnixMilli()
	case "custom":
		from, _ = strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
		to2, _ := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
		if to2 > 0 {
			to = to2
		}
	default:
		from = now.AddDate(0, 0, -7).UnixMilli()
	}
	if q == "today" {
		t, _ := time.ParseInLocation("2006-01-02", JakartaDate(now), time.FixedZone("WIB", 7*3600))
		from = t.UnixMilli()
	}
	return from, to
}

func adminCharts(from, to int64, accs []GameAccount, matches []StoredMatch) map[string]any {
	days := map[string]map[string]int{}
	bump := func(day, k string) {
		if days[day] == nil {
			days[day] = map[string]int{}
		}
		days[day][k]++
	}
	for _, a := range accs {
		if a.CreatedAt >= from && a.CreatedAt <= to {
			bump(JakartaDate(time.UnixMilli(a.CreatedAt)), "newUsers")
		}
	}
	playersDay := map[string]map[string]bool{}
	for _, m := range matches {
		if m.StartedAt < from || m.StartedAt > to {
			continue
		}
		d := JakartaDate(time.UnixMilli(m.StartedAt))
		bump(d, "matches")
		if playersDay[d] == nil {
			playersDay[d] = map[string]bool{}
		}
		for _, p := range m.Players {
			playersDay[d][p.UserID] = true
		}
	}
	for d, set := range playersDay {
		if days[d] == nil {
			days[d] = map[string]int{}
		}
		days[d]["dap"] = len(set)
	}
	sub := map[string]int{}
	correct, total := 0, 0
	if DefaultHub.Lobby != nil && DefaultHub.Lobby.Attempts != nil {
		for _, a := range DefaultHub.Lobby.Attempts.All() {
			if a.CreatedAt < from || a.CreatedAt > to {
				continue
			}
			total++
			sub[a.Subject]++
			if a.Correct {
				correct++
			}
		}
	}
	acc := 0.0
	if total > 0 {
		acc = float64(correct) / float64(total) * 100
	}
	return map[string]any{"daily": days, "accuracy": acc, "subjects": sub, "questions": total}
}

func adminStatus() map[string]any {
	st := func(ok bool) string {
		if ok {
			return "HEALTHY"
		}
		return "DOWN"
	}
	api := DefaultHub != nil
	db := DefaultHub != nil && DefaultHub.Accounts != nil && DefaultHub.Progress != nil
	ws := DefaultHub != nil
	mm := DefaultHub != nil && DefaultHub.Matchmaker != nil
	qs := DefaultEduBank() != nil
	errors := []OpsError{}
	if DefaultHub != nil && DefaultHub.Ops != nil {
		errors = DefaultHub.Ops.Errors()
	}
	backupOk := !ularLastBackupAt.IsZero() && ularLastBackupErr == ""
	backupSt := "DOWN"
	if backupOk {
		backupSt = "HEALTHY"
	}
	return map[string]any{
		"api": st(api), "database": st(db), "websocket": st(ws),
		"matchmaking": st(mm), "questionService": st(qs),
		"backup":       backupSt,
		"backupDetail": UlarBackupStatus(),
		"errors": errors,
	}
}

func adminListPlayers(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	status := strings.ToUpper(r.URL.Query().Get("status"))
	rank := strings.ToUpper(r.URL.Query().Get("rank"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	levelMin, _ := strconv.Atoi(r.URL.Query().Get("levelMin"))
	levelMax, _ := strconv.Atoi(r.URL.Query().Get("levelMax"))
	from, _ := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	to, _ := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	accs := DefaultHub.Accounts.AllAccounts()
	items := make([]map[string]any, 0)
	online := map[string]bool{}
	if DefaultHub.Lobby != nil {
		DefaultHub.Lobby.mu.Lock()
		for id := range DefaultHub.Lobby.online {
			online[id] = true
		}
		DefaultHub.Lobby.mu.Unlock()
	}
	for _, a := range accs {
		if q != "" && !strings.Contains(strings.ToLower(a.Username), q) && !strings.Contains(strings.ToLower(a.PlayerID), q) {
			continue
		}
		if from > 0 && a.CreatedAt < from {
			continue
		}
		if to > 0 && a.CreatedAt > to {
			continue
		}
		banned := DefaultHub.Ops != nil && DefaultHub.Ops.ActiveBan(a.PlayerID) != nil
		st := "OFFLINE"
		if banned {
			st = "BANNED"
		} else if online[a.PlayerID] {
			st = "ONLINE"
		}
		if status != "" && st != status {
			continue
		}
		row := map[string]any{"userId": a.PlayerID, "username": a.Username, "role": NormalizeRole(a.Role), "status": st, "createdAt": a.CreatedAt}
		if DefaultHub.Progress != nil {
			if p, ok := DefaultHub.Progress.Get(a.PlayerID); ok {
				rs := p.RankState()
				if rank != "" && rs.Tier != rank {
					continue
				}
				if levelMin > 0 && p.Level < levelMin {
					continue
				}
				if levelMax > 0 && p.Level > levelMax {
					continue
				}
				row["level"] = p.Level
				row["xp"] = p.XP
				row["coins"] = p.Coins
				row["rank"] = rs.Label()
				row["rr"] = rs.RR
				row["matches"] = p.TotalMatches
				row["wins"] = p.Wins
			}
		}
		if DefaultHub.Social != nil {
			row["lastSeen"] = DefaultHub.Social.Status(a.PlayerID, a.PlayerID)
		}
		items = append(items, row)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": pageSlice(items, page, 20), "total": len(items), "page": page})
}

func adminPlayerDetail(w http.ResponseWriter, id string) {
	acc := DefaultHub.Accounts.AccountByID(id)
	if acc == nil {
		writeAccountErr(w, http.StatusNotFound, "tidak ditemukan")
		return
	}
	out := map[string]any{"account": acc, "sanctions": DefaultHub.Ops.SanctionsFor(id)}
	if DefaultHub.Progress != nil {
		out["profile"] = DefaultHub.Progress.ViewFor(id, acc.Username)
		out["history"] = DefaultHub.Progress.History(id, 0)
		out["rankHistory"] = DefaultHub.Progress.RankHistory(id, 0)
		out["coins"] = DefaultHub.Progress.TxCoins(id, 0)
		out["xp"] = DefaultHub.Progress.TxXP(id, 0)
		out["achievements"] = DefaultHub.Progress.Unlocks(id)
	}
	if DefaultHub.Social != nil {
		reps := make([]PlayerReport, 0)
		for _, r := range DefaultHub.Social.AllReports() {
			if r.ReportedID == id || r.ReporterID == id {
				reps = append(reps, r)
			}
		}
		out["reports"] = reps
	}
	writeJSON(w, http.StatusOK, out)
}

func adminSanction(w http.ResponseWriter, r *http.Request, a *adminCtx, userID string) {
	if !adminLimit.allow("ban:"+a.ID, 8, time.Minute) {
		writeAccountErr(w, http.StatusTooManyRequests, "rate limit")
		return
	}
	var in struct {
		Type, Reason, Confirm string
		EndAt                 int64
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Reason) == "" {
		writeAccountErr(w, http.StatusBadRequest, "reason wajib")
		return
	}
	typ := strings.ToUpper(strings.TrimSpace(in.Type))
	switch typ {
	case SanctionWarning, SanctionChatMute, SanctionTempBan, SanctionPermBan:
	default:
		writeAccountErr(w, http.StatusBadRequest, "tipe sanksi")
		return
	}
	if typ == SanctionPermBan && strings.ToUpper(strings.TrimSpace(in.Confirm)) != "PERMANENT BAN" {
		writeAccountErr(w, http.StatusBadRequest, "ketik PERMANENT BAN")
		return
	}
	sn := DefaultHub.Ops.AddSanction(UserSanction{UserID: userID, Type: typ, Reason: in.Reason, IssuedBy: a.ID, EndAt: in.EndAt})
	if typ == SanctionTempBan || typ == SanctionPermBan {
		DefaultHub.Accounts.RevokeSessions(userID)
	}
	a.audit("PLAYER_BANNED", "player", userID, "", clipJSON(sn), r.RemoteAddr)
	writeJSON(w, http.StatusOK, sn)
}

func qStats() map[string]map[string]int {
	out := map[string]map[string]int{}
	if DefaultHub.Lobby == nil || DefaultHub.Lobby.Attempts == nil {
		return out
	}
	for _, a := range DefaultHub.Lobby.Attempts.All() {
		row := out[a.QuestionID]
		if row == nil {
			row = map[string]int{}
		}
		row["asked"]++
		if a.Timeout {
			row["timeout"]++
		} else if a.Correct {
			row["correct"]++
		} else {
			row["wrong"]++
		}
		out[a.QuestionID] = row
	}
	return out
}

func adminListQuestions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	sub := r.URL.Query().Get("subject")
	diff := r.URL.Query().Get("difficulty")
	grade := r.URL.Query().Get("grade")
	q := r.URL.Query().Get("q")
	active := r.URL.Query().Get("active")
	bank := DefaultEduBank()
	all := bank.ListAdmin(sub, diff, q)
	stats := qStats()
	thr := LiveConfig().ReviewAccuracy
	if thr <= 0 {
		thr = 25
	}
	items := make([]map[string]any, 0, len(all))
	for _, it := range all {
		if grade != "" && it.Grade != grade {
			continue
		}
		if active == "1" && (!it.Active || it.Deleted) {
			continue
		}
		if active == "0" && it.Active && !it.Deleted {
			continue
		}
		st := stats[it.ID]
		asked := st["asked"]
		acc := 0.0
		if asked > 0 {
			acc = float64(st["correct"]) / float64(asked) * 100
		}
		warn := ""
		review := false
		if asked >= 20 && acc < thr {
			warn = "Mungkin terlalu sulit."
			review = true
		}
		status := "ACTIVE"
		if it.Deleted {
			status = "DELETED"
		} else if !it.Active {
			status = "INACTIVE"
		}
		items = append(items, map[string]any{
			"id": it.ID, "subject": it.Subject, "grade": it.Grade, "difficulty": it.Difficulty,
			"question": it.Question, "correctAnswer": it.CorrectAnswer, "status": status,
			"createdAt": it.CreatedAt, "timesAsked": asked, "correctCount": st["correct"],
			"wrongCount": st["wrong"], "timeoutCount": st["timeout"], "accuracy": acc,
			"warning": warn, "needsReview": review, "optionA": it.OptionA, "optionB": it.OptionB,
			"optionC": it.OptionC, "optionD": it.OptionD, "explanation": it.Explanation, "active": it.Active,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": pageSlice(items, page, 20), "total": len(items), "page": page, "summary": DefaultEduBank().Validate()})
}

func adminUpsertQuestion(w http.ResponseWriter, r *http.Request, a *adminCtx, create bool) {
	var item EduQuestion
	if json.NewDecoder(r.Body).Decode(&item) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	if create && item.ID == "" {
		item.ID = "q-" + shortID()
	}
	item.Active = true
	if err := DefaultEduBank().Upsert(item); err != nil {
		writeAccountErr(w, http.StatusBadRequest, err.Error())
		return
	}
	act := "QUESTION_UPDATED"
	if create {
		act = "QUESTION_CREATED"
	}
	a.audit(act, "question", item.ID, "", clipJSON(item.ID), r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": item.ID})
}

func adminUpsertQuestionID(w http.ResponseWriter, r *http.Request, a *adminCtx, id string) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	var item EduQuestion
	if json.Unmarshal(raw, &item) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	item.ID = id
	old, ok := DefaultEduBank().Get(id)
	if ok {
		keys := map[string]json.RawMessage{}
		_ = json.Unmarshal(raw, &keys)
		if _, has := keys["active"]; !has {
			item.Active = old.Active
		}
		if item.CreatedAt == 0 {
			item.CreatedAt = old.CreatedAt
		}
	}
	if err := DefaultEduBank().Upsert(item); err != nil {
		writeAccountErr(w, http.StatusBadRequest, err.Error())
		return
	}
	a.audit("QUESTION_UPDATED", "question", id, clipJSON(old), clipJSON(item), r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func adminExportQuestions(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=questions.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"subject", "grade", "difficulty", "question", "optionA", "optionB", "optionC", "optionD", "correctAnswer", "explanation"})
	for _, q := range DefaultEduBank().ListAdmin("", "", "") {
		if q.Deleted {
			continue
		}
		_ = cw.Write([]string{q.Subject, q.Grade, q.Difficulty, q.Question, q.OptionA, q.OptionB, q.OptionC, q.OptionD, q.CorrectAnswer, q.Explanation})
	}
	cw.Flush()
}

func adminImportQuestions(w http.ResponseWriter, r *http.Request, a *adminCtx) {
	raw, _ := io.ReadAll(r.Body)
	preview := r.URL.Query().Get("commit") != "1"
	cr := csv.NewReader(strings.NewReader(string(raw)))
	rows, err := cr.ReadAll()
	if err != nil || len(rows) < 2 {
		writeAccountErr(w, http.StatusBadRequest, "csv")
		return
	}
	seen := map[string]bool{}
	for _, q := range DefaultEduBank().ListAdmin("", "", "") {
		seen[strings.ToLower(strings.TrimSpace(q.Question))] = true
	}
	type rowOut struct {
		Row      int    `json:"row"`
		Status   string `json:"status"`
		Question string `json:"question"`
		Error    string `json:"error,omitempty"`
		Warning  string `json:"warning,omitempty"`
	}
	out := make([]rowOut, 0)
	valid := make([]EduQuestion, 0)
	existing := DefaultEduBank().ListAdmin("", "", "")
	for i, rec := range rows[1:] {
		if len(rec) < 10 {
			out = append(out, rowOut{Row: i + 2, Status: "INVALID", Error: "kolom kurang"})
			continue
		}
		q := EduQuestion{
			Subject: normSubject(rec[0]), Grade: strings.ToUpper(strings.TrimSpace(rec[1])),
			Difficulty: strings.ToUpper(strings.TrimSpace(rec[2])), Question: strings.TrimSpace(rec[3]),
			OptionA: rec[4], OptionB: rec[5], OptionC: rec[6], OptionD: rec[7],
			CorrectAnswer: strings.ToUpper(strings.TrimSpace(rec[8])), Explanation: rec[9], Active: true,
			ID: "q-" + shortID(),
		}
		key := strings.ToLower(q.Question)
		if seen[key] {
			out = append(out, rowOut{Row: i + 2, Status: "DUPLICATE", Question: q.Question})
			continue
		}
		tmp := []EduQuestion{q}
		if validateItems(tmp).Invalid > 0 {
			out = append(out, rowOut{Row: i + 2, Status: "INVALID", Question: q.Question, Error: "validasi gagal"})
			continue
		}
		warn := ""
		for _, old := range existing {
			if similarQuestion(q.Question, old.Question) {
				warn = "mirip soal yang sudah ada"
				break
			}
		}
		seen[key] = true
		out = append(out, rowOut{Row: i + 2, Status: "VALID", Question: q.Question, Warning: warn})
		valid = append(valid, q)
	}
	if preview {
		writeJSON(w, http.StatusOK, map[string]any{"preview": true, "rows": out, "valid": len(valid)})
		return
	}
	for _, row := range out {
		if row.Status == "INVALID" {
			writeAccountErr(w, http.StatusBadRequest, "validation gagal")
			return
		}
	}
	for i := range valid {
		if valid[i].ID == "" {
			valid[i].ID = "q-" + shortID()
		}
		if err := DefaultEduBank().Upsert(valid[i]); err != nil {
			writeAccountErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	a.audit("QUESTION_CREATED", "question", "import", "", strconv.Itoa(len(valid)), r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]any{"committed": len(valid), "rows": out})
}

func adminListMatches(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	mode := strings.ToUpper(r.URL.Query().Get("mode"))
	st := strings.ToUpper(r.URL.Query().Get("status"))
	live := make([]map[string]any, 0)
	if DefaultHub.Lobby != nil {
		DefaultHub.Lobby.mu.Lock()
		for _, rm := range DefaultHub.Lobby.rooms {
			if rm == nil || rm.Match == nil {
				continue
			}
			ids := make([]string, 0)
			for _, p := range rm.Players {
				if p != nil {
					ids = append(ids, p.Username)
				}
			}
			cur := ""
			dur := 0
			if rm.Match.StartedAt.Unix() > 0 {
				dur = int(time.Since(rm.Match.StartedAt).Seconds())
			}
			if rm.Match.CurrentPlayerID != "" {
				cur = rm.Match.CurrentPlayerID
			}
			live = append(live, map[string]any{
				"matchId": rm.Match.ID, "mode": rm.Mode, "players": ids, "status": rm.Status,
				"duration": dur, "currentTurn": cur, "live": true, "roomCode": rm.RoomCode,
			})
		}
		DefaultHub.Lobby.mu.Unlock()
	}
	hist := make([]map[string]any, 0)
	if DefaultHub.Lobby != nil && DefaultHub.Lobby.Store != nil {
		for _, m := range DefaultHub.Lobby.Store.All() {
			if mode != "" && strings.ToUpper(m.Mode) != mode {
				continue
			}
			if st != "" && strings.ToUpper(m.Status) != st {
				continue
			}
			names := make([]string, 0)
			for _, p := range m.Players {
				names = append(names, p.Username)
			}
			hist = append(hist, map[string]any{"matchId": m.ID, "mode": m.Mode, "players": names, "status": m.Status, "startedAt": m.StartedAt, "finishedAt": m.FinishedAt, "live": false})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"live": live, "history": pageSlice(hist, page, 20), "total": len(hist), "analytics": matchAnalytics(hist)})
}

func matchAnalytics(hist []map[string]any) map[string]any {
	n := len(hist)
	var durSum float64
	durN := 0
	abandoned, terminated := 0, 0
	for _, m := range hist {
		st, _ := m["status"].(string)
		switch strings.ToUpper(st) {
		case "ABANDONED":
			abandoned++
		case "ADMIN_TERMINATED":
			terminated++
		}
		sa := anyInt64(m["startedAt"])
		fa := anyInt64(m["finishedAt"])
		if sa > 0 && fa > sa {
			durSum += float64(fa-sa) / 1000
			durN++
		}
	}
	qTot, qOk := 0, 0
	if DefaultHub.Lobby != nil && DefaultHub.Lobby.Attempts != nil {
		for _, a := range DefaultHub.Lobby.Attempts.All() {
			qTot++
			if a.Correct {
				qOk++
			}
		}
	}
	avgDur, avgQ, corr, abandR, discR := 0.0, 0.0, 0.0, 0.0, 0.0
	if durN > 0 {
		avgDur = durSum / float64(durN)
	}
	if n > 0 {
		avgQ = float64(qTot) / float64(n)
		abandR = float64(abandoned) / float64(n) * 100
		discR = float64(terminated) / float64(n) * 100
	}
	if qTot > 0 {
		corr = float64(qOk) / float64(qTot) * 100
	}
	return map[string]any{
		"count": n, "averageMatchDuration": avgDur, "averageQuestions": avgQ,
		"correctRate": corr, "disconnectRate": discR, "abandonRate": abandR,
	}
}

func anyInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	default:
		return 0
	}
}

func adminMatchDetail(w http.ResponseWriter, id string) {
	if DefaultHub.Lobby != nil {
		DefaultHub.Lobby.mu.Lock()
		for _, rm := range DefaultHub.Lobby.rooms {
			if rm != nil && rm.Match != nil && rm.Match.ID == id {
				snap := DefaultHub.Lobby.snapshotLocked(rm, "")
				DefaultHub.Lobby.mu.Unlock()
				writeJSON(w, http.StatusOK, map[string]any{"live": true, "snapshot": snap})
				return
			}
		}
		DefaultHub.Lobby.mu.Unlock()
		if DefaultHub.Lobby.Store != nil {
			for _, m := range DefaultHub.Lobby.Store.All() {
				if m.ID == id {
					writeJSON(w, http.StatusOK, map[string]any{"live": false, "match": m})
					return
				}
			}
		}
	}
	writeAccountErr(w, http.StatusNotFound, "tidak ditemukan")
}

func adminTerminate(w http.ResponseWriter, r *http.Request, a *adminCtx, id string) {
	if !adminLimit.allow("term:"+a.ID, 8, time.Minute) {
		writeAccountErr(w, http.StatusTooManyRequests, "rate limit")
		return
	}
	var in struct{ Reason, Confirm string }
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Reason) == "" {
		writeAccountErr(w, http.StatusBadRequest, "reason wajib")
		return
	}
	if strings.ToUpper(strings.TrimSpace(in.Confirm)) != "TERMINATE" {
		writeAccountErr(w, http.StatusBadRequest, "konfirmasi TERMINATE")
		return
	}
	if err := DefaultHub.terminateMatch(id, in.Reason); err != "" {
		writeAccountErr(w, http.StatusBadRequest, err)
		return
	}
	a.audit("MATCH_TERMINATED", "match", id, "", in.Reason, r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Hub) terminateMatch(id, reason string) string {
	if h.Lobby == nil {
		return "lobby tidak siap"
	}
	h.Lobby.mu.Lock()
	defer h.Lobby.mu.Unlock()
	var room *UlarRoom
	for _, rm := range h.Lobby.rooms {
		if rm != nil && rm.Match != nil && rm.Match.ID == id {
			room = rm
			break
		}
	}
	if room == nil || room.Match == nil {
		return "match tidak ditemukan"
	}
	now := time.Now()
	room.Match.Terminated = true
	room.Match.TerminateReason = reason
	room.Match.Status = UlarTerminated
	room.Match.FinishedAt = &now
	room.Status = UlarTerminated
	h.persistMatchLocked(room)
	h.Lobby.closeLocked(room)
	return ""
}

func adminListReports(w http.ResponseWriter, r *http.Request) {
	if DefaultHub.Social == nil {
		writeJSON(w, http.StatusOK, map[string]any{"items": []PlayerReport{}, "total": 0})
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	st := strings.ToUpper(r.URL.Query().Get("status"))
	all := DefaultHub.Social.AllReports()
	items := make([]PlayerReport, 0)
	for i := len(all) - 1; i >= 0; i-- {
		if st != "" && strings.ToUpper(all[i].Status) != st {
			continue
		}
		items = append(items, all[i])
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": pageSlice(items, page, 20), "total": len(items)})
}

func adminResolveReport(w http.ResponseWriter, r *http.Request, a *adminCtx, id string) {
	var in struct {
		Status, Note, Action, UserID string
		EndAt                        int64
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Note) == "" {
		writeAccountErr(w, http.StatusBadRequest, "resolution note wajib")
		return
	}
	status := in.Status
	if status == "" {
		status = ReportResolved
	}
	rep, msg := DefaultHub.Social.ResolveReport(id, status, in.Note, a.ID)
	if msg != "" {
		writeAccountErr(w, http.StatusBadRequest, msg)
		return
	}
	if in.Action != "" && in.UserID != "" && HasPerm(a.Role, PermPlayerBan) {
		DefaultHub.Ops.AddSanction(UserSanction{UserID: in.UserID, Type: strings.ToUpper(in.Action), Reason: in.Note, IssuedBy: a.ID, EndAt: in.EndAt})
	}
	a.audit("REPORT_RESOLVED", "report", id, "", clipJSON(rep), r.RemoteAddr)
	writeJSON(w, http.StatusOK, rep)
}

func liveAchievements() []UlarAchievement {
	base := append([]UlarAchievement{}, AchievementCatalog...)
	if DefaultHub.Ops == nil {
		return base
	}
	over := DefaultHub.Ops.AchievementOverlay()
	if len(over) == 0 {
		return base
	}
	idx := map[string]UlarAchievement{}
	for _, a := range over {
		idx[a.ID] = a
	}
	for i, a := range base {
		if o, ok := idx[a.ID]; ok {
			base[i] = o
		}
	}
	for _, o := range over {
		found := false
		for _, a := range base {
			if a.ID == o.ID {
				found = true
				break
			}
		}
		if !found {
			base = append(base, o)
		}
	}
	return base
}

func adminSaveAchievements(w http.ResponseWriter, r *http.Request, a *adminCtx) {
	var in struct {
		Items []UlarAchievement `json:"items"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || len(in.Items) == 0 {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	DefaultHub.Ops.OverlayAchievements(in.Items)
	a.audit("CONFIG_CHANGED", "achievement", "catalog", "", strconv.Itoa(len(in.Items)), r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func adminPutConfig(w http.ResponseWriter, r *http.Request, a *adminCtx) {
	var c GameConfig
	if json.NewDecoder(r.Body).Decode(&c) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	if c.QuestionTimeLimit < 5 || c.WrongAnswerPenalty < 1 || c.MaxPlayers < 2 || c.MinPlayers < 2 {
		writeAccountErr(w, http.StatusBadRequest, "config tidak valid")
		return
	}
	if c.ReconnectGrace < 0 || c.MatchmakingTimeout < 0 {
		writeAccountErr(w, http.StatusBadRequest, "timeout tidak valid")
		return
	}
	// Reward + RR values must be validated fully for critical config updates.
	if c.XPCorrect < 0 || c.XPWrong < 0 || c.XPTimeout < 0 || c.XPMatchComplete < 0 || c.XPWin < 0 ||
		c.CoinMatch < 0 || c.CoinWin < 0 || c.CoinAchievement < 0 ||
		c.RankWinRR < 0 || c.RankLossRR < 0 {
		writeAccountErr(w, http.StatusBadRequest, "reward tidak boleh negatif")
		return
	}
	if len(c.DailyCoins) != 0 && len(c.DailyCoins) != 7 {
		writeAccountErr(w, http.StatusBadRequest, "siklus daily harus 7 hari")
		return
	}
	if len(c.DailyXP) != 0 && len(c.DailyXP) != 7 {
		writeAccountErr(w, http.StatusBadRequest, "siklus daily xp harus 7 hari")
		return
	}
	if len(c.DailyCoins) == 7 {
		for _, n := range c.DailyCoins {
			if n < 0 {
				writeAccountErr(w, http.StatusBadRequest, "daily negatif")
				return
			}
		}
	}
	if len(c.DailyXP) == 7 {
		for _, n := range c.DailyXP {
			if n < 0 {
				writeAccountErr(w, http.StatusBadRequest, "daily xp negatif")
				return
			}
		}
	}
	before := clipJSON(LiveConfig())
	saved := DefaultHub.Ops.PutConfig(c, a.ID)
	a.audit("CONFIG_CHANGED", "config", strconv.Itoa(saved.Version), before, clipJSON(saved), r.RemoteAddr)
	writeJSON(w, http.StatusOK, saved)
}

func adminRollback(w http.ResponseWriter, r *http.Request, a *adminCtx) {
	var in struct {
		Version int `json:"version"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Version <= 0 {
		writeAccountErr(w, http.StatusBadRequest, "version")
		return
	}
	c, msg := DefaultHub.Ops.Rollback(in.Version)
	if msg != "" {
		writeAccountErr(w, http.StatusBadRequest, msg)
		return
	}
	a.audit("CONFIG_CHANGED", "config", strconv.Itoa(in.Version), "", "rollback", r.RemoteAddr)
	writeJSON(w, http.StatusOK, c)
}

func adminUpsertSeason(w http.ResponseWriter, r *http.Request, a *adminCtx) {
	var se OpsSeason
	if json.NewDecoder(r.Body).Decode(&se) != nil || se.Name == "" {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	saved, msg := DefaultHub.Ops.UpsertSeason(se)
	if msg != "" {
		writeAccountErr(w, http.StatusBadRequest, msg)
		return
	}
	a.audit("SEASON_STARTED", "season", saved.ID, "", clipJSON(saved), r.RemoteAddr)
	writeJSON(w, http.StatusOK, saved)
}

func adminSeasonPreview(w http.ResponseWriter) {
	samples := []struct{ From, To string }{
		{"GRANDMASTER", "DIAMOND I"},
		{"MASTER", "PLATINUM I"},
		{"DIAMOND I", "GOLD I"},
	}
	n := 0
	if DefaultHub.Progress != nil {
		n = len(DefaultHub.Progress.AllProfiles())
	}
	writeJSON(w, http.StatusOK, map[string]any{"players": n, "examples": samples, "note": "soft reset 70% index, XP/coins/achievements tetap"})
}

func adminEndSeason(w http.ResponseWriter, r *http.Request, a *adminCtx, id string) {
	var in struct {
		Confirm string `json:"confirm"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if strings.ToUpper(strings.TrimSpace(in.Confirm)) != "END SEASON" {
		writeAccountErr(w, http.StatusBadRequest, "ketik END SEASON")
		return
	}
	n := 0
	if DefaultHub.Progress != nil {
		n = DefaultHub.Progress.SoftResetAllRanks()
	}
	for _, se := range DefaultHub.Ops.Seasons() {
		if se.ID == id {
			se.Active = false
			se.EndedBy = a.ID
			_, _ = DefaultHub.Ops.UpsertSeason(se)
		}
	}
	a.audit("SEASON_ENDED", "season", id, "", strconv.Itoa(n), r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "playersReset": n})
}

func adminPutRewards(w http.ResponseWriter, r *http.Request, a *adminCtx) {
	var in struct {
		DailyCoins      []int `json:"dailyCoins"`
		DailyXP         []int `json:"dailyXp"`
		XPCorrect       *int  `json:"xpCorrect"`
		XPWrong         *int  `json:"xpWrong"`
		XPTimeout       *int  `json:"xpTimeout"`
		XPMatch         *int  `json:"xpMatch"`
		XPWin           *int  `json:"xpWin"`
		CoinMatch       *int  `json:"coinMatch"`
		CoinWin         *int  `json:"coinWin"`
		CoinAchievement *int  `json:"coinAchievement"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	c := LiveConfig()
	if len(in.DailyCoins) == 7 {
		for _, n := range in.DailyCoins {
			if n < 0 {
				writeAccountErr(w, http.StatusBadRequest, "daily negatif")
				return
			}
		}
		c.DailyCoins = in.DailyCoins
	} else if in.DailyCoins != nil {
		writeAccountErr(w, http.StatusBadRequest, "siklus daily harus 7 hari")
		return
	}
	if len(in.DailyXP) == 7 {
		for _, n := range in.DailyXP {
			if n < 0 {
				writeAccountErr(w, http.StatusBadRequest, "daily xp negatif")
				return
			}
		}
		c.DailyXP = in.DailyXP
	}
	set := func(dst *int, src *int) {
		if src != nil {
			if *src < 0 {
				*dst = -1
				return
			}
			*dst = *src
		}
	}
	set(&c.XPCorrect, in.XPCorrect)
	set(&c.XPWrong, in.XPWrong)
	set(&c.XPTimeout, in.XPTimeout)
	set(&c.XPMatchComplete, in.XPMatch)
	set(&c.XPWin, in.XPWin)
	set(&c.CoinMatch, in.CoinMatch)
	set(&c.CoinWin, in.CoinWin)
	set(&c.CoinAchievement, in.CoinAchievement)
	if c.XPCorrect < 0 || c.XPWrong < 0 || c.XPTimeout < 0 || c.XPMatchComplete < 0 || c.XPWin < 0 || c.CoinMatch < 0 || c.CoinWin < 0 {
		writeAccountErr(w, http.StatusBadRequest, "reward tidak boleh negatif")
		return
	}
	before := clipJSON(LiveConfig())
	saved := DefaultHub.Ops.PutConfig(c, a.ID)
	a.audit("CONFIG_CHANGED", "config", strconv.Itoa(saved.Version), before, clipJSON(saved), r.RemoteAddr)
	writeJSON(w, http.StatusOK, saved)
}

func adminPutRanks(w http.ResponseWriter, r *http.Request, a *adminCtx) {
	var in struct {
		WinRr  *int `json:"winRr"`
		LossRr *int `json:"lossRr"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	c := LiveConfig()
	if in.WinRr != nil {
		if *in.WinRr < 0 {
			writeAccountErr(w, http.StatusBadRequest, "RR tidak boleh negatif")
			return
		}
		c.RankWinRR = *in.WinRr
	}
	if in.LossRr != nil {
		if *in.LossRr < 0 {
			writeAccountErr(w, http.StatusBadRequest, "RR tidak boleh negatif")
			return
		}
		c.RankLossRR = *in.LossRr
	}
	before := clipJSON(LiveConfig())
	saved := DefaultHub.Ops.PutConfig(c, a.ID)
	a.audit("CONFIG_CHANGED", "rank", strconv.Itoa(saved.Version), before, clipJSON(saved), r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]any{"tiers": RankTiers, "divisions": RankDivisions, "winRr": saved.RankWinRR, "lossRr": saved.RankLossRR, "rrPerDivision": RRPerDivision, "note": "tidak mengubah rank player yang sudah ada"})
}

func normSubject(s string) string {
	s = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", "_"))
	switch s {
	case "PAI":
		return SubjectPAI
	case "MATEMATIKA", "MATH":
		return SubjectMath
	case "BAHASA_INGGRIS", "ENGLISH":
		return SubjectEnglish
	case "BAHASA_JAWA", "JAWA":
		return SubjectJawa
	default:
		return s
	}
}

func similarQuestion(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == "" || b == "" || a == b {
		return false
	}
	if len(a) < 24 || len(b) < 24 {
		return false
	}
	return strings.Contains(a, b) || strings.Contains(b, a)
}
