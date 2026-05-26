Component({
  properties: {
    visible: {
      type: Boolean,
      value: false
    }
  },
  data: {
    privacyChecked: false,
    logining: false
  },
  methods: {
    togglePrivacy() {
      this.setData({ privacyChecked: !this.data.privacyChecked })
    },
    onLogin() {
      if (!this.data.privacyChecked) {
        wx.showToast({ title: '请先同意隐私协议', icon: 'none' })
        return
      }
      this.setData({ logining: true })
      this.triggerEvent('login')
    },
    onCancel() {
      this.triggerEvent('cancel')
    },
    onBrowse() {
      this.triggerEvent('browse')
    },
    stopPropagation() {},
    onViewPrivacy() {
      wx.showModal({
        title: '隐私政策',
        content: '我们重视您的隐私安全，收集的信息仅用于社保规划服务，不会向第三方共享您的个人信息。',
        showCancel: false
      })
    },
    onViewAgreement() {
      wx.showModal({
        title: '用户协议',
        content: '使用AI社保智筹即表示您同意本服务条款。我们提供社保规划建议，不构成正式法律意见。',
        showCancel: false
      })
    }
  }
})
