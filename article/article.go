package main
import(
	"fmt"
	"bytes"
	"html/template"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"time"
	"net/http"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"io"
	"flag"
)
var (
	articleDB = "article.db"
	timeFormat_ = "20060102"
	artDB *bolt.DB
	port = flag.String("p",":8081","port")
)
func main(){
	select{}
}
func init(){
	var err error
	artDB,err = bolt.Open(articleDB,0600,nil)
	if err != nil {
		panic(err)
	}
	Router := gin.Default()
	Router.LoadHTMLGlob("./templates/*")
	Router.GET("/",func(c *gin.Context){
		template.ParseFiles()
		c.HTML(http.StatusOK,"article.tmpl",nil)
	})
	Router.GET("/view",func(c *gin.Context){
		begin,err := strconv.Atoi(c.DefaultQuery("begin","0"))
		if err != nil {
			return
		}
		count,err := strconv.Atoi(c.DefaultQuery("count","30"))
		if err != nil {
			return
		}
		ars:=make([]*Article,0,count)
		err = forEachArticle(begin,func(ar *Article)error{
			ars = append(ars,ar)
			count--
			if count<=0{
				return io.EOF
			}
			return nil
		})
		if err != nil {
			if err != io.EOF {
				return
			}
		}
		c.JSON(http.StatusOK,gin.H{"dbs":ars})
	})
	Router.GET("/view/:id",func(c *gin.Context){
		id,err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return
		}
		ar := &Article{
			Date:int64(id),
		}
		err = ar.Load()
		if err != nil {
			return
		}
		c.JSON(http.StatusOK,gin.H{"db":ar})
	})
	Router.GET("/del/:id",func(c *gin.Context){
		id,err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return
		}
		ar := &Article{
			Date:int64(id),
		}
		err = ar.Del()
		if err != nil {
			return
		}
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
	})
	Router.POST("/update",func(c *gin.Context){
		id,err := strconv.Atoi(c.DefaultQuery("id","0"))
		if err != nil {
			fmt.Println(err)
			return
		}
		ar := &Article{Date:int64(id)}
		err = json.NewDecoder(c.Request.Body).Decode(&ar)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ar.Save()
		if err != nil {
			return
		}
		c.JSON(http.StatusOK,gin.H{"msg":"success"})
	})
	go Router.Run(*port)
}


type Article struct{
	Title string
	Date int64
	Content string
	Source string
	Author string
	key []byte
	dir []byte
}
func forEachArticle(begin int,hand func(*Article)error)error {
	beginAr := &Article{Date:int64(begin)}
	if beginAr.Date ==0 {
		beginAr.Date = time.Now().UnixNano()
	}
	var err error
	var ar Article
	return artDB.View(func(t *bolt.Tx)error{
		c := t.Cursor()
		k,_ := c.Seek(beginAr.getDir())
		//k,v := c.Last()
		if k == nil {
			return io.EOF
		}
		for k,_ := c.Last();k!=nil;k,_ = c.Prev(){
			c_ := t.Bucket(k).Cursor()
			_k,_v := c_.Seek(beginAr.getKey())
			if _k == nil {
				return io.EOF
			}
			if bytes.Equal(_k,beginAr.getKey()){
				_k,_v = c_.Prev()
			}
			for ;_k != nil;_k,_v = c_.Prev(){
				//ar = &Article{}
				err = json.Unmarshal(_v,&ar)
				if err != nil {
					return err
				}
				err = hand(&ar)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
func (self *Article)getDir() []byte {
	if self.dir != nil {
		return self.dir
	}
	if self.Date == 0 {
		return nil
	}
	self.dir=make([]byte,4)
	y,m,d := time.Unix(self.Date/int64(time.Millisecond),0).Date()
	binary.BigEndian.PutUint32(self.dir,uint32(time.Date(y,m,d,0,0,0,0,time.Local).Unix()))
	return self.dir
}
func (self *Article)getKey() []byte {
	if self.key != nil {
		return self.key
	}
	if self.Date == 0 {
		return nil
	}
	self.key=make([]byte,8)
	binary.BigEndian.PutUint64(self.key,uint64(self.Date))
	return self.key
}

func(self *Article)Del()error{
	return artDB.Batch(func(t *bolt.Tx)error{
		b := t.Bucket(self.getDir())
		if b == nil {
			return nil
		}
		return b.Delete(self.getKey())
	})
}
func(self *Article)Load()error{
	if self.Date == 0 {
		return fmt.Errorf("date = 0")
	}
	//key := self.getKey()
	return artDB.View(func(t *bolt.Tx)error{
		b := t.Bucket(self.getDir())
		if b == nil {
			return fmt.Errorf("find not")
		}
		body := b.Get(self.getKey())
		if body == nil {
			return fmt.Errorf("find not")
		}
		return self.loadByte(body)
	})

}
func(self *Article) Save()error{
	if self.Title == "" || self.Content == "" {
		return fmt.Errorf("title or content is nil")
	}
	if self.Date == 0 {
		self.Date = time.Now().UnixNano()
	}
	if self.Author == "" {
		self.Author = "Admin"
	}
	return artDB.Batch(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(self.getDir())
		if err != nil {
			return err
		}
		return b.Put(self.getKey(),self.toByte())
	})

}

func (self *Article) loadByte(db []byte)error {
	return json.Unmarshal(db,&self)
}
func (self *Article) toByte() []byte {
	db,err := json.Marshal(self)
	if err != nil {
		panic(err)
	}
	return db
}

