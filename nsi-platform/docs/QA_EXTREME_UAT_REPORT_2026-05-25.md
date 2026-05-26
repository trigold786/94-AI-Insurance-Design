# 极致深度测试与 UAT 验收报告 (Extreme QA Report)

## 1. 测试执行概览
- **项目名称**: NeuroSocialInsurance (神经社保智能平台)
- **验收目标**: 100% 符合 PRD 验收标准
- **测试深度**: 覆盖业务逻辑、底层鲁棒性、安全渗透、UX 交互、极端环境及并发性能
- **审计范围**: api-server (Go), actuarial-engine (Go), policy-crawler (Go), 前端 (微信小程序/支付宝小程序/WebClient), Docker 基础设施
- **执行日期**: 2026-05-25
- **执行状态**: **发现风险 — 不可交付，需修复 P0 级问题后方可上线**

---

## 2. 深度测试矩阵 (Deep Testing Matrix)

| 需求 ID | 需求描述 | 验收标准 (AC) | 安全审计 | 鲁棒性检查 | 性能表现 | 验收状态 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| SEC-01 | 用户认证 | JWT/Session 鉴权 | **FAIL**: 使用 x-user-id 头伪造身份，零加密验证 | 任何用户可冒充任意身份 | N/A | **P0 阻塞** |
| SEC-02 | 管理面板安全 | 管理接口需鉴权 | **FAIL**: /admin/* 全部无认证 | 519条政策数据、LLM API Key 完全暴露 | N/A | **P0 阻塞** |
| SEC-03 | API Key 保护 | 密钥加密存储/传输 | **FAIL**: DeepSeek API Key 明文存储+明文返回 | `sk-f22c79fa7bf846c3a41ff31a0398a733` 已泄露 | N/A | **P0 阻塞** |
| SEC-04 | 数据库安全 | 非默认密码+SSL | **FAIL**: 默认密码 postgres123, sslmode=disable | 连接明文传输 | N/A | **P0 阻塞** |
| SEC-05 | CORS 策略 | 限定可信源 | **FAIL**: Access-Control-Allow-Origin: * | 任意网站可跨域请求+伪造用户 | N/A | **P0 阻塞** |
| SEC-06 | IDOR 防护 | 资源归属校验 | **FAIL**: MarkAlertRead 无 ownership 校验 | 用户 A 可标记用户 B 的提醒 | N/A | **HIGH** |
| SEC-07 | 速率限制 | 防暴力/防DoS | **FAIL**: 零速率限制 | 无限调用方案计算引擎 | N/A | **HIGH** |
| SEC-08 | CSRF 防护 | SameSite/Token | **FAIL**: 无 CSRF 保护 | POST/PUT/DELETE 跨站可操作 | N/A | **MEDIUM** |
| SQL-01 | SQL 注入防护 | 参数化查询 | **PASS**: 全部使用 $1/$2 占位符 | 边界值未触发注入 | N/A | **PASS** |
| SQL-02 | 输入校验 | 字段格式验证 | **PARTIAL**: Age 有范围校验, 其他字段缺校验 | monthly_income=-1 未拦截 | N/A | **MEDIUM** |
| BIZ-01 | 方案生成 | 政策数据驱动计算 | **PASS**: 从 policy_claims 提取费率 | 极端值(age=999)已拦截 | 200ms 内响应 | **PASS** |
| BIZ-02 | 精算引擎 | 纯数学无硬编码 | **PASS**: city.go 已移除 | 多方案+现金流计算正确 | <100ms | **PASS** |
| BIZ-03 | 政策爬取 | 自动爬取+LLM提取 | **PASS**: 14源519条 | Chromium 渲染正常 | N/A | **PASS** |
| UX-01 | 前端连通 | API调用正确 | **FAIL**: 3端硬编码 127.0.0.1:39401 | 生产环境无法连接 | N/A | **HIGH** |
| UX-02 | 安全响应头 | CSP/X-Frame/X-CTO | **FAIL**: 全部缺失 | XSS/点击劫持风险 | N/A | **HIGH** |
| INF-01 | 容器安全 | 非 root 运行 | **FAIL**: 全部 root 运行 | 容器逃逸风险 | N/A | **HIGH** |
| INF-02 | 优雅关闭 | SIGTERM 安全处理 | **FAIL**: 使用 log.Fatal+ListenAndServe | 在途请求被丢弃 | N/A | **MEDIUM** |
| INF-03 | 连接池 | 数据库连接管理 | **FAIL**: 未配置 MaxOpen/MaxIdle/MaxLifetime | 高并发下连接耗尽 | N/A | **MEDIUM** |
| INF-04 | 请求超时 | Read/Write/Idle | **FAIL**: 未配置 Server 超时 | Slowloris 攻击风险 | N/A | **MEDIUM** |
| INF-05 | 资源限制 | Docker mem/cpu | **FAIL**: 无限制 | OOM 风险 | N/A | **MEDIUM** |

---

## 3. 问题发现与风险预警 (Issue Tracker)

### 3.1 P0 — 严重漏洞 (Critical/Blocker) — 共 5 项

#### C-01: 身份认证形同虚设 — x-user-id 头伪造
> **位置**: `shared/middleware/middleware.go:24-35`
> **描述**: 认证中间件仅检查 `x-user-id` 头是否存在，无 JWT/Session/签名验证。config.go 中定义了 JWTSecret 但从未使用。
> **复现 Payload**:
> ```bash
> curl -H "x-user-id: admin" http://localhost:39401/v1/profile
> # 返回 200 + 完整用户资料
> ```
> **影响**: 全量用户身份冒充，任意用户资料/方案/缴费/提醒可读写。
> **修复建议**: 实现 JWT Bearer Token 验证，从 token claims 提取 userID。

#### C-02: 管理面板零认证
> **位置**: `policy-crawler/cmd/main.go:90-111`
> **描述**: 10+ 个 /admin/* 端点无任何认证。端口 39403 映射到宿主机。
> **复现 Payload**:
> ```bash
> curl http://localhost:39403/admin/llm/config
> # 返回: {"api_key":"sk-f22c79fa7bf846c3a41ff31a0398a733",...}
> curl http://localhost:39403/admin/dashboard
> # 返回: 519条政策数据统计
> ```
> **影响**: 政策数据篡改、爬取源操控、LLM 密钥窃取、任意政策注入。
> **修复建议**: 添加 Basic Auth 或集成平台认证体系。生产环境不暴露端口。

#### C-03: LLM API Key 明文泄露
> **位置**: `policy-crawler/internal/admin/admin_llm.go:30-38`
> **描述**: DeepSeek API Key 以明文存储在 llm_configs 表，GET 接口完整返回。
> **复现 Payload**:
> ```bash
> curl http://localhost:39403/admin/llm/config
> # {"api_key":"sk-f22c79fa7bf846c3a41ff31a0398a733"}
> ```
> **影响**: 已泄露密钥可被滥用，产生费用。需立即轮换。
> **修复建议**: AES-256-GCM 加密存储，API 返回时脱敏（仅显示末4位）。

#### C-04: 数据库默认密码 + SSL 禁用
> **位置**: `docker-compose.yml:11`, `shared/config/config.go:24`
> **描述**: 默认密码 `postgres123`，`sslmode=disable` 明文传输。
> **影响**: 数据库凭据已知，连接未加密，PII 数据暴露。
> **修复建议**: 移除默认密码，强制 sslmode=require，轮换密码。

#### C-05: CORS 通配符 + 无 CSRF = 全面跨站攻击
> **位置**: `shared/middleware/middleware.go:55`
> **描述**: `Access-Control-Allow-Origin: *` 配合 x-user-id 头认证，任意恶意网站可冒充用户。
> **复现 Payload**:
> ```bash
> curl -H "Origin: http://evil.com" -H "x-user-id: victim" http://localhost:39401/v1/profile
> # 200 OK — 请求被接受处理
> ```
> **修复建议**: 限定 CORS 为可信域名，添加 CSRF Token，SameSite=Strict。

### 3.2 P1 — 高危问题 (High) — 共 10 项

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| H-01 | actuarial-engine 零认证 | `actuarial-engine/cmd/main.go:17-33` | 任意调用计算引擎 |
| H-02 | IDOR — MarkAlertRead 无归属校验 | `rights_handler.go:85-106`, `rights_repo.go:125` | 用户可标记他人提醒 |
| H-03 | 内部错误详情泄露给客户端 | `handler.go:33`, `repository.go:48` | 泄露 DB schema/连接信息 |
| H-04 | 前端 3 端硬编码 localhost | `weapp/api.js:1`, `alipay/api.js:1` | 生产环境无法连接 |
| H-05 | 全部容器以 root 运行 | 所有 Dockerfile | 容器逃逸风险 |
| H-06 | 无速率限制 | 全局 | DoS/暴力攻击无防护 |
| H-07 | 无 CSP/X-Frame-Options 头 | middleware.go | XSS/点击劫持无防护 |
| H-08 | MinIO 默认凭据 minioadmin | `docker-compose.infra.yml:42-43` | 对象存储可被接管 |
| H-09 | 内部服务端口全暴露 | `docker-compose.yml` 全部 ports | 数据库/Redis/引擎直接可达 |
| H-10 | 精算引擎明文 HTTP 通信 | `plan_handler.go:77`, `cmd/main.go:37` | 用户财务数据明文传输 |

### 3.3 P2 — 中危问题 (Medium) — 共 12 项

| # | 问题 | 位置 |
|---|------|------|
| M-01 | /webclient 端点无认证暴露 | `api-server/cmd/main.go:64` |
| M-02 | ReportProxyHandler 用户 ID 可伪造 | `webclient_handler.go:16-31` |
| M-03 | Profile 输入字段缺验证 | `handler.go:73-123` |
| M-04 | 查询结果无 LIMIT 上限 | `handler.go:128-129`, `policy_repo.go:73-77` |
| M-05 | PostgreSQL/Redis 端口暴露宿主 | `docker-compose.infra.yml:12,25` |
| M-06 | 服务间明文 HTTP | `api-server/cmd/main.go:37` |
| M-07 | Chromium SSRF 风险 | `renderer.go:23-34` |
| M-08 | 无 CSRF 保护 | 全局 |
| M-09 | 无优雅关闭 | 全部 main.go 使用 log.Fatal |
| M-10 | 数据库连接池未配置 | `shared/db/db.go:10-26` |
| M-11 | HTTP Server 无超时配置 | 全部 ListenAndServe |
| M-12 | LLM 配置保存非事务(竞态) | `store.go:351-358` |

### 3.4 P3 — 低危问题 (Low) — 共 6 项

| # | 问题 | 位置 |
|---|------|------|
| L-01 | 反馈内容无最大长度限制 | `feedback_handler.go:33-36` |
| L-02 | 反馈联系方式无格式验证 | `feedback_handler.go:41` |
| L-03 | 日期解析逻辑脆弱 | `handler.go:85-97` |
| L-04 | 用户 ID 被记录到日志 | `rights_handler.go:177-178` |
| L-05 | .env.example 包含默认凭据 | `.env.example` |
| L-06 | EventBus 潜在 goroutine 泄漏 | `eventbus.go:75-111` |

---

## 4. 模拟真实用户试用报告 (User Trial Simulation)

### 4.1 核心业务流程测试

| 流程 | 步骤 | 结果 | 耗时 |
|------|------|------|------|
| 用户画像填写 | POST /v1/profile (age=30, gender=male, income=10000) | 200 OK | <50ms |
| 方案生成 | POST /v1/plans/generate | 200 OK, 3方案+30年现金流 | ~200ms |
| 方案报告查看 | GET /v1/plans/report?plan_id=xxx | 200 OK, HTML报告 | <100ms |
| 政策搜索 | GET /v1/policies/search?region_code=SH | 200 OK, 政策列表 | <50ms |
| 缴费查询 | GET /v1/rights/payment-status | 200 OK | <50ms |
| 提醒列表 | GET /v1/rights/alerts | 200 OK | <50ms |
| 反馈提交 | POST /v1/feedback | 200 OK | <50ms |

### 4.2 正面发现 (What Works Well)

- **SQL 注入防护**: 全部参数化查询，无注入风险
- **业务逻辑正确**: 方案计算从真实政策数据提取费率，不再硬编码
- **精算引擎解耦**: 纯数学计算器，零城市依赖
- **请求体大小限制**: 全部 POST/PUT 使用 MaxBytesReader (1MB)
- **Panic 恢复**: RecoveryMiddleware 防止崩溃泄露堆栈
- **数据库 CHECK 约束**: 迁移文件定义了 age/gender/employment_status 等约束
- **最小依赖**: 仅 lib/pq + go-redis，无已知 CVE

### 4.3 第一印象与操作体验

- **第一印象**: 业务流程完整，方案计算专业且基于真实政策数据
- **操作流畅度**: 90% 的 API 响应在 100ms 内
- **潜在隐患**: 认证机制严重不足，任何"用户"可冒充任意身份；管理面板完全裸奔

---

## 5. 安全风险评估表

| 风险等级 | 数量 | 关键风险 |
|----------|------|---------|
| **CRITICAL (P0)** | 5 | 身份伪造、管理面板裸奔、API Key 泄露、默认密码、CORS 通配 |
| **HIGH (P1)** | 10 | IDOR、引擎无认证、错误泄露、容器 root、无速率限制 |
| **MEDIUM (P2)** | 12 | CSRF、SSRF、无超时、连接池、明文通信 |
| **LOW (P3)** | 6 | 输入验证、日志、竞态 |
| **PASS** | 3 | SQL注入防护、方案生成、精算引擎 |

---

## 6. 修复优先级路线图

### 阶段一: 紧急修复 (上线前必须完成) — P0

| # | 修复项 | 工作量 | 验证方法 |
|---|--------|--------|---------|
| 1 | 实现 JWT 认证替换 x-user-id | 2天 | curl -H "Authorization: Bearer <invalid>" 返回 401 |
| 2 | /admin/* 添加认证中间件 | 0.5天 | curl 无认证返回 401 |
| 3 | API Key 加密存储+返回脱敏 | 1天 | GET 返回 "sk-****a733" |
| 4 | 移除默认密码, 强制环境变量 | 0.5天 | 无 DATABASE_URL 时启动失败 |
| 5 | CORS 限定可信域名 | 0.5天 | Origin: evil.com 返回 403 |

### 阶段二: 加固 (上线后1周内) — P1

| # | 修复项 | 工作量 |
|---|--------|--------|
| 6 | actuarial-engine 添加内部认证 | 0.5天 |
| 7 | 修复 IDOR (添加 user_id WHERE 条件) | 0.5天 |
| 8 | 错误响应脱敏 | 0.5天 |
| 9 | 前端 API 地址环境化 | 0.5天 |
| 10 | Dockerfile 添加非 root 用户 | 0.5天 |
| 11 | 添加速率限制中间件 | 1天 |
| 12 | 添加 CSP/X-Frame-Options 头 | 0.5天 |
| 13 | 修改默认 MinIO 凭据 | 0.5天 |
| 14 | 移除内部端口暴露 | 0.5天 |
| 15 | 服务间通信升级 HTTPS | 1天 |

### 阶段三: 优化 (上线后2周内) — P2

| # | 修复项 | 工作量 |
|---|--------|--------|
| 16 | 添加 CSRF Token | 1天 |
| 17 | 实现优雅关闭 | 0.5天 |
| 18 | 配置数据库连接池 | 0.5天 |
| 19 | 配置 HTTP Server 超时 | 0.5天 |
| 20 | 添加 Docker 资源限制 | 0.5天 |
| 21 | 输入验证增强 | 1天 |
| 22 | 查询结果 LIMIT 上限 | 0.5天 |

---

## 7. 最终验收结论

> **结论**: 经全平台极致深度审计，系统在**业务逻辑层面表现正确**（方案计算基于真实政策数据、精算引擎解耦完成、SQL 注入防护到位），但在**安全层面存在 5 项 P0 级致命漏洞**，**不可进入生产环境**。
>
> **核心阻塞点**: 身份认证机制形同虚设（C-01），管理面板完全裸奔且暴露 API Key（C-02/C-03），数据库使用默认密码且禁用 SSL（C-04），CORS 通配允许任意网站跨域攻击（C-05）。
>
> **建议**: 完成阶段一全部 5 项 P0 修复后，重新执行渗透测试验证，通过后方可进入灰度发布。预计阶段一修复工作量约 4.5 人天。

---

*报告由 OpenCode Omni-Platform QA Expert 自动生成 — 2026-05-25*
*审计工具: 静态代码分析 + 实时 API 渗透测试 + 基础设施审计 + 前端代码审查*
