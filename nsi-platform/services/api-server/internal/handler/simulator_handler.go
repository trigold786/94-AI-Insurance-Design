package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type SimulatorRequest struct {
	CityCode     string  `json:"city_code"`
	Gender       string  `json:"gender"`
	Age          int     `json:"age"`
	BasePercent  int     `json:"base_percent"`
	PaidYears    int     `json:"paid_years"`
	PlanYears    int     `json:"plan_years"`
	Employment   string  `json:"employment"`
	IsLocalHukou bool    `json:"is_local_hukou"`
}

type SimulatorResponse struct {
	Cost           SimCost         `json:"cost"`
	Pension        SimPension      `json:"pension"`
	Subsidy        SimSubsidy      `json:"subsidy"`
	NetMonthly     float64         `json:"net_monthly"`
	Thresholds     SimThresholds   `json:"thresholds"`
	Qualifications []SimQual       `json:"qualifications"`
	Cashflow       []SimCashflow   `json:"cashflow"`
	Comparison     SimComparison   `json:"comparison"`
	BreakEvenAge   int             `json:"break_even_age"`
	PolicyTriggers []SimTrigger    `json:"policy_triggers"`
}

type SimCost struct {
	MonthlyTotal   float64 `json:"monthly_total"`
	MonthlyPension float64 `json:"monthly_pension"`
	MonthlyMedical float64 `json:"monthly_medical"`
	AnnualTotal    float64 `json:"annual_total"`
}
type SimPension struct {
	ProjectedMonthly    float64 `json:"projected_monthly"`
	PersonalAccountTotal float64 `json:"personal_account_total"`
	BasePension         float64 `json:"base_pension"`
	AccountPension      float64 `json:"account_pension"`
}
type SimSubsidy struct {
	AnnualTotal float64          `json:"annual_total"`
	Items       []SimSubsidyItem `json:"items"`
}
type SimSubsidyItem struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	PolicyID string  `json:"policy_id"`
	ClaimID  string  `json:"claim_id"`
}
type SimThresholds struct {
	MinContributionYears float64 `json:"min_contribution_years"`
	RetirementYear       int     `json:"retirement_year"`
	MeetsMinYears        bool    `json:"meets_min_years"`
	YearsShortfall       float64 `json:"years_shortfall"`
}
type SimQual struct {
	Name       string  `json:"name"`
	Qualified  bool    `json:"qualified"`
	YearsUntil float64 `json:"years_until,omitempty"`
	Detail     string  `json:"detail"`
}
type SimCashflow struct {
	Year    int     `json:"year"`
	Payment float64 `json:"payment"`
	Subsidy float64 `json:"subsidy"`
	Net     float64 `json:"net"`
}
type SimComparison struct {
	At60  SimCompareItem `json:"at_60"`
	At100 SimCompareItem `json:"at_100"`
	At300 SimCompareItem `json:"at_300"`
}
type SimCompareItem struct {
	MonthlyCost      float64 `json:"monthly_cost"`
	ProjectedPension float64 `json:"projected_pension"`
}
type SimTrigger struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

var cityAvgSalary = map[string]float64{
	"310000": 12183,
	"110000": 11297,
	"440300": 13731,
	"440100": 11848,
	"330100": 10048,
}

func SimulatorHandler(
	thresholdResolver *ThresholdResolver,
	policyRepo PolicyQuerier,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, 405, map[string]interface{}{"code": "METHOD_NOT_ALLOWED"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req SimulatorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}
		validCities := map[string]bool{"310000": true, "110000": true, "440300": true, "440100": true, "330100": true}
		if !validCities[req.CityCode] {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid city_code"})
			return
		}
		if req.Gender != "male" && req.Gender != "female" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "gender must be male or female"})
			return
		}
		if req.Age < 16 || req.Age > 70 {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "age must be 16-70"})
			return
		}
		if req.BasePercent < 60 || req.BasePercent > 300 {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "base_percent must be 60-300"})
			return
		}
		if req.PaidYears < 0 || req.PaidYears > 35 {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "paid_years must be 0-35"})
			return
		}
		if req.PlanYears < 0 || req.PlanYears > 35 {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "plan_years must be 0-35"})
			return
		}

		resp := calculateSimulation(r.Context(), req, thresholdResolver, policyRepo)
		respondJSON(w, 200, map[string]interface{}{"code": 0, "data": resp})
	})
}

