# ÊîØ‰ªòÂÆùÂ∞èÁ®ãÂ∫è MVP ÂÆûÁé∞ËÆ°Âàí

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build complete ÊîØ‰ªòÂÆùÂ∞èÁ®ãÂ∫è MVP with 7 pages (P1-P7) + API integration

**Architecture:** Native Alipay Mini Program (AXML+ACSS+JS), 7 independent pages under `pages/`, API layer in `services/api.js`. Navigation flows login ‚Ü?index ‚Ü?city-picker ‚Ü?profile ‚Ü?loading ‚Ü?preview ‚Ü?plan. Data flows via `app.globalData`.

**Tech Stack:** AXML, ACSS, JavaScript, Alipay Mini Program SDK (my.* API), existing backend REST API. Direct translation from WeChat Mini Program with `wx.*` ‚Ü?`my.*` API mapping.

---

### Task 1: Project Scaffold + Global Config

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `frontend/alipay/app.json`
- Create: `frontend/alipay/app.js`
- Create: `frontend/alipay/app.acss`
- Create: `frontend/alipay/utils/constants.js`
- Modify: `frontend/alipay/services/api.js` (overwrite with Alipay API)

- [ ] **Step 1: Create directories**

```bash
New-Item -ItemType Directory -Path "nsi-platform/frontend/alipay/utils" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/alipay/pages/login" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/alipay/pages/index" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/alipay/pages/city-picker" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/alipay/pages/profile" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/alipay/pages/loading" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/alipay/pages/preview" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/alipay/pages/plan" -Force
```

- [ ] **Step 2: Write app.json**

```json
{
  "pages": [
    "pages/login/login",
    "pages/index/index",
    "pages/city-picker/city-picker",
    "pages/profile/profile",
    "pages/loading/loading",
    "pages/preview/preview",
    "pages/plan/plan"
  ],
  "window": {
    "defaultTitle": "AIÁ§æ‰øùÊô∫Á≠π",
    "titleBarColor": "#1A56DB",
    "backgroundColor": "#F5F7FA"
  },
  "permission": {
    "scope.userLocation": {
      "desc": "Áî®‰∫éËá™Âä®ÂÆö‰ΩçÊÇ®ÊâÄÂú®ÁöÑÂüéÂ∏Ç"
    }
  }
}
```

- [ ] **Step 3: Write app.js**

```javascript
App({
  globalData: {
    userInfo: null,
    currentCity: null,
    currentCityCode: null,
    planResult: null,
    profileData: null,
  },
  onLaunch() {
    const that = this;
    my.getSetting({
      success: (res) => {
        if (res.authSetting['scope.userInfo']) {
          my.getAuthCode({
            scopes: 'auth_user',
            success: (res) => {
              that.globalData.userInfo = { nickName: res.authCode || 'default' };
            },
          });
        }
      },
    });
  },
});
```

- [ ] **Step 4: Write app.acss** (same as WeChat app.wxss)

```css
page {
  background-color: #F5F7FA;
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', sans-serif;
  color: #333;
}
.container {
  padding: 20rpx 30rpx;
}
.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.06);
}
.btn-primary {
  background: linear-gradient(135deg, #1A56DB, #3B82F6);
  color: #fff;
  border-radius: 48rpx;
  padding: 24rpx 0;
  text-align: center;
  font-size: 32rpx;
  font-weight: 600;
}
.btn-primary:active {
  opacity: 0.85;
}
.text-muted {
  color: #9CA3AF;
  font-size: 26rpx;
}
.text-error {
  color: #EF4444;
  font-size: 26rpx;
}
```

- [ ] **Step 5: Write utils/constants.js** (same as WeChat)

```javascript
const CITIES = [
  { code: '310000', name: '‰∏äÊµ∑' },
  { code: '110000', name: 'Âåó‰∫¨' },
  { code: '440300', name: 'Ê∑±Âú≥' },
  { code: '440100', name: 'ÂπøÂ∑û' },
  { code: '330100', name: 'Êù≠Â∑û' },
];

const EMPLOYMENT_STATUS = [
  { value: 'employed', label: '‰ºÅ‰∏öÂ∞±‰∏ö' },
  { value: 'flexible', label: 'ÁÅµÊ¥ªÂ∞±‰∏ö' },
  { value: 'self_employed', label: 'Ëá™Èõá' },
  { value: 'unemployed', label: 'Â§±‰∏ö' },
];

const GENDER_OPTIONS = [
  { value: 'male', label: 'Áî? },
  { value: 'female', label: 'Â•? },
];

const POLICY_TYPES = [
  { type: 'subsidy', label: 'Ë°•Ë¥¥ÊîøÁ≠ñ' },
  { type: 'training', label: 'ÂüπËÆ≠ÊîøÁ≠ñ' },
  { type: 'pension', label: 'ÂÖªËÄÅ‰øùÈô? },
];

module.exports = { CITIES, EMPLOYMENT_STATUS, GENDER_OPTIONS, POLICY_TYPES };
```

- [ ] **Step 6: Write services/api.js** (Alipay version with my.request)

