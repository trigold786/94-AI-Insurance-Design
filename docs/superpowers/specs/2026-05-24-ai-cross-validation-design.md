# AI 交叉验证 + 语义搜索 设计文档

## 概述

为 AI社保智筹 系统增加基于嵌入向量的政策语义相似度搜索和交叉验证能力，提升政策匹配准确性和重复检测能力。

## 架构

```
用户请求 → API Handler → EmbeddingCache.Search() → 排序返回
                                      ↕
提取器 ProcessOne → 生成嵌入 → EmbeddingCache.Search() → 矛盾检测 → 调整 status/confidence
                                      ↕
                               PostgreSQL float8[]
                                      ↕
                              Timer (每分钟) 全量刷新缓存
```

## 组件设计

### 1. `internal/embeddings/similarity.go`

```go
// SimilarResult 相似搜索结果
type SimilarResult struct {
    ClaimID    string  `json:"claim_id"`
    PolicyID   string  `json:"policy_id"`
    PolicyType string  `json:"policy_type"`
    RegionCode string  `json:"region_code"`
    Score      float64 `json:"score"`       // 余弦相似度 [0,1]
    SourceName string  `json:"source_name"`
    PolicyURL  string  `json:"policy_url"`
    Status     string  `json:"status"`
}

// CosineSimilarity 计算两个向量的余弦相似度
func CosineSimilarity(a, b []float64) float64

// SearchSimilar 在缓存中搜索最相似的 K 条结果
// threshold: 最低相似度阈值；limit: 最大返回条数
// filter: 可选的 region_code / policy_type 过滤
func (c *EmbeddingCache) SearchSimilar(emb []float64, threshold float64, limit int, filter *SearchFilter) []SimilarResult
```

### 2. `internal/embeddings/cache.go`

```go
// EmbeddedClaim 缓存的嵌入数据
type EmbeddedClaim struct {
    ClaimID    string
    PolicyID   string
    PolicyType string
    RegionCode string
    Embedding  []float64
    SourceName string
    Status     string
}

// EmbeddingCache 内存嵌入缓存
type EmbeddingCache struct {
    mu      sync.RWMutex
    claims  []EmbeddedClaim
    loader  func() ([]EmbeddedClaim, error)  // 从 DB 加载的回调
}

func NewEmbeddingCache(loader func() ([]EmbeddedClaim, error)) *EmbeddingCache
func (c *EmbeddingCache) Load() error           // 首次/全量加载
func (c *EmbeddingCache) Refresh() error         // 全量刷新（定时器）
func (c *EmbeddingCache) Add(ec EmbeddedClaim)   // 单条新增
func (c *EmbeddingCache) SearchSimilar(...) []SimilarResult
```

### 3. API 端点

**外部 API：** `POST /v1/policies/similar`

请求体：
```json
{
  "claim_id": "LLM-xxx",     // 二选一
  "text": "北京灵活就业补贴", // 二选一
  "region": "110000",        // 可选
  "policy_type": "subsidy",  // 可选
  "limit": 10
}
```

响应：
```json
{
  "code": 0,
  "data": [
    {"claim_id": "...", "policy_type": "subsidy", "region_code": "110000",
     "score": 0.92, "source_name": "北京人社局", "policy_url": "https://..."}
  ]
}
```

**管理后台：** `GET /admin/llm/search?q=失业补贴&region=110000&type=subsidy&limit=20`

- q 参数通过 `embeddings.FromText()` 转为嵌入向量后搜索
- 返回含 score 的政策列表

### 4. 提取器集成

修改 `extractor.go` 的 `ProcessOne` 方法，在 Step 4（LLM 解析）和 Step 5（构建 Claim）之间插入交叉验证：

