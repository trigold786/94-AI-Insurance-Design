# LLM 方案生成系统实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新建 llm-gateway 独立微服务，改造 api-server 方案生成从精算模型切换为 LLM 智能生成，保留精算引擎作双向验证。

**Architecture:** 新增 `llm-gateway` 服务（:39404）统一管理 LLM 调用（多 provider + fallback）。api-server 的 `GeneratePlanHandler` 改为：检索三级政策 → 组装 prompt → 调 llm-gateway → 解析双视图输出 → 精算引擎验证偏差 → 返回方案。前端增加视图切换和政策依据展示。

**Tech Stack:** Go 1.24+, PostgreSQL, Docker, Chart.js, LLM APIs (DeepSeek/阿里云百炼/火山方舟/OpenCode Go)

**Spec:** `docs/superpowers/specs/2026-05-28-llm-plan-generation-design.md`

---

## File Structure

### 新建文件

| File | Responsibility |
|------|---------------|
| `services/llm-gateway/cmd/main.go` | 入口，HTTP server + 路由 |
| `services/llm-gateway/go.mod` | Go module 定义 |
| `services/llm-gateway/internal/gateway/gateway.go` | 核心网关：provider 选择、fallback、限流 |
| `services/llm-gateway/internal/provider/provider.go` | Provider 接口 + OpenAI 兼容实现 |
| `services/llm-gateway/internal/provider/bailian.go` | 阿里云百炼实现 |
| `services/llm-gateway/internal/config/config.go` | DB-backed provider 配置 CRUD |
| `services/llm-gateway/internal/usage/usage.go` | 用量日志记录 |
| `services/llm-gateway/internal/admin/admin.go` | 管理 API handler |
| `services/llm-gateway/internal/admin/admin_page.go` | 管理页面 HTML |
| `services/llm-gateway/migrations/001_init.sql` | llm_providers + llm_usage_logs 建表 |
| `services/llm-gateway/Dockerfile` | Docker 镜像 |

### 修改文件

| File | Change |
|------|--------|
| `shared/models/models.go` | 新增 LLMScheme, PolicyReference, VerificationResult, DeviationDetail；扩展 PlanSnapshot |
| `shared/config/config.go` | 新增 LLMGatewayURL 字段 |
| `services/api-server/internal/handler/plan_handler.go` | 重写：检索三级政策 + 调 llm-gateway + 解析双视图 |
| `services/api-server/internal/handler/webclient_handler.go` | 改造方案展示：双视图 + 政策依据 |
| `services/api-server/internal/repository/policy_repo.go` | 新增 QueryByRegionHierarchy + policy source fields 查询 |
| `services/api-server/internal/repository/plan_repo.go` | 适配新 PlanSnapshot 结构 |
| `services/api-server/cmd/main.go` | 新增 LLMGatewayURL 配置，传给 handler |
| `services/api-server/migrations/008_plan_verification.sql` | plan_verification_logs 建表 + plan_snapshots 新列 |
| `docker-compose.yml` | 新增 llm-gateway 服务 + api-server 环境变量 |
| `scripts/migrate.sh` | 新增 llm-gateway migrations |
| `Makefile` | 新增 build-llm-gateway, test-llm-gateway |
| `services/api-server/go.mod` | 无需改动（调 llm-gateway 用 HTTP） |

---

## Task 1: llm-gateway 项目脚手架 + go.mod

**Files:**
- Create: `services/llm-gateway/go.mod`
- Create: `services/llm-gateway/cmd/main.go` (minimal healthz only)
- Create: `services/llm-gateway/Dockerfile`

- [ ] **Step 1: 创建目录结构和 go.mod**

```sh
mkdir -p services/llm-gateway/cmd
mkdir -p services/llm-gateway/internal/gateway
mkdir -p services/llm-gateway/internal/provider
mkdir -p services/llm-gateway/internal/config
mkdir -p services/llm-gateway/internal/usage
mkdir -p services/llm-gateway/internal/admin
mkdir -p services/llm-gateway/migrations
```

创建 `services/llm-gateway/go.mod`:
```
module github.com/trigold786/94-AI-Insurance-Design/llm-gateway

go 1.24

require (
	github.com/lib/pq v1.10.9
)

replace github.com/trigold786/94-AI-Insurance-Design/shared => ../../shared
```

- [ ] **Step 2: 创建最小化 main.go（仅 healthz）**

`services/llm-gateway/cmd/main.go`:
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "39404"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
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
}

```

修复 import（需要加 `"time"`）。

- [ ] **Step 3: 创建 Dockerfile**

`services/llm-gateway/Dockerfile`:
```dockerfile
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY bin/llm-gateway /llm-gateway
RUN adduser -D -g '' appuser
USER appuser
EXPOSE 39404
ENTRYPOINT ["/llm-gateway"]
```

- [ ] **Step 4: 编译验证**

```sh
cd services/llm-gateway
GOPROXY=https://goproxy.cn,direct go mod tidy
GOOS=linux GOARCH=amd64 go build -o bin/llm-gateway ./cmd/main.go
```

Expected: 编译成功，无错误。

- [ ] **Step 5: Commit**

```bash
git add services/llm-gateway/
git commit -m "feat(llm-gateway): scaffold project with healthz endpoint"
```

---

## Task 2: llm-gateway Provider 接口 + OpenAI 兼容实现

**Files:**
- Create: `services/llm-gateway/internal/provider/provider.go`
- Create: `services/llm-gateway/internal/provider/provider_test.go`

- [ ] **Step 1: 编写 Provider 接口测试**

`services/llm-gateway/internal/provider/provider_test.go`:
```go
package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAICompatChat_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}
		resp := ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "hello world"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := &OpenAICompatProvider{
		Endpoint:  srv.URL,
		APIKey:    "test-key",
		ModelName: "test-model",
		MaxTokens: 100,
		HTTPClient: srv.Client(),
	}

	content, err := p.Chat("system prompt", "user content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", content)
	}
}

func TestOpenAICompatChat_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer srv.Close()

	p := &OpenAICompatProvider{
		Endpoint:   srv.URL,
		APIKey:     "test-key",
		ModelName:  "test-model",
		MaxTokens:  100,
		HTTPClient: srv.Client(),
	}

	_, err := p.Chat("sys", "user")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestOpenAICompatChat_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	p := &OpenAICompatProvider{
		Endpoint:   srv.URL,
		APIKey:     "test-key",
		ModelName:  "test-model",
		MaxTokens:  100,
		HTTPClient: srv.Client(),
	}

	_, err := p.Chat("sys", "user")
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```sh
cd services/llm-gateway
go test ./internal/provider/ -v -run TestOpenAICompat
```

Expected: 编译失败（类型未定义）。

- [ ] **Step 3: 实现 Provider 接口 + OpenAI 兼容**

