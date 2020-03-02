package shopping
import(
	"fmt"
	"sort"
	"time"
	"crypto/md5"
	"io"
	//"sync"
	"io/ioutil"
	"encoding/json"
	"strings"
	//"bytes"
	//"strconv"
	//"regexp"
	"github.com/zaddone/studySystem/request"
	"net/url"
	//"github.com/boltdb/bolt"
	//"encoding/binary"
	"github.com/PuerkitoBio/goquery"
)
var (
	JdUrl = "https://router.jd.com/api"
	JdUrl_ = "https://api.jd.com/routerjson"
	//PddErrNum int = 0
	//pdd.ddk.theme.goods.search
	//pdd.ddk.goods.search
	//JdToken = "0619e9dd75e448dea0ab1b0449de3d89wu5z"

	//JdToken = "8fb30ead08284c52a879444d6a47c8bdywqw"
	//JdOrderDB *bolt.DB
	//dbTime = []byte("time")
	//dbUser = []byte("user")
	dbLast = []byte("last")
	dbPhone = []byte("Phone")
	orderTimeFormat = "2006010215"
	jdSiteid = "2009626993"
	//week = []string{""}


)

func NewJd(sh *ShoppingInfo) (ShoppingInterface){
	//fmt.Println("jd")
	return &Jd{Info:sh}
	//if !Open{
	//	return p
	//}
	//var err error
	//p.OrderDB,err = bolt.Open("jdorderDB",0600,nil)
	//if err != nil {
	//	panic(err)
	//}
	//return p
}
type Jd struct{
	Info *ShoppingInfo
	Pid string
	//OrderDB *bolt.DB
}

func (self *Jd)OrderMsg(_db interface{})(str string){
	db := _db.(map[string]interface{})
	db__ := db["skuList"].([]interface{})
	str = ""
	var sumFee float64
	for _,_db_:= range db__{
		db_:=_db_.(map[string]interface{})
		fee := db_["actualFee"].(float64)
		if db_["validCode"].(float64) == 17 {
			sumFee+=fee
		}
		str += fmt.Sprintf("%s\n￥%.2f\n佣金￥%.2f \n技术服务费￥%.2f\n",
			db_["skuName"].(string),
			db_["actualCosPrice"].(float64),
			fee,fee*0.1,
		)
	}
	if sumFee == 0 {
		str +=iMsg+"订单完成8日后返利\n"
	}else{
		finishTime :=time.Unix(int64(db["finishTime"].(float64)/1000),0).Add(time.Hour*24*8)
		//sumFee *= 0.9
		str += fmt.Sprintf("%s\n返￥%.2f\n预计%s到帐\n",
			iMsg,
			sumFee*0.9,
			finishTime.Format("1月2日"),
		)
	}
	return
}
func (self *Jd)addSign(u *url.Values){
	u.Add("app_key",self.Info.Client_id)
	//u.Add("access_token","8fb30ead08284c52a879444d6a47c8bdywqw")
	u.Add("format","json")
	now := time.Now().Add(-(time.Minute*2))
	u.Add("sign_method","md5")
	//time.Now().Format("2006-01-02 15:04:05")
	//timestamp
	u.Add("timestamp",now.Format(timeFormat))
	var li []string
	for k,_ := range *u {
		li = append(li,k)
	}
	sort.Strings(li)
	sign := self.Info.Client_secret
	for _,k :=range li {
		sign+=k+u.Get(k)
	}
	sign+=self.Info.Client_secret
	//fmt.Println(sign)
	u.Add("sign",fmt.Sprintf("%X", md5.Sum([]byte(sign))))
	//fmt.Println(u.Get("sign"))
}

