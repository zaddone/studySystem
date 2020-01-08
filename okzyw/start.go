//package okzyw
package main
import(
	"fmt"
	"io"
	"strings"
	"log"
	"regexp"
	"encoding/json"
	"github.com/zaddone/studySystem/request"
	"github.com/PuerkitoBio/goquery"
	"github.com/boltdb/bolt"
	"github.com/zaddone/studySystem/wxmsg"
)
var (
	rootUrl string = "http://www.okzyw.com"
	regS *regexp.Regexp
	regG *regexp.Regexp
	//TimeFormat string = "2006-01-02 15:04:05"
)
func init(){
	regS = regexp.MustCompile("\\s+")
	regG = regexp.MustCompile("解说|福利|色情")
}
type page struct {
	title string
	Update string
	Li []string
}
func (self *page) liTostring() string{
	lis := make([]string,0,len(self.Li))
	for _,l := range self.Li {
		lis = append(lis,fmt.Sprintf("\"%s\"",l))
	}
	return strings.Join(lis,",")
}
func (self *page) toAddString() string{
	return fmt.Sprintf("{_id:\"%s\",li:[%s]}",self.title,self.liTostring())
}
func (self *page) toUpdateString() string{
	return fmt.Sprintf("{li:[%s]}",self.liTostring())
}
func getList(page int,readPage func(uri string,datetime string)error) error{
	return request.ClientHttp(fmt.Sprintf("%s/?m=vod-index-pg-%d.html",rootUrl,page),"GET",[]int{304,200},nil,func(body io.Reader)error{
		doc,err := goquery.NewDocumentFromReader(body)
		if err != nil {
			return err
		}
		doc.Find(".xing_vb li").EachWithBreak(func(i int,s *goquery.Selection)bool {
			if regG.MatchString(s.Find("span.xing_vb5").Text()){
				//fmt.Println(s.Find("span.xing_vb4").Text())
				return true
			}
			val,ok := s.Find("span.xing_vb4 a").Attr("href")
			if !ok{
				return true
			}
			strup := s.Find("span.xing_vb6").Text()
			if strup=="" {
				strup = s.Find("span.xing_vb7").Text()
				if strup == "" {
					return true
				}
			}

			err = readPage(val,strup)
			if err != nil {
				if err == io.EOF {
					return false
				}
				fmt.Println(err)
			}
			return true
		})
		//s.RemoveFiltered("script")

		return nil
	})

}


func main(){
	i:=1
	var err error
	db,err := bolt.Open("okzyw.db",0600,nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	tx,err := db.Begin(true)
	if err != nil {
		panic(err)
	}
	defer tx.Commit()
	b,err := tx.CreateBucketIfNotExists([]byte("list"))
	if err != nil {
		panic(err)
	}
	for{
		err := getList(i,func(uri string,update string)error{
			return request.ClientHttp(fmt.Sprintf("%s%s",rootUrl,uri),"GET",[]int{304,200},nil,func(body io.Reader)error{
				doc,err := goquery.NewDocumentFromReader(body)
				if err != nil {
					return err
				}

				p := &page{}
				p.title = doc.Find(".vodInfo .vodh h2").Text()
				db:= b.Get([]byte(p.title))
				isUpdate:= false
				if db != nil {
					err = json.Unmarshal(db,p)
					if err != nil {
						return err
					}
					if p.Update == update {
						return io.EOF
					}
					isUpdate = true
				}
				p.Update = update
				p.Li = nil

				doc.Find(".ibox.playBox .vodplayinfo ul li").EachWithBreak(func(i int,s *goquery.Selection)bool{
					t :=regS.ReplaceAllString(s.Text(),"")
					if len(t) < 10  {
						return true
					}
					t = strings.ToLower(t)
					//fmt.Println(t)
					if strings.HasSuffix(t,".m3u8"){
						p.Li = append(p.Li,t)
					}
					return true
				})

				if !isUpdate {
					err = wxmsg.AddToWXDB("vod",p.toAddString())
					if err != nil {
						return err
					}

				}else{
					wxmsg.UpdateWXDB("vod",p.title,p.toUpdateString())
					if err != nil {
						return err
					}
				}
				fmt.Println(p.title,p.Update)
				db,err = json.Marshal(p)
				if err != nil {
					panic(err)
				}
				return b.Put([]byte(p.title),db)
				//return nil
			})
		})
		//fmt.Println(err)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println(err)
		}
		i++
	}

}

