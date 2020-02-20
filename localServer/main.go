package main
import(
	"fmt"
	"github.com/zaddone/studySystem/request"
	"github.com/gin-gonic/gin"
	"net/url"
	"time"
	"sort"
	"encoding/hex"
	"crypto/sha1"
	"strings"
	"io"
	"net/http"
)

var(
	WXtoken = "zhaoweijie2020"
	Router = gin.Default()
	Remote = "https://www.zaddone.com/v1/"
)

func Sha1(data []byte) string {
	sha1 := sha1.New()
	sha1.Write(data)
	return hex.EncodeToString(sha1.Sum([]byte(nil)))
}
func addSign(u *url.Values){
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	li := []string{WXtoken}
	for _,v := range *u{
		li = append(li,v...)
	}
	sort.Strings(li)
	u.Add("sign",Sha1([]byte(strings.Join(li,""))))
}

func init(){
	Router.Any("updatesite/:py",func(c *gin.Context){
		url_ := c.Request.URL.Query()
		addSign(&url_)
		//c.JSON(http.StatusOK,gin.H{"msg":err.Error(),"content":sh})

		err := request.ClientHttp__(Remote+"updatesite/"+c.Param("py")+"?"+url_.Encode(),"GET",nil,nil,func(body io.Reader,res *http.Response)error{
			c.DataFromReader(res.StatusCode,res.ContentLength,res.Header.Get("content-type"),res.Body,nil)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
		}
		return
	})
}
func main(){
	go Router.Run(":8088")
	fmt.Println("run")
	select{}

}
