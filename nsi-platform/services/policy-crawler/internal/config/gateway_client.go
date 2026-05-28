package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/llm"
)

type GatewayConfigClient struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.RWMutex
	cache      map[string]*cacheEntry
	ttl        time.Duration
}

type cacheEntry struct {
	data      *modelConfigResponse
	fetchedAt time.Time
}

type modelConfigResponse struct {
	Code int `json:"code"`
	Data struct {
		ID                int             `json:"id"`
		FunctionKey       string          `json:"function_key"`
		Provider          string          `json:"provider"`
		ModelID           string          `json:"model_id"`
		APIEndpoint       string          `json:"api_endpoint"`
		APIKey            string          `json:"api_key"`
		ExtraParams       json.RawMessage `json:"extra_params"`
		MaxTokens         int             `json:"max_tokens"`
		Enabled           bool            `json:"enabled"`
		BackupProvider    string          `json:"backup_provider"`
		BackupModelID     string          `json:"backup_model_id"`
		BackupAPIEndpoint string          `json:"backup_api_endpoint"`
		BackupAPIKey      string          `json:"backup_api_key"`
	} `json:"data"`
}

type EmbeddingConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	Dimensions int
}

type ASRConfig struct {
	Provider            string
	APIKey              string
	Endpoint            string
	AppID               string
	ResourceID          string
	Language            string
	SampleRate          int
	MaxWaitSeconds      int
	PollIntervalSeconds int
	Enabled             bool
}

func NewGatewayConfigClient(baseURL string) *GatewayConfigClient {
	if baseURL == "" {
		baseURL = "http://localhost:39404"
	}
	return &GatewayConfigClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]*cacheEntry),
		ttl:   60 * time.Second,
	}
}

func (g *GatewayConfigClient) fetchConfig(ctx context.Context, functionKey string) (*modelConfigResponse, error) {
	g.mu.RLock()
	if entry, ok := g.cache[functionKey]; ok && time.Since(entry.fetchedAt) < g.ttl {
		g.mu.RUnlock()
		return entry.data, nil
	}
	g.mu.RUnlock()

	url := fmt.Sprintf("%s/internal/model-configs/%s", g.baseURL, functionKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch config %s: %w", functionKey, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result modelConfigResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("gateway error: code=%d", result.Code)
	}

	g.mu.Lock()
	g.cache[functionKey] = &cacheEntry{data: &result, fetchedAt: time.Now()}
	g.mu.Unlock()
	return &result, nil
}

func (g *GatewayConfigClient) GetLLMConfig(ctx context.Context) (llm.Config, *llm.Config, error) {
	resp, err := g.fetchConfig(ctx, "llm_extract")
	if err != nil {
		return llm.Config{}, nil, fmt.Errorf("get llm config: %w", err)
	}

	cfg := llm.Config{
		Provider:  llm.ParseProvider(resp.Data.Provider),
		APIKey:    resp.Data.APIKey,
		Endpoint:  resp.Data.APIEndpoint,
		ModelName: resp.Data.ModelID,
		MaxTokens: resp.Data.MaxTokens,
		Enabled:   resp.Data.Enabled,
	}

	var backup *llm.Config
	if resp.Data.BackupProvider != "" && resp.Data.BackupAPIKey != "" {
		backup = &llm.Config{
			Provider:  llm.ParseProvider(resp.Data.BackupProvider),
			APIKey:    resp.Data.BackupAPIKey,
			Endpoint:  resp.Data.BackupAPIEndpoint,
			ModelName: resp.Data.BackupModelID,
			MaxTokens: resp.Data.MaxTokens,
			Enabled:   true,
		}
	}

	return cfg, backup, nil
}

func (g *GatewayConfigClient) GetEmbeddingConfig(ctx context.Context) (EmbeddingConfig, error) {
	resp, err := g.fetchConfig(ctx, "embedding")
	if err != nil {
		return EmbeddingConfig{}, fmt.Errorf("get embedding config: %w", err)
	}

	ec := EmbeddingConfig{
		APIKey:  resp.Data.APIKey,
		BaseURL: resp.Data.APIEndpoint,
		Model:   resp.Data.ModelID,
	}

	type extraParams struct {
		Dimensions int `json:"dimensions"`
	}
	if resp.Data.ExtraParams != nil {
		var ep extraParams
		if err := json.Unmarshal(resp.Data.ExtraParams, &ep); err == nil && ep.Dimensions > 0 {
			ec.Dimensions = ep.Dimensions
		}
	}
	if ec.Dimensions <= 0 {
		ec.Dimensions = 1024
	}

	return ec, nil
}

func (g *GatewayConfigClient) GetASRConfig(ctx context.Context) (ASRConfig, error) {
	resp, err := g.fetchConfig(ctx, "asr")
	if err != nil {
		return ASRConfig{}, fmt.Errorf("get asr config: %w", err)
	}

	ac := ASRConfig{
		Provider: resp.Data.Provider,
		APIKey:   resp.Data.APIKey,
		Endpoint: resp.Data.APIEndpoint,
		Enabled:  resp.Data.Enabled,
	}

	type asrExtra struct {
		AppID               string `json:"app_id"`
		ResourceID          string `json:"resource_id"`
		Language            string `json:"language"`
		SampleRate          int    `json:"sample_rate"`
		MaxWaitSeconds      int    `json:"max_wait_seconds"`
		PollIntervalSeconds int    `json:"poll_interval_seconds"`
	}
	if resp.Data.ExtraParams != nil {
		var ep asrExtra
		if err := json.Unmarshal(resp.Data.ExtraParams, &ep); err == nil {
			ac.AppID = ep.AppID
			ac.ResourceID = ep.ResourceID
			ac.Language = ep.Language
			ac.SampleRate = ep.SampleRate
			ac.MaxWaitSeconds = ep.MaxWaitSeconds
			ac.PollIntervalSeconds = ep.PollIntervalSeconds
		}
	}

	return ac, nil
}
