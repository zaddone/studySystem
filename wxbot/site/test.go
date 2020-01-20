package main
import(
	//"unicode/utf8"
	"github.com/boltdb/bolt"
	"encoding/binary"
	//"net/url"
	//"strconv"
	//"strings"
	"unsafe"
	//"bytes"
	"fmt"
	//"strconv"
)
var (
	Contact = []byte("ContactList")
)
func BytesString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func ByteToBstring(b []byte)[]string{
	s := make([]string,0,len(b))
	//var m,k byte
	var j uint
	for _,_b := range b{
		var str string = ""
		for j=0;j<8;j++{
			if _b &^ (^(1<<j)) == 0 {
				str="0"+str
			}else{
				str="1"+str
			}
		}
		s = append(s,str)
	}
	return s
}

func toDec(db []byte) []byte {

	var _db []byte
	for _, r := range []rune(string(db)){
		//fmt.Println(r)
		//b_ := make([]byte,8)
		r_ := string(r)
		le := len(r_)
		//fmt.Println(len(r_))
		if le == 1 {
			_db = append(_db,[]byte(r_)...)
			continue
		}

		b__ := make([]byte,8)
		//binary.LittleEndian.PutUint64(b_,uint64(r))
		binary.BigEndian.PutUint64(b__,uint64(r))
		for i,b := range b__{
			if b==0{
				continue
			}
			b__ = b__[i:]
			break
		}
		//fmt.Println(b_)
		//fmt.Println(b__)
		_db = append(_db,b__...)
		//v := binary.BigEndian.Uint64([]byte(string(r)))
		//fmt.Println(r,v)
	}
	return _db //fmt.Println(string(_db))

}


func main(){
	//fmt.Println([]byte)
	WXDB,err := bolt.Open("WXDB",0600,nil)
	if err != nil {
		panic(err)
	}
	err = WXDB.View(func(t *bolt.Tx)error{
		b := t.Bucket(Contact)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k,v []byte)error{
			k_:= toDec([]byte(k))
			k_s := ByteToBstring(k_)
			fmt.Println(string(k_))
			fmt.Println(k_s)
			return nil
		})

	})
	if err != nil {
		panic(err)
	}
}
