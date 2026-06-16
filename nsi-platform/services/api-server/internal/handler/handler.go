package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/api-server/internal/repository"
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
	respondJSON(w, http.StatusInternalServerError, map[string]interface{}{"code": "INTERNAL_ERROR", "message": "internal server error"})
}

func HealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{"status": "ok"})
	})
}

type profileInput struct {
	Age                      int      `json:"age"`
	Gender                   string   `json:"gender"`
	HouseholdRegionCode      string   `json:"household_region_code"`
	CurrentResidenceCode     string   `json:"current_residence_code"`
	EmploymentStatus         string   `json:"employment_status"`
	UnemploymentRegDate      *string  `json:"unemployment_reg_date,omitempty"`
	FlexibleEmploymentRegDate *string `json:"flexible_employment_reg_date,omitempty"`
	SocialSecurityYears      int      `json:"social_security_years"`
	SkillCertificateLevel    *string  `json:"skill_certificate_level,omitempty"`
	HasChildren              bool     `json:"has_children"`
	DateOfBirth              string   `json:"date_of_birth"`
	ContributionMonths       int      `json:"contribution_months"`
	PensionTotalAmount       float64  `json:"pension_total_amount"`
	PensionPersonalAmount    float64  `json:"pension_personal_amount"`
	IsLocalHukou             bool     `json:"is_local_hukou"`
	ChildAgeRange            string   `json:"child_age_range"`
	HasElderlyDependents     bool     `json:"has_elderly_dependents"`
	MonthlyIncome            float64  `json:"monthly_income"`
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

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var input profileInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}

		validEmployment := map[string]bool{"employed": true, "flexible": true, "self-employed": true, "unemployed": true}
		if input.EmploymentStatus != "" && !validEmployment[input.EmploymentStatus] {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid employment_status"})
			return
		}
		validGender := map[string]bool{"male": true, "female": true}
		if input.Gender != "" && !validGender[input.Gender] {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid gender"})
			return
		}
		if input.SocialSecurityYears < 0 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "social_security_years must be >= 0"})
			return
		}

		age := input.Age
		if age == 0 && input.DateOfBirth != "" {
			if parts := strings.Split(input.DateOfBirth, "-"); len(parts) == 2 {
				if by, err := strconv.Atoi(parts[0]); err == nil {
					if bm, err := strconv.Atoi(parts[1]); err == nil {
						now := time.Now()
						age = now.Year() - by
						if int(now.Month()) < bm {
							age--
						}
					}
				}
			}
		}
		profile := &models.UserProfile{
			UserID:                   userID,
			Age:                      age,
			Gender:                   input.Gender,
			HouseholdRegionCode:      input.HouseholdRegionCode,
			CurrentResidenceCode:     input.CurrentResidenceCode,
			EmploymentStatus:         input.EmploymentStatus,
			UnemploymentRegDate:      input.UnemploymentRegDate,
			FlexibleEmploymentRegDate: input.FlexibleEmploymentRegDate,
			SocialSecurityYears:      input.SocialSecurityYears,
			SkillCertificateLevel:    input.SkillCertificateLevel,
			HasChildren:              input.HasChildren,
			DateOfBirth:              input.DateOfBirth,
			ContributionMonths:       input.ContributionMonths,
			PensionTotalAmount:       input.PensionTotalAmount,
			PensionPersonalAmount:    input.PensionPersonalAmount,
			IsLocalHukou:             input.IsLocalHukou,
			ChildAgeRange:            input.ChildAgeRange,
			HasElderlyDependents:     input.HasElderlyDependents,
			MonthlyIncome:            input.MonthlyIncome,
		}

		if err := repo.Upsert(r.Context(), profile); err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": profile})
	})
}

func QueryPoliciesHandler(repo repository.PolicyRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		limit, _ := strconv.Atoi(q.Get("limit"))
		if limit <= 0 {
			limit = 50
		}
		if limit > 200 {
			limit = 200
		}
		offset, _ := strconv.Atoi(q.Get("offset"))

		filter := repository.PolicyFilter{
			RegionCode: q.Get("region_code"),
			PolicyType: q.Get("policy_type"),
			Status:     q.Get("status"),
			Keyword:    q.Get("keyword"),
			Limit:      limit,
			Offset:     offset,
		}

		claims, err := repo.Query(r.Context(), filter)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": claims})
	})
}
