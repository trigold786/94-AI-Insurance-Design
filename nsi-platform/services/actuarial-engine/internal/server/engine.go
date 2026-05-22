package server

import (
	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/calculator"
	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/cashflow"
	"github.com/trigold786/94-AI-Insurance-Design/actuarial-engine/internal/optimizer"
)

type PlanRequest struct {
	RegionCode       string
	Age              int
	Gender           string
	Employment       string
	ContributionYears int
	CurrentBalance   float64
	MonthlyBudget    float64
	Priority         string
	LocalAvgSalary   float64
}

type PlanResponse struct {
	Schemes           []PlanScheme
	CalculationTimeMs float64
}

type PlanScheme struct {
	Name             string
	BaseSalary       int
	MonthlyCost      float64
	AnnualSubsidy    float64
	ProjectedPension float64
	Cashflow         []cashflow.CashFlowItem
}

func CalculatePlan(req PlanRequest) PlanResponse {
	schemes := calculator.GenerateSchemes(calculator.GenerateInput{
		Age:               req.Age,
		Gender:            req.Gender,
		ContributionYears: req.ContributionYears,
		LocalAvgSalary:    req.LocalAvgSalary,
		CurrentBalance:    req.CurrentBalance,
		MonthlyBudget:     req.MonthlyBudget,
		PensionAge:        getPensionAge(req.Gender),
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
	frontier := optimizer.FindParetoFrontier(optSchemes)

	frontierMap := make(map[string]bool)
	for _, f := range frontier {
		frontierMap[f.Name] = true
	}

	yearsToRetirement := getPensionAge(req.Gender) - req.Age
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
		if baseScheme != nil {
			baseSalary = baseScheme.BaseSalary
		}

		cf := cashflow.Project(cashflow.CashflowInput{
			Years:            yearsToRetirement,
			AnnualPayment:    s.MonthlyCost * 12,
			AnnualSubsidy:    0,
			InitialBalance:   req.CurrentBalance,
			SalaryGrowthRate: 0.05,
			FundReturnRate:   0.03,
		})

		planSchemes = append(planSchemes, PlanScheme{
			Name:             s.Name,
			BaseSalary:       baseSalary,
			MonthlyCost:      s.MonthlyCost,
			AnnualSubsidy:    0,
			ProjectedPension: s.ProjectedPension,
			Cashflow:         cf,
		})
	}

	return PlanResponse{
		Schemes: planSchemes,
	}
}

func getPensionAge(gender string) int {
	if gender == "female" {
		return 55
	}
	return 60
}
