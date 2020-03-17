package alimama
import(
	"fmt"
	"encoding/base64"
	"encoding/json"
	"os"
	"time"
	//"strings"
	"strconv"
	"github.com/zaddone/studySystem/chromeServer"
	"github.com/zaddone/studySystem/config"
	"net/url"
)
var(
	//uri = "https://www.alimama.com/member/login.htm"
	orderUrl string = "https://pub.alimama.com/openapi/param2/1/gateway.unionpub/report.getTbkOrderDetails.json?t=1581752650893&_tb_token_=f087e0e378633&jumpType=0&positionIndex=&pageNo=1&startTime=2020-01-01&endTime=2020-02-15&queryType=2&tkStatus=&memberType=&pageSize=40"
	uri string = "https://login.taobao.com/member/login.jhtml?style=mini&newMini2=true&from=alimama"
	//uri string = "https://login.taobao.com/member/login.jhtml?"
	tb_token = "_tb_token_"
	indexUrl = "https://www.alimama.com/index.htm"
	//Png = config.Conf.Static+"/xcode.png"
	Png = "xcode.png"
	RunConTrol chan bool
	HandOrder func(interface{}) = nil
	Begin time.Time
	orderTimeFormat = "2006-01-02"
	orderTime = "2006-01-02 15:04:05"
	LoginPhone = "192.168.1.51"
	ShowPhone = "192.168.1.52"
	TaobaoLoginEvent func(path string) = nil

)
func init(){
	var err error
	Begin,err = time.Parse(orderTimeFormat,"2020-01-01")
	if err != nil {
		panic(err)
	}
	TaobaoLoginEvent = ControlPhoneEvent
}
func ControlPhoneEvent(path string){
	RunConTrol = make(chan bool)
	go func(){
		if err := InitPhone(LoginPhone,RunConTrol,func(p string){
			TaobaoLoginCheck(p)
		}); err != nil {
			panic(err)
		}
	}()
	go func(){
		if err := InitPhone(ShowPhone,RunConTrol,func(p string){
			ShowBrowser(p)
		}); err != nil {
			panic(err)
		}
	}()
}
func Run() error {
	//"https://www.alimama.com/index.htm"
	chromeServer.HandleResponse = CheckLoginPage
	chromeServer.PageNavigate(uri,func(db map[string]interface{}){
		fmt.Println(uri)
	})
	//GetOrderPage()
	return chromeServer.Run()
}
//func GetOrderPage(){
//	chromeServer.HandleResponse = GetOrder_
//	ourl,err := url.Parse(orderUrl)
//	if err != nil {
//		panic(err)
//	}
//	uVal,err := url.ParseQuery(ourl.RawQuery)
//	if err != nil {
//		panic(err)
//	}
//	uVal.Set("startTime",Begin.Format(orderTimeFormat))
//	uVal.Set("endTime",time.Now().Format(orderTimeFormat))
//	uVal.Set("queryType","1")
//	uVal.Set("pageNo","1")
//	uVal.Set("t",fmt.Sprintf("%d",time.Now().Unix()))
//	orderUrl = fmt.Sprintf("%s://%s%s?%s",ourl.Scheme,ourl.Host,ourl.Path,uVal.Encode())
//	chromeServer.PageNavigate(orderUrl,func(res map[string]interface{}){
//		fmt.Println(res)
//		//getBody(res,qu)
//	})
//
//}

func NextPage(){

	ourl,err := url.Parse(orderUrl)
	if err != nil {
		panic(err)
	}
	uVal,err := url.ParseQuery(ourl.RawQuery)
	if err != nil {
		panic(err)
	}
	uVal.Set("t",fmt.Sprintf("%d",time.Now().Unix()))
	page,err :=strconv.Atoi(uVal.Get("pageNo"))
	if err != nil {
		panic(err)
	}
	uVal.Set("pageNo",fmt.Sprintf("%d",page+1))
	orderUrl = fmt.Sprintf("%s://%s%s?%s",ourl.Scheme,ourl.Host,ourl.Path,uVal.Encode())
	fmt.Println(orderUrl)
	chromeServer.PageNavigate(orderUrl,func(res map[string]interface{}){
		fmt.Println(res)
		//getBody(res,qu)
	})

}

