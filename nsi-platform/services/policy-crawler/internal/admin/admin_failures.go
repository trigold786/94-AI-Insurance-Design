package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func FailureSummaryHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		summary, err := store.GetFailureSummary()
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("summary error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": summary})
	})
}

func FailureTrendHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		days := 7
		if d := r.URL.Query().Get("days"); d != "" {
			if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
			if days > 90 {
				days = 90
			}
		}
		}
		points, err := store.GetFailureTrend(days)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("trend error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": points})
	})
}

func FailureBySourceHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries, err := store.GetFailureBySource()
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("by-source error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": entries})
	})
}

func FailureTopReasonsHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := 10
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
				if limit > 100 {
					limit = 100
				}
			}
		}
		reasons, err := store.GetTopFailureReasons(limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("top-reasons error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": reasons})
	})
}

func FailureRawTextsHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sourceID := r.URL.Query().Get("source_id")
		failureType := r.URL.Query().Get("type")
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
				if limit > 500 {
					limit = 500
				}
			}
		}
		entries, err := store.GetFailedRawTexts(sourceID, failureType, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed-raw-texts error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": entries})
	})
}

func FailureRetryHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			RawTextID *int64 `json:"raw_text_id"`
			SourceID  string `json:"source_id"`
			All       bool   `json:"all"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.RawTextID != nil {
			if err := store.RetryRawText(*req.RawTextID); err != nil {
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("retry error: %v", err))
				return
			}
			respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": "retry queued"})
			return
		}

		if req.SourceID != "" && req.All {
			count, err := store.RetryAllFailed(req.SourceID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("retry all error: %v", err))
				return
			}
			respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": fmt.Sprintf("retried %d items", count)})
			return
		}

		respondError(w, http.StatusBadRequest, "provide raw_text_id or source_id+all")
	})
}
