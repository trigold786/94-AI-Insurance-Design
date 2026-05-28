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
	structuredSchemesJSON, err := json.Marshal(plan.StructuredSchemes)
	if err != nil {
		return fmt.Errorf("failed to marshal structured_schemes: %w", err)
	}
	policyRefsJSON, err := json.Marshal(plan.PolicyReferences)
	if err != nil {
		return fmt.Errorf("failed to marshal policy_references: %w", err)
	}

	var verificationResultJSON []byte
	if plan.VerificationResult != nil {
		verificationResultJSON, err = json.Marshal(plan.VerificationResult)
		if err != nil {
			return fmt.Errorf("failed to marshal verification_result: %w", err)
		}
	}

	query := `
		INSERT INTO plan_snapshots (plan_id, user_id, policy_version_snapshot_id,
			recommended_schemes, free_form_text, structured_schemes, policy_references,
			recommendation, recommendation_reason, verification_result,
			total_cost, total_subsidy, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (plan_id) DO UPDATE SET
			policy_version_snapshot_id = EXCLUDED.policy_version_snapshot_id,
			recommended_schemes = EXCLUDED.recommended_schemes,
			free_form_text = EXCLUDED.free_form_text,
			structured_schemes = EXCLUDED.structured_schemes,
			policy_references = EXCLUDED.policy_references,
			recommendation = EXCLUDED.recommendation,
			recommendation_reason = EXCLUDED.recommendation_reason,
			verification_result = EXCLUDED.verification_result,
			total_cost = EXCLUDED.total_cost,
			total_subsidy = EXCLUDED.total_subsidy,
			updated_at = NOW()
	`

	_, err = r.db.ExecContext(ctx, query,
		plan.PlanID,
		plan.UserID,
		plan.PolicyVersionSnapshotID,
		schemesJSON,
		plan.FreeFormText,
		structuredSchemesJSON,
		policyRefsJSON,
		plan.Recommendation,
		plan.RecommendationReason,
		verificationResultJSON,
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
		SELECT plan_id, user_id, policy_version_snapshot_id,
			recommended_schemes, free_form_text, structured_schemes, policy_references,
			recommendation, recommendation_reason, verification_result,
			total_cost, total_subsidy, generated_at
		FROM plan_snapshots
		WHERE plan_id = $1
	`

	var (
		planIDOut                string
		userID                   string
		policyVersionSnapshotID  string
		schemesJSON              []byte
		freeFormText             string
		structuredSchemesJSON    []byte
		policyRefsJSON           []byte
		recommendation           string
		recommendationReason     string
		verificationResultJSON   []byte
		totalCost                float64
		totalSubsidy             float64
		generatedAt              time.Time
	)

	err := r.db.QueryRowContext(ctx, query, planID).Scan(
		&planIDOut, &userID, &policyVersionSnapshotID,
		&schemesJSON, &freeFormText, &structuredSchemesJSON, &policyRefsJSON,
		&recommendation, &recommendationReason, &verificationResultJSON,
		&totalCost, &totalSubsidy, &generatedAt,
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

	var structuredSchemes []models.LLMScheme
	if err := json.Unmarshal(structuredSchemesJSON, &structuredSchemes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal structured_schemes: %w", err)
	}

	var policyRefs []models.PolicyReference
	if err := json.Unmarshal(policyRefsJSON, &policyRefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy_references: %w", err)
	}

	var verificationResult *models.VerificationResult
	if len(verificationResultJSON) > 0 {
		verificationResult = &models.VerificationResult{}
		if err := json.Unmarshal(verificationResultJSON, verificationResult); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verification_result: %w", err)
		}
	}

	return &models.PlanSnapshot{
		PlanID:                  planIDOut,
		UserID:                  userID,
		PolicyVersionSnapshotID: policyVersionSnapshotID,
		RecommendedSchemes:      schemes,
		FreeFormText:            freeFormText,
		StructuredSchemes:       structuredSchemes,
		PolicyReferences:        policyRefs,
		Recommendation:          recommendation,
		RecommendationReason:    recommendationReason,
		VerificationResult:      verificationResult,
		TotalCost:               totalCost,
		TotalSubsidy:            totalSubsidy,
		GeneratedAt:             generatedAt,
	}, nil
}
