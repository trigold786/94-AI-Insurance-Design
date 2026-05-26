package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type PlanRepository interface {
	Save(ctx context.Context, plan *models.PlanSnapshot) error
	GetByID(ctx context.Context, planID string) (*models.PlanSnapshot, error)
}

type planRepository struct {
	db *sql.DB
}

func NewPlanRepository(db *sql.DB) (PlanRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &planRepository{db: db}, nil
}

func (r *planRepository) Save(ctx context.Context, plan *models.PlanSnapshot) error {
	schemesJSON, err := json.Marshal(plan.RecommendedSchemes)
	if err != nil {
		return fmt.Errorf("failed to marshal schemes: %w", err)
	}

	query := `
		INSERT INTO plan_snapshots (plan_id, user_id, recommended_schemes, total_cost, total_subsidy, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (plan_id) DO UPDATE SET
			recommended_schemes = EXCLUDED.recommended_schemes,
			total_cost = EXCLUDED.total_cost,
			total_subsidy = EXCLUDED.total_subsidy,
			updated_at = NOW()
	`

	_, err = r.db.ExecContext(ctx, query,
		plan.PlanID,
		plan.UserID,
		schemesJSON,
		plan.TotalCost,
		plan.TotalSubsidy,
		plan.GeneratedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}
	return nil
}

func (r *planRepository) GetByID(ctx context.Context, planID string) (*models.PlanSnapshot, error) {
	query := `
		SELECT plan_id, user_id, recommended_schemes, total_cost, total_subsidy, generated_at
		FROM plan_snapshots
		WHERE plan_id = $1
	`

	var (
		planIDOut  string
		userID     string
		schemesJSON []byte
		totalCost  float64
		totalSubsidy float64
		generatedAt time.Time
	)

	err := r.db.QueryRowContext(ctx, query, planID).Scan(
		&planIDOut, &userID, &schemesJSON, &totalCost, &totalSubsidy, &generatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	var schemes []models.Scheme
	if err := json.Unmarshal(schemesJSON, &schemes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schemes: %w", err)
	}

	return &models.PlanSnapshot{
		PlanID:             planIDOut,
		UserID:             userID,
		RecommendedSchemes: schemes,
		TotalCost:          totalCost,
		TotalSubsidy:       totalSubsidy,
		GeneratedAt:        generatedAt,
	}, nil
}
