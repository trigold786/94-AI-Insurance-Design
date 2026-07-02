# Social Media Relevance Filtering & Video Extraction Implementation Plan

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement two-level pipeline for Douyin/WeChat crawlers with admin-configurable relevance filtering and async video content extraction (subtitle-first, ASR-fallback).

**Architecture:** Level 1 (fast discovery + keyword pre-filter in Fetch()) 鈫?Level 2 (async VideoExtractWorker: yt-dlp subtitle 鈫?ASR fallback 鈫?enriched content). Relevance rules stored in DB, managed via admin UI. ASR config as separate unit. LLM primary+backup.

**Tech Stack:** Go 1.22, PostgreSQL, yt-dlp (CLI), Volcano Engine ASR API, Chart.js (admin UI)

---

## File Structure

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**New files:**
- `services/policy-crawler/migrations/023_relevance_video.sql` 鈥?relevance_rules, relevance_thresholds, video_extract_status, seed data
- `services/policy-crawler/migrations/024_asr_llm_backup.sql` 鈥?asr_configs table, LLM backup columns
- `services/policy-crawler/internal/crawler/relevance.go` 鈥?RelevanceFilter, Rule, Score()
- `services/policy-crawler/internal/crawler/video_extract.go` 鈥?VideoExtractWorker, task queue, pipeline
- `services/policy-crawler/internal/crawler/asr.go` 鈥?ASRProvider interface + volcengine impl
- `services/policy-crawler/internal/admin/admin_relevance.go` 鈥?relevance rules admin API
- `services/policy-crawler/internal/admin/admin_asr.go` 鈥?ASR config admin API

**Modified files:**
- `services/policy-crawler/internal/crawler/crawler.go` 鈥?CrawlResult: add VideoURL, NeedsVideoExtract, ContentType
- `services/policy-crawler/internal/crawler/douyin_crawler.go` 鈥?strict author discovery, relevance pre-filter, set NeedsVideoExtract
- `services/policy-crawler/internal/crawler/wechat_crawler.go` 鈥?mixed mode, __biz discovery, anti-bot mitigation
- `services/policy-crawler/internal/crawler/manager.go` 鈥?VideoExtractQueue init, recovery, crawlAndProcess dispatch
- `services/policy-crawler/internal/crawler/store.go` 鈥?SaveRawTextReturningID, video_extract_status CRUD, GetPendingVideoExtracts, GetUnprocessedRawTexts update
- `services/policy-crawler/internal/admin/admin_llm.go` 鈥?LLMConfig: add backup fields
- `services/policy-crawler/internal/admin/admin_page.go` 鈥?new "鐩稿叧鎬ц鍒? tab, ASR section, test panel
- `services/policy-crawler/internal/llm/llm.go` 鈥?Client: add backup config, fallback logic
- `services/policy-crawler/cmd/main.go` 鈥?init worker, routes, recovery
- `services/policy-crawler/Dockerfile` 鈥?add python3, ffmpeg, yt-dlp

---

### Task 1: Migrations 023 + 024

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/migrations/023_relevance_video.sql`
- Create: `services/policy-crawler/migrations/024_asr_llm_backup.sql`

- [ ] **Step 1: Create migration 023**

```sql
-- 023_relevance_video.sql
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

CREATE TABLE relevance_thresholds (
    source_id TEXT PRIMARY KEY REFERENCES policy_sources(source_id),
    level1_min_score INT DEFAULT 1,
    level2_min_score INT DEFAULT 2,
    extra_keywords TEXT DEFAULT '',
    updated_at TIMESTAMPTZ DEFAULT now()
);

ALTER TABLE policy_raw_texts ADD COLUMN video_extract_status TEXT DEFAULT NULL;
CREATE INDEX idx_raw_texts_vextract ON policy_raw_texts(video_extract_status)
    WHERE video_extract_status IN ('pending', 'processing');

