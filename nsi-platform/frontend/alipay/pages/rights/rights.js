const app = getApp();

Page({
  data: {
    paymentSummary: {},
    alerts: [],
    loading: true,
  },
  onShow() {
    this.loadData();
  },
  loadData() {
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    this.setData({ loading: true });

    const that = this;
    const baseUrl = 'http://127.0.0.1:39401';

    my.request({
      url: baseUrl + '/v1/rights/payment-status',
      method: 'GET',
      headers: { 'x-user-id': userID },
      success(res) {
        if (res.data && res.data.code === 0) {
          that.setData({ paymentSummary: res.data.data || {} });
        }
      },
      complete() {
        my.request({
          url: baseUrl + '/v1/rights/alerts?unread_only=false',
          method: 'GET',
          headers: { 'x-user-id': userID },
          success(res2) {
            if (res2.data && res2.data.code === 0) {
              that.setData({ alerts: res2.data.data || [], loading: false });
            }
          },
          fail() { that.setData({ loading: false }); },
        });
      },
    });
  },
});
