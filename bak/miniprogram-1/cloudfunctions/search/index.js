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
    //_id: db.RegExp({
     // regexp: event.words.join('|'),
    //})
  }).get()
  let Idmap = new Map()
  if (res.data.length===0){
    //return
    res = await db.collection(config.data.collword).where({
      //_id: db.command.in(event.words)
      _id: db.RegExp({
       regexp: event.words.join('|'),
      })
    }).get()
    if (res.data.length === 0) return;
    for (let k of res.data) {
      for (let l of k.link) {
        //Idmap.set(l, ((Idmap.get(l) || 0) + (1 / k.link.length) * k._id.length))
        Idmap.set(l, ((Idmap.get(l) || 0) + (1 / k.link.length)))    
      }
    }

  }else{
    for (let k of res.data) {
      for (let l of k.link) {
        Idmap.set(l, ((Idmap.get(l) || 0) + (1 / k.link.length) * k._id.length))
        //Idmap.set(l, ((Idmap.get(l) || 0) + (1 / k.link.length)))    
      }
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
  let i = 0 
  let listarr = []
  let listset = new Set()
  //listset.add(maxli._id)    
  for (let d of list) {
    listarr.push(d._id)
    //listset.add(d._id) 
    i++
    if (i>10)break
  }  
  let page = []
  let nlist = []
  let res2 = await db.collection(config.data.collname).where({
    _id: db.command.in(listarr)
  }).get()
  let listMap = new Map()
  for (let k of res2.data){
    listMap[k._id] = k
    listset.add(k._id)    
    //page.push(k)
  }
  for (let l of listarr){
    let li = listMap[l] 
    if (!li)continue
    page.push(li)
    let lp = li.par
    if (lp){
      if (!listset.has(li.par)) nlist.push(li.par)
    }    
  }
  if (page.length<10 && nlist.length>0){
    let res3 = await db.collection(config.data.collname).where({
      _id: db.command.in(nlist)
    }).get()
    for (let k of res3.data) {
      page.push(k)
    }
  }
   return page
}