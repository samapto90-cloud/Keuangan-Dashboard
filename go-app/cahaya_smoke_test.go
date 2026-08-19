package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"keuangan/mmo"
)

func TestCahayaStagingHTTPSmoke(t *testing.T) {
	t.Setenv("CAHAYA_SOCIAL_STORE", filepath.Join(t.TempDir(), "cahaya-social.json"))
	mux := http.NewServeMux()
	registerHealthReady(mux)
	mountCahayaGame(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	get := func(path string) (int, string) {
		t.Helper()
		res, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		raw, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(raw)
	}

	code, body := get("/health")
	if code != 200 {
		t.Fatalf("health %d", code)
	}
	var health map[string]string
	if err := json.Unmarshal([]byte(body), &health); err != nil {
		t.Fatal(err)
	}
	if health["status"] != "ok" || health["version"] != mmo.GameVersion || health["phase"] != mmo.GamePhase {
		t.Fatalf("health body %s", body)
	}

	code, body = get("/ready")
	if code != 200 || !strings.Contains(body, `"status":"ready"`) || !strings.Contains(body, `"writable":true`) {
		t.Fatalf("ready %d %s", code, body)
	}

	code, body = get("/cahaya/")
	if code != 200 {
		t.Fatalf("cahaya %d", code)
	}
	if !strings.Contains(body, "Ular Tangga Nusantara") || !strings.Contains(body, "0.1.0-phase1") {
		t.Fatalf("title missing: %s", body)
	}

	var ok, fail atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := http.Get(srv.URL + "/health")
			if err != nil {
				fail.Add(1)
				return
			}
			res.Body.Close()
			if res.StatusCode == 200 {
				ok.Add(1)
			} else {
				fail.Add(1)
			}
		}()
	}
	wg.Wait()
	if fail.Load() != 0 || ok.Load() != 25 {
		t.Fatalf("concurrent health ok=%d fail=%d", ok.Load(), fail.Load())
	}

	jsPath := ""
	for _, part := range strings.Split(body, `src="`) {
		if strings.Contains(part, "/cahaya/assets/") && strings.Contains(part, ".js") {
			jsPath = strings.Split(part, `"`)[0]
			break
		}
	}
	if jsPath == "" {
		t.Fatal("bundle script missing from index")
	}
	req, err := http.NewRequest(http.MethodGet, srv.URL+jsPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	js, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || len(js) < 1_000 {
		t.Fatalf("bundle %d bytes=%d enc=%s", res.StatusCode, len(js), res.Header.Get("Content-Encoding"))
	}
	if !strings.Contains(string(js), "Ular Tangga") && !strings.Contains(string(js), "PLAY ONLINE") {
		t.Fatal("bundle body truncated or wrong type")
	}
}

func TestCahayaGzipDoesNotTruncateBundle(t *testing.T) {
	mux := http.NewServeMux()
	mountCahayaGame(mux)
	srv := httptest.NewServer(withGzip(mux))
	t.Cleanup(srv.Close)
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/cahaya/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	html := string(raw)
	if res.StatusCode != 200 || !strings.Contains(html, "Ular Tangga Nusantara") {
		t.Fatalf("gzip index %d enc=%s body=%s", res.StatusCode, res.Header.Get("Content-Encoding"), html[:min(200, len(html))])
	}
}
