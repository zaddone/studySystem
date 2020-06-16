package main
import(
	"fmt"
	"encoding/json"
	"github.com/zaddone/studySystem/shopping"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/config"
	"github.com/gin-gonic/gin"
	"crypto/md5"
	"encoding/xml"
	"time"
	"bytes"
	"strings"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
)
var(
	Remote = flag.String("r", "http://127.0.0.1:8080/v2","remote")
	TimeFormat = "20060102150405"
)
func noPemSign(u map[string]interface{}){
	u["nonce_str"] = RandString(16)
	u["mch_id"] = *MerchantId
	var li []string
	for k,_ := range u{
		li = append(li,k)
	}
	sort.Strings(li)
	sign_:=[]string{}
	for _,k := range li {
		sign_  = append(sign_,k+"="+fmt.Sprintln(u[k]))
	}
	fmt.Println(sign_)
	u["sign"]=fmt.Sprintf("%X", md5.Sum([]byte(strings.Join(sign_,"&")+config.Conf.Apikeyv3)))
}

func requestHttp(path,Method string,u url.Values, body io.Reader,hand func(io.Reader,*http.Response)error)error{
	if u == nil {
		u = url.Values{}
	}
	addSign(&u)
	return request.ClientHttp__(*Remote+path+"?"+u.Encode(),Method,body,nil,hand)
}

func addSign(u *url.Values){
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	li := []string{config.Conf.Minitoken}
	for _,v := range *u{
		li = append(li,v...)
	}
	sort.Strings(li)
	u.Add("sign",shopping.Sha1([]byte(strings.Join(li,""))))
}

func initAlibaba(hand func(*shopping.Alibaba)error)error{
	Info := &shopping.ShoppingInfo{}
	err  := requestHttp("/shopping/1688","GET",nil,nil,func(body io.Reader,res *http.Response)error{
		return json.NewDecoder(body).Decode(Info)
	})
	if err != nil {
		return err
	}
	return hand(shopping.NewAlibaba(Info,""))
}

func ShopPay(o *shopping.AlAddrForOrder,p *shopping.AlProductForOrder,hand func(interface{})error) error {
	return initAlibaba(func(ali *shopping.Alibaba)error {
		return hand(ali.CreateOrder(o,[]*shopping.AlProductForOrder{p}))
	})
}

type clientInfo struct{
	Appid string
	Openid string
	Clientip string
}
type OrderInfo struct {
	Goods shopping.AlProductForOrder
	Addr  shopping.AlAddrForOrder
	Client clientInfo
}
func (self *OrderInfo)unifiedorder(orderid string,fee int,hand func(db interface{})error )error{

	u := map[string]interface{}{}
	u["appid"]=self.Client.Appid
	u["body"] = "米果小店-订单支付"
	u["out_trade_no"] = orderid
	u["total_fee"] = fee
	u["spbill_create_ip"] = self.Client.Clientip
	begin := time.Now()
	u["time_start"] = begin.Format(TimeFormat)
	u["time_expire"] = begin.AddDate(0,0,5).Format(TimeFormat)
	u["notify_url"] = "https://www.zaddone.com/wxpay/pay/notify_url"
	u["trade_type"] = "JSAPI"
	u["product_id"] = self.Goods.SpecId
	u["openid"] = self.Client.Openid

	noPemSign(u)
	body, err := xml.MarshalIndent(Map(u), "", "  ")
	if err != nil {
		return err
	}
	//body,err := xml.Marshal(u)
	//fmt.Println(string(body),err)
	//return err
	uri :="https://api.mch.weixin.qq.com/pay/unifiedorder"
	return request.ClientHttp__(uri,"POST",bytes.NewReader(body),nil,func(Body io.Reader,res *http.Response)error{
		if res.StatusCode != 200 {
			return fmt.Errorf(res.Status)
		}
		db,err := ioutil.ReadAll(Body)
		if err != nil {
			return err
		}
		var db map[string]interface{}
		err = xml.NewDecoder(Body).Decode(&db)
		if err != nil {
			return err
		}
		//db[]
		//fmt.Println(string(db))
		return hand(db)
	})

}

func init(){
	pay := Router.Group("pay",func() gin.HandlerFunc {
		return checkManage
	}())
	pay.GET("/notify_url",func(c *gin.Context){
		//https://api.mch.weixin.qq.com/pay/unifiedorder

	})
	pay.POST("/postordertoalibaba",func(c *gin.Context){
		fmt.Println(c.Request.RemoteAddr,c.Request.Header.Get("X-Forwarded-For"))
		db,err := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		fmt.Println(string(db))
		var o OrderInfo
		_db,_ := json.Marshal(&o)
		fmt.Println(string(_db))

		err = json.Unmarshal(db,&o)
		fmt.Printf("%+v\n",o)
		//err := json.NewDecoder(c.Request.Body).Decode(&o)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		//o.unifiedorder("",10)
		//c.JSON(http.StatusOK,gin.H{"msg":"success"})
		//return
		err = ShopPay(&(o.Addr),&(o.Goods),func(db interface{})error{
			//o.res = db
			res := db.(map[string]interface{})["result"]
			if res == nil {
				c.JSON(http.StatusNotFound,gin.H{"msg":db})
				return nil
			}
			res_ := res.(map[string]interface{})
			orderid := res_["orderId"]
			if orderid == nil {
				c.JSON(http.StatusNotFound,gin.H{"msg":db})
				return nil
			}
			return o.unifiedorder(orderid.(string),int(res_["totalSuccessAmount"].(float64)*1.1),func(body interface{})error{
				c.JSON(http.StatusOK,gin.H{"msg":db})
				return nil
			})
			//fmt.Println(db)
			//return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
		}
		return
	})
}
