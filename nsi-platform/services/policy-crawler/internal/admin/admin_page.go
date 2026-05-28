package admin

import "net/http"

func AdminPageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(adminHTML))
	})
}

const adminHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>AI社保智筹 - 管理后台</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#F5F7FA;color:#333;font-size:14px}
.header{background:linear-gradient(135deg,#1A56DB,#3B82F6);color:#fff;padding:14px 24px;display:flex;justify-content:space-between;align-items:center}
.header h1{font-size:18px;font-weight:600}.header span{font-size:12px;opacity:0.8}
.nav{display:flex;gap:0;background:#fff;border-bottom:2px solid #E5E7EB;padding:0 8px;overflow-x:auto}
.nav-item{padding:10px 18px;cursor:pointer;font-size:13px;color:#6B7280;border-bottom:2px solid transparent;margin-bottom:-2px;white-space:nowrap}
.flex-row{display:flex;flex-wrap:wrap}
.nav-item.active{color:#1A56DB;border-bottom-color:#1A56DB;font-weight:600}
.nav-item:hover{color:#1A56DB}
.container{max-width:1200px;margin:16px auto;padding:0 16px}
.card{background:#fff;border-radius:10px;padding:16px;margin-bottom:12px;box-shadow:0 1px 4px rgba(0,0,0,0.06)}
.stat-row{display:flex;gap:12px;flex-wrap:wrap}
.stat-card{flex:1;min-width:120px;background:#fff;border-radius:10px;padding:16px;text-align:center;box-shadow:0 1px 4px rgba(0,0,0,0.06)}
.stat-num{font-size:28px;font-weight:700;display:block}
.stat-label{font-size:12px;color:#6B7280;display:block;margin-top:4px}
.c-green{color:#059669}.c-orange{color:#D97706}.c-red{color:#EF4444}.c-blue{color:#1A56DB}
table{width:100%;border-collapse:collapse;font-size:13px}
th{text-align:left;padding:8px 10px;border-bottom:2px solid #E5E7EB;color:#6B7280;font-weight:600;font-size:12px}
td{padding:8px 10px;border-bottom:1px solid #F3F4F6}
tr:hover{background:#F9FAFB}
.badge{display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:500}
.bg-green{background:#D1FAE5;color:#059669}.bg-yellow{background:#FEF3C7;color:#D97706}
.bg-red{background:#FEE2E2;color:#EF4444}.bg-blue{background:#DBEAFE;color:#1A56DB}.bg-purple{background:#EDE9FE;color:#7C3AED}
.btn{padding:6px 14px;border:none;border-radius:6px;font-size:12px;cursor:pointer;font-weight:500}
.btn-sm{padding:4px 10px;font-size:11px}
.btn-primary{background:#1A56DB;color:#fff}
.btn-success{background:#059669;color:#fff}
.btn-danger{background:#EF4444;color:#fff}
.btn-outline{background:#fff;border:1px solid #D1D5DB;color:#374151}
.btn:disabled{opacity:0.5;cursor:not-allowed}
.toggle{position:relative;display:inline-block;width:36px;height:20px}
.toggle input{opacity:0;width:0;height:0}
.slider{position:absolute;cursor:pointer;top:0;left:0;right:0;bottom:0;background:#ccc;border-radius:20px;transition:.3s}
.slider:before{content:'';position:absolute;height:14px;width:14px;left:3px;bottom:3px;background:#fff;border-radius:50%;transition:.3s}
.toggle input:checked+.slider{background:#1A56DB}
.toggle input:checked+.slider:before{transform:translateX(16px)}
.toast{position:fixed;bottom:20px;right:20px;padding:10px 16px;border-radius:8px;color:#fff;font-size:13px;z-index:999;display:none}
.spinner{display:inline-block;width:14px;height:14px;border:2px solid #E5E7EB;border-top-color:#1A56DB;border-radius:50%;animation:spin .6s linear infinite;vertical-align:middle;margin-right:6px}
@keyframes spin{to{transform:rotate(360deg)}}
textarea{width:100%;min-height:200px;border:1px solid #D1D5DB;border-radius:6px;padding:10px;font-size:13px;font-family:monospace;resize:vertical}
.progress-bar{height:8px;background:#E5E7EB;border-radius:4px;overflow:hidden;margin:8px 0}
.progress-fill{height:100%;background:#1A56DB;border-radius:4px;transition:width .3s;width:0%}
</style>
</head>
<body>
<div class="header"><h1>AI社保智筹 - 管理后台</h1><span>v1.0.0</span></div>
<div class="nav" id="navBar"></div>
<div class="container" id="app"><div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>加载中...</div></div>
<div class="toast" id="toast"></div>

<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js"></script>
<script>
var navItems=[
  {id:'dashboard',label:'仪表盘'},
  {id:'sources',label:'数据源管理'},
  {id:'claims',label:'政策审核'},
  {id:'search',label:'语义搜索'},
  {id:'extract',label:'AI提取'},
  {id:'relevance',label:'相关性规则'},
  {id:'logs',label:'爬取日志'},
  {id:'pipeline',label:'数据流水线'},
  {id:'extractLogs',label:'提取日志'},
  {id:'failures',label:'失败分析'},
  {id:'import',label:'+导入政策',style:'color:#059669;font-weight:600'},
  {id:'llm-gateway',label:'模型管理'}
];
var currentPanel='dashboard';

function initNav(){var h='';navItems.forEach(function(n){h+='<div class="nav-item'+(n.id==currentPanel?' active':'')+'" data-panel="'+n.id+'"'+(n.style?' style="'+n.style+'"':'')+'>'+n.label+'</div>'});document.getElementById('navBar').innerHTML=h;
document.getElementById('navBar').addEventListener('click',function(e){var item=e.target.closest('.nav-item');if(item){switchPanel(item.dataset.panel)}})}
initNav();

function showToast(msg,type){var t=document.getElementById('toast');t.textContent=msg;t.style.display='block';t.style.background=type==='success'?'#059669':'#EF4444';t.style.opacity='1';setTimeout(function(){t.style.display='none'},3000)}
function esc(s){if(typeof s!=='string')return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML}

function switchPanel(id){
  currentPanel=id;
  window.location.hash='#'+id;
  document.querySelectorAll('.nav-item').forEach(function(n){n.classList.remove('active')});
  var el=document.querySelector('[data-panel="'+id+'"]');if(el)el.classList.add('active');
  var app=document.getElementById('app');
  app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>加载中...</div>';
  if(id==='dashboard')loadDashboard();
  else if(id==='sources')loadSources();
  else if(id==='claims')loadClaims();
  else if(id==='search')loadSearch();
  else if(id==='extract')loadExtract();
  else if(id==='relevance')loadRelevanceRules();
  else if(id==='logs')loadLogs();
  else if(id==='extractLogs')loadExtractLogs();
   else if(id==='failures')loadFailures();
   else if(id==='pipeline')loadPipeline();
   else if(id==='import')showImportForm();
   else if(id==='llm-gateway'){
     var h='';
     h+='<div class="card"><iframe src="/llm-gateway/admin" style="width:100%;height:800px;border:0;"></iframe></div>';
     app.innerHTML=h;
   }
}

// 支持通过 hash 直接定位页面
window.addEventListener('hashchange',function(){
  var h=window.location.hash.replace('#','');
  if(h && navItems.some(function(n){return n.id===h})) switchPanel(h);
});
// 初始化时读取 hash
(function(){
  var h=window.location.hash.replace('#','');
  if(h && navItems.some(function(n){return n.id===h})){switchPanel(h);return}
  switchPanel('dashboard');
})();

function loadDashboard(){
  fetch('/admin/dashboard').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    var s=d.data;
    // 顶部指标卡
    var h='<div class="stat-row">'+
      '<div class="stat-card"><span class="stat-num c-blue">'+s.total_sources+'</span><span class="stat-label">总数据源</span></div>'+
      '<div class="stat-card"><span class="stat-num c-green">'+s.active_sources+'</span><span class="stat-label">活跃中</span></div>'+
      '<div class="stat-card"><span class="stat-num c-blue">'+s.total_claims+'</span><span class="stat-label">政策总量</span></div>'+
      '<div class="stat-card"><span class="stat-num c-green">'+s.verified_claims+'</span><span class="stat-label">已通过</span></div>'+
      '<div class="stat-card"><span class="stat-num c-orange">'+s.pending_claims+'</span><span class="stat-label">待审核</span></div>'+
      '<div class="stat-card"><span class="stat-num c-red">'+s.unverified_claims+'</span><span class="stat-label">已驳回</span></div>'+
      '</div>'+
      '<div class="stat-row">'+
      '<div class="stat-card"><span class="stat-num c-blue">'+s.with_embedding+'</span><span class="stat-label">有向量 ('+(s.total_claims>0?Math.round(s.with_embedding/s.total_claims*100):0)+'%)</span></div>'+
      '<div class="stat-card"><span class="stat-num c-green">'+s.with_policy_url+'</span><span class="stat-label">有原文链接</span></div>'+
      '<div class="stat-card"><span class="stat-num c-orange">'+s.pending_extraction+'</span><span class="stat-label">待提取</span></div>'+
      '<div class="stat-card"><span class="stat-num '+(s.today_crawls>0?'c-blue':'c-orange')+'">'+s.today_crawls+'</span><span class="stat-label">今日爬取</span></div>'+
      '<div class="stat-card"><span class="stat-num '+(s.failed_crawls>0?'c-red':'c-green')+'">'+s.failed_crawls+'</span><span class="stat-label">今日失败</span></div>'+
      '<div class="stat-card"><span class="stat-num '+(s.extract_success_rate>80?'c-green':'c-orange')+'">'+(s.extract_success_rate||0).toFixed(1)+'%</span><span class="stat-label">提取成功率</span></div>'+
      '</div>';

    // 政策类型分布
    h+='<div class="flex-row" style="gap:12px;flex-wrap:wrap">'+
      '<div class="card" style="flex:1;min-width:300px"><h3 style="font-size:15px;margin-bottom:12px">政策类型分布</h3><div id="typeChart">';
    var typeMap={'subsidy':'补贴','pension':'养老','medical':'医疗','unemployment':'失业','injury':'工伤','maternity':'生育','housing_fund':'公积金','training':'培训'};
    var typeColors={'subsidy':'#059669','pension':'#1A56DB','medical':'#D97706','unemployment':'#7C3AED','injury':'#EF4444','maternity':'#EC4899','housing_fund':'#0EA5E9','training':'#F59E0B'};
    if(s.policy_type_dist){
      var maxType=0;Object.values(s.policy_type_dist).forEach(function(v){if(v>maxType)maxType=v});
      Object.keys(s.policy_type_dist).forEach(function(k){
        var pct=s.policy_type_dist[k]/maxType*100;
        h+='<div style="display:flex;align-items:center;margin:6px 0;font-size:13px">'+
          '<span style="width:60px;color:#6B7280">'+(typeMap[k]||k)+'</span>'+
          '<span style="flex:1;height:20px;background:#F3F4F6;border-radius:4px;overflow:hidden;margin:0 8px">'+
          '<span style="display:block;height:100%;width:'+pct+'%;background:'+(typeColors[k]||'#1A56DB')+';border-radius:4px"></span></span>'+
          '<span style="width:40px;text-align:right;font-weight:600;font-size:13px">'+s.policy_type_dist[k]+'</span></div>';
      });
    }
    h+='</div></div>';

    // 城市分布
    h+='<div class="card" style="flex:1;min-width:300px"><h3 style="font-size:15px;margin-bottom:12px">城市分布 TOP10</h3><div id="regionChart">';
    var regionMap={'110000':'北京','310000':'上海','440100':'广州','330100':'杭州','440300':'深圳','310115':'上海浦东','330108':'杭州滨江','000000':'未知','N/A':'N/A','NOT_FOUND':'未识别'};
    if(s.region_dist){
      var maxReg=0;Object.values(s.region_dist).forEach(function(v){if(v>maxReg)maxReg=v});
      Object.keys(s.region_dist).forEach(function(k){
        var pct=s.region_dist[k]/maxReg*100;
        h+='<div style="display:flex;align-items:center;margin:6px 0;font-size:13px">'+
          '<span style="width:70px;color:#6B7280">'+(regionMap[k]||k)+'</span>'+
          '<span style="flex:1;height:20px;background:#F3F4F6;border-radius:4px;overflow:hidden;margin:0 8px">'+
          '<span style="display:block;height:100%;width:'+pct+'%;background:#1A56DB;border-radius:4px"></span></span>'+
          '<span style="width:40px;text-align:right;font-weight:600;font-size:13px">'+s.region_dist[k]+'</span></div>';
      });
    }
    h+='</div></div></div>';

    // 7天爬取趋势
    h+='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">近7天爬取趋势</h3><div style="display:flex;align-items:end;gap:6px;height:120px;padding:10px 0">';
    if(s.crawl_trend_7d){
      var maxCrawl=0;s.crawl_trend_7d.forEach(function(v){if(v>maxCrawl)maxCrawl=v});
      var days=['7天前','6天前','5天前','4天前','3天前','昨天','今天'];
      s.crawl_trend_7d.forEach(function(v,i){
        var hgt=maxCrawl>0?v/maxCrawl*100:0;
        h+='<div style="flex:1;display:flex;flex-direction:column;align-items:center;height:100%;justify-content:end">'+
          '<span style="font-size:11px;color:#6B7280;margin-bottom:4px">'+v+'</span>'+
          '<span style="display:block;width:100%;max-width:40px;height:'+hgt+'%;background:'+(i===6?'#1A56DB':'#93C5FD')+';border-radius:4px 4px 0 0;min-height:'+(v>0?'8px':'0')+'"></span>'+
          '<span style="font-size:10px;color:#9CA3AF;margin-top:4px">'+days[i]+'</span></div>';
      });
    }
    h+='</div></div>';

    // 失败分析
    h+='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">爬取失败分析（按来源+原因）</h3><div id="crawlFailChart">加载中...</div></div>'+
      '<div class="card"><h3 style="font-size:15px;margin-bottom:12px">AI提取失败分析（按来源+原因）</h3><div id="extractFailChart">加载中...</div></div>';

    // 数据源一览
    h+='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">数据源一览</h3><div id="miniSources">加载中...</div></div>';
    document.getElementById('app').innerHTML=h;

    // 加载失败数据
    fetch('/admin/failures').then(function(r){return r.json()}).then(function(d3){
      if(d3.code!==0)return;
      var f=d3.data;
      var renderFail=function(containerId,items,label,color){
        if(!items||items.length===0){document.getElementById(containerId).innerHTML='<span style="color:#9CA3AF;font-size:13px">暂无失败记录</span>';return}
        var max=1;items.forEach(function(x){if(x.count>max)max=x.count});
        var html='<table style="width:100%;font-size:13px">';
        items.forEach(function(x){
          var pct=x.count/max*100;
          html+='<tr><td style="padding:4px 8px;white-space:nowrap;font-weight:600;width:120px">'+esc(x.source_name||x.source_id)+'</td>'+
            '<td style="padding:4px 8px;color:#6B7280;font-size:12px;max-width:300px;overflow:hidden;text-overflow:ellipsis">'+esc(x.error_message)+'</td>'+
            '<td style="padding:4px 8px;width:120px"><span style="display:inline-block;height:18px;width:'+pct+'%;min-width:12px;background:'+color+';border-radius:3px;vertical-align:middle"></span></td>'+
            '<td style="padding:4px 8px;text-align:right;font-weight:600;width:40px">'+x.count+'</td></tr>';
        });
        html+='</table>';
        document.getElementById(containerId).innerHTML=html;
      };
      renderFail('crawlFailChart',f.crawl_failures,'爬取失败','#EF4444');
      renderFail('extractFailChart',f.extract_failures,'提取失败','#D97706');
    }).catch(function(){document.getElementById('crawlFailChart').innerHTML='<span style="color:#EF4444">加载失败</span>'});

    fetch('/admin/sources').then(function(r2){return r2.json()}).then(function(d2){
      if(d2.code!==0)return;
      var h2='<table><tr><th>名称</th><th>级别</th><th>地区</th><th>状态</th><th>政策数</th><th>最近爬取</th></tr>';
      d2.data.slice(0,10).forEach(function(s){
        h2+='<tr><td style="font-weight:600;font-size:13px">'+esc(s.source_name)+'</td><td><span class="badge '+(s.source_level==='HIGH'?'bg-green':'bg-yellow')+'">'+s.source_level+'</span></td><td>'+(s.region_code||'-')+'</td><td>'+(s.enabled?'<span class="badge bg-green">启用</span>':'<span class="badge bg-red">禁用</span>')+'</td><td>'+(s.claims_count||'-')+'</td><td style="font-size:11px;color:#9CA3AF">'+(s.last_crawl||'-')+'</td></tr>'
      });
      document.getElementById('miniSources').innerHTML=h2;
    }).catch(function(){document.getElementById('miniSources').innerHTML='加载失败'});
  }).catch(function(e){document.getElementById('app').innerHTML='<div class="card"><h3>加载失败</h3><p>'+esc(e.message)+'</p></div>'});
}

var allSourcesData=[];

function loadSources(){
  fetch('/admin/sources').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    allSourcesData=d.data||[];
    renderSources();
  }).catch(function(e){document.getElementById('app').innerHTML='<div class="card"><h3>加载失败</h3><p>'+esc(e.message)+'</p></div>'});
}

function renderSources(){
  var ft=document.getElementById('src_ft')?document.getElementById('src_ft').value:'';
  var fl=document.getElementById('src_fl')?document.getElementById('src_fl').value:'';
  var fr=document.getElementById('src_fr')?document.getElementById('src_fr').value:'';
  var fi=document.getElementById('src_fi')?document.getElementById('src_fi').value.toLowerCase():'';
  var filtered=allSourcesData.filter(function(s){
    if(ft && s.crawl_type!==ft)return false;
    if(fl && s.source_level!==fl)return false;
    if(fr && (s.region_code||'')!==fr)return false;
    if(fi && s.source_id.toLowerCase().indexOf(fi)<0 && s.source_name.toLowerCase().indexOf(fi)<0)return false;
    return true;
  });
  var regions=[];allSourcesData.forEach(function(s){if(s.region_code && regions.indexOf(s.region_code)<0)regions.push(s.region_code)});
  var h='<div style="display:flex;gap:8px;margin-bottom:10px;flex-wrap:wrap;align-items:center">'+
    '<span style="font-size:12px;color:#6B7280">筛选:</span>'+
    '<select id="src_ft" onchange="renderSources()" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
    '<option value="">全部类型</option>'+sourceTypeList.map(function(t){return '<option value="'+t.v+'"'+(ft===t.v?' selected':'')+'>'+t.l+'</option>'}).join('')+'</select>'+
    '<select id="src_fl" onchange="renderSources()" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
    '<option value="">全部级别</option><option value="HIGH"'+(fl==='HIGH'?' selected':'')+'>HIGH</option><option value="MEDIUM"'+(fl==='MEDIUM'?' selected':'')+'>MEDIUM</option><option value="LOW"'+(fl==='LOW'?' selected':'')+'>LOW</option></select>'+
    '<select id="src_fr" onchange="renderSources()" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
    '<option value="">全部地区</option>';
  regions.forEach(function(r){h+='<option value="'+r+'"'+(fr===r?' selected':'')+'>'+r+'</option>'});
  h+='</select>'+
    '<input id="src_fi" oninput="renderSources()" placeholder="搜索ID/名称..." value="'+esc(fi)+'" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px;width:160px">'+
    '<span style="color:#6B7280;font-size:13px">共 '+filtered.length+' / '+allSourcesData.length+' 个数据源</span>'+
    '<span style="flex:1"></span><button class="btn btn-success" onclick="showSourceForm()">+ 新增数据源</button></div>';
  h+='<div style="overflow-x:auto"><table><tr><th>启用</th><th>名称</th><th>URL</th><th>级别</th><th>类型</th><th>地区</th><th>间隔</th><th>最近爬取</th><th>状态</th><th>操作</th></tr>';
  filtered.forEach(function(s){
      var actions='<div style="display:flex;gap:4px;flex-wrap:nowrap">';
      actions+='<button class="btn btn-outline btn-sm" onclick="showSourceForm(\''+esc(s.source_id)+'\')" title="编辑">✎</button>';
      actions+='<button class="btn btn-danger btn-sm" onclick="deleteSource(\''+esc(s.source_id)+'\',\''+esc(s.source_name).replace(/'/g,"\\'")+'\')" title="删除">✕</button>';
      if(s.crawl_type!=='manual'){
        actions+='<button class="btn btn-primary btn-sm" onclick="triggerCrawl(\''+esc(s.source_id)+'\')" title="触发爬取">▶</button>';
      }
      if(s.crawl_type==='rss'){
        actions+='<button class="btn btn-outline btn-sm" onclick="testRSS(\''+esc(s.source_url).replace(/'/g,"\\'")+'\')" title="测试RSS">🔍</button>';
      }
      if(s.crawl_type==='manual'){
        actions+='<button class="btn btn-outline btn-sm" onclick="showImportModal(\''+esc(s.source_id)+'\',\''+esc(s.source_name).replace(/'/g,"\\'")+'\')" title="导入内容">📤</button>';
      }
      actions+='</div>';
      h+='<tr><td><label class="toggle"><input type="checkbox" '+(s.enabled?'checked':'')+' data-sid="'+esc(s.source_id)+'" onchange="toggleSource(this)"><span class="slider"></span></label></td>'+
      '<td style="font-weight:600;font-size:13px">'+esc(s.source_name)+'</td>'+
      '<td style="font-size:11px;max-width:180px;overflow:hidden;text-overflow:ellipsis" title="'+esc(s.source_url)+'">'+esc(s.source_url||'-')+'</td>'+
      '<td><span class="badge '+(sourceLevelBadge[s.source_level]||'bg-yellow')+'">'+esc(s.source_level)+'</span></td>'+
      '<td><span class="badge '+(sourceTypeBadge[s.crawl_type]||'bg-blue')+'">'+(sourceTypeMap[s.crawl_type]||esc(s.crawl_type))+'</span></td>'+
      '<td>'+(s.region_code||'-')+'</td>'+
      '<td style="font-size:12px">'+(s.crawl_type==='manual'?'-':s.interval_sec+'s')+'</td>'+
      '<td style="font-size:11px;color:#9CA3AF">'+(s.last_crawl||'-')+'</td>'+
      '<td>'+(s.last_status==='success'?'<span class="badge bg-green">成功</span>':s.last_status==='failed'?'<span class="badge bg-red">失败</span>':'-')+'</td>'+
      '<td>'+actions+'</td></tr>';
    });
    h+='</table></div>'+
      '<div id="sourceModal" style="display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.4);z-index:1000">'+
      '<div style="background:#fff;border-radius:10px;padding:20px;max-width:500px;margin:60px auto;box-shadow:0 4px 20px rgba(0,0,0,0.15)">'+
      '<h3 id="sourceModalTitle" style="font-size:16px;margin-bottom:16px">新增数据源</h3>'+
      '<div style="display:grid;grid-template-columns:1fr 1fr;gap:10px">'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">数据源 ID</label><input id="sf_id" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">名称</label><input id="sf_name" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">类型</label><select id="sf_type" onchange="onTypeChange()" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      sourceTypeList.map(function(t){return '<option value="'+t.v+'">'+t.l+'</option>'}).join('')+'</select></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">级别</label><select id="sf_level" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      '<option value="HIGH">HIGH</option><option value="MEDIUM" selected>MEDIUM</option><option value="LOW">LOW</option></select></div>'+
      '<div id="sf_url_wrap" style="grid-column:1/3"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">URL</label><input id="sf_url" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px" placeholder="https://..."></div>'+
      '<div id="sf_interval_wrap"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">爬取间隔(秒)</label><input id="sf_interval" type="number" value="86400" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">省份 <span style="color:#EF4444">*</span></label><select id="sf_province" onchange="onProvinceChange()" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></select></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">城市</label><select id="sf_city" onchange="onCityChange()" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></select></div>'+
      '<div id="sf_district_wrap"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">区县</label><select id="sf_district" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></select></div>'+
      '<div style="grid-column:1/3;border-top:1px solid #E5E7EB;padding-top:8px;margin-top:4px"><span style="font-size:11px;color:#9CA3AF">爬取配置</span></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">代理URL (可选)</label><input id="sf_proxy" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px" placeholder="http://proxy:port"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">请求间隔(ms)</label><input id="sf_delay" type="number" value="0" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">最大并发</label><input id="sf_concurrent" type="number" value="1" min="1" max="10" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div style="display:flex;align-items:center;padding-top:18px"><label style="font-size:13px;display:flex;align-items:center;gap:4px;cursor:pointer"><input id="sf_robots" type="checkbox" checked>遵守 robots.txt</label></div>'+
      '</div>'+
      '<div style="display:flex;gap:8px;margin-top:16px;justify-content:flex-end">'+
      '<button class="btn btn-outline" onclick="closeSourceModal()">取消</button>'+
      '<button class="btn btn-primary" onclick="saveSource()">保存</button></div>'+
      '</div></div>'+
      '<div id="importModal" style="display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.4);z-index:1000">'+
      '<div style="background:#fff;border-radius:10px;padding:20px;max-width:550px;margin:60px auto;box-shadow:0 4px 20px rgba(0,0,0,0.15)">'+
      '<h3 id="importModalTitle" style="font-size:16px;margin-bottom:16px">导入内容</h3>'+
      '<input type="hidden" id="imp_source_id">'+
      '<div style="margin-bottom:10px"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">标题</label><input id="imp_title" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div style="margin-bottom:10px"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">来源URL (可选)</label><input id="imp_url" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div style="margin-bottom:10px"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">内容</label><textarea id="imp_content" style="width:100%;min-height:180px;border:1px solid #D1D5DB;border-radius:6px;padding:10px;font-size:13px;font-family:monospace;resize:vertical"></textarea></div>'+
      '<div style="display:flex;gap:8px;justify-content:flex-end">'+
      '<button class="btn btn-outline" onclick="closeImportModal()">取消</button>'+
      '<button class="btn btn-success" onclick="doSourceImport()">导入</button></div>'+
      '</div></div>'+
      '<div id="rssPreview" style="display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.4);z-index:1000">'+
      '<div style="background:#fff;border-radius:10px;padding:20px;max-width:500px;margin:60px auto;box-shadow:0 4px 20px rgba(0,0,0,0.15)">'+
      '<h3 style="font-size:16px;margin-bottom:12px">RSS 预览</h3>'+
      '<div id="rssPreviewContent"></div>'+
      '<div style="margin-top:12px;text-align:right"><button class="btn btn-outline" onclick="closeRSSPreview()">关闭</button></div>'+
      '</div></div>';
    document.getElementById('app').innerHTML=h;
}

function toggleSource(el){
  var id=el.dataset.sid,enabled=el.checked;
  fetch('/admin/sources/update',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({source_id:id,enabled:enabled})})
  .then(function(r){return r.json()}).then(function(d){showToast(enabled?'\u5df2\u542f\u7528':'\u5df2\u7981\u7528','success')})
  .catch(function(){showToast('\u64cd\u4f5c\u5931\u8d25','error')});
}

var sourceTypeMap={'govsite':'政府网站','file':'文件','rss':'RSS','manual':'手动','douyin':'抖音','wechat':'微信公众号'};
var sourceTypeBadge={'govsite':'bg-blue','file':'bg-yellow','rss':'bg-green','manual':'bg-yellow','douyin':'bg-red','wechat':'bg-purple'};
var sourceTypeList=[{v:'govsite',l:'政府网站'},{v:'file',l:'文件'},{v:'rss',l:'RSS'},{v:'manual',l:'手动'},{v:'douyin',l:'抖音'},{v:'wechat',l:'微信公众号'}];
var sourceLevelBadge={'HIGH':'bg-green','MEDIUM':'bg-yellow','LOW':'bg-blue'};
var editingSourceId='';

function onTypeChange(){
  var t=document.getElementById('sf_type').value;
  var urlWrap=document.getElementById('sf_url_wrap');
  var intvWrap=document.getElementById('sf_interval_wrap');
  if(t==='manual'){
    urlWrap.style.display='none';
    intvWrap.style.display='none';
  }else{
    urlWrap.style.display='';
    intvWrap.style.display='';
  }
}

function onProvinceChange(){
  var p=document.getElementById('sf_province').value;
  var citySel=document.getElementById('sf_city');
  var distSel=document.getElementById('sf_district');
  citySel.innerHTML='<option value="">请选择城市</option>';
  distSel.innerHTML='<option value="">请先选择城市</option>';
  if(!p)return;
  fetch('/admin/regions?parent='+p).then(function(r){return r.json()}).then(function(d){
    if(d.code===0){
      d.data.forEach(function(c){citySel.innerHTML+='<option value="'+c.code+'">'+c.name+'</option>'});
    }
  });
}

function onCityChange(){
  var c=document.getElementById('sf_city').value;
  var distSel=document.getElementById('sf_district');
  distSel.innerHTML='<option value="">请选择区县</option>';
  if(!c)return;
  fetch('/admin/regions?parent='+c).then(function(r){return r.json()}).then(function(d){
    if(d.code===0){
      d.data.forEach(function(d2){distSel.innerHTML+='<option value="'+d2.code+'">'+d2.name+'</option>'});
    }
  });
}

function showSourceForm(id){
  editingSourceId=id||'';
  var modal=document.getElementById('sourceModal');
  document.getElementById('sourceModalTitle').textContent=id?'编辑数据源':'新增数据源';
  var idInput=document.getElementById('sf_id');
  // 加载省份列表
  fetch('/admin/regions').then(function(r){return r.json()}).then(function(d){
    if(d.code===0){
      var sel=document.getElementById('sf_province');
      sel.innerHTML='<option value="">请选择省份</option>';
      d.data.provinces.forEach(function(p){sel.innerHTML+='<option value="'+p.code+'">'+p.name+'</option>'});
    }
  });
  document.getElementById('sf_city').innerHTML='<option value="">请先选择省份</option>';
  document.getElementById('sf_district').innerHTML='<option value="">请先选择城市</option>';
  if(id){
    idInput.value=id;idInput.readOnly=true;
    fetch('/admin/sources').then(function(r){return r.json()}).then(function(d){
      var s=d.data.find(function(x){return x.source_id===id});
      if(s){
        document.getElementById('sf_name').value=s.source_name;
        document.getElementById('sf_type').value=s.crawl_type;
        document.getElementById('sf_level').value=s.source_level;
        document.getElementById('sf_url').value=s.source_url;
        document.getElementById('sf_interval').value=s.interval_sec;
        document.getElementById('sf_proxy').value=s.proxy_url||'';
        document.getElementById('sf_delay').value=s.request_delay_ms||0;
        document.getElementById('sf_concurrent').value=s.max_concurrent||1;
        document.getElementById('sf_robots').checked=s.respect_robots!==false;
        // 加载并设置地区
        var rc=s.region_code||'';
        if(rc){
          var provCode=rc.slice(0,2)+'0000';
          var cityCode=rc.slice(0,4)+'00';
          var distCode=rc;
          fetch('/admin/regions?parent='+provCode).then(function(r){return r.json()}).then(function(d2){
            if(d2.code===0){
              var citySel=document.getElementById('sf_city');
              citySel.innerHTML='<option value="">请选择城市</option>';
              d2.data.forEach(function(c){citySel.innerHTML+='<option value="'+c.code+'"'+(c.code===cityCode?' selected':'')+'>'+c.name+'</option>'});
              if(cityCode>=provCode){
                fetch('/admin/regions?parent='+cityCode).then(function(r){return r.json()}).then(function(d3){
                  if(d3.code===0){
                    var distSel=document.getElementById('sf_district');
                    distSel.innerHTML='<option value="">请选择区县</option>';
                    d3.data.forEach(function(d){distSel.innerHTML+='<option value="'+d.code+'"'+(d.code===distCode?' selected':'')+'>'+d.name+'</option>'});
                  }
                });
              }
            }
          });
          var provSel=document.getElementById('sf_province');
          setTimeout(function(){
            for(var i=0;i<provSel.options.length;i++){if(provSel.options[i].value===provCode){provSel.selectedIndex=i;break}}
          },100);
        }
        onTypeChange();
      }
    });
  }else{
    idInput.value='';idInput.readOnly=false;
    document.getElementById('sf_name').value='';
    document.getElementById('sf_type').value='govsite';
    document.getElementById('sf_level').value='MEDIUM';
    document.getElementById('sf_url').value='';
    document.getElementById('sf_interval').value='86400';
    document.getElementById('sf_proxy').value='';
    document.getElementById('sf_delay').value='0';
    document.getElementById('sf_concurrent').value='1';
    document.getElementById('sf_robots').checked=true;
    onTypeChange();
  }
  modal.style.display='block';
}

function closeSourceModal(){document.getElementById('sourceModal').style.display='none'}

function saveSource(){
  var id=document.getElementById('sf_id').value.trim();
  var name=document.getElementById('sf_name').value.trim();
  if(!id||!name){showToast('ID和名称不能为空','error');return}
  // 从三级地区选择器组装 region_code
  var prov=document.getElementById('sf_province').value;
  var city=document.getElementById('sf_city').value;
  var dist=document.getElementById('sf_district').value;
  var region_code=dist||city||prov||'';
  var payload={
    source_id:id,source_name:name,
    crawl_type:document.getElementById('sf_type').value,
    source_level:document.getElementById('sf_level').value,
    source_url:document.getElementById('sf_url').value,
    interval_sec:parseInt(document.getElementById('sf_interval').value)||86400,
    region_code:region_code,
    proxy_url:document.getElementById('sf_proxy').value,
    request_delay_ms:parseInt(document.getElementById('sf_delay').value)||0,
    max_concurrent:parseInt(document.getElementById('sf_concurrent').value)||1,
    respect_robots:document.getElementById('sf_robots').checked
  };
  if(editingSourceId){
    payload.source_id=editingSourceId;
    fetch('/admin/sources/update',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
    .then(function(r){return r.json()}).then(function(d){
      if(d.code===0){showToast('已保存','success');closeSourceModal();loadSources()}
      else{showToast('保存失败: '+(d.error||''),'error')}
    }).catch(function(){showToast('请求失败','error')});
  }else{
    fetch('/admin/sources/create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
    .then(function(r){return r.json()}).then(function(d){
      if(d.code===0){showToast('已创建','success');closeSourceModal();loadSources()}
      else{showToast('创建失败: '+(d.error||''),'error')}
    }).catch(function(){showToast('请求失败','error')});
  }
}

function deleteSource(id,name){
  if(!confirm('确定要删除 "'+name+'" 吗？'))return;
  fetch('/admin/sources/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({source_id:id})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('已删除','success');loadSources()}
    else{showToast('删除失败: '+(d.error||''),'error')}
  }).catch(function(){showToast('请求失败','error')});
}

function triggerCrawl(id){
  fetch('/admin/sources/crawl',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({source_id:id})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0)showToast('爬取已触发','success');
    else showToast('触发失败: '+(d.error||''),'error');
  }).catch(function(){showToast('请求失败','error')});
}

function testRSS(url){
  fetch('/admin/sources/test-rss',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url:url})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code!==0){showToast('测试失败: '+(d.error||''),'error');return}
    var items=d.data.items||[];
    var total=d.data.total||0;
    var h='<p style="font-size:13px;color:#6B7280;margin-bottom:8px">共 '+total+' 条，显示前 '+items.length+' 条</p>';
    items.forEach(function(it){
      h+='<div style="padding:6px 0;border-bottom:1px solid #F3F4F6;font-size:13px">'+
        '<div style="font-weight:600">'+esc(it.title||'(无标题)')+'</div>'+
        '<a href="'+esc(it.link)+'" target="_blank" style="font-size:11px;color:#1A56DB;text-decoration:none">'+esc(it.link)+'</a></div>';
    });
    document.getElementById('rssPreviewContent').innerHTML=h;
    document.getElementById('rssPreview').style.display='block';
  }).catch(function(){showToast('请求失败','error')});
}

function closeRSSPreview(){document.getElementById('rssPreview').style.display='none'}

function showImportModal(sid,sname){
  document.getElementById('imp_source_id').value=sid;
  document.getElementById('importModalTitle').textContent='导入内容 — '+sname;
  document.getElementById('imp_title').value='';
  document.getElementById('imp_url').value='';
  document.getElementById('imp_content').value='';
  document.getElementById('importModal').style.display='block';
}

function closeImportModal(){document.getElementById('importModal').style.display='none'}

function doSourceImport(){
  var sid=document.getElementById('imp_source_id').value;
  var title=document.getElementById('imp_title').value.trim();
  var content=document.getElementById('imp_content').value.trim();
  if(!title||!content){showToast('标题和内容不能为空','error');return}
  fetch('/admin/sources/import',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({source_id:sid,title:title,content:content,source_url:document.getElementById('imp_url').value})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('导入成功，待审核','success');closeImportModal()}
    else{showToast('导入失败: '+(d.error||''),'error')}
  }).catch(function(){showToast('请求失败','error')});
}

var claimFilter='';
var claimRegion='';
var claimSource='';
var claimType='';
var claimLevel='';

function loadClaims(filter,region,source,ptype,plevel){
  if(filter!==undefined)claimFilter=filter;
  if(region!==undefined)claimRegion=region;
  if(source!==undefined)claimSource=source;
  if(ptype!==undefined)claimType=ptype;
  if(plevel!==undefined)claimLevel=plevel;
  var params=[];
  if(claimFilter)params.push('status='+claimFilter);
  if(claimRegion)params.push('region_code='+claimRegion);
  if(claimSource)params.push('source_id='+encodeURIComponent(claimSource));
  if(claimType)params.push('policy_type='+claimType);
  if(claimLevel)params.push('source_level='+claimLevel);
  var url='/admin/claims'+(params.length?'?'+params.join('&'):'');
  Promise.all([
    fetch(url).then(function(r){return r.json()}),
    fetch('/admin/sources').then(function(r){return r.json()})
  ]).then(function(results){
    var d=results[0],srcD=results[1];
    var claims=d.data||d.claims||[];
    var sources=(srcD.data||[]).sort(function(a,b){return (a.source_name||a.source_id).localeCompare(b.source_name||b.source_id)});
    var regionMap={'':'全部地区','110000':'北京','310000':'上海','330100':'杭州','440100':'广州','440300':'深圳'};
    var typeMap={'':'全部类型','subsidy':'补贴','pension':'养老','medical':'医疗','unemployment':'失业','injury':'工伤','maternity':'生育','housing_fund':'公积金','training':'培训'};
    var levelMap={'':'全部级别','HIGH':'HIGH','MEDIUM':'MEDIUM','LOW':'LOW'};
    var h='<div style="display:flex;gap:6px;margin-bottom:12px;flex-wrap:wrap;align-items:center">'+
      '<button class="btn '+(claimFilter===''?'btn-primary':'btn-outline')+'" onclick="loadClaims(\'\')">全部 ('+claims.length+')</button>'+
      '<button class="btn '+(claimFilter==='pending_review'?'btn-primary':'btn-outline')+'" onclick="loadClaims(\'pending_review\')">待审核</button>'+
      '<button class="btn '+(claimFilter==='verified'?'btn-primary':'btn-outline')+'" onclick="loadClaims(\'verified\')">已通过</button>'+
      '<button class="btn '+(claimFilter==='unverified'?'btn-primary':'btn-outline')+'" onclick="loadClaims(\'unverified\')">已驳回</button>'+
      '<span style="width:1px;height:24px;background:#E5E7EB;margin:0 4px"></span>'+
      '<span style="font-size:12px;color:#6B7280">类型:</span> '+
      '<select onchange="loadClaims(undefined,undefined,undefined,this.value)" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">';
    Object.keys(typeMap).forEach(function(k){
      h+='<option value="'+k+'" '+(claimType===k?'selected':'')+'>'+typeMap[k]+'</option>';
    });
    h+='</select>'+
      '<span style="font-size:12px;color:#6B7280">级别:</span> '+
      '<select onchange="loadClaims(undefined,undefined,undefined,undefined,this.value)" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">';
    Object.keys(levelMap).forEach(function(k){
      h+='<option value="'+k+'" '+(claimLevel===k?'selected':'')+'>'+levelMap[k]+'</option>';
    });
    h+='</select>'+
      '<span style="font-size:12px;color:#6B7280">地区:</span> '+
      '<select onchange="loadClaims(undefined,this.value)" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">';
    Object.keys(regionMap).forEach(function(k){
      h+='<option value="'+k+'" '+(claimRegion===k?'selected':'')+'>'+regionMap[k]+'</option>';
    });
    h+='</select>'+
      '<span style="font-size:12px;color:#6B7280">数据源:</span> '+
      '<select onchange="loadClaims(undefined,undefined,this.value)" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px;max-width:200px">'+
      '<option value="">全部数据源</option>';
    sources.forEach(function(s){
      var label=esc(s.source_name||s.source_id)+(s.crawl_type?' ('+s.crawl_type+')':'');
      h+='<option value="'+esc(s.source_id)+'" '+(claimSource===s.source_id?'selected':'')+'>'+label+'</option>';
    });
    h+='</select></div>';
    claims.forEach(function(c){
      var badge=c.status==='verified'?'<span class="badge bg-green">已通过</span>':c.status==='unverified'?'<span class="badge bg-red">已驳回</span>':'<span class="badge bg-yellow">待审核</span>';
      var actions='';
      if(c.status!=='verified')actions+='<button class="btn btn-success btn-sm" onclick="updateClaim(\''+esc(c.claim_id)+'\',\'verified\')">通过</button>';
      if(c.status!=='unverified')actions+='<button class="btn btn-danger btn-sm" onclick="updateClaim(\''+esc(c.claim_id)+'\',\'unverified\')">驳回</button>';
      var src='';
      if(c.source_name)src+='<span style="font-size:11px;color:#6B7280">来源: '+esc(c.source_name)+'</span>';
      var policyLink=c.policy_url||c.source_url;
      if(policyLink)src+=' <a href="'+esc(policyLink)+'" target="_blank" style="font-size:11px;color:#1A56DB;text-decoration:none" title="'+esc(policyLink)+'">[政策原文]</a>';
      h+='<div class="card"><div style="display:flex;justify-content:space-between;margin-bottom:6px"><span style="font-weight:600;color:#1A56DB;font-size:13px">'+esc(c.claim_id)+'</span>'+badge+'</div>'+
      '<div style="font-size:12px;color:#6B7280;margin-bottom:4px">ID: '+esc(c.policy_id)+' | 类型: '+esc(c.policy_type)+' | 地区: '+esc(c.region_code)+' | 置信度: '+Math.round((c.confidence_score||0)*100)+'%</div>'+
      '<div style="font-size:12px">'+esc(c.subsidy_calc_method)+'</div>'+
      (src?'<div style="font-size:12px;margin-top:4px">'+src+'</div>':'')+
      (actions?'<div style="margin-top:8px;display:flex;gap:6px">'+actions+'</div>':'')+'</div>'
    });
    if(claims.length===0)h+='<div class="card" style="text-align:center;color:#9CA3AF;padding:40px">暂无政策数据</div>';
    document.getElementById('app').innerHTML=h;
  }).catch(function(e){document.getElementById('app').innerHTML='<div class="card" style="text-align:center;color:#9CA3AF;padding:40px">加载失败</div>'});
}

function updateClaim(id,status){
  fetch('/admin/claims/'+id,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:status,confidence_score:status==='verified'?1.0:0})})
  .then(function(r){return r.json()}).then(function(d){showToast(status==='verified'?'\u5df2\u901a\u8fc7':'\u5df2\u9a73\u56de','success');loadClaims()})
  .catch(function(){showToast('\u64cd\u4f5c\u5931\u8d25','error')});
}

var extractProviderMap={
  'deepseek':{name:'DeepSeek',endpoint:'https://api.deepseek.com/v1/chat/completions',model:'deepseek-chat'},
  'ali_bailian':{name:'阿里云百炼',endpoint:'https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation',model:'qwen-plus'},
  'volc_ark':{name:'火山方舟',endpoint:'https://ark.cn-beijing.volces.com/api/v3/chat/completions',model:'doubao-pro-32k'},
  'opencode_go':{name:'OpenCode Go',endpoint:'http://localhost:11434/v1/chat/completions',model:'opencode-go'}
};

var embeddingProviderMap={
  'volc_ark':{name:'火山方舟 Doubao',endpoint:'https://ark.cn-beijing.volces.com/api/v3/embeddings/multimodal',dims:1024},
  'openai':{name:'OpenAI',endpoint:'https://api.openai.com/v1/embeddings',dims:1536}
};

function loadExtract(){
  var app=document.getElementById('app');app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>加载中...</div>';
  Promise.all([
    Promise.resolve({code:0,data:{provider:'',model_name:'',endpoint:'',api_key:'',enabled:false,embedding_model:'',embedding_dimensions:0,embedding_api_key:'',embedding_endpoint:'',backup_provider:'',backup_api_key:'',backup_endpoint:'',backup_model_name:''}}),
    fetch('/admin/llm/status').then(function(r){return r.json()}),
    fetch('/admin/llm/pending').then(function(r){return r.json()})
  ]).then(function(results){
    var cfg=results[0].data,st=results[1].data,pending=results[2].data||[];
    var prov=extractProviderMap[cfg.provider]||extractProviderMap['deepseek'];
    var h='<div class="card" style="background:#EFF6FF;border:1px solid #BFDBFE;padding:12px 16px;margin-bottom:12px;border-radius:8px"><div style="display:flex;align-items:center;gap:8px"><span style="font-size:16px">&#9881;</span><div><div style="font-weight:600;font-size:14px;color:#1A56DB">模型配置已迁移至 LLM Gateway</div><div style="font-size:12px;color:#6B7280;margin-top:2px">LLM、Embedding、ASR 统一在 Gateway 管理</div></div><a href="http://localhost:39404/admin/#model-configs" target="_blank" style="margin-left:auto;padding:6px 14px;background:#1A56DB;color:#fff;border-radius:6px;text-decoration:none;font-size:12px">前往配置 →</a></div></div>';
    h+='<div class="card"><h3 style="font-size:16px;margin-bottom:12px">AI 提取状态</h3>';
    h+='<div style="padding:8px 0;color:#6B7280;font-size:13px">模型配置请前往 LLM Gateway 管理</div>';
    h+='</div>';

    h+='<div class="card" style="margin-top:12px;background:#EFF6FF;border:1px solid #BFDBFE;padding:12px 16px;border-radius:8px"><div style="display:flex;align-items:center;gap:8px"><span style="font-size:16px">&#9881;</span><div><div style="font-weight:600;font-size:14px;color:#1A56DB">ASR 配置已迁移至 LLM Gateway</div><div style="font-size:12px;color:#6B7280;margin-top:2px">在 Gateway 管理 ASR 模型配置</div></div><a href="http://localhost:39404/admin/#model-configs" target="_blank" style="margin-left:auto;padding:6px 14px;background:#1A56DB;color:#fff;border-radius:6px;text-decoration:none;font-size:12px">前往配置 →</a></div></div>';
    h+='<div class="card" style="margin-top:12px"><h3 style="font-size:16px;margin-bottom:12px">ASR 语音识别配置（火山引擎大模型）</h3>';
    h+='<div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">';
    h+='<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">服务版本</label>'+
      '<select id="asrResourceId" onchange="updateASRDefaults()" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      '<option value="volc.bigasr.auc">标准版（推荐）</option><option value="volc.bigasr.auc_idle">闲时版（低成本）</option></select></div>';
    h+='<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">APP ID</label><input id="asrAppId" placeholder="火山引擎 APP ID" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>';
    h+='<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Access Token</label><input id="asrKey" type="password" placeholder="Access Token" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>';
    h+='<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Language</label><input id="asrLang" value="zh" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>';
    h+='<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">最大等待时间（秒）</label><input id="asrMaxWait" type="number" value="300" min="30" max="3600" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>';
    h+='<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">轮询间隔（秒）</label><input id="asrPollInterval" type="number" value="5" min="2" max="30" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>';
    h+='<div style="grid-column:1/3"><label style="font-size:13px;display:flex;align-items:center;gap:4px;cursor:pointer"><input id="asrEnabled" type="checkbox">启用 ASR</label></div>';
    h+='<div style="text-align:right;grid-column:1/3;display:flex;gap:8px;justify-content:flex-end">'+
      '<button class="btn" style="background:#6B7280;color:#fff;border:none" onclick="testASR()">测试连接</button>'+
      '<button class="btn btn-primary" onclick="saveASRConfig()">保存 ASR 配置</button></div>';
    h+='</div>';
    h+='<div id="asrTestResult" style="margin-top:8px;display:none"></div>';
    h+='</div>';

    h+='<div class="card"><h3 style="font-size:16px;margin-bottom:12px">提取状态</h3>';
    h+='<div style="display:flex;gap:12px;flex-wrap:wrap">';
    h+='<div style="background:#F3F4F6;border-radius:8px;padding:12px 20px;text-align:center"><span style="font-size:24px;font-weight:700;color:#1A56DB;display:block">'+st.unprocessed+'</span><span style="font-size:12px;color:#6B7280">待提取原始文本</span></div>';
    h+='<div style="background:#F3F4F6;border-radius:8px;padding:12px 20px;text-align:center"><span style="font-size:24px;font-weight:700;color:'+(st.llm_configured?'#059669':'#D97706')+';display:block">'+(st.llm_configured?'已配置':'未配置')+'</span><span style="font-size:12px;color:#6B7280">LLM 状态</span></div>';
    h+='</div>';
    if(st.llm_configured){
      h+='<div id="extractProgressArea" style="margin-top:12px">'+
        '<div class="progress-bar"><div class="progress-fill" id="progressBar"></div></div>'+
        '<div style="display:flex;justify-content:space-between;align-items:center;gap:8px;flex-wrap:wrap">'+
        '<button class="btn btn-success" onclick="runExtraction()" id="extractBtn">开始批量提取</button>'+
        '<span id="extractStatus" style="font-size:13px;color:#6B7280">'+st.unprocessed+' 条待提取</span></div></div>';
    }else{
      h+='<div style="margin-top:12px;font-size:13px;color:#D97706">请先配置 API Key 并启用 LLM</div>';
    }
    h+='</div>';

    // 待提取列表
    if(pending.length>0){
      h+='<div class="card"><div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">'+
        '<h3 style="font-size:15px">待提取列表 ('+pending.length+' 条)</h3>'+
        '<div style="display:flex;gap:6px;align-items:center">'+
        '<span style="font-size:12px;color:#6B7280">类型:</span>'+
        '<select id="ext_ft" onchange="filterExtract()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
        '<option value="">全部</option>'+sourceTypeList.map(function(t){return '<option value="'+t.v+'">'+t.l+'</option>'}).join('')+'</select>'+
        '<span style="font-size:12px;color:#6B7280">级别:</span>'+
        '<select id="ext_fl" onchange="filterExtract()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
        '<option value="">全部</option><option value="HIGH">HIGH</option><option value="MEDIUM">MEDIUM</option><option value="LOW">LOW</option></select>'+
        '<span style="font-size:12px;color:#6B7280">地区:</span>'+
        '<select id="ext_fr" onchange="filterExtract()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
        '<option value="">全部</option>';
      var extRegions=[];pending.forEach(function(p){if(p.region_code && extRegions.indexOf(p.region_code)<0)extRegions.push(p.region_code)});
      extRegions.sort().forEach(function(r){h+='<option value="'+r+'">'+r+'</option>'});
      h+='</select></div></div>'+
        '<div id="extListWrap"><div style="overflow-x:auto"><table><tr><th>来源</th><th>标题</th><th>URL</th><th>抓取时间</th></tr>';
      pending.forEach(function(p){
        h+='<tr data-ct="'+esc(p.crawl_type)+'" data-sl="'+esc(p.source_level)+'" data-rc="'+esc(p.region_code)+'"><td style="font-weight:600;font-size:13px">'+esc(p.source_name||p.source_id)+'</td>'+
          '<td>'+esc(p.title||p.source_url)+'</td>'+
          '<td style="font-size:11px;max-width:200px;overflow:hidden;text-overflow:ellipsis">'+
          '<a href="'+esc(p.source_url)+'" target="_blank" style="color:#1A56DB;text-decoration:none">[链接]</a></td>'+
          '<td style="font-size:11px;color:#9CA3AF">'+esc(p.fetched_at)+'</td></tr>';
      });
      h+='</table></div></div></div>';
    }else if(st.unprocessed>0){
      h+='<div class="card" style="padding:12px;text-align:center;color:#9CA3AF;font-size:13px">共有 '+st.unprocessed+' 条待提取，详细信息请启动提取后查看进度</div>';
    }

    app.innerHTML=h;
    loadASRConfig();
  }).catch(function(e){app.innerHTML='<div class="card"><h3>加载失败</h3><p>'+esc(e.message)+'</p></div>'});
}

function updateLLMEndpoint(){
  var p=document.getElementById('llmProvider').value;
  var info=extractProviderMap[p]||extractProviderMap['deepseek'];
  document.getElementById('llmEndpoint').value=info.endpoint;
  document.getElementById('llmModel').value=info.model;
}

function updateEmbeddingEndpoint(){
  var p=document.getElementById('embProvider').value;
  var info=embeddingProviderMap[p]||embeddingProviderMap['volc_ark'];
  document.getElementById('embEndpoint').value=info.endpoint;
  document.getElementById('embDims').value=info.dims;
}

function filterExtract(){
  var ft=document.getElementById('ext_ft').value;
  var fl=document.getElementById('ext_fl').value;
  var fr=document.getElementById('ext_fr').value;
  var rows=document.querySelectorAll('#extListWrap tr[data-ct]');
  var vis=0;
  rows.forEach(function(r){
    var show=true;
    if(ft && r.dataset.ct!==ft)show=false;
    if(fl && r.dataset.sl!==fl)show=false;
    if(fr && r.dataset.rc!==fr)show=false;
    r.style.display=show?'':'none';
    if(show)vis++;
  });
}

function saveLLMConfig(){
  var p=document.getElementById('llmProvider').value;
  fetch('/admin/llm/config/save',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({
      provider:p,
      api_key:document.getElementById('llmKey').value,
      endpoint:document.getElementById('llmEndpoint').value,
      model_name:document.getElementById('llmModel').value,
      max_tokens:4096,
      enabled:!!document.getElementById('llmKey').value,
      embedding_model:document.getElementById('embModel').value,
      embedding_dimensions:parseInt(document.getElementById('embDims').value)||1024,
      embedding_api_key:document.getElementById('embKey').value,
      embedding_endpoint:document.getElementById('embEndpoint').value,
      backup_provider:document.getElementById('backupProvider').value,
      backup_api_key:document.getElementById('backupKey').value,
      backup_endpoint:document.getElementById('backupEndpoint').value,
      backup_model_name:document.getElementById('backupModel').value
    })}).then(function(r){return r.json()}).then(function(d){
    showToast('\u914d\u7f6e\u5df2\u4fdd\u5b58','success')
  }).catch(function(){showToast('\u4fdd\u5b58\u5931\u8d25','error')});
}

var extractTimer=null;

function runExtraction(){
  var btn=document.getElementById('extractBtn');
  btn.disabled=true;
  btn.textContent='启动中...';
  fetch('/admin/llm/extract',{method:'POST'}).then(function(r){return r.json()}).then(function(d){
    if(d.code!==0){showToast('启动失败: '+d.message,'error');btn.disabled=false;btn.textContent='开始批量提取';return}
    btn.textContent='提取中...';
    pollExtractProgress();
  }).catch(function(){showToast('请求失败','error');btn.disabled=false;btn.textContent='开始批量提取'});
}

function pollExtractProgress(){
  fetch('/admin/llm/progress').then(function(r){return r.json()}).then(function(d){
    var p=d.data;
    var statusEl=document.getElementById('extractStatus');
    var barEl=document.getElementById('progressBar');
    if(!statusEl||!barEl)return;
    var pct=p.total>0?Math.round((p.completed+p.failed)/p.total*100):0;
    barEl.style.width=pct+'%';
    statusEl.innerHTML=p.completed+'/'+p.total+' 完成, '+p.failed+' 失败'+(p.current_src?' (当前: '+esc(p.current_src)+')':'');
    if(p.done){
      showToast('提取完成: '+p.completed+' 成功, '+p.failed+' 失败','success');
      var btn=document.getElementById('extractBtn');
      if(btn){btn.disabled=false;btn.textContent='开始批量提取'}
      if(extractTimer){clearInterval(extractTimer);extractTimer=null}
      loadExtract();
      return;
    }
    if(p.running){
      extractTimer=setTimeout(pollExtractProgress,2000);
    }
  }).catch(function(){
    extractTimer=setTimeout(pollExtractProgress,2000);
  });
}

function loadSearch(){
  var app=document.getElementById('app');
  app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>\u52a0\u8f7d\u4e2d...</div>'+
    '<iframe src="/admin/search_page" style="width:100%;height:800px;border:none;" onload="var s=this.previousSibling;if(s)s.style.display=\'none\'"></iframe>';
}

var logStartDate='',logEndDate='',logFt='',logFl='',logSt='';

function loadLogs(){
  var app=document.getElementById('app');
  var today=new Date();var weekAgo=new Date(today);weekAgo.setDate(weekAgo.getDate()-7);
  var sd=logStartDate||weekAgo.toISOString().slice(0,10);
  var ed=logEndDate||today.toISOString().slice(0,10);
  var ft=logFt,vl=logFl,st=logSt;
  var url='/admin/logs?start_date='+sd+'&end_date='+ed+'&source_type='+ft+'&source_level='+vl+'&status='+st;
  fetch(url).then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    var items=d.data||[];
    var h='<div style="display:flex;gap:6px;align-items:center;margin-bottom:10px;flex-wrap:wrap">'+
      '<span style="font-size:12px;color:#6B7280">起始</span>'+
      '<input type="date" id="logStart" value="'+sd+'" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<span style="font-size:12px;color:#6B7280">结束</span>'+
      '<input type="date" id="logEnd" value="'+ed+'" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<span style="font-size:12px;color:#6B7280">类型:</span>'+
      '<select id="logFt" onchange="filterLogs()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<option value="">全部</option>'+sourceTypeList.map(function(t){return '<option value="'+t.v+'"'+(ft===t.v?' selected':'')+'>'+t.l+'</option>'}).join('')+'</select>'+
      '<span style="font-size:12px;color:#6B7280">级别:</span>'+
      '<select id="logFl" onchange="filterLogs()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<option value="">全部</option><option value="HIGH"'+(vl==='HIGH'?' selected':'')+'>HIGH</option><option value="MEDIUM"'+(vl==='MEDIUM'?' selected':'')+'>MEDIUM</option><option value="LOW"'+(vl==='LOW'?' selected':'')+'>LOW</option></select>'+
      '<span style="font-size:12px;color:#6B7280">状态:</span>'+
      '<select id="logSt" onchange="filterLogs()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<option value="">全部</option><option value="true"'+(st==='true'?' selected':'')+'>成功</option><option value="false"'+(st==='false'?' selected':'')+'>失败</option></select>'+
      '<button class="btn btn-primary btn-sm" onclick="filterLogs()">查询</button>'+
      '<span style="font-size:13px;color:#6B7280">共 '+items.length+' 条记录</span></div>'+
      '<div style="overflow-x:auto"><table><tr><th>\u65f6\u95f4</th><th>\u6570\u636e\u6e90</th><th>\u72b6\u6001</th><th>\u63d0\u53d6ID</th><th>\u5185\u5bb9\u6982\u8981</th><th>\u9519\u8bef\u4fe1\u606f</th></tr>';
    items.forEach(function(l){
      var claimCell=l.extracted_claim_id?'<a href="#claims" onclick="loadClaims(undefined,undefined,undefined,undefined);return false" style="color:#1A56DB;text-decoration:none;font-size:12px">'+esc(l.extracted_claim_id)+'</a>':'-';
      var summaryCell=l.content_summary?'<span style="font-size:12px;color:#374151">'+esc(l.content_summary)+'</span>':'-';
      h+='<tr><td style="font-size:11px;color:#9CA3AF">'+esc(l.crawled_at)+'</td><td>'+esc(l.source_name||l.source_id)+'</td>'+
      '<td>'+(l.status==='true'||l.status===true?'<span class="badge bg-green">\u6210\u529f</span>':'<span class="badge bg-red">\u5931\u8d25</span>')+'</td>'+
      '<td>'+claimCell+'</td>'+
      '<td style="max-width:250px;overflow:hidden;text-overflow:ellipsis">'+summaryCell+'</td>'+
      '<td style="font-size:12px;color:#EF4444;max-width:200px;overflow:hidden;text-overflow:ellipsis" title="'+esc(l.error_message)+'">'+esc(l.error_message||'-')+'</td></tr>'
    });
    h+='</table></div>';app.innerHTML=h;
  }).catch(function(e){app.innerHTML='<div class="card"><h3>\u52a0\u8f7d\u5931\u8d25</h3><p>'+esc(e.message)+'</p></div>'});
}

function filterLogs(){
  logStartDate=document.getElementById('logStart').value;
  logEndDate=document.getElementById('logEnd').value;
  logFt=document.getElementById('logFt').value;
  logFl=document.getElementById('logFl').value;
  logSt=document.getElementById('logSt').value;
  loadLogs();
}

var extLogStartDate='',extLogEndDate='',extLogFt='',extLogFl='',extLogFr='',extLogSt='';

function loadExtractLogs(){
  var app=document.getElementById('app');
  var today=new Date();var weekAgo=new Date(today);weekAgo.setDate(weekAgo.getDate()-7);
  var sd=extLogStartDate||weekAgo.toISOString().slice(0,10);
  var ed=extLogEndDate||today.toISOString().slice(0,10);
  var ft=extLogFt,vl=extLogFl,rc=extLogFr,st=extLogSt;
  var url='/admin/extract-logs?start_date='+sd+'&end_date='+ed+'&source_type='+ft+'&source_level='+vl+'&region_code='+rc+'&status='+st;
  fetch(url).then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    var items=d.data||[];
    var succ=0,fail=0;items.forEach(function(l){if(l.status==='success')succ++;else fail++});
    // 生成筛选栏
    var h='<div style="display:flex;gap:6px;align-items:center;margin-bottom:10px;flex-wrap:wrap">'+
      '<span style="font-size:12px;color:#6B7280">起始</span>'+
      '<input type="date" id="extLogStart" value="'+sd+'" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<span style="font-size:12px;color:#6B7280">结束</span>'+
      '<input type="date" id="extLogEnd" value="'+ed+'" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<span style="font-size:12px;color:#6B7280">类型:</span>'+
      '<select id="extLogFt" onchange="filterExtractLogs()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<option value="">全部</option>'+sourceTypeList.map(function(t){return '<option value="'+t.v+'"'+(ft===t.v?' selected':'')+'>'+t.l+'</option>'}).join('')+'</select>'+
      '<span style="font-size:12px;color:#6B7280">级别:</span>'+
      '<select id="extLogFl" onchange="filterExtractLogs()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<option value="">全部</option><option value="HIGH"'+(vl==='HIGH'?' selected':'')+'>HIGH</option><option value="MEDIUM"'+(vl==='MEDIUM'?' selected':'')+'>MEDIUM</option><option value="LOW"'+(vl==='LOW'?' selected':'')+'>LOW</option></select>'+
      '<span style="font-size:12px;color:#6B7280">状态:</span>'+
      '<select id="extLogSt" onchange="filterExtractLogs()" style="padding:3px 6px;border:1px solid #D1D5DB;border-radius:4px;font-size:12px">'+
      '<option value="">全部</option><option value="success"'+(st==='success'?' selected':'')+'>成功</option><option value="failed"'+(st==='failed'?' selected':'')+'>失败</option></select>'+
      '<button class="btn btn-primary btn-sm" onclick="filterExtractLogs()">查询</button>'+
      '<span class="badge bg-green">成功 '+succ+'</span>'+
      '<span class="badge bg-red">失败 '+fail+'</span>'+
      '<span style="font-size:13px;color:#6B7280">共 '+items.length+' 条</span></div>'+
      '<div style="overflow-x:auto"><table><tr><th>\u65f6\u95f4</th><th>\u6570\u636e\u6e90</th><th>\u6a21\u578b</th><th>\u6807\u9898</th><th>\u63d0\u53d6\u5185\u5bb9\u6982\u8981</th><th>\u63d0\u53d6ID</th><th>\u72b6\u6001</th><th>\u4fe1\u606f</th></tr>';
    items.forEach(function(l){
      var claimCell=l.claim_id?'<span style="font-size:11px;color:#1A56DB">'+esc(l.claim_id)+'</span>':'-';
      var titleCell=l.title?'<span style="font-size:12px;color:#374151;max-width:150px;display:inline-block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">'+esc(l.title)+'</span>':'-';
      var summaryCell=l.content_summary?'<span style="font-size:12px;color:#374151">'+esc(l.content_summary)+'</span>':'-';
      h+='<tr><td style="font-size:11px;color:#9CA3AF">'+esc(l.created_at)+'</td>'+
      '<td style="font-size:12px">'+esc(l.source_name||l.source_id)+'</td>'+
      '<td style="font-size:11px;color:#6B7280">'+esc(l.model_name||'-')+'</td>'+
      '<td>'+titleCell+'</td>'+
      '<td style="max-width:200px;overflow:hidden;text-overflow:ellipsis">'+summaryCell+'</td>'+
      '<td>'+claimCell+'</td>'+
      '<td>'+(l.status==='success'?'<span class="badge bg-green">\u6210\u529f</span>':'<span class="badge bg-red">\u5931\u8d25</span>')+'</td>'+
      '<td style="font-size:12px;color:'+(l.status==='success'?'#059669':'#EF4444')+';max-width:250px;overflow:hidden;text-overflow:ellipsis" title="'+esc(l.error_message)+'">'+esc(l.error_message||'-')+'</td></tr>'
    });
    h+='</table></div>';app.innerHTML=h;
  }).catch(function(e){app.innerHTML='<div class="card"><h3>\u52a0\u8f7d\u5931\u8d25</h3><p>'+esc(e.message)+'</p></div>'});
}

function filterExtractLogs(){
  extLogStartDate=document.getElementById('extLogStart').value;
  extLogEndDate=document.getElementById('extLogEnd').value;
  extLogFt=document.getElementById('extLogFt').value;
  extLogFl=document.getElementById('extLogFl').value;
  extLogSt=document.getElementById('extLogSt').value;
  loadExtractLogs();
}

function loadPipeline(){
  var app=document.getElementById('app');
  app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>\u52a0\u8f7d\u4e2d...</div>';
  fetch('/admin/pipeline').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    var items=d.data||[];
    var succ=0,fail=0,total=0;
    items.forEach(function(e){
      if(e.enabled)total++;
      if(e.crawl_status==='success')succ++;
      if(e.crawl_status==='failed')fail++;
    });
    var h='<div class="stat-row" style="margin-bottom:12px">'+
      '<div class="stat-card"><span class="stat-num c-blue">'+items.length+'</span><span class="stat-label">数据源</span></div>'+
      '<div class="stat-card"><span class="stat-num c-blue">'+total+'</span><span class="stat-label">启用的</span></div>'+
      '<div class="stat-card"><span class="stat-num c-green">'+succ+'</span><span class="stat-label">爬取成功</span></div>'+
      '<div class="stat-card"><span class="stat-num c-red">'+fail+'</span><span class="stat-label">爬取失败</span></div>'+
      '</div>';
    items.forEach(function(e){
      var crawlBadge=e.crawl_status==='success'?'<span class="badge bg-green">成功</span>':
        e.crawl_status==='failed'?'<span class="badge bg-red">失败</span>':
        e.crawl_status==='never'?'<span class="badge bg-yellow">从未</span>':'<span class="badge bg-yellow">'+esc(e.crawl_status)+'</span>';
      var extractBadge=e.extract_status==='success'?'<span class="badge bg-green">成功</span>':
        e.extract_status==='failed'?'<span class="badge bg-red">失败</span>':'<span class="badge bg-yellow">未处理</span>';
      var claimBadge=e.claim_status==='verified'?'<span class="badge bg-green">已通过</span>':
        e.claim_status==='unverified'?'<span class="badge bg-red">已驳回</span>':
        e.claim_status==='pending_review'?'<span class="badge bg-yellow">待审核</span>':'<span class="badge bg-yellow">无</span>';
      var rawIcon=e.has_raw_text?'<span style="color:#059669">\u2713</span>':'<span style="color:#EF4444">\u2717</span>';
      h+='<div class="card" style="margin-bottom:8px;padding:12px 16px">'+
        '<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:6px">'+
        '<span style="font-weight:600;font-size:14px">'+esc(e.source_name||e.source_id)+' <span style="font-size:11px;color:#6B7280;font-weight:normal">('+esc(e.crawl_type)+', '+esc(e.source_level)+')</span></span>'+
        '<span>'+(e.enabled?'<span class="badge bg-green">启用</span>':'<span class="badge bg-red">禁用</span>')+'</span></div>'+
        '<div style="display:flex;gap:16px;font-size:12px;color:#6B7280;flex-wrap:wrap">'+
        '<div><span style="color:#374151;font-weight:500">爬取:</span> '+crawlBadge+' <span style="color:#9CA3AF">'+esc(e.last_crawl_at||'')+'</span><br>'+
        '<span style="font-size:11px;color:#EF4444">'+esc(e.crawl_error)+'</span></div>'+
        '<div><span style="color:#374151;font-weight:500">原文:</span> '+rawIcon+'</div>'+
        '<div><span style="color:#374151;font-weight:500">提取:</span> '+extractBadge+' <span style="color:#9CA3AF">'+esc(e.last_extract_at||'')+'</span></div>'+
        '<div><span style="color:#374151;font-weight:500">政策:</span> '+claimBadge+' '+
        (e.claim_id?'<span style="font-size:11px;color:#1A56DB">'+esc(e.claim_id)+'</span>':'')+
        (e.confidence>0?' <span style="font-size:11px;color:#D97706">'+(Math.round(e.confidence*100))+'%</span>':'')+
        '</div></div></div>';
    });
    app.innerHTML=h;
  }).catch(function(e){app.innerHTML='<div class="card"><h3>\u52a0\u8f7d\u5931\u8d25</h3><p>'+esc(e.message)+'</p></div>'});
}

function showImportForm(){
  document.getElementById('app').innerHTML=
  '<div class="card"><h3 style="font-size:16px;margin-bottom:8px">\u5bfc\u5165\u653f\u7b56</h3><p style="font-size:13px;color:#6B7280;margin-bottom:12px">\u7c98\u8d34\u7ed3\u6784\u5316\u653f\u7b56\u6587\u672c\u3002\u4e5f\u53ef\u5c06 .txt \u6587\u4ef6\u653e\u5165 /data/policies/ \u76ee\u5f55\u81ea\u52a8\u5bfc\u5165\u3002</p>'+
  '<textarea id="policyText" placeholder="\u653f\u7b56ID: SH-2025-FLEX-SUBSIDY\n\u5730\u533a\u4ee3\u7801: 310000\n\u653f\u7b56\u7c7b\u578b: subsidy\n\u9002\u7528\u4eba\u7fa4: flexible_employment, 4050\n\u8ba1\u7b97\u65b9\u6cd5: \u57fa\u6570*50%\n\u751f\u6548\u65e5\u671f: 2025-01-01"></textarea>'+
  '<div style="display:flex;gap:8px;margin-top:12px"><button class="btn btn-success" onclick="doImport()" id="importBtn">\u89e3\u6790\u5e76\u5bfc\u5165</button>'+
  '<button class="btn btn-outline" onclick="fillSample()">\u586b\u5165\u793a\u4f8b</button></div></div>';
}

function fillSample(){
  document.getElementById('policyText').value=[
    '\u653f\u7b56ID: SH-2025-FLEX-SUBSIDY-DEMO',
    '\u5730\u533a\u4ee3\u7801: 310000',
    '\u653f\u7b56\u7c7b\u578b: subsidy',
    '\u9002\u7528\u4eba\u7fa4: flexible_employment, 4050',
    '\u8ba1\u7b97\u65b9\u6cd5: \u57fa\u6570*50%',
    '\u751f\u6548\u65e5\u671f: 2025-01-01',
  ].join('\n');
}

function doImport(){
  var text=document.getElementById('policyText').value;
  if(!text.trim()){showToast('\u8bf7\u8f93\u5165\u653f\u7b56\u6587\u672c','error');return}
  var btn=document.getElementById('importBtn');btn.disabled=true;btn.textContent='\u89e3\u6790\u4e2d...';
  fetch('/admin/ingest',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({text:text})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('\u5bfc\u5165\u6210\u529f\uff0c\u5f85\u5ba1\u6838','success');document.getElementById('policyText').value=''}
    else{showToast('\u5bfc\u5165\u5931\u8d25: '+(d.error||d.message||''),'error')}
    btn.disabled=false;btn.textContent='\u89e3\u6790\u5e76\u5bfc\u5165';
  }).catch(function(){showToast('\u8bf7\u6c42\u5931\u8d25','error');btn.disabled=false;btn.textContent='\u89e3\u6790\u5e76\u5bfc\u5165'});
}

function loadASRConfig(){
  fetch('/admin/asr/config').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0||!d.data)return;
    var c=d.data;
    var ridEl=document.getElementById('asrResourceId');
    var aidEl=document.getElementById('asrAppId');
    var kEl=document.getElementById('asrKey');
    var lEl=document.getElementById('asrLang');
    var mwEl=document.getElementById('asrMaxWait');
    var piEl=document.getElementById('asrPollInterval');
    var enEl=document.getElementById('asrEnabled');
    if(ridEl&&c.resource_id)ridEl.value=c.resource_id;
    if(aidEl&&c.app_id)aidEl.value=c.app_id;
    if(kEl&&c.api_key)kEl.value=c.api_key;
    if(lEl&&c.language)lEl.value=c.language;
    if(mwEl&&c.max_wait_seconds)mwEl.value=c.max_wait_seconds;
    if(piEl&&c.poll_interval_seconds)piEl.value=c.poll_interval_seconds;
    if(enEl&&c.enabled!==undefined)enEl.checked=!!c.enabled;
  }).catch(function(){});
}

function updateASRDefaults(){
  var rid=document.getElementById('asrResourceId').value;
  var mwEl=document.getElementById('asrMaxWait');
  var piEl=document.getElementById('asrPollInterval');
  if(rid==='volc.bigasr.auc_idle'){
    if(mwEl)mwEl.value='3600';
    if(piEl)piEl.value='10';
  }else{
    if(mwEl)mwEl.value='300';
    if(piEl)piEl.value='5';
  }
}

function saveASRConfig(){
  var payload={
    provider:'volcengine',
    app_id:document.getElementById('asrAppId').value,
    api_key:document.getElementById('asrKey').value,
    resource_id:document.getElementById('asrResourceId').value,
    language:document.getElementById('asrLang').value,
    max_wait_seconds:parseInt(document.getElementById('asrMaxWait').value)||300,
    poll_interval_seconds:parseInt(document.getElementById('asrPollInterval').value)||5,
    enabled:document.getElementById('asrEnabled').checked
  };
  fetch('/admin/asr/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0)showToast('ASR\u914d\u7f6e\u5df2\u4fdd\u5b58','success');
    else showToast('\u4fdd\u5b58\u5931\u8d25: '+(d.error||''),'error');
  }).catch(function(){showToast('\u4fdd\u5b58\u5931\u8d25','error')});
}

function testASR(){
  var rid=document.getElementById('asrResourceId').value;
  var aid=document.getElementById('asrAppId').value;
  var key=document.getElementById('asrKey').value;
  if(!aid||!key){showToast('\u8bf7\u5148\u586b\u5199 APP ID \u548c Access Token','error');return}
  var resultEl=document.getElementById('asrTestResult');
  resultEl.style.display='block';
  resultEl.innerHTML='<div style="padding:8px;background:#FEF3C7;border-radius:4px;font-size:13px">\u6d4b\u8bd5\u4e2d...</div>';
  var testURL='https://tos-volc-ai.bytedance.com/resource/doubao_bigmodel_asr_test.wav';
  var payload={test_url:testURL};
  fetch('/admin/asr/test',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){
      var text=(d.data&&d.data.text)||'';
      var chars=(d.data&&d.data.char_count)||text.length;
      resultEl.innerHTML='<div style="padding:8px;background:#D1FAE5;border-radius:4px;font-size:13px">'+
        '<strong style="color:#059669">\u2705 \u8fde\u63a5\u6210\u529f</strong> ('+chars+' \u5b57)'+
        (text?'<br><span style="color:#374151;margin-top:4px;display:block">\u8f6c\u5f55\u9884\u89c8: '+esc(text.substring(0,200))+(text.length>200?'...':'')+'</span>':'')+
        '</div>';
    }else{
      resultEl.innerHTML='<div style="padding:8px;background:#FEE2E2;border-radius:4px;font-size:13px">'+
        '<strong style="color:#DC2626">\u274c \u8fde\u63a5\u5931\u8d25</strong><br><span style="color:#374151">'+esc(d.error||'\u672a\u77e5\u9519\u8bef')+'</span></div>';
    }
  }).catch(function(e){
    resultEl.innerHTML='<div style="padding:8px;background:#FEE2E2;border-radius:4px;font-size:13px">'+
      '<strong style="color:#DC2626">\u274c \u8bf7\u6c42\u5931\u8d25</strong><br>'+esc(e.message)+'</div>';
  });
}

