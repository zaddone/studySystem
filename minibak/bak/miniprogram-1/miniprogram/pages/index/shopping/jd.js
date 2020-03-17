// miniprogram/pages/index/shopping/jd.js
var app = getApp();
Page({

  /**
   * 页面的初始数据
   */
  data: {

  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad: function (options) {
    console.log(options)
    if (options.goods) {
      let that = this
      app.getGoodsData('jd', options.goods, function (db) {
        that.setData({
          tab: { py: 'jd' },
          db: db,
        })
        wx.setNavigationBarTitle({ title: db.Tag })
      }, function () {
        wx.navigateTo({
          url: '/pages/index/index',
        })
      })
      return
    }
    wx.navigateTo({
      url: '/pages/index/index',
    })        
  },
  searchTap() {
    wx.navigateTo({
      url: '/pages/index/list?q=' + this.data.db.Name,
    })
  },
  getRouter: function (e) {
    
    wx.request({
      url: 'https://www.zaddone.com/goods/' + this.data.tab.py,
      data: { goodsid: e.currentTarget.dataset.id, session: app.globalData.userInfo.OPENID },
      success: function (res) {
        console.log(res)
        let uri = res.data.jd_kpl_open_promotion_pidurlconvert_response.clickUrl.shortUrl;
        //pages/proxy/union/union?isUnion=1&spreadUrl=
        console.log(uri)
        wx.navigateToMiniProgram({
          appId: 'wx1edf489cb248852c',
          path:"pages/proxy/union/union?spreadUrl="+uri,
        })
      },
    })
    
  },
  showuser(e) {
    wx.navigateTo({
      url: '/pages/user/form',
    })
  },
  showorder(e) {
    wx.navigateTo({
      url: '/pages/user/order',
    })
  },
  showhelp(e) {
    wx.navigateTo({
      url: '/pages/user/help',
    })
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
    return {
      title: this.data.db.Name,
      path: '/pages/index/shopping/jd?goods=' + this.data.db.Id
    }
  }
})