package main
import(
	"os/exec"
	"fmt"
	"io"
	"log"
	"strings"
	"bufio"
	"regexp"
	"strconv"
	"time"
)
var(
	ip string = "192.168.1."
	port string ="5555"
	regName = regexp.MustCompile(`^\S+`)
	regMax = regexp.MustCompile(`max (\d+),`)
	regMax_ = regexp.MustCompile(`(\d+)x(\d+)`)
)
func cmdToAdb(h func(string)error,opt ...string)(err error){
	runout := func(r io.Reader){
		buf := bufio.NewReader(r)
		for{
			s,err:= buf.ReadString('\n')
			if err != nil {
				if err != io.EOF{
					log.Println(err)
				}
				return
			}
			if len(s)>0 && h != nil {
				er := h(s)
				if er != nil{
					err = er
					break
				}
			}
		}
	}
	fmt.Println(opt)
	cmd := exec.Command("adb",opt...)
	out,err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return err
		//log.Fatal(err)
	}
	outerr,err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return err
		//log.Fatal(err)
	}
	defer out.Close()
	defer outerr.Close()
	go runout(out)
	go runout(outerr)
	er :=  cmd.Run()
	if er != nil {
		panic(er)
	}
	return

}
func getRate(name string)(rateX,rateY,maxX,maxY float64){
//func getRate(name string)error{
	//adb shell getevent -p | grep -e "0035" -e "0036"
	var r1,r2 []float64
	err := cmdToAdb(func(s string)error{
		s_ := regMax.FindStringSubmatch(s)
		if len(s_)<2{
			return nil
		}
		r,err := strconv.ParseFloat(s_[1],64)
		if err != nil {
			return err
		}
		r1 = append(r1,r)
		return nil
	},"-s",name,"shell","getevent","-p","|","grep","-e","0035","-e","0036")
	if err != nil {
		return
		//panic(err)
	}
	if len(r1)!=2{
		return
	}
	//fmt.Println(r1)
	//adb shell wm size
	err = cmdToAdb(func(s string)error{
		s_ := regMax_.FindStringSubmatch(s)
		if len(s_)<3{
			return nil
		}
		//fmt.Println(s_)
		r,err := strconv.ParseFloat(s_[1],64)
		if err != nil {
			return err
		}
		r2 = append(r2,r)
		r,err = strconv.ParseFloat(s_[2],64)
		if err != nil {
			return err
		}
		r2 = append(r2,r)
		return nil
	},"-s",name,"shell","wm","size")
	if err != nil {
		return
		//panic(err)
	}
	//fmt.Println(r2)

	return r1[0]/r2[0],r1[1]/r2[1],r2[0],r2[1]
}
func inputSwipe(name,x1,y1,x2,y2 string) error {
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"shell", "input", "swipe",x1,y1,x2,y2)
}
func inputTap(name,x,y string) error {
	return cmdToAdb(func(s string)error{
		fmt.Println("tap",s)
		return nil
	},"-s",name,"shell", "input", "tap",x,y)
}
func inputKeyevent(name,key string) error {
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"shell", "input", "keyevent",key)
}
func inputText(name,text string) error {
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"shell", "input", "text",text)
}
func screencap(name string)error{
	cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"shell", "screencap", "/sdcard/screen.png")
	cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"pull", "/sdcard/screen.png")
	return nil
}

func connect(ip string) error {
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"connect",ip)
}
func tcpip(port string) error {
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"tcpip",port)
}
func devices() (ClientList []string){
	err := cmdToAdb(func(s string)error{
		s = regName.FindString(s)
		//fmt.Println(s,len(s))
		//if strings.HasPrefix(s,ip){
		if len(s)>0 && s != "List" {
			ClientList = append(ClientList,strings.Split(s," ")[0])
		}
			//fmt.Println(s)
		//}
		return nil
	},"devices")
	if err != nil {
		panic(err)
	}
	return ClientList
}

//func getScreencap(){
//	list := devices()
//	if len(list)==0{
//		return
//	}
//	fmt.Println(list)
//	and := list[0]
//	err := inputKeyevent(and,"26")
//	if err != nil {
//		panic(err)
//	}
//	time.Sleep(1*time.Second)
//	screencap(and)
//
//}
func main(){
	list := devices()
	if len(list)==0{
		return
	}
	fmt.Println(list)
	and := list[0]
	rx,ry,maxX,maxY := getRate(and)
	fmt.Println(rx,ry,maxX,maxY)
	err := inputKeyevent(and,"26")
	if err != nil {
		panic(err)
	}
	time.Sleep(1*time.Second)

	x1 := maxX/2*rx
	y1 := maxY/2*ry
	y2 := y1 - y1/2
	inputSwipe(and,fmt.Sprintf("%.0f",x1),fmt.Sprintf("%.0f",y1),fmt.Sprintf("%.0f",x1),fmt.Sprintf("%.0f",y2))

	time.Sleep(1*time.Second)

	//screencap(and)
	//860,1650,140
	fmt.Println("run")
	x := fmt.Sprintf("%.0f",float64(860+70)*rx)
	y := fmt.Sprintf("%.0f",float64(1650+70)*ry)
	inputTap(and,x,y)

	time.Sleep(1*time.Second)

	screencap(and)

}
