package alimama
import(
	"fmt"
	"encoding/base64"
	"os"
	"time"
	"strings"
	"strconv"
	"github.com/zaddone/studySystem/chromeServer"
	"net/url"
)
var(
	//uri = "https://www.alimama.com/member/login.htm"
	orderUrl string = "https://pub.alimama.com/openapi/param2/1/gateway.unionpub/report.getTbkOrderDetails.json?t=1581752650893&_tb_token_=f087e0e378633&jumpType=0&positionIndex=&pageNo=1&startTime=2020-01-01&endTime=2020-02-15&queryType=2&tkStatus=&memberType=&pageSize=40"
	uri string = "https://login.taobao.com/member/login.jhtml?style=mini&newMini2=true&from=alimama"
	tb_token = "_tb_token_"
	indexUrl = "https://www.alimama.com/index.htm"
	Png = "xcode.png"
	RunConTrol chan bool
	HandOrder func(interface{}) = nil
	Begin time.Time
	orderTimeFormat = "2006-01-02"

)
func init(){
	var err error
	Begin,err = time.Parse(orderTimeFormat,"2020-01-01")
	if err != nil {
		panic(err)
	}
}
func Run(){
	//"https://www.alimama.com/index.htm"
	chromeServer.HandleResponse = CheckLogin
	chromeServer.Run(uri)
}

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
	chromeServer.PageNavigate(orderUrl,func(res map[string]interface{}){
		fmt.Println(res)
		//getBody(res,qu)
	})

}

func GetOrder(_db interface{}){
	chromeServer.GetBody(_db,"gateway.unionpub",func(__id float64,result map[string]interface{}){
		if HandOrder != nil {
			li := result["data"].(map[string]interface{})["result"]
			if li == nil {
				return
			}
			li_ := li.([]interface{})
			for _,l := range li_ {
				HandOrder(l)
			}
			if len(li_) == 40 {
				NextPage()
			}
		}else{
			fmt.Println(result)
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
			return
		}

		uVal.Set("queryType","3")
		uVal.Set("pageNo","0")
		orderUrl = fmt.Sprintf("%s://%s%s?%s",ourl.Scheme,ourl.Host,ourl.Path,uVal.Encode())
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
		chromeServer.ShowCookies(func(db_ map[string]interface{}){
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
			for _,_c_ := range db_["cookies"].([]interface{}) {
				c_ := _c_.(map[string]interface{})
				name := c_["name"].(string)
				if strings.EqualFold(tb_token,name){
					uVal.Set(tb_token,c_["value"].(string))
					uVal.Set("t",fmt.Sprintf("%d",time.Now().Unix()))
					//fmt.Println(c_)
					break
				}
			}
			orderUrl = fmt.Sprintf("%s://%s%s?%s",ourl.Scheme,ourl.Host,ourl.Path,uVal.Encode())
			chromeServer.PageNavigate(orderUrl,func(res map[string]interface{}){
				fmt.Println(res)
				//getBody(res,qu)
			})
			return
		})
	})
}

func CheckLogin(_db interface{}){

	if !chromeServer.GetBody(_db,Png,func(__id float64,result map[string]interface{}){
		body,err := base64.StdEncoding.DecodeString(result["body"].(string))
		if err != nil {
			panic(err)
		}
		f, _ := os.OpenFile(Png, os.O_RDWR|os.O_CREATE, os.ModePerm)
		f.Write(body)
		f.Close()
		RunConTrol = make(chan bool)
		go func(){
			if err := InitPhone(*LoginPhone,RunConTrol,func(p string){
				TaobaoLoginCheck(p)
			}); err != nil {
				panic(err)
			}
		}()
		go func(){
			if err := InitPhone(*ShowPhone,RunConTrol,func(p string){
				ShowBrowser(p)
			}); err != nil {
				panic(err)
			}
		}()
		fmt.Println("http://127.0.0.1"+":8001/"+Png)
		chromeServer.HandleResponse = LoginSession
		return
	}){
		LoginSession(_db)
	}

}
