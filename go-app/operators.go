package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

type PortalOperator struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
}

func newOperatorID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte("fallback01"))
	}
	return hex.EncodeToString(b)
}

func normalizeOperatorUsername(u string) string {
	return strings.ToLower(strings.TrimSpace(u))
}

func migratePortalOperators(cfg *PortalAuthConfig) bool {
	if cfg == nil {
		return false
	}
	changed := false
	if cfg.Operators == nil {
		cfg.Operators = []PortalOperator{}
		changed = true
	}
	if len(cfg.Operators) == 0 && strings.TrimSpace(cfg.OperatorUsername) != "" {
		cfg.Operators = append(cfg.Operators, PortalOperator{
			ID:       newOperatorID(),
			Username: strings.TrimSpace(cfg.OperatorUsername),
			Password: cfg.OperatorPassword,
			Name:     firstNonEmpty(cfg.OperatorName, "Operator SIPKEU"),
			Enabled:  true,
		})
		changed = true
	}
	for i := range cfg.Operators {
		if strings.TrimSpace(cfg.Operators[i].ID) == "" {
			cfg.Operators[i].ID = newOperatorID()
			changed = true
		}
		cfg.Operators[i].Username = strings.TrimSpace(cfg.Operators[i].Username)
		cfg.Operators[i].Name = strings.TrimSpace(cfg.Operators[i].Name)
	}
	syncLegacyOperatorFields(cfg)
	return changed
}

func syncLegacyOperatorFields(cfg *PortalAuthConfig) {
	if cfg == nil {
		return
	}
	for _, op := range cfg.Operators {
		if !op.Enabled || strings.TrimSpace(op.Username) == "" {
			continue
		}
		cfg.OperatorUsername = op.Username
		cfg.OperatorPassword = op.Password
		cfg.OperatorName = firstNonEmpty(op.Name, "Operator")
		return
	}
	if len(cfg.Operators) == 0 {
		return
	}
	// Keep last known credentials blank if all disabled.
	cfg.OperatorUsername = ""
	cfg.OperatorPassword = ""
	cfg.OperatorName = ""
}

func findOperatorByUsername(cfg PortalAuthConfig, username string) (PortalOperator, bool) {
	u := normalizeOperatorUsername(username)
	if u == "" {
		return PortalOperator{}, false
	}
	for _, op := range cfg.Operators {
		if !op.Enabled {
			continue
		}
		if normalizeOperatorUsername(op.Username) == u {
			return op, true
		}
	}
	// Legacy fallback
	if normalizeOperatorUsername(cfg.OperatorUsername) == u && strings.TrimSpace(cfg.OperatorUsername) != "" {
		return PortalOperator{
			ID:       "legacy",
			Username: cfg.OperatorUsername,
			Password: cfg.OperatorPassword,
			Name:     firstNonEmpty(cfg.OperatorName, "Operator"),
			Enabled:  true,
		}, true
	}
	return PortalOperator{}, false
}

func operatorUsernameTaken(cfg PortalAuthConfig, username, exceptID string) bool {
	u := normalizeOperatorUsername(username)
	if u == "" {
		return false
	}
	if normalizeOperatorUsername(cfg.AdminUsername) == u {
		return true
	}
	for _, op := range cfg.Operators {
		if exceptID != "" && op.ID == exceptID {
			continue
		}
		if normalizeOperatorUsername(op.Username) == u {
			return true
		}
	}
	return false
}

func maskOperators(ops []PortalOperator) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(ops))
	for _, op := range ops {
		item := map[string]interface{}{
			"id":       op.ID,
			"username": op.Username,
			"name":     op.Name,
			"enabled":  op.Enabled,
		}
		if strings.TrimSpace(op.Password) != "" {
			item["password"] = passwordMask
			item["has_password"] = true
		} else {
			item["password"] = ""
			item["has_password"] = false
		}
		out = append(out, item)
	}
	return out
}

func requirePortalAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := getSession(r)
		if sess == nil {
			jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
			return
		}
		if sess.Role != "admin" {
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Hanya admin portal yang dapat mengelola operator"})
			return
		}
		if isAdminOnlyPortal(sess.AppModule) {
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Portal ini tidak mendukung akun operator"})
			return
		}
		if !containsPortalID(sess.AppModule) {
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Portal tidak valid"})
			return
		}
		next(w, r)
	}
}

type operatorWritePayload struct {
	Username string                  `json:"username"`
	Password string                  `json:"password"`
	Name     string                  `json:"name"`
	Enabled  *bool                   `json:"enabled"`
	Perms    *OperatorPermissionSet  `json:"perms"`
}

