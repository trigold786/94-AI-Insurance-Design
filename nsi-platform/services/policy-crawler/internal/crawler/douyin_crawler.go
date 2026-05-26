package crawler

import (
	"crypto/sha256"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

var douyinVideoRe = regexp.MustCompile(`href="([^"]*?douyin\.com/video/\d+[^"]*?)"`)

type DouyinCrawler struct {
	config    SourceConfig
	renderer  PageRenderer
	maxItems  int
	processed map[string]bool
}

func NewDouyinCrawler(cfg SourceConfig) *DouyinCrawler {
	return &DouyinCrawler{
		config:    cfg,
		maxItems:  50,
		processed: make(map[string]bool),
	}
}

func (d *DouyinCrawler) SetRenderer(r PageRenderer) { d.renderer = r }
func (d *DouyinCrawler) SourceID() string            { return d.config.SourceID }
func (d *DouyinCrawler) SourceLevel() string          { return d.config.SourceLevel }

func (d *DouyinCrawler) Interval() time.Duration {
	if d.config.IntervalSec <= 0 {
		return 168 * time.Hour
	}
	return time.Duration(d.config.IntervalSec) * time.Second
}

func (d *DouyinCrawler) Fetch() ([]*CrawlResult, error) {
	if d.config.SourceURL == "" {
		return nil, nil
	}
	if d.renderer == nil {
		log.Printf("[douyin] %s: Chrome renderer not available, skipping", d.config.SourceID)
		return nil, nil
	}

	urls := parseDouyinURLs(d.config.SourceURL)
	var videoURLs []string
	for _, u := range urls {
		if isDouyinUserURL(u) {
			discovered, err := d.discoverVideosFromUserPage(u)
			if err != nil {
				log.Printf("[douyin] discover videos from %s error: %v", u, err)
				continue
			}
			videoURLs = append(videoURLs, discovered...)
		} else {
			videoURLs = append(videoURLs, u)
		}
	}

	var results []*CrawlResult
	for _, u := range videoURLs {
		if d.processed[u] {
			continue
		}
		d.processed[u] = true

		result, err := d.fetchVideoPage(u)
		if err != nil {
			log.Printf("[douyin] fetch %s error: %v", u, err)
			continue
		}
		if result != nil {
			results = append(results, result)
		}
		if len(results) >= d.maxItems {
			break
		}
	}
	return results, nil
}

func (d *DouyinCrawler) discoverVideosFromUserPage(userURL string) ([]string, error) {
	cleanURL := stripQueryParams(userURL)
	log.Printf("[douyin] discovering videos from user page: %s", cleanURL)

	html, err := d.renderer.RenderWithVirtualTime(cleanURL, 15000)
	if err != nil {
		return nil, fmt.Errorf("render user page: %w", err)
	}

	matches := douyinVideoRe.FindAllStringSubmatch(html, -1)
	seen := make(map[string]bool)
	var videoURLs []string
	for _, m := range matches {
		raw := m[1]
		raw = stripQueryParams(raw)
		if seen[raw] {
			continue
		}
		seen[raw] = true
		videoURLs = append(videoURLs, raw)
	}

	log.Printf("[douyin] discovered %d videos from user page %s", len(videoURLs), cleanURL)
	return videoURLs, nil
}

func isDouyinUserURL(u string) bool {
	return strings.Contains(u, "douyin.com/user/")
}

func stripQueryParams(u string) string {
	if idx := strings.Index(u, "?"); idx >= 0 {
		return u[:idx]
	}
	return u
}

func (d *DouyinCrawler) fetchVideoPage(videoURL string) (*CrawlResult, error) {
	html, err := d.renderer.Render(videoURL)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	title := extractDouyinTitle(html)
	desc := extractDouyinDesc(html)
	content := title
	if desc != "" {
		content = title + "\n" + desc
	}
	if content == "" {
		content = html
	}

	hash := sha256.Sum256([]byte(videoURL))
	return &CrawlResult{
		SourceID:    d.config.SourceID,
		SourceLevel: d.config.SourceLevel,
		RawText:     content,
		Title:       title,
		SourceURL:   videoURL,
		FetchedAt:   time.Now(),
		VersionHash: fmt.Sprintf("%x", hash),
	}, nil
}

func parseDouyinURLs(rawURL string) []string {
	var urls []string
	seen := make(map[string]bool)
	for _, part := range strings.Split(rawURL, "\n") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "http") && strings.Contains(part, "douyin.com") {
			if !seen[part] {
				seen[part] = true
				urls = append(urls, part)
			}
		}
	}
	if len(urls) == 0 && strings.HasPrefix(rawURL, "http") && strings.Contains(rawURL, "douyin.com") {
		urls = append(urls, rawURL)
	}
	return urls
}

func extractDouyinTitle(html string) string {
	markers := []string{
		`<title>`,
		`property="og:title" content="`,
		`name="description" content="`,
	}
	for _, m := range markers {
		idx := strings.Index(html, m)
		if idx == -1 {
			continue
		}
		start := idx + len(m)
		end := strings.Index(html[start:], `"`)
		if end == -1 {
			end = strings.Index(html[start:], "<")
		}
		if end == -1 {
			continue
		}
		title := html[start : start+end]
		title = strings.TrimSpace(title)
		title = strings.ReplaceAll(title, " - 抖音", "")
		title = strings.ReplaceAll(title, " | 抖音", "")
		if title != "" && title != "抖音" {
			return title
		}
	}
	return ""
}

func extractDouyinDesc(html string) string {
	marker := `name="description" content="`
	idx := strings.Index(html, marker)
	if idx == -1 {
		marker = `property="og:description" content="`
		idx = strings.Index(html, marker)
	}
	if idx == -1 {
		return ""
	}
	start := idx + len(marker)
	end := strings.Index(html[start:], `"`)
	if end == -1 {
		return ""
	}
	desc := html[start : start+end]
	return strings.TrimSpace(desc)
}

