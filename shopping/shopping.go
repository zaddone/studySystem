package shopping
import(
	"fmt"
	"github.com/boltdb/bolt"
	"encoding/gob"
	"encoding/json"
	"io"
	"bytes"
	//"strings"
	"strconv"
	//"encoding/gob"
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

	OrderMsgHand func(...interface{}) = nil
	SiteList  = []byte("siteList")
	//orderDB = []byte("orderDB")
	order = []byte("order")
	UserInfo = []byte("UserInfo")
	iMsg = "请仔细核对商品，若有问题及时申请售后\n"
	//ShoppingMap = map[string]ShoppingInterface{}
	ShoppingMap = sync.Map{}// map[string]ShoppingInterface{}
	//siteDB string = "SiteDB"

	orderTime = []byte("orderTime")
	orderUser = []byte("orderUser")
	orderDB string = "orderDB"
	//userDB string = "userDB"

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
	DownOrder = make(chan bool,100)
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
	return DB.Batch(func(t *bolt.Tx)error{
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
	//OrderSearch(...string)interface{}
	OutUrl(interface{}) string
	OrderMsg(interface{}) string
	ProductSearch(...string)[]interface{}
	GoodsAppMini(...string)interface{}
	OrderDown(hand func(interface{}))error
	OrderDownSelf(hand func(interface{}))error
	Test()interface{}
}
//OrderDown(orderid string,db interface{})error
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
type OrderApplyStruct struct{
	Fee float64
	//FeeP float64
	Date int64
	//Date_ int64
}
func (self *OrderApplyStruct) encode() []byte {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(self)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (self *OrderApplyStruct) decode(data []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(self)
}
func (self *OrderApplyStruct) getkey() []byte {
	if self.Date == 0 {
		self.Date = time.Now().Unix()
	}
	var k_ [8]byte
	binary.BigEndian.PutUint64(k_[:],uint64(self.Date))
	return k_[:]
}
func OrderApplyUpdateDB(userid,orderid string,ordermap interface{},t *bolt.Tx)error {
	o := []byte(orderid)
	u := []byte(userid)
	var order_map map[string]interface{}
	if ordermap == nil {
		_b := t.Bucket(order)
		if _b == nil {
			return fmt.Errorf("order bucket is nil")
		}
		order_val := _b.Get(o)
		if order_val == nil {
			return fmt.Errorf("order is nil")
		}
		err := json.Unmarshal(order_val,&order_map)
		if err != nil {
			return err
		}
	}else{
		order_map = ordermap.(map[string]interface{})
	}

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
	//uub_,err := ub.CreateBucketIfNotExists([]byte("time"))
	//if err != nil {
	//	return err
	//}
	var Sum float64
	sumFee := ub.Get([]byte("sum"))
	if sumFee != nil {
		Sum,err = strconv.ParseFloat(string(sumFee),64)
		if err != nil {
			return err
			//panic(err)
		}
	}
	oa := &OrderApplyStruct{}
	vid := uub.Get(o)
	if vid != nil {
		oa.decode(vid)
		Sum -= oa.Fee
		//uub_.Delete(oa.getkey())
	}

	//fmt.Println("order map")
	//fmt.Println(order_map)
	if order_map["payTime"] == nil{
		oa.Date = int64(order_map["time"].(float64))
		oa.Fee = 0
	}else{
		oa.Date =int64( order_map["payTime"].(float64))
		oa.Fee = order_map["fee"].(float64)
	}
	//fmt.Println("order map 1")
	Sum += oa.Fee
	err = ub.Put([]byte("sum"),[]byte(fmt.Sprintf("%.2f",Sum)))
	if err != nil {
		return err
	}
	return uub.Put(o,oa.encode())
	//if err != nil {
	//	return err
	//}
	//return uub_.Put(oa.getkey(),o)
}

func OrderApplyUpdate(userid,orderid string)error {
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		t,err := DB.Begin(true)
		if err != nil {
			return err
		}
		err = OrderApplyUpdateDB(userid,orderid,nil,t)
		if err != nil {
			return err
		}
		return t.Commit()
	})
}
func GetUserSum(userid string) interface{} {
	var val string
	err := openSiteDB(orderDB,func(DB *bolt.DB)error{
		t,err := DB.Begin(false)
		if err != nil {
			return err
		}
		b_ := t.Bucket(orderUser)
		if b_ == nil {
			return io.EOF
		}
		ub :=b_.Bucket([]byte(userid))
		if ub == nil {
			return io.EOF
		}
		v := ub.Get([]byte("sum"))
		if v == nil {
			return fmt.Errorf("sum is nil")
		}
		val = string(v)
		return nil
	})
	if err != nil {
		return err
	}
	return val
}

func DownOrderAll(hand func(db interface{})) {
	var w sync.WaitGroup
	ShoppingMap.Range(func(k,v interface{})bool{
		//fmt.Println(k)
		w.Add(1)
		go func(s ShoppingInterface){
			err := s.OrderDownSelf(hand)
			if err != nil {
				fmt.Println(err)
			}

			w.Done()
			fmt.Println(k)
		}(v.(ShoppingInterface))
		return true
	})
	w.Wait()
	return
}

func OrderApplyDB(userid,orderid string,t *bolt.Tx,hand func(interface{}))error{
	b,err := t.CreateBucketIfNotExists(order)
	if err != nil {
		return err
	}
	o := []byte(orderid)
	v := b.Get(o)
	var db map[string]interface{}
	if v == nil {
		DownOrderAll(func(_db interface{}){
			fmt.Println(_db)
			__db := _db.(map[string]interface{})
			oid_ := __db["order_id"].(string)
			if oid_ == orderid {
				__db["userid"] = userid
				hand(__db)
			}
			err = OrderUpdateDB(oid_,__db,t)
			if err != nil {
				panic(err)
			}
		})
		return nil
	}
	if err = json.Unmarshal(v,&db); err != nil {
		return err
	}
	if db["userid"] != nil {
		us := db["userid"].(string)
		if len(us)>0 {
			if us != userid {
				return nil
			}else{
				hand(db)
			}
			return nil
		}
	}
	db["userid"] = userid
	hand(db)
	v,_ = json.Marshal(db)
	if err = b.Put(o,v); err != nil {
		return err
	}
	return OrderApplyUpdateDB(userid,orderid,db,t)

}


func OrderApply(userid,orderid string,hand func(interface{}))error{
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		t,err := DB.Begin(true)
		if err != nil {
			return err
		}
		err = OrderApplyDB(userid,orderid,t,hand)
		if err != nil {
			return err
		}
		return t.Commit()
	})
}

