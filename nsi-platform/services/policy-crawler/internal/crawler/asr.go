package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type ASRConfig struct {
	ID                 int64  `json:"id"`
	Provider           string `json:"provider"`
	APIKey             string `json:"api_key"`
	AppID              string `json:"app_id"`
	Endpoint           string `json:"endpoint"`
	ResourceID         string `json:"resource_id"`
	Language           string `json:"language"`
	SampleRate         int    `json:"sample_rate"`
	MaxWaitSeconds     int    `json:"max_wait_seconds"`
	PollIntervalSeconds int   `json:"poll_interval_seconds"`
	Enabled            bool   `json:"enabled"`
}

type ASRProvider interface {
	Transcribe(audioPath string, audioURL string) (string, error)
}

type VolcengineASR struct {
	config ASRConfig
	client *http.Client
}

func NewVolcengineASR(cfg ASRConfig) *VolcengineASR {
	maxWait := cfg.MaxWaitSeconds
	if maxWait <= 0 {
		maxWait = 300
	}
	pollInterval := cfg.PollIntervalSeconds
	if pollInterval <= 0 {
		pollInterval = 5
	}
	cfg.MaxWaitSeconds = maxWait
	cfg.PollIntervalSeconds = pollInterval
	return &VolcengineASR{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (v *VolcengineASR) Transcribe(audioPath string, audioURL string) (string, error) {
	if audioURL == "" {
		url, err := v.getDirectAudioURL(audioPath)
		if err != nil {
			return "", fmt.Errorf("no audio URL available and cannot extract from yt-dlp: %w", err)
		}
		audioURL = url
	}

	submitURL := "https://openspeech.bytedance.com/api/v3/auc/bigmodel/submit"
	resourceID := v.config.ResourceID
	if resourceID == "" {
		resourceID = "volc.bigasr.auc"
	}
	if strings.Contains(resourceID, "idle") {
		submitURL = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/idle/submit"
	}

	taskID := fmt.Sprintf("nsi-asr-%d", time.Now().UnixNano())

	taskID, err := v.submitTask(submitURL, audioURL, taskID, resourceID)
	if err != nil {
		return "", fmt.Errorf("submit ASR task: %w", err)
	}
	log.Printf("[asr] submitted task %s for url=%s (resource=%s)", taskID, truncateForLog(audioURL, 80), resourceID)

	queryURL := "https://openspeech.bytedance.com/api/v3/auc/bigmodel/query"
	if strings.Contains(resourceID, "idle") {
		queryURL = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/idle/query"
	}

	text, err := v.pollResult(queryURL, taskID, resourceID)
	if err != nil {
		return "", fmt.Errorf("poll ASR result: %w", err)
	}
	if text == "" {
		return "", fmt.Errorf("ASR returned empty text")
	}
	return text, nil
}

func (v *VolcengineASR) submitTask(submitURL, audioURL, taskID, resourceID string) (string, error) {
	reqBody := map[string]interface{}{
		"user":    map[string]string{"uid": v.config.AppID},
		"audio":   map[string]string{"url": audioURL},
		"request": map[string]string{"model_name": "bigmodel"},
	}
	body, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequest("POST", submitURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create submit request: %w", err)
	}
	v.setHeaders(httpReq, taskID, resourceID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("submit API call: %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.Header.Get("X-Api-Status-Code")
	if statusCode != "20000000" && statusCode != "20000001" && statusCode != "20000002" {
		respBody, _ := io.ReadAll(resp.Body)
		msg := resp.Header.Get("X-Api-Message")
		return "", fmt.Errorf("submit failed: status=%s message=%s body=%s", statusCode, msg, truncateBytes(respBody, 200))
	}

	return taskID, nil
}

func (v *VolcengineASR) pollResult(queryURL, taskID, resourceID string) (string, error) {
	deadline := time.Now().Add(time.Duration(v.config.MaxWaitSeconds) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(time.Duration(v.config.PollIntervalSeconds) * time.Second)

		httpReq, err := http.NewRequest("POST", queryURL, bytes.NewReader([]byte("{}")))
		if err != nil {
			return "", fmt.Errorf("create query request: %w", err)
		}
		v.setHeaders(httpReq, taskID, resourceID)
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := v.client.Do(httpReq)
		if err != nil {
			log.Printf("[asr] query error for task %s: %v", taskID, err)
			continue
		}

		statusCode := resp.Header.Get("X-Api-Status-Code")
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		switch statusCode {
		case "20000000":
			var result struct {
				Result struct {
					Text string `json:"text"`
				} `json:"result"`
			}
			if err := json.Unmarshal(respBody, &result); err != nil {
				return "", fmt.Errorf("parse ASR response: %w (body=%s)", err, truncateBytes(respBody, 200))
			}
			return result.Result.Text, nil
		case "20000001", "20000002":
			log.Printf("[asr] task %s still processing (status=%s)", taskID, statusCode)
			continue
		default:
			msg := resp.Header.Get("X-Api-Message")
			return "", fmt.Errorf("ASR query failed: status=%s message=%s", statusCode, msg)
		}
	}
	return "", fmt.Errorf("ASR task %s timed out after %d seconds", taskID, v.config.MaxWaitSeconds)
}

func (v *VolcengineASR) setHeaders(req *http.Request, taskID, resourceID string) {
	req.Header.Set("X-Api-App-Key", v.config.AppID)
	req.Header.Set("X-Api-Access-Key", v.config.APIKey)
	req.Header.Set("X-Api-Resource-Id", resourceID)
	req.Header.Set("X-Api-Request-Id", taskID)
	req.Header.Set("X-Api-Sequence", "-1")
}

func (v *VolcengineASR) getDirectAudioURL(audioPath string) (string, error) {
	return "", fmt.Errorf("audio URL extraction not supported for local files, need yt-dlp video URL")
}

func GetDirectAudioURLFromVideo(videoURL string) (string, error) {
	cmd := exec.Command("yt-dlp", "-g", "-f", "bestaudio", videoURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp get-url: %w: %s", err, string(out))
	}
	url := strings.TrimSpace(string(out))
	if url == "" {
		return "", fmt.Errorf("yt-dlp returned empty URL")
	}
	return url, nil
}

func NewASRProviderFromConfig(cfg ASRConfig) ASRProvider {
	switch cfg.Provider {
	case "volcengine":
		return NewVolcengineASR(cfg)
	default:
		log.Printf("[asr] unknown provider %q, defaulting to volcengine", cfg.Provider)
		return NewVolcengineASR(cfg)
	}
}

func truncateBytes(data []byte, maxLen int) string {
	if len(data) <= maxLen {
		return string(data)
	}
	return string(data[:maxLen]) + "..."
}

func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
