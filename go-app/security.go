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
	maxLoginBodyBytes   = 8 << 10  // 8 KiB
	loginMaxFails       = 5
	loginLockDuration   = 15 * time.Minute
	defaultAPIRateMax    = 2400
	defaultAPIRateWin    = time.Minute
	defaultLoginRateMax  = 120
	defaultIPRateMax     = 1200
	defaultAssetRateMax  = 3000
	defaultPortalRateMax = 240
	defaultMaxConnPerIP  = 64
	loginRateWindow     = time.Minute
	bcryptCost          = 10
	maxBcryptConcurrent = 160
)

type loginGuard struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

type loginAttempt struct {
	fails       int
	lockedUntil time.Time
}

type apiRateBucket struct {
	count       int
	windowStart time.Time
}

type loginRateGuard struct {
	mu      sync.Mutex
	buckets map[string]*apiRateBucket
}

var (
	loginLimiter     loginGuard
	loginRateLimiter loginRateGuard
	trustProxy       bool
	apiRateLimitMax  int
	apiRateWindow    time.Duration
	loginRateMax     int
	ipRateLimitMax   int
	assetRateLimitMax int
	portalStatusRateMax int
	maxConnPerIP     int
	bcryptSem        = make(chan struct{}, maxBcryptConcurrent)
)

func initSecurity() {
	loginLimiter.attempts = map[string]*loginAttempt{}
	loginRateLimiter.buckets = map[string]*apiRateBucket{}
	trustProxy = strings.TrimSpace(os.Getenv("SIPKEU_TRUST_PROXY")) == "1"
	apiRateLimitMax = securityEnvInt("SIPKEU_API_RATE_LIMIT", defaultAPIRateMax)
	if apiRateLimitMax < 120 {
		apiRateLimitMax = 120
	}
	apiRateWindow = defaultAPIRateWin
	loginRateMax = securityEnvInt("SIPKEU_LOGIN_RATE_LIMIT", defaultLoginRateMax)
	if loginRateMax < 15 {
		loginRateMax = 15
	}
	ipRateLimitMax = securityEnvInt("SIPKEU_IP_RATE_LIMIT", defaultIPRateMax)
	if ipRateLimitMax < 120 {
		ipRateLimitMax = 120
	}
	assetRateLimitMax = securityEnvInt("SIPKEU_ASSET_RATE_LIMIT", defaultAssetRateMax)
	if assetRateLimitMax < 300 {
		assetRateLimitMax = 300
	}
	portalStatusRateMax = securityEnvInt("SIPKEU_PORTAL_STATUS_RATE", defaultPortalRateMax)
	if portalStatusRateMax < 30 {
		portalStatusRateMax = 30
	}
	maxConnPerIP = securityEnvInt("SIPKEU_MAX_CONN_PER_IP", defaultMaxConnPerIP)
	if maxConnPerIP < 8 {
		maxConnPerIP = 8
	}
	go purgeExpiredSessionsLoop()
	go purgeLoginAttemptsLoop()
	go purgeLoginRateBucketsLoop()
	go purgeShardedRateGuardsLoop()
	initGlobalConnLimit()
	warnWeakProductionSecrets()
}

func securityEnvInt(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func hashPasswordStore(plain string) (string, error) {
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", nil
	}
	if strings.HasPrefix(plain, "$2a$") || strings.HasPrefix(plain, "$2b$") || strings.HasPrefix(plain, "$2y$") {
		return plain, nil
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func storePasswordIfProvided(current, incoming string) string {
	in := strings.TrimSpace(incoming)
	if in == "" || in == passwordMask {
		return current
	}
	hashed, err := hashPasswordStore(in)
	if err != nil {
		log.Printf("Peringatan: gagal hash password: %v", err)
		return in
	}
	return hashed
}

func clientIP(r *http.Request) string {
	if trustProxy {
		if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
		if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
			return xrip
		}
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

func (g *loginRateGuard) allow(ip string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now()
	b := g.buckets[ip]
	if b == nil || now.Sub(b.windowStart) >= loginRateWindow {
		g.buckets[ip] = &apiRateBucket{count: 1, windowStart: now}
		return true
	}
	if b.count >= loginRateMax {
		return false
	}
	b.count++
	return true
}

func purgeLoginRateBucketsLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		now := time.Now()
		loginRateLimiter.mu.Lock()
		for k, b := range loginRateLimiter.buckets {
			if now.Sub(b.windowStart) > 2*loginRateWindow {
				delete(loginRateLimiter.buckets, k)
			}
		}
		loginRateLimiter.mu.Unlock()
	}
}

func loginRateAllow(ip string) bool {
	return loginRateLimiter.allow(ip)
}

func purgeLoginAttemptsLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	for range ticker.C {
		now := time.Now()
		loginLimiter.mu.Lock()
		for ip, a := range loginLimiter.attempts {
			if now.After(a.lockedUntil) && a.fails == 0 {
				delete(loginLimiter.attempts, ip)
			}
		}
		loginLimiter.mu.Unlock()
	}
}

