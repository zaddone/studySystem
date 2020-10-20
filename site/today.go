package main
import(
	"github.com/zaddone/studySystem/request"
	"github.com/gin-gonic/gin"
	"encoding/json"
	//"golang.org/x/image/webp"
	//"image"
	//"image/jpeg"
	//"bytes"
	"fmt"
	"io"
	//"strings"
	"io/ioutil"
	"net/http"
	"net/url"
	//"regexp"
	"github.com/PuerkitoBio/goquery"
	"sync"
)


type ApiHand func() (interface{},error)
var(
	//imgFile = regexp.MustCompile(`(png|jpeg|jpg).*`)
	todayMap = new(sync.Map)
	todayChan = make(chan ApiHand,5)
	//todayMap = map[string]ApiHand{
	//	"1":shanbayApi,
	//	"2":icibaApi,
	//	"3":oneApi,
	//	//3:dujinApi,
	//}
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

			//img,_ := doc.Find(".fp-one-imagen").Attr("src")
			txt := doc.Find(".fp-one-cita a").First().Text()
			o = &Pop{
				Txt:[]string{txt},
				Imgurl: "https://api.dujin.org/bing/1366.php",
				//Imgurl:img,
			}
			saveCache(uri,o)
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

//func dujinApi() (o interface{},err error ) {
//	//https://api.dujin.org/bing/1366.php
//	o = &Pop{
//		Txt:[]string{},
//		Imgurl: "https://api.dujin.org/bing/1366.php",
//	}
//	return
//}

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
				//Imgurl: imgFile.ReplaceAllString(data["origin_img_urls"].([]interface{})[1].(string),"$1"),
				Imgurl: data["origin_img_urls"].([]interface{})[1].(string),
			}
			//o.Imgurl = imgFile.ReplaceAllString(o.ImgUrl,"$1")
			saveCache(uri,o)
			return nil

		},
	)
	return
}

func init(){
	todayMap.Store("1",shanbayApi)
	todayMap.Store("2",icibaApi)
	todayMap.Store("3",oneApi)
	todayMap.Range(func(k,v interface{})bool{
		todayChan<-v.(func() (interface{},error))
		return true
	})
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

	//contentTpye :=strings.ToUpper(res.Header.Get("Content-Type"))
	//var body io.Reader
	//if strings.Contains(contentTpye,"WEBP"){
	//	var im image.Image
	//	im,err = webp.Decode(res.Body)
	//	if err != nil {
	//		panic(err)
	//	}
	//	var out  bytes.Buffer
	//	err = jpeg.Encode(&out,im,nil)
	//	if err != nil {
	//		panic(err)
	//	}
	//	if err != nil {
	//		c.String(http.StatusNotFound,err.Error())
	//		return
	//	}
	//	var out  bytes.Buffer
	//	err =  jpeg.Encode(&out,im,nil)
	//	if err != nil {
	//		c.String(http.StatusNotFound,err.Error())
	//		return
	//	}
		res.Header.Set("Content-Type","image/jpeg")
	//	body = &out
	//}else{
	//	body = res.Body
	//}
	//bo,err := ioutil.ReadAll(body)
	bo,err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.String(http.StatusNotFound,err.Error())
		return
	}


	c.Data(res.StatusCode,res.Header.Get("Content-Type"),bo)
	return

}


func reqToday(c *gin.Context){

	k := c.Query("pop")
	if k != "" {
		v,_ := todayMap.Load(k)
		if v == nil {
			return
		}
		o,err := (v.(ApiHand))()
		if err == nil {
			c.JSON(http.StatusOK,o)
		}else{
			fmt.Println(err)
		}
		return
	}
	for{
	select{
	case hand := <-todayChan:
		todayChan<-hand
		o,err := hand()
		if err == nil {
			c.JSON(http.StatusOK,o)
			return
		}else{
			fmt.Println(err)
		}
	default:
		return
	}
	}
	return

}
