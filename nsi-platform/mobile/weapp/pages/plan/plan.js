const api = require('../../services/api')

Page({
  data: {
    cityName: '',
    cityCode: '',
    recommended: null,
    schemes: [],
    cashflow: [],
    actions: [],
    selectedIndex: 0,
    loading: true
  },
  onLoad() {
    const app = getApp()
    const city = app.globalData.currentCity
    if (city) {
      this.setData({ cityName: city.displayName, cityCode: city.code })
    }
    this.loadPlanDetail()
  },
  loadPlanDetail() {
    const app = getApp()
    const planId = app.globalData.planId
    if (planId) {
      api.getPlanDetail(planId).then((res) => {
        this.renderPlanData(res)
      }).catch(() => {
        this.renderFallbackData()
      })
    } else {
      this.renderFallbackData()
    }
  },
  renderPlanData(res) {
    const data = res.plan || res.data || res
    const schemes = (data.schemes || []).map((s, i) => ({
      name: s.name || '方案' + (i + 1),
      baseAmount: (s.base_amount || s.baseAmount || '0').toString(),
      monthlyPayment: (s.monthly_payment || s.monthlyPayment || '0').toString(),
      pension: (s.pension || '0').toString(),
      saving: (s.saving || s.saving_amount || '').toString(),
      badge: s.badge || (i === 0 ? '推荐' : ''),
      badgeColor: i === 0 ? '#059669' : '#6B7280',
      selected: i === 0
    }))
    const recommended = schemes.length > 0 ? {
      name: schemes[0].name,
      pension: schemes[0].pension
    } : null
    this.setData({
      schemes,
      recommended,
      cashflow: data.cashflow || data.cash_flow || this.generateMockCashflow(),
      actions: data.actions || data.suggestions || ['继续缴纳社保以累积更多年限', '关注当地社保政策变动', '定期更新个人信息'],
      loading: false
    })
  },
  renderFallbackData() {
    this.setData({
      recommended: { name: '智筹方案A', pension: '3,280' },
      schemes: [
        { name: '智筹方案A', baseAmount: '5,000', monthlyPayment: '1,850', pension: '3,280', saving: '6,000', badge: '推荐', badgeColor: '#059669', selected: true },
        { name: '标准方案B', baseAmount: '6,500', monthlyPayment: '2,405', pension: '3,850', saving: '3,200', badge: '', badgeColor: '#6B7280', selected: false },
        { name: '基础方案C', baseAmount: '3,500', monthlyPayment: '1,295', pension: '2,450', saving: '', badge: '经济', badgeColor: '#F59E0B', selected: false }
      ],
      cashflow: this.generateMockCashflow(),
      actions: ['继续缴纳社保以累积更多年限', '关注当地社保政策变动', '定期更新个人信息'],
      loading: false
    })
  },
  generateMockCashflow() {
    const rows = []
    for (let i = 1; i <= 15; i++) {
      const year = 2026 + i
      rows.push({
        year: year,
        monthlyPayment: Math.round(1850 - i * 30),
        yearlyPayment: Math.round((1850 - i * 30) * 12),
        cumulative: Math.round(i * 22000),
        pension: Math.round(3280 + i * 80)
      })
    }
    return rows
  },
  onSelectScheme(e) {
    const index = e.currentTarget.dataset.index
    const schemes = this.data.schemes.map((s, i) => ({
      ...s,
      selected: i === index
    }))
    this.setData({ schemes })
  },
  viewReport() {
    const app = getApp()
    const planId = app.globalData.planId
    if (planId) {
      api.getPlanReport(planId).then((res) => {
        if (res.report_url || res.url) {
          wx.showModal({
            title: '查看报告',
            content: '报告生成成功，请在浏览器中查看',
            confirmText: '知道了'
          })
        } else {
          wx.showToast({ title: '报告加载中，请稍后查看', icon: 'none' })
        }
      }).catch(() => {
        wx.showToast({ title: '报告生成失败，请重试', icon: 'none' })
      })
    } else {
      wx.showToast({ title: '暂无报告数据', icon: 'none' })
    }
  }
})
