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
		COALESCE(is_local_hukou,false), employment_status, unemployment_reg_date, flexible_employment_reg_date,
		social_security_years, skill_certificate_level, has_children,
		COALESCE(child_age_range,''), COALESCE(has_elderly_dependents,false),
		COALESCE(date_of_birth,''), COALESCE(contribution_months,0),
		COALESCE(pension_total_amount,0), COALESCE(pension_personal_amount,0),
		COALESCE(monthly_income,0)
		FROM user_profiles WHERE user_id = $1`

	var p models.UserProfile
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&p.UserID, &p.TenantID, &p.Age, &p.Gender,
		&p.HouseholdRegionCode, &p.CurrentResidenceCode,
		&p.IsLocalHukou, &p.EmploymentStatus, &p.UnemploymentRegDate, &p.FlexibleEmploymentRegDate,
		&p.SocialSecurityYears, &p.SkillCertificateLevel, &p.HasChildren,
		&p.ChildAgeRange, &p.HasElderlyDependents,
		&p.DateOfBirth, &p.ContributionMonths, &p.PensionTotalAmount, &p.PensionPersonalAmount,
		&p.MonthlyIncome,
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

	r.db.ExecContext(ctx, `INSERT INTO users (user_id, tenant_id) VALUES ($1, COALESCE($2, 'default')) ON CONFLICT (user_id) DO NOTHING`,
		profile.UserID, profile.TenantID)

	query := `INSERT INTO user_profiles (user_id, tenant_id, age, gender, household_region_code,
		current_residence_code, is_local_hukou, employment_status, unemployment_reg_date,
		flexible_employment_reg_date, social_security_years, skill_certificate_level, has_children,
		child_age_range, has_elderly_dependents, monthly_income,
		date_of_birth, contribution_months, pension_total_amount, pension_personal_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (user_id) DO UPDATE SET
			age = EXCLUDED.age, gender = EXCLUDED.gender,
			household_region_code = EXCLUDED.household_region_code,
			current_residence_code = EXCLUDED.current_residence_code,
			is_local_hukou = EXCLUDED.is_local_hukou,
			employment_status = EXCLUDED.employment_status,
			unemployment_reg_date = EXCLUDED.unemployment_reg_date,
			flexible_employment_reg_date = EXCLUDED.flexible_employment_reg_date,
			social_security_years = EXCLUDED.social_security_years,
			skill_certificate_level = EXCLUDED.skill_certificate_level,
			has_children = EXCLUDED.has_children,
			child_age_range = EXCLUDED.child_age_range,
			has_elderly_dependents = EXCLUDED.has_elderly_dependents,
			monthly_income = EXCLUDED.monthly_income,
			date_of_birth = EXCLUDED.date_of_birth,
			contribution_months = EXCLUDED.contribution_months,
			pension_total_amount = EXCLUDED.pension_total_amount,
			pension_personal_amount = EXCLUDED.pension_personal_amount,
			updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query,
		profile.UserID, profile.TenantID, profile.Age, profile.Gender,
		profile.HouseholdRegionCode, profile.CurrentResidenceCode,
		profile.IsLocalHukou, profile.EmploymentStatus, profile.UnemploymentRegDate, profile.FlexibleEmploymentRegDate,
		profile.SocialSecurityYears, profile.SkillCertificateLevel, profile.HasChildren,
		profile.ChildAgeRange, profile.HasElderlyDependents, profile.MonthlyIncome,
		profile.DateOfBirth, profile.ContributionMonths, profile.PensionTotalAmount, profile.PensionPersonalAmount,
	)
	if err != nil {
		return errors.NewInternal("failed to upsert profile")
	}
	return nil
}
