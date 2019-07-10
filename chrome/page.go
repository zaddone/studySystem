package chrome
import(
	"fmt"
	"time"
	"log"
	"strings"
	"regexp"
	"encoding/binary"
	"encoding/json"
	"net/url"
	//"sync"
	"github.com/boltdb/bolt"
	"github.com/zaddone/studySystem/config"
	//"math"
	//"sort"
)

var (
	WordDB = "word.db"
	PageDB = "page.db"
	pageBucket = []byte("page")
	WordBucket = []byte("word")
	regTitle *regexp.Regexp = regexp.MustCompile(`[\p{Han}]+`)

)
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
	relevant []byte
	//class []byte
}
func ViewPageBucket(Bucket []byte,hand func(*bolt.Bucket)error) error {
	db,err := bolt.Open(PageDB,0600,nil)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.View(func(t *bolt.Tx)error{
		b := t.Bucket(Bucket)
		if b == nil {
			return fmt.Errorf("b==nil")
		}
		return hand(b)
	})
}

//
//func (self *Page)LoadPage(id []byte,handle func([]byte, *bolt.Bucket)) error {
//
//	db,err := bolt.Open(PageDB,0600,nil)
//	if err != nil {
//		return err
//	}
//	defer db.Close()
//	return db.View(func(t *bolt.Tx)error{
//		b := t.Bucket(pageBucket)
//		if b == nil {
//			return fmt.Errorf("b==nil")
//		}
//		return json.Unmarshal(b.Get(id),self)
//	})
//
//}
func findSetPage(id []byte,b *bolt.Bucket,handle func(*Page)) (p *Page) {

	p = &Page{}
	err := json.Unmarshal(b.Get(id),p)
	if err != nil {
		return nil
	}
	handle(p)
	for i:=0;i< len(p.Children);i+=8{
		 findSetPage(p.Children[i:i+8],b,handle)
	}
	return p

}


