# AI社保智筹 (AI Social Insurance Planner)

为 AI 时代灵活就业人员提供社保定制化筹划的智能平台。

## 项目结构

```
nsi-platform/
├── services/
│   ├── api-server/          # 业务核心 API (Go/Gin)
│   ├── policy-crawler/      # 政策采集与验证 (Go)
│   └── actuarial-engine/    # 精算引擎 (Go/gRPC)
├── frontend/
│   ├── ios/                 # SwiftUI 原生 iOS
│   ├── android/             # Jetpack Compose 原生 Android
│   ├── weapp/               # 微信小程序
│   └── alipay/              # 支付宝小程序
├── shared/                  # 共享库 (数据模型、配置、工具)
├── infra/                   # 基础设施配置
└── scripts/                 # 工具脚本
```

## 环境要求

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 18
- Redis 7

## 快速开始

```bash
# 启动基础设施 (PostgreSQL, Redis, MinIO)
docker compose -f infra/docker-compose.infra.yml up -d

# 启动所有服务
docker compose up -d
```

## 文档

详见 `docs/` 目录。
