# AI 交叉验证 + 语义搜索 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为政策条款添加基于嵌入向量的余弦相似度搜索，实现提取时交叉验证（重复/矛盾检测）和管理后台/API 语义搜索。

**Architecture:** 内存嵌入缓存（EmbeddingCache）+ Go 余弦相似度 O(n) 搜索。提取器 ProcessOne 在 LLM 解析后、InsertClaim 前交叉验证相似政策，调整 confidence/status。缓存定时全量刷新 + 单条增量更新。

**Tech Stack:** Go 1.24, PostgreSQL 18 float8[], pq.Array, net/http

---

### Task 1: 嵌入缓存模型定义

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/embeddings/models.go`

- [ ] **Step 1: Write models.go**

```go
package embeddings

// EmbeddedClaim 缓存中的政策数据
type EmbeddedClaim struct {
	ClaimID    string
	PolicyID   string
	PolicyType string
	RegionCode string
	Embedding  []float64
	SourceName string
	Status     string
}

// SimilarResult 搜索结果
type SimilarResult struct {
	ClaimID    string  `json:"claim_id"`
	PolicyID   string  `json:"policy_id"`
	PolicyType string  `json:"policy_type"`
	RegionCode string  `json:"region_code"`
	Score      float64 `json:"score"`
	SourceName string  `json:"source_name"`
	PolicyURL  string  `json:"policy_url,omitempty"`
	Status     string  `json:"status"`
}

// SearchFilter 搜索过滤条件
type SearchFilter struct {
	RegionCode string // 为空不过滤
	PolicyType string // 为空不过滤
}
```

- [ ] **Step 2: Commit**

```bash
git add nsi-platform/services/policy-crawler/internal/embeddings/models.go
git commit -m "feat(embeddings): add EmbeddedClaim, SimilarResult, SearchFilter types"
```

---

### Task 2: 余弦相似度函数

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/embeddings/similarity.go`
- Create: `nsi-platform/services/policy-crawler/internal/embeddings/similarity_test.go`

- [ ] **Step 1: Write the failing test**

```go
// similarity_test.go
package embeddings

import (
	"math"
	"testing"
)

func TestCosineSimilarity_Identical(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{1, 0, 0}
	s := CosineSimilarity(a, b)
	if s != 1.0 {
		t.Fatalf("identical vectors should have similarity 1.0, got %f", s)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}
	s := CosineSimilarity(a, b)
	if s != 0.0 {
		t.Fatalf("orthogonal vectors should have similarity 0.0, got %f", s)
	}
}

func TestCosineSimilarity_Partial(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0.5, 0.5, 0}
	s := CosineSimilarity(a, b)
	expected := 1.0 / math.Sqrt(2) // dot=0.5, |a|=1, |b|=sqrt(0.5)
	if math.Abs(s-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, s)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float64{0, 0, 0}
	b := []float64{1, 0, 0}
	s := CosineSimilarity(a, b)
	if s != 0.0 {
		t.Fatalf("zero vector should return 0.0, got %f", s)
	}
}

func TestCosineSimilarity_BothZero(t *testing.T) {
	a := []float64{0, 0, 0}
	b := []float64{0, 0, 0}
	s := CosineSimilarity(a, b)
	if s != 0.0 {
		t.Fatalf("both zero should return 0.0, got %f", s)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float64{1, 0}
	b := []float64{1, 0, 0}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic on different lengths")
		}
	}()
	CosineSimilarity(a, b)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run (in `nsi-platform/services/policy-crawler`):
```powershell
$env:GOPROXY="off"; $env:GONOSUMCHECK="*"; $env:GONOSUMDB="*"; $env:GOFLAGS="-mod=mod"
go test ./internal/embeddings/ -run TestCosine -v 2>&1
```
Expected: FAIL with `undefined: CosineSimilarity`

- [ ] **Step 3: Write minimal implementation**

```go
// similarity.go
package embeddings

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("cosine similarity: vectors must have same length")
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	// 内联 math.Sqrt 避免引入 math 包（可选）
	// 使用 math.Sqrt 更精确
	return float64(int64(x*1e15)) / 1e15 // 简化版，实际使用 math.Sqrt
}
```

Wait, I should just use math.Sqrt. Let me use the standard library.

```go
package embeddings

