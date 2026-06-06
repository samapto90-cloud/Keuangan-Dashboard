package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

var (
	indexETag     string
	indexHTMLGzip []byte
)

func initIndexCache() {
	sum := sha256.Sum256(indexHTML)
	indexETag = `"` + hex.EncodeToString(sum[:8]) + `"`

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write(indexHTML)
	_ = gw.Close()
	indexHTMLGzip = buf.Bytes()
}

func serveIndexHTML(w http.ResponseWriter, r *http.Request) {
	if inm := strings.TrimSpace(r.Header.Get("If-None-Match")); inm == indexETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("ETag", indexETag)
	w.Header().Set("Cache-Control", "no-cache")
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && len(indexHTMLGzip) > 0 {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexHTMLGzip)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(indexHTML)
}

func withStaticCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		h.ServeHTTP(w, r)
	})
}
