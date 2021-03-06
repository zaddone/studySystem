package chrome

import(

	"fmt"
	"sync"
	"io/ioutil"
	"os/exec"
	"io"
	"os"
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
	"github.com/boltdb/bolt"
	"net/http"
	"net/url"

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
		"https://www.toutiao.com/",
	}
	Ourl = "http://127.0.0.1:"+port
	k = []byte{10,68,101,118,84,111,111,108,115,32}
	retitle *regexp.Regexp
	rec *regexp.Regexp
	rea *regexp.Regexp
	rej *regexp.Regexp
	regTag *regexp.Regexp

	uris = config.Conf.ToutiaoUri
	//WXDBPushChan = make(chan *UpdateId,100)
	//WXDBDeleteChan = make(chan *DelId,100)
	WXDBChan = make(chan interface{},100)
	WordDB = config.Conf.CollPath+"/"+config.Conf.CollWordName+".db"
	PageDB = config.Conf.CollPath+"/"+config.Conf.CollPageName+".db"
	pageBucket = []byte("page")
	pageListBucket = []byte("pageList")
	//pageVodBucket = []byte("pageVod")
	WordBucket = []byte("word")

	DbWord *bolt.DB
	DbPage *bolt.DB
	//WordTmp string
	//pageTmp string
	HandIDFunc = map[float64]func(i float64,db interface{}){}
	Num float64 = 1

)
type pageInterface interface {
	GetId() uint64
	ToWXString() (string,[]string)
	GetUpdate() bool
	GetTitle() string
}

type DelId struct{
	coll string
	ids []string
}
func NewDelId(c string,i []string) *DelId {
	return &DelId{coll:c,ids:i}
}
//type UpdateId struct{
//	id uint64
//	ids []string
//}
type UpdateFile struct{
	coll string
	uri string
}

func syncPushWXDB(){
	for{
		p := <-WXDBChan
		switch rs := p.(type) {
		case string:
			f,err := os.OpenFile(config.Conf.CollPath+"/"+config.Conf.CollPageName,os.O_APPEND|os.O_CREATE|os.O_RDWR,0777)
			if err != nil {
				//return err
				panic(err)
			}
			_,err = f.WriteString(rs)
			if err != nil {
				panic(err)
			}
			f.Close()
		case []string:
			f,err := os.OpenFile(config.Conf.CollPath+"/"+config.Conf.CollWordName,os.O_APPEND|os.O_CREATE|os.O_RDWR,0777)
			if err != nil {
				//return err
				panic(err)
			}
			for _,r := range rs{
				_,err = f.WriteString(r)
				if err != nil {
					panic(err)
				}
			}
			f.Close()


		case *DelId:
			fmt.Println("del")
			err := wxmsg.DBDelete(rs.coll,rs.ids)
			if err != nil {
				log.Println("del",err)
			}
		//case *UpdateId:
		//	fmt.Println("update")
		//	err := wxmsg.UpdateToWXDB(rs.id,rs.ids)
		//	if err != nil {
		//		log.Println("update",err)
		//	}
		case *UpdateFile:
			fmt.Println("file",rs)
			uri := config.Conf.CollPath+"/"+rs.uri
			err := wxmsg.UpDBToWX(rs.coll,uri)
			if err != nil {
				log.Println(err)
				if _,err = os.Stat(uri);err== nil {
					WXDBChan<-rs
				}
			}
		default:
			log.Println("default",p,rs)
		}
		//select{
		//case p:=<-WXDBPushChan:
		//	err = wxmsg.UpdateToWXDB(p.id,p.ids)
		//case idsc := <-WXDBDeleteChan:
		//	err := wxmsg.DBDelete(idsc.coll,idsc.ids)
		//	if err != nil {
		//		log.Println(err)
		//	}
		//}
	}
}

