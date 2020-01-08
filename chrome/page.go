package chrome
import(
	"fmt"
	"time"
	"log"
	"strings"
	"bytes"
	"regexp"
	"encoding/binary"
	"encoding/json"
	"net/url"
	//"os"
	//"sync"
	"github.com/boltdb/bolt"
	"github.com/zaddone/studySystem/config"
	//"math"
	//"sort"
)

var (
	regTitle *regexp.Regexp = regexp.MustCompile(`[\p{Han}]+`)
	regT *regexp.Regexp = regexp.MustCompile(`[0-9|a-z|A-Z|\p{Han}]+`)
	regK *regexp.Regexp = regexp.MustCompile(`[0-9a-zA-Z]+|\p{Han}`)
)

func clearLocalDB(hand func([]string,[]string)error) error {

	//db,err := bolt.Open(PageDB,0600,nil)
	//if err != nil {
	//	return err
	//}
	//defer db.Close()
	tx,err := DbPage.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Commit()
	//tx.Commit()
	b := tx.Bucket(pageListBucket)
	if b == nil {
		return fmt.Errorf("b == nil")
	}
	li := b.Get([]byte("page"))
	pli := len(li)  - config.Conf.MaxPage*8

	if pli < 1  {
		return fmt.Errorf("%d",pli)
	}

	klink:=li[:pli]
	var klinkStr []string
	cTag := []byte(contentTag)
	for i:=0;i<pli;{
		I := i+8
		k := li[i:I]
		cou := b.Get(k)
		if cou== nil {
			continue
		}
		if bytes.Contains(cou,cTag){
			continue
		}
		err = b.Delete(k)
		if err != nil {
			panic(err)
		}
		klinkStr = append(klinkStr,fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(k)))

		//err = b.Delete(k)
		i = I
	}
	b.Put([]byte("page"),li[pli:])

	tx_,err := DbWord.Begin(true)
	if err != nil {
		return err
	}
	defer tx_.Commit()
	b_ := tx_.Bucket(WordBucket)
	if b_ == nil {
		return fmt.Errorf("b == nil")
	}
	var saveKey []string
	c_ := b_.Cursor()
	var klinkWord []string
	for k,v := c_.First();k!= nil;k,v = c_.Next(){
		vlen := len(v)
		var nv []byte
		nolist := make([]string,0,vlen/8)
		for i:=0;i<vlen;i+=8{
			_v := v[i:i+8]
			if !bytes.Contains(klink,_v){
				nv=append(nv,_v...)
				nolist = append(nolist,fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(_v)))
			}
		}
		lenv := len(nv)
		if lenv == 0 {
			//fmt.Println("-",string(k))
			b_.Delete(k)
			if lenv/8 < 50 {
				klinkWord = append(klinkWord,fmt.Sprintf("\"%s\"",string(k)))
			}
		}else{
			if lenv != vlen {
				b_.Put(k,nv)
				if !bytes.HasPrefix(nv,[]byte("vod")){
					saveKey = append(saveKey,fmt.Sprintf("{_id:\"%s\",link:[%s]}",string(k),strings.Join(nolist,",")))
				}
			}
		}
	}
	WXDBChan<-saveKey
	//f.Close()
	return hand(klinkStr,klinkWord)
	//fmt.Println("hand",err)

}
func reverse(s string) (s_ []rune) {
	s_ = []rune(s)
	for i, j := 0, len(s_)-1; i < j; i, j = i+1, j-1 {
		s_[i], s_[j] = s_[j], s_[i]
	}
	return s_
}

