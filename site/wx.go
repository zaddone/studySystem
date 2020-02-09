package main
import(
	"sort"
	"strings"
	"github.com/gin-gonic/gin"
	//"crypto/hmac"
	"encoding/hex"
	"crypto/sha1"
	"net/http"
	"strconv"
	//"io/ioutil"
	"fmt"
	"encoding/xml"
	"time"
	"regexp"
)
var(
	WXtoken = "zhaoweijie2020"
	OrderIDReg = regexp.MustCompile(`(jd|京东)[\s|\S]*(\d{12})`)
	httpReg = regexp.MustCompile(`http`)
	jdReg = regexp.MustCompile(`\/(\d+)\.html`)
	jdReg_ = regexp.MustCompile(`sku=(\d+)`)
	pddReg = regexp.MustCompile(`goods_id=(\d+)`);
	cmdReg = regexp.MustCompile(`([a-zA-Z|\p{Han}]+)(\d+)`)
	msg = "查优惠卷返利 https://www.zaddone.com"


)
type wxMsg struct{
	XMLName  xml.Name `xml:"xml"`
	ToUserName string
	FromUserName string
	CreateTime int64
	MsgType string
	Content string
}
type wxRevMsg struct{
	XMLName  xml.Name `xml:"xml"`
	ToUserName string
	FromUserName string
	CreateTime int64
	MsgType string
	Content string
	MsgId int
}
func Sha1(data []byte) string {
	sha1 := sha1.New()
	sha1.Write(data)
	return hex.EncodeToString(sha1.Sum([]byte(nil)))
}
func handMsg(str string,hand func(string)){
	n := httpReg.FindStringIndex(str)
	if len(n)>0{
		handHttp(str[n[0]:],hand)
		return
	}
	s := cmdReg.FindStringSubmatch(str)
	fmt.Println(s)
	if len(s)==3 {
		handCmd(s[1],s[2],hand)
		return
	}
	hand(str)
}
func handCmd(name,num string,h func(string)){
	h(name+":"+num)
}
func jdRev(id string)string{
	obj := ShoppingMap["jd"]
	info:=obj.GoodsDetail(id)
	if info == nil {
		return msg
	}
	//fmt.Println(info)
	result := info.(map[string]interface{})["jd_kpl_open_xuanpin_searchgoods_response"].(map[string]interface{})["result"]
	if result == nil{
		return msg
	}
	detail := result.(map[string]interface{})["queryVo"].([]interface{})[0].(map[string]interface{})
	//out := obj.GoodsUrl(id)
	//if out == nil {
	//	return msg
	//}
	//fmt.Println(out)
	price,err :=strconv.ParseFloat(detail["price"].(string),64)
	if err != nil {
		fmt.Println(err)
		return msg
	}
	ratio,err := strconv.ParseFloat(detail["commisionRatioWl"].(string),64)
	if err != nil {
		fmt.Println(err)
		return msg
	}
	pricef := price*(ratio/100)*0.9
	//
	ur := "https://www.zaddone.com/?keyword=https%3A%2F%2Fitem.jd.com%2F"+id+".html"
	return fmt.Sprintf("%s\n%s\n%.2f-%.2f\nRebate %.2f\nfee %.2f",
		detail["wareName"].(string),
		ur,
		price,
		pricef,
		pricef*0.9,
		pricef*0.1,
	)
}
func pddRev(id string)string{
	obj := ShoppingMap["pinduoduo"]
	info:=obj.GoodsDetail(id)
	if info == nil {
		return "msg error"
	}
	detail := info.(map[string]interface{})["goods_detail_response"].(map[string]interface{})["goods_details"].([]interface{})[0].(map[string]interface{})

	out := obj.GoodsUrl(id)
	if out == nil {
		return "msg error"
	}
	//ur := "https://www.zaddone.com/?keyword=https%3A%2F%2Fitem.jd.com%2F"+id+".html"
	price := detail["min_group_price"].(float64)/100
	pricef := price*(detail["promotion_rate"].(float64)/1000)
	return fmt.Sprintf("%s\n%s\n%s\n%.2f-%.2f\nRebate %.2f\nfee %.2f",
		detail["mall_name"].(string),
		detail["goods_name"].(string),
		out.(map[string]interface{})["goods_promotion_url_generate_response"].(map[string]interface{})["goods_promotion_url_list"].([]interface{})[0].(map[string]interface{})["short_url"].(string),
		price,
		pricef,
		pricef*0.9,
		pricef*0.1,
	)
}
func handHttp(str string,h func(string)) {
	ss := pddReg.FindStringSubmatch(str)
	if len(ss) >1 {
		h(pddRev(ss[1]))
		return
	}
	ss = jdReg.FindStringSubmatch(str)
	if len(ss) >1 {
		h(jdRev(ss[1]))
		return
	}
	ss = jdReg_.FindStringSubmatch(str)
	if len(ss) >1 {
		h(jdRev(ss[1]))
		return
	}
	h(str)
}

func handWxPost(c *gin.Context){
	var db wxRevMsg
	err := xml.NewDecoder(c.Request.Body).Decode(&db)
	if err != nil {
		fmt.Println(err)
		c.String(http.StatusOK,"")
		return
	}
	content := "success"
	handMsg(db.Content,func(s string){
		content = s
	})
	sendstr,err := xml.Marshal(&wxMsg{
		ToUserName:db.FromUserName,
		FromUserName:db.ToUserName,
		CreateTime:time.Now().Unix(),
		MsgType:"text",
		Content:content})
	if err != nil {
		fmt.Println(err)
		c.String(http.StatusOK,"")
		return
	}
	//fmt.Println(string(sendstr))
	c.String(http.StatusOK,string(sendstr))
	return
}
func handWxQuery(c *gin.Context){
	timestamp:=c.Query("timestamp")
	if timestamp == ""{
		c.String(http.StatusOK,"")
		return
	}
	signature := c.Query("signature")
	nonce := c.Query("nonce")
	//echostr:= c.Query("echostr")
	li := []string{WXtoken,timestamp,nonce}
	sort.Strings(li)
	li_ := Sha1([]byte(strings.Join(li,"")))
	if signature != li_ {
		c.String(http.StatusOK,"")
		return
	}

	//c.String(http.StatusOK,"")
	//return
	handWxPost(c)
	//c.String(http.StatusOK,echostr)

	//c.JSON(http.StatusOK,gin.H{"msg":timestamp})
}
