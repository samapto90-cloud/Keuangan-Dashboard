package main

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

func parseAllowedOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func originHostname(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return strings.ToLower(origin)
	}
	return strings.ToLower(u.Hostname())
}

func requestOriginFromHost(r *http.Request) string {
	if r == nil {
		return ""
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = strings.ToLower(strings.Split(proto, ",")[0])
	}
	return scheme + "://" + host
}

func originAllowed(origin string, allowed []string, r *http.Request) bool {
	if len(allowed) == 0 {
		return true
	}
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return true
	}
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	reqOrigin := requestOriginFromHost(r)
	if reqOrigin != "" && origin == reqOrigin {
		return true
	}
	oh := originHostname(origin)
	if oh == "" {
		return false
	}
	for _, a := range allowed {
		if oh == originHostname(a) && oh != "" {
			return true
		}
	}
	if localOriginAllowed(origin) {
		return true
	}
	return false
}

func localOriginAllowed(origin string) bool {
	if strings.TrimSpace(os.Getenv("SIPKEU_ALLOW_LOCALHOST")) != "1" {
		return false
	}
	h := originHostname(origin)
	return h == "localhost" || h == "127.0.0.1"
}

func corsAllowOriginHeader(origin string, allowed []string, r *http.Request) string {
	origin = strings.TrimSpace(origin)
	if origin != "" && originAllowed(origin, allowed, r) {
		return origin
	}
	if len(allowed) == 1 {
		return allowed[0]
	}
	reqOrigin := requestOriginFromHost(r)
	if reqOrigin != "" && originAllowed(reqOrigin, allowed, r) {
		return reqOrigin
	}
	if len(allowed) > 0 {
		return allowed[0]
	}
	return ""
}
