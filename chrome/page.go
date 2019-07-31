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
	//"sync"
	"github.com/boltdb/bolt"
	"github.com/zaddone/studySystem/config"
	//"math"
	//"sort"
)

var (
	regTitle *regexp.Regexp = regexp.MustCompile(`[\p{Han}]+`)
)
func clearLocalDB(hand func([]string)error) error {

	//db,err := bolt.Open(PageDB,0600,nil)
	//if err != nil {
	//	return err
	//}
	//defer db.Close()
	tx,err := DbPage.Begin(true)
	if err != nil {
		return err
	}
	//tx.Commit()
	b := tx.Bucket(pageBucket)
	if b == nil {
		return fmt.Errorf("b == nil")
	}
	c := b.Cursor()
	//p := &Page{}
	var klink []byte
	var klinkStr []string
	p:= &Page{}
	for k,v := c.First();k!=nil&&len(klinkStr)<100;k,v = c.Next(){
		err = json.Unmarshal(v,p)
		if err == nil {
			if strings.HasSuffix(p.Content,contentTag){
				continue
			}
		}
		klink = append(klink,k...)
		klinkStr = append(klinkStr,fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(k)))
		err = b.Delete(k)
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println(klinkStr)
	//db_,err := bolt.Open(WordDB,0600,nil)
	//if err != nil {
	//	return err
	//}
	//defer db_.Close()
	tx_,err := DbWord.Begin(true)
	if err != nil {
		return err
	}
	b_ := tx_.Bucket(WordBucket)
	if b_ == nil {
		return fmt.Errorf("b == nil")
	}
	c_ := b_.Cursor()
	for k,v := c_.First();k!= nil;k,v = c_.Next(){
		vlen := len(v)
		var nv []byte
		for i:=0;i<vlen;i+=8{
			_v := v[i:i+8]
			if !bytes.Contains(klink,_v){
				nv=append(nv,_v...)
			}
		}
		lenv := len(nv)
		if lenv == 0 {
			//fmt.Println("-",string(k))
			b_.Delete(k)
		}else{
			if lenv != vlen {
				b_.Put(k,nv)
			}
		}
	}
	err = hand(klinkStr)
	//fmt.Println("hand",err)
	if err != nil {
		return err
	}
	tx_.Commit()
	tx.Commit()
	return nil

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
	relevant []byte
	//class []byte
	update bool
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
func getWord() (wordlist map[string][]string,err error) {
	wordlist = make(map[string][]string)
	err = DbWord.View(func(tx *bolt.Tx)error{
		b := tx.Bucket(WordBucket)
		if b == nil {
			return fmt.Errorf("b = nil")
		}
		c := b.Cursor()
		for k,v := c.First(); k!= nil ; k,v = c.Next() {
			le := len(v)
			lev := le/8
			if lev < 2  || lev>50 {
				continue
			}
			var noId []string
			for i:=0;i<le;i+=8 {
				pid := v[i:i+8]
				noId = append(
					noId,
				fmt.Sprintf("\"%d\"",binary.BigEndian.Uint64(pid)) )
			}
			wordlist[string(k)] = noId
		}
		return nil
	})
	return

}

func ViewPageBucket(Bucket []byte,hand func(*bolt.Bucket)error) error {
	return DbPage.View(func(t *bolt.Tx)error{
		b := t.Bucket(Bucket)
		if b == nil {
			return fmt.Errorf("b==nil")
		}
		return hand(b)
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

	P := findSetPage(
		lid,
		b,
		func(p *Page) bool {
			self.relevant = append(self.relevant,p.Id...)
			return len(self.relevant)<=10
		},
	)
	if P == nil {
		return fmt.Errorf("Not Find Page")
	}
	self.Par = lid

	P.Children = append(P.Children,self.Id...)
	return P.SaveDBBucket(b)
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
		//self.relevant = append(self.relevant,p.Id...)
		//if !handle(p){
		//	return nil
		//}
		self.relevant = append(p.Id,p.Children...)
		//p = findSetPage(
		//	lid,
		//	b,
		//	func(p *Page)bool{
		//		self.relevant = append(self.relevant,p.Id...)
		//		return len(self.relevant)<=10
		//	},
		//)
		//if p == nil {
		//	return fmt.Errorf("Not Find Page")
		//}
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
		Id:make([]byte,8),
		//class:[]byte(class),
	}
	binary.BigEndian.PutUint64(p.Id,uint64(time.Now().UnixNano()))
	return
}


func (self *Page) ToWXString() (string,[]string) {
	var link []string
	for i:=0;i<len(self.relevant);i+=8{
		link = append(
		link,
		fmt.Sprintf("\"%d\"",
		binary.BigEndian.Uint64(self.relevant[i:i+8])))
	}
	if len(link)>10{
		link = link[:10]
	}
	//fmt.Println(link)
	return fmt.Sprintf(
		"{_id:\"%d\",link:[%s],title:\"%s\",text:\"%s\"}",
		binary.BigEndian.Uint64(self.Id),
		strings.Join(link,","),
		self.Title,
		url.QueryEscape(self.Content)),link

}

func (self *Page) SaveDBBucket(b *bolt.Bucket) error {
	v,err := json.Marshal(self)
	if err != nil {
		return err
	}
	return b.Put(self.Id,v)
}

func (self *Page) SaveDB() error {
	v,err := json.Marshal(self)
	if err != nil {
		return err
	}
	//db,err := bolt.Open(PageDB,0600,nil)
	//if err != nil {
	//	return err
	//}
	//defer db.Close()
	return DbPage.Update(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(pageBucket)
		if err != nil {
			return err
		}
		return b.Put(self.Id,v)
	})

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
	if (float64(max)/float64(len(W))) > 0.9 {
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

	return DbWord.Update(func(tx *bolt.Tx)error{

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
			for _j:=j+2;_j<=len(lr);_j++ {
				key[string(lr[j:_j])]+=1
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
	titl := regTitle.FindAllString(self.Title,-1)
	newTi := make([]string,0,len(titl))
	for _,t := range titl{
		if len(t)>2 {
			newTi = append(newTi,t)
		}
	}
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
	return split_(append(newTi,regTitle.FindAllString(self.Content,-1)...))

}
