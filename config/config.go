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
	//flag.Parse()
	Conf = NewConfig(*LogFileName)
}
type Config struct {
	ArticleServer string
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
	CollWordName string
	CollPath string
	ToutiaoUri []string
	//UserInfo *url.Values
	OutKey string
	MaxPage int
	//Site []*SitePage
	Minitoken string
	WXtoken string
	AliyunKeyid string
	AliyunSecret string
	Apikeyv3 string
	MerchantId string
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
		c.ArticleServer = "127.0.0.1:8080"
		c.Coll = true
		c.Proxy = ""
		c.MaxPage = 60000
		c.Static = "static"
		c.Port=":8080"
		c.Templates = "./templates/*"
		c.WXAppid = ""
		c.WXSec = ""
		c.CollPageName = "page"
		c.CollWordName = "word"
		c.CollPath = "/data"
		c.Minitoken = ""
		c.WXtoken = ""
		c.Apikeyv3 = ""
		c.MerchantId = ""
		_,err = os.Stat(c.CollPath)
		if err != nil {
			err = os.MkdirAll(c.CollPath,0777)
			if err != nil {
				panic(err)
			}
		}
		c.OutKey="头条客户端|头条号|转载|(点击[\\s\\S]+?关注)|(购买[\\s\\S]+?优惠)"
		c.ToutiaoUri = []string{
			"https://www.toutiao.com",
			"https://www.toutiao.com/ch/news_hot/",
			"https://www.toutiao.com/ch/news_finance/",
			"https://www.toutiao.com/ch/news_tech/",
			"https://www.toutiao.com/ch/news_entertainment/",
			"https://www.toutiao.com/ch/news_game/",
			"https://www.toutiao.com/ch/news_car/",
			"https://www.toutiao.com/ch/funny/",
			"https://www.toutiao.com/ch/news_baby/",
			"https://www.toutiao.com/ch/news_regimen/",
			"https://www.toutiao.com/ch/news_sports/",
			"https://www.toutiao.com/ch/news_essay/",
			"https://www.toutiao.com/ch/news_military/",
			"https://www.toutiao.com/ch/news_fashion/",
			"https://www.toutiao.com/ch/news_discovery/",
			"https://www.toutiao.com/ch/news_regimen/",
			"https://www.toutiao.com/ch/news_history/",
			"https://www.toutiao.com/ch/news_world/",
			"https://www.toutiao.com/ch/news_travel/",
			"https://www.toutiao.com/ch/news_food/",
		}
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
