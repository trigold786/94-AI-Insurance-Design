package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	client := p.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	reqBody := map[string]interface{}{
		"model": p.model,
		"input": texts,
	}
	if p.dimensions > 0 {
		reqBody["dimensions"] = p.dimensions
	}

	body, _ := json.Marshal(reqBody)
	hReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := client.Do(hReq)
	if err != nil {
		return nil, fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		snippet := string(respBody)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return nil, fmt.Errorf("embedding API %d: %s", resp.StatusCode, snippet)
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	vecs := make([][]float64, len(result.Data))
	for i, d := range result.Data {
		vecs[i] = d.Embedding
	}
	return vecs, nil
}

func (p *OpenAIProvider) Dimensions() int  { return p.dimensions }
func (p *OpenAIProvider) ModelName() string { return p.model }
