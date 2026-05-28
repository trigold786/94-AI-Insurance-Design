package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type PlanRepository interface {
	Save(ctx context.Context, plan *models.PlanSnapshot) error
	GetByID(ctx context.Context, planID string) (*models.PlanSnapshot, error)
}

type ProfileLookuper interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
}

type LLMGatewayClient struct {
	URL string
}

type LLMChatRequest struct {
	SystemPrompt string `json:"system_prompt"`
	UserContent  string `json:"user_content"`
	MaxTokens    int    `json:"max_tokens"`
	Caller       string `json:"caller"`
}

type LLMChatResponse struct {
	Code         int    `json:"code"`
	Content      string `json:"content"`
	ProviderUsed string `json:"provider_used"`
	LatencyMs    int64  `json:"latency_ms"`
}

func (c *LLMGatewayClient) Chat(ctx context.Context, systemPrompt, userContent string) (*LLMChatResponse, error) {
	body, _ := json.Marshal(LLMChatRequest{
		SystemPrompt: systemPrompt,
		UserContent:  userContent,
		MaxTokens:    8192,
		Caller:       "api-server",
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/v1/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm-gateway call failed: %w", err)
	}
	defer resp.Body.Close()
	var llmResp LLMChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("llm-gateway error: %s", llmResp.Content)
	}
	return &llmResp, nil
}

func buildPlanSystemPrompt() string {
	return `你是一位资深社保政策顾问。根据用户的个人情况和所在地的社保政策，为用户量身定制最优社保参保方案。

规则：
1. 所有政策依据必须来自提供的政策库数据，不可编造
2. 必须理解上位法与下位法的关系：国家法律 > 省级规定 > 市级细则 > 区级优惠
3. 当地方政策与上位法冲突时，以有利于用户的原则解释
4. 方案必须包含所有适用的优惠政策，不能遗漏
5. 每一条建议都必须标注所依据的政策（标题+文号+链接）
6. 数值计算必须精确，基于提供的费率和基数

输出格式（必须严格遵守）：
===FREE_FORM_START===
[自由文本格式的方案建议书，面向普通用户，通俗易懂，2000字以内]
===FREE_FORM_END===
===STRUCTURED_START===
{
  "summary": "一句话总结",
  "schemes": [
    {
      "name": "方案名称",
      "description": "方案描述",
      "monthly_cost": 0,
      "annual_subsidy": 0,
      "projected_pension": 0,
      "total_cost": 0,
      "contribution_base": 0,
      "pension_employee_rate": 0,
      "pension_employer_rate": 0,
      "medical_employee_rate": 0,
      "analysis": "该方案的详细分析",
      "applicable_policies": ["claim_id_1"]
    }
  ],
  "policy_references": [
    {
      "claim_id": "",
      "policy_title": "",
      "document_number": "",
      "policy_url": "",
      "relevant_excerpt": "提取的原文片段",
      "how_applied": "如何应用于本方案"
    }
  ],
  "recommendation": {
    "recommended_scheme": "方案名称",
    "reasoning": "推荐理由"
  }
}
===STRUCTURED_END===`
}

func buildPlanUserContent(profile *models.UserProfile, policies []models.PolicyClaim) string {
	policyJSON, _ := json.Marshal(policies)
	profileJSON, _ := json.Marshal(profile)
	return fmt.Sprintf(`## 用户画像
%s

## 适用政策（共%d条）
%s

## 请基于以上信息，为该用户生成最优社保参保方案。`, string(profileJSON), len(policies), string(policyJSON))
}

func parseLLMResponse(raw string) (freeForm string, structured *models.LLMSchemeResponse, err error) {
	freeStart := strings.Index(raw, "===FREE_FORM_START===")
	freeEnd := strings.Index(raw, "===FREE_FORM_END===")
	if freeStart >= 0 && freeEnd > freeStart {
		freeForm = strings.TrimSpace(raw[freeStart+len("===FREE_FORM_START===") : freeEnd])
	}
	structStart := strings.Index(raw, "===STRUCTURED_START===")
	structEnd := strings.Index(raw, "===STRUCTURED_END===")
	if structStart >= 0 && structEnd > structStart {
		jsonStr := strings.TrimSpace(raw[structStart+len("===STRUCTURED_START===") : structEnd])
		var resp models.LLMSchemeResponse
		if jsonErr := json.Unmarshal([]byte(jsonStr), &resp); jsonErr != nil {
			return freeForm, nil, fmt.Errorf("parse structured JSON: %w", jsonErr)
		}
		structured = &resp
	}
	if freeForm == "" && structured == nil {
		return "", nil, fmt.Errorf("LLM response missing required markers")
	}
	return freeForm, structured, nil
}

