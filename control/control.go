package control
import(
	"bufio"
	"fmt"
	"os/exec"
	"log"
	"io"
	"strings"
	"strconv"
	"regexp"
	"flag"
	//"time"
)

var(
	//ip string = "192.168.1."
	//port string ="5555"
	regName = regexp.MustCompile(`([0-9a-zA-Z\.\:]+)`)
	regMax = regexp.MustCompile(`max (\d+),`)
	regMax_ = regexp.MustCompile(`(\d+)x(\d+)`)
	//InitIp = flag.String("ip","192.168.1.","ip")
	Port = flag.Int("port",5555,"port")
	//MainPhone string
	Screen  = "screen.png"
)
func init(){
	flag.Parse()
	return

}

func InitDevices(ip string) error{
	li := devices()
	fmt.Println(li)
	for _,l := range li {
		if !strings.Contains(l,"192.168.1.") {
			err:= tcpip(l,*Port)
			if err != nil {
				return err
			}
			fmt.Println(l)
			return connect(l,ip)
		}
	}
	return io.EOF
}
func Connect(name,ip string)error{
	err:= tcpip(name,*Port)
	if err != nil {
		return err
	}
	//fmt.Println(l)
	return connect(name,ip)
}

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
	//fmt.Println(opt)
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
	return cmd.Run()
	//if er != nil {
	//	//return er
	//	panic(er)
	//}
	//return

}
func GetRate(name string)(X,Y,W,H float64){
	return getRate(name)
}
func GetTapXY(name string,x,y int)(X,Y string){
	rateX,rateY,_,_:= getRate(name)
	X = fmt.Sprintf("%.0f",float64(x)*rateX)
	Y = fmt.Sprintf("%.0f",float64(y)*rateY)
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
func InputTap(name,x,y string) error {
	return inputTap(name,x,y)
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
func InputText(name,text string) error {
	return inputText(name,text)
}
func inputText(name,text string) error {
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"shell", "input", "text",text)
}
func Screencap(name,img string)error{
	cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"shell", "screencap", "/sdcard/"+img)
	cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"pull", "/sdcard/"+img)
	return nil
}

func Push(name ,f ,t string)error{
	return push(name,f,t)
}
func push(name ,f ,t string)error{
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"push", f, t)
}

func connect(name,ip string) error {
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"connect",ip)
}
func tcpip(name string,port int) error {
	return cmdToAdb(func(s string)error{
		fmt.Println(s)
		return nil
	},"-s",name,"tcpip",fmt.Sprintf("%d",port))
}

func OpenPower(name string)error{
	return inputKeyevent(name,"26")

}
func OpenScreen(name string)error{
	//err := inputKeyevent(name,"26")
	//if err != nil {
	//	return err
	//}
	//time.Sleep(1*time.Second)
	rx,ry,maxX,maxY := getRate(name)
	x1 := maxX/2*rx
	y1 := maxY*0.9*ry
	y2 := y1 - maxY/2
	err := inputSwipe(name,fmt.Sprintf("%.0f",x1),fmt.Sprintf("%.0f",y1),fmt.Sprintf("%.0f",x1),fmt.Sprintf("%.0f",y2))
	if err != nil {
		return err
	}

	//time.Sleep(1*time.Second)
	return nil
}
func OpenApp(name,app string) error{
	return cmdToAdb(func(s_ string)error{
		fmt.Println(s_)
		return nil
	},"-s",name,"shell", "am", "start",app)
}

func CloseApp(name,app string) error{
	//fmt.Println("close",app)
	return cmdToAdb(func(s_ string)error{
		fmt.Println(s_)
		return nil
	},"-s",name,"shell", "am", "force-stop",app)
}
func Devices() (ClientList []string){
	return devices()
}
func devices() (ClientList []string){
	err := cmdToAdb(func(s_ string)error{
		ss := regName.FindAllString(s_,-1)
		if len(ss)!=2 {
			return nil
		}
		if strings.EqualFold(ss[1],"device"){
			ClientList = append(ClientList,ss[0])
		}
		return nil
		//s := regName.FindString(s_)
		//fmt.Println(s,len(s))
		//if strings.HasPrefix(s,ip){
		//if s[0] != "List" {
		//	fmt.Println(s_)
		//	ClientList = append(ClientList,strings.Split(s," ")[0])
		//}
			//fmt.Println(s)
		//}
		//return nil
	},"devices")
	if err != nil {
		fmt.Println(err)
		//panic(err)
	}
	return ClientList
}
