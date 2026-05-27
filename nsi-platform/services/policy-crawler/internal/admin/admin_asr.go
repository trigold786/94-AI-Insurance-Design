package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func ASRConfigGetHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var id int64
		var provider, apiKey, endpoint, language string
		var sampleRate int
		var enabled bool
		err := db.QueryRow(
			"SELECT id, provider, api_key, endpoint, language, sample_rate, enabled FROM asr_configs LIMIT 1",
		).Scan(&id, &provider, &apiKey, &endpoint, &language, &sampleRate, &enabled)
		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"provider":    "",
					"api_key":     "",
					"endpoint":    "",
					"language":    "",
					"sample_rate": 0,
					"enabled":     false,
				},
			})
			return
		}
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("query error: %v", err))
			return
		}

		masked := maskAPIKey(apiKey)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"id":          id,
				"provider":    provider,
				"api_key":     masked,
				"endpoint":    endpoint,
				"language":    language,
				"sample_rate": sampleRate,
				"enabled":     enabled,
			},
		})
	})
}

func ASRConfigSaveHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Provider   string `json:"provider"`
			APIKey     string `json:"api_key"`
			Endpoint   string `json:"endpoint"`
			Language   string `json:"language"`
			SampleRate int    `json:"sample_rate"`
			Enabled    bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Provider == "" {
			respondError(w, http.StatusBadRequest, "provider required")
			return
		}
		if req.APIKey == "" {
			respondError(w, http.StatusBadRequest, "api_key required")
			return
		}

		_, err := db.Exec(
			`UPDATE asr_configs SET provider=?, api_key=?, endpoint=?, language=?, sample_rate=?, enabled=? WHERE id = (SELECT MIN(id) FROM asr_configs)`,
			req.Provider, req.APIKey, req.Endpoint, req.Language, req.SampleRate, req.Enabled,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("update error: %v", err))
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "ASR config saved",
		})
	})
}


