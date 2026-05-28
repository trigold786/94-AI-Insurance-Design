package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatChoice struct {
	Message ChatMessage `json:"message"`
}

type ChatResponse struct {
	Choices []ChatChoice `json:"choices"`
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

	type chatRequest struct {
		Model     string        `json:"model"`
		Messages  []ChatMessage `json:"messages"`
		MaxTokens int           `json:"max_tokens,omitempty"`
	}

	req := chatRequest{
		Model: p.ModelName,
		Messages: []ChatMessage{
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
		truncated := string(respBody)
		if len(truncated) > 200 {
			truncated = truncated[:200]
		}
		return "", fmt.Errorf("API %d: %s", resp.StatusCode, truncated)
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
