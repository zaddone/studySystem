var ShoppingMap =  new Map()
var firstSite = "";
//var publicObj;
var urlReg = new RegExp(/(http[\:\.\/\?\=\-\w]+)[\+| ]?/);
var globalCode = "";

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
  //$("#siteName").text("")
  $('#p').html(key.substring(0,6))
  //$('#searchKey').val(key)
  //$('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  jsonGetSearch(py,key)
  //jsonGetSearch('jd',key)
  //jsonGet('/search/vip',key,VipPageHtml)
}

function checkInputIsUrl(word){
  //let pos = word.indexOf("http");
  //if (pos<0)return false;
  let rui_ =  urlReg.exec(word)
  //console.log(rui_,rui_.length)
  if (rui_ && rui_.length>1){
    console.log(rui_[1])
    $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
    $.ajax({
      type:"get",
      url:"https://www.zaddone.com/site/goodsurl",
      dataType:"jsonp",
      data:{"url":rui_[1]},
      success:function(db){
     	$('.wait').html('')
	console.log(db.py)
        let obj  = ShoppingMap.get(db.py) 
        obj.func(db.db)
      },
      error:function(db){
     //console.log(db)
     $('.wait').html('<div class="alert alert-warning alert-dismissible fade show" role="alert">没有找到 <a href="javascript:Search()"> 重试</a><button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button></div>')
    },
    })

    return true
  }
  return false
}

function ShowGoods(py,goodsid){
  let obj  = ShoppingMap.get(py) 
  //console.log(obj)
  if (!obj)return
  obj.key = goodsid
  //obj.funcHand = "ShowGoods"
  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  $.ajax({
    type: "get",
    dataType: "jsonp",
    cache:false,
    url: 'https://www.zaddone.com/site/goodsid/'+py,
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
    if (!py){
      $('#menu').toggle()
      $('#search').toggle()
      py = firstSite
    }
    let obj = ShoppingMap.get(py)
    if (!obj){
	return
        //jsonGetSearch('pinduoduo',key)
	//return
        //obj = ShoppingMap.get(firstSite)
	////console.log(obj)
        //$('#menu').toggle()
        //$('#search').toggle()
	//$('.wait').html("")
	//return
    }
    console.log(obj)
    window.scrollTo(0,0);
    $('#pyinput').val(obj.py)
    $('.active').removeClass('active');
    $('#nav'+obj.py).addClass('active');
    $("#dropdownMenuButton").html(obj.name+'<svg class="bi bi-caret-down-fill" width="1em" height="1em" viewBox="0 0 16 16" fill="currentColor" xmlns="http://www.w3.org/2000/svg"><path d="M7.247 11.14L2.451 5.658C1.885 5.013 2.345 4 3.204 4h9.592a1 1 0 0 1 .753 1.659l-4.796 5.48a1 1 0 0 1-1.506 0z"/></svg>');
    //$("#siteName").html(obj.name)
    if (obj.db.length>0){
	obj.db.forEach(function(val,k){
		obj.html(k,val)
	})
    	return
    }
    if (!key){
    	//key = $(".p").text()
  	key = getQueryString("keyword")
	if (!key)return
    }
    obj.key = key
    jsonGet('https://www.zaddone.com/site/search/'+py+"?keyword="+encodeURI(key)+"&ext="+globalCode,obj)
}
function ShowSearch(){
    $('#searchKey').focus()
    $('.list').html("")
}
function jsonGet(uri_,obj){

  $('.wait').html('<div class="col-lg-12 d-flex justify-content-center"><div class="spinner-border" role="status"> <span class="sr-only">Loading...</span></div></div>')
  $.ajax({
    type: "get",
    dataType: "jsonp",
    //cache:false,
    url: uri_,
    //data:{"keyword":key_},	  
    success:function(db){
	//obj.funcHand="jsonGetSearch"
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
function getSiteList(){
	$.ajax({
		type:"get",
		url:'https://www.zaddone.com/site/',
		data:{'content_type':'json'},
		dataType:"jsonp",
		cache:false,
		success:function(db){
		 db.forEach(function(k){
		  firstSite = k.py
		  let obj
		  if (k.py ==="taobao"){
		    obj = {func:eval(k.py+'PageHtml'),db:[],page:0,html:eval('html'+k.py),py:k.py,name:k.Name}
		  }else{
		    obj = {func:PageHtml,db:[],page:0,html:html,py:k.py,name:k.Name}
		  }
		  ShoppingMap.set(k.py,obj)
		  $('#siteMenu').prepend('<li class="nav-item"><a class="nav-link" tabindex="-1" aria-disabled="true" id="nav'+k.py+'" href="javascript:jsonGetSearch(\''+k.py+'\')">'+k.Name+'</a></li>')
		 })
		  
	        //console.log(firstSite)
		Search()
		}
	});
}
function getCiteCode(){
	var x = document.cookie;
	if (x.search('codecity') > -1)return
	//console.log(x.search('codecity'))
	//console.log(x)
	//if (x.search('codecity'))
	let u = 'https://ipservice.suning.com/ipQuery.do'
	$.ajax({
		type:"get",
		url:u,
		dataType:"jsonp",
		//jsonpCallback:"cookieCallback",
		success:function(db){
			console.log(db)
			//document.cookie = "codecity="+db.cityMDMId;
			//document.cookie = "codecity="+db.cityCommerceId;
			document.cookie = "codecity="+db.cityLESId;
			//if (db.cityLESId)globalCode=db.citylesid
		}
	});
}
