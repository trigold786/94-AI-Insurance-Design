package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
	Upsert(ctx context.Context, profile *models.UserProfile) error
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		respondJSON(w, appErr.HTTPStatus, map[string]interface{}{"code": appErr.Code, "message": appErr.Message})
		return
	}
	respondJSON(w, http.StatusInternalServerError, map[string]interface{}{"code": "INTERNAL_ERROR", "message": err.Error()})
}

func HealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{"status": "ok"})
	})
}

type profileInput struct {
	Age                      int     `json:"age"`
	Gender                   string  `json:"gender"`
	HouseholdRegionCode      string  `json:"household_region_code"`
	CurrentResidenceCode     string  `json:"current_residence_code"`
	EmploymentStatus         string  `json:"employment_status"`
	UnemploymentRegDate      *string `json:"unemployment_reg_date,omitempty"`
	FlexibleEmploymentRegDate *string `json:"flexible_employment_reg_date,omitempty"`
	SocialSecurityYears      int     `json:"social_security_years"`
	SkillCertificateLevel    *string `json:"skill_certificate_level,omitempty"`
	HasChildren              bool    `json:"has_children"`
}

func GetProfileHandler(repo ProfileRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		profile, err := repo.GetByUserID(r.Context(), userID)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": profile})
	})
}

func UpdateProfileHandler(repo ProfileRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		var input profileInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}

		profile := &models.UserProfile{
			UserID:                   userID,
			Age:                      input.Age,
			Gender:                   input.Gender,
			HouseholdRegionCode:      input.HouseholdRegionCode,
			CurrentResidenceCode:     input.CurrentResidenceCode,
			EmploymentStatus:         input.EmploymentStatus,
			UnemploymentRegDate:      input.UnemploymentRegDate,
			FlexibleEmploymentRegDate: input.FlexibleEmploymentRegDate,
			SocialSecurityYears:      input.SocialSecurityYears,
			SkillCertificateLevel:    input.SkillCertificateLevel,
			HasChildren:              input.HasChildren,
		}

		if err := repo.Upsert(r.Context(), profile); err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": profile})
	})
}
