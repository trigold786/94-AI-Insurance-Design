package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type PolicyQuerier interface {
	QueryByRegionAndStatus(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error)
}

type ComplianceEvaluator struct{}

func (e *ComplianceEvaluator) Evaluate(user *models.UserProfile, policy *models.PolicyClaim) (isEligible bool, unmet []string) {
	if user == nil || policy == nil {
		return false, nil
	}

	for _, tag := range policy.TargetGroupTags {
		if e.matchTag(tag, user) {
			continue
		}
		if !e.isBuiltinTag(tag) {
			continue
		}
		unmet = append(unmet, tag)
	}

	if len(policy.Conditions) > 0 {
		var conditions []models.ComplianceCondition
		if err := json.Unmarshal(policy.Conditions, &conditions); err == nil {
			for _, c := range conditions {
				if c.TagMatch != "" && !e.matchTag(c.TagMatch, user) {
					unmet = append(unmet, c.Name)
				}
			}
		}
	}

	isEligible = len(unmet) == 0
	return
}

func (e *ComplianceEvaluator) matchTag(tag string, user *models.UserProfile) bool {
	switch tag {
	case "flexible_employment":
		return user.EmploymentStatus == "flexible"
	case "unemployed":
		return user.EmploymentStatus == "unemployed"
	case "employed":
		return user.EmploymentStatus == "employed"
	case "has_children":
		return user.HasChildren
	case "age_40_plus", "4050":
		return user.Age >= 40
	case "age_50_plus":
		return user.Age >= 50
	case "female":
		return user.Gender == "female"
	case "male":
		return user.Gender == "male"
	case "low_income":
		return false // requires income field on UserProfile (Phase 2)
	default:
		return false
	}
}

func (e *ComplianceEvaluator) isBuiltinTag(tag string) bool {
	builtin := map[string]bool{
		"flexible_employment": true, "unemployed": true, "employed": true,
		"has_children": true, "age_40_plus": true, "age_50_plus": true,
		"4050": true, "female": true, "male": true, "low_income": true,
	}
	return builtin[tag]
}

func ComplianceChecklistHandler(evaluator *ComplianceEvaluator, policyRepo PolicyQuerier, profileRepo ProfileRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		if userID == "" {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": "UNAUTHORIZED", "message": "missing user"})
			return
		}

		profile, err := profileRepo.GetByUserID(r.Context(), userID)
		if err != nil {
			respondError(w, err)
			return
		}

		cityCode := r.URL.Query().Get("city_code")
		if cityCode == "" {
			cityCode = profile.CurrentResidenceCode
			if cityCode == "" {
				cityCode = profile.HouseholdRegionCode
			}
		}
		if cityCode == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "city_code is required"})
			return
		}

		policies, err := policyRepo.QueryByRegionAndStatus(r.Context(), cityCode, "verified")
		if err != nil {
			respondError(w, err)
			return
		}

		today := time.Now().Truncate(24 * time.Hour)

		var matchedPolicies []models.PolicyCompliance
		allDocsMap := make(map[string]models.RequiredDocument)
		var eligibleTags []string

		for _, p := range policies {
			if p.EffectiveDate != "" {
				effectiveDate, parseErr := time.Parse("2006-01-02", p.EffectiveDate)
				if parseErr != nil {
					log.Printf("[compliance] failed to parse effective_date for %s: %v", p.ClaimID, parseErr)
					continue
				}
				if effectiveDate.After(today) {
					continue
				}
			}
			if p.ExpireDate != nil && *p.ExpireDate != "" {
				expireDate, parseErr := time.Parse("2006-01-02", *p.ExpireDate)
				if parseErr == nil && expireDate.Before(today) {
					continue
				}
			}

			pc := models.PolicyCompliance{
				PolicyID:          p.PolicyID,
				PolicyType:        p.PolicyType,
				ClaimID:           p.ClaimID,
				SubsidyCalcMethod: p.SubsidyCalcMethod,
			}

			if len(p.Conditions) > 0 {
				var conds []models.ComplianceCondition
				if err := json.Unmarshal(p.Conditions, &conds); err == nil {
					pc.Conditions = conds
				} else {
					log.Printf("[compliance] failed to parse conditions for %s: %v", p.ClaimID, err)
				}
			}

			if len(p.RequiredDocuments) > 0 {
				var docs []models.RequiredDocument
				if err := json.Unmarshal(p.RequiredDocuments, &docs); err == nil {
					pc.RequiredDocs = docs
					for _, d := range docs {
						allDocsMap[d.Name] = d
					}
				}
			}

			pc.ProcessingSteps = getProcessingSteps(p.PolicyType, pc.Conditions, pc.RequiredDocs)

			isEligible, unmet := evaluator.Evaluate(profile, &p)
			pc.IsEligible = isEligible
			pc.UnmetConditions = unmet

			if isEligible {
				eligibleTags = append(eligibleTags, p.TargetGroupTags...)
			}

			matchedPolicies = append(matchedPolicies, pc)
		}

		var allDocs []models.RequiredDocument
		for _, d := range allDocsMap {
			allDocs = append(allDocs, d)
		}

		result := models.ComplianceChecklist{
			UserID:          userID,
			CityCode:        cityCode,
			MatchedPolicies: matchedPolicies,
			RequiredDocs:    allDocs,
			EligibleTags:    uniqueStrings(eligibleTags),
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": result})
	})
}

