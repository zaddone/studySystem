package main
import(
	"fmt"
	"sort"
	"time"
	"crypto/md5"
	"io"
	//"sync"
	"io/ioutil"
	"encoding/json"
	//"strings"
	"bytes"
	//"strconv"
	"regexp"
	"github.com/zaddone/studySystem/request"
	"net/url"
	"github.com/boltdb/bolt"
	"encoding/binary"
)
var (
	JdUrl = "https://router.jd.com/api"
	JdUrl_ = "https://api.jd.com/routerjson"
	//PddErrNum int = 0
	//pdd.ddk.theme.goods.search
	//pdd.ddk.goods.search
	//JdToken = "0619e9dd75e448dea0ab1b0449de3d89wu5z"
	JdToken = "8fb30ead08284c52a879444d6a47c8bdywqw"
	//JdOrderDB *bolt.DB
	//dbTime = []byte("time")
	dbId = []byte("order")
	//dbUser = []byte("user")
	dbLast = []byte("last")
	dbPhone = []byte("Phone")
	timeFormat = "2006-01-02 15:04:05"
	orderTimeFormat = "2006010215"
	//week = []string{""}
	jdReg = regexp.MustCompile(`\/(\d+)\.html`)
	jdReg_ = regexp.MustCompile(`sku=(\d+)`)
	jdOrderReg = regexp.MustCompile(`\d{12,}`)

)

