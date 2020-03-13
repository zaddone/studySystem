function jdPageHtml(data){

  if (data.length === 0){
        $('.list').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">没有找到相关信息<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
  	return
  }
 let that = this
  that.db = []
 $.each(data, function(key, val) {	
  val.fprice = (parseFloat(val.commisionRatioWl)/100 * parseFloat(val.price)).toFixed(2);
  val.imgurl = 'https://img14.360buyimg.com/ads/'+val.imageUrl
  //if (val.coupon_discount)val.c = '<span class="badge badge-warning">券</span>';
  //console.log(val)
  that.db.push(val)
  that.html(key,val)
 })
}
function htmljd(key,val){
  $('.list').append('<div class="col-lg-2 top" ><a class="goods card btn btn-link" href="javascript:ShowJdBox('+key+')"><span class="position-absolute"><span class="badge badge-secondary">京东</span></span><img src="'+val.imgurl+'" class="card-img-top" alt="'+val.wareName+'"><div class="overflow-hidden" style="height:100px"><div class="card-text"><span class="badge badge-danger">￥'+val.price+'-'+val.fprice+'</span><p class="name">'+val.wareName+'</p></div></div></a></div>')
}
function ShowJdBox(key){
  let obj = ShoppingMap.get('jd')
  let val = obj.db[key]
  $('#myLargeModalLabel').modal('toggle')
  $('.modal-title').html(val.wareName)
  $('.carousel-inner').html('<div class="carousel-item active"><img src="'+val.imgurl+'" class="d-block w-100" ></div>')
  $('.text').html('<span class="badge badge-danger">￥'+val.price+'返'+val.fprice+'</span>')
  $('.pmsg').html('<p><a  class="btn btn-info"   target="_blank"  href="https://www.zaddone.com/p/jd/'+val.skuId+'">点击这里 到"京东"下单</a></p>')
  $('.pmsg').append('<p>关注微信公众号 米果推荐 zaddone_com,输入订单号获取返利详情</p>')
  $('.pmsg').show(); 
  
}