func getProcessingSteps(policyType string, conditions []models.ComplianceCondition, requiredDocs []models.RequiredDocument) []models.ProcessingStep {
	var docDesc string
	if len(requiredDocs) > 0 {
		names := make([]string, 0, len(requiredDocs))
		for _, d := range requiredDocs {
			names = append(names, d.Name)
		}
		docDesc = "根据清单准备以下材料: " + strings.Join(names, "、")
	} else {
		docDesc = "根据清单准备所有必需材料，包括身份证、户口本、就业创业证等"
	}

	step1 := models.ProcessingStep{Order: 1, Name: "准备材料", Description: docDesc}
	step2 := models.ProcessingStep{Order: 2, Name: "提交申请", Description: "登录当地人社局官网或政务服务平台提交申请"}

	common := []models.ProcessingStep{step1, step2}

	switch policyType {
	case "subsidy":
		return append(common, []models.ProcessingStep{
			{Order: 3, Name: "审核公示", Description: "人社部门审核材料，符合条件者进行公示"},
			{Order: 4, Name: "补贴发放", Description: "审核通过后，补贴资金发放至个人账户"},
		}...)
	case "pension":
		return append(common, []models.ProcessingStep{
			{Order: 3, Name: "参保登记", Description: "携带身份证到社保经办机构或通过线上平台办理灵活就业人员参保登记"},
			{Order: 4, Name: "选择缴费基数", Description: "根据经济能力选择60%-300%的缴费基数档次"},
			{Order: 5, Name: "按月缴费", Description: "每月按时足额缴纳养老保险费"},
			{Order: 6, Name: "待遇申领", Description: "达到法定退休年龄且缴费满15年，申领养老金"},
		}...)
	case "medical":
		return append(common, []models.ProcessingStep{
			{Order: 3, Name: "参保登记", Description: "办理灵活就业人员职工医保或城乡居民医保参保"},
			{Order: 4, Name: "缴费确认", Description: "按月缴纳医疗保险费，确认医保待遇生效"},
			{Order: 5, Name: "就医报销", Description: "在定点医疗机构就医，持社保卡直接结算"},
		}...)
	case "training":
		return append(common, []models.ProcessingStep{
			{Order: 3, Name: "报名培训", Description: "选择政府认定的培训机构报名技能培训课程"},
			{Order: 4, Name: "完成培训", Description: "按时参加培训并通过考核"},
			{Order: 5, Name: "申领补贴", Description: "凭培训合格证书向人社部门申请培训补贴"},
		}...)
	default:
		return common
	}
}

func uniqueStrings(s []string) []string {
	seen := map[string]bool{}
	var r []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			r = append(r, v)
		}
	}
	return r
}
