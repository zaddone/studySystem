package chrome
import(
	"io"
	//"os"
	"bytes"
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
	//regW *regexp.Regexp = regexp.MustCompile(`[0-9|a-z|\p{Han}]+`)
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
		keyMap := map[string]int{"vod"+self.Title:1}
		ts := regT.FindAllString(self.Title,-1)
		//self.Title = strings.Join(ts," ")
		for _,l := range ts{
			keyMap[l]+=1
		}
		doc.Find(".vodinfobox li span").Each(func(i int,s *goquery.Selection){
			for _,l := range regT.FindAllString(s.Text(),-1){
				kl := regM.ReplaceAllString(l,"")
				if len(kl) ==0 {
					continue
				}
				keyMap[l]+=1
			}
		})
		for k,_:= range keyMap {
			self.key=append(self.key,strings.ToLower(k))
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
	wb,err := tx_.CreateBucketIfNotExists(WordBucket)
	if err != nil {
		return err
	}
	//wb := tx_.Bucket(WordBucket)


	tx,err := DbPage.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Commit()
	pageb,err := tx.CreateBucketIfNotExists(pageBucket)
	if err != nil {
		return err
	}

	oldid := wb.Get([]byte("vod"+self.Title))

	if len(oldid)>0 {
		db := pageb.Get(oldid)
		p_ := &Pagevod{}
		err := json.Unmarshal(db,p_)
		if err != nil {
			panic(err)
		}
		self.Id = p_.Id
		if !strings.EqualFold(self.Content,p_.Content){
			self.SaveDBBucket(pageb)
			self.getTitleKey(wb)
			self.getTitlePar(wb,pageb)
			return nil
		}
		self.getTitleKey(wb)
		if !self.getTitlePar(wb,pageb){
			return fmt.Errorf("is same %s %s",self.Title,p_.Title)
		}
		self.SaveDBBucket(pageb)
		return nil
		//}
	}
	self.getTitlePar(wb,pageb)
	self.SaveDBBucket(pageb)
	fmt.Println(self.Title)
	var upWord []string
	for _,_k := range self.key{
		if len([]rune(_k)) <2 {
			continue
		}
		k := []byte(_k)
		b_ :=wb.Get(k)
		if len(b_) == 0 {
			b_=self.Id
			//wb.Put(k,self.Id)
		}else{
			b_ = append(b_,self.Id...)
			//wb.Put(k,append(b_,self.Id...))
		}
		lev := len(b_)
		if lev/8 > 20 {
			b_ = b_[8:]
		}
		err := wb.Put(k,b_)
		if err != nil {
			panic(err)
		}
		if strings.HasPrefix(_k,"vod"){
			continue
		}
		nolist := make([]string,0,lev/8)
		for i:=0;i<lev;i+=8{
			nolist = append(nolist,fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(b_[i:i+8])))
		}
		upWord =append(upWord,fmt.Sprintf("{_id:\"%s\",link:[%s]}",_k,strings.Join(nolist,",")))
	}
	WXDBChan<-upWord
	self.getTitleKey(wb)

	return nil
}
func (self *Pagevod)getTitlePar(wb,pageb *bolt.Bucket)bool{
	IdMap := map[string]float64{}
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
			_id := d_[i:I]
			if !bytes.Equal(_id,self.Id){
				IdMap[string(_id)]+=leds/1
			}
			i = I
		}
	}
	if len(IdMap) ==0 {
		return false
	}
	var maxID string
	var max float64 = 0
	for k,v := range IdMap {
		if v > max {
			max = v
			maxID = k
		}
	}
	pid := []byte(maxID)
	if len(self.Par)>0 && bytes.Equal(self.Par,pid){
		return false
	}
	if len(self.Children)>0 && bytes.Contains(self.Children,pid){
		return false
	}
	db := pageb.Get(pid)
	p_ := &Page{}
	err := json.Unmarshal(db,p_)
	if err != nil {
		panic(err)
	}
	self.Par = pid
	p_.Children = append(p_.Children,self.Id...)
	p_.SaveDBBucket(pageb)
	return true


}
func (self *Pagevod)getTitleKey(wb *bolt.Bucket){
	var upWord []string
	key := map[string]int{}
	for _,l := range regT.FindAllString(self.Title,-1){
		lr := regK.FindAllString(l,-1)
		for j:=0;j<len(lr);j++{
			for _j:=j+1;_j<=len(lr);_j++ {
				k :=strings.ToLower(strings.Join(lr[j:_j],""))
				if len([]rune(k))>1{
					key[k]+=1
				}
			}
		}
	}
	keyStr := strings.Join(self.key,",")
	for k_,_ := range key {
		k:=[]byte(k_)
		b_ :=wb.Get(k)
		if len(b_)==0 || bytes.Contains(b_,self.Id) {
			continue
		}
		//self.key
		if !strings.Contains(keyStr,k_){
			self.key = append(self.key,k_)
		}
		b_ = append(b_,self.Id...)

		lev := len(b_)
		out := lev/8 > 20
		if out {
			b_ = b_[8:]
		}
		err := wb.Put(k,b_)
		if err != nil {
			panic(err)
		}
		nolist := make([]string,0,lev/8)
		for i:=0;i<lev;i+=8{
			nolist = append(nolist,fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(b_[i:i+8])))
		}
		upWord =append(upWord,fmt.Sprintf("{_id:\"%s\",link:[%s]}",k_,strings.Join(nolist,",")))

	}
	WXDBChan<-upWord

}

func syncRunPageVod(max int){
	for{
		findPageVod(max)
		<-time.After(1*time.Hour)
	}
}
func findPageVod(max int){
	i:=1
	for c:=0;c<max;{
	//for c:=0;c<30000;{
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
