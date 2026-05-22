package main

import (
	"log"
	"net/http"
	"os"

	"github.com/trigold786/94-AI-Insurance-Design/api-server/internal/handler"
	"github.com/trigold786/94-AI-Insurance-Design/api-server/internal/repository"
	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
	"github.com/trigold786/94-AI-Insurance-Design/shared/db"
	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	profileRepo, err := repository.NewProfileRepository(database)
	if err != nil {
		log.Fatalf("failed to create profile repository: %v", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/healthz", handler.HealthCheckHandler())

	profileHandler := methodRouter(map[string]http.Handler{
		http.MethodGet:  handler.GetProfileHandler(profileRepo),
		http.MethodPut:  handler.UpdateProfileHandler(profileRepo),
	})
	mux.Handle("/v1/profile", middleware.Chain(
		middleware.RecoveryMiddleware(),
		middleware.AuthMiddleware,
		profileHandler,
	))

	// Stub routes — implemented in subsequent sprints
	mux.Handle("/v1/policies", middleware.Chain(
		middleware.RecoveryMiddleware(),
		stubHandler(http.StatusOK, `{"code":0,"data":[]}`),
	))
	mux.Handle("/v1/plans/", middleware.Chain(
		middleware.RecoveryMiddleware(),
		stubHandler(http.StatusOK, `{"code":0,"data":{}}`),
	))

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":30001"
	}

	log.Printf("api-server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func methodRouter(handlers map[string]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := handlers[r.Method]
		if !ok {
			http.Error(w, `{"code":"METHOD_NOT_ALLOWED","message":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func stubHandler(status int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write([]byte(body))
	})
}
