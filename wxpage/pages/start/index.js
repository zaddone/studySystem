var app = getApp();
Page({
  data: {
    statusBarHeight: app.globalData.sysinfo.statusBarHeight,
    windowsHeight: app.globalData.sysinfo.screenHeight,
    windowsWidth: app.globalData.sysinfo.screenWidth,
    show: "hidden",
    //listshow: "hidden",
    theme: app.theme_start(),
    top: app.globalData.sysinfo.statusBarHeight + 10,
    adId : "null"
  },
  showModle(event){
    //console.log(event.currentTarget.dataset)
    wx.showModal({
      title: event.currentTarget.dataset.item.title,
      content: event.currentTarget.dataset.item.text,
      showCancel:false,
      confirmText: '关闭',
    })
  },
  toSearch(event){
    //console.log(event)
    wx.navigateTo({ url:"../search/search"})
  },
  upWeaterInfo(province, city, district){
    let that = this
    app.getCityWeather(province, city, district, function (data) {
      that.setData(data,function(){
        //console.log("event")
        //setTimeout(that.setPageListAddr,1000)
      } )
      wx.vibrateShort();
      //console.log(data)
      for (let adb of data.alarm)
        wx.showModal({
          title: adb.title,
          content: adb.text,
          showCancel: false,
          confirmText: '关闭',
        })
    })
  },
  toFace(){
    wx.navigateTo({ url: "../face/face" })
  },
  onLoad(obj) {

    let that = this   
    app.getDBColl(function(){
      that.setData({
        adId: app.globalData.adId,
      })
    })
    app.getToday(function(t){
      that.setData({
        today:t
      })
    })
    //that.setIndexList()
    if (obj && obj.loc) {      
      this.setData(app.globalData.weatcher)
      wx.vibrateShort();
      return
    }
    //console.log(wx.getSystemInfoSync())
   if (!wx.getStorageSync("openMap")){
     app.getIpCityName(that.upWeaterInfo)
    }else{ 
    wx.getLocation({
      type: 'gcj02',
      fail(res){
        app.getIpCityName(that.upWeaterInfo)
        //console.log(res)
      },
      success(res) {
        app.globalData.loc = {
          latitude: res.latitude,
          longitude: res.longitude
        }
        app.getAddrWeather(that.data.loc,
          function (province, city, district) {
            app.getCityinfo(province, city, district, that.upWeaterInfo)
               
          })
      },
    })
    }
  },
    
  setPageListAddr(){
    //return
    let that = this
     wx.createSelectorQuery().select('.scroll-view_H').boundingClientRect(function (rect) {
      //let height = rect.height
       let top = (rect.height || 0)
    let height = that.data.windowsHeight - top
      that.setData({
        //scrollviewbotton: height-50,
        pagelistheight: height,
        scrollviewtop: top,

      },function(){
        return
        wx.createSelectorQuery().select('#up').boundingClientRect(function (rect) {
          //console.log(rect)
          let h = rect ? rect.height || 0 : 0
          //let height = (that.data.windowsHeight - (rect.height || 0))
          that.setData({
            scrollviewbotton: height - h,
            //pagelistheight: height,
          })
        }).exec();
      })


    }).exec();
  },
   
  onPullDownRefresh() {
    wx.stopPullDownRefresh();
    this.onLoad()
  },
  openMap: function() {
    wx.vibrateShort();
    wx.reLaunch({
      url: '../map/map?latitude=' + app.globalData.loc.latitude + '&longitude=' + app.globalData.loc.longitude
    })
  },
  onShareAppMessage: function (res) {
    return {
      title: "今日天气",
      path: 'pages/start/index'
    }
  },
})