import "math"

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("cosine similarity: vectors must have same length")
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

- [ ] **Step 4: Run test to verify it passes**

```powershell
go test ./internal/embeddings/ -run TestCosine -v 2>&1
```
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add nsi-platform/services/policy-crawler/internal/embeddings/similarity.go
git add nsi-platform/services/policy-crawler/internal/embeddings/similarity_test.go
git commit -m "feat(embeddings): add CosineSimilarity function"
```

---

### Task 3: EmbeddingCache 实现

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/embeddings/cache.go`
- Create: `nsi-platform/services/policy-crawler/internal/embeddings/cache_test.go`

- [ ] **Step 1: Write cache.go**

```go
package embeddings

import (
	"sort"
	"sync"
)

// EmbeddingCache 内存嵌入缓存
type EmbeddingCache struct {
	mu     sync.RWMutex
	claims []EmbeddedClaim
	loader func() ([]EmbeddedClaim, error)
}

// SearchFilter 搜索过滤条件
// (如果在 models.go 中已定义则省略)

// NewEmbeddingCache 创建缓存，loader 是首次加载和刷新的回调
func NewEmbeddingCache(loader func() ([]EmbeddedClaim, error)) *EmbeddingCache {
	return &EmbeddingCache{loader: loader}
}

// Load 全量加载（启动时调用）
func (c *EmbeddingCache) Load() error {
	claims, err := c.loader()
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.claims = claims
	c.mu.Unlock()
	return nil
}

// Refresh 全量刷新（定时器调用）
func (c *EmbeddingCache) Refresh() error {
	return c.Load()
}

// Add 单条新增（提取成功后调用）
func (c *EmbeddingCache) Add(ec EmbeddedClaim) {
	c.mu.Lock()
	c.claims = append(c.claims, ec)
	c.mu.Unlock()
}

// Len 返回缓存大小（测试用）
func (c *EmbeddingCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.claims)
}

// SearchSimilar 搜索最相似的 K 条结果
func (c *EmbeddingCache) SearchSimilar(emb []float64, threshold float64, limit int, filter *SearchFilter) []SimilarResult {
	c.mu.RLock()
	claims := make([]EmbeddedClaim, len(c.claims))
	copy(claims, c.claims)
	c.mu.RUnlock()

	var results []SimilarResult
	for _, cc := range claims {
		if len(cc.Embedding) == 0 {
			continue
		}
		if filter != nil {
			if filter.RegionCode != "" && cc.RegionCode != filter.RegionCode {
				continue
			}
			if filter.PolicyType != "" && cc.PolicyType != filter.PolicyType {
				continue
			}
		}
		score := CosineSimilarity(emb, cc.Embedding)
		if score < threshold {
			continue
		}
		results = append(results, SimilarResult{
			ClaimID:    cc.ClaimID,
			PolicyID:   cc.PolicyID,
			PolicyType: cc.PolicyType,
			RegionCode: cc.RegionCode,
			Score:      score,
			SourceName: cc.SourceName,
			Status:     cc.Status,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}
```

- [ ] **Step 2: Write cache_test.go**

