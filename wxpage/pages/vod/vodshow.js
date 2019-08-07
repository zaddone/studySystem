// miniprogram/pages/vod/vodshow.js
var app = getApp();
var util = require('../../utils/md5.js')
Page({
  /**
   * 页面的初始数据
   */
  data: {
    adId: app.globalData.adId,
    vod:[],
    link:[],
  },

  next(event) {
    this.playvod(this.voddb.indexid + 1)
  },
  playshow(event){
    this.playvod(event.currentTarget.id)
  },
  playvod(index) {
    if (!this.voddb) return
    
    //console.log(this.voddb, index)
    if (this.voddb.vod.length - 1 < index) return
    wx.setNavigationBarTitle({ "title": this.voddb.title })
    this.voddb.indexid = index;    
    wx.setStorageSync(this.voddb._id, this.voddb)
    let that = this
    for (let v of this.voddb.vod){
     v.played=""
    }
    this.voddb.vod[this.voddb.indexid].played='play'
   
    this.setData({ 
      src: this.voddb.vod[this.voddb.indexid].url,
      vodlist: this.voddb.vod,
     }, function () {
     
      that.v.play()
      //that.v.requestFullScreen()
       //console.log(this.voddb.vod)
    })
  },
  onLoad: function (options) {
   
   // return 
    if (!options || !options.id) {
          wx.reLaunch({ url: "/pages/search/search" })
      return
    }else{
      this.id = options.id
    }
    //wx.setClipboardData({ data: this.id })
    let that = this    
    app.getDBColl(function () {
      that.setData({
        adId: app.globalData.adId
      })
    })
   
    if (options.index) this.data.index = option.index;
    this.v = wx.createVideoContext('myVideo', this)

  },
  
  /**
   * 生命周期函数--监听页面初次渲染完成
   */
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

  onReady: function () {
    
    if (!this.id) return
    wx.setBackgroundColor({
      backgroundColor:'#000000',
      backgroundColorTop: '#000000',
      backgroundColorBottom: '#000000',
    })

    let that = this
    this.voddb = wx.getStorageSync(this.id)
    if (this.voddb && (Date.parse(new Date()) / 1000) < this.voddb.timeOut ) {      
      this.playvod(this.data.index || this.voddb.indexid || 0)      
      app.showlink(this.voddb, this.setDBTolink)  
      return
    }
    //console.log(this.id)
    app.downDB(this.id,
      function (d) {
        app.addDBList("list", d._id);
        that.voddb = app.toDBStyle(d)
        wx.setStorageSync(that.voddb._id, that.voddb);
        that.playvod(that.data.index || 0)
        app.showlink(that.voddb, that.setDBTolink)  
        //that.showlink(_d.link)
      },
      function () {
        //console.log("out",that.id)
        wx.redirectTo({
          url: '/pages/search/search',
        })
      })
  },
  setDBTolink(_d) {
    console.log(_d)
    if (_d.vod) {
      this.data.vod.push({ _id: _d._id, title: _d.title + "(" + _d.vod.length + ")" })
      console.log(this.data.vod)
      this.setData({ vod: this.data.vod })
    } else {
      this.data.link.push({ _id: _d._id, title: _d.title })
      this.setData({ link: this.data.link })
    }
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow: function () {

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

  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom: function () {

  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage: function () {
    return {
      title: this.voddb.title,
      path: '/pages/vod/vodshow?id=' + this.id
    }
  }
})