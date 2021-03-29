package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func FindBT(ctx *context.Context,nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <- (*ctx).Done():
				return
			default:

			}
		}
	}()
	return out
}

func Download(inCh <-chan int) <-chan string {
	out := make(chan string)
	cT := time.NewTimer(1 * time.Second)
	defer cT.Stop()
	go func() {
		defer close(out)
		for {
			select {
			case n := <- inCh:
				if n >0 {
					out <- fmt.Sprintf("%d 下载完成",n)
				}else{
					return
				}
			case <-cT.C:
				return
			default:

			}
		}
	}()

	return out
}

func assgin(cs ...<-chan string) <-chan string {
	out := make(chan string)

	var wg sync.WaitGroup

	collect := func(in <-chan string) {
		defer wg.Done()
		for n := range in {
			out <- n
		}
	}

	wg.Add(3)

	for _, c := range cs {
		go collect(c)
	}

	//wg.Wait()
	//close(out)

	// 正确方式
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	ctx,cancel := context.WithTimeout(ctx,1 * time.Second)
	defer cancel()
	t := time.Now()
	want := FindBT(&ctx,1, 2, 3, 4)
	ctx := context.Background()

	// FAN-OUT    >1s  1取消掉  资源销毁
	movie1 := Download(want)
	movie2 := Download(want)
	movie3 := Download(want)

	// FAN-IN
	for ret := range assgin(movie1, movie2, movie3) {
		fmt.Printf("%s \n", ret)
	}

	fmt.Println("app elapsed:", time.Since(t))
	fmt.Println()
}
