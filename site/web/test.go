package main
import(
	//"regexp"
	//"flag"
	"fmt"
	"time"
)
var(
	//WXtoken = "zhaoweijie2020"
	//OrderIDReg = regexp.MustCompile(`(jd|京东)[\s|\S]*(\d{12})`)
	//httpReg = regexp.MustCompile(`http`)
	//jdReg = regexp.MustCompile(`\/(\d+)\.html`)
	//jdReg_ = regexp.MustCompile(`sku=(\d+)`)
	//pddReg = regexp.MustCompile(`goods_id=(\d+)`);
	//cmdReg = regexp.MustCompile(`([a-zA-Z|\p{Han}]+)(\d+)`)


	//cmd  = flag.String("c","www.zaddone.com:443","cmd")
	orderTimeFormat = "2006010215"
)

func main(){
	t_,err := time.Parse(orderTimeFormat,"2020010101")
	if err != nil {
		panic(err)
	}
	fmt.Println(t_)
}
