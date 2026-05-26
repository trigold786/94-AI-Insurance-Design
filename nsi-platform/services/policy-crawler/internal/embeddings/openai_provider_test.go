package embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProvider_Dimensions(t *testing.T) {
	p := &OpenAIProvider{dimensions: 1536}
	if p.Dimensions() != 1536 {
		t.Fatalf("expected 1536, got %d", p.Dimensions())
	}
}

func TestOpenAIProvider_ModelName(t *testing.T) {
	p := &OpenAIProvider{model: "text-embedding-3-small"}
	if p.ModelName() != "text-embedding-3-small" {
		t.Fatalf("expected text-embedding-3-small, got %s", p.ModelName())
	}
}

func TestOpenAIProvider_Embed_Success(t *testing.T) {
	embedding := make([]float64, 1536)
	embedding[0] = 0.1
	embedding[1] = 0.2
	resp := map[string]interface{}{
		"data": []map[string]interface{}{
			{"embedding": embedding},
		},
	}
	body, _ := json.Marshal(resp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(401)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		w.Write(body)
	}))
	defer server.Close()

	p := &OpenAIProvider{
		apiKey:     "test-key",
		baseURL:    server.URL,
		model:      "text-embedding-3-small",
		dimensions: 1536,
	}
	vecs, err := p.Embed(context.Background(), []string{"test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 1 {
		t.Fatalf("expected 1 vector, got %d", len(vecs))
	}
	if len(vecs[0]) != 1536 {
		t.Fatalf("expected 1536 dims, got %d", len(vecs[0]))
	}
	if vecs[0][0] != 0.1 || vecs[0][1] != 0.2 {
		t.Fatalf("expected [0.1, 0.2, ...], got [%f, %f, ...]", vecs[0][0], vecs[0][1])
	}
}

func TestOpenAIProvider_Embed_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer server.Close()

	p := &OpenAIProvider{
		apiKey:     "test-key",
		baseURL:    server.URL,
		model:      "text-embedding-3-small",
		dimensions: 1536,
	}
	_, err := p.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestOpenAIProvider_Embed_EmptyResponse(t *testing.T) {
	resp := map[string]interface{}{"data": []map[string]interface{}{}}
	body, _ := json.Marshal(resp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}))
	defer server.Close()

	p := &OpenAIProvider{
		apiKey:     "test-key",
		baseURL:    server.URL,
		model:      "text-embedding-3-small",
		dimensions: 1536,
	}
	_, err := p.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}
