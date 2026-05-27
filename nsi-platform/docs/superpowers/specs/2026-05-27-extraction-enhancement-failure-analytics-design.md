# LLM Extraction Enhancement & Failure Analytics Dashboard — Design Spec

**Date**: 2026-05-27
**Scope**: policy-crawler service (Sub-project 2 + Sub-project 3)

---

## Background

Current state:
- **748 claims** extracted from **2564 raw texts**, with **10643 crawl logs** and **4015 extract logs**
- **Extract failure rate: 78%** (3136 failed vs 879 success) — primary pain point
- `extractPlainText` hard-truncates at 8000 chars; long PDF/DOCX policies (20k+ chars) lose information
- `parseExtractionResult` fails on non-standard JSON output from LLM (markdown wrapping, missing commas, trailing text)
- Current `FailureAnalysisHandler` only aggregates by source+error_message, no charts/trends/retry
- No visibility into video extraction failures (`video_extract_status='failed'`)

## Goals

1. **Raise extraction success rate** from 22% to >70% via fault-tolerant parsing + document splitting
2. **Enrich extracted fields** with valuable policy metadata (title, authority, doc number, process, contact, type)
3. **Add quality metadata** to track extraction method, raw text length, model used
4. **Build failure analytics dashboard** covering crawl/extract/video-extract with charts, trends, and retry

---

## Sub-project 2: LLM Extraction Enhancement

### 2.1 Smart Document Splitting

**File**: `services/policy-crawler/internal/extractor/splitter.go` (new)

```go
func splitDocument(text string, maxChunkSize int) []string
```

- Split by paragraph boundaries (`\n\n`)
- Each chunk ≤ `maxChunkSize` chars (default 4000, leaving room for system prompt)
- Short documents (≤6000 chars): single-pass extraction, zero overhead
- Long documents: each chunk extracted independently → `[]ExtractionResult`
- Max 5 chunks (excess truncated, logged)
- Merge step: LLM call with prompt "以下是同一政策文档不同片段的提取结果，请合并为一个完整结果，保留所有信息"

**Flow**:
```
raw_text → clean → len ≤ 6000?
  YES → single LLM call → parse
  NO  → split → N LLM calls → N results → merge LLM call → parse
```

### 2.2 Fault-Tolerant Parsing

**File**: `services/policy-crawler/internal/extractor/parser.go` (new)

Three-level degradation:

**Level 1 — Standard JSON** (existing):
- Find `{...}` block, `json.Unmarshal`

**Level 2 — Repair parsing** (new):
- Strip markdown code block wrapping (```json ... ```)
- Strip trailing non-JSON text after last `}`
- Fix common LLM errors: single→double quotes, missing commas, trailing commas before `}`
- Re-attempt `json.Unmarshal`

**Level 3 — Regex fallback** (new):
- Extract fields individually via regex patterns from raw LLM output
- Fill what's extractable, leave rest empty
- Mark `extraction_method = 'regex_fallback'`

### 2.3 Field Enhancement

**New fields added to `ExtractionResult` struct and LLM prompt:**

| Field | DB Column | Type | Description |
|-------|-----------|------|-------------|
| `policy_title` | `policy_title` | text | Official policy title |
| `issuing_authority` | `issuing_authority` | text | Issuing government body |
| `document_number` | `document_number` | text | Document number (e.g. "沪人社规〔2024〕1号") |
| `application_process` | `application_process` | jsonb | Steps array `[{"step":1,"action":"...","description":"..."}]` |
| `contact_info` | `contact_info` | text | Phone/address/website for consultation |
| `source_type` | `source_type` | text | `gov_doc`/`social_media`/`news`/`rumor` |

All new columns are nullable — no breaking changes to existing data.

### 2.4 Quality Metadata

| Field | DB Column | Type | Default |
|-------|-----------|------|---------|
| extraction method | `extraction_method` | text | `'full'` |
| raw text length | `raw_text_length` | int | `0` |
| split count | `split_count` | int | `0` |

Values for `extraction_method`:
- `'full'` — single-pass extraction succeeded
- `'split'` — document was split, merged successfully
- `'regex_fallback'` — JSON parsing failed, regex extraction used

### 2.5 Migration 025

```sql
-- 025_extraction_enhancement.sql
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS policy_title text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS issuing_authority text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS document_number text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS application_process jsonb DEFAULT '[]'::jsonb;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS contact_info text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS source_type text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS extraction_method text DEFAULT 'full';
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS raw_text_length int DEFAULT 0;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS split_count int DEFAULT 0;
```

### 2.6 Updated LLM System Prompt

Extended JSON schema in system prompt:

```json
{
  "policy_id": "...",
  "policy_title": "政策正式标题",
  "issuing_authority": "发文机关",
  "document_number": "文号",
  "region_code": "6位地区代码",
  "policy_type": "...",
  "target_groups": ["..."],
  "subsidy_calc_method": "...",
  "amount_min": ...,
  "amount_max": ...,
  "subsidy_duration": ...,
  "effective_date": "YYYY-MM-DD",
  "expire_date": "YYYY-MM-DD",
  "policy_url": "...",
  "brief_summary": "一句话概括(≤50字)",
  "source_type": "gov_doc|social_media|news|rumor",
  "application_process": [{"step":1,"action":"步骤","description":"描述"}],
  "contact_info": "咨询电话/地址",
  "conditions": [{"name":"...","description":"...","tag_match":"..."}],
  "required_documents": [{"name":"...","description":"...","source":"user/gov","optional":false}]
}
```

---

