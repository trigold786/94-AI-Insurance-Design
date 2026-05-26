package calculator

import (
	"fmt"
	"strings"
)

type Scheme struct {
	Name                  string  `json:"name"`
	BaseSalary            int     `json:"base_salary"`
	MonthlyCost           float64 `json:"monthly_cost"`
	AnnualSubsidy         float64 `json:"annual_subsidy"`
	SubsidyPolicy         string  `json:"subsidy_policy"`
	SubsidyCondition      string  `json:"subsidy_condition"`
	PaidMonths            int     `json:"paid_months"`
	TargetMonths          int     `json:"target_months"`
	RemainingMonths       int     `json:"remaining_months"`
	TotalPersonalCost     float64 `json:"total_personal_cost"`
	RemainingPersonalCost float64 `json:"remaining_personal_cost"`
	ProjectedPension      float64 `json:"projected_pension"`
}

type GenerateInput struct {
	Age                  int
	Gender               string
	Employment           string
	ContributionYears    int
	ContributionMonths   int
	LocalAvgSalary       float64
	CurrentBalance       float64
	MonthlyBudget        float64
	PensionAge           int
	PensionEmployeeRate  float64
	PensionEmployerRate  float64
	MedicalEmployeeRate  float64
	MedicalEmployerRate  float64
	ContributionBaseMin  float64
	ContributionBaseMax  float64
	SubsidyCalcMethod    string
}

func generateSalaryTiers(localAvgSalary, baseMin, baseMax float64) []int {
	if baseMin <= 0 && baseMax <= 0 {
		pcts := []float64{0.6, 1.0, 3.0}
		var tiers []int
		for _, p := range pcts {
			v := int(float64(localAvgSalary) * p)
			if v > 0 {
				tiers = append(tiers, v)
			}
		}
		return tiers
	}
	if baseMin <= 0 {
		baseMin = 1
	}
	if baseMax < baseMin {
		baseMax = baseMin
	}
	start := int(baseMin)
	if start%100 != 0 {
		start = (start/100 + 1) * 100
	}
	end := int(baseMax)
	var tiers []int
	for s := start; s <= end; s += 100 {
		tiers = append(tiers, s)
	}
	if len(tiers) == 0 {
		tiers = append(tiers, start)
	}
	return tiers
}

func parseSubsidyRateFromMethod(method string) float64 {
	if method == "" {
		return 0
	}
	parts := strings.Split(method, "+")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "%"); idx >= 0 {
			start := strings.LastIndex(p[:idx], "*")
			if start >= 0 {
				s := p[start+1 : idx]
				var r float64
				if _, err := fmt.Sscanf(s, "%f", &r); err == nil {
					return r / 100
				}
			}
		}
	}
	return 0
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
	pensionRate := input.PensionEmployeeRate
	if pensionRate <= 0 {
		pensionRate = 0.08
	}
	medicalRate := input.MedicalEmployeeRate
	if medicalRate <= 0 {
		medicalRate = 0.02
	}
	employerPensionRate := input.PensionEmployerRate
	if employerPensionRate <= 0 {
		employerPensionRate = 0.16
	}

	subsidyPolicy := ""
	subsidyCondition := ""
	subsidyRate := 1.0
	switch input.Employment {
	case "flexible":
		subsidyRate = 0.5
		subsidyPolicy = "灵活就业人员社会保险补贴政策"
		subsidyCondition = "已办理灵活就业登记，女性40岁以上/男性50岁以上可享受更高补贴比例"
	case "unemployed":
		subsidyRate = 0.3
		subsidyPolicy = "失业人员社保补贴 + 职业技能培训补贴"
		subsidyCondition = "已办理失业登记，正在领取失业保险金期间"
	case "employed":
		subsidyRate = 0.0
		subsidyPolicy = "单位职工社保（单位缴纳统筹部分）"
		subsidyCondition = "由所在单位依法缴纳"
	default:
		subsidyRate = 0.0
	}

	if input.SubsidyCalcMethod != "" {
		if parsed := parseSubsidyRateFromMethod(input.SubsidyCalcMethod); parsed > 0 {
			subsidyRate = parsed
			subsidyPolicy = "政策性补贴（动态计算）"
			subsidyCondition = "根据匹配政策自动计算补贴比例"
		}
	}

	paidMonths := input.ContributionMonths
	if paidMonths <= 0 {
		paidMonths = input.ContributionYears * 12
	}
	if paidMonths < 0 {
		paidMonths = 0
	}
	targetMonths := 240
	remainingMonths := targetMonths - paidMonths
	if remainingMonths < 0 {
		remainingMonths = 0
	}
	totalYears := (paidMonths + remainingMonths) / 12

	baseSalaries := generateSalaryTiers(input.LocalAvgSalary, input.ContributionBaseMin, input.ContributionBaseMax)

	var schemes []Scheme
	for _, baseSalary := range baseSalaries {
		if baseSalary <= 0 {
			continue
		}

		monthlyCost := CalculateMonthlyContribution(float64(baseSalary), pensionRate, medicalRate)
		if monthlyCost > input.MonthlyBudget {
			continue
		}

		projectedBalance := projectAccountBalance(input.CurrentBalance, float64(baseSalary), pensionRate, input.Age, input.PensionAge)
		personalPension := CalculatePersonalAccountPension(projectedBalance, input.PensionAge)
		totalBasic := CalculateBasicPension(input.LocalAvgSalary, float64(baseSalary), totalYears)
		projectedPension := totalBasic + personalPension
		annualSubsidy := float64(baseSalary) * employerPensionRate * 12 * subsidyRate
		personalCostPerMonth := monthlyCost * (1 - subsidyRate)
		remainingPersonalCost := personalCostPerMonth * float64(remainingMonths)
		totalPersonalCost := personalCostPerMonth * float64(paidMonths+remainingMonths)

		schemes = append(schemes, Scheme{
			Name:                  fmt.Sprintf("缴费基数 %d", baseSalary),
			BaseSalary:            baseSalary,
			MonthlyCost:           monthlyCost,
			AnnualSubsidy:         annualSubsidy,
			SubsidyPolicy:         subsidyPolicy,
			SubsidyCondition:      subsidyCondition,
			PaidMonths:            paidMonths,
			TargetMonths:          targetMonths,
			RemainingMonths:       remainingMonths,
			TotalPersonalCost:     totalPersonalCost,
			RemainingPersonalCost: remainingPersonalCost,
			ProjectedPension:      projectedPension,
		})
	}

	if schemes == nil {
		schemes = []Scheme{}
	}

	return schemes
}

func projectAccountBalance(currentBalance, baseSalary, personalRate float64, currentAge, pensionAge int) float64 {
	years := float64(pensionAge - currentAge)
	if years <= 0 {
		return currentBalance
	}
	monthlyContribution := baseSalary * personalRate
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
