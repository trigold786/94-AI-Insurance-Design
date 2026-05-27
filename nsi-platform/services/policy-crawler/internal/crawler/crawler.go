package crawler

import "time"

// Source 政策数据源接口 (PRD §4.1.1)
type Source interface {
	SourceID() string
	SourceLevel() string // HIGH, MEDIUM, LOW
	Interval() time.Duration
	Fetch() ([]*CrawlResult, error)
}

// CrawlResult 单次爬取结果
type CrawlResult struct {
	SourceID          string
	SourceLevel       string
	RawText           string
	Title             string
	SourceURL         string
	FetchedAt         time.Time
	VersionHash       string
	Error             string
	VideoURL          string
	NeedsVideoExtract bool
	ContentType       string
}

// SourceConfig 来源配置（对应 DB sources 表）
type SourceConfig struct {
	SourceID       string `db:"source_id" json:"source_id"`
	SourceName     string `db:"source_name" json:"source_name"`
	SourceURL      string `db:"source_url" json:"source_url"`
	SourceLevel    string `db:"source_level" json:"source_level"`
	CrawlType      string `db:"crawl_type" json:"crawl_type"`
	IntervalSec    int    `db:"interval_sec" json:"interval_sec"`
	RegionCode     string `db:"region_code" json:"region_code"`
	Enabled        bool   `db:"enabled" json:"enabled"`
	ProxyURL       string `db:"proxy_url" json:"proxy_url"`
	RequestDelayMs int    `db:"request_delay_ms" json:"request_delay_ms"`
	MaxConcurrent  int    `db:"max_concurrent" json:"max_concurrent"`
	RespectRobots  bool   `db:"respect_robots" json:"respect_robots"`
}

// Store 爬取结果持久化接口
type Store interface {
	ListEnabledSources() ([]SourceConfig, error)
	SaveRawText(sourceID, title, content, sourceURL, versionHash string) error
	SaveCrawlLog(sourceID string, success bool, errMsg string)
	SaveCrawlLogWithDetails(sourceID string, success bool, errMsg string, claimID string, summary string)
}
