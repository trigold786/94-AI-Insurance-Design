package crawler

import (
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var govUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

type GovSiteCrawler struct {
	config   SourceConfig
	client   *http.Client
	renderer PageRenderer
}

func NewGovSiteCrawler(cfg SourceConfig) *GovSiteCrawler {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &GovSiteCrawler{
		config: cfg,
		client: &http.Client{
			Timeout:   20 * time.Second,
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (g *GovSiteCrawler) SetRenderer(r PageRenderer) { g.renderer = r }

func (g *GovSiteCrawler) SourceID() string   { return g.config.SourceID }
func (g *GovSiteCrawler) SourceLevel() string { return g.config.SourceLevel }

func (g *GovSiteCrawler) Interval() time.Duration {
	if g.config.IntervalSec <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(g.config.IntervalSec) * time.Second
}

const maxSubPages = 10

func (g *GovSiteCrawler) Fetch() ([]*CrawlResult, error) {
	var mainResult *CrawlResult
	var err error
	if g.renderer != nil {
		mainResult, err = g.fetchWithRenderer()
	} else {
		mainResult, err = g.fetchHTTP()
	}
	if err != nil {
		return nil, err
	}

	var results []*CrawlResult
	results = append(results, mainResult)

	links := extractPolicyLinks(mainResult.RawText, g.config.SourceURL)
	for i, link := range links {
		if i >= maxSubPages {
			break
		}
		sub, err := g.fetchURL(link)
		if err != nil {
			log.Printf("[crawler] sub-page %s fetch error: %v", link, err)
			continue
		}
		if sub != nil {
			results = append(results, sub)
		}
	}

	if len(links) > 0 {
		log.Printf("[crawler] %s: fetched main + %d sub-pages", g.config.SourceID, len(results)-1)
	}
	return results, nil
}

func (g *GovSiteCrawler) fetchWithRenderer() (*CrawlResult, error) {
	html, err := g.renderer.Render(g.config.SourceURL)
	if err != nil {
		log.Printf("[crawler] renderer failed for %s: %v, falling back to HTTP", g.config.SourceID, err)
		return g.fetchHTTP()
	}
	hash := sha256.Sum256([]byte(html))
	return &CrawlResult{
		SourceID:    g.config.SourceID,
		SourceLevel: g.config.SourceLevel,
		RawText:     html,
		Title:       g.config.SourceName,
		SourceURL:   g.config.SourceURL,
		FetchedAt:   time.Now(),
		VersionHash: fmt.Sprintf("%x", hash),
	}, nil
}

func (g *GovSiteCrawler) fetchHTTP() (*CrawlResult, error) {
	var lastErr error
	for _, ua := range govUserAgents {
		result, err := g.tryFetch(ua)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("all user-agents failed: %w", lastErr)
}

func (g *GovSiteCrawler) fetchURL(rawURL string) (*CrawlResult, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", govUserAgents[0])
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.5")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if len(body) < 100 {
		return nil, fmt.Errorf("response too short (%d bytes)", len(body))
	}

	finalURL := rawURL
	if resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	hash := sha256.Sum256(body)
	sourceID := g.config.SourceID
	return &CrawlResult{
		SourceID:    sourceID,
		SourceLevel: g.config.SourceLevel,
		RawText:     string(body),
		Title:       g.config.SourceName,
		SourceURL:   finalURL,
		FetchedAt:   time.Now(),
		VersionHash: fmt.Sprintf("%x", hash),
	}, nil
}

func (g *GovSiteCrawler) tryFetch(userAgent string) (*CrawlResult, error) {
	req, err := http.NewRequest("GET", g.config.SourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if len(body) < 100 {
		return nil, fmt.Errorf("response too short (%d bytes)", len(body))
	}

	finalURL := g.config.SourceURL
	if resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	hash := sha256.Sum256(body)

	return &CrawlResult{
		SourceID:    g.config.SourceID,
		SourceLevel: g.config.SourceLevel,
		RawText:     string(body),
		Title:       g.config.SourceName,
		SourceURL:   finalURL,
		FetchedAt:   time.Now(),
		VersionHash: fmt.Sprintf("%x", hash),
	}, nil
}

func extractPolicyLinks(html, baseURL string) []string {
	seen := make(map[string]bool)
	var links []string

	lower := strings.ToLower(html)
	idx := 0
	for {
		start := strings.Index(lower[idx:], "<a ")
		if start == -1 {
			break
		}
		start += idx
		end := strings.Index(lower[start:], "</a>")
		if end == -1 {
			break
		}
		end += start

		tag := html[start : end+4]

		hrefStart := strings.Index(strings.ToLower(tag), "href=\"")
		if hrefStart == -1 {
			hrefStart = strings.Index(strings.ToLower(tag), "href='")
		}
		if hrefStart == -1 {
			idx = end + 4
			continue
		}
		quote := string(tag[hrefStart+5])
		contentStart := hrefStart + 6
		hrefEnd := strings.Index(tag[contentStart:], quote)
		if hrefEnd == -1 {
			idx = end + 4
			continue
		}
		href := tag[contentStart : contentStart+hrefEnd]
		idx = end + 4

		if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
			continue
		}
		if strings.HasPrefix(href, "//") {
			href = "https:" + href
		}
		if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
			base, err := url.Parse(baseURL)
			if err != nil {
				continue
			}
			rel, err := url.Parse(href)
			if err != nil {
				continue
			}
			href = base.ResolveReference(rel).String()
		}

		if seen[href] {
			continue
		}
		seen[href] = true

		text := extractLinkText(tag)
		if !isPolicyLink(href, text) {
			continue
		}
		links = append(links, href)
	}

	return links
}

func extractLinkText(tag string) string {
	start := strings.Index(tag, ">")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(tag, "</a>")
	if end == -1 || end <= start {
		return ""
	}
	text := tag[start+1 : end]
	text = strings.Join(strings.Fields(text), " ")
	return text
}

func isPolicyLink(href, text string) bool {
	keywords := []string{
		"政策", "通知", "公告", "办法", "规定", "方案", "意见",
		"2025", "2026", "2027", "2028",
		"社保", "养老", "医疗", "失业", "工伤", "生育",
		"补贴", "就业", "公积金",
	}
	combined := strings.ToLower(href + " " + text)
	for _, kw := range keywords {
		if strings.Contains(combined, strings.ToLower(kw)) {
			return true
		}
	}
	extensions := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx"}
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(href), ext) {
			return false
		}
	}
	return false
}
