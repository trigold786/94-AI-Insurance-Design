const BASE_URL = 'https://api.nsinsurance.cn';
const DEV_BASE_URL = 'http://127.0.0.1:39401';
const ENV = 'prod';
const API_BASE_URL = ENV === 'dev' ? DEV_BASE_URL : BASE_URL;

function request(method, path, data, userID) {
  return new Promise((resolve, reject) => {
    wx.showNavigationBarLoading();
    wx.request({
      url: API_BASE_URL + path,
      method,
      data,
      header: {
        'x-user-id': userID || 'default',
        'Content-Type': 'application/json',
      },
      success: (res) => {
        if (res.statusCode === 401) {
          wx.showToast({ title: '请先登录', icon: 'none' });
          reject(new Error('unauthorized'));
          return;
        }
        if (res.statusCode === 404) {
          reject(new Error('not_found'));
          return;
        }
        if (res.statusCode >= 400) {
          const msg = (res.data && (res.data.msg || res.data.message)) || '请求失败';
          wx.showToast({ title: msg, icon: 'none' });
          reject(new Error(msg));
          return;
        }
        const body = res.data;
        if (body.code !== 0) {
          const msg = body.msg || body.message || '服务异常';
          wx.showToast({ title: msg, icon: 'none' });
          reject(new Error(msg));
          return;
        }
        resolve(body.data !== undefined ? body.data : body);
      },
      fail: () => {
        wx.showToast({ title: '网络异常，请检查连接', icon: 'none' });
        reject(new Error('network_error'));
      },
      complete: () => {
        wx.hideNavigationBarLoading();
      },
    });
  });
}

function getProfile(userID)           { return request('GET', '/v1/profile', null, userID); }
function updateProfile(userID, data)  { return request('PUT', '/v1/profile', data, userID); }
function queryPolicies(userID, opts) {
  let path = '/v1/policies';
  const params = [];
  if (opts) {
    if (opts.regionCode) params.push('region_code=' + opts.regionCode);
    if (opts.policyType) params.push('policy_type=' + opts.policyType);
    if (opts.status) params.push('status=' + opts.status);
  }
  if (params.length) path += '?' + params.join('&');
  return request('GET', path, null, userID);
}
function generatePlan(userID, data)   { return request('POST', '/v1/plans/generate', data, userID); }
function getPlanDetail(userID, pid)  { return request('GET', '/v1/plans/' + pid, null, userID); }
function getPlanReport(userID, pid)  { return request('GET', '/v1/plans/report?plan_id=' + pid, null, userID); }
function getCompliance(userID, city) { return request('GET', '/v1/compliance/checklist?city_code=' + city, null, userID); }
function getGuide(userID, city)      { return request('GET', '/v1/guide?city_code=' + city, null, userID); }
function getPaymentStatus(userID)    { return request('GET', '/v1/rights/payment-status', null, userID); }
function getAlerts(userID)           { return request('GET', '/v1/rights/alerts', null, userID); }
function markAlertRead(userID, id)   { return request('POST', '/v1/rights/alerts/read', { alert_id: id }, userID); }
function submitFeedback(userID, data){ return request('POST', '/v1/feedback', data, userID); }

module.exports = {
  getProfile, updateProfile, queryPolicies, generatePlan, getPlanDetail, getPlanReport,
  getCompliance, getGuide, getPaymentStatus, getAlerts, markAlertRead, submitFeedback,
};
