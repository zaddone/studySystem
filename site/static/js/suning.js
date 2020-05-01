function suningPageHtml(data){
  if (data.length === 0){
        $('.list').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">没有找到相关信息<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
  	return
  }
 let that = this
  that.db = []
 $.each(data, function(key, val) {	
 //val.fprice = (parseFloat(val.commisionRatioWl)/100 * parseFloat(val.price)).toFixed(2);
 //val.imgurl = 'https://img14.360buyimg.com/ads/'+val.imageUrl
  //if (val.coupon_discount)val.c = '<span class="badge badge-warning">券</span>';
  //console.log(val)
  //val.Fprice = val.Fprice.toFixed(2)
  that.db.push(val)
  that.html(key,val)
 })
}
function htmlsuning(key,val){
  if (!val.Img){
	$('.list').append('<div class="col-lg-2 top" ><p class="name">'+val.Name+'</p><p>没有返利</p></div>')
  	return
  }
  $('.list').append('<div class="col-lg-2 top" ><a class="goods card btn btn-link" href="javascript:ShowSuningBox('+key+')"><span class="position-absolute"><span class="badge badge-secondary">苏宁</span><span class="badge badge-dark">'+val.Tag+'</span></span><img src="'+val.Img[0]+'" class="card-img-top" alt="'+val.Name+'"><div class="overflow-hidden" style="height:100px"><div class="card-text"><span class="badge badge-danger">￥'+val.Price+'-'+val.Fprice+'</span><p class="name">'+val.Name+'</p></div></div></a></div>')
}
function ShowSuningBox(key){
  let obj = ShoppingMap.get('suning')
  let val = obj.db[key]
  $('#myLargeModalLabel').modal('toggle')
  $('.modal-title').html(val.Name)
  if (val.Img && val.Img.length >0 ) {
    $('.carousel-inner').append('<div class="carousel-item active"><img src="'+val.Img.shift()+'" class="d-block w-100" ></div>')
    $.each(val.Img,function(k,v){
      $('.carousel-inner').append('<div class="carousel-item"><img src="'+v+'" class="d-block w-100" ></div>')
    })
  }
  //$('.carousel-inner').html('<div class="carousel-item active"><img src="'+val.Img+'" class="d-block w-100" ></div>')
  $('.text').html('<span class="badge badge-danger">￥'+val.Price+'返'+val.Fprice+'</span>')
  $('.text').append('<p>'+val.Show+'</p>')
  $('.pmsg').html('<p><a  class="btn btn-info"   target="_blank"  href="https://www.zaddone.com/p/suning/'+val.Id+'">点击这里 到"苏宁"下单</a></p>')
  $('.pmsg').append('<p>关注微信公众号 米果推荐 zaddone_com,输入订单号获取返利详情</p>')
  $('.pmsg').show(); 
  
}
