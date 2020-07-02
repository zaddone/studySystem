package main
import(
	"fmt"
	//"math"
	"encoding/json"
	//"encoding/gob"
	"github.com/zaddone/studySystem/shopping"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/config"
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	//"github.com/go-gomail/gomail"
	"strconv"
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
	"crypto/tls"
	"crypto/x509"
	"log"
)

var(
	Remote = flag.String("r", "http://127.0.0.1:8080/v2","remote")
	TimeFormat = "20060102150405"
	wxOrderDB = "wxOrder.db"
	//aliInfo *shopping.Alibaba = nil
)

func openDB(isWrite bool,hand func(*bolt.Tx)error)error{
	db,err := bolt.Open(wxOrderDB,0600,nil)
	if err != nil {
		return err
	}
	defer db.Close()
	t,err := db.Begin(isWrite)
	if err != nil {
		return err
	}
	if isWrite {
		defer t.Commit()
	}
	return hand(t)
}

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
		sign_  = append(sign_,k+"="+fmt.Sprint(u[k]))
	}
	sign :=fmt.Sprintf("%s&key=%s",strings.Join(sign_,"&"),config.Conf.Apikeyv3)
	fmt.Println(sign)
	u["sign"]=fmt.Sprintf("%X", md5.Sum([]byte(sign)))
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
	//if aliInfo != nil {
	//	return hand(aliInfo)
	//}
	Info := &shopping.ShoppingInfo{}
	err  := requestHttp("/shopping/1688","GET",nil,nil,func(body io.Reader,res *http.Response)error{
		return json.NewDecoder(body).Decode(Info)
	})
	if err != nil {
		return err
	}
	//aliInfo = shopping.NewAlibaba(Info,"")
	return hand(shopping.NewAlibaba(Info,""))
}

func ViewPay(o *shopping.AlAddrForOrder,p []*shopping.AlProductForOrder,hand func(interface{})error) error {
	return initAlibaba(func(ali *shopping.Alibaba)error {
		return hand(ali.PreviewCreateOrder(o,p))
	})
}

func ShopPay(o *shopping.AlAddrForOrder,p []*shopping.AlProductForOrder,hand func(interface{})error) error {
	return initAlibaba(func(ali *shopping.Alibaba)error {
		return hand(ali.CreateOrder(o,p))
	})
}

type clientInfo struct{
	Appid string
	Openid string
	Clientip string
	SumPayment float64
	Name string
	Tag string
	Img string
}

type OrderInfo struct {
	Goods []*shopping.AlProductForOrder
	Addr  *shopping.AlAddrForOrder
	Client *clientInfo
	Notify *notify
	Alibaba map[string]interface{}
	//TraceView string
	Orderid string
}
func GetOrderList(openid,orderid string,hand func(*OrderInfo)error)error{
	var err error
	return openDB(false,func(t *bolt.Tx)error{
		b := t.Bucket([]byte(openid))
		if b == nil {
			return fmt.Errorf("openid is nil")
		}
		c:=b.Cursor()
		var k,v []byte
		if len(orderid) == 0 {
			k,v = c.First()
		}else{
			ord := []byte(orderid)
			k,v = c.Seek(ord)
			if bytes.EqualFold(k,ord){
				k,v = c.Next()
			}
		}
		for ;k != nil;k,v = c.Next(){
			var o OrderInfo
			err = json.Unmarshal(v,&o)
			//err = gob.NewDecoder(bytes.NewReader(v)).Decode(&o)
			if err != nil {
				return err
			}
			o.Orderid = string(k)
			err = hand(&o)
			if err != nil {
				return nil
				//return err
			}
		}
		return nil
		//return gob.NewDecoder(bytes.NewReader(b.Get([]byte(orderid)))).Decode(self)
	})
}

func (self *OrderInfo)ToByte(hand func([]byte)error) (err error) {
	db,err := json.Marshal(self)
	//var network bytes.Buffer
	//err = gob.NewEncoder(&network).Encode(self)
	if err != nil {
		return err
	}
	return hand(db)
}

func (self *OrderInfo)Save(orderid string)error{
	return openDB(true,func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists([]byte(self.Client.Openid))
		if err != nil {
			return err
		}
		return self.ToByte(func(val []byte)error{
			return b.Put([]byte(orderid),val)
		})
	})
}

func (self *OrderInfo)Load(openid,orderid string)error{
	return openDB(false,func(t *bolt.Tx)error{
		b := t.Bucket([]byte(openid))
		if b == nil {
			return fmt.Errorf("openid is nil")
		}
		return json.Unmarshal(b.Get([]byte(orderid)),self)
		//return gob.NewDecoder(bytes.NewReader()).Decode(self)
	})
}

