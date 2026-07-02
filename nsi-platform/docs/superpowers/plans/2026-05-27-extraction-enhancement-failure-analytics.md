# LLM Extraction Enhancement & Failure Analytics Dashboard �?Implementation Plan

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Raise extraction success rate from 22% to >70%, enrich extracted policy fields, and add a failure analytics dashboard with charts and retry.

**Architecture:** Extend the existing `extractor` package with smart document splitting, fault-tolerant JSON parsing, and new extraction fields. Add failure analytics as new store methods + admin API handlers + new admin UI tab. Single migration (025) for all DB changes.

**Tech Stack:** Go 1.22, PostgreSQL with pgvector, Chart.js v4 via CDN, existing admin HTML template pattern.

---

## File Structure

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `migrations/025_extraction_enhancement.sql` | New columns on policy_claims |
| Create | `internal/extractor/splitter.go` | Document splitting logic |
| Create | `internal/extractor/parser.go` | Fault-tolerant JSON parsing + regex fallback |
| Create | `internal/extractor/splitter_test.go` | Tests for splitter |
| Create | `internal/extractor/parser_test.go` | Tests for parser |
| Modify | `internal/extractor/extractor.go` | Integration of splitter+parser+new fields into ProcessOne |
| Modify | `shared/models/models.go` | New fields on PolicyClaim struct |
| Modify | `internal/crawler/store.go` | Updated INSERT/SELECT for new columns + failure analytics queries + retry methods |
| Create | `internal/crawler/failure_queries.go` | Failure analytics SQL queries |
| Create | `internal/admin/admin_failures.go` | Failure analytics HTTP handlers |
| Modify | `internal/admin/admin_dashboard.go` | DashboardStore interface additions |
| Modify | `internal/admin/admin_page.go` | New "失败分析" tab with charts + retry UI |
| Modify | `cmd/main.go` | Register new `/admin/failures/*` routes |

---

### Task 1: Migration 025 �?New Columns on policy_claims

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/migrations/025_extraction_enhancement.sql`

- [ ] **Step 1: Write the migration SQL**

```sql
-- 025_extraction_enhancement.sql
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS policy_title text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS issuing_authority text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS document_number text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS application_process jsonb DEFAULT '[]'::jsonb;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS contact_info text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS source_type text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS extraction_method text DEFAULT 'full';
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS raw_text_length int DEFAULT 0;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS split_count int DEFAULT 0;
```

- [ ] **Step 2: Verify migration runs**

Run: `docker compose restart policy-crawler` then check logs for migration success.

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/migrations/025_extraction_enhancement.sql
git commit -m "feat: migration 025 - extraction enhancement columns on policy_claims"
```

---

### Task 2: PolicyClaim Model �?New Fields

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Modify: `shared/models/models.go:30-57` (PolicyClaim struct)

- [ ] **Step 1: Add new fields to PolicyClaim struct**

Add these fields after `PolicyURL` (line 50):

```go
	PolicyTitle         string          `db:"policy_title" json:"policy_title,omitempty"`
	IssuingAuthority    string          `db:"issuing_authority" json:"issuing_authority,omitempty"`
	DocumentNumber      string          `db:"document_number" json:"document_number,omitempty"`
	ApplicationProcess  json.RawMessage `db:"application_process" json:"application_process,omitempty"`
	ContactInfo         string          `db:"contact_info" json:"contact_info,omitempty"`
	SourceType          string          `db:"source_type" json:"source_type,omitempty"`
	ExtractionMethod    string          `db:"extraction_method" json:"extraction_method,omitempty"`
	RawTextLength       int             `db:"raw_text_length" json:"raw_text_length,omitempty"`
	SplitCount          int             `db:"split_count" json:"split_count,omitempty"`
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./shared/...` from `services/policy-crawler`

- [ ] **Step 3: Commit**

```bash
git add shared/models/models.go
git commit -m "feat: add extraction enhancement fields to PolicyClaim model"
```

---

