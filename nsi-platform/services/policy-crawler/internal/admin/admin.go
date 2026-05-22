package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type ClaimStore interface {
	ListByStatus(status string) ([]models.PolicyClaim, error)
	UpdateStatus(claimID, status string, confidence float64) error
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

		claims, err := store.ListByStatus(status)
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

func UpdateClaimHandler(store ClaimStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
		claimID := parts[len(parts)-1]
		if claimID == "" || claimID == "claims" {
			respondError(w, http.StatusBadRequest, "claim ID is required")
			return
		}

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
