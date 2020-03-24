package shopping
import(
	"github.com/zaddone/studySystem/request"
	"encoding/base64"
	"io"
	"io/ioutil"
	"time"
	"net/url"
	"strings"
	"net/http"
	"crypto/md5"
	"strconv"
	"bytes"
	"encoding/json"
	"fmt"
)
var (
	sunUri = "https://open.suning.com/api/http/sopRequest/"
	uriVal =&url.Values{}
)

func NewSuning(sh *ShoppingInfo) (ShoppingInterface){
	return &Suning{
		Info:sh,
		u:url.Values{
			"appkey":[]string{sh.Client_id},
			"versionNo":[]string{"v1.2"},
		},
		//pid:"658414",
	}
}
type Suning struct{
	Info *ShoppingInfo
	u url.Values
	//pid string
}

func (self *Suning)addSign(body []byte){
	body_ := base64.StdEncoding.EncodeToString(body)
	//u.Add("appSecret",self.Info.Client_secret)
	self.u.Set("appRequestTime",time.Now().Format(timeFormat))
	//u.Get("appkey")
	sign := self.Info.Client_secret+self.u.Get("appMethod")+self.u.Get("appRequestTime")+self.u.Get("appkey")+self.u.Get("versionNo")+body_
	self.u.Set("signInfo",fmt.Sprintf("%x", md5.Sum([]byte(sign))))
}