func apiRateKey(r *http.Request) string {
	if tok := bearerToken(r); tok != "" {
		n := len(tok)
		if n > 16 {
			n = 16
		}
		return "t:" + tok[:n]
	}
	return "ip:" + clientIP(r)
}

func ipRateLimitForPath(path string) (max int, window time.Duration, skip bool) {
	switch {
	case path == "/health":
		return 0, 0, true
	case path == "/data/auth/login":
		return 0, 0, true
	case path == "/data/portals/status":
		return portalStatusRateMax, time.Minute, false
	case strings.HasPrefix(path, "/assets/"), path == "/favicon.ico":
		return assetRateLimitMax, time.Minute, false
	default:
		return ipRateLimitMax, time.Minute, false
	}
}

func suspiciousUserAgent(ua string) bool {
	ua = strings.ToLower(strings.TrimSpace(ua))
	if ua == "" {
		return false
	}
	bad := []string{
		"sqlmap", "nikto", "masscan", "nmap", "acunetix", "nessus", "dirbuster",
		"gobuster", "wpscan", "havij", "zgrab",
	}
	for _, b := range bad {
		if strings.Contains(ua, b) {
			return true
		}
	}
	return false
}

func withIPShield(next http.Handler) http.Handler {
	const maxPathLen = 2048
	const maxQueryLen = 4096
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > maxPathLen || len(r.URL.RawQuery) > maxQueryLen {
			http.Error(w, "URI Too Long", http.StatusRequestURITooLong)
			return
		}

		if strings.HasPrefix(path, "/data/") && suspiciousUserAgent(r.Header.Get("User-Agent")) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ip := clientIP(r)
		if path != "/health" {
			if !acquireGlobalConn() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "3")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error":"Server sedang penuh. Coba lagi sebentar."}`))
				return
			}
			defer releaseGlobalConn()
			if !shardedConnLimiter.acquire(ip, maxConnPerIP) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "5")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error":"Server sibuk. Coba lagi sebentar."}`))
				return
			}
			defer shardedConnLimiter.release(ip)
		}

		max, window, skip := ipRateLimitForPath(path)
		if !skip {
			key := "ip:" + ip + ":" + path
			if strings.HasPrefix(path, "/assets/") {
				key = "ip:" + ip + ":assets"
			} else if strings.HasPrefix(path, "/data/") {
				key = "ip:" + ip + ":data"
			}
			if !shardedIPRate.allow(key, max, window) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "30")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Terlalu banyak permintaan dari jaringan ini."}`))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func withAPIRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/health" || path == "/favicon.ico" || path == "/data/auth/login" {
			next.ServeHTTP(w, r)
			return
		}
		if path == "/data/portals/status" {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(path, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}
		if path == "/" && r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		if !shardedAPIRate.allow(apiRateKey(r), apiRateLimitMax, apiRateWindow) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Terlalu banyak permintaan. Coba lagi sebentar."}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC [%s %s]: %v", r.Method, r.URL.Path, rec)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Kesalahan internal server"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func passwordMatches(stored, input string) bool {
	stored = strings.TrimSpace(stored)
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		bcryptSem <- struct{}{}
		defer func() { <-bcryptSem }()
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(input)) == nil
	}
	if len(stored) != len(input) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(stored), []byte(input)) == 1
}

func withBlockSuspiciousPaths(next http.Handler) http.Handler {
	blocked := []string{
		"/.env", "/.git", "/wp-admin", "/wp-login", "/phpmyadmin", "/admin.php",
		"/cgi-bin", "/vendor/phpunit", "/.aws", "/config.php", "/shell",
		"/xmlrpc.php", "/.well-known/security.txt", "/actuator", "/server-status",
		"/telescope", "/debug", "/_profiler", "/solr", "/manager/html",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.ToLower(r.URL.Path)
		if strings.Contains(p, "..") || strings.Contains(p, "\x00") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		for _, b := range blocked {
			if strings.HasPrefix(p, b) || strings.Contains(p, b) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodPost &&
			r.Method != http.MethodPut && r.Method != http.MethodDelete && r.Method != http.MethodOptions {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
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
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
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

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

func withGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "/health" || strings.HasPrefix(path, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		gw := gzipWriterPool.Get().(*gzip.Writer)
		gw.Reset(w)
		defer func() {
			_ = gw.Close()
			gzipWriterPool.Put(gw)
		}()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, gw: gw}, r)
	})
}
