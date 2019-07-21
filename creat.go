package main
import(
	//"fmt"
	"github.com/zaddone/studySystem/wxmsg"
)
func main(){
	err := wxmsg.CreateColl("pageColl")
	if err != nil {
		//fmt.Println(err)
		panic(err)
	}
}
