package main

import (
	"sync"
	"time"
)

const portalStatusCacheTTL = 45 * time.Second

var (
	portalStatusCacheMu sync.RWMutex
	portalStatusCached  map[string]interface{}
	portalStatusExpires time.Time
)

func cachedPortalStatus(build func() map[string]interface{}) map[string]interface{} {
	now := time.Now()
	portalStatusCacheMu.RLock()
	if portalStatusCached != nil && now.Before(portalStatusExpires) {
		out := portalStatusCached
		portalStatusCacheMu.RUnlock()
		return out
	}
	portalStatusCacheMu.RUnlock()

	out := build()
	portalStatusCacheMu.Lock()
	portalStatusCached = out
	portalStatusExpires = now.Add(portalStatusCacheTTL)
	portalStatusCacheMu.Unlock()
	return out
}

func invalidatePortalStatusCache() {
	portalStatusCacheMu.Lock()
	portalStatusCached = nil
	portalStatusExpires = time.Time{}
	portalStatusCacheMu.Unlock()
}
