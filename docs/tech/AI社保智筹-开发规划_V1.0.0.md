# AI社保智筹 — 开发规划 V1.1.0

| 文档属性 | 内容 |
| :--- | :--- |
| **版本号** | V1.1.0 |
| **状态** | 生效 |
| **发布日期** | 2026-07-03 |
| **更新说明** | V1.1.0: 更新Stage 1.0完成状态；新增Stage 2.0规划；新增沙盘模拟器/AI顾问服务；新增llm-gateway服务；更新数据覆盖目标。 |

## 1. 项目结构

```
nsi-platform/
├── services/
│   ├── api-server/                # Go/Gin — 用户、匹配、方案、合规、沙盘、AI顾问、订单
│   ├── policy-crawler/            # Go — 政策采集 + 验证 + 版本管理
│   ├── actuarial-engine/          # Go — NSGA-II优化 + 精算
│   └── llm-gateway/               # Go — 统一模型配置管理(4功能点)
├── frontend/
│   ├── ios/                       # SwiftUI — 原生 iOS (14页面)
│   ├── android/                   # Jetpack Compose — 原生 Android (14页面)
│   ├── weapp/                     # 微信小程序（原生，14页面+滑块验证码）
│   └── alipay/                    # 支付宝小程序（原生，14页面+滑块验证码）
├── shared/                        # Go 共享库（模型、加密、通知、配置）
├── scripts/                       # 工具脚本（数据同步等）
├── docs/                          # 文档库
├── docker-compose.yml
└── README.md
```

## 2. 微服务边界

| 服务 | 职责 | 外部依赖 | 数据库 |
|------|------|---------|--------|
| **api-server** | 用户注册登录、用户画像、政策匹配查询、方案展示、支付、合规引导、权益监测、社保沙盘模拟、AI顾问、订单管理、设置 | Redis(会话)、PG(数据)、actuarial-engine、llm-gateway | postgres(nsi_api) |
| **policy-crawler** | 多源政策采集、NLP结构化、交叉验证、置信度评分、人工审核、版本管理、ASR视频转录 | PG(政策库)、LLM API、Embedding API | postgres(nsi_crawler) |
| **actuarial-engine** | NSGA-II三目标帕累托优化(成本↓+养老金↑+公平性↑)、现金流预测、税务计算 | 无 | 无状态 |
| **llm-gateway** | 统一模型配置管理(4功能点: llm_extract/llm_plan/embedding/asr)、多Provider代理 | 火山引擎、DeepSeek、Ollama | postgres(nsi_llm) |
| **actuarial-engine** | 帕累托最优计算、三方案生成、现金流模拟、个税计算、灵敏度分析 | Redis(缓存)、PG(政策数据) | postgres(缓存结果) |

## 3. 开发阶段

### Phase 1（MVP）
参见 PRD V1.2.1 §3.1

Timeline: 目标 6-8 周

#### Sprint 1-2：基础设施 + 数据层
- [ ] 项目骨架搭建（Go mod、目录结构）
- [ ] PostgreSQL schema 设计与迁移（Goose）
- [ ] Docker Compose 本地开发环境
- [ ] GitHub Actions CI（lint + test + build）
- [ ] 共享库：数据模型、错误码、配置加载

#### Sprint 3-4：政策采集 MVG（Minimum Viable Gather）
- [ ] policy-crawler 基础爬虫（上海 5 个街道，硬编码演示）
- [ ] 政策文本 → LLM API → JSON 结构化管道
- [ ] 置信度评分初版（简化版，人工设为 1.0）
- [ ] 人工审核工作台 Web 页面（基础 CRUD）
- [ ] 验证：单条政策全流程走通

#### Sprint 5-6：核心匹配 + 方案生成
- [ ] api-server：用户画像 CRUD（认证由外部网关处理，从 `x-user-id` 获取身份）
- [ ] api-server：认证网关集成（读取 `x-user-id` 头，开发环境用 mock）
- [ ] api-server：政策查询接口（按城市+人群标签过滤）
- [ ] actuarial-engine：单城市养老金精算模型
- [ ] actuarial-engine：帕累托优化初版（基列举+多目标排序）
- [ ] api-server → actuarial-engine gRPC 集成

#### Sprint 7-8：前端 MVP + 联调
- [ ] 微信小程序：首页、信息采集、方案展示（MVP UI）
- [ ] iOS/Android：同上（MVP UI）
- [ ] 全链路联调：用户注册 → 填信息 → 匹配 → 方案 → 展示
- [ ] UAT 环境部署

### Phase 2（完整 V1.2.1）
- [ ] 剩余 4 城市政策采集（北京/深圳/广州/杭州）
- [ ] 完整帕累托（多险种组合）
- [ ] 合规性认定 + 材料清单
- [ ] 权益监测 + 风险预警
- [ ] 支付集成
- [ ] 支付宝小程序
- [ ] 完整 UI/UX（方案对比、图表、PDF 报告）

## 4. API 设计（高层）

### api-server 主要接口

> 注：用户认证（注册/登录/Token）由外部认证微服务提供。api-server 通过认证网关注入的 `x-user-id` 头获取用户身份。

```
GET    /v1/profile                 # 获取用户画像 (需认证, x-user-id)
PUT    /v1/profile                 # 更新用户画像

GET    /v1/policies                # 政策查询
  ?city=310000&tags=4050,flexible
  &status=verified

POST   /v1/plans/generate          # 触发方案生成（异步）
GET    /v1/plans/{id}              # 获取方案结果
GET    /v1/plans/{id}/report       # 获取 PDF 报告

POST   /v1/payments                # 创建支付订单
POST   /v1/payments/{id}/callback  # 支付回调

GET    /v1/alerts                  # 权益预警列表
```

### policy-crawler 内部接口

```
POST   /internal/v1/policies/ingest        # 人工录入政策
POST   /internal/v1/policies/{id}/verify   # 人工审核通过/驳回
GET    /internal/v1/policies/pending       # 待审核列表
POST   /internal/v1/sources               # 新增数据源配置
```

### actuarial-engine gRPC

```protobuf
service ActuarialEngine {
  rpc CalculatePlan(PlanRequest) returns (PlanResponse);
  rpc SensitivityAnalysis(SensitivityRequest) returns (SensitivityResponse);
}
```

## 5. 数据流向

```
用户 → 微信小程序 / iOS / Android App
         ↓ HTTPS
    认证网关 (外部微服务 / MVP 阶段简化版)
         ↓ x-user-id header
    api-server (Go/Gin)
         ↓              ↘
    policy-crawler    actuarial-engine
    (定时任务 + LLM)    (gRPC 计算)
         ↓                 ↓
    PostgreSQL 18 <——— pgvector 语义检索
         ↑
    Redis (缓存, 非会话)
```

## 6. CI/CD 流水线（GitHub Actions）

| 阶段 | 触发 | 操作 |
|------|------|------|
| lint | push/pr | golangci-lint |
| test | push/pr | go test ./... |
| build | push/pr | docker build |
| deploy UAT | push to main | Docker Compose 部署到 ECS |
| deploy Prod | tag release | K8s 部署（后续） |

## 7. 关键约束（SOP L1 强制执行）

- 端口分配：30000-50000 段
- API 版本：URL 路径带 `/v1/`
- API 废弃：旧版本保留 6 个月兼容期
- 数据库迁移：Goose，禁止手动改库
- 国密：SM4 加密敏感字段，TLS 1.3
- 密钥：KMS/环境变量注入，禁止明文（不留现有代码库的坑）
- 代码评审：PR 需至少 1 人审批
- TDD：红黄绿流程
