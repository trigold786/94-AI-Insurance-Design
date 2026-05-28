package admin

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func ASRConfigGetHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var id int64
		var provider, apiKey, appID, endpoint, resourceID, language string
		var sampleRate, maxWait, pollInterval int
		var enabled bool
		err := db.QueryRow(
			"SELECT id, provider, api_key, app_id, endpoint, resource_id, language, sample_rate, max_wait_seconds, poll_interval_seconds, enabled FROM asr_configs LIMIT 1",
		).Scan(&id, &provider, &apiKey, &appID, &endpoint, &resourceID, &language, &sampleRate, &maxWait, &pollInterval, &enabled)
		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"provider":              "",
					"api_key":               "",
					"app_id":                "",
					"endpoint":              "",
					"resource_id":           "volc.bigasr.auc",
					"language":              "zh",
					"sample_rate":           16000,
					"max_wait_seconds":      300,
					"poll_interval_seconds": 5,
					"enabled":               false,
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
				"id":                    id,
				"provider":              provider,
				"api_key":               masked,
				"app_id":                appID,
				"endpoint":              endpoint,
				"resource_id":           resourceID,
				"language":              language,
				"sample_rate":           sampleRate,
				"max_wait_seconds":      maxWait,
				"poll_interval_seconds": pollInterval,
				"enabled":               enabled,
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
			Provider            string `json:"provider"`
			APIKey              string `json:"api_key"`
			AppID               string `json:"app_id"`
			Endpoint            string `json:"endpoint"`
			ResourceID          string `json:"resource_id"`
			Language            string `json:"language"`
			SampleRate          int    `json:"sample_rate"`
			MaxWaitSeconds      int    `json:"max_wait_seconds"`
			PollIntervalSeconds int    `json:"poll_interval_seconds"`
			Enabled             bool   `json:"enabled"`
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
		if req.AppID == "" {
			respondError(w, http.StatusBadRequest, "app_id required")
			return
		}
		if req.ResourceID == "" {
			req.ResourceID = "volc.bigasr.auc"
		}
		if req.MaxWaitSeconds <= 0 {
			req.MaxWaitSeconds = 300
		}
		if req.PollIntervalSeconds <= 0 {
			req.PollIntervalSeconds = 5
		}

		_, err := db.Exec(
			`UPDATE asr_configs SET provider=?, api_key=?, app_id=?, endpoint=?, resource_id=?, language=?, sample_rate=?, max_wait_seconds=?, poll_interval_seconds=?, enabled=? WHERE id = (SELECT MIN(id) FROM asr_configs)`,
			req.Provider, req.APIKey, req.AppID, req.Endpoint, req.ResourceID, req.Language, req.SampleRate, req.MaxWaitSeconds, req.PollIntervalSeconds, req.Enabled,
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

func ASRTestHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			TestURL string `json:"test_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.TestURL == "" {
			respondError(w, http.StatusBadRequest, "test_url required")
			return
		}

		var appID, apiKey, resourceID string
		var maxWait, pollInterval int
		err := db.QueryRow(
			"SELECT app_id, api_key, resource_id, max_wait_seconds, poll_interval_seconds FROM asr_configs LIMIT 1",
		).Scan(&appID, &apiKey, &resourceID, &maxWait, &pollInterval)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("load config error: %v", err))
			return
		}
		if appID == "" || apiKey == "" {
			respondError(w, http.StatusBadRequest, "APP ID and Access Token must be configured first")
			return
		}
		if resourceID == "" {
			resourceID = "volc.bigasr.auc"
		}
		if maxWait <= 0 {
			maxWait = 300
		}
		if pollInterval <= 0 {
			pollInterval = 5
		}

		submitURL := "https://openspeech.bytedance.com/api/v3/auc/bigmodel/submit"
		queryURL := "https://openspeech.bytedance.com/api/v3/auc/bigmodel/query"
		if resourceID == "volc.bigasr.auc_idle" {
			submitURL = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/idle/submit"
			queryURL = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/idle/query"
		}

		taskID := fmt.Sprintf("nsi-test-%d", time.Now().UnixNano())

		reqBody := map[string]interface{}{
			"user":    map[string]string{"uid": appID},
			"audio":   map[string]string{"url": req.TestURL},
			"request": map[string]string{"model_name": "bigmodel"},
		}
		body, _ := json.Marshal(reqBody)

		httpReq, err := http.NewRequest("POST", submitURL, bytes.NewReader(body))
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("create request: %v", err))
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-Api-App-Key", appID)
		httpReq.Header.Set("X-Api-Access-Key", apiKey)
		httpReq.Header.Set("X-Api-Resource-Id", resourceID)
		httpReq.Header.Set("X-Api-Request-Id", taskID)
		httpReq.Header.Set("X-Api-Sequence", "-1")

		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code":  1,
				"error": fmt.Sprintf("submit failed: %v", err),
			})
			return
		}
		resp.Body.Close()

		statusCode := resp.Header.Get("X-Api-Status-Code")
		if statusCode != "20000000" && statusCode != "20000001" && statusCode != "20000002" {
			msg := resp.Header.Get("X-Api-Message")
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code":  1,
				"error": fmt.Sprintf("submit rejected: status=%s message=%s", statusCode, msg),
			})
			return
		}

		deadline := time.Now().Add(time.Duration(maxWait) * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(time.Duration(pollInterval) * time.Second)

			qReq, _ := http.NewRequest("POST", queryURL, bytes.NewReader([]byte("{}")))
			qReq.Header.Set("Content-Type", "application/json")
			qReq.Header.Set("X-Api-App-Key", appID)
			qReq.Header.Set("X-Api-Access-Key", apiKey)
			qReq.Header.Set("X-Api-Resource-Id", resourceID)
			qReq.Header.Set("X-Api-Request-Id", taskID)

			qResp, err := client.Do(qReq)
			if err != nil {
				continue
			}
			qBody, _ := io.ReadAll(qResp.Body)
			qResp.Body.Close()

			sc := qResp.Header.Get("X-Api-Status-Code")
			switch sc {
			case "20000000":
				var result struct {
					Result struct {
						Text string `json:"text"`
					} `json:"result"`
				}
				json.Unmarshal(qBody, &result)
				respondJSON(w, http.StatusOK, map[string]interface{}{
					"code": 0,
					"data": map[string]interface{}{
						"text":       result.Result.Text,
						"char_count": len(result.Result.Text),
					},
				})
				return
			case "20000001", "20000002":
				continue
			default:
				msg := qResp.Header.Get("X-Api-Message")
				respondJSON(w, http.StatusOK, map[string]interface{}{
					"code":  1,
					"error": fmt.Sprintf("query failed: status=%s message=%s", sc, msg),
				})
				return
			}
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":  1,
			"error": fmt.Sprintf("test timed out after %d seconds", maxWait),
		})
	})
}
