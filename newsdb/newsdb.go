package newsdb
import(
	"io"
	"io/ioutil"
	"fmt"
	"time"
	"regexp"
	"strings"
	"net/url"
	//"bufio"
	"math/rand"
	//"encoding/json"
	//"github.com/PuerkitoBio/goquery"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/wxmsg"
	"github.com/zaddone/studySystem/config"
	"github.com/lunny/html2md"
	//"github.com/gomarkdown/markdown"
	//"github.com/gomarkdown/markdown/html"
	//"github.com/russross/blackfriday"
	"github.com/json-iterator/go"
	"github.com/boltdb/bolt"
	//"reflect"
	"encoding/binary"
)
const (
	TimeFormat string = "2006-01-02 15:04"
	//TimeFormatL string = "2006-01-02 15:04:05"
	TimeFormatR string = "2006-01-02T15:04:05.000"
	NewsDB string = "NewsList.db"
)
var (
	LastTime string
	Loc *time.Location
	RegS *regexp.Regexp
	ReLower *regexp.Regexp
	ReStyle *regexp.Regexp
	//ReStyleL *regexp.Regexp
	ReScript *regexp.Regexp
	ReZhu *regexp.Regexp
	ReZhuA *regexp.Regexp
	ReA *regexp.Regexp
	ReBody *regexp.Regexp
	//ReImg *regexp.Regexp
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)
func init() {
	var err error
	Loc,err = time.LoadLocation("Etc/GMT-8")
	if err != nil {
		panic(err)
	}
	RegS,err = regexp.Compile("[\\t\\n\\f\\r]")
	if err != nil {
		panic(err)
	}
	ReLower,err = regexp.Compile("\\<[\\S\\s]+?\\>")
	if err != nil {
		panic(err)
	}
	ReStyle, err= regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	if err != nil {
		panic(err)
	}
	ReScript, err = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	if err != nil {
		panic(err)
	}
	//ReImg, err = regexp.Compile("\\<img[\\S\\s]+?\\>")
	//if err != nil {
	//	panic(err)
	//}
	ReZhuA, err = regexp.Compile("[\"]")
	if err != nil {
		panic(err)
	}
	ReZhu, err = regexp.Compile("\\n")
	if err != nil {
		panic(err)
	}
	ReA, err = regexp.Compile("\\<a[\\S\\s]+?\\</a\\>")
	if err != nil {
		panic(err)
	}
	ReBody, err = regexp.Compile("\\<\\!--repaste.body.begin--\\>[\\S\\s]+?\\<\\!--repaste.body.end--\\>")
	if err != nil {
		panic(err)
	}

	go func(){
		for{
			DownNews()
			fmt.Println(time.Now())
			time.Sleep(5*time.Minute)
		}
	}()

}

type News struct {
	//_id		string
	Update		int64
	Title		string
	Description	string
	Content		string
	//Url		string 
	//Type		string 
	DateTime	int64
	Text		string
	Link		[]string
	Sign		int
	//linkId		[]int
}
//func msToTime(ms string) (time.Time, error) {
//	msInt, err := strconv.ParseInt(ms, 10, 64)
//	if err != nil {
//		return time.Time{}, err
//	}
//	tm := time.Unix(0, msInt*int64(time.Millisecond))
//	fmt.Println(tm.Format("2006-02-01 15:04:05.000"))
//	return tm, nil
//}
func NewNews(n map[string]interface{}) (e *News) {

	if n["type"].(string) != "ARTI" {
		return nil
	}
	update := time.Now().UnixNano()
	dt,err := time.ParseInLocation(TimeFormat,n["dateTime"].(string),Loc)
	if err != nil {
		panic(err)
	}
	e = &News{
		Title:n["title"].(string),
		Description:n["description"].(string),
		Content:n["content"].(string),
		DateTime:dt.Unix(),
		//_id:time.Now().Format(TimeFormatR),
		Update:update,
	}
	e.GetText(n["url"].(string))
	return e

}
func (self *News) GetText(u string){

	err := request.ClientHttp(u,"GET",[]int{304,200},nil,func(body io.Reader)error{
		db,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}

		//doc,err := goquery.NewDocumentFromReader(body)
		//if err != nil {
		//	return err
		//}
		//h,err := doc.Find(".cnt_bd").Html()
		////s.RemoveFiltered("script")
		//if err != nil {
		//	return err
		//}

		//h := string(RegS.ReplaceAll(ReBody.Find(db),[]byte{}))
		//h := string(RegS.ReplaceAll(ReBody.Find(db),[]byte{}))
		//h := string(ReBody.Find(db))
		h := strings.Replace(string(ReBody.Find(db)),"\"","",0)
		//h = strings.Replace(h,"\n","",0)
		//h = strings.Replace(h,"\\r","",0)
		h = RegS.ReplaceAllString(h,"")
		h = ReLower.ReplaceAllStringFunc(h, strings.ToLower)
		h = ReStyle.ReplaceAllString(h, "")
		h = ReScript.ReplaceAllString(h, "")
		h = ReA.ReplaceAllString(h, "")
		//fmt.Println(h)
		//self.Text = ReZhu.ReplaceAllString(html2md.Convert(h),"\\\n")
		self.Text = html2md.Convert(h)
		//self.Text = ReZhuA.ReplaceAllString(self.Text, "\\'")
		//self.Text = h
		//self.Text = ReZhuA.ReplaceAllString(self.Text,"")
		//self.Text =html2md.Convert(h)
		//fmt.Println(self.Text)

		return nil
	})
	if err != nil {
		panic(err)
	}
	//fmt.Println(self.Text)

}

