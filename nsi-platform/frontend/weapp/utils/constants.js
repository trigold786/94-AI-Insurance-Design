const CITIES = [
  { code: '310000', name: '上海' },
  { code: '110000', name: '北京' },
  { code: '440300', name: '深圳' },
  { code: '440100', name: '广州' },
  { code: '330100', name: '杭州' },
];

const EMPLOYMENT_STATUS = [
  { value: 'employed', label: '企业就业' },
  { value: 'flexible', label: '灵活就业' },
  { value: 'self_employed', label: '自雇' },
  { value: 'unemployed', label: '失业' },
];

const GENDER_OPTIONS = [
  { value: 'male', label: '男' },
  { value: 'female', label: '女' },
];

const POLICY_TYPES = [
  { type: 'subsidy', label: '补贴政策' },
  { type: 'training', label: '培训政策' },
  { type: 'pension', label: '养老保险' },
];

module.exports = { CITIES, EMPLOYMENT_STATUS, GENDER_OPTIONS, POLICY_TYPES };
