package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const passwordMask = "********"

var sipkeuPortalIDs = []string{"sekretariat", "paud", "sd", "smp", "kas-belanja"}

type PortalAuthConfig struct {
	AdminUsername    string `json:"admin_username"`
	AdminPassword    string `json:"admin_password"`
	AdminName        string `json:"admin_name"`
	OperatorUsername string `json:"operator_username"`
	OperatorPassword string `json:"operator_password"`
	OperatorName     string `json:"operator_name"`
}

type OperatorPermissionSet struct {
	ViewDashboard   bool `json:"view_dashboard"`
	ViewRekap       bool `json:"view_rekap"`
	ViewPejabat     bool `json:"view_pejabat"`
	AddTransaksi    bool `json:"add_transaksi"`
	EditTransaksi   bool `json:"edit_transaksi"`
	DeleteTransaksi bool `json:"delete_transaksi"`
	DeleteBulk      bool `json:"delete_bulk"`
	DeleteAll       bool `json:"delete_all"`
	ExportTransaksi bool `json:"export_transaksi"`
	ImportTransaksi bool `json:"import_transaksi"`
	ImportAnggaran  bool `json:"import_anggaran"`
	ViewRegister    bool `json:"view_register"`
	CetakKwitansi   bool `json:"cetak_kwitansi"`
	CetakNp2d       bool `json:"cetak_np2d"`
}

type SystemSettings struct {
	SettingsPortalUsername string                           `json:"settings_portal_username"`
	SettingsPortalPassword string                           `json:"settings_portal_password"`
	SettingsPortalName     string                           `json:"settings_portal_name"`
	Portals                map[string]PortalAuthConfig      `json:"portals"`
	OperatorPerms          map[string]OperatorPermissionSet `json:"operator_perms"`
}

var (
	systemSettings   SystemSettings
	systemSettingsMu sync.RWMutex
)

func defaultOperatorPerms() OperatorPermissionSet {
	return OperatorPermissionSet{
		AddTransaksi:    true,
		EditTransaksi:   true,
		DeleteTransaksi: true,
		ViewRegister:    true,
		CetakKwitansi:   true,
		CetakNp2d:       true,
	}
}

func defaultPortalAuth() PortalAuthConfig {
	return PortalAuthConfig{
		AdminUsername:    envOr("SIPKEU_ADMIN_USER", "admin"),
		AdminPassword:    envOr("SIPKEU_ADMIN_PASSWORD", "admin2026"),
		AdminName:        "Administrator SIPKEU",
		OperatorUsername: envOr("SIPKEU_OPERATOR_USER", "operator"),
		OperatorPassword: envOr("SIPKEU_OPERATOR_PASSWORD", "operator2026"),
		OperatorName:     "Operator SIPKEU",
	}
}

func clearOperatorAuth(cfg *PortalAuthConfig) {
	cfg.OperatorUsername = ""
	cfg.OperatorPassword = ""
	cfg.OperatorName = ""
}

func sanitizeKasBelanjaSettings(s *SystemSettings) {
	if s == nil {
		return
	}
	if s.OperatorPerms != nil {
		delete(s.OperatorPerms, "kas-belanja")
	}
	if s.Portals != nil {
		if cfg, ok := s.Portals["kas-belanja"]; ok {
			clearOperatorAuth(&cfg)
			s.Portals["kas-belanja"] = cfg
		}
	}
}

func defaultSystemSettings() SystemSettings {
	portals := map[string]PortalAuthConfig{}
	perms := map[string]OperatorPermissionSet{}
	auth := defaultPortalAuth()
	for _, id := range sipkeuPortalIDs {
		p := auth
		if id == "kas-belanja" {
			clearOperatorAuth(&p)
		}
		portals[id] = p
		if id != "kas-belanja" {
			perms[id] = defaultOperatorPerms()
		}
	}
	return SystemSettings{
		SettingsPortalUsername: "199010132019031001",
		SettingsPortalPassword: "Hasanah050393",
		SettingsPortalName:     "Administrator Sistem SIPKEU",
		Portals:                portals,
		OperatorPerms:          perms,
	}
}

