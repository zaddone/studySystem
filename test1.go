package main

import (
	//"context"
	"log"
	"fmt"
	"io"
	//"strings"
	"bytes"
	//"time"
	//"github.com/chromedp/chromedp"
	//"github.com/chromedp/chromedp/cdp"
	//"github.com/chromedp/cdproto/cdp"
	//"github.com/chromedp/chromedp/runner"
	"os/exec"
	//"golang.org/x/net/websocket"
	"github.com/gorilla/websocket"
	//"regexp"
)

func main() {

	//reg := regexp.MustCompile(`\s`)
	k := []byte{10,68,101,118,84,111,111,108,115,32}
	port := "9222"

	//google-chrome --remote-debugging-port=9222 --headless --no-sandbox --no-default-browser-check --disable-gpu
	//fmt.Println([]byte("\\r"))
	runout := func(r io.ReadCloser){
		var db [8192]byte
		for{
			n,err := r.Read(db[:])
			if err != nil {
				fmt.Println(err)
				r.Close()
				return
			}
			//fmt.Println(string(db[:n]))
			if bytes.HasPrefix(db[:n],k){
				//fmt.Println(string(db[23:n-1]))
				Monitor(string(db[23:n-1]))
				//websocket.Dial(string(db[23:n-1]),"","")
			}
		}
	}
	op :=[]string{
		"--remote-debugging-port="+port,
		//"--headless",
		//"--disable-gpu",
		"--no-sandbox",
		"--no-default-browser-check",
		"https://www.toutiao.com/ch/news_baby/",
	}
	cmd := exec.Command("google-chrome",op... )
	out,err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	outerr,err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	go runout(out)
	go runout(outerr)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Wait()

}

func Monitor(ws string){
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	//defer c.Close()
	//fmt.Println("start ",ws)
	err = c.WriteJSON(map[string]interface{}{"Browser":"grantPermissions"})
	if err != nil {
		panic(err)
	}

	go func(){
		var db interface{}
		for{
			err = c.ReadJSON(&db)
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Println(db)
		}
	}()

}
