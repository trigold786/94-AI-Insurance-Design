package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/admin"
	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/config"
	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/gateway"
	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/modelconfig"
	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/provider"
	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/usage"
	sharedDB "github.com/trigold786/94-AI-Insurance-Design/shared/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL env var is required")
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "39404"
	}
	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		adminPass = "changeme"
	}

	db, err := sharedDB.Connect(databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	configStore, err := config.NewConfigStore(db)
	if err != nil {
		log.Fatalf("create config store: %v", err)
	}

	usageStore, err := usage.NewUsageStore(db)
	if err != nil {
		log.Fatalf("create usage store: %v", err)
	}

	mcStore, err := modelconfig.NewStore(db)
	if err != nil {
		log.Fatalf("create model config store: %v", err)
	}

	ctx := context.Background()
	enabledProviders, err := configStore.GetEnabledProviders(ctx)
	if err != nil {
		log.Fatalf("load enabled providers: %v", err)
	}

	var providers []provider.Provider
	for _, cfg := range enabledProviders {
		switch cfg.ProviderName {
		case "ali_bailian":
			providers = append(providers, &provider.BailianProvider{
				Endpoint:  cfg.Endpoint,
				APIKey:    cfg.APIKey,
				ModelName: cfg.ModelName,
				MaxTokens: cfg.MaxTokens,
			})
		default:
			providers = append(providers, &provider.OpenAICompatProvider{
				Endpoint:  cfg.Endpoint,
				APIKey:    cfg.APIKey,
				ModelName: cfg.ModelName,
				MaxTokens: cfg.MaxTokens,
			})
		}
	}

	gw := gateway.New(providers)
	adminHandler := admin.NewHandler(configStore, usageStore, mcStore, adminUser, adminPass)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "only POST is supported",
			})
			return
		}

		var req struct {
			SystemPrompt string `json:"system_prompt"`
			UserContent  string `json:"user_content"`
			MaxTokens    int    `json:"max_tokens"`
			Caller       string `json:"caller"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    "BAD_REQUEST",
				"message": "invalid JSON body",
			})
			return
		}

		start := time.Now()
		content, providerUsed, err := gw.Chat(req.SystemPrompt, req.UserContent)
		latencyMs := time.Since(start).Milliseconds()

		usageLog := &usage.UsageLog{
			ProviderName: providerUsed,
			Caller:       req.Caller,
			LatencyMs:    int(latencyMs),
		}

		if err != nil {
			usageLog.Status = "error"
			usageLog.ErrorMessage = err.Error()
			_ = usageStore.Record(r.Context(), usageLog)
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code":    "LLM_ERROR",
				"message": err.Error(),
			})
			return
		}

		usageLog.Status = "ok"
		usageLog.Model = providerUsed
		_ = usageStore.Record(r.Context(), usageLog)

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":          0,
			"content":       content,
			"provider_used": providerUsed,
			"latency_ms":    latencyMs,
		})
	})

	mux.HandleFunc("/admin/providers", adminHandler.BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminHandler.ListProviders(w, r)
		case http.MethodPost:
			adminHandler.SaveProvider(w, r)
		default:
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "only GET and POST are supported",
			})
		}
	}))

	mux.HandleFunc("/admin/providers/test", adminHandler.BasicAuth(adminHandler.TestProvider))
	mux.HandleFunc("/admin/usage", adminHandler.BasicAuth(adminHandler.GetUsage))

	mux.HandleFunc("/admin/model-configs", adminHandler.BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminHandler.ListModelConfigs(w, r)
		default:
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "only GET is supported",
			})
		}
	}))
	mux.HandleFunc("/admin/model-configs/save", adminHandler.BasicAuth(adminHandler.SaveModelConfig))
	mux.HandleFunc("/admin/model-configs/test", adminHandler.BasicAuth(adminHandler.TestModelConfig))
	mux.HandleFunc("/admin/model-configs/", adminHandler.BasicAuth(adminHandler.GetModelConfig))
	mux.HandleFunc("/internal/model-configs/", adminHandler.GetModelConfigInternal)

	mux.HandleFunc("/admin/", adminHandler.BasicAuth(adminHandler.AdminPage))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 180 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("llm-gateway starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
	log.Println("server stopped")
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
