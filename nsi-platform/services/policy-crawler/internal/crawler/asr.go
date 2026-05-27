package crawler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ASRConfig struct {
	ID         int64  `json:"id"`
	Provider   string `json:"provider"`
	APIKey     string `json:"api_key"`
	Endpoint   string `json:"endpoint"`
	Language   string `json:"language"`
	SampleRate int    `json:"sample_rate"`
	Enabled    bool   `json:"enabled"`
}

type ASRProvider interface {
	Transcribe(audioPath string) (string, error)
}

type VolcengineASR struct {
	config ASRConfig
	client *http.Client
}

func NewVolcengineASR(cfg ASRConfig) *VolcengineASR {
	return &VolcengineASR{
		config: cfg,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (v *VolcengineASR) Transcribe(audioPath string) (string, error) {
	data, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("read audio: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	reqBody := map[string]interface{}{
		"audio":       encoded,
		"language":    v.config.Language,
		"sample_rate": v.config.SampleRate,
		"format":      detectAudioFormat(audioPath),
	}
	body, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequest("POST", v.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+v.config.APIKey)
	resp, err := v.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ASR API call: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ASR API %d: %s", resp.StatusCode, string(respBody)[:minInt(len(respBody), 200)])
	}
	var result struct {
		Data struct {
			Text string `json:"text"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse ASR response: %w", err)
	}
	if result.Data.Text == "" {
		return "", fmt.Errorf("ASR returned empty text")
	}
	return result.Data.Text, nil
}

func detectAudioFormat(path string) string {
	if len(path) > 4 {
		ext := path[len(path)-4:]
		switch ext {
		case ".mp3":
			return "mp3"
		case ".wav":
			return "wav"
		case ".m4a":
			return "m4a"
		}
	}
	return "mp3"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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
