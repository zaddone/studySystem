package wxmsg
import(
	"io"
	"fmt"
	"bytes"
	"strings"
	"net/url"
	"net/http"
	//"io/ioutil"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/config"
	"encoding/json"
	"time"
)
var(
	//wxAppId = "wx92ebd09c7b0d944f"
	//wxSec	= "b3005d3c298e27b60ee1f90d188a9d86"
	wxToKenUrl= "https://api.weixin.qq.com/cgi-bin/token"
	//wxToKenUrl= "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=wx92ebd09c7b0d944f&secret=b3005d3c298e27b60ee1f90d188a9d86"
	toKen string
	env string = "guomi-2i7wu"
	MaxCount float64 = 10000
	//ExpiresIn int64
)

func setToken() int {
	db := map[string]interface{}{}
	err := request.ClientHttp(wxToKenUrl,"GET",[]int{200},nil,func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&db)
	})
	if err != nil {
		panic(err)
	}
	if db["access_token"]==nil {
		return setToken()
	}
	toKen = db["access_token"].(string)
	return int(db["expires_in"].(float64)) - 100

}
func init(){
	wxToKenUrl = fmt.Sprintf("%s?%s",wxToKenUrl,
	(&url.Values{
		"grant_type":	[]string{"client_credential"},
		"appid":	[]string{config.Conf.WXAppid},
		"secret":	[]string{config.Conf.WXSec},
	}).Encode())

	//fmt.Println(wxToKenUrl)
	k := setToken()
	fmt.Println("setToKen",k)
	//k := time.Duration(setToken())*time.Second
	go func(){
		for{

			time.Sleep(time.Duration(k)*time.Second)
			k = setToken()
		}
	}()

}
func PostRequest(url string,PostBody map[string]interface{},h func(io.Reader)error) error {
	url = fmt.Sprintf("%s?access_token=%s",url,toKen)
	PostBody["env"]=env
	db,err := json.Marshal(PostBody)
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(db))
	return request.ClientHttp_(url,"POST",bytes.NewReader(db),http.Header{"Content-Type":[]string{"application/x-www-form-urlencoded","multipart/form-data"}},func(body io.Reader,st int)error{
		if st == 200 {
			return h(body)
		}
		var da [8192]byte
		n,err := body.Read(da[:])
		return fmt.Errorf("status code %d %s %s", st, url,string(da[:n]),err)
	})
	//return request.ClientHttp(fmt.Sprintf("%s?access_token=%s",url,toKen),"POST",[]int{200},PostBody,h)

}
func CollectionClearDB(hand func()error)error {

	return PostRequest("https://api.weixin.qq.com/tcb/databasecollectionget",map[string]interface{}{},func(body io.Reader)error{
		var res  map[string]interface{}
		json.NewDecoder(body).Decode(&res)

		errcode := res["errcode"].(float64)
		if errcode != 0 {
			return fmt.Errorf("%.0f %s",errcode,res["errmsg"].(string))
		}
		for _,coll := range res["collections"].([]interface{}){
			c := coll.(map[string]interface{})
			if !strings.EqualFold(c["name"].(string),config.Conf.CollPageName){
				continue
			}
			if MaxCount > c["count"].(float64) {
				return nil
			}
			return hand()

		}
		return nil
	})

}

func DeleteColl(c_name string) error {

	return PostRequest("https://api.weixin.qq.com/tcb/databasecollectiondelete",map[string]interface{}{"collection_name":c_name},func(body io.Reader)error{
		var res  map[string]interface{}
		json.NewDecoder(body).Decode(&res)
		if res["errcode"].(float64) == 0 {
			return nil
		}
		//fmt.Println(res,res["errcode"].(float64),res["errmsg"].(string))
		return fmt.Errorf(res["errmsg"].(string))
	})

}
func CreateColl(c_name string) error {

	return PostRequest("https://api.weixin.qq.com/tcb/databasecollectionadd",map[string]interface{}{"collection_name":c_name},func(body io.Reader)error{
		var res  map[string]interface{}
		json.NewDecoder(body).Decode(&res)
		if res["errcode"].(float64) == 0 {
			return nil
		}
		//fmt.Println(res,res["errcode"].(float64),res["errmsg"].(string))
		return fmt.Errorf(res["errmsg"].(string))
	})

}

func DBDelete(ids []string)error {
	fmt.Println(ids)
	return PostRequest(
		"https://api.weixin.qq.com/tcb/databasedelete",
		map[string]interface{}{
			"query":fmt.Sprintf(
				"db.collection(\"%s\").where({_id:db.command.in([%s])}).remove()",
				config.Conf.CollPageName,
				strings.Join(ids,","))},
		func(body io.Reader)error{

		var res  map[string]interface{}
		json.NewDecoder(body).Decode(&res)
		errcode := res["errcode"].(float64)
		if errcode == 0 {
			return nil
		}
		return fmt.Errorf("%.0f %s",errcode,res["errmsg"].(string))
	})
}

func UpdateToWXDB(id uint64,ids []string) error {
	//fmt.Println(id,ids)
	var res  map[string]interface{}
	err := PostRequest(
		"https://api.weixin.qq.com/tcb/databaseupdate",
		map[string]interface{}{
			"query":fmt.Sprintf("db.collection(\"%s\").where({_id:db.command.in(%s)}).update({data:{link:db.command.unshift(\"%d\")}})",config.Conf.CollPageName,fmt.Sprintf("[%s]",strings.Join(ids,",")),id)},
		func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&res)
	})
	if err != nil {
		return err
	}
	if res["errcode"].(float64) != 0 {
		return fmt.Errorf("%.0f %s",res["errcode"].(float64),res["errmsg"].(string))
	}
	return nil
}

func AddToWXDB(coll string,body string) error {
	//fmt.Println(body)
	var res  map[string]interface{}
	err := PostRequest(
		"https://api.weixin.qq.com/tcb/databaseadd",
		map[string]interface{}{
			"query":fmt.Sprintf("db.collection(\"%s\").add({data:[%s]})",coll,body)},
		func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&res)
	})
	if err != nil {
		return err
	}
	if res["errcode"].(float64) != 0 {
		return fmt.Errorf("%.0f %s",res["errcode"].(float64),res["errmsg"].(string))
	}
	return nil
}
func SaveToWXDB(body string) error {
	return AddToWXDB(config.Conf.CollPageName,body)
	//fmt.Println(body)
	//var res  map[string]interface{}
	//err := PostRequest(
	//	"https://api.weixin.qq.com/tcb/databaseadd",
	//	map[string]interface{}{
	//		"query":fmt.Sprintf("db.collection(\"%s\").add({data:[%s]})",config.Conf.CollPageName,body)},
	//	func(body io.Reader)error{
	//	return json.NewDecoder(body).Decode(&res)
	//})
	//if err != nil {
	//	return err
	//}
	//if res["errcode"].(float64) != 0 {
	//	return fmt.Errorf("%.0f %s",res["errcode"].(float64),res["errmsg"].(string))
	//}
	//return nil
}