//func SetField(obj interface{}, name string, value interface{}) error {
//
//	structValue := reflect.ValueOf(obj).Elem()
//	structFieldValue := structValue.FieldByName(name)
//	if !structFieldValue.IsValid() {
//	    return fmt.Errorf("No such field: %s in obj", name)
//	}
//	if !structFieldValue.CanSet() {
//	    return fmt.Errorf("Cannot set %s field value", name)
//	}
//	structFieldType := structFieldValue.Type()
//	val := reflect.ValueOf(value)
//	if structFieldType != val.Type() {
//		return fmt.Errorf("00")
//	}
//	structFieldValue.Set(val)
//	return nil
//
//}
//func (self *News) loadMap(db map[string]interface{})(err error){
//	for k,v := range db {
//		err = SetField(self,k,v)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}



//func (self *News) toByte() []byte{
//	json.Marshal(self.toMap())
//	self.toMap()
//}
func DownNews(){
	var db map[string]interface{}
	rand.Seed(time.Now().UnixNano())
	u:= fmt.Sprintf("http://news.cctv.com/data/index.json?r=%f",rand.Float64())
	err := request.ClientHttp(u,"GET",[]int{304,200},nil,func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&db)
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//d := db["updateTime"]

	n := GetNewsDBlast()
	fmt.Println(n)
	for _,m := range db["rollData"].([]interface{}){
		news := NewNews(m.(map[string]interface{}))
		if n != nil {
			if news.DateTime <= n.DateTime {
				break
			}
		}
		if news == nil {
			continue
		}
		err = news.SaveToWXDB()
		if err != nil {
			panic(err)
		}
		err = news.SaveTolocalDB()
		if err != nil {
			panic(err)
		}
		fmt.Println(news.Title)
	}




}
func GetNewsDBlast() (n *News) {


	k := make([]byte,8)
	binary.BigEndian.PutUint64(k,uint64(time.Now().UnixNano()))
	db,err := bolt.Open(NewsDB,0600,nil)
	if err != nil {
		panic(err)
	}
	err = db.View(func(tx *bolt.Tx)error{
		b := tx.Bucket([]byte(NewsDB))
		if b == nil {
			return fmt.Errorf("b==nil")
		}
		k,v := b.Cursor().Seek(k)
		if k != nil {
			n = &News{}
			json.Unmarshal(v,n)
		}
		return nil
	})
	db.Close()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return n

}

func (self *News) SaveTolocalDB() error {

	m,err := json.Marshal(self)
	if err != nil {
		return err
	}
	k := make([]byte,8)
	binary.BigEndian.PutUint64(k,uint64(self.Update))
	db,err := bolt.Open(NewsDB,0600,nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx)error{
		b,err := tx.CreateBucketIfNotExists([]byte(NewsDB))
		if err != nil {
			return err
		}
		return b.Put(k,m)
	})

	db.Close()
	//db.Close()
	return err

}
func (self *News) ToWXString() string {

	//htmlFlags := html.CommonFlags | html.HrefTargetBlank
	//opts := html.RendererOptions{Flags: htmlFlags}
	//renderer := html.NewRenderer(opts)
	//text := markdown.ToHTML([]byte(self.Text),nil,renderer)
	//return fmt.Sprintf("{update:%d,link:%d,title:\"%s\",text:\"%s\"}", self.Update,self.Link,self.Title,ReZhuA.ReplaceAll(RegS.ReplaceAll(blackfriday.MarkdownCommon([]byte(self.Text)),[]byte{}),"\\'"))

	//t,err := json.Marshal(self.Text)
	//if err != nil {
	//	panic(err)
	//}
	//t := url.QueryEscape(self.Text)
	//t :=ReZhuA.ReplaceAllString(ReZhu.ReplaceAllString(self.Text,"\\\n"),"\\\"")

	return fmt.Sprintf("{_id:\"%d\",sign:%d,link:%d,title:\"%s\",text:\"%s\"}", self.Update,self.Sign,self.Link,self.Title,url.QueryEscape(self.Text))
	//return fmt.Sprintf("{Title:'%s',DateTime:%d,Link:%d}", self.Title,self.DateTime,self.Link)
	//return fmt.Sprintf("{DateTime:%d,Link:%d}",self.DateTime,self.Link)

}

func (self *News) SaveToWXDB() error {

	//db,err := json.Marshal(self)
	//if err != nil {
	//	panic(err)
	//}
	//return nil

	var res  map[string]interface{}
	err := wxmsg.PostRequest(
		"https://api.weixin.qq.com/tcb/databaseadd",
		map[string]interface{}{
			"query":fmt.Sprintf("db.collection(\"%s\").add({data:[%s]})",config.Conf.CollPageName,self.ToWXString())},
		func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&res)
	})
	if err != nil {
		return err
	}
	if res["errcode"].(float64) != 0 {
		return fmt.Errorf("%.0f %s",res["errcode"].(float64),res["errmsg"].(string))
	}
	//self._id = res["id_list"].([]interface{})[0].(string)
	//fmt.Println(self._id)
	return nil
	//fmt.Println(string(db))

}
