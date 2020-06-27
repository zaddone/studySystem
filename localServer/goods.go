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
	"encoding/json"
	"io/ioutil"
)

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
				catid := obj_["productInfo"].(map[string]interface{})["categoryID"].(float64)
				obj_["cat"] = ali.GetCategory(fmt.Sprintf("%.0f",catid))
				err := UpdateGoods(productId,obj)
				//err := ali.SaveProduct(productId,obj)
				if err != nil {
					//return err
					//return err
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
