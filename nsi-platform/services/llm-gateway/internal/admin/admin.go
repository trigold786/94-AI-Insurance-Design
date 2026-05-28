package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/config"
	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/provider"
	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/usage"
)

type Handler struct {
	configStore *config.ConfigStore
	usageStore  *usage.UsageStore
	adminUser   string
	adminPass   string
}

func NewHandler(cs *config.ConfigStore, us *usage.UsageStore, user, pass string) *Handler {
	return &Handler{
		configStore: cs,
		usageStore:  us,
		adminUser:   user,
		adminPass:   pass,
	}
}

func (h *Handler) BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != h.adminUser || pass != h.adminPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="llm-gateway admin"`)
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "unauthorized",
			})
			return
		}
		next(w, r)
	}
}

func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.configStore.ListProviders(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"code":    500,
			"message": fmt.Sprintf("list providers: %v", err),
		})
		return
	}
	for i := range providers {
		providers[i].APIKey = maskKey(providers[i].APIKey)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": providers,
	})
}

func (h *Handler) SaveProvider(w http.ResponseWriter, r *http.Request) {
	var cfg config.ProviderConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "invalid JSON",
		})
		return
	}
	if cfg.ProviderName == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "provider_name is required",
		})
		return
	}
	if err := h.configStore.SaveProvider(r.Context(), &cfg); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"code":    500,
			"message": fmt.Sprintf("save provider: %v", err),
		})
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "saved",
	})
}

func (h *Handler) TestProvider(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProviderName string `json:"provider_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "invalid JSON",
		})
		return
	}
	providers, err := h.configStore.ListProviders(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"code":    500,
			"message": fmt.Sprintf("load providers: %v", err),
		})
		return
	}
	var cfg *config.ProviderConfig
	for i := range providers {
		if providers[i].ProviderName == req.ProviderName {
			cfg = &providers[i]
			break
		}
	}
	if cfg == nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"code":    404,
			"message": "provider not found",
		})
		return
	}

	var p provider.Provider
	switch cfg.ProviderName {
	case "ali_bailian":
		p = &provider.BailianProvider{
			Endpoint:  cfg.Endpoint,
			APIKey:    cfg.APIKey,
			ModelName: cfg.ModelName,
			MaxTokens: cfg.MaxTokens,
		}
	default:
		p = &provider.OpenAICompatProvider{
			Endpoint:  cfg.Endpoint,
			APIKey:    cfg.APIKey,
			ModelName: cfg.ModelName,
			MaxTokens: cfg.MaxTokens,
		}
	}

	start := time.Now()
	result, err := p.Chat("You are a test assistant.", "Say hello in one sentence.")
	latency := time.Since(start)
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "test failed",
			"data": map[string]interface{}{
				"provider_name": cfg.ProviderName,
				"latency_ms":    latency.Milliseconds(),
				"status":        "error",
				"error":         err.Error(),
			},
		})
		return
	}
	preview := result
	if len(preview) > 200 {
		preview = preview[:200]
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "test passed",
		"data": map[string]interface{}{
			"provider_name":   cfg.ProviderName,
			"latency_ms":      latency.Milliseconds(),
			"status":          "ok",
			"response_preview": preview,
		},
	})
}

func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	data, err := h.usageStore.GetDailyUsage(r.Context(), 30)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"code":    500,
			"message": fmt.Sprintf("get usage: %v", err),
		})
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": data,
	})
}

func (h *Handler) AdminPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(adminHTML))
}

func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
