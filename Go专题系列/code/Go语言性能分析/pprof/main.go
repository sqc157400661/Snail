package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // 第一步～
)

// 创建map不指定容量
func MakeMap1() map[int]int {
	mp := make(map[int]int)
	for i := 0; i < 100000; i++ {
		mp[i] = i
	}
	return mp
}

// 创建map指定容量
func MakeMap2() map[int]int {
	mp := make(map[int]int, 100000)
	for i := 0; i < 100000; i++ {
		mp[i] = i
	}
	return mp
}

func test1(w http.ResponseWriter, r *http.Request) {
	MakeMap1()
	MakeMap2()
	var ii interface{}
	ii = 1
	ii = "32342"
	fmt.Println(ii)
}

func main() {
	// 路由配置
	http.HandleFunc("/test1", test1)
	_ = http.ListenAndServe("0.0.0.0:6061", nil)
}
