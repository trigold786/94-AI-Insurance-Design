Component({
  properties: {
    title: { type: String, value: '' },
    description: { type: String, value: '' },
    icon: { type: String, value: '📋' },
    iconBg: { type: String, value: '#EBF0FF' },
    tags: { type: Array, value: [] }
  },
  methods: {
    onTap() {
      this.triggerEvent('tap', { title: this.properties.title })
    }
  }
})
