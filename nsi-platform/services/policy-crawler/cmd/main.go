package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/admin"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/scheduler"
	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
)

func main() {
	cfg := config.Load()

	crawlScheduler := scheduler.NewScheduler(24 * time.Hour)
	crawlScheduler.Task = func() {
		log.Println("crawl cycle started")
		// Fetch from sources, parse, verify — implemented in subsequent sprints
	}
	crawlScheduler.Start()

	mux := http.NewServeMux()
	mux.Handle("/admin/claims", admin.ListClaimsHandler(nil))
	mux.Handle("/admin/claims/", admin.UpdateClaimHandler(nil))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":30002"
	}

	log.Printf("policy-crawler starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
