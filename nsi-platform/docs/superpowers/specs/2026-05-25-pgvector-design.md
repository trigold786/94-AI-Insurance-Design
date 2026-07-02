# pgvector 向量检索设计方�?
## 目标

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

�?policy-crawler 的内存暴力余弦相似度搜索替换�?PostgreSQL pgvector 向量检索，同时�?hash bag-of-words embedding 升级为可配置的语�?embedding（默�?OpenAI text-embedding-3-small，降级为 hash）�?
## 方案：完全替换单�?vector + 可配�?provider

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### 决策依据

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- �?hash embedding �?256 �?`float8[]`，无法转�?1536 �?`vector(1536)`，需清空
- 单列比双�?独立表更简单，�?JOIN 开销
- 旧数�?embedding 清空后自然降级为关键词搜索（混合兼容�?
---

## 1. 基础设施�?
### 1.1 Docker 镜像切换

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

`infra/docker-compose.infra.yml` �?postgres 服务�?
```yaml
postgres:
  image: pgvector/pgvector:pg18
```

替换 `postgres:18-alpine`。删�?`infra/postgres/Dockerfile`（不再需要手动编译）�?
### 1.2 Migration 012（对 nsi_crawler 库）

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```sql
-- 012_pgvector.sql

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 清空�?hash embedding�?56维无法转�?536�?vector�?UPDATE policy_claims SET embedding = NULL WHERE embedding IS NOT NULL;

-- 类型转换 float8[] �?vector(1536)
ALTER TABLE policy_claims
  ALTER COLUMN embedding TYPE vector(1536)
  USING NULL::vector;

-- IVFFlat 索引（适合 <10万条记录�?CREATE INDEX idx_policy_claims_embedding
  ON policy_claims
  USING ivfflat (embedding vector_cosine_ops)
  WITH (lists = 100);

-- llm_configs 增加 embedding 配置
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT 'text-embedding-3-small';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_dimensions INT DEFAULT 1536;
```

### 1.3 Go 依赖

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

新增 `github.com/pgvector/pgvector-go`。保�?`github.com/lib/pq` 不变�?
---

## 2. Embedding Provider 接口

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### 2.1 接口定义

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

文件：`policy-crawler/internal/embeddings/provider.go`

```go
type EmbeddingProvider interface {
    Embed(ctx context.Context, texts []string) ([][]float64, error)
    Dimensions() int
    ModelName() string
}
```

### 2.2 OpenAI Provider

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

文件：`policy-crawler/internal/embeddings/openai_provider.go`

- �?`LLMConfig` 读取 `api_key`、`endpoint`（支持自定义 base_url）、`embedding_model`、`embedding_dimensions`
- 调用 `POST {base_url}/embeddings`，model 使用配置�?- 复用现有 LLM HTTP client 模式（`http.Client` + `Authorization: Bearer`�?- OpenAI-compatible API 格式（兼�?DeepSeek/百炼/volc_ark 等国�?provider �?embedding endpoint�?- 请求结构：`{"model":"text-embedding-3-small","input":["text1","text2"]}`
- 响应结构：`{"data":[{"embedding":[0.1,...],...}]}`
- 批量上限 2048 条，超时 30s，重�?2 次指数退�?
### 2.3 Hash Provider（降级）

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

文件：`policy-crawler/internal/embeddings/hash_provider.go`

- 保持现有 `FromText()` 逻辑�?56�?hash bag-of-words�?- `Dimensions()` 返回 1536，`Embed()` 输出�?256 维填充原 hash 值，其余补零，然�?L2 归一�?- �?`LLMConfig` �?API Key �?embedding API 调用失败时自动降�?
### 2.4 Provider 工厂

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```go
func NewProviderFromConfig(apiKey, baseURL, model string, dimensions int) EmbeddingProvider {
    if apiKey != "" {
        return &OpenAIProvider{...}
    }
    return &HashProvider{}
}
```

### 2.5 写入流程

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

�?claim 提取完成后（extractor.go step 7）：
1. 组装文本：`policy_type + " " + subsidy_calc_method + " " + policy_id + " " + region_code + conditions文本`
2. 调用 `provider.Embed(ctx, []string{text})` 获取向量
3. 写入 DB：`UPDATE policy_claims SET embedding = $1 WHERE claim_id = $2`（使�?`pgvector.Vector` 类型�?
---

## 3. 搜索层替�?
### 3.1 VectorSearcher

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

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

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

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

