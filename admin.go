package main
import(
	"github.com/zaddone/studySystem/chrome"
	"github.com/zaddone/studySystem/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"encoding/binary"
	"fmt"
	"io"
	"net/url"
	"encoding/base64"

	//"encoding/json"
)

func main(){

	Router := gin.Default()
	Router.Static("/"+config.Conf.Static,"./"+config.Conf.Static)
	Router.LoadHTMLGlob(config.Conf.Templates+"/*")
	Router.GET("/",func(c *gin.Context){
		c.HTML(http.StatusOK,"index.tmpl",nil)
	})
	Router.GET("/del",func(c *gin.Context){
		id_,err := url.QueryUnescape(c.Query("id"))
		if err != nil{
			c.String(http.StatusNotFound,fmt.Sprintln(err))
			return
		}
		id,err := base64.StdEncoding.DecodeString(id_)
		if err != nil{
			c.String(http.StatusNotFound,fmt.Sprintln(err))
			return
		}
		err = chrome.DelPage(id)
		if err != nil {
			c.String(http.StatusNotFound,fmt.Sprintln(err))
			return
		}
		chrome.WXDBChan <- chrome.NewDelId(config.Conf.CollPageName,[]string{fmt.Sprintf("%d",binary.BigEndian.Uint64(id))})

		c.JSON(http.StatusOK,gin.H{"msg":"Success"})
	})
	Router.GET("/search/:key",func(c *gin.Context){
		plist := make([]*chrome.Page,0,10)
		err := chrome.SearchPage(c.Param("key"),func(p *chrome.Page){
			//p.Title += binary.BigEndian.Uint64(p.Id)
			p.Title += fmt.Sprintln(binary.BigEndian.Uint64(p.Id))
			plist = append(plist,p)
		})
		if err != nil {
			c.String(http.StatusNotFound,fmt.Sprintln(err))
			return
		}
		c.JSON(http.StatusOK,gin.H{"dblist":plist,"count":len(plist)})
	})
	Router.GET("/pagejson",func(c *gin.Context){
		c.JSON(http.StatusOK,gin.H{"msg":fmt.Sprintln(chrome.PageJsonFile())})
	})
	Router.GET("/showlist/:max",func(c *gin.Context){
		be := c.DefaultQuery("begin","")
		var err error
		var beg []byte
		if be != "" {
			//beg = make([]byte,8)
			beg,err = base64.StdEncoding.DecodeString(be)
			if err != nil {
				c.String(http.StatusNotFound,fmt.Sprintln(err))
				return
			}
			//binary.BigEndian.PutUint64(beg,binary.BigEndian.Uint64(beg))
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
