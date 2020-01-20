package util
import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"encoding/binary"
	//"unicode/utf8"
)

type Charset string

const (
   UTF8    = Charset("UTF-8")
   GB18030 = Charset("GB18030")
)
func ToDec(db []byte) []byte {

	var _db []byte
	for _, r := range []rune(string(db)){
		b__ := make([]byte,8)
		binary.BigEndian.PutUint64(b__,uint64(r))
		for i,b := range b__{
			if b==0{
				continue
			}
			b__ = b__[i:]
			break
		}
		_db = append(_db,b__...)
	}
	return _db //fmt.Println(string(_db))

}

func ConvertByte2String(b []byte, charset Charset) string {

   var str string
   switch charset {
   case GB18030:
      var decodeBytes,_=simplifiedchinese.GB18030.NewDecoder().Bytes(b)
      str= string(decodeBytes)
   case UTF8:
      fallthrough
   default:
      str = string(b)
   }

   return str
}
//func GetDOMBody(DOM interface{}) (body interface{}) {
//
//	//DOM.describeNode
//	root := ((__db.(map[string]interface{})["result"]).(map[string]interface{})["root"]).(map[string]interface{})
//	return (root["children"].([]interface{})[1].(map[string]interface{}))["children"].([]interface{})[2]
//
//}
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
