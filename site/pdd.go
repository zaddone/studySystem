package main
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
	"github.com/boltdb/bolt"
	//"net/http"
	"regexp"
)
var (
	PddUrl = "https://gw-api.pinduoduo.com/api/router"
	PddOrderDB *bolt.DB
	//PddErrNum int = 0
	//pdd.ddk.theme.goods.search
	//pdd.ddk.goods.search
	pddReg = regexp.MustCompile(`goods_id=(\d+)`);
	pddOrderReg = regexp.MustCompile(`\d{6}-\d{15}`)
)
type Pdd struct{
	Info *ShoppingInfo
	PddPid []string
	OrderDB *bolt.DB
}
func NewPdd(sh *ShoppingInfo) (p *Pdd) {
	p = &Pdd{Info:sh}
	var err error
	p.OrderDB,err = bolt.Open("pddOrder",0600,nil)
	if err != nil {
		panic(err)
	}
	return
}

func (self *Pdd)addSign(u *url.Values){
	u.Add("client_id",self.Info.Client_id)
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
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
	fmt.Println(sign)
	u.Add("sign",fmt.Sprintf("%X", md5.Sum([]byte(sign))))
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
	//u.Add("generate_we_app","true")
	//if multi{
	u.Add("multi_group","true")
	//}
	return self.ClientHttp(u)
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

func (self *Pdd) SearchGoods(words ...string)interface{}{
	db :=  self.searchGoods(words...)
	if db == nil {
		return nil
	}
	res := db.(map[string]interface{})["goods_search_response"]
	if res == nil {
		return nil
	}
	return res.(map[string]interface{})["goods_list"].([]interface{})
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
	return res.(map[string]interface{})["goods_details"]
	//db_.goods_detail_response.goods_details
}
func (self *Pdd)OrderSearch(keys ...string)interface{}{
	//pdd.ddk.order.detail.get
	orderid := keys[0]
	userid := keys[1]
	err := self.orderGet(orderid,userid)
	if err != nil {
		return nil
	}
	u := &url.Values{}
	u.Add("type","pdd.ddk.order.detail.get")
	u.Add("order_sn",orderid)
	db:= self.ClientHttp(u)
	if db == nil {
		return nil
	}
	err = self.orderSave(orderid,userid,db)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(db)
	return db
	//return nil
}
func (self *Pdd) orderGet (orderid,userid string) error {

	//oid := []byte(orderid)
	return self.OrderDB.View(func(t *bolt.Tx)error{
		b := t.Bucket(dbId)
		if b == nil{
			//panic(err)
			return nil
		}
		v := b.Get([]byte(orderid))
		if v == nil {
			return nil
		}
		var db map[string]interface{}
		err := json.Unmarshal(v,&db)
		if err != nil {
			return err
		}
		uid := db["userid"]
		if uid!=nil && uid.(string)!=userid {
			return io.EOF
		}
		return nil
	})
}
func (self *Pdd) orderSave (orderid,userid string,db interface{}) error {
	response := db.(map[string]interface{})["order_detail_response"]
	if response == nil {
		return nil
	}
	data := response.(map[string]interface{})
	//fmt.Println(db)
	data["userid"] = userid
	return self.OrderDB.Update(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(dbId)
		if err != nil {
			return err
		}
		d_,err := json.Marshal(data)
		if err != nil{
			return err
		}
		return b.Put([]byte(orderid),d_)
	})
	//order_status_desc
	//return nil
}
func (self *Pdd)OutUrl(db interface{}) string {
	res := db.(map[string]interface{})["goods_promotion_url_generate_response"]
	if res == nil {
		return ""
	}
	res_ := res.(map[string]interface{})["goods_promotion_url_list"]
	if res == nil {
		return ""
	}
	res__ := res_.([]interface{})
	if len(res__)== 0 {
		return ""
	}
	res___ := res__[0].(map[string]interface{})
	if res___ == nil {
		return ""
	}
	return res___["short_url"].(string)
}
func(self *Pdd)GetInfo()*ShoppingInfo {
	return self.Info
}

func (self *Pdd) ProductSearch(words ...string)(result []interface{}){
	return self.searchGoods(words...).([]interface{})
}
