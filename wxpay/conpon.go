package main
import(
	"net/http"
	"bytes"
	"fmt"
	"math/rand"
	"github.com/zaddone/studySystem/request"
	"github.com/zaddone/studySystem/config"
	_rand "crypto/rand"
	"encoding/json"
	"crypto/x509"
	"crypto/rsa"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha1"
	"crypto"
	"encoding/base64"
	"encoding/pem"
	"time"
	"os"
	"io/ioutil"
	"net/url"
	"io"
	"flag"
)
var (
	pemcert = flag.String("pemcert","cert/apiclient_cert.pem","pemcert")
	pemkey = flag.String("pemkey","cert/apiclient_key.pem","pemkey")
	//rand_str = rand.New(rand.NewSource(time.Now().Unix()))
	MerchantId = flag.String("m","1589104921","merchant")
	header = http.Header{}
	rand_str = rand.New(rand.NewSource(time.Now().Unix()))
	publicKey []byte
	rfcTime = "2006-01-02T15:04:05.000Z"
	//timeFormat = "2006010215"

)
func RandString(len int) string {
	_bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := rand_str.Intn(26) + 65
		_bytes[i] = byte(b)
	}
	return string(_bytes)
}

func init(){
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json")
	header.Set("User-Agent","zaddone")
	header.Set("Accept-Language","zh-CN")
	certificates()
}


func CheckSign(sign map[string]string) (bool, error) {
	timestamp := sign["timestamp"]
	nonce := sign["nonce"]
	signature := sign["signature"]
	body := sign["body"]
	//wxSerial := sign["wxSerial"]
	//验签之前需要先验证平台证书序列号是否正确一致
	//此处部分代码省略
	//if cert.SerialNo != wxSerial {
	//	glog.Error("证书号错误或已过期")
	//	return false, err
	//}
	checkStr := timestamp + "\n" + nonce + "\n" + body + "\n"

	//key, err := ioutil.ReadFile(*pemcert)
	//if err != nil {
	//	panic(err)
	//}

	blocks, _ := pem.Decode(publicKey)
	fmt.Printf("%+v",blocks)
	//if blocks == nil || blocks.Type != "PUBLIC KEY" {
	if blocks == nil {
		//fmt.Println(blocks.Type)
		//fmt.Println("failed to decode PUBLIC KEY")
		return false, nil
	}
	oldSign, err := base64.StdEncoding.DecodeString(signature)
	pub, err := x509.ParsePKIXPublicKey(blocks.Bytes)
	hashed := sha256.Sum256([]byte(checkStr))
	err = rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA256, hashed[:], oldSign)

	return true, err
}

func Sign(method, uri, body string) error {


	mchid := *MerchantId
	serial_no := "1FFC8ABAC3E9F21708F5FAA65EB05AD872D0175F"
	header.Add("Wechatpay-Serial",serial_no)
	nonce_str := RandString(16)
	timestamp := time.Now().Unix()
	signature, err := SHA256WithRsaBase64(fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n",method,uri,timestamp,nonce_str,body))
	if err != nil {
		return err
		//panic(err)
	}
	//signature = "test"
	header.Set("Authorization",fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",serial_no=\"%s\",nonce_str=\"%s\",timestamp=\"%d\",signature=\"%s\"",mchid,serial_no,nonce_str,timestamp,signature))
	//fmt.Println(signature)
	header.Set("Request-ID", nonce_str)
	//ok,err := CheckSign(map[string]string{
	//	"timestamp":fmt.Sprintf("%d",timestamp),
	//	"nonce":nonce_str,
	//	"signature":signature,
	//	"body":body,
	//})
	//fmt.Println(ok,err)
	//fmt.Println(header)
	//return header

	return nil
}
func RsaEncrypt(origData []byte) (string, error) {
	//publicKey := []byte(`平台证书公钥`)
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return "",fmt.Errorf("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	secretMessage := origData
	rng := _rand.Reader

	cipherdata, err := rsa.EncryptOAEP(sha1.New(), rng, pubInterface.(*rsa.PublicKey), secretMessage, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
	}

	ciphertext := base64.StdEncoding.EncodeToString(cipherdata)
	fmt.Printf("Ciphertext: %x\n", ciphertext)
	return ciphertext, err
}
//func AesDecrypt(ciphertext string) string {
//	key := []byte(conf.AppConf.MinKey) // 加密的密钥
//	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
//	if err != nil {
//		fmt.Println(err)
//		return ""
//	}
//
//	genKey := make([]byte, 16)
//	copy(genKey, key)
//	for i := 16; i < len(key); { //		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
//			genKey[j] ^= key[i]
//		}
//	}
//
//	cipher, _ := aes.NewCipher(genKey)
//	decrypted := make([]byte, len(encrypted))
//
//	for bs, be := 0, cipher.BlockSize(); bs < len(encrypted); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
//		cipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
//	}
//
//	trim := 0
//	if len(decrypted) > 0 {
//		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
//	}
//
//	decrypted = decrypted[:trim]
//
//	log.Println("解密结果：", string(decrypted))
//	return string(decrypted)
//}

//func RsaDecrypt(ciphertext, nonce2, associatedData2 string) (plaintext string, err error) {
func RsaDecrypt(cert map[string]interface{}) ( err error) {
	key := []byte(config.Conf.Apikeyv3)
	additionalData := []byte(cert["associated_data"].(string))
	nonce := []byte(cert["nonce"].(string))

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		return err
	}
	cipherdata, _ := base64.StdEncoding.DecodeString(cert["ciphertext"].(string))
	//if err != nil {
	//	panic(err)
	//}
	publicKey, err = aesgcm.Open(nil, nonce, cipherdata, additionalData)
	//fmt.Println("plaintext: ", string(plaindata))
	return err
	//return string(plaindata), err

}