var allRelevanceRules=[];
function loadRelevanceRules(){
  var app=document.getElementById('app');
  app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>\u52a0\u8f7d\u4e2d...</div>';
  fetch('/admin/relevance/rules').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    allRelevanceRules=d.data||[];
    renderRelevanceRules();
  }).catch(function(e){app.innerHTML='<div class="card"><h3>\u52a0\u8f7d\u5931\u8d25</h3><p>'+esc(e.message)+'</p></div>'});
}

function renderRelevanceRules(){
  var rules=allRelevanceRules;
  var grouped={};
  rules.forEach(function(r){
    var cat=r.category||'default';
    if(!grouped[cat])grouped[cat]=[];
    grouped[cat].push(r);
  });
  var h='<div class="card"><h3 style="font-size:16px;margin-bottom:12px">\u76f8\u5173\u6027\u89c4\u5219\u7ba1\u7406</h3>';
  h+='<div style="display:grid;grid-template-columns:1fr 1fr 1fr 1fr auto;gap:8px;margin-bottom:12px;align-items:end">'+
    '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">\u5173\u952e\u8bcd</label><input id="rl_kw" placeholder="\u5173\u952e\u8bcd" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
    '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">\u5206\u7c7b</label><select id="rl_cat" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
    '<option value="social_insurance">\u793e\u4fdd</option><option value="housing_fund">\u516c\u79ef\u91d1</option><option value="employment">\u5c31\u4e1a</option><option value="medical">\u533b\u7597</option><option value="pension">\u517b\u8001</option><option value="subsidy">\u8865\u8d34</option><option value="default">\u9ed8\u8ba4</option></select></div>'+
    '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">\u6743\u91cd</label><select id="rl_wt" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
    '<option value="1">1</option><option value="2">2</option></select></div>'+
    '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">\u8303\u56f4</label><select id="rl_scope" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
    '<option value="all">\u5168\u90e8</option><option value="douyin">\u6296\u97f3</option><option value="wechat">\u5fae\u4fe1</option><option value="govsite">\u653f\u5e9c\u7f51\u7ad9</option></select></div>'+
    '<button class="btn btn-success" onclick="addRelevanceRule()">\u6dfb\u52a0</button></div>';
  h+='</div>';

  Object.keys(grouped).forEach(function(cat){
    h+='<div class="card"><h4 style="font-size:14px;margin-bottom:8px;color:#1A56DB">\u5206\u7c7b: '+esc(cat)+' ('+grouped[cat].length+' \u6761)</h4>';
    h+='<div style="overflow-x:auto"><table><tr><th>\u542f\u7528</th><th>\u5173\u952e\u8bcd</th><th>\u6743\u91cd</th><th>\u8303\u56f4</th><th>\u64cd\u4f5c</th></tr>';
    grouped[cat].forEach(function(r){
      h+='<tr><td><label class="toggle"><input type="checkbox" '+(r.enabled!==false?'checked':'')+' onchange="toggleRule('+r.id+',this.checked)"><span class="slider"></span></label></td>'+
        '<td style="font-weight:600;font-size:13px">'+esc(r.keyword)+'</td>'+
        '<td>'+esc(r.weight)+'</td>'+
        '<td><span class="badge bg-blue">'+esc(r.scope||'all')+'</span></td>'+
        '<td><button class="btn btn-danger btn-sm" onclick="deleteRelevanceRule('+r.id+')">\u5220\u9664</button></td></tr>';
    });
    h+='</table></div></div>';
  });
  if(rules.length===0)h+='<div class="card" style="text-align:center;color:#9CA3AF;padding:30px">\u6682\u65e0\u89c4\u5219\uff0c\u8bf7\u6dfb\u52a0</div>';

  h+='<div class="card" style="margin-top:12px"><h3 style="font-size:16px;margin-bottom:8px">\u6d4b\u8bd5\u76f8\u5173\u6027</h3>';
  h+='<div style="margin-bottom:8px"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">\u8f93\u5165\u6587\u672c</label>'+
    '<textarea id="rl_test_text" style="min-height:80px" placeholder="\u7c98\u8d34\u6587\u672c\u5185\u5bb9..."></textarea></div>';
  h+='<div style="display:grid;grid-template-columns:1fr 1fr auto;gap:8px;align-items:end;margin-bottom:8px">'+
    '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Source ID</label><input id="rl_test_src" placeholder="source_id" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
    '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">Crawl Type</label><select id="rl_test_ct" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
    '<option value="">\u65e0</option>'+sourceTypeList.map(function(t){return '<option value="'+t.v+'">'+t.l+'</option>'}).join('')+'</select></div>'+
    '<button class="btn btn-primary" onclick="testRelevance()">\u6d4b\u8bd5</button></div>';
  h+='<div id="rl_test_result"></div></div>';

  h+='<div class="card" style="margin-top:12px"><h3 style="font-size:16px;margin-bottom:8px">\u6279\u91cf\u5bfc\u5165</h3>';
  h+='<div style="margin-bottom:8px"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">JSON \u6570\u7ec4</label>'+
    '<textarea id="rl_bulk_json" style="min-height:120px" placeholder=\'[{"keyword":"\u793e\u4fdd","category":"social_insurance","weight":1,"scope":"all"}]\'></textarea></div>'+
    '<button class="btn btn-success" onclick="bulkImportRules()">\u5bfc\u5165</button></div>';

  document.getElementById('app').innerHTML=h;
}

