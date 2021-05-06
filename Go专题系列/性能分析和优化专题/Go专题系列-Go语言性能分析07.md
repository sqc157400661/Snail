# Go调优工具-GODEBUG

GODEBUG是Go语言中十分强大的工具之一，GODEBUG可以让Go程序在运行时输出调试信息。通过GODEBUG我们可以很直观地看到调度器或垃圾回收等详细信息，并且不需要安装其他插件，非常方便。

## GODEBUG基础知识

Go Scheduler的主要功能是对在处理器上运行的OS线程分发可运行的goroutine。一般来说，一提到调度器，就不得不提三个缩写，具体如下。

1. G：goroutine，实际上，每次调用go func时都会生成一个G。
2.  P：Processor，处理器，一般P的数量就是处理器的核数，可以通过GOMAXPROCS 进行修改。
3.  M：Machine，OS线程。

这三者的交互实际上来源于Go的M：N 调度模型，也就是说，M必须与P进行绑定，然后不断地在M上循环寻找可运行的G来执行相应的任务，具体内容可以详细阅读[Go Runtime Scheduler](https://www.jianshu.com/p/2f5b0aaec856)。Go Scheduler的工作流程如图

![pprof_gongneng](images/godebug-01.png)



- 当执行go func（）时，实际上就是创建了一个全新的goroutine，我们称它为G。·
- 新创建的G会被放入P的本地队列（local queue）或全局队列（global queue）中，准备下一步的动作，注意这里的P指的是创建G的P。
- 唤醒或创建M以便执行G。
- 不断地进行事件循环。
- 寻找可用状态下的G执行任务。
- 清除后，重新进入事件循环。

上面提到的全局队列和本地队列，从功能上来说都是存放正在等待运行的G的，不同之处在于，本地队列有数量限制，即不允许超过256个，并且在新建G时，会优先选择P的本地队列。如果本地队列满了，则将P的本地队列中的一半的G移到全局队列，这其实可以理解为调度资源的共享和再平衡。



另外，上图中的steal行为是用来做什么的呢？当创建新的G或者G变成可运行状态时，它会被推送并加入当前P的本地队列中。当P执行G完毕后，它开始“干活”，它会从本地队列中弹出G，同时检查当前本地队列是否为空。如果为空，则会随机地从其他P的本地队列中尝试窃取一半可运行的G到自己的名下，如下图所示。

![pprof_gongneng](images/godebug-02.png)

在这个例子中，P2 在本地队列中找不到可以运行的 G，因而它会执行steal 调度算法，随机选择其他处理器P1，并从P1的本地队列中窃取三个G到它自己的本地队列中。至此，P1、P2都拥有了可执行的G。P1中多余的G也不会被浪费，调度资源会更加平均的在多个处理器中进行流转。

## 开始GODEBUG

GODEBUG可以控制运行时的调试变量，参数之间以逗号分隔，格式为name=val。在调度器观察上会使用如下两个参数：·

- `schedtrace`：设置`schedtrace=X`，可以在运行时每X毫秒发出一行调度器的摘要信息到标准err输出中。
- `scheddetail`：设置`schedtrace=X`和`scheddetail=1`，可以在运行时每X毫秒发出一次详细的多行信息，信息内容包括调度程序、处理器、OS线程和goroutine的状态。

### 示例：

下面创建一个main.go文件，写入示例代码：

```
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"sync"
)

func calcSum(w *sync.WaitGroup, idx int) {
	defer w.Done()
	var sum, n int64
	for ; n < 1000000000; n++ {
		sum += n
	}
	fmt.Println(idx, sum)
}

func main() {
	runtime.GOMAXPROCS(1)

	f, _ := os.Create("trace.output")
	defer f.Close()

	_ = trace.Start(f)
	defer trace.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go calcSum(&wg, i)
	}
	wg.Wait()
}
```

#### 使用schedtrace

```
GODEBUG=schedtrace=1000 go run demo.go

SCHED 0ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=1 idlethreads=0 runqueue=0 [0 0]
# command-line-arguments
SCHED 0ms: gomaxprocs=2 idleprocs=0 threads=5 spinningthreads=0 idlethreads=1 runqueue=0 [0 0]
# command-line-arguments
SCHED 0ms: gomaxprocs=2 idleprocs=1 threads=5 spinningthreads=0 idlethreads=1 runqueue=0 [0 0]
SCHED 0ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=1 idlethreads=0 runqueue=0 [0 0]
SCHED 1013ms: gomaxprocs=2 idleprocs=2 threads=8 spinningthreads=0 idlethreads=4 runqueue=0 [0 0]
SCHED 1029ms: gomaxprocs=2 idleprocs=0 threads=5 spinningthreads=0 idlethreads=1 runqueue=0 [5 4]
SCHED 2025ms: gomaxprocs=2 idleprocs=2 threads=8 spinningthreads=0 idlethreads=4 runqueue=0 [0 0]
SCHED 2043ms: gomaxprocs=2 idleprocs=0 threads=5 spinningthreads=0 idlethreads=1 runqueue=3 [3 2]
SCHED 3038ms: gomaxprocs=2 idleprocs=2 threads=8 spinningthreads=0 idlethreads=4 runqueue=0 [0 0]
SCHED 3048ms: gomaxprocs=2 idleprocs=0 threads=5 spinningthreads=0 idlethreads=1 runqueue=4 [3 1]
SCHED 4054ms: gomaxprocs=2 idleprocs=2 threads=8 spinningthreads=0 idlethreads=4 runqueue=0 [0 0]
SCHED 4053ms: gomaxprocs=2 idleprocs=0 threads=5 spinningthreads=0 idlethreads=1 runqueue=6 [2 0]
SCHED 5089ms: gomaxprocs=2 idleprocs=2 threads=8 spinningthreads=0 idlethreads=4 runqueue=0 [0 0]
。。。。。。
SCHED 5056ms: gomaxprocs=2 idleprocs=0 threads=5 spinningthreads=0 idlethreads=1 runqueue=0 [0 0]
。。。。。。。

```

- sched：每一行都代表调度器的调试信息，后面提示的毫秒数表示从启动到现在的运行时间，输出的时间间隔受schedtrace值的影响。
- gomaxprocs：当前的CPU核心数（GOMAXPROCS的当前值）。
- idleprocs：空闲的处理器数量，后面的数字表示当前的空闲数量。
- threads：OS线程数量，后面的数字表示当前正在运行的线程数量。
- spinningthreads：自旋状态的OS线程数量。
- idlethreads：空闲的线程数量。
- runqueue：全局队列中的goroutine数量，后面的[0 0 ]分别代表这2个P的本地队列正在运行的goroutine数量。

下面讲解“自旋线程”这个概念，在Head First of Golang Scheduler中，对“自旋线程”的说明如下：Go Scheduler 的设计者在考虑了“OS的资源利用率”和“频繁的线程抢占给 OS 带来的负载”之后，提出了“Spinning Thread”（自旋线程）这个概念。也就是说，当“自旋线程”没有找到可供其调度执行的 goroutine 时，并不会销毁该线程，而是采取“自旋”的操作保存了下来。虽然看起来浪费了一些资源，但是考虑一下 syscall 的情景就可以知道，比起“自旋”操作，线程间频繁的抢占、创建和销毁操作带来的危害更大。

#### 使用scheddetail

如果想要看到调度器的完整信息，则可以增加scheddetail参数，这样即可更进一步地查看调度的细节逻辑，代码如下：

```
$ GODEBUG=scheddetail=1,schedtrace=1000 go run demo.go
SCHED 0ms: gomaxprocs=2 idleprocs=0 threads=5 spinningthreads=0 idlethreads=1 runqueue=0 gcwaiting=0 nmidlelocked=0 stopwait=0 sysmonwait=0
  P0: status=0 schedtick=0 syscalltick=0 m=-1 runqsize=0 gfreecnt=0
  P1: status=0 schedtick=3 syscalltick=0 m=-1 runqsize=0 gfreecnt=0
  M4: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=-1
  M3: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=true lockedg=-1
  M2: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 spinning=false blocked=false lockedg=-1
  M1: p=-1 curg=17 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=17
  M0: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 spinning=false blocked=false lockedg=1
  G1: status=4(chan receive) m=-1 lockedm=0
  G17: status=6() m=1 lockedm=1
  G2: status=4(force gc (idle)) m=-1 lockedm=-1
  G3: status=4(GC sweep wait) m=-1 lockedm=-1
  G4: status=1() m=-1 lockedm=-1
..........
```

这里抽取了1000ms时的调试信息，信息量比较大，涉及核心的GMP，下面先从每一个字段开始了解。

##### （1）G.

- status：G的运行状态。
- m：隶属哪一个M。
- lockedm：是否有锁定M。

**G的运行状态共有9种**，对于分析内部流转非常有帮助，如下表所示。

| 状态              | 指示 | 含义                                                         |
| ----------------- | ---- | ------------------------------------------------------------ |
| _Gidle            | 0    | 刚刚被分配，还没有进行的初始化                               |
| _Grunnable        | 1    | 已经在运行队列中，还没有执行用户代码                         |
| _Grunning         | 2    | 不在运行队列中，已经可以执行用户代码，此时已经分配了M和P     |
| _Gsyscall         | 3    | 正在执行系统调用，此时分配了M                                |
| _Gwaiting         | 4    | 在运行时被阻止，没有执行用户代码，也不再运行队列中，此时在正在某处阻塞等待中 |
| _Gmoribund_unused | 5    | 尚未使用，但是在gbd中进行了硬编码                            |
| _Gdead            | 6    | 尚未使用，这个状态可能是刚退出或者刚被初始化，此时它并没有执行用户代码，也有可能是没有分配堆栈 |
| _Genqueue_unused  | 7    | 尚未使用                                                     |
| _Gcopystack       | 8    | 正在复制堆栈，并没有执行用户代码，也不在运行队列中           |

在了解了各类状态的含义后，再来看看下面这部分代码：

```
G1: status=4(semacquire) m=-1 lockedm=-1
G2: status=4(force gc (idle)) m=-1 lockedm=-1
G3: status=4(GC sweep wait) m=-1 lockedm=-1
G17: status=1() m=-1 lockedm=-1
G18: status=2() m=4 lockedm=-1
```

在这段代码中，G1的运行状态为Gwaiting，并没有分配M和锁定。括号中的semacquire是什么含义呢？因为 status=4 表示的是 goroutine 在运行时时被阻止，而阻止它的事件正是semacquire事件。semacquire会检查信号量的情况，在合适的时机调用goparkunlock函数，把当前的goroutine放进等待队列，并把它设为Gwaiting状态。在实际运行中还有什么原因会导致这种现象呢？具体如下：



那么在实际运行中还有什么原因会导致这种现象呢，我们一起看看，如下

```
waitReasonZero                                    *// ""*
waitReasonGCAssistMarking                         *// "GC assist marking"*
waitReasonIOWait                                  *// "IO wait"*
waitReasonChanReceiveNilChan                      *// "chan receive (nil chan)"*
waitReasonChanSendNilChan                         *// "chan send (nil chan)"*
waitReasonDumpingHeap                             *// "dumping heap"*
waitReasonGarbageCollection                       *// "garbage collection"*
waitReasonGarbageCollectionScan                   *// "garbage collection scan"*
waitReasonPanicWait                               *// "panicwait"*
waitReasonSelect                                  *// "select"*
waitReasonSelectNoCases                           *// "select (no cases)"*
waitReasonGCAssistWait                            *// "GC assist wait"*
waitReasonGCSweepWait                             *// "GC sweep wait"*
waitReasonChanReceive                             *// "chan receive"*
waitReasonChanSend                                *// "chan send"*
waitReasonFinalizerWait                           *// "finalizer wait"*
waitReasonForceGGIdle                             *// "force gc (idle)"*
waitReasonSemacquire                              *// "semacquire"*
waitReasonSleep                                   *// "sleep"
waitReasonSyncCondWait                            *// "sync.Cond.Wait"*
waitReasonTimerGoroutineIdle                      *// "timer goroutine (idle)"*
waitReasonTraceReaderBlocked                      *// "trace reader (blocked)"*
waitReasonWaitForGCCycle                          *// "wait for GC cycle"*
waitReasonGCWorkerIdle                            *// "GC worker (idle)*
```

我们通过以上 `waitReason` 可以了解到 `Goroutine` 会被暂停运行的原因要素，也就是会出现在括号中的事件。

##### （2）M

- p：隶属哪一个 P。
- curg：当前正在使用哪个 G。
- runqsize：运行队列中的 G 数量。
- gfreecnt：可用的G（状态为 Gdead）。
- mallocing：是否正在分配内存。
- throwing：是否抛出异常。
- preemptoff：不等于空字符串的话，保持 curg 在这个 m 上运行。

##### （3）P

- status：P 的运行状态。
- schedtick：P 的调度次数。
- syscalltick：P 的系统调用次数。
- m：隶属哪一个 M。
- runqsize：运行队列中的 G 数量。
- gfreecnt：可用的G（状态为 Gdead）。

| 状态      | 值   | 含义                                                         |
| :-------- | :--- | :----------------------------------------------------------- |
| _Pidle    | 0    | 刚刚被分配，还没有进行进行初始化。                           |
| _Prunning | 1    | 当 M 与 P 绑定调用 acquirep 时，P 的状态会改变为 _Prunning。 |
| _Psyscall | 2    | 正在执行系统调用。                                           |
| _Pgcstop  | 3    | 暂停运行，此时系统正在进行 GC，直至 GC 结束后才会转变到下一个状态阶段。 |
| _Pdead    | 4    | 废弃，不再使用。                                             |

### 小小结

本节我们学习了调度的一些基础知识，并通过GODEBUG工具掌握了观察调度器的方法。通常我们会把GODEBUG和go tool trace工具结合使用，在实际使用中，类似的方法还有很多，组合巧用是重点。





## 参考：

1. [Debugging performance issues in Go programs](https://software.intel.com/en-us/blogs/2014/05/10/debugging-performance-issues-in-go-programs)
*   [A whirlwind tour of Go’s runtime environment variables](https://dave.cheney.net/tag/godebug)
*   [Go调度器系列（2）宏观看调度器](https://mp.weixin.qq.com/s?__biz=Mzg3MTA0NDQ1OQ==&amp;mid=2247483907&amp;idx=2&amp;sn=c955372683bc0078e14227702ab0a35e&amp;chksm=ce85c607f9f24f116158043f63f7ca11dc88cd519393ba182261f0d7fc328c7b6a94fef4e416&amp;scene=38#wechat_redirect)
*   [Go's work-stealing scheduler](https://rakyll.org/scheduler/)
*   [Scheduler Tracing In Go](https://www.ardanlabs.com/blog/2015/02/scheduler-tracing-in-go.html)
*   [Head First of Golang Scheduler](https://zhuanlan.zhihu.com/p/42057783)
*   [goroutine 的状态切换](http://xargin.com/state-of-goroutine/)
*   [Environment_Variables](https://golang.org/pkg/runtime/#hdr-Environment_Variables)

