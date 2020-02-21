package shopping
import(
	"fmt"
	"time"
	"io/ioutil"
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
func Hmac(key, data string) string {
	hmac := hmac.New(md5.New, []byte(key))
	hmac.Write([]byte(data))
	return hex.EncodeToString(hmac.Sum([]byte("")))
}
type Vip struct{
	Info *ShoppingInfo
	VipPid []string
}
func (self *Vip)addSign(u *url.Values,body string){
	u.Add("appKey",self.Info.Client_id)
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	u.Add("format","json")
	var li []string
	for k,_ := range *u {
		li = append(li,k)
	}
	sort.Strings(li)
	sign := self.Info.Client_secret
	for _,k :=range li {
		sign+=k+u.Get(k)
	}
	//sign+=self.Info.Client_secret
	//sign = Hmac(self.Info.Client_secret,sign+body)
	//fmt.Println(sign)
	u.Add("sign",Hmac(self.Info.Client_secret,sign+body))
}
func (self *Vip) ClientHttp(u *url.Values,body string)( out interface{}){

	self.addSign(u,body)
	//ht := http.Header{}
	//ht.Add("Content-Type","application/json")
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
func (self *Vip) SearchGoods(words ...string)interface{}{
	u := &url.Values{}
	u.Add("service","com.vip.adp.api.open.service.UnionGoodsService")
	u.Add("method","query")
	u.Add("version","1.0.0")
	body,err :=json.Marshal(
		map[string]interface{}{
			"keyword":words[0],
			"page":"1",
			"requestId":words[1],
		})
	if err != nil {
		panic(err)
	}
	//body_ := string(body)
	//self.addSign(u,body_)
	return self.ClientHttp(u,string(body))

}

func (self *Vip) GoodsUrl(words ...string) interface{}{
	return nil
}

func (self *Vip) GoodsDetail(words ...string)interface{}{
	return nil
}
func (self *Vip)OrderSearch(keys ...string)interface{}{
	return nil
}
func (self *Vip)OutUrl(db interface{}) string {
	return ""
}
func(self *Vip)GetInfo()*ShoppingInfo {
	return self.Info
}
