package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"  // 第一步～
)

// 一段有问题的代码
func do() {
	var c chan int
	for {
		select {
		case v := <-c:
			fmt.Printf("我是有问题的那一行，通道无法读取到值：%v", v)
		default:

		}
	}
}

func main() {
	// 执行一段有问题的代码
	for i := 0; i < 4; i++ {
		go do()
	}
	_ =http.ListenAndServe("0.0.0.0:6061", nil)
}
