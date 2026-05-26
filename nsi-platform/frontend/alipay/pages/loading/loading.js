const app = getApp();
const api = require('../../services/api');

Page({
  data: {
    progressText: '正在分析您的社保情况...',
  },
  onShow() {
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    const input = app.globalData.planInput;

    if (!input) {
      my.redirectTo({ url: '/pages/profile/profile' });
      return;
    }

    this.setData({ progressText: '正在计算最优方案...' });

    api.generatePlan(userID, input).then((result) => {
      my.setStorageSync({ key: 'planResult', data: result });
      setTimeout(() => {
        my.redirectTo({ url: '/pages/preview/preview' });
      }, 800);
    }).catch(() => {
      this.setData({ progressText: '生成失败，请重试' });
      setTimeout(() => {
        my.redirectTo({ url: '/pages/profile/profile' });
      }, 2000);
    });
  },
});
