const app = getApp();
const api = require('../../services/api');
const { CITIES, GENDER_OPTIONS, EMPLOYMENT_STATUS } = require('../../utils/constants');

const DEBOUNCE_MS = 150;

Page({
  data: {
    cities: CITIES,
    genderOptions: GENDER_OPTIONS,
    employmentOptions: EMPLOYMENT_STATUS,
    cityIndex: 0,
    genderIndex: 0,
    employmentIndex: 0,
    age: 30,
    basePercent: 100,
    paidYears: 10,
    planYears: 15,
    hukou: true,
    calculating: false,
    result: {
      monthly_cost: 0,
      monthly_pension: 0,
      subsidy: 0,
      net: 0,
      triggers: [],
      chart: [],
    },
  },
  onLoad() {
    this.scheduleCalc();
  },
  getUserID() {
    return app.globalData.userInfo ? app.globalData.userInfo.nickName : 'default';
  },
  onCityChange(e) {
    this.setData({ cityIndex: Number(e.detail.value) });
    this.scheduleCalc();
  },
  onGenderChange(e) {
    this.setData({ genderIndex: Number(e.detail.value) });
    this.scheduleCalc();
  },
  onEmploymentChange(e) {
    this.setData({ employmentIndex: Number(e.detail.value) });
    this.scheduleCalc();
  },
  onAgeChanging(e) { this.setData({ age: e.detail.value }); },
  onAgeChange() { this.scheduleCalc(); },
  onBaseChanging(e) { this.setData({ basePercent: e.detail.value }); },
  onBaseChange() { this.scheduleCalc(); },
  onPaidChanging(e) { this.setData({ paidYears: e.detail.value }); },
  onPaidChange() { this.scheduleCalc(); },
  onPlanChanging(e) { this.setData({ planYears: e.detail.value }); },
  onPlanChange() { this.scheduleCalc(); },
  onHukouChange(e) {
    this.setData({ hukou: e.detail.value });
    this.scheduleCalc();
  },
  scheduleCalc() {
    if (this._timer) clearTimeout(this._timer);
    this._timer = setTimeout(() => { this.calc(); }, DEBOUNCE_MS);
  },
  calc() {
    this.setData({ calculating: true });
    const payload = {
      city_code: CITIES[this.data.cityIndex].code,
      gender: GENDER_OPTIONS[this.data.genderIndex].value,
      age: this.data.age,
      base_percent: this.data.basePercent,
      paid_years: this.data.paidYears,
      plan_years: this.data.planYears,
      employment: EMPLOYMENT_STATUS[this.data.employmentIndex].value,
      hukou: this.data.hukou,
    };
    api.calculateSimulator(this.getUserID(), payload)
      .then((data) => {
        this.setData({ calculating: false, result: this.normalize(data) });
      })
      .catch(() => { this.setData({ calculating: false }); });
  },
  normalize(data) {
    const d = data || {};
    const chart = (d.chart || []).map((c) => ({
      label: c.label,
      value: c.value,
      height: 8 + Math.round((Number(c.value) || 0) / (d.chart_max || 1) * 220),
      color: c.color || '#1A56DB',
    }));
    return {
      monthly_cost: d.monthly_cost || 0,
      monthly_pension: d.monthly_pension || 0,
      subsidy: d.subsidy || 0,
      net: d.net || 0,
      triggers: d.triggers || [],
      chart,
      chart_max: d.chart_max || 1,
    };
  },
});
