package handler

import (
	"net/http"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type VersionLister interface {
	ListVersions(policyID string) ([]models.VersionSnapshot, error)
	GetVersionAtTime(policyID string, timestamp string) (*models.VersionSnapshot, error)
}

func VersionsHandler(store VersionLister) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		policyID := r.URL.Query().Get("policy_id")
		if policyID == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": -1, "msg": "policy_id query param required"})
			return
		}

		if ts := r.URL.Query().Get("as_of"); ts != "" {
			vs, err := store.GetVersionAtTime(policyID, ts)
			if err != nil {
				respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": nil, "msg": "no version found at that time"})
				return
			}
			respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": vs})
			return
		}

		versions, err := store.ListVersions(policyID)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{"code": -1, "msg": err.Error()})
			return
		}
		if versions == nil {
			versions = []models.VersionSnapshot{}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": versions})
	})
}
