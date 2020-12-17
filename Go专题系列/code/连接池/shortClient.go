package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   1000 * time.Millisecond,
			KeepAlive: 5000 * time.Second,
		}).DialContext,
		// MaxIdleConns:        100,
		// MaxIdleConnsPerHost: 100,
		// IdleConnTimeout:     2 * time.Second,
		DisableKeepAlives:true,
	},
	Timeout: 1000 * time.Millisecond,
}
var syncw sync.WaitGroup

func main() {
	bT := time.Now()            // 开始时间

	syncw.Add(10000)
	for i:=0;i<10000;i++{
		go DoReq()
		time.Sleep(1 * time.Millisecond)
	}
	syncw.Wait()
	eT := time.Since(bT)      // 从开始到当前所消耗的时间
	fmt.Println("Run time: ", eT)

}


func DoReq(){
	fmt.Println("start")
	url := "https://121.41.101.109/index.php"
	req,_ := http.NewRequest("GET",url,nil)
	resp,rerr := httpClient.Do(req)
	if rerr == nil{
		defer resp.Body.Close()
	}
	syncw.Done()
	fmt.Println("end")
}