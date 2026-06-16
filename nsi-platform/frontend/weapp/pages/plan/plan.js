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
      wx.showToast({ title: '参数错误', icon: 'none' });
      return;
    }
    api.checkUnlock(app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default', planId).then((res) => {
      if (!res.unlocked) {
        wx.showModal({ title: '提示', content: '请先解锁完整报告', showCancel: false });
      }
    }).catch(() => {});
    const cached = wx.getStorageSync('planResult');
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
    wx.navigateTo({ url: '/pages/compliance/compliance' });
  },
  onSavePDF() {
    wx.showToast({ title: 'PDF 报告导出（Phase 2）', icon: 'none' });
  },
  onShare() {
    wx.showShareMenu({ withShareTicket: true });
  },
});
