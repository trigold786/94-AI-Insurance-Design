# Sub-Project 1: Crawler Capability Enhancement Design

| **°æ±¾ºÅ** | V1.0.0 |
| **×´Ì¬** | ÒÑÉúÐ§ |
| **·¢²¼ÈÕÆÚ** | 2026-06-15 |

**Date**: 2026-05-27
**Scope**: Policy-crawler service â€?crawling infrastructure improvements
**Order**: First of 3 sub-projects (1: Crawler, 2: Extraction, 3: Analytics)

---

## 1. Document File Parsing

| **°æ±¾ºÅ** | V1.0.0 |
| **×´Ì¬** | ÒÑÉúÐ§ |
| **·¢²¼ÈÕÆÚ** | 2026-06-15 |

### Problem
GovSite crawler explicitly rejects `.pdf/.doc/.docx/.xls/.xlsx` links via `isPolicyLink()`. Government policy documents are frequently published as PDF files.

### Solution
- New file `internal/crawler/doc_parser.go` with a `DocParser` interface
- For PDF: use `github.com/ledongthuc/pdf` Go library (pure Go, no CGO required)
- For DOCX: use `github.com/nguyenthenguyen/docx` or extract from ZIP + read `word/document.xml`
- GovSite `isPolicyLink()` updated to allow document links
- GovSite `fetchURL()` detects content-type, routes PDF/DOCX to parser, HTML stays as-is

### Architecture
```
fetchURL() â†?HTTP GET â†?check Content-Type header
  â”œâ”€â”€ text/html â†?existing HTML processing
  â”œâ”€â”€ application/pdf â†?DocParser.ParsePDF(responseBody) â†?plain text
  â”œâ”€â”€ application/...word... â†?DocParser.ParseDOCX(responseBody) â†?plain text
  â””â”€â”€ other â†?skip (log warning)
```

### Files
- Create: `internal/crawler/doc_parser.go`
- Modify: `internal/crawler/govsite.go` â€?`isPolicyLink()`, `fetchURL()`
- Add Go dependency: `github.com/ledongthuc/pdf`

---

## 2. Robust Retry + Proxy Support

| **°æ±¾ºÅ** | V1.0.0 |
| **×´Ì¬** | ÒÑÉúÐ§ |
| **·¢²¼ÈÕÆÚ** | 2026-06-15 |

### Problem
Current HTTP client has no retry logic. Transient errors (timeouts, 5xx, DNS blips) cause permanent crawl failures. No proxy support for geo-restricted government sites.

### Solution
- New file `internal/crawler/http_client.go` with `NewResilientHTTPClient(config)` 
- Exponential backoff: 3 retries, delays 1s â†?4s â†?16s
- Error classification for retry decision:
  - **Retry**: 5xx, network timeout, DNS temporary failure, connection refused
  - **No retry**: 4xx (client errors), DNS not found (permanent)
- Proxy: read `proxy_url` from source config, set on HTTP transport
- Configurable timeout (default 30s, override via source config)

### Architecture
```go
type HTTPClientConfig struct {
    MaxRetries    int
    BaseDelay     time.Duration
    Timeout       time.Duration
    ProxyURL      string
    UserAgent     string
}

func (c *ResilientHTTPClient) Get(url string) (*http.Response, error)
```

### Files
- Create: `internal/crawler/http_client.go`
- Modify: `internal/crawler/govsite.go` â€?use new client
- Modify: `internal/crawler/rss_crawler.go` â€?use new client
- Modify: `internal/crawler/wechat_crawler.go` â€?use new client
- Migration: `022_crawler_ethics.sql` adds `proxy_url TEXT DEFAULT ''` to `policy_sources`

---

## 3. Crawling Ethics Configuration

| **°æ±¾ºÅ** | V1.0.0 |
| **×´Ì¬** | ÒÑÉúÐ§ |
| **·¢²¼ÈÕÆÚ** | 2026-06-15 |

### Problem
No rate limiting between requests. No robots.txt checking. Crawlers could overwhelm government servers or get IP-banned.

### Solution
- New columns on `policy_sources`:
  - `request_delay_ms INT DEFAULT 1000` â€?delay between requests in milliseconds
  - `max_concurrent INT DEFAULT 1` â€?max concurrent requests per source
  - `respect_robots BOOLEAN DEFAULT true` â€?check robots.txt before crawling
- New file `internal/crawler/robots.go` â€?fetch and parse robots.txt, check Allow/Disallow rules
- Crawler manager enforces `request_delay_ms` between fetches via `time.Sleep`
- Source management admin UI gets new configuration fields

### Architecture
```
Manager.crawlAndProcess(source)
  â†?read source.RequestDelayMs, source.RespectRobots, source.ProxyURL
  â†?if RespectRobots: robots.Check(sourceURL, userAgent) â†?skip if disallowed
  â†?fetch with delay between requests
  â†?use ResilientHTTPClient with proxy
```

### Files
- Create: `internal/crawler/robots.go`
- Modify: `internal/crawler/manager.go` â€?read ethics config, enforce delays
- Modify: `internal/admin/admin_sources.go` â€?new fields in source update/create
- Modify: `internal/admin/admin_page.go` â€?source editor UI fields
- Migration: `022_crawler_ethics.sql`

---

## 4. Migration: 022_crawler_ethics.sql

| **°æ±¾ºÅ** | V1.0.0 |
| **×´Ì¬** | ÒÑÉúÐ§ |
| **·¢²¼ÈÕÆÚ** | 2026-06-15 |

```sql
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS proxy_url TEXT DEFAULT '';
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS request_delay_ms INT DEFAULT 1000;
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS max_concurrent INT DEFAULT 1;
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS respect_robots BOOLEAN DEFAULT true;
```

---

## Modified Files Summary

| **°æ±¾ºÅ** | V1.0.0 |
| **×´Ì¬** | ÒÑÉúÐ§ |
| **·¢²¼ÈÕÆÚ** | 2026-06-15 |

| File | Change |
|------|--------|
| `internal/crawler/doc_parser.go` | NEW â€?PDF/DOCX text extraction |
| `internal/crawler/http_client.go` | NEW â€?resilient HTTP client with retry + proxy |
| `internal/crawler/robots.go` | NEW â€?robots.txt parser and checker |
| `internal/crawler/govsite.go` | Use new HTTP client, allow document links, route to doc parser |
| `internal/crawler/rss_crawler.go` | Use new HTTP client |
| `internal/crawler/wechat_crawler.go` | Use new HTTP client |
| `internal/crawler/manager.go` | Read ethics config, enforce delays, check robots |
| `internal/admin/admin_sources.go` | New source config fields |
| `internal/admin/admin_page.go` | Source editor UI for new fields |
| `migrations/022_crawler_ethics.sql` | New columns on policy_sources |

## No Changes To
- Actuarial engine
- API server
- Shared models
- Extraction pipeline (sub-project 2)
- Admin dashboard analytics (sub-project 3)
- Existing crawler types that already work (douyin, file, manual)
