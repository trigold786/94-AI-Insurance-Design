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
      app.globalData.currentCity = '上海';
      app.globalData.currentCityCode = '310000';
      this.setData({ currentCity: '上海', currentCityCode: '310000' });
      this.loadPolicies('310000');
    }
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
