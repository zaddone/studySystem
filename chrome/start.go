package chrome

import(
	"fmt"
	"io/ioutil"
	"os/exec"
	"io"
	"bytes"
	"strings"
	"regexp"
	"time"
	"log"
	"html"
	"encoding/json"
	"github.com/lunny/html2md"
	"github.com/zaddone/studySystem/request"
	"github.com/gorilla/websocket"
	"github.com/zaddone/studySystem/config"
)
var (
	port = "9222"

	rooturl string = "https://www.toutiao.com"
	op =[]string{
		"--remote-debugging-port="+port,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--no-default-browser-check",
		//"https://www.toutiao.com/ch/news_baby/",
	}
	Ourl = "http://127.0.0.1:"+port
	k = []byte{10,68,101,118,84,111,111,108,115,32}
	retitle *regexp.Regexp
	rec *regexp.Regexp

)
func init(){
	var err error
	retitle, err = regexp.Compile(`title: \'[\s\s]+?\'`)
	if err != nil {
		panic(err)
	}
	rec, err = regexp.Compile(`content: \'[\s\s]+?\'`)
	if err != nil {
		panic(err)
	}
	start(func(in string){
		fmt.Println(in)
		open("https://www.toutiao.com/ch/news_baby/",func(v interface{})error{
			//fmt.Println(v)
			requestId:=""
			step:=0
			id:=float64(time.Now().Unix())
			runStream((v.(map[string]interface{}))["webSocketDebuggerUrl"].(string),func(_v interface{},c *websocket.Conn){
				__v :=_v.(map[string]interface{})
				if step == 0{
					//fmt.Println(__v)
					if __v["method"] !="Network.responseReceived"{
						return
					}
					u := (__v["params"].(map[string]interface{}))
					_u := u["response"].(map[string]interface{})
					uri:= _u["url"].(string)
					if strings.Contains(uri,"/api/pc/feed/"){
						//u["RequestId"].(string)
						requestId = u["requestId"].(string)
						fmt.Println(uri,requestId)
						//panic(0)
						step = 1
					}
					return
				}else if step ==1{
					if __v["method"] !="Network.loadingFinished"{
						return
					}
					if !strings.EqualFold((__v["params"].(map[string]interface{}))["requestId"].(string),requestId){
						return
					}
					fmt.Println(__v)
					//err := c.WriteJSON(map[string]interface{}{"method":"Network.disable","id":99})
					//if err != nil {
					//	log.Fatal("w:", err)
					//}

					err := c.WriteJSON(map[string]interface{}{
						"method":"Network.getResponseBody",
						"id":id,
						"params":map[string]interface{}{"requestId":requestId},
					})
					if err != nil {
						log.Fatal("w:", err)
					}
					step = 2
					return
				}else if step ==2 {
					//if (__v["params"].(map[string]interface{}))["body"] == nil{
					//	return
					//}
					//fmt.Println(__v)
					//if strings.Contains(fmt.Sprintln(__v),"body"){
					//	fmt.Println(__v)
					//}
					//return
					_id := __v["id"]
					if _id==nil {
						return
					}
					if id != __v["id"].(float64){
						return
					}
					//fmt.Println(__v)
					body := __v["result"].(map[string]interface{})["body"].(string)
					body_:= map[string]interface{}{}
					json.Unmarshal([]byte(body),&body_)
					for _,d := range body_["data"].([]interface{}){
						if err := extract(rooturl + d.(map[string]interface{})["source_url"].(string)); err != nil {
							fmt.Println(err)
						}
					}
					err := c.WriteJSON(map[string]interface{}{"method":"Network.disable","id":99})
					if err != nil {
						log.Fatal("w:", err)
					}


				}

				//fmt.Println(["Network.responseReceived"])
			})
			return nil
		})
	})

}
func start(hand func(string)){
	runout := func(r io.ReadCloser){
		var db [8192]byte
		for{
			n,err := r.Read(db[:])
			if err != nil {
				if err != io.EOF{
					panic(err)
				}
			}
			//fmt.Println(string(db[:n]))

			if bytes.HasPrefix(db[:n],k){

				//fmt.Println(string(db[23:n-1]))
				hand(string(db[23:n-1]))
				//websocket.Dial(string(db[23:n-1]),"","")
			}
		}
	}
	cmd := exec.Command("google-chrome",op... )
	out,err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	outerr,err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	go runout(out)
	go runout(outerr)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}
func open(u string,hand func(interface{})error){
	err := request.ClientHttp_(Ourl+"/json/new/?"+u,"GET",nil,nil,func(body io.Reader,st int)error {
		if st != 200 {
			db,err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			return fmt.Errorf("%d %s",st,db)
		}
		var k interface{}
		err := json.NewDecoder(body).Decode(&k)
		if err != nil {
			return err
		}
		return hand(k)

	})
	if err != nil {
		panic(err)
	}

}
func runStream(u string,hand func(interface{},*websocket.Conn)){
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	//DOM.enable
	err = c.WriteJSON(map[string]interface{}{"method":"Network.enable","id":1})
	if err != nil {
		log.Fatal("w:", err)
	}
	//err = c.WriteJSON(map[string]interface{}{"method":"DOM.getFlattenedDocument","id":2})
	//if err != nil {
	//	log.Fatal("w:", err)
	//}
	go func(){
		var db interface{}
		for{
			err = c.ReadJSON(&db)
			if err != nil {
				log.Println(err)
				return
			}
			hand(db,c)
			//fmt.Println(db)
		}
	}()
}

func extract(uri string) error {
	fmt.Println(uri)
	return request.ClientHttp_(uri,"GET",nil,config.Conf.Header,
	func(body io.Reader,st int)error {
		db,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		if st != 200 {
			return fmt.Errorf("%d %s",st,db)
		}
		//fmt.Println(string(db))
		_extract(db)
		//fmt.Println(string(db))
		return nil
	})

}
func _extract(body []byte) {
	//retitle, err := regexp.Compile("title: \\'[\\s\\s]+?\\'")
	//if err != nil {
	//	panic(err)
	//}
	//rec, err := regexp.Compile("content: \\'[\\s\\s]+?\\'")
	//if err != nil {
	//	panic(err)
	//}
	loc := retitle.FindIndex(body)
	fmt.Println(loc)
	if len(loc)==0 {
		//fmt.println(string(body))
		return
	}
	title := string(body[loc[0]+8:loc[1]-1])
	fmt.Println(title)
	loc_ := rec.FindIndex(body)
	if len(loc_)==0 {
		//fmt.println(title)
		fmt.Println(string(body))
		return
	}
	content:=html2md.Convert(html.UnescapeString(string(body[loc_[0]+10:loc_[1]-1])))
	fmt.Println(content)
	//fmt.println(string(db[loc[0]+8:loc[1]-1]))

}

