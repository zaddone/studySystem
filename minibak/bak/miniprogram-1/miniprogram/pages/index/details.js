// miniprogram/pages/index/details.js
Page({

  /**
   * 页面的初始数据
   */
  data: {

  },
  handleContact(e) {
    console.log(e.detail.path)
    console.log(e.detail.query)
  },
  /**
   * 生命周期函数--监听页面加载
   */
  onLoad: function (options) {
    //console.log(options)
    var pages = getCurrentPages();
    var prevPage = pages[pages.length - 2]
    let tab = prevPage.data.tabs[options.tab]
    //console.log(tab)
    let db = tab.db[options.id]
    console.log(db)
    this.setData({ db: db, tab: tab})
    wx.setNavigationBarTitle({ title: db.title })
  },

  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady: function () {

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

  }
})