// miniprogram/pages/index/details.js
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
    if (options.goods && options.py) {
      let that = this
      app.getGoodsData(options.py, options.goods, function (db) {
        that.setData({
          //tab: { py: options.py },
          py:options.py,
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
  pinduoduoRouter: function () {
    let that = this
    app.getUserInfo(function (info) {
    wx.request({
      url: 'https://www.zaddone.com/goods/' + that.data.py,
      data: { goodsid: that.data.db.Id, session: info.OPENID },
      success: function (res) {
        let r = res.data.goods_promotion_url_generate_response.goods_promotion_url_list[0].we_app_info
        wx.navigateToMiniProgram({
          appId: r.app_id,
          path: r.page_path,
        })
        //console.log()
      },
    })  
    })  
  },
  tapDialogButton(e) {
    this.data.db.show = false
    this.setData({
      db:this.data.db
      //dialogShow: false,
      //showOneButtonDialog: false
    })
  },
  
  taobaoRouter:function(){
    let that = this
    app.getUserInfo(function (info) {
      wx.request({
        url: 'https://www.zaddone.com/goods/taobao',
        data: {
          goodsid: that.data.db.Ext,
          ext: that.data.db.Name,
          session: info.OPENID,
        },
        success: function (res) {          
          if (res.statusCode!=200 || !res.data || res.data.length===0){
            wx.showToast({
              title: '网络错误',
              icon:'none'
            })
            return;
          }
          console.log(res)
          //let code = 
          that.data.db.code = res.data.tbk_tpwd_create_response.data.model
          that.data.db.show = true
          that.setData({            
            db: that.data.db
          })
                    
          //that.setData({ code: res.data.tbk_tpwd_create_response.data.model })

        }
      })

    });
  },
  jdRouter:function(){
    let that = this
    app.getUserInfo(function (info) {
      wx.request({
        url: 'https://www.zaddone.com/goods/' + that.data.py,
        data: { goodsid: that.data.db.Id, session: info.OPENID },
        success: function (res) {
          console.log(res)
          let uri = res.data.jd_kpl_open_promotion_pidurlconvert_response.clickUrl.shortUrl;
          //pages/proxy/union/union?isUnion=1&spreadUrl=
          console.log(uri)
          wx.navigateToMiniProgram({
            appId: 'wx1edf489cb248852c',
            path: "pages/proxy/union/union?spreadUrl=" + uri,
          })
        },
      })
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
    if (!this.data.db)return;
    return {
      title: this.data.db.Name,
      path: '/pages/index/details?py='+this.data.py+'&goods=' + this.data.db.Id
    }
  } 
})