func calculateSimulation(ctx context.Context, req SimulatorRequest, tr *ThresholdResolver, policyRepo PolicyQuerier) SimulatorResponse {
	avgSalary := cityAvgSalary[req.CityCode]
	if avgSalary == 0 {
		avgSalary = 10000
	}
	baseSalary := avgSalary * float64(req.BasePercent) / 100

	pensionRate := 0.20
	medicalRate := 0.08
	monthlyPensionCost := baseSalary * pensionRate
	monthlyMedicalCost := baseSalary * medicalRate
	monthlyTotal := monthlyPensionCost + monthlyMedicalCost

	totalYears := req.PaidYears + req.PlanYears
	retAge := 60
	if req.Gender == "female" {
		retAge = 55
	}
	retirementYear := currentYear() + (retAge - req.Age)
	if retirementYear < currentYear() {
		retirementYear = currentYear()
	}

	minYears, thresholdDetail := float64(15), ""
	if tr != nil {
		minYears, thresholdDetail = tr.ResolveMinContributionYears(ctx, req.CityCode, retirementYear)
	}

	personalAccountRate := 0.08
	personalAccountTotal := baseSalary * personalAccountRate * 12 * float64(totalYears)
	accountPensionMonthly := 0.0
	if retAge > 0 {
		divisor := 170.0
		if req.Gender == "male" {
			divisor = 139.0
		}
		accountPensionMonthly = personalAccountTotal / divisor
	}
	avgBaseSalary := avgSalary * 0.6
	basePensionMonthly := (avgBaseSalary + baseSalary) / 2 * float64(totalYears) * 0.01
	projectedPension := basePensionMonthly + accountPensionMonthly
	if projectedPension < 0 {
		projectedPension = 0
	}

	subsidyItems := []SimSubsidyItem{}
	annualSubsidy := 0.0
	if req.Employment == "flexible" || req.Employment == "unemployed" {
		if policies, err := policyRepo.QueryByRegionAndStatus(ctx, req.CityCode, "verified"); err == nil {
			for _, p := range policies {
				if p.PolicyType == "subsidy" && p.SubsidyAmountMax != nil && *p.SubsidyAmountMax > 0 {
					qualified := true
					for _, tag := range p.TargetGroupTags {
						if tag == "4050" {
							if req.Gender == "female" && req.Age < 40 {
								qualified = false
							}
							if req.Gender == "male" && req.Age < 50 {
								qualified = false
							}
						}
						if tag == "flexible_employment" && req.Employment != "flexible" {
							qualified = false
						}
						if tag == "unemployed" && req.Employment != "unemployed" {
							qualified = false
						}
					}
					if qualified {
						amount := *p.SubsidyAmountMax
						subsidyItems = append(subsidyItems, SimSubsidyItem{
							Name: p.PolicyTitle, Amount: amount,
							PolicyID: p.PolicyID, ClaimID: p.ClaimID,
						})
						annualSubsidy += amount
					}
				}
			}
		}
	}

	netMonthly := monthlyTotal - annualSubsidy/12
	if netMonthly < 0 {
		netMonthly = 0
	}

	meetsMin := float64(totalYears) >= minYears
	shortfall := 0.0
	if !meetsMin {
		shortfall = minYears - float64(totalYears)
	}

	qualifications := calculateQualifications(req, totalYears, retirementYear)
	cashflow := calculateCashflow(req, baseSalary, monthlyTotal, annualSubsidy, retAge)
	comparison := calculateComparison(avgSalary, totalYears, req.Gender)
	breakEven := calculateBreakEven(baseSalary, monthlyTotal, annualSubsidy, projectedPension, req.Age, retAge)

	var triggers []SimTrigger
	if !meetsMin {
		triggers = append(triggers, SimTrigger{
			Type: "threshold_warning", Severity: "warning",
			Message: fmt.Sprintf("您计划总缴费%d年，%d年退休需%.1f年，还差%.1f年", totalYears, retirementYear, minYears, shortfall),
		})
	} else {
		triggers = append(triggers, SimTrigger{
			Type: "threshold_ok", Severity: "info",
			Message: fmt.Sprintf("%s，当前计划%d年✅", thresholdDetail, totalYears),
		})
	}
	if req.Gender == "female" && req.Age >= 40 {
		triggers = append(triggers, SimTrigger{
			Type: "policy_trigger", Severity: "success",
			Message: fmt.Sprintf("💡 您已符合4050社保补贴条件，每年可省约¥%.0f", annualSubsidy),
		})
	} else if req.Gender == "female" && req.Age >= 35 {
		triggers = append(triggers, SimTrigger{
			Type: "policy_trigger", Severity: "info",
			Message: fmt.Sprintf("💡 再过%d年（满40岁）即可申请4050社保补贴", 40-req.Age),
		})
	}
	if req.Gender == "male" && req.Age >= 50 {
		triggers = append(triggers, SimTrigger{
			Type: "policy_trigger", Severity: "success",
			Message: fmt.Sprintf("💡 您已符合4050社保补贴条件，每年可省约¥%.0f", annualSubsidy),
		})
	} else if req.Gender == "male" && req.Age >= 45 {
		triggers = append(triggers, SimTrigger{
			Type: "policy_trigger", Severity: "info",
			Message: fmt.Sprintf("💡 再过%d年（满50岁）即可申请4050社保补贴", 50-req.Age),
		})
	}

	return SimulatorResponse{
		Cost: SimCost{
			MonthlyTotal: round2(monthlyTotal), MonthlyPension: round2(monthlyPensionCost),
			MonthlyMedical: round2(monthlyMedicalCost), AnnualTotal: round2(monthlyTotal * 12),
		},
		Pension: SimPension{
			ProjectedMonthly: round2(projectedPension), PersonalAccountTotal: round2(personalAccountTotal),
			BasePension: round2(basePensionMonthly), AccountPension: round2(accountPensionMonthly),
		},
		Subsidy: SimSubsidy{AnnualTotal: round2(annualSubsidy), Items: subsidyItems},
		NetMonthly: round2(netMonthly),
		Thresholds: SimThresholds{
			MinContributionYears: minYears, RetirementYear: retirementYear,
			MeetsMinYears: meetsMin, YearsShortfall: round2(shortfall),
		},
		Qualifications: qualifications,
		Cashflow:       cashflow,
		Comparison:     comparison,
		BreakEvenAge:   breakEven,
		PolicyTriggers: triggers,
	}
}

