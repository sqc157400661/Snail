package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // 第一步～
	"runtime"
	"time"
)


func sleeper(ch chan int){
	ch <- 111
}
func read(ch chan int){
	<-ch
}

func test1(w http.ResponseWriter, r *http.Request){
	ch := make(chan int)
	for i:=0;i<999;i++ {
		go sleeper(ch)
		go read(ch)
	}
	go read(ch)
	time.Sleep(time.Second * 3)
	fmt.Println(111)
}

func init() {
	runtime.GOMAXPROCS(1) // 限制 CPU 使用数，避免过载
	//runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
	runtime.SetBlockProfileRate(1) // 开启对阻塞操作的跟踪
}
func main() {
	// 路由配置
	http.HandleFunc("/test1", test1)
	_ =http.ListenAndServe("0.0.0.0:6061", nil)
}













//