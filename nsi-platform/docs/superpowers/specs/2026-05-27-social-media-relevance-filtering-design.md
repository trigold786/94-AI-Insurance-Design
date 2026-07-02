# Social Media Relevance Filtering & Video Content Extraction

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Date**: 2026-05-27
**Status**: Draft
**Sub-projects**: This is part of the Crawler Enhancement roadmap (Sub-project 1.5, between Sub-project 1 and 2)

## Problem Statement

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Social media crawlers (Douyin + WeChat) have **zero relevance filtering** and produce content too short for LLM extraction:

| Metric | Douyin | WeChat |
|--------|--------|--------|
| Raw texts stored | 268 | 3 |
| Avg content length | ~160 bytes | ~490 bytes |
| LLM extraction threshold | 500 bytes | 500 bytes |
| Items qualifying for LLM | 0 (0%) | 0 (0%) |
| Policy claims produced | 0 | 0 |
| Irrelevant content | ~80% (football, novels, cars, diapers) | Anti-bot pages |

Root causes:
1. Douyin user pages return platform recommendations mixed with creator's own videos �?no author filtering at discovery
2. No keyword/topic relevance check on any social media content
3. Only title + meta description extracted from Douyin (~160 bytes) �?no audio/subtitle extraction
4. WeChat articles return anti-bot JavaScript pages, not real content

## Design Goals

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

1. **Douyin**: Only crawl the specific creator's own videos, filter by relevance, extract video audio/subtitles
2. **WeChat**: Fix anti-bot issues, implement mixed discovery mode (search + account-based)
3. **Relevance filtering**: Admin-configurable keyword rules with weights, applied at two pipeline stages
4. **Video content extraction**: Subtitle-first, ASR-fallback for audio-to-text
5. **Configurability**: ASR has its own config unit; LLM has primary + backup

## Architecture: Two-Level Pipeline

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```
Level 1: Fast Discovery + Pre-filter (in Crawler.Fetch())
  ├─ Douyin: strict author filter �?keyword pre-filter on title+desc �?return lightweight metadata
  └─ WeChat: search + account discovery �?fetch full article �?keyword filter �?return full text
           �?           �?Manager.crawlAndProcess()
  ├─ SaveRawText (lightweight for Douyin, full for WeChat)
  ├─ WeChat: direct to LLM extraction
  └─ Douyin + videoURL: mark video_extract_status='pending', enqueue to VideoExtractQueue
           �?           �?Level 2: Async Video Deep Extract (VideoExtractWorker)
  ├─ yt-dlp: extract subtitles (priority)
  ├─ No subtitle �?download audio �?ASR �?transcript
  ├─ Merge: title + desc + transcript �?enriched text (~2000+ bytes)
  ├─ RelevanceFilter.Score(enriched) �?Level 2 threshold
  �?  ├─ Pass �?UpdateRawText, mark video_extract_status='done'
  �?  └─ Fail �?mark video_extract_status='discarded', extracted=true
  └─ Ready for LLM extraction pipeline
```

## Component 1: Relevance Filter Engine

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### Storage

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**`relevance_rules` table** (migration 023):

```sql
CREATE TABLE relevance_rules (
    id SERIAL PRIMARY KEY,
    category TEXT NOT NULL,           -- 险种/政策动词/金额时间/人群/政策文档/自定�?    keyword TEXT NOT NULL,
    weight INT NOT NULL DEFAULT 1,    -- 1=medium, 2=high
    scope TEXT NOT NULL DEFAULT 'all', -- all/douyin/wechat/govsite
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_relevance_rules_scope ON relevance_rules(scope) WHERE enabled;
```

**`relevance_thresholds` table** (per-source override):

