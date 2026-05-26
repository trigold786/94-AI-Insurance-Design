const authUtil = require('../../utils/auth')
const { CITIES } = require('../../utils/constants')

Page({
  data: {
    agreed: false,
    logging: false,
    currentCity: '北京市',
    currentCityCode: '110000',
    features: [
      { icon: '🛡️', text: 'AI 智能分析社保政策，为您定制最优方案' },
      { icon: '📊', text: '多方案对比，直观展示缴费与收益差异' },
      { icon: '💰', text: '合规节费，合理节省社保缴费成本' }
    ]
  },
  onLoad() {
    const app = getApp()
    if (app.globalData.currentCity) {
      this.setData({
        currentCity: app.globalData.currentCity.displayName,
        currentCityCode: app.globalData.currentCity.code
      })
    }
  },
  onShow() {
    const app = getApp()
    if (app.globalData.currentCity) {
      this.setData({
        currentCity: app.globalData.currentCity.displayName,
        currentCityCode: app.globalData.currentCity.code
      })
    }
  },
  toggleAgreement() {
    this.setData({ agreed: !this.data.agreed })
  },
  onStart() {
    if (!this.data.agreed) return
    this.setData({ logging: true })
    authUtil.login().then(() => {
      return authUtil.getUserProfile()
    }).then((userInfo) => {
      const app = getApp()
      app.globalData.userInfo = userInfo
      app.globalData.isLoggedIn = true
      this.setData({ logging: false })
      wx.redirectTo({ url: '/pages/index/index' })
    }).catch((err) => {
      this.setData({ logging: false })
      if (err.message === '用户拒绝授权') {
        wx.showModal({
          title: '提示',
          content: '授权后可获得更精准的社保规划服务，您也可以先浏览',
          confirmText: '继续浏览',
          success() {
            wx.redirectTo({ url: '/pages/index/index' })
          }
        })
      } else {
        wx.showToast({ title: '登录失败，请重试', icon: 'none' })
      }
    })
  },
  goCityPicker() {
    wx.navigateTo({ url: '/pages/city-picker/city-picker?from=login' })
  },
  onPrivacy() {
    wx.showModal({
      title: '隐私政策',
      content: '我们重视您的隐私安全。收集的信息仅用于社保规划服务，严格保密，不会向第三方共享您的个人信息。',
      showCancel: false
    })
  },
  onTerms() {
    wx.showModal({
      title: '用户协议',
      content: '使用AI社保智筹即表示您同意本服务条款。我们提供的社保规划建议仅供参考，不构成正式法律意见。',
      showCancel: false
    })
  }
})
