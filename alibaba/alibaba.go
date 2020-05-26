package alibaba
import(
	"fmt"
	//"strconv"
	"github.com/zaddone/studySystem/chromeServer"
	//"github.com/zaddone/studySystem/shopping"
	"net/url"
	"encoding/json"
)
var (
	Url_1= "https://guanjia.1688.com/page/portal.htm"
	Url_2 = "https://guanjia.1688.com/page/start.htm"
	indexUrl = "https://guanjia.1688.com/event/app/newchannel_fx_selloffer/querySuplierProducts.htm?_input_charset=utf8&keyword=&pageNum=1"
	HandGoods func(interface{}) = nil
)
func GetGoodsList(_db interface{}){
	if !chromeServer.GetBody(_db,"page/start.htm",func(__id float64,result map[string]interface{}){
		chromeServer.PageNavigate(indexUrl,func(res map[string]interface{}){
			fmt.Println(res)
		})

	}){
		chromeServer.GetBody(_db,"querySuplierProducts.htm",func(__id float64,result map[string]interface{}){
			if HandGoods == nil {
				return
			}
			body := result["body"]
			if body == nil {
				chromeServer.ClosePage()
				return
			}
			var re map[string]interface{}
			err := json.Unmarshal([]byte(body.(string)),&re)
			if err != nil {
				panic(err)
			}
			re_ := re["result"].(map[string]interface{})
			for _,l := range  re_["sellOfferVOList"].([]interface{}){
				HandGoods(l)
			}

			num := re_["pageNum"].(float64)
			if re_["pageCount"].(float64) == num {
				chromeServer.ClosePage()
				return
			}
			u,err := url.Parse(indexUrl)
			if err != nil {
				panic(err)
			}
			val := u.Query()
			val.Set("pageNum",fmt.Sprintf("%.0f",num+1))
			uri_ := fmt.Sprintf("%s://%s%s?%s",u.Scheme,u.Host,u.Path,val.Encode())
			fmt.Println(uri_)
			chromeServer.PageNavigate(uri_,func(res map[string]interface{}){
				fmt.Println(res)
			})
			//fmt.Println(result)
		})
	}
}
func Run() error {
	chromeServer.HandleResponse = GetGoodsList
	return chromeServer.Run(Url_2)
}