func NewJd(sh *ShoppingInfo) (p *Jd){
	p = &Jd{Info:sh}
	var err error
	p.OrderDB,err = bolt.Open("jdorderDB",0600,nil)
	if err != nil {
		panic(err)
	}
	return
}
type Jd struct{
	Info *ShoppingInfo
	Pid string
	OrderDB *bolt.DB
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
func (self *Jd)HandOrderDB(db interface{},hand func(map[string]interface{}))error{

	if db == nil {
		return fmt.Errorf("db == nil")
	}
	var db_ map[string]interface{}
	err := json.Unmarshal([]byte(db.(map[string]interface{})["jd_union_open_order_query_response"].(map[string]interface{})["result"].(string)),&db_)
	if err != nil {
		return err
	}
	if db_["code"].(float64) != 200 {
		return fmt.Errorf("%v",db)
	}
	datalist := db_["data"]
	if datalist == nil {
		return nil
		//return fmt.Errorf("data = nil")
	}
	for _,l := range datalist.([]interface{}){
		hand(l.(map[string]interface{}))
	}
	return nil

}
func (self *Jd)orderDownAll() (err error){
	t_:=self.orderLast()
	//fmt.Println(t_)
	t,err := self.OrderDB.Begin(true)
	if err != nil {
		return
	}
	defer t.Commit()
	//tb ,err := t.CreateBucketIfNotExists(dbTime)
	//if err != nil {
	//	return
	//}
	ob ,err := t.CreateBucketIfNotExists(dbId)
	if err != nil {
		return
	}
	ol ,err := t.CreateBucketIfNotExists(dbLast)
	if err != nil {
		return
	}
	for{
		//db := self.orderDown(t_)
		fmt.Println(t_)
		err = self.HandOrderDB(self.orderDown(t_),func(v map[string]interface{}){
			//fmt.Println(v)
			orderId :=[]byte(fmt.Sprintf("%.0f",v["orderId"].(float64)))
			v_, err:= json.Marshal(v)
			if err != nil {
				panic(err)
			}
			ob.Put(orderId,v_)
		})
		if err != nil {
			return err
		}

		t_ = t_.Add(time.Hour)
		if time.Now().Unix() < t_.Unix(){
			break
		}
		last_ := make([]byte,8)
		binary.BigEndian.PutUint64(last_,uint64(t_.Unix()))
		ol.Put(dbLast,last_)
	}
	return

}
func (self *Jd)orderDown(t time.Time) interface{} {
	u := &url.Values{}
	u.Add("method","jd.union.open.order.query")
	u.Add("v","1.0")
	u.Add("access_token",JdToken)
	query := map[string]interface{}{
		//"orderId":keys[0],
		"orderReq":map[string]interface{}{
			"pageIndex":1,
			"pageSize":500,
			"type":1,
			"time":t.Format(orderTimeFormat),
		},
		//"bin":"zaddone",

	}
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	u.Add("param_json",string(body))
	return self.ClientHttp(JdUrl,u)
}
func (self *Jd)orderLast() (t_ time.Time) {
	//t_ = nil
	isO:= false
	err := self.OrderDB.View(func(t *bolt.Tx)(err error){
		b := t.Bucket(dbLast)
		if b == nil {
			return
		}
		v := b.Get(dbLast)
		if v == nil {
			return
		}
		t_ = time.Unix(int64(binary.BigEndian.Uint64(v)),0)
		isO = true
		//fmt.Println(t_)
		return
	})
	if err != nil {
		panic(err)
	}
	if !isO {
		t_,err = time.Parse(orderTimeFormat,"2020010101")
		if err != nil {
			panic(err)
		}
	}
	return

}
func (self *Jd)orderGet(orderid,userid string,hand func(interface{}))error{

	key := []byte(orderid)
	//fmt.Println(orderid)
	//u:= []byte(userid)
	var db map[string]interface{}
	err :=  self.OrderDB.View(func(t *bolt.Tx)error{
		b := t.Bucket(dbId)
		if b == nil {
			return io.EOF
		}
		v := b.Get(key)
		if v == nil {
			return io.EOF
		}
		err := json.Unmarshal(v,&db)
		if err != nil {
			//return err
			panic(err)
		}
		fmt.Println(db)
		uid := db["userid"]
		if uid != nil &&  uid.(string) != userid {
			return io.EOF
		}
		//db["userid"] = uid
		return nil
	})
	if err != nil {
		return nil
	}
	return self.HandOrderDB(self.orderDown(time.Unix(int64(db["orderTime"].(float64)/1000),0)),func(v map[string]interface{}){
		if !bytes.Equal(key,[]byte(fmt.Sprintf("%.0f",v["orderId"].(float64)))){
			return
		}
		db = v
		db["userid"] = userid
		t,err := self.OrderDB.Begin(true)
		if err != nil {
			return
		}
		v_, err:= json.Marshal(v)
		if err != nil {
			panic(err)
		}
		t.Bucket(dbId).Put(key,v_)
		t.Commit()
		hand(db)
	})
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
	fmt.Println(sign)
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
	u.Add("access_token",JdToken)
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	//u.Add("360buy_param_json",fmt.Sprintf("{\"goodsReqDTO\":{\"keyword\":\"%s\"}}",words[0]))
	u.Add("param_json",string(body))

	//u.Add("custom_parameters",words[1])
	return self.ClientHttp(JdUrl,u)
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
	u.Add("access_token",JdToken)
	return self.ClientHttp(JdUrl,u)
	//return nil
}
func (self *Jd) GoodsUrl(words ...string) interface{}{

	u := &url.Values{}
	u.Add("method","jd.kpl.open.promotion.pidurlconvert")
	u.Add("v","2.0")
	u.Add("access_token",JdToken)
	query := map[string]interface{}{
		"webId":"0",
		"positionId":0,
		"materalId":fmt.Sprintf("https://item.jd.com/%s.html",words[0]),
		"kplClick":1,
		"shortUrl":1,
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
	err := self.orderGet(keys[0],keys[1],func(db interface{}){
		d = db
		//d = string(db.([]byte))
	})
	if err != nil {
		panic(err)
	}
	if d != nil {
		return d
	}
	//fmt.Println("run down")
	ok := make(chan int,1)
	go func(){
		err := self.orderDownAll()
		if err != nil {
			fmt.Println(err)
		}
		ok<-1
	}()
	select{
	case <- time.After(time.Second*4):
		//fmt.Println("time out")
		return
	case <- ok:
		//fmt.Println("ok")
		err = self.orderGet(keys[0],keys[1],func(db interface{}){
			d = db
			//d = string(db.([]byte))
		})
		if err != nil {
			panic(err)
		}
	}
	return
	//u := &url.Values{}
	////jd.kepler.order.getorderdetail
	////jd.kepler.order.getlist
	////jd.union.open.order.row.query
	////jd.union.open.order.query
	//u.Add("method","jd.union.open.order.query")
	//u.Add("v","1.0")
	//u.Add("access_token",JdToken)
	//query := map[string]interface{}{
	//	//"orderId":keys[0],
	//	"orderReq":map[string]interface{}{
	//		"pageIndex":1,
	//		"pageSize":500,
	//		"type":1,
	//		"time":"",
	//	}
	//	//"bin":"zaddone",
	//}
	//body,err := json.Marshal(query)
	//if err != nil {
	//	panic(err)
	//}
	//u.Add("param_json",string(body))
	//return self.ClientHttp(JdUrl,u)

}
func (self *Jd)OutUrl(db interface{}) string {
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
