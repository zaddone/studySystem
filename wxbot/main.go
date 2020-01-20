package main
import(
	"fmt"
	"os/exec"
	"io"
	"regexp"
	//"net/url"
	"bytes"
	"strings"
	"log"
	"sync"
	"io/ioutil"
	"encoding/json"
	"encoding/base64"
	"time"
	"github.com/zaddone/studySystem/request"
	//"github.com/zaddone/wxbot/util"
	//"github.com/zaddone/studySystem/wxmsg"
	"github.com/gorilla/websocket"

)
type Hfunc func(interface{})
var(
	port = "9222"
	rooturl string = "https://wx.qq.com/?lang=zh_CN"
	op =[]string{
		"--remote-debugging-port="+port,
		//"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--no-default-browser-check",
		rooturl,
	}
	Ourl = "http://127.0.0.1:"+port
	chromekey = []byte{10,68,101,118,84,111,111,108,115,32}

	handleFinish = map[string]func(string,map[string]interface{}){}
	handleResponse Hfunc = nil
	//handleIDReq Hfunc = nil
	handMap = map[float64]func(float64,map[string]interface{}){}
	CachePng = make(chan interface{},1)
	//MsgPool = make(chan *Msg,10000)
	Num float64 = 100
	//Mx,My float64
	NowUserName string
	DialogueMap =map[string]chan *Msg {}
	//RunDialog = make(chan string)
	writeChan = make(chan interface{},5)

	regK *regexp.Regexp = regexp.MustCompile(`[0-9a-zA-Z]+|\p{Han}`)
)

func main(){
	//fmt.Println("ok")
	start(func(u string)error{
		w := new(sync.WaitGroup)
		handleResponse = CheckWXLogin
		return openPage_(func(db interface{})error{
		//return openPage(rooturl,func(db interface{})error{
			_vb := db.(map[string]interface{})
			runStream(_vb["webSocketDebuggerUrl"].(string),w,func(db interface{},c *websocket.Conn){
				__v := db.(map[string]interface{})
				id__ := __v["id"]
				if id__ != nil{
					_id__ := id__.(float64)
					h := handMap[_id__]
					if h != nil {
						//result := (db.(map[string]interface{})["result"]).(map[string]interface{})
						go h(_id__,(db.(map[string]interface{})["result"]).(map[string]interface{}))
					}

					return
				}
				//fmt.Println(__v["method"])
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
					//fmt.Println(__v["method"],u["requestId"])
					hand(rid,u)
				default:
					//if __v["params"] == nil {
					//	return
					//}
					//u:= __v["params"].(map[string]interface{})
					//fmt.Println(__v["method"],u["requestId"])
					//if u["requestId"] == nil {
					//	return
					//}
					//rid := u["requestId"].(string)
					//hand := handleFinish[rid]
					//if hand == nil {
					//	return
					//}
					//hand(rid,u,write)
					//fmt.Println(db)
				}

			})
			w.Wait()
			closePage(_vb["id"].(string))
			return io.EOF
		})

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
func openPage(u string,hand func(interface{})error) error {

	return request.ClientHttp_(Ourl+"/json/new?"+u,"GET",nil,nil,func(body io.Reader,st int)error {
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
		er := hand(k)
		if er != nil {
			panic(er)
		}
		//close(MsgPool)
		panic(0)
		return nil

	})
	//if err != nil {
	//	log.Println(err)
	//	//panic(err)
	//}

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
	//err = c.WriteJSON(map[string]interface{}{"method":"Network.clearBrowserCookies","id":2})
	//if err != nil {
	//	log.Fatal("w:", err)
	//}

	//err = c.WriteJSON(map[string]interface{}{"method":"Page.enable","id":3})
	//if err != nil {
	//	panic(err)
	//	log.Fatal("w:", err)
	//}


	//err = c.WriteJSON(map[string]interface{}{"method":"DOM.enable","id":4})
	//if err != nil {
	//	panic(err)
	//	log.Fatal("w:", err)
	//}

	//err = c.WriteJSON(map[string]interface{}{
	//	"method":"Page.navigate",
	//	"id":5,
	//	"params":map[string]interface{}{"url":rooturl},
	//})
	//if err != nil {
	//	panic(err)
	//	log.Fatal("w:", err)
	//}

	//writeChan := make(chan interface{},5)
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
			//fmt.Println(db)
			hand(db,c)
			//fmt.Println(db)
		}
	}()
	go func(){
		//fmt.Println(u,"w")
		defer w.Done()
		for{
			select{
			case w:= <-writeChan:
				if w == nil {
					log.Println("stream w")
					return
				}
				fmt.Println(w)
				err := c.WriteJSON(w)
				if err != nil {
					fmt.Println(err)
					return
				}
				//log.Println(w)

			case <-stop:
				log.Println("stop stream w")
				return
			}
		}
	}()
	//c.Close()
	return
}


