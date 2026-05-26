App({
  globalData: {
    userInfo: null,
    currentCity: null,
    currentCityCode: null,
    planResult: null,
    profileData: null,
  },
  onLaunch() {
    this.globalData.userInfo = { nickName: 'default' };
    const { CITIES } = require('/utils/constants');
    this.globalData.currentCity = CITIES[0].name;
    this.globalData.currentCityCode = CITIES[0].code;
  },
});
