package parser

import (
	"os"
	"testing"
)

func TestParseLLMResponseValid(t *testing.T) {
	json := `{
		"policy_id": "SH-2025-001",
		"region_code": "310000",
		"policy_type": "subsidy",
		"target_groups": ["flexible_employment", "unemployed"],
		"subsidy_calc_method": "fixed_amount",
		"amount_min": 500,
		"amount_max": 1500,
		"effective_date": "2025-01-01",
		"expire_date": "2025-12-31"
	}`

	claim, err := ParseLLMResponse(json)
	if err != nil {
		t.Fatalf("ParseLLMResponse() returned error: %v", err)
	}
	if claim.PolicyID != "SH-2025-001" {
		t.Errorf("expected SH-2025-001, got %s", claim.PolicyID)
	}
	if claim.PolicyType != "subsidy" {
		t.Errorf("expected subsidy, got %s", claim.PolicyType)
	}
	if len(claim.TargetGroups) != 2 {
		t.Errorf("expected 2 target groups, got %d", len(claim.TargetGroups))
	}
	if *claim.AmountMin != 500 {
		t.Errorf("expected amount_min 500, got %.0f", *claim.AmountMin)
	}
}

func TestParseLLMResponseInvalidJSON(t *testing.T) {
	_, err := ParseLLMResponse(`{invalid json}`)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseLLMResponseEmpty(t *testing.T) {
	_, err := ParseLLMResponse("")
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
}

func TestParseLLMResponseMissingRequiredFields(t *testing.T) {
	json := `{"policy_id": "TEST-001"}` // missing policy_type, region_code
	_, err := ParseLLMResponse(json)
	if err == nil {
		t.Fatal("expected error for missing fields, got nil")
	}
}

func TestParseLLMResponseMinimal(t *testing.T) {
	json := `{
		"policy_id": "BJ-2025-001",
		"region_code": "110000",
		"policy_type": "pension",
		"target_groups": ["employed"],
		"subsidy_calc_method": "percentage",
		"amount_min": 0,
		"effective_date": "2025-01-01"
	}`

	claim, err := ParseLLMResponse(json)
	if err != nil {
		t.Fatalf("ParseLLMResponse() returned error: %v", err)
	}
	if claim.PolicyID != "BJ-2025-001" {
		t.Errorf("expected BJ-2025-001, got %s", claim.PolicyID)
	}
	if claim.AmountMax != nil {
		t.Error("expected AmountMax to be nil for optional field")
	}
	if claim.ExpireDate != nil {
		t.Error("expected ExpireDate to be nil for optional field")
	}
}

func TestParseLLMResponseWrongTypes(t *testing.T) {
	json := `{
		"policy_id": "TEST-001",
		"region_code": "110000",
		"policy_type": "pension",
		"target_groups": "not_an_array",
		"subsidy_calc_method": "fixed",
		"amount_min": "should_be_number",
		"effective_date": "2025-01-01"
	}`

	_, err := ParseLLMResponse(json)
	if err == nil {
		t.Fatal("expected error for wrong types, got nil")
	}
}

func TestParsePolicyTextIntegration(t *testing.T) {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		t.Skip("LLM_API_KEY not set, skipping integration test")
	}

	text := `上海市人力资源和社会保障局关于2025年灵活就业人员社会保险补贴的通知。
	补贴标准：每月500-1500元。适用对象：灵活就业人员、失业人员。
	有效期限：2025年1月1日至2025年12月31日。`

	claim, err := ParsePolicyText(text, apiKey, "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation")
	if err != nil {
		t.Fatalf("ParsePolicyText() returned error: %v", err)
	}
	if claim.PolicyID == "" {
		t.Error("expected non-empty policy_id")
	}
	if claim.RegionCode == "" {
		t.Error("expected non-empty region_code")
	}
}
