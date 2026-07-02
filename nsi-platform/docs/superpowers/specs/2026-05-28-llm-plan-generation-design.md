# LLM 方案生成系统设计

**日期**: 2026-05-28
**状�?*: 已批�?

## 目标

将社保方案生成从确定性精算模型（actuarial-engine）改�?LLM 智能生成，实现：
1. 方案完全�?LLM 根据用户画像 + 政策库数据定�?
2. 政策依据仅来自系统爬�?提取的真实政策数�?
3. 方案包含所有利用到的政策依据（原文片段+链接，真实可查）
4. 精算引擎保留作为双向验证�?

## 架构

新增独立微服�?`llm-gateway`（端�?39404），统一管理 LLM 调用�?

```
┌──────────────�?    ┌──────────────�?    ┌──────────────�?
�? api-server  │────▶│ llm-gateway  │────▶│ LLM Provider �?
�? :39401      �?    �? :39404      �?    �?(DeepSeek�? �?
└──────┬───────�?    └──────┬───────�?    └──────────────�?
       �?                   �?
       �?                   �?
┌──────────────�?    ┌──────────────�?
�? nsi_api DB  �?    �? nsi_llm DB  �?
└──────────────�?    └──────────────�?

actuarial-engine :39402 保留，用作数值验�?
```

## 新建服务：llm-gateway

### 目录结构

```
services/llm-gateway/
├── cmd/main.go
├── internal/
�?  ├── gateway/
�?  �?  ├── gateway.go             # 核心路由：provider选择、fallback、限�?
�?  �?  └── gateway_test.go
�?  ├── provider/
�?  �?  ├── provider.go            # Provider接口定义
�?  �?  ├── openai_compat.go       # DeepSeek/火山方舟等OpenAI兼容
�?  �?  ├── bailian.go             # 阿里云百�?
�?  �?  └── provider_test.go
�?  ├── config/
�?  �?  ├── config.go              # DB-backed provider配置CRUD
�?  �?  └── config_test.go
�?  ├── admin/
�?  �?  ├── admin_handler.go       # 管理页面+API
�?  �?  └── admin_page.go          # 管理UI HTML（内嵌，同crawler模式�?
�?  └── usage/
�?      ├── usage.go               # 用量统计记录
�?      └── usage_test.go
├── migrations/
�?  ├── 001_providers.sql
�?  └── 002_usage_logs.sql
└── Dockerfile
```

### API 端点

**公共端点**�?
- `POST /v1/chat` �?发�?prompt，返�?LLM 响应
  - 请求：`{"model_preference":"deepseek","system_prompt":"...","user_content":"...","max_tokens":4096,"caller":"api-server"}`
  - 响应：`{"content":"...","provider_used":"deepseek","model":"deepseek-v4-flash","tokens_in":500,"tokens_out":2000,"latency_ms":1200}`
- `GET /healthz`

**管理员端�?*（Basic Auth）：
- `GET /admin/providers` �?当前 provider 配置列表
- `POST /admin/providers` �?保存 provider 配置
- `POST /admin/providers/test` �?测试 provider 连通�?
- `GET /admin/usage` �?用量统计（按 provider/日期 聚合�?
- `GET /admin/` �?管理页面 HTML

### 支持�?Provider

| Provider | endpoint | model | 格式 |
|----------|----------|-------|------|
| DeepSeek | `https://api.deepseek.com/v1/chat/completions` | `deepseek-v4-flash` | OpenAI 兼容 |
| 阿里云百�?| `https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions` | `qwen3.6-plus` | OpenAI 兼容 |
| 火山方舟 | `https://ark.cn-beijing.volces.com/api/v3/chat/completions` | `doubao-pro-32k` | OpenAI 兼容 |
| OpenCode Go | `http://localhost:11434/v1/chat/completions` | `opencode-go` | OpenAI 兼容 |

### Fallback 策略

