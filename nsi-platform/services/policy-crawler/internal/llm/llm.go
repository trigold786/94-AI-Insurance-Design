package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		ModelName: "deepseek-chat",
		MaxTokens: 4096,
		Enabled:   false,
	}
}

// 各 Provider 默认值
var providerDefaults = map[Provider]struct {
	endpoint  string
	modelName string
}{
	ProviderDeepSeek:    {"https://api.deepseek.com/v1/chat/completions", "deepseek-chat"},
	ProviderAliBailian:  {"https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation", "qwen-plus"},
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
	config Config
	http   *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		config: cfg,
		http:   &http.Client{Timeout: 60 * time.Second},
	}
}

// Chat 发送对话请求，返回文本响应
func (c *Client) Chat(systemPrompt, userContent string) (string, error) {
	switch c.config.Provider {
	case ProviderAliBailian:
		return c.chatBailian(systemPrompt, userContent)
	default:
		return c.chatOpenAI(systemPrompt, userContent)
	}
}

// chatOpenAI OpenAI 兼容格式（DeepSeek/火山方舟/OpenCode Go）
func (c *Client) chatOpenAI(systemPrompt, userContent string) (string, error) {
	req := ChatRequest{
		Model: c.config.ModelName,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		MaxTokens: c.config.MaxTokens,
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.http.Do(httpReq)
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

// chatBailian 阿里云百炼（通义千问）格式
func (c *Client) chatBailian(systemPrompt, userContent string) (string, error) {
	req := bailianRequest{
		Model: c.config.ModelName,
		Parameters: map[string]interface{}{
			"max_tokens": c.config.MaxTokens,
			"result_format": "text",
		},
	}
	req.Input.Messages = []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent},
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

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
