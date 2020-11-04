package main

import (
	"net/http"
	_ "net/http/pprof" // 第一步～
	"runtime"
)

func test1(w http.ResponseWriter, r *http.Request){
	var ch chan int = make(chan int)
	for i:=0;i<999;i++ {
		go func(i int) {
			ch<-i
		}(i)
	}

}

func init() {
	runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
	runtime.SetBlockProfileRate(1)
}
func main() {
	// 路由配置
	http.HandleFunc("/test1", test1)
	_ =http.ListenAndServe("0.0.0.0:6061", nil)
}













//