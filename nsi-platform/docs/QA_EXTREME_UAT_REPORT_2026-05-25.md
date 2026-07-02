# 极致深度测试�?UAT 验收报告 (Extreme QA Report)

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

## 1. 测试执行概览
- **项目名称**: NeuroSocialInsurance (神经社保智能平台)
- **验收目标**: 100% 符合 PRD 验收标准
- **测试深度**: 覆盖业务逻辑、底层鲁棒性、安全渗透、UX 交互、极端环境及并发性能
- **审计范围**: api-server (Go), actuarial-engine (Go), policy-crawler (Go), 前端 (微信小程�?支付宝小程序/WebClient), Docker 基础设施
- **执行日期**: 2026-05-25
- **执行状�?*: **发现风险 �?不可交付，需修复 P0 级问题后方可上线**

---

## 2. 深度测试矩阵 (Deep Testing Matrix)

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| 需�?ID | 需求描�?| 验收标准 (AC) | 安全审计 | 鲁棒性检�?| 性能表现 | 验收状�?|
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| SEC-01 | 用户认证 | JWT/Session 鉴权 | **FAIL**: 使用 x-user-id 头伪造身份，零加密验�?| 任何用户可冒充任意身�?| N/A | **P0 阻塞** |
| SEC-02 | 管理面板安全 | 管理接口需鉴权 | **FAIL**: /admin/* 全部无认�?| 519条政策数据、LLM API Key 完全暴露 | N/A | **P0 阻塞** |
| SEC-03 | API Key 保护 | 密钥加密存储/传输 | **FAIL**: DeepSeek API Key 明文存储+明文返回 | `sk-f22c79fa7bf846c3a41ff31a0398a733` 已泄�?| N/A | **P0 阻塞** |
| SEC-04 | 数据库安�?| 非默认密�?SSL | **FAIL**: 默认密码 postgres123, sslmode=disable | 连接明文传输 | N/A | **P0 阻塞** |
| SEC-05 | CORS 策略 | 限定可信�?| **FAIL**: Access-Control-Allow-Origin: * | 任意网站可跨域请�?伪造用�?| N/A | **P0 阻塞** |
| SEC-06 | IDOR 防护 | 资源归属校验 | **FAIL**: MarkAlertRead �?ownership 校验 | 用户 A 可标记用�?B 的提�?| N/A | **HIGH** |
| SEC-07 | 速率限制 | 防暴�?防DoS | **FAIL**: 零速率限制 | 无限调用方案计算引擎 | N/A | **HIGH** |
| SEC-08 | CSRF 防护 | SameSite/Token | **FAIL**: �?CSRF 保护 | POST/PUT/DELETE 跨站可操�?| N/A | **MEDIUM** |
| SQL-01 | SQL 注入防护 | 参数化查�?| **PASS**: 全部使用 $1/$2 占位�?| 边界值未触发注入 | N/A | **PASS** |
| SQL-02 | 输入校验 | 字段格式验证 | **PARTIAL**: Age 有范围校�? 其他字段缺校�?| monthly_income=-1 未拦�?| N/A | **MEDIUM** |
| BIZ-01 | 方案生成 | 政策数据驱动计算 | **PASS**: �?policy_claims 提取费率 | 极端�?age=999)已拦�?| 200ms 内响�?| **PASS** |
| BIZ-02 | 精算引擎 | 纯数学无硬编�?| **PASS**: city.go 已移�?| 多方�?现金流计算正�?| <100ms | **PASS** |
| BIZ-03 | 政策爬取 | 自动爬取+LLM提取 | **PASS**: 14�?19�?| Chromium 渲染正常 | N/A | **PASS** |
| UX-01 | 前端连�?| API调用正确 | **FAIL**: 3端硬编码 127.0.0.1:39401 | 生产环境无法连接 | N/A | **HIGH** |
| UX-02 | 安全响应�?| CSP/X-Frame/X-CTO | **FAIL**: 全部缺失 | XSS/点击劫持风险 | N/A | **HIGH** |
| INF-01 | 容器安全 | �?root 运行 | **FAIL**: 全部 root 运行 | 容器逃逸风�?| N/A | **HIGH** |
| INF-02 | 优雅关闭 | SIGTERM 安全处理 | **FAIL**: 使用 log.Fatal+ListenAndServe | 在途请求被丢弃 | N/A | **MEDIUM** |
| INF-03 | 连接�?| 数据库连接管�?| **FAIL**: 未配�?MaxOpen/MaxIdle/MaxLifetime | 高并发下连接耗尽 | N/A | **MEDIUM** |
| INF-04 | 请求超时 | Read/Write/Idle | **FAIL**: 未配�?Server 超时 | Slowloris 攻击风险 | N/A | **MEDIUM** |
| INF-05 | 资源限制 | Docker mem/cpu | **FAIL**: 无限�?| OOM 风险 | N/A | **MEDIUM** |

---

## 3. 问题发现与风险预�?(Issue Tracker)

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### 3.1 P0 �?严重漏洞 (Critical/Blocker) �?�?5 �?
#### C-01: 身份认证形同虚设 �?x-user-id 头伪�?> **位置**: `shared/middleware/middleware.go:24-35`
> **描述**: 认证中间件仅检�?`x-user-id` 头是否存在，�?JWT/Session/签名验证。config.go 中定义了 JWTSecret 但从未使用�?> **复现 Payload**:
> ```bash
> curl -H "x-user-id: admin" http://localhost:39401/v1/profile
> # 返回 200 + 完整用户资料
> ```
> **影响**: 全量用户身份冒充，任意用户资�?方案/缴费/提醒可读写�?> **修复建议**: 实现 JWT Bearer Token 验证，从 token claims 提取 userID�?
#### C-02: 管理面板零认�?> **位置**: `policy-crawler/cmd/main.go:90-111`
> **描述**: 10+ �?/admin/* 端点无任何认证。端�?39403 映射到宿主机�?> **复现 Payload**:
> ```bash
> curl http://localhost:39403/admin/llm/config
> # 返回: {"api_key":"sk-f22c79fa7bf846c3a41ff31a0398a733",...}
> curl http://localhost:39403/admin/dashboard
> # 返回: 519条政策数据统�?> ```
> **影响**: 政策数据篡改、爬取源操控、LLM 密钥窃取、任意政策注入�?> **修复建议**: 添加 Basic Auth 或集成平台认证体系。生产环境不暴露端口�?
#### C-03: LLM API Key 明文泄露
> **位置**: `policy-crawler/internal/admin/admin_llm.go:30-38`
> **描述**: DeepSeek API Key 以明文存储在 llm_configs 表，GET 接口完整返回�?> **复现 Payload**:
> ```bash
> curl http://localhost:39403/admin/llm/config
> # {"api_key":"sk-f22c79fa7bf846c3a41ff31a0398a733"}
> ```
> **影响**: 已泄露密钥可被滥用，产生费用。需立即轮换�?> **修复建议**: AES-256-GCM 加密存储，API 返回时脱敏（仅显示末4位）�?
#### C-04: 数据库默认密�?+ SSL 禁用
> **位置**: `docker-compose.yml:11`, `shared/config/config.go:24`
> **描述**: 默认密码 `postgres123`，`sslmode=disable` 明文传输�?> **影响**: 数据库凭据已知，连接未加密，PII 数据暴露�?> **修复建议**: 移除默认密码，强�?sslmode=require，轮换密码�?
#### C-05: CORS 通配�?+ �?CSRF = 全面跨站攻击
> **位置**: `shared/middleware/middleware.go:55`
> **描述**: `Access-Control-Allow-Origin: *` 配合 x-user-id 头认证，任意恶意网站可冒充用户�?> **复现 Payload**:
> ```bash
> curl -H "Origin: http://evil.com" -H "x-user-id: victim" http://localhost:39401/v1/profile
> # 200 OK �?请求被接受处�?> ```
> **修复建议**: 限定 CORS 为可信域名，添加 CSRF Token，SameSite=Strict�?
### 3.2 P1 �?高危问题 (High) �?�?10 �?
| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| H-01 | actuarial-engine 零认�?| `actuarial-engine/cmd/main.go:17-33` | 任意调用计算引擎 |
| H-02 | IDOR �?MarkAlertRead 无归属校�?| `rights_handler.go:85-106`, `rights_repo.go:125` | 用户可标记他人提�?|
| H-03 | 内部错误详情泄露给客户端 | `handler.go:33`, `repository.go:48` | 泄露 DB schema/连接信息 |
| H-04 | 前端 3 端硬编码 localhost | `weapp/api.js:1`, `alipay/api.js:1` | 生产环境无法连接 |
| H-05 | 全部容器�?root 运行 | 所�?Dockerfile | 容器逃逸风�?|
| H-06 | 无速率限制 | 全局 | DoS/暴力攻击无防�?|
| H-07 | �?CSP/X-Frame-Options �?| middleware.go | XSS/点击劫持无防�?|
| H-08 | MinIO 默认凭据 minioadmin | `docker-compose.infra.yml:42-43` | 对象存储可被接管 |
| H-09 | 内部服务端口全暴�?| `docker-compose.yml` 全部 ports | 数据�?Redis/引擎直接可达 |
| H-10 | 精算引擎明文 HTTP 通信 | `plan_handler.go:77`, `cmd/main.go:37` | 用户财务数据明文传输 |

### 3.3 P2 �?中危问题 (Medium) �?�?12 �?
| # | 问题 | 位置 |
|---|------|------|
| M-01 | /webclient 端点无认证暴�?| `api-server/cmd/main.go:64` |
| M-02 | ReportProxyHandler 用户 ID 可伪�?| `webclient_handler.go:16-31` |
| M-03 | Profile 输入字段缺验�?| `handler.go:73-123` |
| M-04 | 查询结果�?LIMIT 上限 | `handler.go:128-129`, `policy_repo.go:73-77` |
| M-05 | PostgreSQL/Redis 端口暴露宿主 | `docker-compose.infra.yml:12,25` |
| M-06 | 服务间明�?HTTP | `api-server/cmd/main.go:37` |
| M-07 | Chromium SSRF 风险 | `renderer.go:23-34` |
| M-08 | �?CSRF 保护 | 全局 |
| M-09 | 无优雅关�?| 全部 main.go 使用 log.Fatal |
| M-10 | 数据库连接池未配�?| `shared/db/db.go:10-26` |
| M-11 | HTTP Server 无超时配�?| 全部 ListenAndServe |
| M-12 | LLM 配置保存非事�?竞�? | `store.go:351-358` |

### 3.4 P3 �?低危问题 (Low) �?�?6 �?
| # | 问题 | 位置 |
|---|------|------|
| L-01 | 反馈内容无最大长度限�?| `feedback_handler.go:33-36` |
| L-02 | 反馈联系方式无格式验�?| `feedback_handler.go:41` |
| L-03 | 日期解析逻辑脆弱 | `handler.go:85-97` |
| L-04 | 用户 ID 被记录到日志 | `rights_handler.go:177-178` |
| L-05 | .env.example 包含默认凭据 | `.env.example` |
| L-06 | EventBus 潜在 goroutine 泄漏 | `eventbus.go:75-111` |

---

## 4. 模拟真实用户试用报告 (User Trial Simulation)

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### 4.1 核心业务流程测试

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

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

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- **SQL 注入防护**: 全部参数化查询，无注入风�?- **业务逻辑正确**: 方案计算从真实政策数据提取费率，不再硬编�?- **精算引擎解�?*: 纯数学计算器，零城市依赖
- **请求体大小限�?*: 全部 POST/PUT 使用 MaxBytesReader (1MB)
- **Panic 恢复**: RecoveryMiddleware 防止崩溃泄露堆栈
- **数据�?CHECK 约束**: 迁移文件定义�?age/gender/employment_status 等约�?- **最小依�?*: �?lib/pq + go-redis，无已知 CVE

### 4.3 第一印象与操作体�?
- **第一印象**: 业务流程完整，方案计算专业且基于真实政策数据
- **操作流畅�?*: 90% �?API 响应�?100ms �?- **潜在隐患**: 认证机制严重不足，任�?用户"可冒充任意身份；管理面板完全裸奔

---

## 5. 安全风险评估�?
| 风险等级 | 数量 | 关键风险 |
|----------|------|---------|
| **CRITICAL (P0)** | 5 | 身份伪造、管理面板裸奔、API Key 泄露、默认密码、CORS 通配 |
| **HIGH (P1)** | 10 | IDOR、引擎无认证、错误泄露、容�?root、无速率限制 |
| **MEDIUM (P2)** | 12 | CSRF、SSRF、无超时、连接池、明文通信 |
| **LOW (P3)** | 6 | 输入验证、日志、竞�?|
| **PASS** | 3 | SQL注入防护、方案生成、精算引�?|

---

## 6. 修复优先级路线图

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### 阶段一: 紧急修�?(上线前必须完�? �?P0

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| # | 修复�?| 工作�?| 验证方法 |
|---|--------|--------|---------|
| 1 | 实现 JWT 认证替换 x-user-id | 2�?| curl -H "Authorization: Bearer <invalid>" 返回 401 |
| 2 | /admin/* 添加认证中间�?| 0.5�?| curl 无认证返�?401 |
| 3 | API Key 加密存储+返回脱敏 | 1�?| GET 返回 "sk-****a733" |
| 4 | 移除默认密码, 强制环境变量 | 0.5�?| �?DATABASE_URL 时启动失�?|
| 5 | CORS 限定可信域名 | 0.5�?| Origin: evil.com 返回 403 |

### 阶段�? 加固 (上线�?周内) �?P1

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| # | 修复�?| 工作�?|
|---|--------|--------|
| 6 | actuarial-engine 添加内部认证 | 0.5�?|
| 7 | 修复 IDOR (添加 user_id WHERE 条件) | 0.5�?|
| 8 | 错误响应脱敏 | 0.5�?|
| 9 | 前端 API 地址环境�?| 0.5�?|
| 10 | Dockerfile 添加�?root 用户 | 0.5�?|
| 11 | 添加速率限制中间�?| 1�?|
| 12 | 添加 CSP/X-Frame-Options �?| 0.5�?|
| 13 | 修改默认 MinIO 凭据 | 0.5�?|
| 14 | 移除内部端口暴露 | 0.5�?|
| 15 | 服务间通信升级 HTTPS | 1�?|

### 阶段�? 优化 (上线�?周内) �?P2

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| # | 修复�?| 工作�?|
|---|--------|--------|
| 16 | 添加 CSRF Token | 1�?|
| 17 | 实现优雅关闭 | 0.5�?|
| 18 | 配置数据库连接池 | 0.5�?|
| 19 | 配置 HTTP Server 超时 | 0.5�?|
| 20 | 添加 Docker 资源限制 | 0.5�?|
| 21 | 输入验证增强 | 1�?|
| 22 | 查询结果 LIMIT 上限 | 0.5�?|

---

## 7. 最终验收结�?
> **结论**: 经全平台极致深度审计，系统在**业务逻辑层面表现正确**（方案计算基于真实政策数据、精算引擎解耦完成、SQL 注入防护到位），但在**安全层面存在 5 �?P0 级致命漏�?*�?*不可进入生产环境**�?>
> **核心阻塞�?*: 身份认证机制形同虚设（C-01），管理面板完全裸奔且暴�?API Key（C-02/C-03），数据库使用默认密码且禁用 SSL（C-04），CORS 通配允许任意网站跨域攻击（C-05）�?>
> **建议**: 完成阶段一全部 5 �?P0 修复后，重新执行渗透测试验证，通过后方可进入灰度发布。预计阶段一修复工作量约 4.5 人天�?
---

*报告�?OpenCode Omni-Platform QA Expert 自动生成 �?2026-05-25*
*审计工具: 静态代码分�?+ 实时 API 渗透测�?+ 基础设施审计 + 前端代码审查*
