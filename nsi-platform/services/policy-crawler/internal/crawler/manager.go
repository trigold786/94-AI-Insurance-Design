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
	log.Printf("[crawler] source %s not found for crawl", sourceID)
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
			m.store.SaveCrawlLog(s.SourceID(), true, "HTML stored, awaiting LLM extraction")
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

		m.store.SaveCrawlLog(s.SourceID(), true, "")
		log.Printf("[crawler] stored claim %s (confidence=%.2f, status=%s)", claim.ClaimID, confidence, status)
	}
}

// calculateConfidence 置信度评分 (PRD §4.2.2)
func (m *CrawlerManager) calculateConfidence(sourceLevel string, claim *parser.PolicyClaim) float64 {
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

	// 尝试交叉验证：查找同 region + 同 type 的已有政策
	var matchCount, conflictCount int
	if m.claimDB != nil {
		existing, err := m.claimDB.ListByRegionAndType(claim.RegionCode, claim.PolicyType)
		if err == nil {
			for _, e := range existing {
				if e.Status != "unverified" {
					matchCount++
				}
			}
		}
	}

	// 简化评分公式 (PRD §4.2.2 完整公式需分布式权重配置)
	score := (0.4*wSource + 0.3*float64(matchCount)*0.1 - 0.1*float64(conflictCount)*0.1 + 0.2*0.8 + 0.1*0) / 0.7
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


