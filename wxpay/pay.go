package main
import(
	"fmt"
	"encoding/json"
	"github.com/zaddone/studySystem/shopping"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/config"
	"github.com/gin-gonic/gin"
	"time"
	"strings"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
)
var(
	Remote = flag.String("r", "http://127.0.0.1:8080/v2","remote")
)

func requestHttp(path,Method string,u url.Values, body io.Reader,hand func(io.Reader,*http.Response)error)error{
	if u == nil {
		u = url.Values{}
	}
	addSign(&u)
	return request.ClientHttp__(*Remote+path+"?"+u.Encode(),Method,body,nil,hand)
}

func addSign(u *url.Values){
	u.Add("timestamp",fmt.Sprintf("%d",time.Now().Unix()))
	li := []string{config.Conf.Minitoken}
	for _,v := range *u{
		li = append(li,v...)
	}
	sort.Strings(li)
	u.Add("sign",shopping.Sha1([]byte(strings.Join(li,""))))
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

func ShopPay(o *shopping.AlAddrForOrder,p *shopping.AlProductForOrder,hand func(interface{})error) error {
	return initAlibaba(func(ali *shopping.Alibaba)error {
		return hand(ali.CreateOrder(o,[]*shopping.AlProductForOrder{p}))
	})
}
type OrderInfo struct {
	Goods shopping.AlProductForOrder    `json:"Goods"`
	Addr  shopping.AlAddrForOrder `json:"Addr"`
}
func init(){
	pay := Router.Group("pay",func() gin.HandlerFunc {
		return checkManage
	}())
	pay.POST("/postordertoalibaba",func(c *gin.Context){
		db,err := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		fmt.Println(string(db))
		var o OrderInfo
		_db,_ := json.Marshal(&o)
		fmt.Println(string(_db))

		err = json.Unmarshal(db,&o)
		fmt.Printf("%+v\n",o)
		//err := json.NewDecoder(c.Request.Body).Decode(&o)
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
			return
		}
		//fmt.Printf("%v,%v\n",o.Goods,o.Addr)
		err = ShopPay(&(o.Addr),&(o.Goods),func(db interface{})error{
			c.JSON(http.StatusOK,gin.H{"msg":db})
			return nil
		})
		if err != nil {
			c.JSON(http.StatusNotFound,gin.H{"msg":err})
		}
		return
	})
}