`services/llm-gateway/internal/provider/provider.go`:
```go
package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Provider interface {
	Chat(systemPrompt, userContent string) (string, error)
	Name() string
	Model() string
}

type OpenAICompatProvider struct {
	Endpoint   string
	APIKey     string
	ModelName  string
	MaxTokens  int
	HTTPClient *http.Client
}

func (p *OpenAICompatProvider) Name() string  { return p.ModelName }
func (p *OpenAICompatProvider) Model() string { return p.ModelName }

func (p *OpenAICompatProvider) Chat(systemPrompt, userContent string) (string, error) {
	client := p.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}

	type chatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type chatRequest struct {
		Model     string        `json:"model"`
		Messages  []chatMessage `json:"messages"`
		MaxTokens int           `json:"max_tokens,omitempty"`
	}

	req := chatRequest{
		Model: p.ModelName,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		MaxTokens: p.MaxTokens,
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", p.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API %d: %s", resp.StatusCode, string(respBody)[:min(200, len(string(respBody)))])
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty response choices")
	}
	return chatResp.Choices[0].Message.Content, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

- [ ] **Step 4: 运行测试确认通过**

```sh
cd services/llm-gateway
go test ./internal/provider/ -v -run TestOpenAICompat
```

Expected: 3 tests PASS。

- [ ] **Step 5: Commit**

```bash
git add services/llm-gateway/internal/provider/
git commit -m "feat(llm-gateway): add Provider interface + OpenAI compat implementation"
```

---

## Task 3: 阿里云百炼 Provider

**Files:**
- Modify: `services/llm-gateway/internal/provider/provider.go` (无需改动，已在此文件加类型)
- Create: `services/llm-gateway/internal/provider/bailian.go`
- Modify: `services/llm-gateway/internal/provider/provider_test.go`

- [ ] **Step 1: 编写百炼 provider 测试**

在 `provider_test.go` 末尾追加：
```go
func TestBailianChat_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"output":{"text":"百炼回复"}}`))
	}))
	defer srv.Close()

	p := &BailianProvider{
		Endpoint:   srv.URL,
		APIKey:     "test-key",
		ModelName:  "qwen-plus",
		MaxTokens:  100,
		HTTPClient: srv.Client(),
	}

	content, err := p.Chat("sys", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "百炼回复" {
		t.Errorf("expected '百炼回复', got '%s'", content)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```sh
cd services/llm-gateway
go test ./internal/provider/ -v -run TestBailian
```

Expected: 编译失败（BailianProvider 未定义）。

- [ ] **Step 3: 实现 BailianProvider**

`services/llm-gateway/internal/provider/bailian.go`:
```go
package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BailianProvider struct {
	Endpoint   string
	APIKey     string
	ModelName  string
	MaxTokens  int
	HTTPClient *http.Client
}

func (p *BailianProvider) Name() string  { return "ali_bailian" }
func (p *BailianProvider) Model() string { return p.ModelName }

type bailianRequest struct {
	Model string `json:"model"`
	Input struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	} `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type bailianResponse struct {
	Output struct {
		Text string `json:"text"`
	} `json:"output"`
}

func (p *BailianProvider) Chat(systemPrompt, userContent string) (string, error) {
	client := p.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}

	req := bailianRequest{
		Model: p.ModelName,
		Parameters: map[string]interface{}{
			"max_tokens":    p.MaxTokens,
			"result_format": "text",
		},
	}
	req.Input.Messages = []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent},
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", p.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bailian API %d: %s", resp.StatusCode, string(respBody)[:min(200, len(string(respBody)))])
	}

	var bResp bailianResponse
	if err := json.Unmarshal(respBody, &bResp); err != nil {
		return "", fmt.Errorf("parse bailian response: %w", err)
	}
	return bResp.Output.Text, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

```sh
cd services/llm-gateway
go test ./internal/provider/ -v
```

Expected: 4 tests PASS。

- [ ] **Step 5: Commit**

```bash
git add services/llm-gateway/internal/provider/
git commit -m "feat(llm-gateway): add Bailian (通义千问) provider"
```

---

## Task 4: DB 配置存储 (config)

**Files:**
- Create: `services/llm-gateway/internal/config/config.go`

- [ ] **Step 1: 实现 DB-backed 配置 CRUD**

`services/llm-gateway/internal/config/config.go`:
```go
package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ProviderConfig struct {
	ID           int       `json:"id"`
	ProviderName string   `json:"provider_name"`
	APIKey       string   `json:"api_key"`
	Endpoint     string   `json:"endpoint"`
	ModelName    string    `json:"model_name"`
	MaxTokens    int       `json:"max_tokens"`
	IsPrimary    bool      `json:"is_primary"`
	IsEnabled    bool      `json:"is_enabled"`
	Priority     int       `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ConfigStore struct {
	db *sql.DB
}

func NewConfigStore(db *sql.DB) (*ConfigStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &ConfigStore{db: db}, nil
}

func (s *ConfigStore) ListProviders(ctx context.Context) ([]ProviderConfig, error) {
	query := `SELECT id, provider_name, api_key, endpoint, model_name, max_tokens,
		is_primary, is_enabled, priority, created_at, updated_at
		FROM llm_providers ORDER BY priority ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query providers: %w", err)
	}
	defer rows.Close()

	var providers []ProviderConfig
	for rows.Next() {
		var p ProviderConfig
		if err := rows.Scan(&p.ID, &p.ProviderName, &p.APIKey, &p.Endpoint,
			&p.ModelName, &p.MaxTokens, &p.IsPrimary, &p.IsEnabled,
			&p.Priority, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		providers = append(providers, p)
	}
	if providers == nil {
		providers = []ProviderConfig{}
	}
	return providers, nil
}

