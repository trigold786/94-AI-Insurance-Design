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
			Choices: []ChatChoice{
				{Message: ChatMessage{Content: "hello world"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := &OpenAICompatProvider{
		Endpoint:   srv.URL,
		APIKey:     "test-key",
		ModelName:  "test-model",
		MaxTokens:  100,
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

func TestOpenAICompatChat_NameModel(t *testing.T) {
	p := &OpenAICompatProvider{ModelName: "deepseek-chat"}
	if p.Name() != "deepseek-chat" {
		t.Errorf("expected 'deepseek-chat', got '%s'", p.Name())
	}
	if p.Model() != "deepseek-chat" {
		t.Errorf("expected 'deepseek-chat', got '%s'", p.Model())
	}
}

func TestBailianChat_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}
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

func TestBailianChat_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer srv.Close()

	p := &BailianProvider{
		Endpoint: srv.URL, APIKey: "k", ModelName: "m", MaxTokens: 100, HTTPClient: srv.Client(),
	}
	_, err := p.Chat("sys", "user")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestBailianChat_NameModel(t *testing.T) {
	p := &BailianProvider{ModelName: "qwen-plus"}
	if p.Name() != "ali_bailian" {
		t.Errorf("expected 'ali_bailian', got '%s'", p.Name())
	}
	if p.Model() != "qwen-plus" {
		t.Errorf("expected 'qwen-plus', got '%s'", p.Model())
	}
}
