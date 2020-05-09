package main
import(
	"fmt"
	"io"
	"net/http"
	"crypto/md5"
	"bytes"
	"time"
	"sort"
	"encoding/json"
	"github.com/zaddone/studySystem/request"
	"github.com/gin-gonic/gin"
)
var(
	//uri = "https://openapi.fulu.com/api/getway"
	uri = "https://pre-openapi.fulu.com/api/getway"
	//appSecret = "1bff4e1a3ce545598c582615f6164195"
	//appkey = "pXdSGGz15gDX5vifEWPvHRNFfHWnbnLpIT908akIyQumbE0YY+vwLGxOCLhJPufm"
	appkey = "i4esv1l+76l/7NQCL3QudG90Fq+YgVfFGJAWgT+7qO1Bm9o/adG/1iwO2qXsAXNB"
	appSecret = "0a091b3aa4324435aab703142518a8f7"
	timeFormat = "2006-01-02 15:04:05"
	Router = gin.Default()
)

type sortRunes []rune

func (s sortRunes) Less(i, j int) bool {
    return s[i] < s[j]
}

func (s sortRunes) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}

func (s sortRunes) Len() int {
    return len(s)
}

func SortString(s string) string {
    r := []rune(s)
    sort.Sort(sortRunes(r))
    return string(r)
}


func addSign(u map[string]interface{} ){
	u["app_key"]=appkey
	u["timestamp"]=time.Now().Format(timeFormat)
	u["format"]="json"
	u["charset"]="utf-8"
	u["sign_type"]="md5"
	u["app_auth_token"]=""
	u["version"]="2.0"
	db,err := json.Marshal(u)
	if err != nil {
		panic(err)
	}
	sign := SortString(string(db)) + appSecret
	u["sign"]=fmt.Sprintf("%x", md5.Sum([]byte(sign)))
	//fmt.Println(u)

}
func ClientHttp(body map[string]interface{})( out interface{}){
	ht := http.Header{}
	ht.Add("Content-Type","application/json")
	db,err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	err = request.ClientHttp_(
		uri,
		"POST",
		bytes.NewBuffer(db),
		ht,
		func(b io.Reader,s int)error{
			fmt.Println(s)
			return json.NewDecoder(b).Decode(&out)
		},
	)
	if err != nil {
		fmt.Println(err,out)
		out = err
	}
	return out

}
func OrderMobileAdd(phone string,num int,orderid string) interface{}{
	u := map[string]interface{}{}
	u["method"]="fulu.order.mobile.add"
	u["biz_content"] = map[string]interface{}{
		"charge_phone":phone,
		"charge_value":num,
		"customer_order_no":orderid
	}
	addSign(u)
	req := ClientHttp(u)
	if req == nil {
		return io.EOF
	}
	switch r:= req.(type){
	case error:
		return r
	}
	req_ := req.(map[string]interface{})
	if req_ == nil {
		return io.EOF
	}
	code := int(req_["code"].(float64))
	if code != 0{
		return fmt.Errorf("%d %s",int(code),req_["message"].(string))
	}
	return req_["result"]

}

func getGoodsTemplate(id string) interface{} {
	u := map[string]interface{}{}
	u["method"]="fulu.goods.template.get"
	u["biz_content"] = map[string]interface{}{
		"template_id":id,
	}
	addSign(u)
	req := ClientHttp(u)
	//fmt.Println(req)
	if req == nil {
		return io.EOF
	}
	switch r:= req.(type){
	case error:
		return r
	}
	req_ := req.(map[string]interface{})
	if req_ == nil {
		return io.EOF
	}
	code := int(req_["code"].(float64))
	if code != 0{
		return fmt.Errorf("%d %s",int(code),req_["message"].(string))
	}
	fmt.Println(req_["result"])
	tem := &GoodsTemplate{}
	json.Unmarshal([]byte(req_["result"].(string)),tem)
	return tem

}
func getGoodsList(hand func(*GoodsInfo)error) error{
	u := map[string]interface{}{}
	u["method"]="fulu.goods.list.get"
	u["biz_content"]="{}"
	addSign(u)
	req := ClientHttp(u)
	if req == nil {
		return io.EOF
	}
	switch r:= req.(type){
	case error:
		return r
	}
	req_ := req.(map[string]interface{})
	if req_ == nil {
		return io.EOF
	}
	code := int(req_["code"].(float64))
	if code != 0{
		return fmt.Errorf("%d %s",int(code),req_["message"].(string))
	}
	var res []*GoodsInfo
	err := json.Unmarshal([]byte(req_["result"].(string)),&res)
	if err != nil {
		return err
	}
	for _,v := range res {
		//tem := getGoodsTemplate(v.Template_id)
		//switch te := tem.(type){
		//case *GoodsTemplate:
		//	v.Template = te
		//case error:
		//	fmt.Println(tem)
		//}
		err := hand(v)
		if err != nil{
			return err
		}
	}
	return nil
	//fmt.Println(out)
}
func init(){
	Router.POST("/orderback",func(c *gin.Context){
		c.Request.Body
	})
	Router.GET("/goodslist",func(c *gin.Context){
		var li []*GoodsInfo
		err := getGoodsList(func(g *GoodsInfo)error{
			li = append(li,g)
			return nil
		})
		if err != nil {
			c.JSONP(http.StatusOK,gin.H{"msg":err})
			return
		}
		c.JSONP(http.StatusOK,gin.H{"msg":"success","result":li})
	})
}
func main(){
	Router.Run(":8083")
	select{}
}
