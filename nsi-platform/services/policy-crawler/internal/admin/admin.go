package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/parser"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

var validStatuses = map[string]bool{
	"": true, "verified": true, "pending_review": true, "unverified": true,
}

// PendingRawText 待提取原始文本摘要
type PendingRawText struct {
	ID         int64  `json:"id"`
	SourceID   string `json:"source_id"`
	SourceName string `json:"source_name"`
	Title      string `json:"title"`
	SourceURL  string `json:"source_url"`
	FetchedAt  string `json:"fetched_at"`
	CrawlType  string `json:"crawl_type"`
	SourceLevel string `json:"source_level"`
	RegionCode string `json:"region_code"`
}

// ExtProgress 提取进度
type ExtProgress struct {
	mu         sync.Mutex
	Total      int    `json:"total"`
	Completed  int    `json:"completed"`
	Failed     int    `json:"failed"`
	Running    bool   `json:"running"`
	Done       bool   `json:"done"`
	CurrentID  int64  `json:"current_id"`
	CurrentSrc string `json:"current_src"`
}

// GlobalExtProgress 全局提取进度（用于异步提取）
var GlobalExtProgress = &ExtProgress{}

func (p *ExtProgress) Lock()   { p.mu.Lock() }
func (p *ExtProgress) Unlock() { p.mu.Unlock() }

type ClaimStore interface {
	ListByStatus(status string, regionCode string, sourceID string, policyType string, sourceLevel string) ([]models.PolicyClaim, error)
	UpdateStatus(claimID, status string, confidence float64) error
	Ingest(claim *models.PolicyClaim) error
	GetClaimByID(claimID string) (*models.PolicyClaim, error)
	SearchSimilarClaims(claimID string, limit int) ([]models.PolicyClaim, error)
	UpdateClaimFields(claimID string, fields map[string]interface{}) error
}

func SourceImportHandler(store SourceImportStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		var req struct {
			SourceID  string `json:"source_id"`
			Title     string `json:"title"`
			Content   string `json:"content"`
			SourceURL string `json:"source_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.SourceID == "" || req.Title == "" || req.Content == "" {
			respondError(w, http.StatusBadRequest, "source_id, title, and content are required")
			return
		}

		src, err := store.GetSourceByID(req.SourceID)
		if err != nil || src == nil {
			respondError(w, http.StatusNotFound, "source not found")
			return
		}

		now := time.Now()
		claim := &models.PolicyClaim{
			ClaimID:         fmt.Sprintf("MANUAL-%d", now.UnixNano()),
			PolicyID:        req.Title,
			RegionCode:      src.RegionCode,
			PolicyType:      "subsidy",
			TargetGroupTags:  []string{},
			SubsidyCalcMethod: "参见原文",
			EffectiveDate:   now.Format("2006-01-02"),
			ConfidenceScore: 0.3,
			Status:          "unverified",
			VersionNumber:   1,
			Conditions:      []byte("[]"),
			RequiredDocuments: []byte("[]"),
			SourceID:        req.SourceID,
			SourceName:      src.SourceName,
			SourceURL:       req.SourceURL,
		}

		if err := store.SaveRawText(req.SourceID, req.Title, req.Content, req.SourceURL, ""); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("save raw text: %v", err))
			return
		}

		if err := store.Ingest(claim); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("ingest: %v", err))
			return
		}

		store.MarkRawTextExtractedByTitle(req.SourceID, req.Title)

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "导入成功，待人工审核",
			"data":    claim,
		})
	})
}

type SourceImportStore interface {
	ClaimStore
	GetSourceByID(sourceID string) (*SourceInfo, error)
	SaveRawText(sourceID, title, content, sourceURL, versionHash string) error
	MarkRawTextExtractedByTitle(sourceID, title string)
}

type updateRequest struct {
	Status          string  `json:"status"`
	ConfidenceScore float64 `json:"confidence_score"`
}

func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, code int, msg string) {
	respondJSON(w, code, map[string]string{"error": msg})
}

func ListClaimsHandler(store ClaimStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		regionCode := r.URL.Query().Get("region_code")
		sourceID := r.URL.Query().Get("source_id")
		policyType := r.URL.Query().Get("policy_type")
		sourceLevel := r.URL.Query().Get("source_level")

		validStatuses := map[string]bool{
			"": true, "verified": true, "pending_review": true, "unverified": true,
		}
		if !validStatuses[status] {
			respondError(w, http.StatusBadRequest, "invalid status: must be verified, pending_review, or unverified")
			return
		}

		if store == nil {
			respondJSON(w, http.StatusOK, map[string]interface{}{"claims": []interface{}{}})
			return
		}

		claims, err := store.ListByStatus(status, regionCode, sourceID, policyType, sourceLevel)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list claims")
			return
		}
		if claims == nil {
			claims = []models.PolicyClaim{}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"claims": claims})
	})
}

func IngestPolicyHandler(store ClaimStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Text == "" {
			respondError(w, http.StatusBadRequest, "text is required")
			return
		}

		parsed, conditions, docs, err := parser.ParseStructuredText(req.Text)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("parse error: %v", err))
			return
		}

		condJSON, _ := json.Marshal(conditions)
		docJSON, _ := json.Marshal(docs)

		now := time.Now()
		claim := &models.PolicyClaim{
			ClaimID:          fmt.Sprintf("IMP-%d", now.UnixNano()),
			PolicyID:         parsed.PolicyID,
			RegionCode:       parsed.RegionCode,
			PolicyType:       parsed.PolicyType,
			TargetGroupTags:  parsed.TargetGroups,
			SubsidyCalcMethod: parsed.SubsidyCalcMethod,
			SubsidyAmountMin: parsed.AmountMin,
			SubsidyAmountMax: parsed.AmountMax,
			SubsidyDuration:  parsed.SubsidyDuration,
			EffectiveDate:    parsed.EffectiveDate,
			ExpireDate:       parsed.ExpireDate,
			ConfidenceScore:  0.7,
			Status:           "pending_review",
			VersionNumber:    1,
			Conditions:       condJSON,
			RequiredDocuments: docJSON,
		}

		if store != nil {
			if err := store.Ingest(claim); err != nil {
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("store error: %v", err))
				return
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "政策导入成功，待审核",
			"data":    claim,
		})
	})
}

func BatchUpdateHandler(store ClaimStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			ClaimIDs []string `json:"claim_ids"`
			Status   string   `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		validStatuses := map[string]bool{"verified": true, "unverified": true, "pending_review": true}
		if !validStatuses[req.Status] {
			respondError(w, http.StatusBadRequest, "invalid status")
			return
		}
		var succeeded, failed int
		var lastErr string
		for _, id := range req.ClaimIDs {
			if store != nil {
				if err := store.UpdateStatus(id, req.Status, 1.0); err != nil {
					failed++
					lastErr = err.Error()
				} else {
					succeeded++
				}
			}
		}
		if failed > 0 {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code": 0, "succeeded": succeeded, "failed": failed,
				"message": fmt.Sprintf("batch updated: %d succeeded, %d failed", succeeded, failed),
				"last_error": lastErr,
			})
		} else {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code": 0, "succeeded": succeeded, "message": fmt.Sprintf("batch updated %d claims", succeeded),
			})
		}
	})
}

