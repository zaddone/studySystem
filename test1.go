package main

import (
	"fmt"
	"regexp"
)

func main() {

	str := "tsdfasfasdf title: 'afdasdfasdfasdfasdfasdfasdfasfd';asdfasfasdfasd"
	retitle := regexp.MustCompile(`title: \'([\S\s]+?)\'`)
	fmt.Println(retitle.FindAllStringSubmatch(str,-1)[0][1])


}
