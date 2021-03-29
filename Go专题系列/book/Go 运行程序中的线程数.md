# Go 运行程序中的线程数

稍微入门Go语言的程序员都知道，`GOMAXPROCS`变量可以限制并发运行用户态Go代码操作系统的最大线程数，你甚至可以通过调用函数`func GOMAXPROCS(n int) int`在程序运行时改变最大线程数的大小，但是当你进一步阅读文档，或者更深入的应用Go语言开发的时候，你就会发现，实际线程数要比你设置的这个数要大，有时候甚至远远大于你设置的数值，更悲剧的是，即使你的并发任务回退到没有几个的时候，这些线程数还没有降下来，白白浪费内存空间和CPU的调度。

- [disk io引起golang线程数暴涨的问题](http://xiaorui.cc/archives/5171)
- [golang创建大量线程的问题分析](https://yuerblog.cc/2020/03/02/golang创建大量线程的问题分析/)
- [一个 Go 程序系统线程暴涨的问题](https://zhuanlan.zhihu.com/p/22474724)
- [极端情况下收缩 Go 的线程数](https://xargin.com/shrink-go-threads/)

Go的文档也说明了实际的Thread可能不受`GOMAXPROCS`限制，如下面的文档所说，Go代码进行系统调用的时候被block的线程数不受这个变量限制：

> The GOMAXPROCS variable limits the number of operating system threads that can execute user-level Go code simultaneously. There is no limit to the number of threads that can be blocked in system calls on behalf of Go code; those do not count against the GOMAXPROCS limit. This package's GOMAXPROCS function queries and changes the limit.

如果并发的blocking的系统调用很多，Go就会创建大量的线程，但是当系统调用完成后，这些线程因为Go运行时的设计，却不会被回收掉。具体讨论见[go issue #14592](https://github.com/golang/go/issues/14592)。这个issue已经是2016的issue了，都4年多了，从Go 1.6推到现在，依然没有人动手尝试修复或者改进它。很显然，这并不是一个很容易修复的工作。

我重新整理一下，加深一下自己对这个知识点的理解。读者看到这篇文章后也多看看文中提到的链接，看看大家遇到的情况和解决办法。

## 什么是blocking的系统调用?

那么什么是blocking的系统调用(system call)呢？[stackoverflow](https://stackoverflow.com/questions/19309136/what-is-meant-by-blocking-system-call/19313275)有一个问答，很好的回答了这个问题：

> A blocking system call is one that must wait until the action can be completed. read() would be a good example - if no input is ready, it'll sit there and wait until some is (provided you haven't set it to non-blocking, of course, in which case it wouldn't be a blocking system call). Obviously, while one thread is waiting on a blocking system call, another thread can be off doing something else.

阻塞的系统调用就是系统调用执行时，在完成之前调用者必须等待。`read()`就是一个很好的例子，如果没有数据可读，调用者就一直等待直到一些数据可读(在你没有将它设置为 `non-blocking`情况下)。

那么如此一来Go从网络`I/O`中read数据岂不是每个读取goroutine都会占用一个系统线程了么？不会的!Go使用netpoller处理[网络读写](https://morsmachine.dk/netpoller)，它使用`epoll(linux)、kqueue(BSD、Darwin)、IoCompletionPort(Windows)`的方式可以`poll network I/O`的状态。一旦接受了一个连接，连接的文件描述符就被设置为`non-blocking`，这也意味着一旦连接中没有数据，从其中read数据并不会被阻塞，而是返回一个特定的错误，因此Go标准库的网络读写不会产生大量的线程，除非你把`GOMAXPROCS`设置的非常大，或者把底层的网络连接文件描述符又设置回了blocking模式。

但是`cgo`或者其它一些阻塞的系统调用可能就会导致线程大量增加并无法回收了，比如下面的例子。

## 线程数暴涨的简单测试

上面列出了各位大咖都是实际产品中遇到的例子，我来举一个简单的例子，你就可以看到大量的线程产生了。

```
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
```

Go提供了两种查询域名的方式，CGO方式或者纯Go方式，比如net库中的`Dial`、`LookupHost`、`LookupAddr`这些函数都会间接或者直接的与域名程序相关，比如上面的例子中使用`LookupHost`，采用不同的方式并发情况下产生的线程会不同。

比如采用纯Go的方式,程序在退出的时候会有10个线程：

```
$ GODEBUG=netdns=go go run main.go
threads in starting: 7
threads after LookupHost: 10
```

而采用cgo的方式，程序在退出的时候会有几十个甚至上百线程：

```
$ GODEBUG=netdns=cgo go run main.go
threads in starting: 7
threads after LookupHost: 109
```

## 无限暴涨？不可能!

Go运行时不会回收线程，而是会在需要的时候重用它们。但是你如果创建大量的线程，根本就是不需要的，理论上值保留一小部分线程重用就可以了。

如果程序设计的不合理，就会导致大量的空闲线程。如果你在http的处理程序中调用了类似的blocking系统调用或者CGO代码，或者微服务服务端调用了类似的代码，都有可能在客户端高并发访问时产生“线程泄露”的情况。

但是，系统的线程也不是无限创建，一来每个线程都会占用一定的内存资源，大量的线程导致内存枯竭，二来Go运行时其实对运行时创建的线程的数量还是有一个限制的，默认是10000个线程。

你可以使用`debug.SetMaxThreads`函数进行设置。比如你可以在上面的例子中将最大线程数设置为100:

```
   ......
   // 开始前的线程数
fmt.Printf(("threads in starting: %d\n"), threadProfile.Count())
debug.SetMaxThreads(100)
var wg sync.WaitGroup
wg.Add(100)
   ......
```

再运行上面的程序就会crash:

```
$ GODEBUG=netdns=cgo go run main.go
threads in starting: 7
runtime: program exceeds 100-thread limit
fatal error: thread exhaustion
runtime stack:
runtime.throw(0x54c3e2, 0x11)
        /usr/local/go/src/runtime/panic.go:1116 +0x72
runtime.checkmcount()
        /usr/local/go/src/runtime/proc.go:622 +0xac
runtime.mReserveID(0x62c878)
        /usr/local/go/src/runtime/proc.go:636 +0
......
```

## 减少线程

官方issue中也有人提供使用`LockOSThread`杀掉线程的方法，比如曹春晖大牛提供的一个函数逐个杀掉线程：

```go
// KillOne kills a thread
func KillOne() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		return
	}()
	wg.Wait()
}
```

`LockOSThread`函数会把当前goroutine绑定在当前的系统线程上，这个goroutine总是在这个线程中执行，而且也不会有其它goroutine在这个线程中执行。只有这个goroutine调用了相同次数的`UnlockOSThread`函数之后，才会进行解绑。

如果goroutine在退出的时候没有unlock这个线程，那么这个线程会被终止。我们正好可以利用这个特性将线程杀掉。我们可以启动一个goroutine,调用`LockOSThread`占住一个线程，尽然当前有很多空闲的线程，所以正好可以重用一个，goroutine退出的时候不调用`UnlockOSThread`，也就导致这个线程被终止了。

当然也有网友在官方issue提供了担心，杀掉一个空闲的线程有可能导致子进程会收到KIll信号。

你可以扩展这个方法，提供`Kill(n int)`可以终止多个线程的方法，当然原理都是类似的。从实践上上来看，你可以启动一个值守goroutine,检查到线程数超过某个阈值后就回收一部分线程，或者提供一个接口，可以手工调用某个API终止一部分线程，在官方还没有解决这个问题之前也不失是一种可用的方法