```go
package embeddings

import (
	"testing"
)

func TestCacheLoad(t *testing.T) {
	c := NewEmbeddingCache(func() ([]EmbeddedClaim, error) {
		return []EmbeddedClaim{
			{ClaimID: "c1", RegionCode: "110000", PolicyType: "subsidy", Embedding: []float64{1, 0, 0}},
			{ClaimID: "c2", RegionCode: "310000", PolicyType: "pension", Embedding: []float64{0, 1, 0}},
		}, nil
	})
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	if c.Len() != 2 {
		t.Fatalf("expected 2 claims, got %d", c.Len())
	}
}

func TestCacheAdd(t *testing.T) {
	c := NewEmbeddingCache(func() ([]EmbeddedClaim, error) {
		return nil, nil
	})
	c.Load()
	c.Add(EmbeddedClaim{ClaimID: "c3", Embedding: []float64{0, 0, 1}})
	if c.Len() != 1 {
		t.Fatalf("expected 1 claim after add, got %d", c.Len())
	}
}

func TestSearchSimilar_ExactMatch(t *testing.T) {
	c := NewEmbeddingCache(func() ([]EmbeddedClaim, error) {
		return []EmbeddedClaim{
			{ClaimID: "c1", Embedding: []float64{1, 0, 0}},
			{ClaimID: "c2", Embedding: []float64{0, 1, 0}},
		}, nil
	})
	c.Load()
	results := c.SearchSimilar([]float64{1, 0, 0}, 0.5, 10, nil)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ClaimID != "c1" {
		t.Fatalf("expected c1 first (score=1.0), got %s", results[0].ClaimID)
	}
}

func TestSearchSimilar_Threshold(t *testing.T) {
	c := NewEmbeddingCache(func() ([]EmbeddedClaim, error) {
		return []EmbeddedClaim{
			{ClaimID: "c1", Embedding: []float64{1, 0, 0}},
			{ClaimID: "c2", Embedding: []float64{0, 1, 0}},
		}, nil
	})
	c.Load()
	results := c.SearchSimilar([]float64{1, 0, 0}, 0.99, 10, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result (threshold 0.99), got %d", len(results))
	}
}

func TestSearchSimilar_RegionFilter(t *testing.T) {
	c := NewEmbeddingCache(func() ([]EmbeddedClaim, error) {
		return []EmbeddedClaim{
			{ClaimID: "c1", RegionCode: "110000", Embedding: []float64{1, 0, 0}},
			{ClaimID: "c2", RegionCode: "310000", Embedding: []float64{1, 0, 0}},
		}, nil
	})
	c.Load()
	results := c.SearchSimilar([]float64{1, 0, 0}, 0, 10, &SearchFilter{RegionCode: "110000"})
	if len(results) != 1 || results[0].ClaimID != "c1" {
		t.Fatalf("expected 1 result for region 110000, got %d", len(results))
	}
}

func TestSearchSimilar_EmptyCache(t *testing.T) {
	c := NewEmbeddingCache(func() ([]EmbeddedClaim, error) {
		return nil, nil
	})
	c.Load()
	results := c.SearchSimilar([]float64{1, 0, 0}, 0, 10, nil)
	if len(results) != 0 {
		t.Fatalf("empty cache should return 0 results, got %d", len(results))
	}
}

func TestCacheConcurrency(t *testing.T) {
	c := NewEmbeddingCache(func() ([]EmbeddedClaim, error) {
		return []EmbeddedClaim{{ClaimID: "c1", Embedding: []float64{1, 0, 0}}}, nil
	})
	c.Load()
	done := make(chan bool)
	go func() {
		c.SearchSimilar([]float64{1, 0, 0}, 0, 10, nil)
		done <- true
	}()
	go func() {
		c.Add(EmbeddedClaim{ClaimID: "c2", Embedding: []float64{0, 1, 0}})
		done <- true
	}()
	<-done
	<-done
	// 没有 panic 即通过
}
```

- [ ] **Step 3: Run tests**

