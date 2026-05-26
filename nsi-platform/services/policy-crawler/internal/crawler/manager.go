package crawler

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/parser"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

// CrawlerManager 爬取管理器：管理所有数据源的爬取调度
type CrawlerManager struct {
	store     Store
	claimDB   ClaimDB
	crawlers  []Source
	stopCh    chan struct{}
	renderer  PageRenderer
}

// ClaimDB 政策原子存储接口
type ClaimDB interface {
	Ingest(claim *models.PolicyClaim) error
	ListByRegionAndType(regionCode, policyType string) ([]models.PolicyClaim, error)
}

func NewCrawlerManager(store Store, claimDB ClaimDB) *CrawlerManager {
	return &CrawlerManager{
		store:   store,
		claimDB: claimDB,
		stopCh:  make(chan struct{}),
	}
}

func (m *CrawlerManager) SetRenderer(r PageRenderer) { m.renderer = r }

func (m *CrawlerManager) Init(sources []SourceConfig, watchDir string) {
	for _, cfg := range sources {
		if !cfg.Enabled {
			continue
		}
		var s Source
		switch cfg.CrawlType {
		case "govsite":
			gc := NewGovSiteCrawler(cfg)
			if m.renderer != nil {
				gc.SetRenderer(m.renderer)
			}
			s = gc
		case "file":
			s = NewFileWatcherCrawler(cfg, watchDir)
		case "rss":
			s = NewRSSCrawler(cfg)
		case "manual":
			s = NewManualCrawler(cfg)
		case "douyin":
			dc := NewDouyinCrawler(cfg)
			if m.renderer != nil {
				dc.SetRenderer(m.renderer)
			}
			s = dc
		case "wechat":
			wc := NewWeChatCrawler(cfg)
			if m.renderer != nil {
				wc.SetRenderer(m.renderer)
			}
			s = wc
		default:
			log.Printf("[crawler] unknown crawl type %q for source %s", cfg.CrawlType, cfg.SourceID)
			continue
		}
		m.crawlers = append(m.crawlers, s)
		log.Printf("[crawler] registered source %s (%s, interval=%v)", cfg.SourceID, cfg.SourceLevel, s.Interval())
	}
}

func (m *CrawlerManager) CrawlAll() {
	for _, s := range m.crawlers {
		m.crawlAndProcess(s)
	}
	log.Printf("[crawler] crawled all %d sources", len(m.crawlers))
}

func (m *CrawlerManager) CrawlSource(sourceID string) {
	for _, s := range m.crawlers {
		if s.SourceID() == sourceID {
			m.crawlAndProcess(s)
			return
		}
	}
	if m.store != nil {
		s, err := m.loadAndRegisterSource(sourceID)
		if err != nil {
			log.Printf("[crawler] dynamic load source %s error: %v", sourceID, err)
			return
		}
		if s != nil {
			m.crawlAndProcess(s)
			return
		}
	}
	log.Printf("[crawler] source %s not found for crawl", sourceID)
}

func (m *CrawlerManager) loadAndRegisterSource(sourceID string) (Source, error) {
	cfgs, err := m.store.ListEnabledSources()
	if err != nil {
		return nil, fmt.Errorf("query sources: %w", err)
	}
	for _, cfg := range cfgs {
		if cfg.SourceID != sourceID {
			continue
		}
		for _, existing := range m.crawlers {
			if existing.SourceID() == sourceID {
				return existing, nil
			}
		}
		var s Source
		switch cfg.CrawlType {
		case "govsite":
			gc := NewGovSiteCrawler(cfg)
			if m.renderer != nil {
				gc.SetRenderer(m.renderer)
			}
			s = gc
		case "file":
			s = NewFileWatcherCrawler(cfg, "")
		case "rss":
			s = NewRSSCrawler(cfg)
		case "manual":
			s = NewManualCrawler(cfg)
		case "douyin":
			dc := NewDouyinCrawler(cfg)
			if m.renderer != nil {
				dc.SetRenderer(m.renderer)
			}
			s = dc
		case "wechat":
			wc := NewWeChatCrawler(cfg)
			if m.renderer != nil {
				wc.SetRenderer(m.renderer)
			}
			s = wc
		default:
			return nil, fmt.Errorf("unknown crawl type %q", cfg.CrawlType)
		}
		m.crawlers = append(m.crawlers, s)
		log.Printf("[crawler] dynamically registered source %s (%s)", cfg.SourceID, cfg.CrawlType)
		return s, nil
	}
	return nil, nil
}

