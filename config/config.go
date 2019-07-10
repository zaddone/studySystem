package config
import(
	"net/http"
	"github.com/BurntSushi/toml"
	"flag"
	"os"
)
var(
	LogFileName   = flag.String("c", "conf.log", "config log")
	Conf *Config
)
func init(){
	//EntryList = make(chan *Entry,1000)
	flag.Parse()
	Conf = NewConfig(*LogFileName)
}
type Config struct {
	Proxy string
	Port string
	DbPath string
	KvDbPath string
	DeduPath string
	Templates string
	Static string
	Header http.Header
	WeixinUrl string
	Coll bool
	WXAppid string
	WXSec string
	CollPageName string
	//UserInfo *url.Values
	OutKey string
	//Site []*SitePage
}
func (self *Config) Save(fileName string){
	fi,err := os.OpenFile(fileName,os.O_CREATE|os.O_WRONLY,0777)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	e := toml.NewEncoder(fi)
	err = e.Encode(self)
	if err != nil {
		panic(err)
	}
}
func NewConfig(fileName string)  *Config {
	var c Config
	_,err := os.Stat(fileName)
	if err != nil {
		//c.UserInfo=&url.Values{
		//"username":[]string{""},
		//"password":[]string{""},
		//"randCode":[]string{""}}
		//c.UserArr=[]string{"lqylqjd","lqylxhsq","lqyyhsq","lqyjpc"}
		c.Coll = true
		c.Proxy = ""
		c.KvDbPath="MyKV.db"
		c.DeduPath="dedu.db"
		c.Static = "static"
		c.Port=":8080"
		c.DbPath = "foo.db"
		c.Templates = "./templates/*"
		c.WeixinUrl = "https://weixin.sogou.com/weixin?type=1&s_from=input&query=longquanjy&ie=utf8"
		c.WXAppid = "wx92ebd09c7b0d944f"
		c.WXSec = "b3005d3c298e27b60ee1f90d188a9d86"
		c.CollPageName = "pageColl"
		c.OutKey="头条号|公众号|打赏|转载|订购|(点击[\\s\\S]+?关注)|(购买[\\s\\S]+?优惠)"
		c.Header = http.Header{
			//"Content-Type":[]string{"application/x-www-form-urlencoded","multipart/form-data"},
			"Upgrade-Insecure-Requests":[]string{"1"},
			"Pragma": []string{"no-cache"},
			"Cache-Control": []string{"no-cache"},
			"TE":[]string{"Trailers"},
			"Accept":[]string{"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
			"Connection":[]string{"keep-alive"},
			"Accept-Encoding":[]string{"gzip, deflate, sdch"},
			"Accept-Language":[]string{"zh-CN,zh;q=0.8"},
			"User-Agent":[]string{"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:66.0) Gecko/20100101 Firefox/66.0"}}
		c.Save(fileName)
	}else{
		if _,err := toml.DecodeFile(fileName,&c);err != nil {
			panic(err)
		}
	}
	return &c
}