func GeneratePlanHandler(llmGatewayURL string, repo PlanRepository, profileRepo ProfileLookuper, policyRepo PolicyQuerier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Age           int     `json:"age"`
			Gender        string  `json:"gender"`
			Employment    string  `json:"employment"`
			MonthlyBudget float64 `json:"monthly_budget"`
			Priority      string  `json:"priority"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}
		if req.Age < 16 || req.Age > 70 {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "age must be between 16 and 70"})
			return
		}
		if req.Gender != "male" && req.Gender != "female" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "gender must be male or female"})
			return
		}

		var profile *models.UserProfile
		if profileRepo != nil {
			if p, err := profileRepo.GetByUserID(r.Context(), userID); err == nil {
				profile = p
			}
		}
		if profile == nil {
			profile = &models.UserProfile{
				UserID:           userID,
				Age:              req.Age,
				Gender:           req.Gender,
				EmploymentStatus: req.Employment,
			}
		}

		var policies []models.PolicyClaim
		code := profile.CurrentResidenceCode
		if code == "" {
			code = profile.HouseholdRegionCode
		}
		if policyRepo != nil && code != "" {
			policies, _ = policyRepo.QueryByRegionAndStatus(r.Context(), code, "verified")
		}
		if policies == nil {
			policies = []models.PolicyClaim{}
		}

		if llmGatewayURL == "" {
			respondJSON(w, 500, map[string]interface{}{"code": "CONFIG_ERROR", "message": "LLM gateway not configured"})
			return
		}
		client := &LLMGatewayClient{URL: llmGatewayURL}
		systemPrompt := buildPlanSystemPrompt()
		userContent := buildPlanUserContent(profile, policies)
		llmResp, err := client.Chat(r.Context(), systemPrompt, userContent)
		if err != nil {
			respondError(w, err)
			return
		}

		freeForm, structured, err := parseLLMResponse(llmResp.Content)
		if err != nil {
			respondError(w, err)
			return
		}

		snapshot := &models.PlanSnapshot{
			PlanID:       fmt.Sprintf("plan-%d", time.Now().UnixNano()),
			UserID:       userID,
			FreeFormText: freeForm,
			GeneratedAt:  time.Now(),
		}

		if structured != nil {
			snapshot.StructuredSchemes = structured.Schemes
			snapshot.PolicyReferences = structured.PolicyReferences
			snapshot.Recommendation = structured.Recommendation.RecommendedScheme
			snapshot.RecommendationReason = structured.Recommendation.Reasoning

			var totalCost, totalSubsidy float64
			for _, s := range structured.Schemes {
				totalCost += s.TotalCost
				totalSubsidy += s.AnnualSubsidy
			}
			snapshot.TotalCost = totalCost
			snapshot.TotalSubsidy = totalSubsidy
		}

		if repo != nil {
			if err := repo.Save(r.Context(), snapshot); err != nil {
				respondError(w, err)
				return
			}
		}

		respondJSON(w, 200, map[string]interface{}{"code": 0, "data": snapshot})
	})
}

func PlanDetailHandler(repo PlanRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		planID := strings.TrimPrefix(r.URL.Path, "/v1/plans/")
		if planID == "" || strings.Contains(planID, "/") {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid plan_id"})
			return
		}

		plan, err := repo.GetByID(r.Context(), planID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				respondJSON(w, http.StatusNotFound, map[string]interface{}{"code": "NOT_FOUND", "message": err.Error()})
				return
			}
			respondError(w, err)
			return
		}

		if plan.UserID != userID {
			respondJSON(w, http.StatusNotFound, map[string]interface{}{"code": "NOT_FOUND", "message": "plan not found"})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": plan})
	})
}

func calcPensionAge(dob, gender string, originalPensionAge int) int {
	baseAge := 60
	if gender == "female" {
		baseAge = 55
		if originalPensionAge >= 50 && originalPensionAge <= 60 {
			baseAge = originalPensionAge
		}
	}
	var pace, maxDelay int
	switch baseAge {
	case 50:
		pace = 2
		maxDelay = 60
	case 55:
		pace = 4
		maxDelay = 36
	default:
		pace = 4
		maxDelay = 36
	}
	if len(dob) < 7 {
		return baseAge + maxDelay/12
	}
	parts := strings.Split(dob, "-")
	if len(parts) < 2 {
		return baseAge + maxDelay/12
	}
	by, _ := strconv.Atoi(parts[0])
	bm, _ := strconv.Atoi(parts[1])
	baseYear := 2025 - baseAge
	monthsSincePolicy := (by-baseYear)*12 + bm - 1
	if monthsSincePolicy < 0 {
		return baseAge
	}
	delay := monthsSincePolicy/pace + 1
	if delay > maxDelay {
		delay = maxDelay
	}
	return baseAge + delay/12
}

func parseRate(method, label string) float64 {
	if method == "" {
		return 0
	}
	var rates []float64
	parts := strings.Split(method, "+")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "%"); idx >= 0 {
			start := strings.LastIndex(p[:idx], "*")
			if start >= 0 {
				s := p[start+1 : idx]
				var r float64
				if _, err := fmt.Sscanf(s, "%f", &r); err == nil {
					rates = append(rates, r/100)
				}
			}
		}
	}
	if label == "单位" && len(rates) >= 2 {
		return rates[1]
	}
	if len(rates) >= 1 {
		return rates[0]
	}
	return 0
}
