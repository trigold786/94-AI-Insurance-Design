package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/server"
	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
)

func main() {
	_ = config.Load()

	var cache server.Cache = server.NoopCache{}
	if enabled, _ := strconv.ParseBool(os.Getenv("CALC_CACHE_ENABLED")); enabled {
		cache = server.NewInMemoryCache(10 * time.Minute)
		log.Println("in-memory calculation cache enabled (TTL: 24h, cleanup: 10m)")
	}

	mux := http.NewServeMux()
	actuarySecret := os.Getenv("ACTUARY_SECRET")

	mux.HandleFunc("/v1/calculate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"code":"METHOD_NOT_ALLOWED"}`, http.StatusMethodNotAllowed)
			return
		}

		if actuarySecret != "" {
			if r.Header.Get("X-Actuary-Secret") != actuarySecret {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"invalid actuary secret"}`, http.StatusUnauthorized)
				return
			}
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req server.PlanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"code":"VALIDATION_ERROR","message":"invalid JSON"}`, http.StatusBadRequest)
			return
		}

		resp := server.CalculatePlan(req, cache)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":39402"
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("actuarial-engine (HTTP) starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
