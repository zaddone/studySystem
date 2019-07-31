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
	//"encoding/binary"
	"github.com/lunny/html2md"
	"github.com/zaddone/studySystem/request"
	"github.com/gorilla/websocket"
	"github.com/zaddone/studySystem/config"
	"github.com/zaddone/studySystem/wxmsg"
	"github.com/boltdb/bolt"

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
	WXDBPushChan = make(chan pageInterface,100)
	WXDBDeleteChan = make(chan []string,100)
	WordDB = "word.db"
	PageDB = "page.db"
	pageBucket = []byte("page")
	WordBucket = []byte("word")

	DbWord *bolt.DB
	DbPage *bolt.DB

)
type pageInterface interface {
	GetId() uint64
	ToWXString() (string,[]string)
	GetUpdate() bool
	GetTitle() string
}
func syncPushWXDB(){
	for{
		select{
		case p:=<-WXDBPushChan:
			fmt.Println(p.GetTitle())
			body,ids := p.ToWXString()
			if p.GetUpdate(){
				err := wxmsg.UpdateWXDB(config.Conf.CollPageName,fmt.Sprintf("%d",p.GetId()),body)
				if err != nil {
					log.Println(err)
				}
			}else{
				err := wxmsg.SaveToWXDB(body)
				if (err == nil) && (len(ids)>0) {
					err = wxmsg.UpdateToWXDB(p.GetId(),ids[:1])
				}
			}
		case ids := <-WXDBDeleteChan:
			err := wxmsg.DBDelete(ids)
			if err != nil {
				log.Println(err)
			}

		}
	}

}
func init(){
	//fmt.Println("init")
	var err error
	DbPage,err = bolt.Open(PageDB,0600,nil)
	if err != nil {
		panic(err)
	}
	DbWord,err = bolt.Open(WordDB,0600,nil)
	if err != nil {
		panic(err)
	}
	return
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

	go syncPushWXDB()
	//go syncRunPageVod()
	//fmt.Println("run vod")
	//return
	go start(func(in string)error{
		//findPageVod()
		UpWord()
		w:=new(sync.WaitGroup)
		for{
			for _,u := range uris {
				err = Coll(u,w)
				if err != nil {
					return err
				}
				w.Wait()
			}
			findPageVod()
			UpWord()
			ClearDB(500)
			<-time.After(15 * time.Minute)

		}
		return nil
	})
	log.Println("run")
}

func UpWord(){
	word := "word"
	w,err := getWord()
	log.Println("word begin",len(w))
	if err != nil {
		log.Println(err)
		return
	}
	err = wxmsg.DeleteColl(word)
	if err != nil {
		log.Println(err)
	}
	err = wxmsg.CreateColl(word)
	if err != nil {
		log.Println(err)
		return
	}

	ci :=0
	var db [][]string
	var d []string
	for k,v := range w {
		ci++
		d = append(d,fmt.Sprintf("{_id:\"%s\",link:[%s]}",k,strings.Join(v,",")))
		if ci>=100{
			ci=0
			db = append(db,d)
			d = nil
		}

	}
	for _,d := range db {
		//fmt.Println(strings.Join(d,","))
		err = wxmsg.AddToWXDB(word,strings.Join(d,","))
		if err != nil {
			fmt.Println(err)
		}
	}


}

func ClearDB(max int) error {

	fmt.Println("begin Clear")
	return wxmsg.CollectionClearDB(func()error{
		//return clearLocalDB(max,wxmsg.DBDelete)
		return clearLocalDB(max,func(ids []string)error{
			WXDBDeleteChan<-ids
			return nil
		})
	})

}

