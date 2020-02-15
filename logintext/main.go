package main
import(
	//"net/http"
	"net/url"
	"log"
	"fmt"
	"github.com/zaddone/studySystem/request"
	"io"
	"io/ioutil"
	"os/exec"
	"sync"
	"bytes"
	"time"
	"strings"
	"github.com/gorilla/websocket"
	"encoding/json"
	"encoding/base64"
	"os"
)

type Hfunc func(interface{})
var (
	port = "9222"
	rooturl string = "https://login.taobao.com/member/login.jhtml?style=mini&newMini2=true&from=alimama"
	orderUrl string = "https://pub.alimama.com/openapi/param2/1/gateway.unionpub/report.getTbkOrderDetails.json?t=1581752650893&_tb_token_=f087e0e378633&jumpType=0&positionIndex=&pageNo=1&startTime=2020-01-01&endTime=2020-02-15&queryType=2&tkStatus=&memberType=&pageSize=40"
	tb_token = "_tb_token_"
	indexUrl = "https://www.alimama.com/index.htm"
	viewOrder = "https://pub.alimama.com/manage/effect/overview_orders.htm"
	chromekey = []byte{10,68,101,118,84,111,111,108,115,32}
	op =[]string{
		"--remote-debugging-port="+port,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--no-default-browser-check",
		rooturl,
	}
	Ourl = "http://127.0.0.1:"+port
	handMap = map[float64]func(float64,map[string]interface{}){}
	handleFinish = map[string]func(string,map[string]interface{}){}
	handleResponse Hfunc = nil
	Num float64 = 0
	png = "xcode.png"
	writeChan = make(chan interface{},5)
)
func InputText(str string,endHand func()){
	Num++
	handMap[Num] = func(id__ float64,req_ map[string]interface{}){
		delete(handMap,id__)
		if endHand != nil{
			endHand()
		}
		//ClickBoxModel(sendbtn,func(){
			//fmt.Println(sendbtn)
		//})

	}
	writeChan<-map[string]interface{}{
		"method":"Input.insertText",
		"id":Num,
		"params":map[string]interface{}{"text":str},
	}
}
func ShowCookies(hand func(map[string]interface{})){
	Num++
	handMap[Num] = func(id__ float64,req_ map[string]interface{}){
		delete(handMap,id__)
		if hand != nil{
			hand(req_)
		}
		//ClickBoxModel(sendbtn,func(){
			//fmt.Println(sendbtn)
		//})

	}
	writeChan<-map[string]interface{}{
		"method":"Network.getAllCookies",
		"id":Num,
		//"params":map[string]interface{}{"text":str},
	}
}
func PageNavigate(str string,hand func(map[string]interface{})){
	//Page.navigate
	Num++
	handMap[Num] = func(id__ float64,req_ map[string]interface{}){
		delete(handMap,id__)
		if hand != nil{
			hand(req_)
		}
		//ClickBoxModel(sendbtn,func(){
			//fmt.Println(sendbtn)
		//})

	}
	writeChan<-map[string]interface{}{
		"method":"Page.navigate",
		"id":Num,
		"params":map[string]interface{}{"url":str},
	}
}
func LoginSession(_db interface{}){
	getBody(_db,indexUrl,func(__id float64,result map[string]interface{}){
	//PageNavigate(viewOrder,func(res map[string]interface{}){
		ShowCookies(func(db_ map[string]interface{}){
			//fmt.Println(db)
			ourl,err := url.Parse(orderUrl)
			if err != nil {
				panic(err)
			}
			uVal,err := url.ParseQuery(ourl.RawQuery)
			if err != nil {
				panic(err)
			}
			for _,_c_ := range db_["cookies"].([]interface{}) {
				c_ := _c_.(map[string]interface{})
				name := c_["name"].(string)
				if strings.EqualFold(tb_token,name){
					uVal.Set(tb_token,c_["value"].(string))
					uVal.Set("t",fmt.Sprintf("%d",time.Now().Unix()))
					//fmt.Println(c_)
					break
				}
			}
			qu := fmt.Sprintf("%s://%s%s?%s",ourl.Scheme,ourl.Host,ourl.Path,uVal.Encode())
			PageNavigate(qu,func(res map[string]interface{}){
				fmt.Println(res)
			})
			return

		})
		//})
	})
}
func CheckLogin(_db interface{}){
	if !getBody(_db,png,func(__id float64,result map[string]interface{}){
		//fmt.Println("--------------")
		//fmt.Println(result)
		body,err := base64.StdEncoding.DecodeString(result["body"].(string))
		if err != nil {
			panic(err)
		}
		f, _ := os.OpenFile(png, os.O_RDWR|os.O_CREATE, os.ModePerm)
		f.Write(body)
		f.Close()
		fmt.Println("http://127.0.0.1"+":8001/"+png)
		//fmt.Println(result["body"].(string))
		handleResponse = LoginSession
		return
	}){
		LoginSession(_db)
	}

}

