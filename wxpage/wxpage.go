package wxpage
import(
	"fmt"
	"io"
	//"io/ioutil"
	"github.com/zaddone/studySystem/request"
	"github.com/PuerkitoBio/goquery"
)
var(
	Wxurl="https://www.toutiao.com/ch/news_baby/"
)
func init(){
	downPage(Wxurl)
}
type page struct{


	Update	int64
	Title	string
	Text	string
	Source	string
	Similar	int64
	BeSimilar	[]int64

}
func downPage(u string){
	err := request.ClientHttp(u,"GET",[]int{304,200},nil,func(body io.Reader)error{
		doc,err := goquery.NewDocumentFromReader(body)
		if err != nil {
			//fmt.Println(err)
			return err
		}
		doc.Find(".txt-box h3 a").Each(func(i int, s *goquery.Selection){
			href,b := s.Attr("href")
			fmt.Println(s.Text(),href,b)
		})
		//d,err := ioutil.ReadAll(body)
		//fmt.Println(string(d))
		return nil
		//return json.NewDecoder(body).Decode(&db)
	})
	if err != nil {
		panic(err)
	}
}
