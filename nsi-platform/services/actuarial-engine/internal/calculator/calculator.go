package calculator

import "fmt"

type Scheme struct {
	Name            string  `json:"name"`
	BaseSalary      int     `json:"base_salary"`
	MonthlyCost     float64 `json:"monthly_cost"`
	AnnualSubsidy   float64 `json:"annual_subsidy"`
	ProjectedPension float64 `json:"projected_pension"`
}

type GenerateInput struct {
	Age              int
	Gender           string
	ContributionYears int
	LocalAvgSalary   float64
	CurrentBalance   float64
	MonthlyBudget    float64
	PensionAge       int
}

func CalculateBasicPension(localAvgSalary, indexedAvgSalary float64, years int) float64 {
	if years <= 0 || localAvgSalary <= 0 || indexedAvgSalary <= 0 {
		return 0
	}
	return (localAvgSalary + indexedAvgSalary) / 2 * float64(years) * 0.01
}

func CalculatePersonalAccountPension(balance float64, retireAge int) float64 {
	if balance <= 0 {
		return 0
	}
	divisor := getDivisor(retireAge)
	return balance / divisor
}

func CalculateMonthlyContribution(salary float64, pensionRate, medicalRate float64) float64 {
	if salary <= 0 {
		return 0
	}
	return salary * (pensionRate + medicalRate)
}

func GenerateSchemes(input GenerateInput) []Scheme {
	tiers := []struct {
		name    string
		pct     float64
	}{
		{"最低基数", 0.6},
		{"中等基数", 1.0},
		{"最高基数", 3.0},
	}

	pensionRate := 0.08
	medicalRate := 0.02

	var schemes []Scheme
	for _, tier := range tiers {
		baseSalary := int(float64(input.LocalAvgSalary) * tier.pct)
		if baseSalary <= 0 {
			continue
		}

		monthlyCost := CalculateMonthlyContribution(float64(baseSalary), pensionRate, medicalRate)
		if monthlyCost > input.MonthlyBudget {
			continue
		}

		basicPension := CalculateBasicPension(input.LocalAvgSalary, float64(baseSalary), input.ContributionYears)
		projectedBalance := projectAccountBalance(input.CurrentBalance, float64(baseSalary), input.Age, input.PensionAge)
		personalPension := CalculatePersonalAccountPension(projectedBalance, input.PensionAge)

		projectedPension := basicPension + personalPension

		schemes = append(schemes, Scheme{
			Name:            fmt.Sprintf("%s (%.0f%%)", tier.name, tier.pct*100),
			BaseSalary:      baseSalary,
			MonthlyCost:     monthlyCost,
			AnnualSubsidy:   0,
			ProjectedPension: projectedPension,
		})
	}

	if schemes == nil {
		schemes = []Scheme{}
	}

	return schemes
}

func projectAccountBalance(currentBalance, baseSalary float64, currentAge, pensionAge int) float64 {
	years := float64(pensionAge - currentAge)
	if years <= 0 {
		return currentBalance
	}
	monthlyContribution := baseSalary * 0.08
	annualContribution := monthlyContribution * 12
	rate := 0.03

	balance := currentBalance
	for y := 0; y < int(years); y++ {
		balance = balance*(1+rate) + annualContribution
	}
	return balance
}

func getDivisor(age int) float64 {
	switch {
	case age >= 60:
		return 139
	case age >= 55:
		return 170
	case age >= 50:
		return 195
	default:
		return 139
	}
}
