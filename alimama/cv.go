package alimama
import(
	//"gocv.io/x/gocv"
	"image"
	"strings"
	"fmt"
	"time"
	//"io"
	"os"
	"image/png"
	"github.com/zaddone/studySystem/control"
	"image/color"
	"flag"
)
var(
	MainImg = flag.String("img","screen.png","img")
	ShowImg = flag.String("img_","screen_1.png","img")
	CodeToPath = flag.String("Codepath","/sdcard/Pictures/Screenshots/xcode.png","codePath")
	Code = flag.String("Code","xcode.png","code")
	MainTag = flag.String("tag","./img/t1.png","tag")
	MainTag2 = flag.String("tag2","./img/t2.png","tag")
	MainTag3 = flag.String("tag3","./img/t3.png","tag")
	MainTag4 = flag.String("tag4","./img/t4.png","tag")
	Taobao = "com.taobao.taobao/com.taobao.tao.TBMainActivity"
	Browser = "com.android.browser/.BrowserActivity"
	LoginPhone = flag.String("Login","192.168.1.51","ip")
	ShowPhone = flag.String("Show","192.168.1.52","ip")
)
//func init(){
//	flag.Parse()
//}
func InitPhone(phone string,stop chan bool,hand func(string)) (err error){
	for{
		select{
		case <-stop:
			return
		default:

			li := control.Devices()
			for _,l := range li {
				if strings.Contains(l,phone){
					//cphone = l
					hand(l)
					return
					//continue
				}
			}
			for _,l := range li {
				if !strings.Contains(l,"192.168.1.") {
					err = control.Connect(l,phone)
					if err != nil {
						return
					}
				}
			}
			time.Sleep(time.Second)
		}
	}
	return
	//return InitPhone(phone)
}
//func main(){
//	fmt.Println(InitPhone(*ShowPhone))
//}
func openImg(src string) image.Image {
	//fi,err := os.Stat(src)
	//if err != nil {
	//	return nil
	//	panic(err)
	//}
	//if fi.Size() == 0 {
	//	fmt.Println("size=0")
	//	return nil
	//}
	f,err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return nil
		//panic(err)
	}
	defer f.Close()
	//png.Decode(
	img,err := png.Decode(f)
	if err != nil {
		fmt.Println(err)
		return nil
		//panic(err)
	}
	return img
}
func ContainsImg(img,subimg image.Image,point *image.Point,rale float64) bool {
	var k,v float64
	w := subimg.Bounds().Dx()
	h := subimg.Bounds().Dy()
	w_ := img.Bounds().Dx()
	h_ := img.Bounds().Dy()
	//fmt.Println(w,h)
	for i,I,j,J:=0,point.X,0,point.Y ; i<w && j<h && I<w_ && J<h_ ; i,I,j,J = i+1,I+1,j+1,J+1{
		k++
		if EqualColor(img.At(I,J),subimg.At(i,j)){
		v++
		}
	}
	return  (v/k) >= rale
}
func EqualColor(src,dis color.Color) bool{
	g,b,a,r := src.RGBA()
	g_,b_,a_,r_ := dis.RGBA()
	fmt.Println(g_,b_,a_,r_ )
	if g != g_ {
		return false
	}
	if b != b_ {
		return false
	}
	if a != a_ {
		return false
	}
	if r != r_ {
		return false
	}
	return true
}
func getZeroColor() color.Color {
	r:= uint8(255)
	z:=uint8(0)
	return color.RGBA{z,z,z,r}
}
func TsTap(name string)error{
	var i int
	max := 100
	for i=0;i<max;i++{
		err := control.Screencap(name,*MainImg)
		if err != nil {
			return err
			//panic(err)
		}
		//time.Sleep(time.Second)
		fi,err := os.Stat(*MainImg)
		if err != nil {
			return err
		}
		if fi.Size() > 0 {
			continue
		}
		X,Y,W,H := control.GetRate(name)
		control.InputTap(name,fmt.Sprintf("%.0f",W*0.5*X),fmt.Sprintf("%.0f",H*0.67 *Y))
		return nil
	}
	if i == max{
		return fmt.Errorf("i == 100")
	}
	return nil
}
func checkTag(name,tag,img string,po *image.Point,hand func(x,y int)) error {
	t := openImg(tag)
	var i int
	var max int =100
	for i=0;i<max;i++{
		err := control.Screencap(name,img)
		if err != nil {
			return err
			//panic(err)
		}
		//time.Sleep(time.Second)
		img_ := openImg(img)
		if img_ == nil{
			continue
			//return fmt.Errorf("img == nil")
		}
		//t.Bounds().Dx()/2
		if ContainsImg(img_,t,po,0.95){
			w := t.Bounds().Dx()/2
			h := t.Bounds().Dy()/2
			hand(po.X+w,po.Y+h)
			//control.InputTap(control.MainPhone,)
			break
		}
	}
	if i == max{
		return fmt.Errorf("i == 100")
	}
	return nil
}
func checkScreenIsEm(at color.Color) bool{
	g,b,a,_ := at.RGBA()
	return g+b+a == 0
}

func runApp(appact,name,img string)error{
	err := control.CloseApp(name,strings.Split(appact,"/")[0])
	if err != nil {
		return err
		panic(err)
	}
	err = control.Screencap(name,img)
	if err != nil {
		return err
	}
	img_ := openImg(img)
	//if img == nil{
	//	return fmt.Errorf("img == nil")
	//}
	//zero := getZeroColor()
	//t := img.At(10,10)
	if img_ == nil || checkScreenIsEm(img_.At(10,10)){
		//fmt.Println(img.At)
		err := control.OpenPower(name)
		if err != nil {
			return err
		}
	}
	err = control.OpenScreen(name)
	if err != nil {
		return err
	}
	err = control.OpenApp(name,appact)
	if err != nil {
		return err
	}
	return nil

}
func TaobaoLoginCheck(loginPhone string){
	err := runApp(Taobao,loginPhone,*MainImg)
	if err != nil {
		panic(err)
	}
	p := image.Pt(32,112)
	err = checkTag(loginPhone,*MainTag,*MainImg,&p,func(x,y int){
		X,Y := control.GetTapXY(loginPhone,x,y)
		control.InputTap(loginPhone,X,Y)
	})
	if err != nil {
		panic(err)
	}
	err = TsTap(loginPhone)
	if err != nil {
		panic(err)
	}
	err = control.CloseApp(loginPhone,strings.Split(Taobao,"/")[0])
	if err != nil {
		panic(err)
	}
	//p3 := image.Pt(32,48)
	//err = checkTag(loginPhone,*MainTag3,&p3,func(x,y int){
	//	//time.Sleep(1*time.Second)
	//	//X,Y := control.GetTapXY(loginPhone,192,1220)
	//	//control.InputTap(loginPhone,X,Y)
	//})
	//if err != nil {
	//	panic(err)
	//}
	return
}
func ShowBrowser(name string){
	err := runApp(Browser,name,*ShowImg)
	if err != nil {
		panic(err)
	}
	//control.InputText(name,"http://192.168.1.30:8001:/xcode.png")

}
func main(){
	fmt.Println("run")
	RunConTrol := make(chan bool)
	if err := InitPhone(*ShowPhone,RunConTrol,func(p string){
		fmt.Println(p)
		ShowBrowser(p)
	});err != nil {
		panic(err)
	}
}
