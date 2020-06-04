package main
import(
	"fmt"
	"io"
	"strconv"
	//"time"
	"net/url"
	"github.com/gin-gonic/gin"
	"github.com/zaddone/studySystem/request"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"bytes"
	"regexp"
	"strings"
	//"sync"
	//"math"
	//"os"
	"github.com/nfnt/resize"
	"image/jpeg"
	"github.com/PuerkitoBio/goquery"
)

var(
	Category = "Category.db"
	WxCategory = []byte("WxCategory")
	WxWorks = []byte("works")
	DelWorks = []byte("delworks")
	keywordsReg = regexp.MustCompile(`[a-z0-9A-Z\p{Han}]+`)
	//DelKeywords = regexp.MustCompile(`定制|批发|一件代发|厂家直销|`)
	//imgUrlReg = regexp.MustCompile(`http[s?]\://[a-z0-9A-Z./_]+`)
)
//https://api.weixin.qq.com/merchant/express/add
func GetallExpress(hand func(interface{})error)error{
	//http.Post()
	return Request(
		"https://api.weixin.qq.com/merchant/express/getall",
		nil,
		hand)
}

func GetKeywords(w string,hand func(string)){
	for _,key := range keywordsReg.FindAllString(w,-1){
		for i,_w := range []rune(key){
			word := []rune{_w}
			for _,__w := range []rune(key)[i+1:]{
				word =append(word,__w)
				hand(string(word))
			}
		}
	}
}
func GoodsWithAlibabaToWX(obj interface{},hand func(interface{})) error {

	db := obj.(map[string]interface{})
	cats := db["cat"]
	if cats == nil {
		return fmt.Errorf("cat is not find")
	}
	product := db["productInfo"].(map[string]interface{})
	subject := product["subject"].(string)
	content := product["description"].(string)
	images := product["image"].(map[string]interface{})["images"].([]interface{})
	for i,img := range images{
		err := uploadImg("https://cbu01.alicdn.com/"+img.(string),
		func(_img string){
			images[i] = _img
		})
		if err != nil {
			fmt.Println(err)
			return err
		}
		//images[i] = "https://cbu01.alicdn.com/"+img.(string)
	}
	return openCategoryDB(false,func(t *bolt.Tx)error{
		b := t.Bucket(WxWorks)
		if b == nil {
			return nil
		}
		//bc := t.Bucket(WxCategory)
		//if bc == nil {
		//	return nil
		//}
		valMap := map[string]int{}
		sk := ""
		for _,cat := range cats.([]interface{}){
			catname := cat.(map[string]interface{})["name"].(string)
			key :=  keywordsReg.FindAllString(catname,-1)
			sk += catname
			for _,c := range key {
				v := b.Get([]byte(c))
				if v != nil {
					valMap[string(v)]++
				}
			}
		}
		if len(valMap) == 0 {
			GetKeywords(sk+subject,func(w string){
				//fmt.Println(subject,w)
				v := b.Get([]byte(w))
				if v != nil {
					valMap[string(v)]++
				}
			})

		}
		if len(valMap)==0{
			fmt.Println(subject)
			return nil
		}
		var max int
		var catn string
		for k,v:= range valMap{
			if v>max{
				catn = k
			}
		}

		//b__ := bc.Bucket([]byte(catn))
		//if b__ == nil {
		//	panic(0)
		//}
		doc,err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			return err
		}
		detail := []interface{}{
			map[string]interface{}{
				"text": "test first",
			},
		}
		doc.Find("img").Each(func(i int,s *goquery.Selection){
			img,_:=s.Attr("src")
			err := uploadImg(img,func(_img string){
				detail = append(detail,map[string]interface{}{
					"img":_img,
				})
			})
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(img)
			//detail = append(detail,map[string]interface{}{
			//	"img":img,
			//})
		})
		//imgs := imgUrlReg.FindAllString(content,-1)
		//fmt.Println(imgs)

		var catinfo []interface{}
		err = GetCategory(catn,t.Bucket(WxCategory),true,func(n string,db interface{})error{
			catinfo =append(catinfo,db)
			//fmt.Println(catinfo)
			return nil
		})
		if err != nil {
			return err
		}
		base_attr := map[string]interface{}{
			"main_img":images[0],
			"img":images,
			"name":subject,
			"detail":detail,
			"property":[]interface{}{},
			"buy_limit":"0",
			//"category_id":[]string{catinfo["id"].(string)},
			//"cats":catinfo,
		}
		var cats []string
		for _,c := range catinfo{
			cats = append(cats,c.(map[string]interface{})["id"].(string))
		}
		base_attr["category_id"] = cats
		var sku_list []interface{}
		//sku_info
		skumap := map[string]map[string]int{}
		pr_ :=strings.Split(product["referencePrice"].(string),"~")
		pr,err := strconv.ParseFloat(pr_[len(pr_)-1],64)
		if err != nil {
			return err
			panic(err)
		}
		pr = pr/0.99*100
		//pr,_ = strconv.ParseFloat(fmt.Sprintf("%.2f",pr/0.99),64)
		//fmt.Println(product)
		for _,p := range product["skuInfos"].([]interface{}){
			p_ := p.(map[string]interface{})

			obj := map[string]interface{}{
				"quantity":fmt.Sprintf("%.0f",p_["amountOnSale"].(float64)),
				"product_code":p_["specId"],
				//"ori_price":pr,
				"price":int(pr),
				//"icon_url":p_[""]
			}
			var attrl  []string
			for _,attr := range p_["attributes"].([]interface{}){
				attr_ := attr.(map[string]interface{})
				skuname := attr_["attributeDisplayName"].(string)
				skuval := attr_["attributeValue"].(string)
				l:= fmt.Sprintf("$%s:$%s",skuname,skuval)
				attrl = append(attrl,l)
				_val := skumap[skuname]
				if _val == nil {
					skumap[skuname] = map[string]int{skuval:1}
				}else{
					skumap[skuname][skuval]++
				}
				//skumap[l]=map[string]interface{}{
				//	"id":"$"+skuname,
				//	"vid":"$"+skuval,
				//}
				if attr_["skuImageUrl"] != nil {
					obj["icon_url"] ="https://cbu01.alicdn.com/" + attr_["skuImageUrl"].(string)
				}
			}

			if obj["icon_url"] != nil {
				err := uploadImg(obj["icon_url"].(string),func(_img string){
					obj["icon_url"] = _img
				})
				if err != nil {
					return err
				}
			}
			obj["sku_id"] = strings.Join(attrl,";")
			sku_list = append(sku_list,obj)
		}
		var sku_info []interface{}
		for k,v:= range skumap {
			//v_ := v.(map[string]interface{})
			//val_ := skumap_[v_["id"].(string)]
			//if val_ == nil {
			//	
			//}
			var v_ []string
			for _k,_ := range v {
				v_ = append(v_,"$"+_k)
			}
			sku_info = append(sku_info,map[string]interface{}{"id":"$"+k,"vid":v_})
		}
		base_attr["sku_info"] = sku_info

		delivery_info := map[string]interface{}{
			//"delivery_type":0,
			//"template_id":0,
			//"express":[]interface{}{
			//	map[string]interface{}{
			//		"id": 10000027,
			//		"price": 100,
			//	},
			//	map[string]interface{}{
			//		"id": 10000028,
			//		"price": 100,
			//	},
			//	map[string]interface{}{
			//		"id": 10000029,
			//		"price": 100,
			//	},
			//},
		}
		attrext := map[string]interface{}{
			"isPostFree":"1",
			"isHasReceipt":"0",
			"isUnderGuaranty":"0",
			"isSupportReplace":0,
			"location":map[string]interface{}{
				"country": "中国",
				"province": "四川省",
				"city": "成都市",
				"address":"",
			},
		}
		//fmt.Println(base_attr)
		base:=map[string]interface{}{
			"product_base":base_attr,
			"delivery_info":delivery_info,
			"sku_list":sku_list,
			"attrext":attrext,
		}
		hand(base)
		//return nil
		//return fmt.Errorf("%+v",base)
		return Request(
			"https://api.weixin.qq.com/merchant/create",
			base,
			func(res interface{})error{
				fmt.Println(res)
				return nil
			},
		)
		//fmt.Println(subject,catn)
		//return nil
	})

}

