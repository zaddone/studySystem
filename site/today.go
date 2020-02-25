package main
import(
	"github.com/zaddone/studySystem/request"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"github.com/PuerkitoBio/goquery"
)


type ApiHand func() (interface{},error)
var(
	todayMap = map[int]ApiHand{
		0:shanbayApi,
		1:dujinApi,
		2:icibaApi,
		3:oneApi,
	}
)
type Pop struct{
	Txt []string
	Imgurl string
}

func oneApi() (o interface{},err error) {
	uri := []byte("oneApi")
	o = checkCache(uri)
	if o != nil {
		return
	}

	err = request.ClientHttp(
		"http://www.wufazhuce.com/",
		"GET",[]int{200},nil,
		func(body io.Reader)error{
			doc,err := goquery.NewDocumentFromReader(body)
			if err != nil {
				return err
			}
			img,_ := doc.Find(".fp-one-imagen").Attr("src")
			txt := doc.Find(".fp-one-cita a").First().Text()
			o = &Pop{
				Txt:[]string{txt},
				Imgurl:img,
			}
			return nil
		},
	)
	return

}
func icibaApi() (o interface{},err error) {
	uri := []byte("icibaApi")
	o = checkCache(uri)
	if o != nil {
		return
	}
	err = request.ClientHttp(
		"http://open.iciba.com/dsapi/",
		"GET",[]int{200},nil,
		func(body io.Reader)error{
			var db map[string]interface{}
			d,err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			err = json.Unmarshal(d,&db)
			if err != nil {
				return err
			}
			//db["note"].(string)
			o = &Pop{
				Txt:[]string{
					db["note"].(string),
					db["content"].(string),
				},
				Imgurl: db["picture4"].(string),
			}
			saveCache(uri,o)
			return nil

		},
	)
	return

}

func dujinApi() (o interface{},err error ) {
	//https://api.dujin.org/bing/1366.php
	o = &Pop{
		Txt:[]string{},
		Imgurl: "https://api.dujin.org/bing/1366.php",
	}
	return
}

func shanbayApi() (o interface{},err error ) {
	uri := []byte("shanbay")
	o = checkCache(uri)
	if o != nil {
		return
	}
	err = request.ClientHttp(
		"https://rest.shanbay.com/api/v2/quote/quotes/today/",
		"GET",[]int{200},nil,
		func(body io.Reader)error{
			var db map[string]interface{}
			err := json.NewDecoder(body).Decode(&db)
			if err != nil {
				return err
			}
			if db["data"] == nil {
				return fmt.Errorf("data = nil")
			}
			data := db["data"].(map[string]interface{})
			o = &Pop{
				Txt:[]string{
					data["translation"].(string),
					data["content"].(string),
				},
				Imgurl: data["origin_img_urls"].([]interface{})[1].(string),
			}
			saveCache(uri,o)
			return nil

		},
	)
	return
}

func init(){
	Router.GET("/today",reqToday)
	Router.GET("/img",reqImg)
}
func getImgUrl(uri string) string{
	u := url.Values{}
	u.Add("url",uri)
	return "https://www.zaddone.com/img?"+u.Encode()
}

func reqImg(c *gin.Context){
	uri :=c.Query("url")
	if uri == "" {
		c.String(http.StatusNotFound,"uri = nil")
		return
	}
	req,err := http.NewRequest("GET",uri,nil)
	if err != nil {
		c.String(http.StatusNotFound,err.Error())
		return
	}
	Cli := http.Client{}
	res, err := Cli.Do(req)
	if err != nil {
		c.String(http.StatusNotFound,err.Error())
		return
	}
	bo,err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.String(http.StatusNotFound,err.Error())
		return
	}
	c.Data(res.StatusCode,res.Header.Get("Content-Type"),bo)
	return

}


func reqToday(c *gin.Context){
	//o,err := dujinApi()
	//if err == nil {
	//	c.JSON(http.StatusOK,o)
	//	return
	//}
	//fmt.Println(err)
	//return

	for _,hand := range todayMap{
		o,err := hand()
		if err == nil {
			//switch c := o.(type){
			//case *Pop :
			//	c.Imgurl = getImgUrl(c.Imgurl)
			//default:
			//	o.(map[string]interface{})["Imgurl"] = getImgUrl( o.(map[string]interface{})["Imgurl"].(string))
			//}
			c.JSON(http.StatusOK,o)
			return
		}else{
			fmt.Println(err)
		}

	}
	return

}
