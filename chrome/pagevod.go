package chrome
import(
	"io"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"github.com/zaddone/studySystem/request"
	"github.com/PuerkitoBio/goquery"
	"github.com/boltdb/bolt"
	"encoding/json"
	"encoding/binary"
)
var (
	regG *regexp.Regexp = regexp.MustCompile("伦理|福利|色情")
	regW *regexp.Regexp = regexp.MustCompile(`[0-9|a-z|\p{Han}]+`)
	regM *regexp.Regexp = regexp.MustCompile(`[0-9]+`)
	regS = regexp.MustCompile(`\S+\$\S+\.m3u8`)
	rootUrl string = "http://www.okzyw.com"
	contentTag string = "vod|"
)
type Pagevod struct{
	Page
	//Update string
	vod []string
	key []string
	//IsVod bool
	//word *bolt.Bucket
	//page *bolt.Bucket
}
func NewPagevod() (v *Pagevod) {

	v = &Pagevod{
		//Id:make([]byte,8),
		//IsVod:true,
	}
	v.Id=make([]byte,8)
	binary.BigEndian.PutUint64(v.Id,uint64(time.Now().UnixNano()))
	return

}
func (self *Pagevod) loadPage(uri string,) error {

	return request.ClientHttp(fmt.Sprintf("%s%s",rootUrl,uri),"GET",[]int{304,200},nil,func(body io.Reader)error{
		doc,err := goquery.NewDocumentFromReader(body)
		if err != nil {
			return err
		}
		self.Title = doc.Find(".vodInfo .vodh h2").Text()
		keyMap := map[string]int{self.Title:1}
		for _,l := range regW.FindAllString(self.Title,-1){
			keyMap[l]+=1
		}
		doc.Find(".vodinfobox li span").Each(func(i int,s *goquery.Selection){
			for _,l := range regW.FindAllString(s.Text(),-1){
				kl := regM.ReplaceAllString(l,"")
				if len(kl) ==0 {
					continue
				}
				keyMap[l]+=1
			}
		})
		for k,_:= range keyMap {
			self.key=append(self.key,k)
		}
		//fmt.Println(self.key)

		tt := doc.Find(".ibox.playBox .vodplayinfo").Text()
		self.vod = regS.FindAllString(tt,-1)
		if len(self.vod) == 0{
			fmt.Println(tt)
			return fmt.Errorf("find Not vod")
		}

		return self.CheckOldVod()

	})
}
func (self *Pagevod) SaveVod(wb,pb *bolt.Bucket)error{

	self.Content = contentTag + strings.Join(self.vod,"|")
	IdMap := map[string]float64{}
	for _,_k := range self.key{
		k := []byte(_k)
		d_ := wb.Get(k)
		if d_ == nil {
			wb.Put(k,self.Id)
			continue
		}
		wb.Put(k,append(d_,self.Id...))
		led := len(d_)
		leds := float64(led)
		for i:=0;i<led;i+=8{
			IdMap[string(d_[i:i+8])]+=leds/1
		}
	}
	if len(IdMap)>0{
		var maxK string
		var maxV float64
		for k,v := range IdMap {
			if v> maxV {
				maxK = k
			}
		}
		err := self.linkBucket([]byte(maxK),pb)
		if err != nil {
			log.Println(err)
		}
	}
	return self.SaveDBBucket(pb)


}

func (self *Pagevod)CheckOldVod()error{

	tx_,err := DbWord.Begin(true)
	if err != nil {
		return err
	}
	defer tx_.Commit()
	wordb := tx_.Bucket(WordBucket)
	tx,err := DbPage.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Commit()
	pageb := tx.Bucket(pageBucket)

	kt := wordb.Get([]byte(self.Title))
	if kt == nil{
		return self.SaveVod(wordb,pageb)
	}
	for i:=0; i<len(kt); i+=8 {
		err = json.Unmarshal(pageb.Get(kt[i:i+8]),self)
		if err != nil{
			log.Println(err)
			continue
		}
		if !strings.HasSuffix(self.Content,contentTag){
			continue
		}
		if len(self.vod)==(len(strings.Split(self.Content,"|"))-1) {
			return fmt.Errorf("is Same")
		}

		self.update = true
		break
	}
	return self.SaveVod(wordb,pageb)

}
func syncRunPageVod(){
	for{
		findPageVod()
		<-time.After(1*time.Hour)
	}
}
func findPageVod(){
	i:=1
	for c:=0;c<20000;{
		err:= getList(i,func(u,d string)error{
			pv := NewPagevod()
			err:=  pv.loadPage(u)
			if err != nil {
				return err
			}
			WXDBPushChan<-pv
			c++
			return nil
		})
		if err == io.EOF {
			break
		}
		i++
	}

}

func getList(page int,readPage func(uri string,datetime string)error) error{
	return request.ClientHttp(fmt.Sprintf("%s/?m=vod-index-pg-%d.html",rootUrl,page),"GET",[]int{304,200},nil,func(body io.Reader)error{
		doc,err := goquery.NewDocumentFromReader(body)
		if err != nil {
			return err
		}
		count :=0
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

			if err = readPage(val,strup);err != nil {
				fmt.Println(err)
				return true
			}
			count++
			return true
		})
		//s.RemoveFiltered("script")
		if count ==0 {
			return io.EOF
		}
		return nil
	})

}
