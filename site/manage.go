package main
import(
	"fmt"
	"github.com/zaddone/studySystem/shopping"
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	"sort"
	"encoding/json"
	"strconv"
	"time"
	"strings"
	"net/http"
)
func init(){
	manageFunc := func() gin.HandlerFunc {
		return checkManage
	}()
	v1 := Router.Group("/v1",manageFunc)
	v1.GET("/shopping",func(c *gin.Context){
		var li []interface{}
		shopping.ShoppingMap.Range(func(k,v interface{})bool{
			li = append(li,v.(shopping.ShoppingInterface).GetInfo())
			return true
		})
		c.JSON(http.StatusOK,li)
	})
	v1.GET("shopping/:py",func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh == nil {
			c.JSON(http.StatusNotFound,nil)
			return
		}
		c.JSON(http.StatusOK,sh.(shopping.ShoppingInterface).GetInfo())
		return
	})
	v1.GET("delsite/:py",func(c *gin.Context){
		py := c.Param("py")
		shopping.ShoppingMap.Delete(py)
		err := shopping.OpenSiteDB(*siteDB,func(db *bolt.DB)error{
			return db.Update(func(t *bolt.Tx)error{
				b := t.Bucket(shopping.SiteList)
				if b == nil {
					return fmt.Errorf("b == nil")
				}
				return b.Delete([]byte(py))
			})
		})
		c.JSON(http.StatusOK,gin.H{"msg":err.Error()})
	})
	v1.GET("updatesite/:py",func(c *gin.Context){
		py := c.Param("py")
		sh_,_ := shopping.ShoppingMap.Load(py)
		var sh *shopping.ShoppingInfo
		if sh_ == nil{
			sh = &shopping.ShoppingInfo{
				Py:py,
			}
		}else{
			sh = sh_.(shopping.ShoppingInterface).GetInfo()
		}
		err := shopping.OpenSiteDB(*siteDB,func(db *bolt.DB)error{
			//sh.Load(db)
			sh.Name = c.DefaultQuery("name",sh.Name)
			sh.Img = c.DefaultQuery("img",sh.Img)
			sh.Uri = c.DefaultQuery("uri",sh.Uri)
			sh.Client_id = c.DefaultQuery("clientid",sh.Client_id)
			sh.Client_secret = c.DefaultQuery("clientsecret",sh.Client_secret)
			sh.Token = c.DefaultQuery("token",sh.Token)
			return sh.SaveToDB(db)

		})
		if err == nil {
			err = fmt.Errorf("success")
		}
		c.JSON(http.StatusOK,gin.H{"msg":err.Error(),"content":sh})
	})
	v1.POST("updateorder/:py",func(c *gin.Context){
		sh_,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh_ == nil {
			c.JSON(http.StatusNotFound,nil)
			return
		}
		sh := sh_.(shopping.ShoppingInterface)
		orderid := c.Query("orderid")
		if orderid == ""{
			c.JSON(http.StatusNotFound,gin.H{"msg":"orderid error"})
			return
		}
		var db interface{}
		err := json.NewDecoder(c.Request.Body).Decode(&db)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			return
		}
		err = sh.OrderUpdate(orderid,db)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			return
		}
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
		return


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
		//fmt.Println("stamp",err)
		c.Abort()
		return
	}
	d := time.Now().Unix() - stamp
	if d<0 {
		d=-d
	}
	if d>6 {
		//fmt.Println("stamp",d)
		c.Abort()
		return
	}
	signature := c.Query("sign")
	if signature == "" {
		c.Abort()
		return
	}
	qu := []string{WXtoken}
	for k,v := range c.Request.URL.Query(){
		if k == "sign" {
			continue
		}
		qu = append(qu,v...)
		//fmt.Println(k,v)
	}
	sort.Strings(qu)
	//li_ := Sha1([]byte(strings.Join(qu,"")))
	if signature != shopping.Sha1([]byte(strings.Join(qu,""))) {
		c.Abort()
		return
	}
	c.Next()
}
