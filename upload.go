package main
import(
	"fmt"
	"os"
	"flag"
	"time"
	"strings"
	"net/url"
	"path/filepath"
	"github.com/zaddone/studySystem/wxmsg"
	//"github.com/zaddone/studySystem/config"
)

var (
	DirPath=flag.String("path", "/home/dimon/Music/pgMusic/", "dir path")
	Album = flag.String("album", "Album1", "Album name")
)
func init(){
	flag.Parse()
}
type fileInfo struct{
	path string
	name string
	upName string
	cid string
}
func (self *fileInfo) ToString() string {
	//dir := strings.Split(*DirPath,"/")
	return fmt.Sprintf(
		"\"%s|%s\"",
		url.QueryEscape(strings.Split(self.name,".")[0]),
		self.cid,
		//dir[len(dir)-1],
	)
}

func getFileList() ( fl []*fileInfo) {
	err := filepath.Walk(*DirPath,func(path string,_f os.FileInfo,err error)error{
		if err != nil {
			return err
		}
		if _f.IsDir(){
			return nil
		}
		if !strings.HasSuffix(_f.Name(),".mp3"){
			return nil
		}
		//fmt.Println(_f.Name())
		fl = append(
			fl,
			&fileInfo{
				path:path,
				name:_f.Name(),
				upName:fmt.Sprintf("%s/%d.mp3",*Album,time.Now().UnixNano()),
			})
		//_,fid,err := wxmsg.UpFileToWX_(path)
		//fmt.Println(fid)
		//f.WriteString(fid)
		//fmt.Println(err)
		return nil

	})
	if err != nil {
		panic(err)
	}
	return

}
func Up(fi []*fileInfo){

	newFi := make([]*fileInfo,0,len(fi))
	for _,f :=range fi {
		_,fid,err := wxmsg.UpFileToWX_(f.path,f.upName)
		if err != nil {
			newFi = append(newFi,f)
			continue
		}
		f.cid = fid
		fmt.Println(fid,f.name)
	}
	if len(newFi)>0 {
		Up(newFi)
	}

}

func main(){

	//f,err := os.OpenFile("updb.log",os.O_CREATE|os.O_APPEND|os.O_RDWR,0777)
	//if err != nil {
	//	panic(err)
	//}
	//defer f.Close()
	fi := getFileList()
	//fmt.Println(fi)
	//return
	Up(fi)
	str := make([]string,0,len(fi))
	for _,_f := range fi {
		str = append(str,_f.ToString())
	}
	//dir := strings.Split(*DirPath,"/")
	//db := fmt.Sprintf("{_id:\"%s\",list:[%s]}",*DirPath,strings.Join(str,","))
	//fmt.Println(db)
	err := wxmsg.AddToWXDB("music",fmt.Sprintf("{_id:\"%s\",list:[%s]}",*Album,strings.Join(str,",")))
	if err != nil {
		panic(err)
	}
	//wxmsg.UpFileToWX(uri)

}
