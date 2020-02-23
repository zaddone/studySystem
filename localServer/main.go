package main
import(
	"fmt"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/shopping"
	//"github.com/zaddone/studySystem/alimama"
	"github.com/gin-gonic/gin"
	"net/url"
	"time"
	"sort"
	"strings"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"encoding/json"
	//"flag"
)

var(
	WXtoken = "zhaoweijie2020"
	Router = gin.Default()
	Remote = "https://www.zaddone.com/v1"
	//siteDB  = flag.String("db","/home/dimon/Documents/wxbot/bin/SiteDB","db")
)
func Sign(c *gin.Context){
	url_ := c.Request.URL.Query()
	addSign(&url_)
}
func addSign(u *url.Values){
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	li := []string{WXtoken}
	for _,v := range *u{
		li = append(li,v...)
	}
	sort.Strings(li)
	u.Add("sign",shopping.Sha1([]byte(strings.Join(li,""))))
}
func HandForward(c *gin.Context){
	//c.Request.Body.Close()
	err := requestHttp(
		c.Request.URL.Path,
		c.Request.Method,
		c.Request.URL.Query(),
		c.Request.Body,
		func(body io.Reader,res *http.Response)error{
			c.DataFromReader(res.StatusCode,res.ContentLength,res.Header.Get("content-type"),res.Body,nil)
			return nil
		},
	)

	if err != nil {
		c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
	}
}
func requestHttp(path,Method string,u url.Values, body io.Reader,hand func(io.Reader,*http.Response)error)error{
	addSign(&u)
	return request.ClientHttp__(Remote+path+"?"+u.Encode(),Method,body,nil,hand)
}
func InitShoppingMap(){
	requestHttp("/shopping","GET",url.Values{},nil,func(body io.Reader,res *http.Response)error{
		var db []*shopping.ShoppingInfo
		err := json.NewDecoder(body).Decode(&db)
		if err != nil {
			return err
		}
		for _,sh := range db {
			hand := shopping.FuncMap[sh.Py]
			if hand != nil {
				shopping.ShoppingMap.Store(sh.Py,hand(sh))
			}
		}
		//fmt.Println(shopping.ShoppingMap)
		return nil
	})
}

func init(){
	//flag.Parse()
	//shopping.InitShoppingMap(*siteDB)
	Router.GET("updatesite/:py",HandForward)
	Router.GET("shopping/:py",HandForward)
	Router.GET("shopping",HandForward)
	Router.GET("order/:py",HandForward)
	//Router.POST("updateorder/:py",HandForward)
	go Router.Run(":8088")
}
func DownOrder(){
	shopping.ShoppingMap.Range(func(k,v interface{})bool{
		v_ := v.(shopping.ShoppingInterface)
		//v_.GetInfo().Update = 0
		//var orderlist []interface{}
		err := v_.OrderDown(func(db interface{}){
			db_,err := json.Marshal(db)
			if err != nil {
				panic(err)
				fmt.Println(err)
				return
			}
			u := url.Values{}
			u.Add("orderid",db.(map[string]interface{})["order_id"].(string))
			//var req interface{}
			err = requestHttp("/updateorder/"+k.(string),"POST",u,bytes.NewReader(db_),func(body io.Reader,res *http.Response)error{
				db,err := ioutil.ReadAll(body)
				fmt.Println("order",string(db))
				return err
				//return json.NewDecoder(body).Decode(&req)
			})
			if err != nil {
				panic(err)
				fmt.Println(err)
			}
			//fmt.Println(req)
			//orderlist = append(orderlist,db)
			//fmt.Println(db)
		})
		if err != nil {
			fmt.Println(k,err)
			return true
		}
		u_:= url.Values{}
		u_.Set("update",fmt.Sprintf("%d",v_.GetInfo().Update))
		//var req_ interface{}
		err = requestHttp("/updatesite/"+k.(string),"GET",u_,nil,func(body io.Reader,res *http.Response)error{
			//return json.NewDecoder(body).Decode(&req_)
			db,err := ioutil.ReadAll(body)
			//fmt.Println(db)
			fmt.Println("site",string(db))
			//fmt.Println(string(db))
			return err
		})
		if err != nil {
			panic(err)
			fmt.Println(err)
		}
		//fmt.Println(req_)
		return true
	})
}
func main(){
	InitShoppingMap()
	DownOrder()
	//sh,_ := shopping.ShoppingMap.Load("pinduoduo")

	//select{}
}
