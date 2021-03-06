package main
import(
	"fmt"
	"github.com/zaddone/studySystem/shopping"
	"github.com/zaddone/studySystem/config"
	//"github.com/zaddone/studySystem/article"
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	//"encoding/binary"
	"io"
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
	v1 := Router.Group("/v2",manageFunc)
	v1.POST("/goods/update",func(c *gin.Context){
		id := c.Query("id")
		if id == "" {
			c.JSON(http.StatusFound,fmt.Errorf("id = nil"))
			return
		}
		var db interface{}
		err := json.NewDecoder(c.Request.Body).Decode(&db)
		if err != nil {
			c.JSON(http.StatusFound,err)
			return
		}
		err = shopping.AlibabaShopping.SaveProduct(id,db)
		if err != nil {
			c.JSON(http.StatusFound,err)
			return
		}
		c.JSON(http.StatusOK,db)
	})
	v1.GET("/goods/update/list_t",func(c *gin.Context){
		sum,err :=strconv.Atoi(c.DefaultQuery("con","20"))
		if err != nil {
			c.JSON(http.StatusFound,err)
			return
		}
		var li []interface{}
		err = shopping.AlibabaShopping.HandGoodsListT(
			c.Query("goodsid"),
			("false" != c.DefaultQuery("show","false")),
			func(db interface{})error{
			li = append(li,db)
			if len(li)>=sum{
				return io.EOF
			}
			return nil
		})
		if len(li)>0{
			c.JSON(http.StatusOK,li)
			return
		}
		c.JSON(http.StatusFound,err)
		return

	})
	//v1.GET("/goods/update/list",func(c *gin.Context){
	//	var li []interface{}
	//	sum,err :=strconv.Atoi(c.DefaultQuery("con","20"))
	//	if err != nil {
	//		c.JSON(http.StatusFound,err)
	//		return
	//	}
	//	//lis := c.Query("goodsids")
	//	//if len(lis) == 0{
	//	//	c.JSON(http.StatusFound,fmt.Errorf("goodsids is nil"))
	//	//	return
	//	//}
	//	err = shopping.AlibabaShopping.HandGoodsList(
	//		c.Query("goodsids"),
	//		func(db interface{})error{
	//		li = append(li,db)
	//		if len(li)>=sum{
	//			return io.EOF
	//		}
	//		return nil
	//	})
	//	if len(li)>0{
	//		c.JSON(http.StatusOK,li)
	//		return
	//	}
	//	//fmt.Println(err)
	//	c.JSON(http.StatusFound,err)
	//	return
	//})
	v1.GET("/goods/update/del",func(c *gin.Context){
		id := c.Query("id")
		if id == "" {
			c.JSON(http.StatusFound,fmt.Errorf("id = nil"))
			return
		}
		err := shopping.AlibabaShopping.DelGoods(id)
		if err != nil {
			c.JSON(http.StatusFound,err)
			return
		}
		c.JSON(http.StatusOK,map[string]interface{}{
			"msg":"success",
		})
	})
	v1.POST("/goods/update/edit",func(c *gin.Context){
		id := c.Query("id")
		if id == "" {
			c.JSON(http.StatusFound,fmt.Errorf("id = nil"))
			return
		}
		fmt.Println(id)
		var db interface{}
		err := json.NewDecoder(c.Request.Body).Decode(&db)
		if err != nil {
			fmt.Println("json",err)
			c.JSON(http.StatusFound,err)
			return
		}
		err = shopping.AlibabaShopping.SaveGoods(id,db)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusFound,err)
			return
		}
		c.JSON(http.StatusOK,db)

	})
	v1.GET("/goods/get",func(c *gin.Context){
		gid := c.Query("goodsid")
		if len(gid)  == 0 {
			return
		}
		err := shopping.AlibabaShopping.GoodsGet(gid,func(db interface{}){
			c.JSON(http.StatusOK,db)
		})
		if err != nil {
			c.JSON(http.StatusFound,err)
			return
		}
	})
	v1.GET("/goods/list",func(c *gin.Context){
		var li []interface{}
		sum,err :=strconv.Atoi(c.DefaultQuery("con","20"))
		if err != nil {
			c.JSON(http.StatusFound,err)
			return
		}
		err = shopping.AlibabaShopping.GoodsShow(
			[]byte(c.Query("goodsid")),
			func(db interface{})error{
			li = append(li,db)
			if len(li)>= sum{
				return io.EOF
			}
			return nil
		})
		if len(li)>0{
			c.JSON(http.StatusOK,li)
			return
		}
		//fmt.Println(err)
		c.JSON(http.StatusFound,err)
		return
	})

	v1.GET("/wxtoken",func(c *gin.Context){
		c.JSON(http.StatusOK,gin.H{"msg":toKen})
	})
	v1.GET("/shopping",func(c *gin.Context){
		var li []interface{}
		shopping.ShoppingMap.Range(func(k,v interface{})bool{
			li = append(li,v.(shopping.ShoppingInterface).GetInfo())
			return true
		})
		//fmt.Println(li)
		c.JSON(http.StatusOK,li)
	})
	v1.GET("/shopping/:py",func(c *gin.Context){
		py := c.Param("py")
		sh,_ := shopping.ShoppingMap.Load(py)
		if sh == nil {
			err := shopping.OpenSiteDB(*siteDB,func(db *bolt.DB)error{
				return db.View(func(t *bolt.Tx)error{
					b := t.Bucket(shopping.SiteList)
					if b == nil {
						return fmt.Errorf("b == nil")
					}
					v := b.Get([]byte(py))
					if v == nil {
						return fmt.Errorf("v == nil")
					}
					Shinfo := &shopping.ShoppingInfo{}
					err := Shinfo.LoadByte(v)
					if err != nil {
						return err
					}
					c.JSON(http.StatusOK,Shinfo)
					return nil
				})
			})
			if err != nil {
				c.JSON(http.StatusNotFound,nil)
			}
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
			sh.ReToken = c.DefaultQuery("retoken",sh.ReToken)
			//sh.UpOrder = c.Query("update")
			upO := c.Query("uporder")
			if upO != "" {
				in,err := strconv.Atoi(upO)
				if err != nil {
					return err
				}
				sh.UpOrder = int64(in)
			}
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
			shopping.InitShoppingMap(*siteDB)
			err = fmt.Errorf("success")
		}
		fmt.Println(sh)
		c.JSON(http.StatusOK,gin.H{"msg":err.Error(),"content":sh})
	})
	v1.GET("/order/del",func(c *gin.Context){
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
	v1.GET("/order/list",func(c *gin.Context){
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
	//v1.GET("/order/time",func(c *gin.Context){
	//	t := c.Query("t")
	//	if t == "" {
	//		c.JSON(http.StatusNotFound,gin.H{"msg":"t error"})
	//		return
	//	}
	//	dbMap := map[string][]interface{}{}
	//	shopping.OrderWithTime([]byte(t),func(k string,db interface{}){
	//		dbMap[k] = append(dbMap[k],db)
	//	})
	//	if len(dbMap) == 0 {
	//		c.JSON(http.StatusNotFound,gin.H{"msg":"error"})
	//		return
	//	}
	//	c.JSON(http.StatusOK,dbMap)
	//})
	v1.GET("order_apply/update",func(c *gin.Context){
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
	v1.GET("order_apply",func(c *gin.Context){
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
		err := shopping.OrderApply(userid,orderid,func(db interface{}){
			c.JSON(http.StatusOK,db)
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			return
		}
		return

	})

	v1.POST("/updateorder/:py",func(c *gin.Context){
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
	v1.GET("/user/get",func(c *gin.Context){
		u:=c.Query("userid")
		if u == "" {
			return
		}
		user := shopping.User{UserId:u}
		err := user.Get()
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
			return
		}
		c.JSON(http.StatusOK,gin.H{"msg":user})
	})
	//v1.GET("user/order/del",func(c *gin.Context){
	//	u := c.Query("userid")
	//	if u == "" {
	//		return
	//	}
	//	o := c.Query("numid")
	//	if o == "" {
	//		return
	//	}
	//	
	//})
	v1.GET("user/order",func(c *gin.Context){
		u := c.Query("userid")
		if u == "" {
			return
		}
		o := c.Query("numid")
		cou,err := strconv.Atoi(c.DefaultQuery("count","20"))
		if err != nil {
			//panic(err)
			return
		}
		var li []interface{}
		err = shopping.OrderListWithUser(o,u,func(db interface{})error{
			li = append(li,db)
			cou--
			if cou == 0 {
				return io.EOF
			}
			return nil
		})
		if err != nil {
			//panic(err)
			if err != io.EOF {
				fmt.Println(err)
			}
		}
		c.JSON(http.StatusOK,li)
		return
	})
	//v1.GET("user/update",func(c *gin.Context){
	//	m := c.Query("mobile")
	//	if m == "" {
	//		return
	//	}
	//	u:=c.Query("userid")
	//	if u == "" {
	//		return
	//	}
	//	n:=c.Query("name")
	//	if n == "" {
	//		return
	//	}
	//	err := CheckPhoneCode(m,n)
	//	if err != nil {
	//		c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
	//		return
	//	}

	//	user := shopping.User{Mobile:m,UserId:u}
	//	err = user.Update()
	//	if err != nil {
	//		c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
	//		return
	//	}
	//	c.JSON(http.StatusOK,gin.H{"msg":user})
	//})

}

func checkManage(c *gin.Context){
	//fmt.Println(c.Request.PostForm)
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
	if d>60 {
		//fmt.Println("stamp",d)
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
