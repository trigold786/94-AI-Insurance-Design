const app = getApp();

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
    my.request({
      url: 'http://127.0.0.1:39401/v1/compliance/checklist?city_code=' + cityCode,
      method: 'GET',
      headers: { 'x-user-id': userID },
      success(res) {
        if (res.status === 200 && res.data.code === 0) {
          that.setData({
            matchedPolicies: res.data.data.matched_policies || [],
            requiredDocs: res.data.data.required_docs || [],
            eligibleTags: res.data.data.eligible_tags || [],
            loading: false,
          });
        }
      },
      fail() {
        my.showToast({ content: '加载失败', type: 'none' });
        that.setData({ loading: false });
      },
    });
  },
});