function addRelevanceRule(){
  var kw=document.getElementById('rl_kw').value.trim();
  if(!kw){showToast('\u5173\u952e\u8bcd\u4e0d\u80fd\u4e3a\u7a7a','error');return}
  fetch('/admin/relevance/rules/create',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({
      keyword:kw,
      category:document.getElementById('rl_cat').value,
      weight:parseInt(document.getElementById('rl_wt').value)||1,
      scope:document.getElementById('rl_scope').value
    })}).then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('\u5df2\u6dfb\u52a0','success');loadRelevanceRules()}
    else showToast('\u6dfb\u52a0\u5931\u8d25: '+(d.error||''),'error');
  }).catch(function(){showToast('\u8bf7\u6c42\u5931\u8d25','error')});
}

function deleteRelevanceRule(id){
  if(!confirm('\u786e\u5b9a\u5220\u9664\u8be5\u89c4\u5219\uff1f'))return;
  fetch('/admin/relevance/rules/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id:id})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('\u5df2\u5220\u9664','success');loadRelevanceRules()}
    else showToast('\u5220\u9664\u5931\u8d25: '+(d.error||''),'error');
  }).catch(function(){showToast('\u8bf7\u6c42\u5931\u8d25','error')});
}

function toggleRule(id,enabled){
  fetch('/admin/relevance/rules/update',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({id:id,enabled:enabled})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0)showToast(enabled?'\u5df2\u542f\u7528':'\u5df2\u7981\u7528','success');
    else showToast('\u64cd\u4f5c\u5931\u8d25','error');
  }).catch(function(){showToast('\u64cd\u4f5c\u5931\u8d25','error')});
}

