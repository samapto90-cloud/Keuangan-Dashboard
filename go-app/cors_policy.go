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

// expandAllowedOrigins menambah varian www dan port :8888 agar login konsisten
// meskipun .env server hanya berisi satu URL dasar.
func expandAllowedOrigins(origins []string) []string {
	if len(origins) == 0 {
		return origins
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(origins)*4)
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, o := range origins {
		add(o)
		u, err := url.Parse(o)
		if err != nil || u.Hostname() == "" {
			continue
		}
		scheme := u.Scheme
		if scheme == "" {
			scheme = "https"
		}
		base := normalizeHostname(u.Hostname())
		if base == "" || base == "localhost" || base == "127.0.0.1" {
			continue
		}
		for _, h := range []string{base, "www." + base} {
			for _, port := range []string{"", ":8888"} {
				add(scheme + "://" + h + port)
			}
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
	if err != nil || u.Hostname() == "" {
		return normalizeHostname(origin)
	}
	return normalizeHostname(u.Hostname())
}

func refererHostname(r *http.Request) string {
	if r == nil {
		return ""
	}
	ref := strings.TrimSpace(r.Header.Get("Referer"))
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil || u.Hostname() == "" {
		return ""
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

func hostnameInAllowedList(host string, allowed []string) bool {
	host = normalizeHostname(host)
	if host == "" {
		return false
	}
	for _, a := range allowed {
		if host == originHostname(a) {
			return true
		}
	}
	return false
}

func originAllowed(origin string, allowed []string, r *http.Request) bool {
	if len(allowed) == 0 {
		return true
	}
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return true
	}
	if origin == "null" {
		if rh := refererHostname(r); rh != "" && hostnameInAllowedList(rh, allowed) {
			return true
		}
		if rh := requestHostname(r); rh != "" && hostnameInAllowedList(rh, allowed) {
			return true
		}
		return false
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

func secFetchSiteSafe(r *http.Request) bool {
	sfs := strings.ToLower(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site")))
	if sfs == "" {
		return true
	}
	return sfs == "same-origin" || sfs == "same-site"
}

func mutatingRequestSafe(r *http.Request, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
	default:
		return true
	}
	if !secFetchSiteSafe(r) {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin != "" && origin != "null" {
		return originAllowed(origin, allowed, r)
	}
	if origin == "null" {
		return originAllowed(origin, allowed, r)
	}
	if rh := refererHostname(r); rh != "" {
		return hostnameInAllowedList(rh, allowed)
	}
	rh := requestHostname(r)
	return rh != "" && hostnameInAllowedList(rh, allowed)
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
