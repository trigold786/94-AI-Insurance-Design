const { EMPLOYMENT_STATUS, CITIES } = require('../../utils/constants')
const api = require('../../services/api')

Page({
  data: {
    stepIndex: 1,
    submitting: false,
    employmentOptions: EMPLOYMENT_STATUS,
    cities: CITIES,
    formData: {
      age: '',
      gender: '',
      hukou: '',
      residence: '',
      employment: '',
      socialYears: '',
      unemploymentDate: '',
      hasChildren: '',
      childrenCount: 0,
      hasSkills: ''
    }
  },
  onLoad() {
    const app = getApp()
    if (app.globalData.profileData && Object.keys(app.globalData.profileData).length > 0) {
      this.setData({ formData: { ...this.data.formData, ...app.globalData.profileData } })
    }
    if (app.globalData.currentCity) {
      this.setData({
        'formData.hukou': app.globalData.currentCity.displayName,
        'formData.residence': app.globalData.currentCity.displayName
      })
    }
  },
  onInput(e) {
    const field = e.currentTarget.dataset.field
    this.setData({ ['formData.' + field]: e.detail.value })
  },
  setGender(e) {
    this.setData({ 'formData.gender': e.currentTarget.dataset.value })
  },
  setEmployment(e) {
    this.setData({ 'formData.employment': e.currentTarget.dataset.value })
  },
  setHasChildren(e) {
    this.setData({ 'formData.hasChildren': e.currentTarget.dataset.value })
  },
  setChildrenCount(e) {
    this.setData({ 'formData.childrenCount': parseInt(e.currentTarget.dataset.value) })
  },
  setSkills(e) {
    this.setData({ 'formData.hasSkills': e.currentTarget.dataset.value })
  },
  setUnemploymentDate(e) {
    this.setData({ 'formData.unemploymentDate': e.detail.value })
  },
  onPickHukou() {
    this.showCityPicker('hukou')
  },
  onPickResidence() {
    this.showCityPicker('residence')
  },
  showCityPicker(field) {
    const cityList = this.data.cities.map(c => c.displayName)
    wx.showActionSheet({
      itemList: cityList,
      success: (res) => {
        if (res.tapIndex >= 0 && res.tapIndex < cityList.length) {
          this.setData({ ['formData.' + field]: cityList[res.tapIndex] })
        }
      }
    })
  },
  validateStep() {
    const d = this.data.formData
    if (this.data.stepIndex === 1) {
      const age = parseInt(d.age)
      if (!d.age || isNaN(age) || age < 16 || age > 70) {
        wx.showToast({ title: '请输入有效年龄（16-70岁）', icon: 'none' })
        return false
      }
      if (!d.gender) {
        wx.showToast({ title: '请选择性别', icon: 'none' })
        return false
      }
      if (!d.hukou) {
        wx.showToast({ title: '请选择户籍所在地', icon: 'none' })
        return false
      }
      if (!d.residence) {
        wx.showToast({ title: '请选择居住地', icon: 'none' })
        return false
      }
    } else if (this.data.stepIndex === 2) {
      if (!d.employment) {
        wx.showToast({ title: '请选择就业状态', icon: 'none' })
        return false
      }
      const years = parseInt(d.socialYears)
      if (!d.socialYears || isNaN(years) || years < 0 || years > 50) {
        wx.showToast({ title: '请输入有效社保年限（0-50年）', icon: 'none' })
        return false
      }
      if (d.employment === 'unemployed' && !d.unemploymentDate) {
        wx.showToast({ title: '请选择失业登记日期', icon: 'none' })
        return false
      }
    } else if (this.data.stepIndex === 3) {
      if (!d.hasChildren) {
        wx.showToast({ title: '请选择是否有子女', icon: 'none' })
        return false
      }
      if (!d.hasSkills) {
        wx.showToast({ title: '请选择是否有技能证书', icon: 'none' })
        return false
      }
    }
    return true
  },
  nextStep() {
    if (!this.validateStep()) return
    const app = getApp()
    app.globalData.profileData = { ...this.data.formData }
    if (this.data.stepIndex < 3) {
      this.setData({ stepIndex: this.data.stepIndex + 1 })
    } else {
      this.submitProfile()
    }
  },
  prevStep() {
    if (this.data.stepIndex > 1) {
      this.setData({ stepIndex: this.data.stepIndex - 1 })
    }
  },
  goStep(e) {
    const step = parseInt(e.currentTarget.dataset.step)
    if (step < this.data.stepIndex) {
      this.setData({ stepIndex: step })
    }
  },
  submitProfile() {
    this.setData({ submitting: true })
    const data = this.data.formData
    const payload = {
      age: parseInt(data.age),
      gender: data.gender,
      hukou: data.hukou,
      residence: data.residence,
      employment: data.employment,
      social_years: parseInt(data.socialYears),
      unemployment_date: data.unemploymentDate || null,
      has_children: data.hasChildren === 'yes',
      children_count: data.childrenCount || 0,
      has_skills: data.hasSkills === 'yes'
    }
    const app = getApp()
    app.globalData.stepData = payload
    api.saveProfile(payload).then((res) => {
      this.setData({ submitting: false })
      wx.redirectTo({ url: '/pages/loading/loading' })
    }).catch(() => {
      this.setData({ submitting: false })
      wx.showToast({ title: '提交失败，请检查网络后重试', icon: 'none' })
    })
  }
})
