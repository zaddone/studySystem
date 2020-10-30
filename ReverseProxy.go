package main
import (
    "net/http"
    "net/http/httputil"
    "log"
    "strings"
    "sync"
    //"fmt"
)
var (
	pathList = map[string]string{
		"shopping":"127.0.0.1:8087",
		"calendar":"127.0.0.1:8086",
		"content":"127.0.0.1:8085",
		"wxserver":"127.0.0.1:8084",
		"wxpay":"127.0.0.1:8083",
		"v1":"127.0.0.1:8082",
		"article":"127.0.0.1:8081",
		"site":"127.0.0.1:8080",
	}
	mutex sync.Mutex
)
func main() {
	redi := http.NewServeMux()
	for key,_ := range pathList {
		redi.HandleFunc("/"+key+"/", func(w http.ResponseWriter, r *http.Request) {
			director := func(req *http.Request) {
				pa := strings.Split(req.URL.Path,"/")[1]
				req.URL.Path = req.URL.Path[len(pa)+1:]
				req.URL.Scheme = "http"
				mutex.Lock()
				req.URL.Host = pathList[pa]
				mutex.Unlock()
				//log.Println(req.URL,key,val)
			}
			proxy := &httputil.ReverseProxy{Director: director}
			proxy.ServeHTTP(w, r)
		})
	}

	redi.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://www.miguotuijian.cn", http.StatusMovedPermanently)
		return
	})
	go func(){
		log.Fatal(http.ListenAndServeTLS(":443","./3375181_zaddone.com.pem","./3375181_zaddone.com.key", redi))
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://www.miguotuijian.cn", http.StatusMovedPermanently)
		return
	})
	go func(){
		log.Fatal(http.ListenAndServe(":80", mux))
	}()
	select{}
}