function testRelevance(){
  var text=document.getElementById('rl_test_text').value.trim();
  if(!text){showToast('\u8bf7\u8f93\u5165\u6d4b\u8bd5\u6587\u672c','error');return}
  fetch('/admin/relevance/test',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({
      text:text,
      source_id:document.getElementById('rl_test_src').value,
      crawl_type:document.getElementById('rl_test_ct').value
    })}).then(function(r){return r.json()}).then(function(d){
    if(d.code!==0){showToast('\u6d4b\u8bd5\u5931\u8d25: '+(d.error||''),'error');return}
    var res=d.data;
    var h='<div style="padding:12px;background:#F9FAFB;border-radius:6px;font-size:13px">'+
      '<div style="margin-bottom:6px"><strong>\u5f97\u5206:</strong> <span style="font-size:18px;font-weight:700;color:'+(res.score>=0.5?'#059669':'#EF4444')+'">'+(res.score||0).toFixed(3)+'</span></div>'+
      '<div><strong>\u5339\u914d\u5173\u952e\u8bcd:</strong> '+(res.matched_keywords&&res.matched_keywords.length>0?
      res.matched_keywords.map(function(k){return '<span class="badge bg-blue" style="margin:2px">'+esc(k)+'</span>'}).join(''):
      '<span style="color:#9CA3AF">\u65e0</span>')+'</div></div>';
    document.getElementById('rl_test_result').innerHTML=h;
  }).catch(function(){showToast('\u8bf7\u6c42\u5931\u8d25','error')});
}

