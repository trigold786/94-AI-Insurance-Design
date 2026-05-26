const app = getApp();

Page({
  data: {
    blurredAmount: '约 ¥XX,XXX/年',
    blurredMonthly: '约 ¥XXX/月',
  },
  onUnlock() {
    const planResult = my.getStorageSync({ key: 'planResult' }).data;
    if (planResult && planResult.plan_id) {
      app.globalData.planResult = planResult;
      my.navigateTo({ url: '/pages/plan/plan?planId=' + planResult.plan_id });
    } else {
      my.showToast({ content: '方案数据异常', type: 'none' });
    }
  },
});
