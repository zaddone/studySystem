package main
import(
	//"github.com/zaddone/studySystem/config"
	//"encoding/json"
	"github.com/gin-gonic/gin"
	//"github.com/unrolled/secure"
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
	Site  = flag.String("Site","www.zaddone.com:443","site")
	siteDB  = flag.String("db","SiteDB","db")
	//SiteDB *bolt.DB
	ShoppingMap = map[string]ShoppingInterface{}
	MapSession = sync.Map{}
	Router = gin.Default()
	Router_ = gin.Default()
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

type ShoppingInterface interface{
	GetInfo()*ShoppingInfo
	SearchGoods(...string)interface{}
	GoodsUrl(...string)interface{}
	GoodsDetail(...string)interface{}
	OrderSearch(...string)interface{}
	OutUrl(interface{}) string
	OrderMsg(interface{}) string
	ProductSearch(...string)[]interface{}
}

func checkSession(c *gin.Context){
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

func initShoppingMap(){
	err := ReadShoppingList(*siteDB,func(sh *ShoppingInfo)error{
		switch sh.Py {
		case "pinduoduo":
			ShoppingMap[sh.Py] = NewPdd(sh)
		//case "vip":
		//	ShoppingMap[sh.Py] = &Vip{Info:sh}
		case "jd":
			ShoppingMap[sh.Py] = NewJd(sh)
		case "taobao":
			ShoppingMap[sh.Py] = NewTaobao(sh)
		default:
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
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
	gin.SetMode(gin.ReleaseMode)
	go runServerClearMap()
	flag.Parse()
	initShoppingMap()
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
		for k,v := range ShoppingMap {
			sh := v.GetInfo()
			li = append(li,
			map[string]string{
				"Name":sh.Name,
				"Img":sh.Img,
				"Uri":sh.Uri,
				"py":k,
				//"script":fmt.Sprintf("{func:%sPageHtml,db:[],page:0,html: html%s,py:\"%s\",name:\"%s\"}",sh.Py,sh.Py,sh.Py,sh.Py,sh.Name),
			})
		}
		c.HTML(http.StatusOK,"index.tmpl",li)
	})
	Router.GET("/script",func(c *gin.Context){
		js:=""
		for k,v := range ShoppingMap {
			sh := v.GetInfo()
			js+=fmt.Sprintf("ShoppingMap.set('%s',{func:%sPageHtml,db:[],page:0,html: html%s,py:'%s',name:'%s'});",k,k,k,k,sh.Name)
		}
		c.Data(http.StatusOK,"application/javascript",[]byte(js))
	})


	Router.Use(secureFunc)
	{
		Router.GET("/p/:py/:id",func(c *gin.Context){
			sh := ShoppingMap[c.Param("py")]
			if sh == nil {
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
				return
			}
			//id := c.Param("id")
			u := sh.OutUrl(sh.GoodsUrl(c.Param("id")))
			if u == "" {
				c.JSON(http.StatusNotFound,gin.H{"msg":""})
			}
			c.Redirect(http.StatusMovedPermanently,u)

		})
		Router.GET("goodsid/:py",func(c *gin.Context){
			sh := ShoppingMap[c.Param("py")]
			if sh == nil {
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
				return
			}
			keyword := c.DefaultQuery("goodsid","")
			if keyword == "" {
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
				return
			}
			db := sh.GoodsDetail(keyword)
			if db == nil {
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
				return
			}
			c.JSON(http.StatusOK,db)
			return
		})
		Router.GET("goods/:py",func(c *gin.Context){
			sh := ShoppingMap[c.Param("py")]
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
			db := sh.GoodsUrl(keyword,c.DefaultQuery("ext",""))
			if db == nil {
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not"})
				return
			}
			c.JSON(http.StatusOK,db)
			return
		})
		Router.GET("search/:py",func(c *gin.Context){
			sh := ShoppingMap[c.Param("py")]
			if sh == nil {
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not1"})
				return
			}
			keyword := c.DefaultQuery("keyword","")
			if keyword == "" {
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not2"})
				return
			}
			db := sh.SearchGoods(keyword)
			if db == nil{
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not3"})
				return
			}
			c.JSON(http.StatusOK,db)
			return
		})
	}
	//Router.LoadHTMLGlob(config.Conf.Templates+"/*")


}
func main(){
	go Router.RunTLS(":443","./3375181_zaddone.com.pem","./3375181_zaddone.com.key")
	go Router_.Run(":80")
	select{}
}
