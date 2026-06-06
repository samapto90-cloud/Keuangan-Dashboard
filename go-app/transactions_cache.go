package main

import (
	"sync"
	"time"
)

const transactionsCacheTTL = 25 * time.Second

type transactionsCacheEntry struct {
	data    []Transaction
	expires time.Time
}

var (
	transactionsCacheMu sync.RWMutex
	transactionsCache   = map[string]transactionsCacheEntry{}
)

func invalidateTransactionsCache(moduleID string) {
	transactionsCacheMu.Lock()
	delete(transactionsCache, moduleID)
	transactionsCacheMu.Unlock()
}

func cachedModuleTransactions(mod *SipkeuModule) []Transaction {
	if mod == nil {
		return nil
	}
	moduleID := mod.ID
	now := time.Now()

	transactionsCacheMu.RLock()
	if c, ok := transactionsCache[moduleID]; ok && now.Before(c.expires) {
		out := append([]Transaction(nil), c.data...)
		transactionsCacheMu.RUnlock()
		return out
	}
	transactionsCacheMu.RUnlock()

	data := moduleTransactionsCopy(mod)
	transactionsCacheMu.Lock()
	transactionsCache[moduleID] = transactionsCacheEntry{data: data, expires: now.Add(transactionsCacheTTL)}
	transactionsCacheMu.Unlock()
	return append([]Transaction(nil), data...)
}
