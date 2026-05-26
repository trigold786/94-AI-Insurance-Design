# pgvector 向量检索设计方案

## 目标

将 policy-crawler 的内存暴力余弦相似度搜索替换为 PostgreSQL pgvector 向量检索，同时将 hash bag-of-words embedding 升级为可配置的语义 embedding（默认 OpenAI text-embedding-3-small，降级为 hash）。

## 方案：完全替换单列 vector + 可配置 provider

### 决策依据

- 旧 hash embedding 是 256 维 `float8[]`，无法转为 1536 维 `vector(1536)`，需清空
- 单列比双列/独立表更简单，无 JOIN 开销
- 旧数据 embedding 清空后自然降级为关键词搜索（混合兼容）

---

## 1. 基础设施层

### 1.1 Docker 镜像切换

`infra/docker-compose.infra.yml` 中 postgres 服务：

```yaml
postgres:
  image: pgvector/pgvector:pg18
```

替换 `postgres:18-alpine`。删除 `infra/postgres/Dockerfile`（不再需要手动编译）。

### 1.2 Migration 012（对 nsi_crawler 库）

```sql
-- 012_pgvector.sql

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 清空旧 hash embedding（256维无法转为1536维 vector）
UPDATE policy_claims SET embedding = NULL WHERE embedding IS NOT NULL;

-- 类型转换 float8[] → vector(1536)
ALTER TABLE policy_claims
  ALTER COLUMN embedding TYPE vector(1536)
  USING NULL::vector;

-- IVFFlat 索引（适合 <10万条记录）
CREATE INDEX idx_policy_claims_embedding
  ON policy_claims
  USING ivfflat (embedding vector_cosine_ops)
  WITH (lists = 100);

-- llm_configs 增加 embedding 配置
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT 'text-embedding-3-small';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_dimensions INT DEFAULT 1536;
```

### 1.3 Go 依赖

新增 `github.com/pgvector/pgvector-go`。保持 `github.com/lib/pq` 不变。

---

## 2. Embedding Provider 接口

### 2.1 接口定义

文件：`policy-crawler/internal/embeddings/provider.go`

```go
type EmbeddingProvider interface {
    Embed(ctx context.Context, texts []string) ([][]float64, error)
    Dimensions() int
    ModelName() string
}
```

### 2.2 OpenAI Provider

文件：`policy-crawler/internal/embeddings/openai_provider.go`

- 从 `LLMConfig` 读取 `api_key`、`endpoint`（支持自定义 base_url）、`embedding_model`、`embedding_dimensions`
- 调用 `POST {base_url}/embeddings`，model 使用配置值
- 复用现有 LLM HTTP client 模式（`http.Client` + `Authorization: Bearer`）
- OpenAI-compatible API 格式（兼容 DeepSeek/百炼/volc_ark 等国内 provider 的 embedding endpoint）
- 请求结构：`{"model":"text-embedding-3-small","input":["text1","text2"]}`
- 响应结构：`{"data":[{"embedding":[0.1,...],...}]}`
- 批量上限 2048 条，超时 30s，重试 2 次指数退避

### 2.3 Hash Provider（降级）

文件：`policy-crawler/internal/embeddings/hash_provider.go`

- 保持现有 `FromText()` 逻辑（256维 hash bag-of-words）
- `Dimensions()` 返回 1536，`Embed()` 输出前 256 维填充原 hash 值，其余补零，然后 L2 归一化
- 当 `LLMConfig` 无 API Key 或 embedding API 调用失败时自动降级

### 2.4 Provider 工厂

```go
func NewProviderFromConfig(apiKey, baseURL, model string, dimensions int) EmbeddingProvider {
    if apiKey != "" {
        return &OpenAIProvider{...}
    }
    return &HashProvider{}
}
```

### 2.5 写入流程

新 claim 提取完成后（extractor.go step 7）：
1. 组装文本：`policy_type + " " + subsidy_calc_method + " " + policy_id + " " + region_code + conditions文本`
2. 调用 `provider.Embed(ctx, []string{text})` 获取向量
3. 写入 DB：`UPDATE policy_claims SET embedding = $1 WHERE claim_id = $2`（使用 `pgvector.Vector` 类型）

---

## 3. 搜索层替换

### 3.1 VectorSearcher

文件：`policy-crawler/internal/embeddings/pgvector_search.go`

```go
type VectorSearcher struct {
    db       *sql.DB
    provider EmbeddingProvider
}

func NewVectorSearcher(db *sql.DB, provider EmbeddingProvider) *VectorSearcher

func (s *VectorSearcher) GetEmbedding(claimID string) []float64
func (s *VectorSearcher) SearchSimilar(emb []float64, threshold float64, limit int, filter *SearchFilter) []SimilarResult
func (s *VectorSearcher) SearchByText(ctx context.Context, query string, threshold float64, limit int, filter *SearchFilter) ([]SimilarResult, error)
func (s *VectorSearcher) KeywordSearch(query string, limit int, filter *SearchFilter) ([]SimilarResult, error)
```

### 3.2 pgvector SQL 查询

```sql
SELECT claim_id, policy_id, policy_type, region_code,
       COALESCE(source_name, ''), COALESCE(policy_url, ''), COALESCE(status, 'pending_review'),
       1 - (embedding <=> $1::vector) AS score
FROM policy_claims
WHERE embedding IS NOT NULL
  AND ($2::text = '' OR region_code = $2)
  AND ($3::text = '' OR policy_type = $3)
  AND 1 - (embedding <=> $1::vector) >= $4
ORDER BY embedding <=> $1::vector
LIMIT $5
```

