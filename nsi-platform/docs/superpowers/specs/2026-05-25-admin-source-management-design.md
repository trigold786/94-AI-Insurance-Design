# Admin UI: Multi-Level Source Management

## Goal

Enable full CRUD + operational management of all policy source types (govsite, file, rss, manual) in the existing admin panel's "数据源管理" tab.

## Current State

The admin panel (`/admin`) has a "数据源管理" tab that displays all sources in a table with enable/disable toggles. It calls:
- `GET /admin/sources` — list all sources
- `POST /admin/sources/update` — toggle `enabled` or change `interval_sec`
- `POST /admin/sources/import` — import content for a source (exists but not wired into sources tab UI)

Missing: create, edit all fields, delete, manual trigger, RSS preview.

## Scope

### Backend Changes

#### New File: `services/policy-crawler/internal/admin/admin_sources.go`

Handler functions:

1. **`SourceCreateHandler(store SourceCRUDStore)`** — `POST /admin/sources/create`
   - Request: `{source_id, source_name, source_url, source_level, crawl_type, interval_sec, region_code}`
   - Validates required fields, calls `store.CreateSource()`
   - Returns `{code:0, data: source}`

2. **`SourceDeleteHandler(store SourceCRUDStore)`** — `POST /admin/sources/delete`
   - Request: `{source_id}`
   - Calls `store.DeleteSource()`
   - Returns `{code:0, message: "deleted"}`

3. **`SourceCrawlTriggerHandler(mgr CrawlTrigger)`** — `POST /admin/sources/crawl`
   - Request: `{source_id}`
   - Calls `mgr.CrawlSource(sourceID)` in a goroutine
   - Returns `{code:0, message: "crawl started"}`

4. **`RSSTestHandler()`** — `POST /admin/sources/test-rss`
   - Request: `{url: string}`
   - Fetches URL, parses as RSS/Atom, returns top 5 items: `{items: [{title, link}]}`
   - Returns `{code:0, data: {items}}` or error if fetch/parse fails

#### Store Interface Extension

Add to a new `SourceCRUDStore` interface in `admin_sources.go`:

```go
type SourceCRUDStore interface {
    DashboardStore
    CreateSource(src *SourceInfo) error
    DeleteSource(sourceID string) error
}
```

Implement in `crawler/store.go`:
- `CreateSource(src *admin.SourceInfo)` — `INSERT INTO policy_sources`
- `DeleteSource(sourceID string)` — `DELETE FROM policy_sources WHERE source_id = $1`

#### Existing Handler Extension

`SourceUpdateHandler` in `admin_dashboard.go` — extend the request struct to accept:
- `source_name *string`
- `source_url *string`
- `source_level *string`
- `crawl_type *string`
- `region_code *string`

Extend `UpdateSource()` in store to handle these fields.

#### Route Registration

Add to `cmd/main.go` route setup (after existing admin routes):
```
POST /admin/sources/create  → SourceCreateHandler
POST /admin/sources/delete  → SourceDeleteHandler
POST /admin/sources/crawl   → SourceCrawlTriggerHandler
POST /admin/sources/test-rss → RSSTestHandler
```

### Frontend Changes

All changes in `admin_page.go` inline HTML, replacing the `loadSources()` function.

#### New UI Elements

1. **"+ 新增数据源" button** at top of sources tab
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
   - Edit (pencil icon) → opens same form pre-filled
   - Delete (trash icon) → confirmation dialog → `POST /admin/sources/delete`
   - Crawl (play icon, hidden for manual) → `POST /admin/sources/crawl`
   - Import (upload icon, shown for manual only) → opens import modal with title/url/textarea → `POST /admin/sources/import`
   - Test RSS (magnifier icon, shown for rss only) → `POST /admin/sources/test-rss` → shows preview
4. **crawl_type column** shows badges:
   - govsite → blue "政府网站"
   - file → gray "文件"
   - rss → green "RSS"
   - manual → yellow "手动"

#### JS Functions

- `loadSources()` — rewritten to render enhanced table with action buttons
- `showSourceForm(source?)` — opens modal for create/edit
- `saveSource()` — POST create or update
- `deleteSource(id)` — confirm + delete
- `triggerCrawl(id)` — POST crawl trigger
- `testRSS(url)` — POST test-rss, show preview
- `showImportModal(sourceId, sourceName)` — opens import textarea modal
- `doSourceImport()` — calls existing `/admin/sources/import`

### CrawlTrigger Interface

```go
type CrawlTrigger interface {
    CrawlSource(sourceID string)
}
```

`CrawlerManager` already implements `CrawlSource(sourceID string)`. Pass the manager instance to the handler.

### RSS Test Implementation

Reuse `parseFeed()` from `crawler/rss_crawler.go` by extracting it into a shared function or by having the handler import and call it. Since `parseFeed` is unexported in the `crawler` package, options:
- Export it as `ParseFeed()` — simplest
- Duplicate the parsing logic in admin handler — avoids cross-package dependency
- **Chosen: Export `ParseFeed()`** from `crawler` package. One-line change.

## Files Modified

| File | Change |
|------|--------|
| `internal/admin/admin_sources.go` | **NEW** — 4 handler functions + interfaces |
| `internal/admin/admin_dashboard.go` | Extend `SourceUpdateHandler` + `DashboardStore` to support full edit |
| `internal/admin/admin_page.go` | Rewrite `loadSources()` JS, add modal HTML + 7 JS functions |
| `internal/crawler/store.go` | Add `CreateSource()`, `DeleteSource()`, extend `UpdateSource()` |
| `internal/crawler/rss_crawler.go` | Export `parseFeed` → `ParseFeed` |
| `cmd/main.go` | Register 4 new routes, pass CrawlerManager to crawl trigger handler |

## Non-Goals

- No new migration needed (policy_sources table already has all columns)
- No changes to api-server, actuarial-engine, or frontend clients
- No pagination for sources (current count is small, <20)
