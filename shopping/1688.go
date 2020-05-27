package shopping
import(
	"fmt"
	"time"
	"strings"
	"net/url"
	"crypto/hmac"
	"crypto/sha1"
	"github.com/zaddone/studySystem/request"
	"github.com/boltdb/bolt"
	"encoding/json"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
)
var (
	//1688Url = "https://gw.open.1688.com/openapi/param2/%s/6020087"
	Url1688 = "https://gw.open.1688.com/openapi/"
	BuyShopping *Alibaba
	alibabatimeFormat = "20060102150405000-0700"
)

type Alibaba struct{
	Info *ShoppingInfo
	DbPath string
	//Pid string
}
type AlAddrForOrder struct {
	FullName string `json:"fullName"`
	Mobile string `json:"mobile"`
	CityText string `json:"cityText"`
	ProvinceText string `json:"provinceText"`
	AreaText string `json:"areaText"`
	TownText string `json:"townText"`
	Address string `json:"address"`
}

type AlProductForOrder struct {
	Offerid float64 `json:"offerid"`
	SpecId string `json:"specId"`
	Quantity float64 `json:"quantity"`
}

func NewAlibaba(sh *ShoppingInfo,siteDB string) (*Alibaba){

	j:= &Alibaba{Info:sh,DbPath:"alibaba.db"}
	if siteDB == "" {
		return j
	}
	//return j
	go func(){
		for{
			if j.Info.ReTimeOut == "" {
				j.Info.ReTimeOut = "20201120194552000+0800"
			}
			reTimeOut,err := time.Parse(alibabatimeFormat,j.Info.ReTimeOut)
			if err != nil {
				panic(err)
			}
			//if err != nil {
			//	fmt.Println(err)
			//	//panic(err)
			//	err := j.ReToken_(siteDB)
			//	if err != nil {

			//		fmt.Println(err)
			//		panic(err)
			//	}
			//}
			select{
			case <- time.After(time.Second*time.Duration(j.Info.TimeOut - time.Now().Unix())):
				err := j.ReToken(siteDB)
				if err != nil {
					panic(err)
				}
			case <- time.After(time.Second*time.Duration(reTimeOut.Unix() - time.Now().Unix())):
				err := j.ReToken_(siteDB)
				if err != nil {
					panic(err)
				}
			}
		}
	}()
	//go func(){
	//	for{
	//	//time.Now().Unix()
	//	err := j.ReToken(siteDB)
	//	if err != nil {
	//		panic(err)
	//	}
	//	time.Sleep(time.Second*time.Duration(j.Info.TimeOut-time.Now().Unix()))
	//	}
	//}()
	return j

}
func (self *Alibaba) ReToken_ (siteDB string) error {
	uri := "https://gw.open.1688.com/openapi/param2/1/system.oauth2/postponeToken/"+self.Info.Client_id
	u := url.Values{}
	u.Set("client_id",self.Info.Client_id)
	u.Set("client_secret",self.Info.Client_secret)
	u.Set("refresh_token",self.Info.ReToken)
	u.Set("access_token",self.Info.Token)
	return request.ClientHttp_(
		uri+"?"+u.Encode(),
		"POST",nil,nil,
		func(body io.Reader,start int)error{
			var res map[string]interface{}
			err := json.NewDecoder(body).Decode(&res)
			if err != nil {
				return err
			}
			fmt.Println(res)
			if res["access_token"] == nil {
				return io.EOF
			}
			self.Info.Token = res["access_token"].(string)
			self.Info.ReToken=res["refresh_token"].(string)
			//self.Info.TimeOut =time.Now().Unix()+int64(res["expires_in"].(float64))
			self.Info.Uri = res["memberId"].(string)
			self.Info.ReTimeOut = res["refresh_token_timeout"].(string)

			exp,err := strconv.Atoi(res["expires_in"].(string))
			if err != nil {
				return err
			}
			self.Info.TimeOut =time.Now().Unix()+int64(exp)
			return OpenSiteDB(siteDB,self.Info.SaveToDB)
		},
	)

}
func (self *Alibaba) ReToken (siteDB string) error {

	uri := "https://gw.open.1688.com/openapi/param2/1/system.oauth2/getToken/"+self.Info.Client_id
	u := url.Values{}
	u.Set("client_id",self.Info.Client_id)
	u.Set("client_secret",self.Info.Client_secret)
	u.Set("grant_type","refresh_token")
	u.Set("refresh_token",self.Info.ReToken)
	return request.ClientHttp_(
		uri+"?"+u.Encode(),
		"POST",nil,nil,
		func(body io.Reader,start int)error{
			var res map[string]interface{}
			err := json.NewDecoder(body).Decode(&res)
			if err != nil {
				return err
			}
			fmt.Println(res)
			if res["access_token"] == nil {
				return io.EOF
			}
			self.Info.Token = res["access_token"].(string)
			//self.Info.ReToken=res["refresh_token"].(string)
			//self.Info.TimeOut =time.Now().Unix()+int64(res["expires_in"].(float64))
			exp,err := strconv.Atoi(res["expires_in"].(string))
			if err != nil {
				return err
			}
			self.Info.TimeOut =time.Now().Unix()+int64(exp)
			self.Info.Uri = res["memberId"].(string)
			return OpenSiteDB(siteDB,self.Info.SaveToDB)
		},
	)

}
func (self *Alibaba) ClientHttp(uri string,u *url.Values)( out interface{}){

	u.Add("memberId",self.Info.Uri)
	u.Add("_aop_timestamp",fmt.Sprintf("%d",time.Now().Unix()*1000))
	var li []string
	for k,_ := range *u {
		li = append(li,k+u.Get(k))
	}
	sort.Strings(li)
	uri = "param2/"+uri+"/"+self.Info.Client_id
	sign := uri+strings.Join(li,"")
	//key := []byte("123456")
	fmt.Println(self.Info,sign)
	mac := hmac.New(sha1.New, []byte(self.Info.Client_secret))
	mac.Write([]byte(sign))
	u.Add("_aop_signature",fmt.Sprintf("%X", mac.Sum(nil)))


	var err error
	uri = Url1688 + uri+"?"+u.Encode()
	err = request.ClientHttp_(
		uri,
		"POST",nil,
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
func (self *Alibaba) CreateOrder(
	a *AlAddrForOrder,
	p *AlProductForOrder)interface{}{
	addr_,err := json.Marshal(a)
	if err != nil {
		return err
	}
	product_,err := json.Marshal(p)
	if err != nil {
		return err
	}
	//com.alibaba.trade:alibaba.trade.fastCreateOrder-1
	uri := "1/com.alibaba.trade/alibaba.trade.fastCreateOrder"
	u := &url.Values{}
	u.Add("flow","saleproxy" )
	u.Add("addressParam",string(addr_))
	u.Add("cargoParamList",string(product_))
	u.Add("access_token",self.Info.Token)
	obj := self.ClientHttp(uri,u)
	//fmt.Println(obj)
	return obj


}
func (self *Alibaba) GoodsDetail(words ...string)interface{}{
	uri := "1/com.alibaba.product/alibaba.agent.product.simple.get"
	u := &url.Values{}
	u.Add("webSite","1688" )
	u.Add("productID",words[0] )
	u.Add("access_token",self.Info.Token )
	obj := self.ClientHttp(uri,u)
	fmt.Println(obj)
	return obj
}

func (self *Alibaba) OpenDB (read bool,hand func(*bolt.Tx)error) error {

	db,err := bolt.Open(self.DbPath,0600,nil)
	if err != nil {
		return err
	}
	defer db.Close()
	t,err := db.Begin(read)
	if err != nil {
		return err
	}
	if read {
		defer t.Commit()
	}
	return hand(t)

}
func (self *Alibaba) SaveProduct(k string ,obj interface{}) error {

	return self.OpenDB(true,func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists([]byte("product"))
		if err != nil {
			return err
		}
		v,err := json.Marshal(obj)
		if err != nil {
			return err
		}
		return b.Put([]byte(k),v)
	})
}


func (self *Alibaba) SearchGoods(words ...string)interface{}{

	uri := "1/com.alibaba.search/alibaba.search.cbu.general"
	u := &url.Values{}
	u.Add("q",words[0] )
	u.Add("access_token",self.Info.Token )
	obj := self.ClientHttp(uri,u)
	fmt.Println(obj)
	return obj

}
