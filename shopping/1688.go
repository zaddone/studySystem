package shopping

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/zaddone/studySystem/request"
	"net/http"
	"net/url"
	"strings"
	"time"

	//"golang.org/x/text/encoding/simplifiedchinese"
	//"golang.org/x/text/transform"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
)

var (
	//1688Url = "https://gw.open.1688.com/openapi/param2/%s/6020087"
	Url1688           = "https://gw.open.1688.com/openapi/"
	AlibabaShopping   *Alibaba
	alibabatimeFormat = "20060102150405000-0700"
	goodsDB           = []byte("product")
	goodsList         = []byte("productL")
	//GoodsListDB = []byte("productList")
)

//
//func GbkToUtf8(s []byte) ([]byte, error) {
//	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
//	d, e := ioutil.ReadAll(reader)
//	if e != nil {
//		return nil, e
//	}
//	return d, nil
//}

type Alibaba struct {
	Info   *ShoppingInfo
	DbPath string
	//Pid string
}

type AlAddrForOrder struct {
	FullName     string `json:"fullName"`
	Mobile       string `json:"mobile"`
	CityText     string `json:"cityText"`
	ProvinceText string `json:"provinceText"`
	AreaText     string `json:"areaText"`
	TownText     string `json:"townText"`
	Address      string `json:"address"`
}

func (self *AlAddrForOrder) LoadTestDB() {
	self.FullName = "赵伟杰"
	self.Mobile = "18628175526"
	self.CityText = "成都市"
	self.ProvinceText = "四川省"
	self.AreaText = "郫都区"
	self.TownText = "犀浦街道"
	self.Address = "校园路55号交大卡布里城1栋1单元1708号"
}

type AlProductForOrder struct {
	Offerid  float64 `json:"offerId"`
	SpecId   string  `json:"specId"`
	Quantity float64 `json:"quantity"`
}

func (self *AlProductForOrder) LoadTestDB() {
	self.Offerid = 586899647105
	self.SpecId = "4fac2a41ec29cef08c68c5cac25382d8"
	self.Quantity = 1
}

func NewAlibaba(sh *ShoppingInfo, siteDB string) *Alibaba {
	j := &Alibaba{Info: sh, DbPath: "alibaba.db"}
	if siteDB == "" {
		return j
	}
	return j
	//return j
	go func() {
		for {
			if j.Info.ReTimeOut == "" {
				j.Info.ReTimeOut = "20201120194552000+0800"
			}
			reTimeOut, err := time.Parse(alibabatimeFormat, j.Info.ReTimeOut)
			if err != nil {
				panic(err)
			}
			select {
			case <-time.After(time.Second * time.Duration(j.Info.TimeOut-time.Now().Unix())):
				err := j.ReToken(siteDB)
				if err != nil {
					panic(err)
				}
			case <-time.After(time.Second * time.Duration(reTimeOut.Unix()-time.Now().Unix())):
				err := j.ReToken_(siteDB)
				if err != nil {
					panic(err)
				}
			}
		}
	}()
	return j
}

func (self *Alibaba) ReToken_(siteDB string) error {
	uri := "https://gw.open.1688.com/openapi/param2/1/system.oauth2/postponeToken/" + self.Info.Client_id
	u := url.Values{}
	u.Set("client_id", self.Info.Client_id)
	u.Set("client_secret", self.Info.Client_secret)
	u.Set("refresh_token", self.Info.ReToken)
	u.Set("access_token", self.Info.Token)
	return request.ClientHttp_(
		uri+"?"+u.Encode(),
		"POST", nil, nil,
		func(body io.Reader, start int) error {
			var res map[string]interface{}
			err := json.NewDecoder(body).Decode(&res)
			if err != nil {
				return err
			}
			fmt.Println("retoken_", res)
			if res["access_token"] == nil {
				return io.EOF
			}
			self.Info.Token = res["access_token"].(string)
			self.Info.ReToken = res["refresh_token"].(string)
			//self.Info.TimeOut =time.Now().Unix()+int64(res["expires_in"].(float64))
			self.Info.Uri = res["memberId"].(string)
			self.Info.ReTimeOut = res["refresh_token_timeout"].(string)

			exp, err := strconv.Atoi(res["expires_in"].(string))
			if err != nil {
				return err
			}
			self.Info.TimeOut = time.Now().Unix() + int64(exp)
			return OpenSiteDB(siteDB, self.Info.SaveToDB)
		},
	)

}

