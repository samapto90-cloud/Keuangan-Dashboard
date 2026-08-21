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
	mux.HandleFunc("/cahaya/api/realtime/poll", mmo.HandleRealtimePoll)
	mux.HandleFunc("/cahaya/api/realtime/send", mmo.HandleRealtimeSend)
	mux.HandleFunc("/cahaya/api/register", mmo.HandleRegister)
	mux.HandleFunc("/cahaya/api/login", mmo.HandleLogin)
	mux.HandleFunc("/cahaya/api/logout", mmo.HandleLogout)
	mux.HandleFunc("/cahaya/api/reset-password", mmo.HandleResetPassword)
	mux.HandleFunc("/cahaya/api/profile", mmo.HandleProfile)
	mux.HandleFunc("/cahaya/api/profile/stats", mmo.HandleProgressStats)
	mux.HandleFunc("/cahaya/api/profile/achievements", mmo.HandleProgressAchievements)
	mux.HandleFunc("/cahaya/api/profile/history", mmo.HandleProgressHistory)
	mux.HandleFunc("/cahaya/api/profile/update", mmo.HandleProgressUpdate)
	mux.HandleFunc("/cahaya/api/daily-reward", mmo.HandleDailyStatus)
	mux.HandleFunc("/cahaya/api/daily-reward/claim", mmo.HandleDailyClaim)
	mux.HandleFunc("/cahaya/api/admin/coins", mmo.HandleAdminAdjustCoins)
	mux.HandleFunc("/cahaya/api/friends", mmo.HandleFriends)
	mux.HandleFunc("/cahaya/api/friends/search", mmo.HandleFriendSearch)
	mux.HandleFunc("/cahaya/api/friends/request", mmo.HandleFriendAction)
	mux.HandleFunc("/cahaya/api/friends/respond", mmo.HandleFriendAction)
	mux.HandleFunc("/cahaya/api/friends/remove", mmo.HandleFriendAction)
	mux.HandleFunc("/cahaya/api/friends/block", mmo.HandleFriendAction)
	mux.HandleFunc("/cahaya/api/friends/unblock", mmo.HandleFriendAction)
	mux.HandleFunc("/cahaya/api/players/public", mmo.HandlePublicPlayer)
	mux.HandleFunc("/cahaya/api/notifications", mmo.HandleNotifications)
	mux.HandleFunc("/cahaya/api/privacy", mmo.HandlePrivacy)
	mux.HandleFunc("/cahaya/api/report", mmo.HandleReportPlayer)
	mux.HandleFunc("/cahaya/api/feedback", mmo.HandleFeedback)
	mux.HandleFunc("/cahaya/api/season", mmo.HandleSeason)
	mux.HandleFunc("/cahaya/api/rank", mmo.HandleRankMe)
	mux.HandleFunc("/cahaya/api/rank/history", mmo.HandleRankHistoryHTTP)
	mux.HandleFunc("/cahaya/api/leaderboard", mmo.HandleLeaderboard)
	mux.HandleFunc("/cahaya/api/ular/board", mmo.HandleUlarBoard)
	mux.HandleFunc("/cahaya/api/ular/resolve", mmo.HandleUlarResolve)
	mux.HandleFunc("/cahaya/api/ular/questions/validate", mmo.HandleQuestionValidate)
	mux.HandleFunc("/cahaya/api/ular/questions/stats", mmo.HandleQuestionStats)
	mux.HandleFunc("/cahaya/api/ular/practice/question", mmo.HandlePracticeQuestion)
	mux.HandleFunc("/cahaya/api/ular/practice/answer", mmo.HandlePracticeAnswer)
	mux.HandleFunc("/cahaya/api/admin/", mmo.HandleAdminAPI)
	mux.HandleFunc("/cahaya/api/admin", mmo.HandleAdminAPI)
	mux.HandleFunc("/cahaya/api/ular/admin/questions", mmo.HandleAdminQuestions)
	mux.HandleFunc("/admin/api/", mmo.HandleAdminAPI)
	mux.HandleFunc("/admin/api", mmo.HandleAdminAPI)
	mux.HandleFunc("/admin", serveCahayaAdmin)
	mux.HandleFunc("/admin/", serveCahayaAdmin)
	mux.HandleFunc("/cahaya/admin", serveCahayaAdmin)
	mux.HandleFunc("/cahaya/admin/", serveCahayaAdmin)
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

func serveCahayaAdmin(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/admin/api") {
		mmo.HandleAdminAPI(w, r)
		return
	}
	raw, err := cahayaDistFS.ReadFile("cahaya-dist/index.html")
	if err != nil {
		http.Error(w, "admin tidak tersedia", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
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
		case strings.HasSuffix(p, ".webmanifest"):
			w.Header().Set("Content-Type", "application/manifest+json")
		case strings.HasSuffix(p, "sw.js"):
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			w.Header().Set("Service-Worker-Allowed", "/cahaya/")
		}
		next.ServeHTTP(w, r)
	})
}
