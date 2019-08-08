var app = getApp();
Page({
  data: {
    pageLen: getCurrentPages().length,
    isOpen: false,
    adId: app.globalData.adId,
    theme: app.theme_start(),
    windowsHeight: app.globalData.sysinfo.screenHeight,
    animationData: {},
    isUp: false
  },
  onShow: function () {
    //console.log(this.data.pageLen)
    let info = wx.getMenuButtonBoundingClientRect()
    this.sTop = 0;
    this.setData({
      narHeight: info.bottom + 5,
      natTop: info.top + info.height / 2,
      top: info.top,
    })
    var animation = wx.createAnimation({
      duration: 300,
      timingFunction: 'linear',
    })
    this.animation = animation
  },
  onPageScroll: function (object) {
    if (!this.data.isUp && (object.scrollTop - this.sTop) > 10) {
      this.animation.translate(0, -this.data.narHeight).step({ duration: 200 })
      this.setData({
        animationData: this.animation.export()
      })
      this.data.isUp = true;
    } else if (this.data.isUp && ((this.sTop - object.scrollTop) > 10 || object.scrollTop === 0)) {

      this.animation.translate(0, 0).step({ duration: 200 })
      this.setData({
        animationData: this.animation.export()
      })
      this.data.isUp = false;
    }
    this.sTop = object.scrollTop
    //console.log(object)
  },
  __bind_touchmove(event) {
  },
  __bind_touchend(event) {
  },
  __bind_touchstart(event) {
  },
  __bind_tap(event) {
    if (this.data.isOpen) return;
    this.data.isOpen = true;
    let that = this;
    if (event.target.dataset._el.tag != "image") return;
    let url = event.target.dataset._el.attr.src;
    wx.previewImage({
      current: url,
      urls: [url],
      success: function () {
        that.data.isOpen = false;
      },
    })
  },
  loadDB: function (event) {
    wx.redirectTo({
      url: '/pages/page/index?id=' + event.target.id,
    })
  },
  loadVod: function (event) {
    wx.redirectTo({
      url: '/pages/vod/vodshow?id=' + event.target.id,
    })
  },

  goBack() {
    wx.navigateBack({
      delta: 1
    })
  },

  onPullDownRefresh() {
    //wx.reLaunch({ url: "/pages/search/search" })

    //return
    if (getCurrentPages().length > 1)
      wx.navigateBack({
        delta: 1
      })
    else {
      wx.reLaunch({ url: "/pages/search/search" })
    }

  },
  onLoad(obj) {
    wx.showLoading({
      title: "loading"
    })
    if (obj && obj.id) {

      app.getPageDBE(obj.id, this.viewPage);
      return;
    }
    wx.reLaunch({ url: "/pages/search/search" })
  },
  toSearch(event) {
    wx.reLaunch({ url: "/pages/search/search" })
  },
  viewPage: function (db) {
    //console.log(db)
    db.article.theme = app.theme();
    let that = this
    app.getDBColl(function () {
      that.setData({
        db: db,
        theme: app.theme(),
        adId: app.globalData.adId,
        link: [],
        vod: [],
        pageLen: getCurrentPages().length,
      }, wx.hideLoading());
      wx.pageScrollTo({
        scrollTop: 0,
        duration: 0,
      })

    })
  },
  setDBTolink(_d){
    if (_d.vod) {
      this.data.vod.push({ _id: _d._id, title: _d.title + "(" + _d.vod.length + ")" })
      this.setData({ vod: this.data.vod })
    } else {
      this.data.link.push({ _id: _d._id, title: _d.title })
      this.setData({ link: this.data.link })
    }      
  },  
  

  onReachBottom() {
    if (this.data.link && this.data.link.length > 0) {
      return;
    }
    if (this.data.vod && this.data.vod.length > 0) {
      return;
    }
    if (this.data.db) {
      this.onSign()
      app.showlink(this.data.db, this.setDBTolink)
    }
  },
  onSign: function () {
    if (this.data.db.sign) return
    this.data.db.sign = true
    wx.setStorage({
      key: this.data.db._id,
      data: this.data.db,
    });
  },
  onShareAppMessage: function (res) {
    //if (!this.input) return
    return {
      title: this.data.db.title,
      path: '/pages/page/index?id=' + this.data.db._id,
    }
  },
})