func (self *Jd) ClientHttp(uri string,u *url.Values)( out interface{}){

	self.addSign(u)
	//ht := http.Header{}
	//ht.Add("Content-Type","application/json")
	var err error
	err = request.ClientHttp_(
		uri+"?"+u.Encode(),
		"GET",nil,
		nil,
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
func (self *Jd) SearchGoods(words ...string)interface{}{
	//self.ProductSearch(words...)
	u := &url.Values{}
	//jd.kpl.open.xuanpin.search.sku
	//jd.kpl.open.xuanpin.searchgoods
	u.Add("method","jd.kpl.open.xuanpin.searchgoods")
	u.Add("v","1.0")
	query := map[string]interface{}{
		"queryParam":map[string]interface{}{"keywords":words[0]},
		"pageParam":map[string]interface{}{"pageSize":20,"pageNum":1},
		"orderField":0,
	}
	u.Add("access_token",self.Info.Token)
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	//u.Add("360buy_param_json",fmt.Sprintf("{\"goodsReqDTO\":{\"keyword\":\"%s\"}}",words[0]))
	u.Add("param_json",string(body))
	//u.Add("custom_parameters",words[1])
	//data.jd_kpl_open_xuanpin_searchgoods_response.result.queryVo
	db := self.ClientHttp(JdUrl,u)
	if db == nil {
		return nil
	}
	res := db.(map[string]interface{})["jd_kpl_open_xuanpin_searchgoods_response"]
	if res == nil {
		return nil
	}
	return res.(map[string]interface{})["result"].(map[string]interface{})["queryVo"]
}
func (self *Jd) GoodsDetail(words ...string)interface{}{
	u := &url.Values{}
	u.Add("method","jd.kpl.open.xuanpin.searchgoods")
	u.Add("v","1.0")
	query := map[string]interface{}{
		//"queryParam":map[string]interface{}{"keywords":words[0]},
		"queryParam":map[string]interface{}{"skuId":words[0]},
		"pageParam":map[string]interface{}{"pageSize":1,"pageNum":1},
		"orderField":0,
	}
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	u.Add("param_json",string(body))
	//jd.kpl.open.item.getmobilewarestyleandjsbywareid
	u.Add("access_token",self.Info.Token)
	//return self.ClientHttp(JdUrl,u)
	db := self.ClientHttp(JdUrl,u)
	if db == nil {
		return nil
	}
	res := db.(map[string]interface{})["jd_kpl_open_xuanpin_searchgoods_response"]
	if res == nil {
		return nil
	}
	//return res["result"].(map[string]interface{})["queryVo"]
	return res.(map[string]interface{})["result"].(map[string]interface{})["queryVo"]
	//return nil
}
func (self *Jd) GoodsUrl(words ...string) interface{}{
	u := &url.Values{}
	u.Add("method","jd.union.open.promotion.common.get")
	u.Add("v","1.0")
	//u.Add("access_token",self.Info.Token)
	query := map[string]map[string]interface{}{
		"promotionCodeReq":map[string]interface{}{
		"siteId":jdSiteid,
		"materialId":fmt.Sprintf("https://item.jd.com/%s.html",words[0]),
	},
	}
	if len(words)>1 {
		query["promotionCodeReq"]["ext1"] = words[1]
	}
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	u.Add("param_json",string(body))
	//defer fmt.Println(u)
	return self.ClientHttp(JdUrl,u)
}
func (self *Jd) GoodsUrl_(words ...string) interface{}{

	u := &url.Values{}
	u.Add("method","jd.kpl.open.promotion.pidurlconvert")
	u.Add("v","2.0")
	u.Add("access_token",self.Info.Token)
	query := map[string]interface{}{
		"webId":"0",
		"positionId":0,
		"materalId":fmt.Sprintf("https://item.jd.com/%s.html",words[0]),
		"kplClick":1,
		"shortUrl":1,
	}

	if len(words)>1 {
		query["subUnionId"] = words[1]
	}
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	u.Add("param_json",string(body))
	return self.ClientHttp(JdUrl,u)

}
func (self *Jd)OrderSearch(keys ...string)(d interface{}){
	if len(keys)<2 {
		return
	}
	err := self.Info.orderGet(keys[0],keys[1],func(db interface{}){
		d = db
		//d = string(db.([]byte))
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return
}
func (self *Jd)OutUrl(db interface{}) string {
	//fmt.Println(db)
	res := db.(map[string]interface{})["jd_union_open_promotion_common_get_response"]
	if res == nil {
		fmt.Println("root is nil")
		return ""
	}
	res_ := res.(map[string]interface{})["result"]
	if res_ == nil {
		fmt.Println("result is nil")
		return ""
	}
	var res__ map[string]interface{}
	err := json.Unmarshal([]byte(res_.(string)),&res__)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return res__["data"].(map[string]interface{})["clickURL"].(string)


}
func (self *Jd)OutUrl_(db interface{}) string {
	//db.jd_kpl_open_promotion_pidurlconvert_response.clickUrl.clickURL
	res := db.(map[string]interface{})["jd_kpl_open_promotion_pidurlconvert_response"]
	if res == nil {
		return ""
	}
	res_ := res.(map[string]interface{})["clickUrl"]
	if res_ == nil {
		return ""
	}
	return res_.(map[string]interface{})["clickURL"].(string)

}
func(self *Jd)GetInfo()*ShoppingInfo {
	return self.Info
}
func (self *Jd) ProductSearch(words ...string)(result []interface{}){
	//https://search.jd.com/Search?keyword=
	u := &url.Values{}
	u.Add("keyword",words[0])
	err:= request.ClientHttp_("https://search.jd.com/Search?"+u.Encode(),"GET",nil,nil,func(body io.Reader,st int)error{
		_,err := goquery.NewDocumentFromReader(body)
		//db,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		//fmt.Println(string(db))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return nil

}

func (self *Jd) getOrder(t time.Time,page int)interface{}{
	u := &url.Values{}
	u.Add("method","jd.union.open.order.query")
	u.Add("v","1.0")
	query := map[string]interface{}{
		"orderReq":map[string]interface{}{
			"pageNo":page,
			"type":3,
			"time":t.Format(orderTimeFormat),
		},
	}
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	u.Add("param_json",string(body))
	return self.ClientHttp(JdUrl,u)

}
func (self *Jd) OrderDown(hand func(interface{}))error{
	//fmt.Println("jd down")
	var begin time.Time
	if self.Info.Update == 0 {
		var err error
		begin,err = time.Parse(timeFormat,"2020-02-03 16:00:00")
		if err != nil {
			panic(err)
		}
	}else{
		begin = time.Unix(self.Info.Update,0)
	}

	fmt.Println(begin)
	for{
		page := 1
		//fmt.Println(begin)
		for {
			db := self.getOrder(begin,page)
			if db == nil {
				continue
			}
			page++
			res := db.(map[string]interface{})["jd_union_open_order_query_response"]
			if res == nil {
				fmt.Println("response",db)
				return io.EOF
			}
			res_ := res.(map[string]interface{})["result"]
			if res_ == nil {
				fmt.Println("result",db)
				return io.EOF
			}
			var data map[string]interface{}
			err := json.Unmarshal([]byte(res_.(string)),&data)
			if err != nil {
				panic(err)
			}
			//fmt.Println(data)
			li := data["data"]
			if li == nil {
				break
			}
			li_ := li.([]interface{})
			for _,l := range li_ {
				l_ := l.(map[string]interface{})
				l_["order_id"] =fmt.Sprintf("%.0f", l_["orderId"].(float64))
				l_["status"] = false
				var goodid []string
				var sumFee float64
				for _, _db_:= range l_["skuList"].([]interface{}){
					db_:=_db_.(map[string]interface{})
					goodid = append(goodid,fmt.Sprintf("%.0f",db_["skuId"].(float64)))
					fee := db_["actualFee"].(float64)

					sumFee+=fee
					//if db_["validCode"].(float64) == 17 {
					//	l_["status"] = true
					//}
				}
				l_["goodsid"] = strings.Join(goodid,",")
				l_["fee"] = sumFee
				if l_["finishTime"].(float64)!=0 {
					l_["status"] = true
					l_["endTime"] = l_["finishTime"]
				}
				//fmt.Println(l_)
				hand(l_)
			}
			if len(li_) <20 {
				break
			}
		}
		begin = begin.Add(1*time.Hour)
		now := time.Now()
		if now.Unix()< begin.Unix() && now.Hour() < begin.Hour(){
			break
		}
	}
	self.Info.Update = begin.Unix()
	return nil
	//jd.union.open.order.query
	//return io.EOF
}