func (self *OrderInfo)payRefund(hand func(interface{})error )error{
	uri := "https://api.mch.weixin.qq.com/secapi/pay/refund"
	u := map[string]interface{}{}
	u["appid"]=self.Client.Appid
	u["out_trade_no"]=self.Notify.Out_trade_no
	u["out_refund_no"] = RandString(64)
	u["total_fee"] = self.Notify.Total_fee
	u["refund_fee"] = self.Notify.Total_fee
	//u["cash_fee"] = self.Notify.Cash_fee
	//u["notify_url"] = "https://www.zaddone.com/wxpay/pay/notify_url_refund"
	noPemSign(u)
	body, err := xml.MarshalIndent(Map(u), "", "  ")
	if err != nil {
		fmt.Println("xml is err",err)
		return err
	}
	res,err := KeyHttpsPost(uri,"application/xml;charset=utf-8",bytes.NewReader(body))
	if err != nil {
		return err
	}
	//res.Body
	if res.StatusCode != 200 {
		db,err := ioutil.ReadAll(res.Body)
		fmt.Println(db,err)
		return fmt.Errorf("%d %s %s",res.StatusCode,res.Status,string(db))
	}
	Res := &payRefundRes{}
	err = xml.NewDecoder(res.Body).Decode(Res)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return hand(Res)

}

func (self *OrderInfo)unifiedorder(orderid string,fee int,hand func(interface{})error )error{

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
	//u["product_id"] = self.Goods.SpecId
	u["openid"] = self.Client.Openid

	noPemSign(u)
	body, err := xml.MarshalIndent(Map(u), "", "  ")
	if err != nil {
		fmt.Println("xml is err",err)
		return err
	}
	//body,err := xml.Marshal(u)
	//fmt.Println(string(body),err)
	//return err
	uri :="https://api.mch.weixin.qq.com/pay/unifiedorder"
	return request.ClientHttp__(uri,"POST",bytes.NewReader(body),nil,func(Body io.Reader,res *http.Response)error{
		if res.StatusCode != 200 {
			db,err := ioutil.ReadAll(Body)
			fmt.Println(db,err)
			return fmt.Errorf("%d %s %s",res.StatusCode,res.Status,string(db))
		}
		Res := &unifiedRes{}
		err = xml.NewDecoder(Body).Decode(Res)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return hand(Res)
	})

}
func KeyHttpsPost(url string, contentType string, body io.Reader) (*http.Response, error) {
	//var wechatPayCert = *pemcert
	//var wechatPayKey = *pemkey
	//var rootCa = "F:/cert/rootca.pem"
	var tr *http.Transport
	// 微信提供的API证书,证书和证书密钥 .pem格式
	certs, err := tls.LoadX509KeyPair(*pemcert, *pemkey)
	if err != nil {
		log.Println("certs load err:", err)
		return nil,err

	} else {
		// 微信支付HTTPS服务器证书的根证书  .pem格式
		rootCa, err := ioutil.ReadFile(*rootca)
		if err != nil {
			log.Println("err2222:", err)
		} else {
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(rootCa){
				//fmt.Println(string(publicKey))
				return nil,fmt.Errorf("public err")
			}
			//pool.AppendCertsFromPEM(rootCa)
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      pool,
					Certificates: []tls.Certificate{certs},
				},
			}
		}

	}
	client := &http.Client{Transport: tr}
	return client.Post(url, contentType, body)
}

