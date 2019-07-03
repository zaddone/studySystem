package chrome
import(
	"fmt"
	"time"
	"regexp"
	"encoding/binary"
	"encoding/json"
	//"sync"
	"github.com/boltdb/bolt"
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
	Id int64
	Title string
	Content string
	Par int64
	Children []int64
}
func NewPage(title,content string) *Page {
	return &Page{
		Id:time.Now().UnixNano(),
		Title:title,
		Content:content,
	}
}
func (self *Page) SaveDB() error {

	k := make([]byte,8)
	binary.BigEndian.PutUint64(k,uint64(self.Id))
	v,err := json.Marshal(self)
	if err != nil {
		return
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
		b.Put(k,v)
	})

}
func (self *Page) StudyWord() error {

	db,err := bolt.Open(WordDB,0600,nil)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	b,err := tx.CreateBucketIfNotExists(WordBucket)
	if err != nil {
		return err
	}
	c := b.Cursor()
	self.splitWord(func(w string,c int){
	//	for i:= len(w);i>1;i--{
	//		//fmt.Println(string(w[i-2:i]))
		k,v := c.Seek([]byte(string(w[i-2:i])))
	//	}
	})

	if err := tx.Commit(); err != nil {
		return err
		log.Fatal(err)
	}

	return nil
}

func split(li []string)(work map[string]int){
	work = map[string]int{}
	le := len(li)
	for i:=0;i<le;i++ {
		s_bak := []rune(li[i])
		I := i+1
		_li := li[I:]
		if len(_li) > 0 {
			for j:= len(s_bak)-1;j>=0;j--{
				sk:= s_bak[j]
				for _i,_s := range _li{
					t := strings.IndexRune(_s,sk)
					if t<=0 {
						continue
					}
					li[I+_i] = _s[:t]+string(sk)
					li = append(li,_s[t:])
					le = len(li)
					work[string(s_bak[j:])] +=1
					s_bak = append(s_bak[:j],sk)
				}
			}
		}
		work[string(s_bak)] +=1
	}
	return
}

func (self *Page) splitWord(h func(k string ,v int)) {
	m := map[string]int{}
	li := regTitle.FindAllString(self.Title,-1)
	li = append(li,regTitle.FindAllString(self.Content,-1))
	m := split(li)
	for k,v := split(li) {
		h(k,v)
	}


}