func (s *ConfigStore) GetEnabledProviders(ctx context.Context) ([]ProviderConfig, error) {
	query := `SELECT id, provider_name, api_key, endpoint, model_name, max_tokens,
		is_primary, is_enabled, priority, created_at, updated_at
		FROM llm_providers WHERE is_enabled = true ORDER BY priority ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query enabled providers: %w", err)
	}
	defer rows.Close()

	var providers []ProviderConfig
	for rows.Next() {
		var p ProviderConfig
		if err := rows.Scan(&p.ID, &p.ProviderName, &p.APIKey, &p.Endpoint,
			&p.ModelName, &p.MaxTokens, &p.IsPrimary, &p.IsEnabled,
			&p.Priority, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		providers = append(providers, p)
	}
	if providers == nil {
		providers = []ProviderConfig{}
	}
	return providers, nil
}

func (s *ConfigStore) SaveProvider(ctx context.Context, p *ProviderConfig) error {
	query := `INSERT INTO llm_providers (provider_name, api_key, endpoint, model_name, max_tokens, is_primary, is_enabled, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (provider_name) DO UPDATE SET
			api_key = EXCLUDED.api_key,
			endpoint = EXCLUDED.endpoint,
			model_name = EXCLUDED.model_name,
			max_tokens = EXCLUDED.max_tokens,
			is_primary = EXCLUDED.is_primary,
			is_enabled = EXCLUDED.is_enabled,
			priority = EXCLUDED.priority,
			updated_at = NOW()`

	_, err := s.db.ExecContext(ctx, query,
		p.ProviderName, p.APIKey, p.Endpoint, p.ModelName,
		p.MaxTokens, p.IsPrimary, p.IsEnabled, p.Priority)
	if err != nil {
		return fmt.Errorf("save provider: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add services/llm-gateway/internal/config/
git commit -m "feat(llm-gateway): add DB-backed provider config store"
```

---

## Task 5: 用量日志 (usage)

**Files:**
- Create: `services/llm-gateway/internal/usage/usage.go`

- [ ] **Step 1: 实现用量记录**

`services/llm-gateway/internal/usage/usage.go`:
```go
package usage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type UsageLog struct {
	ID           int64     `json:"id"`
	ProviderName string    `json:"provider_name"`
	Model        string    `json:"model"`
	Caller       string    `json:"caller"`
	TokensIn     int       `json:"tokens_in"`
	TokensOut    int       `json:"tokens_out"`
	LatencyMs    int       `json:"latency_ms"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type UsageStore struct {
	db *sql.DB
}

func NewUsageStore(db *sql.DB) (*UsageStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &UsageStore{db: db}, nil
}

func (s *UsageStore) Record(ctx context.Context, log *UsageLog) error {
	query := `INSERT INTO llm_usage_logs (provider_name, model, caller, tokens_in, tokens_out, latency_ms, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.db.ExecContext(ctx, query,
		log.ProviderName, log.Model, log.Caller,
		log.TokensIn, log.TokensOut, log.LatencyMs,
		log.Status, log.ErrorMessage)
	if err != nil {
		return fmt.Errorf("record usage: %w", err)
	}
	return nil
}

type DailyUsage struct {
	Date         string  `json:"date"`
	ProviderName string  `json:"provider_name"`
	TotalCalls   int     `json:"total_calls"`
	TotalTokensIn  int   `json:"total_tokens_in"`
	TotalTokensOut int   `json:"total_tokens_out"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
}

func (s *UsageStore) GetDailyUsage(ctx context.Context, days int) ([]DailyUsage, error) {
	query := `SELECT DATE(created_at) as date, provider_name,
		COUNT(*) as total_calls,
		COALESCE(SUM(tokens_in), 0) as total_tokens_in,
		COALESCE(SUM(tokens_out), 0) as total_tokens_out,
		COALESCE(AVG(latency_ms), 0) as avg_latency_ms
		FROM llm_usage_logs
		WHERE created_at >= NOW() - ($1 || ' days')::INTERVAL
		GROUP BY DATE(created_at), provider_name
		ORDER BY date DESC, provider_name`

	rows, err := s.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("query daily usage: %w", err)
	}
	defer rows.Close()

	var result []DailyUsage
	for rows.Next() {
		var u DailyUsage
		if err := rows.Scan(&u.Date, &u.ProviderName, &u.TotalCalls,
			&u.TotalTokensIn, &u.TotalTokensOut, &u.AvgLatencyMs); err != nil {
			return nil, fmt.Errorf("scan daily usage: %w", err)
		}
		result = append(result, u)
	}
	if result == nil {
		result = []DailyUsage{}
	}
	return result, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add services/llm-gateway/internal/usage/
git commit -m "feat(llm-gateway): add usage logging store"
```

---

## Task 6: Gateway 核心（fallback + 路由）

**Files:**
- Create: `services/llm-gateway/internal/gateway/gateway.go`
- Create: `services/llm-gateway/internal/gateway/gateway_test.go`

- [ ] **Step 1: 编写 Gateway 测试**

`services/llm-gateway/internal/gateway/gateway_test.go`:
```go
package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/config"
	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/provider"
)

func TestGateway_ChatWithFallback(t *testing.T) {
	callCount := 0
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv1.Close()

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := provider.ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: "fallback ok"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv2.Close()

	providers := []provider.Provider{
		&provider.OpenAICompatProvider{
			Endpoint: srv1.URL, APIKey: "k1", ModelName: "model1", MaxTokens: 100, HTTPClient: srv1.Client(),
		},
		&provider.OpenAICompatProvider{
			Endpoint: srv2.URL, APIKey: "k2", ModelName: "model2", MaxTokens: 100, HTTPClient: srv2.Client(),
		},
	}

	gw := &Gateway{providers: providers}
	content, provName, err := gw.Chat("sys", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "fallback ok" {
		t.Errorf("expected 'fallback ok', got '%s'", content)
	}
	if provName != "model2" {
		t.Errorf("expected model2, got '%s'", provName)
	}
	if callCount != 1 {
		t.Errorf("expected primary to be called once, got %d", callCount)
	}
}

func TestGateway_AllProvidersFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	providers := []provider.Provider{
		&provider.OpenAICompatProvider{
			Endpoint: srv.URL, APIKey: "k", ModelName: "m", MaxTokens: 100, HTTPClient: srv.Client(),
		},
	}

	gw := &Gateway{providers: providers}
	_, _, err := gw.Chat("sys", "user")
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```sh
cd services/llm-gateway
go test ./internal/gateway/ -v
```

Expected: 编译失败（Gateway 未定义）。

- [ ] **Step 3: 实现 Gateway**

`services/llm-gateway/internal/gateway/gateway.go`:
```go
package gateway

import (
	"fmt"
	"log"
	"sync"

	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/provider"
)

type Gateway struct {
	mu         sync.RWMutex
	providers  []provider.Provider
}

func New(providers []provider.Provider) *Gateway {
	return &Gateway{providers: providers}
}

func (g *Gateway) UpdateProviders(providers []provider.Provider) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.providers = providers
}

func (g *Gateway) GetProviders() []provider.Provider {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.providers
}

func (g *Gateway) Chat(systemPrompt, userContent string) (content string, providerUsed string, err error) {
	g.mu.RLock()
	providers := make([]provider.Provider, len(g.providers))
	copy(providers, g.providers)
	g.mu.RUnlock()

	if len(providers) == 0 {
		return "", "", fmt.Errorf("no providers configured")
	}

	var lastErr error
	for _, p := range providers {
		result, err := p.Chat(systemPrompt, userContent)
		if err != nil {
			log.Printf("[gateway] provider %s failed: %v", p.Name(), err)
			lastErr = err
			continue
		}
		return result, p.Name(), nil
	}

	return "", "", fmt.Errorf("all providers failed, last error: %v", lastErr)
}
```

- [ ] **Step 4: 运行测试确认通过**

```sh
cd services/llm-gateway
go test ./internal/gateway/ -v
```

Expected: 2 tests PASS。

- [ ] **Step 5: Commit**

```bash
git add services/llm-gateway/internal/gateway/
git commit -m "feat(llm-gateway): add gateway core with fallback routing"
```

---

## Task 7: llm-gateway Migration + DB 初始化

**Files:**
- Create: `services/llm-gateway/migrations/001_init.sql`
- Modify: `scripts/migrate.sh` — 新增 llm-gateway migrations
- Modify: `docker-compose.yml` — db-init 新增 nsi_llm 库
- Modify: `docker-compose.yml` — db-migrate 新增 llm-gateway migrations 挂载

- [ ] **Step 1: 创建 migration 文件**

`services/llm-gateway/migrations/001_init.sql`:
```sql
BEGIN;

CREATE TABLE IF NOT EXISTS llm_providers (
  id SERIAL PRIMARY KEY,
  provider_name VARCHAR(50) NOT NULL UNIQUE,
  api_key TEXT NOT NULL,
  endpoint TEXT NOT NULL,
  model_name VARCHAR(100) NOT NULL,
  max_tokens INT DEFAULT 4096,
  is_primary BOOLEAN DEFAULT false,
  is_enabled BOOLEAN DEFAULT true,
  priority INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS llm_usage_logs (
  id BIGSERIAL PRIMARY KEY,
  provider_name VARCHAR(50),
  model VARCHAR(100),
  caller VARCHAR(50),
  tokens_in INT,
  tokens_out INT,
  latency_ms INT,
  status VARCHAR(20),
  error_message TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_logs_created_at ON llm_usage_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_usage_logs_provider ON llm_usage_logs(provider_name);

COMMIT;
```

- [ ] **Step 2: 修改 migrate.sh**

在 `scripts/migrate.sh` 的 `echo 'Migrations complete.'` 行之前追加：

```sh
echo 'Running llm-gateway migrations...'
run_migrations nsi_llm /migrations/llm
```

- [ ] **Step 3: 修改 docker-compose.yml — db-init 新增 nsi_llm**

在 `db-init` 的 `entrypoint` 的 `psql` 命令行中，在 `CREATE DATABASE nsi_crawler;` 之后追加：
```
psql -h postgres -U postgres -c 'CREATE DATABASE nsi_llm;' 2>/dev/null || true;
```

- [ ] **Step 4: 修改 docker-compose.yml — db-migrate 新增挂载**

在 `db-migrate` 的 `volumes` 中追加：
```yaml
      - ./services/llm-gateway/migrations:/migrations/llm:ro
```

- [ ] **Step 5: Commit**

```bash
git add services/llm-gateway/migrations/ scripts/migrate.sh docker-compose.yml
git commit -m "feat(llm-gateway): add DB migration and docker integration"
```

---

## Task 8: llm-gateway 管理 API + 管理页面

**Files:**
- Create: `services/llm-gateway/internal/admin/admin.go`
- Create: `services/llm-gateway/internal/admin/admin_page.go`

- [ ] **Step 1: 实现管理 API handler**

`services/llm-gateway/internal/admin/admin.go`:
```go
package admin

import (
	"encoding/json"
	"fmt"
	"log"
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
		u, p, ok := r.BasicAuth()
		if !ok || u != h.adminUser || p != h.adminPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="llm-gateway admin"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.configStore.ListProviders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": providers})
}

func (h *Handler) SaveProvider(w http.ResponseWriter, r *http.Request) {
	var p config.ProviderConfig
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid JSON", 400)
		return
	}
	if p.ProviderName == "" {
		http.Error(w, "provider_name required", 400)
		return
	}
	if err := h.configStore.SaveProvider(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
}

func (h *Handler) TestProvider(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProviderName string `json:"provider_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", 400)
		return
	}

	providers, err := h.configStore.ListProviders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var target *config.ProviderConfig
	for i := range providers {
		if providers[i].ProviderName == input.ProviderName {
			target = &providers[i]
			break
		}
	}
	if target == nil {
		http.Error(w, "provider not found", 404)
		return
	}

	var p provider.Provider
	switch target.ProviderName {
	case "ali_bailian":
		p = &provider.BailianProvider{
			Endpoint: target.Endpoint, APIKey: target.APIKey,
			ModelName: target.ModelName, MaxTokens: target.MaxTokens,
		}
	default:
		p = &provider.OpenAICompatProvider{
			Endpoint: target.Endpoint, APIKey: target.APIKey,
			ModelName: target.ModelName, MaxTokens: target.MaxTokens,
		}
	}

	start := time.Now()
	content, err := p.Chat("你是测试助手，请回复：连接成功", "测试连接")
	latency := time.Since(start).Milliseconds()

	result := map[string]interface{}{
		"provider":   target.ProviderName,
		"latency_ms": latency,
	}
	if err != nil {
		result["status"] = "failed"
		result["error"] = err.Error()
		log.Printf("[admin] test provider %s failed: %v", target.ProviderName, err)
	} else {
		result["status"] = "ok"
		result["response_preview"] = content
		if len(content) > 200 {
			result["response_preview"] = content[:200]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": result})
}

func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	daily, err := h.usageStore.GetDailyUsage(r.Context(), 30)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": daily})
}

