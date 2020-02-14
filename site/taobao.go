package main
import(
	"fmt"
	"sort"
	"crypto/md5"
	"time"
	"io"
	"io/ioutil"
	"encoding/json"
	"github.com/zaddone/studySystem/request"
	"net/url"
	"github.com/boltdb/bolt"
	//"strconv"
	"regexp"
)
var (
	taobaoid = regexp.MustCompile(`[\?|\&]id=(\d+)`)
	getTaobaoUrl=regexp.MustCompile(`var url = '(\S+)';`)
	//Pid = "109998500026"
)
type Taobao struct{
	Info *ShoppingInfo
	Pid string
	OrderDB *bolt.DB
	Url string
}
func NewTaobao(sh *ShoppingInfo)(t *Taobao) {
	t = &Taobao{Info:sh}
	t.Pid = "109998500026"
	t.Url = "https://eco.taobao.com/router/rest"
	var err error
	t.OrderDB,err = bolt.Open("taobaoOrder",0600,nil)
	if err != nil {
		panic(err)
	}
	return
}
func (self *Taobao)addSign(u *url.Values){
	u.Add("format","json")
	u.Add("v","2.0")
	u.Add("sign_method","md5")
	u.Add("app_key",self.Info.Client_id)
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
func (self *Taobao) ClientHttp(u *url.Values)( out interface{}){

	self.addSign(u)
	//ht := http.Header{}
	//ht.Add("Content-Type","application/json")
	var err error
	err = request.ClientHttp_(
		self.Url+"?"+u.Encode(),
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
func (self *Taobao) SearchGoods(words ...string)interface{}{
	//taobao.tbk.dg.material.optional
	u := &url.Values{}
	u.Add("adzone_id",self.Pid)
	u.Add("q",words[0])
	u.Add("platform","2")
	u.Add("method","taobao.tbk.dg.material.optional")
	return self.ClientHttp(u)
	//fmt.Println(db)
	//return db
	//return nil
}
func (self *Taobao) GoodsUrl(words ...string) interface{}{
	//taobao.tbk.spread.get
	//taobao.tbk.tpwd.create
	u := &url.Values{}
	u.Add("method","taobao.tbk.tpwd.create")
	u.Add("text","来自米果推荐 zaddone.com "+words[1])
	u.Add("url",words[0])
	return self.ClientHttp(u)
	//req := map[string]interface{}{"url":}
	//u.Add("requests",)
	//return nil
}
func (self *Taobao) goodsUrlToId(uri string) interface{}{
	//taobao.tbk.item.click.extract
	u := &url.Values{}
	u.Add("method","taobao.tbk.item.click.extract")
	u.Add("click_url",uri)
	return self.ClientHttp(u)
}
func (self *Taobao) goodsInfo(id string) interface{} {
	//taobao.tbk.item.info.get
	u := &url.Values{}
	u.Add("method","taobao.tbk.item.info.get")
	u.Add("num_iids",id)
	return self.ClientHttp(u)
}
func (self *Taobao) GoodsDetail(words ...string)interface{}{
	//taobao.tbk.item.click.extract
	uri := words[0]
	ids := taobaoid.FindStringSubmatch(uri)
	if len(ids) == 0 {
		err:= request.ClientHttp_(uri,"GET",nil,nil,func(body io.Reader,st int)error{
			db,err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			if st != 200 {
				return io.EOF
			}

			//fmt.Println(string(db))
			uri = string(getTaobaoUrl.Find(db))
			//fmt.Println(st,uri)
			return nil
		})
		if err != nil {
			fmt.Println(err)
			return nil

		}
		ids = taobaoid.FindStringSubmatch(uri)
		if len(ids) == 0 {
			return nil
		}
		//return nil
	}
	id := ids[1]
	//return self.SearchGoods(id)
	//fmt.Println(ids,id)
	goodinfo := self.goodsInfo(id)
	if goodinfo == nil {
		return nil
	}
	fmt.Println(goodinfo)
	res := goodinfo.(map[string]interface{})["tbk_item_info_get_response"]
	if res == nil {
		return nil
	}
	li := res.(map[string]interface{})["results"].(map[string]interface{})["n_tbk_item"].([]interface{})
	if len(li) == 0 {
		return nil
	}
	//db := li[0].(map[string]interface{})
	db := self.SearchGoods(li[0].(map[string]interface{})["title"].(string))
	if db == nil {
		return nil
	}
	res_ := db.(map[string]interface{})["tbk_dg_material_optional_response"]
	if res_ == nil {
		return nil
	}
	var li_  []interface{}
	reslist := res_.(map[string]interface{})["result_list"].(map[string]interface{})
	for _,v := range reslist["map_data"].([]interface{}){
		if id == fmt.Sprintf("%.0f",v.(map[string]interface{})["item_id"].(float64)){
			li_ = []interface{}{v}
			break
		}
	}
	reslist["map_data"] = li_
	return db
}
func (self *Taobao) OrderSearch(keys ...string)interface{}{
	return nil
}
func (self *Taobao) GetInfo() *ShoppingInfo {
	return self.Info
}
func (self *Taobao) OutUrl(db interface{}) string {
	return ""
}
func (self *Taobao)OrderMsg(_db interface{}) (str string){
	return ""
}

