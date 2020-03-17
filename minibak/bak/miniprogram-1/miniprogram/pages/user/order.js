var app = getApp();
function formatDateTime(inputTime) {
  var date = new Date(inputTime);
  var y = date.getFullYear();
  var m = date.getMonth() + 1;
  m = m < 10 ? ('0' + m) : m;
  var d = date.getDate();
  d = d < 10 ? ('0' + d) : d;
  return y + '-' + m + '-' + d;
};
Page({

  /**
   * 页面的初始数据
   */
  data: {
    lastorderid:"",
  },
  showuser(e) {
    wx.navigateTo({
      url: '/pages/user/form',
    })
  },
  /**
   * 生命周期函数--监听页面加载
   */
  onLoad: function (options) {
    //app.get
    wx.showLoading({ title:'加载中'})
    this.getOrderlist();
    
  },
  showpage(e){
    console.log(e)
    wx.navigateTo({
      url: '/pages/index/shopping/' + e.currentTarget.dataset.site + '?goods=' + e.currentTarget.dataset.id,
    })
  },
  getOrderlist(){
    let that = this;
    app.getUserInfo(function (info) {
      let reg = { userid: info.OPENID, orderid: that.data.lastorderid };
      app.addsign(reg)
      wx.request({
        url: 'https://www.zaddone.com/v1/user/order',
        data: reg,
        success:function(res){
          console.log(res.data);  
          var li = []          
          var sum=0 ;
          if (!res.data)return;
          res.data.forEach(function(v,k){
            if (v.payTime){
              v.date = formatDateTime(new Date(v.payTime*1000))
            }
            li.push(v)
            sum += v.fee 
          })    
          li.sort(function(a,b){
            return a.time - b.time
          })
           
          that.setData({
            show:true,
            sum:sum,            
            db:li,          
          })
        },
        complete:function(){
          wx.hideLoading();
        }
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

  }
})