package main
import (
    "net/http"
    "net/http/httputil"
    "log"
)

func main() {
	redi := http.NewServeMux()
	redi.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		director := func(req *http.Request) {
			req.URL.Path = req.URL.Path[3:]
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:8082"
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	})
	redi.HandleFunc("/article/", func(w http.ResponseWriter, r *http.Request) {
		director := func(req *http.Request) {
			req.URL.Path = req.URL.Path[8:]
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:8081"
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	})
	redi.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() =="/"{
			http.Redirect(w, r, "https://www.miguotuijian.cn", http.StatusMovedPermanently)
			return
		}
		//log.Println(r.URL)
		director := func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:8080"
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	})
	go func(){
		log.Fatal(http.ListenAndServeTLS(":443","./3375181_zaddone.com.pem","./3375181_zaddone.com.key", redi))
		//log.Fatal(http.ListenAndServe(":80",redi))
	}()
	mux := http.NewServeMux()
	//mux.HandleFunc("/", Redirect301Handler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//r.Header.Add("Access-Control-Allow-Origin","https://www.zaddone.com")
		if r.URL.String() =="/"{
			http.Redirect(w, r, "https://www.miguotuijian.cn", http.StatusMovedPermanently)
			return
		}
		director := func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:8080"
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	})
	//err := http.ListenAndServe(":80", mux)
	go func(){
		log.Fatal(http.ListenAndServe(":80", mux))
	}()
	select{}
}