func openCategoryDB(write bool,hand func(*bolt.Tx)error)error{
	db,err := bolt.Open(Category,0600,nil)
	if err != nil {
		return err
	}
	defer db.Close()
	t,err := db.Begin(write)
	if err != nil {
		return err
	}
	if write {
		defer t.Commit()
	}
	return hand(t)
}

func GetToken(hand func(string)error) error {
	u := url.Values{}
	addSign(&u)
	return request.ClientHttp_("https://www.zaddone.com/wxserver/token?"+u.Encode(),"GET",nil,nil,func(body io.Reader,re int)error{
		db,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		//fmt.P
		return hand(string(db))
	})
}

func GetCategory(na string,b *bolt.Bucket,isg bool,hand func(string,interface{})error)error{
	b_ := b.Bucket([]byte(na))
	if b_ == nil {
		return nil
	}
	val := map[string]interface{}{"name":na}
	err := b_.ForEach(func(k,v []byte)error{
		//fmt.Println(string(k),string(v))
		k_ := string(k)
		if k_ == "id"{
			val[k_] = string(v)
		}else{
			var db interface{}
			if json.Unmarshal(v,&db) == nil {
				val[k_] = db
			}else{
				val[k_] = string(v)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if isg && val["pName"] != nil {
		GetCategory(val["pName"].(string),b,true,hand)
	}
	//fmt.Println(val)
	return hand(na,val)
}

func ShowCategory(name string,hand func(string,interface{})error)error{

	return openCategoryDB(false,func(t *bolt.Tx)error{
		b := t.Bucket(WxCategory)
		if b == nil {
			return nil
		}
		if len(name) > 0 {
			return GetCategory(name,b,false,hand)
		}
		c := b.Cursor()
		for b_,_ := c.First();b_ != nil;b_,_ = c.Next() {
			err := GetCategory(string(b_),b,false,hand)
			//val := map[string]interface{}{}
			//b__:= b.Bucket(b_)
			//err :=  b__.ForEach(func(k,v []byte)error{
			//	val[string(k)] = string(v)
			//	return nil
			//})
			//if err != nil {
			//	return err
			//}
			//err = hand(string(b_),val)
			if err != nil {
				return err
			}
		}
		return nil
	})

}

func GetCategoryInfo(id string,hand func(interface{})error)error{
	return Request(
		"https://api.weixin.qq.com/merchant/category/getproperty",
		map[string]interface{}{"cate_id":id},
		hand)

}
func GetCategorySku(id string,hand func(interface{})error)error{
	//dbMap := map[string]interface{}{"cate_id":id}
	return Request(
		"https://api.weixin.qq.com/merchant/category/getsku",
		map[string]interface{}{"cate_id":id},
		hand)
}
func Request(uri string,dbMap interface{},hand func(interface{})error) error {
	db_ ,err := json.Marshal(dbMap)
	if err != nil {
		return err
	}
	return GetToken(func(token string)error{
	u := url.Values{}
	u.Set("access_token",token)
	//header := http.Header{}
	//header.Add("Content-Type","multipart/form-data")
	//header.Add("Content-Type","application/json")
	//header.Add("Accept","application/json")
	resp,err := http.Post(uri+"?"+u.Encode(),"multipart/form-data",bytes.NewReader(db_))
	if err != nil {
		return err
	}
	var val interface{}
	err = json.NewDecoder(resp.Body).Decode(&val)
	resp.Body.Close()
	if err != nil {
		return err
	}
	return hand(val)
	//return request.ClientHttp_(uri+"?"+u.Encode(),"POST",bytes.NewReader(db_),header,
	//	func(body io.Reader,re int)error{
	//		var val interface{}
	//		err := json.NewDecoder(body).Decode(&val)
	//		if err != nil {
	//			return err
	//		}
	//		return hand(val)
	//	})
	})

}

func DownCategory(id string,hand func(interface{})error)error{

	u := url.Values{}
	return GetToken(func(token string)error{
		u.Set("access_token",token)
		//fmt.Println(u)
		dbMap := map[string]interface{}{
			"cate_id":id,
		}
		db_ ,err := json.Marshal(dbMap)
		if err != nil {
			return err
		}
		uri := "https://api.weixin.qq.com/merchant/category/getsub?"+u.Encode()
		fmt.Println(uri,string(db_))
		return request.ClientHttp_(
			uri,"POST",bytes.NewReader(db_),nil,
		func(body io.Reader,re int)error{
			var val interface{}
			err := json.NewDecoder(body).Decode(&val)
			if err != nil {
				return err
			}
			return hand(val)
			//return nil
		})
	})

}
func sizeImg(w,h float64) (w_ float64,h_ float64){

	s:= w/h
	if w>640 {
		w_ = 640
		h_ = 640/s
		if h_ > 600 {
			h_ = 600
			w_ = h_*s
		}
	}
	return
}

func uploadImg(uri string,hand func(string)) error {
	//hand(uri)
	res,err := http.Get(uri)
	if err != nil {
		return err
	}
	//fmt.Println(res.Header.Get("Content-Type"))
	//fmt.Println(res.Request.Header)
	//fmt.Println("get ",uri)
	img, err := jpeg.Decode(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	w_,h_ := sizeImg(float64(img.Bounds().Dx()),float64(img.Bounds().Dy()))
	m := resize.Resize(uint(w_), uint(h_), img, resize.Lanczos3)
	r_img,w_img := io.Pipe()
	//var wait sync.WaitGroup
	//wait.Add(2)
	go func(){
		err = jpeg.Encode(w_img,m,nil)
		if err != nil {
			fmt.Println(err)
			//panic(err)
		}
		w_img.Close()
		//wait.Done()
	//	fmt.Println("end encode")
	}()

	db,err := ioutil.ReadAll(r_img)
	if err != nil {
		return err
	}
	//go func(){
	return GetToken(func(token string)error{
		us := strings.Split(uri,"/")
		u := url.Values{}
		//us := strings.Split(uri,"/")
		u.Set("filename",us[len(us)-1])
		u.Set("access_token",token)
		resp,err := http.Post("https://api.weixin.qq.com/merchant/common/upload_img?"+u.Encode(),"multipart/form-data",bytes.NewReader(db))
		if err != nil {
			return err
		}
		if resp.StatusCode  != 200 {
			fmt.Println(resp.Request.Header)
			return fmt.Errorf("%d %s",resp.StatusCode,resp.Status)
		}
		var req interface{}
		err = json.NewDecoder(resp.Body).Decode(&req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		img_url := req.(map[string]interface{})["image_url"]
		if img_url == nil {
			return fmt.Errorf("%+v",req)
		}
		//fmt.Println(req)
		hand(img_url.(string))
		return nil


		//return request.ClientHttp_(
		//	"https://api.weixin.qq.com/merchant/common/upload_img?"+u.Encode(),
		//	//"https://api.weixin.qq.com/cgi-bin/media/uploadimg?"+u.Encode(),
		//	"POST",
		//	r_img,
		//	header,
		//	//http.Header{"Content-Type":[]string{"multipart/form-data"}},
		//	//http.Header{"Content-Type":[]string{"application/x-www-form-urlencoded"}},
		//	func(body io.Reader,re int)error{
		//	//fmt.Println(re)
		//	if re != 200 {
		//		db,err := ioutil.ReadAll(body)
		//		if err != nil {
		//			return err
		//		}
		//		return fmt.Errorf("err:%s %d",string(db),re)
		//	}
		//	var req interface{}
		//	json.NewDecoder(body).Decode(&req)
		//	img_url := req.(map[string]interface{})["image_url"]
		//	if img_url == nil {
		//		return fmt.Errorf("%+v",req)
		//	}
		//	fmt.Println(req)
		//	hand(img_url.(string))
		//	return nil
		//})
	})
	//if err != nil {
	//	fmt.Println(err)
	//	//panic(err)
	//}
	////wait.Done()
	//
	//wait.Wait()
	//return nil

}

func handCategory(pidName string,db interface{},pb,b *bolt.Bucket)error {

	db_ := db.(map[string]interface{})["cate_list"]
	if db_ == nil {
		return nil
	}
	var li []string
	for _,d := range db_.([]interface{}){
		d_ := d.(map[string]interface{})
		//go func(d_ map[string]interface{}){
		//fmt.Println(d_)
		name := d_["name"].(string)
		li = append(li,name)
		id := d_["id"].(string)
		//b.CreateBucketIfNotExists(
		b_,err := b.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			fmt.Println(err)
			return err
		}
		if len(pidName)>0{
			err := b_.Put([]byte("pName"),[]byte(pidName))
			if err != nil {
				return err
				//return
			}
		}
		err = b_.Put([]byte("id"),[]byte(id))
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = DownCategory(id,func(__db interface{})error{
			return handCategory(name,__db,b_,b)
		})
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = GetCategorySku(id,func(vb interface{})error{
			//fmt.Println(vb)
			sku := vb.(map[string]interface{})["sku_table"]
			if sku != nil {
				db,err := json.Marshal(sku)
				if err != nil {
					return err
				}
				return b_.Put([]byte("sku"),db)
			}
			return nil
		})
		if err != nil {
			return err
		}
		err = GetCategoryInfo(id,func(vb interface{})error{
			//fmt.Println(vb)
			sku := vb.(map[string]interface{})["properties"]
			if sku != nil {
				db,err := json.Marshal(sku)
				if err != nil {
					return err
				}
				return b_.Put([]byte("properties"),db)
			}
			return nil
		})
		if err != nil {
			return err
		}

	}
	if pb!=nil && len(li)>0{
		pb.Put([]byte("children"),[]byte(strings.Join(li,",")))
	}
	//pb := b.Bucket([]byte(pidName))
	//c.JSON(http.StatusOK,db)
	return nil
}

func init(){
	shop := Router.Group("wxshop")
	shop.GET("category/show",func(c *gin.Context){
		var li []interface{}
		err := ShowCategory(c.Query("key"),func(name string,db interface{})error{
			db_ := db.(map[string]interface{})
			if db_["children"] == nil {
				li = append(li,map[string]interface{}{"name":name,"val":db})
			}
			//https://api.weixin.qq.com/merchant/category/getsku?access_token=ACCESS_TOKEN
			return nil
		})
		if err != nil {
			c.JSON(http.StatusFound,err)
			return
		}

		err = openCategoryDB(true,func(t *bolt.Tx)error{
			b,err := t.CreateBucketIfNotExists(WxWorks)
			if err != nil {
				return err
			}
			for _,l := range li {
				name :=[]byte( l.(map[string]interface{})["name"].(string))
				for _,s := range bytes.Split(name,[]byte{'/'}){
					v := b.Get(s)
					if v != nil {
						fmt.Println(string(s),string(v))
						continue
					}
					err := b.Put(s,name)
					if err != nil {
						fmt.Println(err,string(s),string(name))
						//return err
					}
				}
			}
			return nil
		})
		if err != nil {
			//panic(err)
			c.JSON(http.StatusFound,err)
			return
		}

		c.JSON(http.StatusOK,li)
		return
	})
	shop.GET("category/down",func(c *gin.Context){
		err := openCategoryDB(true,func(t *bolt.Tx)error{
			rootCat:="1"
			b,err := t.CreateBucketIfNotExists(WxCategory)
			if err != nil {
				return err
			}
			return DownCategory(rootCat,func(db interface{})error{
				return handCategory("",db,nil,b)
			})
		})
		if err != nil {
			c.JSON(http.StatusFound,err)
		}
		c.String(http.StatusOK,"success")
	})

}
