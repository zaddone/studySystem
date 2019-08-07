var app = getApp();
var logo = "../../images/shejiyeiconoiujmbtptap.png";
Page({
  /**
   * 页面的初始数据
   */
  data: {
    top: app.globalData.sysinfo.statusBarHeight + 10,
    weatcher: app.globalData.weatcher,
    theme: app.theme_start(),
  },
  gohome: function () {
    wx.reLaunch({
      url: '../start/index'
    })
  },
  regionchange(event){
    this.getCenterLocation()
    //console.log(event)
  },
  handmap:function(event){
    //console.log(event)
  },
  markertap: function (event) {
    //app.globalData.loc.latitude = this.data.markers[event.markerId].latitude;
    //app.globalData.loc.longitude = this.data.markers[event.markerId].longitude;
    wx.reLaunch({
      url: '../start/index?loc=' + this.data.markers[event.markerId].latitude + ',' + this.data.markers[event.markerId].longitude
    })
    //console.log(event)
  },
  /**
   * 生命周期函数--监听页面加载
   */
  onReady: function (e) {
    // 使用 wx.createMapContext 获取 map 上下文
  
  },
  setMapInfo(latitude, longitude,call){
    app.globalData.loc = {
      longitude: longitude,
      latitude: latitude,
    };
    this.setData({
      latitude: latitude,
      longitude: longitude,
      windowsHeight: app.globalData.sysinfo.windowHeight,
      markers: [{
        iconPath: logo,
        id: 0,
        latitude: latitude,
        longitude: longitude,
        width: 50,
        height: 50,
        //callout: { content: this.data.weatcher, display: 'ALWAYS',},
        callout:call,
      }],

    })
  },
  onLoad: function (options) {
    wx.setStorageSync("openMap",true)
    let that = this
    this.mapCtx = wx.createMapContext('myMap')
    wx.getLocation({
      type: 'gcj02',
      fail(res) {
        //console.log("fail",res)
        if (!options.longitude) {
          return
        }
        that.setMapInfo(options.latitude, options.longitude)
      },
      success(res) {
        //console.log("success", res)
        that.setMapInfo(res.latitude, res.longitude)
      },
    })

   
  },
  showData(res){
    this.setData({
      markers: [{
        //  iconPath: "../../images/shejiyeiconoiujmbtptap.png",
        iconPath: logo,
        id: 0,
        latitude: res.latitude,
        longitude: res.longitude,
        width: 50,
        height: 50,
        callout: {
          content: app.globalData.weatcher.w.weather + '  ' + app.globalData.weatcher.w.degree + '℃\n' + app.globalData.weatcher.cityname + "\n",
          display: 'ALWAYS',
          bgColor: '#f1f1f1',
          borderWidth: 1,
          padding: 5,
          borderRadius: 10,
        },
      }],
    })
  },
  getCenterLocation: function () {
    let that = this;
    this.mapCtx.getCenterLocation({
      success: function (res) {
        app.globalData.loc = {
          longitude: res.longitude,
          latitude: res.latitude,
        };
        //app.globalData.loc.latitude = res.latitude;
        app.getAddrWeather(app.globalData.loc,
          function (province, city, district) {
            //console.log(province, city, district)
            if (city + " " + district === app.globalData.weatcher.cityname){
              that.showData(res)
              return
            }


            app.getCityinfo(province, city, district,
              function (province, city, district) {
                app.getCityWeather(province, city, district, function (data) {
                  //console.log(data)
                 that.showData(res)
                })
              })
          })

      

      }
    })
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