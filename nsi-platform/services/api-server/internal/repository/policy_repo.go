package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type PolicyFilter struct {
	RegionCode string
	PolicyType string
	Status     string
	Keyword    string
	Limit      int
	Offset     int
}

type PolicyRepository interface {
	Query(ctx context.Context, filter PolicyFilter) ([]models.PolicyClaim, error)
	GetByID(ctx context.Context, policyID string) (*models.PolicyClaim, error)
	QueryByRegionAndStatus(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error)
	QueryByRegionHierarchy(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error)
}

type policyRepository struct {
	db *sql.DB
}

func NewPolicyRepository(db *sql.DB) (PolicyRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &policyRepository{db: db}, nil
}

const policyColumns = `claim_id, policy_id, region_code, policy_type, target_group_tags,
	subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
	effective_date, expire_date, confidence_score, status, version_number,
	conditions, required_documents, source_id, source_name, source_url,
	policy_url, policy_title, issuing_authority, document_number, application_process`

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanPolicyClaim(s scanner) (models.PolicyClaim, error) {
	var c models.PolicyClaim
	tags := []string{}
	var conditions, reqDocs, appProcess []byte
	err := s.Scan(
		&c.ClaimID, &c.PolicyID, &c.RegionCode, &c.PolicyType, pq.Array(&tags),
		&c.SubsidyCalcMethod, &c.SubsidyAmountMin, &c.SubsidyAmountMax,
		&c.SubsidyDuration, &c.EffectiveDate, &c.ExpireDate,
		&c.ConfidenceScore, &c.Status, &c.VersionNumber,
		&conditions, &reqDocs,
		&c.SourceID, &c.SourceName, &c.SourceURL,
		&c.PolicyURL, &c.PolicyTitle, &c.IssuingAuthority,
		&c.DocumentNumber, &appProcess,
	)
	if err != nil {
		return c, err
	}
	if len(conditions) > 0 {
		c.Conditions = json.RawMessage(conditions)
	}
	if len(reqDocs) > 0 {
		c.RequiredDocuments = json.RawMessage(reqDocs)
	}
	if len(appProcess) > 0 {
		c.ApplicationProcess = json.RawMessage(appProcess)
	}
	c.TargetGroupTags = tags
	return c, nil
}

func (r *policyRepository) Query(ctx context.Context, filter PolicyFilter) ([]models.PolicyClaim, error) {
	query := `SELECT ` + policyColumns + ` FROM policy_claims WHERE 1=1`

	var args []interface{}
	argIdx := 1

	if filter.RegionCode != "" {
		query += fmt.Sprintf(" AND region_code = $%d", argIdx)
		args = append(args, filter.RegionCode)
		argIdx++
	}
	if filter.PolicyType != "" {
		query += fmt.Sprintf(" AND policy_type = $%d", argIdx)
		args = append(args, filter.PolicyType)
		argIdx++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Keyword != "" {
		query += fmt.Sprintf(" AND (policy_id ILIKE $%d OR claim_id ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Keyword+"%")
		argIdx++
	}

	query += " ORDER BY updated_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewInternal("failed to query policies")
	}
	defer rows.Close()

	var claims []models.PolicyClaim
	for rows.Next() {
		c, err := scanPolicyClaim(rows)
		if err != nil {
			return nil, errors.NewInternal("failed to scan policy")
		}
		claims = append(claims, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewInternal("rows iteration error")
	}

	if claims == nil {
		claims = []models.PolicyClaim{}
	}

	return claims, nil
}

func (r *policyRepository) QueryByRegionAndStatus(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error) {
	query := `SELECT ` + policyColumns + ` FROM policy_claims WHERE region_code = $1 AND status = $2
		ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, regionCode, status)
	if err != nil {
		return nil, errors.NewInternal("failed to query policies")
	}
	defer rows.Close()

	var claims []models.PolicyClaim
	for rows.Next() {
		c, err := scanPolicyClaim(rows)
		if err != nil {
			return nil, errors.NewInternal("failed to scan policy")
		}
		claims = append(claims, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewInternal("rows iteration error")
	}

	if claims == nil {
		claims = []models.PolicyClaim{}
	}

	return claims, nil
}

func (r *policyRepository) QueryByRegionHierarchy(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error) {
	codes := buildRegionHierarchy(regionCode)

	query := `SELECT ` + policyColumns + ` FROM policy_claims
		WHERE region_code = ANY($1) AND status = $2 AND (expire_date IS NULL OR expire_date > NOW())
		ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(codes), status)
	if err != nil {
		return nil, errors.NewInternal("failed to query policies")
	}
	defer rows.Close()

	var claims []models.PolicyClaim
	for rows.Next() {
		c, err := scanPolicyClaim(rows)
		if err != nil {
			return nil, errors.NewInternal("failed to scan policy")
		}
		claims = append(claims, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewInternal("rows iteration error")
	}

	if claims == nil {
		claims = []models.PolicyClaim{}
	}

	return claims, nil
}

func buildRegionHierarchy(code string) []string {
	codes := []string{"000000"}
	if len(code) >= 2 {
		codes = append(codes, code[:2]+"0000")
	}
	if len(code) >= 4 {
		codes = append(codes, code[:4]+"00")
	}
	if len(code) >= 6 {
		codes = append(codes, code)
	}
	return codes
}

func (r *policyRepository) GetByID(ctx context.Context, policyID string) (*models.PolicyClaim, error) {
	if policyID == "" {
		return nil, fmt.Errorf("policyID cannot be empty")
	}

	query := `SELECT ` + policyColumns + ` FROM policy_claims WHERE policy_id = $1`

	c, err := scanPolicyClaim(r.db.QueryRowContext(ctx, query, policyID))
	if err == sql.ErrNoRows {
		return nil, errors.NewNotFound("policy claim", policyID)
	}
	if err != nil {
		return nil, errors.NewInternal("failed to get policy")
	}
	return &c, nil
}