```powershell
go test ./internal/embeddings/ -v 2>&1
```
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add nsi-platform/services/policy-crawler/internal/embeddings/cache.go
git add nsi-platform/services/policy-crawler/internal/embeddings/cache_test.go
git commit -m "feat(embeddings): add EmbeddingCache with SearchSimilar"
```

---

### Task 4: DBStore 增加 LoadEmbeddings 方法

**Files:**
- Modify: `nsi-platform/services/policy-crawler/internal/crawler/store.go`
- Modify: `nsi-platform/services/policy-crawler/internal/crawler/store_test.go`（如果存在）

- [ ] **Step 1: 修改 store.go，增加 LoadEmbeddings**

在 `SaveEmbedding` 方法后面添加：

```go
func (s *DBStore) LoadEmbeddings() ([]embeddings.EmbeddedClaim, error) {
	rows, err := s.db.Query(`
		SELECT claim_id, policy_id, policy_type, region_code, embedding, 
		       COALESCE(source_name, ''), COALESCE(status, 'pending_review')
		FROM policy_claims 
		WHERE embedding IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("query embeddings: %w", err)
	}
	defer rows.Close()

	var result []embeddings.EmbeddedClaim
	for rows.Next() {
		var ec embeddings.EmbeddedClaim
		if err := rows.Scan(&ec.ClaimID, &ec.PolicyID, &ec.PolicyType, &ec.RegionCode,
			pq.Array(&ec.Embedding), &ec.SourceName, &ec.Status); err != nil {
			return nil, fmt.Errorf("scan embedding: %w", err)
		}
		result = append(result, ec)
	}
	return result, rows.Err()
}
```

需要在文件顶部 import 中添加 `"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"`。

- [ ] **Step 2: 编译验证**

```powershell
go build ./internal/crawler/ 2>&1
```
Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add nsi-platform/services/policy-crawler/internal/crawler/store.go
git commit -m "feat(crawler): add LoadEmbeddings method to DBStore"
```

---

### Task 5: 提取器集成交叉验证

**Files:**
- Modify: `nsi-platform/services/policy-crawler/internal/extractor/extractor.go`
- Modify: `nsi-platform/services/policy-crawler/internal/embeddings/embedding.go`（如需要）

- [ ] **Step 1: 修改 Extractor struct 增加 ReferenceChecker 接口**

在 `extractor.go` 中新增接口，并修改 Extractor struct 和构造函数：

```go
// ReferenceChecker 交叉验证接口
type ReferenceChecker interface {
	SearchSimilar(emb []float64, threshold float64, limit int, filter *embeddings.SearchFilter) []embeddings.SimilarResult
}

// Extractor LLM 政策提取器
type Extractor struct {
	store    RawTextStore
	client   *llm.Client
	checker  ReferenceChecker // 新增
}

func NewExtractor(store RawTextStore, client *llm.Client) *Extractor {
	return &Extractor{store: store, client: client}
}

// 可选：设置交叉验证器（避免改构造函数签名）
func (e *Extractor) SetReferenceChecker(c ReferenceChecker) {
	e.checker = c
}
```

- [ ] **Step 2: 在 ProcessOne 中插入交叉验证逻辑**

在 Step 4 和 Step 5 之间（parsed 解析后、构建 claim 前），添加：

```go
	// 4f. 交叉验证（如果配置了 checker）
	if e.checker != nil {
		embedText := parsed.PolicyType + " " + parsed.SubsidyCalcMethod + " " + parsed.PolicyID + " " + parsed.RegionCode
		if len(parsed.Conditions) > 0 {
			for _, c := range parsed.Conditions {
				if name, ok := c["name"].(string); ok {
					embedText += " " + name
				}
				if desc, ok := c["description"].(string); ok {
					embedText += " " + desc
				}
			}
		}
		emb := embeddings.FromText(embedText)
		similar := e.checker.SearchSimilar(emb, 0.5, 10, &embeddings.SearchFilter{RegionCode: parsed.RegionCode})
		maxScore := 0.0
		var bestMatch *embeddings.SimilarResult
		for i := range similar {
			if similar[i].Score > maxScore {
				maxScore = similar[i].Score
				bestMatch = &similar[i]
			}
		}
		if bestMatch != nil && parsed.RegionCode != "" && bestMatch.RegionCode == parsed.RegionCode {
			switch {
			case maxScore > 0.85:
				claim.Status = "pending_review"
			case maxScore > 0.7:
				if parsed.AmountMin != nil && bestMatch.Score > 0 {
					// 计算金额差异
					diff := math.Abs(*parsed.AmountMin-amountFromResult(bestMatch)) / math.Max(*parsed.AmountMin, amountFromResult(bestMatch))
					if diff > 0.5 {
						claim.Status = "unverified"
						claim.ConfidenceScore *= 0.5
					} else {
						claim.Status = "verified"
						claim.ConfidenceScore += 0.05
					}
				}
			}
		}
	}
```

需要添加 `"math"` 到 import。

辅助函数：
```go
func amountFromResult(r *embeddings.SimilarResult) float64 {
	// 从 DB 查询具体金额（简化：返回 0，实际场景可扩展）
	return 0
}
```

实际上这里设计有误——similar results 没有 amount 字段。金额需要在搜索时额外关联。简化起见，当 maxScore > 0.7 且 parsed.AmountMin != nil 时，标记为 pending_review 让人工审核即可。矛盾检测在 MVP 阶段做到这个粒度足够。

修正后的逻辑：

```go
	if bestMatch != nil && parsed.RegionCode != "" && bestMatch.RegionCode == parsed.RegionCode {
		switch {
		case maxScore > 0.85:
			claim.Status = "pending_review" // 疑似重复
		case maxScore > 0.7:
			claim.Status = "pending_review" // 高度相似，需人工确认
			claim.ConfidenceScore *= 0.9
		}
	}
```

- [ ] **Step 3: 编译验证**

```powershell
go build ./internal/extractor/ 2>&1
```
Expected: 无错误

- [ ] **Step 4: 修改 admin_llm.go 传递 ReferenceChecker**

将 `LLMExtractRunHandler` 签名改为：

```go
func LLMExtractRunHandler(store interface{}, checker extractor.ReferenceChecker) http.Handler {
```

在 goroutine 中创建 extractor 后设置 checker：

```go
ext := extractor.NewExtractor(rawStore, client)
if checker != nil {
    ext.SetReferenceChecker(checker)
}
```

- [ ] **Step 5: 编译验证**

```powershell
go build ./internal/extractor/ ./internal/admin/ 2>&1
```
Expected: 无错误

- [ ] **Step 6: Commit**

```bash
git add nsi-platform/services/policy-crawler/internal/extractor/extractor.go
git add nsi-platform/services/policy-crawler/internal/admin/admin_llm.go
git commit -m "feat(extractor): add cross-validation via ReferenceChecker"
```

---

### Task 6: 外部 API 端点

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/handler/similar_handler.go`
- Modify: `nsi-platform/services/policy-crawler/cmd/main.go`

- [ ] **Step 1: 创建 similar handler**

```go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
)

type SimilarRequest struct {
	ClaimID    string `json:"claim_id"`    // 二选一
	Text       string `json:"text"`        // 二选一
	RegionCode string `json:"region"`
	PolicyType string `json:"policy_type"`
	Limit      int    `json:"limit"`
}

type SimilarCache interface {
	SearchSimilar(emb []float64, threshold float64, limit int, filter *embeddings.SearchFilter) []embeddings.SimilarResult
}

func SimilarSearchHandler(cache SimilarCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req SimilarRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"code":-1,"msg":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if req.ClaimID == "" && req.Text == "" {
			http.Error(w, `{"code":-1,"msg":"claim_id or text required"}`, http.StatusBadRequest)
			return
		}
		limit := req.Limit
		if limit <= 0 || limit > 50 {
			limit = 10
		}

		// 优先用 claim_id 获取嵌入
		var emb []float64
		if req.ClaimID != "" {
			// 从缓存或 DB 加载该 claim 的嵌入
			emb = loadEmbeddingByClaimID(req.ClaimID)
		}
		if emb == nil && req.Text != "" {
			emb = embeddings.FromText(req.Text)
		}
		if emb == nil {
			http.Error(w, `{"code":-1,"msg":"cannot generate embedding"}`, http.StatusBadRequest)
			return
		}

		var filter *embeddings.SearchFilter
		if req.RegionCode != "" || req.PolicyType != "" {
			filter = &embeddings.SearchFilter{
				RegionCode: req.RegionCode,
				PolicyType: req.PolicyType,
			}
		}
		results := cache.SearchSimilar(emb, 0, limit, filter)
		if results == nil {
			results = []embeddings.SimilarResult{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": results,
		})
	})
}

// loadEmbeddingByClaimID 从缓存加载指定 claim 的嵌入向量
// 由于缓存只存了 []EmbeddedClaim，需要遍历查找
func loadEmbeddingByClaimID(claimID string) []float64 {
	// 这是一个桩函数，需要从 EmbeddingCache 中查找
	// 实际实现需要修改 EmbeddingCache 暴露一个 GetByID 方法
	return nil
}
```

这个方案有缺陷——`loadEmbeddingByClaimID` 无法访问 cache。更好的方案是让 EmbeddingCache 暴露 `GetEmbedding(claimID string) []float64` 方法。

改为先给 EmbeddingCache 增加 GetEmbedding 方法：

```go
// GetEmbedding 根据 claimID 获取嵌入向量
func (c *EmbeddingCache) GetEmbedding(claimID string) []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, cc := range c.claims {
		if cc.ClaimID == claimID {
			return cc.Embedding
		}
	}
	return nil
}
```

- [ ] **Step 2: 在 EmbeddingCache 中增加 GetEmbedding 方法**

```go
// GetEmbedding 根据 claimID 获取嵌入向量
func (c *EmbeddingCache) GetEmbedding(claimID string) []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, cc := range c.claims {
		if cc.ClaimID == claimID {
			return cc.Embedding
		}
	}
	return nil
}
```

- [ ] **Step 3: 完善 SimilarSearchHandler**

```go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
)

type SimilarRequest struct {
	ClaimID    string `json:"claim_id"`
	Text       string `json:"text"`
	RegionCode string `json:"region"`
	PolicyType string `json:"policy_type"`
	Limit      int    `json:"limit"`
}

type EmbeddingSource interface {
	GetEmbedding(claimID string) []float64
	SearchSimilar(emb []float64, threshold float64, limit int, filter *embeddings.SearchFilter) []embeddings.SimilarResult
}

func SimilarSearchHandler(cache EmbeddingSource) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req SimilarRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": -1, "msg": "invalid JSON"})
			return
		}
		if req.ClaimID == "" && req.Text == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": -1, "msg": "claim_id or text required"})
			return
		}
		limit := req.Limit
		if limit <= 0 || limit > 50 {
			limit = 10
		}

		var emb []float64
		if req.ClaimID != "" {
			emb = cache.GetEmbedding(req.ClaimID)
		}
		if emb == nil && req.Text != "" {
			emb = embeddings.FromText(req.Text)
		}
		if emb == nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": -1, "msg": "cannot generate embedding"})
			return
		}

		var filter *embeddings.SearchFilter
		if req.RegionCode != "" || req.PolicyType != "" {
			filter = &embeddings.SearchFilter{RegionCode: req.RegionCode, PolicyType: req.PolicyType}
		}
		results := cache.SearchSimilar(emb, 0, limit, filter)
		if results == nil {
			results = []embeddings.SimilarResult{}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": results})
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
```

- [ ] **Step 4: 编译验证**

```powershell
go build ./internal/handler/ 2>&1
```
Expected: 无错误

- [ ] **Step 5: Commit**

```bash
git add nsi-platform/services/policy-crawler/internal/embeddings/cache.go
mkdir -p nsi-platform/services/policy-crawler/internal/handler
git add nsi-platform/services/policy-crawler/internal/handler/similar_handler.go
git commit -m "feat(api): add POST /v1/policies/similar endpoint"
```

---

### Task 7: Admin 语义搜索 Handler

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/admin/admin_search.go`
- Modify: `nsi-platform/services/policy-crawler/internal/admin/admin.go`
- Modify: `nsi-platform/services/policy-crawler/internal/admin/admin_page.go`
- Modify: `nsi-platform/services/policy-crawler/cmd/main.go`

- [ ] **Step 1: 创建 admin_search.go**

```go
package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
)

type EmbeddingSearcher interface {
	SearchSimilar(emb []float64, threshold float64, limit int, filter *embeddings.SearchFilter) []embeddings.SimilarResult
}

func AdminSearchHandler(searcher EmbeddingSearcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		region := strings.TrimSpace(r.URL.Query().Get("region"))
		policyType := strings.TrimSpace(r.URL.Query().Get("type"))
		limitStr := r.URL.Query().Get("limit")

		limit := 20
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
			limit = n
		}

		if q == "" {
			respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": []embeddings.SimilarResult{}})
			return
		}

		emb := embeddings.FromText(q)
		var filter *embeddings.SearchFilter
		if region != "" || policyType != "" {
			filter = &embeddings.SearchFilter{RegionCode: region, PolicyType: policyType}
		}
		results := searcher.SearchSimilar(emb, 0, limit, filter)
		if results == nil {
			results = []embeddings.SimilarResult{}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": results})
	})
}

