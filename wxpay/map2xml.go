package main
import(
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"github.com/zaddone/studySystem/config"
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
	Coupon_type_0 string `xml:"coupon_type_0"`
	Coupon_id_0 string `xml:"coupon_id_0"`
	Coupon_fee_0 int `xml:"coupon_fee_0"`
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
		//fmt.Printf("%s: %v = %v\n", field.Name, field.Type, value)
		//fmt.Println(field.Tag)
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
	sort.Strings(sign_)
	sign := fmt.Sprintf("%s&key=%s",strings.Join(sign_,"&"),config.Conf.Apikeyv3)
	_sign := fmt.Sprintf("%X", md5.Sum([]byte(sign)))
	fmt.Println(sign,_sign,nsign)
	//fmt.Println(_sign,nsign)
	return strings.EqualFold(_sign,nsign)

}
type payRefundRes struct{
	XMLName   xml.Name `xml:"xml"`
	Return_code string `xml:"return_code"`
	Return_msg  string `xml:"return_msg"`
	Appid	    string `xml:"appid"`
	Mch_id      string `xml:"mch_id"`
	Nonce_str   string `xml:"nonce_str"`
	Openid      string `xml:"openid"`
	Sign	    string `xml:"sign"`
	Result_code string `xml:"result_code"`
	Transaction_id string `xml:"transaction_id"`
	Out_trade_no string `xml:"out_trade_no"`
	Out_refund_no string `xml:"out_refund_no"`
	Refund_id   string `xml:"refund_id"`
	Refund_fee   string `xml:"refund_fee"`
	Settlement_refund_fee   string `xml:"settlement_refund_fee"`
	Total_fee   string `xml:"total_fee"`
	Settlement_total_fee   string `xml:"settlement_total_fee"`
	Fee_type  string `xml:"fee_type"`
	Cash_fee  string `xml:"cash_type"`
	Cash_refund_fee  string `xml:"cash_refund_fee"`
	Coupon_refund_fee  string `xml:"coupon_refund_fee"`
	Coupon_refund_count  string `xml:"coupon_refund_count"`
	Coupon_type_0  string `xml:"coupon_type_0"`
	Coupon_refund_id_0  string `xml:"coupon_refund_id_0"`
	Coupon_refund_fee_0  string `xml:"coupon_refund_fee_0"`
	Err_code    string `xml:"err_code"`
	Err_code_des string `xml:"err_code_des"`
}

type unifiedRes struct {
	XMLName   xml.Name `xml:"xml"`
	Return_code string `xml:"return_code"`
	Return_msg  string `xml:"return_msg"`
	Appid	    string `xml:"appid"`
	Mch_id      string `xml:"mch_id"`
	Nonce_str   string `xml:"nonce_str"`
	Openid      string `xml:"openid"`
	Sign	    string `xml:"sign"`
	Result_code string `xml:"result_code"`
	Prepay_id   string `xml:"prepay_id"`
	Trade_type  string `xml:"trade_type"`
	Err_code    string `xml:"err_code"`
	Err_code_des string `xml:"err_code_des"`
	Device_info string `xml:"device_info"`
}
type Map map[string]interface{}

type xmlMapEntry struct {
    XMLName xml.Name
    Value   string `xml:",innerxml"`
}

type xmlMapEntryString struct {
    XMLName xml.Name
    Value   string `xml:",chardata"`
}
type xmlMapEntryInt struct {
    XMLName xml.Name
    Value   int `xml:",chardata"`
}
func (m Map) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
    if len(m) == 0 {
        return nil
    }
    start.Name = xml.Name{Local: "xml"}

    err := e.EncodeToken(start)
    if err != nil {
        return err
    }

    for k, v := range m {
	switch _v := v.(type){
	case string:
	    e.Encode(xmlMapEntryString{XMLName: xml.Name{Local: k}, Value: _v})
	case int:
	    e.Encode(xmlMapEntryInt{XMLName: xml.Name{Local: k}, Value: _v})
	}
    }

    return e.EncodeToken(start.End())
}

func (m *Map) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
    *m = Map{}
    for {
        var e xmlMapEntry
        err := d.Decode(&e)
        if err == io.EOF {
		break
        } else if err != nil {
		fmt.Println(err)
		return err
        }

        (*m)[e.XMLName.Local] = e.Value
    }
    return nil
}
func _main() {
    // The Map
    m := map[string]interface{}{
        "key_1": "Value One",
        "key_2": 1231231,
    }
    fmt.Println(m)
    // Encode to XML
    x, _ := xml.MarshalIndent(Map(m), "", "  ")
    fmt.Println(string(x))

    // Decode back from XML
    var rm map[string]interface{}
    xml.Unmarshal(x, (*Map)(&rm))
    fmt.Println(rm)
}
