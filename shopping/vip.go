package shopping
import(
	"fmt"
	"time"
	"io/ioutil"
	"strconv"
	"net/url"
	"sort"
	"encoding/json"
	"strings"
	"io"
	"github.com/zaddone/studySystem/request"
	"crypto/md5"
	"crypto/hmac"
	"encoding/hex"
)
var (
	VipUrl = "https://gw.vipapis.com"
	//VipUrl = "http://sandbox.vipapis.com"
	//pdd.ddk.theme.goods.search
	//pdd.ddk.goods.search
)
func NewVip(sh *ShoppingInfo,r string) ShoppingInterface{
	t := &Vip{Info:sh}
	t.Url = "https://gw.vipapis.com"
	return t
}
func Hmac(key, data string) string {
	hmac := hmac.New(md5.New, []byte(key))
	hmac.Write([]byte(data))
	return strings.ToUpper(hex.EncodeToString(hmac.Sum([]byte(""))))
}
type Vip struct{
	Info *ShoppingInfo
	VipPid []string
	Url string
}
func (self *Vip)addSign(u *url.Values,body string){
	u.Add("appKey",self.Info.Client_id)
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	u.Add("format","json")
	u.Add("version","1.0.0")
	var li []string
	for k,_ := range *u {
		li = append(li,k)
	}
	sort.Strings(li)
	//sign := self.Info.Client_secret
	var sign string
	for _,k :=range li {
		sign+=k+u.Get(k)
	}
	//sign+=self.Info.Client_secret
	//sign = Hmac(self.Info.Client_secret,sign+body)
	//fmt.Println(sign,body)
	u.Add("sign",Hmac(self.Info.Client_secret,sign+body))
}
func (self *Vip) ClientHttp(u *url.Values,body string)( out interface{}){

	self.addSign(u,body)
	//ht := http.Header{}
	//ht.Add("Content-Type","application/json")
	//fmt.Println(u)
	var err error
	err = request.ClientHttp_(
		VipUrl+"?"+u.Encode(),
		"POST",
		strings.NewReader(body),
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
//com.vip.adp.api.open.service.UnionGoodsService 1.0.0
//query
func (self *Vip)stuctured(d_ interface{})(g Goods){
	d:=d_.(map[string]interface{})
	p,err := strconv.ParseFloat(d["vipPrice"].(string),64)
	if err != nil {
		panic(err)
	}
	f,err := strconv.ParseFloat(d["commission"].(string),64)
	if err != nil {
		panic(err)
	}
	img := []string{d["goodsMainPicture"].(string)}
	if d["goodsCarouselPictures"] != nil {
		img = append(img,d["goodsCarouselPictures"].([]string)...)
	}
	return Goods{
		Id:d["goodsId"].(string),
		Name:d["goodsName"].(string),
		Img:img,
		Price:p,
		Fprice:fmt.Sprintf("%.2f",f*Rate),
		Tag:d["brandName"].(string),
	}
}
func (self *Vip) SearchGoods(words ...string)interface{}{
	u := &url.Values{}
	u.Add("service","com.vip.adp.api.open.service.UnionGoodsService")
	u.Add("method","query")
	if len(words) <2{
		words = append(words,"zaddone")
	}
	body,err :=json.Marshal(
		map[string]interface{}{
			"request":map[string]interface{}{
				"keyword":words[0],
				"page":"1",
				"requestId":words[1],
			},
		})
	if err != nil {
		panic(err)
	}
	db := self.ClientHttp(u,string(body))
	if db == nil {
		return nil
	}
	req := db.(map[string]interface{})["result"]
	if req == nil {
		return nil
	}
	goodslist := req.(map[string]interface{})["goodsInfoList"]
	if goodslist == nil {
		return nil
	}
	var li []interface{}
	for _,d := range goodslist.([]interface{}){
		li = append(li,self.stuctured(d))
	}
	return li

}

func (self *Vip) GoodsUrl(words ...string) interface{}{
	u := &url.Values{}
	u.Add("service","com.vip.adp.api.open.service.UnionUrlService")
	u.Add("method","genByGoodsId")
	body,err :=json.Marshal(
		map[string]interface{}{
			"goodsIdList":[]string{words[0]},
			"requestId":words[1],
		})
	if err != nil {
		panic(err)
	}
	db := self.ClientHttp(u,string(body))
	//fmt.Println(db)
	if db == nil {
		return nil
	}
	req := db.(map[string]interface{})["result"]
	if req == nil {
		return nil
	}
	li := req.(map[string]interface{})["urlInfoList"]
	if li == nil {
		return nil
	}
	return li.([]interface{})[0]
	//return li.([]interface{})[0].(map[string]interface{})["longUrl"]

	//return nil
}

func (self *Vip) GoodsDetail(words ...string)interface{}{
	return nil
}
func (self *Vip)OrderSearch(keys ...string)interface{}{
	return nil
}
func (self *Vip)OutUrl(db interface{}) string {
	if db == nil {
		return ""
	}
	return db.(map[string]interface{})["longUrl"].(string)
	//return db.(string)
}
func(self *Vip)GetInfo()*ShoppingInfo {
	return self.Info
}
func (self *Vip)OrderMsg(interface{}) string{
	return ""
}
func (self *Vip)ProductSearch(...string)[]interface{}{
	return nil
}
func (self *Vip)GoodsAppMini(words ...string)interface{}{
	db := self.GoodsUrl(words...)
	if db == nil {
		return nil
	}
	return map[string]interface{}{
		"appid":"wxe9714e742209d35f",
		"url":db.(map[string]interface{})["vipWxUrl"].(string),
	}
}
func (self *Vip)OrderDown(hand func(interface{}))error{
	return nil
}
func (self *Vip)OrderDownSelf(hand func(interface{}))error{
	return nil
}
