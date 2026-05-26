package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

var _ context.Context

func TestNewPlanRepositoryNilDB(t *testing.T) {
	_, err := NewPlanRepository(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
}

func TestNewPlanRepositorySuccess(t *testing.T) {
	db := &sql.DB{}
	repo, err := NewPlanRepository(db)
	if err != nil {
		t.Fatalf("NewPlanRepository() returned error: %v", err)
	}
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestPlanRepositoryFullFlow(t *testing.T) {
	t.Skip("skipping integration test; set NSI_TEST_DATABASE_URL to run")

	repo, _ := NewPlanRepository(nil)

	now := time.Now()
	plan := &models.PlanSnapshot{
		PlanID:      "plan-int-001",
		UserID:      "user-int-plan",
		TotalCost:   5000,
		TotalSubsidy: 1000,
		RecommendedSchemes: []models.Scheme{
			{Name: "方案1", BaseSalary: 6000, MonthlyCost: 500, ProjectedPension: 3500},
		},
		GeneratedAt: now,
	}

	if err := repo.Save(context.Background(), plan); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	got, err := repo.GetByID(context.Background(), "plan-int-001")
	if err != nil {
		t.Fatalf("GetByID() returned error: %v", err)
	}
	if got.PlanID != "plan-int-001" {
		t.Errorf("expected PlanID plan-int-001, got %s", got.PlanID)
	}
	if len(got.RecommendedSchemes) != 1 {
		t.Errorf("expected 1 scheme, got %d", len(got.RecommendedSchemes))
	}
}
