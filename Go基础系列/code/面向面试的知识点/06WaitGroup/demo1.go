package main

import (
	"sync"
)

const N = 10

var wg = &sync.WaitGroup{}


func main() {

	for i:= 0; i< N; i++ {
		wg.Add(1)
		go func(i int) {
			//wg.Add(1) // 错误的使用
			println(i)
			defer wg.Done()
		}(i)
	}

	wg.Wait()
}
