package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func putTestSession(t *testing.T, token string, sess Session) {
	t.Helper()
	sess.Expires = time.Now().Add(time.Hour)
	sess.CreatedAt = time.Now()
	sess.LastSeen = time.Now()
	sessionsMu.Lock()
	sessions[token] = sess
	sessionsMu.Unlock()
	t.Cleanup(func() {
		sessionsMu.Lock()
		delete(sessions, token)
		sessionsMu.Unlock()
	})
}

func TestRequireSettingsAdminForbidsNonSettingsAdmin(t *testing.T) {
	okHandler := requireAuth(requireSettingsAdmin(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]string{"ok": "1"})
	}))

	cases := []struct {
		name   string
		token  string
		sess   *Session
		want   int
	}{
		{name: "no session", want: http.StatusUnauthorized},
		{
			name:  "portal admin",
			token: "tok-admin-sek",
			sess:  &Session{Username: "admin", Role: "admin", AppModule: "sekretariat", Name: "Admin"},
			want:  http.StatusForbidden,
		},
		{
			name:  "operator",
			token: "tok-op",
			sess:  &Session{Username: "op1", Role: "operator", AppModule: "sekretariat", Name: "Op"},
			want:  http.StatusForbidden,
		},
		{
			name:  "settings-admin wrong module",
			token: "tok-sa-wrong",
			sess:  &Session{Username: "sa", Role: "settings-admin", AppModule: "sekretariat", Name: "SA"},
			want:  http.StatusForbidden,
		},
		{
			name:  "settings-admin ok",
			token: "tok-sa-ok",
			sess:  &Session{Username: "sa", Role: "settings-admin", AppModule: "pengaturan", Name: "SA"},
			want:  http.StatusOK,
		},
	}

	endpoints := []string{
		"/data/admin/data-ops",
		"/data/admin/data-ops/export",
		"/data/admin/tahun",
		"/data/admin/portal-hero",
	}

	for _, ep := range endpoints {
		for _, tc := range cases {
			t.Run(ep+"/"+tc.name, func(t *testing.T) {
				if tc.sess != nil {
					putTestSession(t, tc.token, *tc.sess)
				}
				req := httptest.NewRequest(http.MethodGet, ep, nil)
				if tc.token != "" {
					req.Header.Set("Authorization", "Bearer "+tc.token)
				}
				rr := httptest.NewRecorder()
				okHandler(rr, req)
				if rr.Code != tc.want {
					t.Fatalf("status=%d want=%d body=%s", rr.Code, tc.want, rr.Body.String())
				}
			})
		}
	}
}

func TestAdminDataOpsClearRejectsUnknownPortal(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"confirm":            "HAPUS",
		"clear_transactions": true,
		"modules":            []string{"bukan-portal"},
	})
	req := httptest.NewRequest(http.MethodPost, "/data/admin/data-ops/clear", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleAdminDataClear(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d want 400 body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdminPortalHeroUploadRejectsNonMP4Magic(t *testing.T) {
	dir := t.TempDir()
	prev := dataDir
	dataDir = dir
	t.Cleanup(func() {
		dataDir = prev
		invalidatePortalHeroMetaCache()
	})
	invalidatePortalHeroMetaCache()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("video", "fake.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(part, "not-a-real-mp4-file"); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/data/admin/portal-hero", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()
	handleAdminPortalHeroUpload(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d want 400 body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	errMsg := out["error"]
	if errMsg == "" || (!bytes.Contains([]byte(errMsg), []byte("ftyp")) && !bytes.Contains([]byte(errMsg), []byte("MP4"))) {
		t.Fatalf("unexpected error payload %#v", out)
	}
	custom := filepath.Join(dir, "portal-hero", "custom.mp4")
	if _, err := os.Stat(custom); err == nil {
		t.Fatal("custom.mp4 should not remain after rejected upload")
	}
}

func TestAdminPortalHeroUploadAcceptsFtypMP4(t *testing.T) {
	dir := t.TempDir()
	prev := dataDir
	dataDir = dir
	t.Cleanup(func() {
		dataDir = prev
		invalidatePortalHeroMetaCache()
	})
	invalidatePortalHeroMetaCache()

	payload := make([]byte, 32)
	copy(payload[4:], []byte("ftypisom"))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("video", "ok.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/data/admin/portal-hero", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()
	handleAdminPortalHeroUpload(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d want 200 body=%s", rr.Code, rr.Body.String())
	}
	custom := filepath.Join(dir, "portal-hero", "custom.mp4")
	if st, err := os.Stat(custom); err != nil || st.Size() == 0 {
		t.Fatalf("expected custom.mp4 written: %v", err)
	}
}
