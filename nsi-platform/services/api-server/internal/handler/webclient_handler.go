package handler

import (
	"html/template"
	"net/http"
)

func WebClientHandler(apiBaseURL string) http.Handler {
	tmpl := template.Must(template.New("webclient").Parse(webClientHTML))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, map[string]string{"APIBaseURL": apiBaseURL})
	})
}

func ReportProxyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		planID := r.URL.Query().Get("plan_id")
		userID := r.URL.Query().Get("x-user-id")
		if planID == "" {
			http.Error(w, "plan_id required", http.StatusBadRequest)
			return
		}
		if userID == "" {
			userID = "default-user"
		}
		r2, _ := http.NewRequest("GET", "/v1/plans/report?plan_id="+planID, nil)
		r2.Header.Set("x-user-id", userID)
		http.Redirect(w, r, "/v1/plans/report?plan_id="+planID, http.StatusFound)
	})
}

const webClientHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>AI社保智筹 - 测试客户端</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:"Microsoft YaHei","PingFang SC",sans-serif;background:#F5F7FA;color:#1F2937;font-size:14px}
.header{background:linear-gradient(135deg,#1A56DB,#3B82F6);color:#fff;padding:14px 24px;display:flex;justify-content:space-between;align-items:center}
.header h1{font-size:18px;font-weight:600}
.nav{display:flex;gap:0;background:#fff;border-bottom:2px solid #E5E7EB;padding:0 8px;overflow-x:auto}
.nav-item{padding:10px 16px;cursor:pointer;font-size:13px;color:#6B7280;border-bottom:2px solid transparent;margin-bottom:-2px;white-space:nowrap}
.nav-item.active{color:#1A56DB;border-bottom-color:#1A56DB;font-weight:600}
.nav-item:hover{color:#1A56DB}
.container{max-width:1000px;margin:16px auto;padding:0 16px}
.card{background:#fff;border-radius:10px;padding:20px;margin-bottom:12px;box-shadow:0 1px 4px rgba(0,0,0,0.06)}
h2{font-size:16px;color:#1A56DB;margin-bottom:12px;border-bottom:2px solid #E5E7EB;padding-bottom:6px}
label{display:block;font-size:13px;color:#6B7280;margin-bottom:4px;margin-top:10px}
input,select,textarea{width:100%;padding:8px 12px;border:1px solid #D1D5DB;border-radius:6px;font-size:14px}
input:focus,select:focus,textarea:focus{outline:none;border-color:#1A56DB;box-shadow:0 0 0 3px rgba(26,86,219,0.1)}
textarea{min-height:80px;font-family:monospace}
.btn{padding:8px 20px;border:none;border-radius:6px;font-size:13px;cursor:pointer;font-weight:500;margin:4px}
.btn-primary{background:#1A56DB;color:#fff}
.btn-success{background:#059669;color:#fff}
.btn-outline{background:#fff;border:1px solid #D1D5DB;color:#374151}
.btn:disabled{opacity:0.5;cursor:not-allowed}
table{width:100%;border-collapse:collapse;font-size:13px;margin:8px 0}
th{text-align:left;padding:8px 10px;border-bottom:2px solid #E5E7EB;color:#6B7280;font-size:12px}
td{padding:7px 10px;border-bottom:1px solid #F3F4F6}
tr:hover{background:#F9FAFB}
.badge{display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:500}
.bg-green{background:#D1FAE5;color:#059669}
.bg-yellow{background:#FEF3C7;color:#D97706}
.bg-red{background:#FEE2E2;color:#EF4444}
.bg-blue{background:#DBEAFE;color:#1A56DB}
.form-row{display:flex;gap:12px;flex-wrap:wrap}
.form-row>*{flex:1;min-width:140px}
.result-box{background:#F9FAFB;border-radius:8px;padding:12px;font-family:monospace;font-size:12px;max-height:300px;overflow:auto;white-space:pre-wrap;margin-top:8px}
.alert-error{background:#FEE2E2;color:#EF4444;padding:8px 12px;border-radius:6px;margin:8px 0;font-size:13px}
.alert-success{background:#D1FAE5;color:#059669;padding:8px 12px;border-radius:6px;margin:8px 0;font-size:13px}
.spinner{display:inline-block;width:14px;height:14px;border:2px solid #E5E7EB;border-top-color:#1A56DB;border-radius:50%;animation:spin .6s linear infinite;vertical-align:middle;margin-right:6px}
@keyframes spin{to{transform:rotate(360deg)}}
.ym-picker{position:relative;display:inline-block;width:100%}
.ym-display{width:100%;padding:8px 12px;border:1px solid #D1D5DB;border-radius:6px;font-size:14px;cursor:pointer;background:#fff;text-align:left;position:relative}
.ym-display::after{content:'';position:absolute;right:10px;top:50%;border:5px solid transparent;border-top:6px solid #6B7280;transform:translateY(-3px)}
.ym-panel{display:none;position:absolute;top:100%;left:0;right:0;z-index:999;background:#fff;border:1px solid #D1D5DB;border-radius:8px;box-shadow:0 8px 24px rgba(0,0,0,0.15);margin-top:4px;padding:16px}
.ym-panel.open{display:block}
.ym-panel-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:12px;padding-bottom:8px;border-bottom:1px solid #E5E7EB}
.ym-panel-header span{font-size:14px;font-weight:600;color:#1F2937}
.ym-btn{padding:4px 12px;border:1px solid #D1D5DB;border-radius:4px;background:#fff;cursor:pointer;font-size:12px}
.ym-btn-primary{background:#1A56DB;color:#fff;border-color:#1A56DB;padding:6px 20px;font-weight:500}
.ym-col-row{display:flex;gap:8px}
.ym-col{flex:1;text-align:center}
.ym-col label{font-size:12px;color:#6B7280;margin-bottom:6px;display:block}
.ym-scroll{height:180px;overflow-y:auto;border:1px solid #E5E7EB;border-radius:6px;scroll-snap-type:y mandatory;-webkit-overflow-scrolling:touch}
.ym-scroll::-webkit-scrollbar{width:4px}
.ym-scroll::-webkit-scrollbar-thumb{background:#D1D5DB;border-radius:2px}
.ym-opt{padding:8px 4px;cursor:pointer;font-size:14px;scroll-snap-align:center;transition:background 0.15s}
.ym-opt:hover{background:#EFF6FF}
.ym-opt.sel{background:#1A56DB;color:#fff;font-weight:600;border-radius:4px;margin:0 4px}
.ym-months{display:grid;grid-template-columns:repeat(4,1fr);gap:4px;height:auto;max-height:180px;overflow-y:auto;border:1px solid #E5E7EB;border-radius:6px;padding:4px}
.ym-mopt{padding:8px 4px;cursor:pointer;font-size:13px;text-align:center;border-radius:4px;transition:background 0.15s}
.ym-mopt:hover{background:#EFF6FF}
.ym-mopt.sel{background:#1A56DB;color:#fff;font-weight:600}
.chart-container{position:relative;height:280px;width:100%;margin:8px 0}
.scheme-card{border:1px solid #E5E7EB;border-radius:10px;padding:16px;margin:10px 0;transition:box-shadow 0.2s}
.scheme-card:hover{box-shadow:0 4px 12px rgba(0,0,0,0.1)}
.scheme-card.recommended{border-color:#1A56DB;border-width:2px}
.scheme-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:8px;flex-wrap:wrap;gap:6px}
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
</style>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js"></script>
</head>
<body>

<div class="header"><h1>AI社保智筹 - 测试客户端</h1></div>
<div class="nav" id="navBar"></div>
<div class="container"><div id="app" class="card" style="text-align:center;padding:40px;color:#9CA3AF">选择一个功能开始测试</div></div>
<div id="resultArea" style="display:none" class="container"><div class="card"><h2>API 响应</h2><div id="resultBox" class="result-box"></div></div></div>

<script>
var API = '{{.APIBaseURL}}';
var UID = 'default-user';
var _token = '';
var _formData = {dob:'1996-01',gender:'male',origAge:0,employment:'flexible',months:60,budget:2000,hasChildren:false,pensionTotal:30000,pensionPersonal:15000,city:'310000'};

function ensureToken(){
  if(_token) return Promise.resolve(_token);
  return fetch(API+'/v1/auth/token',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({user_id:UID})})
    .then(function(r){return r.json()})
    .then(function(d){
      if(d.code===0&&d.data&&d.data.token){_token=d.data.token;return _token;}
      throw new Error(d.message||'token fetch failed');
    });
}

var navItems=[
  {id:'profile',label:'1.用户画像'},
  {id:'plan',label:'2.方案生成'},
  {id:'report',label:'3.报告'},
  {id:'compliance',label:'4.合规'},
  {id:'guide',label:'5.指南'},
  {id:'policies',label:'6.政策'},
  {id:'rights',label:'7.权益'},
  {id:'feedback',label:'8.反馈'},
];
function renderNav(){
  var h='';navItems.forEach(function(n){h+='<div class="nav-item" data-id="'+n.id+'">'+n.label+'</div>'});
  document.getElementById('navBar').innerHTML=h;
  document.getElementById('navBar').addEventListener('click',function(e){
    var item=e.target.closest('.nav-item');if(item)switchTab(item.dataset.id);
  });
}
renderNav();

function switchTab(id){
  document.querySelectorAll('.nav-item').forEach(function(n){n.classList.remove('active')});
  var el=document.querySelector('[data-id="'+id+'"]');if(el)el.classList.add('active');
  document.getElementById('resultArea').style.display='none';
  if(id==='profile')showProfile();
  else if(id==='plan')showPlan();
  else if(id==='report')showReport();
  else if(id==='compliance')showCompliance();
  else if(id==='guide')showGuide();
  else if(id==='policies')showPolicies();
  else if(id==='rights')showRights();
  else if(id==='feedback')showFeedback();
}

function api(method,path,data){
  return ensureToken().then(function(tk){
    var opts={method:method,headers:{'Authorization':'Bearer '+tk,'Content-Type':'application/json'}};
    if(data)opts.body=JSON.stringify(data);
    return fetch(API+path,opts).then(function(r){return r.json()}).then(function(d){
      if(d.code!==0&&d.code!==undefined)throw new Error(d.msg||d.message||'API error');
      return d.data;
    });
  });
}
function showResult(data){
  document.getElementById('resultArea').style.display='block';
  document.getElementById('resultBox').textContent=typeof data==='string'?data:JSON.stringify(data,null,2);
}
function esc(s){if(typeof s!=='string')return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML}

var _ymPickerVal='';
function initYMPicker(initialVal,onchange){
  _ymPickerVal=initialVal||'1996-01';
  var parts=_ymPickerVal.split('-');
  var selY=parseInt(parts[0])||1996,selM=parseInt(parts[1])||1;
  var curY=new Date().getFullYear();
  var minY=1940,maxY=curY-16;
  function render(){
    var disp=document.getElementById('ym-display');
    if(disp)disp.textContent=_ymPickerVal;
  }
  window._ymOpen=function(){
    var panel=document.getElementById('ym-panel');
    if(!panel)return;
    panel.classList.add('open');
    var scrollEl=document.getElementById('ym-year-scroll');
    if(scrollEl){var el=scrollEl.querySelector('.sel');if(el)el.scrollIntoView({block:'center'});}
  };
  window._ymClose=function(){
    var panel=document.getElementById('ym-panel');
    if(panel)panel.classList.remove('open');
  };
  window._ymToggle=function(e){
    e.stopPropagation();
    var panel=document.getElementById('ym-panel');
    if(panel)panel.classList.toggle('open');
  };
  window._ymSelYear=function(y){
    selY=y;
    if(selM>12){selM=12;}
    _ymPickerVal=y+'-'+(selM<10?'0':'')+selM;
    var disp=document.getElementById('ym-display');if(disp)disp.textContent=_ymPickerVal;
    buildYears();buildMonths();
  };
  window._ymSelMonth=function(m){
    selM=m;
    _ymPickerVal=selY+'-'+(m<10?'0':'')+m;
    var disp=document.getElementById('ym-display');if(disp)disp.textContent=_ymPickerVal;
    buildMonths();
  };
  window._ymPrevYear=function(){if(selY>minY){selY--;_ymPickerVal=selY+'-'+(selM<10?'0':'')+selM;var disp=document.getElementById('ym-display');if(disp)disp.textContent=_ymPickerVal;buildYears();buildMonths();}};
  window._ymNextYear=function(){if(selY<maxY){selY++;_ymPickerVal=selY+'-'+(selM<10?'0':'')+selM;var disp=document.getElementById('ym-display');if(disp)disp.textContent=_ymPickerVal;buildYears();buildMonths();}};
  window._ymConfirm=function(){
    var panel=document.getElementById('ym-panel');if(panel)panel.classList.remove('open');
    if(onchange)onchange(_ymPickerVal);
  };  function buildYears(){
    var el=document.getElementById('ym-year-scroll');
    if(!el)return;
    var h='';
    for(var y=minY;y<=maxY;y++){
      h+='<div class="ym-opt'+(y===selY?' sel':'')+'" onclick="_ymSelYear('+y+')">'+y+'年</div>';
    }
    el.innerHTML=h;
  }
  function buildMonths(){
    var el=document.getElementById('ym-month-grid');
    if(!el)return;
    var h='';
    for(var m=1;m<=12;m++){
      h+='<div class="ym-mopt'+(m===selM?' sel':'')+'" onclick="_ymSelMonth('+m+')">'+m+'月</div>';
    }
    el.innerHTML=h;
  }
  buildYears();buildMonths();render();
  document.addEventListener('click',function(e){
    var picker=document.querySelector('.ym-picker');
    if(picker&&!picker.contains(e.target)){var panel=document.getElementById('ym-panel');if(panel)panel.classList.remove('open');}
  });
  var panelEl=document.getElementById('ym-panel');
  if(panelEl){panelEl.addEventListener('click',function(e){e.stopPropagation();});}
}

function calcPensionAge(dob,gender,origAge){
  if(!dob||dob.length<7){
    var d=new Date();return{years:d.getFullYear(),months:d.getMonth()+1};
  }
  var base=60;
  if(gender==='female')base=55;
  if(gender==='female'&&origAge>=50&&origAge<=60)base=origAge;
  var pace=(base===50?2:4),maxDelay=(base===50?60:36);
  var p=dob.split('-');
  var by=parseInt(p[0]),bm=parseInt(p[1]);
  var baseYear=2025-base;
  var months=(by-baseYear)*12+bm-1;
  if(months<0){var ry=by+base,rm=bm;if(rm>12){rm-=12;ry++;}return{years:ry,months:rm};}
  var delay=Math.floor(months/pace)+1;
  if(delay>maxDelay)delay=maxDelay;
  var pay=base+Math.floor(delay/12),pam=delay%12;
  var ry=by+pay+Math.floor((bm-1+pam)/12);
  var rm=((bm-1+pam)%12)+1;
  return{years:ry,months:rm};
}
function retireTimeline(dob,gender,origAge){
  var ri=calcPensionAge(dob,gender,origAge);
  var baseAge=60;if(gender==='female'){baseAge=55;if(origAge>=50&&origAge<=60)baseAge=origAge;}
  var p=dob.split('-');var by=parseInt(p[0])||1996;var bm=parseInt(p[1])||1;
  var curYear=new Date().getFullYear();
  var ageNow=curYear-by;
  var totalSpan=baseAge-ageNow;if(totalSpan<1)totalSpan=1;
  var oldRetireYear=by+baseAge;
  var newRetireYear=ri.years;
  var el=baseAge-ageNow;if(el<0)el=0;if(el>totalSpan)el=totalSpan;
  var baseS=baseAge-ageNow;if(baseS<0)baseS=0;if(baseS>totalSpan)baseS=totalSpan;
  var elPct=Math.round(el/totalSpan*100);
  var basePct=Math.round((baseS-el)/totalSpan*100);
  var delayPct=100-elPct-basePct;if(delayPct<0)delayPct=0;
  var h='<div class="timeline-bar">';
  if(elPct>0)h+='<div class="timeline-segment" style="width:'+elPct+'%;background:#9CA3AF">已过 '+el+'年</div>';
  if(basePct>0)h+='<div class="timeline-segment" style="width:'+basePct+'%;background:#3B82F6">距法定退休 '+(baseS-el)+'年</div>';
  if(delayPct>0)h+='<div class="timeline-segment" style="width:'+delayPct+'%;background:#F59E0B">延迟 +'+(newRetireYear-oldRetireYear)+'年</div>';
  h+='</div>';
  h+='<div class="timeline-labels"><span>'+by+'年(出生)</span><span>'+oldRetireYear+'年(原法定)</span><span>'+newRetireYear+'年'+ri.months+'月(实际)</span></div>';
  if(newRetireYear>oldRetireYear){
    h+='<div style="margin-top:6px;font-size:12px;color:#D97706;background:#FEF3C7;padding:6px 10px;border-radius:6px">2025延迟退休新政：延迟 '+(newRetireYear-oldRetireYear)+'年'+(ri.months<10?'0':'')+ri.months+'月，预计退休 <strong>'+newRetireYear+'年'+ri.months+'月</strong></div>';
  }else{
    h+='<div style="margin-top:6px;font-size:12px;color:#059669;background:#D1FAE5;padding:6px 10px;border-radius:6px">2025延迟退休新政：延迟幅度为0，退休时间不受影响</div>';
  }
  return h;
}
// === 1. Profile ===
function showProfile(){
  var d=_formData;
  document.getElementById('app').innerHTML=
    '<h2>用户画像</h2><div class="form-row">'+
    '<div><label>出生年月</label><div class="ym-picker">'+
    '<div id="ym-display" class="ym-display" onclick="_ymToggle(event)">'+d.dob+'</div>'+
    '<div id="ym-panel" class="ym-panel">'+
    '<div class="ym-panel-header"><button class="ym-btn" onclick="_ymPrevYear(event)">&lt;</button><span id="ym-header-title">选择年月</span><button class="ym-btn" onclick="_ymNextYear(event)">&gt;</button></div>'+
    '<div class="ym-col-row"><div class="ym-col"><label>年份</label><div id="ym-year-scroll" class="ym-scroll"></div></div><div class="ym-col"><label>月份</label><div id="ym-month-grid" class="ym-months"></div></div></div>'+
    '<div style="text-align:right;margin-top:10px"><button class="ym-btn ym-btn-primary" onclick="_ymConfirm()" style="padding:6px 24px;font-size:13px">确认</button></div>'+
    '</div></div></div>'+
    '<div><label>性别</label><select id="pf-gender" onchange="onGenderChange()"><option value="male" '+(d.gender==='male'?'selected':'')+'>男(60岁)</option><option value="female" '+(d.gender==='female'?'selected':'')+'>女</option></select>'+
    '<div id="pf-orig-age-wrap"'+(d.gender!=='female'?' style="display:none"':'')+'><label>原退休年龄</label><select id="pf-orig-age" onchange="onOrigAgeChange()"><option value="55" '+(d.origAge==55?'selected':'')+'>55岁(管理岗)</option><option value="50" '+(d.origAge==50?'selected':'')+'>50岁(工人岗)</option></select></div></div>'+
    '<div><label>户籍所在地</label><select id="pf-city"><option value="310000" '+(d.city==='310000'?'selected':'')+'>上海</option><option value="110000" '+(d.city==='110000'?'selected':'')+'>北京</option><option value="440100" '+(d.city==='440100'?'selected':'')+'>广州</option><option value="330100" '+(d.city==='330100'?'selected':'')+'>杭州</option><option value="440300" '+(d.city==='440300'?'selected':'')+'>深圳</option></select></div></div>'+
    '<div id="pfRetireTimeline" style="margin:12px 0">'+retireTimeline(d.dob,d.gender,d.origAge)+'</div>'+
    '<div class="form-row"><div><label>就业状态</label><select id="pf-employment"><option value="flexible" '+(d.employment==='flexible'?'selected':'')+'>灵活就业</option><option value="employed" '+(d.employment==='employed'?'selected':'')+'>在职</option><option value="unemployed" '+(d.employment==='unemployed'?'selected':'')+'>失业</option></select></div>'+
    '<div><label>累计缴费月数</label><input id="pf-months" type="number" value="'+d.months+'" min="0"></div>'+
    '<div><label>月预算(元)</label><input id="pf-budget" type="number" value="'+d.budget+'" min="0"></div></div>'+
    '<div class="form-row"><div><label>子女</label><select id="pf-children"><option value="false" '+(d.hasChildren===false?'selected':'')+'>无</option><option value="true" '+(d.hasChildren===true?'selected':'')+'>有</option></select></div>'+
    '<div><label>养老金本息总额(元)</label><input id="pf-pension-total" type="number" value="'+d.pensionTotal+'"></div>'+
    '<div><label>养老金总额个人部分(元)</label><input id="pf-pension-personal" type="number" value="'+d.pensionPersonal+'"></div></div>'+
    '<p style="font-size:12px;color:#6B7280;margin-top:8px">当地月均工资: 将由系统根据城市自动确定</p>'+
    '<button class="btn btn-primary mt-16" onclick="onSaveProfile()">保存画像</button><div id="pfResult"></div>';
  initYMPicker(d.dob,function(v){_formData.dob=v;recalcRetireAlert();});
}
function onGenderChange(){
  var g=document.getElementById('pf-gender').value;
  _formData.gender=g;
  var wrap=document.getElementById('pf-orig-age-wrap');
  if(wrap)wrap.style.display=g==='female'?'':'none';
  if(g!=='female')document.getElementById('pf-orig-age').value='55';
  recalcRetireAlert();
}
function onOrigAgeChange(){
  recalcRetireAlert();
}
function recalcRetireAlert(){
  var oa=parseInt(document.getElementById('pf-orig-age')?.value)||0;
  var g=document.getElementById('pf-gender').value;
  var el=document.getElementById('pfRetireTimeline');
  if(el)el.innerHTML=retireTimeline(_formData.dob,g,oa);
}
function onSaveProfile(){
  _formData={
    dob:_ymPickerVal||_formData.dob||'1996-01',
    gender:document.getElementById('pf-gender').value,
    origAge:parseInt(document.getElementById('pf-orig-age')?.value)||0,
    city:document.getElementById('pf-city').value,
    employment:document.getElementById('pf-employment').value,
    months:parseInt(document.getElementById('pf-months').value)||0,
    budget:parseFloat(document.getElementById('pf-budget').value)||0,
    hasChildren:document.getElementById('pf-children').value==='true',
    pensionTotal:parseFloat(document.getElementById('pf-pension-total').value)||0,
    pensionPersonal:parseFloat(document.getElementById('pf-pension-personal').value)||0,
  };
  api('PUT','/v1/profile',{
    date_of_birth:_formData.dob,
    gender:_formData.gender,
    household_region_code:_formData.city,
    employment_status:_formData.employment,
    contribution_months:_formData.months,
    has_children:_formData.hasChildren,
    pension_total_amount:_formData.pensionTotal,
    pension_personal_amount:_formData.pensionPersonal,
    social_security_years:Math.floor(_formData.months/12),
    age:calcAge(_formData.dob),
  }).then(function(d){
    document.getElementById('pfResult').innerHTML='<div class="alert-success">画像已保存</div>';
    showResult(d);
  }).catch(function(e){
    document.getElementById('pfResult').innerHTML='<div class="alert-error">'+esc(e.message)+'</div>';
    showResult({error:e.message});
  });
}
function calcAge(dob){
  if(!dob||dob.length<7)return 30;
  var p=dob.split('-');var y=parseInt(p[0]),m=parseInt(p[1]);
  var n=new Date();var ny=n.getFullYear(),nm=n.getMonth()+1;
  var a=ny-y;if(nm<m)a--;
  return a>16&&a<70?a:30;
}

// === 2. Plan ===
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
        {label:'月缴费(元)',data:schemes.map(function(s){return s.monthly_cost}),backgroundColor:'rgba(59,130,246,0.7)'},
        {label:'预计月养老金(元)',data:schemes.map(function(s){return s.projected_pension}),backgroundColor:'rgba(5,150,105,0.7)'},
        {label:'税后月领(元)',data:schemes.map(function(s){return s.after_tax_pension||0}),backgroundColor:'rgba(124,58,237,0.7)'}
      ]
    },
    options:{responsive:true,maintainAspectRatio:false,plugins:{legend:{position:'bottom',labels:{font:{size:11}}}},scales:{y:{beginAtZero:true,ticks:{callback:function(v){return v+'元'}}}}}
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
      labels:cf.map(function(c){return '第'+c.year+'年'}),
      datasets:[
        {label:'年缴费',data:cf.map(function(c){return c.payment}),borderColor:'#3B82F6',backgroundColor:'rgba(59,130,246,0.1)',fill:true,tension:0.3},
        {label:'年补贴',data:cf.map(function(c){return c.subsidy}),borderColor:'#F59E0B',backgroundColor:'rgba(245,158,11,0.1)',fill:true,tension:0.3},
        {label:'累计余额',data:cf.map(function(c){return c.balance}),borderColor:'#059669',backgroundColor:'rgba(5,150,105,0.1)',fill:true,tension:0.3}
      ]
    },
    options:{responsive:true,maintainAspectRatio:false,plugins:{legend:{position:'bottom',labels:{font:{size:11}}}},scales:{y:{ticks:{callback:function(v){return (v/10000).toFixed(1)+'万'}}}}}
  });
}
function onSimChange(){
  var base=parseFloat(document.getElementById('sim-base')?.value)||5000;
  var years=parseInt(document.getElementById('sim-years')?.value)||15;
  var budget=parseFloat(document.getElementById('sim-budget')?.value)||2000;
  var baseVal=document.getElementById('sim-base-val');if(baseVal)baseVal.textContent=base+'元';
  var yearsVal=document.getElementById('sim-years-val');if(yearsVal)yearsVal.textContent=years+'年';
  var budgetVal=document.getElementById('sim-budget-val');if(budgetVal)budgetVal.textContent=budget+'元';
  var cityCode=_formData.city||'310000';
  var avgSalary=12383,pensionRate=0.08,medicalRate=0.02;
  var cities={'310000':{avgSalary:12383},'110000':{avgSalary:15764},'440300':{avgSalary:14530},'440100':{avgSalary:13795},'330100':{avgSalary:9625}};
  var ci=cities[cityCode];if(ci)avgSalary=ci.avgSalary;
  var monthlyCost=base*(pensionRate+medicalRate);
  var personalBalance=base*pensionRate*years*12*Math.pow(1.03,years/2);
  var retireAge=_formData.gender==='male'?63:58;
  if(_formData.gender==='female'&&_formData.origAge===50)retireAge=55;
  var divisor=retireAge>=60?139:retireAge>=55?170:195;
  var personalPension=personalBalance/divisor;
  var basicPension=(avgSalary+base)/2*years*0.01;
  var total=Math.round(basicPension+personalPension);
  var el=document.getElementById('sim-result');
  if(el)el.textContent='￥'+total.toLocaleString()+' /月';
  var detail=document.getElementById('sim-detail');
  if(detail)detail.textContent='基础养老金 '+Math.round(basicPension)+' + 个人账户 '+Math.round(personalPension)+' | 月缴 '+Math.round(monthlyCost)+'元 | 投入产出比 '+(monthlyCost>0?(total/monthlyCost).toFixed(1):'-')+'x';
}
function showPlan(){
  var age=calcAge(_formData.dob);
  document.getElementById('app').innerHTML=
    '<h2>方案生成</h2><p style="font-size:13px;color:#6B7280;margin-bottom:12px">基于保存的用户画像生成社保方案</p>'+
    '<div style="font-size:13px;color:#374151;margin-bottom:12px;padding:12px;background:#F9FAFB;border-radius:8px">'+
    '出生: '+_formData.dob+' | 性别: '+(_formData.gender==='male'?'男':'女')+' | 就业: '+_formData.employment+' | '+
    '缴费: '+_formData.months+'月 | 预算: '+_formData.budget+'元/月 | 养老金总额: '+_formData.pensionTotal+'元</div>'+
    '<button class="btn btn-primary" onclick="onGeneratePlan()" id="genBtn">生成方案</button>'+
    '<div id="planResult"></div>';
}
function onGeneratePlan(){
  var btn=document.getElementById('genBtn');btn.disabled=true;btn.textContent='生成中...';
  var age=calcAge(_formData.dob);
  var req={
    age:age,gender:_formData.gender,
    original_pension_age:_formData.origAge||0,
    employment:_formData.employment,
    contribution_years:Math.floor(_formData.months/12),
    contribution_months:_formData.months,
    current_balance:_formData.pensionTotal,
    local_avg_salary:0,
    monthly_budget:_formData.budget,
    priority:'balanced',
  };
  api('POST','/v1/plans/generate',req).then(function(plan){
    btn.disabled=false;btn.textContent='再次生成';
    window._lastPlan=plan;
    window._lastPlanId=plan.plan_id;
    var schemes=plan.recommended_schemes||[];
    var maxROI=0,schemesWithROI=schemes.map(function(s){
      var roi=s.monthly_cost>0?s.projected_pension/s.monthly_cost:0;
      if(roi>maxROI)maxROI=roi;
      s._roi=roi;
      return s;
    });
    var h='<div class="alert-success" style="margin-top:12px">方案生成成功！共 '+schemes.length+' 个方案</div>';
    h+='<div class="chart-container" style="margin:12px 0"><canvas id="schemeChart"></canvas></div>';
    schemesWithROI.forEach(function(s,i){
      var isRec=s._roi===maxROI;
      var roiClass=s._roi>=3?'roi-green':s._roi>=2?'roi-yellow':'roi-red';
      h+='<div class="scheme-card'+(isRec?' recommended':'')+'">';
      h+='<div class="scheme-header"><strong style="font-size:15px">'+esc(s.name)+'</strong>';
      if(isRec)h+='<span class="rec-badge">推荐</span>';
      h+='<span class="roi-badge '+roiClass+'">ROI '+s._roi.toFixed(1)+'x</span></div>';
      h+='<div class="scheme-metrics">';
      h+='<div class="scheme-metric"><div class="value">'+s.monthly_cost.toFixed(0)+'</div><div class="label">月缴(元)</div></div>';
      h+='<div class="scheme-metric"><div class="value" style="color:#059669">'+s.projected_pension.toFixed(0)+'</div><div class="label">预计月养老金</div></div>';
      if(s.after_tax_pension>0)h+='<div class="scheme-metric"><div class="value" style="color:#7C3AED">'+s.after_tax_pension.toFixed(0)+'</div><div class="label">税后月领</div></div>';
      if(s.annual_subsidy>0)h+='<div class="scheme-metric"><div class="value" style="color:#D97706">'+s.annual_subsidy.toFixed(0)+'</div><div class="label">年补贴</div></div>';
      h+='<div class="scheme-metric"><div class="value" style="font-size:16px;color:#6B7280">'+s.remaining_months+'</div><div class="label">还需缴(月)</div></div>';
      h+='</div>';
      if(s.subsidy_policy)h+='<div style="margin-top:8px;font-size:12px;color:#6B7280">'+esc(s.subsidy_policy)+'</div>';
      if(s.cashflow&&s.cashflow.length>0){
        h+='<div class="cashflow-toggle" onclick="toggleCashflow('+i+')">查看现金流趋势 ▼</div>';
        h+='<div class="cashflow-panel" id="cf-panel-'+i+'"><div class="chart-container"><canvas id="cf-chart-'+i+'"></canvas></div></div>';
      }
      h+='</div>';
    });
    if(schemes.length>0){
      h+='<div class="sim-section"><h3 style="font-size:14px;color:#1A56DB;margin-bottom:8px">退休金模拟器</h3>';
      h+='<div class="sim-slider-row"><label>缴费基数</label><input type="range" id="sim-base" min="'+schemes[0].base_salary+'" max="'+schemes[schemes.length-1].base_salary+'" step="100" value="'+schemes[0].base_salary+'" oninput="onSimChange()"><span class="sim-val" id="sim-base-val">'+schemes[0].base_salary+'元</span></div>';
      h+='<div class="sim-slider-row"><label>缴费年限</label><input type="range" id="sim-years" min="1" max="40" step="1" value="'+Math.max(1,Math.floor(_formData.months/12))+'" oninput="onSimChange()"><span class="sim-val" id="sim-years-val">'+Math.max(1,Math.floor(_formData.months/12))+'年</span></div>';
      h+='<div class="sim-slider-row"><label>月预算</label><input type="range" id="sim-budget" min="500" max="10000" step="100" value="'+_formData.budget+'" oninput="onSimChange()"><span class="sim-val" id="sim-budget-val">'+_formData.budget+'元</span></div>';
      h+='<div class="sim-result" id="sim-result">计算中...</div>';
      h+='<div class="sim-result-sub" id="sim-detail"></div></div>';
    }
    h+='<button class="btn btn-outline" style="margin-top:12px" id="viewReportBtn" data-plan-id="'+esc(plan.plan_id)+'">查看完整报告</button>';
    document.getElementById('planResult').innerHTML=h;
    var vrBtn=document.getElementById('viewReportBtn');
    if(vrBtn)vrBtn.onclick=function(){viewReport(this.getAttribute('data-plan-id'));};
    _cfCharts={};
    renderSchemeChart(schemesWithROI);
    schemesWithROI.forEach(function(s,i){
      if(s.cashflow&&s.cashflow.length>0)renderCashflowChart(i,s.cashflow);
    });
    if(schemes.length>0)onSimChange();
  }).catch(function(e){btn.disabled=false;btn.textContent='生成方案';document.getElementById('planResult').innerHTML='<div class="alert-error">'+esc(e.message)+'</div>'});
}
function viewReport(pid){
  window._lastPlanId=pid;
  switchTab('report');
}

// === 3. Report ===
function showReport(){
  var pid=window._lastPlanId||'';
  document.getElementById('app').innerHTML=
    '<h2>方案报告</h2><div class="form-row"><div style="flex:3"><label>方案ID</label><input id="rpt-id" value="'+esc(pid)+'" placeholder="输入方案ID"></div>'+
    '<div style="flex:1;display:flex;align-items:end"><button class="btn btn-primary" onclick="onLoadReport()">查看报告</button></div></div>'+
    '<div id="reportFrame" style="margin-top:12px;display:none"><iframe id="rptIframe" style="width:100%;height:600px;border:1px solid #E5E7EB;border-radius:8px"></iframe></div>';
}
function onLoadReport(){
  var pid=document.getElementById('rpt-id').value;
  if(!pid)return;
  document.getElementById('reportFrame').style.display='block';
  fetch(API+'/v1/plans/report?plan_id='+encodeURIComponent(pid),{headers:{'Authorization':'Bearer '+_token}})
    .then(function(r){return r.text()})
    .then(function(html){
      document.getElementById('rptIframe').srcdoc=html;
    });
}

// === 4. Compliance ===
function showCompliance(){
  document.getElementById('app').innerHTML=
    '<h2>合规检查清单</h2><div class="form-row"><div style="flex:3"><label>城市</label><select id="comp-city"><option value="310000">上海</option><option value="110000">北京</option><option value="440100">广州</option><option value="330100">杭州</option><option value="440300">深圳</option></select></div>'+
    '<div style="flex:1;display:flex;align-items:end"><button class="btn btn-primary" onclick="onGetCompliance()">查询</button></div></div>'+
    '<div id="compResult"></div>';
}
function onGetCompliance(){
  var city=document.getElementById('comp-city').value;
  api('GET','/v1/compliance/checklist?city_code='+city).then(function(data){
    var h='<div style="margin-top:12px"><div class="alert-success">匹配 '+(data.matched_policies||[]).length+' 条政策</div>';
    (data.matched_policies||[]).forEach(function(p){
      h+='<div class="card" style="padding:12px;margin:8px 0">'+
        '<div style="display:flex;justify-content:space-between;align-items:center">'+
        '<strong>'+esc(p.policy_type)+'</strong>'+
        (p.is_eligible?'<span class="badge bg-green">可申请</span>':'<span class="badge bg-red">条件不足</span>')+'</div>'+
        '<div style="font-size:12px;color:#6B7280;margin-top:4px">'+esc(p.subsidy_calc_method)+'</div>';
      if(p.unmet_conditions&&p.unmet_conditions.length>0)
        h+='<div style="margin-top:6px;font-size:12px;color:#EF4444">未满足: '+p.unmet_conditions.join(', ')+'</div>';
      if(p.processing_steps&&p.processing_steps.length>0)
        h+='<div style="margin-top:6px;font-size:12px;color:#6B7280">流程: '+p.processing_steps.map(function(s){return s.name}).join(' → ')+'</div>';
      h+='</div>';
    });
    h+='</div>';
    document.getElementById('compResult').innerHTML=h;
    showResult(data);
  }).catch(function(e){document.getElementById('compResult').innerHTML='<div class="alert-error">'+esc(e.message)+'</div>'});
}

// === 5. Guide ===
function showGuide(){
  document.getElementById('app').innerHTML=
    '<h2>办理指南</h2><div class="form-row"><div style="flex:3"><label>城市</label><select id="guide-city"><option value="310000">上海</option><option value="110000">北京</option><option value="440100">广州</option><option value="330100">杭州</option><option value="440300">深圳</option></select></div>'+
    '<div style="flex:1;display:flex;align-items:end"><button class="btn btn-primary" onclick="onLoadGuide()">查看</button></div></div>'+
    '<div id="guideFrame" style="margin-top:12px;display:none"><iframe id="guideIframe" style="width:100%;height:600px;border:1px solid #E5E7EB;border-radius:8px"></iframe></div>';
}
function onLoadGuide(){
  var city=document.getElementById('guide-city').value;
  document.getElementById('guideFrame').style.display='block';
  ensureToken().then(function(){fetch(API+'/v1/guide?city_code='+city,{headers:{'Authorization':'Bearer '+_token}})
    .then(function(r){return r.text()})
    .then(function(html){
      document.getElementById('guideIframe').srcdoc=html;
    });});
}

// === 6. Policies ===
function showPolicies(){
  document.getElementById('app').innerHTML=
    '<h2>政策查询</h2><div class="form-row"><div><label>城市</label><select id="pol-city"><option value="">全部</option><option value="310000">上海</option><option value="110000">北京</option><option value="440100">广州</option><option value="330100">杭州</option><option value="440300">深圳</option></select></div>'+
    '<div><label>类型</label><select id="pol-type"><option value="">全部</option><option value="subsidy">补贴</option><option value="pension">养老</option><option value="medical">医疗</option><option value="unemployment">失业</option><option value="injury">工伤</option><option value="maternity">生育</option><option value="housing_fund">公积金</option><option value="training">培训</option></select></div>'+
    '<div style="display:flex;align-items:end"><button class="btn btn-primary" onclick="onQueryPolicies()">查询</button></div></div>'+
    '<div id="polResult"></div>';
}
function onQueryPolicies(){
  var city=document.getElementById('pol-city').value;
  var type=document.getElementById('pol-type').value;
  var params='';if(city)params+='region_code='+city;if(type)params+=(params?'&':'')+'policy_type='+type;
  api('GET','/v1/policies'+(params?'?'+params:'')).then(function(items){
    if(!items||items.length===0){document.getElementById('polResult').innerHTML='<div class="alert-error">没有找到政策</div>';return}
    var h='<div style="margin-top:12px;font-size:13px;color:#6B7280">共 '+items.length+' 条</div><table><tr><th>类型</th><th>城市</th><th>计算方式</th><th>来源</th><th>状态</th></tr>';
    items.slice(0,20).forEach(function(p){
      h+='<tr><td><span class="badge bg-blue">'+esc(p.policy_type)+'</span></td><td>'+esc(p.region_code)+'</td><td>'+esc(p.subsidy_calc_method||'-')+'</td><td style="font-size:11px">'+esc(p.source_name||'-')+'</td><td><span class="badge '+(p.status==='verified'?'bg-green':p.status==='pending_review'?'bg-yellow':'bg-red')+'">'+esc(p.status)+'</span></td></tr>';
    });
    h+='</table>';
    document.getElementById('polResult').innerHTML=h;
    showResult(items);
  }).catch(function(e){document.getElementById('polResult').innerHTML='<div class="alert-error">'+esc(e.message)+'</div>'});
}

// === 7. Rights ===
function showRights(){
  document.getElementById('app').innerHTML=
    '<h2>权益监控</h2><div style="display:flex;gap:12px;flex-wrap:wrap;margin-bottom:12px">'+
    '<button class="btn btn-primary" onclick="onGetPayment()">缴费状态</button>'+
    '<button class="btn btn-outline" onclick="onGetAlerts()">告警列表</button></div><div id="rightsResult"></div>';
}
function onGetPayment(){
  api('GET','/v1/rights/payment-status').then(function(d){
    document.getElementById('rightsResult').innerHTML='';
    showResult(d);
  }).catch(function(e){document.getElementById('rightsResult').innerHTML='<div class="alert-error">'+esc(e.message)+'</div>'});
}
function onGetAlerts(){
  api('GET','/v1/rights/alerts').then(function(items){
    var h='<div style="margin-top:12px;font-size:13px;color:#6B7280">共 '+(items||[]).length+' 条告警</div>';
    (items||[]).forEach(function(a){
      h+='<div class="card" style="padding:12px;margin:8px 0"><div style="display:flex;justify-content:space-between">'+
        '<span class="badge '+(a.severity==='high'?'bg-red':a.severity==='medium'?'bg-yellow':'bg-blue')+'">'+esc(a.alert_type)+'</span>'+
        (a.is_read?'<span style="font-size:11px;color:#9CA3AF">已读</span>':'<span class="badge bg-yellow">未读</span>')+'</div>'+
        '<div style="font-weight:600;margin-top:4px">'+esc(a.title)+'</div>'+
        '<div style="font-size:12px;color:#6B7280">'+esc(a.message)+'</div></div>';
    });
    if(!items||items.length===0)h+='<div style="text-align:center;color:#9CA3AF;padding:20px">暂无告警</div>';
    document.getElementById('rightsResult').innerHTML=h;
    showResult(items);
  }).catch(function(e){document.getElementById('rightsResult').innerHTML='<div class="alert-error">'+esc(e.message)+'</div>'});
}

// === 8. Feedback ===
function showFeedback(){
  document.getElementById('app').innerHTML=
    '<h2>意见反馈</h2><div class="form-row"><div><label>类型</label><select id="fb-cat"><option value="general">功能建议</option><option value="bug">问题报告</option><option value="other">其他</option></select></div></div>'+
    '<div><label>反馈内容</label><textarea id="fb-content" placeholder="请详细描述..."></textarea></div>'+
    '<div><label>联系方式</label><input id="fb-contact" placeholder="手机或邮箱(选填)"></div>'+
    '<button class="btn btn-success" onclick="onSubmitFeedback()">提交反馈</button><div id="fbResult"></div>';
}
function onSubmitFeedback(){
  var data={category:document.getElementById('fb-cat').value,content:document.getElementById('fb-content').value,contact:document.getElementById('fb-contact').value};
  if(!data.content){document.getElementById('fbResult').innerHTML='<div class="alert-error">请输入反馈内容</div>';return}
  api('POST','/v1/feedback',data).then(function(){
    document.getElementById('fbResult').innerHTML='<div class="alert-success">反馈已提交，感谢您的意见！</div>';
  }).catch(function(e){document.getElementById('fbResult').innerHTML='<div class="alert-error">'+esc(e.message)+'</div>'});
}
</script>
</body>
</html>`