### 3.3 关键词降级

当 embedding provider 不可用或 embedding 列为 NULL 时：

```sql
SELECT claim_id, policy_id, policy_type, region_code,
       COALESCE(source_name, ''), COALESCE(policy_url, ''), status,
       confidence_score AS score
FROM policy_claims
WHERE (policy_id ILIKE '%' || $1 || '%' OR subsidy_calc_method ILIKE '%' || $1 || '%')
  AND ($2::text = '' OR region_code = $2)
  AND ($3::text = '' OR policy_type = $3)
ORDER BY confidence_score DESC
LIMIT $4
```

### 3.4 调用方替换

| 调用点 | 当前 | 替换为 |
|--------|------|--------|
| `POST /v1/policies/similar` | `EmbeddingCache` | `VectorSearcher` |
| `GET /admin/llm/search?q=` | `EmbeddingCache` | `VectorSearcher` |
| extractor.go 去重 | `ReferenceChecker → EmbeddingCache` | `VectorSearcher` |
| `GET /admin/llm/extract` | 传 `embedCache` 作 checker | 传 `VectorSearcher` 作 checker |

### 3.5 移除 EmbeddingCache

- 删除 `cache.go` 中后台 goroutine（每分钟全量加载）
- 删除 `EmbeddingCache` 结构体
- 保留 `CosineSimilarity()`（用于测试）
- 保留 `EmbeddedClaim`、`SimilarResult`、`SearchFilter` 结构体

### 3.6 接口兼容

`VectorSearcher` 同时实现现有两个接口：

```go
// handler/similar_handler.go
type EmbeddingSource interface {
    GetEmbedding(claimID string) []float64
    SearchSimilar(emb []float64, threshold float64, limit int, filter *SearchFilter) []SimilarResult
}

// admin/admin_search.go
type EmbeddingSearcher interface {
    SearchSimilar(emb []float64, threshold float64, limit int, filter *SearchFilter) []SimilarResult
}

// extractor/extractor.go
type ReferenceChecker interface {
    SearchSimilar(emb []float64, threshold float64, limit int, filter *SearchFilter) []SimilarResult
}
```

`VectorSearcher` 实现所有三个接口，无需改动接口定义或 handler/extractor 签名。

---

## 4. 接口变更汇总

### 新增文件

| 文件 | 内容 |
|------|------|
| `migrations/012_pgvector.sql` | pgvector 扩展 + 类型转换 + 索引 + llm_configs 字段 |
| `internal/embeddings/provider.go` | `EmbeddingProvider` 接口 + 工厂 |
| `internal/embeddings/openai_provider.go` | OpenAI-compatible embedding provider |
| `internal/embeddings/hash_provider.go` | Hash 降级 provider（1536 维输出） |
| `internal/embeddings/pgvector_search.go` | `VectorSearcher` 实现 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `infra/docker-compose.infra.yml` | postgres 镜像切换 |
| `cmd/main.go` | `EmbeddingCache` → `VectorSearcher`，provider 初始化 |
| `internal/crawler/store.go` | `SaveEmbedding()` 用 `pgvector.Vector` 替代 `pq.Array`；`LoadEmbeddings()` 移除（不再需要全量加载） |
| `internal/extractor/extractor.go` | step 7 embedding 生成改用 provider |
| `internal/embeddings/embedding.go` | `Dim` 常量改为 1536 或导出为可配置 |
| `shared/models/models.go` | `Embedding` 字段类型可能需调整 |

### 删除文件

| 文件 | 原因 |
|------|------|
| `infra/postgres/Dockerfile` | 改用官方预编译镜像 |

### 保留不动的文件

- `internal/embeddings/similarity.go` — `CosineSimilarity()` 仍用于测试
- `internal/embeddings/models.go` — `SimilarResult`、`SearchFilter` 结构体保留
- `internal/handler/similar_handler.go` — 接口不变，注入实现改为 `VectorSearcher`
- `internal/admin/admin_search.go` — 同上
- `internal/embeddings/cache_test.go` — 需更新或重写为 `pgvector_search_test.go`
- `internal/embeddings/similarity_test.go` — 不变

---

## 5. 测试计划

1. **OpenAI Provider 单元测试**：mock HTTP，验证请求/响应解析
2. **Hash Provider 单元测试**：验证 1536 维输出（前 256 非零，后 1280 零）
3. **VectorSearcher 单元测试**：mock `*sql.DB`，验证 SQL 查询构建
4. **集成测试**：需要 pgvector PostgreSQL 实例
5. **降级测试**：验证 provider 失败时回退到关键词搜索
6. **现有测试修复**：`cache_test.go` → `pgvector_search_test.go`

---

## 6. 风险与缓解

| 风险 | 缓解 |
|------|------|
| pgvector 镜像拉不下来 | 可用 `pgvector/pgvector:pg17` 或 `ankane/pgvector` 替代 |
| 旧 embedding 数据丢失 | 旧数据 embedding 全部置 NULL，搜索走关键词降级，新写入自动用新模型 |
| OpenAI API 费用 | embedding 价格极低（$0.02/1M tokens），且只在 claim 入库时调用 |
| `lib/pq` 不支持 vector 类型 | 使用 `pgvector-go` 库的 `pgvector.Vector` 类型 + text cast |
| migration 失败（数据量大） | `UPDATE SET embedding = NULL` 大表可能慢，可分批处理 |