INSERT INTO relevance_rules (category, keyword, weight, scope) VALUES
('闄╃','绀句繚',2,'all'),('闄╃','鍏昏�?,2,'all'),('闄╃','鍏昏�侀櫓',2,'all'),('闄╃','鍏昏�侀噾',2,'all'),
('闄╃','鍖荤枟',2,'all'),('闄╃','鍖讳繚',2,'all'),('闄╃','澶变笟',2,'all'),('闄╃','宸ヤ激',2,'all'),
('闄╃','鐢熻偛',2,'all'),('闄╃','鍏Н閲?,2,'all'),
('鏀跨瓥鍔ㄨ瘝','缂磋垂',2,'all'),('鏀跨瓥鍔ㄨ瘝','琛ョ即',2,'all'),('鏀跨瓥鍔ㄨ瘝','寰呴亣',2,'all'),
('鏀跨瓥鍔ㄨ瘝','棰嗗彇',2,'all'),('鏀跨瓥鍔ㄨ瘝','鍔炵悊',2,'all'),('鏀跨瓥鍔ㄨ瘝','閫�浼?,2,'all'),
('鏀跨瓥鍔ㄨ瘝','寤惰繜閫�浼?,2,'all'),('鏀跨瓥鍔ㄨ瘝','閫�浼戝勾榫?,2,'all'),('鏀跨瓥鍔ㄨ瘝','鍙備繚',2,'all'),
('鏀跨瓥鍔ㄨ瘝','鍙備繚浜?,2,'all'),('鏀跨瓥鍔ㄨ瘝','缂磋垂骞撮檺',2,'all'),
('閲戦鏃堕棿','琛ヨ创',1,'all'),('閲戦鏃堕棿','鎶ラ攢',1,'all'),('閲戦鏃堕棿','鍩烘暟',1,'all'),
('閲戦鏃堕棿','姣斾緥',1,'all'),('閲戦鏃堕棿','鏍囧噯',1,'all'),('閲戦鏃堕棿','閲戦',1,'all'),
('閲戦鏃堕棿','璋冩暣',1,'all'),('閲戦鏃堕棿','涓婃定',1,'all'),
('浜虹兢','鑱屽伐',1,'all'),('浜虹兢','鐏垫椿灏变笟',1,'all'),('浜虹兢','閫�浼戜汉鍛?,1,'all'),
('浜虹兢','灞呮皯',1,'all'),('浜虹兢','澶栧浗浜?,1,'all'),('浜虹兢','涓綋鎴?,1,'all'),
('鏀跨瓥鏂囨。','鏀跨瓥',1,'all'),('鏀跨瓥鏂囨。','閫氱煡',1,'all'),('鏀跨瓥鏂囨。','鍏憡',1,'all'),
('鏀跨瓥鏂囨。','鍔炴硶',1,'all'),('鏀跨瓥鏂囨。','瑙勫畾',1,'all'),('鏀跨瓥鏂囨。','鏂规',1,'all'),
('鏀跨瓥鏂囨。','鎰忚',1,'all'),('鏀跨瓥鏂囨。','鎰忚绋?,1,'all');
```

- [ ] **Step 2: Create migration 024**

```sql
-- 024_asr_llm_backup.sql
CREATE TABLE asr_configs (
    id SERIAL PRIMARY KEY,
    provider TEXT NOT NULL DEFAULT 'volcengine',
    api_key TEXT NOT NULL DEFAULT '',
    endpoint TEXT NOT NULL DEFAULT '',
    language TEXT NOT NULL DEFAULT 'zh',
    sample_rate INT NOT NULL DEFAULT 16000,
    enabled BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ DEFAULT now()
);
INSERT INTO asr_configs (provider, api_key, endpoint, enabled)
VALUES ('volcengine', '', 'https://openspeech.bytedance.com/api/v1/auc', false);

ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_provider TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_api_key TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_endpoint TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_model_name TEXT DEFAULT '';
```

- [ ] **Step 3: Verify migrations compile**

Run: `docker compose up -d db-migrate && docker logs nsi-db-migrate 2>&1 | Select-Object -Last 20`
Expected: migration 023 and 024 applied successfully

- [ ] **Step 4: Commit**

```bash
git add services/policy-crawler/migrations/023_relevance_video.sql services/policy-crawler/migrations/024_asr_llm_backup.sql
git commit -m "feat: migrations for relevance rules, video extract status, ASR config, LLM backup"
```

---

### Task 2: CrawlResult Extension

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/crawler/crawler.go`

- [ ] **Step 1: Add new fields to CrawlResult**

In `crawler.go`, update the `CrawlResult` struct:

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
	VideoURL          string
	NeedsVideoExtract bool
	ContentType       string
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...` in `services/policy-crawler`
Expected: compiles clean (new fields are opt-in, no breakage)

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/internal/crawler/crawler.go
git commit -m "feat: add VideoURL, NeedsVideoExtract, ContentType to CrawlResult"
```

---

### Task 3: Relevance Filter Engine

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/crawler/relevance.go`
- Create: `services/policy-crawler/internal/crawler/relevance_test.go`

- [ ] **Step 1: Write test for RelevanceFilter.Score()**

Create `relevance_test.go`:

```go
package crawler

import "testing"

func TestRelevanceScoreBasic(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "绀句繚", Weight: 2, Scope: "all"},
		{Keyword: "鍏昏�?, Weight: 2, Scope: "all"},
		{Keyword: "琛ヨ创", Weight: 1, Scope: "all"},
		{Keyword: "鑱屽伐", Weight: 1, Scope: "all"},
	})
	score, matched := filter.Score("涓婃捣绀句繚缂磋垂鍩烘暟璋冩暣閫氱煡", "DOUYIN-test", "douyin")
	if score < 2 {
		t.Errorf("expected score >= 2 for 绀句繚+璋冩暣, got %d", score)
	}
	if len(matched) == 0 {
		t.Error("expected at least one match")
	}
}

func TestRelevanceScoreIrrelevant(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "绀句繚", Weight: 2, Scope: "all"},
		{Keyword: "鍏昏�?, Weight: 2, Scope: "all"},
	})
	score, _ := filter.Score("鐣寗鐣呭惉鍏嶈垂鐪嬪悗缁俯鏆栧尰鐢熷皬璇?, "DOUYIN-test", "douyin")
	if score != 0 {
		t.Errorf("expected score 0 for irrelevant text, got %d", score)
	}
}

func TestRelevanceScoreScope(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "绀句繚", Weight: 2, Scope: "douyin"},
		{Keyword: "鍏昏�?, Weight: 2, Scope: "wechat"},
	})
	score1, _ := filter.Score("绀句繚缂磋垂", "SRC", "douyin")
	if score1 != 2 {
		t.Errorf("expected 2 for douyin scope match, got %d", score1)
	}
	score2, _ := filter.Score("绀句繚缂磋垂", "SRC", "wechat")
	if score2 != 0 {
		t.Errorf("expected 0 for wechat scope miss, got %d", score2)
	}
	score3, _ := filter.Score("鍏昏�?, "SRC", "wechat")
	if score3 != 2 {
		t.Errorf("expected 2 for wechat scope match, got %d", score3)
	}
}

func TestRelevanceExtraKeywords(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "绀句繚", Weight: 2, Scope: "all"},
	})
	filter.SetExtraKeywords("SRC1", []string{"闂佃", "娴︿笢"})
	score, _ := filter.Score("闂佃鍖虹ぞ淇濅腑蹇?, "SRC1", "douyin")
	if score < 3 {
		t.Errorf("expected score >= 3 (绀句繚+闂佃), got %d", score)
	}
	score2, _ := filter.Score("闂佃鍖虹ぞ淇濅腑蹇?, "SRC2", "douyin")
	if score2 != 2 {
		t.Errorf("expected score 2 (only 绀句繚, no extra for SRC2), got %d", score2)
	}
}

func TestRelevanceThresholds(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "绀句繚", Weight: 2, Scope: "all"},
	})
	filter.SetThresholds("SRC1", 3, 5)
	if filter.MinScore("SRC1", "level1") != 3 {
		t.Error("level1 threshold should be 3")
	}
	if filter.MinScore("SRC1", "level2") != 5 {
		t.Error("level2 threshold should be 5")
	}
	if filter.MinScore("SRC2", "level1") != 1 {
		t.Error("default level1 threshold should be 1")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/crawler/ -run TestRelevance -v`
Expected: FAIL (NewRelevanceFilter not defined)

- [ ] **Step 3: Implement RelevanceFilter**

Create `relevance.go`:

```go
package crawler

import (
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"
)

type Rule struct {
	Keyword string
	Weight  int
	Scope   string
}

type RelevanceFilter struct {
	mu       sync.RWMutex
	rules    []Rule
	extra    map[string][]Rule
	l1Min    map[string]int
	l2Min    map[string]int
	db       *sql.DB
	lastLoad time.Time
}

func NewRelevanceFilter(rules []Rule) *RelevanceFilter {
	return &RelevanceFilter{
		rules: rules,
		extra: make(map[string][]Rule),
		l1Min: make(map[string]int),
		l2Min: make(map[string]int),
	}
}

func (f *RelevanceFilter) SetExtraKeywords(sourceID string, keywords []string) {
	var rules []Rule
	for _, kw := range keywords {
		rules = append(rules, Rule{Keyword: kw, Weight: 1, Scope: "all"})
	}
	f.mu.Lock()
	f.extra[sourceID] = rules
	f.mu.Unlock()
}

func (f *RelevanceFilter) SetThresholds(sourceID string, l1, l2 int) {
	f.mu.Lock()
	f.l1Min[sourceID] = l1
	f.l2Min[sourceID] = l2
	f.mu.Unlock()
}

func (f *RelevanceFilter) MinScore(sourceID, level string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	m := f.l1Min
	if level == "level2" {
		m = f.l2Min
	}
	if v, ok := m[sourceID]; ok {
		return v
	}
	if level == "level1" {
		return 1
	}
	return 2
}

func (f *RelevanceFilter) Score(text, sourceID, crawlType string) (int, []string) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	lower := strings.ToLower(text)
	var total int
	var matched []string
	allRules := f.collectRules(sourceID, crawlType)
	for _, r := range allRules {
		if strings.Contains(lower, strings.ToLower(r.Keyword)) {
			total += r.Weight
			matched = append(matched, r.Keyword)
		}
	}
	return total, matched
}

