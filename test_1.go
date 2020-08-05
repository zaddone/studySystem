package main
import(
	"fmt"
	"net/http"
	"encoding/json"
)
func main(){
	url := "http://www.kuaidi.com/index-ajaxselectcourierinfo-4307091472913-yunda.html"
	res,err := http.Get(url)
	if err != nil {
		panic(err)
	}
	var db interface{}
	err = json.NewDecoder(res.Body).Decode(&db)
	if err != nil {
		panic(err)
	}
	fmt.Println(db)
}
