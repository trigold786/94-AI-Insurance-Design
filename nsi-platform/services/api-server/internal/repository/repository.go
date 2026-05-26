package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
	Upsert(ctx context.Context, profile *models.UserProfile) error
}

type profileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) (ProfileRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &profileRepository{db: db}, nil
}

func (r *profileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	query := `SELECT user_id, tenant_id, age, gender, household_region_code, current_residence_code,
		employment_status, unemployment_reg_date, flexible_employment_reg_date,
		social_security_years, skill_certificate_level, has_children,
		COALESCE(date_of_birth,''), COALESCE(contribution_months,0),
		COALESCE(pension_total_amount,0), COALESCE(pension_personal_amount,0)
		FROM user_profiles WHERE user_id = $1`

	var p models.UserProfile
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&p.UserID, &p.TenantID, &p.Age, &p.Gender,
		&p.HouseholdRegionCode, &p.CurrentResidenceCode,
		&p.EmploymentStatus, &p.UnemploymentRegDate, &p.FlexibleEmploymentRegDate,
		&p.SocialSecurityYears, &p.SkillCertificateLevel, &p.HasChildren,
		&p.DateOfBirth, &p.ContributionMonths, &p.PensionTotalAmount, &p.PensionPersonalAmount,
	)
	if err == sql.ErrNoRows {
		return nil, errors.NewNotFound("user profile", userID)
	}
	if err != nil {
		return nil, errors.NewInternal("failed to get profile")
	}
	return &p, nil
}

func (r *profileRepository) Upsert(ctx context.Context, profile *models.UserProfile) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}

	query := `INSERT INTO user_profiles (user_id, tenant_id, age, gender, household_region_code,
		current_residence_code, employment_status, unemployment_reg_date,
		flexible_employment_reg_date, social_security_years, skill_certificate_level, has_children,
		date_of_birth, contribution_months, pension_total_amount, pension_personal_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (user_id) DO UPDATE SET
			age = EXCLUDED.age, gender = EXCLUDED.gender,
			household_region_code = EXCLUDED.household_region_code,
			current_residence_code = EXCLUDED.current_residence_code,
			employment_status = EXCLUDED.employment_status,
			unemployment_reg_date = EXCLUDED.unemployment_reg_date,
			flexible_employment_reg_date = EXCLUDED.flexible_employment_reg_date,
			social_security_years = EXCLUDED.social_security_years,
			skill_certificate_level = EXCLUDED.skill_certificate_level,
			has_children = EXCLUDED.has_children,
			date_of_birth = EXCLUDED.date_of_birth,
			contribution_months = EXCLUDED.contribution_months,
			pension_total_amount = EXCLUDED.pension_total_amount,
			pension_personal_amount = EXCLUDED.pension_personal_amount,
			updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query,
		profile.UserID, profile.TenantID, profile.Age, profile.Gender,
		profile.HouseholdRegionCode, profile.CurrentResidenceCode,
		profile.EmploymentStatus, profile.UnemploymentRegDate, profile.FlexibleEmploymentRegDate,
		profile.SocialSecurityYears, profile.SkillCertificateLevel, profile.HasChildren,
		profile.DateOfBirth, profile.ContributionMonths, profile.PensionTotalAmount, profile.PensionPersonalAmount,
	)
	if err != nil {
		return errors.NewInternal("failed to upsert profile")
	}
	return nil
}
