var ShoppingMap =  new Map()
//var publicObj;
var jdReg = new RegExp(/\/(\d+)\.html/);
var jdReg_ = new RegExp(/sku=(\d+)/);
var pddReg = new RegExp(/goods_id=(\d+)(\&|$)/);

function Search(){
  let key = getQueryString("keyword")
  let py = getQueryString("py")
  if (isEmpty(key))return;
  if (checkInputIsUrl(key)) {
    return
  }
  $("#nav"+py).addClass('active');
  $('#p').html(key.substring(0,6))
  //$('#searchKey').val(key)
  $('#clear').toggle()
  $('#form').toggle()
  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  jsonGetSearch(py,key)
  //jsonGetSearch('jd',key)
  //jsonGet('/search/vip',key,VipPageHtml)
}
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
    ShowGoods("pinduoduo",_li[1])
    return true
  }
  pos = res.indexOf("jd.com")
  if (pos>0){
    //let req = parseQueryString(res)
    let _li = jdReg.exec(res)
    if (!_li){
    	_li = jdReg_.exec(res)
    }
    if (!_li){
	    return
    }
    //console.log(_li)
    //jdGoods(_li[1])
    ShowGoods("jd",_li[1])
    return true
  }
  return false
}

function ShowGoods(py,goodsid){
  let obj  = ShoppingMap.get(py) 
  if (!obj)return
  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  $.getJSON({
    url: '/goodsid/'+py,
    data:{"goodsid":goodsid},	  
    success: function(db){
     //console.log(db)
     $('.wait').html('')
     obj.func(db,false)
    },
    error:function(db){
      $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">'+db+'<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
    },
  });
}
function jdPageHtml(data){

 let that = this
 $.each(data.jd_kpl_open_xuanpin_searchgoods_response.result.queryVo, function(key, val) {	
  val.fprice = (parseFloat(val.commisionRatioWl)/100 * parseFloat(val.price)).toFixed(2);
  //console.log(val)
  that.db.push(val)
  that.html(val)
 })
}
function htmljd(val){
  $('.list').append('<div class="col-lg-2 top" ><div class="card"> <a class="btn btn-link" target="_blank" href="https://www.zaddone.com/p/jd/'+val.skuId+'" ><span class="position-absolute"><span class="badge badge-secondary">京东</span></span><img src="http://img14.360buyimg.com/ads/'+val.imageUrl+'" class="card-img-top" alt="'+val.wareName+'"><div class="overflow-hidden" style="height:100px"><p class="card-text"><span class="badge badge-danger">￥'+val.price+'-'+val.fprice+'</span></br>'+val.wareName+'</p></div></a></div></div>')
}
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
  that.db.push(val)
  that.html(val)
 })
}
function htmlpinduoduo(val){
  $('.list').append('<div class="col-lg-2 top" ><div class="card"><a class="btn btn-link" target="_blank" href="https://www.zaddone.com/p/pinduoduo/'+val.goods_id+'" ><span class="position-absolute"><span class="badge badge-secondary">拼多多</span><span class="badge badge-dark">'+val.mall_name+'</span></span><img src="'+val.goods_thumbnail_url+'" class="card-img-top" alt="'+val.goods_desc+'"><div class="overflow-hidden" style="height:100px"><p class="card-text"><span class="badge badge-danger">￥'+val.min_group_price+'-'+val.fprice+'</span></br>'+val.goods_name+'</p></div></button></div></div>')
}
function jsonGetSearch(py,key){
    $('.list').html("")
    let obj = ShoppingMap.get(py)
    if (!obj){
        jsonGetSearch('pinduoduo',key)
	return
        //obj = ShoppingMap.get("jd")
	//console.log(obj)
        //$('#form').toggle()
        //$('#clear').toggle()
	//$('.wait').html("")
	//return
    }
    $('#pyinput').val(obj.py)
    $('.dropdown-item').removeClass('active');
    $('#nav'+obj.py).addClass('active');
    $("#dropdownMenuButton").html(obj.name);
    if (obj.db.length>0){
	obj.db.forEach(function(val){
		obj.html(val)
	})
    	return
    }
    if (!key){
    	//key = $(".p").text()
  	key = getQueryString("keyword")
	if (!key)return
    }
    jsonGet('/search/'+py,key,obj)
}
function ShowSearch(){
    $('#form').toggle()
    $('#clear').toggle()
    $('#searchKey').focus()
    $('.list').html("")
}
function jsonGet(uri_,key_,obj){
  $.getJSON({
    url: uri_,
    data:{"keyword":key_},	  
    success:function(db){
 	$('.wait').html("")
	obj.func(db,true)
	//success_(db)
    },
    error:function(db){
      $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">'+db+'<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
    },
  });
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
