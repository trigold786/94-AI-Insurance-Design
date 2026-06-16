const app = getApp();
const api = require('../../services/api');

Page({
  data: {
    blurredAmount: '约 ¥XX,XXX/年',
    blurredMonthly: '约 ¥XXX/月',
  },
  onUnlock() {
    const planResult = wx.getStorageSync('planResult');
    if (!planResult || !planResult.plan_id) {
      wx.showToast({ title: '方案数据异常', icon: 'none' });
      return;
    }
    const planId = planResult.plan_id;
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    wx.showLoading({ title: '创建订单...' });
    api.createOrder(userID, planId).then((order) => {
      wx.hideLoading();
      wx.showModal({
        title: '确认支付',
        content: '支付 ¥19.90 解锁完整报告？',
        success: (res) => {
          if (res.confirm) {
            wx.showLoading({ title: '支付中...' });
            api.payOrder(userID, order.order_id).then(() => {
              wx.hideLoading();
              app.globalData.planResult = planResult;
              wx.redirectTo({ url: '/pages/plan/plan?planId=' + planId });
            }).catch(() => { wx.hideLoading(); });
          }
        },
      });
    }).catch(() => { wx.hideLoading(); });
  },
});