func (self *Alibaba) ReToken(siteDB string) error {

	uri := "https://gw.open.1688.com/openapi/param2/1/system.oauth2/getToken/" + self.Info.Client_id
	u := url.Values{}
	u.Set("client_id", self.Info.Client_id)
	u.Set("client_secret", self.Info.Client_secret)
	u.Set("grant_type", "refresh_token")
	u.Set("refresh_token", self.Info.ReToken)
	return request.ClientHttp_(
		uri+"?"+u.Encode(),
		"POST", nil, nil,
		func(body io.Reader, start int) error {
			var res map[string]interface{}
			err := json.NewDecoder(body).Decode(&res)
			if err != nil {
				return err
			}
			//fmt.Println(res)
			fmt.Println("retoken", res)
			if res["access_token"] == nil {
				return io.EOF
			}
			self.Info.Token = res["access_token"].(string)
			//self.Info.ReToken=res["refresh_token"].(string)
			//self.Info.TimeOut =time.Now().Unix()+int64(res["expires_in"].(float64))
			exp, err := strconv.Atoi(res["expires_in"].(string))
			if err != nil {
				return err
			}
			self.Info.TimeOut = time.Now().Unix() + int64(exp)
			self.Info.Uri = res["memberId"].(string)
			return OpenSiteDB(siteDB, self.Info.SaveToDB)
		},
	)

}
func (self *Alibaba) ClientHttp(uri string, u *url.Values) (out interface{}) {

	u.Add("memberId", self.Info.Uri)
	u.Add("_aop_timestamp", fmt.Sprintf("%d", time.Now().Unix()*1000))
	var li []string
	for k, _ := range *u {
		li = append(li, k+u.Get(k))
	}
	sort.Strings(li)
	uri = "param2/" + uri + "/" + self.Info.Client_id
	sign := uri + strings.Join(li, "")
	//key := []byte("123456")
	fmt.Println(self.Info, sign)
	mac := hmac.New(sha1.New, []byte(self.Info.Client_secret))
	mac.Write([]byte(sign))
	u.Add("_aop_signature", fmt.Sprintf("%X", mac.Sum(nil)))

	var err error
	uri = Url1688 + uri + "?" + u.Encode()
	err = request.ClientHttp_(
		uri,
		"POST", nil,
		nil,
		func(body io.Reader, start int) error {
			if start != 200 {
				db, err := ioutil.ReadAll(body)
				if err != nil {
					return err
				}
				return fmt.Errorf("%s", db)
			}
			return json.NewDecoder(body).Decode(&out)
		})
	if err != nil {
		fmt.Println(err, out)
		out = err
		//time.Sleep(time.Second*1)
		//return self.ClientHttp(u)
		//panic(err)
	}
	return
}

func (self *Alibaba) GetTraceView(id string) interface{} {
	//com.alibaba.logistics:alibaba.trade.getLogisticsInfos.buyerView-1
	uri := "1/com.alibaba.logistics/alibaba.trade.getLogisticsInfos.buyerView"
	u := &url.Values{}
	u.Add("orderId", id)
	u.Add("webSite", "1688")
	u.Add("fields", "company.name,sender,receiver,sendgood")
	u.Add("access_token", self.Info.Token)
	return self.ClientHttp(uri, u)
}
func (self *Alibaba) GetTraceInfo(id string) interface{} {
	//com.alibaba.logistics:alibaba.trade.getLogisticsTraceInfo.buyerView-1
	uri := "1/com.alibaba.logistics/alibaba.trade.getLogisticsTraceInfo.buyerView"
	u := &url.Values{}
	u.Add("orderId", id)
	u.Add("webSite", "1688")
	u.Add("access_token", self.Info.Token)
	return self.ClientHttp(uri, u)

}

