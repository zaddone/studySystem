package main

import (
	"fmt"
	//"bytes"
	//"time"
	"io"
	//"os"
	"net/http"
	//"net/url"
	//"strings"
	//"strconv"
	"github.com/gin-gonic/gin"
	//"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/wxmsgb"
	//"github.com/zaddone/studySystem/chromeServer"
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/zaddone/studySystem/alibaba"
	"github.com/zaddone/studySystem/shopping"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"io/ioutil"
	"strings"
)

var (
	decoder *encoding.Decoder = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
)

//func MapToStr(db map[string]interface{})string{
//
//	var vl  []string
//	for k,v := range db {
//		switch _v := v.(type){
//		case string:
//			vl = append(vl,fmt.Sprintf("%s:\"%s\"",k,_v)
//		case []string:
//			vl = append(vl,fmt.Sprintf("%s:[\"%s\"]",k,strings.Join(_v,"\",\""))
//		case float64:
//			vl = append(vl,fmt.Sprintf("%s:%.0f",k,_v)
//		default:
//			strdb,err := json.Marshal(_v)
//			if err != nil { //				panic(err)
//			}
//			vl = append(vl,fmt.Sprintf("%s:\"%s\"",k,string(strdb))
//
//		}
//	}
//	fmt.Sprintf("{%s}",strings.Join(vl,","))
//}

func handTaobaoFile(f io.Reader, hand func(interface{})) error {

	buf := bufio.NewReader(decoder.Reader(f))
	var dbval []byte
	line := 0
	//var fields []string
	for {
		li, isp, err := buf.ReadLine()
		if err != nil {
			return err
			fmt.Println(err)
			break
		}
		if isp {
			dbval = append(dbval, li...)
			//fmt.Println(len(dbval))
			continue
			//panic(err)
		}
		if len(dbval) > 0 {
			li = append(dbval, li...)
			dbval = nil
		}
		if line > 1 {
			lis := strings.Split(string(li), "	")
			hand(lis)
			//for _i,l := range lis {
			//fmt.Println(_i,l)
			//}
		}
		line++
	}
	return nil
}

func getDesImg(des string, hand func(string)) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(des))
	if err != nil {
		return err
	}
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		v, _ := s.Attr("src")
		//fmt.Println(v,e)
		hand(v)
	})
	return nil
}
func PreviewOrder(obj map[string]interface{}, ali *shopping.Alibaba) error {
	pro := obj["productInfo"].(map[string]interface{})
	skus := pro["skuInfos"]
	if skus == nil {
		return fmt.Errorf("obj is nil")
	}
	//var sku map[string]interface{}
	var ss []interface{}
	for _, s := range skus.([]interface{}) {
		if s.(map[string]interface{})["amountOnSale"].(float64) > 1 {
			ss = append(ss, s)
		}
	}
	pro["skuInfos"] = ss
	po := &shopping.AlProductForOrder{
		Offerid:  pro["productID"].(float64),
		SpecId:   ss[0].(map[string]interface{})["specId"].(string),
		Quantity: 1,
	}
	addr := &shopping.AlAddrForOrder{}
	addr.LoadTestDB()
	for {
		res := ali.PreviewCreateOrder(addr, []*shopping.AlProductForOrder{po})
		switch r := res.(type) {
		case error:
			return r
		default:
			errcode := r.(map[string]interface{})["errorCode"]
			if errcode == nil {
				obj["Preview"] = r
				return nil
			}
			ec := errcode.(string)
			if ec == "500_005" || ec == "500_006" {
				fmt.Println(r)
				po.Quantity++
			} else {
				return fmt.Errorf("%s", errcode)
			}
		}
	}
	return nil

}

