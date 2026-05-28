# Crawler Gateway Config Integration Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace crawler's local LLM/ASR/Embedding config reads with calls to llm-gateway's unified model_configs API.

**Architecture:** Add an internal (unauthenticated) endpoint to llm-gateway for service-to-service config reads. Add a `GatewayConfigClient` in the crawler that reads from this endpoint with a 60s TTL cache. Replace all local DB config reads with gateway client calls.

**Tech Stack:** Go, HTTP, Docker networking

---

### Task 1: Add internal model-config endpoint to llm-gateway

**Files:**
- Modify: `services/llm-gateway/cmd/main.go:182` (add new route)
- Modify: `services/llm-gateway/internal/admin/admin.go:257` (add GetModelConfigInternal)

- [ ] **Step 1: Add GetModelConfigInternal handler in admin.go**

Add this method to `admin.go` after the existing `GetModelConfig` method (after line 280):

```go
func (h *Handler) GetModelConfigInternal(w http.ResponseWriter, r *http.Request) {
	functionKey := strings.TrimPrefix(r.URL.Path, "/internal/model-configs/")
	functionKey = strings.TrimSuffix(functionKey, "/")
	if functionKey == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "function_key is required",
		})
		return
	}

	cfg, err := h.mcStore.GetByKey(r.Context(), functionKey)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"code":    404,
			"message": err.Error(),
		})
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": cfg,
	})
}
```

Note: This returns the **unmasked** `cfg` (not `cfg.ToMasked()`) so the crawler gets the real API key.

- [ ] **Step 2: Add route in main.go**

In `cmd/main.go`, add this route after line 182 (after the existing `/admin/model-configs/` route):

```go
mux.HandleFunc("/internal/model-configs/", adminHandler.GetModelConfigInternal)
```

This route has NO `BasicAuth` wrapper — it's unauthenticated for Docker-internal use only.

- [ ] **Step 3: Build and deploy**

Run:
```bash
cd nsi-platform/services/llm-gateway
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:GOPROXY="https://goproxy.cn,direct"
go build -o llm-gateway ./cmd/
docker restart nsi-llm-gateway
```

- [ ] **Step 4: Verify the internal endpoint**

Run:
```bash
curl http://localhost:39404/internal/model-configs/llm_extract
```

Expected: JSON with `"code":0` and `"data"` containing full config with unmasked `api_key`.

- [ ] **Step 5: Commit**

```bash
git add services/llm-gateway/internal/admin/admin.go services/llm-gateway/cmd/main.go
git commit -m "feat(llm-gateway): add internal model-configs endpoint for service-to-service reads"
```

---

### Task 2: Create GatewayConfigClient in policy-crawler

**Files:**
- Create: `services/policy-crawler/internal/config/gateway_client.go`

- [ ] **Step 1: Create the GatewayConfigClient**

Create `services/policy-crawler/internal/config/gateway_client.go`:

