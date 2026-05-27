package extractor

import (
	"testing"
)

func TestParseExtractionResult_StandardJSON(t *testing.T) {
	input := `{"policy_id":"P001","region_code":"310000","policy_type":"pension","target_groups":[],"subsidy_calc_method":"按月","amount_min":500,"brief_summary":"测试政策"}`
	result, method, err := parseExtractionResultRobust(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "full" {
		t.Fatalf("expected method=full, got %s", method)
	}
	if result.PolicyID != "P001" {
		t.Fatalf("expected P001, got %s", result.PolicyID)
	}
}

func TestParseExtractionResult_MarkdownWrapped(t *testing.T) {
	input := "```json\n{\"policy_id\":\"P002\",\"region_code\":\"110000\",\"policy_type\":\"medical\",\"target_groups\":[],\"subsidy_calc_method\":\"按年\",\"brief_summary\":\"测试\"}\n```"
	result, method, err := parseExtractionResultRobust(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "full" {
		t.Fatalf("expected method=full, got %s", method)
	}
	if result.PolicyID != "P002" {
		t.Fatalf("expected P002, got %s", result.PolicyID)
	}
}

func TestParseExtractionResult_TrailingText(t *testing.T) {
	input := `{"policy_id":"P003","region_code":"","policy_type":"subsidy","target_groups":[],"subsidy_calc_method":"","brief_summary":""} 这是一段多余的说明文字，不是JSON的一部分。`
	result, method, err := parseExtractionResultRobust(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "full" {
		t.Fatalf("expected method=full, got %s", method)
	}
	if result.PolicyID != "P003" {
		t.Fatalf("expected P003, got %s", result.PolicyID)
	}
}

func TestParseExtractionResult_TrailingComma(t *testing.T) {
	input := `{"policy_id":"P004","region_code":"","policy_type":"subsidy","target_groups":[],"subsidy_calc_method":"","brief_summary":"",}`
	result, _, err := parseExtractionResultRobust(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PolicyID != "P004" {
		t.Fatalf("expected P004, got %s", result.PolicyID)
	}
}

func TestParseExtractionResult_RegexFallback(t *testing.T) {
	input := `这段文字没有JSON格式。policy_id是P005，地区代码为440300，政策类型为training。补贴金额最低200元。生效日期2024-01-01。`
	result, method, err := parseExtractionResultRobust(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "regex_fallback" {
		t.Fatalf("expected method=regex_fallback, got %s", method)
	}
	if result.PolicyID != "P005" {
		t.Fatalf("expected P005, got %s", result.PolicyID)
	}
}

func TestParseExtractionResult_CompleteGarbage(t *testing.T) {
	input := `这是一段完全没有任何可提取信息的文字内容。`
	_, _, err := parseExtractionResultRobust(input)
	if err == nil {
		t.Fatal("expected error for garbage input")
	}
}
