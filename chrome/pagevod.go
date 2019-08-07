package chrome
import(
	"io"
	//"os"
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
		//page.Tag:"vod",
		//Id:make([]byte,8),
		//IsVod:true,
	}
	v.Tag = "vod"
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
		keyMap := map[string]int{}
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

		return self.CheckToSaveVod()

	})
}
func (self *Pagevod) SaveVod(wb,pb *bolt.Bucket)error{

	self.Content = contentTag + strings.Join(self.vod,"|")
	if wb == nil {
		return self.SaveDBBucket(pb)
	}
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
func (self *Pagevod)CheckToSaveVod()error{
	self.Content = contentTag + strings.Join(self.vod,"|")
	tx_,err := DbWord.Begin(true)
	if err != nil {
		return err
	}
	defer tx_.Commit()
	tx,err := DbPage.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Commit()

	wb := tx_.Bucket(WordBucket)
	IdMap := map[string]float64{}
	IdMapN := map[string]int{}
	for _,_k := range self.key{
		k := []byte(_k)
		d_ := wb.Get(k)
		if d_ == nil {
			//IdMap[_k]
			//wb.Put(k,self.Id)
			continue
		}
		//wb.Put(k,append(d_,self.Id...))
		led := len(d_)
		leds := float64(led)
		for i:=0;i<led;{
			I := i+8
			IdMap[string(d_[i:I])]+=leds/1
			IdMapN[string(d_[i:I])]++
			i = I
		}
	}
	var maxN int = 0
	var maxID string
	for k,v := range IdMapN {
		if v > maxN {
			maxN = v
			maxID = k
		}
	}


	pageb := tx.Bucket(pageBucket)
	if maxN == len(self.key){
		db := pageb.Get([]byte(maxID))
		p_ := &Page{}
		err := json.Unmarshal(db,p_)
		if err != nil {
			panic(err)
		}
		if strings.EqualFold(self.Title,p_.Title){
			if len(self.Content) > len(p_.Content) {
				p_.Content = self.Content
				fmt.Println(p_.Title)
				p_.SaveDBBucket(pageb)
				return nil
			}else{
				return fmt.Errorf("is same %s %s",self.Title,p_.Title)
			}
		}
	}

	var max float64 = 0
	for k,v := range IdMap {
		if v > max {
			max = v
			maxID = k
		}
	}
	pid := []byte(maxID)
	db := pageb.Get(pid)
	p_ := &Page{}
	err = json.Unmarshal(db,p_)
	if err != nil {
		panic(err)
	}
	self.Par = pid
	p_.Children = append(p_.Children,self.Id...)
	fmt.Println(self.Title)
	self.SaveDBBucket(pageb)
	p_.SaveDBBucket(pageb)
	for _,_k := range self.key{
		k := []byte(_k)
		b_ :=wb.Get(k)
		if len(b_) == 0 {
			wb.Put(k,self.Id)
		}else{
			wb.Put(k,append(b_,self.Id...))
		}
	}
	return nil
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
			//fmt.Println(u,d)
			pv := NewPagevod()
			err:=  pv.loadPage(u)
			if err != nil {
				return err
			}
			c++
			//fmt.Println(pv)
			//body,ids := pv.ToWXString()
			//fmt.Println(pv.Title)
			//WXDBChan<-&UpdateId{pv.GetId(),ids}
			//f,err := os.OpenFile(config.Conf.CollPageName,os.O_APPEND|os.O_CREATE|os.O_RDWR,0777)
			//if err != nil{
			//	return err
			//}
			//defer f.Close()
			//_,err = f.WriteString(pv.ToWXString())
			//if err != nil {
			//	return err
			//}
			return pv.SaveToList()
		})
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		i++
	}

}

func getList(page int,readPage func(uri string,datetime string)error) error{
	//fmt.Println("page",page)
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
