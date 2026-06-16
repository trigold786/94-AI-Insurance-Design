package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Provider 大模型提供商配置
type Provider int

const (
	ProviderDeepSeek    Provider = iota // 默认：DeepSeek
	ProviderAliBailian                  // 阿里云百炼（通义千问）
	ProviderVolcArk                     // 火山方舟
	ProviderOpenCodeGo                  // OpenCode Go 订阅
)

func (p Provider) String() string {
	switch p {
	case ProviderDeepSeek:
		return "deepseek"
	case ProviderAliBailian:
		return "ali_bailian"
	case ProviderVolcArk:
		return "volc_ark"
	case ProviderOpenCodeGo:
		return "opencode_go"
	}
	return "deepseek"
}

func ParseProvider(s string) Provider {
	switch s {
	case "deepseek":
		return ProviderDeepSeek
	case "ali_bailian":
		return ProviderAliBailian
	case "volc_ark":
		return ProviderVolcArk
	case "opencode_go":
		return ProviderOpenCodeGo
	}
	return ProviderDeepSeek
}

// Config 大模型配置
type Config struct {
	Provider   Provider `json:"provider"`
	APIKey     string   `json:"api_key"`
	Endpoint   string   `json:"endpoint"`
	ModelName  string   `json:"model_name"`
	MaxTokens  int      `json:"max_tokens"`
	Enabled    bool     `json:"enabled"`
}

// DefaultConfig 默认配置（DeepSeek）
func DefaultConfig() Config {
	return Config{
		Provider:  ProviderDeepSeek,
		Endpoint:  "https://api.deepseek.com/v1/chat/completions",
		ModelName: "deepseek-v4-flash",
		MaxTokens: 4096,
		Enabled:   false,
	}
}

// 各 Provider 默认值
var providerDefaults = map[Provider]struct {
	endpoint  string
	modelName string
}{
	ProviderDeepSeek:    {"https://api.deepseek.com/v1/chat/completions", "deepseek-v4-flash"},
	ProviderAliBailian:  {"https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", "qwen3.6-plus"},
	ProviderVolcArk:     {"https://ark.cn-beijing.volces.com/api/v3/chat/completions", "doubao-pro-32k"},
	ProviderOpenCodeGo:  {"http://localhost:11434/v1/chat/completions", "opencode-go"},
}

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest OpenAI 兼容请求格式
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"`
}

// ChatResponse OpenAI 兼容响应格式
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// 通义千问请求格式
type bailianRequest struct {
	Model string `json:"model"`
	Input struct {
		Messages []ChatMessage `json:"messages"`
	} `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type bailianResponse struct {
	Output struct {
		Text string `json:"text"`
	} `json:"output"`
}

// Client 统一的 LLM 客户端
type Client struct {
	config      Config
	backup      *Config
	http        *http.Client
	usedBackup  bool
}

func NewClient(cfg Config) *Client {
	return &Client{
		config: cfg,
		http:   &http.Client{Timeout: 300 * time.Second},
	}
}

func NewClientWithBackup(cfg Config, backup *Config) *Client {
	return &Client{
		config: cfg,
		backup: backup,
		http:   &http.Client{Timeout: 300 * time.Second},
	}
}

func (c *Client) ModelName() string {
	if c.usedBackup && c.backup != nil {
		return c.backup.ModelName
	}
	return c.config.ModelName
}

func (c *Client) UsedBackup() bool {
	return c.usedBackup
}

func (c *Client) chatPrimary(systemPrompt, userContent string) (string, error) {
	switch c.config.Provider {
	case ProviderAliBailian:
		return c.chatBailianWithConfig(c.config, systemPrompt, userContent)
	default:
		return c.chatOpenAIWithConfig(c.config, systemPrompt, userContent)
	}
}

func (c *Client) Chat(systemPrompt, userContent string) (string, error) {
	c.usedBackup = false
	const maxPrimaryRetries = 5
	for attempt := 1; attempt <= maxPrimaryRetries; attempt++ {
		result, err := c.chatPrimary(systemPrompt, userContent)
		if err == nil {
			return result, nil
		}
		if c.backup == nil {
			return "", err
		}
		log.Printf("[llm] primary model attempt %d/%d failed: %v", attempt, maxPrimaryRetries, err)
		if attempt < maxPrimaryRetries {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}
	log.Printf("[llm] primary model failed %d times, falling back to backup", maxPrimaryRetries)
	c.usedBackup = true
	return c.chatWithBackup(systemPrompt, userContent)
}

func (c *Client) chatWithBackup(systemPrompt, userContent string) (string, error) {
	switch c.backup.Provider {
	case ProviderAliBailian:
		return c.chatBailianWithConfig(*c.backup, systemPrompt, userContent)
	default:
		return c.chatOpenAIWithConfig(*c.backup, systemPrompt, userContent)
	}
}

// chatOpenAI OpenAI 兼容格式（DeepSeek/火山方舟/OpenCode Go）
func doChatOpenAI(cfg Config, httpClient *http.Client, systemPrompt, userContent string) (string, error) {
	req := ChatRequest{
		Model: cfg.ModelName,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		MaxTokens: cfg.MaxTokens,
	}

	endpoint := cfg.Endpoint
	if !strings.Contains(endpoint, "/chat/completions") && !strings.Contains(endpoint, "/embeddings") {
		endpoint = strings.TrimRight(endpoint, "/") + "/chat/completions"
	}
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API %d: %s", resp.StatusCode, string(respBody)[:min(200, len(respBody))])
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

func (c *Client) chatOpenAIWithConfig(cfg Config, systemPrompt, userContent string) (string, error) {
	return doChatOpenAI(cfg, c.http, systemPrompt, userContent)
}

// chatBailian 阿里云百炼（通义千问）格式
func (c *Client) chatBailianWithConfig(cfg Config, systemPrompt, userContent string) (string, error) {
	// 2026-05: 阿里云百炼已全面支持 OpenAI 兼容模式
	if strings.Contains(cfg.Endpoint, "compatible-mode") {
		return c.chatOpenAIWithConfig(cfg, systemPrompt, userContent)
	}

	req := bailianRequest{
		Model: cfg.ModelName,
		Parameters: map[string]interface{}{
			"max_tokens": cfg.MaxTokens,
			"result_format": "text",
		},
	}
	req.Input.Messages = []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent},
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API %d: %s", resp.StatusCode, string(respBody)[:min(200, len(respBody))])
	}

	var bailianResp bailianResponse
	if err := json.Unmarshal(respBody, &bailianResp); err != nil {
		return "", fmt.Errorf("parse bailian response: %w", err)
	}
	return bailianResp.Output.Text, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
