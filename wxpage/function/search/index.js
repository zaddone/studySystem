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
    //_id: db.command.in(event.words)
    _id: db.RegExp({
      regexp: event.words.join('|'),
    })
  }).get()
  if (res.data.length===0)return;
  let Idmap = new Map()
  for (let k of res.data){
    //console.log(k)
    let max =1
    for (let w of event.words){
      if (k._id.indexOf(w)!= -1){
        if (w.length>max)max=w.length
      }
    }
    //console.log(k)
    for (let l of k.link){      
      Idmap.set(l, ((Idmap.get(l) || 0) + (1 / k.link.length )* max) )    
      //console.log(l)  
      //Idmap.set(l, ((Idmap.get(l) || 0) + 1))      
    }
  }
  let list = []
  Idmap.forEach(function (value, key, map) {
    list.push({ _id: key, val: value })
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
  })
  //console.log(list)
  let maxli = list[0];
  let i = 0 
  let listset = new Set()
  //listset.add(maxli._id)    
  for (let d of list) {
    listset.add(d._id) 
    i++
    if (i>10)break
  }  
  let page = []
  if (i<10){
    let first = list[0]._id
    let res1 = await db.collection(config.data.collname).doc(maxli._id).get()
    for(;i<10;i++){
      if (!res1.data.link)break
      listset.add(res1.data.link.pop())
    }    
    listset.delete(first)
    page.push(res1.data)
  }
  
  let res2 = await db.collection(config.data.collname).where({
    _id: db.command.in(Array.from(listset))
  }).get()
  for (let k of res2.data){
    page.push(k)
  }
  return page
}