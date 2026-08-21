package mmo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func adminSetup(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("ULAR_ADMIN_TOKEN", "phase8-admin-token")
	t.Setenv("ULAR_OPS_STORE", filepath.Join(dir, "ops.json"))
	t.Setenv("CAHAYA_ACCOUNT_STORE", filepath.Join(dir, "accounts.json"))
	t.Setenv("ULAR_PROGRESS_STORE", filepath.Join(dir, "progress.json"))
	t.Setenv("CAHAYA_SOCIAL_STORE", filepath.Join(dir, "social.json"))
	t.Setenv("ULAR_MATCH_STORE", filepath.Join(dir, "matches.json"))
	DefaultHub.Accounts = OpenAccountStore(filepath.Join(dir, "accounts.json"))
	DefaultHub.Progress = OpenProgressStore(filepath.Join(dir, "progress.json"))
	DefaultHub.Ops = OpenOpsStore(filepath.Join(dir, "ops.json"))
	DefaultHub.Social = OpenSocialStore(filepath.Join(dir, "social.json"))
	DefaultHub.Lobby = NewUlarLobby()
}

func adminReq(t *testing.T, method, path string, body any, token, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *bytes.Reader
	if body == nil {
		rdr = bytes.NewReader(nil)
	} else if s, ok := body.(string); ok {
		rdr = bytes.NewReader([]byte(s))
	} else {
		raw, _ := json.Marshal(body)
		rdr = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, rdr)
	if token != "" {
		req.Header.Set("X-Ular-Admin", token)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if body != nil {
		if _, ok := body.(string); !ok {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	rec := httptest.NewRecorder()
	HandleAdminAPI(rec, req)
	return rec
}

func registerNamed(t *testing.T, user, email string) sessionOut {
	t.Helper()
	reg, _ := json.Marshal(map[string]string{
		"username": user, "email": email, "password": "Rahasia1", "confirmPassword": "Rahasia1",
	})
	req := httptest.NewRequest(http.MethodPost, "/cahaya/api/register", bytes.NewReader(reg))
	req.RemoteAddr = t.Name() + ":1"
	rec := httptest.NewRecorder()
	HandleRegister(rec, req)
	if rec.Code != 200 {
		t.Fatalf("register %s %d %s", user, rec.Code, rec.Body.String())
	}
	var sess sessionOut
	if json.Unmarshal(rec.Body.Bytes(), &sess) != nil || sess.Token == "" {
		t.Fatal("session")
	}
	return sess
}

func sampleQuestion(id, text string) EduQuestion {
	src := DefaultEduBank().ListAdmin("", "", "")[0]
	src.ID = id
	src.Question = text
	src.Active = true
	src.Deleted = false
	return src
}

func TestAdminQuestionCRUDAndAudit(t *testing.T) {
	adminSetup(t)
	q := sampleQuestion("q-p8-create", "Apa lambang negara Indonesia pada tes fase delapan CRUD?")
	rec := adminReq(t, http.MethodPost, "/cahaya/api/admin/questions", q, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	got, ok := DefaultEduBank().Get("q-p8-create")
	if !ok || got.Question != q.Question {
		t.Fatal("not saved")
	}
	q.Explanation = "Garuda Pancasila. " + q.Explanation
	rec = adminReq(t, http.MethodPut, "/cahaya/api/admin/questions/q-p8-create", q, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("edit %d %s", rec.Code, rec.Body.String())
	}
	logs, _ := DefaultHub.Ops.AuditPage(0, 20, "QUESTION_UPDATED")
	if len(logs) == 0 || logs[0].AdminID == "" || logs[0].TargetID != "q-p8-create" || logs[0].CreatedAt == 0 {
		t.Fatalf("audit %+v", logs)
	}
	q.Active = false
	rec = adminReq(t, http.MethodPut, "/cahaya/api/admin/questions/q-p8-create", q, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("deactivate %d %s", rec.Code, rec.Body.String())
	}
	for _, it := range DefaultEduBank().List("", "", "", false) {
		if it.ID == "q-p8-create" {
			t.Fatal("inactive leaked into match pool")
		}
	}
	if !DefaultEduBank().Delete("q-p8-create") {
		t.Fatal("cleanup")
	}
}

func TestAdminQuestionImportPreviewAndBlock(t *testing.T) {
	adminSetup(t)
	uniq := "Soal import fase delapan unik " + shortID()
	csv := "subject,grade,difficulty,question,optionA,optionB,optionC,optionD,correctAnswer,explanation\n" +
		"PAI,SMA,EASY," + uniq + ",A1,B1,C1,D1,A,Penjelasan wajib ada.\n" +
		"PAI,SMA,EASY,,A1,B1,C1,D1,A,Penjelasan wajib ada.\n" +
		"PAI,SMA,EASY," + uniq + ",A1,B1,C1,D1,A,Penjelasan wajib ada.\n"
	rec := adminReq(t, http.MethodPost, "/cahaya/api/admin/questions/import", csv, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("preview %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Preview bool `json:"preview"`
		Rows    []struct {
			Status string `json:"status"`
		} `json:"rows"`
	}
	if json.Unmarshal(rec.Body.Bytes(), &out) != nil || !out.Preview || len(out.Rows) != 3 {
		t.Fatalf("preview body %s", rec.Body.String())
	}
	if out.Rows[0].Status != "VALID" || out.Rows[1].Status != "INVALID" || out.Rows[2].Status != "DUPLICATE" {
		t.Fatalf("rows %+v", out.Rows)
	}
	rec = adminReq(t, http.MethodPost, "/cahaya/api/admin/questions/import?commit=1", csv, "phase8-admin-token", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("commit invalid want 400 got %d %s", rec.Code, rec.Body.String())
	}
}

func TestAdminPlayerBanBlocksLoginAndMatch(t *testing.T) {
	adminSetup(t)
	sess := registerNamed(t, "BanTarget", "ban@example.com")
	rec := adminReq(t, http.MethodPost, "/cahaya/api/admin/players/"+sess.PlayerID+"/sanction", map[string]any{
		"type": "PERMANENT_BAN", "reason": "uji ban", "confirm": "PERMANENT BAN",
	}, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("ban %d %s", rec.Code, rec.Body.String())
	}
	login, _ := json.Marshal(map[string]string{"username": "BanTarget", "password": "Rahasia1"})
	req := httptest.NewRequest(http.MethodPost, "/cahaya/api/login", bytes.NewReader(login))
	req.RemoteAddr = t.Name() + "-login:1"
	lr := httptest.NewRecorder()
	HandleLogin(lr, req)
	if lr.Code == 200 {
		t.Fatal("banned player still logged in")
	}
	p := testPlayer(sess.PlayerID, "BanTarget")
	if msg := DefaultHub.queueJoin(p, "CASUAL", "", "", 2); msg == "" {
		t.Fatal("banned joined matchmaking")
	}
	if _, msg := DefaultHub.Lobby.Create(p); msg == "" {
		t.Fatal("banned created room")
	}
	rec = adminReq(t, http.MethodPost, "/cahaya/api/admin/players/"+sess.PlayerID+"/unban", map[string]string{}, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("unban %d %s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "/cahaya/api/login", bytes.NewReader(login))
	req.RemoteAddr = t.Name() + "-login2:1"
	lr = httptest.NewRecorder()
	HandleLogin(lr, req)
	if lr.Code != 200 {
		t.Fatalf("unban login %d %s", lr.Code, lr.Body.String())
	}
}

func TestPlayerForbiddenOnAdmin(t *testing.T) {
	adminSetup(t)
	sess := registerNamed(t, "PlayerOnly", "player@example.com")
	rec := adminReq(t, http.MethodGet, "/cahaya/api/admin/questions", nil, "", sess.Token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d %s", rec.Code, rec.Body.String())
	}
	req := httptest.NewRequest(http.MethodGet, "/cahaya/api/ular/admin/questions", nil)
	req.Header.Set("Authorization", "Bearer "+sess.Token)
	rr := httptest.NewRecorder()
	HandleAdminQuestions(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("legacy want 403 got %d", rr.Code)
	}
}

func TestModeratorCannotEndSeason(t *testing.T) {
	adminSetup(t)
	sess := registerNamed(t, "ModUser", "mod@example.com")
	if msg := DefaultHub.Accounts.SetRole(sess.PlayerID, RoleModerator); msg != "" {
		t.Fatal(msg)
	}
	rec := adminReq(t, http.MethodPost, "/cahaya/api/admin/seasons/season-1/end", map[string]string{"confirm": "END SEASON"}, "", sess.Token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d %s", rec.Code, rec.Body.String())
	}
}

func TestAdminConfigVersionRollbackAndFreeze(t *testing.T) {
	adminSetup(t)
	c := DefaultGameConfig()
	c.QuestionTimeLimit = 30
	rec := adminReq(t, http.MethodPut, "/cahaya/api/admin/config", c, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("v1 %d %s", rec.Code, rec.Body.String())
	}
	if LiveQuestionTime() != 30*time.Second {
		t.Fatal(LiveQuestionTime())
	}
	prev := ularCountdown
	ularCountdown = time.Millisecond
	t.Cleanup(func() { ularCountdown = prev })
	a := testPlayer("cfg-a", "Andi")
	DefaultHub.Lobby.Connect(a)
	room, errc := DefaultHub.Lobby.Create(a)
	if errc != "" {
		t.Fatal(errc)
	}
	for i, name := range []string{"Budi", "Citra", "Deni"} {
		p := testPlayer("cfg-"+string(rune('b'+i)), name)
		DefaultHub.Lobby.Connect(p)
		if _, errc := DefaultHub.Lobby.Join(p, room.RoomCode); errc != "" {
			t.Fatal(errc)
		}
	}
	for _, id := range []string{"cfg-a", "cfg-b", "cfg-c", "cfg-d"} {
		if _, errc := DefaultHub.Lobby.SetReady(id, true); errc != "" {
			t.Fatal(errc)
		}
	}
	if errc := DefaultHub.startMatch("cfg-a"); errc != "" {
		t.Fatal(errc)
	}
	DefaultHub.Lobby.mu.Lock()
	frozen := room.Match.QuestionLimit
	mid := room.Match.ID
	DefaultHub.Lobby.mu.Unlock()
	if frozen != 30*time.Second {
		t.Fatalf("frozen %s", frozen)
	}
	c.QuestionTimeLimit = 20
	rec = adminReq(t, http.MethodPut, "/cahaya/api/admin/config", c, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("v2 %d %s", rec.Code, rec.Body.String())
	}
	if LiveQuestionTime() != 20*time.Second {
		t.Fatal("new config not live")
	}
	DefaultHub.Lobby.mu.Lock()
	still := room.Match.QuestionLimit
	DefaultHub.Lobby.mu.Unlock()
	if still != 30*time.Second {
		t.Fatalf("old match mutated %s", still)
	}
	rec = adminReq(t, http.MethodPost, "/cahaya/api/admin/config/rollback", map[string]int{"version": 1}, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("rollback %d %s", rec.Code, rec.Body.String())
	}
	if DefaultHub.Ops.ActiveConfig().Version != 1 || LiveQuestionTime() != 30*time.Second {
		t.Fatalf("active %+v", DefaultHub.Ops.ActiveConfig())
	}
	_ = mid
}

func TestAdminFeatureFlagRanked(t *testing.T) {
	adminSetup(t)
	rec := adminReq(t, http.MethodPut, "/cahaya/api/admin/flags", FeatureFlags{EnableDailyReward: true, EnableChat: true}, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("flags %d %s", rec.Code, rec.Body.String())
	}
	if LiveFlags().EnableRanked {
		t.Fatal("ranked still on")
	}
	p := testPlayer("flag-u", "Andi")
	if msg := DefaultHub.queueJoin(p, "RANKED", "", "", 2); msg == "" {
		t.Fatal("ranked started while disabled")
	}
}

func TestAdminTerminateRequiresConfirmAndSaves(t *testing.T) {
	adminSetup(t)
	prev := ularCountdown
	ularCountdown = time.Millisecond
	t.Cleanup(func() { ularCountdown = prev })
	a := testPlayer("tm-a", "Andi")
	DefaultHub.Lobby.Connect(a)
	room, _ := DefaultHub.Lobby.Create(a)
	for i, name := range []string{"Budi", "Citra", "Deni"} {
		p := testPlayer("tm-"+string(rune('b'+i)), name)
		DefaultHub.Lobby.Connect(p)
		_, _ = DefaultHub.Lobby.Join(p, room.RoomCode)
	}
	for _, id := range []string{"tm-a", "tm-b", "tm-c", "tm-d"} {
		_, _ = DefaultHub.Lobby.SetReady(id, true)
	}
	if errc := DefaultHub.startMatch("tm-a"); errc != "" {
		t.Fatal(errc)
	}
	DefaultHub.Lobby.mu.Lock()
	mid := room.Match.ID
	DefaultHub.Lobby.mu.Unlock()
	rec := adminReq(t, http.MethodPost, "/cahaya/api/admin/matches/"+mid+"/terminate", map[string]string{"reason": "darurat"}, "phase8-admin-token", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("no confirm %d %s", rec.Code, rec.Body.String())
	}
	rec = adminReq(t, http.MethodPost, "/cahaya/api/admin/matches/"+mid+"/terminate", map[string]string{"reason": "darurat", "confirm": "TERMINATE"}, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("terminate %d %s", rec.Code, rec.Body.String())
	}
	found := false
	for _, m := range DefaultHub.Lobby.Store.All() {
		if m.ID == mid {
			found = true
			if m.Status != string(UlarTerminated) {
				t.Fatalf("status %s", m.Status)
			}
		}
	}
	if !found {
		t.Fatal("match not persisted")
	}
}

func TestAdminReportResolveAudited(t *testing.T) {
	adminSetup(t)
	a := registerNamed(t, "RepA", "ra@example.com")
	b := registerNamed(t, "RepB", "rb@example.com")
	rep, msg := DefaultHub.Social.Report(a.PlayerID, b.PlayerID, "Spam", "uji laporan", "")
	if msg != "" {
		t.Fatal(msg)
	}
	rec := adminReq(t, http.MethodPost, "/cahaya/api/admin/reports/"+rep.ID+"/resolve", map[string]string{
		"status": ReportResolved, "note": "sudah ditindak",
	}, "phase8-admin-token", "")
	if rec.Code != 200 {
		t.Fatalf("resolve %d %s", rec.Code, rec.Body.String())
	}
	logs, _ := DefaultHub.Ops.AuditPage(0, 20, "REPORT_RESOLVED")
	if len(logs) == 0 || logs[0].AdminID != "env-token" || logs[0].TargetID != rep.ID {
		t.Fatalf("audit %+v", logs)
	}
}

func TestAdminMatchViewHasNoDiceEndpoint(t *testing.T) {
	adminSetup(t)
	rec := adminReq(t, http.MethodPost, "/cahaya/api/admin/matches/x/roll", map[string]int{"value": 6}, "phase8-admin-token", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 got %d %s", rec.Code, rec.Body.String())
	}
}
