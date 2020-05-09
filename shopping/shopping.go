package shopping
import(
	"fmt"
	"github.com/boltdb/bolt"
	"encoding/gob"
	"encoding/json"
	"io"
	"bytes"
	//"strings"
	//"strconv"
	"time"
	"io/ioutil"
	"encoding/hex"
	"crypto/sha1"
	//"crypto/md5"
	"sync"
	"encoding/binary"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)
type NewShopping func(*ShoppingInfo,string) ShoppingInterface
var (

	OrderMsgHand func(...interface{})
	SiteList  = []byte("siteList")
	//orderDB = []byte("orderDB")
	order = []byte("order")
	orderTime = []byte("orderTime")
	orderUser = []byte("orderUser")
	UserInfo = []byte("UserInfo")
	iMsg = "请仔细核对商品，若有问题及时申请售后\n"
	//ShoppingMap = map[string]ShoppingInterface{}
	ShoppingMap = sync.Map{}// map[string]ShoppingInterface{}
	//siteDB string = "SiteDB"
	orderDB string = "orderDB"
	timeFormat = "2006-01-02 15:04:05"
	//siteDB  = flag.String("db","SiteDB","db")
	FuncMap = map[string]NewShopping{
		"jd":NewJd,
		"pinduoduo":NewPdd,
		"taobao":NewTaobao,
		"suning":NewSuning,
		"mogu":NewMogu,
		"vip":NewVip,
	}
	Rate = 0.9

)
func InterfaceToString(v interface{}) string {
	switch _v := v.(type){
	case string :
		return _v
	case float64:
		return fmt.Sprintf("%f",_v)
	case bool:
		return fmt.Sprintf("%b",_v)
	default:
		return ""
	}
	return ""
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
type Goods struct{
	Id string
	Img []string
	Name string
	Price float64
	Fprice string
	Tag string
	Coupon bool
	Show string
	Ext string
}
type User struct{
	//Mobile string
	//Name string
	UserId string
	//Users string
	Session string

}
func (self *User) Get() error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
	return DB.View(func(t *bolt.Tx)error{
		b := t.Bucket(UserInfo)
		if b == nil {
			return io.EOF
		}
		db := b.Get([]byte(self.UserId))
		if db == nil {
			return io.EOF
		}
		return json.Unmarshal(db,&self)
	})
	})

}
func (self *User) Update() error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
	return DB.Update(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(UserInfo)
		if err != nil {
			return err
		}
		db,err := json.Marshal(&self)
		if err != nil {
			return err
		}
		return b.Put([]byte(self.UserId),db)
	})
	})
}
func OrderDelDB(orderid string,DB *bolt.DB) error {
	o := []byte(orderid)
	return DB.Batch(func(t *bolt.Tx)error{
		b := t.Bucket(order)
		if b == nil {
			return io.EOF
		}
		return b.Delete(o)
	})
}
func OrderDel(orderid string)error {
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		return OrderDelDB(orderid,DB)
	})
}
func GetShoppingMap(py string) ShoppingInterface {
	v,_ := ShoppingMap.Load(py)
	if v == nil {
		return nil
	}
	return v.(ShoppingInterface)
}
//func Md5(data []byte) string {
//	m := md5.New()
//	_,err := m.Write(data)
//	if err != nil {
//		panic(err)
//	}
//	return hex.EncodeToString(md5.Sum([16]byte))
//}
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
	GoodsAppMini(...string)interface{}
	OrderDown(hand func(interface{}))error
	OrderDownSelf(hand func(interface{}))error
	//OrderDown(orderid string,db interface{})error

}
//F96BF1AC420D7D482DFFB7153173B5BE
type ShoppingInfo struct {
	Py string
	Name string
	Img string
	Uri string
	Client_id string
	Client_secret string
	Token string
	ReToken string
	TimeOut int64
	Update int64
	UpOrder int64
}
func OrderApplyUpdateDB(userid,orderid string,DB *bolt.DB)error {
	o := []byte(orderid)
	u := []byte(userid)
	//ti := time.Now().Unix()
	var k_ [8]byte
	binary.BigEndian.PutUint64(k_[:],uint64(time.Now().Unix()))
	return DB.Batch(func(t *bolt.Tx)error{
		b_,err := t.CreateBucketIfNotExists(orderUser)
		if err != nil {
			return err
		}
		ub,err :=b_.CreateBucketIfNotExists(u)
		if err != nil {
			return err
		}
		uub,err := ub.CreateBucketIfNotExists([]byte("order"))
		if err != nil {
			return err
		}
		uub_,err := ub.CreateBucketIfNotExists([]byte("time"))
		if err != nil {
			return err
		}

		vid := uub.Get(o)
		if vid != nil {
			uub_.Delete(vid)
			//return nil
		}
		err = uub.Put(o,k_[:])
		if err != nil {
			return err
		}
		return uub_.Put(k_[:],o)
	})
}
func OrderApplyUpdate(userid,orderid string)error {
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		return OrderApplyUpdateDB(userid,orderid,DB)
	})
}
func DownOrderAll(hand func(db interface{})) {
	var w sync.WaitGroup
	ShoppingMap.Range(func(k,v interface{})bool{
		w.Add(1)
		go func(s ShoppingInterface){
			err := s.OrderDownSelf(hand)
			if err != nil {
				fmt.Println(err)
			}
			w.Done()
		}(v.(ShoppingInterface))
		return true
	})
	w.Wait()
	return

}
func OrderApplyDB(userid,orderid string,DB *bolt.DB,hand func(interface{}))error{
	o := []byte(orderid)
	//u := []byte(userid)
	ti := time.Now().Unix()
	var db map[string]interface{}  = nil
	down := func(){
		DownOrderAll(func(_db interface{}){
			__db := _db.(map[string]interface{})
			oid_ := __db["order_id"].(string)
			OrderUpdateDB(oid_,_db,DB)
			if oid_ == orderid {
				db =  __db
				hand(db)
				if err := OrderApplyUpdateDB(userid,orderid,DB);err != nil {
					panic(err)
				}
			}
		})
	}
	return DB.Batch(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(order)
		if err != nil {
			return err
		}
		v := b.Get(o)
		if v != nil {
			err = json.Unmarshal(v,&db)
			if err != nil {
				return err
			}
			if db["userid"] != nil  {
				us := db["userid"].(string)
				if len(us)>0{
					if us != userid {
						return io.EOF
					}
					if db["order_id"] != nil {
						hand(db)
					}else{
						down()
					}
					return nil
				}
			}
			db["userid"] = userid
			hand(db)
			err=OrderApplyUpdateDB(userid,orderid,DB)
			if err != nil {
				panic(err)
			}
			return nil
		}
		down()
		if db != nil {
			return nil
		}
		db = map[string]interface{}{
			"userid":userid,
			"time":ti,
		}
		val,err := json.Marshal(db)
		if err != nil {
			return err
		}
		return b.Put(o,val)


	})
}
func OrderApply(userid,orderid string,hand func(interface{}))error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		return OrderApplyDB(userid,orderid,DB,hand)
	})
}
func OrderUpdateDB(orderid string,db interface{},DB *bolt.DB)error{
	o := []byte(orderid)
	return DB.Batch(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(order)
		if err != nil {
			return err
		}
		db_ := db.(map[string]interface{})
		val := b.Get(o)
		if val != nil {
			var valdb map[string]interface{}
			err := json.Unmarshal(val,&valdb)
			if err != nil {
				return err
			}
			db_["userid"] = valdb["userid"]
			if valdb["order_id"] == nil {
				OrderMsgHand(valdb["userid"],db_)
			}
		}
		str,err := json.Marshal(db_)
		if err != nil {
			return err
		}
		err=  b.Put(o,str)
		if err != nil {
			return err
		}
		if db_["userid"]==nil || len(db_["userid"].(string)) == 0{
			return nil
		}
		return OrderApplyUpdateDB(db_["userid"].(string),orderid,DB)
	})

}