```sql
CREATE TABLE relevance_thresholds (
    source_id TEXT PRIMARY KEY REFERENCES policy_sources(source_id),
    level1_min_score INT DEFAULT 1,
    level2_min_score INT DEFAULT 2,
    extra_keywords TEXT DEFAULT '',    -- comma-separated, weight 1
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

**Default seed data** (migration 023):

| category | keyword | weight | scope |
|----------|---------|--------|-------|
| 险种 | 社保 | 2 | all |
| 险种 | 养�?| 2 | all |
| 险种 | 医疗 | 2 | all |
| 险种 | 失业 | 2 | all |
| 险种 | 工伤 | 2 | all |
| 险种 | 生育 | 2 | all |
| 险种 | 公积�?| 2 | all |
| 政策动词 | 缴费 | 2 | all |
| 政策动词 | 补缴 | 2 | all |
| 政策动词 | 待遇 | 2 | all |
| 政策动词 | 领取 | 2 | all |
| 政策动词 | 办理 | 2 | all |
| 政策动词 | 退�?| 2 | all |
| 政策动词 | 延迟退�?| 2 | all |
| 金额时间 | 补贴 | 1 | all |
| 金额时间 | 报销 | 1 | all |
| 金额时间 | 基数 | 1 | all |
| 金额时间 | 比例 | 1 | all |
| 金额时间 | 标准 | 1 | all |
| 人群 | 职工 | 1 | all |
| 人群 | 灵活就业 | 1 | all |
| 人群 | 退休人�?| 1 | all |
| 人群 | 居民 | 1 | all |
| 政策文档 | 政策 | 1 | all |
| 政策文档 | 通知 | 1 | all |
| 政策文档 | 公告 | 1 | all |
| 政策文档 | 办法 | 1 | all |

### Go API

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```go
type RelevanceFilter struct {
    rules    []Rule
    byScope  map[string][]Rule  // scope �?rules
    extra    map[string][]Rule  // sourceID �?extra rules
    l1Min    map[string]int     // sourceID �?level1 threshold
    l2Min    map[string]int     // sourceID �?level2 threshold
}

type Rule struct {
    Keyword string
    Weight  int
    Scope   string
}

func (f *RelevanceFilter) Score(text, sourceID, level string) (score int, matched []string)
func (f *RelevanceFilter) LoadFromDB(db *sql.DB) error
func (f *RelevanceFilter) Reload() error  // hot reload
```

- Startup: `LoadFromDB()` loads all enabled rules + thresholds
- Background: every 5 minutes call `Reload()`, or admin triggers via channel
- `Score()` iterates all matching rules (scope=all OR scope matches crawlType), plus source-specific extra_keywords
- Returns total weighted score and list of matched keywords

### Admin API

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/admin/relevance/rules` | GET | List rules (filter by category, scope) |
| `/admin/relevance/rules` | POST | Create rule |
| `/admin/relevance/rules` | PUT | Update rule (weight, enabled, scope) |
| `/admin/relevance/rules` | DELETE | Delete rule |
| `/admin/relevance/thresholds/{source_id}` | GET/PUT | Get/set per-source thresholds + extra keywords |
| `/admin/relevance/test` | POST | Test text against current rules, return score + matched keywords |
| `/admin/relevance/bulk-import` | POST | Bulk import rules (JSON array) |

### Admin UI

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

New "相关性规�? tab in admin dashboard:
1. Rules table grouped by category, inline edit weight/enable/disable
2. Add rule form: keyword + category dropdown + weight + scope
3. Bulk import (paste JSON/CSV)
4. Per-source threshold config in source edit modal
5. Test panel: input text �?show matched keywords and score in real-time

## Component 2: Douyin Crawler Redesign

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### Discovery: Strict Author Filtering

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Current problem: `discoverVideosFromUserPage()` scrapes the user page HTML and finds ALL `/video/` links including platform recommendations.

New approach:
1. Use Douyin's internal API endpoint `/aweme/v1/web/aweme/post/` which returns only the creator's own videos (paginated)
2. Fallback: Chrome-render the user page, but only parse the "作品" (works) section �?ignore the "推荐" (recommended) section
3. For each discovered video URL, verify author nickname matches the SourceID embedded name before adding to results

