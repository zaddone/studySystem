package shopping
import(
	"fmt"
	"sort"
	"time"
	"crypto/md5"
	"io"
	"io/ioutil"
	"encoding/json"
	//"strings"
	//"bytes"
	"github.com/zaddone/studySystem/request"
	"net/url"
	//"github.com/boltdb/bolt"
	//"net/http"
	//"regexp"
)
var (
	PddUrl = "https://gw-api.pinduoduo.com/api/router"
	//PddOrderDB *bolt.DB
	//PddErrNum int = 0
	//pdd.ddk.theme.goods.search
	//pdd.ddk.goods.search
)
type Pdd struct{
	Info *ShoppingInfo
	PddPid []string
	//OrderDB *bolt.DB
}
func NewPdd(sh *ShoppingInfo) (ShoppingInterface) {
	return &Pdd{Info:sh}
	//if !o {
	//	return p
	//}
	//var err error
	//p.OrderDB,err = bolt.Open("pddOrder",0600,nil)
	//if err != nil {
	//	panic(err)
	//}
	//return p
}

func (self *Pdd)addSign(u *url.Values){
	u.Set("client_id",self.Info.Client_id)
	u.Set("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
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
	u.Set("sign",fmt.Sprintf("%X", md5.Sum([]byte(sign))))
}
func (self *Pdd) ClientHttp(u *url.Values)( out interface{}){

	self.addSign(u)
	//ht := http.Header{}
	//ht.Add("Content-Type","application/json")
	var err error
	err = request.ClientHttp_(
		PddUrl+"?"+u.Encode(),
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
func (self *Pdd) getPid()error {
	req := self.pidQuery()
	switch r := req.(type){
	case error:
		return r
	}
	pid := (req.(map[string]interface{})["p_id_query_response"]).(map[string]interface{})
	if pid["total_count"].(float64) >0 {
		for _,p_ := range pid["p_id_list"].([]interface{}){
			self.PddPid = append(self.PddPid,(p_.(map[string]interface{})["p_id"]).(string))
		}
		return nil
	}
	req = self.pidGenerate(1)
	switch r := req.(type){
	case error:
		return r
	}
	_pid := ((req.(map[string]interface{})["p_id_generate_response"]).(map[string]interface{})["p_id_list"]).([]interface{})
	for _,p_ := range _pid{
		self.PddPid = append(self.PddPid,(p_.(map[string]interface{})["p_id"]).(string))
	}

	return nil

}
//pdd.ddk.goods.pid.generate
func (self *Pdd) pidGenerate(n int) interface{}{
	u := &url.Values{}
	u.Add("type","pdd.ddk.goods.pid.generate")
	u.Add("number",fmt.Sprintf("%d",n))
	return self.ClientHttp(u)
}
//pdd.ddk.goods.pid.query
func (self *Pdd) pidQuery() interface{}{
	u := &url.Values{}
	u.Add("type","pdd.ddk.goods.pid.query")
	return self.ClientHttp(u)
}
func (self *Pdd)GoodsAppMini(words ...string)interface{}{
	goodsid := words[0]
	if len(self.PddPid) == 0 {
		err := self.getPid()
		if err != nil {
			return err
		}
	}
	u := &url.Values{}
	u.Add("type","pdd.ddk.goods.promotion.url.generate")
	u.Add("goods_id_list","["+goodsid+"]")
	u.Add("p_id",self.PddPid[0])
	u.Add("generate_short_url","true")
	//if words[len(words)-1] =="mini"{
	//	return self.GoodsAppMini(words[:len(words)-1]...)
	u.Add("generate_we_app","true")
	//}
	//if multi{
	u.Add("multi_group","true")
	if len(words)>1 {
		u.Add("custom_parameters",words[1])
	}
	//}
	db := self.ClientHttp(u)
	res := db.(map[string]interface{})["goods_promotion_url_generate_response"]
	if res == nil {
		return nil
	}
	res_ := res.(map[string]interface{})["goods_promotion_url_list"]
	if res == nil {
		return nil
	}
	res__ := res_.([]interface{})
	if len(res__)== 0 {
		return nil
	}

	res___ := res__[0].(map[string]interface{})
	if res___ == nil {
		return nil
	}
	if res___["we_app_info"] == nil {
		return nil
	}
	app := res___["we_app_info"].(map[string]interface{})
	return map[string]interface{}{
		"appid":app["app_id"].(string),
		"url":app["page_path"].(string),
	}

}
//pdd.ddk.goods.promotion.url.generate
func (self *Pdd) GoodsUrl(words ...string) interface{}{

	goodsid := words[0]
	if len(self.PddPid) == 0 {
		err := self.getPid()
		if err != nil {
			return err
		}
	}
	u := &url.Values{}
	u.Add("type","pdd.ddk.goods.promotion.url.generate")
	u.Add("goods_id_list","["+goodsid+"]")
	u.Add("p_id",self.PddPid[0])
	u.Add("generate_short_url","true")
	//if words[len(words)-1] =="mini"{
	//	return self.GoodsAppMini(words[:len(words)-1]...)
	//	u.Add("generate_we_app","true")
	//}
	//if multi{
	u.Add("multi_group","true")
	if len(words)>1 {
		u.Add("custom_parameters",words[1])
	}
	//}
	db := self.ClientHttp(u)
	res := db.(map[string]interface{})["goods_promotion_url_generate_response"]
	if res == nil {
		return nil
	}
	res_ := res.(map[string]interface{})["goods_promotion_url_list"]
	if res == nil {
		return nil
	}
	res__ := res_.([]interface{})
	if len(res__)== 0 {
		return nil
	}

	res___ := res__[0].(map[string]interface{})
	if res___ == nil {
		return nil
	}
	//if res___["we_app_info"] != nil {
	//	app := res___["we_app_info"].(map[string]interface{})
	//	return map[string]interface{}{
	//		"appid":app["app_id"].(string),
	//		"url":app["page_path"].(string),
	//	}
	//}

	return res___
	//return res___["short_url"].(string)


}

func (self *Pdd)OrderMsg(_db interface{}) (str string){
	db := _db.(map[string]interface{})
	res := db["order_detail_response"].(map[string]interface{})

	fee := res["promotion_amount"].(float64)/100
	str = fmt.Sprintf("%s\n￥%.2f\n佣金￥%.2f \n技术服务费￥%.2f\n",
		res["goods_name"].(string),
		res["order_amount"].(float64)/100,
		fee,fee*0.1,
	)
	if res["order_status"].(float64) == 2{
		finishTime :=time.Unix(int64(db["order_receive_time"].(float64)/1000),0).Add(time.Hour*24*15)
		str += fmt.Sprintf("%s\n返￥%.2f\n预计%s到帐\n",
			iMsg,
			fee*0.9,
			finishTime.Format("1月2日"),
		)
	}else{
		str +=iMsg+"订单完成15日后返利\n"
	}
	return str
}

func (self *Pdd) stuctured(data interface{}) (g Goods){
	d_ := data.(map[string]interface{})
	//p:= d_["min_group_price"].(float64)/100
	return Goods{
		Id:fmt.Sprintf("%.0f",d_["goods_id"].(float64)),
		Img:[]string{d_["goods_thumbnail_url"].(string)},
		Name:d_["goods_name"].(string),
		Tag:d_["mall_name"].(string),
		Price:d_["min_group_price"].(float64)/100,
		Fprice:d_["promotion_rate"].(float64)/1000.0,
		Coupon:d_["coupon_discount"].(float64)>0,
		//Show:d_["goods_desc"].(string),
	}
}

func (self *Pdd) SearchGoods(words ...string)interface{}{
	db :=  self.searchGoods(words...)
	if db == nil {
		return nil
	}
	res := db.(map[string]interface{})["goods_search_response"]
	if res == nil {
		return nil
	}
	var li []interface{}
	for _,d := range res.(map[string]interface{})["goods_list"].([]interface{}){
		li = append(li,self.stuctured(d))
	}
	return li
	//return res.(map[string]interface{})["goods_list"].([]interface{})
}
func (self *Pdd) searchGoods(words ...string)interface{}{
	u := &url.Values{}
	u.Add("type","pdd.ddk.goods.search")
	u.Add("keyword",words[0])
	u.Add("page_size","30")
	//u.Add("custom_parameters",words[1])
	return self.ClientHttp(u)
}
//pdd.ddk.goods.detail

func (self *Pdd) goodsDetail(words ...string)interface{}{
	goodsid := words[0]
	u := &url.Values{}
	u.Add("type","pdd.ddk.goods.detail")
	u.Add("goods_id_list","["+goodsid+"]")
	return self.ClientHttp(u)
}
func (self *Pdd) GoodsDetail(words ...string)interface{}{
	db := self.goodsDetail(words...)
	if db == nil {
		return nil
	}
	res := db.(map[string]interface{})["goods_detail_response"]
	if res == nil {
		return nil
	}

	var li []interface{}
	for _,d := range res.(map[string]interface{})["goods_details"].([]interface{}){
		li = append(li,self.stuctured(d))
	}
	return li
	//return res.(map[string]interface{})["goods_details"]
	//db_.goods_detail_response.goods_details
}

func (self *Pdd)OrderSearch(keys ...string)(d interface{}){
	//pdd.ddk.order.detail.get
	if len(keys)<2 {
		return nil
	}
	err := orderGet(keys[0],keys[1],func(db interface{}){
		d = db
		//d = string(db.([]byte))
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return
	//return nil
}
func (self *Pdd)OutUrl(db interface{}) string {
	return db.(map[string]interface{})["short_url"].(string)
	//res := db.(map[string]interface{})["goods_promotion_url_generate_response"]
	//if res == nil {
	//	return ""
	//}
	//res_ := res.(map[string]interface{})["goods_promotion_url_list"]
	//if res == nil {
	//	return ""
	//}
	//res__ := res_.([]interface{})
	//if len(res__)== 0 {
	//	return ""
	//}
	//res___ := res__[0].(map[string]interface{})
	//if res___ == nil {
	//	return ""
	//}
	//return res___["short_url"].(string)
}
func(self *Pdd)GetInfo()*ShoppingInfo {
	return self.Info
}

func (self *Pdd) ProductSearch(words ...string)(result []interface{}){
	return self.searchGoods(words...).([]interface{})
}
func (self *Pdd) OrderDown(hand func(interface{}))error{
	var begin,end time.Time
	if self.Info.Update == 0 {
		var err error
		begin,err = time.Parse(timeFormat,"2020-01-01 00:00:00")
		if err != nil {
			panic(err)
		}
	}else{
		begin = time.Unix(self.Info.Update,0)
	}
	//self.Info.Update = end.Unix()
	for{
		isOut := false
		end = begin.Add(24*time.Hour)
		Now := time.Now()
		if !Now.After(end){
			//fmt.Println()
			end = Now
			isOut = true
		}
		//fmt.Println(begin,end)
		page := 1
		for {
			db := self.getOrder(begin,end,page)
			if db == nil {
				continue
			}
			res := db.(map[string]interface{})["order_list_get_response"]
			if res == nil {
				fmt.Println(db)
				return io.EOF
			}
			li := res.(map[string]interface{})["order_list"].([]interface{})
			for _,l := range li{
				l_ := l.(map[string]interface{})
				l_["order_id"] = l_["order_sn"]
				l_["status"] = false
				//if l_["order_status"].(float64) == 2{
				//	l_["status"] = true
				//	l_["endTime"] = l_["order_receive_time"]
				//}
				l_["fee"] = l_["promotion_amount"].(float64)/100
				l_["goodsid"] =fmt.Sprintf("%.0f",l_["goods_id"].(float64))
				l_["goodsName"] = l_["goods_name"]
				l_["goodsImg"] = l_["goods_thumbnail_url"]
				l_["site"] = self.Info.Py
				l_["userid"] = l_["custom_parameters"]
				l_["time"] = time.Now().Unix()
				l_["text"] = l_["order_status_desc"]
				//if l_["order_verify_time"] != nil {
				if l_["order_receive_time"] != nil {
					//l_["status"] = true
					l_["endTime"] = int64(l_["order_receive_time"].(float64))
					var ver time.Time
					if l_["order_verify_time"] == nil {
						ver = time.Unix(int64(l_["order_receive_time"].(float64)),0)
					}else{
						ver = time.Unix(int64(l_["order_verify_time"].(float64)),0)
					}
					y,m,d := ver.Date()
					if d >15{
						y,m,_ = ver.AddDate(0,1,0).Date()
					}
					fmt.Println(y,m)
					l_["payTime"] = time.Date(y,m,21,0,0,0,0,ver.Location()).Unix()
				}
				hand(l)
			}
			if len(li) <40 {
				break
			}
			page++
		}
		begin = end
		if  isOut {
			break
		}
	}
	self.Info.Update = begin.Unix()
	return nil
	//return openSiteDB(siteDB,self.Info.SaveToDB)

}
func (self *Pdd) getOrder(begin,end time.Time,page int)interface{}{
	u := &url.Values{}
	u.Add("type","pdd.ddk.order.list.increment.get")
	u.Add("page_size","40")
	u.Add("return_count","false")
	u.Add("start_update_time",fmt.Sprintf("%d",begin.Unix()))
	u.Add("end_update_time",fmt.Sprintf("%d",end.Unix()))
	u.Add("page",fmt.Sprintf("%d",page))
	return self.ClientHttp(u)
}
