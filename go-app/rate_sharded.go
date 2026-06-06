package main

import (
	"hash/fnv"
	"sync"
	"time"
)

const rateShardCount = 64

type rateShard struct {
	mu      sync.Mutex
	buckets map[string]*apiRateBucket
}

type shardedRateGuard struct {
	shards [rateShardCount]rateShard
}

func newShardedRateGuard() *shardedRateGuard {
	g := &shardedRateGuard{}
	for i := range g.shards {
		g.shards[i].buckets = map[string]*apiRateBucket{}
	}
	return g
}

func rateShardIndex(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32() % rateShardCount
}

func (g *shardedRateGuard) allow(key string, max int, window time.Duration) bool {
	if max <= 0 {
		return true
	}
	sh := &g.shards[rateShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()
	now := time.Now()
	b := sh.buckets[key]
	if b == nil || now.Sub(b.windowStart) >= window {
		sh.buckets[key] = &apiRateBucket{count: 1, windowStart: now}
		return true
	}
	if b.count >= max {
		return false
	}
	b.count++
	return true
}

func (g *shardedRateGuard) purge(stale time.Duration) {
	now := time.Now()
	for i := range g.shards {
		sh := &g.shards[i]
		sh.mu.Lock()
		for k, b := range sh.buckets {
			if now.Sub(b.windowStart) > stale {
				delete(sh.buckets, k)
			}
		}
		sh.mu.Unlock()
	}
}

var (
	shardedAPIRate  = newShardedRateGuard()
	shardedIPRate   = newShardedRateGuard()
	globalConnGuard struct {
		mu    sync.Mutex
		count int
	}
	globalMaxConns int
)

func initGlobalConnLimit() {
	globalMaxConns = securityEnvInt("SIPKEU_GLOBAL_MAX_CONN", 22000)
	if globalMaxConns < 2000 {
		globalMaxConns = 2000
	}
}

func acquireGlobalConn() bool {
	globalConnGuard.mu.Lock()
	defer globalConnGuard.mu.Unlock()
	if globalConnGuard.count >= globalMaxConns {
		return false
	}
	globalConnGuard.count++
	return true
}

func releaseGlobalConn() {
	globalConnGuard.mu.Lock()
	if globalConnGuard.count > 0 {
		globalConnGuard.count--
	}
	globalConnGuard.mu.Unlock()
}

func purgeShardedRateGuardsLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		shardedAPIRate.purge(3 * apiRateWindow)
		shardedIPRate.purge(3 * time.Minute)
	}
}

type shardedConnGuard struct {
	shards [rateShardCount]struct {
		mu     sync.Mutex
		active map[string]int
	}
}

func newShardedConnGuard() *shardedConnGuard {
	g := &shardedConnGuard{}
	for i := range g.shards {
		g.shards[i].active = map[string]int{}
	}
	return g
}

func (g *shardedConnGuard) acquire(ip string, max int) bool {
	sh := &g.shards[rateShardIndex("conn:"+ip)]
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if sh.active[ip] >= max {
		return false
	}
	sh.active[ip]++
	return true
}

func (g *shardedConnGuard) release(ip string) {
	sh := &g.shards[rateShardIndex("conn:"+ip)]
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if sh.active[ip] > 0 {
		sh.active[ip]--
		if sh.active[ip] == 0 {
			delete(sh.active, ip)
		}
	}
}

var shardedConnLimiter = newShardedConnGuard()
