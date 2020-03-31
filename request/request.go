package request
import(
	//"strings"
	"io"
	"fmt"
	"net/http"
	//"net/url"
	"net/http/cookiejar"
	"github.com/zaddone/studySystem/config"
	"compress/gzip"
	"compress/flate"
)

var (
	Jar *cookiejar.Jar
)
func init(){
	var err error
	Jar,err =cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
}
func ClientHttp__(path string,ty string,r io.Reader,h http.Header, hand func(io.Reader,*http.Response)error) error {
	Req, err := http.NewRequest(ty,path,r)
	if err != nil {
		return err
	}
	if h != nil {
		//Req.Header = h
		for k,v := range h {
			for _,_v := range v{
				Req.Header.Set(k,_v)
			}
		}
	}else{
		//Req.Header = config.Conf.Header
		for k,v := range config.Conf.Header {
			for _,_v := range v {
				Req.Header.Set(k,_v)
			}
		}
	}
	Cli := &http.Client{Jar:Jar}
	res, err := Cli.Do(Req)
	if err != nil {
		return err
	}
	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
		//defer reader.Close()
	default:
		reader = res.Body
	}
	if hand != nil {
		err = hand(reader,res)
	}
	reader.Close()
	return err
}

func ClientHttpR_(path string,ty string,r io.Reader,referer string,h http.Header, hand func(io.Reader,int)error) error {
	Req, err := http.NewRequest(ty,path,r)
	if err != nil {
		return err
	}
	if h != nil {
		for k,v := range h {
			for _,_v := range v{
				Req.Header.Add(k,_v)
			}
		}
	}
	Req.Header.Add("Referer",referer)
	Cli := &http.Client{Jar:Jar}
	res, err := Cli.Do(Req)
	if err != nil {
		return err
	}

	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
		//defer reader.Close()
	default:
		reader = res.Body
	}
	if hand != nil {
		err = hand(reader,res.StatusCode)
	}
	reader.Close()
	return err

}
func ClientHttp_(path string,ty string,r io.Reader,h http.Header, hand func(io.Reader,int)error) error {
	//fmt.Println(path)
	Req, err := http.NewRequest(ty,path,r)
	if err != nil {
		return err
	}
	if h != nil {
		Req.Header = h
	}
	//	for k,v := range h {
	//		for _,_v := range v{
	//			Req.Header.Set(k,_v)
	//		}
	//	}
	//}
	//fmt.Println(Req.Header)
	Cli := &http.Client{Jar:Jar}
	res, err := Cli.Do(Req)
	if err != nil {
		return err
	}

	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
		//defer reader.Close()
	default:
		reader = res.Body
	}
	if hand != nil {
		err = hand(reader,res.StatusCode)
	}
	reader.Close()
	return err

}
func ClientHttp(path string,ty string,statu []int,PostDB io.Reader, hand func(io.Reader)error) error {
	return ClientHttp_(path,ty,PostDB,config.Conf.Header,func(body io.Reader,st int) error {
		for _,s := range statu {
			if s == st {
				return hand(body)
			}
		}
		var da [8192]byte
		n,err := body.Read(da[:])
		return fmt.Errorf("status code %d %s %s", st, path,string(da[:n]),err)
	})
}

