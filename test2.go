package main
import(
	"fmt"
	"regexp"
	"strings"
	//"runes"
)

var (
	reg *regexp.Regexp = regexp.MustCompile(`[\p{Han}]+`)
)

//func main_(){
//	fmt.Println(strings.IndexAny("widuu", "acbu"))
//	fmt.Println(strings.IndexAny("acbeu","widuu" ))
//}

func main(){
	m  := "在Go当中 string底层改变是用byte数组存的，并且是不可以改变的。"
	work := map[int][]bool{}
	li := reg.FindAllString(m,-1)
	for i,l := range li {
		work[i] = make([]bool,len([]rune(l)))
	}
	fmt.Println(li)
	le := len(li)
	//var list [][2]int
	for i:=0;i<le;i++ {
		s_bak := []rune(li[i])
		leb := len(s_bak)
		I := i+1
		_li := li[I:]
		if len(_li) == 0 {
			break
			//work[string(s_bak)] += 1
		}
		for j:=0;j<leb;j++{
		//for j,sk := range s_bak{
			sk := s_bak[j]
			for _i,_s := range _li {
				t := strings.IndexRune(_s,sk)
				//t = []rune(_s[:t])
				if t<=0 {
					continue
				}
				_j := len([]rune(_s[:t]))
				work[i][j] = true
				work[I+_i][_j] = true
				//work[i] = work[i],j)
				//work[_i] = append(work[_i],len([]rune(_s[:t])))
				fmt.Println(string([]rune(_s)[_j:]),string(s_bak[j:]))
				//j = j+_t-1
			}
		}
	}

	key := map[string]int{}
	for k,v := range work {
		fmt.Println(k,v,li[k])
		ls := []rune(li[k])
		var list []string
		var str string
		for i,l := range ls {
			if v[i] {
				if str != "" {
					list = append(list,str)
					str = ""
				}
				list = append(list,string(l))
			}else{
				str += string(l)
			}
		}
		if str != "" {
			list = append(list,str)
		}
		fmt.Println(list)
		le:= len(list)
		for i:=0;i<le;i++{
			s:= list[i]
			key[s]+=1
			for j:=i+1 ; j < le ; j++ {
				s+=list[j]
				key[s]+=1
			}
		}
	}
	for k,v := range key {
		fmt.Println(k,v)
	}
	//fmt.Println(li)
	return

	//strings.IndexRune(m)
	//return
	//m_ := []rune(m)

	//fmt.Println(m)
	//m1:="是否存在某个字符或子串"
	//li := reg.FindAllString(m,-1)
	//le := len(li)

	//for i:=0; i<le; i++ {
	//	s_bak := []rune(li[i])
	//	I := i+1
	//	_li := li[I:]
	//	if len(_li) > 0 {
	//		leb := len(s_bak)-1
	//		for j := leb ; j>=0 ; j-- {
	//			sk:= s_bak[j]
	//			//fmt.Println(string(sk),_li)
	//			//le_ := len(_li)
	//			//for _i:=0 ; _i<le_ ; _i++ {
	//			for _i,_s := range _li {
	//				t := strings.IndexRune(_s,sk)
	//				if t<=0 {
	//					continue
	//				}
	//				li[I+_i] = _s[:t]+string(sk)
	//				li = append(li,_s[t:])
	//				//_li = append(_li,_s[t:])
	//				//le++
	//				le = len(li)
	//				work[string(s_bak[j:])] +=1
	//				s_bak = append(s_bak[:j],sk)
	//			}
	//		}
	//	}
	//	work[string(s_bak)] +=1
	//}
	//for k,v := range work{
	//	fmt.Println(k,v)
	//}

}
