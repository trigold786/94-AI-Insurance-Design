const app = getApp();
const api = require('../../services/api');

Page({
  data: {
    schemes: [],
    currentScheme: 0,
    currentSchemeData: null,
  },
  onLoad(query) {
    const planId = query.planId;
    if (!planId) {
      my.showToast({ content: '参数错误', type: 'none' });
      return;
    }
    api.checkUnlock(app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default', planId).then((res) => {
      if (!res.unlocked) {
        my.alert({ title: '提示', content: '请先解锁完整报告' });
      }
    }).catch(() => {});
    const cached = my.getStorageSync({ key: 'planResult' }).data;
    if (cached && cached.recommended_schemes) {
      this.setData({ schemes: cached.recommended_schemes });
      this.selectScheme(0);
    } else {
      const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
      api.getPlanDetail(userID, planId).then((plan) => {
        if (plan && plan.recommended_schemes) {
          this.setData({ schemes: plan.recommended_schemes });
          this.selectScheme(0);
        }
      }).catch(() => {});
    }
  },
  selectScheme(index) {
    const schemes = this.data.schemes;
    if (schemes && schemes[index]) {
      this.setData({
        currentScheme: index,
        currentSchemeData: schemes[index],
      });
    }
  },
  onSchemeTap(e) {
    this.selectScheme(e.currentTarget.dataset.index);
  },
  onApplySubsidy() {
    my.navigateTo({ url: '/pages/compliance/compliance' });
  },
  onSavePDF() {
    my.showToast({ content: 'PDF 报告导出（Phase 2）', type: 'none' });
  },
  onShare() {
    my.showShareMenu();
  },
});
