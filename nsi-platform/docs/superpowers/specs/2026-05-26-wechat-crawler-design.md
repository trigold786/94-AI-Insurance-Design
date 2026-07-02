# WeChat Official Account Crawler Design

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

## Overview

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

Crawl social insurance policy articles from WeChat official accounts (鍏紬鍙?. Uses Bing search for article discovery with Sogou fallback and manual URL as last resort.

## Architecture

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

```
Admin creates wechat-type source
  鈹溾攢 Mode 1: Public account name or keyword
  鈹?   鈫?Bing search: site:mp.weixin.qq.com + keyword
  鈹?   鈫?Extract article URLs from search results
  鈹?   鈫?(Fallback) Sogou WeChat search if Bing blocked
  鈹?   鈫?Render each article page 鈫?extract content
  鈹?  鈹斺攢 Mode 2: Direct mp.weixin.qq.com URL
       鈫?Chrome render single article 鈫?extract content
```

## URL Format Detection

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

SourceURL field supports 3 formats:
- **Account name/keyword**: `涓婃捣绀句繚` or `keyword:绀句繚琛ョ即鏀跨瓥` 鈫?triggers Bing search discovery
- **Direct URL**: `mp.weixin.qq.com/s/...` or `https://mp.weixin.qq.com/s/...` 鈫?direct render
- **Multiple URLs**: newline-separated list (same as Douyin)

## Discovery Flow

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Step 1: Bing Search (primary)
```
GET https://cn.bing.com/search?q=site%3Amp.weixin.qq.com+{keyword}
Chrome RenderWithVirtualTime(url, 15000)
Parse search result HTML 鈫?extract mp.weixin.qq.com links
Limit: top 20 articles per crawl
```

### Step 2: Sogou Fallback
If Bing returns no results or is blocked:
```
GET https://weixin.sogou.com/weixin?type=2&query={keyword}
Chrome RenderWithVirtualTime(url, 15000)
Parse search result HTML 鈫?extract mp.weixin.qq.com links
```

### Step 3: Manual URL (always available)
Admin can paste direct article URLs as SourceURL, newline-separated.

## Content Extraction

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

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

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- Reuse existing `chromeMu` mutex (serial Chrome rendering)
- 3-5 second delay between article fetches
- CAPTCHA detection: if page contains verification form, log and skip
- Max 20 articles per crawl cycle
- Default crawl interval: 168 hours (7 days)

## Data Model

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Source Configuration
```sql
INSERT INTO policy_sources (source_id, source_name, source_url, source_level, crawl_type, interval_sec, region_code, enabled)
VALUES ('WECHAT-涓婃捣绀句繚', '寰俊鍏紬鍙?涓婃捣绀句繚', '涓婃捣绀句繚', 'MEDIUM', 'wechat', 604800, '310000', true);
```

### CrawlType Registration
`manager.go` switch adds `case "wechat"` 鈫?`NewWeChatCrawler(cfg).SetRenderer(renderer)`

## File Changes

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

| File | Change |
|------|--------|
| `internal/crawler/wechat_crawler.go` | New file, ~300 lines |
| `internal/crawler/manager.go` | Add `wechat` case in both `Init()` and `loadAndRegisterSource()` |
| `migrations/017_wechat_source.sql` | Seed data for initial WeChat sources |

## WeChatCrawler Interface

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

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

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- `discoverArticles(keyword)` 鈥?Bing search 鈫?extract article URLs
- `discoverArticlesSogou(keyword)` 鈥?Sogou fallback
- `fetchArticle(url)` 鈥?Chrome render single article 鈫?extract title + content
- `extractWeChatContent(html)` 鈥?Parse rendered HTML for title, content, author, date
- `isWeChatArticleURL(url)` 鈥?Check if URL is a WeChat article
- `parseWeChatURLs(rawURL)` 鈥?Parse SourceURL field (handle multiple formats)