```go
func (d *DouyinCrawler) discoverOwnVideos() ([]string, error) {
    // Strategy 1: API endpoint (preferred)
    urls, err := d.discoverViaAPI()
    if err == nil && len(urls) > 0 {
        return urls, nil
    }
    // Strategy 2: Chrome render + works section parsing
    return d.discoverViaChromeWorks()
}
```

API endpoint approach:
- URL: `https://www.douyin.com/aweme/v1/web/aweme/post/?sec_user_id=<UID>&count=20&max_cursor=0`
- Requires: `User-Agent`, `Cookie` (optional but helps), `Referer` headers
- Response: JSON with `aweme_list` containing video URLs and metadata
- Extract `aweme_id` �?construct URL: `https://www.douyin.com/video/<aweme_id>`

Chrome works section parsing:
- Render user page with `?showTab=works` parameter
- Parse only links within the works tab container (identified by specific CSS class or data attribute)
- Filter: link must contain `/video/` AND be within the creator's video grid

### Level 1: Pre-filter

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

After discovery, for each video:
1. Fetch video page (Chrome or HTTP) to get title + description
2. `RelevanceFilter.Score(title + " " + description, sourceID, "level1")`
3. If score < threshold �?skip, log reason
4. If passes �?return `CrawlResult{NeedsVideoExtract: true, VideoURL: videoURL}`

### CrawlResult Extension

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```go
type CrawlResult struct {
    SourceID          string
    SourceLevel       string
    RawText           string
    Title             string
    SourceURL         string
    FetchedAt         time.Time
    VersionHash       string
    Error             string
    VideoURL          string   // NEW: original video URL
    NeedsVideoExtract bool     // NEW: marks for Level 2 processing
    ContentType       string   // NEW: text / html / video-meta
}
```

## Component 3: Video Content Extractor

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### Docker Image Update

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Dockerfile additions:
```dockerfile
RUN apk add --no-cache python3 py3-pip ffmpeg
RUN pip3 install --no-cache-dir yt-dlp
```

### VideoExtractWorker

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```go
type VideoExtractWorker struct {
    store     Store
    queue     chan VideoExtractTask
    filter    *RelevanceFilter
    asrConfig ASRConfig
    maxWorkers int
}

type VideoExtractTask struct {
    RawTextID   int64
    SourceID    string
    VideoURL    string
    Title       string
    Description string
    RetryCount  int
}
```

Pipeline per task:
1. Set `video_extract_status = 'processing'` on raw_text row
2. Try yt-dlp subtitle extraction: `yt-dlp --write-sub --sub-lang zh --skip-download --output <tmp> <videoURL>`
3. If subtitle found �?read .vtt/.srt file �?clean �?transcript
4. If no subtitle:
   a. Download audio: `yt-dlp -x --audio-format mp3 --output <tmp> <videoURL>`
   b. Call ASR API with audio file
   c. transcript = ASR response text
5. Enriched content = format:
   ```
   【标题�?title>
   【描述�?description>
   【视频转录�?transcript>
   ```
6. `RelevanceFilter.Score(enriched, sourceID, "level2")`
   - Pass �?`UPDATE policy_raw_texts SET content = enriched, video_extract_status = 'done' WHERE id = rawTextID`
   - Fail �?`UPDATE policy_raw_texts SET video_extract_status = 'discarded', extracted = true WHERE id = rawTextID`
7. On error: increment retry_count, if < 3 �?re-enqueue with delay; else �?mark `video_extract_status = 'failed'`

### Concurrency & Rate Limiting

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- Worker count: configurable via env `VIDEO_EXTRACT_WORKERS` (default 2)
- Queue buffer: 100 tasks
- Rate limit: configurable per-source `request_delay_ms` applies between tasks for same source
- yt-dlp rate limit: `--ratelimit 1M` to avoid IP bans
- Temporary files: stored in `/tmp/video-extract/`, cleaned after processing

### Recovery on Restart

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

On startup:
```sql
SELECT id, source_id, content, source_url, title
FROM policy_raw_texts
WHERE video_extract_status IN ('pending', 'processing')
```
Re-enqueue all as `VideoExtractTask` (content is the lightweight title+desc).

