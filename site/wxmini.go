package main

import(
	"sort"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/shopping"
	"github.com/zaddone/studySystem/config"
	"github.com/gin-gonic/gin"
	"net/url"
	"encoding/json"
	"time"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var(
	wxToKenUrl= "https://api.weixin.qq.com/cgi-bin/token"
	toKen string
	sendMiniMsgUrl = "https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token="

	SessionChan = make(chan interface{},100)
)

func init(){
	wxToKenUrl = fmt.Sprintf("%s?%s",wxToKenUrl,
	(&url.Values{
		"grant_type":	[]string{"client_credential"},
		"appid":	[]string{config.Conf.WXAppid},
		"secret":	[]string{config.Conf.WXSec},
	}).Encode())
	//fmt.Println(wxToKenUrl)
	k := setToken()
	fmt.Println("setToKen",k)
	//k := time.Duration(setToken())*time.Second
	go func(){
		for{

			time.Sleep(time.Duration(k)*time.Second)
			k = setToken()
		}
	}()
	Router.POST("/wxmini",handMiniQuery)
	Router.GET("/wxmini",handMiniQuery)
}

func setToken() int {
	db := map[string]interface{}{}
	err := request.ClientHttp(wxToKenUrl,"GET",[]int{200},nil,func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&db)
	})
	if err != nil {
		panic(err)
	}
	if db["access_token"]==nil {

		fmt.Println(db)
		time.Sleep(1*time.Hour)
		return setToken()
	}
	toKen = db["access_token"].(string)
	return int(db["expires_in"].(float64)) - 100

}
func handMiniQuery(c *gin.Context){
	timestamp:=c.Query("timestamp")
	if timestamp == ""{
		fmt.Println("timestamp = nil")
		c.String(http.StatusOK,"")
		return
	}
	stamp,err := strconv.ParseInt(timestamp,10,64)
	if err != nil {
		fmt.Println(err)
		c.String(http.StatusOK,"")
		return
	}
	d := time.Now().Unix() - stamp
	if d<0 {
		d=-d
	}
	if d>60 {
		fmt.Println("time out")
		c.String(http.StatusOK,"")
		return
	}
	signature := c.Query("signature")
	nonce := c.Query("nonce")
	echostr:= c.Query("echostr")
	li := []string{config.Conf.Minitoken,timestamp,nonce}
	sort.Strings(li)
	li_ := shopping.Sha1([]byte(strings.Join(li,"")))
	if signature != li_ {
		fmt.Println("sign is er",li_,signature)
		c.String(http.StatusOK,"")
		return
	}
	c.String(http.StatusOK,echostr)
	handBody(c)
}
func handBody(c *gin.Context){
	var db map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&db)
	if err != nil {
		panic(err)
	}
	fmt.Println(db)
	SessionChan<-db
}
