package main

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const adminRekapCacheTTL = 45 * time.Second

type adminRekapCacheEntry struct {
	rows    []adminRekapRow
	expires time.Time
}

var (
	adminRekapCacheMu sync.RWMutex
	adminRekapCache   = map[string]adminRekapCacheEntry{}
)

func adminRekapCacheKey(portals []string, mode, from, to, jenis string) string {
	ps := append([]string(nil), portals...)
	sort.Strings(ps)
	return strings.Join(ps, ",") + "\x00" + mode + "\x00" + from + "\x00" + to + "\x00" + jenis
}

func invalidateAdminRekapCache() {
	adminRekapCacheMu.Lock()
	adminRekapCache = map[string]adminRekapCacheEntry{}
	adminRekapCacheMu.Unlock()
}

func cachedAdminRekapRows(portals []string, mode, from, to, jenis string, build func() []adminRekapRow) []adminRekapRow {
	key := adminRekapCacheKey(portals, mode, from, to, jenis)
	now := time.Now()
	adminRekapCacheMu.RLock()
	if c, ok := adminRekapCache[key]; ok && now.Before(c.expires) {
		rows := append([]adminRekapRow(nil), c.rows...)
		adminRekapCacheMu.RUnlock()
		return rows
	}
	adminRekapCacheMu.RUnlock()

	rows := build()
	adminRekapCacheMu.Lock()
	adminRekapCache[key] = adminRekapCacheEntry{
		rows:    append([]adminRekapRow(nil), rows...),
		expires: now.Add(adminRekapCacheTTL),
	}
	adminRekapCacheMu.Unlock()
	return rows
}

func adminRekapRowsToPPTKStats(rows []adminRekapRow) []adminRekapPPTKStat {
	stats := make([]adminRekapPPTKStat, 0, len(rows))
	for _, r := range rows {
		stats = append(stats, adminRekapPPTKStat{
			PortalID:    r.PortalID,
			PortalLabel: r.PortalLabel,
			PPTK:        r.PPTK,
			Anggaran:    r.Anggaran,
			Realisasi:   r.Realisasi,
			Sisa:        r.Sisa,
			Count:       r.Count,
			Pct:         r.Pct,
		})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Realisasi != stats[j].Realisasi {
			return stats[i].Realisasi > stats[j].Realisasi
		}
		return stats[i].Anggaran > stats[j].Anggaran
	})
	return stats
}