type Page struct {

	Id []byte
	Title string
	Content string
	Par []byte
	Children []byte
	//relevant []byte
	//class []byte
	update bool
	Tag string
}
func (self *Page) GetTitle() string{
	return self.Title
}
func(self *Page) GetId() uint64 {
	return binary.BigEndian.Uint64(self.Id)
}
func (self *Page)GetUpdate() bool {
	return self.update
}
//func getWord() (wordlist map[string][]string,err error) {
//	wordlist = make(map[string][]string)
//	err = DbWord.Update(func(tx *bolt.Tx)error{
//		b := tx.Bucket(WordBucket)
//		if b == nil {
//			return fmt.Errorf("b = nil")
//		}
//		c := b.Cursor()
//		for k,v := c.First(); k!= nil ; k,v = c.Next() {
//			le := len(v)
//			//lev := le/8
//			noId := map[string][]byte{}
//			for i:=0;i<le;i+=8 {
//				pid := v[i:i+8]
//				noId[fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(pid))] = pid
//			}
//
//			nolist := make([]string,0,len(noId))
//			v_ := make([]byte,0,le)
//			for k,t := range noId {
//				nolist = append(nolist,k)
//				v_ = append(v_,t...)
//			}
//			if len(v_) < le {
//				b.Put(k,v_)
//			}
//			if len(noId)>50 {
//				continue
//			}
//			wordlist[string(k)] = nolist
//		}
//		return nil
//	})
//	return
//
//}
//func getWord_() (wordlist map[string][]string,err error) {
//	wordlist = make(map[string][]string)
//	err = DbWord.View(func(tx *bolt.Tx)error{
//		b := tx.Bucket(WordBucket)
//		if b == nil {
//			return fmt.Errorf("b = nil")
//		}
//		c := b.Cursor()
//		for k,v := c.First(); k!= nil ; k,v = c.Next() {
//			le := len(v)
//			lev := le/8
//			if lev>50 {
//				continue
//			}
//			var noId []string
//			for i:=0;i<le;i+=8 {
//				pid := v[i:i+8]
//				noId = append(
//					noId,
//				fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(pid)) )
//			}
//			wordlist[string(k)] = noId
//		}
//		return nil
//	})
//	return
//
//}

func ViewPageBucket(Bucket []byte,hand func(*bolt.Bucket)error) error {
	return DbPage.View(func(t *bolt.Tx)error{
		b := t.Bucket(Bucket)
		if b == nil {
			return fmt.Errorf("b==nil")
		}
		return hand(b)
	})
}

func DelPage(id []byte) (err error) {
	return DbPage.Batch(func(tx *bolt.Tx)error{
		b := tx.Bucket(pageBucket)
		if b== nil {
			return fmt.Errorf("b == nil")
		}
		return b.Delete(id)
	})
}

func findSetPage(id []byte,b *bolt.Bucket,handle func(*Page)bool) (p *Page) {
	pagedb := b.Get(id)
	if pagedb == nil {
		return
	}
	p = &Page{}
	err := json.Unmarshal(pagedb,p)
	if err != nil {
		return nil
	}
	if !handle(p){
		return nil
	}
	for i:=0;i< len(p.Children);i+=8{
		 findSetPage(p.Children[i:i+8],b,handle)
	}
	return p

}

func (self *Page) linkBucket(lid []byte,b *bolt.Bucket) error {


	pagedb := b.Get(lid)
	p := &Page{}
	err := json.Unmarshal(pagedb,p)
	if err != nil {
		return err
	}

	self.Par = lid
	//self.relevant = append(p.Id,p.Children...)
	p.Children = append(p.Children,self.Id...)
	return p.SaveDBBucket(b)
}

