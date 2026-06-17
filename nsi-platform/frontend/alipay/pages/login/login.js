const app = getApp();

Page({
  data: {
    agreed: false,
    loading: false,
  },
  onAgreeChange(e) {
    this.setData({ agreed: e.detail.value.length > 0 });
  },
  onViewPrivacy() {
    my.alert({
      title: '隐私政策',
      content: '我们收集您的信息用于社保规划服务...',
    });
  },
  onViewTerms() {
    my.alert({
      title: '用户协议',
      content: '欢迎使用AI社保智筹...',
    });
  },
  onLogin() {
    if (!this.data.agreed) return;
    this.setData({ loading: true });
    const api = require('../../services/api');
    my.getAuthCode({
      scopes: 'auth_user',
      success: (res) => {
        const userID = res.authCode || 'default';
        api.getToken(userID).then(() => {
          app.globalData.userID = userID;
          app.globalData.userInfo = { nickName: userID };
          this.getLocationAndNavigate();
        }).catch(() => {
          this.setData({ loading: false });
          my.showToast({ content: '登录失败，请重试', type: 'none' });
        });
      },
      fail: () => {
        this.setData({ loading: false });
        my.showToast({ content: '登录失败，请重试', type: 'none' });
      },
    });
  },
  getLocationAndNavigate() {
    my.getLocation({
      type: 1,
      success: () => {
        const { CITIES } = require('../../utils/constants');
        app.globalData.currentCity = CITIES[0].name;
        app.globalData.currentCityCode = CITIES[0].code;
        my.redirectTo({ url: '/pages/index/index' });
      },
      fail: () => {
        my.redirectTo({ url: '/pages/index/index' });
      },
    });
  },
});
