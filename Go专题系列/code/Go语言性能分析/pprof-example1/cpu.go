package main

import (
	"net/http"
	_ "net/http/pprof" // 第一步～
)

func main() {
	// 路由配置
	http.HandleFunc("/cpu", myPrint)
	_ = http.ListenAndServe("0.0.0.0:6062", nil)
}

func myPrint(writer http.ResponseWriter, request *http.Request) {
	go func() {
		for i := 0; i < 10000000; i++ {
			//fmt.Println()
		}
	}()
	_, _ = writer.Write([]byte("cpu"))
}
