package main

import (
"fmt"
	"time"
)

func FindBT(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func Download(inCh <-chan int) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for n := range inCh {
			out <- fmt.Sprintf("%d 下载完成",n)
		}
	}()

	return out
}

func main() {
	t := time.Now()
	want := FindBT(1, 2, 3, 4)
	movie := Download(want)

	// consumer
	for ret := range movie {
		fmt.Printf("%s \n", ret)
	}
	fmt.Println("app elapsed:", time.Since(t).Seconds())
	fmt.Println("app elapsed:", time.Duration(111))
	fmt.Println()
}