func SHA256WithRsaBase64(origData string) (sign string, err error) {
	key, err := ioutil.ReadFile(*pemkey)
	//fmt.Println(key)
	blocks, _ := pem.Decode(key)
	if blocks == nil || blocks.Type != "PRIVATE KEY" {
		fmt.Println("failed to decode PRIVATE KEY")
		return
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(blocks.Bytes)
	if err != nil {
		panic(err)
	}
	//fmt.Println(privateKey)

	h := sha256.New()
	h.Write([]byte(origData))
	//digest := h.Sum([]byte("miguotuijian2020miguotuijian2020"))
	digest := h.Sum(nil)
	s, _ := rsa.SignPKCS1v15(nil, privateKey.(*rsa.PrivateKey), crypto.SHA256, digest)
	sign = base64.StdEncoding.EncodeToString(s)
	return sign, err
}
func getRequestNo(mid string) string {
	return fmt.Sprintf("%s_%d",mid,time.Now().Unix())
}
func couponCreate(amount int,hand func(interface{})error)error{
	uri,err := url.Parse("https://api.mch.weixin.qq.com/v3/marketing/favor/coupon-stocks")
	if err != nil {
		return err
	}
	begin :=time.Now()
	end :=  begin.Add(36*time.Hour)
	//end :=  begin.AddDate(0,0,2)
	//am := amount*100
	body := map[string]interface{}{
		"stock_name":fmt.Sprintf("%.2f元代金券",float64(amount)/100),
		"belong_merchant":*MerchantId,
		"available_begin_time":begin.Format(rfcTime),
		"available_end_time":end.Format(rfcTime),
		"no_cash":true,
		"stock_type":"NORMAL",
		"out_request_no":getRequestNo(*MerchantId),
		"stock_use_rule":map[string]interface{}{
			"max_coupons":1,
			"max_amount":amount,
			"max_coupons_per_user":1,
			"prevent_api_abuse":true,
			"natural_person_limit":false,
		},
		"coupon_use_rule":map[string]interface{}{
			//"coupon_available_time":map[string]interface{}{
			//	//"fix_available_time":map[string]interface{}{
			//	//	"begin_time":beginDay,
			//	//	"end_time":beginDay+7200,
			//	//},
			//	"available_time_after_receive":120,
			//	//"second_day_available":false,
			//},
			"fixed_normal_coupon":map[string]interface{}{
				"coupon_amount":amount,
				"transaction_minimum":amount,
			},
			"available_merchants":[]string{*MerchantId},
		},
	}
	str,err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = Sign("POST",uri.Path,string(str))
	if err != nil {
		panic(err)
	}
	var db map[string]interface{}
	err = request.ClientHttp_(
		uri.String(),
		"POST",
		bytes.NewReader(str),
		header,
		func(res io.Reader,st int)error{
			if st != 200 {
				_db,_ := ioutil.ReadAll(res)
				fmt.Println(body,_db)
				return fmt.Errorf("%s",_db)
			}
			return json.NewDecoder(res).Decode(&db)
		},
	)
	if err != nil {
		return err
	}
	body["req"] = db
	return hand(body)


}

func couponGet(stockid,uid,appid,requestNo string,amount int) error {
	uri,err :=url.Parse(fmt.Sprintf("https://api.mch.weixin.qq.com/v3/marketing/favor/users/%s/coupons",stockid))
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"stock_id":stockid,
		"openid":uid,
		"out_request_no":requestNo,
		"appid":appid,
		"stock_creator_mchid":*MerchantId,
		"coupon_value":amount,
		"coupon_minimum":amount,
	}
	str,err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = Sign("POST",uri.Path,string(str))
	if err != nil {
		panic(err)
	}
	return request.ClientHttp_(
		uri.String(),
		"POST",
		bytes.NewReader(str),
		header,
		func(res io.Reader,st int)error{
			if st != 200 {
				db,_ := ioutil.ReadAll(res)
				return fmt.Errorf("%s",db)
			}
			return nil
			//return json.NewDecoder(res).Decode(&db)
		},
	)

}

