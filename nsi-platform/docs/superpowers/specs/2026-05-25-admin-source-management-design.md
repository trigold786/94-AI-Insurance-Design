# Admin UI: Multi-Level Source Management

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

## Goal

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

Enable full CRUD + operational management of all policy source types (govsite, file, rss, manual) in the existing admin panel's "鏁版嵁婧愮鐞? tab.

## Current State

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

The admin panel (`/admin`) has a "鏁版嵁婧愮鐞? tab that displays all sources in a table with enable/disable toggles. It calls:
- `GET /admin/sources` 鈥?list all sources
- `POST /admin/sources/update` 鈥?toggle `enabled` or change `interval_sec`
- `POST /admin/sources/import` 鈥?import content for a source (exists but not wired into sources tab UI)

Missing: create, edit all fields, delete, manual trigger, RSS preview.

## Scope

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Backend Changes

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

#### New File: `services/policy-crawler/internal/admin/admin_sources.go`

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

Handler functions:

1. **`SourceCreateHandler(store SourceCRUDStore)`** 鈥?`POST /admin/sources/create`
   - Request: `{source_id, source_name, source_url, source_level, crawl_type, interval_sec, region_code}`
   - Validates required fields, calls `store.CreateSource()`
   - Returns `{code:0, data: source}`

2. **`SourceDeleteHandler(store SourceCRUDStore)`** 鈥?`POST /admin/sources/delete`
   - Request: `{source_id}`
   - Calls `store.DeleteSource()`
   - Returns `{code:0, message: "deleted"}`

3. **`SourceCrawlTriggerHandler(mgr CrawlTrigger)`** 鈥?`POST /admin/sources/crawl`
   - Request: `{source_id}`
   - Calls `mgr.CrawlSource(sourceID)` in a goroutine
   - Returns `{code:0, message: "crawl started"}`

4. **`RSSTestHandler()`** 鈥?`POST /admin/sources/test-rss`
   - Request: `{url: string}`
   - Fetches URL, parses as RSS/Atom, returns top 5 items: `{items: [{title, link}]}`
   - Returns `{code:0, data: {items}}` or error if fetch/parse fails

#### Store Interface Extension

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

Add to a new `SourceCRUDStore` interface in `admin_sources.go`:

```go
type SourceCRUDStore interface {
    DashboardStore
    CreateSource(src *SourceInfo) error
    DeleteSource(sourceID string) error
}
```

Implement in `crawler/store.go`:
- `CreateSource(src *admin.SourceInfo)` 鈥?`INSERT INTO policy_sources`
- `DeleteSource(sourceID string)` 鈥?`DELETE FROM policy_sources WHERE source_id = $1`

#### Existing Handler Extension

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

`SourceUpdateHandler` in `admin_dashboard.go` 鈥?extend the request struct to accept:
- `source_name *string`
- `source_url *string`
- `source_level *string`
- `crawl_type *string`
- `region_code *string`

Extend `UpdateSource()` in store to handle these fields.

#### Route Registration

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

Add to `cmd/main.go` route setup (after existing admin routes):
```
POST /admin/sources/create  鈫?SourceCreateHandler
POST /admin/sources/delete  鈫?SourceDeleteHandler
POST /admin/sources/crawl   鈫?SourceCrawlTriggerHandler
POST /admin/sources/test-rss 鈫?RSSTestHandler
```

### Frontend Changes

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

All changes in `admin_page.go` inline HTML, replacing the `loadSources()` function.

#### New UI Elements

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

1. **"+ 鏂板鏁版嵁婧? button** at top of sources tab
2. **Source form modal** with fields:
   - source_id (text, required for new, readonly for edit)
   - source_name (text, required)
   - crawl_type (select: govsite/file/rss/manual)
   - source_url (text, required for govsite/rss, hidden for manual)
   - source_level (select: HIGH/MEDIUM/LOW)
   - interval_sec (number, hidden for manual)
   - region_code (text)
   - Dynamic: URL and interval fields show/hide based on crawl_type
3. **Per-row action buttons**:
   - Edit (pencil icon) 鈫?opens same form pre-filled
   - Delete (trash icon) 鈫?confirmation dialog 鈫?`POST /admin/sources/delete`
   - Crawl (play icon, hidden for manual) 鈫?`POST /admin/sources/crawl`
   - Import (upload icon, shown for manual only) 鈫?opens import modal with title/url/textarea 鈫?`POST /admin/sources/import`
   - Test RSS (magnifier icon, shown for rss only) 鈫?`POST /admin/sources/test-rss` 鈫?shows preview
4. **crawl_type column** shows badges:
   - govsite 鈫?blue "鏀垮簻缃戠珯"
   - file 鈫?gray "鏂囦欢"
   - rss 鈫?green "RSS"
   - manual 鈫?yellow "鎵嬪姩"

#### JS Functions

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- `loadSources()` 鈥?rewritten to render enhanced table with action buttons
- `showSourceForm(source?)` 鈥?opens modal for create/edit
- `saveSource()` 鈥?POST create or update
- `deleteSource(id)` 鈥?confirm + delete
- `triggerCrawl(id)` 鈥?POST crawl trigger
- `testRSS(url)` 鈥?POST test-rss, show preview
- `showImportModal(sourceId, sourceName)` 鈥?opens import textarea modal
- `doSourceImport()` 鈥?calls existing `/admin/sources/import`

### CrawlTrigger Interface

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

```go
type CrawlTrigger interface {
    CrawlSource(sourceID string)
}
```

`CrawlerManager` already implements `CrawlSource(sourceID string)`. Pass the manager instance to the handler.

### RSS Test Implementation

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

Reuse `parseFeed()` from `crawler/rss_crawler.go` by extracting it into a shared function or by having the handler import and call it. Since `parseFeed` is unexported in the `crawler` package, options:
- Export it as `ParseFeed()` 鈥?simplest
- Duplicate the parsing logic in admin handler 鈥?avoids cross-package dependency
- **Chosen: Export `ParseFeed()`** from `crawler` package. One-line change.

## Files Modified

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

| File | Change |
|------|--------|
| `internal/admin/admin_sources.go` | **NEW** 鈥?4 handler functions + interfaces |
| `internal/admin/admin_dashboard.go` | Extend `SourceUpdateHandler` + `DashboardStore` to support full edit |
| `internal/admin/admin_page.go` | Rewrite `loadSources()` JS, add modal HTML + 7 JS functions |
| `internal/crawler/store.go` | Add `CreateSource()`, `DeleteSource()`, extend `UpdateSource()` |
| `internal/crawler/rss_crawler.go` | Export `parseFeed` 鈫?`ParseFeed` |
| `cmd/main.go` | Register 4 new routes, pass CrawlerManager to crawl trigger handler |

## Non-Goals

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- No new migration needed (policy_sources table already has all columns)
- No changes to api-server, actuarial-engine, or frontend clients
- No pagination for sources (current count is small, <20)