func (h *Handler) AdminPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(adminHTML))
}
```

- [ ] **Step 2: 实现管理页面 HTML**

`services/llm-gateway/internal/admin/admin_page.go`:
（内嵌 HTML，参考 crawler 的 admin_page.go 模式。包含 3 个 tab：Provider 配置、用量统计、连通性测试。使用 Chart.js 展示用量。页面较长，此处给出核心结构，具体 HTML 实现参照 `services/policy-crawler/internal/admin/admin_page.go` 的模式，用 Go const string 内嵌。）

页面核心功能：
- Provider 列表：显示已配置的 provider，可编辑 API Key/Endpoint/Model/优先级/启用状态
- 保存按钮：POST /admin/providers
- 测试按钮：POST /admin/providers/test
- 用量图表：GET /admin/usage → Chart.js 柱状图
- 默认 provider 预填：DeepSeek, 阿里云百炼, 火山方舟, OpenCode Go

- [ ] **Step 3: Commit**

```bash
git add services/llm-gateway/internal/admin/
git commit -m "feat(llm-gateway): add admin API and management page"
```

---

## Task 9: llm-gateway main.go 完整化 + 公共 Chat API

**Files:**
- Modify: `services/llm-gateway/cmd/main.go`

- [ ] **Step 1: 完善 main.go — 连接 DB + 注册所有路由**

重写 `services/llm-gateway/cmd/main.go`，完整版本包含：
- DB 连接（`shared/db.Connect`）
- 初始化 ConfigStore, UsageStore
- 从 DB 加载 providers 构建 Gateway
- 公共端点 `POST /v1/chat`：接收请求 → Gateway.Chat → 记录 usage → 返回响应
- 管理端点：Basic Auth 保护
- 请求格式：`{"system_prompt":"...","user_content":"...","max_tokens":4096,"caller":"api-server"}`
- 响应格式：`{"content":"...","provider_used":"deepseek","model":"deepseek-v4-flash","latency_ms":1200}`

关键逻辑：
```go
func chatHandler(gw *gateway.Gateway, us *usage.UsageStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SystemPrompt string `json:"system_prompt"`
			UserContent  string `json:"user_content"`
			MaxTokens    int    `json:"max_tokens"`
			Caller       string `json:"caller"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "BAD_REQUEST"})
			return
		}
		if req.SystemPrompt == "" || req.UserContent == "" {
			respondJSON(w, 400, map[string]interface{}{"code": "BAD_REQUEST", "message": "system_prompt and user_content required"})
			return
		}

		start := time.Now()
		content, providerUsed, err := gw.Chat(req.SystemPrompt, req.UserContent)
		latency := time.Since(start).Milliseconds()

		status := "success"
		errMsg := ""
		if err != nil {
			status = "failed"
			errMsg = err.Error()
			us.Record(r.Context(), &usage.UsageLog{ProviderName: "none", Caller: req.Caller, LatencyMs: int(latency), Status: status, ErrorMessage: errMsg})
			respondJSON(w, 502, map[string]interface{}{"code": "LLM_ERROR", "message": errMsg})
			return
		}

		us.Record(r.Context(), &usage.UsageLog{
			ProviderName: providerUsed, Caller: req.Caller,
			LatencyMs: int(latency), Status: status,
		})

		respondJSON(w, 200, map[string]interface{}{
			"code":          0,
			"content":       content,
			"provider_used": providerUsed,
			"latency_ms":    latency,
		})
	}
}
```

注意：需要引入 `shared/db`，go.mod 中已有 replace 指令。需要在 go.mod require 中加入 `github.com/trigold786/94-AI-Insurance-Design/shared v0.0.0`。

- [ ] **Step 2: 编译验证**

```sh
cd services/llm-gateway
GOPROXY=https://goproxy.cn,direct go mod tidy
GOOS=linux GOARCH=amd64 go build -o bin/llm-gateway ./cmd/main.go
```

Expected: 编译成功。

- [ ] **Step 3: Commit**

```bash
git add services/llm-gateway/
git commit -m "feat(llm-gateway): complete main.go with chat API and admin routes"
```

---

## Task 10: shared/models 更新 — 新增 LLM 方案类型

**Files:**
- Modify: `shared/models/models.go`

- [ ] **Step 1: 在 models.go 末尾追加新类型**

在 `shared/models/models.go` 的 `VersionSnapshot` 结构体之后追加：

```go
type LLMScheme struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	MonthlyCost         float64  `json:"monthly_cost"`
	AnnualSubsidy       float64  `json:"annual_subsidy"`
	ProjectedPension    float64  `json:"projected_pension"`
	TotalCost           float64  `json:"total_cost"`
	ContributionBase    float64  `json:"contribution_base"`
	PensionEmployeeRate float64  `json:"pension_employee_rate"`
	PensionEmployerRate float64  `json:"pension_employer_rate"`
	MedicalEmployeeRate float64  `json:"medical_employee_rate"`
	Analysis            string   `json:"analysis"`
	ApplicablePolicies  []string `json:"applicable_policies"`
}

type PolicyReference struct {
	ClaimID         string `json:"claim_id"`
	PolicyTitle     string `json:"policy_title"`
	DocumentNumber  string `json:"document_number"`
	PolicyURL       string `json:"policy_url"`
	RelevantExcerpt string `json:"relevant_excerpt"`
	HowApplied      string `json:"how_applied"`
}

type DeviationDetail struct {
	Metric       string  `json:"metric"`
	LLMValue     float64 `json:"llm_value"`
	ActuaryValue float64 `json:"actuary_value"`
	DeviationPct float64 `json:"deviation_pct"`
}

type VerificationResult struct {
	Status       string            `json:"status"`
	MaxDeviation float64           `json:"max_deviation_pct"`
	Details      []DeviationDetail `json:"details"`
}

type LLMSchemeResponse struct {
	Summary          string            `json:"summary"`
	Schemes          []LLMScheme       `json:"schemes"`
	PolicyReferences []PolicyReference `json:"policy_references"`
	Recommendation   struct {
		RecommendedScheme string `json:"recommended_scheme"`
		Reasoning         string `json:"reasoning"`
	} `json:"recommendation"`
}
```

- [ ] **Step 2: 更新 PlanSnapshot 结构体**

将现有的 `PlanSnapshot` 替换为：

```go
type PlanSnapshot struct {
	PlanID               string             `db:"plan_id" json:"plan_id"`
	UserID               string             `db:"user_id" json:"user_id"`
	PolicyVersionSnapshotID string          `db:"policy_version_snapshot_id" json:"policy_version_snapshot_id"`
	RecommendedSchemes   []Scheme           `db:"recommended_schemes" json:"recommended_schemes"`
	FreeFormText         string             `db:"free_form_text" json:"free_form_text"`
	StructuredSchemes    []LLMScheme        `db:"structured_schemes" json:"structured_schemes"`
	PolicyReferences     []PolicyReference  `db:"policy_references" json:"policy_references"`
	Recommendation       string             `db:"recommendation" json:"recommendation"`
	RecommendationReason string             `db:"recommendation_reason" json:"recommendation_reason"`
	VerificationResult   *VerificationResult `db:"verification_result" json:"verification_result,omitempty"`
	TotalCost            float64            `db:"total_cost" json:"total_cost"`
	TotalSubsidy         float64            `db:"total_subsidy" json:"total_subsidy"`
	GeneratedAt          time.Time          `db:"generated_at" json:"generated_at"`
}
```

注意：保留 `RecommendedSchemes []Scheme`（旧格式）以兼容现有方案数据和报告功能。新增字段都有 `db` tag 以支持后续 DB 存储。

- [ ] **Step 3: 编译验证 shared**

```sh
cd shared && go build ./...
```

Expected: 编译成功。

- [ ] **Step 4: Commit**

```bash
git add shared/models/models.go
git commit -m "feat(shared): add LLM scheme types, policy references, verification result"
```

---

## Task 11: shared/config 更新 — 新增 LLMGatewayURL

**Files:**
- Modify: `shared/config/config.go`

- [ ] **Step 1: 在 Config 结构体和 Load 函数中添加 LLMGatewayURL**

在 `shared/config/config.go` 的 `Config` 结构体中追加：
```go
LLMGatewayURL string
```

在 `Load()` 函数中追加：
```go
LLMGatewayURL: getEnv("LLM_GATEWAY_URL", "http://localhost:39404"),
```

- [ ] **Step 2: 编译验证**

```sh
cd shared && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add shared/config/config.go
git commit -m "feat(shared): add LLMGatewayURL config field"
```

---

## Task 12: api-server migration — plan_snapshots 新列 + verification_logs

**Files:**
- Create: `services/api-server/migrations/008_plan_llm_fields.sql`

- [ ] **Step 1: 创建 migration**

`services/api-server/migrations/008_plan_llm_fields.sql`:
```sql
BEGIN;

ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS free_form_text TEXT DEFAULT '';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS structured_schemes JSONB DEFAULT '[]';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS policy_references JSONB DEFAULT '[]';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS recommendation TEXT DEFAULT '';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS recommendation_reason TEXT DEFAULT '';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS verification_result JSONB;

CREATE TABLE IF NOT EXISTS plan_verification_logs (
  id BIGSERIAL PRIMARY KEY,
  plan_id VARCHAR(100),
  llm_provider VARCHAR(50),
  llm_scheme_name VARCHAR(200),
  metric VARCHAR(50),
  llm_value DECIMAL(15,2),
  actuary_value DECIMAL(15,2),
  deviation_pct DECIMAL(5,2),
  root_cause VARCHAR(50),
  resolution TEXT,
  resolved BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verification_logs_plan_id ON plan_verification_logs(plan_id);
CREATE INDEX IF NOT EXISTS idx_verification_logs_resolved ON plan_verification_logs(resolved);

COMMIT;
```

- [ ] **Step 2: Commit**

```bash
git add services/api-server/migrations/008_plan_llm_fields.sql
git commit -m "feat(api-server): add migration for LLM plan fields and verification logs"
```

---

## Task 13: api-server policy_repo — QueryByRegionHierarchy + 溯源字段

**Files:**
- Modify: `services/api-server/internal/repository/policy_repo.go`

- [ ] **Step 1: 新增 QueryByRegionHierarchy 方法**

在 `policy_repo.go` 的 `PolicyRepository` 接口中追加：
```go
QueryByRegionHierarchy(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error)
```

实现此方法，核心逻辑：

```go
func (r *policyRepository) QueryByRegionHierarchy(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error) {
	codes := buildRegionHierarchy(regionCode)

	query := `SELECT claim_id, policy_id, region_code, policy_type, target_group_tags,
		subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
		effective_date, expire_date, confidence_score, status, version_number,
		conditions, required_documents,
		source_id, source_name, source_url, policy_url, policy_title,
		issuing_authority, document_number, application_process
		FROM policy_claims
		WHERE region_code = ANY($1) AND status = $2
		AND (expire_date IS NULL OR expire_date > NOW())
		ORDER BY
			CASE region_code
				WHEN '000000' THEN 1
				ELSE 2
			END, updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(codes), status)
	// ... scan 同 QueryByRegionAndStatus，但额外扫描 source 字段
	// 返回 claims
}

