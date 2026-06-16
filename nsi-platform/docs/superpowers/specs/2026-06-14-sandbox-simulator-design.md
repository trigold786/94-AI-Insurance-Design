# 社保沙盘模拟器 + AI顾问 设计规格

| 属性 | 内容 |
|------|------|
| 版本 | V1.0.0 |
| 日期 | 2026-06-14 |
| 依据 | PRD V1.2.1 §4.4/§4.7 + MRD用户痛点分析 |
| 优先级 | 沙盘模拟器 P0；AI顾问 P1 |

---

## 1. 背景与目标

目标用户（前程序员/分析师/设计师转灵活就业）最大的痛点是"缺乏专业筹划"——有具体的what-if问题无人能答。现有产品是单向"表单→报告"流程，不支持实时探索。

**本升级将产品从"工具"升级为"顾问"**：用户拖动滑块即可实时看到不同决策对养老金、补贴、资格的影响，并可追问AI顾问。

## 2. 核心原则

- **所有阈值动态化**：缴费年限、年龄线、补贴金额等全部从`policy_claims.conditions`读取，不硬编码
- **渐进式政策支持**：最低缴费年限2029年15年→2039年20年（每年+6个月），根据用户预计退休年份动态计算
- **实时响应<500ms**：滑块拖动150ms防抖后请求，后端缓存5分钟
- **AI回复简洁**：≤3句话，先结论后依据，必须引用政策来源

## 3. 沙盘模拟器

### 3.1 输入参数

| 参数 | 类型 | 范围 | 影响 |
|------|------|------|------|
| city_code | 下拉 | 5个MVP城市 | 切换→重新匹配政策集 |
| gender | 开关 | male/female | 影响退休年龄、4050判定 |
| age | 滑块 | 16-70 | 跨越40(女)/50(男)→触发4050 |
| base_percent | 滑块 | 60-300, 步长10 | 缴费基数=当地平均工资×percent% |
| paid_years | 滑块 | 0-35 | 已缴年限，影响退休金、购房资格 |
| plan_years | 滑块 | 0-35 | 计划继续缴费年限 |
| employment | 下拉 | flexible/unemployed/employed | 影响补贴资格 |
| is_local_hukou | 开关 | true/false | 影响政策适用范围 |

### 3.2 输出结构

```
SimulatorResponse {
  cost: { monthly_total, monthly_pension, monthly_medical, annual_total }
  pension: { projected_monthly, personal_account_total, base_pension, account_pension }
  subsidy: { annual_total, items: [{ name, amount, policy_id, claim_id }] }
  net_monthly: float
  thresholds: { min_contribution_years, retirement_year, meets_min_years, years_shortfall }
  qualifications: [{ name, qualified, years_until?, detail }]
  cashflow: [{ year, payment, subsidy, net }]
  comparison: { at_60, at_100, at_300 }
  break_even_age: int
  policy_triggers: [{ type, message, severity }]
}
```

### 3.3 后端架构

**新增端点**：`POST /v1/simulator/calculate`（api-server）

**内部数据流**：
1. `ThresholdResolver` — 从policy_claims.conditions JSONB解析阈值（年限/年龄/金额），支持渐进式政策
2. `actuarial-engine /v1/calculate` — 计算养老金、缴费、现金流（已有算法，新增base_percent参数）
3. `ComplianceEvaluator` — 用当前参数匹配eligible政策（复用已有逻辑）
4. 合并返回

**新增模块**：`ThresholdResolver`
- 输入：city_code, gender, age, paid_years, plan_years, retirement_year
- 从DB读取相关政策的conditions JSONB
- 解析出适用阈值：min_contribution_years（渐进式）、4050年龄线、购房年限要求等
- 不硬编码任何数字

**缓存**：Redis，key=`sim:{city}:{gender}:{age}:{base}:{paid}:{plan}:{emp}:{hukou}`，TTL=5min

### 3.4 前端布局

WebClient新增Tab"社保沙盘"：
- 左面板：8个输入控件（滑块/下拉/开关）
- 右面板：4个数字卡片 + 政策触发提示 + 3个Chart.js图表 + 资格状态列表
- 底部：方案保存（最多3个）+ 方案对比弹窗
- 滑块150ms防抖，拖动时实时更新

## 4. AI社保顾问

### 4.1 交互

沙盘页底部对话框，用户输入问题→POST /v1/advisor/ask→返回简洁文字。

### 4.2 回复规则

- ≤3句话，先结论后依据
- 必须引用政策来源（"根据XX号文"）
- 不确定时说"建议咨询12333确认"
- 自动注入当前沙盘参数+用户画像作为上下文

### 4.3 后端

**端点**：`POST /v1/advisor/ask`

**数据流**：
1. 将问题+当前沙盘参数+用户画像组装为context
2. pgvector语义检索Top-5相关政策
3. LLM Gateway生成回复（system prompt约束≤3句话+引用来源）
4. 返回纯文本

## 5. 数据模型变更

### 5.1 新增表：simulator_scenarios（方案保存）

```sql
CREATE TABLE IF NOT EXISTS simulator_scenarios (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '方案A',
    params JSONB NOT NULL,
    result JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 5.2 渐进式政策数据

需要在policy_claims中添加或更新政策conditions，记录最低缴费年限的渐进式变化。示例conditions JSONB：
```json
[{"name":"最低缴费年限","type":"gradual_min_years","base_year":2030,"base_value":15,"increment_year":0.5,"max_value":20,"max_year":2039}]
```

ThresholdResolver识别`type=gradual_min_years`并按退休年份插值计算。

## 6. 实施计划

1. **后端-ThresholdResolver** — 新模块，从DB解析动态阈值
2. **后端-SimulatorHandler** — 新端点，调用actuary+policy+threshold合并返回
3. **后端-AdvisorHandler** — AI顾问端点，RAG+LLM
4. **前端-沙盘UI** — WebClient新Tab，滑块+图表+卡片
5. **前端-AI对话框** — 沙盘页底部对话
6. **迁移SQL** — simulator_scenarios表 + 渐进式政策conditions
7. **构建+部署+测试**
