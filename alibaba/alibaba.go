package alibaba

import (
	"fmt"
	//"strconv"
	"github.com/zaddone/studySystem/chromeServer"
	//"github.com/zaddone/studySystem/shopping"
	"encoding/json"
	"net/url"
	"time"
)

var (
	Url_1         = "https://guanjia.1688.com/page/portal.htm"
	Url_2         = "https://guanjia.1688.com/page/start.htm"
	refUrl        = "https://guanjia.1688.com/page/offers.htm?menuCode=dx_offers"
	attrUrl       = "https://widget.1688.com/front/getJsonComponent.json"
	indexUrl      = "https://guanjia.1688.com/event/app/newchannel_fx_selloffer/querySuplierProducts.htm?_input_charset=utf8&keyword=&pageNum=1"
	goodslistUrl  = "https://widget.1688.com/front/getJsonComponent.json"
	goodslistUrl_ = url.Values{
		"props":      []string{"{\"loginId\":\"\",\"pageNo\":12,\"keyword\":\"\",\"memberId\":\"\",\"offerType\":\"normal\",\"supportDF\":true,\"supportNY\":false,\"supportCYS\":false}"},
		"namespace":  []string{"AlifeCsbcDxManagentOfferListActionsQueryOffers"},
		"widgetId":   []string{"AlifeCsbcDxManagentOfferListActionsQueryOffers"},
		"methodName": []string{"execute"},
	}
	goodsProps = map[string]interface{}{
		"loginId":    "",
		"pageNo":     0,
		"keyword":    "",
		"memberId":   "",
		"offerType":  "normal",
		"supportDF":  true,
		"supportNY":  false,
		"supportCYS": false,
	}
	HandGoods func(interface{}) = nil
	Public    string
	VideoInfo string
	stop      chan int
)

func GetGoodsAttr(_db interface{}) {
	///event/app/videoInfo/getVideoById.htm
	if chromeServer.GetBody(_db, "event/app/videoInfo/getVideoById.htm", func(__id float64, result map[string]interface{}) {
		VideoInfo = result["body"].(string)
		//fmt.Println(result["body"])

	}) {

		return
	}
	if chromeServer.GetBody(_db, "1688.com/offer/", func(__id float64, result map[string]interface{}) {
		if len(Public) == 0 {
			Public = result["body"].(string)
		}

		go func() {
			select {
			case <-time.After(1 * time.Second):
				chromeServer.InputKey(34, nil)
			case <-stop:
				return
			}
		}()

		//https://img.alicdn.com/tfscom
		//chromeServer.ClosePage()
	}) {
		return
	}
	chromeServer.GetBody(_db, "img.alicdn.com/tfscom", func(__id float64, result map[string]interface{}) {
		//fmt.Println(Public)

		HandGoods(map[string]string{
			"body":   Public,
			"tfscom": result["body"].(string),
			"video":  VideoInfo,
		})
		close(stop)
		chromeServer.ClosePage()
	})
}
func getPageUrl() string {
	no := goodsProps["pageNo"].(int)
	no++
	goodsProps["pageNo"] = no

	props, err := json.Marshal(goodsProps)
	if err != nil {
		panic(err)
	}
	goodslistUrl_.Set("props", string(props))
	u := goodslistUrl + "?" + goodslistUrl_.Encode()
	fmt.Println(u, string(props))
	return u

}
func GetGoodsList(_db interface{}) {
	if !chromeServer.GetBody(_db, "page/start.htm", func(__id float64, result map[string]interface{}) {

		chromeServer.PageNavigate_(getPageUrl(), refUrl, func(res map[string]interface{}) {
			fmt.Println(res)
		})

	}) {
		chromeServer.GetBody(_db, goodslistUrl, func(__id float64, result map[string]interface{}) {
			if HandGoods == nil {
				return
			}
			body := result["body"]
			if body == nil {
				chromeServer.ClosePage()
				return
			}
			var re map[string]interface{}
			err := json.Unmarshal([]byte(body.(string)), &re)
			if err != nil {
				fmt.Println(result)
				panic(err)
			}
			re_ := re["content"]
			if re_ == nil {
				fmt.Println(re)
				return
			}
			li := re_.(map[string]interface{})["list"]
			if li == nil {
				fmt.Println(re_)
				//chromeServer.ClosePage()
				return
			}
			li_ := li.([]interface{})
			if len(li_) == 0 {
				chromeServer.ClosePage()
				return

			}
			for _, l := range li_ {
				HandGoods(l)
			}

			chromeServer.ClosePage()
			return

			chromeServer.PageNavigate_(getPageUrl(), refUrl, func(res map[string]interface{}) {
				fmt.Println(res)
			})
		})
	}
}

func RunDetail(id string) error {
	Public = ""
	VideoInfo = ""
	stop = make(chan int)
	chromeServer.HandleResponse = GetGoodsAttr
	return chromeServer.Run(fmt.Sprintf("https://detail.1688.com/offer/%s.html", id))
}
func Run() error {
	chromeServer.HandleResponse = GetGoodsList
	return chromeServer.Run(Url_2)
}