func init(){

	Router.POST("pay/notify_url",func(c *gin.Context){
		//https://api.mch.weixin.qq.com/pay/unifiedorder
		//c.Request.Body
		var db notify
		err := xml.NewDecoder(c.Request.Body).Decode(&db)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		if !db.CheckSign(){
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		c.XML(http.StatusOK,gin.H{
			"return_code":"SUCCESS",
			"return_msg":"OK",
		})
		//return

		oi:= &OrderInfo{}
		err = oi.Load(db.Openid,db.Out_trade_no)
		if err != nil {
			fmt.Println(err)
			//c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		oi.Notify = &db


		go func(){
			err := ShopPay(oi.Addr,oi.Goods,func(_db interface{})error{
				//res,err := json.Marshal(_db.(map[string]interface{})["result"])
				//if err != nil {
				//	return err
				//}
				oi.Alibaba = _db.(map[string]interface{})["result"].(map[string]interface{})
				return oi.Save(db.Out_trade_no)
			})
			if err != nil {
				fmt.Println(err)
			}
		}()
		return

	})
	pay := Router.Group("pay",func() gin.HandlerFunc {
		return checkManage
	}())
	pay.GET("/pay_refund",func(c *gin.Context){
		oi:= &OrderInfo{}
		err := oi.Load(c.Query("openid"),c.Query("orderid"))
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		err = oi.payRefund(func(db interface{})error{
			c.JSON(http.StatusOK,gin.H{"msg":err})
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
	})
	pay.GET("/getorderlist",func(c *gin.Context){
		pages,err :=strconv.Atoi(c.DefaultQuery("page","10"))
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		var li []interface{}
		err = GetOrderList(c.Query("openid"),c.Query("orderid"),func(o *OrderInfo)error{
			li = append(li,o)
			if len(li)>=pages{
				return io.EOF
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		c.JSON(http.StatusOK,gin.H{"list":li})
		return
	})
	pay.GET("/gettraceview",func(c *gin.Context){
		oi:= &OrderInfo{}
		err := oi.Load(c.Query("openid"),c.Query("orderid"))
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		//var bab interface{}
		//err = json.Unmarshal([]byte(oi.Alibaba),&bab)
		//if err != nil {
		//	c.JSON(http.StatusNotFound,gin.H{"msg":err})
		//	return
		//}
		//fmt.Println(bab)

		err =  initAlibaba(func(ali *shopping.Alibaba)error {
			c.JSON(http.StatusOK,ali.GetTraceView(oi.Alibaba["orderId"].(string)))
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
	})
	pay.GET("/cleartrace",func(c *gin.Context){
		oi:= &OrderInfo{}
		err := oi.Load(c.Query("openid"),c.Query("orderid"))
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		err =  initAlibaba(func(ali *shopping.Alibaba)error {
			c.JSON(http.StatusOK,ali.ClearOrder(oi.Alibaba["orderId"].(string)))
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}

	})
	pay.GET("/gettraceinfo",func(c *gin.Context){
		//orderid := c.Param("id")
		oi:= &OrderInfo{}
		err := oi.Load(c.Query("openid"),c.Query("orderid"))
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		//var bab interface{}
		//err = json.Unmarshal([]byte(oi.Alibaba),&bab)
		//if err != nil {
		//	c.JSON(http.StatusNotFound,gin.H{"msg":err})
		//	return
		//}

		err =  initAlibaba(func(ali *shopping.Alibaba)error {
			c.JSON(http.StatusOK,ali.GetTraceInfo(oi.Alibaba["orderId"].(string)))
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}

	})
	pay.POST("/postordertoalibababuy",func(c *gin.Context){
		db,err := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		o := &OrderInfo{}
		if err = json.Unmarshal(db,o); err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		oid := RandString(32)
		fmt.Println(o.Client.SumPayment)
		//err = o.unifiedorder(oid,int(o.Client.SumPayment),func(_db interface{})error{
		err = o.unifiedorder(oid,int(1),func(_db interface{})error{
			if len(_db.(*unifiedRes).Prepay_id)==0{
				c.JSON(http.StatusNotFound,_db)
				return nil
			}
			c.JSON(http.StatusOK,gin.H{"msg":_db,"orderid":oid})
			return o.Save(oid)
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
		}
		return


	})
	pay.POST("/postordertoalibaba",func(c *gin.Context){
		fmt.Println(c.Request.RemoteAddr,c.Request.Header.Get("X-Forwarded-For"))
		db,err := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		//fmt.Println(string(db))
		o:= &OrderInfo{}
		//_db,_ := json.Marshal(&o)
		//fmt.Println(string(_db))
		err = json.Unmarshal(db,o)

		//fmt.Println(o)
		//o.Client.Clientip =c.Request.Header.Get("X-Forwarded-For")
		//fmt.Printf("%+v %+v\n",o.Addr,o.Goods)
		//err := json.NewDecoder(c.Request.Body).Decode(&o)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		//err = o.unifiedorder(RandString(16),10,func(body interface{})error{
		//	fmt.Println(body)
		//	c.JSON(http.StatusOK,body)
		//	return nil
		//})
		//if err != nil {
		//	c.JSON(http.StatusNotFound,gin.H{"msg":err})
		//	return
		//}
		////c.JSON(http.StatusOK,gin.H{"msg":"success"})
		//return
		err = ViewPay(o.Addr,o.Goods,func(db interface{})error{
			//o.res = db
			fmt.Println(db)
			c.JSON(http.StatusOK,db)
			return nil

			//res := db.(map[string]interface{})["result"]
			//if res == nil {
			//	c.JSON(http.StatusNotFound,gin.H{"msg1":db})
			//	return nil
			//}
			//res_ := res.(map[string]interface{})
			//orderid := res_["orderId"]
			//if orderid == nil {
			//	c.JSON(http.StatusNotFound,gin.H{"msg2":db})
			//	return nil
			//}
			//return o.unifiedorder(orderid.(string),int(res_["totalSuccessAmount"].(float64)*1.1),func(_db interface{})error{
			//	fmt.Println(_db)
			//	c.JSON(http.StatusOK,_db)
			//	return nil
			//})
			//fmt.Println(db)

		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
		}
		return
	})
}