1. 构建 claimText = policy_type + subsidy_calc_method + conditions
2. 用 `embeddings.FromText(claimText)` 生成嵌入向量
3. 调用 `cache.SearchSimilar(emb, 0.5, 10, &SearchFilter{RegionCode: parsed.RegionCode})`
4. 运行矛盾检测规则（见下方）
5. 根据检测结果调整 confidence 和 status
6. 继续原有流程（InsertClaim → SaveEmbedding → MarkExtracted）

### 5. 矛盾检测规则

```
同城（region_code 相同）:
  相似度 > 0.85 → 疑似重复，status=pending_review，confidence 不变
  相似度 0.7-0.85 + amount_min 差异 > 50% → 疑似矛盾，status=unverified，confidence×0.5
  相似度 0.7-0.85 + amount_min 差异 ≤ 50% → 一致，status=verified，confidence+0.05
  相似度 < 0.7 → 正常，status=verified，confidence 不变

不同城:
  相似度 > 0.6 → 状态不变，API 返回时附带"相关城市政策"链接
  相似度 ≤ 0.6 → 不影响

> 差异计算：|a - b| / max(a, b)，确保分母不为 0 时结果在 [0,1] 区间。
> 不同城政策不改变 status，仅在 API 响应中附加 references 字段提示参考。

### 6. 管理后台 UI

在 Admin 新增「语义搜索」Tab：

- 搜索输入框（文本）+ 城市下拉筛选 + 政策类型下拉筛选 + 搜索按钮
- 结果列表：政策 ID、类型、城市、来源、相似度百分比（颜色标识：>0.8 绿色，0.6-0.8 黄色，<0.6 灰色）
- 每行有"查看详情"链接跳转到政策审核 Tab

## 数据流

### 提取时交叉验证
```
LLM JSON → 构建 claimText → FromText() → 嵌入向量
    → cache.SearchSimilar(emb, 0.5, 10)
    → 矛盾检测 → 调整 confidence/status
    → InsertClaim (含调整后的值)
    → SaveEmbedding
    → cache.Add(embeddedClaim)  // 立即更新缓存
    → MarkExtracted
```

### Admin 搜索
```
用户输入 q → FromText(q) → 嵌入向量
    → cache.SearchSimilar(emb, threshold=0, limit=20, filter={region,type})
    → 返回排序结果
```

## 测试

### 单元测试

`similarity_test.go`:
- `CosineSimilarity` 相同向量 → 1.0
- `CosineSimilarity` 正交向量 → 0.0
- `CosineSimilarity` 部分重叠 → [0,1] 区间
- `SearchSimilar` 空缓存 → 空结果
- `SearchSimilar` 精确匹配 → 最高分第一
- `SearchSimilar` 阈值过滤 → 低于阈值不返回
- `SearchSimilar` region 过滤 → 只返指定城市

`cache_test.go`:
- `Load()` 全量加载
- `Add()` 增量添加后 Search 包含新数据
- `Refresh()` 全量刷新
- 并发安全（goroutine 读写）

### 集成测试

- 提取器交叉验证集成：mock 缓存 → 验证 confidence/status 调整
- API `/v1/policies/similar` 端点测试

## 文件变更清单

| 文件 | 操作 |
|------|------|
| `internal/embeddings/similarity.go` | **新增** |
| `internal/embeddings/cache.go` | **新增** |
| `internal/embeddings/embedding.go` | 无变更（复用 FromText） |
| `internal/extractor/extractor.go` | **修改**：插入交叉验证逻辑 |
| `internal/crawler/store.go` | **修改**：增加 LoadEmbeddings 方法 |
| `internal/admin/admin_page.go` | **修改**：增加语义搜索 Tab |
| `internal/admin/admin_search.go` | **新增**：搜索 API handler |
| `internal/admin/admin.go` | **修改**：注册搜索路由 |
| `cmd/main.go` | **修改**：初始化 EmbeddingCache + 传参 |
| `shared/models/models.go` | **修改**：增加 SimilarResult 等类型（可选） |
| 对应 `*_test.go` | **新增/修改** |
