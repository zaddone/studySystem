function pinduoduoPageHtml(data){
  if (data.length === 0){
        $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">没有找到<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
  	return
  }
  let that = this
  that.db = []
 $.each(data, function(key, val) {	
  val.c="";
  val.Fprice = val.Fprice.toFixed(2)
  if (val.Coupon)val.c = '<span class="badge badge-warning">券</span>';
  that.db.push(val)
  that.html(key,val)
 })
}
function htmlpinduoduo(key,val){
  $('.list').append('<div class="col-lg-2 top" ><a class="goods card btn btn-link" href="javascript:ShowPddBox('+key+')" ><span class="position-absolute"><span class="badge badge-danger">拼多多</span><span class="badge badge-dark">'+val.Tag+'</span></span><img src="'+val.Img+'" class="card-img-top" ><div class="overflow-hidden" style="height:100px"><div class="card-text"><span class="badge badge-danger">￥'+val.Price+'-'+val.Fprice+'</span>'+val.c+'<p class="name">'+val.Name+'</p></div></div></a></div>')
}
function ShowPddBox(key){
  let obj = ShoppingMap.get('pinduoduo')
  let val = obj.db[key]
  $('#myLargeModalLabel').modal('toggle')
  $('.modal-title').html(val.Name)
  $('.carousel-inner').html('<div class="carousel-item active"><img src="'+val.Img+'" class="d-block w-100" ></div>')
  $('.text').html('<span class="badge badge-danger">￥'+val.Price+'返'+val.Fprice+'</span>'+val.c)
  $('.text').append('<p><span class="badge badge-secondary">'+val.Tag+'</span></p>')
  //$('.text').append('<p>'+val.goods_desc+'</p>')
  $('.pmsg').html('<p><a  target="_blank"  class="btn btn-info"  href="https://www.zaddone.com/p/pinduoduo/'+val.Id+'">点击这里 到"拼多多"下单</a></p>')
  $('.pmsg').append('<p>关注微信公众号 米果推荐 zaddone_com,输入订单号获取返利详情</p>')
  $('.pmsg').show(); 
}
