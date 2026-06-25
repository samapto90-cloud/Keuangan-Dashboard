package main

import (
	"net"
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

func normalizeHostname(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	if strings.HasPrefix(h, "www.") {
		h = h[4:]
	}
	return h
}

func originHostname(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" || origin == "null" {
		return ""
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return normalizeHostname(origin)
	}
	return normalizeHostname(u.Hostname())
}

func effectiveRequestHost(r *http.Request) string {
	if r == nil {
		return ""
	}
	host := strings.TrimSpace(r.Host)
	if trustProxy {
		if xf := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); xf != "" {
			host = strings.TrimSpace(strings.Split(xf, ",")[0])
		}
	}
	return host
}

func requestHostname(r *http.Request) string {
	host := effectiveRequestHost(r)
	if host == "" {
		return ""
	}
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		return normalizeHostname(host)
	}
	return normalizeHostname(h)
}

func requestOriginFromHost(r *http.Request) string {
	if r == nil {
		return ""
	}
	host := effectiveRequestHost(r)
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
	if origin == "" || origin == "null" {
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

	// Same hostname as incoming request (www / port / http vs https variants).
	if rh := requestHostname(r); rh != "" && oh == rh {
		return true
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