func getBody(_db interface{},uri_ string, bodyMap func(float64,map[string]interface{})) bool {
	u := _db.(map[string]interface{})
	_u := u["response"].(map[string]interface{})
	_uri := _u["url"].(string)
	if !strings.Contains(_uri,uri_){
		return false
	}
	rid := u["requestId"].(string)
	//fmt.Println(_uri,rid)
	handleFinish[rid] =func(id_ string ,db map[string]interface{}){
		//id := float64(time.Now().Unix())
		delete(handleFinish,id_)
		Num++
		handMap[Num] = func(__id float64,__db map[string]interface{}){
			delete(handMap,__id)
			//fmt.Println(__db)
			bodyMap(__id,__db)
		}
		writeChan<-map[string]interface{}{
			"method":"Network.getResponseBody",
			"id":Num,
			"params":map[string]interface{}{"requestId":id_},
		}
	}
	return true

}
func runStream(u string,w *sync.WaitGroup,hand func(interface{},*websocket.Conn)){
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	err = c.WriteJSON(map[string]interface{}{"method":"Network.enable","id":1})
	if err != nil {
		log.Fatal("w:", err)
	}
	err = c.WriteJSON(map[string]interface{}{"method":"Network.clearBrowserCookies","id":1})
	if err != nil {
		log.Fatal("w:", err)
	}
	stop := make(chan bool)

	w.Add(2)
	go func(){
		defer w.Done()
		var db interface{}
		for{
			err = c.ReadJSON(&db)
			if err != nil {
				close(stop)
				log.Println("stream",err)
				break
			}
			c.SetReadDeadline(time.Now().Add(time.Minute*2))
			hand(db,c)
		}
	}()
	go func(){
		defer w.Done()
		for{
			select{
			case w:= <-writeChan:
				if w == nil {
					log.Println("stream w")
					return
				}
				err := c.WriteJSON(w)
				if err != nil {
					fmt.Println(err)
					return
				}

			case <-stop:
				log.Println("stop stream w")
				return
			}
		}
	}()
	return
}