## Component 4: Database Changes

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### Migration 023: Relevance + Video Extract

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```sql
-- Relevance rules
CREATE TABLE relevance_rules (
    id SERIAL PRIMARY KEY,
    category TEXT NOT NULL,
    keyword TEXT NOT NULL,
    weight INT NOT NULL DEFAULT 1,
    scope TEXT NOT NULL DEFAULT 'all',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_relevance_rules_scope ON relevance_rules(scope) WHERE enabled;

-- Per-source relevance thresholds
CREATE TABLE relevance_thresholds (
    source_id TEXT PRIMARY KEY REFERENCES policy_sources(source_id),
    level1_min_score INT DEFAULT 1,
    level2_min_score INT DEFAULT 2,
    extra_keywords TEXT DEFAULT '',
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Video extraction status on raw texts
ALTER TABLE policy_raw_texts ADD COLUMN video_extract_status TEXT DEFAULT NULL;
-- Values: NULL (not video), 'pending', 'processing', 'done', 'failed', 'discarded'
CREATE INDEX idx_raw_texts_vextract ON policy_raw_texts(video_extract_status) 
    WHERE video_extract_status IN ('pending', 'processing');

-- Seed default rules
INSERT INTO relevance_rules (category, keyword, weight, scope) VALUES
('险种', '社保', 2, 'all'),
('险种', '养�?, 2, 'all'),
('险种', '养老险', 2, 'all'),
('险种', '养老金', 2, 'all'),
('险种', '医疗', 2, 'all'),
('险种', '医保', 2, 'all'),
('险种', '失业', 2, 'all'),
('险种', '工伤', 2, 'all'),
('险种', '生育', 2, 'all'),
('险种', '公积�?, 2, 'all'),
('政策动词', '缴费', 2, 'all'),
('政策动词', '补缴', 2, 'all'),
('政策动词', '待遇', 2, 'all'),
('政策动词', '领取', 2, 'all'),
('政策动词', '办理', 2, 'all'),
('政策动词', '退�?, 2, 'all'),
('政策动词', '延迟退�?, 2, 'all'),
('政策动词', '退休年�?, 2, 'all'),
('政策动词', '参保', 2, 'all'),
('政策动词', '参保�?, 2, 'all'),
('政策动词', '缴费年限', 2, 'all'),
('金额时间', '补贴', 1, 'all'),
('金额时间', '报销', 1, 'all'),
('金额时间', '基数', 1, 'all'),
('金额时间', '比例', 1, 'all'),
('金额时间', '标准', 1, 'all'),
('金额时间', '金额', 1, 'all'),
('金额时间', '调整', 1, 'all'),
('金额时间', '上涨', 1, 'all'),
('人群', '职工', 1, 'all'),
('人群', '灵活就业', 1, 'all'),
('人群', '退休人�?, 1, 'all'),
('人群', '居民', 1, 'all'),
('人群', '外国�?, 1, 'all'),
('人群', '个体�?, 1, 'all'),
('政策文档', '政策', 1, 'all'),
('政策文档', '通知', 1, 'all'),
('政策文档', '公告', 1, 'all'),
('政策文档', '办法', 1, 'all'),
('政策文档', '规定', 1, 'all'),
('政策文档', '方案', 1, 'all'),
('政策文档', '意见', 1, 'all'),
('政策文档', '意见�?, 1, 'all');
```

### Migration 024: ASR Config + LLM Backup

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```sql
-- ASR configuration (separate config unit)
CREATE TABLE asr_configs (
    id SERIAL PRIMARY KEY,
    provider TEXT NOT NULL DEFAULT 'volcengine',  -- volcengine / xfyun / whisper-local
    api_key TEXT NOT NULL DEFAULT '',
    endpoint TEXT NOT NULL DEFAULT '',
    language TEXT NOT NULL DEFAULT 'zh',
    sample_rate INT NOT NULL DEFAULT 16000,
    enabled BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ DEFAULT now()
);
INSERT INTO asr_configs (provider, api_key, endpoint, enabled) 
VALUES ('volcengine', '', 'https://openspeech.bytedance.com/api/v1/auc', false);

-- LLM backup config columns
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_provider TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_api_key TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_endpoint TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_model_name TEXT DEFAULT '';
```

