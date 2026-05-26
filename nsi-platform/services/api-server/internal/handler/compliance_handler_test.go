package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type mockPolicyQuerier struct {
	policies []models.PolicyClaim
	err      error
}

func (m *mockPolicyQuerier) QueryByRegionAndStatus(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.policies, nil
}

func TestComplianceEvaluator_Evaluate(t *testing.T) {
	eval := &ComplianceEvaluator{}

	t.Run("flexible employment user matches flexible tag", func(t *testing.T) {
		user := &models.UserProfile{EmploymentStatus: "flexible", Age: 35, Gender: "male"}
		policy := &models.PolicyClaim{TargetGroupTags: []string{"flexible_employment"}}
		eligible, unmet := eval.Evaluate(user, policy)
		if !eligible {
			t.Errorf("expected eligible, unmet: %v", unmet)
		}
	})

	t.Run("employed user does not match flexible tag", func(t *testing.T) {
		user := &models.UserProfile{EmploymentStatus: "employed", Age: 35}
		policy := &models.PolicyClaim{TargetGroupTags: []string{"flexible_employment"}}
		eligible, unmet := eval.Evaluate(user, policy)
		if eligible {
			t.Error("expected not eligible")
		}
		if len(unmet) == 0 {
			t.Error("expected unmet conditions")
		}
	})

	t.Run("user age 45 matches 4050 tag", func(t *testing.T) {
		user := &models.UserProfile{Age: 45, EmploymentStatus: "flexible"}
		policy := &models.PolicyClaim{TargetGroupTags: []string{"4050"}}
		eligible, _ := eval.Evaluate(user, policy)
		if !eligible {
			t.Error("expected eligible for 4050")
		}
	})

	t.Run("user age 35 does not match 4050 tag", func(t *testing.T) {
		user := &models.UserProfile{Age: 35}
		policy := &models.PolicyClaim{TargetGroupTags: []string{"4050"}}
		eligible, _ := eval.Evaluate(user, policy)
		if eligible {
			t.Error("expected not eligible for 4050")
		}
	})

	t.Run("has_children tag matches", func(t *testing.T) {
		user := &models.UserProfile{HasChildren: true}
		policy := &models.PolicyClaim{TargetGroupTags: []string{"has_children"}}
		eligible, _ := eval.Evaluate(user, policy)
		if !eligible {
			t.Error("expected eligible with children")
		}
	})

	t.Run("multiple conditions all met", func(t *testing.T) {
		user := &models.UserProfile{EmploymentStatus: "flexible", Age: 45, HasChildren: true, Gender: "male"}
		policy := &models.PolicyClaim{TargetGroupTags: []string{"flexible_employment", "4050"}}
		eligible, unmet := eval.Evaluate(user, policy)
		if !eligible {
			t.Errorf("expected all met, unmet: %v", unmet)
		}
	})

	t.Run("unknown tag is ignored", func(t *testing.T) {
		user := &models.UserProfile{EmploymentStatus: "employed"}
		policy := &models.PolicyClaim{TargetGroupTags: []string{"unknown_tag"}}
		eligible, _ := eval.Evaluate(user, policy)
		if !eligible {
			t.Error("unknown tags should be ignored")
		}
	})

	t.Run("nil user or policy returns false", func(t *testing.T) {
		eval := &ComplianceEvaluator{}
		user := &models.UserProfile{EmploymentStatus: "flexible"}
		policy := &models.PolicyClaim{TargetGroupTags: []string{"flexible_employment"}}
		eligible, _ := eval.Evaluate(nil, policy)
		if eligible {
			t.Error("nil user should not be eligible")
		}
		eligible, _ = eval.Evaluate(user, nil)
		if eligible {
			t.Error("nil policy should not be eligible")
		}
	})

	t.Run("condition with tag_match is evaluated", func(t *testing.T) {
		user := &models.UserProfile{EmploymentStatus: "employed"}
		conds, _ := json.Marshal([]models.ComplianceCondition{
			{Name: "必须为灵活就业人员", TagMatch: "flexible_employment"},
		})
		policy := &models.PolicyClaim{
			TargetGroupTags: []string{},
			Conditions:      conds,
		}
		eligible, unmet := eval.Evaluate(user, policy)
		if eligible {
			t.Error("should not be eligible with unmet condition")
		}
		if len(unmet) == 0 {
			t.Error("expected unmet condition name")
		}
	})
}

