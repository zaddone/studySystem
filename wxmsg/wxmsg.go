package wxmsg
import(
	"os"
	"io"
	"io/ioutil"
	"fmt"
	"bytes"
	"strings"
	"net/url"
	"net/http"
	"mime/multipart"
	//"io/ioutil"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/config"
	"encoding/json"
	"path/filepath"
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

		fmt.Println(db)
		time.Sleep(1*time.Hour)
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

func DBDelete(coll string,ids []string)error {
	fmt.Println(ids)
	return PostRequest(
		"https://api.weixin.qq.com/tcb/databasedelete",
		map[string]interface{}{
			"query":fmt.Sprintf(
				"db.collection(\"%s\").where({_id:db.command.in([%s])}).remove()",
				coll,
				//config.Conf.CollPageName,
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
func UpdateWXDB(coll string,_id string,body string) error {
	//fmt.Println(body)
	var res  map[string]interface{}
	err := PostRequest(
		"https://api.weixin.qq.com/tcb/databaseupdate",
		map[string]interface{}{
			"query":fmt.Sprintf("db.collection(\"%s\").doc(\"%s\").update({data:{%s}})",coll,_id,body)},
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
}
func UpDBToWX(coll,uri string)error{
	fp,pid,err := UpFileToWX(uri)
	if err != nil {
		//panic(err)
		return err
	}
	fmt.Println(fp,pid)
	var res map[string]interface{}
	err = PostRequest(
		"https://api.weixin.qq.com/tcb/databasemigrateimport",
		map[string]interface{}{
			"collection_name":coll,
			"file_path":fp,
			"file_type":1,
			"stop_on_error":false,
			"conflict_mode":2,
		},
		func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&res)
	})
	if err != nil{
		//panic(err)
		return err
	}
	fmt.Println("databasemigrateimport",res)
	if res["errcode"].(float64) != 0 {
		return fmt.Errorf(res["errmsg"].(string))
	}
	job_id := res["job_id"]
	for {
		<-time.After(5*time.Second)
		err = PostRequest(
			"https://api.weixin.qq.com/tcb/databasemigratequeryinfo",
			map[string]interface{}{
				"job_id":job_id,
			},
			func(body io.Reader)error{
			return json.NewDecoder(body).Decode(&res)
		})
		if err != nil {
			return err
			//log.Println(err)
		}
		fmt.Println("info",res)
		if res["errcode"].(float64) != 0 {
			return fmt.Errorf(res["errmsg"].(string))
		}

		if strings.EqualFold(res["status"].(string),"fail"){
			panic(res)
		}
		if strings.EqualFold(res["status"].(string),"success"){
			fmt.Println(res)
			break
		}

	}
	err = PostRequest(
			"https://api.weixin.qq.com/tcb/batchdeletefile",
			map[string]interface{}{
				"fileid_list":[]string{pid},
			},
			func(body io.Reader)error{
			return json.NewDecoder(body).Decode(&res)
		})

	fmt.Println("del",res)
	if res["errcode"].(float64) != 0 {
		fmt.Println(res)
		return fmt.Errorf(res["errmsg"].(string))
	}
	return os.Remove(uri)

}

func UpFileToWX_(uri,fileName string) (string,string,error) {
	_,err := os.Stat(uri)
	if err != nil {
		return "","",err
	}

	//fileName := fmt.Sprintf("%d.mp3",time.Now().Unix())
	var res  map[string]interface{}
	err = PostRequest(
		"https://api.weixin.qq.com/tcb/uploadfile",
		map[string]interface{}{
			"path":fileName,
		},
		func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&res)
	})
	if err != nil {
		panic(err)
		return "","",err
	}
	params := map[string]io.Reader{
		"key":strings.NewReader(fileName),
		"signature":strings.NewReader(res["authorization"].(string)),
		"x-cos-security-token":strings.NewReader(res["token"].(string)),
		"x-cos-meta-fileid":strings.NewReader(res["cos_file_id"].(string)),
		"file":mustOpen(uri),
	}
	//client := http.DefaultClient
	err = Upload(res["url"].(string), params,fileName)
	if err != nil {
		return "","",err
		//panic(err)
	}
	//fmt.Println(res["file_id"].(string))
	//return res["file_id"].(string),nil
	return fileName,res["file_id"].(string),nil
}

func UpFileToWX(uri string) (string,string,error) {

	//fmt.Println(uri)
	//setToken()
	fi,err := os.Stat(uri)
	if err != nil {
		return "","",err
	}

	fileName := fmt.Sprintf("%s/%d",fi.Name(),time.Now().Unix())
	var res  map[string]interface{}
	err = PostRequest(
		"https://api.weixin.qq.com/tcb/uploadfile",
		map[string]interface{}{
			"path":fileName,
		},
		func(body io.Reader)error{
		return json.NewDecoder(body).Decode(&res)
	})
	if err != nil {
		panic(err)
		return "","",err
	}
	params := map[string]io.Reader{
		"key":strings.NewReader(fileName),
		"signature":strings.NewReader(res["authorization"].(string)),
		"x-cos-security-token":strings.NewReader(res["token"].(string)),
		"x-cos-meta-fileid":strings.NewReader(res["cos_file_id"].(string)),
		"file":mustOpen(uri),
	}
	//client := http.DefaultClient
	err = Upload(res["url"].(string), params,fileName)
	if err != nil {
		return "","",err
		//panic(err)
	}
	//fmt.Println(res["file_id"].(string))
	//return res["file_id"].(string),nil
	return fileName,res["file_id"].(string),nil

	//req,err := NewUploadRequest(res["url"].(string),params,"file",uri)
	//if err != nil {
	//	panic(err)
	//	return "",err
	//}
	//resp, err := http.DefaultClient.Do(req)
	////resp, err := http.Client.Do(req)
	//if err != nil {
	//	panic(err)
	//	return "", err
	//}
	//defer resp.Body.Close()
	//db ,err  := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return "",err
	//}
	//if bytes.Contains(db,[]byte("Error")){
	//	return "",fmt.Errorf(string(db))
	//}
	//var ret interface{}
	//if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
	//	panic(err)
	//	return "",err
	//}
	//fmt.Println(ret)
	//return res["file_id"].(string),nil

}
func Upload(url string, values map[string]io.Reader,fileName string) (err error) {
    // Prepare a form that you will submit to that URL.
    var b bytes.Buffer
    w := multipart.NewWriter(&b)
    for key, r := range values {
        var fw io.Writer
        if x, ok := r.(io.Closer); ok {
            defer x.Close()
        }
        // Add an image file
        if _, ok := r.(*os.File); ok {
	//	fmt.Println(x.Name())
            if fw, err = w.CreateFormFile(key,fileName); err != nil {
                return
            }
        } else {
            // Add other fields
            if fw, err = w.CreateFormField(key); err != nil {
                return
            }
        }
        if _, err = io.Copy(fw, r); err != nil {
            return err
        }

    }

    w.Close()
    req, err := http.NewRequest("POST", url, &b)
    if err != nil {
        return
    }
    req.Header.Set("Content-Type", w.FormDataContentType())
    //fmt.Println()
    Cli := &http.Client{}
    res, err := Cli.Do(req)
    if err != nil {
        return
    }
    if res.StatusCode != 204 {

	db ,err  := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(db))
        return fmt.Errorf("bad status: %s", res.Status)
    }
    return
}

func mustOpen(f string) *os.File {
    r, err := os.Open(f)
    if err != nil {
        panic(err)
    }
    return r
}

func NewUploadRequest(link string, params map[string]string, name, path string) (*http.Request, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	//part, err := writer.CreateFormField(name)
	part, err := writer.CreateFormFile(name, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, fp)
	if err != nil {
		return nil, err
	}
	for key, val := range params {
		//fmt.Println(key,val)
		err = writer.WriteField(key, val)
		if err != nil {
			return nil, err
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", link, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "multipart/form-data")
	//fmt.Println(req.MultipartForm.Value)
	return req, nil
}
