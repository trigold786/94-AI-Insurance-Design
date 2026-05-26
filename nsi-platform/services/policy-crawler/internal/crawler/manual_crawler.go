package crawler

import "time"

type ManualCrawler struct {
	config SourceConfig
}

func NewManualCrawler(cfg SourceConfig) *ManualCrawler {
	return &ManualCrawler{config: cfg}
}

func (m *ManualCrawler) SourceID() string   { return m.config.SourceID }
func (m *ManualCrawler) SourceLevel() string { return m.config.SourceLevel }

func (m *ManualCrawler) Interval() time.Duration {
	return 0
}

func (m *ManualCrawler) Fetch() ([]*CrawlResult, error) {
	return nil, nil
}
