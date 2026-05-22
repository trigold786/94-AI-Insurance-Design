package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/server"
	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/calculate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"code":"METHOD_NOT_ALLOWED"}`, http.StatusMethodNotAllowed)
			return
		}

		var req server.PlanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"code":"VALIDATION_ERROR","message":"invalid JSON"}`, http.StatusBadRequest)
			return
		}

		resp := server.CalculatePlan(req)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":50051"
	}

	log.Printf("actuarial-engine (HTTP) starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
