package main

import (
	"fmt"
	//"regexp"
)

func main() {

	//str := "tsdfasfasdf title: 'afdasdfasdfasdfasdfasdfasdfasfd';asdfasfasdfasd"
	//retitle := regexp.MustCompile(`title: \'([\S\s]+?)\'`)
	//fmt.Println(retitle.FindAllStringSubmatch(str,-1)[0][1])
	str := map[string][]byte{}
	for i:=0;i<10;i++{
		str[fmt.Sprintf("%d",i)]=nil
	}
	for k,v:= range str {
		fmt.Println(k,v)
	}


}
