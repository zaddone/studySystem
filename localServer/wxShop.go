package main
import(
	"fmt"
	"io"
	"net/url"
	"github.com/gin-gonic/gin"
	"github.com/zaddone/studySystem/request"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"bytes"
)
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

func DownCategory(id int,hand func(interface{})error)error{

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

func init(){
	shop := Router.Group("wxshop")
	shop.GET("category/update",func(c *gin.Context){
		err := DownCategory(1,func(db interface{})error{
			//fmt.Println(db)
			c.JSON(http.StatusOK,db)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusFound,err)
		}
	})
}
