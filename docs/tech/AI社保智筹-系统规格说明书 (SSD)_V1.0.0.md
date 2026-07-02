# AI社保智筹 — 系统规格说明书 (SSD) V1.1.0

| 文档属性 | 内容 |
| :--- | :--- |
| **文档级别** | L4 (Standard) |
| **版本号** | V1.1.0 |
| **状态** | 生效 |
| **编写依据** | PRD V1.2.1 / 开发规划 V1.1.0 / SOP L1-一号文件 |
| **发布日期** | 2026-07-03 |
| **更新说明** | V1.1.0: 新增社保沙盘模拟器技术规格、AI顾问RAG架构、NSGA-II优化器规格、数据加密层设计、推送通知层设计；更新接口清单；V1.0.0归档。 |

---

## 目录

1. [系统概述](#1-系统概述)
2. [架构设计](#2-架构设计)
3. [组件规格](#3-组件规格)
4. [数据设计](#4-数据设计)
5. [接口规格](#5-接口规格)
6. [安全设计](#6-安全设计)
7. [部署设计](#7-部署设计)
8. [质量保障](#8-质量保障)
9. [运维设计](#9-运维设计)
10. [前瞻性设计](#10-前瞻性设计)
11. [附录](#11-附录)

---

## 1. 系统概述

### 1.1 产品定位

AI社保智筹是为 AI 时代灵活就业人员提供社保定制化筹划的智能平台。核心价值是解决政策信息不对称问题，通过 AI 技术实现街道级政策精准匹配与帕累托最优方案生成。

### 1.2 核心业务流程 → 技术映射

| 业务步骤 (PRD §4) | 技术服务 | 关键指标 | 外部依赖 |
|-------------------|---------|---------|---------|
| 用户认证 | 外部认证微服务 (待对接) | — | 未来对接 |
| 用户填写信息 → 画像构建 | api-server 用户画像 CRUD + LBS 定位 | 3 步完成基础画像 | 无 |
| 政策匹配 → 展示可申请补贴 | api-server 多维度标签过滤 + pgvector 语义检索 | 响应 < 1.5s |
| 方案生成 → 帕累托最优解集 | actuarial-engine gRPC 异步计算 | 响应 < 5s |
| 合规认定 → 材料清单 | api-server 动态渲染 policy `conditions` 字段 | 清单与政策库实时同步 |
| 权益监测 → 风险预警 | api-server 定时任务 + 模板消息推送 | 断缴前 7 天预警 |
| 付费 → 解锁完整报告 | api-server 支付回调 (Phase 2) | 支付后即时解锁 |

### 1.3 技术栈总览

| 层 | 技术 | 版本/选型理由 |
|---|------|-------------|
| 后端语言 | Go | SOP L1 强制，高并发 + 低内存 |
| Web 框架 | Gin | Go 生态最成熟，性能极佳 |
| 数据库 | PostgreSQL 18 + pgvector | 一库多用：关系数据 + 向量检索 |
| 缓存 | Redis 7 | 会话、热点政策、计算结果缓存 |
| 服务间通信 | gRPC (actuarial-engine) / 同库直读 (policy-crawler) | 高性能内部调用 |
| 消息队列 | Redis Streams (Dev/Test) → Kafka (Prod) | SOP L1 抽象接口 |
| 任务调度 | Asynq | SOP L1 强制 |
| 对象存储 | MinIO (Dev/Test) → 阿里云 OSS (Prod) | S3 协议兼容，零代码切换 |
| 移动端 iOS | Swift + SwiftUI | 原生性能最优 |
| 移动端 Android | Kotlin + Jetpack Compose | 原生性能最优 |
| 微信小程序 | 微信原生框架 | PRD 要求 |
| 支付宝小程序 | 支付宝原生框架 | PRD 要求 |
| LLM API | 通义千问 / GPT-4o | 政策文本 → JSON 结构化 |
| CI/CD | GitHub Actions | 决策确定 |
| 监控 | VictoriaMetrics + Vector + OpenTelemetry | SOP L1 强制 |
| 国密 | SM4 (数据加密) | SOP L1 §1.2.5 强制 |

---

## 2. 架构设计

### 2.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              客户端层                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ iOS App  │  │ Android  │  │ 微信小程序    │  │ 支付宝小程序      │   │
│  │ SwiftUI  │  │ Compose  │  │ 原生框架      │  │ 原生框架          │   │
│  └────┬─────┘  └────┬─────┘  └──────┬───────┘  └────────┬─────────┘   │
│       └──────────────┴──────────────┴────────────────────┘             │
│                              │ HTTPS                                    │
├──────────────────────────────┼─────────────────────────────────────────┤
│  ┌───────────────────────────┴──────────────┐                          │
│  │         认证网关 (未来对接外部微服务)       │                          │
│  │   MVP 阶段：简单 JWT 校验 + x-user-id     │                          │
│  │   未来：对接 account-center 微服务         │                          │
│  └──────────────────────┬───────────────────┘                          │
│                         │ x-user-id header                              │
├─────────────────────────┼───────────────────────────────────────────────┤
│                     API 网关 / SLB                                      │
├──────────────────────────────┼─────────────────────────────────────────┤
│                              ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     api-server (Go/Gin)                         │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │   │
│  │  │ 用户模块  │ │ 画像模块  │ │ 匹配模块  │ │ 合规/权益模块    │  │   │
│  │  └──────────┘ └──────────┘ └────┬─────┘ └──────────────────┘  │   │
│  │                                  │ gRPC                         │   │
│  └──────────────────────────────────┼──────────────────────────────┘   │
│                                     ▼                                   │
│  ┌──────────────────────────────┐  ┌──────────────────────────────┐   │
│  │   policy-crawler (Go)        │  │  actuarial-engine (Go)       │   │
│  │   ┌──────────────────────┐   │  │  ┌──────────────────────┐   │   │
│  │   │ 爬虫引擎 (定时任务)   │   │  │  │ 帕累托优化引擎       │   │   │
│  │   ├──────────────────────┤   │  │  ├──────────────────────┤   │   │
│  │   │ LLM 结构化管道       │   │  │  │ 现金流模拟器         │   │   │
│  │   ├──────────────────────┤   │  │  ├──────────────────────┤   │   │
│  │   │ 验证与置信度评分     │   │  │  │ 个税计算器           │   │   │
│  │   └──────────────────────┘   │  │  └──────────────────────┘   │   │
│  └──────────────────────────────┘  └──────────────────────────────┘   │
│                                     │                                  │
│  ┌──────────────────────────────────┼──────────────────────────────┐   │
│  │              PostgreSQL 18 + pgvector                            │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────────────────────┐  │   │
│  │  │ 用户数据    │ │ 政策数据    │ │ 方案快照/日志             │  │   │
│  │  └────────────┘ └────────────┘ └────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│  ┌──────────────────────────┐ ┌─────────────────────────────────┐      │
│  │     Redis 7              │ │     MinIO / 阿里云 OSS          │      │
│  │  会话/缓存/Streams       │ │   PDF 报告/政策附件             │      │
│  └──────────────────────────┘ └─────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 服务通信模式

| 通信对 | 协议 | 理由 |
|--------|------|------|
| 客户端 ↔ 认证网关 | HTTPS | 未来由外部 auth 微服务处理，当前用简单 JWT |
| api-server ↔ actuarial-engine | gRPC (同步) | 方案生成是同步等待的用户操作，需即时响应 |
| policy-crawler ↔ PostgreSQL | SQL (直读) | 单向数据写入，无中间服务依赖 |
| api-server ↔ PostgreSQL | SQL | 强一致数据操作 |
| api-server ↔ Redis | 原生协议 | 高性能缓存 + session |
| policy-crawler ↔ LLM API | HTTPS REST | 外部 API 调用 |

### 2.3 关键设计决策

| 决策 | 方案 | 理由 |
|------|------|------|
| 政策库共享 vs 独立 | policy-crawler 和 api-server 共享同一 PG 实例，但 schema 隔离 | 避免跨服务事务，同时保证数据一致性 |
| 精算同步 vs 异步 | gRPC 同步调用 (非 WebSocket) | 方案生成 < 5s，同步更简单，无状态利于横向扩展 |
| NLP 在线 vs 离线 | 离线（crawler 定时任务处理） | 政策非实时变更，离线处理成本低，可人工审核后再上线 |
| 用户画像存储 | PG 关系模型 (非文档型) | 画像字段稳定且关联查询多，关系模型更合适 |

---

## 3. 组件规格

### 3.1 api-server

**职责边界**：业务核心编排（不含用户认证，由外部微服务提供）

| 子模块 | 职责 | 关键接口 | 认证依赖 |
|--------|------|---------|---------|
| profile | 画像 CRUD + LBS 定位处理 | `GET|PUT /v1/profile` | 依赖外部 auth |
| policy-query | 政策检索 + 语义匹配（只读） | `GET /v1/policies` | 依赖外部 auth |
| plan-orchestration | 调用 actuarial-engine + 结果缓存 | `POST /v1/plans/generate` | 依赖外部 auth |
| compliance | 材料清单 + 认定条件渲染 | `GET /v1/compliance/*` | 依赖外部 auth |
| alert | 断缴计算 + 推送 | 内部定时任务 | 内部系统用户 |
| payment | 支付订单 + 回调 (Phase 2) | `POST /v1/payments/*` | 依赖外部 auth |
| admin | 运营后台接口 | `GET|POST /v1/admin/*` | 依赖外部 auth |

**认证集成策略**：
- **MVP 阶段**：认证网关从请求头 `Authorization: Bearer <token>` 中解析用户身份，将解析后的 `x-user-id`、`x-user-role` 等头注入转发到 api-server。api-server 信任网关注入的头，不再自行解析 JWT。
- **未来对接**：将认证网关替换为 account-center 微服务，实现注册/登录/Token 管理/权限控制，api-server 零代码变更。
- **公开接口**：`GET /v1/configs/*`、`GET /v1/policies`（匿名查询）无需认证，认证网关直接放行。

**关键设计**：
- 所有业务逻辑在 service 层，handler 只做参数校验和响应序列化
- actuator-engine 的 gRPC 客户端使用连接池，超时设 10s
- 用户身份通过请求头 `x-user-id` 获取，不可从 URL/body 参数中提取（防篡改）

### 3.2 policy-crawler

**职责边界**：政策全生命周期管理（采集 → 结构化 → 验证 → 发布）

| 子模块 | 职责 | 
|--------|------|
| scheduler | Asynq 定时任务管理器，控制各信源采集频率 |
| fetcher | HTTP 爬虫模块，支持 HIGH 源每日抓取、MEDIUM 源每周扫描 |
| parser | LLM API 调用层，将非结构化文本转为结构化 JSON |
| verifier | 置信度评分计算 + 交叉比对 + 三级入库决策 |
| admin-api | 人工审核工作台后端（嵌入 api-server admin 模块） |

**信源采集流程**：

```
定时触发 → 检查信源配置 → HTTP 请求 → 原始文本
    ↓
LLM API 提取 → {policy_id, region_code, policy_type, conditions, subsidy...}
    ↓
语义检索 pgvector(余弦相似度) → Top-K 相似政策
    ↓
字段级 Diff → 匹配/冲突计数 → 置信度计算
    ↓
≥ 0.85 → Verified → 写入 policy_claims + 版本快照
0.6~0.85 → Pending_Review → 写入审核队列
< 0.6 → Unverified → 保留历史
    ↓
人工审核工作台 → 确认/驳回/修改 → 更新置信度 → 发布
```

### 3.3 actuarial-engine

**职责边界**：纯计算服务，无状态，无数据库直连

| 子模块 | 职责 |
|--------|------|
| calculator | 基础养老金/医保计算器（按城市政策参数） |
| optimizer | 帕累托多目标优化（NSGA-II 或加权和） |
| cashflow | 年度现金流模拟 + 图表数据生成 |
| tax | 个税计算（工资薪金 + 劳务报酬 + 专项附加扣除） |

**帕累托优化流程**：

```
输入: {用户画像, 城市政策集合, 约束条件}
    ↓
基数枚举 [下限, 上限] 步长 100 元
    ↓
险种组合 [养老, 医疗, 失业] 各档位
    ↓
对每组 (基数, 险种):
    ├── 补贴计算 (政策条件匹配 → 补贴金额)
    ├── 权益评估 (预期养老金、医保报销比、积分)
    └── 成本计算 (月度缴纳额)
    ↓
多目标排序 → 帕累托前沿 → Top-5 推荐方案
    ↓
输出: [{方案详情, 现金流, 对比数据}]
```

**缓存策略**：
- 相同 `(城市, 年龄, 性别, 基数范围)` 的结果缓存 24h
- 政策参数变更时主动失效相关缓存
- Redis key 模式：`actuary:plan:{md5(params)}`

---

## 4. 数据设计

### 4.1 逻辑模型 (Policy Domain)

```
policy_sources            ← 信源配置（HIGH/MEDIUM/LOW + URL + 权重）
    │
    ├── policy_raw_texts  ← 原始文本（按次抓取结果 + version_hash）
    │
    └── policy_claims     ← 结构化政策原子（核心数据）
            │
            ├── policy_versions      ← 版本快照（不可变，按时间回溯）
            ├── policy_matches       ← 用户 × 政策匹配记录
            ├── verification_logs    ← 验证日志（置信度变更记录）
            └── user_feedback        ← 用户纠错反馈
```

### 4.2 逻辑模型 (User Domain)

```
users ← auth_base
  │
  └── profiles ← 画像（年龄/户籍/居住地/就业状态/技能/家庭）
        │
        ├── plan_snapshots ← 方案快照（包含政策版本引用 + 推荐方案 JSON）
        ├── payment_records ← 支付记录
        └── alert_subscriptions ← 预警订阅
```

### 4.3 schema.sql 关键表结构（迁移用 Goose）

详见 `docs/schema.sql`，以下为表名与用途速览：

| Schema | 表 | 用途 |
|--------|-----|------|
| public | users | 认证基础信息 |
| public | user_profiles | 用户画像扩展 |
| public | policy_sources | 政策数据源配置 |
| public | policy_raw_texts | 政策原始文本 |
| public | policy_claims | 结构化政策原子 |
| public | policy_versions | 政策版本快照 |
| public | verification_logs | 验证日志 |
| public | policy_matches | 用户×政策匹配 |
| public | plan_snapshots | 方案快照 |
| public | payment_records | 支付记录 |
| public | user_feedback | 用户反馈 |
| public | alert_subscriptions | 预警订阅 |

### 4.4 数据库迁移规范

- **工具**：Goose (Go)
- **目录**：`services/api-server/migrations/`, `services/policy-crawler/migrations/`
- **命名**：`YYYYMMDDHHMMSS_description.sql` (时间戳前缀)
- **每个迁移包含**：`-- +goose Up` 和 `-- +goose Down` 双向
- **禁止**：手动直连修改数据库结构

### 4.5 pgvector 向量检索设计

| 用途 | 嵌入内容 | 维度 | 索引类型 |
|------|---------|------|---------|
| 政策相似语义检索 | 政策 title + conditions + policy_type 拼接文本 | 1536 (text-embedding-3) | IVFFlat |
| 用户×政策匹配 | 用户标签向量 vs 政策 target_group_tags 向量 | 1536 | IVFFlat |

向量在 policy-crawler 的 LLM 管道中生成，随政策入库一起写入。

---

## 5. 接口规格

### 5.1 通用规范 (SOP L1 强制)

| 规范项 | 标准 |
|--------|------|
| URL 路径 | `/v1/{resource}` |
| 请求体 | JSON, Content-Type: `application/json` |
| 响应格式 | `{"code": 0, "data": {...}, "message": "ok"}` |
| 错误码 | `code != 0` 表示错误，`message` 描述 |
| 限流 | 响应头 `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` |
| 幂等 | 写操作必须携带 `Idempotency-Key` 请求头 |
| 认证 | 由外部认证网关处理。api-server 从请求头 `x-user-id` 获取用户身份。网关对匿名公开接口放行，其余接口校验 JWT 后转发。 |
| 用户身份传递 | 认证网关负责在转发的请求中注入 `x-user-id`、`x-user-role`。api-server **禁止**自行从请求中解析 Token。 |
| API 废弃 | 旧版本保留 6 个月，响应头标注 `Deprecated: true`，提前 30 天通知 |

### 5.2 核心 API 规范

> **注**：用户认证相关 API（注册/登录/Token 刷新/密码管理/登出）由外部认证微服务提供，不在本 SSD 范围内。api-server 通过认证网关注入的 `x-user-id` 头识别用户身份。

#### 用户画像

```yaml
GET /v1/profile
  响应: {"code": 0, "data": {"age": 35, "gender": "male", "region": "310000", ...}}

PUT /v1/profile
  请求: {"birth_date": "1990-01-15", "gender": "male", "region": "310000", "employment": "flexible"}
  响应: {"code": 0, "data": {"profile": {...}}}
```

#### 政策匹配

```yaml
GET /v1/policies?region=310000&tags=4050,flexible&status=verified
  响应: {
    "code": 0,
    "data": [
      {
        "policy_id": "sh-4050-001",
        "title": "灵活就业社会保险费补贴",
        "region": "310000",
        "policy_type": "养老保险补贴",
        "subsidy": {"method": "ratio", "rate": 0.5, "max_amount": 12000},
        "conditions": [...],
        "confidence": 0.95
      }
    ]
  }
```

#### 方案生成

```yaml
POST /v1/plans/generate
  请求: {"preferences": {"monthly_budget": 2000, "priority": "balance"}}
  响应: {
    "code": 0,
    "data": {
      "plan_id": "plan_abc123",
      "status": "completed",
      "schemes": [
        {
          "name": "balanced",
          "monthly_cost": 1850,
          "annual_subsidy": 7200,
          "details": {...},
          "cashflow": [...],
          "equity": {...}
        }
      ]
    }
  }
```

### 5.3 gRPC 定义 (actuarial-engine)

```protobuf
syntax = "proto3";
package actuarial;

service ActuarialEngine {
  rpc CalculatePlan (PlanRequest) returns (PlanResponse);
}

message PlanRequest {
  string region_code = 1;
  int32 age = 2;
  string gender = 3;
  string employment = 4;
  int32 contribution_years = 5;
  double current_balance = 6;
  double monthly_budget = 7;
  string priority = 8;   // "min_cost", "max_pension", "balance"
  double local_avg_salary = 9;
  repeated PolicyParam policies = 10;
}

message PolicyParam {
  string policy_id = 1;
  string calc_method = 2;
  double amount_min = 3;
  double amount_max = 4;
}

message PlanResponse {
  repeated Scheme schemes = 1;
  double calculation_time_ms = 2;
}

message Scheme {
  string name = 1;
  int32 base_salary = 2;        // 推荐缴费基数
  double monthly_cost = 3;      // 月缴费
  double annual_subsidy = 4;    // 年补贴
  double projected_pension = 5; // 预期月养老金
  repeated CashFlowItem cashflow = 6;
}

message CashFlowItem {
  int32 year = 1;
  double payment = 2;    // 年缴费
  double subsidy = 3;    // 年补贴
  double balance = 4;    // 账户余额
}
```

---

## 6. 安全设计

### 6.1 认证与授权

| 组件 | 方案 | 备注 |
|------|------|------|
| 用户认证 | 由外部认证微服务提供，不在本项目范围 | MVP 阶段由认证网关做简单 JWT 校验 |
| 身份传递 | 认证网关在请求头注入 `x-user-id`、`x-user-role` | api-server 信任网关，禁止自行解析 Token |
| Token 黑名单 | 由外部认证微服务管理，api-server 不参与 | 认证网关自行处理 |
| 密码/验证码 | 由外部认证微服务管理 | 不在本项目范围 |
| API 认证 | 认证网关拦截 + 转发。api-server 按 `x-user-id` 是否存在判断登录态 | 公开接口（政策查询等）无需认证 |
| 服务间认证 | gRPC mTLS 或内部网络隔离 | |

### 6.2 数据加密

| 数据类型 | 加密方式 | 备注 |
|---------|---------|------|
| 身份证号 | SM4 (应用层加密) | 存储为 ciphertext |
| 手机号 | SM4 | 脱敏输出 (138****8000) |
| 邮箱 | SM4 | 脱敏输出 |
| 银行卡号 | SM4 | 存储为 ciphertext |
| 传输链路 | TLS 1.3 | 全链路 HTTPS |
| 数据库 | PostgreSQL TDE | 启用透明加密 |

### 6.3 密钥管理

- 数据库密码、API Key、SM4 密钥**全部通过环境变量或 KMS 注入**
- docker-compose.yml 中只允许 `${VAR}` 引用，**禁止明文值**
- 本地开发使用 `.env` 文件（已加入 `.gitignore`）

### 6.4 隐私合规 (PIPL 映射)

| PIPL 要求 | 技术实现 |
|-----------|---------|
| 最小必要 | 仅采集 PRD §4.3.1 列表中的字段，拒绝非必要授权不影响基本功能 |
| 明示同意 | 首次启动展示用户协议 + 隐私政策，逐项授权弹窗 |
| 撤回同意 | 设置页提供授权管理开关，一键注销账号 |
| 数据删除 | 注销触发匿名化（SM4 字段置 NULL，保留匿名日志） |
| 数据本地化 | 所有服务器部署在中国境内，不使用境外云服务 |
| 数据可移植性 | 提供数据导出接口（JSON 格式） |

### 6.5 安全验收指标

| 验收项 | 标准 |
|--------|------|
| 渗透测试 | 第三方机构执行，零高危漏洞 |
| 敏感数据扫描 | 代码库中无明文密码/AK/SK/证书 |
| 依赖扫描 | `go mod verify` 零已知漏洞 |
| 输入校验 | 所有用户输入经过白名单校验，防 SQL/XSS 注入 |

---

## 7. 部署设计

### 7.1 环境矩阵

| 环境 | 架构 | 基础设施 | 数据库 | 缓存 | 对象存储 | 监控 | CI/CD |
|------|------|---------|--------|------|---------|------|-------|
| Dev | Docker Compose 独立项目 | 本地宿主机 | PG 18 (本地) | Redis 7 (本地) | MinIO (本地) | 无 | 手动 |
| Test | Docker Compose 独立项目 | 本地宿主机 | PG 18 (本地) | Redis 7 (本地) | MinIO (本地) | 无 | GitHub Actions |
| UAT | Docker Compose 共享平台 | 阿里云 ECS 单机 | PG 18 (自建) | Redis 7 (自建) | MinIO (自建) | VictoriaMetrics | GitHub Actions |
| Prod | K8s | 阿里云 ACK | RDS PG 高可用 | 云托管 Redis 集群 | 阿里云 OSS | VictoriaMetrics 集群 | GitHub Actions |

### 7.2 Docker Compose (UAT)

```yaml
# infra/docker-compose.uat.yml (示意)
services:
  postgres:  { image: postgres:18-alpine, volumes: [...] }
  redis:     { image: redis:7-alpine }
  minio:     { image: minio/minio }
  api-server:
    build: services/api-server
    environment:
      DATABASE_URL: ${DATABASE_URL}
      REDIS_URL: ${REDIS_URL}
      JWT_SECRET: ${JWT_SECRET}
      ACTUARY_GRPC_ADDR: actuarial-engine:50051
  policy-crawler:
    build: services/policy-crawler
    environment:
      LLM_API_KEY: ${LLM_API_KEY}
      DATABASE_URL: ${DATABASE_URL}
  actuarial-engine:
    build: services/actuarial-engine
    ports: ["50051:50051"]
```

### 7.3 CI/CD Pipeline (GitHub Actions)

```yaml
# .github/workflows/ci.yml (示意)
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps: [uses: golangci/golangci-lint-action]
  test:
    runs-on: ubuntu-latest
    services: { postgres, redis }
    steps: [run: go test ./... -race -count=1]
  build:
    runs-on: ubuntu-latest
    steps: [docker build, docker push]
```

### 7.4 端口分配

| 服务 | 端口 | 备注 |
|------|------|------|
| api-server | 30001 | 核心业务入口 |
| policy-crawler | 30002 | 内部管理接口 |
| actuarial-engine | 50051 | gRPC 端口 |
| PostgreSQL | 35432 | 非标准端口防扫描 |
| Redis | 36379 | 非标准端口防扫描 |

---

## 8. 质量保障

### 8.1 测试策略

| 测试层级 | 覆盖范围 | 工具 | 目标覆盖率 |
|---------|---------|------|-----------|
| 单元测试 | 核心算法、工具函数 | Go testing | ≥ 80% |
| 集成测试 | API handler + 数据库 | testify + testcontainers-go | 每条 API ≥ 1 个集成用例 |
| 契约测试 | gRPC client/server | (放在集成测试中) | 所有 gRPC 方法 |
| E2E 测试 | 关键用户路径 | Playwright / 小程序自动化 | MVP 核心路径 |
| 压力测试 | 方案生成、政策匹配 | k6 / wrk | P99 < 5s at 500 TPS |

### 8.2 PRD 验收指标 → 技术验证方法

| PRD §8 验收项 | 技术验证方法 | 通过标准 |
|--------------|------------|---------|
| 政策匹配准确率 ≥ 99% | 人工抽检 100 条置信度 ≥ 0.85 的政策，记录错误数 | 错误 ≤ 1 条 |
| 方案金额误差 < 0.5% | 用已知计算案例的输入输出做回归测试 | 所有案例通过 |
| HIGH 源更新 < 12h | 监控 crawler 采集时间戳 vs 政策发布时间 | 90% 分位 < 12h |
| LBS 街道级识别 | 模拟 10 个已知坐标点，验证返回的 region_code | 100% 正确 |
| 并发 5000 TPS | k6 压力测试脚本持续 5 分钟 | P99 响应 < 1.5s, 0 错误 |
| 支付后即时解锁 | E2E 测试模拟支付回调 → 查询方案状态 | 回调后 < 1s 解锁 |
| 渗透测试零高危 | 第三方安全机构报告 | 高危 = 0 |

### 8.3 质量门禁 (SOP L1 强制)

- 单元测试覆盖率 **≥ 80%**
- PR 提交后 **1 个工作日内** 完成评审
- PR 需 **至少 1 人审批** 方可合并
- 安全扫描 **零高危漏洞**
- 任务粒度 **30 分钟内可完成**

---

## 9. 运维设计

### 9.1 监控与告警

| 指标 | 采集方式 | 告警阈值 |
|------|---------|---------|
| API P99 响应时间 | OpenTelemetry 埋点 → VictoriaMetrics | > 3s 持续 5min |
| 服务 CPU 使用率 | Vector 采集 → VictoriaMetrics | > 80% 持续 10min |
| 政策采集成功率 | crawler 业务指标 | < 90% 持续 1h |
| 验证码发送失败率 | api-server 业务指标 | > 5% 持续 5min |
| 精算 gRPC 错误率 | actuarial-engine 业务指标 | > 1% 持续 5min |
| 数据库连接数 | PG 内置监控 | > 80% 最大连接数 |

### 9.2 日志规范

| 日志类型 | 格式 | 存储 | 保留周期 |
|---------|------|------|---------|
| 业务日志 | JSON (level, service, trace_id, message) | Vector → 文件 | 30 天 |
| 访问日志 | JSON (method, path, status, latency) | Vector → 文件 | 30 天 |
| 审计日志 | JSON (user_id, action, target, timestamp) | PostgreSQL audit_log 表 | 永久（不可删改） |
| 慢查询日志 | PG 原生 | 文件 | 7 天 |

### 9.3 灾备与恢复

| 项目 | 目标 |
|------|------|
| RPO | ≤ 1 小时（生产环境） |
| RTO | ≤ 30 分钟（生产环境） |
| 备份策略 | 每日自动全量备份 + WAL 持续归档 |
| 恢复演练 | 每季度至少执行 1 次 |

---

## 10. 前瞻性设计

本章记录为保证系统长期可扩展性而预先作出的设计决策。这些设计在 MVP 阶段可能只用到一部分，但架构已为之留好接口。

### 10.1 认证解耦

| 当前 (MVP) | 未来 |
|-----------|------|
| 认证网关做简单 JWT 校验，注入 `x-user-id` | 对接 account-center 微服务，支持注册/登录/OAuth/SSO |
| api-server 无感，通过 `x-user-id` 头获取用户身份 | 无代码变更，仅替换认证网关实现 |

**设计原则**：api-server 对认证机制零依赖。认证是独立的安全层，可被替换、升级、拆分而不影响业务代码。

### 10.2 事件驱动架构预留

当前服务间以同步调用（gRPC/SQL）为主。未来当服务数量增长（如新增通知服务、数据分析服务、B2B 网关），同步调用将形成网状依赖，需引入事件总线。

**预留设计**：

```go
// shared/eventbus/eventbus.go (接口抽象)
type EventBus interface {
    Publish(ctx context.Context, topic string, event interface{}) error
    Subscribe(topic string, handler Handler) (Subscription, error)
}

// MVP 实现：Redis Streams (embedded)
// 未来实现：Kafka / RocketMQ (零代码变更)
```

**关键业务事件**（当前只记日志，未来可订阅）：

| 事件 | 触发点 | 未来可能的订阅者 |
|------|--------|----------------|
| `policy.updated` | policy-crawler 政策变更发布 | 通知服务（推送用户）、分析服务 |
| `plan.generated` | actuarial-engine 返回方案 | B2B 网关、数据服务 |
| `user.feedback` | 用户上报政策偏差 | 审核工作台、信源权重调整服务 |
| `alert.triggered` | 断缴/政策变更预警触发 | 通知服务、运营看板 |

### 10.3 服务网格准备

当前 3 个服务通过 Docker Compose 内部网络通信。未来扩展至 10+ 微服务时，建议引入服务网格（Service Mesh）。

**预留设计**：
- 所有服务间通信使用 HTTP/gRPC 标准协议，避免私有协议
- 健康检查端点统一为 `/healthz` 和 `/readyz`
- 链路追踪通过 `x-trace-id` / `x-span-id` 请求头传递（OpenTelemetry 标准）
- 服务发现通过环境变量配置（静态），未来迁移到 Consul/Nacos 零代码变更

### 10.4 政策信源插件化

当前直接内置 HIGH/MEDIUM/LOW 三级的采集逻辑。未来新信源类型（如政府 API 直连、第三方数据供应商）应可通过插件注册。

**预留设计**：

```go
// shared/crawler/plugin.go (接口抽象)
type SourcePlugin interface {
    Name() string
    Level() SourceLevel
    Fetch(ctx context.Context) ([]RawDocument, error)
    Parse(raw RawDocument) (*PolicyClaim, error)
}

// 注册机制
var registry = make(map[string]SourcePlugin)
func Register(plugin SourcePlugin) { registry[plugin.Name()] = plugin }
```

当前 policy-crawler 内置的 HIGH/MEDIUM/LOW 逻辑各自实现 `SourcePlugin` 接口。新信源只需实现该接口并调用 `Register()`，无需修改核心管道。

### 10.5 多租户预留

当前仅服务 C 端个人用户。未来 B2B 模式（与人力资源平台、行业协会合作）需要多租户隔离。

**预留设计**：

```go
// 所有核心表预留 tenant_id 字段 (MVP 阶段可为空或默认值)
type UserProfile struct {
    TenantID  string `db:"tenant_id"`   // MVP: "default"
    UserID    string `db:"user_id"`
    Age       int    `db:"age"`
    // ...
}
```

- 查询层：所有业务查询 WHERE `tenant_id = ?`
- Middleware：从请求头 `x-tenant-id` 提取租户上下文
- MVP 阶段 `tenant_id` 为 "default"，B2B 阶段按租户隔离数据和配置

### 10.6 多区域部署预留

当前仅部署在国内单区域。未来如扩展至港澳台或其他国家，需支持：

**预留设计**：
- 政策数据和行政区划代码与区域绑定（`region_code` 字段已设计）
- 数据库按区域分库，应用层无状态可水平扩展
- 配置文件与环境绑定，同一份代码通过配置适配不同区域的政策规则

### 10.7 B2B Open API 预留

未来开放平台需要对外提供 API 给合作伙伴。

**预留设计**：
- API 路径已统一为 `/v1/{resource}`，版本化从第一天开始
- `Idempotency-Key` 幂等机制已内置，满足 B2B 对账需求
- 限流响应头已规范，B2B 场景可基于 `X-Api-Key` 做按客户限流
- 数据模型已预留 `tenant_id`，B2B 客户的策略和数据天然隔离

### 10.8 已有设计中的前瞻性元素（汇总）

| 设计元素 | 对应 SSD 章节 | 为未来预留的能力 |
|---------|-------------|----------------|
| pgvector 嵌入语义检索 | §4.5 | 支持未来更复杂的 ML 匹配模型 |
| 政策版本快照不可变 | §4.1 | 审计追溯 + B2B 合规要求 |
| MinIO → OSS 协议一致 | §7.1 | 环境切换零代码变更 |
| Redis Streams → Kafka 接口抽象 | §10.2 | 消息中间件无感替换 |
| gRPC 服务定义先行 | §5.3 | 跨语言服务调用标准化 |
| 精算结果缓存策略 | §3.3 | 高并发场景下降级可用 |

---

## 11. 附录

### A. 与 PRD 需求追踪矩阵的映射

| PRD §A 功能需求编号 | SSD 对应章节 | 验收方法 (SSD §8.2) |
|---------------------|-------------|-------------------|
| 1.1 政策数据自动化采集 | §3.2 policy-crawler | HIGH 源 < 12h |
| 1.2 政策文本AI结构化 | §3.2 parser (LLM管道) | 人工抽检 100 条 |
| 2.1 基于LBS区域匹配 | §3.1 api-server policy-query | LBS 10 点 100% |
| 3.1 帕累托基数推荐 | §3.3 actuarial-engine optimizer | 金额误差 < 0.5% |
| 4.1 条件与材料清单 | §3.1 api-server compliance | 动态渲染验证 |
| 8.1 支付后即时解锁 | §5.2 API 方案生成 | E2E 测试 < 1s |
| 9.1 社保沙盘模拟器 | §3.5 simulator + §5.3 API | 150ms防抖+实时计算 |
| 9.2 AI社保顾问 | §3.6 advisor RAG | ≤3句+引用来源 |
| 10.1 数据加密(AES-256) | §3.7 crypto | 加解密一致性 |
| 10.2 推送通知 | §3.8 notifier | SMS/模板消息降级 |

### B. 参考文档

- PRD V1.2.1 — 产品需求说明书
- BRD V1.2.0 — 业务需求说明书
- 开发规划 V1.1.0 — 开发阶段与排期
- SOP L1-一号文件 v2.3.0 — 系统开发与文档管理规范
