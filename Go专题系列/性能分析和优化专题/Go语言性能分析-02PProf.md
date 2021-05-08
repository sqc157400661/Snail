# 性能剖析工具PProf

1.压力测试工具ab

2.PProf简介

3.PProf的简单使用

4.通过交互式终端使用PProf

5.通过可视化界面使用PProf

6.排查CPU占用过高问题

7.排查内存占用过高问题

8.排查频繁GC问题

9.排查协程泄漏问题

10.排查锁竞争问题

11.排查阻塞问题

12.性能优化案列

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

> pprof is a tool for visualization and analysis of profiling data.

PProf是分析性能、分析数据的工具，并支持可视化的图形分析。**是Go语言中必知必会的技能点**。

### PProf使用姿势

采样方式

- `runtime/pprof`：采集程序（非Server）指定区块的运行数据进行分析。·
- `net/http/pprof`：基于HTTP Server运行，并且可以采集运行时的数据进行分析。·
- `go test`：通过运行测试用例，指定所需标识进行采集。

使用方式

- Report Generation：报告生成。    [格式：`pprof <format> [options] source`]
- Interactive Terminal Use：交互式终端使用。  [格式：`pprof [options] source`]
- Web Interface：Web界面。 [格式：`pprof -http=[host]:[port] [options] source`]

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

