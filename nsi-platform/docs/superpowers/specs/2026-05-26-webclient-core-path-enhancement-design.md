# WebClient Core Path Enhancement Design

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Date**: 2026-05-26
**Scope**: User-facing webclient (`/webclient`), focused on "Profile 鈫?Plan 鈫?Report" core path
**Chart Library**: Chart.js v4 via CDN (~60KB, no build tools needed)

---

## 1. Backend Bug Fix: Cashflow & AfterTaxPension Passthrough

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Problem
`plan_handler.go` defines `SchemeResult` (line 38) without `AfterTaxPension` or `Cashflow` fields. When decoding the actuarial engine response (line 108-111), these fields are silently dropped. The mapping loop (lines 207-223) also omits them.

### Fix
1. Add `AfterTaxPension float64` and `Cashflow []models.CashFlowItem` to `SchemeResult` struct in `plan_handler.go`
2. Add `AfterTaxPension` field to `models.Scheme` struct in `shared/models/models.go`
3. Map both fields in the scheme conversion loop (lines 210-223)

### Files
- `services/api-server/internal/handler/plan_handler.go`
- `shared/models/models.go` (add `AfterTaxPension` field to `Scheme`)

### Impact
No DB migration needed 鈥?`recommended_schemes` is stored as JSON blob in `plan_snapshots`, so the new fields are automatically included.

---

## 2. Delayed Retirement Visualization (Profile Tab Enhancement)

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Current
Single line of text: "棰勮鍒濇棰嗗彇閫�浼戦噾: 2056骞?鏈?

### Enhancement
Visual timeline bar showing:
```
[Now 2026] ---- [娉曞畾 60宀?2040] ---- [寤惰繜鍚?60y8m 2040-08]
                  鈫?base              鈫?actual (+8 months)
```

- Color-coded segments: elapsed (gray), remaining to base (blue), delay months (orange)
- Label showing "2025鏂版斂绛栧欢杩?+8涓湀"
- Comparison: old policy age vs new policy age

### Implementation
Pure HTML/CSS in webclient JS. No new API calls.

---

## 3. Scheme Comparison Cards + ROI Chart (Plan Tab Enhancement)

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Current
Plain HTML table with columns: 鏂规/缂磋垂鍩烘暟/鏈堢即/骞磋ˉ璐?棰勮鏈堝吇鑰侀噾

### Enhancement

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

#### 3a. Card Layout
Each scheme rendered as a card with:
- Header: scheme name (e.g. "鏂规A - 60%绀惧钩")
- Key metrics in large font: 鏈堢即 鈫?棰勮鏈堝吇鑰侀噾
- ROI indicator: `projected_pension / monthly_cost` ratio with color (green > 3x, yellow 2-3x, red < 2x)
- Badge: "鎺ㄨ崘" for highest ROI scheme
- Subsidy info: "骞磋ˉ璐?楼X,XXX"
- After-tax pension: "绋庡悗鏈堥 楼X,XXX" (from fix #1)

#### 3b. Comparison Bar Chart
Chart.js horizontal grouped bar chart:
- X axis: schemes
- Y axis: monthly cost (blue) vs projected pension (green)
- Side panel: total cost comparison

### Implementation
Chart.js CDN + vanilla JS. Uses data already returned by `POST /v1/plans/generate`.

---

## 4. Cashflow Trend Chart (Plan Detail Enhancement)

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Current
Cashflow data is dropped (bug #1).

### Enhancement (after fix #1 provides data)
For each scheme card, add an expandable cashflow section:
- Chart.js line chart with 3 lines: Annual Payment (red), Annual Subsidy (green), Cumulative Balance (blue)
- X axis: years (1 to retirement)
- Hover tooltip: exact amounts
- Summary stats below chart: total paid, total subsidy received, final balance at retirement

### Implementation
Chart.js line chart. Data from `scheme.cashflow[]` array (available after fix #1).

---

## 5. Retirement Pension Simulator (New Tab or Section in Plan Tab)

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Current
Fixed parameters, one-shot generation.

### Enhancement
Interactive parameter sliders + real-time result update:
- Slider 1: 缂磋垂鍩烘暟 (contribution base, range from min to max based on city)
- Slider 2: 缂磋垂骞撮檺 target (180-360 months)
- Slider 3: 鏈堥绠?(monthly budget)

When user adjusts sliders:
- Frontend recalculates estimated pension using simplified formula: `base_pension + personal_account_pension`
- Displays projected monthly pension as a large number that updates in real-time
- Shows a reference line for "褰撳湴鏈�浣庣敓娲讳繚闅? (city minimum living standard)

### Implementation
No new API needed. Frontend JS uses:
- `local_avg_salary` and rates from the existing plan generation response
- Chinese pension formula: basic pension = (avg_salary + indexed_salary) / 2 脳 years 脳 1%; personal account = balance / divisor_months
- Sliders trigger JS recalculation, no backend call

### Formula (simplified frontend calc)
```
basic_pension = (local_avg_salary + contribution_base) / 2 * contribution_years * 0.01
personal_pension = contribution_base * 0.08 * contribution_years * 12 * 1.03^years / divisor_months
total = basic_pension + personal_pension
```

---

## Technical Notes

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### Chart.js Integration
```html
<script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>
```
Single CDN script tag in the `<head>` section of `webClientHTML`. No module bundler needed.

### Modified Files Summary
| File | Change |
|------|--------|
| `shared/models/models.go` | Add `AfterTaxPension` to `Scheme` |
| `services/api-server/internal/handler/plan_handler.go` | Add fields to `SchemeResult`, map cashflow + afterTaxPension |
| `services/api-server/internal/handler/webclient_handler.go` | Rewrite Tab 1 (retirement viz), Tab 2 (cards + charts + simulator), Chart.js CDN |

### No Changes To
- Actuarial engine (already returns complete data)
- Policy-crawler service
- Admin endpoints
- Other webclient tabs (compliance, guide, policies, rights, feedback)
- Database schema

---

## Out of Scope (Future Enhancements)
- Downloadable PDF report (`/v1/plans/report/pdf` exposure)
- Semantic policy search (crawler `/v1/policies/similar`)
- Payment record submission (`POST /v1/rights/payment-records`)
- Policy version history
