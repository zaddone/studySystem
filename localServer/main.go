package main
import(
	"fmt"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/shopping"
	"github.com/zaddone/studySystem/config"
	"github.com/zaddone/studySystem/alimama"
	"github.com/zaddone/studySystem/chromeServer"
	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
	"net/url"
	"time"
	"sort"
	"strings"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"sync"
	"flag"
)

var(
	//flag.String("site","127.0.0.1")
	WXtoken = config.Conf.Minitoken
	Router = gin.Default()
	Remote = flag.String("r", "https://www.zaddone.com/v1","remote")
	wsupgrader = websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 5 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)
func Sign(c *gin.Context){
	url_ := c.Request.URL.Query()
	addSign(&url_)
}
func addSign(u *url.Values){
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	li := []string{WXtoken}
	for _,v := range *u{
		li = append(li,v...)
	}
	sort.Strings(li)
	u.Add("sign",shopping.Sha1([]byte(strings.Join(li,""))))
}
func HandForward(c *gin.Context){
	//c.Request.Body.Close()
	err := requestHttp(
		c.Request.URL.Path,
		c.Request.Method,
		c.Request.URL.Query(),
		c.Request.Body,
		func(body io.Reader,res *http.Response)error{
			c.DataFromReader(res.StatusCode,res.ContentLength,res.Header.Get("content-type"),res.Body,nil)
			return nil
		},
	)
	if err != nil {
		c.JSON(http.StatusNotFound,gin.H{"msg":err.Error()})
	}

}
func requestHttp(path,Method string,u url.Values, body io.Reader,hand func(io.Reader,*http.Response)error)error{
	addSign(&u)
	return request.ClientHttp__(*Remote+path+"?"+u.Encode(),Method,body,nil,hand)
}
func InitShoppingMap()error{
	return requestHttp("/shopping","GET",url.Values{},nil,func(body io.Reader,res *http.Response)error{
		var db []*shopping.ShoppingInfo
		err := json.NewDecoder(body).Decode(&db)
		if err != nil {
			return err
		}
		for _,sh := range db {
			hand := shopping.FuncMap[sh.Py]
			if hand != nil {
				shopping.ShoppingMap.Store(sh.Py,hand(sh,""))
			}
		}
		//fmt.Println(shopping.ShoppingMap)
		return nil
	})
}

