//下面通过代码直接插入的方式来获取trace. 内容会涉及到网络请求，涉及协程异步执行等。
package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"runtime/trace"
	"strconv"
	"sync"
	"time"
)


var wg sync.WaitGroup
var httpClient = &http.Client{Timeout: 30 * time.Second}

func SleepSomeTime() time.Duration{
	return time.Microsecond * time.Duration(rand.Int()%1000)
}

func create(readChan chan int) {
	defer wg.Done()
	for i := 0; i < 500; i++ {
		readChan <- getBodySize()
		SleepSomeTime()
	}
	close(readChan)
}

func convert(readChan chan int, output chan string) {
	defer wg.Done()
	for readChan := range readChan {
		output <- strconv.Itoa(readChan)
		SleepSomeTime()
	}
	close(output)
}

func outputStr(output chan string) {
	defer wg.Done()
	for _ = range output {
		// do nothing
		SleepSomeTime()
	}
}

// 获取taobao 页面大小
func getBodySize() int {
	resp, _ := httpClient.Get("https://taobao.com")
	res, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return len(res)
}

func run() {
	readChan, output := make(chan int), make(chan string)
	wg.Add(3)
	go create(readChan)
	go convert(readChan, output)
	go outputStr(output)
}

func main() {
	f, _ := os.Create("trace.out")
	defer f.Close()
	_ = trace.Start(f)
	defer trace.Stop()
	run()
	wg.Wait()
}