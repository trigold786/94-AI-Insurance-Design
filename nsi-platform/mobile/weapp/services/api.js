const BASE_URL = 'https://api.nsinsurance.cn'
const DEV_BASE_URL = 'http://127.0.0.1:39401'
const ENV = 'prod'
const API_BASE_URL = ENV === 'dev' ? DEV_BASE_URL : BASE_URL

const request = (method, path, data = {}, extraHeader = {}) => {
  const app = getApp()
  const userId = app.globalData.userInfo ? 'wx_' + app.globalData.userInfo.nickName : 'mock_user_001'
  return new Promise((resolve, reject) => {
    wx.request({
      url: API_BASE_URL + path,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        'x-user-id': userId,
        ...extraHeader
      },
      success(res) {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data)
        } else if (res.statusCode === 401) {
          wx.showToast({ title: '登录已过期，请重新授权', icon: 'none' })
          reject(res.data)
        } else {
          wx.showToast({
            title: res.data && res.data.message ? res.data.message : '请求失败',
            icon: 'none'
          })
          reject(res.data)
        }
      },
      fail(err) {
        wx.showToast({ title: '网络异常，请检查网络连接', icon: 'none' })
        reject(err)
      }
    })
  })
}

module.exports = {
  healthz() {
    return request('GET', '/healthz')
  },
  saveProfile(profile) {
    return request('PUT', '/v1/profile', profile)
  },
  getPolicies(region) {
    return request('GET', '/v1/policies', { region })
  },
  generatePlan(params) {
    return request('POST', '/v1/plans/generate', params)
  },
  getPlanDetail(id) {
    return request('GET', '/v1/plans/' + id)
  },
  getPlanReport(planId) {
    return request('GET', '/v1/plans/report', { plan_id: planId })
  },
  getComplianceChecklist(cityCode) {
    return request('GET', '/v1/compliance/checklist', { city_code: cityCode })
  },
  getGuide(cityCode) {
    return request('GET', '/v1/guide', { city_code: cityCode })
  },
  submitFeedback(feedback) {
    return request('POST', '/v1/feedback', feedback)
  },
  getPaymentStatus() {
    return request('GET', '/v1/rights/payment-status')
  },
  getAlerts() {
    return request('GET', '/v1/rights/alerts')
  }
}
