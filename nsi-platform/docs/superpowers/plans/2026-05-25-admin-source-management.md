# Admin Source Management Implementation Plan

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add full CRUD + operational management (create, edit, delete, crawl trigger, RSS preview, manual import) for multi-level policy sources in the admin panel.

**Architecture:** Extend the existing inline-HTML admin panel in policy-crawler. Add 4 new API handlers in a new `admin_sources.go` file, extend the store with `CreateSource`/`DeleteSource`/full-field `UpdateSource`, export `ParseFeed` from the crawler package, and rewrite the `loadSources()` JS function with modal forms and per-row action buttons.

**Tech Stack:** Go 1.25, net/http, PostgreSQL (pgvector), inline HTML/JS

---

## File Structure

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

| Action | File | Responsibility |
|--------|------|----------------|
| CREATE | `internal/admin/admin_sources.go` | 4 new handlers + interfaces |
| MODIFY | `internal/admin/admin_dashboard.go` | Extend `SourceUpdateHandler` for full-field edits |
| MODIFY | `internal/admin/admin_page.go` | Rewrite `loadSources()` JS, add modal HTML + 7 JS functions |
| MODIFY | `internal/crawler/store.go` | Add `CreateSource()`, `DeleteSource()`, extend `UpdateSource()` |
| MODIFY | `internal/crawler/rss_crawler.go` | Export `parseFeed` ‚Ü?`ParseFeed` |
| MODIFY | `cmd/main.go` | Register 4 new routes, pass manager to crawl trigger |
| CREATE | `internal/admin/admin_sources_test.go` | Handler unit tests |

All paths relative to `services/policy-crawler/`.

---

### Task 1: Export `ParseFeed` from crawler package

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Modify: `internal/crawler/rss_crawler.go:102`

- [ ] **Step 1: Rename `parseFeed` to `ParseFeed` and update its callers**

In `internal/crawler/rss_crawler.go`, rename the function:

```go
func ParseFeed(data []byte) ([]FeedItem, error) {
```

The only caller is `Fetch()` in the same file at line 56, which already calls `parseFeed(body)`. Change it to `ParseFeed(body)`.

- [ ] **Step 2: Verify build**

Run: `cd services/policy-crawler && go build ./...`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/internal/crawler/rss_crawler.go
git commit -m "refactor: export ParseFeed for admin RSS preview"
```

---

### Task 2: Add store methods ‚Ä?CreateSource, DeleteSource, extend UpdateSource

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Modify: `internal/crawler/store.go`
- Modify: `internal/admin/admin_dashboard.go` (DashboardStore interface)

- [ ] **Step 1: Add `CreateSource` method to `DBStore`**

In `internal/crawler/store.go`, add after `UpdateSource` (around line 250):

```go
func (s *DBStore) CreateSource(src *admin.SourceInfo) error {
	_, err := s.db.Exec(`INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code)
		VALUES ($1, $2, $3, $4, 0.7, true, $5, $6, $7)`,
		src.SourceID, src.SourceName, src.SourceURL, src.SourceLevel,
		src.CrawlType, src.IntervalSec, src.RegionCode)
	return err
}

