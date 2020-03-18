var app = getApp();
Page({
  data: {
    loading: false,
    color: '#000',
    background: '#f8f8f8',
    show: true,
    title:"查询",
    animated: false,
    keywords:"关键词 链接 订单号",

  },
  onLoad: function (options) {
    let sitelist = []
    let that = this
    app.getShoppings(function(db){
      db.forEach(function(v){
        sitelist.push(v.Name)
      })
      that.setData({site:sitelist})
      console.log(db)
    })
  },   
  toClear:function(e){
    e.detail.value = ''
    this.keywords=""
  },
  toSearch: function (e) {
    if (!e.detail.value || e.detail.value.length === 0) return;
    app.handKeyWords(e.detail.value, app.showModeToPage, app.handOrder, this.search)
    //this.search(e.detail.value)
  },
  showuser(e){
    wx.navigateTo({
      url: '/pages/user/form',
    })
  },
  showorder(e){
    wx.navigateTo({
      url: '/pages/user/order',
    })
  },
  showhelp(e) {
    wx.navigateTo({
      url: '/pages/user/help',
    })
  },
  search(q){
    wx.navigateTo({
      url: '/pages/index/list?q=' + q,
    })
  },
  onShow() {
    //console.log("show")
    //app.clipboardData();
  },
  onShareAppMessage: function () {
    return {
      title: "网购返利",
      path: '/pages/index/index'
    }
  }
});