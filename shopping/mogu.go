package shopping
import(
	"net/url"
	"crypto/md5"
	"sort"
	"encoding/json"
	"github.com/zaddone/studySystem/request"
	//"github.com/zaddone/studySystem/config"
	"strings"
	"strconv"
	"time"
	"fmt"
	"io"
	"net/http"
	"io/ioutil"
	"regexp"
)
var (
	regId = regexp.MustCompile(`itemId=([a-zA-Z0-9]+)`)
)

func NewMogu(sh *ShoppingInfo,siteDB string) (ShoppingInterface){
	m := &Mogu{Info:sh}
	//return m
	if siteDB == "" {
		return m
	}

	//go func (){
	//	for _ = range m.DownChan{
	//		m.OrderDownSelf(func(db interface{}){
	//			err := OrderUpdate(db.(map[string]interface{})["order_id"].(string),db)
	//			if err != nil {
	//				fmt.Println(err)
	//			}
	//		})
	//	}
	//}()
	go func(){
		for{
			//fmt.Println(m.Info)
		err := m.ReToken(siteDB)
		if err != nil {
			fmt.Println(err)
			//panic(err)
		}
		long := m.Info.TimeOut - time.Now().Unix() - 7200
		//if long < 3600 {
		//	long = 3600
		//}
		time.Sleep(time.Duration(long)*time.Second)
		}
	}()
	return m
}

type Mogu struct{
	Info *ShoppingInfo
	//Pid string
	//OrderDB *bolt.DB
	//DownChan chan bool
}

func (self *Mogu) ReToken (siteDB string) error {
	u := url.Values{}
	u.Set("app_key",self.Info.Client_id)
	u.Set("app_secret",self.Info.Client_secret)
	u.Set("grant_type","refresh_token")
	u.Set("refresh_token",self.Info.ReToken)
	uri := "https://oauth.mogujie.com/token?"+u.Encode()
	fmt.Println(uri)
	return request.ClientHttp_(
		uri,
		"GET",nil,nil,
		func(body io.Reader,start int)error{
			var res map[string]interface{}
			err := json.NewDecoder(body).Decode(&res)
			if err != nil {
				return err
			}
			fmt.Println(res)
			if res["statusCode"] != "0000000" {
				return fmt.Errorf(res["errorMsg"].(string))
			}
			self.Info.Token = res["access_token"].(string)
			self.Info.ReToken = res["refresh_token"].(string)
			self.Info.TimeOut = int64(res["access_expires_in"].(float64))
			//fmt.Println(siteDB)
			return OpenSiteDB(siteDB,self.Info.SaveToDB)
		},
	)
}
func (self *Mogu)addSign(u *url.Values){
	u.Add("app_key",self.Info.Client_id)
	u.Add("version","1.0")
	u.Add("format","json")
	u.Add("sign_method","md5")
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
	//fmt.Println(sign)
	u.Add("sign",fmt.Sprintf("%X", md5.Sum([]byte(sign))))
	//fmt.Println(u.Get("sign"))
}