func buildRegionHierarchy(code string) []string {
	if len(code) < 6 {
		return []string{code, "000000"}
	}
	var codes []string
	codes = append(codes, "000000")
	codes = append(codes, code[:2]+"0000")
	codes = append(codes, code[:4]+"00")
	codes = append(codes, code)
	return codes
}
```

注意：需要在 `QueryByRegionAndStatus` 的 SELECT 和 Scan 中也加上 `source_id, source_name, source_url, policy_url, policy_title, issuing_authority, document_number, application_process` 这些溯源字段，目前它们在 DB 中有但在查询中未返回。这是"链接 + 提取的原文片段"展示的前提。

- [ ] **Step 2: 同时修改现有的 Query 和 QueryByRegionAndStatus 方法**

在所有查询的 SELECT 列表中追加溯源字段，Scan 中追加对应变量。

- [ ] **Step 3: 运行 api-server 测试确保不破坏现有功能**

```sh
cd services/api-server && go test ./internal/repository/ -v -run TestPolicy
```

- [ ] **Step 4: Commit**

```bash
git add services/api-server/internal/repository/policy_repo.go
git commit -m "feat(api-server): add QueryByRegionHierarchy and source fields in policy queries"
```

---

## Task 14: api-server plan_repo 适配新 PlanSnapshot

**Files:**
- Modify: `services/api-server/internal/repository/plan_repo.go`

- [ ] **Step 1: 更新 Save 方法以存储新字段**

修改 `plan_repo.go` 的 `Save` 方法：
```go
func (r *planRepository) Save(ctx context.Context, plan *models.PlanSnapshot) error {
	schemesJSON, _ := json.Marshal(plan.RecommendedSchemes)
	structuredJSON, _ := json.Marshal(plan.StructuredSchemes)
	refsJSON, _ := json.Marshal(plan.PolicyReferences)
	var verificationJSON []byte
	if plan.VerificationResult != nil {
		verificationJSON, _ = json.Marshal(plan.VerificationResult)
	}

	query := `
		INSERT INTO plan_snapshots (plan_id, user_id, recommended_schemes,
			free_form_text, structured_schemes, policy_references,
			recommendation, recommendation_reason, verification_result,
			total_cost, total_subsidy, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (plan_id) DO UPDATE SET
			recommended_schemes = EXCLUDED.recommended_schemes,
			free_form_text = EXCLUDED.free_form_text,
			structured_schemes = EXCLUDED.structured_schemes,
			policy_references = EXCLUDED.policy_references,
			recommendation = EXCLUDED.recommendation,
			recommendation_reason = EXCLUDED.recommendation_reason,
			verification_result = EXCLUDED.verification_result,
			total_cost = EXCLUDED.total_cost,
			total_subsidy = EXCLUDED.total_subsidy,
			updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query,
		plan.PlanID, plan.UserID, schemesJSON,
		plan.FreeFormText, structuredJSON, refsJSON,
		plan.Recommendation, plan.RecommendationReason,
		verificationJSON,
		plan.TotalCost, plan.TotalSubsidy, plan.GeneratedAt)
	if err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: 更新 GetByID 以读取新字段**

修改 `GetByID` 的 Scan 以包含新列。注意 `verification_result` 可能为 NULL，需要用 `[]byte` 接收。

- [ ] **Step 3: 运行测试**

```sh
cd services/api-server && go test ./internal/repository/ -v -run TestPlan
```

- [ ] **Step 4: Commit**

```bash
git add services/api-server/internal/repository/plan_repo.go
git commit -m "feat(api-server): update plan_repo for LLM scheme fields"
```

---

## Task 15: api-server plan_handler 重写 — LLM 方案生成

**Files:**
- Modify: `services/api-server/internal/handler/plan_handler.go`
- Modify: `services/api-server/internal/handler/plan_handler_test.go`

这是核心改造。新的 `GeneratePlanHandler` 不再接受 `Calculator` 参数，改为接受 `LLMGatewayURL`。

- [ ] **Step 1: 重写 plan_handler.go**

新签名：
```go
func GeneratePlanHandler(llmGatewayURL string, repo PlanRepository, profileRepo ProfileLookuper, policyRepo PolicyQuerier) http.Handler
```

核心流程：
1. 解析请求（保留现有验证：age 16-70, gender male/female, monthly_budget > 0）
2. 获取用户画像 → 确定地区代码
3. 调用 `policyRepo.QueryByRegionHierarchy(ctx, code, "verified")` 获取三级政策
4. 过滤 target_group_tags 匹配用户条件的政策
5. 组装 LLM Prompt（system prompt + user content = 画像JSON + 政策数据JSON）
6. HTTP POST 调 llm-gateway `/v1/chat`
7. 解析 LLM 响应（提取 `===FREE_FORM_START===` 和 `===STRUCTURED_START===` 之间的内容）
8. [异步 goroutine] 精算引擎验证：从 LLM 输出中提取数值 → 调 actuarial-engine → 比对 → 记录偏差
9. 构建 PlanSnapshot 并保存
10. 返回响应

LLM Gateway 客户端（内联在 plan_handler.go 中）：
```go
type LLMGatewayClient struct {
	URL string
}

type LLMChatRequest struct {
	SystemPrompt string `json:"system_prompt"`
	UserContent  string `json:"user_content"`
	MaxTokens    int    `json:"max_tokens"`
	Caller       string `json:"caller"`
}

type LLMChatResponse struct {
	Code         int    `json:"code"`
	Content      string `json:"content"`
	ProviderUsed string `json:"provider_used"`
	LatencyMs    int64  `json:"latency_ms"`
}

func (c *LLMGatewayClient) Chat(ctx context.Context, systemPrompt, userContent string) (*LLMChatResponse, error) {
	body, _ := json.Marshal(LLMChatRequest{
		SystemPrompt: systemPrompt,
		UserContent:  userContent,
		MaxTokens:    8192,
		Caller:       "api-server",
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/v1/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm-gateway call failed: %w", err)
	}
	defer resp.Body.Close()

	var llmResp LLMChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("llm-gateway error: %s", llmResp.Content)
	}
	return &llmResp, nil
}
```

Prompt 组装函数：
```go
func buildPlanPrompt(profile *models.UserProfile, policies []models.PolicyClaim) (systemPrompt, userContent string) {
	systemPrompt = `你是一位资深社保政策顾问。根据用户的个人情况和所在地的社保政策，为用户量身定制最优社保参保方案。

规则：
1. 所有政策依据必须来自提供的政策库数据，不可编造
2. 必须理解上位法与下位法的关系：国家法律 > 省级规定 > 市级细则 > 区级优惠
3. 当地方政策与上位法冲突时，以有利于用户的原则解释
4. 方案必须包含所有适用的优惠政策，不能遗漏
5. 每一条建议都必须标注所依据的政策（标题+文号+链接）
6. 数值计算必须精确，基于提供的费率和基数

输出格式（必须严格遵守）：
===FREE_FORM_START===
[自由文本格式的方案建议书，面向普通用户，通俗易懂，2000字以内]
===FREE_FORM_END===
===STRUCTURED_START===
{
  "summary": "一句话总结",
  "schemes": [
    {
      "name": "方案名称",
      "description": "方案描述",
      "monthly_cost": 0,
      "annual_subsidy": 0,
      "projected_pension": 0,
      "total_cost": 0,
      "contribution_base": 0,
      "pension_employee_rate": 0,
      "pension_employer_rate": 0,
      "medical_employee_rate": 0,
      "analysis": "该方案的详细分析",
      "applicable_policies": ["claim_id_1"]
    }
  ],
  "policy_references": [
    {
      "claim_id": "",
      "policy_title": "",
      "document_number": "",
      "policy_url": "",
      "relevant_excerpt": "提取的原文片段",
      "how_applied": "如何应用于本方案"
    }
  ],
  "recommendation": {
    "recommended_scheme": "方案名称",
    "reasoning": "推荐理由"
  }
}
===STRUCTURED_END===`

	policyJSON, _ := json.Marshal(policies)
	profileJSON, _ := json.Marshal(profile)

	userContent = fmt.Sprintf(`## 用户画像
%s

## 适用政策（共%d条）
%s

## 请基于以上信息，为该用户生成最优社保参保方案。`, string(profileJSON), len(policies), string(policyJSON))

	return systemPrompt, userContent
}
```

LLM 响应解析函数：
```go
func parseLLMResponse(raw string) (freeForm string, structured *models.LLMSchemeResponse, err error) {
	freeStart := strings.Index(raw, "===FREE_FORM_START===")
	freeEnd := strings.Index(raw, "===FREE_FORM_END===")
	if freeStart >= 0 && freeEnd > freeStart {
		freeForm = strings.TrimSpace(raw[freeStart+len("===FREE_FORM_START===") : freeEnd])
	}

	structStart := strings.Index(raw, "===STRUCTURED_START===")
	structEnd := strings.Index(raw, "===STRUCTURED_END===")
	if structStart >= 0 && structEnd > structStart {
		jsonStr := strings.TrimSpace(raw[structStart+len("===STRUCTURED_START===") : structEnd])
		var resp models.LLMSchemeResponse
		if jsonErr := json.Unmarshal([]byte(jsonStr), &resp); jsonErr != nil {
			return freeForm, nil, fmt.Errorf("parse structured JSON: %w", jsonErr)
		}
		structured = &resp
	}

	if freeForm == "" && structured == nil {
		return "", nil, fmt.Errorf("LLM response missing required markers")
	}
	return freeForm, structured, nil
}
```

注意：`LLMSchemeResponse` 需要在 `shared/models/models.go` 中定义：
```go
type LLMSchemeResponse struct {
	Summary          string            `json:"summary"`
	Schemes          []LLMScheme       `json:"schemes"`
	PolicyReferences []PolicyReference `json:"policy_references"`
	Recommendation   struct {
		RecommendedScheme string `json:"recommended_scheme"`
		Reasoning         string `json:"reasoning"`
	} `json:"recommendation"`
}
```

将此类型加到 Task 10 的 models.go 中。

- [ ] **Step 2: 重写 plan_handler_test.go**

更新 mock 类型。移除 `mockCalculator`，新增 `mockLLMGateway`：

```go
type mockLLMGateway struct {
	resp *LLMChatResponse
	err  error
}
```

测试用例：
- TestGeneratePlanHandlerSuccess：mock LLM 返回合法双视图响应
- TestGeneratePlanHandlerInvalidAge：保留
- TestGeneratePlanHandlerInvalidGender：保留
- TestGeneratePlanHandlerInvalidBudget：保留
- TestGeneratePlanHandlerLLMError：mock LLM 返回错误
- TestGeneratePlanHandlerParseError：mock LLM 返回无 marker 的文本

- [ ] **Step 3: 运行测试**

```sh
cd services/api-server && go test ./internal/handler/ -v -run TestGeneratePlan
```

Expected: 所有测试 PASS。

- [ ] **Step 4: Commit**

```bash
git add services/api-server/internal/handler/plan_handler.go services/api-server/internal/handler/plan_handler_test.go
git commit -m "feat(api-server): rewrite plan_handler for LLM-based generation"
```

---

## Task 16: api-server main.go 更新

**Files:**
- Modify: `services/api-server/cmd/main.go`

- [ ] **Step 1: 修改路由注册**

在 `cmd/main.go` 中：

1. 将 `GeneratePlanHandler` 的调用从 `GeneratePlanHandler(calculator, ...)` 改为 `GeneratePlanHandler(cfg.LLMGatewayURL, ...)`
2. 保留 `calculator` 初始化（用于精算验证），但不再传给 GeneratePlanHandler
3. 在 `mux.Handle("/v1/plans/generate", ...)` 处使用新签名

关键改动：
```go
// 旧：handler.GeneratePlanHandler(calculator, planRepo, profileRepo, policyRepo)
// 新：handler.GeneratePlanHandler(cfg.LLMGatewayURL, planRepo, profileRepo, policyRepo)
```

- [ ] **Step 2: 编译验证**

```sh
cd services/api-server && go build ./cmd/main.go
```

- [ ] **Step 3: Commit**

```bash
git add services/api-server/cmd/main.go
git commit -m "feat(api-server): wire LLM gateway URL into plan handler"
```

---

## Task 17: WebClient 改造 — 双视图 + 政策依据

**Files:**
- Modify: `services/api-server/internal/handler/webclient_handler.go`

- [ ] **Step 1: 修改方案生成 tab 的渲染逻辑**

在 `webclient_handler.go` 的 `onGeneratePlan()` JS 函数中：

1. 发送请求到 `/v1/plans/generate`（请求体不变）
2. 解析响应中的新字段：`free_form_text`, `structured_schemes`, `policy_references`, `recommendation`, `verification_result`
3. 渲染方案结果区域，包含：
   - 视图切换下拉：`自由文本` / `结构化分析`
   - 自由文本视图：渲染 `free_form_text`（支持简单 markdown → HTML 转换：换行→`<br>`，`**text**`→`<strong>`）
   - 结构化视图：schemes 列表（表格形式：方案名/缴费基数/月缴费/年补贴/预计养老金）+ 每个方案的 analysis 文字
   - 推荐方案高亮卡片 + 推荐理由
   - 政策依据区域：每个 PolicyReference 显示为卡片，包含：
     - `policy_title` + `document_number`
     - 可点击的 `policy_url` 链接
     - `relevant_excerpt` 原文片段（灰色背景引用块）
     - `how_applied` 应用说明

- [ ] **Step 2: 验证 webclient 页面可正常加载**

```sh
curl -s http://localhost:39401/webclient | head -20
```

Expected: HTML 页面正常返回。

- [ ] **Step 3: Commit**

```bash
git add services/api-server/internal/handler/webclient_handler.go
git commit -m "feat(api-server): add dual-view plan display with policy references"
```

---

## Task 18: Docker 集成 — llm-gateway 服务

**Files:**
- Modify: `docker-compose.yml`

- [ ] **Step 1: 在 docker-compose.yml 中添加 llm-gateway 服务**

在 `policy-crawler` 服务之后、`db-init` 之前追加：

```yaml
  llm-gateway:
    image: nsi-llm-gateway:latest
    container_name: nsi-llm-gateway
    ports:
      - "39404:39404"
    environment:
      DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD:?POSTGRES_PASSWORD must be set}@postgres:5432/nsi_llm?sslmode=disable
      SERVER_PORT: "39404"
      ADMIN_USER: ${ADMIN_USERNAME:-admin}
      ADMIN_PASS: ${ADMIN_PASSWORD:-changeme}
    depends_on:
      postgres:
        condition: service_healthy
      db-migrate:
        condition: service_completed_successfully
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:39404/healthz"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