- �?`priority` 字段排序（越小越优先�?
- �?provider 请求失败（网络错�?HTTP �?200/空响应）自动切到下一�?
- 每次调用记录使用�?provider 和是�?fallback

### DB Schema

```sql
-- nsi_llm database

CREATE TABLE llm_providers (
  id SERIAL PRIMARY KEY,
  provider_name VARCHAR(50) NOT NULL UNIQUE,
  api_key TEXT NOT NULL,
  endpoint TEXT NOT NULL,
  model_name VARCHAR(100) NOT NULL,
  max_tokens INT DEFAULT 4096,
  is_primary BOOLEAN DEFAULT false,
  is_enabled BOOLEAN DEFAULT true,
  priority INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE llm_usage_logs (
  id BIGSERIAL PRIMARY KEY,
  provider_name VARCHAR(50),
  model VARCHAR(100),
  caller VARCHAR(50),
  tokens_in INT,
  tokens_out INT,
  latency_ms INT,
  status VARCHAR(20),  -- success / fallback / failed
  error_message TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## 改造：api-server 方案生成

### 新的 GeneratePlanHandler 流程

```
1. 解析请求 + 获取用户画像
2. 检索三级政策（国家/�?�?区）+ 排除已过�?
3. 组装 LLM Prompt（用户画�?+ 政策数据 + 输出格式要求�?
4. 调用 llm-gateway POST /v1/chat
5. 解析 LLM 响应（Part A 自由文本 + Part B 结构化JSON�?
6. [异步] 精算引擎双向验证
7. 保存方案快照
8. 返回双视图方�?+ 政策依据
```

### 政策检索改�?

新增 `QueryByRegionHierarchy(ctx, regionCode, status)` 方法�?

```sql
-- 示例：用�?regionCode = '310115'
-- 查询国家�?000000) + 省级(310000) + 市级(310100) + 区级(310115)
SELECT ... FROM policy_claims
WHERE region_code IN ('000000', '310000', '310100', '310115')
  AND status = 'verified'
  AND (expire_date IS NULL OR expire_date > NOW())
ORDER BY
  CASE region_code
    WHEN '000000' THEN 1   -- 国家级优先级最低（兜底�?
    WHEN '310000' THEN 2
    WHEN '310100' THEN 3
    WHEN '310115' THEN 4   -- 区级最具体，优先级最�?
  END
```

### LLM Prompt 设计

**System Prompt**�?
```
你是一位资深社保政策顾问。根据用户的个人情况和所在地的社保政策，
为用户量身定制最优社保参保方案�?

规则�?
1. 所有政策依据必须来自提供的政策库数据，不可编�?
2. 必须理解上位法与下位法的关系：国家法�?> 省级规定 > 市级细则 > 区级优惠
3. 当地方政策与上位法冲突时，以有利于用户的原则解释
4. 方案必须包含所有适用的优惠政策，不能遗漏
5. 每一条建议都必须标注所依据的政策（标题+文号+链接�?
6. 数值计算必须精确，基于提供的费率和基数

输出格式（必须严格遵守）�?
===FREE_FORM_START===
[自由文本格式的方案建议书，面向普通用户，通俗易懂�?000字以内]
===FREE_FORM_END===
===STRUCTURED_START===
{
  "summary": "一句话总结",
  "schemes": [
    {
      "name": "方案名称",
      "description": "方案描述",
      "monthly_cost": 0,
      "annual_subsidy": 0,
      "projected_pension": 0,
      "total_cost": 0,
      "contribution_base": 0,
      "pension_employee_rate": 0,
      "pension_employer_rate": 0,
      "medical_employee_rate": 0,
      "analysis": "该方案的详细分析",
      "applicable_policies": ["claim_id_1"]
    }
  ],
  "policy_references": [
    {
      "claim_id": "",
      "policy_title": "",
      "document_number": "",
      "policy_url": "",
      "relevant_excerpt": "提取的原文片�?,
      "how_applied": "如何应用于本方案"
    }
  ],
  "recommendation": {
    "recommended_scheme": "方案名称",
    "reasoning": "推荐理由"
  }
}
===STRUCTURED_END===
```

**User Content**�?
```
## 用户画像
{用户画像JSON}