func main(){
	start(func(u string)error{
		w := new(sync.WaitGroup)
		handleResponse = CheckLogin
		time.Sleep(100*time.Millisecond)
		return openPage_(func(db interface{})error{
			_vb := db.(map[string]interface{})
			time.Sleep(100*time.Millisecond)
			runStream(_vb["webSocketDebuggerUrl"].(string),w,func(db interface{},c *websocket.Conn){
				__v := db.(map[string]interface{})
				id__ := __v["id"]
				if id__ != nil{
					_id__ := id__.(float64)
					h := handMap[_id__]
					if h != nil {
						//fmt.Println(db)
						db_ := db.(map[string]interface{})
						if db_["result"] == nil {
							go h(_id__,db_)
						}else{
							go h(_id__,(db_["result"]).(map[string]interface{}))
						}
					}

					return
				}
				switch __v["method"]{
				case "Network.responseReceived":
					if handleResponse == nil {
						return
					}
					handleResponse(__v["params"])
				case "Network.loadingFinished":
					u:= __v["params"].(map[string]interface{})
					rid := u["requestId"].(string)
					hand := handleFinish[rid]
					if hand == nil {
						return
					}
					hand(rid,u)
				default:
				}

			})
			w.Wait()
			closePage(_vb["id"].(string))
			return io.EOF
		})
	})
}
func closePage(id string){
	log.Println("close",id)
	err := request.ClientHttp_(Ourl+"/json/close/"+id,"GET",nil,nil,func(body io.Reader,st int)error {
		if st != 200 {
			db,err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			return fmt.Errorf("%d %s",st,db)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func openPage_(hand func(interface{})error) error {
	return request.ClientHttp_(Ourl+"/json","GET",nil,nil,func(body io.Reader,st int)error {
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
		//fmt.Println(k)
		for _,v := range k.([]interface{}){
			er := hand(v)
			if er != nil {
				panic(er)
			}
		}
		//close(MsgPool)
		panic(0)
		return nil

	})
}

func start(hand func(string)error){
	runout := func(r io.Reader){
		var db [8192]byte
		for{
			n,err := r.Read(db[:])
			if err != nil {
				if err != io.EOF{
					log.Println(err)
					return
				}
			}
			//fmt.Println(string(db[:n]))
			//continue
			if bytes.HasPrefix(db[:n],chromekey){
				err = hand(string(db[23:n-1]))
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
	for{
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
		cmd.Wait()
		out.Close()
		outerr.Close()
		log.Println("cmd end")
		err = exec.Command("pkill","chrome").Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
func GetDoc(h func(map[string]interface{})){
	Num++
	handMap[Num] = func(_id float64,_db map[string]interface{}){
		delete(handMap,_id)
		h(_db)
	}
	writeChan<-map[string]interface{}{
		"method":"DOM.getDocument",
		"id":Num,
		"params":map[string]interface{}{"depth":-1},
	}
}
func findNodeValue(val string,root map[string]interface{},hand func(map[string]interface{})bool){
	toChrldren(root,func(db map[string]interface{}) bool {
		nodeValue := db["nodeValue"]
		if nodeValue == nil {
			return true
		}
		if strings.EqualFold(nodeValue.(string),val){
			return hand(db)
		}
		return true
	})
}

func findAttributes(userName string,root map[string]interface{},hand func(map[string]interface{})) {
	toChrldren(root,func(db map[string]interface{}) bool {
		attr := db["attributes"]
		if attr == nil {
			return true
		}
		for _,d := range attr.([]interface{}){
			switch c := d.(type){
			case string:
				if strings.EqualFold(userName,c){
					hand(db)
					return false
				}
			default:
				continue
			}
		}
		return true
	})
}
func toChrldren(node map[string]interface{},hand func(map[string]interface{})bool)bool{
	cnode := node["children"]
	if cnode == nil {
		return true
	}
	for _,d := range cnode.([]interface{}){
		d_ := d.(map[string]interface{})
		if !hand(d_){
			return false
		}
		if !toChrldren(d_,hand){
			return false
		}
	}
	return true
}
func ClickBoxModel(node map[string]interface{},hand func()){
	Num++
	handMap[Num] = func(__id_ float64,result map[string]interface{}){
		delete(handMap,__id_)
		//fmt.Println(result)
		xy := ((result["quads"].([]interface{}))[0]).([]interface{})
		Mx :=xy[0].(float64) + (xy[2].(float64)-xy[0].(float64))/2
		My :=xy[1].(float64) + (xy[7].(float64)-xy[1].(float64))/2
		//fmt.Println(xy,Mx,My)
		handMap[Num] = func(__id float64,__db map[string]interface{}){
			//fmt.Println("released",__db)
			delete(handMap,__id)
			handMap[__id] = func(_id float64,_db map[string]interface{}){
				//fmt.Println("over",_db)
				delete(handMap,_id)
				hand()
			}
			//time.Sleep(time.Millisecond*100)
			//fmt.Println(Mx,My)
			writeChan<-map[string]interface{}{
				"method":"Input.dispatchMouseEvent",
				"id":__id,
				"params":map[string]interface{}{
					"type":"mouseReleased",
					"x":Mx,
					"y":My,
					"button":"left",
					"buttons":1,
					"clickCount":1,
				},
			}
		}
		writeChan<-map[string]interface{}{
			"method":"Input.dispatchMouseEvent",
			"id":Num,
			"params":map[string]interface{}{
				"type":"mousePressed",
				"x":Mx,
				"y":My,
				"button":"left",
				"buttons":1,
				"clickCount":1,
			},
		}

	}
	//writeChan<-map[string]interface{}{
	//	"method":"DOM.focus",
	//	"id":0,
	//	"params":map[string]interface{}{"nodeId":node["nodeId"]},
	//}
	writeChan<-map[string]interface{}{
		"method":"DOM.getContentQuads",
		"id":Num,
		"params":map[string]interface{}{"nodeId":node["nodeId"]},
	}
}
