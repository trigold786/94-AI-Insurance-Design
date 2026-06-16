package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LLMClaimResponse struct {
	PolicyID           string   `json:"policy_id"`
	RegionCode         string   `json:"region_code"`
	PolicyType         string   `json:"policy_type"`
	TargetGroups       []string `json:"target_groups"`
	SubsidyCalcMethod  string   `json:"subsidy_calc_method"`
	AmountMin          *float64 `json:"amount_min"`
	AmountMax          *float64 `json:"amount_max,omitempty"`
	SubsidyDuration    *int     `json:"subsidy_duration,omitempty"`
	EffectiveDate      string   `json:"effective_date"`
	ExpireDate         *string  `json:"expire_date,omitempty"`
}

type PolicyClaim struct {
	PolicyID          string
	RegionCode        string
	PolicyType        string
	TargetGroups      []string
	SubsidyCalcMethod string
	AmountMin         *float64
	AmountMax         *float64
	SubsidyDuration   *int
	EffectiveDate     string
	ExpireDate        *string
}

func ParseLLMResponse(response string) (*PolicyClaim, error) {
	if strings.TrimSpace(response) == "" {
		return nil, fmt.Errorf("empty LLM response")
	}

	var parsed LLMClaimResponse
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response JSON: %w", err)
	}

	if parsed.PolicyID == "" {
		return nil, fmt.Errorf("missing required field: policy_id")
	}
	if parsed.RegionCode == "" {
		return nil, fmt.Errorf("missing required field: region_code")
	}
	if parsed.PolicyType == "" {
		return nil, fmt.Errorf("missing required field: policy_type")
	}
	if parsed.EffectiveDate == "" {
		return nil, fmt.Errorf("missing required field: effective_date")
	}
	if len(parsed.TargetGroups) == 0 {
		return nil, fmt.Errorf("missing required field: target_groups")
	}
	if parsed.SubsidyCalcMethod == "" {
		return nil, fmt.Errorf("missing required field: subsidy_calc_method")
	}

	return &PolicyClaim{
		PolicyID:          parsed.PolicyID,
		RegionCode:        parsed.RegionCode,
		PolicyType:        parsed.PolicyType,
		TargetGroups:      parsed.TargetGroups,
		SubsidyCalcMethod: parsed.SubsidyCalcMethod,
		AmountMin:         parsed.AmountMin,
		AmountMax:         parsed.AmountMax,
		SubsidyDuration:   parsed.SubsidyDuration,
		EffectiveDate:     parsed.EffectiveDate,
		ExpireDate:        parsed.ExpireDate,
	}, nil
}

func ParsePolicyText(text, apiKey, endpoint string) (*PolicyClaim, error) {
	prompt := fmt.Sprintf(`你是一个社保政策分析专家。请从以下政策文本中提取结构化信息，返回JSON格式。
政策文本：
%s

请返回以下JSON格式（只返回JSON，不要其他文字）：
{
  "policy_id": "唯一政策ID",
  "region_code": "地区行政代码",
  "policy_type": "政策类型(pension/medical/unemployment/injury/maternity/housing_fund/subsidy/training)",
  "target_groups": ["适用人群标签"],
  "subsidy_calc_method": "补贴计算方法",
  "amount_min": 最低金额(数字),
  "amount_max": 最高金额(数字,可选),
  "subsidy_duration": 补贴时长(月,可选),
  "effective_date": "生效日期 YYYY-MM-DD",
  "expire_date": "失效日期 YYYY-MM-DD(可选)"
}`, text)

	var reqBody []byte
	var isCompat bool
	if strings.Contains(endpoint, "compatible-mode") {
		isCompat = true
		reqBody, _ = json.Marshal(map[string]interface{}{
			"model": "qwen3.6-plus",
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
		})
	} else {
		reqBody, _ = json.Marshal(map[string]interface{}{
			"model": "qwen-plus",
			"input": map[string]interface{}{
				"messages": []map[string]string{
					{"role": "user", "content": prompt},
				},
			},
		})
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read LLM response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
	}

	if isCompat {
		var apiResp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to parse LLM API response: %w", err)
		}
		if len(apiResp.Choices) == 0 {
			return nil, fmt.Errorf("empty response choices")
		}
		return ParseLLMResponse(apiResp.Choices[0].Message.Content)
	}

	var apiResp struct {
		Output struct {
			Text string `json:"text"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM API response: %w", err)
	}

	return ParseLLMResponse(apiResp.Output.Text)
}