func handGoods(obj map[string]interface{}) interface{} {
	pro := obj["productInfo"].(map[string]interface{})
	if pro["skuInfos"] == nil {
		return nil
	}
	op_ := obj["Preview"].(map[string]interface{})["orderPreviewResuslt"]
	if op_ == nil {
		return nil
	}
	op_list := op_.([]interface{})
	if len(op_list) == 0 {
		return nil
	}
	//pro["subject_old"] = pro["subject"]
	resu := op_list[0].(map[string]interface{})
	pro["price"] = fmt.Sprintf("%.2f", resu["sumPayment"].(float64)*1.1/100)
	cl := resu["cargoList"].([]interface{})[0].(map[string]interface{})
	pro["NumMin"] = cl["amount"].(float64) / cl["finalUnitPrice"].(float64)
	if pro["NumMin"].(float64) < 1 {
		pro["NumMin"] = 1
	}
	Carriage := resu["sumCarriage"].(float64) / 100
	attrName := []string{}
	for i, _v := range pro["skuInfos"].([]interface{}) {
		v := _v.(map[string]interface{})
		skuName := ""
		for _, v_ := range v["attributes"].([]interface{}) {
			_v_ := v_.(map[string]interface{})
			if _v_["skuImageUrl"] != nil {
				v["imageUrl"] = "https://cbu01.alicdn.com/" + _v_["skuImageUrl"].(string)
			}
			skuName += _v_["attributeValue"].(string)
			if i == 0 {
				attrName = append(attrName, _v_["attributeDisplayName"].(string))
			}
		}
		v["skuName"] = skuName
		if v["price"] == nil {
			v["price"] = pro["price"]
		} else {
			v["price"] = fmt.Sprintf("%.2f", (v["price"].(float64)+Carriage)*1.1)
		}
	}
	pro["attrName"] = strings.Join(attrName, "/")

	images := pro["image"].(map[string]interface{})["images"].([]interface{})
	for i, image := range images {
		images[i] = "https://cbu01.alicdn.com/" + image.(string)
	}
	return pro
}

//func UpdateGoods(id string,obj interface{})error{
//	if obj == nil {
//		return fmt.Errorf("obj is nil")
//	}
//	u := url.Values{}
//	u.Add("id",id)
//	addSign(&u)
//	db,err := json.Marshal(obj)
//	if err != nil {
//		return err
//	}
//	return request.ClientHttp_("https://www.zaddone.com/site/v2/goods/update?"+u.Encode(),"POST",bytes.NewReader(db),nil,func(body io.Reader,re int)error{
//		if re != 200 {
//			d,_ := ioutil.ReadAll(body)
//			return fmt.Errorf("%d %s",re,d)
//		}
//		return nil
//	})
//}

func initAlibaba(hand func(*shopping.Alibaba) error) error {
	Info := &shopping.ShoppingInfo{}
	err := requestHttp("/shopping/1688", "GET", nil, nil, func(body io.Reader, res *http.Response) error {
		return json.NewDecoder(body).Decode(Info)
	})
	if err != nil {
		return err
	}
	return hand(shopping.NewAlibaba(Info, ""))
}

