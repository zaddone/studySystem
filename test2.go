package main
import(
	"fmt"
	"regexp"
	//"strings"
	//"runes"
)

var (
	reg *regexp.Regexp = regexp.MustCompile(`[0-9a-zA-Z]+|\p{Han}`)
	regT *regexp.Regexp = regexp.MustCompile(`[0-9|a-z|A-Z|\p{Han}]+`)
)
func main(){
	m  := "在Go当中 string底层改变是用byte数组存的 100t，并且是不可以改变的。"
	//work := map[int][]bool{}
	li := reg.FindAllString(m,-1)
	for i,l := range li {
		fmt.Println(i,l,len([]rune(l)))
	}
	return


}
