package main
import(
	"regexp"
	//"flag"
	"fmt"
	//"time"
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
	//orderTimeFormat = "2006010215"
	taobaoid = regexp.MustCompile(`[\?|\&]id=(\d+)`)
)

func main(){
	//t_,err := time.Parse(orderTimeFormat,"2020010101")
	//if err != nil {
	//	panic(err)
	//}
	str := taobaoid.FindStringSubmatch("https://item.taobao.com/item.htm?spm=a219r.lm869.14.1.1d667ee4vRNM0F&id=611800287553&ns=1&abbucket=12#detail")
	fmt.Println(str)
}
