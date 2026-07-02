# Crawler Capability Enhancement Implementation Plan

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enhance the policy-crawler with document file parsing, resilient HTTP client (retry + proxy), and crawling ethics configuration.

**Architecture:** New `http_client.go` wraps all HTTP requests with retry/proxy. New `doc_parser.go` extracts text from PDF/DOCX. New `robots.go` checks robots.txt. SourceConfig gets 4 new fields via migration.

**Tech Stack:** Go, `github.com/ledongthuc/pdf` (pure Go PDF), existing PostgreSQL + Docker infrastructure

---

## File Structure

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

| File | Responsibility |
|------|---------------|
| `migrations/022_crawler_ethics.sql` | Add proxy_url, request_delay_ms, max_concurrent, respect_robots columns |
| `internal/crawler/crawler.go` | Add 4 new fields to SourceConfig struct |
| `internal/crawler/store.go` | Update SQL queries to read new fields |
| `internal/crawler/http_client.go` | NEW 鈥?ResilientHTTPClient with retry + proxy |
| `internal/crawler/doc_parser.go` | NEW 鈥?PDF/DOCX text extraction |
| `internal/crawler/robots.go` | NEW 鈥?robots.txt parser and checker |
| `internal/crawler/govsite.go` | Use new HTTP client, allow document links, route to doc parser, check robots |
| `internal/crawler/rss_crawler.go` | Use new HTTP client |
| `internal/crawler/manager.go` | Enforce request delay between fetches |
| `internal/admin/admin_sources.go` | New source config fields in create/update |
| `internal/admin/admin_page.go` | Source editor UI for new fields |

---

### Task 1: Migration + SourceConfig struct update

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `migrations/022_crawler_ethics.sql`
- Modify: `internal/crawler/crawler.go`
- Modify: `internal/crawler/store.go`

- [ ] **Step 1: Create migration file**

Create `migrations/022_crawler_ethics.sql`:

```sql
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS proxy_url TEXT DEFAULT '';
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS request_delay_ms INT DEFAULT 1000;
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS max_concurrent INT DEFAULT 1;
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS respect_robots BOOLEAN DEFAULT true;
```

- [ ] **Step 2: Add fields to SourceConfig struct**

In `internal/crawler/crawler.go`, update the `SourceConfig` struct (add after `Enabled` field):

```go
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
```

- [ ] **Step 3: Update ListEnabledSources SQL query**

In `internal/crawler/store.go`, update `ListEnabledSources()` (around line 33). Change the SQL query and Scan call:

```go
func (s *DBStore) ListEnabledSources() ([]SourceConfig, error) {
	rows, err := s.db.Query(`SELECT source_id, source_name, source_url, source_level, crawl_type, interval_sec, COALESCE(region_code,''), enabled, COALESCE(proxy_url,''), COALESCE(request_delay_ms,1000), COALESCE(max_concurrent,1), COALESCE(respect_robots,true) FROM policy_sources WHERE enabled = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	var cfgs []SourceConfig
	for rows.Next() {
		var c SourceConfig
		if err := rows.Scan(&c.SourceID, &c.SourceName, &c.SourceURL, &c.SourceLevel, &c.CrawlType, &c.IntervalSec, &c.RegionCode, &c.Enabled, &c.ProxyURL, &c.RequestDelayMs, &c.MaxConcurrent, &c.RespectRobots); err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		cfgs = append(cfgs, c)
	}
	return cfgs, rows.Err()
}
```

- [ ] **Step 4: Run migration and verify compilation**

Run migration inside Docker: `docker exec nsi-policy-crawler /app/scripts/migrate.sh`
Run: `go build ./...` (in `services/policy-crawler`)
Expected: compiles without errors

---

### Task 2: Resilient HTTP Client

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `internal/crawler/http_client.go`

- [ ] **Step 1: Create http_client.go**

Create `internal/crawler/http_client.go`:

```go
package crawler

import (
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClientConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	Timeout    time.Duration
	ProxyURL   string
	UserAgent  string
}

