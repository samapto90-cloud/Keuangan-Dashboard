package mmo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func progressHTTPSetup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	DefaultHub.Accounts = OpenAccountStore(filepath.Join(dir, "accounts.json"))
	DefaultHub.Progress = OpenProgressStore(filepath.Join(dir, "progress.json"))
	reg, _ := json.Marshal(map[string]string{
		"username": "PemainSatu", "email": "p1@example.com",
		"password": "Rahasia1", "confirmPassword": "Rahasia1",
	})
	req := httptest.NewRequest(http.MethodPost, "/cahaya/api/register", bytes.NewReader(reg))
	rec := httptest.NewRecorder()
	HandleRegister(rec, req)
	if rec.Code != 200 {
		t.Fatalf("register %d %s", rec.Code, rec.Body.String())
	}
	var sess sessionOut
	if json.Unmarshal(rec.Body.Bytes(), &sess) != nil || sess.Token == "" {
		t.Fatal("session")
	}
	return sess.Token
}

func TestProfileAPINoSecretsAndDefaults(t *testing.T) {
	tok := progressHTTPSetup(t)
	req := httptest.NewRequest(http.MethodGet, "/cahaya/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	HandleProfile(rec, req)
	if rec.Code != 200 {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "password") || strings.Contains(body, "PasswordHash") || strings.Contains(body, "token") {
		t.Fatal("secrets leaked")
	}
	var view map[string]any
	if json.Unmarshal(rec.Body.Bytes(), &view) != nil {
		t.Fatal("json")
	}
	if view["level"] != float64(1) || view["xp"] != float64(0) || view["coins"] != float64(0) {
		t.Fatalf("%v", view)
	}
}

func TestProfileUpdateRejectsClientXP(t *testing.T) {
	tok := progressHTTPSetup(t)
	body, _ := json.Marshal(map[string]any{"xpAmount": 999999, "coinAmount": 999999, "level": 100, "achievementUnlocked": "CENTURY"})
	req := httptest.NewRequest(http.MethodPost, "/cahaya/api/profile/update", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	HandleProgressUpdate(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d %s", rec.Code, rec.Body.String())
	}
	sess := DefaultHub.Accounts.Lookup(tok)
	p, ok := DefaultHub.Progress.Get(sess.PlayerID)
	if ok && (p.XP != 0 || p.Coins != 0 || p.Level != 1) {
		t.Fatalf("mutated %+v", p)
	}
}

func TestDailyRewardHTTPOnce(t *testing.T) {
	tok := progressHTTPSetup(t)
	req := httptest.NewRequest(http.MethodPost, "/cahaya/api/daily-reward/claim", bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	HandleDailyClaim(rec, req)
	if rec.Code != 200 {
		t.Fatalf("claim %d %s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "/cahaya/api/daily-reward/claim", bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	HandleDailyClaim(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("second %d %s", rec.Code, rec.Body.String())
	}
}

func TestHistoryPagination(t *testing.T) {
	tok := progressHTTPSetup(t)
	sess := DefaultHub.Accounts.Lookup(tok)
	for i := 0; i < 21; i++ {
		_, msg := DefaultHub.Progress.SettleMatch(MatchSettlement{
			MatchID: fmt.Sprintf("mh-%d", i),
			Players: []MatchSettlePlayer{{UserID: sess.PlayerID, Username: sess.Username, Rank: 2, Position: 10}},
		})
		if msg != "" {
			t.Fatal(msg)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/cahaya/api/profile/history?page=0", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	HandleProgressHistory(rec, req)
	var out struct {
		Items []MatchPlayerResult `json:"items"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if len(out.Items) != MaxMatchHistoryPage {
		t.Fatalf("page0 %d", len(out.Items))
	}
}

func TestUsernameLengthRejected(t *testing.T) {
	tok := progressHTTPSetup(t)
	body, _ := json.Marshal(map[string]string{"username": "ab"})
	req := httptest.NewRequest(http.MethodPost, "/cahaya/api/profile/update", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	HandleProgressUpdate(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}
