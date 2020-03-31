package shopping
import(
	"fmt"
	"sort"
	"crypto/md5"
	"time"
	"io"
	"io/ioutil"
	"encoding/json"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/alimama"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strconv"
	//"github.com/boltdb/bolt"
	//"strconv"
	"regexp"
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)
var (
	taobaoid = regexp.MustCompile(`[\?|\&]id=(\d+)`)
	getTaobaoUrl=regexp.MustCompile(`var url = '(\S+)';`)
	getPageInfo = regexp.MustCompile(`var extraData = (\{.+\})`)
	checkGoodsID = regexp.MustCompile(`\D`)
	//Pid = "109998500026"
)
func DecodeGBK(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}
type Taobao struct{
	Info *ShoppingInfo
	Pid string
	//OrderDB *bolt.DB
	Url string
}
func NewTaobao(sh *ShoppingInfo)(ShoppingInterface) {
	t := &Taobao{Info:sh}
	t.Pid = "109998500026"
	t.Url = "https://eco.taobao.com/router/rest"
	//if !o{
	//	return t
	//}
	//var err error
	//t.OrderDB,err = bolt.Open("taobaoOrder",0600,nil)
	//if err != nil {
	//	panic(err)
	//}
	return t
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
	u := &url.Values{}
	u.Add("q",words[0])
	u.Add("adzone_id",self.Pid)
	u.Add("platform","2")
	u.Add("is_tmall","false")
	u.Add("method","taobao.tbk.dg.material.optional")
	db := self.ClientHttp(u)
	if db == nil {
		return nil
	}
	res_ := db.(map[string]interface{})["tbk_dg_material_optional_response"]
	if res_ == nil {
		return nil
	}

	var li []interface{}
	for _,d := range res_.(map[string]interface{})["result_list"].(map[string]interface{})["map_data"].([]interface{}){
		li = append(li,self.stuctured(d))
	}
	return li
	//return res_.(map[string]interface{})["result_list"].(map[string]interface{})["map_data"]

}