## 适用政策（共N条）
{政策数据JSON数组，每条包含：claim_id, policy_title, region_code, policy_type,
 subsidy_calc_method, subsidy_amount_min/max, conditions, document_number,
 policy_url, source_url, effective_date, expire_date, target_group_tags}

## 请基于以上信息，为该用户生成最优社保参保方案�?
```

### 新的数据模型

```go
type LLMScheme struct {
    Name                string   `json:"name"`
    Description         string   `json:"description"`
    MonthlyCost         float64  `json:"monthly_cost"`
    AnnualSubsidy       float64  `json:"annual_subsidy"`
    ProjectedPension    float64  `json:"projected_pension"`
    TotalCost           float64  `json:"total_cost"`
    ContributionBase    float64  `json:"contribution_base"`
    PensionEmployeeRate float64  `json:"pension_employee_rate"`
    PensionEmployerRate float64  `json:"pension_employer_rate"`
    MedicalEmployeeRate float64  `json:"medical_employee_rate"`
    Analysis            string   `json:"analysis"`
    ApplicablePolicies  []string `json:"applicable_policies"`
}

type PolicyReference struct {
    ClaimID          string `json:"claim_id"`
    PolicyTitle      string `json:"policy_title"`
    DocumentNumber   string `json:"document_number"`
    PolicyURL        string `json:"policy_url"`
    RelevantExcerpt  string `json:"relevant_excerpt"`
    HowApplied       string `json:"how_applied"`
}

type LLMSchemeResponse struct {
    Summary          string            `json:"summary"`
    Schemes          []LLMScheme       `json:"schemes"`
    PolicyReferences []PolicyReference `json:"policy_references"`
    Recommendation   struct {
        RecommendedScheme string `json:"recommended_scheme"`
        Reasoning         string `json:"reasoning"`
    } `json:"recommendation"`
}
```

### 扩展 PlanSnapshot

```go
type PlanSnapshot struct {
    PlanID               string             `json:"plan_id"`
    UserID               string             `json:"user_id"`
    FreeFormText         string             `json:"free_form_text"`
    StructuredSchemes    []LLMScheme        `json:"structured_schemes"`
    PolicyReferences     []PolicyReference  `json:"policy_references"`
    Recommendation       string             `json:"recommendation"`
    RecommendationReason string             `json:"recommendation_reason"`
    VerificationResult   *VerificationResult `json:"verification_result,omitempty"`
    GeneratedAt          time.Time          `json:"generated_at"`
}

type VerificationResult struct {
    Status       string            `json:"status"` // pass / warning / failed
    MaxDeviation float64           `json:"max_deviation_pct"`
    Details      []DeviationDetail `json:"details"`
}
```

## 精算引擎双向验证

### 核心原则

**不预设哪一方是对的**。偏差发现后，双向排查客观归因：

```
LLM输出数�?←比对→ 精算引擎计算结果
         �?
    偏差 > 阈�?5%)�?
    ├─ �?�?通过，记录日�?
    └─ �?�?记录偏差详情，进入根因分析：
         ├─ 1. 检查LLM是否正确理解了政策费率（对照政策原文验证�?
         ├─ 2. 检查精算引擎的参数是否过时（如 hardcoded 默认值）
         ├─ 3. 检查两者使用的政策数据是否一致（新政策未同步�?
         ├─ 4. 检查精算引擎的计算公式是否覆盖了特殊优惠场�?
         └─ 归因�?
              ├─ llm_error �?优化 prompt / 增加 few-shot 示例
              ├─ actuary_bug �?修复精算引擎代码
              ├─ data_sync �?修复数据同步问题
              └─ policy_mismatch �?完善政策数据