func (self *Alibaba) GetCategory(id string) interface{} {
	uri := "1/com.alibaba.product/alibaba.category.get"
	u := &url.Values{}
	u.Add("categoryID", id)
	u.Add("access_token", self.Info.Token)
	obj := self.ClientHttp(uri, u)
	if obj == nil {
		return nil
	}
	return obj.(map[string]interface{})["categoryInfo"]
	//fmt.Println(obj)
	//return obj
	//com.alibaba.product:alibaba.category.get-1
}
func (self *Alibaba) ClearOrder(orderid string) interface{} {
	//alibaba.trade.cancel
	uri := "1/com.alibaba.trade/alibaba.trade.cancel"
	u := &url.Values{}
	u.Add("webSite", "1688")
	u.Add("tradeID", orderid)
	u.Add("cancelReason", "other")
	u.Add("access_token", self.Info.Token)
	return self.ClientHttp(uri, u)

}

func (self *Alibaba) PreviewCreateOrder(a *AlAddrForOrder, p []*AlProductForOrder) interface{} {
	addr_, err := json.Marshal(a)
	if err != nil {
		return err
	}
	product_, err := json.Marshal(p)
	if err != nil {
		return err
	}
	//com.alibaba.trade:alibaba.trade.fastCreateOrder-1
	uri := "1/com.alibaba.trade/alibaba.createOrder.preview"
	u := &url.Values{}
	u.Add("flow", "saleproxy")
	u.Add("addressParam", string(addr_))
	//u.Add("addressParam","")
	u.Add("cargoParamList", string(product_))
	u.Add("access_token", self.Info.Token)
	u.Add("invoiceParam", "")
	return self.ClientHttp(uri, u)
	//fmt.Println(obj)

}
func (self *Alibaba) CreateOrder(a *AlAddrForOrder, p []*AlProductForOrder) interface{} {

	addr_, err := json.Marshal(a)
	if err != nil {
		return err
	}
	product_, err := json.Marshal(p)
	if err != nil {
		return err
	}
	//com.alibaba.trade:alibaba.trade.fastCreateOrder-1
	uri := "1/com.alibaba.trade/alibaba.trade.fastCreateOrder"
	u := &url.Values{}
	u.Add("flow", "saleproxy")
	u.Add("addressParam", string(addr_))
	u.Add("cargoParamList", string(product_))
	u.Add("access_token", self.Info.Token)
	return self.ClientHttp(uri, u)

}

