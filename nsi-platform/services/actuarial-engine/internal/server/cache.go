package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, data []byte, ttl time.Duration)
}

type NoopCache struct{}

func (NoopCache) Get(key string) ([]byte, bool) { return nil, false }
func (NoopCache) Set(key string, data []byte, ttl time.Duration) {}

type cacheEntry struct {
	data    []byte
	expires time.Time
}

type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

func NewInMemoryCache(cleanupInterval time.Duration) *InMemoryCache {
	c := &InMemoryCache{
		entries: make(map[string]cacheEntry),
	}
	if cleanupInterval > 0 {
		go c.cleanupLoop(cleanupInterval)
	}
	return c
}

func (c *InMemoryCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expires) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}
	return entry.data, true
}

func (c *InMemoryCache) Set(key string, data []byte, ttl time.Duration) {
	c.mu.Lock()
	c.entries[key] = cacheEntry{
		data:    data,
		expires: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

func (c *InMemoryCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.entries {
			if now.After(v.expires) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}

func cacheKey(req PlanRequest) string {
	data, _ := json.Marshal(req)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("actuary:calc:%x", hash)
}
