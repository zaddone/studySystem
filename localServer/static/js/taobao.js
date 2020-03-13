function taobaoPageHtml(data){
  //console.log(db_)
  if (data.length ===0){
        $('.list').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">没有找到相关信息<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
  	return
  }
  
  let that = this
  that.db = []
 $.each(data, function(key, val) {	
  val.tname = "淘宝"
  if (val.user_type === 1) val.tname = "天猫";
  val.url_ ='https:'+val.url;
  if (val.coupon_share_url)val.url_ ='https:'+val.coupon_share_url;
  val.c="";
  if (val.coupon_id) val.c = '<span class="badge badge-warning">券</span>'
  val.price = val.zk_final_price
  val.fprice = (val.price*(val.commission_rate/10000.0)).toFixed(2);
  that.db.push(val)
  that.html(key,val)
 })
}

function htmltaobao(key,val){
  $('.list').append('<div class="col-lg-2 top" ><a class="goods card  btn btn-link" href="javascript:ShowTaobaoBox('+key+')"><span class="position-absolute"><span class="badge badge-warning">'+val.tname+'</span><span class="badge badge-dark">'+val.provcity+' '+val.shop_title+'</span></span><img src="'+val.pict_url+'" class="card-img-top" alt="'+val.title+'"><div class="overflow-hidden" style="height:100px"><div class="card-text"><span class="badge badge-danger">￥'+val.price+'-'+val.fprice+'</span>'+val.c+'<p class="name">'+val.title+'</p><p>'+val.item_description+'</p></div></div></a></div>')
}
function ShowTaobaoBox(key){

  let obj = ShoppingMap.get('taobao')
  let val = obj.db[key]
  $('#myLargeModalLabel').modal('toggle')
  $('.modal-title').html(val.short_title)
  $('.carousel-inner').html('<div class="carousel-item active"><img src="'+val.pict_url+'" class="d-block w-100" ></div>')
  if (val.small_images){
  $.each(val.small_images.string,function(k,v){
    $('.carousel-inner').append('<div class="carousel-item"><img src="'+v+'" class="d-block w-100" ></div>')
  })
  }
  $('.text').html('<span class="badge badge-danger">￥'+val.price+'返'+val.fprice+'</span>')
  if (val.coupon_info)$('.text').append('<span  class="badge badge-warning">'+val.coupon_info+'</span>')
  $('.text').append('<p><span class="badge badge-secondary">'+val.provcity+' '+val.shop_title+'</span></p>')
  $('.text').append('<p>'+val.item_description+'</p>')
  $('.pmsg').hide(); 
  //if (!isWl()){
  //   $('.pmsg').html('<p><a  target="_blank" class="btn btn-info"  href="'+val.url_+'">点击这里 到'+val.tname+'领卷下单</a></p>')
  //   $('.pmsg').append('<p>关注微信公众号 米果推荐 zaddone_com,输入订单号获取返利详情</p>')
  //   $('.pmsg').show(); 
  //  return
  //}
  $.ajax({
    type: "get",
    dataType: "json",
    cache:false,
    url: '/goods/taobao',
    data:{"goodsid":val.url_,"ext":val.title},	  
    success: function(db){
     //console.log(db)
     //$('.pmsg').html("")
     let code = db.tbk_tpwd_create_response.data.model
     //$('.pmsg').html('<a href="#'+code+'"  class="btn btn-info" >'+code+'</a><p>长按上面代码拷贝复制后,打开'+val.tname+'应用查看</p>')
     $('.pmsg').html('<textarea>'+code+'</textarea><p>拷贝复制上面代码后,打开'+val.tname+'应用查看</p>')
     $('.pmsg').append('<p>关注微信公众号 米果推荐 zaddone_com,输入订单号获取返利详情</p>')

     if (isWeixin()){
     	ShowWX('.pmsg')
     }
     $('.pmsg').show(); 
     //obj.func(db,false)
    },
  });

}
