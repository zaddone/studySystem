package main
import(
	//"gocv.io/x/gocv"
	"image"
	//"strings"
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
	CodeToPath = flag.String("Codepath","/sdcard/Pictures/Screenshots/xcode.png","codePath")
	Code = flag.String("Code","xcode.png","code")
	MainTag = flag.String("tag","./img/t1.png","tag")
	MainTag2 = flag.String("tag2","./img/t2.png","tag")
	MainTag3 = flag.String("tag3","./img/t3.png","tag")
	MainTag4 = flag.String("tag4","./img/t4.png","tag")
	Taobao = "com.taobao.taobao/com.taobao.tao.TBMainActivity"
)
func init(){
	flag.Parse()
	return
}
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
	fmt.Println(w,h)
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
func checkTag(tag string,po *image.Point,hand func(x,y int)) error {
	t := openImg(tag)
	var i int
	for i=0;i<10;i++{
		err := control.Screencap(control.MainPhone)
		if err != nil {
			return err
			//panic(err)
		}
		time.Sleep(time.Second)
		img := openImg(*MainImg)
		if img == nil{
			continue
			//return fmt.Errorf("img == nil")
		}
		//t.Bounds().Dx()/2
		if ContainsImg(img,t,po,0.95){
			w := t.Bounds().Dx()/2
			h := t.Bounds().Dy()/2
			hand(po.X+w,po.Y+h)
			//control.InputTap(control.MainPhone,)
			break
		}
	}
	if i>=10{
		return fmt.Errorf("i == 10")
	}
	return nil
}

func runApp(appact string)error{
	//err := control.CloseApp(control.MainPhone,app)
	//if err != nil {
	//	return err
	//	panic(err)
	//}
	err := control.Screencap(control.MainPhone)
	if err != nil {
		return err
		panic(err)
	}
	img := openImg(*MainImg)
	if img == nil{
		return fmt.Errorf("img == nil")
	}
	zero := getZeroColor()
	//t := img.At(10,10)
	if EqualColor(zero,img.At(10,10)){
		err := control.OpenScreen(control.MainPhone)
		if err != nil {
			return err
			panic(err)
		}
	}
	err = control.OpenApp(control.MainPhone,appact)
	if err != nil {
		return err
		panic(err)
	}
	return nil

}
func TaobaoLoginCheck(){
	err := control.Push(control.MainPhone,*Code,*CodeToPath)
	if err != nil {
		panic(err)
	}
	err = runApp(Taobao)
	if err != nil {
		panic(err)
	}
	p := image.Pt(32,112)
	err = checkTag(*MainTag,&p,func(x,y int){
		X,Y := control.GetTapXY(control.MainPhone,x,y)
		control.InputTap(control.MainPhone,X,Y)
	})
	if err != nil {
		panic(err)
	}
	p2 := image.Pt(120,1324)
	err = checkTag(*MainTag2,&p2,func(x,y int){
		X,Y := control.GetTapXY(control.MainPhone,x,y)
		control.InputTap(control.MainPhone,X,Y)
	})
	if err != nil {
		panic(err)
	}
	p3 := image.Pt(32,48)
	err = checkTag(*MainTag3,&p3,func(x,y int){
		time.Sleep(1*time.Second)
		X,Y := control.GetTapXY(control.MainPhone,192,1220)
		control.InputTap(control.MainPhone,X,Y)
	})
	if err != nil {
		panic(err)
	}
	p4 := image.Pt(120,48)
	err = checkTag(*MainTag4,&p4,func(x,y int){
		time.Sleep(1*time.Second)
		X,Y := control.GetTapXY(control.MainPhone,144,330)
		control.InputTap(control.MainPhone,X,Y)
	})
	if err != nil {
		panic(err)
	}
	time.Sleep(5*time.Second)
	err = control.Screencap(control.MainPhone)
	if err != nil {
		panic(err)
	}


}