func (self *Suning) ClientHttp(body []byte)( out interface{}){

	self.addSign(body)
	ht := http.Header{}
	ht.Add("Content-Type","application/json;charset=utf-8")
	//apiMeth:=self.u.Get("appMethod")
	//self.u.Del("appMethod")
	for k,_ := range self.u{
		ht.Set(k,self.u.Get(k))
	}
	//fmt.Println("http",self.u)
	//fmt.Println(ht)
	var err error
	err = request.ClientHttp_(
		"https://open.suning.com/api/http/sopRequest",
		"POST",bytes.NewReader(body),
		ht,
		func(body io.Reader,start int)error{
		if start != 200 {
			db,err := ioutil.ReadAll(body)
			if err!= nil {
				return err
			}
			return fmt.Errorf("%s",db)
		}
		return json.NewDecoder(body).Decode(&out)
	})
	if err != nil {
		fmt.Println(err,out)
		out = err
		//time.Sleep(time.Second*1)
		//return self.ClientHttp(u)
		//panic(err)
	}
	return

}
func (self *Suning)GetInfo()*ShoppingInfo{
	return self.Info
}
func (self *Suning) stuctured(data interface{}) (g Goods){
	db := data.(map[string]interface{})
	info := db["commodityInfo"].(map[string]interface{})
	fmt.Println(db)
	f,err := strconv.ParseFloat(info["rate"].(string),64)
	if err != nil {
		panic(err)
	}
	price := info["commodityPrice"]
	if price == nil {
		price = info["snPrice"]
		if price == nil {
			price = "0"
		}
	}
	p,err := strconv.ParseFloat(price.(string),64)
	if err != nil {
		panic(err)
	}
	g = Goods{
		Id:info["supplierCode"].(string)+"-"+info["commodityCode"].(string),
		Name:info["commodityName"].(string),
		//Img:[]string{
		//	info["pictureUrl"].(map[string]interface{})["picUrl"].(string),
		//},
		Price:p,
		Fprice:f,
		Tag:info["supplierName"].(string),
		Show:info["sellingPoint"].(string),
	}
	if info["pictureUrl"] != nil{
		switch pic := info["pictureUrl"].(type){
		case []interface{}:
			for _, p := range pic{
				g.Img = append(g.Img,p.(map[string]interface{})["picUrl"].(string))
			}
		}


	}
	if db["couponInfo"] != nil && db["couponInfo"].(map[string]interface{})["couponUrl"] != nil{
		g.Coupon = true
		g.Ext = db["couponInfo"].(map[string]interface{})["couponUrl"].(string)
	}
	return g

}
func (self *Suning)SearchGoods(words ...string)interface{}{
	//https://ipservice.suning.com/ipQuery.do?callback=cookieCallback1

	self.u.Set("appMethod","suning.netalliance.searchcommodity.query")
	_body :=map[string]interface{}{
		//"cityCode":words[1],
		"keyword":words[0],
		//"sortType":2,
		//"suningService":1,
		//"size":20,
	}
	if len(words)>1{
		_body["cityCode"] = words[1]
	}
	body := map[string]interface{}{
		"sn_request":map[string]interface{}{
			"sn_body":map[string]interface{}{
				"querySearchcommodity":_body,
			},
		},
	}
	fmt.Println(body)
	b,err:= json.Marshal(body)
	if err != nil {
		return nil
	}
	db := self.ClientHttp(b)
	//fmt.Println(db)
	db_ := db.(map[string]interface{})["sn_responseContent"].(map[string]interface{})["sn_body"]
	if db_ == nil {
		return nil
	}
	var li []interface{}
	for _,b :=range db_.(map[string]interface{})["querySearchcommodity"].([]interface{}){
		li = append(li,self.stuctured(b))
	}
	return li

}
func (self *Suning)GoodsAppMini(words ...string)interface{}{
	self.u.Set("appMethod","suning.netalliance.appletextensionlink.get")
	db := map[string]interface{}{
		"productUrl":"https://product.suning.com/"+strings.Replace(words[0],"-","/",-1)+".html",
		"subUser":words[1],
	}
	if len(words)>2{
		db["quanUrl"] = words[2]
	}
	body := map[string]interface{}{
		"sn_request":map[string]interface{}{
			"sn_body":map[string]interface{}{
				"getAppletextensionlink":db,
			},
		},
	}
	b,err:= json.Marshal(body)
	if err != nil {
		return nil
	}
	db_ := self.ClientHttp(b)
	if db_ == nil {
		return nil
	}
	d := db_.(map[string]interface{})["sn_responseContent"].(map[string]interface{})["sn_body"]
	if d == nil {
		return nil
	}
	d_ := d.(map[string]interface{})["getAppletextensionlink"].(map[string]interface{})
	return map[string]interface{}{
		"appid":d_["appId"].(string),
		"url":d_["appletExtendUrl"].(string),
	}
	//return db

}
func (self *Suning)GoodsUrl(words ...string)interface{}{
	if words[len(words)-1] =="mini"{
		return self.GoodsAppMini(words[:len(words)-1]...)
	}
	db := map[string]interface{}{
		"productUrl":"https://product.suning.com/"+strings.Replace(words[0],"-","/",-1)+".html",
		"subUser":words[1],
	}
	if len(words)>2{
		db["quanUrl"] = words[2]
	}
	//suning.netalliance.storepromotionurl.query
	self.u.Set("appMethod","suning.netalliance.extensionlink.get")
	body := map[string]interface{}{
		"sn_request":map[string]interface{}{
			"sn_body":map[string]interface{}{
				"getExtensionlink":db,
			},
		},
	}
	b,err:= json.Marshal(body)
	if err != nil {
		return nil
	}
	db_ := self.ClientHttp(b)
	if db_ == nil {
		return nil
	}
	d := db_.(map[string]interface{})["sn_responseContent"].(map[string]interface{})["sn_body"]
	if d == nil {
		return nil
	}
	d_ := d.(map[string]interface{})["getExtensionlink"].(map[string]interface{})
	return map[string]interface{}{
		//"appid":d_["appId"].(string),
		"url":d_["shortLink"].(string),
	}
}
func (self *Suning)GoodsDetail(words ...string)interface{}{
	//suning.netalliance.unioninfomation.get
	self.u.Set("appMethod","suning.netalliance.unioninfomation.get")
	body := map[string]interface{}{
		"sn_request":map[string]interface{}{
			"sn_body":map[string]interface{}{
				"getUnionInfomation":map[string]interface{}{
					"goodsCode":strings.Split(words[0],"-")[1],
				},
			},
		},
	}
	b,err:= json.Marshal(body)
	if err != nil {
		return nil
	}
	db := self.ClientHttp(b)
	fmt.Println(db)
	db_ := db.(map[string]interface{})["sn_responseContent"].(map[string]interface{})["sn_body"]
	if db_ == nil {
		return nil
	}
	db_ = db_.(map[string]interface{})["getUnionInfomation"]
	if db_ == nil {
		return nil
	}
	query := db_.([]interface{})[0].(map[string]interface{})

	p := query["suningPrice"].(float64)
	f := query["rate"].(float64)
	var t string
	if query["operatingModel"].(float64) == 1{
		t = "苏宁自营"
	}else{
		t = query["vendorName"].(string)
	}
	return []interface{}{
		Goods{
		Id:query["mertCode"].(string)+"-"+query["goodsCode"].(string),
		Name:query["goodsName"].(string),
		Price:p,
		Fprice:f,
		Tag:t,
		Show:query["promoDesc"].(string),
		Img:strings.Split(query["pictureUrl"].(string),","),
		Coupon:false,
		},
	}
}
func (self *Suning)OrderSearch(words ...string)interface{}{
	return nil
}
func (self *Suning)OutUrl(v interface{}) string{
	return v.(map[string]interface{})["url"].(string)
}
func (self *Suning)OrderMsg(interface{}) string{
	return ""
}
func (self *Suning)ProductSearch(k ...string)[]interface{}{
	return nil
}
func (self *Suning)OrderDown(hand func(interface{}))error{
	return nil
}