```javascript
const BASE_URL = 'http://localhost:30001';
const TOAST_DURATION = 2000;

function showError(msg) {
  my.showToast({ content: msg, type: 'none', duration: TOAST_DURATION });
}

function request(method, path, data, userID) {
  return new Promise((resolve, reject) => {
    my.request({
      url: BASE_URL + path,
      method,
      data,
      headers: {
        'x-user-id': userID || 'default',
        'Content-Type': 'application/json',
      },
      success: (res) => {
        if (res.status >= 400) {
          const msg = res.data && res.data.message ? res.data.message : 'ËØ∑Ê±ÇÂ§±Ë¥•(' + res.status + ')';
          showError(msg);
          reject(new Error(msg));
          return;
        }
        const { code, data: body } = res.data;
        if (code !== 0) {
          const msg = body && body.message ? body.message : 'ÊúçÂä°ÂºÇÂ∏∏';
          showError(msg);
          reject(new Error(msg));
          return;
        }
        resolve(body);
      },
      fail: () => {
        showError('ÁΩëÁªúËØ∑Ê±ÇÂ§±Ë¥•ÔºåËØ∑Ê£ÄÊü•ÁΩëÁªúËøûÊé?);
        reject(new Error('Network error'));
      },
    });
  });
}

function getProfile(userID) {
  return request('GET', '/v1/profile', null, userID);
}

function updateProfile(userID, data) {
  return request('PUT', '/v1/profile', data, userID);
}

function queryPolicies(userID, { regionCode, policyType, status } = {}) {
  let path = '/v1/policies';
  const params = [];
  if (regionCode) params.push('region_code=' + regionCode);
  if (policyType) params.push('policy_type=' + policyType);
  if (status) params.push('status=' + status);
  if (params.length) path += '?' + params.join('&');
  return request('GET', path, null, userID);
}

function generatePlan(userID, data) {
  return request('POST', '/v1/plans/generate', data, userID);
}

function getPlanDetail(userID, planID) {
  return request('GET', '/v1/plans/' + planID, null, userID);
}

module.exports = { getProfile, updateProfile, queryPolicies, generatePlan, getPlanDetail };
```

Note differences from WeChat: `my.request()` uses `headers` (not `header`), `my.showToast()` uses `content` (not `title`), `res.status` (not `res.statusCode`), `my.getAuthCode` for auth.

---

### Task 2: P1 Login Page (Alipay)

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `frontend/alipay/pages/login/login.js`
- Create: `frontend/alipay/pages/login/login.axml`
- Create: `frontend/alipay/pages/login/login.acss`

- [ ] **Step 1: Write login.axml** (same AXML as WeChat WXML)

```xml
<view class="container">
  <view class="header">
    <view class="logo-placeholder">
      <text class="logo-text">Á§æ‰øù</text>
    </view>
    <text class="title">AIÁ§æ‰øùÊô∫Á≠π</text>
    <text class="subtitle">AIÈ©±Âä®ÁöÑÁ§æ‰øùËßÑÂàíÂä©Êâ?/text>
  </view>

  <view class="policy-card">
    <checkbox class="policy-checkbox" checked="{{agreed}}" onChange="onAgreeChange"/>
    <text class="policy-text">ÊàëÂ∑≤ÈòÖËØªÂπ∂ÂêåÊÑ?/text>
    <text class="policy-link" onTap="onViewPrivacy">„ÄäÈöêÁßÅÊîøÁ≠ñ„Ä?/text>
    <text class="policy-text">Âí?/text>
    <text class="policy-link" onTap="onViewTerms">„ÄäÁî®Êà∑ÂçèËÆÆ„Ä?/text>
  </view>

  <button class="btn-primary {{!agreed ? 'btn-disabled' : ''}}" 
          disabled="{{!agreed}}" 
          onTap="onLogin">
    <text a:if="{{!loading}}">ÊîØ‰ªòÂÆù‰∏ÄÈîÆÁôªÂΩ?/text>
    <text a:else>ÁôªÂΩï‰∏?..</text>
  </button>

  <view class="footer">
    <text class="text-muted">ÁôªÂΩïÂç≥Ë°®Á§∫ÊÇ®ÂêåÊÑè‰∏äËø∞Êù°Ê¨æ</text>
  </view>
</view>
```

- [ ] **Step 2: Write login.acss** (same as WeChat wxss)

```css
.container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 60rpx 0;
  min-height: 100vh;
  background: linear-gradient(180deg, #EEF2FF 0%, #F5F7FA 100%);
}
.header {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 80rpx;
}
.logo-placeholder {
  width: 160rpx;
  height: 160rpx;
  border-radius: 40rpx;
  background: linear-gradient(135deg, #1A56DB, #3B82F6);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
}
.logo-text { font-size: 48rpx; font-weight: 700; color: #fff; }
.title { font-size: 48rpx; font-weight: 700; color: #1A56DB; margin-bottom: 12rpx; }
.subtitle { font-size: 28rpx; color: #6B7280; }
.policy-card {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 48rpx;
  width: 100%;
}
.policy-checkbox { margin-right: 12rpx; }
.policy-text { font-size: 24rpx; color: #6B7280; }
.policy-link { font-size: 24rpx; color: #1A56DB; text-decoration: underline; }
.btn-disabled { opacity: 0.5; }
.footer { margin-top: 48rpx; text-align: center; }
```

