function PageHtml(data){
  if (!data || data.length === 0){
    $('.list').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">没有找到相关信息<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
    return
  }
 $('.list').html("")
 let that = this
  that.db = []
 $.each(data, function(key, val) {	
  that.db.push(val)
  that.html(key,val)
 })
}
function html(key,val){
  if (!val.Img){
	$('.list').append('<div class="col-lg-2 top" ><p class="name">'+val.Name+'</p><p>没有返利</p></div>')
  	return
  }
  $('.list').append('<div class="col-lg-2 top" ><a class="goods card btn btn-link" href="javascript:ShowCurrencyBox('+key+',\''+this.name+'\',\''+this.py+'\')"><span class="position-absolute"><span class="badge badge-secondary">'+this.name+'</span><span class="badge badge-dark">'+val.Tag+'</span></span><img src="'+val.Img[0]+'" class="card-img-top" alt="'+val.Name+'"><div class="overflow-hidden" style="height:100px"><div class="card-text"><span class="badge badge-danger">￥'+val.Price+'-'+val.Fprice+'</span><p class="name">'+val.Name+'</p></div></div></a></div>')
}
function ShowCurrencyBox(key,tname,tpy){
  let obj = ShoppingMap.get(tpy)
  let val = obj.db[key]
  $('#myLargeModalLabel').modal('toggle')
  $('.modal-title').html(val.Name)
  $('.carousel-inner').html("")
  if (val.Img && val.Img.length >0 ) {
    $('.carousel-inner').append('<div class="carousel-item active"><img src="'+val.Img.shift()+'" class="d-block w-100" ></div>')
    $.each(val.Img,function(k,v){
      $('.carousel-inner').append('<div class="carousel-item"><img src="'+v+'" class="d-block w-100" ></div>')
    })
  }
  $('.text').html('<span class="badge badge-danger">￥'+val.Price+'返'+val.Fprice+'</span>')
  $('.pmsg').html('<p><a  class="btn btn-info"   target="_blank"  href="https://www.zaddone.com/site/p/'+tpy+'/'+val.Id+'">点击这里 到"'+tname+'"下单</a></p>')
  $('.pmsg').append('<p>完成订单后,搜索订单号获取返利详情</p>')
  $('.pmsg').show(); 
  
}
