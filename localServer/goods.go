package main
import(
	"fmt"
	"bytes"
	"io"
	"net/http"
	"net/url"
	//"strings"
	//"strconv"
	"github.com/gin-gonic/gin"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/alibaba"
	"github.com/zaddone/studySystem/shopping"
	"github.com/PuerkitoBio/goquery"
	"encoding/json"
	"io/ioutil"
	"strings"
)
func getDesImg(des string,hand func(string))error{
	doc,err := goquery.NewDocumentFromReader(strings.NewReader(des))
	if err != nil {
		return err
	}
	doc.Find("img").Each(func(i int,s *goquery.Selection){
		v,_ := s.Attr("src")
		//fmt.Println(v,e)
		hand(v)
	})
	return nil
}
func PreviewOrder(obj map[string]interface{},ali *shopping.Alibaba)error{
	pro := obj["productInfo"].(map[string]interface{})
	skus := pro["skuInfos"]
	if skus == nil {
		return fmt.Errorf("obj is nil")
	}
	//var sku map[string]interface{}
	var ss []interface{}
	for _,s := range skus.([]interface{}){
		if s.(map[string]interface{})["amountOnSale"].(float64) >1 {
			ss = append(ss,s)
		}
	}
	pro["skuInfos"] = ss
	po := &shopping.AlProductForOrder{
		Offerid:pro["productID"].(float64),
		SpecId:ss[0].(map[string]interface{})["specId"].(string),
		Quantity:1,
	}
	addr := &shopping.AlAddrForOrder{}
	addr.LoadTestDB()
	for {
		res := ali.PreviewCreateOrder(addr,[]*shopping.AlProductForOrder{po})
		switch r := res.(type){
		case error:
			return r
		default:
			errcode := r.(map[string]interface{})["errorCode"]
			if errcode == nil{
				obj["Preview"] = r
				return nil
			}
			ec := errcode.(string)
			if ec == "500_005" || ec == "500_006" {
				fmt.Println(r)
				po.Quantity++
			}else{
				return fmt.Errorf("%s",errcode)
			}
		}
	}
	return nil



}
func GoodsListHand(li []string)(dbs []interface{}){

	lis := strings.Join(li,",")
	u := url.Values{}
	u.Add("goodsids",lis)
	addSign(&u)
	//var li []interface{}
	err := request.ClientHttp_("https://www.zaddone.com/site/v2/goods/update/list?"+u.Encode(),"GET",nil,nil,func(body io.Reader,re int)error{
		if re != 200 {
			return fmt.Errorf("start is %d",re)
		}
		return json.NewDecoder(body).Decode(&dbs)
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return

}
func handGoods(obj map[string]interface{})interface{}{
	pro := obj["productInfo"].(map[string]interface{})
	if pro["skuInfos"] == nil {
		return nil
	}
	op_ :=obj["Preview"].(map[string]interface{})["orderPreviewResuslt"]
	if op_ == nil {
		return nil
	}
	op_list := op_.([]interface{})
	if len(op_list)==0 {
		return nil
	}
	//pro["subject_old"] = pro["subject"]
	resu := op_list[0].(map[string]interface{})
	pro["price"] =fmt.Sprintf("%.2f", resu["sumPayment"].(float64)*1.1/100)
	cl := resu["cargoList"].([]interface{})[0].(map[string]interface{})
	pro["NumMin"] = cl["amount"].(float64)/cl["finalUnitPrice"].(float64)
	if pro["NumMin"].(float64)<1 {
		pro["NumMin"] = 1
	}
	Carriage := resu["sumCarriage"].(float64)/100
	attrName := []string{}
	for i,_v := range pro["skuInfos"].([]interface{}){
		v := _v.(map[string]interface{})
		skuName := ""
		for _,v_ := range v["attributes"].([]interface{}){
			_v_ := v_.(map[string]interface{})
			if _v_["skuImageUrl"]!= nil {
				v["imageUrl"] = "https://cbu01.alicdn.com/"+_v_["skuImageUrl"].(string)
			}
			skuName += _v_["attributeValue"].(string)
			if i == 0 {
				attrName =append(attrName, _v_["attributeDisplayName"].(string))
			}
		}
		v["skuName"] = skuName
		if v["price"]== nil {
			v["price"] = pro["price"]
		}else{
			v["price"] =fmt.Sprintf("%.2f",(v["price"].(float64)+Carriage)*1.1)
		}
	}
	pro["attrName"] = strings.Join(attrName,"/")

	images := pro["image"].(map[string]interface{})["images"].([]interface{})
	for i,image := range images {
		images[i] = "https://cbu01.alicdn.com/"+image.(string)
	}
	return pro
}


func UpdateGoods(id string,obj interface{})error{
	if obj == nil {
		return fmt.Errorf("obj is nil")
	}
	u := url.Values{}
	u.Add("id",id)
	addSign(&u)
	db,err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return request.ClientHttp_("https://www.zaddone.com/site/v2/goods/update?"+u.Encode(),"POST",bytes.NewReader(db),nil,func(body io.Reader,re int)error{
		if re != 200 {
			d,_ := ioutil.ReadAll(body)
			return fmt.Errorf("%d %s",re,d)
		}
		return nil
	})
}

func initAlibaba(hand func(*shopping.Alibaba)error)error{
	Info := &shopping.ShoppingInfo{}
	err  := requestHttp("/shopping/1688","GET",nil,nil,func(body io.Reader,res *http.Response)error{
		return json.NewDecoder(body).Decode(Info)
	})
	if err != nil {
		return err
	}
	return hand(shopping.NewAlibaba(Info,""))
}

func init(){
	goods := Router.Group("goods")
	goods.GET("/",func(c *gin.Context){
		c.HTML(http.StatusOK,"goods.tmpl",nil)
	})
	goods.POST("/save",func(c *gin.Context){
		id := c.Query("id")
		if id == "" {
			c.JSON(http.StatusFound,fmt.Errorf("id = nil"))
			return
		}
		u := url.Values{}
		u.Set("id",id)
		addSign(&u)
		err := request.ClientHttp_("https://www.zaddone.com/site/v2/goods/update/edit?"+u.Encode(),"POST",c.Request.Body,nil,func(body io.Reader,re int)error{

			db,err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			c.String(re,string(db))
			return nil
		})
		if err != nil {
			c.JSON(http.StatusFound,err)
		}


	})
	goods.GET("/show",func(c *gin.Context){
		u := url.Values{}
		u.Add("goodsid",c.Query("id"))
		u.Add("show",c.Query("show"))
		u.Add("con","8")
		addSign(&u)
		err := request.ClientHttp_("https://www.zaddone.com/site/v2/goods/update/list_t?"+u.Encode(),"GET",nil,nil,func(body io.Reader,re int)error{
			if re != 200 {
				return fmt.Errorf("%d",re)
			}
			var db interface{}
			err := json.NewDecoder(body).Decode(&db)
			if err != nil {
				return err
			}
			c.JSON(http.StatusOK,db)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,err)
			return
		}

	})
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

	goods.GET("/down",func(c *gin.Context){
		//var li []interface{}
		list := []string{}
		err := initAlibaba(func(ali *shopping.Alibaba)error {
			alibaba.HandGoods = func(db interface{}){
				//li = append(li,db)
				db_:= db.(map[string]interface{})
				productId :=fmt.Sprintf("%.0f",db_["productId"].(float64))
				itemId := fmt.Sprintf("%.0f",db_["itemId"].(float64))
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
				err := getDesImg(des,func(img string){
					//fmt.Println(img)
					des_img = append(des_img,img)
				})
				if err != nil {
					panic(err)
				}
				product["des_img"] = des_img

				err = PreviewOrder(obj_,ali)
				if err != nil {
					fmt.Println(err)
					return
				}
				pro:= handGoods(obj_)
				if pro == nil {
					return
				}
				err = UpdateGoods(productId,pro)
				if err != nil {
					fmt.Println(err)
					return
				}

				err = ali.Crossborder(productId)
				if err != nil {
					//return err
					panic(err)
				}
				//li = append(li,pro)

				list = append(list,productId)
			}
			return alibaba.Run()
		})
		if len(list)>0{

			c.JSON(http.StatusOK,GoodsListHand(list))
			return
		}
		c.JSON(http.StatusOK,err)
		return
	})

}
