var ListDB = new Map();
ListDB.set("pinduoduo",{func:pddMode,li:[]})
ListDB.set("vip",{func:vipMode,li:[]})
ListDB.set("jd",{func:jdMode,li:[]})
var jdReg = new RegExp(/\/(\d+)\.html/);
var pddReg = new RegExp(/goods_id=(\d+)(\&|$)/);
function Search(){
  let key = getQueryString("keyword")
  if (isEmpty(key))return;
  if (checkInputIsUrl(key)) {
    return
  }
  $('#searchKey').val(key)
  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  jsonGet('/search/pinduoduo',key,PddPageHtml)
  jsonGet('/search/jd',key,JdPageHtml)
  //jsonGet('/search/vip',key,VipPageHtml)
}

$(document).ready(function(){
  Search()
  ShowMode()
});

function checkInputIsUrl(word){
  let pos = word.indexOf("http");
  if (pos<0)return false;
  word.slice(pos, word.length)
  let ret = parseQueryString(word)
  ret.goods_id
  let res = word.slice(pos, word.length);
  pos = res.indexOf("yangkeduo.com")
  if (pos>0){
    let _li = pddReg.exec(res)
    //let req = parseQueryString(res)
    pddGoods(_li[1])
    return true
  }
  pos = res.indexOf("jd.com")
  if (pos>0){
    //let req = parseQueryString(res)
    let _li = jdReg.exec(res)
    jdGoods(_li[1])
    return true
  }
  return false
}

function jdGoods(goodsid){
  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  $.getJSON({
    url: '/goodsid/jd',
    data:{"goodsid":goodsid},	  
    success: function(db){
     //console.log(db)
     JdPageHtml(db)
     $('.wait').html('')
    },
    error:function(db){
      $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">'+db+'<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
    },
  });
}
function pddGoods(goodsid){
  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  $.getJSON({
    url: '/goodsid/pinduoduo',
    data:{"goodsid":goodsid},	  
    success: function(db){
     //console.log(db)
     PddPageHtml(db)
     $('.wait').html('')
    },
    error:function(db){
      $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">'+db+'<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
    },
  });
}


function JdPageHtml(data){
	//http://img14.360buyimg.com/ads/
  let obj = ListDB.get("jd")
  obj.li=[]
 $.each(data.jd_kpl_open_xuanpin_searchgoods_response.result.queryVo, function(key, val) {	
  console.log(val)
  val.fprice = (parseFloat(val.commisionRatioWl)/100 * parseFloat(val.price)).toFixed(2);
  obj.li.push(val)
  //ListDB.push(val)
  $('.list').append('<div class="col-lg-2 top" ><div class="card"> <button data-toggle="modal" data-target=".bd-modal-lg" type="button" class="btn btn-link" data-id="'+key+'" data-site="jd" ><span class="position-absolute"><span class="badge badge-secondary">京东</span></span><img src="http://img14.360buyimg.com/ads/'+val.imageUrl+'" class="card-img-top" alt="'+val.wareName+'"><div class="overflow-hidden" style="height:100px"><p class="card-text"><span class="badge badge-danger">￥'+val.price+'-'+val.fprice+'</span></br>'+val.wareName+'</p></div></button></div></div>')
 })
}
function VipPageHtml(data){
  let obj = ListDB.get("vip")
  obj.li = []
  //ListDB["vip"].li =[]
 $.each(data, function(key, val) {	
  console.log(val)
  obj.li.push(val)
  //$('.list').append('<div class="col-lg-2 top" ><div class="card"> <button data-toggle="modal" data-target=".bd-modal-lg" type="button" class="btn btn-link" data-id="'+key+'" data-site="pinduoduo" ><span class="position-absolute"><span class="badge badge-secondary">拼多多</span><span class="badge badge-dark">'+val.mall_name+'</span></span><img src="'+val.goods_thumbnail_url+'" class="card-img-top" alt="'+val.goods_desc+'"><div class="overflow-hidden" style="height:100px"><p class="card-text"><span class="badge badge-danger">￥'+val.min_group_price+'-'+fprice+'</span></br>'+val.goods_name+'</p></div></button></div></div>')
 })
}
function PddPageHtml(db){
  let obj = ListDB.get("pinduoduo")
  obj.li=[]
  let data
  try{
  data = db.goods_search_response.goods_list
  }catch(err) {
  data = db.goods_detail_response.goods_details

  }
 $.each(data, function(key, val) {	
  val.min_group_price /= 100.0
  val.min_normal_price /= 100.0
  obj.li.push(val)
  val.fprice = (val.min_group_price*(val.promotion_rate/1000.0)).toFixed(2);
  //<span class="badge badge-dark">Dark</span>
  $('.list').append('<div class="col-lg-2 top" ><div class="card"> <button data-toggle="modal" data-target=".bd-modal-lg" type="button" class="btn btn-link" data-id="'+key+'" data-site="pinduoduo" ><span class="position-absolute"><span class="badge badge-secondary">拼多多</span><span class="badge badge-dark">'+val.mall_name+'</span></span><img src="'+val.goods_thumbnail_url+'" class="card-img-top" alt="'+val.goods_desc+'"><div class="overflow-hidden" style="height:100px"><p class="card-text"><span class="badge badge-danger">￥'+val.min_group_price+'-'+val.fprice+'</span></br>'+val.goods_name+'</p></div></button></div></div>')
  //$('.list').append('<div class="col-lg-2 top" ><div class="card"> <button data-toggle="modal" data-target=".bd-modal-lg" type="button" class="btn btn-link" data-id="'+key+'" data-site="pinduoduo" ><span class="position-absolute"><span class="badge badge-secondary">拼多多</span><span class="badge badge-dark">'+val.mall_name+'</span></span><img src="'+val.goods_thumbnail_url+'" class="card-img-top" alt="'+val.goods_desc+'"><div class="overflow-hidden" style="height:100px"><p class="card-text"><span class="badge badge-danger">￥'+val.min_group_price+'-'+fprice+'</span></br>'+val.goods_name+'</p></div></button></div></div>')
 })
}
function jsonGet(uri_,key_,success_){
  $.getJSON({
    url: uri_,
    data:{"keyword":key_},	  
    success:function(db){
 	$('.wait').html("")
	success_(db)
    },
    error:function(db){
      $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">'+db+'<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
    },
  });
}

