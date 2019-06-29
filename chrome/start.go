package chrome

import(
	"fmt"
	"io/ioutil"
	"os/exec"
	"io"
	"bytes"
	"strings"
	//"net/http"
	"time"
	"log"
	"encoding/json"
	"github.com/zaddone/studySystem/request"
	"github.com/gorilla/websocket"
)
var (
	port = "9222"
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

)
func init(){
	start(func(in string){
		fmt.Println(in)
		open("https://www.toutiao.com/ch/news_baby/",func(v interface{})error{
			//fmt.Println(v)
			requestId:=""
			step:=0
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
						"id":time.Now().Unix(),
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
					if strings.Contains(fmt.Sprintln(__v),"body"){
						fmt.Println(__v)
						panic(0)

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

