package shopping
import(
	"fmt"
	"github.com/boltdb/bolt"
	"encoding/gob"
	"encoding/json"
	"io"
	"bytes"
	//"flag"
	"encoding/hex"
	"crypto/sha1"
	"sync"
	//"encoding/binary"
)
type NewShopping func(*ShoppingInfo) ShoppingInterface
var (
	SiteList  = []byte("siteList")
	order = []byte("order")
	orderTime = []byte("orderTime")
	orderUser = []byte("orderUser")
	UserInfo = []byte("UserInfo")
	iMsg = "请仔细核对商品，若有问题及时申请售后\n"
	//ShoppingMap = map[string]ShoppingInterface{}
	ShoppingMap = sync.Map{}// map[string]ShoppingInterface{}
	siteDB string
	timeFormat = "2006-01-02 15:04:05"
	//siteDB  = flag.String("db","SiteDB","db")
	FuncMap = map[string]NewShopping{
		"jd":NewJd,
		"pinduoduo":NewPdd,
		"taobao":NewTaobao,
	}

)
func Sha1(data []byte) string {
	sha1 := sha1.New()
	sha1.Write(data)
	return hex.EncodeToString(sha1.Sum([]byte(nil)))
}
type ShoppingInterface interface{
	GetInfo()*ShoppingInfo
	SearchGoods(...string)interface{}
	GoodsUrl(...string)interface{}
	GoodsDetail(...string)interface{}
	OrderSearch(...string)interface{}
	OutUrl(interface{}) string
	OrderMsg(interface{}) string
	ProductSearch(...string)[]interface{}
	OrderDown(hand func(interface{}))error
	//OrderUpdate(orderid string,db interface{})error

}

type ShoppingInfo struct {
	Py string
	Name string
	Img string
	Uri string
	Client_id string
	Client_secret string
	Token string
	Update int64
}
func OrderApply(userid,orderid string)error{
	o := []byte(orderid)
	u := []byte(userid)
	return openSiteDB(siteDB,func(DB *bolt.DB)error{
	return DB.Update(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(order)
		if err != nil {
			return err
		}
		v := b.Get(o)
		var db map[string]interface{}
		if v != nil {
			err = json.Unmarshal(v,&db)
			if err != nil {
				return err
			}
			if db["userid"] != nil {
				if db["userid"].(string) == userid {
					return nil
				}
				return io.EOF
			}
		}else{
			db = map[string]interface{}{
				"userid":userid,
				//"order_id":orderid,
			}
		}
		val,err := json.Marshal(db)
		if err != nil {
			return err
		}
		err = b.Put(o,val)
		if err != nil {
			return err
		}
		b_,err := t.CreateBucketIfNotExists(orderUser)
		if err != nil {
			return err
		}
		ub,err := b_.CreateBucketIfNotExists(u)
		if err != nil {
			return err
		}
		err = ub.Put(o,[]byte{0})
		if err != nil {
			return err
		}
		return nil

	})
	})
}

func OrderUpdate(orderid string,db interface{})error{

	return openSiteDB(siteDB,func(DB *bolt.DB)error{
	return DB.Batch(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(order)
		if err != nil {
			return err
		}
		db_ := db.(map[string]interface{})
		val := b.Get([]byte(orderid))
		if val != nil {
			var valdb map[string]interface{}
			err := json.Unmarshal(val,&valdb)
			if err != nil {
				return err
			}
			db_["userid"] = valdb["userid"]
		}
		str,err := json.Marshal(db_)
		if err != nil {
			return err
		}
		return b.Put([]byte(orderid),str)
	})
	})
}
func OrderList(orderid string,hand func(map[string]interface{})error)error{
	return openSiteDB(siteDB,func(DB *bolt.DB)error{
		return  DB.View(func(t *bolt.Tx)error{
			b := t.Bucket(order)
			if b == nil {
				return io.EOF
			}
			//fmt.Println("read",order)
			c:=b.Cursor()
			//k_,_ := c.First()
			//fmt.Println(string(k_))
			//var k,v []byte
			var k,v []byte
			if len(orderid) == 0 {
				k,v = c.First()
			}else{
				k,v = c.Seek([]byte(orderid))
			}
			for ;k!= nil;k,v=c.Next(){
				fmt.Println(string(k))
				var db  map[string]interface{}
				err := json.Unmarshal(v,&db)
				if err != nil {
					return err
				}
				err = hand(db)
				if err != nil {
					return err
				}
			}
			return nil
		})
	})

}

