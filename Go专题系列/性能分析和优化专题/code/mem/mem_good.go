package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof" // 第一步～
	"sync"
	"time"
)

var HttpClient *http.Client
var Once sync.Once

func HttpClientInstance() *http.Client {
	Once.Do(func() {
		HttpClient = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   50 * time.Millisecond,
					KeepAlive: 10 * time.Second,
				}).DialContext,
				MaxIdleConns:        200,
				MaxIdleConnsPerHost: 100,
				MaxConnsPerHost:     50,
				IdleConnTimeout:     1 * time.Second},
			Timeout: 1 * time.Second,
		}
	})
	return HttpClient
}

func main() {
	// 路由配置
	http.HandleFunc("/mem", myPrint)
	_ = http.ListenAndServe("0.0.0.0:6062", nil)
}

func myPrint(writer http.ResponseWriter, request *http.Request) {
	go doSomeThing()
	_, _ = writer.Write([]byte("mem"))
}
func doSomeThing() {
	for i := 0; i < 100; i++ {
		ticker := time.NewTicker(100 * time.Millisecond) //指定定时器间隔时间为1S
		go func() {
			<-ticker.C
			h()
		}()
		time.Sleep(5 * time.Second) //休眠10S为了看到效果，不然直接停了
		ticker.Stop()//停止该定时器 +++
	}
}


func h() []*int {
	_ = getjson()
	s := []*int{new(int), new(int), new(int), new(int)}
	// 使用此s切片 ...
	time.Sleep(1 * time.Second) //休眠10S为了看到效果，不然直接停了
	s[0], s[len(s)-1] = nil, nil // 重置首尾元素指针 +++
	return s[1:3:3]
}

func getjson() error{
	req, rerr := http.NewRequest("GET", "http://blog.xiaot123.com/mix-manifest.json", bytes.NewBuffer([]byte{}))
	if rerr != nil {
		return rerr
	}
	req.Header.Set("Content-Type", "application/json")

	resp, rserr := HttpClientInstance().Do(req)
	if rserr != nil {
		return  rserr
	}
	defer resp.Body.Close() // 关闭资源 +++
	res, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("resp byte length", len(res))
	return nil
}