func calculateQualifications(req SimulatorRequest, totalYears int, retirementYear int) []SimQual {
	var quals []SimQual

	purchaseYears := 5
	if req.CityCode == "310000" || req.CityCode == "110000" || req.CityCode == "440100" {
		purchaseYears = 5
	}
	if req.CityCode == "440300" {
		purchaseYears = 3
	}
	qualified := totalYears >= purchaseYears
	detail := fmt.Sprintf("已缴%d年≥%d年要求", totalYears, purchaseYears)
	if !qualified {
		detail = fmt.Sprintf("已缴%d年，需%d年，差%d年", totalYears, purchaseYears, purchaseYears-totalYears)
	}
	quals = append(quals, SimQual{Name: "购房资格", Qualified: qualified, Detail: detail})

	hukouPoints := totalYears * 3
	quals = append(quals, SimQual{
		Name: "落户积分", Qualified: hukouPoints >= 60, Detail: fmt.Sprintf("社保积分%d分（每年3分×%d年）", hukouPoints, totalYears),
	})

	age4050 := 40
	if req.Gender == "male" {
		age4050 = 50
	}
	if req.Age >= age4050 {
		quals = append(quals, SimQual{Name: "4050补贴", Qualified: true, Detail: "已符合年龄条件"})
	} else {
		quals = append(quals, SimQual{Name: "4050补贴", Qualified: false, YearsUntil: float64(age4050 - req.Age), Detail: fmt.Sprintf("还需%d年", age4050-req.Age)})
	}

	return quals
}

