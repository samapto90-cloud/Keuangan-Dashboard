package main

import (
	"net/http/httptest"
	"testing"
)

func TestOriginAllowed(t *testing.T) {
	allowed := expandAllowedOrigins(parseAllowedOrigins("https://sakubijak.com"))
	req := httptest.NewRequest("POST", "https://sakubijak.com:8888/data/auth/login", nil)
	req.Host = "sakubijak.com:8888"
	req.Header.Set("X-Forwarded-Proto", "https")

	cases := []struct {
		origin string
		want   bool
	}{
		{"https://sakubijak.com", true},
		{"https://sakubijak.com:8888", true},
		{"https://www.sakubijak.com", true},
		{"https://www.sakubijak.com:8888", true},
		{"http://localhost:3000", false},
		{"", true},
	}
	for _, c := range cases {
		got := originAllowed(c.origin, allowed, req)
		if got != c.want {
			t.Fatalf("originAllowed(%q) = %v, want %v", c.origin, got, c.want)
		}
	}
}

func TestOriginNullWithReferer(t *testing.T) {
	allowed := expandAllowedOrigins(parseAllowedOrigins("https://sakubijak.com"))
	req := httptest.NewRequest("POST", "https://sakubijak.com:8888/data/auth/login", nil)
	req.Host = "sakubijak.com:8888"
	req.Header.Set("Referer", "https://sakubijak.com:8888/")
	req.Header.Set("X-Forwarded-Proto", "https")

	if !originAllowed("null", allowed, req) {
		t.Fatal("null origin with valid referer should be allowed")
	}
	req2 := httptest.NewRequest("POST", "https://evil.example/data/auth/login", nil)
	if originAllowed("null", allowed, req2) {
		t.Fatal("null origin without trusted referer should be rejected")
	}
}

func TestExpandAllowedOrigins(t *testing.T) {
	got := expandAllowedOrigins(parseAllowedOrigins("https://sakubijak.com"))
	if len(got) < 4 {
		t.Fatalf("expected expanded origins, got %d: %v", len(got), got)
	}
}

func TestOriginAllowedProxyForwardedHost(t *testing.T) {
	trustProxy = true
	t.Cleanup(func() { trustProxy = false })
	allowed := expandAllowedOrigins(parseAllowedOrigins("https://sakubijak.com"))
	req := httptest.NewRequest("POST", "http://127.0.0.1:8888/data/auth/login", nil)
	req.Host = "127.0.0.1:8888"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "sakubijak.com")

	if !originAllowed("https://sakubijak.com", allowed, req) {
		t.Fatal("expected https://sakubijak.com allowed via X-Forwarded-Host proxy path")
	}
}

func TestOriginAllowedLocalhostDev(t *testing.T) {
	t.Setenv("SIPKEU_ALLOW_LOCALHOST", "1")
	allowed := expandAllowedOrigins(parseAllowedOrigins("https://sakubijak.com"))
	if !originAllowed("http://localhost:3000", allowed, nil) {
		t.Fatal("localhost should be allowed when SIPKEU_ALLOW_LOCALHOST=1")
	}
	if !originAllowed("http://127.0.0.1:8888", allowed, nil) {
		t.Fatal("127.0.0.1 should be allowed when SIPKEU_ALLOW_LOCALHOST=1")
	}
}

func TestMutatingRequestBlocksCrossSite(t *testing.T) {
	allowed := expandAllowedOrigins(parseAllowedOrigins("https://sakubijak.com"))
	req := httptest.NewRequest("POST", "https://sakubijak.com:8888/data/auth/login", nil)
	req.Host = "sakubijak.com:8888"
	req.Header.Set("Origin", "https://sakubijak.com:8888")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	if mutatingRequestSafe(req, allowed) {
		t.Fatal("cross-site POST should be blocked")
	}
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	if !mutatingRequestSafe(req, allowed) {
		t.Fatal("same-origin POST should be allowed")
	}
}

func TestNormalizeHostname(t *testing.T) {
	if normalizeHostname("WWW.Sakubijak.com") != "sakubijak.com" {
		t.Fatal("www prefix should be stripped")
	}
}
