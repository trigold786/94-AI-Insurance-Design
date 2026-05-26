const api = require('../../services/api')
const { LOADING_MESSAGES } = require('../../utils/constants')

Page({
  data: {
    progress: 0,
    currentMessage: LOADING_MESSAGES[0],
    messageIndex: 0,
    tips: [
      { text: '读取个人信息档案', done: false },
      { text: '分析城市社保政策', done: false },
      { text: '智能匹配缴费方案', done: false },
      { text: '计算未来养老收益', done: false },
      { text: '生成个性规划报告', done: false }
    ]
  },
  onLoad() {
    this.startProgress()
    this.startMessageCycle()
    this.generatePlan()
  },
  onUnload() {
    this.clearTimers()
  },
  clearTimers() {
    if (this._progressTimer) clearInterval(this._progressTimer)
    if (this._messageTimer) clearInterval(this._messageTimer)
  },
  startProgress() {
    this._progressTimer = setInterval(() => {
      let p = this.data.progress
      if (p < 90) {
        p += Math.random() * 8 + 2
        if (p > 90) p = 90
        this.setData({ progress: Math.round(p) })
      }
    }, 800)
  },
  startMessageCycle() {
    this._messageTimer = setInterval(() => {
      let idx = this.data.messageIndex
      idx = (idx + 1) % LOADING_MESSAGES.length
      this.setData({
        currentMessage: LOADING_MESSAGES[idx],
        messageIndex: idx
      })
      const doneCount = Math.min(idx + 1, this.data.tips.length)
      const tips = this.data.tips.map((t, i) => ({ ...t, done: i < doneCount }))
      this.setData({ tips })
    }, 2000)
  },
  generatePlan() {
    const app = getApp()
    const profile = app.globalData.stepData || {}
    const cityCode = app.globalData.currentCity ? app.globalData.currentCity.code : '110000'
    const params = {
      city_code: cityCode,
      age: profile.age || 30,
      gender: profile.gender || 'male',
      employment: profile.employment || 'employed',
      social_years: profile.social_years || 5
    }
    api.generatePlan(params).then((res) => {
      const planId = res.plan_id || res.id || ''
      app.globalData.planId = planId
      app.globalData.planResult = res
      this.setData({ progress: 100 })
      this.clearTimers()
      const tips = this.data.tips.map(t => ({ ...t, done: true }))
      this.setData({ tips, currentMessage: '方案生成完成！' })
      setTimeout(() => {
        wx.redirectTo({ url: '/pages/preview/preview' })
      }, 800)
    }).catch(() => {
      this.clearTimers()
      this.setData({ progress: 100, currentMessage: '生成完成，加载方案中...' })
      setTimeout(() => {
        wx.redirectTo({ url: '/pages/preview/preview' })
      }, 500)
    })
  }
})
