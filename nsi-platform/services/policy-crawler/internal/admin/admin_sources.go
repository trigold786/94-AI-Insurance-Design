package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SourceCRUDStore interface {
	DashboardStore
}

type CrawlTrigger interface {
	CrawlSource(sourceID string)
}

type FeedParserFunc func(data []byte) ([]FeedPreviewItem, error)

var RSSFeedParser FeedParserFunc

func SourceCreateHandler(store SourceCRUDStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var src SourceInfo
		if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if src.SourceID == "" || src.SourceName == "" {
			respondError(w, http.StatusBadRequest, "source_id and source_name required")
			return
		}
		if src.CrawlType == "" {
			src.CrawlType = "govsite"
		}
		if src.SourceLevel == "" {
			src.SourceLevel = "MEDIUM"
		}
		if src.IntervalSec == 0 {
			src.IntervalSec = 86400
		}
		if err := store.CreateSource(&src); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("create error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": src})
	})
}

func SourceDeleteHandler(store SourceCRUDStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			SourceID string `json:"source_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.SourceID == "" {
			respondError(w, http.StatusBadRequest, "source_id required")
			return
		}
		if err := store.DeleteSource(req.SourceID); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("delete error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": "deleted"})
	})
}

func SourceCrawlTriggerHandler(mgr CrawlTrigger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			SourceID string `json:"source_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.SourceID == "" {
			respondError(w, http.StatusBadRequest, "source_id required")
			return
		}
		go mgr.CrawlSource(req.SourceID)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": fmt.Sprintf("crawl triggered for %s", req.SourceID),
		})
	})
}

type FeedPreviewItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func RSSTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.URL == "" {
			respondError(w, http.StatusBadRequest, "url required")
			return
		}

		resp, err := http.DefaultClient.Get(req.URL)
		if err != nil {
			respondError(w, http.StatusBadGateway, fmt.Sprintf("fetch error: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			respondError(w, http.StatusBadGateway, fmt.Sprintf("HTTP %d", resp.StatusCode))
			return
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("read error: %v", err))
			return
		}

		if RSSFeedParser == nil {
			respondError(w, http.StatusInternalServerError, "RSS parser not configured")
			return
		}
		items, err := RSSFeedParser(body)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("parse error: %v", err))
			return
		}

		limit := 5
		if len(items) < limit {
			limit = len(items)
		}
		preview := make([]FeedPreviewItem, limit)
		for i := 0; i < limit; i++ {
			preview[i] = items[i]
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"total": len(items),
				"items": preview,
			},
		})
	})
}