### Task 3: Smart Document Splitter

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/extractor/splitter.go`
- Create: `services/policy-crawler/internal/extractor/splitter_test.go`

- [ ] **Step 1: Write the test**

```go
package extractor

import (
	"strings"
	"testing"
)

func TestSplitDocument_Short(t *testing.T) {
	text := "短文档内容，不需要分片�?
	chunks := splitDocument(text, 4000)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != text {
		t.Fatalf("chunk content mismatch")
	}
}

func TestSplitDocument_Long(t *testing.T) {
	paras := make([]string, 20)
	for i := range paras {
		paras[i] = strings.Repeat("这是�?+string(rune('A'+i))+"段内容�?, 100)
	}
	text := strings.Join(paras, "\n\n")
	chunks := splitDocument(text, 4000)
	if len(chunks) < 2 {
		t.Fatalf("expected >= 2 chunks for long doc, got %d", len(chunks))
	}
	for i, c := range chunks {
		if len([]rune(c)) > 4000 {
			t.Fatalf("chunk %d too long: %d chars", i, len([]rune(c)))
		}
	}
	totalLen := 0
	for _, c := range chunks {
		totalLen += len(c)
	}
	if totalLen < len(text)-500 {
		t.Fatalf("too much content lost: original=%d chunks_total=%d", len(text), totalLen)
	}
}

func TestSplitDocument_MaxChunks(t *testing.T) {
	paras := make([]string, 30)
	for i := range paras {
		paras[i] = strings.Repeat("段落内容"+string(rune('A'+i)), 200)
	}
	text := strings.Join(paras, "\n\n")
	chunks := splitDocument(text, 2000)
	if len(chunks) > 5 {
		t.Fatalf("expected max 5 chunks, got %d", len(chunks))
	}
}

