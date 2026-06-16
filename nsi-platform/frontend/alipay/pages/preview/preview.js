const app = getApp();
const api = require('../../services/api');

Page({
  data: {
    blurredAmount: '约 ¥XX,XXX/年',
    blurredMonthly: '约 ¥XXX/月',
  },
  onUnlock() {
    const planResult = my.getStorageSync({ key: 'planResult' }).data;
    if (!planResult || !planResult.plan_id) {
      my.showToast({ content: '方案数据异常', type: 'none' });
      return;
    }
    const planId = planResult.plan_id;
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    my.showLoading({ content: '创建订单...' });
    api.createOrder(userID, planId).then((order) => {
      my.hideLoading();
      my.confirm({
        title: '确认支付',
        content: '支付 ¥19.90 解锁完整报告？',
        confirmButtonText: '确认支付',
        cancelButtonText: '取消',
        success: (res) => {
          if (res.confirm) {
            my.showLoading({ content: '支付中...' });
            api.payOrder(userID, order.order_id).then(() => {
              my.hideLoading();
              app.globalData.planResult = planResult;
              my.redirectTo({ url: '/pages/plan/plan?planId=' + planId });
            }).catch(() => { my.hideLoading(); });
          }
        },
      });
    }).catch(() => { my.hideLoading(); });
  },
});
