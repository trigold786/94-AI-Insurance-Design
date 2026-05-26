# 个性化 PDF 报告 设计文档

## 概述

为 AI社保智筹 系统增加方案报告生成能力。用户查看方案后可通过 `GET /v1/plans/{id}/report` 获取一份完整的 HTML 报告页面，包含方案对比、现金流预测、合规清单、权益分析和行动步骤。

## 架构

```
用户/移动端 → GET /v1/plans/{id}/report
                     → planRepo.GetPlan(planID) → 方案快照
                     → compliance_handler → 合规数据
                     → html/template 渲染 → HTML
                     → 返回 Content-Type: text/html
```

**核心策略**：HTML 内联 CSS + `@media print` 规则。移动端 WebView 显示报告，用户可通过平台原生打印功能保存为 PDF。无需额外依赖库。

## API

**`GET /v1/plans/{id}/report`**

响应：`Content-Type: text/html; charset=utf-8`

HTML 页面结构：
1. 报告 Header（标题、生成时间、用户城市）
2. 方案概览（三档对比表）
3. 现金流预测（表格：年份 × 缴费/领取金额）
4. 合规清单（材料列表 + 状态）
5. 权益说明（政策影响分析）
6. 行动步骤（办理流程）

## 组件

### 1. `internal/handler/report_handler.go`

```go
func PlanReportHandler(repo PlanRepository) http.Handler
```

- 解析 `{id}` 从 URL path
- 调用 `repo.GetPlan(planID)` 获取方案数据
- 调用 `repo.GetCompliance(planID)` 获取合规数据（可选）
- 用 `html/template` 渲染报告
- 返回 HTML

### 2. 数据获取

`PlanRepository` 接口需要的方法：
- `GetPlan(planID string) (*models.PlanSnapshot, error)` — 已有
- `GetCompliance(userID, cityCode string) (*models.ComplianceChecklist, error)` — 已有

### 3. 模板

使用 Go `html/template` 包。模板放在 `internal/handler/report_template.go`（Go 字符串常量，内联）。

## 报告内容

### 方案概览表
| 指标 | 档次 1 (60%) | 档次 2 (100%) | 档次 3 (300%) |
|------|-------------|--------------|--------------|
| 月缴费基数 | xx 元 | xx 元 | xx 元 |
| 个人月缴 | xx 元 | xx 元 | xx 元 |
| 政府月补 | xx 元 | xx 元 | xx 元 |
| 年缴费总计 | xx 元 | xx 元 | xx 元 |
| 预计月领 | xx 元 | xx 元 | xx 元 |
| 回本周期 | xx 月 | xx 月 | xx 月 |

### 现金流表
| 年份 | 年龄 | 年缴费 | 累计缴费 | 年领取 | 累计领取 | 账户余额 |
|------|------|--------|---------|--------|---------|---------|

### 合规清单
| 材料名称 | 来源 | 说明 | 状态 |
|---------|------|------|------|

### 权益分析
- 养老保险：xx
- 医疗保险：xx
- 其他：xx

### 行动步骤
1. xxx
2. xxx
3. xxx

## 测试

- `TestPlanReport_Success` — 正常生成报告
- `TestPlanReport_NotFound` — 方案不存在
- `TestPlanReport_InvalidID` — 空 ID
- 验证 HTML 包含关键内容标签（方案对比表、合规清单）