func (f *RelevanceFilter) collectRules(sourceID, crawlType string) []Rule {
	var result []Rule
	for _, r := range f.rules {
		if r.Scope == "all" || r.Scope == crawlType {
			result = append(result, r)
		}
	}
	if extra, ok := f.extra[sourceID]; ok {
		result = append(result, extra...)
	}
	return result
}

func (f *RelevanceFilter) LoadFromDB(db *sql.DB) error {
	f.db = db
	return f.Reload()
}

func (f *RelevanceFilter) Reload() error {
	if f.db == nil {
		return nil
	}
	rows, err := f.db.Query(`SELECT keyword, weight, scope FROM relevance_rules WHERE enabled`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var rules []Rule
	for rows.Next() {
		var r Rule
		if err := rows.Scan(&r.Keyword, &r.Weight, &r.Scope); err != nil {
			return err
		}
		rules = append(rules, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	tRows, err := f.db.Query(`SELECT source_id, level1_min_score, level2_min_score, extra_keywords FROM relevance_thresholds`)
	if err != nil {
		return err
	}
	defer tRows.Close()
	extra := make(map[string][]Rule)
	l1Min := make(map[string]int)
	l2Min := make(map[string]int)
	for tRows.Next() {
		var sid string
		var l1, l2 int
		var ek string
		if err := tRows.Scan(&sid, &l1, &l2, &ek); err != nil {
			return err
		}
		l1Min[sid] = l1
		l2Min[sid] = l2
		if ek != "" {
			for _, kw := range strings.Split(ek, ",") {
				kw = strings.TrimSpace(kw)
				if kw != "" {
					extra[sid] = append(extra[sid], Rule{Keyword: kw, Weight: 1, Scope: "all"})
				}
			}
		}
	}

	f.mu.Lock()
	f.rules = rules
	f.extra = extra
	f.l1Min = l1Min
	f.l2Min = l2Min
	f.lastLoad = time.Now()
	f.mu.Unlock()
	log.Printf("[relevance] loaded %d rules, %d source overrides", len(rules), len(l1Min))
	return nil
}

func (f *RelevanceFilter) StartReloadLoop(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := f.Reload(); err != nil {
				log.Printf("[relevance] reload error: %v", err)
			}
		case <-stopCh:
			return
		}
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/crawler/ -run TestRelevance -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/policy-crawler/internal/crawler/relevance.go services/policy-crawler/internal/crawler/relevance_test.go
git commit -m "feat: RelevanceFilter with weighted keyword scoring, scope, extra keywords, thresholds"
```

---

### Task 4: Store Updates for Video Extract

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/crawler/store.go`

- [ ] **Step 1: Add SaveRawTextReturningID**

Add to `store.go` after `SaveRawText`:

```go
func (s *DBStore) SaveRawTextReturningID(sourceID, title, content, sourceURL, versionHash string) (int64, error) {
	if versionHash != "" {
		var existingID int64
		err := s.db.QueryRow(`SELECT id FROM policy_raw_texts WHERE version_hash = $1 LIMIT 1`, versionHash).Scan(&existingID)
		if err == nil {
			return existingID, nil
		}
	}
	var id int64
	err := s.db.QueryRow(
		`INSERT INTO policy_raw_texts (source_id, title, content, source_url, version_hash, fetch_time, video_extract_status) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		sourceID, title, content, sourceURL, versionHash, time.Now(), nil,
	).Scan(&id)
	return id, err
}
```

- [ ] **Step 2: Add video_extract_status helpers**

```go
func (s *DBStore) SetVideoExtractStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE policy_raw_texts SET video_extract_status = $1 WHERE id = $2`, status, id)
	return err
}

func (s *DBStore) UpdateRawTextContent(id int64, content string) error {
	_, err := s.db.Exec(`UPDATE policy_raw_texts SET content = $1 WHERE id = $2`, content, id)
	return err
}

func (s *DBStore) MarkExtractedByID(id int64) error {
	_, err := s.db.Exec(`UPDATE policy_raw_texts SET extracted = true WHERE id = $1`, id)
	return err
}

type PendingVideoExtract struct {
	ID        int64
	SourceID  string
	VideoURL  string
	Title     string
	Content   string
}

func (s *DBStore) GetPendingVideoExtracts() ([]PendingVideoExtract, error) {
	rows, err := s.db.Query(`SELECT id, source_id, source_url, COALESCE(title,''), COALESCE(content,'') FROM policy_raw_texts WHERE video_extract_status IN ('pending','processing')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PendingVideoExtract
	for rows.Next() {
		var p PendingVideoExtract
		if err := rows.Scan(&p.ID, &p.SourceID, &p.VideoURL, &p.Title, &p.Content); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}
```

- [ ] **Step 3: Update GetUnprocessedRawTexts to exclude video-pending**

In `store.go`, update the SQL in `GetUnprocessedRawTexts` to add:

```sql
AND (prt.video_extract_status IS NULL OR prt.video_extract_status = 'done')
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./...` in `services/policy-crawler`
Expected: compiles clean

- [ ] **Step 5: Commit**

```bash
git add services/policy-crawler/internal/crawler/store.go
git commit -m "feat: SaveRawTextReturningID, video_extract_status CRUD, pending video extract recovery"
```

---

### Task 5: ASR Provider

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/crawler/asr.go`

- [ ] **Step 1: Implement ASR provider**

Create `asr.go`:

```go
package crawler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ASRConfig struct {
	ID         int64  `json:"id"`
	Provider   string `json:"provider"`
	APIKey     string `json:"api_key"`
	Endpoint   string `json:"endpoint"`
	Language   string `json:"language"`
	SampleRate int    `json:"sample_rate"`
	Enabled    bool   `json:"enabled"`
}

type ASRProvider interface {
	Transcribe(audioPath string) (string, error)
}

type VolcengineASR struct {
	config ASRConfig
	client *http.Client
}

func NewVolcengineASR(cfg ASRConfig) *VolcengineASR {
	return &VolcengineASR{
		config: cfg,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (v *VolcengineASR) Transcribe(audioPath string) (string, error) {
	data, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("read audio: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	reqBody := map[string]interface{}{
		"audio":      encoded,
		"language":   v.config.Language,
		"sample_rate": v.config.SampleRate,
		"format":     detectAudioFormat(audioPath),
	}
	body, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequest("POST", v.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+v.config.APIKey)
	resp, err := v.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ASR API call: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ASR API %d: %s", resp.StatusCode, string(respBody)[:min(len(respBody), 200)])
	}
	var result struct {
		Data struct {
			Text string `json:"text"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse ASR response: %w", err)
	}
	if result.Data.Text == "" {
		return "", fmt.Errorf("ASR returned empty text")
	}
	return result.Data.Text, nil
}