function bulkImportRules(){
  var raw=document.getElementById('rl_bulk_json').value.trim();
  if(!raw){showToast('\u8bf7\u8f93\u5165JSON','error');return}
  var arr;
  try{arr=JSON.parse(raw)}catch(e){showToast('JSON\u89e3\u6790\u5931\u8d25: '+e.message,'error');return}
  if(!Array.isArray(arr)){showToast('\u8bf7\u8f93\u5165JSON\u6570\u7ec4','error');return}
  fetch('/admin/relevance/bulk-import',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({rules:arr})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('\u5df2\u5bfc\u5165 '+arr.length+' \u6761\u89c4\u5219','success');loadRelevanceRules()}
    else showToast('\u5bfc\u5165\u5931\u8d25: '+(d.error||''),'error');
  }).catch(function(){showToast('\u8bf7\u6c42\u5931\u8d25','error')});
}

var failureTrendChart=null,failureSourceChart=null,failureReasonChart=null;
var failureTrendDays=7;

function loadFailures(){
  var app=document.getElementById('app');
  app.innerHTML='<div class="stat-row" id="fSummary"></div>'+
    '<div class="card"><h3 style="margin-bottom:8px">\u5931\u8d25\u8d8b\u52bf <button class="btn btn-sm btn-outline" onclick="loadFailureTrend(7)" id="fTrend7">7\u5929</button> <button class="btn btn-sm btn-outline" onclick="loadFailureTrend(30)" id="fTrend30">30\u5929</button></h3><canvas id="fTrendCanvas" height="200"></canvas></div>'+
    '<div class="flex-row"><div class="card" style="flex:1;min-width:300px"><h3 style="margin-bottom:8px">\u6309\u6765\u6e90\u5206\u5e03</h3><canvas id="fSourceCanvas" height="250"></canvas></div>'+
    '<div class="card" style="flex:1;min-width:300px"><h3 style="margin-bottom:8px">Top \u5931\u8d25\u539f\u56e0</h3><canvas id="fReasonCanvas" height="250"></canvas></div></div>'+
    '<div class="card"><div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px"><h3>\u5931\u8d25\u660e\u7ec6</h3><div><select id="fTypeFilter" onchange="loadFailureDetail()" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px"><option value="">\u5168\u90e8</option><option value="extract">\u63d0\u53d6\u5931\u8d25</option><option value="video">\u89c6\u9891\u63d0\u53d6\u5931\u8d25</option></select> <button class="btn btn-primary btn-sm" onclick="retrySelected()">\u91cd\u8bd5\u9009\u4e2d</button></div></div><div id="fDetailTable"></div></div>';

  fetch('/admin/failures/summary').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)return;
    var s=d.data;
    document.getElementById('fSummary').innerHTML=
      '<div class="stat-card"><span class="stat-num c-red">'+s.crawl_failures+'</span><span class="stat-label">\u722c\u53d6\u5931\u8d25</span></div>'+
      '<div class="stat-card"><span class="stat-num c-orange">'+s.extract_failures+'</span><span class="stat-label">\u63d0\u53d6\u5931\u8d25</span></div>'+
      '<div class="stat-card"><span class="stat-num" style="color:#7C3AED">'+s.video_failures+'</span><span class="stat-label">\u89c6\u9891\u63d0\u53d6\u5931\u8d25</span></div>';
  });

  loadFailureTrend(7);
  loadFailureSource();
  loadFailureReasons();
  loadFailureDetail();
}