func (m *CrawlerManager) Stop() {
	close(m.stopCh)
}

func (m *CrawlerManager) crawlAndProcess(s Source) {
	log.Printf("[crawler] fetching source %s (%s)", s.SourceID(), s.SourceLevel())

	results, err := s.Fetch()
	if err != nil {
		log.Printf("[crawler] fetch error for %s: %v", s.SourceID(), err)
		m.store.SaveCrawlLog(s.SourceID(), false, err.Error())
		return
	}
	if len(results) == 0 {
		m.store.SaveCrawlLogWithDetails(s.SourceID(), true, "no new content", "", "")
		return // 没有新内容
	}

	for _, result := range results {
		if result == nil {
			continue
		}
		// 存储原文（始终保存原始抓取内容）
		m.store.SaveRawText(s.SourceID(), result.Title, result.RawText, result.SourceURL, result.VersionHash)
		log.Printf("[crawler] fetched %d bytes from %s (%s)", len(result.RawText), s.SourceID(), result.SourceURL)

		// 判断内容类型：HTML 页面 vs 结构化文本
		isHTML := !strings.Contains(result.RawText, "政策ID:")
		if !isHTML && len(result.RawText) > 0 {
			first := strings.TrimSpace(result.RawText)
			isHTML = first[0] == '<' || strings.Contains(first, "<body") || strings.Contains(first, "<html")
		}

		if isHTML {
			log.Printf("[crawler] %s returned HTML page (gov site), stored raw. LLM extraction pending.", s.SourceID())
			summary := truncateSummary(result.Title, 120)
			if summary == "" {
				summary = truncateSummary(result.SourceURL, 120)
			}
			m.store.SaveCrawlLogWithDetails(s.SourceID(), true, "HTML stored, awaiting LLM extraction", "", summary)
			continue
		}

		// NLP 解析（仅对结构化文本）
		parsed, conditions, docs, err := parser.ParseStructuredText(result.RawText)
		if err != nil {
			log.Printf("[crawler] parse error for %s: %v", s.SourceID(), err)
			m.store.SaveCrawlLog(s.SourceID(), false, "parse error: "+err.Error())
			continue
		}

		condJSON, _ := json.Marshal(conditions)
		docJSON, _ := json.Marshal(docs)

		// 交叉验证 & 置信度评分 (PRD §4.2.2)
		confidence := m.calculateConfidence(s.SourceLevel(), parsed)
		status := decideStatus(confidence)

		// 构建政策原子
		claim := &models.PolicyClaim{
			ClaimID:           fmt.Sprintf("CRAWL-%d", time.Now().UnixNano()),
			PolicyID:          parsed.PolicyID,
			RegionCode:        parsed.RegionCode,
			PolicyType:        parsed.PolicyType,
			TargetGroupTags:   parsed.TargetGroups,
			SubsidyCalcMethod: parsed.SubsidyCalcMethod,
			SubsidyAmountMin:  parsed.AmountMin,
			SubsidyAmountMax:  parsed.AmountMax,
			SubsidyDuration:   parsed.SubsidyDuration,
			EffectiveDate:     parsed.EffectiveDate,
			ExpireDate:        parsed.ExpireDate,
			ConfidenceScore:   confidence,
			Status:            status,
			VersionNumber:     1,
			Conditions:        condJSON,
			RequiredDocuments: docJSON,
			SourceID:          s.SourceID(),
			SourceURL:         result.SourceURL,
			SourceName:        "",
		}

		// 入库
		if err := m.claimDB.Ingest(claim); err != nil {
			log.Printf("[crawler] store error for %s: %v", s.SourceID(), err)
			m.store.SaveCrawlLog(s.SourceID(), false, "store error: "+err.Error())
			continue
		}

		m.store.SaveCrawlLogWithDetails(s.SourceID(), true, "", claim.ClaimID, truncateSummary(claim.SubsidyCalcMethod, 120))
		log.Printf("[crawler] stored claim %s (confidence=%.2f, status=%s)", claim.ClaimID, confidence, status)
	}
}