func toChrldren_(node map[string]interface{},hand func(map[string]interface{},map[string]interface{})bool)bool{
	cnode := node["children"]
	if cnode == nil {
		return true
	}
	for _,d := range cnode.([]interface{}){
		d_ := d.(map[string]interface{})
		if !hand(d_,node){
			return false
		}
		if !toChrldren_(d_,hand){
			return false
		}
	}
	return true
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

func findMsgUserNode(userName string,root map[string]interface{},hand func(map[string]interface{})) {
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

func ClickBoxModel(node map[string]interface{},hand func()){
	Num++
	handMap[Num] = func(__id_ float64,result map[string]interface{}){
		delete(handMap,__id_)
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

func handSyncMsg(_db *Msg){

	msglist := DialogueMap[_db.FromUserName]
	if msglist == nil {
		msglist = make(chan *Msg,100)
		DialogueMap[_db.FromUserName] = msglist
	}
	_db.Content = ""
	//fmt.Println(_db)
	msglist<-_db

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
func findText(node map[string]interface{},hand func(string)){
	toChrldren(node,func(node map[string]interface{})bool{
		if strings.EqualFold(node["nodeName"].(string),"#text"){
			hand(node["nodeValue"].(string))
			return false
		}
		return true
	})
}

func checkKeyNode(n map[string]interface{},k string) (bool){
	attr := n["attributes"]
	if attr == nil {
		return false
	}
	for _,d := range attr.([]interface{}){
		if !strings.Contains(d.(string),k){
			continue
		}
		return true
	}
	return false

}
func checkMsgNode(n map[string]interface{},m *Msg) (bool){
	//return checkKeyNode(n,m.MsgId)
	attr := n["attributes"]
	if attr == nil {
		return false
	}
	for _,d := range attr.([]interface{}){
		if !strings.Contains(d.(string),m.MsgId){
			continue
		}
		findText(n,func(t string){
			m.Content = t
		})
		return true
	}
	return false
}

func findDialogue(msg chan *Msg,success func(*Msg),complete func(map[string]interface{})) {
	var root map[string]interface{} = nil
	findD := func(node_r map[string]interface{},_m *Msg){
		//fmt.Println("d",_m)
		toChrldren(node_r,func(node map[string]interface{})bool{
			if checkMsgNode(node,_m){
				fmt.Println("find",_m)
				success(_m)
				//root = fnode
				return false
			}
			return true
		})
	}

	var w sync.WaitGroup
	G:
	//for {
	for m:= range msg {

		time.Sleep(100*time.Millisecond)
		w.Add(1)
		GetDoc(func(result map[string]interface{}){
			root = result["root"].(map[string]interface{})
			findD(root,m)
			w.Done()
		})
		w.Wait()
		fmt.Println("-----------")
		if m.Content == "" {
			//fmt.Println(m)
			//panic(0)
			msg<-m
			continue G
		}
		for{
			select{
			case m_:= <-msg:
				findD(root,m_)
				if m_.Content == "" {
					//fmt.Println(m)
					//panic(0)
					msg<-m_
					continue G
				}
			default:
				complete(root)
				continue G
			}
		}
	}
}
func BackMsg(uid string,root map[string]interface{}){
	//btn btn_send
	var sendbtn,editArea map[string]interface{}
	toChrldren(root,func(node map[string]interface{})bool{
		if checkKeyNode(node,"btn btn_send"){
			sendbtn = node
			return false
		}
		return true
	})
	toChrldren(root,func(node map[string]interface{})bool{
		if checkKeyNode(node,"editArea"){
			editArea = node
			return false
		}
		return true
	})
	words := make(map[string]int)
	n:=0
	GetMsgf(uid,func(m *Msg)error{
		lr := regK.FindAllString(m.Content,-1)
		for j:=0;j<len(lr);j++{
			for _j:=j+2;_j<=len(lr);_j++ {
				words[strings.Join(lr[j:_j],"") ]+=1
			}
		}
		n++
		if n>5 {
			return io.EOF
		}
		return nil
	})
	str := ""
	for k,_ := range words{
		str += k
		if len(str)>10 {
			break
		}
	}
	Num++
	handMap[Num]= func(id_ float64,req map[string]interface{}){
		delete(handMap,id_)
		Num++
		handMap[Num] = func(id__ float64,req_ map[string]interface{}){
			delete(handMap,id__)
			ClickBoxModel(sendbtn,func(){
				fmt.Println(sendbtn)
			})

		}
		writeChan<-map[string]interface{}{
			"method":"Input.insertText",
			"id":Num,
			"params":map[string]interface{}{"text":str},
		}
	}
	writeChan<-map[string]interface{}{
		"method":"DOM.focus",
		"id":Num,
		"params":map[string]interface{}{"nodeId":editArea["nodeId"]},
	}

}
func OpenDialogue(uid string,msg chan *Msg) {

	if NowUserName == uid {
		findDialogue(
			msg,
			func(m *Msg){
				if err := UpdateMsg(m); err != nil {
					panic(err)
				}
			},
			func(r map[string]interface{}){
				BackMsg(uid,r)
				delete(DialogueMap,uid)
				close(msg)
			},
		)
		return
	}

	var w sync.WaitGroup
	w.Add(1)
	GetDoc(func(result map[string]interface{}){
	findMsgUserNode(uid,result["root"].(map[string]interface{}),func(node map[string]interface{}){
	ClickBoxModel(node,func(){
		NowUserName = uid
		findDialogue(
			msg,
			func(m *Msg){
				if err := UpdateMsg(m); err != nil {
					panic(err)
				}
			},
			func(r map[string]interface{}){
				BackMsg(uid,r)
				delete(DialogueMap,uid)
				close(msg)
				w.Done()
			},
		)
		//time.Sleep(500*time.Millisecond)
		//findBackMsg()
	})
	})
	})
	w.Wait()

}


func GetWebwxsync(_db interface{}){
	getBody(_db,"https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxsync",
	func(__id float64,result map[string]interface{}){
		body :=result["body"]
		if body == nil {
			return
		}
		//fmt.Println(result)
		var _msgList MsgList
		err := json.Unmarshal([]byte(body.(string)),&_msgList)
		if err != nil {
			panic(err)
		}
		//fmt.Println(_msgList)
		for _,b_ := range _msgList.AddMsgList{
			//b_ := b.(map[string]interface{})
			//fmt.Println(b_)
			switch b_.MsgType{
			case 1:
				handSyncMsg(b_)
			default:
				//fmt.Println(b_["MsgType"])
			}
		}
	})
}

func GetContactList(_db interface{}){

	getBody(_db,"https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetcontact",
	func(__id float64,__db map[string]interface{}){
		//fmt.Println(__db)
		go HandDialogueMap()
		HandleWXContact(__db)
		handleResponse = GetWebwxsync
	})

}
func HandDialogueMap(){
	for{
		for u,m := range DialogueMap {
			OpenDialogue(u,m)
			//fmt.Println(u,len(m))
		}
		time.Sleep(time.Second)
	}
}
func HandleWXContact(result map[string]interface{}){
	body :=result["body"]
	if body == nil {
		return
	}
	var _body MemberList
	err := json.Unmarshal([]byte(body.(string)),&_body)
	if err != nil {
		panic(err)
	}

	for _,d_ := range _body.MemberList{
		NickName   := d_.NickName
		RemarkName := d_.RemarkName
		UserName := d_.UserName
		if RemarkName  == "" {
			RemarkName = NickName
		}
		AddContact(RemarkName,UserName)
	}
}
func CheckWXLogin(_db interface{}){
	if !getBody(_db,"https://login.weixin.qq.com/qrcode/",
	func(__id float64,result map[string]interface{}){
		body,err :=base64.StdEncoding.DecodeString(result["body"].(string))
		if err != nil {
			panic(err)
		}
		select{
		case <-CachePng:
			CachePng <- body
		default:
			CachePng <- body
		}
		fmt.Println("http://127.0.0.1"+":8001"+"/loginwx")
		handleResponse = GetContactList
	}){
		GetContactList(_db)
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
		Num++
		handMap[Num] = func(__id float64,__db map[string]interface{}){
			bodyMap(__id,__db)
			delete(handMap,__id)
		}
		writeChan<-map[string]interface{}{
			"method":"Network.getResponseBody",
			"id":Num,
			"params":map[string]interface{}{"requestId":id_},
		}
		delete(handleFinish,id_)
	}
	return true

}
