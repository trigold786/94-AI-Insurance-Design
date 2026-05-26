const app = getApp();
const { CITIES } = require('../../utils/constants');

Page({
  data: {
    cities: CITIES,
    currentCode: app.globalData.currentCityCode || '310000',
  },
  onSelect(e) {
    const { code, name } = e.currentTarget.dataset;
    app.globalData.currentCity = name;
    app.globalData.currentCityCode = code;
    my.navigateBack();
  },
});
