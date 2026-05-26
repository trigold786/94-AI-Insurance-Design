package server

import (
	"encoding/json"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/calculator"
	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/cashflow"
	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/optimizer"
	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/tax"
)

type PlanRequest struct {
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
	SubsidyCalcMethod    string  `json:"subsidy_calc_method"`
}

type PlanResponse struct {
	Schemes           []PlanScheme `json:"schemes"`
	CalculationTimeMs float64      `json:"calculation_time_ms"`
}

type PlanScheme struct {
	Name                  string                 `json:"name"`
	BaseSalary            int                    `json:"base_salary"`
	MonthlyCost           float64                `json:"monthly_cost"`
	AnnualSubsidy         float64                `json:"annual_subsidy"`
	SubsidyPolicy         string                 `json:"subsidy_policy"`
	SubsidyCondition      string                 `json:"subsidy_condition"`
	PaidMonths            int                    `json:"paid_months"`
	TargetMonths          int                    `json:"target_months"`
	RemainingMonths       int                    `json:"remaining_months"`
	TotalPersonalCost     float64                `json:"total_personal_cost"`
	RemainingPersonalCost float64                `json:"remaining_personal_cost"`
	ProjectedPension      float64                `json:"projected_pension"`
	AfterTaxPension       float64                `json:"after_tax_pension"`
	Cashflow              []cashflow.CashFlowItem `json:"cashflow,omitempty"`
}

func CalculatePlan(req PlanRequest, cache Cache) PlanResponse {
	key := cacheKey(req)
	if cached, ok := cache.Get(key); ok {
		var resp PlanResponse
		if err := json.Unmarshal(cached, &resp); err == nil {
			return resp
		}
	}

	start := time.Now()

	pensionAge := req.GenderPensionAge
	if pensionAge <= 0 {
		pensionAge = 63
		if req.Gender == "female" {
			pensionAge = 58
		}
		if req.OriginalPensionAge == 50 {
			pensionAge = 55
		}
	}

	schemes := calculator.GenerateSchemes(calculator.GenerateInput{
		Age:                  req.Age,
		Gender:               req.Gender,
		Employment:           req.Employment,
		ContributionYears:    req.ContributionYears,
		ContributionMonths:   req.ContributionMonths,
		LocalAvgSalary:       req.LocalAvgSalary,
		CurrentBalance:       req.CurrentBalance,
		MonthlyBudget:        req.MonthlyBudget,
		PensionAge:           pensionAge,
		PensionEmployeeRate:  req.PensionEmployeeRate,
		PensionEmployerRate:  req.PensionEmployerRate,
		MedicalEmployeeRate:  req.MedicalEmployeeRate,
		MedicalEmployerRate:  req.MedicalEmployerRate,
		ContributionBaseMin:  req.ContributionBaseMin,
		ContributionBaseMax:  req.ContributionBaseMax,
		SubsidyCalcMethod:    req.SubsidyCalcMethod,
	})

	var optSchemes []optimizer.Scheme
	for _, s := range schemes {
		optSchemes = append(optSchemes, optimizer.Scheme{
			Name:             s.Name,
			MonthlyCost:      s.MonthlyCost,
			ProjectedPension: s.ProjectedPension,
		})
	}

	optSchemes = optimizer.RankByEfficiency(optSchemes)
	optSchemes = optimizer.FilterParetoOptimal(optSchemes, 0)

	yearsToRetirement := pensionAge - req.Age
	if yearsToRetirement < 0 {
		yearsToRetirement = 0
	}

	var planSchemes []PlanScheme
	for _, s := range optSchemes {
		var baseScheme *calculator.Scheme
		for i := range schemes {
			if schemes[i].Name == s.Name {
				baseScheme = &schemes[i]
				break
			}
		}

		baseSalary := 0
		annualSubsidy := 0.0
		subsidyPolicy := ""
		subsidyCondition := ""
		paidMonths := 0
		targetMonths := 240
		remainingMonths := 0
		totalPersonalCost := 0.0
		remainingPersonalCost := 0.0
		if baseScheme != nil {
			baseSalary = baseScheme.BaseSalary
			annualSubsidy = baseScheme.AnnualSubsidy
			subsidyPolicy = baseScheme.SubsidyPolicy
			subsidyCondition = baseScheme.SubsidyCondition
			paidMonths = baseScheme.PaidMonths
			targetMonths = baseScheme.TargetMonths
			remainingMonths = baseScheme.RemainingMonths
			totalPersonalCost = baseScheme.TotalPersonalCost
			remainingPersonalCost = baseScheme.RemainingPersonalCost
		}

		cf := cashflow.Project(cashflow.CashflowInput{
			Years:            yearsToRetirement,
			AnnualPayment:    s.MonthlyCost * 12,
			AnnualSubsidy:    annualSubsidy,
			InitialBalance:   req.CurrentBalance,
			SalaryGrowthRate: 0.05,
			FundReturnRate:   0.03,
		})

		annualPension := s.ProjectedPension * 12
		annualTax, _ := tax.CalculateTax(annualPension, float64(baseSalary))
		afterTaxPension := s.ProjectedPension - annualTax/12
		if afterTaxPension < 0 {
			afterTaxPension = 0
		}

		planSchemes = append(planSchemes, PlanScheme{
			Name:                  s.Name,
			BaseSalary:            baseSalary,
			MonthlyCost:           s.MonthlyCost,
			AnnualSubsidy:         annualSubsidy,
			SubsidyPolicy:         subsidyPolicy,
			SubsidyCondition:      subsidyCondition,
			PaidMonths:            paidMonths,
			TargetMonths:          targetMonths,
			RemainingMonths:       remainingMonths,
			TotalPersonalCost:     totalPersonalCost,
			RemainingPersonalCost: remainingPersonalCost,
			ProjectedPension:      s.ProjectedPension,
			AfterTaxPension:       afterTaxPension,
			Cashflow:              cf,
		})
	}

	resp := PlanResponse{
		Schemes:           planSchemes,
		CalculationTimeMs: float64(time.Since(start).Microseconds()) / 1000,
	}

	data, err := json.Marshal(resp)
	if err == nil {
		cache.Set(key, data, 24*time.Hour)
	}

	return resp
}