function vipMode(that,site,goods){
}
function jdMode(that,site,goods){
    $(that).find('.modal-header h6').html('<span class="badge badge-secondary">'+goods.wareName+'</span><span class="badge badge-danger">￥'+goods.price+'-'+goods.fprice+'</span></br>'+goods.wareName)
    let innet = $(that).find('.carousel-inner')
    innet.html("")
    innet.append('<div class="carousel-item active"><img src="http://img14.360buyimg.com/ads/'+goods.imageUrl+'" class="d-block w-100" alt=""></div>')
    var modal = $(that)
    modal.find(".down").html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
    $.getJSON({
      url: '/goods/'+site,
      data:{"goodsid":goods.skuId},	  
      success: function(db){
       console.log(db)
	      //jd_kpl_open_promotion_pidurlconvert_response.
       let g = db.jd_kpl_open_promotion_pidurlconvert_response.clickUrl.clickURL
       //var down  = '<a target="_blank" class="btn btn-success" href="'+g+'">京东下单 返'+goods.promotion_rate/10.0+'%</a>'
       var down  = '<a target="_blank" class="btn btn-success" href="'+g+'">京东下单</a>'
       modal.find(".down").html(down)
       modal.find(".footerdown").html(down)
      },
      error:function(db){
       //console.log(db)
       modal.find(".modal-footer").html('<div class="alert alert-warning alert-dismissible fade show" role="alert">'+db+'<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
      },
    });
}

function pddMode(that,site,goods){
    $(that).find('.modal-header h6').html('<span class="badge badge-secondary">'+goods.mall_name+'</span><span class="badge badge-danger">￥'+goods.min_group_price+'-'+goods.fprice+'</span></br>'+goods.goods_name)
    let innet = $(that).find('.carousel-inner')
    innet.html("")
    innet.append('<div class="carousel-item active"><img src="'+goods.goods_image_url+'" class="d-block w-100" alt=""></div>')
    if (goods.goods_gallery_urls){
      goods.goods_gallery_urls.forEach(function(value,i){
        innet.append('<div class="carousel-item"><img src="'+value+'" class="d-block w-100" alt=""></div>')
      })
    }
    var modal = $(that)
    modal.find(".down").html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
    $.getJSON({
      url: '/goods/'+site,
      data:{"goodsid":goods.goods_id},	  
      success: function(db){
       console.log(db)
       let g = db.goods_promotion_url_generate_response.goods_promotion_url_list[0]
       //var down  = '<a target="_blank" class="btn btn-success" href="'+g.short_url+'">拼多多下单 返'+goods.promotion_rate/10.0+'%</a>'
       var down  = '<a target="_blank" class="btn btn-success" href="'+g.short_url+'">拼多多下单</a>'
        modal.find(".down").html(down)
        modal.find(".footerdown").html(down)
      },
      error:function(db){
       //console.log(db)
       modal.find(".modal-footer").html('<div class="alert alert-warning alert-dismissible fade show" role="alert">'+db+'<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
      },
    });
}
function ShowMode(){
  $("#myLargeModalLabel").on("show.bs.modal", function (event) {
    var button = $(event.relatedTarget)
    var site = button.data('site'); 
    let obj = ListDB.get(site)
    var goods = obj.li[parseInt(button.data('id'))]
    obj.func(this,site,obj.li[parseInt(button.data('id'))])
  })
}


function parseQueryString(url) {
 var reg_url = /^[^\?]+\?([\w\W]+)$/,
  reg_para = /([^&=]+)=([\w\W]*?)(&|$|#)/g,
  arr_url = reg_url.exec(url),
  ret = {};
 if (arr_url && arr_url[1]) {
  var str_para = arr_url[1], result;
  while ((result = reg_para.exec(str_para)) != null) {
   ret[result[1]] = result[2];
  }
 }
 return ret;
}
function getQueryString(name) {
 var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)", "i");
 var r = window.location.search.substr(1).match(reg);
 //console.log(r)
 if (r != null) return decodeURIComponent(r[2]);
 return null;
}
function trim(a){
  if(typeof a =='string'){
    return a.replace(/\s+/,'');
  }else {
    return a;
  }
}
function isEmpty(a){
  var b = trim(a);
  if((typeof a) == 'string'  && b){
    return false;
   }else {
      return true;
   }
}
function randomNum(minNum,maxNum){ 
    switch(arguments.length){ 
        case 1: 
            return parseInt(Math.random()*minNum+1,10); 
        break; 
        case 2: 
            return parseInt(Math.random()*(maxNum-minNum+1)+minNum,10); 
        break; 
            default: 
                return 0; 
            break; 
    } 
}
