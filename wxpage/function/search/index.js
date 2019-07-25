// 云函数入口文件
const cloud = require('wx-server-sdk')
cloud.init()

// 云函数入口函数
exports.main = async (event, context) => {
  const db = cloud.database()

  const res = await db.collection('word').where({
    _id: db.RegExp({
      regexp: event.words.join('|'),
    })
  }).get()
  let Idmap = new Map()
  for (let k of res.data){
    //console.log(k)
    for (let l of k.link){      
      Idmap.set(l,((Idmap.get(l)||0)+1/k.link.length)*k._id.length)      
    }
  }
  let list = []
  Idmap.forEach(function (value, key, map){
    list.push({ _id: key, val: value})    
  })
  list.sort(function (obj1, obj2) {
    var val1 = obj1.val;
    var val2 = obj2.val;
    if (val1 > val2) {
      return -1;
    } else if (val1 < val2) {
      return 1;
    } else {
      return 0;
    }
  } )
  let i = 0
  let pageid = []
  for (let d of list){
    pageid.push(d._id)
    i++;
    if (i>5)break
  }
  let plist =await db.collection('pageColl').where({
    _id: db.command.in(pageid)
  }).get()

  const vod = await db.collection('vod').where({
    _id: db.RegExp({
      regexp: event.words.join('|'),
    })
  }).limit(5).get()
  let row = []
  for (let v of plist.data){
    row.push(v)
  }
  for (let v of vod.data){
    row.push(v)
  }
  return row
}