package userManage
import(
	"fmt"
	"encoding/json"
	"flag"
	"io"
	"time"
	"strings"
	"sort"
	//"sync"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/zaddone/studySystem/shopping"
	"github.com/zaddone/studySystem/config"
	"github.com/zaddone/studySystem/request"
	"net/http"
	"net/url"
)

var(
	WXtoken = config.Conf.Minitoken
	Router = gin.Default()
	Port = flag.String("p",":8082","port")
	Remote = flag.String("r", "http://127.0.0.1:8080/v2","remote")
)

func addSign(u *url.Values){
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	li := []string{WXtoken}
	for _,v := range *u{
		li = append(li,v...)
	}
	sort.Strings(li)
	u.Add("sign",shopping.Sha1([]byte(strings.Join(li,""))))
}
func requestHttp(path,Method string,u url.Values, body io.Reader,hand func(io.Reader,*http.Response)error)error{
	addSign(&u)
	return request.ClientHttp__(*Remote+path+"?"+u.Encode(),Method,body,nil,hand)
}

func UpdateShopping(sh *shopping.ShoppingInfo) error {
	return requestHttp(
		"/updatesite/"+sh.Py,
		"GET",
		url.Values{"update":[]string{fmt.Sprintf("%d",sh.Update)}},
		nil,
		func(body io.Reader,res *http.Response)error{
			if res.StatusCode != 200 {
				return fmt.Errorf(res.Status)
			}
			fmt.Println(sh)
			return nil
	})
}
func InitShoppingMap()error{
	return requestHttp("/shopping","GET",url.Values{},nil,func(body io.Reader,res *http.Response)error{
		var db []*shopping.ShoppingInfo
		err := json.NewDecoder(body).Decode(&db)
		if err != nil {
			return err
		}
		for _,sh := range db {
			hand := shopping.FuncMap[sh.Py]
			if hand != nil {
				shopping.ShoppingMap.Store(sh.Py,hand(sh,""))
			}
		}
		//fmt.Println(shopping.ShoppingMap)
		return nil
	})
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
func init(){
	err := InitShoppingMap()
	if err != nil {
		panic(err)
	}

	manageFunc := func() gin.HandlerFunc {
		return checkManage
	}()
	Router.GET("/order/del",manageFunc,func(c *gin.Context){
		orderid := c.Query("orderid")
		if orderid == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"orderid error"})
			return
		}
		err := shopping.OrderDel(orderid)
		if err != nil {
			c.String(http.StatusNotFound,err.Error())
			return
		}
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
	})
	Router.GET("/order/list",manageFunc,func(c *gin.Context){
		num,err := strconv.Atoi(c.DefaultQuery("count","0"))
		if err != nil {
			c.String(http.StatusNotFound,err.Error())
			return
		}
		var li []interface{}
		err = shopping.OrderList(c.Query("orderid"),func(v map[string]interface{})error{
			li = append(li,v)
			num--
			if num<=0{
				return io.EOF
			}
			return nil
		})
		if err!= nil && err != io.EOF {
			c.String(http.StatusNotFound,err.Error())
			return
		}
		c.JSON(http.StatusOK,li)
	})
	Router.GET("order_apply/update",manageFunc,func(c *gin.Context){
		orderid := c.Query("orderid")
		if orderid == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"orderid error"})
			return
		}
		userid := c.Query("userid")
		if userid == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"userid error"})
			return
		}
		err := shopping.OrderApplyUpdate(userid,orderid)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
		}
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
	})
	Router.GET("order_apply",manageFunc,func(c *gin.Context){
		orderid := c.Query("orderid")
		if orderid == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"orderid error"})
			return
		}
		userid := c.Query("userid")
		if userid == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"userid error"})
			return
		}
		w := make(chan bool)
		go func (){
			err := shopping.OrderApply(userid,orderid,func(db interface{}){
				c.JSON(http.StatusOK,db)
				w<-true
			})
			fmt.Println(err)
			close(w)
			if err == io.EOF{
			shopping.ShoppingMap.Range(func(k,v interface{})bool{
				sh := v.(shopping.ShoppingInterface)
				err = UpdateShopping(sh.GetInfo())
				if err != nil {
					fmt.Println(err)
				}
				return true
			})
			}
		}()
		<-w
		return
	})

	Router.GET("test/:py",func(c *gin.Context){
		sh_,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh_ == nil {
			return
		}
		c.JSON(http.StatusOK,gin.H{"msg":sh_.(shopping.ShoppingInterface).Test()})

	})
	Router.POST("/updateorder/:py",manageFunc,func(c *gin.Context){
		py := c.Param("py")
		orderid := c.Query("orderid")
		if orderid == ""{
			c.JSON(http.StatusNotFound,gin.H{"msg":"orderid error"})
			return
		}
		var db map[string]interface{}
		err := json.NewDecoder(c.Request.Body).Decode(&db)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			return
		}
		err = shopping.OrderUpdate(orderid,db)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			return
		}
		endTime := db["endTime"]
		//db["py"] = py
		if endTime != nil && endTime.(float64) !=0{
			err = shopping.OrderUpdateTime(
				py,
				orderid,
				[]byte(time.Unix(int64(endTime.(float64)),0).Format("20060102")),
			)
			if err != nil {
				c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			}
		}
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
		return
	})

	Router.GET("user/sum",manageFunc,func(c *gin.Context){
		u := c.Query("userid")
		if u == "" {
			return
		}
		val := shopping.GetUserSum(u)
		switch v:=val.(type){
			case error:
				c.JSON(http.StatusNotFound,gin.H{"msg":v.Error()})
				return
			case string:
				c.JSON(http.StatusOK,gin.H{"sum":v})
				return
		}
		return

	})
	Router.GET("user/order",manageFunc,func(c *gin.Context){
		u := c.Query("userid")
		if u == "" {
			return
		}
		cou,err := strconv.Atoi(c.DefaultQuery("count","20"))
		if err != nil {
			//panic(err)
			return
		}
		var li []interface{}
		err = shopping.OrderListWithUser(c.Query("numid"),u,func(db interface{})error{
			li = append(li,db)
			cou--
			fmt.Println(cou)
			if cou == 0 {
				return io.EOF
			}
			return nil
		})
		if err != nil {
			//panic(err)
			if err != io.EOF {
				fmt.Println(err)
				return
			}
		}
		c.JSON(http.StatusOK,li)
		return
	})
	go Router.Run(*Port)
}

