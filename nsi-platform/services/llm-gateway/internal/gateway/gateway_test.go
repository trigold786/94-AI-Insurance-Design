package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
			Choices: []provider.ChatChoice{
				{Message: provider.ChatMessage{Content: "fallback ok"}},
			},
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

	gw := New(providers)
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

	gw := New(providers)
	_, _, err := gw.Chat("sys", "user")
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestGateway_NoProviders(t *testing.T) {
	gw := New(nil)
	_, _, err := gw.Chat("sys", "user")
	if err == nil {
		t.Fatal("expected error with no providers")
	}
}

func TestGateway_UpdateProviders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := provider.ChatResponse{
			Choices: []provider.ChatChoice{
				{Message: provider.ChatMessage{Content: "new provider"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	gw := New(nil)
	gw.UpdateProviders([]provider.Provider{
		&provider.OpenAICompatProvider{
			Endpoint: srv.URL, APIKey: "k", ModelName: "new-model", MaxTokens: 100, HTTPClient: srv.Client(),
		},
	})

	content, name, err := gw.Chat("sys", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "new provider" {
		t.Errorf("expected 'new provider', got '%s'", content)
	}
	if name != "new-model" {
		t.Errorf("expected 'new-model', got '%s'", name)
	}
}
