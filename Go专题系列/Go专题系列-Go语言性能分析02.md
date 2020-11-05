# 性能剖析工具PProf

## 第一节：压力测试工具ab

### ab工具介绍

ab是apache自带的压力测试工具。ab非常实用，它不仅可以对apache服务器进行网站访问压力测试，也可以对或其它类型的服务器进行压力测试。比如nginx、tomcat、IIS等

### 安装

```csharp
yum -y install httpd-tools # centos，redhat

apt-get install apache2-utils   #ubuntu Debian 
```

### 使用说明

```swift
# 格式
ab -n1000 -c 10 http:/www.snail.com/ 
```

常用参数说明：

```
-n  即requests，用于指定压力测试总共的执行次数。
-c  即concurrency，用于指定压力测试的并发数。
-t  即timelimit，等待响应的最大时间(单位：秒)。
```

```
# -n发出800个请求，-c模拟800并发，相当800人同时访问，后面是测试url
ab -n 800 -c 800  http://192.168.0.10/ 

#在60秒内发请求，一次100个并发请求。
ab -t 60 -c 100 http://192.168.0.10/ 
```

更多使用方法详见 [ab 官方文档](http://httpd.apache.org/docs/2.0/programs/ab.html)

### 结果说明

```
Server Software:        Apache          #服务器软件
Server Hostname:        www.taoquan.ink #域名
Server Port:            80              #请求端口号

Document Path:          /               #文件路径
Document Length:        40888 bytes     #页面字节数

Concurrency Level:      10              #请求的并发数
Time taken for tests:   27.300 seconds  #总访问时间
Complete requests:      1000            #请求成功数量
Failed requests:        0               #请求失败数量
Write errors:           0
Total transferred:      41054242 bytes  #请求总数据大小（包括header头信息）
HTML transferred:       40888000 bytes  #html页面实际总字节数
Requests per second:    36.63 [#/sec] (mean)  #每秒多少请求，这个是非常重要的参数数值，服务器的吞吐量
Time per request:       272.998 [ms] (mean)     #用户平均请求等待时间 
Time per request:       27.300 [ms] (mean, across all concurrent requests) # 服务器平均处理时间，也就是服务器吞吐量的倒数                  
Transfer rate:          1468.58 [Kbytes/sec] received  #每秒获取的数据长度

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:       43   47   2.4     47      53
Processing:   189  224  40.7    215     895
Waiting:      102  128  38.6    118     794
Total:        233  270  41.3    263     945

Percentage of the requests served within a certain time (ms)
  50%    263    #50%用户请求在263ms内返回
  66%    271    #66%用户请求在271ms内返回
  75%    279    #75%用户请求在279ms内返回
  80%    285    #80%用户请求在285ms内返回
  90%    303    #90%用户请求在303ms内返回
  95%    320    #95%用户请求在320ms内返回
  98%    341    #98%用户请求在341ms内返回
  99%    373    #99%用户请求在373ms内返回
 100%    945 (longest request)
```







------

## 第二节：PProf简介

### PProf是什么

PProf是分析性能、分析数据的工具，并支持可视化的图形分析。**是Go语言中必知必会的技能点**。

### PProf使用姿势

采样方式

- `runtime/pprof`：采集程序（非Server）指定区块的运行数据进行分析。·
- `net/http/pprof`：基于HTTP Server运行，并且可以采集运行时的数据进行分析。·
- `go test`：通过运行测试用例，指定所需标识进行采集。

使用方式

- Report Generation：报告生成。
- Interactive Terminal Use：交互式终端使用。
- Web Interface：Web界面。

### PProf可以做什么

-  CPU Profiling：CPU分析。按照一定的频率采集所监听的应用程序CPU（含寄存器）的使用情况，确定应用程序在主动消耗CPU周期时花费时间的位置。·
- Memory Profiling：内存分析。在应用程序进行堆分配时记录堆栈跟踪，用于监视当前和历史内存使用情况，以及检查内存泄漏。·
- Block Profiling：阻塞分析。记录goroutine阻塞等待同步（包括定时器通道）的位置，默认不开启，需要调用runtime.SetBlockProfileRate进行设置。·
- Mutex Profiling：互斥锁分析。报告互斥锁的竞争情况，默认不开启，需要调用runtime.SetMutexProfileFraction进行设置。
- Goroutine Profiling：goroutine分析，可以对当前应用程序正在运行的goroutine进行堆栈跟踪和分析。这项功能在实际排查中会经常用到，因为很多问题出现时的表象就是goroutine暴增，而这时候我们要做的事情之一就是查看应用程序中的 goroutine 正在做什么事情，因为什么阻塞了，然后再进行下一步分析。



------



## 第三节：PProf的简单使用

### 一个简单的例子：

```
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // 第一步～
)

// 创建map不指定容量
func makeMap1() map[int]int {
	mp := make(map[int]int)
	for i:=0;i<100000;i++{
		mp[i] = i
	}
	return mp
}
// 创建map指定容量
func makeMap2() map[int]int {
	mp := make(map[int]int,100000)
	for i:=0;i<100000;i++{
		mp[i] = i
	}
	return mp
}

func test1(w http.ResponseWriter, r *http.Request){
	makeMap1()
	makeMap2()
	fmt.Println(1111)
}

func main() {
	// 路由配置
	http.HandleFunc("/test1", test1)
	_ =http.ListenAndServe("0.0.0.0:6061", nil)
}
```

说明：

1. 在import中添加对`“net/http/pprof”`的引用

   ​	如果应用使用了自定义的 `Mux`，则需要手动注册一些路由规则：

   ```
   r.HandleFunc("/debug/pprof/", pprof.Index)
   r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
   r.HandleFunc("/debug/pprof/profile", pprof.Profile)
   r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
   r.HandleFunc("/debug/pprof/trace", pprof.Trace)
   ```

2. 访问http://服务器地址:端口/debug/pprof/地址，检查是否正常响应

### 通过浏览器访问

首页信息

![pprof_home](images/pprof_home.png)



| 类型         | 描述                                      |
| ------------ | ----------------------------------------- |
| allocs       | **内存分配**情况的采样信息                |
| blocks       | **阻塞**操作情况的采样信息                |
| cmdline      | 显示程序启动命令和其完整路径              |
| goroutine    | 显示当前所有**协程**的堆栈信息            |
| heap         | **堆**上的内存分配情况的采样信息          |
| mutex        | **锁**竞争情况的采样信息                  |
| profile      | **cpu**占用情况的采样信息，点击会下载文件 |
| threadcreate | 系统OS**线程**创建情况的采样信息          |
| trace        | 程序**运行跟踪**信息                      |

说明：

1. `?debug=1`,可以直接在浏览器中访问。
2. 不新增debug参数，那么将会直接下载对应的profile文件。
3. 在部署环境中，我们为了网络安全，通常不会直接对外网暴露 PProf 的相关端口，因此会通过curl、wget等方式进行profile文件的间接拉取
4. 在实际场景中，我们常常需要及时将当前状态下的profile文件给存储下来，便于二次分析。

## 第四节：通过交互式终端使用

第二种方式是直接通过命令行完成对正在运行的应用程序PProf进行抓取和分析。

### CPU Profiling:

```
go tool pprof http://127.0.0.1:6061/debug/pprof/profile?seconds=60
```

1. 在执行该命令后，需等待60 s（可调整 seconds 的值），PProf会进行 CPU Profiling，结束后将默认进入PProf的命令行交互式模式，查看或导出分析结果。

2. 输入查询命令 top10，查看对应资源开销（例如，CPU 就是执行耗时/开销、Memory 就是内存占用大小）排名前十的函数，命令如下：

   ```
   (pprof) top 15
   Showing nodes accounting for 10820ms, 85.60% of 12640ms total
   Dropped 161 nodes (cum <= 63.20ms)
   Showing top 15 nodes out of 99
         flat  flat%   sum%        cum   cum%
       5150ms 40.74% 40.74%     8950ms 70.81%  runtime.mapassign_fast64
        760ms  6.01% 46.76%      760ms  6.01%  runtime.memclrNoHeapPointers
        760ms  6.01% 52.77%      760ms  6.01%  runtime.stdcall3
        650ms  5.14% 57.91%     2100ms 16.61%  runtime.evacuate_fast64
        560ms  4.43% 62.34%      560ms  4.43%  runtime.bucketShift
        540ms  4.27% 66.61%      540ms  4.27%  runtime.stdcall1
        430ms  3.40% 70.02%      430ms  3.40%  runtime.cgocall
        320ms  2.53% 72.55%      320ms  2.53%  runtime.add
        320ms  2.53% 75.08%      320ms  2.53%  runtime.isEmpty
        300ms  2.37% 77.45%      300ms  2.37%  runtime.aeshash64
        230ms  1.82% 79.27%      230ms  1.82%  runtime.evacuated
        230ms  1.82% 81.09%      230ms  1.82%  runtime.procyield
        210ms  1.66% 82.75%     5740ms 45.41%  main.makeMap1
        210ms  1.66% 84.41%      210ms  1.66%  runtime.(*guintptr).cas
        150ms  1.19% 85.60%      150ms  1.19%  runtime.memmove
   (pprof)
   
   ```
   
   参数说明：
   
   - `flat`：当前函数的运行耗时。
   - `flat%`：当前函数占CPU运行总耗时的比例。
   - `sum%`：当前函数累积使用占CPU运行总耗时比例。
   - ` cum`：当前函数加上调用当前函数的函数占用CPU的总耗时。
   -  `cum%`：当前函数加上调用当前函数的函数占用CPU的总耗时百分比。
   - `最后一列`：函数名。

在大多数情况下，我们可以得出一个应用程序的运行情况，知道当前是什么函数，正在做什么事情，占用了多少资源等等，以此得到一个初步的分析方向。

我们还可以使用`list 函数名`命令查看具体的函数分析，例如执行`list makeMap1`查看我们编写的函数的详细分析 (若函数名不明确，则默认对该函数名进行模糊匹配):

```
(pprof) list makeMap1
Total: 2s
ROUTINE ======================== main.makeMap1 in D:\www\Snail\Go涓撻绯诲垪\code\Go璇█鎬ц兘鍒嗘瀽\pprof\main.go
      20ms      960ms (flat, cum) 48.00% of Total
         .          .      7:)
         .          .      8:
         .          .      9:// 鍒涘缓map涓嶆寚瀹氬閲?         .          .     10:func makeMap1() map[int]int {
         .          .     11:   mp := make(map[int]int)
      20ms       20ms     12:   for i:=0;i<100000;i++{
         .      940ms     13:           mp[i] = i
         .          .     14:   }
         .          .     15:   return mp
         .          .     16:}
         .          .     17:// 鍒涘缓map鎸囧畾瀹归噺
         .          .     18:func makeMap2() map[int]int {
(pprof)

```

可以看出该函数那一行占用CPU资源最多。

##### 常用交互命令行

- help  可以查看所有命令的使用说明

- **top** 	可以查看TOP多少分配情况

- **list** 	 展示源码及相应损耗

- **web** 	使用浏览器视图展开

- tree 	以树状显示

- png 	以图片格式输出

- svg 	生成浏览器可以识别的svg文件

- traces 打印所有调用栈信息

  注意：PProf中的所有功能都会根据 Profile的不同类型展示不同的对应结果

### Heap Profiling:

```
 go tool pprof http://127.0.0.1:6061/debug/pprof/heap
```

1. 在执行该命令后，能够很快地拉取到结果，因为它不像CPU Profiling那样需要做采样等待。
2. 它还有j几个个参数选项，默认选项是`inuse_space`
   1. inuse_space：未释放空间数。
   2. alloc_space : 所有分配空间数
   3. inuse_objects : 未释放对象数
   4. alloc_objects：所有分配对象数。

```
// 结果:
Fetching profile over HTTP from http://127.0.0.1:6061/debug/pprof/heap
Saved profile in C:\Users\viruser.v-desktop\pprof\pprof.alloc_objects.alloc_space.inuse_objects.inuse_space.007.pb.gz
Type: inuse_space
Time: Nov 4, 2020 at 10:37am (CST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 880.32kB, 100% of 880.32kB total
      flat  flat%   sum%        cum   cum%
  880.32kB   100%   100%   880.32kB   100%  main.makeMap1
         0     0%   100%   880.32kB   100%  main.test1
         0     0%   100%   880.32kB   100%  net/http.(*ServeMux).ServeHTTP
         0     0%   100%   880.32kB   100%  net/http.(*conn).serve
         0     0%   100%   880.32kB   100%  net/http.HandlerFunc.ServeHTTP
         0     0%   100%   880.32kB   100%  net/http.serverHandler.ServeHTTP
(pprof)

```



### Goroutine Profiling:

```
go tool pprof http://127.0.0.1:6061/debug/pprof/goroutine
```

这里我们在使用一个新的交互命令:

1. goroutine 时可以使用traces命令，这个命令会打印出对应的所有调用栈，以及指标信息。
2. 查看整个调用链路详情。在哪里使用了多少个goroutine,并且通过分析可以知道谁才是真正的调用方

```
(pprof) traces
Type: goroutine
Time: Nov 4, 2020 at 11:10am (CST)
-----------+-------------------------------------------------------
         1   runtime.gopark
             runtime.netpollblock
             internal/poll.runtime_pollWait
             internal/poll.(*pollDesc).wait
             internal/poll.(*ioSrv).ExecIO
             internal/poll.(*FD).Read
             net.(*netFD).Read
             net.(*conn).Read
             net/http.(*connReader).backgroundRead
-----------+-------------------------------------------------------
         1   runtime.gopark
             runtime.netpollblock
             internal/poll.runtime_pollWait
             internal/poll.(*pollDesc).wait
             internal/poll.(*ioSrv).ExecIO
             internal/poll.(*FD).acceptOne
             internal/poll.(*FD).Accept
             net.(*netFD).accept
             net.(*TCPListener).accept
             net.(*TCPListener).Accept
             net/http.(*Server).Serve
             net/http.(*Server).ListenAndServe
             net/http.ListenAndServe
             main.main
             runtime.main
-----------+-------------------------------------------------------
         1   runtime/pprof.writeRuntimeProfile
             runtime/pprof.writeGoroutine
             runtime/pprof.(*Profile).WriteTo
             net/http/pprof.handler.ServeHTTP
             net/http/pprof.Index
             net/http.HandlerFunc.ServeHTTP
             net/http.(*ServeMux).ServeHTTP
             net/http.serverHandler.ServeHTTP
             net/http.(*conn).serve
-----------+-------------------------------------------------------
(pprof)
```

说明：

1. 调用栈上的展示是自下而上的，也就是说 runtime.main方法调用了 main.main方法，而main.main方法又调用了 net/http.ListenAndServe 方法，排查起来比较方便。
2. 每个调用栈信息都是用 ------- 分割，函数方法前面的是指标数据，例如，Gorutine Profiling 展示的是该方法占用的 goroutine的数量



### Mutex Profiling:

在调用 chan （通道）、sync.Mutex （同步锁）或者 time.Sleep() 时会造成阻塞，为了验证互斥锁的竞争持有者的堆栈跟踪情况，我们调整先前的示例代码

```
package main

import (
	"net/http"
	_ "net/http/pprof" 
	"runtime"
	"sync"
)

func test1(w http.ResponseWriter, r *http.Request){
	var m sync.Mutex
	var datas = make(map[int]struct{})
	for i:=0;i<999;i++ {
		go func(i int) {
			m.Lock()
			defer m.Unlock()
			datas[i] = struct{}{}
		}(i)
	}
}

func init() {
	runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
}
func main() {
	// 路由配置
	http.HandleFunc("/test1", test1)
	_ =http.ListenAndServe("0.0.0.0:6061", nil)
}
```

**特别注意** ：`runtime.SetMutexProfileFraction`语句，如果未来希望对互斥锁进行采集，则需要调用该方法设置采集频率。如果没有设置，或设置的数值小于0，则不进行采集。

```
 go tool pprof http://127.0.0.1:6061/debug/pprof/mutex
```

调用top命令，查看互斥量的排名：

```
(pprof) top
Showing nodes accounting for 655.27us, 100% of 655.27us total
      flat  flat%   sum%        cum   cum%
  655.27us   100%   100%   655.27us   100%  sync.(*Mutex).Unlock
         0     0%   100%   655.27us   100%  main.test1.func1
(pprof)
```

调用 list命令 查看指定函数的代码情况 （包含特定的指标信息，如耗时）,这个地方表示引起互斥锁函数，以及锁开销的位置。

```
(pprof) list test1
Total: 655.27us
ROUTINE ======================== main.test1.func1 in D:\www\Snail\Go涓撻绯诲垪\code\Go璇█鎬ц兘鍒嗘瀽\pprof\main.go
         0   655.27us (flat, cum)   100% of Total
         .          .     13:   for i:=0;i<999;i++ {
         .          .     14:           go func(i int) {
         .          .     15:                   m.Lock()
         .          .     16:                   defer m.Unlock()
         .          .     17:                   datas[i] = struct{}{}
         .   655.27us     18:           }(i)
         .          .     19:   }
         .          .     20:}
         .          .     21:
         .          .     22:func init() {
         .          .     23:   runtime.SetMutexProfileFraction(1) // 寮€鍚閿佽皟鐢ㄧ殑璺熻釜
(pprof)
```



### Block Profiling:

与 Mutex 的 runtime.SetMutexProfileFraction 语句类似，Block也需要调用 runtime.SetBlockProfileRate 语句进行设置，如果没有设置，或者设置数值小于0，则不进行采集

```
package main

import (
	"net/http"
	_ "net/http/pprof" // 第一步～
	"runtime"
)

func test1(w http.ResponseWriter, r *http.Request){
	var ch chan int = make(chan int)
	for i:=0;i<999;i++ {
		go func(i int) {
			ch<-i
		}(i)
	}
}

func init() {
	runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
	runtime.SetBlockProfileRate(1)
}
func main() {
	// 路由配置
	http.HandleFunc("/test1", test1)
	_ =http.ListenAndServe("0.0.0.0:6061", nil)
}
```

```
go tool pprof http://127.0.0.1:6061/debug/pprof/block
```

调用 top 命令，查看阻塞情况排名：

```
(pprof) top
Showing nodes accounting for 331.46us, 100% of 331.46us total
      flat  flat%   sum%        cum   cum%
  331.46us   100%   100%   331.46us   100%  sync.(*Cond).Wait
         0     0%   100%   331.46us   100%  net/http.(*conn).serve
         0     0%   100%   331.46us   100%  net/http.(*connReader).abortPendingRead
         0     0%   100%   331.46us   100%  net/http.(*response).finishRequest
(pprof)
```

ps：

Cond的主要作用就是获取锁之后，wait()方法会等待一个通知，来进行下一步锁释放等操作，以此控制锁合适的释放，释放频率等。适用于在并发环境下goroutine的等待和通知。



## 第五节：可视化界面

```
wget http://127.0.0.1:6061/debug/pprof/profile
```

默认需要等待30s，执行完毕后在当前目录下可发现采集的profile文件。下面咱们来生成可视化界面：

```
// 这里端口自定义 只要你能访问到就行
go tool pprof -http=:8000 profile
```

可能会出现`Could not execute dot; may need to install graphviz.`，那么意味着需要安装 graphviz组件。

windows下安装：

1. `Graphviz.7z`解压后
2. 将graphviz安装目录下的bin文件夹添加到Path环境变量中。
3. 在终端输入dot -version查看是否安装成功。

通过PProf提供的可视化界面，我们能够更方便、更直观地看到Go应用程序的调用链和使用情况等。另外，在View菜单栏中，PProf还支持多种分析方式的切换，如图





接下来我们将对基于CPU Profiling抓取的profile文件进行一一介绍。profile类型的分析模式是互通的，只需了解一种即可。



https://blog.csdn.net/skh2015java/article/details/102748222



https://www.liwenzhou.com/posts/Go/performance_optimisation/

