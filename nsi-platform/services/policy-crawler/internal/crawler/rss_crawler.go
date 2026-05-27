package crawler

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"
)

type RSSCrawler struct {
	config        SourceConfig
	resilient     *ResilientHTTPClient
	maxItems      int
	processed     map[string]bool
	robotsChecker *RobotsChecker
}

func NewRSSCrawler(cfg SourceConfig) *RSSCrawler {
	maxItems := 20
	rhcCfg := DefaultHTTPClientConfig()
	rhcCfg.ProxyURL = cfg.ProxyURL
	rhcCfg.Timeout = 30 * time.Second
	return &RSSCrawler{
		config:        cfg,
		resilient:     NewResilientHTTPClient(rhcCfg),
		maxItems:      maxItems,
		processed:     make(map[string]bool),
		robotsChecker: NewRobotsChecker(),
	}
}

func (r *RSSCrawler) SourceID() string   { return r.config.SourceID }
func (r *RSSCrawler) SourceLevel() string { return r.config.SourceLevel }

func (r *RSSCrawler) Interval() time.Duration {
	if r.config.IntervalSec <= 0 {
		return 168 * time.Hour
	}
	return time.Duration(r.config.IntervalSec) * time.Second
}

func (r *RSSCrawler) Fetch() ([]*CrawlResult, error) {
	if r.config.RespectRobots {
		if err := CheckRobotsBeforeCrawl(r.robotsChecker, r.config.SourceURL, DefaultHTTPClientConfig().UserAgent); err != nil {
			return nil, err
		}
	}

	resp, err := r.resilient.Get(r.config.SourceURL)
	if err != nil {
		return nil, fmt.Errorf("RSS fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("RSS HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("RSS read: %w", err)
	}

	items, err := ParseFeed(body)
	if err != nil {
		return nil, fmt.Errorf("RSS parse: %w", err)
	}

	var results []*CrawlResult
	for i, item := range items {
		if i >= r.maxItems {
			break
		}
		hash := sha256.Sum256([]byte(item.Link))
		hashStr := fmt.Sprintf("%x", hash)
		if r.processed[hashStr] {
			continue
		}
		r.processed[hashStr] = true

		content := item.Content
		if content == "" {
			content = item.Description
		}
		if content == "" {
			content = item.Title
		}

		results = append(results, &CrawlResult{
			SourceID:    r.config.SourceID,
			SourceLevel: r.config.SourceLevel,
			RawText:     content,
			Title:       item.Title,
			SourceURL:   item.Link,
			FetchedAt:   time.Now(),
			VersionHash: hashStr,
		})
	}

	return results, nil
}

type FeedItem struct {
	Title       string
	Link        string
	Description string
	Content     string
}

func ParseFeed(data []byte) ([]FeedItem, error) {
	trimmed := strings.TrimSpace(string(data))

	if strings.HasPrefix(trimmed, "<feed") || strings.Contains(trimmed, "<feed ") {
		return parseAtom(data)
	}
	return parseRSS(data)
}

type rssXML struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			Content     string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
		} `xml:"item"`
	} `xml:"channel"`
}

func parseRSS(data []byte) ([]FeedItem, error) {
	var feed rssXML
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	items := make([]FeedItem, 0, len(feed.Channel.Items))
	for _, it := range feed.Channel.Items {
		if it.Title == "" && it.Link == "" {
			continue
		}
		items = append(items, FeedItem{
			Title:       it.Title,
			Link:        it.Link,
			Description: it.Description,
			Content:     it.Content,
		})
	}
	return items, nil
}

type atomXML struct {
	XMLName xml.Name `xml:"feed"`
	Entries []struct {
		Title string `xml:"title"`
		Link  []struct {
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
		} `xml:"link"`
		Content string `xml:"content"`
		Summary string `xml:"summary"`
	} `xml:"entry"`
}

func parseAtom(data []byte) ([]FeedItem, error) {
	var feed atomXML
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	items := make([]FeedItem, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		link := ""
		for _, l := range entry.Link {
			if l.Rel == "" || l.Rel == "alternate" {
				link = l.Href
				break
			}
		}
		if entry.Title == "" && link == "" {
			continue
		}
		items = append(items, FeedItem{
			Title:       entry.Title,
			Link:        link,
			Description: entry.Summary,
			Content:     entry.Content,
		})
	}
	return items, nil
}
