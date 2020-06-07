package main
import(
	"github.com/zaddone/studySystem/shopping"
	//"github.com/zaddone/studySystem/article"
	//"github.com/zaddone/studySystem/config"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	"encoding/json"
	//"compress/gzip"
	//"io"
	"regexp"
	"net/http"
	"strings"
	"strconv"
	"fmt"
	"flag"
	"time"
	"sync"
	//"net/http/httputil"
)
var(

	timeFormat = "20060102"
	cacheDB = "cache.db"
	cacheList = []byte("cachelist")
	//cacheTime = []byte("cachetime")
	MapSession = sync.Map{}
	Router = gin.Default()
	//Router_ = gin.Default()
	siteDB  = flag.String("db","SiteDB","db")
	SessionId = "session_id"
	Port = flag.String("p",":8080","port")
)

func runServerClearMap(){
	for{
		time.Sleep(time.Hour*1)
		MapSession = sync.Map{}
	}
}

func getGoodsDetail(py,id string) interface{}{
	sh,_ := shopping.ShoppingMap.Load(py)
	if sh == nil {
		return nil
	}
	return sh.(shopping.ShoppingInterface).GoodsDetail(id)
}

func saveCache(uri []byte,val interface{}){
	//fmt.Println("save",string(uri))
	err:= shopping.OpenSiteDB(cacheDB,func(db *bolt.DB)error{
		return db.Batch(func(t *bolt.Tx)error{
			b,err := t.CreateBucketIfNotExists(cacheList)
			if err != nil {
				return err
			}
			Now := []byte{byte(time.Now().Day())}
			b_ := b.Bucket(Now)
			if b_ == nil {
				err = b.ForEach(func(k,v []byte)error{
					if v == nil{
						return b.DeleteBucket(k)
					}
					return b.Delete(k)
				})
				if err != nil {
					return err
				}
				b_,err = b.CreateBucketIfNotExists(Now)
				if err != nil {
					return err
				}
			}
			v,err := json.Marshal(val)
			if err != nil {
				return err
			}
			return b_.Put(uri,v)
		})
	})
	if err != nil {
		panic(err)
	}
}
func checkCache(uri []byte) (c interface{}) {
	//uri := []byte(c.Request.URL.String())
	//fmt.Println("check",string(uri))
	err:= shopping.OpenSiteDB(cacheDB,func(db *bolt.DB)error{
		return db.View(func(t *bolt.Tx)error{
			b := t.Bucket(cacheList)
			if b == nil {
				return nil
			}
			b_ := b.Bucket([]byte{byte(time.Now().Day())})
			if b_ == nil {
				return nil
			}
			v := b_.Get(uri)
			if v == nil {
				return nil
			}
			return json.Unmarshal(v,&c)
		})
	})
	if err != nil {
		panic(err)
		fmt.Println(err)
	}
	return
}

func checkSession(c *gin.Context){
	//fmt.Println("remote addr",c.Request.RemoteAddr)
	//fmt.Printf("%s",c.Request.Header["X-Forwarded-For"])
	ip := IpStrToByte(c.Request.Header.Get("X-Forwarded-For"))
	if ip == nil {
		c.Abort()
		return
	}
	s := string(ip)
	v,ok := MapSession.Load(s)
	now := time.Now().Unix()
	if !ok {
		MapSession.Store(s,now)
		c.Next()
		return
	}
	if now == v.(int64){
		c.Abort()
		return
	}
	MapSession.Store(s,now)
	c.Next()
	return
}

func ClearSessionMap(t time.Time){
	MapSession.Range(func(k,v interface{})bool{
		if (t.Unix() - v.(int64))>86400 {
			MapSession.Delete(k)
		}
		return true
	})
}

func IpStrToByte(s string) []byte {
	ips := strings.Split(s,":")
	if len(ips) < 1 {
		return nil
	}
	var ipaddr [4]byte
	for i,p := range strings.Split(ips[0],"."){
		n,err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		ipaddr[i] = byte(n)
	}
	return ipaddr[:]
}

