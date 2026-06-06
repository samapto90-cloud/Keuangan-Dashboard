package main

import (
	"net/http"
	"strings"
)

func rejectIfPortalHeaderMismatch(w http.ResponseWriter, r *http.Request) bool {
	sess := getSession(r)
	if sess == nil {
		return false
	}
	hdr := strings.TrimSpace(r.Header.Get("X-SIPKEU-App"))
	if hdr == "" || hdr == sess.AppModule {
		return false
	}
	if !isValidAppModule(hdr) {
		return false
	}
	jsonResponse(w, http.StatusForbidden, map[string]string{
		"error": "Portal tidak sesuai sesi login. Keluar lalu login ulang di portal yang benar.",
	})
	return true
}

func withPortalSessionMatch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/data/auth/") || path == "/data/portals/status" {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(path, "/data/") {
			next.ServeHTTP(w, r)
			return
		}
		if rejectIfPortalHeaderMismatch(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}
