package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type CalculateRequest struct {
	Age                  int     `json:"age"`
	Gender               string  `json:"gender"`
	GenderPensionAge     int     `json:"gender_pension_age"`
	OriginalPensionAge   int     `json:"original_pension_age"`
	Employment           string  `json:"employment"`
	ContributionYears    int     `json:"contribution_years"`
	ContributionMonths   int     `json:"contribution_months"`
	CurrentBalance       float64 `json:"current_balance"`
	MonthlyBudget        float64 `json:"monthly_budget"`
	Priority             string  `json:"priority"`
	LocalAvgSalary       float64 `json:"local_avg_salary"`
	PensionEmployeeRate  float64 `json:"pension_employee_rate"`
	PensionEmployerRate  float64 `json:"pension_employer_rate"`
	MedicalEmployeeRate  float64 `json:"medical_employee_rate"`
	MedicalEmployerRate  float64 `json:"medical_employer_rate"`
	ContributionBaseMin  float64 `json:"contribution_base_min"`
	ContributionBaseMax  float64 `json:"contribution_base_max"`
}

type SchemeResult struct {
	Name                  string                `json:"name"`
	BaseSalary            int                   `json:"base_salary"`
	MonthlyCost           float64               `json:"monthly_cost"`
	AnnualSubsidy         float64               `json:"annual_subsidy"`
	SubsidyPolicy         string                `json:"subsidy_policy"`
	SubsidyCondition      string                `json:"subsidy_condition"`
	PaidMonths            int                   `json:"paid_months"`
	TargetMonths          int                   `json:"target_months"`
	RemainingMonths       int                   `json:"remaining_months"`
	TotalPersonalCost     float64               `json:"total_personal_cost"`
	RemainingPersonalCost float64               `json:"remaining_personal_cost"`
	ProjectedPension      float64               `json:"projected_pension"`
	AfterTaxPension       float64               `json:"after_tax_pension"`
	Cashflow              []models.CashFlowItem `json:"cashflow,omitempty"`
}

type CalculateResponse struct {
	Schemes           []SchemeResult `json:"schemes"`
	CalculationTimeMs float64        `json:"calculation_time_ms"`
}

type Calculator interface {
	Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error)
}

type PlanRepository interface {
	Save(ctx context.Context, plan *models.PlanSnapshot) error
	GetByID(ctx context.Context, planID string) (*models.PlanSnapshot, error)
}

type ProfileLookuper interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
}

type HTTPCalculator struct {
	Endpoint string
	Client   *http.Client
}

func (c *HTTPCalculator) getClient() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func (c *HTTPCalculator) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint+"/v1/calculate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if secret := os.Getenv("ACTUARY_SECRET"); secret != "" {
		httpReq.Header.Set("X-Actuary-Secret", secret)
	}

	resp, err := c.getClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call calculator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("calculator returned status %d", resp.StatusCode)
	}

	var calcResp CalculateResponse
	if err := json.NewDecoder(resp.Body).Decode(&calcResp); err != nil {
		return nil, fmt.Errorf("failed to decode calculator response: %w", err)
	}

	return &calcResp, nil
}

