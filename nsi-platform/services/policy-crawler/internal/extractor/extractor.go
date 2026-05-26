package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/llm"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

// RawTextStore 原始文本存储接口
type RawTextStore interface {
	GetUnprocessedRawTexts(limit int) ([]RawTextEntry, error)
	MarkExtracted(id int64, claimID string) error
	InsertClaim(claim *models.PolicyClaim) error
	SaveExtractLog(sourceID string, success bool, msg string)
	SaveExtractLogDetailed(rawTextID int64, sourceID string, success bool, msg string, claimID string, title string, modelName string, summary string)
	SaveEmbedding(claimID string, embedding []float64) error
	SaveSnapshot(claim *models.PolicyClaim) error
	MarkSuperseded(oldClaimID, newClaimID string) error
	ListVersions(policyID string) ([]models.VersionSnapshot, error)
}

type RawTextEntry struct {
	ID         int64  `json:"id"`
	SourceID   string `json:"source_id"`
	Content    string `json:"content"`
	SourceURL  string `json:"source_url"`
	SourceName string `json:"source_name"`
	Title      string `json:"title"`
}

// ReferenceChecker 交叉验证接口
type ReferenceChecker interface {
	SearchSimilar(emb []float64, threshold float64, limit int, filter *embeddings.SearchFilter) []embeddings.SimilarResult
}

// Extractor LLM 政策提取器
type Extractor struct {
	store   RawTextStore
	client  *llm.Client
	checker ReferenceChecker
	embProv embeddings.EmbeddingProvider
}

func NewExtractor(store RawTextStore, client *llm.Client) *Extractor {
	return &Extractor{store: store, client: client}
}

func (e *Extractor) SetReferenceChecker(c ReferenceChecker) {
	e.checker = c
}

func (e *Extractor) SetEmbeddingProvider(p embeddings.EmbeddingProvider) {
	e.embProv = p
}

func (e *Extractor) embedText(text string) []float64 {
	if e.embProv != nil {
		vecs, err := e.embProv.Embed(context.Background(), []string{text})
		if err != nil {
			log.Printf("[extractor] embedding provider failed, using hash fallback: %v", err)
			return e.hashEmbed(text)
		}
		if len(vecs) > 0 {
			return vecs[0]
		}
	}
	return e.hashEmbed(text)
}

func (e *Extractor) hashEmbed(text string) []float64 {
	raw := embeddings.FromText(text)
	padded := make([]float64, 1536)
	copy(padded, raw)
	embeddings.Normalize(padded)
	return padded
}

func buildEmbedText(policyType, subsidyCalcMethod, policyID, regionCode string, conditions []map[string]interface{}) string {
	text := policyType + " " + subsidyCalcMethod + " " + policyID + " " + regionCode
	for _, c := range conditions {
		if name, ok := c["name"].(string); ok {
			text += " " + name
		}
		if desc, ok := c["description"].(string); ok {
			text += " " + desc
		}
	}
	return text
}

// ProcessUnprocessed 批量处理所有未提取的原始文本
func (e *Extractor) ProcessUnprocessed(limit int) (int, int, error) {
	entries, err := e.store.GetUnprocessedRawTexts(limit)
	if err != nil {
		return 0, 0, fmt.Errorf("query unprocessed: %w", err)
	}

	var succeeded, failed int
	for _, entry := range entries {
		if err := e.ProcessOne(entry); err != nil {
			log.Printf("[extractor] failed id=%d source=%s: %v", entry.ID, entry.SourceID, err)
			e.store.SaveExtractLogDetailed(entry.ID, entry.SourceID, false, err.Error(), "", entry.Title, e.client.ModelName(), "")
			failed++
		} else {
			succeeded++
		}
	}
	return succeeded, failed, nil
}

