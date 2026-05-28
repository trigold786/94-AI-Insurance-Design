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