- [ ] **Step 3: Write login.js** (Alipay my.* APIs)

```javascript
const app = getApp();

Page({
  data: {
    agreed: false,
    loading: false,
  },
  onAgreeChange(e) {
    this.setData({ agreed: e.detail.value.length > 0 });
  },
  onViewPrivacy() {
    my.alert({
      title: 'ÈöêÁßÅÊîøÁ≠ñ',
      content: 'Êàë‰ª¨Êî∂ÈõÜÊÇ®ÁöÑ‰ø°ÊÅØÁî®‰∫éÁ§æ‰øùËßÑÂàíÊúçÂä°...',
    });
  },
  onViewTerms() {
    my.alert({
      title: 'Áî®Êà∑ÂçèËÆÆ',
      content: 'Ê¨¢Ëøé‰ΩøÁî®AIÁ§æ‰øùÊô∫Á≠π...',
    });
  },
  onLogin() {
    if (!this.data.agreed) return;
    this.setData({ loading: true });
    my.getAuthCode({
      scopes: 'auth_user',
      success: (res) => {
        app.globalData.userInfo = { nickName: res.authCode || 'default' };
        this.getLocationAndNavigate();
      },
      fail: () => {
        this.getLocationAndNavigate();
      },
    });
  },
  getLocationAndNavigate() {
    my.getLocation({
      type: 1,
      success: () => {
        const { CITIES } = require('../../utils/constants');
        app.globalData.currentCity = CITIES[0].name;
        app.globalData.currentCityCode = CITIES[0].code;
        my.redirectTo({ url: '/pages/index/index' });
      },
      fail: () => {
        my.redirectTo({ url: '/pages/index/index' });
      },
    });
  },
});
```

- [ ] **Step 4: Commit** (style: `feat(alipay): P1 login page`)

---

### Task 3: P2-P7 Pages (Alipay)

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:** Create 6 page directories with .axml, .acss, .js for each.

API mapping from WeChat to Alipay:
- `wx.request` ‚Ü?`my.request` (with `headers` instead of `header`, `res.status` instead of `res.statusCode`)
- `wx.navigateTo` ‚Ü?`my.navigateTo`
- `wx.redirectTo` ‚Ü?`my.redirectTo`
- `wx.navigateBack` ‚Ü?`my.navigateBack`
- `wx.showToast` ‚Ü?`my.showToast` (uses `content` instead of `title`)
- `wx.showModal` ‚Ü?`my.confirm`
- `wx.setStorageSync` ‚Ü?`my.setStorageSync`
- `wx.getStorageSync` ‚Ü?`my.getStorageSync`
- `wx.getUserProfile` ‚Ü?`my.getAuthCode`
- `a:if`/`a:else` instead of `wx:if`/`wx:else`
- `onTap` instead of `bindtap`
- `onChange` instead of `bindchange`

Pages to create (all follow WeChat pattern with `a:` prefix and `my.*` APIs):

1. `pages/index/` ‚Ä?P2 Home with city bar, hero CTA, policy scroll (use `onTap`, `a:if`, `my.navigateTo`, `my.request`)
2. `pages/city-picker/` ‚Ä?P3 City picker with city list, `my.navigateBack`
3. `pages/profile/` ‚Ä?P4 Multi-step form (3 steps), `my.request` for API
4. `pages/loading/` ‚Ä?P5 Loading spinner, `my.request` for generatePlan
5. `pages/preview/` ‚Ä?P6 Preview with blurred values
6. `pages/plan/` ‚Ä?P7 Plan detail with scheme comparison

Each page's WXML ‚Ü?AXML, WXSS ‚Ü?ACSS, JS API mapping as above.

- [ ] **Step 1: Copy WeChat page logic for all 6 pages**
- [ ] **Step 2: Replace all `wx.` with `my.` and `wx:` with `a:`**
- [ ] **Step 3: Verify all 7 pages exist with .axml, .acss, .js**

---

### Task 4: Final Verification

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

- [ ] **Step 1: Verify file structure**

```bash
Get-ChildItem -Recurse -LiteralPath "nsi-platform/frontend/alipay" -Include "*.js","*.axml","*.acss","*.json"
```

Expected: 7 page dirs √ó 3 files each + app.js, app.json, app.acss, utils/constants.js, services/api.js

- [ ] **Step 2: Verify navigation flow**

login ‚Ü?index ‚Ü?city-picker ‚Ü?profile ‚Ü?loading ‚Ü?preview ‚Ü?plan ‚Ä?all using `my.navigateTo`/`my.redirectTo`

- [ ] **Step 3: Verify API mapping**

All `wx.request` ‚Ü?`my.request`, `wx.showToast` ‚Ü?`my.showToast` (with `content` param), etc.
