package main

import (
	"fmt"
	"time"
)

func main()  {
	a:=1
	go func() {
		a = 2
	}()
	go func() {
		a = 3
	}()
	fmt.Println(a)
	time.Sleep(1 * time.Second)
}