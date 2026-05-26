const app = getApp();
const api = require('../../services/api');

Page({
  data: {
    category: 'general',
    content: '',
    contact: '',
    submitting: false,
  },
  onCategoryTap(e) {
    this.setData({ category: e.currentTarget.dataset.cat });
  },
  onContentInput(e) { this.setData({ content: e.detail.value }); },
  onContactInput(e) { this.setData({ contact: e.detail.value }); },
  onSubmit() {
    if (!this.data.content.trim()) {
      wx.showToast({ title: '请输入反馈内容', icon: 'none' });
      return;
    }
    this.setData({ submitting: true });
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    api.submitFeedback(userID, {
      category: this.data.category,
      content: this.data.content,
      contact: this.data.contact,
    }).then(() => {
      wx.showToast({ title: '感谢您的反馈！', icon: 'success' });
      setTimeout(() => wx.navigateBack(), 1500);
    }).catch(() => {
      this.setData({ submitting: false });
    });
  },
});