func (self *Page) link(lid []byte) error {
	var p *Page
	err := ViewPageBucket(pageBucket,func(b *bolt.Bucket)error {
		p = findSetPage(
			lid,
			b,
			func(p *Page){
				self.relevant = append(self.relevant,p.Id...)
			},
		)
		if p == nil {
			return fmt.Errorf("Not Find Page")
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



func NewPage(title,content string) (p *Page) {
	p = &Page{
		//Id:time.Now().UnixNano(),
		Title:title,
		Content:content,
		//class:[]byte(class),
	}
	p.Id = make([]byte,8)
	binary.BigEndian.PutUint64(p.Id,uint64(time.Now().UnixNano()))
	return
}


func (self *Page) ToWXString() (string,[]string) {
	var link,_link []string
	for i:=0;i<len(self.relevant);i+=8{
		link = append(
		link,
		fmt.Sprintf("\"%d\"",
		binary.BigEndian.Uint64(self.relevant[i:i+8])))
	}
	if len(link)>10{
		_link = link[:10]
	}else{
		_link = link
	}
	//fmt.Println(link)
	return fmt.Sprintf("{_id:\"%d\",link:%s,title:\"%s\",text:\"%s\"}", binary.BigEndian.Uint64(self.Id),fmt.Sprintf("[%s]",strings.Join(_link,",")),self.Title,url.QueryEscape(self.Content)),link

}


func (self *Page) SaveDB() error {
	v,err := json.Marshal(self)
	if err != nil {
		return err
	}
	db,err := bolt.Open(PageDB,0600,nil)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Update(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(pageBucket)
		if err != nil {
			return err
		}
		return b.Put(self.Id,v)
	})

}


//func addWork(k string,tx *bolt.Tx,W map[string][]byte){
//	//lk := len([]rune(k))
//	//if lk>255 {
//	//	lk = 255
//	//}
//	b := tx.Bucket(WordBucket)
//	if b != nil {
//		db := b.Get([]byte(k))
//		if db != nil {
//			W[k] = db
//			return
//		}
//	}
//	W[k] = []byte{}
//	//for i:=1 ; i<lk ; i++{
//	//	b := tx.Bucket([]byte{byte(i)})
//	//	b_ := tx.Bucket([]byte{byte(lk-i)})
//	//	handWork(
//	//		b,
//	//		b_,
//	//		[]byte(string([]rune(k)[i:])),
//	//		[]byte(string([]rune(k)[:i])),
//	//		W,
//	//	)
//	//	handWork(
//	//		b_,
//	//		b,
//	//		[]byte(string([]rune(k)[:i])),
//	//		[]byte(string([]rune(k)[i:])),
//	//		W,
//	//	)
//	//}
//}

//func handWork(b,b_ *bolt.Bucket,key,key_ []byte,W_ map[string][]byte){
//
//	if b == nil {
//		return
//	}
//	val := b.Get(key)
//	if val == nil {
//		return
//	}
//	W_[string(key)] = val
//	if b_ == nil {
//		return
//	}
//	val = b_.Get(key_)
//	if val == nil {
//		val = []byte{}
//	}
//	W_[string(key_)] = val
//
//}

func (self *Page) CheckUpdateWork() error {

	work := self.getSplitWord()
	//fmt.Println(work)
	if len(work)<100 {
		return fmt.Errorf("work < 100")
	}
	db,err := bolt.Open(WordDB,0600,nil)
	if err != nil {
		return err
	}
	defer db.Close()
	W :=map[string][]byte{}
	//for k,_ := range work {
	//	W[k] = []byte{}
	//}
	//lenWord:=len(work)
	db.View(func(tx *bolt.Tx)error{
		b := tx.Bucket(WordBucket)
		if b != nil {
			for k,_ := range work {
				W[k] = b.Get([]byte(k))
			}
		}else{
			for k,_ := range work {
				W[k] = nil
			}
		}
		return nil
	})
	vm := map[string]int{}
	vm_ := map[string]float64{}
	for _,v := range W {
		le := len(v)
		lev := float64(le/8)
		//fmt.Println(lev)
		for i:=0;i<le;i+=8{
			v_ := string(v[i:i+8])
			vm[v_]+=1
			vm_[v_]+= 1.0/lev
		}
	}
	var max int
	for _,v := range vm {
		//fmt.Println(binary.BigEndian.Uint64([]byte(k)),v)
		if v>max{
			max = v
		}
	}
	fmt.Println(max,len(vm),len(W))
	if max >= len(W) {
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

	err = self.link([]byte(maxk))
	if err != nil {
		log.Println(err)
	}

	return db.Update(func(tx *bolt.Tx)error{

		b,err := tx.CreateBucketIfNotExists(WordBucket)
		if err != nil {
			return err
		}
		for k,v := range W {
			//lk := len([]rune(k))
			//if lk>255 {
			//	lk = 255
			//}

			err:= b.Put([]byte(k),append(v,self.Id...))
			if err != nil {
				return err
			}

		}
		return nil
	})

}

func split_(li []string)(map[string]int){

	key := map[string]int{}
	for _,l := range li {
		lr := []rune(l)
		for j:=0;j<len(lr);j++{
			for _j:=j+2;_j<len(lr);_j++{
				key[string(lr[j:_j])]+=1
			}
		}
	}

	for k,v := range key {
		if v<=1 {
			delete(key,k)
			continue
		}
		//fmt.Println(k,v)
	}
	//fmt.Println(len(key))
	//for k,v := range key{
	//	fmt.Println(k,v)
	//}
	return key

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
	return split_(append(regTitle.FindAllString(self.Title,-1),regTitle.FindAllString(self.Content,-1)...))

	//m = map[string]int{}
	//li := append(regTitle.FindAllString(self.Title,-1),regTitle.FindAllString(self.Content,-1)...)
	//var count,begin int
	//for i,l := range li{
	//	count += len([]rune(l))
	//	if count<150 {
	//		continue
	//	}
	//	for k,v := range split(li[begin:i]){
	//		if len([]rune(k))>1{
	//			m[k]+=v
	//		}
	//	}
	//	count=0
	//	begin=i
	//}
	//for k,v := range m {
	//	fmt.Println(k,v)
	//}
	//return m
}
//func (self *Page) splitWord(h func(k string ,v int)) {
//	m := map[string]int{}
//	li := regTitle.FindAllString(self.Title,-1)
//	li = append(li,regTitle.FindAllString(self.Content,-1))
//	//m := split(li)
//	for k,v := split(li) {
//		h(k,v)
//	}
//
//}
