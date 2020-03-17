// miniprogram/pages/index/shopping/taobao.js
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
    //console.log(options)    
    if (options.goods) {
      let that = this
      app.getGoodsData('taobao', options.goods,function(db){
        console.log(db)
        that.setData({
          tab: { py: 'taobao' },
          db: db,
        } )
        wx.setNavigationBarTitle({ title: db.Tag })
      },function(){
        wx.navigateTo({
          url: '/pages/index/index',
        })
      })
      return
    }
    wx.navigateTo({
      url: '/pages/index/index',
    })
    return
 
  },
  copycode(e){
    wx.setClipboardData({
      data: e.currentTarget.dataset.code,       
    })
  },
  searchTap(){
    wx.navigateTo({
      url: '/pages/index/list?q=' + this.data.db.Name,
    })     
  },
  getTaobaoCode:function(){
    let that = this
    app.getUserInfo(function(info){
      wx.request({
        url: 'https://www.zaddone.com/goods/taobao',
        data: {
          goodsid: that.data.db.Ext ,
          ext: that.data.db.Name,
          session: info.OPENID,
        },
        success: function (res) {
          console.log(res)
          let code = res.data.tbk_tpwd_create_response.data.model
          wx.setClipboardData({
            data: code,
            success(res) {
              wx.showModal({
                showCancel:false,
                content: '成功复制' + code,
                title: '打开(淘宝/天猫)app 查看商品详情',
              })
            }
          })
          that.setData({ code: res.data.tbk_tpwd_create_response.data.model })
          
        }
      })

    });
    
    
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
  copyCode(){
    let that = this;
    wx.setClipboardData({
      data: that.data.code,      
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
      path: '/pages/index/shopping/taobao?goods='+this.data.db.Id
    }
  }
})