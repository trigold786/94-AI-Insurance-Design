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

var errPlanNotFound = fmt.Errorf("plan not found")
var errDB = fmt.Errorf("database error")

func TestPlanDetailHandlerSuccess(t *testing.T) {
	repo := &mockPlanRepo{
		savedPlan: &models.PlanSnapshot{
			PlanID: "plan-001",
			UserID: "user-1",
		},
	}
	handler := PlanDetailHandler(repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-1")
	req := httptest.NewRequest("GET", "/v1/plans/plan-001", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "plan-001") {
		t.Errorf("expected plan-001 in response, got %s", w.Body.String())
	}
}

func TestPlanDetailHandlerNotFound(t *testing.T) {
	repo := &mockPlanRepo{err: errPlanNotFound}
	handler := PlanDetailHandler(repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-1")
	req := httptest.NewRequest("GET", "/v1/plans/plan-nonexistent", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestPlanDetailHandlerRepoError(t *testing.T) {
	repo := &mockPlanRepo{err: errDB}
	handler := PlanDetailHandler(repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-1")
	req := httptest.NewRequest("GET", "/v1/plans/plan-001", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestPlanDetailHandlerWrongUser(t *testing.T) {
	repo := &mockPlanRepo{
		savedPlan: &models.PlanSnapshot{
			PlanID: "plan-001",
			UserID: "user-owner",
		},
	}
	handler := PlanDetailHandler(repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-attacker")
	req := httptest.NewRequest("GET", "/v1/plans/plan-001", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for wrong user, got %d", w.Code)
	}
}

func TestPlanDetailHandlerEmptyID(t *testing.T) {
	repo := &mockPlanRepo{}
	handler := PlanDetailHandler(repo)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-1")
	req := httptest.NewRequest("GET", "/v1/plans/", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
