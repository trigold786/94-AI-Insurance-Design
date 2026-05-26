const app = getApp();
const api = require('../../services/api');

Page({
  data: {
    paymentSummary: {},
    alerts: [],
    loading: true,
    error: '',
  },
  onShow() { this.loadData(); },
  loadData() {
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    this.setData({ loading: true, error: '' });

    Promise.all([
      api.getPaymentStatus(userID).catch(() => ({})),
      api.getAlerts(userID).catch(() => []),
    ]).then(([summary, alerts]) => {
      this.setData({ paymentSummary: summary || {}, alerts: alerts || [], loading: false });
    });
  },
  onMarkRead(e) {
    const id = e.currentTarget.dataset.id;
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    api.markAlertRead(userID, id).then(() => { this.loadData(); });
  },
});
