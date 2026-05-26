function login() {
  return new Promise((resolve, reject) => {
    wx.login({
      success(res) {
        if (res.code) {
          resolve(res.code)
        } else {
          wx.showToast({ title: '微信登录失败', icon: 'none' })
          reject(res)
        }
      },
      fail(err) {
        wx.showToast({ title: '登录异常，请稍后重试', icon: 'none' })
        reject(err)
      }
    })
  })
}

function getUserProfile() {
  return new Promise((resolve, reject) => {
    wx.getUserProfile({
      desc: '用于完善社保个人信息',
      success(res) {
        resolve(res.userInfo)
      },
      fail(err) {
        if (err.errMsg.indexOf('auth deny') !== -1 || err.errMsg.indexOf('fail') !== -1) {
          reject(new Error('用户拒绝授权'))
        } else {
          reject(err)
        }
      }
    })
  })
}

function checkLocationAuth() {
  return new Promise((resolve) => {
    wx.getSetting({
      success(res) {
        if (res.authSetting['scope.userLocation']) {
          resolve(true)
        } else {
          resolve(false)
        }
      },
      fail() {
        resolve(false)
      }
    })
  })
}

function requestLocationAuth() {
  return new Promise((resolve, reject) => {
    wx.authorize({
      scope: 'scope.userLocation',
      success() {
        resolve(true)
      },
      fail() {
        wx.showModal({
          title: '需要位置权限',
          content: '我们需要您的地理位置信息来推荐城市社保政策',
          showCancel: true,
          confirmText: '去设置',
          success(res) {
            if (res.confirm) {
              wx.openSetting({
                success(settingRes) {
                  if (settingRes.authSetting['scope.userLocation']) {
                    resolve(true)
                  } else {
                    resolve(false)
                  }
                },
                fail() {
                  resolve(false)
                }
              })
            } else {
              resolve(false)
            }
          }
        })
      }
    })
  })
}

module.exports = {
  login,
  getUserProfile,
  checkLocationAuth,
  requestLocationAuth
}
