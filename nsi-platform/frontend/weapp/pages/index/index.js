const app = getApp();
const api = require('../../services/api');

Page({
  data: {
    currentCity: '加载中...',
    currentCityCode: '',
    policies: [],
  },
  onShow() {
    const city = app.globalData.currentCity;
    const code = app.globalData.currentCityCode;
    if (city) {
      this.setData({ currentCity: city, currentCityCode: code });
      this.loadPolicies(code);
    } else {
      this.detectLocation();
    }
  },
  detectLocation() {
    wx.getLocation({
      type: 'wgs84',
      success: (res) => {
        const c = this.mapLocationToCity(res.latitude, res.longitude);
        this.applyCity(c.name, c.code);
      },
      fail: () => { this.applyCity('上海', '310000'); },
    });
  },
  applyCity(name, code) {
    app.globalData.currentCity = name;
    app.globalData.currentCityCode = code;
    this.setData({ currentCity: name, currentCityCode: code });
    this.loadPolicies(code);
  },
  mapLocationToCity(lat, lng) {
    const cities = [
      { lat: 31.23, lng: 121.47, code: '310000', name: '上海' },
      { lat: 39.90, lng: 116.40, code: '110000', name: '北京' },
      { lat: 22.54, lng: 114.06, code: '440300', name: '深圳' },
      { lat: 23.13, lng: 113.27, code: '440100', name: '广州' },
      { lat: 30.27, lng: 120.15, code: '330100', name: '杭州' },
    ];
    let best = cities[0], bestDist = Infinity;
    for (const c of cities) {
      const d = (c.lat - lat) ** 2 + (c.lng - lng) ** 2;
      if (d < bestDist) { bestDist = d; best = c; }
    }
    return best;
  },
  loadPolicies(code) {
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    api.queryPolicies(userID, { regionCode: code })
      .then((claims) => {
        this.setData({ policies: claims || [] });
      })
      .catch(() => {});
  },
  onComplianceTap() {
    wx.navigateTo({ url: '/pages/compliance/compliance' });
  },
  onRightsTap() {
    wx.navigateTo({ url: '/pages/rights/rights' });
  },
  onCityTap() {
    wx.navigateTo({ url: '/pages/city-picker/city-picker' });
  },
  onStartPlan() {
    wx.navigateTo({ url: '/pages/profile/profile' });
  },
  onPolicyTap(e) {
    const item = e.currentTarget.dataset.item;
    wx.showModal({ title: item.policy_id, content: item.subsidy_calc_method });
  },
});
