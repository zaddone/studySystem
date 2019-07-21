package chrome

import(
	"fmt"
	"testing"
	"time"
)
func Test_run(t *testing.T){
	start(func(in string){
		w := runBrowserStream(in,func(db interface{}){
			fmt.Println(in)
		})
		w<-map[string]interface{}{"method":"Browser.close","id":time.Now().Unix()}
		select{}
	})
	//select{}
	//fmt.Println("t",t)
	//t.Log("test-------------------")

}
