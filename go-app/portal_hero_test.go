package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPortalHeroDefaultMeta(t *testing.T) {
	dir := t.TempDir()
	prev := dataDir
	dataDir = dir
	t.Cleanup(func() { dataDir = prev; invalidatePortalHeroMetaCache() })
	invalidatePortalHeroMetaCache()

	m := getPortalHeroMeta()
	if m.Source != "default" {
		t.Fatalf("source=%s", m.Source)
	}
	url := portalHeroPublicURL(m)
	if url == "" || url[:len("/assets/portal-hero/")] != "/assets/portal-hero/" {
		t.Fatalf("url=%s", url)
	}
}

func TestPortalHeroCustomServe(t *testing.T) {
	dir := t.TempDir()
	prev := dataDir
	dataDir = dir
	t.Cleanup(func() { dataDir = prev; invalidatePortalHeroMetaCache() })
	invalidatePortalHeroMetaCache()

	if err := os.MkdirAll(portalHeroDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	custom := []byte("fake-mp4-bytes")
	if err := os.WriteFile(portalHeroCustomPath(), custom, 0o644); err != nil {
		t.Fatal(err)
	}
	invalidatePortalHeroMetaCache()
	m := getPortalHeroMeta()
	if m.Source != "custom" || m.SizeBytes != int64(len(custom)) {
		t.Fatalf("meta=%+v", m)
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/portal-hero/portal-intro.mp4", nil)
	rr := httptest.NewRecorder()
	servePortalHeroAssets(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if rr.Body.String() != string(custom) {
		t.Fatalf("body mismatch")
	}
	if rr.Header().Get("X-Portal-Hero-Source") != "custom" {
		t.Fatalf("source header %s", rr.Header().Get("X-Portal-Hero-Source"))
	}
	_ = filepath.Base(portalHeroCustomPath())
}

func TestIsAllowedPortalHeroUpload(t *testing.T) {
	if !isAllowedPortalHeroUpload("hero.mp4", "") {
		t.Fatal("mp4 should allow")
	}
	if isAllowedPortalHeroUpload("hero.webm", "video/webm") {
		t.Fatal("webm should reject")
	}
}

func TestFileLooksLikeMP4(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.mp4")
	bad := filepath.Join(dir, "bad.mp4")
	// Minimal ISO BMFF-like header with ftyp
	payload := make([]byte, 32)
	copy(payload[4:], []byte("ftypisom"))
	if err := os.WriteFile(good, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("not-a-video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !fileLooksLikeMP4(good) {
		t.Fatal("expected ftyp detect")
	}
	if fileLooksLikeMP4(bad) {
		t.Fatal("expected reject")
	}
}