func init(){
	fmt.Println("chrome init")
	var err error
	DbPage,err = bolt.Open(PageDB,0600,nil)
	if err != nil {
		panic(err)
	}
	DbWord,err = bolt.Open(WordDB,0600,nil)
	if err != nil {
		panic(err)
	}
	if !config.Conf.Coll{
		return
	}
	//rej, err = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	//if err != nil {
	//	panic(err)
	//}
	rea, err = regexp.Compile("\\<a[\\S\\s]+?\\</a\\>")
	if err != nil {
		panic(err)
	}
	//retitle, err = regexp.Compile(`title: \'[\s\s]+\'`)
	//chineseTag: '军事',
	regTag,err = regexp.Compile(`chineseTag: \'[\S\s]+?\'`)
	if err != nil {
		panic(err)
	}

	retitle, err = regexp.Compile(`title: \'[\S\s]+?\'`)
	if err != nil {
		panic(err)
	}
	rec, err = regexp.Compile(`content: \'[\S\s]+?\'`)
	if err != nil {
		panic(err)
	}

	go syncPushWXDB()
	go func(){
	for{

		err = ClearDB()
		if err != nil {
			panic(err)
			fmt.Println(err)
		}
		urls_:=make([]string,0,len(uris))
		for _,u := range uris{
			urls_ = append(urls_,u)
		}
		runStart(func(u string)error{
		//fmt.Println(u)
		return openPage_(func(v interface{})error{
		_vb := v.(map[string]interface{})
		runStream_(_vb["webSocketDebuggerUrl"].(string),func(db interface{},write chan interface{},wc *websocket.Conn){
			db_ :=db.(map[string]interface{})
			id_ := db_["id"]
			if  id_ != nil {
				id__ := id_.(float64)
				fun :=HandIDFunc[id__]
				if fun != nil {
					go fun(id__,db)
				}
				return
			}
			if db_["method"] !="Network.responseReceived"{
				return
			}
			u := (db_["params"].(map[string]interface{}))
			_u := u["response"].(map[string]interface{})
			_uri := _u["url"].(string)
			if !strings.Contains(_uri,"/api/pc/feed/"){
				return
			}
			GetPageCookies(write,_vb["url"].(string),_uri,func(){
				if len(urls_) == 0 {
					wc.Close()
					return
				}
				PageNavigate(write,urls_[0])
				urls_ = urls_[1:]
			})


		})
		return nil
		})
		})

		findPageVod(config.Conf.MaxPage)
		updateFileToWX()
		<-time.After(5 * time.Hour)
	}
	}()
	log.Println("run")
}
func runStart(hand func(string)error){
	runout := func(r io.Reader){
		//defer wait.Done()
		var db [8192]byte
		for{

			n,err := r.Read(db[:])
			if err != nil {
				if err != io.EOF{
					log.Println(err)
				}
				return
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
	cmd := exec.Command("google-chrome",op...)
	//cmd := exec.Command("google-chrome",append(op,uri)... )
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
	//err = cmd.Start()
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	//cmd.Wait()
	out.Close()
	outerr.Close()
	fmt.Println("over start")
}
//func runPage(){
//	runStart(func(u string)error{
//		fmt.Println(u)
//		return nil
//	})
//}
func updateFileToWX() error {

	//<-time.After(5 * time.Minute)
	//_,err := os.Stat(config.Conf.CollWordName)
	//if err != nil {
	//	err = WordJsonFile()
	//	if err != nil {
	//		return err
	//	}
	//}
	//fmt.Println(len(WXDBChan))
	WXDBChan<-&UpdateFile{config.Conf.CollWordName,config.Conf.CollWordName}
	WXDBChan<-&UpdateFile{config.Conf.CollPageName,config.Conf.CollPageName}
	return nil

}


func ClearDB() error {

	//fmt.Println("begin Clear")
	//defer fmt.Println("end Clear")

	return wxmsg.CollectionClearDB(func() error {
		//return clearLocalDB(max,wxmsg.DBDelete)
		return clearLocalDB(func(ids []string,idw []string)error{
			WXDBChan<-&DelId{config.Conf.CollPageName,ids}
			WXDBChan<-&DelId{config.Conf.CollWordName,idw}
			return nil
		})

	})

}
func requestPageList(uri,_uri string) error{
	return request.ClientHttpR_(_uri,"GET",nil,uri,config.Conf.Header,
	func(body io.Reader,st int)error {
		//db,err := ioutil.ReadAll(body)
		//if err != nil {
		//	return err
		//}
		if st !=  200 {
			return fmt.Errorf("s = %d",st)
		}
		db:= map[string]interface{}{}
		err:= json.NewDecoder(body).Decode(&db)
		if err != nil {
			//panic(err)
			return err
		}
		if db["data"] == nil {
			return nil
		}
		for _,d := range db["data"].([]interface{}){
			//fmt.Println(d)
			if err := extract(rooturl + d.(map[string]interface{})["source_url"].(string)); err != nil {
				fmt.Println(err)
			}
			time.Sleep(1*time.Second)
		}

		//fmt.Println(st,string(db),len(db))
		//time.Sleep(5*time.Second)
		return nil
		//return fmt.Errorf("%d %s",st,db)

	})

}
func AddAction(write chan interface{},info map[string]interface{},h func(interface{})){
	Num++
	HandIDFunc[Num] = func(i float64,db interface{}){
		delete(HandIDFunc,i)
		h(db)
	}
	info["id"] = Num
	write <- info
}
func GetPageCookies(write chan interface{},u string,u_ string,success func()){
	AddAction(
		write,
		map[string]interface{}{
			"method":"Network.getAllCookies",
		},
		func(db_ interface{}){
		cookies :=(((db_.(map[string]interface{}))["result"].(map[string]interface{}))["cookies"]).([]interface{})
		var cos []*http.Cookie = make([]*http.Cookie,0,len(cookies))
		ur,err := url.Parse(u)
		if err != nil {
			panic(err)
		}
		for _,_c_ := range cookies {
			c_ := _c_.(map[string]interface{})
			domain := c_["domain"].(string)

			if !strings.Contains(ur.Hostname(),domain){
				continue
			}
			cos = append(cos,&http.Cookie{
				Name:c_["name"].(string),
				Value:c_["value"].(string),
				Domain:domain,
				Path:c_["path"].(string),
				Expires:time.Unix(int64(c_["expires"].(float64)),0),
				HttpOnly:c_["httpOnly"].(bool),
				Secure:c_["secure"].(bool),
				//SameSite:c_["sameSite"].(http.SameSite),
			})

		}
		request.Jar.SetCookies(ur,cos)
		err = requestPageList(u,u_)
		if err != nil {
			fmt.Println(err)
			//panic(err)
		}else{
			success()
		}
		//fmt.Println(db)
		},
	)
}
func PageNavigate(write chan interface{},u string){
	fmt.Println(u)
	AddAction(
		write,
		map[string]interface{}{
			"method":"Page.navigate",
			"params":map[string]interface{}{
				"url":u,
			},
		},
		func(db interface{}){
			fmt.Println(db)
		},
	)

}
//func Coll(uri string)error{
//	var id_1 float64 = 999
//	return openPage_(uri,func(v interface{})error{
//		_vb := v.(map[string]interface{})
//		var _uri_ string = ""
//		runStream(_vb["webSocketDebuggerUrl"].(string),func(_v interface{},writeChan chan interface{},wc *websocket.Conn){
//			__v :=_v.(map[string]interface{})
//			if _uri_ != "" {
//				if (__v["id"] == nil) ||
//				(__v["id"].(float64) != id_1) {
//					return
//				}
//				defer wc.Close()
//				ur,err := url.Parse(_uri_)
//				if err != nil {
//					panic(err)
//				}
//				if __v["result"] == nil {
//					return
//				}
//				cookies :=((__v["result"].(map[string]interface{}))["cookies"]).([]interface{})
//				var cos []*http.Cookie = make([]*http.Cookie,0,len(cookies))
//				for _,_c_ := range cookies {
//					c_ := _c_.(map[string]interface{})
//					domain := c_["domain"].(string)
//					if !strings.Contains(ur.Hostname(),domain){
//						continue
//					}
//					cos = append(cos,&http.Cookie{
//						Name:c_["name"].(string),
//						Value:c_["value"].(string),
//						Domain:domain,
//						Path:c_["path"].(string),
//						Expires:time.Unix(int64(c_["expires"].(float64)),0),
//						HttpOnly:c_["httpOnly"].(bool),
//						Secure:c_["secure"].(bool),
//						//SameSite:c_["sameSite"].(http.SameSite),
//					})
//				}
//				request.Jar.SetCookies(ur,cos)
//				err = requestPageList(uri,_uri_)
//				if err != nil {
//					panic(err)
//				}
//				fmt.Println("close")
//				err = wc.Close()
//				if err != nil {
//					panic(err)
//				}
//			}
//			if __v["method"] !="Network.responseReceived"{
//				return
//			}
//			u := (__v["params"].(map[string]interface{}))
//			_u := u["response"].(map[string]interface{})
//			_uri := _u["url"].(string)
//			if !strings.Contains(_uri,"/api/pc/feed/"){
//				return
//			}
//			//err := requestPageList(uri,_uri)
//			//if err != nil {
//			//	panic(err)
//			//}
//			//fmt.Println("close")
//			//err = wc.Close()
//			//if err != nil {
//			//	panic(err)
//			//}
//
//			//fmt.Println(_u["headers"])
//			//fmt.Println(__v)
//			_uri_ = _uri
//			writeChan<-map[string]interface{}{
//				"method":"Network.getAllCookies",
//				"id":id_1,
//			}
//			return
//		})
//		return nil
//	})
//
//}

func start(uri string,hand func(string)error){
	//var wait sync.WaitGroup
	//wait.Add(2)
	runout := func(r io.Reader){
		//defer wait.Done()
		var db [8192]byte
		for{

			n,err := r.Read(db[:])
			if err != nil {
				if err != io.EOF{
					log.Println(err)
				}
				return
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
	cmd := exec.Command("google-chrome",append(op,uri)... )
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
	err = cmd.Run()
	//err = cmd.Start()
	if err != nil {
		fmt.Println(err)
		//panic(err)
		//log.Fatal(err)
	}
	//wait.Wait()
	//cmd.Wait()
	out.Close()
	outerr.Close()
	fmt.Println("over start")
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

func openPage_(hand func(interface{})error) error {
	for{
	//	fmt.Println(Ourl)
	err := request.ClientHttp_(Ourl+"/json","GET",nil,nil,func(body io.Reader,st int)error {
		fmt.Println(st)
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
		k_ := k.([]interface{})
		if len(k_) == 0 {
			return fmt.Errorf("find not")
		}
		for _,v_ := range k_ {
			if er := hand(v_); er != nil {
				panic(er)
			}
			break
		}
		return nil
	})
	if err == nil {
		log.Println("cmd end")
		err = exec.Command("pkill","chrome").Run()
		if err != nil {
			fmt.Println(err)
		}
		return nil
	}
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
//func runBrowserStream(u string,hand func(interface{})) (chan interface{}){
//	c, _, err := websocket.DefaultDialer.Dial(u, nil)
//	if err != nil {
//		log.Fatal("dial:", err)
//	}
//	writeChan := make(chan interface{},5)
//	go func(){
//		var db interface{}
//		for{
//			err = c.ReadJSON(&db)
//			if err != nil {
//				log.Println("stream",err)
//				return
//			}
//			go hand(db)
//		}
//	}()
//	go func(){
//		for{
//			w:= <-writeChan
//			if w == nil {
//				log.Println("stream w")
//				return
//			}
//			log.Println(w)
//			err := c.WriteJSON(w)
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//		}
//	}()
//	return writeChan
//
//}

func runStream_(u string,hand func(interface{},chan interface{},*websocket.Conn)){
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	err = c.WriteJSON(map[string]interface{}{"method":"Network.enable","id":1})
	if err != nil {
		log.Fatal("w:", err)
	}
	writeChan := make(chan interface{},5)
	//PageNavigate(writeChan,
	var wait sync.WaitGroup
	wait.Add(2)
	go func(){
		defer wait.Done()
		var db interface{}
		for{
			err = c.ReadJSON(&db)
			if err != nil {
				close(writeChan)
				log.Println("stream",err)
				break
			}
			c.SetReadDeadline(time.Now().Add(time.Minute*2))
			go hand(db,writeChan,c)
		}
	}()
	go func(){
		defer wait.Done()
		for w := range writeChan{
			err := c.WriteJSON(w)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}()
	wait.Wait()
	fmt.Println("runStream")
	return

}
func runStream(u string,hand func(interface{},chan interface{},*websocket.Conn)){
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
	//stop := make(chan bool)
	//err = c.WriteJSON(map[string]interface{}{"method":"DOM.getFlattenedDocument","id":2})
	//if err != nil {
	//	log.Fatal("w:", err)
	//}
	var wait sync.WaitGroup
	wait.Add(2)
	go func(){
		defer wait.Done()
		var db interface{}
		for{
			err = c.ReadJSON(&db)
			if err != nil {
				close(writeChan)
				//close(writeChan)
				//writeChan = nil
				log.Println("stream",err)
				break
			}
			c.SetReadDeadline(time.Now().Add(time.Minute*2))
			go hand(db,writeChan,c)
		}
	}()
	go func(){
		defer wait.Done()
		for w := range writeChan{
			err := c.WriteJSON(w)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}()
	wait.Wait()


	//c.Close()
	return
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
		err,_ = _extract(db)
		return err
		//if err != nil {
		//	return err
		//}
		//f,err := os.OpenFile(config.Conf.CollPageName,os.O_APPEND|os.O_CREATE|os.O_RDWR,0777)
		//if err != nil {
		//	panic(err)
		//}
		////s_ := p.ToWXString()
		////WXDBChan<-&UpdateId{p.GetId(),ids}
		//fmt.Println(p.Title)
		//_,err = f.WriteString(p.ToWXString())
		//return f.Close()
		//err := wxmsg.SaveToWXDB(body)
		//if (err == nil) && (len(ids)>0) {
		//	err = wxmsg.UpdateToWXDB(binary.BigEndian.Uint64(p.Id),ids[:1])
		//}
		//if err != nil {
		//	fmt.Println(err)
		//}
		//fmt.Println(string(db))
		//return nil
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

	//tag := regTag.Findindex(body)
	//if len(loc)==0 {
	//	return fmt.Errorf("Not Found tag"),nil
	//}
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
		"page",
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

			//p_.Title += binary.BigEndian.Uint64(id)
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

func EachDBf(db *bolt.DB,Bucket []byte,beginkey []byte,hand func(b *bolt.Bucket,k,v []byte)error)error{

	return db.View(func(tx *bolt.Tx)error{
		b := tx.Bucket(Bucket)
		if b == nil {
			return fmt.Errorf("%s ==nil",Bucket)
		}
		c := b.Cursor()
		var k,v []byte
		if len(beginkey) == 0 {
			k,v = c.Last()
		}else{
			k,v = c.Seek(beginkey)
		}
		for ;k!= nil;k,v = c.Prev(){
			err := hand(b,k,v)
			if err != nil {
				return err
			}
		}
		return nil
	})

}

func EachDB(db *bolt.DB,Bucket []byte,beginkey []byte,hand func(b *bolt.Bucket,k,v []byte)error)error{

	return db.View(func(tx *bolt.Tx)error{
		b := tx.Bucket(Bucket)
		if b == nil {
			return fmt.Errorf("%s ==nil",Bucket)
		}
		c := b.Cursor()
		for k,v := c.Seek(beginkey);k!= nil;k,v = c.Next(){
			err := hand(b,k,v)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func getFileTmpName(w []byte) string{
	return fmt.Sprintf("%s_%d",w,time.Now().UnixNano())
}
//func WordJsonFile()error{
//	//WordTmp = getFileTmpName(WordBucket)
//	//f,err := os.OpenFile(config.Conf.CollWordName,os.O_APPEND|os.O_CREATE|os.O_RDWR,0777)
//	//if err != nil {
//	//	return err
//	//}
//	//defer f.Close()
//	var wordUp []string
//	defer func(){
//		WXDBChan<-wordUp
//	}()
//	return EachDB(DbWord,WordBucket,[]byte{0},func(b *bolt.Bucket,k,v []byte)error{
//		le := len(v)
//		lev := le/8
//		if lev>50 {
//			return nil
//		}
//		nolist := make([]string,0,lev)
//		for i:=0;i<le;i+=8 {
//			pid := v[i:i+8]
//			nolist = append(nolist,fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(pid)))
//		}
//		if !bytes.HasPrefix(k,[]byte("vod")){
//
//		}
//
//		 _,err = f.WriteString(fmt.Sprintf("{_id:\"%s\",link:[%s]}",string(k),strings.Join(nolist,",")))
//		return err
//	})
//
//}
func PageJsonFile()error{

	return nil
	//f,err := os.OpenFile(config.Conf.CollPageName,os.O_APPEND|os.O_CREATE|os.O_RDWR,0777)
	//if err != nil {
	//	return err
	//}
	//defer f.Close()
	//p := &Page{}
	////p_ := &Page{}
	//var list,vodlist []byte
	//err = EachDB(DbPage,pageBucket,[]byte{0},func(b *bolt.Bucket,k,v []byte)error{
	//	err = json.Unmarshal(v,p)
	//	if err != nil {
	//		return err
	//	}
	//	if strings.HasPrefix(p.Content,contentTag){
	//		vodlist = append(vodlist,k...)
	//	}else{
	//		list = append(list,k...)
	//	}
	//	p.relevant = append(p.Children,p.Par...)
	//	p_db,_ := p.ToWXString()
	//	_,err = f.WriteString(p_db)
	//	return err

	//})
	//if err != nil {
	//	return err
	//}
	//tx,err := DbPage.Begin(true)
	//if err != nil{
	//	return err
	//}
	//bl,err := tx.CreateBucketIfNotExists(pageListBucket)
	//if err != nil{
	//	return err
	//}
	//err = bl.Put([]byte("page"),list)
	//if err != nil {
	//	return err
	//}
	//err = bl.Put([]byte("vod"),vodlist)
	//if err != nil {
	//	return err
	//}
	//fmt.Println(len(list)/8,len(vodlist)/8)
	//return tx.Commit()


}

func ReadPageList(begin []byte,max int) (list []*Page,err error){

	list = make([]*Page,0,max)
	err = EachDBf(DbPage,pageBucket,begin,func(b *bolt.Bucket,k,v []byte)error{
		p := &Page{}
		er := json.Unmarshal(v,p)
		if er == nil && !strings.HasPrefix(p.Content,contentTag) {
			p.Title += fmt.Sprintln(binary.BigEndian.Uint64(k))
			list = append(list,p)
			if len(list)>=max {
				return io.EOF
			}
		}
		return nil
	})
	return

}
