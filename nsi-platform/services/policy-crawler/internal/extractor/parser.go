package extractor

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func parseExtractionResultRobust(llmOutput string) (*ExtractionResult, string, error) {
	result, err := tryStandardParse(llmOutput)
	if err == nil {
		return result, "full", nil
	}

	result, err = tryRepairParse(llmOutput)
	if err == nil {
		return result, "full", nil
	}

	result, err = tryRegexFallback(llmOutput)
	if err == nil {
		return result, "regex_fallback", nil
	}

	return nil, "", fmt.Errorf("all parsing methods failed: %w", err)
}

func tryStandardParse(input string) (*ExtractionResult, error) {
	start := strings.Index(input, "{")
	end := strings.LastIndex(input, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found")
	}
	var result ExtractionResult
	if err := json.Unmarshal([]byte(input[start:end+1]), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func tryRepairParse(input string) (*ExtractionResult, error) {
	cleaned := input

	cbMatch := reCodeBlock.FindStringSubmatch(cleaned)
	if len(cbMatch) > 1 {
		cleaned = cbMatch[1]
	}

	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object in repaired input")
	}
	jsonStr := cleaned[start : end+1]

	jsonStr = reTrailingCommaObj.ReplaceAllString(jsonStr, "}")
	jsonStr = reTrailingCommaArr.ReplaceAllString(jsonStr, "]")

	var result ExtractionResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

var (
	reCodeBlock       = regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.*?)\\n?```")
	reTrailingCommaObj = regexp.MustCompile(`,\s*}`)
	reTrailingCommaArr = regexp.MustCompile(`,\s*]`)
	rePolicyID    = regexp.MustCompile(`policy[_ ]?id[：:"是为]?\s*["']?([A-Za-z0-9\-_]+)`)
	reRegionCode  = regexp.MustCompile(`(?:地区代码|region[_ ]?code)[：:"]?\s*["']?(\d{6})`)
	rePolicyType  = regexp.MustCompile(`(?:政策类型|policy[_ ]?type)[：:"]?\s*["']?(pension|medical|unemployment|injury|maternity|housing_fund|subsidy|training)`)
	reAmountMin   = regexp.MustCompile(`(?:最低补贴|amount[_ ]?min)[：:"]?\s*["']?([\d.]+)`)
	reAmountMax   = regexp.MustCompile(`(?:最高补贴|amount[_ ]?max)[：:"]?\s*["']?([\d.]+)`)
	reEffectiveDt = regexp.MustCompile(`(?:生效日期|effective[_ ]?date)[：:"]?\s*["']?(\d{4}[-/]\d{2}[-/]\d{2})`)
	reBriefSumm   = regexp.MustCompile(`(?:brief[_ ]?summary|概括|要点)[：:"]?\s*["']([^"]{1,100})`)
)

func tryRegexFallback(input string) (*ExtractionResult, error) {
	result := &ExtractionResult{
		TargetGroups:      []string{},
		SubsidyCalcMethod: "参见政策原文",
	}

	if m := rePolicyID.FindStringSubmatch(input); len(m) > 1 {
		result.PolicyID = m[1]
	}
	if m := reRegionCode.FindStringSubmatch(input); len(m) > 1 {
		result.RegionCode = m[1]
	}
	if m := rePolicyType.FindStringSubmatch(input); len(m) > 1 {
		result.PolicyType = m[1]
	}
	if m := reAmountMin.FindStringSubmatch(input); len(m) > 1 {
		var v float64
		if _, err := fmt.Sscanf(m[1], "%f", &v); err == nil {
			result.AmountMin = &v
		}
	}
	if m := reAmountMax.FindStringSubmatch(input); len(m) > 1 {
		var v float64
		if _, err := fmt.Sscanf(m[1], "%f", &v); err == nil {
			result.AmountMax = &v
		}
	}
	if m := reEffectiveDt.FindStringSubmatch(input); len(m) > 1 {
		result.EffectiveDate = strings.ReplaceAll(m[1], "/", "-")
	}
	if m := reBriefSumm.FindStringSubmatch(input); len(m) > 1 {
		result.BriefSummary = m[1]
	}

	if result.PolicyID == "" && result.RegionCode == "" && result.PolicyType == "" {
		return nil, fmt.Errorf("regex fallback found no extractable fields")
	}
	return result, nil
}