func (e *Extractor) ProcessOne(entry RawTextEntry) error {
	// 1. 提取正文（去除 HTML 标签）
	cleanText := extractPlainText(entry.Content)
	if len(cleanText) < 50 {
		return fmt.Errorf("content too short (%d bytes) after cleaning", len(cleanText))
	}

	// 2. 构建 LLM 提示词
	systemPrompt := `你是一个专业的中国社保政策分析专家。你的任务是从政府政策文本中提取结构化信息。
请提取以下字段，只返回JSON，不要其他文字：
{
  "policy_id": "唯一政策ID",
  "region_code": "地区行政代码(6位)",
  "policy_type": "政策类型(pension/medical/unemployment/injury/maternity/housing_fund/subsidy/training)",
  "target_groups": ["适用人群标签(flexible_employment/unemployed/employed/4050/has_children/female/male/low_income)"],
  "subsidy_calc_method": "补贴计算方法描述",
  "amount_min": 最低补贴金额(数字),
  "amount_max": 最高补贴金额(数字,可选),
  "subsidy_duration": 补贴期限(月,可选),
  "effective_date": "生效日期YYYY-MM-DD",
  "expire_date": "失效日期YYYY-MM-DD(可选)",
  "policy_url": "该政策原文的网址(从页面文本中提取完整的URL,必填)",
  "brief_summary": "用一句话概括该社保政策的要点(不超过50字)",
  "conditions": [{"name":"条件名称","description":"条件描述","tag_match":"对应人群标签"}],
  "required_documents": [{"name":"材料名称","description":"描述","source":"user/gov","optional":false}]
}`

	// 3. 调用 LLM
	llmResp, err := e.client.Chat(systemPrompt, cleanText)
	if err != nil {
		return fmt.Errorf("LLM call: %w", err)
	}

	// 4. 解析 LLM 响应
	parsed, err := parseExtractionResult(llmResp)
	if err != nil {
		return fmt.Errorf("parse LLM result: %w", err)
	}

	// 4a. 补充缺失字段
	if parsed.PolicyID == "" {
		parsed.PolicyID = fmt.Sprintf("AUTO-%s-%d", entry.SourceID, time.Now().UnixMilli())
	}
	if parsed.RegionCode == "" {
		parsed.RegionCode = "000000"
	}

	// 4b. 校验 policy_type（映射到允许的值）
	parsed.PolicyType = validatePolicyType(parsed.PolicyType)

	// 4c. 确保 target_group_tags 非 nil（DB 约束 NOT NULL）
	if parsed.TargetGroups == nil {
		parsed.TargetGroups = []string{}
	}

	// 4d. 确保 subsidy_calc_method 非空（DB 约束 NOT NULL）
	if parsed.SubsidyCalcMethod == "" {
		parsed.SubsidyCalcMethod = "参见政策原文"
	}

	// 4e. 清洗日期字段（LLM 可能返回 N/A/null/空字符串）
	if parsed.EffectiveDate != "" && parsed.EffectiveDate != "N/A" && parsed.EffectiveDate != "null" && parsed.EffectiveDate != "NULL" {
		if _, err := time.Parse("2006-01-02", parsed.EffectiveDate); err != nil {
			parsed.EffectiveDate = time.Now().Format("2006-01-02")
		}
	} else {
		parsed.EffectiveDate = time.Now().Format("2006-01-02")
	}
	if parsed.ExpireDate != nil {
		d := strings.TrimSpace(*parsed.ExpireDate)
		if d == "" || d == "N/A" || d == "null" || d == "NULL" {
			parsed.ExpireDate = nil
		} else if _, err := time.Parse("2006-01-02", d); err != nil {
			parsed.ExpireDate = nil
		}
	}

	// 5. 构建 PolicyClaim
	condJSON, _ := json.Marshal(parsed.Conditions)
	docJSON, _ := json.Marshal(parsed.RequiredDocuments)

	claim := &models.PolicyClaim{
		ClaimID:            fmt.Sprintf("LLM-%d", time.Now().UnixNano()),
		PolicyID:           parsed.PolicyID,
		RegionCode:         parsed.RegionCode,
		PolicyType:         parsed.PolicyType,
		TargetGroupTags:    parsed.TargetGroups,
		SubsidyCalcMethod:  parsed.SubsidyCalcMethod,
		SubsidyAmountMin:   parsed.AmountMin,
		SubsidyAmountMax:   parsed.AmountMax,
		SubsidyDuration:    parsed.SubsidyDuration,
		EffectiveDate:      parsed.EffectiveDate,
		ExpireDate:         parsed.ExpireDate,
		ConfidenceScore:    0.85,
		Status:             "verified",
		VersionNumber:      1,
		Conditions:         condJSON,
		RequiredDocuments:  docJSON,
		SourceID:           entry.SourceID,
		SourceName:         entry.SourceName,
		SourceURL:          entry.SourceURL,
		PolicyURL:          parsed.PolicyURL,
	}

	// 5a. 交叉验证（如果配置了 checker）
	if e.checker != nil {
		embedText := buildEmbedText(parsed.PolicyType, parsed.SubsidyCalcMethod, parsed.PolicyID, parsed.RegionCode, parsed.Conditions)
		emb := e.embedText(embedText)
		similar := e.checker.SearchSimilar(emb, 0.5, 10, &embeddings.SearchFilter{RegionCode: parsed.RegionCode})
		maxScore := 0.0
		for i := range similar {
			if similar[i].Score > maxScore {
				maxScore = similar[i].Score
			}
		}
		if maxScore > 0.85 && parsed.RegionCode != "" {
			e.store.SaveExtractLogDetailed(entry.ID, entry.SourceID, true,
				fmt.Sprintf("duplicate of %s (score=%.2f), skipped insert", similar[0].ClaimID, maxScore),
				similar[0].ClaimID, entry.Title, e.client.ModelName(), "")
			if err := e.store.MarkExtracted(entry.ID, similar[0].ClaimID); err != nil {
				return fmt.Errorf("mark extracted: %w", err)
			}
			log.Printf("[extractor] skipped claim (duplicate of %s, score=%.2f) from raw_text id=%d", similar[0].ClaimID, maxScore, entry.ID)
			return nil
		} else if maxScore > 0.7 && parsed.RegionCode != "" {
			claim.Status = "pending_review"
			claim.ConfidenceScore *= 0.9
		}
	}

	// 6. 入库
	if err := e.store.InsertClaim(claim); err != nil {
		return fmt.Errorf("insert claim: %w", err)
	}

	// 6a. 保存版本快照
	if err := e.store.SaveSnapshot(claim); err != nil {
		log.Printf("[extractor] failed to save snapshot for %s: %v", claim.ClaimID, err)
	}

	// 7. 生成向量并存储
	embedText := buildEmbedText(claim.PolicyType, claim.SubsidyCalcMethod, claim.PolicyID, claim.RegionCode, parsed.Conditions)
	emb := e.embedText(embedText)
	if err := e.store.SaveEmbedding(claim.ClaimID, emb); err != nil {
		log.Printf("[extractor] failed to save embedding for %s: %v", claim.ClaimID, err)
	}

	// 8. 标记已提取
	if err := e.store.MarkExtracted(entry.ID, claim.ClaimID); err != nil {
		return fmt.Errorf("mark extracted: %w", err)
	}

	e.store.SaveExtractLogDetailed(entry.ID, entry.SourceID, true, fmt.Sprintf("claim=%s", claim.ClaimID), claim.ClaimID, entry.Title, e.client.ModelName(), parsed.BriefSummary)
	log.Printf("[extractor] extracted claim %s from raw_text id=%d", claim.ClaimID, entry.ID)
	return nil
}

