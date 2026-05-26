const app = getApp();
const api = require('../../services/api');
const { GENDER_OPTIONS, EMPLOYMENT_STATUS } = require('../../utils/constants');

Page({
  data: {
    currentStep: 1,
    genders: GENDER_OPTIONS.map(g => g.label),
    employmentStatus: EMPLOYMENT_STATUS.map(e => e.label),
    childOptions: ['无', '有'],
    age: '', gender: '', household: '',
    employment: '', years: '',
    hasChildrenText: '', budget: '', balance: '',
    submitting: false, error: '',
  },
  onShow() {
    this.loadProfile();
  },
  loadProfile() {
    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    const that = this;
    wx.request({
      url: 'http://127.0.0.1:39401/v1/profile',
      header: { 'x-user-id': userID },
      success(res) {
        if (res.statusCode === 200 && res.data.code === 0 && res.data.data) {
          const p = res.data.data;
          const genderLabel = GENDER_OPTIONS.find(g => g.value === p.gender);
          const employmentLabel = EMPLOYMENT_STATUS.find(e => e.value === p.employment_status);
          that.setData({
            age: String(p.age || ''),
            gender: p.gender || '',
            household: p.household_region_code || '',
            employment: p.employment_status || '',
            years: String(p.social_security_years || ''),
            hasChildrenText: p.has_children ? '有' : '无',
          });
        }
      },
      fail() {},
    });
  },
  onGenderChange(e) {
    this.setData({ gender: GENDER_OPTIONS[e.detail.value].value });
  },
  onAgeInput(e) {
    this.setData({ age: e.detail.value });
  },
  onRegionChange(e) {
    this.setData({ household: e.detail.value.join(' ') });
  },
  onEmploymentChange(e) {
    this.setData({ employment: EMPLOYMENT_STATUS[e.detail.value].value });
  },
  onYearsInput(e) {
    this.setData({ years: e.detail.value });
  },
  onChildChange(e) {
    this.setData({ hasChildrenText: e.detail.value === 0 ? '无' : '有' });
  },
  onBudgetInput(e) {
    this.setData({ budget: e.detail.value });
  },
  onBalanceInput(e) {
    this.setData({ balance: e.detail.value });
  },
  onNextStep() {
    if (this.data.currentStep < 3) {
      this.setData({ currentStep: this.data.currentStep + 1, error: '' });
    }
  },
  onPrevStep() {
    if (this.data.currentStep > 1) {
      this.setData({ currentStep: this.data.currentStep - 1, error: '' });
    }
  },
  onSubmit() {
    const age = parseInt(this.data.age);
    if (age < 16 || age > 70) {
      this.setData({ error: '年龄必须在16-70之间' });
      return;
    }
    const budget = parseFloat(this.data.budget);
    if (!budget || budget <= 0) {
      this.setData({ error: '请输入有效月预算' });
      return;
    }
    this.setData({ submitting: true, error: '' });

    const userID = app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
    const profileData = {
      age,
      gender: this.data.gender || 'male',
      household_region_code: app.globalData.currentCityCode,
      current_residence_code: app.globalData.currentCityCode,
      employment_status: this.data.employment || 'flexible',
      social_security_years: parseInt(this.data.years) || 0,
      has_children: this.data.hasChildrenText === '有',
    };

    api.updateProfile(userID, profileData).then(() => {
      app.globalData.profileData = profileData;
      app.globalData.planInput = {
        age,
        gender: this.data.gender || 'male',
        employment: this.data.employment || 'flexible',
        contribution_years: parseInt(this.data.years) || 0,
        current_balance: parseFloat(this.data.balance) || 0,
        monthly_budget: budget,
        local_avg_salary: budget * 2,
      };
      wx.redirectTo({ url: '/pages/loading/loading' });
    }).catch((err) => {
      this.setData({ submitting: false, error: err.message || '提交失败' });
    });
  },
});
