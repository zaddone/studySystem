package main
import(
	"github.com/zaddone/studySystem/config"
	"github.com/zaddone/studySystem/shopping"
	//"encoding/json"
	"sort"
	"strings"
	"fmt"
	"flag"
	"time"
	"io/ioutil"
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	"net/http"
	"strconv"
)

var (
	Port = flag.String("p",":8083","port")
	Router = gin.Default()
	stocksDB = flag.String("stocks", "stocks.db","stocks")
	stockBucket = []byte("stocks")
)

func openSiteDB(h func(*bolt.DB)error)error{
	db ,err := bolt.Open(*stocksDB,0600,nil)
	if err != nil {
		return err
	}
	//fmt.Println("open",dbname)
	defer func(){
		err := db.Close()
		if err != nil {
			panic(err)
		}
		//fmt.Println("close",dbname)
	}()
	return h(db)
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
	fmt.Println("next")
	c.Next()
}

func init(){
	flag.Parse()
	manageFunc := func() gin.HandlerFunc {
		return checkManage
	}()
	Router.POST("/callback",func(c *gin.Context){
		res,err := ioutil.ReadAll(c.Request.Body)
		fmt.Println(res,err)
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
	})
	Router.GET("/coupon/test",func(c *gin.Context){
		amount,err :=strconv.Atoi(c.Query("amount"))
		if err != nil {
			return
		}
		err = couponCreate(amount,func(db interface{})error{
			_db := db.(map[string]interface{})
			_res := _db["req"].(map[string]interface{})
			stock_id := _res["stock_id"].(string)
			err := couponOpen(stock_id)
			if err != nil {
				return err
			}
			c.JSON(http.StatusOK,db)
			//request_no := _db["out_request_no"].(string)
			//return couponGet(stock_id,userid,appid,request_no,amount)
			return nil
		})
		if err != nil {
			//c.JSON(http.StatusOK,err)
			return
		}
		//c.JSON(http.StatusOK,gin.H{"msg":"success"})
	})
	Router.GET("/coupon/get",manageFunc,func(c *gin.Context){
		amount,err :=strconv.Atoi(c.Query("amount"))
		if err != nil {
			return
		}
		userid := c.Query("userid")
		if userid == "" {
			return
		}
		appid := c.Query("appid")
		if appid == "" {
			return
		}
		err = couponCreate(amount,func(db interface{})error{
			_db := db.(map[string]interface{})
			_res := _db["req"].(map[string]interface{})
			stock_id := _res["stock_id"].(string)
			err := couponOpen(stock_id)
			if err != nil {
				return err
			}
			request_no := _db["out_request_no"].(string)
			return couponGet(stock_id,userid,appid,request_no,amount)
		})
		if err != nil {
			c.JSON(http.StatusOK,gin.H{"msg":err.Error()})
			return
		}
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
	})
	go Router.Run(*Port)
}


func main(){
	select{}
}

