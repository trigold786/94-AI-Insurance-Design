package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type mockReportRepo struct {
	plan *models.PlanSnapshot
	err  error
}

func (m *mockReportRepo) GetByID(_ context.Context, planID string) (*models.PlanSnapshot, error) {
	if m.plan == nil && m.err == nil {
		return nil, fmt.Errorf("not found")
	}
	return m.plan, m.err
}
func (m *mockReportRepo) Save(_ context.Context, plan *models.PlanSnapshot) error {
	return nil
}
func (m *mockReportRepo) QueryByRegionAndStatus(_ context.Context, _, _ string) ([]models.PolicyClaim, error) {
	return nil, nil
}

func TestPlanReport_Success(t *testing.T) {
	repo := &mockReportRepo{
		plan: &models.PlanSnapshot{
			PlanID: "test-1",
			UserID: "user1",
			RecommendedSchemes: []models.Scheme{
				{Name: "方案A", BaseSalary: 5000, MonthlyCost: 1000, AnnualSubsidy: 3000, ProjectedPension: 2500},
			},
			TotalCost:    360000,
			TotalSubsidy: 90000,
		},
	}
	h := PlanReportHandler(repo, repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user1")
	req := httptest.NewRequest("GET", "/v1/plans/report?plan_id=test-1", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Fatalf("expected text/html, got %s", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "方案A") || !strings.Contains(body, "社保规划报告") {
		t.Fatal("HTML missing key content")
	}
}

func TestPlanReport_NotFound(t *testing.T) {
	repo := &mockReportRepo{plan: nil, err: nil}
	h := PlanReportHandler(repo, repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user1")
	req := httptest.NewRequest("GET", "/v1/plans/report?plan_id=nonexistent", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPlanReport_EmptyID(t *testing.T) {
	repo := &mockReportRepo{}
	h := PlanReportHandler(repo, repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user1")
	req := httptest.NewRequest("GET", "/v1/plans/report", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPlanReport_WrongUser(t *testing.T) {
	repo := &mockReportRepo{
		plan: &models.PlanSnapshot{PlanID: "test-1", UserID: "user2"},
	}
	h := PlanReportHandler(repo, repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user1")
	req := httptest.NewRequest("GET", "/v1/plans/report?plan_id=test-1", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for wrong user, got %d", w.Code)
	}
}


