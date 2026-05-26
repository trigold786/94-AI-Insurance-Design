# 移动端前端 MVP 实现设计 V1.0.0

## 1. 概述

本文档定义 AI社保智筹 MVP 移动端前端的实现方案。PRD V1.2.1 定义了 10+ 页面，本设计将其按平台 + 开发顺序落地。

## 2. 跨平台策略

| 平台 | 技术栈 | 状态 |
|------|--------|------|
| 微信小程序 | WXML + WXSS + JS (原生) | 已有 API 服务层 scaffold |
| iOS | SwiftUI + NSIAPI SDK | 已有 Client.swift + NSIAPI SDK scaffold |
| Android | Jetpack Compose + Kotlin | 已有 Client.kt scaffold |
| 支付宝小程序 | AXML + ACSS + JS (原生) | 已有 API 服务层 scaffold |

各平台独立项目，无共享组件库。开发顺序：微信 → iOS → Android → 支付宝。

## 3. 页面清单（所有平台一致）

| ID | 页面 | 核心功能 | API 依赖 |
|----|------|---------|----------|
| P1 | 启动/授权页 | 品牌展示、隐私政策、用户协议同意、一键登录(onboarding)、地理位置授权 | - |
| P2 | 首页 | 城市显示、政策速览、搜索框(P2)、开始筹划按钮 | `GET /v1/policies` |
| P3 | 城市选择页 | 5 个 MVP 城市列表 | - |
| P4 | 信息采集页 | 分步表单（基本→就业→家庭），进度条 | `PUT /v1/profile` |
| P5 | 加载页 | 生成动画 + 进度文字 | `POST /v1/plans/generate` |
| P6 | 方案预览页 | 模糊数据展示、解锁按钮 | - |
| P7 | 方案详情页 | 方案对比表格、现金流图表、行动清单 | `GET /v1/plans/{id}` |
| P8(后续) | 合规性引导页 | 身份认定条件动态渲染、材料清单 | `GET /v1/policies` |

非 MVP（Phase 2）页面：登录注册页（使用外部网关 mock）、设置页（已有基线）、用户反馈管理。

## 4. 微信小程序详细设计

### 4.1 项目结构

```
frontend/weapp/
├── app.js                 # 全局逻辑（onLaunch 定位、登录态检查）
├── app.json               # 全局配置（页面注册、TabBar 如有）
├── app.wxss               # 全局样式（主题色、字体）
├── services/
│   └── api.js             # API 客户端（已有 scaffold）
├── utils/
│   ├── auth.js            # 登录、地理位置授权工具
│   └── constants.js       # 常量（城市列表、政策类型映射）
├── components/
│   ├── auth-modal/        # 授权弹窗组件（隐私协议 + 一键登录）
│   ├── progress-bar/      # 分步进度条组件
│   ├── policy-card/       # 政策速览卡片组件
│   ├── scheme-card/       # 方案对比卡片组件
│   └── blurred-view/      # 模糊遮罩组件
└── pages/
    ├── login/             # P1 启动/授权页
    ├── index/             # P2 首页
    ├── city-picker/       # P3 城市选择页
    ├── profile/           # P4 信息采集（分步表单）
    ├── loading/           # P5 方案生成中
    ├── preview/           # P6 方案预览
    └── plan/              # P7 方案详情
```

### 4.2 页面间导航流

```
login → index → city-picker → index
     → index → profile → loading → preview → plan
```

使用 wx.navigateTo / wx.redirectTo 进行页面跳转。数据传递通过全局数据（app.globalData）或 URL 参数。

### 4.3 关键组件设计

**授权弹窗 (auth-modal):**
- 首次启动显示：品牌 Logo、隐私政策 checkbox、用户协议 checkbox、"微信一键登录"按钮
- 拒绝授权 → 仅浏览模式（部分功能受限）
- 通过 `wx.getUserProfile` 获取用户信息

**分步表单 (profile 页面):**
- 步骤 1：基本信息（年龄、性别、户籍地、居住地）
- 步骤 2：就业信息（就业状态、社保年限、失业登记日期）
- 步骤 3：家庭信息（子女情况、技能证书）
- 每步 ≤ 4 个问题，带进度条指示器
- `wx.chooseLocation` 辅助城市选择
- 提交后调用 `PUT /v1/profile`，成功则跳转到 loading 页

**方案对比卡片 (scheme-card):**
- 显示方案名称、缴费基数、月缴费额、预计养老金
- 卡片选中态切换（高亮边框）
- 支持横向滑动对比

**模糊遮罩 (blurred-view):**
- `backdrop-filter: blur(10px)` CSS 实现毛玻璃效果
- 覆盖核心数据区域（预计年补贴额、月节省额）
- "解锁完整报告"按钮触发支付弹窗（MVP 中显示提示，跳转到 plan 页）

### 4.4 数据流

```
Page.onLoad → 检查登录态 → 获取地理位置 → 初始化数据
    ↓
用户交互 → 调用 services/api.js → 请求 api-server
    ↓
API 响应 → Page.setData → 渲染 UI
    ↓
错误处理 → wx.showToast / 错误页面
```

### 4.5 错误处理

- 网络错误：wx.showToast + 重试按钮
- API 错误：根据 code 显示中文提示
- 输入校验：前端必填校验 + 格式校验（年龄 16-70、手机号格式等）
- 授权拒绝：降级到仅浏览模式

### 4.6 后续平台差异要点

**iOS (SwiftUI):**
- 使用 `NSIAPI.NSIClient` 替代 api.js
- 地理位置：`CLLocationManager`
- 授权：`SignInWithAppleButton` 或只读模式
- 模糊效果：`.blur(radius:)` modifier
- TabView / NavigationStack 导航

**Android (Jetpack Compose):**
- 使用 `NSIClient` 替代 api.js
- 地理位置：`FusedLocationProviderClient`
- 授权：运行时权限请求
- 模糊效果：`RenderScript` 或 `BlurMaskFilter`
- NavHost + Scaffold 导航

**支付宝小程序:**
- 使用 `my.request`（已有 scaffold）
- 授权：`my.getAuthCode`
- 页面结构与微信一致，适配支付宝 API
- 组件样式按 ACSS 规范

## 5. 验收标准

- 每个平台的每个页面能正常运行
- API 调用正确（通过 mock/integration 测试）
- 导航流程完整：login → index → profile → loading → preview → plan
- 错误状态有 UI 反馈
- 所有平台的页面布局、功能一致
- 微信小程序可预览（开发者工具）
- iOS/Android 可构建（Xcode / Android Studio）

## 6. 后续迭代

- Phase 2: 合规性引导页、支付集成、PDF 报告
- 无障碍适配（字体调整、语音播报）
- E2E 测试