func downHandler(w http.ResponseWriter, r *http.Request) {
	var conn *websocket.Conn
	var err error
	conn, err = wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	defer conn.Close()
        t, reply, err := conn.ReadMessage()
        if err != nil {
		fmt.Println(err)
		return
        }
	fmt.Println(t,string(reply))
}
func WsHandler(w http.ResponseWriter, r *http.Request) {
	var conn *websocket.Conn
	var err error
	conn, err = wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	defer conn.Close()
    //for {
        t, reply, err := conn.ReadMessage()
        if err != nil {
		fmt.Println(err)
		return
        }
	fmt.Println(t,string(reply))
	err = InitShoppingMap()
	if err != nil {
		err = conn.WriteMessage(t,[]byte(err.Error()))
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	chanmsg := make(chan interface{},100)
	defer close(chanmsg)
	go func(){
	for m := range chanmsg {
		//fmt.Println("chan",m)
		switch m_ := m.(type){
		case string:
			err = conn.WriteMessage(t,[]byte(m_))
			if err != nil {
				fmt.Println(err)
			}

		case []byte:
			err = conn.WriteMessage(t,m_)
			if err != nil {
				fmt.Println(err)
			}
		default:
			err = conn.WriteJSON(m_)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	}()
	alimama.TaobaoLoginEvent = func(path string){
		fmt.Println("event",path)
		chanmsg<-"/"+config.Conf.Static+"/"+path
	}
	var wait sync.WaitGroup
	shopping.ShoppingMap.Range(func(_k,_v interface{})bool{
		//fmt.Println(k.(string))
		//chanmsg<-k.(string)
		wait.Add(1)
		go func(k interface{},v_ shopping.ShoppingInterface){
			defer wait.Done()
			//v_ := v.(shopping.ShoppingInterface)
			err := v_.OrderDown(func(db interface{}){
				fmt.Println(k.(string),db)
				db_,err := json.Marshal(db)
				if err != nil {
					panic(err)
					fmt.Println(err)
					return
				}
				u := url.Values{}
				u.Add("orderid",db.(map[string]interface{})["order_id"].(string))
				err = requestHttp("/updateorder/"+k.(string),"POST",u,bytes.NewReader(db_),func(body io.Reader,res *http.Response)error{
					data,err := ioutil.ReadAll(body)
					if err != nil {
						return err
					}
					chanmsg<-data
					return nil
					//return conn.WriteMessage(t,db)
				})
				if err != nil {
					fmt.Println(err)
					return
				}
			})
			if err != nil {
				fmt.Println(err)
				return
			}
			u_:= url.Values{}
			u_.Set("update",fmt.Sprintf("%d",v_.GetInfo().Update))
			err = requestHttp("/updatesite/"+k.(string),"GET",u_,nil,func(body io.Reader,res *http.Response)error{
				db,err := ioutil.ReadAll(body)
				fmt.Println("site",string(db))
				return err
			})
			if err != nil {
				//panic(err)
				fmt.Println(err)
			}
		}(_k,_v.(shopping.ShoppingInterface))
		return true
	})
	wait.Wait()
	conn.CloseHandler()(2,"end")

}

func init(){
	flag.Parse()
	//shopping.InitShoppingMap(*siteDB)

	Router.Static("/"+config.Conf.Static,"./"+config.Conf.Static)
	Router.LoadHTMLGlob(config.Conf.Templates)
	Router.GET("/",func(c *gin.Context){
		c.HTML(http.StatusOK,"index.tmpl",nil)
	})
	Router.GET("ws",func(c *gin.Context){
		WsHandler(c.Writer, c.Request)
	})
	Router.GET("updatesite/:py",HandForward)
	Router.GET("shopping/:py",HandForward)
	Router.GET("shopping",HandForward)
	Router.GET("/delsite/:py",HandForward)
	//Router.GET("order/:py",HandForward)
	//Router.GET("order",HandForward)
	Router.GET("order/list",HandForward)
	Router.GET("order/time",HandForward)
	Router.GET("order_apply",HandForward)
	Router.GET("order/del",HandForward)
	Router.GET("wxtoken",HandForward)
	Router.GET("down",func(c *gin.Context){
		downHandler(c.Writer, c.Request)
	})

	//Router.GET("init",func(c *gin.Context){
	//	InitShoppingMap()
	//	c.String(http.StatusOK,"success")
	//})
	//Router.GET("run",func(c *gin.Context){
	//	go DownOrder()
	//	c.String(http.StatusOK,"success")
	//})
	//Router.POST("updateorder/:py",HandForward)
	go Router.Run(config.Conf.Port)
}
func DownOrder(){
	shopping.ShoppingMap.Range(func(k,v interface{})bool{
		fmt.Println(k.(string))
		v_ := v.(shopping.ShoppingInterface)
		err := v_.OrderDown(func(db interface{}){
			db_,err := json.Marshal(db)
			if err != nil {
				panic(err)
				fmt.Println(err)
				return
			}
			u := url.Values{}
			u.Add("orderid",db.(map[string]interface{})["order_id"].(string))
			//var req interface{}
			err = requestHttp("/updateorder/"+k.(string),"POST",u,bytes.NewReader(db_),func(body io.Reader,res *http.Response)error{
				db,err := ioutil.ReadAll(body)
				fmt.Println("order",string(db))
				return err
				//return json.NewDecoder(body).Decode(&req)
			})
			if err != nil {
				panic(err)
				fmt.Println(err)
			}
			//fmt.Println(req)
			//orderlist = append(orderlist,db)
			//fmt.Println(db)
		})
		if err != nil {
			fmt.Println(k,err)
			return true
		}
		u_:= url.Values{}
		u_.Set("update",fmt.Sprintf("%d",v_.GetInfo().Update))
		//var req_ interface{}
		err = requestHttp("/updatesite/"+k.(string),"GET",u_,nil,func(body io.Reader,res *http.Response)error{
			//return json.NewDecoder(body).Decode(&req_)
			db,err := ioutil.ReadAll(body)
			//fmt.Println(db)
			fmt.Println("site",string(db))
			//fmt.Println(string(db))
			return err
		})
		if err != nil {
			panic(err)
			fmt.Println(err)
		}
		//fmt.Println(req_)
		return true
	})
}
func main(){
	//InitShoppingMap()
	//DownOrder()
	//sh,_ := shopping.ShoppingMap.Load("pinduoduo")
	rooturl := fmt.Sprintf("http://127.0.0.1%s",config.Conf.Port)
	fmt.Println(rooturl)
	chromeServer.View(rooturl)
	//select{}
}
