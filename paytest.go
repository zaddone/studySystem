package main
import(
	"fmt"
	"reflect"
	"strings"
	"crypto/md5"
)
type notify struct{
	Return_code string `xml:"return_code"`
	Return_msg  string `xml:"return_msg"`
	Appid       string `xml:"appid"`
	Mch_id      string `xml:"mch_id"`
	Device_info string `xml:"device_info"`
	Nonce_str   string `xml:"nonce_str"`
	Sign        string `xml:"sign"`
	Sign_type   string `xml:"sign_type"`
	Result_code string `xml:"result_code"`
	Err_code    string `xml:"err_code"`
	Err_code_des string `xml:"err_code_des"`
	Openid      string `xml:"openid"`
	Is_subscribe string `xml:"is_subscribe"`
	Trade_type  string `xml:"trade_type"`
	Bank_type   string `xml:"bank_type"`
	Total_fee   int `xml:"total_fee"`
	Settlement_total_fee int `xml:"settlement_total_fee"`
	Fee_type string `xml:"fee_type"`
	Cash_fee int `xml:"cash_fee"`
	Cash_fee_type string `xml:"cash_fee_type"`
	Coupon_fee   int  `xml:"coupon_fee"`
	Coupon_count int `xml:"coupon_count"`
	Coupon_type_n string `xml:"coupon_type_$n"`
	Coupon_id_n string `xml:"coupon_id_$n"`
	Coupon_fee_n int `xml:"coupon_fee_$n"`
	Transaction_id string `xml:"transaction_id"`
	Out_trade_no string `xml:"out_trade_no"`
	Attach string `xml:"attach"`
	Time_end string `xml:"time_end"`
}
func (self notify) CheckSign() bool {
	getType := reflect.TypeOf(self)
	getValue := reflect.ValueOf(self)
	sign_:=[]string{}
	var nsign string
	for i := 0; i < getType.NumField(); i++ {
		field := getType.Field(i)
		value := getValue.Field(i).Interface()
		if strings.EqualFold(field.Name,"Sign"){
			nsign = value.(string)
			continue
		}
		//fmt.Println(field.Tag.Get("xml"))
		//fmt.Printf("%s: %v = %v\n", field.Name, field.Type, value)
		switch v := value.(type) {
		case int:
			if v != 0{
				sign_ = append(sign_,fmt.Sprintf("%s=%d",field.Tag.Get("xml"),v))
			}
		case string:
			if len(v) >0 {
				sign_ = append(sign_,fmt.Sprintf("%s=%s",field.Tag.Get("xml"),v))
			}
		}
	}
	sign := fmt.Sprintf("%s&key=12313131231",strings.Join(sign_,"&"))
	_sign := fmt.Sprintf("%X", md5.Sum([]byte(sign)))
	fmt.Println(sign,_sign,nsign)
	//fmt.Println(_sign,nsign)
	return strings.EqualFold(_sign,nsign)

}
func main(){
	no:= &notify{}
	no.Appid="abcde"
	no.Cash_fee = 100
	fmt.Println(no.CheckSign())
}
