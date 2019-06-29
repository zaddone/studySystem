package ttpage
import(
	"fmt"
	"io"
	"html"
	//"bufio"
	//"strings"
	"regexp"
	"os/exec"
	"io/ioutil"
	"net/http"
	"net/url"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/config"
	"github.com/lunny/html2md"
	"runtime"
)
var (
	commands = map[string]string{
		"windows": "explorer",
		"darwin":  "open",
		"linux":   "xdg-open",
	}
	rooturl string = "https://www.toutiao.com"
	Port string = ":80"
)
func Open(uri string) error {

	run,ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}
	cmd := exec.Command(run,"http://127.0.0.1"+Port+uri)
	return cmd.Start()
}
func extract(body []byte) {

	//ReS, err := regexp.Compile("articleInfo: \\{[\\S\\s]+?\\}")
	//if err != nil {
	//	panic(err)
	//}
	ReTitle, err := regexp.Compile("title: \\'[\\S\\s]+?\\'")
	if err != nil {
		panic(err)
	}
	ReC, err := regexp.Compile("content: \\'[\\S\\s]+?\\'")
	if err != nil {
		panic(err)
	}
	//db := ReS.Find(body)
	//if len(db)==0 {
	//	return
	//}
	loc := ReTitle.FindIndex(body)
	if len(loc)==0 {
		//fmt.Println(string(body))
		return
	}
	title := string(body[loc[0]+8:loc[1]-1])
	fmt.Println(title)
	loc_ := ReC.FindIndex(body)
	if len(loc_)==0 {
		//fmt.Println(title)
		fmt.Println(string(body))
		return
	}
	content:=html2md.Convert(html.UnescapeString(string(body[loc_[0]+10:loc_[1]-1])))
	fmt.Println(content)
	//fmt.Println(string(db[loc[0]+8:loc[1]-1]))

}
func getPageUrl(val map[string]interface{},header http.Header) error {
	//val:=make(map[string]interface{})
	//err := json.NewDecoder(body).Decode(&val)
	//if err != nil {
	//	return err
	//}
	for _,v := range val["data"].([]interface{}){
		err := request.ClientHttp_(rooturl + v.(map[string]interface{})["source_url"].(string),
		"GET",
		nil,
		config.Conf.Header,
		func(body io.Reader,st int)error {
			db,err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			if st != 200 {
				return fmt.Errorf("%d %s",st,db)
			}
			extract(db)
			//fmt.Println(string(db))
			return nil
		})
		if err != nil {
			fmt.Println(err)
			continue
		}


		//_v["title"]
		//_v["source_url"]
	}
	return nil
}
func init(){
	//gin.SetMode(gin.ReleaseMode)
	Router := gin.Default()

	Router.GET("/api/pc/feed/",func(c *gin.Context){
		err := request.ClientHttp_(rooturl+"/api/pc/feed/?" + url.Values{
			"category":c.QueryArray("category"),
			"utm_source":c.QueryArray("utm_source"),
			"widen":c.QueryArray("widen"),
			"max_behot_time":c.QueryArray("max_behot_time"),
			"max_behot_time_tmp":c.QueryArray("max_behot_time_tmp"),
			"tadrequire":c.QueryArray("tadrequire"),
			"as":c.QueryArray("as"),
			"cp":c.QueryArray("cp"),
			"_signature":c.QueryArray("_signature"),

		}.Encode(),"GET",nil,c.Request.Header,func(body io.Reader,st int)error{
			if st != 200 {
				return fmt.Errorf("%d",st)
			}

			val:=make(map[string]interface{})
			err := json.NewDecoder(body).Decode(&val)
			if err != nil {
				return err
			}
			c.JSON(http.StatusOK,val)

			//c.Data(http.StatusOK,"application/json",nil)
			return getPageUrl(val,c.Request.Header)
			//db,err := ioutil.ReadAll(body)
			//if err != nil {
			//	return err
			//}
			//if st != 200 {
			//	return fmt.Errorf("%d %s",st,string(db))
			//}
			//return nil
		})
		if err != nil {
			c.String(http.StatusNotFound,fmt.Sprintln(err))
			return
		}

		//c.Data(http.StatusOK,"text/html; charset=utf-8",db)
		return
	})
	Router.GET("/ch/:cn",func(c *gin.Context){
		fmt.Println("ch start")
		err := request.ClientHttp(fmt.Sprintf("https://www.toutiao.com/ch/%s/",c.Param("cn")),"GET",[]int{200},nil,func(body io.Reader)error{
			db,err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			c.Data(http.StatusOK,"text/html; charset=utf-8",db)
			return nil
		})
		if err != nil {
			c.String(http.StatusNotFound,fmt.Sprintln(err))
			return
		}
		return
	})
	go Router.Run(":80")
	go Router.RunTLS(":443","server.crt","server.key")
}
