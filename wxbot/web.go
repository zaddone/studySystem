package main
import(
	//"github.com/zaddone/studySystem/wxmsg"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"encoding/binary"
	"net/http"
	"fmt"
	"time"
	//"bytes"
)
var (
	WXDB *bolt.DB
	Contact = []byte("ContactList")
	CantactMapTmp =  map[string]string{}
	//CantactMsgTmp =  map[string][]*Msg{}
)

func init(){
	var err error
	WXDB,err = bolt.Open("WXDB",0600,nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("init")
	Router := gin.Default()
	//Router.Static("/"+config.Conf.Static,"./"+config.Conf.Static)
	Router.LoadHTMLGlob("templates/*")
	Router.GET("/",func(c *gin.Context){
		c.HTML(http.StatusOK,"index.tmpl",nil)
	})
	Router.GET("/loginwx",func(c *gin.Context){
		c.Data(http.StatusOK,"image/jpeg",(<-CachePng).([]byte))
		//c.JSON(http.StatusOK,gin.H{"start":true})
	})
	Router.GET("/contactlist",func(c *gin.Context){
		li:=make([]string,0,100)
		err := WXDB.View(func(t *bolt.Tx)error{
			b := t.Bucket(Contact)
			if b == nil {
				return nil
			}
			return b.ForEach(func(k,v []byte)error{
				li = append(li,string(k))
				return nil
			})

		})
		if err != nil {
			c.JSON(http.StatusOK,gin.H{"state":false,"msg":err.Error()})
		}
		c.JSON(http.StatusOK,gin.H{"state":true,"list":li})
	})
	go Router.Run(":8001")

}

func GetMsgf(userid string,hand func(*Msg)error ) (err error) {
	name := CantactMapTmp[userid]
	if name == "" {
		return fmt.Errorf("%s = nil",userid)
	}
	return WXDB.View(func(t *bolt.Tx)error{
		b := t.Bucket([]byte(name))
		if b == nil {
			return fmt.Errorf("%s not found for db",name)
		}
		c := b.Cursor()
		c.Bucket()
		for k,v := c.Last();k!=nil;k,v = c.Prev(){
			msg := &Msg{}
			err = msg.LoadByte(v)
			if err != nil {
				return err
			}
			err=hand(msg)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func GetMsg(userid string,hand func(*Msg)error ) error {
	name := CantactMapTmp[userid]
	if name == "" {
		return fmt.Errorf("%s = nil",userid)
	}
	return WXDB.View(func(t *bolt.Tx)error{
		b := t.Bucket([]byte(name))
		if b == nil {
			return fmt.Errorf("%s not found for db",name)
		}
		return b.ForEach(func(k,v []byte)error{
			msg := &Msg{}
			err := msg.LoadByte(v)
			if err != nil {
				panic(err)
				return err
			}
			return hand(msg)
		})
	})
}
func UpdateMsg(msg *Msg ) error {
	userid := msg.FromUserName
	name := CantactMapTmp[userid]
	if name == "" {
		return fmt.Errorf("%s = nil",userid)
	}
	return WXDB.Update(func(t *bolt.Tx)error{
		b := t.Bucket([]byte(name))
		if b == nil {
			return fmt.Errorf("%s not found for db",name)
		}
		k := make([]byte,8)
		binary.BigEndian.PutUint64(k,uint64(time.Now().UnixNano()))
		return b.Put(k,msg.ToByte())
	})
}
func AddContact(name,userid string)error{

	CantactMapTmp[userid] = name
	return WXDB.Update(func(t *bolt.Tx)error{
		_,err := t.CreateBucketIfNotExists([]byte(name))
		return err
	})
}
