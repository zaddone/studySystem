package conf
import(
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"flag"
)
type Config map[string]interface{}
func (self *Config) Save(){
	fi,err := os.OpenFile(*LogFileName,os.O_CREATE|os.O_WRONLY,0777)
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
func (self *Config) Get(key string) interface{} {
	return self[key]
}

func (self *Config) Set(key string,interface{}) interface{} {
	self[key] = val
	self.Save()
}
var(
	LogFileName   = flag.String("c", "confeasy.log", "config log")
	Conf Config
)
func init(){
	//EntryList = make(chan *Entry,1000)
	//flag.Parse()
	Conf = NewConfig()
}
func NewConfig()  Config {
	_,err := os.Stat(*LogFileName)
	if err != nil {
		return Config{}
	}
	var c Config
	if _,err := toml.DecodeFile(fileName,&c);err != nil {
		panic(err)
	}
	return c

}