//func GetOrder_(_db interface{}){
//	if !chromeServer.GetBody(_db,uri,func(__id float64,result map[string]interface{}){
//		chromeServer.PageNavigate(Uri,func(res map[string]interface{}){
//			chromeServer.HandleResponse = CheckLoginPage
//		})
//	}){
//
//		CheckLogin(_db)
//	}
//}
func GetOrder(_db interface{}){
	chromeServer.GetBody(_db,"gateway.unionpub",func(__id float64,result map[string]interface{}){
		//fmt.Println(result)
		if HandOrder == nil {
			return
		}
		body := result["body"]
		if body == nil {
			return
		}
		var data map[string]interface{}
		err := json.Unmarshal([]byte(body.(string)),&data)
		if err != nil {
			fmt.Println(err)
			return
		}
		res := data["data"].(map[string]interface{})["result"]
		if res == nil {
			fmt.Println(data)
			return
		}
		li_ := res.([]interface{})
		for _,l := range li_ {
			l_ := l.(map[string]interface{})
			l_["order_id"] = l_["tradeId"]
			l_["status"] = false
			l_["fee"] = l_["pubSharePreFee"]
			l_["goodsid"] =fmt.Sprintf("%.0f",l_["itemId"].(float64))
			l_["goodsName"] = l_["itemTitle"]
			l_["goodsImg"] = l_["itemImg"]
			l_["site"] = "taobao"
			l_["time"] = time.Now().Unix()
			l_["text"] = l_["tkStatusText"]
			if l_["tkEarningTime"] != nil {
				endt,err := time.Parse(orderTime, l_["tkEarningTime"].(string))
				if err != nil {
					continue
					//panic(err)
				}
				//y,m,d := endt.Date()
				//l_["status"] = true
				l_["endTime"]= endt.Unix()
				y,m,_ := endt.AddDate(0,1,0).Date()
				l_["payTime"]= time.Date(y,m,21,0,0,0,0,endt.Location()).Unix()

			}
			//if l_["tkStatusText"].(string) =="已结算" {
			//	l_["status"] = true
			//	endt,err := time.Parse(orderTime, l_["tkEarningTime"].(string))
			//	if err != nil {
			//		panic(err)
			//	}
			//	l_["endTime"] =endt.Unix()

			//}
			HandOrder(l)
		}
		if len(li_) == 40 {
			NextPage()
			return
		}
		ourl,err := url.Parse(orderUrl)
		if err != nil {
			panic(err)
		}
		uVal,err := url.ParseQuery(ourl.RawQuery)
		if err != nil {
			panic(err)
		}
		if uVal.Get("queryType") == "3"{
			Begin = time.Now()
			chromeServer.ClosePage()
			return
		}

		uVal.Set("queryType","3")
		uVal.Set("pageNo","0")
		orderUrl = fmt.Sprintf("%s://%s%s?%s",ourl.Scheme,ourl.Host,ourl.Path,uVal.Encode())
		fmt.Println(orderUrl)
		chromeServer.PageNavigate(orderUrl,func(res map[string]interface{}){
			fmt.Println(res)
			//getBody(res,qu)
		})

		//NextPage()
		//chromeServer.HandleResponse = GetOrder
	})
}
func LoginSession(_db interface{}){
	chromeServer.GetBody(_db,indexUrl,func(__id float64,result map[string]interface{}){
		if RunConTrol != nil {
			close(RunConTrol)
		}
	//PageNavigate(viewOrder,func(res map[string]interface{}){
		//chromeServer.ShowCookies(func(db_ map[string]interface{}){
			//fmt.Println(db)
		chromeServer.HandleResponse = GetOrder
		ourl,err := url.Parse(orderUrl)
		if err != nil {
			panic(err)
		}
		uVal,err := url.ParseQuery(ourl.RawQuery)
		if err != nil {
			panic(err)
		}
		uVal.Set("startTime",Begin.Format(orderTimeFormat))
		uVal.Set("endTime",time.Now().Format(orderTimeFormat))
		uVal.Set("queryType","1")
		uVal.Set("pageNo","1")
		uVal.Set("t",fmt.Sprintf("%d",time.Now().Unix()))
		//for _,_c_ := range db_["cookies"].([]interface{}) {
		//	c_ := _c_.(map[string]interface{})
		//	name := c_["name"].(string)
		//	if strings.EqualFold(tb_token,name){
		//		uVal.Set(tb_token,c_["value"].(string))
		//		break
		//	}
		//}
		orderUrl = fmt.Sprintf("%s://%s%s?%s",ourl.Scheme,ourl.Host,ourl.Path,uVal.Encode())
		//fmt.Println(orderUrl)
		chromeServer.PageNavigate(orderUrl,func(res map[string]interface{}){
			fmt.Println(res)
			//getBody(res,qu)
		})
		return
		//})
	})
}

func CheckLoginPage(_db interface{}){
	if !chromeServer.GetBody(_db,uri,func(__id float64,result map[string]interface{}){
		fmt.Println("getBody")
	//for{
		time.Sleep(3*time.Second)
		chromeServer.GetDoc(func(body map[string]interface{}){
		//fmt.Println(body)
		chromeServer.FindAttributes("J_Static2Quick",body["root"].(map[string]interface{}),func(node map[string]interface{})bool{
			fmt.Println("find")
			//isseccess := true
			chromeServer.ClickBoxModel(node["nodeId"].(float64),func(){
				fmt.Println(node)
				//isseccess = false
			//	fmt.Println("click")
			//	chromeServer.InputText("zaddone",func(){
			//		fmt.Println("seccess")
			//	})
			})
			return false
		})
		})
		//if !isseccess {
		//	return
		//}
	//}
	}){
		CheckLogin(_db)
	}
}

func CheckLogin(_db interface{}){
	if !chromeServer.GetBody(_db,Png,func(__id float64,result map[string]interface{}){
		switch result["body"].(type){
		case string:
			body,err := base64.StdEncoding.DecodeString(result["body"].(string))
			if err != nil {
				panic(err)
			}
			f, _ := os.OpenFile(config.Conf.Static+"/"+Png, os.O_RDWR|os.O_CREATE, os.ModePerm)
			f.Write(body)
			f.Close()

			if TaobaoLoginEvent != nil {
				TaobaoLoginEvent(Png)
			}
			//fmt.Println("http://127.0.0.1"+":8001/"+Png)
			chromeServer.HandleResponse = LoginSession
			//chromeServer.HandleResponse = GetOrder
		default:

			fmt.Println(result["body"])
			chromeServer.PageNavigate(uri,func(res map[string]interface{}){
				fmt.Println(res)
				//getBody(res,qu)
			})
		}
		return
	}){
		//LoginSession(_db)
		GetOrder(_db)
	}

}