func TestComplianceChecklistHandler(t *testing.T) {
	eval := &ComplianceEvaluator{}

	t.Run("returns checklist with matched policies", func(t *testing.T) {
		policyRepo := &mockPolicyQuerier{
			policies: []models.PolicyClaim{
				{
					ClaimID:           "claim-1",
					PolicyID:          "policy-1",
					PolicyType:        "subsidy",
					RegionCode:        "310000",
					TargetGroupTags:   []string{"flexible_employment"},
					SubsidyCalcMethod: "每月补贴600元",
					Status:            "verified",
				},
			},
		}
		profileRepo := &mockProfileRepo{
			profile: &models.UserProfile{UserID: "user-1", Age: 35, EmploymentStatus: "flexible"},
		}

		handler := middleware.AuthMiddleware("")(ComplianceChecklistHandler(eval, policyRepo, profileRepo))

		req := httptest.NewRequest("GET", "/v1/compliance/checklist?city_code=310000", nil)
		req.Header.Set("x-user-id", "user-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Code int                          `json:"code"`
			Data models.ComplianceChecklist   `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Data.UserID != "user-1" {
			t.Errorf("expected user-1, got %s", resp.Data.UserID)
		}
		if len(resp.Data.MatchedPolicies) != 1 {
			t.Errorf("expected 1 policy, got %d", len(resp.Data.MatchedPolicies))
		}
		if !resp.Data.MatchedPolicies[0].IsEligible {
			t.Error("expected policy to be eligible")
		}
	})

	t.Run("returns 400 without city_code", func(t *testing.T) {
		profileRepo := &mockProfileRepo{profile: &models.UserProfile{UserID: "user-1"}}
		handler := middleware.AuthMiddleware("")(ComplianceChecklistHandler(eval, &mockPolicyQuerier{}, profileRepo))
		req := httptest.NewRequest("GET", "/v1/compliance/checklist", nil)
		req.Header.Set("x-user-id", "user-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("returns 401 without user", func(t *testing.T) {
		handler := ComplianceChecklistHandler(eval, &mockPolicyQuerier{}, &mockProfileRepo{})
		req := httptest.NewRequest("GET", "/v1/compliance/checklist?city_code=310000", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("returns 404 when profile not found", func(t *testing.T) {
		profileRepo := &mockProfileRepo{err: errors.NewNotFound("profile", "user-999")}
		handler := middleware.AuthMiddleware("")(ComplianceChecklistHandler(eval, &mockPolicyQuerier{}, profileRepo))
		req := httptest.NewRequest("GET", "/v1/compliance/checklist?city_code=310000", nil)
		req.Header.Set("x-user-id", "user-999")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("shows unmet conditions for ineligible user", func(t *testing.T) {
		policyRepo := &mockPolicyQuerier{
			policies: []models.PolicyClaim{
				{
					ClaimID:         "claim-2",
					PolicyID:        "policy-2",
					PolicyType:      "subsidy",
					RegionCode:      "310000",
					TargetGroupTags: []string{"flexible_employment"},
					Status:          "verified",
				},
			},
		}
		profileRepo := &mockProfileRepo{
			profile: &models.UserProfile{UserID: "user-2", Age: 35, EmploymentStatus: "employed"},
		}

		handler := middleware.AuthMiddleware("")(ComplianceChecklistHandler(eval, policyRepo, profileRepo))
		req := httptest.NewRequest("GET", "/v1/compliance/checklist?city_code=310000", nil)
		req.Header.Set("x-user-id", "user-2")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var resp struct {
			Data models.ComplianceChecklist `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Data.MatchedPolicies) == 0 {
			t.Fatal("expected matched policies")
		}
		if resp.Data.MatchedPolicies[0].IsEligible {
			t.Error("expected ineligible")
		}
		if len(resp.Data.MatchedPolicies[0].UnmetConditions) == 0 {
			t.Error("expected unmet conditions")
		}
	})
}
