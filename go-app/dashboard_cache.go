package main

import (
	"sync"
	"time"
)

const dashboardCacheTTL = 30 * time.Second

type dashboardCacheEntry struct {
	stats   DashboardStats
	expires time.Time
}

var (
	dashboardCacheMu sync.RWMutex
	dashboardCache   = map[string]dashboardCacheEntry{}
)

func invalidateDashboardCache(moduleID string) {
	dashboardCacheMu.Lock()
	delete(dashboardCache, moduleID)
	dashboardCacheMu.Unlock()
}

func cachedDashboardStats(moduleID string, build func() DashboardStats) DashboardStats {
	now := time.Now()
	dashboardCacheMu.RLock()
	if c, ok := dashboardCache[moduleID]; ok && now.Before(c.expires) {
		stats := c.stats
		dashboardCacheMu.RUnlock()
		return stats
	}
	dashboardCacheMu.RUnlock()

	stats := build()
	dashboardCacheMu.Lock()
	dashboardCache[moduleID] = dashboardCacheEntry{stats: stats, expires: now.Add(dashboardCacheTTL)}
	dashboardCacheMu.Unlock()
	return stats
}
