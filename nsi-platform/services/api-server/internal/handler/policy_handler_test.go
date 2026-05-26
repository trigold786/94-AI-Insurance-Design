package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/api-server/internal/repository"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type mockPolicyRepo struct {
	claims []models.PolicyClaim
	err    error
}

func (m *mockPolicyRepo) Query(ctx context.Context, filter repository.PolicyFilter) ([]models.PolicyClaim, error) {
	return m.claims, m.err
}

func (m *mockPolicyRepo) GetByID(ctx context.Context, policyID string) (*models.PolicyClaim, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.claims) > 0 {
		return &m.claims[0], nil
	}
	return nil, errors.NewNotFound("policy", policyID)
}

func (m *mockPolicyRepo) QueryByRegionAndStatus(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error) {
	if m.err != nil {
		return nil, m.err
	}
	var filtered []models.PolicyClaim
	for _, c := range m.claims {
		if c.RegionCode == regionCode && c.Status == status {
			filtered = append(filtered, c)
		}
	}
	if filtered == nil {
		filtered = []models.PolicyClaim{}
	}
	return filtered, nil
}

func TestQueryPoliciesHandlerSuccess(t *testing.T) {
	repo := &mockPolicyRepo{
		claims: []models.PolicyClaim{
			{ClaimID: "CLM-001", PolicyID: "SH-2025-001", RegionCode: "310000", PolicyType: "subsidy", Status: "verified"},
		},
	}
	handler := QueryPoliciesHandler(repo)

	req := httptest.NewRequest("GET", "/v1/policies?region_code=310000&policy_type=subsidy", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "SH-2025-001") {
		t.Errorf("expected policy in response, got %s", w.Body.String())
	}
}

func TestQueryPoliciesHandlerNoResults(t *testing.T) {
	repo := &mockPolicyRepo{claims: []models.PolicyClaim{}}
	handler := QueryPoliciesHandler(repo)

	req := httptest.NewRequest("GET", "/v1/policies?region_code=999999", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestQueryPoliciesHandlerError(t *testing.T) {
	repo := &mockPolicyRepo{err: errors.NewInternal("db error")}
	handler := QueryPoliciesHandler(repo)

	req := httptest.NewRequest("GET", "/v1/policies", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestQueryPoliciesHandlerAllFilters(t *testing.T) {
	repo := &mockPolicyRepo{
		claims: []models.PolicyClaim{
			{ClaimID: "CLM-002", PolicyID: "BJ-2025-001", RegionCode: "110000", PolicyType: "pension", Status: "verified"},
		},
	}
	handler := QueryPoliciesHandler(repo)

	req := httptest.NewRequest("GET", "/v1/policies?region_code=110000&policy_type=pension&status=verified&keyword=BJ", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
