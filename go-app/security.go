package main

import (
	"compress/gzip"
	"crypto/subtle"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	maxRequestBodyBytes = 12 << 20 // 12 MiB (impor Excel)
	loginMaxFails       = 5
	loginLockDuration   = 15 * time.Minute
)

type loginGuard struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

type loginAttempt struct {
	fails       int
	lockedUntil time.Time
}

var loginLimiter loginGuard

func initSecurity() {
	loginLimiter.attempts = map[string]*loginAttempt{}
	go purgeExpiredSessionsLoop()
	warnWeakProductionSecrets()
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func (g *loginGuard) check(ip string) (blocked bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	a := g.attempts[ip]
	if a == nil {
		return false
	}
	if time.Now().Before(a.lockedUntil) {
		return true
	}
	if a.fails >= loginMaxFails {
		a.fails = 0
	}
	return false
}

func (g *loginGuard) recordFail(ip string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	a := g.attempts[ip]
	if a == nil {
		a = &loginAttempt{}
		g.attempts[ip] = a
	}
	a.fails++
	if a.fails >= loginMaxFails {
		a.lockedUntil = time.Now().Add(loginLockDuration)
	}
}

func (g *loginGuard) recordSuccess(ip string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.attempts, ip)
}

func passwordMatches(stored, input string) bool {
	stored = strings.TrimSpace(stored)
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(input)) == nil
	}
	if len(stored) != len(input) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(stored), []byte(input)) == 1
}

func sessionLifetime() time.Duration {
	raw := strings.TrimSpace(os.Getenv("SIPKEU_SESSION_HOURS"))
	if raw == "" {
		return 8 * time.Hour
	}
	hours, err := strconv.ParseFloat(raw, 64)
	if err != nil || hours <= 0 || hours > 24 {
		return 8 * time.Hour
	}
	return time.Duration(hours * float64(time.Hour))
}

func warnWeakProductionSecrets() {
	if strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN")) == "" {
		return
	}
	weak := map[string]bool{
		"admin2026": true, "operator2026": true, "admin": true, "password": true,
		"123456": true, "operator": true,
	}
	check := func(name, val string) {
		if weak[strings.TrimSpace(val)] {
			log.Printf("PERINGATAN KEAMANAN: %s masih password default/lelemah — ganti di .env server.", name)
		}
	}
	check("SIPKEU_ADMIN_PASSWORD", envOr("SIPKEU_ADMIN_PASSWORD", "admin2026"))
	check("SIPKEU_OPERATOR_PASSWORD", envOr("SIPKEU_OPERATOR_PASSWORD", "operator2026"))
}

func purgeExpiredSessionsLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		now := time.Now()
		sessionsMu.Lock()
		for token, sess := range sessions {
			if now.After(sess.Expires) {
				delete(sessions, token)
			}
		}
		sessionsMu.Unlock()
	}
}

func withMaxBody(max int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && max > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, max)
		}
		next.ServeHTTP(w, r)
	})
}

func securityCSP() string {
	return strings.Join([]string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com",
		"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com",
		"font-src 'self' https://cdn.jsdelivr.net https://fonts.gstatic.com data:",
		"img-src 'self' data: blob:",
		"connect-src 'self'",
		"frame-ancestors 'self'",
		"base-uri 'self'",
		"form-action 'self'",
		"object-src 'none'",
	}, "; ")
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
		w.Header().Set("Content-Security-Policy", securityCSP())
		if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") || r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gw          *gzip.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.gw.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func withGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		gw := gzip.NewWriter(w)
		defer gw.Close()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, gw: gw}, r)
	})
}
