package crawler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/parser"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/verifier"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

// CrawlerManager 爬取管理器：管理所有数据源的爬取调度
type CrawlerManager struct {
	store       Store
	dbStore     *DBStore
	claimDB     ClaimDB
	crawlers    []Source
	sourceCfgs  map[string]SourceConfig
	stopCh      chan struct{}
	renderer    PageRenderer
	filter      *RelevanceFilter
	videoWorker *VideoExtractWorker
}

// ClaimDB 政策原子存储接口
type ClaimDB interface {
	Ingest(claim *models.PolicyClaim) error
	ListByRegionAndType(regionCode, policyType string) ([]models.PolicyClaim, error)
}

func NewCrawlerManager(store Store, dbStore *DBStore, claimDB ClaimDB) *CrawlerManager {
	return &CrawlerManager{
		store:      store,
		dbStore:    dbStore,
		claimDB:    claimDB,
		sourceCfgs: make(map[string]SourceConfig),
		stopCh:     make(chan struct{}),
	}
}

func (m *CrawlerManager) SetRenderer(r PageRenderer) { m.renderer = r }

func (m *CrawlerManager) GetFilter() *RelevanceFilter { return m.filter }

func (m *CrawlerManager) InitFilterAndWorker(db *sql.DB, asrCfg ASRConfig) {
	m.filter = NewRelevanceFilter(nil)
	m.filter.LoadFromDB(db)
	go m.filter.StartReloadLoop(5*time.Minute, m.stopCh)

	var asr ASRProvider
	if asrCfg.Enabled {
		asr = NewASRProviderFromConfig(asrCfg)
	}
	m.videoWorker = NewVideoExtractWorker(m.dbStore, m.filter, asr, 2)
	m.videoWorker.Start()
	m.recoverPendingVideoExtracts()
}

func (m *CrawlerManager) recoverPendingVideoExtracts() {
	if m.dbStore == nil {
		return
	}
	tasks, err := m.dbStore.GetPendingVideoExtracts()
	if err != nil {
		log.Printf("[video-extract] recovery query failed: %v", err)
		return
	}
	for _, t := range tasks {
		m.videoWorker.Queue() <- VideoExtractTask{
			RawTextID: t.ID,
			SourceID:  t.SourceID,
			VideoURL:  t.VideoURL,
			Title:     t.Title,
		}
	}
	if len(tasks) > 0 {
		log.Printf("[video-extract] recovered %d pending tasks", len(tasks))
	}
}

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
			dc := NewDouyinCrawler(cfg, m.filter)
			if m.renderer != nil {
				dc.SetRenderer(m.renderer)
			}
			s = dc
		case "wechat":
			wc := NewWeChatCrawler(cfg, m.filter)
			if m.renderer != nil {
				wc.SetRenderer(m.renderer)
			}
			s = wc
		case "bilibili":
			s = NewBilibiliCrawler(cfg, m.filter)
		default:
			log.Printf("[crawler] unknown crawl type %q for source %s", cfg.CrawlType, cfg.SourceID)
			continue
		}
		m.crawlers = append(m.crawlers, s)
		m.sourceCfgs[cfg.SourceID] = cfg
		log.Printf("[crawler] registered source %s (%s, interval=%v)", cfg.SourceID, cfg.SourceLevel, s.Interval())
	}
}

func (m *CrawlerManager) CrawlAll() {
	for i, s := range m.crawlers {
		if i > 0 {
			delay := m.getDelayForSource(s.SourceID())
			if delay > 0 {
				log.Printf("[crawler] waiting %v before crawling %s", delay, s.SourceID())
				time.Sleep(delay)
			}
		}
		m.crawlAndProcess(s)
	}
	log.Printf("[crawler] crawled all %d sources", len(m.crawlers))
}

func (m *CrawlerManager) getDelayForSource(sourceID string) time.Duration {
	cfg, ok := m.sourceCfgs[sourceID]
	if !ok || cfg.RequestDelayMs <= 0 {
		return 2 * time.Second
	}
	return time.Duration(cfg.RequestDelayMs) * time.Millisecond
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
			dc := NewDouyinCrawler(cfg, m.filter)
			if m.renderer != nil {
				dc.SetRenderer(m.renderer)
			}
			s = dc
		case "wechat":
			wc := NewWeChatCrawler(cfg, m.filter)
			if m.renderer != nil {
				wc.SetRenderer(m.renderer)
			}
			s = wc
		case "bilibili":
			s = NewBilibiliCrawler(cfg, m.filter)
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

		if result.NeedsVideoExtract && m.videoWorker != nil && m.dbStore != nil {
			rawTextID, err := m.dbStore.SaveRawTextReturningID(result.SourceID, result.Title, result.RawText, result.SourceURL, result.VersionHash)
			if err != nil {
				log.Printf("[crawler] save raw text error for %s: %v", s.SourceID(), err)
				continue
			}
			m.dbStore.SetVideoExtractStatus(rawTextID, "pending")
			m.videoWorker.Queue() <- VideoExtractTask{
				RawTextID: rawTextID,
				SourceID:  result.SourceID,
				VideoURL:  result.VideoURL,
				Title:     result.Title,
			}
			m.store.SaveCrawlLogWithDetails(s.SourceID(), true, "pending video extraction", "", truncateSummary(result.Title, 120))
			continue
		}

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
		status := verifier.DecideStatus(confidence)

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

func (m *CrawlerManager) calculateConfidence(sourceLevel string, claim *parser.PolicyClaim) float64 {
	matchRate := 0.5
	if m.claimDB != nil {
		existing, err := m.claimDB.ListByRegionAndType(claim.RegionCode, claim.PolicyType)
		if err == nil && len(existing) > 0 {
			verifiedCount := 0
			for _, e := range existing {
				if e.Status == "verified" {
					verifiedCount++
				}
			}
			if verifiedCount > 0 {
				matchRate = float64(verifiedCount) / float64(len(existing))
			} else {
				matchRate = 0.1
			}
		}
	}

	hasAmount := claim.AmountMin != nil || claim.AmountMax != nil
	fields := 0
	if claim.PolicyID != "" { fields++ }
	if claim.RegionCode != "" && claim.RegionCode != "000000" { fields++ }
	if claim.PolicyType != "" { fields++ }
	if claim.TargetGroups != nil && len(claim.TargetGroups) > 0 { fields++ }
	if claim.SubsidyCalcMethod != "" { fields++ }
	if hasAmount { fields++ }
	structureScore := float64(fields) / 6.0
	if structureScore < 0.3 { structureScore = 0.3 }

	completeness := 0.0
	compFields := 0
	if claim.AmountMin != nil && *claim.AmountMin > 0 { compFields++ }
	if claim.SubsidyDuration != nil && *claim.SubsidyDuration > 0 { compFields++ }
	if claim.EffectiveDate != "" { compFields++ }
	completeness = float64(compFields) / 3.0

	timeliness := 0.5
	if claim.EffectiveDate != "" { timeliness = 0.8 }

	mc := &models.PolicyClaim{
		SourceLevel: sourceLevel,
		MatchRate:   matchRate,
		ConflictScore: 1.0 - structureScore*0.3,
		FetchedAt:   time.Now().Format("2006-01-02"),
	}
	prdScore := verifier.CalculateConfidence(mc, verifier.DefaultConfidenceConfig())

	combinedScore := 0.6*prdScore + 0.15*structureScore + 0.15*completeness + 0.10*timeliness
	if combinedScore > 1.0 { combinedScore = 1.0 }
	if combinedScore < 0 { combinedScore = 0 }
	return combinedScore
}


