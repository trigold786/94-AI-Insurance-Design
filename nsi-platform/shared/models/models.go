package models

import (
	"encoding/json"
	"time"
)

// UserProfile 用户画像
type UserProfile struct {
	UserID                   string   `db:"user_id" json:"user_id"`
	TenantID                 string   `db:"tenant_id" json:"tenant_id"`
	Age                      int      `db:"age" json:"age"`
	Gender                   string   `db:"gender" json:"gender"`
	HouseholdRegionCode      string   `db:"household_region_code" json:"household_region_code"`
	CurrentResidenceCode     string   `db:"current_residence_code" json:"current_residence_code"`
	IsLocalHukou             bool     `db:"is_local_hukou" json:"is_local_hukou"`
	EmploymentStatus         string   `db:"employment_status" json:"employment_status"`
	UnemploymentRegDate      *string  `db:"unemployment_reg_date" json:"unemployment_reg_date,omitempty"`
	FlexibleEmploymentRegDate *string `db:"flexible_employment_reg_date" json:"flexible_employment_reg_date,omitempty"`
	SocialSecurityYears      int      `db:"social_security_years" json:"social_security_years"`
	SkillCertificateLevel    *string  `db:"skill_certificate_level" json:"skill_certificate_level,omitempty"`
	HasChildren              bool     `db:"has_children" json:"has_children"`
	ChildAgeRange            string   `db:"child_age_range" json:"child_age_range"`
	HasElderlyDependents     bool     `db:"has_elderly_dependents" json:"has_elderly_dependents"`
	MonthlyIncome            float64  `db:"monthly_income" json:"monthly_income"`
	// 新增字段
	DateOfBirth           string  `db:"date_of_birth" json:"date_of_birth"`
	ContributionMonths    int     `db:"contribution_months" json:"contribution_months"`
	PensionTotalAmount    float64 `db:"pension_total_amount" json:"pension_total_amount"`
	PensionPersonalAmount float64 `db:"pension_personal_amount" json:"pension_personal_amount"`
}

// PolicyClaim 结构化政策原子
type PolicyClaim struct {
	ClaimID           string           `db:"claim_id" json:"claim_id"`
	PolicyID          string           `db:"policy_id" json:"policy_id"`
	RegionCode        string           `db:"region_code" json:"region_code"`
	PolicyType        string           `db:"policy_type" json:"policy_type"`
	TargetGroupTags   []string         `db:"target_group_tags" json:"target_group_tags"`
	SubsidyCalcMethod string           `db:"subsidy_calc_method" json:"subsidy_calc_method"`
	SubsidyAmountMin  *float64         `db:"subsidy_amount_min" json:"subsidy_amount_min,omitempty"`
	SubsidyAmountMax  *float64         `db:"subsidy_amount_max" json:"subsidy_amount_max,omitempty"`
	SubsidyDuration   *int             `db:"subsidy_duration" json:"subsidy_duration,omitempty"`
	EffectiveDate     string           `db:"effective_date" json:"effective_date"`
	ExpireDate        *string          `db:"expire_date" json:"expire_date,omitempty"`
	PublishDate       string           `db:"publish_date" json:"publish_date,omitempty"`
	ConfidenceScore   float64          `db:"confidence_score" json:"confidence_score"`
	Status            string           `db:"status" json:"status"`
	VersionNumber     int              `db:"version_number" json:"version_number"`
	Conditions        json.RawMessage  `db:"conditions" json:"conditions,omitempty"`
	RequiredDocuments json.RawMessage  `db:"required_documents" json:"required_documents,omitempty"`
	SourceID          string           `db:"source_id" json:"source_id,omitempty"`
	SourceName        string           `db:"source_name" json:"source_name,omitempty"`
	SourceURL         string           `db:"source_url" json:"source_url,omitempty"`
	PolicyURL         string           `db:"policy_url" json:"policy_url,omitempty"`
	PolicyTitle       string           `db:"policy_title" json:"policy_title,omitempty"`
	IssuingAuthority  string           `db:"issuing_authority" json:"issuing_authority,omitempty"`
	DocumentNumber    string           `db:"document_number" json:"document_number,omitempty"`
	ApplicationProcess json.RawMessage `db:"application_process" json:"application_process,omitempty"`
	ContactInfo       string           `db:"contact_info" json:"contact_info,omitempty"`
	SourceType        string           `db:"source_type" json:"source_type,omitempty"`
	ExtractionMethod  string           `db:"extraction_method" json:"extraction_method,omitempty"`
	RawTextLength     int              `db:"raw_text_length" json:"raw_text_length,omitempty"`
	SplitCount        int              `db:"split_count" json:"split_count,omitempty"`
	SourceLevel       string           `db:"source_level" json:"source_level,omitempty"`
	FetchedAt         string           `db:"fetched_at" json:"fetched_at,omitempty"`
	VerifiedBy        string           `db:"verified_by" json:"verified_by,omitempty"`
	MatchRate         float64          `db:"match_rate" json:"match_rate,omitempty"`
	ConflictScore     float64          `db:"conflict_score" json:"conflict_score,omitempty"`
	Embedding         []float64        `db:"embedding" json:"embedding,omitempty"`
}