func init(){
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
	//Router.Use(gzip.Gzip(gzip.DefaultCompression))
	shopping.InitShoppingMap(*siteDB)
	go runServerClearMap()
	secureFunc := func() gin.HandlerFunc {
		return checkSession
	}()
	Router.Static("/static","./static")
	//Router.Static("/article","./article")
	Router.LoadHTMLGlob("./templates/*")
	//Router_.GET("/",func(c *gin.Context){
	//	c.Redirect(301,"https://www.zaddone.com")
	//	//c.Data(http.StatusOK,"text/html",Html)
	//	//c.Data(200,"text/html",Html)
	//})
	//Router.Group("/article",ReverseProxy())
	//Router.GET("/sendsms",func(c *gin.Context){
	//	phone := c.Query("phone")
	//	if phone == "" {
	//		return
	//	}
	//	//c.JSON(http.StatusOK,gin.H{"msg":"success"})
	//	//return
	//	err := PhoneCode(phone)
	//	if err != nil {
	//		c.JSONP(http.StatusNotFound,gin.H{"msg":err.Error()})
	//		return
	//	}
	//	//c.JSON(http.StatusOK,gin.H{"code":randCode()})
	//	c.JSONP(http.StatusOK,gin.H{"msg":"success"})
	//})
	//Router.GET("/checksms",func(c *gin.Context){
	//	phone := c.Query("phone")
	//	if phone == "" {
	//		return
	//	}
	//	code := c.Query("code")
	//	if code == "" {
	//		return
	//	}
	//	//session := c.Query("seccion")
	//	//if session == "" {
	//	//	return
	//	//}
	//	err := CheckPhoneCode(phone,code)
	//	if err != nil {
	//		c.JSONP(http.StatusNotFound,gin.H{"msg":err.Error()})
	//		return
	//	}
	//	user := shopping.User{UserId:phone}
	//	err = user.Get()
	//	if err != nil {
	//		if err != io.EOF {
	//			c.JSONP(http.StatusNotFound,gin.H{"msg":err.Error()})
	//			return
	//		}
	//		user.Session = shopping.Sha1([]byte(fmt.Sprintf("%s%s%s",time.Now(),c.Request.RemoteAddr,phone)))
	//		err = user.Update()
	//		if err != nil {
	//			c.JSONP(http.StatusNotFound,gin.H{"msg":err.Error()})
	//			return
	//		}
	//	}
	//	c.JSONP(http.StatusOK,gin.H{"msg":"success","user":user})
	//})
	//search := Router.Group("/search",manageFunc)
	Router.GET("/",gzip.Gzip(gzip.DefaultCompression),func(c *gin.Context){
		//session,err := c.Cookie(SessionId)
		//if err != nil {
		//	session = shopping.Sha1([]byte(fmt.Sprintf("%s%s",time.Now(),c.Request.RemoteAddr)))
		//	c.SetCookie(SessionId,session[:32],3600*24*365*10,"/",".zaddone.com",false,false)
		//}
		if c.Query("content_type") == "json"{

			session,err := c.Cookie(SessionId)
			if err != nil {
				session = shopping.Sha1([]byte(fmt.Sprintf("%s%s",time.Now(),c.Request.RemoteAddr)))
				c.SetCookie(SessionId,session[:32],3600*24*365*10,"/",".zaddone.com",false,false)
			}
			var li []map[string]string
			shopping.ShoppingMap.Range(func(k,v interface{})bool{
				sh := v.(shopping.ShoppingInterface).GetInfo()
				li = append(li,
				map[string]string{
					"Name":sh.Name,
					"Img":sh.Img,
					"Uri":sh.Uri,
					"py":k.(string),
				})
				return true
			})
			c.JSONP(http.StatusOK,li)
			return
		//}else if len(c.Query("keyword"))>0 {
		//	var li []map[string]string
		//	shopping.ShoppingMap.Range(func(k,v interface{})bool{
		//		sh := v.(shopping.ShoppingInterface).GetInfo()
		//		li = append(li,
		//		map[string]string{
		//			"Name":sh.Name,
		//			"Img":sh.Img,
		//			"Uri":sh.Uri,
		//			"py":k.(string),
		//		})
		//		return true
		//	})
		//	c.HTML(http.StatusOK,"index_1.tmpl",gin.H{"site":li})
		//	return
		//}else{
		//	c.HTML(http.StatusOK,"search.tmpl",nil)
		}
		c.HTML(http.StatusOK,"index_1.tmpl",nil)
	})
	Router.GET("/script",gzip.Gzip(gzip.DefaultCompression),func(c *gin.Context){
		session,err := c.Cookie(SessionId)
		if err != nil {
			session = shopping.Sha1([]byte(fmt.Sprintf("%s%s",time.Now(),c.Request.RemoteAddr)))
			c.SetCookie(SessionId,session[:32],3600*24*365*10,"/",".zaddone.com",false,false)
		}
		js:=""
		shopping.ShoppingMap.Range(func(k,v interface{})bool{
			sh := v.(shopping.ShoppingInterface).GetInfo()
		//for k,v := range shopping.ShoppingMap {
			//sh := v.GetInfo()
			js+=fmt.Sprintf("ShoppingMap.set('%s',{func:%sPageHtml,db:[],page:0,html: html%s,py:'%s',name:'%s'});",k,k,k,k,sh.Name)
			return true
		})
		c.Data(http.StatusOK,"application/javascript",[]byte(js))
	})

	Router.GET("/p/:py/:id",secureFunc,func(c *gin.Context){
		sh_,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh_ == nil {
			c.JSONP(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		sh := sh_.(shopping.ShoppingInterface)
		//id := c.Param("id")
		val := []string{c.Param("id")}
		session:= c.Query("session")
		if session != ""{
			val = append(val,session)
		}else{
			session,err := c.Cookie(SessionId)
			if err == nil{
				val = append(val,session)
			}
		}
		db := sh.GoodsUrl(val...)
		u := sh.OutUrl(db)
		fmt.Println(u,db)
		if u == "" {
			c.JSONP(http.StatusNotFound,gin.H{"msg":""})
		}
		c.Redirect(http.StatusMovedPermanently,u)

	})
	//Router.GET("goodsurl",secureFunc,gzip.Gzip(gzip.DefaultCompression),func(c *gin.Context){
	Router.GET("goodsurl",secureFunc,func(c *gin.Context){
		uri := c.Query("url")
		if uri == "" {
			return
		}
		if regexp.MustCompile(`yangkeduo`).MatchString(uri){
			str := regexp.MustCompile(`goods_id=(\d+)`).FindStringSubmatch(uri)
			if len(str)<2{
				return
			}
			//c.JSON(http.StatusOK,getGoodsDetail("pinduoduo",str[1]))
			c.JSONP(http.StatusOK,gin.H{"py":"pinduoduo","db":getGoodsDetail("pinduoduo",str[1])})
			return
		}
		if regexp.MustCompile(`jd`).MatchString(uri){
			str := regexp.MustCompile(`(\d+)\.html`).FindStringSubmatch(uri)
			if len(str)<2{
				str = regexp.MustCompile(`sku=(\d+)`).FindStringSubmatch(uri)
				if len(str)<2{
					str = regexp.MustCompile(`sku/(\d+)`).FindStringSubmatch(uri)
					if len(str)<2{
						return
					}
				}
			}
			c.JSONP(http.StatusOK,gin.H{"py":"jd","db":getGoodsDetail("jd",str[1])})
			//c.JSON(http.StatusOK,getGoodsDetail("jd",str[1]))
			return
		}
		if regexp.MustCompile(`tb|taobao`).MatchString(uri){
			//c.JSON(http.StatusOK,getGoodsDetail("taobao",uri))
			c.JSONP(http.StatusOK,gin.H{"py":"taobao","db":getGoodsDetail("taobao",uri)})
			return
		}
		if regexp.MustCompile(`mogu`).MatchString(uri){
			//c.JSON(http.StatusOK,getGoodsDetail("mogu",uri))
			c.JSONP(http.StatusOK,gin.H{"py":"mogu","db":getGoodsDetail("mogu",uri)})
			return
		}
		if regexp.MustCompile(`suning`).MatchString(uri){
			str := regexp.MustCompile(`(\d+\/\d+)\.html`).FindStringSubmatch(uri)
			if len(str)<2{
				return
			}
			//fmt.Println("suning",str)
			c.JSONP(http.StatusOK,gin.H{"py":"suning","db":getGoodsDetail("suning",strings.Replace(str[1],"/","-",-1))})
			return
		}
		if regexp.MustCompile(`vip`).MatchString(uri){
			//https://detail.vip.com/detail-1711197624-6918740352580831640.html
			str := regexp.MustCompile(`\-(\d+)\.html`).FindStringSubmatch(uri)
			//fmt.Println(str)
			if len(str)<2{
				return
			}
			c.JSONP(http.StatusOK,gin.H{"py":"vip","db":getGoodsDetail("vip",str[1])})
			return
		}
		return

	})
	Router.GET("goodsid/:py",secureFunc,gzip.Gzip(gzip.DefaultCompression),func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		//sh := shopping.ShoppingMap[c.Param("py")]
		if sh == nil {
			c.JSONP(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		keyword := c.DefaultQuery("goodsid","")
		if keyword == "" {
			c.JSONP(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		uri := []byte(c.Request.URL.String())
		db := checkCache(uri)
		if db == nil{
			db = sh.(shopping.ShoppingInterface).GoodsDetail(keyword)
			if db == nil {
				c.JSONP(http.StatusNotFound,gin.H{"msg":"fond not"})
				return
			}
			saveCache(uri,db)
		}
		c.JSONP(http.StatusOK,db)
		return
	})
	Router.GET("miniapp/:py",secureFunc,gzip.Gzip(gzip.DefaultCompression),func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh == nil {
			return
		}
		keyword := c.Query("goodsid")
		if keyword == "" {
			return
		}
		val := []string{keyword}
		session:= c.Query("session")
		if session != ""{
			val = append(val,session)
		}
		ext := c.Query("ext")
		if ext != "" {
			val = append(val,ext)
		}
		uri := []byte(c.Request.URL.String())
		db := checkCache(uri)
		if db == nil{
			db = sh.(shopping.ShoppingInterface).GoodsAppMini(val...)
			if db == nil {
				c.JSONP(http.StatusNotFound,gin.H{"msg":"fond not"})
				return
			}
			saveCache(uri,db)
		}
		//fmt.Println(db)
		c.JSONP(http.StatusOK,db)
		return

	})
	Router.GET("goods/:py",secureFunc,gzip.Gzip(gzip.DefaultCompression),func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		if sh == nil {
			c.JSONP(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}
		keyword := c.DefaultQuery("goodsid","")
		if keyword == "" {
			c.JSONP(http.StatusNotFound,gin.H{"msg":"fond not"})
			return
		}

		val := []string{keyword}

		session:= c.Query("session")
		if session != ""{
			val = append(val,session)
		}else{
			session,err := c.Cookie(SessionId)
			if err != nil{
				return
			}
			val = append(val,session)
		}
		ext := c.Query("ext")
		if ext != "" {
			val = append(val,ext)
		}
		//mini := c.Query("mini")
		//if mini != "" {
		//	val = append(val,mini)
		//}

		uri := []byte(c.Request.URL.String())
		db := checkCache(uri)
		if db == nil{
			db = sh.(shopping.ShoppingInterface).GoodsUrl(val...)
			if db == nil {
				c.JSONP(http.StatusNotFound,gin.H{"msg":"fond not"})
				return
			}
			saveCache(uri,db)
		}
		c.JSONP(http.StatusOK,db)
		return
	})

	//Router.GET("search/:py",secureFunc,gzip.Gzip(gzip.DefaultCompression),func(c *gin.Context){
	Router.GET("search/:py",secureFunc,func(c *gin.Context){
		sh,_ := shopping.ShoppingMap.Load(c.Param("py"))
		fmt.Println("py",sh)
		if sh == nil {
			//c.JSON(http.StatusNotFound,gin.H{"msg":"fond not1"})
			return
		}

		keyword := c.Query("keyword")
		if keyword == "" {
			//c.JSON(http.StatusNotFound,gin.H{"msg":"fond not2"})
			return
		}
		key := []string{keyword}
		ext,_ := c.Cookie("codecity")
		//ext := c.Query("ext")
		if ext != "" {
			key = append(key,ext)
		}
		session:= c.Query("session")
		if session != ""{
			key = append(key,session)
		}else{
			session,err := c.Cookie(SessionId)
			if err == nil{
				key = append(key,session)
			}
		}

		//session,_ := c.Cookie(SessionId)
		uri := []byte(c.Request.URL.String())
		db := checkCache(uri)
		if db == nil{
			db = sh.(shopping.ShoppingInterface).SearchGoods(key...)
			if db == nil{
				//c.JSON(http.StatusNotFound,gin.H{"msg":"fond not3"})
				return
			}
			saveCache(uri,db)
		}
		c.JSONP(http.StatusOK,db)
		return
	})
	//Router.GET("test",func(c *gin.Context){
	//	c.JSON(http.StatusOK,shopping.BuyShopping.GoodsDetail("618279451491"))
	//	return
	//})
	Router.GET("robots.txt",func(c *gin.Context){
		c.String(http.StatusOK,"User-agent: *\nAllow:　/")
		return
	})

}
func main(){
	//go Router.RunTLS(":443","./3375181_zaddone.com.pem","./3375181_zaddone.com.key")
	go Router.Run(*Port)
	select{}
}
