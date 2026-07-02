# 个性化 PDF 报告 设计文档

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

## 概述

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

�?AI社保智筹 系统增加方案报告生成能力。用户查看方案后可通过 `GET /v1/plans/{id}/report` 获取一份完整的 HTML 报告页面，包含方案对比、现金流预测、合规清单、权益分析和行动步骤�?
## 架构

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```
用户/移动�?�?GET /v1/plans/{id}/report
                     �?planRepo.GetPlan(planID) �?方案快照
                     �?compliance_handler �?合规数据
                     �?html/template 渲染 �?HTML
                     �?返回 Content-Type: text/html
```

**核心策略**：HTML 内联 CSS + `@media print` 规则。移动端 WebView 显示报告，用户可通过平台原生打印功能保存�?PDF。无需额外依赖库�?
## API

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**`GET /v1/plans/{id}/report`**

响应：`Content-Type: text/html; charset=utf-8`

HTML 页面结构�?1. 报告 Header（标题、生成时间、用户城市）
2. 方案概览（三档对比表�?3. 现金流预测（表格：年�?× 缴费/领取金额�?4. 合规清单（材料列�?+ 状态）
5. 权益说明（政策影响分析）
6. 行动步骤（办理流程）

## 组件

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### 1. `internal/handler/report_handler.go`

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```go
func PlanReportHandler(repo PlanRepository) http.Handler
```

- 解析 `{id}` �?URL path
- 调用 `repo.GetPlan(planID)` 获取方案数据
- 调用 `repo.GetCompliance(planID)` 获取合规数据（可选）
- �?`html/template` 渲染报告
- 返回 HTML

### 2. 数据获取

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

`PlanRepository` 接口需要的方法�?- `GetPlan(planID string) (*models.PlanSnapshot, error)` �?已有
- `GetCompliance(userID, cityCode string) (*models.ComplianceChecklist, error)` �?已有

### 3. 模板

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

使用 Go `html/template` 包。模板放�?`internal/handler/report_template.go`（Go 字符串常量，内联）�?
## 报告内容

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### 方案概览�?| 指标 | 档次 1 (60%) | 档次 2 (100%) | 档次 3 (300%) |
|------|-------------|--------------|--------------|
| 月缴费基�?| xx �?| xx �?| xx �?|
| 个人月缴 | xx �?| xx �?| xx �?|
| 政府月补 | xx �?| xx �?| xx �?|
| 年缴费总计 | xx �?| xx �?| xx �?|
| 预计月领 | xx �?| xx �?| xx �?|
| 回本周期 | xx �?| xx �?| xx �?|

### 现金流表
| 年份 | 年龄 | 年缴�?| 累计缴费 | 年领�?| 累计领取 | 账户余额 |
|------|------|--------|---------|--------|---------|---------|

### 合规清单
| 材料名称 | 来源 | 说明 | 状�?|
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

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- `TestPlanReport_Success` �?正常生成报告
- `TestPlanReport_NotFound` �?方案不存�?- `TestPlanReport_InvalidID` �?�?ID
- 验证 HTML 包含关键内容标签（方案对比表、合规清单）
