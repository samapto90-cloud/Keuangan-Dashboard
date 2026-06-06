package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type portalOverview struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Enabled      bool   `json:"enabled"`
	TxCount      int    `json:"tx_count"`
	ActiveUsers  int    `json:"active_sessions"`
	AdminUser    string `json:"admin_user"`
	HasOperator  bool   `json:"has_operator"`
	OperatorUser string `json:"operator_user"`
}

func portalLabel(id string) string {
	labels := map[string]string{
		"sekretariat": "SIPKEU Sekretariat",
		"paud":        "SIPKEU PAUD",
		"sd":          "SIPKEU SD",
		"smp":         "SIPKEU SMP",
		"kas-belanja": "Realisasi Kas Belanja",
		"pengaturan":  "Command Center",
	}
	if l, ok := labels[id]; ok {
		return l
	}
	return id
}

func isPortalEnabled(id string) bool {
	if id == "pengaturan" {
		return true
	}
	sys := getSystemSettingsCopy()
	if sys.PortalStatus == nil {
		return true
	}
	st, ok := sys.PortalStatus[id]
	if !ok {
		return true
	}
	return st.Enabled
}

func activeSessionsByModule() map[string]int {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	now := time.Now()
	counts := map[string]int{}
	for _, s := range sessions {
		if now.Before(s.Expires) {
			counts[s.AppModule]++
		}
	}
	return counts
}

func buildCommandCenterOverview() map[string]interface{} {
	sys := getSystemSettingsCopy()
	sessCounts := activeSessionsByModule()
	portals := []portalOverview{}

	for _, id := range sipkeuPortalIDs {
		cfg := sys.Portals[id]
		txCount := 0
		if mod := sipkeuModules[id]; mod != nil {
			mod.mu.Lock()
			txCount = len(mod.txs)
			mod.mu.Unlock()
		}
		enabled := true
		if sys.PortalStatus != nil {
			if st, ok := sys.PortalStatus[id]; ok {
				enabled = st.Enabled
			}
		}
		portals = append(portals, portalOverview{
			ID:           id,
			Label:        portalLabel(id),
			Enabled:      enabled,
			TxCount:      txCount,
			ActiveUsers:  sessCounts[id],
			AdminUser:    cfg.AdminUsername,
			HasOperator:  strings.TrimSpace(cfg.OperatorUsername) != "",
			OperatorUser: cfg.OperatorUsername,
		})
	}

	totalSess := 0
	for _, c := range sessCounts {
		totalSess += c
	}

	return map[string]interface{}{
		"generated_at":     time.Now().Format(time.RFC3339),
		"active_sessions":  totalSess,
		"portals":          portals,
		"security": map[string]interface{}{
			"rate_limit_per_min": apiRateLimitMax,
			"login_max_fails":    loginMaxFails,
			"session_hours":      sessionLifetime().Hours(),
			"trust_proxy":        trustProxy,
		},
		"recent_audit": auditLogCopy(15),
	}
}

func sessionListPublic() []map[string]interface{} {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	now := time.Now()
	out := []map[string]interface{}{}
	for token, s := range sessions {
		if now.After(s.Expires) {
			continue
		}
		masked := token
		if len(masked) > 8 {
			masked = masked[:8] + "…"
		}
		out = append(out, map[string]interface{}{
			"token_id":   masked,
			"token":      token,
			"username":   s.Username,
			"role":       s.Role,
			"name":       s.Name,
			"app_module": s.AppModule,
			"client_ip":  s.ClientIP,
			"created_at": s.CreatedAt.Format(time.RFC3339),
			"last_seen":  s.LastSeen.Format(time.RFC3339),
			"expires_at": s.Expires.Format(time.RFC3339),
		})
	}
	return out
}

func revokeSessionToken(token string) bool {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if _, ok := sessions[token]; ok {
		delete(sessions, token)
		return true
	}
	return false
}

func revokeSessionsForModule(moduleID string) int {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	n := 0
	for tok, s := range sessions {
		if s.AppModule == moduleID {
			delete(sessions, tok)
			n++
		}
	}
	return n
}

func handleAdminCommandCenter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	jsonResponse(w, http.StatusOK, buildCommandCenterOverview())
}

func handleAdminSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"sessions": sessionListPublic(),
			"total":    len(sessionListPublic()),
		})
	case http.MethodPost:
		var body struct {
			Token    string `json:"token"`
			ModuleID string `json:"module_id"`
			All      bool   `json:"all"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		sess := getSession(r)
		revoked := 0
		if body.All {
			revoked = revokeAllSessionsExcept(bearerToken(r))
			recordAudit(sess.Username, "revoke_all_sessions", "security", "Semua sesi diakhiri kecuali sesi admin", clientIP(r))
		} else if body.ModuleID != "" {
			revoked = revokeSessionsForModule(strings.TrimSpace(body.ModuleID))
			recordAudit(sess.Username, "revoke_module_sessions", body.ModuleID, "Sesi portal diakhiri", clientIP(r))
		} else if body.Token != "" {
			if revokeSessionToken(strings.TrimSpace(body.Token)) {
				revoked = 1
			}
			recordAudit(sess.Username, "revoke_session", "security", "Sesi tunggal diakhiri", clientIP(r))
		} else {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "token, module_id, atau all wajib diisi"})
			return
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"revoked": revoked,
			"message": "Sesi berhasil diakhiri",
		})
	default:
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func revokeAllSessionsExcept(keepToken string) int {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	n := 0
	for tok := range sessions {
		if tok == keepToken {
			continue
		}
		delete(sessions, tok)
		n++
	}
	return n
}

func handleAdminAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	limit := 100
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"entries": auditLogCopy(limit),
	})
}
