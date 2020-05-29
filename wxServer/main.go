package main
import(
	"fmt"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/config"
	"github.com/zaddone/studySystem/shopping"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"sort"
	"crypto/sha1"
	"strings"
	"strconv"
	"net/http"
	"flag"
	"io"
	"io/ioutil"
	"net/url"
	"time"
	//"github.com/boltdb/bolt"
)
var (
	Port = flag.String("p",":8084","port")
	Router = gin.Default()
	AppSecret = flag.String("secret","c6f9455b3cfef5813f749ee86d9f8c17","app secret")
	AppKey = flag.String("key","wx4ff6f10c37ce208d","app secret")
	Sign = flag.String("sign","miguotuijian2020miguotuijian2020","sign")
	WxTokenUrl = "https://api.weixin.qq.com/cgi-bin/token"
	WXTOKEN *WxToken = new(WxToken)
)
//{"access_token":"ACCESS_TOKEN","expires_in":7200}
type WxToken struct {
	Access_token string `json:"access_token"`
	Expires_in float64 `json:"expires_in"`
	EndTime int64
}
func (self *WxToken)GetToken() error {

	//https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET
	u := url.Values{}
	u.Set("grant_type","client_credential")
	u.Set("appid",*AppKey)
	u.Set("secret",*AppSecret)
	//fmt.Println(u)
	//WxTokenUrl+"?"+u.Encode()
	return request.ClientHttp_(
		WxTokenUrl+"?"+u.Encode(),
		"GET",
		nil,
		nil,
		func(res io.Reader,st int)error{
			_db,_ := ioutil.ReadAll(res)
			fmt.Println(string(_db))
			//fmt.Println(st)
			if st != 200 {
				//_db,_ := ioutil.ReadAll(res)
				//fmt.Println(_db)
				return fmt.Errorf("%s",_db)
			}
			err := json.Unmarshal(_db,self)
			if err != nil {
				return err
			}

			//fmt.Println(res)
			//fmt.Println(self)
			//err := json.NewDecoder(res).Decode(self)
			//if err != nil {
			//	_db,_ := ioutil.ReadAll(res)
			//	fmt.Println(_db)
			//	//panic(err)
			//	//fmt.Println(err)
			//	return err
			//}
			self.EndTime = time.Now().Unix()+int64(self.Expires_in)
			//fmt.Println(self)
			return nil
		},
	)

}
func(self *WxToken) String() string {
	if time.Now().Unix()>self.EndTime{
		err := self.GetToken()
		if err != nil {
			panic(err)
		}
	}
	return self.Access_token

}
func checkManage(c *gin.Context){
	timestamp:=c.Query("timestamp")
	if timestamp == ""{
		fmt.Println("stamp = nil")
		c.Abort()
		return
	}
	stamp,err := strconv.ParseInt(timestamp,10,64)
	if err != nil {
		c.Abort()
		return
	}
	d := time.Now().Unix() - stamp
	if d<0 {
		d=-d
	}
	if d>60 {
		c.Abort()
		return
	}
	signature := c.Query("sign")
	if signature == "" {
		c.Abort()
		return
	}
	qu := []string{config.Conf.WXtoken}
	for k,v := range c.Request.URL.Query(){
		if k == "sign" {
			continue
		}
		qu = append(qu,v...)
	}
	sort.Strings(qu)
	if signature != shopping.Sha1([]byte(strings.Join(qu,""))) {
		c.Abort()
		return
	}
	//fmt.Println("next")
	c.Next()
}

func checkWXserver(c *gin.Context){

	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce     := c.Query("nonce")
	li:=[]string{*Sign,timestamp,nonce}
	sort.Strings(li)
	fmt.Println(li)
	//config.Conf.Apikeyv3
	//key := []byte(config.Conf.Apikeyv3)
	//fmt.Println(config.Conf.Apikeyv3)
	s := sha1.New()
	_,err := io.WriteString(s, strings.Join(li,""))
	if err != nil {
		fmt.Println(err)
		c.Abort()
		return
	}
	//_sign :=  fmt.Sprintf("%x", s.Sum(nil))
	//mac := hmac.New(sha1.New, key)
	//mac.Write([]byte(strings.Join(li,"")))
	//sign := fmt.Sprintf("%x\n", mac.Sum(nil))
	//fmt.Println(sign)
	if fmt.Sprintf("%x", s.Sum(nil)) != signature {
		c.Abort()
		return
	}
	c.Next()
}

func init(){
	flag.Parse()
	wxserverFunc := func() gin.HandlerFunc {
		return checkWXserver
	}()
	manageFunc := func() gin.HandlerFunc {
		return checkManage
	}()
	Router.GET("/",wxserverFunc,func(c *gin.Context){
		c.String(http.StatusOK,c.Query("echostr"))
	})
	Router.GET("/token_test",func(c *gin.Context){
		c.String(http.StatusOK,WXTOKEN.String())
	})
	Router.GET("/token",manageFunc,func(c *gin.Context){
		c.String(http.StatusOK,WXTOKEN.String())
	})


}
func main(){
	Router.Run(*Port)
	//select{}
}
