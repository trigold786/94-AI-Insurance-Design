# WebClient Core Path Enhancement Design

**Date**: 2026-05-26
**Scope**: User-facing webclient (`/webclient`), focused on "Profile → Plan → Report" core path
**Chart Library**: Chart.js v4 via CDN (~60KB, no build tools needed)

---

## 1. Backend Bug Fix: Cashflow & AfterTaxPension Passthrough

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
No DB migration needed — `recommended_schemes` is stored as JSON blob in `plan_snapshots`, so the new fields are automatically included.

---

## 2. Delayed Retirement Visualization (Profile Tab Enhancement)

### Current
Single line of text: "预计初次领取退休金: 2056年3月"

### Enhancement
Visual timeline bar showing:
```
[Now 2026] ---- [法定 60岁 2040] ---- [延迟后 60y8m 2040-08]
                  ↑ base              ↑ actual (+8 months)
```

- Color-coded segments: elapsed (gray), remaining to base (blue), delay months (orange)
- Label showing "2025新政策延迟 +8个月"
- Comparison: old policy age vs new policy age

### Implementation
Pure HTML/CSS in webclient JS. No new API calls.

---

## 3. Scheme Comparison Cards + ROI Chart (Plan Tab Enhancement)

### Current
Plain HTML table with columns: 方案/缴费基数/月缴/年补贴/预计月养老金

### Enhancement

#### 3a. Card Layout
Each scheme rendered as a card with:
- Header: scheme name (e.g. "方案A - 60%社平")
- Key metrics in large font: 月缴 → 预计月养老金
- ROI indicator: `projected_pension / monthly_cost` ratio with color (green > 3x, yellow 2-3x, red < 2x)
- Badge: "推荐" for highest ROI scheme
- Subsidy info: "年补贴 ¥X,XXX"
- After-tax pension: "税后月领 ¥X,XXX" (from fix #1)

#### 3b. Comparison Bar Chart
Chart.js horizontal grouped bar chart:
- X axis: schemes
- Y axis: monthly cost (blue) vs projected pension (green)
- Side panel: total cost comparison

### Implementation
Chart.js CDN + vanilla JS. Uses data already returned by `POST /v1/plans/generate`.

---

## 4. Cashflow Trend Chart (Plan Detail Enhancement)

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

### Current
Fixed parameters, one-shot generation.

### Enhancement
Interactive parameter sliders + real-time result update:
- Slider 1: 缴费基数 (contribution base, range from min to max based on city)
- Slider 2: 缴费年限 target (180-360 months)
- Slider 3: 月预算 (monthly budget)

When user adjusts sliders:
- Frontend recalculates estimated pension using simplified formula: `base_pension + personal_account_pension`
- Displays projected monthly pension as a large number that updates in real-time
- Shows a reference line for "当地最低生活保障" (city minimum living standard)

### Implementation
No new API needed. Frontend JS uses:
- `local_avg_salary` and rates from the existing plan generation response
- Chinese pension formula: basic pension = (avg_salary + indexed_salary) / 2 × years × 1%; personal account = balance / divisor_months
- Sliders trigger JS recalculation, no backend call

### Formula (simplified frontend calc)
```
basic_pension = (local_avg_salary + contribution_base) / 2 * contribution_years * 0.01
personal_pension = contribution_base * 0.08 * contribution_years * 12 * 1.03^years / divisor_months
total = basic_pension + personal_pension
```

---

## Technical Notes

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
