package crawler

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var douyinVideoRe = regexp.MustCompile(`href="([^"]*?douyin\.com/video/\d+[^"]*?)"`)
var douyinAuthorRe = regexp.MustCompile(`@([\p{L}\p{N}\-_]+)`)

type DouyinCrawler struct {
	config    SourceConfig
	renderer  PageRenderer
	filter    *RelevanceFilter
	maxItems  int
	processed map[string]bool
}

func NewDouyinCrawler(cfg SourceConfig, filter *RelevanceFilter) *DouyinCrawler {
	return &DouyinCrawler{
		config:    cfg,
		filter:    filter,
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
	log.Printf("[douyin] discovering own videos from user page: %s", cleanURL)

	if d.renderer != nil {
		worksURL := cleanURL + "?showTab=works"
		html, err := d.renderer.RenderWithVirtualTime(worksURL, 30000)
		if err == nil {
			videoURLs := extractDouyinVideoURLs(html)
			if len(videoURLs) > 0 {
				log.Printf("[douyin] discovered %d videos via Chrome (works tab) from %s", len(videoURLs), cleanURL)
				return videoURLs, nil
			}
		}
		log.Printf("[douyin] Chrome render failed for works tab %s: %v", cleanURL, err)
	}

	html, err := httpFetch(cleanURL)
	if err != nil {
		return nil, fmt.Errorf("discover videos from %s: %w", cleanURL, err)
	}
	videoURLs := extractDouyinVideoURLs(html)
	if len(videoURLs) == 0 {
		videoURLs = extractDouyinVideoLinks(html)
	}
	if len(videoURLs) == 0 {
		return nil, fmt.Errorf("no videos discovered from %s", cleanURL)
	}
	log.Printf("[douyin] discovered %d videos via HTTP from %s", len(videoURLs), cleanURL)
	return videoURLs, nil
}

var douyinShareURLRe = regexp.MustCompile(`https?://(?:www\.)?douyin\.com/video/\d+`)

func extractDouyinVideoURLs(html string) []string {
	matches := douyinVideoRe.FindAllStringSubmatch(html, -1)
	seen := make(map[string]bool)
	var urls []string
	for _, m := range matches {
		raw := m[1]
		raw = stripQueryParams(raw)
		if seen[raw] {
			continue
		}
		seen[raw] = true
		urls = append(urls, raw)
	}
	return urls
}

func extractDouyinVideoLinks(html string) []string {
	matches := douyinShareURLRe.FindAllString(html, -1)
	seen := make(map[string]bool)
	var urls []string
	for _, u := range matches {
		u = stripQueryParams(u)
		if seen[u] {
			continue
		}
		seen[u] = true
		urls = append(urls, u)
	}
	return urls
}

func isDouyinUserURL(u string) bool {
	return strings.Contains(u, "douyin.com/user/")
}

func extractNicknameFromSourceID(sourceID string) string {
	if strings.HasPrefix(sourceID, "DOUYIN-") {
		return sourceID[len("DOUYIN-"):]
	}
	return ""
}

// extractDouyinAuthorNickname finds the first @mention in a Douyin video page
func extractDouyinAuthorNickname(html string) string {
	markers := []string{
		`name="description" content="`,
		`property="og:description" content="`,
		`property="og:title" content="`,
	}
	for _, m := range markers {
		idx := strings.Index(html, m)
		if idx == -1 {
			continue
		}
		start := idx + len(m)
		end := strings.Index(html[start:], `"`)
		if end == -1 || end > 500 {
			continue
		}
		snippet := html[start : start+end]
		matches := douyinAuthorRe.FindStringSubmatch(snippet)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func stripQueryParams(u string) string {
	if idx := strings.Index(u, "?"); idx >= 0 {
		return u[:idx]
	}
	return u
}

func (d *DouyinCrawler) fetchVideoPage(videoURL string) (*CrawlResult, error) {
	var html string
	var err error

	if d.renderer != nil {
		html, err = d.renderer.RenderWithVirtualTime(videoURL, 15000)
		if err != nil {
			log.Printf("[douyin] Chrome render failed for %s, falling back to HTTP: %v", videoURL, err)
		}
	}

	if html == "" {
		html, err = httpFetch(videoURL)
		if err != nil {
			return nil, fmt.Errorf("http fetch: %w", err)
		}
	}

	expectedNickname := extractNicknameFromSourceID(d.config.SourceID)
	if expectedNickname != "" {
		htmlLower := strings.ToLower(html)
		actualAuthor := extractDouyinAuthorNickname(html)
		if strings.Contains(html, "@"+expectedNickname) || strings.Contains(htmlLower, "@"+strings.ToLower(expectedNickname)) {
		} else if actualAuthor != "" {
			actualLower := strings.ToLower(actualAuthor)
			expectedLower := strings.ToLower(expectedNickname)
			if !strings.Contains(actualLower, expectedLower) && !strings.Contains(expectedLower, actualLower) {
				log.Printf("[douyin] skipping video %s: author %q does not match expected %q", videoURL, actualAuthor, expectedNickname)
				return nil, nil
			}
		}
	}

	title := extractDouyinTitle(html)
	desc := extractDouyinDesc(html)
	content := title
	if desc != "" {
		content = title + "\n" + desc
	}
	if content == "" {
		content = extractDouyinTextFromHTML(html)
	}
	if content == "" {
		return nil, nil
	}

	if d.filter != nil {
		score, matched := d.filter.Score(title+" "+desc, d.config.SourceID, "douyin")
		threshold := d.filter.MinScore(d.config.SourceID, "level1")
		if score < threshold {
			log.Printf("[douyin] filtered out %s: relevance score %d < threshold %d (matched: %v)", videoURL, score, threshold, matched)
			return nil, nil
		}
		log.Printf("[douyin] passed relevance filter for %s: score=%d matched=%v", videoURL, score, matched)
	}

	hash := sha256.Sum256([]byte(videoURL))
	return &CrawlResult{
		SourceID:          d.config.SourceID,
		SourceLevel:       d.config.SourceLevel,
		RawText:           content,
		Title:             title,
		SourceURL:         videoURL,
		FetchedAt:         time.Now(),
		VersionHash:       fmt.Sprintf("%x", hash),
		VideoURL:          videoURL,
		NeedsVideoExtract: true,
		ContentType:       "video-meta",
	}, nil
}

func httpFetch(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
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

	// fallback: try <title>.*</title> via regex
	re := regexp.MustCompile(`<title>([^<]+)</title>`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		t := strings.TrimSpace(matches[1])
		t = strings.ReplaceAll(t, " - 抖音", "")
		t = strings.ReplaceAll(t, " | 抖音", "")
		return t
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

func extractDouyinTextFromHTML(html string) string {
	stripTags := regexp.MustCompile(`<[^>]*>`)
	text := stripTags.ReplaceAllString(html, " ")
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) > 500 {
		text = string(runes[:500])
	}
	return strings.TrimSpace(text)
}