// HTMLSearchPage 返回语义搜索页面 HTML
func HTMLSearchPage() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>语义搜索 - AI社保智筹</title>
<style>
body { font-family: "Microsoft YaHei", sans-serif; margin: 20px; background: #f5f5f5; }
.container { max-width: 1000px; margin: 0 auto; }
.search-box { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
.search-box input, .search-box select { padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
.search-box input[type="text"] { width: 60%%; margin-right: 10px; }
.search-box button { padding: 8px 20px; background: #1890ff; color: #fff; border: none; border-radius: 4px; cursor: pointer; }
.search-box button:hover { background: #40a9ff; }
.result-item { background: #fff; margin-top: 12px; padding: 16px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.result-item .score { float: right; font-weight: bold; }
.score-high { color: #52c41a; }
.score-mid { color: #faad14; }
.score-low { color: #999; }
.result-item .meta { color: #666; font-size: 13px; margin-top: 6px; }
.result-item .meta span { margin-right: 16px; }
</style>
</head>
<body>
<div class="container">
<h1>🔍 语义搜索</h1>
<div class="search-box">
<input type="text" id="searchInput" placeholder="输入搜索关键词，如：灵活就业补贴 北京">
<select id="regionSelect">
<option value="">全部城市</option>
<option value="110000">北京</option>
<option value="310000">上海</option>
<option value="440100">广州</option>
<option value="330100">杭州</option>
<option value="440300">深圳</option>
</select>
<select id="typeSelect">
<option value="">全部类型</option>
<option value="subsidy">补贴</option>
<option value="pension">养老</option>
<option value="medical">医疗</option>
<option value="unemployment">失业</option>
<option value="injury">工伤</option>
<option value="maternity">生育</option>
<option value="housing_fund">公积金</option>
<option value="training">培训</option>
</select>
<button onclick="search()">搜索</button>
</div>
<div id="results"></div>
</div>
<script>
function search() {
  const q = document.getElementById('searchInput').value;
  const region = document.getElementById('regionSelect').value;
  const type = document.getElementById('typeSelect').value;
  if (!q) return;
  document.getElementById('results').innerHTML = '<p>搜索中...</p>';
  fetch('/admin/llm/search?q=' + encodeURIComponent(q) + '&region=' + region + '&type=' + type)
    .then(r => r.json())
    .then(d => {
      if (d.code !== 0) { document.getElementById('results').innerHTML = '<p style="color:red">错误: ' + d.msg + '</p>'; return; }
      const items = d.data || [];
      if (items.length === 0) { document.getElementById('results').innerHTML = '<p>没有找到匹配的政策</p>'; return; }
      let html = '';
      for (const item of items) {
        const cls = item.score >= 0.8 ? 'score-high' : (item.score >= 0.6 ? 'score-mid' : 'score-low');
        html += '<div class="result-item">';
        html += '<div><strong>' + item.policy_id + '</strong> <span class="score ' + cls + '">' + (item.score*100).toFixed(1) + '%%</span></div>';
        html += '<div class="meta">';
        html += '<span>类型: ' + item.policy_type + '</span>';
        html += '<span>城市: ' + item.region_code + '</span>';
        html += '<span>来源: ' + (item.source_name || '-') + '</span>';
        html += '<span>状态: ' + item.status + '</span>';
        html += '</div></div>';
      }
      document.getElementById('results').innerHTML = html;
    })
    .catch(e => { document.getElementById('results').innerHTML = '<p style="color:red">请求失败: ' + e.message + '</p>'; });
}
</script>
</body>
</html>`)
}
```

- [ ] **Step 2: 在 admin.go 中注册路由**

在 `RegisterAdminRoutes` 函数中添加（如果没有该函数，则在已有的路由注册处添加）：

```go
// 语义搜索
mux.Handle("/admin/search", middleware.RecoveryMiddleware()(AdminSearchHandler(searcher)))
mux.Handle("/admin/llm/search", middleware.RecoveryMiddleware()(AdminSearchHandler(searcher)))
mux.Handle("/admin/search_page", middleware.RecoveryMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(HTMLSearchPage()))
})))
```

- [ ] **Step 3: 在 admin_page.go 中增加"语义搜索"Tab**

在 `const adminHTML` 的 `navItems` 数组中（第 70-77 行），在 `{id:'import', ...}` 前面添加：

```javascript
  {id:'search',label:'\u8bed\u4e49\u641c\u7d22'},
```

在 `switchPanel` 函数中（第 87-100 行），`else if(id==='import')` 前面添加：

```javascript
  else if(id==='search')loadSearch();
```

在 `loadExtract` 函数后面添加 `loadSearch` 函数：

```javascript
function loadSearch(){
  var app=document.getElementById('app');
  app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>\u52a0\u8f7d\u4e2d...</div>'+
    '<iframe src="/admin/search_page" style="width:100%;height:800px;border:none;" onload="this.previousSibling.style.display=\'none\'"></iframe>';
}
```

- [ ] **Step 4: 编译验证**

```powershell
go build ./internal/admin/ 2>&1
```
Expected: 无错误

- [ ] **Step 5: Commit**

```bash
git add nsi-platform/services/policy-crawler/internal/admin/admin_search.go
git add nsi-platform/services/policy-crawler/internal/admin/admin.go
git add nsi-platform/services/policy-crawler/internal/admin/admin_page.go
git commit -m "feat(admin): add semantic search tab and handler"
```

---

### Task 8: main.go 初始化 EmbeddingCache 和路由

**Files:**
- Modify: `nsi-platform/services/policy-crawler/cmd/main.go`

- [ ] **Step 1: 修改 main.go**

在 `store` 初始化之后、`manager` 初始化之前，添加 EmbeddingCache 初始化：

```go
import (
	// ... 现有 import
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/handler"
)

// 在初始化 store 之后、manager 初始化之前
embedCache := embeddings.NewEmbeddingCache(store.LoadEmbeddings)
if err := embedCache.Load(); err != nil {
	log.Printf("[embeddings] warning: initial load failed: %v", err)
}
// 每分钟刷新缓存
go func() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		if err := embedCache.Refresh(); err != nil {
			log.Printf("[embeddings] refresh failed: %v", err)
		}
	}
}()
```

在路由注册部分：

```go
// 语义搜索 API
mux.Handle("/v1/policies/similar", middleware.RecoveryMiddleware()(handler.SimilarSearchHandler(embedCache)))

// Admin 语义搜索（已包含在 admin 注册中，确保 searcher 参数传递）
```

修改 `RegisterAdminRoutes` 调用或直接注册，将 `embedCache` 传给 admin 搜索 handler。

- [ ] **Step 2: 全量编译验证**

```powershell
go build ./cmd/ 2>&1
```
Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add nsi-platform/services/policy-crawler/cmd/main.go
git commit -m "feat(main): init EmbeddingCache with timer, register similarity routes"
```

---

### Task 9: 集成测试

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/embeddings/similarity_integration_test.go`（可选）
- Create: `nsi-platform/services/policy-crawler/cmd/integration_test.go` 或修改现有测试

- [ ] **Step 1: 构建新二进制**

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
cd nsi-platform/services/policy-crawler
go build -o bin/policy-crawler ./cmd/main.go 2>&1
```
Expected: 无输出（成功）

- [ ] **Step 2: 重建 Docker 镜像并重启**

```powershell
cd nsi-platform
docker build -t nsi-policy-crawler:latest services/policy-crawler
docker compose up -d policy-crawler
```
Expected: 容器成功启动

- [ ] **Step 3: 验证服务健康**

```powershell
Start-Sleep -Seconds 10
curl.exe -s http://127.0.0.1:39403/admin/llm/progress
```
Expected: 返回 JSON（缓存已加载）

- [ ] **Step 4: 验证语义搜索 Admin API**

```powershell
curl.exe -s "http://127.0.0.1:39403/admin/llm/search?q=失业补贴&limit=5"
```
Expected: 返回含 score 的搜索结果 JSON

- [ ] **Step 5: 验证外部 API**

```powershell
curl.exe -s -X POST http://127.0.0.1:39403/v1/policies/similar -H "Content-Type: application/json" -d "{\"text\":\"北京灵活就业社保补贴\",\"region\":\"110000\",\"limit\":5}"
```
Expected: 返回相似政策列表，score 从高到低排序

- [ ] **Step 6: 验证管理后台搜索页面**

访问 `http://127.0.0.1:39403/admin` 并在导航中点击"语义搜索" Tab，输入关键词搜索。
Expected: 显示搜索结果表格

- [ ] **Step 7: 提交**

```bash
git add nsi-platform/services/policy-crawler/bin/policy-crawler
git commit -m "test: verify cross-validation end-to-end via curl"
```

---

### 自检清单

- [ ] 每个任务的代码块包含完整实现
- [ ] 无 TBD/TODO/占位符
- [ ] 类型签名跨任务一致（EmbeddedClaim, SimilarResult, SearchFilter）
- [ ] 测试覆盖核心函数（余弦相似度、缓存搜索、并发安全、空值边界）
- [ ] 文件路径准确，无拼写错误