- [ ] **Step 2: 修改 api-server 环境变量**

在 api-server 的 `environment` 中追加：
```yaml
      LLM_GATEWAY_URL: http://llm-gateway:39404
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml
git commit -m "feat(docker): add llm-gateway service to docker-compose"
```

---

## Task 19: Makefile 更新

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: 添加 llm-gateway 构建和测试目标**

在 `Makefile` 中：

1. 修改 `build` 和 `test` 目标添加 `build-llm-gateway` 和 `test-llm-gateway`
2. 新增：
```makefile
build-llm-gateway:
	cd services/llm-gateway && go build ./cmd/main.go

test-llm-gateway:
	cd services/llm-gateway && go test ./... -count=1 -v
```
3. 修改 `lint` 目标添加 llm-gateway
4. 修改 `clean` 目标添加 llm-gateway

- [ ] **Step 2: Commit**

```bash
git add Makefile
git commit -m "feat(makefile): add llm-gateway build and test targets"
```

---

## Task 20: 构建部署 + 端到端验证

**Files:** 无新文件

- [ ] **Step 1: 编译 llm-gateway Docker 镜像**

```sh
cd services/llm-gateway
GOOS=linux GOARCH=amd64 go build -o bin/llm-gateway ./cmd/main.go
docker build -t nsi-llm-gateway:latest .
```