func (s *DBStore) DeleteSource(sourceID string) error {
	_, err := s.db.Exec(`DELETE FROM policy_sources WHERE source_id = $1`, sourceID)
	return err
}
```

- [ ] **Step 2: Extend `UpdateSource` to support all fields**

Replace the existing `UpdateSource` method (lines 238-250 in store.go):

```go
func (s *DBStore) UpdateSource(id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	var sets []string
	var args []interface{}
	argIdx := 1
	allowed := map[string]string{
		"enabled": "boolean", "interval_sec": "int",
		"source_name": "text", "source_url": "text",
		"source_level": "text", "crawl_type": "text", "region_code": "text",
	}
	for col, typ := range allowed {
		val, ok := updates[col]
		if !ok {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE policy_sources SET %s WHERE source_id = $%d",
		strings.Join(sets, ", "), argIdx)
	_, err := s.db.Exec(query, args...)
	return err
}
```

- [ ] **Step 3: Update `DashboardStore` interface**

In `internal/admin/admin_dashboard.go`, add to the `DashboardStore` interface (after `UpdateSource`):

```go
CreateSource(src *SourceInfo) error
DeleteSource(sourceID string) error
```

Change the `UpdateSource` signature in the interface from:

```go
UpdateSource(id string, enabled *bool, intervalSec *int) error
```

to:

```go
UpdateSource(id string, updates map[string]interface{}) error
```

- [ ] **Step 4: Update `SourceUpdateHandler` for full-field updates**

In `internal/admin/admin_dashboard.go`, replace `SourceUpdateHandler`:

```go
func SourceUpdateHandler(store DashboardStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			SourceID   string  `json:"source_id"`
			Enabled    *bool   `json:"enabled,omitempty"`
			IntervalSec *int   `json:"interval_sec,omitempty"`
			SourceName *string `json:"source_name,omitempty"`
			SourceURL  *string `json:"source_url,omitempty"`
			SourceLevel *string `json:"source_level,omitempty"`
			CrawlType  *string `json:"crawl_type,omitempty"`
			RegionCode *string `json:"region_code,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.SourceID == "" {
			respondError(w, http.StatusBadRequest, "source_id required")
			return
		}
		updates := make(map[string]interface{})
		if req.Enabled != nil {
			updates["enabled"] = *req.Enabled
		}
		if req.IntervalSec != nil {
			updates["interval_sec"] = *req.IntervalSec
		}
		if req.SourceName != nil {
			updates["source_name"] = *req.SourceName
		}
		if req.SourceURL != nil {
			updates["source_url"] = *req.SourceURL
		}
		if req.SourceLevel != nil {
			updates["source_level"] = *req.SourceLevel
		}
		if req.CrawlType != nil {
			updates["crawl_type"] = *req.CrawlType
		}
		if req.RegionCode != nil {
			updates["region_code"] = *req.RegionCode
		}
		if err := store.UpdateSource(req.SourceID, updates); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("update error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": "updated"})
	})
}
```

- [ ] **Step 5: Update `toggleSource` in admin_page.go to use new API**

In `admin_page.go`, the `toggleSource` function (line 219) sends `{source_id, enabled}`. This is still compatible with the updated handler, so no change needed.

- [ ] **Step 6: Verify build**

Run: `cd services/policy-crawler && go build ./...`
Expected: success

- [ ] **Step 7: Commit**

```bash
git add services/policy-crawler/internal/crawler/store.go services/policy-crawler/internal/admin/admin_dashboard.go
git commit -m "feat: add CreateSource/DeleteSource, extend UpdateSource for full-field edits"
```

---

### Task 3: Add 4 new admin handlers

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `internal/admin/admin_sources.go`

- [ ] **Step 1: Create `admin_sources.go`**

```go
package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/crawler"
)

type SourceCRUDStore interface {
	DashboardStore
}

type CrawlTrigger interface {
	CrawlSource(sourceID string)
}

func SourceCreateHandler(store SourceCRUDStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var src SourceInfo
		if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if src.SourceID == "" || src.SourceName == "" {
			respondError(w, http.StatusBadRequest, "source_id and source_name required")
			return
		}
		if src.CrawlType == "" {
			src.CrawlType = "govsite"
		}
		if src.SourceLevel == "" {
			src.SourceLevel = "MEDIUM"
		}
		if src.IntervalSec == 0 {
			src.IntervalSec = 86400
		}
		if err := store.CreateSource(&src); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("create error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": src})
	})
}

func SourceDeleteHandler(store SourceCRUDStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			SourceID string `json:"source_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.SourceID == "" {
			respondError(w, http.StatusBadRequest, "source_id required")
			return
		}
		if err := store.DeleteSource(req.SourceID); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("delete error: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": "deleted"})
	})
}

