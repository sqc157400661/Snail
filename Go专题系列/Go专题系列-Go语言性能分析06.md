# Go调优工具-trace介绍

很多时候，仅仅使用 PProf 不一定能完整地观察并解决问题，这是因为在真实的程序中包含了许多的隐藏动作，例如，goroutine在执行时会做哪些操作？执行/阻塞了多长时间？在什么时候阻止的？在哪里被阻止的？谁又锁/解锁了它们？GC是如何影响goroutine的执行的？这些问题用PProf是很难分析出来的，这时可以用本节的主角trace来解决。

## 如何使用trace

1. 标准库导入runtime/trace

2. 使用trace.Start() 和 trace.Stop()开启和关闭trace，并生成跟踪文件

3. 使用`go tool trace trace文件` 解析跟踪文件，并使用可视化程序打开浏览器

```
package main
import (
	"fmt"
	"os"
	"runtime/trace"
	"time"
)

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = trace.Start(f)
	ch := make(chan string)
	go func(){
		time.Sleep(time.Second)
		ch<- "裸奔的蜗牛，黑乎乎"
		say := make(chan string)
		go sayHello(say)
		fmt.Println(<-say)
	}()
	fmt.Println(<-ch)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()
	// Your program here
}

func sayHello(s chan string){
	time.Sleep(time.Second)
	s<- "hello"
}
```

执行：

```
 go tool trace trace.out
```

开启浏览器：

![pprof_gongneng](images/trace-01.png)

1. View trace：查看跟踪。
2.  Goroutine analysis：goroutine 分析。
3. Network blocking_profile：网络阻塞概况。
4. Synchronization blocking_profile：同步阻塞概况。
5. Syscall blocking_profile：系统调用阻塞概况。
6. Scheduler latency profile：调度延迟概况。
7. User defined tasks：用户自定义任务。
8.  User defined regions：用户自定义区域。
9. Minimum mutator utilization：最低 mutator 利用率。

## trace分析说明