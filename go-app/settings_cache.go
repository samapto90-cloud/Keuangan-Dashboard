package main

import (
	"sync"
	"time"
)

const settingsCacheTTL = 45 * time.Second

type settingsCacheEntry struct {
	data    AppSettings
	expires time.Time
}

var (
	settingsCacheMu sync.RWMutex
	settingsCache   = map[string]settingsCacheEntry{}
)

func invalidateSettingsCache(moduleID string) {
	settingsCacheMu.Lock()
	delete(settingsCache, moduleID)
	settingsCacheMu.Unlock()
}

func cachedModuleSettings(mod *SipkeuModule) AppSettings {
	if mod == nil {
		return AppSettings{}
	}
	moduleID := mod.ID
	now := time.Now()

	settingsCacheMu.RLock()
	if c, ok := settingsCache[moduleID]; ok && now.Before(c.expires) {
		out := c.data
		settingsCacheMu.RUnlock()
		return out
	}
	settingsCacheMu.RUnlock()

	mod.mu.Lock()
	copyAnggaran := cloneAnggaranMap(mod.settings.AnggaranKegiatan)
	rakCopy := cloneRakRows(mod.settings.Rak)
	pa, bend := effectivePejabatValues(mod.ID, mod.settings.PA, mod.settings.Bendahara, mod.defaultSettings.PA, mod.defaultSettings.Bendahara)
	out := AppSettings{
		PA:               pa,
		Bendahara:        bend,
		AnggaranKegiatan: copyAnggaran,
		Rak:              rakCopy,
		RakMeta:          mod.settings.RakMeta,
	}
	mod.mu.Unlock()

	settingsCacheMu.Lock()
	settingsCache[moduleID] = settingsCacheEntry{data: out, expires: now.Add(settingsCacheTTL)}
	settingsCacheMu.Unlock()
	return out
}
