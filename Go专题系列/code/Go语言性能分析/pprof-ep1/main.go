package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // 第一步～
)


func DoOrder(w http.ResponseWriter, r *http.Request) {
	user := getUser()
	goods := getGoods1()
	order := createOrder(user,goods,2)
	fmt.Println(order)
}

func main() {
	// 路由配置
	http.HandleFunc("/order-create", DoOrder)
	_ = http.ListenAndServe("0.0.0.0:6061", nil)
}
