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


func UpdateGoods(id string,obj interface{})error{
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
	goods.GET("/order",func(c *gin.Context){
		err := initAlibaba(func(ali *shopping.Alibaba)error {
			o := new(shopping.AlAddrForOrder)
			p := new(shopping.AlProductForOrder)
			o.LoadTestDB()
			p.LoadTestDB()
			obj := ali.CreateOrder(o,[]*shopping.AlProductForOrder{p})
			c.JSON(http.StatusOK,obj)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusOK,err)
			return
		}
	})

	goods.GET("/down",func(c *gin.Context){
		var li []interface{}
		err := initAlibaba(func(ali *shopping.Alibaba)error {
			//err := ali.ClearProduct()
			//if err != nil {
			//	return err
			//}
			alibaba.HandGoods = func(db interface{}){
				//li = append(li,db)
				db_:= db.(map[string]interface{})
				productId :=fmt.Sprintf("%.0f",db_["productId"].(float64))
				itemId := fmt.Sprintf("%.0f",db_["itemId"].(float64))
				obj := ali.GoodsDetail(productId)
				obj_ := obj.(map[string]interface{})
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
				err = UpdateGoods(productId,obj)
				if err != nil {
					panic(err)
				}

				err = ali.Crossborder(productId)
				if err != nil {
					//return err
					panic(err)
				}
				li = append(li,obj)
			}
			return alibaba.Run()
		})
		if len(li)>0{
			c.JSON(http.StatusOK,li)
			return
		}
		c.JSON(http.StatusOK,err)
		return
	})

}
