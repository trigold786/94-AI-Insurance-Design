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
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/verifier"
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
	GetMaxVersionNumber(policyID string) (int, error)
	GetLatestClaimByPolicyID(policyID string) (string, error)
}

type RawTextEntry struct {
	ID          int64  `json:"id"`
	SourceID    string `json:"source_id"`
	Content     string `json:"content"`
	SourceURL   string `json:"source_url"`
	SourceName  string `json:"source_name"`
	Title       string `json:"title"`
	SourceLevel string `json:"source_level"`
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

func (e *Extractor) computeConfidence(sourceLevel string) float64 {
	if sourceLevel == "" {
		sourceLevel = "MEDIUM"
	}
	mc := &models.PolicyClaim{
		SourceLevel:   sourceLevel,
		MatchRate:     0.5,
		ConflictScore: 1.0,
		FetchedAt:     time.Now().Format("2006-01-02"),
	}
	score := verifier.CalculateConfidence(mc, verifier.DefaultConfidenceConfig())
	status := verifier.DecideStatusWithConfig(score, verifier.DefaultStatusThresholds())
	if status == "verified" {
		return score
	}
	return score * 0.9
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
			e.store.MarkExtracted(entry.ID, "")
			failed++
		} else {
			succeeded++
		}
	}
	return succeeded, failed, nil
}

