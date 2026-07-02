# 移动端前�?MVP 实现设计 V1.0.0

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

## 1. 概述

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

本文档定�?AI社保智筹 MVP 移动端前端的实现方案。PRD V1.2.1 定义�?10+ 页面，本设计将其按平�?+ 开发顺序落地�?
## 2. 跨平台策�?
| 平台 | 技术栈 | 状�?|
|------|--------|------|
| 微信小程�?| WXML + WXSS + JS (原生) | 已有 API 服务�?scaffold |
| iOS | SwiftUI + NSIAPI SDK | 已有 Client.swift + NSIAPI SDK scaffold |
| Android | Jetpack Compose + Kotlin | 已有 Client.kt scaffold |
| 支付宝小程序 | AXML + ACSS + JS (原生) | 已有 API 服务�?scaffold |

各平台独立项目，无共享组件库。开发顺序：微信 �?iOS �?Android �?支付宝�?
## 3. 页面清单（所有平台一致）

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

| ID | 页面 | 核心功能 | API 依赖 |
|----|------|---------|----------|
| P1 | 启动/授权�?| 品牌展示、隐私政策、用户协议同意、一键登�?onboarding)、地理位置授�?| - |
| P2 | 首页 | 城市显示、政策速览、搜索框(P2)、开始筹划按�?| `GET /v1/policies` |
| P3 | 城市选择�?| 5 �?MVP 城市列表 | - |
| P4 | 信息采集�?| 分步表单（基本→就业→家庭），进度条 | `PUT /v1/profile` |
| P5 | 加载�?| 生成动画 + 进度文字 | `POST /v1/plans/generate` |
| P6 | 方案预览�?| 模糊数据展示、解锁按�?| - |
| P7 | 方案详情�?| 方案对比表格、现金流图表、行动清�?| `GET /v1/plans/{id}` |
| P8(后续) | 合规性引导页 | 身份认定条件动态渲染、材料清�?| `GET /v1/policies` |

�?MVP（Phase 2）页面：登录注册页（使用外部网关 mock）、设置页（已有基线）、用户反馈管理�?
## 4. 微信小程序详细设�?
### 4.1 项目结构

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```
frontend/weapp/
├── app.js                 # 全局逻辑（onLaunch 定位、登录态检查）
├── app.json               # 全局配置（页面注册、TabBar 如有�?├── app.wxss               # 全局样式（主题色、字体）
├── services/
�?  └── api.js             # API 客户端（已有 scaffold�?├── utils/
�?  ├── auth.js            # 登录、地理位置授权工�?�?  └── constants.js       # 常量（城市列表、政策类型映射）
├── components/
�?  ├── auth-modal/        # 授权弹窗组件（隐私协�?+ 一键登录）
�?  ├── progress-bar/      # 分步进度条组�?�?  ├── policy-card/       # 政策速览卡片组件
�?  ├── scheme-card/       # 方案对比卡片组件
�?  └── blurred-view/      # 模糊遮罩组件
└── pages/
    ├── login/             # P1 启动/授权�?    ├── index/             # P2 首页
    ├── city-picker/       # P3 城市选择�?    ├── profile/           # P4 信息采集（分步表单）
    ├── loading/           # P5 方案生成�?    ├── preview/           # P6 方案预览
    └── plan/              # P7 方案详情
```

### 4.2 页面间导航流

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

```
login �?index �?city-picker �?index
     �?index �?profile �?loading �?preview �?plan
```

使用 wx.navigateTo / wx.redirectTo 进行页面跳转。数据传递通过全局数据（app.globalData）或 URL 参数�?
### 4.3 关键组件设计

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**授权弹窗 (auth-modal):**
- 首次启动显示：品�?Logo、隐私政�?checkbox、用户协�?checkbox�?微信一键登�?按钮
- 拒绝授权 �?仅浏览模式（部分功能受限�?- 通过 `wx.getUserProfile` 获取用户信息

**分步表单 (profile 页面):**
- 步骤 1：基本信息（年龄、性别、户籍地、居住地�?- 步骤 2：就业信息（就业状态、社保年限、失业登记日期）
- 步骤 3：家庭信息（子女情况、技能证书）
- 每步 �?4 个问题，带进度条指示�?- `wx.chooseLocation` 辅助城市选择
- 提交后调�?`PUT /v1/profile`，成功则跳转�?loading �?
**方案对比卡片 (scheme-card):**
- 显示方案名称、缴费基数、月缴费额、预计养老金
- 卡片选中态切换（高亮边框�?- 支持横向滑动对比

**模糊遮罩 (blurred-view):**
- `backdrop-filter: blur(10px)` CSS 实现毛玻璃效�?- 覆盖核心数据区域（预计年补贴额、月节省额）
- "解锁完整报告"按钮触发支付弹窗（MVP 中显示提示，跳转�?plan 页）

### 4.4 数据�?
```
Page.onLoad �?检查登录�?�?获取地理位置 �?初始化数�?    �?用户交互 �?调用 services/api.js �?请求 api-server
    �?API 响应 �?Page.setData �?渲染 UI
    �?错误处理 �?wx.showToast / 错误页面
```

### 4.5 错误处理

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- 网络错误：wx.showToast + 重试按钮
- API 错误：根�?code 显示中文提示
- 输入校验：前端必填校�?+ 格式校验（年�?16-70、手机号格式等）
- 授权拒绝：降级到仅浏览模�?
### 4.6 后续平台差异要点

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

**iOS (SwiftUI):**
- 使用 `NSIAPI.NSIClient` 替代 api.js
- 地理位置：`CLLocationManager`
- 授权：`SignInWithAppleButton` 或只读模�?- 模糊效果：`.blur(radius:)` modifier
- TabView / NavigationStack 导航

**Android (Jetpack Compose):**
- 使用 `NSIClient` 替代 api.js
- 地理位置：`FusedLocationProviderClient`
- 授权：运行时权限请求
- 模糊效果：`RenderScript` �?`BlurMaskFilter`
- NavHost + Scaffold 导航

**支付宝小程序:**
- 使用 `my.request`（已�?scaffold�?- 授权：`my.getAuthCode`
- 页面结构与微信一致，适配支付�?API
- 组件样式�?ACSS 规范

## 5. 验收标准

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- 每个平台的每个页面能正常运行
- API 调用正确（通过 mock/integration 测试�?- 导航流程完整：login �?index �?profile �?loading �?preview �?plan
- 错误状态有 UI 反馈
- 所有平台的页面布局、功能一�?- 微信小程序可预览（开发者工具）
- iOS/Android 可构建（Xcode / Android Studio�?
## 6. 后续迭代

| **�汾��** | V1.0.0 |
| **״̬** | ����Ч |
| **��������** | 2026-06-15 |

- Phase 2: 合规性引导页、支付集成、PDF 报告
- 无障碍适配（字体调整、语音播报）
- E2E 测试
