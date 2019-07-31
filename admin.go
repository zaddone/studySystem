package main
import(
	"github.com/zaddone/studySystem/chrome"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"encoding/binary"
	"fmt"
	"io"
	//"net/url"
	"encoding/base64"

	//"encoding/json"
)

func main(){

	Router := gin.Default()
	Router.Static("/static","./static")
	Router.LoadHTMLGlob("./templates/*")
	Router.GET("/",func(c *gin.Context){
		c.HTML(http.StatusOK,"index.tmpl",nil)
	})
	Router.GET("/search/:key",func(c *gin.Context){
		plist := make([]*chrome.Page,0,10)
		err := chrome.SearchPage(c.Param("key"),func(p *chrome.Page){
			plist = append(plist,p)
		})
		if err != nil {
			c.String(http.StatusNotFound,fmt.Sprintln(err))
		}
		c.JSON(http.StatusOK,gin.H{"dblist":plist,"count":len(plist)})
	})
	Router.GET("/showlist/:max",func(c *gin.Context){
		be := c.DefaultQuery("begin","")
		var err error
		beg := make([]byte,8)
		if be != "" {
			beg,err = base64.StdEncoding.DecodeString(be)
			if err != nil {
				c.String(http.StatusNotFound,fmt.Sprintln(err))
				return
			}
			binary.BigEndian.PutUint64(beg,binary.BigEndian.Uint64(beg)+1)
		}
		max,err := strconv.Atoi(c.Param("max"))
		if err != nil {
			//c.String(http.StatusNotFound,err)
			c.String(http.StatusNotFound,fmt.Sprintln(err))
			return
		}
		//binary.BigEndian.PutUint64(beg,uint64(begin))
		list,err := chrome.ReadPageList(beg,max)
		if err != nil && err != io.EOF {
			//c.String(http.StatusNotFound,err)
			c.String(http.StatusNotFound,fmt.Sprintln(err))
			return
		}
		c.JSON(http.StatusOK,gin.H{"dblist":list,"count":len(list)})
	})
	Router.Run(":8080")

}
