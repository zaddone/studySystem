package main
import(
)

type GoodsInfo struct {
	Product_id int `json:"product_id"`
	Product_name string `json:"product_name"`
	Product_type  string `json:"product_type"`
	Face_value float64 `json:"face_value"`
	Purchase_price float64 `json:"purchase_price"`
	Sales_status string `json:"sales_status"`
	Stock_status  string `json:"stock_status"`
	Template_id string `json:"template_id"`
	Details string `json:"details"`
	Template interface{}
}
type GoodsTemplate struct{
	AddressId string
	ElementInfo string
	IsServiceArea bool
	GameTempaltePreviewList string

}
type OrderInfo struct {
	Order_id string `json:"order_id"`
	Charge_finish_time string `json:"charge_finish_time"`
	Customer_order_no string `json:"customer_order_no"`
	Order_status string `json:"order_status"`
	Recharge_description string `json:"recharge_description"`
}
