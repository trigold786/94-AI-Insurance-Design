const BASE_URL = 'https://api.nsinsurance.cn';
const DEV_BASE_URL = 'http://127.0.0.1:39401';
const ENV = 'prod';
const API_BASE_URL = ENV === 'dev' ? DEV_BASE_URL : BASE_URL;

function request(method, path, data, userID) {
  return new Promise((resolve, reject) => {
    my.showLoading({ content: '加载中...' });
    my.request({
      url: API_BASE_URL + path,
      method,
      data,
      headers: {
        'x-user-id': userID || 'default',
        'Content-Type': 'application/json',
      },
      success: (res) => {
        if (res.status === 401) { my.alert({ title: '提示', content: '请先登录' }); reject(new Error('unauthorized')); return; }
        if (res.status === 404) { reject(new Error('not_found')); return; }
        if (res.status >= 400) {
          const msg = (res.data && (res.data.msg || res.data.message)) || '请求失败';
          my.showToast({ content: msg, type: 'none' });
          reject(new Error(msg)); return;
        }
        const body = res.data;
        if (body.code !== 0) {
          my.showToast({ content: body.msg || body.message || '服务异常', type: 'none' });
          reject(new Error(body.msg)); return;
        }
        resolve(body.data !== undefined ? body.data : body);
      },
      fail: () => { my.showToast({ content: '网络异常', type: 'none' }); reject(new Error('network_error')); },
      complete: () => { my.hideLoading(); },
    });
  });
}

function getProfile(u)           { return request('GET', '/v1/profile', null, u); }
function updateProfile(u, d)     { return request('PUT', '/v1/profile', d, u); }
function queryPolicies(u, opts) {
  let p = '/v1/policies'; const q = [];
  if (opts) {
    if (opts.regionCode) q.push('region_code=' + opts.regionCode);
    if (opts.policyType) q.push('policy_type=' + opts.policyType);
  }
  if (q.length) p += '?' + q.join('&');
  return request('GET', p, null, u);
}
function generatePlan(u, d)      { return request('POST', '/v1/plans/generate', d, u); }
function getPlanDetail(u, pid)   { return request('GET', '/v1/plans/' + pid, null, u); }
function getCompliance(u, city)  { return request('GET', '/v1/compliance/checklist?city_code=' + city, null, u); }
function getGuide(u, city)       { return request('GET', '/v1/guide?city_code=' + city, null, u); }
function getPaymentStatus(u)     { return request('GET', '/v1/rights/payment-status', null, u); }
function getAlerts(u)            { return request('GET', '/v1/rights/alerts', null, u); }
function markAlertRead(u, id)    { return request('POST', '/v1/rights/alerts/read', { alert_id: id }, u); }
function submitFeedback(u, d)    { return request('POST', '/v1/feedback', d, u); }
function createOrder(u, planId) { return request('POST', '/v1/orders', { plan_id: planId }, u); }
function payOrder(u, oid)       { return request('POST', '/v1/orders/' + oid + '/pay', { payment_method: 'wechat' }, u); }
function checkUnlock(u, planId) { return request('GET', '/v1/orders/check-unlock?plan_id=' + planId, null, u); }
function getSettings(u, data)              { return request('GET', '/v1/settings', null, u); }
function saveSettings(u, data)             { return request('POST', '/v1/settings', data, u); }
function calculateSimulator(u, data)       { return request('POST', '/v1/simulator/calculate', data, u); }
function askAdvisor(u, data)               { return request('POST', '/v1/advisor/ask', data, u); }

module.exports = { API_BASE_URL, getProfile, updateProfile, queryPolicies, generatePlan, getPlanDetail,
  getCompliance, getGuide, getPaymentStatus, getAlerts, markAlertRead, submitFeedback,
  createOrder, payOrder, checkUnlock,
  getSettings, saveSettings, calculateSimulator, askAdvisor };
