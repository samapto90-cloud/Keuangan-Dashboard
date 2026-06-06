package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type stepResult struct {
	ok      bool
	latency time.Duration
	code    int
	step    string
}

func main() {
	base := env("LOADTEST_URL", "http://127.0.0.1:3099")
	desktop := envInt("LOADTEST_DESKTOP", 5000)
	mobile := envInt("LOADTEST_MOBILE", 5000)
	concurrency := envInt("LOADTEST_CONCURRENCY", 400)
	user := env("LOADTEST_USER", "admin")
	pass := env("LOADTEST_PASS", "admin2026")
	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err == nil && n > 0 {
			desktop = n
			mobile = 0
		}
	}
	total := desktop + mobile
	fmt.Printf("Load test SIPKEU\n  URL: %s\n  Desktop: %d | Mobile: %d | Total: %d\n  Concurrency: %d\n",
		base, desktop, mobile, total, concurrency)

	transport := &http.Transport{
		MaxIdleConns:          concurrency * 6,
		MaxIdleConnsPerHost:   concurrency * 6,
		MaxConnsPerHost:       concurrency * 3,
		IdleConnTimeout:       120 * time.Second,
		ResponseHeaderTimeout: 45 * time.Second,
		ForceAttemptHTTP2:     false,
		DisableCompression:    false,
	}
	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	var okCount, failCount uint64
	results := make([]stepResult, 0, total*5)
	var resMu sync.Mutex
	start := time.Now()
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	runUser := func(id int, mobile bool) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()

		ua := desktopUA
		if mobile {
			ua = mobileUA
		}
		vip := virtualIP(id)

		record := func(step string, ok bool, code int, lat time.Duration) {
			resMu.Lock()
			results = append(results, stepResult{ok: ok, latency: lat, code: code, step: step})
			resMu.Unlock()
			if ok {
				atomic.AddUint64(&okCount, 1)
			} else {
				atomic.AddUint64(&failCount, 1)
			}
		}

		do := func(step, method, path string, body []byte, auth string) bool {
			for attempt := 0; attempt < 3; attempt++ {
				if attempt > 0 {
					time.Sleep(time.Duration(attempt*40) * time.Millisecond)
				}
				t0 := time.Now()
				var bodyReader io.Reader
				if body != nil {
					bodyReader = bytes.NewReader(body)
				}
				req, err := http.NewRequest(method, base+path, bodyReader)
				if err != nil {
					continue
				}
				req.Header.Set("User-Agent", ua)
				req.Header.Set("Accept-Encoding", "gzip")
				req.Header.Set("X-Forwarded-For", vip)
				req.Header.Set("X-Real-IP", vip)
				if body != nil {
					req.Header.Set("Content-Type", "application/json")
				}
				if auth != "" {
					req.Header.Set("Authorization", "Bearer "+auth)
				}
				req.Header.Set("X-SIPKEU-App", "sekretariat")
				res, err := client.Do(req)
				lat := time.Since(t0)
				if err != nil {
					if attempt == 2 {
						record(step, false, 0, lat)
					}
					continue
				}
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
				ok := res.StatusCode >= 200 && res.StatusCode < 300
				if ok {
					record(step, true, res.StatusCode, lat)
					return res.StatusCode == 200
				}
				if attempt == 2 {
					record(step, false, res.StatusCode, lat)
				}
			}
			return false
		}

		if !do("index", http.MethodGet, "/", nil, "") {
			return
		}
		if !do("portal_status", http.MethodGet, "/data/portals/status", nil, "") {
			return
		}
		loginBody, _ := json.Marshal(map[string]string{"username": user, "password": pass})
		t0 := time.Now()
		req, err := http.NewRequest(http.MethodPost, base+"/data/auth/login", bytes.NewReader(loginBody))
		if err != nil {
			record("login", false, 0, time.Since(t0))
			return
		}
		req.Header.Set("User-Agent", ua)
		req.Header.Set("X-Forwarded-For", vip)
		req.Header.Set("X-Real-IP", vip)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-SIPKEU-App", "sekretariat")
		res, err := client.Do(req)
		lat := time.Since(t0)
		if err != nil {
			record("login", false, 0, lat)
			return
		}
		var loginResp map[string]interface{}
		json.NewDecoder(res.Body).Decode(&loginResp)
		res.Body.Close()
		token, _ := loginResp["token"].(string)
		okLogin := res.StatusCode == 200 && token != ""
		record("login", okLogin, res.StatusCode, lat)
		if !okLogin {
			return
		}

		if !do("dashboard", http.MethodGet, "/data/dashboard", nil, token) {
			return
		}
		do("transactions", http.MethodGet, "/data/transactions", nil, token)
		do("favicon", http.MethodGet, "/favicon.ico", nil, "")
	}

	for i := 0; i < desktop; i++ {
		wg.Add(1)
		go runUser(i, false)
	}
	for i := 0; i < mobile; i++ {
		wg.Add(1)
		go runUser(10000+i, true)
	}
	wg.Wait()
	elapsed := time.Since(start)

	printReport(results, elapsed, okCount, failCount)
	if failCount > 0 && float64(failCount)/float64(okCount+failCount) > 0.05 {
		os.Exit(1)
	}
}

func printReport(results []stepResult, elapsed time.Duration, ok, fail uint64) {
	fmt.Printf("\nDone in %v — OK steps: %d, Fail steps: %d, throughput: %.1f req/s\n",
		elapsed, ok, fail, float64(ok+fail)/elapsed.Seconds())

	byStep := map[string][]time.Duration{}
	failByStep := map[string]int{}
	failCodes := map[string]map[int]int{}
	for _, r := range results {
		if r.ok {
			byStep[r.step] = append(byStep[r.step], r.latency)
		} else {
			failByStep[r.step]++
			if failCodes[r.step] == nil {
				failCodes[r.step] = map[int]int{}
			}
			failCodes[r.step][r.code]++
		}
	}
	steps := []string{"index", "portal_status", "login", "dashboard", "transactions", "favicon"}
	for _, step := range steps {
		lats := byStep[step]
		if len(lats) == 0 {
			if failByStep[step] > 0 {
				fmt.Printf("  %-16s FAIL=%d codes=%v\n", step+":", failByStep[step], failCodes[step])
			}
			continue
		}
		sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })
		fmt.Printf("  %-16s n=%-5d p50=%-8v p95=%-8v max=%-8v fail=%d codes=%v\n",
			step+":", len(lats), lats[pct(lats, 50)], lats[pct(lats, 95)], lats[len(lats)-1], failByStep[step], failCodes[step])
	}
}

func pct(lats []time.Duration, p int) int {
	if len(lats) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(p)/100*float64(len(lats)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(lats) {
		idx = len(lats) - 1
	}
	return idx
}

func virtualIP(id int) string {
	a := 10 + (id / 65025)
	b := (id / 255) % 255
	c := id % 255
	if a > 250 {
		a = 10 + a%240
	}
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, 1+(id%200))
}

const desktopUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
const mobileUA = "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36"

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
