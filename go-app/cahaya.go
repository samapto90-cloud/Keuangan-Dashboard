package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"keuangan/mmo"
)

//go:embed all:cahaya-dist
var cahayaDistFS embed.FS

func mountCahayaGame(mux *http.ServeMux) {
	mux.HandleFunc("/cahaya/ws", mmo.HandleWS)
	mux.HandleFunc("/cahaya/api/register", mmo.HandleRegister)
	mux.HandleFunc("/cahaya/api/login", mmo.HandleLogin)
	mux.HandleFunc("/cahaya/api/logout", mmo.HandleLogout)
	mux.HandleFunc("/cahaya/api/reset-password", mmo.HandleResetPassword)
	mux.HandleFunc("/cahaya/api/profile", mmo.HandleProfile)
	sub, err := fs.Sub(cahayaDistFS, "cahaya-dist")
	if err != nil {
		log.Printf("cahaya-dist tidak bisa dilayani: %v", err)
		return
	}
	files := http.FileServer(http.FS(sub))
	mux.HandleFunc("/cahaya", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/cahaya/", http.StatusFound)
	})
	mux.Handle("/cahaya/", http.StripPrefix("/cahaya/", withCahayaStaticTypes(files)))
}

func withCahayaStaticTypes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" || p == "" || p == "/index.html" || strings.HasSuffix(p, ".html") {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		switch {
		case strings.HasSuffix(p, ".js"):
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		case strings.HasSuffix(p, ".css"):
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case strings.HasSuffix(p, ".svg"):
			w.Header().Set("Content-Type", "image/svg+xml")
		case strings.HasSuffix(p, ".png"):
			w.Header().Set("Content-Type", "image/png")
		}
		next.ServeHTTP(w, r)
	})
}
