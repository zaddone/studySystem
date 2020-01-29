package main
import(
	//"github.com/zaddone/studySystem/config"
	//"encoding/json"
	//"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	//"github.com/unrolled/secure"
	//"net/http"
	//"fmt"
	"flag"
)
var(
	//Release  = flag.Bool("Release",false,"Release")
	//Site  = flag.String("Site","www.zaddone.com:443","site")
	//SiteDB *bolt.DB
	//Tls  = flag.Bool("TLS",false,"TLS")
)

func init(){
	flag.Parse()
	//var err error
	//SiteDB,err = bolt.Open("SiteDB",0600,nil)
	//if err != nil {
	//	panic(err)
	//}

	//fmt.Println(*Site)
	//secureFunc := func() gin.HandlerFunc {
	//	return func(c *gin.Context) {
	//		secureMiddleware := secure.New(secure.Options{
	//		    SSLRedirect: true,
	//		    SSLHost:     *Site,
	//		})
	//		err := secureMiddleware.Process(c.Writer, c.Request)
	//		if err != nil {
	//		    return
	//		}
	//		c.Next()
	//	}
	//}()
	//fmt.Println("init")
	Router := gin.Default()
	Router.Static("/","./static")
	//Router.Static("/static","./static")
	//Router.LoadHTMLGlob("./templates/*")


	//if *Release{
	//	gin.SetMode(gin.ReleaseMode)
	//}else{
	//	Router.GET("delsite/:py",func(c *gin.Context){
	//		err := SiteDB.Update(func(t *bolt.Tx)error{
	//			b := t.Bucket(SiteList)
	//			if b == nil {
	//				return fmt.Errorf("b == nil")
	//			}
	//			return b.Delete([]byte(c.Param("py")))
	//		})
	//		c.JSON(http.StatusOK,gin.H{"msg":err.Error()})
	//	})
	//	Router.GET("updatesite/:py",func(c *gin.Context){
	//		sh := ShoppingInfo{
	//			Py:c.Param("py"),
	//		}
	//		err := sh.Load(SiteDB)
	//		//if err != nil {
	//		//	c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
	//		//	return
	//		//}
	//		sh.Name = c.DefaultQuery("name",sh.Name)
	//		sh.Img = c.DefaultQuery("img",sh.Img)
	//		sh.Uri = c.DefaultQuery("uri",sh.Uri)
	//		sh.Client_id = c.DefaultQuery("clientid",sh.Client_id)
	//		sh.Client_secret = c.DefaultQuery("clientsecret",sh.Client_secret)
	//		err = sh.SaveToDB(SiteDB)
	//		if err != nil {
	//			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
	//			return
	//		}
	//		if err == nil {
	//			err = fmt.Errorf("success")
	//		}
	//		c.JSON(http.StatusOK,gin.H{"msg":err.Error(),"content":sh})
	//	})
	//}

	//Router.Use(secureFunc)
	//{
	//	Router.GET("/",func(c *gin.Context){
	//		var li []map[string]string
	//		err := ReadShoppingList(SiteDB,func(sh *ShoppingInfo)error{
	//			li = append(li,
	//			map[string]string{
	//				"Name":sh.Name,
	//				"Img":sh.Img,
	//				"Uri":sh.Uri,
	//			})
	//			return nil
	//		})
	//		if err != nil {
	//			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
	//			return
	//		}
	//		c.HTML(http.StatusOK,"index.tmpl",li)
	//	})
	//	Router.GET("/shopping",func(c *gin.Context){
	//		var li []map[string]string
	//		err := ReadShoppingList(SiteDB,func(sh *ShoppingInfo)error{
	//			li = append(li,
	//			map[string]string{
	//				"Name":sh.Name,
	//				"Img":sh.Img,
	//				"Uri":sh.Uri,
	//			})
	//			return nil
	//		})
	//		if err != nil {
	//			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
	//			return
	//		}
	//		c.JSON(http.StatusOK,gin.H{"state":"success","content":li})
	//		//c.HTML(http.StatusOK,"search.tmpl",nil)
	//	})
	//}
	////Router.LoadHTMLGlob(config.Conf.Templates+"/*")
	go Router.RunTLS(":443","./3375181_zaddone.com.pem","./3375181_zaddone.com.key")
	go Router.Run(":80")

}
func main(){
	select{}
}