func (self *Mogu) ClientHttp(u *url.Values)( out interface{}){

	self.addSign(u)
	//ht := http.Header{}
	//ht.Add("Content-Type","application/json")
	var err error
	err = request.ClientHttp_(
		"https://openapi.mogujie.com/invoke?"+u.Encode(),
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
func(self *Mogu)GetInfo()*ShoppingInfo{
	return self.Info
}

func(self *Mogu)stuctured(data interface{})(g Goods){
	d_ := data.(map[string]interface{})
	//p
	f,err :=strconv.ParseFloat(strings.Replace(d_["commissionRate"].(string),"%","",-1),64)
	if err != nil {
		panic(err)
	}
	p,err :=strconv.ParseFloat(d_["zkPrice"].(string),64)
	if err != nil {
		panic(err)
	}
	g= Goods{
		Id:d_["itemId"].(string),
		Img:[]string{strings.Replace(d_["pictUrlForH5"].(string),"http:","",-1)},
		Name:d_["title"].(string),
		Tag:d_["shopTitle"].(string),
		Price:p,
		Fprice:fmt.Sprintf("%.2f",f/100*p*Rate),
		Show:d_["extendDesc"].(string),
		Coupon:!strings.EqualFold(d_["dayLeft"].(string), "0"),
	}
	if g.Coupon {
		g.Id =g.Id+"_"+ d_["promid"].(string)
	}
	return g

}
func(self *Mogu)SearchGoods(words ...string)interface{}{
	//xiaodian.cpsdata.promitem.get
	u := &url.Values{}
	u.Set("method","xiaodian.cpsdata.promitem.get")
	u.Set("access_token",self.Info.Token)
	query := map[string]interface{}{
		"keyword":words[0],
	}
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	u.Set("promInfoQuery",string(body))
	db := self.ClientHttp(u)
	if db == nil {
		return nil
	}
	//fmt.Println(db)
	res := db.(map[string]interface{})["result"]
	if res == nil {
		return nil
	}
	data := res.(map[string]interface{})["data"]
	if data == nil {
		return nil
	}
	items:= data.(map[string]interface{})["items"]
	if items == nil {
		return nil
	}
	var li []interface{}
	for _,l := range items.([]interface{}){
		li = append(li,self.stuctured(l))
	}
	return li
}
func (self *Mogu)GoodsAppMini(words ...string)interface{}{
	ids :=strings.Split(words[0],"_")
	body := map[string]interface{}{
		"itemId":ids[0],
		"uid":"1exztg0",
		//"gid":words[1],
		"genWxcode":false,
	}
	if len(ids)>1{
		body["promId"] = ids[1]
	}
	b,err:= json.Marshal(body)
	if err != nil {
		return nil
	}
	u := &url.Values{}
	u.Set("method","xiaodian.cpsdata.wxcode.get")
	u.Set("access_token",self.Info.Token)
	u.Set("wxcodeParam",string(b))
	db := self.ClientHttp(u)
	if db == nil {
		return nil
	}
	//fmt.Println(db)
	res := db.(map[string]interface{})["result"]
	if res == nil {
		return nil
	}
	data := res.(map[string]interface{})["data"]
	if data == nil {
		return nil
	}
	d_:= data.(map[string]interface{})
	return map[string]interface{}{
		"appid":"wxca3957e5474b3670",
		"url":d_["path"].(string)+"&feedback="+words[1],
	}

}

func(self *Mogu)GoodsUrl(words ...string)interface{}{
	//https://www.mogujie.com/cps/open/track
	if words[len(words)-1] =="mini"{
		return self.GoodsAppMini(words[:len(words)-1]...)
	}
	ids :=strings.Split(words[0],"_")
	u := &url.Values{}
	var uri string
	if len(ids)==1 {
		u.Set("uid","1exztg0")
		u.Set("target","https://shop.mogu.com/detail/"+ids[0])
		u.Set("feedback",words[1])
		uri = "https://www.mogujie.com/cps/open/track?"+u.Encode()
	}else if len(ids)==2 {
		u.Set("userid","1exztg0")
		u.Set("itemid",ids[0])
		u.Set("promid",ids[1])
		u.Set("feedback",words[1])
		uri = "https://union.mogujie.com/jump?"+u.Encode()
	}
	//xiaodian.cpsdata.wxcode.get
	return uri
	//u := &url.Values{}
	//u.Set("method","xiaodian.cpsdata.url.shorten")
	//u.Set("access_token",self.Info.Token)
	//u_ := &url.Values{}
	//u_.Set("userid","1exztg0")
	//u_.Set("userid","1exztg0")
	//u.Set("url","https://union.mogujie.com/jump?"+u_.Encode())
	//query := map[string]interface{}{
	//	"keyword":words[0],
	//}
	//return nil
}

func (self *Mogu) GoodsRequest(uri string) interface{} {

	//var err error
	//fmt.Println(uri)
	//u,err := url.Parse(uri)
	//if err != nil {
	//	panic(err)
	//}
	//h:= config.Conf.Header
	//h.Set("Host",u.Host)
	var id string
	err := request.ClientHttp__(uri,"GET",nil,nil,func(body io.Reader,res *http.Response)error{
		fmt.Println(res.Request.URL.String())
		str := regId.FindAllStringSubmatch(res.Request.URL.String(),-1)
		if len(str) >0 {
			id = str[0][1]
			return nil
		}
		//fmt.Println(str,res.Request.URL.String())
		return io.EOF

	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return self.goodsDetail(id)
	//return self.GoodsDetail("http://shop.mogujie.com/detail/"+id)

}
func (self *Mogu) goodsDetail(words string) interface{}{
	fmt.Println("detail",words)
	u := &url.Values{}
	u.Set("method","xiaodian.cpsdata.item.get")
	u.Set("access_token",self.Info.Token)
	u.Set("url",words)
	db := self.ClientHttp(u)
	fmt.Println(db)
	if db == nil {
		return nil
	}
	res := db.(map[string]interface{})["result"]
	if res == nil {
		return nil
	}
	data := res.(map[string]interface{})["data"]
	if data == nil {
		return nil
	}
	return []interface{}{self.stuctured(data)}
}

func (self *Mogu) GoodsDetail(words ...string) interface{} {
	//xiaodian.cpsdata.item.get
	uri := words[0]
	if !strings.Contains(uri,"http"){
		return self.goodsDetail(uri)
	}
	if !strings.Contains(uri,"detail"){
		return self.GoodsRequest(uri)
	}
	return self.goodsDetail(uri)

}
func(self *Mogu)OrderSearch(...string)interface{}{
	return nil
}
func(self *Mogu)OutUrl(d interface{}) string{
	return d.(string)
}
func(self *Mogu)OrderMsg(interface{}) string{
	return ""
}
func(self *Mogu)ProductSearch(...string)[]interface{}{
	return nil
}
func (self *Mogu)getOrder(begin,end time.Time,page int) interface{} {
	//xiaodian.cpsdata.order.list.get
	u := &url.Values{}
	u.Set("method","xiaodian.cpsdata.order.list.get")
	u.Set("access_token",self.Info.Token)
	//end
	query := map[string]interface{}{
		"start":begin.Format(payTimeFormat),
		"end":end.Format(payTimeFormat),
		"pagesize":20,
		"page":page,
	}
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	u.Set("orderInfoQuery",string(body))
	return self.ClientHttp(u)
	//if db == nil {
	//	return nil
	//}
	//res := db.(map[string]interface{})["result"]
	//if res == nil {
	//	return nil
	//}
	//data := res.(map[string]interface{})["data"]
	//if data == nil {
	//	return nil
	//}
	//order := data.(map[string]interface{})["orders"]
	//if order == nil {
	//	return nil
	//}
	//return order

}
func(self *Mogu)OrderDownSelf(hand func(interface{}))error{
	return self.OrderDown(hand)
}
func(self *Mogu)OrderDown(hand func(interface{}))error{

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
		end := begin.AddDate(0,0,5)
		page := 1
		for {
			db := self.getOrder(begin,end,page)
			if db == nil {
				//break
				time.Sleep(1*time.Second)
				continue
			}
			//fmt.Println(db)
			res := db.(map[string]interface{})["result"]
			if res == nil {
				break
			}
			data := res.(map[string]interface{})["data"]
			if data == nil {
				break
			}
			order:= data.(map[string]interface{})["orders"]
			if order == nil {
				break
			}
			li := order.([]interface{})
			for _,l :=range li{
				l_ := l.(map[string]interface{})
				l_["order_id"] = fmt.Sprintf("%.0f",l_["orderNo"].(float64))
				//l_["userid"] = l_["feedback"].(string)
				var id []string
				var name []string
				for _,p := range l_["products"].([]interface{}){
					p_ := p.(map[string]interface{})
					id = append(id,p_["productNo"].(string))
					name = append(name,p_["name"].(string))
				}
				l_["goodsid"] = strings.Join(id,",")
				l_["goodsName"] = strings.Join(name,",")
				//l_["fee"] = l_["expense"].(float64)
				fee,err := strconv.ParseFloat(l_["expense"].(string),64)
				if err == nil {
					l_["fee"] = fee
					//panic(err)
				}else{
					fmt.Println(err)
				}
				l_["site"] = self.Info.Py
				pay := l_["chargeDate"].(float64)
				//l_["order_id"] = fmt.Sprintf("%.0f",l_["chargeDate"].(float64))
				//pay,err := time.Parse(payTimeFormat,l_["chargeDate"].(string))
				if pay != 0 {
					l_["endTime"] = int64(l_["updateTime"].(float64))
					l_["PayTime"] = int64(pay)
					//panic(err)
				//}else{
				//	fmt.Println(err)
				}
				//l_[""]
				hand(l_)
			}
			if len(li) < 20 {
				break
			}
			page++
		}
		Now := time.Now().Unix()
		if end.Unix()> Now {
			self.Info.Update = Now
			break
		}
		begin = end.AddDate(0,0,1)
		time.Sleep(1*time.Second)
	}
	return nil

}
func (self *Mogu) Test()interface{}{
	return nil
}