func OrderUpdate(orderid string,db interface{})error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		return OrderUpdateDB(orderid,db,DB)
	})
}
func OrderListDB(orderid string,DB *bolt.DB,hand func(map[string]interface{})error)error{
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
			//fmt.Println(string(k))
			var db  map[string]interface{}
			err := json.Unmarshal(v,&db)
			if err != nil {
				return err
			}
			if db["order_id"] == nil {
				continue
				//db["order_id"] = string(k)
			}
			err = hand(db)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func OrderList(orderid string,hand func(map[string]interface{})error)error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		return OrderListDB(orderid,DB,hand)
	})
}
func OrderUserDel(numid,userid string) error {
	return nil
}
func OrderListWithUser(numid,userid string,hand func(interface{})error)error{
	getDB := func(b *bolt.Bucket,k,id []byte) error {
		v := b.Get(id)
		if v == nil {
			return fmt.Errorf("is nil")
		}
		var db map[string]interface{}
		err := json.Unmarshal(v,&db)
		if err != nil {
			return err
		}
		if db["goodsid"] == nil {
			return nil
		}
		db["numid"] = binary.BigEndian.Uint64(k)
		return hand(db)
	}
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		var del [][][]byte
		err := DB.View(func(t *bolt.Tx)error{
			b_ := t.Bucket(order)
			if b_ == nil {
				return io.EOF
			}
			b := t.Bucket(orderUser)
			if b == nil {
				return io.EOF
			}
			b = b.Bucket([]byte(userid))
			if b == nil {
				return io.EOF
			}
			b = b.Bucket([]byte("time"))
			c := b.Cursor()
			var orid,k,v []byte
			if len(numid) == 0 {
				k,v = c.Last()
				err := getDB(b_,k,v)
				if err != nil {
					if err == io.EOF{
						return err
					}
					del = append(del,[][]byte{k,v})
				}
			}else{
				orid = []byte(numid)
				k,v = c.Seek(orid)
				if k == nil {
					return nil
				}
			}
			for k,v = c.Prev();k!=nil;k,v=c.Prev(){
				err := getDB(b_,k,v)
				if err != nil {
					if err == io.EOF {
						return err
					}
					//del = append(del,k)
					del = append(del,[][]byte{k,v})
				}
			}
			return nil
		})
		if err != nil {
			if err != io.EOF{
				fmt.Println(err)
			}
			return err
		}
		if len(del) == 0 {
			return nil
		}
		return DB.Batch(func(t *bolt.Tx)error{
			b := t.Bucket(orderUser)
			if b == nil {
				return io.EOF
			}
			b = b.Bucket([]byte(userid))
			if b == nil {
				return io.EOF
			}
			b_o := b.Bucket([]byte("order"))
			b_t := b.Bucket([]byte("time"))
			for _,_id := range del{
				b_o.Delete(_id[1])
				b_t.Delete(_id[0])
			}
			return nil
		})

	})

}
func OrderGet(orderid,userid string,hand func(interface{}))error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		return  DB.View(func(t *bolt.Tx)error{
			b := t.Bucket(orderUser)
			//b := t.Bucket(order)
			if b == nil {
				return io.EOF
			}
			b = b.Bucket([]byte(userid))
			if b == nil {
				return io.EOF
			}
			v := b.Get([]byte(orderid))
			if v == nil {
				return io.EOF
			}
			b = t.Bucket(order)
			if b == nil {
				return io.EOF
			}
			v = b.Get([]byte(orderid))
			if v == nil {
				return io.EOF
			}
			var db map[string]interface{}
			err := json.Unmarshal(v,&db)
			if err != nil {
				return err
			}
			if db["goodsid"]== nil {
				return io.EOF
			}
			hand(db)
			return nil
		})
	})
	//return orderGet(orderid,userid,hand)
}

func orderGet(orderid,userid string,hand func(interface{}))error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		return  DB.View(func(t *bolt.Tx)error{
			b := t.Bucket(order)
			if b == nil {
				return io.EOF
			}
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
func OrderUpdateTime(orderid,py string,ti []byte)error{
	return OpenSiteDB(orderDB,func(DB *bolt.DB)error{
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
	//siteDB = dbname
	err := ReadShoppingList(dbname,func(sh *ShoppingInfo)error{
		//fmt.Println(sh)
		hand := FuncMap[sh.Py]
		if hand != nil {
			ShoppingMap.Store(sh.Py,hand(sh,dbname))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