func OrderGet(orderid,userid string,hand func(interface{}))error{
	return orderGet(orderid,userid,hand)
}
func orderGet(orderid,userid string,hand func(interface{}))error{
	return openSiteDB(siteDB,func(DB *bolt.DB)error{
		return  DB.View(func(t *bolt.Tx)error{
			b := t.Bucket(order)
			if b == nil {
				return io.EOF
			}
			//b = b.Bucket([]byte(self.Py))
			v := b.Get([]byte(orderid))
			if v == nil {
				return io.EOF
			}
			var db map[string]interface{}
			err := json.Unmarshal(v,&db)
			if err != nil {
				return err
			}
			if userid != ""{
			//fmt.Println(db)
			uid := db["userid"]
			if uid != nil &&  uid.(string) != userid {
				return io.EOF
			}
			}
			hand(db)
			return nil
		})
	})
}
func (self *ShoppingInfo)Load(db *bolt.DB) error {
	if self.Py == "" {
		return fmt.Errorf("name = nil")
	}
	return db.View(func(t *bolt.Tx) error{
		b := t.Bucket(SiteList)
		if b == nil {
			return fmt.Errorf("b == nil")
		}
		v := b.Get([]byte(self.Py))
		if v == nil{
			return nil
		}
		return self.loadByte(v)

	})

}

func (self *ShoppingInfo) loadByte(b []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(b)).Decode(self)
}
func (self *ShoppingInfo) toByte() []byte {
	var db bytes.Buffer
	err := gob.NewEncoder(&db).Encode(self)
	if err != nil {
		panic(err)
	}
	return db.Bytes()
}
func (self *ShoppingInfo) SaveToDB(db *bolt.DB) error {
	return db.Update(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(SiteList)
		if err != nil {
			return err
		}
		return b.Put([]byte(self.Py),self.toByte())
	})
}

func OrderWithTime(ti []byte,hand func(string,interface{}))error{
	return OpenSiteDB(siteDB,func(DB *bolt.DB)error{
		return DB.View(func(t *bolt.Tx)error{
			b:= t.Bucket(orderTime)
			if b == nil {
				return nil
			}
			b_:= b.Bucket(ti)
			if b_ == nil {
				return nil
			}
			return b_.ForEach(func(k,v []byte)error{
				py := string(v)
				//info := ShoppingInfo{Py:string(v)}
				return orderGet(string(k),"",func(db interface{}){
					hand(py,db)
				})
			})
		})
	})
}
func OrderUpdateTime(orderid,py string,ti []byte)error{
	return OpenSiteDB(siteDB,func(DB *bolt.DB)error{
		return DB.Batch(func(t *bolt.Tx)error{
			b,err := t.CreateBucketIfNotExists(orderTime)
			if err != nil {
				return err
			}
			b_,err := b.CreateBucketIfNotExists(ti)
			if err != nil {
				return err
			}
			//b__,err := b_.CreateBucketIfNotExists([]byte(self.Py))
			//if err != nil {
			//	return err
			//}
			return b_.Put([]byte(orderid),[]byte(py))
		})

	})

}

func OpenSiteDB(dbname string,h func(*bolt.DB)error)error{
	return openSiteDB(dbname,h)
}
func openSiteDB(dbname string,h func(*bolt.DB)error)error{
	db ,err := bolt.Open(dbname,0600,nil)
	if err != nil {
		return err
	}
	//fmt.Println("open",dbname)
	defer func(){
		err := db.Close()
		if err != nil {
			panic(err)
		}
		//fmt.Println("close",dbname)
	}()
	return h(db)
}
func ReadShoppingList(dbname string,h func(*ShoppingInfo)error)error{
	return openSiteDB(dbname,func(db *bolt.DB)error{
		return db.View(func(t *bolt.Tx) error{
			b := t.Bucket(SiteList)
			if b == nil {
				return fmt.Errorf("b == nil")
			}
			return b.ForEach(func(k,v []byte)error{
				sh := &ShoppingInfo{}
				er := sh.loadByte(v)
				if er != nil {
					return er
				}
				return h(sh)
			})
		})
	})

}
func InitShoppingMap(dbname string){
	siteDB = dbname
	err := ReadShoppingList(dbname,func(sh *ShoppingInfo)error{
		hand := FuncMap[sh.Py]
		if hand != nil {
			ShoppingMap.Store(sh.Py,hand(sh))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
