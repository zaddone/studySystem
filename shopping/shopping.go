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

func OrderDel(orderid string)error {
	o := []byte(orderid)
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
	return DB.Batch(func(t *bolt.Tx)error{
		b := t.Bucket(order)
		if b == nil {
			return io.EOF
		}
		db := b.Get(o)
		var or  Order
		err := json.Unmarshal(db,&or)
		if err != nil {
			return err
		}
		err = b.Delete(o)
		if err != nil {
			return err
		}

		if len(or.UserId)==0{
			return nil
		}
		b = t.Bucket(orderUser)
		if b == nil {
			return nil
		}
		b = b.Bucket([]byte(or.UserId))
		if b == nil {
			return nil
		}
		//fmt.Println(or.UserId,or.GoodsName)
		return b.Delete(o)
		//return b.Delete(o)
	})
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
func OrderApplyUpdate(userid,orderid string)error {
	o := []byte(orderid)
	u := []byte(userid)
	ti := time.Now().Unix()
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
	return DB.Update(func(t *bolt.Tx)error{
		b_,err := t.CreateBucketIfNotExists(orderUser)
		if err != nil {
			return err
		}
		ub,err := b_.CreateBucketIfNotExists(u)
		if err != nil {
			return err
		}
		var k_ [8]byte
		binary.BigEndian.PutUint64(k_[:],uint64(ti))
		return ub.Put(k_[:],o)
	})
	})
}
func OrderApply(userid,orderid string,hand func(interface{}))error{
	o := []byte(orderid)
	u := []byte(userid)
	ti := time.Now().Unix()
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
	return DB.Update(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(order)
		if err != nil {
			return err
		}
		v := b.Get(o)
		var db map[string]interface{}  = nil
		if v != nil {
			err = json.Unmarshal(v,&db)
			if err != nil {
				return err
			}
			if db["userid"] != nil {
				us := db["userid"].(string)
				if len(us)>0 {
					if us != userid {
						//hand(db)
						//return nil
						return io.EOF
					}
					//return io.EOF
				}else{

					db["userid"] = userid
				}
			}
			ti = int64(db["time"].(float64))
		}else{
			var w sync.WaitGroup
			dbchan:=make(chan interface{},1000)
			go func(){
				for dbc := range dbchan {
					__db := dbc.(map[string]interface{})
					oid_ := __db["order_id"].(string)
					OrderUpdate(oid_,dbc)
					if db == nil && orderid == oid_{
						db = __db
					}
				}
			}()
			ShoppingMap.Range(func(k,v interface{})bool{
				w.Add(1)
				go func(s ShoppingInterface){
				err := s.OrderDownSelf(func(_db interface{}){
					dbchan<-_db
				})
				if err != nil {
					fmt.Println(err)
				}
				w.Done()
				}(v.(ShoppingInterface))
				return true
			})
			w.Wait()
			close(dbchan)
			if db == nil {
				db = map[string]interface{}{
					"userid":userid,
					"time":ti,
					//"order_id":orderid,
				}
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
		var k_ [8]byte
		binary.BigEndian.PutUint64(k_[:],uint64(ti))
		err = ub.Put(k_[:],o)
		//err = ub.Put([]byte(fmt.Sprintf("%d",ti)),o)
		if err != nil {
			return err
		}
		hand(db)
		return nil

	})
	})
}

func OrderUpdate(orderid string,db interface{})error{

	o := []byte(orderid)
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
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
		b_,err := t.CreateBucketIfNotExists(orderUser)
		if err != nil {
			return err
		}
		ub,err := b_.CreateBucketIfNotExists([]byte(db_["userid"].(string)))
		if err != nil {
			return err
		}
		return ub.Put(o,[]byte(fmt.Sprintf("%d",time.Now().Unix())))
	})
	})
}
func OrderList(orderid string,hand func(map[string]interface{})error)error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
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
					db["order_id"] = string(k)
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
func OrderUserDel(numid,userid string) error {
	return nil
}
func OrderListWithUser(numid,userid string,hand func(interface{})error)error{
	getDB := func(b *bolt.Bucket,k,id []byte) error {
		v := b.Get(id)
		t_ := int64(binary.BigEndian.Uint64(k))
		//fmt.Println(t_,string(id),string(v))
		if v == nil {
			//v = b.Get(k)
			//if v == nil {
			//t_,err := strconv.Atoi(string(k))
			//if err != nil {
			//	return err
			//}
			tk := (time.Now().Unix() - t_)
			if (tk<0) || (tk>604800) {
				return fmt.Errorf("time out")
			}
			return nil
			//}
		}
		var db map[string]interface{}
		err := json.Unmarshal(v,&db)
		if err != nil {
			return err
		}
		if db["goodsid"] == nil {
			return nil
		}
		db["numid"] = t_
		return hand(db)
	}
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		var del [][]byte
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
			c := b.Cursor()
			var orid,k,v []byte
			if len(numid) == 0 {
				k,v = c.Last()
				//fmt.Println(string(v))
				err := getDB(b_,k,v)
				if err != nil {
					if err == io.EOF {
						return err
					}
					//fmt.Println(err)
					del = append(del,k)
					//fmt.Println(err,k)
					//err = b.Delete(k)
					//fmt.Println(err)
					//return err
				}
				//orid = []byte{0}
			}else{
				orid = []byte(numid)
				k,v = c.Seek(orid)
				if k == nil {
					return nil
				}
			}
			//orid := []byte(orderid)
			//k,v := c.Seek(orid)
			//var err error
			//if !bytes.Equal(orid,k){
			//	err = getDB(b_,k,v)
			//	if err != nil {
			//		return err
			//	}
			//}
			for k,v = c.Prev();k!=nil;k,v=c.Prev(){
				err := getDB(b_,k,v)
				if err != nil {
					if err == io.EOF {
						return err
					}

					del = append(del,k)
					//fmt.Println(err)
					//err = b.Delete(k)
					//fmt.Println(err)
				}
			}
			return nil
		})
		if err != nil {
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
			for _,_id := range del{
				b.Delete(_id)
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

func OrderWithTime(ti []byte,hand func(string,interface{}))error{
	return OpenSiteDB(orderDB,func(DB *bolt.DB)error{
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
