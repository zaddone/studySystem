package main
import(
	"sort"
	"github.com/zaddone/studySystem/shopping"
	"strings"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	//"crypto/hmac"
	//"encoding/hex"
	//"crypto/sha1"
	"net/http"
	"strconv"
	//"io/ioutil"
	//"net/url"
	"fmt"
	"encoding/xml"
	"time"
	"regexp"
	"io"
)
var(
	WXtoken = "zhaoweijie2020"
	httpReg = regexp.MustCompile(`http`)
	//jdNameReg = regexp.MustCompile(`(jd)|(京东)`)
	//pddNameReg = regexp.MustCompile(`(pdd)|(pinduoduo)|(拼多多)`)
	//jdReg = regexp.MustCompile(`\/(\d+)\.html`)
	//jdReg_ = regexp.MustCompile(`sku=(\d+)`)
	//jdOrderReg = regexp.MustCompile(`\d{12}`)

	//pddReg = regexp.MustCompile(`goods_id=(\d+)`);
	//pddOrderReg = regexp.MustCompile(`\d{6}-\d{15}`)
	phoneReg = regexp.MustCompile(`1\d{10}`)
	typeReg = regexp.MustCompile(`(微信)|(支付宝)`)
	//cmdReg = regexp.MustCompile(`(\d+)([a-zA-Z|\p{Han}]+)`)
	msg = "查优惠卷返利 https://www.zaddone.com"
	OrderErrMsg = "没有找到订单!\n请核对订单号或者稍候再试"
	phoneMsg = "请确认开通了手机转账功能\n支付宝/微信？手机号？\n发给我!"
	UserDB *bolt.DB
	msgPhone = []byte("phone")
	msgType = []byte("type")
	welcome = "zaddone_com米果推荐\n1、支持输入（淘宝、京东、拼多多）网购商品链接，查询产品价格和返利下单链接\n 2、输入订单号、手机号和（微信|支付宝）设置系统自动转账信息，定时到帐。3、发送其他信息，可获取账户金额等相关信息"

	jdReg = regexp.MustCompile(`\/(\d+)\.html`)
	jdReg_ = regexp.MustCompile(`sku=(\d+)`)
	jdOrderReg = regexp.MustCompile(`\d{12,}`)
	pddReg = regexp.MustCompile(`goods_id=(\d+)`);
	pddOrderReg = regexp.MustCompile(`\d{6}-\d{15}`)

)
func init(){
	var err error
	UserDB,err = bolt.Open("UserDB",0600,nil)
	if err != nil {
		panic(err)
	}
	Router.POST("/wx",handWxQuery)
	Router.GET("/wx",handWxQuery)
}
func GetUserMsg(userid string,h func(string))error{
	return UserDB.View(func(t *bolt.Tx)error{
		b := t.Bucket([]byte(userid))
		if b == nil {
			h(phoneMsg)
			return nil
		}
		phone := b.Get(msgPhone)
		if phone == nil {
			h(phoneMsg)
			return nil
		}
		ty := b.Get(msgType)
		h(fmt.Sprintf("请核对手机转账信息:\n%s %s\n如果错了,请重新发给我!",ty,phone))
		return nil
	})
}
func SaveUserMsg(userid,msg string)error{
	userid_ := []byte(userid)
	var key,val [][]byte
	//buck :=[]byte("msg")
	phone := phoneReg.FindString(msg)
	if len(phone)>0{
		key=append(key,msgPhone)
		val=append(val, []byte(phone))
	}
	ty := typeReg.FindString(msg)
	if len(ty)>0{
		key= append(key,msgType)
		val= append(val,[]byte(ty))
	}
	if len(key) == 0 {
		return io.EOF
	}
	return UserDB.Batch(func(t *bolt.Tx)error{
		b,err := t.CreateBucketIfNotExists(userid_)
		if err != nil {
			return err
		}
		for i,k := range key {
			err = b.Put(k,val[i])
			if err != nil {
				return err
			}
		}
		return nil

	})
}

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
	Event string
	MsgId int
}

