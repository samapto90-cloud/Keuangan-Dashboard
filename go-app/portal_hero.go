package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	portalHeroAssetName   = "portal-intro.mp4"
	portalHeroCustomName  = "custom.mp4"
	portalHeroMetaName    = "meta.json"
	portalHeroMaxUploadMB = 48
)

type portalHeroMeta struct {
	Version      string `json:"version"`
	Source       string `json:"source"` // custom | default
	OriginalName string `json:"original_name,omitempty"`
	SizeBytes    int64  `json:"size_bytes"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	UpdatedBy    string `json:"updated_by,omitempty"`
	ContentType  string `json:"content_type,omitempty"`
}

var (
	portalHeroMetaMu sync.RWMutex
	portalHeroCached *portalHeroMeta
	portalHeroEmbed  http.Handler
)

func portalHeroDir() string {
	return filepath.Join(dataDir, "portal-hero")
}

func portalHeroCustomPath() string {
	return filepath.Join(portalHeroDir(), portalHeroCustomName)
}

func portalHeroMetaPath() string {
	return filepath.Join(portalHeroDir(), portalHeroMetaName)
}

func initPortalHeroServe(embedRoot fs.FS) {
	sub, err := fs.Sub(embedRoot, "assets/portal-hero")
	if err != nil {
		return
	}
	portalHeroEmbed = http.FileServer(http.FS(sub))
}

func loadPortalHeroMetaFromDisk() portalHeroMeta {
	custom := portalHeroCustomPath()
	st, err := os.Stat(custom)
	if err != nil || st.IsDir() || st.Size() == 0 {
		return portalHeroMeta{
			Version: "default",
			Source:  "default",
		}
	}
	meta := portalHeroMeta{
		Version:   fmt.Sprintf("%d", st.ModTime().UnixMilli()),
		Source:    "custom",
		SizeBytes: st.Size(),
		UpdatedAt: st.ModTime().UTC().Format(time.RFC3339),
	}
	raw, err := os.ReadFile(portalHeroMetaPath())
	if err == nil {
		var saved portalHeroMeta
		if json.Unmarshal(raw, &saved) == nil && saved.Source == "custom" {
			if saved.Version != "" {
				meta.Version = saved.Version
			}
			meta.OriginalName = saved.OriginalName
			meta.UpdatedBy = saved.UpdatedBy
			meta.ContentType = saved.ContentType
			if saved.UpdatedAt != "" {
				meta.UpdatedAt = saved.UpdatedAt
			}
			if saved.SizeBytes > 0 {
				meta.SizeBytes = saved.SizeBytes
			}
		}
	}
	return meta
}

func getPortalHeroMeta() portalHeroMeta {
	portalHeroMetaMu.RLock()
	if portalHeroCached != nil {
		m := *portalHeroCached
		portalHeroMetaMu.RUnlock()
		return m
	}
	portalHeroMetaMu.RUnlock()

	portalHeroMetaMu.Lock()
	defer portalHeroMetaMu.Unlock()
	if portalHeroCached != nil {
		return *portalHeroCached
	}
	m := loadPortalHeroMetaFromDisk()
	portalHeroCached = &m
	return m
}

func invalidatePortalHeroMetaCache() {
	portalHeroMetaMu.Lock()
	portalHeroCached = nil
	portalHeroMetaMu.Unlock()
}

func persistPortalHeroMeta(m portalHeroMeta) error {
	if err := os.MkdirAll(portalHeroDir(), 0o755); err != nil {
		return err
	}
	return writeJSONAtomic(portalHeroMetaPath(), m)
}

func portalHeroPublicURL(m portalHeroMeta) string {
	v := strings.TrimSpace(m.Version)
	if v == "" {
		v = "default"
	}
	return "/assets/portal-hero/" + portalHeroAssetName + "?v=" + v
}

func portalHeroStatusPayload() map[string]interface{} {
	m := getPortalHeroMeta()
	return map[string]interface{}{
		"url":           portalHeroPublicURL(m),
		"source":        m.Source,
		"version":       m.Version,
		"original_name": m.OriginalName,
		"size_bytes":    m.SizeBytes,
		"updated_at":    m.UpdatedAt,
		"updated_by":    m.UpdatedBy,
		"content_type":  m.ContentType,
		"max_upload_mb": portalHeroMaxUploadMB,
	}
}

func portalHeroPublicStatusPayload() map[string]interface{} {
	m := getPortalHeroMeta()
	return map[string]interface{}{
		"url":           portalHeroPublicURL(m),
		"source":        m.Source,
		"version":       m.Version,
		"size_bytes":    m.SizeBytes,
		"updated_at":    m.UpdatedAt,
		"max_upload_mb": portalHeroMaxUploadMB,
	}
}

func handlePortalHeroPublic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
	jsonResponse(w, http.StatusOK, portalHeroPublicStatusPayload())
}

func handleAdminPortalHero(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonResponse(w, http.StatusOK, portalHeroStatusPayload())
	case http.MethodPost, http.MethodPut:
		handleAdminPortalHeroUpload(w, r)
	case http.MethodDelete:
		handleAdminPortalHeroReset(w, r)
	default:
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func handleAdminPortalHeroUpload(w http.ResponseWriter, r *http.Request) {
	maxBytes := int64(portalHeroMaxUploadMB) << 20
	// Body sudah dibatasi middleware (file + slack multipart). Jangan wrap MaxBytesReader lagi.
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Upload gagal / terlalu besar (maks %d MB)", portalHeroMaxUploadMB),
		})
		return
	}
	file, hdr, err := r.FormFile("video")
	if err != nil {
		file, hdr, err = r.FormFile("file")
	}
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "File video wajib (field: video)"})
		return
	}
	defer file.Close()

	name := strings.ToLower(strings.TrimSpace(hdr.Filename))
	ct := strings.ToLower(strings.TrimSpace(hdr.Header.Get("Content-Type")))
	if !isAllowedPortalHeroUpload(name, ct) {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Format harus MP4 (video/mp4)"})
		return
	}
	if hdr.Size > maxBytes && hdr.Size > 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Ukuran melebihi %d MB", portalHeroMaxUploadMB),
		})
		return
	}

	if err := os.MkdirAll(portalHeroDir(), 0o755); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal siapkan folder video"})
		return
	}

	tmp := portalHeroCustomPath() + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal menulis file"})
		return
	}
	written, copyErr := io.Copy(out, io.LimitReader(file, maxBytes+1))
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil {
		os.Remove(tmp)
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal menyimpan video"})
		return
	}
	if written == 0 {
		os.Remove(tmp)
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "File video kosong"})
		return
	}
	if written > maxBytes {
		os.Remove(tmp)
		jsonResponse(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Ukuran melebihi %d MB", portalHeroMaxUploadMB),
		})
		return
	}
	if !fileLooksLikeMP4(tmp) {
		_ = os.Remove(tmp)
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "File bukan MP4 yang valid (header ftyp tidak ditemukan)"})
		return
	}
	if err := os.Rename(tmp, portalHeroCustomPath()); err != nil {
		os.Remove(tmp)
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Gagal mengaktifkan video"})
		return
	}

	sess := getSession(r)
	actor := ""
	if sess != nil {
		actor = sess.Username
	}
	now := time.Now().UTC()
	meta := portalHeroMeta{
		Version:      fmt.Sprintf("%d", now.UnixMilli()),
		Source:       "custom",
		OriginalName: filepath.Base(hdr.Filename),
		SizeBytes:    written,
		UpdatedAt:    now.Format(time.RFC3339),
		UpdatedBy:    actor,
		ContentType:  "video/mp4",
	}
	_ = persistPortalHeroMeta(meta)
	invalidatePortalHeroMetaCache()
	invalidatePortalStatusCache()

	if sess != nil {
		recordAudit(sess.Username, "portal_hero_upload", "pengaturan",
			fmt.Sprintf("Video portal diganti (%s, %d bytes)", meta.OriginalName, meta.SizeBytes), clientIP(r))
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Video portal berhasil diganti",
		"hero":    portalHeroStatusPayload(),
	})
}

func handleAdminPortalHeroReset(w http.ResponseWriter, r *http.Request) {
	_ = os.Remove(portalHeroCustomPath())
	_ = os.Remove(portalHeroMetaPath())
	invalidatePortalHeroMetaCache()
	invalidatePortalStatusCache()
	sess := getSession(r)
	if sess != nil {
		recordAudit(sess.Username, "portal_hero_reset", "pengaturan", "Video portal dikembalikan ke default", clientIP(r))
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Video portal dikembalikan ke bawaan sistem",
		"hero":    portalHeroStatusPayload(),
	})
}

func isAllowedPortalHeroUpload(filename, contentType string) bool {
	if strings.HasSuffix(filename, ".mp4") {
		return true
	}
	switch contentType {
	case "video/mp4", "application/mp4", "video/quicktime":
		return true
	default:
		return false
	}
}

func fileLooksLikeMP4(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 64)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false
	}
	return bytes.Contains(buf[:n], []byte("ftyp"))
}

func servePortalHeroAssets(w http.ResponseWriter, r *http.Request) {
	rel := strings.TrimPrefix(r.URL.Path, "/assets/portal-hero/")
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || strings.Contains(rel, "..") {
		http.NotFound(w, r)
		return
	}
	base := filepath.Base(rel)
	if base == portalHeroAssetName || base == portalHeroCustomName {
		custom := portalHeroCustomPath()
		if st, err := os.Stat(custom); err == nil && !st.IsDir() && st.Size() > 0 {
			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Cache-Control", "private, no-cache, must-revalidate")
			w.Header().Set("X-Portal-Hero-Source", "custom")
			http.ServeFile(w, r, custom)
			return
		}
		rel = portalHeroAssetName
	}
	if portalHeroEmbed == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Portal-Hero-Source", "default")
	r2 := r.Clone(r.Context())
	r2.URL.Path = "/" + rel
	portalHeroEmbed.ServeHTTP(w, r2)
}