func couponOpen(stock_id string) error {
	uri,err :=url.Parse(fmt.Sprintf("https://api.mch.weixin.qq.com/v3/marketing/favor/stocks/%s/start",stock_id))
	body := map[string]interface{}{
		"stock_creator_mchid":*MerchantId,
	}
	str,err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = Sign("POST",uri.Path,string(str))
	if err != nil {
		return err
	}
	return  request.ClientHttp_(
		uri.String(),
		"POST",
		bytes.NewReader(str),
		header,
		func(res io.Reader,st int)error{
			if st != 200 {
				db,_ := ioutil.ReadAll(res)
				return fmt.Errorf("%s",db)
			}
			return nil
			//return json.NewDecoder(res).Decode(&db)
		},
	)
}

func certificates() error {
	//https://api.mch.weixin.qq.com/v3/certificates
	uri,err := url.Parse("https://api.mch.weixin.qq.com/v3/certificates")
	if err != nil {
		return err
	}
	err = Sign("GET",uri.Path,"")
	if err != nil {
		return err
	}
	return request.ClientHttp_(
		uri.String(),
		"GET",
		nil,
		header,
		func(res io.Reader,st int)error{
			if st != 200 {
				return fmt.Errorf("request is not 200")
			}
			var db map[string]interface{}
			err := json.NewDecoder(res).Decode(&db)
			if err != nil {
				return err
			}
			//db["data"].(map[string]interface{})
			var lastDB map[string]interface{}
			for _,d := range db["data"].([]interface{}){
				lastDB = d.(map[string]interface{})
			}
			cert :=lastDB["encrypt_certificate"].(map[string]interface{})
			RsaDecrypt(cert)
			//fmt.Println()
			return nil
		},
	)

}
func callbacks(){
	uri,err := url.Parse("https://api.mch.weixin.qq.com/v3/marketing/favor/callbacks")
	if err != nil {
		panic(err)
	}
	body := map[string]interface{}{
		"mchid":*MerchantId,
		"notify_url":"https://www.zaddone.com/wxpay",
		"switch":true,
	}
	str,err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	err = Sign("POST",uri.Path,string(str))
	if err != nil {
		panic(err)
	}
	err = request.ClientHttp_(
		uri.String(),
		"POST",
		bytes.NewReader(str),
		header,
		func(res io.Reader,st int)error{
			db,err := ioutil.ReadAll(res)
			fmt.Println(st)
			fmt.Println(string(db),err)
			return nil
		},
	)
	fmt.Println(err)
}

func stockslist(){
	//https://api.mch.weixin.qq.com/v3/marketing/favor/stocks
	u:=&url.Values{}
	u.Set("offset","0")
	u.Set("limit","10")
	u.Set("stock_creator_mchid",*MerchantId)
	uri,err := url.Parse("https://api.mch.weixin.qq.com/v3/marketing/favor/stocks?"+u.Encode())
	if err != nil {
		panic(err)
	}
	err = Sign("GET",uri.Path+"?"+u.Encode(),"")
	if err != nil {
		panic(err)
	}
	err = request.ClientHttp_(
		uri.String(),
		"GET",
		nil,
		header,
		func(res io.Reader,st int)error{
			db,err := ioutil.ReadAll(res)
			fmt.Println(st)
			fmt.Println(string(db),err)
			return nil
		},
	)
	//fmt.Println(uri,uri.)
}
