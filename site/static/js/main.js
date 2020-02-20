var ShoppingMap =  new Map()
//var publicObj;
var jdReg = new RegExp(/\/(\d+)\.html/);
var jdReg_ = new RegExp(/[\?|\&]sku=(\d+)/);
var pddReg = new RegExp(/[\?|\&]goods_id=(\d+)/);
var tbReg = new RegExp(/[\?|\&]id=(\d+)/);
var tburlReg = new RegExp(/(taobao|tmall|tb)/);
var urlReg = new RegExp(/(http[\S]+)[\+| ]?/);

function isWeixin () {
  let wx = navigator.userAgent.toLowerCase()
  if (wx.match(/MicroMessenger/i) === 'micromessenger') {
    return true
  } else {
    return false
  }
}
function isWl(){
  return (/Android|webOS|iPhone|iPod|BlackBerry/i.test(navigator.userAgent)) 
}
function ShowErr(func){

     $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">error<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
	
}
function ShowWX(id){

    $(id).append('<div class="card"><img src="/static/img/gzh.jpg" class="card-img-top" alt="米果推荐 购物查价"><div class="card-body"><p>米果推荐公众号</p></div></div>')
    return 
}
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
  //$('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
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
    console.log(_li)
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
  if (tburlReg.test(res)>=0){
    console.log(urlReg.exec(res))
    let rui_ =  urlReg.exec(res)
    ShowGoods("taobao",rui_[1])
    return true
  }
  
  return false
}

function ShowGoods(py,goodsid){
  let obj  = ShoppingMap.get(py) 
  if (!obj)return
  obj.key = goodsid
  //obj.funcHand = "ShowGoods"
  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  $.ajax({
    type: "get",
    dataType: "json",
    cache:false,
    url: '/goodsid/'+py,
    data:{"goodsid":goodsid},	  
    success: function(db){
     //console.log(db)
     $('.wait').html('')
     obj.func(db)
    },
    error:function(db){
     //console.log(db)
     $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">没有找到 <a href="javascript:ShowGoods(\''+py+'\',\''+goodsid+'\')"> 重试</a><button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
    },
  });
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
	obj.db.forEach(function(val,key){
		obj.html(key,val)
	})
    	return
    }
    if (!key){
    	//key = $(".p").text()
  	key = getQueryString("keyword")
	if (!key)return
    }
    obj.key = key
    jsonGet('/search/'+py+"?keyword="+encodeURI(key),obj)
}
function ShowSearch(){
    $('#form').toggle()
    $('#clear').toggle()
    $('#searchKey').focus()
    $('.list').html("")
}
function jsonGet(uri_,obj){

  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  $.ajax({
    type: "get",
    dataType: "json",
    //cache:false,
    url: uri_,
    //data:{"keyword":key_},	  
    success:function(db){
	obj.funcHand="jsonGetSearch"
 	$('.wait').html("")
	obj.func(db,true)
	//success_(db)
    },
    error:function(db){
      $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">没有找到 <a href="javascript:jsonGetSearch(\''+obj.py+'\',\''+obj.key+'\')"> 重试</a><button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
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