function loadFailureTrend(days){
  failureTrendDays=days;
  var btn7=document.getElementById('fTrend7'),btn30=document.getElementById('fTrend30');
  if(btn7){btn7.className='btn btn-sm '+(days===7?'btn-primary':'btn-outline')}
  if(btn30){btn30.className='btn btn-sm '+(days===30?'btn-primary':'btn-outline')}
  fetch('/admin/failures/trend?days='+days).then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)return;
    var data=d.data;
    var labels=[],crawl=[],extract=[],video=[];
    data.forEach(function(p){labels.push(p.date);crawl.push(p.crawl_failures);extract.push(p.extract_failures);video.push(p.video_failures)});
    if(failureTrendChart)failureTrendChart.destroy();
    var ctx=document.getElementById('fTrendCanvas');
    if(!ctx)return;
    failureTrendChart=new Chart(ctx,{type:'line',data:{labels:labels,datasets:[
      {label:'\u722c\u53d6\u5931\u8d25',data:crawl,borderColor:'#EF4444',backgroundColor:'rgba(239,68,68,0.1)',fill:true,tension:0.3},
      {label:'\u63d0\u53d6\u5931\u8d25',data:extract,borderColor:'#D97706',backgroundColor:'rgba(217,119,6,0.1)',fill:true,tension:0.3},
      {label:'\u89c6\u9891\u5931\u8d25',data:video,borderColor:'#7C3AED',backgroundColor:'rgba(124,58,237,0.1)',fill:true,tension:0.3}
    ]},options:{responsive:true,scales:{y:{beginAtZero:true}},plugins:{legend:{position:'bottom'}}}});
  });
}

