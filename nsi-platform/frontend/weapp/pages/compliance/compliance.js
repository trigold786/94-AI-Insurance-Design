const app = getApp();
const api = require('../../services/api');

Page({
  data: {
    matchedPolicies: [],
    requiredDocs: [],
    eligibleTags: [],
    loading: true,
  },
  onShow() {
    this.loadCompliance();
  },
  loadCompliance() {
    const cityCode = app.globalData.currentCityCode || '310000';
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    this.setData({ loading: true });

    const that = this;
    wx.request({
      url: 'http://127.0.0.1:39401/v1/compliance/checklist?city_code=' + cityCode,
      method: 'GET',
      header: { 'x-user-id': userID },
      success(res) {
        if (res.statusCode === 200 && res.data.code === 0) {
          that.setData({
            matchedPolicies: res.data.data.matched_policies || [],
            requiredDocs: res.data.data.required_docs || [],
            eligibleTags: res.data.data.eligible_tags || [],
            loading: false,
          });
        }
      },
      fail() {
        wx.showToast({ title: '加载失败', icon: 'none' });
        that.setData({ loading: false });
      },
    });
  },
  onGuideTap() {
    const cityCode = app.globalData.currentCityCode || '310000';
    const url = api.API_BASE_URL + '/v1/guide?city_code=' + cityCode;
    wx.navigateTo({ url: '/pages/webview/webview?url=' + encodeURIComponent(url) });
  },
});
