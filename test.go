package main
import(
	"encoding/csv"
	"strings"
	//"io/ioutil"
	//"bytes"
	"fmt"
	"os"
	"flag"
	"bufio"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)
var(
	path = flag.String("p","goods/20200727.csv","path")
	decoder *encoding.Decoder
)
func init(){
	decoder = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	//encoder = 
}

func loadTaobaoCsv() error{
	f,err := os.Open(*path)
	if err != nil {
		return err
	}
	defer f.Close()
	//cf := csv.NewReader(decoder.Reader(f))
	cf := csv.NewReader(f)
	for{
		r,err := cf.Read()
		fmt.Println(len(r),err)
		if err != nil {
			return err
		}
		fmt.Println(r)

		//for _,l := range r {
		//	fmt.Println(l)
		//}
	}

	return err
}

func loadCsv() error{

	f,err := os.Open(*path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(decoder.Reader(f))
	var dbval []byte
	line:=0
	var fields []string
	for{
		li,isp,err := buf.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}
		if isp{
			dbval = append(dbval,li...)
			//fmt.Println(len(dbval))
			continue
			//panic(err)
		}
		if len(dbval)>0 {
			li = append(dbval,li...)
			dbval = nil
		}
		//str, _ := decoder.Bytes(li)
		if line == 0{
			fields = strings.Fields(string(li))
			fmt.Println("fields",fields)
		}else if line<6 {
			//fields = strings.Fields(string(li))
			//fmt.Println(len(fields),fields)
			//str := strings.ReplaceAll(string(li),"\"\"","\\\"")
			lis := strings.Split(string(li),"	")
			for _i,l := range lis {
				fmt.Println(_i,l)
			}
		}else{
			break
		}
		line++
	}

	return nil

}
func main(){
	//err := loadTaobaoCsv()
	err := loadCsv()
	fmt.Println(err)
}
