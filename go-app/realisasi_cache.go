package main

import (
	"fmt"
	"sync"
	"time"
)

const realisasiCacheTTL = 45 * time.Second

type realisasiCacheEntry struct {
	report  RealisasiReport
	expires time.Time
}

var (
	realisasiCacheMu sync.RWMutex
	realisasiCache   = map[string]realisasiCacheEntry{}
)

func realisasiCacheKey(moduleID string, f realisasiFilters) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%d|%d|%s|%s",
		moduleID, f.Tahun, f.Kegiatan, f.SubKegiatan, f.KodeRekening, f.PPTK, f.Bulan, f.Dari, f.Sampai, f.Search, f.Page, f.PageSize, f.Sort, f.SortDir)
}

func invalidateRealisasiCache(moduleID string) {
	realisasiCacheMu.Lock()
	if moduleID == "" {
		realisasiCache = map[string]realisasiCacheEntry{}
	} else {
		for k := range realisasiCache {
			if len(k) >= len(moduleID) && k[:len(moduleID)] == moduleID {
				delete(realisasiCache, k)
			}
		}
	}
	realisasiCacheMu.Unlock()
}

func cachedRealisasiReport(moduleID string, f realisasiFilters, build func() RealisasiReport) RealisasiReport {
	key := realisasiCacheKey(moduleID, f)
	now := time.Now()
	realisasiCacheMu.RLock()
	if c, ok := realisasiCache[key]; ok && now.Before(c.expires) {
		report := c.report
		realisasiCacheMu.RUnlock()
		return report
	}
	realisasiCacheMu.RUnlock()

	report := build()
	realisasiCacheMu.Lock()
	realisasiCache[key] = realisasiCacheEntry{report: report, expires: now.Add(realisasiCacheTTL)}
	realisasiCacheMu.Unlock()
	return report
}
