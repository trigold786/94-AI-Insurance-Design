package server

import (
	"testing"
)

func TestCalculatePlanReturnsSchemes(t *testing.T) {
	resp := CalculatePlan(PlanRequest{
		Age:               30,
		Gender:            "male",
		Employment:        "flexible",
		ContributionYears: 10,
		CurrentBalance:    50000,
		MonthlyBudget:     3000,
		Priority:          "balanced",
		LocalAvgSalary:    10000,
	}, NoopCache{})

	if len(resp.Schemes) == 0 {
		t.Fatal("expected at least 1 scheme")
	}
}

func TestCalculatePlanAllProperties(t *testing.T) {
	resp := CalculatePlan(PlanRequest{
		Age:               35,
		Gender:            "female",
		Employment:        "flexible",
		ContributionYears: 5,
		CurrentBalance:    20000,
		MonthlyBudget:     2000,
		Priority:          "max_pension",
		LocalAvgSalary:    8000,
	}, NoopCache{})

	for _, s := range resp.Schemes {
		if s.Name == "" {
			t.Error("scheme name should not be empty")
		}
		if s.MonthlyCost <= 0 {
			t.Errorf("scheme %s: expected positive monthly cost, got %.2f", s.Name, s.MonthlyCost)
		}
		if s.ProjectedPension < 0 {
			t.Errorf("scheme %s: expected non-negative pension, got %.2f", s.Name, s.ProjectedPension)
		}
	}
}

func TestCalculatePlanCashflowForMale(t *testing.T) {
	resp := CalculatePlan(PlanRequest{
		Age:               40,
		Gender:            "male",
		Employment:        "employed",
		ContributionYears: 15,
		CurrentBalance:    100000,
		MonthlyBudget:     5000,
		Priority:          "balanced",
		LocalAvgSalary:    12000,
	}, NoopCache{})

	for _, s := range resp.Schemes {
		if len(s.Cashflow) == 0 {
			t.Errorf("scheme %s: expected non-empty cashflow", s.Name)
		}
		// Male retirement at 63 (new policy), starting from 40 = 23 years
		if len(s.Cashflow) != 23 {
			t.Errorf("scheme %s: expected 23 cashflow years (40->63), got %d", s.Name, len(s.Cashflow))
		}
	}
}

func TestCalculatePlanCashflowForFemale(t *testing.T) {
	resp := CalculatePlan(PlanRequest{
		Age:               30,
		Gender:            "female",
		Employment:        "flexible",
		ContributionYears: 10,
		CurrentBalance:    50000,
		MonthlyBudget:     3000,
		Priority:          "balanced",
		LocalAvgSalary:    10000,
	}, NoopCache{})

	// Female retirement at 58 (new policy), starting from 30 = 28 years
	for _, s := range resp.Schemes {
		if len(s.Cashflow) != 28 {
			t.Errorf("scheme %s: expected 28 cashflow years, got %d", s.Name, len(s.Cashflow))
		}
	}
}

func TestCalculatePlanBudgetConstraint(t *testing.T) {
	budget := 500.0
	resp := CalculatePlan(PlanRequest{
		Age:               25,
		Gender:            "male",
		Employment:        "unemployed",
		ContributionYears: 0,
		CurrentBalance:    0,
		MonthlyBudget:     budget,
		Priority:          "min_cost",
		LocalAvgSalary:    5000,
	}, NoopCache{})

	for _, s := range resp.Schemes {
		if s.MonthlyCost > budget*1.01 {
			t.Errorf("scheme %s: cost %.2f exceeds budget %.2f", s.Name, s.MonthlyCost, budget)
		}
	}
}

func TestCalculatePlanSchemesOrderedByEfficiency(t *testing.T) {
	resp := CalculatePlan(PlanRequest{
		Age:               30,
		Gender:            "male",
		Employment:        "flexible",
		ContributionYears: 10,
		CurrentBalance:    50000,
		MonthlyBudget:     5000,
		Priority:          "balanced",
		LocalAvgSalary:    10000,
	}, NoopCache{})

	for i := 1; i < len(resp.Schemes); i++ {
		efficiency := func(cost, pension float64) float64 {
			if cost <= 0 {
				return pension
			}
			return pension / cost
		}
		prevEff := efficiency(resp.Schemes[i-1].MonthlyCost, resp.Schemes[i-1].ProjectedPension)
		currEff := efficiency(resp.Schemes[i].MonthlyCost, resp.Schemes[i].ProjectedPension)
		if currEff > prevEff {
			t.Errorf("schemes not ordered by descending efficiency")
		}
	}
}