func (e *Extractor) ProcessOne(entry RawTextEntry) error {
	cleanText := extractPlainText(entry.Content)
	rawTextLen := len(cleanText)
	if rawTextLen < 50 {
		return fmt.Errorf("content too short (%d bytes) after cleaning", rawTextLen)
	}

	systemPrompt := `你是一个专业的中国社保政策分析专家。你的任务是从政府政策文本中提取结构化信息。
请提取以下字段，只返回JSON，不要其他文字：
{
  "policy_id": "唯一政策ID",
  "policy_title": "政策正式标题",
  "issuing_authority": "发文机关",
  "document_number": "文号(如沪人社规〔2024〕1号)",
  "publish_date": "政策发布日期YYYY-MM-DD(可选)",
  "region_code": "地区行政代码(6位)",
  "policy_type": "政策类型(pension/medical/unemployment/injury/maternity/housing_fund/subsidy/training)",
  "target_groups": ["适用人群标签(flexible_employment/unemployed/employed/4050/has_children/female/male/low_income/is_local_hukou)"],
  "subsidy_calc_method": "补贴计算方法描述",
  "amount_min": 最低补贴金额(数字,必须从文中提取具体的金额数字),
  "amount_max": 最高补贴金额(数字,必须从文中提取具体的金额数字),
  "subsidy_duration": 补贴期限(月,可选),
  "effective_date": "生效日期YYYY-MM-DD",
  "expire_date": "失效日期YYYY-MM-DD(可选)",
  "policy_url": "该政策原文的网址(从页面文本中提取完整的URL,必填)",
  "brief_summary": "用一句话概括该社保政策的要点(不超过50字)",
  "source_type": "原文类型(gov_doc/social_media/news/rumor)",
  "application_process": [{"step":1,"action":"办理步骤","description":"步骤描述"}],
  "contact_info": "咨询电话或办理地址",
  "conditions": [{"name":"条件名称","description":"条件描述","tag_match":"对应人群标签(必须是以下之一:flexible_employment/unemployed/employed/4050/has_children/female/male/low_income/is_local_hukou)"}],
  "required_documents": [{"name":"材料名称","description":"描述","source":"user/gov","optional":false}]
}

提取规则：
1. amount_min和amount_max必须是具体的数字，从文中提取，不要猜测
2. conditions中的tag_match必须是以下值之一: flexible_employment/unemployed/employed/4050/has_children/female/male/low_income/is_local_hukou
3. 如果文中没有明确的金额信息，amount_min和amount_max设为0
4. 如果是政策通知或公告（不是具体补贴政策），policy_type设为对应的险种类型
5. target_groups必须从文中提取，不要猜测`

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
	appProcJSON, _ := json.Marshal(parsed.ApplicationProcess)

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
		PublishDate:        parsed.PublishDate,
		ConfidenceScore:    e.computeConfidence(entry.SourceLevel),
		Status:             "verified",
		VersionNumber:      1,
		Conditions:         condJSON,
		RequiredDocuments:  docJSON,
		SourceID:           entry.SourceID,
		SourceName:         entry.SourceName,
		SourceURL:          entry.SourceURL,
		PolicyURL:          parsed.PolicyURL,
		PolicyTitle:        parsed.PolicyTitle,
		IssuingAuthority:   parsed.IssuingAuthority,
		DocumentNumber:     parsed.DocumentNumber,
		ApplicationProcess: appProcJSON,
		ContactInfo:        parsed.ContactInfo,
		SourceType:         parsed.SourceType,
		ExtractionMethod:   extractionMethod,
		RawTextLength:      rawTextLen,
		SplitCount:         splitCount,
	}

	// 5a. 交叉验证（如果配置了 checker）
	var supersedeOldID string
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

		if maxScore > 0.95 && parsed.RegionCode != "" {
			e.store.SaveExtractLogDetailed(entry.ID, entry.SourceID, true,
				fmt.Sprintf("exact duplicate of %s (score=%.2f), skipped", similar[0].ClaimID, maxScore),
				similar[0].ClaimID, entry.Title, e.client.ModelName(), "")
			if err := e.store.MarkExtracted(entry.ID, similar[0].ClaimID); err != nil {
				return fmt.Errorf("mark extracted: %w", err)
			}
			log.Printf("[extractor] skipped claim (exact duplicate of %s, score=%.2f) from raw_text id=%d", similar[0].ClaimID, maxScore, entry.ID)
			return nil
		} else if maxScore > 0.85 && parsed.RegionCode != "" {
			if parsed.PolicyID != "" {
				existingVer, verErr := e.store.GetMaxVersionNumber(parsed.PolicyID)
				if verErr == nil && existingVer > 0 {
					claim.VersionNumber = existingVer + 1
					log.Printf("[extractor] policy %s version %d -> %d (score=%.2f, raw_text id=%d)",
						parsed.PolicyID, existingVer, claim.VersionNumber, maxScore, entry.ID)
					if oldClaimID, oldErr := e.store.GetLatestClaimByPolicyID(parsed.PolicyID); oldErr == nil && oldClaimID != "" {
						supersedeOldID = oldClaimID
					}
				} else {
					claim.Status = "pending_review"
					claim.ConfidenceScore *= 0.9
				}
			} else {
				claim.Status = "pending_review"
				claim.ConfidenceScore *= 0.9
			}
		} else if maxScore > 0.7 && parsed.RegionCode != "" {
			claim.Status = "pending_review"
			claim.ConfidenceScore *= 0.9
		}
	}

	// 6. 入库
	if err := e.store.InsertClaim(claim); err != nil {
		return fmt.Errorf("insert claim: %w", err)
	}

	if supersedeOldID != "" {
		if err := e.store.MarkSuperseded(supersedeOldID, claim.ClaimID); err != nil {
			log.Printf("[extractor] failed to mark supersede %s -> %s: %v", supersedeOldID, claim.ClaimID, err)
		}
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

	modelLabel := e.client.ModelName()
	if e.client.UsedBackup() {
		log.Printf("[extractor] used backup model %s for raw_text id=%d", modelLabel, entry.ID)
	}
	e.store.SaveExtractLogDetailed(entry.ID, entry.SourceID, true, fmt.Sprintf("claim=%s", claim.ClaimID), claim.ClaimID, entry.Title, modelLabel, parsed.BriefSummary)
	log.Printf("[extractor] extracted claim %s from raw_text id=%d", claim.ClaimID, entry.ID)
	return nil
}

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
		if len(p.ApplicationProcess) > 0 && len(merged.ApplicationProcess) == 0 {
			merged.ApplicationProcess = p.ApplicationProcess
		}
	}
	return merged
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
	PublishDate       string                   `json:"publish_date"`
	Conditions        []map[string]interface{} `json:"conditions"`
	RequiredDocuments []map[string]interface{} `json:"required_documents"`
	PolicyURL         string                   `json:"policy_url"`
	BriefSummary      string                   `json:"brief_summary"`
	PolicyTitle       string                   `json:"policy_title"`
	IssuingAuthority  string                   `json:"issuing_authority"`
	DocumentNumber    string                   `json:"document_number"`
	ApplicationProcess []map[string]interface{} `json:"application_process"`
	ContactInfo       string                   `json:"contact_info"`
	SourceType        string                   `json:"source_type"`
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

	// 提取页面标题（政府政策页面的标题很有用）
	title := ""
	titleRe := regexp.MustCompile(`<title>([^<]+)</title>`)
	if m := titleRe.FindStringSubmatch(html); len(m) > 1 {
		title = strings.TrimSpace(m[1])
		title = strings.ReplaceAll(title, " - 抖音", "")
		title = strings.ReplaceAll(title, " | 抖音", "")
	}

	// 提取 meta description（常有政策摘要）
	desc := ""
	descRe := regexp.MustCompile(`<meta\s+name="description"\s+content="([^"]+)"`)
	if m := descRe.FindStringSubmatch(html); len(m) > 1 {
		desc = strings.TrimSpace(m[1])
	}
	if desc == "" {
		descRe2 := regexp.MustCompile(`<meta\s+property="og:description"\s+content="([^"]+)"`)
		if m := descRe2.FindStringSubmatch(html); len(m) > 1 {
			desc = strings.TrimSpace(m[1])
		}
	}

	// 去除 <style>，<script>，<!-- comments -->
	html = safeReplace(`(?is)<style[^>]*>.*?</style>`, html, "")
	html = safeReplace(`(?is)<script[^>]*>.*?</script>`, html, "")
	html = safeReplace(`(?is)<!--.*?-->`, html, "")

	// 去除 HTML 标签
	html = safeReplace(`<[^>]*>`, html, "")

	// 解码 HTML 实体
	html = strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", "\"", "&apos;", "'").Replace(html)

	// 合并空白
	html = safeReplace(`\s+`, html, " ")

	// 用标点分段
	html = strings.ReplaceAll(html, "。", "。\n")
	html = strings.ReplaceAll(html, "；", "；\n")
	html = strings.ReplaceAll(html, "，", "，\n")

	// 过滤短行（降低阈值到 5 个字符以保留更多中文片段）
	lines := strings.Split(html, "\n")
	var cleaned []string
	if title != "" {
		cleaned = append(cleaned, title)
	}
	if desc != "" && desc != title && !strings.Contains(title, desc) && !strings.Contains(desc, title) {
		cleaned = append(cleaned, desc)
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len([]rune(line)) > 5 {
			cleaned = append(cleaned, line)
		}
	}

	result := strings.Join(cleaned, "\n")
	if len([]rune(result)) > 8000 {
		result = string([]rune(result)[:8000])
	}
	return result
}
