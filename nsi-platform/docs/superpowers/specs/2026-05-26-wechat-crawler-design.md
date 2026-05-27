# WeChat Official Account Crawler Design

## Overview

Crawl social insurance policy articles from WeChat official accounts (公众号). Uses Bing search for article discovery with Sogou fallback and manual URL as last resort.

## Architecture

```
Admin creates wechat-type source
  ├─ Mode 1: Public account name or keyword
  │    → Bing search: site:mp.weixin.qq.com + keyword
  │    → Extract article URLs from search results
  │    → (Fallback) Sogou WeChat search if Bing blocked
  │    → Render each article page → extract content
  │
  └─ Mode 2: Direct mp.weixin.qq.com URL
       → Chrome render single article → extract content
```

## URL Format Detection

SourceURL field supports 3 formats:
- **Account name/keyword**: `上海社保` or `keyword:社保补缴政策` → triggers Bing search discovery
- **Direct URL**: `mp.weixin.qq.com/s/...` or `https://mp.weixin.qq.com/s/...` → direct render
- **Multiple URLs**: newline-separated list (same as Douyin)

## Discovery Flow

### Step 1: Bing Search (primary)
```
GET https://cn.bing.com/search?q=site%3Amp.weixin.qq.com+{keyword}
Chrome RenderWithVirtualTime(url, 15000)
Parse search result HTML → extract mp.weixin.qq.com links
Limit: top 20 articles per crawl
```

### Step 2: Sogou Fallback
If Bing returns no results or is blocked:
```
GET https://weixin.sogou.com/weixin?type=2&query={keyword}
Chrome RenderWithVirtualTime(url, 15000)
Parse search result HTML → extract mp.weixin.qq.com links
```

### Step 3: Manual URL (always available)
Admin can paste direct article URLs as SourceURL, newline-separated.

## Content Extraction

### WeChat Article Page (`mp.weixin.qq.com/s/...`)
- **Title**: `<h1 class="rich_media_title">` or `<meta property="og:title">`
- **Content**: `#js_content` div (main article body)
- **Author**: `<span class="rich_media_meta_nickname">` or meta tag
- **Date**: `#publish_time` element
- **Fallback**: strip HTML tags if structured extraction fails

### Bing Search Results Page
- Extract all `<a href>` containing `mp.weixin.qq.com/s/`
- Deduplicate by URL

## Anti-Scraping Strategy

- Reuse existing `chromeMu` mutex (serial Chrome rendering)
- 3-5 second delay between article fetches
- CAPTCHA detection: if page contains verification form, log and skip
- Max 20 articles per crawl cycle
- Default crawl interval: 168 hours (7 days)

## Data Model

### Source Configuration
```sql
INSERT INTO policy_sources (source_id, source_name, source_url, source_level, crawl_type, interval_sec, region_code, enabled)
VALUES ('WECHAT-上海社保', '微信公众号-上海社保', '上海社保', 'MEDIUM', 'wechat', 604800, '310000', true);
```

### CrawlType Registration
`manager.go` switch adds `case "wechat"` → `NewWeChatCrawler(cfg).SetRenderer(renderer)`

## File Changes

| File | Change |
|------|--------|
| `internal/crawler/wechat_crawler.go` | New file, ~300 lines |
| `internal/crawler/manager.go` | Add `wechat` case in both `Init()` and `loadAndRegisterSource()` |
| `migrations/017_wechat_source.sql` | Seed data for initial WeChat sources |

## WeChatCrawler Interface

```go
type WeChatCrawler struct {
    config    SourceConfig
    renderer  PageRenderer
    maxItems  int
    processed map[string]bool
}

func NewWeChatCrawler(cfg SourceConfig) *WeChatCrawler
func (w *WeChatCrawler) SetRenderer(r PageRenderer)
func (w *WeChatCrawler) SourceID() string
func (w *WeChatCrawler) SourceLevel() string
func (w *WeChatCrawler) Interval() time.Duration
func (w *WeChatCrawler) Fetch() ([]*CrawlResult, error)
```

## Key Methods

- `discoverArticles(keyword)` — Bing search → extract article URLs
- `discoverArticlesSogou(keyword)` — Sogou fallback
- `fetchArticle(url)` — Chrome render single article → extract title + content
- `extractWeChatContent(html)` — Parse rendered HTML for title, content, author, date
- `isWeChatArticleURL(url)` — Check if URL is a WeChat article
- `parseWeChatURLs(rawURL)` — Parse SourceURL field (handle multiple formats)