func (self *Taobao) stuctured(data interface{}) (g Goods){
	d_ := data.(map[string]interface{})
	p,err := strconv.ParseFloat(d_["zk_final_price"].(string),64)
	if err != nil {
		panic(err)
	}
	r,err := strconv.ParseFloat(d_["commission_rate"].(string),64)
	if err != nil {
		panic(err)
	}
	g = Goods{
		Id:fmt.Sprintf("%.0f",d_["item_id"].(float64)),
		Img:[]string{d_["pict_url"].(string)},
		Name:d_["title"].(string),
		Tag:d_["shop_title"].(string),
		Price:p,
		Fprice:r/10000,
		//Ext:"https:"+d_["coupon_share_url"].(string),
		Coupon:len(d_["coupon_id"].(string))>0,
		Show:d_["item_description"].(string),
		//Coupon:

	}
	if !g.Coupon {
		g.Ext = "https:"+d_["url"].(string)
	}else{
		g.Ext = "https:"+d_["coupon_share_url"].(string)
	}
	if d_["small_images"] != nil {
	for _,m := range d_["small_images"].(map[string]interface{})["string"].([]interface{}){
		g.Img = append(g.Img,m.(string))
	}
	}
	return

}
func (self *Taobao) ProductSearch(words ...string)(result []interface{}){
	//taobao.tbk.dg.material.optional
	u := &url.Values{}
	u.Add("q",words[0])
	u.Add("type","p")
	//var result []interface{}
	err:= request.ClientHttp_("https://list.tmall.com/search_product.htm?"+u.Encode(),"GET",nil,nil,func(body io.Reader,st int)error{
		doc,err := goquery.NewDocumentFromReader(body)
		//db,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		//fmt.Println(string(db))
		I := 0
		doc.Find(".product").EachWithBreak(func(i int,s *goquery.Selection)bool{
			title,err := DecodeGBK([]byte(s.Find(".productTitle").Text()))
			if err != nil {
				panic(err)
			}
			uri,_ := s.Find(".productTitle a").First().Attr("href")
			ids := taobaoid.FindStringSubmatch(uri)
			//fmt.Println(title,ids[1])
			db := self.SearchGoods(string(title))
			if db == nil{
				return true
			}
			id_ := ids[1]
			for _,v := range self.getResList(db){
				if id_ == fmt.Sprintf("%.0f",v.(map[string]interface{})["item_id"].(float64)){
					I++
					result = append(result,v)
					break
				}
			}

			return I<4

		})
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return result

}

func (self *Taobao)GoodsAppMini(words ...string)interface{}{
	return nil
}
func (self *Taobao) GoodsUrl(words ...string) interface{}{
	//taobao.tbk.spread.get
	//taobao.tbk.tpwd.create
	u := &url.Values{}
	u.Add("method","taobao.tbk.tpwd.create")
	u.Add("text",words[2])
	u.Add("url",words[0])
	u.Add("ext",fmt.Sprintf("{session:\"%s\"}",words[1]))
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

func (self *Taobao) goodsForId(id string)interface{}{
	goodinfo := self.goodsInfo(id)
	if goodinfo == nil {
		return nil
	}
	res := goodinfo.(map[string]interface{})["tbk_item_info_get_response"]
	if res == nil {
		return nil
	}
	li := res.(map[string]interface{})["results"].(map[string]interface{})["n_tbk_item"].([]interface{})
	if len(li) == 0 {
		return nil
	}
	data := li[0].(map[string]interface{})
	goods := Goods{
		Img:[]string{data["pict_url"].(string)},
		Name:data["title"].(string),
	}
	db := self.SearchGoods(goods.Name)
	if db == nil {
		return []interface{}{goods}
	}
	for _,v := range db.([]interface{}) {
		if id == v.(Goods).Id{
			return []interface{}{v}
		}
	}
	return []interface{}{goods}


}

func (self *Taobao) GoodsDetail(words ...string)interface{}{
	//taobao.tbk.item.click.extract
	uri := words[0]
	c := checkGoodsID.FindAllString(uri,-1)
	if len(c) == 0 {
		return self.goodsForId(uri)
	}

	ids := taobaoid.FindStringSubmatch(uri)
	var data map[string]interface{}
	//pageinfo := ""
	if len(ids) >0 {
		return self.goodsForId(ids[1])
	}
	err:= request.ClientHttp_(uri,"GET",nil,nil,func(body io.Reader,st int)error{
		db,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		if st != 200 {
			return io.EOF
		}
		uri = string(getTaobaoUrl.Find(db))
		page:=getPageInfo.FindAllSubmatch(db,-1)
		//fmt.Println(string(page))
		//for i_,p_:= range page{
		//	for i,p := range p_{
		//		fmt.Println(i_,i,string(p))
		//	}
		//}
		if len(page)==1 && len(page[0])==2{
			return json.Unmarshal(page[0][1],&data)
		}
		//fmt.Println(string(db))
		//fmt.Println(st,uri)
		return io.EOF
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if data == nil {
		return nil
	}
	ids = taobaoid.FindStringSubmatch(uri)
	if len(ids) == 0 {
		return nil
	}
	id := ids[1]

	//return self.SearchGoods(id)
	//fmt.Println(ids,id)
	//goodinfo := self.goodsInfo(id)
	//if goodinfo == nil {
	//	return nil
	//}
	////fmt.Println(goodinfo)
	//res := goodinfo.(map[string]interface{})["tbk_item_info_get_response"]
	//if res == nil {
	//	return nil
	//}
	//li := res.(map[string]interface{})["results"].(map[string]interface{})["n_tbk_item"].([]interface{})
	//if len(li) == 0 {
	//	return nil
	//}
	//db := li[0].(map[string]interface{})
	p,err:= strconv.ParseFloat(data["priceL"].(string),64)
	if err != nil {
		return nil
	}
	goods :=Goods{
		Price:p,
		Img:[]string{data["pic"].(string)},
		Name:data["title"].(string),
	}
	db := self.SearchGoods(data["title"].(string))
	if db == nil {
		return []interface{}{goods}
	}

	for _,v := range db.([]interface{}) {
		if id == v.(Goods).Id{
			return []interface{}{v}
		}
	}
	return []interface{}{goods}

}
func (self *Taobao) getResList(db interface{}) []interface{} {
	res_ := db.(map[string]interface{})["tbk_dg_material_optional_response"]
	if res_ == nil {
		return nil
	}
	return res_.(map[string]interface{})["result_list"].(map[string]interface{})["map_data"].([]interface{})
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

func (self *Taobao)OrderDown(hand func(interface{}))error{
	//fmt.Println("taobao")
	if self.Info.Update !=0 {
		alimama.Begin = time.Unix(self.Info.Update,0)
	}
	alimama.HandOrder = hand
	defer func(){
		self.Info.Update = alimama.Begin.Unix()
	}()
	return alimama.Run()
	//return  nil
	//return nil
}