// ExtractionResult LLM 提取的中间结果
type ExtractionResult struct {
	PolicyID          string                   `json:"policy_id"`
	RegionCode        string                   `json:"region_code"`
	PolicyType        string                   `json:"policy_type"`
	TargetGroups      []string                 `json:"target_groups"`
	SubsidyCalcMethod string                   `json:"subsidy_calc_method"`
	AmountMin         *float64                 `json:"amount_min"`
	AmountMax         *float64                 `json:"amount_max"`
	SubsidyDuration   *int                     `json:"subsidy_duration"`
	EffectiveDate     string                   `json:"effective_date"`
	ExpireDate        *string                  `json:"expire_date"`
	Conditions        []map[string]interface{} `json:"conditions"`
	RequiredDocuments []map[string]interface{} `json:"required_documents"`
	PolicyURL         string                   `json:"policy_url"`
	BriefSummary      string                   `json:"brief_summary"`
}

func parseExtractionResult(llmOutput string) (*ExtractionResult, error) {
	// 查找 JSON 块
	start := strings.Index(llmOutput, "{")
	end := strings.LastIndex(llmOutput, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON found in LLM output")
	}

	var result ExtractionResult
	if err := json.Unmarshal([]byte(llmOutput[start:end+1]), &result); err != nil {
		return nil, fmt.Errorf("JSON parse: %w", err)
	}
	return &result, nil
}

// validatePolicyType 校验/映射政策类型到允许的值
func validatePolicyType(t string) string {
	valid := map[string]string{
		"pension":        "pension",
		"medical":        "medical",
		"unemployment":   "unemployment",
		"injury":         "injury",
		"maternity":      "maternity",
		"housing_fund":   "housing_fund",
		"subsidy":        "subsidy",
		"training":       "training",
		// 常见 LLM 映射
		"social_insurance":   "pension",
		"social_subsidy":    "subsidy",
		"employment_subsidy": "subsidy",
		"housing_subsidy":   "subsidy",
		"medical_subsidy":   "medical",
		"endowment_insurance": "pension",
		"pension_insurance": "pension",
	}
	if v, ok := valid[t]; ok {
		return v
	}
	return "subsidy"
}

// extractPlainText 从 HTML 提取纯文本
func extractPlainText(html string) string {
	if len(html) == 0 {
		return ""
	}

	// 安全的正则替换
	safeReplace := func(pattern, src, repl string) string {
		defer func() { recover() }()
		r := regexp.MustCompile(pattern)
		return r.ReplaceAllString(src, repl)
	}

	// 去除 <style> 和 <script> 块 (Go RE2 不支持反向引用 \1，分开匹配)
	html = safeReplace(`(?is)<style[^>]*>.*?</style>`, html, "")
	html = safeReplace(`(?is)<script[^>]*>.*?</script>`, html, "")

	// 去除 HTML 标签
	html = safeReplace(`<[^>]*>`, html, "")

	// 解码 HTML 实体
	html = strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", "\"").Replace(html)

	// 合并空白
	html = safeReplace(`\s+`, html, " ")

	// 保留中文句号后的分段
	html = strings.ReplaceAll(html, "。", "。\n")
	html = strings.ReplaceAll(html, "；", "；\n")

	// 过滤短行
	lines := strings.Split(html, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 10 {
			cleaned = append(cleaned, line)
		}
	}

	result := strings.Join(cleaned, "\n")
	if len(result) > 8000 {
		result = result[:8000]
	}
	return result
}
