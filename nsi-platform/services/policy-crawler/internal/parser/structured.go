package parser

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
)

// ParseStructuredText parses a structured policy text format without LLM API.
// Expected format:
//  政策ID: SH-2025-FLEX-SUBSIDY
//  地区代码: 310000
//  政策类型: subsidy
//  适用人群: flexible_employment, 4050
//  计算方法: 基数*50%
//  补贴下限: 300
//  补贴上限: 800
//  补贴时长: 24
//  生效日期: 2025-01-01
//  失效日期: 2026-12-31
//  认定条件:
//    - name: 灵活就业登记
//      desc: 已办理灵活就业登记
//      tag: flexible_employment
//    - name: 年龄要求
//      desc: 女性40岁以上/男性50岁以上
//      tag: 4050
//  必需材料:
//    - name: 身份证
//      desc: 原件及复印件
//      source: user
func ParseStructuredText(text string) (*PolicyClaim, []string, []string, error) {
	lines := strings.Split(text, "\n")
	result := &PolicyClaim{}
	var conditions []string
	var documents []string
	inConditions := false
	inDocuments := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			inConditions = false
			inDocuments = false
			continue
		}

		if strings.HasPrefix(line, "认定条件:") {
			inConditions = true
			inDocuments = false
			continue
		}
		if strings.HasPrefix(line, "必需材料:") {
			inDocuments = true
			inConditions = false
			continue
		}

		if inConditions {
			if match := regexp.MustCompile(`name:\s*(.+)$`).FindStringSubmatch(line); match != nil {
				conditions = append(conditions, match[1])
			}
			continue
		}
		if inDocuments {
			if match := regexp.MustCompile(`name:\s*(.+)$`).FindStringSubmatch(line); match != nil {
				documents = append(documents, match[1])
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "政策ID:"):
			result.PolicyID = strings.TrimSpace(strings.TrimPrefix(line, "政策ID:"))
		case strings.HasPrefix(line, "地区代码:"):
			result.RegionCode = strings.TrimSpace(strings.TrimPrefix(line, "地区代码:"))
		case strings.HasPrefix(line, "政策类型:"):
			result.PolicyType = strings.TrimSpace(strings.TrimPrefix(line, "政策类型:"))
		case strings.HasPrefix(line, "适用人群:"):
			parts := strings.Split(strings.TrimPrefix(line, "适用人群:"), ",")
			for _, p := range parts {
				tag := strings.TrimSpace(p)
				if tag != "" {
					result.TargetGroups = append(result.TargetGroups, tag)
				}
			}
		case strings.HasPrefix(line, "计算方法:"):
			result.SubsidyCalcMethod = strings.TrimSpace(strings.TrimPrefix(line, "计算方法:"))
		case strings.HasPrefix(line, "补贴下限:"):
			if v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "补贴下限:")), 64); err == nil {
				result.AmountMin = &v
			}
		case strings.HasPrefix(line, "补贴上限:"):
			if v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "补贴上限:")), 64); err == nil {
				result.AmountMax = &v
			}
		case strings.HasPrefix(line, "补贴时长:"):
			if v, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "补贴时长:"))); err == nil {
				result.SubsidyDuration = &v
			}
		case strings.HasPrefix(line, "生效日期:"):
			result.EffectiveDate = strings.TrimSpace(strings.TrimPrefix(line, "生效日期:"))
		case strings.HasPrefix(line, "失效日期:"):
			v := strings.TrimSpace(strings.TrimPrefix(line, "失效日期:"))
			result.ExpireDate = &v
		}
	}

	if result.PolicyID == "" {
		return nil, nil, nil, fmt.Errorf("missing required field: 政策ID")
	}
	if result.RegionCode == "" {
		return nil, nil, nil, fmt.Errorf("missing required field: 地区代码")
	}
	if result.PolicyType == "" {
		return nil, nil, nil, fmt.Errorf("missing required field: 政策类型")
	}
	if result.EffectiveDate == "" {
		return nil, nil, nil, fmt.Errorf("missing required field: 生效日期")
	}

	return result, conditions, documents, nil
}
