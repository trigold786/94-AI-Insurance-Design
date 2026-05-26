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
.bg-red{background:#FEE2E2;color:#EF4444}.bg-blue{background:#DBEAFE;color:#1A56DB}
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

<script>
var navItems=[
  {id:'dashboard',label:'\u4eea\u8868\u76d8'},
  {id:'sources',label:'\u6570\u636e\u6e90\u7ba1\u7406'},
  {id:'claims',label:'\u653f\u7b56\u5ba1\u6838'},
  {id:'search',label:'\u8bed\u4e49\u641c\u7d22'},
  {id:'extract',label:'AI\u63d0\u53d6'},
  {id:'logs',label:'\u722c\u53d6\u65e5\u5fd7'},
  {id:'import',label:'+\u5bfc\u5165\u653f\u7b56',style:'color:#059669;font-weight:600'}
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
  else if(id==='logs')loadLogs();
  else if(id==='import')showImportForm();
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

    // 数据源一览
    h+='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">数据源一览</h3><div id="miniSources">加载中...</div></div>';
    document.getElementById('app').innerHTML=h;

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
    '<option value="">全部类型</option><option value="govsite"'+(ft==='govsite'?' selected':'')+'>政府网站</option><option value="rss"'+(ft==='rss'?' selected':'')+'>RSS</option><option value="douyin"'+(ft==='douyin'?' selected':'')+'>抖音</option><option value="manual"'+(ft==='manual'?' selected':'')+'>手动</option><option value="file"'+(ft==='file'?' selected':'')+'>文件</option></select>'+
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
      '<option value="govsite">政府网站</option><option value="file">文件</option><option value="rss">RSS</option><option value="manual">手动</option><option value="douyin">抖音</option></select></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">级别</label><select id="sf_level" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      '<option value="HIGH">HIGH</option><option value="MEDIUM" selected>MEDIUM</option><option value="LOW">LOW</option></select></div>'+
      '<div id="sf_url_wrap" style="grid-column:1/3"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">URL</label><input id="sf_url" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px" placeholder="https://..."></div>'+
      '<div id="sf_interval_wrap"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">爬取间隔(秒)</label><input id="sf_interval" type="number" value="86400" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>'+
      '<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">地区代码</label><input id="sf_region" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px" placeholder="310000"></div>'+
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

var sourceTypeMap={'govsite':'政府网站','file':'文件','rss':'RSS','manual':'手动','douyin':'抖音'};
var sourceTypeBadge={'govsite':'bg-blue','file':'bg-yellow','rss':'bg-green','manual':'bg-yellow','douyin':'bg-red'};
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

function showSourceForm(id){
  editingSourceId=id||'';
  var modal=document.getElementById('sourceModal');
  document.getElementById('sourceModalTitle').textContent=id?'编辑数据源':'新增数据源';
  var idInput=document.getElementById('sf_id');
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
        document.getElementById('sf_region').value=s.region_code;
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
    document.getElementById('sf_region').value='';
    onTypeChange();
  }
  modal.style.display='block';
}

function closeSourceModal(){document.getElementById('sourceModal').style.display='none'}

function saveSource(){
  var id=document.getElementById('sf_id').value.trim();
  var name=document.getElementById('sf_name').value.trim();
  if(!id||!name){showToast('ID和名称不能为空','error');return}
  var payload={
    source_id:id,source_name:name,
    crawl_type:document.getElementById('sf_type').value,
    source_level:document.getElementById('sf_level').value,
    source_url:document.getElementById('sf_url').value,
    interval_sec:parseInt(document.getElementById('sf_interval').value)||86400,
    region_code:document.getElementById('sf_region').value
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
  'ali_bailian':{name:'\u963f\u91cc\u4e91\u767e\u70bc',endpoint:'https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation',model:'qwen-plus'},
  'volc_ark':{name:'\u706b\u5c71\u65b9\u821f',endpoint:'https://ark.cn-beijing.volces.com/api/v3/chat/completions',model:'doubao-pro-32k'},
  'opencode_go':{name:'OpenCode Go',endpoint:'http://localhost:11434/v1/chat/completions',model:'opencode-go'}
};

function loadExtract(){
  var app=document.getElementById('app');app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>加载中...</div>';
  Promise.all([
    fetch('/admin/llm/config').then(function(r){return r.json()}),
    fetch('/admin/llm/status').then(function(r){return r.json()}),
    fetch('/admin/llm/pending').then(function(r){return r.json()})
  ]).then(function(results){
    var cfg=results[0].data,st=results[1].data,pending=results[2].data||[];
    var prov=extractProviderMap[cfg.provider]||extractProviderMap['deepseek'];
    var h='<div class="card"><h3 style="font-size:16px;margin-bottom:12px">AI 政策提取配置</h3>';
    h+='<div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">';
    h+='<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">提供商</label>'+
      '<select id="llmProvider" onchange="updateLLMEndpoint()" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      '<option value="deepseek" '+(cfg.provider==='deepseek'?'selected':'')+'>DeepSeek</option>'+
      '<option value="ali_bailian" '+(cfg.provider==='ali_bailian'?'selected':'')+'>阿里云百炼</option>'+
      '<option value="volc_ark" '+(cfg.provider==='volc_ark'?'selected':'')+'>火山方舟</option>'+
      '<option value="opencode_go" '+(cfg.provider==='opencode_go'?'selected':'')+'>OpenCode Go</option>'+
      '</select></div>';
    h+='<div><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">模型</label><input id="llmModel" value="'+esc(cfg.model_name||prov.model)+'" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>';
    h+='<div style="grid-column:1/3"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">API 地址</label><input id="llmEndpoint" value="'+esc(cfg.endpoint||prov.endpoint)+'" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>';
    h+='<div style="grid-column:1/3"><label style="font-size:12px;color:#6B7280;display:block;margin-bottom:4px">API Key</label><input id="llmKey" type="password" value="'+esc(cfg.api_key)+'" placeholder="输入 API Key" style="width:100%;padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px"></div>';
    h+='<div><label class="toggle"><input type="checkbox" id="llmEnabled" '+(cfg.api_key?'checked':'')+' onchange="saveLLMConfig()"><span class="slider"></span></label><span style="font-size:13px;margin-left:8px">启用 LLM 提取</span></div>';
    h+='<div style="text-align:right"><button class="btn btn-primary" onclick="saveLLMConfig()">保存配置</button></div>';
    h+='</div></div>';

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
        '<option value="">全部</option><option value="govsite">政府网站</option><option value="rss">RSS</option><option value="douyin">抖音</option><option value="manual">手动</option><option value="file">文件</option></select>'+
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
  }).catch(function(e){app.innerHTML='<div class="card"><h3>加载失败</h3><p>'+esc(e.message)+'</p></div>'});
}

function updateLLMEndpoint(){
  var p=document.getElementById('llmProvider').value;
  var info=extractProviderMap[p]||extractProviderMap['deepseek'];
  document.getElementById('llmEndpoint').value=info.endpoint;
  document.getElementById('llmModel').value=info.model;
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
      enabled:!!document.getElementById('llmKey').value
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

var logStartDate='',logEndDate='';

function loadLogs(){
  var app=document.getElementById('app');
  var today=new Date();var weekAgo=new Date(today);weekAgo.setDate(weekAgo.getDate()-7);
  var sd=logStartDate||weekAgo.toISOString().slice(0,10);
  var ed=logEndDate||today.toISOString().slice(0,10);
  var url='/admin/logs?start_date='+sd+'&end_date='+ed;
  fetch(url).then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    var h='<div style="display:flex;gap:8px;align-items:center;margin-bottom:12px;flex-wrap:wrap">'+
      '<span style="font-size:13px;color:#6B7280">起始</span>'+
      '<input type="date" id="logStart" value="'+sd+'" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      '<span style="font-size:13px;color:#6B7280">结束</span>'+
      '<input type="date" id="logEnd" value="'+ed+'" style="padding:4px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px">'+
      '<button class="btn btn-primary btn-sm" onclick="filterLogs()">\u67e5\u8be2</button>'+
      '<span style="font-size:13px;color:#6B7280">\u5171 '+d.data.length+' \u6761\u8bb0\u5f55</span></div>'+
      '<div style="overflow-x:auto"><table><tr><th>\u65f6\u95f4</th><th>\u6570\u636e\u6e90</th><th>\u72b6\u6001</th><th>\u63d0\u53d6ID</th><th>\u5185\u5bb9\u6982\u8981</th><th>\u9519\u8bef\u4fe1\u606f</th></tr>';
    d.data.forEach(function(l){
      var claimCell=l.extracted_claim_id?'<a href="#claims" onclick="loadClaims(undefined,undefined,undefined,undefined);return false" style="color:#1A56DB;text-decoration:none;font-size:12px">'+esc(l.extracted_claim_id)+'</a>':'-';
      var summaryCell=l.content_summary?'<span style="font-size:12px;color:#374151">'+esc(l.content_summary)+'</span>':'-';
      h+='<tr><td style="font-size:11px;color:#9CA3AF">'+esc(l.crawled_at)+'</td><td>'+esc(l.source_name||l.source_id)+'</td>'+
      '<td>'+(l.status==='success'?'<span class="badge bg-green">\u6210\u529f</span>':'<span class="badge bg-red">\u5931\u8d25</span>')+'</td>'+
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
  loadLogs();
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
  }).catch(function(){showToast('\u8bf7\u6c42\u5931\u8d25','error')});
  btn.disabled=false;btn.textContent='\u89e3\u6790\u5e76\u5bfc\u5165';
}

</script>
</body>
</html>`