func SourceCrawlTriggerHandler(mgr CrawlTrigger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			SourceID string `json:"source_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.SourceID == "" {
			respondError(w, http.StatusBadRequest, "source_id required")
			return
		}
		go mgr.CrawlSource(req.SourceID)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": fmt.Sprintf("crawl triggered for %s", req.SourceID),
		})
	})
}

type FeedPreviewItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func RSSTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.URL == "" {
			respondError(w, http.StatusBadRequest, "url required")
			return
		}

		resp, err := http.DefaultClient.Get(req.URL)
		if err != nil {
			respondError(w, http.StatusBadGateway, fmt.Sprintf("fetch error: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			respondError(w, http.StatusBadGateway, fmt.Sprintf("HTTP %d", resp.StatusCode))
			return
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("read error: %v", err))
			return
		}

		items, err := crawler.ParseFeed(body)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("parse error: %v", err))
			return
		}

		limit := 5
		if len(items) < limit {
			limit = len(items)
		}
		preview := make([]FeedPreviewItem, limit)
		for i := 0; i < limit; i++ {
			preview[i] = FeedPreviewItem{Title: items[i].Title, Link: items[i].Link}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"total": len(items),
				"items": preview,
			},
		})
	})
}
```

- [ ] **Step 2: Verify build**

Run: `cd services/policy-crawler && go build ./...`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_sources.go
git commit -m "feat: add source create/delete/crawl-trigger/rss-test handlers"
```

---

### Task 4: Register new routes in main.go

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Modify: `cmd/main.go:123-131`

- [ ] **Step 1: Add 4 routes after existing source routes**

In `cmd/main.go`, after the line `mux.Handle("/admin/sources/import", ...)` (line 130), add:

```go
		mux.Handle("/admin/sources/create", adminAuth(admin.SourceCreateHandler(store)))
		mux.Handle("/admin/sources/delete", adminAuth(admin.SourceDeleteHandler(store)))
		mux.Handle("/admin/sources/crawl", adminAuth(admin.SourceCrawlTriggerHandler(manager)))
		mux.Handle("/admin/sources/test-rss", adminAuth(admin.RSSTestHandler()))
```

- [ ] **Step 2: Verify build**

Run: `cd services/policy-crawler && go build ./...`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/cmd/main.go
git commit -m "feat: register source CRUD and RSS test routes"
```

---

### Task 5: Rewrite admin UI sources tab

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Modify: `internal/admin/admin_page.go` (replace `loadSources` function + `toggleSource`, add modal HTML + new JS functions)

This is the largest task. The changes are all within the `<script>` block of `adminHTML` in `admin_page.go`.

- [ ] **Step 1: Replace `loadSources()` function (lines 203-216)**

Replace the entire `loadSources` function with:

```javascript
var sourceTypeMap={'govsite':'ÊîøÂ∫úÁΩëÁ´ô','file':'Êñá‰ª∂','rss':'RSS','manual':'ÊâãÂä®'};
var sourceTypeBadge={'govsite':'bg-blue','file':'bg-yellow','rss':'bg-green','manual':'bg-yellow'};
var sourceLevelBadge={'HIGH':'bg-green','MEDIUM':'bg-yellow','LOW':'bg-blue'};

function loadSources(){
  fetch('/admin/sources').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    var h='<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">'+
      '<span style="color:#6B7280;font-size:13px">ÂÖ?'+d.data.length+' ‰∏™Êï∞ÊçÆÊ∫ê</span>'+
      '<button class="btn btn-success" onclick="showSourceForm()">+ Êñ∞Â¢ûÊï∞ÊçÆÊ∫?/button></div>'+
      '<div style="overflow-x:auto"><table><tr><th>ÂêØÁî®</th><th>ÂêçÁß∞</th><th>URL</th><th>Á∫ßÂà´</th><th>Á±ªÂûã</th><th>Âú∞Âå∫</th><th>Èó¥Èöî</th><th>ÊúÄËøëÁà¨Âè?/th><th>Áä∂ÊÄ?/th><th>Êìç‰Ωú</th></tr>';
    d.data.forEach(function(s){
      var actions='<div style="display:flex;gap:4px;flex-wrap:nowrap">';
      actions+='<button class="btn btn-outline btn-sm" onclick="showSourceForm(\''+esc(s.source_id)+'\')" title="ÁºñËæë">‚ú?/button>';
      actions+='<button class="btn btn-danger btn-sm" onclick="deleteSource(\''+esc(s.source_id)+'\',\''+esc(s.source_name)+'\')" title="Âà†Èô§">‚ú?/button>';
      if(s.crawl_type!=='manual'){
        actions+='<button class="btn btn-primary btn-sm" onclick="triggerCrawl(\''+esc(s.source_id)+'\')" title="Ëß¶ÂèëÁà¨Âèñ">‚ñ?/button>';
      }
      if(s.crawl_type==='rss'){
        actions+='<button class="btn btn-outline btn-sm" onclick="testRSS(\''+esc(s.source_url).replace(/'/g,"\\'")+'\')" title="ÊµãËØïRSS">üîç</button>';
      }
      if(s.crawl_type==='manual'){
        actions+='<button class="btn btn-outline btn-sm" onclick="showImportModal(\''+esc(s.source_id)+'\',\''+esc(s.source_name)+'\')" title="ÂØºÂÖ•ÂÜÖÂÆπ">üì§</button>';
      }
      actions+='</div>';
      h+='<tr><td><label class="toggle"><input type="checkbox" '+(s.enabled?'checked':'')+' data-sid="'+esc(s.source_id)+'" onchange="toggleSource(this)"><span class="slider"></span></label></td>'+
      '<td style="font-weight:600;font-size:13px">'+esc(s.source_name)+'</td>'+
      '<td style="font-size:11px;max-width:180px;overflow:hidden;text-overflow:ellipsis" title="'+esc(s.source_url)+'">'+esc(s.source_url||'-')+'</td>'+
      '<td><span class="badge '+(sourceLevelBadge[s.source_level]||'bg-yellow')+'">'+esc(s.source_level)+'</span></td>'+
      '<td><span class="badge '+(sourceTypeBadge[s.crawl_type]||'bg-blue')+'">'+(sourceTypeMap[s.crawl_type]||esc(s.crawl_type))+'</span></td>'+
      '<td>'+(s.region_code||'-')+'</td>'+
      '<td style="font-size:12px">'+(s.crawl_type==='manual'?'-':s.interval_sec+'s')+'</td>'+
      '<td style="font-size:11px;color:#9CA3AF">'+(s.last_crawl||'-')+'</td>'+
      '<td>'+(s.last_status==='success'?'<span class="badge bg-green">ÊàêÂäü</span>':s.last_status==='failed'?'<span class="badge bg-red">Â§±Ë¥•</span>':'-')+'</td>'+
      '<td>'+actions+'</td></tr>';
    });
    h+='</table></div>'+
      '<div id="sourceModal" style="display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.4);z-index:1000;display:none">'+
      '<div style="background:#fff;border-radius:10px;padding:20px;max-width:500px;margin:60px auto;box-shadow:0 4px 20px rgba(0,0,0,0.15)">'+
      '<h3 id="sourceModalTitle" style="font-size:16px;margin-bottom:16px">Êñ∞Â¢ûÊï∞ÊçÆÊ∫?/h3>'+
      '<div style="display:grid;grid-template-columns:1fr 1fr;gap:10px">'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Êï∞ÊçÆÊ∫?ID</label><input id="sf_id" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">ÂêçÁß∞</label><input id="sf_name" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Á±ªÂûã</label><select id="sf_type" onchange="onTypeChange()" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      '<option value="govsite">ÊîøÂ∫úÁΩëÁ´ô</option><option value="file">Êñá‰ª∂</option><option value="rss">RSS</option><option value="manual">ÊâãÂä®</option></select></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Á∫ßÂà´</label><select id="sf_level" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      '<option value="HIGH">HIGH</option><option value="MEDIUM">MEDIUM</option><option value="LOW">LOW</option></select></div>'+
      '<div id="sf_url_wrap" style="grid-column:1/3"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">URL</label><input id="sf_url" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px" placeholder="https://..."></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Áà¨ÂèñÈó¥Èöî(Áß?</label><input id="sf_interval" type="number" value="86400" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Âú∞Âå∫‰ª£Á†Å</label><input id="sf_region" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px" placeholder="310000"></div>'+
      '</div>'+
      '<div style="display:flex;gap:8px;margin-top:16px;justify-content:flex-end">'+
      '<button class="btn btn-outline" onclick="closeSourceModal()">ÂèñÊ∂à</button>'+
      '<button class="btn btn-primary" onclick="saveSource()">‰øùÂ≠ò</button></div>'+
      '</div></div>'+
      '<div id="importModal" style="display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.4);z-index:1000">'+
      '<div style="background:#fff;border-radius:10px;padding:20px;max-width:550px;margin:60px auto;box-shadow:0 4px 20px rgba(0,0,0,0.15)">'+
      '<h3 id="importModalTitle" style="font-size:16px;margin-bottom:16px">ÂØºÂÖ•ÂÜÖÂÆπ</h3>'+
      '<input type="hidden" id="imp_source_id">'+
      '<div style="margin-bottom:10px"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Ê†áÈ¢ò</label><input id="imp_title" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div style="margin-bottom:10px"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Êù•Ê∫êURL (ÂèØÈÄ?</label><input id="imp_url" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div style="margin-bottom:10px"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">ÂÜÖÂÆπ</label><textarea id="imp_content" style="width:100%;min-height:180px;border:1px solid #D1D5DB;border-radius:6px;padding:10px;font-size:13px;font-family:monospace;resize:vertical"></textarea></div>'+
      '<div style="display:flex;gap:8px;justify-content:flex-end">'+
      '<button class="btn btn-outline" onclick="closeImportModal()">ÂèñÊ∂à</button>'+
      '<button class="btn btn-success" onclick="doSourceImport()">ÂØºÂÖ•</button></div>'+
      '</div></div>'+
      '<div id="rssPreview" style="display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.4);z-index:1000">'+
      '<div style="background:#fff;border-radius:10px;padding:20px;max-width:500px;margin:60px auto;box-shadow:0 4px 20px rgba(0,0,0,0.15)">'+
      '<h3 style="font-size:16px;margin-bottom:12px">RSS È¢ÑËßà</h3>'+
      '<div id="rssPreviewContent"></div>'+
      '<div style="margin-top:12px;text-align:right"><button class="btn btn-outline" onclick="closeRSSPreview()">ÂÖ≥Èó≠</button></div>'+
      '</div></div>';
    document.getElementById('app').innerHTML=h;
  }).catch(function(e){document.getElementById('app').innerHTML='<div class="card"><h3>Âä†ËΩΩÂ§±Ë¥•</h3><p>'+esc(e.message)+'</p></div>'});
}
```

- [ ] **Step 2: Add helper JS functions after `toggleSource`**

After the existing `toggleSource` function (line 224), add:

```javascript
var editingSourceId='';

function onTypeChange(){
  var t=document.getElementById('sf_type').value;
  var urlWrap=document.getElementById('sf_url_wrap');
  var intv=document.getElementById('sf_interval');
  if(t==='manual'){
    urlWrap.style.display='none';
    intv.parentElement.style.display='none';
  }else{
    urlWrap.style.display='';
    intv.parentElement.style.display='';
  }
}

function showSourceForm(id){
  editingSourceId=id||'';
  var modal=document.getElementById('sourceModal');
  document.getElementById('sourceModalTitle').textContent=id?'ÁºñËæëÊï∞ÊçÆÊ∫?:'Êñ∞Â¢ûÊï∞ÊçÆÊ∫?;
  var idInput=document.getElementById('sf_id');
  if(id){
    idInput.value=id;idInput.readOnly=true;
    fetch('/admin/sources').then(function(r){return r.json()}).then(function(d){
      var s=d.data.find(function(x){return x.source_id===id});
      if(s){
        document.getElementById('sf_name').value=s.source_name;
        document.getElementById('sf_type').value=s.crawl_type;
        document.getElementById('sf_level').value=s.source_level;
        document.getElementById('sf_url').value=s.source_url;
        document.getElementById('sf_interval').value=s.interval_sec;
        document.getElementById('sf_region').value=s.region_code;
        onTypeChange();
      }
    });
  }else{
    idInput.value='';idInput.readOnly=false;
    document.getElementById('sf_name').value='';
    document.getElementById('sf_type').value='govsite';
    document.getElementById('sf_level').value='MEDIUM';
    document.getElementById('sf_url').value='';
    document.getElementById('sf_interval').value='86400';
    document.getElementById('sf_region').value='';
    onTypeChange();
  }
  modal.style.display='block';
}

function closeSourceModal(){document.getElementById('sourceModal').style.display='none'}

function saveSource(){
  var id=document.getElementById('sf_id').value.trim();
  var name=document.getElementById('sf_name').value.trim();
  if(!id||!name){showToast('IDÂíåÂêçÁß∞‰∏çËÉΩ‰∏∫Á©?,'error');return}
  var payload={
    source_id:id,source_name:name,
    crawl_type:document.getElementById('sf_type').value,
    source_level:document.getElementById('sf_level').value,
    source_url:document.getElementById('sf_url').value,
    interval_sec:parseInt(document.getElementById('sf_interval').value)||86400,
    region_code:document.getElementById('sf_region').value
  };
  if(editingSourceId){
    payload.source_id=editingSourceId;
    fetch('/admin/sources/update',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
    .then(function(r){return r.json()}).then(function(d){
      if(d.code===0){showToast('Â∑≤‰øùÂ≠?,'success');closeSourceModal();loadSources()}
      else{showToast('‰øùÂ≠òÂ§±Ë¥•: '+(d.error||''),'error')}
    }).catch(function(){showToast('ËØ∑Ê±ÇÂ§±Ë¥•','error')});
  }else{
    fetch('/admin/sources/create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
    .then(function(r){return r.json()}).then(function(d){
      if(d.code===0){showToast('Â∑≤ÂàõÂª?,'success');closeSourceModal();loadSources()}
      else{showToast('ÂàõÂª∫Â§±Ë¥•: '+(d.error||''),'error')}
    }).catch(function(){showToast('ËØ∑Ê±ÇÂ§±Ë¥•','error')});
  }
}

function deleteSource(id,name){
  if(!confirm('Á°ÆÂÆöË¶ÅÂà†Èô?"'+name+'" ÂêóÔºü'))return;
  fetch('/admin/sources/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({source_id:id})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('Â∑≤Âà†Èô?,'success');loadSources()}
    else{showToast('Âà†Èô§Â§±Ë¥•: '+(d.error||''),'error')}
  }).catch(function(){showToast('ËØ∑Ê±ÇÂ§±Ë¥•','error')});
}

function triggerCrawl(id){
  fetch('/admin/sources/crawl',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({source_id:id})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0)showToast('Áà¨ÂèñÂ∑≤Ëß¶Âè?,'success');
    else showToast('Ëß¶ÂèëÂ§±Ë¥•: '+(d.error||''),'error');
  }).catch(function(){showToast('ËØ∑Ê±ÇÂ§±Ë¥•','error')});
}

function testRSS(url){
  fetch('/admin/sources/test-rss',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url:url})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code!==0){showToast('ÊµãËØïÂ§±Ë¥•: '+(d.error||''),'error');return}
    var items=d.data.items||[];
    var total=d.data.total||0;
    var h='<p style="font-size:13px;color:#6B7280;margin-bottom:8px">ÂÖ?'+total+' Êù°ÔºåÊòæÁ§∫Ââ?'+items.length+' Êù?/p>';
    items.forEach(function(it){
      h+='<div style="padding:6px 0;border-bottom:1px solid #F3F4F6;font-size:13px">'+
        '<div style="font-weight:600">'+esc(it.title||'(Êó†Ê†áÈ¢?')+'</div>'+
        '<a href="'+esc(it.link)+'" target="_blank" style="font-size:11px;color:#1A56DB;text-decoration:none">'+esc(it.link)+'</a></div>';
    });
    document.getElementById('rssPreviewContent').innerHTML=h;
    document.getElementById('rssPreview').style.display='block';
  }).catch(function(){showToast('ËØ∑Ê±ÇÂ§±Ë¥•','error')});
}

function closeRSSPreview(){document.getElementById('rssPreview').style.display='none'}

function showImportModal(sid,sname){
  document.getElementById('imp_source_id').value=sid;
  document.getElementById('importModalTitle').textContent='ÂØºÂÖ•ÂÜÖÂÆπ ‚Ä?'+sname;
  document.getElementById('imp_title').value='';
  document.getElementById('imp_url').value='';
  document.getElementById('imp_content').value='';
  document.getElementById('importModal').style.display='block';
}

function closeImportModal(){document.getElementById('importModal').style.display='none'}

function doSourceImport(){
  var sid=document.getElementById('imp_source_id').value;
  var title=document.getElementById('imp_title').value.trim();
  var content=document.getElementById('imp_content').value.trim();
  if(!title||!content){showToast('Ê†áÈ¢òÂíåÂÜÖÂÆπ‰∏çËÉΩ‰∏∫Á©?,'error');return}
  fetch('/admin/sources/import',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({source_id:sid,title:title,content:content,source_url:document.getElementById('imp_url').value})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('ÂØºÂÖ•ÊàêÂäüÔºåÂæÖÂÆ°Ê†∏','success');closeImportModal()}
    else{showToast('ÂØºÂÖ•Â§±Ë¥•: '+(d.error||''),'error')}
  }).catch(function(){showToast('ËØ∑Ê±ÇÂ§±Ë¥•','error')});
}
```

- [ ] **Step 3: Verify build**

Run: `cd services/policy-crawler && go build ./...`
Expected: success

- [ ] **Step 4: Run existing tests**

Run: `cd services/policy-crawler && go test ./internal/admin/... -v`
Expected: all existing tests pass (no regressions)

- [ ] **Step 5: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_page.go
git commit -m "feat: admin UI source management with CRUD, crawl trigger, RSS preview, manual import"
```

---

### Task 6: Add handler unit tests

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `internal/admin/admin_sources_test.go`

- [ ] **Step 1: Create test file**

```go
package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSourceCreateHandler_MissingFields(t *testing.T) {
	handler := SourceCreateHandler(nil)
	body := `{"source_id":"TEST-1"}`
	req := httptest.NewRequest("POST", "/admin/sources/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSourceCreateHandler_InvalidJSON(t *testing.T) {
	handler := SourceCreateHandler(nil)
	req := httptest.NewRequest("POST", "/admin/sources/create", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSourceDeleteHandler_MissingID(t *testing.T) {
	handler := SourceDeleteHandler(nil)
	body := `{}`
	req := httptest.NewRequest("POST", "/admin/sources/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSourceCrawlTriggerHandler_MissingID(t *testing.T) {
	handler := SourceCrawlTriggerHandler(nil)
	body := `{}`
	req := httptest.NewRequest("POST", "/admin/sources/crawl", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSourceCrawlTriggerHandler_Success(t *testing.T) {
	var triggered string
	mock := &mockCrawlTrigger{fn: func(id string) { triggered = id }}
	handler := SourceCrawlTriggerHandler(mock)
	body := `{"source_id":"TEST-SRC"}`
	req := httptest.NewRequest("POST", "/admin/sources/crawl", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["code"] != float64(0) {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
}

func TestRSSTestHandler_MissingURL(t *testing.T) {
	handler := RSSTestHandler()
	body := `{}`
	req := httptest.NewRequest("POST", "/admin/sources/test-rss", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRSSTestHandler_InvalidURL(t *testing.T) {
	handler := RSSTestHandler()
	body := `{"url":"http://127.0.0.1:1/nonexistent"}`
	req := httptest.NewRequest("POST", "/admin/sources/test-rss", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Errorf("expected non-200 for invalid URL, got %d", w.Code)
	}
}

func TestSourceUpdateHandler_FullFields(t *testing.T) {
	handler := SourceUpdateHandler(nil)
	body := `{"source_id":"TEST","source_name":"new name","source_url":"http://x","source_level":"LOW","crawl_type":"rss","region_code":"110000","interval_sec":3600}`
	req := httptest.NewRequest("POST", "/admin/sources/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

type mockCrawlTrigger struct {
	fn func(string)
}

func (m *mockCrawlTrigger) CrawlSource(id string) {
	if m.fn != nil {
		m.fn(id)
	}
}
```

- [ ] **Step 2: Run tests**

Run: `cd services/policy-crawler && go test ./internal/admin/... -v`
Expected: all tests pass

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_sources_test.go
git commit -m "test: add unit tests for source CRUD handlers"
```

---

### Task 7: Full build + test verification

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

- [ ] **Step 1: Build all services**

Run: `cd services/policy-crawler && go build ./...`

- [ ] **Step 2: Run all policy-crawler tests**

Run: `cd services/policy-crawler && go test ./... -v`

- [ ] **Step 3: Verify no other service is broken**

Run: `cd services/api-server && go build ./...` and `cd services/actuarial-engine && go build ./...`

Expected: all pass

- [ ] **Step 4: Final commit if any fixes needed**

```bash
git add -A && git commit -m "fix: resolve build/test issues from source management feature"
```
