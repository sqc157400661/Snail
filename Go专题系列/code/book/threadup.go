package main
import (
	"fmt"
	"net"
	"runtime/pprof"
	"sync"
)
var threadProfile = pprof.Lookup("threadcreate")
func main() {
	// 开始前的线程数
	fmt.Printf(("threads in starting: %d\n"), threadProfile.Count())
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				net.LookupHost("www.google.com")
			}
		}()
	}
	wg.Wait()
	// goroutine执行完后的线程数
	fmt.Printf(("threads after LookupHost: %d\n"), threadProfile.Count())
}