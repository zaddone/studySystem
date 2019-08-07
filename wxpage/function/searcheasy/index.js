// 云函数入口文件
const cloud = require('wx-server-sdk')
cloud.init()

// 云函数入口函数
exports.main = async (event, context) => {
  //return
  const db = cloud.database()
  const config = await db.collection('config').doc('dbcoll').get()
  //console.log(config)
  if (!config.data) return
  const res = await db.collection(config.data.collword).where({
    _id: db.command.in(event.words)   
  }).get()
  if (res.data.length === 0) return;
  let Idmap = new Map()
  for (let k of res.data) {
    //console.log(k)
    let max = 1
    for (let w of event.words) {
      if (k._id.indexOf(w) != -1) {
        if (w.length > max) max = w.length
      }
    }
    //console.log(k)
    for (let l of k.link) {
      Idmap.set(l, ((Idmap.get(l) || 0) + (1 / k.link.length) * max))
      //console.log(l)  
      //Idmap.set(l, ((Idmap.get(l) || 0) + 1))      
    }
  }
  let maxlist = []
  let max = 0
  Idmap.forEach(function (value, key, map) {
    if (value>max){
      maxlist = [key]
      max = value
    }else if (value ===max){
      maxlist.push(key)
    }   
  })
  return maxlist
  
}
