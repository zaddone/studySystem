package main
import(
	"fmt"
	"github.com/zaddone/studySystem/shopping"
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	//"encoding/binary"
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
	v1.GET("/shopping/:py",func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh == nil {
			c.JSON(http.StatusNotFound,nil)
			return
		}
		c.JSON(http.StatusOK,sh.(shopping.ShoppingInterface).GetInfo())
		return
	})
	v1.GET("/delsite/:py",func(c *gin.Context){
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
	v1.GET("/updatesite/:py",func(c *gin.Context){
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
			up := c.Query("update")
			if up != "" {
				in,err := strconv.Atoi(up)
				if err != nil {
					return err
				}
				sh.Update = int64(in)
			}
			//sh.Update = c.DefaultQuery("update",sh.Update)
			return sh.SaveToDB(db)

		})
		if err == nil {
			err = fmt.Errorf("success")
		}
		c.JSON(http.StatusOK,gin.H{"msg":err.Error(),"content":sh})
	})
	v1.POST("/order",func(c *gin.Context){
		t := c.Query("t")
		if t != "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"t error"})
			return
		}
		dbMap := map[string][]interface{}{}
		shopping.OrderWithTime([]byte(t),func(k string,db interface{}){
			dbMap[k] = append(dbMap[k],db)
		})
		if len(dbMap) == 0 {
			c.JSON(http.StatusNotFound,gin.H{"msg":"error"})
			return
		}
		c.JSON(http.StatusOK,dbMap)


	})
	v1.POST("/order/:py",func(c *gin.Context){
		py := c.Param("py")
		sh_,_ := shopping.ShoppingMap.Load(py)
		if sh_ == nil {
			//c.JSON(http.StatusNotFound,nil)
			c.JSON(http.StatusNotFound,gin.H{"msg":"py error"+py})
			return
		}
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
		db := sh_.(shopping.ShoppingInterface).OrderSearch(orderid,userid)
		if db == nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":"order error"})
			return
		}
		c.JSON(http.StatusOK,db)
		return

	})
	v1.POST("/updateorder/:py",func(c *gin.Context){
		py := c.Param("py")
		sh_,_ := shopping.ShoppingMap.Load(py)
		if sh_ == nil {
			//c.JSON(http.StatusNotFound,nil)
			c.JSON(http.StatusNotFound,gin.H{"msg":"py error"+py})
			return
		}
		sh := sh_.(shopping.ShoppingInterface)
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
		err = sh.GetInfo().OrderUpdate(orderid,db)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			return
		}
		endTime := db["endTime"]
		//db["py"] = py
		if endTime != nil && endTime.(float64) !=0{
			err = sh.GetInfo().OrderUpdateTime(
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
}
//func SaveOrder(py,orderid string,endTime uint64)error{
//	end := make([]byte,8)
//	binary.BigEndian.PutUint64(end,endTime)
//	return shopping.OpenSiteDB(OrderDB,func(DB *bolt.DB)error{
//		return DB.Batch(func(t *bolt.Tx)error{
//			b,err := t.CreateBucketIfNotExists(end)
//			if err != nil {
//				return err
//			}
//			b_,err := b.CreateBucketIfNotExists([]byte(py))
//			if err != nil {
//				return err
//			}
//			return b_.Put([]byte(orderid),[]byte{'0'})
//		})
//
//	})
//
//}
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
