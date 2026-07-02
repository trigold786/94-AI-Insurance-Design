# WebClient Core Path Enhancement Implementation Plan

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix backend cashflow/afterTaxPension passthrough, then enhance webclient with retirement visualization, scheme comparison cards, cashflow charts, and a pension simulator.

**Architecture:** Backend fix in api-server handler layer (no actuarial-engine changes). Frontend all in single `webClientHTML` string constant in `webclient_handler.go`, using Chart.js v4 CDN for charts.

**Tech Stack:** Go 1.x, Chart.js v4 (CDN), vanilla JS/CSS, existing test patterns (httptest + table-driven)

---

## File Structure

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

| File | Responsibility |
|------|---------------|
| `shared/models/models.go` | Add `AfterTaxPension` field to `Scheme` struct |
| `services/api-server/internal/handler/plan_handler.go` | Add fields to `SchemeResult`, map cashflow + afterTaxPension |
| `services/api-server/internal/handler/webclient_handler.go` | Rewrite Tab 1/2 HTML+JS, add Chart.js CDN, add simulator section |
| `services/api-server/internal/handler/plan_handler_test.go` | Add test for cashflow/afterTaxPension passthrough |

---

### Task 1: Add AfterTaxPension to models.Scheme

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `shared/models/models.go:118-132`

- [ ] **Step 1: Add field to Scheme struct**

In `shared/models/models.go`, add `AfterTaxPension` field after `ProjectedPension` (line 130):

```go
type Scheme struct {
	Name                  string         `json:"name"`
	BaseSalary            int            `json:"base_salary"`
	MonthlyCost           float64        `json:"monthly_cost"`
	AnnualSubsidy         float64        `json:"annual_subsidy"`
	SubsidyPolicy         string         `json:"subsidy_policy"`
	SubsidyCondition      string         `json:"subsidy_condition"`
	PaidMonths            int            `json:"paid_months"`
	TargetMonths          int            `json:"target_months"`
	RemainingMonths       int            `json:"remaining_months"`
	TotalPersonalCost     float64        `json:"total_personal_cost"`
	RemainingPersonalCost float64        `json:"remaining_personal_cost"`
	ProjectedPension      float64        `json:"projected_pension"`
	AfterTaxPension       float64        `json:"after_tax_pension"`
	Cashflow              []CashFlowItem `json:"cashflow,omitempty"`
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./shared/...`
Expected: compiles without errors

---

### Task 2: Fix SchemeResult and mapping in plan_handler.go

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/api-server/internal/handler/plan_handler.go:38-51` (SchemeResult)
- Modify: `services/api-server/internal/handler/plan_handler.go:207-224` (mapping loop)

- [ ] **Step 1: Add fields to SchemeResult struct**

Replace the `SchemeResult` struct (lines 38-51):

```go
type SchemeResult struct {
	Name                  string                `json:"name"`
	BaseSalary            int                   `json:"base_salary"`
	MonthlyCost           float64               `json:"monthly_cost"`
	AnnualSubsidy         float64               `json:"annual_subsidy"`
	SubsidyPolicy         string                `json:"subsidy_policy"`
	SubsidyCondition      string                `json:"subsidy_condition"`
	PaidMonths            int                   `json:"paid_months"`
	TargetMonths          int                   `json:"target_months"`
	RemainingMonths       int                   `json:"remaining_months"`
	TotalPersonalCost     float64               `json:"total_personal_cost"`
	RemainingPersonalCost float64               `json:"remaining_personal_cost"`
	ProjectedPension      float64               `json:"projected_pension"`
	AfterTaxPension       float64               `json:"after_tax_pension"`
	Cashflow              []models.CashFlowItem `json:"cashflow,omitempty"`
}
```

- [ ] **Step 2: Fix the mapping loop**

Replace lines 207-224 (the scheme conversion loop):

```go
		var totalCost, totalSubsidy float64
		var schemes []models.Scheme
		for _, s := range calcResp.Schemes {
			totalCost += s.MonthlyCost * 12
			totalSubsidy += s.AnnualSubsidy
			schemes = append(schemes, models.Scheme{
				Name:                  s.Name,
				BaseSalary:            s.BaseSalary,
				MonthlyCost:           s.MonthlyCost,
				AnnualSubsidy:         s.AnnualSubsidy,
				SubsidyPolicy:         s.SubsidyPolicy,
				SubsidyCondition:      s.SubsidyCondition,
				PaidMonths:            s.PaidMonths,
				TargetMonths:          s.TargetMonths,
				RemainingMonths:       s.RemainingMonths,
				TotalPersonalCost:     s.TotalPersonalCost,
				RemainingPersonalCost: s.RemainingPersonalCost,
				ProjectedPension:      s.ProjectedPension,
				AfterTaxPension:       s.AfterTaxPension,
				Cashflow:              s.Cashflow,
			})
		}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./services/api-server/...`
Expected: compiles without errors

---

### Task 3: Add test for cashflow/afterTaxPension passthrough

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/api-server/internal/handler/plan_handler_test.go`

