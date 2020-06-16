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
	if d>3600 {
		fmt.Println("sign is nil")
		c.Abort()
		return
	}
	signature := c.Query("sign")
	if signature == "" {
		fmt.Println("sign is nil")
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
		fmt.Println("sign is err")
		c.Abort()
		return
	}
	fmt.Println("next")
	c.Next()
}

func init(){
	flag.Parse()
	Router.POST("/callback",func(c *gin.Context){
		res,err := ioutil.ReadAll(c.Request.Body)
		fmt.Println(res,err)
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
	})
	go Router.Run(*Port)
}


func main(){
	select{}
}

