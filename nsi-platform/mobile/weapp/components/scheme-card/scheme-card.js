Component({
  properties: {
    name: { type: String, value: '' },
    baseAmount: { type: String, value: '0' },
    monthlyPayment: { type: String, value: '0' },
    pension: { type: String, value: '0' },
    saving: { type: String, value: '' },
    badge: { type: String, value: '' },
    badgeColor: { type: String, value: '#1A56DB' },
    selected: { type: Boolean, value: false }
  },
  methods: {
    onSelect() {
      this.triggerEvent('select', { name: this.properties.name })
    }
  }
})
