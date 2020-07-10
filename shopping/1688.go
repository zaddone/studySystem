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
	"bytes"
	"strconv"
)
var (
	//1688Url = "https://gw.open.1688.com/openapi/param2/%s/6020087"
	Url1688 = "https://gw.open.1688.com/openapi/"
	AlibabaShopping *Alibaba
	alibabatimeFormat = "20060102150405000-0700"
	goodsDB = []byte("product")
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

func (self *AlAddrForOrder) LoadTestDB() {
	self.FullName = "赵伟杰"
	self.Mobile = "18628175526"
	self.CityText = "成都市"
	self.ProvinceText = "四川省"
	self.AreaText = "郫都区"
	self.TownText = "犀浦街道"
	self.Address = "校园路55号交大卡布里城1栋1单元1708号"
}

type AlProductForOrder struct {
	Offerid float64 `json:"offerId"`
	SpecId string `json:"specId"`
	Quantity float64 `json:"quantity"`
}

func (self *AlProductForOrder)LoadTestDB(){
	self.Offerid = 586899647105
	self.SpecId = "4fac2a41ec29cef08c68c5cac25382d8"
	self.Quantity = 1
}

func NewAlibaba(sh *ShoppingInfo,siteDB string) (*Alibaba){
	j:= &Alibaba{Info:sh,DbPath:"alibaba.db"}
	if siteDB == "" {
		return j
	}
	go func(){
		for{
			if j.Info.ReTimeOut == "" {
				j.Info.ReTimeOut = "20201120194552000+0800"
			}
			reTimeOut,err := time.Parse(alibabatimeFormat,j.Info.ReTimeOut)
			if err != nil {
				panic(err)
			}
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
			fmt.Println("retoken_",res)
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
			//fmt.Println(res)
			fmt.Println("retoken",res)
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

func (self *Alibaba) GetTraceView(id string) interface{} {
	//com.alibaba.logistics:alibaba.trade.getLogisticsInfos.buyerView-1
	uri := "1/com.alibaba.logistics/alibaba.trade.getLogisticsInfos.buyerView"
	u := &url.Values{}
	u.Add("orderId",id)
	u.Add("webSite","1688")
	u.Add("fields","company.name,sender,receiver,sendgood")
	u.Add("access_token",self.Info.Token)
	return self.ClientHttp(uri,u)
}
func (self *Alibaba) GetTraceInfo(id string) interface{} {
	//com.alibaba.logistics:alibaba.trade.getLogisticsTraceInfo.buyerView-1
	uri := "1/com.alibaba.logistics/alibaba.trade.getLogisticsTraceInfo.buyerView"
	u := &url.Values{}
	u.Add("orderId",id)
	u.Add("webSite","1688")
	u.Add("access_token",self.Info.Token)
	return self.ClientHttp(uri,u)


}

func (self *Alibaba) GetCategory(id string) interface{} {
	uri := "1/com.alibaba.product/alibaba.category.get"
	u := &url.Values{}
	u.Add("categoryID",id)
	u.Add("access_token",self.Info.Token)
	obj := self.ClientHttp(uri,u)
	if obj == nil {
		return nil
	}
	return obj.(map[string]interface{})["categoryInfo"]
	//fmt.Println(obj)
	//return obj
	//com.alibaba.product:alibaba.category.get-1
}
func (self *Alibaba) ClearOrder(orderid string)interface{}{
	//alibaba.trade.cancel
	uri := "1/com.alibaba.trade/alibaba.trade.cancel"
	u := &url.Values{}
	u.Add("webSite","1688")
	u.Add("tradeID",orderid)
	u.Add("cancelReason","other")
	u.Add("access_token",self.Info.Token)
	return self.ClientHttp(uri,u)

}

func (self *Alibaba) PreviewCreateOrder(a *AlAddrForOrder,p []*AlProductForOrder)interface{}{
	addr_,err := json.Marshal(a)
	if err != nil {
		return err
	}
	product_,err := json.Marshal(p)
	if err != nil {
		return err
	}
	//com.alibaba.trade:alibaba.trade.fastCreateOrder-1
	uri := "1/com.alibaba.trade/alibaba.createOrder.preview"
	u := &url.Values{}
	u.Add("flow","saleproxy" )
	u.Add("addressParam",string(addr_))
	//u.Add("addressParam","")
	u.Add("cargoParamList",string(product_))
	u.Add("access_token",self.Info.Token)
	u.Add("invoiceParam","")
	return self.ClientHttp(uri,u)
	//fmt.Println(obj)

}
func (self *Alibaba) CreateOrder(a *AlAddrForOrder,p []*AlProductForOrder)interface{}{

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
	return self.ClientHttp(uri,u)


}

func (self *Alibaba) GoodsGet(goodsId string,hand func(interface{}))error {
	return self.OpenDB(false,func(t *bolt.Tx)error{
		b := t.Bucket(goodsDB)
		if b == nil {
			return nil
		}
		val := b.Get([]byte(goodsId))
		if val == nil {
			return fmt.Errorf("find not")
		}
		var db interface{}
		err := json.Unmarshal(val,&db)
		if err != nil {
			return err
		}
		hand(db)
		return nil
	})
}

func (self *Alibaba) GoodsShow(num []byte,hand func(interface{})error)error {
	return self.OpenDB(false,func(t *bolt.Tx)error{
		b := t.Bucket(goodsDB)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		var k,v []byte
		if len(num) == 0 {
			k,v = c.First()
			if k == nil {
				return nil
			}
		}else{
			k,v = c.Seek(num)
			if bytes.Equal(k,num){
				k,v  = c.Next()
			}
		}
		var err error
		for ;k != nil;k,v = c.Next() {
			var db interface{}
			err = json.Unmarshal(v,&db)
			if err != nil {
				return err
			}
			err = hand(db)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (self *Alibaba) GoodsDetail(words ...string)interface{}{
	uri := "1/com.alibaba.product/alibaba.agent.product.simple.get"
	u := &url.Values{}
	u.Add("webSite","1688" )
	u.Add("productID",words[0] )
	u.Add("access_token",self.Info.Token )
	obj := self.ClientHttp(uri,u)
	//fmt.Println(obj)
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

func (self *Alibaba) ClearProduct() error {

	return self.OpenDB(true,func(t *bolt.Tx)error{
		return t.DeleteBucket(goodsDB)
	})

}

func (self *Alibaba) SaveProduct(k string ,obj interface{}) error {

	return self.OpenDB(true,func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(goodsDB)
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

func (self *Alibaba) Crossborder(id string) error {

	uri := "1/com.alibaba.product/alibaba.product.follow.crossborder"
	u := &url.Values{}
	u.Add("productId",id)
	u.Add("access_token",self.Info.Token )
	obj := self.ClientHttp(uri,u)
	if obj.(map[string]interface{})["code"].(float64) == 0 {
		return nil
	}
	//fmt.Println(obj)
	return fmt.Errorf("%v",obj)
}

func (self *Alibaba) UnCrossborder(id string) error {

	uri := "1/com.alibaba.product/alibaba.product.unfollow.crossborder"
	u := &url.Values{}
	u.Add("productId",id)
	u.Add("access_token",self.Info.Token )
	obj := self.ClientHttp(uri,u)
	//fmt.Println(obj)
	//return obj
	if obj.(map[string]interface{})["code"].(float64) == 0 {
		return nil
	}
	//fmt.Println(obj)
	return fmt.Errorf("%v",obj)
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
