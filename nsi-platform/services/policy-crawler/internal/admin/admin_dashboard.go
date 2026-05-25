package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DashboardStore 管理后台数据接口
type DashboardStore interface {
	ClaimStore
	ListAllSources() ([]SourceInfo, error)
	UpdateSource(id string, updates map[string]interface{}) error
	CreateSource(src *SourceInfo) error
	DeleteSource(sourceID string) error
	GetCrawlLogs(limit int) ([]CrawlLogEntry, error)
	GetCrawlLogsFiltered(startDate, endDate string, limit int) ([]CrawlLogEntry, error)
	GetDashboardStats() (*DashboardStats, error)
}

type SourceInfo struct {
	SourceID     string `json:"source_id"`
	SourceName   string `json:"source_name"`
	SourceURL    string `json:"source_url"`
	SourceLevel  string `json:"source_level"`
	CrawlType    string `json:"crawl_type"`
	IntervalSec  int    `json:"interval_sec"`
	RegionCode   string `json:"region_code"`
	Enabled      bool   `json:"enabled"`
	LastCrawl    string `json:"last_crawl,omitempty"`
	LastStatus   string `json:"last_status,omitempty"`
	ClaimsCount  int    `json:"claims_count"`
}

type CrawlLogEntry struct {
	ID           int    `json:"id"`
	SourceID     string `json:"source_id"`
	SourceName   string `json:"source_name"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
	CrawledAt    string `json:"crawled_at"`
}

type DashboardStats struct {
	TotalSources      int            `json:"total_sources"`
	ActiveSources     int            `json:"active_sources"`
	HighSources       int            `json:"high_sources"`
	MediumSources     int            `json:"medium_sources"`
	LowSources        int            `json:"low_sources"`
	TotalClaims       int            `json:"total_claims"`
	VerifiedClaims    int            `json:"verified_claims"`
	PendingClaims     int            `json:"pending_claims"`
	UnverifiedClaims  int            `json:"unverified_claims"`
	TodayCrawls       int            `json:"today_crawls"`
	FailedCrawls      int            `json:"failed_crawls"`
	WithEmbedding     int            `json:"with_embedding"`
	WithPolicyURL     int            `json:"with_policy_url"`
	PendingExtraction  int            `json:"pending_extraction"`
	PolicyTypeDist    map[string]int `json:"policy_type_dist"`
	RegionDist        map[string]int `json:"region_dist"`
	CrawlTrend7d      []int          `json:"crawl_trend_7d"`
	ExtractSuccessRate float64       `json:"extract_success_rate"`
}

func DashboardHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats, err := store.GetDashboardStats()
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("stats error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": stats})
	})
}

func SourceListHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sources, err := store.ListAllSources()
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("sources error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": sources})
	})
}

func SourceUpdateHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			SourceID    string  `json:"source_id"`
			Enabled     *bool   `json:"enabled,omitempty"`
			IntervalSec *int    `json:"interval_sec,omitempty"`
			SourceName  *string `json:"source_name,omitempty"`
			SourceURL   *string `json:"source_url,omitempty"`
			SourceLevel *string `json:"source_level,omitempty"`
			CrawlType   *string `json:"crawl_type,omitempty"`
			RegionCode  *string `json:"region_code,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.SourceID == "" {
			respondError(w, http.StatusBadRequest, "source_id required")
			return
		}
		updates := make(map[string]interface{})
		if req.Enabled != nil {
			updates["enabled"] = *req.Enabled
		}
		if req.IntervalSec != nil {
			updates["interval_sec"] = *req.IntervalSec
		}
		if req.SourceName != nil {
			updates["source_name"] = *req.SourceName
		}
		if req.SourceURL != nil {
			updates["source_url"] = *req.SourceURL
		}
		if req.SourceLevel != nil {
			updates["source_level"] = *req.SourceLevel
		}
		if req.CrawlType != nil {
			updates["crawl_type"] = *req.CrawlType
		}
		if req.RegionCode != nil {
			updates["region_code"] = *req.RegionCode
		}
		if store == nil {
			respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": "updated (no store)"})
			return
		}
		if err := store.UpdateSource(req.SourceID, updates); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("update error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": "updated"})
	})
}

func CrawlLogsHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")
		limit := 5000 // 最多返回 5000 条
		logs, err := store.GetCrawlLogsFiltered(startDate, endDate, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("logs error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": logs})
	})
}