- [ ] **Step 1: Write test**

Append to `plan_handler_test.go`:

```go
func TestGeneratePlanHandlerCashflowPassthrough(t *testing.T) {
	calc := &mockCalculator{
		resp: &CalculateResponse{
			Schemes: []SchemeResult{
				{
					Name:            "缂磋垂鍩烘暟 6000",
					BaseSalary:      6000,
					MonthlyCost:     600,
					ProjectedPension: 2500,
					AfterTaxPension: 2300,
					Cashflow: []models.CashFlowItem{
						{Year: 1, Payment: 7200, Subsidy: 1200, Balance: 8400},
						{Year: 2, Payment: 7560, Subsidy: 1260, Balance: 17220},
					},
				},
			},
		},
	}
	repo := &mockPlanRepo{}
	handler := GeneratePlanHandler(calc, repo, nil, nil)

	body := `{"age":30,"gender":"male","employment":"flexible","contribution_years":10,"current_balance":50000,"monthly_budget":3000,"local_avg_salary":10000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if repo.savedPlan == nil {
		t.Fatal("expected plan to be saved")
	}
	scheme := repo.savedPlan.RecommendedSchemes[0]
	if scheme.AfterTaxPension != 2300 {
		t.Errorf("expected afterTaxPension 2300, got %f", scheme.AfterTaxPension)
	}
	if len(scheme.Cashflow) != 2 {
		t.Fatalf("expected 2 cashflow items, got %d", len(scheme.Cashflow))
	}
	if scheme.Cashflow[0].Year != 1 || scheme.Cashflow[0].Payment != 7200 {
		t.Errorf("unexpected cashflow item: %+v", scheme.Cashflow[0])
	}
	respBody := w.Body.String()
	if !strings.Contains(respBody, "after_tax_pension") {
		t.Error("response should contain after_tax_pension field")
	}
	if !strings.Contains(respBody, "cashflow") {
		t.Error("response should contain cashflow field")
	}
}
```

- [ ] **Step 2: Run test**

Run: `go test ./services/api-server/internal/handler/ -run TestGeneratePlanHandlerCashflowPassthrough -v`
Expected: PASS

- [ ] **Step 3: Run all existing plan tests to verify no regression**

Run: `go test ./services/api-server/internal/handler/ -run TestGeneratePlan -v`
Expected: all PASS

---

### Task 4: Build and deploy backend fix

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:** None new 鈥?just build and deploy

- [ ] **Step 1: Cross-compile for Linux**

Run: `$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o api-server ./cmd/` (in `services/api-server/`)

- [ ] **Step 2: Deploy to Docker**

Run:
```bash
docker cp services/api-server/api-server nsi-api-server:/api-server
docker restart nsi-api-server
```

- [ ] **Step 3: Verify health**

Run: `curl -s http://localhost:39401/healthz`
Expected: `{"status":"ok"}`

