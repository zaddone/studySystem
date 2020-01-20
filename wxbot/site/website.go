package main
import(
	//"github.com/zaddone/studySystem/config"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"net/http"
	//"fmt"
	"flag"
)
var(
	Release  = flag.Bool("Release",false,"Release")
	//Tls  = flag.Bool("TLS",false,"TLS")
)

func init(){
	flag.Parse()
	if *Release{
		gin.SetMode(gin.ReleaseMode)
	}
	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			secureMiddleware := secure.New(secure.Options{
			    SSLRedirect: true,
			    SSLHost:     "www.zaddone.com:443",
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
	Router.LoadHTMLGlob("./templates/*")
	Router.Use(secureFunc)
	{
		Router.GET("/",func(c *gin.Context){
			c.HTML(http.StatusOK,"index.tmpl",nil)
		})
	}
	//Router.LoadHTMLGlob(config.Conf.Templates+"/*")
	go Router.RunTLS(":443","./3375181_zaddone.com.pem","./3375181_zaddone.com.key")
	go Router.Run(":80")

}
func main(){
	select{}
}
