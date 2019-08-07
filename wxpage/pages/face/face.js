var app = getApp();
Page({

  /**
   * Page initial data
   */
  data: {
    height: app.globalData.sysinfo.windowHeight,
    width: app.globalData.sysinfo.windowWidth,
    scrolltop: 1,
    buttonTop: app.globalData.sysinfo.windowHeight / 2,
    buttonLeft: app.globalData.sysinfo.windowWidth / 3 * 2,
    src: "/images/th.jpeg",
    text: [{ txt: "我们聊会儿吧", self: "" }],
  },
  loadDB() {
    console.log("text")
  },
  toSearch() {
    let that = this
    let input = ""
    for (let i = that.data.text.length - 1; i >= 0; i--) {
      let l = that.data.text[i]
      console.log(l)
      if (!l.id) {
        input += l.txt + " "
        if (input.length > 30) break
      }
    }

    if (input.length === 0) return
    wx.cloud.callFunction({
      name: 'searcheasy',
      data: {
        words: app.getSearchKey(input),
      },
      success: function (res) {
        if (res.result.length === 0) return
        let id_ = res.result[0]
        for (let l of that.data.text) {
          if (l.id) {
            if (l.id == id_) return
          }
        }
        app.getPageDBE(id_, function (b) {
          if (b.vod) {
            that.data.text.push({ txt: b.title, self: "", vod: true, id: b._id })
          } else {
            that.data.text.push({ txt: b.title, self: "", id: b._id })
          }

          that.setData({
            text: that.data.text
          }, that.setBottom)
        })
        console.log(res.result)

      },
      fail: console.error,
    })
  },
  setBottom() {
    //if (this.int) clearInterval(this.int)
    let that = this
    setTimeout(function () {
      that.setData({
        bottomid: "bottomid",
         
      })
    }, 1000)
  },
  runRecord() {
    wx.showLoading({
      title: "说话吧",
    })
    let that = this;
    wx.vibrateShort()
    wx.startRecord({
      success(res) {
        console.log(res)
        app.soundToTxt(res.tempFilePath, function (txt) {
          if (txt.length < 2) return

          app.textChat(txt, function (ftxt) {
            app.playtts(ftxt)
            that.data.text.push({ txt: ftxt, self: "" })
            that.toSearch();
            that.setData({
              text: that.data.text,
            }, that.setBottom)
          })
          that.data.text.push({ txt: txt, self: "self" })
          that.setData({
            text: that.data.text,
          }, that.setBottom)
        })
      },
      fail: console.error,
    })
    setTimeout(function () {
      wx.stopRecord({ success: wx.hideLoading() })
    }, 10000)
  },
  touchStart() {
    let that = this
    //this.startT = 0
    wx.getSetting({
      success(res) {
        if (!res.authSetting['scope.record']) {
          wx.authorize({
            scope: 'scope.record',
          })
        } else {
          that.runRecord()
          //that.startT = setTimeout(that.runRecord, 200)
          //that.runRecord()   
          //setTimeout(that.runRecord(),1000)
        }
      }
    })
    console.log("touchStart")
  },
  touchEnd() {
    wx.stopRecord({ success: wx.hideLoading() })
    console.log("touchEnd")
  },
  touchMove() {
    //clearTimeout(this.startT)
  },
  /**
   * Lifecycle function--Called when page load
   */
  getDateDB() {
    let that = this
    wx.request({
      url: "https://www.sojson.com/open/api/lunar/json.shtml",
      success: function (res) {
      }
    })
  },
  onLoad: function (options) {
    let that = this
    wx.request({
      url: "https://rest.shanbay.com/api/v2/quote/quotes/today/",
      success: function (res) {
        console.log(res)
        //let src = res.data.data.origin_img_urls[0].replace(/png.*/, "png");
        //console.log(src)
        //let textlist = []         
        //that.data.text.push({ txt: res.data.data.translation, self: "" })        
        that.setData({
          src: res.data.data.origin_img_urls[0].replace(/png.*/, "png"),
          text: [{ txt: res.data.data.translation, self: "" }],

          //scrollTop: that.data.height,
        }, that.toSearch())

      },

    })
    //https://rest.shanbay.com/api/v2/quote/quotes/today/
  },
  scrolltolower: function (e) {
    //console.log('stop', e)
    //clearInterval(this.int)

  },
  /**
   * Lifecycle function--Called when page is initially rendered
   */
  onReady: function () {
  },

  /**
   * Lifecycle function--Called when page show
   */
  onShow: function () {

  },

  /**
   * Lifecycle function--Called when page hide
   */
  onHide: function () {

  },

  /**
   * Lifecycle function--Called when page unload
   */
  onUnload: function () {

  },

  /**
   * Page event handler function--Called when user drop down
   */

  onPullDownRefresh: function () {
    wx.reLaunch({
      url: "/pages/search/search",
    })
  },

  /**
   * Called when page reach bottom
   */
  onReachBottom: function () {

  },

  /**
   * Called when user click on the top right corner to share
   */
  onShareAppMessage: function () {
    return {
      title:'我们聊会儿吧',
      path: '/pages/face/face'
    }
  }
})