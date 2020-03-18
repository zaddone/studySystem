// 云函数入口文件
const cloud = require('wx-server-sdk')

cloud.init()

// 云函数入口函数
exports.main = async (event, context) => {
  //const wxContext = cloud.getWXContext()
  const db = cloud.database()
  const config = await db.collection('config').doc('dbcoll').get()
  return {
    //event,
    userInfo: cloud.getWXContext(),
    config: config.data,
    //appid: wxContext.APPID,
    //unionid: wxContext.UNIONID,
  }
}