## Component 5: ASR Provider

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### Configuration

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Stored in `asr_configs` table, managed via admin UI.

Go struct:
```go
type ASRConfig struct {
    ID         int64  `json:"id"`
    Provider   string `json:"provider"`    // volcengine / xfyun / whisper-local
    APIKey     string `json:"api_key"`
    Endpoint   string `json:"endpoint"`
    Language   string `json:"language"`
    SampleRate int    `json:"sample_rate"`
    Enabled    bool   `json:"enabled"`
}
```

### Volcano Engine ASR Integration

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

API: ByteDance OpenSpeech (火山引擎语音技�?
- Endpoint: configurable (default `https://openspeech.bytedance.com/api/v1/auc`)
- Auth: Bearer token via API key
- Input: audio file (mp3/wav, mono, 16kHz)
- Output: transcript text with timestamps
- Flow:
  1. Read audio file �?base64 encode
  2. POST to endpoint with audio data
  3. Parse response for text segments
  4. Concatenate into single transcript

### Fallback Chain

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

1. Try configured ASR provider
2. If ASR fails: log error, mark task for retry
3. If retry exhausted: store whatever content we have (title+desc only), mark `video_extract_status='failed'`

## Component 6: LLM Backup Config

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### Storage

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Added to existing `llm_configs` table as `backup_*` columns.

### Behavior

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```go
func (e *LLMExtractor) callLLM(prompt string) (string, error) {
    // Try primary
    resp, err := e.callProvider(e.config.Endpoint, e.config.APIKey, e.config.ModelName, prompt)
    if err == nil {
        return resp, nil
    }
    log.Printf("[llm] primary failed: %v, trying backup", err)
    // Try backup
    if e.config.BackupProvider != "" && e.config.BackupAPIKey != "" {
        resp, err = e.callProvider(e.config.BackupEndpoint, e.config.BackupAPIKey, e.config.BackupModelName, prompt)
        if err == nil {
            return resp, nil
        }
        log.Printf("[llm] backup also failed: %v", err)
    }
    return "", fmt.Errorf("both primary and backup LLM failed")
}
```

## Component 7: WeChat Mixed Mode

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### Search Engine Discovery (existing, improved)

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Keep current Baidu/Sogou/Bing search, but add:
- Extract `__biz` parameter from discovered article URLs
- Deduplicate by `__biz + mid` (article identifier)

### Account-Based Discovery (new)

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

For WeChat sources with article URLs containing `__biz`:
1. Extract `__biz` from source_url
2. Construct refined search: `site:mp.weixin.qq.com __biz=<BIZ_ID>` via search engines
3. This finds more articles from the same account without needing profile page access
4. Merge with search engine results, deduplicate

### Anti-Bot Mitigation

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Current problem: Chrome renderer returns anti-bot JavaScript pages.

Improvements:
1. Increase Chrome render wait time for WeChat articles (15s �?25s)
2. Add cookie support: load cookies from file (exported from browser session)
3. Use `RenderWithVirtualTime` with longer virtual time (simulate human reading)
4. Fallback: if Chrome fails, try plain HTTP with mobile User-Agent (WeChat mobile web is simpler)

### Content Extraction Fix

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Current `extractWeChatContent()` looks for `id="js_content"`. Anti-bot pages don't have this.
- Add check: if `js_content` not found AND page contains "验证" / "环境异常" �?mark as anti-bot, skip
- Log anti-bot detections to help tune rendering parameters

## Component 8: Manager Pipeline Changes

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### crawlAndProcess() Update

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```go
func (m *CrawlerManager) crawlAndProcess(s Source) {
    results, err := s.Fetch()
    // ... error handling ...

    for _, result := range results {
        rawTextID, err := m.store.SaveRawTextReturningID(result.SourceID, result.Title, result.RawText, result.SourceURL, result.VersionHash)
        // SaveRawTextReturningID does dedup check + INSERT RETURNING id (or returns existing id if dedup)

        if result.NeedsVideoExtract {
            // Level 2: enqueue for async video extraction
            m.store.SetVideoExtractStatusByID(rawTextID, "pending")
            m.videoQueue <- VideoExtractTask{
                RawTextID: rawTextID,
                SourceID:  result.SourceID,
                VideoURL:  result.VideoURL,
                Title:     result.Title,
            }
            m.store.SaveCrawlLogWithDetails(result.SourceID, true, "pending video extraction", "", "")
            continue
        }

        // Existing processing path (wechat full text, govsite, etc.)
        // ... unchanged ...
    }
}
```

### LLM Extraction Query Update

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

Update `GetUnprocessedRawTexts` to also pick up completed video extractions:

```sql
WHERE NOT prt.extracted 
  AND LENGTH(prt.content) >= 500 
  AND (prt.video_extract_status IS NULL OR prt.video_extract_status = 'done')
```

This ensures 'failed' and 'discarded' items are excluded from LLM extraction.

## Component 9: Startup Recovery

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

On `CrawlerManager` initialization:
1. Query for pending/processing video extracts
2. Re-enqueue them as tasks
3. Log count for monitoring

```go
func (m *CrawlerManager) recoverPendingVideoExtracts() {
    tasks, err := m.store.GetPendingVideoExtracts()
    if err != nil {
        log.Printf("[video-extract] recovery query failed: %v", err)
        return
    }
    for _, t := range tasks {
        m.videoQueue <- t
    }
    if len(tasks) > 0 {
        log.Printf("[video-extract] recovered %d pending tasks", len(tasks))
    }
}
```

## File Structure

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

New files:
```
services/policy-crawler/internal/
├── crawler/
�?  ├── relevance.go          # RelevanceFilter + Rule types + Score()
�?  ├── video_extract.go      # VideoExtractWorker, task queue, pipeline
�?  └── asr.go                # ASR provider interface + volcengine impl
├── admin/
�?  ├── admin_relevance.go    # Relevance rules admin API handlers
�?  └── admin_asr.go          # ASR config admin API handlers
migrations/
├── 023_relevance_video.sql   # relevance_rules, relevance_thresholds, video_extract_status
└── 024_asr_llm_backup.sql   # asr_configs, backup LLM columns
```

Modified files:
```
services/policy-crawler/internal/
├── crawler/
�?  ├── crawler.go            # CrawlResult: add VideoURL, NeedsVideoExtract, ContentType
�?  ├── douyin_crawler.go     # Redesign discovery (API + works section), add pre-filter
�?  ├── wechat_crawler.go     # Mixed mode, anti-bot mitigation, __biz discovery
�?  ├── manager.go            # VideoExtractQueue integration, recovery on startup
�?  └── store.go              # video_extract_status CRUD, GetPendingVideoExtracts
├── admin/
�?  ├── admin_page.go         # New "相关性规�? tab, ASR config section, test panel
�?  ├── admin_dashboard.go    # SourceInfo: no changes needed
�?  └── admin_llm.go          # LLMConfig: add backup fields, save/load handlers
├── extractor/
�?  └── llm_extractor.go      # Primary/backup fallback logic
cmd/main.go                   # Init VideoExtractWorker, recovery, routes for new admin APIs
services/policy-crawler/Dockerfile  # Add yt-dlp, ffmpeg, python3
```

## Success Metrics

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| Metric | Current | Target |
|--------|---------|--------|
| Douyin relevance (stored) | ~20% | >90% |
| Douyin content avg length | ~160 bytes | >1500 bytes |
| Douyin items qualifying for LLM | 0 | >80% of stored items |
| Douyin policy claims | 0 | >5 per crawl cycle |
| WeChat anti-bot rate | ~100% | <20% |
| WeChat articles with real content | 0 | >10 per crawl cycle |
| False positives (irrelevant stored) | ~80% | <10% |
