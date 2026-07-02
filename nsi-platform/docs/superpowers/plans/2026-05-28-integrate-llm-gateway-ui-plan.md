# 集成LLM Gateway UI到Policy Crawler Admin实现计划

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把llm-gateway的所有管理功能集成到policy-crawler�?9403端口管理后台，使用proxy + iframe方案�?
**Architecture:** 
1. policy-crawler新增proxy handler，转发`/llm-gateway/*`到llm-gateway内部地址
2. 在policy-crawler的admin导航栏新增“模型管理”项
3. 点击该项显示iframe加载`/llm-gateway/admin`

**Tech Stack:** Go, http reverse proxy, Docker networking, HTML/JS iframe integration

---

### Task 1: 实现LLM Gateway Proxy Handler

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/admin/proxy.go`
- Modify: `nsi-platform/services/policy-crawler/cmd/main.go:143-175`

- [ ] **Step 1: 创建proxy.go文件**

在`services/policy-crawler/internal/admin/proxy.go`里写:
```go
package admin

import (
	"io"
	"net/http"
	"os"
)

func LLMGatewayProxy(w http.ResponseWriter, r *http.Request) {
	llmGatewayURL := os.Getenv("LLM_GATEWAY_URL")
	if llmGatewayURL == "" {
		llmGatewayURL = "http://llm-gateway:39404"
	}
	// 去掉/llm-gateway前缀
	path := r.URL.Path[len("/llm-gateway"):]
	targetURL := llmGatewayURL + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 复制header，保留Authorization (Basic Auth)
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应header
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
```

- [ ] **Step 2: 在main.go里添加路�?*

在`services/policy-crawler/cmd/main.go`的HTTP路由配置部分(�?43-175�?, 在其他admin路由后面加上:
```go
	mux.Handle("/llm-gateway/", adminAuth(admin.LLMGatewayProxy))
```

- [ ] **Step 3: 验证代码编译**

运行:
```bash
cd nsi-platform/services/policy-crawler
go build ./...
```

Expected: No errors

- [ ] **Step 4: 提交代码**
```bash
git add nsi-platform/services/policy-crawler/internal/admin/proxy.go
git add nsi-platform/services/policy-crawler/cmd/main.go
git commit -m "feat: add llm-gateway proxy handler"
```

---

### Task 2: 在Admin导航栏新�?模型管理"�?
**Files:**
- Modify: `nsi-platform/services/policy-crawler/internal/admin/admin_page.go`

- [ ] **Step 1: 找到navItems定义**

在`admin_page.go`里找到navItems:
```javascript
var navItems=[
  {id:'dashboard',label:'仪表�?},
  ...
]
```

- [ ] **Step 2: 添加新导航项**

在navItems最后加:
```javascript
,{id:'llm-gateway',label:'模型管理'}
```

- [ ] **Step 3: 添加处理逻辑**

找到处理panel切换的代码（应该有switchPanel函数），为`llm-gateway`添加:
```javascript
function switchPanel(id){
  ...
  else if(id==='llm-gateway'){
    var h=''
    h+='<div class="card">'
    h+='<iframe src="/llm-gateway/admin" style="width:100%;height:800px;border:0;"></iframe>'
    h+='</div>'
    app.innerHTML=h
  }
  ...
}
```

- [ ] **Step 4: 验证修改后的代码文件没有语法错误**

虽然是HTML/JS，确保edit操作正确�?
- [ ] **Step 5: 提交代码**
```bash
git add nsi-platform/services/policy-crawler/internal/admin/admin_page.go
git commit -m "feat: add llm-gateway nav item and iframe panel"
```

---

### Task 3: 添加LLM_GATEWAY_URL�?env�?env.example

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**Files:**
- Modify: `nsi-platform/.env`
- Modify: `nsi-platform/.env.example`

- [ ] **Step 1: 更新.env**

在`nsi-platform/.env`里加一�?
```
LLM_GATEWAY_URL=http://llm-gateway:39404
```

注意: 在docker network里，llm-gateway的容器名是`94-nsip-llm-gateway-1`，但也可以用服务名`llm-gateway`。Docker Compose会自动处理服务名解析�?
- [ ] **Step 2: 更新.env.example**

在`nsi-platform/.env.example`里加:
```
# LLM Gateway地址
LLM_GATEWAY_URL=http://localhost:39404
```

- [ ] **Step 3: 提交更改**
```bash
git add nsi-platform/.env.example
git commit -m "docs: add LLM_GATEWAY_URL to .env.example"
```

---

### Task 4: 编译、部署、测试集�?
**Files:** None (compile/deploy/test only)

- [ ] **Step 1: 重新编译policy-crawler**

```bash
cd nsi-platform/services/policy-crawler
GOOS=linux GOARCH=amd64 go build -o policy-crawler ./cmd/
cp policy-crawler bin/policy-crawler
```

- [ ] **Step 2: 重新构建policy-crawler镜像并启�?*

```bash
cd nsi-platform
docker compose build policy-crawler
docker compose up -d policy-crawler
```

- [ ] **Step 3: 验证容器正常启动**

```bash
docker compose ps
```

Expected: policy-crawler up and healthy

- [ ] **Step 4: 测试集成**

打开浏览器访问`http://localhost:39403/admin`，点导航栏里的“模型管理”，看看iframe是否能正常加载llm-gateway页面�?
测试�?
1. 模型配置页正常显�?2. 可以编辑保存配置
3. 可以测试连通�?4. 可以看用量统�?
---

### 计划自审

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- �?Spec覆盖：所有要求都有任�?- �?没有占位符：所有步骤都有明确代�?命令
- �?Type一致：没有不一致的函数/变量�?