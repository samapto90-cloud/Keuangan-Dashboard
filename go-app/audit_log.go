package main

import (
	"sync"
	"time"
)

const auditLogMax = 500

type AuditEntry struct {
	At       time.Time `json:"at"`
	Actor    string    `json:"actor"`
	Action   string    `json:"action"`
	Module   string    `json:"module"`
	Detail   string    `json:"detail"`
	ClientIP string    `json:"client_ip"`
}

var (
	auditMu   sync.RWMutex
	auditRing []AuditEntry
)

func recordAudit(actor, action, module, detail, ip string) {
	e := AuditEntry{
		At:       time.Now(),
		Actor:    actor,
		Action:   action,
		Module:   module,
		Detail:   detail,
		ClientIP: ip,
	}
	auditMu.Lock()
	auditRing = append(auditRing, e)
	if len(auditRing) > auditLogMax {
		auditRing = append([]AuditEntry(nil), auditRing[len(auditRing)-auditLogMax:]...)
	}
	auditMu.Unlock()
}

func auditLogCopy(limit int) []AuditEntry {
	if limit <= 0 || limit > auditLogMax {
		limit = auditLogMax
	}
	auditMu.RLock()
	defer auditMu.RUnlock()
	n := len(auditRing)
	if n == 0 {
		return []AuditEntry{}
	}
	start := 0
	if n > limit {
		start = n - limit
	}
	out := make([]AuditEntry, n-start)
	copy(out, auditRing[start:])
	// newest first
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
