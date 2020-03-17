package shopping
import(
	//"github.com/boltdb/bolt"
)

type Order struct {
	OrderId string `json:"order_id"`
	GoodsId string `json:"goodsid"`
	UserId string `json:"userid"`
	GoodsName string `json:"goodsName"`
	Status bool `json:"status"`
	Fee float64 `json:"fee"`
	Site string `json:"site"`
	EndTime int64 `json:"endTime"`
	Time int64 `json:"time"`
	Text string `json:"text"`
	PayTime int64 `json:"payTime"`
}