func Coll(uri string,w *sync.WaitGroup)error{
	//runStream(in,func(v interface{},bc *websocket.Conn){
	//	fmt.Println("b",v)
	//})
	fmt.Println(uri)
	return openPage(uri,func(v interface{})error{
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
			if bytes.HasPrefix(db[:n],k){
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
	}

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
func openPage(u string,hand func(interface{})error) error {
	return request.ClientHttp_(Ourl+"/json/new/?"+u,"GET",nil,nil,func(body io.Reader,st int)error {
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
		er := hand(k)
		if er != nil {
			log.Println(er)
		}
		return nil

	})
	//if err != nil {
	//	log.Println(err)
	//	//panic(err)
	//}

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
			//fmt.Println(p.Title)
			WXDBPushChan<-p

			//body,ids := p.ToWXString()
			//err := wxmsg.SaveToWXDB(body)
			//if (err == nil) && (len(ids)>0) {
			//	err = wxmsg.UpdateToWXDB(binary.BigEndian.Uint64(p.Id),ids[:1])
			//}
			//if err != nil {
			//	fmt.Println(err)
			//}

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

func SearchPage(key string,hand func(p *Page)) error {
	if len(key) == 0 {
		return fmt.Errorf("key ==0")
	}
	ks := map[string]int{}

	lr := []rune(key)
	for j:=0;j<len(lr);j++{
		for _j:=j+2;_j<=len(lr);_j++ {
			ks[string(lr[j:_j])]+=1
		}
	}
	fmt.Println(ks)
	var keys [][]byte
	for str,_ := range ks{
		keys = append(keys,[]byte(str))
	}
	pageMap := map[string]float64{}
	err := DbWord.View(func(tx *bolt.Tx)error{
		b := tx.Bucket(WordBucket)
		if b == nil {
			return fmt.Errorf("%s == nil",WordBucket)
		}
		c := b.Cursor()
		for _,str := range keys {
			k,v := c.Seek(str)
			if bytes.Equal(k,str) || bytes.Contains(k,str){
				vf := 1.0/float64(len(v)/8)*float64(len(str))
				for i:=0;i<len(v);i+=8{
					pageMap[string(v[i:i+8])] += vf
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(pageMap) == 0 {
		return fmt.Errorf("kmap == 0")
	}
	var max float64
	var maxid [][]byte
	for k,v := range pageMap {
		if v>max {
			max = v
			maxid = [][]byte{[]byte(k)}
		}else if v==max{
			maxid = append(maxid,[]byte(k))
		}
	}
	//var pa []*Page
	return DbPage.View(func(tx *bolt.Tx) error{
		b := tx.Bucket(pageBucket)
		if b == nil {
			return fmt.Errorf("%s == nil",pageBucket)
		}
		for _,id := range maxid{
			d_ := b.Get(id)
			if d_ == nil {
				continue
			}
			p_ := &Page{}
			err := json.Unmarshal(d_,p_)
			if err != nil{
				fmt.Println(err)
				continue
			}
			hand(p_)
			//pa = append(pa,p_)
			for i:=0;i< len(p_.Children);i+=8{
				d__ := b.Get(p_.Children[i:i+8])
				if d__ == nil {
					continue
				}
				p__ := &Page{}
				err =  json.Unmarshal(d__,p__)
				if err != nil {
					fmt.Println(err)
					continue
				}
				hand(p__)
				//pa = append(pa,p__)
			}
		}
		return nil
	})
	//if err != nil {
	//	return err
	//	//fmt.Println(err)
	//}
	//if len(pa)==0 {
	//	return fmt.Errorf("page == 0 ")
	//}


}

func EachDB(db *bolt.DB,Bucket []byte,beginkey []byte,hand func(k,v []byte)error)error{

	return db.View(func(tx *bolt.Tx)error{
		b := tx.Bucket(Bucket)
		if b == nil {
			return fmt.Errorf("%s ==nil",Bucket)
		}
		c := b.Cursor()
		for k,v := c.Seek(beginkey);k!= nil;k,v = c.Next(){
			err := hand(k,v)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func ReadPageList(begin []byte,max int) (list []*Page,err error){
	list = make([]*Page,0,max)
	err = EachDB(DbPage,pageBucket,begin,func(k,v []byte)error{
		p := &Page{}
		er := json.Unmarshal(v,p)
		if er == nil {
			list = append(list,p)
			if len(list)>=max {
				return io.EOF
			}
		}
		return nil
	})
	return

}