func (self *Alibaba) GoodsGetWithTaobao(taobaoId string, hand func(interface{})) error {
	return self.OpenDB(false, func(t *bolt.Tx) error {
		b_ := t.Bucket(goodsList)
		if b_ == nil {
			return nil
		}
		b := t.Bucket(goodsDB)
		if b == nil {
			return nil
		}
		val := b.Get(b_.Get([]byte(taobaoId)))
		if val == nil {
			return fmt.Errorf("find not")
		}
		var db interface{}
		err := json.Unmarshal(val, &db)
		if err != nil {
			return err
		}
		hand(db)
		return nil
	})

}
func (self *Alibaba) GoodsGet(goodsId string, hand func(interface{})) error {
	return self.OpenDB(false, func(t *bolt.Tx) error {
		b := t.Bucket(goodsDB)
		if b == nil {
			return nil
		}
		val := b.Get([]byte(goodsId))
		if val == nil {
			return fmt.Errorf("find not")
		}
		var db interface{}
		err := json.Unmarshal(val, &db)
		if err != nil {
			return err
		}
		hand(db)
		return nil
	})
}
func (self *Alibaba) GoodsShowList(num []byte, hand func(k, v []byte) error) error {
	return self.OpenDB(false, func(t *bolt.Tx) error {
		b := t.Bucket(goodsDB)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		var k, v []byte
		if len(num) == 0 {
			k, v = c.First()
			if k == nil {
				return nil
			}
		} else {
			k, v = c.Seek(num)
			if bytes.Equal(k, num) {
				k, v = c.Next()
			}
		}
		var err error
		for ; k != nil; k, v = c.Next() {
			err = hand(k, v)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (self *Alibaba) GoodsShow(num []byte, hand func(interface{}) error) error {
	return self.OpenDB(false, func(t *bolt.Tx) error {
		b := t.Bucket(goodsDB)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		var k, v []byte
		if len(num) == 0 {
			k, v = c.First()
			if k == nil {
				return nil
			}
		} else {
			k, v = c.Seek(num)
			if bytes.Equal(k, num) {
				k, v = c.Next()
			}
		}
		var err error
		for ; k != nil; k, v = c.Next() {
			var db interface{}
			err = json.Unmarshal(v, &db)
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
}
func GetWordsKey(db string) []string {
	regK := regexp.MustCompile(`[0-9a-zA-Z]+|\p{Han}`)
	keymap := map[string]int{}
	for _, str := range regexp.MustCompile(`[0-9|a-z|A-Z|\p{Han}]+`).FindAllString(db, -1) {
		//fmt.Println(i, str)
		words := regK.FindAllString(str, -1)
		for i_, w := range words {
			//s := w_
			keymap[w]++
			for _, w_ := range words[i_+1:] {
				w += w_
				keymap[w]++
				//fmt.Println(w)
			}
		}
	}
	sum := 0
	var words []string
	for k, _ := range keymap {
		words = append(words, k)
		//if v > 1 {
		//	sum++
		//	fmt.Println(k, v)
		//}
	}
	fmt.Println(len(keymap), sum)
	//regK := regexp.MustCompile(`[0-9a-zA-Z]+|\p{Han}`)
	//for i, k := range regK.FindAllString(db, -1) {
	//	fmt.Println(i, k)
	//}
	return words
}

func Get1688GoodsDetail(db_ interface{}) interface{} {
	db_m := db_.(map[string]string)
	db := db_m["body"]
	cou := db_m["tfscom"]

	//video := regexp.MustCompile(`"address":"([\w\:\.\/_]+)"`).FindStringSubmatch(db_m["video"])
	video := regexp.MustCompile(`"address":"([^"]+)"`).FindStringSubmatch(db_m["video"])
	video_str := ""
	if len(video) > 1 {
		video_str = video[1]
	}
	title_str := ""
	title := regexp.MustCompile(`property="og:title" content="([^"]+)"`).FindStringSubmatch(db)
	if len(title) > 1 {
		title_str = title[1]
	}
	var keywords []string

	//wordskey := regexp.MustCompile(`property="og:description" content="([^"]+)"`).FindStringSubmatch(db)
	//if len(wordskey) > 1 {
	//	keywords = GetWordsKey(wordskey[1])
	//}

	price_str := ""
	price := regexp.MustCompile(`refPrice:'([^']+)'`).FindStringSubmatch(db)
	if len(price) > 1 {
		price_str = price[1]
	}
	fmt.Println(video_str, title_str, price_str)
	//fmt.Println(video, db_m["video"][1])
	//fmt.Println(cou)
	var des_img []string
	for _, im := range regexp.MustCompile(`src=\\"([^\\]+)\\"`).FindAllStringSubmatch(cou, -1) {
		des_img = append(des_img, im[1])
	}
	fmt.Println(des_img)
	var imgs []string
	for _, im := range regexp.MustCompile(`data-imgs='.+`).FindAllString(db, -1) {
		_ims := regexp.MustCompile(`"original":"(.+)"`).FindStringSubmatch(im)
		//fmt.Println(_ims)
		imgs = append(imgs, _ims[1])
	}
	skumap_ := regexp.MustCompile(`skuMap:(.+)`).FindStringSubmatch(db)
	var skumap_db interface{}
	var skumap__ string
	if len(skumap_) > 1 {
		skumap__ = skumap_[1][0 : len(skumap_[1])-1]

		err := json.Unmarshal([]byte(skumap__), &skumap_db)
		if err != nil {
			fmt.Println(string(skumap__))
			return err
		}
	}
	var props_db interface{}
	var props__ string

	props_ := regexp.MustCompile(`skuProps:(.+)`).FindStringSubmatch(db)
	if len(props_) > 1 {
		props__ = props_[1][0 : len(props_[1])-1]
		err := json.Unmarshal([]byte(props__), &props_db)
		if err != nil {
			fmt.Println(props__)
			return err
		}
	}
	md5str := fmt.Sprintf("%x", md5.Sum([]byte(title_str+price_str+video_str+props__+skumap__+strings.Join(imgs, "")+strings.Join(des_img, ""))))
	//fmt.Println(md5str)

	return map[string]interface{}{
		"videoUrl":     video_str,
		"props":        props_db,
		"skumap":       skumap_db,
		"imgs":         imgs,
		"des_img":      des_img,
		"productTitle": title_str,
		"SellPrice":    price_str,
		"md5str":       md5str,
		"keywords":     keywords,
	}

}

func (self *Alibaba) GoodsDetailForUrl(words ...string) interface{} {

	u := fmt.Sprintf("https://detail.1688.com/offer/%s.html", words[0])
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res.Status)
	}
	db, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	//props,err :=GbkToUtf8(regexp.MustCompile(`skuProps:.+`).Find(db))
	//if err != nil {
	//	return err
	//}
	var imgs []string
	for _, im := range regexp.MustCompile(`data-imgs='.+`).FindAll(db, -1) {
		_ims := regexp.MustCompile(`"original":"(.+)"`).FindStringSubmatch(string(im))
		//fmt.Println(_ims)
		imgs = append(imgs, _ims[1])
	}
	skumap_ := regexp.MustCompile(`skuMap:(.+)`).FindSubmatch(db)
	if len(skumap_) < 2 {
		fmt.Println(string(db))
		return fmt.Errorf("find not skumap")
	}
	skumap__, err := GbkToUtf8(skumap_[1][0 : len(skumap_[1])-1])
	if err != nil {
		return err
	}
	var skumap_db interface{}
	err = json.Unmarshal(skumap__, &skumap_db)
	if err != nil {
		fmt.Println(string(skumap__))
		return err
	}
	props_ := regexp.MustCompile(`skuProps:(.+)`).FindSubmatch(db)
	if len(props_) < 2 {
		return fmt.Errorf("find not props")
	}
	props__, err := GbkToUtf8(props_[1][0 : len(props_[1])-1])
	if err != nil {
		return err
	}
	var props_db interface{}
	err = json.Unmarshal(props__, &props_db)
	if err != nil {
		fmt.Println(string(props__))
		return err
	}

	return map[string]interface{}{
		"videoUrl": fmt.Sprintf(
			"http://cloud.video.taobao.com/play/u/%s/p/2/e/6/t/1/%s.mp4",
			string(regexp.MustCompile(`"userId":"(\w+)"`).FindSubmatch(db)[1]),
			string(regexp.MustCompile(`"videoId":"(\w+)"`).FindSubmatch(db)[1]),
		),
		"props":  props_db,
		"skumap": skumap_db,
		"imgs":   imgs,
	}

	//return string(db)
	//fmt.Println(string(db))
	//return nil

}

func (self *Alibaba) GoodsDetail(words ...string) interface{} {
	uri := "1/com.alibaba.product/alibaba.agent.product.simple.get"
	u := &url.Values{}
	u.Add("webSite", "1688")
	u.Add("productID", words[0])
	u.Add("access_token", self.Info.Token)
	obj := self.ClientHttp(uri, u)
	//fmt.Println(obj)
	return obj
}

func (self *Alibaba) OpenDB(read bool, hand func(*bolt.Tx) error) error {

	db, err := bolt.Open(self.DbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	t, err := db.Begin(read)
	if err != nil {
		return err
	}
	if read {
		defer t.Commit()
	}
	return hand(t)

}

func (self *Alibaba) DelGoods(k string) error {
	return self.OpenDB(true, func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists(goodsDB)
		if err != nil {
			return err
		}
		return b.Delete([]byte(k))
	})
}

func (self *Alibaba) SaveGoods(k string, obj interface{}) error {
	return self.OpenDB(true, func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists(goodsDB)
		if err != nil {
			return err
		}
		db, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		return b.Put([]byte(k), db)
	})
}
func (self *Alibaba) ClearDB(db []byte) error {
	return self.OpenDB(true, func(t *bolt.Tx) error {
		return t.DeleteBucket(db)
	})
}
func (self *Alibaba) ClearProduct() error {

	return self.OpenDB(true, func(t *bolt.Tx) error {
		return t.DeleteBucket(goodsDB)
	})

}
func (self *Alibaba) HandGoodsListT(id string, show bool, up func(interface{}) error) error {
	id_ := []byte(id)
	return self.OpenDB(true, func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists(goodsDB)
		if err != nil {
			return err
		}
		//b_,err := t.CreateBucketIfNotExists(goodsListDB)
		//if err != nil {
		//	return err
		//}
		c := b.Cursor()
		var k, v []byte
		if len(id) == 0 {
			k, v = c.First()
		} else {
			k, v = c.Seek(id_)
			if bytes.Equal(id_, k) {
				k, v = c.Next()
			}
		}
		for ; k != nil; k, v = c.Next() {
			//if len(v) == 1 {
			//	if show{
			//		continue
			//	}
			//	v = b.Get(k)
			//}
			var db interface{}
			err = json.Unmarshal(v, &db)
			if err != nil {
				return err
			}
			err = up(db)
			if err != nil {
				return err
			}
		}
		return nil

	})
}

//func (self *Alibaba) HandGoodsList(lis string,up func(interface{})error)error{
//	return self.OpenDB(true,func(t *bolt.Tx)error{
//		b,err := t.CreateBucketIfNotExists(goodsDB)
//		if err != nil {
//			return err
//		}
//		//b_,err := t.CreateBucketIfNotExists(GoodsListDB)
//		//if err != nil {
//		//	return err
//		//}
//		return b.ForEach(func(k,v []byte)error{
//			if len(lis)>0{
//				if !strings.Contains(lis,string(k)){
//					return b_.Delete(k)
//				}
//			}
//			if len(v) == 1{
//				var db interface{}
//				err = json.Unmarshal(b.Get(k),&db)
//				if err != nil {
//					return err
//				}
//				return up(db)
//			}
//			return nil
//		})
//
//	})
//}
func (self *Alibaba) SaveProductList(k, pid string) error {
	return self.OpenDB(true, func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists(goodsList)
		if err != nil {
			return err
		}
		return b.Put([]byte(pid), []byte(k))
	})
}
func (self *Alibaba) SaveProduct(k string, obj interface{}) error {
	return self.OpenDB(true, func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists(goodsDB)
		if err != nil {
			return err
		}
		//b_,err := t.CreateBucketIfNotExists(GoodsListDB)
		//if err != nil {
		//	return err
		//}
		key := []byte(k)
		v, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		//_db := b.Get(key)
		//if len(_db) == len(v)  {
		//	return nil
		//}
		//err = b_.Put(key,[]byte{0})
		//if err != nil {
		//	return err
		//}

		return b.Put(key, v)

	})
}

func (self *Alibaba) Crossborder(id string) error {

	uri := "1/com.alibaba.product/alibaba.product.follow.crossborder"
	u := &url.Values{}
	u.Add("productId", id)
	u.Add("access_token", self.Info.Token)
	obj := self.ClientHttp(uri, u)
	if obj.(map[string]interface{})["code"].(float64) == 0 {
		return nil
	}
	//fmt.Println(obj)
	return fmt.Errorf("%v", obj)
}

func (self *Alibaba) UnCrossborder(id string) error {

	uri := "1/com.alibaba.product/alibaba.product.unfollow.crossborder"
	u := &url.Values{}
	u.Add("productId", id)
	u.Add("access_token", self.Info.Token)
	obj := self.ClientHttp(uri, u)
	//fmt.Println(obj)
	//return obj
	if obj.(map[string]interface{})["code"].(float64) == 0 {
		return nil
	}
	//fmt.Println(obj)
	return fmt.Errorf("%v", obj)
}

func (self *Alibaba) SearchGoods(words ...string) interface{} {

	uri := "1/com.alibaba.search/alibaba.search.cbu.general"
	u := &url.Values{}
	u.Add("q", words[0])
	u.Add("access_token", self.Info.Token)
	obj := self.ClientHttp(uri, u)
	fmt.Println(obj)
	return obj

}