func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		Timeout:    30 * time.Second,
		UserAgent:  govUserAgents[0],
	}
}

type ResilientHTTPClient struct {
	config HTTPClientConfig
	client *http.Client
}

func NewResilientHTTPClient(cfg HTTPClientConfig) *ResilientHTTPClient {
	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	if cfg.ProxyURL != "" {
		if proxyURL, err := url.Parse(cfg.ProxyURL); err == nil {
			tr.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &ResilientHTTPClient{
		config: cfg,
		client: &http.Client{
			Timeout:   cfg.Timeout,
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

func (c *ResilientHTTPClient) Get(urlStr string) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * c.config.BaseDelay
			time.Sleep(delay)
		}
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", c.config.UserAgent)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryableError(err) {
				return nil, err
			}
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, urlStr)
			resp.Body.Close()
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("after %d retries: %w", c.config.MaxRetries, lastErr)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline") {
		return true
	}
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "EOF") {
		return true
	}
	if strings.Contains(msg, "temporary") || strings.Contains(msg, "TLS handshake") {
		return true
	}
	if strings.Contains(msg, "no such host") {
		return false
	}
	return true
}

func (c *ResilientHTTPClient) HTTPClient() *http.Client {
	return c.client
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...` (in `services/policy-crawler`)

---

### Task 3: Document File Parser

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `internal/crawler/doc_parser.go`

- [ ] **Step 1: Add pdf dependency**

Run: `go get github.com/ledongthuc/pdf` (in `services/policy-crawler`)

- [ ] **Step 2: Create doc_parser.go**

Create `internal/crawler/doc_parser.go`:

```go
package crawler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ExtractPDFText(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("pdf open: %w", err)
	}
	var buf strings.Builder
	n := reader.NumPage()
	for i := 1; i <= n; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}
	result := strings.TrimSpace(buf.String())
	if len(result) == 0 {
		return "", fmt.Errorf("pdf: no text extracted from %d pages", n)
	}
	return result, nil
}

func ExtractDOCXText(data []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("docx open: %w", err)
	}
	var buf strings.Builder
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("docx read document.xml: %w", err)
			}
			defer rc.Close()
			content, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("docx read content: %w", err)
			}
			text := stripXMLTags(string(content))
			text = strings.TrimSpace(text)
			if len(text) > 0 {
				buf.WriteString(text)
			}
		}
	}
	result := buf.String()
	if len(result) == 0 {
		return "", fmt.Errorf("docx: no text extracted")
	}
	return result, nil
}

func stripXMLTags(s string) string {
	var buf strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			buf.WriteString(" ")
			continue
		}
		if !inTag {
			buf.WriteRune(r)
		}
	}
	result := buf.String()
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	for strings.Contains(result, "\n ") {
		result = strings.ReplaceAll(result, "\n ", "\n")
	}
	result = strings.ReplaceAll(result, "\t", " ")
	return result
}

func IsPDFContentType(ct string) bool {
	return strings.Contains(ct, "application/pdf")
}

func IsDOCXContentType(ct string) bool {
	return strings.Contains(ct, "officedocument.wordprocessingml") ||
		strings.Contains(ct, "msword") ||
		strings.Contains(ct, "application/doc")
}

