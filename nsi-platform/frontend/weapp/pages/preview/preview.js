const app = getApp();

Page({
  data: {
    blurredAmount: '约 ¥XX,XXX/年',
    blurredMonthly: '约 ¥XXX/月',
  },
  onUnlock() {
    const planResult = wx.getStorageSync('planResult');
    if (planResult && planResult.plan_id) {
      app.globalData.planResult = planResult;
      wx.navigateTo({ url: '/pages/plan/plan?planId=' + planResult.plan_id });
    } else {
      wx.showToast({ title: '方案数据异常', icon: 'none' });
    }
  },
});