---

### Task 5: Add Chart.js CDN and global styles to webclient

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/api-server/internal/handler/webclient_handler.go` 鈥?the `webClientHTML` const

- [ ] **Step 1: Add Chart.js script tag**

In the `<head>` section, after the closing `</style>` tag and before `</head>`, add:

```html
<script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>
```

- [ ] **Step 2: Add chart container CSS**

Inside the existing `<style>` block, add these styles (before `</style>`):

```css
.chart-container{position:relative;height:280px;width:100%;margin:8px 0}
.scheme-card{border:1px solid #E5E7EB;border-radius:10px;padding:16px;margin:10px 0;transition:box-shadow 0.2s}
.scheme-card:hover{box-shadow:0 4px 12px rgba(0,0,0,0.1)}
.scheme-card.recommended{border-color:#1A56DB;border-width:2px}
.scheme-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:8px}
.scheme-metrics{display:flex;gap:16px;flex-wrap:wrap}
.scheme-metric{text-align:center;min-width:80px}
.scheme-metric .value{font-size:20px;font-weight:700;color:#1F2937}
.scheme-metric .label{font-size:11px;color:#6B7280;margin-top:2px}
.roi-badge{display:inline-block;padding:3px 10px;border-radius:12px;font-size:12px;font-weight:600}
.roi-green{background:#D1FAE5;color:#059669}
.roi-yellow{background:#FEF3C7;color:#D97706}
.roi-red{background:#FEE2E2;color:#EF4444}
.rec-badge{background:#1A56DB;color:#fff;padding:3px 10px;border-radius:12px;font-size:11px;font-weight:600}
.timeline-bar{display:flex;height:36px;border-radius:8px;overflow:hidden;margin:12px 0;font-size:11px;color:#fff}
.timeline-segment{display:flex;align-items:center;justify-content:center;min-width:40px;padding:0 8px}
.timeline-labels{display:flex;justify-content:space-between;font-size:11px;color:#6B7280;margin-top:4px}
.sim-section{background:#F0F5FF;border-radius:10px;padding:16px;margin:12px 0}
.sim-slider-row{display:flex;align-items:center;gap:12px;margin:8px 0}
.sim-slider-row label{min-width:100px;font-size:13px;color:#374151}
.sim-slider-row input[type=range]{flex:1}
.sim-slider-row .sim-val{min-width:80px;text-align:right;font-weight:600;color:#1A56DB}
.sim-result{font-size:28px;font-weight:700;color:#1A56DB;text-align:center;padding:12px}
.sim-result-sub{font-size:13px;color:#6B7280;text-align:center}
.cashflow-toggle{cursor:pointer;color:#1A56DB;font-size:13px;margin-top:8px;display:inline-block}
.cashflow-panel{margin-top:8px;display:none}
.cashflow-panel.open{display:block}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./services/api-server/...`
Expected: compiles without errors

---

### Task 6: Implement delayed retirement visualization in Tab 1

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/api-server/internal/handler/webclient_handler.go` 鈥?the JS in `webClientHTML`

- [ ] **Step 1: Add retireTimeline() helper function**

Add after the `calcPensionAge` function (after the closing `}` around line 251):

```js
function retireTimeline(dob,gender,origAge){
  var ri=calcPensionAge(dob,gender,origAge);
  var baseAge=60;if(gender==='female'){baseAge=55;if(origAge>=50&&origAge<=60)baseAge=origAge;}
  var baseY=ri.years-baseAge+baseAge;
  var oldRetireYear=baseY;
  var newRetireYear=ri.years;
  var curYear=new Date().getFullYear();
  var curMonth=new Date().getMonth()+1;
  var p=dob.split('-');var by=parseInt(p[0])||1996;
  var ageNow=curYear-by;
  var totalSpan=baseAge-ageNow;if(totalSpan<1)totalSpan=1;
  var delayM=(newRetireYear-oldRetireYear)*12+(ri.months-((parseInt(dob.split('-')[1])||1)));
  var el=baseAge-ageNow;if(el<0)el=0;if(el>totalSpan)el=totalSpan;
  var baseS=baseAge-ageNow;if(baseS<0)baseS=0;if(baseS>totalSpan)baseS=totalSpan;
  var delayLen=Math.max(baseS>el?1:0,0.5);
  var elPct=Math.round(el/totalSpan*100);
  var basePct=Math.round((baseS-el)/totalSpan*100);
  var delayPct=100-elPct-basePct;if(delayPct<0)delayPct=0;
  var h='<div class="timeline-bar">';
  if(elPct>0)h+='<div class="timeline-segment" style="width:'+elPct+'%;background:#9CA3AF">宸茶繃 '+el+'骞?/div>';
  if(basePct>0)h+='<div class="timeline-segment" style="width:'+basePct+'%;background:#3B82F6">璺濇硶瀹氶��浼?'+(baseS-el)+'骞?/div>';
  if(delayPct>0)h+='<div class="timeline-segment" style="width:'+delayPct+'%;background:#F59E0B">寤惰繜 +'+(ri.years-oldRetireYear)+'骞?+ri.months+'鏈?/div>';
  h+='</div>';
  h+='<div class="timeline-labels"><span>'+by+'骞?(鍑虹敓)</span><span>'+oldRetireYear+'骞?(鍘熸硶瀹?</span><span>'+newRetireYear+'骞?+ri.months+'鏈?(瀹為檯閫�浼?</span></div>';
  if(newRetireYear>oldRetireYear){
    h+='<div style="margin-top:6px;font-size:12px;color:#D97706;background:#FEF3C7;padding:6px 10px;border-radius:6px">2025寤惰繜閫�浼戞柊鏀匡細寤惰繜 '+(ri.years-oldRetireYear)+'骞?+(ri.months<10?'0':'')+ri.months+'鏈?/div>';
  } else {
    h+='<div style="margin-top:6px;font-size:12px;color:#059669;background:#D1FAE5;padding:6px 10px;border-radius:6px">2025寤惰繜閫�浼戞柊鏀匡細鎮ㄧ殑寤惰繜骞呭害涓?锛岄��浼戞椂闂翠笉鍙楀奖鍝?/div>';
  }
  return h;
}
```

- [ ] **Step 2: Replace retireInfo display in showProfile()**

In the `showProfile()` function, find the line that builds the `retireInfo=calcPensionAge(...)` and the `retireYear>0?'...'` section. Replace the retire alert div with the timeline:

Find (around line 269):
```js
    (retireYear>0?'<div id="pfRetireAlert" class="alert-success" style="margin:8px 0;font-size:14px;text-align:center">棰勮鍒濇棰嗗彇閫�浼戦噾: <strong>'+retireYear+'骞?+retireMonth+'鏈?/strong></div>':'')+
```

Replace with:
```js
    '<div id="pfRetireTimeline" style="margin:12px 0">'+retireTimeline(d.dob,d.gender,d.origAge)+'</div>'+
```

- [ ] **Step 3: Fix recalcRetireAlert to update timeline**

Replace the `recalcRetireAlert` function:

```js
function recalcRetireAlert(){
  var oa=parseInt(document.getElementById('pf-orig-age')?.value)||0;
  var g=document.getElementById('pf-gender').value;
  var el=document.getElementById('pfRetireTimeline');
  if(el)el.innerHTML=retireTimeline(_formData.dob,g,oa);
}
```

- [ ] **Step 4: Remove the now-unused retireYear/retireMonth variables from showProfile()**

In `showProfile()`, remove these 3 lines near the top:
```js
  var retireInfo=calcPensionAge(d.dob,d.gender,d.origAge);
  var retireYear=retireInfo.years,retireMonth=retireInfo.months;
```

They are no longer needed since `retireTimeline()` handles the display directly.

- [ ] **Step 5: Verify compilation**

Run: `go build ./services/api-server/...`

---

### Task 7: Implement scheme comparison cards + ROI chart in Tab 2

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/api-server/internal/handler/webclient_handler.go` 鈥?the `showPlan()` and `onGeneratePlan()` functions

- [ ] **Step 1: Replace showPlan() function**

Replace the entire `showPlan` function (lines ~339-347):

```js
function showPlan(){
  var age=calcAge(_formData.dob);
  document.getElementById('app').innerHTML=
    '<h2>鏂规鐢熸垚</h2><p style="font-size:13px;color:#6B7280;margin-bottom:12px">鍩轰簬淇濆瓨鐨勭敤鎴风敾鍍忕敓鎴愮ぞ淇濇柟妗?/p>'+
    '<div style="font-size:13px;color:#374151;margin-bottom:12px;padding:12px;background:#F9FAFB;border-radius:8px">'+
    '鍑虹敓: '+_formData.dob+' | 鎬у埆: '+(_formData.gender==='male'?'鐢?:'濂?)+' | 灏变笟: '+_formData.employment+' | '+
    '缂磋垂: '+_formData.months+'鏈?| 棰勭畻: '+_formData.budget+'鍏?鏈?| 鍏昏�侀噾鎬婚: '+_formData.pensionTotal+'鍏?/div>'+
    '<button class="btn btn-primary" onclick="onGeneratePlan()" id="genBtn">鐢熸垚鏂规</button>'+
    '<div id="planResult"></div>';
}
```

- [ ] **Step 2: Replace onGeneratePlan() function with card + chart rendering**

Replace the entire `onGeneratePlan` function (lines ~349-377):

```js
function onGeneratePlan(){
  var btn=document.getElementById('genBtn');btn.disabled=true;btn.textContent='鐢熸垚涓?..';
  var age=calcAge(_formData.dob);
  var req={
    age:age,gender:_formData.gender,
    original_pension_age:_formData.origAge||0,
    employment:_formData.employment,
    contribution_years:Math.floor(_formData.months/12),
    current_balance:_formData.pensionTotal,
    local_avg_salary:0,
    monthly_budget:_formData.budget,
    priority:'balanced',
  };
  api('POST','/v1/plans/generate',req).then(function(plan){
    btn.disabled=false;btn.textContent='鍐嶆鐢熸垚';
    window._lastPlan=plan;
    window._lastPlanId=plan.plan_id;
    var schemes=plan.recommended_schemes||[];
    var maxROI=0,schemesWithROI=schemes.map(function(s){
      var roi=s.monthly_cost>0?s.projected_pension/s.monthly_cost:0;
      if(roi>maxROI)maxROI=roi;
      s._roi=roi;
      return s;
    });
    var h='<div class="alert-success mt-16" style="margin-top:12px">鏂规鐢熸垚鎴愬姛锛佸叡 '+schemes.length+' 涓柟妗?/div>';
    h+='<div style="margin:12px 0"><canvas id="schemeChart" class="chart-container"></canvas></div>';
    schemesWithROI.forEach(function(s,i){
      var isRec=s._roi===maxROI;
      var roiClass=s._roi>=3?'roi-green':s._roi>=2?'roi-yellow':'roi-red';
      h+='<div class="scheme-card'+(isRec?' recommended':'')+'">';
      h+='<div class="scheme-header"><strong style="font-size:15px">'+esc(s.name)+'</strong>';
      if(isRec)h+='<span class="rec-badge">鎺ㄨ崘</span>';
      h+='<span class="roi-badge '+roiClass+'">ROI '+s._roi.toFixed(1)+'x</span></div>';
      h+='<div class="scheme-metrics">';
      h+='<div class="scheme-metric"><div class="value">'+s.monthly_cost.toFixed(0)+'</div><div class="label">鏈堢即(鍏?</div></div>';
      h+='<div class="scheme-metric"><div class="value" style="color:#059669">'+s.projected_pension.toFixed(0)+'</div><div class="label">棰勮鏈堝吇鑰侀噾</div></div>';
      if(s.after_tax_pension>0)h+='<div class="scheme-metric"><div class="value" style="color:#7C3AED">'+s.after_tax_pension.toFixed(0)+'</div><div class="label">绋庡悗鏈堥</div></div>';
      if(s.annual_subsidy>0)h+='<div class="scheme-metric"><div class="value" style="color:#D97706">'+s.annual_subsidy.toFixed(0)+'</div><div class="label">骞磋ˉ璐?/div></div>';
      h+='<div class="scheme-metric"><div class="value" style="font-size:16px;color:#6B7280">'+s.remaining_months+'</div><div class="label">杩橀渶缂?鏈?</div></div>';
      h+='</div>';
      if(s.subsidy_policy)h+='<div style="margin-top:8px;font-size:12px;color:#6B7280">'+esc(s.subsidy_policy)+'</div>';
      if(s.cashflow&&s.cashflow.length>0){
        h+='<div class="cashflow-toggle" onclick="toggleCashflow('+i+')">鏌ョ湅鐜伴噾娴佽秼鍔?鈻?/div>';
        h+='<div class="cashflow-panel" id="cf-panel-'+i+'"><canvas id="cf-chart-'+i+'" class="chart-container"></canvas></div>';
      }
      h+='</div>';
    });
    h+='<button class="btn btn-outline mt-16" onclick="viewReport(\''+esc(plan.plan_id)+'\')">鏌ョ湅瀹屾暣鎶ュ憡</button>';
    document.getElementById('planResult').innerHTML=h;
    showResult(plan);
    renderSchemeChart(schemesWithROI);
    schemesWithROI.forEach(function(s,i){
      if(s.cashflow&&s.cashflow.length>0)renderCashflowChart(i,s.cashflow);
    });
  }).catch(function(e){btn.disabled=false;btn.textContent='鐢熸垚鏂规';document.getElementById('planResult').innerHTML='<div class="alert-error">'+esc(e.message)+'</div>'});
}
```

- [ ] **Step 3: Add chart rendering functions**

Add these functions before the `showPlan` function:

```js
var _schemeChart=null;
var _cfCharts={};
function renderSchemeChart(schemes){
  var ctx=document.getElementById('schemeChart');
  if(!ctx)return;
  if(_schemeChart)_schemeChart.destroy();
  _schemeChart=new Chart(ctx,{
    type:'bar',
    data:{
      labels:schemes.map(function(s){return s.name}),
      datasets:[
        {label:'鏈堢即璐?鍏?',data:schemes.map(function(s){return s.monthly_cost}),backgroundColor:'rgba(59,130,246,0.7)'},
        {label:'棰勮鏈堝吇鑰侀噾(鍏?',data:schemes.map(function(s){return s.projected_pension}),backgroundColor:'rgba(5,150,105,0.7)'},
        {label:'绋庡悗鏈堥(鍏?',data:schemes.map(function(s){return s.after_tax_pension||0}),backgroundColor:'rgba(124,58,237,0.7)'}
      ]
    },
    options:{responsive:true,maintainAspectRatio:false,plugins:{legend:{position:'bottom',labels:{font:{size:11}}}},scales:{y:{beginAtZero:true,ticks:{callback:function(v){return v+'鍏?}}}}}
  });
}
function toggleCashflow(i){
  var el=document.getElementById('cf-panel-'+i);
  if(el)el.classList.toggle('open');
}
function renderCashflowChart(i,cf){
  var ctx=document.getElementById('cf-chart-'+i);
  if(!ctx)return;
  if(_cfCharts[i])_cfCharts[i].destroy();
  _cfCharts[i]=new Chart(ctx,{
    type:'line',
    data:{
      labels:cf.map(function(c){return '绗?+c.year+'骞?}),
      datasets:[
        {label:'骞寸即璐?,data:cf.map(function(c){return c.payment}),borderColor:'#3B82F6',backgroundColor:'rgba(59,130,246,0.1)',fill:true,tension:0.3},
        {label:'骞磋ˉ璐?,data:cf.map(function(c){return c.subsidy}),borderColor:'#F59E0B',backgroundColor:'rgba(245,158,11,0.1)',fill:true,tension:0.3},
        {label:'绱浣欓',data:cf.map(function(c){return c.balance}),borderColor:'#059669',backgroundColor:'rgba(5,150,105,0.1)',fill:true,tension:0.3}
      ]
    },
    options:{responsive:true,maintainAspectRatio:false,plugins:{legend:{position:'bottom',labels:{font:{size:11}}}},scales:{y:{ticks:{callback:function(v){return (v/10000).toFixed(1)+'涓?}}}}}
  });
}
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./services/api-server/...`

---

### Task 8: Implement retirement pension simulator in Tab 2

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Files:**
- Modify: `services/api-server/internal/handler/webclient_handler.go` 鈥?add simulator section in plan result

- [ ] **Step 1: Add simulator HTML generation in onGeneratePlan()**

After the `h+='<</div>';` that closes the last scheme card, and before the "鏌ョ湅瀹屾暣鎶ュ憡" button, insert:

```js
    h+='<div class="sim-section"><h3 style="font-size:14px;color:#1A56DB;margin-bottom:8px">閫�浼戦噾妯℃嫙鍣?/h3>';
    h+='<div class="sim-slider-row"><label>缂磋垂鍩烘暟</label><input type="range" id="sim-base" min="'+schemes[0].base_salary+'" max="'+schemes[schemes.length-1].base_salary+'" step="100" value="'+schemes[0].base_salary+'" oninput="onSimChange()"><span class="sim-val" id="sim-base-val">'+schemes[0].base_salary+'鍏?/span></div>';
    h+='<div class="sim-slider-row"><label>缂磋垂骞撮檺</label><input type="range" id="sim-years" min="1" max="40" step="1" value="'+Math.floor(_formData.months/12)+'" oninput="onSimChange()"><span class="sim-val" id="sim-years-val">'+Math.floor(_formData.months/12)+'骞?/span></div>';
    h+='<div class="sim-slider-row"><label>鏈堥绠?/label><input type="range" id="sim-budget" min="500" max="10000" step="100" value="'+_formData.budget+'" oninput="onSimChange()"><span class="sim-val" id="sim-budget-val">'+_formData.budget+'鍏?/span></div>';
    h+='<div class="sim-result" id="sim-result">璁＄畻涓?..</div>';
    h+='<div class="sim-result-sub" id="sim-detail"></div></div>';
```

- [ ] **Step 2: Add simulator calculation function**

Add after the `renderCashflowChart` function:

```js
function onSimChange(){
  var base=parseFloat(document.getElementById('sim-base').value)||5000;
  var years=parseInt(document.getElementById('sim-years').value)||15;
  var budget=parseFloat(document.getElementById('sim-budget').value)||2000;
  document.getElementById('sim-base-val').textContent=base+'鍏?;
  document.getElementById('sim-years-val').textContent=years+'骞?;
  document.getElementById('sim-budget-val').textContent=budget+'鍏?;
  var cityCode=_formData.city||'310000';
  var avgSalary=12383;
  var pensionRate=0.08,medicalRate=0.02;
  var ci=GetCityInfoJS(cityCode);
  if(ci){avgSalary=ci.avgSalary;pensionRate=ci.pensionRate||0.08;medicalRate=ci.medicalRate||0.02;}
  var monthlyCost=base*(pensionRate+medicalRate);
  var personalBalance=base*pensionRate*years*12*Math.pow(1.03,years/2);
  var retireAge=_formData.gender==='male'?63:58;
  if(_formData.gender==='female'&&_formData.origAge===50)retireAge=55;
  var divisor=retireAge>=60?139:retireAge>=55?170:195;
  var personalPension=personalBalance/divisor;
  var basicPension=(avgSalary+base)/2*years*0.01;
  var total=Math.round(basicPension+personalPension);
  var el=document.getElementById('sim-result');
  if(el)el.textContent='锟?+total.toLocaleString()+' /鏈?;
  var detail=document.getElementById('sim-detail');
  if(detail)detail.textContent='鍩虹鍏昏�侀噾 '+Math.round(basicPension)+' + 涓汉璐︽埛 '+Math.round(personalPension)+' | 鏈堢即 '+Math.round(monthlyCost)+'鍏?| 鎶曞叆浜у嚭姣?'+(monthlyCost>0?(total/monthlyCost).toFixed(1):'-')+'x';
}
function GetCityInfoJS(code){
  var cities={'310000':{name:'涓婃捣',avgSalary:12383,pensionRate:0.08,medicalRate:0.02},'110000':{name:'鍖椾含',avgSalary:15764,pensionRate:0.08,medicalRate:0.02},'440300':{name:'娣卞湷',avgSalary:14530,pensionRate:0.08,medicalRate:0.02},'440100':{name:'骞垮窞',avgSalary:13795,pensionRate:0.08,medicalRate:0.02},'330100':{name:'鏉窞',avgSalary:9625,pensionRate:0.08,medicalRate:0.02}};
  return cities[code]||null;
}
```

- [ ] **Step 3: Add onSimChange() call at the end of onGeneratePlan()**

At the very end of the `.then(function(plan){...})` callback in `onGeneratePlan`, after rendering charts, add:

```js
    onSimChange();
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./services/api-server/...`

---

### Task 9: Build, deploy, and verify

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- [ ] **Step 1: Cross-compile**

Run: `$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o api-server ./cmd/` (in `services/api-server/`)

- [ ] **Step 2: Deploy**

Run:
```bash
docker cp services/api-server/api-server nsi-api-server:/api-server
docker restart nsi-api-server
```

- [ ] **Step 3: Verify health**

Run: `Start-Sleep -Seconds 3; curl.exe -s http://localhost:39401/healthz`
Expected: `{"status":"ok"}`

- [ ] **Step 4: Verify webclient loads**

Run: `curl.exe -s -o /dev/null -w "%{http_code}" http://localhost:39401/webclient`
Expected: 200

- [ ] **Step 5: Smoke test 鈥?generate a plan**

Run:
```bash
TOKEN=$(curl -s -X POST http://localhost:39401/v1/auth/token -H 'Content-Type:application/json' -d '{"user_id":"test-user"}' | jq -r '.data.token')
curl -s -X POST http://localhost:39401/v1/plans/generate -H "Authorization: Bearer $TOKEN" -H 'Content-Type:application/json' -d '{"age":30,"gender":"male","employment":"flexible","contribution_years":10,"current_balance":50000,"monthly_budget":3000,"local_avg_salary":12383}'
```

Expected: response contains `after_tax_pension`, `cashflow` fields with data in each scheme.

- [ ] **Step 6: Commit**

```bash
git add shared/models/models.go services/api-server/internal/handler/plan_handler.go services/api-server/internal/handler/plan_handler_test.go services/api-server/internal/handler/webclient_handler.go
git commit -m "feat(webclient): enhance core path with retirement viz, scheme cards, cashflow charts, pension simulator"
```

---

## Self-Review Checklist

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- [x] Spec coverage: Task 1-4 = backend fix, Task 5-8 = frontend enhancements (all 5 items from spec covered)
- [x] No placeholders: all code shown in full
- [x] Type consistency: `SchemeResult.AfterTaxPension` (float64) 鈫?`models.Scheme.AfterTaxPension` (float64), `SchemeResult.Cashflow` ([]models.CashFlowItem) 鈫?`models.Scheme.Cashflow` ([]CashFlowItem) 鈥?types match
- [x] Test coverage: Task 3 tests the backend fix end-to-end with mock calculator