// calculateConfidence 多因子加权置信度评分
func truncateSummary(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// 因子: 数据源权威性(0.30) + 交叉验证匹配(0.25) + 内容结构化程度(0.20) + 数据完整性(0.15) + 时效性(0.10)
func (m *CrawlerManager) calculateConfidence(sourceLevel string, claim *parser.PolicyClaim) float64 {
	// 1. 数据源权威性权重 (0.30)
	var wSource float64
	switch sourceLevel {
	case "HIGH":
		wSource = 1.0
	case "MEDIUM":
		wSource = 0.7
	case "LOW":
		wSource = 0.3
	default:
		wSource = 0.5
	}
	sourceScore := wSource

	// 2. 交叉验证匹配度 (0.25)
	// 同 region + 同 type 的已有政策中已通过的比例越高越可信
	var matchRate float64
	if m.claimDB != nil {
		existing, err := m.claimDB.ListByRegionAndType(claim.RegionCode, claim.PolicyType)
		if err == nil && len(existing) > 0 {
			var verifiedCount int
			for _, e := range existing {
				if e.Status == "verified" {
					verifiedCount++
				}
			}
			matchRate = float64(verifiedCount) / float64(len(existing))
			// 如果全部通过则满分，否则按比例降分
			if verifiedCount == 0 {
				matchRate = 0.1 // 至少给一点分表示存在同类政策
			}
		} else if err == nil {
			matchRate = 0.5 // 暂无同类政策可对比，给中间分
		}
	}
	crossScore := matchRate

	// 3. 内容结构化程度 (0.20)
	// 解析出的字段越完整分数越高
	var fieldScore float64
	fields := 0
	totalFields := 6
	if claim.PolicyID != "" {
		fields++
	}
	if claim.RegionCode != "" && claim.RegionCode != "000000" {
		fields++
	}
	if claim.PolicyType != "" {
		fields++
	}
	if claim.TargetGroups != nil && len(claim.TargetGroups) > 0 {
		fields++
	}
	if claim.SubsidyCalcMethod != "" {
		fields++
	}
	if claim.AmountMin != nil || claim.AmountMax != nil {
		fields++
	}
	fieldScore = float64(fields) / float64(totalFields)
	if fieldScore < 0.3 {
		fieldScore = 0.3 // 最小保证分
	}

	// 4. 数据完整性 (0.15)
	// 有效的金额范围和期限信息
	var completeness float64
	compFields := 0
	if claim.AmountMin != nil && *claim.AmountMin > 0 {
		compFields++
	}
	if claim.SubsidyDuration != nil && *claim.SubsidyDuration > 0 {
		compFields++
	}
	if claim.EffectiveDate != "" {
		compFields++
	}
	completeness = float64(compFields) / 3.0

	// 5. 时效性 (0.10)
	// 新的政策比旧的政策更可信
	var timeliness float64 = 0.5
	if claim.EffectiveDate != "" {
		timeliness = 0.8 // 有生效日期的相对可靠
	}

	// 加权综合评分
	score := 0.30*sourceScore + 0.25*crossScore + 0.20*fieldScore + 0.15*completeness + 0.10*timeliness

	// 边界裁剪
	if score > 1.0 {
		score = 1.0
	}
	if score < 0 {
		score = 0
	}
	return score
}

func decideStatus(score float64) string {
	switch {
	case score >= 0.85:
		return "verified"
	case score >= 0.6:
		return "pending_review"
	default:
		return "unverified"
	}
}


