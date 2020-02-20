package main
import(
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	"sort"
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

	v1.GET("test/:py",func(c *gin.Context){
		obj := ShoppingMap[c.Param("py")]
		if obj == nil {
			c.JSON(http.StatusOK,gin.H{"msg":""})
			return
		}
		order := c.Query("order")
		if order == "" {
			c.JSON(http.StatusOK,gin.H{"msg":""})
			return
		}
		c.JSON(http.StatusOK,obj.OrderSearch(order))
		return
	})
	v1.GET("delsite/:py",func(c *gin.Context){
		err := openSiteDB(*siteDB,func(db *bolt.DB)error{
			return db.Update(func(t *bolt.Tx)error{
				b := t.Bucket(SiteList)
				if b == nil {
					return fmt.Errorf("b == nil")
				}
				return b.Delete([]byte(c.Param("py")))
			})
		})
		c.JSON(http.StatusOK,gin.H{"msg":err.Error()})
	})
	v1.GET("updatesite/:py",func(c *gin.Context){
		sh := ShoppingInfo{
			Py:c.Param("py"),
		}
		err := openSiteDB(*siteDB,func(db *bolt.DB)error{
			sh.Load(db)
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
	v1.GET("/shopping",func(c *gin.Context){
		var li []interface{}
		err := ReadShoppingList(*siteDB,func(sh *ShoppingInfo)error{
			li = append(li,sh)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			return
		}
		c.JSON(http.StatusOK,li)
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
	if signature != Sha1([]byte(strings.Join(qu,""))) {
		c.Abort()
		return
	}
	c.Next()
}
