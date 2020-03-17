// miniprogram/pages/index/shopping/pinduoduo.js
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
      app.getGoodsData('pinduoduo', options.goods, function (db) {
        that.setData({
          tab: { py: 'pinduoduo' },
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
  getRouter:function(e){
    //console.log(e)
    //return;
    wx.request({
      url: 'https://www.zaddone.com/goods/'+this.data.tab.py,
      data: { goodsid: e.currentTarget.dataset.id, session: app.globalData.userInfo.OPENID},
      success:function(res){
        let r = res.data.goods_promotion_url_generate_response.goods_promotion_url_list[0].we_app_info
        wx.navigateToMiniProgram({
          appId:r.app_id,
          path:r.page_path,
        })
        //console.log()
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
      path: '/pages/index/shopping/pinduoduo?goods=' + this.data.db.Id
    }
  }
})