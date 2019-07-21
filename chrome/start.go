package chrome

import(

	"fmt"
	"sync"
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
	"encoding/binary"
	"github.com/lunny/html2md"
	"github.com/zaddone/studySystem/request"
	"github.com/gorilla/websocket"
	"github.com/zaddone/studySystem/config"
	"github.com/zaddone/studySystem/wxmsg"

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
	rea *regexp.Regexp
	rej *regexp.Regexp
	uris = config.Conf.ToutiaoUri
	//[]string{
	//	"https://www.toutiao.com/ch/news_finance/",
	//	"https://www.toutiao.com/ch/news_finance/",
	//	"https://www.toutiao.com/ch/news_baby/",
	//	"https://www.toutiao.com/ch/news_regimen/",
	//	"https://www.toutiao.com/ch/news_sports/",
	//	"https://www.toutiao.com/ch/news_essay/",
	//}

)
func init(){
	//fmt.Println("init")
	var err error
	//rej, err = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	//if err != nil {
	//	panic(err)
	//}
	rea, err = regexp.Compile("\\<a[\\S\\s]+?\\</a\\>")
	if err != nil {
		panic(err)
	}
	//retitle, err = regexp.Compile(`title: \'[\s\s]+\'`)
	retitle, err = regexp.Compile(`title: \'[\S\s]+?\'`)
	if err != nil {
		panic(err)
	}
	rec, err = regexp.Compile(`content: \'[\S\s]+?\'`)
	if err != nil {
		panic(err)
	}
	go start(func(in string){
		i:=0
		err := ClearDB()
		if err != nil {
			panic(err)
		}
		w:=new(sync.WaitGroup)
		for{
			Coll(uris[i],w)
			w.Wait()
			log.Println("wait")
			i++
			if i>=len(uris){
				err := ClearDB()
				if err != nil {
					panic(err)
				}
				i=0
			}
			<-time.After(2 * time.Minute)
		}

	})
	log.Println("run")
}

func ClearDB() error {

	fmt.Println("begin Clear")
	return wxmsg.CollectionClearDB(func()error{
		return clearLocalDB(wxmsg.DBDelete)
	})

}

func Coll(uri string,w *sync.WaitGroup){
	//runStream(in,func(v interface{},bc *websocket.Conn){
	//	fmt.Println("b",v)
	//})
	fmt.Println(uri)
	openPage(uri,func(v interface{})error{
		requestId:=""
		//count :=0
		step:=0
		id:=float64(time.Now().Unix())
		var id_1 float64 = 0
		_vb := v.(map[string]interface{})
		var stop chan bool = nil
		runStream(_vb["webSocketDebuggerUrl"].(string),w,func(_v interface{},writeChan chan interface{}){

			//timeOut = time.After(time.Minute*2)
			__v :=_v.(map[string]interface{})
			//fmt.Println(__v)
			if step == 0 {
				//if __v["method"] == "Network.loadingFailed"{
				//	if stop != nil {
				//		close(stop)
				//		stop = nil
				//	}
				//	closePage(_vb["id"].(string))
				//	return
				//	//fmt.Println(__v)
				//}
				if __v["method"] !="Network.responseReceived"{
					return
				}
				u := (__v["params"].(map[string]interface{}))
				_u := u["response"].(map[string]interface{})
				//_uri:= _u["url"].(string)
				if !strings.Contains(_u["url"].(string),"/api/pc/feed/"){
					return
				}

				//fmt.Println(__v)
				requestId = u["requestId"].(string)
				step = 1
				if stop != nil {
					close(stop)
					stop = nil
				}
				return
			}else if step ==1 {
				if __v["method"] !="Network.loadingFinished"{
					return
				}
				if !strings.EqualFold((__v["params"].(map[string]interface{}))["requestId"].(string),requestId){
					return
				}
				//fmt.Println(__v)
				writeChan<-map[string]interface{}{
					"method":"Network.getResponseBody",
					"id":id,
					"params":map[string]interface{}{"requestId":requestId},
				}
				step = 2
				return
			}else if step ==2 {
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
				count := 0
				d__ :=  body_["data"]
				if d__ != nil {
				for _,d := range d__.([]interface{}){
					if err := extract(rooturl + d.(map[string]interface{})["source_url"].(string)); err != nil {
						fmt.Println(err)
					}else{
						count++
					}
					time.Sleep(1*time.Second)
				}
				}
				if count == 0 {

					if stop != nil {
						close(stop)
						stop = nil
					}
					closePage(_vb["id"].(string))
					return
					//panic(0)
				}
				go func (){
					stop = make(chan bool)
					for{
						select{
							case <-time.After(5 * time.Second):
								writeChan<-map[string]interface{}{
									"method":"Input.dispatchKeyEvent",
									"id":id_1,
									"params":map[string]interface{}{
										"type":"keyDown",
										"windowsVirtualKeyCode":int(0x22),
										"nativeVirtualKeyCode":int(0x22),
									},
								}
							case <- stop:
								return
						}
					}
				}()
				step = 0

			}
		})

		return nil
	})

}

