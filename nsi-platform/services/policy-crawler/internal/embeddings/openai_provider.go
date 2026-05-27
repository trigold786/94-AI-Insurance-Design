package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	client := p.client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	isMultimodal := strings.Contains(p.baseURL, "/multimodal")

	var reqBody map[string]interface{}
	if isMultimodal {
		input := make([]map[string]interface{}, len(texts))
		for i, t := range texts {
			input[i] = map[string]interface{}{
				"type": "text",
				"text": t,
			}
		}
		reqBody = map[string]interface{}{
			"model": p.model,
			"input": input,
		}
	} else {
		reqBody = map[string]interface{}{
			"model": p.model,
			"input": texts,
		}
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
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Try array format first (OpenAI), then object format (ARK multimodal)
	var embeddings []float64
	if err := json.Unmarshal(result.Data, &embeddings); err == nil {
		return [][]float64{embeddings}, nil
	}

	var arr []struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.Unmarshal(result.Data, &arr); err == nil && len(arr) > 0 {
		vecs := make([][]float64, len(arr))
		for i, d := range arr {
			vecs[i] = d.Embedding
		}
		return vecs, nil
	}

	var obj struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.Unmarshal(result.Data, &obj); err == nil && obj.Embedding != nil {
		return [][]float64{obj.Embedding}, nil
	}

	return nil, fmt.Errorf("unexpected embedding response format: %s", string(result.Data))
}

func (p *OpenAIProvider) Dimensions() int  { return p.dimensions }
func (p *OpenAIProvider) ModelName() string { return p.model }