// ComplianceCondition 认定条件
type ComplianceCondition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	TagMatch    string `json:"tag_match,omitempty"`
}

// RequiredDocument 必需材料
type RequiredDocument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"` // user, gov, employer
	Optional    bool   `json:"optional"`
}

// ComplianceChecklist 合规检查清单
type ComplianceChecklist struct {
	UserID           string               `json:"user_id"`
	CityCode         string               `json:"city_code"`
	CityName         string               `json:"city_name"`
	MatchedPolicies  []PolicyCompliance   `json:"matched_policies"`
	RequiredDocs     []RequiredDocument   `json:"required_docs"`
	EligibleTags     []string             `json:"eligible_tags"`
}

// ProcessingStep 办理流程步骤
type ProcessingStep struct {
	Order       int    `json:"order"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
}

// PolicyCompliance 单个政策的合规信息
type PolicyCompliance struct {
	PolicyID         string                `json:"policy_id"`
	PolicyType       string                `json:"policy_type"`
	ClaimID          string                `json:"claim_id"`
	SubsidyCalcMethod string               `json:"subsidy_calc_method"`
	Conditions       []ComplianceCondition `json:"conditions"`
	ConditionStatuses map[string]bool      `json:"condition_statuses,omitempty"`
	RequiredDocs     []RequiredDocument    `json:"required_docs"`
	IsEligible       bool                  `json:"is_eligible"`
	UnmetConditions  []string              `json:"unmet_conditions,omitempty"`
	ProcessingSteps  []ProcessingStep      `json:"processing_steps,omitempty"`
}

// PlanSnapshot 方案快照
type PlanSnapshot struct {
	PlanID                   string              `db:"plan_id" json:"plan_id"`
	UserID                   string              `db:"user_id" json:"user_id"`
	PolicyVersionSnapshotID  string              `db:"policy_version_snapshot_id" json:"policy_version_snapshot_id"`
	RecommendedSchemes       []Scheme            `db:"recommended_schemes" json:"recommended_schemes"`
	FreeFormText             string              `db:"free_form_text" json:"free_form_text"`
	StructuredSchemes        []LLMScheme         `db:"structured_schemes" json:"structured_schemes"`
	PolicyReferences         []PolicyReference   `db:"policy_references" json:"policy_references"`
	Recommendation           string              `db:"recommendation" json:"recommendation"`
	RecommendationReason     string              `db:"recommendation_reason" json:"recommendation_reason"`
	VerificationResult       *VerificationResult `db:"verification_result" json:"verification_result,omitempty"`
	TotalCost                float64             `db:"total_cost" json:"total_cost"`
	TotalSubsidy             float64             `db:"total_subsidy" json:"total_subsidy"`
	GeneratedAt              time.Time           `db:"generated_at" json:"generated_at"`
}

// Scheme 推荐方案
type Scheme struct {
	Name                  string         `json:"name"`
	BaseSalary            int            `json:"base_salary"`
	MonthlyCost           float64        `json:"monthly_cost"`
	AnnualSubsidy         float64        `json:"annual_subsidy"`
	SubsidyPolicy         string         `json:"subsidy_policy"`
	SubsidyCondition      string         `json:"subsidy_condition"`
	PaidMonths            int            `json:"paid_months"`
	TargetMonths          int            `json:"target_months"`
	RemainingMonths       int            `json:"remaining_months"`
	TotalPersonalCost     float64        `json:"total_personal_cost"`
	RemainingPersonalCost float64        `json:"remaining_personal_cost"`
	ProjectedPension      float64        `json:"projected_pension"`
	AfterTaxPension       float64        `json:"after_tax_pension"`
	Cashflow              []CashFlowItem `json:"cashflow,omitempty"`
}

// CashFlowItem 现金流
type CashFlowItem struct {
	Year    int     `json:"year"`
	Payment float64 `json:"payment"`
	Subsidy float64 `json:"subsidy"`
	Balance float64 `json:"balance"`
}

// PaymentRecord 缴费记录
type PaymentRecord struct {
	RecordID   string  `db:"record_id" json:"record_id"`
	UserID     string  `db:"user_id" json:"user_id"`
	PolicyType string  `db:"policy_type" json:"policy_type"`
	Month      string  `db:"month" json:"month"`       // YYYY-MM
	Amount     float64 `db:"amount" json:"amount"`
	Status     string  `db:"status" json:"status"`     // paid, pending, missed
	DueDate    string  `db:"due_date" json:"due_date"` // YYYY-MM-DD
	PaidDate   *string `db:"paid_date" json:"paid_date,omitempty"`
}

// Alert 权益预警
type Alert struct {
	AlertID    string `db:"alert_id" json:"alert_id"`
	UserID     string `db:"user_id" json:"user_id"`
	AlertType  string `db:"alert_type" json:"alert_type"`   // disconnection_risk, policy_change
	Severity   string `db:"severity" json:"severity"`       // high, medium, low
	Title      string `db:"title" json:"title"`
	Message    string `db:"message" json:"message"`
	IsRead     bool   `db:"is_read" json:"is_read"`
	CreatedAt  string `db:"created_at" json:"created_at"`
	PolicyID   *string `db:"policy_id" json:"policy_id,omitempty"`
}

// PolicySource 政策数据源
type PolicySource struct {
	SourceID    string  `db:"source_id" json:"source_id"`
	SourceName  string  `db:"source_name" json:"source_name"`
	SourceURL   string  `db:"source_url" json:"source_url"`
	SourceLevel string  `db:"source_level" json:"source_level"` // HIGH, MEDIUM, LOW
	Weight      float64 `db:"weight" json:"weight"`
}

// VersionSnapshot 政策版本快照
type VersionSnapshot struct {
	ID            int              `db:"id" json:"id"`
	ClaimID       string           `db:"claim_id" json:"claim_id"`
	PolicyID      string           `db:"policy_id" json:"policy_id"`
	VersionNumber int              `db:"version_number" json:"version_number"`
	SnapshotData  *json.RawMessage  `db:"snapshot_data" json:"snapshot_data"`
	SupersededBy  string           `db:"superseded_by" json:"superseded_by,omitempty"`
	CreatedAt     string           `db:"created_at" json:"created_at"`
}

type LLMScheme struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	MonthlyCost         float64  `json:"monthly_cost"`
	AnnualSubsidy       float64  `json:"annual_subsidy"`
	ProjectedPension    float64  `json:"projected_pension"`
	TotalCost           float64  `json:"total_cost"`
	ContributionBase    float64  `json:"contribution_base"`
	PensionEmployeeRate float64  `json:"pension_employee_rate"`
	PensionEmployerRate float64  `json:"pension_employer_rate"`
	MedicalEmployeeRate float64  `json:"medical_employee_rate"`
	Analysis            string   `json:"analysis"`
	ApplicablePolicies  []string `json:"applicable_policies"`
}

type PolicyReference struct {
	ClaimID         string `json:"claim_id"`
	PolicyTitle     string `json:"policy_title"`
	DocumentNumber  string `json:"document_number"`
	PolicyURL       string `json:"policy_url"`
	RelevantExcerpt string `json:"relevant_excerpt"`
	HowApplied      string `json:"how_applied"`
}

type DeviationDetail struct {
	Metric       string  `json:"metric"`
	LLMValue     float64 `json:"llm_value"`
	ActuaryValue float64 `json:"actuary_value"`
	DeviationPct float64 `json:"deviation_pct"`
}

type VerificationResult struct {
	Status       string            `json:"status"`
	MaxDeviation float64           `json:"max_deviation_pct"`
	Details      []DeviationDetail `json:"details"`
}

type LLMSchemeResponse struct {
	Summary          string            `json:"summary"`
	Schemes          []LLMScheme       `json:"schemes"`
	PolicyReferences []PolicyReference `json:"policy_references"`
	Recommendation   struct {
		RecommendedScheme string `json:"recommended_scheme"`
		Reasoning         string `json:"reasoning"`
	} `json:"recommendation"`
}

type OrderData struct {
	OrderID       string  `db:"order_id" json:"order_id"`
	UserID        string  `db:"user_id" json:"user_id"`
	PlanID        string  `db:"plan_id" json:"plan_id"`
	Amount        float64 `db:"amount" json:"amount"`
	Status        string  `db:"status" json:"status"`
	PaymentMethod string  `db:"payment_method" json:"payment_method"`
	PaidAt        *string `db:"paid_at" json:"paid_at,omitempty"`
	CreatedAt     string  `db:"created_at" json:"created_at"`
}

type SimScenario struct {
	ID        int             `db:"id" json:"id"`
	UserID    string          `db:"user_id" json:"user_id"`
	Name      string          `db:"name" json:"name"`
	Params    json.RawMessage `db:"params" json:"params"`
	Result    json.RawMessage `db:"result" json:"result"`
	CreatedAt string          `db:"created_at" json:"created_at"`
}

type ThresholdData struct {
	ClaimID    string          `db:"claim_id" json:"claim_id"`
	Conditions json.RawMessage `db:"conditions" json:"conditions"`
}

type SettingsData struct {
	FontScale      string `json:"font_scale"`
	DefaultTab     string `json:"default_tab"`
	NotificationsOn bool   `json:"notifications_on"`
}