func start(hand func(string)){
	runout := func(r io.ReadCloser){
		var db [8192]byte
		for{
			n,err := r.Read(db[:])
			if err != nil {
				if err != io.EOF{
					log.Println(err)
					break
					//panic(err)
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
	cmd.Wait()
	//select{}
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
func openPage(u string,hand func(interface{})error){
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
		log.Println(err)
		//panic(err)
	}

}
func runBrowserStream(u string,hand func(interface{})) (chan interface{}){
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	writeChan := make(chan interface{},5)
	go func(){
		var db interface{}
		for{
			err = c.ReadJSON(&db)
			if err != nil {
				log.Println("stream",err)
				return
			}
			go hand(db)
		}
	}()
	go func(){
		for{
			w:= <-writeChan
			if w == nil {
				log.Println("stream w")
				return
			}
			log.Println(w)
			err := c.WriteJSON(w)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}()
	return writeChan

}

func runStream(u string,w *sync.WaitGroup,hand func(interface{},chan interface{}))*websocket.Conn{
//func runStream(u string,hand func(interface{}))*websocket.Conn{
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	//DOM.enable
	err = c.WriteJSON(map[string]interface{}{"method":"Network.enable","id":1})
	if err != nil {
		log.Fatal("w:", err)
	}
	writeChan := make(chan interface{},5)
	stop := make(chan bool)
	//err = c.WriteJSON(map[string]interface{}{"method":"DOM.getFlattenedDocument","id":2})
	//if err != nil {
	//	log.Fatal("w:", err)
	//}
	w.Add(2)
	go func(){
		//fmt.Println(u)
		defer w.Done()
		var db interface{}
		for{
			err = c.ReadJSON(&db)
			if err != nil {
				close(stop)
				//close(writeChan)
				//writeChan = nil
				log.Println("stream",err)
				break
			}
			c.SetReadDeadline(time.Now().Add(time.Minute*2))
			go hand(db,writeChan)

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
				log.Println(w)
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
	return c
}

func extract(uri string) error {
	//fmt.Println(uri)
	return request.ClientHttp_(uri,"GET",nil,config.Conf.Header,
	func(body io.Reader,st int)error {
		db,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		if st != 200 {
			return fmt.Errorf("%d %s",st,db)
		}
		err,p := _extract(db)
		if err != nil {
			//fmt.Println(err)
			return err
		}else{
			//wxmsg.SaveToWXDB(p.ToWXString())
			fmt.Println(p.Title)
			body,ids := p.ToWXString()
			err := wxmsg.SaveToWXDB(body)
			if (err == nil) && (len(ids)>0) {
				err = wxmsg.UpdateToWXDB(binary.BigEndian.Uint64(p.Id),ids[:1])
			}
			if err != nil {
				fmt.Println(err)
			}

		}
		//fmt.Println(string(db))
		return nil
	})

}

func html2Text(t string) string {
	t = strings.Replace(t,`\u003C`,"<",-1)
	t = strings.Replace(t,`\u003E`,">",-1)
	t = strings.Replace(t,`\u002F`,"/",-1)
	t = strings.Replace(t,"\\","",-1)
	//t = strings.Replace(t,"\"","",-1)
	return t
}

func _extract(body []byte) (error,*Page) {

	loc := retitle.FindIndex(body)
	if len(loc)==0 {
		return fmt.Errorf("Not Found title"),nil
	}
	loc_ := rec.FindIndex(body)
	if len(loc_)==0 {
		return fmt.Errorf("Not Found content"),nil
	}
	//content :=html2Text(html.UnescapeString(string(body[loc_[0]+10:loc_[1]-1])))
	//fmt.Println(content)
	p := NewPage(
		strings.Replace(html.UnescapeString(string(body[loc[0]+8:loc[1]-1])),"\"","",-1),
		strings.Replace(html2md.Convert(
		rea.ReplaceAllString(
		html2Text(
		html.UnescapeString(
		string(body[loc_[0]+10:loc_[1]-1]))),"")),"\"","",-1),
		//className,
	)
	//fmt.Println(p.Content)
	err := p.CheckUpdateWork()
	if err != nil {
		return err,p
	}
	return p.SaveDB(),p

}