func OrderUpdateDB(orderid string,db interface{},t *bolt.Tx)error{

	o := []byte(orderid)
	//fmt.Println("r orderupdateDB")
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
		if valdb["userid"] != nil &&
		len(valdb["userid"].(string))>0 {
			db_["userid"] = valdb["userid"]
			//OrderMsgHand(db_)
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
	if db_["userid"]!=nil && len(db_["userid"].(string)) > 0{
		return OrderApplyUpdateDB(db_["userid"].(string),orderid,db_,t)
	}
	return nil

}

func OrderUpdate(orderid string,db interface{})error{
	return openSiteDB(orderDB,func(DB *bolt.DB) error {
		t,err := DB.Begin(true)
		if err != nil {
			return err
		}
		err = OrderUpdateDB(orderid,db,t)
		if err != nil {
			return err
		}
		return t.Commit()
	})
}

func OrderListDB(orderid string,DB *bolt.DB,hand func(map[string]interface{})error)error{
	return  DB.View(func(t *bolt.Tx)error{
		b := t.Bucket(order)
		if b == nil {
			return io.EOF
		}
		//defer fmt.Println("end read list db")
		//fmt.Println("read",order)
		c := b.Cursor()
		//k_,_ := c.First()
		//fmt.Println(string(k_))
		//var k,v []byte
		var k,v []byte
		if len(orderid) == 0 {
			k,v = c.First()
		}else{
			k,v = c.Seek([]byte(orderid))
			if bytes.Equal(k,[]byte(orderid)){
				k,v = c.Next()
			}
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
				//fmt.Println(err)
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
	getDB := func(b *bolt.Bucket,k []byte) error {
		//fmt.Println(string(id))
		v := b.Get(k)
		//t:= int64(binary.BigEndian.Uint64(k))
		if v == nil {
			//if time.Now().Unix()-t > 604800{
			//	return fmt.Errorf("is nil")
			//}else{
			return io.EOF
			//}
		}
		var db map[string]interface{}
		err := json.Unmarshal(v,&db)
		if err != nil {
			return err
		}
		//if db["goodsid"] == nil {
		//	return nil
		//}
		db["numid"] = string(k)
		return hand(db)
	}
	return openSiteDB(orderDB,func(DB *bolt.DB)error{
		//var del [][][]byte
		return DB.View(func(t *bolt.Tx)error{
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

			b = b.Bucket([]byte("order"))
			c := b.Cursor()
			var orid,k []byte
			if len(numid) == 0 {
				k,_ = c.Last()
				err := getDB(b_,k)
				fmt.Println(err,k)
				if err != nil {
					//if err == io.EOF{
					return err
					//}
					//del = append(del,[][]byte{k,v})
				}
			}else{
				orid = []byte(numid)
				k,_ = c.Seek(orid)
				if k == nil {
					return nil
				}
			}
			for k,_ = c.Prev();k!=nil;k,_=c.Prev(){
				//fmt.Println(k)
				err := getDB(b_,k)
				if err != nil {
					//if err == io.EOF {
					return err
					//}
					//del = append(del,k)
					//del = append(del,[][]byte{k,v})
				}
			}
			//fmt.Println("end")
			return nil
		})
		//if err != nil {
		//	if err != io.EOF{
		//		fmt.Println(err)
		//	}
		//	return err
		//}
		//if len(del) == 0 {
		//	return nil
		//}
		//return DB.Update(func(t *bolt.Tx)error{
		//	b := t.Bucket(orderUser)
		//	if b == nil {
		//		return io.EOF
		//	}
		//	b = b.Bucket([]byte(userid))
		//	if b == nil {
		//		return io.EOF
		//	}
		//	b_o := b.Bucket([]byte("order"))
		//	b_t := b.Bucket([]byte("time"))
		//	for _,_id := range del{
		//		b_o.Delete(_id[1])
		//		b_t.Delete(_id[0])
		//	}
		//	return nil
		//})

	})

}
//func OrderGet(orderid,userid string,hand func(interface{}))error{
//	return openSiteDB(orderDB,func(DB *bolt.DB)error{
//		return  DB.View(func(t *bolt.Tx)error{
//			b := t.Bucket(orderUser)
//			//b := t.Bucket(order)
//			if b == nil {
//				return io.EOF
//			}
//			b = b.Bucket([]byte(userid))
//			if b == nil {
//				return io.EOF
//			}
//			v := b.Get([]byte(orderid))
//			if v == nil {
//				return io.EOF
//			}
//			b = t.Bucket(order)
//			if b == nil {
//				return io.EOF
//			}
//			v = b.Get([]byte(orderid))
//			if v == nil {
//				return io.EOF
//			}
//			var db map[string]interface{}
//			err := json.Unmarshal(v,&db)
//			if err != nil {
//				return err
//			}
//			if db["goodsid"]== nil {
//				return io.EOF
//			}
//			hand(db)
//			return nil
//		})
//	})
//	//return orderGet(orderid,userid,hand)
//}
//func orderGet(orderid,userid string,hand func(interface{}))error{
//	return openSiteDB(orderDB,func(DB *bolt.DB)error{
//		return  DB.View(func(t *bolt.Tx)error{
//			b := t.Bucket(order)
//			if b == nil {
//				return io.EOF
//			}
//			v := b.Get([]byte(orderid))
//			if v == nil {
//				return io.EOF
//			}
//			var db map[string]interface{}
//			err := json.Unmarshal(v,&db)
//			if err != nil {
//				return err
//			}
//			if userid != ""{
//			//fmt.Println(db)
//			uid := db["userid"]
//			if uid != nil &&  uid.(string) != userid {
//				return io.EOF
//			}
//			}
//			hand(db)
//			return nil
//		})
//	})
//}
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
	return db.Batch(func(t *bolt.Tx)error{
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
			shop := hand(sh,dbname)
			ShoppingMap.Store(sh.Py,shop)

		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