function loadFailureSource(){
  fetch('/admin/failures/by-source').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)return;
    var data=d.data.slice(0,8);
    var labels=[],counts=[];
    data.forEach(function(e){var total=e.crawl_failures+e.extract_failures+e.video_failures;labels.push(e.source_name||e.source_id);counts.push(total)});
    if(failureSourceChart)failureSourceChart.destroy();
    var ctx=document.getElementById('fSourceCanvas');
    if(!ctx)return;
    failureSourceChart=new Chart(ctx,{type:'doughnut',data:{labels:labels,datasets:[{data:counts,backgroundColor:['#EF4444','#D97706','#7C3AED','#1A56DB','#059669','#EC4899','#6366F1','#14B8A6']}]},options:{responsive:true,plugins:{legend:{position:'bottom',labels:{boxWidth:12}}}}});
  });
}

function loadFailureReasons(){
  fetch('/admin/failures/top-reasons?limit=10').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)return;
    var data=d.data;
    var labels=[],counts=[];
    data.forEach(function(r){labels.push(r.reason.length>50?r.reason.substring(0,50)+'...':r.reason);counts.push(r.count)});
    if(failureReasonChart)failureReasonChart.destroy();
    var ctx=document.getElementById('fReasonCanvas');
    if(!ctx)return;
    failureReasonChart=new Chart(ctx,{type:'bar',data:{labels:labels,datasets:[{label:'\u6b21\u6570',data:counts,backgroundColor:'#EF4444'}]},options:{indexAxis:'y',responsive:true,scales:{x:{beginAtZero:true}},plugins:{legend:{display:false}}}});
  });
}

