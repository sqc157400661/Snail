package main

import "time"

// Orders ...
type Orders struct {
	ID              string    `json:"id"`
	Money           float64   `json:"money"`
	GoodsId         int       `json:"goods_id"`
	ReceiverAddress string    `json:"receiverAddress"`
	ReceiverName    string    `json:"receiverName"`
	ReceiverPhone   string    `json:"receiverPhone"`
	Paystate        int32     `json:"paystate"`
	Ordertime       time.Time `json:"ordertime"`
	UserID          int       `json:"user_id"`
}

/*
	创建订单方法
*/
func createOrder(user User, goods Goods, num float64) Orders {
	money := goods.GoodsPrice * num
	return Orders{
		Money:   money,
		GoodsId: goods.ID,
		UserID:  user.ID,
	}
}
