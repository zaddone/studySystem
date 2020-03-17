var app = getApp();
Page({

  /**
   * 页面的初始数据
   */
  data: {
    //tabs: [],
    activeTab: 0,
  },
  onTabCLick(e) {
    //console.log("click")
    const index = e.detail.index
    this.setData({ activeTab: index })
    this.showGoods(e.detail.index)
  },
  onChange(e) {
    //console.log("change")
    const index = e.detail.index
    this.setData({ activeTab: index })
    this.showGoods(e.detail.index)
  },
  details:function(e){
    //console.log(e.currentTarget.dataset.id)
    let tab = this.data.tabs[e.currentTarget.dataset.tab]
    //app.globalData.goodsdb = tab.db[e.currentTarget.dataset.goods]
    wx.navigateTo({
      url: '/pages/index/shopping/' + tab.py + '?goods=' + e.currentTarget.dataset.goods,
    })
  },
  showGoods:function(index){
    let tab = this.data.tabs[index]
    if (tab.db)return;
    console.log("tab show",index)
    wx.showLoading({
      title: '载入中',
    })
    let that = this;
    //let showHandle = app.globalData.shoppingMap.get(tab.py)
    this.getSearchGoods(tab.py, this.title, function (li){
      //console.log(li)
      if (!li || li.length===0) return
      let db = []
      li.forEach(function (v) {
        //console.log(v)
        v.Fprice = (v.Price * v.Fprice).toFixed(2)
        db.push(v)
      })
      that.data.tabs[index].db = db
      //goodsMap.set(py,db)
      that.setData({
        tabs: that.data.tabs,
        //db:db,
      },
        wx.hideLoading(),
      )
    })
  },
   
  getSearchGoods: function (py, key,hand) {
    let that = this
    wx.request({
      url: "https://www.zaddone.com/search/" + py,
      data: { keyword: key },
      success: function (res) {
        console.log(res)
        hand(res.data)         
      },
      fail: function () {
        wx.hideLoading();
        wx.showModal({
          title: '提示',
          content: '请求错误',
          confirmText:'重试',
          success(res) {
            if (res.confirm) {
              that.getSearchGoods(py,key,hand)
              //console.log('用户点击确定')
            } else if (res.cancel) {
              //console.log('用户点击取消')
            }
          }
        })
        //wx.navigateBack({})
      },
      complete:function(){
        wx.hideLoading();
      }
    })
  },
  /**
   * 生命周期函数--监听页面加载
   */
  onLoad: function (options) {
    wx.showLoading({
      title: '加载中',
    }) 
    //this.update = (this.title === options.q)
    this.title = options.q
    wx.setNavigationBarTitle({title:options.q})   
    //console.log(options.q)
    let that = this
    app.getShoppings(function (shoppings) {
      console.log(shoppings)
      let tabs = []
      shoppings.forEach(function(v,k){
        v.db=null
        tabs.push(v)
      })
      that.setData({ tabs: tabs })
      //hand(shoppings[0].py)      
      that.getSearchGoods(tabs[0].py, options.q, function (list) {
        let db = []
        list.forEach(function (v) {
          v.Fprice = (v.Price * v.Fprice).toFixed(2)
          db.push(v)
        })
        that.data.tabs[0].db = db
        that.setData({
          tabs: that.data.tabs,
        },
        wx.hideLoading(),
        )
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
    return {
      title: this.title,
      path: '/pages/index/list?q=' + this.title
    }
  }
})