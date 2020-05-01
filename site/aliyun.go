package main
import(
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/zaddone/studySystem/config"
	"fmt"
	"time"
	"sync"
	"math/rand"
)
const (
	Num int = 4
)
var (
	CodeMap =  sync.Map{}
	rand_service_handler = rand_generator(10)
)
type phoneCodeStruct struct {
	p string
	stop chan bool
}

func rand_generator(n int) chan int {
	rand.Seed(time.Now().UnixNano())
	out := make(chan int)
	go func(x int) {
		for {
			out <- rand.Intn(x)
		}
	}(n)
	return out
}

func randCode() (s string) {
	for i:=0;i<Num;i++{
		s += string('0'+byte(<-rand_service_handler))
	}
	_,ok := CodeMap.Load(s)
	fmt.Println(ok)
	if ok {
		return randCode()
	}
	return

}
func CheckPhoneCode(phone,code string) error {
	return nil
	v,ok := CodeMap.Load(code)
	if !ok{
		return fmt.Errorf("code is error")
	}
	ph:=v.(phoneCodeStruct)
	if ph.p != phone {
		return fmt.Errorf("phone is error")
	}
	close(ph.stop)
	return nil
}
func PhoneCode(phone string) error {
	code := randCode()
	err := sendSms(phone,code)
	if err != nil {
		return err
	}
	s := make(chan bool)
	CodeMap.Store(code,phoneCodeStruct{p:phone,stop:s})
	go func(_s string,st chan bool){
		defer func(){
			CodeMap.Delete(_s)
			fmt.Println("del",_s)
		}()
		select{
		case <-time.After(5*time.Minute):
			return
		case <-st:
			return
		}

	}(code,s)
	return nil
}

func sendSms(phone,code string) error {
	client, err := dysmsapi.NewClientWithAccessKey("cn-zhangjiakou", config.Conf.AliyunKeyid, config.Conf.AliyunSecret)
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phone
	request.SignName = "米果推荐"
	request.TemplateCode = "SMS_187560736"
	request.TemplateParam = fmt.Sprintf("{\"code\":\"%s\"}",code)

	response, err := client.SendSms(request)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	if response.Code == "OK"{
		return nil
	}
	//fmt.Println(response)
	return fmt.Errorf(response.Message)
}
