package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/extractor"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/llm"
)

type LLMConfig struct {
	Provider            string `json:"provider"`
	APIKey              string `json:"api_key"`
	Endpoint            string `json:"endpoint"`
	ModelName           string `json:"model_name"`
	MaxTokens           int    `json:"max_tokens"`
	Enabled             bool   `json:"enabled"`
	EmbeddingModel      string `json:"embedding_model"`
	EmbeddingDimensions int    `json:"embedding_dimensions"`
	EmbeddingAPIKey     string `json:"embedding_api_key"`
	EmbeddingEndpoint   string `json:"embedding_endpoint"`
	BackupProvider      string `json:"backup_provider"`
	BackupAPIKey        string `json:"backup_api_key"`
	BackupEndpoint      string `json:"backup_endpoint"`
	BackupModelName     string `json:"backup_model_name"`
}

type LLMStore interface {
	GetLLMConfig() (*LLMConfig, error)
	SaveLLMConfig(cfg *LLMConfig) error
	GetUnprocessedCount() (int, error)
	RunExtraction(limit int) (int, int, error)
	GetPendingRawTexts(limit int) ([]PendingRawText, error)
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func LLMConfigGetHandler(store LLMStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, err := store.GetLLMConfig()
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("get config: %v", err))
			return
		}
		masked := *cfg
		masked.APIKey = maskAPIKey(cfg.APIKey)
		masked.EmbeddingAPIKey = maskAPIKey(cfg.EmbeddingAPIKey)
		masked.BackupAPIKey = maskAPIKey(cfg.BackupAPIKey)
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": masked})
	})
}

func LLMConfigSaveHandler(store LLMStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var cfg LLMConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if cfg.Provider == "" {
			respondError(w, http.StatusBadRequest, "provider required")
			return
		}
		if strings.Contains(cfg.APIKey, "****") {
			existing, err := store.GetLLMConfig()
			if err == nil && existing != nil {
				cfg.APIKey = existing.APIKey
			}
		}
		if strings.Contains(cfg.EmbeddingAPIKey, "****") {
			existing, err := store.GetLLMConfig()
			if err == nil && existing != nil {
				cfg.EmbeddingAPIKey = existing.EmbeddingAPIKey
			}
		}
		if strings.Contains(cfg.BackupAPIKey, "****") {
			existing, err := store.GetLLMConfig()
			if err == nil && existing != nil {
				cfg.BackupAPIKey = existing.BackupAPIKey
			}
		}
		if err := store.SaveLLMConfig(&cfg); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("save: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": "saved"})
	})
}

func LLMStatusHandler(store LLMStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count, err := store.GetUnprocessedCount()
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("count: %v", err))
			return
		}
		cfg, _ := store.GetLLMConfig()
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"unprocessed":    count,
				"llm_configured": cfg != nil && cfg.Enabled && cfg.APIKey != "",
				"provider":       mapValue(cfg, func(c *LLMConfig) string { return c.Provider }),
			},
		})
	})
}

func LLMPendingHandler(store LLMStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, err := store.GetPendingRawTexts(100)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("pending: %v", err))
			return
		}
		if items == nil {
			items = []PendingRawText{}
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": items,
		})
	})
}

func LLMProgressHandler(store LLMStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := GlobalExtProgress
		p.Lock()
		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"total":       p.Total,
				"completed":   p.Completed,
				"failed":      p.Failed,
				"running":     p.Running,
				"done":        p.Done,
				"current":     p.CurrentID,
				"current_src": p.CurrentSrc,
			},
		}
		p.Unlock()
		respondJSON(w, http.StatusOK, resp)
	})
}

func LLMExtractRunHandler(store interface{}, checker extractor.ReferenceChecker, embedProv embeddings.EmbeddingProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 类型断言：store 需要同时实现 LLMStore 和 extractor.RawTextStore
		llmStore, ok1 := store.(LLMStore)
		rawStore, ok2 := store.(extractor.RawTextStore)
		if !ok1 || !ok2 {
			respondError(w, http.StatusInternalServerError, "store does not support extraction")
			return
		}

		p := GlobalExtProgress
		p.Lock()
		if p.Running {
			p.Unlock()
			respondError(w, http.StatusConflict, "提取任务已在运行中")
			return
		}
		p.Total = 0
		p.Completed = 0
		p.Failed = 0
		p.Running = true
		p.Done = false
		p.CurrentID = 0
		p.CurrentSrc = ""
		p.Unlock()

		go func() {
			cfg, err := llmStore.GetLLMConfig()
			if err != nil {
				log.Printf("[extract] get config: %v", err)
				finishExtract("get config error: " + err.Error())
				return
			}

			client := llm.NewClient(llm.Config{
				Provider:  llm.ParseProvider(cfg.Provider),
				APIKey:    cfg.APIKey,
				Endpoint:  cfg.Endpoint,
				ModelName: cfg.ModelName,
				MaxTokens: cfg.MaxTokens,
				Enabled:   cfg.Enabled,
			})

			entries, err := rawStore.GetUnprocessedRawTexts(100)
			if err != nil {
				log.Printf("[extract] get pending: %v", err)
				finishExtract("get pending error: " + err.Error())
				return
			}

			p.Lock()
			p.Total = len(entries)
			p.Unlock()

			ext := extractor.NewExtractor(rawStore, client)
			if checker != nil {
				ext.SetReferenceChecker(checker)
			}
			if embedProv != nil {
				ext.SetEmbeddingProvider(embedProv)
			}

			for _, entry := range entries {
				p.Lock()
				p.CurrentID = entry.ID
				p.CurrentSrc = entry.SourceID
				p.Unlock()

				if err := ext.ProcessOne(entry); err != nil {
					log.Printf("[extract] failed id=%d source=%s: %v", entry.ID, entry.SourceID, err)
					rawStore.SaveExtractLog(entry.SourceID, false, err.Error())
					p.Lock()
					p.Failed++
					p.Unlock()
				} else {
					p.Lock()
					p.Completed++
					p.Unlock()
				}
			}

			finishExtract("")
		}()

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "提取任务已启动",
		})
	})
}

func finishExtract(msg string) {
	p := GlobalExtProgress
	p.Lock()
	p.Running = false
	p.Done = true
	if msg != "" {
		log.Printf("[extract] finished with error: %s", msg)
	}
	p.Unlock()
}

func mapValue[T any, R any](v *T, fn func(*T) R) R {
	var zero R
	if v == nil {
		return zero
	}
	return fn(v)
}
