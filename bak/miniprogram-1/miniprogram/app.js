//app.js
var util = require('utils/sha1.js')
App({
  onLaunch: function () {    
    if (!wx.cloud) {
      console.error('请使用 2.2.3 或以上的基础库以使用云能力')
    } else {
      wx.cloud.init({
        // env 参数说明：
        //   env 参数决定接下来小程序发起的云开发调用（wx.cloud.xxx）会默认请求到哪个云环境的资源
        //   此处请填入环境 ID, 环境 ID 可打开云控制台查看
        //   如不填则使用默认环境（第一个创建的环境）
        // env: 'my-env-id',
        //env: 'guomi-2i7wu',
        env:'zaddone-5gsor',
        traceUser: true,
      })
    }

    this.globalData = {}
    this.getUserInfo();
    this.getSystemInfo();
    //this.initShoppingMap();
    //this.clipboardData()
    //this.addSign();
  },
  getShoppings: function (hand) {
    //console.log(app.globalData.sha1("test"))
    if (this.globalData.shoppings) {
      hand(this.globalData.shoppings)
      return;
    }
    // else {
   //   let sh = wx.getStorageSync("shoppings");
   //   if (sh) {
   //     this.globalData.shoppings = sh.db 
   //     //this.setData({ tabs: sh.db })
   //     if (sh.date === (new Date()).getDate()) {
   //       hand(this.globalData.shoppings)
   //       return
   //     }
   //   }
   // }
    let that = this
    wx.request({
      url: "https://www.zaddone.com",
      data: { content_type: "json" },
      success: function (res) {
        //console.log(res)
        let shopping = []
        let other
        res.data.forEach(function (v, i) {
          if (v.py==='taobao'){
            v.title = '其他'
            other = v
            return
          }
          v.title = v.Name  
          shopping.push(v)                  
        }) 
        shopping.push(other)
        that.globalData.shoppings = shopping;     
        hand(that.globalData.shoppings)
        //wx.setStorage({
        //  key: 'shoppings',
        //  data: { db: res.data, date: (new Date()).getDate() },
        //})
      },
      fail: function () {
        wx.navigateBack({})
      },
    })
  },
  getUserInfo:function(hand){
    let that = this
    if (that.globalData.userInfo){
      if (hand)hand(that.globalData.userInfo)
      return
    }
    wx.cloud.callFunction({
      name: 'userinfo',
      complete: function(res){
        console.log(res.result)
        that.globalData.userInfo = res.result.userInfo
        that.globalData.config = res.result.config
        if (hand)hand(that.globalData.userInfo)
      },
    })
  },
  getSystemInfo:function(){
    let that = this
    wx.getSystemInfo({ success:function(res){
      //console.log(res.global.window)
      that.globalData.systemInfo = res
    }})
  },
  handKeyWords(data,handhttp,handorderid,handother){
    let val = /(http[\S]+)[\+| ]?/.exec(data)
    //console.log(val) 
    if (val) {
      //let db = {url:val[1]}
      var url = val[1];
      var py, id
      if (/yangkeduo/.test(url)) {
        py = 'pinduoduo';
        id = /[\?|\&]goods_id=(\d+)/.exec(url)[1];
      } else if (/jd.com/.test(url)) {
        py = 'jd';
        id = /\/(\d+)\.html/.exec(url)[1];
        console.log("jd")
      } else if (/tb.cn/.test(url)) {
        py = 'taobao';
        id = url;
      } else if (/taobao.com/.test(url)) {
        py = 'taobao';
        id = url;
      } else return;

      handhttp(py,id)
      return;
    }
    let num = /\d{6}-\d{15}/.exec(data)
    if (num) {
      handorderid(num[0])
      //that.handOrder(num[0]);
      return;
    }
    num = /\d+/.exec(data)

    if (num) {
      let n = num[0]
      let len = n.length;
      console.log(n, len)
      if (len == 12 || len == 18) {
        handorderid(n);
        return
      }
    }
    handother(data)
  },
  clipboardData:function(){
    let that = this    
    wx.getClipboardData({
      success(res) {
        if (!res.data){
          //that.globalData.clipboard=''
          return;
        }
        that.handKeyWords(res.data, 
        function(py,id){
          wx.setClipboardData({
            data: "",
            success: function () {
              //wx.hideLoading();
              wx.hideToast()
              that.showModeToPage(py, id)
            },
          })
        },
        function(orderid){
          wx.setClipboardData({
            data: "",
            success: function () {
              //wx.hideLoading();
              wx.hideToast()
              that.handOrder(orderid)
            },
          })
        },
        function(key){

        })
        return        
      }
    })
  },
  addsign(reg){
    reg.timestamp =Date.parse(new Date()) / 1000;
    let sortli = [];
    for (var i in reg) {
      //console.log(v)
      sortli.push(reg[i])
    }
    sortli.push(this.globalData.config.zaddone)
    sortli.sort()
    console.log(sortli)
    reg.sign = util(sortli.join(''))
  },
  handOrder(orderid){
    
    let reg={
      orderid:orderid,      
    }    
    let that = this
    this.getUserInfo(function (info) {
      reg.userid = info.OPENID;
      that.addsign(reg)
      wx.request({
        url: 'https://www.zaddone.com/v1/order_apply',
        data:reg,
        success:function(req){
          console.log(req.data)
          if (req.statusCode != 200){
            wx.showToast({
              title: '没有找到订单',
              icon:'none'
            })
            return;
          } 
          
          wx.showModal({
            title:'添加订单' ,
            content: req.data.goodsName+req.data.text,
            showCancel:false,
            confirmText:"关闭",
            success(res) {               
            }
          })
        },
      })
      //hand(reg)
    })
  },   
  getGoodsData(py, id, hand, fail){
    let that = this
    if (that.globalData.goodsdb && that.globalData.goodsdb.Id == id){
      console.log("find")
      hand(that.globalData.goodsdb)
      return
    }
    wx.showLoading({
      title: '请稍等',
    })
    wx.request({
      url: 'https://www.zaddone.com/goodsid/'+py,
      data: { goodsid: id },
      success: function (res) {
        if (res.statusCode != 200 || !res.data || res.data.length===0) {
          wx.showToast({
            title: '没有找到',
            icon: 'none'
          })
          return;
        }          
        let db = res.data[0]
        if (db.Fprice ===0){
          db.Fprice = '没有返利'
        }else{
          db.Fprice = (db.Price * db.Fprice).toFixed(2)
        }
        
        that.globalData.goodsdb = db;  
        hand(db)       
      },  
      fail:fail,
      complete:function(){
        wx.hideLoading();
      }    
    })
  },
  showModeToPage(py,id){
    this.getGoodsData(py,id,function(db){
      wx.showModal({
        title: '点击确定查看',
        content: db.Name,
        success(res) {
          if (res.confirm) {
            wx.navigateTo({
              url: '/pages/index/details?py='+py+'&goods='+id,
            })
            //console.log('用户点击确定')
          } else if (res.cancel) {
            console.log('用户点击取消')
          }
        }
      })      
    },function(){
      wx.showToast({
        title: '没有找到',
        duration: 5000,
      })    
    })    
  },
  onHide(){
  },
  onShow(){
    //console.log("show")
    //this.clipboardData()
    let that = this
    let k = setInterval(
      function () {
        that.clipboardData()
        clearInterval(k)
      }, 1000)
  },
})
