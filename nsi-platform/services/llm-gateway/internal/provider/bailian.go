package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
		Messages []ChatMessage `json:"messages"`
	} `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type bailianResponse struct {
	Output struct {
		Text string `json:"text"`
	} `json:"output"`
}

func (p *BailianProvider) Chat(systemPrompt, userContent string) (string, error) {
	// 2026-05: 阿里云百炼已全面支持 OpenAI 兼容模式
	// 若 endpoint 包含 compatible-mode，使用标准 OpenAI 格式；否则回退到旧版 Bailian 格式
	if strings.Contains(p.Endpoint, "compatible-mode") {
		compat := &OpenAICompatProvider{
			Endpoint:  p.Endpoint,
			APIKey:    p.APIKey,
			ModelName: p.ModelName,
			MaxTokens: p.MaxTokens,
			HTTPClient: p.HTTPClient,
		}
		return compat.Chat(systemPrompt, userContent)
	}

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
	req.Input.Messages = []ChatMessage{
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
		truncated := string(respBody)
		if len(truncated) > 200 {
			truncated = truncated[:200]
		}
		return "", fmt.Errorf("bailian API %d: %s", resp.StatusCode, truncated)
	}

	var bResp bailianResponse
	if err := json.Unmarshal(respBody, &bResp); err != nil {
		return "", fmt.Errorf("parse bailian response: %w", err)
	}
	return bResp.Output.Text, nil
}
