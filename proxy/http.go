package main

import (
    "flag"
    //"fmt"
    "io"
    "log"
    "net"
    "net/http"
    "strings"
)

type Pxy struct {
    Cfg Cfg
}

type Cfg struct {
    Addr        string   // 监听地址
    Port        string   // 监听端口
    IsAnonymous bool     // 高匿名模式
    Debug       bool     // 调试模式
}

func main() {

    faddr := flag.String("addr","0.0.0.0","监听地址，默认0.0.0.0")
    fprot := flag.String("port","8080","监听端口，默认8080")
    fanonymous :=  flag.Bool("anonymous",true,"高匿名，默认高匿名")
    fdebug :=  flag.Bool("debug",true,"调试模式显示更多信息，默认关闭")
    flag.Parse()

    cfg := &Cfg{}
    cfg.Addr = *faddr
    cfg.Port = *fprot
    cfg.IsAnonymous = *fanonymous
    cfg.Debug = *fdebug
    Run(cfg)

}

func Run(cfg *Cfg) {
    pxy := NewPxy()
    pxy.SetPxyCfg(cfg)
    log.Printf("HttpPxoy is runing on %s:%s \n", cfg.Addr, cfg.Port)
    // http.Handle("/", pxy)
    bindAddr := cfg.Addr + ":" + cfg.Port
    log.Fatalln(http.ListenAndServe(bindAddr, pxy))
}


func NewPxy() *Pxy {
    return &Pxy{
        Cfg: Cfg{
            Addr:        "",
            Port:        "8081",
            IsAnonymous: true,
            Debug:       false,
        },
    }
}

func (p *Pxy) SetPxyCfg(cfg *Cfg) {
    if cfg.Addr != "" {
        p.Cfg.Addr = cfg.Addr
    }
    if cfg.Port != "" {
        p.Cfg.Port = cfg.Port
    }
    if cfg.IsAnonymous != p.Cfg.IsAnonymous {
        p.Cfg.IsAnonymous = cfg.IsAnonymous
    }
    if cfg.Debug != p.Cfg.Debug {
        p.Cfg.Debug = cfg.Debug
    }

}

func (p *Pxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    // debug
    if p.Cfg.Debug {

	log.Printf("Received request %s %s %s\n", req.Method, req.Host, req.RemoteAddr)

    }

    // http && https
    if req.Method != "CONNECT" {
        // 处理http
        p.HTTP(rw, req)
    } else {
        // 处理https
        // 直通模式不做任何中间处理
        p.HTTPS(rw, req)
    }

}

func (p *Pxy) HTTP(rw http.ResponseWriter, req *http.Request) {

    transport := http.DefaultTransport
    outReq := new(http.Request)
    *outReq = *req
    if p.Cfg.IsAnonymous == false {
        if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
            if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
                clientIP = strings.Join(prior, ", ") + ", " + clientIP
            }
            outReq.Header.Set("X-Forwarded-For", clientIP)
        }
    }
    res, err := transport.RoundTrip(outReq)
    if err != nil {
        rw.WriteHeader(http.StatusBadGateway)
        rw.Write([]byte(err.Error()))
        return
    }
    for key, value := range res.Header {
        for _, v := range value {
            rw.Header().Add(key, v)
        }
    }
    rw.WriteHeader(res.StatusCode)
    io.Copy(rw, res.Body)
    res.Body.Close()
}

func (p *Pxy) HTTPS(rw http.ResponseWriter, req *http.Request) {

    host := req.URL.Host
    hij, ok := rw.(http.Hijacker)
    if !ok {
        log.Printf("HTTP Server does not support hijacking")
    }

    client, _, err := hij.Hijack()
    if err != nil {
        return
    }
    server, err := net.Dial("tcp", host)
    if err != nil {
        return
    }
    client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
    go io.Copy(server, client)
    go io.Copy(client, server)
}