func (self *Page) link(lid []byte) error {
	var p *Page
	err := ViewPageBucket(pageBucket,func(b *bolt.Bucket)error {
		pagedb := b.Get(lid)
		if pagedb == nil {
			return fmt.Errorf("find not page")
		}
		p = &Page{}
		err := json.Unmarshal(pagedb,p)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	self.Par = lid
	p.Children = append(p.Children,self.Id...)
	return p.SaveDB()
}


func NewPage(title,content,tag string) (p *Page) {
	p = &Page{
		//Id:time.Now().UnixNano(),
		Title:title,
		Content:content,
		Id:make([]byte,8),
		Tag:tag,
		//class:[]byte(class),
	}
	binary.BigEndian.PutUint64(p.Id,uint64(time.Now().UnixNano()))
	return
}


func (self *Page) ToWXString() (string) {
	var link []string
	for i:=0;i<len(self.Children);i+=8{
		link = append(
		link,
		fmt.Sprintf("\"%d\"",
		binary.BigEndian.Uint64(self.Children[i:i+8])))
	}
	//if len(link)>10{
	//	link = link[:10]
	//}
	//fmt.Println(link)
	par :=""
	if len(self.Par)>0{
		par = fmt.Sprintf("%d",binary.BigEndian.Uint64(self.Par))
	}

	return fmt.Sprintf(
		"{_id:\"%d\",par:\"%s\",children:[%s],title:\"%s\",text:\"%s\"}",
		binary.BigEndian.Uint64(self.Id),
		par,
		strings.Join(link,","),
		strings.Join(regT.FindAllString(self.Title,-1)," "),
		url.QueryEscape(self.Content))

}
func (self *Page) SaveToList()error{
	return DbPage.Update(func(tx *bolt.Tx)error{
		b,err := tx.CreateBucketIfNotExists(pageListBucket)
		if err != nil {
			return err
		}
		tag := []byte(self.Tag)
		return b.Put(tag,append(b.Get([]byte(self.Tag)),self.Id...,))
	})
}

func (self *Page) SaveDBBucket(b *bolt.Bucket) error {
	v,err := json.Marshal(self)
	if err != nil {
		return err
	}

	WXDBChan<-self.ToWXString()

	return b.Put(self.Id,v)
}

func (self *Page) SaveDB() error {

	v,err := json.Marshal(self)
	if err != nil {
		return err
	}
	err = DbPage.Update(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(pageBucket)
		if err != nil {
			return err
		}
		return b.Put(self.Id,v)
	})
	if err != nil {
		return err
	}

	WXDBChan <- self.ToWXString()

	return self.SaveToList()

}



func (self *Page) CheckUpdateWork() error {

	work := self.getSplitWord()
	if len(work) == 0 {
		return fmt.Errorf("work = 0")
	}
	W :=map[string][]byte{}
	DbWord.View(func(tx *bolt.Tx)error{
		b := tx.Bucket(WordBucket)
		if b != nil {
			c := b.Cursor()
			for k,_ := range work {
				k__ := []byte(k)
				k_,v_ := c.Seek(k__)
				if bytes.Contains(k_,k__){
					W[k] = v_
				}else{
					W[k] = []byte{}
				}
			}
		}else{
			for k,_ := range work {
				W[k] = []byte{}
			}
		}
		return nil
	})
	vm := map[string]int{}
	vm_ := map[string]float64{}
	for k,v := range W {
		le := len(v)
		lev := float64(le/8)
		for i:=0;i<le;i+=8{
			v_ := string(v[i:i+8])
			vm[v_]+=1
			vm_[v_]+= 1.0/lev * float64(len(k))
		}
	}
	var max int
	for _,v := range vm {
		if v>max{
			max = v
		}
	}
	//fmt.Println(max,len(vm),len(W))
	if (float64(max)/float64(len(W))) > 0.7 {
		return fmt.Errorf("is Same %d %d",max,len(W))
	}
	var max_ float64
	var maxk string
	for k,v := range vm_ {
		if v>max_ {
			max_ = v
			maxk = k
		}
	}

	err := self.link([]byte(maxk))
	if err != nil {
		log.Println(err)
	}
	var upWord []string

	err = DbWord.Update(func(tx *bolt.Tx)error{

		b,err := tx.CreateBucketIfNotExists(WordBucket)
		if err != nil {
			return err
		}
		for k,v := range W {
			//lk := len([]rune(k))
			//if lk>255 {
			//	lk = 255
			//}

			if bytes.Contains(v,self.Id){
				continue
			}
			v = append(v,self.Id...)
			lev := len(v)
			if lev/8 > 20 {
				v= v[8:]
				//continue
			}
			nolist := make([]string,0,lev/8)
			for i:=0;i<lev;i+=8{
				v_ := v[i:i+8]
				nolist = append(nolist,fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(v_)))
			}
			//if !strings.HasPrefix(k,"vod"){
			if len(nolist)>0{
			upWord =append(upWord,fmt.Sprintf("{_id:\"%s\",link:[%s]}",k,strings.Join(nolist,",")))
			}
			err:= b.Put([]byte(k),v)
			if err != nil {
				return err
			}
			//}


		}
		return nil
	})
	if err != nil {
		return err
	}
	WXDBChan <- upWord

	return nil

}

