package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type ThresholdResolver struct {
	policyRepo PolicyThresholdQuerier
}

type PolicyThresholdQuerier interface {
	QueryThresholds(ctx context.Context, regionCode, policyType string) ([]models.ThresholdData, error)
}

type GradualThreshold struct {
	BaseYear         int     `json:"base_year"`
	BaseValue        float64 `json:"base_value"`
	IncrementPerYear float64 `json:"increment_per_year"`
	MaxValue         float64 `json:"max_value"`
	MaxYear          int     `json:"max_year"`
}

func NewThresholdResolver(policyRepo PolicyThresholdQuerier) *ThresholdResolver {
	return &ThresholdResolver{policyRepo: policyRepo}
}

func (tr *ThresholdResolver) ResolveMinContributionYears(ctx context.Context, regionCode string, retirementYear int) (float64, string) {
	if tr.policyRepo == nil {
		if retirementYear >= 2039 {
			return 20, "根据渐进式延迟退休政策，2039年起最低缴费年限为20年"
		}
		if retirementYear >= 2030 {
			years := 15 + float64(retirementYear-2030)*0.5
			if years > 20 {
				years = 20
			}
			return years, fmt.Sprintf("根据渐进式延迟退休政策，%d年退休最低缴费年限为%.1f年", retirementYear, years)
		}
		return 15, "当前最低缴费年限为15年"
	}

	thresholds, err := tr.policyRepo.QueryThresholds(ctx, regionCode, "pension")
	if err != nil || len(thresholds) == 0 {
		return tr.fallbackMinYears(retirementYear)
	}

	for _, t := range thresholds {
		var conditions []map[string]interface{}
		if err := json.Unmarshal(t.Conditions, &conditions); err != nil {
			continue
		}
		for _, c := range conditions {
			if cType, ok := c["type"].(string); ok && cType == "gradual_min_years" {
				gt := parseGradualThreshold(c)
				return gt.Resolve(retirementYear), fmt.Sprintf("根据《%s》(%s)", c["description"], t.ClaimID)
			}
		}
	}

	return tr.fallbackMinYears(retirementYear)
}

func (tr *ThresholdResolver) fallbackMinYears(retirementYear int) (float64, string) {
	if retirementYear >= 2039 {
		return 20, "2039年起最低缴费年限20年"
	}
	if retirementYear >= 2030 {
		years := 15 + float64(retirementYear-2030)*0.5
		return years, fmt.Sprintf("%d年退休需缴费%.1f年", retirementYear, years)
	}
	return 15, "当前最低缴费年限15年"
}

func parseGradualThreshold(m map[string]interface{}) GradualThreshold {
	gt := GradualThreshold{
		BaseYear:         2030,
		BaseValue:        15,
		IncrementPerYear: 0.5,
		MaxValue:         20,
		MaxYear:          2039,
	}
	if v, ok := m["base_year"].(float64); ok {
		gt.BaseYear = int(v)
	}
	if v, ok := m["base_value"].(float64); ok {
		gt.BaseValue = v
	}
	if v, ok := m["increment_per_year"].(float64); ok {
		gt.IncrementPerYear = v
	}
	if v, ok := m["max_value"].(float64); ok {
		gt.MaxValue = v
	}
	if v, ok := m["max_year"].(float64); ok {
		gt.MaxYear = int(v)
	}
	return gt
}

func (gt GradualThreshold) Resolve(year int) float64 {
	if year <= gt.BaseYear {
		return gt.BaseValue
	}
	if year >= gt.MaxYear {
		return gt.MaxValue
	}
	result := gt.BaseValue + float64(year-gt.BaseYear)*gt.IncrementPerYear
	return math.Round(result*10) / 10
}

func currentYear() int {
	return time.Now().Year()
}
