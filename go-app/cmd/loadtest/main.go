package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	base := env("LOADTEST_URL", "http://127.0.0.1:3000")
	users := envInt("LOADTEST_USERS", 200)
	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err == nil && n > 0 {
			users = n
		}
	}
	fmt.Printf("Load test: %d virtual users -> %s\n", users, base)

	var ok, fail uint64
	start := time.Now()
	var wg sync.WaitGroup
	sem := make(chan struct{}, 50)

	for i := 0; i < users; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			client := &http.Client{Timeout: 15 * time.Second}
			if code := get(client, base+"/health"); code == 200 {
				atomic.AddUint64(&ok, 1)
			} else {
				atomic.AddUint64(&fail, 1)
			}

			loginBody, _ := json.Marshal(map[string]string{
				"username": "admin",
				"password": "admin2026",
			})
			req, _ := http.NewRequest(http.MethodPost, base+"/data/auth/login", bytes.NewReader(loginBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-SIPKEU-App", "sekretariat")
			res, err := client.Do(req)
			if err != nil || res.StatusCode != 200 {
				atomic.AddUint64(&fail, 1)
				return
			}
			var loginResp map[string]interface{}
			json.NewDecoder(res.Body).Decode(&loginResp)
			res.Body.Close()
			token, _ := loginResp["token"].(string)
			if token == "" {
				atomic.AddUint64(&fail, 1)
				return
			}

			h := http.Header{}
			h.Set("Authorization", "Bearer "+token)
			h.Set("X-SIPKEU-App", "sekretariat")
			dreq, _ := http.NewRequest(http.MethodGet, base+"/data/dashboard", nil)
			dreq.Header = h
			dres, err := client.Do(dreq)
			if err != nil || dres.StatusCode != 200 {
				atomic.AddUint64(&fail, 1)
				return
			}
			io.Copy(io.Discard, dres.Body)
			dres.Body.Close()
			atomic.AddUint64(&ok, 1)
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	total := ok + fail
	fmt.Printf("Done in %v — OK: %d, Fail: %d, RPS: %.1f\n", elapsed, ok, fail, float64(total)/elapsed.Seconds())
	if fail > total/10 {
		os.Exit(1)
	}
}

func get(client *http.Client, url string) int {
	res, err := client.Get(url)
	if err != nil {
		return 0
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)
	return res.StatusCode
}

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
