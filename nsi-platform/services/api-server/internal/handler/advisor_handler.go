package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

func AdvisorHandler(llmGatewayURL string, policyRepo PolicyQuerier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondJSON(w, 405, map[string]interface{}{"code": "METHOD_NOT_ALLOWED"})
			return
		}

		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Question string          `json:"question"`
			Context  json.RawMessage `json:"context"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "question required"})
			return
		}

		var simCtx map[string]interface{}
		if len(req.Context) > 0 {
			json.Unmarshal(req.Context, &simCtx)
		}
		if simCtx == nil {
			simCtx = map[string]interface{}{}
		}
		simCtx["user_id"] = userID

		policyContext := searchRelevantPolicies(r, req.Question, simCtx, policyRepo)

		systemPrompt := `你是AI社保智筹平台的社保顾问。回答规则：
1. 回答不超过3句话，先给结论，再说依据
2. 必须引用政策来源（政策标题或文号）
3. 用通俗语言，不用专业术语
4. 如果不确定，说"建议咨询12333确认"
5. 不要编造政策，只基于提供的政策数据回答
6. 纯文本回答，不要使用markdown格式`

		userContent := fmt.Sprintf("## 用户当前沙盘参数\n%s%s\n\n## 用户问题\n%s",
			mustJSON(simCtx), policyContext, req.Question)

		llmResp, err := callLLMGateway(r.Context(), llmGatewayURL, systemPrompt, userContent)
		if err != nil {
			fallback := generatePolicyAnswer(req.Question, simCtx, policyRepo, r.Context())
			respondJSON(w, 200, map[string]interface{}{
				"code": 0,
				"data": map[string]string{"answer": fallback},
			})
			return
		}

		respondJSON(w, 200, map[string]interface{}{
			"code": 0,
			"data": map[string]string{"answer": truncateSentences(llmResp, 3)},
		})
	})
}

func searchRelevantPolicies(r *http.Request, question string, simCtx map[string]interface{}, policyRepo PolicyQuerier) string {
	cityCode, _ := simCtx["city_code"].(string)
	
	crawlerURL := os.Getenv("POLICY_CRAWER_URL")
	if crawlerURL != "" {
		searchURL := strings.TrimRight(crawlerURL, "/") + "/v1/policies/search?q=" + url.QueryEscape(question) + "&limit=5"
		if cityCode != "" {
			searchURL += "&region_code=" + cityCode
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(searchURL)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if len(body) > 10 {
				return fmt.Sprintf("\n## 语义检索相关政策（Top 5）\n%s", string(body))
			}
		}
	}

	if cityCode != "" {
		if policies, err := policyRepo.QueryByRegionAndStatus(r.Context(), cityCode, "verified"); err == nil && len(policies) > 0 {
			keywords := extractKeywords(question)
			scored := scorePoliciesByKeyword(policies, keywords)
			topN := scored
			if len(topN) > 5 {
				topN = topN[:5]
			}
			policyJSON, _ := json.Marshal(topN)
			return fmt.Sprintf("\n## 相关政策（Top 5）\n%s", string(policyJSON))
		}
	}
	return ""
}

func extractKeywords(q string) []string {
	stopWords := map[string]bool{"的": true, "了": true, "是": true, "在": true, "和": true, "与": true, "或": true, "会": true, "怎样": true, "怎么样": true, "如果": true, "可以": true, "能": true, "吗": true, "呢": true, "什么": true, "怎么": true, "多少": true, "哪些": true, "哪个": true, "有": true, "没有": true, "不": true, "要": true, "需": true, "想": true, "我": true, "你": true, "他": true, "她": true, "它": true}
	var keywords []string
	for _, w := range strings.Fields(q) {
		if !stopWords[w] && len(w) >= 2 {
			keywords = append(keywords, w)
		}
	}
	policyKeywords := []string{"养老", "医疗", "失业", "工伤", "生育", "社保", "补贴", "4050", "基数", "缴费", "退休", "断缴", "转移", "落户", "购房", "积分", "灵活就业", "失业登记"}
	for _, pk := range policyKeywords {
		if strings.Contains(q, pk) {
			keywords = append(keywords, pk)
		}
	}
	return keywords
}

func scorePoliciesByKeyword(policies []models.PolicyClaim, keywords []string) []models.PolicyClaim {
	type scored struct {
		policy models.PolicyClaim
		score  int
	}
	var scoredList []scored
	for _, p := range policies {
		s := 0
		text := p.PolicyTitle + " " + p.SubsidyCalcMethod + " " + p.PolicyType + " " + strings.Join(p.TargetGroupTags, " ")
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				s += strings.Count(text, kw)
			}
		}
		scoredList = append(scoredList, scored{policy: p, score: s})
	}
	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].score > scoredList[j].score
	})
	result := make([]models.PolicyClaim, len(scoredList))
	for i, sl := range scoredList {
		result[i] = sl.policy
	}
	return result
}

func callLLMGateway(ctx context.Context, gatewayURL, systemPrompt, userContent string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"system_prompt": systemPrompt,
		"user_content":  userContent,
		"max_tokens":    1024,
		"caller":        "advisor",
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", gatewayURL+"/v1/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm call: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Code    int    `json:"code"`
		Content string `json:"content"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Content == "" {
		return "", fmt.Errorf("empty response")
	}
	return result.Content, nil
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func truncateSentences(text string, maxSentences int) string {
	var enders = []string{"。", ".", "！", "！", "？", "?"}
	pos := 0
	count := 0
	for i := 0; i < len(text) && count < maxSentences; i++ {
		for _, e := range enders {
			if i+len(e) <= len(text) && text[i:i+len(e)] == e {
				pos = i + len(e)
				count++
				break
			}
		}
	}
	if pos == 0 || count <= maxSentences {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(text[:pos])
}

func generatePolicyAnswer(question string, simCtx map[string]interface{}, policyRepo PolicyQuerier, ctx context.Context) string {
	cityCode, _ := simCtx["city_code"].(string)
	if cityCode == "" {
		cityCode = "310000"
	}
	gender, _ := simCtx["gender"].(string)
	age := 0
	if v, ok := simCtx["age"].(float64); ok {
		age = int(v)
	}

	policies, err := policyRepo.QueryByRegionAndStatus(ctx, cityCode, "verified")
	if err != nil || len(policies) == 0 {
		return "暂未查询到您所在城市的社保政策信息。建议拨打12333社保热线咨询当地最新政策。"
	}

	keywords := extractKeywords(question)
	scored := scorePoliciesByKeyword(policies, keywords)
	topN := scored
	if len(topN) > 3 {
		topN = topN[:3]
	}

	var answer strings.Builder
	answer.WriteString("根据您所在城市的社保政策：\n")
	for i, p := range topN {
		if i >= 3 {
			break
		}
		line := fmt.Sprintf("%d. %s", i+1, p.PolicyTitle)
		if p.SubsidyCalcMethod != "" {
			line += "（" + p.SubsidyCalcMethod + "）"
		}
		if p.IssuingAuthority != "" {
			line += "，来源：" + p.IssuingAuthority
		}
		answer.WriteString(line + "。\n")
	}

	if age >= 40 && gender == "female" {
		answer.WriteString("您已符合4050社保补贴条件（女性40岁以上），建议咨询当地人社局申请。")
	} else if age >= 50 && gender == "male" {
		answer.WriteString("您已符合4050社保补贴条件（男性50岁以上），建议咨询当地人社局申请。")
	} else if age > 0 {
		answer.WriteString(fmt.Sprintf("您今年%d岁，建议关注当地社保补贴政策，符合条件时及时申请。", age))
	}

	return strings.TrimSpace(answer.String())
}