```go
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/crawler"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/llm"
)

type GatewayConfigClient struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex
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
	g.mu.Lock()
	defer g.mu.Unlock()

	if entry, ok := g.cache[functionKey]; ok && time.Since(entry.fetchedAt) < g.ttl {
		return entry.data, nil
	}

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

	body, err := io.ReadAll(resp.Body)
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

	g.cache[functionKey] = &cacheEntry{data: &result, fetchedAt: time.Now()}
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

func (g *GatewayConfigClient) GetASRConfig(ctx context.Context) (crawler.ASRConfig, error) {
	resp, err := g.fetchConfig(ctx, "asr")
	if err != nil {
		return crawler.ASRConfig{}, fmt.Errorf("get asr config: %w", err)
	}

	ac := crawler.ASRConfig{
		Provider: resp.Data.Provider,
		APIKey:   resp.Data.APIKey,
		Endpoint: resp.Data.APIEndpoint,
		Enabled:  resp.Data.Enabled,
	}

	type asrExtra struct {
		AppID      string `json:"app_id"`
		ResourceID string `json:"resource_id"`
		Language   string `json:"language"`
	}
	if resp.Data.ExtraParams != nil {
		var ep asrExtra
		if err := json.Unmarshal(resp.Data.ExtraParams, &ep); err == nil {
			ac.AppID = ep.AppID
			ac.ResourceID = ep.ResourceID
			ac.Language = ep.Language
		}
	}

	return ac, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run:
```bash
cd nsi-platform/services/policy-crawler
$env:GOPROXY="https://goproxy.cn,direct"
go build ./internal/config/
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/internal/config/gateway_client.go
git commit -m "feat(crawler): add GatewayConfigClient for llm-gateway config reads"
```

---

### Task 3: Wire GatewayConfigClient into crawler main.go

**Files:**
- Modify: `services/policy-crawler/cmd/main.go`

- [ ] **Step 1: Update imports**

Add import for the new config package. Change the import block to include:
```go
gwconfig "github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/config"
```

- [ ] **Step 2: Replace embedding config initialization (lines 50-79)**

Replace the embedding initialization block (lines 50-79) with:

```go
	var embedProv embeddings.EmbeddingProvider
	gwClient := gwconfig.NewGatewayConfigClient(cfg.LLMGatewayURL)

	embedCfg, embedErr := gwClient.GetEmbeddingConfig(context.Background())
	if embedErr == nil && embedCfg.APIKey != "" {
		embedProv = embeddings.NewProviderFromConfig(embedCfg.APIKey, embedCfg.BaseURL, embedCfg.Model, embedCfg.Dimensions)
		log.Printf("[embeddings] using %s (dims=%d) via %s [from llm-gateway]", embedProv.ModelName(), embedCfg.Dimensions, embedCfg.BaseURL)
	} else {
		if embedErr != nil {
			log.Printf("[embeddings] warning: cannot read from llm-gateway: %v", embedErr)
		}
		embedProv = embeddings.NewProviderFromConfig("", "", "", 1536)
		log.Println("[embeddings] no embedding config, using hash-bow fallback")
	}
```

- [ ] **Step 3: Replace ASR config initialization (lines 106-107)**

Replace lines 106-107:
```go
	asrCfg, _ := store.GetASRConfig()
	manager.InitFilterAndWorker(database, asrCfg)
```

With:
```go
	asrCfg, asrErr := gwClient.GetASRConfig(context.Background())
	if asrErr != nil {
		log.Printf("[asr] warning: cannot read config from llm-gateway: %v", asrErr)
	}
	manager.InitFilterAndWorker(database, asrCfg)