func init() {

	goods := Router.Group("goods")
	goods.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "goods.tmpl", nil)
	})
	goods.GET("/add", func(c *gin.Context) {
		//id := c.Query("id")
		c.HTML(http.StatusOK, "add.tmpl", nil)
	})
	//goods.POST("/save",func(c *gin.Context){
	//	id := c.Query("id")
	//	if id == "" {
	//		c.JSON(http.StatusFound,fmt.Errorf("id = nil"))
	//		return
	//	}
	//	u := url.Values{}
	//	u.Set("id",id)
	//	addSign(&u)
	//	err := request.ClientHttp_("https://www.zaddone.com/site/v2/goods/update/edit?"+u.Encode(),"POST",c.Request.Body,nil,func(body io.Reader,re int)error{

	//		db,err := ioutil.ReadAll(body)
	//		if err != nil {
	//			return err
	//		}
	//		c.String(re,string(db))
	//		return nil
	//	})
	//	if err != nil {
	//		c.JSON(http.StatusFound,err)
	//	}

	//})
	//goods.GET("/show",func(c *gin.Context){
	//	u := url.Values{}
	//	u.Add("goodsid",c.Query("id"))
	//	u.Add("show",c.Query("show"))
	//	u.Add("con","8")
	//	addSign(&u)
	//	err := request.ClientHttp_("https://www.zaddone.com/site/v2/goods/update/list_t?"+u.Encode(),"GET",nil,nil,func(body io.Reader,re int)error{
	//		if re != 200 {
	//			return fmt.Errorf("%d",re)
	//		}
	//		var db interface{}
	//		err := json.NewDecoder(body).Decode(&db)
	//		if err != nil {
	//			return err
	//		}
	//		c.JSON(http.StatusOK,db)
	//		return nil
	//	})
	//	if err != nil {
	//		c.JSON(http.StatusNotFound,err)
	//		return
	//	}

	//})
	//goods.GET("/order",func(c *gin.Context){
	//	err := initAlibaba(func(ali *shopping.Alibaba)error {
	//		o := new(shopping.AlAddrForOrder)
	//		p := new(shopping.AlProductForOrder)
	//		o.LoadTestDB()
	//		p.LoadTestDB()
	//		obj := ali.CreateOrder(o,[]*shopping.AlProductForOrder{p})
	//		c.JSON(http.StatusOK,obj)
	//		return nil
	//	})
	//	if err != nil {
	//		c.JSON(http.StatusOK,err)
	//		return
	//	}
	//})
	goods.POST("/func", func(c *gin.Context) {
		db, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			panic(err)
		}
		c.Request.Body.Close()
		fmt.Println(string(db))
		err = wxmsgb.FuncWXDB(c.Query("name"), bytes.NewReader(db), func(db interface{}) error {
			c.JSON(http.StatusOK, db)
			return nil
		})

		if err != nil {
			c.JSON(http.StatusNotFound, err)
		}

	})

	goods.POST("/upload", func(c *gin.Context) {
		file, _, err := c.Request.FormFile("upload")
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusOK, err)
			return
		}
		//filename := header.Filename
		defer file.Close()
		var li []interface{}

		err = initAlibaba(func(ali *shopping.Alibaba) error {
			return handTaobaoFile(file, func(l interface{}) {
				for i, l_ := range l.([]string) {
					fmt.Println(i, l_)
				}
				//fmt.Println(l)
				li = append(li, l)
				//l_ := l.([]interface{})
				//ali.GoodsGetWithTaobao(l_[36].(string),func(db interface{}){
				//	//db
				//	li = append(li,db)
				//})
			})
		})
		if len(li) > 0 {
			c.JSON(http.StatusOK, gin.H{"db": li})
			return
		}
		c.JSON(http.StatusOK, err)
	})
	goods.GET("/url", func(c *gin.Context) {

		alibaba.HandGoods = func(db interface{}) {
			//fmt.Println(db)
			c.JSON(http.StatusOK, shopping.Get1688GoodsDetail(db))
		}
		err := alibaba.RunDetail(c.Query("u"))
		fmt.Println(err)

	})
	goods.GET("/down_test", func(c *gin.Context) {
		list := []interface{}{}
		fn := "goods"

		err := initAlibaba(func(ali *shopping.Alibaba) error {
			alibaba.HandGoods = func(db interface{}) {
				//db_:= db.(map[string]interface{})
				//productId :=fmt.Sprintf("%.0f",db_["productId"].(float64))
				//itemId := fmt.Sprintf("%.0f",db_["itemId"].(float64))
				//fmt.Println(db_)
				list = append(list, db)
				//detail_ := ali.GoodsDetailForUrl(productId)
				//switch detail:=detail_.(type){
				//case error:
				//	fmt.Println(detail)
				//	return
				//case map[string]interface{}:
				//	detail["productTitle"] = db_["productTitle"]
				//	detail["productId"] = productId
				//	detail["itemId"] = itemId
				//	detail["PurchasePrice"] = db_["minPurchasePrice"]
				//	detail["SellPrice"] = db_["minTbSellPrice"]
				//	body,err := json.Marshal(detail)
				//	if err != nil {
				//		panic(err)
				//	}
				//	//fmt.Println(detail)
				//	f.WriteString(fmt.Sprintf("{_id:\"%s\",body:%s}",productId,string(body)))
				//	list = append(list,string(body))
				//	time.Sleep(10*time.Second)
				//}
			}
			return alibaba.Run()
		})
		if len(list) > 0 {
			//f,err := os.OpenFile(fn,os.O_APPEND|os.O_CREATE|os.O_RDWR,0777)
			//if err != nil {
			//	c.JSON(http.StatusNotFound,err)
			//	return
			//}
			for i, li := range list {
				//i_ := i
				db_ := li.(map[string]interface{})
				productId := fmt.Sprintf("%.0f", db_["productId"].(float64))
				itemId := fmt.Sprintf("%.0f", db_["itemId"].(float64))
				alibaba.HandGoods = func(db interface{}) {
					d := shopping.Get1688GoodsDetail(db)
					switch detail := d.(type) {
					case error:
						fmt.Println(detail)
						return
					case map[string]interface{}:
						detail["productTitle"] = db_["productTitle"]
						detail["productId"] = productId
						detail["itemId"] = itemId
						detail["PurchasePrice"] = db_["minPurchasePrice"]
						detail["SellPrice"] = db_["minTbSellPrice"]
						body, err := json.Marshal(detail)
						if err != nil {
							panic(err)
						}
						fmt.Println(detail)
						//f.WriteString(fmt.Sprintf("{_id:\"%s\",body:%s}",productId,string(body)))
						list[i] = detail
						err = wxmsgb.UpdateWXDB(fn, productId, string(body))
						if err != nil {
							panic(err)
						}

						//time.Sleep(10*time.Second)
					}
					//fmt.Println(d)
				}
				err := alibaba.RunDetail(productId)
				fmt.Println(err)
			}
			//f.Close()
			//err = wxmsgb.UpDBToWX(fn,fn)
			//fmt.Println("end",err)
			//if err != nil {
			//	c.String(http.StatusNotFound,fmt.Sprint(err))
			//	return
			//}
			c.JSON(http.StatusOK, list)
			return
		}
		c.JSON(http.StatusNotFound, err)
		return
	})
	goods.GET("/down", func(c *gin.Context) {
		//var li []interface{}
		list := []string{}
		err := initAlibaba(func(ali *shopping.Alibaba) error {
			alibaba.HandGoods = func(db interface{}) {
				//li = append(li,db)
				db_ := db.(map[string]interface{})
				productId := fmt.Sprintf("%.0f", db_["productId"].(float64))
				itemId := fmt.Sprintf("%.0f", db_["itemId"].(float64))

				obj := ali.GoodsDetail(productId)
				obj_ := obj.(map[string]interface{})
				if obj_["productInfo"] == nil {
					return
				}
				fmt.Println(obj_)
				obj_["itemid"] = itemId
				product := obj_["productInfo"].(map[string]interface{})
				des := product["description"].(string)
				//obj_["des_img"] = []string{}
				var des_img []string
				err := getDesImg(des, func(img string) {
					//fmt.Println(img)
					des_img = append(des_img, img)
				})
				if err != nil {
					panic(err)
				}
				product["des_img"] = des_img

				err = PreviewOrder(obj_, ali)
				if err != nil {
					fmt.Println(err)
					return
				}
				pro := handGoods(obj_)
				if pro == nil {
					return
				}
				err = ali.SaveProduct(productId, pro)
				//err = UpdateGoods(productId,pro)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = ali.SaveProductList(productId, itemId)
				if err != nil {
					fmt.Println(err)
					return
				}
				//err = ali.Crossborder(productId)
				//if err != nil {
				//	//return err
				//	panic(err)
				//}
				//li = append(li,pro)
				list = append(list, productId)
			}
			return alibaba.Run()
		})
		if len(list) > 0 {
			c.JSON(http.StatusOK, list)
			return
		}
		c.JSON(http.StatusOK, err)
		return
	})

}