func systemSettingsPath() string {
	return filepath.Join(dataDir, "system-settings.json")
}

func initSystemSettings() {
	systemSettings = defaultSystemSettings()
	hashPasswordsInSettings(&systemSettings)
	raw, err := os.ReadFile(systemSettingsPath())
	if err != nil {
		persistSystemSettings()
		log.Printf("Pengaturan sistem: file baru dibuat di %s", systemSettingsPath())
		return
	}
	var loaded SystemSettings
	if err := json.Unmarshal(raw, &loaded); err != nil {
		log.Printf("Peringatan: system-settings.json rusak, pakai default: %v", err)
		persistSystemSettings()
		return
	}
	mergeSystemSettings(&loaded)
	if hashPasswordsInSettings(&loaded) {
		systemSettings = loaded
		persistSystemSettings()
		log.Printf("Password pengaturan sistem di-hash (migrasi keamanan)")
	} else {
		systemSettings = loaded
	}
	log.Printf("Pengaturan sistem dimuat dari %s", systemSettingsPath())
}

func mergeSystemSettings(s *SystemSettings) {
	def := defaultSystemSettings()
	if strings.TrimSpace(s.SettingsPortalUsername) == "" {
		s.SettingsPortalUsername = def.SettingsPortalUsername
	}
	if strings.TrimSpace(s.SettingsPortalPassword) == "" {
		s.SettingsPortalPassword = def.SettingsPortalPassword
	}
	if strings.TrimSpace(s.SettingsPortalName) == "" {
		s.SettingsPortalName = def.SettingsPortalName
	}
	if s.Portals == nil {
		s.Portals = map[string]PortalAuthConfig{}
	}
	if s.OperatorPerms == nil {
		s.OperatorPerms = map[string]OperatorPermissionSet{}
	}
	for _, id := range sipkeuPortalIDs {
		if _, ok := s.Portals[id]; !ok {
			s.Portals[id] = def.Portals[id]
		}
		if id != "kas-belanja" {
			if _, ok := s.OperatorPerms[id]; !ok {
				s.OperatorPerms[id] = def.OperatorPerms[id]
			}
		}
	}
	sanitizeKasBelanjaSettings(s)
}

func isBcryptHash(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}

func hashPasswordsInSettings(s *SystemSettings) bool {
	if s == nil {
		return false
	}
	changed := false
	if !isBcryptHash(s.SettingsPortalPassword) && strings.TrimSpace(s.SettingsPortalPassword) != "" {
		if h, err := hashPasswordStore(s.SettingsPortalPassword); err == nil && h != "" {
			s.SettingsPortalPassword = h
			changed = true
		}
	}
	for id, cfg := range s.Portals {
		if !isBcryptHash(cfg.AdminPassword) && strings.TrimSpace(cfg.AdminPassword) != "" {
			if h, err := hashPasswordStore(cfg.AdminPassword); err == nil && h != "" {
				cfg.AdminPassword = h
				changed = true
			}
		}
		if strings.TrimSpace(cfg.OperatorUsername) != "" && !isBcryptHash(cfg.OperatorPassword) && strings.TrimSpace(cfg.OperatorPassword) != "" {
			if h, err := hashPasswordStore(cfg.OperatorPassword); err == nil && h != "" {
				cfg.OperatorPassword = h
				changed = true
			}
		}
		s.Portals[id] = cfg
	}
	return changed
}

func persistSystemSettings() {
	systemSettingsMu.RLock()
	snap := systemSettings
	systemSettingsMu.RUnlock()
	if err := writeJSONAtomic(systemSettingsPath(), snap); err != nil {
		log.Printf("Gagal simpan pengaturan sistem: %v", err)
	}
}

func getSystemSettingsCopy() SystemSettings {
	systemSettingsMu.RLock()
	defer systemSettingsMu.RUnlock()
	return systemSettings
}

