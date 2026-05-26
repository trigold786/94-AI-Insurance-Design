const CITIES = [
  { code: '110000', name: '北京', displayName: '北京市' },
  { code: '310000', name: '上海', displayName: '上海市' },
  { code: '440100', name: '广州', displayName: '广州市' },
  { code: '330100', name: '杭州', displayName: '杭州市' },
  { code: '440300', name: '深圳', displayName: '深圳市' }
]

const POLICY_TYPE_MAP = {
  yanglao: { label: '养老保险', icon: '🔵', color: '#1A56DB' },
  yiliao: { label: '医疗保险', icon: '🟢', color: '#059669' },
  shiye: { label: '失业保险', icon: '🟠', color: '#F59E0B' },
  gongshang: { label: '工伤保险', icon: '🔴', color: '#EF4444' },
  shengyu: { label: '生育保险', icon: '🟣', color: '#8B5CF6' },
  zhufang: { label: '住房公积金', icon: '🔷', color: '#06B6D4' }
}

const EMPLOYMENT_STATUS = [
  { value: 'employed', label: '在职职工' },
  { value: 'unemployed', label: '失业人员' },
  { value: 'freelance', label: '灵活就业' },
  { value: 'retired', label: '已退休' },
  { value: 'other', label: '其他' }
]

const GENDER_OPTIONS = [
  { value: 'male', label: '男' },
  { value: 'female', label: '女' }
]

const YES_NO_OPTIONS = [
  { value: 'yes', label: '是' },
  { value: 'no', label: '否' }
]

const LOADING_MESSAGES = [
  'AI 正在读取您的社保档案...',
  '正在分析城市政策数据库...',
  '智能匹配最优缴费方案...',
  '计算未来养老金收益...',
  '生成个性化社保规划报告...',
  '方案优化中，请稍候...'
]

const PLAN_FEATURES = [
  { icon: '🛡️', title: '社保安全盾', desc: '全方位保障您的社保权益' },
  { icon: '📊', title: '智能分析', desc: '基于AI的个性化规划方案' },
  { icon: '💰', title: '节省优化', desc: '平均每年节省30%缴费成本' },
  { icon: '📋', title: '合规保障', desc: '严格遵守各地社保政策法规' }
]

module.exports = {
  CITIES,
  POLICY_TYPE_MAP,
  EMPLOYMENT_STATUS,
  GENDER_OPTIONS,
  YES_NO_OPTIONS,
  LOADING_MESSAGES,
  PLAN_FEATURES
}
