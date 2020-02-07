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
	//"strconv"
	"github.com/zaddone/studySystem/request"
	"net/url"
)
var (
	JdUrl = "https://router.jd.com/api"
	//PddErrNum int = 0
	//pdd.ddk.theme.goods.search
	//pdd.ddk.goods.search
	JdToken = "0619e9dd75e448dea0ab1b0449de3d89wu5z"
)
type Jd struct{
	Info *ShoppingInfo
	Pid string
}

func (self *Jd)addSign(u *url.Values){
	u.Add("app_key",self.Info.Client_id)
	//u.Add("access_token","8fb30ead08284c52a879444d6a47c8bdywqw")
	u.Add("format","json")
	now := time.Now().Add(-(time.Minute*2))
	u.Add("sign_method","md5")
	//time.Now().Format("2006-01-02 15:04:05")
	//timestamp
	u.Add("timestamp",now.Format("2006-01-02 15:04:05"))
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

func (self *Jd) ClientHttp(u *url.Values)( out interface{}){

	self.addSign(u)
	//ht := http.Header{}
	//ht.Add("Content-Type","application/json")
	var err error
	err = request.ClientHttp_(
		JdUrl+"?"+u.Encode(),
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
	return self.ClientHttp(u)
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
	return self.ClientHttp(u)
	//return nil
}
func (self *Jd) GoodsUrl(words ...string) interface{}{

	u := &url.Values{}
	//jd.kpl.open.promotion.pidurlconvert
	u.Add("method","jd.kpl.open.promotion.pidurlconvert")
	u.Add("v","2.0")
	u.Add("access_token",JdToken)
	query := map[string]interface{}{
		"webId":"0",
		"positionId":0,
		"materalId":fmt.Sprintf("https://item.jd.com/%s.html",words[0]),
		"kplClick":1,
	}
	body,err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	u.Add("param_json",string(body))

	return self.ClientHttp(u)
}
