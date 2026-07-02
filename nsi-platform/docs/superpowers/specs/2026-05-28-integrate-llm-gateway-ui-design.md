# 集成LLM Gateway UI到Policy Crawler Admin设计

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

## 目标
把llm-gateway的所有管理功能（模型配置、用量统计、连通性测试、Provider配置）集成到policy-crawler�?9403端口管理后台�?
## 方案
**proxy + iframe集成**
- policy-crawler代理所有llm-gateway的admin请求
- iframe用相对路径，整个集成�?9403端口�?- 复用现有Basic Auth（两者账号密码相同）

## 架构
```
┌────────────────────────────────────────────────────�?�?        用户浏览�?(访问 http://localhost:39403)  �?�?                                                   �?�? ┌─────────────────────────────────────────────�?�?�? �? Policy Crawler Admin UI (nav items)       �?�?�? └─────────────────┬───────────────────────────�?�?└────────────────────┼──────────────────────────────�?                     �?                     �?┌────────────────────────────────────────────────────�?�? Policy Crawler Service (39403)                   �?�?                                                   �?�? ┌─────────────────────────────────────────────�?�?�? �? /admin/... (现有管理后台)                   �?�?�? └─────────────────────────────────────────────�?�?�? ┌─────────────────────────────────────────────�?�?�? �? /llm-gateway/* �?proxy �?llm-gateway:39404�?�?�? �? (转发请求，保留Basic Auth)                 �?�?�? └─────────────────────────────────────────────�?�?└────────────────────┬──────────────────────────────�?                     �?                     �?┌────────────────────────────────────────────────────�?�? LLM Gateway Service (39404, 容器�? 94-nsip-llm-gateway-1) �?└────────────────────────────────────────────────────�?```

## 详细设计

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

### 1. Proxy Handler
在policy-crawler内部新增一个proxy handler:
- 路径：`/llm-gateway/*`
- 转发目的地：`http://llm-gateway:39404/` (用容器内部名�?
- 处理：把请求原封不动转发（headers、body、path、query），把响应原封不动返�?- 认证：保留Basic Auth header，因为llm-gateway和policy-crawler用相同账号密�?
### 2. Policy Crawler导航更新
在policy-crawler的`admin_page.go`的navItems里加一个新项：
- ID: `llm-gateway`
- Label: `模型管理`
- 内容：iframe加载 `/llm-gateway/admin`

### 3. 路由配置
在policy-crawler的`main.go`里加新路由：
- `mux.Handle("/llm-gateway/", adminAuth(proxyHandler))` (用policy-crawler自己的auth)

### 4. Proxy Handler实现
实现一个简单的HTTP reverse proxy:
```go
func llmGatewayProxy(w http.ResponseWriter, r *http.Request) {
    // 构造llm-gateway内部地址
    // 去掉前缀/llm-gateway
    // proxy请求到llm-gateway
    // 保留Basic Auth header
    // 原样返回响应
}
```

## 注意事项
- llm-gateway内部地址在docker network里是`94-nsip-llm-gateway-1:39404`还是直接用服务名�?- iframe需要加载`/llm-gateway/admin`，这样所有相对路径的请求都走proxy
- policy-crawler和llm-gateway的ADMIN_USERNAME/ADMIN_PASSWORD相同，proxy直接传Basic Auth即可

## 实现步骤
1. 在policy-crawler里实现proxy handler
2. 在policy-crawler main.go里加/llm-gateway路由
3. 在policy-crawler admin_page.go导航栏加新项
4. 测试集成
