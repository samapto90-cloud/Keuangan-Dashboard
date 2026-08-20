package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"keuangan/mmo"

	"github.com/gorilla/websocket"
)

func marshalEnvelope(typ string, data any) []byte {
	raw, _ := json.Marshal(data)
	env := mmo.Envelope{Type: typ, Data: raw}
	out, _ := json.Marshal(env)
	return out
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func main() {
	var (
		wsURL       = flag.String("ws", "", "WebSocket URL (e.g. ws://127.0.0.1:3000/cahaya/ws)")
		token       = flag.String("token", "", "Auth token for WS (required)")
		clients     = flag.Int("clients", 10, "Number of concurrent WS clients")
		durationSec = flag.Int("duration", 20, "Test duration seconds")
	)
	flag.Parse()

	if *wsURL == "" {
		*wsURL = os.Getenv("ULAR_LOADTEST_WS")
	}
	if *token == "" {
		*token = os.Getenv("ULAR_LOADTEST_TOKEN")
	}

	if *wsURL == "" || *token == "" {
		log.Fatalf("missing ws/token: use -ws/-token or env ULAR_LOADTEST_WS/ULAR_LOADTEST_TOKEN")
	}

	dur := time.Duration(*durationSec) * time.Second
	endAt := time.Now().Add(dur)

	var (
		muLat     sync.Mutex
		lats      []float64 // ms
		muErr     sync.Mutex
		errCount  int
		reqCount  int
	)

	var wg sync.WaitGroup
	wg.Add(*clients)
	for i := 0; i < *clients; i++ {
		go func(clientID int) {
			defer wg.Done()
			c, _, err := websocket.DefaultDialer.Dial(*wsURL, nil)
			if err != nil {
				muErr.Lock()
				errCount++
				muErr.Unlock()
				return
			}
			defer c.Close()

			_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))

			// Auth
			if err := c.WriteMessage(websocket.TextMessage, marshalEnvelope(mmo.TypeAuth, mmo.AuthIn{Token: *token})); err != nil {
				muErr.Lock()
				errCount++
				muErr.Unlock()
				return
			}

			// Wait for TypeAuthOk
			joined := false
			for time.Now().Before(endAt) && !joined {
				_, raw, err := c.ReadMessage()
				if err != nil {
					muErr.Lock()
					errCount++
					muErr.Unlock()
					return
				}
				var env mmo.Envelope
				if err := json.Unmarshal(raw, &env); err != nil {
					continue
				}
				if env.Type == mmo.TypeAuthOk {
					// Join Ular lobby (server will route Phase1 Ular events).
					if err := c.WriteMessage(websocket.TextMessage, marshalEnvelope(mmo.TypeJoinLobby, struct{}{})); err != nil {
						muErr.Lock()
						errCount++
						muErr.Unlock()
						return
					}
					joined = true
				}
			}
			if !joined {
				muErr.Lock()
				errCount++
				muErr.Unlock()
				return
			}

			// Request loop: measure ROOM_LIST RTT.
			for time.Now().Before(endAt) {
				t0 := time.Now()
				if err := c.WriteMessage(websocket.TextMessage, marshalEnvelope(mmo.TypeRoomList, struct{}{})); err != nil {
					muErr.Lock()
					errCount++
					muErr.Unlock()
					return
				}

				// Wait until we see ROOM_LIST response.
				for {
					_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
					_, raw, err := c.ReadMessage()
					if err != nil {
						muErr.Lock()
						errCount++
						muErr.Unlock()
						return
					}
					var env mmo.Envelope
					if err := json.Unmarshal(raw, &env); err != nil {
						continue
					}
					if env.Type == mmo.TypeRoomList {
						ms := float64(time.Since(t0).Microseconds()) / 1000.0
						muLat.Lock()
						lats = append(lats, ms)
						muLat.Unlock()
						muErr.Lock()
						reqCount++
						muErr.Unlock()
						break
					}
				}
			}
		}(i + 1)
	}

	wg.Wait()

	muLat.Lock()
	sort.Float64s(lats)
	latSorted := append([]float64(nil), lats...)
	muLat.Unlock()

	muErr.Lock()
	localErrCount := errCount
	localReqCount := reqCount
	muErr.Unlock()

	p50 := percentile(latSorted, 0.50)
	p95 := percentile(latSorted, 0.95)

	fmt.Printf("ULAR loadtest (ROOM_LIST): clients=%d req=%d err=%d p50=%.1fms p95=%.1fms max=%.1fms\n",
		*clients, localReqCount, localErrCount, p50, p95, func() float64 {
			if len(latSorted) == 0 {
				return math.NaN()
			}
			return latSorted[len(latSorted)-1]
		}(),
	)
}

