package main

import (
	"net/http/httptest"
	"testing"
)

func TestOriginAllowed(t *testing.T) {
	allowed := parseAllowedOrigins("https://sakubijak.com")
	req := httptest.NewRequest("POST", "https://sakubijak.com:8888/data/auth/login", nil)
	req.Host = "sakubijak.com:8888"
	req.Header.Set("X-Forwarded-Proto", "https")

	cases := []struct {
		origin string
		want   bool
	}{
		{"https://sakubijak.com", true},
		{"https://sakubijak.com:8888", true},
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

func TestOriginAllowedLocalhostDev(t *testing.T) {
	t.Setenv("SIPKEU_ALLOW_LOCALHOST", "1")
	allowed := parseAllowedOrigins("https://sakubijak.com")
	if !originAllowed("http://localhost:3000", allowed, nil) {
		t.Fatal("localhost should be allowed when SIPKEU_ALLOW_LOCALHOST=1")
	}
	if !originAllowed("http://127.0.0.1:8888", allowed, nil) {
		t.Fatal("127.0.0.1 should be allowed when SIPKEU_ALLOW_LOCALHOST=1")
	}
}

func TestParseAllowedOriginsMultiple(t *testing.T) {
	got := parseAllowedOrigins("https://sakubijak.com, https://sakubijak.com:8888")
	if len(got) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(got))
	}
}