var failureDetailData=[];
function loadFailureDetail(){
  var type=document.getElementById('fTypeFilter')?document.getElementById('fTypeFilter').value:'';
  fetch('/admin/failures/failed-raw-texts?type='+type+'&limit=100').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)return;
    failureDetailData=d.data;
    var h='<table><thead><tr><th><input type="checkbox" id="fCheckAll" onchange="toggleAllFailCheck()"></th><th>ID</th><th>\u6765\u6e90</th><th>\u6807\u9898</th><th>\u9519\u8bef</th><th>\u7c7b\u578b</th><th>\u65f6\u95f4</th><th>\u64cd\u4f5c</th></tr></thead><tbody>';
    d.data.forEach(function(e,i){
      var typeBadge=e.failure_type==='video'?'bg-purple':'bg-orange';
      h+='<tr><td><input type="checkbox" class="fCheck" data-idx="'+i+'"></td><td>'+e.id+'</td><td>'+esc(e.source_name||e.source_id)+'</td><td>'+esc(e.title)+'</td><td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="'+esc(e.error_reason)+'">'+esc(e.error_reason)+'</td><td><span class="badge '+typeBadge+'">'+(e.failure_type==='video'?'\u89c6\u9891':'\u63d0\u53d6')+'</span></td><td>'+(e.failed_at||'')+'</td><td><button class="btn btn-sm btn-outline" onclick="retryOne('+e.id+')">\u91cd\u8bd5</button></td></tr>';
    });
    h+='</tbody></table>';
    if(d.data.length===0)h='<div style="text-align:center;padding:20px;color:#6B7280">\u6682\u65e0\u5931\u8d25\u8bb0\u5f55</div>';
    document.getElementById('fDetailTable').innerHTML=h;
  });
}

function toggleAllFailCheck(){
  var checked=document.getElementById('fCheckAll').checked;
  document.querySelectorAll('.fCheck').forEach(function(cb){cb.checked=checked});
}

function retryOne(id){
  fetch('/admin/failures/retry',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({raw_text_id:id})}).then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('\u5df2\u91cd\u65b0\u5165\u961f','success');loadFailureDetail()}
    else showToast(d.message||'\u91cd\u8bd5\u5931\u8d25','error');
  }).catch(function(){showToast('\u8bf7\u6c42\u5931\u8d25','error')});
}

function retrySelected(){
  var ids=[];
  document.querySelectorAll('.fCheck:checked').forEach(function(cb){
    var idx=parseInt(cb.dataset.idx);
    if(failureDetailData[idx])ids.push(failureDetailData[idx].id);
  });
  if(ids.length===0){showToast('\u8bf7\u5148\u9009\u62e9\u8981\u91cd\u8bd5\u7684\u9879\u76ee','error');return}
  var promises=ids.map(function(id){
    return fetch('/admin/failures/retry',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({raw_text_id:id})}).then(function(r){return r.json()});
  });
  Promise.all(promises).then(function(){
    showToast('\u5df2\u91cd\u65b0\u5165\u961f '+ids.length+' \u9879','success');
    loadFailureDetail();
  }).catch(function(){showToast('\u90e8\u5206\u91cd\u8bd5\u5931\u8d25','error')});
}

</script>
</body>
</html>`
