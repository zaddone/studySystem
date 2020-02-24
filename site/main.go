package main
import(
	"github.com/zaddone/studySystem/shopping"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	//"compress/gzip"
	"net/http"
	"strings"
	"strconv"
	"fmt"
	"flag"
	"time"
	"sync"
)
var(
	//Release  = flag.Bool("Release",false,"Release")
	//Site  = flag.String("Site","www.zaddone.com:443","site")
	//siteDB  = flag.String("db","SiteDB","db")
	//SiteDB *bolt.DB
	//ShoppingMap = map[string]ShoppingInterface{}
	OrderDB = "order.db"
	MapSession = sync.Map{}
	Router = gin.Default()
	Router_ = gin.Default()
	siteDB  = flag.String("db","SiteDB","db")
	SessionId = "session_id"
	//UpdateMap = time.Now()
	Html = []byte(`
<!doctype html>
<html lang="zh" class="h-100">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <meta name="keywords" content="zaddone,米果报,米果,推荐,网购,查价,优惠卷,省钱,链接,交换">
    <meta name="description" content="zaddone.com,米果报,米果推荐,网购查价,优惠卷省钱,链接交换">
    <meta name="author" content="zaddone, 米果报">
    <meta name="generator" content="Jekyll v3.8.6">
    <title>zaddone米果报</title>
    </head>
<body><script>
window.location.replace("https://www.zaddone.com")
</script>
</body>
</html>
    `)
)

func runServerClearMap(){
	for{
		time.Sleep(time.Hour*1)
		MapSession = sync.Map{}
	}
}

func checkSession(c *gin.Context){
	//s,err := c.Cookie(SessionId)
	//if err != nil {
	//	c.Abort()
	//	return
	//}
	ip := IpStrToByte(c.Request.RemoteAddr)
	if ip == nil {
		c.Abort()
		return
	}
	s := string(ip)
	v,ok := MapSession.Load(s)
	now := time.Now().Unix()
	MapSession.Store(s,now)
	if !ok {
		c.Next()
		return
	}
	if (now - v.(int64)) < 3{
		c.Abort()
		return
	}
	c.Next()
	return
}

func ClearSessionMap(t time.Time){
	MapSession.Range(func(k,v interface{})bool{
		if (t.Unix() - v.(int64))>86400 {
			MapSession.Delete(k)
		}
		return true
	})
}
func IpStrToByte(s string) []byte {
	ips := strings.Split(s,":")
	if len(ips) !=2 {
		return nil
	}
	var ipaddr [4]byte
	for i,p := range strings.Split(ips[0],"."){
		n,err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		ipaddr[i] = byte(n)
	}
	return ipaddr[:]
}

func init(){
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
	Router.Use(gzip.Gzip(gzip.DefaultCompression))
	shopping.InitShoppingMap(*siteDB)
	go runServerClearMap()
	secureFunc := func() gin.HandlerFunc {
		return checkSession
	}()

	Router.Static("/static","./static")
	Router.LoadHTMLGlob("./templates/*")
	Router_.GET("/",func(c *gin.Context){
		//c.Data(http.StatusOK,"text/html",Html)
		c.Data(200,"text/html",Html)
	})

	Router.GET("/",func(c *gin.Context){
		var li []map[string]string
		shopping.ShoppingMap.Range(func(k,v interface{})bool{
			sh := v.(shopping.ShoppingInterface).GetInfo()
			li = append(li,
			map[string]string{
				"Name":sh.Name,
				"Img":sh.Img,
				"Uri":sh.Uri,
				"py":k.(string),
			})
			return true
		})
		c.HTML(http.StatusOK,"index.tmpl",li)
	})
	Router.GET("/script",func(c *gin.Context){
		session,err := c.Cookie(SessionId)
		if err != nil {
			session = shopping.Sha1([]byte( fmt.Sprintf("%.0f%s",time.Now().UnixNano(),c.Request.RemoteAddr)))
			c.SetCookie(SessionId,session,3600*24*365,"/",".zaddone.com",false,false)
		}
		js:=""
		shopping.ShoppingMap.Range(func(k,v interface{})bool{
			sh := v.(shopping.ShoppingInterface).GetInfo()
		//for k,v := range shopping.ShoppingMap {
			//sh := v.GetInfo()
			js+=fmt.Sprintf("ShoppingMap.set('%s',{func:%sPageHtml,db:[],page:0,html: html%s,py:'%s',name:'%s'});",k,k,k,k,sh.Name)
			return true
		})
		c.Data(http.StatusOK,"application/javascript",[]byte(js))
	})

	Router.GET("/p/:py/:id",secureFunc,func(c *gin.Context){
		sh_,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh_ == nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		sh := sh_.(shopping.ShoppingInterface)
		//id := c.Param("id")
		val := []string{c.Param("id")}
		session,err := c.Cookie(SessionId)
		if err == nil{
			val = append(val,session)
		}
		u := sh.OutUrl(sh.GoodsUrl(val...))
		if u == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":""})
		}
		c.Redirect(http.StatusMovedPermanently,u)

	})
	Router.GET("goodsid/:py",secureFunc,func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		//sh := shopping.ShoppingMap[c.Param("py")]
		if sh == nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		keyword := c.DefaultQuery("goodsid","")
		if keyword == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		db := sh.(shopping.ShoppingInterface).GoodsDetail(keyword)
		if db == nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		c.JSON(http.StatusOK,db)
		return
	})
	Router.GET("goods/:py",secureFunc,func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh == nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		keyword := c.DefaultQuery("goodsid","")
		if keyword == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		//ext := c.DefaultQuery("ext","")
		db := sh.(shopping.ShoppingInterface).GoodsUrl(keyword,c.DefaultQuery("ext",""))
		if db == nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		c.JSON(http.StatusOK,db)
		return
	})
	Router.GET("search/:py",secureFunc,func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh == nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not1"})
			return
		}
		keyword := c.DefaultQuery("keyword","")
		if keyword == "" {
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not2"})
			return
		}
		session,_ := c.Cookie(SessionId)
		db := sh.(shopping.ShoppingInterface).SearchGoods(keyword,session)
		if db == nil{
			c.JSON(http.StatusNotFound,gin.H{"msg":"fond not3"})
			return
		}
		c.JSON(http.StatusOK,db)
		return
	})


}
func main(){
	go Router.RunTLS(":443","./3375181_zaddone.com.pem","./3375181_zaddone.com.key")
	go Router_.Run(":80")
	select{}
}
