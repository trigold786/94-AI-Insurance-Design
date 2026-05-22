package calculator

import (
	"math"
	"testing"
)

func TestCalculateBasicPension(t *testing.T) {
	result := CalculateBasicPension(10000, 12000, 15)
	expected := (10000+12000)/2*15*0.01 // 1650
	if !almostEqual(result, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestCalculateBasicPensionZeroYears(t *testing.T) {
	result := CalculateBasicPension(10000, 12000, 0)
	if result != 0 {
		t.Errorf("expected 0 for 0 years, got %.2f", result)
	}
}

func TestCalculateBasicPensionZeroSalary(t *testing.T) {
	result := CalculateBasicPension(0, 0, 15)
	if result != 0 {
		t.Errorf("expected 0 for zero salaries, got %.2f", result)
	}
}

func TestCalculatePersonalAccountPension(t *testing.T) {
	result := CalculatePersonalAccountPension(100000, 60)
	expected := 100000.0 / 139.0 // 719.42
	if !almostEqual(result, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestCalculatePersonalAccountPensionZeroBalance(t *testing.T) {
	result := CalculatePersonalAccountPension(0, 60)
	if result != 0 {
		t.Errorf("expected 0 for zero balance, got %.2f", result)
	}
}

func TestCalculatePersonalAccountPensionDifferentAges(t *testing.T) {
	ages := map[int]float64{60: 139, 55: 170, 50: 195}
	for age, divisor := range ages {
		result := CalculatePersonalAccountPension(100000, age)
		expected := 100000.0 / divisor
		if !almostEqual(result, expected, 0.01) {
			t.Errorf("age %d: expected %.2f, got %.2f", age, expected, result)
		}
	}
}

func TestCalculateMonthlyContribution(t *testing.T) {
	result := CalculateMonthlyContribution(10000, 0.08, 0.02)
	expected := 10000 * (0.08 + 0.02) // 1000
	if !almostEqual(result, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestCalculateMonthlyContributionDifferentRates(t *testing.T) {
	result := CalculateMonthlyContribution(20000, 0.08, 0.12)
	expected := 20000 * 0.20 // 4000
	if !almostEqual(result, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestCalculateMonthlyContributionZeroSalary(t *testing.T) {
	result := CalculateMonthlyContribution(0, 0.08, 0.02)
	if result != 0 {
		t.Errorf("expected 0 for zero salary, got %.2f", result)
	}
}

func TestGenerateSchemes(t *testing.T) {
	schemes := GenerateSchemes(GenerateInput{
		Age:              30,
		Gender:           "male",
		ContributionYears: 15,
		LocalAvgSalary:   10000,
		CurrentBalance:    50000,
		MonthlyBudget:    3000,
		PensionAge:       60,
	})

	if len(schemes) == 0 {
		t.Fatal("expected at least 1 scheme")
	}

	for _, s := range schemes {
		if s.BaseSalary <= 0 {
			t.Errorf("expected positive base salary, got %d", s.BaseSalary)
		}
		if s.MonthlyCost < 0 {
			t.Errorf("expected non-negative monthly cost, got %.2f", s.MonthlyCost)
		}
		if s.ProjectedPension < 0 {
			t.Errorf("expected non-negative projected pension, got %.2f", s.ProjectedPension)
		}
	}
}

func TestGenerateSchemesAllWithinBudget(t *testing.T) {
	budget := 2000.0
	schemes := GenerateSchemes(GenerateInput{
		Age:              35,
		Gender:           "female",
		ContributionYears: 10,
		LocalAvgSalary:   8000,
		CurrentBalance:    20000,
		MonthlyBudget:    budget,
		PensionAge:       55,
	})

	for _, s := range schemes {
		if s.MonthlyCost > budget*1.01 {
			t.Errorf("scheme %s: monthly cost %.2f exceeds budget %.2f",
				s.Name, s.MonthlyCost, budget)
		}
	}
}

func TestGenerateSchemesMinTiers(t *testing.T) {
	schemes := GenerateSchemes(GenerateInput{
		Age:              25,
		Gender:           "male",
		ContributionYears: 0,
		LocalAvgSalary:   5000,
		CurrentBalance:    0,
		MonthlyBudget:    2000,
		PensionAge:       60,
	})

	if len(schemes) < 3 {
		t.Errorf("expected at least 3 scheme tiers, got %d", len(schemes))
	}
}

func TestGenerateSchemesOrderedByCost(t *testing.T) {
	schemes := GenerateSchemes(GenerateInput{
		Age:              40,
		Gender:           "male",
		ContributionYears: 5,
		LocalAvgSalary:   12000,
		CurrentBalance:    10000,
		MonthlyBudget:    5000,
		PensionAge:       60,
	})

	for i := 1; i < len(schemes); i++ {
		if schemes[i].MonthlyCost < schemes[i-1].MonthlyCost {
			t.Errorf("schemes not ordered by cost: %s (%.2f) < %s (%.2f)",
				schemes[i].Name, schemes[i].MonthlyCost,
				schemes[i-1].Name, schemes[i-1].MonthlyCost)
		}
	}
}

func TestSchemeNameNotEmpty(t *testing.T) {
	schemes := GenerateSchemes(GenerateInput{
		Age:              30,
		Gender:           "male",
		ContributionYears: 10,
		LocalAvgSalary:   10000,
		CurrentBalance:    0,
		MonthlyBudget:    2000,
		PensionAge:       60,
	})

	for _, s := range schemes {
		if s.Name == "" {
			t.Error("scheme name should not be empty")
		}
	}
}

func almostEqual(a, b, eps float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < eps || math.IsNaN(a) && math.IsNaN(b)
}