func authenticatePortalUser(username, password, appModule string) (UserAccount, bool) {
	username = strings.ToLower(strings.TrimSpace(username))
	sys := getSystemSettingsCopy()

	if appModule == "pengaturan" {
		if username == strings.ToLower(strings.TrimSpace(sys.SettingsPortalUsername)) &&
			passwordMatches(sys.SettingsPortalPassword, password) {
			return UserAccount{
				Password: sys.SettingsPortalPassword,
				Role:     "settings-admin",
				Name:     sys.SettingsPortalName,
			}, true
		}
		return UserAccount{}, false
	}

	cfg, ok := sys.Portals[appModule]
	if !ok {
		user, found := defaultUsers[username]
		if found && passwordMatches(user.Password, password) {
			return user, true
		}
		return UserAccount{}, false
	}

	if username == strings.ToLower(strings.TrimSpace(cfg.AdminUsername)) &&
		passwordMatches(cfg.AdminPassword, password) {
		return UserAccount{
			Password: cfg.AdminPassword,
			Role:     "admin",
			Name:     firstNonEmpty(cfg.AdminName, "Administrator"),
		}, true
	}
	if appModule != "kas-belanja" &&
		username == strings.ToLower(strings.TrimSpace(cfg.OperatorUsername)) &&
		passwordMatches(cfg.OperatorPassword, password) {
		return UserAccount{
			Password: cfg.OperatorPassword,
			Role:     "operator",
			Name:     firstNonEmpty(cfg.OperatorName, "Operator"),
		}, true
	}
	return UserAccount{}, false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func operatorPermsForModule(moduleID string) OperatorPermissionSet {
	sys := getSystemSettingsCopy()
	if p, ok := sys.OperatorPerms[moduleID]; ok {
		return p
	}
	return defaultOperatorPerms()
}

func allOperatorPermsTrue() OperatorPermissionSet {
	return OperatorPermissionSet{
		ViewDashboard: true, ViewRekap: true, ViewPejabat: true,
		AddTransaksi: true, EditTransaksi: true, DeleteTransaksi: true,
		DeleteBulk: true, DeleteAll: true, ExportTransaksi: true,
		ImportTransaksi: true, ImportAnggaran: true, ViewRegister: true,
		CetakKwitansi: true, CetakNp2d: true,
	}
}

func sessionHasPermission(sess *Session, perm string) bool {
	if sess == nil {
		return false
	}
	if sess.Role == "admin" || sess.Role == "settings-admin" {
		return true
	}
	if sess.Role != "operator" {
		return false
	}
	p := operatorPermsForModule(sess.AppModule)
	switch perm {
	case "view_dashboard":
		return p.ViewDashboard
	case "view_rekap":
		return p.ViewRekap
	case "view_pejabat":
		return p.ViewPejabat
	case "add_transaksi":
		return p.AddTransaksi
	case "edit_transaksi":
		return p.EditTransaksi
	case "delete_transaksi":
		return p.DeleteTransaksi
	case "delete_bulk":
		return p.DeleteBulk
	case "delete_all":
		return p.DeleteAll
	case "export_transaksi":
		return p.ExportTransaksi
	case "import_transaksi":
		return p.ImportTransaksi
	case "import_anggaran":
		return p.ImportAnggaran
	case "view_register":
		return p.ViewRegister
	case "cetak_kwitansi":
		return p.CetakKwitansi
	case "cetak_np2d":
		return p.CetakNp2d
	default:
		return false
	}
}

func requirePermission(perm string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sess := getSession(r)
			if sess == nil {
				jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
				return
			}
			if !sessionHasPermission(sess, perm) {
				jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses ditolak — hak operator tidak mencukupi"})
				return
			}
			next(w, r)
		}
	}
}

func requireSettingsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := getSession(r)
		if sess == nil {
			jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
			return
		}
		if sess.Role != "settings-admin" || sess.AppModule != "pengaturan" {
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses hanya untuk Portal Pengaturan Sistem"})
			return
		}
		next(w, r)
	}
}

func maskPortalAuth(cfg PortalAuthConfig) map[string]interface{} {
	out := map[string]interface{}{
		"admin_username": cfg.AdminUsername,
		"admin_password": passwordMask,
		"admin_name":     cfg.AdminName,
	}
	if strings.TrimSpace(cfg.OperatorUsername) != "" {
		out["operator_username"] = cfg.OperatorUsername
		out["operator_password"] = passwordMask
		out["operator_name"] = cfg.OperatorName
	} else {
		out["operator_username"] = ""
		out["operator_password"] = ""
		out["operator_name"] = ""
	}
	return out
}

