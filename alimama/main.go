package main
import(
	"fmt"
	"encoding/base64"
	"os"
	"time"
	"strings"
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

)
func main(){
	//"https://www.alimama.com/index.htm"
	chromeServer.HandleResponse = CheckLogin
	chromeServer.Run(uri)
}
func GetOrder(_db interface{}){
	chromeServer.GetBody(_db,"gateway.unionpub",func(__id float64,result map[string]interface{}){
		
		fmt.Println(result)
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
			qu := fmt.Sprintf("%s://%s%s?%s",ourl.Scheme,ourl.Host,ourl.Path,uVal.Encode())
			chromeServer.PageNavigate(qu,func(res map[string]interface{}){
				fmt.Println(res)
				//getBody(res,qu)
			})
			return
		})
	})
}

func CheckLogin(_db interface{}){

	if !chromeServer.GetBody(_db,Png,func(__id float64,result map[string]interface{}){
		//fmt.Println("--------------")
		//fmt.Println(result)
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
				//TaobaoLoginCheck(p)
			}); err != nil {
				panic(err)
			}
		}()
		//InitPhone(*LoginPhone,)
		fmt.Println("http://127.0.0.1"+":8001/"+Png)
		//TaobaoLoginCheck()
		//fmt.Println(result["body"].(string))
		chromeServer.HandleResponse = LoginSession
		return
	}){
		LoginSession(_db)
	}

}