func split_(li []string)(map[string]int){

	key := map[string]int{}
	for _,l := range li {
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
	nkey := map[string]int{}
	for k,v := range key {
		if v<=1 {
			delete(key,k)
			continue
		}
		nkey[k] = v
		//fmt.Println(k,v)
	}
	G:
	for k,v := range nkey {
		//delete(key,k)
		for _k,_v := range key {
			if len(k) >= len(_k) {
				continue
			}
			if strings.Contains(_k,k) && (v==_v) {
				//fmt.Println(_k,k)
				delete(nkey,k)
				continue G
			}
		}
	}
	//fmt.Println(len(key))
	//for k,v := range nkey{
	//	fmt.Println(k,v)
	//}
	return nkey

}
func split(li []string)(key map[string]int){

	key = map[string]int{}
	work := map[int][]bool{}
	//li := reg.FindAllString(m,-1)
	for i,l := range li {
		work[i] = make([]bool,len([]rune(l)))
	}
	//fmt.Println(li)
	le := len(li)
	//var list [][2]int
	for i:=0;i<le;i++ {
		s_bak := []rune(li[i])
		leb := len(s_bak)
		I := i+1
		_li := li[I:]
		if len(_li) == 0 {
			break
			//work[string(s_bak)] += 1
		}
		for j:=0;j<leb;j++{
		//for j,sk := range s_bak{
			sk := s_bak[j]
			for _i,_s := range _li {
				t := strings.IndexRune(_s,sk)
				//t = []rune(_s[:t])
				if t<=0 {
					continue
				}
				_j := len([]rune(_s[:t]))
				work[i][j] = true
				work[I+_i][_j] = true
				//work[i] = work[i],j)
				//work[_i] = append(work[_i],len([]rune(_s[:t])))
				//fmt.Println(string([]rune(_s)[_j:]),string(s_bak[j:]))
				//j = j+_t-1
			}
		}
	}
	for k,v := range work {
		//fmt.Println(k,v,li[k])
		ls := []rune(li[k])
		var list []string
		var str string
		for i,l := range ls {
			if v[i] {
				if str != "" {
					list = append(list,str)
					str = ""
				}
				list = append(list,string(l))
			}else{
				str += string(l)
			}
		}
		if str != "" {
			list = append(list,str)
		}
		//fmt.Println(list)
		le:= len(list)
		for i:=0;i<le;i++{
			s:= list[i]
			//key[string(reverse(s))]+=1
			key[s]+=1
			for j:=i+1 ; j < le ; j++ {
				s+=list[j]
				//key[string(reverse(s))]+=1
				key[s]+=1
			}
		}
	}
	return

}

func (self *Page) getSplitWord() (m map[string]int) {

	if float64(len(self.Title))/float64(len(self.Content)) > 0.1 {
		return
	}
	titl := regT.FindAllString(self.Title,-1)
	newTi := make([]string,0,len(titl))
	for _,t := range titl{
		if len(t)>2 {
			newTi = append(newTi,t)
		}
	}
	//self.Title = strings.Join(titl," ")
	//for 
	outkey,err :=regexp.Compile(config.Conf.OutKey)
	if err != nil {
		panic(err)
	}
	var con []string
	G:
	for _,s := range strings.Split(self.Content,"\n"){
		if len(s)<2 {
			con = append(con,s)
			continue
		}
		if outkey.MatchString(s){
			fmt.Println("key",s)
			continue G
		}
		con = append(con,s)
	}
	self.Content = strings.Join(con,"\n")
	return split_(append(newTi,regT.FindAllString(self.Content,-1)...))

}