func systemSettingsPublicResponse() map[string]interface{} {
	sys := getSystemSettingsCopy()
	portals := map[string]interface{}{}
	for id, cfg := range sys.Portals {
		portals[id] = maskPortalAuth(cfg)
	}
	return map[string]interface{}{
		"settings_portal_username": sys.SettingsPortalUsername,
		"settings_portal_password": passwordMask,
		"settings_portal_name":     sys.SettingsPortalName,
		"portals":                  portals,
		"operator_perms":           sys.OperatorPerms,
	}
}

func applyPasswordIfProvided(current, incoming string) string {
	return storePasswordIfProvided(current, incoming)
}

func handleSystemSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if sess := getSession(r); sess == nil || sess.Role != "settings-admin" {
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses hanya untuk Portal Pengaturan Sistem"})
			return
		}
		jsonResponse(w, http.StatusOK, systemSettingsPublicResponse())

	case http.MethodPut:
		if sess := getSession(r); sess == nil || sess.Role != "settings-admin" {
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses hanya untuk Portal Pengaturan Sistem"})
			return
		}
		var incoming SystemSettings
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}

		systemSettingsMu.Lock()
		cur := systemSettings

		if strings.TrimSpace(incoming.SettingsPortalUsername) != "" {
			cur.SettingsPortalUsername = strings.TrimSpace(incoming.SettingsPortalUsername)
		}
		cur.SettingsPortalPassword = applyPasswordIfProvided(cur.SettingsPortalPassword, incoming.SettingsPortalPassword)
		if strings.TrimSpace(incoming.SettingsPortalName) != "" {
			cur.SettingsPortalName = strings.TrimSpace(incoming.SettingsPortalName)
		}

		if incoming.Portals != nil {
			for id, p := range incoming.Portals {
				if !containsPortalID(id) {
					continue
				}
				existing := cur.Portals[id]
				if strings.TrimSpace(p.AdminUsername) != "" {
					existing.AdminUsername = strings.TrimSpace(p.AdminUsername)
				}
				existing.AdminPassword = applyPasswordIfProvided(existing.AdminPassword, p.AdminPassword)
				if strings.TrimSpace(p.AdminName) != "" {
					existing.AdminName = strings.TrimSpace(p.AdminName)
				}
				if id == "kas-belanja" {
					clearOperatorAuth(&existing)
				} else {
					if strings.TrimSpace(p.OperatorUsername) != "" {
						existing.OperatorUsername = strings.TrimSpace(p.OperatorUsername)
					}
					existing.OperatorPassword = applyPasswordIfProvided(existing.OperatorPassword, p.OperatorPassword)
					if strings.TrimSpace(p.OperatorName) != "" {
						existing.OperatorName = strings.TrimSpace(p.OperatorName)
					}
				}
				cur.Portals[id] = existing
			}
		}

		if incoming.OperatorPerms != nil {
			for id, p := range incoming.OperatorPerms {
				if containsPortalID(id) && id != "kas-belanja" {
					cur.OperatorPerms[id] = p
				}
			}
		}

		sanitizeKasBelanjaSettings(&cur)
		hashPasswordsInSettings(&cur)

		systemSettings = cur
		systemSettingsMu.Unlock()
		persistSystemSettings()
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"message": "Pengaturan sistem berhasil disimpan",
			"data":    systemSettingsPublicResponse(),
		})

	default:
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func containsPortalID(id string) bool {
	for _, p := range sipkeuPortalIDs {
		if p == id {
			return true
		}
	}
	return false
}

func permissionsForSession(sess *Session) OperatorPermissionSet {
	if sess == nil {
		return OperatorPermissionSet{}
	}
	if sess.Role == "admin" || sess.Role == "settings-admin" {
		return allOperatorPermsTrue()
	}
	return operatorPermsForModule(sess.AppModule)
}
