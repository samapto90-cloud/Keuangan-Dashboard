package mmo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAccountHTTPRegisterLoginLogout(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CAHAYA_ACCOUNT_STORE", filepath.Join(dir, "accounts.json"))
	DefaultHub.Accounts = OpenAccountStore(filepath.Join(dir, "accounts.json"))
	DefaultHub.Progress = OpenProgressStore(filepath.Join(dir, "progress.json"))

	reg, _ := json.Marshal(map[string]string{
		"username": "UlarSatu", "email": "ular1@example.com",
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

	login, _ := json.Marshal(map[string]string{"username": "UlarSatu", "password": "Rahasia1"})
	req = httptest.NewRequest(http.MethodPost, "/cahaya/api/login", bytes.NewReader(login))
	rec = httptest.NewRecorder()
	HandleLogin(rec, req)
	if rec.Code != 200 {
		t.Fatalf("login %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/cahaya/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+sess.Token)
	rec = httptest.NewRecorder()
	HandleProfile(rec, req)
	if rec.Code != 200 {
		t.Fatalf("profile %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/cahaya/api/logout", bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+sess.Token)
	rec = httptest.NewRecorder()
	HandleLogout(rec, req)
	if rec.Code != 200 {
		t.Fatalf("logout %d", rec.Code)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "accounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("Rahasia1")) {
		t.Fatal("password plaintext")
	}
}
