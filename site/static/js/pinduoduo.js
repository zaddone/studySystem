function pinduoduoPageHtml(db_){
  let data
  try{
  data = db_.goods_search_response.goods_list
  }catch(err) {
  data = db_.goods_detail_response.goods_details
  }
  let that = this
 $.each(data, function(key, val) {	
  val.min_group_price /= 100.0
  val.min_normal_price /= 100.0
  val.fprice = (val.min_group_price*(val.promotion_rate/1000.0)).toFixed(2);
  val.c="";
  if (val.coupon_discount)val.c = '<span class="badge badge-warning">券</span>';
  that.db.push(val)
  that.html(key,val)
 })
}
function htmlpinduoduo(key,val){
  $('.list').append('<div class="col-lg-2 top" ><a class="card btn btn-link" href="javascript:ShowPddBox('+key+')" ><span class="position-absolute"><span class="badge badge-secondary">拼多多</span><span class="badge badge-dark">'+val.mall_name+'</span></span><img src="'+val.goods_thumbnail_url+'" class="card-img-top" alt="'+val.goods_desc+'"><div class="overflow-hidden" style="height:100px"><p class="card-text"><span class="badge badge-danger">￥'+val.min_group_price+'-'+val.fprice+'</span>'+val.c+'</br>'+val.goods_name+'</p></div></a></div>')
}
function ShowPddBox(key){
  let obj = ShoppingMap.get('pinduoduo')
  let val = obj.db[key]
  $('#myLargeModalLabel').modal('toggle')
  $('.modal-title').html(val.goods_name)
  $('.carousel-inner').html('<div class="carousel-item active"><img src="'+val.goods_image_url+'" class="d-block w-100" ></div>')
  $('.text').html('<span class="badge badge-danger">￥'+val.min_group_price+'返'+val.fprice+'</span>'+val.c)
  $('.text').append('<p>'+val.mall_name+'</p>')
  //$('.text').append('<p>'+val.goods_desc+'</p>')
  $('.pmsg').html('<p><a  target="_blank"  href="https://www.zaddone.com/p/pinduoduo/'+val.goods_id+'">点击这里 到"拼多多"下单</a></p>')
  $('.pmsg').append('<p>关注微信公众号 米果推荐 zaddone_com,</p><p>输入订单号获取返利详情</p>')
  $('.pmsg').show(); 
}