```

### 偏差日志�?

```sql
-- �?nsi_api �?
CREATE TABLE plan_verification_logs (
  id BIGSERIAL PRIMARY KEY,
  plan_id VARCHAR(100),
  llm_provider VARCHAR(50),
  llm_scheme_name VARCHAR(200),
  metric VARCHAR(50),              -- monthly_cost / projected_pension �?
  llm_value DECIMAL(15,2),
  actuary_value DECIMAL(15,2),
  deviation_pct DECIMAL(5,2),
  root_cause VARCHAR(50),          -- llm_error / actuary_bug / data_sync / policy_mismatch / unknown
  resolution TEXT,
  resolved BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 管理页面

llm-gateway 管理页面（`http://localhost:39404/admin/`）包含：
- **Provider 配置** tab：多 provider 列表，启�?禁用，优先级，API Key/Endpoint/Model
- **连通性测�?*：一键测试各 provider
- **用量统计** tab：按�?�?provider �?token 消耗图表（Chart.js�?
- **偏差分析** tab：偏差列表（按严重度排序），管理员标�?root_cause �?resolution，LLM/精算/数据问题比例统计

## WebClient 改�?

**方案生成 tab 改�?*�?
- 增加"视图切换"下拉：`自由文本` / `结构化分析`
- 自由文本视图：渲�?`free_form_text`（支�?markdown 格式�?
- 结构化视图：渲染 schemes 列表 + 每个方案�?analysis 文字 + 适用政策卡片
- 政策依据区域：policy_title + document_number + 可点�?policy_url 链接 + relevant_excerpt 原文片段 + how_applied 应用说明
- 推荐方案高亮 + 推荐理由

## Docker 集成

### docker-compose.yml 新增

```yaml
llm-gateway:
  build:
    context: ./services/llm-gateway
    dockerfile: Dockerfile
  ports:
    - "39404:39404"
  environment:
    DATABASE_URL: postgres://nsi:nsi_pass@postgres:5432/nsi_llm?sslmode=disable
    SERVER_PORT: 39404
    ADMIN_USER: admin
    ADMIN_PASS: "${LLM_GATEWAY_ADMIN_PASS:-changeme}"
  depends_on:
    - postgres

# api-server 新增环境变量
api-server:
  environment:
    LLM_GATEWAY_URL: http://llm-gateway:39404
```

### PostgreSQL 初始�?

新增 `nsi_llm` database，migration 自动创建 llm_providers �?llm_usage_logs 表�?

## 改动范围总结

| 组件 | 改动类型 | 说明 |
|------|---------|------|
| `services/llm-gateway/` | **新建** | 独立 LLM 代理服务，多 provider + fallback + 用量统计 |
| `services/api-server/internal/handler/plan_handler.go` | **重写** | 改为检索三级政�?+ �?llm-gateway + 解析双视�?|
| `services/api-server/internal/handler/webclient_handler.go` | **改�?* | 双视图切�?+ 政策依据展示 |
| `services/api-server/internal/repository/policy_repo.go` | **新增方法** | `QueryByRegionHierarchy` |
| `shared/models/models.go` | **更新** | 新增 LLMScheme/PolicyReference/VerificationResult |
| `docker-compose.yml` | **更新** | 新增 llm-gateway 容器 + api-server 环境变量 |
| `docker/postgres-init/` | **更新** | 新增 nsi_llm 库初始化 |

## 约束

- LLM 配置由管理员�?llm-gateway 后台设置，不硬编�?
- 政策依据只引用系统中已提取的 policy_claims，LLM 不可编造政�?
- 精算引擎保留运行，仅用作验证层，不参与方案决�?
- 偏差处理遵循双向排查原则，不预设任何一方有问题
- `GOOS=linux GOARCH=amd64` 交叉编译
- `GOPROXY=https://goproxy.cn,direct`
