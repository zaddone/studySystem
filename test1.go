package main

import (
	"fmt"
	"regexp"
)

func main() {

	st :=`播放类型：kuyun

    第01集$https://iqiyi.com-t-iqiyi.com/share/8cbe9ce23f42628c98f80fa0fac8b19a
    第02集$https://iqiyi.com-t-iqiyi.com/share/c457d7ae48d08a6b84bc0b1b9bd7d474
    第03集$https://iqiyi.com-t-iqiyi.com/share/0b96d81f0494fde5428c7aea243c9157

全选    
播放类型：ckm3u8

    第01集$https://iqiyi.com-t-iqiyi.com/20190717/4951_2d0beba8/index.m3u8
    第02集$https://iqiyi.com-t-iqiyi.com/20190719/4989_e422265d/index.m3u8
    第03集$https://iqiyi.com-t-iqiyi.com/20190724/5195_0501bf0e/index.m3u8

全选
   
影片下载：
下载类型：迅雷下载

    第01集$http://okxzy.xzokzyzy.com/20190717/4951_2d0beba8/Mr.临时老师01.mp4
    第02集$http://okxzy.xzokzyzy.com/20190719/4989_e422265d/临时工先生02.mp4
    第03集$http://okxzy.xzokzyzy.com/20190724/5195_0501bf0e/临时工先生03.mp4

全选
    `
	regS := regexp.MustCompile(`\S+\$\S+\.m3u8`)
	fmt.Println(regS.FindAllString(st,-1))



}
