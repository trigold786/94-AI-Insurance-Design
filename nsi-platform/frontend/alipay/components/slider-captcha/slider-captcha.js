Component({
  properties: {},
  data: {
    verified: false,
    left: 0,
    maxLeft: 0,
    text: '请拖动滑块验证',
  },
  didMount() {
    const q = my.createSelectorQuery().in(this);
    q.select('#track').boundingClientRect();
    q.select('#handle').boundingClientRect();
    q.exec((res) => {
      const track = res[0];
      const handle = res[1];
      if (track && handle) {
        this.setData({ maxLeft: Math.max(0, track.width - handle.width) });
      }
    });
  },
  methods: {
    onChange(e) {
      if (this.data.verified) return;
      this._currentX = e.detail.x;
    },
    onTouchEnd() {
      if (this.data.verified) return;
      const x = this._currentX || 0;
      if (x >= this.data.maxLeft - 6) {
        this.setData({ verified: true, left: this.data.maxLeft, text: '验证成功' });
        this.triggerEvent('verified', { success: true });
      } else {
        this.setData({ left: 0 });
        this._currentX = 0;
      }
    },
    reset() {
      this.setData({ verified: false, left: 0, text: '请拖动滑块验证' });
      this._currentX = 0;
    },
  },
});
