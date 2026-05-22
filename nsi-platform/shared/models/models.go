package models

import "time"

// UserProfile 用户画像
type UserProfile struct {
	UserID                   string   `db:"user_id" json:"user_id"`
	TenantID                 string   `db:"tenant_id" json:"tenant_id"`
	Age                      int      `db:"age" json:"age"`
	Gender                   string   `db:"gender" json:"gender"`
	HouseholdRegionCode      string   `db:"household_region_code" json:"household_region_code"`
	CurrentResidenceCode     string   `db:"current_residence_code" json:"current_residence_code"`
	EmploymentStatus         string   `db:"employment_status" json:"employment_status"`
	UnemploymentRegDate      *string  `db:"unemployment_reg_date" json:"unemployment_reg_date,omitempty"`
	FlexibleEmploymentRegDate *string `db:"flexible_employment_reg_date" json:"flexible_employment_reg_date,omitempty"`
	SocialSecurityYears      int      `db:"social_security_years" json:"social_security_years"`
	SkillCertificateLevel    *string  `db:"skill_certificate_level" json:"skill_certificate_level,omitempty"`
	HasChildren              bool     `db:"has_children" json:"has_children"`
}

// PolicyClaim 结构化政策原子
type PolicyClaim struct {
	ClaimID          string   `db:"claim_id" json:"claim_id"`
	PolicyID         string   `db:"policy_id" json:"policy_id"`
	RegionCode       string   `db:"region_code" json:"region_code"`
	PolicyType       string   `db:"policy_type" json:"policy_type"`
	TargetGroupTags  []string `db:"target_group_tags" json:"target_group_tags"`
	SubsidyCalcMethod string  `db:"subsidy_calc_method" json:"subsidy_calc_method"`
	SubsidyAmountMin *float64 `db:"subsidy_amount_min" json:"subsidy_amount_min,omitempty"`
	SubsidyAmountMax *float64 `db:"subsidy_amount_max" json:"subsidy_amount_max,omitempty"`
	SubsidyDuration  *int     `db:"subsidy_duration" json:"subsidy_duration,omitempty"`
	EffectiveDate    string   `db:"effective_date" json:"effective_date"`
	ExpireDate       *string  `db:"expire_date" json:"expire_date,omitempty"`
	ConfidenceScore  float64  `db:"confidence_score" json:"confidence_score"`
	Status           string   `db:"status" json:"status"` // verified, pending_review, unverified
	VersionNumber    int      `db:"version_number" json:"version_number"`
}

// PlanSnapshot 方案快照
type PlanSnapshot struct {
	PlanID              string    `db:"plan_id" json:"plan_id"`
	UserID              string    `db:"user_id" json:"user_id"`
	PolicyVersionSnapshotID string `db:"policy_version_snapshot_id" json:"policy_version_snapshot_id"`
	RecommendedSchemes  []Scheme  `db:"recommended_schemes" json:"recommended_schemes"`
	TotalCost           float64   `db:"total_cost" json:"total_cost"`
	TotalSubsidy        float64   `db:"total_subsidy" json:"total_subsidy"`
	GeneratedAt         time.Time `db:"generated_at" json:"generated_at"`
}

// Scheme 推荐方案
type Scheme struct {
	Name            string           `json:"name"`
	BaseSalary      int              `json:"base_salary"`
	MonthlyCost     float64          `json:"monthly_cost"`
	AnnualSubsidy   float64          `json:"annual_subsidy"`
	ProjectedPension float64          `json:"projected_pension"`
	Cashflow        []CashFlowItem   `json:"cashflow,omitempty"`
}

// CashFlowItem 现金流
type CashFlowItem struct {
	Year    int     `json:"year"`
	Payment float64 `json:"payment"`
	Subsidy float64 `json:"subsidy"`
	Balance float64 `json:"balance"`
}

// PolicySource 政策数据源
type PolicySource struct {
	SourceID    string  `db:"source_id" json:"source_id"`
	SourceName  string  `db:"source_name" json:"source_name"`
	SourceURL   string  `db:"source_url" json:"source_url"`
	SourceLevel string  `db:"source_level" json:"source_level"` // HIGH, MEDIUM, LOW
	Weight      float64 `db:"weight" json:"weight"`
}
