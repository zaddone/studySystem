package main
import(
	//"github.com/zaddone/studySystem/config"
	//"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"net/http"
	"fmt"
	"flag"
	"time"
)
var(
	Release  = flag.Bool("Release",false,"Release")
	Site  = flag.String("Site","www.zaddone.com:443","site")
	siteDB  = flag.String("db","SiteDB","db")
	//SiteDB *bolt.DB
	ShoppingMap = map[string]ShoppingInterface{}

	//Tls  = flag.Bool("TLS",false,"TLS")
)


type ShoppingInterface interface{
	GetInfo()*ShoppingInfo
	SearchGoods(...string)interface{}
	GoodsUrl(...string)interface{}
	GoodsDetail(...string)interface{}
	OrderSearch(...string)interface{}
	OutUrl(interface{}) string
	OrderMsg(interface{}) string

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
		default:
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func init(){
	flag.Parse()
	//var err error
	//SiteDB,err = bolt.Open("SiteDB",0600,nil)
	//if err != nil {
	//	panic(err)
	//}
	initShoppingMap()

	//fmt.Println(*Site)
	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			secureMiddleware := secure.New(secure.Options{
			    SSLRedirect: true,
			    SSLHost:     *Site,
			})
			err := secureMiddleware.Process(c.Writer, c.Request)
			if err != nil {
			    return
			}
			c.Next()
		}
	}()
	//fmt.Println("init")
	Router := gin.Default()
	Router.Static("/static","./static")
	//Router.Static("/","./static")
	Router.LoadHTMLGlob("./templates/*")


	if *Release{
		gin.SetMode(gin.ReleaseMode)
	}else{
		Router.GET("test/:py",func(c *gin.Context){
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

		Router.GET("delsite/:py",func(c *gin.Context){
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
		Router.GET("updatesite/:py",func(c *gin.Context){
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
				return sh.SaveToDB(db)

			})
			if err == nil {
				err = fmt.Errorf("success")
			}
			c.JSON(http.StatusOK,gin.H{"msg":err.Error(),"content":sh})
		})
		Router.GET("/shopping",func(c *gin.Context){
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
	Router.Use(secureFunc)
	{
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
			_,err := c.Cookie("session_id")
			if err != nil {
				c.SetCookie("session_id",fmt.Sprintf("%d",time.Now().UnixNano()),3600*24*365,"/","www.zaddone.com",true,true)
			}
			js:=""
			for k,v := range ShoppingMap {
				sh := v.GetInfo()
				js+=fmt.Sprintf("ShoppingMap.set('%s',{func:%sPageHtml,db:[],page:0,html: html%s,py:'%s',name:'%s'});",k,k,k,k,sh.Name)
			}
			c.Data(http.StatusOK,"application/javascript",[]byte(js))
		})
		Router.POST("/wx",handWxQuery)
		Router.GET("/wx",handWxQuery)
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
			c.JSON(http.StatusOK,sh.GoodsDetail(keyword))
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
			c.JSON(http.StatusOK,sh.GoodsUrl(keyword))
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

			session,err := c.Cookie("session_id")
			if err != nil{
				c.JSON(http.StatusNotFound,gin.H{"msg":"fond not2"})
				return
			}
			//fmt.Println(session,err)
			c.JSON(http.StatusOK,sh.SearchGoods(keyword,session))
			return
		})
	}
	//Router.LoadHTMLGlob(config.Conf.Templates+"/*")
	go Router.RunTLS(":443","./3375181_zaddone.com.pem","./3375181_zaddone.com.key")
	go Router.Run(":80")

}
func main(){
	select{}
}
