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
	Sumlist   int = 0
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
		if len(Public) > 0 || result["body"] == nil {
			return
		}
		Public = result["body"].(string)
		go func() {
			t := time.NewTicker(time.Second * 1)
			defer func() {
				fmt.Println("run goods end")
				t.Stop()
			}()
			fmt.Println("run goods")
			aft := time.After(15 * time.Second)
			for {
				select {
				case <-aft:
					fmt.Println("out Time")
					chromeServer.ClosePage()
					//defer t.Stop()
					return
				case <-stop:
					//t.Stop()
					return
				case <-t.C:
					chromeServer.InputKey(34, nil)
					//return
				}
			}
		}()

		//https://img.alicdn.com/tfscom
		//chromeServer.ClosePage()
	}) {
		return
	}
	chromeServer.GetBody(_db, "img.alicdn.com/tfscom", func(__id float64, result map[string]interface{}) {
		//fmt.Println(Public)
		if result["body"] == nil {
			return
		}
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
	//fmt.Println(no)
	goodsProps["pageNo"] = no

	props, err := json.Marshal(goodsProps)
	if err != nil {
		panic(err)
	}
	goodslistUrl_.Set("props", string(props))
	u := goodslistUrl + "?" + goodslistUrl_.Encode()
	fmt.Println(string(props))
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
			re__ := re_.(map[string]interface{})
			li := re__["list"]
			if li == nil {
				fmt.Println(re_)
				chromeServer.ClosePage()
				return
			}
			li_ := li.([]interface{})
			if len(li_) == 0 {
				chromeServer.ClosePage()
				return

			}
			Sumlist += len(li_)

			for _, l := range li_ {
				HandGoods(l)
			}
			if int(re__["total"].(float64)) < Sumlist {
				chromeServer.ClosePage()
				return
			}
			//fmt.Println(props)

			//chromeServer.ClosePage()
			//return

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
	//chromeServer.HandleResponse = GetGoodsAttr
	return chromeServer.Run(fmt.Sprintf("https://detail.1688.com/offer/%s.html", id), GetGoodsAttr)
}
func Run() error {
	Sumlist = 0
	//chromeServer.HandleResponse = GetGoodsList
	return chromeServer.Run(Url_2, GetGoodsList)
}
