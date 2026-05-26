package admin

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
)

type EmbeddingSearcher interface {
	SearchByText(ctx context.Context, query string, threshold float64, limit int, filter *embeddings.SearchFilter) ([]embeddings.SimilarResult, error)
}

func AdminSearchHandler(searcher EmbeddingSearcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		region := strings.TrimSpace(r.URL.Query().Get("region"))
		policyType := strings.TrimSpace(r.URL.Query().Get("type"))
		limitStr := r.URL.Query().Get("limit")

		limit := 20
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
			limit = n
		}

		if q == "" {
			respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": []embeddings.SimilarResult{}})
			return
		}

		var filter *embeddings.SearchFilter
		if region != "" || policyType != "" {
			filter = &embeddings.SearchFilter{RegionCode: region, PolicyType: policyType}
		}
		results, err := searcher.SearchByText(r.Context(), q, 0, limit, filter)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{"code": -1, "msg": "search failed"})
			return
		}
		if results == nil {
			results = []embeddings.SimilarResult{}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": results})
	})
}

func HTMLSearchPage() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>语义搜索 - AI社保智筹</title>
<style>
body{font-family:"Microsoft YaHei",sans-serif;margin:20px;background:#f5f5f5}
.container{max-width:1000px;margin:0 auto}
.search-box{background:#fff;padding:20px;border-radius:8px;box-shadow:0 2px 4px rgba(0,0,0,0.1)}
.search-box input,.search-box select{padding:8px 12px;border:1px solid #ddd;border-radius:4px;font-size:14px}
.search-box input[type="text"]{width:60%;margin-right:10px}
.search-box button{padding:8px 20px;background:#1890ff;color:#fff;border:none;border-radius:4px;cursor:pointer}
.search-box button:hover{background:#40a9ff}
.result-item{background:#fff;margin-top:12px;padding:16px;border-radius:8px;box-shadow:0 1px 3px rgba(0,0,0,0.1)}
.result-item .score{float:right;font-weight:bold}
.score-high{color:#52c41a}
.score-mid{color:#faad14}
.score-low{color:#999}
.result-item .meta{color:#666;font-size:13px;margin-top:6px}
.result-item .meta span{margin-right:16px}
</style>
</head>
<body>
<div class="container">
<h1>语义搜索</h1>
<div class="search-box">
<input type="text" id="searchInput" placeholder="输入搜索关键词，如：灵活就业补贴 北京">
<select id="regionSelect">
<option value="">全部城市</option>
<option value="110000">北京</option>
<option value="310000">上海</option>
<option value="440100">广州</option>
<option value="330100">杭州</option>
<option value="440300">深圳</option>
</select>
<select id="typeSelect">
<option value="">全部类型</option>
<option value="subsidy">补贴</option>
<option value="pension">养老</option>
<option value="medical">医疗</option>
<option value="unemployment">失业</option>
<option value="injury">工伤</option>
<option value="maternity">生育</option>
<option value="housing_fund">公积金</option>
<option value="training">培训</option>
</select>
<button onclick="search()">搜索</button>
</div>
<div id="results"></div>
</div>
<script>
function search(){
  var q=document.getElementById('searchInput').value;
  var region=document.getElementById('regionSelect').value;
  var type=document.getElementById('typeSelect').value;
  if(!q)return;
  document.getElementById('results').innerHTML='<p>搜索中...</p>';
  fetch('/admin/llm/search?q='+encodeURIComponent(q)+'&region='+region+'&type='+type)
    .then(function(r){return r.json()})
    .then(function(d){
      if(d.code!==0){document.getElementById('results').innerHTML='<p style="color:red">错误</p>';return}
      var items=d.data||[];
      if(items.length===0){document.getElementById('results').innerHTML='<p>没有找到匹配的政策</p>';return}
      var html='';
      for(var i=0;i<items.length;i++){
        var item=items[i];
        var cls=item.score>=0.8?'score-high':(item.score>=0.6?'score-mid':'score-low');
        html+='<div class="result-item">';
        html+='<div><strong>'+esc(item.policy_id)+'</strong> <span class="score '+cls+'">'+(item.score*100).toFixed(1)+'%</span></div>';
        html+='<div class="meta">';
        html+='<span>ID: '+esc(item.claim_id)+'</span>';
        html+='<span>类型: '+item.policy_type+'</span>';
        html+='<span>城市: '+item.region_code+'</span>';
        html+='<span>来源: '+(item.source_name||'-')+'</span>';
        html+='<span>状态: '+item.status+'</span>';
        html+='<a href="/admin#claims" style="color:#1A56DB;font-size:12px;text-decoration:none" target="_parent">审查 &rarr;</a>';
        html+='</div></div>';
      }
      document.getElementById('results').innerHTML=html;
    })
    .catch(function(e){document.getElementById('results').innerHTML='<p style="color:red">请求失败: '+e.message+'</p>'});
}
function esc(s){if(typeof s!=='string')return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML}
</script>
</body>
</html>`
}
