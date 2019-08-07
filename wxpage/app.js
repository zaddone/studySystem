var util = require('utils/md5.js')
const Towxml = require('towxml/main');
var QQMapWX = require('qqmap/qqmap-wx-jssdk.js');
var qqmapsdk = new QQMapWX({
  key: 'DFPBZ-F7YKX-QBQ4Z-TIIXH-VSLRE-JFFGJ'
});
var path = `${wx.env.USER_DATA_PATH}/tts.wav`
var patt1=/[a-zA-Z0-9]+|[\u3007\u3400-\u4DB5\u4E00-\u9FCB\uE815-\uE864]/g;
var patt2=/[^\u3007\u3400-\u4DB5\u4E00-\u9FCB\uE815-\uE864a-zA-z0-9]/g;
//var api_key ="wx0124084220537d9f"
//var secret_key = "2076410d21b7ce51244ea0b0e7cd8c08"
//var App_ID =2119096253
//var App_Key= omr9HcgFL9KaeM2i
//app.js
var Week = ['日', '一', '二', '三', '四', '五', '六'];
var DD = ['', '明', '后']
var NowDay = new Date().getDate()
App({
  onLaunch: function () {
    //console.log("run");
    if (!wx.cloud) {
      console.error('请使用 2.2.3 或以上的基础库以使用云能力')
    } else {
      wx.cloud.init({
        env: 'guomi-2i7wu',
        traceUser: true,
      })
    }
    this.fm = wx.getFileSystemManager()
    this.innerAudioContext = wx.createInnerAudioContext()

    this.innerAudioContext.autoplay = true
    this.innerAudioContext.onPlay(() => {
      console.log('开始播放')
    })
    this.innerAudioContext.onError((res) => {
      console.log(res.errMsg)
      console.log(res.errCode)
    })
    let that = this
    this.innerAudioContext.onEnded(function () {
      console.log("end", path)
     
      that.fm.stat({
        path: path,
        success: res => {
          console.log(res)
          that.fm.unlink({
            filePath: path,
            success: res => {
              console.log(res)
            },
            fail: res => {
              console.log(res)
            }
          })
        },
        fail: function (res) {
          console.log(res)
        }
      })
    })
    this.globalData = {
      searchMap: new Map(),
      sysinfo: wx.getSystemInfoSync(),
      dbColl: "",
      rz: 6,
      rw: 20,
      adId: "ad_id"
    }
    this.getDBColl()
    this.clearStorageDB()
    //this.getKey()
  },

  getDBColl(handle) {
    let that = this;
    if (this.globalData.dbColl != "") {
      handle()
      return
    }
    const db = wx.cloud.database();
    db.collection('config').doc('dbcoll').get({
      success: function (res) {
        that.globalData.dbColl = res.data.collname;
        that.globalData.adId = res.data.ad1;
        that.globalData.appid = res.data.qqapp_id;
        that.globalData.appkey = res.data.qqapp_key;
        handle()
      },
    })
  },
  theme_start: function () {
    //return "light_big";
    var h = (new Date()).getHours();
    if (h > this.globalData.rz && h < this.globalData.rw) {
      return "light_big";
    } else {
      return "dark_big";
    }
  },
  theme: function () {
    //return "light_big";
    var h = (new Date()).getHours();
    if (h > this.globalData.rz && h < this.globalData.rw) {
      wx.setBackgroundTextStyle({
        textStyle: 'light'
      })
      wx.setBackgroundColor({
        backgroundColor: '#f1f1f1',
        backgroundColorTop: '#f1f1f1',
        backgroundColorBottom: '#f1f1f1',
      });
      return "light_big";
    } else {
      wx.setBackgroundTextStyle({
        textStyle: 'dark'
      })
      wx.setBackgroundColor({
        backgroundColor: '#808080',
        backgroundColorTop: '#808080',
        backgroundColorBottom: '#808080',
      })
      return "dark_big";
    }
  },

  randomString(len) {
    len = len || 32;
    var $chars = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678';
    var maxPos = $chars.length;
    var pwd = '';
    for (let i = 0; i < len; i++) {
      pwd += $chars.charAt(Math.floor(Math.random() * maxPos));
    }
    return pwd;
  },
  getReqSign(req) {

    let sortli = []
    Object.keys(req).forEach(function (key) {
      //console.log(key, req[k"   b
      sortli.push(key)
    });

    //req.forEach(function (v, k, m) {
    //   sortli.push(k)
    // })
    sortli.sort()
    let str = ""
    for (let k of sortli) {
      str += k + "=" + encodeURIComponent(req[k]) + "&"

    }
    str += "app_key=" + this.globalData.appkey
    //console.log(str)
    return util.hexMD5(str).toUpperCase()

  },
  //https://api.ai.qq.com/fcgi-bin/nlp/nlp_textchat
  textChat(txt,hand){
    //this.session = this.randomString(32)
    let req = {
      app_id: this.globalData.appid,
      time_stamp: Date.parse(new Date()) / 1000,
      nonce_str: this.randomString(12),      
      session: "wxxiaochengxu",
      question: txt,
    }
    req.sign = this.getReqSign(req)
    //console.log(req)
    //req.set('sign', getReqSign(req))
    let that = this
    wx.request({
      url: 'https://api.ai.qq.com/fcgi-bin/nlp/nlp_textchat',
      data: req,
      success(r) {
        const data = r.data
        if (data.ret != 0) {
          console.log(data.msg)
          return
        }
        hand(data.data.answer)
      },
    })
  },
  soundToTxt(tmpfile, hand) {
    let that = this
    this.fm.readFile({
      filePath: tmpfile,
      encoding: 'base64',
      success: function (res) {
        //console.log(res.data)
        //const sp = wx.arrayBufferToBase64(res.data)
        //console.log(sp)
        let req = {
          app_id: that.globalData.appid,
          time_stamp: Date.parse(new Date()) / 1000,
          nonce_str: that.randomString(12),
          format: 4,
          speech: res.data,
        }
        req.sign = that.getReqSign(req)
        wx.request({
          url: 'https://api.ai.qq.com/fcgi-bin/aai/aai_asr',
          method: 'POST',
          header: { 'content-type': "application/x-www-form-urlencoded" },

          data: req,
          success(r) {

            const data = r.data
            if (r.data.ret != 0) {
              console.log(r.data.msg)
              return
            }
            hand(r.data.data.text)
          },
          fail: console.error,
        })
      }
    })


  },

  playtts(txt) {
    txt = txt.replace(/ |~|"|'/g, ',')
    console.log(txt)    
    let req = {
      app_id: this.globalData.appid,
      time_stamp: Date.parse(new Date()) / 1000,
      nonce_str: this.randomString(12),
      speaker: 6,
      format: 3,
      volume: 10,
      speed: 100,
      text: txt,
      aht: 0,
      apc: 58,
    }
    req.sign = this.getReqSign(req)
    //console.log(req)
    //req.set('sign', getReqSign(req))
    let that = this
    wx.request({
      url: 'https://api.ai.qq.com/fcgi-bin/aai/aai_tts',
      data: req,
      success(res) {
        if (res.data.ret != 0) {
          console.log(res.data)
          return
        }
        //console.log("w", path)
        that.fm.stat({
          path: path,
          success: res => {
            console.log(res)
            that.fm.unlink({
              filePath: path,
              success: res => {
                console.log(res)
              },
              fail: res => {
                console.log(res)
              }
            })
          },
          fail: function (res) {
            console.log(res)
          }
        })

        path = wx.env.USER_DATA_PATH + "/" + that.randomString(12) + ".wav"
        that.fm.writeFile({
          filePath: path,
          data: wx.base64ToArrayBuffer(res.data.data.speech),
          encoding: 'binary',
          success: function (src) {
            console.log(src)
            that.innerAudioContext.src = path
            that.innerAudioContext.play()
          },
          fail: console.error,
        })
      },
    })
  },
  showlink(d, hand) {
    console.log("showlink")
    console.log(d)
    for (let l of d.children) {
      this.getPageDBE(l, function (_db) {
        hand(_db)
      })
    }    
    if (!d.par) return
    let that = this
    that.getPageDBE(d.par, function (_db) {
      hand(_db)
      let li = new Set(d.par.children)
      li.delete(d._id)
      for (let l of li) {
        that.getPageDBE(l, function (__db) {
          hand(__db)
        })
      }
    })
  },
  toDBStyle(_d) {
    //console.log(text,_d.text)
    let timeOut = Date.parse(new Date()) / 1000 + 3600
    let text = decodeURIComponent(_d.text)
    if (text.startsWith("vod|")) {
      //text = text.replace(/\+/g, "|")
      let vod = []
      let listvod = text.replace(/\+/g, "|").split('|')
      listvod.shift()
      for (let l of listvod) {
        let l_ = l.split('$')
        vod.push({ title: l_[0], url: l_[1] })
      }
      return {
        vod: vod,
        title: _d.title,
        _id: _d._id,
        par: _d.par,
        children: _d.children,
        sign: false,
        timeOut: timeOut
      }
    }
    return {
      article: this.towxml.toJson("\n***\n# " + _d.title + "\n" + text.replace(/\+/g, " ") + "    \n***\n", 'markdown'),
      title: _d.title,
      _id: _d._id,
      par: _d.par,
      children: _d.children,
      sign: false,
      timeOut: timeOut
    }
  },
  inDB: function (data, hand) {
    let len = data.length;
    if (len == 0) return;
    for (let i = len - 1, _d; i >= 0; i--) {
      _d = data[i];
      if (!wx.getStorageSync(_d._id)) {
        wx.setStorageSync(_d._id, this.toDBStyle(_d));
        if (hand) hand(_d._id);
      }
    }
  },
  downDB: function (id, hand, fail) {
    let that = this
    that.getDBColl(function () {
      wx.cloud.database().collection(that.globalData.dbColl).doc(id).get({
        success(res) {
          //console.log(res)
          if (!res.data) {
            fail()
            return
          }
          hand(res.data)
          //hand(app.toDBStyle(res.data))
        },
        fail(err) {
          console.log(err)
          fail()
        }
      })
    })
  },
  clearStorageDB(){
    let o = (Date.parse(new Date()) / 1000) 
    wx.getStorageInfo({
      success(res) {
        for (let l of res.keys){
          wx.getStorage({
            key:l,
            success(r){
              if (r.data.timeOut && o > r.data.timeOut ){
                wx.removeStorage({key:l})
              }
          }
          })
        }
        
      }
    })
  },
  getPageDBE: function (key, hand) {
    let db = wx.getStorageSync(key);
    //console.log(key,db);
    let that = this
    if (db && (Date.parse(new Date()) / 1000) < db.timeOut) {
      hand(db);
      return
    }
    that.downDB(key, function (d) {
      let _d = that.toDBStyle(d)
      hand(_d);
      wx.setStorageSync(_d._id, _d);
    })
    //return false;

  },
  getPageDBExt: function (key, hand) {
    let db = wx.getStorageSync(key);
    //console.log(key,db);
    let that = this
    if (!db) {
      that.downDB(key, function (d) {
        let _d = that.toDBStyle(d)
        that.addDBList("list", key);
        hand(_d);
        wx.setStorageSync(_d._id, _d);
      })
      //return false;
    } else {
      hand(db);
    }
  },
  getPageDB: function (key, hand) {
    let db = wx.getStorageSync(key);
    //console.log(key,db);
    let that = this
    if (!db || (Date.parse(new Date()) / 1000) > db.timeOut) {
      that.downDB(key, function (d) {
        let _d = that.toDBStyle(d)
        hand(_d);
        that.addDBList("history", key);
        wx.setStorageSync(_d._id, _d);
      },
        function () {
          wx.reLaunch({
            url: "pages/search/search",
          })
        })
      //return false;
    } else {
      hand(db);
      that.addDBList("history", key);
    }
  },
  getShareDB: function (hand) {

    let si = wx.getStorageSync("sign");
    let that = this
    if (!si) {
      hand([])
      return
    }
    this.getDBColl(function () {
      const db = wx.cloud.database();
      db.collection(that.globalData.dbColl).where({
        _id: db.command.in(si)
      }).get({
        success: function (res) {
          let his = new Set(wx.getStorageSync("history") || []);
          //console.log(his)
          let link = new Set()
          for (let d of res.data) {
            for (let ld of d.link) {
              link.add(ld)
            }
          }

          let outlink = []
          for (let l of link) {
            if (!his.has(l)) outlink.push(l)
          }
          hand(outlink)
        }
      })
    })
    //let hisSet = new Set(wx.getStorageSync('history') || [])    

  },
  getStartdb: function (obj) {
    var that = this;
    this.getDBColl(function () {
      const db = wx.cloud.database();
      var dbq = db.collection(that.globalData.dbColl).orderBy('_id', 'desc');
      if (obj.t) {
        dbq = dbq.where({
          _id: db.command.nin(obj.t)
        })
      }
      dbq.get({
        success: function (res) {
          //console.log(res)
          obj.success(res.data);
        },
      })
    })
  },
  initdb: function (obj) {
    let that = this
    this.getDBColl(function () {
      var list = new Array();
      let his = wx.getStorageSync('history');
      const db = wx.cloud.database();
      var dbq = db.collection(that.globalData.dbColl)
      if (his) {
        dbq = dbq.where({ _id: db.command.nin(his) })
      }
      dbq.orderBy('_id', 'desc').get({
        success: function (res) {
          that.inDB(res.data, function (_id) { list.push(_id); });
          if (obj) obj(list);
          else wx.setStorageSync('list', list);
        },
      })
    })
  },

  updatedb: function (list, obj) {
    //let his = wx.getStorageSync('history');
    //if (his) this.checkClearHistorydb(his);
    var that = this;
    this.getDBColl(function () {
      wx.cloud.database().collection(that.globalData.dbColl).orderBy('_id', 'desc').get({
        success: function (res) {
          that.inDB(res.data, function (_id) { list.unshift(_id) });
          //wx.setStorageSync('list', list);        
          //console.log(res.data)
          if (obj) obj(list);
          else wx.setStorageSync('list', list);
        },
      });
    })
  },

  towxml: new Towxml(),

  addDBList: function (dbname, key) {
    wx.getStorage({
      key: dbname,
      success: function (res) {
        if (new Set(res.data).has(key)) return;
        res.data.push(key);
        if (res.data.length > 100) wx.removeStorageSync(res.data.shift())
        wx.setStorage({
          key: dbname,
          data: res.data
        });
      },
      fail: function () {
        wx.setStorage({
          key: dbname,
          data: [key]
        })
      },
    })
  },
  getIpCityName(handCity) {
    //var jskey = 'DFPBZ-F7YKX-QBQ4Z-TIIXH-VSLRE-JFFGJ'
    //var jssk = 'lQGrBLF7QPvSv1nZcddDSN5ILnY5SD9h'    
    let that = this
    wx.request({
      url: 'https://apis.map.qq.com/ws/location/v1/ip',
      data: {
        key: 'DFPBZ-F7YKX-QBQ4Z-TIIXH-VSLRE-JFFGJ',
        sig: util.hexMD5("/ws/location/v1/ip?key=DFPBZ-F7YKX-QBQ4Z-TIIXH-VSLRE-JFFGJlQGrBLF7QPvSv1nZcddDSN5ILnY5SD9h"),
      },
      success(res) {
        //console.log(res.data.result.ad_info)
        that.globalData.loc = {
          latitude: res.data.result.location.lat,
          longitude: res.data.result.location.lng
        }
        let db = res.data.result.ad_info
        let province = db.province.substring(0, db.province.length - 1)
        let city = db.city.substring(0, db.city.length - 1)
        handCity(province, city, "")
      },
      fail(res) {
        console.log(res)
        wx.reLaunch({ url: "/pages/search/search" })
      }
    })
  },
  objToArray(map) {
    let resstr = [];
    for (let w of map) {
      resstr.push(w)
    }
    resstr.sort(function (obj1, obj2) {
      var val1 = obj1.length;
      var val2 = obj2.length;
      if (val1 > val2) {
        return -1;
      } else if (val1 < val2) {
        return 1;
      } else {
        return 0;
      }
    })
    return resstr
  },
  getSearchKey(w) {
    let word = new Set(w.toLowerCase().split(patt2))
    //console.log(word)
    for (let x of word) {
      if (x.length <= 1) {
        word.delete(x)
        continue
      }
      let _x = x.match(patt1);
      for (let i = 0; i < _x.length; i++) {
        let str = _x[i]
        if (str.length>1)
        word.add(str)
        for (let _i = i + 1; _i < _x.length; _i++) {
          str += _x[_i]
          word.add(str)
        }
      }
    }
    for (let x of word) {
      console.log(x)
    }
    return this.objToArray(word)

  },
  getCityinfo(province, city, district, handCity) {
    let c = district
    let that = this
    if (c.length == 0) c = city
    wx.request({
      url: 'https://wis.qq.com/city/matching',
      data: {
        source: 'pc',
        city: c,
      },
      success(res) {
        //console.log(res.data)
        for (let v in res.data.data.internal) {
          let addr = res.data.data.internal[v].split(",")
          //console.log(addr)
          if (city === addr[1].trim()) {
            if (c === district) district = addr[2].trim()
            handCity(province, city, district)
            //that.getCityWeather(province, city, district)
            return
          }
        }
        if (district.length > 0) that.getCityinfo(province, city, district.substring(0, district.length - 1), handCity);
      },
      fail(res) {
        console.log(res)
        wx.reLaunch({ url: "/pages/search/search" })
      }
    })
  },
  getCityWeather(province, city, district, handweatcher) {
    let that = this
    wx.request({
      url: 'https://wis.qq.com/weather/common',
      data: {
        source: 'pc',
        weather_type: 'observe|forecast_1h|forecast_24h|alarm|tips|rise',
        province: province,
        city: city,
        county: district,
      },
      success(res) {
        let dataW = []
        //console.log(res.data.data)
        let alarm = []
        for (let i in res.data.data.alarm) {
          let ala = res.data.data.alarm[i]
          let adb = {
            title: ala.province + ala.city + ala.county + ala.type_name + ala.level_name + '预警 ',
            text: ala.detail,
          }

          alarm.push(adb)
        }
        let week = -1
        for (let i in res.data.data.forecast_24h) {
          let w = res.data.data.forecast_24h[i]
          let t = new Date(w.time)
          if (week === t.getDay()) {
            continue
          }
          week = t.getDay()

          let month = t.getMonth() + 1

          dataW.push({
            day_weather_code: w.day_weather_code,
            night_weather_code: w.night_weather_code,
            max_degree: w.max_degree,
            min_degree: w.min_degree,
            date: month + "-" + t.getDate(),
            week: '周' + Week[week],
            today: (week === new Date().getDay()) ? 'today' : '',
          })
        }
        let hd = []
        if (res.data.data.rise.length > 0) {
          that.globalData.rz = parseInt(res.data.data.rise[0].sunrise.substring(0, 2))
          that.globalData.rw = parseInt(res.data.data.rise[0].sunset.substring(0, 2))
        }

        //console.log(app.globalData.rz)
        for (let i in res.data.data.forecast_1h) {
          let w = res.data.data.forecast_1h[i]
          let h = parseInt(w.update_time.substring(8, 10))
          let c = (h > that.globalData.rz && h < that.globalData.rw) ? "d" : 'n'
          hd.push({
            weather: w.weather,
            weather_code: c + w.weather_code,
            degree: w.degree,
            wind: w.wind_direction,
            time_tag: DD[parseInt(w.update_time.substring(6, 8)) - NowDay],
            update_time: h,
          })

        }
        let tips = []
        if (res.data.data.tips.forecast_24h) {
          for (let i in res.data.data.tips.forecast_24h) {
            tips.push(res.data.data.tips.forecast_24h[i])
          }
        }
        if (res.data.data.tips.observe)
          for (let i in res.data.data.tips.observe) {
            tips.push(res.data.data.tips.observe[i])
          }
        let nowH = new Date().getHours()
        //console.log(nowH)
        let base = {
          degree: res.data.data.observe.degree,
          weather_code: ((nowH > that.globalData.rz && nowH < that.globalData.rw) ? 'd' : 'n') + res.data.data.observe.weather_code,
          weather: res.data.data.observe.weather,

        }

        that.globalData.weatcher = {
          alarm: alarm,
          cityname: city + ' ' + district,
          show: "visible",
          w: base,
          dayW: dataW,
          hd: hd,
          tips: tips[Math.floor(Math.random() * tips.length)],
        }
        handweatcher(that.globalData.weatcher)

        //app.globalData.weatcher = res.data.data.observe.weather + res.data.data.observe.degree + "℃"
        //that.getPageList()
        //that.getCityinfo(province, city, district)
      },
      fail(res) {
        console.log(res)
        wx.reLaunch({ url: "/pages/search/search" })
      }
    })
  },

  getAddrWeather(location, hand) {
    //let that = this
    qqmapsdk.reverseGeocoder({
      sig: "lQGrBLF7QPvSv1nZcddDSN5ILnY5SD9h",
      location: location,
      success: function (res) {
        let db = res.result.address_component
        let province = db.province.substring(0, db.province.length - 1)
        let city = db.city.substring(0, db.city.length - 1)
        let district = db.district
        if (district.length > 2) {
          district = district.substring(0, db.city.length - 1)
        }
        hand(province, city, district)
      },
      fail(res) {
        console.log(res)
        wx.reLaunch({ url: "/pages/search/search" })
      }

    })
  },
  onPageNotFound() {
    wx.redirectTo({
      url: '/pages/search/search',
    })
  },
})