func TestSplitDocument_SingleHugeParagraph(t *testing.T) {
	text := strings.Repeat("超长段落不分段�?, 2000)
	chunks := splitDocument(text, 4000)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk for single para, got %d", len(chunks))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/policy-crawler && go test ./internal/extractor/ -run TestSplitDocument -v`
Expected: compilation error (splitDocument not defined)

- [ ] **Step 3: Write the implementation**

```go
package extractor

import (
	"strings"
)

func splitDocument(text string, maxChunkSize int) []string {
	runes := []rune(text)
	if len(runes) <= maxChunkSize {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var current strings.Builder

	for _, para := range paragraphs {
		paraRunes := []rune(para)
		if len(paraRunes) > maxChunkSize {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
			}
			chunks = append(chunks, para)
			continue
		}
		if current.Len() > 0 && current.Len()+len(paraRunes)+2 > maxChunkSize {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(para)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	if len(chunks) > 5 {
		var truncated []string
		totalLen := 0
		for _, c := range chunks {
			if totalLen+len(c) > maxChunkSize*5 {
				break
			}
			truncated = append(truncated, c)
			totalLen += len(c)
		}
		chunks = truncated
	}

	return chunks
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/policy-crawler && go test ./internal/extractor/ -run TestSplitDocument -v`
Expected: all 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add services/policy-crawler/internal/extractor/splitter.go services/policy-crawler/internal/extractor/splitter_test.go
git commit -m "feat: smart document splitter for long policy texts"
```

---

### Task 4: Fault-Tolerant Parser

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/extractor/parser.go`
- Create: `services/policy-crawler/internal/extractor/parser_test.go`

- [ ] **Step 1: Write the tests**

```go
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
	result, method, err := parseExtractionResultRobust(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PolicyID != "P004" {
		t.Fatalf("expected P004, got %s", result.PolicyID)
	}
}

func TestParseExtractionResult_RegexFallback(t *testing.T) {
	input := `这段文字没有JSON格式。policy_id是P005，地区代码为440300，政策类型为training。补贴金额最�?00元。生效日�?024-01-01。`
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/policy-crawler && go test ./internal/extractor/ -run TestParseExtractionResult_Robust -v`
Expected: compilation error

- [ ] **Step 3: Write the implementation**

```go
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

	cbMatch := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.*?)\\n?```").FindStringSubmatch(cleaned)
	if len(cbMatch) > 1 {
		cleaned = cbMatch[1]
	}

	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object in repaired input")
	}
	jsonStr := cleaned[start : end+1]

	jsonStr = regexp.MustCompile(`,\s*}`).ReplaceAllString(jsonStr, "}")
	jsonStr = regexp.MustCompile(`,\s*]`).ReplaceAllString(jsonStr, "]")

	var result ExtractionResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

var (
	rePolicyID    = regexp.MustCompile(`policy[_ ]?id[�?"]?\s*["']?([A-Za-z0-9\-_]+)`)
	reRegionCode  = regexp.MustCompile(`(?:地区代码|region[_ ]?code)[�?"]?\s*["']?(\d{6})`)
	rePolicyType  = regexp.MustCompile(`(?:政策类型|policy[_ ]?type)[�?"]?\s*["']?(pension|medical|unemployment|injury|maternity|housing_fund|subsidy|training)`)
	reAmountMin   = regexp.MustCompile(`(?:最低补贴|amount[_ ]?min)[�?"]?\s*["']?([\d.]+)`)
	reAmountMax   = regexp.MustCompile(`(?:最高补贴|amount[_ ]?max)[�?"]?\s*["']?([\d.]+)`)
	reEffectiveDt = regexp.MustCompile(`(?:生效日期|effective[_ ]?date)[�?"]?\s*["']?(\d{4}[-/]\d{2}[-/]\d{2})`)
	reBriefSumm   = regexp.MustCompile(`(?:brief[_ ]?summary|概括|要点)[�?"]?\s*["']([^"]{1,100})`)
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/policy-crawler && go test ./internal/extractor/ -run TestParseExtractionResult_Robust -v`
Expected: all 6 tests PASS

- [ ] **Step 5: Commit**

```bash
git add services/policy-crawler/internal/extractor/parser.go services/policy-crawler/internal/extractor/parser_test.go
git commit -m "feat: fault-tolerant parser with 3-level degradation (standard/repair/regex)"
```

---

### Task 5: Update Extractor �?Integrate Splitter + Parser + New Fields

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/extractor/extractor.go`

This is the core integration task. Changes to `ProcessOne()`:

1. Use `splitDocument()` for long texts
2. Use `parseExtractionResultRobust()` instead of `parseExtractionResult()`
3. Add merge step for split results
4. Extend `ExtractionResult` struct with 6 new fields
5. Populate new `PolicyClaim` fields from parsed result
6. Track `extractionMethod`, `rawTextLength`, `splitCount`
7. Update LLM system prompt with new fields

- [ ] **Step 1: Add new fields to ExtractionResult struct**

In `extractor.go`, after line 292 (`BriefSummary` field), add:

```go
	PolicyTitle        string                   `json:"policy_title"`
	IssuingAuthority   string                   `json:"issuing_authority"`
	DocumentNumber     string                   `json:"document_number"`
	ApplicationProcess []map[string]interface{} `json:"application_process"`
	ContactInfo        string                   `json:"contact_info"`
	SourceType         string                   `json:"source_type"`
```

- [ ] **Step 2: Update LLM system prompt**

Replace the system prompt (lines 128-145) with:

```go
	systemPrompt := `你是一个专业的中国社保政策分析专家。你的任务是从政府政策文本中提取结构化信息�?请提取以下字段，只返回JSON，不要其他文字：
{
  "policy_id": "唯一政策ID",
  "policy_title": "政策正式标题",
  "issuing_authority": "发文机关",
  "document_number": "文号(如沪人社规�?024�?�?",
  "region_code": "地区行政代码(6�?",
  "policy_type": "政策类型(pension/medical/unemployment/injury/maternity/housing_fund/subsidy/training)",
  "target_groups": ["适用人群标签(flexible_employment/unemployed/employed/4050/has_children/female/male/low_income)"],
  "subsidy_calc_method": "补贴计算方法描述",
  "amount_min": 最低补贴金�?数字),
  "amount_max": 最高补贴金�?数字,可�?,
  "subsidy_duration": 补贴期限(�?可�?,
  "effective_date": "生效日期YYYY-MM-DD",
  "expire_date": "失效日期YYYY-MM-DD(可�?",
  "policy_url": "该政策原文的网址(从页面文本中提取完整的URL,必填)",
  "brief_summary": "用一句话概括该社保政策的要点(不超�?0�?",
  "source_type": "原文类型(gov_doc/social_media/news/rumor)",
  "application_process": [{"step":1,"action":"办理步骤","description":"步骤描述"}],
  "contact_info": "咨询电话或办理地址",
  "conditions": [{"name":"条件名称","description":"条件描述","tag_match":"对应人群标签"}],
  "required_documents": [{"name":"材料名称","description":"描述","source":"user/gov","optional":false}]
}`
```

- [ ] **Step 3: Update ProcessOne to use splitter + robust parser**

Replace lines 122-157 (clean text �?single LLM call �?parse) with:

```go
	cleanText := extractPlainText(entry.Content)
	rawTextLen := len(cleanText)
	if rawTextLen < 50 {
		return fmt.Errorf("content too short (%d bytes) after cleaning", rawTextLen)
	}

	extractionMethod := "full"
	splitCount := 0

	var parsed *ExtractionResult

	chunks := splitDocument(cleanText, 4000)
	if len(chunks) == 1 {
		llmResp, err := e.client.Chat(systemPrompt, chunks[0])
		if err != nil {
			return fmt.Errorf("LLM call: %w", err)
		}
		var method string
		parsed, method, err = parseExtractionResultRobust(llmResp)
		if err != nil {
			return fmt.Errorf("parse LLM result: %w", err)
		}
		extractionMethod = method
	} else {
		splitCount = len(chunks)
		extractionMethod = "split"
		var partialResults []*ExtractionResult
		for i, chunk := range chunks {
			llmResp, err := e.client.Chat(systemPrompt, chunk)
			if err != nil {
				log.Printf("[extractor] LLM call failed for chunk %d/%d: %v", i+1, splitCount, err)
				continue
			}
			pr, method, err := parseExtractionResultRobust(llmResp)
			if err != nil {
				log.Printf("[extractor] parse failed for chunk %d/%d: %v", i+1, splitCount, err)
				continue
			}
			if method == "regex_fallback" {
				extractionMethod = "regex_fallback"
			}
			partialResults = append(partialResults, pr)
		}
		if len(partialResults) == 0 {
			return fmt.Errorf("all %d chunks failed extraction", splitCount)
		}
		if len(partialResults) == 1 {
			parsed = partialResults[0]
		} else {
			parsed = e.mergeResults(partialResults)
		}
	}
```

- [ ] **Step 4: Add mergeResults method**

Add new method to `Extractor`:

```go
func (e *Extractor) mergeResults(parts []*ExtractionResult) *ExtractionResult {
	merged := &ExtractionResult{
		TargetGroups:      []string{},
		SubsidyCalcMethod: "参见政策原文",
		Conditions:        []map[string]interface{}{},
		RequiredDocuments: []map[string]interface{}{},
	}
	for _, p := range parts {
		if p.PolicyID != "" && merged.PolicyID == "" {
			merged.PolicyID = p.PolicyID
		}
		if p.PolicyTitle != "" && merged.PolicyTitle == "" {
			merged.PolicyTitle = p.PolicyTitle
		}
		if p.RegionCode != "" && merged.RegionCode == "" {
			merged.RegionCode = p.RegionCode
		}
		if p.PolicyType != "" && merged.PolicyType == "" {
			merged.PolicyType = p.PolicyType
		}
		if p.IssuingAuthority != "" && merged.IssuingAuthority == "" {
			merged.IssuingAuthority = p.IssuingAuthority
		}
		if p.DocumentNumber != "" && merged.DocumentNumber == "" {
			merged.DocumentNumber = p.DocumentNumber
		}
		if p.SubsidyCalcMethod != "" && merged.SubsidyCalcMethod == "参见政策原文" {
			merged.SubsidyCalcMethod = p.SubsidyCalcMethod
		}
		if p.AmountMin != nil && merged.AmountMin == nil {
			merged.AmountMin = p.AmountMin
		}
		if p.AmountMax != nil && merged.AmountMax == nil {
			merged.AmountMax = p.AmountMax
		}
		if p.SubsidyDuration != nil && merged.SubsidyDuration == nil {
			merged.SubsidyDuration = p.SubsidyDuration
		}
		if p.EffectiveDate != "" && merged.EffectiveDate == "" {
			merged.EffectiveDate = p.EffectiveDate
		}
		if p.ExpireDate != nil && merged.ExpireDate == nil {
			merged.ExpireDate = p.ExpireDate
		}
		if p.PolicyURL != "" && merged.PolicyURL == "" {
			merged.PolicyURL = p.PolicyURL
		}
		if p.BriefSummary != "" && merged.BriefSummary == "" {
			merged.BriefSummary = p.BriefSummary
		}
		if p.SourceType != "" && merged.SourceType == "" {
			merged.SourceType = p.SourceType
		}
		if p.ContactInfo != "" && merged.ContactInfo == "" {
			merged.ContactInfo = p.ContactInfo
		}
		for _, tg := range p.TargetGroups {
			found := false
			for _, existing := range merged.TargetGroups {
				if existing == tg {
					found = true
					break
				}
			}
			if !found {
				merged.TargetGroups = append(merged.TargetGroups, tg)
			}
		}
		merged.Conditions = append(merged.Conditions, p.Conditions...)
		merged.RequiredDocuments = append(merged.RequiredDocuments, p.RequiredDocuments...)
		if p.ApplicationProcess != nil && merged.ApplicationProcess == nil {
			merged.ApplicationProcess = p.ApplicationProcess
		}
	}
	return merged
}
```

- [ ] **Step 5: Update PolicyClaim construction to include new fields**

In the claim construction block (around line 201), add after `PolicyURL`:

```go
		PolicyTitle:        parsed.PolicyTitle,
		IssuingAuthority:   parsed.IssuingAuthority,
		DocumentNumber:     parsed.DocumentNumber,
		ApplicationProcess: appProcJSON,
		ContactInfo:        parsed.ContactInfo,
		SourceType:         parsed.SourceType,
		ExtractionMethod:   extractionMethod,
		RawTextLength:      rawTextLen,
		SplitCount:         splitCount,
```

Where `appProcJSON` is computed earlier:

```go
	appProcJSON, _ := json.Marshal(parsed.ApplicationProcess)
```

- [ ] **Step 6: Verify compilation**

Run: `cd services/policy-crawler && go build ./...`

- [ ] **Step 7: Commit**

```bash
git add services/policy-crawler/internal/extractor/extractor.go
git commit -m "feat: integrate splitter + robust parser + new fields into extraction pipeline"
```

---

### Task 6: Update Store �?INSERT/SELECT for New Columns

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/crawler/store.go`

- [ ] **Step 1: Update InsertClaim (used by extractor) �?add 9 new columns**

The `InsertClaim` method at line 611 needs the new columns in its INSERT statement. Add the 9 new columns and parameters ($21-$29).

- [ ] **Step 2: Update the other InsertClaim (used by admin) at line 102**

Same pattern �?add 9 new columns.

- [ ] **Step 3: Update ListByStatus SELECT query at line 153**

Add the 9 new columns to the SELECT list and Scan call.

- [ ] **Step 4: Verify compilation**

Run: `cd services/policy-crawler && go build ./...`

- [ ] **Step 5: Commit**

```bash
git add services/policy-crawler/internal/crawler/store.go
git commit -m "feat: update store INSERT/SELECT for extraction enhancement columns"
```

---

### Task 7: Failure Analytics �?Data Layer

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/crawler/failure_queries.go`
- Modify: `services/policy-crawler/internal/admin/admin_dashboard.go` (DashboardStore interface)

- [ ] **Step 1: Create failure_queries.go with all query methods**

New types and 6 methods: `GetFailureSummary`, `GetFailureTrend`, `GetFailureBySource`, `GetTopFailureReasons`, `GetFailedRawTexts`, `RetryRawText`, `RetryAllFailed`.

Types to define in `admin_dashboard.go`:

```go
type FailureSummary struct {
	CrawlFailures    int `json:"crawl_failures"`
	ExtractFailures  int `json:"extract_failures"`
	VideoFailures    int `json:"video_failures"`
}

type FailureTrendPoint struct {
	Date            string `json:"date"`
	CrawlFailures   int    `json:"crawl_failures"`
	ExtractFailures int    `json:"extract_failures"`
	VideoFailures   int    `json:"video_failures"`
}

type FailureBySourceEntry struct {
	SourceID        string `json:"source_id"`
	SourceName      string `json:"source_name"`
	CrawlFailures   int    `json:"crawl_failures"`
	ExtractFailures int    `json:"extract_failures"`
	VideoFailures   int    `json:"video_failures"`
}

type TopFailureReason struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type FailedRawTextEntry struct {
	ID          int64  `json:"id"`
	SourceID    string `json:"source_id"`
	SourceName  string `json:"source_name"`
	Title       string `json:"title"`
	ErrorReason string `json:"error_reason"`
	FailedAt    string `json:"failed_at"`
	FailureType string `json:"failure_type"`
}
```

Add to `DashboardStore` interface:

```go
	GetFailureSummary() (*FailureSummary, error)
	GetFailureTrend(days int) ([]FailureTrendPoint, error)
	GetFailureBySource() ([]FailureBySourceEntry, error)
	GetTopFailureReasons(limit int) ([]TopFailureReason, error)
	GetFailedRawTexts(sourceID string, failureType string, limit int) ([]FailedRawTextEntry, error)
	RetryRawText(id int64) error
	RetryAllFailed(sourceID string) (int, error)
```

- [ ] **Step 2: Implement failure_queries.go**

Each method runs SQL against `crawl_logs`, `extract_logs`, and `policy_raw_texts`:

- `GetFailureSummary`: 3 COUNT queries
- `GetFailureTrend(days)`: UNION ALL of 3 subqueries grouped by date
- `GetFailureBySource`: UNION ALL of 3 subqueries grouped by source_id, JOIN policy_sources for name
- `GetTopFailureReasons(limit)`: UNION ALL of error messages from crawl_logs + extract_logs
- `GetFailedRawTexts(sourceID, failureType, limit)`: query extract_logs with status='failed' or policy_raw_texts with video_extract_status='failed'
- `RetryRawText(id)`: `UPDATE policy_raw_texts SET extracted=false, video_extract_status=NULL WHERE id=$1`
- `RetryAllFailed(sourceID)`: reset video failures + re-queue failed extract raw texts

- [ ] **Step 3: Verify compilation**

Run: `cd services/policy-crawler && go build ./...`

- [ ] **Step 4: Commit**

```bash
git add services/policy-crawler/internal/crawler/failure_queries.go services/policy-crawler/internal/admin/admin_dashboard.go
git commit -m "feat: failure analytics data layer with 7 query/retry methods"
```

---

### Task 8: Failure Analytics �?Admin API Handlers

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/admin/admin_failures.go`
- Modify: `services/policy-crawler/cmd/main.go` (register routes)

- [ ] **Step 1: Create admin_failures.go with 6 handlers**

Handlers:
- `FailureSummaryHandler(store DashboardStore)` �?GET `/admin/failures/summary`
- `FailureTrendHandler(store DashboardStore)` �?GET `/admin/failures/trend?days=7`
- `FailureBySourceHandler(store DashboardStore)` �?GET `/admin/failures/by-source`
- `FailureTopReasonsHandler(store DashboardStore)` �?GET `/admin/failures/top-reasons?limit=10`
- `FailureRawTextsHandler(store DashboardStore)` �?GET `/admin/failures/failed-raw-texts?source_id=&type=&limit=50`
- `FailureRetryHandler(store DashboardStore)` �?POST `/admin/failures/retry` with body `{"raw_text_id":123}` or `{"source_id":"X","all":true}`

- [ ] **Step 2: Register routes in cmd/main.go**

Add after line 173 (`mux.Handle("/admin/failures", ...)`):

```go
	mux.Handle("/admin/failures/summary", adminAuth(admin.FailureSummaryHandler(store)))
	mux.Handle("/admin/failures/trend", adminAuth(admin.FailureTrendHandler(store)))
	mux.Handle("/admin/failures/by-source", adminAuth(admin.FailureBySourceHandler(store)))
	mux.Handle("/admin/failures/top-reasons", adminAuth(admin.FailureTopReasonsHandler(store)))
	mux.Handle("/admin/failures/failed-raw-texts", adminAuth(admin.FailureRawTextsHandler(store)))
	mux.Handle("/admin/failures/retry", adminAuth(admin.FailureRetryHandler(store)))
```

- [ ] **Step 3: Verify compilation**

Run: `cd services/policy-crawler && go build ./...`

- [ ] **Step 4: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_failures.go services/policy-crawler/cmd/main.go
git commit -m "feat: failure analytics admin API with 6 endpoints"
```

---

### Task 9: Failure Analytics �?Admin UI Tab

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/admin/admin_page.go`

- [ ] **Step 1: Add nav item**

In `navItems` array (around line 71), add before the import entry:

```js
  {id:'failures',label:'\u5931\u8d25\u5206\u6790'},
```

- [ ] **Step 2: Add switchPanel routing**

In `switchPanel()` (around line 105), add:

```js
  else if(id==='failures')loadFailures();
```

- [ ] **Step 3: Implement loadFailures() function**

Add a `loadFailures()` function that:
1. Fetches `/admin/failures/summary`, `/admin/failures/trend?days=7`, `/admin/failures/by-source`, `/admin/failures/top-reasons?limit=10`
2. Renders summary stat cards (crawl/extract/video failure counts)
3. Renders Chart.js line chart for trend (7-day default with 30-day toggle)
4. Renders Chart.js doughnut for by-source distribution
5. Renders Chart.js horizontal bar for top reasons
6. Renders failed raw_texts table with retry button per row
7. Has "Retry Selected" batch action

Chart.js v4 is already loaded via CDN (`<script src="https://cdn.jsdelivr.net/npm/chart.js">`). The admin page already has canvas/chart patterns from the dashboard tab.

- [ ] **Step 4: Verify the UI loads in browser**

Build, deploy, navigate to `http://<container>:39403/admin#failures`, verify tab renders.

- [ ] **Step 5: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_page.go
git commit -m "feat: failure analytics admin UI tab with charts and retry"
```

---

### Task 10: Build, Deploy, Verify

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/Dockerfile` (no changes needed, just rebuild)

- [ ] **Step 1: Cross-compile binary**

Run:
```bash
cd services/policy-crawler
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o policy-crawler ./cmd/main.go
Copy-Item policy-crawler bin/policy-crawler -Force
```

- [ ] **Step 2: Build Docker image**

Run: `docker build -t nsi-policy-crawler:latest .` from `services/policy-crawler`

- [ ] **Step 3: Deploy**

Run: `docker compose up -d policy-crawler` from project root

- [ ] **Step 4: Verify health**

Run: `docker logs nsi-policy-crawler 2>&1 | Select-Object -Last 20`

Expected: no errors, migration 025 applied, server listening.

- [ ] **Step 5: Verify admin UI**

Open `http://<container>:39403/admin#failures` �?verify failure analytics tab loads with charts.

- [ ] **Step 6: Verify extraction pipeline**

Check logs for next extraction cycle �?verify new fields are being populated, splitter/parser working.

- [ ] **Step 7: Final commit**

```bash
git add -A
git commit -m "feat: LLM extraction enhancement & failure analytics dashboard �?complete deployment"
```
