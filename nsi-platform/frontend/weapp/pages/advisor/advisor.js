const app = getApp();
const api = require('../../services/api');

const WELCOME = '您好，我是 AI 社保顾问，可以为您解答养老、医疗、补贴等社保相关问题，请问有什么可以帮您？';

Page({
  data: {
    messages: [{ role: 'assistant', content: WELCOME }],
    input: '',
    loading: false,
    scrollIntoView: '',
  },
  onLoad() {
    const self = this;
    setTimeout(() => { self.scrollToEnd(); }, 100);
  },
  getUserID() {
    return app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
  },
  onInput(e) { this.setData({ input: e.detail.value }); },
  onSend() {
    const question = (this.data.input || '').trim();
    if (!question || this.data.loading) return;
    const messages = this.data.messages.concat({ role: 'user', content: question });
    this.setData({ messages, input: '', loading: true });
    this.scrollToEnd();
    api.askAdvisor(this.getUserID(), { question })
      .then((data) => {
        const answer = (data && (data.answer || data.content)) || '抱歉，我暂时无法回答这个问题。';
        const newMsgs = this.data.messages.concat({ role: 'assistant', content: answer });
        this.setData({ messages: newMsgs, loading: false });
        this.scrollToEnd();
      })
      .catch(() => {
        this.setData({ loading: false });
        wx.showToast({ title: '回复失败，请重试', icon: 'none' });
      });
  },
  scrollToEnd() {
    const idx = this.data.messages.length - 1;
    this.setData({ scrollIntoView: 'msg-' + idx });
  },
});
