const { PLAN_FEATURES } = require('../../utils/constants')

Page({
  data: {
    features: PLAN_FEATURES
  },
  goToPlan() {
    wx.showModal({
      title: '解锁完整方案',
      content: '您将查看完整的社保规划方案详情，包含方案对比、现金流分析等全部内容',
      confirmText: '查看方案',
      success(res) {
        if (res.confirm) {
          wx.redirectTo({ url: '/pages/plan/plan' })
        }
      }
    })
  }
})
