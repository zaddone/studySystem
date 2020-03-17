// miniprogram/pages/user/form.js
var app = getApp();
Page({

  /**
   * 页面的初始数据
   */
  data: {
    //mobile:12345678901,
    //name:"asdfa",
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad: function (options) {
    this.getUser()
  },
  getUser(){
    let that = this
    app.getUserInfo(function (info) {
      let reg = {userid : info.OPENID};
      app.addsign(reg)
      wx.request({
        url: 'https://www.zaddone.com/v1/user/get',
        data: reg,
        success: function (req) {
          that.setData({ mobile: req.data.msg.Mobile, name: req.data.msg.Name})
          //console.log(req.data.msg.Mobile)
        },
      })
      //hand(reg)
    })
    
  },
  formSubmit:function(e){
    console.log(e)
    this.data.name = e.detail.value.name
    this.data.mobile = e.detail.value.mobile
    let reg = { name: e.detail.value.name, mobile: e.detail.value.mobile}

    app.getUserInfo(function (info) {
      reg.userid = info.OPENID;
      app.addsign(reg)
      wx.request({
        url: 'https://www.zaddone.com/v1/user/update',
        data: reg,
        success: function (req) {
          console.log(req)
          wx.showModal({
            title: '提示',
            content: '提现信息设置成功。到期返利将主动转入微信钱包，本平台收取10%技术服务费，欢迎使用',
            showCancel:false,
            success(res) {
              if (res.confirm) {
                console.log('用户点击确定')
              } else if (res.cancel) {
                console.log('用户点击取消')
              }
            }
          })
        },
      })
      //hand(reg)
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

  },
  getPhoneNumber(e) {
    console.log(e.detail)

  },
})