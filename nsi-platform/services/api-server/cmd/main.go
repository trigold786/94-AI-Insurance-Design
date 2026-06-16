package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/api-server/internal/handler"
	"github.com/trigold786/94-AI-Insurance-Design/api-server/internal/repository"
	"github.com/trigold786/94-AI-Insurance-Design/shared/auth"
	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
	"github.com/trigold786/94-AI-Insurance-Design/shared/crypto"
	"github.com/trigold786/94-AI-Insurance-Design/shared/db"
	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/notifier"
)

func main() {
	cfg := config.Load()
	crypto.Init()

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

	codeStore := handler.NewMemoryCodeStore()

	mux.HandleFunc("/v1/auth/sms/send", func(w http.ResponseWriter, r *http.Request) {
		handler.SendSMSCodeHandler(codeStore).ServeHTTP(w, r)
	})

	mux.HandleFunc("/v1/auth/sms/verify", func(w http.ResponseWriter, r *http.Request) {
		handler.VerifySMSCodeHandler(codeStore, cfg.JWTSecret).ServeHTTP(w, r)
	})

	mux.Handle("/healthz", handler.HealthCheckHandler())

	authMW := middleware.AuthMiddleware(cfg.JWTSecret)

	mux.Handle("/v1/auth/delete-account", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.DeleteAccountHandler(codeStore)))

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

	mux.Handle("/v1/rights/policy-change-notify", middleware.Chain(
		middleware.RecoveryMiddleware(),
	)(handler.PolicyChangeNotifyHandler(rightsRepo, os.Getenv("SERVICE_SECRET"))))

	scheduler := handler.NewAlertScheduler(rightsRepo)
	scheduler.SetPusher(&pusherAdapter{push: notifier.NewPushService()})
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
	)(handler.GeneratePlanHandler(cfg.LLMGatewayURL, cfg.ActuaryEngineURL, planRepo, profileRepo, policyRepo)))

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

	orderRepo, err := repository.NewOrderRepository(database)
	if err != nil {
		log.Fatalf("failed to create order repository: %v", err)
	}

	mux.Handle("/v1/orders", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.CreateOrderHandler(orderRepo, planRepo)))

	mux.Handle("/v1/orders/check-unlock", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(http.HandlerFunc(handler.CheckUnlockHandler(orderRepo))))

	mux.Handle("/v1/orders/", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/pay") {
			handler.PayOrderHandler(orderRepo).ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})))

	thresholdResolver := handler.NewThresholdResolver(policyRepo)
	mux.Handle("/v1/simulator/calculate", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.SimulatorHandler(thresholdResolver, policyRepo)))

	simScenarioRepo, err := repository.NewSimulatorScenarioRepository(database)
	if err != nil {
		log.Fatalf("failed to create simulator scenario repository: %v", err)
	}
	mux.Handle("/v1/simulator/scenarios", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(methodRouter(map[string]http.Handler{
		http.MethodPost: handler.SimulatorScenarioSaveHandler(simScenarioRepo),
		http.MethodGet:  handler.SimulatorScenarioListHandler(simScenarioRepo),
	})))

	mux.Handle("/v1/advisor/ask", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.AdvisorHandler(cfg.LLMGatewayURL, policyRepo)))

	settingsRepo, err := repository.NewSettingsRepository(database)
	if err != nil {
		log.Fatalf("failed to create settings repository: %v", err)
	}
	mux.Handle("/v1/settings", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(methodRouter(map[string]http.Handler{
		http.MethodGet:  handler.GetSettingsHandler(settingsRepo),
		http.MethodPost: handler.SaveSettingsHandler(settingsRepo),
	})))

	mux.Handle("/v1/auth/delete-account-v2", middleware.Chain(
		middleware.RecoveryMiddleware(),
		authMW,
	)(handler.DeleteAccountHandlerV2(settingsRepo)))

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	if envAddr := os.Getenv("LISTEN_ADDR"); envAddr != "" {
		addr = envAddr
	}

	tlsCert := os.Getenv("TLS_CERT")
	tlsKey := os.Getenv("TLS_KEY")
	if tlsCert == "" && os.Getenv("TLS_AUTO") == "true" {
		tlsCert, tlsKey = generateSelfSignedCert()
		log.Println("[main] TLS auto-cert generated (self-signed for development)")
	}

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

type pusherAdapter struct {
	push *notifier.PushService
}

func (p *pusherAdapter) NotifyAlert(payload handler.NotifierPayload) {
	p.push.NotifyAlert(notifier.AlertPayload{
		UserID:  payload.UserID,
		Phone:   payload.Phone,
		Title:   payload.Title,
		Message: payload.Message,
		Type:    payload.Type,
	})
}

func generateSelfSignedCert() (string, string) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"NSI Dev"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}
	derBytes, _ := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyBytes, _ := x509.MarshalECPrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	certFile := "/tmp/nsi-cert.pem"
	keyFile := "/tmp/nsi-key.pem"
	os.WriteFile(certFile, certPEM, 0644)
	os.WriteFile(keyFile, keyPEM, 0644)
	return certFile, keyFile
}


