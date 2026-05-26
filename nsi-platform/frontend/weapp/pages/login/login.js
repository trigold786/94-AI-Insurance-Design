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
    wx.showModal({
      title: '隐私政策',
      content: '我们收集您的信息用于社保规划服务...',
      showCancel: true,
    });
  },
  onViewTerms() {
    wx.showModal({
      title: '用户协议',
      content: '欢迎使用AI社保智筹...',
      showCancel: true,
    });
  },
  onLogin() {
    if (!this.data.agreed) return;
    this.setData({ loading: true });
    wx.getUserProfile({
      desc: '用于完善用户资料',
      success: (res) => {
        app.globalData.userInfo = res.userInfo;
        this.getLocationAndNavigate();
      },
      fail: () => {
        this.getLocationAndNavigate();
      },
    });
  },
  getLocationAndNavigate() {
    wx.getLocation({
      type: 'wgs84',
      success: (res) => {
        app.globalData.currentCity = this.getCityName(res.latitude, res.longitude);
        wx.redirectTo({ url: '/pages/index/index' });
      },
      fail: () => {
        wx.redirectTo({ url: '/pages/index/index' });
      },
    });
  },
  getCityName(lat, lng) {
    const { CITIES } = require('../../utils/constants');
    app.globalData.currentCity = CITIES[0].name;
    app.globalData.currentCityCode = CITIES[0].code;
    return CITIES[0].name;
  },
});