## Sub-project 3: Failure Analytics Dashboard Enhancement

### 3.1 Three-Dimension Coverage

| Dimension | Source | Filter |
|-----------|--------|--------|
| Crawl failures | `crawl_logs WHERE status='failed'` | by source, date range |
| Extract failures | `extract_logs WHERE status='failed'` | by source, date range |
| Video extract failures | `policy_raw_texts WHERE video_extract_status='failed'` | by source, date range |

### 3.2 Data Layer — New Store Methods

**File**: `services/policy-crawler/internal/crawler/store.go` (extend)

```go
type FailureTrendPoint struct {
    Date           string `json:"date"`
    CrawlFailures  int    `json:"crawl_failures"`
    ExtractFailures int   `json:"extract_failures"`
    VideoFailures  int    `json:"video_failures"`
}

type FailureBySource struct {
    SourceID       string `json:"source_id"`
    SourceName     string `json:"source_name"`
    CrawlFailures  int    `json:"crawl_failures"`
    ExtractFailures int   `json:"extract_failures"`
    VideoFailures  int    `json:"video_failures"`
}

type TopFailureReason struct {
    Reason string `json:"reason"`
    Count  int    `json:"count"`
    Source string `json:"source"`
}

type FailedRawText struct {
    ID          int64  `json:"id"`
    SourceID    string `json:"source_id"`
    SourceName  string `json:"source_name"`
    Title       string `json:"title"`
    ErrorReason string `json:"error_reason"`
    FailedAt    string `json:"failed_at"`
    Type        string `json:"type"` // "extract" | "video"
}

func (s *DBStore) GetFailureTrend(days int) ([]FailureTrendPoint, error)
func (s *DBStore) GetFailureBySource() ([]FailureBySource, error)
func (s *DBStore) GetTopFailureReasons(limit int) ([]TopFailureReason, error)
func (s *DBStore) GetFailedRawTexts(sourceID string, limit int) ([]FailedRawText, error)
func (s *DBStore) RetryRawText(id int64) error
func (s *DBStore) RetryAllFailed(sourceID string) (int, error)
```

**Retry logic**:
- `RetryRawText(id)`: `UPDATE policy_raw_texts SET extracted=false, video_extract_status=NULL WHERE id=$1` — re-queues a single raw_text for the full pipeline (LLM extract + video extract)
- `RetryAllFailed(sourceID)`: Two-step reset:
  1. `UPDATE policy_raw_texts SET video_extract_status='pending' WHERE source_id=$1 AND video_extract_status='failed'` — re-queue video extracts
  2. Look up extract_logs for failed extractions of this source, get raw_text_ids, then `UPDATE policy_raw_texts SET extracted=false WHERE id = ANY($1)` — re-queue LLM extraction

### 3.3 Admin API Endpoints

**File**: `services/policy-crawler/internal/admin/admin_failures.go` (new)

| Endpoint | Method | Params | Description |
|----------|--------|--------|-------------|
| `/admin/failures/summary` | GET | — | Total counts for 3 dimensions |
| `/admin/failures/trend` | GET | `?days=7\|30` | Daily failure trend |
| `/admin/failures/by-source` | GET | — | Failures aggregated by source |
| `/admin/failures/top-reasons` | GET | `?limit=10` | Top N failure reasons |
| `/admin/failures/failed-raw-texts` | GET | `?source_id=&limit=50` | Failed raw_text detail list |
| `/admin/failures/retry` | POST | `{"raw_text_id":123}` or `{"source_id":"X","all":true}` | Retry failed items |

### 3.4 Admin UI — Failure Analytics Tab

**File**: `services/policy-crawler/internal/admin/admin_page.go` (extend)

New tab "失败分析" in admin navigation with:

1. **Summary cards** (top): 3 stat cards showing crawl/extract/video failure totals
2. **Trend chart**: Chart.js line chart, 7-day / 30-day toggle, 3 lines (crawl=red, extract=orange, video=purple)
3. **By-source pie chart**: Chart.js doughnut showing failure distribution across sources
4. **Top 10 reasons bar chart**: Chart.js horizontal bar
5. **Failed items table**: ID, source, title, error, time, retry button (per-row)
6. **Batch retry**: Select multiple rows → "Retry Selected" button

Chart.js v4 already loaded via CDN. CSP already allows `cdn.jsdelivr.net`.

### 3.5 Existing Code Impact

- `admin_dashboard.go`: `DashboardStore` interface adds new methods
- `failures.go`: `GetFailureAnalysis()` remains for backward compat, new methods added alongside
- `admin_page.go`: New tab added to HTML template
- `store.go`: New query methods + retry methods
- `cmd/main.go`: Register new `/admin/failures/*` routes

---

## Implementation Order

1. Migration 025 (DB schema)
2. Splitter (`splitter.go`)
3. Parser with fault-tolerance (`parser.go`)
4. Update `extractor.go` to use splitter + enhanced parser + new fields
5. Update store methods for new columns in INSERT/SELECT
6. Failure analytics data layer (store queries)
7. Failure analytics admin API handlers
8. Failure analytics admin UI tab
9. Build, deploy, verify

## Constraints

- No new Go dependencies (splitting/parsing done with stdlib + existing regex)
- New DB columns all nullable or have defaults (no migration risk)
- Existing data untouched (new columns NULL for old rows)
- Admin UI remains single-file HTML (`admin_page.go` const)
- Chart.js v4 via CDN (already configured)
- `GOOS=linux GOARCH=amd64` cross-compilation
- Container binary path: `/policy-crawler`