func detectAudioFormat(path string) string {
	if len(path) > 4 {
		ext := path[len(path)-4:]
		switch ext {
		case ".mp3":
			return "mp3"
		case ".wav":
			return "wav"
		case ".m4a":
			return "m4a"
		}
	}
	return "mp3"
}

func NewASRProviderFromConfig(cfg ASRConfig) ASRProvider {
	switch cfg.Provider {
	case "volcengine":
		return NewVolcengineASR(cfg)
	default:
		log.Printf("[asr] unknown provider %q, defaulting to volcengine", cfg.Provider)
		return NewVolcengineASR(cfg)
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...` in `services/policy-crawler`
Expected: compiles clean

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/internal/crawler/asr.go
git commit -m "feat: ASR provider with Volcano Engine implementation"
```

---

### Task 6: Video Extract Worker

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/crawler/video_extract.go`

- [ ] **Step 1: Implement VideoExtractWorker**

Create `video_extract.go`:

```go
package crawler

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type VideoExtractTask struct {
	RawTextID  int64
	SourceID   string
	VideoURL   string
	Title      string
	RetryCount int
}

type VideoExtractWorker struct {
	store     *DBStore
	filter    *RelevanceFilter
	asr       ASRProvider
	queue     chan VideoExtractTask
	workers   int
	stopCh    <-chan struct{}
	tmpDir    string
}

func NewVideoExtractWorker(store *DBStore, filter *RelevanceFilter, asr ASRProvider, workers int) *VideoExtractWorker {
	if workers <= 0 {
		workers = 2
	}
	tmpDir := "/tmp/video-extract"
	os.MkdirAll(tmpDir, 0755)
	return &VideoExtractWorker{
		store:   store,
		filter:  filter,
		asr:     asr,
		queue:   make(chan VideoExtractTask, 100),
		workers: workers,
		tmpDir:  tmpDir,
	}
}

func (w *VideoExtractWorker) Queue() chan VideoExtractTask { return w.queue }

func (w *VideoExtractWorker) Start() {
	for i := 0; i < w.workers; i++ {
		go w.run(i)
	}
	log.Printf("[video-extract] started %d workers", w.workers)
}

func (w *VideoExtractWorker) run(id int) {
	for {
		select {
		case task := <-w.queue:
			w.process(task)
		case <-w.stopCh:
			return
		}
	}
}

func (w *VideoExtractWorker) process(task VideoExtractTask) {
	log.Printf("[video-extract] processing raw_text=%d url=%s retry=%d", task.RawTextID, task.VideoURL, task.RetryCount)
	w.store.SetVideoExtractStatus(task.RawTextID, "processing")

	transcript, err := w.extractTranscript(task.VideoURL)
	if err != nil {
		log.Printf("[video-extract] transcript error for %s: %v", task.VideoURL, err)
		w.handleFailure(task, err)
		return
	}

	enriched := fmt.Sprintf("銆愭爣棰樸�?s\n銆愯棰戣浆褰曘�?s", task.Title, transcript)
	if task.RetryCount == 0 {
		desc := ""
		if parts := strings.SplitN(task.Title, "\n", 2); len(parts) == 2 {
			desc = parts[1]
		}
		if desc != "" {
			enriched = fmt.Sprintf("銆愭爣棰樸�?s\n銆愭弿杩般�?s\n銆愯棰戣浆褰曘�?s", parts[0], desc, transcript)
		}
	}

	score, matched := w.filter.Score(enriched, task.SourceID, "level2")
	threshold := w.filter.MinScore(task.SourceID, "level2")
	if score < threshold {
		log.Printf("[video-extract] discarded raw_text=%d score=%d<threshold=%d matched=%v", task.RawTextID, score, threshold, matched)
		w.store.UpdateRawTextContent(task.RawTextID, enriched)
		w.store.SetVideoExtractStatus(task.RawTextID, "discarded")
		w.store.MarkExtractedByID(task.RawTextID)
		return
	}

	w.store.UpdateRawTextContent(task.RawTextID, enriched)
	w.store.SetVideoExtractStatus(task.RawTextID, "done")
	log.Printf("[video-extract] done raw_text=%d enriched=%d bytes score=%d", task.RawTextID, len(enriched), score)
}

func (w *VideoExtractWorker) extractTranscript(videoURL string) (string, error) {
	tmpBase := filepath.Join(w.tmpDir, fmt.Sprintf("%d", time.Now().UnixNano()))
	defer w.cleanup(tmpBase)

	subtitle, err := w.extractSubtitle(videoURL, tmpBase)
	if err == nil && subtitle != "" {
		log.Printf("[video-extract] got subtitle for %s (%d chars)", videoURL, len(subtitle))
		return subtitle, nil
	}
	log.Printf("[video-extract] no subtitle for %s, falling back to ASR: %v", videoURL, err)
	return w.extractViaASR(videoURL, tmpBase)
}

func (w *VideoExtractWorker) extractSubtitle(videoURL, tmpBase string) (string, error) {
	outPath := tmpBase + ".vtt"
	cmd := exec.Command("yt-dlp", "--write-sub", "--sub-lang", "zh,zh-Hans,zh-CN", "--skip-download",
		"--convert-subs", "srt", "--output", tmpBase, videoURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp subtitle: %w: %s", err, string(out))
	}
	srtPath := tmpBase + ".srt"
	data, err := os.ReadFile(srtPath)
	if err != nil {
		data, err = os.ReadFile(outPath)
		if err != nil {
			return "", fmt.Errorf("no subtitle file found")
		}
	}
	text := cleanSubtitle(string(data))
	if len(text) < 20 {
		return "", fmt.Errorf("subtitle too short: %d chars", len(text))
	}
	return text, nil
}

func (w *VideoExtractWorker) extractViaASR(videoURL, tmpBase string) (string, error) {
	audioPath := tmpBase + ".mp3"
	cmd := exec.Command("yt-dlp", "-x", "--audio-format", "mp3", "--audio-quality", "5",
		"--output", tmpBase, "--ratelimit", "1M", videoURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp download: %w: %s", err, string(out))
	}
	if _, err := os.Stat(audioPath); err != nil {
		return "", fmt.Errorf("audio file not found at %s", audioPath)
	}
	if w.asr == nil {
		return "", fmt.Errorf("ASR provider not configured")
	}
	return w.asr.Transcribe(audioPath)
}

func (w *VideoExtractWorker) handleFailure(task VideoExtractTask, err error) {
	if task.RetryCount < 3 {
		task.RetryCount++
		log.Printf("[video-extract] retrying raw_text=%d (attempt %d)", task.RawTextID, task.RetryCount)
		time.AfterFunc(time.Duration(task.RetryCount)*10*time.Second, func() {
			w.queue <- task
		})
	} else {
		w.store.SetVideoExtractStatus(task.RawTextID, "failed")
		log.Printf("[video-extract] giving up raw_text=%d after %d retries: %v", task.RawTextID, task.RetryCount, err)
	}
}

func (w *VideoExtractWorker) cleanup(base string) {
	patterns := []string{base + ".*", base}
	for _, p := range patterns {
		matches, _ := filepath.Glob(p)
		for _, f := range matches {
			os.Remove(f)
		}
	}
}

func cleanSubtitle(srt string) string {
	var lines []string
	for _, line := range strings.Split(srt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "-->") {
			continue
		}
		if len(line) > 0 && line[0] >= '0' && line[0] <= '9' && !strings.Contains(line, " ") && len(line) <= 5 {
			continue
		}
		line = strings.ReplaceAll(line, "<b>", "")
		line = strings.ReplaceAll(line, "</b>", "")
		line = strings.ReplaceAll(line, "<i>", "")
		line = strings.ReplaceAll(line, "</i>", "")
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, " ")
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...` in `services/policy-crawler`
Expected: compiles clean

- [ ] **Step 3: Commit**

```bash
git add services/policy-crawler/internal/crawler/video_extract.go
git commit -m "feat: VideoExtractWorker with subtitle-first, ASR-fallback pipeline"
```

---

### Task 7: Douyin Crawler Redesign

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/crawler/douyin_crawler.go`

- [ ] **Step 1: Add relevance filter to DouyinCrawler struct**

Update struct and constructor:

```go
type DouyinCrawler struct {
	config    SourceConfig
	renderer  PageRenderer
	filter    *RelevanceFilter
	maxItems  int
	processed map[string]bool
}

func NewDouyinCrawler(cfg SourceConfig, filter *RelevanceFilter) *DouyinCrawler {
	return &DouyinCrawler{
		config:    cfg,
		filter:    filter,
		maxItems:  50,
		processed: make(map[string]bool),
	}
}
```

- [ ] **Step 2: Rewrite discoverVideosFromUserPage with strict author filtering**

Replace the `discoverVideosFromUserPage` method:

```go
func (d *DouyinCrawler) discoverVideosFromUserPage(userURL string) ([]string, error) {
	cleanURL := stripQueryParams(userURL)
	log.Printf("[douyin] discovering own videos from user page: %s", cleanURL)

	if d.renderer != nil {
		worksURL := cleanURL + "?showTab=works"
		html, err := d.renderer.RenderWithVirtualTime(worksURL, 30000)
		if err == nil {
			videoURLs := extractDouyinVideoURLs(html)
			if len(videoURLs) > 0 {
				log.Printf("[douyin] discovered %d videos via Chrome (works tab) from %s", len(videoURLs), cleanURL)
				return videoURLs, nil
			}
		}
		log.Printf("[douyin] Chrome render failed for works tab %s: %v", cleanURL, err)
	}

	html, err := httpFetch(cleanURL)
	if err != nil {
		return nil, fmt.Errorf("discover videos from %s: %w", cleanURL, err)
	}
	videoURLs := extractDouyinVideoURLs(html)
	if len(videoURLs) == 0 {
		videoURLs = extractDouyinVideoLinks(html)
	}
	if len(videoURLs) == 0 {
		return nil, fmt.Errorf("no videos discovered from %s", cleanURL)
	}
	log.Printf("[douyin] discovered %d videos via HTTP from %s", len(videoURLs), cleanURL)
	return videoURLs, nil
}
```

- [ ] **Step 3: Update fetchVideoPage to add relevance filter and set NeedsVideoExtract**

Replace the `fetchVideoPage` method. After author check and content extraction, add relevance scoring and set `NeedsVideoExtract`:

```go
func (d *DouyinCrawler) fetchVideoPage(videoURL string) (*CrawlResult, error) {
	var html string
	var usedChrome bool
	var err error

	if d.renderer != nil {
		html, err = d.renderer.RenderWithVirtualTime(videoURL, 15000)
		usedChrome = err == nil
		if err != nil {
			log.Printf("[douyin] Chrome render failed for %s, falling back to HTTP: %v", videoURL, err)
		}
	}
	if html == "" {
		html, err = httpFetch(videoURL)
		if err != nil {
			return nil, fmt.Errorf("http fetch: %w", err)
		}
	}

	expectedNickname := extractNicknameFromSourceID(d.config.SourceID)
	if expectedNickname != "" {
		actualAuthor := extractDouyinAuthorNickname(html)
		if strings.Contains(html, "@"+expectedNickname) {
		} else if actualAuthor != "" && !strings.Contains(actualAuthor, expectedNickname) && !strings.Contains(expectedNickname, actualAuthor) {
			log.Printf("[douyin] skipping video %s: author %q does not match expected %q", videoURL, actualAuthor, expectedNickname)
			return nil, nil
		} else if usedChrome {
			log.Printf("[douyin] skipping video %s: @%s not found in Chrome-rendered page", videoURL, expectedNickname)
			return nil, nil
		}
	}

	title := extractDouyinTitle(html)
	desc := extractDouyinDesc(html)
	content := title
	if desc != "" {
		content = title + "\n" + desc
	}
	if content == "" {
		content = extractDouyinTextFromHTML(html)
	}
	if content == "" {
		return nil, nil
	}

	if d.filter != nil {
		score, matched := d.filter.Score(title+" "+desc, d.config.SourceID, "douyin")
		threshold := d.filter.MinScore(d.config.SourceID, "level1")
		if score < threshold {
			log.Printf("[douyin] filtered out %s: relevance score %d < threshold %d (matched: %v)", videoURL, score, threshold, matched)
			return nil, nil
		}
		log.Printf("[douyin] passed relevance filter for %s: score=%d matched=%v", videoURL, score, matched)
	}

	hash := sha256.Sum256([]byte(videoURL))
	return &CrawlResult{
		SourceID:          d.config.SourceID,
		SourceLevel:       d.config.SourceLevel,
		RawText:           content,
		Title:             title,
		SourceURL:         videoURL,
		FetchedAt:         time.Now(),
		VersionHash:       fmt.Sprintf("%x", hash),
		VideoURL:          videoURL,
		NeedsVideoExtract: true,
		ContentType:       "video-meta",
	}, nil
}
```

- [ ] **Step 4: Update all NewDouyinCrawler calls**

In `manager.go`, update both calls to `NewDouyinCrawler(cfg)` to `NewDouyinCrawler(cfg, m.filter)` (the filter field will be added in Task 9). For now, pass `nil` as placeholder and fix in Task 9.

Actually, update the calls to accept the filter parameter now:

In `manager.go` `Init()`:
```go
case "douyin":
    dc := NewDouyinCrawler(cfg, m.filter)
```

In `manager.go` `loadAndRegisterSource()`:
```go
case "douyin":
    dc := NewDouyinCrawler(cfg, m.filter)
```

- [ ] **Step 5: Verify compilation**

Run: `go build ./...` in `services/policy-crawler`
Expected: compiles (may need `m.filter` field on CrawlerManager 鈥?add `filter *RelevanceFilter` to struct, initialized to nil for now, will be wired in Task 9)

- [ ] **Step 6: Commit**

```bash
git add services/policy-crawler/internal/crawler/douyin_crawler.go services/policy-crawler/internal/crawler/manager.go
git commit -m "feat: douyin crawler strict author filter, relevance pre-filter, NeedsVideoExtract"
```

---

### Task 8: WeChat Mixed Mode + Anti-Bot

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/crawler/wechat_crawler.go`

- [ ] **Step 1: Add relevance filter + __biz discovery**

Update struct:

```go
type WeChatCrawler struct {
	config    SourceConfig
	renderer  PageRenderer
	filter    *RelevanceFilter
	maxItems  int
	processed map[string]bool
}

func NewWeChatCrawler(cfg SourceConfig, filter *RelevanceFilter) *WeChatCrawler {
	return &WeChatCrawler{
		config:    cfg,
		filter:    filter,
		maxItems:  20,
		processed: make(map[string]bool),
	}
}
```

- [ ] **Step 2: Add __biz account-based discovery**

Add new method after `discoverArticles`:

```go
func (w *WeChatCrawler) discoverByBiz(bizID string) []string {
	encoded := fmt.Sprintf("__biz=%s", bizID)
	searchURL := fmt.Sprintf("https://www.bing.com/search?q=site%%3Amp.weixin.qq.com%%20%s", encoded)
	log.Printf("[wechat] discovering articles by biz: %s", bizID)
	html, err := w.renderer.RenderWithVirtualTime(searchURL, 20000)
	if err != nil {
		log.Printf("[wechat] biz search render error: %v", err)
		return nil
	}
	return w.extractWeChatURLs(html, "BizSearch", bizID)
}
```

- [ ] **Step 3: Update Fetch() with mixed mode + relevance filter + anti-bot detection**

Update `Fetch()` to:
1. After gathering article URLs from search + __biz discovery, deduplicate
2. Fetch each article, check for anti-bot page, run relevance filter
3. If passes, return CrawlResult WITHOUT NeedsVideoExtract (WeChat has full text already)

Key changes to `Fetch()`:
- After `discoverArticles(kw)`, also try `discoverByBiz` if `__biz` found in source URLs
- In `fetchArticle()`, add anti-bot check before returning content
- After getting content, run `filter.Score(title+" "+content, ...)` with level1 threshold
- If passes, return `CrawlResult{NeedsVideoExtract: false, ContentType: "text"}`

- [ ] **Step 4: Update NewWeChatCrawler calls in manager.go**

Same pattern as Douyin: `NewWeChatCrawler(cfg, m.filter)`

- [ ] **Step 5: Verify compilation**

Run: `go build ./...` in `services/policy-crawler`

- [ ] **Step 6: Commit**

```bash
git add services/policy-crawler/internal/crawler/wechat_crawler.go services/policy-crawler/internal/crawler/manager.go
git commit -m "feat: wechat mixed mode, __biz discovery, anti-bot detection, relevance filter"
```

---

### Task 9: Manager Pipeline Integration

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/crawler/manager.go`

- [ ] **Step 1: Add VideoExtractWorker + RelevanceFilter to CrawlerManager**

Update struct:

```go
type CrawlerManager struct {
	store       Store
	dbStore     *DBStore
	claimDB     ClaimDB
	crawlers    []Source
	sourceCfgs  map[string]SourceConfig
	stopCh      chan struct{}
	renderer    PageRenderer
	filter      *RelevanceFilter
	videoWorker *VideoExtractWorker
}
```

- [ ] **Step 2: Update NewCrawlerManager to accept DBStore**

```go
func NewCrawlerManager(store Store, dbStore *DBStore, claimDB ClaimDB) *CrawlerManager {
	return &CrawlerManager{
		store:      store,
		dbStore:    dbStore,
		claimDB:    claimDB,
		sourceCfgs: make(map[string]SourceConfig),
		stopCh:     make(chan struct{}),
	}
}
```

- [ ] **Step 3: Add InitFilterAndWorker method**

```go
func (m *CrawlerManager) InitFilterAndWorker(db *sql.DB, asrCfg ASRConfig) {
	m.filter = NewRelevanceFilter(nil)
	m.filter.LoadFromDB(db)
	go m.filter.StartReloadLoop(5*time.Minute, m.stopCh)

	var asr ASRProvider
	if asrCfg.Enabled {
		asr = NewASRProviderFromConfig(asrCfg)
	}
	m.videoWorker = NewVideoExtractWorker(m.dbStore, m.filter, asr, 2)
	m.videoWorker.Start()
	m.recoverPendingVideoExtracts()
}

func (m *CrawlerManager) recoverPendingVideoExtracts() {
	if m.dbStore == nil {
		return
	}
	tasks, err := m.dbStore.GetPendingVideoExtracts()
	if err != nil {
		log.Printf("[video-extract] recovery query failed: %v", err)
		return
	}
	for _, t := range tasks {
		m.videoWorker.Queue() <- VideoExtractTask{
			RawTextID: t.ID,
			SourceID:  t.SourceID,
			VideoURL:  t.VideoURL,
			Title:     t.Title,
		}
	}
	if len(tasks) > 0 {
		log.Printf("[video-extract] recovered %d pending tasks", len(tasks))
	}
}
```

- [ ] **Step 4: Update crawlAndProcess for video dispatch**

In `crawlAndProcess()`, after `SaveRawText`, change the handling:

```go
func (m *CrawlerManager) crawlAndProcess(s Source) {
	log.Printf("[crawler] fetching source %s (%s)", s.SourceID(), s.SourceLevel())
	results, err := s.Fetch()
	if err != nil {
		log.Printf("[crawler] fetch error for %s: %v", s.SourceID(), err)
		m.store.SaveCrawlLog(s.SourceID(), false, err.Error())
		return
	}
	if len(results) == 0 {
		m.store.SaveCrawlLogWithDetails(s.SourceID(), true, "no new content", "", "")
		return
	}

	for _, result := range results {
		if result == nil {
			continue
		}

		if result.NeedsVideoExtract && m.videoWorker != nil && m.dbStore != nil {
			rawTextID, err := m.dbStore.SaveRawTextReturningID(result.SourceID, result.Title, result.RawText, result.SourceURL, result.VersionHash)
			if err != nil {
				log.Printf("[crawler] save raw text error for %s: %v", s.SourceID(), err)
				continue
			}
			m.dbStore.SetVideoExtractStatus(rawTextID, "pending")
			m.videoWorker.Queue() <- VideoExtractTask{
				RawTextID: rawTextID,
				SourceID:  result.SourceID,
				VideoURL:  result.VideoURL,
				Title:     result.Title,
			}
			m.store.SaveCrawlLogWithDetails(s.SourceID(), true, "pending video extraction", "", truncateSummary(result.Title, 120))
			continue
		}

		m.store.SaveRawText(s.SourceID(), result.Title, result.RawText, result.SourceURL, result.VersionHash)
		log.Printf("[crawler] fetched %d bytes from %s (%s)", len(result.RawText), s.SourceID(), result.SourceURL)

		isHTML := !strings.Contains(result.RawText, "鏀跨瓥ID:")
		if !isHTML && len(result.RawText) > 0 {
			first := strings.TrimSpace(result.RawText)
			isHTML = first[0] == '<' || strings.Contains(first, "<body") || strings.Contains(first, "<html")
		}
		if isHTML {
			summary := truncateSummary(result.Title, 120)
			if summary == "" {
				summary = truncateSummary(result.SourceURL, 120)
			}
			m.store.SaveCrawlLogWithDetails(s.SourceID(), true, "HTML stored, awaiting LLM extraction", "", summary)
			continue
		}

		parsed, conditions, docs, err := parser.ParseStructuredText(result.RawText)
		if err != nil {
			log.Printf("[crawler] parse error for %s: %v", s.SourceID(), err)
			m.store.SaveCrawlLog(s.SourceID(), false, "parse error: "+err.Error())
			continue
		}

		condJSON, _ := json.Marshal(conditions)
		docJSON, _ := json.Marshal(docs)
		confidence := m.calculateConfidence(s.SourceLevel(), parsed)
		status := decideStatus(confidence)

		claim := &models.PolicyClaim{
			ClaimID:           fmt.Sprintf("CRAWL-%d", time.Now().UnixNano()),
			PolicyID:          parsed.PolicyID,
			RegionCode:        parsed.RegionCode,
			PolicyType:        parsed.PolicyType,
			TargetGroupTags:   parsed.TargetGroups,
			SubsidyCalcMethod: parsed.SubsidyCalcMethod,
			SubsidyAmountMin:  parsed.AmountMin,
			SubsidyAmountMax:  parsed.AmountMax,
			SubsidyDuration:   parsed.SubsidyDuration,
			EffectiveDate:     parsed.EffectiveDate,
			ExpireDate:        parsed.ExpireDate,
			ConfidenceScore:   confidence,
			Status:            status,
			VersionNumber:     1,
			Conditions:        condJSON,
			RequiredDocuments: docJSON,
			SourceID:          s.SourceID(),
			SourceURL:         result.SourceURL,
			SourceName:        "",
		}

		if err := m.claimDB.Ingest(claim); err != nil {
			log.Printf("[crawler] store error for %s: %v", s.SourceID(), err)
			m.store.SaveCrawlLog(s.SourceID(), false, "store error: "+err.Error())
			continue
		}
		m.store.SaveCrawlLogWithDetails(s.SourceID(), true, "", claim.ClaimID, truncateSummary(claim.SubsidyCalcMethod, 120))
		log.Printf("[crawler] stored claim %s (confidence=%.2f, status=%s)", claim.ClaimID, confidence, status)
	}
}
```

- [ ] **Step 5: Update main.go to wire everything**

In `cmd/main.go`:
1. Load ASR config from DB
2. Call `manager.InitFilterAndWorker(db, asrCfg)`
3. Update `NewCrawlerManager(store, dbStore, claimDB)` call

- [ ] **Step 6: Verify compilation**

Run: `go build ./...` in `services/policy-crawler`

- [ ] **Step 7: Commit**

```bash
git add services/policy-crawler/internal/crawler/manager.go services/policy-crawler/cmd/main.go
git commit -m "feat: manager pipeline integration with video extract worker + recovery"
```

---

### Task 10: LLM Backup Config

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/admin/admin_llm.go`
- Modify: `services/policy-crawler/internal/crawler/store.go` (SaveLLMConfig/GetLLMConfig SQL)
- Modify: `services/policy-crawler/internal/llm/llm.go`

- [ ] **Step 1: Add backup fields to LLMConfig**

In `admin_llm.go`:

```go
type LLMConfig struct {
	Provider            string `json:"provider"`
	APIKey              string `json:"api_key"`
	Endpoint            string `json:"endpoint"`
	ModelName           string `json:"model_name"`
	MaxTokens           int    `json:"max_tokens"`
	Enabled             bool   `json:"enabled"`
	EmbeddingModel      string `json:"embedding_model"`
	EmbeddingDimensions int    `json:"embedding_dimensions"`
	EmbeddingAPIKey     string `json:"embedding_api_key"`
	EmbeddingEndpoint   string `json:"embedding_endpoint"`
	BackupProvider      string `json:"backup_provider"`
	BackupAPIKey        string `json:"backup_api_key"`
	BackupEndpoint      string `json:"backup_endpoint"`
	BackupModelName     string `json:"backup_model_name"`
}
```

- [ ] **Step 2: Update store.go GetLLMConfig/SaveLLMConfig SQL**

Add `backup_provider, backup_api_key, backup_endpoint, backup_model_name` to SELECT and UPDATE queries.

- [ ] **Step 3: Add backup fallback to llm.Client**

In `llm.go`, add `BackupConfig` field:

```go
type Client struct {
	config       Config
	backup       *Config
	http         *http.Client
}

func NewClientWithBackup(cfg Config, backup *Config) *Client {
	return &Client{
		config: cfg,
		backup: backup,
		http:   &http.Client{Timeout: 60 * time.Second},
	}
}
```

Update `Chat()` to try backup on failure:

```go
func (c *Client) Chat(systemPrompt, userContent string) (string, error) {
	resp, err := c.chat(systemPrompt, userContent, c.config)
	if err == nil {
		return resp, nil
	}
	if c.backup != nil && c.backup.APIKey != "" {
		log.Printf("[llm] primary failed: %v, trying backup %s", err, c.backup.ModelName)
		resp2, err2 := c.chat(systemPrompt, userContent, *c.backup)
		if err2 == nil {
			return resp2, nil
		}
		log.Printf("[llm] backup also failed: %v", err2)
	}
	return "", err
}

func (c *Client) chat(systemPrompt, userContent string, cfg Config) (string, error) {
	switch cfg.Provider {
	case ProviderAliBailian:
		return c.chatBailianWithConfig(systemPrompt, userContent, cfg)
	default:
		return c.chatOpenAIWithConfig(systemPrompt, userContent, cfg)
	}
}
```

Refactor existing `chatOpenAI` and `chatBailian` to accept a `Config` parameter.

- [ ] **Step 4: Update main.go to wire backup config**

- [ ] **Step 5: Verify compilation**

Run: `go build ./...`

- [ ] **Step 6: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_llm.go services/policy-crawler/internal/crawler/store.go services/policy-crawler/internal/llm/llm.go services/policy-crawler/cmd/main.go
git commit -m "feat: LLM backup config with primary/fallback"
```

---

### Task 11: Admin Relevance Rules API

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/admin/admin_relevance.go`

- [ ] **Step 1: Implement handlers**

Create `admin_relevance.go` with handlers for:
- `RelevanceRulesListHandler(store)` 鈥?GET, list rules with optional category/scope filter
- `RelevanceRulesCreateHandler(store)` 鈥?POST, create rule
- `RelevanceRulesUpdateHandler(store)` 鈥?PUT, update rule (weight, enabled, scope)
- `RelevanceRulesDeleteHandler(store)` 鈥?DELETE, delete rule
- `RelevanceThresholdGetHandler(store)` 鈥?GET `/admin/relevance/thresholds/{source_id}`
- `RelevanceThresholdSetHandler(store)` 鈥?PUT `/admin/relevance/thresholds/{source_id}`
- `RelevanceTestHandler(filter)` 鈥?POST, test text score
- `RelevanceBulkImportHandler(store)` 鈥?POST, bulk import rules

All using `*sql.DB` as the store interface (or a thin interface wrapping the DB).

- [ ] **Step 2: Register routes in main.go**

Add routes:
```go
mux.Handle("/admin/relevance/rules", adminAuth(admin.RelevanceRulesListHandler(db)))
mux.Handle("/admin/relevance/rules/create", adminAuth(admin.RelevanceRulesCreateHandler(db)))
mux.Handle("/admin/relevance/rules/update", adminAuth(admin.RelevanceRulesUpdateHandler(db)))
mux.Handle("/admin/relevance/rules/delete", adminAuth(admin.RelevanceRulesDeleteHandler(db)))
mux.Handle("/admin/relevance/thresholds/", adminAuth(admin.RelevanceThresholdHandler(db)))
mux.Handle("/admin/relevance/test", adminAuth(admin.RelevanceTestHandler(filter)))
mux.Handle("/admin/relevance/bulk-import", adminAuth(admin.RelevanceBulkImportHandler(db)))
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_relevance.go services/policy-crawler/cmd/main.go
git commit -m "feat: admin relevance rules API (CRUD + test + bulk import)"
```

---

### Task 12: Admin ASR Config API

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Create: `services/policy-crawler/internal/admin/admin_asr.go`

- [ ] **Step 1: Implement handlers**

Create `admin_asr.go` with:
- `ASRConfigGetHandler(db)` 鈥?GET current config
- `ASRConfigSaveHandler(db)` 鈥?POST update config

- [ ] **Step 2: Register routes in main.go**

- [ ] **Step 3: Verify compilation**

- [ ] **Step 4: Commit**

```bash
git add services/policy-crawler/internal/admin/admin_asr.go services/policy-crawler/cmd/main.go
git commit -m "feat: admin ASR config API"
```

---

### Task 13: Admin UI 鈥?Relevance Rules + ASR + LLM Backup

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/internal/admin/admin_page.go`

- [ ] **Step 1: Add "鐩稿叧鎬ц鍒? tab**

Add a new tab to the admin navigation that renders:
1. Rules table grouped by category
2. Add rule form (keyword + category + weight + scope)
3. Bulk import (textarea for JSON array)
4. Test panel (input text 鈫?show score + matched keywords)
5. Per-source threshold config section

- [ ] **Step 2: Update LLM config section for backup fields**

Add 4 fields: backup_provider, backup_api_key, backup_endpoint, backup_model_name

- [ ] **Step 3: Add ASR config section**

Add ASR config card: provider, api_key, endpoint, enabled toggle

- [ ] **Step 4: Verify and commit**

```bash
git add services/policy-crawler/internal/admin/admin_page.go
git commit -m "feat: admin UI for relevance rules, ASR config, LLM backup"
```

---

### Task 14: Dockerfile + Build + Deploy

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/policy-crawler/Dockerfile`

- [ ] **Step 1: Update Dockerfile**

```dockerfile
FROM chromium-test:latest
RUN apk add --no-cache python3 py3-pip ffmpeg && pip3 install --no-cache-dir yt-dlp
RUN adduser -D -g '' appuser
RUN mkdir -p /data/policies /tmp/video-extract && chown appuser:appuser /data/policies /tmp/video-extract
COPY bin/policy-crawler /policy-crawler
USER appuser
EXPOSE 39403
ENV CHROME_ENABLED=true
ENTRYPOINT ["/policy-crawler"]
```

- [ ] **Step 2: Cross-compile, rebuild image, deploy**

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o policy-crawler ./cmd/main.go
Copy-Item policy-crawler bin/policy-crawler -Force
docker build --no-cache -t nsi-policy-crawler:latest .
docker compose up -d db-migrate
Start-Sleep 3
docker compose up -d policy-crawler
```

- [ ] **Step 3: Verify**

```powershell
Start-Sleep 10
docker logs nsi-policy-crawler 2>&1 | Select-Object -Last 30
```

Expected: "loaded N rules", "started 2 workers", crawling with relevance filtering

- [ ] **Step 4: Commit**

```bash
git add services/policy-crawler/Dockerfile
git commit -m "feat: Dockerfile with yt-dlp, ffmpeg for video extraction"
```

---

### Task 15: End-to-End Verification

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- [ ] **Step 1: Verify migrations applied**

```powershell
docker exec nsi-postgres psql -U postgres -d nsi_crawler -c "SELECT COUNT(*) FROM relevance_rules WHERE enabled;"
docker exec nsi-postgres psql -U postgres -d nsi_crawler -c "SELECT column_name FROM information_schema.columns WHERE table_name='policy_raw_texts' AND column_name='video_extract_status';"
docker exec nsi-postgres psql -U postgres -d nsi_crawler -c "SELECT * FROM asr_configs LIMIT 1;"
```

Expected: 44 rules, video_extract_status column exists, asr_configs has 1 row

- [ ] **Step 2: Verify admin API**

```powershell
docker exec nsi-policy-crawler wget -qO- http://localhost:39403/healthz
```

Expected: `{"status":"ok"}`

- [ ] **Step 3: Verify relevance filter is loaded**

```powershell
docker logs nsi-policy-crawler 2>&1 | Select-String "relevance"
```

Expected: "loaded 44 rules" log line

- [ ] **Step 4: Trigger a crawl and verify**

```powershell
docker exec nsi-policy-crawler wget -qO- --post-data '{"source_id":"DOUYIN-涓婃捣绀句繚瑙勫垝鎵綱icky"}' --header='Content-Type: application/json' http://admin:changeme@localhost:39403/admin/sources/crawl
```

Then check logs for relevance filter activity and video extract queue.

- [ ] **Step 5: Final commit (if any fixes needed)**
