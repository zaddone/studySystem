var util = require('../../utils/md5.js')
//var tencoding = require('../../text-encoding/index.js')  
//var comm = require('../../utils/comm.js')  
var app = getApp();

// miniprogram/pages/search/search.js
Page({
  /**
   * 页面的初始数据
   */
  data: {
    statusBarHeight: app.globalData.sysinfo.statusBarHeight,
    windowsHeight: app.globalData.sysinfo.screenHeight,
    windowsWidth: app.globalData.sysinfo.screenWidth,
    show: "none",
    adId: app.globalData.adId,
    //listshow: "hidden",
    theme: app.theme_start(),
    top: app.globalData.sysinfo.statusBarHeight + 10,
    focus:true,
  },

  /**
   * 生命周期函数--监听页面加载
   */
  showtts() {
    //if (!app.globalData.access_token) return;

    this.setData({ show: "block" })
    //console.log("showtts", app.globalData.access_token)

  },
  inputid(){
    this.setData({
      input:"",
      focus: true,
    })
  },
  navBotton() {
    if (getCurrentPages().length > 1)
      wx.navigateBack({
        delta: 1
      })
    else {
      wx.reLaunch({ url: "/pages/start/index" })
    }
  },
  toStart(event) {
    wx.reLaunch({ url: "/pages/start/index" })
  },
  goBack() {
    wx.navigateBack({
      delta: 1
    })
  },
  quest: function (n, t) {
    if (t === 0) {
      wx.hideLoading()
      return
    }
    let that = this
    wx.request({
      url: 'https://api.weixin.qq.com/cgi-bin/media/voice/queryrecoresultfortext?access_token=' + app.globalData.access_token.token + '&voice_id=' + n,
      method: "POST",
      success(r) {
        if (!r.data.is_end) {
          setTimeout(that.quest, 1000, n, t - 1)
          return
        }
        wx.hideLoading()
        //console.log(r.data)
        that.setData({ input: r.data.result })
        that.onSearch({ detail: { value: r.data.result } })
      },
      fail() {
        wx.hideLoading()
      }
    })
  },
  onLoad: function (options) {
    this.recorderManager = wx.getRecorderManager()
    let that = this
    app.getDBColl(function(){
      that.setData({
      adId: app.globalData.adId
      })
    })
    //console.log(options)
    if (options.words && options.words.length>0){
      this.setData({ input: options.words })
      this.onSearch({ detail: { value: options.words } })
      return
    }   

    this.recorderManager.onStart(() => {
      wx.vibrateShort()
      //console.log(res)
    })
    this.recorderManager.onStop((res) => {
      wx.showLoading({
        title: '稍等',
      })
      const { tempFilePath } = res
      let n = util.hexMD5(tempFilePath)
      console.log(app.globalData.access_token)
      wx.uploadFile({
        url: 'https://api.weixin.qq.com/cgi-bin/media/voice/addvoicetorecofortext?access_token=' + app.globalData.access_token.token + '&format=mp3&voice_id=' + n,
        filePath: tempFilePath,
        name: n,
        success(res) {
          console.log(res)

          setTimeout(that.quest, 1000, n, 8)

        },
        fail() {
          wx.hideLoading()
        }
      })
      console.log('recorder stop', tempFilePath)
    })


    let info = wx.getMenuButtonBoundingClientRect()
    //let pageLen = getCurrentPages().length
    //console.log(pageLen)
    if (info){
    this.setData({
      narHeight: info.bottom + 5,
      natTop: info.top + info.height / 2,
      top: info.top,
      pageLen: getCurrentPages().length,
    })
  }


  },
  
  touchEnd(event) {
    wx.stopRecord()
    //this.recorderManager.stop();
  },

  touchStart(event) {
    
    console.log(event)
    this.setData({ input: "" })
    let that = this
    wx.startRecord({
      success(res) {
        console.log(res)
        app.soundToTxt(res.tempFilePath, function (txt) {
          that.setData({ input: txt })
          that.onSearch({ detail: { value: txt} })
        })
      },
      fail: console.error,
    })
    setTimeout(function () {
      wx.stopRecord()
    }, 10000)
    
  },
  setPagelist(data){
    let li = []
    let vod = []
    //let i = 0
    //for (let d of res.result.data) {
      for (let d of data) {
      //if (i === 0) playtts(d.title)
      //i++
      if (!wx.getStorageSync(d._id)) {
        app.addDBList("list", d._id);
      }
      let _d = app.toDBStyle(d)
      wx.setStorageSync(_d._id, _d);
      if (_d.vod) {
        vod.push({ _id: _d._id, title: _d.title + "(" + _d.vod.length + ")" })
      } else {
        li.push({ _id: _d._id, title: _d.title })
      }
      //console.log(_d)
    }
     

    let tmpli = { list: li, vod: vod }
    
    this.setData(tmpli, wx.hideLoading())
    app.globalData.searchMap[this.input] = tmpli;
  },
  onSearch(event) {
    //console.log(event);
    //this.setData({ input:""})
    this.setData({ show: "none" })
    
    if (!event.detail.value) return
    this.input = event.detail.value
    wx.showLoading({
      title: '搜索中',
    })
    

    let tmpli = app.globalData.searchMap[event.detail.value]
    if (tmpli) {     
      //console.log("tmpli")
      if (tmpli.list.length>0){
        playtts(tmpli.list[parseInt(Math.random() * tmpli.list.length)].title)
      }
      this.setData(tmpli, wx.hideLoading())
      return
    }
    let that = this      
      wx.cloud.callFunction({
        name: 'search',
        data: {
          words: app.getSearchKey(that.input),
        },
        success: function (res) {
          console.log(res)
          if (!res || !res.result || !res.result) {
            wx.hideLoading()
            return
          }
          that.setPagelist(res.result)
          
        },
        fail: function (res) {
          console.error(res)
          wx.hideLoading()
        },
      })
 
  },
  showVodlist(event) {
    console.log(event)
  },
  gohome: function () {
    wx.reLaunch({
      url: '../start/index'
    })
  },
  onShareAppMessage:function(res){
    if (!this.input) return
    return {
      title: this.input,
      path: '/pages/search/search?words=' + this.input
    }
  },
  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady: function () {
    //playtts("今天已经太晚了！")

  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow: function () {

    /* 
    let that = this
    wx.onKeyboardHeightChange(res => {
      console.log(res.height)
      if (res.height>0){
        that.setData({ kbh: res.height })     
      }      
    })
    */
  },

  /**
   * 生命周期函数--监听页面隐藏
   */
  onHide: function () {

  },

  /**
   * 生命周期函数--监听页面卸载
   */
  onUnload: function () {

  },

  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh: function () {
    this.navBotton()
  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom: function () {

  },


})