### 3.3 关键词降�?
�?embedding provider 不可用或 embedding 列为 NULL 时：

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

### 3.4 调用方替�?
| 调用�?| 当前 | 替换�?|
|--------|------|--------|
| `POST /v1/policies/similar` | `EmbeddingCache` | `VectorSearcher` |
| `GET /admin/llm/search?q=` | `EmbeddingCache` | `VectorSearcher` |
| extractor.go 去重 | `ReferenceChecker �?EmbeddingCache` | `VectorSearcher` |
| `GET /admin/llm/extract` | �?`embedCache` �?checker | �?`VectorSearcher` �?checker |

### 3.5 移除 EmbeddingCache

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- 删除 `cache.go` 中后�?goroutine（每分钟全量加载�?- 删除 `EmbeddingCache` 结构�?- 保留 `CosineSimilarity()`（用于测试）
- 保留 `EmbeddedClaim`、`SimilarResult`、`SearchFilter` 结构�?
### 3.6 接口兼容

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

`VectorSearcher` 同时实现现有两个接口�?
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

`VectorSearcher` 实现所有三个接口，无需改动接口定义�?handler/extractor 签名�?
---

## 4. 接口变更汇�?
### 新增文件

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| 文件 | 内容 |
|------|------|
| `migrations/012_pgvector.sql` | pgvector 扩展 + 类型转换 + 索引 + llm_configs 字段 |
| `internal/embeddings/provider.go` | `EmbeddingProvider` 接口 + 工厂 |
| `internal/embeddings/openai_provider.go` | OpenAI-compatible embedding provider |
| `internal/embeddings/hash_provider.go` | Hash 降级 provider�?536 维输出） |
| `internal/embeddings/pgvector_search.go` | `VectorSearcher` 实现 |

### 修改文件

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| 文件 | 变更 |
|------|------|
| `infra/docker-compose.infra.yml` | postgres 镜像切换 |
| `cmd/main.go` | `EmbeddingCache` �?`VectorSearcher`，provider 初始�?|
| `internal/crawler/store.go` | `SaveEmbedding()` �?`pgvector.Vector` 替代 `pq.Array`；`LoadEmbeddings()` 移除（不再需要全量加载） |
| `internal/extractor/extractor.go` | step 7 embedding 生成改用 provider |
| `internal/embeddings/embedding.go` | `Dim` 常量改为 1536 或导出为可配�?|
| `shared/models/models.go` | `Embedding` 字段类型可能需调整 |

### 删除文件

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| 文件 | 原因 |
|------|------|
| `infra/postgres/Dockerfile` | 改用官方预编译镜�?|

### 保留不动的文�?
- `internal/embeddings/similarity.go` �?`CosineSimilarity()` 仍用于测�?- `internal/embeddings/models.go` �?`SimilarResult`、`SearchFilter` 结构体保�?- `internal/handler/similar_handler.go` �?接口不变，注入实现改�?`VectorSearcher`
- `internal/admin/admin_search.go` �?同上
- `internal/embeddings/cache_test.go` �?需更新或重写为 `pgvector_search_test.go`
- `internal/embeddings/similarity_test.go` �?不变

---

## 5. 测试计划

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

1. **OpenAI Provider 单元测试**：mock HTTP，验证请�?响应解析
2. **Hash Provider 单元测试**：验�?1536 维输出（�?256 非零，后 1280 零）
3. **VectorSearcher 单元测试**：mock `*sql.DB`，验�?SQL 查询构建
4. **集成测试**：需�?pgvector PostgreSQL 实例
5. **降级测试**：验�?provider 失败时回退到关键词搜索
6. **现有测试修复**：`cache_test.go` �?`pgvector_search_test.go`

---

## 6. 风险与缓�?
| 风险 | 缓解 |
|------|------|
| pgvector 镜像拉不下来 | 可用 `pgvector/pgvector:pg17` �?`ankane/pgvector` 替代 |
| �?embedding 数据丢失 | 旧数�?embedding 全部�?NULL，搜索走关键词降级，新写入自动用新模�?|
| OpenAI API 费用 | embedding 价格极低�?0.02/1M tokens），且只�?claim 入库时调�?|
| `lib/pq` 不支�?vector 类型 | 使用 `pgvector-go` 库的 `pgvector.Vector` 类型 + text cast |
| migration 失败（数据量大） | `UPDATE SET embedding = NULL` 大表可能慢，可分批处�?|