func GeneratePlanHandler(calc Calculator, repo PlanRepository, profileRepo ProfileLookuper, policyRepo PolicyQuerier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var calcReq CalculateRequest
		if err := json.NewDecoder(r.Body).Decode(&calcReq); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}

		if calcReq.Age < 16 || calcReq.Age > 70 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "age must be between 16 and 70"})
			return
		}
		if calcReq.Gender != "male" && calcReq.Gender != "female" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "gender must be male or female"})
			return
		}
		if calcReq.MonthlyBudget <= 0 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "monthly_budget must be positive"})
			return
		}
		if calcReq.ContributionYears < 0 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "contribution_years must be non-negative"})
			return
		}
		if profileRepo != nil {
			if profile, err := profileRepo.GetByUserID(r.Context(), userID); err == nil {
				code := profile.CurrentResidenceCode
				if code == "" {
					code = profile.HouseholdRegionCode
				}
				// 从实际政策数据中提取费率参数
				if policyRepo != nil && code != "" {
					policies, err := policyRepo.QueryByRegionAndStatus(r.Context(), code, "verified")
					if err == nil {
						for _, p := range policies {
							switch p.PolicyType {
							case "pension":
								calcReq.PensionEmployeeRate = parseRate(p.SubsidyCalcMethod, "个人")
								calcReq.PensionEmployerRate = parseRate(p.SubsidyCalcMethod, "单位")
								calcReq.GenderPensionAge = calcPensionAge(profile.DateOfBirth, profile.Gender, calcReq.OriginalPensionAge)
							case "medical":
								calcReq.MedicalEmployeeRate = parseRate(p.SubsidyCalcMethod, "个人")
								calcReq.MedicalEmployerRate = parseRate(p.SubsidyCalcMethod, "单位")
							case "subsidy":
								if p.SubsidyAmountMin != nil {
									calcReq.ContributionBaseMin = *p.SubsidyAmountMin
								}
								if p.SubsidyAmountMax != nil {
									calcReq.ContributionBaseMax = *p.SubsidyAmountMax
									if calcReq.LocalAvgSalary <= 0 {
										calcReq.LocalAvgSalary = *p.SubsidyAmountMax
									}
								}
							}
						}
					}
				}
				if calcReq.LocalAvgSalary <= 0 {
					if ci := GetCityInfo(code); ci != nil {
						calcReq.LocalAvgSalary = ci.AvgSalary
					}
				}
			}
		}
		if calcReq.LocalAvgSalary <= 0 {
			// fallback: 使用上海平均工资
			if ci := GetCityInfo("310000"); ci != nil {
				calcReq.LocalAvgSalary = ci.AvgSalary
			}
		}
		if calcReq.LocalAvgSalary <= 0 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "无法确定当地平均工资，请完善用户信息"})
			return
		}

		if calc == nil {
			respondError(w, fmt.Errorf("calculator not configured"))
			return
		}

		calcResp, err := calc.Calculate(r.Context(), &calcReq)
		if err != nil {
			respondError(w, err)
			return
		}

		var totalCost, totalSubsidy float64
		var schemes []models.Scheme
		for _, s := range calcResp.Schemes {
			totalCost += s.MonthlyCost * 12
			totalSubsidy += s.AnnualSubsidy
			schemes = append(schemes, models.Scheme{
				Name:                  s.Name,
				BaseSalary:            s.BaseSalary,
				MonthlyCost:           s.MonthlyCost,
				AnnualSubsidy:         s.AnnualSubsidy,
				SubsidyPolicy:         s.SubsidyPolicy,
				SubsidyCondition:      s.SubsidyCondition,
				PaidMonths:            s.PaidMonths,
				TargetMonths:          s.TargetMonths,
				RemainingMonths:       s.RemainingMonths,
				TotalPersonalCost:     s.TotalPersonalCost,
				RemainingPersonalCost: s.RemainingPersonalCost,
				ProjectedPension:      s.ProjectedPension,
				AfterTaxPension:       s.AfterTaxPension,
				Cashflow:              s.Cashflow,
			})
		}

		snapshot := &models.PlanSnapshot{
			PlanID:             fmt.Sprintf("plan-%d", time.Now().UnixNano()),
			UserID:             userID,
			RecommendedSchemes: schemes,
			TotalCost:          totalCost,
			TotalSubsidy:       totalSubsidy,
			GeneratedAt:        time.Now(),
		}

		if repo != nil {
			if err := repo.Save(r.Context(), snapshot); err != nil {
				respondError(w, err)
				return
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": snapshot})
	})
}

func PlanDetailHandler(repo PlanRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		planID := strings.TrimPrefix(r.URL.Path, "/v1/plans/")
		if planID == "" || strings.Contains(planID, "/") {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid plan_id"})
			return
		}

		plan, err := repo.GetByID(r.Context(), planID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				respondJSON(w, http.StatusNotFound, map[string]interface{}{"code": "NOT_FOUND", "message": err.Error()})
				return
			}
			respondError(w, err)
			return
		}

		if plan.UserID != userID {
			respondJSON(w, http.StatusNotFound, map[string]interface{}{"code": "NOT_FOUND", "message": "plan not found"})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": plan})
	})
}

func calcPensionAge(dob, gender string, originalPensionAge int) int {
	baseAge := 60
	if gender == "female" {
		baseAge = 55
		if originalPensionAge >= 50 && originalPensionAge <= 60 {
			baseAge = originalPensionAge
		}
	}
	var pace, maxDelay int
	switch baseAge {
	case 50:
		pace = 2
		maxDelay = 60
	case 55:
		pace = 4
		maxDelay = 36
	default:
		pace = 4
		maxDelay = 36
	}
	if len(dob) < 7 {
		return baseAge + maxDelay/12
	}
	parts := strings.Split(dob, "-")
	if len(parts) < 2 {
		return baseAge + maxDelay/12
	}
	by, _ := strconv.Atoi(parts[0])
	bm, _ := strconv.Atoi(parts[1])
	baseYear := 2025 - baseAge
	monthsSincePolicy := (by-baseYear)*12 + bm - 1
	if monthsSincePolicy < 0 {
		return baseAge
	}
	delay := monthsSincePolicy/pace + 1
	if delay > maxDelay {
		delay = maxDelay
	}
	return baseAge + delay/12
}

// parseRate 从 subsidy_calc_method 中解析费率
// 格式如 "基数*8%+基数*16%" 中提取个人部分(第一个)和单位部分(第二个)
// label: "个人" 返回第一个费率, "单位" 返回第二个费率
func parseRate(method, label string) float64 {
	if method == "" {
		return 0
	}
	var rates []float64
	parts := strings.Split(method, "+")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "%"); idx >= 0 {
			start := strings.LastIndex(p[:idx], "*")
			if start >= 0 {
				s := p[start+1 : idx]
				var r float64
				if _, err := fmt.Sscanf(s, "%f", &r); err == nil {
					rates = append(rates, r/100)
				}
			}
		}
	}
	if label == "单位" && len(rates) >= 2 {
		return rates[1]
	}
	if len(rates) >= 1 {
		return rates[0]
	}
	return 0
}