```

- [ ] **Step 4: Build and verify compilation**

Run:
```bash
cd nsi-platform/services/policy-crawler
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:GOPROXY="https://goproxy.cn,direct"
go build ./cmd/
```

Expected: No errors.

- [ ] **Step 5: Commit**

```bash
git add services/policy-crawler/cmd/main.go
git commit -m "feat(crawler): use GatewayConfigClient for embedding and ASR config"
```

---

### Task 4: Replace LLM config reads in extraction handlers

**Files:**
- Modify: `services/policy-crawler/internal/admin/admin_llm.go:157-198` (LLMExtractRunHandler)
- Modify: `services/policy-crawler/internal/crawler/store.go:828-845` (RunExtraction)

- [ ] **Step 1: Update LLMExtractRunHandler to accept GatewayConfigClient**

Change the `LLMExtractRunHandler` signature to accept a `*gwconfig.GatewayConfigClient` parameter:

The handler function signature changes from:
```go
func LLMExtractRunHandler(store interface{}, checker extractor.ReferenceChecker, embedProv embeddings.EmbeddingProvider) http.Handler {
```
To:
```go
func LLMExtractRunHandler(store interface{}, checker extractor.ReferenceChecker, embedProv embeddings.EmbeddingProvider, gwClient *gwconfig.GatewayConfigClient) http.Handler {
```

Add the import for the config package:
```go
gwconfig "github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/config"
```

- [ ] **Step 2: Replace the config read in LLMExtractRunHandler's goroutine**

Replace lines 183-198 (inside the goroutine):
```go
		cfg, err := llmStore.GetLLMConfig()
		if err != nil {
			log.Printf("[extract] get config: %v", err)
			finishExtract("get config error: " + err.Error())
			return
		}

		client := llm.NewClient(llm.Config{
			Provider:  llm.ParseProvider(cfg.Provider),
			APIKey:    cfg.APIKey,
			Endpoint:  cfg.Endpoint,
			ModelName: cfg.ModelName,
			MaxTokens: cfg.MaxTokens,
			Enabled:   cfg.Enabled,
		})
```

With:
```go
		llmCfg, backupCfg, err := gwClient.GetLLMConfig(r.Context())
		if err != nil {
			log.Printf("[extract] get config from gateway: %v", err)
			finishExtract("get config error: " + err.Error())
			return
		}

		var client *llm.Client
		if backupCfg != nil {
			client = llm.NewClientWithBackup(llmCfg, backupCfg)
		} else {
			client = llm.NewClient(llmCfg)
		}
```

Note: `r.Context()` won't work inside a goroutine (context cancelled when request ends). Use `context.Background()` instead:
```go
		llmCfg, backupCfg, err := gwClient.GetLLMConfig(context.Background())
```

Also add `"context"` to imports if not already present.

- [ ] **Step 3: Update RunExtraction in store.go**

The `RunExtraction` method at line 828 currently calls `s.GetLLMConfig()`. This method is used by the `LLMStore` interface. We need to keep it working but have the caller (if any) use the gateway client instead.

Since `RunExtraction` is called via the `LLMStore` interface by `LLMStatusHandler`, and the interface requires `RunExtraction(limit int) (int, int, error)`, we have two options:
1. Add a new method to the store that accepts a gateway client
2. Keep the existing method but make the handler call through the gateway

**Simpler approach:** Keep `RunExtraction` as-is (it still reads from local DB as fallback), but the primary extraction path (`LLMExtractRunHandler`) now uses the gateway client. The `RunExtraction` method is not called from any active path — it's exposed via `LLMStore` interface but only used by the now-secondary `LLMStatusHandler`.

Actually, let's verify — check if `RunExtraction` is called anywhere:

Search for callers of `RunExtraction` — if it's only used via `LLMStore` interface and only called from tests or inactive handlers, we can skip modifying it for now. The primary extraction path is `LLMExtractRunHandler`.

For this step, do NOT modify `store.go:RunExtraction`. The change in Step 2 already covers the primary extraction path.

- [ ] **Step 4: Update the main.go route registration**

In `cmd/main.go`, update line 170 to pass the gateway client:

```go
mux.Handle("/admin/llm/extract", adminAuth(middleware.RecoveryMiddleware()(admin.LLMExtractRunHandler(store, searcher, embedProv, gwClient))))
```

- [ ] **Step 5: Build and verify**

Run:
```bash
cd nsi-platform/services/policy-crawler
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:GOPROXY="https://goproxy.cn,direct"
go build ./cmd/
```

Expected: No errors.

- [ ] **Step 6: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_llm.go services/policy-crawler/cmd/main.go
git commit -m "feat(crawler): use GatewayConfigClient for LLM extraction config"
```

---

### Task 5: Update crawler admin page to link to llm-gateway for model config

**Files:**
- Modify: `services/policy-crawler/internal/admin/admin_page.go` (replace LLM/ASR config panels)
- Modify: `services/policy-crawler/cmd/main.go` (remove or redirect old LLM/ASR config routes)

- [ ] **Step 1: Find the extract panel in admin_page.go**

The crawler admin page uses `switchPanel` to show panels by ID. The `extract` panel loads the LLM config form. We need to replace this form with a link to the llm-gateway admin UI.

Search for the function `loadExtract` or the panel with id `extract` in the JavaScript. The panel likely fetches from `/admin/llm/config` and renders a form with provider, API key, etc.

Find the section that renders the LLM config form (it will have fields like `llmProvider`, `llmAPIKey`, etc.) and replace it with a card that links to the llm-gateway.

The exact replacement depends on the current HTML structure. In the `loadExtract` function, after the progress/status section, replace the LLM config form with:

```html
<div class="card">
  <div class="card-header">模型配置</div>
  <p style="margin:8px 0;color:#6B7280;font-size:13px">
    LLM、Embedding 和 ASR 模型配置已统一迁移到 LLM Gateway 管理后台。
  </p>
  <a href="http://localhost:39404/admin/#model-configs" target="_blank" 
     style="display:inline-block;padding:8px 16px;background:#1A56DB;color:#fff;border-radius:6px;text-decoration:none;font-size:13px">
    前往 LLM Gateway 配置模型 →
  </a>
</div>
```

Similarly for the ASR config section if it exists in the page.

- [ ] **Step 2: Remove old config routes from main.go**

Remove or redirect these routes in `cmd/main.go`:

```go
// These routes are no longer needed - config is managed in llm-gateway
// mux.Handle("/admin/llm/config", adminAuth(admin.LLMConfigGetHandler(store)))
// mux.Handle("/admin/llm/config/save", adminAuth(admin.LLMConfigSaveHandler(store)))
// mux.Handle("/admin/asr/config", adminAuth(admin.ASRConfigGetHandler(database)))
// mux.Handle("/admin/asr/config/save", adminAuth(admin.ASRConfigSaveHandler(database)))
```

Keep:
- `/admin/llm/extract` (triggers extraction)
- `/admin/llm/status` (shows status)
- `/admin/llm/pending` (shows pending count)
- `/admin/llm/progress` (shows progress)
- `/admin/asr/test` (tests ASR — but it needs DB access, may need updating later)

- [ ] **Step 3: Build and verify**

Run:
```bash
cd nsi-platform/services/policy-crawler
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:GOPROXY="https://goproxy.cn,direct"
go build ./cmd/
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_page.go services/policy-crawler/cmd/main.go
git commit -m "feat(crawler): redirect model config to llm-gateway admin UI"
```

---

### Task 6: Build, deploy, and end-to-end test

**Files:** None (deployment only)

- [ ] **Step 1: Build both services**

```bash
cd nsi-platform/services/llm-gateway
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:GOPROXY="https://goproxy.cn,direct"
go build -o llm-gateway ./cmd/

cd ../policy-crawler
go build -o policy-crawler ./cmd/
```

- [ ] **Step 2: Restart containers**

```bash
docker restart nsi-llm-gateway
docker restart nsi-policy-crawler
```

- [ ] **Step 3: Verify llm-gateway internal endpoint**

```bash
curl http://localhost:39404/internal/model-configs/llm_extract
curl http://localhost:39404/internal/model-configs/embedding
curl http://localhost:39404/internal/model-configs/asr
```

Expected: All return `{"code":0,"data":{...}}` with unmasked API keys.

- [ ] **Step 4: Check crawler logs for gateway config reads**

```bash
docker logs nsi-policy-crawler 2>&1 | Select-String "llm-gateway|gateway|embeddings"
```

Expected: Log lines showing config read from llm-gateway, e.g. `[embeddings] using ... [from llm-gateway]`.

- [ ] **Step 5: Verify crawler admin page**

Open `http://localhost:39403/admin` in browser. Navigate to the "AI提取" panel. It should show a link to the llm-gateway admin UI instead of the old config form.

- [ ] **Step 6: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix: deployment fixes for crawler gateway config integration"
```
