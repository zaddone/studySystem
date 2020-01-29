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
	"net/http"
)
var (
	PddUrl = "https://gw-api.pinduoduo.com/api/router"
	PddErrNum int = 0
	//pdd.ddk.theme.goods.search
	//pdd.ddk.goods.search
)
type Pdd struct{
	Info *ShoppingInfo
	PddPid []string
}
func (self *Pdd)addSign(u *url.Values){
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
	ht := http.Header{}
	ht.Add("Content-Type","application/json")
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
	u.Add("client_id",self.Info.Client_id)
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	u.Add("type","pdd.ddk.goods.pid.generate")
	u.Add("number",fmt.Sprintf("%d",n))
	return self.ClientHttp(u)
}
//pdd.ddk.goods.pid.query
func (self *Pdd) pidQuery() interface{}{
	u := &url.Values{}
	u.Add("client_id",self.Info.Client_id)
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	u.Add("type","pdd.ddk.goods.pid.query")
	return self.ClientHttp(u)
}
//pdd.ddk.goods.promotion.url.generate
func (self *Pdd) GoodsUrl(goodsid string) interface{}{
	if len(self.PddPid) == 0 {
		err := self.getPid()
		if err != nil {
			return err
		}
	}

	u := &url.Values{}
	u.Add("client_id",self.Info.Client_id)
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	u.Add("type","pdd.ddk.goods.promotion.url.generate")
	u.Add("goods_id_list",goodsid)
	u.Add("p_id",self.PddPid[0])
	return self.ClientHttp(u)

}

func (self *Pdd) SearchGoods(words ...string)interface{}{
	u := &url.Values{}
	u.Add("client_id",self.Info.Client_id)
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	u.Add("type","pdd.ddk.goods.search")
	for _,k := range words{
		u.Add("keyword",k)
	}
	return self.ClientHttp(u)
}

