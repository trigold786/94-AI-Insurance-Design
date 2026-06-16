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
.card-header{font-size:15px;font-weight:600;margin-bottom:12px;padding-bottom:8px;border-bottom:1px solid #E5E7EB;display:flex;align-items:center;justify-content:space-between}
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
input,select{padding:6px 8px;border:1px solid #D1D5DB;border-radius:4px;font-size:13px;width:100%}
input:focus,select:focus{outline:none;border-color:#1A56DB;box-shadow:0 0 0 2px rgba(26,86,219,0.15)}
input[readonly]{background:#F3F4F6;color:#6B7280;cursor:default}
label{font-size:12px;color:#6B7280;display:block;margin-bottom:4px}
.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:10px}
.form-full{grid-column:1/-1}
.section-actions{display:flex;gap:8px;justify-content:flex-end;margin-top:12px;padding-top:10px;border-top:1px solid #F3F4F6}
.collapsible-header{display:flex;align-items:center;gap:6px;cursor:pointer;padding:8px 0;font-size:13px;color:#6B7280;user-select:none}
.collapsible-header:hover{color:#1A56DB}
.collapsible-header .arrow{transition:transform .2s;display:inline-block}
.collapsible-header.open .arrow{transform:rotate(90deg)}
.collapsible-body{display:none;padding-top:8px}
.collapsible-body.open{display:block}
.key-wrapper{display:flex;gap:4px;align-items:center}
.key-wrapper input{flex:1}
.key-toggle{font-size:11px;cursor:pointer;color:#6B7280;padding:4px 6px;border:1px solid #D1D5DB;border-radius:4px;background:#fff;white-space:nowrap}
.key-toggle:hover{color:#1A56DB;border-color:#1A56DB}
.inline-toggle{display:flex;align-items:center;gap:6px;font-size:13px;color:#333}
.inline-toggle input[type=checkbox]{width:auto}
</style>
</head>
<body>
<div class="header"><h1>LLM Gateway - 管理后台</h1><span>v2.1.0</span></div>
<div class="nav" id="navBar"></div>
<div class="container" id="app"><div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>加载中...</div></div>
<div class="toast" id="toast"></div>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<script>
var MODEL_REGISTRY={
  'deepseek':{
    name:'DeepSeek',
    models:{
      'deepseek-v4-pro':{name:'DeepSeek V4 Pro',api:'https://api.deepseek.com/v1/chat/completions',type:'llm',format:'openai',maxTokens:8192},
      'deepseek-v4-flash':{name:'DeepSeek V4 Flash',api:'https://api.deepseek.com/v1/chat/completions',type:'llm',format:'openai',maxTokens:8192}
    }
  },
  'volc_ark':{
    name:'火山方舟 (豆包)',
    models:{
      'doubao-seed-2.0-pro':{name:'Doubao Seed 2.0 Pro',api:'https://ark.cn-beijing.volces.com/api/coding/v3',type:'llm',format:'openai',maxTokens:8192},
      'doubao-seed-2.0-lite':{name:'Doubao Seed 2.0 Lite',api:'https://ark.cn-beijing.volces.com/api/coding/v3',type:'llm',format:'openai',maxTokens:8192},
      'doubao-seed-2.0-code':{name:'Doubao Seed 2.0 Code',api:'https://ark.cn-beijing.volces.com/api/coding/v3',type:'llm',format:'openai',maxTokens:8192},
      'doubao-pro-32k':{name:'豆包 Pro 32K',api:'https://ark.cn-beijing.volces.com/api/v3/chat/completions',type:'llm',format:'openai',maxTokens:4096},
      'doubao-pro-128k':{name:'豆包 Pro 128K',api:'https://ark.cn-beijing.volces.com/api/v3/chat/completions',type:'llm',format:'openai',maxTokens:4096},
      'doubao-lite-32k':{name:'豆包 Lite 32K',api:'https://ark.cn-beijing.volces.com/api/v3/chat/completions',type:'llm',format:'openai',maxTokens:4096},
      'doubao-embedding-vision':{name:'豆包 Embedding Vision',api:'https://ark.cn-beijing.volces.com/api/v3/embeddings/multimodal',type:'embedding',format:'ark_multimodal',dims:[1024,2048]}
    }
  },
  'ali_bailian':{
    name:'阿里云百炼 (通义千问)',
    models:{
      'qwen3.7-max':{name:'通义千问 3.7 Max',api:'https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions',type:'llm',format:'openai',maxTokens:8192},
      'qwen3.6-plus':{name:'通义千问 3.6 Plus',api:'https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions',type:'llm',format:'openai',maxTokens:8192},
      'qwen3.6-flash':{name:'通义千问 3.6 Flash',api:'https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions',type:'llm',format:'openai',maxTokens:8192},
      'text-embedding-v4':{name:'通义文本 Embedding V4',api:'https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings',type:'embedding',format:'openai',dims:[1024,1536]},
      'tongyi-embedding-vision-plus':{name:'通义图文 Embedding Plus',api:'https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings',type:'embedding',format:'openai',dims:[1024,1536]}
    }
  },
  'ollama':{
    name:'Ollama (本地)',
    editable_endpoint:true,
    models:{
      'gemma4:26b':{name:'Gemma 4 26B',api:'http://192.168.1.11:11434/v1',type:'llm',format:'openai',maxTokens:8192},
      'gemma4:31b':{name:'Gemma 4 31B',api:'http://192.168.1.11:11434/v1',type:'llm',format:'openai',maxTokens:8192}
    }
  },
  'opencode_go':{
    name:'OpenCode Go (本地)',
    models:{
      'opencode-go':{name:'OpenCode Go',api:'http://localhost:11434/v1/chat/completions',type:'llm',format:'openai',maxTokens:4096}
    }
  },
  'volcengine_asr':{
    name:'火山引擎 (语音)',
    models:{
      'volc.bigasr.auc':{name:'大模型语音识别 (标准版)',api:'https://openspeech.bytedance.com/api/v3/auc/bigmodel',type:'asr',format:'volcengine_asr'},
      'volc.bigasr.auc_idle':{name:'大模型语音识别 (闲时版)',api:'https://openspeech.bytedance.com/api/v3/auc/bigmodel',type:'asr',format:'volcengine_asr'}
    }
  }
};

var providerLabels={};
for(var pk in MODEL_REGISTRY){providerLabels[pk]=MODEL_REGISTRY[pk].name}

var navItems=[
  {id:'model-configs',label:'模型配置'},
  {id:'usage',label:'用量统计'},
  {id:'connectivity',label:'连通性测试'},
  {id:'providers',label:'Provider配置'}
];
var currentPanel='model-configs';
var allProviders=[];
var modelConfigs={};
var originalKeys={};

function initNav(){
  var h='';
  navItems.forEach(function(n){
    h+='<div class="nav-item'+(n.id==currentPanel?' active':'')+'" data-panel="'+n.id+'">'+n.label+'</div>';
  });
  document.getElementById('navBar').innerHTML=h;
  document.getElementById('navBar').addEventListener('click',function(e){
    var item=e.target.closest('.nav-item');
    if(item)switchPanel(item.dataset.panel);
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
  if(el)el.classList.add('active');
  var app=document.getElementById('app');
  app.innerHTML='<div class="card" style="text-align:center;padding:40px"><div class="spinner"></div>加载中...</div>';
  if(id==='model-configs')loadModelConfigs();
  else if(id==='usage')loadUsage();
  else if(id==='connectivity')loadConnectivity();
  else if(id==='providers')loadProviders();
}

window.addEventListener('hashchange',function(){
  var h=window.location.hash.replace('#','');
  if(h&&navItems.some(function(n){return n.id===h}))switchPanel(h);
});
(function(){
  var h=window.location.hash.replace('#','');
  if(h&&navItems.some(function(n){return n.id===h})){switchPanel(h);return}
  switchPanel('model-configs');
})();

function getProvidersForType(type){
  var result=[];
  for(var pk in MODEL_REGISTRY){
    var p=MODEL_REGISTRY[pk];
    var hasType=false;
    for(var mk in p.models){if(p.models[mk].type===type){hasType=true;break}}
    if(hasType)result.push({key:pk,name:p.name});
  }
  return result;
}

function getModelsForProviderAndType(providerKey,type){
  var p=MODEL_REGISTRY[providerKey];
  if(!p)return[];
  var result=[];
  for(var mk in p.models){
    if(p.models[mk].type===type)result.push({key:mk,info:p.models[mk]});
  }
  return result;
}

function loadModelConfigs(){
  fetch('/admin/model-configs').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    var configs=d.data||[];
    modelConfigs={};
    originalKeys={};
    configs.forEach(function(c){
      modelConfigs[c.function_key]=c;
    });
    renderModelConfigs();
  }).catch(function(e){
    document.getElementById('app').innerHTML='<div class="card"><h3>加载失败</h3><p>'+esc(e.message)+'</p></div>';
  });
}

function renderModelConfigs(){
  var h='';

  // Section 1: LLM 数据提取
  h+=renderLLMSection('llm_extract','LLM 数据提取');

  // Section 2: LLM 方案生成
  h+=renderLLMSection('llm_plan','LLM 方案生成');

  // Section 3: Embedding 向量
  h+=renderEmbeddingSection();

  // Section 4: ASR 语音识别
  h+=renderASRSection();

  document.getElementById('app').innerHTML=h;
  bindCollapsibles();
  bindKeyToggles();
  triggerProviderChanges();
}

function renderLLMSection(fnKey,title){
  var cfg=modelConfigs[fnKey]||{};
  var providers=getProvidersForType('llm');
  var html='<div class="card" id="card_'+fnKey+'">';
  html+='<div class="card-header"><span>'+esc(title)+'</span>';
  html+='<span class="badge '+(cfg.enabled?'bg-green':'bg-red')+'">'+(cfg.enabled?'已启用':'未启用')+'</span></div>';
  html+='<div class="form-grid">';

  html+='<div><label>Provider</label><select id="'+fnKey+'_provider" onchange="onProviderChangeLLM(\''+fnKey+'\')" data-type="llm">';
  providers.forEach(function(p){html+='<option value="'+p.key+'"'+(cfg.provider===p.key?' selected':'')+'>'+esc(p.name)+'</option>'});
  html+='</select></div>';

  var models=getModelsForProviderAndType(cfg.provider||providers[0].key,'llm');
  html+='<div><label>Model</label><select id="'+fnKey+'_model" onchange="onModelChangeLLM(\''+fnKey+'\')">';
  models.forEach(function(m){html+='<option value="'+m.key+'"'+(cfg.model_id===m.key?' selected':'')+'>'+esc(m.info.name)+'</option>'});
  html+='</select></div>';

  var epReadOnly='';
  var regP=MODEL_REGISTRY[cfg.provider||providers[0].key];
  if(regP&&!regP.editable_endpoint){epReadOnly=' readonly'}
  html+='<div class="form-full"><label>API Endpoint</label><input id="'+fnKey+'_endpoint"'+epReadOnly+' value="'+esc(cfg.api_endpoint||models[0].info.api)+'"></div>';

  html+='<div><label>API Key</label><div class="key-wrapper"><input id="'+fnKey+'_apikey" type="password" placeholder="输入 API Key" data-original="'+(cfg.api_key_masked?'masked':'')+'"><span class="key-toggle" onclick="toggleKeyEdit(\''+fnKey+'_apikey\')">编辑</span></div></div>';

  html+='<div><label>Max Tokens</label><input id="'+fnKey+'_maxtokens" type="number" value="'+(cfg.max_tokens||models[0].info.maxTokens||4096)+'"></div>';

  html+='<div class="form-full"><label class="inline-toggle"><input type="checkbox" id="'+fnKey+'_enabled" '+(cfg.enabled?'checked':'')+'>启用</label></div>';

  html+='</div>';

  // Backup collapsible
  html+='<div class="collapsible-header" onclick="toggleCollapsible(this)"><span class="arrow">\u25B6</span> 备用配置 (可选)</div>';
  html+='<div class="collapsible-body">';
  html+='<div class="form-grid">';
  html+='<div><label>备用 Provider</label><select id="'+fnKey+'_backup_provider" onchange="onBackupProviderChange(\''+fnKey+'\')" data-type="llm">';
  html+='<option value="">-- 不使用备用 --</option>';
  providers.forEach(function(p){html+='<option value="'+p.key+'"'+(cfg.backup_provider===p.key?' selected':'')+'>'+esc(p.name)+'</option>'});
  html+='</select></div>';

  var bkModels=[];
  if(cfg.backup_provider){
    bkModels=getModelsForProviderAndType(cfg.backup_provider,'llm');
  }
  html+='<div><label>备用 Model</label><select id="'+fnKey+'_backup_model" onchange="onBackupModelChange(\''+fnKey+'\')">';
  if(bkModels.length===0){html+='<option value="">-- 先选择 Provider --</option>'}
  else{bkModels.forEach(function(m){html+='<option value="'+m.key+'"'+(cfg.backup_model_id===m.key?' selected':'')+'>'+esc(m.info.name)+'</option>'})}
  html+='</select></div>';

  html+='<div class="form-full"><label>备用 API Endpoint</label><input id="'+fnKey+'_backup_endpoint" readonly value="'+esc(cfg.backup_api_endpoint||(bkModels.length?bkModels[0].info.api:''))+'"></div>';
  html+='<div><label>备用 API Key</label><div class="key-wrapper"><input id="'+fnKey+'_backup_apikey" type="password" placeholder="输入备用 API Key" data-original=""><span class="key-toggle" onclick="toggleKeyEdit(\''+fnKey+'_backup_apikey\')">编辑</span></div></div>';
  html+='</div></div>';

  html+='<div class="section-actions">';
  html+='<button class="btn btn-success btn-sm" onclick="testModelConfig(\''+fnKey+'\')">测试</button>';
  html+='<button class="btn btn-primary btn-sm" onclick="saveModelConfig(\''+fnKey+'\')">保存</button>';
  html+='</div>';
  html+='</div>';
  return html;
}

function renderEmbeddingSection(){
  var fnKey='embedding';
  var cfg=modelConfigs[fnKey]||{};
  var providers=getProvidersForType('embedding');
  var html='<div class="card" id="card_'+fnKey+'">';
  html+='<div class="card-header"><span>Embedding 向量</span>';
  html+='<span class="badge '+(cfg.enabled?'bg-green':'bg-red')+'">'+(cfg.enabled?'已启用':'未启用')+'</span></div>';
  html+='<div class="form-grid">';

  html+='<div><label>Provider</label><select id="'+fnKey+'_provider" onchange="onProviderChangeEmbedding()">';
  providers.forEach(function(p){html+='<option value="'+p.key+'"'+(cfg.provider===p.key?' selected':'')+'>'+esc(p.name)+'</option>'});
  html+='</select></div>';

  var models=getModelsForProviderAndType(cfg.provider||providers[0].key,'embedding');
  html+='<div><label>Model</label><select id="'+fnKey+'_model" onchange="onModelChangeEmbedding()">';
  models.forEach(function(m){html+='<option value="'+m.key+'"'+(cfg.model_id===m.key?' selected':'')+'>'+esc(m.info.name)+'</option>'});
  html+='</select></div>';

  var epReadOnlyE='';
  var regPE=MODEL_REGISTRY[cfg.provider||providers[0].key];
  if(regPE&&!regPE.editable_endpoint){epReadOnlyE=' readonly'}
  html+='<div class="form-full"><label>API Endpoint</label><input id="embedding_endpoint"'+epReadOnlyE+' value="'+esc(cfg.api_endpoint||(models.length?models[0].info.api:''))+'"></div>';

  html+='<div><label>API Key</label><div class="key-wrapper"><input id="'+fnKey+'_apikey" type="password" placeholder="输入 API Key" data-original="'+(cfg.api_key_masked?'masked':'')+'"><span class="key-toggle" onclick="toggleKeyEdit(\''+fnKey+'_apikey\')">编辑</span></div></div>';

  var defaultDims=1024;
  if(models.length&&models[0].info.dims){defaultDims=models[0].info.dims[0]}
  if(cfg.extra_params&&cfg.extra_params.dimensions){defaultDims=parseInt(cfg.extra_params.dimensions)}

  html+='<div><label>Dimensions</label><input id="'+fnKey+'_dimensions" type="number" value="'+defaultDims+'"></div>';
  html+='<div class="form-full"><label class="inline-toggle"><input type="checkbox" id="'+fnKey+'_enabled" '+(cfg.enabled?'checked':'')+'>启用</label></div>';
  html+='</div>';

  html+='<div class="section-actions">';
  html+='<button class="btn btn-success btn-sm" onclick="testModelConfig(\''+fnKey+'\')">测试</button>';
  html+='<button class="btn btn-primary btn-sm" onclick="saveModelConfig(\''+fnKey+'\')">保存</button>';
  html+='</div>';
  html+='</div>';
  return html;
}

function renderASRSection(){
  var fnKey='asr';
  var cfg=modelConfigs[fnKey]||{};
  var providers=getProvidersForType('asr');
  var html='<div class="card" id="card_'+fnKey+'">';
  html+='<div class="card-header"><span>ASR 语音识别</span>';
  html+='<span class="badge '+(cfg.enabled?'bg-green':'bg-red')+'">'+(cfg.enabled?'已启用':'未启用')+'</span></div>';
  html+='<div class="form-grid">';

  html+='<div><label>Provider</label><select id="'+fnKey+'_provider" onchange="onProviderChangeASR()">';
  providers.forEach(function(p){html+='<option value="'+p.key+'"'+(cfg.provider===p.key?' selected':'')+'>'+esc(p.name)+'</option>'});
  html+='</select></div>';

  var models=getModelsForProviderAndType(cfg.provider||providers[0].key,'asr');
  html+='<div><label>Model</label><select id="'+fnKey+'_model" onchange="onModelChangeASR()">';
  models.forEach(function(m){html+='<option value="'+m.key+'"'+(cfg.model_id===m.key?' selected':'')+'>'+esc(m.info.name)+'</option>'});
  html+='</select></div>';

  var epReadOnlyA='';
  var regPA=MODEL_REGISTRY[cfg.provider||providers[0].key];
  if(regPA&&!regPA.editable_endpoint){epReadOnlyA=' readonly'}
  html+='<div class="form-full"><label>API Endpoint</label><input id="asr_endpoint"'+epReadOnlyA+' value="'+esc(cfg.api_endpoint||(models.length?models[0].info.api:''))+'"></div>';

  var asrEP=cfg.extra_params||{};
  html+='<div><label>APP ID</label><input id="'+fnKey+'_appid" value="'+esc(asrEP.app_id||'')+'"></div>';

  html+='<div><label>Access Token (API Key)</label><div class="key-wrapper"><input id="'+fnKey+'_apikey" type="password" placeholder="输入 Access Token" data-original="'+(cfg.api_key_masked?'masked':'')+'"><span class="key-toggle" onclick="toggleKeyEdit(\''+fnKey+'_apikey\')">编辑</span></div></div>';

  html+='<div><label>Language</label><input id="'+fnKey+'_language" value="'+esc(asrEP.language||'zh')+'"></div>';
  html+='<div><label>Max Wait (秒)</label><input id="'+fnKey+'_maxwait" type="number" value="'+(asrEP.max_wait_seconds||300)+'"></div>';
  html+='<div><label>Poll Interval (秒)</label><input id="'+fnKey+'_pollinterval" type="number" value="'+(asrEP.poll_interval_seconds||5)+'"></div>';
  html+='<div class="form-full"><label class="inline-toggle"><input type="checkbox" id="'+fnKey+'_enabled" '+(cfg.enabled?'checked':'')+'>启用</label></div>';
  html+='</div>';

  html+='<div class="section-actions">';
  html+='<button class="btn btn-success btn-sm" onclick="testModelConfig(\''+fnKey+'\')">测试</button>';
  html+='<button class="btn btn-primary btn-sm" onclick="saveModelConfig(\''+fnKey+'\')">保存</button>';
  html+='</div>';
  html+='</div>';
  return html;
}

function onProviderChangeLLM(fnKey){
  var providerKey=document.getElementById(fnKey+'_provider').value;
  var models=getModelsForProviderAndType(providerKey,'llm');
  var sel=document.getElementById(fnKey+'_model');
  sel.innerHTML='';
  models.forEach(function(m){
    sel.innerHTML+='<option value="'+m.key+'">'+esc(m.info.name)+'</option>';
  });
  if(models.length){
    document.getElementById(fnKey+'_endpoint').value=models[0].info.api;
    document.getElementById(fnKey+'_maxtokens').value=models[0].info.maxTokens||4096;
  }
}

function onModelChangeLLM(fnKey){
  var providerKey=document.getElementById(fnKey+'_provider').value;
  var modelKey=document.getElementById(fnKey+'_model').value;
  var reg=MODEL_REGISTRY[providerKey];
  if(reg&&reg.models[modelKey]){
    document.getElementById(fnKey+'_endpoint').value=reg.models[modelKey].api;
    document.getElementById(fnKey+'_maxtokens').value=reg.models[modelKey].maxTokens||4096;
  }
}

function onBackupProviderChange(fnKey){
  var providerKey=document.getElementById(fnKey+'_backup_provider').value;
  var sel=document.getElementById(fnKey+'_backup_model');
  sel.innerHTML='';
  if(!providerKey){
    sel.innerHTML='<option value="">-- 先选择 Provider --</option>';
    document.getElementById(fnKey+'_backup_endpoint').value='';
    return;
  }
  var models=getModelsForProviderAndType(providerKey,'llm');
  models.forEach(function(m){
    sel.innerHTML+='<option value="'+m.key+'">'+esc(m.info.name)+'</option>';
  });
  if(models.length){
    document.getElementById(fnKey+'_backup_endpoint').value=models[0].info.api;
  }
}

function onBackupModelChange(fnKey){
  var providerKey=document.getElementById(fnKey+'_backup_provider').value;
  var modelKey=document.getElementById(fnKey+'_backup_model').value;
  var reg=MODEL_REGISTRY[providerKey];
  if(reg&&reg.models&&reg.models[modelKey]){
    document.getElementById(fnKey+'_backup_endpoint').value=reg.models[modelKey].api;
  }
}

function onProviderChangeEmbedding(){
  var providerKey=document.getElementById('embedding_provider').value;
  var models=getModelsForProviderAndType(providerKey,'embedding');
  var sel=document.getElementById('embedding_model');
  sel.innerHTML='';
  models.forEach(function(m){
    sel.innerHTML+='<option value="'+m.key+'">'+esc(m.info.name)+'</option>';
  });
  if(models.length){
    document.getElementById('embedding_endpoint').value=models[0].info.api;
    if(models[0].info.dims){
      document.getElementById('embedding_dimensions').value=models[0].info.dims[0];
    }
  }
}

function onModelChangeEmbedding(){
  var providerKey=document.getElementById('embedding_provider').value;
  var modelKey=document.getElementById('embedding_model').value;
  var reg=MODEL_REGISTRY[providerKey];
  if(reg&&reg.models[modelKey]){
    document.getElementById('embedding_endpoint').value=reg.models[modelKey].api;
    if(reg.models[modelKey].dims){
      document.getElementById('embedding_dimensions').value=reg.models[modelKey].dims[0];
    }
  }
}

function onProviderChangeASR(){
  var providerKey=document.getElementById('asr_provider').value;
  var models=getModelsForProviderAndType(providerKey,'asr');
  var sel=document.getElementById('asr_model');
  sel.innerHTML='';
  models.forEach(function(m){
    sel.innerHTML+='<option value="'+m.key+'">'+esc(m.info.name)+'</option>';
  });
  if(models.length){
    document.getElementById('asr_endpoint').value=models[0].info.api;
  }
}

function onModelChangeASR(){
  var providerKey=document.getElementById('asr_provider').value;
  var modelKey=document.getElementById('asr_model').value;
  var reg=MODEL_REGISTRY[providerKey];
  if(reg&&reg.models&&reg.models[modelKey]){
    document.getElementById('asr_endpoint').value=reg.models[modelKey].api;
  }
}

function triggerProviderChanges(){
  ['llm_extract','llm_plan'].forEach(function(fnKey){
    onProviderChangeLLM(fnKey);
    var bkProv=document.getElementById(fnKey+'_backup_provider');
    if(bkProv&&bkProv.value){onBackupProviderChange(fnKey)}
  });
  var embProv=document.getElementById('embedding_provider');
  if(embProv)onProviderChangeEmbedding();
  var asrProv=document.getElementById('asr_provider');
  if(asrProv)onProviderChangeASR();

  // Restore saved values after triggering
  ['llm_extract','llm_plan'].forEach(function(fnKey){
    var cfg=modelConfigs[fnKey]||{};
    if(cfg.model_id){
      var sel=document.getElementById(fnKey+'_model');
      if(sel){var opts=sel.querySelectorAll('option');for(var i=0;i<opts.length;i++){if(opts[i].value===cfg.model_id){sel.selectedIndex=i;onModelChangeLLM(fnKey);break}}}
    }
    if(cfg.max_tokens){var mt=document.getElementById(fnKey+'_maxtokens');if(mt)mt.value=cfg.max_tokens}
  });
  var embCfg=modelConfigs.embedding||{};
  if(embCfg.model_id){
    var esel=document.getElementById('embedding_model');
    if(esel){var eopts=esel.querySelectorAll('option');for(var i=0;i<eopts.length;i++){if(eopts[i].value===embCfg.model_id){esel.selectedIndex=i;onModelChangeEmbedding();break}}}
  }
  if(embCfg.dimensions){var ed=document.getElementById('embedding_dimensions');if(ed)ed.value=embCfg.dimensions}
  var asrCfg=modelConfigs.asr||{};
  if(asrCfg.model_id){
    var asel=document.getElementById('asr_model');
    if(asel){var aopts=asel.querySelectorAll('option');for(var i=0;i<aopts.length;i++){if(aopts[i].value===asrCfg.model_id){asel.selectedIndex=i;onModelChangeASR();break}}}
  }
}

function saveModelConfig(fnKey){
  var payload={function_key:fnKey};
  payload.provider=document.getElementById(fnKey+'_provider').value;
  payload.model_id=document.getElementById(fnKey+'_model').value;
  payload.api_endpoint=document.getElementById(fnKey+'_endpoint').value;
  payload.enabled=document.getElementById(fnKey+'_enabled').checked;

  var apiKeyEl=document.getElementById(fnKey+'_apikey');
  if(apiKeyEl.dataset.edited==='true'||apiKeyEl.value){
    payload.api_key=apiKeyEl.value;
  }

  if(fnKey==='llm_extract'||fnKey==='llm_plan'){
    payload.max_tokens=parseInt(document.getElementById(fnKey+'_maxtokens').value)||4096;
    var bkProv=document.getElementById(fnKey+'_backup_provider').value;
    if(bkProv){
      payload.backup_provider=bkProv;
      payload.backup_model_id=document.getElementById(fnKey+'_backup_model').value;
      payload.backup_api_endpoint=document.getElementById(fnKey+'_backup_endpoint').value;
      var bkKeyEl=document.getElementById(fnKey+'_backup_apikey');
      if(bkKeyEl.dataset.edited==='true'||bkKeyEl.value){
        payload.backup_api_key=bkKeyEl.value;
      }
    }
  }else if(fnKey==='embedding'){
    var ep={};
    ep.dimensions=parseInt(document.getElementById('embedding_dimensions').value)||1024;
    payload.extra_params=JSON.stringify(ep);
  }else if(fnKey==='asr'){
    var ep={};
    ep.app_id=document.getElementById('asr_appid').value;
    ep.language=document.getElementById('asr_language').value||'zh';
    ep.max_wait_seconds=parseInt(document.getElementById('asr_maxwait').value)||300;
    ep.poll_interval_seconds=parseInt(document.getElementById('asr_pollinterval').value)||5;
    payload.extra_params=JSON.stringify(ep);
  }

  fetch('/admin/model-configs/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code===0){showToast('已保存','success');loadModelConfigs()}
    else{showToast('保存失败: '+(d.message||''),'error')}
  }).catch(function(){showToast('请求失败','error')});
}

function testModelConfig(fnKey){
  var btn=document.querySelector('#card_'+fnKey+' .btn-success');
  if(btn){btn.disabled=true;btn.textContent='测试中...'}
  showToast('正在测试 '+fnKey+'...','success');
  fetch('/admin/model-configs/test',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({function_key:fnKey})})
  .then(function(r){return r.json()}).then(function(d){
    if(btn){btn.disabled=false;btn.textContent='测试'}
    if(d.code!==0){showToast('测试失败: '+(d.message||''),'error');return}
    var r2=d.data;
    if(r2&&r2.status==='ok'){
      showToast(fnKey+' 测试通过 ('+r2.latency_ms+'ms)','success');
    }else{
      showToast(fnKey+' 测试失败: '+(r2?r2.error:'未知错误'),'error');
    }
  }).catch(function(){
    if(btn){btn.disabled=false;btn.textContent='测试'}
    showToast('请求失败','error');
  });
}

function toggleCollapsible(el){
  el.classList.toggle('open');
  var body=el.nextElementSibling;
  if(body)body.classList.toggle('open');
}

function bindCollapsibles(){
  // Auto handled by onclick
}

function toggleKeyEdit(inputId){
  var el=document.getElementById(inputId);
  if(!el)return;
  el.dataset.edited='true';
  el.value='';
  el.type='text';
  el.focus();
  el.addEventListener('blur',function(){
    if(!el.value)el.type='password';
  },{once:true});
}

function bindKeyToggles(){
  // Auto handled by onclick
}

// ==================== Usage Tab ====================

function loadUsage(){
  fetch('/admin/usage').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    var data=d.data||[];
    var h='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">近30天用量统计</h3>';

    if(data.length===0){
      h+='<div style="text-align:center;color:#9CA3AF;padding:40px">暂无用量数据</div>';
    }else{
      var dateSet={},providerSet={};
      data.forEach(function(u){dateSet[u.date]=true;providerSet[u.provider_name]=true});
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
          var u=data.find(function(x){return x.date===dt&&x.provider_name===p});
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

// ==================== Connectivity Tab ====================

var testResults=[];

function loadConnectivity(){
  var h='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">连通性测试</h3>';
  h+='<p style="font-size:13px;color:#6B7280;margin-bottom:12px">测试各功能点及已启用 Provider 的连通性。</p>';
  h+='<div style="display:flex;gap:8px;margin-bottom:16px;flex-wrap:wrap">';
  h+='<button class="btn btn-primary" onclick="testAllProviders()" id="testAllBtn">全部 Provider 测试</button>';
  ['llm_extract','llm_plan','embedding','asr'].forEach(function(fnKey){
    var labels={'llm_extract':'LLM 数据提取','llm_plan':'LLM 方案生成','embedding':'Embedding 向量','asr':'ASR 语音识别'};
    h+='<button class="btn btn-outline btn-sm" onclick="testModelConfigConnectivity(\''+fnKey+'\')">'+labels[fnKey]+'</button>';
  });
  h+='</div>';
  h+='<div id="testResults">';
  if(testResults.length>0){h+=renderTestResults()}
  else{h+='<div style="text-align:center;color:#9CA3AF;padding:30px">尚未进行测试</div>'}
  h+='</div></div>';
  document.getElementById('app').innerHTML=h;
}

function renderTestResults(){
  var h='<table><tr><th>目标</th><th>延迟</th><th>状态</th><th>响应预览</th></tr>';
  testResults.forEach(function(r){
    var statusBadge=r.status==='ok'?'<span class="badge bg-green">成功</span>':'<span class="badge bg-red">失败</span>';
    var label=r.provider_name?(providerLabels[r.provider_name]||r.provider_name):(r.function_key||'');
    h+='<tr>'+
      '<td><span class="badge bg-blue">'+esc(label)+'</span></td>'+
      '<td>'+(r.latency_ms||'-')+'ms</td>'+
      '<td>'+statusBadge+'</td>'+
      '<td style="font-size:12px;max-width:300px;overflow:hidden;text-overflow:ellipsis" title="'+esc(r.response_preview||r.error||'')+'">'+esc(r.response_preview||r.error||'-')+'</td>'+
      '</tr>';
  });
  h+='</table>';
  return h;
}

function testAllProviders(){
  var btn=document.getElementById('testAllBtn');
  btn.disabled=true;btn.textContent='测试中...';
  testResults=[];
  document.getElementById('testResults').innerHTML='<div style="text-align:center;padding:20px"><div class="spinner"></div>测试中...</div>';

  var promises=allProviders.filter(function(p){return p.is_enabled}).map(function(p){
    return fetch('/admin/providers/test',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({provider_name:p.provider_name})})
    .then(function(r){return r.json()}).then(function(d){
      if(d.code===0)testResults.push(d.data);
    }).catch(function(){});
  });

  Promise.all(promises).then(function(){
    btn.disabled=false;btn.textContent='全部 Provider 测试';
    document.getElementById('testResults').innerHTML=renderTestResults();
    var allOk=testResults.every(function(r){return r.status==='ok'});
    showToast(allOk?'全部测试通过':'部分测试失败',allOk?'success':'error');
  });
}

function testModelConfigConnectivity(fnKey){
  var labels={'llm_extract':'LLM 数据提取','llm_plan':'LLM 方案生成','embedding':'Embedding 向量','asr':'ASR 语音识别'};
  showToast('正在测试 '+labels[fnKey]+'...','success');
  fetch('/admin/model-configs/test',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({function_key:fnKey})})
  .then(function(r){return r.json()}).then(function(d){
    if(d.code!==0){testResults.push({function_key:fnKey,status:'fail',error:d.message});showToast('测试失败: '+(d.message||''),'error');document.getElementById('testResults').innerHTML=renderTestResults();return}
    var r2=d.data;
    if(r2){
      r2.function_key=fnKey;
      testResults.push(r2);
      document.getElementById('testResults').innerHTML=renderTestResults();
      if(r2.status==='ok'){showToast(labels[fnKey]+' 测试通过 ('+r2.latency_ms+'ms)','success')}
      else{showToast(labels[fnKey]+' 测试失败: '+(r2.error||''),'error')}
    }
  }).catch(function(e){showToast('请求失败','error')});
}

// ==================== Legacy Provider Tab ====================

var providerDefaults={
  deepseek:{endpoint:'https://api.deepseek.com/v1/chat/completions',model:'deepseek-v4-flash'},
  ali_bailian:{endpoint:'https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions',model:'qwen3.6-plus'},
  volc_ark:{endpoint:'https://ark.cn-beijing.volces.com/api/coding/v3',model:'doubao-seed-2.0-pro'},
  opencode_go:{endpoint:'http://localhost:11434/v1/chat/completions',model:'opencode-go'}
};

function loadProviders(){
  fetch('/admin/providers').then(function(r){return r.json()}).then(function(d){
    if(d.code!==0)throw new Error(d.message||'error');
    allProviders=d.data||[];
    renderProviders();
  }).catch(function(e){
    document.getElementById('app').innerHTML='<div class="card"><h3>加载失败</h3><p>'+esc(e.message)+'</p></div>';
  });
}

function renderProviders(){
  var h='<div class="card"><h3 style="font-size:15px;margin-bottom:12px">已配置 Provider <span class="badge bg-yellow" style="font-size:10px;margin-left:8px">Legacy</span></h3>';
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
  if(!p)return;
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
</script>
</body>
</html>`