func UpdateClaimHandler(store ClaimStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimRight(r.URL.Path, "/")
		if strings.HasSuffix(path, "/compare") {
			ClaimCompareHandler(store).ServeHTTP(w, r)
			return
		}

		parts := strings.Split(path, "/")
		claimID := parts[len(parts)-1]
		if claimID == "" || claimID == "claims" {
			respondError(w, http.StatusBadRequest, "claim ID is required")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req updateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		validStatuses := map[string]bool{"verified": true, "unverified": true, "pending_review": true}
		if !validStatuses[req.Status] {
			respondError(w, http.StatusBadRequest, "invalid status: must be verified, pending_review, or unverified")
			return
		}

		if store != nil {
			if err := store.UpdateStatus(claimID, req.Status, req.ConfidenceScore); err != nil {
				respondError(w, http.StatusInternalServerError, "failed to update claim")
				return
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "claim updated",
		})
	})
}

type FieldDiff struct {
	Field      string  `json:"field"`
	CurrentVal string  `json:"current_val"`
	OtherVal   string  `json:"other_val"`
	DiffScore  float64 `json:"diff_score"`
	ClaimID    string  `json:"claim_id"`
}

func ClaimCompareHandler(store ClaimStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
		claimID := parts[len(parts)-1]
		if claimID == "" {
			respondError(w, http.StatusBadRequest, "claim ID is required")
			return
		}

		claim, err := store.GetClaimByID(claimID)
		if err != nil || claim == nil {
			respondError(w, http.StatusNotFound, "claim not found")
			return
		}

		similar, err := store.SearchSimilarClaims(claimID, 10)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to search similar")
			return
		}

		var diffs []FieldDiff
		compareFields := []string{
			"subsidy_calc_method", "effective_date", "policy_type",
			"region_code", "policy_id",
		}
		if claim.SubsidyAmountMin != nil {
			compareFields = append(compareFields, "subsidy_amount_min")
		}
		if claim.SubsidyAmountMax != nil {
			compareFields = append(compareFields, "subsidy_amount_max")
		}

		for _, other := range similar {
			if other.ClaimID == claimID {
				continue
			}
			for _, field := range compareFields {
				curVal, otherVal := getClaimField(claim, field), getClaimField(&other, field)
				score := fieldDiffScore(curVal, otherVal)
				if score > 0 {
					diffs = append(diffs, FieldDiff{
						Field:      field,
						CurrentVal: curVal,
						OtherVal:   otherVal,
						DiffScore:  score,
						ClaimID:    other.ClaimID,
					})
				}
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":      0,
			"claim":     claim,
			"similar":   similar,
			"diffs":     diffs,
		})
	})
}

func getClaimField(c *models.PolicyClaim, field string) string {
	switch field {
	case "subsidy_calc_method":
		return c.SubsidyCalcMethod
	case "effective_date":
		return c.EffectiveDate
	case "policy_type":
		return c.PolicyType
	case "region_code":
		return c.RegionCode
	case "policy_id":
		return c.PolicyID
	case "subsidy_amount_min":
		if c.SubsidyAmountMin != nil {
			return fmt.Sprintf("%.2f", *c.SubsidyAmountMin)
		}
	case "subsidy_amount_max":
		if c.SubsidyAmountMax != nil {
			return fmt.Sprintf("%.2f", *c.SubsidyAmountMax)
		}
	}
	return ""
}

func fieldDiffScore(a, b string) float64 {
	if a == b {
		return 0
	}
	if a == "" || b == "" {
		return 1.0
	}
	return 1.0
}
