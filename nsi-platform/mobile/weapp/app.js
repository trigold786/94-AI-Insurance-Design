App({
  globalData: {
    userInfo: null,
    isLoggedIn: false,
    currentCity: null,
    currentCityCode: '',
    policies: [],
    planId: '',
    profileData: {},
    planResult: null,
    stepData: {}
  },
  onLaunch() {
    const that = this
    wx.getSetting({
      success(res) {
        if (res.authSetting['scope.userInfo']) {
          wx.getUserInfo({
            success(infoRes) {
              that.globalData.userInfo = infoRes.userInfo
              that.globalData.isLoggedIn = true
            }
          })
        }
      }
    })
    wx.getLocation({
      type: 'wgs84',
      success(locRes) {
        that.globalData.location = {
          latitude: locRes.latitude,
          longitude: locRes.longitude
        }
      },
      fail() {}
    })
  },
  checkLogin() {
    return new Promise((resolve) => {
      if (this.globalData.isLoggedIn) {
        resolve(true)
      } else {
        wx.getSetting({
          success(res) {
            if (res.authSetting['scope.userInfo']) {
              resolve(true)
            } else {
              resolve(false)
            }
          },
          fail() {
            resolve(false)
          }
        })
      }
    })
  }
})
