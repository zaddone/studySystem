package main
import(
	"regexp"
	"flag"
	"fmt"
)
var(
	//WXtoken = "zhaoweijie2020"
	OrderIDReg = regexp.MustCompile(`(jd|京东)[\s|\S]*(\d{12})`)
	httpReg = regexp.MustCompile(`http`)
	jdReg = regexp.MustCompile(`\/(\d+)\.html`)
	jdReg_ = regexp.MustCompile(`sku=(\d+)`)
	pddReg = regexp.MustCompile(`goods_id=(\d+)`);
	cmdReg = regexp.MustCompile(`([a-zA-Z|\p{Han}]+)(\d+)`)


	cmd  = flag.String("c","www.zaddone.com:443","cmd")
)


func init(){
	flag.Parse()
}
func handMsg(str string){
	n := httpReg.FindStringIndex(str)
	if len(n)>0{
		handHttp(str[n[0]:])
		return
	}
	s := cmdReg.FindStringSubmatch(str)
	fmt.Println(s)
	if len(s)==3 {
		handCmd(s[1],s[2])
		return
	}
	fmt.Println("hand",str)
}
func handCmd(name,num string){
	fmt.Println(name,num)
}
func handHttp(str string) {
	ss := pddReg.FindStringSubmatch(str)
	if len(ss) >1 {
		fmt.Println("pdd",ss[1])
		return
	}
	ss = jdReg.FindStringSubmatch(str)
	if len(ss) >1 {
		fmt.Println("jd",ss[1])
		return
	}
	ss = jdReg_.FindStringSubmatch(str)
	if len(ss) >1 {
		fmt.Println("jd",ss[1])
		return
	}

	fmt.Println("find not",str)
}
func main(){
	handMsg(*cmd)
	//handMsg("https://wqitem.jd.com/item/view?sku=59319389027&_fromappid=jwx2c49e82e87e57ff0&_ph_=1&utm_source=iosapp&utm_medium=appshare&utm_campaign=t_335139774&utm_term=CopyURL&ad_od=share")
}
