const { CITIES } = require('../../utils/constants')

Page({
  data: {
    selectedCode: '110000',
    selectedName: '北京',
    searchKey: '',
    cities: CITIES,
    filteredCities: CITIES,
    from: ''
  },
  onLoad(options) {
    const app = getApp()
    this.setData({
      from: options.from || '',
      selectedCode: app.globalData.currentCity ? app.globalData.currentCity.code : '110000'
    })
  },
  onSearch(e) {
    const key = e.detail.value.trim().toLowerCase()
    const filtered = this.data.cities.filter(
      c => c.name.includes(key) || c.displayName.includes(key) || c.code.includes(key)
    )
    this.setData({ searchKey: e.detail.value, filteredCities: filtered })
  },
  clearSearch() {
    this.setData({ searchKey: '', filteredCities: this.data.cities })
  },
  selectCity(e) {
    const { code, name } = e.currentTarget.dataset
    this.setData({ selectedCode: code, selectedName: name })
  },
  onConfirm() {
    const city = this.data.cities.find(c => c.code === this.data.selectedCode)
    if (!city) return
    const app = getApp()
    app.globalData.currentCity = city
    wx.showToast({ title: '已选择 ' + city.displayName, icon: 'success' })
    const pages = getCurrentPages()
    if (pages.length >= 2) {
      const prevPage = pages[pages.length - 2]
      if (prevPage.route === 'pages/login/login' || prevPage.route === 'pages/index/index') {
        wx.navigateBack()
        return
      }
    }
    wx.redirectTo({ url: '/pages/index/index' })
  }
})