![pprof_home](http://cdn.xiaot123.com/blog/2021-04/pprof_home.png-blog)



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

附：Heap 参数详解

```

Alloc uint64 //golang语言框架堆空间分配的字节数
TotalAlloc uint64 //从服务开始运行至今分配器为分配的堆空间总 和，只有增加，释放的时候不减少
Sys uint64 //服务现在系统使用的内存
Lookups uint64 //被runtime监视的指针数
Mallocs uint64 //服务malloc heap objects的次数
Frees uint64 //服务回收的heap objects的次数
HeapAlloc uint64 //服务分配的堆内存字节数
HeapSys uint64 //系统分配的作为运行栈的内存
HeapIdle uint64 //申请但是未分配的堆内存或者回收了的堆内存（空闲）字节数
HeapInuse uint64 //正在使用的堆内存字节数
HeapReleased uint64 //返回给OS的堆内存，类似C/C++中的free。
HeapObjects uint64 //堆内存块申请的量
StackInuse uint64 //正在使用的栈字节数
StackSys uint64 //系统分配的作为运行栈的内存
MSpanInuse uint64 //用于测试用的结构体使用的字节数
MSpanSys uint64 //系统为测试用的结构体分配的字节数
MCacheInuse uint64 //mcache结构体申请的字节数(不会被视为垃圾回收)
MCacheSys uint64 //操作系统申请的堆空间用于mcache的字节数
BuckHashSys uint64 //用于剖析桶散列表的堆空间
GCSys uint64 //垃圾回收标记元信息使用的内存
OtherSys uint64 //golang系统架构占用的额外空间
NextGC uint64 //垃圾回收器检视的内存大小
LastGC uint64 // 垃圾回收器最后一次执行时间。
PauseTotalNs uint64 // 垃圾回收或者其他信息收集导致服务暂停的次数。
PauseNs [256]uint64 //一个循环队列，记录最近垃圾回收系统中断的时间
PauseEnd [256]uint64 //一个循环队列，记录最近垃圾回收系统中断的时间开始点。
NumForcedGC uint32 //服务调用runtime.GC()强制使用垃圾回收的次数。
GCCPUFraction float64 //垃圾回收占用服务CPU工作的时间总和。如果有100个goroutine，垃圾回收的时间为1S,那么就占用了100S。
BySize //内存分配器使用情况
```



## 第四节：通过交互式终端使用

第二种方式是直接通过命令行完成对正在运行的应用程序PProf进行抓取和分析。

**注意事项：** 获取的 Profiling 数据是动态的，要想获得有效的数据，请保证应用处于较大的负载（比如正在生成中运行的服务，或者通过其他工具模拟访问压力）。否则如果应用处于空闲状态，得到的结果可能没有任何意义。

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
   - ` cum`：当前函数加上调用当前函数的函数占用CPU的总耗时。**通俗点说其实反映的是一个堆栈信息，可以对应grahp图和火焰图里的指标**
   -  `cum%`：当前函数加上调用当前函数的函数占用CPU的总耗时百分比。
   - `最后一列`：函数名。
   
   **举例说明：**函数`b`由三部分组成：调用函数`c`、自己直接处理一些事情、调用函数`d`，其中调用函数`c`耗时1秒，自己直接处理事情耗时3秒，调用函数`d`耗时2秒，那么函数`b`的`flat`耗时就是3秒，`cum`耗时就是6秒。
   
   ```swift
   // 该示例在文末参考列表的博客中
   func b() {
       c() // takes 1s
       do something directly // takes 3s
       d() // takes 2s
   }
   ```

##### 常用交互命令行

- help  可以查看所有命令的使用说明

- **top** 	可以查看TOP多少分配情况 ：`top -cum 15` 按照cum进行排序取前15个

- **list** 	 展示源码及相应损耗，可以看到那块代码耗时最多，那些可以做优化，一目了然

- **web** 	使用浏览器视图展开

- tree 	以树状显示

- png-blog 	以图片格式输出

- svg 	生成浏览器可以识别的svg文件

- traces 打印所有调用栈信息

  注意：PProf中的所有功能都会根据 Profile的不同类型展示不同的对应结果

例子：

可以使用`list 函数名`命令查看具体的函数分析，例如执行`list makeMap1`查看我们编写的函数的详细分析 (若函数名不明确，则默认对该函数名进行模糊匹配):

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

## 第五节：可视化界面

```
wget http://127.0.0.1:6061/debug/pprof/profile
```

**默认需要等待30s**，执行完毕后在当前目录下可发现采集的profile文件。下面咱们来生成可视化界面：

```
// 这里端口自定义 只要你能访问到就行
go tool pprof -http=:8000 profile
```

可能会出现`Could not execute dot; may need to install graphviz.`，那么意味着需要安装 graphviz组件。http://www.graphviz.org/download/

windows下安装：

1. `Graphviz.7z`解压后
2. 将graphviz安装目录下的bin文件夹添加到Path环境变量中。
3. 在终端输入dot -version查看是否安装成功。

通过PProf提供的可视化界面，我们能够更方便、更直观地看到Go应用程序的调用链和使用情况等。另外，在View菜单栏中，PProf还支持多种分析方式，如图

![pprof_gongneng](http://cdn.xiaot123.com/blog/2021-04/pprof_gongneng.png-blog)



### profile文件分析CPU Profiling

我们将对基于`CPU Profiling`抓取的profile文件进行一一介绍。其实profile文件类型的分析模式是互通的，只需了解一种即可。



#### Top

该视图与前面讲解命令行交互的top命令的作用和含义是一样的

ps：点击栏目可以进行相关的排序

![pprof_gongneng](http://cdn.xiaot123.com/blog/2021-04/pprof_top.png-blog)

#### Graph视图

视图展示的是整体的函数调用流程，框越大、线越粗、框颜色越鲜艳（红色），代表它占用的时间越久，开销越大。相反，框越小、线越浅、框颜色越淡，则代表在整体的函数调用流程中，它的开销越小。

### Interpreting the Callgraph  解释gragh图怎么看

- **Node Color**:节点颜色
  - large positive cum values are red.  消耗比较大的cum节点用红色标记
  - large negative cum values are green.减少比较多的cum节点用绿色标记（用于2个报告做对比时候）
  - cum values close to zero are grey.消耗比较小，接近零值的用灰色标记
- **Node Font Size**: 字体大小
  - larger font size means larger absolute flat values. 字体越大代码其代表的相应的值约大
  - smaller font size means smaller absolute flat values.字体越小，代表其相应的值约小
- **Edge Weight**:连线粗细
  - thicker edges indicate more resources were used along that path.较粗的连线，代表沿改路径使用了更多的资源
  - thinner edges indicate fewer resources were used along that path.较细的连线，代表沿改路径使用了较少的资源
- **Edge Color**: 连线颜色
  - large positive values are red. 大的正值是红色的
  - large negative values are green.大的负值是绿色的
  - values close to zero are grey.接近零的值为灰色
- **Dashed Edges**: some locations between the two connected locations were removed.删除了两个相连位置之间的某些位置。
- **Solid Edges**: one location directly calls the other.一个位置直接调用另一位置
- **"(inline)" Edge Marker**: the call has been inlined into the caller.

![pprof_gongneng](http://cdn.xiaot123.com/blog/2021-04/pprof_Graph.png-blog)

因此我们可以用此视图分析谁才是开销大头，它又是因为什么调用流程而被调用的。

#### Flame Graph视图

1. Flame Graph（火焰图）是动态的，调用顺序由上到下（A→B→C→D）
2. x 轴表示抽样数，如果一个函数在 x 轴占据的宽度越宽，就表示它被抽到的次数多，即执行的时间长。
3. y 轴表示调用栈，每一层都是一个函数。调用栈越深，火焰层越多
4. 每一块代表一个函数、区块越大，代表占用CPU的时间越长。同时它还支持点击块进行深入分析。

![pprof_gongneng](http://cdn.xiaot123.com/blog/2021-04/pprof_Flame_Graph.png-blog)

#### Peek视图

此视图与Top视图相比，增加了所属上下文信息的展示，即函数的输出调用者和被调用者。

![pprof_gongneng](http://cdn.xiaot123.com/blog/2021-04/pprof_Peek.png-blog)

#### Source视图

该视图主要增加了面向源代码的追踪和分析，可以看到其开销主要消耗在哪里。

![pprof_gongneng](http://cdn.xiaot123.com/blog/2021-04/pprof_Source.png-blog)



## 第六节：与性能测试结合做剖析

`go test`命令有两个参数和 pprof 相关，它们分别指定生成的 CPU 和 Memory profiling 保存的文件：

- -cpuprofile：cpu profiling 数据要保存的文件地址
- -memprofile：memory profiling 数据要报文的文件地址



### CPU profiling 

```
//执行性能测试的同时，也会执行 CPU profiling，并把结果保存在 cpu.prof 文件中：
go test -bench=. -cpuprofile=cpu.profile

//分析查看报告
go tool pprof -http=:8001 cpu.profile
```

### Memory profiling

```
//执行测试的同时，也会执行 Mem profiling，并把结果保存在 cpu.prof 文件中：
go test -bench . -memprofile=mem.profile

//分析查看报告
go tool pprof -http=:8001 mem.profile
```

需要注意的是，Profiling 一般和性能测试一起使用，这个原因在前文也提到过，只有应用在负载高的情况下 Profiling 才有意义。配合单元测试可以针对函数进行提前的性能优化

## 第七节：排查CPU占用过高问题

```
package main

import (
	"net/http"
	_ "net/http/pprof" // 第一步～
	"regexp"
)

func main() {
	// 路由配置
	http.HandleFunc("/cpu", myPrint)
	_ = http.ListenAndServe("0.0.0.0:6062", nil)
}

func myPrint(writer http.ResponseWriter, request *http.Request) {
	go func() {
		for i := 0; i < 100000; i++ {
			getPhone([]string{"18505921256", "13489594009", "12759029321"})
		}
	}()
	_, _ = writer.Write([]byte("cpu"))
}

func getPhone(s []string) bool{
	reg := `^1([38][0-9]|14[57]|5[^4])\d{8}$`
	rgx := regexp.MustCompile(reg)
	for _, v := range s {
		if rgx.MatchString(v) {
			return true
		}
	}
	return false
}
```

**第一步：**

```
ab -c100 -n100 'http://127.0.0.1:6062/cpu'
```

![1620379392570](D:\www\Snail\Go专题系列\images\1620379392570.png)

可以看到 CPU 占用相当高，这显然是有问题的，我们使用 `go tool pprof` 来排场一下：

**第二步：**

```
go tool pprof http://localhost:6062/debug/pprof/profile
```

采样完毕之后自动进入 pprof 的交互命令行界面：

![1620379518241](D:\www\Snail\Go专题系列\images\1620379518241.png)

 输入 top 命令，查看 CPU 占用较高的调用：

```
top -cum
```

![1620379594887](D:\www\Snail\Go专题系列\images\1620379594887.png)

可以看到 myprint和getPhone函数占用的资源最多

```
list myPrint

list getPhone
```

![1620383524085](D:\www\Snail\Go专题系列\images\1620383524085.png)

上面其实已经可以定位到比较慢的执行逻辑来，但是咱们仍然去利用web interface工具来看看

**第三步:**

```
go tool pprof -http=:8000 profile
```

结点比较多的时候输出会自动把一些耗时少的结点 drop 掉。也是合理的，没性能问题的流程你优化个啥啊，耍流氓？

![1620383981124](D:\www\Snail\Go专题系列\images\1620383981124.png)



这里同样也可以定位到是正则这里有问题，那么咱们接着往下走

**第四部：**比 pprof 更直观的火焰图

![1620384215613](D:\www\Snail\Go专题系列\images\1620384215613.png)



理论上输出火焰图之后我们最主要应该关注的是较宽的这些“平顶山”，定位代码的问题几乎就是秒级了，！也可能有人觉得火焰图虽直观，但并不具体。比如我想知道一个函数里每行消耗比较大的调用在整个过程中占用了多少时间，占用了百分之多少，这样我优化了以后才好去吹牛逼，我这次的优化使性能提高了百分之多少多少。这要怎么办呢。这时候还是得用 pprof。



**第五步：优化**

```
func myPrint(writer http.ResponseWriter, request *http.Request) {
	go func() {
		//for i := 0; i < 100000; i++ {
		getPhone([]string{"18505921256", "13489594009", "12759029321"})
		//}
	}()
	_, _ = writer.Write([]byte("cpu"))
}
```

![1620385374078](D:\www\Snail\Go专题系列\images\1620385374078.png)

可以看到性能提升还是比较明显的，

```
#重新生成报告
 go tool pprof -http=:8000  profile
#生成报告对比
go tool pprof -http=:8000 --base profile0 profile
```

![1620385564081](D:\www\Snail\Go专题系列\images\1620385564081.png)



## 第八节：排查内存占用过高问题



```
 go tool pprof http://127.0.0.1:6061/debug/pprof/heap
```

1. 在执行该命令后，能够很快地拉取到结果，因为它不像CPU Profiling那样需要做采样等待。
2. 它还有j几个个参数选项，默认选项是`inuse_space`
   1. inuse_space：收集实时的正在使用的分配空间数。当我们认为应用程序占据的 RSS 过大时，首先关注该指标。
   2. alloc_space : 收集自程序启动以来，累计的分配空间数。当应用历史上发生过内存使用大量上升时，首先关注该指标。
   3. inuse_objects : 收集实时的正在使用的分配对象数。当我们认为内存中的驻留对象过多时，首先关注该指标。
   4. alloc_objects：收集自程序启动以来，累计的分配对象数。当应用曾经发生过历史上的大量内存分配行为导致 CPU 或内存使用大幅上升时，首先关注该指标。
   
   > 网关类应用因为海量连接的关系，会导致进程消耗大量内存，所以我们经常看到相关的优化文章，主要就是降低应用的 inuse_space。
   >
   > 
   >
   > 两个对象数指标主要是为 GC 优化提供依据，当我们进行 GC 调优时，会同时关注应用分配的对象数、正在使用的对象数，以及 GC 的 CPU 占用的指标。

```

package main

import (
	"bytes"
	"fmt"
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
	}
}

func h() []*int {
	_ = getjson()
	s := []*int{new(int), new(int), new(int), new(int)}
	// 使用此s切片 ...
	time.Sleep(1 * time.Second) //休眠10S为了看到效果，不然直接停了
	return s[1:3:3]
}

func getjson() error {
	req, rerr := http.NewRequest("GET", "http://blog.xiaot123.com/mix-manifest.json", nil)
	if rerr != nil {
		return rerr
	}
	req.Header.Set("Content-Type", "application/json")

	resp, rserr := HttpClientInstance().Do(req)
	if rserr != nil {
		return rserr
	}
	var byteSlice []byte
	byteSlice = make([]byte, 0, 10*1024)
	buffer := bytes.NewBuffer(byteSlice)
	_, _ = buffer.ReadFrom(resp.Body) // ioutil.ReadAll(resp.Body) 这里一般用这个 是对这块的buffer.ReadFrom封装
	res := buffer.Bytes()
	fmt.Println("resp byte length", len(res))
	return nil
}

```

**第一步：首先来看一下火焰图**

alloc_objects：

因为cpu也很高，所以咱们先看一下alloc_objects

![1620455654531](D:\www\Snail\Go专题系列\images\1620455654531.png)

可以看到是getJson方式里，发生http请求里申请了大量的内存对象

inuse_objects ：

![1620455824222](D:\www\Snail\Go专题系列\images\1620455824222.png)



**第二步：代码优化**

```golang
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
```



第三步：对比数据

go tool pprof -http=:8000 --base heap.1 heap.2



























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





### trace分析

一般情况下我们是不需要使用 trace 来定位性能问题的，通过压测 + profile 就可以解决大部分问题，除非我们的问题与 runtime 本身的问题相关。

比如 STW 时间比预想中长，超过百毫秒，向官方反馈问题时，才需要出具相关的 trace 文件。比如类似 [long stw](https://github.com/golang/go/issues/19378) 这样的 issue。

采集 trace 对系统的性能影响还是比较大的，即使我们只是开启 gctrace，把 gctrace 日志重定向到文件，对系统延迟也会有一定影响，因为 gctrace 的 print 是在 stw 期间来做的：[gc trace 阻塞调度](http://xiaorui.cc/archives/6232)。

先采集trace信息，并下载分析文件

```
wget  http://127.0.0.1:6061/debug/pprof/trace?seconds=10
```

trace文件存下来后，需要用到`go tool` 的`trace`命令来解析，才能像其他的命令一样看到详细的tracec信息。
执行`go tool trace trace文件`解析trace文件，之后会有一个端口，在这个端口上查看trace信息。如下如：

```
$ go tool trace trace
2020/11/06 17:09:49 Parsing trace...
2020/11/06 17:09:50 Splitting trace...
2020/11/06 17:09:50 Opening browser. Trace viewer is listening on http://127.0.0.1:51073
```











## 第七节：对比分析









https://pdf.us/2019/02/18/2772.html

https://blog.csdn.net/qq_30549833/article/details/89381790