func IsDocumentContentType(ct string) bool {
	return IsPDFContentType(ct) || IsDOCXContentType(ct)
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`

---

### Task 4: Robots.txt Parser

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `internal/crawler/robots.go`

- [ ] **Step 1: Create robots.go**

Create `internal/crawler/robots.go`:

```go
package crawler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type RobotsChecker struct {
	mu     sync.RWMutex
	cache  map[string]*robotsRule
	client *http.Client
}

type robotsRule struct {
	allowed    map[string]bool
	fetched    time.Time
	disallowAll bool
}

func NewRobotsChecker() *RobotsChecker {
	return &RobotsChecker{
		cache: make(map[string]*robotsRule),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (rc *RobotsChecker) IsAllowed(targetURL, userAgent string) bool {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return true
	}
	rule := rc.getRule(parsed.Scheme + "://" + parsed.Host)
	if rule == nil {
		return true
	}
	if rule.disallowAll {
		return false
	}
	return rule.isAllowed(parsed.Path, userAgent)
}

func (rc *RobotsChecker) getRule(baseURL string) *robotsRule {
	rc.mu.RLock()
	r, ok := rc.cache[baseURL]
	rc.mu.RUnlock()
	if ok && time.Since(r.fetched) < 6*time.Hour {
		return r
	}
	r = rc.fetchRule(baseURL)
	rc.mu.Lock()
	rc.cache[baseURL] = r
	rc.mu.Unlock()
	return r
}

func (rc *RobotsChecker) fetchRule(baseURL string) *robotsRule {
	resp, err := rc.client.Get(baseURL + "/robots.txt")
	if err != nil {
		return &robotsRule{allowed: map[string]bool{}, fetched: time.Now()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return &robotsRule{allowed: map[string]bool{}, fetched: time.Now()}
	}
	var buf strings.Builder
	buf.ReadFrom(resp.Body)
	return parseRobotsTxt(buf.String())
}

func parseRobotsTxt(content string) *robotsRule {
	rule := &robotsRule{
		allowed: map[string]bool{},
		fetched: time.Now(),
	}
	var currentUA string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.SplitN(line, "#", 2)[0]
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "user-agent:") {
			currentUA = strings.TrimSpace(line[len("user-agent:"):])
			continue
		}
		if currentUA != "*" && !strings.EqualFold(currentUA, "Mozilla") {
			continue
		}
		if strings.HasPrefix(lower, "disallow:") {
			path := strings.TrimSpace(line[len("disallow:"):])
			if path == "/" {
				rule.disallowAll = true
				return rule
			}
			if path != "" {
				rule.allowed[path] = false
			}
		}
		if strings.HasPrefix(lower, "allow:") {
			path := strings.TrimSpace(line[len("allow:"):])
			if path != "" {
				rule.allowed[path] = true
			}
		}
	}
	return rule
}

func (r *robotsRule) isAllowed(path, userAgent string) bool {
	longestMatch := ""
	allowed := true
	for pattern, isAllowed := range r.allowed {
		if strings.HasPrefix(path, pattern) && len(pattern) > len(longestMatch) {
			longestMatch = pattern
			allowed = isAllowed
		}
	}
	return allowed
}

func CheckRobotsBeforeCrawl(checker *RobotsChecker, targetURL, userAgent string) error {
	if checker == nil {
		return nil
	}
	if !checker.IsAllowed(targetURL, userAgent) {
		return fmt.Errorf("robots.txt disallows crawling: %s", targetURL)
	}
	return nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`

---

### Task 5: Integrate into GovSiteCrawler

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `internal/crawler/govsite.go`

- [ ] **Step 1: Replace HTTP client creation in NewGovSiteCrawler**

Replace the `NewGovSiteCrawler` function:

```go
func NewGovSiteCrawler(cfg SourceConfig) *GovSiteCrawler {
	httpCfg := HTTPClientConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		Timeout:    30 * time.Second,
		ProxyURL:   cfg.ProxyURL,
		UserAgent:  govUserAgents[0],
	}
	rc := NewResilientHTTPClient(httpCfg)
	return &GovSiteCrawler{
		config:   cfg,
		client:   rc.HTTPClient(),
		rc:       rc,
		renderer: nil,
	}
}
```

- [ ] **Step 2: Add rc field to GovSiteCrawler struct**

Replace the struct definition:

```go
type GovSiteCrawler struct {
	config   SourceConfig
	client   *http.Client
	rc       *ResilientHTTPClient
	renderer PageRenderer
}
```

- [ ] **Step 3: Allow document links in isPolicyLink**

Find the `isPolicyLink` function. It currently has a block that rejects `.pdf/.doc/.docx/.xls/.xlsx`. Change it to **allow** those extensions (remove the exclusion block for document files). The relevant code to find and remove/modify is the section checking `hasExt(link, ".pdf")` etc. 鈥?replace the rejection with a simple `return true` for document links:

Find the pattern like:
```go
if hasExt(link, ".pdf") || hasExt(link, ".doc") || ...
```

Replace the rejection with:
```go
	if hasExt(link, ".pdf") || hasExt(link, ".doc") || hasExt(link, ".docx") {
		return true
	}
	if hasExt(link, ".xls") || hasExt(link, ".xlsx") || hasExt(link, ".ppt") || hasExt(link, ".pptx") {
		return false
	}
```

- [ ] **Step 4: Add document content-type handling in fetchURL**

In the `fetchURL` function, after the HTTP GET succeeds and before the HTML processing, add content-type detection. Find where `resp.Body` is read and add after the read:

```go
func (g *GovSiteCrawler) fetchURL(rawURL string) (string, string, error) {
	resp, err := g.rc.Get(rawURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if IsPDFContentType(ct) {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", "", fmt.Errorf("read pdf: %w", err)
		}
		text, err := ExtractPDFText(data)
		if err != nil {
			return "", "", fmt.Errorf("parse pdf: %w", err)
		}
		title := extractTitleFromURL(rawURL)
		return text, title, nil
	}
	if IsDOCXContentType(ct) {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", "", fmt.Errorf("read docx: %w", err)
		}
		text, err := ExtractDOCXText(data)
		if err != nil {
			return "", "", fmt.Errorf("parse docx: %w", err)
		}
		title := extractTitleFromURL(rawURL)
		return text, title, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	...
```

Add a helper function:

```go
func extractTitleFromURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	parts := strings.Split(parsed.Path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.TrimSpace(parts[i])
		if p != "" && !strings.HasPrefix(p, ".") {
			return p
		}
	}
	return parsed.Path
}
```

- [ ] **Step 5: Add robots.txt check in Fetch()**

At the beginning of the `Fetch()` method, add robots check:

```go
func (g *GovSiteCrawler) Fetch() ([]*CrawlResult, error) {
	if g.config.RespectRobots {
		checker := NewRobotsChecker()
		if err := CheckRobotsBeforeCrawl(checker, g.config.SourceURL, govUserAgents[0]); err != nil {
			log.Printf("[govsite] %s: %v", g.config.SourceID, err)
			return nil, err
		}
	}
	... rest of existing Fetch() code
```

- [ ] **Step 6: Verify compilation**

Run: `go build ./...`

---

### Task 6: Integrate into RSS and WeChat crawlers

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `internal/crawler/rss_crawler.go`
- Modify: `internal/crawler/wechat_crawler.go`

- [ ] **Step 1: Update RSS crawler to use ResilientHTTPClient**

In `rss_crawler.go`, find where the `*http.Client` is created in `NewRSSCrawler` or `Fetch`. Replace with `ResilientHTTPClient`:

```go
type RSSCrawler struct {
	config SourceConfig
	rc     *ResilientHTTPClient
}

func NewRSSCrawler(cfg SourceConfig) *RSSCrawler {
	httpCfg := DefaultHTTPClientConfig()
	httpCfg.ProxyURL = cfg.ProxyURL
	rc := NewResilientHTTPClient(httpCfg)
	return &RSSCrawler{config: cfg, rc: rc}
}
```

Update the `Fetch()` method to use `g.rc.Get()` instead of `g.client.Get()` or raw `http.Get`.

- [ ] **Step 2: Update WeChat crawler similarly**

Same pattern in `wechat_crawler.go` 鈥?replace HTTP client creation with `ResilientHTTPClient`.

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`

---

### Task 7: Manager 鈥?enforce request delay

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `internal/crawler/manager.go`

- [ ] **Step 1: Add delay enforcement in crawlAndProcess**

In `crawlAndProcess()`, after getting the source config and before calling `s.Fetch()`, add:

```go
func (m *CrawlerManager) crawlAndProcess(s Source) {
	sourceID := s.SourceID()
	...
	if delay := getRequestDelay(s); delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	...
```

Add helper to read delay from config. Since `Source` interface doesn't expose config, add a method:

```go
type DelayConfigurable interface {
	RequestDelayMs() int
}
```

Then in the crawler types, add:

```go
func (g *GovSiteCrawler) RequestDelayMs() int { return g.config.RequestDelayMs }
func (r *RSSCrawler) RequestDelayMs() int     { return r.config.RequestDelayMs }
```

And in manager:

```go
func getRequestDelay(s Source) int {
	if d, ok := s.(DelayConfigurable); ok {
		return d.RequestDelayMs()
	}
	return 1000
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`

---

### Task 8: Admin UI 鈥?new source config fields

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `internal/admin/admin_sources.go`
- Modify: `internal/admin/admin_page.go`

- [ ] **Step 1: Update admin source create/update handlers**

In `admin_sources.go`, update the SQL INSERT/UPDATE statements to include `proxy_url, request_delay_ms, max_concurrent, respect_robots`. Also update the SELECT queries and Scan calls in source listing.

Find the existing INSERT and UPDATE SQL for `policy_sources` and add the 4 new columns. Find the existing SELECT for listing sources and add the new columns to the query and Scan.

- [ ] **Step 2: Update admin page source editor**

In `admin_page.go`, find the source editor form (HTML template). Add input fields for the 4 new configuration options:

```html
<div class="form-row">
  <div><label>浠ｇ悊 URL</label><input id="src-proxy" placeholder="http://proxy:port"></div>
  <div><label>璇锋眰闂撮殧(ms)</label><input id="src-delay" type="number" value="1000" min="0"></div>
  <div><label>鏈�澶у苟鍙?/label><input id="src-concurrent" type="number" value="1" min="1" max="10"></div>
  <div><label>閬靛畧Robots.txt</label><select id="src-robots"><option value="true">鏄?/option><option value="false">鍚?/option></select></div>
</div>
```

Update the JavaScript that reads form values and sends them to the API.

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`

---

### Task 9: Build, deploy, and verify

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- [ ] **Step 1: Run migration**

Run: `docker exec nsi-policy-crawler /app/scripts/migrate.sh`

- [ ] **Step 2: Cross-compile**

Run: `$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o policy-crawler ./cmd/` (in `services/policy-crawler`)

- [ ] **Step 3: Deploy**

Run:
```bash
docker cp services/policy-crawler/policy-crawler nsi-policy-crawler:/policy-crawler
docker restart nsi-policy-crawler
```

- [ ] **Step 4: Verify health**

Run: `Start-Sleep -Seconds 3; curl.exe -s -u admin:nsi_admin_2026 http://localhost:39403/admin/dashboard`
Expected: JSON with dashboard stats

- [ ] **Step 5: Verify migration applied**

Run: `docker exec nsi-policy-crawler /app/scripts/migrate.sh`
Expected: migration 022 logged as already applied

- [ ] **Step 6: Test source update with new fields**

Run:
```bash
curl -s -u admin:nsi_admin_2026 -X POST http://localhost:39403/admin/sources/update -H "Content-Type: application/json" -d '{"source_id":"SH-GOV-HRSS","proxy_url":"","request_delay_ms":2000,"max_concurrent":1,"respect_robots":true}'
```

Expected: success response

- [ ] **Step 7: Commit**

```bash
git add services/policy-crawler/
git commit -m "feat(crawler): add document parsing, resilient HTTP client, crawling ethics config"
```

---

## Self-Review

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- [x] Spec coverage: Task 1 = migration + config, Task 2 = HTTP client, Task 3 = doc parser, Task 4 = robots, Task 5 = govsite integration, Task 6 = rss/wechat integration, Task 7 = manager delay, Task 8 = admin UI, Task 9 = deploy
- [x] No placeholders: all code shown in full
- [x] Type consistency: `SourceConfig` fields match across crawler.go, store.go, admin_sources.go; `ResilientHTTPClient` used consistently in govsite, rss, wechat
