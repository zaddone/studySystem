package main
import (
    "net/http"
    "net/http/httputil"
    "log"
    //"fmt"
)

func main() {
	redi := http.NewServeMux()
	redi.HandleFunc("/wxpay/", func(w http.ResponseWriter, r *http.Request) {
		director := func(req *http.Request) {
			req.URL.Path = req.URL.Path[6:]
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:8083"
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	})
	redi.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		director := func(req *http.Request) {
			req.URL.Path = req.URL.Path[3:]
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:8082"
			log.Println(req.URL)
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
	redi.HandleFunc("/site/", func(w http.ResponseWriter, r *http.Request) {
		director := func(req *http.Request) {
			req.URL.Path = req.URL.Path[5:]
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:8080"
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	})
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
