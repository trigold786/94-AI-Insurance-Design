package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
)

type SimilarRequest struct {
	ClaimID    string `json:"claim_id"`
	Text       string `json:"text"`
	RegionCode string `json:"region"`
	PolicyType string `json:"policy_type"`
	Limit      int    `json:"limit"`
}

type EmbeddingSource interface {
	GetEmbedding(claimID string) []float64
	SearchSimilar(emb []float64, threshold float64, limit int, filter *embeddings.SearchFilter) []embeddings.SimilarResult
	SearchByText(ctx context.Context, query string, threshold float64, limit int, filter *embeddings.SearchFilter) ([]embeddings.SimilarResult, error)
}

func SimilarSearchHandler(cache EmbeddingSource) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req SimilarRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": -1, "msg": "invalid JSON"})
			return
		}
		if req.ClaimID == "" && req.Text == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": -1, "msg": "claim_id or text required"})
			return
		}
		limit := req.Limit
		if limit <= 0 || limit > 50 {
			limit = 10
		}

		var filter *embeddings.SearchFilter
		if req.RegionCode != "" || req.PolicyType != "" {
			filter = &embeddings.SearchFilter{RegionCode: req.RegionCode, PolicyType: req.PolicyType}
		}

		var results []embeddings.SimilarResult
		if req.ClaimID != "" {
			emb := cache.GetEmbedding(req.ClaimID)
			if emb != nil {
				results = cache.SearchSimilar(emb, 0, limit, filter)
			}
		}
		if req.Text != "" && (len(results) == 0 || req.ClaimID == "") {
			var err error
			results, err = cache.SearchByText(r.Context(), req.Text, 0, limit, filter)
			if err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]interface{}{"code": -1, "msg": "search failed"})
				return
			}
		}

		if results == nil {
			results = []embeddings.SimilarResult{}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": results})
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
