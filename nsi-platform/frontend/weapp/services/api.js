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
        'Authorization': 'Bearer ' + (wx.getStorageSync('token') || ''),
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
function createOrder(userID, planId) { return request('POST', '/v1/orders', { plan_id: planId }, userID); }
function payOrder(userID, orderId)   { return request('POST', '/v1/orders/' + orderId + '/pay', { payment_method: 'wechat' }, userID); }
function checkUnlock(userID, planId) { return request('GET', '/v1/orders/check-unlock?plan_id=' + planId, null, userID); }
function getSettings(userID)              { return request('GET', '/v1/settings', null, userID); }
function saveSettings(userID, data)       { return request('POST', '/v1/settings', data, userID); }
function calculateSimulator(userID, data) { return request('POST', '/v1/simulator/calculate', data, userID); }
function askAdvisor(userID, data)         { return request('POST', '/v1/advisor/ask', data, userID); }

function getToken(userID) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: API_BASE_URL + '/v1/auth/token',
      method: 'POST',
      header: { 'Content-Type': 'application/json' },
      data: { user_id: userID },
      success: res => {
        if (res.data && res.data.data && res.data.data.token) {
          wx.setStorageSync('token', res.data.data.token);
          resolve(res.data.data.token);
        } else { reject(new Error('No token')); }
      },
      fail: reject
    });
  });
}

function sendSMS(phone) {
  return request('POST', '/v1/auth/sms/send', { phone: phone }, null);
}

function verifySMS(phone, code) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: API_BASE_URL + '/v1/auth/sms/verify',
      method: 'POST',
      header: { 'Content-Type': 'application/json' },
      data: { phone: phone, code: code },
      success: res => {
        if (res.data && res.data.data && res.data.data.token) {
          wx.setStorageSync('token', res.data.data.token);
          resolve(res.data.data);
        } else { reject(res.data); }
      },
      fail: reject
    });
  });
}

function deleteAccount() {
  return request('POST', '/v1/auth/delete-account-v2', { confirm: 'DELETE' }, null);
}

function submitPaymentRecord(data) {
  return request('POST', '/v1/rights/payment-records', data, null);
}

function saveScenario(name, params) {
  return request('POST', '/v1/simulator/scenarios', { name: name, params: params }, null);
}

function listScenarios() {
  return request('GET', '/v1/simulator/scenarios', null, null);
}

module.exports = {
  API_BASE_URL,
  getProfile, updateProfile, queryPolicies, generatePlan, getPlanDetail, getPlanReport,
  getCompliance, getGuide, getPaymentStatus, getAlerts, markAlertRead, submitFeedback,
  createOrder, payOrder, checkUnlock,
  getSettings, saveSettings, calculateSimulator, askAdvisor,
  getToken, sendSMS, verifySMS, deleteAccount, submitPaymentRecord, saveScenario, listScenarios,
};