- [ ] **Step 2: 重新编译 api-server Docker 镜像**

```sh
cd services/api-server
GOOS=linux GOARCH=amd64 go build -o bin/api-server ./cmd/main.go
docker build -t nsi-api-server:latest .
```

- [ ] **Step 3: docker compose up**

```sh
docker compose up -d --build
```

- [ ] **Step 4: 等待健康检查通过**

```sh
docker compose ps
```

Expected: 所有容器 healthy，包括 `nsi-llm-gateway`。

- [ ] **Step 5: 测试 llm-gateway healthz**

```sh
curl http://localhost:39404/healthz
```

Expected: `{"status":"ok"}`

- [ ] **Step 6: 访问 llm-gateway 管理页面**

浏览器访问 `http://localhost:39404/admin/`，用 admin/changeme 登录。配置一个 provider（如 DeepSeek），测试连通性。

- [ ] **Step 7: 通过 webclient 生成方案**

浏览器访问 `http://localhost:39401/webclient`，填写用户画像，生成方案。验证：
- 自由文本视图显示
- 结构化视图显示
- 政策依据卡片显示
- 视图切换正常

- [ ] **Step 8: Commit 验证通过标记**

```bash
git tag -a v-llm-plan-gen -m "LLM plan generation system deployed and verified"
```
