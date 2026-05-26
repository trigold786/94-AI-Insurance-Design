package calculator

import (
	"math"
	"testing"
)

func TestGenerateSalaryTiers_Fallback(t *testing.T) {
	tiers := generateSalaryTiers(10000, 0, 0)
	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}
	expected := []int{6000, 10000, 30000}
	for i, v := range tiers {
		if v != expected[i] {
			t.Errorf("tier %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestGenerateSalaryTiers_Enumeration(t *testing.T) {
	tiers := generateSalaryTiers(0, 5000, 5400)
	if len(tiers) != 5 {
		t.Fatalf("expected 5 tiers for 5000-5400 step 100, got %d", len(tiers))
	}
	expected := []int{5000, 5100, 5200, 5300, 5400}
	for i, v := range tiers {
		if v != expected[i] {
			t.Errorf("tier %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestGenerateSalaryTiers_RoundUp(t *testing.T) {
	tiers := generateSalaryTiers(0, 5010, 5200)
	if len(tiers) < 2 {
		t.Fatalf("expected at least 2 tiers, got %d", len(tiers))
	}
	if tiers[0] != 5100 {
		t.Errorf("first tier should round up to 5100, got %d", tiers[0])
	}
	if tiers[len(tiers)-1] != 5200 {
		t.Errorf("last tier should be 5200, got %d", tiers[len(tiers)-1])
	}
}

func TestGenerateSalaryTiers_SingleValue(t *testing.T) {
	tiers := generateSalaryTiers(0, 5000, 5000)
	if len(tiers) != 1 || tiers[0] != 5000 {
		t.Errorf("expected [5000], got %v", tiers)
	}
}

func TestParseSubsidyRateFromMethod(t *testing.T) {
	tests := []struct {
		method string
		want   float64
	}{
		{"基数*50%", 0.5},
		{"基数*8%+基数*16%", 0.08},
		{"基数*30%", 0.3},
		{"", 0},
		{"基数*0%", 0},
	}
	for _, tc := range tests {
		got := parseSubsidyRateFromMethod(tc.method)
		if !almostEqual(got, tc.want, 0.001) {
			t.Errorf("parseSubsidyRateFromMethod(%q) = %.4f, want %.4f", tc.method, got, tc.want)
		}
	}
}

func TestGenerateSchemes_WithEnumeration(t *testing.T) {
	schemes := GenerateSchemes(GenerateInput{
		Age:                 30,
		Gender:              "male",
		Employment:          "employed",
		ContributionYears:   10,
		LocalAvgSalary:      10000,
		CurrentBalance:      50000,
		MonthlyBudget:       5000,
		PensionAge:          60,
		ContributionBaseMin: 5000,
		ContributionBaseMax: 5200,
	})
	if len(schemes) == 0 {
		t.Fatal("expected at least 1 scheme from enumeration")
	}
	for _, s := range schemes {
		if s.BaseSalary < 5000 || s.BaseSalary > 5200 {
			t.Errorf("base salary %d out of range [5000,5200]", s.BaseSalary)
		}
		if s.BaseSalary%100 != 0 {
			t.Errorf("base salary %d not a multiple of 100", s.BaseSalary)
		}
		if s.MonthlyCost > 5000 {
			t.Errorf("monthly cost %.2f exceeds budget", s.MonthlyCost)
		}
	}
}

func TestGenerateSchemes_WithSubsidyCalcMethod(t *testing.T) {
	schemes := GenerateSchemes(GenerateInput{
		Age:               30,
		Gender:            "female",
		Employment:        "flexible",
		ContributionYears: 15,
		LocalAvgSalary:    8000,
		CurrentBalance:    0,
		MonthlyBudget:     5000,
		PensionAge:        55,
		SubsidyCalcMethod: "基数*80%",
	})
	if len(schemes) == 0 {
		t.Fatal("expected at least 1 scheme with subsidy override")
	}
	for _, s := range schemes {
		if s.AnnualSubsidy <= 0 {
			t.Errorf("expected positive subsidy with 80%% rate, got %.2f", s.AnnualSubsidy)
		}
		if s.SubsidyPolicy != "政策性补贴（动态计算）" {
			t.Errorf("expected dynamic subsidy policy name, got %q", s.SubsidyPolicy)
		}
	}
}

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

func TestCalculateBasicPension_ZeroYears(t *testing.T) {
	if v := CalculateBasicPension(10000, 8000, 0); v != 0 {
		t.Fatalf("zero years should return 0, got %.2f", v)
	}
}

func TestCalculateBasicPension_NegativeSalary(t *testing.T) {
	if v := CalculateBasicPension(-100, 8000, 15); v != 0 {
		t.Fatalf("negative salary should return 0, got %.2f", v)
	}
}

func TestCalculateMonthlyContribution_ZeroSalary(t *testing.T) {
	if v := CalculateMonthlyContribution(0, 0.08, 0.02); v != 0 {
		t.Fatalf("zero salary should return 0, got %.2f", v)
	}
}

func TestGenerateSchemes_ZeroBudget(t *testing.T) {
	schemes := GenerateSchemes(GenerateInput{
		Age: 30, Gender: "male", ContributionYears: 15,
		LocalAvgSalary: 10000, CurrentBalance: 0,
		MonthlyBudget: 0, PensionAge: 60,
	})
	if len(schemes) != 0 {
		t.Fatalf("zero budget should produce 0 schemes, got %d", len(schemes))
	}
}

func TestGenerateSchemes_AlreadyRetired(t *testing.T) {
	schemes := GenerateSchemes(GenerateInput{
		Age: 65, Gender: "male", ContributionYears: 0,
		LocalAvgSalary: 10000, CurrentBalance: 50000,
		MonthlyBudget: 5000, PensionAge: 60,
	})
	if len(schemes) == 0 {
		t.Fatal("already retired with high budget should still produce schemes")
	}
}

func TestProjectAccountBalance_NegativeYears(t *testing.T) {
	v := projectAccountBalance(10000, 5000, 0.08, 60, 55)
	if v != 10000 {
		t.Fatalf("past pension age should return current balance, got %.2f", v)
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