func handMsg(str,userid string,hand func(string)){
	n := httpReg.FindStringIndex(str)
	if len(n)>0{
		handHttp(str[n[0]:],userid,hand)
		return
	}
	oid := jdOrderReg.FindString(str)
	if len(oid)==12 {
		obj_ ,_ := shopping.ShoppingMap.Load("jd")
		obj := obj_.(shopping.ShoppingInterface)
		db := obj.OrderSearch(oid,userid)
		if db == nil {
			hand(OrderErrMsg)
			return
		}
		msg := obj.OrderMsg(db)
		err := GetUserMsg(userid,func(s string){
			msg += "\n"+s
		})
		if err != nil {
			panic(err)
		}
		hand(msg)
		return
	}
	if len(oid) == 18 {
		hand("taobao order:"+oid)
		return
	}
	oid = pddOrderReg.FindString(str)
	if len(oid)>0 {
	//if pddOrderReg.MatchString(str){
		obj_,_ := shopping.ShoppingMap.Load("pinduoduo")
		obj := obj_.(shopping.ShoppingInterface)
		db := obj.OrderSearch(oid,userid)
		if db == nil {
			hand(OrderErrMsg)
			return
		}
		msg := obj.OrderMsg(db)
		err := GetUserMsg(userid,func(s string){
			msg += "\n"+s
		})
		if err != nil {
			panic(err)
		}
		hand(msg)
		return
	}
	err := SaveUserMsg(userid,str)
	if err == io.EOF{
		hand(str)
		return
	}
	var msg string
	GetUserMsg(userid,func(s string){
		msg = s
	})
	if len(msg) >0  {
		hand(msg)
		return
	}
	hand(str)
}



func jdRev(id,uesrid string)string{
	obj_,_ := shopping.ShoppingMap.Load("jd")
	obj := obj_.(shopping.ShoppingInterface)
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
	msg := fmt.Sprintf("%s\nhttps://www.zaddone.com/p/jd/%s\n￥%.2f-%.2f\n技术服务费 %.2f\n预计返利 %.2f\n完成订单后一定复制粘贴订单号给我!",
		detail["wareName"].(string),
		id,
		price,
		pricef,
		pricef*0.1,
		pricef*0.9,
	)
	GetUserMsg(uesrid,func(s string){
		msg+="\n"+s
	})
	return msg
}
func pddRev(id,uesrid string)string{
	obj_,_ := shopping.ShoppingMap.Load("pinduoduo")
	obj := obj_.(shopping.ShoppingInterface)
	info:=obj.GoodsDetail(id)
	if info == nil {
		return "msg error"
	}
	detail := info.(map[string]interface{})["goods_detail_response"].(map[string]interface{})["goods_details"].([]interface{})[0].(map[string]interface{})

	//out := obj.GoodsUrl(id)
	//if out == nil {
	//	return "msg error"
	//}
	//ur := "https://www.zaddone.com/?keyword=https%3A%2F%2Fitem.jd.com%2F"+id+".html"
	//out.(map[string]interface{})["goods_promotion_url_generate_response"].(map[string]interface{})["goods_promotion_url_list"].([]interface{})[0].(map[string]interface{})["short_url"].(string),
	price := detail["min_group_price"].(float64)/100
	pricef := price*(detail["promotion_rate"].(float64)/1000)
	//uri := "https://www.zaddone.com/p/pinduoduo/"+id
	msg :=  fmt.Sprintf("%s\n%s\nhttps://www.zaddone.com/p/pinduoduo/%s\n￥%.2f-%.2f\n技术服务费%.2f\n预计返利 %.2f\n完成订单后一定复制粘贴订单号给我!",
		detail["mall_name"].(string),
		detail["goods_name"].(string),
		id,
		price,
		pricef,
		pricef*0.1,
		pricef*0.9,
	)

	GetUserMsg(uesrid,func(s string){
		msg+="\n"+s
	})
	return msg
}
func handHttp(str,userid string,h func(string)) {
	ss := pddReg.FindStringSubmatch(str)
	if len(ss) >1 {
		h(pddRev(ss[1],userid))
		return
	}
	ss = jdReg.FindStringSubmatch(str)
	if len(ss) >1 {
		h(jdRev(ss[1],userid))
		return
	}
	ss = jdReg_.FindStringSubmatch(str)
	if len(ss) >1 {
		h(jdRev(ss[1],userid))
		return
	}
	h(str)
}
func sendMsg(c *gin.Context,db *wxRevMsg,content string) {
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
	c.String(http.StatusOK,string(sendstr))
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
	if db.Event != ""{
		if db.Event == "subscribe" {
			sendMsg(c,&db,welcome)
		}
		return
	}
	handMsg(db.Content,db.FromUserName,func(s string){
		content = s
	})

	sendMsg(c,&db,content)
	return
}
func handWxQuery(c *gin.Context){
	timestamp:=c.Query("timestamp")
	if timestamp == ""{
		c.String(http.StatusOK,"")
		return
	}
	stamp,err := strconv.ParseInt(timestamp,10,64)
	if err != nil {
		c.String(http.StatusOK,"")
		return
	}
	d := time.Now().Unix() - stamp
	if d<0 {
		d=-d
	}
	if d>6 {
		c.String(http.StatusOK,"")
		return
	}
	signature := c.Query("signature")
	nonce := c.Query("nonce")
	//echostr:= c.Query("echostr")
	li := []string{WXtoken,timestamp,nonce}
	sort.Strings(li)
	li_ := shopping.Sha1([]byte(strings.Join(li,"")))
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