func calculateCashflow(req SimulatorRequest, baseSalary, monthlyTotal, annualSubsidy float64, retAge int) []SimCashflow {
	var cf []SimCashflow
	years := retAge - req.Age
	if years > 30 {
		years = 30
	}
	if years < 1 {
		years = 1
	}
	for i := 0; i < years; i++ {
		growthRate := math.Pow(1.05, float64(i))
		payment := monthlyTotal * 12 * growthRate
		subsidy := annualSubsidy
		if subsidy > 0 {
			subsidy *= growthRate
		}
		cf = append(cf, SimCashflow{
			Year: currentYear() + i,
			Payment: round2(payment), Subsidy: round2(subsidy),
			Net: round2(payment - subsidy),
		})
	}
	return cf
}

func calculateComparison(avgSalary float64, totalYears int, gender string) SimComparison {
	calc := func(pct int) SimCompareItem {
		base := avgSalary * float64(pct) / 100
		monthlyCost := base * 0.28
		personalAccount := base * 0.08 * 12 * float64(totalYears)
		divisor := 170.0
		if gender == "male" {
			divisor = 139.0
		}
		accountPension := personalAccount / divisor
		basePension := (avgSalary*0.6 + base) / 2 * float64(totalYears) * 0.01
		return SimCompareItem{
			MonthlyCost: round2(monthlyCost), ProjectedPension: round2(basePension + accountPension),
		}
	}
	return SimComparison{At60: calc(60), At100: calc(100), At300: calc(300)}
}

func calculateBreakEven(baseSalary, monthlyTotal, annualSubsidy, projectedPension float64, currentAge, retAge int) int {
	totalInvested := monthlyTotal * 12 * float64(retAge-currentAge)
	totalSubsidy := annualSubsidy * float64(retAge-currentAge)
	netInvested := totalInvested - totalSubsidy
	if projectedPension <= 0 {
		return 0
	}
	annualPension := projectedPension * 12
	if annualPension <= 0 {
		return 0
	}
	yearsToBreakEven := netInvested / annualPension
	breakAge := retAge + int(math.Ceil(yearsToBreakEven))
	if breakAge > 100 {
		breakAge = 100
	}
	return breakAge
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func SimulatorScenarioSaveHandler(repo SimulatorScenarioRepo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, 405, map[string]interface{}{"code": "METHOD_NOT_ALLOWED"})
			return
		}
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Name   string          `json:"name"`
			Params SimulatorRequest `json:"params"`
			Result json.RawMessage  `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR"})
			return
		}
		if req.Name == "" {
			req.Name = "方案" + strconv.Itoa(int(time.Now().UnixNano()%3+65))
		}
		paramsJSON, _ := json.Marshal(req.Params)
		resultJSON := req.Result
		if len(resultJSON) == 0 {
			resultJSON = json.RawMessage(`{}`)
		}
		if err := repo.SaveScenario(r.Context(), userID, req.Name, paramsJSON, resultJSON); err != nil {
			log.Printf("[simulator] failed to save scenario for %s: %v", userID, err)
			respondError(w, err)
			return
		}
		respondJSON(w, 200, map[string]interface{}{"code": 0, "message": "方案已保存"})
	})
}

func SimulatorScenarioListHandler(repo SimulatorScenarioRepo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		scenarios, err := repo.ListScenarios(r.Context(), userID)
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, 200, map[string]interface{}{"code": 0, "data": scenarios})
	})
}

type SimulatorScenarioRepo interface {
	SaveScenario(ctx context.Context, userID, name string, params, result json.RawMessage) error
	ListScenarios(ctx context.Context, userID string) ([]models.SimScenario, error)
}
