const app = getApp();
const api = require('../../services/api');

const FONT_OPTIONS = [
  { value: 'small', label: '小' },
  { value: 'medium', label: '中' },
  { value: 'large', label: '大' },
];
const LANDING_OPTIONS = [
  { value: 'index', label: '首页' },
  { value: 'profile', label: '画像' },
  { value: 'sandbox', label: '沙盘' },
  { value: 'compliance', label: '合规' },
  { value: 'rights', label: '权益' },
];

Page({
  data: {
    fontOptions: FONT_OPTIONS,
    landingOptions: LANDING_OPTIONS,
    fontSize: 'medium',
    landingPage: 'index',
    landingIndex: 0,
    notifications: true,
    loaded: false,
  },
  onShow() {
    this.loadSettings();
  },
  getUserID() {
    return app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
  },
  loadSettings() {
    wx.showLoading({ title: '加载中', mask: true });
    api.getSettings(this.getUserID())
      .then((data) => {
        wx.hideLoading();
        if (!data) return;
        const landingValue = data.landing_page || 'index';
        const idx = Math.max(0, LANDING_OPTIONS.findIndex((o) => o.value === landingValue));
        this.setData({
          fontSize: data.font_size || 'medium',
          landingPage: landingValue,
          landingIndex: idx,
          notifications: data.notifications !== false,
          loaded: true,
        });
      })
      .catch(() => { wx.hideLoading(); });
  },
  onFontChange(e) {
    const value = e.detail.value;
    this.setData({ fontSize: value });
    this.save({ font_size: value });
  },
  onLandingChange(e) {
    const idx = Number(e.detail.value);
    const value = LANDING_OPTIONS[idx].value;
    this.setData({ landingIndex: idx, landingPage: value });
    this.save({ landing_page: value });
  },
  onNotifyChange(e) {
    const value = e.detail.value;
    this.setData({ notifications: value });
    this.save({ notifications: value });
  },
  save(patch) {
    api.saveSettings(this.getUserID(), patch)
      .then(() => { wx.showToast({ title: '已保存', icon: 'none' }); })
      .catch(() => {});
  },
  onLogout() {
    wx.showModal({
      title: '退出登录',
      content: '确定要退出当前账号吗？',
      confirmColor: '#EF4444',
      success: (res) => {
        if (!res.confirm) return;
        try { wx.removeStorageSync('token'); } catch (e) {}
        app.globalData.userInfo = null;
        wx.reLaunch({ url: '/pages/login/login' });
      },
    });
  },
});
