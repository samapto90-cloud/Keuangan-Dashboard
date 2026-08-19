package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	base := strings.TrimRight(env("SMOKE_URL", "http://127.0.0.1:38991"), "/")
	client := &http.Client{Timeout: 8 * time.Second}

	fail := 0
	check := func(name, path, want string) {
		res, err := client.Get(base + path)
		if err != nil {
			fmt.Printf("FAIL %s: %v\n", name, err)
			fail++
			return
		}
		defer res.Body.Close()
		raw, _ := io.ReadAll(res.Body)
		body := string(raw)
		if res.StatusCode != 200 || (want != "" && !strings.Contains(body, want)) {
			fmt.Printf("FAIL %s: status=%d body=%s\n", name, res.StatusCode, truncate(body, 240))
			fail++
			return
		}
		fmt.Printf("PASS %s\n", name)
	}

	check("health", "/health", `"status":"ok"`)
	check("version", "/health", `"version":"1.0.0-beta"`)
	check("phase", "/health", `"phase":"30/30"`)
	check("ready", "/ready", `"status":"ready"`)
	check("cahaya", "/cahaya/", "Petualangan Menuju Cahaya")
	check("kicker", "/cahaya/", "1.0.0-beta")

	res, err := client.Get(base + "/health")
	if err == nil {
		defer res.Body.Close()
		var health map[string]string
		_ = json.NewDecoder(res.Body).Decode(&health)
		fmt.Printf("health=%v\n", health)
	}

	if fail > 0 {
		os.Exit(1)
	}
	fmt.Println("SMOKE PASS")
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
