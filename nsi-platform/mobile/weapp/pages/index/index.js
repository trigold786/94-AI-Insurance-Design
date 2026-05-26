const api = require('../../services/api')
const { POLICY_TYPE_MAP } = require('../../utils/constants')

Page({
  data: {
    cityName: '北京市',
    cityCode: '110000',
    userName: '',
    greeting: '你好',
    policies: [],
    loading: false
  },
  onLoad() {
    this.loadCityInfo()
    this.updateGreeting()
  },
  onShow() {
    const app = getApp()
    if (app.globalData.currentCity) {
      this.setData({
        cityName: app.globalData.currentCity.displayName,
        cityCode: app.globalData.currentCity.code
      }, () => {
        this.loadPolicies()
      })
    } else if (this.data.cityCode) {
      this.loadPolicies()
    }
  },
  loadCityInfo() {
    const app = getApp()
    if (app.globalData.currentCity) {
      this.setData({
        cityName: app.globalData.currentCity.displayName,
        cityCode: app.globalData.currentCity.code
      })
    }
    if (app.globalData.userInfo) {
      this.setData({ userName: app.globalData.userInfo.nickName || '' })
    }
  },
  updateGreeting() {
    const hour = new Date().getHours()
    let greeting = '你好'
    if (hour < 6) greeting = '夜深了'
    else if (hour < 9) greeting = '早上好'
    else if (hour < 12) greeting = '上午好'
    else if (hour < 14) greeting = '中午好'
    else if (hour < 18) greeting = '下午好'
    else greeting = '晚上好'
    this.setData({ greeting })
  },
  loadPolicies() {
    this.setData({ loading: true })
    api.getPolicies(this.data.cityCode).then((res) => {
      const policies = (res.policies || res.data || []).map((p, i) => {
        const typeInfo = POLICY_TYPE_MAP[p.type] || { label: '社保政策', icon: '📋', color: '#6B7280' }
        return {
          title: p.title || p.name || typeInfo.label,
          description: p.description || p.summary || '',
          icon: typeInfo.icon || '📋',
          iconBg: (typeInfo.color || '#6B7280') + '20',
          tags: p.tags || [typeInfo.label],
          raw: p
        }
      })
      this.setData({ policies, loading: false })
    }).catch(() => {
      this.setData({
        loading: false,
        policies: [
          { title: '养老保险', description: '为退休后的生活提供基本收入保障', icon: '🔵', iconBg: '#1A56DB20', tags: ['养老保障'] },
          { title: '医疗保险', description: '报销医疗费用，减轻看病负担', icon: '🟢', iconBg: '#05966920', tags: ['医疗保障'] },
          { title: '失业保险', description: '失业期间提供基本生活保障', icon: '🟠', iconBg: '#F59E0B20', tags: ['失业保障'] }
        ]
      })
    })
  },
  goCityPicker() {
    wx.navigateTo({ url: '/pages/city-picker/city-picker?from=index' })
  },
  goProfile() {
    const app = getApp()
    if (!app.globalData.isLoggedIn) {
      wx.showModal({
        title: '提示',
        content: '请先授权登录以获取个性化方案',
        confirmText: '去登录',
        success(res) {
          if (res.confirm) {
            wx.redirectTo({ url: '/pages/login/login' })
          }
        }
      })
      return
    }
    wx.navigateTo({ url: '/pages/profile/profile' })
  },
  onPolicyTap(e) {
    const policy = e.detail.title
    wx.showToast({ title: '政策详情：' + policy, icon: 'none' })
  }
})
