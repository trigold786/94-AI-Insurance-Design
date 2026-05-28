package gateway

import (
	"fmt"
	"log"
	"sync"

	"github.com/trigold786/94-AI-Insurance-Design/llm-gateway/internal/provider"
)

type Gateway struct {
	mu        sync.RWMutex
	providers []provider.Provider
}

func New(providers []provider.Provider) *Gateway {
	return &Gateway{providers: providers}
}

func (g *Gateway) UpdateProviders(providers []provider.Provider) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.providers = providers
}

func (g *Gateway) GetProviders() []provider.Provider {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.providers
}

func (g *Gateway) Chat(systemPrompt, userContent string) (content string, providerUsed string, err error) {
	g.mu.RLock()
	providers := make([]provider.Provider, len(g.providers))
	copy(providers, g.providers)
	g.mu.RUnlock()

	if len(providers) == 0 {
		return "", "", fmt.Errorf("no providers configured")
	}

	var lastErr error
	for _, p := range providers {
		result, err := p.Chat(systemPrompt, userContent)
		if err != nil {
			log.Printf("[gateway] provider %s failed: %v", p.Name(), err)
			lastErr = err
			continue
		}
		return result, p.Name(), nil
	}

	return "", "", fmt.Errorf("all providers failed, last error: %v", lastErr)
}
