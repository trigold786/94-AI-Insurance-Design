package admin

const adminHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>LLM Gateway - 管理后台</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#F5F7FA;color:#333;font-size:14px}
.header{background:linear-gradient(135deg,#1A56DB,#3B82F6);color:#fff;padding:14px 24px;display:flex;justify-content:space-between;align-items:center}
.header h1{font-size:18px;font-weight:600}.header span{font-size:12px;opacity:0.8}
.nav{display:flex;gap:0;background:#fff;border-bottom:2px solid #E5E7EB;padding:0 8px;overflow-x:auto}
.nav-item{padding:10px 18px;cursor:pointer;font-size:13px;color:#6B7280;border-bottom:2px solid transparent;margin-bottom:-2px;white-space:nowrap}
.nav-item.active{color:#1A56DB;border-bottom-color:#1A56DB;font-weight:600}
.nav-item:hover{color:#1A56DB}
.container{max-width:1200px;margin:16px auto;padding:0 16px}
.card{background:#fff;border-radius:10px;padding:16px;margin-bottom:12px;box-shadow:0 1px 4px rgba(0,0,0,0.06)}
table{width:100%;border-collapse:collapse;font-size:13px}
th{text-align:left;padding:8px 10px;border-bottom:2px solid #E5E7EB;color:#6B7280;font-weight:600;font-size:12px}
td{padding:8px 10px;border-bottom:1px solid #F3F4F6}
tr:hover{background:#F9FAFB}
.badge{display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:500}
.bg-green{background:#D1FAE5;color:#059669}.bg-yellow{background:#FEF3C7;color:#D97706}
.bg-red{background:#FEE2E2;color:#EF4444}.bg-blue{background:#DBEAFE;color:#1A56DB}
.btn{padding:6px 14px;border:none;border-radius:6px;font-size:12px;cursor:pointer;font-weight:500}
.btn-sm{padding:4px 10px;font-size:11px}
.btn-primary{background:#1A56DB;color:#fff}.btn-primary:hover{background:#1648B8}
.btn-success{background:#059669;color:#fff}.btn-success:hover{background:#047857}
.btn-danger{background:#EF4444;color:#fff}.btn-danger:hover{background:#DC2626}
.btn-outline{background:#fff;border:1px solid #D1D5DB;color:#374151}.btn-outline:hover{background:#F9FAFB}
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
input,select{padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px}
input:focus,select:focus{outline:none;border-color:#1A56DB;box-shadow:0 0 0 2px rgba(26,86,219,0.15)}
label{font-size:12px;color:#6B7280;display:block;margin-bottom:4px}
.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:10px}
.form-full{grid-column:1/3}
</style>
</head>
<body>
<div class="header"><h1>LLM Gateway - 管理后台</h1><span>v1.0.0</span></div>
<div class="nav" id="navBar"></div>
<div class="container" id="app"><div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>加载中...</div></div>
<div class="toast" id="toast"></div>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<script>
var navItems=[
  {id:'providers',label:'Provider配置'},
  {id:'usage',label:'用量统计'},
  {id:'connectivity',label:'连通性测试'}
];
var currentPanel='providers';
var providerDefaults={
  deepseek:{endpoint:'https://api.deepseek.com/v1/chat/completions',model:'deepseek-chat'},
  ali_bailian:{endpoint:'https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation',model:'qwen-plus'},
  volc_ark:{endpoint:'https://ark.cn-beijing.volces.com/api/v3/chat/completions',model:'doubao-pro-32k'},
  opencode_go:{endpoint:'http://localhost:11434/v1/chat/completions',model:'opencode-go'}
};
var providerLabels={
  deepseek:'DeepSeek',
  ali_bailian:'阿里云百炼',
  volc_ark:'火山方舟',
  opencode_go:'OpenCode Go'
};
var allProviders=[];

function initNav(){
  var h='';
  navItems.forEach(function(n){
    h+='<div class="nav-item'+(n.id==currentPanel?' active':'')+'" data-panel="'+n.id+'">'+n.label+'</div>';
  });
  document.getElementById('navBar').innerHTML=h;
  document.getElementById('navBar').addEventListener('click',function(e){
    var item=e.target.closest('.nav-item');
    if(item) switchPanel(item.dataset.panel);
  });
}
initNav();

function showToast(msg,type){
  var t=document.getElementById('toast');
  t.textContent=msg;t.style.display='block';
  t.style.background=type==='success'?'#059669':'#EF4444';
  t.style.opacity='1';
  setTimeout(function(){t.style.display='none'},3000);
}
function esc(s){if(typeof s!=='string')return '';var d=document.createElement('div');d.textContent=s;return d.innerHTML}

function switchPanel(id){
  currentPanel=id;
  window.location.hash='#'+id;
  document.querySelectorAll('.nav-item').forEach(function(n){n.classList.remove('active')});
  var el=document.querySelector('[data-panel="'+id+'"]');
  if(el) el.classList.add('active');
  var app=document.getElementById('app');
  app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>加载中...</div>';
  if(id==='providers') loadProviders();
  else if(id==='usage') loadUsage();
  else if(id==='connectivity') loadConnectivity();
}

window.addEventListener('hashchange',function(){
  var h=window.location.hash.replace('#','');
  if(h && navItems.some(function(n){return n.id===h})) switchPanel(h);
});
(function(){
  var h=window.location.hash.replace('#','');
  if(h && navItems.some(function(n){return n.id===h})){switchPanel(h);return}
  switchPanel('providers');
})();

function loadProviders(){
  fetch('/admin/providers').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0) throw new Error(d.message||'error');
    allProviders=d.data||[];
    renderProviders();
  }).catch(function(e){
    document.getElementById('app').innerHTML='<div class="card"><h3>加载失败</h3><p>'+esc(e.message)+'</p></div>';
  });
}

function renderProviders(){
  var h='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">已配置 Provider</h3>';
  h+='<div style="overflow-x:auto"><table><tr><th>Provider</th><th>Endpoint</th><th>Model</th><th>MaxTokens</th><th>Priority</th><th>启用</th><th>主用</th><th>操作</th></tr>';
  allProviders.forEach(function(p){
    h+='<tr>'+
      '<td><span class="badge bg-blue">'+esc(providerLabels[p.provider_name]||p.provider_name)+'</span></td>'+
      '<td style="font-size:11px;max-width:240px;overflow:hidden;text-overflow:ellipsis" title="'+esc(p.endpoint)+'">'+esc(p.endpoint)+'</td>'+
      '<td style="font-weight:600">'+esc(p.model_name)+'</td>'+
      '<td>'+(p.max_tokens||'-')+'</td>'+
      '<td>'+(p.priority||0)+'</td>'+
      '<td>'+(p.is_enabled?'<span class="badge bg-green">是</span>':'<span class="badge bg-red">否</span>')+'</td>'+
      '<td>'+(p.is_primary?'<span class="badge bg-green">是</span>':'<span class="badge bg-yellow">否</span>')+'</td>'+
      '<td><button class="btn btn-outline btn-sm" onclick="editProvider(\''+esc(p.provider_name)+'\')">编辑</button> '+
      '<button class="btn btn-primary btn-sm" onclick="testOne(\''+esc(p.provider_name)+'\')">测试</button></td>'+
      '</tr>';
  });
  if(allProviders.length===0){
    h+='<tr><td colspan="8" style="text-align:center;color:#9CA3AF;padding:20px">暂无 Provider 配置</td></tr>';
  }
  h+='</table></div></div>';

  h+='<div class="card"><h3 style="font-size:15px;margin-bottom:12px" id="formTitle">新增 Provider</h3>';
  h+='<div class="form-grid">'+
    '<div><label>Provider</label>'+
    '<select id="f_provider" onchange="onProviderChange()" style="width:100%">'+
    '<option value="deepseek">DeepSeek</option>'+
    '<option value="ali_bailian">阿里云百炼</option>'+
    '<option value="volc_ark">火山方舟</option>'+
    '<option value="opencode_go">OpenCode Go</option>'+
    '</select></div>'+
    '<div><label>API Key</label><input id="f_apikey" type="password" style="width:100%" placeholder="输入 API Key"></div>'+
    '<div class="form-full"><label>Endpoint</label><input id="f_endpoint" style="width:100%" placeholder="API Endpoint"></div>'+
    '<div><label>Model Name</label><input id="f_model" style="width:100%" placeholder="模型名称"></div>'+
    '<div><label>Max Tokens</label><input id="f_maxtokens" type="number" value="4096" style="width:100%"></div>'+
    '<div><label>Priority</label><input id="f_priority" type="number" value="0" style="width:100%"></div>'+
    '<div style="display:flex;align-items:center;gap:16px;padding-top:18px">'+
    '<label style="display:flex;align-items:center;gap:4px;cursor:pointer;font-size:13px;color:#333"><input type="checkbox" id="f_enabled" checked>启用</label>'+
    '<label style="display:flex;align-items:center;gap:4px;cursor:pointer;font-size:13px;color:#333"><input type="checkbox" id="f_primary">主用</label>'+
    '</div>'+
    '</div>'+
    '<div style="display:flex;gap:8px;margin-top:16px;justify-content:flex-end">'+
    '<button class="btn btn-outline" onclick="resetForm()">重置</button>'+
    '<button class="btn btn-primary" onclick="saveProvider()">保存</button>'+
    '</div></div>';

  document.getElementById('app').innerHTML=h;
  onProviderChange();
}

function onProviderChange(){
  var sel=document.getElementById('f_provider').value;
  var d=providerDefaults[sel];
  if(d){
    document.getElementById('f_endpoint').value=d.endpoint;
    document.getElementById('f_model').value=d.model;
  }
}

function resetForm(){
  document.getElementById('f_provider').value='deepseek';
  document.getElementById('f_apikey').value='';
  document.getElementById('f_maxtokens').value='4096';
  document.getElementById('f_priority').value='0';
  document.getElementById('f_enabled').checked=true;
  document.getElementById('f_primary').checked=false;
  document.getElementById('formTitle').textContent='新增 Provider';
  onProviderChange();
}

function editProvider(name){
  var p=allProviders.find(function(x){return x.provider_name===name});
  if(!p) return;
  document.getElementById('f_provider').value=p.provider_name;
  document.getElementById('f_endpoint').value=p.endpoint;
  document.getElementById('f_model').value=p.model_name;
  document.getElementById('f_maxtokens').value=p.max_tokens||4096;
  document.getElementById('f_priority').value=p.priority||0;
  document.getElementById('f_apikey').value='';
  document.getElementById('f_enabled').checked=p.is_enabled;
  document.getElementById('f_primary').checked=p.is_primary;
  document.getElementById('formTitle').textContent='编辑 Provider: '+(providerLabels[p.provider_name]||p.provider_name);
}

function saveProvider(){
  var payload={
    provider_name:document.getElementById('f_provider').value,
    api_key:document.getElementById('f_apikey').value,
    endpoint:document.getElementById('f_endpoint').value,
    model_name:document.getElementById('f_model').value,
    max_tokens:parseInt(document.getElementById('f_maxtokens').value)||4096,
    priority:parseInt(document.getElementById('f_priority').value)||0,
    is_enabled:document.getElementById('f_enabled').checked,
    is_primary:document.getElementById('f_primary').checked
  };
  if(!payload.provider_name){showToast('请选择 Provider','error');return}
  if(!payload.api_key){
    var existing=allProviders.find(function(x){return x.provider_name===payload.provider_name});
    if(!existing){showToast('新增 Provider 必须填写 API Key','error');return}
  }
  fetch('/admin/providers',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('已保存','success');resetForm();loadProviders()}
    else{showToast('保存失败: '+(d.message||''),'error')}
  }).catch(function(){showToast('请求失败','error')});
}

function testOne(name){
  showToast('正在测试 '+esc(providerLabels[name]||name)+'...','success');
  fetch('/admin/providers/test',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({provider_name:name})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code!==0){showToast('测试失败: '+(d.message||''),'error');return}
    var r2=d.data;
    if(r2.status==='ok'){
      showToast((providerLabels[name]||name)+' 测试通过 ('+r2.latency_ms+'ms)','success');
    }else{
      showToast((providerLabels[name]||name)+' 测试失败: '+(r2.error||''),'error');
    }
  }).catch(function(){showToast('请求失败','error')});
}

function loadUsage(){
  fetch('/admin/usage').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0) throw new Error(d.message||'error');
    var data=d.data||[];
    var h='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">近30天用量统计</h3>';

    if(data.length===0){
      h+='<div style="text-align:center;color:#9CA3AF;padding:40px">暂无用量数据</div>';
    }else{
      var dateSet={};
      var providerSet={};
      data.forEach(function(u){
        dateSet[u.date]=true;
        providerSet[u.provider_name]=true;
      });
      var dates=Object.keys(dateSet).sort();
      var providers=Object.keys(providerSet);

      h+='<div style="overflow-x:auto"><table><tr><th>日期</th>';
      providers.forEach(function(p){h+='<th>'+esc(providerLabels[p]||p)+' 输入</th><th>'+esc(providerLabels[p]||p)+' 输出</th><th>调用次数</th>'});
      h+='</tr>';
      var dataMap={};
      data.forEach(function(u){dataMap[u.date+'_'+u.provider_name]=u});
      dates.forEach(function(dt){
        h+='<tr><td style="font-weight:600">'+dt+'</td>';
        providers.forEach(function(p){
          var u=dataMap[dt+'_'+p];
          if(u){
            h+='<td>'+u.total_tokens_in.toLocaleString()+'</td><td>'+u.total_tokens_out.toLocaleString()+'</td><td>'+u.total_calls+'</td>';
          }else{
            h+='<td>0</td><td>0</td><td>0</td>';
          }
        });
        h+='</tr>';
      });
      h+='</table></div>';
      h+='<div style="margin-top:16px"><canvas id="usageChart" height="300"></canvas></div>';
    }
    h+='</div>';
    document.getElementById('app').innerHTML=h;

    if(data.length>0){
      var dateSet2={},providerSet2={};
      data.forEach(function(u){dateSet2[u.date]=true;providerSet2[u.provider_name]=true});
      var dates2=Object.keys(dateSet2).sort();
      var providers2=Object.keys(providerSet2);
      var colors=['#1A56DB','#059669','#D97706','#7C3AED','#EC4899','#0EA5E9'];
      var datasets=[];
      providers2.forEach(function(p,pi){
        var inData=[],outData=[];
        dates2.forEach(function(dt){
          var u=data.find(function(x){return x.date===dt && x.provider_name===p});
          inData.push(u?u.total_tokens_in:0);
          outData.push(u?u.total_tokens_out:0);
        });
        var c=colors[pi%colors.length];
        datasets.push({label:(providerLabels[p]||p)+' 输入',data:inData,backgroundColor:c+'CC',borderColor:c,borderWidth:1});
        datasets.push({label:(providerLabels[p]||p)+' 输出',data:outData,backgroundColor:c+'44',borderColor:c+'88',borderWidth:1,borderDash:[4,4]});
      });
      new Chart(document.getElementById('usageChart'),{
        type:'bar',
        data:{labels:dates2,datasets:datasets},
        options:{
          responsive:true,
          plugins:{legend:{position:'bottom',labels:{font:{size:11}}}},
          scales:{x:{stacked:false},y:{beginAtZero:true,ticks:{callback:function(v){return v>=1000?(v/1000)+'k':v}}}}
        }
      });
    }
  }).catch(function(e){
    document.getElementById('app').innerHTML='<div class="card"><h3>加载失败</h3><p>'+esc(e.message)+'</p></div>';
  });
}

var testResults=[];

function loadConnectivity(){
  var h='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">连通性测试</h3>';
  h+='<p style="font-size:13px;color:#6B7280;margin-bottom:12px">点击"全部测试"验证所有已启用 Provider 的连通性。</p>';
  h+='<button class="btn btn-primary" onclick="testAll()" id="testAllBtn">全部测试</button>';
  h+='<div id="testResults" style="margin-top:16px">';
  if(testResults.length>0){
    h+=renderTestResults();
  }else{
    h+='<div style="text-align:center;color:#9CA3AF;padding:30px">尚未进行测试</div>';
  }
  h+='</div></div>';
  document.getElementById('app').innerHTML=h;
}

function renderTestResults(){
  var h='<table><tr><th>Provider</th><th>延迟</th><th>状态</th><th>响应预览</th></tr>';
  testResults.forEach(function(r){
    var statusBadge=r.status==='ok'?'<span class="badge bg-green">成功</span>':'<span class="badge bg-red">失败</span>';
    h+='<tr>'+
      '<td><span class="badge bg-blue">'+esc(providerLabels[r.provider_name]||r.provider_name)+'</span></td>'+
      '<td>'+r.latency_ms+'ms</td>'+
      '<td>'+statusBadge+'</td>'+
      '<td style="font-size:12px;max-width:300px;overflow:hidden;text-overflow:ellipsis" title="'+esc(r.response_preview||r.error||'')+'">'+esc(r.response_preview||r.error||'-')+'</td>'+
      '</tr>';
  });
  h+='</table>';
  return h;
}

function testAll(){
  var btn=document.getElementById('testAllBtn');
  btn.disabled=true;btn.textContent='测试中...';
  testResults=[];
  document.getElementById('testResults').innerHTML='<div style="text-align:center;padding:20px"><div class="spinner"></div>测试中...</div>';

  var promises=allProviders.filter(function(p){return p.is_enabled}).map(function(p){
    return fetch('/admin/providers/test',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({provider_name:p.provider_name})})
    .then(function(r){return r.json()}).then(function(d){
      if(d.code===0) testResults.push(d.data);
    }).catch(function(){});
  });

  Promise.all(promises).then(function(){
    btn.disabled=false;btn.textContent='全部测试';
    document.getElementById('testResults').innerHTML=renderTestResults();
    var allOk=testResults.every(function(r){return r.status==='ok'});
    showToast(allOk?'全部测试通过':'部分测试失败',allOk?'success':'error');
  });
}
</script>
</body>
</html>`
