package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/api-server/internal/handler"
	"github.com/trigold786/94-AI-Insurance-Design/api-server/internal/repository"
	"github.com/trigold786/94-AI-Insurance-Design/shared/auth"
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

	policyRepo, err := repository.NewPolicyRepository(database)
	if err != nil {
		log.Fatalf("failed to create policy repository: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/auth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": "METHOD_NOT_ALLOWED"})
			return
		}
		var req struct {
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "BAD_REQUEST", "message": "user_id required"})
			return
		}
		token, err := auth.GenerateToken(cfg.JWTSecret, req.UserID, 24*time.Hour)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{"code": "INTERNAL_ERROR"})
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": map[string]string{"token": token}})
	})

	mux.Handle("/healthz", handler.HealthCheckHandler())

	authMW := middleware.AuthMiddleware(cfg.JWTSecret)

	profileHandler := middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(methodRouter(map[string]http.Handler{
		http.MethodGet:  handler.GetProfileHandler(profileRepo),
		http.MethodPut:  handler.UpdateProfileHandler(profileRepo),
	}))

	mux.Handle("/v1/profile", profileHandler)

	mux.Handle("/v1/policies", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.QueryPoliciesHandler(policyRepo)))

	mux.Handle("/v1/compliance/checklist", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.ComplianceChecklistHandler(&handler.ComplianceEvaluator{}, policyRepo, profileRepo)))

	mux.Handle("/webclient", handler.WebClientHandler(cfg.WebClientAPIBaseURL))

	mux.Handle("/v1/guide", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.GuideHandler(&handler.ComplianceEvaluator{}, policyRepo, profileRepo)))

	rightsRepo, err := repository.NewRightsRepository(database)
	if err != nil {
		log.Fatalf("failed to create rights repository: %v", err)
	}

	mux.Handle("/v1/rights/payment-status", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.PaymentStatusHandler(rightsRepo)))

	mux.Handle("/v1/rights/alerts", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.AlertListHandler(rightsRepo)))

	mux.Handle("/v1/rights/alerts/read", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.MarkAlertReadHandler(rightsRepo)))

	mux.Handle("/v1/rights/payment-records", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.SubmitPaymentRecordHandler(rightsRepo)))

	scheduler := handler.NewAlertScheduler(rightsRepo)
	scheduler.Start(24 * time.Hour)
	defer scheduler.Stop()

	feedbackRepo, err := repository.NewFeedbackRepo(database)
	if err != nil {
		log.Fatalf("failed to create feedback repo: %v", err)
	}
	mux.Handle("/v1/feedback", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.SubmitFeedbackHandler(feedbackRepo)))

	planRepo, err := repository.NewPlanRepository(database)
	if err != nil {
		log.Fatalf("failed to create plan repository: %v", err)
	}

	mux.Handle("/v1/plans/generate", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.GeneratePlanHandler(cfg.LLMGatewayURL, planRepo, profileRepo, policyRepo)))

	mux.Handle("/v1/plans/", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.PlanDetailHandler(planRepo)))

	mux.Handle("/v1/plans/report/pdf", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.PlanReportPDFHandler(planRepo, policyRepo, profileRepo)))

	mux.Handle("/v1/plans/report", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.PlanReportHandler(planRepo, policyRepo)))

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	if envAddr := os.Getenv("LISTEN_ADDR"); envAddr != "" {
		addr = envAddr
	}

	tlsCert := os.Getenv("TLS_CERT")
	tlsKey := os.Getenv("TLS_KEY")

	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	h := middleware.Chain(
		middleware.SecurityHeadersMiddleware(),
		middleware.CORSMiddleware(cfg.AllowedOrigins),
		middleware.RateLimitMiddleware(rateLimiter),
	)(mux)

	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("api-server starting on %s", addr)
		var err error
		if tlsCert != "" && tlsKey != "" {
			err = srv.ListenAndServeTLS(tlsCert, tlsKey)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
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

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}