func handleOperators(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if sess == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
		return
	}
	moduleID := sess.AppModule

	switch r.Method {
	case http.MethodGet:
		systemSettingsMu.RLock()
		cfg := systemSettings.Portals[moduleID]
		perms := systemSettings.OperatorPerms[moduleID]
		systemSettingsMu.RUnlock()
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"portal":    moduleID,
			"operators": maskOperators(cfg.Operators),
			"perms":     perms,
		})

	case http.MethodPost:
		var body operatorWritePayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		username := strings.TrimSpace(body.Username)
		name := firstNonEmpty(strings.TrimSpace(body.Name), "Operator")
		password := strings.TrimSpace(body.Password)
		if username == "" {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Username operator wajib diisi"})
			return
		}
		if password == "" {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Password operator wajib diisi"})
			return
		}
		if len(password) < 6 {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Password minimal 6 karakter"})
			return
		}

		systemSettingsMu.Lock()
		cfg := systemSettings.Portals[moduleID]
		migratePortalOperators(&cfg)
		if operatorUsernameTaken(cfg, username, "") {
			systemSettingsMu.Unlock()
			jsonResponse(w, http.StatusConflict, map[string]string{"error": "Username sudah dipakai di portal ini"})
			return
		}
		hashed, err := hashPasswordStore(password)
		if err != nil || hashed == "" {
			systemSettingsMu.Unlock()
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal mengamankan password"})
			return
		}
		op := PortalOperator{
			ID:       newOperatorID(),
			Username: username,
			Password: hashed,
			Name:     name,
			Enabled:  true,
		}
		if body.Enabled != nil {
			op.Enabled = *body.Enabled
		}
		cfg.Operators = append(cfg.Operators, op)
		syncLegacyOperatorFields(&cfg)
		systemSettings.Portals[moduleID] = cfg
		if body.Perms != nil {
			systemSettings.OperatorPerms[moduleID] = *body.Perms
		}
		systemSettingsMu.Unlock()
		persistSystemSettings()
		recordAudit(sess.Username, "operator_create", moduleID, "Operator dibuat: "+username, clientIP(r))
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"message":   "Operator berhasil dibuat",
			"operators": maskOperators(cfg.Operators),
			"perms":     operatorPermsForModule(moduleID),
		})

	default:
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func handleOperatorByID(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if sess == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
		return
	}
	moduleID := sess.AppModule
	id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/data/operators/"))
	if id == "" || strings.Contains(id, "/") {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "ID operator tidak valid"})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var body operatorWritePayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}

		systemSettingsMu.Lock()
		cfg := systemSettings.Portals[moduleID]
		migratePortalOperators(&cfg)
		idx := -1
		for i := range cfg.Operators {
			if cfg.Operators[i].ID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			systemSettingsMu.Unlock()
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Operator tidak ditemukan"})
			return
		}
		if strings.TrimSpace(body.Username) != "" {
			if operatorUsernameTaken(cfg, body.Username, id) {
				systemSettingsMu.Unlock()
				jsonResponse(w, http.StatusConflict, map[string]string{"error": "Username sudah dipakai di portal ini"})
				return
			}
			cfg.Operators[idx].Username = strings.TrimSpace(body.Username)
		}
		if strings.TrimSpace(body.Name) != "" {
			cfg.Operators[idx].Name = strings.TrimSpace(body.Name)
		}
		if body.Enabled != nil {
			cfg.Operators[idx].Enabled = *body.Enabled
		}
		if pwd := strings.TrimSpace(body.Password); pwd != "" && pwd != passwordMask {
			if len(pwd) < 6 {
				systemSettingsMu.Unlock()
				jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Password minimal 6 karakter"})
				return
			}
			cfg.Operators[idx].Password = applyPasswordIfProvided(cfg.Operators[idx].Password, pwd)
		}
		syncLegacyOperatorFields(&cfg)
		systemSettings.Portals[moduleID] = cfg
		if body.Perms != nil {
			systemSettings.OperatorPerms[moduleID] = *body.Perms
		}
		systemSettingsMu.Unlock()
		persistSystemSettings()
		recordAudit(sess.Username, "operator_update", moduleID, "Operator diperbarui: "+cfg.Operators[idx].Username, clientIP(r))
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"message":   "Operator berhasil diperbarui",
			"operators": maskOperators(cfg.Operators),
			"perms":     operatorPermsForModule(moduleID),
		})

	case http.MethodDelete:
		systemSettingsMu.Lock()
		cfg := systemSettings.Portals[moduleID]
		migratePortalOperators(&cfg)
		next := make([]PortalOperator, 0, len(cfg.Operators))
		foundName := ""
		for _, op := range cfg.Operators {
			if op.ID == id {
				foundName = op.Username
				continue
			}
			next = append(next, op)
		}
		if foundName == "" {
			systemSettingsMu.Unlock()
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Operator tidak ditemukan"})
			return
		}
		cfg.Operators = next
		syncLegacyOperatorFields(&cfg)
		systemSettings.Portals[moduleID] = cfg
		systemSettingsMu.Unlock()
		persistSystemSettings()
		recordAudit(sess.Username, "operator_delete", moduleID, "Operator dihapus: "+foundName, clientIP(r))
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"message":   "Operator berhasil dihapus",
			"operators": maskOperators(cfg.Operators),
			"perms":     operatorPermsForModule(moduleID),
		})

	default:
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func handleOperatorPerms(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if sess == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
		return
	}
	if r.Method != http.MethodPut {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	var perms OperatorPermissionSet
	if err := json.NewDecoder(r.Body).Decode(&perms); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	moduleID := sess.AppModule
	systemSettingsMu.Lock()
	systemSettings.OperatorPerms[moduleID] = perms
	systemSettingsMu.Unlock()
	persistSystemSettings()
	recordAudit(sess.Username, "operator_perms_update", moduleID, "Hak akses operator diperbarui", clientIP(r))
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Hak akses operator berhasil disimpan",
		"perms":   perms,
	})
}
